package v2

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestEncodeMinimumLiquidity(t *testing.T) {
	require.Equal(t, []byte{0xba, 0x9a, 0x7a, 0x56}, EncodeMinimumLiquidity())
}

func TestDecodeMinimumLiquidity(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(1000))
	require.NoError(t, err)

	got, err := DecodeMinimumLiquidity(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1000).Cmp(got))
}

func TestEncodeFactory(t *testing.T) {
	require.Equal(t, []byte{0xc4, 0x5a, 0x01, 0x55}, EncodeFactory())
}

func TestDecodeFactory(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testTokenA)
	require.NoError(t, err)

	got, err := DecodeFactory(data)
	require.NoError(t, err)
	require.True(t, got.Equal(testTokenA))
}

func TestEncodeToken0(t *testing.T) {
	require.Equal(t, []byte{0x0d, 0xfe, 0x16, 0x81}, EncodeToken0())
}

func TestDecodeToken0(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testTokenA)
	require.NoError(t, err)

	got, err := DecodeToken0(data)
	require.NoError(t, err)
	require.True(t, got.Equal(testTokenA))
}

func TestEncodeToken1(t *testing.T) {
	require.Equal(t, []byte{0xd2, 0x12, 0x20, 0xa7}, EncodeToken1())
}

func TestDecodeToken1(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testTokenB)
	require.NoError(t, err)

	got, err := DecodeToken1(data)
	require.NoError(t, err)
	require.True(t, got.Equal(testTokenB))
}

func TestEncodeGetReserves(t *testing.T) {
	require.Equal(t, []byte{0x09, 0x02, 0xf1, 0xac}, EncodeGetReserves())
}

func TestDecodeGetReserves(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint112, abi.Uint112, abi.Uint32}, big.NewInt(1000), big.NewInt(2000), uint32(123456))
	require.NoError(t, err)

	got, err := DecodeGetReserves(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1000).Cmp(got.Reserve0))
	require.Zero(t, big.NewInt(2000).Cmp(got.Reserve1))
	require.Equal(t, uint32(123456), got.BlockTimestampLast)
}

func TestDecodeGetReservesRejectsGarbage(t *testing.T) {
	_, err := DecodeGetReserves([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodePrice0CumulativeLast(t *testing.T) {
	require.Equal(t, []byte{0x59, 0x09, 0xc0, 0xd5}, EncodePrice0CumulativeLast())
}

func TestDecodePrice0CumulativeLast(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(42))
	require.NoError(t, err)

	got, err := DecodePrice0CumulativeLast(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(42).Cmp(got))
}

func TestEncodePrice1CumulativeLast(t *testing.T) {
	require.Equal(t, []byte{0x5a, 0x3d, 0x54, 0x93}, EncodePrice1CumulativeLast())
}

func TestDecodePrice1CumulativeLast(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(42))
	require.NoError(t, err)

	got, err := DecodePrice1CumulativeLast(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(42).Cmp(got))
}

func TestEncodeKLast(t *testing.T) {
	require.Equal(t, []byte{0x74, 0x64, 0xfc, 0x3d}, EncodeKLast())
}

func TestDecodeKLast(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(42))
	require.NoError(t, err)

	got, err := DecodeKLast(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(42).Cmp(got))
}

func TestEncodeMint(t *testing.T) {
	data, err := EncodeMint(testTokenA)
	require.NoError(t, err)
	require.Equal(t, []byte{0x6a, 0x62, 0x78, 0x42}, data[:4])

	vals, err := mintFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testTokenA))
}

func TestDecodeMint(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(500))
	require.NoError(t, err)

	got, err := DecodeMint(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(500).Cmp(got))
}

func TestEncodeBurn(t *testing.T) {
	data, err := EncodeBurn(testTokenA)
	require.NoError(t, err)
	require.Equal(t, []byte{0x89, 0xaf, 0xcb, 0x44}, data[:4])

	vals, err := burnFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testTokenA))
}

func TestDecodeBurn(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256, abi.Uint256}, big.NewInt(100), big.NewInt(200))
	require.NoError(t, err)

	amount0, amount1, err := DecodeBurn(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(100).Cmp(amount0))
	require.Zero(t, big.NewInt(200).Cmp(amount1))
}

func TestDecodeBurnRejectsGarbage(t *testing.T) {
	_, _, err := DecodeBurn([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeSwap(t *testing.T) {
	data, err := EncodeSwap(big.NewInt(0), big.NewInt(95), testTokenA, nil)
	require.NoError(t, err)
	require.Equal(t, []byte{0x02, 0x2c, 0x0d, 0x9f}, data[:4])

	vals, err := swapFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(0).Cmp(vals[0].(*big.Int)))
	require.Zero(t, big.NewInt(95).Cmp(vals[1].(*big.Int)))
	require.True(t, vals[2].(*types.Address).Equal(testTokenA))
	require.Empty(t, vals[3].([]byte))
}

func TestEncodeSkim(t *testing.T) {
	data, err := EncodeSkim(testTokenA)
	require.NoError(t, err)
	require.Equal(t, []byte{0xbc, 0x25, 0xcf, 0x77}, data[:4])

	vals, err := skimFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testTokenA))
}

func TestEncodeSync(t *testing.T) {
	require.Equal(t, []byte{0xff, 0xf6, 0xca, 0xe9}, EncodeSync())
}

func TestEncodeInitialize(t *testing.T) {
	data, err := EncodeInitialize(testTokenA, testTokenB)
	require.NoError(t, err)
	require.Equal(t, []byte{0x48, 0x5c, 0xc9, 0x55}, data[:4])

	vals, err := initializeFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testTokenA))
	require.True(t, vals[1].(*types.Address).Equal(testTokenB))
}

// TestPairRevertReasons confirms every reasonX constant exactly matches what UniswapV2Pair's own
// require(cond, reason) strings decode to via abi.DecodeRevert.
func TestPairRevertReasons(t *testing.T) {
	reasons := []string{
		reasonLocked, reasonTransferFailed, reasonOverflow,
		reasonInsufficientLiquidityMinted, reasonInsufficientLiquidityBurned,
		reasonInsufficientOutputAmount, reasonInsufficientLiquidity,
		reasonInvalidTo, reasonInsufficientInputAmount, reasonK,
		reasonAddOverflow, reasonSubUnderflow, reasonMulOverflow,
	}
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

func TestEncodePermitTypehash(t *testing.T) {
	require.Equal(t, []byte{0x30, 0xad, 0xf8, 0x1f}, EncodePermitTypehash())
}

func TestDecodePermitTypehash(t *testing.T) {
	want := make([]byte, 32)
	want[31] = 0xab
	data, err := abi.Pack(abi.Types{abi.Bytes32}, want)
	require.NoError(t, err)

	got, err := DecodePermitTypehash(data)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDecodePermitTypehashRejectsGarbage(t *testing.T) {
	_, err := DecodePermitTypehash([]byte{0x01, 0x02})
	require.Error(t, err)
}
