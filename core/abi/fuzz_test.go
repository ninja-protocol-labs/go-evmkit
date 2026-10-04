package abi

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// go test -run '^$' -fuzz '^FuzzPackBytes$' -fuzztime=10s ./core/abi
func FuzzPackBytes(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x01})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 33))
	f.Add(make([]byte, 65))

	f.Fuzz(func(t *testing.T, data []byte) {
		out := packBytes(data)

		require.Zero(t, len(out)%32)
		require.GreaterOrEqual(t, len(out), 32)

		require.Equal(t, encodeUint64Word(uint64(len(data))), out[:32])
		require.Equal(t, data, out[32:32+len(data)])
		require.Equal(t, make([]byte, len(out)-32-len(data)), out[32+len(data):])
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeFixedBytes$' -fuzztime=10s ./core/abi
func FuzzEncodeFixedBytes(f *testing.F) {
	f.Add([]byte{}, uint8(0))
	f.Add([]byte{0xff}, uint8(0))
	f.Add(make([]byte, 32), uint8(31))
	f.Add(make([]byte, 4), uint8(10))

	f.Fuzz(func(t *testing.T, data []byte, sizeByte uint8) {
		size := int(sizeByte%32) + 1

		got, err := encodeFixedBytes(size, data)
		if len(data) != size {
			require.ErrorIs(t, err, ErrByteLengthMismatch)
			return
		}
		require.NoError(t, err)
		require.Len(t, got, 32)
		require.Equal(t, data, got[:size])
		require.Equal(t, make([]byte, 32-size), got[size:])
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeUintBigInt$' -fuzztime=10s ./core/abi
func FuzzEncodeUintBigInt(f *testing.F) {
	f.Add([]byte{}, uint8(0))
	f.Add([]byte{0x01}, uint8(0))
	f.Add(make([]byte, 32), uint8(255))

	f.Fuzz(func(t *testing.T, raw []byte, sizeByte uint8) {
		size := 65 + int(sizeByte)%192
		x := new(big.Int).SetBytes(raw)

		got, err := encodeUint(size, x)

		bnd, ok := boundsFor(size)
		require.True(t, ok)

		if x.Cmp(bnd.uintMax) >= 0 {
			require.ErrorIs(t, err, ErrIntegerOutOfRange)
			return
		}
		require.NoError(t, err)
		require.Len(t, got, 32)

		gotVal := new(big.Int).SetBytes(got)
		require.Zero(t, x.Cmp(gotVal))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeIntBigInt$' -fuzztime=10s ./core/abi
func FuzzEncodeIntBigInt(f *testing.F) {
	f.Add([]byte{}, false, uint8(0))
	f.Add([]byte{0x01}, true, uint8(0))
	f.Add(make([]byte, 32), true, uint8(255))

	f.Fuzz(func(t *testing.T, raw []byte, neg bool, sizeByte uint8) {
		size := 65 + int(sizeByte)%192
		x := new(big.Int).SetBytes(raw)
		if neg {
			x.Neg(x)
		}

		got, err := encodeInt(size, x)

		bnd, ok := boundsFor(size)
		require.True(t, ok)

		if x.Cmp(bnd.intMin) < 0 || x.Cmp(bnd.intMax) > 0 {
			require.ErrorIs(t, err, ErrIntegerOutOfRange)
			return
		}
		require.NoError(t, err)
		require.Len(t, got, 32)

		gotVal := new(big.Int).SetBytes(got)
		if got[0]&0x80 != 0 {
			gotVal.Sub(gotVal, twoPow256)
		}
		require.Zero(t, x.Cmp(gotVal))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeSliceOfUint8$' -fuzztime=10s ./core/abi
func FuzzEncodeSliceOfUint8(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x01, 0x02, 0x03})
	f.Add(make([]byte, 64))

	f.Fuzz(func(t *testing.T, data []byte) {
		got, err := encodeArg(Slice(Uint8), data)
		require.NoError(t, err)
		require.Len(t, got, 32+32*len(data))
		require.Equal(t, encodeUint64Word(uint64(len(data))), got[:32])

		for i, b := range data {
			want, err := encodeValue(Uint8, b)
			require.NoError(t, err)
			require.Equal(t, want, got[32+32*i:32+32*(i+1)])
		}
	})
}

// go test -run '^$' -fuzz '^FuzzPackTupleMixed$' -fuzztime=10s ./core/abi
func FuzzPackTupleMixed(f *testing.F) {
	f.Add(uint64(0), []byte(""), []byte{})
	f.Add(uint64(123456789), []byte("hello"), []byte{0xde, 0xad})
	f.Add(uint64(1)<<63, make([]byte, 40), make([]byte, 70))

	f.Fuzz(func(t *testing.T, amount uint64, strBytes []byte, dynBytes []byte) {
		typs := Types{Uint64, String, Bytes}
		str := string(strBytes)
		args := []any{amount, str, dynBytes}

		got, err := packTuple(typs, args)
		require.NoError(t, err)
		require.Zero(t, len(got)%32)

		const headSize = 32 * 3
		require.GreaterOrEqual(t, len(got), headSize)

		wantHead0, err := encodeValue(Uint64, amount)
		require.NoError(t, err)
		require.Equal(t, wantHead0, got[:32])

		off1 := new(big.Int).SetBytes(got[32:64]).Int64()
		off2 := new(big.Int).SetBytes(got[64:96]).Int64()

		wantTail1, err := encodeDynamicBytes(String, str)
		require.NoError(t, err)
		wantTail2, err := encodeDynamicBytes(Bytes, dynBytes)
		require.NoError(t, err)

		require.EqualValues(t, headSize, off1)
		require.EqualValues(t, headSize+len(wantTail1), off2)

		require.Equal(t, wantTail1, got[off1:off1+int64(len(wantTail1))])
		require.Equal(t, wantTail2, got[off2:off2+int64(len(wantTail2))])
		require.EqualValues(t, len(got), int(off2)+len(wantTail2))
	})
}

// go test -run '^$' -fuzz '^FuzzBoundsFor$' -fuzztime=10s ./core/abi
func FuzzBoundsFor(f *testing.F) {
	f.Add(0)
	f.Add(64)
	f.Add(65)
	f.Add(256)
	f.Add(257)
	f.Add(-100)

	f.Fuzz(func(t *testing.T, size int) {
		bnd, ok := boundsFor(size)
		if size < 65 || size > 256 {
			require.False(t, ok)
			return
		}
		require.True(t, ok)
		require.NotNil(t, bnd.uintMax)
		require.NotNil(t, bnd.intMin)
		require.NotNil(t, bnd.intMax)

		wantMax := new(big.Int).Lsh(big.NewInt(1), uint(size))
		require.Zero(t, wantMax.Cmp(bnd.uintMax))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeSliceOfAddress$' -fuzztime=10s ./core/abi
func FuzzEncodeSliceOfAddress(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 20))
	f.Add(make([]byte, 45))

	f.Fuzz(func(t *testing.T, raw []byte) {
		n := len(raw) / 20
		addrs := make([]*types.Address, n)
		for i := range n {
			addrs[i] = types.NewAddressFromBytes(raw[i*20 : (i+1)*20])
		}

		got, err := encodeArg(Slice(Address), addrs)
		require.NoError(t, err)
		require.Len(t, got, 32+32*n)
		require.Equal(t, encodeUint64Word(uint64(n)), got[:32])

		for i, a := range addrs {
			want, err := encodeValue(Address, a)
			require.NoError(t, err)
			require.Equal(t, want, got[32+32*i:32+32*(i+1)])
		}
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeArrayOfUint256$' -fuzztime=10s ./core/abi
func FuzzEncodeArrayOfUint256(f *testing.F) {
	f.Add(uint64(0), uint64(0), uint64(0))
	f.Add(uint64(1), uint64(2), uint64(3))

	f.Fuzz(func(t *testing.T, a, b, c uint64) {
		typ, err := Array(Uint256, 3)
		require.NoError(t, err)

		vals := []*big.Int{
			new(big.Int).SetUint64(a),
			new(big.Int).SetUint64(b),
			new(big.Int).SetUint64(c),
		}
		got, err := encodeArg(typ, vals)
		require.NoError(t, err)
		require.Len(t, got, 96)

		for i, v := range vals {
			want, err := encodeValue(Uint256, v)
			require.NoError(t, err)
			require.Equal(t, want, got[32*i:32*(i+1)])
		}
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeTupleValueStrings$' -fuzztime=10s ./core/abi
func FuzzEncodeTupleValueStrings(f *testing.F) {
	f.Add([]byte("a"), []byte("bb"))
	f.Add([]byte{}, []byte{})

	f.Fuzz(func(t *testing.T, s1, s2 []byte) {
		typ := Tuple(String, String)
		args := []any{string(s1), string(s2)}

		got, err := encodeTupleValue(typ, args)
		require.NoError(t, err)

		want, err := packTuple(typ.Components, args)
		require.NoError(t, err)
		require.Equal(t, want, got)
	})
}
