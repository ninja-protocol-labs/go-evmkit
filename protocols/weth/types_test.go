package weth

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewDeposit(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	dst := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	wad := big.NewInt(1000)

	d := NewDeposit(contract, dst, wad)
	require.Same(t, contract, d.Contract)
	require.Same(t, dst, d.Dst)
	require.Same(t, wad, d.Wad)
}

func TestNewWithdrawal(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	src := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	wad := big.NewInt(2000)

	w := NewWithdrawal(contract, src, wad)
	require.Same(t, contract, w.Contract)
	require.Same(t, src, w.Src)
	require.Same(t, wad, w.Wad)
}
