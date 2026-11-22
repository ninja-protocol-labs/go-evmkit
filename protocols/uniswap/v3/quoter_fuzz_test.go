package v3

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
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

// go test -run '^$' -fuzz '^FuzzQuoterQuoteExactInputThroughClient$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzQuoterQuoteExactInputThroughClient(f *testing.F) {
	f.Add([]byte{}, []byte{}, uint8(1), uint32(3000), []byte{0x01})
	f.Add([]byte{0x01}, []byte{0x02, 0x03}, uint8(4), uint32(500), make([]byte, 32))

	f.Fuzz(func(t *testing.T, quoterRaw, fromRaw []byte, hopsRaw uint8, feeRaw uint32, amountRaw []byte) {
		if len(amountRaw) > 32 {
			amountRaw = amountRaw[len(amountRaw)-32:]
		}
		quoter := types.NewAddressFromBytes(quoterRaw)
		from := types.NewAddressFromBytes(fromRaw)
		amountIn := new(big.Int).SetBytes(amountRaw)
		hops := int(hopsRaw%5) + 1
		tokens := make([]*types.Address, hops+1)
		fees := make([]*big.Int, hops)
		for i := range tokens {
			tokens[i] = types.NewAddressFromBytes([]byte{byte(i + 1)})
		}
		for i := range fees {
			fees[i] = big.NewInt(int64(feeRaw % (1 << 24)))
		}
		path, err := EncodePath(tokens, fees)
		require.NoError(t, err)

		result, err := quoteExactInputFn.EncodeReturn(big.NewInt(1), []*big.Int{}, []uint32{}, big.NewInt(2))
		require.NoError(t, err)

		cli := mocks.NewClient(t)
		cli.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			call := e.Params.([]any)[0].(map[string]any)
			require.Equal(t, quoter.String(), call["to"])
			require.Equal(t, from.String(), call["from"])
			calldata, err := encoding.Hex.Decode(call["data"].(string))
			require.NoError(t, err)
			vals, err := quoteExactInputFn.DecodeCall(calldata)
			require.NoError(t, err)
			require.Equal(t, path, vals[0].([]byte))
			require.Zero(t, amountIn.Cmp(vals[1].(*big.Int)))
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(result)
			return nil
		}).Once()

		quote, err := NewQuoter(cli).CallQuoteExactInput(t.Context(), bind.NewCallOptions().WithFrom(from), NewQuoteExactInputRequest(quoter, path, amountIn))
		require.NoError(t, err)
		require.Zero(t, big.NewInt(1).Cmp(quote.AmountOut))
	})
}

// go test -run '^$' -fuzz '^FuzzQuoterReadsSurviveArbitraryResult$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzQuoterReadsSurviveArbitraryResult(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 128))
	f.Add(make([]byte, 8*32))

	quoter := types.MustNewAddressFromHex("0x61fFE014bA17989E743c5F6cB21bF9697530B21e")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	path := append(append(append([]byte{}, a.Bytes()...), 0x00, 0x0b, 0xb8), a.Bytes()...)

	f.Fuzz(func(t *testing.T, result []byte) {
		cli := mocks.NewClient(t)
		cli.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(result)
			return nil
		}).Times(4)
		q := NewQuoter(cli)

		inSingle, err := q.CallQuoteExactInputSingle(t.Context(), nil, NewQuoteExactInputSingleRequest(quoter, NewQuoteExactInputSingleParams(a, a, big.NewInt(1), big.NewInt(3000), big.NewInt(0))))
		require.True(t, (err == nil) != (inSingle == nil))
		in, err := q.CallQuoteExactInput(t.Context(), nil, NewQuoteExactInputRequest(quoter, path, big.NewInt(1)))
		require.True(t, (err == nil) != (in == nil))
		outSingle, err := q.CallQuoteExactOutputSingle(t.Context(), nil, NewQuoteExactOutputSingleRequest(quoter, NewQuoteExactOutputSingleParams(a, a, big.NewInt(1), big.NewInt(3000), big.NewInt(0))))
		require.True(t, (err == nil) != (outSingle == nil))
		out, err := q.CallQuoteExactOutput(t.Context(), nil, NewQuoteExactOutputRequest(quoter, path, big.NewInt(1)))
		require.True(t, (err == nil) != (out == nil))
	})
}
