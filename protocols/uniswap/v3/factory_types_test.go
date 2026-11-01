package v3

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewOwnerChanged(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	oldOwner := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	newOwner := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")

	oc := NewOwnerChanged(contract, oldOwner, newOwner)
	require.Same(t, contract, oc.Contract)
	require.Same(t, oldOwner, oc.OldOwner)
	require.Same(t, newOwner, oc.NewOwner)
}

func TestNewPoolCreated(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	token0 := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	token1 := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	pool := types.MustNewAddressFromHex("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc")
	fee := big.NewInt(3000)
	tickSpacing := big.NewInt(60)

	pc := NewPoolCreated(contract, token0, token1, fee, tickSpacing, pool)
	require.Same(t, contract, pc.Contract)
	require.Same(t, token0, pc.Token0)
	require.Same(t, token1, pc.Token1)
	require.Same(t, fee, pc.Fee)
	require.Same(t, tickSpacing, pc.TickSpacing)
	require.Same(t, pool, pc.Pool)
}

func TestNewFeeAmountEnabled(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	fee := big.NewInt(500)
	tickSpacing := big.NewInt(10)

	fae := NewFeeAmountEnabled(contract, fee, tickSpacing)
	require.Same(t, contract, fae.Contract)
	require.Same(t, fee, fae.Fee)
	require.Same(t, tickSpacing, fae.TickSpacing)
}
