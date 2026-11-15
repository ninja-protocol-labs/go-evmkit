package v4

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ninja-protocol-labs/go-evmkit/standards/deploy"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-evmkit/standards/multicall3"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

type RouterSuite struct {
	suite.Suite

	Enable                     bool
	RPCURL                     string
	PrivateKeyHex              string
	PoolManagerHex             string
	PoolSwapTestHex            string
	PoolModifyLiquidityTestHex string
	MulticallHex               string
	NativeSymbol               string
	Interval                   int
	Timeout                    int
	ByteCode                   string
	Constructor                *abi.Function
	Salt                       *types.Hash
	MintValue                  *big.Int
	LiquidityDelta             *big.Int
	DecreaseLiquidityDelta     *big.Int
	SwapAmountIn               *big.Int
	SwapAmountOut              *big.Int

	ctx    context.Context
	cancel context.CancelFunc

	client    rpc.Client
	deployer  *deploy.NicksDeployer
	signer    core.Signer
	from      *types.Address
	multicall *types.Address

	poolManager             *PoolManager
	erc20                   *erc20.ERC20
	poolManagerContract     *types.Address
	poolSwapTest            *types.Address
	poolModifyLiquidityTest *types.Address

	currency0 *types.Address
	currency1 *types.Address
	key       *PoolKey

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestRouterSuite(t *testing.T) {
	suite.Run(t, &RouterSuite{
		Enable:                     false,
		RPCURL:                     "",
		PrivateKeyHex:              "",
		PoolManagerHex:             "",
		PoolSwapTestHex:            "",
		PoolModifyLiquidityTestHex: "",
		MulticallHex:               multicall3.DefaultMulticall3Address.String(),
		NativeSymbol:               "",
		Interval:                   1500,
		Timeout:                    60,
		ByteCode:                   erc20ByteCodeHex,
		Constructor:                abi.NewFunction("constructor", abi.Types{abi.String, abi.String, abi.Uint8}, nil),
		Salt:                       core.Keccak256([]byte("go-evmkit-v4-router")),
		MintValue:                  new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil),
		LiquidityDelta:             big.NewInt(1_000_000),
		DecreaseLiquidityDelta:     big.NewInt(-200_000),
		SwapAmountIn:               big.NewInt(1_000),
		SwapAmountOut:              big.NewInt(500),
	})
}

func (s *RouterSuite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Minute)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.poolManager = New(s.client)
	s.erc20 = erc20.New(s.client)

	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())

	s.poolManagerContract, err = types.NewAddressFromHex(s.PoolManagerHex)
	s.Require().NoError(err)
	s.poolSwapTest, err = types.NewAddressFromHex(s.PoolSwapTestHex)
	s.Require().NoError(err)
	s.poolModifyLiquidityTest, err = types.NewAddressFromHex(s.PoolModifyLiquidityTestHex)
	s.Require().NoError(err)
	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	runtime := fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())

	names := [2]string{"evmkit v4 router test token A", "evmkit v4 router test token B"}
	symbols := [2]string{"EVRA", "EVRB"}
	tokens := [2]*types.Address{}

	for i := range 2 {
		var salt [32]byte
		copy(salt[:], core.Keccak256(fmt.Appendf(nil, "%s-%d", runtime, i)).Bytes())

		deployment := deploy.NewDeterministicContractDeployment(
			deploy.NewContractDeployment(bytecode, s.Constructor.Inputs, []any{names[i], symbols[i], uint8(18)}),
			salt,
		)

		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(150)

		unsigned, err := s.deployer.CreateDeployment(s.ctx, builder, s.from, deploy.DefaultNicksMethodFactory, big.NewInt(0), deployment)
		s.Require().NoError(err)
		s.Require().NoError(unsigned.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, unsigned.MaxFee())
		s.mu.Unlock()

		hash, err := s.deployer.Deploy(s.ctx, unsigned)
		s.Require().NoError(err)

		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "deploy transaction reverted")

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

		tokenMintFn := abi.NewFunction("mint", abi.NewTypes(abi.Address, abi.Uint256), nil)
		calldata, err := tokenMintFn.EncodeCall(s.from, s.MintValue)
		s.Require().NoError(err)

		mintUnsigned, err := tx.NewDynamicFeeBuilder().Build(s.from, token, big.NewInt(0), calldata).Pack(s.ctx, s.client)
		s.Require().NoError(err)
		s.Require().NoError(mintUnsigned.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, mintUnsigned.MaxFee())
		s.mu.Unlock()

		mintRaw, err := mintUnsigned.EncodeRLP()
		s.Require().NoError(err)
		mintHash, err := tx.Broadcast(s.ctx, s.client, mintRaw)
		s.Require().NoError(err)

		mintReceipt, err := tx.WaitForReceipt(s.ctx, s.client, mintHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(mintReceipt.Status, "mint transaction reverted")

		for _, spender := range []*types.Address{s.poolModifyLiquidityTest, s.poolSwapTest} {
			approveBuilder := tx.NewDynamicFeeBuilder()
			unsignedApprove, err := s.erc20.CreateApprove(s.ctx, approveBuilder, token, s.from, spender, s.MintValue)
			s.Require().NoError(err)
			s.Require().NoError(unsignedApprove.Sign(s.signer))
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())
			s.mu.Unlock()

			approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
			s.Require().NoError(err)

			approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
			s.Require().NoError(err)
			s.Require().True(approveReceipt.Status, "approve transaction reverted")
		}

		s.T().Logf("SetupSuite: deployed, minted, and approved token %s %s (tx %s)", symbols[i], token, hash)
	}

	s.currency0, s.currency1 = tokens[0], tokens[1]
	if s.currency0.Big().Cmp(s.currency1.Big()) > 0 {
		s.currency0, s.currency1 = s.currency1, s.currency0
	}
	s.T().Logf("SetupSuite: sorted currency0=%s currency1=%s", s.currency0, s.currency1)

	s.key = NewPoolKey(s.currency0, s.currency1, big.NewInt(3000), big.NewInt(60), types.ZeroAddress())
	sqrtPriceX96 := new(big.Int).Lsh(big.NewInt(1), 96)

	initUnsigned, err := s.poolManager.CreateInitialize(s.ctx, tx.NewDynamicFeeBuilder(), s.poolManagerContract, s.from, s.key, sqrtPriceX96)
	s.Require().NoError(err)
	s.Require().NoError(initUnsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, initUnsigned.MaxFee())
	s.mu.Unlock()

	initHash, err := s.poolManager.Initialize(s.ctx, initUnsigned)
	s.Require().NoError(err)

	initReceipt, err := tx.WaitForReceipt(s.ctx, s.client, initHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(initReceipt.Status, "initialize transaction reverted")
	s.T().Logf("SetupSuite: initialized pool currency0=%s currency1=%s (tx %s)", s.currency0, s.currency1, initHash)
}

func (s *RouterSuite) TearDownSuite() {
	if s.totalFee != nil {
		base := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
		whole, frac := new(big.Int), new(big.Int)
		whole.DivMod(s.totalFee, base, frac)

		fracStr := strings.TrimRight(fmt.Sprintf("%018s", frac.String()), "0")
		decimal := whole.String()
		if fracStr != "" {
			decimal += "." + fracStr
		}

		s.T().Logf("RouterSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *RouterSuite) Test01ModifyLiquidity() {
	modifyLiquidityTestFn := abi.NewFunction(
		"modifyLiquidity",
		abi.NewTypes(poolKeyTuple, liquidityParamsTuple, abi.Bytes, abi.Bool, abi.Bool),
		abi.NewTypes(abi.Bytes32),
	)

	var salt [32]byte
	calldata, err := modifyLiquidityTestFn.EncodeCall(
		[]any{s.key.Currency0, s.key.Currency1, s.key.Fee, s.key.TickSpacing, s.key.Hooks},
		[]any{big.NewInt(-600), big.NewInt(600), s.LiquidityDelta, salt[:]},
		[]byte{}, false, false,
	)
	s.Require().NoError(err)

	unsigned, err := tx.NewDynamicFeeBuilder().Build(s.from, s.poolModifyLiquidityTest, big.NewInt(0), calldata).Pack(s.ctx, s.client)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()

	raw, err := unsigned.EncodeRLP()
	s.Require().NoError(err)
	hash, err := tx.Broadcast(s.ctx, s.client, raw)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "modifyLiquidity transaction reverted")

	s.T().Logf("RouterSuite Test01ModifyLiquidity Result")
	s.T().Logf("\tPoolModifyLiquidityTest: %s", s.poolModifyLiquidityTest)
	s.T().Logf("\tLiquidityDelta: %s", s.LiquidityDelta)
	s.T().Logf("\tTx: %s", hash)
}

func (s *RouterSuite) Test02Swap() {
	testSettingsTuple := abi.Tuple(abi.Bool, abi.Bool)
	poolSwapTestFn := abi.NewFunction(
		"swap",
		abi.NewTypes(poolKeyTuple, swapAmountParamsTuple, testSettingsTuple, abi.Bytes),
		abi.NewTypes(abi.Bytes32),
	)

	minSqrtPriceX96Plus1 := big.NewInt(4295128740)
	amountSpecified := new(big.Int).Neg(s.SwapAmountIn)

	calldata, err := poolSwapTestFn.EncodeCall(
		[]any{s.key.Currency0, s.key.Currency1, s.key.Fee, s.key.TickSpacing, s.key.Hooks},
		[]any{true, amountSpecified, minSqrtPriceX96Plus1},
		[]any{false, false},
		[]byte{},
	)
	s.Require().NoError(err)

	unsigned, err := tx.NewDynamicFeeBuilder().Build(s.from, s.poolSwapTest, big.NewInt(0), calldata).Pack(s.ctx, s.client)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()

	raw, err := unsigned.EncodeRLP()
	s.Require().NoError(err)
	hash, err := tx.Broadcast(s.ctx, s.client, raw)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swap transaction reverted")

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].PoolManager.Equal(s.poolManagerContract))

	s.T().Logf("RouterSuite Test02Swap Result")
	s.T().Logf("\tPoolSwapTest: %s", s.poolSwapTest)
	s.T().Logf("\tAmountIn: %s", s.SwapAmountIn)
	s.T().Logf("\tAmount0: %s", swaps[0].Amount0)
	s.T().Logf("\tAmount1: %s", swaps[0].Amount1)
	s.T().Logf("\tSqrtPriceX96 after: %s", swaps[0].SqrtPriceX96)
	s.T().Logf("\tTx: %s", hash)
}

func (s *RouterSuite) Test03SwapOneForZero() {
	testSettingsTuple := abi.Tuple(abi.Bool, abi.Bool)
	poolSwapTestFn := abi.NewFunction(
		"swap",
		abi.NewTypes(poolKeyTuple, swapAmountParamsTuple, testSettingsTuple, abi.Bytes),
		abi.NewTypes(abi.Bytes32),
	)

	maxSqrtPriceX96Minus1, _ := new(big.Int).SetString("1461446703485210103287273052203988822378723970341", 10)
	amountSpecified := new(big.Int).Neg(s.SwapAmountIn)

	calldata, err := poolSwapTestFn.EncodeCall(
		[]any{s.key.Currency0, s.key.Currency1, s.key.Fee, s.key.TickSpacing, s.key.Hooks},
		[]any{false, amountSpecified, maxSqrtPriceX96Minus1},
		[]any{false, false},
		[]byte{},
	)
	s.Require().NoError(err)

	unsigned, err := tx.NewDynamicFeeBuilder().Build(s.from, s.poolSwapTest, big.NewInt(0), calldata).Pack(s.ctx, s.client)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()

	raw, err := unsigned.EncodeRLP()
	s.Require().NoError(err)
	hash, err := tx.Broadcast(s.ctx, s.client, raw)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swap transaction reverted")

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].PoolManager.Equal(s.poolManagerContract))
	s.Require().Zero(amountSpecified.Cmp(swaps[0].Amount1), "amount1 should equal the exact input specified")

	s.T().Logf("RouterSuite Test03SwapOneForZero Result")
	s.T().Logf("\tAmountIn: %s", s.SwapAmountIn)
	s.T().Logf("\tAmount0: %s", swaps[0].Amount0)
	s.T().Logf("\tAmount1: %s", swaps[0].Amount1)
	s.T().Logf("\tSqrtPriceX96 after: %s", swaps[0].SqrtPriceX96)
	s.T().Logf("\tTx: %s", hash)
}

func (s *RouterSuite) Test04SwapExactOutput() {
	testSettingsTuple := abi.Tuple(abi.Bool, abi.Bool)
	poolSwapTestFn := abi.NewFunction(
		"swap",
		abi.NewTypes(poolKeyTuple, swapAmountParamsTuple, testSettingsTuple, abi.Bytes),
		abi.NewTypes(abi.Bytes32),
	)

	minSqrtPriceX96Plus1 := big.NewInt(4295128740)

	calldata, err := poolSwapTestFn.EncodeCall(
		[]any{s.key.Currency0, s.key.Currency1, s.key.Fee, s.key.TickSpacing, s.key.Hooks},
		[]any{true, s.SwapAmountOut, minSqrtPriceX96Plus1},
		[]any{false, false},
		[]byte{},
	)
	s.Require().NoError(err)

	unsigned, err := tx.NewDynamicFeeBuilder().Build(s.from, s.poolSwapTest, big.NewInt(0), calldata).Pack(s.ctx, s.client)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()

	raw, err := unsigned.EncodeRLP()
	s.Require().NoError(err)
	hash, err := tx.Broadcast(s.ctx, s.client, raw)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swap transaction reverted")

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].PoolManager.Equal(s.poolManagerContract))
	s.Require().Zero(s.SwapAmountOut.Cmp(swaps[0].Amount1), "amount1 (the output currency's credit) should equal the exact output requested")

	s.T().Logf("RouterSuite Test04SwapExactOutput Result")
	s.T().Logf("\tAmountOut: %s", s.SwapAmountOut)
	s.T().Logf("\tAmount0: %s", swaps[0].Amount0)
	s.T().Logf("\tAmount1: %s", swaps[0].Amount1)
	s.T().Logf("\tSqrtPriceX96 after: %s", swaps[0].SqrtPriceX96)
	s.T().Logf("\tTx: %s", hash)
}

func (s *RouterSuite) Test05DecreaseLiquidity() {
	modifyLiquidityTestFn := abi.NewFunction(
		"modifyLiquidity",
		abi.NewTypes(poolKeyTuple, liquidityParamsTuple, abi.Bytes, abi.Bool, abi.Bool),
		abi.NewTypes(abi.Bytes32),
	)

	var salt [32]byte
	calldata, err := modifyLiquidityTestFn.EncodeCall(
		[]any{s.key.Currency0, s.key.Currency1, s.key.Fee, s.key.TickSpacing, s.key.Hooks},
		[]any{big.NewInt(-600), big.NewInt(600), s.DecreaseLiquidityDelta, salt[:]},
		[]byte{}, false, false,
	)
	s.Require().NoError(err)

	unsigned, err := tx.NewDynamicFeeBuilder().Build(s.from, s.poolModifyLiquidityTest, big.NewInt(0), calldata).Pack(s.ctx, s.client)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()

	raw, err := unsigned.EncodeRLP()
	s.Require().NoError(err)
	hash, err := tx.Broadcast(s.ctx, s.client, raw)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "modifyLiquidity transaction reverted")

	modifies, err := ExtractModifyLiquidity(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(modifies, 1)
	s.Require().Zero(s.DecreaseLiquidityDelta.Cmp(modifies[0].LiquidityDelta))

	s.T().Logf("RouterSuite Test05DecreaseLiquidity Result")
	s.T().Logf("\tPoolModifyLiquidityTest: %s", s.poolModifyLiquidityTest)
	s.T().Logf("\tLiquidityDelta: %s", s.DecreaseLiquidityDelta)
	s.T().Logf("\tTx: %s", hash)
}

func (s *RouterSuite) Test06SwapAmountCannotBeZero() {
	testSettingsTuple := abi.Tuple(abi.Bool, abi.Bool)
	poolSwapTestFn := abi.NewFunction(
		"swap",
		abi.NewTypes(poolKeyTuple, swapAmountParamsTuple, testSettingsTuple, abi.Bytes),
		abi.NewTypes(abi.Bytes32),
	)

	minSqrtPriceX96Plus1 := big.NewInt(4295128740)

	calldata, err := poolSwapTestFn.EncodeCall(
		[]any{s.key.Currency0, s.key.Currency1, s.key.Fee, s.key.TickSpacing, s.key.Hooks},
		[]any{true, big.NewInt(0), minSqrtPriceX96Plus1},
		[]any{false, false},
		[]byte{},
	)
	s.Require().NoError(err)

	_, err = tx.NewDynamicFeeBuilder().Build(s.from, s.poolSwapTest, big.NewInt(0), calldata).Pack(s.ctx, s.client)
	s.Require().Error(err)
	s.Require().ErrorIs(s.poolManager.wrapRevertError(err), ErrSwapAmountCannotBeZero)

	s.T().Logf("RouterSuite Test06SwapAmountCannotBeZero Result")
	s.T().Logf("\tError: %s", err)
}

func (s *RouterSuite) Test07PoolNotInitialized() {
	uninitializedKey := NewPoolKey(s.currency0, s.currency1, big.NewInt(500), big.NewInt(10), types.ZeroAddress())

	testSettingsTuple := abi.Tuple(abi.Bool, abi.Bool)
	poolSwapTestFn := abi.NewFunction(
		"swap",
		abi.NewTypes(poolKeyTuple, swapAmountParamsTuple, testSettingsTuple, abi.Bytes),
		abi.NewTypes(abi.Bytes32),
	)

	minSqrtPriceX96Plus1 := big.NewInt(4295128740)

	calldata, err := poolSwapTestFn.EncodeCall(
		[]any{uninitializedKey.Currency0, uninitializedKey.Currency1, uninitializedKey.Fee, uninitializedKey.TickSpacing, uninitializedKey.Hooks},
		[]any{true, new(big.Int).Neg(s.SwapAmountIn), minSqrtPriceX96Plus1},
		[]any{false, false},
		[]byte{},
	)
	s.Require().NoError(err)

	_, err = tx.NewDynamicFeeBuilder().Build(s.from, s.poolSwapTest, big.NewInt(0), calldata).Pack(s.ctx, s.client)
	s.Require().Error(err)
	s.Require().ErrorIs(s.poolManager.wrapRevertError(err), ErrPoolNotInitialized)

	s.T().Logf("RouterSuite Test07PoolNotInitialized Result")
	s.T().Logf("\tError: %s", err)
}

func (s *RouterSuite) Test08SwapUpdatesBalances() {
	before, err := s.erc20.AddressBalances(s.ctx, s.multicall, s.from, []*types.Address{s.currency0, s.currency1}, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(before.Tokens, 2)
	balanceBefore0 := before.Tokens[0].Amount
	balanceBefore1 := before.Tokens[1].Amount

	testSettingsTuple := abi.Tuple(abi.Bool, abi.Bool)
	poolSwapTestFn := abi.NewFunction(
		"swap",
		abi.NewTypes(poolKeyTuple, swapAmountParamsTuple, testSettingsTuple, abi.Bytes),
		abi.NewTypes(abi.Bytes32),
	)

	minSqrtPriceX96Plus1 := big.NewInt(4295128740)
	amountSpecified := new(big.Int).Neg(s.SwapAmountIn)

	calldata, err := poolSwapTestFn.EncodeCall(
		[]any{s.key.Currency0, s.key.Currency1, s.key.Fee, s.key.TickSpacing, s.key.Hooks},
		[]any{true, amountSpecified, minSqrtPriceX96Plus1},
		[]any{false, false},
		[]byte{},
	)
	s.Require().NoError(err)

	unsigned, err := tx.NewDynamicFeeBuilder().Build(s.from, s.poolSwapTest, big.NewInt(0), calldata).Pack(s.ctx, s.client)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()

	raw, err := unsigned.EncodeRLP()
	s.Require().NoError(err)
	hash, err := tx.Broadcast(s.ctx, s.client, raw)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swap transaction reverted")

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)

	var after *erc20.AddressBalances
	s.Require().Eventually(func() bool {
		after, err = s.erc20.AddressBalances(s.ctx, s.multicall, s.from, []*types.Address{s.currency0, s.currency1}, false, rpc.BlockTagLatest)
		return err == nil
	}, 10*time.Second, 500*time.Millisecond, "balances not readable after swap")
	s.Require().Len(after.Tokens, 2)

	wantBalance0 := new(big.Int).Add(balanceBefore0, swaps[0].Amount0)
	wantBalance1 := new(big.Int).Add(balanceBefore1, swaps[0].Amount1)
	s.Require().Zero(wantBalance0.Cmp(after.Tokens[0].Amount), "currency0 balance didn't move by exactly amount0")
	s.Require().Zero(wantBalance1.Cmp(after.Tokens[1].Amount), "currency1 balance didn't move by exactly amount1")

	s.T().Logf("RouterSuite Test08SwapUpdatesBalances Result")
	s.T().Logf("\tBalance0 before/after: %s -> %s", balanceBefore0, after.Tokens[0].Amount)
	s.T().Logf("\tBalance1 before/after: %s -> %s", balanceBefore1, after.Tokens[1].Amount)
	s.T().Logf("\tTx: %s", hash)
}
