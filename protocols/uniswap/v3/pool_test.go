package v3

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/internal/testenv"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/deploy"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

type PoolSuite struct {
	suite.Suite

	RPCURL              string
	PrivateKeyHex       string
	FactoryHex          string
	MulticallHex        string
	NicksFactoryHex     string
	NativeSymbol        string
	NativeDecimals      int
	Fee                 *big.Int
	Timeout             time.Duration
	ReceiptPollInterval time.Duration
	ReceiptTimeout      time.Duration
	ByteCode            string
	Constructor         *abi.Function
	Salt                *types.Hash

	ctx    context.Context
	cancel context.CancelFunc

	client    rpc.Client
	signer    core.Signer
	deployer  *deploy.NicksDeployer
	from      *types.Address
	factory   *types.Address
	multicall *types.Address
	nicks     *types.Address
	tokenA    *types.Address
	tokenB    *types.Address
	pool      *types.Address

	v3Factory *Factory
	v3Pool    *Pool
	erc20     *erc20.ERC20

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestPoolSuite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)
	if c.Addresses.Protocol.Uniswap.V3.Factory == "" {
		t.Skip("testenv: no Uniswap V3 factory configured for this network")
	}

	suite.Run(t, &PoolSuite{
		RPCURL:              c.RPCURL,
		PrivateKeyHex:       c.ActorPrivateKeyHex,
		FactoryHex:          c.Addresses.Protocol.Uniswap.V3.Factory,
		MulticallHex:        c.Addresses.Standard.Multicall,
		NicksFactoryHex:     c.Addresses.Standard.NicksFactory,
		NativeSymbol:        c.Constant.NativeSymbol,
		NativeDecimals:      c.Constant.NativeDecimals,
		Fee:                 big.NewInt(FeeMedium),
		Timeout:             10 * time.Minute,
		ReceiptPollInterval: c.Constant.ReceiptPollInterval,
		ReceiptTimeout:      c.Constant.ReceiptTimeout,
		ByteCode:            testenv.ERC20ByteCodeHex,
		Constructor:         abi.NewFunction("constructor", abi.Types{abi.String, abi.String, abi.Uint8}, nil),
		Salt:                core.Keccak256([]byte("go-evmkit")),
	})
}

func (s *PoolSuite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(s.T().Context(), s.Timeout)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.v3Factory = NewFactory(s.client)
	s.v3Pool = NewPool(s.client)
	s.erc20 = erc20.New(s.client)
	s.T().Logf("SetupSuite: client ready for %s", s.RPCURL)

	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())
	s.T().Logf("SetupSuite: signing as %s", s.from)

	s.factory, err = types.NewAddressFromHex(s.FactoryHex)
	s.Require().NoError(err)
	s.T().Logf("SetupSuite: target UniswapV3Factory %s (fee tier %s)", s.factory, s.Fee)

	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)

	s.nicks, err = types.NewAddressFromHex(s.NicksFactoryHex)
	s.Require().NoError(err)

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	runtime := fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())

	names := [2]string{"evmkit test token A", "evmkit test token B"}
	symbols := [2]string{"EVKA", "EVKB"}
	tokens := [2]*types.Address{}

	for i := range 2 {
		var salt [32]byte
		copy(salt[:], core.Keccak256(fmt.Appendf(nil, "%s-%d", runtime, i)).Bytes())

		deployment := deploy.NewDeterministicContractDeployment(
			deploy.NewContractDeployment(bytecode, s.Constructor.Inputs, []any{names[i], symbols[i], uint8(18)}),
			salt,
		)

		predicted, err := s.deployer.DeployAddress(deploy.NewCreate2Request(s.nicks, deployment))
		s.Require().NoError(err)
		s.T().Logf("SetupSuite: deploying token %s %s, predicted address %s", names[i], symbols[i], predicted)

		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(150)

		deployTx, hash, err := s.deployer.Deploy(s.ctx, bind.NewTransactOptions(s.signer, builder), deploy.NewCreate2Request(s.nicks, deployment))
		s.Require().NoError(err)

		s.mu.Lock()
		s.totalFee.Add(s.totalFee, deployTx.MaxFee())
		s.mu.Unlock()
		s.T().Logf("SetupSuite: signed deployment of %s (nonce %d, maxFee %s)", symbols[i], deployTx.AddressNonce(), deployTx.MaxFee())

		s.T().Logf("SetupSuite: broadcast deployment of %s, tx %s", symbols[i], hash)

		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "deploy transaction reverted")
		s.T().Logf("SetupSuite: %s mined in block %d (gas %d)", symbols[i], receipt.BlockNumber, receipt.GasUsed)

		token, err := s.deployer.DeployAddress(deploy.NewCreate2Request(s.nicks, deployment))
		s.Require().NoError(err)
		tokens[i] = token

		s.Require().Eventually(func() bool {
			var codeHex string
			if err := s.client.Call(s.ctx, rpc.ETHGetCode("code", token.String(), rpc.BlockTagLatest, &codeHex)); err != nil {
				return false
			}
			code, err := encoding.Hex.Decode(codeHex)
			return err == nil && len(code) > 0
		}, 20*time.Second, s.ReceiptPollInterval, "no code visible at deployed token %s", token)

		s.T().Logf("SetupSuite: deployed token %s %s (nonce %d, tx %s)", symbols[i], token, deployTx.AddressNonce(), hash)
	}

	s.tokenA, s.tokenB = tokens[0], tokens[1]
	if s.tokenA.Big().Cmp(s.tokenB.Big()) > 0 {
		s.tokenA, s.tokenB = s.tokenB, s.tokenA
	}
	s.T().Logf("SetupSuite: sorted token0=%s token1=%s", s.tokenA, s.tokenB)

	createBuilder := tx.NewDynamicFeeBuilder()
	createBuilder.PriceBuffer(150)
	createTx, createHash, err := s.v3Factory.CreatePool(s.ctx, bind.NewTransactOptions(s.signer, createBuilder), NewCreatePoolRequest(s.factory, s.tokenA, s.tokenB, s.Fee))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, createTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("SetupSuite: sent createPool (nonce %d), tx %s", createTx.AddressNonce(), createHash)

	createReceipt, err := tx.WaitForReceipt(s.ctx, s.client, createHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(createReceipt.Status, "createPool transaction reverted")

	created, err := ExtractPoolCreated(createReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(created, 1)
	s.pool = created[0].Pool
	s.T().Logf("SetupSuite: created pool %s for (%s, %s, %s) tickSpacing=%s (block %d, gas %d)", s.pool, s.tokenA, s.tokenB, s.Fee, created[0].TickSpacing, createReceipt.BlockNumber, createReceipt.GasUsed)
}

func (s *PoolSuite) TearDownSuite() {
	if s.tokenA != nil && s.tokenB != nil {
		balances, err := s.erc20.AddressBalances(s.ctx, nil, erc20.NewAddressBalancesRequest(s.multicall, s.from, []*types.Address{s.tokenA, s.tokenB}, false))
		s.Require().NoError(err)

		tokenBurnFn := abi.NewFunction("burn", abi.NewTypes(abi.Address, abi.Uint256), nil)

		for _, tb := range balances.Tokens {
			if tb.Amount == nil || tb.Amount.Sign() == 0 {
				continue
			}

			calldata, err := tokenBurnFn.EncodeCall(s.from, tb.Amount)
			s.Require().NoError(err)

			unsignedBurn, err := tx.NewDynamicFeeBuilder().Build(s.from, tb.Token, big.NewInt(0), calldata).Pack(s.ctx, s.client)
			s.Require().NoError(err)
			s.Require().NoError(unsignedBurn.Sign(s.signer))
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, unsignedBurn.MaxFee())
			s.mu.Unlock()

			raw, err := unsignedBurn.EncodeRLP()
			s.Require().NoError(err)
			burnHash, err := tx.Broadcast(s.ctx, s.client, raw)
			s.Require().NoError(err)

			burnReceipt, err := tx.WaitForReceipt(s.ctx, s.client, burnHash, s.ReceiptPollInterval, s.ReceiptTimeout)
			s.Require().NoError(err)
			s.Require().True(burnReceipt.Status, "burn transaction reverted")

			s.T().Logf("TearDownSuite: burned %s of %s from %s (tx %s, block %d, gas %d)", tb.Amount, tb.Token, s.from, burnHash, burnReceipt.BlockNumber, burnReceipt.GasUsed)
		}
	}

	if s.totalFee != nil {
		base := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(s.NativeDecimals)), nil)
		whole, frac := new(big.Int), new(big.Int)
		whole.DivMod(s.totalFee, base, frac)

		fracStr := strings.TrimRight(fmt.Sprintf("%0*s", s.NativeDecimals, frac.String()), "0")
		decimal := whole.String()
		if fracStr != "" {
			decimal += "." + fracStr
		}

		s.T().Logf("PoolSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *PoolSuite) Test01Initialize() {
	before, err := s.v3Pool.Slot0(s.ctx, nil, NewSlot0Request(s.pool))
	s.Require().NoError(err)
	s.Require().Zero(before.SqrtPriceX96.Sign(), "pool shouldn't be initialized yet")

	sqrtPriceX96 := new(big.Int).Lsh(big.NewInt(1), 96)

	initBuilder := tx.NewDynamicFeeBuilder()
	initBuilder.PriceBuffer(150)
	sentTx, hash, err := s.v3Pool.Initialize(s.ctx, bind.NewTransactOptions(s.signer, initBuilder), NewInitializeRequest(s.pool, sqrtPriceX96))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test01Initialize: sent initialize(%s) (nonce %d, maxFee %s), tx %s", sqrtPriceX96, sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "initialize transaction reverted")

	initialized, err := ExtractInitialize(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(initialized, 1)
	s.Require().True(initialized[0].Contract.Equal(s.pool))
	s.Require().Zero(sqrtPriceX96.Cmp(initialized[0].SqrtPriceX96))
	s.Require().Zero(initialized[0].Tick.Sign())

	var after *Slot0
	s.Require().Eventually(func() bool {
		var err error
		after, err = s.v3Pool.Slot0(s.ctx, bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber)), NewSlot0Request(s.pool))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "slot0 not readable at the receipt block")
	s.Require().Zero(sqrtPriceX96.Cmp(after.SqrtPriceX96))
	s.Require().Zero(after.Tick.Sign())
	s.Require().True(after.Unlocked)

	s.T().Logf("UniswapV3Pool %s Test01Initialize Result", s.pool)
	s.T().Logf("\tInitialize event: sqrtPriceX96=%s tick=%s", initialized[0].SqrtPriceX96, initialized[0].Tick)
	s.T().Logf("\tSlot0: sqrtPriceX96=%s tick=%s observationIndex=%s observationCardinality=%s observationCardinalityNext=%s feeProtocol=%s unlocked=%v", after.SqrtPriceX96, after.Tick, after.ObservationIndex, after.ObservationCardinality, after.ObservationCardinalityNext, after.FeeProtocol, after.Unlocked)
	s.T().Logf("\tTx: %s (nonce %d, block %d, gas %d)", hash, sentTx.AddressNonce(), receipt.BlockNumber, receipt.GasUsed)
}

func (s *PoolSuite) Test02InitializeAlreadyInitialized() {
	sqrtPriceX96 := new(big.Int).Lsh(big.NewInt(1), 96)

	revertTx, revertHash, err := s.v3Pool.Initialize(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewInitializeRequest(s.pool, sqrtPriceX96))
	s.Require().ErrorIs(err, ErrAlreadyInitialized)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("UniswapV3Pool %s Test02InitializeAlreadyInitialized Result", s.pool)
	s.T().Logf("\tRequest: sqrtPriceX96=%s", sqrtPriceX96)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolSuite) Test03ReadMethods() {
	factory, err := s.v3Pool.Factory(s.ctx, nil, NewFactoryRequest(s.pool))
	s.Require().NoError(err)
	s.Require().True(factory.Equal(s.factory))

	token0, err := s.v3Pool.Token0(s.ctx, nil, NewToken0Request(s.pool))
	s.Require().NoError(err)
	s.Require().True(token0.Equal(s.tokenA))

	token1, err := s.v3Pool.Token1(s.ctx, nil, NewToken1Request(s.pool))
	s.Require().NoError(err)
	s.Require().True(token1.Equal(s.tokenB))

	fee, err := s.v3Pool.Fee(s.ctx, nil, NewFeeRequest(s.pool))
	s.Require().NoError(err)
	s.Require().Zero(s.Fee.Cmp(fee))

	tickSpacing, err := s.v3Pool.TickSpacing(s.ctx, nil, NewTickSpacingRequest(s.pool))
	s.Require().NoError(err)
	s.Require().Zero(big.NewInt(60).Cmp(tickSpacing))

	maxLiquidityPerTick, err := s.v3Pool.MaxLiquidityPerTick(s.ctx, nil, NewMaxLiquidityPerTickRequest(s.pool))
	s.Require().NoError(err)
	s.Require().Positive(maxLiquidityPerTick.Sign())

	slot0, err := s.v3Pool.Slot0(s.ctx, nil, NewSlot0Request(s.pool))
	s.Require().NoError(err)
	s.Require().NotZero(slot0.SqrtPriceX96.Sign())
	s.Require().True(slot0.Unlocked)

	liquidity, err := s.v3Pool.Liquidity(s.ctx, nil, NewLiquidityRequest(s.pool))
	s.Require().NoError(err)
	s.Require().Zero(liquidity.Sign(), "no liquidity minted yet")

	protocolFees, err := s.v3Pool.ProtocolFees(s.ctx, nil, NewProtocolFeesRequest(s.pool))
	s.Require().NoError(err)
	s.Require().Zero(protocolFees.Token0.Sign())
	s.Require().Zero(protocolFees.Token1.Sign())

	feeGrowthGlobal0X128, err := s.v3Pool.FeeGrowthGlobal0X128(s.ctx, nil, NewFeeGrowthGlobal0X128Request(s.pool))
	s.Require().NoError(err)
	s.Require().Zero(feeGrowthGlobal0X128.Sign())

	feeGrowthGlobal1X128, err := s.v3Pool.FeeGrowthGlobal1X128(s.ctx, nil, NewFeeGrowthGlobal1X128Request(s.pool))
	s.Require().NoError(err)
	s.Require().Zero(feeGrowthGlobal1X128.Sign())

	tickInfo, err := s.v3Pool.Ticks(s.ctx, nil, NewTicksRequest(s.pool, big.NewInt(-60)))
	s.Require().NoError(err)
	s.Require().False(tickInfo.Initialized)
	s.Require().Zero(tickInfo.LiquidityGross.Sign())

	bitmap, err := s.v3Pool.TickBitmap(s.ctx, nil, NewTickBitmapRequest(s.pool, 0))
	s.Require().NoError(err)
	s.Require().Zero(bitmap.Sign())

	var key [32]byte
	copy(key[:], core.Keccak256(append(append(s.from.Bytes(), 0xff, 0xff, 0xc4), 0x00, 0x00, 0x3c)).Bytes())
	position, err := s.v3Pool.Positions(s.ctx, nil, NewPositionsRequest(s.pool, key))
	s.Require().NoError(err)
	s.Require().Zero(position.Liquidity.Sign())
	s.Require().Zero(position.TokensOwed0.Sign())
	s.Require().Zero(position.TokensOwed1.Sign())

	observation, err := s.v3Pool.Observations(s.ctx, nil, NewObservationsRequest(s.pool, big.NewInt(0)))
	s.Require().NoError(err)
	s.Require().True(observation.Initialized)
	s.Require().Positive(observation.BlockTimestamp.Sign())

	observed, err := s.v3Pool.Observe(s.ctx, nil, NewObserveRequest(s.pool, []uint32{0}))
	s.Require().NoError(err)
	s.Require().Len(observed.TickCumulatives, 1)
	s.Require().Len(observed.SecondsPerLiquidityCumulativeX128s, 1)

	s.T().Logf("UniswapV3Pool %s Test03ReadMethods Result", s.pool)
	s.T().Logf("\tFactory: %s", factory)
	s.T().Logf("\tToken0: %s", token0)
	s.T().Logf("\tToken1: %s", token1)
	s.T().Logf("\tFee: %s", fee)
	s.T().Logf("\tTickSpacing: %s", tickSpacing)
	s.T().Logf("\tMaxLiquidityPerTick: %s", maxLiquidityPerTick)
	s.T().Logf("\tSlot0: sqrtPriceX96=%s tick=%s observationCardinality=%s unlocked=%v", slot0.SqrtPriceX96, slot0.Tick, slot0.ObservationCardinality, slot0.Unlocked)
	s.T().Logf("\tLiquidity: %s", liquidity)
	s.T().Logf("\tProtocolFees: token0=%s token1=%s", protocolFees.Token0, protocolFees.Token1)
	s.T().Logf("\tFeeGrowthGlobal0X128/1X128: %s/%s", feeGrowthGlobal0X128, feeGrowthGlobal1X128)
	s.T().Logf("\tTicks(-60): liquidityGross=%s liquidityNet=%s initialized=%v", tickInfo.LiquidityGross, tickInfo.LiquidityNet, tickInfo.Initialized)
	s.T().Logf("\tTickBitmap(0): %s", bitmap)
	s.T().Logf("\tPositions(%s, -60, 60): liquidity=%s tokensOwed0=%s tokensOwed1=%s", s.from, position.Liquidity, position.TokensOwed0, position.TokensOwed1)
	s.T().Logf("\tObservations(0): blockTimestamp=%s tickCumulative=%s initialized=%v", observation.BlockTimestamp, observation.TickCumulative, observation.Initialized)
	s.T().Logf("\tObserve([0]): tickCumulatives=%v secondsPerLiquidityCumulativeX128s=%v", observed.TickCumulatives, observed.SecondsPerLiquidityCumulativeX128s)
}

func (s *PoolSuite) Test04BurnNoPosition() {
	tickLower, tickUpper := big.NewInt(-60), big.NewInt(60)

	amount0, amount1, err := s.v3Pool.CallBurn(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewBurnRequest(s.pool, tickLower, tickUpper, big.NewInt(0)))
	s.Require().ErrorIs(err, ErrNoPosition)
	s.Require().Nil(amount0)
	s.Require().Nil(amount1)

	s.T().Logf("UniswapV3Pool %s Test04BurnNoPosition Result", s.pool)
	s.T().Logf("\tCallBurn(%s, %s, 0) error: %v", tickLower, tickUpper, err)

	revertTx, revertHash, err := s.v3Pool.Burn(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewBurnRequest(s.pool, tickLower, tickUpper, big.NewInt(0)))
	s.Require().ErrorIs(err, ErrNoPosition)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("\tBurn(%s, %s, 0) error: %v", tickLower, tickUpper, err)
}

func (s *PoolSuite) Test05CollectZeroLiquidity() {
	tickLower, tickUpper := big.NewInt(-60), big.NewInt(60)

	predictedAmount0, predictedAmount1, err := s.v3Pool.CallCollect(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewPoolCollectRequest(s.pool, s.from, tickLower, tickUpper, big.NewInt(0), big.NewInt(0)))
	s.Require().NoError(err)
	s.Require().Zero(predictedAmount0.Sign())
	s.Require().Zero(predictedAmount1.Sign())

	collectBuilder := tx.NewDynamicFeeBuilder()
	collectBuilder.PriceBuffer(150)
	sentTx, hash, err := s.v3Pool.Collect(s.ctx, bind.NewTransactOptions(s.signer, collectBuilder), NewPoolCollectRequest(s.pool, s.from, tickLower, tickUpper, big.NewInt(0), big.NewInt(0)))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "collect transaction reverted")

	collected, err := ExtractCollect(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(collected, 1)
	s.Require().True(collected[0].Contract.Equal(s.pool))
	s.Require().True(collected[0].Owner.Equal(s.from))
	s.Require().True(collected[0].Recipient.Equal(s.from))
	s.Require().Zero(tickLower.Cmp(collected[0].TickLower))
	s.Require().Zero(tickUpper.Cmp(collected[0].TickUpper))
	s.Require().Zero(collected[0].Amount0.Sign())
	s.Require().Zero(collected[0].Amount1.Sign())

	s.T().Logf("UniswapV3Pool %s Test05CollectZeroLiquidity Result", s.pool)
	s.T().Logf("\tCallCollect predicted amount0=%s amount1=%s", predictedAmount0, predictedAmount1)
	s.T().Logf("\tCollect event: owner=%s recipient=%s tickLower=%s tickUpper=%s amount0=%s amount1=%s", collected[0].Owner, collected[0].Recipient, collected[0].TickLower, collected[0].TickUpper, collected[0].Amount0, collected[0].Amount1)
	s.T().Logf("\tTx: %s (nonce %d, block %d, gas %d)", hash, sentTx.AddressNonce(), receipt.BlockNumber, receipt.GasUsed)
}

func (s *PoolSuite) Test06IncreaseObservationCardinalityNext() {
	before, err := s.v3Pool.Slot0(s.ctx, nil, NewSlot0Request(s.pool))
	s.Require().NoError(err)

	next := uint16(before.ObservationCardinalityNext.Uint64() + 5)

	increaseBuilder := tx.NewDynamicFeeBuilder()
	increaseBuilder.PriceBuffer(150)
	sentTx, hash, err := s.v3Pool.IncreaseObservationCardinalityNext(s.ctx, bind.NewTransactOptions(s.signer, increaseBuilder), NewIncreaseObservationCardinalityNextRequest(s.pool, next))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "increaseObservationCardinalityNext transaction reverted")

	increased, err := ExtractIncreaseObservationCardinalityNext(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(increased, 1)
	s.Require().True(increased[0].Contract.Equal(s.pool))
	s.Require().Zero(before.ObservationCardinalityNext.Cmp(increased[0].Old))
	s.Require().Zero(big.NewInt(int64(next)).Cmp(increased[0].New))

	var after *Slot0
	s.Require().Eventually(func() bool {
		var err error
		after, err = s.v3Pool.Slot0(s.ctx, bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber)), NewSlot0Request(s.pool))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "slot0 not readable at the receipt block")
	s.Require().Zero(big.NewInt(int64(next)).Cmp(after.ObservationCardinalityNext))

	s.T().Logf("UniswapV3Pool %s Test06IncreaseObservationCardinalityNext Result", s.pool)
	s.T().Logf("\tRequest: observationCardinalityNext=%d", next)
	s.T().Logf("\tIncreaseObservationCardinalityNext event: old=%s new=%s", increased[0].Old, increased[0].New)
	s.T().Logf("\tSlot0.ObservationCardinalityNext: %s -> %s", before.ObservationCardinalityNext, after.ObservationCardinalityNext)
	s.T().Logf("\tTx: %s (nonce %d, block %d, gas %d)", hash, sentTx.AddressNonce(), receipt.BlockNumber, receipt.GasUsed)
}

func (s *PoolSuite) Test07BurnTickLowerGreaterThanOrEqualUpper() {
	revertTx, revertHash, err := s.v3Pool.Burn(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewBurnRequest(s.pool, big.NewInt(60), big.NewInt(-60), big.NewInt(0)))
	s.Require().ErrorIs(err, ErrTickLowerGreaterThanOrEqual)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("UniswapV3Pool %s Test07BurnTickLowerGreaterThanOrEqualUpper Result", s.pool)
	s.T().Logf("\tRequest: tickLower=60 tickUpper=-60 amount=0")
	s.T().Logf("\tError: %v", err)
}

func (s *PoolSuite) Test08BurnTickLowerTooLow() {
	tooLow := big.NewInt(-887273)

	revertTx, revertHash, err := s.v3Pool.Burn(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewBurnRequest(s.pool, tooLow, big.NewInt(60), big.NewInt(0)))
	s.Require().ErrorIs(err, ErrTickLowerTooLow)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("UniswapV3Pool %s Test08BurnTickLowerTooLow Result", s.pool)
	s.T().Logf("\tRequest: tickLower=%s (MIN_TICK-1) tickUpper=60 amount=0", tooLow)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolSuite) Test09BurnTickUpperTooHigh() {
	tooHigh := big.NewInt(887273)

	revertTx, revertHash, err := s.v3Pool.Burn(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewBurnRequest(s.pool, big.NewInt(-60), tooHigh, big.NewInt(0)))
	s.Require().ErrorIs(err, ErrTickUpperTooHigh)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("UniswapV3Pool %s Test09BurnTickUpperTooHigh Result", s.pool)
	s.T().Logf("\tRequest: tickLower=-60 tickUpper=%s (MAX_TICK+1) amount=0", tooHigh)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolSuite) Test10SetFeeProtocolForbidden() {
	owner, err := s.v3Factory.Owner(s.ctx, nil, NewOwnerRequest(s.factory))
	s.Require().NoError(err)
	s.Require().False(owner.Equal(s.from), "s.from shouldn't be the factory owner")

	revertTx, revertHash, err := s.v3Pool.SetFeeProtocol(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSetFeeProtocolRequest(s.pool, 4, 10))
	s.Require().ErrorIs(err, ErrReverted)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("UniswapV3Pool %s Test10SetFeeProtocolForbidden Result", s.pool)
	s.T().Logf("\tCaller: %s (factory owner is %s)", s.from, owner)
	s.T().Logf("\tRequest: feeProtocol0=4 feeProtocol1=10")
	s.T().Logf("\tError: %v", err)
}

func (s *PoolSuite) Test11CollectProtocolForbidden() {
	owner, err := s.v3Factory.Owner(s.ctx, nil, NewOwnerRequest(s.factory))
	s.Require().NoError(err)

	amount0, amount1, err := s.v3Pool.CallCollectProtocol(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewCollectProtocolRequest(s.pool, s.from, big.NewInt(0), big.NewInt(0)))
	s.Require().ErrorIs(err, ErrReverted)
	s.Require().Nil(amount0)
	s.Require().Nil(amount1)

	s.T().Logf("UniswapV3Pool %s Test11CollectProtocolForbidden Result", s.pool)
	s.T().Logf("\tCaller: %s (factory owner is %s)", s.from, owner)
	s.T().Logf("\tCallCollectProtocol error: %v", err)

	revertTx, revertHash, err := s.v3Pool.CollectProtocol(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewCollectProtocolRequest(s.pool, s.from, big.NewInt(0), big.NewInt(0)))
	s.Require().ErrorIs(err, ErrReverted)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("\tCollectProtocol error: %v", err)
}

func (s *PoolSuite) Test12MintZeroAmount() {
	amount0, amount1, err := s.v3Pool.CallMint(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewPoolMintRequest(s.pool, s.from, big.NewInt(-60), big.NewInt(60), big.NewInt(0), nil))
	s.Require().ErrorIs(err, ErrReverted)
	s.Require().Nil(amount0)
	s.Require().Nil(amount1)

	s.T().Logf("UniswapV3Pool %s Test12MintZeroAmount Result", s.pool)
	s.T().Logf("\tCallMint(-60, 60, 0) error: %v", err)

	revertTx, revertHash, err := s.v3Pool.Mint(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewPoolMintRequest(s.pool, s.from, big.NewInt(-60), big.NewInt(60), big.NewInt(0), nil))
	s.Require().ErrorIs(err, ErrReverted)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("\tMint(-60, 60, 0) error: %v", err)
}

func (s *PoolSuite) Test13MintWithoutCallback() {
	amount0, amount1, err := s.v3Pool.CallMint(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewPoolMintRequest(s.pool, s.from, big.NewInt(-60), big.NewInt(60), big.NewInt(1_000_000), nil))
	s.Require().ErrorIs(err, ErrReverted)
	s.Require().Nil(amount0)
	s.Require().Nil(amount1)

	s.T().Logf("UniswapV3Pool %s Test13MintWithoutCallback Result", s.pool)
	s.T().Logf("\tCaller: %s (EOA, no uniswapV3MintCallback)", s.from)
	s.T().Logf("\tCallMint(-60, 60, 1000000) error: %v", err)
}

func (s *PoolSuite) Test14SwapAmountSpecifiedZero() {
	amount0, amount1, err := s.v3Pool.CallSwap(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewSwapRequest(s.pool, s.from, true, big.NewInt(0), big.NewInt(4295128740), nil))
	s.Require().ErrorIs(err, ErrAmountSpecifiedZero)
	s.Require().Nil(amount0)
	s.Require().Nil(amount1)

	s.T().Logf("UniswapV3Pool %s Test14SwapAmountSpecifiedZero Result", s.pool)
	s.T().Logf("\tCallSwap(zeroForOne=true, amountSpecified=0) error: %v", err)

	revertTx, revertHash, err := s.v3Pool.Swap(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSwapRequest(s.pool, s.from, true, big.NewInt(0), big.NewInt(4295128740), nil))
	s.Require().ErrorIs(err, ErrAmountSpecifiedZero)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("\tSwap(zeroForOne=true, amountSpecified=0) error: %v", err)
}

func (s *PoolSuite) Test15SwapInvalidSqrtPriceLimit() {
	slot0, err := s.v3Pool.Slot0(s.ctx, nil, NewSlot0Request(s.pool))
	s.Require().NoError(err)

	limit := new(big.Int).Add(slot0.SqrtPriceX96, big.NewInt(1))

	amount0, amount1, err := s.v3Pool.CallSwap(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewSwapRequest(s.pool, s.from, true, big.NewInt(1_000), limit, nil))
	s.Require().ErrorIs(err, ErrInvalidSqrtPriceLimit)
	s.Require().Nil(amount0)
	s.Require().Nil(amount1)

	s.T().Logf("UniswapV3Pool %s Test15SwapInvalidSqrtPriceLimit Result", s.pool)
	s.T().Logf("\tSlot0.SqrtPriceX96: %s", slot0.SqrtPriceX96)
	s.T().Logf("\tRequest: zeroForOne=true amountSpecified=1000 sqrtPriceLimitX96=%s (above current price)", limit)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolSuite) Test16FlashNoLiquidity() {
	liquidity, err := s.v3Pool.Liquidity(s.ctx, nil, NewLiquidityRequest(s.pool))
	s.Require().NoError(err)
	s.Require().Zero(liquidity.Sign())

	revertTx, revertHash, err := s.v3Pool.Flash(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewFlashRequest(s.pool, s.from, big.NewInt(1), big.NewInt(0), nil))
	s.Require().ErrorIs(err, ErrNoLiquidity)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("UniswapV3Pool %s Test16FlashNoLiquidity Result", s.pool)
	s.T().Logf("\tLiquidity: %s", liquidity)
	s.T().Logf("\tRequest: recipient=%s amount0=1 amount1=0", s.from)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolSuite) Test17SnapshotCumulativesInsideUninitializedTicks() {
	snapshot, err := s.v3Pool.SnapshotCumulativesInside(s.ctx, nil, NewSnapshotCumulativesInsideRequest(s.pool, big.NewInt(-60), big.NewInt(60)))
	s.Require().ErrorIs(err, ErrReverted)
	s.Require().Nil(snapshot)

	s.T().Logf("UniswapV3Pool %s Test17SnapshotCumulativesInsideUninitializedTicks Result", s.pool)
	s.T().Logf("\tRequest: tickLower=-60 tickUpper=60 (neither initialized)")
	s.T().Logf("\tError: %v", err)
}
