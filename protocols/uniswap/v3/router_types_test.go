package v3

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewExactInputSingleParams(t *testing.T) {
	tokenIn := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	tokenOut := types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222")
	fee := big.NewInt(1)
	recipient := types.MustNewAddressFromHex("0x3333333333333333333333333333333333333333")
	amountIn := big.NewInt(3)
	amountOutMinimum := big.NewInt(4)
	sqrtPriceLimitX96 := big.NewInt(5)

	p := NewExactInputSingleParams(tokenIn, tokenOut, fee, recipient, amountIn, amountOutMinimum, sqrtPriceLimitX96)
	require.Same(t, tokenIn, p.TokenIn)
	require.Same(t, tokenOut, p.TokenOut)
	require.Same(t, fee, p.Fee)
	require.Same(t, recipient, p.Recipient)
	require.Same(t, amountIn, p.AmountIn)
	require.Same(t, amountOutMinimum, p.AmountOutMinimum)
	require.Same(t, sqrtPriceLimitX96, p.SqrtPriceLimitX96)
}

func TestNewExactInputParams(t *testing.T) {
	path := []byte{0x01, 0x02}
	recipient := types.MustNewAddressFromHex("0x3333333333333333333333333333333333333333")
	amountIn := big.NewInt(2)
	amountOutMinimum := big.NewInt(3)

	p := NewExactInputParams(path, recipient, amountIn, amountOutMinimum)
	require.Same(t, recipient, p.Recipient)
	require.Same(t, amountIn, p.AmountIn)
	require.Same(t, amountOutMinimum, p.AmountOutMinimum)
	require.Equal(t, path, p.Path)
}

func TestNewExactOutputSingleParams(t *testing.T) {
	tokenIn := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	tokenOut := types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222")
	fee := big.NewInt(1)
	recipient := types.MustNewAddressFromHex("0x3333333333333333333333333333333333333333")
	amountOut := big.NewInt(3)
	amountInMaximum := big.NewInt(4)
	sqrtPriceLimitX96 := big.NewInt(5)

	p := NewExactOutputSingleParams(tokenIn, tokenOut, fee, recipient, amountOut, amountInMaximum, sqrtPriceLimitX96)
	require.Same(t, tokenIn, p.TokenIn)
	require.Same(t, tokenOut, p.TokenOut)
	require.Same(t, fee, p.Fee)
	require.Same(t, recipient, p.Recipient)
	require.Same(t, amountOut, p.AmountOut)
	require.Same(t, amountInMaximum, p.AmountInMaximum)
	require.Same(t, sqrtPriceLimitX96, p.SqrtPriceLimitX96)
}

func TestNewExactOutputParams(t *testing.T) {
	path := []byte{0x03, 0x04}
	recipient := types.MustNewAddressFromHex("0x3333333333333333333333333333333333333333")
	amountOut := big.NewInt(2)
	amountInMaximum := big.NewInt(3)

	p := NewExactOutputParams(path, recipient, amountOut, amountInMaximum)
	require.Same(t, recipient, p.Recipient)
	require.Same(t, amountOut, p.AmountOut)
	require.Same(t, amountInMaximum, p.AmountInMaximum)
	require.Equal(t, path, p.Path)
}
