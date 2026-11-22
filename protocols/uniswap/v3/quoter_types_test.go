package v3

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewQuoteExactInputSingleParams(t *testing.T) {
	tokenIn := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	tokenOut := types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222")
	amountIn := big.NewInt(1)
	fee := big.NewInt(3000)
	sqrtPriceLimitX96 := big.NewInt(2)

	p := NewQuoteExactInputSingleParams(tokenIn, tokenOut, amountIn, fee, sqrtPriceLimitX96)
	require.Same(t, tokenIn, p.TokenIn)
	require.Same(t, tokenOut, p.TokenOut)
	require.Same(t, amountIn, p.AmountIn)
	require.Same(t, fee, p.Fee)
	require.Same(t, sqrtPriceLimitX96, p.SqrtPriceLimitX96)
}

func TestNewQuoteExactOutputSingleParams(t *testing.T) {
	tokenIn := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	tokenOut := types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222")
	amount := big.NewInt(1)
	fee := big.NewInt(3000)
	sqrtPriceLimitX96 := big.NewInt(2)

	p := NewQuoteExactOutputSingleParams(tokenIn, tokenOut, amount, fee, sqrtPriceLimitX96)
	require.Same(t, tokenIn, p.TokenIn)
	require.Same(t, tokenOut, p.TokenOut)
	require.Same(t, amount, p.Amount)
	require.Same(t, fee, p.Fee)
	require.Same(t, sqrtPriceLimitX96, p.SqrtPriceLimitX96)
}

func TestNewQuoteExactInputSingleResult(t *testing.T) {
	amountOut := big.NewInt(1)
	sqrtPriceX96After := big.NewInt(2)
	initializedTicksCrossed := big.NewInt(3)
	gasEstimate := big.NewInt(4)

	r := NewQuoteExactInputSingleResult(amountOut, sqrtPriceX96After, initializedTicksCrossed, gasEstimate)
	require.Same(t, amountOut, r.AmountOut)
	require.Same(t, sqrtPriceX96After, r.SqrtPriceX96After)
	require.Same(t, initializedTicksCrossed, r.InitializedTicksCrossed)
	require.Same(t, gasEstimate, r.GasEstimate)
}

func TestNewQuoteExactInputResult(t *testing.T) {
	amountOut := big.NewInt(1)
	sqrtPriceX96AfterList := []*big.Int{big.NewInt(2)}
	initializedTicksCrossedList := []*big.Int{big.NewInt(3)}
	gasEstimate := big.NewInt(4)

	r := NewQuoteExactInputResult(amountOut, sqrtPriceX96AfterList, initializedTicksCrossedList, gasEstimate)
	require.Same(t, amountOut, r.AmountOut)
	require.Equal(t, sqrtPriceX96AfterList, r.SqrtPriceX96AfterList)
	require.Equal(t, initializedTicksCrossedList, r.InitializedTicksCrossedList)
	require.Same(t, gasEstimate, r.GasEstimate)
}

func TestNewQuoteExactOutputSingleResult(t *testing.T) {
	amountIn := big.NewInt(1)
	sqrtPriceX96After := big.NewInt(2)
	initializedTicksCrossed := big.NewInt(3)
	gasEstimate := big.NewInt(4)

	r := NewQuoteExactOutputSingleResult(amountIn, sqrtPriceX96After, initializedTicksCrossed, gasEstimate)
	require.Same(t, amountIn, r.AmountIn)
	require.Same(t, sqrtPriceX96After, r.SqrtPriceX96After)
	require.Same(t, initializedTicksCrossed, r.InitializedTicksCrossed)
	require.Same(t, gasEstimate, r.GasEstimate)
}

func TestNewQuoteExactOutputResult(t *testing.T) {
	amountIn := big.NewInt(1)
	sqrtPriceX96AfterList := []*big.Int{big.NewInt(2)}
	initializedTicksCrossedList := []*big.Int{big.NewInt(3)}
	gasEstimate := big.NewInt(4)

	r := NewQuoteExactOutputResult(amountIn, sqrtPriceX96AfterList, initializedTicksCrossedList, gasEstimate)
	require.Same(t, amountIn, r.AmountIn)
	require.Equal(t, sqrtPriceX96AfterList, r.SqrtPriceX96AfterList)
	require.Equal(t, initializedTicksCrossedList, r.InitializedTicksCrossedList)
	require.Same(t, gasEstimate, r.GasEstimate)
}

func TestQuoterRequestsValidate(t *testing.T) {
	quoter := types.MustNewAddressFromHex("0x61fFE014bA17989E743c5F6cB21bF9697530B21e")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	path := append(append(append([]byte{}, a.Bytes()...), 0x00, 0x0b, 0xb8), b.Bytes()...)

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"quoteExactInputSingle nil", (*QuoteExactInputSingleRequest)(nil).Validate(), "invalid request: request is nil"},
		{"quoteExactInputSingle empty", (&QuoteExactInputSingleRequest{}).Validate(), "invalid request: contract is nil"},
		{"quoteExactInputSingle nil contract", (&QuoteExactInputSingleRequest{Params: &QuoteExactInputSingleParams{TokenIn: a, TokenOut: b, AmountIn: big.NewInt(1000), Fee: big.NewInt(3000), SqrtPriceLimitX96: big.NewInt(0)}}).Validate(), "invalid request: contract is nil"},
		{"quoteExactInputSingle nil params", (&QuoteExactInputSingleRequest{Contract: quoter}).Validate(), "invalid request: params is nil"},
		{"quoteExactInputSingle nil params.tokenIn", (&QuoteExactInputSingleRequest{Contract: quoter, Params: &QuoteExactInputSingleParams{TokenOut: b, AmountIn: big.NewInt(1000), Fee: big.NewInt(3000), SqrtPriceLimitX96: big.NewInt(0)}}).Validate(), "invalid request: params.tokenIn is nil"},
		{"quoteExactInputSingle nil params.tokenOut", (&QuoteExactInputSingleRequest{Contract: quoter, Params: &QuoteExactInputSingleParams{TokenIn: a, AmountIn: big.NewInt(1000), Fee: big.NewInt(3000), SqrtPriceLimitX96: big.NewInt(0)}}).Validate(), "invalid request: params.tokenOut is nil"},
		{"quoteExactInputSingle nil params.amountIn", (&QuoteExactInputSingleRequest{Contract: quoter, Params: &QuoteExactInputSingleParams{TokenIn: a, TokenOut: b, Fee: big.NewInt(3000), SqrtPriceLimitX96: big.NewInt(0)}}).Validate(), "invalid request: params.amountIn is nil"},
		{"quoteExactInputSingle nil params.fee", (&QuoteExactInputSingleRequest{Contract: quoter, Params: &QuoteExactInputSingleParams{TokenIn: a, TokenOut: b, AmountIn: big.NewInt(1000), SqrtPriceLimitX96: big.NewInt(0)}}).Validate(), "invalid request: params.fee is nil"},
		{"quoteExactInputSingle nil params.sqrtPriceLimitX96", (&QuoteExactInputSingleRequest{Contract: quoter, Params: &QuoteExactInputSingleParams{TokenIn: a, TokenOut: b, AmountIn: big.NewInt(1000), Fee: big.NewInt(3000)}}).Validate(), "invalid request: params.sqrtPriceLimitX96 is nil"},
		{"quoteExactInputSingle valid", (&QuoteExactInputSingleRequest{Contract: quoter, Params: &QuoteExactInputSingleParams{TokenIn: a, TokenOut: b, AmountIn: big.NewInt(1000), Fee: big.NewInt(3000), SqrtPriceLimitX96: big.NewInt(0)}}).Validate(), ""},
		{"quoteExactOutputSingle nil", (*QuoteExactOutputSingleRequest)(nil).Validate(), "invalid request: request is nil"},
		{"quoteExactOutputSingle empty", (&QuoteExactOutputSingleRequest{}).Validate(), "invalid request: contract is nil"},
		{"quoteExactOutputSingle nil contract", (&QuoteExactOutputSingleRequest{Params: &QuoteExactOutputSingleParams{TokenIn: a, TokenOut: b, Amount: big.NewInt(1000), Fee: big.NewInt(3000), SqrtPriceLimitX96: big.NewInt(0)}}).Validate(), "invalid request: contract is nil"},
		{"quoteExactOutputSingle nil params", (&QuoteExactOutputSingleRequest{Contract: quoter}).Validate(), "invalid request: params is nil"},
		{"quoteExactOutputSingle nil params.tokenIn", (&QuoteExactOutputSingleRequest{Contract: quoter, Params: &QuoteExactOutputSingleParams{TokenOut: b, Amount: big.NewInt(1000), Fee: big.NewInt(3000), SqrtPriceLimitX96: big.NewInt(0)}}).Validate(), "invalid request: params.tokenIn is nil"},
		{"quoteExactOutputSingle nil params.tokenOut", (&QuoteExactOutputSingleRequest{Contract: quoter, Params: &QuoteExactOutputSingleParams{TokenIn: a, Amount: big.NewInt(1000), Fee: big.NewInt(3000), SqrtPriceLimitX96: big.NewInt(0)}}).Validate(), "invalid request: params.tokenOut is nil"},
		{"quoteExactOutputSingle nil params.amount", (&QuoteExactOutputSingleRequest{Contract: quoter, Params: &QuoteExactOutputSingleParams{TokenIn: a, TokenOut: b, Fee: big.NewInt(3000), SqrtPriceLimitX96: big.NewInt(0)}}).Validate(), "invalid request: params.amount is nil"},
		{"quoteExactOutputSingle nil params.fee", (&QuoteExactOutputSingleRequest{Contract: quoter, Params: &QuoteExactOutputSingleParams{TokenIn: a, TokenOut: b, Amount: big.NewInt(1000), SqrtPriceLimitX96: big.NewInt(0)}}).Validate(), "invalid request: params.fee is nil"},
		{"quoteExactOutputSingle nil params.sqrtPriceLimitX96", (&QuoteExactOutputSingleRequest{Contract: quoter, Params: &QuoteExactOutputSingleParams{TokenIn: a, TokenOut: b, Amount: big.NewInt(1000), Fee: big.NewInt(3000)}}).Validate(), "invalid request: params.sqrtPriceLimitX96 is nil"},
		{"quoteExactOutputSingle valid", (&QuoteExactOutputSingleRequest{Contract: quoter, Params: &QuoteExactOutputSingleParams{TokenIn: a, TokenOut: b, Amount: big.NewInt(1000), Fee: big.NewInt(3000), SqrtPriceLimitX96: big.NewInt(0)}}).Validate(), ""},
		{"quoteExactInput nil", (*QuoteExactInputRequest)(nil).Validate(), "invalid request: request is nil"},
		{"quoteExactInput empty", (&QuoteExactInputRequest{}).Validate(), "invalid request: contract is nil"},
		{"quoteExactInput nil contract", (&QuoteExactInputRequest{Path: path, AmountIn: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"quoteExactInput empty path", (&QuoteExactInputRequest{Contract: quoter, Path: []byte{}, AmountIn: big.NewInt(1)}).Validate(), "invalid request: path is empty"},
		{"quoteExactInput nil amountIn", (&QuoteExactInputRequest{Contract: quoter, Path: path}).Validate(), "invalid request: amountIn is nil"},
		{"quoteExactInput valid", (&QuoteExactInputRequest{Contract: quoter, Path: path, AmountIn: big.NewInt(1)}).Validate(), ""},
		{"quoteExactOutput nil", (*QuoteExactOutputRequest)(nil).Validate(), "invalid request: request is nil"},
		{"quoteExactOutput empty", (&QuoteExactOutputRequest{}).Validate(), "invalid request: contract is nil"},
		{"quoteExactOutput nil contract", (&QuoteExactOutputRequest{Path: path, AmountOut: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"quoteExactOutput empty path", (&QuoteExactOutputRequest{Contract: quoter, Path: []byte{}, AmountOut: big.NewInt(1)}).Validate(), "invalid request: path is empty"},
		{"quoteExactOutput nil amountOut", (&QuoteExactOutputRequest{Contract: quoter, Path: path}).Validate(), "invalid request: amountOut is nil"},
		{"quoteExactOutput valid", (&QuoteExactOutputRequest{Contract: quoter, Path: path, AmountOut: big.NewInt(1)}).Validate(), ""},
		{"quoteExactInput malformed path accepted", (&QuoteExactInputRequest{Contract: quoter, Path: []byte{0x01}, AmountIn: big.NewInt(1)}).Validate(), ""},
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

func TestNewQuoterRequestsSetFields(t *testing.T) {
	quoter := types.MustNewAddressFromHex("0x61fFE014bA17989E743c5F6cB21bF9697530B21e")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	path := append(append(append([]byte{}, a.Bytes()...), 0x00, 0x0b, 0xb8), b.Bytes()...)
	inParams := NewQuoteExactInputSingleParams(a, b, big.NewInt(1000), big.NewInt(3000), big.NewInt(0))
	outParams := NewQuoteExactOutputSingleParams(a, b, big.NewInt(1000), big.NewInt(3000), big.NewInt(0))

	require.Equal(t, &FactoryRequest{Contract: quoter}, NewFactoryRequest(quoter))
	require.Equal(t, &WETH9Request{Contract: quoter}, NewWETH9Request(quoter))
	require.Equal(t, &QuoteExactInputSingleRequest{Contract: quoter, Params: inParams}, NewQuoteExactInputSingleRequest(quoter, inParams))
	require.Equal(t, &QuoteExactInputRequest{Contract: quoter, Path: path, AmountIn: big.NewInt(5)}, NewQuoteExactInputRequest(quoter, path, big.NewInt(5)))
	require.Equal(t, &QuoteExactOutputSingleRequest{Contract: quoter, Params: outParams}, NewQuoteExactOutputSingleRequest(quoter, outParams))
	require.Equal(t, &QuoteExactOutputRequest{Contract: quoter, Path: path, AmountOut: big.NewInt(6)}, NewQuoteExactOutputRequest(quoter, path, big.NewInt(6)))
}
