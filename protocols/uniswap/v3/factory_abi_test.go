package v3

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

var (
	testTokenA = types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	testTokenB = types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	testPool   = types.MustNewAddressFromHex("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc")
	testOwner  = types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
)

func TestEncodeOwner(t *testing.T) {
	require.Equal(t, []byte{0x8d, 0xa5, 0xcb, 0x5b}, EncodeOwner())
}

func TestDecodeOwner(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testOwner)
	require.NoError(t, err)

	got, err := DecodeOwner(data)
	require.NoError(t, err)
	require.True(t, got.Equal(testOwner))
}

func TestDecodeOwnerRejectsGarbage(t *testing.T) {
	_, err := DecodeOwner([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeFeeAmountTickSpacing(t *testing.T) {
	data, err := EncodeFeeAmountTickSpacing(big.NewInt(3000))
	require.NoError(t, err)
	require.Equal(t, []byte{0x22, 0xaf, 0xcc, 0xcb}, data[:4])

	vals, err := feeAmountTickSpacingFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(3000).Cmp(vals[0].(*big.Int)))
}

func TestDecodeFeeAmountTickSpacing(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Int24}, big.NewInt(60))
	require.NoError(t, err)

	got, err := DecodeFeeAmountTickSpacing(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(60).Cmp(got))
}

func TestDecodeFeeAmountTickSpacingRejectsGarbage(t *testing.T) {
	_, err := DecodeFeeAmountTickSpacing([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeGetPool(t *testing.T) {
	data, err := EncodeGetPool(testTokenA, testTokenB, big.NewInt(500))
	require.NoError(t, err)
	require.Equal(t, []byte{0x16, 0x98, 0xee, 0x82}, data[:4])

	vals, err := getPoolFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testTokenA))
	require.True(t, vals[1].(*types.Address).Equal(testTokenB))
	require.Zero(t, big.NewInt(500).Cmp(vals[2].(*big.Int)))
}

func TestDecodeGetPool(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testPool)
	require.NoError(t, err)

	got, err := DecodeGetPool(data)
	require.NoError(t, err)
	require.True(t, got.Equal(testPool))
}

func TestDecodeGetPoolRejectsGarbage(t *testing.T) {
	_, err := DecodeGetPool([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeCreatePool(t *testing.T) {
	data, err := EncodeCreatePool(testTokenA, testTokenB, big.NewInt(10000))
	require.NoError(t, err)
	require.Equal(t, []byte{0xa1, 0x67, 0x12, 0x95}, data[:4])

	vals, err := createPoolFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testTokenA))
	require.True(t, vals[1].(*types.Address).Equal(testTokenB))
	require.Zero(t, big.NewInt(10000).Cmp(vals[2].(*big.Int)))
}

func TestDecodeCreatePool(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testPool)
	require.NoError(t, err)

	got, err := DecodeCreatePool(data)
	require.NoError(t, err)
	require.True(t, got.Equal(testPool))
}

func TestDecodeCreatePoolRejectsGarbage(t *testing.T) {
	_, err := DecodeCreatePool([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeSetOwner(t *testing.T) {
	data, err := EncodeSetOwner(testOwner)
	require.NoError(t, err)
	require.Equal(t, []byte{0x13, 0xaf, 0x40, 0x35}, data[:4])

	vals, err := setOwnerFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testOwner))
}

func TestEncodeEnableFeeAmount(t *testing.T) {
	data, err := EncodeEnableFeeAmount(big.NewInt(100), big.NewInt(1))
	require.NoError(t, err)
	require.Equal(t, []byte{0x8a, 0x7c, 0x19, 0x5f}, data[:4])

	vals, err := enableFeeAmountFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(100).Cmp(vals[0].(*big.Int)))
	require.Zero(t, big.NewInt(1).Cmp(vals[1].(*big.Int)))
}

func TestFeeAmountTickSpacingZeroFeeIsDisabled(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Int24}, big.NewInt(0))
	require.NoError(t, err)

	got, err := DecodeFeeAmountTickSpacing(data)
	require.NoError(t, err)
	require.Zero(t, got.Sign())
}

func TestEncodeFeeAmountTickSpacingRejectsUint24Overflow(t *testing.T) {
	overflow := new(big.Int).Lsh(big.NewInt(1), 24)
	_, err := EncodeFeeAmountTickSpacing(overflow)
	require.Error(t, err)
}

func TestEncodeFeeAmountTickSpacingAcceptsUint24Max(t *testing.T) {
	m := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 24), big.NewInt(1))
	data, err := EncodeFeeAmountTickSpacing(m)
	require.NoError(t, err)

	vals, err := feeAmountTickSpacingFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, m.Cmp(vals[0].(*big.Int)))
}

func TestEncodeEnableFeeAmountRejectsInt24Overflow(t *testing.T) {
	overflow := new(big.Int).Lsh(big.NewInt(1), 23)
	_, err := EncodeEnableFeeAmount(big.NewInt(1), overflow)
	require.Error(t, err)
}

func TestEncodeEnableFeeAmountRejectsInt24Underflow(t *testing.T) {
	underflow := new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 23))
	underflow.Sub(underflow, big.NewInt(1))
	_, err := EncodeEnableFeeAmount(big.NewInt(1), underflow)
	require.Error(t, err)
}

func TestDecodeFeeAmountTickSpacingNegativeTickSpacing(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Int24}, big.NewInt(-60))
	require.NoError(t, err)

	got, err := DecodeFeeAmountTickSpacing(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(-60).Cmp(got))
}

func TestEncodeGetPoolDistinguishesTokenOrder(t *testing.T) {
	dataAB, err := EncodeGetPool(testTokenA, testTokenB, big.NewInt(500))
	require.NoError(t, err)
	dataBA, err := EncodeGetPool(testTokenB, testTokenA, big.NewInt(500))
	require.NoError(t, err)
	require.NotEqual(t, dataAB, dataBA)
}

func TestEncodeCreatePoolRejectsNilTokens(t *testing.T) {
	_, err := EncodeCreatePool(nil, testTokenB, big.NewInt(500))
	require.Error(t, err)
}
