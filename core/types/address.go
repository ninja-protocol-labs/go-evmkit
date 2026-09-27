package types

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
	"github.com/ninja-protocol-labs/go-lib-cryptography/keccak"
)

// AddressLength is the size of an Ethereum address in bytes.
const AddressLength = 20

// Address is a 20-byte Ethereum account or contract address.
type Address struct {
	bytes [AddressLength]byte
	hex   string
}

// NewAddressFromBytes builds an Address from raw bytes, normalizing to AddressLength.
func NewAddressFromBytes(b []byte) *Address {
	a := new(Address)
	a.SetBytes(b)
	return a
}

// NewAddressFromHex parses a hex string (with or without a 0x prefix) into an Address.
func NewAddressFromHex(s string) (*Address, error) {
	b, err := encoding.Hex.Decode(s)
	if err != nil {
		return nil, fmt.Errorf("types: invalid address hex: %w", err)
	}
	if len(b) != AddressLength {
		return nil, fmt.Errorf("types: invalid address length: %d", len(b))
	}
	return NewAddressFromBytes(b), nil
}

// NewAddressFromBig builds an Address from a big-endian integer.
func NewAddressFromBig(a *big.Int) *Address {
	return NewAddressFromBytes(a.Bytes())
}

// IsHexAddress reports whether s is a validly formatted address (with or without a 0x prefix).
func IsHexAddress(s string) bool {
	b, err := encoding.Hex.Decode(s)
	return err == nil && len(b) == AddressLength
}

// SetBytes copies b into the address, left-padding or right-truncating to AddressLength.
func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a.bytes) {
		b = b[len(b)-AddressLength:]
	}
	copy(a.bytes[AddressLength-len(b):], b)
}

// Bytes returns a copy of the address's raw bytes.
func (a *Address) Bytes() []byte {
	b := make([]byte, AddressLength)
	copy(b, a.bytes[:])
	return b
}

// IsZero reports whether the address is nil or the all-zero address.
func (a *Address) IsZero() bool {
	return a == nil || a.bytes == [AddressLength]byte{}
}

// Equal reports whether a and o are the same address. It is nil-safe.
func (a *Address) Equal(o *Address) bool {
	if a == nil || o == nil {
		return a == o
	}
	return a.bytes == o.bytes
}

// String returns the EIP-55 mixed-case checksum encoding.
func (a *Address) String() string {
	if a.hex == "" {
		a.hex = a.checksum()
	}
	return a.hex
}

// Big returns the address's bytes interpreted as a big-endian integer.
func (a *Address) Big() *big.Int {
	return new(big.Int).SetBytes(a.bytes[:])
}

// checksum implements EIP-55: keccak256 of the lowercase hex (without the
// 0x prefix, as ASCII) decides, nibble by nibble, whether each hex letter
// in the address is upper- or lowercased.
func (a *Address) checksum() string {
	lower := encoding.Hex.Encode(a.bytes[:])
	digest := keccak.Hash256([]byte(lower)).Bytes()

	buf := make([]byte, len(lower))
	for i := 0; i < len(lower); i++ {
		c := lower[i]
		if c >= 'a' && c <= 'f' {
			var nibble byte
			if i%2 == 0 {
				nibble = digest[i/2] >> 4
			} else {
				nibble = digest[i/2] & 0x0f
			}
			if nibble >= 8 {
				c -= 'a' - 'A'
			}
		}
		buf[i] = c
	}
	return "0x" + string(buf)
}
