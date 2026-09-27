package types

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"
)

func TestGeneratePrivateKey(t *testing.T) {
	k1, err := GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	k2, err := GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	if k1.Equal(k2) {
		t.Error("two generated keys should not be equal")
	}
	if len(k1.Bytes()) != PrivateKeyLength {
		t.Errorf("length = %d, want %d", len(k1.Bytes()), PrivateKeyLength)
	}
}

func TestNewPrivateKeyFromBytes(t *testing.T) {
	k1, err := GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	k2, err := NewPrivateKeyFromBytes(k1.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if !k1.Equal(k2) {
		t.Error("roundtrip through bytes changed the key")
	}

	tests := []struct {
		name string
		in   []byte
	}{
		{"too short", make([]byte, PrivateKeyLength-1)},
		{"too long", make([]byte, PrivateKeyLength+1)},
		{"all zero", make([]byte, PrivateKeyLength)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewPrivateKeyFromBytes(tt.in); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestNewPrivateKeyFromHex(t *testing.T) {
	k1, err := GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	hexStr := "0x" + hex.EncodeToString(k1.Bytes())

	k2, err := NewPrivateKeyFromHex(hexStr)
	if err != nil {
		t.Fatal(err)
	}
	if !k1.Equal(k2) {
		t.Error("roundtrip through hex changed the key")
	}

	if _, err := NewPrivateKeyFromHex("0x1234"); err == nil {
		t.Error("expected error for too-short hex, got nil")
	}
	if _, err := NewPrivateKeyFromHex("0x" + "zz" + hexStr[4:]); err == nil {
		t.Error("expected error for invalid hex chars, got nil")
	}
}

func TestNewPrivateKeyFromBig(t *testing.T) {
	k1, err := GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	k2, err := NewPrivateKeyFromBig(k1.Big())
	if err != nil {
		t.Fatal(err)
	}
	if !k1.Equal(k2) {
		t.Error("roundtrip through big.Int changed the key")
	}

	tooLarge := new(big.Int).Lsh(big.NewInt(1), 8*(PrivateKeyLength+1))
	if _, err := NewPrivateKeyFromBig(tooLarge); err == nil {
		t.Error("expected error for oversized big.Int, got nil")
	}
}

func TestPrivateKeyPublicKey(t *testing.T) {
	k, err := GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	pub1 := k.PublicKey()
	pub2 := k.PublicKey()
	if !pub1.Equal(pub2) {
		t.Error("PublicKey() should be deterministic")
	}
}

func TestPrivateKeyBytesIsCopy(t *testing.T) {
	k, err := GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	b1 := k.Bytes()
	b1[0] ^= 0xFF
	b2 := k.Bytes()
	if bytes.Equal(b1, b2) {
		t.Error("Bytes() must return a fresh copy each time")
	}
}

func TestPrivateKeyEqual(t *testing.T) {
	k1, err := GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	k2, err := NewPrivateKeyFromBytes(k1.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	k3, err := GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	var nilKey *PrivateKey

	if !k1.Equal(k2) {
		t.Error("equal keys reported as different")
	}
	if k1.Equal(k3) {
		t.Error("different keys reported as equal")
	}
	if k1.Equal(nilKey) {
		t.Error("non-nil key equal to nil key")
	}
	if !nilKey.Equal(nil) {
		t.Error("nil key should equal nil")
	}
}

func TestPrivateKeyIsZero(t *testing.T) {
	var nilKey *PrivateKey
	if !nilKey.IsZero() {
		t.Error("nil key should be zero")
	}

	k, err := GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	if k.IsZero() {
		t.Error("generated key reported as zero")
	}
}

func TestPrivateKeySign(t *testing.T) {
	k, err := GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	digest := NewHashFromBytes([]byte("some 32 byte digest padded out!"))

	sig, err := k.Sign(digest)
	if err != nil {
		t.Fatal(err)
	}
	if !k.PublicKey().Verify(digest, sig) {
		t.Error("signature does not verify against the signer's public key")
	}

	recovered, err := sig.ECRecover(digest)
	if err != nil {
		t.Fatal(err)
	}
	if !recovered.Equal(k.PublicKey()) {
		t.Error("ECRecover did not return the signer's public key")
	}
}
