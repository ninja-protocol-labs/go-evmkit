package abi

import (
	"errors"
	"fmt"
	"strings"
)

// Sentinel errors returned across the abi package: by the Type
// constructors, by encoding/decoding, and by the signature-string parser.
var (
	ErrInvalidArraySize   = errors.New("abi: array size must be non-negative")
	ErrTupleNameMismatch  = errors.New("abi: tuple names and components length mismatch")
	ErrUnsupportedKind    = errors.New("abi: unsupported kind")
	ErrInvalidGoType      = errors.New("abi: invalid go type for abi type")
	ErrNilAddress         = errors.New("abi: address is nil")
	ErrByteLengthMismatch = errors.New("abi: byte length mismatch")
	ErrIntegerOutOfRange  = errors.New("abi: integer out of range")
	ErrArgCountMismatch   = errors.New("abi: argument count mismatch")
	ErrInvalidTypeString  = errors.New("abi: invalid type string")
	ErrSelectorMismatch   = errors.New("abi: selector mismatch")
)

// Kind identifies which Solidity ABI type an Type represents.
type Kind int

const (
	KindBool       Kind = iota // bool
	KindUint                   // uintN, 8-256 bits
	KindInt                    // intN, 8-256 bits
	KindAddress                // address
	KindBytes                  // dynamic bytes
	KindFixedBytes             // bytesN
	KindString                 // dynamic string
	KindArray                  // T[k], fixed length
	KindSlice                  // T[], dynamic length
	KindTuple                  // (T1, T2, ...)
	KindFunction               // address (20 bytes) + selector (4 bytes), encoded as bytes24
)

// Type describes a Solidity ABI type, recursively for arrays and tuples.
type Type struct {
	Kind Kind

	// Size is the bit width for Uint/Int (8-256, multiples of 8), the byte
	// length for FixedBytes (1-32), or the element count for Array.
	Size int

	// Elem is the element type for Array and Slice. Nil otherwise.
	Elem *Type

	// Components is the field types for Tuple, in order. Nil otherwise.
	Components Types

	// Names holds Tuple field names parallel to Components. Empty when the
	// tuple's fields are unnamed.
	Names []string
}

// Types is a list of Type, e.g. a function's parameter types.
type Types []Type

// Selector is a 4-byte ABI function selector: keccak256(signature)[:4].
type Selector [4]byte

// IsDynamic reports whether values of this type require head/tail encoding
// (bytes, string, T[], or anything containing one of those).
func (t *Type) IsDynamic() bool {
	switch t.Kind {
	case KindBytes, KindString, KindSlice:
		return true
	case KindArray:
		return t.Elem.IsDynamic()
	case KindTuple:
		for i := range t.Components {
			if t.Components[i].IsDynamic() {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// String returns t's canonical ABI type name, e.g. "uint256", "address[]",
// or "(address,uint256)" for a tuple.
func (t *Type) String() string {
	switch t.Kind {
	case KindBool:
		return "bool"
	case KindUint:
		return fmt.Sprintf("uint%d", t.Size)
	case KindInt:
		return fmt.Sprintf("int%d", t.Size)
	case KindAddress:
		return "address"
	case KindBytes:
		return "bytes"
	case KindFixedBytes:
		return fmt.Sprintf("bytes%d", t.Size)
	case KindString:
		return "string"
	case KindSlice:
		return t.Elem.String() + "[]"
	case KindArray:
		return fmt.Sprintf("%s[%d]", t.Elem.String(), t.Size)
	case KindTuple:
		names := make([]string, len(t.Components))
		for i := range t.Components {
			names[i] = t.Components[i].String()
		}
		return "(" + strings.Join(names, ",") + ")"
	case KindFunction:
		return "function"
	default:
		return "unknown"
	}
}

// NewTypes builds a Types list from individual Type values.
func NewTypes(components ...Type) Types {
	return components
}

// Canonical scalar Type values, so common types don't need to be
// constructed by hand.
var (
	Bool         = Type{Kind: KindBool}
	Address      = Type{Kind: KindAddress}
	String       = Type{Kind: KindString}
	Bytes        = Type{Kind: KindBytes}
	FunctionType = Type{Kind: KindFunction}

	Uint8   = Type{Kind: KindUint, Size: 8}
	Uint16  = Type{Kind: KindUint, Size: 16}
	Uint32  = Type{Kind: KindUint, Size: 32}
	Uint64  = Type{Kind: KindUint, Size: 64}
	Uint128 = Type{Kind: KindUint, Size: 128}
	Uint256 = Type{Kind: KindUint, Size: 256}

	Int8   = Type{Kind: KindInt, Size: 8}
	Int16  = Type{Kind: KindInt, Size: 16}
	Int32  = Type{Kind: KindInt, Size: 32}
	Int64  = Type{Kind: KindInt, Size: 64}
	Int128 = Type{Kind: KindInt, Size: 128}
	Int256 = Type{Kind: KindInt, Size: 256}

	Bytes1  = Type{Kind: KindFixedBytes, Size: 1}
	Bytes2  = Type{Kind: KindFixedBytes, Size: 2}
	Bytes3  = Type{Kind: KindFixedBytes, Size: 3}
	Bytes4  = Type{Kind: KindFixedBytes, Size: 4}
	Bytes5  = Type{Kind: KindFixedBytes, Size: 5}
	Bytes6  = Type{Kind: KindFixedBytes, Size: 6}
	Bytes7  = Type{Kind: KindFixedBytes, Size: 7}
	Bytes8  = Type{Kind: KindFixedBytes, Size: 8}
	Bytes9  = Type{Kind: KindFixedBytes, Size: 9}
	Bytes10 = Type{Kind: KindFixedBytes, Size: 10}
	Bytes11 = Type{Kind: KindFixedBytes, Size: 11}
	Bytes12 = Type{Kind: KindFixedBytes, Size: 12}
	Bytes13 = Type{Kind: KindFixedBytes, Size: 13}
	Bytes14 = Type{Kind: KindFixedBytes, Size: 14}
	Bytes15 = Type{Kind: KindFixedBytes, Size: 15}
	Bytes16 = Type{Kind: KindFixedBytes, Size: 16}
	Bytes17 = Type{Kind: KindFixedBytes, Size: 17}
	Bytes18 = Type{Kind: KindFixedBytes, Size: 18}
	Bytes19 = Type{Kind: KindFixedBytes, Size: 19}
	Bytes20 = Type{Kind: KindFixedBytes, Size: 20}
	Bytes21 = Type{Kind: KindFixedBytes, Size: 21}
	Bytes22 = Type{Kind: KindFixedBytes, Size: 22}
	Bytes23 = Type{Kind: KindFixedBytes, Size: 23}
	Bytes24 = Type{Kind: KindFixedBytes, Size: 24}
	Bytes25 = Type{Kind: KindFixedBytes, Size: 25}
	Bytes26 = Type{Kind: KindFixedBytes, Size: 26}
	Bytes27 = Type{Kind: KindFixedBytes, Size: 27}
	Bytes28 = Type{Kind: KindFixedBytes, Size: 28}
	Bytes29 = Type{Kind: KindFixedBytes, Size: 29}
	Bytes30 = Type{Kind: KindFixedBytes, Size: 30}
	Bytes31 = Type{Kind: KindFixedBytes, Size: 31}
	Bytes32 = Type{Kind: KindFixedBytes, Size: 32}
)

// Slice returns the dynamic-length array type T[] for elem.
func Slice(elem Type) Type {
	return Type{Kind: KindSlice, Elem: &elem}
}

// Array returns the fixed-length array type T[size] for elem. size must be
// non-negative.
func Array(elem Type, size int) (Type, error) {
	if size < 0 {
		return Type{}, fmt.Errorf("%w: got %d", ErrInvalidArraySize, size)
	}
	return Type{Kind: KindArray, Elem: &elem, Size: size}, nil
}

// Tuple returns the unnamed tuple type (T1, T2, ...) for components.
func Tuple(components ...Type) Type {
	return Type{Kind: KindTuple, Components: components}
}

// NamedTuple returns the tuple type (T1, T2, ...) for components, with
// field names parallel to components (e.g. a struct's members). len(names)
// must equal len(components).
func NamedTuple(names []string, components ...Type) (Type, error) {
	if len(names) != len(components) {
		return Type{}, fmt.Errorf("%w: %d names but %d components", ErrTupleNameMismatch, len(names), len(components))
	}
	return Type{Kind: KindTuple, Components: components, Names: names}, nil
}
