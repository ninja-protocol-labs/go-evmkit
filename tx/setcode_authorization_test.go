package tx

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
)

func TestSetCodePackRejectsForeignChainAuthorization(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	from := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	to := types.MustNewAddressFromHex("0xe05fcC23807536bEe418f142D19fa0d21BB0cfF7")
	auth := core.Authorization{ChainID: big.NewInt(10), Address: types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222"), Nonce: 0}
	require.NoError(t, auth.Sign(key))

	packed, err := NewSetCodeTxConfig(from, to, big.NewInt(0), nil, []core.Authorization{auth}).
		WithChainID(big.NewInt(8453)).WithNonce(5).WithGasLimit(100_000).WithGasTipCap(big.NewInt(1)).WithGasFeeCap(big.NewInt(2)).
		Pack(t.Context(), mocks.NewClient(t))
	require.ErrorIs(t, err, ErrAuthorizationChainID)
	require.EqualError(t, err, "tx: pack set code tx: authorization 0: authorization chain id does not match the tx chain id: 10, tx 8453")
	require.Nil(t, packed)
}

func TestSetCodePackAcceptsMatchingAndAnyChainAuthorization(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	from := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	to := types.MustNewAddressFromHex("0xe05fcC23807536bEe418f142D19fa0d21BB0cfF7")
	delegate := types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222")

	for _, chainID := range []*big.Int{big.NewInt(8453), big.NewInt(0), nil} {
		auth := core.Authorization{ChainID: chainID, Address: delegate, Nonce: 0}
		require.NoError(t, auth.Sign(key))

		packed, err := NewSetCodeTxConfig(from, to, big.NewInt(0), nil, []core.Authorization{auth}).
			WithChainID(big.NewInt(8453)).WithNonce(5).WithGasLimit(100_000).WithGasTipCap(big.NewInt(1)).WithGasFeeCap(big.NewInt(2)).
			Pack(t.Context(), mocks.NewClient(t))
		require.NoError(t, err, chainID)
		require.Equal(t, core.SetCodeTxType, packed.Type())
	}
}

func TestSetCodePackRejectsSelfSponsoredAuthorizationNonce(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	from := core.PubkeyToAddress(key.PublicKey())

	for _, tc := range []struct {
		name  string
		nonce uint64
		want  string
	}{
		{"same as tx nonce", 5, "tx: pack set code tx: authorization 0: self-sponsored authorization nonce must be the tx nonce + 1: nonce 5, tx nonce 5"},
		{"tx nonce + 2", 7, "tx: pack set code tx: authorization 0: self-sponsored authorization nonce must be the tx nonce + 1: nonce 7, tx nonce 5"},
		{"below tx nonce", 4, "tx: pack set code tx: authorization 0: self-sponsored authorization nonce must be the tx nonce + 1: nonce 4, tx nonce 5"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			auth := core.Authorization{ChainID: big.NewInt(8453), Address: types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222"), Nonce: tc.nonce}
			require.NoError(t, auth.Sign(key))

			packed, err := NewSetCodeTxConfig(from, from, big.NewInt(0), nil, []core.Authorization{auth}).
				WithChainID(big.NewInt(8453)).WithNonce(5).WithGasLimit(100_000).WithGasTipCap(big.NewInt(1)).WithGasFeeCap(big.NewInt(2)).
				Pack(t.Context(), mocks.NewClient(t))
			require.ErrorIs(t, err, ErrAuthorizationNonce)
			require.EqualError(t, err, tc.want)
			require.Nil(t, packed)
		})
	}
}

func TestSetCodePackAcceptsSelfSponsoredNextNonce(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	from := core.PubkeyToAddress(key.PublicKey())
	auth := core.Authorization{ChainID: big.NewInt(8453), Address: types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222"), Nonce: 6}
	require.NoError(t, auth.Sign(key))

	packed, err := NewSetCodeTxConfig(from, from, big.NewInt(0), nil, []core.Authorization{auth}).
		WithChainID(big.NewInt(8453)).WithNonce(5).WithGasLimit(100_000).WithGasTipCap(big.NewInt(1)).WithGasFeeCap(big.NewInt(2)).
		Pack(t.Context(), mocks.NewClient(t))
	require.NoError(t, err)
	require.Equal(t, uint64(5), packed.AddressNonce())
}

func TestSetCodePackSkipsNonceCheckForSponsoredAuthority(t *testing.T) {
	sponsor := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	authorityKey, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	auth := core.Authorization{ChainID: big.NewInt(8453), Address: types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222"), Nonce: 5}
	require.NoError(t, auth.Sign(authorityKey))

	packed, err := NewSetCodeTxConfig(sponsor, core.PubkeyToAddress(authorityKey.PublicKey()), big.NewInt(0), nil, []core.Authorization{auth}).
		WithChainID(big.NewInt(8453)).WithNonce(5).WithGasLimit(100_000).WithGasTipCap(big.NewInt(1)).WithGasFeeCap(big.NewInt(2)).
		Pack(t.Context(), mocks.NewClient(t))
	require.NoError(t, err)
	require.Equal(t, core.SetCodeTxType, packed.Type())
}

func TestSetCodePackSkipsNonceCheckForUnsignedAuthorization(t *testing.T) {
	from := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	auth := core.Authorization{ChainID: big.NewInt(8453), Address: types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222"), Nonce: 5}

	packed, err := NewSetCodeTxConfig(from, from, big.NewInt(0), nil, []core.Authorization{auth}).
		WithChainID(big.NewInt(8453)).WithNonce(5).WithGasLimit(100_000).WithGasTipCap(big.NewInt(1)).WithGasFeeCap(big.NewInt(2)).
		Pack(t.Context(), mocks.NewClient(t))
	require.NoError(t, err)
	require.Equal(t, core.SetCodeTxType, packed.Type())
}
