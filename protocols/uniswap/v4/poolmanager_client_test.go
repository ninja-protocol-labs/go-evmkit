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
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc165"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc6909"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

func TestPoolManagerReadsSendCallAndDecode(t *testing.T) {
	pm := testPoolManager
	from := testSender
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	opts := bind.NewCallOptions().WithFrom(from).WithBlock("0x1b4")
	slotA, slotB := [32]byte{0x0a}, [32]byte{0x0b}
	valueA, valueB := [32]byte{0x01, 0x02}, [32]byte{0xff}
	var salt [32]byte
	salt[31] = 0x07

	feesWord, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(12345))
	require.NoError(t, err)
	controllerWord, err := abi.Pack(abi.NewTypes(abi.Address), from)
	require.NoError(t, err)
	valueWord, err := abi.Pack(abi.NewTypes(abi.Bytes32), valueA[:])
	require.NoError(t, err)
	valuesWord, err := abi.Pack(abi.NewTypes(abi.Slice(abi.Bytes32)), [][]byte{valueA[:], valueB[:]})
	require.NoError(t, err)
	emptyValuesWord, err := abi.Pack(abi.NewTypes(abi.Slice(abi.Bytes32)), [][]byte{})
	require.NoError(t, err)
	tickWord, err := abi.Pack(abi.NewTypes(abi.Int24), big.NewInt(-887272))
	require.NoError(t, err)
	unlockWord, err := abi.Pack(abi.NewTypes(abi.Bytes), []byte{0xca, 0xfe})
	require.NoError(t, err)
	paidWord, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(77))
	require.NoError(t, err)
	collectedWord, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(99))
	require.NoError(t, err)
	modifyWord, err := encoding.Hex.Decode("0xffffffffffffffffffffffffffffff9c000000000000000000000000000000c800000000000000000000000000000003fffffffffffffffffffffffffffffffc")
	require.NoError(t, err)
	swapWord, err := encoding.Hex.Decode("0xfffffffffffffffffffffffffffffc18000000000000000000000000000003e5")
	require.NoError(t, err)
	donateWord, err := encoding.Hex.Decode("0xfffffffffffffffffffffffffffffffbfffffffffffffffffffffffffffffffa")
	require.NoError(t, err)

	protocolFeesData, err := EncodeProtocolFeesAccrued(testCurrency0)
	require.NoError(t, err)
	controllerData, err := EncodeProtocolFeeController()
	require.NoError(t, err)
	extsloadData, err := EncodeExtsload(slotA)
	require.NoError(t, err)
	extsloadRangeData, err := EncodeExtsloadRange(slotA, big.NewInt(2))
	require.NoError(t, err)
	extsloadSlotsData, err := EncodeExtsloadSlots([][32]byte{slotA, slotB})
	require.NoError(t, err)
	extsloadNoSlotsData, err := EncodeExtsloadSlots(nil)
	require.NoError(t, err)
	exttloadData, err := EncodeExttload(slotB)
	require.NoError(t, err)
	exttloadSlotsData, err := EncodeExttloadSlots([][32]byte{slotB, slotA})
	require.NoError(t, err)
	initializeData, err := EncodeInitialize(key, big.NewInt(4295128739))
	require.NoError(t, err)
	liq := NewLiquidityParams(big.NewInt(-60), big.NewInt(60), big.NewInt(1000), salt)
	modifyData, err := EncodeModifyLiquidity(key, liq, []byte{0x01})
	require.NoError(t, err)
	swapParams := NewSwapAmountParams(true, big.NewInt(-1000), big.NewInt(4295128740))
	swapData, err := EncodeSwap(key, swapParams, nil)
	require.NoError(t, err)
	donateData, err := EncodeDonate(key, big.NewInt(5), big.NewInt(6), []byte{0x02})
	require.NoError(t, err)
	unlockData, err := EncodeUnlock([]byte{0xbe, 0xef})
	require.NoError(t, err)
	settleForData, err := EncodeSettleFor(testHooks)
	require.NoError(t, err)
	settleData, err := EncodeSettle()
	require.NoError(t, err)
	collectData, err := EncodeCollectProtocolFees(from, testCurrency1, big.NewInt(0))
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		result   []byte
		call     func(m *PoolManager) (any, error)
		want     any
	}{
		{
			"protocolFeesAccrued",
			protocolFeesData,
			feesWord,
			func(m *PoolManager) (any, error) {
				return m.ProtocolFeesAccrued(t.Context(), opts, NewProtocolFeesAccruedRequest(pm, testCurrency0))
			},
			big.NewInt(12345),
		},
		{
			"protocolFeeController",
			controllerData,
			controllerWord,
			func(m *PoolManager) (any, error) {
				return m.ProtocolFeeController(t.Context(), opts, NewProtocolFeeControllerRequest(pm))
			},
			from,
		},
		{
			"extsload",
			extsloadData,
			valueWord,
			func(m *PoolManager) (any, error) {
				return m.Extsload(t.Context(), opts, NewExtsloadRequest(pm, slotA))
			},
			valueA,
		},
		{
			"extsloadRange",
			extsloadRangeData,
			valuesWord,
			func(m *PoolManager) (any, error) {
				return m.ExtsloadRange(t.Context(), opts, NewExtsloadRangeRequest(pm, slotA, big.NewInt(2)))
			},
			[][32]byte{valueA, valueB},
		},
		{
			"extsloadSlots",
			extsloadSlotsData,
			valuesWord,
			func(m *PoolManager) (any, error) {
				return m.ExtsloadSlots(t.Context(), opts, NewExtsloadSlotsRequest(pm, [][32]byte{slotA, slotB}))
			},
			[][32]byte{valueA, valueB},
		},
		{
			"extsloadSlots empty",
			extsloadNoSlotsData,
			emptyValuesWord,
			func(m *PoolManager) (any, error) {
				return m.ExtsloadSlots(t.Context(), opts, NewExtsloadSlotsRequest(pm, nil))
			},
			[][32]byte{},
		},
		{
			"exttload",
			exttloadData,
			valueWord,
			func(m *PoolManager) (any, error) {
				return m.Exttload(t.Context(), opts, NewExttloadRequest(pm, slotB))
			},
			valueA,
		},
		{
			"exttloadSlots",
			exttloadSlotsData,
			valuesWord,
			func(m *PoolManager) (any, error) {
				return m.ExttloadSlots(t.Context(), opts, NewExttloadSlotsRequest(pm, [][32]byte{slotB, slotA}))
			},
			[][32]byte{valueA, valueB},
		},
		{
			"callInitialize",
			initializeData,
			tickWord,
			func(m *PoolManager) (any, error) {
				return m.CallInitialize(t.Context(), opts, NewInitializeRequest(pm, key, big.NewInt(4295128739)))
			},
			big.NewInt(-887272),
		},
		{
			"callModifyLiquidity",
			modifyData,
			modifyWord,
			func(m *PoolManager) (any, error) {
				caller, fees, err := m.CallModifyLiquidity(t.Context(), opts, NewModifyLiquidityRequest(pm, key, liq, []byte{0x01}))
				if err != nil {
					return nil, err
				}
				return []BalanceDelta{*caller, *fees}, nil
			},
			[]BalanceDelta{*NewBalanceDelta(big.NewInt(-100), big.NewInt(200)), *NewBalanceDelta(big.NewInt(3), big.NewInt(-4))},
		},
		{
			"callSwap",
			swapData,
			swapWord,
			func(m *PoolManager) (any, error) {
				return m.CallSwap(t.Context(), opts, NewSwapRequest(pm, key, swapParams, nil))
			},
			NewBalanceDelta(big.NewInt(-1_000), big.NewInt(997)),
		},
		{
			"callDonate",
			donateData,
			donateWord,
			func(m *PoolManager) (any, error) {
				return m.CallDonate(t.Context(), opts, NewDonateRequest(pm, key, big.NewInt(5), big.NewInt(6), []byte{0x02}))
			},
			NewBalanceDelta(big.NewInt(-5), big.NewInt(-6)),
		},
		{
			"callUnlock",
			unlockData,
			unlockWord,
			func(m *PoolManager) (any, error) {
				return m.CallUnlock(t.Context(), opts, NewUnlockRequest(pm, []byte{0xbe, 0xef}))
			},
			[]byte{0xca, 0xfe},
		},
		{
			"callSettle",
			settleData,
			paidWord,
			func(m *PoolManager) (any, error) {
				return m.CallSettle(t.Context(), opts, NewSettleRequest(pm, nil))
			},
			big.NewInt(77),
		},
		{
			"callSettleFor",
			settleForData,
			paidWord,
			func(m *PoolManager) (any, error) {
				return m.CallSettleFor(t.Context(), opts, NewSettleForRequest(pm, testHooks, nil))
			},
			big.NewInt(77),
		},
		{
			"callCollectProtocolFees",
			collectData,
			collectedWord,
			func(m *PoolManager) (any, error) {
				return m.CallCollectProtocolFees(t.Context(), opts, NewCollectProtocolFeesRequest(pm, from, testCurrency1, big.NewInt(0)))
			},
			big.NewInt(99),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				require.Equal(t, rpc.MethodETHCall, e.Method)
				params, ok := e.Params.([]any)
				require.True(t, ok)
				require.Len(t, params, 2)
				call, ok := params[0].(map[string]any)
				require.True(t, ok)
				require.Equal(t, pm.String(), call["to"])
				require.Equal(t, from.String(), call["from"])
				require.Equal(t, encoding.Hex.EncodePrefixed(tc.calldata), call["data"])
				require.NotContains(t, call, "value")
				require.Equal(t, "0x1b4", params[1])
				*e.Result.(*string) = encoding.Hex.EncodePrefixed(tc.result)
				return nil
			}).Once()

			got, err := tc.call(New(c))
			require.NoError(t, err)
			require.Equal(t, fmt.Sprint(tc.want), fmt.Sprint(got))
		})
	}
}

func TestPoolManagerCallSettleSendsNativeAmountAsValue(t *testing.T) {
	pm := testPoolManager
	paidWord, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(1_000_000))
	require.NoError(t, err)

	for _, tc := range []struct {
		name string
		call func(m *PoolManager, opts *bind.CallOptions) (*big.Int, error)
	}{
		{"callSettle", func(m *PoolManager, opts *bind.CallOptions) (*big.Int, error) {
			return m.CallSettle(t.Context(), opts, NewSettleRequest(pm, big.NewInt(1_000_000)))
		}},
		{"callSettleFor", func(m *PoolManager, opts *bind.CallOptions) (*big.Int, error) {
			return m.CallSettleFor(t.Context(), opts, NewSettleForRequest(pm, testSender, big.NewInt(1_000_000)))
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, opts := range []*bind.CallOptions{nil, bind.NewCallOptions().WithFrom(testSender).WithBlock("0x10")} {
				c := mocks.NewClient(t)
				c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
					call := e.Params.([]any)[0].(map[string]any)
					require.Equal(t, "0xf4240", call["value"])
					if opts == nil {
						require.NotContains(t, call, "from")
						require.Equal(t, rpc.BlockTagLatest, e.Params.([]any)[1])
					} else {
						require.Equal(t, testSender.String(), call["from"])
						require.Equal(t, "0x10", e.Params.([]any)[1])
					}
					*e.Result.(*string) = encoding.Hex.EncodePrefixed(paidWord)
					return nil
				}).Once()

				paid, err := tc.call(New(c), opts)
				require.NoError(t, err)
				require.Zero(t, big.NewInt(1_000_000).Cmp(paid))
				if opts != nil {
					require.Nil(t, opts.Value)
				}
			}

			paid, err := tc.call(New(mocks.NewClient(t)), bind.NewCallOptions().WithValue(big.NewInt(1)))
			require.Nil(t, paid)
			require.ErrorIs(t, err, bind.ErrInvalidOptions)
			require.ErrorContains(t, err, "value must be nil")
		})
	}
}

func TestPoolManagerReadsMapRevertToSentinel(t *testing.T) {
	pm := testPoolManager
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)

	for _, tc := range []struct {
		name string
		err  *abi.Error
		args []any
		raw  jsontext.Value
		want error
	}{
		{"CurrencyNotSettled", errCurrencyNotSettled, nil, nil, ErrCurrencyNotSettled},
		{"PoolNotInitialized", errPoolNotInitialized, nil, nil, ErrPoolNotInitialized},
		{"AlreadyUnlocked", errAlreadyUnlocked, nil, nil, ErrAlreadyUnlocked},
		{"ManagerLocked", errManagerLocked, nil, nil, ErrManagerLocked},
		{"TickSpacingTooLarge", errTickSpacingTooLarge, []any{big.NewInt(40000)}, nil, ErrTickSpacingTooLarge},
		{"TickSpacingTooSmall", errTickSpacingTooSmall, []any{big.NewInt(0)}, nil, ErrTickSpacingTooSmall},
		{"CurrenciesOutOfOrderOrEqual", errCurrenciesOutOfOrderOrEqual, []any{testCurrency1, testCurrency0}, nil, ErrCurrenciesOutOfOrderOrEqual},
		{"UnauthorizedDynamicLPFeeUpdate", errUnauthorizedDynamicLPFeeUpdate, nil, nil, ErrUnauthorizedDynamicLPFeeUpdate},
		{"SwapAmountCannotBeZero", errSwapAmountCannotBeZero, nil, nil, ErrSwapAmountCannotBeZero},
		{"NonzeroNativeValue", errNonzeroNativeValue, nil, nil, ErrNonzeroNativeValue},
		{"MustClearExactPositiveDelta", errMustClearExactPositiveDelta, nil, nil, ErrMustClearExactPositiveDelta},
		{"ProtocolFeeTooLarge", errProtocolFeeTooLarge, []any{big.NewInt(1001)}, nil, ErrProtocolFeeTooLarge},
		{"InvalidCaller", errInvalidCaller, nil, nil, ErrInvalidCaller},
		{"ProtocolFeeCurrencySynced", errProtocolFeeCurrencySynced, nil, nil, ErrProtocolFeeCurrencySynced},
		{"PoolAlreadyInitialized", errPoolAlreadyInitialized, nil, nil, ErrPoolAlreadyInitialized},
		{"TicksMisordered", errTicksMisordered, []any{big.NewInt(60), big.NewInt(-60)}, nil, ErrTicksMisordered},
		{"TickLowerOutOfBounds", errTickLowerOutOfBounds, []any{big.NewInt(-887273)}, nil, ErrTickLowerOutOfBounds},
		{"TickUpperOutOfBounds", errTickUpperOutOfBounds, []any{big.NewInt(887273)}, nil, ErrTickUpperOutOfBounds},
		{"TickLiquidityOverflow", errTickLiquidityOverflow, []any{big.NewInt(60)}, nil, ErrTickLiquidityOverflow},
		{"NoLiquidityToReceiveFees", errNoLiquidityToReceiveFees, nil, nil, ErrNoLiquidityToReceiveFees},
		{"TickMisaligned", errTickMisaligned, []any{big.NewInt(61), big.NewInt(60)}, nil, ErrTickMisaligned},
		{"InvalidTick", errInvalidTick, []any{big.NewInt(900000)}, nil, ErrInvalidTick},
		{"InvalidSqrtPrice", errInvalidSqrtPrice, []any{big.NewInt(1)}, nil, ErrInvalidSqrtPrice},
		{"HookAddressNotValid", errHookAddressNotValid, []any{testHooks}, nil, ErrHookAddressNotValid},
		{"CannotUpdateEmptyPosition", errCannotUpdateEmptyPosition, nil, nil, ErrCannotUpdateEmptyPosition},
		{"DelegateCallNotAllowed", errDelegateCallNotAllowed, nil, nil, ErrDelegateCallNotAllowed},
		{"NativeTransferFailed", errNativeTransferFailed, nil, nil, ErrNativeTransferFailed},
		{"ERC20TransferFailed", errERC20TransferFailed, nil, nil, ErrERC20TransferFailed},
		{"SafeCastOverflow", errSafeCastOverflow, nil, nil, ErrSafeCastOverflow},
		{"PriceLimitAlreadyExceeded", errPriceLimitAlreadyExceeded, []any{big.NewInt(2), big.NewInt(3)}, nil, ErrPriceLimitAlreadyExceeded},
		{"PriceLimitOutOfBounds", errPriceLimitOutOfBounds, []any{big.NewInt(1)}, nil, ErrPriceLimitOutOfBounds},
		{"InvalidFeeForExactOut", errInvalidFeeForExactOut, nil, nil, ErrInvalidFeeForExactOut},
		{"InvalidHookResponse", errInvalidHookResponse, nil, nil, ErrInvalidHookResponse},
		{"HookDeltaExceedsSwapAmount", errHookDeltaExceedsSwapAmount, nil, nil, ErrHookDeltaExceedsSwapAmount},
		{"LPFeeTooLarge", errLPFeeTooLarge, []any{big.NewInt(1_000_001)}, nil, ErrLPFeeTooLarge},
		{"InvalidPriceOrLiquidity", errInvalidPriceOrLiquidity, nil, nil, ErrInvalidPriceOrLiquidity},
		{"InvalidPrice", errInvalidPrice, nil, nil, ErrInvalidPrice},
		{"PriceOverflow", errPriceOverflow, nil, nil, ErrPriceOverflow},
		{"SqrtPriceMathNotEnoughLiquidity", errSqrtPriceMathNotEnoughLiquidity, nil, nil, ErrNotEnoughLiquidity},
		{"WrappedError", errWrappedError, []any{testHooks, []byte{0x12, 0x34, 0x56, 0x78}, []byte{}, []byte{0xde, 0xad, 0xbe, 0xef}}, nil, ErrWrappedError},
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

			tick, err := New(c).CallInitialize(t.Context(), nil, NewInitializeRequest(pm, key, big.NewInt(1)))
			require.Nil(t, tick)
			require.ErrorIs(t, err, tc.want)
			require.ErrorContains(t, err, "uniswapv4: callInitialize: ")
			got, ok := errors.AsType[*rpc.ResponseError](err)
			require.True(t, ok)
			require.Same(t, rpcErr, got)
			if !errors.Is(tc.want, ErrUnknown) {
				require.NotErrorIs(t, err, ErrUnknown)
			}
		})
	}
}

func TestPoolManagerWrapRevertErrorHookCallFailed(t *testing.T) {
	wrappedSel := errWrappedError.Selector()
	hookCallFailed := errHookCallFailed.Selector()

	failedArgs, err := abi.Pack(errWrappedError.Inputs, testHooks, []byte{0x12, 0x34, 0x56, 0x78}, []byte{0x01}, hookCallFailed[:])
	require.NoError(t, err)
	rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(append(wrappedSel[:], failedArgs...)) + `"`)}

	wrapped := (&PoolManager{}).wrapRevertError(rpcErr)
	require.ErrorIs(t, wrapped, ErrWrappedError)
	require.ErrorIs(t, wrapped, ErrHookCallFailed)
	require.NotErrorIs(t, wrapped, ErrUnknown)

	shortDetailsArgs, err := abi.Pack(errWrappedError.Inputs, testHooks, []byte{0x12, 0x34, 0x56, 0x78}, []byte{}, []byte{0x01, 0x02})
	require.NoError(t, err)

	for _, data := range [][]byte{
		append(wrappedSel[:], shortDetailsArgs...),
		append(wrappedSel[:], 0x00),
	} {
		wrapped := (&PoolManager{}).wrapRevertError(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(data) + `"`)})
		require.ErrorIs(t, wrapped, ErrWrappedError)
		require.NotErrorIs(t, wrapped, ErrHookCallFailed)
	}
}

func TestPoolManagerWrapRevertErrorPassesThroughUndecodable(t *testing.T) {
	m := &PoolManager{}
	plain := errors.New("dial tcp: connection refused")
	require.Same(t, plain, m.wrapRevertError(plain))
	require.Nil(t, m.wrapRevertError(nil))

	for _, data := range []jsontext.Value{nil, jsontext.Value(`"0x"`), jsontext.Value(`"0x123456"`), jsontext.Value(`"0xzz"`), jsontext.Value(`{"x":1}`), jsontext.Value(`42`)} {
		t.Run(string(data), func(t *testing.T) {
			rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: data}
			require.Same(t, rpcErr, m.wrapRevertError(rpcErr))

			wrapped := fmt.Errorf("rpc: %s: %w", rpc.MethodETHEstimateGas, rpcErr)
			require.Same(t, wrapped, m.wrapRevertError(wrapped))
		})
	}
}

func TestPoolManagerRevertSelectorsAreUnique(t *testing.T) {
	seen := map[abi.Selector]string{}
	for _, e := range []*abi.Error{
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
		sel string
	}{
		{errPoolAlreadyInitialized, "PoolAlreadyInitialized()", "0x7983c051"},
		{errTicksMisordered, "TicksMisordered(int24,int24)", "0xc4433ed5"},
		{errTickLowerOutOfBounds, "TickLowerOutOfBounds(int24)", "0xd5e2f7ab"},
		{errTickUpperOutOfBounds, "TickUpperOutOfBounds(int24)", "0x1ad777f8"},
		{errTickLiquidityOverflow, "TickLiquidityOverflow(int24)", "0xb8e3c385"},
		{errNoLiquidityToReceiveFees, "NoLiquidityToReceiveFees()", "0xa74f97ab"},
		{errTickMisaligned, "TickMisaligned(int24,int24)", "0xd4d8f3e6"},
		{errInvalidTick, "InvalidTick(int24)", "0x8b86327a"},
		{errInvalidSqrtPrice, "InvalidSqrtPrice(uint160)", "0x61487524"},
		{errHookAddressNotValid, "HookAddressNotValid(address)", "0xe65af6a0"},
		{errCannotUpdateEmptyPosition, "CannotUpdateEmptyPosition()", "0xaefeb924"},
		{errDelegateCallNotAllowed, "DelegateCallNotAllowed()", "0x0d89438e"},
		{errNativeTransferFailed, "NativeTransferFailed()", "0xf4b3b1bc"},
		{errERC20TransferFailed, "ERC20TransferFailed()", "0xf27f64e4"},
	} {
		require.Equal(t, tc.sig, tc.err.Signature())
		sel := tc.err.Selector()
		require.Equal(t, tc.sel, encoding.Hex.EncodePrefixed(sel[:]), tc.sig)
	}
}

func TestPoolManagerReadsRejectMalformedResult(t *testing.T) {
	pm := testPoolManager
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	liq := NewLiquidityParams(big.NewInt(-60), big.NewInt(60), big.NewInt(1), [32]byte{})
	swap := NewSwapAmountParams(true, big.NewInt(-1), big.NewInt(1))

	for _, result := range []string{"0x", "0xzz", "0x01", "not hex"} {
		t.Run(result, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				*e.Result.(*string) = result
				return nil
			}).Times(16)
			m := New(c)
			ctx := t.Context()

			for _, tc := range []struct {
				prefix string
				call   func() (any, error)
			}{
				{"protocolFeesAccrued", func() (any, error) {
					return m.ProtocolFeesAccrued(ctx, nil, NewProtocolFeesAccruedRequest(pm, testCurrency0))
				}},
				{"protocolFeeController", func() (any, error) {
					return m.ProtocolFeeController(ctx, nil, NewProtocolFeeControllerRequest(pm))
				}},
				{"extsload", func() (any, error) {
					v, err := m.Extsload(ctx, nil, NewExtsloadRequest(pm, [32]byte{}))
					if err == nil {
						return v, nil
					}
					require.Equal(t, [32]byte{}, v)
					return nil, err
				}},
				{"extsloadRange", func() (any, error) {
					return m.ExtsloadRange(ctx, nil, NewExtsloadRangeRequest(pm, [32]byte{}, big.NewInt(1)))
				}},
				{"extsloadSlots", func() (any, error) {
					return m.ExtsloadSlots(ctx, nil, NewExtsloadSlotsRequest(pm, [][32]byte{{}}))
				}},
				{"exttload", func() (any, error) {
					v, err := m.Exttload(ctx, nil, NewExttloadRequest(pm, [32]byte{}))
					if err == nil {
						return v, nil
					}
					require.Equal(t, [32]byte{}, v)
					return nil, err
				}},
				{"exttloadSlots", func() (any, error) {
					return m.ExttloadSlots(ctx, nil, NewExttloadSlotsRequest(pm, [][32]byte{{}}))
				}},
				{"callInitialize", func() (any, error) {
					return m.CallInitialize(ctx, nil, NewInitializeRequest(pm, key, big.NewInt(1)))
				}},
				{"callModifyLiquidity", func() (any, error) {
					caller, fees, err := m.CallModifyLiquidity(ctx, nil, NewModifyLiquidityRequest(pm, key, liq, nil))
					require.Nil(t, caller)
					require.Nil(t, fees)
					return nil, err
				}},
				{"callSwap", func() (any, error) {
					return m.CallSwap(ctx, nil, NewSwapRequest(pm, key, swap, nil))
				}},
				{"callDonate", func() (any, error) {
					return m.CallDonate(ctx, nil, NewDonateRequest(pm, key, big.NewInt(1), big.NewInt(1), nil))
				}},
				{"callUnlock", func() (any, error) {
					return m.CallUnlock(ctx, nil, NewUnlockRequest(pm, nil))
				}},
				{"callSettle", func() (any, error) {
					return m.CallSettle(ctx, nil, NewSettleRequest(pm, nil))
				}},
				{"callSettleFor", func() (any, error) {
					return m.CallSettleFor(ctx, nil, NewSettleForRequest(pm, testSender, nil))
				}},
				{"callCollectProtocolFees", func() (any, error) {
					return m.CallCollectProtocolFees(ctx, nil, NewCollectProtocolFeesRequest(pm, testSender, testCurrency0, big.NewInt(0)))
				}},
				{"extsloadRange again", func() (any, error) {
					return m.ExtsloadRange(ctx, nil, NewExtsloadRangeRequest(pm, [32]byte{}, big.NewInt(0)))
				}},
			} {
				out, err := tc.call()
				require.Error(t, err, tc.prefix)
				require.ErrorContains(t, err, "uniswapv4: ")
				require.Nil(t, out, tc.prefix)
			}
		})
	}
}

func TestPoolManagerMethodsRejectInvalidRequestWithoutRPC(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	pm := testPoolManager
	poolKey := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	opts := bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder())
	m := New(mocks.NewClient(t))
	ctx := t.Context()

	for _, tc := range []struct {
		name   string
		prefix string
		call   func() error
		want   error
	}{
		{"protocolFeesAccrued", "uniswapv4: protocolFeesAccrued: ", func() error {
			_, err := m.ProtocolFeesAccrued(ctx, nil, NewProtocolFeesAccruedRequest(pm, nil))
			return err
		}, bind.ErrInvalidRequest},
		{"protocolFeeController", "uniswapv4: protocolFeeController: ", func() error {
			_, err := m.ProtocolFeeController(ctx, nil, &ProtocolFeeControllerRequest{})
			return err
		}, bind.ErrInvalidRequest},
		{"extsload", "uniswapv4: extsload: ", func() error {
			_, err := m.Extsload(ctx, nil, &ExtsloadRequest{})
			return err
		}, bind.ErrInvalidRequest},
		{"extsloadRange", "uniswapv4: extsloadRange: ", func() error {
			_, err := m.ExtsloadRange(ctx, nil, NewExtsloadRangeRequest(pm, [32]byte{}, nil))
			return err
		}, bind.ErrInvalidRequest},
		{"extsloadRange negative", "uniswapv4: extsloadRange: ", func() error {
			_, err := m.ExtsloadRange(ctx, nil, NewExtsloadRangeRequest(pm, [32]byte{}, big.NewInt(-1)))
			return err
		}, abi.ErrIntegerOutOfRange},
		{"extsloadSlots", "uniswapv4: extsloadSlots: ", func() error {
			_, err := m.ExtsloadSlots(ctx, nil, NewExtsloadSlotsRequest(nil, nil))
			return err
		}, bind.ErrInvalidRequest},
		{"exttload", "uniswapv4: exttload: ", func() error {
			_, err := m.Exttload(ctx, nil, &ExttloadRequest{})
			return err
		}, bind.ErrInvalidRequest},
		{"exttloadSlots", "uniswapv4: exttloadSlots: ", func() error {
			_, err := m.ExttloadSlots(ctx, nil, &ExttloadSlotsRequest{})
			return err
		}, bind.ErrInvalidRequest},
		{"callInitialize", "uniswapv4: callInitialize: ", func() error {
			_, err := m.CallInitialize(ctx, nil, NewInitializeRequest(pm, nil, big.NewInt(1)))
			return err
		}, bind.ErrInvalidRequest},
		{"callInitialize fee overflow", "uniswapv4: callInitialize: ", func() error {
			_, err := m.CallInitialize(ctx, nil, NewInitializeRequest(pm, NewPoolKey(testCurrency0, testCurrency1, big.NewInt(1<<24), big.NewInt(60), testHooks), big.NewInt(1)))
			return err
		}, abi.ErrIntegerOutOfRange},
		{"callModifyLiquidity", "uniswapv4: callModifyLiquidity: ", func() error {
			_, _, err := m.CallModifyLiquidity(ctx, nil, NewModifyLiquidityRequest(pm, poolKey, nil, nil))
			return err
		}, bind.ErrInvalidRequest},
		{"callSwap", "uniswapv4: callSwap: ", func() error {
			_, err := m.CallSwap(ctx, nil, NewSwapRequest(pm, poolKey, &SwapAmountParams{}, nil))
			return err
		}, bind.ErrInvalidRequest},
		{"callDonate", "uniswapv4: callDonate: ", func() error {
			_, err := m.CallDonate(ctx, nil, NewDonateRequest(pm, poolKey, nil, big.NewInt(1), nil))
			return err
		}, bind.ErrInvalidRequest},
		{"callUnlock", "uniswapv4: callUnlock: ", func() error {
			_, err := m.CallUnlock(ctx, nil, &UnlockRequest{})
			return err
		}, bind.ErrInvalidRequest},
		{"callSettle", "uniswapv4: callSettle: ", func() error {
			_, err := m.CallSettle(ctx, nil, NewSettleRequest(pm, big.NewInt(-1)))
			return err
		}, bind.ErrInvalidRequest},
		{"callSettleFor", "uniswapv4: callSettleFor: ", func() error {
			_, err := m.CallSettleFor(ctx, nil, NewSettleForRequest(pm, nil, nil))
			return err
		}, bind.ErrInvalidRequest},
		{"callCollectProtocolFees", "uniswapv4: callCollectProtocolFees: ", func() error {
			_, err := m.CallCollectProtocolFees(ctx, nil, NewCollectProtocolFeesRequest(pm, testSender, testCurrency0, nil))
			return err
		}, bind.ErrInvalidRequest},
		{"initialize", "uniswapv4: initialize: ", func() error {
			_, _, err := m.Initialize(ctx, opts, NewInitializeRequest(pm, poolKey, nil))
			return err
		}, bind.ErrInvalidRequest},
		{"modifyLiquidity", "uniswapv4: modifyLiquidity: ", func() error {
			_, _, err := m.ModifyLiquidity(ctx, opts, NewModifyLiquidityRequest(pm, &PoolKey{}, nil, nil))
			return err
		}, bind.ErrInvalidRequest},
		{"modifyLiquidity tick overflow", "uniswapv4: modifyLiquidity: ", func() error {
			_, _, err := m.ModifyLiquidity(ctx, opts, NewModifyLiquidityRequest(pm, poolKey, NewLiquidityParams(big.NewInt(1<<23), big.NewInt(0), big.NewInt(1), [32]byte{}), nil))
			return err
		}, abi.ErrIntegerOutOfRange},
		{"swap", "uniswapv4: swap: ", func() error {
			_, _, err := m.Swap(ctx, opts, NewSwapRequest(pm, poolKey, nil, nil))
			return err
		}, bind.ErrInvalidRequest},
		{"donate", "uniswapv4: donate: ", func() error {
			_, _, err := m.Donate(ctx, opts, NewDonateRequest(nil, poolKey, big.NewInt(1), big.NewInt(1), nil))
			return err
		}, bind.ErrInvalidRequest},
		{"donate negative", "uniswapv4: donate: ", func() error {
			_, _, err := m.Donate(ctx, opts, NewDonateRequest(pm, poolKey, big.NewInt(-1), big.NewInt(1), nil))
			return err
		}, abi.ErrIntegerOutOfRange},
		{"unlock", "uniswapv4: unlock: ", func() error {
			_, _, err := m.Unlock(ctx, opts, &UnlockRequest{})
			return err
		}, bind.ErrInvalidRequest},
		{"sync", "uniswapv4: sync: ", func() error {
			_, _, err := m.Sync(ctx, opts, NewSyncRequest(pm, nil))
			return err
		}, bind.ErrInvalidRequest},
		{"take", "uniswapv4: take: ", func() error {
			_, _, err := m.Take(ctx, opts, NewTakeRequest(pm, testCurrency0, nil, big.NewInt(1)))
			return err
		}, bind.ErrInvalidRequest},
		{"settle", "uniswapv4: settle: ", func() error {
			_, _, err := m.Settle(ctx, opts, &SettleRequest{})
			return err
		}, bind.ErrInvalidRequest},
		{"settleFor", "uniswapv4: settleFor: ", func() error {
			_, _, err := m.SettleFor(ctx, opts, NewSettleForRequest(pm, testSender, big.NewInt(-5)))
			return err
		}, bind.ErrInvalidRequest},
		{"clear", "uniswapv4: clear: ", func() error {
			_, _, err := m.Clear(ctx, opts, NewClearRequest(pm, testCurrency0, nil))
			return err
		}, bind.ErrInvalidRequest},
		{"mint", "uniswapv4: mint: ", func() error {
			_, _, err := m.Mint(ctx, opts, NewMintRequest(pm, testSender, nil, big.NewInt(1)))
			return err
		}, bind.ErrInvalidRequest},
		{"burn", "uniswapv4: burn: ", func() error {
			_, _, err := m.Burn(ctx, opts, NewBurnRequest(pm, nil, big.NewInt(1), big.NewInt(1)))
			return err
		}, bind.ErrInvalidRequest},
		{"updateDynamicLPFee", "uniswapv4: updateDynamicLPFee: ", func() error {
			_, _, err := m.UpdateDynamicLPFee(ctx, opts, NewUpdateDynamicLPFeeRequest(pm, poolKey, nil))
			return err
		}, bind.ErrInvalidRequest},
		{"updateDynamicLPFee overflow", "uniswapv4: updateDynamicLPFee: ", func() error {
			_, _, err := m.UpdateDynamicLPFee(ctx, opts, NewUpdateDynamicLPFeeRequest(pm, poolKey, big.NewInt(1<<24)))
			return err
		}, abi.ErrIntegerOutOfRange},
		{"setProtocolFee", "uniswapv4: setProtocolFee: ", func() error {
			_, _, err := m.SetProtocolFee(ctx, opts, NewSetProtocolFeeRequest(pm, nil, big.NewInt(0)))
			return err
		}, bind.ErrInvalidRequest},
		{"setProtocolFeeController", "uniswapv4: setProtocolFeeController: ", func() error {
			_, _, err := m.SetProtocolFeeController(ctx, opts, NewSetProtocolFeeControllerRequest(pm, nil))
			return err
		}, bind.ErrInvalidRequest},
		{"collectProtocolFees", "uniswapv4: collectProtocolFees: ", func() error {
			_, _, err := m.CollectProtocolFees(ctx, opts, &CollectProtocolFeesRequest{})
			return err
		}, bind.ErrInvalidRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call()
			require.ErrorIs(t, err, tc.want)
			require.ErrorContains(t, err, tc.prefix)
		})
	}

	_, _, err = m.Initialize(ctx, nil, NewInitializeRequest(pm, poolKey, big.NewInt(1)))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv4: initialize: bind: transact: invalid options: options is nil")

	_, _, err = m.Settle(ctx, opts, NewSettleRequest(pm, new(big.Int).Lsh(big.NewInt(1), 256)))
	require.ErrorIs(t, err, bind.ErrInvalidRequest)
	require.EqualError(t, err, "uniswapv4: settle: invalid request: nativeAmount exceeds uint256")

	_, err = m.CallSettleFor(ctx, nil, NewSettleForRequest(pm, testSender, new(big.Int).Lsh(big.NewInt(1), 300)))
	require.ErrorIs(t, err, bind.ErrInvalidRequest)
	require.EqualError(t, err, "uniswapv4: callSettleFor: invalid request: nativeAmount exceeds uint256")

	_, _, err = m.Settle(ctx, opts.WithValue(big.NewInt(1)), NewSettleRequest(pm, nil))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv4: settle: invalid options: value must be nil")

	_, _, err = m.SettleFor(ctx, opts, NewSettleForRequest(pm, testSender, nil))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv4: settleFor: invalid options: value must be nil")
}

func TestPoolManagerWritesSignAndSend(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sender := core.PubkeyToAddress(key.PublicKey())
	pm := testPoolManager
	poolKey := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	liq := NewLiquidityParams(big.NewInt(-120), big.NewInt(120), big.NewInt(-5), [32]byte{0x09})
	swap := NewSwapAmountParams(false, big.NewInt(500), new(big.Int).Lsh(big.NewInt(1), 100))
	native := big.NewInt(1_000_000_007)

	must := func(data []byte, err error) []byte {
		t.Helper()
		require.NoError(t, err)
		return data
	}

	for _, tc := range []struct {
		name     string
		calldata []byte
		value    *big.Int
		send     func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error)
	}{
		{"initialize", must(EncodeInitialize(poolKey, new(big.Int).Lsh(big.NewInt(1), 96))), nil, func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return m.Initialize(t.Context(), opts, NewInitializeRequest(pm, poolKey, new(big.Int).Lsh(big.NewInt(1), 96)))
		}},
		{"modifyLiquidity", must(EncodeModifyLiquidity(poolKey, liq, []byte{0x01, 0x02})), nil, func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return m.ModifyLiquidity(t.Context(), opts, NewModifyLiquidityRequest(pm, poolKey, liq, []byte{0x01, 0x02}))
		}},
		{"swap", must(EncodeSwap(poolKey, swap, nil)), nil, func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return m.Swap(t.Context(), opts, NewSwapRequest(pm, poolKey, swap, nil))
		}},
		{"donate", must(EncodeDonate(poolKey, big.NewInt(10), big.NewInt(0), nil)), nil, func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return m.Donate(t.Context(), opts, NewDonateRequest(pm, poolKey, big.NewInt(10), big.NewInt(0), nil))
		}},
		{"unlock", must(EncodeUnlock([]byte{0xab})), nil, func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return m.Unlock(t.Context(), opts, NewUnlockRequest(pm, []byte{0xab}))
		}},
		{"sync", must(EncodeSync(types.ZeroAddress())), nil, func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return m.Sync(t.Context(), opts, NewSyncRequest(pm, types.ZeroAddress()))
		}},
		{"take", must(EncodeTake(testCurrency0, sender, big.NewInt(42))), nil, func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return m.Take(t.Context(), opts, NewTakeRequest(pm, testCurrency0, sender, big.NewInt(42)))
		}},
		{"settle", must(EncodeSettle()), native, func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return m.Settle(t.Context(), opts, NewSettleRequest(pm, native))
		}},
		{"settle without native", must(EncodeSettle()), nil, func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return m.Settle(t.Context(), opts, NewSettleRequest(pm, nil))
		}},
		{"settleFor", must(EncodeSettleFor(testHooks)), native, func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return m.SettleFor(t.Context(), opts, NewSettleForRequest(pm, testHooks, native))
		}},
		{"clear", must(EncodeClear(testCurrency1, big.NewInt(3))), nil, func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return m.Clear(t.Context(), opts, NewClearRequest(pm, testCurrency1, big.NewInt(3)))
		}},
		{"mint", must(EncodeMint(sender, testCurrency0.Big(), big.NewInt(8))), nil, func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return m.Mint(t.Context(), opts, NewMintRequest(pm, sender, testCurrency0.Big(), big.NewInt(8)))
		}},
		{"burn", must(EncodeBurn(sender, testCurrency1.Big(), big.NewInt(9))), nil, func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return m.Burn(t.Context(), opts, NewBurnRequest(pm, sender, testCurrency1.Big(), big.NewInt(9)))
		}},
		{"updateDynamicLPFee", must(EncodeUpdateDynamicLPFee(poolKey, big.NewInt(500))), nil, func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return m.UpdateDynamicLPFee(t.Context(), opts, NewUpdateDynamicLPFeeRequest(pm, poolKey, big.NewInt(500)))
		}},
		{"setProtocolFee", must(EncodeSetProtocolFee(poolKey, big.NewInt(1000))), nil, func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return m.SetProtocolFee(t.Context(), opts, NewSetProtocolFeeRequest(pm, poolKey, big.NewInt(1000)))
		}},
		{"setProtocolFeeController", must(EncodeSetProtocolFeeController(sender)), nil, func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return m.SetProtocolFeeController(t.Context(), opts, NewSetProtocolFeeControllerRequest(pm, sender))
		}},
		{"collectProtocolFees", must(EncodeCollectProtocolFees(sender, testCurrency0, big.NewInt(0))), nil, func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return m.CollectProtocolFees(t.Context(), opts, NewCollectProtocolFeesRequest(pm, sender, testCurrency0, big.NewInt(0)))
		}},
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
			signed, hash, err := tc.send(New(c), opts)
			require.NoError(t, err)
			require.Nil(t, opts.Value)
			require.True(t, hash.Equal(core.Keccak256(raw)))
			require.Equal(t, uint64(7), signed.AddressNonce())

			decoded, err := core.DecodeTransaction(raw)
			require.NoError(t, err)
			dyn, ok := decoded.(*core.DynamicFeeTx)
			require.True(t, ok)
			require.True(t, dyn.To.Equal(pm))
			if tc.value == nil {
				require.Zero(t, dyn.Value.Sign())
			} else {
				require.Zero(t, tc.value.Cmp(dyn.Value))
			}
			require.Equal(t, tc.calldata, dyn.Data)
			got, err := decoded.Sender()
			require.NoError(t, err)
			require.True(t, got.Equal(sender))
		})
	}
}

func TestPoolManagerWritesMapEstimateRevertToSentinel(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	pm := testPoolManager
	poolKey := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)

	for _, tc := range []struct {
		name string
		err  *abi.Error
		args []any
		raw  jsontext.Value
		send func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error)
		want error
	}{
		{"initialize already initialized", errPoolAlreadyInitialized, nil, nil, func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return m.Initialize(t.Context(), opts, NewInitializeRequest(pm, poolKey, big.NewInt(1)))
		}, ErrPoolAlreadyInitialized},
		{"swap locked", errManagerLocked, nil, nil, func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return m.Swap(t.Context(), opts, NewSwapRequest(pm, poolKey, NewSwapAmountParams(true, big.NewInt(-1), big.NewInt(1)), nil))
		}, ErrManagerLocked},
		{"settle locked", errManagerLocked, nil, nil, func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return m.Settle(t.Context(), opts, NewSettleRequest(pm, big.NewInt(1)))
		}, ErrManagerLocked},
		{"setProtocolFee invalid caller", errInvalidCaller, nil, nil, func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return m.SetProtocolFee(t.Context(), opts, NewSetProtocolFeeRequest(pm, poolKey, big.NewInt(1)))
		}, ErrInvalidCaller},
		{"updateDynamicLPFee unauthorized", errUnauthorizedDynamicLPFeeUpdate, nil, nil, func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return m.UpdateDynamicLPFee(t.Context(), opts, NewUpdateDynamicLPFeeRequest(pm, poolKey, big.NewInt(1)))
		}, ErrUnauthorizedDynamicLPFeeUpdate},
		{"setProtocolFeeController owned", nil, nil, jsontext.Value(`"0x08c379a0"`), func(m *PoolManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return m.SetProtocolFeeController(t.Context(), opts, NewSetProtocolFeeControllerRequest(pm, testSender))
		}, ErrUnknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := tc.raw
			if tc.err != nil {
				sel := tc.err.Selector()
				packed, err := abi.Pack(tc.err.Inputs, tc.args...)
				require.NoError(t, err)
				data = jsontext.Value(`"` + encoding.Hex.EncodePrefixed(append(sel[:], packed...)) + `"`)
			}

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

			signed, hash, err := tc.send(New(c), bind.NewTransactOptions(key, builder))
			require.ErrorIs(t, err, tc.want)
			require.ErrorContains(t, err, "uniswapv4: ")
			require.Nil(t, signed)
			require.Nil(t, hash)
		})
	}
}

func TestPoolManagerWritesReturnSignedTxOnBroadcastFailure(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	pm := testPoolManager
	poolKey := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(9)
	builder.Limit(6_000_000)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(errors.New("dial tcp: connection refused")).Once()

	signed, hash, err := New(c).Initialize(t.Context(), bind.NewTransactOptions(key, builder), NewInitializeRequest(pm, poolKey, big.NewInt(1)))
	require.ErrorContains(t, err, "uniswapv4: initialize: ")
	require.ErrorContains(t, err, "connection refused")
	require.NotErrorIs(t, err, ErrUnknown)
	require.Nil(t, hash)
	require.NotNil(t, signed)
	require.Equal(t, uint64(9), signed.AddressNonce())
}

func TestPoolManagerReadsPassThroughTransportError(t *testing.T) {
	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(errors.New("dial tcp: connection refused")).Once()

	controller, err := New(c).ProtocolFeeController(t.Context(), nil, NewProtocolFeeControllerRequest(testPoolManager))
	require.Nil(t, controller)
	require.EqualError(t, err, "uniswapv4: protocolFeeController: dial tcp: connection refused")
	require.NotErrorIs(t, err, ErrUnknown)
}

func TestPoolManagerCallsPropagateContext(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(t.Context(), marker{}, "poolmanager")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	word, err := abi.Pack(abi.NewTypes(abi.Address), testSender)
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(got context.Context, e rpc.Element) error {
		require.Equal(t, "poolmanager", got.Value(marker{}))
		if e.Method == rpc.MethodETHCall {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(word)
		}
		return nil
	}).Twice()

	_, err = New(c).ProtocolFeeController(ctx, nil, NewProtocolFeeControllerRequest(testPoolManager))
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(1)
	builder.Limit(3_000_000)
	builder.Price(big.NewInt(2), big.NewInt(1))
	_, _, err = New(c).Sync(ctx, bind.NewTransactOptions(key, builder), NewSyncRequest(testPoolManager, testCurrency0))
	require.NoError(t, err)
}

func TestPoolManagerPromotesERC6909Claims(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	id := testCurrency0.Big()
	word, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(1))
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
		switch e.Method {
		case rpc.MethodETHCall:
			call := e.Params.([]any)[0].(map[string]any)
			require.Equal(t, testPoolManager.String(), call["to"])
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(word)
		case rpc.MethodETHSendRawTransaction:
			raw, err := encoding.Hex.Decode(e.Params.([]any)[0].(string))
			require.NoError(t, err)
			decoded, err := core.DecodeTransaction(raw)
			require.NoError(t, err)
			require.True(t, decoded.(*core.DynamicFeeTx).To.Equal(testPoolManager))
			*e.Result.(*string) = core.Keccak256(raw).String()
		default:
			t.Fatalf("unexpected method %s", e.Method)
		}
		return nil
	}).Times(12)

	var m IPoolManager = New(c)
	ctx := t.Context()

	supported, err := m.SupportsInterface(ctx, nil, erc165.NewSupportsInterfaceRequest(testPoolManager, [4]byte{0x0f, 0x63, 0x2f, 0xb3}))
	require.NoError(t, err)
	require.True(t, supported)

	balance, err := m.BalanceOf(ctx, nil, erc6909.NewBalanceOfRequest(testPoolManager, testSender, id))
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1).Cmp(balance))

	allowance, err := m.Allowance(ctx, nil, erc6909.NewAllowanceRequest(testPoolManager, testSender, testHooks, id))
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1).Cmp(allowance))

	operator, err := m.IsOperator(ctx, nil, erc6909.NewIsOperatorRequest(testPoolManager, testSender, testHooks))
	require.NoError(t, err)
	require.True(t, operator)

	approved, err := m.CallApprove(ctx, nil, erc6909.NewApproveRequest(testPoolManager, testHooks, id, big.NewInt(1)))
	require.NoError(t, err)
	require.True(t, approved)

	operatorSet, err := m.CallSetOperator(ctx, nil, erc6909.NewSetOperatorRequest(testPoolManager, testHooks, true))
	require.NoError(t, err)
	require.True(t, operatorSet)

	transferred, err := m.CallTransfer(ctx, nil, erc6909.NewTransferRequest(testPoolManager, testHooks, id, big.NewInt(1)))
	require.NoError(t, err)
	require.True(t, transferred)

	transferredFrom, err := m.CallTransferFrom(ctx, nil, erc6909.NewTransferFromRequest(testPoolManager, testSender, testHooks, id, big.NewInt(1)))
	require.NoError(t, err)
	require.True(t, transferredFrom)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(1)
	builder.Limit(100_000)
	builder.Price(big.NewInt(2), big.NewInt(1))
	opts := bind.NewTransactOptions(key, builder)

	_, hash, err := m.Approve(ctx, opts, erc6909.NewApproveRequest(testPoolManager, testHooks, id, big.NewInt(1)))
	require.NoError(t, err)
	require.NotNil(t, hash)

	_, hash, err = m.SetOperator(ctx, opts, erc6909.NewSetOperatorRequest(testPoolManager, testHooks, true))
	require.NoError(t, err)
	require.NotNil(t, hash)

	_, hash, err = m.Transfer(ctx, opts, erc6909.NewTransferRequest(testPoolManager, testHooks, id, big.NewInt(1)))
	require.NoError(t, err)
	require.NotNil(t, hash)

	_, hash, err = m.TransferFrom(ctx, opts, erc6909.NewTransferFromRequest(testPoolManager, testSender, testHooks, id, big.NewInt(1)))
	require.NoError(t, err)
	require.NotNil(t, hash)
}
