package v2

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
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
