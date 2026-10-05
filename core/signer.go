package core

import "github.com/ninja-protocol-labs/go-evmkit/core/types"

// Signer produces signatures over a digest. *types.PrivateKey satisfies
// it directly; other implementations (hardware wallets, KMS) sign the
// same way without exposing the private key itself.
type Signer interface {
	Sign(digest *types.Hash) (*types.Signature, error)
	PublicKey() *types.PublicKey
}

// PubkeyToAddress returns the Ethereum address derived from pub: the last
// 20 bytes of keccak256(X || Y), the uncompressed point without its 0x04
// prefix.
func PubkeyToAddress(pub *types.PublicKey) *types.Address {
	point := pub.BytesUncompressed()[1:]
	digest := Keccak256(point)
	return types.NewAddressFromBytes(digest.Bytes()[12:])
}

// VerifyAddress reports whether sig, over digest, was produced by the
// private key for address.
func VerifyAddress(digest *types.Hash, sig *types.Signature, address *types.Address) (bool, error) {
	pub, err := sig.ECRecover(digest)
	if err != nil {
		return false, err
	}
	return PubkeyToAddress(pub).Equal(address), nil
}
