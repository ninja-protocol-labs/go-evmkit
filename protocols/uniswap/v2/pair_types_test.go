package v2

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewMint(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	sender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	amount0 := big.NewInt(100)
	amount1 := big.NewInt(200)

	m := NewMint(contract, sender, amount0, amount1)
	require.Same(t, contract, m.Contract)
	require.Same(t, sender, m.Sender)
	require.Same(t, amount0, m.Amount0)
	require.Same(t, amount1, m.Amount1)
}

func TestNewBurn(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	sender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	to := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	amount0 := big.NewInt(100)
	amount1 := big.NewInt(200)

	b := NewBurn(contract, sender, amount0, amount1, to)
	require.Same(t, contract, b.Contract)
	require.Same(t, sender, b.Sender)
	require.Same(t, amount0, b.Amount0)
	require.Same(t, amount1, b.Amount1)
	require.Same(t, to, b.To)
}

func TestNewSwap(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	sender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	to := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	amount0In := big.NewInt(100)
	amount1In := big.NewInt(0)
	amount0Out := big.NewInt(0)
	amount1Out := big.NewInt(95)

	s := NewSwap(contract, sender, amount0In, amount1In, amount0Out, amount1Out, to)
	require.Same(t, contract, s.Contract)
	require.Same(t, sender, s.Sender)
	require.Same(t, amount0In, s.Amount0In)
	require.Same(t, amount1In, s.Amount1In)
	require.Same(t, amount0Out, s.Amount0Out)
	require.Same(t, amount1Out, s.Amount1Out)
	require.Same(t, to, s.To)
}

func TestNewSync(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	reserve0 := big.NewInt(1000)
	reserve1 := big.NewInt(2000)

	s := NewSync(contract, reserve0, reserve1)
	require.Same(t, contract, s.Contract)
	require.Same(t, reserve0, s.Reserve0)
	require.Same(t, reserve1, s.Reserve1)
}

func TestNewReserves(t *testing.T) {
	reserve0 := big.NewInt(1000)
	reserve1 := big.NewInt(2000)
	blockTimestampLast := uint32(123456)

	r := NewReserves(reserve0, reserve1, blockTimestampLast)
	require.Same(t, reserve0, r.Reserve0)
	require.Same(t, reserve1, r.Reserve1)
	require.Equal(t, blockTimestampLast, r.BlockTimestampLast)
}
