package abi

import (
	"encoding/binary"
	"fmt"
	"math/big"

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
		b, ok := v.(bool)
		if !ok {
			return nil, fmt.Errorf("%w: expected bool, got %T", ErrInvalidGoType, v)
		}
		word := make([]byte, 32)
		if b {
			word[31] = 1
		}
		return word, nil

	case KindAddress:
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

	case KindFixedBytes:
		// bytesN is right-padded, unlike the other static types.
		b, ok := v.([]byte)
		if !ok {
			return nil, fmt.Errorf("%w: expected []byte, got %T", ErrInvalidGoType, v)
		}
		if len(b) != t.Size {
			return nil, fmt.Errorf("%w: bytes%d expects %d bytes, got %d", ErrByteLengthMismatch, t.Size, t.Size, len(b))
		}
		word := make([]byte, 32)
		copy(word, b)
		return word, nil

	case KindFunction:
		// 20-byte address + 4-byte selector, right-padded like bytes24.
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

	case KindUint:
		return encodeUint(t.Size, v)
	case KindInt:
		return encodeInt(t.Size, v)

	default:
		return nil, fmt.Errorf("%w: %v is not a static scalar type", ErrUnsupportedKind, t.Kind)
	}
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
