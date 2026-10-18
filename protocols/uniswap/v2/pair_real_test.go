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
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-evmkit/standards/multicall3"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

type RealSwapSuite struct {
	suite.Suite

	Enable        bool
	RPCURL        string
	PrivateKeyHex string
	FactoryHex    string
	TokenInHex    string
	TokenOutHex   string
	MulticallHex  string
	AmountIn      string
	NativeSymbol  string
	RoundTrip     bool
	Interval      int
	Timeout       int

	ctx    context.Context
	cancel context.CancelFunc

	client    rpc.Client
	signer    core.Signer
	from      *types.Address
	factory   *types.Address
	pair      *types.Address
	tokenIn   *types.Address
	tokenOut  *types.Address
	multicall *types.Address
	amount    *big.Int

	v2Factory *Factory
	v2Pair    *Pair
	erc20     *erc20.ERC20

	totalFee *big.Int
}

func TestRealSwapSuite(t *testing.T) {
	suite.Run(t, &RealSwapSuite{
		Enable:        false,
		RoundTrip:     false,
		RPCURL:        "",
		PrivateKeyHex: "",
		FactoryHex:    "",
		TokenInHex:    "",
		TokenOutHex:   "",
		MulticallHex:  multicall3.DefaultMulticall3Address.String(),
		AmountIn:      "1000000000000000",
		NativeSymbol:  "",
		Interval:      200,
		Timeout:       60,
	})
}

func (s *RealSwapSuite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Minute)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.v2Factory = NewFactory(s.client)
	s.v2Pair = NewPair(s.client)
	s.erc20 = erc20.New(s.client)

	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())

	s.factory, err = types.NewAddressFromHex(s.FactoryHex)
	s.Require().NoError(err)

	s.tokenIn, err = types.NewAddressFromHex(s.TokenInHex)
	s.Require().NoError(err)

	s.tokenOut, err = types.NewAddressFromHex(s.TokenOutHex)
	s.Require().NoError(err)

	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)

	amount, ok := new(big.Int).SetString(s.AmountIn, 10)
	s.Require().True(ok)
	s.amount = amount

	s.pair, err = s.v2Factory.GetPair(s.ctx, s.factory, s.tokenIn, s.tokenOut, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().False(s.pair.IsZero(), "no pair exists for (%s, %s)", s.tokenIn, s.tokenOut)
}

func (s *RealSwapSuite) TearDownSuite() {
	if s.totalFee != nil {
		base := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
		whole, frac := new(big.Int), new(big.Int)
		whole.DivMod(s.totalFee, base, frac)

		fracStr := strings.TrimRight(fmt.Sprintf("%018s", frac.String()), "0")
		decimal := whole.String()
		if fracStr != "" {
			decimal += "." + fracStr
		}

		s.T().Logf("RealSwapSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *RealSwapSuite) TestSwap() {
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

	received := s.swap(s.tokenIn, s.tokenOut, s.amount)
	if !s.RoundTrip {
		return
	}

	final := s.swap(s.tokenOut, s.tokenIn, received)
	s.T().Logf("TestSwap: round trip %s -> %s -> %s: sent %s, received back %s", s.tokenIn, s.tokenOut, s.tokenIn, s.amount, final)

	tokenOutBalanceAfter, err := s.erc20.BalanceOf(s.ctx, s.tokenOut, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(tokenOutBalanceAfter.Cmp(tokenOutMeta.Balance), "wallet's TokenOut balance changed (%s -> %s) after swapping it all back", tokenOutMeta.Balance, tokenOutBalanceAfter)
}

func (s *RealSwapSuite) swap(tokenIn, tokenOut *types.Address, amountIn *big.Int) *big.Int {
	token0, err := s.v2Pair.Token0(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)
	token1, err := s.v2Pair.Token1(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)

	inIsToken0 := tokenIn.Equal(token0)
	if inIsToken0 {
		s.Require().True(tokenOut.Equal(token1), "tokenOut %s doesn't match the pair's token1 %s", tokenOut, token1)
	} else {
		s.Require().True(tokenIn.Equal(token1), "tokenIn %s is neither of the pair's tokens (%s, %s)", tokenIn, token0, token1)
		s.Require().True(tokenOut.Equal(token0), "tokenOut %s doesn't match the pair's token0 %s", tokenOut, token0)
	}
	s.T().Logf("swap %s -> %s: token0=%s token1=%s (tokenIn is token0: %v)", tokenIn, tokenOut, token0, token1, inIsToken0)

	balance, err := s.erc20.BalanceOf(s.ctx, tokenIn, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(balance.Cmp(amountIn) >= 0, "wallet holds %s of tokenIn, less than amountIn %s", balance, amountIn)
	s.T().Logf("swap %s -> %s: wallet balance of tokenIn: %s", tokenIn, tokenOut, balance)

	unsignedTransfer, err := s.erc20.CreateTransfer(s.ctx, tx.NewDynamicFeeBuilder(), tokenIn, s.from, s.pair, amountIn)
	s.Require().NoError(err)
	s.Require().NoError(unsignedTransfer.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedTransfer.MaxFee())

	transferHash, err := s.erc20.Transfer(s.ctx, unsignedTransfer)
	s.Require().NoError(err)
	s.T().Logf("swap %s -> %s: broadcast transfer of %s tokenIn to pair, tx %s", tokenIn, tokenOut, amountIn, transferHash)

	transferReceipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(transferReceipt.Status, "transfer transaction reverted")
	s.T().Logf("swap %s -> %s: transfer confirmed in block %d", tokenIn, tokenOut, transferReceipt.BlockNumber)

	var amountOut, amount0Out, amount1Out *big.Int
	var unsignedSwap core.Transaction
	var lastErr error
	attempt := 0
	s.Require().Eventually(func() bool {
		attempt++
		reserves, err := s.v2Pair.GetReserves(s.ctx, s.pair, rpc.BlockTagLatest)
		if err != nil {
			lastErr = err
			return false
		}

		reserveIn, reserveOut := reserves.Reserve0, reserves.Reserve1
		if !inIsToken0 {
			reserveIn, reserveOut = reserves.Reserve1, reserves.Reserve0
		}

		amountInWithFee := new(big.Int).Mul(amountIn, big.NewInt(997))
		numerator := new(big.Int).Mul(amountInWithFee, reserveOut)
		denominator := new(big.Int).Add(new(big.Int).Mul(reserveIn, big.NewInt(1000)), amountInWithFee)
		amountOut = new(big.Int).Div(numerator, denominator)
		amountOut = new(big.Int).Div(new(big.Int).Mul(amountOut, big.NewInt(995)), big.NewInt(1000))

		amount0Out, amount1Out = big.NewInt(0), amountOut
		if !inIsToken0 {
			amount0Out, amount1Out = amountOut, big.NewInt(0)
		}

		s.T().Logf("swap %s -> %s: attempt %d: reserveIn=%s reserveOut=%s amountOut=%s", tokenIn, tokenOut, attempt, reserveIn, reserveOut, amountOut)

		u, err := s.v2Pair.CreateSwap(s.ctx, tx.NewDynamicFeeBuilder(), s.pair, s.from, amount0Out, amount1Out, s.from, nil)
		if err != nil {
			lastErr = err
			s.T().Logf("swap %s -> %s: attempt %d failed: %v", tokenIn, tokenOut, attempt, err)
			return false
		}
		unsignedSwap = u
		return true
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "deposited balance not visible before swap: %v", lastErr)
	s.Require().NoError(unsignedSwap.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedSwap.MaxFee())
	s.T().Logf("swap %s -> %s: CreateSwap succeeded after %d attempt(s), estimated amountOut=%s", tokenIn, tokenOut, attempt, amountOut)

	swapHash, err := s.v2Pair.Swap(s.ctx, unsignedSwap)
	s.Require().NoError(err)
	s.T().Logf("swap %s -> %s: broadcast swap, tx %s", tokenIn, tokenOut, swapHash)

	swapReceipt, err := tx.WaitForReceipt(s.ctx, s.client, swapHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(swapReceipt.Status, "swap transaction reverted")
	s.T().Logf("swap %s -> %s: swap confirmed in block %d", tokenIn, tokenOut, swapReceipt.BlockNumber)

	swaps, err := ExtractSwap(swapReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)

	syncs, err := ExtractSync(swapReceipt.Logs)
	s.Require().NoError(err)

	transfers, err := erc20.ExtractTransfers(swapReceipt.Logs)
	s.Require().NoError(err)

	actualAmountOut := swaps[0].Amount1Out
	if !inIsToken0 {
		actualAmountOut = swaps[0].Amount0Out
	}

	s.T().Logf("UniswapV2Pair %s swap %s -> %s Result", s.pair, tokenIn, tokenOut)
	s.T().Logf("\tAmountIn: %s", amountIn)
	s.T().Logf("\tAmountOut: %s", actualAmountOut)
	s.T().Logf("\tTx: %s", swapHash)

	s.T().Logf("\tLogs:")
	for _, transfer := range transfers {
		s.T().Logf("\t\tTransfer: contract=%s from=%s to=%s value=%s", transfer.Contract, transfer.From, transfer.To, transfer.Value)
	}
	for _, sync := range syncs {
		s.T().Logf("\t\tSync: reserve0=%s reserve1=%s", sync.Reserve0, sync.Reserve1)
	}
	for _, swap := range swaps {
		s.T().Logf("\t\tSwap: sender=%s to=%s amount0In=%s amount1In=%s amount0Out=%s amount1Out=%s",
			swap.Sender, swap.To, swap.Amount0In, swap.Amount1In, swap.Amount0Out, swap.Amount1Out)
	}

	return actualAmountOut
}
