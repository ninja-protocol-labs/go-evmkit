package v3

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// go test -run '^$' -fuzz '^FuzzEncodeExactInputSingleRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeExactInputSingleRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), []byte{0x01, 0xf4}, make([]byte, 20), []byte{0x00}, []byte{0x00}, []byte{0x00})
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, []byte{0xbe, 0xef}, []byte{0xff, 0xff, 0xff}, []byte{0x01}, []byte{0xff}, []byte{0xff}, []byte{0xff})

	f.Fuzz(func(t *testing.T, tokenInRaw, tokenOutRaw, feeRaw, recipientRaw, amountInRaw, amountOutMinimumRaw, sqrtPriceLimitX96Raw []byte) {
		tokenIn := types.NewAddressFromBytes(tokenInRaw)
		tokenOut := types.NewAddressFromBytes(tokenOutRaw)
		fee := new(big.Int).Mod(new(big.Int).SetBytes(feeRaw), new(big.Int).Lsh(big.NewInt(1), 24))
		recipient := types.NewAddressFromBytes(recipientRaw)
		if len(amountInRaw) > 32 {
			amountInRaw = amountInRaw[len(amountInRaw)-32:]
		}
		amountIn := new(big.Int).SetBytes(amountInRaw)
		if len(amountOutMinimumRaw) > 32 {
			amountOutMinimumRaw = amountOutMinimumRaw[len(amountOutMinimumRaw)-32:]
		}
		amountOutMinimum := new(big.Int).SetBytes(amountOutMinimumRaw)
		if len(sqrtPriceLimitX96Raw) > 20 {
			sqrtPriceLimitX96Raw = sqrtPriceLimitX96Raw[len(sqrtPriceLimitX96Raw)-20:]
		}
		sqrtPriceLimitX96 := new(big.Int).SetBytes(sqrtPriceLimitX96Raw)

		data, err := EncodeExactInputSingle(tokenIn, tokenOut, fee, recipient, amountIn, amountOutMinimum, sqrtPriceLimitX96)
		require.NoError(t, err)

		vals, err := exactInputSingleFn.DecodeCall(data)
		require.NoError(t, err)
		fields := vals[0].([]any)
		require.True(t, fields[0].(*types.Address).Equal(tokenIn))
		require.True(t, fields[1].(*types.Address).Equal(tokenOut))
		require.Zero(t, fee.Cmp(fields[2].(*big.Int)))
		require.True(t, fields[3].(*types.Address).Equal(recipient))
		require.Zero(t, amountIn.Cmp(fields[4].(*big.Int)))
		require.Zero(t, amountOutMinimum.Cmp(fields[5].(*big.Int)))
		require.Zero(t, sqrtPriceLimitX96.Cmp(fields[6].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeExactInputRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeExactInputRoundTrip(f *testing.F) {
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, make([]byte, 20), []byte{0x00}, []byte{0x00})
	f.Add([]byte{}, []byte{0xff}, []byte{0xff}, []byte{0xff})

	f.Fuzz(func(t *testing.T, path, recipientRaw, amountInRaw, amountOutMinimumRaw []byte) {
		recipient := types.NewAddressFromBytes(recipientRaw)
		if len(amountInRaw) > 32 {
			amountInRaw = amountInRaw[len(amountInRaw)-32:]
		}
		amountIn := new(big.Int).SetBytes(amountInRaw)
		if len(amountOutMinimumRaw) > 32 {
			amountOutMinimumRaw = amountOutMinimumRaw[len(amountOutMinimumRaw)-32:]
		}
		amountOutMinimum := new(big.Int).SetBytes(amountOutMinimumRaw)

		data, err := EncodeExactInput(path, recipient, amountIn, amountOutMinimum)
		require.NoError(t, err)

		vals, err := exactInputFn.DecodeCall(data)
		require.NoError(t, err)
		fields := vals[0].([]any)
		require.Equal(t, path, fields[0].([]byte))
		require.True(t, fields[1].(*types.Address).Equal(recipient))
		require.Zero(t, amountIn.Cmp(fields[2].(*big.Int)))
		require.Zero(t, amountOutMinimum.Cmp(fields[3].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeExactOutputSingleRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeExactOutputSingleRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), []byte{0x01, 0xf4}, make([]byte, 20), []byte{0x00}, []byte{0x00}, []byte{0x00})
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, []byte{0xbe, 0xef}, []byte{0xff, 0xff, 0xff}, []byte{0x01}, []byte{0xff}, []byte{0xff}, []byte{0xff})

	f.Fuzz(func(t *testing.T, tokenInRaw, tokenOutRaw, feeRaw, recipientRaw, amountOutRaw, amountInMaximumRaw, sqrtPriceLimitX96Raw []byte) {
		tokenIn := types.NewAddressFromBytes(tokenInRaw)
		tokenOut := types.NewAddressFromBytes(tokenOutRaw)
		fee := new(big.Int).Mod(new(big.Int).SetBytes(feeRaw), new(big.Int).Lsh(big.NewInt(1), 24))
		recipient := types.NewAddressFromBytes(recipientRaw)
		if len(amountOutRaw) > 32 {
			amountOutRaw = amountOutRaw[len(amountOutRaw)-32:]
		}
		amountOut := new(big.Int).SetBytes(amountOutRaw)
		if len(amountInMaximumRaw) > 32 {
			amountInMaximumRaw = amountInMaximumRaw[len(amountInMaximumRaw)-32:]
		}
		amountInMaximum := new(big.Int).SetBytes(amountInMaximumRaw)
		if len(sqrtPriceLimitX96Raw) > 20 {
			sqrtPriceLimitX96Raw = sqrtPriceLimitX96Raw[len(sqrtPriceLimitX96Raw)-20:]
		}
		sqrtPriceLimitX96 := new(big.Int).SetBytes(sqrtPriceLimitX96Raw)

		data, err := EncodeExactOutputSingle(tokenIn, tokenOut, fee, recipient, amountOut, amountInMaximum, sqrtPriceLimitX96)
		require.NoError(t, err)

		vals, err := exactOutputSingleFn.DecodeCall(data)
		require.NoError(t, err)
		fields := vals[0].([]any)
		require.True(t, fields[0].(*types.Address).Equal(tokenIn))
		require.True(t, fields[1].(*types.Address).Equal(tokenOut))
		require.Zero(t, fee.Cmp(fields[2].(*big.Int)))
		require.True(t, fields[3].(*types.Address).Equal(recipient))
		require.Zero(t, amountOut.Cmp(fields[4].(*big.Int)))
		require.Zero(t, amountInMaximum.Cmp(fields[5].(*big.Int)))
		require.Zero(t, sqrtPriceLimitX96.Cmp(fields[6].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeExactOutputRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeExactOutputRoundTrip(f *testing.F) {
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, make([]byte, 20), []byte{0x00}, []byte{0x00})
	f.Add([]byte{}, []byte{0xff}, []byte{0xff}, []byte{0xff})

	f.Fuzz(func(t *testing.T, path, recipientRaw, amountOutRaw, amountInMaximumRaw []byte) {
		recipient := types.NewAddressFromBytes(recipientRaw)
		if len(amountOutRaw) > 32 {
			amountOutRaw = amountOutRaw[len(amountOutRaw)-32:]
		}
		amountOut := new(big.Int).SetBytes(amountOutRaw)
		if len(amountInMaximumRaw) > 32 {
			amountInMaximumRaw = amountInMaximumRaw[len(amountInMaximumRaw)-32:]
		}
		amountInMaximum := new(big.Int).SetBytes(amountInMaximumRaw)

		data, err := EncodeExactOutput(path, recipient, amountOut, amountInMaximum)
		require.NoError(t, err)

		vals, err := exactOutputFn.DecodeCall(data)
		require.NoError(t, err)
		fields := vals[0].([]any)
		require.Equal(t, path, fields[0].([]byte))
		require.True(t, fields[1].(*types.Address).Equal(recipient))
		require.Zero(t, amountOut.Cmp(fields[2].(*big.Int)))
		require.Zero(t, amountInMaximum.Cmp(fields[3].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeRouterUnwrapWETH9RoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeRouterUnwrapWETH9RoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, make([]byte, 20))
	f.Add([]byte{0xff}, []byte{0xde, 0xad, 0xbe, 0xef})

	f.Fuzz(func(t *testing.T, amountMinimumRaw, recipientRaw []byte) {
		if len(amountMinimumRaw) > 32 {
			amountMinimumRaw = amountMinimumRaw[len(amountMinimumRaw)-32:]
		}
		amountMinimum := new(big.Int).SetBytes(amountMinimumRaw)
		recipient := types.NewAddressFromBytes(recipientRaw)

		data, err := EncodeRouterUnwrapWETH9(amountMinimum, recipient)
		require.NoError(t, err)

		vals, err := routerUnwrapWETH9Fn.DecodeCall(data)
		require.NoError(t, err)
		require.Zero(t, amountMinimum.Cmp(vals[0].(*big.Int)))
		require.True(t, vals[1].(*types.Address).Equal(recipient))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeUnwrapWETH9WithFeeRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeUnwrapWETH9WithFeeRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, make([]byte, 20), []byte{0x32}, make([]byte, 20))
	f.Add([]byte{0xff}, []byte{0xde, 0xad, 0xbe, 0xef}, []byte{0xff}, []byte{0xbe, 0xef})

	f.Fuzz(func(t *testing.T, amountMinimumRaw, recipientRaw, feeBipsRaw, feeRecipientRaw []byte) {
		if len(amountMinimumRaw) > 32 {
			amountMinimumRaw = amountMinimumRaw[len(amountMinimumRaw)-32:]
		}
		amountMinimum := new(big.Int).SetBytes(amountMinimumRaw)
		recipient := types.NewAddressFromBytes(recipientRaw)
		if len(feeBipsRaw) > 32 {
			feeBipsRaw = feeBipsRaw[len(feeBipsRaw)-32:]
		}
		feeBips := new(big.Int).SetBytes(feeBipsRaw)
		feeRecipient := types.NewAddressFromBytes(feeRecipientRaw)

		data, err := EncodeUnwrapWETH9WithFee(amountMinimum, recipient, feeBips, feeRecipient)
		require.NoError(t, err)

		vals, err := unwrapWETH9WithFeeFn.DecodeCall(data)
		require.NoError(t, err)
		require.Zero(t, amountMinimum.Cmp(vals[0].(*big.Int)))
		require.True(t, vals[1].(*types.Address).Equal(recipient))
		require.Zero(t, feeBips.Cmp(vals[2].(*big.Int)))
		require.True(t, vals[3].(*types.Address).Equal(feeRecipient))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeRouterSweepTokenRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeRouterSweepTokenRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), []byte{0x00}, make([]byte, 20))
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, []byte{0xff}, []byte{0xbe, 0xef})

	f.Fuzz(func(t *testing.T, tokenRaw, amountMinimumRaw, recipientRaw []byte) {
		token := types.NewAddressFromBytes(tokenRaw)
		if len(amountMinimumRaw) > 32 {
			amountMinimumRaw = amountMinimumRaw[len(amountMinimumRaw)-32:]
		}
		amountMinimum := new(big.Int).SetBytes(amountMinimumRaw)
		recipient := types.NewAddressFromBytes(recipientRaw)

		data, err := EncodeRouterSweepToken(token, amountMinimum, recipient)
		require.NoError(t, err)

		vals, err := routerSweepTokenFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(token))
		require.Zero(t, amountMinimum.Cmp(vals[1].(*big.Int)))
		require.True(t, vals[2].(*types.Address).Equal(recipient))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeSweepTokenWithFeeRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeSweepTokenWithFeeRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), []byte{0x00}, make([]byte, 20), []byte{0x32}, make([]byte, 20))
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, []byte{0xff}, []byte{0xbe, 0xef}, []byte{0xff}, []byte{0xaa})

	f.Fuzz(func(t *testing.T, tokenRaw, amountMinimumRaw, recipientRaw, feeBipsRaw, feeRecipientRaw []byte) {
		token := types.NewAddressFromBytes(tokenRaw)
		if len(amountMinimumRaw) > 32 {
			amountMinimumRaw = amountMinimumRaw[len(amountMinimumRaw)-32:]
		}
		amountMinimum := new(big.Int).SetBytes(amountMinimumRaw)
		recipient := types.NewAddressFromBytes(recipientRaw)
		if len(feeBipsRaw) > 32 {
			feeBipsRaw = feeBipsRaw[len(feeBipsRaw)-32:]
		}
		feeBips := new(big.Int).SetBytes(feeBipsRaw)
		feeRecipient := types.NewAddressFromBytes(feeRecipientRaw)

		data, err := EncodeSweepTokenWithFee(token, amountMinimum, recipient, feeBips, feeRecipient)
		require.NoError(t, err)

		vals, err := sweepTokenWithFeeFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(token))
		require.Zero(t, amountMinimum.Cmp(vals[1].(*big.Int)))
		require.True(t, vals[2].(*types.Address).Equal(recipient))
		require.Zero(t, feeBips.Cmp(vals[3].(*big.Int)))
		require.True(t, vals[4].(*types.Address).Equal(feeRecipient))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeRouterMulticallRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeRouterMulticallRoundTrip(f *testing.F) {
	f.Add([]byte{0xde, 0xad}, []byte{0xbe, 0xef})
	f.Add([]byte{}, []byte{})

	f.Fuzz(func(t *testing.T, a, b []byte) {
		calls := [][]byte{a, b}

		data, err := EncodeRouterMulticall(calls)
		require.NoError(t, err)

		vals, err := routerMulticallFn.DecodeCall(data)
		require.NoError(t, err)
		elems := vals[0].([]any)
		require.Len(t, elems, 2)
		require.Equal(t, a, elems[0].([]byte))
		require.Equal(t, b, elems[1].([]byte))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeRouterMulticallRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzDecodeRouterMulticallRoundTrip(f *testing.F) {
	f.Add([]byte{0xde, 0xad}, []byte{0xbe, 0xef})
	f.Add([]byte{}, []byte{})

	f.Fuzz(func(t *testing.T, a, b []byte) {
		results := [][]byte{a, b}

		data, err := routerMulticallFn.EncodeReturn(results)
		require.NoError(t, err)

		got, err := DecodeRouterMulticall(data)
		require.NoError(t, err)
		require.Len(t, got, 2)
		require.Equal(t, a, got[0])
		require.Equal(t, b, got[1])
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeSelfPermitRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeSelfPermitRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), []byte{0x00}, []byte{0x00})
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, []byte{0xff}, []byte{0xff})

	f.Fuzz(func(t *testing.T, tokenRaw, valueRaw, deadlineRaw []byte) {
		token := types.NewAddressFromBytes(tokenRaw)
		if len(valueRaw) > 32 {
			valueRaw = valueRaw[len(valueRaw)-32:]
		}
		value := new(big.Int).SetBytes(valueRaw)
		if len(deadlineRaw) > 32 {
			deadlineRaw = deadlineRaw[len(deadlineRaw)-32:]
		}
		deadline := new(big.Int).SetBytes(deadlineRaw)
		key, err := types.GeneratePrivateKey()
		require.NoError(t, err)
		sig, err := key.Sign(types.NewHashFromBytes([]byte("some 32 byte digest padded out!")))
		require.NoError(t, err)

		data, err := EncodeSelfPermit(token, value, deadline, sig)
		require.NoError(t, err)

		vals, err := selfPermitFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(token))
		require.Zero(t, value.Cmp(vals[1].(*big.Int)))
		require.Zero(t, deadline.Cmp(vals[2].(*big.Int)))
		legacy := sig.LegacyBytes()
		require.Equal(t, legacy[64], vals[3].(uint8))
		require.Equal(t, legacy[:32], vals[4].([]byte))
		require.Equal(t, legacy[32:64], vals[5].([]byte))
	})
}

// go test -run '^$' -fuzz '^FuzzRouterSimpleDecodersDoNotPanic$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzRouterSimpleDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x00})
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 64))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeRouterFactory(data)
		_, _ = DecodeRouterWETH9(data)
		_, _ = DecodeExactInputSingle(data)
		_, _ = DecodeExactInput(data)
		_, _ = DecodeExactOutputSingle(data)
		_, _ = DecodeExactOutput(data)
		_, _ = DecodeRouterMulticall(data)
	})
}

// go test -run '^$' -fuzz '^FuzzRouterMulticallThroughClient$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzRouterMulticallThroughClient(f *testing.F) {
	f.Add([]byte{}, []byte{0x01}, []byte{}, []byte{}, uint64(0), true)
	f.Add([]byte{0x01}, []byte{0xaa, 0xbb}, []byte{0x65, 0x5f, 0x0e, 0x00}, []byte{0x01}, uint64(42), false)
	f.Add(make([]byte, 20), make([]byte, 100), make([]byte, 32), make([]byte, 16), ^uint64(0)-1, true)

	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(f, err)
	signer := core.PubkeyToAddress(key.PublicKey())

	f.Fuzz(func(t *testing.T, routerRaw, callRaw, deadlineRaw, nativeRaw []byte, nonce uint64, dynamic bool) {
		if len(callRaw) == 0 {
			callRaw = []byte{0x00}
		}
		if len(deadlineRaw) > 32 {
			deadlineRaw = deadlineRaw[len(deadlineRaw)-32:]
		}
		if len(nativeRaw) > 16 {
			nativeRaw = nativeRaw[len(nativeRaw)-16:]
		}
		router := types.NewAddressFromBytes(routerRaw)
		calls := [][]byte{callRaw, append([]byte{0x12}, callRaw...)}
		var deadline, native *big.Int
		if len(deadlineRaw) > 0 {
			deadline = new(big.Int).SetBytes(deadlineRaw)
		}
		if len(nativeRaw) > 0 {
			native = new(big.Int).SetBytes(nativeRaw)
		}

		var builder tx.Builder
		if dynamic {
			builder = tx.NewDynamicFeeBuilder()
		} else {
			builder = tx.NewLegacyBuilder()
		}
		builder.ChainID(big.NewInt(1))
		builder.Nonce(nonce)
		builder.Limit(3_000_000)
		builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

		var raw []byte
		cli := mocks.NewClient(t)
		cli.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			raw, err = encoding.Hex.Decode(e.Params.([]any)[0].(string))
			require.NoError(t, err)
			*e.Result.(*string) = core.Keccak256(raw).String()
			return nil
		}).Once()

		opts := bind.NewTransactOptions(key, builder)
		signed, hash, err := NewRouter(cli).Multicall(t.Context(), opts, NewRouterMulticallRequest(router, calls, deadline, native))
		require.NoError(t, err)
		require.True(t, hash.Equal(core.Keccak256(raw)))
		require.Equal(t, nonce, signed.AddressNonce())
		require.Nil(t, opts.Value)

		decoded, err := core.DecodeTransaction(raw)
		require.NoError(t, err)
		from, err := decoded.Sender()
		require.NoError(t, err)
		require.True(t, from.Equal(signer))

		var gotTo *types.Address
		var gotValue *big.Int
		var gotData []byte
		switch v := decoded.(type) {
		case *core.DynamicFeeTx:
			require.True(t, dynamic)
			gotTo, gotValue, gotData = v.To, v.Value, v.Data
		case *core.LegacyTx:
			require.False(t, dynamic)
			gotTo, gotValue, gotData = v.To, v.Value, v.Data
		default:
			t.Fatalf("unexpected tx type %T", decoded)
		}
		require.True(t, gotTo.Equal(router))
		if native == nil {
			require.Zero(t, gotValue.Sign())
		} else {
			require.Zero(t, native.Cmp(gotValue))
		}

		if deadline == nil {
			want, err := EncodeRouterMulticall(calls)
			require.NoError(t, err)
			require.Equal(t, want, gotData)
		} else {
			vals, err := routerMulticallDeadlineFn.DecodeCall(gotData)
			require.NoError(t, err)
			require.Zero(t, deadline.Cmp(vals[0].(*big.Int)))
			gotCalls := vals[1].([]any)
			require.Len(t, gotCalls, len(calls))
			for i, call := range calls {
				require.Equal(t, call, gotCalls[i].([]byte))
			}
		}
	})
}

// go test -run '^$' -fuzz '^FuzzRouterReadsSurviveArbitraryResult$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzRouterReadsSurviveArbitraryResult(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(append(append(make([]byte, 31), 0x20), append(make([]byte, 31), 0x01)...))

	router := types.MustNewAddressFromHex("0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	path := append(append(append([]byte{}, a.Bytes()...), 0x00, 0x0b, 0xb8), a.Bytes()...)

	f.Fuzz(func(t *testing.T, result []byte) {
		cli := mocks.NewClient(t)
		cli.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(result)
			return nil
		}).Times(4)
		r := NewRouter(cli)

		weth9, err := r.WETH9(t.Context(), nil, NewWETH9Request(router))
		require.True(t, (err == nil) != (weth9 == nil))
		amountOut, err := r.CallExactInputSingle(t.Context(), nil, NewExactInputSingleRequest(router, NewExactInputSingleParams(a, a, big.NewInt(3000), a, big.NewInt(1), big.NewInt(0), big.NewInt(0)), nil))
		require.True(t, (err == nil) != (amountOut == nil))
		amountIn, err := r.CallExactOutput(t.Context(), nil, NewExactOutputRequest(router, NewExactOutputParams(path, a, big.NewInt(1), big.NewInt(1)), nil))
		require.True(t, (err == nil) != (amountIn == nil))
		results, err := r.CallMulticall(t.Context(), nil, NewRouterMulticallRequest(router, [][]byte{{0x01}}, nil, nil))
		require.True(t, err == nil || results == nil)
	})
}
