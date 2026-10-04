package abi

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestFunctionSignature(t *testing.T) {
	fn := NewFunction("transfer", Types{Address, Uint256}, Types{Bool})
	require.Equal(t, "transfer(address,uint256)", fn.Signature())
}

func TestFunctionSignatureNoArgs(t *testing.T) {
	fn := NewFunction("totalSupply", Types{}, Types{Uint256})
	require.Equal(t, "totalSupply()", fn.Signature())
}

func TestFunctionSelectorTransfer(t *testing.T) {
	fn := NewFunction("transfer", Types{Address, Uint256}, Types{Bool})
	require.Equal(t, Selector{0xa9, 0x05, 0x9c, 0xbb}, fn.Selector())
}

func TestFunctionSelectorApprove(t *testing.T) {
	fn := NewFunction("approve", Types{Address, Uint256}, Types{Bool})
	require.Equal(t, Selector{0x09, 0x5e, 0xa7, 0xb3}, fn.Selector())
}

func TestFunctionSelectorBalanceOf(t *testing.T) {
	fn := NewFunction("balanceOf", Types{Address}, Types{Uint256})
	require.Equal(t, Selector{0x70, 0xa0, 0x82, 0x31}, fn.Selector())
}

func TestFunctionSelectorTotalSupply(t *testing.T) {
	fn := NewFunction("totalSupply", Types{}, Types{Uint256})
	require.Equal(t, Selector{0x18, 0x16, 0x0d, 0xdd}, fn.Selector())
}

func TestFunctionEncodeCall(t *testing.T) {
	addr, err := types.NewAddressFromHex("0x833e1D0b8Bc979D49d57b65dCF18364694B16D52")
	require.NoError(t, err)
	amount := big.NewInt(1000000000000000000)

	fn := NewFunction("transfer", Types{Address, Uint256}, Types{Bool})
	got, err := fn.EncodeCall(addr, amount)
	require.NoError(t, err)

	sel := fn.Selector()
	require.Equal(t, sel[:], got[:4])

	wantArgs, err := Pack(fn.Inputs, addr, amount)
	require.NoError(t, err)
	require.Equal(t, wantArgs, got[4:])
}

func TestFunctionEncodeCallArgCountMismatch(t *testing.T) {
	fn := NewFunction("transfer", Types{Address, Uint256}, Types{Bool})
	_, err := fn.EncodeCall()
	require.ErrorIs(t, err, ErrArgCountMismatch)
}

func TestFunctionEncodeCallNoArgs(t *testing.T) {
	fn := NewFunction("totalSupply", Types{}, Types{Uint256})
	got, err := fn.EncodeCall()
	require.NoError(t, err)
	sel := fn.Selector()
	require.Equal(t, sel[:], got)
}

func TestFunctionSelectorIsCached(t *testing.T) {
	fn := NewFunction("transfer", Types{Address, Uint256}, Types{Bool})
	first := fn.Selector()
	require.Equal(t, Selector{0xa9, 0x05, 0x9c, 0xbb}, first)

	fn.Name = "somethingElse"
	second := fn.Selector()
	require.Equal(t, first, second)
}
