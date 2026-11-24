package v2

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
	router := types.MustNewAddressFromHex("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	to := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	path := []*types.Address{a, b}
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sig, err := key.Sign(core.Keccak256([]byte("permit")))
	require.NoError(t, err)
	opts := bind.NewCallOptions().WithFrom(a).WithBlock("0x1b4")

	factoryWord, err := abi.Pack(abi.NewTypes(abi.Address), a)
	require.NoError(t, err)
	wethWord, err := abi.Pack(abi.NewTypes(abi.Address), b)
	require.NoError(t, err)
	quoteData, err := EncodeQuote(big.NewInt(1007), big.NewInt(2007), big.NewInt(3007))
	require.NoError(t, err)
	quoteResult, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(11))
	require.NoError(t, err)
	getAmountOutData, err := EncodeGetAmountOut(big.NewInt(1007), big.NewInt(2007), big.NewInt(3007))
	require.NoError(t, err)
	getAmountOutResult, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(11))
	require.NoError(t, err)
	getAmountInData, err := EncodeGetAmountIn(big.NewInt(1007), big.NewInt(2007), big.NewInt(3007))
	require.NoError(t, err)
	getAmountInResult, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(11))
	require.NoError(t, err)
	getAmountsOutData, err := EncodeGetAmountsOut(big.NewInt(1007), path)
	require.NoError(t, err)
	getAmountsOutResult, err := abi.Pack(abi.NewTypes(abi.Slice(abi.Uint256)), []*big.Int{big.NewInt(11), big.NewInt(22), big.NewInt(33)})
	require.NoError(t, err)
	getAmountsInData, err := EncodeGetAmountsIn(big.NewInt(1007), path)
	require.NoError(t, err)
	getAmountsInResult, err := abi.Pack(abi.NewTypes(abi.Slice(abi.Uint256)), []*big.Int{big.NewInt(11), big.NewInt(22), big.NewInt(33)})
	require.NoError(t, err)
	addLiquidityData, err := EncodeAddLiquidity(a, b, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), big.NewInt(4007), to, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	addLiquidityResult, err := abi.Pack(abi.NewTypes(abi.Uint256, abi.Uint256, abi.Uint256), big.NewInt(11), big.NewInt(22), big.NewInt(33))
	require.NoError(t, err)
	addLiquidityETHData, err := EncodeAddLiquidityETH(a, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	addLiquidityETHResult, err := abi.Pack(abi.NewTypes(abi.Uint256, abi.Uint256, abi.Uint256), big.NewInt(11), big.NewInt(22), big.NewInt(33))
	require.NoError(t, err)
	removeLiquidityData, err := EncodeRemoveLiquidity(a, b, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	removeLiquidityResult, err := abi.Pack(abi.NewTypes(abi.Uint256, abi.Uint256), big.NewInt(11), big.NewInt(22))
	require.NoError(t, err)
	removeLiquidityETHData, err := EncodeRemoveLiquidityETH(a, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	removeLiquidityETHResult, err := abi.Pack(abi.NewTypes(abi.Uint256, abi.Uint256), big.NewInt(11), big.NewInt(22))
	require.NoError(t, err)
	removeLiquidityWithPermitData, err := EncodeRemoveLiquidityWithPermit(a, b, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000), true, sig)
	require.NoError(t, err)
	removeLiquidityWithPermitResult, err := abi.Pack(abi.NewTypes(abi.Uint256, abi.Uint256), big.NewInt(11), big.NewInt(22))
	require.NoError(t, err)
	removeLiquidityETHWithPermitData, err := EncodeRemoveLiquidityETHWithPermit(a, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000), true, sig)
	require.NoError(t, err)
	removeLiquidityETHWithPermitResult, err := abi.Pack(abi.NewTypes(abi.Uint256, abi.Uint256), big.NewInt(11), big.NewInt(22))
	require.NoError(t, err)
	swapExactTokensForTokensData, err := EncodeSwapExactTokensForTokens(big.NewInt(1007), big.NewInt(2007), path, to, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	swapExactTokensForTokensResult, err := abi.Pack(abi.NewTypes(abi.Slice(abi.Uint256)), []*big.Int{big.NewInt(11), big.NewInt(22), big.NewInt(33)})
	require.NoError(t, err)
	swapTokensForExactTokensData, err := EncodeSwapTokensForExactTokens(big.NewInt(1007), big.NewInt(2007), path, to, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	swapTokensForExactTokensResult, err := abi.Pack(abi.NewTypes(abi.Slice(abi.Uint256)), []*big.Int{big.NewInt(11), big.NewInt(22), big.NewInt(33)})
	require.NoError(t, err)
	swapExactETHForTokensData, err := EncodeSwapExactETHForTokens(big.NewInt(1007), path, to, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	swapExactETHForTokensResult, err := abi.Pack(abi.NewTypes(abi.Slice(abi.Uint256)), []*big.Int{big.NewInt(11), big.NewInt(22), big.NewInt(33)})
	require.NoError(t, err)
	swapTokensForExactETHData, err := EncodeSwapTokensForExactETH(big.NewInt(1007), big.NewInt(2007), path, to, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	swapTokensForExactETHResult, err := abi.Pack(abi.NewTypes(abi.Slice(abi.Uint256)), []*big.Int{big.NewInt(11), big.NewInt(22), big.NewInt(33)})
	require.NoError(t, err)
	swapExactTokensForETHData, err := EncodeSwapExactTokensForETH(big.NewInt(1007), big.NewInt(2007), path, to, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	swapExactTokensForETHResult, err := abi.Pack(abi.NewTypes(abi.Slice(abi.Uint256)), []*big.Int{big.NewInt(11), big.NewInt(22), big.NewInt(33)})
	require.NoError(t, err)
	swapETHForExactTokensData, err := EncodeSwapETHForExactTokens(big.NewInt(1007), path, to, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	swapETHForExactTokensResult, err := abi.Pack(abi.NewTypes(abi.Slice(abi.Uint256)), []*big.Int{big.NewInt(11), big.NewInt(22), big.NewInt(33)})
	require.NoError(t, err)
	removeLiquidityETHSupportingFeeOnTransferTokensData, err := EncodeRemoveLiquidityETHSupportingFeeOnTransferTokens(a, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	removeLiquidityETHSupportingFeeOnTransferTokensResult, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(11))
	require.NoError(t, err)
	removeLiquidityETHWithPermitSupportingFeeOnTransferTokensData, err := EncodeRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(a, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000), true, sig)
	require.NoError(t, err)
	removeLiquidityETHWithPermitSupportingFeeOnTransferTokensResult, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(11))
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
			factoryWord,
			"",
			func(r *Router) (any, error) {
				return r.Factory(t.Context(), opts, NewFactoryRequest(router))
			},
			a,
		},
		{
			"weth",
			EncodeWETH(),
			wethWord,
			"",
			func(r *Router) (any, error) {
				return r.WETH(t.Context(), opts, NewWETHRequest(router))
			},
			b,
		},
		{
			"quote",
			quoteData,
			quoteResult,
			"",
			func(r *Router) (any, error) {
				return r.Quote(t.Context(), opts, NewQuoteRequest(router, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007)))
			},
			big.NewInt(11),
		},
		{
			"getAmountOut",
			getAmountOutData,
			getAmountOutResult,
			"",
			func(r *Router) (any, error) {
				return r.GetAmountOut(t.Context(), opts, NewGetAmountOutRequest(router, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007)))
			},
			big.NewInt(11),
		},
		{
			"getAmountIn",
			getAmountInData,
			getAmountInResult,
			"",
			func(r *Router) (any, error) {
				return r.GetAmountIn(t.Context(), opts, NewGetAmountInRequest(router, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007)))
			},
			big.NewInt(11),
		},
		{
			"getAmountsOut",
			getAmountsOutData,
			getAmountsOutResult,
			"",
			func(r *Router) (any, error) {
				return r.GetAmountsOut(t.Context(), opts, NewGetAmountsOutRequest(router, big.NewInt(1007), path))
			},
			[]*big.Int{big.NewInt(11), big.NewInt(22), big.NewInt(33)},
		},
		{
			"getAmountsIn",
			getAmountsInData,
			getAmountsInResult,
			"",
			func(r *Router) (any, error) {
				return r.GetAmountsIn(t.Context(), opts, NewGetAmountsInRequest(router, big.NewInt(1007), path))
			},
			[]*big.Int{big.NewInt(11), big.NewInt(22), big.NewInt(33)},
		},
		{
			"callAddLiquidity",
			addLiquidityData,
			addLiquidityResult,
			"",
			func(r *Router) (any, error) {
				return r.CallAddLiquidity(t.Context(), opts, NewAddLiquidityRequest(router, a, b, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), big.NewInt(4007), to, big.NewInt(1_700_000_000)))
			},
			NewAddLiquidityResult(big.NewInt(11), big.NewInt(22), big.NewInt(33)),
		},
		{
			"callAddLiquidityETH",
			addLiquidityETHData,
			addLiquidityETHResult,
			"0xfa7",
			func(r *Router) (any, error) {
				return r.CallAddLiquidityETH(t.Context(), opts, NewAddLiquidityETHRequest(router, a, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000), big.NewInt(4007)))
			},
			NewAddLiquidityETHResult(big.NewInt(11), big.NewInt(22), big.NewInt(33)),
		},
		{
			"callRemoveLiquidity",
			removeLiquidityData,
			removeLiquidityResult,
			"",
			func(r *Router) (any, error) {
				return r.CallRemoveLiquidity(t.Context(), opts, NewRemoveLiquidityRequest(router, a, b, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000)))
			},
			NewRemoveLiquidityResult(big.NewInt(11), big.NewInt(22)),
		},
		{
			"callRemoveLiquidityETH",
			removeLiquidityETHData,
			removeLiquidityETHResult,
			"",
			func(r *Router) (any, error) {
				return r.CallRemoveLiquidityETH(t.Context(), opts, NewRemoveLiquidityETHRequest(router, a, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000)))
			},
			NewRemoveLiquidityETHResult(big.NewInt(11), big.NewInt(22)),
		},
		{
			"callRemoveLiquidityWithPermit",
			removeLiquidityWithPermitData,
			removeLiquidityWithPermitResult,
			"",
			func(r *Router) (any, error) {
				return r.CallRemoveLiquidityWithPermit(t.Context(), opts, NewRemoveLiquidityWithPermitRequest(router, a, b, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000), true, sig))
			},
			NewRemoveLiquidityResult(big.NewInt(11), big.NewInt(22)),
		},
		{
			"callRemoveLiquidityETHWithPermit",
			removeLiquidityETHWithPermitData,
			removeLiquidityETHWithPermitResult,
			"",
			func(r *Router) (any, error) {
				return r.CallRemoveLiquidityETHWithPermit(t.Context(), opts, NewRemoveLiquidityETHWithPermitRequest(router, a, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000), true, sig))
			},
			NewRemoveLiquidityETHResult(big.NewInt(11), big.NewInt(22)),
		},
		{
			"callSwapExactTokensForTokens",
			swapExactTokensForTokensData,
			swapExactTokensForTokensResult,
			"",
			func(r *Router) (any, error) {
				return r.CallSwapExactTokensForTokens(t.Context(), opts, NewSwapExactTokensForTokensRequest(router, big.NewInt(1007), big.NewInt(2007), path, to, big.NewInt(1_700_000_000)))
			},
			[]*big.Int{big.NewInt(11), big.NewInt(22), big.NewInt(33)},
		},
		{
			"callSwapTokensForExactTokens",
			swapTokensForExactTokensData,
			swapTokensForExactTokensResult,
			"",
			func(r *Router) (any, error) {
				return r.CallSwapTokensForExactTokens(t.Context(), opts, NewSwapTokensForExactTokensRequest(router, big.NewInt(1007), big.NewInt(2007), path, to, big.NewInt(1_700_000_000)))
			},
			[]*big.Int{big.NewInt(11), big.NewInt(22), big.NewInt(33)},
		},
		{
			"callSwapExactETHForTokens",
			swapExactETHForTokensData,
			swapExactETHForTokensResult,
			"0x7d7",
			func(r *Router) (any, error) {
				return r.CallSwapExactETHForTokens(t.Context(), opts, NewSwapExactETHForTokensRequest(router, big.NewInt(1007), path, to, big.NewInt(1_700_000_000), big.NewInt(2007)))
			},
			[]*big.Int{big.NewInt(11), big.NewInt(22), big.NewInt(33)},
		},
		{
			"callSwapTokensForExactETH",
			swapTokensForExactETHData,
			swapTokensForExactETHResult,
			"",
			func(r *Router) (any, error) {
				return r.CallSwapTokensForExactETH(t.Context(), opts, NewSwapTokensForExactETHRequest(router, big.NewInt(1007), big.NewInt(2007), path, to, big.NewInt(1_700_000_000)))
			},
			[]*big.Int{big.NewInt(11), big.NewInt(22), big.NewInt(33)},
		},
		{
			"callSwapExactTokensForETH",
			swapExactTokensForETHData,
			swapExactTokensForETHResult,
			"",
			func(r *Router) (any, error) {
				return r.CallSwapExactTokensForETH(t.Context(), opts, NewSwapExactTokensForETHRequest(router, big.NewInt(1007), big.NewInt(2007), path, to, big.NewInt(1_700_000_000)))
			},
			[]*big.Int{big.NewInt(11), big.NewInt(22), big.NewInt(33)},
		},
		{
			"callSwapETHForExactTokens",
			swapETHForExactTokensData,
			swapETHForExactTokensResult,
			"0x7d7",
			func(r *Router) (any, error) {
				return r.CallSwapETHForExactTokens(t.Context(), opts, NewSwapETHForExactTokensRequest(router, big.NewInt(1007), path, to, big.NewInt(1_700_000_000), big.NewInt(2007)))
			},
			[]*big.Int{big.NewInt(11), big.NewInt(22), big.NewInt(33)},
		},
		{
			"callRemoveLiquidityETHSupportingFeeOnTransferTokens",
			removeLiquidityETHSupportingFeeOnTransferTokensData,
			removeLiquidityETHSupportingFeeOnTransferTokensResult,
			"",
			func(r *Router) (any, error) {
				return r.CallRemoveLiquidityETHSupportingFeeOnTransferTokens(t.Context(), opts, NewRemoveLiquidityETHSupportingFeeOnTransferTokensRequest(router, a, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000)))
			},
			big.NewInt(11),
		},
		{
			"callRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens",
			removeLiquidityETHWithPermitSupportingFeeOnTransferTokensData,
			removeLiquidityETHWithPermitSupportingFeeOnTransferTokensResult,
			"",
			func(r *Router) (any, error) {
				return r.CallRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(t.Context(), opts, NewRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest(router, a, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000), true, sig))
			},
			big.NewInt(11),
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

			got, err := tc.call(NewRouter(c))
			require.NoError(t, err)
			require.Equal(t, fmt.Sprint(tc.want), fmt.Sprint(got))
		})
	}
}

func TestRouterReadsMapRevertToSentinel(t *testing.T) {
	router := types.MustNewAddressFromHex("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	to := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	path := []*types.Address{a, b}
	errorSel := []byte{0x08, 0xc3, 0x79, 0xa0}
	all := []error{
		ErrExpired, ErrInsufficientAAmount, ErrInsufficientBAmount, ErrInsufficientOutputAmount,
		ErrExcessiveInputAmount, ErrInvalidPath, ErrIdenticalAddresses, ErrZeroAddress,
		ErrInsufficientAmount, ErrInsufficientLiquidity, ErrInsufficientInputAmount, ErrUnknown,
	}

	for _, tc := range []struct {
		name   string
		reason string
		want   error
	}{
		{"expired", reasonExpired, ErrExpired},
		{"insufficient a amount", reasonInsufficientAAmount, ErrInsufficientAAmount},
		{"insufficient b amount", reasonInsufficientBAmount, ErrInsufficientBAmount},
		{"router insufficient output", reasonRouterInsufficientOutput, ErrInsufficientOutputAmount},
		{"excessive input amount", reasonExcessiveInputAmount, ErrExcessiveInputAmount},
		{"router invalid path", reasonRouterInvalidPath, ErrInvalidPath},
		{"library identical addresses old", reasonLibraryIdenticalAddressesOld, ErrIdenticalAddresses},
		{"library zero address old", reasonLibraryZeroAddressOld, ErrZeroAddress},
		{"library insufficient amount old", reasonLibraryInsufficientAmountOld, ErrInsufficientAmount},
		{"library insufficient liquidity old", reasonLibraryInsufficientLiquidityOld, ErrInsufficientLiquidity},
		{"library insufficient input old", reasonLibraryInsufficientInputOld, ErrInsufficientInputAmount},
		{"library insufficient output old", reasonLibraryInsufficientOutputOld, ErrInsufficientOutputAmount},
		{"library invalid path old", reasonLibraryInvalidPathOld, ErrInvalidPath},
		{"library identical addresses new", reasonLibraryIdenticalAddressesNew, ErrIdenticalAddresses},
		{"library zero address new", reasonLibraryZeroAddressNew, ErrZeroAddress},
		{"library insufficient amount new", reasonLibraryInsufficientAmountNew, ErrInsufficientAmount},
		{"library insufficient liquidity new", reasonLibraryInsufficientLiquidityNew, ErrInsufficientLiquidity},
		{"library insufficient input new", reasonLibraryInsufficientInputNew, ErrInsufficientInputAmount},
		{"library insufficient output new", reasonLibraryInsufficientOutputNew, ErrInsufficientOutputAmount},
		{"library invalid path new", reasonLibraryInvalidPathNew, ErrInvalidPath},
		{"other reason", "UniswapV2Router: SOMETHING_ELSE", ErrUnknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			packed, err := abi.Pack(abi.NewTypes(abi.String), tc.reason)
			require.NoError(t, err)
			revert := append(append([]byte{}, errorSel...), packed...)

			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(revert) + `"`)}).Twice()
			r := NewRouter(c)

			amounts, err := r.CallSwapExactTokensForTokens(t.Context(), bind.NewCallOptions().WithFrom(a), NewSwapExactTokensForTokensRequest(router, big.NewInt(1), big.NewInt(0), path, to, big.NewInt(1)))
			require.Nil(t, amounts)
			require.ErrorIs(t, err, tc.want)
			for _, other := range all {
				if other != tc.want {
					require.NotErrorIs(t, err, other)
				}
			}
			var rpcErr *rpc.ResponseError
			require.ErrorAs(t, err, &rpcErr)

			quote, err := r.Quote(t.Context(), nil, NewQuoteRequest(router, big.NewInt(1), big.NewInt(1), big.NewInt(1)))
			require.Nil(t, quote)
			require.ErrorIs(t, err, tc.want)
			require.ErrorContains(t, err, "uniswapv2: quote: ")
		})
	}

	for _, data := range []jsontext.Value{nil, jsontext.Value(`"0x"`), jsontext.Value(`{"x":1}`), jsontext.Value(`"0xzz"`)} {
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: data}).Once()

		amounts, err := NewRouter(c).GetAmountsOut(t.Context(), nil, NewGetAmountsOutRequest(router, big.NewInt(1), path))
		require.Nil(t, amounts)
		require.ErrorContains(t, err, "uniswapv2: getAmountsOut: ")
		require.NotErrorIs(t, err, ErrUnknown)
	}
}

func TestRouterReadsRejectMalformedResult(t *testing.T) {
	router := types.MustNewAddressFromHex("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	to := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	path := []*types.Address{a, b}

	for _, result := range []string{"0x", "0xzz", "0x01", "0x00", "not hex"} {
		t.Run(result, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				*e.Result.(*string) = result
				return nil
			}).Times(5)
			r := NewRouter(c)

			weth, err := r.WETH(t.Context(), nil, NewWETHRequest(router))
			require.Nil(t, weth)
			require.ErrorContains(t, err, "uniswapv2: weth: ")

			amounts, err := r.GetAmountsOut(t.Context(), nil, NewGetAmountsOutRequest(router, big.NewInt(1), path))
			require.Nil(t, amounts)
			require.ErrorContains(t, err, "uniswapv2: getAmountsOut: ")

			added, err := r.CallAddLiquidity(t.Context(), nil, NewAddLiquidityRequest(router, a, b, big.NewInt(1), big.NewInt(1), big.NewInt(0), big.NewInt(0), to, big.NewInt(1)))
			require.Nil(t, added)
			require.ErrorContains(t, err, "uniswapv2: callAddLiquidity: ")

			removed, err := r.CallRemoveLiquidityETH(t.Context(), nil, NewRemoveLiquidityETHRequest(router, a, big.NewInt(1), big.NewInt(0), big.NewInt(0), to, big.NewInt(1)))
			require.Nil(t, removed)
			require.ErrorContains(t, err, "uniswapv2: callRemoveLiquidityETH: ")

			amountETH, err := r.CallRemoveLiquidityETHSupportingFeeOnTransferTokens(t.Context(), nil, NewRemoveLiquidityETHSupportingFeeOnTransferTokensRequest(router, a, big.NewInt(1), big.NewInt(0), big.NewInt(0), to, big.NewInt(1)))
			require.Nil(t, amountETH)
			require.ErrorContains(t, err, "uniswapv2: callRemoveLiquidityETHSupportingFeeOnTransferTokens: ")
		})
	}
}

func TestRouterMethodsRejectInvalidRequestWithoutRPC(t *testing.T) {
	router := types.MustNewAddressFromHex("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	to := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	path := []*types.Address{a, b}
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	opts := bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder())
	r := NewRouter(mocks.NewClient(t))
	ctx := t.Context()

	_, err = r.Factory(ctx, nil, &FactoryRequest{})
	require.EqualError(t, err, "uniswapv2: factory: "+(&FactoryRequest{}).Validate().Error())
	_, err = r.WETH(ctx, nil, &WETHRequest{})
	require.EqualError(t, err, "uniswapv2: weth: "+(&WETHRequest{}).Validate().Error())
	_, err = r.Quote(ctx, nil, &QuoteRequest{})
	require.EqualError(t, err, "uniswapv2: quote: "+(&QuoteRequest{}).Validate().Error())
	_, err = r.GetAmountOut(ctx, nil, &GetAmountOutRequest{})
	require.EqualError(t, err, "uniswapv2: getAmountOut: "+(&GetAmountOutRequest{}).Validate().Error())
	_, err = r.GetAmountIn(ctx, nil, &GetAmountInRequest{})
	require.EqualError(t, err, "uniswapv2: getAmountIn: "+(&GetAmountInRequest{}).Validate().Error())
	_, err = r.GetAmountsOut(ctx, nil, &GetAmountsOutRequest{})
	require.EqualError(t, err, "uniswapv2: getAmountsOut: "+(&GetAmountsOutRequest{}).Validate().Error())
	_, err = r.GetAmountsIn(ctx, nil, &GetAmountsInRequest{})
	require.EqualError(t, err, "uniswapv2: getAmountsIn: "+(&GetAmountsInRequest{}).Validate().Error())
	_, err = r.CallAddLiquidity(ctx, nil, &AddLiquidityRequest{})
	require.EqualError(t, err, "uniswapv2: callAddLiquidity: "+(&AddLiquidityRequest{}).Validate().Error())
	_, err = r.CallAddLiquidityETH(ctx, nil, &AddLiquidityETHRequest{})
	require.EqualError(t, err, "uniswapv2: callAddLiquidityETH: "+(&AddLiquidityETHRequest{}).Validate().Error())
	_, err = r.CallRemoveLiquidity(ctx, nil, &RemoveLiquidityRequest{})
	require.EqualError(t, err, "uniswapv2: callRemoveLiquidity: "+(&RemoveLiquidityRequest{}).Validate().Error())
	_, err = r.CallRemoveLiquidityETH(ctx, nil, &RemoveLiquidityETHRequest{})
	require.EqualError(t, err, "uniswapv2: callRemoveLiquidityETH: "+(&RemoveLiquidityETHRequest{}).Validate().Error())
	_, err = r.CallRemoveLiquidityWithPermit(ctx, nil, &RemoveLiquidityWithPermitRequest{})
	require.EqualError(t, err, "uniswapv2: callRemoveLiquidityWithPermit: "+(&RemoveLiquidityWithPermitRequest{}).Validate().Error())
	_, err = r.CallRemoveLiquidityETHWithPermit(ctx, nil, &RemoveLiquidityETHWithPermitRequest{})
	require.EqualError(t, err, "uniswapv2: callRemoveLiquidityETHWithPermit: "+(&RemoveLiquidityETHWithPermitRequest{}).Validate().Error())
	_, err = r.CallSwapExactTokensForTokens(ctx, nil, &SwapExactTokensForTokensRequest{})
	require.EqualError(t, err, "uniswapv2: callSwapExactTokensForTokens: "+(&SwapExactTokensForTokensRequest{}).Validate().Error())
	_, err = r.CallSwapTokensForExactTokens(ctx, nil, &SwapTokensForExactTokensRequest{})
	require.EqualError(t, err, "uniswapv2: callSwapTokensForExactTokens: "+(&SwapTokensForExactTokensRequest{}).Validate().Error())
	_, err = r.CallSwapExactETHForTokens(ctx, nil, &SwapExactETHForTokensRequest{})
	require.EqualError(t, err, "uniswapv2: callSwapExactETHForTokens: "+(&SwapExactETHForTokensRequest{}).Validate().Error())
	_, err = r.CallSwapTokensForExactETH(ctx, nil, &SwapTokensForExactETHRequest{})
	require.EqualError(t, err, "uniswapv2: callSwapTokensForExactETH: "+(&SwapTokensForExactETHRequest{}).Validate().Error())
	_, err = r.CallSwapExactTokensForETH(ctx, nil, &SwapExactTokensForETHRequest{})
	require.EqualError(t, err, "uniswapv2: callSwapExactTokensForETH: "+(&SwapExactTokensForETHRequest{}).Validate().Error())
	_, err = r.CallSwapETHForExactTokens(ctx, nil, &SwapETHForExactTokensRequest{})
	require.EqualError(t, err, "uniswapv2: callSwapETHForExactTokens: "+(&SwapETHForExactTokensRequest{}).Validate().Error())
	_, err = r.CallRemoveLiquidityETHSupportingFeeOnTransferTokens(ctx, nil, &RemoveLiquidityETHSupportingFeeOnTransferTokensRequest{})
	require.EqualError(t, err, "uniswapv2: callRemoveLiquidityETHSupportingFeeOnTransferTokens: "+(&RemoveLiquidityETHSupportingFeeOnTransferTokensRequest{}).Validate().Error())
	_, err = r.CallRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(ctx, nil, &RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest{})
	require.EqualError(t, err, "uniswapv2: callRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens: "+(&RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest{}).Validate().Error())
	_, _, err = r.AddLiquidity(ctx, opts, &AddLiquidityRequest{})
	require.EqualError(t, err, "uniswapv2: addLiquidity: "+(&AddLiquidityRequest{}).Validate().Error())
	_, _, err = r.AddLiquidityETH(ctx, opts, &AddLiquidityETHRequest{})
	require.EqualError(t, err, "uniswapv2: addLiquidityETH: "+(&AddLiquidityETHRequest{}).Validate().Error())
	_, _, err = r.RemoveLiquidity(ctx, opts, &RemoveLiquidityRequest{})
	require.EqualError(t, err, "uniswapv2: removeLiquidity: "+(&RemoveLiquidityRequest{}).Validate().Error())
	_, _, err = r.RemoveLiquidityETH(ctx, opts, &RemoveLiquidityETHRequest{})
	require.EqualError(t, err, "uniswapv2: removeLiquidityETH: "+(&RemoveLiquidityETHRequest{}).Validate().Error())
	_, _, err = r.RemoveLiquidityWithPermit(ctx, opts, &RemoveLiquidityWithPermitRequest{})
	require.EqualError(t, err, "uniswapv2: removeLiquidityWithPermit: "+(&RemoveLiquidityWithPermitRequest{}).Validate().Error())
	_, _, err = r.RemoveLiquidityETHWithPermit(ctx, opts, &RemoveLiquidityETHWithPermitRequest{})
	require.EqualError(t, err, "uniswapv2: removeLiquidityETHWithPermit: "+(&RemoveLiquidityETHWithPermitRequest{}).Validate().Error())
	_, _, err = r.SwapExactTokensForTokens(ctx, opts, &SwapExactTokensForTokensRequest{})
	require.EqualError(t, err, "uniswapv2: swapExactTokensForTokens: "+(&SwapExactTokensForTokensRequest{}).Validate().Error())
	_, _, err = r.SwapTokensForExactTokens(ctx, opts, &SwapTokensForExactTokensRequest{})
	require.EqualError(t, err, "uniswapv2: swapTokensForExactTokens: "+(&SwapTokensForExactTokensRequest{}).Validate().Error())
	_, _, err = r.SwapExactETHForTokens(ctx, opts, &SwapExactETHForTokensRequest{})
	require.EqualError(t, err, "uniswapv2: swapExactETHForTokens: "+(&SwapExactETHForTokensRequest{}).Validate().Error())
	_, _, err = r.SwapTokensForExactETH(ctx, opts, &SwapTokensForExactETHRequest{})
	require.EqualError(t, err, "uniswapv2: swapTokensForExactETH: "+(&SwapTokensForExactETHRequest{}).Validate().Error())
	_, _, err = r.SwapExactTokensForETH(ctx, opts, &SwapExactTokensForETHRequest{})
	require.EqualError(t, err, "uniswapv2: swapExactTokensForETH: "+(&SwapExactTokensForETHRequest{}).Validate().Error())
	_, _, err = r.SwapETHForExactTokens(ctx, opts, &SwapETHForExactTokensRequest{})
	require.EqualError(t, err, "uniswapv2: swapETHForExactTokens: "+(&SwapETHForExactTokensRequest{}).Validate().Error())
	_, _, err = r.RemoveLiquidityETHSupportingFeeOnTransferTokens(ctx, opts, &RemoveLiquidityETHSupportingFeeOnTransferTokensRequest{})
	require.EqualError(t, err, "uniswapv2: removeLiquidityETHSupportingFeeOnTransferTokens: "+(&RemoveLiquidityETHSupportingFeeOnTransferTokensRequest{}).Validate().Error())
	_, _, err = r.RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(ctx, opts, &RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest{})
	require.EqualError(t, err, "uniswapv2: removeLiquidityETHWithPermitSupportingFeeOnTransferTokens: "+(&RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest{}).Validate().Error())
	_, _, err = r.SwapExactTokensForTokensSupportingFeeOnTransferTokens(ctx, opts, &SwapExactTokensForTokensSupportingFeeOnTransferTokensRequest{})
	require.EqualError(t, err, "uniswapv2: swapExactTokensForTokensSupportingFeeOnTransferTokens: "+(&SwapExactTokensForTokensSupportingFeeOnTransferTokensRequest{}).Validate().Error())
	_, _, err = r.SwapExactETHForTokensSupportingFeeOnTransferTokens(ctx, opts, &SwapExactETHForTokensSupportingFeeOnTransferTokensRequest{})
	require.EqualError(t, err, "uniswapv2: swapExactETHForTokensSupportingFeeOnTransferTokens: "+(&SwapExactETHForTokensSupportingFeeOnTransferTokensRequest{}).Validate().Error())
	_, _, err = r.SwapExactTokensForETHSupportingFeeOnTransferTokens(ctx, opts, &SwapExactTokensForETHSupportingFeeOnTransferTokensRequest{})
	require.EqualError(t, err, "uniswapv2: swapExactTokensForETHSupportingFeeOnTransferTokens: "+(&SwapExactTokensForETHSupportingFeeOnTransferTokensRequest{}).Validate().Error())
	require.ErrorIs(t, err, bind.ErrInvalidRequest)

	_, _, err = r.SwapExactTokensForTokens(ctx, opts, NewSwapExactTokensForTokensRequest(router, big.NewInt(1), big.NewInt(0), []*types.Address{a, nil}, to, big.NewInt(1)))
	require.EqualError(t, err, "uniswapv2: swapExactTokensForTokens: invalid request: path[1] is nil")

	_, _, err = r.SwapExactTokensForTokens(ctx, opts, NewSwapExactTokensForTokensRequest(router, big.NewInt(-1), big.NewInt(0), path, to, big.NewInt(1)))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)

	_, _, err = r.AddLiquidity(ctx, nil, NewAddLiquidityRequest(router, a, b, big.NewInt(1), big.NewInt(1), big.NewInt(0), big.NewInt(0), to, big.NewInt(1)))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv2: addLiquidity: bind: transact: invalid options: options is nil")

	_, _, err = r.SwapExactETHForTokens(ctx, nil, NewSwapExactETHForTokensRequest(router, big.NewInt(0), path, to, big.NewInt(1), big.NewInt(1)))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv2: swapExactETHForTokens: bind: transact: invalid options: options is nil")

	_, err = r.CallAddLiquidityETH(ctx, bind.NewCallOptions().WithValue(big.NewInt(1)), NewAddLiquidityETHRequest(router, a, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000), big.NewInt(4007)))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv2: callAddLiquidityETH: invalid options: value must be nil")
	_, _, err = r.AddLiquidityETH(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewAddLiquidityETHRequest(router, a, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000), big.NewInt(4007)))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv2: addLiquidityETH: invalid options: value must be nil")
	_, err = r.CallSwapExactETHForTokens(ctx, bind.NewCallOptions().WithValue(big.NewInt(1)), NewSwapExactETHForTokensRequest(router, big.NewInt(1007), path, to, big.NewInt(1_700_000_000), big.NewInt(2007)))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv2: callSwapExactETHForTokens: invalid options: value must be nil")
	_, _, err = r.SwapExactETHForTokens(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewSwapExactETHForTokensRequest(router, big.NewInt(1007), path, to, big.NewInt(1_700_000_000), big.NewInt(2007)))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv2: swapExactETHForTokens: invalid options: value must be nil")
	_, err = r.CallSwapETHForExactTokens(ctx, bind.NewCallOptions().WithValue(big.NewInt(1)), NewSwapETHForExactTokensRequest(router, big.NewInt(1007), path, to, big.NewInt(1_700_000_000), big.NewInt(2007)))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv2: callSwapETHForExactTokens: invalid options: value must be nil")
	_, _, err = r.SwapETHForExactTokens(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewSwapETHForExactTokensRequest(router, big.NewInt(1007), path, to, big.NewInt(1_700_000_000), big.NewInt(2007)))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv2: swapETHForExactTokens: invalid options: value must be nil")
	_, _, err = r.SwapExactETHForTokensSupportingFeeOnTransferTokens(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewSwapExactETHForTokensSupportingFeeOnTransferTokensRequest(router, big.NewInt(1007), path, to, big.NewInt(1_700_000_000), big.NewInt(2007)))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv2: swapExactETHForTokensSupportingFeeOnTransferTokens: invalid options: value must be nil")
}

func TestRouterWritesSignAndSend(t *testing.T) {
	router := types.MustNewAddressFromHex("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	to := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	path := []*types.Address{a, b}
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sig, err := key.Sign(core.Keccak256([]byte("permit")))
	require.NoError(t, err)
	sender := core.PubkeyToAddress(key.PublicKey())

	addLiquidityData, err := EncodeAddLiquidity(a, b, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), big.NewInt(4007), to, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	addLiquidityETHData, err := EncodeAddLiquidityETH(a, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	removeLiquidityData, err := EncodeRemoveLiquidity(a, b, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	removeLiquidityETHData, err := EncodeRemoveLiquidityETH(a, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	removeLiquidityWithPermitData, err := EncodeRemoveLiquidityWithPermit(a, b, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000), true, sig)
	require.NoError(t, err)
	removeLiquidityETHWithPermitData, err := EncodeRemoveLiquidityETHWithPermit(a, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000), true, sig)
	require.NoError(t, err)
	swapExactTokensForTokensData, err := EncodeSwapExactTokensForTokens(big.NewInt(1007), big.NewInt(2007), path, to, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	swapTokensForExactTokensData, err := EncodeSwapTokensForExactTokens(big.NewInt(1007), big.NewInt(2007), path, to, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	swapExactETHForTokensData, err := EncodeSwapExactETHForTokens(big.NewInt(1007), path, to, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	swapTokensForExactETHData, err := EncodeSwapTokensForExactETH(big.NewInt(1007), big.NewInt(2007), path, to, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	swapExactTokensForETHData, err := EncodeSwapExactTokensForETH(big.NewInt(1007), big.NewInt(2007), path, to, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	swapETHForExactTokensData, err := EncodeSwapETHForExactTokens(big.NewInt(1007), path, to, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	removeLiquidityETHSupportingFeeOnTransferTokensData, err := EncodeRemoveLiquidityETHSupportingFeeOnTransferTokens(a, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	removeLiquidityETHWithPermitSupportingFeeOnTransferTokensData, err := EncodeRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(a, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000), true, sig)
	require.NoError(t, err)
	swapExactTokensForTokensSupportingFeeOnTransferTokensData, err := EncodeSwapExactTokensForTokensSupportingFeeOnTransferTokens(big.NewInt(1007), big.NewInt(2007), path, to, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	swapExactETHForTokensSupportingFeeOnTransferTokensData, err := EncodeSwapExactETHForTokensSupportingFeeOnTransferTokens(big.NewInt(1007), path, to, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	swapExactTokensForETHSupportingFeeOnTransferTokensData, err := EncodeSwapExactTokensForETHSupportingFeeOnTransferTokens(big.NewInt(1007), big.NewInt(2007), path, to, big.NewInt(1_700_000_000))
	require.NoError(t, err)

	for _, tc := range []struct {
		name      string
		calldata  []byte
		wantValue *big.Int
		send      func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error)
	}{
		{
			"addLiquidity",
			addLiquidityData,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.AddLiquidity(t.Context(), opts, NewAddLiquidityRequest(router, a, b, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), big.NewInt(4007), to, big.NewInt(1_700_000_000)))
			},
		},
		{
			"addLiquidityETH",
			addLiquidityETHData,
			big.NewInt(4007),
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.AddLiquidityETH(t.Context(), opts, NewAddLiquidityETHRequest(router, a, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000), big.NewInt(4007)))
			},
		},
		{
			"removeLiquidity",
			removeLiquidityData,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.RemoveLiquidity(t.Context(), opts, NewRemoveLiquidityRequest(router, a, b, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000)))
			},
		},
		{
			"removeLiquidityETH",
			removeLiquidityETHData,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.RemoveLiquidityETH(t.Context(), opts, NewRemoveLiquidityETHRequest(router, a, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000)))
			},
		},
		{
			"removeLiquidityWithPermit",
			removeLiquidityWithPermitData,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.RemoveLiquidityWithPermit(t.Context(), opts, NewRemoveLiquidityWithPermitRequest(router, a, b, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000), true, sig))
			},
		},
		{
			"removeLiquidityETHWithPermit",
			removeLiquidityETHWithPermitData,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.RemoveLiquidityETHWithPermit(t.Context(), opts, NewRemoveLiquidityETHWithPermitRequest(router, a, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000), true, sig))
			},
		},
		{
			"swapExactTokensForTokens",
			swapExactTokensForTokensData,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.SwapExactTokensForTokens(t.Context(), opts, NewSwapExactTokensForTokensRequest(router, big.NewInt(1007), big.NewInt(2007), path, to, big.NewInt(1_700_000_000)))
			},
		},
		{
			"swapTokensForExactTokens",
			swapTokensForExactTokensData,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.SwapTokensForExactTokens(t.Context(), opts, NewSwapTokensForExactTokensRequest(router, big.NewInt(1007), big.NewInt(2007), path, to, big.NewInt(1_700_000_000)))
			},
		},
		{
			"swapExactETHForTokens",
			swapExactETHForTokensData,
			big.NewInt(2007),
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.SwapExactETHForTokens(t.Context(), opts, NewSwapExactETHForTokensRequest(router, big.NewInt(1007), path, to, big.NewInt(1_700_000_000), big.NewInt(2007)))
			},
		},
		{
			"swapTokensForExactETH",
			swapTokensForExactETHData,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.SwapTokensForExactETH(t.Context(), opts, NewSwapTokensForExactETHRequest(router, big.NewInt(1007), big.NewInt(2007), path, to, big.NewInt(1_700_000_000)))
			},
		},
		{
			"swapExactTokensForETH",
			swapExactTokensForETHData,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.SwapExactTokensForETH(t.Context(), opts, NewSwapExactTokensForETHRequest(router, big.NewInt(1007), big.NewInt(2007), path, to, big.NewInt(1_700_000_000)))
			},
		},
		{
			"swapETHForExactTokens",
			swapETHForExactTokensData,
			big.NewInt(2007),
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.SwapETHForExactTokens(t.Context(), opts, NewSwapETHForExactTokensRequest(router, big.NewInt(1007), path, to, big.NewInt(1_700_000_000), big.NewInt(2007)))
			},
		},
		{
			"removeLiquidityETHSupportingFeeOnTransferTokens",
			removeLiquidityETHSupportingFeeOnTransferTokensData,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.RemoveLiquidityETHSupportingFeeOnTransferTokens(t.Context(), opts, NewRemoveLiquidityETHSupportingFeeOnTransferTokensRequest(router, a, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000)))
			},
		},
		{
			"removeLiquidityETHWithPermitSupportingFeeOnTransferTokens",
			removeLiquidityETHWithPermitSupportingFeeOnTransferTokensData,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(t.Context(), opts, NewRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest(router, a, big.NewInt(1007), big.NewInt(2007), big.NewInt(3007), to, big.NewInt(1_700_000_000), true, sig))
			},
		},
		{
			"swapExactTokensForTokensSupportingFeeOnTransferTokens",
			swapExactTokensForTokensSupportingFeeOnTransferTokensData,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.SwapExactTokensForTokensSupportingFeeOnTransferTokens(t.Context(), opts, NewSwapExactTokensForTokensSupportingFeeOnTransferTokensRequest(router, big.NewInt(1007), big.NewInt(2007), path, to, big.NewInt(1_700_000_000)))
			},
		},
		{
			"swapExactETHForTokensSupportingFeeOnTransferTokens",
			swapExactETHForTokensSupportingFeeOnTransferTokensData,
			big.NewInt(2007),
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.SwapExactETHForTokensSupportingFeeOnTransferTokens(t.Context(), opts, NewSwapExactETHForTokensSupportingFeeOnTransferTokensRequest(router, big.NewInt(1007), path, to, big.NewInt(1_700_000_000), big.NewInt(2007)))
			},
		},
		{
			"swapExactTokensForETHSupportingFeeOnTransferTokens",
			swapExactTokensForETHSupportingFeeOnTransferTokensData,
			nil,
			func(r *Router, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return r.SwapExactTokensForETHSupportingFeeOnTransferTokens(t.Context(), opts, NewSwapExactTokensForETHSupportingFeeOnTransferTokensRequest(router, big.NewInt(1007), big.NewInt(2007), path, to, big.NewInt(1_700_000_000)))
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
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				require.Equal(t, rpc.MethodETHSendRawTransaction, e.Method)
				raw, err = encoding.Hex.Decode(e.Params.([]any)[0].(string))
				require.NoError(t, err)
				*e.Result.(*string) = core.Keccak256(raw).String()
				return nil
			}).Once()

			opts := bind.NewTransactOptions(key, builder)
			signed, hash, err := tc.send(NewRouter(c), opts)
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
	router := types.MustNewAddressFromHex("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	to := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	path := []*types.Address{a, b}
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)

	packed, err := abi.Pack(abi.NewTypes(abi.String), reasonExpired)
	require.NoError(t, err)
	revert := append([]byte{0x08, 0xc3, 0x79, 0xa0}, packed...)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(7)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	c := mocks.NewClient(t)
	c.EXPECT().Batch(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, elems rpc.Elements) error {
		require.Equal(t, 1, elems.Len())
		require.Equal(t, rpc.MethodETHEstimateGas, elems.GetMethod(0))
		return fmt.Errorf("rpc: %s: %w", rpc.MethodETHEstimateGas, &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(revert) + `"`)})
	}).Once()

	signed, hash, err := NewRouter(c).SwapETHForExactTokens(t.Context(), bind.NewTransactOptions(key, builder), NewSwapETHForExactTokensRequest(router, big.NewInt(1), path, to, big.NewInt(1), big.NewInt(5)))
	require.ErrorIs(t, err, ErrExpired)
	require.NotErrorIs(t, err, ErrUnknown)
	require.ErrorContains(t, err, "uniswapv2: swapETHForExactTokens: ")
	require.Nil(t, signed)
	require.Nil(t, hash)
}

func TestRouterWritesReturnSignedTxOnBroadcastFailure(t *testing.T) {
	router := types.MustNewAddressFromHex("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	to := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(9)
	builder.Limit(3_000_000)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: -32000, Message: "nonce too low"}).Once()

	signed, hash, err := NewRouter(c).AddLiquidityETH(t.Context(), bind.NewTransactOptions(key, builder), NewAddLiquidityETHRequest(router, a, big.NewInt(1000), big.NewInt(0), big.NewInt(0), to, big.NewInt(1), big.NewInt(777)))
	require.ErrorContains(t, err, "uniswapv2: addLiquidityETH: ")
	require.ErrorContains(t, err, "nonce too low")
	require.NotErrorIs(t, err, ErrUnknown)
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
	router := types.MustNewAddressFromHex("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	to := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	path := []*types.Address{a, b}
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	word, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(1))
	require.NoError(t, err)

	cli := mocks.NewClient(t)
	cli.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(got context.Context, e rpc.Element) error {
		require.Equal(t, "router", got.Value(marker{}))
		if e.Method == rpc.MethodETHCall {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(word)
		}
		return nil
	}).Twice()

	_, err = NewRouter(cli).Quote(ctx, nil, NewQuoteRequest(router, big.NewInt(1), big.NewInt(1), big.NewInt(1)))
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(1)
	builder.Limit(3_000_000)
	builder.Price(big.NewInt(2), big.NewInt(1))
	_, _, err = NewRouter(cli).SwapExactETHForTokens(ctx, bind.NewTransactOptions(key, builder), NewSwapExactETHForTokensRequest(router, big.NewInt(0), path, to, big.NewInt(1), big.NewInt(1)))
	require.NoError(t, err)
}

func TestMapRevertErrorCoversRouterAndPair(t *testing.T) {
	for _, tc := range []struct {
		reason string
		want   error
	}{
		{"UniswapV2Router: EXPIRED", ErrExpired},
		{"UniswapV2Library: INVALID_PATH", ErrInvalidPath},
		{"UniswapV2: INSUFFICIENT_OUTPUT_AMOUNT", ErrInsufficientOutputAmount},
		{"UniswapV2: K", ErrK},
		{"UniswapV2: LOCKED", ErrLocked},
		{"ds-math-sub-underflow", ErrSubUnderflow},
	} {
		t.Run(tc.reason, func(t *testing.T) {
			packed, err := abi.Pack(abi.NewTypes(abi.String), tc.reason)
			require.NoError(t, err)
			rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"0x08c379a0` + encoding.Hex.Encode(packed) + `"`)}

			mapped := MapRevertError(rpcErr)
			require.ErrorIs(t, mapped, tc.want)
			require.NotErrorIs(t, mapped, ErrUnknown)
			got, ok := errors.AsType[*rpc.ResponseError](mapped)
			require.True(t, ok)
			require.Same(t, rpcErr, got)
		})
	}

	packed, err := abi.Pack(abi.NewTypes(abi.String), "LOK")
	require.NoError(t, err)
	require.ErrorIs(t, MapRevertError(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"0x08c379a0` + encoding.Hex.Encode(packed) + `"`)}), ErrUnknown)

	custom := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"0xdeadbeef"`)}
	require.Same(t, custom, MapRevertError(custom))
	plain := errors.New("dial tcp: connection refused")
	require.Same(t, plain, MapRevertError(plain))
}
