package types

import (
	"bytes"
	"math/big"
	"testing"
)

func TestNewPublicKeyFromBytes(t *testing.T) {
	priv, err := GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	pub := priv.PublicKey()

	compressed, err := NewPublicKeyFromBytes(pub.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if !pub.Equal(compressed) {
		t.Error("roundtrip through compressed bytes changed the key")
	}

	uncompressed, err := NewPublicKeyFromBytes(pub.BytesUncompressed())
	if err != nil {
		t.Fatal(err)
	}
	if !pub.Equal(uncompressed) {
		t.Error("roundtrip through uncompressed bytes changed the key")
	}

	tests := []struct {
		name string
		in   []byte
	}{
		{"too short", make([]byte, PublicKeyCompressedLength-1)},
		{"bad length", make([]byte, PublicKeyCompressedLength+1)},
		{"all zero compressed", make([]byte, PublicKeyCompressedLength)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewPublicKeyFromBytes(tt.in); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestNewPublicKeyFromHex(t *testing.T) {
	priv, err := GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	pub := priv.PublicKey()

	parsed, err := NewPublicKeyFromHex(pub.String())
	if err != nil {
		t.Fatal(err)
	}
	if !pub.Equal(parsed) {
		t.Error("roundtrip through hex changed the key")
	}

	if _, err := NewPublicKeyFromHex("0x1234"); err == nil {
		t.Error("expected error for too-short hex, got nil")
	}
}

func TestNewPublicKeyFromBig(t *testing.T) {
	priv, err := GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	pub := priv.PublicKey()

	parsed, err := NewPublicKeyFromBig(pub.Big())
	if err != nil {
		t.Fatal(err)
	}
	if !pub.Equal(parsed) {
		t.Error("roundtrip through big.Int changed the key")
	}

	tooLarge := new(big.Int).Lsh(big.NewInt(1), 8*(PublicKeyUncompressedLength+1))
	if _, err := NewPublicKeyFromBig(tooLarge); err == nil {
		t.Error("expected error for oversized big.Int, got nil")
	}
}

func TestPublicKeyBytesIsCopy(t *testing.T) {
	priv, err := GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	pub := priv.PublicKey()

	b1 := pub.Bytes()
	b1[0] ^= 0xFF
	b2 := pub.Bytes()
	if bytes.Equal(b1, b2) {
		t.Error("Bytes() must return a fresh copy each time")
	}
}

func TestPublicKeyEqual(t *testing.T) {
	priv1, err := GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	priv2, err := GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	pub1a := priv1.PublicKey()
	pub1b := priv1.PublicKey()
	pub2 := priv2.PublicKey()
	var nilKey *PublicKey

	if !pub1a.Equal(pub1b) {
		t.Error("equal keys reported as different")
	}
	if pub1a.Equal(pub2) {
		t.Error("different keys reported as equal")
	}
	if pub1a.Equal(nilKey) {
		t.Error("non-nil key equal to nil key")
	}
	if !nilKey.Equal(nil) {
		t.Error("nil key should equal nil")
	}
}

func TestPublicKeyIsZero(t *testing.T) {
	var nilKey *PublicKey
	if !nilKey.IsZero() {
		t.Error("nil key should be zero")
	}

	priv, err := GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	if priv.PublicKey().IsZero() {
		t.Error("derived key reported as zero")
	}
}

func TestPublicKeyStringCached(t *testing.T) {
	priv, err := GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	pub := priv.PublicKey()

	first := pub.String()
	second := pub.String()
	if first != second {
		t.Errorf("cached String() mismatch: %s vs %s", first, second)
	}
}

func TestPublicKeyVerifyRejectsWrongKey(t *testing.T) {
	priv1, err := GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	priv2, err := GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	digest := NewHashFromBytes([]byte("some 32 byte digest padded out!"))

	sig, err := priv1.Sign(digest)
	if err != nil {
		t.Fatal(err)
	}
	if priv2.PublicKey().Verify(digest, sig) {
		t.Error("signature verified against the wrong public key")
	}
}
