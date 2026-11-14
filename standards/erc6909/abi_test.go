package erc6909

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestEncodeBalanceOf(t *testing.T) {
	data, err := EncodeBalanceOf(testAddr1, big.NewInt(42))
	require.NoError(t, err)
	require.Equal(t, []byte{0x00, 0xfd, 0xd5, 0x8e}, data[:4])

	vals, err := balanceOfFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testAddr1))
	require.Zero(t, big.NewInt(42).Cmp(vals[1].(*big.Int)))
}

func TestEncodeBalanceOfRejectsNilArgs(t *testing.T) {
	_, err := EncodeBalanceOf(nil, big.NewInt(42))
	require.Error(t, err)
	_, err = EncodeBalanceOf(testAddr1, nil)
	require.Error(t, err)
}

func TestDecodeBalanceOf(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(42))
	require.NoError(t, err)
	n, err := DecodeBalanceOf(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(42).Cmp(n))
}

func TestDecodeBalanceOfRejectsGarbage(t *testing.T) {
	_, err := DecodeBalanceOf([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeAllowance(t *testing.T) {
	data, err := EncodeAllowance(testAddr1, testAddr2, big.NewInt(1))
	require.NoError(t, err)
	require.Equal(t, []byte{0x59, 0x8a, 0xf9, 0xe7}, data[:4])

	vals, err := allowanceFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testAddr1))
	require.True(t, vals[1].(*types.Address).Equal(testAddr2))
	require.Zero(t, big.NewInt(1).Cmp(vals[2].(*big.Int)))
}

func TestEncodeAllowanceRejectsNilArgs(t *testing.T) {
	_, err := EncodeAllowance(nil, testAddr2, big.NewInt(1))
	require.Error(t, err)
	_, err = EncodeAllowance(testAddr1, nil, big.NewInt(1))
	require.Error(t, err)
	_, err = EncodeAllowance(testAddr1, testAddr2, nil)
	require.Error(t, err)
}

func TestDecodeAllowance(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(100))
	require.NoError(t, err)
	n, err := DecodeAllowance(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(100).Cmp(n))
}

func TestDecodeAllowanceRejectsGarbage(t *testing.T) {
	_, err := DecodeAllowance([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeIsOperator(t *testing.T) {
	data, err := EncodeIsOperator(testAddr1, testAddr2)
	require.NoError(t, err)
	require.Equal(t, []byte{0xb6, 0x36, 0x3c, 0xf2}, data[:4])

	vals, err := isOperatorFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testAddr1))
	require.True(t, vals[1].(*types.Address).Equal(testAddr2))
}

func TestEncodeIsOperatorRejectsNilArgs(t *testing.T) {
	_, err := EncodeIsOperator(nil, testAddr2)
	require.Error(t, err)
	_, err = EncodeIsOperator(testAddr1, nil)
	require.Error(t, err)
}

func TestDecodeIsOperator(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Bool}, true)
	require.NoError(t, err)
	ok, err := DecodeIsOperator(data)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestDecodeIsOperatorRejectsGarbage(t *testing.T) {
	_, err := DecodeIsOperator([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeApprove(t *testing.T) {
	data, err := EncodeApprove(testAddr1, big.NewInt(1), big.NewInt(100))
	require.NoError(t, err)
	require.Equal(t, []byte{0x42, 0x6a, 0x84, 0x93}, data[:4])

	vals, err := approveFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testAddr1))
	require.Zero(t, big.NewInt(1).Cmp(vals[1].(*big.Int)))
	require.Zero(t, big.NewInt(100).Cmp(vals[2].(*big.Int)))
}

func TestEncodeApproveRejectsNilArgs(t *testing.T) {
	_, err := EncodeApprove(nil, big.NewInt(1), big.NewInt(100))
	require.Error(t, err)
	_, err = EncodeApprove(testAddr1, nil, big.NewInt(100))
	require.Error(t, err)
	_, err = EncodeApprove(testAddr1, big.NewInt(1), nil)
	require.Error(t, err)
}

func TestEncodeSetOperator(t *testing.T) {
	data, err := EncodeSetOperator(testAddr1, true)
	require.NoError(t, err)
	require.Equal(t, []byte{0x55, 0x8a, 0x72, 0x97}, data[:4])

	vals, err := setOperatorFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testAddr1))
	require.True(t, vals[1].(bool))
}

func TestEncodeSetOperatorRejectsNilArgs(t *testing.T) {
	_, err := EncodeSetOperator(nil, true)
	require.Error(t, err)
}

func TestEncodeTransfer(t *testing.T) {
	data, err := EncodeTransfer(testAddr1, big.NewInt(1), big.NewInt(100))
	require.NoError(t, err)
	require.Equal(t, []byte{0x09, 0x5b, 0xcd, 0xb6}, data[:4])

	vals, err := transferFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testAddr1))
	require.Zero(t, big.NewInt(1).Cmp(vals[1].(*big.Int)))
	require.Zero(t, big.NewInt(100).Cmp(vals[2].(*big.Int)))
}

func TestEncodeTransferRejectsNilArgs(t *testing.T) {
	_, err := EncodeTransfer(nil, big.NewInt(1), big.NewInt(100))
	require.Error(t, err)
	_, err = EncodeTransfer(testAddr1, nil, big.NewInt(100))
	require.Error(t, err)
	_, err = EncodeTransfer(testAddr1, big.NewInt(1), nil)
	require.Error(t, err)
}

func TestEncodeTransferFrom(t *testing.T) {
	data, err := EncodeTransferFrom(testAddr1, testAddr2, big.NewInt(1), big.NewInt(100))
	require.NoError(t, err)
	require.Equal(t, []byte{0xfe, 0x99, 0x04, 0x9a}, data[:4])

	vals, err := transferFromFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testAddr1))
	require.True(t, vals[1].(*types.Address).Equal(testAddr2))
	require.Zero(t, big.NewInt(1).Cmp(vals[2].(*big.Int)))
	require.Zero(t, big.NewInt(100).Cmp(vals[3].(*big.Int)))
}

func TestEncodeTransferFromRejectsNilArgs(t *testing.T) {
	_, err := EncodeTransferFrom(nil, testAddr2, big.NewInt(1), big.NewInt(100))
	require.Error(t, err)
	_, err = EncodeTransferFrom(testAddr1, nil, big.NewInt(1), big.NewInt(100))
	require.Error(t, err)
	_, err = EncodeTransferFrom(testAddr1, testAddr2, nil, big.NewInt(100))
	require.Error(t, err)
	_, err = EncodeTransferFrom(testAddr1, testAddr2, big.NewInt(1), nil)
	require.Error(t, err)
}

func TestEncodeName(t *testing.T) {
	data, err := EncodeName(big.NewInt(1))
	require.NoError(t, err)
	require.Equal(t, []byte{0x00, 0xad, 0x80, 0x0c}, data[:4])

	vals, err := nameFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1).Cmp(vals[0].(*big.Int)))
}

func TestEncodeNameRejectsNilArgs(t *testing.T) {
	_, err := EncodeName(nil)
	require.Error(t, err)
}

func TestDecodeName(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.String}, "evmkit test token")
	require.NoError(t, err)
	s, err := DecodeName(data)
	require.NoError(t, err)
	require.Equal(t, "evmkit test token", s)
}

func TestDecodeNameRejectsGarbage(t *testing.T) {
	_, err := DecodeName([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeSymbol(t *testing.T) {
	data, err := EncodeSymbol(big.NewInt(1))
	require.NoError(t, err)
	require.Equal(t, []byte{0x4e, 0x41, 0xa1, 0xfb}, data[:4])

	vals, err := symbolFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1).Cmp(vals[0].(*big.Int)))
}

func TestEncodeSymbolRejectsNilArgs(t *testing.T) {
	_, err := EncodeSymbol(nil)
	require.Error(t, err)
}

func TestDecodeSymbol(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.String}, "EVK6909")
	require.NoError(t, err)
	s, err := DecodeSymbol(data)
	require.NoError(t, err)
	require.Equal(t, "EVK6909", s)
}

func TestDecodeSymbolRejectsGarbage(t *testing.T) {
	_, err := DecodeSymbol([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecimals(t *testing.T) {
	data, err := EncodeDecimals(big.NewInt(1))
	require.NoError(t, err)
	require.Equal(t, []byte{0x3f, 0x47, 0xe6, 0x62}, data[:4])

	vals, err := decimalsFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1).Cmp(vals[0].(*big.Int)))
}

func TestEncodeDecimalsRejectsNilArgs(t *testing.T) {
	_, err := EncodeDecimals(nil)
	require.Error(t, err)
}

func TestDecodeDecimals(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint8}, uint8(18))
	require.NoError(t, err)
	d, err := DecodeDecimals(data)
	require.NoError(t, err)
	require.Equal(t, uint8(18), d)
}

func TestDecodeDecimalsRejectsGarbage(t *testing.T) {
	_, err := DecodeDecimals([]byte{0x01, 0x02})
	require.Error(t, err)
}
