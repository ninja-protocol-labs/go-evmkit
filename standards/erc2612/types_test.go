package erc2612

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

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
		NewPermit(spender, spender, big.NewInt(100), big.NewInt(0), big.NewInt(9999999999)), // owner changed
		NewPermit(owner, owner, big.NewInt(100), big.NewInt(0), big.NewInt(9999999999)),     // spender changed
		NewPermit(owner, spender, big.NewInt(101), big.NewInt(0), big.NewInt(9999999999)),   // value changed
		NewPermit(owner, spender, big.NewInt(100), big.NewInt(1), big.NewInt(9999999999)),   // nonce changed
		NewPermit(owner, spender, big.NewInt(100), big.NewInt(0), big.NewInt(9999999998)),   // deadline changed
	}
	for i, v := range variants {
		h, err := v.Hash(domain)
		require.NoError(t, err)
		require.NotEqual(t, baseHash.Bytes(), h.Bytes(), "variant %d", i)
	}
}
