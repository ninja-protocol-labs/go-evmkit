package core

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

var _ Signer = (*types.PrivateKey)(nil)

func TestPrivateKeySatisfiesSigner(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)

	var signer Signer = key
	require.Equal(t, key.PublicKey(), signer.PublicKey())

	digest := Keccak256([]byte("hello"))
	sig, err := signer.Sign(digest)
	require.NoError(t, err)
	require.NotNil(t, sig)
}

// keyBlobSigner simulates an HSM-backed signer: the caller only ever sees
// an opaque blob, never a *types.PrivateKey. The embedded key stands in
// for whatever the HSM does internally (a real client would call out to
// hardware here instead).
type keyBlobSigner struct {
	blob []byte
	key  *types.PrivateKey
}

func (s *keyBlobSigner) Sign(digest *types.Hash) (*types.Signature, error) {
	return s.key.Sign(digest)
}

func (s *keyBlobSigner) PublicKey() *types.PublicKey {
	return s.key.PublicKey()
}

var _ Signer = (*keyBlobSigner)(nil)

func TestKeyBlobStyleSignerSatisfiesInterface(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)

	var signer Signer = &keyBlobSigner{blob: []byte("opaque-hsm-key-handle"), key: key}

	digest := Keccak256([]byte("hello"))
	sig, err := signer.Sign(digest)
	require.NoError(t, err)

	require.True(t, signer.PublicKey().Verify(digest, sig))
}

func TestPubkeyToAddressKnownVector(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(1))
	require.NoError(t, err)

	got := PubkeyToAddress(key.PublicKey())
	want, err := types.NewAddressFromHex("0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf")
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestVerifyAddressRoundTrip(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	address := PubkeyToAddress(key.PublicKey())

	digest := Keccak256([]byte("hello"))
	sig, err := key.Sign(digest)
	require.NoError(t, err)

	ok, err := VerifyAddress(digest, sig, address)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestVerifyAddressRejectsWrongAddress(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	other, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	wrongAddress := PubkeyToAddress(other.PublicKey())

	digest := Keccak256([]byte("hello"))
	sig, err := key.Sign(digest)
	require.NoError(t, err)

	ok, err := VerifyAddress(digest, sig, wrongAddress)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestVerifyAddressRejectsWrongDigest(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	address := PubkeyToAddress(key.PublicKey())

	sig, err := key.Sign(Keccak256([]byte("hello")))
	require.NoError(t, err)

	ok, err := VerifyAddress(Keccak256([]byte("goodbye")), sig, address)
	require.NoError(t, err)
	require.False(t, ok)
}
