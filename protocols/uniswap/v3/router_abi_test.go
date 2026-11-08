package v3

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

var (
	testRouterTokenIn     = types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	testRouterTokenOut    = types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	testRouterRecipient   = types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	testRouterFeeAddr     = types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222")
	testRouterFactoryAddr = types.MustNewAddressFromHex("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc")
)

func TestEncodeRouterFactory(t *testing.T) {
	require.Equal(t, []byte{0xc4, 0x5a, 0x01, 0x55}, EncodeRouterFactory())
}

func TestDecodeRouterFactory(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testRouterFactoryAddr)
	require.NoError(t, err)

	got, err := DecodeRouterFactory(data)
	require.NoError(t, err)
	require.True(t, got.Equal(testRouterFactoryAddr))
}

func TestDecodeRouterFactoryRejectsGarbage(t *testing.T) {
	_, err := DecodeRouterFactory([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeRouterWETH9(t *testing.T) {
	require.Equal(t, []byte{0x4a, 0xa4, 0xa4, 0xfc}, EncodeRouterWETH9())
}

func TestDecodeRouterWETH9(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testRouterTokenIn)
	require.NoError(t, err)

	got, err := DecodeRouterWETH9(data)
	require.NoError(t, err)
	require.True(t, got.Equal(testRouterTokenIn))
}

func TestDecodeRouterWETH9RejectsGarbage(t *testing.T) {
	_, err := DecodeRouterWETH9([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeExactInputSingle(t *testing.T) {
	data, err := EncodeExactInputSingle(testRouterTokenIn, testRouterTokenOut, big.NewInt(3000), testRouterRecipient, big.NewInt(1000000), big.NewInt(1), big.NewInt(0))
	require.NoError(t, err)
	require.Equal(t, []byte{0x04, 0xe4, 0x5a, 0xaf}, data[:4])

	vals, err := exactInputSingleFn.DecodeCall(data)
	require.NoError(t, err)
	require.Len(t, vals, 1)
	fields := vals[0].([]any)
	require.Len(t, fields, 7)
	require.True(t, fields[0].(*types.Address).Equal(testRouterTokenIn))
	require.True(t, fields[1].(*types.Address).Equal(testRouterTokenOut))
	require.Zero(t, big.NewInt(3000).Cmp(fields[2].(*big.Int)))
	require.True(t, fields[3].(*types.Address).Equal(testRouterRecipient))
	require.Zero(t, big.NewInt(1000000).Cmp(fields[4].(*big.Int)))
	require.Zero(t, big.NewInt(1).Cmp(fields[5].(*big.Int)))
	require.Zero(t, big.NewInt(0).Cmp(fields[6].(*big.Int)))
}

func TestDecodeExactInputSingle(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(42))
	require.NoError(t, err)

	got, err := DecodeExactInputSingle(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(42).Cmp(got))
}

func TestDecodeExactInputSingleRejectsGarbage(t *testing.T) {
	_, err := DecodeExactInputSingle([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeExactInputSingleRejectsNilTokenIn(t *testing.T) {
	_, err := EncodeExactInputSingle(nil, testRouterTokenOut, big.NewInt(3000), testRouterRecipient, big.NewInt(1000000), big.NewInt(1), big.NewInt(0))
	require.Error(t, err)
}

func TestEncodeExactInputSingleRejectsUint24Overflow(t *testing.T) {
	overflow := new(big.Int).Lsh(big.NewInt(1), 24)
	_, err := EncodeExactInputSingle(testRouterTokenIn, testRouterTokenOut, overflow, testRouterRecipient, big.NewInt(1000000), big.NewInt(1), big.NewInt(0))
	require.Error(t, err)
}

func TestEncodeExactInput(t *testing.T) {
	path := append(append(testRouterTokenIn.Bytes(), 0x00, 0x0b, 0xb8), testRouterTokenOut.Bytes()...)
	data, err := EncodeExactInput(path, testRouterRecipient, big.NewInt(1000000), big.NewInt(1))
	require.NoError(t, err)
	require.Equal(t, []byte{0xb8, 0x58, 0x18, 0x3f}, data[:4])

	vals, err := exactInputFn.DecodeCall(data)
	require.NoError(t, err)
	fields := vals[0].([]any)
	require.Len(t, fields, 4)
	require.Equal(t, path, fields[0].([]byte))
	require.True(t, fields[1].(*types.Address).Equal(testRouterRecipient))
	require.Zero(t, big.NewInt(1000000).Cmp(fields[2].(*big.Int)))
	require.Zero(t, big.NewInt(1).Cmp(fields[3].(*big.Int)))
}

func TestEncodeExactInputEmptyPath(t *testing.T) {
	data, err := EncodeExactInput([]byte{}, testRouterRecipient, big.NewInt(1000000), big.NewInt(1))
	require.NoError(t, err)

	vals, err := exactInputFn.DecodeCall(data)
	require.NoError(t, err)
	fields := vals[0].([]any)
	require.Empty(t, fields[0].([]byte))
}

func TestDecodeExactInput(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(99))
	require.NoError(t, err)

	got, err := DecodeExactInput(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(99).Cmp(got))
}

func TestDecodeExactInputRejectsGarbage(t *testing.T) {
	_, err := DecodeExactInput([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeExactOutputSingle(t *testing.T) {
	data, err := EncodeExactOutputSingle(testRouterTokenIn, testRouterTokenOut, big.NewInt(500), testRouterRecipient, big.NewInt(1000000), big.NewInt(2000000), big.NewInt(0))
	require.NoError(t, err)
	require.Equal(t, []byte{0x50, 0x23, 0xb4, 0xdf}, data[:4])

	vals, err := exactOutputSingleFn.DecodeCall(data)
	require.NoError(t, err)
	fields := vals[0].([]any)
	require.Len(t, fields, 7)
	require.True(t, fields[0].(*types.Address).Equal(testRouterTokenIn))
	require.True(t, fields[1].(*types.Address).Equal(testRouterTokenOut))
	require.Zero(t, big.NewInt(500).Cmp(fields[2].(*big.Int)))
	require.True(t, fields[3].(*types.Address).Equal(testRouterRecipient))
	require.Zero(t, big.NewInt(1000000).Cmp(fields[4].(*big.Int)))
	require.Zero(t, big.NewInt(2000000).Cmp(fields[5].(*big.Int)))
	require.Zero(t, big.NewInt(0).Cmp(fields[6].(*big.Int)))
}

func TestDecodeExactOutputSingle(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(7))
	require.NoError(t, err)

	got, err := DecodeExactOutputSingle(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(7).Cmp(got))
}

func TestDecodeExactOutputSingleRejectsGarbage(t *testing.T) {
	_, err := DecodeExactOutputSingle([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeExactOutput(t *testing.T) {
	path := append(append(testRouterTokenOut.Bytes(), 0x00, 0x0b, 0xb8), testRouterTokenIn.Bytes()...)
	data, err := EncodeExactOutput(path, testRouterRecipient, big.NewInt(1000000), big.NewInt(2000000))
	require.NoError(t, err)
	require.Equal(t, []byte{0x09, 0xb8, 0x13, 0x46}, data[:4])

	vals, err := exactOutputFn.DecodeCall(data)
	require.NoError(t, err)
	fields := vals[0].([]any)
	require.Len(t, fields, 4)
	require.Equal(t, path, fields[0].([]byte))
	require.True(t, fields[1].(*types.Address).Equal(testRouterRecipient))
	require.Zero(t, big.NewInt(1000000).Cmp(fields[2].(*big.Int)))
	require.Zero(t, big.NewInt(2000000).Cmp(fields[3].(*big.Int)))
}

func TestDecodeExactOutput(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(13))
	require.NoError(t, err)

	got, err := DecodeExactOutput(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(13).Cmp(got))
}

func TestDecodeExactOutputRejectsGarbage(t *testing.T) {
	_, err := DecodeExactOutput([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeRouterUnwrapWETH9(t *testing.T) {
	data, err := EncodeRouterUnwrapWETH9(big.NewInt(0), testRouterRecipient)
	require.NoError(t, err)
	require.Equal(t, []byte{0x49, 0x40, 0x4b, 0x7c}, data[:4])

	vals, err := routerUnwrapWETH9Fn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(0).Cmp(vals[0].(*big.Int)))
	require.True(t, vals[1].(*types.Address).Equal(testRouterRecipient))
}

func TestEncodeUnwrapWETH9WithFee(t *testing.T) {
	data, err := EncodeUnwrapWETH9WithFee(big.NewInt(0), testRouterRecipient, big.NewInt(50), testRouterFeeAddr)
	require.NoError(t, err)
	require.Equal(t, []byte{0x9b, 0x2c, 0x0a, 0x37}, data[:4])

	vals, err := unwrapWETH9WithFeeFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(0).Cmp(vals[0].(*big.Int)))
	require.True(t, vals[1].(*types.Address).Equal(testRouterRecipient))
	require.Zero(t, big.NewInt(50).Cmp(vals[2].(*big.Int)))
	require.True(t, vals[3].(*types.Address).Equal(testRouterFeeAddr))
}

func TestEncodeUnwrapWETH9WithFeeRejectsNilFeeRecipient(t *testing.T) {
	_, err := EncodeUnwrapWETH9WithFee(big.NewInt(0), testRouterRecipient, big.NewInt(50), nil)
	require.Error(t, err)
}

func TestEncodeRouterSweepToken(t *testing.T) {
	data, err := EncodeRouterSweepToken(testRouterTokenIn, big.NewInt(0), testRouterRecipient)
	require.NoError(t, err)
	require.Equal(t, []byte{0xdf, 0x2a, 0xb5, 0xbb}, data[:4])

	vals, err := routerSweepTokenFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testRouterTokenIn))
	require.Zero(t, big.NewInt(0).Cmp(vals[1].(*big.Int)))
	require.True(t, vals[2].(*types.Address).Equal(testRouterRecipient))
}

func TestEncodeSweepTokenWithFee(t *testing.T) {
	data, err := EncodeSweepTokenWithFee(testRouterTokenIn, big.NewInt(0), testRouterRecipient, big.NewInt(100), testRouterFeeAddr)
	require.NoError(t, err)
	require.Equal(t, []byte{0xe0, 0xe1, 0x89, 0xa0}, data[:4])

	vals, err := sweepTokenWithFeeFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testRouterTokenIn))
	require.Zero(t, big.NewInt(0).Cmp(vals[1].(*big.Int)))
	require.True(t, vals[2].(*types.Address).Equal(testRouterRecipient))
	require.Zero(t, big.NewInt(100).Cmp(vals[3].(*big.Int)))
	require.True(t, vals[4].(*types.Address).Equal(testRouterFeeAddr))
}

func TestEncodeRouterRefundETH(t *testing.T) {
	require.Equal(t, []byte{0x12, 0x21, 0x0e, 0x8a}, EncodeRouterRefundETH())
}

func TestEncodeRouterMulticall(t *testing.T) {
	data, err := EncodeRouterMulticall([][]byte{EncodeRouterRefundETH(), {0x01, 0x02, 0x03}})
	require.NoError(t, err)
	require.Equal(t, []byte{0xac, 0x96, 0x50, 0xd8}, data[:4])

	vals, err := routerMulticallFn.DecodeCall(data)
	require.NoError(t, err)
	elems := vals[0].([]any)
	require.Len(t, elems, 2)
	require.Equal(t, EncodeRouterRefundETH(), elems[0].([]byte))
	require.Equal(t, []byte{0x01, 0x02, 0x03}, elems[1].([]byte))
}

func TestEncodeRouterMulticallEmpty(t *testing.T) {
	data, err := EncodeRouterMulticall([][]byte{})
	require.NoError(t, err)

	vals, err := routerMulticallFn.DecodeCall(data)
	require.NoError(t, err)
	require.Empty(t, vals[0].([]any))
}

func TestDecodeRouterMulticall(t *testing.T) {
	data, err := routerMulticallFn.EncodeReturn([][]byte{{0xaa, 0xbb}, {0xcc}})
	require.NoError(t, err)

	got, err := DecodeRouterMulticall(data)
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, []byte{0xaa, 0xbb}, got[0])
	require.Equal(t, []byte{0xcc}, got[1])
}

func TestDecodeRouterMulticallEmpty(t *testing.T) {
	data, err := routerMulticallFn.EncodeReturn([][]byte{})
	require.NoError(t, err)

	got, err := DecodeRouterMulticall(data)
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestDecodeRouterMulticallRejectsGarbage(t *testing.T) {
	_, err := DecodeRouterMulticall([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeSelfPermit(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	sig, err := key.Sign(types.NewHashFromBytes([]byte("some 32 byte digest padded out!")))
	require.NoError(t, err)

	data, err := EncodeSelfPermit(testRouterTokenIn, big.NewInt(1000), big.NewInt(1700000000), sig)
	require.NoError(t, err)
	require.Equal(t, []byte{0xf3, 0x99, 0x5c, 0x67}, data[:4])

	vals, err := selfPermitFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testRouterTokenIn))
	require.Zero(t, big.NewInt(1000).Cmp(vals[1].(*big.Int)))
	require.Zero(t, big.NewInt(1700000000).Cmp(vals[2].(*big.Int)))
	legacy := sig.LegacyBytes()
	require.Equal(t, legacy[64], vals[3].(uint8))
	require.Equal(t, legacy[:32], vals[4].([]byte))
	require.Equal(t, legacy[32:64], vals[5].([]byte))
}

func TestEncodeSelfPermitRejectsNilToken(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	sig, err := key.Sign(types.NewHashFromBytes([]byte("some 32 byte digest padded out!")))
	require.NoError(t, err)

	_, err = EncodeSelfPermit(nil, big.NewInt(1000), big.NewInt(1700000000), sig)
	require.Error(t, err)
}

func TestEncodeSelfPermitIfNecessary(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	sig, err := key.Sign(types.NewHashFromBytes([]byte("some 32 byte digest padded out!")))
	require.NoError(t, err)

	data, err := EncodeSelfPermitIfNecessary(testRouterTokenIn, big.NewInt(1000), big.NewInt(1700000000), sig)
	require.NoError(t, err)
	require.Equal(t, []byte{0xc2, 0xe3, 0x14, 0x0a}, data[:4])

	vals, err := selfPermitIfNecessaryFn.DecodeCall(data)
	require.NoError(t, err)
	require.Equal(t, sig.LegacyBytes()[64], vals[3].(uint8))
}

func TestEncodeSelfPermitAllowed(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	sig, err := key.Sign(types.NewHashFromBytes([]byte("some 32 byte digest padded out!")))
	require.NoError(t, err)

	data, err := EncodeSelfPermitAllowed(testRouterTokenIn, big.NewInt(0), big.NewInt(1700000000), sig)
	require.NoError(t, err)
	require.Equal(t, []byte{0x46, 0x59, 0xa4, 0x94}, data[:4])

	vals, err := selfPermitAllowedFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testRouterTokenIn))
	require.Zero(t, big.NewInt(0).Cmp(vals[1].(*big.Int)))
	require.Zero(t, big.NewInt(1700000000).Cmp(vals[2].(*big.Int)))
}

func TestEncodeSelfPermitAllowedIfNecessary(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	sig, err := key.Sign(types.NewHashFromBytes([]byte("some 32 byte digest padded out!")))
	require.NoError(t, err)

	data, err := EncodeSelfPermitAllowedIfNecessary(testRouterTokenIn, big.NewInt(0), big.NewInt(1700000000), sig)
	require.NoError(t, err)
	require.Equal(t, []byte{0xa4, 0xa7, 0x8f, 0x0c}, data[:4])

	vals, err := selfPermitAllowedIfNecessaryFn.DecodeCall(data)
	require.NoError(t, err)
	require.Equal(t, sig.LegacyBytes()[64], vals[3].(uint8))
}

func TestEncodeExactInputSingleRejectsInt256ForUint24(t *testing.T) {
	_, err := EncodeExactInputSingle(testRouterTokenIn, testRouterTokenOut, big.NewInt(-1), testRouterRecipient, big.NewInt(1000000), big.NewInt(1), big.NewInt(0))
	require.Error(t, err)
}

func TestEncodeExactOutputSingleRejectsNilRecipient(t *testing.T) {
	_, err := EncodeExactOutputSingle(testRouterTokenIn, testRouterTokenOut, big.NewInt(500), nil, big.NewInt(1000000), big.NewInt(2000000), big.NewInt(0))
	require.Error(t, err)
}

func TestEncodeExactInputDistinguishesPathDirection(t *testing.T) {
	pathForward := append(append(testRouterTokenIn.Bytes(), 0x00, 0x0b, 0xb8), testRouterTokenOut.Bytes()...)
	pathBackward := append(append(testRouterTokenOut.Bytes(), 0x00, 0x0b, 0xb8), testRouterTokenIn.Bytes()...)

	dataForward, err := EncodeExactInput(pathForward, testRouterRecipient, big.NewInt(1000000), big.NewInt(1))
	require.NoError(t, err)
	dataBackward, err := EncodeExactInput(pathBackward, testRouterRecipient, big.NewInt(1000000), big.NewInt(1))
	require.NoError(t, err)
	require.NotEqual(t, dataForward, dataBackward)
}

func TestEncodeSweepTokenWithFeeRejectsUint256Overflow(t *testing.T) {
	overflow := new(big.Int).Lsh(big.NewInt(1), 256)
	_, err := EncodeSweepTokenWithFee(testRouterTokenIn, overflow, testRouterRecipient, big.NewInt(100), testRouterFeeAddr)
	require.Error(t, err)
}
