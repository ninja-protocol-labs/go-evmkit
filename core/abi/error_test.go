package abi

import (
	"encoding/hex"
	"math/big"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func mustDecodeHexRevert(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	require.NoError(t, err)
	return b
}

func TestDecodeRevertErrorString(t *testing.T) {
	data := mustDecodeHexRevert(t, "08c379a0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000174d756c746963616c6c333a2063616c6c206661696c6564000000000000000000")

	reason, err := DecodeRevert(data)
	require.NoError(t, err)
	require.Equal(t, "Multicall3: call failed", reason)
}

func TestDecodeRevertValueMismatch(t *testing.T) {
	data := mustDecodeHexRevert(t, "08c379a00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000001a4d756c746963616c6c333a2076616c7565206d69736d61746368000000000000")

	reason, err := DecodeRevert(data)
	require.NoError(t, err)
	require.Equal(t, "Multicall3: value mismatch", reason)
}

func TestDecodeRevertPanic(t *testing.T) {
	data := mustDecodeHexRevert(t, "4e487b710000000000000000000000000000000000000000000000000000000000000011")

	reason, err := DecodeRevert(data)
	require.NoError(t, err)
	require.Equal(t, "panic code 0x11", reason)
}

func TestDecodeRevertRejectsUnknownSelector(t *testing.T) {
	_, err := DecodeRevert([]byte{0xde, 0xad, 0xbe, 0xef})
	require.ErrorIs(t, err, ErrSelectorMismatch)
}

func TestDecodeRevertRejectsShortData(t *testing.T) {
	_, err := DecodeRevert([]byte{0x01, 0x02})
	require.ErrorIs(t, err, ErrByteLengthMismatch)
}

func TestParseErrorAndDecode(t *testing.T) {
	e, err := ParseError("InsufficientBalance(uint256 available, uint256 required)")
	require.NoError(t, err)
	require.Equal(t, "InsufficientBalance(uint256,uint256)", e.Signature())

	data, err := Pack(e.Inputs, big.NewInt(1), big.NewInt(100))
	require.NoError(t, err)
	sel := e.Selector()
	revertData := append(sel[:], data...)

	vals, err := e.Decode(revertData)
	require.NoError(t, err)
	require.Equal(t, 0, vals[0].(*big.Int).Cmp(big.NewInt(1)))
	require.Equal(t, 0, vals[1].(*big.Int).Cmp(big.NewInt(100)))
}

func TestParseErrorWithKeyword(t *testing.T) {
	e, err := ParseError("error InsufficientBalance(uint256 available, uint256 required)")
	require.NoError(t, err)
	require.Equal(t, "InsufficientBalance", e.Name)
	require.Equal(t, "InsufficientBalance(uint256,uint256)", e.Signature())
}

func TestParseErrorWithKeywordAndExtraSpace(t *testing.T) {
	e, err := ParseError("  error   Foo(uint256)  ")
	require.NoError(t, err)
	require.Equal(t, "Foo", e.Name)
}

func TestParseErrorNameStartingWithErrorIsNotStripped(t *testing.T) {
	e, err := ParseError("errorCode(uint256)")
	require.NoError(t, err)
	require.Equal(t, "errorCode", e.Name)
}

func TestParseErrorNoArgs(t *testing.T) {
	e, err := ParseError("Paused()")
	require.NoError(t, err)
	require.Equal(t, "Paused()", e.Signature())
	require.Empty(t, e.Inputs)
}

func TestParseErrorTupleAndArrayArgs(t *testing.T) {
	e, err := ParseError("BatchFailed((address,uint256)[] memory items)")
	require.NoError(t, err)
	require.Equal(t, "BatchFailed((address,uint256)[])", e.Signature())
}

func TestParseErrorRejectsMissingName(t *testing.T) {
	_, err := ParseError("(uint256)")
	require.ErrorIs(t, err, ErrInvalidTypeString)
}

func TestParseErrorRejectsMissingParen(t *testing.T) {
	_, err := ParseError("Foo")
	require.ErrorIs(t, err, ErrInvalidTypeString)
}

func TestParseErrorRejectsTrailingGarbage(t *testing.T) {
	_, err := ParseError("Foo(uint256) extra")
	require.ErrorIs(t, err, ErrInvalidTypeString)
}

func TestParseErrorRejectsMalformedParamList(t *testing.T) {
	tests := []string{"Foo(uint256", "Foo(uint256,)", "Foo(,uint256)", "Foo(uint256 a uint256 b)"}
	for _, sig := range tests {
		_, err := ParseError(sig)
		require.Error(t, err, sig)
	}
}

func TestErrorDecodeRejectsWrongSelector(t *testing.T) {
	e, err := ParseError("Foo(uint256)")
	require.NoError(t, err)

	_, err = e.Decode([]byte{0xde, 0xad, 0xbe, 0xef, 0x00})
	require.ErrorIs(t, err, ErrSelectorMismatch)
}

func TestErrorDecodeRejectsShortData(t *testing.T) {
	e, err := ParseError("Foo(uint256)")
	require.NoError(t, err)

	_, err = e.Decode([]byte{0x01, 0x02})
	require.ErrorIs(t, err, ErrByteLengthMismatch)
}

func TestErrorDecodeRejectsMalformedArgs(t *testing.T) {
	e, err := ParseError("Foo(uint256)")
	require.NoError(t, err)

	sel := e.Selector()
	_, err = e.Decode(sel[:])
	require.Error(t, err)
}

func TestErrorDecodeNoArgs(t *testing.T) {
	e, err := ParseError("Paused()")
	require.NoError(t, err)

	sel := e.Selector()
	vals, err := e.Decode(sel[:])
	require.NoError(t, err)
	require.Empty(t, vals)
}

func TestNewErrorAndParseErrorAgree(t *testing.T) {
	byHand := NewError("InsufficientBalance", Types{Uint256, Uint256})
	parsed, err := ParseError("InsufficientBalance(uint256,uint256)")
	require.NoError(t, err)

	require.Equal(t, byHand.Signature(), parsed.Signature())
	require.Equal(t, byHand.Selector(), parsed.Selector())
}

func TestErrorSelectorConcurrent(t *testing.T) {
	e := NewError("InsufficientBalance", Types{Uint256, Uint256})

	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			require.Equal(t, e.Selector(), e.Selector())
		})
	}
	wg.Wait()
}

func TestErrorSelectorKnownValue(t *testing.T) {
	e := NewError("Foo", Types{Uint256})
	fn := NewFunction("Foo", Types{Uint256}, nil)
	require.Equal(t, fn.Selector(), e.Selector())
}
