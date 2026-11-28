package eip7702

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewAccountCode(t *testing.T) {
	delegate := types.MustNewAddressFromHex("0x63c0c19a282a1b52b07dd5a65b58948a07dae32b")

	require.Equal(t, &AccountCode{Kind: DelegatedCode, Delegate: delegate}, NewAccountCode(DelegatedCode, delegate))
	require.Equal(t, &AccountCode{Kind: EmptyCode}, NewAccountCode(EmptyCode, nil))
	require.Equal(t, &AccountCode{Kind: ContractCode}, NewAccountCode(ContractCode, nil))
}

func TestCodeKindString(t *testing.T) {
	require.Equal(t, "empty", EmptyCode.String())
	require.Equal(t, "delegated", DelegatedCode.String())
	require.Equal(t, "contract", ContractCode.String())
	require.Equal(t, "CodeKind(9)", CodeKind(9).String())
}

func TestNewGetDelegationRequest(t *testing.T) {
	account := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	require.Equal(t, &GetDelegationRequest{Account: account}, NewGetDelegationRequest(account))
}

func TestGetDelegationRequestValidate(t *testing.T) {
	for _, tc := range []struct {
		name string
		req  *GetDelegationRequest
		want string
	}{
		{"nil request", nil, "invalid request: request is nil"},
		{"empty", &GetDelegationRequest{}, "invalid request: account is nil"},
		{"nil account", NewGetDelegationRequest(nil), "invalid request: account is nil"},
		{"account", NewGetDelegationRequest(types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")), ""},
		{"zero address account", NewGetDelegationRequest(types.ZeroAddress()), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			if tc.want == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, bind.ErrInvalidRequest)
			require.EqualError(t, err, tc.want)
		})
	}
}
