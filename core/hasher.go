package core

import (
	"strconv"

	"github.com/ninja-protocol-labs/go-lib-cryptography/keccak"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

const (
	// eip191Prefix is the personal-sign prefix defined by EIP-191.
	eip191Prefix = "\x19Ethereum Signed Message:\n"

	// eip712Prefix is the two-byte version prefix defined by EIP-712.
	eip712Prefix = "\x19\x01"
)

// Keccak256 returns the raw Keccak-256 hash of data: the hash behind the
// KECCAK256 opcode, address derivation, and every ABI selector.
func Keccak256(data []byte) *types.Hash {
	digest := keccak.Hash256(data).Bytes()
	return types.NewHashFromBytes(digest[:])
}

// EIP191Hash returns the Keccak-256 hash of msg with the EIP-191 personal
// message prefix ("\x19Ethereum Signed Message:\n" + len(msg)) prepended,
// matching the digest wallets sign for eth_sign/personal_sign.
func EIP191Hash(msg []byte) *types.Hash {
	prefixed := make([]byte, 0, len(eip191Prefix)+20+len(msg))
	prefixed = append(prefixed, eip191Prefix...)
	prefixed = strconv.AppendInt(prefixed, int64(len(msg)), 10)
	prefixed = append(prefixed, msg...)
	return Keccak256(prefixed)
}

// EIP712Hash returns the final signable EIP-712 digest:
// keccak256("\x19\x01" ++ domainSeparator ++ hashStruct(message)).
func EIP712Hash(domain EIP712Domain, primaryType string, schema EIP712Types, message map[string]any) (*types.Hash, error) {
	domainSep, err := DomainSeparator(domain)
	if err != nil {
		return nil, err
	}
	msgHash, err := schema.HashStruct(primaryType, message)
	if err != nil {
		return nil, err
	}

	buf := append([]byte(eip712Prefix), domainSep.Bytes()...)
	buf = append(buf, msgHash.Bytes()...)
	return Keccak256(buf), nil
}

// EIP1014Address returns the deterministic contract address EIP-1014's
// CREATE2 opcode computes:
// keccak256(0xff ++ deployer ++ salt ++ keccak256(initCode))[12:].
func EIP1014Address(deployer *types.Address, salt [32]byte, initCode []byte) *types.Address {
	initCodeHash := Keccak256(initCode)

	buf := make([]byte, 0, 1+types.AddressLength+32+types.HashLength)
	buf = append(buf, 0xff)
	buf = append(buf, deployer.Bytes()...)
	buf = append(buf, salt[:]...)
	buf = append(buf, initCodeHash.Bytes()...)

	digest := Keccak256(buf)
	return types.NewAddressFromBytes(digest.Bytes()[12:])
}
