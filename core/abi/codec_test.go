package abi

import (
	"encoding/hex"
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestEncodeBool(t *testing.T) {
	b, err := encodeValue(Bool, true)
	require.NoError(t, err)
	want, err := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000001")
	require.NoError(t, err)
	require.Equal(t, want, b)

	b, err = encodeValue(Bool, false)
	require.NoError(t, err)
	require.Equal(t, make([]byte, 32), b)

	_, err = encodeValue(Bool, "true")
	require.ErrorIs(t, err, ErrInvalidGoType)
}

func TestEncodeAddress(t *testing.T) {
	addr, err := types.NewAddressFromHex("0x833e1D0b8Bc979D49d57b65dCF18364694B16D52")
	require.NoError(t, err)

	b, err := encodeValue(Address, addr)
	require.NoError(t, err)
	want, err := hex.DecodeString("000000000000000000000000833e1d0b8bc979d49d57b65dcf18364694b16d52")
	require.NoError(t, err)
	require.Equal(t, want, b)

	_, err = encodeValue(Address, "0x833e1D0b8Bc979D49d57b65dCF18364694B16D52")
	require.ErrorIs(t, err, ErrInvalidGoType)
}

func TestEncodeAddressNilRejected(t *testing.T) {
	var addr *types.Address
	_, err := encodeValue(Address, addr)
	require.ErrorIs(t, err, ErrNilAddress)
}

func TestEncodeUintNative(t *testing.T) {
	tests := []struct {
		typ  Type
		v    any
		want string
	}{
		{Uint8, uint8(255), "00000000000000000000000000000000000000000000000000000000000000ff"},
		{Uint16, uint16(0x1234), "0000000000000000000000000000000000000000000000000000000000001234"},
		{Uint32, uint32(0xdeadbeef), "00000000000000000000000000000000000000000000000000000000deadbeef"},
		{Uint64, uint64(0x0123456789abcdef), "0000000000000000000000000000000000000000000000000123456789abcdef"},
	}
	for _, tt := range tests {
		got, err := encodeValue(tt.typ, tt.v)
		require.NoError(t, err)
		want, err := hex.DecodeString(tt.want)
		require.NoError(t, err)
		require.Equal(t, want, got)
	}
}

func TestEncodeUint256BigInt(t *testing.T) {
	n, ok := new(big.Int).SetString("1000000000000000000", 10)
	require.True(t, ok)

	got, err := encodeValue(Uint256, n)
	require.NoError(t, err)
	want, err := hex.DecodeString("0000000000000000000000000000000000000000000000000de0b6b3a7640000")
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestEncodeUintRejectsWrongGoType(t *testing.T) {
	_, err := encodeValue(Uint8, uint16(1))
	require.ErrorIs(t, err, ErrInvalidGoType)
}

func TestEncodeUintBigIntNegativeRejected(t *testing.T) {
	_, err := encodeValue(Uint256, big.NewInt(-1))
	require.ErrorIs(t, err, ErrIntegerOutOfRange)
}

func TestEncodeUintBigIntOverflowRejected(t *testing.T) {
	overflow := new(big.Int).Lsh(big.NewInt(1), 128)
	_, err := encodeValue(Uint128, overflow)
	require.ErrorIs(t, err, ErrIntegerOutOfRange)

	maxValid := new(big.Int).Sub(overflow, big.NewInt(1))
	_, err = encodeValue(Uint128, maxValid)
	require.NoError(t, err)
}

func TestEncodeIntNativePositive(t *testing.T) {
	got, err := encodeValue(Int32, int32(42))
	require.NoError(t, err)
	want, err := hex.DecodeString("000000000000000000000000000000000000000000000000000000000000002a")
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestEncodeIntNativeNegativeOne(t *testing.T) {
	got, err := encodeValue(Int8, int8(-1))
	require.NoError(t, err)
	want, err := hex.DecodeString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestEncodeIntNegativeTwo(t *testing.T) {
	got, err := encodeValue(Int64, int64(-2))
	require.NoError(t, err)
	want, err := hex.DecodeString("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe")
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestEncodeInt256BigIntNegative(t *testing.T) {
	n := big.NewInt(-1000000000000000000)
	got, err := encodeValue(Int256, n)
	require.NoError(t, err)
	want, err := hex.DecodeString("fffffffffffffffffffffffffffffffffffffffffffffffff21f494c589c0000")
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestEncodeIntBigIntRangeRejected(t *testing.T) {
	half := new(big.Int).Lsh(big.NewInt(1), 127)
	_, err := encodeValue(Int128, half)
	require.ErrorIs(t, err, ErrIntegerOutOfRange)

	maxValid := new(big.Int).Sub(half, big.NewInt(1))
	_, err = encodeValue(Int128, maxValid)
	require.NoError(t, err)

	minValid := new(big.Int).Neg(half)
	_, err = encodeValue(Int128, minValid)
	require.NoError(t, err)

	belowMin := new(big.Int).Sub(minValid, big.NewInt(1))
	_, err = encodeValue(Int128, belowMin)
	require.ErrorIs(t, err, ErrIntegerOutOfRange)
}

func TestEncodeFixedBytes(t *testing.T) {
	got, err := encodeValue(Bytes4, []byte{0xde, 0xad, 0xbe, 0xef})
	require.NoError(t, err)
	want, err := hex.DecodeString("deadbeef00000000000000000000000000000000000000000000000000000000")
	require.NoError(t, err)
	require.Equal(t, want, got)

	full := make([]byte, 32)
	for i := range full {
		full[i] = byte(i + 1)
	}
	got, err = encodeValue(Bytes32, full)
	require.NoError(t, err)
	require.Equal(t, full, got)
}

func TestEncodeFixedBytesWrongLength(t *testing.T) {
	_, err := encodeValue(Bytes4, []byte{0x01, 0x02})
	require.ErrorIs(t, err, ErrByteLengthMismatch)
}

func TestEncodeFunction(t *testing.T) {
	fn := make([]byte, 24)
	for i := range fn {
		fn[i] = byte(i + 1)
	}
	got, err := encodeValue(FunctionType, fn)
	require.NoError(t, err)
	require.Equal(t, fn, got[:24])
	require.Equal(t, make([]byte, 8), got[24:])
}

func TestEncodeFunctionWrongLength(t *testing.T) {
	_, err := encodeValue(FunctionType, make([]byte, 20))
	require.ErrorIs(t, err, ErrByteLengthMismatch)
}

func TestEncodeValueRejectsDynamicKind(t *testing.T) {
	_, err := encodeValue(String, "hi")
	require.ErrorIs(t, err, ErrUnsupportedKind)
}

func TestEncodeUintOutOfRangeSizeRejected(t *testing.T) {
	_, err := encodeValue(Type{Kind: KindUint, Size: 264}, big.NewInt(1))
	require.ErrorIs(t, err, ErrUnsupportedKind)
}

func TestEncodeIntOutOfRangeSizeRejected(t *testing.T) {
	_, err := encodeValue(Type{Kind: KindInt, Size: 264}, big.NewInt(1))
	require.ErrorIs(t, err, ErrUnsupportedKind)
}

func TestEncodeUintEveryNonNativeSizeCovered(t *testing.T) {
	for size := 65; size <= 256; size++ {
		b, ok := boundsFor(size)
		require.True(t, ok, "size %d", size)
		require.NotNil(t, b.uintMax)
		require.NotNil(t, b.intMin)
		require.NotNil(t, b.intMax)
	}
}

func TestBoundsForRejectsOutsideRange(t *testing.T) {
	_, ok := boundsFor(64)
	require.False(t, ok)

	_, ok = boundsFor(257)
	require.False(t, ok)

	_, ok = boundsFor(0)
	require.False(t, ok)

	_, ok = boundsFor(-8)
	require.False(t, ok)
}

func TestEncodeUintNativeBoundaries(t *testing.T) {
	tests := []struct {
		typ Type
		v   any
	}{
		{Uint8, uint8(0)},
		{Uint8, uint8(math.MaxUint8)},
		{Uint16, uint16(0)},
		{Uint16, uint16(math.MaxUint16)},
		{Uint32, uint32(0)},
		{Uint32, uint32(math.MaxUint32)},
		{Uint64, uint64(0)},
		{Uint64, uint64(math.MaxUint64)},
	}
	for _, tt := range tests {
		got, err := encodeValue(tt.typ, tt.v)
		require.NoError(t, err)
		require.Len(t, got, 32)
	}
}

func TestEncodeIntNativeBoundaries(t *testing.T) {
	tests := []struct {
		typ  Type
		v    any
		want string
	}{
		{Int8, int8(0), "0000000000000000000000000000000000000000000000000000000000000000"},
		{Int8, int8(math.MaxInt8), "000000000000000000000000000000000000000000000000000000000000007f"},
		{Int8, int8(math.MinInt8), "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff80"},
		{Int16, int16(math.MaxInt16), "0000000000000000000000000000000000000000000000000000000000007fff"},
		{Int16, int16(math.MinInt16), "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8000"},
		{Int32, int32(math.MaxInt32), "000000000000000000000000000000000000000000000000000000007fffffff"},
		{Int32, int32(math.MinInt32), "ffffffffffffffffffffffffffffffffffffffffffffffffffffffff80000000"},
		{Int64, int64(math.MaxInt64), "0000000000000000000000000000000000000000000000007fffffffffffffff"},
		{Int64, int64(math.MinInt64), "ffffffffffffffffffffffffffffffffffffffffffffffff8000000000000000"},
	}
	for _, tt := range tests {
		got, err := encodeValue(tt.typ, tt.v)
		require.NoError(t, err)
		want, err := hex.DecodeString(tt.want)
		require.NoError(t, err)
		require.Equal(t, want, got)
	}
}

func TestEncodeUintBigIntBoundaries(t *testing.T) {
	for _, size := range []int{65, 72, 128, 255, 256} {
		m, ok := boundsFor(size)
		require.True(t, ok)

		got, err := encodeValue(Type{Kind: KindUint, Size: size}, big.NewInt(0))
		require.NoError(t, err)
		require.Len(t, got, 32)

		maxValid := new(big.Int).Sub(m.uintMax, big.NewInt(1))
		got, err = encodeValue(Type{Kind: KindUint, Size: size}, maxValid)
		require.NoError(t, err)
		require.Len(t, got, 32)

		_, err = encodeValue(Type{Kind: KindUint, Size: size}, m.uintMax)
		require.ErrorIs(t, err, ErrIntegerOutOfRange)

		overMax := new(big.Int).Add(m.uintMax, big.NewInt(1))
		_, err = encodeValue(Type{Kind: KindUint, Size: size}, overMax)
		require.ErrorIs(t, err, ErrIntegerOutOfRange)
	}
}

func TestEncodeIntBigIntBoundaries(t *testing.T) {
	for _, size := range []int{65, 72, 128, 255, 256} {
		bnd, ok := boundsFor(size)
		require.True(t, ok)

		got, err := encodeValue(Type{Kind: KindInt, Size: size}, big.NewInt(0))
		require.NoError(t, err)
		require.Len(t, got, 32)

		got, err = encodeValue(Type{Kind: KindInt, Size: size}, bnd.intMax)
		require.NoError(t, err)
		require.Len(t, got, 32)

		got, err = encodeValue(Type{Kind: KindInt, Size: size}, bnd.intMin)
		require.NoError(t, err)
		require.Len(t, got, 32)

		overMax := new(big.Int).Add(bnd.intMax, big.NewInt(1))
		_, err = encodeValue(Type{Kind: KindInt, Size: size}, overMax)
		require.ErrorIs(t, err, ErrIntegerOutOfRange)

		underMin := new(big.Int).Sub(bnd.intMin, big.NewInt(1))
		_, err = encodeValue(Type{Kind: KindInt, Size: size}, underMin)
		require.ErrorIs(t, err, ErrIntegerOutOfRange)
	}
}

func TestEncodeUintSize64RejectsBigInt(t *testing.T) {
	_, err := encodeValue(Uint64, big.NewInt(1))
	require.ErrorIs(t, err, ErrInvalidGoType)
}

func TestEncodeIntSize64RejectsBigInt(t *testing.T) {
	_, err := encodeValue(Int64, big.NewInt(1))
	require.ErrorIs(t, err, ErrInvalidGoType)
}

func TestEncodeUintSize72RejectsNativeUint64(t *testing.T) {
	_, err := encodeValue(Type{Kind: KindUint, Size: 72}, uint64(1))
	require.ErrorIs(t, err, ErrInvalidGoType)
}

func TestEncodeAddressZeroValue(t *testing.T) {
	addr := types.NewAddressFromBytes(nil)
	got, err := encodeValue(Address, addr)
	require.NoError(t, err)
	require.Equal(t, make([]byte, 32), got)
}

func TestEncodeFixedBytesEmptyRejected(t *testing.T) {
	_, err := encodeValue(Bytes1, []byte{})
	require.ErrorIs(t, err, ErrByteLengthMismatch)
}

func TestEncodeFixedBytesSingleByte(t *testing.T) {
	got, err := encodeValue(Bytes1, []byte{0xff})
	require.NoError(t, err)
	want, err := hex.DecodeString("ff00000000000000000000000000000000000000000000000000000000000000")
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestEncodeFunctionEmptyRejected(t *testing.T) {
	_, err := encodeValue(FunctionType, []byte{})
	require.ErrorIs(t, err, ErrByteLengthMismatch)
}

func TestEncodeBoolWrongGoTypeVariants(t *testing.T) {
	for _, v := range []any{1, uint8(1), nil, []byte{1}} {
		_, err := encodeValue(Bool, v)
		require.ErrorIs(t, err, ErrInvalidGoType)
	}
}
