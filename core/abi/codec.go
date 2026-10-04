package abi

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"reflect"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

var (
	// twoPow256 is 2^256, used for *big.Int two's complement.
	twoPow256 = new(big.Int).Lsh(big.NewInt(1), 256)

	// allFF is a 32-byte source buffer for sign-extension fills.
	allFF = [32]byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	}
)

// bounds is the uint/int range for one bit width above 64.
type bounds struct {
	uintMax, intMin, intMax *big.Int
}

// boundsTable holds precomputed bounds for widths 65..256, indexed by size-65.
var boundsTable [192]bounds

func init() {
	for size := 65; size <= 256; size++ {
		half := new(big.Int).Lsh(big.NewInt(1), uint(size-1))
		boundsTable[size-65] = bounds{
			uintMax: new(big.Int).Lsh(big.NewInt(1), uint(size)),
			intMin:  new(big.Int).Neg(half),
			intMax:  new(big.Int).Sub(half, big.NewInt(1)),
		}
	}
}

// boundsFor returns the precomputed bounds for size, or ok=false if size is outside 65..256.
func boundsFor(size int) (b bounds, ok bool) {
	if size < 65 || size > 256 {
		return bounds{}, false
	}
	return boundsTable[size-65], true
}

// encodeValue encodes a static scalar value as a single 32-byte ABI word.
func encodeValue(t Type, v any) ([]byte, error) {
	switch t.Kind {
	case KindBool:
		return encodeBool(v)
	case KindAddress:
		return encodeAddress(v)
	case KindFixedBytes:
		return encodeFixedBytes(t.Size, v)
	case KindFunction:
		return encodeFunction(v)
	case KindUint:
		return encodeUint(t.Size, v)
	case KindInt:
		return encodeInt(t.Size, v)
	default:
		return nil, fmt.Errorf("%w: %v is not a static scalar type", ErrUnsupportedKind, t.Kind)
	}
}

func encodeBool(v any) ([]byte, error) {
	b, ok := v.(bool)
	if !ok {
		return nil, fmt.Errorf("%w: expected bool, got %T", ErrInvalidGoType, v)
	}
	word := make([]byte, 32)
	if b {
		word[31] = 1
	}
	return word, nil
}

func encodeAddress(v any) ([]byte, error) {
	addr, ok := v.(*types.Address)
	if !ok {
		return nil, fmt.Errorf("%w: expected *types.Address, got %T", ErrInvalidGoType, v)
	}
	if addr == nil {
		return nil, ErrNilAddress
	}
	word := make([]byte, 32)
	copy(word[32-types.AddressLength:], addr.Bytes())
	return word, nil
}

// encodeFixedBytes right-pads b into a 32-byte word, unlike the other static types.
func encodeFixedBytes(size int, v any) ([]byte, error) {
	b, ok := v.([]byte)
	if !ok {
		return nil, fmt.Errorf("%w: expected []byte, got %T", ErrInvalidGoType, v)
	}
	if len(b) != size {
		return nil, fmt.Errorf("%w: bytes%d expects %d bytes, got %d", ErrByteLengthMismatch, size, size, len(b))
	}
	word := make([]byte, 32)
	copy(word, b)
	return word, nil
}

// encodeFunction encodes a 20-byte address + 4-byte selector, right-padded like bytes24.
func encodeFunction(v any) ([]byte, error) {
	b, ok := v.([]byte)
	if !ok {
		return nil, fmt.Errorf("%w: expected []byte, got %T", ErrInvalidGoType, v)
	}
	if len(b) != 24 {
		return nil, fmt.Errorf("%w: function type expects 24 bytes, got %d", ErrByteLengthMismatch, len(b))
	}
	word := make([]byte, 32)
	copy(word, b)
	return word, nil
}

// encodeUint encodes an unsigned integer as a big-endian 32-byte word.
// size<=64 takes the matching native Go type; size>64 takes a *big.Int.
func encodeUint(size int, v any) ([]byte, error) {
	switch size {
	case 8:
		x, ok := v.(uint8)
		if !ok {
			return nil, fmt.Errorf("%w: expected uint8, got %T", ErrInvalidGoType, v)
		}
		return encodeUint64Word(uint64(x)), nil
	case 16:
		x, ok := v.(uint16)
		if !ok {
			return nil, fmt.Errorf("%w: expected uint16, got %T", ErrInvalidGoType, v)
		}
		return encodeUint64Word(uint64(x)), nil
	case 32:
		x, ok := v.(uint32)
		if !ok {
			return nil, fmt.Errorf("%w: expected uint32, got %T", ErrInvalidGoType, v)
		}
		return encodeUint64Word(uint64(x)), nil
	case 64:
		x, ok := v.(uint64)
		if !ok {
			return nil, fmt.Errorf("%w: expected uint64, got %T", ErrInvalidGoType, v)
		}
		return encodeUint64Word(x), nil
	default:
		word := make([]byte, 32)
		x, ok := v.(*big.Int)
		if !ok {
			return nil, fmt.Errorf("%w: expected *big.Int, got %T", ErrInvalidGoType, v)
		}
		if x.Sign() < 0 {
			return nil, fmt.Errorf("%w: uint%d cannot be negative: %s", ErrIntegerOutOfRange, size, x)
		}
		bnd, ok := boundsFor(size)
		if !ok {
			return nil, fmt.Errorf("%w: uint%d is not a standard ABI width", ErrUnsupportedKind, size)
		}
		if x.Cmp(bnd.uintMax) >= 0 {
			return nil, fmt.Errorf("%w: uint%d overflow: %s", ErrIntegerOutOfRange, size, x)
		}
		b := x.Bytes()
		copy(word[32-len(b):], b)
		return word, nil
	}
}

// encodeInt encodes a signed integer as a 32-byte two's complement word.
// size<=64 takes the matching native Go type; size>64 takes a *big.Int.
func encodeInt(size int, v any) ([]byte, error) {
	switch size {
	case 8:
		x, ok := v.(int8)
		if !ok {
			return nil, fmt.Errorf("%w: expected int8, got %T", ErrInvalidGoType, v)
		}
		return encodeInt64Word(int64(x)), nil
	case 16:
		x, ok := v.(int16)
		if !ok {
			return nil, fmt.Errorf("%w: expected int16, got %T", ErrInvalidGoType, v)
		}
		return encodeInt64Word(int64(x)), nil
	case 32:
		x, ok := v.(int32)
		if !ok {
			return nil, fmt.Errorf("%w: expected int32, got %T", ErrInvalidGoType, v)
		}
		return encodeInt64Word(int64(x)), nil
	case 64:
		x, ok := v.(int64)
		if !ok {
			return nil, fmt.Errorf("%w: expected int64, got %T", ErrInvalidGoType, v)
		}
		return encodeInt64Word(x), nil
	default:
		x, ok := v.(*big.Int)
		if !ok {
			return nil, fmt.Errorf("%w: expected *big.Int, got %T", ErrInvalidGoType, v)
		}
		bnd, ok := boundsFor(size)
		if !ok {
			return nil, fmt.Errorf("%w: int%d is not a standard ABI width", ErrUnsupportedKind, size)
		}
		if x.Cmp(bnd.intMin) < 0 || x.Cmp(bnd.intMax) > 0 {
			return nil, fmt.Errorf("%w: int%d overflow: %s", ErrIntegerOutOfRange, size, x)
		}

		word := make([]byte, 32)
		if x.Sign() >= 0 {
			b := x.Bytes()
			copy(word[32-len(b):], b)
			return word, nil
		}
		twosComplement := new(big.Int).Add(twoPow256, x)
		b := twosComplement.Bytes()
		copy(word[32-len(b):], b)
		copy(word[:32-len(b)], allFF[:32-len(b)])
		return word, nil
	}
}

// encodeDynamicBytes encodes a dynamic bytes or string value as its own
// self-contained tail block: enc(len) ++ pad_right(data).
func encodeDynamicBytes(t Type, v any) ([]byte, error) {
	switch t.Kind {
	case KindBytes:
		b, ok := v.([]byte)
		if !ok {
			return nil, fmt.Errorf("%w: expected []byte, got %T", ErrInvalidGoType, v)
		}
		return packBytes(b), nil
	case KindString:
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("%w: expected string, got %T", ErrInvalidGoType, v)
		}
		return packBytes([]byte(s)), nil
	default:
		return nil, fmt.Errorf("%w: %v is not a dynamic leaf type", ErrUnsupportedKind, t.Kind)
	}
}

// encodeInt64Word sign-extends x into a 32-byte two's complement word.
func encodeInt64Word(x int64) []byte {
	word := make([]byte, 32)
	if x < 0 {
		copy(word[:24], allFF[:24])
	}
	binary.BigEndian.PutUint64(word[24:], uint64(x))
	return word
}

// encodeUint64Word left-pads x into a 32-byte big-endian word.
func encodeUint64Word(x uint64) []byte {
	word := make([]byte, 32)
	binary.BigEndian.PutUint64(word[24:], x)
	return word
}

// packBytes encodes b as enc(len) ++ pad_right(b), padded to a multiple of 32 bytes.
func packBytes(b []byte) []byte {
	padded := (len(b) + 31) &^ 31
	out := make([]byte, 32+padded)
	copy(out, encodeUint64Word(uint64(len(b))))
	copy(out[32:], b)
	return out
}

// encodeArg encodes a single value of type t, dispatching to the matching
// encoder for scalars, dynamic leaves, or composite types (which recurse
// into packTuple).
func encodeArg(t Type, v any) ([]byte, error) {
	switch t.Kind {
	case KindBytes, KindString:
		return encodeDynamicBytes(t, v)
	case KindSlice:
		return encodeSlice(t, v)
	case KindArray:
		return encodeArray(t, v)
	case KindTuple:
		return encodeTupleValue(t, v)
	default:
		return encodeValue(t, v)
	}
}

// Pack encodes args according to types using the standard ABI head/tail
// layout, e.g. for a function call's arguments.
func Pack(types Types, args ...any) ([]byte, error) {
	return packTuple(types, args)
}

// packTuple encodes a sequence of typed values using the ABI head/tail
// layout: static values encode in place in the head, dynamic values leave
// a 32-byte offset in the head and their encoding in the tail. It backs
// Tuple, Array, and Slice encoding alike, since all three are "a typed
// sequence of values" at the wire level.
func packTuple(types Types, args []any) ([]byte, error) {
	if len(types) != len(args) {
		return nil, fmt.Errorf("%w: %d types but %d args", ErrArgCountMismatch, len(types), len(args))
	}

	heads := make([][]byte, len(types))
	tails := make([][]byte, len(types))
	headSize := 0

	for i, t := range types {
		if t.IsDynamic() {
			tail, err := encodeArg(t, args[i])
			if err != nil {
				return nil, err
			}
			tails[i] = tail
			headSize += 32
			continue
		}
		head, err := encodeArg(t, args[i])
		if err != nil {
			return nil, err
		}
		heads[i] = head
		headSize += len(head)
	}

	offset := headSize
	for i, t := range types {
		if !t.IsDynamic() {
			continue
		}
		heads[i] = encodeUint64Word(uint64(offset))
		offset += len(tails[i])
	}

	out := make([]byte, 0, offset)
	for _, h := range heads {
		out = append(out, h...)
	}
	for _, tl := range tails {
		out = append(out, tl...)
	}
	return out, nil
}

// encodeSlice encodes a T[] value as enc(count) ++ packTuple(Elem repeated
// count times, elements), accepting any Go slice via reflect.
func encodeSlice(t Type, v any) ([]byte, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return nil, fmt.Errorf("%w: expected a slice, got %T", ErrInvalidGoType, v)
	}

	n := rv.Len()
	elemTypes := make(Types, n)
	elemArgs := make([]any, n)
	for i := range n {
		elemTypes[i] = *t.Elem
		elemArgs[i] = rv.Index(i).Interface()
	}

	body, err := packTuple(elemTypes, elemArgs)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, 32+len(body))
	out = append(out, encodeUint64Word(uint64(n))...)
	out = append(out, body...)
	return out, nil
}

// encodeArray encodes a T[k] value as packTuple(Elem repeated k times,
// elements): a fixed array is a tuple with no length prefix. Accepts any
// Go slice or array via reflect.
func encodeArray(t Type, v any) ([]byte, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, fmt.Errorf("%w: expected a slice or array, got %T", ErrInvalidGoType, v)
	}
	if n := rv.Len(); n != t.Size {
		return nil, fmt.Errorf("%w: array expects %d elements, got %d", ErrArgCountMismatch, t.Size, n)
	}

	elemTypes := make(Types, t.Size)
	elemArgs := make([]any, t.Size)
	for i := 0; i < t.Size; i++ {
		elemTypes[i] = *t.Elem
		elemArgs[i] = rv.Index(i).Interface()
	}
	return packTuple(elemTypes, elemArgs)
}

// encodeTupleValue encodes a tuple value, given positionally as []any
// matching t.Components in order.
func encodeTupleValue(t Type, v any) ([]byte, error) {
	args, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("%w: expected []any, got %T", ErrInvalidGoType, v)
	}
	return packTuple(t.Components, args)
}

// decodeValue decodes a single static 32-byte ABI word into the same Go
// type encodeValue expects for t. It only handles Kinds that are always
// static and fit in one word: Bool, Uint, Int, Address, FixedBytes,
// FunctionType.
func decodeValue(t Type, word []byte) (any, error) {
	if len(word) != 32 {
		return nil, fmt.Errorf("%w: word must be 32 bytes, got %d", ErrByteLengthMismatch, len(word))
	}
	switch t.Kind {
	case KindBool:
		return decodeBool(word), nil
	case KindAddress:
		return decodeAddress(word), nil
	case KindFixedBytes:
		return decodeFixedBytes(t.Size, word), nil
	case KindFunction:
		return decodeFunction(word), nil
	case KindUint:
		return decodeUint(t.Size, word)
	case KindInt:
		return decodeInt(t.Size, word)
	default:
		return nil, fmt.Errorf("%w: %v is not a static scalar type", ErrUnsupportedKind, t.Kind)
	}
}

func decodeBool(word []byte) bool {
	return word[31] != 0
}

func decodeAddress(word []byte) *types.Address {
	return types.NewAddressFromBytes(word[32-types.AddressLength:])
}

func decodeFixedBytes(size int, word []byte) []byte {
	b := make([]byte, size)
	copy(b, word[:size])
	return b
}

func decodeFunction(word []byte) []byte {
	b := make([]byte, 24)
	copy(b, word[:24])
	return b
}

// decodeUint decodes a 32-byte word into an unsigned integer, using the
// same size<=64-native/size>64-*big.Int split as encodeUint.
func decodeUint(size int, word []byte) (any, error) {
	switch size {
	case 8:
		return word[31], nil
	case 16:
		return binary.BigEndian.Uint16(word[30:32]), nil
	case 32:
		return binary.BigEndian.Uint32(word[28:32]), nil
	case 64:
		return binary.BigEndian.Uint64(word[24:32]), nil
	default:
		bnd, ok := boundsFor(size)
		if !ok {
			return nil, fmt.Errorf("%w: uint%d is not a standard ABI width", ErrUnsupportedKind, size)
		}
		n := new(big.Int).SetBytes(word)
		if n.Cmp(bnd.uintMax) >= 0 {
			return nil, fmt.Errorf("%w: uint%d overflow: %s", ErrIntegerOutOfRange, size, n)
		}
		return n, nil
	}
}

// decodeInt decodes a 32-byte two's complement word into a signed integer,
// using the same size<=64-native/size>64-*big.Int split as encodeInt. An
// int64's bit pattern is already two's complement, so the low 8 bytes can
// be reinterpreted directly regardless of sign.
func decodeInt(size int, word []byte) (any, error) {
	switch size {
	case 8:
		return int8(word[31]), nil
	case 16:
		return int16(binary.BigEndian.Uint16(word[30:32])), nil
	case 32:
		return int32(binary.BigEndian.Uint32(word[28:32])), nil
	case 64:
		return int64(binary.BigEndian.Uint64(word[24:32])), nil
	default:
		bnd, ok := boundsFor(size)
		if !ok {
			return nil, fmt.Errorf("%w: int%d is not a standard ABI width", ErrUnsupportedKind, size)
		}
		n := new(big.Int).SetBytes(word)
		if word[0]&0x80 != 0 {
			n.Sub(n, twoPow256)
		}
		if n.Cmp(bnd.intMin) < 0 || n.Cmp(bnd.intMax) > 0 {
			return nil, fmt.Errorf("%w: int%d overflow: %s", ErrIntegerOutOfRange, size, n)
		}
		return n, nil
	}
}
