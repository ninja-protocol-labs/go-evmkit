package eip7702

import (
	"context"
	"encoding/json/v2"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
)

func TestGetDelegationClassifiesCode(t *testing.T) {
	account := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	delegate := types.MustNewAddressFromHex("0x63c0c19a282a1b52b07dd5a65b58948a07dae32b")

	for _, tc := range []struct {
		name   string
		result string
		want   *AccountCode
	}{
		{"empty", "0x", &AccountCode{Kind: EmptyCode}},
		{"delegated", "0xef010063c0c19a282a1b52b07dd5a65b58948a07dae32b", &AccountCode{Kind: DelegatedCode, Delegate: delegate}},
		{"delegated upper case", "0xEF010063C0C19A282A1B52B07DD5A65B58948A07DAE32B", &AccountCode{Kind: DelegatedCode, Delegate: delegate}},
		{"contract", "0x6080604052348015600e575f5ffd5b50", &AccountCode{Kind: ContractCode}},
		{"designator with trailing byte", "0xef010063c0c19a282a1b52b07dd5a65b58948a07dae32b00", &AccountCode{Kind: ContractCode}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				require.Equal(t, rpc.MethodETHGetCode, e.Method)
				require.Equal(t, []any{account.String(), rpc.BlockTagLatest}, e.Params)
				_, ok := e.Result.(*AccountCode)
				require.True(t, ok)
				raw, err := json.Marshal(tc.result)
				require.NoError(t, err)
				return json.Unmarshal(raw, e.Result)
			}).Once()

			got, err := NewEIP7702(c).GetDelegation(t.Context(), nil, NewGetDelegationRequest(account))
			require.NoError(t, err)
			require.Equal(t, tc.want.Kind, got.Kind)
			if tc.want.Delegate == nil {
				require.Nil(t, got.Delegate)
				return
			}
			require.True(t, tc.want.Delegate.Equal(got.Delegate))
		})
	}
}

func TestGetDelegationForwardsBlock(t *testing.T) {
	account := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
		require.Equal(t, []any{account.String(), "0x1b4"}, e.Params)
		return json.Unmarshal([]byte(`"0x"`), e.Result)
	}).Once()

	got, err := NewEIP7702(c).GetDelegation(t.Context(), bind.NewCallOptions().WithBlock("0x1b4").WithFrom(types.MustNewAddressFromHex("0x63c0c19a282a1b52b07dd5a65b58948a07dae32b")), NewGetDelegationRequest(account))
	require.NoError(t, err)
	require.Equal(t, EmptyCode, got.Kind)
}

func TestGetDelegationRejectsInvalidRequest(t *testing.T) {
	for _, tc := range []struct {
		name string
		req  *GetDelegationRequest
		want string
	}{
		{"nil request", nil, "eip7702: getDelegation: invalid request: request is nil"},
		{"nil account", NewGetDelegationRequest(nil), "eip7702: getDelegation: invalid request: account is nil"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewEIP7702(mocks.NewClient(t)).GetDelegation(t.Context(), nil, tc.req)
			require.ErrorIs(t, err, bind.ErrInvalidRequest)
			require.EqualError(t, err, tc.want)
			require.Nil(t, got)
		})
	}
}

func TestGetDelegationRejectsStateOverride(t *testing.T) {
	account := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	opts := bind.NewCallOptions().WithStateOverride(rpc.NewStateOverride().WithCode(account, []byte{0xef, 0x01, 0x00}))

	got, err := NewEIP7702(mocks.NewClient(t)).GetDelegation(t.Context(), opts, NewGetDelegationRequest(account))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "eip7702: getDelegation: invalid options: eth_getCode takes no state override")
	require.Nil(t, got)
}

func TestGetDelegationWrapsRPCError(t *testing.T) {
	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: -32000, Message: "header not found"}).Once()

	got, err := NewEIP7702(c).GetDelegation(t.Context(), nil, NewGetDelegationRequest(types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")))
	rpcErr, ok := errors.AsType[*rpc.ResponseError](err)
	require.True(t, ok)
	require.Equal(t, -32000, rpcErr.Code)
	require.ErrorContains(t, err, "eip7702: getDelegation: ")
	require.Nil(t, got)
}

func TestGetDelegationRejectsMalformedResult(t *testing.T) {
	for _, tc := range []struct {
		name   string
		result string
	}{
		{"null", `null`},
		{"empty string", `""`},
		{"odd length", `"0xef0"`},
		{"invalid character", `"0xzz"`},
		{"not a string", `42`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				return json.Unmarshal([]byte(tc.result), e.Result)
			}).Once()

			got, err := NewEIP7702(c).GetDelegation(t.Context(), nil, NewGetDelegationRequest(types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")))
			require.ErrorContains(t, err, "eip7702: getDelegation: ")
			require.Nil(t, got)
		})
	}
}

func TestGetDelegationPropagatesContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, _ rpc.Element) error {
		return ctx.Err()
	}).Once()

	got, err := NewEIP7702(c).GetDelegation(ctx, nil, NewGetDelegationRequest(types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")))
	require.ErrorIs(t, err, context.Canceled)
	require.Nil(t, got)
}

func TestGetDelegationElement(t *testing.T) {
	account := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	result := NewAccountCode(ContractCode, nil)

	e, err := GetDelegationElement("delegation", bind.NewCallOptions().WithBlock("0x10"), NewGetDelegationRequest(account), result)
	require.NoError(t, err)
	require.Equal(t, "delegation", e.ID)
	require.Equal(t, rpc.MethodETHGetCode, e.Method)
	require.Equal(t, []any{account.String(), "0x10"}, e.Params)
	require.Same(t, result, e.Result)
	require.Equal(t, &AccountCode{}, result)

	require.NoError(t, json.Unmarshal([]byte(`"0xef010063c0c19a282a1b52b07dd5a65b58948a07dae32b"`), e.Result))
	require.Equal(t, DelegatedCode, result.Kind)
	require.True(t, result.Delegate.Equal(types.MustNewAddressFromHex("0x63c0c19a282a1b52b07dd5a65b58948a07dae32b")))
}

func TestGetDelegationElementRejectsInvalidInput(t *testing.T) {
	account := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	for _, tc := range []struct {
		name   string
		opts   *bind.CallOptions
		req    *GetDelegationRequest
		result *AccountCode
		target error
		want   string
	}{
		{"nil request", nil, nil, &AccountCode{}, bind.ErrInvalidRequest, "eip7702: getDelegation: invalid request: request is nil"},
		{"nil result", nil, NewGetDelegationRequest(account), nil, bind.ErrInvalidRequest, "eip7702: getDelegation: invalid request: result is nil"},
		{"state override", bind.NewCallOptions().WithStateOverride(rpc.NewStateOverride().WithNonce(account, 1)), NewGetDelegationRequest(account), &AccountCode{}, bind.ErrInvalidOptions, "eip7702: getDelegation: invalid options: eth_getCode takes no state override"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := GetDelegationElement("delegation", tc.opts, tc.req, tc.result)
			require.ErrorIs(t, err, tc.target)
			require.EqualError(t, err, tc.want)
			require.Equal(t, rpc.Element{}, e)
		})
	}
}

func TestGetDelegationElementBatchesWithNonce(t *testing.T) {
	account := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	var code AccountCode
	var nonceHex string

	e, err := GetDelegationElement("delegation", nil, NewGetDelegationRequest(account), &code)
	require.NoError(t, err)
	elems := rpc.Elements{}
	elems.With(e)
	elems.With(rpc.ETHGetTransactionCount("nonce", account.String(), rpc.BlockTagLatest, &nonceHex))

	c := mocks.NewClient(t)
	c.EXPECT().Batch(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, got rpc.Elements) error {
		require.Equal(t, 2, got.Len())
		require.NoError(t, json.Unmarshal([]byte(`"0x"`), got.GetResult(0)))
		return json.Unmarshal([]byte(`"0x3"`), got.GetResult(1))
	}).Once()

	require.NoError(t, c.Batch(t.Context(), elems))
	require.Equal(t, EmptyCode, code.Kind)
	require.Equal(t, "0x3", nonceHex)
}
