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

// go test -run '^$' -fuzz '^FuzzParseType$' -fuzztime=10s ./core/abi
func FuzzParseType(f *testing.F) {
	f.Add("uint256")
	f.Add("address[]")
	f.Add("(address,uint256)[]")
	f.Add("")
	f.Add("((((")
	f.Add("uint256[][][3]")
	f.Add("bytes33")
	f.Add("uint7")
	f.Add("bool")
	f.Add("function")
	f.Add("uint")
	f.Add("int")
	f.Add("bytes")
	f.Add("bytes32")
	f.Add("bytes0")
	f.Add("uint256[0]")
	f.Add("uint256[][]")
	f.Add("(bool,(address,uint256[]),string)[]")
	f.Add("()")
	f.Add("(())")
	f.Add("((),())")
	f.Add("uint256[-1]")
	f.Add("uint256]")
	f.Add("[uint256]")
	f.Add("uint256,uint256")
	f.Add(",")
	f.Add("(")
	f.Add(")")
	f.Add("uint256 ")
	f.Add("  ")
	f.Add("\t")
	f.Add("uint25600000000000000")
	f.Add("tuple(address,uint256)")

	f.Fuzz(func(t *testing.T, s string) {
		typ, err := ParseType(s)
		if err != nil {
			return
		}

		s2 := typ.String()
		typ2, err := ParseType(s2)
		require.NoError(t, err)
		require.Equal(t, typ, typ2)
	})
}

// go test -run '^$' -fuzz '^FuzzParseFunction$' -fuzztime=10s ./core/abi
func FuzzParseFunction(f *testing.F) {
	f.Add("transfer(address,uint256)")
	f.Add("transfer(address to, uint256 amount) returns (bool)")
	f.Add("function withdraw(uint256 amount) external onlyOwner returns (bool)")
	f.Add("")
	f.Add("garbage(((")
	f.Add("function mul(uint256 a, uint256 b) internal constant returns (uint256, uint256)")
	f.Add("swap(address[] memory path, uint256 amountIn, uint256 amountOutMin)")
	f.Add("swap(bytes calldata data)")
	f.Add("swap((address tokenIn, address tokenOut, uint256 amountIn, uint256 amountOutMin, address[] path, uint256 deadline) calldata params) external returns (uint256[] memory amounts)")
	f.Add("getUser() returns ((address,uint256))")
	f.Add("swap(address[] transient path)")
	f.Add("totalSupply()")
	f.Add("totalSupply() returns (uint256)")
	f.Add("getReserves() returns (uint112 reserve0, uint112 reserve1, uint32 blockTimestampLast)")
	f.Add("function noop() pure")
	f.Add("function noop() public pure virtual override")
	f.Add("foo((address,uint256)[] memory items) returns (bool[] memory)")
	f.Add("foo((address a, (uint256 x, uint256 y) point)[] memory items)")
	f.Add("f()")
	f.Add("f(")
	f.Add("f)")
	f.Add("()")
	f.Add("f(uint256")
	f.Add("f(uint256,)")
	f.Add("f(,uint256)")
	f.Add("f(uint256 a uint256 b)")
	f.Add("function()")
	f.Add("returns(bool)")
	f.Add("f() returns ()")
	f.Add("f() returns")
	f.Add("f() external external")
	f.Add("")
	f.Add("   ")
	f.Add("f(uint256[][2][] memory x)")

	f.Fuzz(func(t *testing.T, sig string) {
		fn, err := ParseFunction(sig, nil)
		if err != nil {
			return
		}

		fn2, err := ParseFunction(fn.Signature(), nil)
		require.NoError(t, err)
		require.Equal(t, fn.Name, fn2.Name)
		require.Equal(t, fn.Inputs, fn2.Inputs)
		require.Equal(t, fn.Selector(), fn2.Selector())
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeUintBigInt$' -fuzztime=10s ./core/abi
func FuzzDecodeUintBigInt(f *testing.F) {
	f.Add([]byte{}, uint8(0))
	f.Add([]byte{0x01}, uint8(0))
	f.Add(make([]byte, 32), uint8(255))

	f.Fuzz(func(t *testing.T, raw []byte, sizeByte uint8) {
		size := 65 + int(sizeByte)%192
		x := new(big.Int).SetBytes(raw)

		word, err := encodeUint(size, x)
		if err != nil {
			return
		}

		got, err := decodeUint(size, word)
		require.NoError(t, err)
		require.Zero(t, x.Cmp(got.(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeIntBigInt$' -fuzztime=10s ./core/abi
func FuzzDecodeIntBigInt(f *testing.F) {
	f.Add([]byte{}, false, uint8(0))
	f.Add([]byte{0x01}, true, uint8(0))
	f.Add(make([]byte, 32), true, uint8(255))

	f.Fuzz(func(t *testing.T, raw []byte, neg bool, sizeByte uint8) {
		size := 65 + int(sizeByte)%192
		x := new(big.Int).SetBytes(raw)
		if neg {
			x.Neg(x)
		}

		word, err := encodeInt(size, x)
		if err != nil {
			return
		}

		got, err := decodeInt(size, word)
		require.NoError(t, err)
		require.Zero(t, x.Cmp(got.(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeValue$' -fuzztime=10s ./core/abi
func FuzzDecodeValue(f *testing.F) {
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 0))
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 33))
	allFFSeed := make([]byte, 32)
	for i := range allFFSeed {
		allFFSeed[i] = 0xff
	}
	f.Add(allFFSeed)

	kinds := []Type{Bool, Address, Uint8, Uint256, Int8, Int256, Bytes4, Bytes32, FunctionType}

	f.Fuzz(func(t *testing.T, raw []byte) {
		for _, typ := range kinds {
			// Must never panic, regardless of word length or content;
			// an error is a perfectly fine outcome for garbage input.
			_, _ = decodeValue(typ, raw)
		}

		word := make([]byte, 32)
		copy(word, raw)
		for _, typ := range kinds {
			_, _ = decodeValue(typ, word)
		}
	})
}

// go test -run '^$' -fuzz '^FuzzUnpackBytes$' -fuzztime=10s ./core/abi
func FuzzUnpackBytes(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 16))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 64))
	hugeLen := make([]byte, 32)
	for i := range hugeLen {
		hugeLen[i] = 0xff
	}
	f.Add(hugeLen)

	f.Fuzz(func(t *testing.T, data []byte) {
		got, err := unpackBytes(data)
		if err != nil {
			return
		}
		require.LessOrEqual(t, len(got), len(data))
	})
}

// go test -run '^$' -fuzz '^FuzzPackUnpackBytesRoundTrip$' -fuzztime=10s ./core/abi
func FuzzPackUnpackBytesRoundTrip(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte("hello"))
	f.Add(make([]byte, 64))

	f.Fuzz(func(t *testing.T, data []byte) {
		packed := packBytes(data)
		got, err := unpackBytes(packed)
		require.NoError(t, err)
		require.Equal(t, data, got)
	})
}

// go test -run '^$' -fuzz '^FuzzPackUnpackTupleMixedRoundTrip$' -fuzztime=10s ./core/abi
func FuzzPackUnpackTupleMixedRoundTrip(f *testing.F) {
	f.Add(uint64(0), []byte(""), []byte{})
	f.Add(uint64(123456789), []byte("hello"), []byte{0xde, 0xad})
	f.Add(uint64(1)<<63, make([]byte, 40), make([]byte, 70))

	f.Fuzz(func(t *testing.T, amount uint64, strBytes []byte, dynBytes []byte) {
		types := Types{Uint64, String, Bytes}
		str := string(strBytes)
		args := []any{amount, str, dynBytes}

		data, err := packTuple(types, args)
		require.NoError(t, err)

		got, err := unpackTuple(types, data)
		require.NoError(t, err)
		require.Equal(t, amount, got[0])
		require.Equal(t, str, got[1])
		require.Equal(t, dynBytes, got[2])
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeArgCrashSafety$' -fuzztime=10s ./core/abi
func FuzzDecodeArgCrashSafety(f *testing.F) {
	f.Add(make([]byte, 0))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 64))
	allFFSeed := make([]byte, 64)
	for i := range allFFSeed {
		allFFSeed[i] = 0xff
	}
	f.Add(allFFSeed)

	innerTuple := Tuple(Uint8, String)
	kinds := []Type{
		Bool, Address, Uint256, Int256, Bytes32, String, Bytes,
		Slice(Uint256), Slice(String), Slice(innerTuple),
	}
	arrType, err := Array(Uint8, 3)
	if err != nil {
		f.Fatal(err)
	}
	kinds = append(kinds, arrType, innerTuple)

	f.Fuzz(func(t *testing.T, data []byte) {
		for _, typ := range kinds {
			_, _, _ = decodeArg(typ, data, 0)
		}
	})
}

// go test -run '^$' -fuzz '^FuzzParseError$' -fuzztime=10s ./core/abi
func FuzzParseError(f *testing.F) {
	f.Add("Foo(uint256)")
	f.Add("error Foo(uint256)")
	f.Add("error InsufficientBalance(uint256 available, uint256 required)")
	f.Add("errorCode(uint256)")
	f.Add("Paused()")
	f.Add("")
	f.Add("   ")
	f.Add("error")
	f.Add("error ")
	f.Add("error (uint256)")
	f.Add("garbage(((")
	f.Add("Foo(")
	f.Add("Foo)")
	f.Add("()")
	f.Add("Foo(uint256")
	f.Add("Foo(uint256,)")
	f.Add("Foo(,uint256)")
	f.Add("Foo(uint256 a uint256 b)")
	f.Add("Foo(uint256) extra")
	f.Add("Foo(uint256) returns (bool)")
	f.Add("BatchFailed((address,uint256)[] memory items)")
	f.Add("Foo(uint256[][2][] memory x)")

	f.Fuzz(func(t *testing.T, sig string) {
		e, err := ParseError(sig)
		if err != nil {
			return
		}

		e2, err := ParseError(e.Signature())
		require.NoError(t, err)
		require.Equal(t, e.Name, e2.Name)
		require.Equal(t, e.Inputs, e2.Inputs)
		require.Equal(t, e.Selector(), e2.Selector())
	})
}

// go test -run '^$' -fuzz '^FuzzErrorDecodeCrashSafety$' -fuzztime=10s ./core/abi
func FuzzErrorDecodeCrashSafety(f *testing.F) {
	e, err := ParseError("InsufficientBalance(uint256,uint256)")
	if err != nil {
		f.Fatal(err)
	}
	sel := e.Selector()

	f.Add([]byte{})
	f.Add(sel[:])
	f.Add(append(sel[:], make([]byte, 64)...))
	f.Add(make([]byte, 4))
	f.Add(make([]byte, 100))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = e.Decode(data)
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeRevertCrashSafety$' -fuzztime=10s ./core/abi
func FuzzDecodeRevertCrashSafety(f *testing.F) {
	f.Add([]byte{})
	f.Add(errorSelector[:])
	f.Add(panicSelector[:])
	errData, _ := Pack(Types{String}, "Multicall3: call failed")
	f.Add(append(errorSelector[:], errData...))
	panicData, _ := Pack(Types{Uint256}, big.NewInt(0x11))
	f.Add(append(panicSelector[:], panicData...))
	f.Add(make([]byte, 4))
	f.Add(make([]byte, 100))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeRevert(data)
	})
}
