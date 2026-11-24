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

type RouterSuite struct {
	suite.Suite

	RPCURL                     string
	PrivateKeyHex              string
	PoolManagerHex             string
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
	LiquidityDelta             *big.Int
	DecreaseLiquidityDelta     *big.Int
	SwapAmountIn               *big.Int
	SwapAmountOut              *big.Int

	client                  rpc.Client
	deployer                *deploy.NicksDeployer
	signer                  core.Signer
	from                    *types.Address
	poolManagerContract     *types.Address
	poolSwapTest            *types.Address
	poolModifyLiquidityTest *types.Address
	multicall               *types.Address
	nicks                   *types.Address

	poolManager *PoolManager
	erc20       *erc20.ERC20

	currency0 *types.Address
	currency1 *types.Address
	key       *PoolKey

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestRouterSuite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)
	v4 := c.Addresses.Protocol.Uniswap.V4
	if v4.PoolManager == "" || v4.PoolSwapTest == "" || v4.PoolModifyLiquidityTest == "" {
		t.Skip("testenv: no Uniswap V4 PoolManager/PoolSwapTest/PoolModifyLiquidityTest configured for this network")
	}

	suite.Run(t, &RouterSuite{
		RPCURL:                     c.RPCURL,
		PrivateKeyHex:              c.ActorPrivateKeyHex,
		PoolManagerHex:             v4.PoolManager,
		PoolSwapTestHex:            v4.PoolSwapTest,
		PoolModifyLiquidityTestHex: v4.PoolModifyLiquidityTest,
		MulticallHex:               c.Addresses.Standard.Multicall,
		NicksFactoryHex:            c.Addresses.Standard.NicksFactory,
		NativeSymbol:               c.Constant.NativeSymbol,
		NativeDecimals:             c.Constant.NativeDecimals,
		Timeout:                    10 * time.Minute,
		ReceiptPollInterval:        c.Constant.ReceiptPollInterval,
		ReceiptTimeout:             c.Constant.ReceiptTimeout,
		ByteCode:                   testenv.ERC20ByteCodeHex,
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
	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.poolManager = New(s.client)
	s.erc20 = erc20.New(s.client)
	s.T().Logf("SetupSuite: client ready for %s", s.RPCURL)

	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())
	s.T().Logf("SetupSuite: signing as %s", s.from)

	s.poolManagerContract, err = types.NewAddressFromHex(s.PoolManagerHex)
	s.Require().NoError(err)
	s.poolSwapTest, err = types.NewAddressFromHex(s.PoolSwapTestHex)
	s.Require().NoError(err)
	s.poolModifyLiquidityTest, err = types.NewAddressFromHex(s.PoolModifyLiquidityTestHex)
	s.Require().NoError(err)
	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)
	s.nicks, err = types.NewAddressFromHex(s.NicksFactoryHex)
	s.Require().NoError(err)
	s.T().Logf("SetupSuite: PoolManager %s, PoolSwapTest %s, PoolModifyLiquidityTest %s", s.poolManagerContract, s.poolSwapTest, s.poolModifyLiquidityTest)

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	runtime := fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())

	names := [2]string{"evmkit v4 router test token A", "evmkit v4 router test token B"}
	symbols := [2]string{"EVRA", "EVRB"}
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
		s.T().Logf("SetupSuite: deployed token %s %s (nonce %d, tx %s, block %d)", symbols[i], token, deployTx.AddressNonce(), hash, receipt.BlockNumber)

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
	s.T().Logf("SetupSuite: sorted currency0=%s currency1=%s", s.currency0, s.currency1)

	s.key = NewPoolKey(s.currency0, s.currency1, big.NewInt(3000), big.NewInt(60), types.ZeroAddress())

	initTx, initHash, err := s.poolManager.Initialize(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewInitializeRequest(s.poolManagerContract, s.key, new(big.Int).Lsh(big.NewInt(1), 96)))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, initTx.MaxFee())
	s.mu.Unlock()

	initReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, initHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(initReceipt.Status, "initialize transaction reverted")
	s.T().Logf("SetupSuite: initialized pool currency0=%s currency1=%s (tx %s, block %d)", s.currency0, s.currency1, initHash, initReceipt.BlockNumber)
}

func (s *RouterSuite) TearDownSuite() {
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

			burnTx, burnHash, err := bind.Transact(ctx, s.client, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), tb.Token, calldata)
			s.Require().NoError(err)
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, burnTx.MaxFee())
			s.mu.Unlock()

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

		s.T().Logf("RouterSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}
}

func (s *RouterSuite) Test01ModifyLiquidity() {
	modifyLiquidityTestFn := abi.NewFunction("modifyLiquidity", abi.NewTypes(poolKeyTuple, liquidityParamsTuple, abi.Bytes, abi.Bool, abi.Bool), abi.NewTypes(abi.Bytes32))

	var salt [32]byte
	calldata, err := modifyLiquidityTestFn.EncodeCall(
		[]any{s.key.Currency0, s.key.Currency1, s.key.Fee, s.key.TickSpacing, s.key.Hooks},
		[]any{big.NewInt(-600), big.NewInt(600), s.LiquidityDelta, salt[:]},
		[]byte{}, false, false,
	)
	s.Require().NoError(err)

	var sentTx core.Transaction
	var hash *types.Hash
	s.Require().Eventually(func() bool {
		sentTx, hash, err = bind.Transact(s.T().Context(), s.client, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), s.poolModifyLiquidityTest, calldata)
		return err == nil || sentTx != nil
	}, 20*time.Second, s.ReceiptPollInterval, "pool still not initialized on the estimating node")
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test01ModifyLiquidity: sent PoolModifyLiquidityTest.modifyLiquidity (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "modifyLiquidity transaction reverted")

	modifies, err := ExtractModifyLiquidity(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(modifies, 1)
	s.Require().True(modifies[0].Contract.Equal(s.poolManagerContract))
	s.Require().True(modifies[0].Sender.Equal(s.poolModifyLiquidityTest))
	s.Require().Zero(s.LiquidityDelta.Cmp(modifies[0].LiquidityDelta))

	s.T().Logf("Router Test01ModifyLiquidity Result")
	s.T().Logf("\tModifyLiquidity event: id=%x sender=%s ticks=[%s, %s] delta=%s", modifies[0].Id, modifies[0].Sender, modifies[0].TickLower, modifies[0].TickUpper, modifies[0].LiquidityDelta)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *RouterSuite) Test02Swap() {
	poolSwapTestFn := abi.NewFunction("swap", abi.NewTypes(poolKeyTuple, swapAmountParamsTuple, abi.Tuple(abi.Bool, abi.Bool), abi.Bytes), abi.NewTypes(abi.Bytes32))
	amountSpecified := new(big.Int).Neg(s.SwapAmountIn)

	calldata, err := poolSwapTestFn.EncodeCall(
		[]any{s.key.Currency0, s.key.Currency1, s.key.Fee, s.key.TickSpacing, s.key.Hooks},
		[]any{true, amountSpecified, big.NewInt(4295128740)},
		[]any{false, false},
		[]byte{},
	)
	s.Require().NoError(err)

	sentTx, hash, err := bind.Transact(s.T().Context(), s.client, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), s.poolSwapTest, calldata)
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test02Swap: sent PoolSwapTest.swap (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swap transaction reverted")

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.poolManagerContract))
	s.Require().Zero(amountSpecified.Cmp(swaps[0].Amount0), "amount0 should equal the exact input specified")
	s.Require().Positive(swaps[0].Amount1.Sign())

	s.T().Logf("Router Test02Swap Result")
	s.T().Logf("\tAmountIn: %s", s.SwapAmountIn)
	s.T().Logf("\tAmount0/Amount1: %s / %s", swaps[0].Amount0, swaps[0].Amount1)
	s.T().Logf("\tSqrtPriceX96 after: %s (tick %s)", swaps[0].SqrtPriceX96, swaps[0].Tick)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *RouterSuite) Test03SwapOneForZero() {
	poolSwapTestFn := abi.NewFunction("swap", abi.NewTypes(poolKeyTuple, swapAmountParamsTuple, abi.Tuple(abi.Bool, abi.Bool), abi.Bytes), abi.NewTypes(abi.Bytes32))
	maxSqrtPriceX96Minus1, ok := new(big.Int).SetString("1461446703485210103287273052203988822378723970341", 10)
	s.Require().True(ok)
	amountSpecified := new(big.Int).Neg(s.SwapAmountIn)

	calldata, err := poolSwapTestFn.EncodeCall(
		[]any{s.key.Currency0, s.key.Currency1, s.key.Fee, s.key.TickSpacing, s.key.Hooks},
		[]any{false, amountSpecified, maxSqrtPriceX96Minus1},
		[]any{false, false},
		[]byte{},
	)
	s.Require().NoError(err)

	sentTx, hash, err := bind.Transact(s.T().Context(), s.client, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), s.poolSwapTest, calldata)
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test03SwapOneForZero: sent PoolSwapTest.swap (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swap transaction reverted")

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.poolManagerContract))
	s.Require().Zero(amountSpecified.Cmp(swaps[0].Amount1), "amount1 should equal the exact input specified")
	s.Require().Positive(swaps[0].Amount0.Sign())

	s.T().Logf("Router Test03SwapOneForZero Result")
	s.T().Logf("\tAmountIn: %s", s.SwapAmountIn)
	s.T().Logf("\tAmount0/Amount1: %s / %s", swaps[0].Amount0, swaps[0].Amount1)
	s.T().Logf("\tSqrtPriceX96 after: %s (tick %s)", swaps[0].SqrtPriceX96, swaps[0].Tick)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *RouterSuite) Test04SwapExactOutput() {
	poolSwapTestFn := abi.NewFunction("swap", abi.NewTypes(poolKeyTuple, swapAmountParamsTuple, abi.Tuple(abi.Bool, abi.Bool), abi.Bytes), abi.NewTypes(abi.Bytes32))

	calldata, err := poolSwapTestFn.EncodeCall(
		[]any{s.key.Currency0, s.key.Currency1, s.key.Fee, s.key.TickSpacing, s.key.Hooks},
		[]any{true, s.SwapAmountOut, big.NewInt(4295128740)},
		[]any{false, false},
		[]byte{},
	)
	s.Require().NoError(err)

	sentTx, hash, err := bind.Transact(s.T().Context(), s.client, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), s.poolSwapTest, calldata)
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test04SwapExactOutput: sent PoolSwapTest.swap (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swap transaction reverted")

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.poolManagerContract))
	s.Require().Zero(s.SwapAmountOut.Cmp(swaps[0].Amount1), "amount1 should equal the exact output requested")
	s.Require().Negative(swaps[0].Amount0.Sign())

	s.T().Logf("Router Test04SwapExactOutput Result")
	s.T().Logf("\tAmountOut: %s", s.SwapAmountOut)
	s.T().Logf("\tAmount0/Amount1: %s / %s", swaps[0].Amount0, swaps[0].Amount1)
	s.T().Logf("\tSqrtPriceX96 after: %s (tick %s)", swaps[0].SqrtPriceX96, swaps[0].Tick)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *RouterSuite) Test05DecreaseLiquidity() {
	modifyLiquidityTestFn := abi.NewFunction("modifyLiquidity", abi.NewTypes(poolKeyTuple, liquidityParamsTuple, abi.Bytes, abi.Bool, abi.Bool), abi.NewTypes(abi.Bytes32))

	var salt [32]byte
	calldata, err := modifyLiquidityTestFn.EncodeCall(
		[]any{s.key.Currency0, s.key.Currency1, s.key.Fee, s.key.TickSpacing, s.key.Hooks},
		[]any{big.NewInt(-600), big.NewInt(600), s.DecreaseLiquidityDelta, salt[:]},
		[]byte{}, false, false,
	)
	s.Require().NoError(err)

	sentTx, hash, err := bind.Transact(s.T().Context(), s.client, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), s.poolModifyLiquidityTest, calldata)
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test05DecreaseLiquidity: sent PoolModifyLiquidityTest.modifyLiquidity (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "modifyLiquidity transaction reverted")

	modifies, err := ExtractModifyLiquidity(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(modifies, 1)
	s.Require().Zero(s.DecreaseLiquidityDelta.Cmp(modifies[0].LiquidityDelta))

	s.T().Logf("Router Test05DecreaseLiquidity Result")
	s.T().Logf("\tModifyLiquidity delta: %s", modifies[0].LiquidityDelta)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *RouterSuite) Test06SwapAmountCannotBeZeroMapsThroughRouter() {
	poolSwapTestFn := abi.NewFunction("swap", abi.NewTypes(poolKeyTuple, swapAmountParamsTuple, abi.Tuple(abi.Bool, abi.Bool), abi.Bytes), abi.NewTypes(abi.Bytes32))

	calldata, err := poolSwapTestFn.EncodeCall(
		[]any{s.key.Currency0, s.key.Currency1, s.key.Fee, s.key.TickSpacing, s.key.Hooks},
		[]any{true, big.NewInt(0), big.NewInt(4295128740)},
		[]any{false, false},
		[]byte{},
	)
	s.Require().NoError(err)

	sentTx, hash, err := bind.Transact(s.T().Context(), s.client, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), s.poolSwapTest, calldata)
	s.Require().Error(err)
	s.Require().Nil(sentTx)
	s.Require().Nil(hash)

	wrapped := (&Router{}).wrapRevertError(err)
	s.Require().ErrorIs(wrapped, ErrSwapAmountCannotBeZero)
	s.Require().NotErrorIs(wrapped, ErrUnknown)

	s.T().Logf("Router Test06SwapAmountCannotBeZeroMapsThroughRouter Result")
	s.T().Logf("\tError: %s", wrapped)
}

func (s *RouterSuite) Test07PoolNotInitializedMapsThroughRouter() {
	uninitializedKey := NewPoolKey(s.currency0, s.currency1, big.NewInt(500), big.NewInt(10), types.ZeroAddress())
	poolSwapTestFn := abi.NewFunction("swap", abi.NewTypes(poolKeyTuple, swapAmountParamsTuple, abi.Tuple(abi.Bool, abi.Bool), abi.Bytes), abi.NewTypes(abi.Bytes32))

	calldata, err := poolSwapTestFn.EncodeCall(
		[]any{uninitializedKey.Currency0, uninitializedKey.Currency1, uninitializedKey.Fee, uninitializedKey.TickSpacing, uninitializedKey.Hooks},
		[]any{true, new(big.Int).Neg(s.SwapAmountIn), big.NewInt(4295128740)},
		[]any{false, false},
		[]byte{},
	)
	s.Require().NoError(err)

	sentTx, hash, err := bind.Transact(s.T().Context(), s.client, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), s.poolSwapTest, calldata)
	s.Require().Error(err)
	s.Require().Nil(sentTx)
	s.Require().Nil(hash)

	wrapped := (&Router{}).wrapRevertError(err)
	s.Require().ErrorIs(wrapped, ErrPoolNotInitialized)
	s.Require().NotErrorIs(wrapped, ErrUnknown)

	s.T().Logf("Router Test07PoolNotInitializedMapsThroughRouter Result")
	s.T().Logf("\tError: %s", wrapped)
}

func (s *RouterSuite) Test08SwapUpdatesBalances() {
	before, err := s.erc20.AddressBalances(s.T().Context(), nil, erc20.NewAddressBalancesRequest(s.multicall, s.from, []*types.Address{s.currency0, s.currency1}, false))
	s.Require().NoError(err)
	s.Require().Len(before.Tokens, 2)

	poolSwapTestFn := abi.NewFunction("swap", abi.NewTypes(poolKeyTuple, swapAmountParamsTuple, abi.Tuple(abi.Bool, abi.Bool), abi.Bytes), abi.NewTypes(abi.Bytes32))
	calldata, err := poolSwapTestFn.EncodeCall(
		[]any{s.key.Currency0, s.key.Currency1, s.key.Fee, s.key.TickSpacing, s.key.Hooks},
		[]any{true, new(big.Int).Neg(s.SwapAmountIn), big.NewInt(4295128740)},
		[]any{false, false},
		[]byte{},
	)
	s.Require().NoError(err)

	sentTx, hash, err := bind.Transact(s.T().Context(), s.client, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), s.poolSwapTest, calldata)
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test08SwapUpdatesBalances: sent PoolSwapTest.swap (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swap transaction reverted")

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))
	beforeAt := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber-1))

	var after *erc20.AddressBalances
	s.Require().Eventually(func() bool {
		after, err = s.erc20.AddressBalances(s.T().Context(), atBlock, erc20.NewAddressBalancesRequest(s.multicall, s.from, []*types.Address{s.currency0, s.currency1}, false))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "balances not readable at the receipt block")
	s.Require().Len(after.Tokens, 2)

	pinnedBefore, err := s.erc20.AddressBalances(s.T().Context(), beforeAt, erc20.NewAddressBalancesRequest(s.multicall, s.from, []*types.Address{s.currency0, s.currency1}, false))
	s.Require().NoError(err)

	want0 := new(big.Int).Add(pinnedBefore.Tokens[0].Amount, swaps[0].Amount0)
	want1 := new(big.Int).Add(pinnedBefore.Tokens[1].Amount, swaps[0].Amount1)
	s.Require().Zero(want0.Cmp(after.Tokens[0].Amount), "currency0 balance didn't move by exactly amount0")
	s.Require().Zero(want1.Cmp(after.Tokens[1].Amount), "currency1 balance didn't move by exactly amount1")

	s.T().Logf("Router Test08SwapUpdatesBalances Result")
	s.T().Logf("\tBalance0: %s -> %s (latest before send %s)", pinnedBefore.Tokens[0].Amount, after.Tokens[0].Amount, before.Tokens[0].Amount)
	s.T().Logf("\tBalance1: %s -> %s (latest before send %s)", pinnedBefore.Tokens[1].Amount, after.Tokens[1].Amount, before.Tokens[1].Amount)
	s.T().Logf("\tAmount0/Amount1: %s / %s", swaps[0].Amount0, swaps[0].Amount1)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}
