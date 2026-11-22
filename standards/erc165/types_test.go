package erc165

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
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
	e := New(mocks.NewClient(t))

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

func TestSupportsInterfaceSendsCallAndDecodesResult(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	interfaceID := [4]byte{0x80, 0xac, 0x58, 0xcd}
	calldata, err := EncodeSupportsInterface(interfaceID)
	require.NoError(t, err)

	for _, tc := range []struct {
		name   string
		result string
		want   bool
	}{
		{"supported", "0x0000000000000000000000000000000000000000000000000000000000000001", true},
		{"unsupported", "0x0000000000000000000000000000000000000000000000000000000000000000", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				require.Equal(t, rpc.MethodETHCall, e.Method)
				params, ok := e.Params.([]any)
				require.True(t, ok)
				require.Len(t, params, 2)
				call, ok := params[0].(map[string]any)
				require.True(t, ok)
				require.Equal(t, contract.String(), call["to"])
				require.Equal(t, encoding.Hex.EncodePrefixed(calldata), call["data"])
				require.NotContains(t, call, "from")
				require.Equal(t, rpc.BlockTagLatest, params[1])

				result, ok := e.Result.(*string)
				require.True(t, ok)
				*result = tc.result
				return nil
			}).Once()

			ok, err := New(c).SupportsInterface(t.Context(), nil, NewSupportsInterfaceRequest(contract, interfaceID))
			require.NoError(t, err)
			require.Equal(t, tc.want, ok)
		})
	}
}

func TestSupportsInterfaceForwardsCallOptions(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	from := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
		params := e.Params.([]any)
		call := params[0].(map[string]any)
		require.Equal(t, from.String(), call["from"])
		require.Equal(t, "0x1b4", params[1])
		*e.Result.(*string) = "0x0000000000000000000000000000000000000000000000000000000000000001"
		return nil
	}).Once()

	ok, err := New(c).SupportsInterface(t.Context(), bind.NewCallOptions().WithFrom(from).WithBlock("0x1b4"), NewSupportsInterfaceRequest(contract, [4]byte{0x01, 0xff, 0xc9, 0xa7}))
	require.NoError(t, err)
	require.True(t, ok)
}

func TestSupportsInterfaceMapsRevertToNotImplemented(t *testing.T) {
	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: 3, Message: "execution reverted"}).Once()

	ok, err := New(c).SupportsInterface(t.Context(), nil, NewSupportsInterfaceRequest(types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006"), [4]byte{0x01, 0xff, 0xc9, 0xa7}))
	require.ErrorIs(t, err, ErrNotImplementedERC165)
	var rpcErr *rpc.ResponseError
	require.ErrorAs(t, err, &rpcErr)
	require.Equal(t, 3, rpcErr.Code)
	require.False(t, ok)
}

func TestSupportsInterfaceMapsUndecodableResultToNotImplemented(t *testing.T) {
	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
		*e.Result.(*string) = "0x01"
		return nil
	}).Once()

	ok, err := New(c).SupportsInterface(t.Context(), nil, NewSupportsInterfaceRequest(types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006"), [4]byte{0x01, 0xff, 0xc9, 0xa7}))
	require.ErrorIs(t, err, ErrNotImplementedERC165)
	require.False(t, ok)
}

func TestSupportsInterfaceTransportErrorIsNotNotImplemented(t *testing.T) {
	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(errors.New("rpc: send http request: connection refused")).Once()

	ok, err := New(c).SupportsInterface(t.Context(), nil, NewSupportsInterfaceRequest(types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006"), [4]byte{0x01, 0xff, 0xc9, 0xa7}))
	require.Error(t, err)
	require.ErrorContains(t, err, "connection refused")
	require.NotErrorIs(t, err, ErrNotImplementedERC165)
	require.NotErrorIs(t, err, bind.ErrInvalidRequest)
	require.False(t, ok)
}
