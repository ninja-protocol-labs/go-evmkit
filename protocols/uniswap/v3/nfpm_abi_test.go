package v3

import (
	"encoding/json/jsontext"
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

var (
	testNfpmToken0      = types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	testNfpmToken1      = types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	testNfpmRecipient   = types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	testNfpmOperator    = types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222")
	testNfpmFactoryAddr = types.MustNewAddressFromHex("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc")
)

func TestEncodeNfpmFactory(t *testing.T) {
	require.Equal(t, []byte{0xc4, 0x5a, 0x01, 0x55}, EncodeNfpmFactory())
}

func TestDecodeNfpmFactory(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testNfpmFactoryAddr)
	require.NoError(t, err)

	got, err := DecodeNfpmFactory(data)
	require.NoError(t, err)
	require.True(t, got.Equal(testNfpmFactoryAddr))
}

func TestDecodeNfpmFactoryRejectsGarbage(t *testing.T) {
	_, err := DecodeNfpmFactory([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeWETH9(t *testing.T) {
	require.Equal(t, []byte{0x4a, 0xa4, 0xa4, 0xfc}, EncodeWETH9())
}

func TestDecodeWETH9(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testNfpmToken0)
	require.NoError(t, err)

	got, err := DecodeWETH9(data)
	require.NoError(t, err)
	require.True(t, got.Equal(testNfpmToken0))
}

func TestDecodeWETH9RejectsGarbage(t *testing.T) {
	_, err := DecodeWETH9([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeCreateAndInitializePoolIfNecessary(t *testing.T) {
	data, err := EncodeCreateAndInitializePoolIfNecessary(testNfpmToken0, testNfpmToken1, big.NewInt(3000), big.NewInt(1))
	require.NoError(t, err)
	require.Equal(t, []byte{0x13, 0xea, 0xd5, 0x62}, data[:4])

	vals, err := createAndInitializePoolIfNecessaryFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testNfpmToken0))
	require.True(t, vals[1].(*types.Address).Equal(testNfpmToken1))
	require.Zero(t, big.NewInt(3000).Cmp(vals[2].(*big.Int)))
	require.Zero(t, big.NewInt(1).Cmp(vals[3].(*big.Int)))
}

func TestDecodeCreateAndInitializePoolIfNecessary(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testNfpmFactoryAddr)
	require.NoError(t, err)

	got, err := DecodeCreateAndInitializePoolIfNecessary(data)
	require.NoError(t, err)
	require.True(t, got.Equal(testNfpmFactoryAddr))
}

func TestDecodeCreateAndInitializePoolIfNecessaryRejectsGarbage(t *testing.T) {
	_, err := DecodeCreateAndInitializePoolIfNecessary([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeCreateAndInitializePoolIfNecessaryRejectsUint160Overflow(t *testing.T) {
	overflow := new(big.Int).Lsh(big.NewInt(1), 160)
	_, err := EncodeCreateAndInitializePoolIfNecessary(testNfpmToken0, testNfpmToken1, big.NewInt(3000), overflow)
	require.Error(t, err)
}

func TestEncodeNfpmPositions(t *testing.T) {
	data, err := EncodeNfpmPositions(big.NewInt(42))
	require.NoError(t, err)
	require.Equal(t, []byte{0x99, 0xfb, 0xab, 0x88}, data[:4])

	vals, err := nfpmPositionsFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(42).Cmp(vals[0].(*big.Int)))
}

func TestDecodeNfpmPositions(t *testing.T) {
	data, err := nfpmPositionsFn.EncodeReturn(big.NewInt(1), testNfpmOperator, testNfpmToken0, testNfpmToken1, big.NewInt(3000), big.NewInt(-60), big.NewInt(60), big.NewInt(1000), big.NewInt(2), big.NewInt(3), big.NewInt(4), big.NewInt(5))
	require.NoError(t, err)

	got, err := DecodeNfpmPositions(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1).Cmp(got.Nonce))
	require.True(t, got.Operator.Equal(testNfpmOperator))
	require.True(t, got.Token0.Equal(testNfpmToken0))
	require.True(t, got.Token1.Equal(testNfpmToken1))
	require.Zero(t, big.NewInt(3000).Cmp(got.Fee))
	require.Zero(t, big.NewInt(-60).Cmp(got.TickLower))
	require.Zero(t, big.NewInt(60).Cmp(got.TickUpper))
	require.Zero(t, big.NewInt(1000).Cmp(got.Liquidity))
	require.Zero(t, big.NewInt(2).Cmp(got.FeeGrowthInside0LastX128))
	require.Zero(t, big.NewInt(3).Cmp(got.FeeGrowthInside1LastX128))
	require.Zero(t, big.NewInt(4).Cmp(got.TokensOwed0))
	require.Zero(t, big.NewInt(5).Cmp(got.TokensOwed1))
}

func TestDecodeNfpmPositionsRejectsGarbage(t *testing.T) {
	_, err := DecodeNfpmPositions([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeNfpmMint(t *testing.T) {
	data, err := EncodeNfpmMint(testNfpmToken0, testNfpmToken1, big.NewInt(3000), big.NewInt(-60), big.NewInt(60), big.NewInt(1000), big.NewInt(1000), big.NewInt(0), big.NewInt(0), testNfpmRecipient, big.NewInt(1700000000))
	require.NoError(t, err)
	require.Equal(t, []byte{0x88, 0x31, 0x64, 0x56}, data[:4])

	vals, err := nfpmMintFn.DecodeCall(data)
	require.NoError(t, err)
	fields := vals[0].([]any)
	require.Len(t, fields, 11)
	require.True(t, fields[0].(*types.Address).Equal(testNfpmToken0))
	require.True(t, fields[1].(*types.Address).Equal(testNfpmToken1))
	require.Zero(t, big.NewInt(3000).Cmp(fields[2].(*big.Int)))
	require.Zero(t, big.NewInt(-60).Cmp(fields[3].(*big.Int)))
	require.Zero(t, big.NewInt(60).Cmp(fields[4].(*big.Int)))
	require.Zero(t, big.NewInt(1000).Cmp(fields[5].(*big.Int)))
	require.Zero(t, big.NewInt(1000).Cmp(fields[6].(*big.Int)))
	require.Zero(t, big.NewInt(0).Cmp(fields[7].(*big.Int)))
	require.Zero(t, big.NewInt(0).Cmp(fields[8].(*big.Int)))
	require.True(t, fields[9].(*types.Address).Equal(testNfpmRecipient))
	require.Zero(t, big.NewInt(1700000000).Cmp(fields[10].(*big.Int)))
}

func TestDecodeNfpmMint(t *testing.T) {
	data, err := nfpmMintFn.EncodeReturn(big.NewInt(1), big.NewInt(2), big.NewInt(3), big.NewInt(4))
	require.NoError(t, err)

	tokenId, liquidity, amount0, amount1, err := DecodeNfpmMint(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1).Cmp(tokenId))
	require.Zero(t, big.NewInt(2).Cmp(liquidity))
	require.Zero(t, big.NewInt(3).Cmp(amount0))
	require.Zero(t, big.NewInt(4).Cmp(amount1))
}

func TestDecodeNfpmMintRejectsGarbage(t *testing.T) {
	_, _, _, _, err := DecodeNfpmMint([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeNfpmMintRejectsNilToken0(t *testing.T) {
	_, err := EncodeNfpmMint(nil, testNfpmToken1, big.NewInt(3000), big.NewInt(-60), big.NewInt(60), big.NewInt(1000), big.NewInt(1000), big.NewInt(0), big.NewInt(0), testNfpmRecipient, big.NewInt(1700000000))
	require.Error(t, err)
}

func TestEncodeNfpmMintRejectsInt24Overflow(t *testing.T) {
	overflow := new(big.Int).Lsh(big.NewInt(1), 23)
	_, err := EncodeNfpmMint(testNfpmToken0, testNfpmToken1, big.NewInt(3000), overflow, big.NewInt(60), big.NewInt(1000), big.NewInt(1000), big.NewInt(0), big.NewInt(0), testNfpmRecipient, big.NewInt(1700000000))
	require.Error(t, err)
}

func TestEncodeIncreaseLiquidity(t *testing.T) {
	data, err := EncodeIncreaseLiquidity(big.NewInt(1), big.NewInt(1000), big.NewInt(1000), big.NewInt(0), big.NewInt(0), big.NewInt(1700000000))
	require.NoError(t, err)
	require.Equal(t, []byte{0x21, 0x9f, 0x5d, 0x17}, data[:4])

	vals, err := increaseLiquidityFn.DecodeCall(data)
	require.NoError(t, err)
	fields := vals[0].([]any)
	require.Len(t, fields, 6)
	require.Zero(t, big.NewInt(1).Cmp(fields[0].(*big.Int)))
	require.Zero(t, big.NewInt(1000).Cmp(fields[1].(*big.Int)))
	require.Zero(t, big.NewInt(1000).Cmp(fields[2].(*big.Int)))
	require.Zero(t, big.NewInt(0).Cmp(fields[3].(*big.Int)))
	require.Zero(t, big.NewInt(0).Cmp(fields[4].(*big.Int)))
	require.Zero(t, big.NewInt(1700000000).Cmp(fields[5].(*big.Int)))
}

func TestDecodeIncreaseLiquidity(t *testing.T) {
	data, err := increaseLiquidityFn.EncodeReturn(big.NewInt(1), big.NewInt(2), big.NewInt(3))
	require.NoError(t, err)

	liquidity, amount0, amount1, err := DecodeIncreaseLiquidity(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1).Cmp(liquidity))
	require.Zero(t, big.NewInt(2).Cmp(amount0))
	require.Zero(t, big.NewInt(3).Cmp(amount1))
}

func TestDecodeIncreaseLiquidityRejectsGarbage(t *testing.T) {
	_, _, _, err := DecodeIncreaseLiquidity([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecreaseLiquidity(t *testing.T) {
	data, err := EncodeDecreaseLiquidity(big.NewInt(1), big.NewInt(500), big.NewInt(0), big.NewInt(0), big.NewInt(1700000000))
	require.NoError(t, err)
	require.Equal(t, []byte{0x0c, 0x49, 0xcc, 0xbe}, data[:4])

	vals, err := decreaseLiquidityFn.DecodeCall(data)
	require.NoError(t, err)
	fields := vals[0].([]any)
	require.Len(t, fields, 5)
	require.Zero(t, big.NewInt(1).Cmp(fields[0].(*big.Int)))
	require.Zero(t, big.NewInt(500).Cmp(fields[1].(*big.Int)))
	require.Zero(t, big.NewInt(0).Cmp(fields[2].(*big.Int)))
	require.Zero(t, big.NewInt(0).Cmp(fields[3].(*big.Int)))
	require.Zero(t, big.NewInt(1700000000).Cmp(fields[4].(*big.Int)))
}

func TestDecodeDecreaseLiquidity(t *testing.T) {
	data, err := decreaseLiquidityFn.EncodeReturn(big.NewInt(1), big.NewInt(2))
	require.NoError(t, err)

	amount0, amount1, err := DecodeDecreaseLiquidity(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1).Cmp(amount0))
	require.Zero(t, big.NewInt(2).Cmp(amount1))
}

func TestDecodeDecreaseLiquidityRejectsGarbage(t *testing.T) {
	_, _, err := DecodeDecreaseLiquidity([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecreaseLiquidityRejectsUint128Overflow(t *testing.T) {
	overflow := new(big.Int).Lsh(big.NewInt(1), 128)
	_, err := EncodeDecreaseLiquidity(big.NewInt(1), overflow, big.NewInt(0), big.NewInt(0), big.NewInt(1700000000))
	require.Error(t, err)
}

func TestEncodeNfpmCollect(t *testing.T) {
	data, err := EncodeNfpmCollect(big.NewInt(1), testNfpmRecipient, big.NewInt(100), big.NewInt(200))
	require.NoError(t, err)
	require.Equal(t, []byte{0xfc, 0x6f, 0x78, 0x65}, data[:4])

	vals, err := nfpmCollectFn.DecodeCall(data)
	require.NoError(t, err)
	fields := vals[0].([]any)
	require.Len(t, fields, 4)
	require.Zero(t, big.NewInt(1).Cmp(fields[0].(*big.Int)))
	require.True(t, fields[1].(*types.Address).Equal(testNfpmRecipient))
	require.Zero(t, big.NewInt(100).Cmp(fields[2].(*big.Int)))
	require.Zero(t, big.NewInt(200).Cmp(fields[3].(*big.Int)))
}

func TestDecodeNfpmCollect(t *testing.T) {
	data, err := nfpmCollectFn.EncodeReturn(big.NewInt(1), big.NewInt(2))
	require.NoError(t, err)

	amount0, amount1, err := DecodeNfpmCollect(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1).Cmp(amount0))
	require.Zero(t, big.NewInt(2).Cmp(amount1))
}

func TestDecodeNfpmCollectRejectsGarbage(t *testing.T) {
	_, _, err := DecodeNfpmCollect([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeNfpmBurn(t *testing.T) {
	data, err := EncodeNfpmBurn(big.NewInt(1))
	require.NoError(t, err)
	require.Equal(t, []byte{0x42, 0x96, 0x6c, 0x68}, data[:4])

	vals, err := nfpmBurnFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1).Cmp(vals[0].(*big.Int)))
}

func TestEncodeUnwrapWETH9(t *testing.T) {
	data, err := EncodeUnwrapWETH9(big.NewInt(0), testNfpmRecipient)
	require.NoError(t, err)
	require.Equal(t, []byte{0x49, 0x40, 0x4b, 0x7c}, data[:4])

	vals, err := unwrapWETH9Fn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(0).Cmp(vals[0].(*big.Int)))
	require.True(t, vals[1].(*types.Address).Equal(testNfpmRecipient))
}

func TestEncodeSweepToken(t *testing.T) {
	data, err := EncodeSweepToken(testNfpmToken0, big.NewInt(0), testNfpmRecipient)
	require.NoError(t, err)
	require.Equal(t, []byte{0xdf, 0x2a, 0xb5, 0xbb}, data[:4])

	vals, err := sweepTokenFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testNfpmToken0))
	require.Zero(t, big.NewInt(0).Cmp(vals[1].(*big.Int)))
	require.True(t, vals[2].(*types.Address).Equal(testNfpmRecipient))
}

func TestEncodeRefundETH(t *testing.T) {
	require.Equal(t, []byte{0x12, 0x21, 0x0e, 0x8a}, EncodeRefundETH())
}

func TestEncodeMulticall(t *testing.T) {
	data, err := EncodeMulticall([][]byte{EncodeRefundETH(), {0x01, 0x02, 0x03}})
	require.NoError(t, err)
	require.Equal(t, []byte{0xac, 0x96, 0x50, 0xd8}, data[:4])

	vals, err := nfpmMulticallFn.DecodeCall(data)
	require.NoError(t, err)
	elems := vals[0].([]any)
	require.Len(t, elems, 2)
	require.Equal(t, EncodeRefundETH(), elems[0].([]byte))
	require.Equal(t, []byte{0x01, 0x02, 0x03}, elems[1].([]byte))
}

func TestEncodeMulticallEmpty(t *testing.T) {
	data, err := EncodeMulticall([][]byte{})
	require.NoError(t, err)

	vals, err := nfpmMulticallFn.DecodeCall(data)
	require.NoError(t, err)
	require.Empty(t, vals[0].([]any))
}

func TestDecodeMulticall(t *testing.T) {
	data, err := nfpmMulticallFn.EncodeReturn([][]byte{{0xaa, 0xbb}, {0xcc}})
	require.NoError(t, err)

	got, err := DecodeMulticall(data)
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, []byte{0xaa, 0xbb}, got[0])
	require.Equal(t, []byte{0xcc}, got[1])
}

func TestDecodeMulticallRejectsGarbage(t *testing.T) {
	_, err := DecodeMulticall([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeTotalSupply(t *testing.T) {
	require.Equal(t, []byte{0x18, 0x16, 0x0d, 0xdd}, EncodeTotalSupply())
}

func TestDecodeTotalSupply(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(42))
	require.NoError(t, err)

	got, err := DecodeTotalSupply(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(42).Cmp(got))
}

func TestDecodeTotalSupplyRejectsGarbage(t *testing.T) {
	_, err := DecodeTotalSupply([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeTokenOfOwnerByIndex(t *testing.T) {
	data, err := EncodeTokenOfOwnerByIndex(testNfpmRecipient, big.NewInt(0))
	require.NoError(t, err)
	require.Equal(t, []byte{0x2f, 0x74, 0x5c, 0x59}, data[:4])

	vals, err := tokenOfOwnerByIndexFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testNfpmRecipient))
	require.Zero(t, big.NewInt(0).Cmp(vals[1].(*big.Int)))
}

func TestDecodeTokenOfOwnerByIndex(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(7))
	require.NoError(t, err)

	got, err := DecodeTokenOfOwnerByIndex(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(7).Cmp(got))
}

func TestDecodeTokenOfOwnerByIndexRejectsGarbage(t *testing.T) {
	_, err := DecodeTokenOfOwnerByIndex([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeTokenByIndex(t *testing.T) {
	data, err := EncodeTokenByIndex(big.NewInt(3))
	require.NoError(t, err)
	require.Equal(t, []byte{0x4f, 0x6c, 0xcc, 0xe7}, data[:4])

	vals, err := tokenByIndexFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(3).Cmp(vals[0].(*big.Int)))
}

func TestDecodeTokenByIndex(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(9))
	require.NoError(t, err)

	got, err := DecodeTokenByIndex(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(9).Cmp(got))
}

func TestDecodeTokenByIndexRejectsGarbage(t *testing.T) {
	_, err := DecodeTokenByIndex([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodePermitTypehash(t *testing.T) {
	require.Equal(t, []byte{0x30, 0xad, 0xf8, 0x1f}, EncodePermitTypehash())
}

func TestDecodePermitTypehash(t *testing.T) {
	var want [32]byte
	want[0] = 0xde
	want[31] = 0xef
	data, err := abi.Pack(abi.Types{abi.Bytes32}, want[:])
	require.NoError(t, err)

	got, err := DecodePermitTypehash(data)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDecodePermitTypehashRejectsGarbage(t *testing.T) {
	_, err := DecodePermitTypehash([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDomainSeparator(t *testing.T) {
	require.Equal(t, []byte{0x36, 0x44, 0xe5, 0x15}, EncodeDomainSeparator())
}

func TestDecodeDomainSeparator(t *testing.T) {
	var want [32]byte
	want[0] = 0xaa
	want[31] = 0xbb
	data, err := abi.Pack(abi.Types{abi.Bytes32}, want[:])
	require.NoError(t, err)

	got, err := DecodeDomainSeparator(data)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDecodeDomainSeparatorRejectsGarbage(t *testing.T) {
	_, err := DecodeDomainSeparator([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeNFTPermit(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	sig, err := key.Sign(types.NewHashFromBytes([]byte("some 32 byte digest padded out!")))
	require.NoError(t, err)

	data, err := EncodeNFTPermit(testNfpmRecipient, big.NewInt(1), big.NewInt(1700000000), sig)
	require.NoError(t, err)
	require.Equal(t, []byte{0x7a, 0xc2, 0xff, 0x7b}, data[:4])

	vals, err := nftPermitFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testNfpmRecipient))
	require.Zero(t, big.NewInt(1).Cmp(vals[1].(*big.Int)))
	require.Zero(t, big.NewInt(1700000000).Cmp(vals[2].(*big.Int)))
	legacy := sig.LegacyBytes()
	require.Equal(t, legacy[64], vals[3].(uint8))
	require.Equal(t, legacy[:32], vals[4].([]byte))
	require.Equal(t, legacy[32:64], vals[5].([]byte))
}

func TestEncodeNFTPermitRejectsNilSpender(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	sig, err := key.Sign(types.NewHashFromBytes([]byte("some 32 byte digest padded out!")))
	require.NoError(t, err)

	_, err = EncodeNFTPermit(nil, big.NewInt(1), big.NewInt(1700000000), sig)
	require.Error(t, err)
}

func TestNfpmRevertReasons(t *testing.T) {
	reasons := []string{
		reasonTransactionTooOld,
		reasonInsufficientWETH9,
		reasonInsufficientToken,
		reasonNotWETH9,
		reasonTransferFromFailed,
		reasonTokenTransferFailed,
		reasonApproveFailed,
		reasonTransferETHFailed,
		reasonInvalidTokenID,
		reasonNotApproved,
		reasonPriceSlippageCheck,
		reasonNotCleared,
		reasonPermitExpired,
		reasonApprovalToCurrentOwner,
		reasonUnauthorized,
		reasonInvalidSignature,
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

func TestNfpmWrapRevertErrorZeroLengthDataFallsBackToErrReverted(t *testing.T) {
	n := &NFPM{}
	rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted"}
	err := n.wrapRevertError(rpcErr)
	require.ErrorIs(t, err, ErrReverted)
}

func TestNfpmWrapRevertErrorEmptyHexDataFallsBackToErrReverted(t *testing.T) {
	n := &NFPM{}
	rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"0x"`)}
	err := n.wrapRevertError(rpcErr)
	require.ErrorIs(t, err, ErrReverted)
}

func TestNfpmWrapRevertErrorKnownReasonMapsToSentinel(t *testing.T) {
	n := &NFPM{}
	revertData, err := abi.NewFunction("Error", abi.Types{abi.String}, nil).EncodeCall(reasonNotApproved)
	require.NoError(t, err)
	hexData := `"` + encoding.Hex.EncodePrefixed(revertData) + `"`
	rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(hexData)}
	wrapped := n.wrapRevertError(rpcErr)
	require.ErrorIs(t, wrapped, ErrNotApproved)
}

func TestNfpmWrapRevertErrorNonRPCErrorPassesThrough(t *testing.T) {
	n := &NFPM{}
	plain := errors.New("boom")
	err := n.wrapRevertError(plain)
	require.Same(t, plain, err)
}
