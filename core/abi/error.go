package abi

import (
	"fmt"
	"strings"

	"github.com/ninja-protocol-labs/go-lib-cryptography/keccak"
)

// errorSelector and panicSelector are the standard encodings Solidity
// uses for a require/revert reason and a compiler-inserted panic
// (overflow, array out-of-bounds, ...), respectively — computed rather
// than hardcoded so there's nothing to independently verify.
var (
	errorSelector = NewFunction("Error", Types{String}, nil).Selector()
	panicSelector = NewFunction("Panic", Types{Uint256}, nil).Selector()
)

// DecodeRevert decodes data — e.g. the Data field of a JSON-RPC
// "execution reverted" error — as a Solidity require/revert reason
// (Error(string)) or a compiler panic (Panic(uint256)), returning a
// human-readable message. It returns ErrSelectorMismatch if data isn't
// either standard encoding; for a Solidity custom error, parse it with
// ParseError and decode data with its Decode method instead.
func DecodeRevert(data []byte) (string, error) {
	if len(data) < 4 {
		return "", fmt.Errorf("%w: revert data must be at least 4 bytes, got %d", ErrByteLengthMismatch, len(data))
	}

	var sel Selector
	copy(sel[:], data[:4])

	switch sel {
	case errorSelector:
		vals, err := Unpack(Types{String}, data[4:])
		if err != nil {
			return "", fmt.Errorf("abi: decode revert reason: %w", err)
		}
		reason, _ := vals[0].(string)
		return reason, nil

	case panicSelector:
		vals, err := Unpack(Types{Uint256}, data[4:])
		if err != nil {
			return "", fmt.Errorf("abi: decode panic code: %w", err)
		}
		return fmt.Sprintf("panic code 0x%x", vals[0]), nil

	default:
		return "", fmt.Errorf("%w: %x is neither Error(string) nor Panic(uint256)", ErrSelectorMismatch, sel)
	}
}

// Error describes a Solidity custom error declaration (e.g. "error
// InsufficientBalance(uint256 available, uint256 required)"): a selector
// and argument list, the same shape as a function call but with no
// return values.
type Error struct {
	Name   string
	Inputs Types

	// set once by NewError and never written afterwards, so concurrent reads are safe
	s *Selector
}

// NewError builds an Error from its name and argument types.
func NewError(name string, inputs Types) *Error {
	e := &Error{
		Name:   name,
		Inputs: inputs,
	}
	sel := e.computeSelector()
	e.s = &sel
	return e
}

// ParseError parses a Solidity custom error signature, with or without a
// leading "error" keyword, e.g. "InsufficientBalance(uint256,uint256)" or
// "error InsufficientBalance(uint256 available, uint256 required)".
func ParseError(signature string) (*Error, error) {
	sig := strings.TrimSpace(signature)
	if rest, ok := stripPrefix(sig, "error"); ok && (rest == "" || !isIdentChar(rest[0])) {
		sig = strings.TrimSpace(rest)
	}

	idx := strings.IndexByte(sig, '(')
	if idx < 0 {
		return nil, fmt.Errorf("%w: missing '(' in %q", ErrInvalidTypeString, signature)
	}
	name := strings.TrimSpace(sig[:idx])
	if name == "" {
		return nil, fmt.Errorf("%w: missing error name in %q", ErrInvalidTypeString, signature)
	}

	inputs, pos, err := parseParamList(sig, idx)
	if err != nil {
		return nil, err
	}
	if pos = skipSpace(sig, pos); pos != len(sig) {
		return nil, fmt.Errorf("%w: unexpected trailing characters in %q", ErrInvalidTypeString, signature)
	}

	return NewError(name, inputs), nil
}

// Signature returns the error's canonical signature, e.g.
// "InsufficientBalance(uint256,uint256)".
func (e *Error) Signature() string {
	names := make([]string, len(e.Inputs))
	for i := range e.Inputs {
		names[i] = e.Inputs[i].String()
	}
	return e.Name + "(" + strings.Join(names, ",") + ")"
}

// Selector returns the 4-byte ABI selector: the first 4 bytes of
// keccak256(Signature()).
func (e *Error) Selector() Selector {
	if e.s != nil {
		return *e.s
	}
	return e.computeSelector()
}

// Decode decodes a revert payload (e.g. the Data field of a JSON-RPC
// "execution reverted" error): its leading 4-byte selector must match
// e.Selector(), and the remaining bytes are decoded according to Inputs.
func (e *Error) Decode(data []byte) ([]any, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("%w: revert data must be at least 4 bytes, got %d", ErrByteLengthMismatch, len(data))
	}
	var got Selector
	copy(got[:], data[:4])
	if want := e.Selector(); got != want {
		return nil, fmt.Errorf("%w: expected %x, got %x", ErrSelectorMismatch, want, got)
	}

	vals, err := Unpack(e.Inputs, data[4:])
	if err != nil {
		return nil, fmt.Errorf("abi: decode error %s: %w", e.Name, err)
	}
	return vals, nil
}

func (e *Error) computeSelector() Selector {
	digest := keccak.Hash256([]byte(e.Signature())).Bytes()
	var sel Selector
	copy(sel[:], digest[:4])
	return sel
}
