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
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

func TestQuoterReadsSendCallAndDecode(t *testing.T) {
	quoter := testPoolManager
	from := testSender
	opts := bind.NewCallOptions().WithFrom(from).WithBlock("0x2a")
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	single := NewQuoteExactSingleParams(key, true, big.NewInt(1_000_000), []byte{0x01})
	multi := NewQuoteExactParams(testCurrency0, []*PathKey{
		NewPathKey(testCurrency1, big.NewInt(500), big.NewInt(10), testHooks, nil),
		NewPathKey(testHooks, big.NewInt(3000), big.NewInt(60), testHooks, []byte{0x02}),
	}, big.NewInt(2_000_000))
	maxUint256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))

	addressWord, err := abi.Pack(abi.NewTypes(abi.Address), testHooks)
	require.NoError(t, err)
	zeroWord, err := abi.Pack(abi.NewTypes(abi.Address), testSender)
	require.NoError(t, err)
	quoteWord, err := abi.Pack(abi.NewTypes(abi.Uint256, abi.Uint256), maxUint256, big.NewInt(84_000))
	require.NoError(t, err)

	inSingleData, err := EncodeQuoteExactInputSingle(single)
	require.NoError(t, err)
	inData, err := EncodeQuoteExactInput(multi)
	require.NoError(t, err)
	outSingleData, err := EncodeQuoteExactOutputSingle(single)
	require.NoError(t, err)
	outData, err := EncodeQuoteExactOutput(multi)
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		result   []byte
		call     func(q *Quoter) (any, error)
		want     any
	}{
		{"poolManager", EncodeQuoterPoolManager(), addressWord, func(q *Quoter) (any, error) {
			return q.PoolManager(t.Context(), opts, NewPoolManagerRequest(quoter))
		}, testHooks},
		{"msgSender", EncodeQuoterMsgSender(), zeroWord, func(q *Quoter) (any, error) {
			return q.MsgSender(t.Context(), opts, NewMsgSenderRequest(quoter))
		}, testSender},
		{"callQuoteExactInputSingle", inSingleData, quoteWord, func(q *Quoter) (any, error) {
			return q.CallQuoteExactInputSingle(t.Context(), opts, NewQuoteExactInputSingleRequest(quoter, single))
		}, NewQuoteExactInputResult(maxUint256, big.NewInt(84_000))},
		{"callQuoteExactInput", inData, quoteWord, func(q *Quoter) (any, error) {
			return q.CallQuoteExactInput(t.Context(), opts, NewQuoteExactInputRequest(quoter, multi))
		}, NewQuoteExactInputResult(maxUint256, big.NewInt(84_000))},
		{"callQuoteExactOutputSingle", outSingleData, quoteWord, func(q *Quoter) (any, error) {
			return q.CallQuoteExactOutputSingle(t.Context(), opts, NewQuoteExactOutputSingleRequest(quoter, single))
		}, NewQuoteExactOutputResult(maxUint256, big.NewInt(84_000))},
		{"callQuoteExactOutput", outData, quoteWord, func(q *Quoter) (any, error) {
			return q.CallQuoteExactOutput(t.Context(), opts, NewQuoteExactOutputRequest(quoter, multi))
		}, NewQuoteExactOutputResult(maxUint256, big.NewInt(84_000))},
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
				require.Equal(t, quoter.String(), call["to"])
				require.Equal(t, from.String(), call["from"])
				require.Equal(t, encoding.Hex.EncodePrefixed(tc.calldata), call["data"])
				require.NotContains(t, call, "value")
				require.Equal(t, "0x2a", params[1])
				*e.Result.(*string) = encoding.Hex.EncodePrefixed(tc.result)
				return nil
			}).Once()

			got, err := tc.call(NewQuoter(c))
			require.NoError(t, err)
			require.Equal(t, fmt.Sprint(tc.want), fmt.Sprint(got))
		})
	}
}

func TestQuoterReadsDefaultToLatestWithoutFrom(t *testing.T) {
	word, err := abi.Pack(abi.NewTypes(abi.Uint256, abi.Uint256), big.NewInt(9), big.NewInt(1))
	require.NoError(t, err)
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)

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

	quote, err := NewQuoter(c).CallQuoteExactInputSingle(t.Context(), nil, NewQuoteExactInputSingleRequest(testPoolManager, NewQuoteExactSingleParams(key, false, big.NewInt(5), nil)))
	require.NoError(t, err)
	require.Zero(t, big.NewInt(9).Cmp(quote.AmountOut))
	require.Zero(t, big.NewInt(1).Cmp(quote.GasEstimate))
}

func TestQuoterCallsMapRevertToSentinel(t *testing.T) {
	quoter := testPoolManager
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)

	poolNotInitializedSel := errPoolNotInitialized.Selector()
	wrappedPoolNotInitialized, err := abi.Pack(errUnexpectedRevertBytes.Inputs, poolNotInitializedSel[:])
	require.NoError(t, err)
	zeroSel := errSwapAmountCannotBeZero.Selector()
	wrappedZero, err := abi.Pack(errUnexpectedRevertBytes.Inputs, zeroSel[:])
	require.NoError(t, err)
	wrappedUnknown, err := abi.Pack(errUnexpectedRevertBytes.Inputs, []byte{0xde, 0xad, 0xbe, 0xef})
	require.NoError(t, err)

	for _, tc := range []struct {
		name   string
		err    *abi.Error
		args   []any
		packed []byte
		raw    jsontext.Value
		want   []error
	}{
		{"NotEnoughLiquidity", errNotEnoughLiquidity, []any{make([]byte, 32)}, nil, nil, []error{ErrNotEnoughLiquidity}},
		{"NotSelf", errNotSelf, nil, nil, nil, []error{ErrNotSelf}},
		{"UnexpectedCallSuccess", errUnexpectedCallSuccess, nil, nil, nil, []error{ErrUnexpectedCallSuccess}},
		{"UnexpectedRevertBytes(PoolNotInitialized)", errUnexpectedRevertBytes, nil, wrappedPoolNotInitialized, nil, []error{ErrUnexpectedRevertBytes, ErrPoolNotInitialized}},
		{"UnexpectedRevertBytes(SwapAmountCannotBeZero)", errUnexpectedRevertBytes, nil, wrappedZero, nil, []error{ErrUnexpectedRevertBytes, ErrSwapAmountCannotBeZero}},
		{"UnexpectedRevertBytes(unknown)", errUnexpectedRevertBytes, nil, wrappedUnknown, nil, []error{ErrUnexpectedRevertBytes}},
		{"PoolManager PoolNotInitialized", errPoolNotInitialized, nil, nil, nil, []error{ErrPoolNotInitialized}},
		{"PoolManager ManagerLocked", errManagerLocked, nil, nil, nil, []error{ErrManagerLocked}},
		{"PoolManager TickMisaligned", errTickMisaligned, []any{big.NewInt(61), big.NewInt(60)}, nil, nil, []error{ErrTickMisaligned}},
		{"Error(string)", nil, nil, nil, jsontext.Value(`"0x08c379a00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000c554e415554484f52495a45440000000000000000000000000000000000000000"`), []error{ErrUnknown}},
		{"unknown selector", nil, nil, nil, jsontext.Value(`"0xdeadbeef"`), []error{ErrUnknown}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := tc.raw
			if tc.err != nil {
				sel := tc.err.Selector()
				packed := tc.packed
				if packed == nil {
					var err error
					packed, err = abi.Pack(tc.err.Inputs, tc.args...)
					require.NoError(t, err)
				}
				data = jsontext.Value(`"` + encoding.Hex.EncodePrefixed(append(sel[:], packed...)) + `"`)
			}
			rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: data}
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).Return(rpcErr).Once()

			quote, err := NewQuoter(c).CallQuoteExactOutputSingle(t.Context(), nil, NewQuoteExactOutputSingleRequest(quoter, NewQuoteExactSingleParams(key, true, big.NewInt(1), nil)))
			require.Nil(t, quote)
			require.ErrorContains(t, err, "uniswapv4: callQuoteExactOutputSingle: ")
			for _, want := range tc.want {
				require.ErrorIs(t, err, want)
			}
			if !errors.Is(tc.want[0], ErrUnknown) {
				require.NotErrorIs(t, err, ErrUnknown)
			}
			got, ok := errors.AsType[*rpc.ResponseError](err)
			require.True(t, ok)
			require.Same(t, rpcErr, got)
		})
	}
}

func TestQuoterReadsRejectMalformedResult(t *testing.T) {
	quoter := testPoolManager
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	single := NewQuoteExactSingleParams(key, true, big.NewInt(1), nil)
	multi := NewQuoteExactParams(testCurrency0, nil, big.NewInt(1))

	for _, result := range []string{"0x", "0xzz", "0x01", "not hex"} {
		t.Run(result, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				*e.Result.(*string) = result
				return nil
			}).Times(6)
			q := NewQuoter(c)
			ctx := t.Context()

			poolManager, err := q.PoolManager(ctx, nil, NewPoolManagerRequest(quoter))
			require.Nil(t, poolManager)
			require.ErrorContains(t, err, "uniswapv4: quoterPoolManager: ")

			msgSender, err := q.MsgSender(ctx, nil, NewMsgSenderRequest(quoter))
			require.Nil(t, msgSender)
			require.ErrorContains(t, err, "uniswapv4: quoterMsgSender: ")

			inSingle, err := q.CallQuoteExactInputSingle(ctx, nil, NewQuoteExactInputSingleRequest(quoter, single))
			require.Nil(t, inSingle)
			require.ErrorContains(t, err, "uniswapv4: callQuoteExactInputSingle: ")

			in, err := q.CallQuoteExactInput(ctx, nil, NewQuoteExactInputRequest(quoter, multi))
			require.Nil(t, in)
			require.ErrorContains(t, err, "uniswapv4: callQuoteExactInput: ")

			outSingle, err := q.CallQuoteExactOutputSingle(ctx, nil, NewQuoteExactOutputSingleRequest(quoter, single))
			require.Nil(t, outSingle)
			require.ErrorContains(t, err, "uniswapv4: callQuoteExactOutputSingle: ")

			out, err := q.CallQuoteExactOutput(ctx, nil, NewQuoteExactOutputRequest(quoter, multi))
			require.Nil(t, out)
			require.ErrorContains(t, err, "uniswapv4: callQuoteExactOutput: ")
		})
	}
}

func TestQuoterMethodsRejectInvalidRequestWithoutRPC(t *testing.T) {
	quoter := testPoolManager
	q := NewQuoter(mocks.NewClient(t))
	ctx := t.Context()
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)

	_, err := q.PoolManager(ctx, nil, &PoolManagerRequest{})
	require.EqualError(t, err, "uniswapv4: quoterPoolManager: invalid request: contract is nil")
	_, err = q.MsgSender(ctx, nil, &MsgSenderRequest{})
	require.EqualError(t, err, "uniswapv4: quoterMsgSender: invalid request: contract is nil")
	_, err = q.CallQuoteExactInputSingle(ctx, nil, NewQuoteExactInputSingleRequest(quoter, NewQuoteExactSingleParams(nil, true, big.NewInt(1), nil)))
	require.EqualError(t, err, "uniswapv4: callQuoteExactInputSingle: invalid request: params.poolKey is nil")
	_, err = q.CallQuoteExactInput(ctx, nil, NewQuoteExactInputRequest(quoter, NewQuoteExactParams(testCurrency0, []*PathKey{nil}, big.NewInt(1))))
	require.EqualError(t, err, "uniswapv4: callQuoteExactInput: invalid request: params.path[0] is nil")
	_, err = q.CallQuoteExactOutputSingle(ctx, nil, NewQuoteExactOutputSingleRequest(quoter, NewQuoteExactSingleParams(key, true, nil, nil)))
	require.EqualError(t, err, "uniswapv4: callQuoteExactOutputSingle: invalid request: params.exactAmount is nil")
	_, err = q.CallQuoteExactOutput(ctx, nil, NewQuoteExactOutputRequest(nil, NewQuoteExactParams(testCurrency0, nil, big.NewInt(1))))
	require.EqualError(t, err, "uniswapv4: callQuoteExactOutput: invalid request: contract is nil")
	require.ErrorIs(t, err, bind.ErrInvalidRequest)

	_, err = q.CallQuoteExactInputSingle(ctx, nil, NewQuoteExactInputSingleRequest(quoter, NewQuoteExactSingleParams(key, true, new(big.Int).Lsh(big.NewInt(1), 128), nil)))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)
	require.ErrorContains(t, err, "uniswapv4: callQuoteExactInputSingle: ")
	_, err = q.CallQuoteExactOutput(ctx, nil, NewQuoteExactOutputRequest(quoter, NewQuoteExactParams(testCurrency0, []*PathKey{NewPathKey(testCurrency1, big.NewInt(1<<24), big.NewInt(60), testHooks, nil)}, big.NewInt(1))))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)
}

func TestQuoterReadsPassThroughTransportError(t *testing.T) {
	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(errors.New("dial tcp: connection refused")).Once()

	msgSender, err := NewQuoter(c).MsgSender(t.Context(), nil, NewMsgSenderRequest(testPoolManager))
	require.Nil(t, msgSender)
	require.EqualError(t, err, "uniswapv4: quoterMsgSender: dial tcp: connection refused")
	require.NotErrorIs(t, err, ErrUnknown)
}

func TestQuoterCallsPropagateContext(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(t.Context(), marker{}, "quoter")
	word, err := abi.Pack(abi.NewTypes(abi.Address), testHooks)
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(got context.Context, e rpc.Element) error {
		require.Equal(t, "quoter", got.Value(marker{}))
		*e.Result.(*string) = encoding.Hex.EncodePrefixed(word)
		return nil
	}).Once()

	poolManager, err := NewQuoter(c).PoolManager(ctx, nil, NewPoolManagerRequest(testPoolManager))
	require.NoError(t, err)
	require.True(t, poolManager.Equal(testHooks))
}

func TestQuoterSatisfiesInterface(t *testing.T) {
	word, err := abi.Pack(abi.NewTypes(abi.Uint256, abi.Uint256), big.NewInt(3), big.NewInt(4))
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
		*e.Result.(*string) = encoding.Hex.EncodePrefixed(word)
		return nil
	}).Once()

	var q IQuoter = NewQuoter(c)
	quote, err := q.CallQuoteExactOutput(t.Context(), nil, NewQuoteExactOutputRequest(testPoolManager, NewQuoteExactParams(testCurrency0, nil, big.NewInt(1))))
	require.NoError(t, err)
	require.Zero(t, big.NewInt(3).Cmp(quote.AmountIn))
	require.Zero(t, big.NewInt(4).Cmp(quote.GasEstimate))
}
