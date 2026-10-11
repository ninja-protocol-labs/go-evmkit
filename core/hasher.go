package core

import (
	"github.com/ninja-protocol-labs/go-lib-cryptography/keccak"

	"github.com/ninja-protocol-labs/go-evmkit/core/rlp"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Keccak256 returns the raw Keccak-256 hash of data: the hash behind the
// KECCAK256 opcode, address derivation, and every ABI selector.
func Keccak256(data []byte) *types.Hash {
	digest := keccak.Hash256(data).Bytes()
	return types.NewHashFromBytes(digest[:])
}

// CreateAddress returns the deterministic contract address the CREATE
// opcode computes for a deployer at a given nonce:
// keccak256(rlp([sender, nonce]))[12:].
func CreateAddress(sender *types.Address, nonce uint64) (*types.Address, error) {
	enc, err := rlp.Encode([]any{sender.Bytes(), nonce})
	if err != nil {
		return nil, err
	}

	digest := Keccak256(enc)
	return types.NewAddressFromBytes(digest.Bytes()[12:]), nil
}
