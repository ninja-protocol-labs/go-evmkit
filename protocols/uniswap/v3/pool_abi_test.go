package v3

import (
	"encoding/json/jsontext"
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

func TestEncodeFactory(t *testing.T) {
	require.Equal(t, []byte{0xc4, 0x5a, 0x01, 0x55}, EncodeFactory())
}

func TestDecodeFactory(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testTokenA)
	require.NoError(t, err)

	got, err := DecodeFactory(data)
	require.NoError(t, err)
	require.True(t, got.Equal(testTokenA))
}

func TestDecodeFactoryRejectsGarbage(t *testing.T) {
	_, err := DecodeFactory([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeToken0(t *testing.T) {
	require.Equal(t, []byte{0x0d, 0xfe, 0x16, 0x81}, EncodeToken0())
}

func TestDecodeToken0(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testTokenA)
	require.NoError(t, err)

	got, err := DecodeToken0(data)
	require.NoError(t, err)
	require.True(t, got.Equal(testTokenA))
}

func TestDecodeToken0RejectsGarbage(t *testing.T) {
	_, err := DecodeToken0([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeToken1(t *testing.T) {
	require.Equal(t, []byte{0xd2, 0x12, 0x20, 0xa7}, EncodeToken1())
}

func TestDecodeToken1(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testTokenB)
	require.NoError(t, err)

	got, err := DecodeToken1(data)
	require.NoError(t, err)
	require.True(t, got.Equal(testTokenB))
}

func TestDecodeToken1RejectsGarbage(t *testing.T) {
	_, err := DecodeToken1([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodePoolFee(t *testing.T) {
	require.Equal(t, []byte{0xdd, 0xca, 0x3f, 0x43}, EncodePoolFee())
}

func TestDecodePoolFee(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint24}, big.NewInt(3000))
	require.NoError(t, err)

	got, err := DecodePoolFee(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(3000).Cmp(got))
}

func TestDecodePoolFeeRejectsGarbage(t *testing.T) {
	_, err := DecodePoolFee([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodePoolTickSpacing(t *testing.T) {
	require.Equal(t, []byte{0xd0, 0xc9, 0x3a, 0x7c}, EncodePoolTickSpacing())
}

func TestDecodePoolTickSpacing(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Int24}, big.NewInt(60))
	require.NoError(t, err)

	got, err := DecodePoolTickSpacing(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(60).Cmp(got))
}

func TestDecodePoolTickSpacingRejectsGarbage(t *testing.T) {
	_, err := DecodePoolTickSpacing([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeMaxLiquidityPerTick(t *testing.T) {
	require.Equal(t, []byte{0x70, 0xcf, 0x75, 0x4a}, EncodeMaxLiquidityPerTick())
}

func TestDecodeMaxLiquidityPerTick(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint128}, big.NewInt(123456789))
	require.NoError(t, err)

	got, err := DecodeMaxLiquidityPerTick(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(123456789).Cmp(got))
}

func TestDecodeMaxLiquidityPerTickRejectsGarbage(t *testing.T) {
	_, err := DecodeMaxLiquidityPerTick([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeSlot0(t *testing.T) {
	require.Equal(t, []byte{0x38, 0x50, 0xc7, 0xbd}, EncodeSlot0())
}

func TestDecodeSlot0(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint160, abi.Int24, abi.Uint16, abi.Uint16, abi.Uint16, abi.Uint8, abi.Bool},
		big.NewInt(1234567890), big.NewInt(-887272), uint16(5), uint16(10), uint16(20), uint8(6), true)
	require.NoError(t, err)

	got, err := DecodeSlot0(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1234567890).Cmp(got.SqrtPriceX96))
	require.Zero(t, big.NewInt(-887272).Cmp(got.Tick))
	require.Zero(t, big.NewInt(5).Cmp(got.ObservationIndex))
	require.Zero(t, big.NewInt(10).Cmp(got.ObservationCardinality))
	require.Zero(t, big.NewInt(20).Cmp(got.ObservationCardinalityNext))
	require.Zero(t, big.NewInt(6).Cmp(got.FeeProtocol))
	require.True(t, got.Unlocked)
}

func TestDecodeSlot0RejectsGarbage(t *testing.T) {
	_, err := DecodeSlot0([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestDecodeSlot0ObservationCardinalityMaxUint16(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint160, abi.Int24, abi.Uint16, abi.Uint16, abi.Uint16, abi.Uint8, abi.Bool},
		big.NewInt(1), big.NewInt(0), uint16(65535), uint16(65535), uint16(65535), uint8(255), false)
	require.NoError(t, err)

	got, err := DecodeSlot0(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(65535).Cmp(got.ObservationIndex))
	require.Zero(t, big.NewInt(65535).Cmp(got.ObservationCardinality))
	require.Zero(t, big.NewInt(65535).Cmp(got.ObservationCardinalityNext))
	require.Zero(t, big.NewInt(255).Cmp(got.FeeProtocol))
	require.False(t, got.Unlocked)
}

func TestEncodeFeeGrowthGlobal0X128(t *testing.T) {
	require.Equal(t, []byte{0xf3, 0x05, 0x83, 0x99}, EncodeFeeGrowthGlobal0X128())
}

func TestDecodeFeeGrowthGlobal0X128(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(999))
	require.NoError(t, err)

	got, err := DecodeFeeGrowthGlobal0X128(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(999).Cmp(got))
}

func TestDecodeFeeGrowthGlobal0X128RejectsGarbage(t *testing.T) {
	_, err := DecodeFeeGrowthGlobal0X128([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeFeeGrowthGlobal1X128(t *testing.T) {
	require.Equal(t, []byte{0x46, 0x14, 0x13, 0x19}, EncodeFeeGrowthGlobal1X128())
}

func TestDecodeFeeGrowthGlobal1X128(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(888))
	require.NoError(t, err)

	got, err := DecodeFeeGrowthGlobal1X128(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(888).Cmp(got))
}

func TestDecodeFeeGrowthGlobal1X128RejectsGarbage(t *testing.T) {
	_, err := DecodeFeeGrowthGlobal1X128([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeProtocolFees(t *testing.T) {
	require.Equal(t, []byte{0x1a, 0xd8, 0xb0, 0x3b}, EncodeProtocolFees())
}

func TestDecodeProtocolFees(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint128, abi.Uint128}, big.NewInt(100), big.NewInt(200))
	require.NoError(t, err)

	got, err := DecodeProtocolFees(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(100).Cmp(got.Token0))
	require.Zero(t, big.NewInt(200).Cmp(got.Token1))
}

func TestDecodeProtocolFeesRejectsGarbage(t *testing.T) {
	_, err := DecodeProtocolFees([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodePoolLiquidity(t *testing.T) {
	require.Equal(t, []byte{0x1a, 0x68, 0x65, 0x02}, EncodePoolLiquidity())
}

func TestDecodePoolLiquidity(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint128}, big.NewInt(42))
	require.NoError(t, err)

	got, err := DecodePoolLiquidity(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(42).Cmp(got))
}

func TestDecodePoolLiquidityRejectsGarbage(t *testing.T) {
	_, err := DecodePoolLiquidity([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeTicks(t *testing.T) {
	data, err := EncodeTicks(big.NewInt(-100))
	require.NoError(t, err)
	require.Equal(t, []byte{0xf3, 0x0d, 0xba, 0x93}, data[:4])

	vals, err := ticksFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(-100).Cmp(vals[0].(*big.Int)))
}

func TestDecodeTicks(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint128, abi.Int128, abi.Uint256, abi.Uint256, abi.Int56, abi.Uint160, abi.Uint32, abi.Bool},
		big.NewInt(1), big.NewInt(-2), big.NewInt(3), big.NewInt(4), big.NewInt(-5), big.NewInt(6), uint32(7), true)
	require.NoError(t, err)

	got, err := DecodeTicks(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1).Cmp(got.LiquidityGross))
	require.Zero(t, big.NewInt(-2).Cmp(got.LiquidityNet))
	require.Zero(t, big.NewInt(3).Cmp(got.FeeGrowthOutside0X128))
	require.Zero(t, big.NewInt(4).Cmp(got.FeeGrowthOutside1X128))
	require.Zero(t, big.NewInt(-5).Cmp(got.TickCumulativeOutside))
	require.Zero(t, big.NewInt(6).Cmp(got.SecondsPerLiquidityOutsideX128))
	require.Zero(t, big.NewInt(7).Cmp(got.SecondsOutside))
	require.True(t, got.Initialized)
}

func TestDecodeTicksRejectsGarbage(t *testing.T) {
	_, err := DecodeTicks([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestDecodeTicksSecondsOutsideMaxUint32(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint128, abi.Int128, abi.Uint256, abi.Uint256, abi.Int56, abi.Uint160, abi.Uint32, abi.Bool},
		big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), uint32(4294967295), false)
	require.NoError(t, err)

	got, err := DecodeTicks(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(4294967295).Cmp(got.SecondsOutside))
}

func TestEncodeTickBitmap(t *testing.T) {
	data, err := EncodeTickBitmap(-5)
	require.NoError(t, err)
	require.Equal(t, []byte{0x53, 0x39, 0xc2, 0x96}, data[:4])

	vals, err := tickBitmapFn.DecodeCall(data)
	require.NoError(t, err)
	require.Equal(t, int16(-5), vals[0].(int16))
}

func TestEncodeTickBitmapMinInt16(t *testing.T) {
	data, err := EncodeTickBitmap(-32768)
	require.NoError(t, err)

	vals, err := tickBitmapFn.DecodeCall(data)
	require.NoError(t, err)
	require.Equal(t, int16(-32768), vals[0].(int16))
}

func TestDecodeTickBitmap(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(123))
	require.NoError(t, err)

	got, err := DecodeTickBitmap(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(123).Cmp(got))
}

func TestDecodeTickBitmapRejectsGarbage(t *testing.T) {
	_, err := DecodeTickBitmap([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodePositions(t *testing.T) {
	var key [32]byte
	key[31] = 0x07
	data, err := EncodePositions(key)
	require.NoError(t, err)
	require.Equal(t, []byte{0x51, 0x4e, 0xa4, 0xbf}, data[:4])

	vals, err := positionsFn.DecodeCall(data)
	require.NoError(t, err)
	require.Equal(t, key[:], vals[0].([]byte))
}

func TestDecodePositions(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint128, abi.Uint256, abi.Uint256, abi.Uint128, abi.Uint128},
		big.NewInt(1), big.NewInt(2), big.NewInt(3), big.NewInt(4), big.NewInt(5))
	require.NoError(t, err)

	got, err := DecodePositions(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1).Cmp(got.Liquidity))
	require.Zero(t, big.NewInt(2).Cmp(got.FeeGrowthInside0LastX128))
	require.Zero(t, big.NewInt(3).Cmp(got.FeeGrowthInside1LastX128))
	require.Zero(t, big.NewInt(4).Cmp(got.TokensOwed0))
	require.Zero(t, big.NewInt(5).Cmp(got.TokensOwed1))
}

func TestDecodePositionsRejectsGarbage(t *testing.T) {
	_, err := DecodePositions([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodePoolObservations(t *testing.T) {
	data, err := EncodePoolObservations(big.NewInt(9))
	require.NoError(t, err)
	require.Equal(t, []byte{0x25, 0x2c, 0x09, 0xd7}, data[:4])

	vals, err := poolObservationsFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(9).Cmp(vals[0].(*big.Int)))
}

func TestDecodePoolObservations(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint32, abi.Int56, abi.Uint160, abi.Bool}, uint32(111), big.NewInt(222), big.NewInt(333), true)
	require.NoError(t, err)

	got, err := DecodePoolObservations(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(111).Cmp(got.BlockTimestamp))
	require.Zero(t, big.NewInt(222).Cmp(got.TickCumulative))
	require.Zero(t, big.NewInt(333).Cmp(got.SecondsPerLiquidityCumulativeX128))
	require.True(t, got.Initialized)
}

func TestDecodePoolObservationsRejectsGarbage(t *testing.T) {
	_, err := DecodePoolObservations([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestDecodePoolObservationsBlockTimestampMaxUint32(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint32, abi.Int56, abi.Uint160, abi.Bool}, uint32(4294967295), big.NewInt(0), big.NewInt(0), false)
	require.NoError(t, err)

	got, err := DecodePoolObservations(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(4294967295).Cmp(got.BlockTimestamp))
}

func TestEncodeObserve(t *testing.T) {
	data, err := EncodeObserve([]uint32{3600, 0})
	require.NoError(t, err)
	require.Equal(t, []byte{0x88, 0x3b, 0xdb, 0xfd}, data[:4])

	vals, err := observeFn.DecodeCall(data)
	require.NoError(t, err)
	secondsAgos := vals[0].([]any)
	require.Len(t, secondsAgos, 2)
}

func TestEncodeObserveEmptySlice(t *testing.T) {
	data, err := EncodeObserve([]uint32{})
	require.NoError(t, err)

	vals, err := observeFn.DecodeCall(data)
	require.NoError(t, err)
	require.Empty(t, vals[0].([]any))
}

func TestDecodeObserve(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Slice(abi.Int56), abi.Slice(abi.Uint160)},
		[]*big.Int{big.NewInt(1), big.NewInt(2)}, []*big.Int{big.NewInt(3), big.NewInt(4)})
	require.NoError(t, err)

	got, err := DecodeObserve(data)
	require.NoError(t, err)
	require.Len(t, got.TickCumulatives, 2)
	require.Zero(t, big.NewInt(1).Cmp(got.TickCumulatives[0]))
	require.Zero(t, big.NewInt(2).Cmp(got.TickCumulatives[1]))
	require.Len(t, got.SecondsPerLiquidityCumulativeX128s, 2)
	require.Zero(t, big.NewInt(3).Cmp(got.SecondsPerLiquidityCumulativeX128s[0]))
	require.Zero(t, big.NewInt(4).Cmp(got.SecondsPerLiquidityCumulativeX128s[1]))
}

func TestDecodeObserveRejectsGarbage(t *testing.T) {
	_, err := DecodeObserve([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestDecodeObserveEmptySlices(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Slice(abi.Int56), abi.Slice(abi.Uint160)}, []*big.Int{}, []*big.Int{})
	require.NoError(t, err)

	got, err := DecodeObserve(data)
	require.NoError(t, err)
	require.Empty(t, got.TickCumulatives)
	require.Empty(t, got.SecondsPerLiquidityCumulativeX128s)
}

func TestEncodeSnapshotCumulativesInside(t *testing.T) {
	data, err := EncodeSnapshotCumulativesInside(big.NewInt(-10), big.NewInt(10))
	require.NoError(t, err)
	require.Equal(t, []byte{0xa3, 0x88, 0x07, 0xf2}, data[:4])

	vals, err := snapshotCumulativesInsideFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(-10).Cmp(vals[0].(*big.Int)))
	require.Zero(t, big.NewInt(10).Cmp(vals[1].(*big.Int)))
}

func TestDecodeSnapshotCumulativesInside(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Int56, abi.Uint160, abi.Uint32}, big.NewInt(-1), big.NewInt(2), uint32(3))
	require.NoError(t, err)

	got, err := DecodeSnapshotCumulativesInside(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(-1).Cmp(got.TickCumulativeInside))
	require.Zero(t, big.NewInt(2).Cmp(got.SecondsPerLiquidityInsideX128))
	require.Zero(t, big.NewInt(3).Cmp(got.SecondsInside))
}

func TestDecodeSnapshotCumulativesInsideRejectsGarbage(t *testing.T) {
	_, err := DecodeSnapshotCumulativesInside([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestDecodeSnapshotCumulativesInsideSecondsInsideMaxUint32(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Int56, abi.Uint160, abi.Uint32}, big.NewInt(0), big.NewInt(0), uint32(4294967295))
	require.NoError(t, err)

	got, err := DecodeSnapshotCumulativesInside(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(4294967295).Cmp(got.SecondsInside))
}

func TestEncodeInitialize(t *testing.T) {
	sqrtPriceX96 := new(big.Int).Lsh(big.NewInt(1), 96)
	data, err := EncodeInitialize(sqrtPriceX96)
	require.NoError(t, err)
	require.Equal(t, []byte{0xf6, 0x37, 0x73, 0x1d}, data[:4])

	vals, err := initializeFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, sqrtPriceX96.Cmp(vals[0].(*big.Int)))
}

func TestEncodeMint(t *testing.T) {
	data, err := EncodeMint(testTokenA, big.NewInt(-100), big.NewInt(100), big.NewInt(500), []byte{0xde, 0xad})
	require.NoError(t, err)
	require.Equal(t, []byte{0x3c, 0x8a, 0x7d, 0x8d}, data[:4])

	vals, err := mintFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testTokenA))
	require.Zero(t, big.NewInt(-100).Cmp(vals[1].(*big.Int)))
	require.Zero(t, big.NewInt(100).Cmp(vals[2].(*big.Int)))
	require.Zero(t, big.NewInt(500).Cmp(vals[3].(*big.Int)))
	require.Equal(t, []byte{0xde, 0xad}, vals[4].([]byte))
}

func TestDecodeMint(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256, abi.Uint256}, big.NewInt(1), big.NewInt(2))
	require.NoError(t, err)

	amount0, amount1, err := DecodeMint(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1).Cmp(amount0))
	require.Zero(t, big.NewInt(2).Cmp(amount1))
}

func TestDecodeMintRejectsGarbage(t *testing.T) {
	_, _, err := DecodeMint([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeCollect(t *testing.T) {
	data, err := EncodeCollect(testTokenA, big.NewInt(-100), big.NewInt(100), big.NewInt(1), big.NewInt(2))
	require.NoError(t, err)
	require.Equal(t, []byte{0x4f, 0x1e, 0xb3, 0xd8}, data[:4])

	vals, err := collectFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testTokenA))
	require.Zero(t, big.NewInt(1).Cmp(vals[3].(*big.Int)))
	require.Zero(t, big.NewInt(2).Cmp(vals[4].(*big.Int)))
}

func TestDecodeCollect(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint128, abi.Uint128}, big.NewInt(1), big.NewInt(2))
	require.NoError(t, err)

	amount0, amount1, err := DecodeCollect(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1).Cmp(amount0))
	require.Zero(t, big.NewInt(2).Cmp(amount1))
}

func TestDecodeCollectRejectsGarbage(t *testing.T) {
	_, _, err := DecodeCollect([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeBurn(t *testing.T) {
	data, err := EncodeBurn(big.NewInt(-100), big.NewInt(100), big.NewInt(500))
	require.NoError(t, err)
	require.Equal(t, []byte{0xa3, 0x41, 0x23, 0xa7}, data[:4])

	vals, err := burnFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(-100).Cmp(vals[0].(*big.Int)))
	require.Zero(t, big.NewInt(100).Cmp(vals[1].(*big.Int)))
	require.Zero(t, big.NewInt(500).Cmp(vals[2].(*big.Int)))
}

func TestDecodeBurn(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256, abi.Uint256}, big.NewInt(1), big.NewInt(2))
	require.NoError(t, err)

	amount0, amount1, err := DecodeBurn(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1).Cmp(amount0))
	require.Zero(t, big.NewInt(2).Cmp(amount1))
}

func TestDecodeBurnRejectsGarbage(t *testing.T) {
	_, _, err := DecodeBurn([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeSwap(t *testing.T) {
	data, err := EncodeSwap(testTokenA, true, big.NewInt(1000), big.NewInt(4295128740), []byte{0xbe, 0xef})
	require.NoError(t, err)
	require.Equal(t, []byte{0x12, 0x8a, 0xcb, 0x08}, data[:4])

	vals, err := swapFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testTokenA))
	require.True(t, vals[1].(bool))
	require.Zero(t, big.NewInt(1000).Cmp(vals[2].(*big.Int)))
}

func TestEncodeSwapNegativeAmountSpecified(t *testing.T) {
	maxSqrtRatio, ok := new(big.Int).SetString("1461446703485210103287273052203988822378723970342", 10)
	require.True(t, ok)
	data, err := EncodeSwap(testTokenA, false, big.NewInt(-1000), maxSqrtRatio, []byte{})
	require.NoError(t, err)

	vals, err := swapFn.DecodeCall(data)
	require.NoError(t, err)
	require.False(t, vals[1].(bool))
	require.Zero(t, big.NewInt(-1000).Cmp(vals[2].(*big.Int)))
}

func TestDecodeSwap(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Int256, abi.Int256}, big.NewInt(-1), big.NewInt(2))
	require.NoError(t, err)

	amount0, amount1, err := DecodeSwap(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(-1).Cmp(amount0))
	require.Zero(t, big.NewInt(2).Cmp(amount1))
}

func TestDecodeSwapRejectsGarbage(t *testing.T) {
	_, _, err := DecodeSwap([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeFlash(t *testing.T) {
	data, err := EncodeFlash(testTokenA, big.NewInt(100), big.NewInt(200), []byte{0x01})
	require.NoError(t, err)
	require.Equal(t, []byte{0x49, 0x0e, 0x6c, 0xbc}, data[:4])

	vals, err := flashFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testTokenA))
	require.Zero(t, big.NewInt(100).Cmp(vals[1].(*big.Int)))
	require.Zero(t, big.NewInt(200).Cmp(vals[2].(*big.Int)))
}

func TestEncodeIncreaseObservationCardinalityNext(t *testing.T) {
	data, err := EncodeIncreaseObservationCardinalityNext(100)
	require.NoError(t, err)
	require.Equal(t, []byte{0x32, 0x14, 0x8f, 0x67}, data[:4])

	vals, err := increaseObservationCardinalityNextFn.DecodeCall(data)
	require.NoError(t, err)
	require.Equal(t, uint16(100), vals[0].(uint16))
}

func TestEncodeIncreaseObservationCardinalityNextMaxUint16(t *testing.T) {
	data, err := EncodeIncreaseObservationCardinalityNext(65535)
	require.NoError(t, err)

	vals, err := increaseObservationCardinalityNextFn.DecodeCall(data)
	require.NoError(t, err)
	require.Equal(t, uint16(65535), vals[0].(uint16))
}

func TestEncodeSetFeeProtocol(t *testing.T) {
	data, err := EncodeSetFeeProtocol(4, 10)
	require.NoError(t, err)
	require.Equal(t, []byte{0x82, 0x06, 0xa4, 0xd1}, data[:4])

	vals, err := setFeeProtocolFn.DecodeCall(data)
	require.NoError(t, err)
	require.Equal(t, uint8(4), vals[0].(uint8))
	require.Equal(t, uint8(10), vals[1].(uint8))
}

func TestEncodeSetFeeProtocolMaxUint8(t *testing.T) {
	data, err := EncodeSetFeeProtocol(255, 255)
	require.NoError(t, err)

	vals, err := setFeeProtocolFn.DecodeCall(data)
	require.NoError(t, err)
	require.Equal(t, uint8(255), vals[0].(uint8))
	require.Equal(t, uint8(255), vals[1].(uint8))
}

func TestEncodeCollectProtocol(t *testing.T) {
	data, err := EncodeCollectProtocol(testTokenA, big.NewInt(1), big.NewInt(2))
	require.NoError(t, err)
	require.Equal(t, []byte{0x85, 0xb6, 0x67, 0x29}, data[:4])

	vals, err := collectProtocolFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testTokenA))
	require.Zero(t, big.NewInt(1).Cmp(vals[1].(*big.Int)))
	require.Zero(t, big.NewInt(2).Cmp(vals[2].(*big.Int)))
}

func TestDecodeCollectProtocol(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint128, abi.Uint128}, big.NewInt(1), big.NewInt(2))
	require.NoError(t, err)

	amount0, amount1, err := DecodeCollectProtocol(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1).Cmp(amount0))
	require.Zero(t, big.NewInt(2).Cmp(amount1))
}

func TestDecodeCollectProtocolRejectsGarbage(t *testing.T) {
	_, _, err := DecodeCollectProtocol([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestPoolRevertReasons(t *testing.T) {
	reasons := []string{
		reasonLocked,
		reasonTickLowerGreaterThanOrEqualUpper,
		reasonTickLowerTooLow,
		reasonTickUpperTooHigh,
		reasonAlreadyInitialized,
		reasonMintToken0Underpaid,
		reasonMintToken1Underpaid,
		reasonAmountSpecifiedZero,
		reasonInvalidSqrtPriceLimit,
		reasonSwapUnderpaid,
		reasonNoLiquidity,
		reasonFlashToken0Underpaid,
		reasonFlashToken1Underpaid,
		reasonLiquiditySubUnderflow,
		reasonLiquidityAddOverflow,
		reasonOracleNotInitialized,
		reasonOracleObservationTooOld,
		reasonNoPosition,
		reasonLiquidityOverflow,
		reasonTickOutOfBounds,
		reasonSqrtRatioOutOfBounds,
		reasonTransferFailed,
	}
	for _, reason := range reasons {
		t.Run(reason, func(t *testing.T) {
			data, err := abi.NewFunction("Error", abi.Types{abi.String}, nil).EncodeCall(reason)
			require.NoError(t, err)

			got, err := abi.DecodeRevert(data)
			require.NoError(t, err)
			require.Equal(t, reason, got)
		})
	}
}

func TestWrapRevertErrorZeroLengthDataFallsBackToErrReverted(t *testing.T) {
	p := &Pool{}
	rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted"}
	err := p.wrapRevertError(rpcErr)
	require.ErrorIs(t, err, ErrReverted)
}

func TestWrapRevertErrorEmptyHexDataFallsBackToErrReverted(t *testing.T) {
	p := &Pool{}
	rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"0x"`)}
	err := p.wrapRevertError(rpcErr)
	require.ErrorIs(t, err, ErrReverted)
}

func TestWrapRevertErrorKnownReasonMapsToSentinel(t *testing.T) {
	p := &Pool{}
	revertData, err := abi.NewFunction("Error", abi.Types{abi.String}, nil).EncodeCall(reasonNoPosition)
	require.NoError(t, err)
	hexData := `"` + encoding.Hex.EncodePrefixed(revertData) + `"`
	rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(hexData)}

	wrapped := p.wrapRevertError(rpcErr)
	require.ErrorIs(t, wrapped, ErrNoPosition)
}

func TestWrapRevertErrorNonRPCErrorPassesThrough(t *testing.T) {
	p := &Pool{}
	plain := errors.New("boom")
	err := p.wrapRevertError(plain)
	require.Same(t, plain, err)
}
