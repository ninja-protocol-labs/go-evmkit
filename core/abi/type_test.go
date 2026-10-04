package abi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsDynamicStaticScalars(t *testing.T) {
	tests := []Type{
		Bool, Address, FunctionType,
		Uint8, Uint16, Uint32, Uint64, Uint128, Uint256,
		Int8, Int16, Int32, Int64, Int128, Int256,
		Bytes1, Bytes20, Bytes32,
	}
	for _, typ := range tests {
		require.False(t, typ.IsDynamic(), "kind %v should be static", typ.Kind)
	}
}

func TestIsDynamicDynamicScalars(t *testing.T) {
	tests := []Type{String, Bytes}
	for _, typ := range tests {
		require.True(t, typ.IsDynamic(), "kind %v should be dynamic", typ.Kind)
	}
}

func TestIsDynamicSlice(t *testing.T) {
	uintSlice := Slice(Uint256)
	require.True(t, uintSlice.IsDynamic())
	boolSlice := Slice(Bool)
	require.True(t, boolSlice.IsDynamic())
}

func TestIsDynamicArray(t *testing.T) {
	staticArr, err := Array(Uint256, 3)
	require.NoError(t, err)
	require.False(t, staticArr.IsDynamic())

	dynamicArr, err := Array(Bytes, 3)
	require.NoError(t, err)
	require.True(t, dynamicArr.IsDynamic())

	nestedArr, err := Array(Slice(Uint256), 2)
	require.NoError(t, err)
	require.True(t, nestedArr.IsDynamic())
}

func TestIsDynamicTuple(t *testing.T) {
	staticTuple := Tuple(Address, Uint256, Bool)
	require.False(t, staticTuple.IsDynamic())

	dynamicTuple := Tuple(Address, String)
	require.True(t, dynamicTuple.IsDynamic())

	nestedDynamicTuple := Tuple(Address, Tuple(Bytes))
	require.True(t, nestedDynamicTuple.IsDynamic())

	arr, err := Array(Uint256, 2)
	require.NoError(t, err)
	tupleWithStaticArray := Tuple(Address, arr)
	require.False(t, tupleWithStaticArray.IsDynamic())

	sliceInTuple := Tuple(Address, Slice(Uint256))
	require.True(t, sliceInTuple.IsDynamic())
}

func TestArrayRejectsNegativeSize(t *testing.T) {
	_, err := Array(Uint256, -1)
	require.ErrorIs(t, err, ErrInvalidArraySize)
}

func TestArrayAllowsZeroSize(t *testing.T) {
	zeroArr, err := Array(Uint256, 0)
	require.NoError(t, err)
	require.Equal(t, 0, zeroArr.Size)
}

func TestNamedTupleLengthMismatch(t *testing.T) {
	_, err := NamedTuple([]string{"a"}, Address, Uint256)
	require.ErrorIs(t, err, ErrTupleNameMismatch)
}

func TestNamedTupleMatchingLength(t *testing.T) {
	typ, err := NamedTuple([]string{"to", "amount"}, Address, Uint256)
	require.NoError(t, err)
	require.Equal(t, []string{"to", "amount"}, typ.Names)
	require.Len(t, typ.Components, 2)
}

func TestNewTypes(t *testing.T) {
	types := NewTypes(Address, Uint256, Bool)
	require.Len(t, types, 3)
	require.Equal(t, KindAddress, types[0].Kind)
	require.Equal(t, KindUint, types[1].Kind)
	require.Equal(t, KindBool, types[2].Kind)
}

func TestCanonicalUintSizes(t *testing.T) {
	tests := []struct {
		typ  Type
		size int
	}{
		{Uint8, 8}, {Uint16, 16}, {Uint32, 32}, {Uint64, 64}, {Uint128, 128}, {Uint256, 256},
	}
	for _, tt := range tests {
		require.Equal(t, KindUint, tt.typ.Kind)
		require.Equal(t, tt.size, tt.typ.Size)
	}
}

func TestCanonicalIntSizes(t *testing.T) {
	tests := []struct {
		typ  Type
		size int
	}{
		{Int8, 8}, {Int16, 16}, {Int32, 32}, {Int64, 64}, {Int128, 128}, {Int256, 256},
	}
	for _, tt := range tests {
		require.Equal(t, KindInt, tt.typ.Kind)
		require.Equal(t, tt.size, tt.typ.Size)
	}
}

func TestCanonicalFixedBytesSizes(t *testing.T) {
	tests := []struct {
		typ  Type
		size int
	}{
		{Bytes1, 1}, {Bytes4, 4}, {Bytes16, 16}, {Bytes20, 20}, {Bytes32, 32},
	}
	for _, tt := range tests {
		require.Equal(t, KindFixedBytes, tt.typ.Kind)
		require.Equal(t, tt.size, tt.typ.Size)
	}
}

func TestSliceElem(t *testing.T) {
	s := Slice(Address)
	require.Equal(t, KindSlice, s.Kind)
	require.NotNil(t, s.Elem)
	require.Equal(t, KindAddress, s.Elem.Kind)
}

func TestArrayElemAndSize(t *testing.T) {
	a, err := Array(Bytes32, 4)
	require.NoError(t, err)
	require.Equal(t, KindArray, a.Kind)
	require.Equal(t, 4, a.Size)
	require.NotNil(t, a.Elem)
	require.Equal(t, KindFixedBytes, a.Elem.Kind)
	require.Equal(t, 32, a.Elem.Size)
}
