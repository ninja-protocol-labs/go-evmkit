package eip7702

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/rlp"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewAuthorization(t *testing.T) {
	chainID := big.NewInt(11155111)
	delegate := types.MustNewAddressFromHex("0x63c0c19a282a1b52b07dd5a65b58948a07dae32b")

	auth, err := NewAuthorization(chainID, delegate, 7)
	require.NoError(t, err)
	require.Equal(t, &core.Authorization{ChainID: big.NewInt(11155111), Address: delegate, Nonce: 7}, auth)
	require.Nil(t, auth.Signature)
}

func TestNewAuthorizationCopiesInputs(t *testing.T) {
	chainID := big.NewInt(11155111)
	delegate := types.MustNewAddressFromHex("0x63c0c19a282a1b52b07dd5a65b58948a07dae32b")

	auth, err := NewAuthorization(chainID, delegate, 7)
	require.NoError(t, err)
	chainID.SetInt64(1)
	delegate.SetBytes([]byte{0x01})
	require.Equal(t, int64(11155111), auth.ChainID.Int64())
	require.True(t, auth.Address.Equal(types.MustNewAddressFromHex("0x63c0c19a282a1b52b07dd5a65b58948a07dae32b")))
}

func TestNewAuthorizationRejectsInvalidInput(t *testing.T) {
	delegate := types.MustNewAddressFromHex("0x63c0c19a282a1b52b07dd5a65b58948a07dae32b")

	for _, tc := range []struct {
		name     string
		chainID  *big.Int
		delegate *types.Address
		nonce    uint64
		target   error
		want     string
	}{
		{"nil delegate", big.NewInt(1), nil, 0, ErrInvalidDelegate, "eip7702: newAuthorization: invalid delegate"},
		{"zero delegate", big.NewInt(1), types.ZeroAddress(), 0, ErrInvalidDelegate, "eip7702: newAuthorization: invalid delegate"},
		{"nil chain id", nil, delegate, 0, ErrInvalidChainID, "eip7702: newAuthorization: invalid chain id"},
		{"zero chain id", big.NewInt(0), delegate, 0, ErrInvalidChainID, "eip7702: newAuthorization: invalid chain id"},
		{"negative chain id", big.NewInt(-1), delegate, 0, ErrInvalidChainID, "eip7702: newAuthorization: invalid chain id"},
		{"max nonce", big.NewInt(1), delegate, math.MaxUint64, ErrInvalidNonce, "eip7702: newAuthorization: invalid nonce"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			auth, err := NewAuthorization(tc.chainID, tc.delegate, tc.nonce)
			require.ErrorIs(t, err, tc.target)
			require.EqualError(t, err, tc.want)
			require.Nil(t, auth)
		})
	}
}

func TestNewAuthorizationAcceptsBoundaryNonce(t *testing.T) {
	auth, err := NewAuthorization(big.NewInt(1), types.MustNewAddressFromHex("0x63c0c19a282a1b52b07dd5a65b58948a07dae32b"), math.MaxUint64-1)
	require.NoError(t, err)
	require.Equal(t, uint64(math.MaxUint64-1), auth.Nonce)
}

func TestNewRevocation(t *testing.T) {
	auth, err := NewRevocation(big.NewInt(84532), 3)
	require.NoError(t, err)
	require.Equal(t, int64(84532), auth.ChainID.Int64())
	require.True(t, auth.Address.IsZero())
	require.NotNil(t, auth.Address)
	require.Equal(t, uint64(3), auth.Nonce)
	require.Nil(t, auth.Signature)
}

func TestNewRevocationRejectsInvalidInput(t *testing.T) {
	for _, tc := range []struct {
		name    string
		chainID *big.Int
		nonce   uint64
		target  error
		want    string
	}{
		{"nil chain id", nil, 0, ErrInvalidChainID, "eip7702: newRevocation: invalid chain id"},
		{"zero chain id", big.NewInt(0), 0, ErrInvalidChainID, "eip7702: newRevocation: invalid chain id"},
		{"max nonce", big.NewInt(1), math.MaxUint64, ErrInvalidNonce, "eip7702: newRevocation: invalid nonce"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			auth, err := NewRevocation(tc.chainID, tc.nonce)
			require.ErrorIs(t, err, tc.target)
			require.EqualError(t, err, tc.want)
			require.Nil(t, auth)
		})
	}
}

func TestNewAuthorizationSignsAndRecoversAuthority(t *testing.T) {
	key, err := types.NewPrivateKeyFromHex("0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318")
	require.NoError(t, err)

	auth, err := NewAuthorization(big.NewInt(11155111), types.MustNewAddressFromHex("0x63c0c19a282a1b52b07dd5a65b58948a07dae32b"), 0)
	require.NoError(t, err)
	require.NoError(t, auth.Sign(key))

	authority, err := auth.Authority()
	require.NoError(t, err)
	require.True(t, authority.Equal(core.PubkeyToAddress(key.PublicKey())))
}

func TestSelfAuthorizationNonce(t *testing.T) {
	for _, tc := range []struct {
		name    string
		txNonce uint64
		want    uint64
	}{
		{"zero", 0, 1},
		{"seven", 7, 8},
		{"largest valid", math.MaxUint64 - 2, math.MaxUint64 - 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SelfAuthorizationNonce(tc.txNonce)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestSelfAuthorizationNonceRejectsOverflow(t *testing.T) {
	for _, txNonce := range []uint64{math.MaxUint64 - 1, math.MaxUint64} {
		got, err := SelfAuthorizationNonce(txNonce)
		require.ErrorIs(t, err, ErrInvalidNonce)
		require.EqualError(t, err, "eip7702: selfAuthorizationNonce: invalid nonce")
		require.Zero(t, got)
	}
}

func TestCheckAuthorization(t *testing.T) {
	key, err := types.NewPrivateKeyFromHex("0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318")
	require.NoError(t, err)
	authority := core.PubkeyToAddress(key.PublicKey())
	delegate := types.MustNewAddressFromHex("0x63c0c19a282a1b52b07dd5a65b58948a07dae32b")

	auth, err := NewAuthorization(big.NewInt(84532), delegate, 3)
	require.NoError(t, err)
	require.NoError(t, auth.Sign(key))

	require.NoError(t, CheckAuthorization(auth, authority, big.NewInt(84532), delegate))
}

func TestCheckAuthorizationRevocation(t *testing.T) {
	key, err := types.NewPrivateKeyFromHex("0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318")
	require.NoError(t, err)

	auth, err := NewRevocation(big.NewInt(84532), 4)
	require.NoError(t, err)
	require.NoError(t, auth.Sign(key))

	require.NoError(t, CheckAuthorization(auth, core.PubkeyToAddress(key.PublicKey()), big.NewInt(84532), types.ZeroAddress()))
}

func TestCheckAuthorizationRejectsMismatch(t *testing.T) {
	key, err := types.NewPrivateKeyFromHex("0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318")
	require.NoError(t, err)
	authority := core.PubkeyToAddress(key.PublicKey())
	delegate := types.MustNewAddressFromHex("0x63c0c19a282a1b52b07dd5a65b58948a07dae32b")
	other := types.MustNewAddressFromHex("0x000000000000000000000000000000000000dEaD")

	signed, err := NewAuthorization(big.NewInt(84532), delegate, 3)
	require.NoError(t, err)
	require.NoError(t, signed.Sign(key))
	anyChain := core.NewAuthorization(big.NewInt(0), delegate, 3)
	require.NoError(t, anyChain.Sign(key))
	maxNonce := core.NewAuthorization(big.NewInt(84532), delegate, math.MaxUint64)
	require.NoError(t, maxNonce.Sign(key))
	unsigned, err := NewAuthorization(big.NewInt(84532), delegate, 3)
	require.NoError(t, err)

	for _, tc := range []struct {
		name      string
		auth      *core.Authorization
		authority *types.Address
		chainID   *big.Int
		delegate  *types.Address
		target    error
	}{
		{"nil authorization", nil, authority, big.NewInt(84532), delegate, bind.ErrInvalidRequest},
		{"nil authority", signed, nil, big.NewInt(84532), delegate, bind.ErrInvalidRequest},
		{"nil delegate", signed, authority, big.NewInt(84532), nil, bind.ErrInvalidRequest},
		{"nil expected chain id", signed, authority, nil, delegate, ErrInvalidChainID},
		{"zero expected chain id", anyChain, authority, big.NewInt(0), delegate, ErrInvalidChainID},
		{"other chain", signed, authority, big.NewInt(8453), delegate, ErrInvalidChainID},
		{"any-chain authorization", anyChain, authority, big.NewInt(84532), delegate, ErrInvalidChainID},
		{"other delegate", signed, authority, big.NewInt(84532), other, ErrInvalidDelegate},
		{"revocation expected", signed, authority, big.NewInt(84532), types.ZeroAddress(), ErrInvalidDelegate},
		{"max nonce", maxNonce, authority, big.NewInt(84532), delegate, ErrInvalidNonce},
		{"unsigned", unsigned, authority, big.NewInt(84532), delegate, ErrInvalidSignature},
		{"other authority", signed, other, big.NewInt(84532), delegate, ErrInvalidAuthority},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := CheckAuthorization(tc.auth, tc.authority, tc.chainID, tc.delegate)
			require.ErrorIs(t, err, tc.target)
			require.ErrorContains(t, err, "eip7702: checkAuthorization: ")
		})
	}
}

func TestCheckAuthorizationRejectsHighS(t *testing.T) {
	key, err := types.NewPrivateKeyFromHex("0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318")
	require.NoError(t, err)
	authority := core.PubkeyToAddress(key.PublicKey())
	delegate := types.MustNewAddressFromHex("0x63c0c19a282a1b52b07dd5a65b58948a07dae32b")

	auth, err := NewAuthorization(big.NewInt(84532), delegate, 3)
	require.NoError(t, err)
	require.NoError(t, auth.Sign(key))

	n, ok := new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	require.True(t, ok)
	highS, err := rlp.SignatureFromRS(auth.Signature.R(), new(big.Int).Sub(n, auth.Signature.S()), byte(auth.Signature.V().Uint64())^1)
	require.NoError(t, err)
	auth.Signature = highS

	recovered, err := auth.Authority()
	require.Error(t, err)
	require.Nil(t, recovered)

	err = CheckAuthorization(auth, authority, big.NewInt(84532), delegate)
	require.ErrorIs(t, err, ErrInvalidSignature)
	require.ErrorContains(t, err, "eip7702: checkAuthorization: invalid signature: ")
}
