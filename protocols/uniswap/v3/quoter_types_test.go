package v3

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

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
