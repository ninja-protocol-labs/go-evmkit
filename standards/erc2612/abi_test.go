package erc2612

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

var (
	testOwner   = types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	testSpender = types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
)

func TestEncodePermit(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	sig, err := key.Sign(types.NewHashFromBytes([]byte("some digest")))
	require.NoError(t, err)
	value := big.NewInt(100)
	deadline := big.NewInt(9999999999)

	data, err := EncodePermit(testOwner, testSpender, value, deadline, sig)
	require.NoError(t, err)
	require.Equal(t, []byte{0xd5, 0x05, 0xac, 0xcf}, data[:4])

	vals, err := permitFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testOwner))
	require.True(t, vals[1].(*types.Address).Equal(testSpender))
	require.Zero(t, value.Cmp(vals[2].(*big.Int)))
	require.Zero(t, deadline.Cmp(vals[3].(*big.Int)))

	wantV := uint8(sig.LegacyV().Uint64())
	require.Equal(t, wantV, vals[4].(uint8))

	wantR := make([]byte, 32)
	sig.R().FillBytes(wantR)
	require.Equal(t, wantR, vals[5].([]byte))

	wantS := make([]byte, 32)
	sig.S().FillBytes(wantS)
	require.Equal(t, wantS, vals[6].([]byte))
}

func TestEncodePermitUsesLegacyV(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	sig, err := key.Sign(types.NewHashFromBytes([]byte("some digest")))
	require.NoError(t, err)

	data, err := EncodePermit(testOwner, testSpender, big.NewInt(1), big.NewInt(1), sig)
	require.NoError(t, err)

	vals, err := permitFn.DecodeCall(data)
	require.NoError(t, err)

	v := vals[4].(uint8)
	require.True(t, v == 27 || v == 28, "v must be legacy-encoded (27/28), got %d", v)
}

func TestEncodeNonces(t *testing.T) {
	data, err := EncodeNonces(testOwner)
	require.NoError(t, err)
	require.Equal(t, []byte{0x7e, 0xce, 0xbe, 0x00}, data[:4])

	vals, err := noncesFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testOwner))
}

func TestDecodeNonces(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(42))
	require.NoError(t, err)

	n, err := DecodeNonces(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(42).Cmp(n))
}

func TestDecodeNoncesRejectsGarbage(t *testing.T) {
	_, err := DecodeNonces([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDomainSeparator(t *testing.T) {
	data := EncodeDomainSeparator()
	require.Equal(t, []byte{0x36, 0x44, 0xe5, 0x15}, data)
}

func TestDecodeDomainSeparator(t *testing.T) {
	want := make([]byte, 32)
	want[31] = 0xab
	data, err := abi.Pack(abi.Types{abi.Bytes32}, want)
	require.NoError(t, err)

	got, err := DecodeDomainSeparator(data)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDecodeDomainSeparatorRejectsGarbage(t *testing.T) {
	_, err := DecodeDomainSeparator([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestCustomErrorSignatures(t *testing.T) {
	tests := []struct {
		name string
		err  *abi.Error
		sig  string
	}{
		{"ERC2612ExpiredSignature", errExpiredSignature, "ERC2612ExpiredSignature(uint256)"},
		{"ERC2612InvalidSigner", errInvalidSigner, "ERC2612InvalidSigner(address,address)"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.sig, tc.err.Signature())

			parsed, err := abi.ParseError(tc.sig)
			require.NoError(t, err)
			require.Equal(t, tc.err.Selector(), parsed.Selector())
		})
	}
}
