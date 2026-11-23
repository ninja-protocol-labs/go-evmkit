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
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

func TestRouterCallExecuteActionsSendsCallAndNativeAmount(t *testing.T) {
	router := testHooks
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	swapData, err := EncodeActionSwapExactInSingle(NewExactInputSingleParams(key, true, big.NewInt(1_000), big.NewInt(900), big.NewInt(0), nil))
	require.NoError(t, err)
	settleData, err := EncodeActionSettleAll(testCurrency0, big.NewInt(1_000))
	require.NoError(t, err)
	takeData, err := EncodeActionTakeAll(testCurrency1, big.NewInt(900))
	require.NoError(t, err)
	actions := []byte{ActionSwapExactInSingle, ActionSettleAll, ActionTakeAll}
	params := [][]byte{swapData, settleData, takeData}
	calldata, err := EncodeExecuteActions(actions, params)
	require.NoError(t, err)

	for _, tc := range []struct {
		name   string
		native *big.Int
		value  string
	}{
		{"with native", big.NewInt(1_000_000), "0xf4240"},
		{"without native", nil, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, opts := range []*bind.CallOptions{nil, bind.NewCallOptions().WithFrom(testSender).WithBlock("0x10")} {
				c := mocks.NewClient(t)
				c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
					require.Equal(t, rpc.MethodETHCall, e.Method)
					call := e.Params.([]any)[0].(map[string]any)
					require.Equal(t, router.String(), call["to"])
					require.Equal(t, encoding.Hex.EncodePrefixed(calldata), call["data"])
					if tc.value == "" {
						require.NotContains(t, call, "value")
					} else {
						require.Equal(t, tc.value, call["value"])
					}
					if opts == nil {
						require.NotContains(t, call, "from")
						require.Equal(t, rpc.BlockTagLatest, e.Params.([]any)[1])
					} else {
						require.Equal(t, testSender.String(), call["from"])
						require.Equal(t, "0x10", e.Params.([]any)[1])
					}
					*e.Result.(*string) = "0x"
					return nil
				}).Once()

				require.NoError(t, NewRouter(c).CallExecuteActions(t.Context(), opts, NewExecuteActionsRequest(router, actions, params, tc.native)))
				if opts != nil {
					require.Nil(t, opts.Value)
				}
			}
		})
	}

	err = NewRouter(mocks.NewClient(t)).CallExecuteActions(t.Context(), bind.NewCallOptions().WithValue(big.NewInt(1)), NewExecuteActionsRequest(router, actions, params, nil))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv4: callExecuteActions: invalid options: value must be nil")
}

func TestRouterCallExecuteActionsMapsRevertToSentinel(t *testing.T) {
	router := testHooks

	for _, tc := range []struct {
		name string
		err  *abi.Error
		args []any
		raw  jsontext.Value
		want error
	}{
		{"V4TooLittleReceived", errV4TooLittleReceived, []any{big.NewInt(900), big.NewInt(800)}, nil, ErrV4TooLittleReceived},
		{"V4TooMuchRequested", errV4TooMuchRequested, []any{big.NewInt(1000), big.NewInt(1100)}, nil, ErrV4TooMuchRequested},
		{"V4TooLittleReceivedPerHop", errV4TooLittleReceivedPerHop, []any{big.NewInt(1), big.NewInt(2), big.NewInt(1)}, nil, ErrV4TooLittleReceivedPerHop},
		{"V4TooMuchRequestedPerHop", errV4TooMuchRequestedPerHop, []any{big.NewInt(0), big.NewInt(2), big.NewInt(3)}, nil, ErrV4TooMuchRequestedPerHop},
		{"V4TooLittleReceivedPerHopSingle", errV4TooLittleReceivedPerHopSingle, []any{big.NewInt(2), big.NewInt(1)}, nil, ErrV4TooLittleReceivedPerHopSingle},
		{"V4TooMuchRequestedPerHopSingle", errV4TooMuchRequestedPerHopSingle, []any{big.NewInt(2), big.NewInt(3)}, nil, ErrV4TooMuchRequestedPerHopSingle},
		{"InvalidHopPriceLength", errInvalidHopPriceLength, nil, nil, ErrInvalidHopPriceLength},
		{"V4ExactOutputUnfilled", errV4ExactOutputUnfilled, []any{big.NewInt(500), big.NewInt(400)}, nil, ErrV4ExactOutputUnfilled},
		{"InputLengthMismatch", errInputLengthMismatch, nil, nil, ErrInputLengthMismatch},
		{"UnsupportedAction", errUnsupportedAction, []any{big.NewInt(0x02)}, nil, ErrUnsupportedAction},
		{"DeltaNotPositive", errDeltaNotPositive, []any{testCurrency0}, nil, ErrDeltaNotPositive},
		{"DeltaNotNegative", errDeltaNotNegative, []any{testCurrency1}, nil, ErrDeltaNotNegative},
		{"InsufficientBalance", errInsufficientBalance, nil, nil, ErrInsufficientBalance},
		{"PoolManager CurrencyNotSettled", errCurrencyNotSettled, nil, nil, ErrCurrencyNotSettled},
		{"PoolManager PoolNotInitialized", errPoolNotInitialized, nil, nil, ErrPoolNotInitialized},
		{"PoolManager PriceLimitAlreadyExceeded", errPriceLimitAlreadyExceeded, []any{big.NewInt(2), big.NewInt(3)}, nil, ErrPriceLimitAlreadyExceeded},
		{"PoolManager WrappedError", errWrappedError, []any{testHooks, []byte{0x12, 0x34, 0x56, 0x78}, []byte{}, []byte{0xde, 0xad, 0xbe, 0xef}}, nil, ErrWrappedError},
		{"Error(string)", nil, nil, jsontext.Value(`"0x08c379a00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000c554e415554484f52495a45440000000000000000000000000000000000000000"`), ErrUnknown},
		{"unknown selector", nil, nil, jsontext.Value(`"0xdeadbeef"`), ErrUnknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := tc.raw
			if tc.err != nil {
				sel := tc.err.Selector()
				packed, err := abi.Pack(tc.err.Inputs, tc.args...)
				require.NoError(t, err)
				data = jsontext.Value(`"` + encoding.Hex.EncodePrefixed(append(sel[:], packed...)) + `"`)
			}
			rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: data}
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).Return(rpcErr).Once()

			err := NewRouter(c).CallExecuteActions(t.Context(), nil, NewExecuteActionsRequest(router, []byte{ActionSettleAll}, [][]byte{{0x01}}, nil))
			require.ErrorIs(t, err, tc.want)
			require.ErrorContains(t, err, "uniswapv4: callExecuteActions: ")
			got, ok := errors.AsType[*rpc.ResponseError](err)
			require.True(t, ok)
			require.Same(t, rpcErr, got)
			if !errors.Is(tc.want, ErrUnknown) {
				require.NotErrorIs(t, err, ErrUnknown)
			}
		})
	}
}

func TestRouterWrapRevertErrorPassesThroughUndecodable(t *testing.T) {
	r := &Router{}
	plain := errors.New("dial tcp: connection refused")
	require.Same(t, plain, r.wrapRevertError(plain))
	require.Nil(t, r.wrapRevertError(nil))

	for _, data := range []jsontext.Value{nil, jsontext.Value(`"0x"`), jsontext.Value(`"0x123456"`), jsontext.Value(`"0xzz"`), jsontext.Value(`{"x":1}`), jsontext.Value(`42`)} {
		t.Run(string(data), func(t *testing.T) {
			rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: data}
			require.Same(t, rpcErr, r.wrapRevertError(rpcErr))

			wrapped := fmt.Errorf("rpc: %s: %w", rpc.MethodETHEstimateGas, rpcErr)
			require.Same(t, wrapped, r.wrapRevertError(wrapped))
		})
	}
}

func TestRouterRevertSelectorsAreUnique(t *testing.T) {
	seen := map[abi.Selector]string{}
	for _, e := range []*abi.Error{
		errV4TooLittleReceived, errV4TooMuchRequested, errV4TooLittleReceivedPerHop, errV4TooMuchRequestedPerHop,
		errV4TooLittleReceivedPerHopSingle, errV4TooMuchRequestedPerHopSingle, errInvalidHopPriceLength, errV4ExactOutputUnfilled,
		errInputLengthMismatch, errUnsupportedAction, errDeltaNotPositive, errDeltaNotNegative, errInsufficientBalance,
		errCurrencyNotSettled, errPoolNotInitialized, errAlreadyUnlocked, errManagerLocked, errTickSpacingTooLarge,
		errTickSpacingTooSmall, errCurrenciesOutOfOrderOrEqual, errUnauthorizedDynamicLPFeeUpdate, errSwapAmountCannotBeZero,
		errNonzeroNativeValue, errMustClearExactPositiveDelta, errProtocolFeeTooLarge, errInvalidCaller, errProtocolFeeCurrencySynced,
		errPoolAlreadyInitialized, errTicksMisordered, errTickLowerOutOfBounds, errTickUpperOutOfBounds, errTickLiquidityOverflow,
		errNoLiquidityToReceiveFees, errTickMisaligned, errInvalidTick, errInvalidSqrtPrice, errHookAddressNotValid,
		errCannotUpdateEmptyPosition, errDelegateCallNotAllowed, errNativeTransferFailed, errERC20TransferFailed,
		errSafeCastOverflow, errPriceLimitAlreadyExceeded, errPriceLimitOutOfBounds, errInvalidFeeForExactOut,
		errInvalidHookResponse, errHookCallFailed, errHookDeltaExceedsSwapAmount, errLPFeeTooLarge, errWrappedError,
		errInvalidPriceOrLiquidity, errInvalidPrice, errPriceOverflow, errSqrtPriceMathNotEnoughLiquidity,
	} {
		prev, dup := seen[e.Selector()]
		require.False(t, dup, "%s collides with %s", e.Signature(), prev)
		seen[e.Selector()] = e.Signature()
	}

	for _, tc := range []struct {
		err *abi.Error
		sig string
	}{
		{errV4TooLittleReceived, "V4TooLittleReceived(uint256,uint256)"},
		{errV4TooLittleReceivedPerHop, "V4TooLittleReceivedPerHop(uint256,uint256,uint256)"},
		{errV4ExactOutputUnfilled, "V4ExactOutputUnfilled(uint256,uint256)"},
		{errInputLengthMismatch, "InputLengthMismatch()"},
		{errUnsupportedAction, "UnsupportedAction(uint256)"},
	} {
		require.Equal(t, tc.sig, tc.err.Signature())
		sel := tc.err.Selector()
		require.Equal(t, core.Keccak256([]byte(tc.sig)).Bytes()[:4], sel[:], tc.sig)
	}
}

func TestRouterMethodsRejectInvalidRequestWithoutRPC(t *testing.T) {
	router := testHooks
	r := NewRouter(mocks.NewClient(t))
	ctx := t.Context()
	opts := bind.NewTransactOptions(nil, tx.NewDynamicFeeBuilder())

	err := r.CallExecuteActions(ctx, nil, &ExecuteActionsRequest{})
	require.EqualError(t, err, "uniswapv4: callExecuteActions: invalid request: contract is nil")
	err = r.CallExecuteActions(ctx, nil, NewExecuteActionsRequest(router, nil, nil, nil))
	require.EqualError(t, err, "uniswapv4: callExecuteActions: invalid request: actions is empty")
	err = r.CallExecuteActions(ctx, nil, NewExecuteActionsRequest(router, []byte{ActionSettleAll, ActionTakeAll}, [][]byte{{0x01}}, nil))
	require.EqualError(t, err, "uniswapv4: callExecuteActions: invalid request: params length 1 does not match actions length 2")
	require.ErrorIs(t, err, bind.ErrInvalidRequest)

	_, _, err = r.ExecuteActions(ctx, opts, NewExecuteActionsRequest(router, []byte{ActionSettleAll}, [][]byte{{0x01}}, big.NewInt(-1)))
	require.EqualError(t, err, "uniswapv4: executeActions: invalid request: nativeAmount is negative")
	_, _, err = r.ExecuteActions(ctx, opts, NewExecuteActionsRequest(router, []byte{ActionSettleAll}, [][]byte{{0x01}}, new(big.Int).Lsh(big.NewInt(1), 256)))
	require.EqualError(t, err, "uniswapv4: executeActions: invalid request: nativeAmount exceeds uint256")
	_, _, err = r.ExecuteActions(ctx, opts, &ExecuteActionsRequest{})
	require.EqualError(t, err, "uniswapv4: executeActions: invalid request: contract is nil")

	_, _, err = r.ExecuteActions(ctx, bind.NewTransactOptions(nil, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewExecuteActionsRequest(router, []byte{ActionSettleAll}, [][]byte{{0x01}}, big.NewInt(1)))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv4: executeActions: invalid options: value must be nil")

	_, _, err = r.ExecuteActions(ctx, nil, NewExecuteActionsRequest(router, []byte{ActionSettleAll}, [][]byte{{0x01}}, big.NewInt(1)))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.ErrorContains(t, err, "uniswapv4: executeActions: ")
}

func TestRouterExecuteActionsSignsAndSends(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sender := core.PubkeyToAddress(key.PublicKey())
	router := testHooks
	poolKey := NewPoolKey(types.ZeroAddress(), testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	swapData, err := EncodeActionSwapExactInSingle(NewExactInputSingleParams(poolKey, true, big.NewInt(1_000), big.NewInt(1), big.NewInt(0), nil))
	require.NoError(t, err)
	settleData, err := EncodeActionSettleAll(types.ZeroAddress(), big.NewInt(1_000))
	require.NoError(t, err)
	takeData, err := EncodeActionTakeAll(testCurrency1, big.NewInt(1))
	require.NoError(t, err)
	actions := []byte{ActionSwapExactInSingle, ActionSettleAll, ActionTakeAll}
	params := [][]byte{swapData, settleData, takeData}
	calldata, err := EncodeExecuteActions(actions, params)
	require.NoError(t, err)

	for _, tc := range []struct {
		name   string
		native *big.Int
	}{
		{"with native", big.NewInt(1_000)},
		{"without native", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			builder := tx.NewDynamicFeeBuilder()
			builder.ChainID(big.NewInt(1))
			builder.Nonce(7)
			builder.Limit(6_000_000)
			builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

			var raw []byte
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				require.Equal(t, rpc.MethodETHSendRawTransaction, e.Method)
				var err error
				raw, err = encoding.Hex.Decode(e.Params.([]any)[0].(string))
				require.NoError(t, err)
				*e.Result.(*string) = core.Keccak256(raw).String()
				return nil
			}).Once()

			opts := bind.NewTransactOptions(key, builder)
			signed, hash, err := NewRouter(c).ExecuteActions(t.Context(), opts, NewExecuteActionsRequest(router, actions, params, tc.native))
			require.NoError(t, err)
			require.Nil(t, opts.Value)
			require.True(t, hash.Equal(core.Keccak256(raw)))
			require.Equal(t, uint64(7), signed.AddressNonce())

			decoded, err := core.DecodeTransaction(raw)
			require.NoError(t, err)
			dyn, ok := decoded.(*core.DynamicFeeTx)
			require.True(t, ok)
			require.True(t, dyn.To.Equal(router))
			if tc.native == nil {
				require.Zero(t, dyn.Value.Sign())
			} else {
				require.Zero(t, tc.native.Cmp(dyn.Value))
			}
			require.Equal(t, calldata, dyn.Data)
			got, err := decoded.Sender()
			require.NoError(t, err)
			require.True(t, got.Equal(sender))
		})
	}
}

func TestRouterExecuteActionsMapsEstimateRevertToSentinel(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)

	for _, tc := range []struct {
		name string
		err  *abi.Error
		args []any
		want error
	}{
		{"too little received", errV4TooLittleReceived, []any{big.NewInt(900), big.NewInt(800)}, ErrV4TooLittleReceived},
		{"unsupported action", errUnsupportedAction, []any{big.NewInt(0x00)}, ErrUnsupportedAction},
		{"pool not initialized", errPoolNotInitialized, nil, ErrPoolNotInitialized},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sel := tc.err.Selector()
			packed, err := abi.Pack(tc.err.Inputs, tc.args...)
			require.NoError(t, err)
			data := jsontext.Value(`"` + encoding.Hex.EncodePrefixed(append(sel[:], packed...)) + `"`)

			builder := tx.NewDynamicFeeBuilder()
			builder.ChainID(big.NewInt(1))
			builder.Nonce(7)
			builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

			c := mocks.NewClient(t)
			c.EXPECT().Batch(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, elems rpc.Elements) error {
				require.Equal(t, 1, elems.Len())
				require.Equal(t, rpc.MethodETHEstimateGas, elems.GetMethod(0))
				return fmt.Errorf("rpc: %s: %w", rpc.MethodETHEstimateGas, &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: data})
			}).Once()

			signed, hash, err := NewRouter(c).ExecuteActions(t.Context(), bind.NewTransactOptions(key, builder), NewExecuteActionsRequest(testHooks, []byte{ActionSettleAll}, [][]byte{{0x01}}, nil))
			require.ErrorIs(t, err, tc.want)
			require.ErrorContains(t, err, "uniswapv4: executeActions: ")
			require.NotErrorIs(t, err, ErrUnknown)
			require.Nil(t, signed)
			require.Nil(t, hash)
		})
	}
}

func TestRouterExecuteActionsReturnsSignedTxOnBroadcastFailure(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(9)
	builder.Limit(6_000_000)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(errors.New("dial tcp: connection refused")).Once()

	signed, hash, err := NewRouter(c).ExecuteActions(t.Context(), bind.NewTransactOptions(key, builder), NewExecuteActionsRequest(testHooks, []byte{ActionSettleAll}, [][]byte{{0x01}}, big.NewInt(5)))
	require.ErrorContains(t, err, "uniswapv4: executeActions: ")
	require.ErrorContains(t, err, "connection refused")
	require.NotErrorIs(t, err, ErrUnknown)
	require.Nil(t, hash)
	require.NotNil(t, signed)
	require.Equal(t, uint64(9), signed.AddressNonce())
	require.Zero(t, big.NewInt(5).Cmp(signed.(*core.DynamicFeeTx).Value))
}

func TestRouterCallPassesThroughTransportError(t *testing.T) {
	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(errors.New("dial tcp: connection refused")).Once()

	err := NewRouter(c).CallExecuteActions(t.Context(), nil, NewExecuteActionsRequest(testHooks, []byte{ActionSettleAll}, [][]byte{{0x01}}, nil))
	require.EqualError(t, err, "uniswapv4: callExecuteActions: dial tcp: connection refused")
	require.NotErrorIs(t, err, ErrUnknown)
}

func TestRouterCallsPropagateContext(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(t.Context(), marker{}, "router")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(got context.Context, e rpc.Element) error {
		require.Equal(t, "router", got.Value(marker{}))
		*e.Result.(*string) = "0x"
		return nil
	}).Twice()

	var r IRouter = NewRouter(c)
	require.NoError(t, r.CallExecuteActions(ctx, nil, NewExecuteActionsRequest(testHooks, []byte{ActionSettleAll}, [][]byte{{0x01}}, nil)))

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(1)
	builder.Limit(3_000_000)
	builder.Price(big.NewInt(2), big.NewInt(1))
	_, _, err = r.ExecuteActions(ctx, bind.NewTransactOptions(key, builder), NewExecuteActionsRequest(testHooks, []byte{ActionSettleAll}, [][]byte{{0x01}}, nil))
	require.NoError(t, err)
}
