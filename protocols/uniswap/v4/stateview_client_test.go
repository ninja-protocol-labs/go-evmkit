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

func TestStateViewReadsSendCallAndDecode(t *testing.T) {
	sv := testPoolManager
	from := testSender
	opts := bind.NewCallOptions().WithFrom(from).WithBlock("0x1b4")
	poolID := PoolID{0xab, 0xcd}
	positionID := [32]byte{0x12, 0x34}
	salt := [32]byte{0x56}
	minTick := big.NewInt(-887272)
	maxUint128 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	minInt128 := new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 127))
	maxUint256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))

	poolManagerWord, err := abi.Pack(abi.NewTypes(abi.Address), testHooks)
	require.NoError(t, err)
	slot0Word, err := abi.Pack(abi.NewTypes(abi.Uint160, abi.Int24, abi.Uint24, abi.Uint24), big.NewInt(4295128739), minTick, big.NewInt(1000), big.NewInt(3000))
	require.NoError(t, err)
	liquidityWord, err := abi.Pack(abi.NewTypes(abi.Uint128), maxUint128)
	require.NoError(t, err)
	pairWord, err := abi.Pack(abi.NewTypes(abi.Uint256, abi.Uint256), maxUint256, big.NewInt(7))
	require.NoError(t, err)
	tickInfoWord, err := abi.Pack(abi.NewTypes(abi.Uint128, abi.Int128, abi.Uint256, abi.Uint256), maxUint128, minInt128, big.NewInt(11), big.NewInt(13))
	require.NoError(t, err)
	tickLiquidityWord, err := abi.Pack(abi.NewTypes(abi.Uint128, abi.Int128), big.NewInt(500), big.NewInt(-500))
	require.NoError(t, err)
	bitmapWord, err := abi.Pack(abi.NewTypes(abi.Uint256), new(big.Int).Lsh(big.NewInt(1), 255))
	require.NoError(t, err)
	positionWord, err := abi.Pack(abi.NewTypes(abi.Uint128, abi.Uint256, abi.Uint256), big.NewInt(1_000_000), big.NewInt(17), big.NewInt(19))
	require.NoError(t, err)
	positionLiquidityWord, err := abi.Pack(abi.NewTypes(abi.Uint128), big.NewInt(42))
	require.NoError(t, err)

	slot0Data, err := EncodeGetSlot0(poolID)
	require.NoError(t, err)
	liquidityData, err := EncodeGetLiquidity(poolID)
	require.NoError(t, err)
	globalsData, err := EncodeGetFeeGrowthGlobals(poolID)
	require.NoError(t, err)
	insideData, err := EncodeGetFeeGrowthInside(poolID, big.NewInt(-600), big.NewInt(600))
	require.NoError(t, err)
	tickInfoData, err := EncodeGetTickInfo(poolID, minTick)
	require.NoError(t, err)
	tickLiquidityData, err := EncodeGetTickLiquidity(poolID, big.NewInt(60))
	require.NoError(t, err)
	outsideData, err := EncodeGetTickFeeGrowthOutside(poolID, big.NewInt(-60))
	require.NoError(t, err)
	bitmapData, err := EncodeGetTickBitmap(poolID, -58)
	require.NoError(t, err)
	positionData, err := EncodeGetPositionInfo(poolID, from, big.NewInt(-60), big.NewInt(60), salt)
	require.NoError(t, err)
	positionByIDData, err := EncodeGetPositionInfoByID(poolID, positionID)
	require.NoError(t, err)
	positionLiquidityData, err := EncodeStateViewGetPositionLiquidity(poolID, positionID)
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		result   []byte
		call     func(v *StateView) (any, error)
		want     any
	}{
		{
			"poolManager",
			EncodeStateViewPoolManager(),
			poolManagerWord,
			func(v *StateView) (any, error) {
				return v.PoolManager(t.Context(), opts, NewPoolManagerRequest(sv))
			},
			testHooks,
		},
		{
			"getSlot0",
			slot0Data,
			slot0Word,
			func(v *StateView) (any, error) {
				return v.GetSlot0(t.Context(), opts, NewGetSlot0Request(sv, poolID))
			},
			NewSlot0(big.NewInt(4295128739), minTick, big.NewInt(1000), big.NewInt(3000)),
		},
		{
			"getLiquidity",
			liquidityData,
			liquidityWord,
			func(v *StateView) (any, error) {
				return v.GetLiquidity(t.Context(), opts, NewGetLiquidityRequest(sv, poolID))
			},
			maxUint128,
		},
		{
			"getFeeGrowthGlobals",
			globalsData,
			pairWord,
			func(v *StateView) (any, error) {
				return v.GetFeeGrowthGlobals(t.Context(), opts, NewGetFeeGrowthGlobalsRequest(sv, poolID))
			},
			NewFeeGrowthGlobals(maxUint256, big.NewInt(7)),
		},
		{
			"getFeeGrowthInside",
			insideData,
			pairWord,
			func(v *StateView) (any, error) {
				return v.GetFeeGrowthInside(t.Context(), opts, NewGetFeeGrowthInsideRequest(sv, poolID, big.NewInt(-600), big.NewInt(600)))
			},
			NewFeeGrowthInside(maxUint256, big.NewInt(7)),
		},
		{
			"getTickInfo",
			tickInfoData,
			tickInfoWord,
			func(v *StateView) (any, error) {
				return v.GetTickInfo(t.Context(), opts, NewGetTickInfoRequest(sv, poolID, minTick))
			},
			NewTickInfo(maxUint128, minInt128, big.NewInt(11), big.NewInt(13)),
		},
		{
			"getTickLiquidity",
			tickLiquidityData,
			tickLiquidityWord,
			func(v *StateView) (any, error) {
				return v.GetTickLiquidity(t.Context(), opts, NewGetTickLiquidityRequest(sv, poolID, big.NewInt(60)))
			},
			NewTickLiquidity(big.NewInt(500), big.NewInt(-500)),
		},
		{
			"getTickFeeGrowthOutside",
			outsideData,
			pairWord,
			func(v *StateView) (any, error) {
				return v.GetTickFeeGrowthOutside(t.Context(), opts, NewGetTickFeeGrowthOutsideRequest(sv, poolID, big.NewInt(-60)))
			},
			NewTickFeeGrowthOutside(maxUint256, big.NewInt(7)),
		},
		{
			"getTickBitmap",
			bitmapData,
			bitmapWord,
			func(v *StateView) (any, error) {
				return v.GetTickBitmap(t.Context(), opts, NewGetTickBitmapRequest(sv, poolID, -58))
			},
			new(big.Int).Lsh(big.NewInt(1), 255),
		},
		{
			"getPositionInfo",
			positionData,
			positionWord,
			func(v *StateView) (any, error) {
				return v.GetPositionInfo(t.Context(), opts, NewGetPositionInfoRequest(sv, poolID, from, big.NewInt(-60), big.NewInt(60), salt))
			},
			NewPositionState(big.NewInt(1_000_000), big.NewInt(17), big.NewInt(19)),
		},
		{
			"getPositionInfoByID",
			positionByIDData,
			positionWord,
			func(v *StateView) (any, error) {
				return v.GetPositionInfoByID(t.Context(), opts, NewGetPositionInfoByIDRequest(sv, poolID, positionID))
			},
			NewPositionState(big.NewInt(1_000_000), big.NewInt(17), big.NewInt(19)),
		},
		{
			"getPositionLiquidity",
			positionLiquidityData,
			positionLiquidityWord,
			func(v *StateView) (any, error) {
				return v.GetPositionLiquidity(t.Context(), opts, NewStateViewGetPositionLiquidityRequest(sv, poolID, positionID))
			},
			big.NewInt(42),
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
				require.Equal(t, sv.String(), call["to"])
				require.Equal(t, from.String(), call["from"])
				require.Equal(t, encoding.Hex.EncodePrefixed(tc.calldata), call["data"])
				require.NotContains(t, call, "value")
				require.Equal(t, "0x1b4", params[1])
				*e.Result.(*string) = encoding.Hex.EncodePrefixed(tc.result)
				return nil
			}).Once()

			got, err := tc.call(NewStateView(c))
			require.NoError(t, err)
			require.Equal(t, fmt.Sprint(tc.want), fmt.Sprint(got))
		})
	}
}

func TestStateViewReadsDefaultToLatestWithoutFrom(t *testing.T) {
	word, err := abi.Pack(abi.NewTypes(abi.Uint128), big.NewInt(9))
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

	liquidity, err := NewStateView(c).GetLiquidity(t.Context(), nil, NewGetLiquidityRequest(testPoolManager, PoolID{}))
	require.NoError(t, err)
	require.Zero(t, big.NewInt(9).Cmp(liquidity))
}

func TestStateViewReadsPassThroughRPCErrors(t *testing.T) {
	sv := testPoolManager

	for _, data := range []jsontext.Value{nil, jsontext.Value(`"0x"`), jsontext.Value(`"0xdeadbeef"`), jsontext.Value(`"0x08c379a0"`)} {
		t.Run(string(data), func(t *testing.T) {
			rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: data}
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).Return(rpcErr).Once()

			slot0, err := NewStateView(c).GetSlot0(t.Context(), nil, NewGetSlot0Request(sv, PoolID{}))
			require.Nil(t, slot0)
			require.ErrorContains(t, err, "uniswapv4: getSlot0: ")
			got, ok := errors.AsType[*rpc.ResponseError](err)
			require.True(t, ok)
			require.Same(t, rpcErr, got)
			require.NotErrorIs(t, err, ErrUnknown)
		})
	}

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(errors.New("dial tcp: connection refused")).Once()

	poolManager, err := NewStateView(c).PoolManager(t.Context(), nil, NewPoolManagerRequest(sv))
	require.Nil(t, poolManager)
	require.EqualError(t, err, "uniswapv4: stateViewPoolManager: dial tcp: connection refused")
}

func TestStateViewReadsRejectMalformedResult(t *testing.T) {
	sv := testPoolManager
	poolID := PoolID{0x01}
	tick := big.NewInt(60)

	for _, result := range []string{"0x", "0xzz", "0x01", "not hex"} {
		t.Run(result, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				*e.Result.(*string) = result
				return nil
			}).Times(12)
			v := NewStateView(c)
			ctx := t.Context()

			poolManager, err := v.PoolManager(ctx, nil, NewPoolManagerRequest(sv))
			require.Nil(t, poolManager)
			require.ErrorContains(t, err, "uniswapv4: stateViewPoolManager: ")

			slot0, err := v.GetSlot0(ctx, nil, NewGetSlot0Request(sv, poolID))
			require.Nil(t, slot0)
			require.ErrorContains(t, err, "uniswapv4: getSlot0: ")

			liquidity, err := v.GetLiquidity(ctx, nil, NewGetLiquidityRequest(sv, poolID))
			require.Nil(t, liquidity)
			require.ErrorContains(t, err, "uniswapv4: getLiquidity: ")

			globals, err := v.GetFeeGrowthGlobals(ctx, nil, NewGetFeeGrowthGlobalsRequest(sv, poolID))
			require.Nil(t, globals)
			require.ErrorContains(t, err, "uniswapv4: getFeeGrowthGlobals: ")

			inside, err := v.GetFeeGrowthInside(ctx, nil, NewGetFeeGrowthInsideRequest(sv, poolID, tick, tick))
			require.Nil(t, inside)
			require.ErrorContains(t, err, "uniswapv4: getFeeGrowthInside: ")

			tickInfo, err := v.GetTickInfo(ctx, nil, NewGetTickInfoRequest(sv, poolID, tick))
			require.Nil(t, tickInfo)
			require.ErrorContains(t, err, "uniswapv4: getTickInfo: ")

			tickLiquidity, err := v.GetTickLiquidity(ctx, nil, NewGetTickLiquidityRequest(sv, poolID, tick))
			require.Nil(t, tickLiquidity)
			require.ErrorContains(t, err, "uniswapv4: getTickLiquidity: ")

			outside, err := v.GetTickFeeGrowthOutside(ctx, nil, NewGetTickFeeGrowthOutsideRequest(sv, poolID, tick))
			require.Nil(t, outside)
			require.ErrorContains(t, err, "uniswapv4: getTickFeeGrowthOutside: ")

			bitmap, err := v.GetTickBitmap(ctx, nil, NewGetTickBitmapRequest(sv, poolID, 0))
			require.Nil(t, bitmap)
			require.ErrorContains(t, err, "uniswapv4: getTickBitmap: ")

			position, err := v.GetPositionInfo(ctx, nil, NewGetPositionInfoRequest(sv, poolID, testSender, tick, tick, [32]byte{}))
			require.Nil(t, position)
			require.ErrorContains(t, err, "uniswapv4: getPositionInfo: ")

			positionByID, err := v.GetPositionInfoByID(ctx, nil, NewGetPositionInfoByIDRequest(sv, poolID, [32]byte{}))
			require.Nil(t, positionByID)
			require.ErrorContains(t, err, "uniswapv4: getPositionInfoByID: ")

			positionLiquidity, err := v.GetPositionLiquidity(ctx, nil, NewStateViewGetPositionLiquidityRequest(sv, poolID, [32]byte{}))
			require.Nil(t, positionLiquidity)
			require.ErrorContains(t, err, "uniswapv4: stateViewGetPositionLiquidity: ")
		})
	}
}

func TestStateViewMethodsRejectInvalidRequestWithoutRPC(t *testing.T) {
	sv := testPoolManager
	v := NewStateView(mocks.NewClient(t))
	ctx := t.Context()
	outOfRange := big.NewInt(1 << 23)

	_, err := v.PoolManager(ctx, nil, nil)
	require.EqualError(t, err, "uniswapv4: stateViewPoolManager: "+(*PoolManagerRequest)(nil).Validate().Error())
	_, err = v.GetSlot0(ctx, nil, &GetSlot0Request{})
	require.EqualError(t, err, "uniswapv4: getSlot0: "+(&GetSlot0Request{}).Validate().Error())
	_, err = v.GetLiquidity(ctx, nil, NewGetLiquidityRequest(nil, PoolID{}))
	require.EqualError(t, err, "uniswapv4: getLiquidity: "+NewGetLiquidityRequest(nil, PoolID{}).Validate().Error())
	_, err = v.GetFeeGrowthGlobals(ctx, nil, nil)
	require.EqualError(t, err, "uniswapv4: getFeeGrowthGlobals: "+(*GetFeeGrowthGlobalsRequest)(nil).Validate().Error())
	_, err = v.GetFeeGrowthInside(ctx, nil, NewGetFeeGrowthInsideRequest(sv, PoolID{}, nil, big.NewInt(60)))
	require.EqualError(t, err, "uniswapv4: getFeeGrowthInside: "+NewGetFeeGrowthInsideRequest(sv, PoolID{}, nil, big.NewInt(60)).Validate().Error())
	_, err = v.GetTickInfo(ctx, nil, NewGetTickInfoRequest(sv, PoolID{}, nil))
	require.EqualError(t, err, "uniswapv4: getTickInfo: "+NewGetTickInfoRequest(sv, PoolID{}, nil).Validate().Error())
	_, err = v.GetTickLiquidity(ctx, nil, NewGetTickLiquidityRequest(nil, PoolID{}, big.NewInt(0)))
	require.EqualError(t, err, "uniswapv4: getTickLiquidity: "+NewGetTickLiquidityRequest(nil, PoolID{}, big.NewInt(0)).Validate().Error())
	_, err = v.GetTickFeeGrowthOutside(ctx, nil, &GetTickFeeGrowthOutsideRequest{Contract: sv})
	require.EqualError(t, err, "uniswapv4: getTickFeeGrowthOutside: "+(&GetTickFeeGrowthOutsideRequest{Contract: sv}).Validate().Error())
	_, err = v.GetTickBitmap(ctx, nil, nil)
	require.EqualError(t, err, "uniswapv4: getTickBitmap: "+(*GetTickBitmapRequest)(nil).Validate().Error())
	_, err = v.GetPositionInfo(ctx, nil, NewGetPositionInfoRequest(sv, PoolID{}, nil, big.NewInt(0), big.NewInt(0), [32]byte{}))
	require.EqualError(t, err, "uniswapv4: getPositionInfo: "+NewGetPositionInfoRequest(sv, PoolID{}, nil, big.NewInt(0), big.NewInt(0), [32]byte{}).Validate().Error())
	_, err = v.GetPositionInfoByID(ctx, nil, &GetPositionInfoByIDRequest{})
	require.EqualError(t, err, "uniswapv4: getPositionInfoByID: "+(&GetPositionInfoByIDRequest{}).Validate().Error())
	_, err = v.GetPositionLiquidity(ctx, nil, NewStateViewGetPositionLiquidityRequest(nil, PoolID{}, [32]byte{}))
	require.EqualError(t, err, "uniswapv4: stateViewGetPositionLiquidity: "+NewStateViewGetPositionLiquidityRequest(nil, PoolID{}, [32]byte{}).Validate().Error())
	require.ErrorIs(t, err, bind.ErrInvalidRequest)

	_, err = v.GetTickInfo(ctx, nil, NewGetTickInfoRequest(sv, PoolID{}, outOfRange))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)
	require.ErrorContains(t, err, "uniswapv4: getTickInfo: ")
	_, err = v.GetTickLiquidity(ctx, nil, NewGetTickLiquidityRequest(sv, PoolID{}, new(big.Int).Neg(new(big.Int).Add(outOfRange, big.NewInt(1)))))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)
	_, err = v.GetTickFeeGrowthOutside(ctx, nil, NewGetTickFeeGrowthOutsideRequest(sv, PoolID{}, outOfRange))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)
	_, err = v.GetFeeGrowthInside(ctx, nil, NewGetFeeGrowthInsideRequest(sv, PoolID{}, big.NewInt(0), outOfRange))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)
	_, err = v.GetPositionInfo(ctx, nil, NewGetPositionInfoRequest(sv, PoolID{}, testSender, outOfRange, big.NewInt(0), [32]byte{}))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)
}

func TestStateViewCallsPropagateContext(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(t.Context(), marker{}, "stateview")
	word, err := abi.Pack(abi.NewTypes(abi.Address), testHooks)
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(got context.Context, e rpc.Element) error {
		require.Equal(t, "stateview", got.Value(marker{}))
		*e.Result.(*string) = encoding.Hex.EncodePrefixed(word)
		return nil
	}).Once()

	poolManager, err := NewStateView(c).PoolManager(ctx, nil, NewPoolManagerRequest(testPoolManager))
	require.NoError(t, err)
	require.True(t, poolManager.Equal(testHooks))
}

func TestStateViewSatisfiesInterface(t *testing.T) {
	word, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(3))
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
		*e.Result.(*string) = encoding.Hex.EncodePrefixed(word)
		return nil
	}).Once()

	var v IStateView = NewStateView(c)
	bitmap, err := v.GetTickBitmap(t.Context(), nil, NewGetTickBitmapRequest(testPoolManager, PoolID{}, -1))
	require.NoError(t, err)
	require.Zero(t, big.NewInt(3).Cmp(bitmap))
}
