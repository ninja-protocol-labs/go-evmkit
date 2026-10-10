package abi

import (
	"fmt"
	"strings"

	"github.com/ninja-protocol-labs/go-lib-cryptography/keccak"
)

// Function describes a Solidity contract function by its input and output
// types, so calldata can be built and returned values decoded without
// parsing a signature string.
type Function struct {
	Name    string
	Inputs  Types
	Outputs Types

	// set once by NewFunction and never written afterwards, so concurrent reads are safe
	s *Selector
}

// NewFunction builds a Function from its name, input types, and output types.
func NewFunction(name string, inputs, outputs Types) *Function {
	f := &Function{
		Name:    name,
		Inputs:  inputs,
		Outputs: outputs,
	}
	sel := f.computeSelector()
	f.s = &sel
	return f
}

// Signature returns the function's canonical signature, e.g.
// "transfer(address,uint256)".
func (f *Function) Signature() string {
	names := make([]string, len(f.Inputs))
	for i := range f.Inputs {
		names[i] = f.Inputs[i].String()
	}
	return f.Name + "(" + strings.Join(names, ",") + ")"
}

// Selector returns the 4-byte ABI selector: the first 4 bytes of
// keccak256(Signature()).
func (f *Function) Selector() Selector {
	if f.s != nil {
		return *f.s
	}
	return f.computeSelector()
}

// SelectorBytes returns the 4-byte ABI selector as a []byte, i.e. the
// calldata for a call with no arguments.
func (f *Function) SelectorBytes() []byte {
	sel := f.Selector()
	return sel[:]
}

func (f *Function) computeSelector() Selector {
	digest := keccak.Hash256([]byte(f.Signature())).Bytes()
	var sel Selector
	copy(sel[:], digest[:4])
	return sel
}

// EncodeCall encodes a full function call: the 4-byte selector followed by
// args ABI-encoded according to Inputs.
func (f *Function) EncodeCall(args ...any) ([]byte, error) {
	packed, err := Pack(f.Inputs, args...)
	if err != nil {
		return nil, fmt.Errorf("abi: encode call to %s: %w", f.Name, err)
	}
	sel := f.Selector()
	return append(sel[:], packed...), nil
}

// EncodeReturn encodes a function's return values according to Outputs,
// e.g. to build synthetic eth_call return data for mocking or simulation
// — the inverse of DecodeReturn.
func (f *Function) EncodeReturn(args ...any) ([]byte, error) {
	packed, err := Pack(f.Outputs, args...)
	if err != nil {
		return nil, fmt.Errorf("abi: encode return from %s: %w", f.Name, err)
	}
	return packed, nil
}

// DecodeCall decodes a full function call (as produced by EncodeCall):
// its leading 4-byte selector must match f.Selector(), and the remaining
// bytes are decoded according to Inputs.
func (f *Function) DecodeCall(calldata []byte) ([]any, error) {
	if len(calldata) < 4 {
		return nil, fmt.Errorf("%w: calldata must be at least 4 bytes, got %d", ErrByteLengthMismatch, len(calldata))
	}
	var got Selector
	copy(got[:], calldata[:4])
	if want := f.Selector(); got != want {
		return nil, fmt.Errorf("%w: expected %x, got %x", ErrSelectorMismatch, want, got)
	}

	vals, err := Unpack(f.Inputs, calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("abi: decode call to %s: %w", f.Name, err)
	}
	return vals, nil
}

// DecodeReturn decodes a function's return data according to Outputs.
func (f *Function) DecodeReturn(data []byte) ([]any, error) {
	vals, err := Unpack(f.Outputs, data)
	if err != nil {
		return nil, fmt.Errorf("abi: decode return from %s: %w", f.Name, err)
	}
	return vals, nil
}

// DecodeSingleReturn decodes data as f's return values, expecting exactly
// one (e.g. a single-value getter like balanceOf), and type-asserts it as T.
func (f *Function) DecodeSingleReturn[T any](data []byte) (T, error) {
	var z T
	vals, err := f.DecodeReturn(data)
	if err != nil {
		return z, err
	}
	if len(vals) == 0 {
		return z, fmt.Errorf("%w: %s has no return values", ErrArgCountMismatch, f.Name)
	}
	v, ok := vals[0].(T)
	if !ok {
		return z, fmt.Errorf("%w: expected %T, got %T", ErrInvalidGoType, z, vals[0])
	}
	return v, nil
}

// DecodeSingleReturnAs is DecodeSingleReturn followed by convert, for a
// single return value that needs wrapping into a domain type (e.g. a raw
// []byte into a *types.Hash). T and R are both inferred from convert.
func (f *Function) DecodeSingleReturnAs[T, R any](data []byte, convert func(T) R) (R, error) {
	var zero R
	v, err := f.DecodeSingleReturn[T](data)
	if err != nil {
		return zero, err
	}
	return convert(v), nil
}
