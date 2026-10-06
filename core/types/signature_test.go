package types

import (
	"bytes"
	"math/big"
	"testing"
)

func testSignature(t *testing.T) (*PrivateKey, *Hash, *Signature) {
	t.Helper()
	priv, err := GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	digest := NewHashFromBytes([]byte("some 32 byte digest padded out!"))
	sig, err := priv.Sign(digest)
	if err != nil {
		t.Fatal(err)
	}
	return priv, digest, sig
}

func TestNewSignatureFromBytes(t *testing.T) {
	_, _, sig := testSignature(t)

	parsed, err := NewSignatureFromBytes(sig.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if !sig.Equal(parsed) {
		t.Error("roundtrip through bytes changed the signature")
	}

	tests := []struct {
		name string
		in   []byte
	}{
		{"too short", make([]byte, SignatureLength-1)},
		{"too long", make([]byte, SignatureLength+1)},
		{"all zero", make([]byte, SignatureLength)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewSignatureFromBytes(tt.in); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestNewSignatureFromHex(t *testing.T) {
	_, _, sig := testSignature(t)

	parsed, err := NewSignatureFromHex(sig.String())
	if err != nil {
		t.Fatal(err)
	}
	if !sig.Equal(parsed) {
		t.Error("roundtrip through hex changed the signature")
	}

	if _, err := NewSignatureFromHex("0x1234"); err == nil {
		t.Error("expected error for too-short hex, got nil")
	}
}

func TestNewSignatureFromBig(t *testing.T) {
	_, _, sig := testSignature(t)

	parsed, err := NewSignatureFromBig(sig.Big())
	if err != nil {
		t.Fatal(err)
	}
	if !sig.Equal(parsed) {
		t.Error("roundtrip through big.Int changed the signature")
	}

	tooLarge := new(big.Int).Lsh(big.NewInt(1), 8*(SignatureLength+1))
	if _, err := NewSignatureFromBig(tooLarge); err == nil {
		t.Error("expected error for oversized big.Int, got nil")
	}
}

func TestSignatureBytesLayout(t *testing.T) {
	_, _, sig := testSignature(t)

	b := sig.Bytes()
	if len(b) != SignatureLength {
		t.Fatalf("length = %d, want %d", len(b), SignatureLength)
	}

	r := sig.R().FillBytes(make([]byte, 32))
	s := sig.S().FillBytes(make([]byte, 32))
	v := byte(sig.V().Uint64())

	want := append(append(append([]byte{}, r...), s...), v)
	if !bytes.Equal(b, want) {
		t.Errorf("Bytes() layout mismatch: got %x want %x", b, want)
	}
}

func TestSignatureBytesIsCopy(t *testing.T) {
	_, _, sig := testSignature(t)

	b1 := sig.Bytes()
	b1[0] ^= 0xFF
	b2 := sig.Bytes()
	if bytes.Equal(b1, b2) {
		t.Error("Bytes() must return a fresh copy each time")
	}
}

func TestSignatureLegacyV(t *testing.T) {
	_, _, sig := testSignature(t)

	want := new(big.Int).Add(sig.V(), big.NewInt(27))
	if sig.LegacyV().Cmp(want) != 0 {
		t.Errorf("got %s want %s", sig.LegacyV(), want)
	}
}

func TestSignatureEIP155V(t *testing.T) {
	_, _, sig := testSignature(t)

	chainID := big.NewInt(1)
	want := new(big.Int).Mul(chainID, big.NewInt(2))
	want.Add(want, big.NewInt(35))
	want.Add(want, sig.V())

	if sig.EIP155V(chainID).Cmp(want) != 0 {
		t.Errorf("got %s want %s", sig.EIP155V(chainID), want)
	}
}

func TestSignatureEqual(t *testing.T) {
	_, digest, sig1 := testSignature(t)
	sig2, err := NewSignatureFromBytes(sig1.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	priv3, err := GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	sig3, err := priv3.Sign(digest)
	if err != nil {
		t.Fatal(err)
	}
	var nilSig *Signature

	if !sig1.Equal(sig2) {
		t.Error("equal signatures reported as different")
	}
	if sig1.Equal(sig3) {
		t.Error("different signatures reported as equal")
	}
	if sig1.Equal(nilSig) {
		t.Error("non-nil signature equal to nil signature")
	}
	if !nilSig.Equal(nil) {
		t.Error("nil signature should equal nil")
	}
}

func TestSignatureIsZero(t *testing.T) {
	var nilSig *Signature
	if !nilSig.IsZero() {
		t.Error("nil signature should be zero")
	}

	_, _, sig := testSignature(t)
	if sig.IsZero() {
		t.Error("generated signature reported as zero")
	}
}

func TestSignatureStringCached(t *testing.T) {
	_, _, sig := testSignature(t)

	first := sig.String()
	second := sig.String()
	if first != second {
		t.Errorf("cached String() mismatch: %s vs %s", first, second)
	}
}

func TestSignatureECRecover(t *testing.T) {
	priv, digest, sig := testSignature(t)

	recovered, err := sig.ECRecover(digest)
	if err != nil {
		t.Fatal(err)
	}
	if !recovered.Equal(priv.PublicKey()) {
		t.Error("ECRecover did not return the signer's public key")
	}

	wrongDigest := NewHashFromBytes([]byte("a completely different digest!!"))
	recovered2, err := sig.ECRecover(wrongDigest)
	if err == nil && recovered2.Equal(priv.PublicKey()) {
		t.Error("ECRecover with the wrong digest should not recover the same key")
	}
}

func TestSignatureCompactRoundTrip(t *testing.T) {
	_, _, sig := testSignature(t)

	compact, err := sig.CompactBytes()
	if err != nil {
		t.Fatal(err)
	}
	if len(compact) != CompactSignatureLength {
		t.Fatalf("compact length = %d, want %d", len(compact), CompactSignatureLength)
	}

	parsed, err := NewSignatureFromCompact(compact)
	if err != nil {
		t.Fatal(err)
	}
	if !sig.Equal(parsed) {
		t.Error("roundtrip through compact bytes changed the signature")
	}
}

func TestSignatureCompactEncodesRecoveryIDInTopBit(t *testing.T) {
	_, _, sig := testSignature(t)

	compact, err := sig.CompactBytes()
	if err != nil {
		t.Fatal(err)
	}

	gotYParity := compact[32] >> 7
	if byte(sig.V().Uint64()) != gotYParity {
		t.Errorf("top bit of compact[32] = %d, want recovery id %d", gotYParity, sig.V().Uint64())
	}
}

func TestSignatureCompactRejectsHighRecoveryID(t *testing.T) {
	_, _, sig := testSignature(t)
	sig.v = 2

	if _, err := sig.CompactBytes(); err == nil {
		t.Error("expected an error for a recovery id above 1")
	}
}

func TestNewSignatureFromCompactRejectsWrongLength(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
	}{
		{"too short", make([]byte, CompactSignatureLength-1)},
		{"too long", make([]byte, CompactSignatureLength+1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewSignatureFromCompact(tt.in); err == nil {
				t.Error("expected an error")
			}
		})
	}
}

func TestSignatureCompactBytesDoesNotAliasCaller(t *testing.T) {
	_, _, sig := testSignature(t)

	compact, err := sig.CompactBytes()
	if err != nil {
		t.Fatal(err)
	}
	original := bytes.Clone(compact)
	compact[0] ^= 0xff

	again, err := sig.CompactBytes()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(again, original) {
		t.Error("mutating a returned compact encoding affected a later call")
	}
}
