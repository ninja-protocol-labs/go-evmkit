package eip7702

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"math/big"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

func FuzzParseDelegationDoesNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0xef, 0x01, 0x00})
	f.Add(append([]byte{0xef, 0x01, 0x00}, make([]byte, 20)...))
	f.Add(append([]byte{0xef, 0x01, 0x00}, make([]byte, 21)...))
	f.Add([]byte{0x60, 0x80, 0x60, 0x40, 0x52})

	f.Fuzz(func(t *testing.T, code []byte) {
		delegate, ok := ParseDelegation(code)
		if !ok {
			require.Nil(t, delegate)
			return
		}
		require.Len(t, code, DelegationLength)
		require.True(t, bytes.HasPrefix(code, []byte{0xef, 0x01, 0x00}))
		require.Equal(t, code[3:], delegate.Bytes())
	})
}

func FuzzEncodeDelegationRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20))
	f.Add([]byte{0x63, 0xc0, 0xc1, 0x9a, 0x28, 0x2a, 0x1b, 0x52, 0xb0, 0x7d, 0xd5, 0xa6, 0x5b, 0x58, 0x94, 0x8a, 0x07, 0xda, 0xe3, 0x2b})
	f.Add([]byte{0x01})

	f.Fuzz(func(t *testing.T, raw []byte) {
		delegate := types.NewAddressFromBytes(raw)

		code, err := EncodeDelegation(delegate)
		if delegate.IsZero() {
			require.ErrorIs(t, err, ErrInvalidDelegate)
			require.Nil(t, code)
			return
		}
		require.NoError(t, err)
		require.Len(t, code, DelegationLength)

		got, ok := ParseDelegation(code)
		require.True(t, ok)
		require.True(t, got.Equal(delegate))
	})
}

func FuzzNewAuthorization(f *testing.F) {
	f.Add([]byte{0x01}, make([]byte, 20), uint64(0))
	f.Add([]byte{}, []byte{0x01}, uint64(7))
	f.Add([]byte{0xaa, 0x36, 0xa7}, []byte{0xde, 0xad}, ^uint64(0))

	f.Fuzz(func(t *testing.T, chainIDBytes []byte, delegateBytes []byte, nonce uint64) {
		chainID := new(big.Int).SetBytes(chainIDBytes)
		delegate := types.NewAddressFromBytes(delegateBytes)

		auth, err := NewAuthorization(chainID, delegate, nonce)
		switch {
		case delegate.IsZero():
			require.ErrorIs(t, err, ErrInvalidDelegate)
			require.Nil(t, auth)
		case chainID.Sign() == 0:
			require.ErrorIs(t, err, ErrInvalidChainID)
			require.Nil(t, auth)
		case nonce == ^uint64(0):
			require.ErrorIs(t, err, ErrInvalidNonce)
			require.Nil(t, auth)
		default:
			require.NoError(t, err)
			require.Zero(t, chainID.Cmp(auth.ChainID))
			require.True(t, delegate.Equal(auth.Address))
			require.Equal(t, nonce, auth.Nonce)
		}
	})
}

func FuzzSelfAuthorizationNonce(f *testing.F) {
	f.Add(uint64(0))
	f.Add(uint64(7))
	f.Add(^uint64(0) - 2)
	f.Add(^uint64(0) - 1)
	f.Add(^uint64(0))

	f.Fuzz(func(t *testing.T, txNonce uint64) {
		got, err := SelfAuthorizationNonce(txNonce)
		if txNonce >= ^uint64(0)-1 {
			require.ErrorIs(t, err, ErrInvalidNonce)
			require.Zero(t, got)
			return
		}
		require.NoError(t, err)
		require.Equal(t, txNonce+1, got)
		require.NotEqual(t, ^uint64(0), got)
	})
}

func FuzzGetDelegationThroughClient(f *testing.F) {
	f.Add("0x")
	f.Add("0xef010063c0c19a282a1b52b07dd5a65b58948a07dae32b")
	f.Add("0x6080604052")
	f.Add("")
	f.Add("0xzz")

	f.Fuzz(func(t *testing.T, result string) {
		raw, err := json.Marshal(result)
		if err != nil {
			t.Skip()
		}
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			return json.Unmarshal(raw, e.Result)
		}).Once()

		got, err := NewEIP7702(c).GetDelegation(t.Context(), nil, NewGetDelegationRequest(types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")))
		code, decodeErr := encoding.Hex.Decode(result)
		if result == "" || decodeErr != nil {
			require.Error(t, err)
			require.Nil(t, got)
			return
		}
		require.NoError(t, err)
		switch {
		case len(code) == 0:
			require.Equal(t, EmptyCode, got.Kind)
			require.Nil(t, got.Delegate)
		case len(code) == DelegationLength && bytes.HasPrefix(code, []byte{0xef, 0x01, 0x00}):
			require.Equal(t, DelegatedCode, got.Kind)
			require.Equal(t, code[3:], got.Delegate.Bytes())
		default:
			require.Equal(t, ContractCode, got.Kind)
			require.Nil(t, got.Delegate)
		}
	})
}

func FuzzCheckAuthorizationDoesNotPanic(f *testing.F) {
	f.Add([]byte{0x01}, make([]byte, 20), uint64(0), make([]byte, 65))
	f.Add([]byte{0x01, 0x4a, 0x34}, []byte{0xde, 0xad}, uint64(3), append(make([]byte, 64), 0x01))
	f.Add([]byte{}, []byte{}, ^uint64(0), []byte{0xff})

	f.Fuzz(func(t *testing.T, chainIDBytes []byte, delegateBytes []byte, nonce uint64, sigBytes []byte) {
		chainID := new(big.Int).SetBytes(chainIDBytes)
		delegate := types.NewAddressFromBytes(delegateBytes)
		auth := core.NewAuthorization(chainID, delegate, nonce)
		if sig, err := types.NewSignatureFromBytes(sigBytes); err == nil {
			auth.Signature = sig
		}

		err := CheckAuthorization(auth, types.MustNewAddressFromHex("0x2c7536E3605D9C16a7a3D7b1898e529396a65c23"), chainID, delegate)
		if err == nil {
			signer, recoverErr := auth.Authority()
			require.NoError(t, recoverErr)
			require.True(t, signer.Equal(types.MustNewAddressFromHex("0x2c7536E3605D9C16a7a3D7b1898e529396a65c23")))
		}
	})
}
