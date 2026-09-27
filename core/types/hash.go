package types

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// HashLength is the size of a 32-byte hash in bytes.
const HashLength = 32

// Hash is a 32-byte hash value.
type Hash struct {
	bytes [HashLength]byte
	hex   string
}

// NewHashFromBytes builds a Hash from raw bytes, normalizing to HashLength.
func NewHashFromBytes(b []byte) *Hash {
	h := new(Hash)
	h.SetBytes(b)
	return h
}

// NewHashFromHex parses a hex string (with or without a 0x prefix) into a Hash.
func NewHashFromHex(s string) (*Hash, error) {
	b, err := encoding.Hex.Decode(s)
	if err != nil {
		return nil, fmt.Errorf("types: invalid hash hex: %w", err)
	}
	if len(b) != HashLength {
		return nil, fmt.Errorf("types: invalid hash length: %d", len(b))
	}
	return NewHashFromBytes(b), nil
}

// NewHashFromBig builds a Hash from a big-endian integer.
func NewHashFromBig(n *big.Int) *Hash {
	return NewHashFromBytes(n.Bytes())
}

// IsHexHash reports whether s is a validly formatted hash (with or without a 0x prefix).
func IsHexHash(s string) bool {
	b, err := encoding.Hex.Decode(s)
	return err == nil && len(b) == HashLength
}

// SetBytes copies b into the hash, left-padding or right-truncating to HashLength.
func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h.bytes) {
		b = b[len(b)-HashLength:]
	}
	copy(h.bytes[HashLength-len(b):], b)
}

// Bytes returns a copy of the hash's raw bytes.
func (h *Hash) Bytes() []byte {
	b := make([]byte, HashLength)
	copy(b, h.bytes[:])
	return b
}

// IsZero reports whether the hash is nil or the all-zero hash.
func (h *Hash) IsZero() bool {
	return h == nil || h.bytes == [HashLength]byte{}
}

// Equal reports whether h and o are the same hash. It is nil-safe.
func (h *Hash) Equal(o *Hash) bool {
	if h == nil || o == nil {
		return h == o
	}
	return h.bytes == o.bytes
}

// Big returns the hash's bytes interpreted as a big-endian integer.
func (h *Hash) Big() *big.Int {
	return new(big.Int).SetBytes(h.bytes[:])
}

// String returns the lowercase hex encoding of the hash.
func (h *Hash) String() string {
	if h.hex == "" {
		h.hex = encoding.Hex.EncodePrefixed(h.bytes[:])
	}
	return h.hex
}
