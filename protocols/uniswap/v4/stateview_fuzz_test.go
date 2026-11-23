package v4

import (
	"context"
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
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// go test -run '^$' -fuzz '^FuzzAllStateViewDecodersDoNotPanic$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzAllStateViewDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 64))
	f.Add(make([]byte, 96))
	f.Add(make([]byte, 128))
	f.Add(append(make([]byte, 31), 0xff))

	f.Fuzz(func(t *testing.T, data []byte) {
		require.NotPanics(t, func() {
			_, _ = DecodeGetSlot0(data)
			_, _ = DecodeGetTickInfo(data)
			_, _ = DecodeGetTickLiquidity(data)
			_, _ = DecodeGetTickFeeGrowthOutside(data)
			_, _ = DecodeGetFeeGrowthGlobals(data)
			_, _ = DecodeGetLiquidity(data)
			_, _ = DecodeGetTickBitmap(data)
			_, _ = DecodeGetPositionInfo(data)
			_, _ = DecodeStateViewGetPositionLiquidity(data)
			_, _ = DecodeGetFeeGrowthInside(data)
			_, _ = DecodeStateViewPoolManager(data)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzCalculatePositionKeyMatchesPackedEncoding$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzCalculatePositionKeyMatchesPackedEncoding(f *testing.F) {
	f.Add([]byte{0x66, 0x67}, int64(-600), int64(600), []byte{0x07})
	f.Add(make([]byte, 20), int64(-8388608), int64(8388607), make([]byte, 32))
	f.Add([]byte{0xff}, int64(-1), int64(0), []byte{0xff, 0xff})
	f.Add([]byte{0x01}, int64(8388608), int64(0), []byte{})
	f.Add([]byte{0x01}, int64(0), int64(-8388609), []byte{})

	f.Fuzz(func(t *testing.T, ownerRaw []byte, tickLower, tickUpper int64, saltRaw []byte) {
		owner := types.NewAddressFromBytes(ownerRaw)
		var salt [32]byte
		saltClamped := saltRaw
		if len(saltClamped) > 32 {
			saltClamped = saltClamped[len(saltClamped)-32:]
		}
		copy(salt[32-len(saltClamped):], saltClamped)

		key, err := CalculatePositionKey(owner, big.NewInt(tickLower), big.NewInt(tickUpper), salt)

		inRange := tickLower >= -(1<<23) && tickLower < 1<<23 && tickUpper >= -(1<<23) && tickUpper < 1<<23
		if !inRange {
			require.ErrorContains(t, err, "out of int24 range")
			return
		}
		require.NoError(t, err)

		twoPow24 := new(big.Int).Lsh(big.NewInt(1), 24)
		lowerWord := new(big.Int).Mod(big.NewInt(tickLower), twoPow24).FillBytes(make([]byte, 3))
		upperWord := new(big.Int).Mod(big.NewInt(tickUpper), twoPow24).FillBytes(make([]byte, 3))
		packed := append(append(append(append([]byte{}, owner.Bytes()...), lowerWord...), upperWord...), salt[:]...)
		require.Len(t, packed, 58)
		require.Equal(t, core.Keccak256(packed).Bytes(), key[:])
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeStateViewTickCallsRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeStateViewTickCallsRoundTrip(f *testing.F) {
	f.Add([]byte{0x57, 0x91}, int64(-20), int64(887220), int16(0))
	f.Add(make([]byte, 32), int64(-8388608), int64(8388607), int16(-32768))
	f.Add([]byte{0xff}, int64(8388608), int64(0), int16(32767))
	f.Add([]byte{}, int64(-8388609), int64(-1), int16(-1))

	f.Fuzz(func(t *testing.T, poolIDRaw []byte, tick, otherTick int64, word int16) {
		var poolID PoolID
		poolIDClamped := poolIDRaw
		if len(poolIDClamped) > 32 {
			poolIDClamped = poolIDClamped[len(poolIDClamped)-32:]
		}
		copy(poolID[32-len(poolIDClamped):], poolIDClamped)

		tickInRange := tick >= -(1<<23) && tick < 1<<23
		otherInRange := otherTick >= -(1<<23) && otherTick < 1<<23

		for _, c := range []struct {
			fn     *abi.Function
			encode func(PoolID, *big.Int) ([]byte, error)
		}{
			{getTickInfoFn, EncodeGetTickInfo},
			{getTickLiquidityFn, EncodeGetTickLiquidity},
			{getTickFeeGrowthOutsideFn, EncodeGetTickFeeGrowthOutside},
		} {
			data, err := c.encode(poolID, big.NewInt(tick))
			if !tickInRange {
				require.Error(t, err)
				continue
			}
			require.NoError(t, err)
			require.Len(t, data, 4+64)
			vals, err := c.fn.DecodeCall(data)
			require.NoError(t, err)
			require.Equal(t, poolID[:], vals[0])
			require.Zero(t, big.NewInt(tick).Cmp(vals[1].(*big.Int)))
		}

		inside, err := EncodeGetFeeGrowthInside(poolID, big.NewInt(tick), big.NewInt(otherTick))
		if !tickInRange || !otherInRange {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			vals, err := getFeeGrowthInsideFn.DecodeCall(inside)
			require.NoError(t, err)
			require.Equal(t, poolID[:], vals[0])
			require.Zero(t, big.NewInt(tick).Cmp(vals[1].(*big.Int)))
			require.Zero(t, big.NewInt(otherTick).Cmp(vals[2].(*big.Int)))
		}

		bitmap, err := EncodeGetTickBitmap(poolID, word)
		require.NoError(t, err)
		vals, err := getTickBitmapFn.DecodeCall(bitmap)
		require.NoError(t, err)
		require.Equal(t, poolID[:], vals[0])
		require.Equal(t, word, vals[1])
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeGetPositionInfoRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeGetPositionInfoRoundTrip(f *testing.F) {
	f.Add([]byte{0x57}, []byte{0x66}, int64(-600), int64(600), []byte{0x07})
	f.Add(make([]byte, 32), make([]byte, 20), int64(-8388608), int64(8388607), make([]byte, 32))
	f.Add([]byte{}, []byte{}, int64(0), int64(0), []byte{})

	f.Fuzz(func(t *testing.T, poolIDRaw, ownerRaw []byte, tickLower, tickUpper int64, saltRaw []byte) {
		var poolID PoolID
		poolIDClamped := poolIDRaw
		if len(poolIDClamped) > 32 {
			poolIDClamped = poolIDClamped[len(poolIDClamped)-32:]
		}
		copy(poolID[32-len(poolIDClamped):], poolIDClamped)

		var salt [32]byte
		saltClamped := saltRaw
		if len(saltClamped) > 32 {
			saltClamped = saltClamped[len(saltClamped)-32:]
		}
		copy(salt[32-len(saltClamped):], saltClamped)

		owner := types.NewAddressFromBytes(ownerRaw)

		byFields, err := EncodeGetPositionInfo(poolID, owner, big.NewInt(tickLower), big.NewInt(tickUpper), salt)
		positionID, keyErr := CalculatePositionKey(owner, big.NewInt(tickLower), big.NewInt(tickUpper), salt)
		require.Equal(t, err == nil, keyErr == nil, "encode and position key must agree on tick validity")
		if err != nil {
			return
		}

		vals, err := getPositionInfoFn.DecodeCall(byFields)
		require.NoError(t, err)
		require.Equal(t, poolID[:], vals[0])
		require.True(t, vals[1].(*types.Address).Equal(owner))
		require.Zero(t, big.NewInt(tickLower).Cmp(vals[2].(*big.Int)))
		require.Zero(t, big.NewInt(tickUpper).Cmp(vals[3].(*big.Int)))
		require.Equal(t, salt[:], vals[4])

		byID, err := EncodeGetPositionInfoByID(poolID, positionID)
		require.NoError(t, err)
		idVals, err := getPositionInfoByIDFn.DecodeCall(byID)
		require.NoError(t, err)
		require.Equal(t, poolID[:], idVals[0])
		require.Equal(t, positionID[:], idVals[1])

		liquidity, err := EncodeStateViewGetPositionLiquidity(poolID, positionID)
		require.NoError(t, err)
		require.Equal(t, byID[4:], liquidity[4:])
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeStateViewReturnsRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzDecodeStateViewReturnsRoundTrip(f *testing.F) {
	f.Add([]byte{0x01}, int64(-20), uint32(0), uint32(3000), []byte{0x01}, []byte{0x80}, true, []byte{0x01}, []byte{0x02})
	f.Add(make([]byte, 20), int64(-8388608), uint32(16777215), uint32(16777215), make([]byte, 16), make([]byte, 16), false, make([]byte, 32), make([]byte, 32))
	f.Add([]byte{}, int64(8388607), uint32(0), uint32(0), []byte{}, []byte{}, false, []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, sqrtPriceRaw []byte, tick int64, protocolFee, lpFee uint32, liquidityGrossRaw, liquidityNetRaw []byte, liquidityNetNegative bool, growth0Raw, growth1Raw []byte) {
		if tick < -(1<<23) || tick >= 1<<23 {
			tick %= 1 << 23
		}
		protocolFee &= 0xffffff
		lpFee &= 0xffffff

		sqrtPriceClamped := sqrtPriceRaw
		if len(sqrtPriceClamped) > 20 {
			sqrtPriceClamped = sqrtPriceClamped[len(sqrtPriceClamped)-20:]
		}
		sqrtPrice := new(big.Int).SetBytes(sqrtPriceClamped)

		liquidityGrossClamped := liquidityGrossRaw
		if len(liquidityGrossClamped) > 16 {
			liquidityGrossClamped = liquidityGrossClamped[len(liquidityGrossClamped)-16:]
		}
		liquidityGross := new(big.Int).SetBytes(liquidityGrossClamped)

		liquidityNetClamped := liquidityNetRaw
		if len(liquidityNetClamped) > 16 {
			liquidityNetClamped = liquidityNetClamped[len(liquidityNetClamped)-16:]
		}
		liquidityNet := new(big.Int).SetBytes(liquidityNetClamped)
		liquidityNet.Rsh(liquidityNet, 1)
		if liquidityNetNegative {
			liquidityNet.Neg(liquidityNet).Sub(liquidityNet, big.NewInt(1))
		}

		growth0Clamped := growth0Raw
		if len(growth0Clamped) > 32 {
			growth0Clamped = growth0Clamped[len(growth0Clamped)-32:]
		}
		growth0 := new(big.Int).SetBytes(growth0Clamped)

		growth1Clamped := growth1Raw
		if len(growth1Clamped) > 32 {
			growth1Clamped = growth1Clamped[len(growth1Clamped)-32:]
		}
		growth1 := new(big.Int).SetBytes(growth1Clamped)

		slot0Data, err := getSlot0Fn.EncodeReturn(sqrtPrice, big.NewInt(tick), big.NewInt(int64(protocolFee)), big.NewInt(int64(lpFee)))
		require.NoError(t, err)
		slot0, err := DecodeGetSlot0(slot0Data)
		require.NoError(t, err)
		require.Zero(t, sqrtPrice.Cmp(slot0.SqrtPriceX96))
		require.Zero(t, big.NewInt(tick).Cmp(slot0.Tick))
		require.Zero(t, big.NewInt(int64(protocolFee)).Cmp(slot0.ProtocolFee))
		require.Zero(t, big.NewInt(int64(lpFee)).Cmp(slot0.LPFee))

		tickInfoData, err := getTickInfoFn.EncodeReturn(liquidityGross, liquidityNet, growth0, growth1)
		require.NoError(t, err)
		info, err := DecodeGetTickInfo(tickInfoData)
		require.NoError(t, err)
		require.Zero(t, liquidityGross.Cmp(info.LiquidityGross))
		require.Zero(t, liquidityNet.Cmp(info.LiquidityNet))
		require.Zero(t, growth0.Cmp(info.FeeGrowthOutside0X128))
		require.Zero(t, growth1.Cmp(info.FeeGrowthOutside1X128))

		tickLiquidity, err := DecodeGetTickLiquidity(tickInfoData[:64])
		require.NoError(t, err)
		require.Zero(t, liquidityGross.Cmp(tickLiquidity.LiquidityGross))
		require.Zero(t, liquidityNet.Cmp(tickLiquidity.LiquidityNet))

		outside, err := DecodeGetTickFeeGrowthOutside(tickInfoData[64:])
		require.NoError(t, err)
		require.Zero(t, growth0.Cmp(outside.FeeGrowthOutside0X128))
		require.Zero(t, growth1.Cmp(outside.FeeGrowthOutside1X128))

		positionData, err := getPositionInfoFn.EncodeReturn(liquidityGross, growth0, growth1)
		require.NoError(t, err)
		position, err := DecodeGetPositionInfo(positionData)
		require.NoError(t, err)
		require.Zero(t, liquidityGross.Cmp(position.Liquidity))
		require.Zero(t, growth0.Cmp(position.FeeGrowthInside0LastX128))
		require.Zero(t, growth1.Cmp(position.FeeGrowthInside1LastX128))

		globals, err := DecodeGetFeeGrowthGlobals(tickInfoData[64:])
		require.NoError(t, err)
		require.Zero(t, growth0.Cmp(globals.FeeGrowthGlobal0))
		inside, err := DecodeGetFeeGrowthInside(tickInfoData[64:])
		require.NoError(t, err)
		require.Zero(t, growth1.Cmp(inside.FeeGrowthInside1X128))
	})
}

// go test -run '^$' -fuzz '^FuzzStateViewTickInfoThroughClient$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzStateViewTickInfoThroughClient(f *testing.F) {
	f.Add(make([]byte, 32), int32(0), uint64(0))
	f.Add([]byte{0xab, 0xcd}, int32(-887272), uint64(1))
	f.Add([]byte{0xff}, int32(1<<23-1), uint64(1<<40))

	word, err := abi.Pack(abi.NewTypes(abi.Uint128, abi.Int128, abi.Uint256, abi.Uint256), big.NewInt(1), big.NewInt(-1), big.NewInt(2), big.NewInt(3))
	require.NoError(f, err)

	f.Fuzz(func(t *testing.T, poolIDRaw []byte, tickRaw int32, block uint64) {
		var poolID PoolID
		copy(poolID[:], poolIDRaw)
		// fold tickRaw into int24's [-2^23, 2^23) range
		tick := big.NewInt(int64(tickRaw<<8) >> 8)
		blockTag := fmt.Sprintf("0x%x", block)

		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			params := e.Params.([]any)
			require.Equal(t, blockTag, params[1])
			call := params[0].(map[string]any)
			require.Equal(t, testPoolManager.String(), call["to"])
			calldata, err := encoding.Hex.Decode(call["data"].(string))
			require.NoError(t, err)
			vals, err := getTickInfoFn.DecodeCall(calldata)
			require.NoError(t, err)
			require.Equal(t, poolID[:], vals[0].([]byte))
			require.Zero(t, tick.Cmp(vals[1].(*big.Int)))
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(word)
			return nil
		}).Once()

		info, err := NewStateView(c).GetTickInfo(t.Context(), bind.NewCallOptions().WithBlock(blockTag), NewGetTickInfoRequest(testPoolManager, poolID, tick))
		require.NoError(t, err)
		require.Zero(t, big.NewInt(1).Cmp(info.LiquidityGross))
		require.Zero(t, big.NewInt(-1).Cmp(info.LiquidityNet))
	})
}

// go test -run '^$' -fuzz '^FuzzStateViewReadsSurviveArbitraryResult$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzStateViewReadsSurviveArbitraryResult(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 96))
	f.Add(make([]byte, 128))
	f.Add(append(make([]byte, 31), 0xff))

	tick := big.NewInt(60)

	f.Fuzz(func(t *testing.T, result []byte) {
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(result)
			return nil
		}).Times(12)
		v := NewStateView(c)
		ctx := t.Context()

		poolManager, err := v.PoolManager(ctx, nil, NewPoolManagerRequest(testPoolManager))
		require.True(t, (err == nil) != (poolManager == nil))
		slot0, err := v.GetSlot0(ctx, nil, NewGetSlot0Request(testPoolManager, PoolID{}))
		require.True(t, (err == nil) != (slot0 == nil))
		liquidity, err := v.GetLiquidity(ctx, nil, NewGetLiquidityRequest(testPoolManager, PoolID{}))
		require.True(t, (err == nil) != (liquidity == nil))
		globals, err := v.GetFeeGrowthGlobals(ctx, nil, NewGetFeeGrowthGlobalsRequest(testPoolManager, PoolID{}))
		require.True(t, (err == nil) != (globals == nil))
		inside, err := v.GetFeeGrowthInside(ctx, nil, NewGetFeeGrowthInsideRequest(testPoolManager, PoolID{}, tick, tick))
		require.True(t, (err == nil) != (inside == nil))
		tickInfo, err := v.GetTickInfo(ctx, nil, NewGetTickInfoRequest(testPoolManager, PoolID{}, tick))
		require.True(t, (err == nil) != (tickInfo == nil))
		tickLiquidity, err := v.GetTickLiquidity(ctx, nil, NewGetTickLiquidityRequest(testPoolManager, PoolID{}, tick))
		require.True(t, (err == nil) != (tickLiquidity == nil))
		outside, err := v.GetTickFeeGrowthOutside(ctx, nil, NewGetTickFeeGrowthOutsideRequest(testPoolManager, PoolID{}, tick))
		require.True(t, (err == nil) != (outside == nil))
		bitmap, err := v.GetTickBitmap(ctx, nil, NewGetTickBitmapRequest(testPoolManager, PoolID{}, 0))
		require.True(t, (err == nil) != (bitmap == nil))
		position, err := v.GetPositionInfo(ctx, nil, NewGetPositionInfoRequest(testPoolManager, PoolID{}, testSender, tick, tick, [32]byte{}))
		require.True(t, (err == nil) != (position == nil))
		positionByID, err := v.GetPositionInfoByID(ctx, nil, NewGetPositionInfoByIDRequest(testPoolManager, PoolID{}, [32]byte{}))
		require.True(t, (err == nil) != (positionByID == nil))
		positionLiquidity, err := v.GetPositionLiquidity(ctx, nil, NewStateViewGetPositionLiquidityRequest(testPoolManager, PoolID{}, [32]byte{}))
		require.True(t, (err == nil) != (positionLiquidity == nil))
		if err == nil {
			require.LessOrEqual(t, positionLiquidity.BitLen(), 128)
		}
	})
}
