package v2

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewAddLiquidityResult(t *testing.T) {
	amountA := big.NewInt(100)
	amountB := big.NewInt(200)
	liquidity := big.NewInt(300)

	r := NewAddLiquidityResult(amountA, amountB, liquidity)
	require.Same(t, amountA, r.AmountA)
	require.Same(t, amountB, r.AmountB)
	require.Same(t, liquidity, r.Liquidity)
}

func TestNewAddLiquidityETHResult(t *testing.T) {
	amountToken := big.NewInt(100)
	amountETH := big.NewInt(200)
	liquidity := big.NewInt(300)

	r := NewAddLiquidityETHResult(amountToken, amountETH, liquidity)
	require.Same(t, amountToken, r.AmountToken)
	require.Same(t, amountETH, r.AmountETH)
	require.Same(t, liquidity, r.Liquidity)
}

func TestNewRemoveLiquidityResult(t *testing.T) {
	amountA := big.NewInt(100)
	amountB := big.NewInt(200)

	r := NewRemoveLiquidityResult(amountA, amountB)
	require.Same(t, amountA, r.AmountA)
	require.Same(t, amountB, r.AmountB)
}

func TestNewRemoveLiquidityETHResult(t *testing.T) {
	amountToken := big.NewInt(100)
	amountETH := big.NewInt(200)

	r := NewRemoveLiquidityETHResult(amountToken, amountETH)
	require.Same(t, amountToken, r.AmountToken)
	require.Same(t, amountETH, r.AmountETH)
}

func TestRouterRequestsValidate(t *testing.T) {
	router := types.MustNewAddressFromHex("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sig, err := key.Sign(core.Keccak256([]byte("permit")))
	require.NoError(t, err)

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"weth nil", (*WETHRequest)(nil).Validate(), "invalid request: request is nil"},
		{"weth empty", (&WETHRequest{}).Validate(), "invalid request: contract is nil"},
		{"weth nil contract", (&WETHRequest{}).Validate(), "invalid request: contract is nil"},
		{"weth valid", (&WETHRequest{Contract: router}).Validate(), ""},
		{"quote nil", (*QuoteRequest)(nil).Validate(), "invalid request: request is nil"},
		{"quote empty", (&QuoteRequest{}).Validate(), "invalid request: contract is nil"},
		{"quote nil contract", (&QuoteRequest{AmountA: big.NewInt(1), ReserveA: big.NewInt(1), ReserveB: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"quote nil amountA", (&QuoteRequest{Contract: router, ReserveA: big.NewInt(1), ReserveB: big.NewInt(1)}).Validate(), "invalid request: amountA is nil"},
		{"quote nil reserveA", (&QuoteRequest{Contract: router, AmountA: big.NewInt(1), ReserveB: big.NewInt(1)}).Validate(), "invalid request: reserveA is nil"},
		{"quote nil reserveB", (&QuoteRequest{Contract: router, AmountA: big.NewInt(1), ReserveA: big.NewInt(1)}).Validate(), "invalid request: reserveB is nil"},
		{"quote valid", (&QuoteRequest{Contract: router, AmountA: big.NewInt(1), ReserveA: big.NewInt(1), ReserveB: big.NewInt(1)}).Validate(), ""},
		{"getAmountOut nil", (*GetAmountOutRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getAmountOut empty", (&GetAmountOutRequest{}).Validate(), "invalid request: contract is nil"},
		{"getAmountOut nil contract", (&GetAmountOutRequest{AmountIn: big.NewInt(1), ReserveIn: big.NewInt(1), ReserveOut: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"getAmountOut nil amountIn", (&GetAmountOutRequest{Contract: router, ReserveIn: big.NewInt(1), ReserveOut: big.NewInt(1)}).Validate(), "invalid request: amountIn is nil"},
		{"getAmountOut nil reserveIn", (&GetAmountOutRequest{Contract: router, AmountIn: big.NewInt(1), ReserveOut: big.NewInt(1)}).Validate(), "invalid request: reserveIn is nil"},
		{"getAmountOut nil reserveOut", (&GetAmountOutRequest{Contract: router, AmountIn: big.NewInt(1), ReserveIn: big.NewInt(1)}).Validate(), "invalid request: reserveOut is nil"},
		{"getAmountOut valid", (&GetAmountOutRequest{Contract: router, AmountIn: big.NewInt(1), ReserveIn: big.NewInt(1), ReserveOut: big.NewInt(1)}).Validate(), ""},
		{"getAmountIn nil", (*GetAmountInRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getAmountIn empty", (&GetAmountInRequest{}).Validate(), "invalid request: contract is nil"},
		{"getAmountIn nil contract", (&GetAmountInRequest{AmountOut: big.NewInt(1), ReserveIn: big.NewInt(1), ReserveOut: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"getAmountIn nil amountOut", (&GetAmountInRequest{Contract: router, ReserveIn: big.NewInt(1), ReserveOut: big.NewInt(1)}).Validate(), "invalid request: amountOut is nil"},
		{"getAmountIn nil reserveIn", (&GetAmountInRequest{Contract: router, AmountOut: big.NewInt(1), ReserveOut: big.NewInt(1)}).Validate(), "invalid request: reserveIn is nil"},
		{"getAmountIn nil reserveOut", (&GetAmountInRequest{Contract: router, AmountOut: big.NewInt(1), ReserveIn: big.NewInt(1)}).Validate(), "invalid request: reserveOut is nil"},
		{"getAmountIn valid", (&GetAmountInRequest{Contract: router, AmountOut: big.NewInt(1), ReserveIn: big.NewInt(1), ReserveOut: big.NewInt(1)}).Validate(), ""},
		{"getAmountsOut nil", (*GetAmountsOutRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getAmountsOut empty", (&GetAmountsOutRequest{}).Validate(), "invalid request: contract is nil"},
		{"getAmountsOut nil contract", (&GetAmountsOutRequest{AmountIn: big.NewInt(1), Path: []*types.Address{a, b}}).Validate(), "invalid request: contract is nil"},
		{"getAmountsOut nil amountIn", (&GetAmountsOutRequest{Contract: router, Path: []*types.Address{a, b}}).Validate(), "invalid request: amountIn is nil"},
		{"getAmountsOut empty path", (&GetAmountsOutRequest{Contract: router, AmountIn: big.NewInt(1), Path: []*types.Address{}}).Validate(), "invalid request: path is empty"},
		{"getAmountsOut nil path[1]", (&GetAmountsOutRequest{Contract: router, AmountIn: big.NewInt(1), Path: []*types.Address{a, nil, b}}).Validate(), "invalid request: path[1] is nil"},
		{"getAmountsOut valid", (&GetAmountsOutRequest{Contract: router, AmountIn: big.NewInt(1), Path: []*types.Address{a, b}}).Validate(), ""},
		{"getAmountsIn nil", (*GetAmountsInRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getAmountsIn empty", (&GetAmountsInRequest{}).Validate(), "invalid request: contract is nil"},
		{"getAmountsIn nil contract", (&GetAmountsInRequest{AmountOut: big.NewInt(1), Path: []*types.Address{a, b}}).Validate(), "invalid request: contract is nil"},
		{"getAmountsIn nil amountOut", (&GetAmountsInRequest{Contract: router, Path: []*types.Address{a, b}}).Validate(), "invalid request: amountOut is nil"},
		{"getAmountsIn empty path", (&GetAmountsInRequest{Contract: router, AmountOut: big.NewInt(1), Path: []*types.Address{}}).Validate(), "invalid request: path is empty"},
		{"getAmountsIn nil path[1]", (&GetAmountsInRequest{Contract: router, AmountOut: big.NewInt(1), Path: []*types.Address{a, nil, b}}).Validate(), "invalid request: path[1] is nil"},
		{"getAmountsIn valid", (&GetAmountsInRequest{Contract: router, AmountOut: big.NewInt(1), Path: []*types.Address{a, b}}).Validate(), ""},
		{"addLiquidity nil", (*AddLiquidityRequest)(nil).Validate(), "invalid request: request is nil"},
		{"addLiquidity empty", (&AddLiquidityRequest{}).Validate(), "invalid request: contract is nil"},
		{"addLiquidity nil contract", (&AddLiquidityRequest{TokenA: a, TokenB: a, AmountADesired: big.NewInt(1), AmountBDesired: big.NewInt(1), AmountAMin: big.NewInt(1), AmountBMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"addLiquidity nil tokenA", (&AddLiquidityRequest{Contract: router, TokenB: a, AmountADesired: big.NewInt(1), AmountBDesired: big.NewInt(1), AmountAMin: big.NewInt(1), AmountBMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: tokenA is nil"},
		{"addLiquidity nil tokenB", (&AddLiquidityRequest{Contract: router, TokenA: a, AmountADesired: big.NewInt(1), AmountBDesired: big.NewInt(1), AmountAMin: big.NewInt(1), AmountBMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: tokenB is nil"},
		{"addLiquidity nil amountADesired", (&AddLiquidityRequest{Contract: router, TokenA: a, TokenB: a, AmountBDesired: big.NewInt(1), AmountAMin: big.NewInt(1), AmountBMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountADesired is nil"},
		{"addLiquidity nil amountBDesired", (&AddLiquidityRequest{Contract: router, TokenA: a, TokenB: a, AmountADesired: big.NewInt(1), AmountAMin: big.NewInt(1), AmountBMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountBDesired is nil"},
		{"addLiquidity nil amountAMin", (&AddLiquidityRequest{Contract: router, TokenA: a, TokenB: a, AmountADesired: big.NewInt(1), AmountBDesired: big.NewInt(1), AmountBMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountAMin is nil"},
		{"addLiquidity nil amountBMin", (&AddLiquidityRequest{Contract: router, TokenA: a, TokenB: a, AmountADesired: big.NewInt(1), AmountBDesired: big.NewInt(1), AmountAMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountBMin is nil"},
		{"addLiquidity nil to", (&AddLiquidityRequest{Contract: router, TokenA: a, TokenB: a, AmountADesired: big.NewInt(1), AmountBDesired: big.NewInt(1), AmountAMin: big.NewInt(1), AmountBMin: big.NewInt(1), Deadline: big.NewInt(1)}).Validate(), "invalid request: to is nil"},
		{"addLiquidity nil deadline", (&AddLiquidityRequest{Contract: router, TokenA: a, TokenB: a, AmountADesired: big.NewInt(1), AmountBDesired: big.NewInt(1), AmountAMin: big.NewInt(1), AmountBMin: big.NewInt(1), To: a}).Validate(), "invalid request: deadline is nil"},
		{"addLiquidity valid", (&AddLiquidityRequest{Contract: router, TokenA: a, TokenB: a, AmountADesired: big.NewInt(1), AmountBDesired: big.NewInt(1), AmountAMin: big.NewInt(1), AmountBMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), ""},
		{"addLiquidityETH nil", (*AddLiquidityETHRequest)(nil).Validate(), "invalid request: request is nil"},
		{"addLiquidityETH empty", (&AddLiquidityETHRequest{}).Validate(), "invalid request: contract is nil"},
		{"addLiquidityETH nil contract", (&AddLiquidityETHRequest{Token: a, AmountTokenDesired: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), AmountETHDesired: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"addLiquidityETH nil token", (&AddLiquidityETHRequest{Contract: router, AmountTokenDesired: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), AmountETHDesired: big.NewInt(1)}).Validate(), "invalid request: token is nil"},
		{"addLiquidityETH nil amountTokenDesired", (&AddLiquidityETHRequest{Contract: router, Token: a, AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), AmountETHDesired: big.NewInt(1)}).Validate(), "invalid request: amountTokenDesired is nil"},
		{"addLiquidityETH nil amountTokenMin", (&AddLiquidityETHRequest{Contract: router, Token: a, AmountTokenDesired: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), AmountETHDesired: big.NewInt(1)}).Validate(), "invalid request: amountTokenMin is nil"},
		{"addLiquidityETH nil amountETHMin", (&AddLiquidityETHRequest{Contract: router, Token: a, AmountTokenDesired: big.NewInt(1), AmountTokenMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), AmountETHDesired: big.NewInt(1)}).Validate(), "invalid request: amountETHMin is nil"},
		{"addLiquidityETH nil to", (&AddLiquidityETHRequest{Contract: router, Token: a, AmountTokenDesired: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), Deadline: big.NewInt(1), AmountETHDesired: big.NewInt(1)}).Validate(), "invalid request: to is nil"},
		{"addLiquidityETH nil deadline", (&AddLiquidityETHRequest{Contract: router, Token: a, AmountTokenDesired: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, AmountETHDesired: big.NewInt(1)}).Validate(), "invalid request: deadline is nil"},
		{"addLiquidityETH nil amountETHDesired", (&AddLiquidityETHRequest{Contract: router, Token: a, AmountTokenDesired: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountETHDesired is nil"},
		{"addLiquidityETH valid", (&AddLiquidityETHRequest{Contract: router, Token: a, AmountTokenDesired: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), AmountETHDesired: big.NewInt(1)}).Validate(), ""},
		{"removeLiquidity nil", (*RemoveLiquidityRequest)(nil).Validate(), "invalid request: request is nil"},
		{"removeLiquidity empty", (&RemoveLiquidityRequest{}).Validate(), "invalid request: contract is nil"},
		{"removeLiquidity nil contract", (&RemoveLiquidityRequest{TokenA: a, TokenB: a, Liquidity: big.NewInt(1), AmountAMin: big.NewInt(1), AmountBMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"removeLiquidity nil tokenA", (&RemoveLiquidityRequest{Contract: router, TokenB: a, Liquidity: big.NewInt(1), AmountAMin: big.NewInt(1), AmountBMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: tokenA is nil"},
		{"removeLiquidity nil tokenB", (&RemoveLiquidityRequest{Contract: router, TokenA: a, Liquidity: big.NewInt(1), AmountAMin: big.NewInt(1), AmountBMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: tokenB is nil"},
		{"removeLiquidity nil liquidity", (&RemoveLiquidityRequest{Contract: router, TokenA: a, TokenB: a, AmountAMin: big.NewInt(1), AmountBMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: liquidity is nil"},
		{"removeLiquidity nil amountAMin", (&RemoveLiquidityRequest{Contract: router, TokenA: a, TokenB: a, Liquidity: big.NewInt(1), AmountBMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountAMin is nil"},
		{"removeLiquidity nil amountBMin", (&RemoveLiquidityRequest{Contract: router, TokenA: a, TokenB: a, Liquidity: big.NewInt(1), AmountAMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountBMin is nil"},
		{"removeLiquidity nil to", (&RemoveLiquidityRequest{Contract: router, TokenA: a, TokenB: a, Liquidity: big.NewInt(1), AmountAMin: big.NewInt(1), AmountBMin: big.NewInt(1), Deadline: big.NewInt(1)}).Validate(), "invalid request: to is nil"},
		{"removeLiquidity nil deadline", (&RemoveLiquidityRequest{Contract: router, TokenA: a, TokenB: a, Liquidity: big.NewInt(1), AmountAMin: big.NewInt(1), AmountBMin: big.NewInt(1), To: a}).Validate(), "invalid request: deadline is nil"},
		{"removeLiquidity valid", (&RemoveLiquidityRequest{Contract: router, TokenA: a, TokenB: a, Liquidity: big.NewInt(1), AmountAMin: big.NewInt(1), AmountBMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), ""},
		{"removeLiquidityETH nil", (*RemoveLiquidityETHRequest)(nil).Validate(), "invalid request: request is nil"},
		{"removeLiquidityETH empty", (&RemoveLiquidityETHRequest{}).Validate(), "invalid request: contract is nil"},
		{"removeLiquidityETH nil contract", (&RemoveLiquidityETHRequest{Token: a, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"removeLiquidityETH nil token", (&RemoveLiquidityETHRequest{Contract: router, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: token is nil"},
		{"removeLiquidityETH nil liquidity", (&RemoveLiquidityETHRequest{Contract: router, Token: a, AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: liquidity is nil"},
		{"removeLiquidityETH nil amountTokenMin", (&RemoveLiquidityETHRequest{Contract: router, Token: a, Liquidity: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountTokenMin is nil"},
		{"removeLiquidityETH nil amountETHMin", (&RemoveLiquidityETHRequest{Contract: router, Token: a, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountETHMin is nil"},
		{"removeLiquidityETH nil to", (&RemoveLiquidityETHRequest{Contract: router, Token: a, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), Deadline: big.NewInt(1)}).Validate(), "invalid request: to is nil"},
		{"removeLiquidityETH nil deadline", (&RemoveLiquidityETHRequest{Contract: router, Token: a, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a}).Validate(), "invalid request: deadline is nil"},
		{"removeLiquidityETH valid", (&RemoveLiquidityETHRequest{Contract: router, Token: a, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), ""},
		{"removeLiquidityWithPermit nil", (*RemoveLiquidityWithPermitRequest)(nil).Validate(), "invalid request: request is nil"},
		{"removeLiquidityWithPermit empty", (&RemoveLiquidityWithPermitRequest{}).Validate(), "invalid request: contract is nil"},
		{"removeLiquidityWithPermit nil contract", (&RemoveLiquidityWithPermitRequest{TokenA: a, TokenB: a, Liquidity: big.NewInt(1), AmountAMin: big.NewInt(1), AmountBMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), ApproveMax: true, Sig: sig}).Validate(), "invalid request: contract is nil"},
		{"removeLiquidityWithPermit nil tokenA", (&RemoveLiquidityWithPermitRequest{Contract: router, TokenB: a, Liquidity: big.NewInt(1), AmountAMin: big.NewInt(1), AmountBMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), ApproveMax: true, Sig: sig}).Validate(), "invalid request: tokenA is nil"},
		{"removeLiquidityWithPermit nil tokenB", (&RemoveLiquidityWithPermitRequest{Contract: router, TokenA: a, Liquidity: big.NewInt(1), AmountAMin: big.NewInt(1), AmountBMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), ApproveMax: true, Sig: sig}).Validate(), "invalid request: tokenB is nil"},
		{"removeLiquidityWithPermit nil liquidity", (&RemoveLiquidityWithPermitRequest{Contract: router, TokenA: a, TokenB: a, AmountAMin: big.NewInt(1), AmountBMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), ApproveMax: true, Sig: sig}).Validate(), "invalid request: liquidity is nil"},
		{"removeLiquidityWithPermit nil amountAMin", (&RemoveLiquidityWithPermitRequest{Contract: router, TokenA: a, TokenB: a, Liquidity: big.NewInt(1), AmountBMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), ApproveMax: true, Sig: sig}).Validate(), "invalid request: amountAMin is nil"},
		{"removeLiquidityWithPermit nil amountBMin", (&RemoveLiquidityWithPermitRequest{Contract: router, TokenA: a, TokenB: a, Liquidity: big.NewInt(1), AmountAMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), ApproveMax: true, Sig: sig}).Validate(), "invalid request: amountBMin is nil"},
		{"removeLiquidityWithPermit nil to", (&RemoveLiquidityWithPermitRequest{Contract: router, TokenA: a, TokenB: a, Liquidity: big.NewInt(1), AmountAMin: big.NewInt(1), AmountBMin: big.NewInt(1), Deadline: big.NewInt(1), ApproveMax: true, Sig: sig}).Validate(), "invalid request: to is nil"},
		{"removeLiquidityWithPermit nil deadline", (&RemoveLiquidityWithPermitRequest{Contract: router, TokenA: a, TokenB: a, Liquidity: big.NewInt(1), AmountAMin: big.NewInt(1), AmountBMin: big.NewInt(1), To: a, ApproveMax: true, Sig: sig}).Validate(), "invalid request: deadline is nil"},
		{"removeLiquidityWithPermit nil sig", (&RemoveLiquidityWithPermitRequest{Contract: router, TokenA: a, TokenB: a, Liquidity: big.NewInt(1), AmountAMin: big.NewInt(1), AmountBMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), ApproveMax: true}).Validate(), "invalid request: sig is nil"},
		{"removeLiquidityWithPermit valid", (&RemoveLiquidityWithPermitRequest{Contract: router, TokenA: a, TokenB: a, Liquidity: big.NewInt(1), AmountAMin: big.NewInt(1), AmountBMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), ApproveMax: true, Sig: sig}).Validate(), ""},
		{"removeLiquidityETHWithPermit nil", (*RemoveLiquidityETHWithPermitRequest)(nil).Validate(), "invalid request: request is nil"},
		{"removeLiquidityETHWithPermit empty", (&RemoveLiquidityETHWithPermitRequest{}).Validate(), "invalid request: contract is nil"},
		{"removeLiquidityETHWithPermit nil contract", (&RemoveLiquidityETHWithPermitRequest{Token: a, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), ApproveMax: true, Sig: sig}).Validate(), "invalid request: contract is nil"},
		{"removeLiquidityETHWithPermit nil token", (&RemoveLiquidityETHWithPermitRequest{Contract: router, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), ApproveMax: true, Sig: sig}).Validate(), "invalid request: token is nil"},
		{"removeLiquidityETHWithPermit nil liquidity", (&RemoveLiquidityETHWithPermitRequest{Contract: router, Token: a, AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), ApproveMax: true, Sig: sig}).Validate(), "invalid request: liquidity is nil"},
		{"removeLiquidityETHWithPermit nil amountTokenMin", (&RemoveLiquidityETHWithPermitRequest{Contract: router, Token: a, Liquidity: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), ApproveMax: true, Sig: sig}).Validate(), "invalid request: amountTokenMin is nil"},
		{"removeLiquidityETHWithPermit nil amountETHMin", (&RemoveLiquidityETHWithPermitRequest{Contract: router, Token: a, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), ApproveMax: true, Sig: sig}).Validate(), "invalid request: amountETHMin is nil"},
		{"removeLiquidityETHWithPermit nil to", (&RemoveLiquidityETHWithPermitRequest{Contract: router, Token: a, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), Deadline: big.NewInt(1), ApproveMax: true, Sig: sig}).Validate(), "invalid request: to is nil"},
		{"removeLiquidityETHWithPermit nil deadline", (&RemoveLiquidityETHWithPermitRequest{Contract: router, Token: a, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, ApproveMax: true, Sig: sig}).Validate(), "invalid request: deadline is nil"},
		{"removeLiquidityETHWithPermit nil sig", (&RemoveLiquidityETHWithPermitRequest{Contract: router, Token: a, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), ApproveMax: true}).Validate(), "invalid request: sig is nil"},
		{"removeLiquidityETHWithPermit valid", (&RemoveLiquidityETHWithPermitRequest{Contract: router, Token: a, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), ApproveMax: true, Sig: sig}).Validate(), ""},
		{"swapExactTokensForTokens nil", (*SwapExactTokensForTokensRequest)(nil).Validate(), "invalid request: request is nil"},
		{"swapExactTokensForTokens empty", (&SwapExactTokensForTokensRequest{}).Validate(), "invalid request: contract is nil"},
		{"swapExactTokensForTokens nil contract", (&SwapExactTokensForTokensRequest{AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"swapExactTokensForTokens nil amountIn", (&SwapExactTokensForTokensRequest{Contract: router, AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountIn is nil"},
		{"swapExactTokensForTokens nil amountOutMin", (&SwapExactTokensForTokensRequest{Contract: router, AmountIn: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountOutMin is nil"},
		{"swapExactTokensForTokens empty path", (&SwapExactTokensForTokensRequest{Contract: router, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(1), Path: []*types.Address{}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: path is empty"},
		{"swapExactTokensForTokens nil path[1]", (&SwapExactTokensForTokensRequest{Contract: router, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(1), Path: []*types.Address{a, nil, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: path[1] is nil"},
		{"swapExactTokensForTokens nil to", (&SwapExactTokensForTokensRequest{Contract: router, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, Deadline: big.NewInt(1)}).Validate(), "invalid request: to is nil"},
		{"swapExactTokensForTokens nil deadline", (&SwapExactTokensForTokensRequest{Contract: router, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a}).Validate(), "invalid request: deadline is nil"},
		{"swapExactTokensForTokens valid", (&SwapExactTokensForTokensRequest{Contract: router, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), ""},
		{"swapTokensForExactTokens nil", (*SwapTokensForExactTokensRequest)(nil).Validate(), "invalid request: request is nil"},
		{"swapTokensForExactTokens empty", (&SwapTokensForExactTokensRequest{}).Validate(), "invalid request: contract is nil"},
		{"swapTokensForExactTokens nil contract", (&SwapTokensForExactTokensRequest{AmountOut: big.NewInt(1), AmountInMax: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"swapTokensForExactTokens nil amountOut", (&SwapTokensForExactTokensRequest{Contract: router, AmountInMax: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountOut is nil"},
		{"swapTokensForExactTokens nil amountInMax", (&SwapTokensForExactTokensRequest{Contract: router, AmountOut: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountInMax is nil"},
		{"swapTokensForExactTokens empty path", (&SwapTokensForExactTokensRequest{Contract: router, AmountOut: big.NewInt(1), AmountInMax: big.NewInt(1), Path: []*types.Address{}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: path is empty"},
		{"swapTokensForExactTokens nil path[1]", (&SwapTokensForExactTokensRequest{Contract: router, AmountOut: big.NewInt(1), AmountInMax: big.NewInt(1), Path: []*types.Address{a, nil, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: path[1] is nil"},
		{"swapTokensForExactTokens nil to", (&SwapTokensForExactTokensRequest{Contract: router, AmountOut: big.NewInt(1), AmountInMax: big.NewInt(1), Path: []*types.Address{a, b}, Deadline: big.NewInt(1)}).Validate(), "invalid request: to is nil"},
		{"swapTokensForExactTokens nil deadline", (&SwapTokensForExactTokensRequest{Contract: router, AmountOut: big.NewInt(1), AmountInMax: big.NewInt(1), Path: []*types.Address{a, b}, To: a}).Validate(), "invalid request: deadline is nil"},
		{"swapTokensForExactTokens valid", (&SwapTokensForExactTokensRequest{Contract: router, AmountOut: big.NewInt(1), AmountInMax: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), ""},
		{"swapExactETHForTokens nil", (*SwapExactETHForTokensRequest)(nil).Validate(), "invalid request: request is nil"},
		{"swapExactETHForTokens empty", (&SwapExactETHForTokensRequest{}).Validate(), "invalid request: contract is nil"},
		{"swapExactETHForTokens nil contract", (&SwapExactETHForTokensRequest{AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1), AmountIn: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"swapExactETHForTokens nil amountOutMin", (&SwapExactETHForTokensRequest{Contract: router, Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1), AmountIn: big.NewInt(1)}).Validate(), "invalid request: amountOutMin is nil"},
		{"swapExactETHForTokens empty path", (&SwapExactETHForTokensRequest{Contract: router, AmountOutMin: big.NewInt(1), Path: []*types.Address{}, To: a, Deadline: big.NewInt(1), AmountIn: big.NewInt(1)}).Validate(), "invalid request: path is empty"},
		{"swapExactETHForTokens nil path[1]", (&SwapExactETHForTokensRequest{Contract: router, AmountOutMin: big.NewInt(1), Path: []*types.Address{a, nil, b}, To: a, Deadline: big.NewInt(1), AmountIn: big.NewInt(1)}).Validate(), "invalid request: path[1] is nil"},
		{"swapExactETHForTokens nil to", (&SwapExactETHForTokensRequest{Contract: router, AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, Deadline: big.NewInt(1), AmountIn: big.NewInt(1)}).Validate(), "invalid request: to is nil"},
		{"swapExactETHForTokens nil deadline", (&SwapExactETHForTokensRequest{Contract: router, AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a, AmountIn: big.NewInt(1)}).Validate(), "invalid request: deadline is nil"},
		{"swapExactETHForTokens nil amountIn", (&SwapExactETHForTokensRequest{Contract: router, AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountIn is nil"},
		{"swapExactETHForTokens valid", (&SwapExactETHForTokensRequest{Contract: router, AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1), AmountIn: big.NewInt(1)}).Validate(), ""},
		{"swapTokensForExactETH nil", (*SwapTokensForExactETHRequest)(nil).Validate(), "invalid request: request is nil"},
		{"swapTokensForExactETH empty", (&SwapTokensForExactETHRequest{}).Validate(), "invalid request: contract is nil"},
		{"swapTokensForExactETH nil contract", (&SwapTokensForExactETHRequest{AmountOut: big.NewInt(1), AmountInMax: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"swapTokensForExactETH nil amountOut", (&SwapTokensForExactETHRequest{Contract: router, AmountInMax: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountOut is nil"},
		{"swapTokensForExactETH nil amountInMax", (&SwapTokensForExactETHRequest{Contract: router, AmountOut: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountInMax is nil"},
		{"swapTokensForExactETH empty path", (&SwapTokensForExactETHRequest{Contract: router, AmountOut: big.NewInt(1), AmountInMax: big.NewInt(1), Path: []*types.Address{}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: path is empty"},
		{"swapTokensForExactETH nil path[1]", (&SwapTokensForExactETHRequest{Contract: router, AmountOut: big.NewInt(1), AmountInMax: big.NewInt(1), Path: []*types.Address{a, nil, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: path[1] is nil"},
		{"swapTokensForExactETH nil to", (&SwapTokensForExactETHRequest{Contract: router, AmountOut: big.NewInt(1), AmountInMax: big.NewInt(1), Path: []*types.Address{a, b}, Deadline: big.NewInt(1)}).Validate(), "invalid request: to is nil"},
		{"swapTokensForExactETH nil deadline", (&SwapTokensForExactETHRequest{Contract: router, AmountOut: big.NewInt(1), AmountInMax: big.NewInt(1), Path: []*types.Address{a, b}, To: a}).Validate(), "invalid request: deadline is nil"},
		{"swapTokensForExactETH valid", (&SwapTokensForExactETHRequest{Contract: router, AmountOut: big.NewInt(1), AmountInMax: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), ""},
		{"swapExactTokensForETH nil", (*SwapExactTokensForETHRequest)(nil).Validate(), "invalid request: request is nil"},
		{"swapExactTokensForETH empty", (&SwapExactTokensForETHRequest{}).Validate(), "invalid request: contract is nil"},
		{"swapExactTokensForETH nil contract", (&SwapExactTokensForETHRequest{AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"swapExactTokensForETH nil amountIn", (&SwapExactTokensForETHRequest{Contract: router, AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountIn is nil"},
		{"swapExactTokensForETH nil amountOutMin", (&SwapExactTokensForETHRequest{Contract: router, AmountIn: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountOutMin is nil"},
		{"swapExactTokensForETH empty path", (&SwapExactTokensForETHRequest{Contract: router, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(1), Path: []*types.Address{}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: path is empty"},
		{"swapExactTokensForETH nil path[1]", (&SwapExactTokensForETHRequest{Contract: router, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(1), Path: []*types.Address{a, nil, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: path[1] is nil"},
		{"swapExactTokensForETH nil to", (&SwapExactTokensForETHRequest{Contract: router, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, Deadline: big.NewInt(1)}).Validate(), "invalid request: to is nil"},
		{"swapExactTokensForETH nil deadline", (&SwapExactTokensForETHRequest{Contract: router, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a}).Validate(), "invalid request: deadline is nil"},
		{"swapExactTokensForETH valid", (&SwapExactTokensForETHRequest{Contract: router, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), ""},
		{"swapETHForExactTokens nil", (*SwapETHForExactTokensRequest)(nil).Validate(), "invalid request: request is nil"},
		{"swapETHForExactTokens empty", (&SwapETHForExactTokensRequest{}).Validate(), "invalid request: contract is nil"},
		{"swapETHForExactTokens nil contract", (&SwapETHForExactTokensRequest{AmountOut: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1), AmountInMax: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"swapETHForExactTokens nil amountOut", (&SwapETHForExactTokensRequest{Contract: router, Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1), AmountInMax: big.NewInt(1)}).Validate(), "invalid request: amountOut is nil"},
		{"swapETHForExactTokens empty path", (&SwapETHForExactTokensRequest{Contract: router, AmountOut: big.NewInt(1), Path: []*types.Address{}, To: a, Deadline: big.NewInt(1), AmountInMax: big.NewInt(1)}).Validate(), "invalid request: path is empty"},
		{"swapETHForExactTokens nil path[1]", (&SwapETHForExactTokensRequest{Contract: router, AmountOut: big.NewInt(1), Path: []*types.Address{a, nil, b}, To: a, Deadline: big.NewInt(1), AmountInMax: big.NewInt(1)}).Validate(), "invalid request: path[1] is nil"},
		{"swapETHForExactTokens nil to", (&SwapETHForExactTokensRequest{Contract: router, AmountOut: big.NewInt(1), Path: []*types.Address{a, b}, Deadline: big.NewInt(1), AmountInMax: big.NewInt(1)}).Validate(), "invalid request: to is nil"},
		{"swapETHForExactTokens nil deadline", (&SwapETHForExactTokensRequest{Contract: router, AmountOut: big.NewInt(1), Path: []*types.Address{a, b}, To: a, AmountInMax: big.NewInt(1)}).Validate(), "invalid request: deadline is nil"},
		{"swapETHForExactTokens nil amountInMax", (&SwapETHForExactTokensRequest{Contract: router, AmountOut: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountInMax is nil"},
		{"swapETHForExactTokens valid", (&SwapETHForExactTokensRequest{Contract: router, AmountOut: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1), AmountInMax: big.NewInt(1)}).Validate(), ""},
		{"removeLiquidityETHSupportingFeeOnTransferTokens nil", (*RemoveLiquidityETHSupportingFeeOnTransferTokensRequest)(nil).Validate(), "invalid request: request is nil"},
		{"removeLiquidityETHSupportingFeeOnTransferTokens empty", (&RemoveLiquidityETHSupportingFeeOnTransferTokensRequest{}).Validate(), "invalid request: contract is nil"},
		{"removeLiquidityETHSupportingFeeOnTransferTokens nil contract", (&RemoveLiquidityETHSupportingFeeOnTransferTokensRequest{Token: a, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"removeLiquidityETHSupportingFeeOnTransferTokens nil token", (&RemoveLiquidityETHSupportingFeeOnTransferTokensRequest{Contract: router, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: token is nil"},
		{"removeLiquidityETHSupportingFeeOnTransferTokens nil liquidity", (&RemoveLiquidityETHSupportingFeeOnTransferTokensRequest{Contract: router, Token: a, AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: liquidity is nil"},
		{"removeLiquidityETHSupportingFeeOnTransferTokens nil amountTokenMin", (&RemoveLiquidityETHSupportingFeeOnTransferTokensRequest{Contract: router, Token: a, Liquidity: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountTokenMin is nil"},
		{"removeLiquidityETHSupportingFeeOnTransferTokens nil amountETHMin", (&RemoveLiquidityETHSupportingFeeOnTransferTokensRequest{Contract: router, Token: a, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountETHMin is nil"},
		{"removeLiquidityETHSupportingFeeOnTransferTokens nil to", (&RemoveLiquidityETHSupportingFeeOnTransferTokensRequest{Contract: router, Token: a, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), Deadline: big.NewInt(1)}).Validate(), "invalid request: to is nil"},
		{"removeLiquidityETHSupportingFeeOnTransferTokens nil deadline", (&RemoveLiquidityETHSupportingFeeOnTransferTokensRequest{Contract: router, Token: a, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a}).Validate(), "invalid request: deadline is nil"},
		{"removeLiquidityETHSupportingFeeOnTransferTokens valid", (&RemoveLiquidityETHSupportingFeeOnTransferTokensRequest{Contract: router, Token: a, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1)}).Validate(), ""},
		{"removeLiquidityETHWithPermitSupportingFeeOnTransferTokens nil", (*RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest)(nil).Validate(), "invalid request: request is nil"},
		{"removeLiquidityETHWithPermitSupportingFeeOnTransferTokens empty", (&RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest{}).Validate(), "invalid request: contract is nil"},
		{"removeLiquidityETHWithPermitSupportingFeeOnTransferTokens nil contract", (&RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest{Token: a, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), ApproveMax: true, Sig: sig}).Validate(), "invalid request: contract is nil"},
		{"removeLiquidityETHWithPermitSupportingFeeOnTransferTokens nil token", (&RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest{Contract: router, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), ApproveMax: true, Sig: sig}).Validate(), "invalid request: token is nil"},
		{"removeLiquidityETHWithPermitSupportingFeeOnTransferTokens nil liquidity", (&RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest{Contract: router, Token: a, AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), ApproveMax: true, Sig: sig}).Validate(), "invalid request: liquidity is nil"},
		{"removeLiquidityETHWithPermitSupportingFeeOnTransferTokens nil amountTokenMin", (&RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest{Contract: router, Token: a, Liquidity: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), ApproveMax: true, Sig: sig}).Validate(), "invalid request: amountTokenMin is nil"},
		{"removeLiquidityETHWithPermitSupportingFeeOnTransferTokens nil amountETHMin", (&RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest{Contract: router, Token: a, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), ApproveMax: true, Sig: sig}).Validate(), "invalid request: amountETHMin is nil"},
		{"removeLiquidityETHWithPermitSupportingFeeOnTransferTokens nil to", (&RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest{Contract: router, Token: a, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), Deadline: big.NewInt(1), ApproveMax: true, Sig: sig}).Validate(), "invalid request: to is nil"},
		{"removeLiquidityETHWithPermitSupportingFeeOnTransferTokens nil deadline", (&RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest{Contract: router, Token: a, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, ApproveMax: true, Sig: sig}).Validate(), "invalid request: deadline is nil"},
		{"removeLiquidityETHWithPermitSupportingFeeOnTransferTokens nil sig", (&RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest{Contract: router, Token: a, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), ApproveMax: true}).Validate(), "invalid request: sig is nil"},
		{"removeLiquidityETHWithPermitSupportingFeeOnTransferTokens valid", (&RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest{Contract: router, Token: a, Liquidity: big.NewInt(1), AmountTokenMin: big.NewInt(1), AmountETHMin: big.NewInt(1), To: a, Deadline: big.NewInt(1), ApproveMax: true, Sig: sig}).Validate(), ""},
		{"swapExactTokensForTokensSupportingFeeOnTransferTokens nil", (*SwapExactTokensForTokensSupportingFeeOnTransferTokensRequest)(nil).Validate(), "invalid request: request is nil"},
		{"swapExactTokensForTokensSupportingFeeOnTransferTokens empty", (&SwapExactTokensForTokensSupportingFeeOnTransferTokensRequest{}).Validate(), "invalid request: contract is nil"},
		{"swapExactTokensForTokensSupportingFeeOnTransferTokens nil contract", (&SwapExactTokensForTokensSupportingFeeOnTransferTokensRequest{AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"swapExactTokensForTokensSupportingFeeOnTransferTokens nil amountIn", (&SwapExactTokensForTokensSupportingFeeOnTransferTokensRequest{Contract: router, AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountIn is nil"},
		{"swapExactTokensForTokensSupportingFeeOnTransferTokens nil amountOutMin", (&SwapExactTokensForTokensSupportingFeeOnTransferTokensRequest{Contract: router, AmountIn: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountOutMin is nil"},
		{"swapExactTokensForTokensSupportingFeeOnTransferTokens empty path", (&SwapExactTokensForTokensSupportingFeeOnTransferTokensRequest{Contract: router, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(1), Path: []*types.Address{}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: path is empty"},
		{"swapExactTokensForTokensSupportingFeeOnTransferTokens nil path[1]", (&SwapExactTokensForTokensSupportingFeeOnTransferTokensRequest{Contract: router, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(1), Path: []*types.Address{a, nil, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: path[1] is nil"},
		{"swapExactTokensForTokensSupportingFeeOnTransferTokens nil to", (&SwapExactTokensForTokensSupportingFeeOnTransferTokensRequest{Contract: router, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, Deadline: big.NewInt(1)}).Validate(), "invalid request: to is nil"},
		{"swapExactTokensForTokensSupportingFeeOnTransferTokens nil deadline", (&SwapExactTokensForTokensSupportingFeeOnTransferTokensRequest{Contract: router, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a}).Validate(), "invalid request: deadline is nil"},
		{"swapExactTokensForTokensSupportingFeeOnTransferTokens valid", (&SwapExactTokensForTokensSupportingFeeOnTransferTokensRequest{Contract: router, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), ""},
		{"swapExactETHForTokensSupportingFeeOnTransferTokens nil", (*SwapExactETHForTokensSupportingFeeOnTransferTokensRequest)(nil).Validate(), "invalid request: request is nil"},
		{"swapExactETHForTokensSupportingFeeOnTransferTokens empty", (&SwapExactETHForTokensSupportingFeeOnTransferTokensRequest{}).Validate(), "invalid request: contract is nil"},
		{"swapExactETHForTokensSupportingFeeOnTransferTokens nil contract", (&SwapExactETHForTokensSupportingFeeOnTransferTokensRequest{AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1), AmountIn: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"swapExactETHForTokensSupportingFeeOnTransferTokens nil amountOutMin", (&SwapExactETHForTokensSupportingFeeOnTransferTokensRequest{Contract: router, Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1), AmountIn: big.NewInt(1)}).Validate(), "invalid request: amountOutMin is nil"},
		{"swapExactETHForTokensSupportingFeeOnTransferTokens empty path", (&SwapExactETHForTokensSupportingFeeOnTransferTokensRequest{Contract: router, AmountOutMin: big.NewInt(1), Path: []*types.Address{}, To: a, Deadline: big.NewInt(1), AmountIn: big.NewInt(1)}).Validate(), "invalid request: path is empty"},
		{"swapExactETHForTokensSupportingFeeOnTransferTokens nil path[1]", (&SwapExactETHForTokensSupportingFeeOnTransferTokensRequest{Contract: router, AmountOutMin: big.NewInt(1), Path: []*types.Address{a, nil, b}, To: a, Deadline: big.NewInt(1), AmountIn: big.NewInt(1)}).Validate(), "invalid request: path[1] is nil"},
		{"swapExactETHForTokensSupportingFeeOnTransferTokens nil to", (&SwapExactETHForTokensSupportingFeeOnTransferTokensRequest{Contract: router, AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, Deadline: big.NewInt(1), AmountIn: big.NewInt(1)}).Validate(), "invalid request: to is nil"},
		{"swapExactETHForTokensSupportingFeeOnTransferTokens nil deadline", (&SwapExactETHForTokensSupportingFeeOnTransferTokensRequest{Contract: router, AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a, AmountIn: big.NewInt(1)}).Validate(), "invalid request: deadline is nil"},
		{"swapExactETHForTokensSupportingFeeOnTransferTokens nil amountIn", (&SwapExactETHForTokensSupportingFeeOnTransferTokensRequest{Contract: router, AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountIn is nil"},
		{"swapExactETHForTokensSupportingFeeOnTransferTokens valid", (&SwapExactETHForTokensSupportingFeeOnTransferTokensRequest{Contract: router, AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1), AmountIn: big.NewInt(1)}).Validate(), ""},
		{"swapExactTokensForETHSupportingFeeOnTransferTokens nil", (*SwapExactTokensForETHSupportingFeeOnTransferTokensRequest)(nil).Validate(), "invalid request: request is nil"},
		{"swapExactTokensForETHSupportingFeeOnTransferTokens empty", (&SwapExactTokensForETHSupportingFeeOnTransferTokensRequest{}).Validate(), "invalid request: contract is nil"},
		{"swapExactTokensForETHSupportingFeeOnTransferTokens nil contract", (&SwapExactTokensForETHSupportingFeeOnTransferTokensRequest{AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"swapExactTokensForETHSupportingFeeOnTransferTokens nil amountIn", (&SwapExactTokensForETHSupportingFeeOnTransferTokensRequest{Contract: router, AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountIn is nil"},
		{"swapExactTokensForETHSupportingFeeOnTransferTokens nil amountOutMin", (&SwapExactTokensForETHSupportingFeeOnTransferTokensRequest{Contract: router, AmountIn: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: amountOutMin is nil"},
		{"swapExactTokensForETHSupportingFeeOnTransferTokens empty path", (&SwapExactTokensForETHSupportingFeeOnTransferTokensRequest{Contract: router, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(1), Path: []*types.Address{}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: path is empty"},
		{"swapExactTokensForETHSupportingFeeOnTransferTokens nil path[1]", (&SwapExactTokensForETHSupportingFeeOnTransferTokensRequest{Contract: router, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(1), Path: []*types.Address{a, nil, b}, To: a, Deadline: big.NewInt(1)}).Validate(), "invalid request: path[1] is nil"},
		{"swapExactTokensForETHSupportingFeeOnTransferTokens nil to", (&SwapExactTokensForETHSupportingFeeOnTransferTokensRequest{Contract: router, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, Deadline: big.NewInt(1)}).Validate(), "invalid request: to is nil"},
		{"swapExactTokensForETHSupportingFeeOnTransferTokens nil deadline", (&SwapExactTokensForETHSupportingFeeOnTransferTokensRequest{Contract: router, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a}).Validate(), "invalid request: deadline is nil"},
		{"swapExactTokensForETHSupportingFeeOnTransferTokens valid", (&SwapExactTokensForETHSupportingFeeOnTransferTokensRequest{Contract: router, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1)}).Validate(), ""},
		{"getAmountsOut single hop path", (&GetAmountsOutRequest{Contract: router, AmountIn: big.NewInt(1), Path: []*types.Address{a}}).Validate(), ""},
		{"removeLiquidityWithPermit approveMax false", (&RemoveLiquidityWithPermitRequest{Contract: router, TokenA: a, TokenB: b, Liquidity: big.NewInt(1), AmountAMin: big.NewInt(0), AmountBMin: big.NewInt(0), To: a, Deadline: big.NewInt(1), ApproveMax: false, Sig: sig}).Validate(), ""},
		{"swapExactETHForTokens zero amountIn", (&SwapExactETHForTokensRequest{Contract: router, AmountOutMin: big.NewInt(0), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(1), AmountIn: big.NewInt(0)}).Validate(), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.want == "" {
				require.NoError(t, tc.err)
				return
			}
			require.ErrorIs(t, tc.err, bind.ErrInvalidRequest)
			require.EqualError(t, tc.err, tc.want)
		})
	}
}

func TestNewRouterRequestsSetFields(t *testing.T) {
	router := types.MustNewAddressFromHex("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sig, err := key.Sign(core.Keccak256([]byte("permit")))
	require.NoError(t, err)

	require.Equal(t, &WETHRequest{Contract: router}, NewWETHRequest(router))
	require.Equal(t, &QuoteRequest{Contract: router, AmountA: big.NewInt(1), ReserveA: big.NewInt(2), ReserveB: big.NewInt(3)}, NewQuoteRequest(router, big.NewInt(1), big.NewInt(2), big.NewInt(3)))
	require.Equal(t, &GetAmountOutRequest{Contract: router, AmountIn: big.NewInt(1), ReserveIn: big.NewInt(2), ReserveOut: big.NewInt(3)}, NewGetAmountOutRequest(router, big.NewInt(1), big.NewInt(2), big.NewInt(3)))
	require.Equal(t, &GetAmountInRequest{Contract: router, AmountOut: big.NewInt(1), ReserveIn: big.NewInt(2), ReserveOut: big.NewInt(3)}, NewGetAmountInRequest(router, big.NewInt(1), big.NewInt(2), big.NewInt(3)))
	require.Equal(t, &GetAmountsOutRequest{Contract: router, AmountIn: big.NewInt(1), Path: []*types.Address{a, b}}, NewGetAmountsOutRequest(router, big.NewInt(1), []*types.Address{a, b}))
	require.Equal(t, &GetAmountsInRequest{Contract: router, AmountOut: big.NewInt(1), Path: []*types.Address{a, b}}, NewGetAmountsInRequest(router, big.NewInt(1), []*types.Address{a, b}))
	require.Equal(t, &AddLiquidityRequest{Contract: router, TokenA: a, TokenB: b, AmountADesired: big.NewInt(3), AmountBDesired: big.NewInt(4), AmountAMin: big.NewInt(5), AmountBMin: big.NewInt(6), To: a, Deadline: big.NewInt(8)}, NewAddLiquidityRequest(router, a, b, big.NewInt(3), big.NewInt(4), big.NewInt(5), big.NewInt(6), a, big.NewInt(8)))
	require.Equal(t, &AddLiquidityETHRequest{Contract: router, Token: a, AmountTokenDesired: big.NewInt(2), AmountTokenMin: big.NewInt(3), AmountETHMin: big.NewInt(4), To: a, Deadline: big.NewInt(6), AmountETHDesired: big.NewInt(7)}, NewAddLiquidityETHRequest(router, a, big.NewInt(2), big.NewInt(3), big.NewInt(4), a, big.NewInt(6), big.NewInt(7)))
	require.Equal(t, &RemoveLiquidityRequest{Contract: router, TokenA: a, TokenB: b, Liquidity: big.NewInt(3), AmountAMin: big.NewInt(4), AmountBMin: big.NewInt(5), To: b, Deadline: big.NewInt(7)}, NewRemoveLiquidityRequest(router, a, b, big.NewInt(3), big.NewInt(4), big.NewInt(5), b, big.NewInt(7)))
	require.Equal(t, &RemoveLiquidityETHRequest{Contract: router, Token: a, Liquidity: big.NewInt(2), AmountTokenMin: big.NewInt(3), AmountETHMin: big.NewInt(4), To: a, Deadline: big.NewInt(6)}, NewRemoveLiquidityETHRequest(router, a, big.NewInt(2), big.NewInt(3), big.NewInt(4), a, big.NewInt(6)))
	require.Equal(t, &RemoveLiquidityWithPermitRequest{Contract: router, TokenA: a, TokenB: b, Liquidity: big.NewInt(3), AmountAMin: big.NewInt(4), AmountBMin: big.NewInt(5), To: b, Deadline: big.NewInt(7), ApproveMax: true, Sig: sig}, NewRemoveLiquidityWithPermitRequest(router, a, b, big.NewInt(3), big.NewInt(4), big.NewInt(5), b, big.NewInt(7), true, sig))
	require.Equal(t, &RemoveLiquidityETHWithPermitRequest{Contract: router, Token: a, Liquidity: big.NewInt(2), AmountTokenMin: big.NewInt(3), AmountETHMin: big.NewInt(4), To: a, Deadline: big.NewInt(6), ApproveMax: true, Sig: sig}, NewRemoveLiquidityETHWithPermitRequest(router, a, big.NewInt(2), big.NewInt(3), big.NewInt(4), a, big.NewInt(6), true, sig))
	require.Equal(t, &SwapExactTokensForTokensRequest{Contract: router, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(2), Path: []*types.Address{a, b}, To: b, Deadline: big.NewInt(5)}, NewSwapExactTokensForTokensRequest(router, big.NewInt(1), big.NewInt(2), []*types.Address{a, b}, b, big.NewInt(5)))
	require.Equal(t, &SwapTokensForExactTokensRequest{Contract: router, AmountOut: big.NewInt(1), AmountInMax: big.NewInt(2), Path: []*types.Address{a, b}, To: b, Deadline: big.NewInt(5)}, NewSwapTokensForExactTokensRequest(router, big.NewInt(1), big.NewInt(2), []*types.Address{a, b}, b, big.NewInt(5)))
	require.Equal(t, &SwapExactETHForTokensRequest{Contract: router, AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(4), AmountIn: big.NewInt(5)}, NewSwapExactETHForTokensRequest(router, big.NewInt(1), []*types.Address{a, b}, a, big.NewInt(4), big.NewInt(5)))
	require.Equal(t, &SwapTokensForExactETHRequest{Contract: router, AmountOut: big.NewInt(1), AmountInMax: big.NewInt(2), Path: []*types.Address{a, b}, To: b, Deadline: big.NewInt(5)}, NewSwapTokensForExactETHRequest(router, big.NewInt(1), big.NewInt(2), []*types.Address{a, b}, b, big.NewInt(5)))
	require.Equal(t, &SwapExactTokensForETHRequest{Contract: router, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(2), Path: []*types.Address{a, b}, To: b, Deadline: big.NewInt(5)}, NewSwapExactTokensForETHRequest(router, big.NewInt(1), big.NewInt(2), []*types.Address{a, b}, b, big.NewInt(5)))
	require.Equal(t, &SwapETHForExactTokensRequest{Contract: router, AmountOut: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(4), AmountInMax: big.NewInt(5)}, NewSwapETHForExactTokensRequest(router, big.NewInt(1), []*types.Address{a, b}, a, big.NewInt(4), big.NewInt(5)))
	require.Equal(t, &RemoveLiquidityETHSupportingFeeOnTransferTokensRequest{Contract: router, Token: a, Liquidity: big.NewInt(2), AmountTokenMin: big.NewInt(3), AmountETHMin: big.NewInt(4), To: a, Deadline: big.NewInt(6)}, NewRemoveLiquidityETHSupportingFeeOnTransferTokensRequest(router, a, big.NewInt(2), big.NewInt(3), big.NewInt(4), a, big.NewInt(6)))
	require.Equal(t, &RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest{Contract: router, Token: a, Liquidity: big.NewInt(2), AmountTokenMin: big.NewInt(3), AmountETHMin: big.NewInt(4), To: a, Deadline: big.NewInt(6), ApproveMax: true, Sig: sig}, NewRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest(router, a, big.NewInt(2), big.NewInt(3), big.NewInt(4), a, big.NewInt(6), true, sig))
	require.Equal(t, &SwapExactTokensForTokensSupportingFeeOnTransferTokensRequest{Contract: router, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(2), Path: []*types.Address{a, b}, To: b, Deadline: big.NewInt(5)}, NewSwapExactTokensForTokensSupportingFeeOnTransferTokensRequest(router, big.NewInt(1), big.NewInt(2), []*types.Address{a, b}, b, big.NewInt(5)))
	require.Equal(t, &SwapExactETHForTokensSupportingFeeOnTransferTokensRequest{Contract: router, AmountOutMin: big.NewInt(1), Path: []*types.Address{a, b}, To: a, Deadline: big.NewInt(4), AmountIn: big.NewInt(5)}, NewSwapExactETHForTokensSupportingFeeOnTransferTokensRequest(router, big.NewInt(1), []*types.Address{a, b}, a, big.NewInt(4), big.NewInt(5)))
	require.Equal(t, &SwapExactTokensForETHSupportingFeeOnTransferTokensRequest{Contract: router, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(2), Path: []*types.Address{a, b}, To: b, Deadline: big.NewInt(5)}, NewSwapExactTokensForETHSupportingFeeOnTransferTokensRequest(router, big.NewInt(1), big.NewInt(2), []*types.Address{a, b}, b, big.NewInt(5)))
}
