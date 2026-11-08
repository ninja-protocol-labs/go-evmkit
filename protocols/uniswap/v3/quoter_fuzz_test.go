package v3

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// go test -run '^$' -fuzz '^FuzzEncodeQuoteExactInputSingleRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeQuoteExactInputSingleRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), []byte{0x00}, []byte{0x01, 0xf4}, []byte{0x00})
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, []byte{0xbe, 0xef}, []byte{0xff}, []byte{0xff, 0xff, 0xff}, []byte{0xff})

	f.Fuzz(func(t *testing.T, tokenInRaw, tokenOutRaw, amountInRaw, feeRaw, sqrtPriceLimitX96Raw []byte) {
		tokenIn := types.NewAddressFromBytes(tokenInRaw)
		tokenOut := types.NewAddressFromBytes(tokenOutRaw)
		if len(amountInRaw) > 32 {
			amountInRaw = amountInRaw[len(amountInRaw)-32:]
		}
		amountIn := new(big.Int).SetBytes(amountInRaw)
		fee := new(big.Int).Mod(new(big.Int).SetBytes(feeRaw), new(big.Int).Lsh(big.NewInt(1), 24))
		if len(sqrtPriceLimitX96Raw) > 20 {
			sqrtPriceLimitX96Raw = sqrtPriceLimitX96Raw[len(sqrtPriceLimitX96Raw)-20:]
		}
		sqrtPriceLimitX96 := new(big.Int).SetBytes(sqrtPriceLimitX96Raw)

		data, err := EncodeQuoteExactInputSingle(tokenIn, tokenOut, amountIn, fee, sqrtPriceLimitX96)
		require.NoError(t, err)

		vals, err := quoteExactInputSingleFn.DecodeCall(data)
		require.NoError(t, err)
		fields := vals[0].([]any)
		require.True(t, fields[0].(*types.Address).Equal(tokenIn))
		require.True(t, fields[1].(*types.Address).Equal(tokenOut))
		require.Zero(t, amountIn.Cmp(fields[2].(*big.Int)))
		require.Zero(t, fee.Cmp(fields[3].(*big.Int)))
		require.Zero(t, sqrtPriceLimitX96.Cmp(fields[4].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeQuoteExactInputRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeQuoteExactInputRoundTrip(f *testing.F) {
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, []byte{0x00})
	f.Add([]byte{}, []byte{0xff})

	f.Fuzz(func(t *testing.T, path, amountInRaw []byte) {
		if len(amountInRaw) > 32 {
			amountInRaw = amountInRaw[len(amountInRaw)-32:]
		}
		amountIn := new(big.Int).SetBytes(amountInRaw)

		data, err := EncodeQuoteExactInput(path, amountIn)
		require.NoError(t, err)

		vals, err := quoteExactInputFn.DecodeCall(data)
		require.NoError(t, err)
		require.Equal(t, path, vals[0].([]byte))
		require.Zero(t, amountIn.Cmp(vals[1].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeQuoteExactOutputSingleRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeQuoteExactOutputSingleRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), []byte{0x00}, []byte{0x01, 0xf4}, []byte{0x00})
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, []byte{0xbe, 0xef}, []byte{0xff}, []byte{0xff, 0xff, 0xff}, []byte{0xff})

	f.Fuzz(func(t *testing.T, tokenInRaw, tokenOutRaw, amountRaw, feeRaw, sqrtPriceLimitX96Raw []byte) {
		tokenIn := types.NewAddressFromBytes(tokenInRaw)
		tokenOut := types.NewAddressFromBytes(tokenOutRaw)
		if len(amountRaw) > 32 {
			amountRaw = amountRaw[len(amountRaw)-32:]
		}
		amount := new(big.Int).SetBytes(amountRaw)
		fee := new(big.Int).Mod(new(big.Int).SetBytes(feeRaw), new(big.Int).Lsh(big.NewInt(1), 24))
		if len(sqrtPriceLimitX96Raw) > 20 {
			sqrtPriceLimitX96Raw = sqrtPriceLimitX96Raw[len(sqrtPriceLimitX96Raw)-20:]
		}
		sqrtPriceLimitX96 := new(big.Int).SetBytes(sqrtPriceLimitX96Raw)

		data, err := EncodeQuoteExactOutputSingle(tokenIn, tokenOut, amount, fee, sqrtPriceLimitX96)
		require.NoError(t, err)

		vals, err := quoteExactOutputSingleFn.DecodeCall(data)
		require.NoError(t, err)
		fields := vals[0].([]any)
		require.True(t, fields[0].(*types.Address).Equal(tokenIn))
		require.True(t, fields[1].(*types.Address).Equal(tokenOut))
		require.Zero(t, amount.Cmp(fields[2].(*big.Int)))
		require.Zero(t, fee.Cmp(fields[3].(*big.Int)))
		require.Zero(t, sqrtPriceLimitX96.Cmp(fields[4].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeQuoteExactOutputRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeQuoteExactOutputRoundTrip(f *testing.F) {
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, []byte{0x00})
	f.Add([]byte{}, []byte{0xff})

	f.Fuzz(func(t *testing.T, path, amountOutRaw []byte) {
		if len(amountOutRaw) > 32 {
			amountOutRaw = amountOutRaw[len(amountOutRaw)-32:]
		}
		amountOut := new(big.Int).SetBytes(amountOutRaw)

		data, err := EncodeQuoteExactOutput(path, amountOut)
		require.NoError(t, err)

		vals, err := quoteExactOutputFn.DecodeCall(data)
		require.NoError(t, err)
		require.Equal(t, path, vals[0].([]byte))
		require.Zero(t, amountOut.Cmp(vals[1].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzQuoterSimpleDecodersDoNotPanic$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzQuoterSimpleDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x00})
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 64))
	f.Add(make([]byte, 128))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeQuoteExactInputSingle(data)
		_, _ = DecodeQuoteExactInput(data)
		_, _ = DecodeQuoteExactOutputSingle(data)
		_, _ = DecodeQuoteExactOutput(data)
	})
}
