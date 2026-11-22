package deploy

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestRequestsValidate(t *testing.T) {
	factory := types.MustNewAddressFromHex("0x4e59b44847b379578588920cA78FbF26c0B4956C")
	deployment := NewContractDeployment([]byte{0x60, 0x80}, nil, nil)

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"create nil", (*CreateRequest)(nil).Validate(), "invalid request: request is nil"},
		{"create empty", (&CreateRequest{}).Validate(), "invalid request: deployment is nil"},
		{"create nil bytecode", (&CreateRequest{Deployment: NewContractDeployment(nil, nil, nil)}).Validate(), "invalid request: deployment bytecode is empty"},
		{"create empty bytecode", (&CreateRequest{Deployment: NewContractDeployment([]byte{}, nil, nil)}).Validate(), "invalid request: deployment bytecode is empty"},
		{"create valid", (&CreateRequest{Deployment: deployment}).Validate(), ""},
		{"create valid with args", (&CreateRequest{Deployment: NewContractDeployment([]byte{0x00}, abi.NewTypes(abi.Uint256), []any{big.NewInt(1)})}).Validate(), ""},
		{"create2 nil", (*Create2Request)(nil).Validate(), "invalid request: request is nil"},
		{"create2 empty", (&Create2Request{}).Validate(), "invalid request: factory is nil"},
		{"create2 nil factory", (&Create2Request{Deployment: NewDeterministicContractDeployment(deployment, [32]byte{1})}).Validate(), "invalid request: factory is nil"},
		{"create2 nil deployment", (&Create2Request{Factory: factory}).Validate(), "invalid request: deployment is nil"},
		{"create2 nil inner deployment", (&Create2Request{Factory: factory, Deployment: NewDeterministicContractDeployment(nil, [32]byte{1})}).Validate(), "invalid request: deployment is nil"},
		{"create2 empty bytecode", (&Create2Request{Factory: factory, Deployment: NewDeterministicContractDeployment(NewContractDeployment(nil, nil, nil), [32]byte{1})}).Validate(), "invalid request: deployment bytecode is empty"},
		{"create2 valid", (&Create2Request{Factory: factory, Deployment: NewDeterministicContractDeployment(deployment, [32]byte{1})}).Validate(), ""},
		{"create2 zero salt", (&Create2Request{Factory: factory, Deployment: NewDeterministicContractDeployment(deployment, [32]byte{})}).Validate(), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.want == "" {
				require.NoError(t, tc.err)
				return
			}
			require.ErrorIs(t, tc.err, bind.ErrInvalidRequest)
			require.EqualError(t, tc.err, tc.want)
		})
	}
}

func TestNewRequestsSetFields(t *testing.T) {
	factory := types.MustNewAddressFromHex("0x4e59b44847b379578588920cA78FbF26c0B4956C")
	deployment := NewContractDeployment([]byte{0x60, 0x80}, abi.NewTypes(abi.Uint256), []any{big.NewInt(1)})
	deterministic := NewDeterministicContractDeployment(deployment, [32]byte{7})

	require.Equal(t, &ContractDeployment{Bytecode: []byte{0x60, 0x80}, Constructor: abi.NewTypes(abi.Uint256), Args: []any{big.NewInt(1)}}, deployment)
	require.Equal(t, &DeterministicContractDeployment{ContractDeployment: deployment, Salt: [32]byte{7}}, deterministic)
	require.Equal(t, &CreateRequest{Deployment: deployment}, NewCreateRequest(deployment))
	require.Equal(t, &Create2Request{Factory: factory, Deployment: deterministic}, NewCreate2Request(factory, deterministic))
}

func TestInitCode(t *testing.T) {
	args, err := abi.Pack(abi.NewTypes(abi.String, abi.Uint8), "token", uint8(18))
	require.NoError(t, err)

	for _, tc := range []struct {
		name       string
		deployment *ContractDeployment
		want       []byte
	}{
		{"no constructor", NewContractDeployment([]byte{0x60, 0x80}, nil, nil), []byte{0x60, 0x80}},
		{"empty constructor", NewContractDeployment([]byte{0x60, 0x80}, abi.NewTypes(), []any{}), []byte{0x60, 0x80}},
		{"with args", NewContractDeployment([]byte{0x60, 0x80}, abi.NewTypes(abi.String, abi.Uint8), []any{"token", uint8(18)}), append([]byte{0x60, 0x80}, args...)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.deployment.InitCode()
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}

	bytecode := []byte{0x60, 0x80}
	got, err := NewContractDeployment(bytecode, abi.NewTypes(abi.Uint8), []any{uint8(1)}).InitCode()
	require.NoError(t, err)
	got[0] = 0xff
	require.Equal(t, []byte{0x60, 0x80}, bytecode)

	for _, tc := range []struct {
		name       string
		deployment *ContractDeployment
	}{
		{"missing arg", NewContractDeployment([]byte{0x60}, abi.NewTypes(abi.Uint256), nil)},
		{"extra arg", NewContractDeployment([]byte{0x60}, nil, []any{big.NewInt(1)})},
		{"wrong type", NewContractDeployment([]byte{0x60}, abi.NewTypes(abi.Uint256), []any{"one"})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.deployment.InitCode()
			require.Nil(t, got)
			require.ErrorContains(t, err, "deploy: initCode: ")
		})
	}
}

func TestEncodeERC2470Deploy(t *testing.T) {
	deployment := NewDeterministicContractDeployment(NewContractDeployment([]byte{0x60, 0x80}, abi.NewTypes(abi.Uint8), []any{uint8(9)}), [32]byte{0xaa, 0xbb})
	initCode, err := deployment.InitCode()
	require.NoError(t, err)

	got, err := EncodeERC2470Deploy(deployment)
	require.NoError(t, err)
	require.Equal(t, []byte{0x4a, 0xf6, 0x3f, 0x02}, got[:4])
	require.Equal(t, "deploy(bytes,bytes32)", erc2470DeployFn.Signature())

	vals, err := erc2470DeployFn.DecodeCall(got)
	require.NoError(t, err)
	require.Equal(t, initCode, vals[0])
	require.Equal(t, deployment.Salt[:], vals[1])

	_, err = EncodeERC2470Deploy(NewDeterministicContractDeployment(NewContractDeployment([]byte{0x60}, abi.NewTypes(abi.Uint256), nil), [32]byte{}))
	require.ErrorContains(t, err, "deploy: encodeERC2470Deploy: ")
}
