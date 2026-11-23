package v4

import (
	"context"
	"encoding/json/jsontext"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// go test -run '^$' -fuzz '^FuzzAllQuoterDecodersDoNotPanic$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzAllQuoterDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x01})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 33))
	f.Add(make([]byte, 63))
	f.Add(make([]byte, 64))
	f.Add(make([]byte, 96))

	f.Fuzz(func(t *testing.T, data []byte) {
		require.NotPanics(t, func() {
			_, _ = DecodeQuoteExactInputSingle(data)
			_, _ = DecodeQuoteExactInput(data)
			_, _ = DecodeQuoteExactOutputSingle(data)
			_, _ = DecodeQuoteExactOutput(data)
			_, _ = DecodeQuoterPoolManager(data)
			_, _ = DecodeQuoterMsgSender(data)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzQuoterWrapRevertErrorDoesNotPanic$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzQuoterWrapRevertErrorDoesNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x61, 0x90, 0xb2, 0xb0})
	f.Add(append([]byte{0x61, 0x90, 0xb2, 0xb0}, make([]byte, 32)...))
	f.Add(append([]byte{0x61, 0x90, 0xb2, 0xb0}, make([]byte, 64)...))
	f.Add([]byte{0x7a, 0x5e, 0xd7, 0x34})
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef})

	f.Fuzz(func(t *testing.T, data []byte) {
		rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(data) + `"`)}
		require.NotPanics(t, func() {
			wrapped := (&Quoter{}).wrapRevertError(rpcErr)
			if len(data) < 4 {
				require.Same(t, rpcErr, wrapped)
				return
			}
			var got *rpc.ResponseError
			require.ErrorAs(t, wrapped, &got)
			require.Same(t, rpcErr, got)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzQuoterWrapRevertErrorUnexpectedRevertBytesRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzQuoterWrapRevertErrorUnexpectedRevertBytesRoundTrip(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x48, 0x6a, 0xa3, 0x07})
	f.Add([]byte{0x7a, 0x5e, 0xd7, 0x34, 0x01})
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef})
	f.Add(make([]byte, 100))

	f.Fuzz(func(t *testing.T, inner []byte) {
		outerSel := errUnexpectedRevertBytes.Selector()
		outerArgs, err := abi.Pack(errUnexpectedRevertBytes.Inputs, inner)
		require.NoError(t, err)
		revertData := append(outerSel[:], outerArgs...)
		rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(revertData) + `"`)}

		wrapped := (&Quoter{}).wrapRevertError(rpcErr)
		require.ErrorIs(t, wrapped, ErrUnexpectedRevertBytes)
		require.NotErrorIs(t, wrapped, ErrUnknown)
		var got *rpc.ResponseError
		require.ErrorAs(t, wrapped, &got)
		require.Same(t, rpcErr, got)

		var innerSel abi.Selector
		if len(inner) >= 4 {
			copy(innerSel[:], inner[:4])
		}
		require.Equal(t, len(inner) >= 4 && innerSel == errPoolNotInitialized.Selector(), errors.Is(wrapped, ErrPoolNotInitialized))
		require.Equal(t, len(inner) >= 4 && innerSel == errSwapAmountCannotBeZero.Selector(), errors.Is(wrapped, ErrSwapAmountCannotBeZero))
		require.Equal(t, len(inner) >= 4 && innerSel == errSafeCastOverflow.Selector(), errors.Is(wrapped, ErrSafeCastOverflow))
		require.Equal(t, len(inner) >= 4 && innerSel == errWrappedError.Selector(), errors.Is(wrapped, ErrWrappedError))
		require.Equal(t, len(inner) >= 4 && innerSel == errPriceLimitAlreadyExceeded.Selector(), errors.Is(wrapped, ErrPriceLimitAlreadyExceeded))
		require.Equal(t, len(inner) >= 4 && (innerSel == errNotEnoughLiquidity.Selector() || innerSel == errSqrtPriceMathNotEnoughLiquidity.Selector()), errors.Is(wrapped, ErrNotEnoughLiquidity))
		if errors.Is(wrapped, ErrHookCallFailed) {
			require.ErrorIs(t, wrapped, ErrWrappedError)
		}
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeQuoteExactSingleRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeQuoteExactSingleRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x11}, []byte{0x22}, []byte{0x0b, 0xb8}, []byte{0xff, 0xff, 0xc4}, true, []byte{0x03, 0xe8}, []byte{0xde, 0xad})
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 20), []byte{}, []byte{}, false, []byte{}, []byte{})
	f.Add([]byte{0xff}, []byte{0xff}, []byte{0xff}, []byte{0xff, 0xff, 0xff}, []byte{0x7f, 0xff, 0xff}, true, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, make([]byte, 65))
	f.Add([]byte{0x01}, []byte{0x02}, []byte{0x03}, []byte{0x01}, []byte{0x80, 0x00, 0x00}, false, []byte{0x01}, []byte{0x00})

	f.Fuzz(func(t *testing.T, currency0Raw, currency1Raw, hooksRaw, feeRaw, tickSpacingRaw []byte, zeroForOne bool, exactAmountRaw, hookData []byte) {
		currency0 := types.NewAddressFromBytes(currency0Raw)
		currency1 := types.NewAddressFromBytes(currency1Raw)
		hooks := types.NewAddressFromBytes(hooksRaw)

		feeClamped := feeRaw
		if len(feeClamped) > 3 {
			feeClamped = feeClamped[len(feeClamped)-3:]
		}
		fee := new(big.Int).SetBytes(feeClamped)

		tickSpacingClamped := tickSpacingRaw
		if len(tickSpacingClamped) > 3 {
			tickSpacingClamped = tickSpacingClamped[len(tickSpacingClamped)-3:]
		}
		tickSpacing := new(big.Int).SetBytes(tickSpacingClamped)
		if len(tickSpacingClamped) == 3 && tickSpacingClamped[0]&0x80 != 0 {
			tickSpacing.Sub(tickSpacing, new(big.Int).Lsh(big.NewInt(1), 24))
		}

		exactAmountClamped := exactAmountRaw
		if len(exactAmountClamped) > 16 {
			exactAmountClamped = exactAmountClamped[len(exactAmountClamped)-16:]
		}
		exactAmount := new(big.Int).SetBytes(exactAmountClamped)

		poolKey := NewPoolKey(currency0, currency1, fee, tickSpacing, hooks)
		params := NewQuoteExactSingleParams(poolKey, zeroForOne, exactAmount, hookData)

		in, err := EncodeQuoteExactInputSingle(params)
		require.NoError(t, err)
		out, err := EncodeQuoteExactOutputSingle(params)
		require.NoError(t, err)
		require.Equal(t, in[4:], out[4:])
		require.Zero(t, (len(in)-4)%32)
		require.Len(t, in, 4+32*10+((len(hookData)+31)/32)*32)

		vals, err := quoteExactInputSingleFn.DecodeCall(in)
		require.NoError(t, err)
		fields := vals[0].([]any)
		poolKeyFields := fields[0].([]any)
		require.True(t, poolKeyFields[0].(*types.Address).Equal(currency0))
		require.True(t, poolKeyFields[1].(*types.Address).Equal(currency1))
		require.Zero(t, fee.Cmp(poolKeyFields[2].(*big.Int)))
		require.Zero(t, tickSpacing.Cmp(poolKeyFields[3].(*big.Int)))
		require.True(t, poolKeyFields[4].(*types.Address).Equal(hooks))
		require.Equal(t, zeroForOne, fields[1])
		require.Zero(t, exactAmount.Cmp(fields[2].(*big.Int)))
		if len(hookData) == 0 {
			require.Empty(t, fields[3])
		} else {
			require.Equal(t, hookData, fields[3])
		}
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeQuoteExactRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeQuoteExactRoundTrip(f *testing.F) {
	f.Add([]byte{0x11}, []byte{}, []byte{0x03, 0xe8})
	f.Add(make([]byte, 20), make([]byte, 49), []byte{})
	f.Add([]byte{0xff}, make([]byte, 49*3), []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	f.Add([]byte{0x01, 0x02}, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x00, 0x0b, 0xb8, 0xff, 0xff, 0xc4, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f, 0x30, 0x31, 0x32, 0x33, 0x34, 0x00, 0x00, 0x05}, []byte{0x30, 0x39})

	f.Fuzz(func(t *testing.T, exactCurrencyRaw, pathRaw, exactAmountRaw []byte) {
		exactCurrency := types.NewAddressFromBytes(exactCurrencyRaw)

		exactAmountClamped := exactAmountRaw
		if len(exactAmountClamped) > 16 {
			exactAmountClamped = exactAmountClamped[len(exactAmountClamped)-16:]
		}
		exactAmount := new(big.Int).SetBytes(exactAmountClamped)

		hops := min(len(pathRaw)/49, 8)
		path := make([]*PathKey, hops)
		for i := range path {
			chunk := pathRaw[i*49 : (i+1)*49]
			tickSpacing := new(big.Int).SetBytes(chunk[23:26])
			if chunk[23]&0x80 != 0 {
				tickSpacing.Sub(tickSpacing, new(big.Int).Lsh(big.NewInt(1), 24))
			}
			hookData := chunk[:int(chunk[48])%49]
			path[i] = NewPathKey(types.NewAddressFromBytes(chunk[0:20]), new(big.Int).SetBytes(chunk[20:23]), tickSpacing, types.NewAddressFromBytes(chunk[26:46]), hookData)
		}
		params := NewQuoteExactParams(exactCurrency, path, exactAmount)

		in, err := EncodeQuoteExactInput(params)
		require.NoError(t, err)
		out, err := EncodeQuoteExactOutput(params)
		require.NoError(t, err)
		require.Equal(t, in[4:], out[4:])
		require.Zero(t, (len(in)-4)%32)

		for _, c := range []struct {
			fn   *abi.Function
			data []byte
		}{{quoteExactInputFn, in}, {quoteExactOutputFn, out}} {
			vals, err := c.fn.DecodeCall(c.data)
			require.NoError(t, err)
			fields := vals[0].([]any)
			require.True(t, fields[0].(*types.Address).Equal(exactCurrency))
			pathVals := fields[1].([]any)
			require.Len(t, pathVals, hops)
			for i, v := range pathVals {
				k := v.([]any)
				require.True(t, k[0].(*types.Address).Equal(path[i].IntermediateCurrency))
				require.Zero(t, path[i].Fee.Cmp(k[1].(*big.Int)))
				require.Zero(t, path[i].TickSpacing.Cmp(k[2].(*big.Int)))
				require.True(t, k[3].(*types.Address).Equal(path[i].Hooks))
				if len(path[i].HookData) == 0 {
					require.Empty(t, k[4])
				} else {
					require.Equal(t, path[i].HookData, k[4])
				}
			}
			require.Zero(t, exactAmount.Cmp(fields[2].(*big.Int)))
		}
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeQuoteResultRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzDecodeQuoteResultRoundTrip(f *testing.F) {
	f.Add([]byte{}, []byte{})
	f.Add([]byte{0x03, 0xe5}, []byte{0x01, 0x4c, 0x08})
	f.Add(make([]byte, 32), make([]byte, 32))
	f.Add([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, []byte{0x01})

	f.Fuzz(func(t *testing.T, amountRaw, gasEstimateRaw []byte) {
		amountClamped := amountRaw
		if len(amountClamped) > 32 {
			amountClamped = amountClamped[len(amountClamped)-32:]
		}
		amount := new(big.Int).SetBytes(amountClamped)

		gasEstimateClamped := gasEstimateRaw
		if len(gasEstimateClamped) > 32 {
			gasEstimateClamped = gasEstimateClamped[len(gasEstimateClamped)-32:]
		}
		gasEstimate := new(big.Int).SetBytes(gasEstimateClamped)

		data, err := quoteExactInputSingleFn.EncodeReturn(amount, gasEstimate)
		require.NoError(t, err)
		require.Len(t, data, 64)

		inSingle, err := DecodeQuoteExactInputSingle(data)
		require.NoError(t, err)
		require.Zero(t, amount.Cmp(inSingle.AmountOut))
		require.Zero(t, gasEstimate.Cmp(inSingle.GasEstimate))

		in, err := DecodeQuoteExactInput(data)
		require.NoError(t, err)
		require.Zero(t, amount.Cmp(in.AmountOut))
		require.Zero(t, gasEstimate.Cmp(in.GasEstimate))

		outSingle, err := DecodeQuoteExactOutputSingle(data)
		require.NoError(t, err)
		require.Zero(t, amount.Cmp(outSingle.AmountIn))
		require.Zero(t, gasEstimate.Cmp(outSingle.GasEstimate))

		out, err := DecodeQuoteExactOutput(data)
		require.NoError(t, err)
		require.Zero(t, amount.Cmp(out.AmountIn))
		require.Zero(t, gasEstimate.Cmp(out.GasEstimate))

		_, err = DecodeQuoteExactInputSingle(data[:63])
		require.Error(t, err)
	})
}

// go test -run '^$' -fuzz '^FuzzQuoterExactInputSingleThroughClient$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzQuoterExactInputSingleThroughClient(f *testing.F) {
	f.Add(true, uint64(0), []byte{}, uint64(0))
	f.Add(false, uint64(1_000_000), []byte{0x01, 0x02}, uint64(1))
	f.Add(true, uint64(1<<63), make([]byte, 64), uint64(1<<40))

	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)

	f.Fuzz(func(t *testing.T, zeroForOne bool, amountRaw uint64, hookData []byte, block uint64) {
		amount := new(big.Int).SetUint64(amountRaw)
		blockTag := fmt.Sprintf("0x%x", block)
		word, err := abi.Pack(abi.NewTypes(abi.Uint256, abi.Uint256), new(big.Int).Add(amount, big.NewInt(1)), big.NewInt(21_000))
		require.NoError(t, err)

		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			params := e.Params.([]any)
			require.Equal(t, blockTag, params[1])
			call := params[0].(map[string]any)
			require.Equal(t, testPoolManager.String(), call["to"])
			require.Equal(t, testSender.String(), call["from"])
			calldata, err := encoding.Hex.Decode(call["data"].(string))
			require.NoError(t, err)
			vals, err := quoteExactInputSingleFn.DecodeCall(calldata)
			require.NoError(t, err)
			tuple := vals[0].([]any)
			require.Equal(t, zeroForOne, tuple[1].(bool))
			require.Zero(t, amount.Cmp(tuple[2].(*big.Int)))
			require.Equal(t, append([]byte{}, hookData...), tuple[3].([]byte))
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(word)
			return nil
		}).Once()

		quote, err := NewQuoter(c).CallQuoteExactInputSingle(t.Context(), bind.NewCallOptions().WithFrom(testSender).WithBlock(blockTag), NewQuoteExactInputSingleRequest(testPoolManager, NewQuoteExactSingleParams(key, zeroForOne, amount, hookData)))
		require.NoError(t, err)
		require.Zero(t, new(big.Int).Add(amount, big.NewInt(1)).Cmp(quote.AmountOut))
		require.Zero(t, big.NewInt(21_000).Cmp(quote.GasEstimate))
	})
}

// go test -run '^$' -fuzz '^FuzzQuoterReadsSurviveArbitraryResult$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzQuoterReadsSurviveArbitraryResult(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 64))
	f.Add(append(make([]byte, 63), 0xff))

	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)

	f.Fuzz(func(t *testing.T, result []byte) {
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(result)
			return nil
		}).Times(6)
		q := NewQuoter(c)
		ctx := t.Context()

		poolManager, err := q.PoolManager(ctx, nil, NewPoolManagerRequest(testPoolManager))
		require.True(t, (err == nil) != (poolManager == nil))
		msgSender, err := q.MsgSender(ctx, nil, NewMsgSenderRequest(testPoolManager))
		require.True(t, (err == nil) != (msgSender == nil))
		inSingle, err := q.CallQuoteExactInputSingle(ctx, nil, NewQuoteExactInputSingleRequest(testPoolManager, NewQuoteExactSingleParams(key, true, big.NewInt(1), nil)))
		require.True(t, (err == nil) != (inSingle == nil))
		in, err := q.CallQuoteExactInput(ctx, nil, NewQuoteExactInputRequest(testPoolManager, NewQuoteExactParams(testCurrency0, nil, big.NewInt(1))))
		require.True(t, (err == nil) != (in == nil))
		outSingle, err := q.CallQuoteExactOutputSingle(ctx, nil, NewQuoteExactOutputSingleRequest(testPoolManager, NewQuoteExactSingleParams(key, false, big.NewInt(1), nil)))
		require.True(t, (err == nil) != (outSingle == nil))
		out, err := q.CallQuoteExactOutput(ctx, nil, NewQuoteExactOutputRequest(testPoolManager, NewQuoteExactParams(testCurrency0, nil, big.NewInt(1))))
		require.True(t, (err == nil) != (out == nil))
	})
}
