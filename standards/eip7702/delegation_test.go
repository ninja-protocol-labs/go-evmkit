package eip7702

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

func TestEncodeDelegation(t *testing.T) {
	delegate := types.MustNewAddressFromHex("0x63c0c19a282a1b52b07dd5a65b58948a07dae32b")

	code, err := EncodeDelegation(delegate)
	require.NoError(t, err)
	require.Len(t, code, DelegationLength)
	require.Equal(t, "0xef010063c0c19a282a1b52b07dd5a65b58948a07dae32b", encoding.Hex.EncodePrefixed(code))
}

func TestEncodeDelegationRejectsInvalidDelegate(t *testing.T) {
	for _, tc := range []struct {
		name     string
		delegate *types.Address
	}{
		{"nil", nil},
		{"zero address", types.ZeroAddress()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			code, err := EncodeDelegation(tc.delegate)
			require.ErrorIs(t, err, ErrInvalidDelegate)
			require.EqualError(t, err, "eip7702: encodeDelegation: invalid delegate")
			require.Nil(t, code)
		})
	}
}

func TestParseDelegation(t *testing.T) {
	code, err := encoding.Hex.Decode("0xef010063c0c19a282a1b52b07dd5a65b58948a07dae32b")
	require.NoError(t, err)

	delegate, ok := ParseDelegation(code)
	require.True(t, ok)
	require.True(t, delegate.Equal(types.MustNewAddressFromHex("0x63c0c19a282a1b52b07dd5a65b58948a07dae32b")))
}

func TestParseDelegationZeroAddressDesignator(t *testing.T) {
	code, err := encoding.Hex.Decode("0xef01000000000000000000000000000000000000000000")
	require.NoError(t, err)

	delegate, ok := ParseDelegation(code)
	require.True(t, ok)
	require.True(t, delegate.IsZero())
}

func TestParseDelegationRejectsNonDesignator(t *testing.T) {
	for _, tc := range []struct {
		name string
		code string
	}{
		{"empty", "0x"},
		{"prefix only", "0xef0100"},
		{"22 bytes", "0xef010063c0c19a282a1b52b07dd5a65b58948a07dae3"},
		{"24 bytes", "0xef010063c0c19a282a1b52b07dd5a65b58948a07dae32b00"},
		{"wrong first byte", "0xee010063c0c19a282a1b52b07dd5a65b58948a07dae32b"},
		{"wrong version byte", "0xef010163c0c19a282a1b52b07dd5a65b58948a07dae32b"},
		{"eof container header", "0xef000163c0c19a282a1b52b07dd5a65b58948a07dae32b"},
		{"contract code", "0x6080604052348015600e575f5ffd5b50603e80601a5f395ff3fe60806040525f5ffdfea2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			code, err := encoding.Hex.Decode(tc.code)
			require.NoError(t, err)

			delegate, ok := ParseDelegation(code)
			require.False(t, ok)
			require.Nil(t, delegate)
		})
	}
}

func TestParseDelegationRejectsNil(t *testing.T) {
	delegate, ok := ParseDelegation(nil)
	require.False(t, ok)
	require.Nil(t, delegate)
}

func TestEncodeParseDelegationRoundTrip(t *testing.T) {
	delegate := types.MustNewAddressFromHex("0x000000000000000000000000000000000000dEaD")

	code, err := EncodeDelegation(delegate)
	require.NoError(t, err)

	got, ok := ParseDelegation(code)
	require.True(t, ok)
	require.True(t, got.Equal(delegate))
}

func TestParseDelegationDoesNotAliasInput(t *testing.T) {
	code, err := encoding.Hex.Decode("0xef010063c0c19a282a1b52b07dd5a65b58948a07dae32b")
	require.NoError(t, err)

	delegate, ok := ParseDelegation(code)
	require.True(t, ok)
	code[3] = 0xff
	require.True(t, delegate.Equal(types.MustNewAddressFromHex("0x63c0c19a282a1b52b07dd5a65b58948a07dae32b")))
}
