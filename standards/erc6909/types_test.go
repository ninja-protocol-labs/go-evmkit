package erc6909

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewOperatorSet(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	owner := types.MustNewAddressFromHex("0x1231deb6f5749ef6ce6943a275a1d3e7486f4eae")
	operator := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	os := NewOperatorSet(contract, owner, operator, true)
	require.Same(t, contract, os.Contract)
	require.Same(t, owner, os.Owner)
	require.Same(t, operator, os.Operator)
	require.True(t, os.Approved)
}

func TestNewApproval(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	owner := types.MustNewAddressFromHex("0x1231deb6f5749ef6ce6943a275a1d3e7486f4eae")
	spender := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	id := big.NewInt(1)
	amount := big.NewInt(100)

	a := NewApproval(contract, owner, spender, id, amount)
	require.Same(t, contract, a.Contract)
	require.Same(t, owner, a.Owner)
	require.Same(t, spender, a.Spender)
	require.Same(t, id, a.Id)
	require.Same(t, amount, a.Amount)
}

func TestNewTransfer(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	caller := types.MustNewAddressFromHex("0x1231deb6f5749ef6ce6943a275a1d3e7486f4eae")
	sender := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	receiver := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	id := big.NewInt(42)
	amount := big.NewInt(100)

	tr := NewTransfer(contract, caller, sender, receiver, id, amount)
	require.Same(t, contract, tr.Contract)
	require.Same(t, caller, tr.Caller)
	require.Same(t, sender, tr.Sender)
	require.Same(t, receiver, tr.Receiver)
	require.Same(t, id, tr.Id)
	require.Same(t, amount, tr.Amount)
}

func TestNewNameUpdated(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	id := big.NewInt(1)

	n := NewNameUpdated(contract, id, "evmkit test token")
	require.Same(t, contract, n.Contract)
	require.Same(t, id, n.Id)
	require.Equal(t, "evmkit test token", n.NewName)
}

func TestNewSymbolUpdated(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	id := big.NewInt(1)

	s := NewSymbolUpdated(contract, id, "EVK6909")
	require.Same(t, contract, s.Contract)
	require.Same(t, id, s.Id)
	require.Equal(t, "EVK6909", s.NewSymbol)
}

func TestNewDecimalsUpdated(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	id := big.NewInt(1)

	d := NewDecimalsUpdated(contract, id, 18)
	require.Same(t, contract, d.Contract)
	require.Same(t, id, d.Id)
	require.Equal(t, uint8(18), d.NewDecimals)
}
