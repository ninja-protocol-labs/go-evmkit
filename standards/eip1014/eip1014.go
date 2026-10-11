// Package eip1014 computes the deterministic contract address EIP-1014's
// CREATE2 opcode produces.
package eip1014

import (
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Address returns the deterministic contract address EIP-1014's CREATE2
// opcode computes:
// keccak256(0xff ++ deployer ++ salt ++ keccak256(initCode))[12:].
func Address(deployer *types.Address, salt [32]byte, initCode []byte) *types.Address {
	initCodeHash := core.Keccak256(initCode)

	buf := make([]byte, 0, 1+types.AddressLength+32+types.HashLength)
	buf = append(buf, 0xff)
	buf = append(buf, deployer.Bytes()...)
	buf = append(buf, salt[:]...)
	buf = append(buf, initCodeHash.Bytes()...)

	digest := core.Keccak256(buf)
	return types.NewAddressFromBytes(digest.Bytes()[12:])
}
