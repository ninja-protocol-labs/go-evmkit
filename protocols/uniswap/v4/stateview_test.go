package v4

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

type StateViewSuite struct {
	suite.Suite

	RPCURL                     string
	PrivateKeyHex              string
	PoolManagerHex             string
	StateViewHex               string
	PoolSwapTestHex            string
	PoolModifyLiquidityTestHex string
	MulticallHex               string
	NicksFactoryHex            string
	NativeSymbol               string
	NativeDecimals             int
	Timeout                    time.Duration
	ReceiptPollInterval        time.Duration
	ReceiptTimeout             time.Duration
	ByteCode                   string
	Constructor                *abi.Function
	Salt                       *types.Hash
	MintValue                  *big.Int
	FullRangeLiquidity         *big.Int
	NarrowLiquidity            *big.Int
	SwapAmountIn               *big.Int

	client                  rpc.Client
	signer                  core.Signer
	deployer                *deploy.NicksDeployer
	from                    *types.Address
	poolManagerContract     *types.Address
	stateViewContract       *types.Address
	poolSwapTest            *types.Address
	poolModifyLiquidityTest *types.Address
	multicall               *types.Address
	nicks                   *types.Address
	currency0               *types.Address
	currency1               *types.Address

	poolManager *PoolManager
	stateView   *StateView
	erc20       *erc20.ERC20

	key            *PoolKey
	poolID         PoolID
	stateSlot      [32]byte
	liquiditySlot  [32]byte
	fullRangeLower *big.Int
	fullRangeUpper *big.Int
	fullRangeSalt  [32]byte
	narrowLower    *big.Int
	narrowUpper    *big.Int
	narrowSalt     [32]byte
	preSwapBlock   string
	swapBlock      string

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestStateViewSuite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)
	if c.Addresses.Protocol.Uniswap.V4.StateView == "" {
		t.Skip("testenv: no Uniswap V4 StateView configured for this network")
	}
	if c.Addresses.Protocol.Uniswap.V4.PoolSwapTest == "" || c.Addresses.Protocol.Uniswap.V4.PoolModifyLiquidityTest == "" {
		t.Skip("testenv: no PoolSwapTest/PoolModifyLiquidityTest configured for this network")
	}

	suite.Run(t, &StateViewSuite{
		RPCURL:                     c.RPCURL,
		PrivateKeyHex:              c.ActorPrivateKeyHex,
		PoolManagerHex:             c.Addresses.Protocol.Uniswap.V4.PoolManager,
		StateViewHex:               c.Addresses.Protocol.Uniswap.V4.StateView,
		PoolSwapTestHex:            c.Addresses.Protocol.Uniswap.V4.PoolSwapTest,
		PoolModifyLiquidityTestHex: c.Addresses.Protocol.Uniswap.V4.PoolModifyLiquidityTest,
		MulticallHex:               c.Addresses.Standard.Multicall,
		NicksFactoryHex:            c.Addresses.Standard.NicksFactory,
		NativeSymbol:               c.Constant.NativeSymbol,
		NativeDecimals:             c.Constant.NativeDecimals,
		Timeout:                    10 * time.Minute,
		ReceiptPollInterval:        c.Constant.ReceiptPollInterval,
		ReceiptTimeout:             c.Constant.ReceiptTimeout,
		ByteCode:                   testenv.ERC20ByteCodeHex,
		Constructor:                abi.NewFunction("constructor", abi.Types{abi.String, abi.String, abi.Uint8}, nil),
		Salt:                       core.Keccak256([]byte("go-evmkit-v4-stateview")),
		MintValue:                  new(big.Int).Mul(big.NewInt(1_000_000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),
		FullRangeLiquidity:         new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
		NarrowLiquidity:            new(big.Int).Mul(big.NewInt(5), new(big.Int).Exp(big.NewInt(10), big.NewInt(17), nil)),
		SwapAmountIn:               new(big.Int).Exp(big.NewInt(10), big.NewInt(15), nil),
	})
}

func (s *StateViewSuite) SetupSuite() {
	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.poolManager = New(s.client)
	s.stateView = NewStateView(s.client)
	s.erc20 = erc20.New(s.client)

	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())
	s.T().Logf("SetupSuite: signing as %s", s.from)

	s.poolManagerContract, err = types.NewAddressFromHex(s.PoolManagerHex)
	s.Require().NoError(err)
	s.stateViewContract, err = types.NewAddressFromHex(s.StateViewHex)
	s.Require().NoError(err)
	s.poolSwapTest, err = types.NewAddressFromHex(s.PoolSwapTestHex)
	s.Require().NoError(err)
	s.poolModifyLiquidityTest, err = types.NewAddressFromHex(s.PoolModifyLiquidityTestHex)
	s.Require().NoError(err)
	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)
	s.nicks, err = types.NewAddressFromHex(s.NicksFactoryHex)
	s.Require().NoError(err)
	s.T().Logf("SetupSuite: StateView %s, PoolManager %s, PoolSwapTest %s, PoolModifyLiquidityTest %s", s.stateViewContract, s.poolManagerContract, s.poolSwapTest, s.poolModifyLiquidityTest)

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	runtime := fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())

	names := [2]string{"evmkit v4 stateview token A", "evmkit v4 stateview token B"}
	symbols := [2]string{"EVSA", "EVSB"}
	tokens := [2]*types.Address{}
	tokenMintFn := abi.NewFunction("mint", abi.NewTypes(abi.Address, abi.Uint256), nil)

	for i := range 2 {
		var salt [32]byte
		copy(salt[:], core.Keccak256(fmt.Appendf(nil, "%s-%d", runtime, i)).Bytes())

		deployment := deploy.NewDeterministicContractDeployment(
			deploy.NewContractDeployment(bytecode, s.Constructor.Inputs, []any{names[i], symbols[i], uint8(18)}),
			salt,
		)

		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(150)

		deployTx, hash, err := s.deployer.Deploy(s.T().Context(), bind.NewTransactOptions(s.signer, builder), deploy.NewCreate2Request(s.nicks, deployment))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, deployTx.MaxFee())
		s.mu.Unlock()

		receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "deploy transaction reverted")

		token, err := s.deployer.DeployAddress(deploy.NewCreate2Request(s.nicks, deployment))
		s.Require().NoError(err)
		tokens[i] = token

		s.Require().Eventually(func() bool {
			var codeHex string
			if err := s.client.Call(s.T().Context(), rpc.ETHGetCode("code", token.String(), rpc.BlockTagLatest, &codeHex)); err != nil {
				return false
			}
			code, err := encoding.Hex.Decode(codeHex)
			return err == nil && len(code) > 0
		}, 20*time.Second, s.ReceiptPollInterval, "no code visible at deployed token %s", token)
		s.T().Logf("SetupSuite: deployed token %s %s (tx %s, block %d)", symbols[i], token, hash, receipt.BlockNumber)

		calldata, err := tokenMintFn.EncodeCall(s.from, s.MintValue)
		s.Require().NoError(err)

		mintBuilder := tx.NewDynamicFeeBuilder()
		mintBuilder.PriceBuffer(150)
		mintTx, mintHash, err := bind.Transact(s.T().Context(), s.client, bind.NewTransactOptions(s.signer, mintBuilder), token, calldata)
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, mintTx.MaxFee())
		s.mu.Unlock()

		mintReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, mintHash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(mintReceipt.Status, "mint transaction reverted")
		s.T().Logf("SetupSuite: minted %s %s to %s (tx %s)", s.MintValue, symbols[i], s.from, mintHash)

		for _, spender := range []*types.Address{s.poolModifyLiquidityTest, s.poolSwapTest} {
			approveTx, approveHash, err := s.erc20.Approve(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), erc20.NewApproveRequest(token, spender, s.MintValue))
			s.Require().NoError(err)
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, approveTx.MaxFee())
			s.mu.Unlock()

			approveReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, approveHash, s.ReceiptPollInterval, s.ReceiptTimeout)
			s.Require().NoError(err)
			s.Require().True(approveReceipt.Status, "approve transaction reverted")
			s.T().Logf("SetupSuite: approved %s for %s (tx %s)", spender, symbols[i], approveHash)
		}
	}

	s.currency0, s.currency1 = tokens[0], tokens[1]
	if s.currency0.Big().Cmp(s.currency1.Big()) > 0 {
		s.currency0, s.currency1 = s.currency1, s.currency0
	}

	s.key = NewPoolKey(s.currency0, s.currency1, big.NewInt(3000), big.NewInt(60), types.ZeroAddress())
	s.poolID, err = s.key.PoolID()
	s.Require().NoError(err)

	var slotKey [64]byte
	copy(slotKey[:32], s.poolID[:])
	slotKey[63] = 6
	copy(s.stateSlot[:], core.Keccak256(slotKey[:]).Bytes())
	new(big.Int).Add(new(big.Int).SetBytes(s.stateSlot[:]), big.NewInt(3)).FillBytes(s.liquiditySlot[:])
	s.T().Logf("SetupSuite: currency0=%s currency1=%s poolId=%x", s.currency0, s.currency1, s.poolID)

	initTx, initHash, err := s.poolManager.Initialize(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewInitializeRequest(s.poolManagerContract, s.key, new(big.Int).Lsh(big.NewInt(1), 96)))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, initTx.MaxFee())
	s.mu.Unlock()

	initReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, initHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(initReceipt.Status, "initialize transaction reverted")
	s.T().Logf("SetupSuite: initialized pool (tx %s, block %d)", initHash, initReceipt.BlockNumber)

	s.fullRangeLower, s.fullRangeUpper = big.NewInt(-887220), big.NewInt(887220)
	copy(s.fullRangeSalt[:], core.Keccak256(fmt.Appendf(nil, "%s-full-range", runtime)).Bytes())
	s.narrowLower, s.narrowUpper = big.NewInt(-600), big.NewInt(600)
	copy(s.narrowSalt[:], core.Keccak256(fmt.Appendf(nil, "%s-narrow", runtime)).Bytes())

	modifyLiquidityTestFn := abi.NewFunction(
		"modifyLiquidity",
		abi.NewTypes(poolKeyTuple, liquidityParamsTuple, abi.Bytes, abi.Bool, abi.Bool),
		abi.NewTypes(abi.Bytes32),
	)

	for _, position := range []struct {
		name      string
		tickLower *big.Int
		tickUpper *big.Int
		liquidity *big.Int
		salt      [32]byte
	}{
		{"full range", s.fullRangeLower, s.fullRangeUpper, s.FullRangeLiquidity, s.fullRangeSalt},
		{"narrow", s.narrowLower, s.narrowUpper, s.NarrowLiquidity, s.narrowSalt},
	} {
		calldata, err := modifyLiquidityTestFn.EncodeCall(
			[]any{s.key.Currency0, s.key.Currency1, s.key.Fee, s.key.TickSpacing, s.key.Hooks},
			[]any{position.tickLower, position.tickUpper, position.liquidity, position.salt[:]},
			[]byte{}, false, false,
		)
		s.Require().NoError(err)

		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(150)
		modifyTx, modifyHash, err := bind.Transact(s.T().Context(), s.client, bind.NewTransactOptions(s.signer, builder), s.poolModifyLiquidityTest, calldata)
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, modifyTx.MaxFee())
		s.mu.Unlock()

		modifyReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, modifyHash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(modifyReceipt.Status, "modifyLiquidity transaction reverted")
		s.T().Logf("SetupSuite: added %s position [%s, %s] liquidity %s (tx %s, block %d)", position.name, position.tickLower, position.tickUpper, position.liquidity, modifyHash, modifyReceipt.BlockNumber)
	}
}

func (s *StateViewSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	if s.currency0 != nil && s.currency1 != nil {
		balances, err := s.erc20.AddressBalances(ctx, nil, erc20.NewAddressBalancesRequest(s.multicall, s.from, []*types.Address{s.currency0, s.currency1}, false))
		s.Require().NoError(err)

		tokenBurnFn := abi.NewFunction("burn", abi.NewTypes(abi.Address, abi.Uint256), nil)

		for _, tb := range balances.Tokens {
			if tb.Amount == nil || tb.Amount.Sign() == 0 {
				continue
			}

			calldata, err := tokenBurnFn.EncodeCall(s.from, tb.Amount)
			s.Require().NoError(err)

			unsignedBurn, err := tx.NewDynamicFeeBuilder().Build(s.from, tb.Token, big.NewInt(0), calldata).Pack(ctx, s.client)
			s.Require().NoError(err)
			s.Require().NoError(unsignedBurn.Sign(s.signer))
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, unsignedBurn.MaxFee())
			s.mu.Unlock()

			raw, err := unsignedBurn.EncodeRLP()
			s.Require().NoError(err)
			burnHash, err := tx.Broadcast(ctx, s.client, raw)
			s.Require().NoError(err)

			burnReceipt, err := tx.WaitForReceipt(ctx, s.client, burnHash, s.ReceiptPollInterval, s.ReceiptTimeout)
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

		s.T().Logf("StateViewSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}
}

func (s *StateViewSuite) Test01PoolManager() {
	poolManager, err := s.stateView.PoolManager(s.T().Context(), nil, NewPoolManagerRequest(s.stateViewContract))
	s.Require().NoError(err)
	s.Require().True(poolManager.Equal(s.poolManagerContract))

	s.T().Logf("StateView %s Test01PoolManager Result", s.stateViewContract)
	s.T().Logf("\tPoolManager: %s", poolManager)
}

func (s *StateViewSuite) Test02Slot0MatchesRawStorage() {
	slot0, err := s.stateView.GetSlot0(s.T().Context(), nil, NewGetSlot0Request(s.stateViewContract, s.poolID))
	s.Require().NoError(err)
	s.Require().Zero(new(big.Int).Lsh(big.NewInt(1), 96).Cmp(slot0.SqrtPriceX96))
	s.Require().Zero(slot0.Tick.Sign())
	s.Require().Zero(slot0.ProtocolFee.Sign())
	s.Require().Zero(s.key.Fee.Cmp(slot0.LPFee))

	raw, err := s.poolManager.Extsload(s.T().Context(), nil, NewExtsloadRequest(s.poolManagerContract, s.stateSlot))
	s.Require().NoError(err)
	word := new(big.Int).SetBytes(raw[:])
	rawPrice := new(big.Int).And(word, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 160), big.NewInt(1)))
	rawTick := new(big.Int).And(new(big.Int).Rsh(word, 160), big.NewInt(0xffffff))
	if rawTick.Bit(23) == 1 {
		rawTick.Sub(rawTick, big.NewInt(1<<24))
	}
	rawProtocolFee := new(big.Int).And(new(big.Int).Rsh(word, 184), big.NewInt(0xffffff))
	rawLPFee := new(big.Int).And(new(big.Int).Rsh(word, 208), big.NewInt(0xffffff))
	s.Require().Zero(rawPrice.Cmp(slot0.SqrtPriceX96))
	s.Require().Zero(rawTick.Cmp(slot0.Tick))
	s.Require().Zero(rawProtocolFee.Cmp(slot0.ProtocolFee))
	s.Require().Zero(rawLPFee.Cmp(slot0.LPFee))

	s.T().Logf("StateView %s Test02Slot0MatchesRawStorage Result", s.stateViewContract)
	s.T().Logf("\tGetSlot0: sqrtPriceX96=%s tick=%s protocolFee=%s lpFee=%s", slot0.SqrtPriceX96, slot0.Tick, slot0.ProtocolFee, slot0.LPFee)
	s.T().Logf("\tExtsload(stateSlot): %x", raw)
}

func (s *StateViewSuite) Test03LiquidityIsSumOfInRangePositions() {
	liquidity, err := s.stateView.GetLiquidity(s.T().Context(), nil, NewGetLiquidityRequest(s.stateViewContract, s.poolID))
	s.Require().NoError(err)
	want := new(big.Int).Add(s.FullRangeLiquidity, s.NarrowLiquidity)
	s.Require().Zero(want.Cmp(liquidity))

	raw, err := s.poolManager.Extsload(s.T().Context(), nil, NewExtsloadRequest(s.poolManagerContract, s.liquiditySlot))
	s.Require().NoError(err)
	s.Require().Zero(liquidity.Cmp(new(big.Int).SetBytes(raw[16:])))

	s.T().Logf("StateView %s Test03LiquidityIsSumOfInRangePositions Result", s.stateViewContract)
	s.T().Logf("\tGetLiquidity: %s (full range %s + narrow %s)", liquidity, s.FullRangeLiquidity, s.NarrowLiquidity)
	s.T().Logf("\tExtsload(liquiditySlot): %x", raw)
}

func (s *StateViewSuite) Test04FullRangePositionByFieldsAndByID() {
	byFields, err := s.stateView.GetPositionInfo(s.T().Context(), nil, NewGetPositionInfoRequest(s.stateViewContract, s.poolID, s.poolModifyLiquidityTest, s.fullRangeLower, s.fullRangeUpper, s.fullRangeSalt))
	s.Require().NoError(err)
	s.Require().Zero(s.FullRangeLiquidity.Cmp(byFields.Liquidity))
	s.Require().Zero(byFields.FeeGrowthInside0LastX128.Sign())
	s.Require().Zero(byFields.FeeGrowthInside1LastX128.Sign())

	positionID, err := CalculatePositionKey(s.poolModifyLiquidityTest, s.fullRangeLower, s.fullRangeUpper, s.fullRangeSalt)
	s.Require().NoError(err)

	byID, err := s.stateView.GetPositionInfoByID(s.T().Context(), nil, NewGetPositionInfoByIDRequest(s.stateViewContract, s.poolID, positionID))
	s.Require().NoError(err)
	s.Require().Equal(byFields, byID)

	liquidity, err := s.stateView.GetPositionLiquidity(s.T().Context(), nil, NewStateViewGetPositionLiquidityRequest(s.stateViewContract, s.poolID, positionID))
	s.Require().NoError(err)
	s.Require().Zero(s.FullRangeLiquidity.Cmp(liquidity))

	s.T().Logf("StateView %s Test04FullRangePositionByFieldsAndByID Result", s.stateViewContract)
	s.T().Logf("\tPosition: owner=%s [%s, %s] salt=%x id=%x", s.poolModifyLiquidityTest, s.fullRangeLower, s.fullRangeUpper, s.fullRangeSalt, positionID)
	s.T().Logf("\tGetPositionInfo: liquidity=%s feeGrowthInside0Last=%s feeGrowthInside1Last=%s", byFields.Liquidity, byFields.FeeGrowthInside0LastX128, byFields.FeeGrowthInside1LastX128)
	s.T().Logf("\tGetPositionLiquidity: %s", liquidity)
}

func (s *StateViewSuite) Test05NarrowPositionByFieldsAndByID() {
	byFields, err := s.stateView.GetPositionInfo(s.T().Context(), nil, NewGetPositionInfoRequest(s.stateViewContract, s.poolID, s.poolModifyLiquidityTest, s.narrowLower, s.narrowUpper, s.narrowSalt))
	s.Require().NoError(err)
	s.Require().Zero(s.NarrowLiquidity.Cmp(byFields.Liquidity))
	s.Require().Zero(byFields.FeeGrowthInside0LastX128.Sign())
	s.Require().Zero(byFields.FeeGrowthInside1LastX128.Sign())

	positionID, err := CalculatePositionKey(s.poolModifyLiquidityTest, s.narrowLower, s.narrowUpper, s.narrowSalt)
	s.Require().NoError(err)

	byID, err := s.stateView.GetPositionInfoByID(s.T().Context(), nil, NewGetPositionInfoByIDRequest(s.stateViewContract, s.poolID, positionID))
	s.Require().NoError(err)
	s.Require().Equal(byFields, byID)

	liquidity, err := s.stateView.GetPositionLiquidity(s.T().Context(), nil, NewStateViewGetPositionLiquidityRequest(s.stateViewContract, s.poolID, positionID))
	s.Require().NoError(err)
	s.Require().Zero(s.NarrowLiquidity.Cmp(liquidity))

	s.T().Logf("StateView %s Test05NarrowPositionByFieldsAndByID Result", s.stateViewContract)
	s.T().Logf("\tPosition: owner=%s [%s, %s] salt=%x id=%x", s.poolModifyLiquidityTest, s.narrowLower, s.narrowUpper, s.narrowSalt, positionID)
	s.T().Logf("\tGetPositionInfo: liquidity=%s feeGrowthInside0Last=%s feeGrowthInside1Last=%s", byFields.Liquidity, byFields.FeeGrowthInside0LastX128, byFields.FeeGrowthInside1LastX128)
	s.T().Logf("\tGetPositionLiquidity: %s", liquidity)
}

func (s *StateViewSuite) Test06PositionUnderEOAOwnerIsEmpty() {
	position, err := s.stateView.GetPositionInfo(s.T().Context(), nil, NewGetPositionInfoRequest(s.stateViewContract, s.poolID, s.from, s.narrowLower, s.narrowUpper, s.narrowSalt))
	s.Require().NoError(err)
	s.Require().Zero(position.Liquidity.Sign(), "positions are owned by PoolModifyLiquidityTest, not the EOA that sent the tx")

	s.T().Logf("StateView %s Test06PositionUnderEOAOwnerIsEmpty Result", s.stateViewContract)
	s.T().Logf("\tGetPositionInfo(owner=%s): liquidity=%s", s.from, position.Liquidity)
}

func (s *StateViewSuite) Test07PositionUnderOtherSaltIsEmpty() {
	position, err := s.stateView.GetPositionInfo(s.T().Context(), nil, NewGetPositionInfoRequest(s.stateViewContract, s.poolID, s.poolModifyLiquidityTest, s.narrowLower, s.narrowUpper, s.fullRangeSalt))
	s.Require().NoError(err)
	s.Require().Zero(position.Liquidity.Sign())

	var zeroSalt [32]byte
	zeroSaltPosition, err := s.stateView.GetPositionInfo(s.T().Context(), nil, NewGetPositionInfoRequest(s.stateViewContract, s.poolID, s.poolModifyLiquidityTest, s.narrowLower, s.narrowUpper, zeroSalt))
	s.Require().NoError(err)
	s.Require().Zero(zeroSaltPosition.Liquidity.Sign())

	s.T().Logf("StateView %s Test07PositionUnderOtherSaltIsEmpty Result", s.stateViewContract)
	s.T().Logf("\tGetPositionInfo(narrow range, full range salt): liquidity=%s", position.Liquidity)
	s.T().Logf("\tGetPositionInfo(narrow range, zero salt): liquidity=%s", zeroSaltPosition.Liquidity)
}

func (s *StateViewSuite) Test08TickInfoAtFullRangeBoundaries() {
	lower, err := s.stateView.GetTickInfo(s.T().Context(), nil, NewGetTickInfoRequest(s.stateViewContract, s.poolID, s.fullRangeLower))
	s.Require().NoError(err)
	s.Require().Zero(s.FullRangeLiquidity.Cmp(lower.LiquidityGross))
	s.Require().Zero(s.FullRangeLiquidity.Cmp(lower.LiquidityNet))

	upper, err := s.stateView.GetTickInfo(s.T().Context(), nil, NewGetTickInfoRequest(s.stateViewContract, s.poolID, s.fullRangeUpper))
	s.Require().NoError(err)
	s.Require().Zero(s.FullRangeLiquidity.Cmp(upper.LiquidityGross))
	s.Require().Zero(new(big.Int).Neg(s.FullRangeLiquidity).Cmp(upper.LiquidityNet))

	lowerLiquidity, err := s.stateView.GetTickLiquidity(s.T().Context(), nil, NewGetTickLiquidityRequest(s.stateViewContract, s.poolID, s.fullRangeLower))
	s.Require().NoError(err)
	s.Require().Zero(lower.LiquidityGross.Cmp(lowerLiquidity.LiquidityGross))
	s.Require().Zero(lower.LiquidityNet.Cmp(lowerLiquidity.LiquidityNet))

	upperOutside, err := s.stateView.GetTickFeeGrowthOutside(s.T().Context(), nil, NewGetTickFeeGrowthOutsideRequest(s.stateViewContract, s.poolID, s.fullRangeUpper))
	s.Require().NoError(err)
	s.Require().Zero(upper.FeeGrowthOutside0X128.Cmp(upperOutside.FeeGrowthOutside0X128))
	s.Require().Zero(upper.FeeGrowthOutside1X128.Cmp(upperOutside.FeeGrowthOutside1X128))

	s.T().Logf("StateView %s Test08TickInfoAtFullRangeBoundaries Result", s.stateViewContract)
	s.T().Logf("\tTick %s: gross=%s net=%s outside0=%s outside1=%s", s.fullRangeLower, lower.LiquidityGross, lower.LiquidityNet, lower.FeeGrowthOutside0X128, lower.FeeGrowthOutside1X128)
	s.T().Logf("\tTick %s: gross=%s net=%s outside0=%s outside1=%s", s.fullRangeUpper, upper.LiquidityGross, upper.LiquidityNet, upper.FeeGrowthOutside0X128, upper.FeeGrowthOutside1X128)
}

func (s *StateViewSuite) Test09TickInfoAtNarrowBoundaries() {
	lower, err := s.stateView.GetTickInfo(s.T().Context(), nil, NewGetTickInfoRequest(s.stateViewContract, s.poolID, s.narrowLower))
	s.Require().NoError(err)
	s.Require().Zero(s.NarrowLiquidity.Cmp(lower.LiquidityGross))
	s.Require().Zero(s.NarrowLiquidity.Cmp(lower.LiquidityNet))

	upper, err := s.stateView.GetTickInfo(s.T().Context(), nil, NewGetTickInfoRequest(s.stateViewContract, s.poolID, s.narrowUpper))
	s.Require().NoError(err)
	s.Require().Zero(s.NarrowLiquidity.Cmp(upper.LiquidityGross))
	s.Require().Zero(new(big.Int).Neg(s.NarrowLiquidity).Cmp(upper.LiquidityNet))

	upperLiquidity, err := s.stateView.GetTickLiquidity(s.T().Context(), nil, NewGetTickLiquidityRequest(s.stateViewContract, s.poolID, s.narrowUpper))
	s.Require().NoError(err)
	s.Require().Zero(upper.LiquidityGross.Cmp(upperLiquidity.LiquidityGross))
	s.Require().Zero(upper.LiquidityNet.Cmp(upperLiquidity.LiquidityNet))

	lowerOutside, err := s.stateView.GetTickFeeGrowthOutside(s.T().Context(), nil, NewGetTickFeeGrowthOutsideRequest(s.stateViewContract, s.poolID, s.narrowLower))
	s.Require().NoError(err)
	s.Require().Zero(lower.FeeGrowthOutside0X128.Cmp(lowerOutside.FeeGrowthOutside0X128))
	s.Require().Zero(lower.FeeGrowthOutside1X128.Cmp(lowerOutside.FeeGrowthOutside1X128))

	s.T().Logf("StateView %s Test09TickInfoAtNarrowBoundaries Result", s.stateViewContract)
	s.T().Logf("\tTick %s: gross=%s net=%s", s.narrowLower, lower.LiquidityGross, lower.LiquidityNet)
	s.T().Logf("\tTick %s: gross=%s net=%s", s.narrowUpper, upper.LiquidityGross, upper.LiquidityNet)
}

func (s *StateViewSuite) Test10TickInfoAtUninitializedTickIsZero() {
	info, err := s.stateView.GetTickInfo(s.T().Context(), nil, NewGetTickInfoRequest(s.stateViewContract, s.poolID, big.NewInt(60)))
	s.Require().NoError(err)
	s.Require().Zero(info.LiquidityGross.Sign())
	s.Require().Zero(info.LiquidityNet.Sign())
	s.Require().Zero(info.FeeGrowthOutside0X128.Sign())
	s.Require().Zero(info.FeeGrowthOutside1X128.Sign())

	s.T().Logf("StateView %s Test10TickInfoAtUninitializedTickIsZero Result", s.stateViewContract)
	s.T().Logf("\tTick 60: gross=%s net=%s", info.LiquidityGross, info.LiquidityNet)
}

func (s *StateViewSuite) Test11TickBitmapFlagsInitializedTicks() {
	s.T().Logf("StateView %s Test11TickBitmapFlagsInitializedTicks Result", s.stateViewContract)

	for _, tick := range []*big.Int{s.fullRangeLower, s.fullRangeUpper, s.narrowLower, s.narrowUpper} {
		compressed := tick.Int64() / s.key.TickSpacing.Int64()
		word := int16(compressed >> 8)
		bit := int(compressed & 0xff)

		bitmap, err := s.stateView.GetTickBitmap(s.T().Context(), nil, NewGetTickBitmapRequest(s.stateViewContract, s.poolID, word))
		s.Require().NoError(err)
		s.Require().Equal(uint(1), bitmap.Bit(bit), "tick %s must be flagged in word %d bit %d", tick, word, bit)

		s.T().Logf("\tTick %s: word %d bit %d, bitmap %x", tick, word, bit, bitmap)
	}
}

func (s *StateViewSuite) Test12TickBitmapEmptyWordIsZero() {
	bitmap, err := s.stateView.GetTickBitmap(s.T().Context(), nil, NewGetTickBitmapRequest(s.stateViewContract, s.poolID, 1))
	s.Require().NoError(err)
	s.Require().Zero(bitmap.Sign())

	s.T().Logf("StateView %s Test12TickBitmapEmptyWordIsZero Result", s.stateViewContract)
	s.T().Logf("\tGetTickBitmap(word 1): %s", bitmap)
}

func (s *StateViewSuite) Test13FeeGrowthGlobalsZeroBeforeSwap() {
	globals, err := s.stateView.GetFeeGrowthGlobals(s.T().Context(), nil, NewGetFeeGrowthGlobalsRequest(s.stateViewContract, s.poolID))
	s.Require().NoError(err)
	s.Require().Zero(globals.FeeGrowthGlobal0.Sign())
	s.Require().Zero(globals.FeeGrowthGlobal1.Sign())

	s.T().Logf("StateView %s Test13FeeGrowthGlobalsZeroBeforeSwap Result", s.stateViewContract)
	s.T().Logf("\tGetFeeGrowthGlobals: %s / %s", globals.FeeGrowthGlobal0, globals.FeeGrowthGlobal1)
}

func (s *StateViewSuite) Test14SwapUpdatesSlot0LiquidityAndFeeGrowth() {
	var preSwapBlockHex string
	s.Require().NoError(s.client.Call(s.T().Context(), rpc.ETHBlockNumber("blockNumber", &preSwapBlockHex)))
	s.preSwapBlock = preSwapBlockHex

	calldata, err := abi.NewFunction(
		"swap",
		abi.NewTypes(poolKeyTuple, swapAmountParamsTuple, abi.Tuple(abi.Bool, abi.Bool), abi.Bytes),
		abi.NewTypes(abi.Bytes32),
	).EncodeCall(
		[]any{s.key.Currency0, s.key.Currency1, s.key.Fee, s.key.TickSpacing, s.key.Hooks},
		[]any{true, new(big.Int).Neg(s.SwapAmountIn), big.NewInt(4295128740)},
		[]any{false, false},
		[]byte{},
	)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	sentTx, hash, err := bind.Transact(s.T().Context(), s.client, bind.NewTransactOptions(s.signer, builder), s.poolSwapTest, calldata)
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swap transaction reverted")
	s.swapBlock = fmt.Sprintf("0x%x", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.poolManagerContract))

	atBlock := bind.NewCallOptions().WithBlock(s.swapBlock)

	var slot0 *Slot0
	s.Require().Eventually(func() bool {
		slot0, err = s.stateView.GetSlot0(s.T().Context(), atBlock, NewGetSlot0Request(s.stateViewContract, s.poolID))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "slot0 not readable at the swap block")
	s.Require().Zero(swaps[0].SqrtPriceX96.Cmp(slot0.SqrtPriceX96))
	s.Require().Zero(swaps[0].Tick.Cmp(slot0.Tick))
	s.Require().Negative(slot0.Tick.Sign())
	s.Require().GreaterOrEqual(slot0.Tick.Cmp(s.narrowLower), 0, "the swap must stay inside the narrow range")

	liquidity, err := s.stateView.GetLiquidity(s.T().Context(), atBlock, NewGetLiquidityRequest(s.stateViewContract, s.poolID))
	s.Require().NoError(err)
	s.Require().Zero(swaps[0].Liquidity.Cmp(liquidity))

	globals, err := s.stateView.GetFeeGrowthGlobals(s.T().Context(), atBlock, NewGetFeeGrowthGlobalsRequest(s.stateViewContract, s.poolID))
	s.Require().NoError(err)
	s.Require().Positive(globals.FeeGrowthGlobal0.Sign(), "zeroForOne swap fees accrue in token0")
	s.Require().Zero(globals.FeeGrowthGlobal1.Sign())

	s.T().Logf("StateView %s Test14SwapUpdatesSlot0LiquidityAndFeeGrowth Result", s.stateViewContract)
	s.T().Logf("\tSwap event: amount0=%s amount1=%s sqrtPriceX96=%s tick=%s liquidity=%s", swaps[0].Amount0, swaps[0].Amount1, swaps[0].SqrtPriceX96, swaps[0].Tick, swaps[0].Liquidity)
	s.T().Logf("\tGetSlot0 at %s: sqrtPriceX96=%s tick=%s", s.swapBlock, slot0.SqrtPriceX96, slot0.Tick)
	s.T().Logf("\tGetFeeGrowthGlobals: %s / %s", globals.FeeGrowthGlobal0, globals.FeeGrowthGlobal1)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *StateViewSuite) Test15PreSwapBlockReadsInitialState() {
	s.Require().NotEmpty(s.preSwapBlock, "Test14 must record the pre-swap block")
	atBlock := bind.NewCallOptions().WithBlock(s.preSwapBlock)

	slot0, err := s.stateView.GetSlot0(s.T().Context(), atBlock, NewGetSlot0Request(s.stateViewContract, s.poolID))
	s.Require().NoError(err)
	s.Require().Zero(new(big.Int).Lsh(big.NewInt(1), 96).Cmp(slot0.SqrtPriceX96))
	s.Require().Zero(slot0.Tick.Sign())

	globals, err := s.stateView.GetFeeGrowthGlobals(s.T().Context(), atBlock, NewGetFeeGrowthGlobalsRequest(s.stateViewContract, s.poolID))
	s.Require().NoError(err)
	s.Require().Zero(globals.FeeGrowthGlobal0.Sign())

	s.T().Logf("StateView %s Test15PreSwapBlockReadsInitialState Result", s.stateViewContract)
	s.T().Logf("\tGetSlot0 at %s: sqrtPriceX96=%s tick=%s", s.preSwapBlock, slot0.SqrtPriceX96, slot0.Tick)
	s.T().Logf("\tGetFeeGrowthGlobals at %s: %s / %s", s.preSwapBlock, globals.FeeGrowthGlobal0, globals.FeeGrowthGlobal1)
}

func (s *StateViewSuite) Test16FeeGrowthInsideMatchesGlobalsWhileInRange() {
	s.Require().NotEmpty(s.swapBlock, "Test14 must record the swap block")
	atBlock := bind.NewCallOptions().WithBlock(s.swapBlock)

	globals, err := s.stateView.GetFeeGrowthGlobals(s.T().Context(), atBlock, NewGetFeeGrowthGlobalsRequest(s.stateViewContract, s.poolID))
	s.Require().NoError(err)

	fullRange, err := s.stateView.GetFeeGrowthInside(s.T().Context(), atBlock, NewGetFeeGrowthInsideRequest(s.stateViewContract, s.poolID, s.fullRangeLower, s.fullRangeUpper))
	s.Require().NoError(err)
	s.Require().Zero(globals.FeeGrowthGlobal0.Cmp(fullRange.FeeGrowthInside0X128))
	s.Require().Zero(globals.FeeGrowthGlobal1.Cmp(fullRange.FeeGrowthInside1X128))

	narrow, err := s.stateView.GetFeeGrowthInside(s.T().Context(), atBlock, NewGetFeeGrowthInsideRequest(s.stateViewContract, s.poolID, s.narrowLower, s.narrowUpper))
	s.Require().NoError(err)
	s.Require().Zero(globals.FeeGrowthGlobal0.Cmp(narrow.FeeGrowthInside0X128))
	s.Require().Zero(globals.FeeGrowthGlobal1.Cmp(narrow.FeeGrowthInside1X128))

	s.T().Logf("StateView %s Test16FeeGrowthInsideMatchesGlobalsWhileInRange Result", s.stateViewContract)
	s.T().Logf("\tGlobals: %s / %s", globals.FeeGrowthGlobal0, globals.FeeGrowthGlobal1)
	s.T().Logf("\tInside [%s, %s]: %s / %s", s.fullRangeLower, s.fullRangeUpper, fullRange.FeeGrowthInside0X128, fullRange.FeeGrowthInside1X128)
	s.T().Logf("\tInside [%s, %s]: %s / %s", s.narrowLower, s.narrowUpper, narrow.FeeGrowthInside0X128, narrow.FeeGrowthInside1X128)
}

func (s *StateViewSuite) Test17FeeGrowthInsideOutOfRangeIsZero() {
	s.Require().NotEmpty(s.swapBlock, "Test14 must record the swap block")

	inside, err := s.stateView.GetFeeGrowthInside(s.T().Context(), bind.NewCallOptions().WithBlock(s.swapBlock), NewGetFeeGrowthInsideRequest(s.stateViewContract, s.poolID, big.NewInt(600), big.NewInt(1200)))
	s.Require().NoError(err)
	s.Require().Zero(inside.FeeGrowthInside0X128.Sign(), "a range the price never entered has no fee growth inside")
	s.Require().Zero(inside.FeeGrowthInside1X128.Sign())

	s.T().Logf("StateView %s Test17FeeGrowthInsideOutOfRangeIsZero Result", s.stateViewContract)
	s.T().Logf("\tInside [600, 1200]: %s / %s", inside.FeeGrowthInside0X128, inside.FeeGrowthInside1X128)
}

func (s *StateViewSuite) Test18PositionFeeGrowthLastUnchangedUntilModified() {
	s.Require().NotEmpty(s.swapBlock, "Test14 must record the swap block")

	position, err := s.stateView.GetPositionInfo(s.T().Context(), bind.NewCallOptions().WithBlock(s.swapBlock), NewGetPositionInfoRequest(s.stateViewContract, s.poolID, s.poolModifyLiquidityTest, s.narrowLower, s.narrowUpper, s.narrowSalt))
	s.Require().NoError(err)
	s.Require().Zero(s.NarrowLiquidity.Cmp(position.Liquidity))
	s.Require().Zero(position.FeeGrowthInside0LastX128.Sign(), "feeGrowthInsideLast only updates when the position is modified")
	s.Require().Zero(position.FeeGrowthInside1LastX128.Sign())

	s.T().Logf("StateView %s Test18PositionFeeGrowthLastUnchangedUntilModified Result", s.stateViewContract)
	s.T().Logf("\tNarrow position: liquidity=%s feeGrowthInside0Last=%s feeGrowthInside1Last=%s", position.Liquidity, position.FeeGrowthInside0LastX128, position.FeeGrowthInside1LastX128)
}

func (s *StateViewSuite) Test19UnknownPoolReadsAsZero() {
	unknown := NewPoolKey(s.currency0, s.currency1, big.NewInt(500), big.NewInt(10), types.ZeroAddress())
	poolID, err := unknown.PoolID()
	s.Require().NoError(err)

	slot0, err := s.stateView.GetSlot0(s.T().Context(), nil, NewGetSlot0Request(s.stateViewContract, poolID))
	s.Require().NoError(err)
	s.Require().Zero(slot0.SqrtPriceX96.Sign())
	s.Require().Zero(slot0.Tick.Sign())
	s.Require().Zero(slot0.ProtocolFee.Sign())
	s.Require().Zero(slot0.LPFee.Sign())

	liquidity, err := s.stateView.GetLiquidity(s.T().Context(), nil, NewGetLiquidityRequest(s.stateViewContract, poolID))
	s.Require().NoError(err)
	s.Require().Zero(liquidity.Sign())

	globals, err := s.stateView.GetFeeGrowthGlobals(s.T().Context(), nil, NewGetFeeGrowthGlobalsRequest(s.stateViewContract, poolID))
	s.Require().NoError(err)
	s.Require().Zero(globals.FeeGrowthGlobal0.Sign())
	s.Require().Zero(globals.FeeGrowthGlobal1.Sign())

	s.T().Logf("StateView %s Test19UnknownPoolReadsAsZero Result", s.stateViewContract)
	s.T().Logf("\tPoolId: %x", poolID)
	s.T().Logf("\tGetSlot0: sqrtPriceX96=%s tick=%s lpFee=%s", slot0.SqrtPriceX96, slot0.Tick, slot0.LPFee)
	s.T().Logf("\tGetLiquidity: %s, GetFeeGrowthGlobals: %s / %s", liquidity, globals.FeeGrowthGlobal0, globals.FeeGrowthGlobal1)
}

func (s *StateViewSuite) Test20ReadsAgainstEOAFailToDecode() {
	slot0, slot0Err := s.stateView.GetSlot0(s.T().Context(), nil, NewGetSlot0Request(s.from, s.poolID))
	s.Require().Error(slot0Err)
	s.Require().Nil(slot0)
	s.Require().ErrorContains(slot0Err, "uniswapv4: getSlot0: ")

	poolManager, poolManagerErr := s.stateView.PoolManager(s.T().Context(), nil, NewPoolManagerRequest(s.from))
	s.Require().Error(poolManagerErr)
	s.Require().Nil(poolManager)
	s.Require().ErrorContains(poolManagerErr, "uniswapv4: stateViewPoolManager: ")

	s.T().Logf("StateView %s Test20ReadsAgainstEOAFailToDecode Result", s.stateViewContract)
	s.T().Logf("\tGetSlot0 on %s error: %v", s.from, slot0Err)
	s.T().Logf("\tPoolManager on %s error: %v", s.from, poolManagerErr)
}
