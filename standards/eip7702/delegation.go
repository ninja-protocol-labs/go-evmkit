package eip7702

import (
	"bytes"
	"fmt"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

var delegationPrefix = [3]byte{0xef, 0x01, 0x00}

// DelegationLength is the size of a delegation designator: the 0xef0100 prefix plus an address.
const DelegationLength = len(delegationPrefix) + types.AddressLength

// EncodeDelegation returns the delegation designator 0xef0100 || delegate.
func EncodeDelegation(delegate *types.Address) ([]byte, error) {
	if delegate.IsZero() {
		return nil, fmt.Errorf("eip7702: encodeDelegation: %w", ErrInvalidDelegate)
	}

	code := make([]byte, 0, DelegationLength)
	code = append(code, delegationPrefix[:]...)
	return append(code, delegate.Bytes()...), nil
}

// ParseDelegation returns the delegate when code is exactly a delegation designator.
func ParseDelegation(code []byte) (*types.Address, bool) {
	if len(code) != DelegationLength || !bytes.HasPrefix(code, delegationPrefix[:]) {
		return nil, false
	}

	return types.NewAddressFromBytes(code[len(delegationPrefix):]), true
}
