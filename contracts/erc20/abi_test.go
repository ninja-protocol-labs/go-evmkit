package erc20

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/contracts/multicall3"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func encodeResultsData(t *testing.T, successes []bool, datas [][]byte) []byte {
	t.Helper()
	require.Len(t, datas, len(successes))

	resultT := abi.Tuple(abi.Bool, abi.Bytes)
	elems := make([]any, len(successes))
	for i := range successes {
		elems[i] = []any{successes[i], datas[i]}
	}
	data, err := abi.Pack(abi.Types{abi.Slice(resultT)}, elems)
	require.NoError(t, err)
	return data
}

var (
	testToken  = mustAddrPanic("0x4200000000000000000000000000000000000006")
	testToken2 = mustAddrPanic("0x833589fcd6edb6e08f4c7c32d4f71b54bda02913")
	testAddr1  = mustAddrPanic("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	testAddr2  = mustAddrPanic("0xba12222222228d8ba445958a75a0704d566bf2c8")
)

// --- simple no-arg functions ---

func TestEncodeNoArgFunctions(t *testing.T) {
	require.Equal(t, []byte{0x06, 0xfd, 0xde, 0x03}, EncodeName())
	require.Equal(t, []byte{0x95, 0xd8, 0x9b, 0x41}, EncodeSymbol())
	require.Equal(t, []byte{0x31, 0x3c, 0xe5, 0x67}, EncodeDecimals())
	require.Equal(t, []byte{0x18, 0x16, 0x0d, 0xdd}, EncodeTotalSupply())
}

func TestDecodeNameSymbol(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.String}, "USD Coin")
	require.NoError(t, err)
	s, err := DecodeName(data)
	require.NoError(t, err)
	require.Equal(t, "USD Coin", s)

	data, err = abi.Pack(abi.Types{abi.String}, "USDC")
	require.NoError(t, err)
	s, err = DecodeSymbol(data)
	require.NoError(t, err)
	require.Equal(t, "USDC", s)
}

func TestDecodeDecimals(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint8}, uint8(6))
	require.NoError(t, err)
	d, err := DecodeDecimals(data)
	require.NoError(t, err)
	require.Equal(t, uint8(6), d)
}

func TestDecodeTotalSupply(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(1_000_000))
	require.NoError(t, err)
	n, err := DecodeTotalSupply(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1_000_000).Cmp(n))
}

func TestSimpleDecodersRejectGarbage(t *testing.T) {
	garbage := []byte{0x01, 0x02}
	_, err := DecodeName(garbage)
	require.Error(t, err)
	_, err = DecodeDecimals(garbage)
	require.Error(t, err)
	_, err = DecodeTotalSupply(garbage)
	require.Error(t, err)
	_, err = DecodeBalanceOf(garbage)
	require.Error(t, err)
}

func TestEncodeBalanceOf(t *testing.T) {
	data, err := EncodeBalanceOf(testAddr1)
	require.NoError(t, err)
	require.Equal(t, []byte{0x70, 0xa0, 0x82, 0x31}, data[:4])

	vals, err := balanceOfFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testAddr1))
}

func TestEncodeBalanceOfRejectsNilAddress(t *testing.T) {
	_, err := EncodeBalanceOf(nil)
	require.Error(t, err)
}

func TestDecodeBalanceOf(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(42))
	require.NoError(t, err)
	n, err := DecodeBalanceOf(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(42).Cmp(n))
}

func TestEncodeAllowance(t *testing.T) {
	data, err := EncodeAllowance(testAddr1, testAddr2)
	require.NoError(t, err)

	vals, err := allowanceFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testAddr1))
	require.True(t, vals[1].(*types.Address).Equal(testAddr2))
}

func TestEncodeTransferDecodeTransfer(t *testing.T) {
	data, err := EncodeTransfer(testAddr1, big.NewInt(100))
	require.NoError(t, err)
	vals, err := transferFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testAddr1))
	require.Zero(t, big.NewInt(100).Cmp(vals[1].(*big.Int)))

	retData, err := abi.Pack(abi.Types{abi.Bool}, true)
	require.NoError(t, err)
	ok, err := DecodeTransfer(retData)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestEncodeApproveDecodeApprove(t *testing.T) {
	data, err := EncodeApprove(testAddr2, big.NewInt(5))
	require.NoError(t, err)
	vals, err := approveFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testAddr2))
	require.Zero(t, big.NewInt(5).Cmp(vals[1].(*big.Int)))

	retData, err := abi.Pack(abi.Types{abi.Bool}, false)
	require.NoError(t, err)
	ok, err := DecodeApprove(retData)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestEncodeTransferFromDecodeTransferFrom(t *testing.T) {
	data, err := EncodeTransferFrom(testAddr1, testAddr2, big.NewInt(9))
	require.NoError(t, err)
	vals, err := transferFromFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testAddr1))
	require.True(t, vals[1].(*types.Address).Equal(testAddr2))
	require.Zero(t, big.NewInt(9).Cmp(vals[2].(*big.Int)))

	retData, err := abi.Pack(abi.Types{abi.Bool}, true)
	require.NoError(t, err)
	ok, err := DecodeTransferFrom(retData)
	require.NoError(t, err)
	require.True(t, ok)
}

// --- Metadata ---

func TestDecodeMetadataAllSuccess(t *testing.T) {
	nameData, _ := abi.Pack(abi.Types{abi.String}, "USD Coin")
	symbolData, _ := abi.Pack(abi.Types{abi.String}, "USDC")
	decimalsData, _ := abi.Pack(abi.Types{abi.Uint8}, uint8(6))
	totalSupplyData, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(1000))

	data := encodeResultsData(t,
		[]bool{true, true, true, true},
		[][]byte{nameData, symbolData, decimalsData, totalSupplyData},
	)

	m, err := DecodeMetadata(data)
	require.NoError(t, err)
	require.Equal(t, "USD Coin", m.Name)
	require.Equal(t, "USDC", m.Symbol)
	require.Equal(t, uint8(6), m.Decimals)
	require.Zero(t, big.NewInt(1000).Cmp(m.TotalSupply))
}

func TestDecodeMetadataPartialFailure(t *testing.T) {
	nameData, _ := abi.Pack(abi.Types{abi.String}, "USD Coin")
	totalSupplyData, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(1000))

	data := encodeResultsData(t,
		[]bool{true, false, false, true},
		[][]byte{nameData, nil, nil, totalSupplyData},
	)

	m, err := DecodeMetadata(data)
	require.NoError(t, err)
	require.Equal(t, "USD Coin", m.Name)
	require.Equal(t, "", m.Symbol)
	require.Equal(t, uint8(0), m.Decimals)
	require.Zero(t, big.NewInt(1000).Cmp(m.TotalSupply))
}

func TestDecodeMetadataRejectsWrongResultCount(t *testing.T) {
	data := encodeResultsData(t, []bool{true}, [][]byte{{}})
	_, err := DecodeMetadata(data)
	require.Error(t, err)
}

func TestDecodeMetadataWithBalance(t *testing.T) {
	nameData, _ := abi.Pack(abi.Types{abi.String}, "USD Coin")
	symbolData, _ := abi.Pack(abi.Types{abi.String}, "USDC")
	decimalsData, _ := abi.Pack(abi.Types{abi.Uint8}, uint8(6))
	totalSupplyData, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(1000))
	balData, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(42))

	data := encodeResultsData(t,
		[]bool{true, true, true, true, true},
		[][]byte{nameData, symbolData, decimalsData, totalSupplyData, balData},
	)

	mb, err := DecodeMetadataWithBalance(data)
	require.NoError(t, err)
	require.Equal(t, "USD Coin", mb.Name)
	require.Zero(t, big.NewInt(42).Cmp(mb.Balance))
}

func TestDecodeMetadataWithBalanceFailedBalance(t *testing.T) {
	nameData, _ := abi.Pack(abi.Types{abi.String}, "USD Coin")
	symbolData, _ := abi.Pack(abi.Types{abi.String}, "USDC")
	decimalsData, _ := abi.Pack(abi.Types{abi.Uint8}, uint8(6))
	totalSupplyData, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(1000))

	data := encodeResultsData(t,
		[]bool{true, true, true, true, false},
		[][]byte{nameData, symbolData, decimalsData, totalSupplyData, nil},
	)

	mb, err := DecodeMetadataWithBalance(data)
	require.NoError(t, err)
	require.Nil(t, mb.Balance)
}

func TestDecodeAllowanceWithBalance(t *testing.T) {
	allowanceData, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(5))
	ownerBalData, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(10))
	spenderBalData, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(20))

	data := encodeResultsData(t,
		[]bool{true, true, true},
		[][]byte{allowanceData, ownerBalData, spenderBalData},
	)

	awb, err := DecodeAllowanceWithBalance(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(5).Cmp(awb.Value))
	require.Zero(t, big.NewInt(10).Cmp(awb.OwnerBalance))
	require.Zero(t, big.NewInt(20).Cmp(awb.SpenderBalance))
}

func TestDecodeAllowanceWithBalanceRejectsWrongResultCount(t *testing.T) {
	data := encodeResultsData(t, []bool{true, true}, [][]byte{{}, {}})
	_, err := DecodeAllowanceWithBalance(data)
	require.Error(t, err)
}

func TestDecodeTokenBalances(t *testing.T) {
	addrs := []*types.Address{testAddr1, testAddr2}
	bal1, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(1))
	bal2, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(2))

	data := encodeResultsData(t, []bool{true, false}, [][]byte{bal1, bal2})

	tb, err := DecodeTokenBalances(testToken, addrs, data)
	require.NoError(t, err)
	require.True(t, tb.Token.Equal(testToken))
	require.Len(t, tb.Addresses, 2)
	require.True(t, tb.Addresses[0].Address.Equal(testAddr1))
	require.Zero(t, big.NewInt(1).Cmp(tb.Addresses[0].Amount))
	require.True(t, tb.Addresses[1].Address.Equal(testAddr2))
	require.Nil(t, tb.Addresses[1].Amount)
}

func TestDecodeTokenBalancesRejectsLengthMismatch(t *testing.T) {
	data := encodeResultsData(t, []bool{true}, [][]byte{{}})
	_, err := DecodeTokenBalances(testToken, []*types.Address{testAddr1, testAddr2}, data)
	require.Error(t, err)
}

func TestDecodeAddressBalances(t *testing.T) {
	toks := []*types.Address{testToken, testToken2}
	bal1, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(3))
	bal2, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(4))

	data := encodeResultsData(t, []bool{true, true}, [][]byte{bal1, bal2})

	ab, err := DecodeAddressBalances(testAddr1, toks, data)
	require.NoError(t, err)
	require.True(t, ab.Address.Equal(testAddr1))
	require.Len(t, ab.Tokens, 2)
	require.True(t, ab.Tokens[0].Token.Equal(testToken))
	require.Zero(t, big.NewInt(3).Cmp(ab.Tokens[0].Amount))
	require.True(t, ab.Tokens[1].Token.Equal(testToken2))
	require.Zero(t, big.NewInt(4).Cmp(ab.Tokens[1].Amount))
}

func TestEncodeBalancePairsRejectsLengthMismatch(t *testing.T) {
	_, err := EncodeBalancePairs([]*types.Address{testToken}, []*types.Address{testAddr1, testAddr2}, false)
	require.Error(t, err)
}

func TestDecodeBalancePairsRejectsLengthMismatch(t *testing.T) {
	_, err := DecodeBalancePairs([]*types.Address{testToken}, []*types.Address{testAddr1, testAddr2}, []byte{})
	require.Error(t, err)
}

func TestDecodeBalancePairs(t *testing.T) {
	toks := []*types.Address{testToken, testToken, testToken2}
	addrs := []*types.Address{testAddr1, testAddr2, testAddr2}
	bal1, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(1))
	bal2, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(2))
	bal3, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(3))

	data := encodeResultsData(t, []bool{true, true, true}, [][]byte{bal1, bal2, bal3})

	pairs, err := DecodeBalancePairs(toks, addrs, data)
	require.NoError(t, err)
	require.Len(t, pairs, 3)
	require.True(t, pairs[0].Token.Equal(testToken))
	require.True(t, pairs[0].Address.Equal(testAddr1))
	require.Zero(t, big.NewInt(1).Cmp(pairs[0].Amount))
	require.True(t, pairs[2].Token.Equal(testToken2))
	require.True(t, pairs[2].Address.Equal(testAddr2))
	require.Zero(t, big.NewInt(3).Cmp(pairs[2].Amount))
}

func TestEncodeMetadataMatchesManualAggregate3(t *testing.T) {
	got, err := EncodeMetadata(testToken, true)
	require.NoError(t, err)

	want, err := multicall3.EncodeAggregate3([]multicall3.Call3{
		multicall3.NewCall3(testToken, true, EncodeName()),
		multicall3.NewCall3(testToken, true, EncodeSymbol()),
		multicall3.NewCall3(testToken, true, EncodeDecimals()),
		multicall3.NewCall3(testToken, true, EncodeTotalSupply()),
	})
	require.NoError(t, err)
	require.Equal(t, want, got)
}
