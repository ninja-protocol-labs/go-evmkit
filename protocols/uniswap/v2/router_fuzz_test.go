package v2

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// go test -run '^$' -fuzz '^FuzzDecodeGetAmountsOutRoundTrip$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzDecodeGetAmountsOutRoundTrip(f *testing.F) {
	f.Add([]byte{0x01}, []byte{0x02}, []byte{0x03})
	f.Add(make([]byte, 32), make([]byte, 32), make([]byte, 32))

	f.Fuzz(func(t *testing.T, a, b, c []byte) {
		amounts := make([]*big.Int, 0, 3)
		for _, raw := range [][]byte{a, b, c} {
			if len(raw) > 32 {
				raw = raw[len(raw)-32:]
			}
			amounts = append(amounts, new(big.Int).SetBytes(raw))
		}

		data, err := abi.Pack(abi.Types{abi.Slice(abi.Uint256)}, amounts)
		require.NoError(t, err)

		got, err := DecodeGetAmountsOut(data)
		require.NoError(t, err)
		require.Len(t, got, len(amounts))
		for i, want := range amounts {
			require.Zero(t, want.Cmp(got[i]))
		}
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeSwapExactTokensForTokensRoundTrip$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzEncodeSwapExactTokensForTokensRoundTrip(f *testing.F) {
	f.Add([]byte{}, []byte{}, make([]byte, 20), make([]byte, 20), make([]byte, 20))
	f.Add(make([]byte, 32), make([]byte, 32), []byte{0xde, 0xad, 0xbe, 0xef}, make([]byte, 20), make([]byte, 20))

	f.Fuzz(func(t *testing.T, amountInRaw, amountOutMinRaw, token0Raw, token1Raw, toRaw []byte) {
		if len(amountInRaw) > 32 {
			amountInRaw = amountInRaw[len(amountInRaw)-32:]
		}
		if len(amountOutMinRaw) > 32 {
			amountOutMinRaw = amountOutMinRaw[len(amountOutMinRaw)-32:]
		}
		amountIn := new(big.Int).SetBytes(amountInRaw)
		amountOutMin := new(big.Int).SetBytes(amountOutMinRaw)
		path := []*types.Address{types.NewAddressFromBytes(token0Raw), types.NewAddressFromBytes(token1Raw)}
		to := types.NewAddressFromBytes(toRaw)
		deadline := big.NewInt(9999999999)

		data, err := EncodeSwapExactTokensForTokens(amountIn, amountOutMin, path, to, deadline)
		require.NoError(t, err)
		require.Equal(t, []byte{0x38, 0xed, 0x17, 0x39}, data[:4])

		vals, err := swapExactTokensForTokensFn.DecodeCall(data)
		require.NoError(t, err)
		require.Zero(t, amountIn.Cmp(vals[0].(*big.Int)))
		require.Zero(t, amountOutMin.Cmp(vals[1].(*big.Int)))
		gotPath, ok := vals[2].([]any)
		require.True(t, ok)
		require.Len(t, gotPath, 2)
		require.True(t, gotPath[0].(*types.Address).Equal(path[0]))
		require.True(t, gotPath[1].(*types.Address).Equal(path[1]))
		require.True(t, vals[3].(*types.Address).Equal(to))
		require.Zero(t, deadline.Cmp(vals[4].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeAddLiquidityRoundTrip$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzDecodeAddLiquidityRoundTrip(f *testing.F) {
	f.Add([]byte{0x01}, []byte{0x02}, []byte{0x03})
	f.Add(make([]byte, 32), make([]byte, 32), make([]byte, 32))

	f.Fuzz(func(t *testing.T, aRaw, bRaw, lRaw []byte) {
		if len(aRaw) > 32 {
			aRaw = aRaw[len(aRaw)-32:]
		}
		if len(bRaw) > 32 {
			bRaw = bRaw[len(bRaw)-32:]
		}
		if len(lRaw) > 32 {
			lRaw = lRaw[len(lRaw)-32:]
		}
		amountA := new(big.Int).SetBytes(aRaw)
		amountB := new(big.Int).SetBytes(bRaw)
		liquidity := new(big.Int).SetBytes(lRaw)

		data, err := abi.Pack(abi.Types{abi.Uint256, abi.Uint256, abi.Uint256}, amountA, amountB, liquidity)
		require.NoError(t, err)

		got, err := DecodeAddLiquidity(data)
		require.NoError(t, err)
		require.Zero(t, amountA.Cmp(got.AmountA))
		require.Zero(t, amountB.Cmp(got.AmountB))
		require.Zero(t, liquidity.Cmp(got.Liquidity))
	})
}

// go test -run '^$' -fuzz '^FuzzRouterSimpleDecodersDoNotPanic$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzRouterSimpleDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 33))
	f.Add(make([]byte, 64))
	f.Add(make([]byte, 96))
	f.Add(make([]byte, 128))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeRouterFactory(data)
		_, _ = DecodeWETH(data)
		_, _ = DecodeQuote(data)
		_, _ = DecodeGetAmountOut(data)
		_, _ = DecodeGetAmountIn(data)
		_, _ = DecodeGetAmountsOut(data)
		_, _ = DecodeGetAmountsIn(data)
		_, _ = DecodeAddLiquidity(data)
		_, _ = DecodeAddLiquidityETH(data)
		_, _ = DecodeRemoveLiquidity(data)
		_, _ = DecodeRemoveLiquidityETH(data)
		_, _ = DecodeRemoveLiquidityWithPermit(data)
		_, _ = DecodeRemoveLiquidityETHWithPermit(data)
		_, _ = DecodeSwapExactTokensForTokens(data)
		_, _ = DecodeSwapTokensForExactTokens(data)
		_, _ = DecodeSwapExactETHForTokens(data)
		_, _ = DecodeSwapTokensForExactETH(data)
		_, _ = DecodeSwapExactTokensForETH(data)
		_, _ = DecodeSwapETHForExactTokens(data)
		_, _ = DecodeRemoveLiquidityETHSupportingFeeOnTransferTokens(data)
		_, _ = DecodeRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(data)
	})
}
