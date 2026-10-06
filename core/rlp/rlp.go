// Package rlp wraps go-ethereum's RLP codec. It is the only package allowed to
// import go-ethereum.
package rlp

import (
	"fmt"
	"math/big"

	grpl "github.com/ethereum/go-ethereum/rlp"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Encode returns the RLP encoding of v.
func Encode(v any) ([]byte, error) {
	b, err := grpl.EncodeToBytes(v)
	if err != nil {
		return nil, fmt.Errorf("rlp: encode: %w", err)
	}
	return b, nil
}

// Decode parses data into v. Non-canonical encodings and trailing bytes are errors.
func Decode(data []byte, v any) error {
	if err := grpl.DecodeBytes(data, v); err != nil {
		return fmt.Errorf("rlp: decode: %w", err)
	}
	return nil
}

// FromAddress converts a to a fixed-size array; nil maps to nil.
func FromAddress(a *types.Address) *[types.AddressLength]byte {
	if a == nil {
		return nil
	}

	var out [types.AddressLength]byte
	copy(out[:], a.Bytes())
	return &out
}

// ToAddress is the inverse of FromAddress.
func ToAddress(a *[types.AddressLength]byte) *types.Address {
	if a == nil {
		return nil
	}
	return types.NewAddressFromBytes(a[:])
}

// BigOrZero returns n, or a new zero value when n is nil.
func BigOrZero(n *big.Int) *big.Int {
	if n == nil {
		return new(big.Int)
	}
	return n
}

// SignatureFromRS builds a signature from r, s and a recovery id.
func SignatureFromRS(r, s *big.Int, recoveryID byte) (*types.Signature, error) {
	if r.BitLen() > 256 || s.BitLen() > 256 {
		return nil, fmt.Errorf("rlp: signature value exceeds 256 bits")
	}
	buf := make([]byte, types.SignatureLength)
	r.FillBytes(buf[:32])
	s.FillBytes(buf[32:64])
	buf[64] = recoveryID
	return types.NewSignatureFromBytes(buf)
}

// RecoveryID returns sig's recovery id, rejecting values above 1.
func RecoveryID(sig *types.Signature) (byte, error) {
	v := sig.V()
	if v.BitLen() > 1 {
		return 0, fmt.Errorf("rlp: unsupported recovery id %s", v)
	}
	return byte(v.Uint64()), nil
}
