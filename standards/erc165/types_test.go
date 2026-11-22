package erc165

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

func TestNewSupportsInterfaceRequest(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	interfaceID := [4]byte{0x80, 0xac, 0x58, 0xcd}

	require.Equal(t, &SupportsInterfaceRequest{Contract: contract, InterfaceID: interfaceID}, NewSupportsInterfaceRequest(contract, interfaceID))
}

func TestSupportsInterfaceRequestValidate(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")

	for _, tc := range []struct {
		name string
		req  *SupportsInterfaceRequest
		want string
	}{
		{"nil request", nil, "invalid request: request is nil"},
		{"empty", &SupportsInterfaceRequest{}, "invalid request: contract is nil"},
		{"nil contract", NewSupportsInterfaceRequest(nil, [4]byte{0x01, 0xff, 0xc9, 0xa7}), "invalid request: contract is nil"},
		{"zero interface id", NewSupportsInterfaceRequest(contract, [4]byte{}), ""},
		{"invalid sentinel interface id", NewSupportsInterfaceRequest(contract, [4]byte{0xff, 0xff, 0xff, 0xff}), ""},
		{"erc165 interface id", NewSupportsInterfaceRequest(contract, [4]byte{0x01, 0xff, 0xc9, 0xa7}), ""},
		{"zero address contract", NewSupportsInterfaceRequest(types.ZeroAddress(), [4]byte{0x01, 0xff, 0xc9, 0xa7}), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			if tc.want == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, bind.ErrInvalidRequest)
			require.NotErrorIs(t, err, ErrNotImplementedERC165)
			require.EqualError(t, err, tc.want)
		})
	}
}

func TestSupportsInterfaceRejectsInvalidRequest(t *testing.T) {
	e := New(rpc.NewClient("http://127.0.0.1:0"))

	for _, tc := range []struct {
		name string
		req  *SupportsInterfaceRequest
		want string
	}{
		{"nil request", nil, "erc165: supportsInterface: invalid request: request is nil"},
		{"nil contract", NewSupportsInterfaceRequest(nil, [4]byte{0x01, 0xff, 0xc9, 0xa7}), "erc165: supportsInterface: invalid request: contract is nil"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ok, err := e.SupportsInterface(t.Context(), nil, tc.req)
			require.ErrorIs(t, err, bind.ErrInvalidRequest)
			require.NotErrorIs(t, err, ErrNotImplementedERC165)
			require.EqualError(t, err, tc.want)
			require.False(t, ok)
		})
	}
}

func TestSupportsInterfaceTransportErrorIsNotNotImplemented(t *testing.T) {
	e := New(rpc.NewClient("http://127.0.0.1:0"))

	ok, err := e.SupportsInterface(t.Context(), nil, NewSupportsInterfaceRequest(types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006"), [4]byte{0x01, 0xff, 0xc9, 0xa7}))
	require.Error(t, err)
	require.NotErrorIs(t, err, ErrNotImplementedERC165)
	require.NotErrorIs(t, err, bind.ErrInvalidRequest)
	require.False(t, ok)
}
