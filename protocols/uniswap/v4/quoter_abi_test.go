package v4

import (
	"encoding/json/jsontext"
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

func TestQuoterFunctionSelectors(t *testing.T) {
	cases := []struct {
		name string
		fn   *abi.Function
		sig  string
		sel  string
	}{
		{"quoteExactInputSingle", quoteExactInputSingleFn, "quoteExactInputSingle(((address,address,uint24,int24,address),bool,uint128,bytes))", "0xaa9d21cb"},
		{"quoteExactInput", quoteExactInputFn, "quoteExactInput((address,(address,uint24,int24,address,bytes)[],uint128))", "0xca253dc9"},
		{"quoteExactOutputSingle", quoteExactOutputSingleFn, "quoteExactOutputSingle(((address,address,uint24,int24,address),bool,uint128,bytes))", "0x58733073"},
		{"quoteExactOutput", quoteExactOutputFn, "quoteExactOutput((address,(address,uint24,int24,address,bytes)[],uint128))", "0x147d2af9"},
		{"poolManager", quoterPoolManagerFn, "poolManager()", "0xdc4c90d3"},
		{"msgSender", quoterMsgSenderFn, "msgSender()", "0xd737d0c7"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.sig, c.fn.Signature())
			sel := c.fn.Selector()
			require.Equal(t, c.sel, encoding.Hex.EncodePrefixed(sel[:]))
		})
	}
}

func TestQuoterErrorSelectors(t *testing.T) {
	cases := []struct {
		name string
		err  *abi.Error
		sig  string
		sel  string
	}{
		{"NotEnoughLiquidity", errNotEnoughLiquidity, "NotEnoughLiquidity(bytes32)", "0x7a5ed734"},
		{"NotSelf", errNotSelf, "NotSelf()", "0x29c3b7ee"},
		{"UnexpectedCallSuccess", errUnexpectedCallSuccess, "UnexpectedCallSuccess()", "0xe0752a5a"},
		{"UnexpectedRevertBytes", errUnexpectedRevertBytes, "UnexpectedRevertBytes(bytes)", "0x6190b2b0"},
		{"SafeCastOverflow", errSafeCastOverflow, "SafeCastOverflow()", "0x93dafdf1"},
		{"PriceLimitAlreadyExceeded", errPriceLimitAlreadyExceeded, "PriceLimitAlreadyExceeded(uint160,uint160)", "0x7c9c6e8f"},
		{"PriceLimitOutOfBounds", errPriceLimitOutOfBounds, "PriceLimitOutOfBounds(uint160)", "0x9e4d7cc7"},
		{"InvalidFeeForExactOut", errInvalidFeeForExactOut, "InvalidFeeForExactOut()", "0x96206246"},
		{"InvalidHookResponse", errInvalidHookResponse, "InvalidHookResponse()", "0x1e048e1d"},
		{"HookCallFailed", errHookCallFailed, "HookCallFailed()", "0xa9e35b2f"},
		{"HookDeltaExceedsSwapAmount", errHookDeltaExceedsSwapAmount, "HookDeltaExceedsSwapAmount()", "0xfa0b71d6"},
		{"LPFeeTooLarge", errLPFeeTooLarge, "LPFeeTooLarge(uint24)", "0x14002113"},
		{"WrappedError", errWrappedError, "WrappedError(address,bytes4,bytes,bytes)", "0x90bfb865"},
		{"InvalidPriceOrLiquidity", errInvalidPriceOrLiquidity, "InvalidPriceOrLiquidity()", "0x4f2461b8"},
		{"InvalidPrice", errInvalidPrice, "InvalidPrice()", "0x00bfc921"},
		{"PriceOverflow", errPriceOverflow, "PriceOverflow()", "0xf5c787f1"},
		{"SqrtPriceMathNotEnoughLiquidity", errSqrtPriceMathNotEnoughLiquidity, "NotEnoughLiquidity()", "0x4323a555"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.sig, c.err.Signature())
			sel := c.err.Selector()
			require.Equal(t, c.sel, encoding.Hex.EncodePrefixed(sel[:]))
		})
	}
}

func TestQuoterErrorSelectorsDoNotCollideWithPoolManagerErrors(t *testing.T) {
	all := []*abi.Error{
		errNotEnoughLiquidity, errNotSelf, errUnexpectedCallSuccess, errUnexpectedRevertBytes,
		errCurrencyNotSettled, errPoolNotInitialized, errAlreadyUnlocked, errManagerLocked,
		errTickSpacingTooLarge, errTickSpacingTooSmall, errCurrenciesOutOfOrderOrEqual,
		errUnauthorizedDynamicLPFeeUpdate, errSwapAmountCannotBeZero, errNonzeroNativeValue,
		errMustClearExactPositiveDelta, errSafeCastOverflow, errPriceLimitAlreadyExceeded,
		errPriceLimitOutOfBounds, errInvalidFeeForExactOut, errInvalidHookResponse, errHookCallFailed,
		errHookDeltaExceedsSwapAmount, errLPFeeTooLarge, errWrappedError, errInvalidPriceOrLiquidity,
		errInvalidPrice, errPriceOverflow, errSqrtPriceMathNotEnoughLiquidity,
	}
	seen := make(map[abi.Selector]string)
	for _, e := range all {
		sel := e.Selector()
		prev, dup := seen[sel]
		require.False(t, dup, "selector %x shared by %s and %s", sel, prev, e.Name)
		seen[sel] = e.Name
	}
}

func TestEncodeQuoteExactInputSingleMatchesGoldenCalldata(t *testing.T) {
	poolKey := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(-60), testHooks)
	params := NewQuoteExactSingleParams(poolKey, true, big.NewInt(1000000), []byte{0xde, 0xad})

	data, err := EncodeQuoteExactInputSingle(params)
	require.NoError(t, err)

	want := "0xaa9d21cb" +
		"0000000000000000000000000000000000000000000000000000000000000020" +
		"000000000000000000000000ba12222222228d8ba445958a75a0704d566bf2c8" +
		"0000000000000000000000001231deb6f5749ef6ce6943a275a1d3e7486f4eae" +
		"0000000000000000000000000000000000000000000000000000000000000bb8" +
		"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc4" +
		"000000000000000000000000b47e3cd837ddf8e4c57f05d70ab865de6e193bbb" +
		"0000000000000000000000000000000000000000000000000000000000000001" +
		"00000000000000000000000000000000000000000000000000000000000f4240" +
		"0000000000000000000000000000000000000000000000000000000000000100" +
		"0000000000000000000000000000000000000000000000000000000000000002" +
		"dead000000000000000000000000000000000000000000000000000000000000"
	require.Equal(t, want, encoding.Hex.EncodePrefixed(data))
}

func TestEncodeQuoteExactOutputSingleMatchesGoldenCalldata(t *testing.T) {
	maxUint128 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	poolKey := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(500), big.NewInt(10), testHooks)
	params := NewQuoteExactSingleParams(poolKey, false, maxUint128, []byte{})

	data, err := EncodeQuoteExactOutputSingle(params)
	require.NoError(t, err)

	want := "0x58733073" +
		"0000000000000000000000000000000000000000000000000000000000000020" +
		"000000000000000000000000ba12222222228d8ba445958a75a0704d566bf2c8" +
		"0000000000000000000000001231deb6f5749ef6ce6943a275a1d3e7486f4eae" +
		"00000000000000000000000000000000000000000000000000000000000001f4" +
		"000000000000000000000000000000000000000000000000000000000000000a" +
		"000000000000000000000000b47e3cd837ddf8e4c57f05d70ab865de6e193bbb" +
		"0000000000000000000000000000000000000000000000000000000000000000" +
		"00000000000000000000000000000000ffffffffffffffffffffffffffffffff" +
		"0000000000000000000000000000000000000000000000000000000000000100" +
		"0000000000000000000000000000000000000000000000000000000000000000"
	require.Equal(t, want, encoding.Hex.EncodePrefixed(data))
}

func TestEncodeQuoteExactInputMatchesGoldenCalldata(t *testing.T) {
	zero := types.MustNewAddressFromHex("0x0000000000000000000000000000000000000000")
	path := []*PathKey{
		NewPathKey(testCurrency0, big.NewInt(3000), big.NewInt(60), testHooks, []byte{0x01}),
		NewPathKey(testCurrency1, big.NewInt(100), big.NewInt(-1), zero, []byte{}),
	}
	params := NewQuoteExactParams(testSender, path, big.NewInt(12345))

	data, err := EncodeQuoteExactInput(params)
	require.NoError(t, err)

	want := "0xca253dc9" +
		"0000000000000000000000000000000000000000000000000000000000000020" +
		"0000000000000000000000006667c8dc9fbfec411e7c1ee2b24de960149f930f" +
		"0000000000000000000000000000000000000000000000000000000000000060" +
		"0000000000000000000000000000000000000000000000000000000000003039" +
		"0000000000000000000000000000000000000000000000000000000000000002" +
		"0000000000000000000000000000000000000000000000000000000000000040" +
		"0000000000000000000000000000000000000000000000000000000000000120" +
		"000000000000000000000000ba12222222228d8ba445958a75a0704d566bf2c8" +
		"0000000000000000000000000000000000000000000000000000000000000bb8" +
		"000000000000000000000000000000000000000000000000000000000000003c" +
		"000000000000000000000000b47e3cd837ddf8e4c57f05d70ab865de6e193bbb" +
		"00000000000000000000000000000000000000000000000000000000000000a0" +
		"0000000000000000000000000000000000000000000000000000000000000001" +
		"0100000000000000000000000000000000000000000000000000000000000000" +
		"0000000000000000000000001231deb6f5749ef6ce6943a275a1d3e7486f4eae" +
		"0000000000000000000000000000000000000000000000000000000000000064" +
		"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff" +
		"0000000000000000000000000000000000000000000000000000000000000000" +
		"00000000000000000000000000000000000000000000000000000000000000a0" +
		"0000000000000000000000000000000000000000000000000000000000000000"
	require.Equal(t, want, encoding.Hex.EncodePrefixed(data))
}

func TestEncodeQuoteExactOutputEmptyPathMatchesGoldenCalldata(t *testing.T) {
	params := NewQuoteExactParams(testSender, []*PathKey{}, big.NewInt(0))

	data, err := EncodeQuoteExactOutput(params)
	require.NoError(t, err)

	want := "0x147d2af9" +
		"0000000000000000000000000000000000000000000000000000000000000020" +
		"0000000000000000000000006667c8dc9fbfec411e7c1ee2b24de960149f930f" +
		"0000000000000000000000000000000000000000000000000000000000000060" +
		"0000000000000000000000000000000000000000000000000000000000000000" +
		"0000000000000000000000000000000000000000000000000000000000000000"
	require.Equal(t, want, encoding.Hex.EncodePrefixed(data))
}

func TestEncodeQuoteExactOutputNilPathEqualsEmptyPath(t *testing.T) {
	nilPath, err := EncodeQuoteExactOutput(NewQuoteExactParams(testSender, nil, big.NewInt(7)))
	require.NoError(t, err)
	emptyPath, err := EncodeQuoteExactOutput(NewQuoteExactParams(testSender, []*PathKey{}, big.NewInt(7)))
	require.NoError(t, err)
	require.Equal(t, emptyPath, nilPath)
}

func TestEncodeQuoteExactSingleNilHookDataEqualsEmpty(t *testing.T) {
	poolKey := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	nilData, err := EncodeQuoteExactInputSingle(NewQuoteExactSingleParams(poolKey, true, big.NewInt(1), nil))
	require.NoError(t, err)
	emptyData, err := EncodeQuoteExactInputSingle(NewQuoteExactSingleParams(poolKey, true, big.NewInt(1), []byte{}))
	require.NoError(t, err)
	require.Equal(t, emptyData, nilData)
}

func TestEncodeQuoteExactInputSingleAndOutputSingleDifferOnlyInSelector(t *testing.T) {
	poolKey := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	params := NewQuoteExactSingleParams(poolKey, true, big.NewInt(1000), []byte{0x01, 0x02, 0x03})

	in, err := EncodeQuoteExactInputSingle(params)
	require.NoError(t, err)
	out, err := EncodeQuoteExactOutputSingle(params)
	require.NoError(t, err)
	require.Len(t, out, len(in))
	require.NotEqual(t, in[:4], out[:4])
	require.Equal(t, in[4:], out[4:])
}

func TestEncodeQuoteExactInputAndOutputDifferOnlyInSelector(t *testing.T) {
	path := []*PathKey{NewPathKey(testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks, []byte{0xaa})}
	params := NewQuoteExactParams(testCurrency0, path, big.NewInt(1000))

	in, err := EncodeQuoteExactInput(params)
	require.NoError(t, err)
	out, err := EncodeQuoteExactOutput(params)
	require.NoError(t, err)
	require.Len(t, out, len(in))
	require.NotEqual(t, in[:4], out[:4])
	require.Equal(t, in[4:], out[4:])
}

func TestEncodeQuoteExactInputSingleRoundTripViaDecodeCall(t *testing.T) {
	poolKey := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(8388607), big.NewInt(-8388608), testHooks)
	maxUint128 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	hookData := make([]byte, 33)
	for i := range hookData {
		hookData[i] = byte(i + 1)
	}
	params := NewQuoteExactSingleParams(poolKey, true, maxUint128, hookData)

	for _, c := range []struct {
		name   string
		fn     *abi.Function
		encode func(*QuoteExactSingleParams) ([]byte, error)
	}{
		{"input", quoteExactInputSingleFn, EncodeQuoteExactInputSingle},
		{"output", quoteExactOutputSingleFn, EncodeQuoteExactOutputSingle},
	} {
		t.Run(c.name, func(t *testing.T) {
			data, err := c.encode(params)
			require.NoError(t, err)
			require.Zero(t, (len(data)-4)%32)

			vals, err := c.fn.DecodeCall(data)
			require.NoError(t, err)
			require.Len(t, vals, 1)
			fields, ok := vals[0].([]any)
			require.True(t, ok)
			require.Len(t, fields, 4)
			poolKeyFields, ok := fields[0].([]any)
			require.True(t, ok)
			require.Len(t, poolKeyFields, 5)
			require.True(t, poolKeyFields[0].(*types.Address).Equal(testCurrency0))
			require.True(t, poolKeyFields[1].(*types.Address).Equal(testCurrency1))
			require.Zero(t, big.NewInt(8388607).Cmp(poolKeyFields[2].(*big.Int)))
			require.Zero(t, big.NewInt(-8388608).Cmp(poolKeyFields[3].(*big.Int)))
			require.True(t, poolKeyFields[4].(*types.Address).Equal(testHooks))
			require.Equal(t, true, fields[1])
			require.Zero(t, maxUint128.Cmp(fields[2].(*big.Int)))
			require.Equal(t, hookData, fields[3])
		})
	}
}

func TestEncodeQuoteExactInputRoundTripViaDecodeCall(t *testing.T) {
	path := make([]*PathKey, 5)
	for i := range path {
		currency := types.NewAddressFromBytes([]byte{byte(i + 1), 0xaa, 0xbb})
		hooks := types.NewAddressFromBytes([]byte{0xcc, byte(i)})
		hookData := make([]byte, i*17)
		for j := range hookData {
			hookData[j] = byte(i ^ j)
		}
		path[i] = NewPathKey(currency, big.NewInt(int64(100*(i+1))), big.NewInt(int64(-i*7)), hooks, hookData)
	}
	params := NewQuoteExactParams(testCurrency0, path, big.NewInt(987654321))

	for _, c := range []struct {
		name   string
		fn     *abi.Function
		encode func(*QuoteExactParams) ([]byte, error)
	}{
		{"input", quoteExactInputFn, EncodeQuoteExactInput},
		{"output", quoteExactOutputFn, EncodeQuoteExactOutput},
	} {
		t.Run(c.name, func(t *testing.T) {
			data, err := c.encode(params)
			require.NoError(t, err)
			require.Zero(t, (len(data)-4)%32)

			vals, err := c.fn.DecodeCall(data)
			require.NoError(t, err)
			require.Len(t, vals, 1)
			fields, ok := vals[0].([]any)
			require.True(t, ok)
			require.Len(t, fields, 3)
			require.True(t, fields[0].(*types.Address).Equal(testCurrency0))
			pathVals, ok := fields[1].([]any)
			require.True(t, ok)
			require.Len(t, pathVals, len(path))
			for i, v := range pathVals {
				k, ok := v.([]any)
				require.True(t, ok)
				require.Len(t, k, 5)
				require.True(t, k[0].(*types.Address).Equal(path[i].IntermediateCurrency), "path[%d].intermediateCurrency", i)
				require.Zero(t, path[i].Fee.Cmp(k[1].(*big.Int)), "path[%d].fee", i)
				require.Zero(t, path[i].TickSpacing.Cmp(k[2].(*big.Int)), "path[%d].tickSpacing", i)
				require.True(t, k[3].(*types.Address).Equal(path[i].Hooks), "path[%d].hooks", i)
				require.Equal(t, path[i].HookData, k[4], "path[%d].hookData", i)
			}
			require.Zero(t, big.NewInt(987654321).Cmp(fields[2].(*big.Int)))
		})
	}
}

func TestEncodeQuoteExactSingleNilParams(t *testing.T) {
	_, err := EncodeQuoteExactInputSingle(nil)
	require.ErrorContains(t, err, "params must not be nil")
	_, err = EncodeQuoteExactOutputSingle(nil)
	require.ErrorContains(t, err, "params must not be nil")
}

func TestEncodeQuoteExactSingleNilPoolKey(t *testing.T) {
	params := NewQuoteExactSingleParams(nil, true, big.NewInt(1), nil)
	_, err := EncodeQuoteExactInputSingle(params)
	require.ErrorContains(t, err, "poolKey must not be nil")
	_, err = EncodeQuoteExactOutputSingle(params)
	require.ErrorContains(t, err, "poolKey must not be nil")
}

func TestEncodeQuoteExactNilParams(t *testing.T) {
	_, err := EncodeQuoteExactInput(nil)
	require.ErrorContains(t, err, "params must not be nil")
	_, err = EncodeQuoteExactOutput(nil)
	require.ErrorContains(t, err, "params must not be nil")
}

func TestEncodeQuoteExactNilPathEntry(t *testing.T) {
	path := []*PathKey{
		NewPathKey(testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks, nil),
		nil,
	}
	params := NewQuoteExactParams(testCurrency0, path, big.NewInt(1))
	_, err := EncodeQuoteExactInput(params)
	require.ErrorContains(t, err, "path[1] must not be nil")
	_, err = EncodeQuoteExactOutput(params)
	require.ErrorContains(t, err, "path[1] must not be nil")
}

func TestEncodeQuoteExactSingleRejectsOutOfRangeValues(t *testing.T) {
	twoPow128 := new(big.Int).Lsh(big.NewInt(1), 128)
	cases := []struct {
		name   string
		params *QuoteExactSingleParams
	}{
		{"exactAmountOverflowsUint128", NewQuoteExactSingleParams(NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks), true, twoPow128, nil)},
		{"exactAmountNegative", NewQuoteExactSingleParams(NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks), true, big.NewInt(-1), nil)},
		{"feeOverflowsUint24", NewQuoteExactSingleParams(NewPoolKey(testCurrency0, testCurrency1, big.NewInt(1<<24), big.NewInt(60), testHooks), true, big.NewInt(1), nil)},
		{"tickSpacingOverflowsInt24", NewQuoteExactSingleParams(NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(1<<23), testHooks), true, big.NewInt(1), nil)},
		{"tickSpacingUnderflowsInt24", NewQuoteExactSingleParams(NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(-(1<<23)-1), testHooks), true, big.NewInt(1), nil)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := EncodeQuoteExactInputSingle(c.params)
			require.Error(t, err)
			_, err = EncodeQuoteExactOutputSingle(c.params)
			require.Error(t, err)
		})
	}
}

func TestEncodeQuoteExactRejectsOutOfRangeValues(t *testing.T) {
	twoPow128 := new(big.Int).Lsh(big.NewInt(1), 128)
	cases := []struct {
		name   string
		params *QuoteExactParams
	}{
		{"exactAmountOverflowsUint128", NewQuoteExactParams(testCurrency0, nil, twoPow128)},
		{"exactAmountNegative", NewQuoteExactParams(testCurrency0, nil, big.NewInt(-1))},
		{"pathFeeOverflowsUint24", NewQuoteExactParams(testCurrency0, []*PathKey{NewPathKey(testCurrency1, big.NewInt(1<<24), big.NewInt(60), testHooks, nil)}, big.NewInt(1))},
		{"pathTickSpacingOverflowsInt24", NewQuoteExactParams(testCurrency0, []*PathKey{NewPathKey(testCurrency1, big.NewInt(3000), big.NewInt(1<<23), testHooks, nil)}, big.NewInt(1))},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := EncodeQuoteExactInput(c.params)
			require.Error(t, err)
			_, err = EncodeQuoteExactOutput(c.params)
			require.Error(t, err)
		})
	}
}

func TestDecodeQuoteResultsGolden(t *testing.T) {
	data, err := encoding.Hex.Decode("0x" +
		"00000000000000000000000000000000000000000000000000000000000003e5" +
		"0000000000000000000000000000000000000000000000000000000000014c08")
	require.NoError(t, err)

	in, err := DecodeQuoteExactInputSingle(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(997).Cmp(in.AmountOut))
	require.Zero(t, big.NewInt(85000).Cmp(in.GasEstimate))

	inMulti, err := DecodeQuoteExactInput(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(997).Cmp(inMulti.AmountOut))
	require.Zero(t, big.NewInt(85000).Cmp(inMulti.GasEstimate))

	out, err := DecodeQuoteExactOutputSingle(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(997).Cmp(out.AmountIn))
	require.Zero(t, big.NewInt(85000).Cmp(out.GasEstimate))

	outMulti, err := DecodeQuoteExactOutput(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(997).Cmp(outMulti.AmountIn))
	require.Zero(t, big.NewInt(85000).Cmp(outMulti.GasEstimate))
}

func TestDecodeQuoteResultsMaxUint256(t *testing.T) {
	maxUint256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	data, err := quoteExactInputSingleFn.EncodeReturn(maxUint256, maxUint256)
	require.NoError(t, err)

	in, err := DecodeQuoteExactInputSingle(data)
	require.NoError(t, err)
	require.Zero(t, maxUint256.Cmp(in.AmountOut))
	require.Zero(t, maxUint256.Cmp(in.GasEstimate))

	inMulti, err := DecodeQuoteExactInput(data)
	require.NoError(t, err)
	require.Zero(t, maxUint256.Cmp(inMulti.AmountOut))

	out, err := DecodeQuoteExactOutputSingle(data)
	require.NoError(t, err)
	require.Zero(t, maxUint256.Cmp(out.AmountIn))

	outMulti, err := DecodeQuoteExactOutput(data)
	require.NoError(t, err)
	require.Zero(t, maxUint256.Cmp(outMulti.AmountIn))
}

func TestDecodeQuoteResultsZero(t *testing.T) {
	data := make([]byte, 64)

	in, err := DecodeQuoteExactInputSingle(data)
	require.NoError(t, err)
	require.Zero(t, in.AmountOut.Sign())
	require.Zero(t, in.GasEstimate.Sign())

	out, err := DecodeQuoteExactOutput(data)
	require.NoError(t, err)
	require.Zero(t, out.AmountIn.Sign())
	require.Zero(t, out.GasEstimate.Sign())
}

func TestDecodeQuoteResultsRejectShortData(t *testing.T) {
	for _, n := range []int{0, 1, 31, 32, 33, 63} {
		data := make([]byte, n)
		_, err := DecodeQuoteExactInputSingle(data)
		require.Error(t, err, "len %d", n)
		_, err = DecodeQuoteExactInput(data)
		require.Error(t, err, "len %d", n)
		_, err = DecodeQuoteExactOutputSingle(data)
		require.Error(t, err, "len %d", n)
		_, err = DecodeQuoteExactOutput(data)
		require.Error(t, err, "len %d", n)
	}
}

func TestDecodeQuoteResultsErrorPrefixes(t *testing.T) {
	_, err := DecodeQuoteExactInputSingle(nil)
	require.ErrorContains(t, err, "uniswapv4: decodeQuoteExactInputSingle")
	_, err = DecodeQuoteExactInput(nil)
	require.ErrorContains(t, err, "uniswapv4: decodeQuoteExactInput")
	_, err = DecodeQuoteExactOutputSingle(nil)
	require.ErrorContains(t, err, "uniswapv4: decodeQuoteExactOutputSingle")
	_, err = DecodeQuoteExactOutput(nil)
	require.ErrorContains(t, err, "uniswapv4: decodeQuoteExactOutput")
}

func TestDecodeQuoteResultsReturnFreshValues(t *testing.T) {
	data, err := quoteExactInputSingleFn.EncodeReturn(big.NewInt(1), big.NewInt(2))
	require.NoError(t, err)

	a, err := DecodeQuoteExactInputSingle(data)
	require.NoError(t, err)
	a.AmountOut.SetInt64(999)
	b, err := DecodeQuoteExactInputSingle(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1).Cmp(b.AmountOut))
}

func TestEncodeQuoterPoolManager(t *testing.T) {
	require.Equal(t, "0xdc4c90d3", encoding.Hex.EncodePrefixed(EncodeQuoterPoolManager()))
}

func TestEncodeQuoterMsgSender(t *testing.T) {
	require.Equal(t, "0xd737d0c7", encoding.Hex.EncodePrefixed(EncodeQuoterMsgSender()))
}

func TestEncodeQuoterGettersReturnFreshSlices(t *testing.T) {
	a := EncodeQuoterPoolManager()
	a[0] = 0x00
	require.Equal(t, "0xdc4c90d3", encoding.Hex.EncodePrefixed(EncodeQuoterPoolManager()))

	b := EncodeQuoterMsgSender()
	b[0] = 0x00
	require.Equal(t, "0xd737d0c7", encoding.Hex.EncodePrefixed(EncodeQuoterMsgSender()))
}

func TestDecodeQuoterPoolManager(t *testing.T) {
	data, err := encoding.Hex.Decode("0x000000000000000000000000fb3e0c6f74eb1a21cc1da29aec80d2dfe6c9a317")
	require.NoError(t, err)
	addr, err := DecodeQuoterPoolManager(data)
	require.NoError(t, err)
	require.True(t, addr.Equal(types.MustNewAddressFromHex("0xFB3e0C6F74eB1a21CC1Da29aeC80D2Dfe6C9a317")))

	_, err = DecodeQuoterPoolManager(data[:31])
	require.ErrorContains(t, err, "uniswapv4: decodeQuoterPoolManager")
}

func TestDecodeQuoterMsgSender(t *testing.T) {
	addr, err := DecodeQuoterMsgSender(make([]byte, 32))
	require.NoError(t, err)
	require.True(t, addr.Equal(types.MustNewAddressFromHex("0x0000000000000000000000000000000000000000")))

	data, err := encoding.Hex.Decode("0x0000000000000000000000006667c8dc9fbfec411e7c1ee2b24de960149f930f")
	require.NoError(t, err)
	addr, err = DecodeQuoterMsgSender(data)
	require.NoError(t, err)
	require.True(t, addr.Equal(testSender))

	_, err = DecodeQuoterMsgSender(nil)
	require.ErrorContains(t, err, "uniswapv4: decodeQuoterMsgSender")
}

func TestQuoterWrapRevertErrorNonRPCErrorPassesThrough(t *testing.T) {
	q := &Quoter{}
	plain := errors.New("boom")
	require.Same(t, plain, q.wrapRevertError(plain))
}

func TestQuoterWrapRevertErrorZeroLengthDataPassesThrough(t *testing.T) {
	q := &Quoter{}
	rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted"}
	require.Same(t, rpcErr, q.wrapRevertError(rpcErr))
}

func TestQuoterWrapRevertErrorMalformedDataPassesThrough(t *testing.T) {
	q := &Quoter{}
	cases := []struct {
		name string
		data string
	}{
		{"notJSONString", `123`},
		{"notHex", `"0xzz"`},
		{"emptyHex", `"0x"`},
		{"shorterThanSelector", `"0x6190b2"`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(c.data)}
			require.Same(t, rpcErr, q.wrapRevertError(rpcErr))
		})
	}
}

func TestQuoterWrapRevertErrorUnknownSelectorMapsToErrUnknown(t *testing.T) {
	q := &Quoter{}
	rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"0xdeadbeef"`)}
	err := q.wrapRevertError(rpcErr)
	require.ErrorIs(t, err, ErrUnknown)
	require.NotErrorIs(t, err, ErrUnexpectedRevertBytes)
	var got *rpc.ResponseError
	require.ErrorAs(t, err, &got)
	require.Same(t, rpcErr, got)
}

func TestQuoterWrapRevertErrorDirectSelectorsMapToSentinel(t *testing.T) {
	var poolID [32]byte
	poolID[0] = 0xab
	poolID[31] = 0xcd
	cases := []struct {
		name string
		err  *abi.Error
		args []any
		want error
	}{
		{"NotEnoughLiquidity", errNotEnoughLiquidity, []any{poolID[:]}, ErrNotEnoughLiquidity},
		{"NotSelf", errNotSelf, nil, ErrNotSelf},
		{"UnexpectedCallSuccess", errUnexpectedCallSuccess, nil, ErrUnexpectedCallSuccess},
	}

	q := &Quoter{}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sel := c.err.Selector()
			argsData, err := abi.Pack(c.err.Inputs, c.args...)
			require.NoError(t, err)
			revertData := append(sel[:], argsData...)
			rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(revertData) + `"`)}

			wrapped := q.wrapRevertError(rpcErr)
			require.ErrorIs(t, wrapped, c.want)
			require.NotErrorIs(t, wrapped, ErrUnknown)
			require.NotErrorIs(t, wrapped, ErrUnexpectedRevertBytes)
			var got *rpc.ResponseError
			require.ErrorAs(t, wrapped, &got)
			require.Same(t, rpcErr, got)
		})
	}
}

func TestQuoterWrapRevertErrorUnexpectedRevertBytesMapsInnerSelector(t *testing.T) {
	var poolID [32]byte
	poolID[0] = 0x12
	cases := []struct {
		name string
		err  *abi.Error
		args []any
		want error
	}{
		{"NotEnoughLiquidity", errNotEnoughLiquidity, []any{poolID[:]}, ErrNotEnoughLiquidity},
		{"NotSelf", errNotSelf, nil, ErrNotSelf},
		{"UnexpectedCallSuccess", errUnexpectedCallSuccess, nil, ErrUnexpectedCallSuccess},
		{"CurrencyNotSettled", errCurrencyNotSettled, nil, ErrCurrencyNotSettled},
		{"PoolNotInitialized", errPoolNotInitialized, nil, ErrPoolNotInitialized},
		{"AlreadyUnlocked", errAlreadyUnlocked, nil, ErrAlreadyUnlocked},
		{"ManagerLocked", errManagerLocked, nil, ErrManagerLocked},
		{"TickSpacingTooLarge", errTickSpacingTooLarge, []any{big.NewInt(32768)}, ErrTickSpacingTooLarge},
		{"TickSpacingTooSmall", errTickSpacingTooSmall, []any{big.NewInt(0)}, ErrTickSpacingTooSmall},
		{"CurrenciesOutOfOrderOrEqual", errCurrenciesOutOfOrderOrEqual, []any{testCurrency1, testCurrency0}, ErrCurrenciesOutOfOrderOrEqual},
		{"UnauthorizedDynamicLPFeeUpdate", errUnauthorizedDynamicLPFeeUpdate, nil, ErrUnauthorizedDynamicLPFeeUpdate},
		{"SwapAmountCannotBeZero", errSwapAmountCannotBeZero, nil, ErrSwapAmountCannotBeZero},
		{"NonzeroNativeValue", errNonzeroNativeValue, nil, ErrNonzeroNativeValue},
		{"MustClearExactPositiveDelta", errMustClearExactPositiveDelta, nil, ErrMustClearExactPositiveDelta},
		{"SafeCastOverflow", errSafeCastOverflow, nil, ErrSafeCastOverflow},
		{"PriceLimitAlreadyExceeded", errPriceLimitAlreadyExceeded, []any{big.NewInt(4295128740), big.NewInt(4295128739)}, ErrPriceLimitAlreadyExceeded},
		{"PriceLimitOutOfBounds", errPriceLimitOutOfBounds, []any{big.NewInt(4295128739)}, ErrPriceLimitOutOfBounds},
		{"InvalidFeeForExactOut", errInvalidFeeForExactOut, nil, ErrInvalidFeeForExactOut},
		{"InvalidHookResponse", errInvalidHookResponse, nil, ErrInvalidHookResponse},
		{"HookDeltaExceedsSwapAmount", errHookDeltaExceedsSwapAmount, nil, ErrHookDeltaExceedsSwapAmount},
		{"LPFeeTooLarge", errLPFeeTooLarge, []any{big.NewInt(1000001)}, ErrLPFeeTooLarge},
		{"InvalidPriceOrLiquidity", errInvalidPriceOrLiquidity, nil, ErrInvalidPriceOrLiquidity},
		{"InvalidPrice", errInvalidPrice, nil, ErrInvalidPrice},
		{"PriceOverflow", errPriceOverflow, nil, ErrPriceOverflow},
		{"SqrtPriceMathNotEnoughLiquidity", errSqrtPriceMathNotEnoughLiquidity, nil, ErrNotEnoughLiquidity},
		{"WrappedErrorOtherContext", errWrappedError, []any{testHooks, []byte{0x57, 0x5e, 0x24, 0xb4}, []byte{0xde, 0xad}, []byte{0x01, 0x02, 0x03, 0x04}}, ErrWrappedError},
	}

	q := &Quoter{}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			innerSel := c.err.Selector()
			innerArgs, err := abi.Pack(c.err.Inputs, c.args...)
			require.NoError(t, err)
			inner := append(innerSel[:], innerArgs...)

			outerSel := errUnexpectedRevertBytes.Selector()
			outerArgs, err := abi.Pack(errUnexpectedRevertBytes.Inputs, inner)
			require.NoError(t, err)
			revertData := append(outerSel[:], outerArgs...)
			rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(revertData) + `"`)}

			wrapped := q.wrapRevertError(rpcErr)
			require.ErrorIs(t, wrapped, ErrUnexpectedRevertBytes)
			require.ErrorIs(t, wrapped, c.want)
			require.NotErrorIs(t, wrapped, ErrUnknown)
			var got *rpc.ResponseError
			require.ErrorAs(t, wrapped, &got)
			require.Same(t, rpcErr, got)
		})
	}
}

func TestQuoterWrapRevertErrorUnexpectedRevertBytesInnerMapsOnlyItsOwnSentinel(t *testing.T) {
	inner := errPoolNotInitialized.Selector()
	outerSel := errUnexpectedRevertBytes.Selector()
	outerArgs, err := abi.Pack(errUnexpectedRevertBytes.Inputs, inner[:])
	require.NoError(t, err)
	revertData := append(outerSel[:], outerArgs...)
	rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(revertData) + `"`)}

	wrapped := (&Quoter{}).wrapRevertError(rpcErr)
	require.ErrorIs(t, wrapped, ErrPoolNotInitialized)
	for _, other := range []error{
		ErrNotEnoughLiquidity, ErrNotSelf, ErrUnexpectedCallSuccess, ErrCurrencyNotSettled,
		ErrAlreadyUnlocked, ErrManagerLocked, ErrTickSpacingTooLarge, ErrTickSpacingTooSmall,
		ErrCurrenciesOutOfOrderOrEqual, ErrUnauthorizedDynamicLPFeeUpdate, ErrSwapAmountCannotBeZero,
		ErrNonzeroNativeValue, ErrMustClearExactPositiveDelta, ErrSafeCastOverflow, ErrPriceLimitAlreadyExceeded,
		ErrPriceLimitOutOfBounds, ErrInvalidFeeForExactOut, ErrInvalidHookResponse, ErrHookCallFailed,
		ErrHookDeltaExceedsSwapAmount, ErrLPFeeTooLarge, ErrWrappedError, ErrInvalidPriceOrLiquidity,
		ErrInvalidPrice, ErrPriceOverflow, ErrUnknown,
	} {
		require.NotErrorIs(t, wrapped, other)
	}
}

func TestQuoterWrapRevertErrorUnexpectedRevertBytesWithoutKnownInner(t *testing.T) {
	invalidTick, err := encoding.Hex.Decode("0x8b86327a00000000000000000000000000000000000000000000000000000000000d8a1c")
	require.NoError(t, err)
	quoteSwap, err := encoding.Hex.Decode("0xecbd98040000000000000000000000000000000000000000000000000000000000000001")
	require.NoError(t, err)
	cases := []struct {
		name  string
		inner []byte
	}{
		{"emptyInner", []byte{}},
		{"innerShorterThanSelector", []byte{0x48, 0x6a, 0xa3}},
		{"unknownInnerSelector", []byte{0xde, 0xad, 0xbe, 0xef, 0x01}},
		{"poolAlreadyInitialized", []byte{0x79, 0x83, 0xc0, 0x51}},
		{"invalidTick", invalidTick},
		{"innerQuoteSwap", quoteSwap},
	}

	q := &Quoter{}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			outerSel := errUnexpectedRevertBytes.Selector()
			outerArgs, err := abi.Pack(errUnexpectedRevertBytes.Inputs, c.inner)
			require.NoError(t, err)
			revertData := append(outerSel[:], outerArgs...)
			rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(revertData) + `"`)}

			wrapped := q.wrapRevertError(rpcErr)
			require.ErrorIs(t, wrapped, ErrUnexpectedRevertBytes)
			require.NotErrorIs(t, wrapped, ErrUnknown)
			require.NotErrorIs(t, wrapped, ErrPoolNotInitialized)
			require.NotErrorIs(t, wrapped, ErrNotEnoughLiquidity)
		})
	}
}

func TestQuoterWrapRevertErrorUnexpectedRevertBytesMalformedPayload(t *testing.T) {
	cases := []struct {
		name string
		data string
	}{
		{"selectorOnly", "0x6190b2b0"},
		{"truncatedOffset", "0x6190b2b0000000000000000000000000000000000000000000000000000000000000"},
		{"offsetOutOfBounds", "0x6190b2b000000000000000000000000000000000000000000000000000000000000000ff"},
		{"lengthOutOfBounds", "0x6190b2b0" +
			"0000000000000000000000000000000000000000000000000000000000000020" +
			"00000000000000000000000000000000000000000000000000000000000000ff"},
	}

	q := &Quoter{}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + c.data + `"`)}
			wrapped := q.wrapRevertError(rpcErr)
			require.ErrorIs(t, wrapped, ErrUnexpectedRevertBytes)
			require.NotErrorIs(t, wrapped, ErrUnknown)
			var got *rpc.ResponseError
			require.ErrorAs(t, wrapped, &got)
			require.Same(t, rpcErr, got)
		})
	}
}

func TestQuoterWrapRevertErrorDoesNotUnwrapNestedUnexpectedRevertBytesTwice(t *testing.T) {
	poolNotInit := errPoolNotInitialized.Selector()
	outerSel := errUnexpectedRevertBytes.Selector()
	middleArgs, err := abi.Pack(errUnexpectedRevertBytes.Inputs, poolNotInit[:])
	require.NoError(t, err)
	middle := append(outerSel[:], middleArgs...)
	outerArgs, err := abi.Pack(errUnexpectedRevertBytes.Inputs, middle)
	require.NoError(t, err)
	revertData := append(outerSel[:], outerArgs...)
	rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(revertData) + `"`)}

	wrapped := (&Quoter{}).wrapRevertError(rpcErr)
	require.ErrorIs(t, wrapped, ErrUnexpectedRevertBytes)
	require.NotErrorIs(t, wrapped, ErrPoolNotInitialized)
}

func TestQuoterWrapRevertErrorWrappedHookCallFailedGolden(t *testing.T) {
	inner, err := encoding.Hex.Decode("0x90bfb865" +
		"000000000000000000000000b47e3cd837ddf8e4c57f05d70ab865de6e193bbb" +
		"575e24b400000000000000000000000000000000000000000000000000000000" +
		"0000000000000000000000000000000000000000000000000000000000000080" +
		"00000000000000000000000000000000000000000000000000000000000000c0" +
		"0000000000000000000000000000000000000000000000000000000000000004" +
		"deadbeef00000000000000000000000000000000000000000000000000000000" +
		"0000000000000000000000000000000000000000000000000000000000000004" +
		"a9e35b2f00000000000000000000000000000000000000000000000000000000")
	require.NoError(t, err)

	vals, err := errWrappedError.Decode(inner)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testHooks))
	require.Equal(t, []byte{0xde, 0xad, 0xbe, 0xef}, vals[2])
	require.Equal(t, []byte{0xa9, 0xe3, 0x5b, 0x2f}, vals[3])

	outerSel := errUnexpectedRevertBytes.Selector()
	outerArgs, err := abi.Pack(errUnexpectedRevertBytes.Inputs, inner)
	require.NoError(t, err)
	revertData := append(outerSel[:], outerArgs...)
	rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(revertData) + `"`)}

	wrapped := (&Quoter{}).wrapRevertError(rpcErr)
	require.ErrorIs(t, wrapped, ErrUnexpectedRevertBytes)
	require.ErrorIs(t, wrapped, ErrWrappedError)
	require.ErrorIs(t, wrapped, ErrHookCallFailed)
	require.NotErrorIs(t, wrapped, ErrUnknown)
	require.NotErrorIs(t, wrapped, ErrInvalidHookResponse)
	require.NotErrorIs(t, wrapped, ErrNotEnoughLiquidity)
	var got *rpc.ResponseError
	require.ErrorAs(t, wrapped, &got)
	require.Same(t, rpcErr, got)
}

func TestQuoterWrapRevertErrorWrappedErrorWithoutHookCallFailed(t *testing.T) {
	cases := []struct {
		name    string
		details []byte
	}{
		{"emptyDetails", []byte{}},
		{"shortDetails", []byte{0xa9, 0xe3, 0x5b}},
		{"otherDetails", []byte{0x01, 0x02, 0x03, 0x04}},
		{"hookCallFailedNotAtStart", []byte{0x00, 0xa9, 0xe3, 0x5b, 0x2f}},
	}

	q := &Quoter{}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			innerSel := errWrappedError.Selector()
			innerArgs, err := abi.Pack(errWrappedError.Inputs, testHooks, []byte{0x57, 0x5e, 0x24, 0xb4}, []byte{0xde, 0xad}, c.details)
			require.NoError(t, err)
			inner := append(innerSel[:], innerArgs...)

			outerSel := errUnexpectedRevertBytes.Selector()
			outerArgs, err := abi.Pack(errUnexpectedRevertBytes.Inputs, inner)
			require.NoError(t, err)
			revertData := append(outerSel[:], outerArgs...)
			rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(revertData) + `"`)}

			wrapped := q.wrapRevertError(rpcErr)
			require.ErrorIs(t, wrapped, ErrUnexpectedRevertBytes)
			require.ErrorIs(t, wrapped, ErrWrappedError)
			require.NotErrorIs(t, wrapped, ErrHookCallFailed)
			require.NotErrorIs(t, wrapped, ErrUnknown)
		})
	}
}

func TestQuoterWrapRevertErrorMalformedWrappedError(t *testing.T) {
	cases := []struct {
		name  string
		inner string
	}{
		{"selectorOnly", "0x90bfb865"},
		{"truncatedHead", "0x90bfb865000000000000000000000000b47e3cd837ddf8e4c57f05d70ab865de6e193bbb"},
		{"offsetOutOfBounds", "0x90bfb865" +
			"000000000000000000000000b47e3cd837ddf8e4c57f05d70ab865de6e193bbb" +
			"575e24b400000000000000000000000000000000000000000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000000fff" +
			"0000000000000000000000000000000000000000000000000000000000000fff"},
	}

	q := &Quoter{}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			inner, err := encoding.Hex.Decode(c.inner)
			require.NoError(t, err)
			outerSel := errUnexpectedRevertBytes.Selector()
			outerArgs, err := abi.Pack(errUnexpectedRevertBytes.Inputs, inner)
			require.NoError(t, err)
			revertData := append(outerSel[:], outerArgs...)
			rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(revertData) + `"`)}

			wrapped := q.wrapRevertError(rpcErr)
			require.ErrorIs(t, wrapped, ErrUnexpectedRevertBytes)
			require.ErrorIs(t, wrapped, ErrWrappedError)
			require.NotErrorIs(t, wrapped, ErrHookCallFailed)
			require.NotErrorIs(t, wrapped, ErrUnknown)
		})
	}
}
