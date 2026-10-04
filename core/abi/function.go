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

	// cache
	s *Selector
}

// NewFunction builds a Function from its name, input types, and output types.
func NewFunction(name string, inputs, outputs Types) *Function {
	return &Function{
		Name:    name,
		Inputs:  inputs,
		Outputs: outputs,
	}
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

	digest := keccak.Hash256([]byte(f.Signature())).Bytes()
	var sel Selector
	copy(sel[:], digest[:4])
	f.s = &sel
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
