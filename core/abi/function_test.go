package abi

import (
	"math/big"
	"sync"
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

func TestFunctionDecodeReturn(t *testing.T) {
	fn := NewFunction("balanceOf", Types{Address}, Types{Uint256})

	returnData, err := Pack(fn.Outputs, big.NewInt(42))
	require.NoError(t, err)

	got, err := fn.DecodeReturn(returnData)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Zero(t, big.NewInt(42).Cmp(got[0].(*big.Int)))
}

func TestFunctionDecodeReturnMultipleOutputs(t *testing.T) {
	fn := NewFunction("getReserves", Types{}, Types{Uint256, Uint256})

	returnData, err := Pack(fn.Outputs, big.NewInt(100), big.NewInt(200))
	require.NoError(t, err)

	got, err := fn.DecodeReturn(returnData)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(100).Cmp(got[0].(*big.Int)))
	require.Zero(t, big.NewInt(200).Cmp(got[1].(*big.Int)))
}

func TestFunctionDecodeReturnNoOutputs(t *testing.T) {
	fn := NewFunction("doSomething", Types{}, Types{})
	got, err := fn.DecodeReturn(nil)
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestFunctionDecodeReturnError(t *testing.T) {
	fn := NewFunction("balanceOf", Types{Address}, Types{Uint256})
	_, err := fn.DecodeReturn(make([]byte, 16))
	require.ErrorIs(t, err, ErrByteLengthMismatch)
}

func TestFunctionEncodeCallDecodeReturnFullRoundTrip(t *testing.T) {
	addr, err := types.NewAddressFromHex("0x833e1D0b8Bc979D49d57b65dCF18364694B16D52")
	require.NoError(t, err)
	amount := big.NewInt(1000000000000000000)

	transfer := NewFunction("transfer", Types{Address, Uint256}, Types{Bool})
	calldata, err := transfer.EncodeCall(addr, amount)
	require.NoError(t, err)
	require.Equal(t, []byte{0xa9, 0x05, 0x9c, 0xbb}, calldata[:4])

	decodedArgs, err := transfer.DecodeCall(calldata)
	require.NoError(t, err)
	require.Equal(t, addr, decodedArgs[0])
	require.Zero(t, amount.Cmp(decodedArgs[1].(*big.Int)))

	returnData, err := transfer.EncodeReturn(true)
	require.NoError(t, err)
	decodedReturn, err := transfer.DecodeReturn(returnData)
	require.NoError(t, err)
	require.Equal(t, true, decodedReturn[0])
}

func TestFunctionEncodeReturnMatchesPack(t *testing.T) {
	fn := NewFunction("balanceOf", Types{Address}, Types{Uint256})

	got, err := fn.EncodeReturn(big.NewInt(42))
	require.NoError(t, err)

	want, err := Pack(fn.Outputs, big.NewInt(42))
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestFunctionEncodeReturnDecodeReturnRoundTrip(t *testing.T) {
	fn := NewFunction("getReserves", Types{}, Types{Uint256, Uint256})

	data, err := fn.EncodeReturn(big.NewInt(100), big.NewInt(200))
	require.NoError(t, err)

	got, err := fn.DecodeReturn(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(100).Cmp(got[0].(*big.Int)))
	require.Zero(t, big.NewInt(200).Cmp(got[1].(*big.Int)))
}

func TestFunctionEncodeReturnArgCountMismatch(t *testing.T) {
	fn := NewFunction("balanceOf", Types{Address}, Types{Uint256})
	_, err := fn.EncodeReturn()
	require.ErrorIs(t, err, ErrArgCountMismatch)
}

func TestFunctionEncodeReturnNoOutputs(t *testing.T) {
	fn := NewFunction("doSomething", Types{}, Types{})
	got, err := fn.EncodeReturn()
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestFunctionDecodeCallWrongSelectorRejected(t *testing.T) {
	transfer := NewFunction("transfer", Types{Address, Uint256}, Types{Bool})
	approve := NewFunction("approve", Types{Address, Uint256}, Types{Bool})

	addr, err := types.NewAddressFromHex("0x833e1D0b8Bc979D49d57b65dCF18364694B16D52")
	require.NoError(t, err)
	calldata, err := approve.EncodeCall(addr, big.NewInt(1))
	require.NoError(t, err)

	_, err = transfer.DecodeCall(calldata)
	require.ErrorIs(t, err, ErrSelectorMismatch)
}

func TestFunctionDecodeCallTooShortRejected(t *testing.T) {
	fn := NewFunction("transfer", Types{Address, Uint256}, Types{Bool})
	_, err := fn.DecodeCall([]byte{0x01, 0x02})
	require.ErrorIs(t, err, ErrByteLengthMismatch)
}

func TestFunctionDecodeCallNoArgs(t *testing.T) {
	fn := NewFunction("totalSupply", Types{}, Types{Uint256})
	calldata, err := fn.EncodeCall()
	require.NoError(t, err)

	got, err := fn.DecodeCall(calldata)
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestFunctionSelectorConcurrent(t *testing.T) {
	fn := NewFunction("transfer", Types{Address, Uint256}, Types{Bool})
	want := Selector{0xa9, 0x05, 0x9c, 0xbb}

	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			require.Equal(t, want, fn.Selector())
		})
	}
	wg.Wait()
}
