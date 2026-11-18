package v4

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

func TestStateViewFunctionSelectors(t *testing.T) {
	cases := []struct {
		name string
		fn   *abi.Function
		sig  string
		sel  string
	}{
		{"getSlot0", getSlot0Fn, "getSlot0(bytes32)", "0xc815641c"},
		{"getTickInfo", getTickInfoFn, "getTickInfo(bytes32,int24)", "0x7c40f1fe"},
		{"getTickLiquidity", getTickLiquidityFn, "getTickLiquidity(bytes32,int24)", "0xcaedab54"},
		{"getTickFeeGrowthOutside", getTickFeeGrowthOutsideFn, "getTickFeeGrowthOutside(bytes32,int24)", "0x8a2bb9e6"},
		{"getFeeGrowthGlobals", getFeeGrowthGlobalsFn, "getFeeGrowthGlobals(bytes32)", "0x9ec538c8"},
		{"getLiquidity", getLiquidityFn, "getLiquidity(bytes32)", "0xfa6793d5"},
		{"getTickBitmap", getTickBitmapFn, "getTickBitmap(bytes32,int16)", "0x1c7ccb4c"},
		{"getPositionInfo", getPositionInfoFn, "getPositionInfo(bytes32,address,int24,int24,bytes32)", "0xdacf1d2f"},
		{"getPositionInfoByID", getPositionInfoByIDFn, "getPositionInfo(bytes32,bytes32)", "0x97fd7b42"},
		{"getPositionLiquidity", stateViewGetPositionLiquidityFn, "getPositionLiquidity(bytes32,bytes32)", "0xf0928f29"},
		{"getFeeGrowthInside", getFeeGrowthInsideFn, "getFeeGrowthInside(bytes32,int24,int24)", "0x53e9c1fb"},
		{"poolManager", stateViewPoolManagerFn, "poolManager()", "0xdc4c90d3"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.sig, c.fn.Signature())
			sel := c.fn.Selector()
			require.Equal(t, c.sel, encoding.Hex.EncodePrefixed(sel[:]))
		})
	}
}

func TestStateViewSelectorsAreUnique(t *testing.T) {
	seen := make(map[abi.Selector]string)
	for _, fn := range []*abi.Function{
		getSlot0Fn, getTickInfoFn, getTickLiquidityFn, getTickFeeGrowthOutsideFn, getFeeGrowthGlobalsFn,
		getLiquidityFn, getTickBitmapFn, getPositionInfoFn, getPositionInfoByIDFn,
		stateViewGetPositionLiquidityFn, getFeeGrowthInsideFn, stateViewPoolManagerFn,
	} {
		sel := fn.Selector()
		prev, dup := seen[sel]
		require.False(t, dup, "selector %x shared by %s and %s", sel, prev, fn.Signature())
		seen[sel] = fn.Signature()
	}
}

func TestPoolKeyPoolIDMatchesGoldenForStateViewTests(t *testing.T) {
	key := NewPoolKey(
		types.MustNewAddressFromHex("0x8382d6168621b72748fB9F39409dE13510b7755D"),
		types.MustNewAddressFromHex("0xC2B5B3dcD56bF9c456E0910CCB2beb96A7dA35fc"),
		big.NewInt(3000), big.NewInt(60),
		types.MustNewAddressFromHex("0x0000000000000000000000000000000000000000"),
	)
	id, err := key.PoolID()
	require.NoError(t, err)
	require.Equal(t, "0x5791fb7021bcb40d4a002179ef633ca0195af031a95b68cad7d68a239f8e09f2", encoding.Hex.EncodePrefixed(id[:]))
}

func TestEncodeStateViewCallsMatchGoldenCalldata(t *testing.T) {
	poolIDBytes, err := encoding.Hex.Decode("0x5791fb7021bcb40d4a002179ef633ca0195af031a95b68cad7d68a239f8e09f2")
	require.NoError(t, err)
	var poolID PoolID
	copy(poolID[:], poolIDBytes)

	positionIDBytes, err := encoding.Hex.Decode("0x61a237c56f3a308261c2d9b5974952f42fdc16df84d6cd9c48d1f21459eadfa7")
	require.NoError(t, err)
	var positionID [32]byte
	copy(positionID[:], positionIDBytes)

	var salt [32]byte
	salt[31] = 7

	slot0, err := EncodeGetSlot0(poolID)
	require.NoError(t, err)
	require.Equal(t, "0xc815641c5791fb7021bcb40d4a002179ef633ca0195af031a95b68cad7d68a239f8e09f2", encoding.Hex.EncodePrefixed(slot0))

	tickInfo, err := EncodeGetTickInfo(poolID, big.NewInt(-887220))
	require.NoError(t, err)
	require.Equal(t, "0x7c40f1fe5791fb7021bcb40d4a002179ef633ca0195af031a95b68cad7d68a239f8e09f2fffffffffffffffffffffffffffffffffffffffffffffffffffffffffff2764c", encoding.Hex.EncodePrefixed(tickInfo))

	bitmap, err := EncodeGetTickBitmap(poolID, -32768)
	require.NoError(t, err)
	require.Equal(t, "0x1c7ccb4c5791fb7021bcb40d4a002179ef633ca0195af031a95b68cad7d68a239f8e09f2ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8000", encoding.Hex.EncodePrefixed(bitmap))

	positionInfo, err := EncodeGetPositionInfo(poolID, testSender, big.NewInt(-600), big.NewInt(600), salt)
	require.NoError(t, err)
	require.Equal(t, "0xdacf1d2f5791fb7021bcb40d4a002179ef633ca0195af031a95b68cad7d68a239f8e09f2"+
		"0000000000000000000000006667c8dc9fbfec411e7c1ee2b24de960149f930f"+
		"fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffda8"+
		"0000000000000000000000000000000000000000000000000000000000000258"+
		"0000000000000000000000000000000000000000000000000000000000000007", encoding.Hex.EncodePrefixed(positionInfo))

	positionInfoByID, err := EncodeGetPositionInfoByID(poolID, positionID)
	require.NoError(t, err)
	require.Equal(t, "0x97fd7b425791fb7021bcb40d4a002179ef633ca0195af031a95b68cad7d68a239f8e09f261a237c56f3a308261c2d9b5974952f42fdc16df84d6cd9c48d1f21459eadfa7", encoding.Hex.EncodePrefixed(positionInfoByID))

	positionLiquidity, err := EncodeStateViewGetPositionLiquidity(poolID, positionID)
	require.NoError(t, err)
	require.Equal(t, "0xf0928f295791fb7021bcb40d4a002179ef633ca0195af031a95b68cad7d68a239f8e09f261a237c56f3a308261c2d9b5974952f42fdc16df84d6cd9c48d1f21459eadfa7", encoding.Hex.EncodePrefixed(positionLiquidity))

	inside, err := EncodeGetFeeGrowthInside(poolID, big.NewInt(-60), big.NewInt(8388607))
	require.NoError(t, err)
	require.Equal(t, "0x53e9c1fb5791fb7021bcb40d4a002179ef633ca0195af031a95b68cad7d68a239f8e09f2ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc400000000000000000000000000000000000000000000000000000000007fffff", encoding.Hex.EncodePrefixed(inside))

	require.Equal(t, "0xdc4c90d3", encoding.Hex.EncodePrefixed(EncodeStateViewPoolManager()))
}

func TestEncodeStateViewCallsRoundTripViaDecodeCall(t *testing.T) {
	var poolID PoolID
	for i := range poolID {
		poolID[i] = byte(i + 1)
	}
	var positionID [32]byte
	for i := range positionID {
		positionID[i] = byte(0xff - i)
	}

	cases := []struct {
		name   string
		fn     *abi.Function
		encode func() ([]byte, error)
		want   []any
	}{
		{"getSlot0", getSlot0Fn, func() ([]byte, error) { return EncodeGetSlot0(poolID) }, []any{poolID[:]}},
		{"getTickInfo", getTickInfoFn, func() ([]byte, error) { return EncodeGetTickInfo(poolID, big.NewInt(-8388608)) }, []any{poolID[:], big.NewInt(-8388608)}},
		{"getTickLiquidity", getTickLiquidityFn, func() ([]byte, error) { return EncodeGetTickLiquidity(poolID, big.NewInt(8388607)) }, []any{poolID[:], big.NewInt(8388607)}},
		{"getTickFeeGrowthOutside", getTickFeeGrowthOutsideFn, func() ([]byte, error) { return EncodeGetTickFeeGrowthOutside(poolID, big.NewInt(0)) }, []any{poolID[:], big.NewInt(0)}},
		{"getFeeGrowthGlobals", getFeeGrowthGlobalsFn, func() ([]byte, error) { return EncodeGetFeeGrowthGlobals(poolID) }, []any{poolID[:]}},
		{"getLiquidity", getLiquidityFn, func() ([]byte, error) { return EncodeGetLiquidity(poolID) }, []any{poolID[:]}},
		{"getTickBitmap", getTickBitmapFn, func() ([]byte, error) { return EncodeGetTickBitmap(poolID, 32767) }, []any{poolID[:], int16(32767)}},
		{"getPositionInfoByID", getPositionInfoByIDFn, func() ([]byte, error) { return EncodeGetPositionInfoByID(poolID, positionID) }, []any{poolID[:], positionID[:]}},
		{"getPositionLiquidity", stateViewGetPositionLiquidityFn, func() ([]byte, error) { return EncodeStateViewGetPositionLiquidity(poolID, positionID) }, []any{poolID[:], positionID[:]}},
		{"getFeeGrowthInside", getFeeGrowthInsideFn, func() ([]byte, error) {
			return EncodeGetFeeGrowthInside(poolID, big.NewInt(-887272), big.NewInt(887272))
		}, []any{poolID[:], big.NewInt(-887272), big.NewInt(887272)}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			data, err := c.encode()
			require.NoError(t, err)
			require.Equal(t, 4+32*len(c.want), len(data))
			vals, err := c.fn.DecodeCall(data)
			require.NoError(t, err)
			require.Len(t, vals, len(c.want))
			for i, want := range c.want {
				switch w := want.(type) {
				case *big.Int:
					require.Zero(t, w.Cmp(vals[i].(*big.Int)), "arg %d", i)
				default:
					require.Equal(t, w, vals[i], "arg %d", i)
				}
			}
		})
	}
}

func TestEncodeGetPositionInfoRoundTripViaDecodeCall(t *testing.T) {
	var poolID PoolID
	poolID[0] = 0xaa
	var salt [32]byte
	salt[0] = 0x80
	salt[31] = 0x01

	data, err := EncodeGetPositionInfo(poolID, testHooks, big.NewInt(-887272), big.NewInt(887272), salt)
	require.NoError(t, err)
	vals, err := getPositionInfoFn.DecodeCall(data)
	require.NoError(t, err)
	require.Len(t, vals, 5)
	require.Equal(t, poolID[:], vals[0])
	require.True(t, vals[1].(*types.Address).Equal(testHooks))
	require.Zero(t, big.NewInt(-887272).Cmp(vals[2].(*big.Int)))
	require.Zero(t, big.NewInt(887272).Cmp(vals[3].(*big.Int)))
	require.Equal(t, salt[:], vals[4])
}

func TestGetPositionInfoOverloadsDifferOnlyInSelectorAndArgs(t *testing.T) {
	var poolID PoolID
	poolID[31] = 1
	var salt [32]byte
	salt[31] = 9

	positionID, err := CalculatePositionKey(testSender, big.NewInt(-60), big.NewInt(60), salt)
	require.NoError(t, err)

	byFields, err := EncodeGetPositionInfo(poolID, testSender, big.NewInt(-60), big.NewInt(60), salt)
	require.NoError(t, err)
	byID, err := EncodeGetPositionInfoByID(poolID, positionID)
	require.NoError(t, err)
	require.NotEqual(t, byFields[:4], byID[:4])
	require.Len(t, byFields, 4+32*5)
	require.Len(t, byID, 4+32*2)
	require.Equal(t, byFields[4:36], byID[4:36])
	require.Equal(t, positionID[:], byID[36:68])
}

func TestEncodeStateViewRejectsOutOfRangeTicks(t *testing.T) {
	var poolID PoolID
	for _, tick := range []*big.Int{big.NewInt(8388608), big.NewInt(-8388609), new(big.Int).Lsh(big.NewInt(1), 200), nil} {
		_, err := EncodeGetTickInfo(poolID, tick)
		require.ErrorContains(t, err, "uniswapv4: encodeGetTickInfo", "tick %v", tick)
		_, err = EncodeGetTickLiquidity(poolID, tick)
		require.ErrorContains(t, err, "uniswapv4: encodeGetTickLiquidity", "tick %v", tick)
		_, err = EncodeGetTickFeeGrowthOutside(poolID, tick)
		require.ErrorContains(t, err, "uniswapv4: encodeGetTickFeeGrowthOutside", "tick %v", tick)
		_, err = EncodeGetFeeGrowthInside(poolID, tick, big.NewInt(0))
		require.ErrorContains(t, err, "uniswapv4: encodeGetFeeGrowthInside", "tick %v", tick)
		_, err = EncodeGetFeeGrowthInside(poolID, big.NewInt(0), tick)
		require.ErrorContains(t, err, "uniswapv4: encodeGetFeeGrowthInside", "tick %v", tick)
		_, err = EncodeGetPositionInfo(poolID, testSender, tick, big.NewInt(0), [32]byte{})
		require.ErrorContains(t, err, "uniswapv4: encodeGetPositionInfo", "tick %v", tick)
		_, err = EncodeGetPositionInfo(poolID, testSender, big.NewInt(0), tick, [32]byte{})
		require.ErrorContains(t, err, "uniswapv4: encodeGetPositionInfo", "tick %v", tick)
	}
}

func TestEncodeGetPositionInfoRejectsNilOwner(t *testing.T) {
	_, err := EncodeGetPositionInfo(PoolID{}, nil, big.NewInt(-60), big.NewInt(60), [32]byte{})
	require.ErrorContains(t, err, "uniswapv4: encodeGetPositionInfo")
}

func TestDecodeGetSlot0LiveGolden(t *testing.T) {
	data, err := encoding.Hex.Decode("0x" +
		"0000000000000000000000000000000000000000ffbe8771b01e02aef00146b1" +
		"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffec" +
		"0000000000000000000000000000000000000000000000000000000000000000" +
		"0000000000000000000000000000000000000000000000000000000000000bb8")
	require.NoError(t, err)

	slot0, err := DecodeGetSlot0(data)
	require.NoError(t, err)
	wantSqrtPrice, ok := new(big.Int).SetString("79149013026817623549647341233", 10)
	require.True(t, ok)
	require.Zero(t, wantSqrtPrice.Cmp(slot0.SqrtPriceX96))
	require.Zero(t, big.NewInt(-20).Cmp(slot0.Tick))
	require.Zero(t, slot0.ProtocolFee.Sign())
	require.Zero(t, big.NewInt(3000).Cmp(slot0.LPFee))
}

func TestDecodeGetFeeGrowthGlobalsLiveGolden(t *testing.T) {
	data, err := encoding.Hex.Decode("0x" +
		"00000000000000000000000000000000000064dd7c56c1fd253f9dcba8642882" +
		"0000000000000000000000000000000000003254e6e221c8a7a41e57d9dba908")
	require.NoError(t, err)

	globals, err := DecodeGetFeeGrowthGlobals(data)
	require.NoError(t, err)
	require.Equal(t, "0x64dd7c56c1fd253f9dcba8642882", "0x"+globals.FeeGrowthGlobal0.Text(16))
	require.Equal(t, "0x3254e6e221c8a7a41e57d9dba908", "0x"+globals.FeeGrowthGlobal1.Text(16))
}

func TestDecodeGetTickInfoExtremes(t *testing.T) {
	data, err := encoding.Hex.Decode("0x" +
		"00000000000000000000000000000000ffffffffffffffffffffffffffffffff" +
		"ffffffffffffffffffffffffffffffff80000000000000000000000000000000" +
		"0000000000000000000000000000000000000000000000000000000000000001" +
		"0000000000000000000000000000000000000000000000000000000000000000")
	require.NoError(t, err)

	info, err := DecodeGetTickInfo(data)
	require.NoError(t, err)
	maxUint128 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	minInt128 := new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 127))
	require.Zero(t, maxUint128.Cmp(info.LiquidityGross))
	require.Zero(t, minInt128.Cmp(info.LiquidityNet))
	require.Zero(t, big.NewInt(1).Cmp(info.FeeGrowthOutside0X128))
	require.Zero(t, info.FeeGrowthOutside1X128.Sign())

	liquidity, err := DecodeGetTickLiquidity(data[:64])
	require.NoError(t, err)
	require.Zero(t, maxUint128.Cmp(liquidity.LiquidityGross))
	require.Zero(t, minInt128.Cmp(liquidity.LiquidityNet))
}

func TestDecodeStateViewReturnsRoundTrip(t *testing.T) {
	maxUint256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	maxUint160 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 160), big.NewInt(1))
	maxUint128 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))

	slot0Data, err := getSlot0Fn.EncodeReturn(maxUint160, big.NewInt(-8388608), big.NewInt(16777215), big.NewInt(1000000))
	require.NoError(t, err)
	slot0, err := DecodeGetSlot0(slot0Data)
	require.NoError(t, err)
	require.Zero(t, maxUint160.Cmp(slot0.SqrtPriceX96))
	require.Zero(t, big.NewInt(-8388608).Cmp(slot0.Tick))
	require.Zero(t, big.NewInt(16777215).Cmp(slot0.ProtocolFee))
	require.Zero(t, big.NewInt(1000000).Cmp(slot0.LPFee))

	outsideData, err := getTickFeeGrowthOutsideFn.EncodeReturn(maxUint256, big.NewInt(7))
	require.NoError(t, err)
	outside, err := DecodeGetTickFeeGrowthOutside(outsideData)
	require.NoError(t, err)
	require.Zero(t, maxUint256.Cmp(outside.FeeGrowthOutside0X128))
	require.Zero(t, big.NewInt(7).Cmp(outside.FeeGrowthOutside1X128))

	liquidityData, err := getLiquidityFn.EncodeReturn(maxUint128)
	require.NoError(t, err)
	liquidity, err := DecodeGetLiquidity(liquidityData)
	require.NoError(t, err)
	require.Zero(t, maxUint128.Cmp(liquidity))

	bitmapData, err := getTickBitmapFn.EncodeReturn(maxUint256)
	require.NoError(t, err)
	bitmap, err := DecodeGetTickBitmap(bitmapData)
	require.NoError(t, err)
	require.Zero(t, maxUint256.Cmp(bitmap))

	positionData, err := getPositionInfoFn.EncodeReturn(maxUint128, maxUint256, big.NewInt(0))
	require.NoError(t, err)
	position, err := DecodeGetPositionInfo(positionData)
	require.NoError(t, err)
	require.Zero(t, maxUint128.Cmp(position.Liquidity))
	require.Zero(t, maxUint256.Cmp(position.FeeGrowthInside0LastX128))
	require.Zero(t, position.FeeGrowthInside1LastX128.Sign())

	positionLiquidity, err := DecodeStateViewGetPositionLiquidity(liquidityData)
	require.NoError(t, err)
	require.Zero(t, maxUint128.Cmp(positionLiquidity))

	insideData, err := getFeeGrowthInsideFn.EncodeReturn(big.NewInt(1), maxUint256)
	require.NoError(t, err)
	inside, err := DecodeGetFeeGrowthInside(insideData)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1).Cmp(inside.FeeGrowthInside0X128))
	require.Zero(t, maxUint256.Cmp(inside.FeeGrowthInside1X128))

	globalsData, err := getFeeGrowthGlobalsFn.EncodeReturn(maxUint256, maxUint256)
	require.NoError(t, err)
	globals, err := DecodeGetFeeGrowthGlobals(globalsData)
	require.NoError(t, err)
	require.Zero(t, maxUint256.Cmp(globals.FeeGrowthGlobal0))
	require.Zero(t, maxUint256.Cmp(globals.FeeGrowthGlobal1))

	pmData, err := encoding.Hex.Decode("0x000000000000000000000000fb3e0c6f74eb1a21cc1da29aec80d2dfe6c9a317")
	require.NoError(t, err)
	pm, err := DecodeStateViewPoolManager(pmData)
	require.NoError(t, err)
	require.True(t, pm.Equal(types.MustNewAddressFromHex("0xFB3e0C6F74eB1a21CC1Da29aeC80D2Dfe6C9a317")))
}

func TestDecodeStateViewRejectsNonCanonicalWords(t *testing.T) {
	cases := []struct {
		name   string
		decode func([]byte) error
		data   string
	}{
		{"slot0SqrtPriceAboveUint160", func(d []byte) error { _, err := DecodeGetSlot0(d); return err }, "0x" +
			"0000000000000000000000010000000000000000000000000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000000000"},
		{"slot0TickAboveInt24", func(d []byte) error { _, err := DecodeGetSlot0(d); return err }, "0x" +
			"0000000000000000000000000000000000000000000000000000000000000001" +
			"0000000000000000000000000000000000000000000000000000000000800000" +
			"0000000000000000000000000000000000000000000000000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000000000"},
		{"slot0LPFeeAboveUint24", func(d []byte) error { _, err := DecodeGetSlot0(d); return err }, "0x" +
			"0000000000000000000000000000000000000000000000000000000000000001" +
			"0000000000000000000000000000000000000000000000000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000001000000"},
		{"tickInfoLiquidityGrossAboveUint128", func(d []byte) error { _, err := DecodeGetTickInfo(d); return err }, "0x" +
			"0000000000000000000000000000000100000000000000000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000000000"},
		{"tickInfoLiquidityNetBelowInt128", func(d []byte) error { _, err := DecodeGetTickInfo(d); return err }, "0x" +
			"0000000000000000000000000000000000000000000000000000000000000000" +
			"ffffffffffffffffffffffffffffffff7fffffffffffffffffffffffffffffff" +
			"0000000000000000000000000000000000000000000000000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000000000"},
		{"liquidityAboveUint128", func(d []byte) error { _, err := DecodeGetLiquidity(d); return err }, "0x" +
			"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
		{"positionLiquidityAboveUint128", func(d []byte) error { _, err := DecodeGetPositionInfo(d); return err }, "0x" +
			"0000000000000000000000000000000100000000000000000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000000000"},
		{"poolManagerDirtyHighBytes", func(d []byte) error { _, err := DecodeStateViewPoolManager(d); return err }, "0x" +
			"010000000000000000000000fb3e0c6f74eb1a21cc1da29aec80d2dfe6c9a317"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			data, err := encoding.Hex.Decode(c.data)
			require.NoError(t, err)
			require.Error(t, c.decode(data))
		})
	}
}

func TestDecodeStateViewRejectsShortData(t *testing.T) {
	for _, n := range []int{0, 1, 31} {
		data := make([]byte, n)
		_, err := DecodeGetLiquidity(data)
		require.ErrorContains(t, err, "uniswapv4: decodeGetLiquidity")
		_, err = DecodeGetTickBitmap(data)
		require.ErrorContains(t, err, "uniswapv4: decodeGetTickBitmap")
		_, err = DecodeStateViewGetPositionLiquidity(data)
		require.ErrorContains(t, err, "uniswapv4: decodeStateViewGetPositionLiquidity")
		_, err = DecodeStateViewPoolManager(data)
		require.ErrorContains(t, err, "uniswapv4: decodeStateViewPoolManager")
	}
	for _, n := range []int{0, 32, 63} {
		data := make([]byte, n)
		_, err := DecodeGetTickLiquidity(data)
		require.ErrorContains(t, err, "uniswapv4: decodeGetTickLiquidity")
		_, err = DecodeGetTickFeeGrowthOutside(data)
		require.ErrorContains(t, err, "uniswapv4: decodeGetTickFeeGrowthOutside")
		_, err = DecodeGetFeeGrowthGlobals(data)
		require.ErrorContains(t, err, "uniswapv4: decodeGetFeeGrowthGlobals")
		_, err = DecodeGetFeeGrowthInside(data)
		require.ErrorContains(t, err, "uniswapv4: decodeGetFeeGrowthInside")
	}
	for _, n := range []int{0, 64, 95} {
		_, err := DecodeGetPositionInfo(make([]byte, n))
		require.ErrorContains(t, err, "uniswapv4: decodeGetPositionInfo")
	}
	for _, n := range []int{0, 96, 127} {
		data := make([]byte, n)
		_, err := DecodeGetSlot0(data)
		require.ErrorContains(t, err, "uniswapv4: decodeGetSlot0")
		_, err = DecodeGetTickInfo(data)
		require.ErrorContains(t, err, "uniswapv4: decodeGetTickInfo")
	}
}

func TestDecodeStateViewZeroPoolIsAllZero(t *testing.T) {
	slot0, err := DecodeGetSlot0(make([]byte, 128))
	require.NoError(t, err)
	require.Zero(t, slot0.SqrtPriceX96.Sign())
	require.Zero(t, slot0.Tick.Sign())
	require.Zero(t, slot0.ProtocolFee.Sign())
	require.Zero(t, slot0.LPFee.Sign())

	info, err := DecodeGetTickInfo(make([]byte, 128))
	require.NoError(t, err)
	require.Zero(t, info.LiquidityGross.Sign())
	require.Zero(t, info.LiquidityNet.Sign())
}

func TestEncodeStateViewPoolManagerReturnsFreshSlice(t *testing.T) {
	a := EncodeStateViewPoolManager()
	a[0] = 0x00
	require.Equal(t, "0xdc4c90d3", encoding.Hex.EncodePrefixed(EncodeStateViewPoolManager()))
}
