package types

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
	"github.com/ninja-protocol-labs/go-lib-cryptography/secp256k1"
)

const (
	// SignatureLength is the byte length of a recoverable (r ∥ s ∥ v) signature.
	SignatureLength = secp256k1.SignatureCompactLen + 1

	// CompactSignatureLength is the byte length of an EIP-2098 compact (r ∥ yParityAndS) signature.
	CompactSignatureLength = secp256k1.SignatureCompactLen
)

// Signature is a recoverable ECDSA signature: r, s and a recovery id.
type Signature struct {
	sig *secp256k1.Signature

	// cache
	bytes []byte
	r     *big.Int
	s     *big.Int
	v     byte
	hex   string
}

// NewSignatureFromBytes parses a recoverable (r ∥ s ∥ v) signature.
func NewSignatureFromBytes(b []byte) (*Signature, error) {
	if len(b) != SignatureLength {
		return nil, fmt.Errorf("types: invalid signature length: %d", len(b))
	}

	sig, err := secp256k1.SignatureFromBytes(b[:secp256k1.SignatureCompactLen])
	if err != nil {
		return nil, fmt.Errorf("types: invalid signature: %w", err)
	}
	return &Signature{
		sig: sig,
		v:   b[secp256k1.SignatureCompactLen],
	}, nil
}

// NewSignatureFromHex parses a hex-encoded recoverable signature (with or without a 0x prefix).
func NewSignatureFromHex(s string) (*Signature, error) {
	b, err := encoding.Hex.Decode(s)
	if err != nil {
		return nil, fmt.Errorf("types: invalid signature hex: %w", err)
	}
	return NewSignatureFromBytes(b)
}

// NewSignatureFromBig builds a Signature from its recoverable (r ∥ s ∥ v) encoding interpreted as a big-endian integer.
func NewSignatureFromBig(n *big.Int) (*Signature, error) {
	b := n.Bytes()
	if len(b) > SignatureLength {
		return nil, fmt.Errorf("types: invalid signature: too large")
	}

	buf := make([]byte, SignatureLength)
	copy(buf[SignatureLength-len(b):], b)
	return NewSignatureFromBytes(buf)
}

// Bytes returns the recoverable encoding, r ∥ s ∥ v, as a copy.
func (s *Signature) Bytes() []byte {
	if s.bytes == nil {
		rs := s.sig.Bytes()

		b := make([]byte, SignatureLength)
		copy(b, rs[:])
		b[SignatureLength-1] = s.v
		s.bytes = b
	}

	b := make([]byte, SignatureLength)
	copy(b, s.bytes)
	return b
}

// Big returns the recoverable encoding, r ∥ s ∥ v, interpreted as a big-endian integer.
func (s *Signature) Big() *big.Int {
	return new(big.Int).SetBytes(s.Bytes())
}

// R returns the signature's r value.
func (s *Signature) R() *big.Int {
	if s.r == nil {
		s.r = new(big.Int).SetBytes(s.Bytes()[:secp256k1.SignatureScalarLen])
	}
	return new(big.Int).Set(s.r)
}

// S returns the signature's s value.
func (s *Signature) S() *big.Int {
	if s.s == nil {
		s.s = new(big.Int).SetBytes(s.Bytes()[secp256k1.SignatureScalarLen:secp256k1.SignatureCompactLen])
	}
	return new(big.Int).Set(s.s)
}

// V returns the raw recovery id (0-3).
func (s *Signature) V() *big.Int {
	return new(big.Int).SetUint64(uint64(s.v))
}

// LegacyV returns the recovery id encoded as a pre-EIP-155 legacy transaction v value.
func (s *Signature) LegacyV() *big.Int {
	return new(big.Int).Add(s.V(), big.NewInt(27))
}

// EIP155V returns the recovery id encoded as an EIP-155 transaction v value for the given chain id.
func (s *Signature) EIP155V(chainID *big.Int) *big.Int {
	v := new(big.Int).Mul(chainID, big.NewInt(2))
	v.Add(v, big.NewInt(35))
	v.Add(v, s.V())
	return v
}

// CompactBytes returns the EIP-2098 compact encoding: r ∥ s, with the top
// bit of s set to the recovery id. This is lossless only for a recovery id
// of 0 or 1, which holds for every signature secp256k1.Sign produces.
func (s *Signature) CompactBytes() ([]byte, error) {
	if s.v > 1 {
		return nil, fmt.Errorf("types: compact signature: unsupported recovery id %d", s.v)
	}

	b := s.Bytes()
	compact := make([]byte, CompactSignatureLength)
	copy(compact, b[:CompactSignatureLength])
	if s.v == 1 {
		compact[secp256k1.SignatureScalarLen] |= 0x80
	}
	return compact, nil
}

// NewSignatureFromCompact parses an EIP-2098 compact (r ∥ yParityAndS) signature.
func NewSignatureFromCompact(b []byte) (*Signature, error) {
	if len(b) != CompactSignatureLength {
		return nil, fmt.Errorf("types: invalid compact signature length: %d", len(b))
	}

	ss := make([]byte, secp256k1.SignatureScalarLen)
	copy(ss, b[secp256k1.SignatureScalarLen:])

	var v byte
	if ss[0]&0x80 != 0 {
		v = 1
		ss[0] &^= 0x80
	}

	full := make([]byte, SignatureLength)
	copy(full[:secp256k1.SignatureScalarLen], b[:secp256k1.SignatureScalarLen])
	copy(full[secp256k1.SignatureScalarLen:secp256k1.SignatureCompactLen], ss)
	full[SignatureLength-1] = v
	return NewSignatureFromBytes(full)
}

// Equal reports whether sig and o hold the same r, s and recovery id. It is nil-safe.
func (s *Signature) Equal(o *Signature) bool {
	if s == nil || o == nil {
		return s == o
	}
	return s.sig.Equal(o.sig) && s.v == o.v
}

// IsZero reports whether the signature is nil or unset.
func (s *Signature) IsZero() bool {
	return s == nil || s.sig.IsZero()
}

// String returns the recoverable encoding as lowercase hex.
func (s *Signature) String() string {
	if s.hex == "" {
		s.hex = encoding.Hex.EncodePrefixed(s.Bytes())
	}
	return s.hex
}

// ECRecover returns the public key that produced sig over digest.
func (s *Signature) ECRecover(digest *Hash) (*PublicKey, error) {
	if s == nil {
		return nil, fmt.Errorf("types: ecrecover: nil signature")
	}

	k, err := secp256k1.Recover(digest.Bytes(), s.sig, s.v)
	if err != nil {
		return nil, fmt.Errorf("types: ecrecover: %w", err)
	}
	return &PublicKey{
		key: k,
	}, nil
}
