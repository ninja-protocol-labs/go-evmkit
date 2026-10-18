package v2

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewPairCreated(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	token0 := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	token1 := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	pair := types.MustNewAddressFromHex("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc")
	allPairsLength := big.NewInt(7)

	pc := NewPairCreated(contract, token0, token1, pair, allPairsLength)
	require.Same(t, contract, pc.Contract)
	require.Same(t, token0, pc.Token0)
	require.Same(t, token1, pc.Token1)
	require.Same(t, pair, pc.Pair)
	require.Same(t, allPairsLength, pc.AllPairsLength)
}
