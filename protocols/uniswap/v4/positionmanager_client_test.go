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
	"github.com/ninja-protocol-labs/go-evmkit/protocols/permit2"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc721"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

func TestPositionManagerReadsSendCallAndDecode(t *testing.T) {
	pm := testPositionManager
	from := testSender
	opts := bind.NewCallOptions().WithFrom(from).WithBlock("0x1b4")
	tokenID := big.NewInt(77)
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	poolID := [25]byte{0xaa, 0xbb, 0xcc}
	sig := []byte{0x01, 0x02, 0x03}
	single := permit2.NewPermitSingle(permit2.NewPermitDetails(testCurrency0, big.NewInt(1), big.NewInt(2), big.NewInt(3)), testHooks, big.NewInt(4))
	batch := permit2.NewPermitBatch([]permit2.PermitDetails{*permit2.NewPermitDetails(testCurrency1, big.NewInt(5), big.NewInt(6), big.NewInt(7))}, testHooks, big.NewInt(8))

	var infoWord [32]byte
	copy(infoWord[:25], poolID[:])
	infoWord[25], infoWord[26], infoWord[27] = 0x00, 0x02, 0x58
	infoWord[28], infoWord[29], infoWord[30] = 0xff, 0xfd, 0xa8
	infoWord[31] = 0x01
	info := NewPositionInfo(poolID, big.NewInt(-600), big.NewInt(600), true)

	addressWord, err := abi.Pack(abi.NewTypes(abi.Address), testHooks)
	require.NoError(t, err)
	uintWord, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(1234))
	require.NoError(t, err)
	liquidityWord, err := abi.Pack(abi.NewTypes(abi.Uint128), big.NewInt(1_000_000))
	require.NoError(t, err)
	poolAndInfoWord, err := abi.Pack(abi.NewTypes(poolKeyTuple, abi.Uint256), []any{key.Currency0, key.Currency1, key.Fee, key.TickSpacing, key.Hooks}, new(big.Int).SetBytes(infoWord[:]))
	require.NoError(t, err)
	poolKeysWord, err := abi.Pack(abi.NewTypes(abi.Address, abi.Address, abi.Uint24, abi.Int24, abi.Address), key.Currency0, key.Currency1, key.Fee, key.TickSpacing, key.Hooks)
	require.NoError(t, err)
	tickWord, err := abi.Pack(abi.NewTypes(abi.Int24), big.NewInt(-887272))
	require.NoError(t, err)
	bytesWord, err := abi.Pack(abi.NewTypes(abi.Bytes), []byte{0xde, 0xad})
	require.NoError(t, err)
	emptyBytesWord, err := abi.Pack(abi.NewTypes(abi.Bytes), []byte{})
	require.NoError(t, err)

	poolManagerData, err := EncodePoolManager()
	require.NoError(t, err)
	weth9Data, err := EncodeWETH9()
	require.NoError(t, err)
	permit2Data, err := EncodePermit2Contract()
	require.NoError(t, err)
	nextTokenIDData, err := EncodeNextTokenID()
	require.NoError(t, err)
	liquidityData, err := EncodeGetPositionLiquidity(tokenID)
	require.NoError(t, err)
	poolAndInfoData, err := EncodeGetPoolAndPositionInfo(tokenID)
	require.NoError(t, err)
	positionInfoData, err := EncodePositionInfo(tokenID)
	require.NoError(t, err)
	poolKeysData, err := EncodePoolKeys(poolID)
	require.NoError(t, err)
	subscriberData, err := EncodeSubscriber(tokenID)
	require.NoError(t, err)
	gasLimitData, err := EncodeUnsubscribeGasLimit()
	require.NoError(t, err)
	noncesData, err := EncodeNonces(from, big.NewInt(3))
	require.NoError(t, err)
	initializePoolData, err := EncodeInitializePool(key, new(big.Int).Lsh(big.NewInt(1), 96))
	require.NoError(t, err)
	permit2PermitData, err := EncodePermit2Permit(from, single, sig)
	require.NoError(t, err)
	permit2PermitBatchData, err := EncodePermit2PermitBatch(from, batch, sig)
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		result   []byte
		call     func(p *PositionManager) (any, error)
		want     any
	}{
		{"poolManager", poolManagerData, addressWord, func(p *PositionManager) (any, error) {
			return p.PoolManager(t.Context(), opts, NewPoolManagerRequest(pm))
		}, testHooks},
		{"weth9", weth9Data, addressWord, func(p *PositionManager) (any, error) {
			return p.WETH9(t.Context(), opts, NewWETH9Request(pm))
		}, testHooks},
		{"permit2", permit2Data, addressWord, func(p *PositionManager) (any, error) {
			return p.Permit2(t.Context(), opts, NewPermit2Request(pm))
		}, testHooks},
		{"nextTokenId", nextTokenIDData, uintWord, func(p *PositionManager) (any, error) {
			return p.NextTokenID(t.Context(), opts, NewNextTokenIDRequest(pm))
		}, big.NewInt(1234)},
		{"getPositionLiquidity", liquidityData, liquidityWord, func(p *PositionManager) (any, error) {
			return p.GetPositionLiquidity(t.Context(), opts, NewGetPositionLiquidityRequest(pm, tokenID))
		}, big.NewInt(1_000_000)},
		{"getPoolAndPositionInfo", poolAndInfoData, poolAndInfoWord, func(p *PositionManager) (any, error) {
			gotKey, gotInfo, err := p.GetPoolAndPositionInfo(t.Context(), opts, NewGetPoolAndPositionInfoRequest(pm, tokenID))
			if err != nil {
				return nil, err
			}
			return []any{*gotKey, *gotInfo}, nil
		}, []any{*key, *info}},
		{"positionInfo", positionInfoData, infoWord[:], func(p *PositionManager) (any, error) {
			return p.PositionInfo(t.Context(), opts, NewPositionInfoRequest(pm, tokenID))
		}, info},
		{"poolKeys", poolKeysData, poolKeysWord, func(p *PositionManager) (any, error) {
			return p.PoolKeys(t.Context(), opts, NewPoolKeysRequest(pm, poolID))
		}, key},
		{"subscriber", subscriberData, addressWord, func(p *PositionManager) (any, error) {
			return p.Subscriber(t.Context(), opts, NewSubscriberRequest(pm, tokenID))
		}, testHooks},
		{"unsubscribeGasLimit", gasLimitData, uintWord, func(p *PositionManager) (any, error) {
			return p.UnsubscribeGasLimit(t.Context(), opts, NewUnsubscribeGasLimitRequest(pm))
		}, big.NewInt(1234)},
		{"nonces", noncesData, uintWord, func(p *PositionManager) (any, error) {
			return p.Nonces(t.Context(), opts, NewNoncesRequest(pm, from, big.NewInt(3)))
		}, big.NewInt(1234)},
		{"callInitializePool", initializePoolData, tickWord, func(p *PositionManager) (any, error) {
			return p.CallInitializePool(t.Context(), opts, NewInitializePoolRequest(pm, key, new(big.Int).Lsh(big.NewInt(1), 96)))
		}, big.NewInt(-887272)},
		{"callPermit2Permit", permit2PermitData, bytesWord, func(p *PositionManager) (any, error) {
			return p.CallPermit2Permit(t.Context(), opts, NewPermit2PermitRequest(pm, from, single, sig))
		}, []byte{0xde, 0xad}},
		{"callPermit2PermitBatch", permit2PermitBatchData, emptyBytesWord, func(p *PositionManager) (any, error) {
			return p.CallPermit2PermitBatch(t.Context(), opts, NewPermit2PermitBatchRequest(pm, from, batch, sig))
		}, []byte{}},
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

			got, err := tc.call(NewPositionManager(c))
			require.NoError(t, err)
			require.Equal(t, fmt.Sprint(tc.want), fmt.Sprint(got))
		})
	}
}

func TestPositionManagerReadsDefaultToLatestWithoutFrom(t *testing.T) {
	word, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(9))
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
		params := e.Params.([]any)
		call := params[0].(map[string]any)
		require.NotContains(t, call, "from")
		require.NotContains(t, call, "value")
		require.Equal(t, rpc.BlockTagLatest, params[1])
		*e.Result.(*string) = encoding.Hex.EncodePrefixed(word)
		return nil
	}).Once()

	next, err := NewPositionManager(c).NextTokenID(t.Context(), nil, NewNextTokenIDRequest(testPositionManager))
	require.NoError(t, err)
	require.Zero(t, big.NewInt(9).Cmp(next))
}

func TestPositionManagerCallMulticallSendsNativeAmountAsValue(t *testing.T) {
	pm := testPositionManager
	data := [][]byte{{0x01, 0x02}, {0x03}}
	calldata, err := EncodeMulticall(data)
	require.NoError(t, err)
	resultWord, err := abi.Pack(abi.NewTypes(abi.Slice(abi.Bytes)), []any{[]byte{0xaa}, []byte{}})
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
					call := e.Params.([]any)[0].(map[string]any)
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
					*e.Result.(*string) = encoding.Hex.EncodePrefixed(resultWord)
					return nil
				}).Once()

				results, err := NewPositionManager(c).CallMulticall(t.Context(), opts, NewMulticallRequest(pm, data, tc.native))
				require.NoError(t, err)
				require.Equal(t, [][]byte{{0xaa}, {}}, results)
				if opts != nil {
					require.Nil(t, opts.Value)
				}
			}
		})
	}

	results, err := NewPositionManager(mocks.NewClient(t)).CallMulticall(t.Context(), bind.NewCallOptions().WithValue(big.NewInt(1)), NewMulticallRequest(pm, data, nil))
	require.Nil(t, results)
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv4: callMulticall: invalid options: value must be nil")
}

func TestPositionManagerReadsMapRevertToSentinel(t *testing.T) {
	pm := testPositionManager
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)

	for _, tc := range []struct {
		name string
		err  *abi.Error
		args []any
		raw  jsontext.Value
		want error
	}{
		{"NotApproved", errNotApproved, []any{testSender}, nil, ErrNotApproved},
		{"DeadlinePassed", errDeadlinePassed, []any{big.NewInt(1_700_000_000)}, nil, ErrDeadlinePassed},
		{"PoolManagerMustBeLocked", errPoolManagerMustBeLocked, nil, nil, ErrPoolManagerMustBeLocked},
		{"SignatureDeadlineExpired", errSignatureDeadlineExpired, nil, nil, ErrSignatureDeadlineExpired},
		{"NoSelfPermit", errNoSelfPermit, nil, nil, ErrNoSelfPermit},
		{"Unauthorized", errUnauthorized, nil, nil, ErrUnauthorized},
		{"NotSubscribed", errNotSubscribed, nil, nil, ErrNotSubscribed},
		{"NoCodeSubscriber", errNoCodeSubscriber, nil, nil, ErrNoCodeSubscriber},
		{"GasLimitTooLow", errGasLimitTooLow, nil, nil, ErrGasLimitTooLow},
		{"SubscriptionReverted", errSubscriptionReverted, []any{testHooks, []byte{0xde, 0xad}}, nil, ErrSubscriptionReverted},
		{"ModifyLiquidityNotificationReverted", errModifyLiquidityNotificationReverted, []any{testHooks, []byte{}}, nil, ErrModifyLiquidityNotificationReverted},
		{"BurnNotificationReverted", errBurnNotificationReverted, []any{testHooks, []byte{0x01}}, nil, ErrBurnNotificationReverted},
		{"AlreadySubscribed", errAlreadySubscribed, []any{big.NewInt(1), testHooks}, nil, ErrAlreadySubscribed},
		{"NonceAlreadyUsed", errNonceAlreadyUsed, nil, nil, ErrNonceAlreadyUsed},
		{"ContractLocked", errContractLocked, nil, nil, ErrContractLocked},
		{"InvalidEthSender", errInvalidEthSender, nil, nil, ErrInvalidEthSender},
		{"DeltaNotPositive", errDeltaNotPositive, []any{testCurrency0}, nil, ErrDeltaNotPositive},
		{"DeltaNotNegative", errDeltaNotNegative, []any{testCurrency1}, nil, ErrDeltaNotNegative},
		{"InsufficientBalance", errInsufficientBalance, nil, nil, ErrInsufficientBalance},
		{"MaximumAmountExceeded", errMaximumAmountExceeded, []any{big.NewInt(1000), big.NewInt(1100)}, nil, ErrMaximumAmountExceeded},
		{"MinimumAmountInsufficient", errMinimumAmountInsufficient, []any{big.NewInt(900), big.NewInt(800)}, nil, ErrMinimumAmountInsufficient},
		{"InvalidSignatureLength", errInvalidSignatureLength, nil, nil, permit2.ErrInvalidSignatureLength},
		{"InvalidSignature", errInvalidSignature, nil, nil, permit2.ErrInvalidSignature},
		{"InvalidSigner", errInvalidSigner, nil, nil, permit2.ErrInvalidSigner},
		{"InvalidContractSignature", errInvalidContractSignature, nil, nil, permit2.ErrInvalidContractSignature},
		{"PoolManager CurrencyNotSettled", errCurrencyNotSettled, nil, nil, ErrCurrencyNotSettled},
		{"PoolManager PoolNotInitialized", errPoolNotInitialized, nil, nil, ErrPoolNotInitialized},
		{"PoolManager ManagerLocked", errManagerLocked, nil, nil, ErrManagerLocked},
		{"PoolManager TickMisaligned", errTickMisaligned, []any{big.NewInt(61), big.NewInt(60)}, nil, ErrTickMisaligned},
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

			tick, err := NewPositionManager(c).CallInitializePool(t.Context(), nil, NewInitializePoolRequest(pm, key, big.NewInt(1)))
			require.Nil(t, tick)
			require.ErrorIs(t, err, tc.want)
			require.ErrorContains(t, err, "uniswapv4: callInitializePool: ")
			got, ok := errors.AsType[*rpc.ResponseError](err)
			require.True(t, ok)
			require.Same(t, rpcErr, got)
			if !errors.Is(tc.want, ErrUnknown) {
				require.NotErrorIs(t, err, ErrUnknown)
			}
		})
	}
}

func TestPositionManagerWrapRevertErrorPassesThroughUndecodable(t *testing.T) {
	plain := errors.New("dial tcp: connection refused")
	require.Same(t, plain, mapPositionManagerRevertError(plain))
	require.Nil(t, mapPositionManagerRevertError(nil))

	for _, data := range []jsontext.Value{nil, jsontext.Value(`"0x"`), jsontext.Value(`"0x123456"`), jsontext.Value(`"0xzz"`), jsontext.Value(`{"x":1}`), jsontext.Value(`42`)} {
		t.Run(string(data), func(t *testing.T) {
			rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: data}
			require.Same(t, rpcErr, mapPositionManagerRevertError(rpcErr))

			wrapped := fmt.Errorf("rpc: %s: %w", rpc.MethodETHEstimateGas, rpcErr)
			require.Same(t, wrapped, mapPositionManagerRevertError(wrapped))
		})
	}
}

func TestPositionManagerRevertSelectorsAreUnique(t *testing.T) {
	seen := map[abi.Selector]string{}
	for _, e := range []*abi.Error{
		errNotApproved, errDeadlinePassed, errPoolManagerMustBeLocked, errSignatureDeadlineExpired, errNoSelfPermit,
		errUnauthorized, errNotSubscribed, errNoCodeSubscriber, errGasLimitTooLow, errSubscriptionReverted,
		errModifyLiquidityNotificationReverted, errBurnNotificationReverted, errAlreadySubscribed, errNonceAlreadyUsed,
		errContractLocked, errInvalidEthSender, errDeltaNotPositive, errDeltaNotNegative, errInsufficientBalance,
		errMaximumAmountExceeded, errMinimumAmountInsufficient, errInvalidSignatureLength, errInvalidSignature,
		errInvalidSigner, errInvalidContractSignature,
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
		{errNotApproved, "NotApproved(address)"},
		{errDeadlinePassed, "DeadlinePassed(uint256)"},
		{errSubscriptionReverted, "SubscriptionReverted(address,bytes)"},
		{errAlreadySubscribed, "AlreadySubscribed(uint256,address)"},
		{errDeltaNotPositive, "DeltaNotPositive(address)"},
		{errMaximumAmountExceeded, "MaximumAmountExceeded(uint128,uint128)"},
		{errInvalidSignatureLength, "InvalidSignatureLength()"},
		{errInvalidContractSignature, "InvalidContractSignature()"},
	} {
		require.Equal(t, tc.sig, tc.err.Signature())
		sel := tc.err.Selector()
		require.Equal(t, core.Keccak256([]byte(tc.sig)).Bytes()[:4], sel[:], tc.sig)
	}
}

func TestPositionManagerReadsRejectMalformedResult(t *testing.T) {
	pm := testPositionManager
	tokenID := big.NewInt(1)
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	single := permit2.NewPermitSingle(permit2.NewPermitDetails(testCurrency0, big.NewInt(1), big.NewInt(2), big.NewInt(3)), testHooks, big.NewInt(4))
	batch := permit2.NewPermitBatch([]permit2.PermitDetails{*permit2.NewPermitDetails(testCurrency1, big.NewInt(5), big.NewInt(6), big.NewInt(7))}, testHooks, big.NewInt(8))

	for _, result := range []string{"0x", "0xzz", "0x01", "not hex"} {
		t.Run(result, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				*e.Result.(*string) = result
				return nil
			}).Times(15)
			p := NewPositionManager(c)
			ctx := t.Context()

			poolManager, err := p.PoolManager(ctx, nil, NewPoolManagerRequest(pm))
			require.Nil(t, poolManager)
			require.ErrorContains(t, err, "uniswapv4: poolManager: ")

			weth9, err := p.WETH9(ctx, nil, NewWETH9Request(pm))
			require.Nil(t, weth9)
			require.ErrorContains(t, err, "uniswapv4: weth9: ")

			permit2Addr, err := p.Permit2(ctx, nil, NewPermit2Request(pm))
			require.Nil(t, permit2Addr)
			require.ErrorContains(t, err, "uniswapv4: permit2: ")

			next, err := p.NextTokenID(ctx, nil, NewNextTokenIDRequest(pm))
			require.Nil(t, next)
			require.ErrorContains(t, err, "uniswapv4: nextTokenId: ")

			liquidity, err := p.GetPositionLiquidity(ctx, nil, NewGetPositionLiquidityRequest(pm, tokenID))
			require.Nil(t, liquidity)
			require.ErrorContains(t, err, "uniswapv4: getPositionLiquidity: ")

			gotKey, gotInfo, err := p.GetPoolAndPositionInfo(ctx, nil, NewGetPoolAndPositionInfoRequest(pm, tokenID))
			require.Nil(t, gotKey)
			require.Nil(t, gotInfo)
			require.ErrorContains(t, err, "uniswapv4: getPoolAndPositionInfo: ")

			info, err := p.PositionInfo(ctx, nil, NewPositionInfoRequest(pm, tokenID))
			require.Nil(t, info)
			require.ErrorContains(t, err, "uniswapv4: positionInfo: ")

			byID, err := p.PoolKeys(ctx, nil, NewPoolKeysRequest(pm, [25]byte{}))
			require.Nil(t, byID)
			require.ErrorContains(t, err, "uniswapv4: poolKeys: ")

			subscriber, err := p.Subscriber(ctx, nil, NewSubscriberRequest(pm, tokenID))
			require.Nil(t, subscriber)
			require.ErrorContains(t, err, "uniswapv4: subscriber: ")

			gasLimit, err := p.UnsubscribeGasLimit(ctx, nil, NewUnsubscribeGasLimitRequest(pm))
			require.Nil(t, gasLimit)
			require.ErrorContains(t, err, "uniswapv4: unsubscribeGasLimit: ")

			nonces, err := p.Nonces(ctx, nil, NewNoncesRequest(pm, testSender, big.NewInt(0)))
			require.Nil(t, nonces)
			require.ErrorContains(t, err, "uniswapv4: nonces: ")

			tick, err := p.CallInitializePool(ctx, nil, NewInitializePoolRequest(pm, key, big.NewInt(1)))
			require.Nil(t, tick)
			require.ErrorContains(t, err, "uniswapv4: callInitializePool: ")

			results, err := p.CallMulticall(ctx, nil, NewMulticallRequest(pm, [][]byte{{0x01}}, nil))
			require.Nil(t, results)
			require.ErrorContains(t, err, "uniswapv4: callMulticall: ")

			permitData, err := p.CallPermit2Permit(ctx, nil, NewPermit2PermitRequest(pm, testSender, single, []byte{0x01}))
			require.Nil(t, permitData)
			require.ErrorContains(t, err, "uniswapv4: callPermit2Permit: ")

			batchData, err := p.CallPermit2PermitBatch(ctx, nil, NewPermit2PermitBatchRequest(pm, testSender, batch, []byte{0x01}))
			require.Nil(t, batchData)
			require.ErrorContains(t, err, "uniswapv4: callPermit2PermitBatch: ")
		})
	}
}

func TestPositionManagerMethodsRejectInvalidRequestWithoutRPC(t *testing.T) {
	pm := testPositionManager
	p := NewPositionManager(mocks.NewClient(t))
	ctx := t.Context()
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	key.Hooks = nil
	call := bind.NewCallOptions()
	opts := bind.NewTransactOptions(nil, tx.NewDynamicFeeBuilder())

	_, err := p.PoolManager(ctx, call, &PoolManagerRequest{})
	require.EqualError(t, err, "uniswapv4: poolManager: "+(&PoolManagerRequest{}).Validate().Error())
	_, err = p.WETH9(ctx, call, &WETH9Request{})
	require.EqualError(t, err, "uniswapv4: weth9: "+(&WETH9Request{}).Validate().Error())
	_, err = p.Permit2(ctx, call, &Permit2Request{})
	require.EqualError(t, err, "uniswapv4: permit2: "+(&Permit2Request{}).Validate().Error())
	_, err = p.NextTokenID(ctx, call, &NextTokenIDRequest{})
	require.EqualError(t, err, "uniswapv4: nextTokenId: "+(&NextTokenIDRequest{}).Validate().Error())
	_, err = p.GetPositionLiquidity(ctx, call, NewGetPositionLiquidityRequest(pm, nil))
	require.EqualError(t, err, "uniswapv4: getPositionLiquidity: "+NewGetPositionLiquidityRequest(pm, nil).Validate().Error())
	_, _, err = p.GetPoolAndPositionInfo(ctx, call, NewGetPoolAndPositionInfoRequest(nil, big.NewInt(1)))
	require.EqualError(t, err, "uniswapv4: getPoolAndPositionInfo: "+NewGetPoolAndPositionInfoRequest(nil, big.NewInt(1)).Validate().Error())
	_, err = p.PositionInfo(ctx, call, NewPositionInfoRequest(pm, nil))
	require.EqualError(t, err, "uniswapv4: positionInfo: "+NewPositionInfoRequest(pm, nil).Validate().Error())
	_, err = p.PoolKeys(ctx, call, &PoolKeysRequest{})
	require.EqualError(t, err, "uniswapv4: poolKeys: "+(&PoolKeysRequest{}).Validate().Error())
	_, err = p.Subscriber(ctx, call, NewSubscriberRequest(pm, nil))
	require.EqualError(t, err, "uniswapv4: subscriber: "+NewSubscriberRequest(pm, nil).Validate().Error())
	_, err = p.UnsubscribeGasLimit(ctx, call, &UnsubscribeGasLimitRequest{})
	require.EqualError(t, err, "uniswapv4: unsubscribeGasLimit: "+(&UnsubscribeGasLimitRequest{}).Validate().Error())
	_, err = p.Nonces(ctx, call, NewNoncesRequest(pm, nil, big.NewInt(0)))
	require.EqualError(t, err, "uniswapv4: nonces: "+NewNoncesRequest(pm, nil, big.NewInt(0)).Validate().Error())
	_, err = p.CallInitializePool(ctx, call, NewInitializePoolRequest(pm, key, big.NewInt(1)))
	require.EqualError(t, err, "uniswapv4: callInitializePool: invalid request: key.hooks is nil")
	_, err = p.CallMulticall(ctx, call, NewMulticallRequest(pm, nil, nil))
	require.EqualError(t, err, "uniswapv4: callMulticall: invalid request: data is empty")
	_, err = p.CallPermit2Permit(ctx, call, NewPermit2PermitRequest(pm, testSender, nil, []byte{0x01}))
	require.EqualError(t, err, "uniswapv4: callPermit2Permit: invalid request: permitSingle is nil")
	_, err = p.CallPermit2PermitBatch(ctx, call, NewPermit2PermitBatchRequest(pm, testSender, permit2.NewPermitBatch(nil, testHooks, big.NewInt(1)), []byte{0x01}))
	require.EqualError(t, err, "uniswapv4: callPermit2PermitBatch: invalid request: permitBatch.details is empty")

	_, _, err = p.ModifyLiquidities(ctx, opts, NewModifyLiquiditiesRequest(pm, []byte{ActionSettlePair}, nil, big.NewInt(1), nil))
	require.EqualError(t, err, "uniswapv4: modifyLiquidities: invalid request: params length 0 does not match actions length 1")
	_, _, err = p.ModifyLiquiditiesWithoutUnlock(ctx, opts, NewModifyLiquiditiesWithoutUnlockRequest(pm, nil, nil, nil))
	require.EqualError(t, err, "uniswapv4: modifyLiquiditiesWithoutUnlock: invalid request: actions is empty")
	_, _, err = p.InitializePool(ctx, opts, NewInitializePoolRequest(pm, nil, big.NewInt(1)))
	require.EqualError(t, err, "uniswapv4: initializePool: invalid request: key is nil")
	_, _, err = p.Multicall(ctx, opts, NewMulticallRequest(pm, [][]byte{{0x01}}, big.NewInt(-1)))
	require.EqualError(t, err, "uniswapv4: multicall: invalid request: nativeAmount is negative")
	_, _, err = p.Subscribe(ctx, opts, NewSubscribeRequest(pm, big.NewInt(1), nil, nil))
	require.EqualError(t, err, "uniswapv4: subscribe: invalid request: subscriber is nil")
	_, _, err = p.Unsubscribe(ctx, opts, NewUnsubscribeRequest(nil, big.NewInt(1)))
	require.EqualError(t, err, "uniswapv4: unsubscribe: invalid request: contract is nil")
	_, _, err = p.RevokeNonce(ctx, opts, NewRevokeNonceRequest(pm, nil))
	require.EqualError(t, err, "uniswapv4: revokeNonce: invalid request: nonce is nil")
	_, _, err = p.ERC721Permit(ctx, opts, NewERC721PermitRequest(pm, testHooks, big.NewInt(1), big.NewInt(1), big.NewInt(1), nil))
	require.EqualError(t, err, "uniswapv4: erc721Permit: invalid request: signature is empty")
	_, _, err = p.ERC721PermitForAll(ctx, opts, NewERC721PermitForAllRequest(pm, testSender, nil, true, big.NewInt(1), big.NewInt(1), []byte{0x01}))
	require.EqualError(t, err, "uniswapv4: erc721PermitForAll: invalid request: operator is nil")
	_, _, err = p.Permit2Permit(ctx, opts, &Permit2PermitRequest{})
	require.EqualError(t, err, "uniswapv4: permit2Permit: invalid request: contract is nil")
	_, _, err = p.Permit2PermitBatch(ctx, opts, NewPermit2PermitBatchRequest(pm, nil, nil, nil))
	require.EqualError(t, err, "uniswapv4: permit2PermitBatch: invalid request: owner is nil")
	require.ErrorIs(t, err, bind.ErrInvalidRequest)

	_, _, err = p.ModifyLiquidities(ctx, bind.NewTransactOptions(nil, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewModifyLiquiditiesRequest(pm, []byte{ActionSettlePair}, [][]byte{{0x01}}, big.NewInt(1), nil))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv4: modifyLiquidities: invalid options: value must be nil")
	_, _, err = p.ModifyLiquiditiesWithoutUnlock(ctx, bind.NewTransactOptions(nil, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(0)), NewModifyLiquiditiesWithoutUnlockRequest(pm, []byte{ActionSettlePair}, [][]byte{{0x01}}, nil))
	require.EqualError(t, err, "uniswapv4: modifyLiquiditiesWithoutUnlock: invalid options: value must be nil")
	_, _, err = p.Multicall(ctx, bind.NewTransactOptions(nil, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewMulticallRequest(pm, [][]byte{{0x01}}, big.NewInt(1)))
	require.EqualError(t, err, "uniswapv4: multicall: invalid options: value must be nil")

	_, _, err = p.ModifyLiquidities(ctx, nil, NewModifyLiquiditiesRequest(pm, []byte{ActionSettlePair}, [][]byte{{0x01}}, big.NewInt(1), big.NewInt(1)))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.ErrorContains(t, err, "uniswapv4: modifyLiquidities: ")
	_, _, err = p.Multicall(ctx, nil, NewMulticallRequest(pm, [][]byte{{0x01}}, nil))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)

	_, err = p.GetPositionLiquidity(ctx, nil, NewGetPositionLiquidityRequest(pm, big.NewInt(-1)))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)
	require.ErrorContains(t, err, "uniswapv4: getPositionLiquidity: ")
	_, err = p.CallInitializePool(ctx, nil, NewInitializePoolRequest(pm, NewPoolKey(testCurrency0, testCurrency1, big.NewInt(1<<24), big.NewInt(60), testHooks), big.NewInt(1)))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)
	_, _, err = p.RevokeNonce(ctx, opts, NewRevokeNonceRequest(pm, new(big.Int).Lsh(big.NewInt(1), 256)))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)
	require.ErrorContains(t, err, "uniswapv4: revokeNonce: ")
}

func TestPositionManagerWritesSignAndSend(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sender := core.PubkeyToAddress(key.PublicKey())
	pm := testPositionManager
	poolKey := NewPoolKey(types.ZeroAddress(), testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	actions := []byte{ActionMintPosition, ActionSettlePair, ActionSweep}
	params := [][]byte{{0x01}, {0x02, 0x03}, {}}
	deadline := big.NewInt(1_700_000_000)
	native := big.NewInt(1_000_000_007)
	sig := []byte{0x0a, 0x0b, 0x0c}
	single := permit2.NewPermitSingle(permit2.NewPermitDetails(testCurrency0, big.NewInt(1), big.NewInt(2), big.NewInt(3)), testHooks, big.NewInt(4))
	batch := permit2.NewPermitBatch([]permit2.PermitDetails{*permit2.NewPermitDetails(testCurrency0, big.NewInt(1), big.NewInt(2), big.NewInt(3)), *permit2.NewPermitDetails(testCurrency1, big.NewInt(5), big.NewInt(6), big.NewInt(7))}, testHooks, big.NewInt(8))

	modifyData, err := EncodeModifyLiquidities(actions, params, deadline)
	require.NoError(t, err)
	withoutUnlockData, err := EncodeModifyLiquiditiesWithoutUnlock(actions, params)
	require.NoError(t, err)
	initializeData, err := EncodeInitializePool(poolKey, new(big.Int).Lsh(big.NewInt(1), 96))
	require.NoError(t, err)
	multicallData, err := EncodeMulticall([][]byte{withoutUnlockData, initializeData})
	require.NoError(t, err)
	subscribeData, err := EncodeSubscribe(big.NewInt(5), testHooks, []byte{0xff})
	require.NoError(t, err)
	unsubscribeData, err := EncodeUnsubscribe(big.NewInt(5))
	require.NoError(t, err)
	revokeData, err := EncodeRevokeNonce(big.NewInt(1 << 40))
	require.NoError(t, err)
	permitData, err := EncodeERC721Permit(testHooks, big.NewInt(5), deadline, big.NewInt(2), sig)
	require.NoError(t, err)
	permitForAllData, err := EncodeERC721PermitForAll(sender, testHooks, true, deadline, big.NewInt(3), sig)
	require.NoError(t, err)
	permit2PermitData, err := EncodePermit2Permit(sender, single, sig)
	require.NoError(t, err)
	permit2PermitBatchData, err := EncodePermit2PermitBatch(sender, batch, sig)
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		value    *big.Int
		send     func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error)
	}{
		{"modifyLiquidities", modifyData, native, func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return p.ModifyLiquidities(t.Context(), opts, NewModifyLiquiditiesRequest(pm, actions, params, deadline, native))
		}},
		{"modifyLiquidities without native", modifyData, nil, func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return p.ModifyLiquidities(t.Context(), opts, NewModifyLiquiditiesRequest(pm, actions, params, deadline, nil))
		}},
		{"modifyLiquiditiesWithoutUnlock", withoutUnlockData, native, func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return p.ModifyLiquiditiesWithoutUnlock(t.Context(), opts, NewModifyLiquiditiesWithoutUnlockRequest(pm, actions, params, native))
		}},
		{"initializePool", initializeData, nil, func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return p.InitializePool(t.Context(), opts, NewInitializePoolRequest(pm, poolKey, new(big.Int).Lsh(big.NewInt(1), 96)))
		}},
		{"multicall", multicallData, native, func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return p.Multicall(t.Context(), opts, NewMulticallRequest(pm, [][]byte{withoutUnlockData, initializeData}, native))
		}},
		{"multicall without native", multicallData, nil, func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return p.Multicall(t.Context(), opts, NewMulticallRequest(pm, [][]byte{withoutUnlockData, initializeData}, nil))
		}},
		{"subscribe", subscribeData, nil, func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return p.Subscribe(t.Context(), opts, NewSubscribeRequest(pm, big.NewInt(5), testHooks, []byte{0xff}))
		}},
		{"unsubscribe", unsubscribeData, nil, func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return p.Unsubscribe(t.Context(), opts, NewUnsubscribeRequest(pm, big.NewInt(5)))
		}},
		{"revokeNonce", revokeData, nil, func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return p.RevokeNonce(t.Context(), opts, NewRevokeNonceRequest(pm, big.NewInt(1<<40)))
		}},
		{"erc721Permit", permitData, nil, func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return p.ERC721Permit(t.Context(), opts, NewERC721PermitRequest(pm, testHooks, big.NewInt(5), deadline, big.NewInt(2), sig))
		}},
		{"erc721PermitForAll", permitForAllData, nil, func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return p.ERC721PermitForAll(t.Context(), opts, NewERC721PermitForAllRequest(pm, sender, testHooks, true, deadline, big.NewInt(3), sig))
		}},
		{"permit2Permit", permit2PermitData, nil, func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return p.Permit2Permit(t.Context(), opts, NewPermit2PermitRequest(pm, sender, single, sig))
		}},
		{"permit2PermitBatch", permit2PermitBatchData, nil, func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return p.Permit2PermitBatch(t.Context(), opts, NewPermit2PermitBatchRequest(pm, sender, batch, sig))
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
			signed, hash, err := tc.send(NewPositionManager(c), opts)
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

func TestPositionManagerWritesMapEstimateRevertToSentinel(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	pm := testPositionManager
	poolKey := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)

	for _, tc := range []struct {
		name string
		err  *abi.Error
		args []any
		raw  jsontext.Value
		send func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error)
		want error
	}{
		{"modifyLiquidities deadline passed", errDeadlinePassed, []any{big.NewInt(1)}, nil, func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return p.ModifyLiquidities(t.Context(), opts, NewModifyLiquiditiesRequest(pm, []byte{ActionSettlePair}, [][]byte{{0x01}}, big.NewInt(1), nil))
		}, ErrDeadlinePassed},
		{"modifyLiquidities currency not settled", errCurrencyNotSettled, nil, nil, func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return p.ModifyLiquidities(t.Context(), opts, NewModifyLiquiditiesRequest(pm, []byte{ActionMintPosition}, [][]byte{{0x01}}, big.NewInt(1), big.NewInt(1)))
		}, ErrCurrencyNotSettled},
		{"modifyLiquiditiesWithoutUnlock locked", errManagerLocked, nil, nil, func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return p.ModifyLiquiditiesWithoutUnlock(t.Context(), opts, NewModifyLiquiditiesWithoutUnlockRequest(pm, []byte{ActionSettlePair}, [][]byte{{0x01}}, nil))
		}, ErrManagerLocked},
		{"initializePool tick spacing", errTickSpacingTooSmall, []any{big.NewInt(0)}, nil, func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return p.InitializePool(t.Context(), opts, NewInitializePoolRequest(pm, poolKey, big.NewInt(1)))
		}, ErrTickSpacingTooSmall},
		{"subscribe no code", errNoCodeSubscriber, nil, nil, func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return p.Subscribe(t.Context(), opts, NewSubscribeRequest(pm, big.NewInt(1), testHooks, nil))
		}, ErrNoCodeSubscriber},
		{"unsubscribe not subscribed", errNotSubscribed, nil, nil, func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return p.Unsubscribe(t.Context(), opts, NewUnsubscribeRequest(pm, big.NewInt(1)))
		}, ErrNotSubscribed},
		{"revokeNonce used", errNonceAlreadyUsed, nil, nil, func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return p.RevokeNonce(t.Context(), opts, NewRevokeNonceRequest(pm, big.NewInt(1)))
		}, ErrNonceAlreadyUsed},
		{"erc721Permit expired", errSignatureDeadlineExpired, nil, nil, func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return p.ERC721Permit(t.Context(), opts, NewERC721PermitRequest(pm, testHooks, big.NewInt(1), big.NewInt(1), big.NewInt(1), []byte{0x01}))
		}, ErrSignatureDeadlineExpired},
		{"erc721PermitForAll bad signer", errInvalidSigner, nil, nil, func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return p.ERC721PermitForAll(t.Context(), opts, NewERC721PermitForAllRequest(pm, testSender, testHooks, true, big.NewInt(1), big.NewInt(1), []byte{0x01}))
		}, permit2.ErrInvalidSigner},
		{"multicall unknown", nil, nil, jsontext.Value(`"0x08c379a0"`), func(p *PositionManager, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return p.Multicall(t.Context(), opts, NewMulticallRequest(pm, [][]byte{{0x01}}, nil))
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

			signed, hash, err := tc.send(NewPositionManager(c), bind.NewTransactOptions(key, builder))
			require.ErrorIs(t, err, tc.want)
			require.ErrorContains(t, err, "uniswapv4: ")
			require.Nil(t, signed)
			require.Nil(t, hash)
		})
	}
}

func TestPositionManagerWritesReturnSignedTxOnBroadcastFailure(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(9)
	builder.Limit(6_000_000)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(errors.New("dial tcp: connection refused")).Once()

	signed, hash, err := NewPositionManager(c).ModifyLiquidities(t.Context(), bind.NewTransactOptions(key, builder), NewModifyLiquiditiesRequest(testPositionManager, []byte{ActionSettlePair}, [][]byte{{0x01}}, big.NewInt(1), big.NewInt(5)))
	require.ErrorContains(t, err, "uniswapv4: modifyLiquidities: ")
	require.ErrorContains(t, err, "connection refused")
	require.NotErrorIs(t, err, ErrUnknown)
	require.Nil(t, hash)
	require.NotNil(t, signed)
	require.Equal(t, uint64(9), signed.AddressNonce())
	require.Zero(t, big.NewInt(5).Cmp(signed.(*core.DynamicFeeTx).Value))
}

func TestPositionManagerReadsPassThroughTransportError(t *testing.T) {
	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(errors.New("dial tcp: connection refused")).Once()

	weth9, err := NewPositionManager(c).WETH9(t.Context(), nil, NewWETH9Request(testPositionManager))
	require.Nil(t, weth9)
	require.EqualError(t, err, "uniswapv4: weth9: dial tcp: connection refused")
	require.NotErrorIs(t, err, ErrUnknown)
}

func TestPositionManagerCallsPropagateContext(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(t.Context(), marker{}, "positionmanager")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	word, err := abi.Pack(abi.NewTypes(abi.Address), testSender)
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(got context.Context, e rpc.Element) error {
		require.Equal(t, "positionmanager", got.Value(marker{}))
		if e.Method == rpc.MethodETHCall {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(word)
		}
		return nil
	}).Twice()

	_, err = NewPositionManager(c).Subscriber(ctx, nil, NewSubscriberRequest(testPositionManager, big.NewInt(1)))
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(1)
	builder.Limit(3_000_000)
	builder.Price(big.NewInt(2), big.NewInt(1))
	_, _, err = NewPositionManager(c).RevokeNonce(ctx, bind.NewTransactOptions(key, builder), NewRevokeNonceRequest(testPositionManager, big.NewInt(3)))
	require.NoError(t, err)
}

func TestPositionManagerPromotesERC721(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	tokenID := big.NewInt(4)
	ownerWord, err := abi.Pack(abi.NewTypes(abi.Address), testSender)
	require.NoError(t, err)
	nameWord, err := abi.Pack(abi.NewTypes(abi.String), "Uniswap v4 Positions NFT")
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
		switch e.Method {
		case rpc.MethodETHCall:
			call := e.Params.([]any)[0].(map[string]any)
			require.Equal(t, testPositionManager.String(), call["to"])
			data, err := encoding.Hex.Decode(call["data"].(string))
			require.NoError(t, err)
			nameSel := abi.NewFunction("name", nil, nil).Selector()
			if [4]byte(data[:4]) == nameSel {
				*e.Result.(*string) = encoding.Hex.EncodePrefixed(nameWord)
			} else {
				*e.Result.(*string) = encoding.Hex.EncodePrefixed(ownerWord)
			}
		case rpc.MethodETHSendRawTransaction:
			raw, err := encoding.Hex.Decode(e.Params.([]any)[0].(string))
			require.NoError(t, err)
			decoded, err := core.DecodeTransaction(raw)
			require.NoError(t, err)
			require.True(t, decoded.(*core.DynamicFeeTx).To.Equal(testPositionManager))
			*e.Result.(*string) = core.Keccak256(raw).String()
		default:
			t.Fatalf("unexpected method %s", e.Method)
		}
		return nil
	}).Times(3)

	var p IPositionManager = NewPositionManager(c)
	ctx := t.Context()

	owner, err := p.OwnerOf(ctx, nil, erc721.NewOwnerOfRequest(testPositionManager, tokenID))
	require.NoError(t, err)
	require.True(t, owner.Equal(testSender))

	name, err := p.Name(ctx, nil, erc721.NewNameRequest(testPositionManager))
	require.NoError(t, err)
	require.Equal(t, "Uniswap v4 Positions NFT", name)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(1)
	builder.Limit(100_000)
	builder.Price(big.NewInt(2), big.NewInt(1))

	_, hash, err := p.Approve(ctx, bind.NewTransactOptions(key, builder), erc721.NewApproveRequest(testPositionManager, testHooks, tokenID))
	require.NoError(t, err)
	require.NotNil(t, hash)
}
