package erc721

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewTransfer(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	from := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	to := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	tokenId := big.NewInt(42)

	tr := NewTransfer(contract, from, to, tokenId)
	require.Same(t, contract, tr.Contract)
	require.Same(t, from, tr.From)
	require.Same(t, to, tr.To)
	require.Same(t, tokenId, tr.TokenId)
}

func TestNewApproval(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	approved := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	tokenId := big.NewInt(7)

	ap := NewApproval(contract, owner, approved, tokenId)
	require.Same(t, contract, ap.Contract)
	require.Same(t, owner, ap.Owner)
	require.Same(t, approved, ap.Approved)
	require.Same(t, tokenId, ap.TokenId)
}

func TestNewApprovalForAll(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	operator := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	afa := NewApprovalForAll(contract, owner, operator, true)
	require.Same(t, contract, afa.Contract)
	require.Same(t, owner, afa.Owner)
	require.Same(t, operator, afa.Operator)
	require.True(t, afa.Approved)
}

func TestNewMetadata(t *testing.T) {
	m := NewMetadata("Bored Ape Yacht Club", "BAYC")
	require.Equal(t, "Bored Ape Yacht Club", m.Name)
	require.Equal(t, "BAYC", m.Symbol)
}

func TestNewMetadataWithTokenID(t *testing.T) {
	m := NewMetadata("Bored Ape Yacht Club", "BAYC")
	tokenId := big.NewInt(1234)
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	mt := NewMetadataWithTokenID(m, tokenId, "ipfs://1234", owner)
	require.Equal(t, *m, mt.Metadata)
	require.Same(t, tokenId, mt.TokenId)
	require.Equal(t, "ipfs://1234", mt.TokenURI)
	require.Same(t, owner, mt.Owner)
}

func TestNewAllowanceWithTokenID(t *testing.T) {
	tokenId := big.NewInt(99)
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	operator := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	awt := NewAllowanceWithTokenID(tokenId, owner, operator)
	require.Same(t, tokenId, awt.TokenId)
	require.Same(t, owner, awt.Owner)
	require.Same(t, operator, awt.Operator)
}

func TestNewApprovalForAllWithBalance(t *testing.T) {
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	operator := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	ownerBalance := big.NewInt(3)
	operatorBalance := big.NewInt(4)

	afawb := NewApprovalForAllWithBalance(owner, operator, true, ownerBalance, operatorBalance)
	require.Same(t, owner, afawb.Owner)
	require.Same(t, operator, afawb.Operator)
	require.True(t, afawb.Approved)
	require.Same(t, ownerBalance, afawb.OwnerBalance)
	require.Same(t, operatorBalance, afawb.OperatorBalance)
}

func TestNewTokenOwner(t *testing.T) {
	tokenId := big.NewInt(5)
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	to := NewTokenOwner(tokenId, owner)
	require.Same(t, tokenId, to.TokenId)
	require.Same(t, owner, to.Owner)
}

func TestNewTokenOwners(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	tokens := []TokenOwner{
		*NewTokenOwner(big.NewInt(1), types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")),
		*NewTokenOwner(big.NewInt(2), types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")),
	}

	tos := NewTokenOwners(contract, tokens)
	require.Same(t, contract, tos.Contract)
	require.Equal(t, tokens, tos.Tokens)
}

func TestNewTokenOperator(t *testing.T) {
	tokenId := big.NewInt(6)
	operator := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	top := NewTokenOperator(tokenId, operator)
	require.Same(t, tokenId, top.TokenId)
	require.Same(t, operator, top.Operator)
}

func TestNewTokenOperators(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	tokens := []TokenOperator{
		*NewTokenOperator(big.NewInt(1), types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")),
		*NewTokenOperator(big.NewInt(2), types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")),
	}

	tops := NewTokenOperators(contract, tokens)
	require.Same(t, contract, tops.Contract)
	require.Equal(t, tokens, tops.Tokens)
}

func TestNewTokenBalance(t *testing.T) {
	token := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	amount := big.NewInt(4)

	tb := NewTokenBalance(token, amount)
	require.Same(t, token, tb.Token)
	require.Same(t, amount, tb.Amount)
}

func TestNewAddressBalances(t *testing.T) {
	address := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	tokens := []TokenBalance{
		*NewTokenBalance(types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d"), big.NewInt(1)),
		*NewTokenBalance(types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8"), big.NewInt(2)),
	}

	ab := NewAddressBalances(address, tokens)
	require.Same(t, address, ab.Address)
	require.Equal(t, tokens, ab.Tokens)
}

func TestNewTokenOwnerPair(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	tokenId := big.NewInt(8)

	top := NewTokenOwnerPair(contract, owner, tokenId)
	require.Same(t, contract, top.Contract)
	require.Same(t, tokenId, top.TokenId)
	require.Same(t, owner, top.Owner)
}

func TestNewAddressBalance(t *testing.T) {
	address := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	amount := big.NewInt(5)

	ab := NewAddressBalance(address, amount)
	require.Same(t, address, ab.Address)
	require.Same(t, amount, ab.Amount)
}

func TestNewTokenBalances(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	addrs := []AddressBalance{
		*NewAddressBalance(types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f"), big.NewInt(1)),
		*NewAddressBalance(types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8"), big.NewInt(2)),
	}

	tbs := NewTokenBalances(contract, addrs)
	require.Same(t, contract, tbs.Contract)
	require.Equal(t, addrs, tbs.Addresses)
}
