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

func TestNFPMReadsSendCallAndDecode(t *testing.T) {
	nfpm := types.MustNewAddressFromHex("0xC36442b4a4522E871399CD717aBDD847Ab11FE88")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	pool := types.MustNewAddressFromHex("0x8ad599c3a0ff1de082011efddc58f1908eb6e6d8")
	opts := bind.NewCallOptions().WithFrom(a).WithBlock("0x1b4")
	var hashWord [32]byte
	hashWord[0], hashWord[31] = 0x49, 0xec

	result0, err := nfpmFactoryFn.EncodeReturn(a)
	require.NoError(t, err)
	result1, err := nfpmWeth9Fn.EncodeReturn(b)
	require.NoError(t, err)
	result2, err := totalSupplyFn.EncodeReturn(big.NewInt(42))
	require.NoError(t, err)
	result3, err := tokenOfOwnerByIndexFn.EncodeReturn(big.NewInt(99))
	require.NoError(t, err)
	tokenOfOwnerByIndexData, err := EncodeTokenOfOwnerByIndex(b, big.NewInt(1))
	require.NoError(t, err)
	result4, err := tokenByIndexFn.EncodeReturn(big.NewInt(77))
	require.NoError(t, err)
	tokenByIndexData, err := EncodeTokenByIndex(big.NewInt(3))
	require.NoError(t, err)
	result5, err := permitTypehashFn.EncodeReturn(hashWord[:])
	require.NoError(t, err)
	result6, err := domainSeparatorFn.EncodeReturn(hashWord[:])
	require.NoError(t, err)
	result7, err := nfpmPositionsFn.EncodeReturn(big.NewInt(1), a, b, a, big.NewInt(3000), big.NewInt(-60), big.NewInt(60), big.NewInt(1000), big.NewInt(2), big.NewInt(3), big.NewInt(4), big.NewInt(5))
	require.NoError(t, err)
	positionsData, err := EncodeNfpmPositions(big.NewInt(7))
	require.NoError(t, err)
	result8, err := createAndInitializePoolIfNecessaryFn.EncodeReturn(pool)
	require.NoError(t, err)
	createData, err := EncodeCreateAndInitializePoolIfNecessary(a, b, big.NewInt(3000), new(big.Int).Lsh(big.NewInt(1), 96))
	require.NoError(t, err)
	result9, err := nfpmMulticallFn.EncodeReturn([][]byte{{0x01}, {0x02, 0x03}})
	require.NoError(t, err)
	multicallData, err := EncodeMulticall([][]byte{{0xaa}, {0xbb, 0xcc}})
	require.NoError(t, err)
	result10, err := nfpmMulticallFn.EncodeReturn([][]byte{})
	require.NoError(t, err)
	result11, err := nfpmMintFn.EncodeReturn(big.NewInt(1), big.NewInt(2), big.NewInt(3), big.NewInt(4))
	require.NoError(t, err)
	mintData, err := EncodeNfpmMint(a, b, big.NewInt(3000), big.NewInt(-60), big.NewInt(60), big.NewInt(1000), big.NewInt(2000), big.NewInt(0), big.NewInt(0), b, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	result12, err := increaseLiquidityFn.EncodeReturn(big.NewInt(5), big.NewInt(6), big.NewInt(7))
	require.NoError(t, err)
	increaseData, err := EncodeIncreaseLiquidity(big.NewInt(7), big.NewInt(1000), big.NewInt(2000), big.NewInt(0), big.NewInt(0), big.NewInt(1_700_000_000))
	require.NoError(t, err)
	result13, err := decreaseLiquidityFn.EncodeReturn(big.NewInt(8), big.NewInt(9))
	require.NoError(t, err)
	decreaseData, err := EncodeDecreaseLiquidity(big.NewInt(7), big.NewInt(100), big.NewInt(0), big.NewInt(0), big.NewInt(1_700_000_000))
	require.NoError(t, err)
	result14, err := nfpmCollectFn.EncodeReturn(big.NewInt(5), big.NewInt(6))
	require.NoError(t, err)
	collectData, err := EncodeNfpmCollect(big.NewInt(7), b, big.NewInt(5), big.NewInt(6))
	require.NoError(t, err)

	for _, tc := range []struct {
		name      string
		calldata  []byte
		result    []byte
		wantValue string
		call      func(n *NFPM) (any, error)
		want      any
	}{
		{
			"factory",
			EncodeNfpmFactory(),
			result0,
			"",
			func(n *NFPM) (any, error) {
				return n.Factory(t.Context(), opts, NewFactoryRequest(nfpm))
			},
			a,
		},
		{
			"weth9",
			EncodeWETH9(),
			result1,
			"",
			func(n *NFPM) (any, error) {
				return n.WETH9(t.Context(), opts, NewWETH9Request(nfpm))
			},
			b,
		},
		{
			"totalSupply",
			EncodeTotalSupply(),
			result2,
			"",
			func(n *NFPM) (any, error) {
				return n.TotalSupply(t.Context(), opts, NewTotalSupplyRequest(nfpm))
			},
			big.NewInt(42),
		},
		{
			"tokenOfOwnerByIndex",
			tokenOfOwnerByIndexData,
			result3,
			"",
			func(n *NFPM) (any, error) {
				return n.TokenOfOwnerByIndex(t.Context(), opts, NewTokenOfOwnerByIndexRequest(nfpm, b, big.NewInt(1)))
			},
			big.NewInt(99),
		},
		{
			"tokenByIndex",
			tokenByIndexData,
			result4,
			"",
			func(n *NFPM) (any, error) {
				return n.TokenByIndex(t.Context(), opts, NewTokenByIndexRequest(nfpm, big.NewInt(3)))
			},
			big.NewInt(77),
		},
		{
			"permitTypehash",
			EncodePermitTypehash(),
			result5,
			"",
			func(n *NFPM) (any, error) {
				return n.PermitTypehash(t.Context(), opts, NewPermitTypehashRequest(nfpm))
			},
			hashWord,
		},
		{
			"domainSeparator",
			EncodeDomainSeparator(),
			result6,
			"",
			func(n *NFPM) (any, error) {
				return n.DomainSeparator(t.Context(), opts, NewDomainSeparatorRequest(nfpm))
			},
			hashWord,
		},
		{
			"positions",
			positionsData,
			result7,
			"",
			func(n *NFPM) (any, error) {
				return n.Positions(t.Context(), opts, NewNFPMPositionsRequest(nfpm, big.NewInt(7)))
			},
			NewPosition(big.NewInt(1), a, b, a, big.NewInt(3000), big.NewInt(-60), big.NewInt(60), big.NewInt(1000), big.NewInt(2), big.NewInt(3), big.NewInt(4), big.NewInt(5)),
		},
		{
			"callCreateAndInitializePoolIfNecessary",
			createData,
			result8,
			"",
			func(n *NFPM) (any, error) {
				return n.CallCreateAndInitializePoolIfNecessary(t.Context(), opts, NewCreateAndInitializePoolIfNecessaryRequest(nfpm, a, b, big.NewInt(3000), new(big.Int).Lsh(big.NewInt(1), 96)))
			},
			pool,
		},
		{
			"callMulticall",
			multicallData,
			result9,
			"0x2a",
			func(n *NFPM) (any, error) {
				return n.CallMulticall(t.Context(), opts, NewNFPMMulticallRequest(nfpm, [][]byte{{0xaa}, {0xbb, 0xcc}}, big.NewInt(42)))
			},
			[][]byte{{0x01}, {0x02, 0x03}},
		},
		{
			"callMulticall without native",
			multicallData,
			result10,
			"",
			func(n *NFPM) (any, error) {
				return n.CallMulticall(t.Context(), opts, NewNFPMMulticallRequest(nfpm, [][]byte{{0xaa}, {0xbb, 0xcc}}, nil))
			},
			[][]byte{},
		},
		{
			"callMint",
			mintData,
			result11,
			"0x3e8",
			func(n *NFPM) (any, error) {
				tokenID, liquidity, amount0, amount1, err := n.CallMint(t.Context(), opts, NewNFPMMintRequest(nfpm, NewMintParams(a, b, big.NewInt(3000), big.NewInt(-60), big.NewInt(60), big.NewInt(1000), big.NewInt(2000), big.NewInt(0), big.NewInt(0), b, big.NewInt(1_700_000_000)), big.NewInt(1000)))
				return []*big.Int{tokenID, liquidity, amount0, amount1}, err
			},
			[]*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3), big.NewInt(4)},
		},
		{
			"callIncreaseLiquidity",
			increaseData,
			result12,
			"",
			func(n *NFPM) (any, error) {
				liquidity, amount0, amount1, err := n.CallIncreaseLiquidity(t.Context(), opts, NewIncreaseLiquidityRequest(nfpm, NewIncreaseLiquidityParams(big.NewInt(7), big.NewInt(1000), big.NewInt(2000), big.NewInt(0), big.NewInt(0), big.NewInt(1_700_000_000)), nil))
				return []*big.Int{liquidity, amount0, amount1}, err
			},
			[]*big.Int{big.NewInt(5), big.NewInt(6), big.NewInt(7)},
		},
		{
			"callDecreaseLiquidity",
			decreaseData,
			result13,
			"",
			func(n *NFPM) (any, error) {
				amount0, amount1, err := n.CallDecreaseLiquidity(t.Context(), opts, NewDecreaseLiquidityRequest(nfpm, NewDecreaseLiquidityParams(big.NewInt(7), big.NewInt(100), big.NewInt(0), big.NewInt(0), big.NewInt(1_700_000_000))))
				return []*big.Int{amount0, amount1}, err
			},
			[]*big.Int{big.NewInt(8), big.NewInt(9)},
		},
		{
			"callCollect",
			collectData,
			result14,
			"",
			func(n *NFPM) (any, error) {
				amount0, amount1, err := n.CallCollect(t.Context(), opts, NewNFPMCollectRequest(nfpm, NewCollectParams(big.NewInt(7), b, big.NewInt(5), big.NewInt(6))))
				return []*big.Int{amount0, amount1}, err
			},
			[]*big.Int{big.NewInt(5), big.NewInt(6)},
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
				require.Equal(t, nfpm.String(), call["to"])
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

			got, err := tc.call(NewNFPM(c))
			require.NoError(t, err)
			require.Equal(t, fmt.Sprint(tc.want), fmt.Sprint(got))
		})
	}
}

func TestNFPMReadsMapRevertToSentinel(t *testing.T) {
	nfpm := types.MustNewAddressFromHex("0xC36442b4a4522E871399CD717aBDD847Ab11FE88")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	errorSel := []byte{0x08, 0xc3, 0x79, 0xa0}
	all := []error{ErrTransactionTooOld, ErrInsufficientWETH9, ErrInsufficientToken, ErrNotWETH9, ErrTransferFromFailed, ErrTokenTransferFailed, ErrApproveFailed, ErrTransferETHFailed, ErrInvalidTokenID, ErrNotApproved, ErrPriceSlippageCheck, ErrNotCleared, ErrPermitExpired, ErrApprovalToCurrentOwner, ErrUnauthorized, ErrInvalidSignature, ErrReverted}

	for _, tc := range []struct {
		name   string
		reason string
		want   error
	}{
		{"transaction too old", reasonTransactionTooOld, ErrTransactionTooOld},
		{"insufficient weth9", reasonInsufficientWETH9, ErrInsufficientWETH9},
		{"insufficient token", reasonInsufficientToken, ErrInsufficientToken},
		{"not weth9", reasonNotWETH9, ErrNotWETH9},
		{"transfer from failed", reasonTransferFromFailed, ErrTransferFromFailed},
		{"token transfer failed", reasonTokenTransferFailed, ErrTokenTransferFailed},
		{"approve failed", reasonApproveFailed, ErrApproveFailed},
		{"transfer eth failed", reasonTransferETHFailed, ErrTransferETHFailed},
		{"invalid token id", reasonInvalidTokenID, ErrInvalidTokenID},
		{"not approved", reasonNotApproved, ErrNotApproved},
		{"price slippage check", reasonPriceSlippageCheck, ErrPriceSlippageCheck},
		{"not cleared", reasonNotCleared, ErrNotCleared},
		{"permit expired", reasonPermitExpired, ErrPermitExpired},
		{"approval to current owner", reasonApprovalToCurrentOwner, ErrApprovalToCurrentOwner},
		{"unauthorized", reasonUnauthorized, ErrUnauthorized},
		{"invalid signature", reasonInvalidSignature, ErrInvalidSignature},
		{"other reason", "XYZ", ErrReverted},
	} {
		t.Run(tc.name, func(t *testing.T) {
			packed, err := abi.Pack(abi.NewTypes(abi.String), tc.reason)
			require.NoError(t, err)
			revert := append(append([]byte{}, errorSel...), packed...)

			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(revert) + `"`)}).Twice()
			n := NewNFPM(c)

			amount0, amount1, err := n.CallCollect(t.Context(), bind.NewCallOptions().WithFrom(a), NewNFPMCollectRequest(nfpm, NewCollectParams(big.NewInt(7), b, big.NewInt(1), big.NewInt(1))))
			require.Nil(t, amount0)
			require.Nil(t, amount1)
			require.ErrorIs(t, err, tc.want)
			for _, other := range all {
				if other != tc.want {
					require.NotErrorIs(t, err, other)
				}
			}
			var rpcErr *rpc.ResponseError
			require.ErrorAs(t, err, &rpcErr)

			position, err := n.Positions(t.Context(), nil, NewNFPMPositionsRequest(nfpm, big.NewInt(7)))
			require.Nil(t, position)
			require.ErrorIs(t, err, tc.want)
			require.ErrorContains(t, err, "uniswapv3: positions: ")
		})
	}

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(errors.New("dial tcp: connection refused")).Once()

	total, err := NewNFPM(c).TotalSupply(t.Context(), nil, NewTotalSupplyRequest(nfpm))
	require.Nil(t, total)
	require.EqualError(t, err, "uniswapv3: totalSupply: dial tcp: connection refused")
	require.NotErrorIs(t, err, ErrReverted)
}

func TestNFPMReadsRejectMalformedResult(t *testing.T) {
	nfpm := types.MustNewAddressFromHex("0xC36442b4a4522E871399CD717aBDD847Ab11FE88")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	for _, result := range []string{"0x", "0xzz", "0x01", "0x00", "not hex"} {
		t.Run(result, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				*e.Result.(*string) = result
				return nil
			}).Times(4)
			n := NewNFPM(c)

			position, err := n.Positions(t.Context(), nil, NewNFPMPositionsRequest(nfpm, big.NewInt(7)))
			require.Nil(t, position)
			require.ErrorContains(t, err, "uniswapv3: positions: ")

			typehash, err := n.PermitTypehash(t.Context(), nil, NewPermitTypehashRequest(nfpm))
			require.Equal(t, [32]byte{}, typehash)
			require.ErrorContains(t, err, "uniswapv3: permitTypehash: ")

			tokenID, liquidity, amount0, amount1, err := n.CallMint(t.Context(), nil, NewNFPMMintRequest(nfpm, NewMintParams(a, b, big.NewInt(3000), big.NewInt(-60), big.NewInt(60), big.NewInt(1), big.NewInt(1), big.NewInt(0), big.NewInt(0), b, big.NewInt(1)), nil))
			require.Nil(t, tokenID)
			require.Nil(t, liquidity)
			require.Nil(t, amount0)
			require.Nil(t, amount1)
			require.ErrorContains(t, err, "uniswapv3: callMint: ")

			results, err := n.CallMulticall(t.Context(), nil, NewNFPMMulticallRequest(nfpm, [][]byte{{0x01}}, nil))
			require.Nil(t, results)
			require.ErrorContains(t, err, "uniswapv3: callMulticall: ")
		})
	}
}

func TestNFPMMethodsRejectInvalidRequestWithoutRPC(t *testing.T) {
	nfpm := types.MustNewAddressFromHex("0xC36442b4a4522E871399CD717aBDD847Ab11FE88")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	opts := bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder())
	n := NewNFPM(mocks.NewClient(t))
	ctx := t.Context()

	_, err = n.Factory(ctx, nil, &FactoryRequest{})
	require.EqualError(t, err, "uniswapv3: factory: "+(&FactoryRequest{}).Validate().Error())
	_, err = n.WETH9(ctx, nil, &WETH9Request{})
	require.EqualError(t, err, "uniswapv3: weth9: "+(&WETH9Request{}).Validate().Error())
	_, err = n.TotalSupply(ctx, nil, &TotalSupplyRequest{})
	require.EqualError(t, err, "uniswapv3: totalSupply: "+(&TotalSupplyRequest{}).Validate().Error())
	_, err = n.TokenOfOwnerByIndex(ctx, nil, &TokenOfOwnerByIndexRequest{})
	require.EqualError(t, err, "uniswapv3: tokenOfOwnerByIndex: "+(&TokenOfOwnerByIndexRequest{}).Validate().Error())
	_, err = n.TokenByIndex(ctx, nil, &TokenByIndexRequest{})
	require.EqualError(t, err, "uniswapv3: tokenByIndex: "+(&TokenByIndexRequest{}).Validate().Error())
	_, err = n.PermitTypehash(ctx, nil, &PermitTypehashRequest{})
	require.EqualError(t, err, "uniswapv3: permitTypehash: "+(&PermitTypehashRequest{}).Validate().Error())
	_, err = n.DomainSeparator(ctx, nil, &DomainSeparatorRequest{})
	require.EqualError(t, err, "uniswapv3: domainSeparator: "+(&DomainSeparatorRequest{}).Validate().Error())
	_, err = n.Positions(ctx, nil, &NFPMPositionsRequest{})
	require.EqualError(t, err, "uniswapv3: positions: "+(&NFPMPositionsRequest{}).Validate().Error())
	_, err = n.CallCreateAndInitializePoolIfNecessary(ctx, nil, &CreateAndInitializePoolIfNecessaryRequest{})
	require.EqualError(t, err, "uniswapv3: callCreateAndInitializePoolIfNecessary: "+(&CreateAndInitializePoolIfNecessaryRequest{}).Validate().Error())
	_, err = n.CallMulticall(ctx, nil, &NFPMMulticallRequest{})
	require.EqualError(t, err, "uniswapv3: callMulticall: "+(&NFPMMulticallRequest{}).Validate().Error())
	_, _, _, _, err = n.CallMint(ctx, nil, &NFPMMintRequest{})
	require.EqualError(t, err, "uniswapv3: callMint: "+(&NFPMMintRequest{}).Validate().Error())
	_, _, _, err = n.CallIncreaseLiquidity(ctx, nil, &IncreaseLiquidityRequest{})
	require.EqualError(t, err, "uniswapv3: callIncreaseLiquidity: "+(&IncreaseLiquidityRequest{}).Validate().Error())
	_, _, err = n.CallDecreaseLiquidity(ctx, nil, &DecreaseLiquidityRequest{})
	require.EqualError(t, err, "uniswapv3: callDecreaseLiquidity: "+(&DecreaseLiquidityRequest{}).Validate().Error())
	_, _, err = n.CallCollect(ctx, nil, &NFPMCollectRequest{})
	require.EqualError(t, err, "uniswapv3: callCollect: "+(&NFPMCollectRequest{}).Validate().Error())
	_, _, err = n.Permit(ctx, opts, &PermitRequest{})
	require.EqualError(t, err, "uniswapv3: permit: "+(&PermitRequest{}).Validate().Error())
	_, _, err = n.CreateAndInitializePoolIfNecessary(ctx, opts, &CreateAndInitializePoolIfNecessaryRequest{})
	require.EqualError(t, err, "uniswapv3: createAndInitializePoolIfNecessary: "+(&CreateAndInitializePoolIfNecessaryRequest{}).Validate().Error())
	_, _, err = n.UnwrapWETH9(ctx, opts, &NFPMUnwrapWETH9Request{})
	require.EqualError(t, err, "uniswapv3: unwrapWETH9: "+(&NFPMUnwrapWETH9Request{}).Validate().Error())
	_, _, err = n.SweepToken(ctx, opts, &NFPMSweepTokenRequest{})
	require.EqualError(t, err, "uniswapv3: sweepToken: "+(&NFPMSweepTokenRequest{}).Validate().Error())
	_, _, err = n.RefundETH(ctx, opts, &NFPMRefundETHRequest{})
	require.EqualError(t, err, "uniswapv3: refundETH: "+(&NFPMRefundETHRequest{}).Validate().Error())
	_, _, err = n.Multicall(ctx, opts, &NFPMMulticallRequest{})
	require.EqualError(t, err, "uniswapv3: multicall: "+(&NFPMMulticallRequest{}).Validate().Error())
	_, _, err = n.Mint(ctx, opts, &NFPMMintRequest{})
	require.EqualError(t, err, "uniswapv3: mint: "+(&NFPMMintRequest{}).Validate().Error())
	_, _, err = n.IncreaseLiquidity(ctx, opts, &IncreaseLiquidityRequest{})
	require.EqualError(t, err, "uniswapv3: increaseLiquidity: "+(&IncreaseLiquidityRequest{}).Validate().Error())
	_, _, err = n.DecreaseLiquidity(ctx, opts, &DecreaseLiquidityRequest{})
	require.EqualError(t, err, "uniswapv3: decreaseLiquidity: "+(&DecreaseLiquidityRequest{}).Validate().Error())
	_, _, err = n.Collect(ctx, opts, &NFPMCollectRequest{})
	require.EqualError(t, err, "uniswapv3: collect: "+(&NFPMCollectRequest{}).Validate().Error())
	_, _, err = n.Burn(ctx, opts, &NFPMBurnRequest{})
	require.EqualError(t, err, "uniswapv3: burn: "+(&NFPMBurnRequest{}).Validate().Error())
	require.ErrorIs(t, err, bind.ErrInvalidRequest)

	_, _, err = n.Mint(ctx, opts, NewNFPMMintRequest(nfpm, &MintParams{Token0: a}, nil))
	require.EqualError(t, err, "uniswapv3: mint: invalid request: params.token1 is nil")

	_, _, err = n.Collect(ctx, opts, NewNFPMCollectRequest(nfpm, NewCollectParams(big.NewInt(7), b, new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(0))))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)

	_, _, err = n.CreateAndInitializePoolIfNecessary(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewCreateAndInitializePoolIfNecessaryRequest(nfpm, a, b, big.NewInt(3000), big.NewInt(1)))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: createAndInitializePoolIfNecessary: invalid options: value must be nil")
	_, _, err = n.UnwrapWETH9(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewNFPMUnwrapWETH9Request(nfpm, big.NewInt(0), b))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: unwrapWETH9: invalid options: value must be nil")
	_, _, err = n.SweepToken(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewNFPMSweepTokenRequest(nfpm, a, big.NewInt(0), b))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: sweepToken: invalid options: value must be nil")
	_, _, err = n.RefundETH(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewNFPMRefundETHRequest(nfpm))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: refundETH: invalid options: value must be nil")
	_, _, err = n.Multicall(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewNFPMMulticallRequest(nfpm, [][]byte{{0x01}}, nil))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: multicall: invalid options: value must be nil")
	_, _, err = n.Mint(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewNFPMMintRequest(nfpm, NewMintParams(a, b, big.NewInt(3000), big.NewInt(-60), big.NewInt(60), big.NewInt(1000), big.NewInt(2000), big.NewInt(0), big.NewInt(0), b, big.NewInt(1_700_000_000)), nil))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: mint: invalid options: value must be nil")
	_, _, err = n.IncreaseLiquidity(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewIncreaseLiquidityRequest(nfpm, NewIncreaseLiquidityParams(big.NewInt(7), big.NewInt(1), big.NewInt(1), big.NewInt(0), big.NewInt(0), big.NewInt(1)), nil))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: increaseLiquidity: invalid options: value must be nil")
	_, err = n.CallCreateAndInitializePoolIfNecessary(ctx, bind.NewCallOptions().WithValue(big.NewInt(1)), NewCreateAndInitializePoolIfNecessaryRequest(nfpm, a, b, big.NewInt(3000), big.NewInt(1)))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: callCreateAndInitializePoolIfNecessary: invalid options: value must be nil")
	_, err = n.CallMulticall(ctx, bind.NewCallOptions().WithValue(big.NewInt(1)), NewNFPMMulticallRequest(nfpm, [][]byte{{0x01}}, nil))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: callMulticall: invalid options: value must be nil")
	_, _, _, _, err = n.CallMint(ctx, bind.NewCallOptions().WithValue(big.NewInt(1)), NewNFPMMintRequest(nfpm, NewMintParams(a, b, big.NewInt(3000), big.NewInt(-60), big.NewInt(60), big.NewInt(1000), big.NewInt(2000), big.NewInt(0), big.NewInt(0), b, big.NewInt(1_700_000_000)), nil))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: callMint: invalid options: value must be nil")
	_, _, _, err = n.CallIncreaseLiquidity(ctx, bind.NewCallOptions().WithValue(big.NewInt(1)), NewIncreaseLiquidityRequest(nfpm, NewIncreaseLiquidityParams(big.NewInt(7), big.NewInt(1), big.NewInt(1), big.NewInt(0), big.NewInt(0), big.NewInt(1)), nil))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: callIncreaseLiquidity: invalid options: value must be nil")

	_, _, err = n.Burn(ctx, nil, NewNFPMBurnRequest(nfpm, big.NewInt(7)))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: burn: bind: transact: invalid options: options is nil")
}

func TestNFPMWritesSignAndSend(t *testing.T) {
	nfpm := types.MustNewAddressFromHex("0xC36442b4a4522E871399CD717aBDD847Ab11FE88")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sender := core.PubkeyToAddress(key.PublicKey())
	sig, err := key.Sign(core.Keccak256([]byte("permit")))
	require.NoError(t, err)

	w0, err := EncodeNFTPermit(b, big.NewInt(7), big.NewInt(1_700_000_000), sig)
	require.NoError(t, err)
	w1, err := EncodeCreateAndInitializePoolIfNecessary(a, b, big.NewInt(3000), new(big.Int).Lsh(big.NewInt(1), 96))
	require.NoError(t, err)
	w2, err := EncodeUnwrapWETH9(big.NewInt(5), b)
	require.NoError(t, err)
	w3, err := EncodeSweepToken(a, big.NewInt(5), b)
	require.NoError(t, err)
	w4 := EncodeRefundETH()
	w5, err := EncodeMulticall([][]byte{{0xaa}, {0xbb}})
	require.NoError(t, err)
	w6, err := EncodeNfpmMint(a, b, big.NewInt(3000), big.NewInt(-60), big.NewInt(60), big.NewInt(1000), big.NewInt(2000), big.NewInt(0), big.NewInt(0), b, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	w8, err := EncodeIncreaseLiquidity(big.NewInt(7), big.NewInt(1000), big.NewInt(2000), big.NewInt(0), big.NewInt(0), big.NewInt(1_700_000_000))
	require.NoError(t, err)
	w9, err := EncodeDecreaseLiquidity(big.NewInt(7), big.NewInt(100), big.NewInt(0), big.NewInt(0), big.NewInt(1_700_000_000))
	require.NoError(t, err)
	w10, err := EncodeNfpmCollect(big.NewInt(7), b, big.NewInt(5), big.NewInt(6))
	require.NoError(t, err)
	w11, err := EncodeNfpmBurn(big.NewInt(7))
	require.NoError(t, err)

	for _, tc := range []struct {
		name      string
		calldata  []byte
		wantValue *big.Int
		send      func(n *NFPM, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error)
	}{
		{
			"permit",
			w0,
			nil,
			func(n *NFPM, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return n.Permit(t.Context(), opts, NewPermitRequest(nfpm, b, big.NewInt(7), big.NewInt(1_700_000_000), sig))
			},
		},
		{
			"createAndInitializePoolIfNecessary",
			w1,
			nil,
			func(n *NFPM, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return n.CreateAndInitializePoolIfNecessary(t.Context(), opts, NewCreateAndInitializePoolIfNecessaryRequest(nfpm, a, b, big.NewInt(3000), new(big.Int).Lsh(big.NewInt(1), 96)))
			},
		},
		{
			"unwrapWETH9",
			w2,
			nil,
			func(n *NFPM, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return n.UnwrapWETH9(t.Context(), opts, NewNFPMUnwrapWETH9Request(nfpm, big.NewInt(5), b))
			},
		},
		{
			"sweepToken",
			w3,
			nil,
			func(n *NFPM, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return n.SweepToken(t.Context(), opts, NewNFPMSweepTokenRequest(nfpm, a, big.NewInt(5), b))
			},
		},
		{
			"refundETH",
			w4,
			nil,
			func(n *NFPM, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return n.RefundETH(t.Context(), opts, NewNFPMRefundETHRequest(nfpm))
			},
		},
		{
			"multicall",
			w5,
			big.NewInt(42),
			func(n *NFPM, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return n.Multicall(t.Context(), opts, NewNFPMMulticallRequest(nfpm, [][]byte{{0xaa}, {0xbb}}, big.NewInt(42)))
			},
		},
		{
			"mint",
			w6,
			big.NewInt(1000),
			func(n *NFPM, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return n.Mint(t.Context(), opts, NewNFPMMintRequest(nfpm, NewMintParams(a, b, big.NewInt(3000), big.NewInt(-60), big.NewInt(60), big.NewInt(1000), big.NewInt(2000), big.NewInt(0), big.NewInt(0), b, big.NewInt(1_700_000_000)), big.NewInt(1000)))
			},
		},
		{
			"mint without native",
			w6,
			nil,
			func(n *NFPM, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return n.Mint(t.Context(), opts, NewNFPMMintRequest(nfpm, NewMintParams(a, b, big.NewInt(3000), big.NewInt(-60), big.NewInt(60), big.NewInt(1000), big.NewInt(2000), big.NewInt(0), big.NewInt(0), b, big.NewInt(1_700_000_000)), nil))
			},
		},
		{
			"increaseLiquidity",
			w8,
			big.NewInt(2000),
			func(n *NFPM, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return n.IncreaseLiquidity(t.Context(), opts, NewIncreaseLiquidityRequest(nfpm, NewIncreaseLiquidityParams(big.NewInt(7), big.NewInt(1000), big.NewInt(2000), big.NewInt(0), big.NewInt(0), big.NewInt(1_700_000_000)), big.NewInt(2000)))
			},
		},
		{
			"decreaseLiquidity",
			w9,
			nil,
			func(n *NFPM, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return n.DecreaseLiquidity(t.Context(), opts, NewDecreaseLiquidityRequest(nfpm, NewDecreaseLiquidityParams(big.NewInt(7), big.NewInt(100), big.NewInt(0), big.NewInt(0), big.NewInt(1_700_000_000))))
			},
		},
		{
			"collect",
			w10,
			nil,
			func(n *NFPM, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return n.Collect(t.Context(), opts, NewNFPMCollectRequest(nfpm, NewCollectParams(big.NewInt(7), b, big.NewInt(5), big.NewInt(6))))
			},
		},
		{
			"burn",
			w11,
			nil,
			func(n *NFPM, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return n.Burn(t.Context(), opts, NewNFPMBurnRequest(nfpm, big.NewInt(7)))
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
			signed, hash, err := tc.send(NewNFPM(c), opts)
			require.NoError(t, err)
			require.True(t, hash.Equal(core.Keccak256(raw)))
			require.Equal(t, uint64(7), signed.AddressNonce())
			require.Nil(t, opts.Value)

			decoded, err := core.DecodeTransaction(raw)
			require.NoError(t, err)
			dyn, ok := decoded.(*core.DynamicFeeTx)
			require.True(t, ok)
			require.True(t, dyn.To.Equal(nfpm))
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

func TestNFPMWritesMapEstimateRevertToSentinel(t *testing.T) {
	nfpm := types.MustNewAddressFromHex("0xC36442b4a4522E871399CD717aBDD847Ab11FE88")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)

	packed, err := abi.Pack(abi.NewTypes(abi.String), reasonNotCleared)
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

	signed, hash, err := NewNFPM(c).Burn(t.Context(), bind.NewTransactOptions(key, builder), NewNFPMBurnRequest(nfpm, big.NewInt(7)))
	require.ErrorIs(t, err, ErrNotCleared)
	require.NotErrorIs(t, err, ErrReverted)
	require.ErrorContains(t, err, "uniswapv3: burn: ")
	require.Nil(t, signed)
	require.Nil(t, hash)
}

func TestNFPMWritesReturnSignedTxOnBroadcastFailure(t *testing.T) {
	nfpm := types.MustNewAddressFromHex("0xC36442b4a4522E871399CD717aBDD847Ab11FE88")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(9)
	builder.Limit(3_000_000)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(errors.New("dial tcp: connection refused")).Once()

	signed, hash, err := NewNFPM(c).Mint(t.Context(), bind.NewTransactOptions(key, builder), NewNFPMMintRequest(nfpm, NewMintParams(a, b, big.NewInt(3000), big.NewInt(-60), big.NewInt(60), big.NewInt(1), big.NewInt(1), big.NewInt(0), big.NewInt(0), b, big.NewInt(1)), big.NewInt(777)))
	require.ErrorContains(t, err, "uniswapv3: mint: ")
	require.ErrorContains(t, err, "connection refused")
	require.NotErrorIs(t, err, ErrReverted)
	require.Nil(t, hash)
	require.NotNil(t, signed)
	require.Equal(t, uint64(9), signed.AddressNonce())
	dyn, ok := signed.(*core.DynamicFeeTx)
	require.True(t, ok)
	require.Zero(t, big.NewInt(777).Cmp(dyn.Value))
}

func TestNFPMCallsPropagateContext(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(t.Context(), marker{}, "nfpm")
	nfpm := types.MustNewAddressFromHex("0xC36442b4a4522E871399CD717aBDD847Ab11FE88")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	word, err := totalSupplyFn.EncodeReturn(big.NewInt(1))
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(got context.Context, e rpc.Element) error {
		require.Equal(t, "nfpm", got.Value(marker{}))
		if e.Method == rpc.MethodETHCall {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(word)
		}
		return nil
	}).Twice()

	_, err = NewNFPM(c).TotalSupply(ctx, nil, NewTotalSupplyRequest(nfpm))
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(1)
	builder.Limit(3_000_000)
	builder.Price(big.NewInt(2), big.NewInt(1))
	_, _, err = NewNFPM(c).Burn(ctx, bind.NewTransactOptions(key, builder), NewNFPMBurnRequest(nfpm, big.NewInt(7)))
	require.NoError(t, err)
}
