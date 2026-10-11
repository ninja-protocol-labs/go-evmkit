package erc1155

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/standards/multicall3"
)

var (
	testAddr1  = types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	testAddr2  = types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	testAddr3  = types.MustNewAddressFromHex("0xb47e3cd837ddf8e4c57f05d70ab865de6e193bbb")
	testTokens = types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
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

func TestEncodeBalanceOfBatch(t *testing.T) {
	accounts := []*types.Address{testAddr1, testAddr2}
	ids := []*big.Int{big.NewInt(1), big.NewInt(2)}

	data, err := EncodeBalanceOfBatch(accounts, ids)
	require.NoError(t, err)
	require.Equal(t, []byte{0x4e, 0x12, 0x73, 0xf4}, data[:4])

	vals, err := balanceOfBatchFn.DecodeCall(data)
	require.NoError(t, err)
	gotAccounts, ok := vals[0].([]any)
	require.True(t, ok)
	require.Len(t, gotAccounts, 2)
	require.True(t, gotAccounts[0].(*types.Address).Equal(testAddr1))
	require.True(t, gotAccounts[1].(*types.Address).Equal(testAddr2))
	gotIds, ok := vals[1].([]any)
	require.True(t, ok)
	require.Len(t, gotIds, 2)
	require.Zero(t, big.NewInt(1).Cmp(gotIds[0].(*big.Int)))
	require.Zero(t, big.NewInt(2).Cmp(gotIds[1].(*big.Int)))
}

func TestEncodeBalanceOfBatchRejectsLengthMismatch(t *testing.T) {
	_, err := EncodeBalanceOfBatch([]*types.Address{testAddr1}, []*big.Int{big.NewInt(1), big.NewInt(2)})
	require.Error(t, err)
}

func TestDecodeBalanceOfBatch(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Slice(abi.Uint256)}, []*big.Int{big.NewInt(100), big.NewInt(200)})
	require.NoError(t, err)

	amounts, err := DecodeBalanceOfBatch(data)
	require.NoError(t, err)
	require.Len(t, amounts, 2)
	require.Zero(t, big.NewInt(100).Cmp(amounts[0]))
	require.Zero(t, big.NewInt(200).Cmp(amounts[1]))
}

func TestDecodeBalanceOfBatchEmpty(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Slice(abi.Uint256)}, []*big.Int{})
	require.NoError(t, err)

	amounts, err := DecodeBalanceOfBatch(data)
	require.NoError(t, err)
	require.NotNil(t, amounts)
	require.Empty(t, amounts)
}

func TestDecodeBalanceOfBatchRejectsGarbage(t *testing.T) {
	_, err := DecodeBalanceOfBatch([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestDecodeIdBalances(t *testing.T) {
	ids := []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3)}
	data, err := abi.Pack(abi.Types{abi.Slice(abi.Uint256)}, []*big.Int{big.NewInt(10), big.NewInt(20), big.NewInt(30)})
	require.NoError(t, err)

	balances, err := DecodeIdBalances(ids, data)
	require.NoError(t, err)
	require.Len(t, balances, 3)
	for i, id := range ids {
		require.Same(t, id, balances[i].Id)
		require.Zero(t, big.NewInt(int64((i+1)*10)).Cmp(balances[i].Amount))
	}
}

func TestDecodeIdBalancesRejectsLengthMismatch(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Slice(abi.Uint256)}, []*big.Int{big.NewInt(10)})
	require.NoError(t, err)

	_, err = DecodeIdBalances([]*big.Int{big.NewInt(1), big.NewInt(2)}, data)
	require.Error(t, err)
}

func TestDecodeIdBalancePairs(t *testing.T) {
	accounts := []*types.Address{testAddr1, testAddr2}
	ids := []*big.Int{big.NewInt(1), big.NewInt(2)}
	data, err := abi.Pack(abi.Types{abi.Slice(abi.Uint256)}, []*big.Int{big.NewInt(10), big.NewInt(20)})
	require.NoError(t, err)

	pairs, err := DecodeIdBalancePairs(accounts, ids, data)
	require.NoError(t, err)
	require.Len(t, pairs, 2)
	require.Same(t, accounts[0], pairs[0].Account)
	require.Same(t, ids[0], pairs[0].Id)
	require.Zero(t, big.NewInt(10).Cmp(pairs[0].Amount))
	require.Same(t, accounts[1], pairs[1].Account)
	require.Same(t, ids[1], pairs[1].Id)
	require.Zero(t, big.NewInt(20).Cmp(pairs[1].Amount))
}

func TestDecodeIdBalancePairsRejectsAccountsIdsLengthMismatch(t *testing.T) {
	_, err := DecodeIdBalancePairs([]*types.Address{testAddr1}, []*big.Int{big.NewInt(1), big.NewInt(2)}, []byte{})
	require.Error(t, err)
}

func TestDecodeIdBalancePairsRejectsResultLengthMismatch(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Slice(abi.Uint256)}, []*big.Int{big.NewInt(10)})
	require.NoError(t, err)

	_, err = DecodeIdBalancePairs([]*types.Address{testAddr1, testAddr2}, []*big.Int{big.NewInt(1), big.NewInt(2)}, data)
	require.Error(t, err)
}

func TestEncodeIsApprovedForAll(t *testing.T) {
	data, err := EncodeIsApprovedForAll(testAddr1, testAddr2)
	require.NoError(t, err)
	require.Equal(t, []byte{0xe9, 0x85, 0xe9, 0xc5}, data[:4])

	vals, err := isApprovedForAllFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testAddr1))
	require.True(t, vals[1].(*types.Address).Equal(testAddr2))
}

func TestEncodeIsApprovedForAllRejectsNilArgs(t *testing.T) {
	_, err := EncodeIsApprovedForAll(nil, testAddr2)
	require.Error(t, err)
	_, err = EncodeIsApprovedForAll(testAddr1, nil)
	require.Error(t, err)
}

func TestDecodeIsApprovedForAll(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Bool}, true)
	require.NoError(t, err)
	ok, err := DecodeIsApprovedForAll(data)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestDecodeIsApprovedForAllRejectsGarbage(t *testing.T) {
	_, err := DecodeIsApprovedForAll([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeSetApprovalForAll(t *testing.T) {
	data, err := EncodeSetApprovalForAll(testAddr1, true)
	require.NoError(t, err)
	require.Equal(t, []byte{0xa2, 0x2c, 0xb4, 0x65}, data[:4])

	vals, err := setApprovalForAllFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testAddr1))
	require.Equal(t, true, vals[1].(bool))
}

func TestEncodeSetApprovalForAllRejectsNilAddress(t *testing.T) {
	_, err := EncodeSetApprovalForAll(nil, true)
	require.Error(t, err)
}

func TestEncodeSafeTransferFrom(t *testing.T) {
	callData := []byte{0xde, 0xad, 0xbe, 0xef}
	data, err := EncodeSafeTransferFrom(testAddr1, testAddr2, big.NewInt(1), big.NewInt(100), callData)
	require.NoError(t, err)
	require.Equal(t, []byte{0xf2, 0x42, 0x43, 0x2a}, data[:4])

	vals, err := safeTransferFromFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testAddr1))
	require.True(t, vals[1].(*types.Address).Equal(testAddr2))
	require.Zero(t, big.NewInt(1).Cmp(vals[2].(*big.Int)))
	require.Zero(t, big.NewInt(100).Cmp(vals[3].(*big.Int)))
	require.Equal(t, callData, vals[4].([]byte))
}

func TestEncodeSafeTransferFromRejectsNilArgs(t *testing.T) {
	_, err := EncodeSafeTransferFrom(nil, testAddr2, big.NewInt(1), big.NewInt(100), nil)
	require.Error(t, err)
	_, err = EncodeSafeTransferFrom(testAddr1, nil, big.NewInt(1), big.NewInt(100), nil)
	require.Error(t, err)
	_, err = EncodeSafeTransferFrom(testAddr1, testAddr2, nil, big.NewInt(100), nil)
	require.Error(t, err)
	_, err = EncodeSafeTransferFrom(testAddr1, testAddr2, big.NewInt(1), nil, nil)
	require.Error(t, err)
}

func TestEncodeSafeBatchTransferFrom(t *testing.T) {
	ids := []*big.Int{big.NewInt(1), big.NewInt(2)}
	values := []*big.Int{big.NewInt(10), big.NewInt(20)}
	callData := []byte{0xde, 0xad, 0xbe, 0xef}

	data, err := EncodeSafeBatchTransferFrom(testAddr1, testAddr2, ids, values, callData)
	require.NoError(t, err)
	require.Equal(t, []byte{0x2e, 0xb2, 0xc2, 0xd6}, data[:4])

	vals, err := safeBatchTransferFromFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testAddr1))
	require.True(t, vals[1].(*types.Address).Equal(testAddr2))
	gotIds, ok := vals[2].([]any)
	require.True(t, ok)
	require.Zero(t, big.NewInt(1).Cmp(gotIds[0].(*big.Int)))
	require.Zero(t, big.NewInt(2).Cmp(gotIds[1].(*big.Int)))
	gotValues, ok := vals[3].([]any)
	require.True(t, ok)
	require.Zero(t, big.NewInt(10).Cmp(gotValues[0].(*big.Int)))
	require.Zero(t, big.NewInt(20).Cmp(gotValues[1].(*big.Int)))
	require.Equal(t, callData, vals[4].([]byte))
}

func TestEncodeSafeBatchTransferFromRejectsLengthMismatch(t *testing.T) {
	_, err := EncodeSafeBatchTransferFrom(testAddr1, testAddr2, []*big.Int{big.NewInt(1), big.NewInt(2)}, []*big.Int{big.NewInt(10)}, nil)
	require.Error(t, err)
}

func TestEncodeSafeBatchTransferFromRejectsNilArgs(t *testing.T) {
	ids := []*big.Int{big.NewInt(1)}
	values := []*big.Int{big.NewInt(10)}
	_, err := EncodeSafeBatchTransferFrom(nil, testAddr2, ids, values, nil)
	require.Error(t, err)
	_, err = EncodeSafeBatchTransferFrom(testAddr1, nil, ids, values, nil)
	require.Error(t, err)
}

func TestEncodeURI(t *testing.T) {
	data, err := EncodeURI(big.NewInt(42))
	require.NoError(t, err)
	require.Equal(t, []byte{0x0e, 0x89, 0x34, 0x1c}, data[:4])

	vals, err := uriFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(42).Cmp(vals[0].(*big.Int)))
}

func TestEncodeURIRejectsNilId(t *testing.T) {
	_, err := EncodeURI(nil)
	require.Error(t, err)
}

func TestDecodeURI(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.String}, "ipfs://Qm.../{id}")
	require.NoError(t, err)
	s, err := DecodeURI(data)
	require.NoError(t, err)
	require.Equal(t, "ipfs://Qm.../{id}", s)
}

func TestDecodeURIRejectsGarbage(t *testing.T) {
	_, err := DecodeURI([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestCustomErrorSignatures(t *testing.T) {
	tests := []struct {
		name string
		err  *abi.Error
		sig  string
	}{
		{"ERC1155InsufficientBalance", errInsufficientBalance, "ERC1155InsufficientBalance(address,uint256,uint256,uint256)"},
		{"ERC1155InvalidSender", errInvalidSender, "ERC1155InvalidSender(address)"},
		{"ERC1155InvalidReceiver", errInvalidReceiver, "ERC1155InvalidReceiver(address)"},
		{"ERC1155MissingApprovalForAll", errMissingApprovalForAll, "ERC1155MissingApprovalForAll(address,address)"},
		{"ERC1155InvalidApprover", errInvalidApprover, "ERC1155InvalidApprover(address)"},
		{"ERC1155InvalidOperator", errInvalidOperator, "ERC1155InvalidOperator(address)"},
		{"ERC1155InvalidArrayLength", errInvalidArrayLength, "ERC1155InvalidArrayLength(uint256,uint256)"},
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

func TestTransferSingleEventSignature(t *testing.T) {
	require.Equal(t, "TransferSingle(address,address,address,uint256,uint256)", transferSingleEvent.Signature())

	t0, ok := transferSingleEvent.Topic0()
	require.True(t, ok)
	require.Equal(t, "0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62", t0.String())

	require.False(t, transferSingleEvent.Anonymous)
	require.Len(t, transferSingleEvent.Inputs, 5)
	require.True(t, transferSingleEvent.Inputs[0].Indexed)
	require.True(t, transferSingleEvent.Inputs[1].Indexed)
	require.True(t, transferSingleEvent.Inputs[2].Indexed)
	require.False(t, transferSingleEvent.Inputs[3].Indexed)
	require.False(t, transferSingleEvent.Inputs[4].Indexed)
}

func TestTransferBatchEventSignature(t *testing.T) {
	require.Equal(t, "TransferBatch(address,address,address,uint256[],uint256[])", transferBatchEvent.Signature())

	t0, ok := transferBatchEvent.Topic0()
	require.True(t, ok)
	require.Equal(t, "0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb", t0.String())

	require.False(t, transferBatchEvent.Anonymous)
	require.Len(t, transferBatchEvent.Inputs, 5)
	require.True(t, transferBatchEvent.Inputs[0].Indexed)
	require.True(t, transferBatchEvent.Inputs[1].Indexed)
	require.True(t, transferBatchEvent.Inputs[2].Indexed)
	require.False(t, transferBatchEvent.Inputs[3].Indexed)
	require.False(t, transferBatchEvent.Inputs[4].Indexed)
}

func TestApprovalForAllEventSignature(t *testing.T) {
	require.Equal(t, "ApprovalForAll(address,address,bool)", approvalForAllEvent.Signature())

	t0, ok := approvalForAllEvent.Topic0()
	require.True(t, ok)
	require.Equal(t, "0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31", t0.String())

	require.False(t, approvalForAllEvent.Anonymous)
	require.Len(t, approvalForAllEvent.Inputs, 3)
	require.True(t, approvalForAllEvent.Inputs[0].Indexed)
	require.True(t, approvalForAllEvent.Inputs[1].Indexed)
	require.False(t, approvalForAllEvent.Inputs[2].Indexed)
}

func TestURIEventSignature(t *testing.T) {
	require.Equal(t, "URI(string,uint256)", uriEvent.Signature())

	t0, ok := uriEvent.Topic0()
	require.True(t, ok)
	require.Equal(t, "0x6bb7ff708619ba0610cba295a58592e0451dee2622938c8755667688daf3529b", t0.String())

	require.False(t, uriEvent.Anonymous)
	require.Len(t, uriEvent.Inputs, 2)
	require.False(t, uriEvent.Inputs[0].Indexed)
	require.True(t, uriEvent.Inputs[1].Indexed)
}

func TestDecodeApprovalForAllWithBalance(t *testing.T) {
	approvedData, _ := abi.Pack(abi.Types{abi.Bool}, true)
	accountBalData, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(3))
	operatorBalData, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(5))

	data := encodeResultsData(t,
		[]bool{true, true, true},
		[][]byte{approvedData, accountBalData, operatorBalData},
	)

	id := big.NewInt(7)
	afawb, err := DecodeApprovalForAllWithBalance(testAddr1, testAddr2, id, data)
	require.NoError(t, err)
	require.True(t, afawb.Account.Equal(testAddr1))
	require.True(t, afawb.Operator.Equal(testAddr2))
	require.Same(t, id, afawb.Id)
	require.True(t, afawb.Approved)
	require.Zero(t, big.NewInt(3).Cmp(afawb.AccountBalance))
	require.Zero(t, big.NewInt(5).Cmp(afawb.OperatorBalance))
}

func TestDecodeApprovalForAllWithBalancePartialFailure(t *testing.T) {
	approvedData, _ := abi.Pack(abi.Types{abi.Bool}, true)

	data := encodeResultsData(t,
		[]bool{true, false, false},
		[][]byte{approvedData, nil, nil},
	)

	afawb, err := DecodeApprovalForAllWithBalance(testAddr1, testAddr2, big.NewInt(7), data)
	require.NoError(t, err)
	require.True(t, afawb.Approved)
	require.Nil(t, afawb.AccountBalance)
	require.Nil(t, afawb.OperatorBalance)
}

func TestDecodeApprovalForAllWithBalanceRejectsWrongResultCount(t *testing.T) {
	data := encodeResultsData(t, []bool{true, true}, [][]byte{{}, {}})
	_, err := DecodeApprovalForAllWithBalance(testAddr1, testAddr2, big.NewInt(7), data)
	require.Error(t, err)
}

func TestEncodeApprovalForAllWithBalanceMatchesManualAggregate3(t *testing.T) {
	id := big.NewInt(7)
	got, err := EncodeApprovalForAllWithBalance(testTokens, testAddr1, testAddr2, id, true)
	require.NoError(t, err)

	approvedCalldata, err := EncodeIsApprovedForAll(testAddr1, testAddr2)
	require.NoError(t, err)
	accountBalCalldata, err := EncodeBalanceOf(testAddr1, id)
	require.NoError(t, err)
	operatorBalCalldata, err := EncodeBalanceOf(testAddr2, id)
	require.NoError(t, err)

	want, err := multicall3.EncodeAggregate3([]multicall3.Call3{
		multicall3.NewCall3(testTokens, true, approvedCalldata),
		multicall3.NewCall3(testTokens, true, accountBalCalldata),
		multicall3.NewCall3(testTokens, true, operatorBalCalldata),
	})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDecodeAccountIdBalances(t *testing.T) {
	contracts := []*types.Address{testTokens, testAddr3}
	bal1, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(3))
	bal2, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(4))

	data := encodeResultsData(t, []bool{true, true}, [][]byte{bal1, bal2})

	id := big.NewInt(9)
	aib, err := DecodeAccountIdBalances(testAddr1, id, contracts, data)
	require.NoError(t, err)
	require.True(t, aib.Account.Equal(testAddr1))
	require.Same(t, id, aib.Id)
	require.Len(t, aib.Contracts, 2)
	require.True(t, aib.Contracts[0].Contract.Equal(testTokens))
	require.Zero(t, big.NewInt(3).Cmp(aib.Contracts[0].Amount))
	require.True(t, aib.Contracts[1].Contract.Equal(testAddr3))
	require.Zero(t, big.NewInt(4).Cmp(aib.Contracts[1].Amount))
}

func TestDecodeAccountIdBalancesPartialFailure(t *testing.T) {
	bal1, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(3))

	data := encodeResultsData(t, []bool{true, false}, [][]byte{bal1, nil})

	aib, err := DecodeAccountIdBalances(testAddr1, big.NewInt(9), []*types.Address{testTokens, testAddr3}, data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(3).Cmp(aib.Contracts[0].Amount))
	require.Nil(t, aib.Contracts[1].Amount)
}

func TestDecodeAccountIdBalancesRejectsLengthMismatch(t *testing.T) {
	data := encodeResultsData(t, []bool{true}, [][]byte{{}})
	_, err := DecodeAccountIdBalances(testAddr1, big.NewInt(9), []*types.Address{testTokens, testAddr3}, data)
	require.Error(t, err)
}

func TestEncodeAccountIdBalancesMatchesManualAggregate3(t *testing.T) {
	contracts := []*types.Address{testTokens, testAddr3}
	id := big.NewInt(9)
	got, err := EncodeAccountIdBalances(testAddr1, id, contracts, true)
	require.NoError(t, err)

	calldata, err := EncodeBalanceOf(testAddr1, id)
	require.NoError(t, err)

	want, err := multicall3.EncodeAggregate3([]multicall3.Call3{
		multicall3.NewCall3(testTokens, true, calldata),
		multicall3.NewCall3(testAddr3, true, calldata),
	})
	require.NoError(t, err)
	require.Equal(t, want, got)
}
