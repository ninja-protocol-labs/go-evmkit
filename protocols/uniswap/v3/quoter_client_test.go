package v3

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

func TestQuoterReadsSendCallAndDecode(t *testing.T) {
	quoter := types.MustNewAddressFromHex("0x61fFE014bA17989E743c5F6cB21bF9697530B21e")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	c := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	opts := bind.NewCallOptions().WithFrom(a).WithBlock("0x1b4")
	path, err := EncodePath([]*types.Address{a, b, c}, []*big.Int{big.NewInt(500), big.NewInt(3000)})
	require.NoError(t, err)

	factoryResult, err := quoterFactoryFn.EncodeReturn(b)
	require.NoError(t, err)
	weth9Result, err := quoterWETH9Fn.EncodeReturn(c)
	require.NoError(t, err)
	singleResult, err := quoteExactInputSingleFn.EncodeReturn(big.NewInt(990), new(big.Int).Lsh(big.NewInt(1), 96), uint32(2), big.NewInt(80000))
	require.NoError(t, err)
	multiResult, err := quoteExactInputFn.EncodeReturn(big.NewInt(980), []*big.Int{big.NewInt(11), big.NewInt(22)}, []uint32{1, 3}, big.NewInt(150000))
	require.NoError(t, err)
	outSingleResult, err := quoteExactOutputSingleFn.EncodeReturn(big.NewInt(1010), big.NewInt(33), uint32(0), big.NewInt(70000))
	require.NoError(t, err)
	outMultiResult, err := quoteExactOutputFn.EncodeReturn(big.NewInt(1020), []*big.Int{big.NewInt(44), big.NewInt(55)}, []uint32{0, 4}, big.NewInt(160000))
	require.NoError(t, err)

	inSingleData, err := EncodeQuoteExactInputSingle(a, b, big.NewInt(1000), big.NewInt(3000), big.NewInt(0))
	require.NoError(t, err)
	inData, err := EncodeQuoteExactInput(path, big.NewInt(1000))
	require.NoError(t, err)
	outSingleData, err := EncodeQuoteExactOutputSingle(a, b, big.NewInt(990), big.NewInt(3000), big.NewInt(0))
	require.NoError(t, err)
	outData, err := EncodeQuoteExactOutput(path, big.NewInt(990))
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		result   []byte
		call     func(q *Quoter) (any, error)
		want     any
	}{
		{
			"factory",
			EncodeQuoterFactory(),
			factoryResult,
			func(q *Quoter) (any, error) {
				return q.Factory(t.Context(), opts, NewFactoryRequest(quoter))
			},
			b,
		},
		{
			"weth9",
			EncodeQuoterWETH9(),
			weth9Result,
			func(q *Quoter) (any, error) {
				return q.WETH9(t.Context(), opts, NewWETH9Request(quoter))
			},
			c,
		},
		{
			"callQuoteExactInputSingle",
			inSingleData,
			singleResult,
			func(q *Quoter) (any, error) {
				return q.CallQuoteExactInputSingle(t.Context(), opts, NewQuoteExactInputSingleRequest(quoter, NewQuoteExactInputSingleParams(a, b, big.NewInt(1000), big.NewInt(3000), big.NewInt(0))))
			},
			NewQuoteExactInputSingleResult(big.NewInt(990), new(big.Int).Lsh(big.NewInt(1), 96), big.NewInt(2), big.NewInt(80000)),
		},
		{
			"callQuoteExactInput",
			inData,
			multiResult,
			func(q *Quoter) (any, error) {
				return q.CallQuoteExactInput(t.Context(), opts, NewQuoteExactInputRequest(quoter, path, big.NewInt(1000)))
			},
			NewQuoteExactInputResult(big.NewInt(980), []*big.Int{big.NewInt(11), big.NewInt(22)}, []*big.Int{big.NewInt(1), big.NewInt(3)}, big.NewInt(150000)),
		},
		{
			"callQuoteExactOutputSingle",
			outSingleData,
			outSingleResult,
			func(q *Quoter) (any, error) {
				return q.CallQuoteExactOutputSingle(t.Context(), opts, NewQuoteExactOutputSingleRequest(quoter, NewQuoteExactOutputSingleParams(a, b, big.NewInt(990), big.NewInt(3000), big.NewInt(0))))
			},
			NewQuoteExactOutputSingleResult(big.NewInt(1010), big.NewInt(33), big.NewInt(0), big.NewInt(70000)),
		},
		{
			"callQuoteExactOutput",
			outData,
			outMultiResult,
			func(q *Quoter) (any, error) {
				return q.CallQuoteExactOutput(t.Context(), opts, NewQuoteExactOutputRequest(quoter, path, big.NewInt(990)))
			},
			NewQuoteExactOutputResult(big.NewInt(1020), []*big.Int{big.NewInt(44), big.NewInt(55)}, []*big.Int{big.NewInt(0), big.NewInt(4)}, big.NewInt(160000)),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cli := mocks.NewClient(t)
			cli.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				require.Equal(t, rpc.MethodETHCall, e.Method)
				params, ok := e.Params.([]any)
				require.True(t, ok)
				require.Len(t, params, 2)
				call, ok := params[0].(map[string]any)
				require.True(t, ok)
				require.Equal(t, quoter.String(), call["to"])
				require.Equal(t, a.String(), call["from"])
				require.Equal(t, encoding.Hex.EncodePrefixed(tc.calldata), call["data"])
				require.NotContains(t, call, "value")
				require.Equal(t, "0x1b4", params[1])
				*e.Result.(*string) = encoding.Hex.EncodePrefixed(tc.result)
				return nil
			}).Once()

			got, err := tc.call(NewQuoter(cli))
			require.NoError(t, err)
			require.Equal(t, fmt.Sprint(tc.want), fmt.Sprint(got))
		})
	}
}

func TestQuoterReadsMapRevertToSentinel(t *testing.T) {
	quoter := types.MustNewAddressFromHex("0x61fFE014bA17989E743c5F6cB21bF9697530B21e")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	errorSel := []byte{0x08, 0xc3, 0x79, 0xa0}

	for _, tc := range []struct {
		name   string
		reason string
		want   error
	}{
		{"locked", reasonLocked, ErrLocked},
		{"amount specified zero", reasonAmountSpecifiedZero, ErrAmountSpecifiedZero},
		{"invalid sqrt price limit", reasonInvalidSqrtPriceLimit, ErrInvalidSqrtPriceLimit},
		{"unexpected error", "Unexpected error", ErrReverted},
	} {
		t.Run(tc.name, func(t *testing.T) {
			packed, err := abi.Pack(abi.NewTypes(abi.String), tc.reason)
			require.NoError(t, err)
			revert := append(append([]byte{}, errorSel...), packed...)

			cli := mocks.NewClient(t)
			cli.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(revert) + `"`)}).Once()

			quote, err := NewQuoter(cli).CallQuoteExactInputSingle(t.Context(), bind.NewCallOptions().WithFrom(a), NewQuoteExactInputSingleRequest(quoter, NewQuoteExactInputSingleParams(a, b, big.NewInt(1), big.NewInt(3000), big.NewInt(0))))
			require.Nil(t, quote)
			require.ErrorIs(t, err, tc.want)
			require.ErrorContains(t, err, "uniswapv3: callQuoteExactInputSingle: ")
			var rpcErr *rpc.ResponseError
			require.ErrorAs(t, err, &rpcErr)
		})
	}

	for _, data := range []jsontext.Value{nil, jsontext.Value(`"0x"`), jsontext.Value(`{"x":1}`)} {
		cli := mocks.NewClient(t)
		cli.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: data}).Once()

		path, err := EncodePath([]*types.Address{a, b}, []*big.Int{big.NewInt(3000)})
		require.NoError(t, err)
		quote, err := NewQuoter(cli).CallQuoteExactOutput(t.Context(), nil, NewQuoteExactOutputRequest(quoter, path, big.NewInt(1)))
		require.Nil(t, quote)
		require.ErrorIs(t, err, ErrReverted)
		require.ErrorContains(t, err, "uniswapv3: callQuoteExactOutput: reverted: ")
	}

	cli := mocks.NewClient(t)
	cli.EXPECT().Call(mock.Anything, mock.Anything).Return(errors.New("dial tcp: connection refused")).Once()

	factory, err := NewQuoter(cli).Factory(t.Context(), nil, NewFactoryRequest(quoter))
	require.Nil(t, factory)
	require.EqualError(t, err, "uniswapv3: factory: dial tcp: connection refused")
	require.NotErrorIs(t, err, ErrReverted)
}

func TestQuoterReadsRejectMalformedResult(t *testing.T) {
	quoter := types.MustNewAddressFromHex("0x61fFE014bA17989E743c5F6cB21bF9697530B21e")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	for _, result := range []string{"0x", "0xzz", "0x01", "0x00", "not hex"} {
		t.Run(result, func(t *testing.T) {
			cli := mocks.NewClient(t)
			cli.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				*e.Result.(*string) = result
				return nil
			}).Times(3)
			q := NewQuoter(cli)
			path, err := EncodePath([]*types.Address{a, b}, []*big.Int{big.NewInt(3000)})
			require.NoError(t, err)

			weth9, err := q.WETH9(t.Context(), nil, NewWETH9Request(quoter))
			require.Nil(t, weth9)
			require.ErrorContains(t, err, "uniswapv3: weth9: ")

			single, err := q.CallQuoteExactOutputSingle(t.Context(), nil, NewQuoteExactOutputSingleRequest(quoter, NewQuoteExactOutputSingleParams(a, b, big.NewInt(1), big.NewInt(3000), big.NewInt(0))))
			require.Nil(t, single)
			require.ErrorContains(t, err, "uniswapv3: callQuoteExactOutputSingle: ")

			multi, err := q.CallQuoteExactInput(t.Context(), nil, NewQuoteExactInputRequest(quoter, path, big.NewInt(1)))
			require.Nil(t, multi)
			require.ErrorContains(t, err, "uniswapv3: callQuoteExactInput: ")
		})
	}
}

func TestQuoterMethodsRejectInvalidRequestWithoutRPC(t *testing.T) {
	quoter := types.MustNewAddressFromHex("0x61fFE014bA17989E743c5F6cB21bF9697530B21e")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	q := NewQuoter(mocks.NewClient(t))
	ctx := t.Context()

	_, err := q.Factory(ctx, nil, &FactoryRequest{})
	require.EqualError(t, err, "uniswapv3: factory: "+(&FactoryRequest{}).Validate().Error())
	_, err = q.WETH9(ctx, nil, &WETH9Request{})
	require.EqualError(t, err, "uniswapv3: weth9: "+(&WETH9Request{}).Validate().Error())
	_, err = q.CallQuoteExactInputSingle(ctx, nil, NewQuoteExactInputSingleRequest(quoter, nil))
	require.EqualError(t, err, "uniswapv3: callQuoteExactInputSingle: invalid request: params is nil")
	_, err = q.CallQuoteExactInput(ctx, nil, NewQuoteExactInputRequest(quoter, nil, big.NewInt(1)))
	require.EqualError(t, err, "uniswapv3: callQuoteExactInput: invalid request: path is empty")
	_, err = q.CallQuoteExactOutputSingle(ctx, nil, NewQuoteExactOutputSingleRequest(quoter, &QuoteExactOutputSingleParams{TokenIn: a}))
	require.EqualError(t, err, "uniswapv3: callQuoteExactOutputSingle: invalid request: params.tokenOut is nil")
	_, err = q.CallQuoteExactOutput(ctx, nil, &QuoteExactOutputRequest{})
	require.EqualError(t, err, "uniswapv3: callQuoteExactOutput: invalid request: contract is nil")
	require.ErrorIs(t, err, bind.ErrInvalidRequest)

	_, err = q.CallQuoteExactInputSingle(ctx, nil, NewQuoteExactInputSingleRequest(quoter, NewQuoteExactInputSingleParams(a, a, big.NewInt(1), big.NewInt(1<<24), big.NewInt(0))))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)
}

func TestQuoterCallsPropagateContext(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(t.Context(), marker{}, "quoter")
	quoter := types.MustNewAddressFromHex("0x61fFE014bA17989E743c5F6cB21bF9697530B21e")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	word, err := quoterFactoryFn.EncodeReturn(a)
	require.NoError(t, err)

	cli := mocks.NewClient(t)
	cli.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(got context.Context, e rpc.Element) error {
		require.Equal(t, "quoter", got.Value(marker{}))
		*e.Result.(*string) = encoding.Hex.EncodePrefixed(word)
		return nil
	}).Once()

	_, err = NewQuoter(cli).Factory(ctx, nil, NewFactoryRequest(quoter))
	require.NoError(t, err)
}
