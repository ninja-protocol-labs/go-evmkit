package v4

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewQuoteExactSingleParams(t *testing.T) {
	currency0 := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	currency1 := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	hooks := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	poolKey := NewPoolKey(currency0, currency1, big.NewInt(3000), big.NewInt(60), hooks)
	exactAmount := big.NewInt(1000)
	hookData := []byte{0xde, 0xad}

	p := NewQuoteExactSingleParams(poolKey, true, exactAmount, hookData)
	require.Same(t, poolKey, p.PoolKey)
	require.True(t, p.ZeroForOne)
	require.Same(t, exactAmount, p.ExactAmount)
	require.Equal(t, hookData, p.HookData)

	p2 := NewQuoteExactSingleParams(poolKey, false, exactAmount, nil)
	require.False(t, p2.ZeroForOne)
	require.Nil(t, p2.HookData)
}

func TestNewQuoteExactSingleParamsSharesPointers(t *testing.T) {
	poolKey := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(500), big.NewInt(10), testHooks)
	exactAmount := big.NewInt(1)
	hookData := []byte{0x01}

	p := NewQuoteExactSingleParams(poolKey, true, exactAmount, hookData)
	exactAmount.SetInt64(42)
	hookData[0] = 0xff
	require.Zero(t, big.NewInt(42).Cmp(p.ExactAmount))
	require.Equal(t, byte(0xff), p.HookData[0])
}

func TestNewQuoteExactParams(t *testing.T) {
	exactCurrency := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	intermediate0 := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	intermediate1 := types.MustNewAddressFromHex("0x1231deb6f5749ef6ce6943a275a1d3e7486f4eae")
	hooks := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	path := []*PathKey{
		NewPathKey(intermediate0, big.NewInt(3000), big.NewInt(60), hooks, nil),
		NewPathKey(intermediate1, big.NewInt(500), big.NewInt(10), hooks, []byte{0x01}),
	}
	exactAmount := big.NewInt(1000)

	p := NewQuoteExactParams(exactCurrency, path, exactAmount)
	require.Same(t, exactCurrency, p.ExactCurrency)
	require.Len(t, p.Path, 2)
	require.Same(t, path[0], p.Path[0])
	require.Same(t, path[1], p.Path[1])
	require.Same(t, exactAmount, p.ExactAmount)
}

func TestNewQuoteExactParamsEmptyPath(t *testing.T) {
	p := NewQuoteExactParams(testCurrency0, nil, big.NewInt(0))
	require.Nil(t, p.Path)
	require.Zero(t, p.ExactAmount.Sign())

	p2 := NewQuoteExactParams(testCurrency0, []*PathKey{}, big.NewInt(0))
	require.NotNil(t, p2.Path)
	require.Empty(t, p2.Path)
}

func TestNewQuoteExactInputResult(t *testing.T) {
	amountOut := big.NewInt(997)
	gasEstimate := big.NewInt(85000)

	r := NewQuoteExactInputResult(amountOut, gasEstimate)
	require.Same(t, amountOut, r.AmountOut)
	require.Same(t, gasEstimate, r.GasEstimate)
}

func TestNewQuoteExactOutputResult(t *testing.T) {
	amountIn := big.NewInt(1003)
	gasEstimate := big.NewInt(91000)

	r := NewQuoteExactOutputResult(amountIn, gasEstimate)
	require.Same(t, amountIn, r.AmountIn)
	require.Same(t, gasEstimate, r.GasEstimate)
}
