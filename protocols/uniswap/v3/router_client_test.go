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
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

func TestRouterReadsSendCallAndDecode(t *testing.T) {
	router := types.MustNewAddressFromHex("0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	c := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	path, err := EncodePath([]*types.Address{a, b}, []*big.Int{big.NewInt(3000)})
	require.NoError(t, err)
	opts := bind.NewCallOptions().WithFrom(a).WithBlock("0x1b4")

	result0, err := routerFactoryFn.EncodeReturn(b)
	require.NoError(t, err)
	result1, err := routerWETH9Fn.EncodeReturn(c)
	require.NoError(t, err)
	result2, err := exactInputSingleFn.EncodeReturn(big.NewInt(995))
	require.NoError(t, err)
	inSingleData, err := EncodeExactInputSingle(a, b, big.NewInt(3000), c, big.NewInt(1000), big.NewInt(990), big.NewInt(0))
	require.NoError(t, err)
	result3, err := exactInputFn.EncodeReturn(big.NewInt(985))
	require.NoError(t, err)
	inData, err := EncodeExactInput(path, c, big.NewInt(1000), big.NewInt(980))
	require.NoError(t, err)
	result4, err := exactOutputSingleFn.EncodeReturn(big.NewInt(510))
	require.NoError(t, err)
	outSingleData, err := EncodeExactOutputSingle(a, b, big.NewInt(3000), c, big.NewInt(500), big.NewInt(600), big.NewInt(0))
	require.NoError(t, err)
	result5, err := exactOutputFn.EncodeReturn(big.NewInt(520))
	require.NoError(t, err)
	outData, err := EncodeExactOutput(path, c, big.NewInt(500), big.NewInt(600))
	require.NoError(t, err)
	result6, err := routerMulticallFn.EncodeReturn([][]byte{{0x01}, {0x02}})
	require.NoError(t, err)
	multicallData, err := EncodeRouterMulticall([][]byte{{0xaa}, {0xbb, 0xcc}})
	require.NoError(t, err)
	result7, err := routerMulticallDeadlineFn.EncodeReturn([][]byte{{0x03}})
	require.NoError(t, err)
	multicallDeadlineData, err := EncodeRouterMulticallWithDeadline(big.NewInt(1_700_000_000), [][]byte{{0xaa}})
	require.NoError(t, err)

	for _, tc := range []struct {
		name      string
		calldata  []byte
		result    []byte
		wantValue string
		call      func(r *Router) (any, error)
		want      any
	}{
		{
			"factory",
			EncodeRouterFactory(),
			result0,
			"",
			func(r *Router) (any, error) {
				return r.Factory(t.Context(), opts, NewFactoryRequest(router))
			},
			b,
		},
		{
			"weth9",
			EncodeRouterWETH9(),
			result1,
			"",
			func(r *Router) (any, error) {
				return r.WETH9(t.Context(), opts, NewWETH9Request(router))
			},
			c,
		},
		{
			"callExactInputSingle",
			inSingleData,
			result2,
			"0x3e8",
			func(r *Router) (any, error) {
				return r.CallExactInputSingle(t.Context(), opts, NewExactInputSingleRequest(router, NewExactInputSingleParams(a, b, big.NewInt(3000), c, big.NewInt(1000), big.NewInt(990), big.NewInt(0)), big.NewInt(1000)))
			},
			big.NewInt(995),
		},
		{
			"callExactInput",
			inData,
			result3,
			"",
			func(r *Router) (any, error) {
				return r.CallExactInput(t.Context(), opts, NewExactInputRequest(router, NewExactInputParams(path, c, big.NewInt(1000), big.NewInt(980)), nil))
			},
			big.NewInt(985),
		},
		{
			"callExactOutputSingle",
			outSingleData,
			result4,
			"0x258",
			func(r *Router) (any, error) {
				return r.CallExactOutputSingle(t.Context(), opts, NewExactOutputSingleRequest(router, NewExactOutputSingleParams(a, b, big.NewInt(3000), c, big.NewInt(500), big.NewInt(600), big.NewInt(0)), big.NewInt(600)))
			},
			big.NewInt(510),
		},
		{
			"callExactOutput",
			outData,
			result5,
			"",
			func(r *Router) (any, error) {
				return r.CallExactOutput(t.Context(), opts, NewExactOutputRequest(router, NewExactOutputParams(path, c, big.NewInt(500), big.NewInt(600)), nil))
			},
			big.NewInt(520),
		},
		{
			"callMulticall",
			multicallData,
			result6,
			"0x2a",
			func(r *Router) (any, error) {
				return r.CallMulticall(t.Context(), opts, NewRouterMulticallRequest(router, [][]byte{{0xaa}, {0xbb, 0xcc}}, nil, big.NewInt(42)))
			},
			[][]byte{{0x01}, {0x02}},
		},
		{
			"callMulticall with deadline",
			multicallDeadlineData,
			result7,
			"",
			func(r *Router) (any, error) {
				return r.CallMulticall(t.Context(), opts, NewRouterMulticallRequest(router, [][]byte{{0xaa}}, big.NewInt(1_700_000_000), nil))
			},
			[][]byte{{0x03}},
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
				require.Equal(t, router.String(), call["to"])
				require.Equal(t, a.String(), call["from"])
				require.Equal(t, encoding.Hex.EncodePrefixed(tc.calldata), call["data"])
				if tc.wantValue == "" {
					require.NotContains(t, call, "value")
				} else {
					require.Equal(t, tc.wantValue, call["value"])
				}
				require.Equal(t, "0x1b4", params[1])
				*e.Result.(*string) = encoding.Hex.EncodePrefixed(tc.result)
				return nil
			}).Once()

			got, err := tc.call(NewRouter(cli))
			require.NoError(t, err)
			require.Equal(t, fmt.Sprint(tc.want), fmt.Sprint(got))
		})
	}
}

func TestRouterReadsMapRevertToSentinel(t *testing.T) {
	router := types.MustNewAddressFromHex("0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	c := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	path, err := EncodePath([]*types.Address{a, b}, []*big.Int{big.NewInt(3000)})
	require.NoError(t, err)
	errorSel := []byte{0x08, 0xc3, 0x79, 0xa0}
	all := []error{ErrTransactionTooOld, ErrTooLittleReceived, ErrTooMuchRequested, ErrInsufficientWETH9, ErrInsufficientToken, ErrNotWETH9, ErrTransferFromFailed, ErrTokenTransferFailed, ErrApproveFailed, ErrTransferETHFailed, ErrInvalidSqrtPriceLimit, ErrLocked, ErrReverted}

	for _, tc := range []struct {
		name   string
		reason string
		want   error
	}{
		{"transaction too old", reasonTransactionTooOld, ErrTransactionTooOld},
		{"too little received", reasonTooLittleReceived, ErrTooLittleReceived},
		{"too much requested", reasonTooMuchRequested, ErrTooMuchRequested},
		{"insufficient weth9", reasonInsufficientWETH9, ErrInsufficientWETH9},
		{"insufficient token", reasonInsufficientToken, ErrInsufficientToken},
		{"not weth9", reasonNotWETH9, ErrNotWETH9},
		{"transfer from failed", reasonTransferFromFailed, ErrTransferFromFailed},
		{"token transfer failed", reasonTokenTransferFailed, ErrTokenTransferFailed},
		{"approve failed", reasonApproveFailed, ErrApproveFailed},
		{"transfer eth failed", reasonTransferETHFailed, ErrTransferETHFailed},
		{"pool sqrt price limit", reasonInvalidSqrtPriceLimit, ErrInvalidSqrtPriceLimit},
		{"pool locked", reasonLocked, ErrLocked},
		{"other reason", "XYZ", ErrReverted},
	} {
		t.Run(tc.name, func(t *testing.T) {
			packed, err := abi.Pack(abi.NewTypes(abi.String), tc.reason)
			require.NoError(t, err)
			revert := append(append([]byte{}, errorSel...), packed...)

			cli := mocks.NewClient(t)
			cli.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(revert) + `"`)}).Once()

			amountOut, err := NewRouter(cli).CallExactInput(t.Context(), bind.NewCallOptions().WithFrom(a), NewExactInputRequest(router, NewExactInputParams(path, c, big.NewInt(1), big.NewInt(0)), nil))
			require.Nil(t, amountOut)
			require.ErrorIs(t, err, tc.want)
			for _, other := range all {
				if other != tc.want {
					require.NotErrorIs(t, err, other)
				}
			}
			require.ErrorContains(t, err, "uniswapv3: callExactInput: ")
			var rpcErr *rpc.ResponseError
			require.ErrorAs(t, err, &rpcErr)
		})
	}

	for _, data := range []jsontext.Value{nil, jsontext.Value(`"0x"`), jsontext.Value(`{"x":1}`)} {
		cli := mocks.NewClient(t)
		cli.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: data}).Once()

		results, err := NewRouter(cli).CallMulticall(t.Context(), nil, NewRouterMulticallRequest(router, [][]byte{{0x01}}, nil, nil))
		require.Nil(t, results)
		require.ErrorIs(t, err, ErrReverted)
	}

	cli := mocks.NewClient(t)
	cli.EXPECT().Call(mock.Anything, mock.Anything).Return(errors.New("dial tcp: connection refused")).Once()

	weth9, err := NewRouter(cli).WETH9(t.Context(), nil, NewWETH9Request(router))
	require.Nil(t, weth9)
	require.EqualError(t, err, "uniswapv3: weth9: dial tcp: connection refused")
	require.NotErrorIs(t, err, ErrReverted)
}

func TestRouterReadsRejectMalformedResult(t *testing.T) {
	router := types.MustNewAddressFromHex("0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	c := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")

	for _, result := range []string{"0x", "0xzz", "0x01", "0x00", "not hex"} {
		t.Run(result, func(t *testing.T) {
			cli := mocks.NewClient(t)
			cli.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				*e.Result.(*string) = result
				return nil
			}).Times(3)
			r := NewRouter(cli)

			factory, err := r.Factory(t.Context(), nil, NewFactoryRequest(router))
			require.Nil(t, factory)
			require.ErrorContains(t, err, "uniswapv3: factory: ")

			amountIn, err := r.CallExactOutputSingle(t.Context(), nil, NewExactOutputSingleRequest(router, NewExactOutputSingleParams(a, b, big.NewInt(3000), c, big.NewInt(1), big.NewInt(1), big.NewInt(0)), nil))
			require.Nil(t, amountIn)
			require.ErrorContains(t, err, "uniswapv3: callExactOutputSingle: ")

			results, err := r.CallMulticall(t.Context(), nil, NewRouterMulticallRequest(router, [][]byte{{0x01}}, nil, nil))
			require.Nil(t, results)
			require.ErrorContains(t, err, "uniswapv3: callMulticall: ")
		})
	}
}

func TestRouterMethodsRejectInvalidRequestWithoutRPC(t *testing.T) {
	router := types.MustNewAddressFromHex("0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	c := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	path, err := EncodePath([]*types.Address{a, b}, []*big.Int{big.NewInt(3000)})
	require.NoError(t, err)
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sig, err := key.Sign(core.Keccak256([]byte("permit")))
	require.NoError(t, err)
	opts := bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder())
	r := NewRouter(mocks.NewClient(t))
	ctx := t.Context()

	_, err = r.Factory(ctx, nil, &FactoryRequest{})
	require.EqualError(t, err, "uniswapv3: factory: "+(&FactoryRequest{}).Validate().Error())
	_, err = r.WETH9(ctx, nil, &WETH9Request{})
	require.EqualError(t, err, "uniswapv3: weth9: "+(&WETH9Request{}).Validate().Error())
	_, err = r.CallExactInputSingle(ctx, nil, &ExactInputSingleRequest{})
	require.EqualError(t, err, "uniswapv3: callExactInputSingle: "+(&ExactInputSingleRequest{}).Validate().Error())
	_, err = r.CallExactInput(ctx, nil, &ExactInputRequest{})
	require.EqualError(t, err, "uniswapv3: callExactInput: "+(&ExactInputRequest{}).Validate().Error())
	_, err = r.CallExactOutputSingle(ctx, nil, &ExactOutputSingleRequest{})
	require.EqualError(t, err, "uniswapv3: callExactOutputSingle: "+(&ExactOutputSingleRequest{}).Validate().Error())
	_, err = r.CallExactOutput(ctx, nil, &ExactOutputRequest{})
	require.EqualError(t, err, "uniswapv3: callExactOutput: "+(&ExactOutputRequest{}).Validate().Error())
	_, err = r.CallMulticall(ctx, nil, &RouterMulticallRequest{})
	require.EqualError(t, err, "uniswapv3: callMulticall: "+(&RouterMulticallRequest{}).Validate().Error())
	_, _, err = r.ExactInputSingle(ctx, opts, &ExactInputSingleRequest{})
	require.EqualError(t, err, "uniswapv3: exactInputSingle: "+(&ExactInputSingleRequest{}).Validate().Error())
	_, _, err = r.ExactInput(ctx, opts, &ExactInputRequest{})
	require.EqualError(t, err, "uniswapv3: exactInput: "+(&ExactInputRequest{}).Validate().Error())
	_, _, err = r.ExactOutputSingle(ctx, opts, &ExactOutputSingleRequest{})
	require.EqualError(t, err, "uniswapv3: exactOutputSingle: "+(&ExactOutputSingleRequest{}).Validate().Error())
	_, _, err = r.ExactOutput(ctx, opts, &ExactOutputRequest{})
	require.EqualError(t, err, "uniswapv3: exactOutput: "+(&ExactOutputRequest{}).Validate().Error())
	_, _, err = r.UnwrapWETH9(ctx, opts, &RouterUnwrapWETH9Request{})
	require.EqualError(t, err, "uniswapv3: unwrapWETH9: "+(&RouterUnwrapWETH9Request{}).Validate().Error())
	_, _, err = r.UnwrapWETH9WithFee(ctx, opts, &RouterUnwrapWETH9WithFeeRequest{})
	require.EqualError(t, err, "uniswapv3: unwrapWETH9WithFee: "+(&RouterUnwrapWETH9WithFeeRequest{}).Validate().Error())
	_, _, err = r.SweepToken(ctx, opts, &RouterSweepTokenRequest{})
	require.EqualError(t, err, "uniswapv3: sweepToken: "+(&RouterSweepTokenRequest{}).Validate().Error())
	_, _, err = r.SweepTokenWithFee(ctx, opts, &RouterSweepTokenWithFeeRequest{})
	require.EqualError(t, err, "uniswapv3: sweepTokenWithFee: "+(&RouterSweepTokenWithFeeRequest{}).Validate().Error())
	_, _, err = r.RefundETH(ctx, opts, &RouterRefundETHRequest{})
	require.EqualError(t, err, "uniswapv3: refundETH: "+(&RouterRefundETHRequest{}).Validate().Error())
	_, _, err = r.Multicall(ctx, opts, &RouterMulticallRequest{})
	require.EqualError(t, err, "uniswapv3: multicall: "+(&RouterMulticallRequest{}).Validate().Error())
	_, _, err = r.SelfPermit(ctx, opts, &SelfPermitRequest{})
	require.EqualError(t, err, "uniswapv3: selfPermit: "+(&SelfPermitRequest{}).Validate().Error())
	_, _, err = r.SelfPermitIfNecessary(ctx, opts, &SelfPermitIfNecessaryRequest{})
	require.EqualError(t, err, "uniswapv3: selfPermitIfNecessary: "+(&SelfPermitIfNecessaryRequest{}).Validate().Error())
	_, _, err = r.SelfPermitAllowed(ctx, opts, &SelfPermitAllowedRequest{})
	require.EqualError(t, err, "uniswapv3: selfPermitAllowed: "+(&SelfPermitAllowedRequest{}).Validate().Error())
	_, _, err = r.SelfPermitAllowedIfNecessary(ctx, opts, &SelfPermitAllowedIfNecessaryRequest{})
	require.EqualError(t, err, "uniswapv3: selfPermitAllowedIfNecessary: "+(&SelfPermitAllowedIfNecessaryRequest{}).Validate().Error())
	require.ErrorIs(t, err, bind.ErrInvalidRequest)

	_, _, err = r.ExactInput(ctx, opts, NewExactInputRequest(router, NewExactInputParams(nil, c, big.NewInt(1), big.NewInt(0)), nil))
	require.EqualError(t, err, "uniswapv3: exactInput: invalid request: params.path is empty")

	_, _, err = r.ExactInputSingle(ctx, opts, NewExactInputSingleRequest(router, NewExactInputSingleParams(a, b, big.NewInt(1<<24), c, big.NewInt(1), big.NewInt(0), big.NewInt(0)), nil))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)

	_, _, err = r.ExactInputSingle(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewExactInputSingleRequest(router, NewExactInputSingleParams(a, b, big.NewInt(3000), c, big.NewInt(1), big.NewInt(0), big.NewInt(0)), nil))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: exactInputSingle: invalid options: value must be nil")
	_, _, err = r.ExactInput(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewExactInputRequest(router, NewExactInputParams(path, c, big.NewInt(1), big.NewInt(0)), nil))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: exactInput: invalid options: value must be nil")
	_, _, err = r.ExactOutputSingle(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewExactOutputSingleRequest(router, NewExactOutputSingleParams(a, b, big.NewInt(3000), c, big.NewInt(1), big.NewInt(1), big.NewInt(0)), nil))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: exactOutputSingle: invalid options: value must be nil")
	_, _, err = r.ExactOutput(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewExactOutputRequest(router, NewExactOutputParams(path, c, big.NewInt(1), big.NewInt(1)), nil))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: exactOutput: invalid options: value must be nil")
	_, _, err = r.UnwrapWETH9(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewRouterUnwrapWETH9Request(router, big.NewInt(0), c))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: unwrapWETH9: invalid options: value must be nil")
	_, _, err = r.UnwrapWETH9WithFee(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewRouterUnwrapWETH9WithFeeRequest(router, big.NewInt(0), c, big.NewInt(1), b))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: unwrapWETH9WithFee: invalid options: value must be nil")
	_, _, err = r.SweepToken(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewRouterSweepTokenRequest(router, a, big.NewInt(0), c))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: sweepToken: invalid options: value must be nil")
	_, _, err = r.SweepTokenWithFee(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewRouterSweepTokenWithFeeRequest(router, a, big.NewInt(0), c, big.NewInt(1), b))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: sweepTokenWithFee: invalid options: value must be nil")
	_, _, err = r.RefundETH(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewRouterRefundETHRequest(router))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: refundETH: invalid options: value must be nil")
	_, _, err = r.Multicall(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewRouterMulticallRequest(router, [][]byte{{0x01}}, nil, nil))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: multicall: invalid options: value must be nil")
	_, _, err = r.SelfPermit(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewSelfPermitRequest(router, a, big.NewInt(1), big.NewInt(1), sig))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: selfPermit: invalid options: value must be nil")
	_, _, err = r.SelfPermitIfNecessary(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewSelfPermitIfNecessaryRequest(router, a, big.NewInt(1), big.NewInt(1), sig))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: selfPermitIfNecessary: invalid options: value must be nil")
	_, _, err = r.SelfPermitAllowed(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewSelfPermitAllowedRequest(router, a, big.NewInt(1), big.NewInt(1), sig))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: selfPermitAllowed: invalid options: value must be nil")
	_, _, err = r.SelfPermitAllowedIfNecessary(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewSelfPermitAllowedIfNecessaryRequest(router, a, big.NewInt(1), big.NewInt(1), sig))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: selfPermitAllowedIfNecessary: invalid options: value must be nil")
	_, err = r.CallExactInputSingle(ctx, bind.NewCallOptions().WithValue(big.NewInt(1)), NewExactInputSingleRequest(router, NewExactInputSingleParams(a, b, big.NewInt(3000), c, big.NewInt(1), big.NewInt(0), big.NewInt(0)), nil))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: callExactInputSingle: invalid options: value must be nil")
	_, err = r.CallExactInput(ctx, bind.NewCallOptions().WithValue(big.NewInt(1)), NewExactInputRequest(router, NewExactInputParams(path, c, big.NewInt(1), big.NewInt(0)), nil))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: callExactInput: invalid options: value must be nil")
	_, err = r.CallExactOutputSingle(ctx, bind.NewCallOptions().WithValue(big.NewInt(1)), NewExactOutputSingleRequest(router, NewExactOutputSingleParams(a, b, big.NewInt(3000), c, big.NewInt(1), big.NewInt(1), big.NewInt(0)), nil))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: callExactOutputSingle: invalid options: value must be nil")
	_, err = r.CallExactOutput(ctx, bind.NewCallOptions().WithValue(big.NewInt(1)), NewExactOutputRequest(router, NewExactOutputParams(path, c, big.NewInt(1), big.NewInt(1)), nil))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: callExactOutput: invalid options: value must be nil")
	_, err = r.CallMulticall(ctx, bind.NewCallOptions().WithValue(big.NewInt(1)), NewRouterMulticallRequest(router, [][]byte{{0x01}}, nil, nil))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: callMulticall: invalid options: value must be nil")

	_, _, err = r.RefundETH(ctx, nil, NewRouterRefundETHRequest(router))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: refundETH: bind: transact: invalid options: options is nil")
}

func TestRouterWritesSignAndSend(t *testing.T) {
	router := types.MustNewAddressFromHex("0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	c := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	path, err := EncodePath([]*types.Address{a, b}, []*big.Int{big.NewInt(3000)})
	require.NoError(t, err)
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sig, err := key.Sign(core.Keccak256([]byte("permit")))
	require.NoError(t, err)
	sender := core.PubkeyToAddress(key.PublicKey())

	w0, err := EncodeExactInputSingle(a, b, big.NewInt(3000), c, big.NewInt(1000), big.NewInt(990), big.NewInt(0))
	require.NoError(t, err)
	w1, err := EncodeExactInput(path, c, big.NewInt(1000), big.NewInt(980))
	require.NoError(t, err)
	w2, err := EncodeExactOutputSingle(a, b, big.NewInt(3000), c, big.NewInt(500), big.NewInt(600), big.NewInt(0))
	require.NoError(t, err)
	w3, err := EncodeExactOutput(path, c, big.NewInt(500), big.NewInt(600))
	require.NoError(t, err)
	w4, err := EncodeRouterUnwrapWETH9(big.NewInt(5), c)
	require.NoError(t, err)
	w5, err := EncodeUnwrapWETH9WithFee(big.NewInt(5), c, big.NewInt(25), b)
	require.NoError(t, err)
	w6, err := EncodeRouterSweepToken(a, big.NewInt(5), c)
	require.NoError(t, err)
	w7, err := EncodeSweepTokenWithFee(a, big.NewInt(5), c, big.NewInt(25), b)
	require.NoError(t, err)
	w8 := EncodeRouterRefundETH()
	w9, err := EncodeRouterMulticall([][]byte{{0xaa}})
	require.NoError(t, err)
	w10, err := EncodeRouterMulticallWithDeadline(big.NewInt(1_700_000_000), [][]byte{{0xaa}})
	require.NoError(t, err)
	w11, err := EncodeSelfPermit(a, big.NewInt(7), big.NewInt(1_700_000_000), sig)
	require.NoError(t, err)
	w12, err := EncodeSelfPermitIfNecessary(a, big.NewInt(7), big.NewInt(1_700_000_000), sig)
	require.NoError(t, err)
	w13, err := EncodeSelfPermitAllowed(a, big.NewInt(3), big.NewInt(1_700_000_000), sig)
	require.NoError(t, err)
	w14, err := EncodeSelfPermitAllowedIfNecessary(a, big.NewInt(3), big.NewInt(1_700_000_000), sig)
	require.NoError(t, err)

	for _, tc := range []struct {
		name      string
		calldata  []byte
		wantValue *big.Int
		send      func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error)
	}{
		{
			"exactInputSingle",
			w0,
			big.NewInt(1000),
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.ExactInputSingle(t.Context(), opts, NewExactInputSingleRequest(router, NewExactInputSingleParams(a, b, big.NewInt(3000), c, big.NewInt(1000), big.NewInt(990), big.NewInt(0)), big.NewInt(1000)))
			},
		},
		{
			"exactInput",
			w1,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.ExactInput(t.Context(), opts, NewExactInputRequest(router, NewExactInputParams(path, c, big.NewInt(1000), big.NewInt(980)), nil))
			},
		},
		{
			"exactOutputSingle",
			w2,
			big.NewInt(600),
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.ExactOutputSingle(t.Context(), opts, NewExactOutputSingleRequest(router, NewExactOutputSingleParams(a, b, big.NewInt(3000), c, big.NewInt(500), big.NewInt(600), big.NewInt(0)), big.NewInt(600)))
			},
		},
		{
			"exactOutput",
			w3,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.ExactOutput(t.Context(), opts, NewExactOutputRequest(router, NewExactOutputParams(path, c, big.NewInt(500), big.NewInt(600)), nil))
			},
		},
		{
			"unwrapWETH9",
			w4,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.UnwrapWETH9(t.Context(), opts, NewRouterUnwrapWETH9Request(router, big.NewInt(5), c))
			},
		},
		{
			"unwrapWETH9WithFee",
			w5,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.UnwrapWETH9WithFee(t.Context(), opts, NewRouterUnwrapWETH9WithFeeRequest(router, big.NewInt(5), c, big.NewInt(25), b))
			},
		},
		{
			"sweepToken",
			w6,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.SweepToken(t.Context(), opts, NewRouterSweepTokenRequest(router, a, big.NewInt(5), c))
			},
		},
		{
			"sweepTokenWithFee",
			w7,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.SweepTokenWithFee(t.Context(), opts, NewRouterSweepTokenWithFeeRequest(router, a, big.NewInt(5), c, big.NewInt(25), b))
			},
		},
		{
			"refundETH",
			w8,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.RefundETH(t.Context(), opts, NewRouterRefundETHRequest(router))
			},
		},
		{
			"multicall",
			w9,
			big.NewInt(42),
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.Multicall(t.Context(), opts, NewRouterMulticallRequest(router, [][]byte{{0xaa}}, nil, big.NewInt(42)))
			},
		},
		{
			"multicall with deadline",
			w10,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.Multicall(t.Context(), opts, NewRouterMulticallRequest(router, [][]byte{{0xaa}}, big.NewInt(1_700_000_000), nil))
			},
		},
		{
			"selfPermit",
			w11,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.SelfPermit(t.Context(), opts, NewSelfPermitRequest(router, a, big.NewInt(7), big.NewInt(1_700_000_000), sig))
			},
		},
		{
			"selfPermitIfNecessary",
			w12,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.SelfPermitIfNecessary(t.Context(), opts, NewSelfPermitIfNecessaryRequest(router, a, big.NewInt(7), big.NewInt(1_700_000_000), sig))
			},
		},
		{
			"selfPermitAllowed",
			w13,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.SelfPermitAllowed(t.Context(), opts, NewSelfPermitAllowedRequest(router, a, big.NewInt(3), big.NewInt(1_700_000_000), sig))
			},
		},
		{
			"selfPermitAllowedIfNecessary",
			w14,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.SelfPermitAllowedIfNecessary(t.Context(), opts, NewSelfPermitAllowedIfNecessaryRequest(router, a, big.NewInt(3), big.NewInt(1_700_000_000), sig))
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			builder := tx.NewDynamicFeeBuilder()
			builder.ChainID(big.NewInt(1))
			builder.Nonce(7)
			builder.Limit(3_000_000)
			builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

			var raw []byte
			cli := mocks.NewClient(t)
			cli.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				require.Equal(t, rpc.MethodETHSendRawTransaction, e.Method)
				raw, err = encoding.Hex.Decode(e.Params.([]any)[0].(string))
				require.NoError(t, err)
				*e.Result.(*string) = core.Keccak256(raw).String()
				return nil
			}).Once()

			opts := bind.NewTransactOptions(key, builder)
			signed, hash, err := tc.send(NewRouter(cli), opts)
			require.NoError(t, err)
			require.True(t, hash.Equal(core.Keccak256(raw)))
			require.Equal(t, uint64(7), signed.AddressNonce())
			require.Nil(t, opts.Value)

			decoded, err := core.DecodeTransaction(raw)
			require.NoError(t, err)
			dyn, ok := decoded.(*core.DynamicFeeTx)
			require.True(t, ok)
			require.True(t, dyn.To.Equal(router))
			if tc.wantValue == nil {
				require.Zero(t, dyn.Value.Sign())
			} else {
				require.Zero(t, tc.wantValue.Cmp(dyn.Value))
			}
			require.Equal(t, tc.calldata, dyn.Data)
			got, err := decoded.Sender()
			require.NoError(t, err)
			require.True(t, got.Equal(sender))
		})
	}
}

func TestRouterWritesMapEstimateRevertToSentinel(t *testing.T) {
	router := types.MustNewAddressFromHex("0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	c := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	path, err := EncodePath([]*types.Address{a, b}, []*big.Int{big.NewInt(3000)})
	require.NoError(t, err)
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)

	packed, err := abi.Pack(abi.NewTypes(abi.String), reasonTooLittleReceived)
	require.NoError(t, err)
	revert := append([]byte{0x08, 0xc3, 0x79, 0xa0}, packed...)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(7)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	cli := mocks.NewClient(t)
	cli.EXPECT().Batch(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, elems rpc.Elements) error {
		require.Equal(t, 1, elems.Len())
		require.Equal(t, rpc.MethodETHEstimateGas, elems.GetMethod(0))
		return fmt.Errorf("rpc: %s: %w", rpc.MethodETHEstimateGas, &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(revert) + `"`)})
	}).Once()

	signed, hash, err := NewRouter(cli).ExactInput(t.Context(), bind.NewTransactOptions(key, builder), NewExactInputRequest(router, NewExactInputParams(path, c, big.NewInt(1), big.NewInt(1_000_000)), nil))
	require.ErrorIs(t, err, ErrTooLittleReceived)
	require.NotErrorIs(t, err, ErrReverted)
	require.ErrorContains(t, err, "uniswapv3: exactInput: ")
	require.Nil(t, signed)
	require.Nil(t, hash)
}

func TestRouterWritesReturnSignedTxOnBroadcastFailure(t *testing.T) {
	router := types.MustNewAddressFromHex("0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	c := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(9)
	builder.Limit(3_000_000)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	cli := mocks.NewClient(t)
	cli.EXPECT().Call(mock.Anything, mock.Anything).Return(errors.New("dial tcp: connection refused")).Twice()

	signed, hash, err := NewRouter(cli).ExactInputSingle(t.Context(), bind.NewTransactOptions(key, builder), NewExactInputSingleRequest(router, NewExactInputSingleParams(a, b, big.NewInt(3000), c, big.NewInt(1), big.NewInt(0), big.NewInt(0)), big.NewInt(777)))
	require.ErrorContains(t, err, "uniswapv3: exactInputSingle: ")
	require.ErrorContains(t, err, "connection refused")
	require.NotErrorIs(t, err, ErrReverted)
	require.Nil(t, hash)
	require.NotNil(t, signed)
	require.Equal(t, uint64(9), signed.AddressNonce())
	dyn, ok := signed.(*core.DynamicFeeTx)
	require.True(t, ok)
	require.Zero(t, big.NewInt(777).Cmp(dyn.Value))
}

func TestRouterCallsPropagateContext(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(t.Context(), marker{}, "router")
	router := types.MustNewAddressFromHex("0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	word, err := routerWETH9Fn.EncodeReturn(router)
	require.NoError(t, err)

	cli := mocks.NewClient(t)
	cli.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(got context.Context, e rpc.Element) error {
		require.Equal(t, "router", got.Value(marker{}))
		if e.Method == rpc.MethodETHCall {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(word)
		}
		return nil
	}).Twice()

	_, err = NewRouter(cli).WETH9(ctx, nil, NewWETH9Request(router))
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(1)
	builder.Limit(3_000_000)
	builder.Price(big.NewInt(2), big.NewInt(1))
	_, _, err = NewRouter(cli).RefundETH(ctx, bind.NewTransactOptions(key, builder), NewRouterRefundETHRequest(router))
	require.NoError(t, err)
}

func TestMapRevertErrorCoversRouterPoolAndNFPM(t *testing.T) {
	for _, tc := range []struct {
		reason string
		want   error
	}{
		{"Too little received", ErrTooLittleReceived},
		{"Transaction too old", ErrTransactionTooOld},
		{"LOK", ErrLocked},
		{"SPL", ErrInvalidSqrtPriceLimit},
		{"Invalid token ID", ErrInvalidTokenID},
		{"Not approved", ErrNotApproved},
	} {
		t.Run(tc.reason, func(t *testing.T) {
			packed, err := abi.Pack(abi.NewTypes(abi.String), tc.reason)
			require.NoError(t, err)
			rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"0x08c379a0` + encoding.Hex.Encode(packed) + `"`)}

			mapped := MapRevertError(rpcErr)
			require.ErrorIs(t, mapped, tc.want)
			require.NotErrorIs(t, mapped, ErrReverted)
			got, ok := errors.AsType[*rpc.ResponseError](mapped)
			require.True(t, ok)
			require.Same(t, rpcErr, got)
		})
	}

	packed, err := abi.Pack(abi.NewTypes(abi.String), "UniswapV2: K")
	require.NoError(t, err)
	require.ErrorIs(t, MapRevertError(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"0x08c379a0` + encoding.Hex.Encode(packed) + `"`)}), ErrReverted)
	require.ErrorIs(t, MapRevertError(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"0xdeadbeef"`)}), ErrReverted)

	plain := errors.New("dial tcp: connection refused")
	require.Same(t, plain, MapRevertError(plain))
	require.NoError(t, MapRevertError(nil))
	var success *rpc.ResponseError
	require.NoError(t, MapRevertError(success))
}
