package types

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
	"github.com/ninja-protocol-labs/go-lib-cryptography/secp256k1"
)

// PrivateKeyLength is the byte length of a secp256k1 private key.
const PrivateKeyLength = secp256k1.SeckeyLen

// PrivateKey is a secp256k1 private key.
type PrivateKey struct {
	key *secp256k1.PrivateKey
}

// GeneratePrivateKey generates a new random private key.
func GeneratePrivateKey() (*PrivateKey, error) {
	k, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("types: generate private key: %w", err)
	}
	return &PrivateKey{
		key: k,
	}, nil
}

// NewPrivateKeyFromBytes parses a 32-byte big-endian scalar into a PrivateKey.
func NewPrivateKeyFromBytes(b []byte) (*PrivateKey, error) {
	k, err := secp256k1.PrivateKeyFromBytes(b)
	if err != nil {
		return nil, fmt.Errorf("types: invalid private key: %w", err)
	}
	return &PrivateKey{
		key: k,
	}, nil
}

// NewPrivateKeyFromHex parses a hex string (with or without a 0x prefix) into a PrivateKey.
func NewPrivateKeyFromHex(s string) (*PrivateKey, error) {
	b, err := encoding.Hex.Decode(s)
	if err != nil {
		return nil, fmt.Errorf("types: invalid private key hex: %w", err)
	}
	return NewPrivateKeyFromBytes(b)
}

// NewPrivateKeyFromBig builds a PrivateKey from a big-endian integer scalar.
func NewPrivateKeyFromBig(n *big.Int) (*PrivateKey, error) {
	b := n.Bytes()
	if len(b) > PrivateKeyLength {
		return nil, fmt.Errorf("types: invalid private key: too large")
	}

	buf := make([]byte, PrivateKeyLength)
	copy(buf[PrivateKeyLength-len(b):], b)
	return NewPrivateKeyFromBytes(buf)
}

// PublicKey returns the public key this private key derives to.
func (k *PrivateKey) PublicKey() *PublicKey {
	return &PublicKey{
		key: k.key.PublicKey(),
	}
}

// Bytes returns the private key's scalar, big-endian, as a copy. It is secret.
func (k *PrivateKey) Bytes() []byte {
	b := k.key.Bytes()
	return b[:]
}

// Big returns the private key's scalar interpreted as a big-endian integer. It is secret.
func (k *PrivateKey) Big() *big.Int {
	return new(big.Int).SetBytes(k.Bytes())
}

// Equal reports whether k and o are the same private key. It is nil-safe and constant-time.
func (k *PrivateKey) Equal(o *PrivateKey) bool {
	if k == nil || o == nil {
		return k == o
	}
	return k.key.Equal(o.key)
}

// IsZero reports whether the private key is nil or unset.
func (k *PrivateKey) IsZero() bool {
	return k == nil || k.key.IsZero()
}

// Sign returns a recoverable, deterministic signature over digest, with s normalized low.
func (k *PrivateKey) Sign(digest *Hash) (*Signature, error) {
	sig, v, err := secp256k1.SignRecoverable(k.key, digest.Bytes())
	if err != nil {
		return nil, fmt.Errorf("types: sign: %w", err)
	}
	return &Signature{
		sig: sig,
		v:   v,
	}, nil
}
