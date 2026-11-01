package v3

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/standards/deploy"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-evmkit/standards/multicall3"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

type PoolSuite struct {
	suite.Suite

	Enable        bool
	RPCURL        string
	PrivateKeyHex string
	FactoryHex    string
	MulticallHex  string
	NativeSymbol  string
	Fee           *big.Int
	Interval      int
	Timeout       int
	ByteCode      string
	Constructor   *abi.Function
	Salt          *types.Hash

	ctx    context.Context
	cancel context.CancelFunc

	client    rpc.Client
	signer    core.Signer
	deployer  *deploy.NicksDeployer
	from      *types.Address
	factory   *types.Address
	multicall *types.Address
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
	suite.Run(t, &PoolSuite{
		Enable:        false,
		RPCURL:        "",
		PrivateKeyHex: "",
		FactoryHex:    "",
		MulticallHex:  multicall3.DefaultMulticall3Address.String(),
		NativeSymbol:  "",
		Fee:           big.NewInt(FeeMedium),
		Interval:      1500,
		Timeout:       60,
		ByteCode:      erc20ByteCodeHex,
		Constructor:   abi.NewFunction("constructor", abi.Types{abi.String, abi.String, abi.Uint8}, nil),
		Salt:          core.Keccak256([]byte("go-evmkit")),
	})
}

func (s *PoolSuite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Minute)

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

		predicted, err := s.deployer.DeployAddress(deploy.DefaultNicksMethodFactory, deployment)
		s.Require().NoError(err)
		s.T().Logf("SetupSuite: deploying token %s %s, predicted address %s", names[i], symbols[i], predicted)

		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(150)

		unsigned, err := s.deployer.CreateDeployment(
			s.ctx,
			builder,
			s.from,
			deploy.DefaultNicksMethodFactory,
			big.NewInt(0),
			deployment,
		)
		s.Require().NoError(err)

		s.Require().NoError(unsigned.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, unsigned.MaxFee())
		s.mu.Unlock()
		s.T().Logf("SetupSuite: signed deployment of %s (nonce %d, maxFee %s)", symbols[i], unsigned.AddressNonce(), unsigned.MaxFee())

		hash, err := s.deployer.Deploy(s.ctx, unsigned)
		s.Require().NoError(err)
		s.T().Logf("SetupSuite: broadcast deployment of %s, tx %s", symbols[i], hash)

		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "deploy transaction reverted")
		s.T().Logf("SetupSuite: %s mined in block %d", symbols[i], receipt.BlockNumber)

		token, err := s.deployer.DeployAddress(deploy.DefaultNicksMethodFactory, deployment)
		s.Require().NoError(err)
		tokens[i] = token

		s.Require().Eventually(func() bool {
			var codeHex string
			if err := s.client.Call(s.ctx, rpc.ETHGetCode("code", token.String(), rpc.BlockTagLatest, &codeHex)); err != nil {
				return false
			}
			code, err := encoding.Hex.Decode(codeHex)
			return err == nil && len(code) > 0
		}, 20*time.Second, time.Duration(s.Interval)*time.Millisecond, "no code visible at deployed token %s", token)

		s.T().Logf("SetupSuite: deployed token %s %s (nonce %d, tx %s)", symbols[i], token, unsigned.AddressNonce(), hash)
	}

	s.tokenA, s.tokenB = tokens[0], tokens[1]
	if s.tokenA.Big().Cmp(s.tokenB.Big()) > 0 {
		s.tokenA, s.tokenB = s.tokenB, s.tokenA
	}
	s.T().Logf("SetupSuite: sorted token0=%s token1=%s", s.tokenA, s.tokenB)

	createBuilder := tx.NewDynamicFeeBuilder()
	createBuilder.PriceBuffer(150)
	unsignedCreate, err := s.v3Factory.CreateCreatePool(s.ctx, createBuilder, s.factory, s.from, s.tokenA, s.tokenB, s.Fee)
	s.Require().NoError(err)
	s.Require().NoError(unsignedCreate.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedCreate.MaxFee())
	s.mu.Unlock()

	createHash, err := s.v3Factory.CreatePool(s.ctx, unsignedCreate)
	s.Require().NoError(err)
	s.T().Logf("SetupSuite: broadcast createPool, tx %s", createHash)

	createReceipt, err := tx.WaitForReceipt(s.ctx, s.client, createHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(createReceipt.Status, "createPool transaction reverted")

	created, err := ExtractPoolCreated(createReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(created, 1)
	s.pool = created[0].Pool
	s.T().Logf("SetupSuite: created pool %s for (%s, %s, %s)", s.pool, s.tokenA, s.tokenB, s.Fee)
}

func (s *PoolSuite) TearDownSuite() {
	if s.tokenA != nil && s.tokenB != nil {
		balances, err := s.erc20.AddressBalances(s.ctx, s.multicall, s.from, []*types.Address{s.tokenA, s.tokenB}, false, rpc.BlockTagLatest)
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

			burnReceipt, err := tx.WaitForReceipt(s.ctx, s.client, burnHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
			s.Require().NoError(err)
			s.Require().True(burnReceipt.Status, "burn transaction reverted")

			s.T().Logf("TearDownSuite: burned %s of %s from %s (tx %s)", tb.Amount, tb.Token, s.from, burnHash)
		}
	}

	if s.totalFee != nil {
		base := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
		whole, frac := new(big.Int), new(big.Int)
		whole.DivMod(s.totalFee, base, frac)

		fracStr := strings.TrimRight(fmt.Sprintf("%018s", frac.String()), "0")
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
	before, err := s.v3Pool.Slot0(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(before.SqrtPriceX96.Sign(), "pool shouldn't be initialized yet")
	s.T().Logf("Test01Initialize: confirmed pool %s not yet initialized", s.pool)

	sqrtPriceX96 := new(big.Int).Lsh(big.NewInt(1), 96)

	initBuilder := tx.NewDynamicFeeBuilder()
	initBuilder.PriceBuffer(150)
	unsigned, err := s.v3Pool.CreateInitialize(s.ctx, initBuilder, s.pool, s.from, sqrtPriceX96)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test01Initialize: signed initialize (nonce %d, maxFee %s)", unsigned.AddressNonce(), unsigned.MaxFee())

	hash, err := s.v3Pool.Initialize(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test01Initialize: broadcast initialize, tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "initialize transaction reverted")
	s.T().Logf("Test01Initialize: mined in block %d", receipt.BlockNumber)

	initialized, err := ExtractInitialize(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(initialized, 1)
	s.Require().True(initialized[0].Contract.Equal(s.pool))
	s.Require().Zero(sqrtPriceX96.Cmp(initialized[0].SqrtPriceX96))
	s.T().Logf("Test01Initialize: Initialize event confirms sqrtPriceX96=%s tick=%s", initialized[0].SqrtPriceX96, initialized[0].Tick)

	after, err := s.v3Pool.Slot0(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(sqrtPriceX96.Cmp(after.SqrtPriceX96))
	s.Require().True(after.Unlocked)
	s.T().Logf("Test01Initialize: slot0 confirms sqrtPriceX96=%s tick=%s", after.SqrtPriceX96, after.Tick)

	s.T().Logf("UniswapV3Pool %s Test01Initialize Result", s.pool)
	s.T().Logf("\tSqrtPriceX96: %s", after.SqrtPriceX96)
	s.T().Logf("\tTick: %s", after.Tick)
	s.T().Logf("\tTx: %s", hash)
}

func (s *PoolSuite) Test02InitializeAlreadyInitialized() {
	sqrtPriceX96 := new(big.Int).Lsh(big.NewInt(1), 96)

	_, err := s.v3Pool.CreateInitialize(s.ctx, tx.NewDynamicFeeBuilder(), s.pool, s.from, sqrtPriceX96)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrAlreadyInitialized)
	s.T().Logf("Test02InitializeAlreadyInitialized: initialize(%s) reverted as expected: %v", sqrtPriceX96, err)
}

func (s *PoolSuite) Test03ReadMethods() {
	factory, err := s.v3Pool.Factory(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(factory.Equal(s.factory))

	token0, err := s.v3Pool.Token0(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(token0.Equal(s.tokenA))

	token1, err := s.v3Pool.Token1(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(token1.Equal(s.tokenB))

	fee, err := s.v3Pool.Fee(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(s.Fee.Cmp(fee))

	tickSpacing, err := s.v3Pool.TickSpacing(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(big.NewInt(60).Cmp(tickSpacing))

	maxLiquidityPerTick, err := s.v3Pool.MaxLiquidityPerTick(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Positive(maxLiquidityPerTick.Sign())

	slot0, err := s.v3Pool.Slot0(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().NotZero(slot0.SqrtPriceX96.Sign())
	s.Require().True(slot0.Unlocked)

	liquidity, err := s.v3Pool.Liquidity(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(liquidity.Sign(), "no liquidity minted yet")

	protocolFees, err := s.v3Pool.ProtocolFees(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(protocolFees.Token0.Sign())
	s.Require().Zero(protocolFees.Token1.Sign())

	feeGrowthGlobal0X128, err := s.v3Pool.FeeGrowthGlobal0X128(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(feeGrowthGlobal0X128.Sign())

	feeGrowthGlobal1X128, err := s.v3Pool.FeeGrowthGlobal1X128(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(feeGrowthGlobal1X128.Sign())

	s.T().Logf("UniswapV3Pool %s Test03ReadMethods Result", s.pool)
	s.T().Logf("\tFactory: %s", factory)
	s.T().Logf("\tToken0: %s", token0)
	s.T().Logf("\tToken1: %s", token1)
	s.T().Logf("\tFee: %s", fee)
	s.T().Logf("\tTickSpacing: %s", tickSpacing)
	s.T().Logf("\tMaxLiquidityPerTick: %s", maxLiquidityPerTick)
	s.T().Logf("\tSlot0.SqrtPriceX96: %s", slot0.SqrtPriceX96)
	s.T().Logf("\tLiquidity: %s", liquidity)
}

func (s *PoolSuite) Test04BurnNoPosition() {
	tickLower, tickUpper := big.NewInt(-60), big.NewInt(60)

	_, _, err := s.v3Pool.CallBurn(s.ctx, s.pool, s.from, tickLower, tickUpper, big.NewInt(0), rpc.BlockTagLatest)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrNoPosition)
	s.T().Logf("Test04BurnNoPosition: callBurn(%s, %s, 0) reverted as expected: %v", tickLower, tickUpper, err)

	_, err = s.v3Pool.CreateBurn(s.ctx, tx.NewDynamicFeeBuilder(), s.pool, s.from, tickLower, tickUpper, big.NewInt(0))
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrNoPosition)
	s.T().Logf("Test04BurnNoPosition: createBurn(%s, %s, 0) reverted as expected: %v", tickLower, tickUpper, err)
}

func (s *PoolSuite) Test05CollectZeroLiquidity() {
	tickLower, tickUpper := big.NewInt(-60), big.NewInt(60)

	predictedAmount0, predictedAmount1, err := s.v3Pool.CallCollect(s.ctx, s.pool, s.from, s.from, tickLower, tickUpper, big.NewInt(0), big.NewInt(0), rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(predictedAmount0.Sign())
	s.Require().Zero(predictedAmount1.Sign())
	s.T().Logf("Test05CollectZeroLiquidity: predicted amount0=%s amount1=%s", predictedAmount0, predictedAmount1)

	collectBuilder := tx.NewDynamicFeeBuilder()
	collectBuilder.PriceBuffer(150)
	unsigned, err := s.v3Pool.CreateCollect(s.ctx, collectBuilder, s.pool, s.from, s.from, tickLower, tickUpper, big.NewInt(0), big.NewInt(0))
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test05CollectZeroLiquidity: signed collect (nonce %d, maxFee %s)", unsigned.AddressNonce(), unsigned.MaxFee())

	hash, err := s.v3Pool.Collect(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test05CollectZeroLiquidity: broadcast collect, tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "collect transaction reverted")
	s.T().Logf("Test05CollectZeroLiquidity: mined in block %d", receipt.BlockNumber)

	collected, err := ExtractCollect(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(collected, 1)
	s.Require().True(collected[0].Owner.Equal(s.from))
	s.Require().True(collected[0].Recipient.Equal(s.from))
	s.Require().Zero(tickLower.Cmp(collected[0].TickLower))
	s.Require().Zero(tickUpper.Cmp(collected[0].TickUpper))
	s.Require().Zero(collected[0].Amount0.Sign())
	s.Require().Zero(collected[0].Amount1.Sign())
	s.T().Logf("Test05CollectZeroLiquidity: Collect event confirms amount0=%s amount1=%s", collected[0].Amount0, collected[0].Amount1)

	s.T().Logf("UniswapV3Pool %s Test05CollectZeroLiquidity Result", s.pool)
	s.T().Logf("\tTickLower: %s", tickLower)
	s.T().Logf("\tTickUpper: %s", tickUpper)
	s.T().Logf("\tAmount0: %s", collected[0].Amount0)
	s.T().Logf("\tAmount1: %s", collected[0].Amount1)
	s.T().Logf("\tTx: %s", hash)
}

func (s *PoolSuite) Test06IncreaseObservationCardinalityNext() {
	before, err := s.v3Pool.Slot0(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test06IncreaseObservationCardinalityNext: observationCardinalityNext before=%s", before.ObservationCardinalityNext)

	next := uint16(before.ObservationCardinalityNext.Uint64() + 5)

	increaseBuilder := tx.NewDynamicFeeBuilder()
	increaseBuilder.PriceBuffer(150)
	unsigned, err := s.v3Pool.CreateIncreaseObservationCardinalityNext(s.ctx, increaseBuilder, s.pool, s.from, next)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test06IncreaseObservationCardinalityNext: signed increaseObservationCardinalityNext(%d) (nonce %d, maxFee %s)", next, unsigned.AddressNonce(), unsigned.MaxFee())

	hash, err := s.v3Pool.IncreaseObservationCardinalityNext(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test06IncreaseObservationCardinalityNext: broadcast increaseObservationCardinalityNext, tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "increaseObservationCardinalityNext transaction reverted")
	s.T().Logf("Test06IncreaseObservationCardinalityNext: mined in block %d", receipt.BlockNumber)

	increased, err := ExtractIncreaseObservationCardinalityNext(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(increased, 1)
	s.Require().True(increased[0].Contract.Equal(s.pool))
	s.Require().Zero(before.ObservationCardinalityNext.Cmp(increased[0].Old))
	s.Require().Zero(big.NewInt(int64(next)).Cmp(increased[0].New))
	s.T().Logf("Test06IncreaseObservationCardinalityNext: event confirms old=%s new=%s", increased[0].Old, increased[0].New)

	after, err := s.v3Pool.Slot0(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(big.NewInt(int64(next)).Cmp(after.ObservationCardinalityNext))

	s.T().Logf("UniswapV3Pool %s Test06IncreaseObservationCardinalityNext Result", s.pool)
	s.T().Logf("\tObservationCardinalityNext: %s -> %s", before.ObservationCardinalityNext, after.ObservationCardinalityNext)
	s.T().Logf("\tTx: %s", hash)
}

func (s *PoolSuite) Test07BurnTickLowerGreaterThanOrEqualUpper() {
	_, err := s.v3Pool.CreateBurn(s.ctx, tx.NewDynamicFeeBuilder(), s.pool, s.from, big.NewInt(60), big.NewInt(-60), big.NewInt(0))
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrTickLowerGreaterThanOrEqual)
	s.T().Logf("Test07BurnTickLowerGreaterThanOrEqualUpper: burn(60, -60, 0) reverted as expected: %v", err)
}

func (s *PoolSuite) Test08BurnTickLowerTooLow() {
	tooLow := big.NewInt(-887273)
	_, err := s.v3Pool.CreateBurn(s.ctx, tx.NewDynamicFeeBuilder(), s.pool, s.from, tooLow, big.NewInt(60), big.NewInt(0))
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrTickLowerTooLow)
	s.T().Logf("Test08BurnTickLowerTooLow: burn(%s, 60, 0) reverted as expected: %v", tooLow, err)
}

func (s *PoolSuite) Test09BurnTickUpperTooHigh() {
	tooHigh := big.NewInt(887273)
	_, err := s.v3Pool.CreateBurn(s.ctx, tx.NewDynamicFeeBuilder(), s.pool, s.from, big.NewInt(-60), tooHigh, big.NewInt(0))
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrTickUpperTooHigh)
	s.T().Logf("Test09BurnTickUpperTooHigh: burn(-60, %s, 0) reverted as expected: %v", tooHigh, err)
}

func (s *PoolSuite) Test10SetFeeProtocolForbidden() {
	_, err := s.v3Factory.Owner(s.ctx, s.factory, rpc.BlockTagLatest)
	s.Require().NoError(err)

	_, err = s.v3Pool.CreateSetFeeProtocol(s.ctx, tx.NewDynamicFeeBuilder(), s.pool, s.from, 4, 10)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrReverted)
	s.T().Logf("Test10SetFeeProtocolForbidden: setFeeProtocol(4, 10) from non-owner %s reverted as expected: %v", s.from, err)
}

func (s *PoolSuite) Test11CollectProtocolForbidden() {
	_, err := s.v3Pool.CreateCollectProtocol(s.ctx, tx.NewDynamicFeeBuilder(), s.pool, s.from, s.from, big.NewInt(0), big.NewInt(0))
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrReverted)
	s.T().Logf("Test11CollectProtocolForbidden: collectProtocol(%s, 0, 0) from non-owner %s reverted as expected: %v", s.from, s.from, err)
}
