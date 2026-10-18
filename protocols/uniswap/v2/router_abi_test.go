package v2

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

var (
	testTokenC = types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	testTo     = types.MustNewAddressFromHex("0x5c69bee701ef814a2b6a3edd4b1652cb9cc5aa6f")
)

func testRouterSignature(t *testing.T) *types.Signature {
	t.Helper()
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	sig, err := key.Sign(types.NewHashFromBytes([]byte("some 32 byte digest padded out!")))
	require.NoError(t, err)
	return sig
}

func requirePermitParts(t *testing.T, sig *types.Signature, vals []any, offset int) {
	t.Helper()
	wantV := uint8(sig.LegacyV().Uint64())
	require.Equal(t, wantV, vals[offset].(uint8))

	wantR := make([]byte, 32)
	sig.R().FillBytes(wantR)
	require.Equal(t, wantR, vals[offset+1].([]byte))

	wantS := make([]byte, 32)
	sig.S().FillBytes(wantS)
	require.Equal(t, wantS, vals[offset+2].([]byte))
}

func requirePath(t *testing.T, path []*types.Address, got []any) {
	t.Helper()
	require.Len(t, got, len(path))
	for i, addr := range path {
		require.True(t, got[i].(*types.Address).Equal(addr))
	}
}

func TestEncodeRouterFactory(t *testing.T) {
	require.Equal(t, []byte{0xc4, 0x5a, 0x01, 0x55}, EncodeRouterFactory())
}

func TestDecodeRouterFactory(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testTokenA)
	require.NoError(t, err)

	got, err := DecodeRouterFactory(data)
	require.NoError(t, err)
	require.True(t, got.Equal(testTokenA))
}

func TestEncodeWETH(t *testing.T) {
	require.Equal(t, []byte{0xad, 0x5c, 0x46, 0x48}, EncodeWETH())
}

func TestDecodeWETH(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testTokenA)
	require.NoError(t, err)

	got, err := DecodeWETH(data)
	require.NoError(t, err)
	require.True(t, got.Equal(testTokenA))
}

func TestEncodeQuote(t *testing.T) {
	data, err := EncodeQuote(big.NewInt(100), big.NewInt(1000), big.NewInt(2000))
	require.NoError(t, err)
	require.Equal(t, []byte{0xad, 0x61, 0x5d, 0xec}, data[:4])

	vals, err := quoteFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(100).Cmp(vals[0].(*big.Int)))
	require.Zero(t, big.NewInt(1000).Cmp(vals[1].(*big.Int)))
	require.Zero(t, big.NewInt(2000).Cmp(vals[2].(*big.Int)))
}

func TestDecodeQuote(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(200))
	require.NoError(t, err)

	got, err := DecodeQuote(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(200).Cmp(got))
}

func TestEncodeGetAmountOut(t *testing.T) {
	data, err := EncodeGetAmountOut(big.NewInt(100), big.NewInt(1000), big.NewInt(2000))
	require.NoError(t, err)
	require.Equal(t, []byte{0x05, 0x4d, 0x50, 0xd4}, data[:4])

	vals, err := getAmountOutFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(100).Cmp(vals[0].(*big.Int)))
	require.Zero(t, big.NewInt(1000).Cmp(vals[1].(*big.Int)))
	require.Zero(t, big.NewInt(2000).Cmp(vals[2].(*big.Int)))
}

func TestDecodeGetAmountOut(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(95))
	require.NoError(t, err)

	got, err := DecodeGetAmountOut(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(95).Cmp(got))
}

func TestEncodeGetAmountIn(t *testing.T) {
	data, err := EncodeGetAmountIn(big.NewInt(95), big.NewInt(1000), big.NewInt(2000))
	require.NoError(t, err)
	require.Equal(t, []byte{0x85, 0xf8, 0xc2, 0x59}, data[:4])

	vals, err := getAmountInFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(95).Cmp(vals[0].(*big.Int)))
	require.Zero(t, big.NewInt(1000).Cmp(vals[1].(*big.Int)))
	require.Zero(t, big.NewInt(2000).Cmp(vals[2].(*big.Int)))
}

func TestDecodeGetAmountIn(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(100))
	require.NoError(t, err)

	got, err := DecodeGetAmountIn(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(100).Cmp(got))
}

func TestEncodeGetAmountsOut(t *testing.T) {
	path := []*types.Address{testTokenA, testTokenB, testTokenC}
	data, err := EncodeGetAmountsOut(big.NewInt(100), path)
	require.NoError(t, err)
	require.Equal(t, []byte{0xd0, 0x6c, 0xa6, 0x1f}, data[:4])

	vals, err := getAmountsOutFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(100).Cmp(vals[0].(*big.Int)))
	requirePath(t, path, vals[1].([]any))
}

func TestDecodeGetAmountsOut(t *testing.T) {
	amounts := []*big.Int{big.NewInt(100), big.NewInt(95), big.NewInt(90)}
	data, err := abi.Pack(abi.Types{abi.Slice(abi.Uint256)}, amounts)
	require.NoError(t, err)

	got, err := DecodeGetAmountsOut(data)
	require.NoError(t, err)
	require.Len(t, got, 3)
	for i, want := range amounts {
		require.Zero(t, want.Cmp(got[i]))
	}
}

func TestDecodeGetAmountsOutRejectsGarbage(t *testing.T) {
	_, err := DecodeGetAmountsOut([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeGetAmountsIn(t *testing.T) {
	path := []*types.Address{testTokenA, testTokenB}
	data, err := EncodeGetAmountsIn(big.NewInt(100), path)
	require.NoError(t, err)
	require.Equal(t, []byte{0x1f, 0x00, 0xca, 0x74}, data[:4])

	vals, err := getAmountsInFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(100).Cmp(vals[0].(*big.Int)))
	requirePath(t, path, vals[1].([]any))
}

func TestDecodeGetAmountsIn(t *testing.T) {
	amounts := []*big.Int{big.NewInt(105), big.NewInt(100)}
	data, err := abi.Pack(abi.Types{abi.Slice(abi.Uint256)}, amounts)
	require.NoError(t, err)

	got, err := DecodeGetAmountsIn(data)
	require.NoError(t, err)
	require.Len(t, got, 2)
	for i, want := range amounts {
		require.Zero(t, want.Cmp(got[i]))
	}
}

func TestDecodeGetAmountsInRejectsGarbage(t *testing.T) {
	_, err := DecodeGetAmountsIn([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeAddLiquidity(t *testing.T) {
	deadline := big.NewInt(9999999999)
	data, err := EncodeAddLiquidity(testTokenA, testTokenB, big.NewInt(1000), big.NewInt(2000), big.NewInt(900), big.NewInt(1900), testTo, deadline)
	require.NoError(t, err)
	require.Equal(t, []byte{0xe8, 0xe3, 0x37, 0x00}, data[:4])

	vals, err := addLiquidityFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testTokenA))
	require.True(t, vals[1].(*types.Address).Equal(testTokenB))
	require.Zero(t, big.NewInt(1000).Cmp(vals[2].(*big.Int)))
	require.Zero(t, big.NewInt(2000).Cmp(vals[3].(*big.Int)))
	require.Zero(t, big.NewInt(900).Cmp(vals[4].(*big.Int)))
	require.Zero(t, big.NewInt(1900).Cmp(vals[5].(*big.Int)))
	require.True(t, vals[6].(*types.Address).Equal(testTo))
	require.Zero(t, deadline.Cmp(vals[7].(*big.Int)))
}

func TestDecodeAddLiquidity(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256, abi.Uint256, abi.Uint256}, big.NewInt(1000), big.NewInt(2000), big.NewInt(1500))
	require.NoError(t, err)

	got, err := DecodeAddLiquidity(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1000).Cmp(got.AmountA))
	require.Zero(t, big.NewInt(2000).Cmp(got.AmountB))
	require.Zero(t, big.NewInt(1500).Cmp(got.Liquidity))
}

func TestDecodeAddLiquidityRejectsGarbage(t *testing.T) {
	_, err := DecodeAddLiquidity([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeAddLiquidityETH(t *testing.T) {
	deadline := big.NewInt(9999999999)
	data, err := EncodeAddLiquidityETH(testTokenA, big.NewInt(1000), big.NewInt(900), big.NewInt(1), testTo, deadline)
	require.NoError(t, err)
	require.Equal(t, []byte{0xf3, 0x05, 0xd7, 0x19}, data[:4])

	vals, err := addLiquidityETHFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testTokenA))
	require.Zero(t, big.NewInt(1000).Cmp(vals[1].(*big.Int)))
	require.Zero(t, big.NewInt(900).Cmp(vals[2].(*big.Int)))
	require.Zero(t, big.NewInt(1).Cmp(vals[3].(*big.Int)))
	require.True(t, vals[4].(*types.Address).Equal(testTo))
	require.Zero(t, deadline.Cmp(vals[5].(*big.Int)))
}

func TestDecodeAddLiquidityETH(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256, abi.Uint256, abi.Uint256}, big.NewInt(1000), big.NewInt(5), big.NewInt(1500))
	require.NoError(t, err)

	got, err := DecodeAddLiquidityETH(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1000).Cmp(got.AmountToken))
	require.Zero(t, big.NewInt(5).Cmp(got.AmountETH))
	require.Zero(t, big.NewInt(1500).Cmp(got.Liquidity))
}

func TestDecodeAddLiquidityETHRejectsGarbage(t *testing.T) {
	_, err := DecodeAddLiquidityETH([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeRemoveLiquidity(t *testing.T) {
	deadline := big.NewInt(9999999999)
	data, err := EncodeRemoveLiquidity(testTokenA, testTokenB, big.NewInt(500), big.NewInt(100), big.NewInt(200), testTo, deadline)
	require.NoError(t, err)
	require.Equal(t, []byte{0xba, 0xa2, 0xab, 0xde}, data[:4])

	vals, err := removeLiquidityFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testTokenA))
	require.True(t, vals[1].(*types.Address).Equal(testTokenB))
	require.Zero(t, big.NewInt(500).Cmp(vals[2].(*big.Int)))
	require.Zero(t, big.NewInt(100).Cmp(vals[3].(*big.Int)))
	require.Zero(t, big.NewInt(200).Cmp(vals[4].(*big.Int)))
	require.True(t, vals[5].(*types.Address).Equal(testTo))
	require.Zero(t, deadline.Cmp(vals[6].(*big.Int)))
}

func TestDecodeRemoveLiquidity(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256, abi.Uint256}, big.NewInt(100), big.NewInt(200))
	require.NoError(t, err)

	got, err := DecodeRemoveLiquidity(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(100).Cmp(got.AmountA))
	require.Zero(t, big.NewInt(200).Cmp(got.AmountB))
}

func TestDecodeRemoveLiquidityRejectsGarbage(t *testing.T) {
	_, err := DecodeRemoveLiquidity([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeRemoveLiquidityETH(t *testing.T) {
	deadline := big.NewInt(9999999999)
	data, err := EncodeRemoveLiquidityETH(testTokenA, big.NewInt(500), big.NewInt(100), big.NewInt(1), testTo, deadline)
	require.NoError(t, err)
	require.Equal(t, []byte{0x02, 0x75, 0x1c, 0xec}, data[:4])

	vals, err := removeLiquidityETHFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testTokenA))
	require.Zero(t, big.NewInt(500).Cmp(vals[1].(*big.Int)))
	require.Zero(t, big.NewInt(100).Cmp(vals[2].(*big.Int)))
	require.Zero(t, big.NewInt(1).Cmp(vals[3].(*big.Int)))
	require.True(t, vals[4].(*types.Address).Equal(testTo))
	require.Zero(t, deadline.Cmp(vals[5].(*big.Int)))
}

func TestDecodeRemoveLiquidityETH(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256, abi.Uint256}, big.NewInt(100), big.NewInt(5))
	require.NoError(t, err)

	got, err := DecodeRemoveLiquidityETH(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(100).Cmp(got.AmountToken))
	require.Zero(t, big.NewInt(5).Cmp(got.AmountETH))
}

func TestDecodeRemoveLiquidityETHRejectsGarbage(t *testing.T) {
	_, err := DecodeRemoveLiquidityETH([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeRemoveLiquidityWithPermit(t *testing.T) {
	sig := testRouterSignature(t)
	deadline := big.NewInt(9999999999)
	data, err := EncodeRemoveLiquidityWithPermit(testTokenA, testTokenB, big.NewInt(500), big.NewInt(100), big.NewInt(200), testTo, deadline, true, sig)
	require.NoError(t, err)
	require.Equal(t, []byte{0x21, 0x95, 0x99, 0x5c}, data[:4])

	vals, err := removeLiquidityWithPermitFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testTokenA))
	require.True(t, vals[1].(*types.Address).Equal(testTokenB))
	require.Zero(t, big.NewInt(500).Cmp(vals[2].(*big.Int)))
	require.Zero(t, big.NewInt(100).Cmp(vals[3].(*big.Int)))
	require.Zero(t, big.NewInt(200).Cmp(vals[4].(*big.Int)))
	require.True(t, vals[5].(*types.Address).Equal(testTo))
	require.Zero(t, deadline.Cmp(vals[6].(*big.Int)))
	require.True(t, vals[7].(bool))
	requirePermitParts(t, sig, vals, 8)
}

func TestDecodeRemoveLiquidityWithPermit(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256, abi.Uint256}, big.NewInt(100), big.NewInt(200))
	require.NoError(t, err)

	got, err := DecodeRemoveLiquidityWithPermit(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(100).Cmp(got.AmountA))
	require.Zero(t, big.NewInt(200).Cmp(got.AmountB))
}

func TestEncodeRemoveLiquidityETHWithPermit(t *testing.T) {
	sig := testRouterSignature(t)
	deadline := big.NewInt(9999999999)
	data, err := EncodeRemoveLiquidityETHWithPermit(testTokenA, big.NewInt(500), big.NewInt(100), big.NewInt(1), testTo, deadline, false, sig)
	require.NoError(t, err)
	require.Equal(t, []byte{0xde, 0xd9, 0x38, 0x2a}, data[:4])

	vals, err := removeLiquidityETHWithPermitFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testTokenA))
	require.Zero(t, big.NewInt(500).Cmp(vals[1].(*big.Int)))
	require.Zero(t, big.NewInt(100).Cmp(vals[2].(*big.Int)))
	require.Zero(t, big.NewInt(1).Cmp(vals[3].(*big.Int)))
	require.True(t, vals[4].(*types.Address).Equal(testTo))
	require.Zero(t, deadline.Cmp(vals[5].(*big.Int)))
	require.False(t, vals[6].(bool))
	requirePermitParts(t, sig, vals, 7)
}

func TestDecodeRemoveLiquidityETHWithPermit(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256, abi.Uint256}, big.NewInt(100), big.NewInt(5))
	require.NoError(t, err)

	got, err := DecodeRemoveLiquidityETHWithPermit(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(100).Cmp(got.AmountToken))
	require.Zero(t, big.NewInt(5).Cmp(got.AmountETH))
}

func TestEncodeSwapExactTokensForTokens(t *testing.T) {
	path := []*types.Address{testTokenA, testTokenB}
	deadline := big.NewInt(9999999999)
	data, err := EncodeSwapExactTokensForTokens(big.NewInt(100), big.NewInt(90), path, testTo, deadline)
	require.NoError(t, err)
	require.Equal(t, []byte{0x38, 0xed, 0x17, 0x39}, data[:4])

	vals, err := swapExactTokensForTokensFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(100).Cmp(vals[0].(*big.Int)))
	require.Zero(t, big.NewInt(90).Cmp(vals[1].(*big.Int)))
	requirePath(t, path, vals[2].([]any))
	require.True(t, vals[3].(*types.Address).Equal(testTo))
	require.Zero(t, deadline.Cmp(vals[4].(*big.Int)))
}

func TestDecodeSwapExactTokensForTokens(t *testing.T) {
	amounts := []*big.Int{big.NewInt(100), big.NewInt(95)}
	data, err := abi.Pack(abi.Types{abi.Slice(abi.Uint256)}, amounts)
	require.NoError(t, err)

	got, err := DecodeSwapExactTokensForTokens(data)
	require.NoError(t, err)
	require.Len(t, got, 2)
	for i, want := range amounts {
		require.Zero(t, want.Cmp(got[i]))
	}
}

func TestEncodeSwapTokensForExactTokens(t *testing.T) {
	path := []*types.Address{testTokenA, testTokenB}
	deadline := big.NewInt(9999999999)
	data, err := EncodeSwapTokensForExactTokens(big.NewInt(90), big.NewInt(100), path, testTo, deadline)
	require.NoError(t, err)
	require.Equal(t, []byte{0x88, 0x03, 0xdb, 0xee}, data[:4])

	vals, err := swapTokensForExactTokensFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(90).Cmp(vals[0].(*big.Int)))
	require.Zero(t, big.NewInt(100).Cmp(vals[1].(*big.Int)))
	requirePath(t, path, vals[2].([]any))
	require.True(t, vals[3].(*types.Address).Equal(testTo))
	require.Zero(t, deadline.Cmp(vals[4].(*big.Int)))
}

func TestDecodeSwapTokensForExactTokens(t *testing.T) {
	amounts := []*big.Int{big.NewInt(105), big.NewInt(100)}
	data, err := abi.Pack(abi.Types{abi.Slice(abi.Uint256)}, amounts)
	require.NoError(t, err)

	got, err := DecodeSwapTokensForExactTokens(data)
	require.NoError(t, err)
	require.Len(t, got, 2)
	for i, want := range amounts {
		require.Zero(t, want.Cmp(got[i]))
	}
}

func TestEncodeSwapExactETHForTokens(t *testing.T) {
	path := []*types.Address{testTokenA, testTokenB}
	deadline := big.NewInt(9999999999)
	data, err := EncodeSwapExactETHForTokens(big.NewInt(90), path, testTo, deadline)
	require.NoError(t, err)
	require.Equal(t, []byte{0x7f, 0xf3, 0x6a, 0xb5}, data[:4])

	vals, err := swapExactETHForTokensFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(90).Cmp(vals[0].(*big.Int)))
	requirePath(t, path, vals[1].([]any))
	require.True(t, vals[2].(*types.Address).Equal(testTo))
	require.Zero(t, deadline.Cmp(vals[3].(*big.Int)))
}

func TestDecodeSwapExactETHForTokens(t *testing.T) {
	amounts := []*big.Int{big.NewInt(100), big.NewInt(95)}
	data, err := abi.Pack(abi.Types{abi.Slice(abi.Uint256)}, amounts)
	require.NoError(t, err)

	got, err := DecodeSwapExactETHForTokens(data)
	require.NoError(t, err)
	require.Len(t, got, 2)
	for i, want := range amounts {
		require.Zero(t, want.Cmp(got[i]))
	}
}

func TestEncodeSwapTokensForExactETH(t *testing.T) {
	path := []*types.Address{testTokenA, testTokenB}
	deadline := big.NewInt(9999999999)
	data, err := EncodeSwapTokensForExactETH(big.NewInt(90), big.NewInt(100), path, testTo, deadline)
	require.NoError(t, err)
	require.Equal(t, []byte{0x4a, 0x25, 0xd9, 0x4a}, data[:4])

	vals, err := swapTokensForExactETHFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(90).Cmp(vals[0].(*big.Int)))
	require.Zero(t, big.NewInt(100).Cmp(vals[1].(*big.Int)))
	requirePath(t, path, vals[2].([]any))
	require.True(t, vals[3].(*types.Address).Equal(testTo))
	require.Zero(t, deadline.Cmp(vals[4].(*big.Int)))
}

func TestDecodeSwapTokensForExactETH(t *testing.T) {
	amounts := []*big.Int{big.NewInt(105), big.NewInt(100)}
	data, err := abi.Pack(abi.Types{abi.Slice(abi.Uint256)}, amounts)
	require.NoError(t, err)

	got, err := DecodeSwapTokensForExactETH(data)
	require.NoError(t, err)
	require.Len(t, got, 2)
	for i, want := range amounts {
		require.Zero(t, want.Cmp(got[i]))
	}
}

func TestEncodeSwapExactTokensForETH(t *testing.T) {
	path := []*types.Address{testTokenA, testTokenB}
	deadline := big.NewInt(9999999999)
	data, err := EncodeSwapExactTokensForETH(big.NewInt(100), big.NewInt(90), path, testTo, deadline)
	require.NoError(t, err)
	require.Equal(t, []byte{0x18, 0xcb, 0xaf, 0xe5}, data[:4])

	vals, err := swapExactTokensForETHFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(100).Cmp(vals[0].(*big.Int)))
	require.Zero(t, big.NewInt(90).Cmp(vals[1].(*big.Int)))
	requirePath(t, path, vals[2].([]any))
	require.True(t, vals[3].(*types.Address).Equal(testTo))
	require.Zero(t, deadline.Cmp(vals[4].(*big.Int)))
}

func TestDecodeSwapExactTokensForETH(t *testing.T) {
	amounts := []*big.Int{big.NewInt(100), big.NewInt(95)}
	data, err := abi.Pack(abi.Types{abi.Slice(abi.Uint256)}, amounts)
	require.NoError(t, err)

	got, err := DecodeSwapExactTokensForETH(data)
	require.NoError(t, err)
	require.Len(t, got, 2)
	for i, want := range amounts {
		require.Zero(t, want.Cmp(got[i]))
	}
}

func TestEncodeSwapETHForExactTokens(t *testing.T) {
	path := []*types.Address{testTokenA, testTokenB}
	deadline := big.NewInt(9999999999)
	data, err := EncodeSwapETHForExactTokens(big.NewInt(100), path, testTo, deadline)
	require.NoError(t, err)
	require.Equal(t, []byte{0xfb, 0x3b, 0xdb, 0x41}, data[:4])

	vals, err := swapETHForExactTokensFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(100).Cmp(vals[0].(*big.Int)))
	requirePath(t, path, vals[1].([]any))
	require.True(t, vals[2].(*types.Address).Equal(testTo))
	require.Zero(t, deadline.Cmp(vals[3].(*big.Int)))
}

func TestDecodeSwapETHForExactTokens(t *testing.T) {
	amounts := []*big.Int{big.NewInt(105), big.NewInt(100)}
	data, err := abi.Pack(abi.Types{abi.Slice(abi.Uint256)}, amounts)
	require.NoError(t, err)

	got, err := DecodeSwapETHForExactTokens(data)
	require.NoError(t, err)
	require.Len(t, got, 2)
	for i, want := range amounts {
		require.Zero(t, want.Cmp(got[i]))
	}
}

func TestEncodeRemoveLiquidityETHSupportingFeeOnTransferTokens(t *testing.T) {
	deadline := big.NewInt(9999999999)
	data, err := EncodeRemoveLiquidityETHSupportingFeeOnTransferTokens(testTokenA, big.NewInt(500), big.NewInt(100), big.NewInt(1), testTo, deadline)
	require.NoError(t, err)
	require.Equal(t, []byte{0xaf, 0x29, 0x79, 0xeb}, data[:4])

	vals, err := removeLiquidityETHSupportingFeeOnTransferTokensFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testTokenA))
	require.Zero(t, big.NewInt(500).Cmp(vals[1].(*big.Int)))
	require.Zero(t, big.NewInt(100).Cmp(vals[2].(*big.Int)))
	require.Zero(t, big.NewInt(1).Cmp(vals[3].(*big.Int)))
	require.True(t, vals[4].(*types.Address).Equal(testTo))
	require.Zero(t, deadline.Cmp(vals[5].(*big.Int)))
}

func TestDecodeRemoveLiquidityETHSupportingFeeOnTransferTokens(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(5))
	require.NoError(t, err)

	got, err := DecodeRemoveLiquidityETHSupportingFeeOnTransferTokens(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(5).Cmp(got))
}

func TestEncodeRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(t *testing.T) {
	sig := testRouterSignature(t)
	deadline := big.NewInt(9999999999)
	data, err := EncodeRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(testTokenA, big.NewInt(500), big.NewInt(100), big.NewInt(1), testTo, deadline, true, sig)
	require.NoError(t, err)
	require.Equal(t, []byte{0x5b, 0x0d, 0x59, 0x84}, data[:4])

	vals, err := removeLiquidityETHWithPermitSupportingFeeOnTransferTokensFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testTokenA))
	require.Zero(t, big.NewInt(500).Cmp(vals[1].(*big.Int)))
	require.Zero(t, big.NewInt(100).Cmp(vals[2].(*big.Int)))
	require.Zero(t, big.NewInt(1).Cmp(vals[3].(*big.Int)))
	require.True(t, vals[4].(*types.Address).Equal(testTo))
	require.Zero(t, deadline.Cmp(vals[5].(*big.Int)))
	require.True(t, vals[6].(bool))
	requirePermitParts(t, sig, vals, 7)
}

func TestDecodeRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(5))
	require.NoError(t, err)

	got, err := DecodeRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(5).Cmp(got))
}

func TestEncodeSwapExactTokensForTokensSupportingFeeOnTransferTokens(t *testing.T) {
	path := []*types.Address{testTokenA, testTokenB}
	deadline := big.NewInt(9999999999)
	data, err := EncodeSwapExactTokensForTokensSupportingFeeOnTransferTokens(big.NewInt(100), big.NewInt(90), path, testTo, deadline)
	require.NoError(t, err)
	require.Equal(t, []byte{0x5c, 0x11, 0xd7, 0x95}, data[:4])

	vals, err := swapExactTokensForTokensSupportingFeeOnTransferTokensFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(100).Cmp(vals[0].(*big.Int)))
	require.Zero(t, big.NewInt(90).Cmp(vals[1].(*big.Int)))
	requirePath(t, path, vals[2].([]any))
	require.True(t, vals[3].(*types.Address).Equal(testTo))
	require.Zero(t, deadline.Cmp(vals[4].(*big.Int)))
}

func TestEncodeSwapExactETHForTokensSupportingFeeOnTransferTokens(t *testing.T) {
	path := []*types.Address{testTokenA, testTokenB}
	deadline := big.NewInt(9999999999)
	data, err := EncodeSwapExactETHForTokensSupportingFeeOnTransferTokens(big.NewInt(90), path, testTo, deadline)
	require.NoError(t, err)
	require.Equal(t, []byte{0xb6, 0xf9, 0xde, 0x95}, data[:4])

	vals, err := swapExactETHForTokensSupportingFeeOnTransferTokensFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(90).Cmp(vals[0].(*big.Int)))
	requirePath(t, path, vals[1].([]any))
	require.True(t, vals[2].(*types.Address).Equal(testTo))
	require.Zero(t, deadline.Cmp(vals[3].(*big.Int)))
}

func TestEncodeSwapExactTokensForETHSupportingFeeOnTransferTokens(t *testing.T) {
	path := []*types.Address{testTokenA, testTokenB}
	deadline := big.NewInt(9999999999)
	data, err := EncodeSwapExactTokensForETHSupportingFeeOnTransferTokens(big.NewInt(100), big.NewInt(90), path, testTo, deadline)
	require.NoError(t, err)
	require.Equal(t, []byte{0x79, 0x1a, 0xc9, 0x47}, data[:4])

	vals, err := swapExactTokensForETHSupportingFeeOnTransferTokensFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(100).Cmp(vals[0].(*big.Int)))
	require.Zero(t, big.NewInt(90).Cmp(vals[1].(*big.Int)))
	requirePath(t, path, vals[2].([]any))
	require.True(t, vals[3].(*types.Address).Equal(testTo))
	require.Zero(t, deadline.Cmp(vals[4].(*big.Int)))
}

func TestRouterRevertReasons(t *testing.T) {
	reasons := []string{
		reasonExpired, reasonInsufficientAAmount, reasonInsufficientBAmount,
		reasonRouterInsufficientOutput, reasonExcessiveInputAmount, reasonRouterInvalidPath,
		reasonLibraryIdenticalAddressesOld, reasonLibraryIdenticalAddressesNew,
		reasonLibraryZeroAddressOld, reasonLibraryZeroAddressNew,
		reasonLibraryInsufficientAmountOld, reasonLibraryInsufficientAmountNew,
		reasonLibraryInsufficientLiquidityOld, reasonLibraryInsufficientLiquidityNew,
		reasonLibraryInsufficientInputOld, reasonLibraryInsufficientInputNew,
		reasonLibraryInsufficientOutputOld, reasonLibraryInsufficientOutputNew,
		reasonLibraryInvalidPathOld, reasonLibraryInvalidPathNew,
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
