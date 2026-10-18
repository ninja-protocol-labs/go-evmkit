package v2

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
	testPair   = types.MustNewAddressFromHex("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc")
)

func TestEncodeFeeTo(t *testing.T) {
	require.Equal(t, []byte{0x01, 0x7e, 0x7e, 0x58}, EncodeFeeTo())
}

func TestDecodeFeeTo(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testTokenA)
	require.NoError(t, err)

	got, err := DecodeFeeTo(data)
	require.NoError(t, err)
	require.True(t, got.Equal(testTokenA))
}

func TestDecodeFeeToRejectsGarbage(t *testing.T) {
	_, err := DecodeFeeTo([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeFeeToSetter(t *testing.T) {
	require.Equal(t, []byte{0x09, 0x4b, 0x74, 0x15}, EncodeFeeToSetter())
}

func TestDecodeFeeToSetter(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testTokenB)
	require.NoError(t, err)

	got, err := DecodeFeeToSetter(data)
	require.NoError(t, err)
	require.True(t, got.Equal(testTokenB))
}

func TestDecodeFeeToSetterRejectsGarbage(t *testing.T) {
	_, err := DecodeFeeToSetter([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeGetPair(t *testing.T) {
	data, err := EncodeGetPair(testTokenA, testTokenB)
	require.NoError(t, err)
	require.Equal(t, []byte{0xe6, 0xa4, 0x39, 0x05}, data[:4])

	vals, err := getPairFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testTokenA))
	require.True(t, vals[1].(*types.Address).Equal(testTokenB))
}

func TestDecodeGetPair(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testPair)
	require.NoError(t, err)

	got, err := DecodeGetPair(data)
	require.NoError(t, err)
	require.True(t, got.Equal(testPair))
}

func TestDecodeGetPairRejectsGarbage(t *testing.T) {
	_, err := DecodeGetPair([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeAllPairs(t *testing.T) {
	data, err := EncodeAllPairs(big.NewInt(7))
	require.NoError(t, err)
	require.Equal(t, []byte{0x1e, 0x3d, 0xd1, 0x8b}, data[:4])

	vals, err := allPairsFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(7).Cmp(vals[0].(*big.Int)))
}

func TestDecodeAllPairs(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testPair)
	require.NoError(t, err)

	got, err := DecodeAllPairs(data)
	require.NoError(t, err)
	require.True(t, got.Equal(testPair))
}

func TestDecodeAllPairsRejectsGarbage(t *testing.T) {
	_, err := DecodeAllPairs([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeAllPairsLength(t *testing.T) {
	require.Equal(t, []byte{0x57, 0x4f, 0x2b, 0xa3}, EncodeAllPairsLength())
}

func TestDecodeAllPairsLength(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(42))
	require.NoError(t, err)

	got, err := DecodeAllPairsLength(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(42).Cmp(got))
}

func TestDecodeAllPairsLengthRejectsGarbage(t *testing.T) {
	_, err := DecodeAllPairsLength([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeCreatePair(t *testing.T) {
	data, err := EncodeCreatePair(testTokenA, testTokenB)
	require.NoError(t, err)
	require.Equal(t, []byte{0xc9, 0xc6, 0x53, 0x96}, data[:4])

	vals, err := createPairFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testTokenA))
	require.True(t, vals[1].(*types.Address).Equal(testTokenB))
}

func TestDecodeCreatePair(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testPair)
	require.NoError(t, err)

	got, err := DecodeCreatePair(data)
	require.NoError(t, err)
	require.True(t, got.Equal(testPair))
}

func TestDecodeCreatePairRejectsGarbage(t *testing.T) {
	_, err := DecodeCreatePair([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeSetFeeTo(t *testing.T) {
	data, err := EncodeSetFeeTo(testTokenA)
	require.NoError(t, err)
	require.Equal(t, []byte{0xf4, 0x69, 0x01, 0xed}, data[:4])

	vals, err := setFeeToFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testTokenA))
}

func TestEncodeSetFeeToSetter(t *testing.T) {
	data, err := EncodeSetFeeToSetter(testTokenB)
	require.NoError(t, err)
	require.Equal(t, []byte{0xa2, 0xe7, 0x4a, 0xf6}, data[:4])

	vals, err := setFeeToSetterFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testTokenB))
}

func TestRevertReasons(t *testing.T) {
	reasons := []string{reasonIdenticalAddresses, reasonZeroAddress, reasonPairExists, reasonForbidden}
	for _, reason := range reasons {
		t.Run(reason, func(t *testing.T) {
			data, err := abi.NewFunction("Error", abi.Types{abi.String}, nil).EncodeCall(reason)
			require.NoError(t, err)

			got, err := abi.DecodeRevert(data)
			require.NoError(t, err)
			require.Equal(t, reason, got)
		})
	}
}
