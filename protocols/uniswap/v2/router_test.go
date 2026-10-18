package v2

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/protocols/weth"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/deploy"
	"github.com/ninja-protocol-labs/go-evmkit/standards/eip712"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc2612"
	"github.com/ninja-protocol-labs/go-evmkit/standards/multicall3"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

type RouterLifeCycleSuite struct {
	suite.Suite

	Enable          bool
	RPCURL          string
	PrivateKeyHex   string
	RouterHex       string
	MulticallHex    string
	Interval        int
	Timeout         int
	DeadlineSeconds int
	LimitBuffer     uint64
	PriceBuffer     uint64
	ByteCode        string
	Constructor     *abi.Function
	MintValue       string
	LiquidityAmount string
	SwapAmount      string
	NativeValue     string
	NativeSymbol    string
	Salt            *types.Hash

	ctx    context.Context
	cancel context.CancelFunc

	client   rpc.Client
	signer   core.Signer
	deployer *deploy.NicksDeployer

	from      *types.Address
	router    *types.Address
	factory   *types.Address
	multicall *types.Address
	wethAddr  *types.Address
	chainID   *big.Int
	token0    *types.Address
	token1    *types.Address
	token2    *types.Address
	pair      *types.Address
	pair12    *types.Address
	pairETH   *types.Address

	mintValue       *big.Int
	liquidityAmount *big.Int
	swapAmount      *big.Int
	nativeValue     *big.Int

	v2Router  *Router
	v2Factory *Factory
	v2Pair    *Pair
	w         *weth.WETH
	erc20     *erc20.ERC20

	totalFee *big.Int
}

func TestRouterSuite(t *testing.T) {
	suite.Run(t, &RouterLifeCycleSuite{
		Enable:          false,
		RPCURL:          "",
		RouterHex:       "",
		PrivateKeyHex:   "",
		MulticallHex:    multicall3.DefaultMulticall3Address.String(),
		Interval:        200,
		Timeout:         60,
		DeadlineSeconds: 600,
		LimitBuffer:     300,
		PriceBuffer:     150,
		ByteCode:        erc20ByteCodeHex,
		Constructor:     abi.NewFunction("constructor", abi.NewTypes(abi.String, abi.String, abi.Uint8), nil),
		MintValue:       "1000000000000000000000000",
		LiquidityAmount: "1000000000000000000000",
		SwapAmount:      "1000000000000000000",
		NativeValue:     "1000000000000000",
		NativeSymbol:    "",
		Salt:            core.Keccak256([]byte("go-evmkit")),
	})
}

func (s *RouterLifeCycleSuite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), time.Duration(s.Timeout)*time.Minute)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.v2Router = NewRouter(s.client)
	s.v2Factory = NewFactory(s.client)
	s.v2Pair = NewPair(s.client)
	s.w = weth.New(s.client)
	s.erc20 = erc20.New(s.client)

	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())

	s.router, err = types.NewAddressFromHex(s.RouterHex)
	s.Require().NoError(err)

	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)

	mintValue, ok := new(big.Int).SetString(s.MintValue, 10)
	s.Require().True(ok)
	s.mintValue = mintValue

	liquidityAmount, ok := new(big.Int).SetString(s.LiquidityAmount, 10)
	s.Require().True(ok)
	s.liquidityAmount = liquidityAmount

	swapAmount, ok := new(big.Int).SetString(s.SwapAmount, 10)
	s.Require().True(ok)
	s.swapAmount = swapAmount

	nativeValue, ok := new(big.Int).SetString(s.NativeValue, 10)
	s.Require().True(ok)
	s.nativeValue = nativeValue

	var chainIDHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHChainID("chainId", &chainIDHex)))
	chainID, ok := new(big.Int).SetString(strings.TrimPrefix(chainIDHex, "0x"), 16)
	s.Require().True(ok)
	s.chainID = chainID

	s.wethAddr, err = s.v2Router.WETH(s.ctx, s.router, rpc.BlockTagLatest)
	s.Require().NoError(err)

	s.factory, err = s.v2Router.Factory(s.ctx, s.router, rpc.BlockTagLatest)
	s.Require().NoError(err)

	s.T().Logf("SetupSuite: router %s -> factory=%s WETH=%s", s.router, s.factory, s.wethAddr)

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	runtime := fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())
	names := [3]string{"evmkit test router token A", "evmkit test router token B", "evmkit test router token C"}
	symbols := [3]string{"EVKRA", "EVKRB", "EVKRC"}
	tokens := [3]*types.Address{}

	tokenMintFn := abi.NewFunction("mint", abi.NewTypes(abi.Address, abi.Uint256), nil)

	for i := range 3 {
		var salt [32]byte
		copy(salt[:], core.Keccak256(fmt.Appendf(nil, "%s-%d", runtime, i)).Bytes())

		deployment := deploy.NewDeterministicContractDeployment(
			deploy.NewContractDeployment(bytecode, s.Constructor.Inputs, []any{names[i], symbols[i], uint8(18)}),
			salt,
		)

		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(s.PriceBuffer)
		builder.LimitBuffer(s.LimitBuffer)

		unsigned, err := s.deployer.CreateDeployment(s.ctx, builder, s.from, deploy.DefaultNicksMethodFactory, big.NewInt(0), deployment)
		s.Require().NoError(err)
		s.Require().NoError(unsigned.Sign(s.signer))
		s.totalFee.Add(s.totalFee, unsigned.MaxFee())

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
		}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "no code visible at deployed token %s", token)

		s.T().Logf("SetupSuite: deployed token %s %s (nonce %d, tx %s)", symbols[i], token, unsigned.AddressNonce(), hash)

		calldata, err := tokenMintFn.EncodeCall(s.from, s.mintValue)
		s.Require().NoError(err)

		var unsignedMint core.Transaction
		s.Require().Eventually(func() bool {
			builder = tx.NewDynamicFeeBuilder()
			builder.PriceBuffer(s.PriceBuffer)
			builder.LimitBuffer(s.LimitBuffer)
			u, err := builder.Build(s.from, token, big.NewInt(0), calldata).Pack(s.ctx, s.client)
			if err != nil || u.Gas() < 40000 {
				return false
			}
			unsignedMint = u
			return true
		}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "mint gas estimate looked like a no-code call")
		s.Require().NoError(unsignedMint.Sign(s.signer))
		s.totalFee.Add(s.totalFee, unsignedMint.MaxFee())

		raw, err := unsignedMint.EncodeRLP()
		s.Require().NoError(err)
		mintHash, err := tx.Broadcast(s.ctx, s.client, raw)
		s.Require().NoError(err)

		mintReceipt, err := tx.WaitForReceipt(s.ctx, s.client, mintHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(mintReceipt.Status, "mint transaction reverted")

		s.T().Logf("SetupSuite: minted %s of %s to owner %s (tx %s)", s.mintValue, token, s.from, mintHash)
	}

	s.token0, s.token1 = tokens[0], tokens[1]
	if s.token0.Big().Cmp(s.token1.Big()) > 0 {
		s.token0, s.token1 = s.token1, s.token0
	}
	s.token2 = tokens[2]
}

func (s *RouterLifeCycleSuite) TearDownSuite() {
	if s.token0 != nil && s.token1 != nil && s.token2 != nil {
		tokenBurnFn := abi.NewFunction("burn", abi.NewTypes(abi.Address, abi.Uint256), nil)

		for _, token := range []*types.Address{s.token0, s.token1, s.token2} {
			balance, err := s.erc20.BalanceOf(s.ctx, token, s.from, rpc.BlockTagLatest)
			s.Require().NoError(err)
			if balance.Sign() == 0 {
				continue
			}

			calldata, err := tokenBurnFn.EncodeCall(s.from, balance)
			s.Require().NoError(err)

			builder := tx.NewDynamicFeeBuilder()
			builder.PriceBuffer(s.PriceBuffer)
			builder.LimitBuffer(s.LimitBuffer)
			unsignedBurn, err := builder.Build(s.from, token, big.NewInt(0), calldata).Pack(s.ctx, s.client)
			s.Require().NoError(err)
			s.Require().NoError(unsignedBurn.Sign(s.signer))
			s.totalFee.Add(s.totalFee, unsignedBurn.MaxFee())

			raw, err := unsignedBurn.EncodeRLP()
			s.Require().NoError(err)
			burnHash, err := tx.Broadcast(s.ctx, s.client, raw)
			s.Require().NoError(err)

			burnReceipt, err := tx.WaitForReceipt(s.ctx, s.client, burnHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
			s.Require().NoError(err)
			s.Require().True(burnReceipt.Status, "burn transaction reverted")

			s.T().Logf("TearDownSuite: burned %s of %s from %s (tx %s)", balance, token, s.from, burnHash)
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

		s.T().Logf("RouterSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *RouterLifeCycleSuite) Test01ReadMethods() {
	factory, err := s.v2Router.Factory(s.ctx, s.router, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().False(factory.IsZero())
	s.Require().True(factory.Equal(s.factory))
	s.T().Logf("Test01ReadMethods: factory() = %s", factory)

	w, err := s.v2Router.WETH(s.ctx, s.router, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().False(w.IsZero())
	s.Require().True(w.Equal(s.wethAddr))
	s.T().Logf("Test01ReadMethods: WETH() = %s", w)
}

func (s *RouterLifeCycleSuite) Test02AddLiquidity() {
	s.T().Logf("Test02AddLiquidity: token0=%s token1=%s liquidityAmount=%s", s.token0, s.token1, s.liquidityAmount)

	for _, token := range []*types.Address{s.token0, s.token1} {
		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(s.PriceBuffer)
		builder.LimitBuffer(s.LimitBuffer)
		unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, token, s.from, s.router, s.liquidityAmount)
		s.Require().NoError(err)
		s.Require().NoError(unsignedApprove.Sign(s.signer))
		s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())

		approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
		s.Require().NoError(err)
		s.T().Logf("Test02AddLiquidity: approve %s to spend %s of %s, broadcast tx %s", s.router, s.liquidityAmount, token, approveHash)

		approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(approveReceipt.Status, "approve transaction reverted")
		var allowance *big.Int
		s.Require().Eventually(func() bool {
			allowance, err = s.erc20.Allowance(s.ctx, token, s.from, s.router, rpc.BlockTagLatest)
			return err == nil && s.liquidityAmount.Cmp(allowance) == 0
		}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
		s.T().Logf("Test02AddLiquidity: approve of %s confirmed in block %d", token, approveReceipt.BlockNumber)
	}

	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
	s.T().Logf("Test02AddLiquidity: deadline=%s", deadline)

	predicted, err := s.v2Router.CallAddLiquidity(s.ctx, s.router, s.from, s.token0, s.token1, s.liquidityAmount, s.liquidityAmount, big.NewInt(0), big.NewInt(0), s.from, deadline, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(s.liquidityAmount.Cmp(predicted.AmountA))
	s.Require().Zero(s.liquidityAmount.Cmp(predicted.AmountB))
	s.Require().True(predicted.Liquidity.Sign() > 0)
	s.T().Logf("Test02AddLiquidity: CallAddLiquidity predicted amountA=%s amountB=%s liquidity=%s", predicted.AmountA, predicted.AmountB, predicted.Liquidity)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.v2Router.CreateAddLiquidity(s.ctx, builder, s.router, s.from, s.token0, s.token1, s.liquidityAmount, s.liquidityAmount, big.NewInt(0), big.NewInt(0), s.from, deadline)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.v2Router.AddLiquidity(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test02AddLiquidity: broadcast addLiquidity tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "addLiquidity transaction reverted")
	s.T().Logf("Test02AddLiquidity: addLiquidity confirmed in block %d", receipt.BlockNumber)

	created, err := ExtractPairCreated(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(created, 1, "addLiquidity should have created the pair itself, via the factory")
	s.pair = created[0].Pair
	s.T().Logf("Test02AddLiquidity: factory created pair %s for (%s, %s), allPairsLength=%s", s.pair, s.token0, s.token1, created[0].AllPairsLength)

	mints, err := ExtractMint(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(mints, 1)
	s.T().Logf("Test02AddLiquidity: Mint sender=%s amount0=%s amount1=%s", mints[0].Sender, mints[0].Amount0, mints[0].Amount1)

	var reserves *Reserves
	s.Require().Eventually(func() bool {
		reserves, err = s.v2Pair.GetReserves(s.ctx, s.pair, rpc.BlockTagLatest)
		return err == nil && reserves.Reserve0.Sign() > 0 && reserves.Reserve1.Sign() > 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "reserves not visible after addLiquidity")
	s.Require().Zero(s.liquidityAmount.Cmp(reserves.Reserve0))
	s.Require().Zero(s.liquidityAmount.Cmp(reserves.Reserve1))
	s.T().Logf("Test02AddLiquidity: pair %s reserves reserve0=%s reserve1=%s", s.pair, reserves.Reserve0, reserves.Reserve1)

	liquidityBalance, err := s.v2Pair.BalanceOf(s.ctx, s.pair, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test02AddLiquidity: %s's LP balance of pair %s: %s", s.from, s.pair, liquidityBalance)
}

func (s *RouterLifeCycleSuite) Test03GetAmounts() {
	path := []*types.Address{s.token0, s.token1}
	s.T().Logf("Test03GetAmounts: path=[%s, %s] swapAmount=%s", s.token0, s.token1, s.swapAmount)

	reserves, err := s.v2Pair.GetReserves(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test03GetAmounts: pair %s reserves reserve0=%s reserve1=%s", s.pair, reserves.Reserve0, reserves.Reserve1)

	amountsOut, err := s.v2Router.GetAmountsOut(s.ctx, s.router, s.swapAmount, path, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(amountsOut, 2)
	s.Require().Zero(s.swapAmount.Cmp(amountsOut[0]))
	s.Require().True(amountsOut[1].Sign() > 0)
	s.T().Logf("Test03GetAmounts: getAmountsOut(%s, [%s, %s]) = %v", s.swapAmount, s.token0, s.token1, amountsOut)

	amountsIn, err := s.v2Router.GetAmountsIn(s.ctx, s.router, amountsOut[1], path, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(amountsIn, 2)
	s.Require().Zero(amountsOut[1].Cmp(amountsIn[1]))
	s.Require().True(amountsIn[0].Sign() > 0)
	s.T().Logf("Test03GetAmounts: getAmountsIn(%s, [%s, %s]) = %v", amountsOut[1], s.token0, s.token1, amountsIn)

	quote, err := s.v2Router.Quote(s.ctx, s.router, s.swapAmount, reserves.Reserve0, reserves.Reserve1, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test03GetAmounts: quote(%s, reserve0=%s, reserve1=%s) = %s (no fee, vs. amountsOut's fee-adjusted %s)", s.swapAmount, reserves.Reserve0, reserves.Reserve1, quote, amountsOut[1])
}

func (s *RouterLifeCycleSuite) Test04SwapExactTokensForTokens() {
	path := []*types.Address{s.token0, s.token1}
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
	s.T().Logf("Test04SwapExactTokensForTokens: path=[%s, %s] swapAmount=%s deadline=%s", s.token0, s.token1, s.swapAmount, deadline)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.token0, s.from, s.router, s.swapAmount)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)
	s.T().Logf("Test04SwapExactTokensForTokens: approve %s to spend %s of %s, broadcast tx %s", s.router, s.swapAmount, s.token0, approveHash)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, s.token0, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && s.swapAmount.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	s.T().Logf("Test04SwapExactTokensForTokens: approve confirmed in block %d", approveReceipt.BlockNumber)

	predicted, err := s.v2Router.CallSwapExactTokensForTokens(s.ctx, s.router, s.from, s.swapAmount, big.NewInt(0), path, s.from, deadline, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(predicted, 2)
	s.T().Logf("Test04SwapExactTokensForTokens: CallSwapExactTokensForTokens predicted amounts=%v", predicted)

	beforeIn, err := s.erc20.BalanceOf(s.ctx, s.token0, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	beforeOut, err := s.erc20.BalanceOf(s.ctx, s.token1, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test04SwapExactTokensForTokens: balances before: token0=%s token1=%s", beforeIn, beforeOut)

	builder = tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.v2Router.CreateSwapExactTokensForTokens(s.ctx, builder, s.router, s.from, s.swapAmount, predicted[1], path, s.from, deadline)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.v2Router.SwapExactTokensForTokens(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test04SwapExactTokensForTokens: broadcast swapExactTokensForTokens tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swapExactTokensForTokens transaction reverted")
	s.T().Logf("Test04SwapExactTokensForTokens: swapExactTokensForTokens confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().Zero(predicted[1].Cmp(swaps[0].Amount1Out))
	s.T().Logf("Test04SwapExactTokensForTokens: Swap sender=%s to=%s amount0In=%s amount1In=%s amount0Out=%s amount1Out=%s",
		swaps[0].Sender, swaps[0].To, swaps[0].Amount0In, swaps[0].Amount1In, swaps[0].Amount0Out, swaps[0].Amount1Out)

	transfers, err := erc20.ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	for _, transfer := range transfers {
		s.T().Logf("Test04SwapExactTokensForTokens: Transfer contract=%s from=%s to=%s value=%s", transfer.Contract, transfer.From, transfer.To, transfer.Value)
	}

	var afterIn, afterOut *big.Int
	s.Require().Eventually(func() bool {
		afterIn, err = s.erc20.BalanceOf(s.ctx, s.token0, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		afterOut, err = s.erc20.BalanceOf(s.ctx, s.token1, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		return new(big.Int).Sub(beforeIn, afterIn).Cmp(s.swapAmount) == 0 &&
			new(big.Int).Sub(afterOut, beforeOut).Cmp(predicted[1]) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "token1 balance not visible after swap")
	s.T().Logf("Test04SwapExactTokensForTokens: balances after: token0=%s (%s -> %s) token1=%s (%s -> %s)",
		afterIn, beforeIn, afterIn, afterOut, beforeOut, afterOut)
}

func (s *RouterLifeCycleSuite) Test05RemoveLiquidity() {
	liquidity, err := s.v2Pair.BalanceOf(s.ctx, s.pair, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(liquidity.Sign() > 0, "no LP balance left to remove")
	s.T().Logf("Test05RemoveLiquidity: pair=%s LP balance=%s", s.pair, liquidity)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.pair, s.from, s.router, liquidity)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)
	s.T().Logf("Test05RemoveLiquidity: approve %s to spend %s of LP token %s, broadcast tx %s", s.router, liquidity, s.pair, approveHash)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, s.pair, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && liquidity.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	s.T().Logf("Test05RemoveLiquidity: approve confirmed in block %d", approveReceipt.BlockNumber)

	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	predicted, err := s.v2Router.CallRemoveLiquidity(s.ctx, s.router, s.from, s.token0, s.token1, liquidity, big.NewInt(0), big.NewInt(0), s.from, deadline, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(predicted.AmountA.Sign() > 0)
	s.Require().True(predicted.AmountB.Sign() > 0)
	s.T().Logf("Test05RemoveLiquidity: CallRemoveLiquidity predicted amountA=%s amountB=%s", predicted.AmountA, predicted.AmountB)

	beforeA, err := s.erc20.BalanceOf(s.ctx, s.token0, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	beforeB, err := s.erc20.BalanceOf(s.ctx, s.token1, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test05RemoveLiquidity: balances before: token0=%s token1=%s", beforeA, beforeB)

	builder = tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.v2Router.CreateRemoveLiquidity(s.ctx, builder, s.router, s.from, s.token0, s.token1, liquidity, predicted.AmountA, predicted.AmountB, s.from, deadline)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.v2Router.RemoveLiquidity(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test05RemoveLiquidity: broadcast removeLiquidity tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "removeLiquidity transaction reverted")
	s.T().Logf("Test05RemoveLiquidity: removeLiquidity confirmed in block %d", receipt.BlockNumber)

	burns, err := ExtractBurn(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(burns, 1)
	s.Require().Zero(predicted.AmountA.Cmp(burns[0].Amount0))
	s.Require().Zero(predicted.AmountB.Cmp(burns[0].Amount1))
	s.T().Logf("Test05RemoveLiquidity: Burn sender=%s amount0=%s amount1=%s to=%s", burns[0].Sender, burns[0].Amount0, burns[0].Amount1, burns[0].To)

	transfers, err := erc20.ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	for _, transfer := range transfers {
		s.T().Logf("Test05RemoveLiquidity: Transfer contract=%s from=%s to=%s value=%s", transfer.Contract, transfer.From, transfer.To, transfer.Value)
	}

	var afterA, afterB, remainingLiquidity *big.Int
	s.Require().Eventually(func() bool {
		afterA, err = s.erc20.BalanceOf(s.ctx, s.token0, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		afterB, err = s.erc20.BalanceOf(s.ctx, s.token1, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		remainingLiquidity, err = s.v2Pair.BalanceOf(s.ctx, s.pair, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		return new(big.Int).Sub(afterA, beforeA).Cmp(predicted.AmountA) == 0 &&
			new(big.Int).Sub(afterB, beforeB).Cmp(predicted.AmountB) == 0 &&
			remainingLiquidity.Sign() == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not visible after removeLiquidity")
	s.T().Logf("Test05RemoveLiquidity: balances after: token0=%s (%s -> %s) token1=%s (%s -> %s), remaining LP balance=%s",
		afterA, beforeA, afterA, afterB, beforeB, afterB, remainingLiquidity)

	reserves, err := s.v2Pair.GetReserves(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test05RemoveLiquidity: pair %s reserves after full exit: reserve0=%s reserve1=%s", s.pair, reserves.Reserve0, reserves.Reserve1)
}

func (s *RouterLifeCycleSuite) Test06AddLiquidityETH() {
	s.T().Logf("Test06AddLiquidityETH: token=%s liquidityAmount=%s nativeValue=%s", s.token0, s.liquidityAmount, s.nativeValue)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.token0, s.from, s.router, s.liquidityAmount)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)
	s.T().Logf("Test06AddLiquidityETH: approve %s to spend %s of %s, broadcast tx %s", s.router, s.liquidityAmount, s.token0, approveHash)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, s.token0, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && s.liquidityAmount.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	s.T().Logf("Test06AddLiquidityETH: approve confirmed in block %d", approveReceipt.BlockNumber)

	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	predicted, err := s.v2Router.CallAddLiquidityETH(s.ctx, s.router, s.from, s.token0, s.liquidityAmount, big.NewInt(0), big.NewInt(0), s.from, deadline, s.nativeValue, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(s.liquidityAmount.Cmp(predicted.AmountToken))
	s.Require().Zero(s.nativeValue.Cmp(predicted.AmountETH))
	s.Require().True(predicted.Liquidity.Sign() > 0)
	s.T().Logf("Test06AddLiquidityETH: CallAddLiquidityETH predicted amountToken=%s amountETH=%s liquidity=%s", predicted.AmountToken, predicted.AmountETH, predicted.Liquidity)

	builder = tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.v2Router.CreateAddLiquidityETH(s.ctx, builder, s.router, s.from, s.token0, s.liquidityAmount, big.NewInt(0), big.NewInt(0), s.from, deadline, s.nativeValue)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.v2Router.AddLiquidityETH(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test06AddLiquidityETH: broadcast addLiquidityETH tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "addLiquidityETH transaction reverted")
	s.T().Logf("Test06AddLiquidityETH: addLiquidityETH confirmed in block %d", receipt.BlockNumber)

	created, err := ExtractPairCreated(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(created, 1, "addLiquidityETH should have created the (token0, WETH) pair itself, via the factory")
	s.pairETH = created[0].Pair
	s.T().Logf("Test06AddLiquidityETH: factory created pair %s for (%s, WETH=%s)", s.pairETH, s.token0, s.wethAddr)

	deposits, err := weth.ExtractDeposits(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(deposits, 1, "addLiquidityETH should have wrapped the attached native value into WETH")
	s.Require().Zero(s.nativeValue.Cmp(deposits[0].Wad))
	s.T().Logf("Test06AddLiquidityETH: Deposit dst=%s wad=%s", deposits[0].Dst, deposits[0].Wad)

	var reserves *Reserves
	s.Require().Eventually(func() bool {
		reserves, err = s.v2Pair.GetReserves(s.ctx, s.pairETH, rpc.BlockTagLatest)
		return err == nil && reserves.Reserve0.Sign() > 0 && reserves.Reserve1.Sign() > 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "reserves not visible after addLiquidityETH")
	s.T().Logf("Test06AddLiquidityETH: pair %s reserves reserve0=%s reserve1=%s", s.pairETH, reserves.Reserve0, reserves.Reserve1)

	liquidityBalance, err := s.v2Pair.BalanceOf(s.ctx, s.pairETH, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test06AddLiquidityETH: %s's LP balance of pair %s: %s", s.from, s.pairETH, liquidityBalance)
}

func (s *RouterLifeCycleSuite) Test07SwapExactETHForTokens() {
	path := []*types.Address{s.wethAddr, s.token0}
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	pairToken0, err := s.v2Pair.Token0(s.ctx, s.pairETH, rpc.BlockTagLatest)
	s.Require().NoError(err)
	reserves, err := s.v2Pair.GetReserves(s.ctx, s.pairETH, rpc.BlockTagLatest)
	s.Require().NoError(err)

	wethReserve := reserves.Reserve1
	if pairToken0.Equal(s.wethAddr) {
		wethReserve = reserves.Reserve0
	}
	amountIn := new(big.Int).Div(wethReserve, big.NewInt(20))
	s.T().Logf("Test07SwapExactETHForTokens: path=[WETH=%s, %s] pairETH reserves reserve0=%s reserve1=%s, amountIn=%s deadline=%s", s.wethAddr, s.token0, reserves.Reserve0, reserves.Reserve1, amountIn, deadline)

	predicted, err := s.v2Router.CallSwapExactETHForTokens(s.ctx, s.router, s.from, big.NewInt(0), path, s.from, deadline, amountIn, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(predicted, 2)
	s.Require().Zero(amountIn.Cmp(predicted[0]))
	s.T().Logf("Test07SwapExactETHForTokens: CallSwapExactETHForTokens predicted amounts=%v", predicted)

	before, err := s.erc20.BalanceOf(s.ctx, s.token0, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test07SwapExactETHForTokens: token0 balance before: %s", before)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.v2Router.CreateSwapExactETHForTokens(s.ctx, builder, s.router, s.from, predicted[1], path, s.from, deadline, amountIn)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.v2Router.SwapExactETHForTokens(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test07SwapExactETHForTokens: broadcast swapExactETHForTokens tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swapExactETHForTokens transaction reverted")
	s.T().Logf("Test07SwapExactETHForTokens: swapExactETHForTokens confirmed in block %d", receipt.BlockNumber)

	deposits, err := weth.ExtractDeposits(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(deposits, 1, "swapExactETHForTokens should have wrapped the attached native value into WETH before swapping")
	s.Require().Zero(amountIn.Cmp(deposits[0].Wad))
	s.T().Logf("Test07SwapExactETHForTokens: Deposit dst=%s wad=%s", deposits[0].Dst, deposits[0].Wad)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.pairETH))
	s.T().Logf("Test07SwapExactETHForTokens: Swap sender=%s to=%s amount0In=%s amount1In=%s amount0Out=%s amount1Out=%s",
		swaps[0].Sender, swaps[0].To, swaps[0].Amount0In, swaps[0].Amount1In, swaps[0].Amount0Out, swaps[0].Amount1Out)

	var after *big.Int
	s.Require().Eventually(func() bool {
		after, err = s.erc20.BalanceOf(s.ctx, s.token0, s.from, rpc.BlockTagLatest)
		return err == nil && new(big.Int).Sub(after, before).Cmp(predicted[1]) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "token0 balance not visible after swap")
	s.T().Logf("Test07SwapExactETHForTokens: token0 balance after: %s (%s -> %s)", after, before, after)
}

func (s *RouterLifeCycleSuite) Test08SwapExactTokensForETH() {
	path := []*types.Address{s.token0, s.wethAddr}
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
	s.T().Logf("Test08SwapExactTokensForETH: path=[%s, WETH=%s] swapAmount=%s deadline=%s", s.token0, s.wethAddr, s.swapAmount, deadline)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.token0, s.from, s.router, s.swapAmount)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)
	s.T().Logf("Test08SwapExactTokensForETH: approve %s to spend %s of %s, broadcast tx %s", s.router, s.swapAmount, s.token0, approveHash)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, s.token0, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && s.swapAmount.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	s.T().Logf("Test08SwapExactTokensForETH: approve confirmed in block %d", approveReceipt.BlockNumber)

	predicted, err := s.v2Router.CallSwapExactTokensForETH(s.ctx, s.router, s.from, s.swapAmount, big.NewInt(0), path, s.from, deadline, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(predicted, 2)
	s.Require().Zero(s.swapAmount.Cmp(predicted[0]))
	s.T().Logf("Test08SwapExactTokensForETH: CallSwapExactTokensForETH predicted amounts=%v", predicted)

	var nativeBeforeHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.from.String(), rpc.BlockTagLatest, &nativeBeforeHex)))
	nativeBefore, ok := new(big.Int).SetString(strings.TrimPrefix(nativeBeforeHex, "0x"), 16)
	s.Require().True(ok)
	s.T().Logf("Test08SwapExactTokensForETH: native balance before: %s", nativeBefore)

	builder = tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.v2Router.CreateSwapExactTokensForETH(s.ctx, builder, s.router, s.from, s.swapAmount, predicted[1], path, s.from, deadline)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.v2Router.SwapExactTokensForETH(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test08SwapExactTokensForETH: broadcast swapExactTokensForETH tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swapExactTokensForETH transaction reverted")
	s.T().Logf("Test08SwapExactTokensForETH: swapExactTokensForETH confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.pairETH))
	s.T().Logf("Test08SwapExactTokensForETH: Swap sender=%s to=%s amount0In=%s amount1In=%s amount0Out=%s amount1Out=%s",
		swaps[0].Sender, swaps[0].To, swaps[0].Amount0In, swaps[0].Amount1In, swaps[0].Amount0Out, swaps[0].Amount1Out)

	withdrawals, err := weth.ExtractWithdrawals(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(withdrawals, 1, "swapExactTokensForETH should have unwrapped WETH back to native before sending it out")
	s.Require().Zero(predicted[1].Cmp(withdrawals[0].Wad))
	s.T().Logf("Test08SwapExactTokensForETH: Withdrawal src=%s wad=%s", withdrawals[0].Src, withdrawals[0].Wad)

	var nativeAfterHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.from.String(), rpc.BlockTagLatest, &nativeAfterHex)))
	nativeAfter, ok := new(big.Int).SetString(strings.TrimPrefix(nativeAfterHex, "0x"), 16)
	s.Require().True(ok)
	s.T().Logf("Test08SwapExactTokensForETH: native balance after: %s (%s -> %s, gas spent reduces the raw delta)", nativeAfter, nativeBefore, nativeAfter)
}

func (s *RouterLifeCycleSuite) Test09SwapETHForExactTokens() {
	path := []*types.Address{s.wethAddr, s.token0}
	amountOut := s.swapAmount
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	amountsIn, err := s.v2Router.GetAmountsIn(s.ctx, s.router, amountOut, path, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(amountsIn, 2)
	needed := amountsIn[0]

	value := new(big.Int).Mul(needed, big.NewInt(2))
	s.T().Logf("Test09SwapETHForExactTokens: path=[WETH=%s, %s] amountOut=%s needed=%s value(overpaid)=%s deadline=%s", s.wethAddr, s.token0, amountOut, needed, value, deadline)

	predicted, err := s.v2Router.CallSwapETHForExactTokens(s.ctx, s.router, s.from, amountOut, path, s.from, deadline, value, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(predicted, 2)
	s.Require().True(predicted[0].Cmp(value) <= 0, "predicted amountIn %s exceeds the value sent %s", predicted[0], value)
	s.Require().Zero(amountOut.Cmp(predicted[1]))
	s.T().Logf("Test09SwapETHForExactTokens: CallSwapETHForExactTokens predicted amounts=%v", predicted)

	var nativeBeforeHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.from.String(), rpc.BlockTagLatest, &nativeBeforeHex)))
	nativeBefore, ok := new(big.Int).SetString(strings.TrimPrefix(nativeBeforeHex, "0x"), 16)
	s.Require().True(ok)
	tokenBefore, err := s.erc20.BalanceOf(s.ctx, s.token0, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test09SwapETHForExactTokens: before: native=%s token0=%s", nativeBefore, tokenBefore)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.v2Router.CreateSwapETHForExactTokens(s.ctx, builder, s.router, s.from, amountOut, path, s.from, deadline, value)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.v2Router.SwapETHForExactTokens(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test09SwapETHForExactTokens: broadcast swapETHForExactTokens tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swapETHForExactTokens transaction reverted")
	s.T().Logf("Test09SwapETHForExactTokens: swapETHForExactTokens confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.pairETH))
	actualOut := swaps[0].Amount0Out
	actualIn := swaps[0].Amount1In
	if actualOut.Sign() == 0 {
		actualOut, actualIn = swaps[0].Amount1Out, swaps[0].Amount0In
	}
	s.Require().Zero(amountOut.Cmp(actualOut))
	s.Require().True(actualIn.Cmp(value) <= 0, "actual amountIn %s exceeds the value sent %s", actualIn, value)
	s.T().Logf("Test09SwapETHForExactTokens: Swap sender=%s to=%s amount0In=%s amount1In=%s amount0Out=%s amount1Out=%s",
		swaps[0].Sender, swaps[0].To, swaps[0].Amount0In, swaps[0].Amount1In, swaps[0].Amount0Out, swaps[0].Amount1Out)

	deposits, err := weth.ExtractDeposits(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(deposits, 1, "swapETHForExactTokens should only wrap the needed amount, not the full overpaid value")
	s.Require().Zero(actualIn.Cmp(deposits[0].Wad))
	s.T().Logf("Test09SwapETHForExactTokens: Deposit dst=%s wad=%s (value sent was %s, confirming the excess %s was never wrapped)", deposits[0].Dst, deposits[0].Wad, value, new(big.Int).Sub(value, deposits[0].Wad))

	var nativeAfterHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.from.String(), rpc.BlockTagLatest, &nativeAfterHex)))
	nativeAfter, ok := new(big.Int).SetString(strings.TrimPrefix(nativeAfterHex, "0x"), 16)
	s.Require().True(ok)

	var tokenAfter *big.Int
	s.Require().Eventually(func() bool {
		tokenAfter, err = s.erc20.BalanceOf(s.ctx, s.token0, s.from, rpc.BlockTagLatest)
		return err == nil && new(big.Int).Sub(tokenAfter, tokenBefore).Cmp(amountOut) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "token0 balance not visible after swap")

	spent := new(big.Int).Sub(nativeBefore, nativeAfter)
	s.Require().True(spent.Cmp(value) < 0, "spent %s should be less than the overpaid value %s if the excess was refunded", spent, value)
	s.T().Logf("Test09SwapETHForExactTokens: after: native=%s (spent %s, vs. value sent %s) token0=%s", nativeAfter, spent, value, tokenAfter)
}

func (s *RouterLifeCycleSuite) Test10SwapExpired() {
	path := []*types.Address{s.token0, s.token1}
	expired := big.NewInt(time.Now().Add(-1 * time.Minute).Unix())

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	_, err := s.v2Router.CreateSwapExactTokensForTokens(s.ctx, builder, s.router, s.from, s.swapAmount, big.NewInt(0), path, s.from, expired)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrExpired)
	s.T().Logf("Test10SwapExpired: deadline=%s rejected as expected: %v", expired, err)
}

func (s *RouterLifeCycleSuite) Test11SwapInvalidPath() {
	path := []*types.Address{s.token1, s.token0}
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	_, err := s.v2Router.CreateSwapExactETHForTokens(s.ctx, builder, s.router, s.from, big.NewInt(0), path, s.from, deadline, s.nativeValue)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInvalidPath)
	s.T().Logf("Test11SwapInvalidPath: path=[%s, %s] rejected as expected: %v", path[0], path[1], err)
}

func (s *RouterLifeCycleSuite) Test12SwapTokensForExactTokens() {
	path := []*types.Address{s.token0, s.token1}
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	pairToken0, err := s.v2Pair.Token0(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)
	currentReserves, err := s.v2Pair.GetReserves(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)
	reserve0, reserve1 := currentReserves.Reserve0, currentReserves.Reserve1
	if !pairToken0.Equal(s.token0) {
		reserve0, reserve1 = currentReserves.Reserve1, currentReserves.Reserve0
	}

	amount0Desired, amount1Desired := s.liquidityAmount, s.liquidityAmount
	if reserve0.Sign() > 0 && reserve1.Sign() > 0 {
		amount1Desired = new(big.Int).Div(new(big.Int).Mul(amount0Desired, reserve1), reserve0)
	}
	s.T().Logf("Test12SwapTokensForExactTokens: re-adding liquidity matching reserve0=%s reserve1=%s -> amount0Desired=%s amount1Desired=%s", reserve0, reserve1, amount0Desired, amount1Desired)

	for i, token := range []*types.Address{s.token0, s.token1} {
		amountDesired := amount0Desired
		if i == 1 {
			amountDesired = amount1Desired
		}

		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(s.PriceBuffer)
		builder.LimitBuffer(s.LimitBuffer)
		unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, token, s.from, s.router, amountDesired)
		s.Require().NoError(err)
		s.Require().NoError(unsignedApprove.Sign(s.signer))
		s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())

		approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
		s.Require().NoError(err)

		approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(approveReceipt.Status, "approve transaction reverted")
		var allowance *big.Int
		s.Require().Eventually(func() bool {
			allowance, err = s.erc20.Allowance(s.ctx, token, s.from, s.router, rpc.BlockTagLatest)
			return err == nil && amountDesired.Cmp(allowance) == 0
		}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	}

	addBuilder := tx.NewDynamicFeeBuilder()
	addBuilder.PriceBuffer(s.PriceBuffer)
	addBuilder.LimitBuffer(s.LimitBuffer)
	unsignedAdd, err := s.v2Router.CreateAddLiquidity(s.ctx, addBuilder, s.router, s.from, s.token0, s.token1, amount0Desired, amount1Desired, big.NewInt(0), big.NewInt(0), s.from, deadline)
	s.Require().NoError(err)
	s.Require().NoError(unsignedAdd.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedAdd.MaxFee())

	addHash, err := s.v2Router.AddLiquidity(s.ctx, unsignedAdd)
	s.Require().NoError(err)
	s.T().Logf("Test12SwapTokensForExactTokens: re-added liquidity, broadcast tx %s", addHash)

	addReceipt, err := tx.WaitForReceipt(s.ctx, s.client, addHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(addReceipt.Status, "addLiquidity transaction reverted")
	s.T().Logf("Test12SwapTokensForExactTokens: re-add liquidity confirmed in block %d", addReceipt.BlockNumber)

	reserves, err := s.v2Pair.GetReserves(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test12SwapTokensForExactTokens: pair %s reserves reserve0=%s reserve1=%s", s.pair, reserves.Reserve0, reserves.Reserve1)

	amountOut := s.swapAmount
	amountsIn, err := s.v2Router.GetAmountsIn(s.ctx, s.router, amountOut, path, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(amountsIn, 2)
	amountInMax := new(big.Int).Mul(amountsIn[0], big.NewInt(2))
	s.T().Logf("Test12SwapTokensForExactTokens: amountOut=%s needed=%s amountInMax(overpaid)=%s", amountOut, amountsIn[0], amountInMax)

	swapApproveBuilder := tx.NewDynamicFeeBuilder()
	swapApproveBuilder.PriceBuffer(s.PriceBuffer)
	swapApproveBuilder.LimitBuffer(s.LimitBuffer)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, swapApproveBuilder, s.token0, s.from, s.router, amountInMax)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, s.token0, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && amountInMax.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")

	predicted, err := s.v2Router.CallSwapTokensForExactTokens(s.ctx, s.router, s.from, amountOut, amountInMax, path, s.from, deadline, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(predicted, 2)
	s.Require().Zero(amountOut.Cmp(predicted[1]))
	s.T().Logf("Test12SwapTokensForExactTokens: CallSwapTokensForExactTokens predicted amounts=%v", predicted)

	beforeIn, err := s.erc20.BalanceOf(s.ctx, s.token0, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	beforeOut, err := s.erc20.BalanceOf(s.ctx, s.token1, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(s.PriceBuffer)
	swapBuilder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.v2Router.CreateSwapTokensForExactTokens(s.ctx, swapBuilder, s.router, s.from, amountOut, amountInMax, path, s.from, deadline)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.v2Router.SwapTokensForExactTokens(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test12SwapTokensForExactTokens: broadcast swapTokensForExactTokens tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swapTokensForExactTokens transaction reverted")
	s.T().Logf("Test12SwapTokensForExactTokens: swapTokensForExactTokens confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().Zero(amountOut.Cmp(swaps[0].Amount1Out))
	actualIn := swaps[0].Amount0In
	s.Require().True(actualIn.Cmp(amountInMax) <= 0, "actual amountIn %s exceeds amountInMax %s", actualIn, amountInMax)
	s.T().Logf("Test12SwapTokensForExactTokens: Swap sender=%s to=%s amount0In=%s amount1In=%s amount0Out=%s amount1Out=%s",
		swaps[0].Sender, swaps[0].To, swaps[0].Amount0In, swaps[0].Amount1In, swaps[0].Amount0Out, swaps[0].Amount1Out)

	var afterIn, afterOut *big.Int
	s.Require().Eventually(func() bool {
		afterIn, err = s.erc20.BalanceOf(s.ctx, s.token0, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		afterOut, err = s.erc20.BalanceOf(s.ctx, s.token1, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		return new(big.Int).Sub(beforeIn, afterIn).Cmp(actualIn) == 0 &&
			new(big.Int).Sub(afterOut, beforeOut).Cmp(amountOut) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not visible after swap")

	spentIn := new(big.Int).Sub(beforeIn, afterIn)
	s.Require().Zero(actualIn.Cmp(spentIn))
	s.T().Logf("Test12SwapTokensForExactTokens: balances after: token0=%s (spent %s, vs. approved max %s) token1=%s (%s -> %s)",
		afterIn, spentIn, amountInMax, afterOut, beforeOut, afterOut)
}

func (s *RouterLifeCycleSuite) Test13SwapTokensForExactETH() {
	path := []*types.Address{s.token0, s.wethAddr}
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	pairToken0, err := s.v2Pair.Token0(s.ctx, s.pairETH, rpc.BlockTagLatest)
	s.Require().NoError(err)
	reserves, err := s.v2Pair.GetReserves(s.ctx, s.pairETH, rpc.BlockTagLatest)
	s.Require().NoError(err)

	wethReserve := reserves.Reserve1
	if pairToken0.Equal(s.wethAddr) {
		wethReserve = reserves.Reserve0
	}
	amountOut := new(big.Int).Div(wethReserve, big.NewInt(20))
	s.T().Logf("Test13SwapTokensForExactETH: path=[%s, WETH=%s] pairETH reserves reserve0=%s reserve1=%s, amountOut=%s", s.token0, s.wethAddr, reserves.Reserve0, reserves.Reserve1, amountOut)

	amountsIn, err := s.v2Router.GetAmountsIn(s.ctx, s.router, amountOut, path, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(amountsIn, 2)
	amountInMax := new(big.Int).Mul(amountsIn[0], big.NewInt(2))
	s.T().Logf("Test13SwapTokensForExactETH: needed=%s amountInMax(overpaid)=%s", amountsIn[0], amountInMax)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.token0, s.from, s.router, amountInMax)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)
	s.T().Logf("Test13SwapTokensForExactETH: approve %s to spend %s of %s, broadcast tx %s", s.router, amountInMax, s.token0, approveHash)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, s.token0, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && amountInMax.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	s.T().Logf("Test13SwapTokensForExactETH: approve confirmed in block %d", approveReceipt.BlockNumber)

	predicted, err := s.v2Router.CallSwapTokensForExactETH(s.ctx, s.router, s.from, amountOut, amountInMax, path, s.from, deadline, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(predicted, 2)
	s.Require().Zero(amountOut.Cmp(predicted[1]))
	s.T().Logf("Test13SwapTokensForExactETH: CallSwapTokensForExactETH predicted amounts=%v", predicted)

	tokenBefore, err := s.erc20.BalanceOf(s.ctx, s.token0, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(s.PriceBuffer)
	swapBuilder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.v2Router.CreateSwapTokensForExactETH(s.ctx, swapBuilder, s.router, s.from, amountOut, amountInMax, path, s.from, deadline)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.v2Router.SwapTokensForExactETH(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test13SwapTokensForExactETH: broadcast swapTokensForExactETH tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swapTokensForExactETH transaction reverted")
	s.T().Logf("Test13SwapTokensForExactETH: swapTokensForExactETH confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.pairETH))
	actualOut, actualIn := swaps[0].Amount0Out, swaps[0].Amount1In
	if actualOut.Sign() == 0 {
		actualOut, actualIn = swaps[0].Amount1Out, swaps[0].Amount0In
	}
	s.Require().Zero(amountOut.Cmp(actualOut))
	s.Require().True(actualIn.Cmp(amountInMax) <= 0, "actual amountIn %s exceeds amountInMax %s", actualIn, amountInMax)
	s.T().Logf("Test13SwapTokensForExactETH: Swap sender=%s to=%s amount0In=%s amount1In=%s amount0Out=%s amount1Out=%s",
		swaps[0].Sender, swaps[0].To, swaps[0].Amount0In, swaps[0].Amount1In, swaps[0].Amount0Out, swaps[0].Amount1Out)

	withdrawals, err := weth.ExtractWithdrawals(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(withdrawals, 1, "swapTokensForExactETH should have unwrapped WETH back to native before sending it out")
	s.Require().Zero(amountOut.Cmp(withdrawals[0].Wad))
	s.T().Logf("Test13SwapTokensForExactETH: Withdrawal src=%s wad=%s", withdrawals[0].Src, withdrawals[0].Wad)

	var tokenAfter *big.Int
	s.Require().Eventually(func() bool {
		tokenAfter, err = s.erc20.BalanceOf(s.ctx, s.token0, s.from, rpc.BlockTagLatest)
		return err == nil && new(big.Int).Sub(tokenBefore, tokenAfter).Cmp(actualIn) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "token0 balance not visible after swap")

	spentIn := new(big.Int).Sub(tokenBefore, tokenAfter)
	s.Require().Zero(actualIn.Cmp(spentIn))
	s.T().Logf("Test13SwapTokensForExactETH: token0 spent %s (vs. approved max %s)", spentIn, amountInMax)
}

func (s *RouterLifeCycleSuite) Test14RemoveLiquidityETH() {
	liquidity, err := s.v2Pair.BalanceOf(s.ctx, s.pairETH, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(liquidity.Sign() > 0, "no LP balance left to remove")
	s.T().Logf("Test14RemoveLiquidityETH: pairETH=%s LP balance=%s", s.pairETH, liquidity)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.pairETH, s.from, s.router, liquidity)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)
	s.T().Logf("Test14RemoveLiquidityETH: approve %s to spend %s of LP token %s, broadcast tx %s", s.router, liquidity, s.pairETH, approveHash)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, s.pairETH, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && liquidity.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	s.T().Logf("Test14RemoveLiquidityETH: approve confirmed in block %d", approveReceipt.BlockNumber)

	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	predicted, err := s.v2Router.CallRemoveLiquidityETH(s.ctx, s.router, s.from, s.token0, liquidity, big.NewInt(0), big.NewInt(0), s.from, deadline, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(predicted.AmountToken.Sign() > 0)
	s.Require().True(predicted.AmountETH.Sign() > 0)
	s.T().Logf("Test14RemoveLiquidityETH: CallRemoveLiquidityETH predicted amountToken=%s amountETH=%s", predicted.AmountToken, predicted.AmountETH)

	tokenBefore, err := s.erc20.BalanceOf(s.ctx, s.token0, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)

	removeBuilder := tx.NewDynamicFeeBuilder()
	removeBuilder.PriceBuffer(s.PriceBuffer)
	removeBuilder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.v2Router.CreateRemoveLiquidityETH(s.ctx, removeBuilder, s.router, s.from, s.token0, liquidity, predicted.AmountToken, predicted.AmountETH, s.from, deadline)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.v2Router.RemoveLiquidityETH(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test14RemoveLiquidityETH: broadcast removeLiquidityETH tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "removeLiquidityETH transaction reverted")
	s.T().Logf("Test14RemoveLiquidityETH: removeLiquidityETH confirmed in block %d", receipt.BlockNumber)

	burns, err := ExtractBurn(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(burns, 1)
	s.T().Logf("Test14RemoveLiquidityETH: Burn sender=%s amount0=%s amount1=%s to=%s", burns[0].Sender, burns[0].Amount0, burns[0].Amount1, burns[0].To)

	withdrawals, err := weth.ExtractWithdrawals(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(withdrawals, 1, "removeLiquidityETH should have unwrapped WETH back to native before sending it out")
	s.Require().Zero(predicted.AmountETH.Cmp(withdrawals[0].Wad))
	s.T().Logf("Test14RemoveLiquidityETH: Withdrawal src=%s wad=%s", withdrawals[0].Src, withdrawals[0].Wad)

	var tokenAfter, remainingLiquidity *big.Int
	s.Require().Eventually(func() bool {
		tokenAfter, err = s.erc20.BalanceOf(s.ctx, s.token0, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		remainingLiquidity, err = s.v2Pair.BalanceOf(s.ctx, s.pairETH, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		return new(big.Int).Sub(tokenAfter, tokenBefore).Cmp(predicted.AmountToken) == 0 && remainingLiquidity.Sign() == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not visible after removeLiquidityETH")
	s.T().Logf("Test14RemoveLiquidityETH: token0 %s -> %s, remaining LP balance=%s", tokenBefore, tokenAfter, remainingLiquidity)
}

func (s *RouterLifeCycleSuite) Test15RemoveLiquidityWithPermit() {
	liquidity, err := s.v2Pair.BalanceOf(s.ctx, s.pair, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(liquidity.Sign() > 0, "no LP balance left to remove")
	s.T().Logf("Test15RemoveLiquidityWithPermit: pair=%s LP balance=%s", s.pair, liquidity)

	nonce, err := s.v2Pair.Nonces(s.ctx, s.pair, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	domain := eip712.NewDomain("Uniswap V2", erc2612.DefaultVersion, s.chainID, s.pair, nil)
	permit := erc2612.NewPermit(s.from, s.router, liquidity, nonce, deadline)
	sig, err := permit.Sign(domain, s.signer)
	s.Require().NoError(err)
	s.T().Logf("Test15RemoveLiquidityWithPermit: signed permit nonce=%s deadline=%s (no Approve transaction needed)", nonce, deadline)

	predicted, err := s.v2Router.CallRemoveLiquidityWithPermit(s.ctx, s.router, s.from, s.token0, s.token1, liquidity, big.NewInt(0), big.NewInt(0), s.from, deadline, false, sig, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(predicted.AmountA.Sign() > 0)
	s.Require().True(predicted.AmountB.Sign() > 0)
	s.T().Logf("Test15RemoveLiquidityWithPermit: CallRemoveLiquidityWithPermit predicted amountA=%s amountB=%s", predicted.AmountA, predicted.AmountB)

	beforeA, err := s.erc20.BalanceOf(s.ctx, s.token0, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	beforeB, err := s.erc20.BalanceOf(s.ctx, s.token1, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.v2Router.CreateRemoveLiquidityWithPermit(s.ctx, builder, s.router, s.from, s.token0, s.token1, liquidity, predicted.AmountA, predicted.AmountB, s.from, deadline, false, sig)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.v2Router.RemoveLiquidityWithPermit(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test15RemoveLiquidityWithPermit: broadcast removeLiquidityWithPermit tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "removeLiquidityWithPermit transaction reverted")
	s.T().Logf("Test15RemoveLiquidityWithPermit: removeLiquidityWithPermit confirmed in block %d", receipt.BlockNumber)

	burns, err := ExtractBurn(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(burns, 1)
	s.Require().Zero(predicted.AmountA.Cmp(burns[0].Amount0))
	s.Require().Zero(predicted.AmountB.Cmp(burns[0].Amount1))
	s.T().Logf("Test15RemoveLiquidityWithPermit: Burn sender=%s amount0=%s amount1=%s to=%s", burns[0].Sender, burns[0].Amount0, burns[0].Amount1, burns[0].To)

	var afterA, afterB, nonceAfter, remainingLiquidity *big.Int
	s.Require().Eventually(func() bool {
		afterA, err = s.erc20.BalanceOf(s.ctx, s.token0, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		afterB, err = s.erc20.BalanceOf(s.ctx, s.token1, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		remainingLiquidity, err = s.v2Pair.BalanceOf(s.ctx, s.pair, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		nonceAfter, err = s.v2Pair.Nonces(s.ctx, s.pair, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		return new(big.Int).Sub(afterA, beforeA).Cmp(predicted.AmountA) == 0 &&
			new(big.Int).Sub(afterB, beforeB).Cmp(predicted.AmountB) == 0 &&
			remainingLiquidity.Sign() == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not visible after removeLiquidityWithPermit")
	s.Require().Zero(new(big.Int).Add(nonce, big.NewInt(1)).Cmp(nonceAfter), "permit nonce should have advanced by exactly 1")
	s.T().Logf("Test15RemoveLiquidityWithPermit: balances after: token0=%s (%s -> %s) token1=%s (%s -> %s), nonce %s -> %s, remaining LP balance=%s",
		afterA, beforeA, afterA, afterB, beforeB, afterB, nonce, nonceAfter, remainingLiquidity)
}

func (s *RouterLifeCycleSuite) Test16RemoveLiquidityETHWithPermit() {
	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.token0, s.from, s.router, s.liquidityAmount)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, s.token0, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && s.liquidityAmount.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")

	addDeadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
	addBuilder := tx.NewDynamicFeeBuilder()
	addBuilder.PriceBuffer(s.PriceBuffer)
	addBuilder.LimitBuffer(s.LimitBuffer)
	unsignedAdd, err := s.v2Router.CreateAddLiquidityETH(s.ctx, addBuilder, s.router, s.from, s.token0, s.liquidityAmount, big.NewInt(0), big.NewInt(0), s.from, addDeadline, s.nativeValue)
	s.Require().NoError(err)
	s.Require().NoError(unsignedAdd.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedAdd.MaxFee())

	addHash, err := s.v2Router.AddLiquidityETH(s.ctx, unsignedAdd)
	s.Require().NoError(err)
	s.T().Logf("Test16RemoveLiquidityETHWithPermit: re-added liquidity to pairETH=%s, broadcast tx %s", s.pairETH, addHash)

	addReceipt, err := tx.WaitForReceipt(s.ctx, s.client, addHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(addReceipt.Status, "addLiquidityETH transaction reverted")
	s.T().Logf("Test16RemoveLiquidityETHWithPermit: re-add liquidity confirmed in block %d", addReceipt.BlockNumber)

	var liquidity *big.Int
	s.Require().Eventually(func() bool {
		liquidity, err = s.v2Pair.BalanceOf(s.ctx, s.pairETH, s.from, rpc.BlockTagLatest)
		return err == nil && liquidity.Sign() > 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "LP balance not visible after re-adding liquidity")
	s.T().Logf("Test16RemoveLiquidityETHWithPermit: pairETH=%s LP balance=%s", s.pairETH, liquidity)

	nonce, err := s.v2Pair.Nonces(s.ctx, s.pairETH, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	domain := eip712.NewDomain("Uniswap V2", erc2612.DefaultVersion, s.chainID, s.pairETH, nil)
	permit := erc2612.NewPermit(s.from, s.router, liquidity, nonce, deadline)
	sig, err := permit.Sign(domain, s.signer)
	s.Require().NoError(err)
	s.T().Logf("Test16RemoveLiquidityETHWithPermit: signed permit nonce=%s deadline=%s (no Approve transaction needed)", nonce, deadline)

	predicted, err := s.v2Router.CallRemoveLiquidityETHWithPermit(s.ctx, s.router, s.from, s.token0, liquidity, big.NewInt(0), big.NewInt(0), s.from, deadline, false, sig, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(predicted.AmountToken.Sign() > 0)
	s.Require().True(predicted.AmountETH.Sign() > 0)
	s.T().Logf("Test16RemoveLiquidityETHWithPermit: CallRemoveLiquidityETHWithPermit predicted amountToken=%s amountETH=%s", predicted.AmountToken, predicted.AmountETH)

	tokenBefore, err := s.erc20.BalanceOf(s.ctx, s.token0, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)

	removeBuilder := tx.NewDynamicFeeBuilder()
	removeBuilder.PriceBuffer(s.PriceBuffer)
	removeBuilder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.v2Router.CreateRemoveLiquidityETHWithPermit(s.ctx, removeBuilder, s.router, s.from, s.token0, liquidity, predicted.AmountToken, predicted.AmountETH, s.from, deadline, false, sig)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.v2Router.RemoveLiquidityETHWithPermit(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test16RemoveLiquidityETHWithPermit: broadcast removeLiquidityETHWithPermit tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "removeLiquidityETHWithPermit transaction reverted")
	s.T().Logf("Test16RemoveLiquidityETHWithPermit: removeLiquidityETHWithPermit confirmed in block %d", receipt.BlockNumber)

	burns, err := ExtractBurn(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(burns, 1)
	s.T().Logf("Test16RemoveLiquidityETHWithPermit: Burn sender=%s amount0=%s amount1=%s to=%s", burns[0].Sender, burns[0].Amount0, burns[0].Amount1, burns[0].To)

	withdrawals, err := weth.ExtractWithdrawals(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(withdrawals, 1, "removeLiquidityETHWithPermit should have unwrapped WETH back to native before sending it out")
	s.Require().Zero(predicted.AmountETH.Cmp(withdrawals[0].Wad))
	s.T().Logf("Test16RemoveLiquidityETHWithPermit: Withdrawal src=%s wad=%s", withdrawals[0].Src, withdrawals[0].Wad)

	var tokenAfter, nonceAfter, remainingLiquidity *big.Int
	s.Require().Eventually(func() bool {
		tokenAfter, err = s.erc20.BalanceOf(s.ctx, s.token0, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		remainingLiquidity, err = s.v2Pair.BalanceOf(s.ctx, s.pairETH, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		nonceAfter, err = s.v2Pair.Nonces(s.ctx, s.pairETH, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		return new(big.Int).Sub(tokenAfter, tokenBefore).Cmp(predicted.AmountToken) == 0 && remainingLiquidity.Sign() == 0 &&
			new(big.Int).Add(nonce, big.NewInt(1)).Cmp(nonceAfter) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not visible after removeLiquidityETHWithPermit")
	s.Require().Zero(new(big.Int).Add(nonce, big.NewInt(1)).Cmp(nonceAfter), "permit nonce should have advanced by exactly 1")
	s.T().Logf("Test16RemoveLiquidityETHWithPermit: token0 %s -> %s, nonce %s -> %s, remaining LP balance=%s", tokenBefore, tokenAfter, nonce, nonceAfter, remainingLiquidity)
}

func (s *RouterLifeCycleSuite) Test17SwapExactTokensForTokensSupportingFeeOnTransferTokens() {
	path := []*types.Address{s.token0, s.token1}
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	pairToken0, err := s.v2Pair.Token0(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)
	currentReserves, err := s.v2Pair.GetReserves(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)
	reserve0, reserve1 := currentReserves.Reserve0, currentReserves.Reserve1
	if !pairToken0.Equal(s.token0) {
		reserve0, reserve1 = currentReserves.Reserve1, currentReserves.Reserve0
	}

	amount0Desired, amount1Desired := s.liquidityAmount, s.liquidityAmount
	if reserve0.Sign() > 0 && reserve1.Sign() > 0 {
		amount1Desired = new(big.Int).Div(new(big.Int).Mul(amount0Desired, reserve1), reserve0)
	}
	s.T().Logf("Test17SwapExactTokensForTokensSupportingFeeOnTransferTokens: re-adding liquidity matching reserve0=%s reserve1=%s -> amount0Desired=%s amount1Desired=%s", reserve0, reserve1, amount0Desired, amount1Desired)

	for i, token := range []*types.Address{s.token0, s.token1} {
		amountDesired := amount0Desired
		if i == 1 {
			amountDesired = amount1Desired
		}

		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(s.PriceBuffer)
		builder.LimitBuffer(s.LimitBuffer)
		unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, token, s.from, s.router, amountDesired)
		s.Require().NoError(err)
		s.Require().NoError(unsignedApprove.Sign(s.signer))
		s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())

		approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
		s.Require().NoError(err)

		approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(approveReceipt.Status, "approve transaction reverted")
		var allowance *big.Int
		s.Require().Eventually(func() bool {
			allowance, err = s.erc20.Allowance(s.ctx, token, s.from, s.router, rpc.BlockTagLatest)
			return err == nil && amountDesired.Cmp(allowance) == 0
		}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	}

	addBuilder := tx.NewDynamicFeeBuilder()
	addBuilder.PriceBuffer(s.PriceBuffer)
	addBuilder.LimitBuffer(s.LimitBuffer)
	unsignedAdd, err := s.v2Router.CreateAddLiquidity(s.ctx, addBuilder, s.router, s.from, s.token0, s.token1, amount0Desired, amount1Desired, big.NewInt(0), big.NewInt(0), s.from, deadline)
	s.Require().NoError(err)
	s.Require().NoError(unsignedAdd.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedAdd.MaxFee())

	addHash, err := s.v2Router.AddLiquidity(s.ctx, unsignedAdd)
	s.Require().NoError(err)
	s.T().Logf("Test17SwapExactTokensForTokensSupportingFeeOnTransferTokens: re-added liquidity, broadcast tx %s", addHash)

	addReceipt, err := tx.WaitForReceipt(s.ctx, s.client, addHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(addReceipt.Status, "addLiquidity transaction reverted")
	s.T().Logf("Test17SwapExactTokensForTokensSupportingFeeOnTransferTokens: re-add liquidity confirmed in block %d", addReceipt.BlockNumber)

	wantToken0Reserve := new(big.Int).Add(reserve0, amount0Desired)
	s.Require().Eventually(func() bool {
		newReserves, err := s.v2Pair.GetReserves(s.ctx, s.pair, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		newToken0Reserve := newReserves.Reserve0
		if !pairToken0.Equal(s.token0) {
			newToken0Reserve = newReserves.Reserve1
		}
		return newToken0Reserve.Cmp(wantToken0Reserve) >= 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "reserves not visible after re-adding liquidity")

	amountsOut, err := s.v2Router.GetAmountsOut(s.ctx, s.router, s.swapAmount, path, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(amountsOut, 2)
	s.T().Logf("Test17SwapExactTokensForTokensSupportingFeeOnTransferTokens: GetAmountsOut predicted amounts=%v", amountsOut)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.token0, s.from, s.router, s.swapAmount)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, s.token0, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && s.swapAmount.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")

	beforeOut, err := s.erc20.BalanceOf(s.ctx, s.token1, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(s.PriceBuffer)
	swapBuilder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.v2Router.CreateSwapExactTokensForTokensSupportingFeeOnTransferTokens(s.ctx, swapBuilder, s.router, s.from, s.swapAmount, amountsOut[1], path, s.from, deadline)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.v2Router.SwapExactTokensForTokensSupportingFeeOnTransferTokens(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test17SwapExactTokensForTokensSupportingFeeOnTransferTokens: broadcast tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swapExactTokensForTokensSupportingFeeOnTransferTokens transaction reverted")
	s.T().Logf("Test17SwapExactTokensForTokensSupportingFeeOnTransferTokens: confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().Zero(amountsOut[1].Cmp(swaps[0].Amount1Out))
	s.T().Logf("Test17SwapExactTokensForTokensSupportingFeeOnTransferTokens: Swap sender=%s to=%s amount0In=%s amount1In=%s amount0Out=%s amount1Out=%s",
		swaps[0].Sender, swaps[0].To, swaps[0].Amount0In, swaps[0].Amount1In, swaps[0].Amount0Out, swaps[0].Amount1Out)

	var afterOut *big.Int
	s.Require().Eventually(func() bool {
		afterOut, err = s.erc20.BalanceOf(s.ctx, s.token1, s.from, rpc.BlockTagLatest)
		return err == nil && new(big.Int).Sub(afterOut, beforeOut).Cmp(amountsOut[1]) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "token1 balance not visible after swap")
	s.T().Logf("Test17SwapExactTokensForTokensSupportingFeeOnTransferTokens: token1 balance %s -> %s", beforeOut, afterOut)
}

func (s *RouterLifeCycleSuite) Test18SwapExactETHForTokensSupportingFeeOnTransferTokens() {
	path := []*types.Address{s.wethAddr, s.token0}
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	pairToken0, err := s.v2Pair.Token0(s.ctx, s.pairETH, rpc.BlockTagLatest)
	s.Require().NoError(err)
	currentReserves, err := s.v2Pair.GetReserves(s.ctx, s.pairETH, rpc.BlockTagLatest)
	s.Require().NoError(err)
	tokenReserve, wethReserve := currentReserves.Reserve0, currentReserves.Reserve1
	if !pairToken0.Equal(s.token0) {
		tokenReserve, wethReserve = currentReserves.Reserve1, currentReserves.Reserve0
	}

	tokenDesired := s.liquidityAmount
	if tokenReserve.Sign() > 0 && wethReserve.Sign() > 0 {
		tokenDesired = new(big.Int).Div(new(big.Int).Mul(s.nativeValue, tokenReserve), wethReserve)
	}
	s.T().Logf("Test18SwapExactETHForTokensSupportingFeeOnTransferTokens: re-adding liquidity matching tokenReserve=%s wethReserve=%s -> tokenDesired=%s nativeValue=%s", tokenReserve, wethReserve, tokenDesired, s.nativeValue)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.token0, s.from, s.router, tokenDesired)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, s.token0, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && tokenDesired.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")

	addDeadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
	addBuilder := tx.NewDynamicFeeBuilder()
	addBuilder.PriceBuffer(s.PriceBuffer)
	addBuilder.LimitBuffer(s.LimitBuffer)
	unsignedAdd, err := s.v2Router.CreateAddLiquidityETH(s.ctx, addBuilder, s.router, s.from, s.token0, tokenDesired, big.NewInt(0), big.NewInt(0), s.from, addDeadline, s.nativeValue)
	s.Require().NoError(err)
	s.Require().NoError(unsignedAdd.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedAdd.MaxFee())

	addHash, err := s.v2Router.AddLiquidityETH(s.ctx, unsignedAdd)
	s.Require().NoError(err)
	s.T().Logf("Test18SwapExactETHForTokensSupportingFeeOnTransferTokens: re-added liquidity to pairETH, broadcast tx %s", addHash)

	addReceipt, err := tx.WaitForReceipt(s.ctx, s.client, addHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(addReceipt.Status, "addLiquidityETH transaction reverted")
	s.T().Logf("Test18SwapExactETHForTokensSupportingFeeOnTransferTokens: re-add liquidity confirmed in block %d", addReceipt.BlockNumber)

	wantWethReserve := new(big.Int).Add(wethReserve, s.nativeValue)
	var reserves *Reserves
	s.Require().Eventually(func() bool {
		reserves, err = s.v2Pair.GetReserves(s.ctx, s.pairETH, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		newWethReserve := reserves.Reserve1
		if pairToken0.Equal(s.wethAddr) {
			newWethReserve = reserves.Reserve0
		}
		return newWethReserve.Cmp(wantWethReserve) >= 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "reserves not visible after addLiquidityETH")

	newWethReserve := reserves.Reserve1
	if pairToken0.Equal(s.wethAddr) {
		newWethReserve = reserves.Reserve0
	}
	amountIn := new(big.Int).Div(newWethReserve, big.NewInt(20))
	s.T().Logf("Test18SwapExactETHForTokensSupportingFeeOnTransferTokens: pairETH reserves reserve0=%s reserve1=%s, amountIn=%s", reserves.Reserve0, reserves.Reserve1, amountIn)

	amountsOut, err := s.v2Router.GetAmountsOut(s.ctx, s.router, amountIn, path, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(amountsOut, 2)
	s.T().Logf("Test18SwapExactETHForTokensSupportingFeeOnTransferTokens: GetAmountsOut predicted amounts=%v", amountsOut)

	before, err := s.erc20.BalanceOf(s.ctx, s.token0, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(s.PriceBuffer)
	swapBuilder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.v2Router.CreateSwapExactETHForTokensSupportingFeeOnTransferTokens(s.ctx, swapBuilder, s.router, s.from, amountsOut[1], path, s.from, deadline, amountIn)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.v2Router.SwapExactETHForTokensSupportingFeeOnTransferTokens(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test18SwapExactETHForTokensSupportingFeeOnTransferTokens: broadcast tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swapExactETHForTokensSupportingFeeOnTransferTokens transaction reverted")
	s.T().Logf("Test18SwapExactETHForTokensSupportingFeeOnTransferTokens: confirmed in block %d", receipt.BlockNumber)

	deposits, err := weth.ExtractDeposits(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(deposits, 1, "swapExactETHForTokensSupportingFeeOnTransferTokens should have wrapped the attached native value into WETH before swapping")
	s.Require().Zero(amountIn.Cmp(deposits[0].Wad))

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.pairETH))
	s.T().Logf("Test18SwapExactETHForTokensSupportingFeeOnTransferTokens: Swap sender=%s to=%s amount0In=%s amount1In=%s amount0Out=%s amount1Out=%s",
		swaps[0].Sender, swaps[0].To, swaps[0].Amount0In, swaps[0].Amount1In, swaps[0].Amount0Out, swaps[0].Amount1Out)

	var after *big.Int
	s.Require().Eventually(func() bool {
		after, err = s.erc20.BalanceOf(s.ctx, s.token0, s.from, rpc.BlockTagLatest)
		return err == nil && new(big.Int).Sub(after, before).Cmp(amountsOut[1]) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "token0 balance not visible after swap")
	s.T().Logf("Test18SwapExactETHForTokensSupportingFeeOnTransferTokens: token0 balance %s -> %s", before, after)
}

func (s *RouterLifeCycleSuite) Test19SwapExactTokensForETHSupportingFeeOnTransferTokens() {
	path := []*types.Address{s.token0, s.wethAddr}
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	pairToken0, err := s.v2Pair.Token0(s.ctx, s.pairETH, rpc.BlockTagLatest)
	s.Require().NoError(err)
	reserves, err := s.v2Pair.GetReserves(s.ctx, s.pairETH, rpc.BlockTagLatest)
	s.Require().NoError(err)
	tokenReserve := reserves.Reserve1
	if pairToken0.Equal(s.token0) {
		tokenReserve = reserves.Reserve0
	}

	amountIn := new(big.Int).Div(tokenReserve, big.NewInt(20))
	s.T().Logf("Test19SwapExactTokensForETHSupportingFeeOnTransferTokens: pairETH reserves reserve0=%s reserve1=%s, amountIn=%s", reserves.Reserve0, reserves.Reserve1, amountIn)

	amountsOut, err := s.v2Router.GetAmountsOut(s.ctx, s.router, amountIn, path, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(amountsOut, 2)
	s.T().Logf("Test19SwapExactTokensForETHSupportingFeeOnTransferTokens: GetAmountsOut predicted amounts=%v", amountsOut)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.token0, s.from, s.router, amountIn)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, s.token0, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && amountIn.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")

	var nativeBeforeHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.from.String(), rpc.BlockTagLatest, &nativeBeforeHex)))
	nativeBefore, ok := new(big.Int).SetString(strings.TrimPrefix(nativeBeforeHex, "0x"), 16)
	s.Require().True(ok)

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(s.PriceBuffer)
	swapBuilder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.v2Router.CreateSwapExactTokensForETHSupportingFeeOnTransferTokens(s.ctx, swapBuilder, s.router, s.from, amountIn, amountsOut[1], path, s.from, deadline)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.v2Router.SwapExactTokensForETHSupportingFeeOnTransferTokens(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test19SwapExactTokensForETHSupportingFeeOnTransferTokens: broadcast tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swapExactTokensForETHSupportingFeeOnTransferTokens transaction reverted")
	s.T().Logf("Test19SwapExactTokensForETHSupportingFeeOnTransferTokens: confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.pairETH))
	s.T().Logf("Test19SwapExactTokensForETHSupportingFeeOnTransferTokens: Swap sender=%s to=%s amount0In=%s amount1In=%s amount0Out=%s amount1Out=%s",
		swaps[0].Sender, swaps[0].To, swaps[0].Amount0In, swaps[0].Amount1In, swaps[0].Amount0Out, swaps[0].Amount1Out)

	withdrawals, err := weth.ExtractWithdrawals(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(withdrawals, 1, "swapExactTokensForETHSupportingFeeOnTransferTokens should have unwrapped WETH back to native before sending it out")
	s.Require().Zero(amountsOut[1].Cmp(withdrawals[0].Wad))
	s.T().Logf("Test19SwapExactTokensForETHSupportingFeeOnTransferTokens: Withdrawal src=%s wad=%s", withdrawals[0].Src, withdrawals[0].Wad)

	var nativeAfterHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.from.String(), rpc.BlockTagLatest, &nativeAfterHex)))
	nativeAfter, ok := new(big.Int).SetString(strings.TrimPrefix(nativeAfterHex, "0x"), 16)
	s.Require().True(ok)
	s.T().Logf("Test19SwapExactTokensForETHSupportingFeeOnTransferTokens: native balance %s -> %s (gas spent reduces the raw delta)", nativeBefore, nativeAfter)
}

func (s *RouterLifeCycleSuite) Test20RemoveLiquidityETHSupportingFeeOnTransferTokens() {
	liquidity, err := s.v2Pair.BalanceOf(s.ctx, s.pairETH, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(liquidity.Sign() > 0, "no LP balance left to remove")
	s.T().Logf("Test20RemoveLiquidityETHSupportingFeeOnTransferTokens: pairETH=%s LP balance=%s", s.pairETH, liquidity)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.pairETH, s.from, s.router, liquidity)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, s.pairETH, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && liquidity.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")

	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	predictedAmountETH, err := s.v2Router.CallRemoveLiquidityETHSupportingFeeOnTransferTokens(s.ctx, s.router, s.from, s.token0, liquidity, big.NewInt(0), big.NewInt(0), s.from, deadline, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(predictedAmountETH.Sign() > 0)
	s.T().Logf("Test20RemoveLiquidityETHSupportingFeeOnTransferTokens: CallRemoveLiquidityETHSupportingFeeOnTransferTokens predicted amountETH=%s", predictedAmountETH)

	tokenBefore, err := s.erc20.BalanceOf(s.ctx, s.token0, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)

	removeBuilder := tx.NewDynamicFeeBuilder()
	removeBuilder.PriceBuffer(s.PriceBuffer)
	removeBuilder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.v2Router.CreateRemoveLiquidityETHSupportingFeeOnTransferTokens(s.ctx, removeBuilder, s.router, s.from, s.token0, liquidity, big.NewInt(0), predictedAmountETH, s.from, deadline)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.v2Router.RemoveLiquidityETHSupportingFeeOnTransferTokens(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test20RemoveLiquidityETHSupportingFeeOnTransferTokens: broadcast tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "removeLiquidityETHSupportingFeeOnTransferTokens transaction reverted")
	s.T().Logf("Test20RemoveLiquidityETHSupportingFeeOnTransferTokens: confirmed in block %d", receipt.BlockNumber)

	burns, err := ExtractBurn(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(burns, 1)
	s.T().Logf("Test20RemoveLiquidityETHSupportingFeeOnTransferTokens: Burn sender=%s amount0=%s amount1=%s to=%s", burns[0].Sender, burns[0].Amount0, burns[0].Amount1, burns[0].To)

	withdrawals, err := weth.ExtractWithdrawals(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(withdrawals, 1, "removeLiquidityETHSupportingFeeOnTransferTokens should have unwrapped WETH back to native before sending it out")
	s.Require().Zero(predictedAmountETH.Cmp(withdrawals[0].Wad))
	s.T().Logf("Test20RemoveLiquidityETHSupportingFeeOnTransferTokens: Withdrawal src=%s wad=%s", withdrawals[0].Src, withdrawals[0].Wad)

	var tokenAfter, remainingLiquidity *big.Int
	s.Require().Eventually(func() bool {
		tokenAfter, err = s.erc20.BalanceOf(s.ctx, s.token0, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		remainingLiquidity, err = s.v2Pair.BalanceOf(s.ctx, s.pairETH, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		return tokenAfter.Cmp(tokenBefore) > 0 && remainingLiquidity.Sign() == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not visible after removeLiquidityETHSupportingFeeOnTransferTokens")
	s.T().Logf("Test20RemoveLiquidityETHSupportingFeeOnTransferTokens: token0 %s -> %s, remaining LP balance=%s", tokenBefore, tokenAfter, remainingLiquidity)
}

func (s *RouterLifeCycleSuite) Test21RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens() {
	pairToken0, err := s.v2Pair.Token0(s.ctx, s.pairETH, rpc.BlockTagLatest)
	s.Require().NoError(err)
	currentReserves, err := s.v2Pair.GetReserves(s.ctx, s.pairETH, rpc.BlockTagLatest)
	s.Require().NoError(err)
	tokenReserve, wethReserve := currentReserves.Reserve0, currentReserves.Reserve1
	if !pairToken0.Equal(s.token0) {
		tokenReserve, wethReserve = currentReserves.Reserve1, currentReserves.Reserve0
	}

	tokenDesired := s.liquidityAmount
	if tokenReserve.Sign() > 0 && wethReserve.Sign() > 0 {
		tokenDesired = new(big.Int).Div(new(big.Int).Mul(s.nativeValue, tokenReserve), wethReserve)
	}
	s.T().Logf("Test21RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens: re-adding liquidity matching tokenReserve=%s wethReserve=%s -> tokenDesired=%s nativeValue=%s", tokenReserve, wethReserve, tokenDesired, s.nativeValue)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.token0, s.from, s.router, tokenDesired)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	var tokenAllowance *big.Int
	s.Require().Eventually(func() bool {
		tokenAllowance, err = s.erc20.Allowance(s.ctx, s.token0, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && tokenDesired.Cmp(tokenAllowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")

	addDeadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
	addBuilder := tx.NewDynamicFeeBuilder()
	addBuilder.PriceBuffer(s.PriceBuffer)
	addBuilder.LimitBuffer(s.LimitBuffer)
	unsignedAdd, err := s.v2Router.CreateAddLiquidityETH(s.ctx, addBuilder, s.router, s.from, s.token0, tokenDesired, big.NewInt(0), big.NewInt(0), s.from, addDeadline, s.nativeValue)
	s.Require().NoError(err)
	s.Require().NoError(unsignedAdd.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedAdd.MaxFee())

	addHash, err := s.v2Router.AddLiquidityETH(s.ctx, unsignedAdd)
	s.Require().NoError(err)
	s.T().Logf("Test21RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens: re-added liquidity to pairETH, broadcast tx %s", addHash)

	addReceipt, err := tx.WaitForReceipt(s.ctx, s.client, addHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(addReceipt.Status, "addLiquidityETH transaction reverted")
	s.T().Logf("Test21RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens: re-add liquidity confirmed in block %d", addReceipt.BlockNumber)

	var liquidity *big.Int
	s.Require().Eventually(func() bool {
		liquidity, err = s.v2Pair.BalanceOf(s.ctx, s.pairETH, s.from, rpc.BlockTagLatest)
		return err == nil && liquidity.Sign() > 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "LP balance not visible after re-adding liquidity")
	s.T().Logf("Test21RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens: pairETH=%s LP balance=%s", s.pairETH, liquidity)

	nonce, err := s.v2Pair.Nonces(s.ctx, s.pairETH, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	domain := eip712.NewDomain("Uniswap V2", erc2612.DefaultVersion, s.chainID, s.pairETH, nil)
	permit := erc2612.NewPermit(s.from, s.router, liquidity, nonce, deadline)
	sig, err := permit.Sign(domain, s.signer)
	s.Require().NoError(err)
	s.T().Logf("Test21RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens: signed permit nonce=%s deadline=%s (no Approve transaction needed)", nonce, deadline)

	predictedAmountETH, err := s.v2Router.CallRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(s.ctx, s.router, s.from, s.token0, liquidity, big.NewInt(0), big.NewInt(0), s.from, deadline, false, sig, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(predictedAmountETH.Sign() > 0)
	s.T().Logf("Test21RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens: CallRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens predicted amountETH=%s", predictedAmountETH)

	tokenBefore, err := s.erc20.BalanceOf(s.ctx, s.token0, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)

	removeBuilder := tx.NewDynamicFeeBuilder()
	removeBuilder.PriceBuffer(s.PriceBuffer)
	removeBuilder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.v2Router.CreateRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(s.ctx, removeBuilder, s.router, s.from, s.token0, liquidity, big.NewInt(0), predictedAmountETH, s.from, deadline, false, sig)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.v2Router.RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test21RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens: broadcast tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "removeLiquidityETHWithPermitSupportingFeeOnTransferTokens transaction reverted")
	s.T().Logf("Test21RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens: confirmed in block %d", receipt.BlockNumber)

	burns, err := ExtractBurn(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(burns, 1)
	s.T().Logf("Test21RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens: Burn sender=%s amount0=%s amount1=%s to=%s", burns[0].Sender, burns[0].Amount0, burns[0].Amount1, burns[0].To)

	withdrawals, err := weth.ExtractWithdrawals(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(withdrawals, 1, "removeLiquidityETHWithPermitSupportingFeeOnTransferTokens should have unwrapped WETH back to native before sending it out")
	s.Require().Zero(predictedAmountETH.Cmp(withdrawals[0].Wad))
	s.T().Logf("Test21RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens: Withdrawal src=%s wad=%s", withdrawals[0].Src, withdrawals[0].Wad)

	var tokenAfter, nonceAfter, remainingLiquidity *big.Int
	s.Require().Eventually(func() bool {
		tokenAfter, err = s.erc20.BalanceOf(s.ctx, s.token0, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		remainingLiquidity, err = s.v2Pair.BalanceOf(s.ctx, s.pairETH, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		nonceAfter, err = s.v2Pair.Nonces(s.ctx, s.pairETH, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		return tokenAfter.Cmp(tokenBefore) > 0 && remainingLiquidity.Sign() == 0 &&
			new(big.Int).Add(nonce, big.NewInt(1)).Cmp(nonceAfter) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not visible after removeLiquidityETHWithPermitSupportingFeeOnTransferTokens")
	s.Require().Zero(new(big.Int).Add(nonce, big.NewInt(1)).Cmp(nonceAfter), "permit nonce should have advanced by exactly 1")
	s.T().Logf("Test21RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens: token0 %s -> %s, nonce %s -> %s, remaining LP balance=%s", tokenBefore, tokenAfter, nonce, nonceAfter, remainingLiquidity)
}

func (s *RouterLifeCycleSuite) Test22QuoteInsufficientAmount() {
	_, err := s.v2Router.Quote(s.ctx, s.router, big.NewInt(0), big.NewInt(1000), big.NewInt(1000), rpc.BlockTagLatest)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInsufficientAmount)
	s.T().Logf("Test22QuoteInsufficientAmount: amountA=0 rejected as expected: %v", err)
}

func (s *RouterLifeCycleSuite) Test23QuoteInsufficientLiquidity() {
	_, err := s.v2Router.Quote(s.ctx, s.router, big.NewInt(1000), big.NewInt(0), big.NewInt(1000), rpc.BlockTagLatest)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInsufficientLiquidity)
	s.T().Logf("Test23QuoteInsufficientLiquidity: reserveA=0 rejected as expected: %v", err)
}

func (s *RouterLifeCycleSuite) Test24AddLiquidityIdenticalAddresses() {
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	_, err := s.v2Router.CreateAddLiquidity(s.ctx, builder, s.router, s.from, s.token0, s.token0, big.NewInt(1000), big.NewInt(1000), big.NewInt(0), big.NewInt(0), s.from, deadline)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrIdenticalAddresses)
	s.T().Logf("Test24AddLiquidityIdenticalAddresses: tokenA==tokenB==%s rejected as expected: %v", s.token0, err)
}

func (s *RouterLifeCycleSuite) Test25SwapInsufficientOutputAmount() {
	path := []*types.Address{s.token0, s.token1}
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.token0, s.from, s.router, s.swapAmount)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, s.token0, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && s.swapAmount.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")

	absurdMin := new(big.Int).Exp(big.NewInt(10), big.NewInt(30), nil)

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(s.PriceBuffer)
	swapBuilder.LimitBuffer(s.LimitBuffer)
	_, err = s.v2Router.CreateSwapExactTokensForTokens(s.ctx, swapBuilder, s.router, s.from, s.swapAmount, absurdMin, path, s.from, deadline)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInsufficientOutputAmount)
	s.T().Logf("Test25SwapInsufficientOutputAmount: amountOutMin=%s rejected as expected: %v", absurdMin, err)
}

func (s *RouterLifeCycleSuite) Test26SwapExcessiveInputAmount() {
	path := []*types.Address{s.token0, s.token1}
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	amountsIn, err := s.v2Router.GetAmountsIn(s.ctx, s.router, s.swapAmount, path, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(amountsIn, 2)
	tooLowMax := new(big.Int).Div(amountsIn[0], big.NewInt(2))
	s.T().Logf("Test26SwapExcessiveInputAmount: needed=%s amountInMax(too low)=%s", amountsIn[0], tooLowMax)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	_, err = s.v2Router.CreateSwapTokensForExactTokens(s.ctx, builder, s.router, s.from, s.swapAmount, tooLowMax, path, s.from, deadline)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrExcessiveInputAmount)
	s.T().Logf("Test26SwapExcessiveInputAmount: rejected as expected: %v", err)
}

func (s *RouterLifeCycleSuite) Test27AddLiquidityInsufficientBAmount() {
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	amountADesired := big.NewInt(1000)
	amountBDesired := new(big.Int).Exp(big.NewInt(10), big.NewInt(30), nil)
	absurdBMin := new(big.Int).Exp(big.NewInt(10), big.NewInt(29), nil)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	_, err := s.v2Router.CreateAddLiquidity(s.ctx, builder, s.router, s.from, s.token0, s.token1, amountADesired, amountBDesired, big.NewInt(0), absurdBMin, s.from, deadline)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInsufficientBAmount)
	s.T().Logf("Test27AddLiquidityInsufficientBAmount: amountBMin=%s rejected as expected: %v", absurdBMin, err)
}

func (s *RouterLifeCycleSuite) Test28AddLiquidityInsufficientAAmount() {
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
	amountADesired := new(big.Int).Exp(big.NewInt(10), big.NewInt(30), nil)
	amountBDesired := big.NewInt(1000)
	absurdAMin := new(big.Int).Exp(big.NewInt(10), big.NewInt(29), nil)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	_, err := s.v2Router.CreateAddLiquidity(s.ctx, builder, s.router, s.from, s.token0, s.token1, amountADesired, amountBDesired, absurdAMin, big.NewInt(0), s.from, deadline)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInsufficientAAmount)
	s.T().Logf("Test28AddLiquidityInsufficientAAmount: amountAMin=%s rejected as expected: %v", absurdAMin, err)
}

func (s *RouterLifeCycleSuite) Test29MultihopSwap() {
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
	for _, token := range []*types.Address{s.token1, s.token2} {
		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(s.PriceBuffer)
		builder.LimitBuffer(s.LimitBuffer)
		unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, token, s.from, s.router, s.liquidityAmount)
		s.Require().NoError(err)
		s.Require().NoError(unsignedApprove.Sign(s.signer))
		s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())

		approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
		s.Require().NoError(err)

		approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(approveReceipt.Status, "approve transaction reverted")
		var allowance *big.Int
		s.Require().Eventually(func() bool {
			allowance, err = s.erc20.Allowance(s.ctx, token, s.from, s.router, rpc.BlockTagLatest)
			return err == nil && s.liquidityAmount.Cmp(allowance) == 0
		}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	}

	addBuilder := tx.NewDynamicFeeBuilder()
	addBuilder.PriceBuffer(s.PriceBuffer)
	addBuilder.LimitBuffer(s.LimitBuffer)
	unsignedAdd, err := s.v2Router.CreateAddLiquidity(s.ctx, addBuilder, s.router, s.from, s.token1, s.token2, s.liquidityAmount, s.liquidityAmount, big.NewInt(0), big.NewInt(0), s.from, deadline)
	s.Require().NoError(err)
	s.Require().NoError(unsignedAdd.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedAdd.MaxFee())

	addHash, err := s.v2Router.AddLiquidity(s.ctx, unsignedAdd)
	s.Require().NoError(err)
	s.T().Logf("Test29MultihopSwap: created (token1, token2) pair, broadcast tx %s", addHash)

	addReceipt, err := tx.WaitForReceipt(s.ctx, s.client, addHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(addReceipt.Status, "addLiquidity transaction reverted")
	s.T().Logf("Test29MultihopSwap: (token1, token2) addLiquidity confirmed in block %d", addReceipt.BlockNumber)

	created, err := ExtractPairCreated(addReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(created, 1, "addLiquidity should have created the (token1, token2) pair itself, via the factory")
	s.pair12 = created[0].Pair
	s.T().Logf("Test29MultihopSwap: factory created pair %s for (%s, %s)", s.pair12, s.token1, s.token2)

	var pair12Reserves *Reserves
	s.Require().Eventually(func() bool {
		pair12Reserves, err = s.v2Pair.GetReserves(s.ctx, s.pair12, rpc.BlockTagLatest)
		return err == nil && pair12Reserves.Reserve0.Sign() > 0 && pair12Reserves.Reserve1.Sign() > 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "reserves not visible after addLiquidity")
	s.T().Logf("Test29MultihopSwap: pair12 %s reserves reserve0=%s reserve1=%s", s.pair12, pair12Reserves.Reserve0, pair12Reserves.Reserve1)

	path := []*types.Address{s.token0, s.token1, s.token2}

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.token0, s.from, s.router, s.swapAmount)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	var swapAllowance *big.Int
	s.Require().Eventually(func() bool {
		swapAllowance, err = s.erc20.Allowance(s.ctx, s.token0, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && s.swapAmount.Cmp(swapAllowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")

	predicted, err := s.v2Router.CallSwapExactTokensForTokens(s.ctx, s.router, s.from, s.swapAmount, big.NewInt(0), path, s.from, deadline, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(predicted, 3, "a 2-hop path should produce 3 amounts: in, intermediate, out")
	s.T().Logf("Test29MultihopSwap: CallSwapExactTokensForTokens predicted amounts=%v (path=[%s, %s, %s])", predicted, path[0], path[1], path[2])

	before, err := s.erc20.BalanceOf(s.ctx, s.token2, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(s.PriceBuffer)
	swapBuilder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.v2Router.CreateSwapExactTokensForTokens(s.ctx, swapBuilder, s.router, s.from, s.swapAmount, predicted[2], path, s.from, deadline)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.v2Router.SwapExactTokensForTokens(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test29MultihopSwap: broadcast swapExactTokensForTokens tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swapExactTokensForTokens transaction reverted")
	s.T().Logf("Test29MultihopSwap: swapExactTokensForTokens confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 2, "a 2-hop path should emit one Swap per pair it crosses")

	var hopOnPair, hopOnPair12 *Swap
	for i := range swaps {
		switch {
		case swaps[i].Contract.Equal(s.pair):
			hopOnPair = &swaps[i]
		case swaps[i].Contract.Equal(s.pair12):
			hopOnPair12 = &swaps[i]
		}
	}
	s.Require().NotNil(hopOnPair, "expected a Swap on (token0, token1) pair %s", s.pair)
	s.Require().NotNil(hopOnPair12, "expected a Swap on (token1, token2) pair %s", s.pair12)
	s.T().Logf("Test29MultihopSwap: hop1 (pair %s) sender=%s to=%s amount0In=%s amount1In=%s amount0Out=%s amount1Out=%s",
		hopOnPair.Contract, hopOnPair.Sender, hopOnPair.To, hopOnPair.Amount0In, hopOnPair.Amount1In, hopOnPair.Amount0Out, hopOnPair.Amount1Out)
	s.T().Logf("Test29MultihopSwap: hop2 (pair12 %s) sender=%s to=%s amount0In=%s amount1In=%s amount0Out=%s amount1Out=%s",
		hopOnPair12.Contract, hopOnPair12.Sender, hopOnPair12.To, hopOnPair12.Amount0In, hopOnPair12.Amount1In, hopOnPair12.Amount0Out, hopOnPair12.Amount1Out)

	hop1Out := hopOnPair.Amount0Out
	if hop1Out.Sign() == 0 {
		hop1Out = hopOnPair.Amount1Out
	}
	s.Require().Zero(predicted[1].Cmp(hop1Out))
	s.Require().True(hopOnPair.To.Equal(s.pair12))

	hop2Out := hopOnPair12.Amount0Out
	if hop2Out.Sign() == 0 {
		hop2Out = hopOnPair12.Amount1Out
	}
	s.Require().Zero(predicted[2].Cmp(hop2Out))
	s.Require().True(hopOnPair12.To.Equal(s.from))

	var after *big.Int
	s.Require().Eventually(func() bool {
		after, err = s.erc20.BalanceOf(s.ctx, s.token2, s.from, rpc.BlockTagLatest)
		return err == nil && new(big.Int).Sub(after, before).Cmp(predicted[2]) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "token2 balance not visible after swap")
	s.T().Logf("Test29MultihopSwap: token2 balance %s -> %s", before, after)
}

type RouterSwapSuite struct {
	suite.Suite

	Enable          bool
	RPCURL          string
	PrivateKeyHex   string
	RouterHex       string
	TokenInHex      string
	TokenOutHex     string
	HopsHex         []string
	NativeIn        bool
	NativeOut       bool
	MulticallHex    string
	AmountIn        string
	NativeSymbol    string
	RoundTrip       bool
	Interval        int
	Timeout         int
	DeadlineSeconds int
	PriceBuffer     uint64
	LimitBuffer     uint64

	ctx    context.Context
	cancel context.CancelFunc

	client    rpc.Client
	signer    core.Signer
	from      *types.Address
	router    *types.Address
	wethAddr  *types.Address
	tokenIn   *types.Address
	tokenOut  *types.Address
	hops      []*types.Address
	multicall *types.Address
	amount    *big.Int

	v2Router *Router
	erc20    *erc20.ERC20

	totalFee *big.Int
}

func TestRouterSwapSuite(t *testing.T) {
	suite.Run(t, &RouterSwapSuite{
		Enable:          false,
		RoundTrip:       false,
		RPCURL:          "",
		PrivateKeyHex:   "",
		RouterHex:       "",
		TokenInHex:      "",
		TokenOutHex:     "",
		HopsHex:         []string{},
		NativeIn:        false,
		NativeOut:       true,
		MulticallHex:    multicall3.DefaultMulticall3Address.String(),
		AmountIn:        "",
		NativeSymbol:    "",
		Interval:        200,
		Timeout:         60,
		DeadlineSeconds: 600,
		PriceBuffer:     150,
		LimitBuffer:     150,
	})
}

func (s *RouterSwapSuite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Minute)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.v2Router = NewRouter(s.client)
	s.erc20 = erc20.New(s.client)

	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())

	s.router, err = types.NewAddressFromHex(s.RouterHex)
	s.Require().NoError(err)

	s.Require().False(s.NativeIn && s.NativeOut, "NativeIn and NativeOut can't both be set -- that's not a swap")

	s.wethAddr, err = s.v2Router.WETH(s.ctx, s.router, rpc.BlockTagLatest)
	s.Require().NoError(err)

	if s.NativeIn {
		s.tokenIn = s.wethAddr
	} else {
		s.tokenIn, err = types.NewAddressFromHex(s.TokenInHex)
		s.Require().NoError(err)
	}

	if s.NativeOut {
		s.tokenOut = s.wethAddr
	} else {
		s.tokenOut, err = types.NewAddressFromHex(s.TokenOutHex)
		s.Require().NoError(err)
	}

	s.hops = make([]*types.Address, 0, len(s.HopsHex))
	for _, h := range s.HopsHex {
		hop, err := types.NewAddressFromHex(h)
		s.Require().NoError(err)
		s.hops = append(s.hops, hop)
	}

	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)

	amount, ok := new(big.Int).SetString(s.AmountIn, 10)
	s.Require().True(ok)
	s.amount = amount

	path := append([]*types.Address{s.tokenIn}, s.hops...)
	path = append(path, s.tokenOut)
	amountsOut, err := s.v2Router.GetAmountsOut(s.ctx, s.router, s.amount, path, rpc.BlockTagLatest)
	s.Require().NoError(err, "no route from tokenIn to tokenOut via this router")
	s.Require().Len(amountsOut, len(path))
}

func (s *RouterSwapSuite) TearDownSuite() {
	if s.totalFee != nil {
		base := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
		whole, frac := new(big.Int), new(big.Int)
		whole.DivMod(s.totalFee, base, frac)

		fracStr := strings.TrimRight(fmt.Sprintf("%018s", frac.String()), "0")
		decimal := whole.String()
		if fracStr != "" {
			decimal += "." + fracStr
		}

		s.T().Logf("RouterSwapSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *RouterSwapSuite) TestSwap() {
	tokenInMeta, err := s.erc20.MetadataWithBalance(s.ctx, s.multicall, s.tokenIn, s.from, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("TestSwap: TokenIn %s", s.tokenIn)
	s.T().Logf("\t %s (%s), decimals: %d", tokenInMeta.Name, tokenInMeta.Symbol, tokenInMeta.Decimals)
	s.T().Logf("\t %s amounts: %s", s.tokenIn, tokenInMeta.Balance.String())

	tokenOutMeta, err := s.erc20.MetadataWithBalance(s.ctx, s.multicall, s.tokenOut, s.from, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("TestSwap: TokenOut %s", s.tokenOut)
	s.T().Logf("\t %s (%s), decimals: %d", tokenOutMeta.Name, tokenOutMeta.Symbol, tokenOutMeta.Decimals)
	s.T().Logf("\t %s amounts: %s", s.tokenOut, tokenOutMeta.Balance.String())

	forwardPath := append([]*types.Address{s.tokenIn}, s.hops...)
	forwardPath = append(forwardPath, s.tokenOut)

	reversePath := make([]*types.Address, len(forwardPath))
	for i, addr := range forwardPath {
		reversePath[len(forwardPath)-1-i] = addr
	}

	received := s.swap(forwardPath, s.amount, s.NativeIn, s.NativeOut)
	if !s.RoundTrip {
		return
	}

	final := s.swap(reversePath, received, s.NativeOut, s.NativeIn)
	s.T().Logf("TestSwap: round trip %s -> %s -> %s: sent %s, received back %s", s.tokenIn, s.tokenOut, s.tokenIn, s.amount, final)

	if s.NativeOut {
		return
	}
	tokenOutBalanceAfter, err := s.erc20.BalanceOf(s.ctx, s.tokenOut, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(tokenOutBalanceAfter.Cmp(tokenOutMeta.Balance), "wallet's TokenOut balance changed (%s -> %s) after swapping it all back", tokenOutMeta.Balance, tokenOutBalanceAfter)
}

func (s *RouterSwapSuite) swap(path []*types.Address, amountIn *big.Int, nativeIn, nativeOut bool) *big.Int {
	tokenIn, tokenOut := path[0], path[len(path)-1]

	if nativeIn {
		var hex string
		s.Require().NoError(s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.from.String(), rpc.BlockTagLatest, &hex)))
		balance, ok := new(big.Int).SetString(strings.TrimPrefix(hex, "0x"), 16)
		s.Require().True(ok)
		s.Require().True(balance.Cmp(amountIn) >= 0, "wallet holds %s native, less than amountIn %s", balance, amountIn)
		s.T().Logf("swap %s -> %s: wallet native balance: %s", tokenIn, tokenOut, balance)
	} else {
		balance, err := s.erc20.BalanceOf(s.ctx, tokenIn, s.from, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().True(balance.Cmp(amountIn) >= 0, "wallet holds %s of tokenIn, less than amountIn %s", balance, amountIn)
		s.T().Logf("swap %s -> %s: wallet balance of tokenIn: %s", tokenIn, tokenOut, balance)

		approveBuilder := tx.NewDynamicFeeBuilder()
		approveBuilder.PriceBuffer(s.PriceBuffer)
		approveBuilder.LimitBuffer(s.LimitBuffer)
		unsignedApprove, err := s.erc20.CreateApprove(s.ctx, approveBuilder, tokenIn, s.from, s.router, amountIn)
		s.Require().NoError(err)
		s.Require().NoError(unsignedApprove.Sign(s.signer))
		s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())

		approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
		s.Require().NoError(err)
		s.T().Logf("swap %s -> %s: approve %s to spend %s, broadcast tx %s", tokenIn, tokenOut, s.router, amountIn, approveHash)

		approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(approveReceipt.Status, "approve transaction reverted")
		var allowance *big.Int
		s.Require().Eventually(func() bool {
			allowance, err = s.erc20.Allowance(s.ctx, tokenIn, s.from, s.router, rpc.BlockTagLatest)
			return err == nil && amountIn.Cmp(allowance) == 0
		}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
		s.T().Logf("swap %s -> %s: approve confirmed in block %d", tokenIn, tokenOut, approveReceipt.BlockNumber)
	}

	var amountOutMin *big.Int
	var unsignedSwap core.Transaction
	var lastErr error
	attempt := 0
	s.Require().Eventually(func() bool {
		attempt++
		amountsOut, err := s.v2Router.GetAmountsOut(s.ctx, s.router, amountIn, path, rpc.BlockTagLatest)
		if err != nil {
			lastErr = err
			return false
		}
		amountOutLast := amountsOut[len(amountsOut)-1]
		amountOutMin = new(big.Int).Div(new(big.Int).Mul(amountOutLast, big.NewInt(995)), big.NewInt(1000))
		s.T().Logf("swap %s -> %s: attempt %d: GetAmountsOut=%s amountOutMin=%s", tokenIn, tokenOut, attempt, amountOutLast, amountOutMin)

		deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(s.PriceBuffer)
		builder.LimitBuffer(s.LimitBuffer)

		var u core.Transaction
		switch {
		case nativeIn:
			u, err = s.v2Router.CreateSwapExactETHForTokens(s.ctx, builder, s.router, s.from, amountOutMin, path, s.from, deadline, amountIn)
		case nativeOut:
			u, err = s.v2Router.CreateSwapExactTokensForETH(s.ctx, builder, s.router, s.from, amountIn, amountOutMin, path, s.from, deadline)
		default:
			u, err = s.v2Router.CreateSwapExactTokensForTokens(s.ctx, builder, s.router, s.from, amountIn, amountOutMin, path, s.from, deadline)
		}
		if err != nil {
			lastErr = err
			s.T().Logf("swap %s -> %s: attempt %d failed: %v", tokenIn, tokenOut, attempt, err)
			return false
		}
		unsignedSwap = u
		return true
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "swap not buildable: %v", lastErr)
	s.Require().NoError(unsignedSwap.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedSwap.MaxFee())
	s.T().Logf("swap %s -> %s: swap tx built after %d attempt(s), amountOutMin=%s", tokenIn, tokenOut, attempt, amountOutMin)

	var swapHash *types.Hash
	var err error
	switch {
	case nativeIn:
		swapHash, err = s.v2Router.SwapExactETHForTokens(s.ctx, unsignedSwap)
	case nativeOut:
		swapHash, err = s.v2Router.SwapExactTokensForETH(s.ctx, unsignedSwap)
	default:
		swapHash, err = s.v2Router.SwapExactTokensForTokens(s.ctx, unsignedSwap)
	}
	s.Require().NoError(err)
	s.T().Logf("swap %s -> %s: broadcast swap, tx %s", tokenIn, tokenOut, swapHash)

	swapReceipt, err := tx.WaitForReceipt(s.ctx, s.client, swapHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(swapReceipt.Status, "swap transaction reverted")
	s.T().Logf("swap %s -> %s: swap confirmed in block %d", tokenIn, tokenOut, swapReceipt.BlockNumber)

	swaps, err := ExtractSwap(swapReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, len(path)-1)

	transfers, err := erc20.ExtractTransfers(swapReceipt.Logs)
	s.Require().NoError(err)

	lastSwap := swaps[len(swaps)-1]
	actualAmountOut := lastSwap.Amount1Out
	if actualAmountOut.Sign() == 0 {
		actualAmountOut = lastSwap.Amount0Out
	}

	s.T().Logf("Router swap %s -> %s Result", tokenIn, tokenOut)
	s.T().Logf("\tAmountIn: %s", amountIn)
	s.T().Logf("\tAmountOut: %s", actualAmountOut)
	s.T().Logf("\tTx: %s", swapHash)
	s.T().Logf("\tLogs:")
	for _, transfer := range transfers {
		s.T().Logf("\t\tTransfer: contract=%s from=%s to=%s value=%s", transfer.Contract, transfer.From, transfer.To, transfer.Value)
	}
	if nativeIn {
		deposits, err := weth.ExtractDeposits(swapReceipt.Logs)
		s.Require().NoError(err)
		for _, d := range deposits {
			s.T().Logf("\t\tDeposit: dst=%s wad=%s", d.Dst, d.Wad)
		}
	}
	if nativeOut {
		withdrawals, err := weth.ExtractWithdrawals(swapReceipt.Logs)
		s.Require().NoError(err)
		for _, w := range withdrawals {
			s.T().Logf("\t\tWithdrawal: src=%s wad=%s", w.Src, w.Wad)
		}
	}
	for _, sw := range swaps {
		s.T().Logf("\t\tSwap: contract=%s sender=%s to=%s amount0In=%s amount1In=%s amount0Out=%s amount1Out=%s",
			sw.Contract, sw.Sender, sw.To, sw.Amount0In, sw.Amount1In, sw.Amount0Out, sw.Amount1Out)
	}

	return actualAmountOut
}
