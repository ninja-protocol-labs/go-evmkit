package v3

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

var (
	testQuoterTokenIn     = types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	testQuoterTokenOut    = types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	testQuoterFactoryAddr = types.MustNewAddressFromHex("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc")
)

func TestEncodeQuoterFactory(t *testing.T) {
	require.Equal(t, []byte{0xc4, 0x5a, 0x01, 0x55}, EncodeQuoterFactory())
}

func TestDecodeQuoterFactory(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testQuoterFactoryAddr)
	require.NoError(t, err)

	got, err := DecodeQuoterFactory(data)
	require.NoError(t, err)
	require.True(t, got.Equal(testQuoterFactoryAddr))
}

func TestDecodeQuoterFactoryRejectsGarbage(t *testing.T) {
	_, err := DecodeQuoterFactory([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeQuoterWETH9(t *testing.T) {
	require.Equal(t, []byte{0x4a, 0xa4, 0xa4, 0xfc}, EncodeQuoterWETH9())
}

func TestDecodeQuoterWETH9(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testQuoterTokenIn)
	require.NoError(t, err)

	got, err := DecodeQuoterWETH9(data)
	require.NoError(t, err)
	require.True(t, got.Equal(testQuoterTokenIn))
}

func TestDecodeQuoterWETH9RejectsGarbage(t *testing.T) {
	_, err := DecodeQuoterWETH9([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeQuoteExactInputSingle(t *testing.T) {
	data, err := EncodeQuoteExactInputSingle(testQuoterTokenIn, testQuoterTokenOut, big.NewInt(1000000), big.NewInt(3000), big.NewInt(0))
	require.NoError(t, err)
	require.Equal(t, []byte{0xc6, 0xa5, 0x02, 0x6a}, data[:4])

	vals, err := quoteExactInputSingleFn.DecodeCall(data)
	require.NoError(t, err)
	require.Len(t, vals, 1)
	fields := vals[0].([]any)
	require.Len(t, fields, 5)
	require.True(t, fields[0].(*types.Address).Equal(testQuoterTokenIn))
	require.True(t, fields[1].(*types.Address).Equal(testQuoterTokenOut))
	require.Zero(t, big.NewInt(1000000).Cmp(fields[2].(*big.Int)))
	require.Zero(t, big.NewInt(3000).Cmp(fields[3].(*big.Int)))
	require.Zero(t, big.NewInt(0).Cmp(fields[4].(*big.Int)))
}

func TestEncodeQuoteExactInputSingleRejectsNilTokenIn(t *testing.T) {
	_, err := EncodeQuoteExactInputSingle(nil, testQuoterTokenOut, big.NewInt(1000000), big.NewInt(3000), big.NewInt(0))
	require.Error(t, err)
}

func TestEncodeQuoteExactInputSingleRejectsUint24Overflow(t *testing.T) {
	overflow := new(big.Int).Lsh(big.NewInt(1), 24)
	_, err := EncodeQuoteExactInputSingle(testQuoterTokenIn, testQuoterTokenOut, big.NewInt(1000000), overflow, big.NewInt(0))
	require.Error(t, err)
}

func TestDecodeQuoteExactInputSingle(t *testing.T) {
	data, err := quoteExactInputSingleFn.EncodeReturn(big.NewInt(42), big.NewInt(100), uint32(2), big.NewInt(50000))
	require.NoError(t, err)

	got, err := DecodeQuoteExactInputSingle(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(42).Cmp(got.AmountOut))
	require.Zero(t, big.NewInt(100).Cmp(got.SqrtPriceX96After))
	require.Zero(t, big.NewInt(2).Cmp(got.InitializedTicksCrossed))
	require.Zero(t, big.NewInt(50000).Cmp(got.GasEstimate))
}

func TestDecodeQuoteExactInputSingleRejectsGarbage(t *testing.T) {
	_, err := DecodeQuoteExactInputSingle([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeQuoteExactInput(t *testing.T) {
	path := append(append(testQuoterTokenIn.Bytes(), 0x00, 0x0b, 0xb8), testQuoterTokenOut.Bytes()...)
	data, err := EncodeQuoteExactInput(path, big.NewInt(1000000))
	require.NoError(t, err)
	require.Equal(t, []byte{0xcd, 0xca, 0x17, 0x53}, data[:4])

	vals, err := quoteExactInputFn.DecodeCall(data)
	require.NoError(t, err)
	require.Len(t, vals, 2)
	require.Equal(t, path, vals[0].([]byte))
	require.Zero(t, big.NewInt(1000000).Cmp(vals[1].(*big.Int)))
}

func TestEncodeQuoteExactInputEmptyPath(t *testing.T) {
	data, err := EncodeQuoteExactInput([]byte{}, big.NewInt(1000000))
	require.NoError(t, err)

	vals, err := quoteExactInputFn.DecodeCall(data)
	require.NoError(t, err)
	require.Empty(t, vals[0].([]byte))
}

func TestDecodeQuoteExactInput(t *testing.T) {
	data, err := quoteExactInputFn.EncodeReturn(big.NewInt(42), []*big.Int{big.NewInt(100), big.NewInt(101)}, []uint32{1, 2}, big.NewInt(70000))
	require.NoError(t, err)

	got, err := DecodeQuoteExactInput(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(42).Cmp(got.AmountOut))
	require.Len(t, got.SqrtPriceX96AfterList, 2)
	require.Zero(t, big.NewInt(100).Cmp(got.SqrtPriceX96AfterList[0]))
	require.Zero(t, big.NewInt(101).Cmp(got.SqrtPriceX96AfterList[1]))
	require.Len(t, got.InitializedTicksCrossedList, 2)
	require.Zero(t, big.NewInt(1).Cmp(got.InitializedTicksCrossedList[0]))
	require.Zero(t, big.NewInt(2).Cmp(got.InitializedTicksCrossedList[1]))
	require.Zero(t, big.NewInt(70000).Cmp(got.GasEstimate))
}

func TestDecodeQuoteExactInputEmptyLists(t *testing.T) {
	data, err := quoteExactInputFn.EncodeReturn(big.NewInt(42), []*big.Int{}, []uint32{}, big.NewInt(70000))
	require.NoError(t, err)

	got, err := DecodeQuoteExactInput(data)
	require.NoError(t, err)
	require.Empty(t, got.SqrtPriceX96AfterList)
	require.Empty(t, got.InitializedTicksCrossedList)
}

func TestDecodeQuoteExactInputRejectsGarbage(t *testing.T) {
	_, err := DecodeQuoteExactInput([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeQuoteExactOutputSingle(t *testing.T) {
	data, err := EncodeQuoteExactOutputSingle(testQuoterTokenIn, testQuoterTokenOut, big.NewInt(1000000), big.NewInt(3000), big.NewInt(0))
	require.NoError(t, err)
	require.Equal(t, []byte{0xbd, 0x21, 0x70, 0x4a}, data[:4])

	vals, err := quoteExactOutputSingleFn.DecodeCall(data)
	require.NoError(t, err)
	fields := vals[0].([]any)
	require.Len(t, fields, 5)
	require.True(t, fields[0].(*types.Address).Equal(testQuoterTokenIn))
	require.True(t, fields[1].(*types.Address).Equal(testQuoterTokenOut))
	require.Zero(t, big.NewInt(1000000).Cmp(fields[2].(*big.Int)))
	require.Zero(t, big.NewInt(3000).Cmp(fields[3].(*big.Int)))
	require.Zero(t, big.NewInt(0).Cmp(fields[4].(*big.Int)))
}

func TestEncodeQuoteExactOutputSingleRejectsNilRecipient(t *testing.T) {
	_, err := EncodeQuoteExactOutputSingle(testQuoterTokenIn, nil, big.NewInt(1000000), big.NewInt(3000), big.NewInt(0))
	require.Error(t, err)
}

func TestDecodeQuoteExactOutputSingle(t *testing.T) {
	data, err := quoteExactOutputSingleFn.EncodeReturn(big.NewInt(7), big.NewInt(100), uint32(2), big.NewInt(50000))
	require.NoError(t, err)

	got, err := DecodeQuoteExactOutputSingle(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(7).Cmp(got.AmountIn))
	require.Zero(t, big.NewInt(100).Cmp(got.SqrtPriceX96After))
	require.Zero(t, big.NewInt(2).Cmp(got.InitializedTicksCrossed))
	require.Zero(t, big.NewInt(50000).Cmp(got.GasEstimate))
}

func TestDecodeQuoteExactOutputSingleRejectsGarbage(t *testing.T) {
	_, err := DecodeQuoteExactOutputSingle([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeQuoteExactOutput(t *testing.T) {
	path := append(append(testQuoterTokenOut.Bytes(), 0x00, 0x0b, 0xb8), testQuoterTokenIn.Bytes()...)
	data, err := EncodeQuoteExactOutput(path, big.NewInt(1000000))
	require.NoError(t, err)
	require.Equal(t, []byte{0x2f, 0x80, 0xbb, 0x1d}, data[:4])

	vals, err := quoteExactOutputFn.DecodeCall(data)
	require.NoError(t, err)
	require.Equal(t, path, vals[0].([]byte))
	require.Zero(t, big.NewInt(1000000).Cmp(vals[1].(*big.Int)))
}

func TestDecodeQuoteExactOutput(t *testing.T) {
	data, err := quoteExactOutputFn.EncodeReturn(big.NewInt(13), []*big.Int{big.NewInt(100)}, []uint32{1}, big.NewInt(70000))
	require.NoError(t, err)

	got, err := DecodeQuoteExactOutput(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(13).Cmp(got.AmountIn))
	require.Len(t, got.SqrtPriceX96AfterList, 1)
	require.Zero(t, big.NewInt(100).Cmp(got.SqrtPriceX96AfterList[0]))
	require.Len(t, got.InitializedTicksCrossedList, 1)
	require.Zero(t, big.NewInt(1).Cmp(got.InitializedTicksCrossedList[0]))
	require.Zero(t, big.NewInt(70000).Cmp(got.GasEstimate))
}

func TestDecodeQuoteExactOutputRejectsGarbage(t *testing.T) {
	_, err := DecodeQuoteExactOutput([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeQuoteExactInputDistinguishesPathDirection(t *testing.T) {
	pathForward := append(append(testQuoterTokenIn.Bytes(), 0x00, 0x0b, 0xb8), testQuoterTokenOut.Bytes()...)
	pathBackward := append(append(testQuoterTokenOut.Bytes(), 0x00, 0x0b, 0xb8), testQuoterTokenIn.Bytes()...)

	dataForward, err := EncodeQuoteExactInput(pathForward, big.NewInt(1000000))
	require.NoError(t, err)
	dataBackward, err := EncodeQuoteExactInput(pathBackward, big.NewInt(1000000))
	require.NoError(t, err)
	require.NotEqual(t, dataForward, dataBackward)
}
