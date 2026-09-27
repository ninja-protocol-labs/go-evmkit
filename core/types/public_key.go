package types

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
	"github.com/ninja-protocol-labs/go-lib-cryptography/secp256k1"
)

// PublicKeyCompressedLength and PublicKeyUncompressedLength are the two SEC 1 point encodings.
const (
	PublicKeyCompressedLength   = secp256k1.PubkeyCompressedLen
	PublicKeyUncompressedLength = secp256k1.PubkeyUncompressedLen
)

// PublicKey is a point on secp256k1.
type PublicKey struct {
	key *secp256k1.PublicKey
}

// NewPublicKeyFromBytes parses a compressed or uncompressed SEC 1 point into a PublicKey.
func NewPublicKeyFromBytes(b []byte) (*PublicKey, error) {
	k, err := secp256k1.PublicKeyFromBytes(b)
	if err != nil {
		return nil, fmt.Errorf("types: invalid public key: %w", err)
	}
	return &PublicKey{
		key: k,
	}, nil
}

// NewPublicKeyFromHex parses a hex-encoded SEC 1 point (with or without a 0x prefix) into a PublicKey.
func NewPublicKeyFromHex(s string) (*PublicKey, error) {
	b, err := encoding.Hex.Decode(s)
	if err != nil {
		return nil, fmt.Errorf("types: invalid public key hex: %w", err)
	}
	return NewPublicKeyFromBytes(b)
}

// NewPublicKeyFromBig builds a PublicKey from its uncompressed SEC 1 encoding interpreted as a big-endian integer.
func NewPublicKeyFromBig(n *big.Int) (*PublicKey, error) {
	b := n.Bytes()
	if len(b) > PublicKeyUncompressedLength {
		return nil, fmt.Errorf("types: invalid public key: too large")
	}

	buf := make([]byte, PublicKeyUncompressedLength)
	copy(buf[PublicKeyUncompressedLength-len(b):], b)
	return NewPublicKeyFromBytes(buf)
}

// Bytes returns the compressed SEC 1 encoding: a parity byte and X.
func (k *PublicKey) Bytes() []byte {
	b := k.key.Bytes()
	return b[:]
}

// BytesUncompressed returns the uncompressed SEC 1 encoding: 0x04, X and Y.
func (k *PublicKey) BytesUncompressed() []byte {
	b := k.key.BytesUncompressed()
	return b[:]
}

// Big returns the uncompressed SEC 1 encoding interpreted as a big-endian integer.
func (k *PublicKey) Big() *big.Int {
	return new(big.Int).SetBytes(k.BytesUncompressed())
}

// Equal reports whether k and o are the same point. It is nil-safe.
func (k *PublicKey) Equal(o *PublicKey) bool {
	if k == nil || o == nil {
		return k == o
	}
	return k.key.Equal(o.key)
}

// IsZero reports whether the public key is nil or unset.
func (k *PublicKey) IsZero() bool {
	return k == nil || k.key.IsZero()
}

// String returns the compressed encoding as lowercase hex.
func (k *PublicKey) String() string {
	return k.key.String()
}

// Verify reports whether sig is k's signature over digest. High-s signatures are rejected.
func (k *PublicKey) Verify(digest *Hash, sig *Signature) bool {
	if k == nil || sig == nil {
		return false
	}
	return secp256k1.Verify(k.key, digest.Bytes(), sig.sig)
}
