package erc2612

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/standards/eip712"
)

func TestNewPermit(t *testing.T) {
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	value := big.NewInt(100)
	nonce := big.NewInt(0)
	deadline := big.NewInt(9999999999)

	p := NewPermit(owner, spender, value, nonce, deadline)
	require.Same(t, owner, p.Owner)
	require.Same(t, spender, p.Spender)
	require.Same(t, value, p.Value)
	require.Same(t, nonce, p.Nonce)
	require.Same(t, deadline, p.Deadline)
}

func TestPermitTypesEncodeType(t *testing.T) {
	got := permitTypes.EncodeType(primaryType)
	want := "Permit(address owner,address spender,uint256 value,uint256 nonce,uint256 deadline)"
	require.Equal(t, want, got)
}

func TestPermitHash(t *testing.T) {
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	p := NewPermit(owner, spender, big.NewInt(100), big.NewInt(0), big.NewInt(9999999999))
	domain := eip712.NewDomain("evmkit test token", DefaultVersion, big.NewInt(1), contract, nil)

	got, err := p.Hash(domain)
	require.NoError(t, err)
	require.Len(t, got.Bytes(), 32)

	got2, err := p.Hash(domain)
	require.NoError(t, err)
	require.Equal(t, got.Bytes(), got2.Bytes())
}

func TestPermitHashMatchesManualEip712Hash(t *testing.T) {
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	value := big.NewInt(100)
	nonce := big.NewInt(3)
	deadline := big.NewInt(9999999999)

	p := NewPermit(owner, spender, value, nonce, deadline)
	domain := eip712.NewDomain("evmkit test token", DefaultVersion, big.NewInt(1), contract, nil)

	got, err := p.Hash(domain)
	require.NoError(t, err)

	want, err := eip712.Hash(domain, "Permit", eip712.NewTypes("Permit",
		eip712.NewField("owner", "address"),
		eip712.NewField("spender", "address"),
		eip712.NewField("value", "uint256"),
		eip712.NewField("nonce", "uint256"),
		eip712.NewField("deadline", "uint256"),
	), map[string]any{
		"owner":    owner,
		"spender":  spender,
		"value":    value,
		"nonce":    nonce,
		"deadline": deadline,
	})
	require.NoError(t, err)
	require.Equal(t, want.Bytes(), got.Bytes())
}

func TestPermitHashSensitiveToEachField(t *testing.T) {
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	domain := eip712.NewDomain("evmkit test token", DefaultVersion, big.NewInt(1), contract, nil)

	base := NewPermit(owner, spender, big.NewInt(100), big.NewInt(0), big.NewInt(9999999999))
	baseHash, err := base.Hash(domain)
	require.NoError(t, err)

	variants := []*Permit{
		NewPermit(spender, spender, big.NewInt(100), big.NewInt(0), big.NewInt(9999999999)),
		NewPermit(owner, owner, big.NewInt(100), big.NewInt(0), big.NewInt(9999999999)),
		NewPermit(owner, spender, big.NewInt(101), big.NewInt(0), big.NewInt(9999999999)),
		NewPermit(owner, spender, big.NewInt(100), big.NewInt(1), big.NewInt(9999999999)),
		NewPermit(owner, spender, big.NewInt(100), big.NewInt(0), big.NewInt(9999999998)),
	}
	for i, v := range variants {
		h, err := v.Hash(domain)
		require.NoError(t, err)
		require.NotEqual(t, baseHash.Bytes(), h.Bytes(), "variant %d", i)
	}
}

func TestNewRequestsSetFields(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sig, err := NewPermit(owner, spender, big.NewInt(1), big.NewInt(0), big.NewInt(2)).Sign(eip712.NewDomain("Token", DefaultVersion, big.NewInt(1), contract, nil), key)
	require.NoError(t, err)

	require.Equal(t, &NoncesRequest{Contract: contract, Owner: owner}, NewNoncesRequest(contract, owner))
	require.Equal(t, &DomainSeparatorRequest{Contract: contract}, NewDomainSeparatorRequest(contract))
	require.Equal(t, &PermitRequest{Contract: contract, Owner: owner, Spender: spender, Amount: big.NewInt(1), Deadline: big.NewInt(2), Signature: sig}, NewPermitRequest(contract, owner, spender, big.NewInt(1), big.NewInt(2), sig))
}

func TestRequestsValidate(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sig, err := NewPermit(owner, spender, big.NewInt(1), big.NewInt(0), big.NewInt(2)).Sign(eip712.NewDomain("Token", DefaultVersion, big.NewInt(1), contract, nil), key)
	require.NoError(t, err)
	amount := big.NewInt(1)
	deadline := big.NewInt(2)

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"nonces nil", (*NoncesRequest)(nil).Validate(), "invalid request: request is nil"},
		{"nonces empty", (&NoncesRequest{}).Validate(), "invalid request: contract is nil"},
		{"nonces nil owner", NewNoncesRequest(contract, nil).Validate(), "invalid request: owner is nil"},
		{"nonces valid", NewNoncesRequest(contract, owner).Validate(), ""},
		{"domainSeparator nil", (*DomainSeparatorRequest)(nil).Validate(), "invalid request: request is nil"},
		{"domainSeparator empty", (&DomainSeparatorRequest{}).Validate(), "invalid request: contract is nil"},
		{"domainSeparator valid", NewDomainSeparatorRequest(contract).Validate(), ""},
		{"permit nil", (*PermitRequest)(nil).Validate(), "invalid request: request is nil"},
		{"permit empty", (&PermitRequest{}).Validate(), "invalid request: contract is nil"},
		{"permit nil contract", NewPermitRequest(nil, owner, spender, amount, deadline, sig).Validate(), "invalid request: contract is nil"},
		{"permit nil owner", NewPermitRequest(contract, nil, spender, amount, deadline, sig).Validate(), "invalid request: owner is nil"},
		{"permit nil spender", NewPermitRequest(contract, owner, nil, amount, deadline, sig).Validate(), "invalid request: spender is nil"},
		{"permit nil amount", NewPermitRequest(contract, owner, spender, nil, deadline, sig).Validate(), "invalid request: amount is nil"},
		{"permit nil deadline", NewPermitRequest(contract, owner, spender, amount, nil, sig).Validate(), "invalid request: deadline is nil"},
		{"permit nil signature", NewPermitRequest(contract, owner, spender, amount, deadline, nil).Validate(), "invalid request: signature is nil"},
		{"permit only contract", NewPermitRequest(contract, nil, nil, nil, nil, nil).Validate(), "invalid request: owner is nil"},
		{"permit zero amount and deadline", NewPermitRequest(contract, owner, spender, big.NewInt(0), big.NewInt(0), sig).Validate(), ""},
		{"permit owner equals spender", NewPermitRequest(contract, owner, owner, amount, deadline, sig).Validate(), ""},
		{"permit valid", NewPermitRequest(contract, owner, spender, amount, deadline, sig).Validate(), ""},
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
