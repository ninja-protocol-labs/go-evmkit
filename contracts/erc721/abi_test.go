package erc721

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/contracts/multicall3"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

var (
	testAddr1   = types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	testAddr2   = types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	cryptopunks = types.MustNewAddressFromHex("0xb47e3cd837ddf8e4c57f05d70ab865de6e193bbb")
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

func TestEncodeNoArgFunctions(t *testing.T) {
	require.Equal(t, []byte{0x06, 0xfd, 0xde, 0x03}, EncodeName())
	require.Equal(t, []byte{0x95, 0xd8, 0x9b, 0x41}, EncodeSymbol())
}

func TestDecodeNameSymbol(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.String}, "Bored Ape Yacht Club")
	require.NoError(t, err)
	s, err := DecodeName(data)
	require.NoError(t, err)
	require.Equal(t, "Bored Ape Yacht Club", s)

	data, err = abi.Pack(abi.Types{abi.String}, "BAYC")
	require.NoError(t, err)
	s, err = DecodeSymbol(data)
	require.NoError(t, err)
	require.Equal(t, "BAYC", s)
}

func TestSimpleDecodersRejectGarbage(t *testing.T) {
	garbage := []byte{0x01, 0x02}
	_, err := DecodeName(garbage)
	require.Error(t, err)
	_, err = DecodeSymbol(garbage)
	require.Error(t, err)
	_, err = DecodeBalanceOf(garbage)
	require.Error(t, err)
	_, err = DecodeOwnerOf(garbage)
	require.Error(t, err)
	_, err = DecodeGetApproved(garbage)
	require.Error(t, err)
	_, err = DecodeIsApprovedForAll(garbage)
	require.Error(t, err)
	_, err = DecodeTokenURI(garbage)
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

func TestEncodeOwnerOf(t *testing.T) {
	data, err := EncodeOwnerOf(big.NewInt(42))
	require.NoError(t, err)
	require.Equal(t, []byte{0x63, 0x52, 0x21, 0x1e}, data[:4])

	vals, err := ownerOfFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(42).Cmp(vals[0].(*big.Int)))
}

func TestEncodeOwnerOfRejectsNilTokenID(t *testing.T) {
	_, err := EncodeOwnerOf(nil)
	require.Error(t, err)
}

func TestDecodeOwnerOf(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testAddr1)
	require.NoError(t, err)
	addr, err := DecodeOwnerOf(data)
	require.NoError(t, err)
	require.True(t, addr.Equal(testAddr1))
}

func TestEncodeGetApproved(t *testing.T) {
	data, err := EncodeGetApproved(big.NewInt(42))
	require.NoError(t, err)
	require.Equal(t, []byte{0x08, 0x18, 0x12, 0xfc}, data[:4])

	vals, err := getApprovedFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(42).Cmp(vals[0].(*big.Int)))
}

func TestEncodeGetApprovedRejectsNilTokenID(t *testing.T) {
	_, err := EncodeGetApproved(nil)
	require.Error(t, err)
}

func TestDecodeGetApproved(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testAddr1)
	require.NoError(t, err)
	addr, err := DecodeGetApproved(data)
	require.NoError(t, err)
	require.True(t, addr.Equal(testAddr1))
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

func TestEncodeIsApprovedForAllRejectsNilAddress(t *testing.T) {
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

func TestEncodeApprove(t *testing.T) {
	data, err := EncodeApprove(testAddr1, big.NewInt(42))
	require.NoError(t, err)
	require.Equal(t, []byte{0x09, 0x5e, 0xa7, 0xb3}, data[:4])

	vals, err := approveFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testAddr1))
	require.Zero(t, big.NewInt(42).Cmp(vals[1].(*big.Int)))
}

func TestEncodeApproveRejectsNilArgs(t *testing.T) {
	_, err := EncodeApprove(nil, big.NewInt(42))
	require.Error(t, err)
	_, err = EncodeApprove(testAddr1, nil)
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

func TestEncodeTransferFrom(t *testing.T) {
	data, err := EncodeTransferFrom(testAddr1, testAddr2, big.NewInt(42))
	require.NoError(t, err)
	require.Equal(t, []byte{0x23, 0xb8, 0x72, 0xdd}, data[:4])

	vals, err := transferFromFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testAddr1))
	require.True(t, vals[1].(*types.Address).Equal(testAddr2))
	require.Zero(t, big.NewInt(42).Cmp(vals[2].(*big.Int)))
}

func TestEncodeTransferFromRejectsNilArgs(t *testing.T) {
	_, err := EncodeTransferFrom(nil, testAddr2, big.NewInt(42))
	require.Error(t, err)
	_, err = EncodeTransferFrom(testAddr1, nil, big.NewInt(42))
	require.Error(t, err)
	_, err = EncodeTransferFrom(testAddr1, testAddr2, nil)
	require.Error(t, err)
}

func TestEncodeSafeTransferFrom(t *testing.T) {
	data, err := EncodeSafeTransferFrom(testAddr1, testAddr2, big.NewInt(42))
	require.NoError(t, err)
	require.Equal(t, []byte{0x42, 0x84, 0x2e, 0x0e}, data[:4])

	vals, err := safeTransferFromFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testAddr1))
	require.True(t, vals[1].(*types.Address).Equal(testAddr2))
	require.Zero(t, big.NewInt(42).Cmp(vals[2].(*big.Int)))
}

func TestEncodeSafeTransferFromRejectsNilArgs(t *testing.T) {
	_, err := EncodeSafeTransferFrom(nil, testAddr2, big.NewInt(42))
	require.Error(t, err)
	_, err = EncodeSafeTransferFrom(testAddr1, nil, big.NewInt(42))
	require.Error(t, err)
	_, err = EncodeSafeTransferFrom(testAddr1, testAddr2, nil)
	require.Error(t, err)
}

func TestEncodeSafeTransferFromWithData(t *testing.T) {
	callData := []byte{0xde, 0xad, 0xbe, 0xef}
	data, err := EncodeSafeTransferFromWithData(testAddr1, testAddr2, big.NewInt(42), callData)
	require.NoError(t, err)
	require.Equal(t, []byte{0xb8, 0x8d, 0x4f, 0xde}, data[:4])

	vals, err := safeTransferFromWithDataFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testAddr1))
	require.True(t, vals[1].(*types.Address).Equal(testAddr2))
	require.Zero(t, big.NewInt(42).Cmp(vals[2].(*big.Int)))
	require.Equal(t, callData, vals[3].([]byte))
}

func TestEncodeSafeTransferFromWithDataRejectsNilArgs(t *testing.T) {
	_, err := EncodeSafeTransferFromWithData(nil, testAddr2, big.NewInt(42), nil)
	require.Error(t, err)
	_, err = EncodeSafeTransferFromWithData(testAddr1, nil, big.NewInt(42), nil)
	require.Error(t, err)
	_, err = EncodeSafeTransferFromWithData(testAddr1, testAddr2, nil, nil)
	require.Error(t, err)
}

func TestEncodeTokenURI(t *testing.T) {
	data, err := EncodeTokenURI(big.NewInt(42))
	require.NoError(t, err)
	require.Equal(t, []byte{0xc8, 0x7b, 0x56, 0xdd}, data[:4])

	vals, err := tokenURIFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(42).Cmp(vals[0].(*big.Int)))
}

func TestEncodeTokenURIRejectsNilTokenID(t *testing.T) {
	_, err := EncodeTokenURI(nil)
	require.Error(t, err)
}

func TestDecodeTokenURI(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.String}, "ipfs://Qm.../42")
	require.NoError(t, err)
	s, err := DecodeTokenURI(data)
	require.NoError(t, err)
	require.Equal(t, "ipfs://Qm.../42", s)
}

func TestCustomErrorSignatures(t *testing.T) {
	tests := []struct {
		name string
		err  *abi.Error
		sig  string
	}{
		{"ERC721InvalidOwner", errInvalidOwner, "ERC721InvalidOwner(address)"},
		{"ERC721NonexistentToken", errNonexistentToken, "ERC721NonexistentToken(uint256)"},
		{"ERC721IncorrectOwner", errIncorrectOwner, "ERC721IncorrectOwner(address,uint256,address)"},
		{"ERC721InvalidSender", errInvalidSender, "ERC721InvalidSender(address)"},
		{"ERC721InvalidReceiver", errInvalidReceiver, "ERC721InvalidReceiver(address)"},
		{"ERC721InsufficientApproval", errInsufficientApproval, "ERC721InsufficientApproval(address,uint256)"},
		{"ERC721InvalidApprover", errInvalidApprover, "ERC721InvalidApprover(address)"},
		{"ERC721InvalidOperator", errInvalidOperator, "ERC721InvalidOperator(address)"},
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

func TestTransferEventSignature(t *testing.T) {
	require.Equal(t, "Transfer(address,address,uint256)", transferEvent.Signature())

	t0, ok := transferEvent.Topic0()
	require.True(t, ok)
	require.Equal(t, "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", t0.String())

	require.False(t, transferEvent.Anonymous)
	require.Len(t, transferEvent.Inputs, 3)
	require.True(t, transferEvent.Inputs[0].Indexed)
	require.True(t, transferEvent.Inputs[1].Indexed)
	require.True(t, transferEvent.Inputs[2].Indexed)
}

func TestApprovalEventSignature(t *testing.T) {
	require.Equal(t, "Approval(address,address,uint256)", approvalEvent.Signature())

	t0, ok := approvalEvent.Topic0()
	require.True(t, ok)
	require.Equal(t, "0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925", t0.String())

	require.False(t, approvalEvent.Anonymous)
	require.Len(t, approvalEvent.Inputs, 3)
	require.True(t, approvalEvent.Inputs[0].Indexed)
	require.True(t, approvalEvent.Inputs[1].Indexed)
	require.True(t, approvalEvent.Inputs[2].Indexed)
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

func TestDecodeMetadataAllSuccess(t *testing.T) {
	nameData, _ := abi.Pack(abi.Types{abi.String}, "Bored Ape Yacht Club")
	symbolData, _ := abi.Pack(abi.Types{abi.String}, "BAYC")

	data := encodeResultsData(t,
		[]bool{true, true},
		[][]byte{nameData, symbolData},
	)

	m, err := DecodeMetadata(data)
	require.NoError(t, err)
	require.Equal(t, "Bored Ape Yacht Club", m.Name)
	require.Equal(t, "BAYC", m.Symbol)
}

func TestDecodeMetadataPartialFailure(t *testing.T) {
	nameData, _ := abi.Pack(abi.Types{abi.String}, "Bored Ape Yacht Club")

	data := encodeResultsData(t,
		[]bool{true, false},
		[][]byte{nameData, nil},
	)

	m, err := DecodeMetadata(data)
	require.NoError(t, err)
	require.Equal(t, "Bored Ape Yacht Club", m.Name)
	require.Equal(t, "", m.Symbol)
}

func TestDecodeMetadataRejectsWrongResultCount(t *testing.T) {
	data := encodeResultsData(t, []bool{true}, [][]byte{{}})
	_, err := DecodeMetadata(data)
	require.Error(t, err)
}

func TestEncodeMetadataMatchesManualAggregate3(t *testing.T) {
	got, err := EncodeMetadata(bayc, true)
	require.NoError(t, err)

	want, err := multicall3.EncodeAggregate3([]multicall3.Call3{
		multicall3.NewCall3(bayc, true, EncodeName()),
		multicall3.NewCall3(bayc, true, EncodeSymbol()),
	})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDecodeMetadataWithTokenID(t *testing.T) {
	nameData, _ := abi.Pack(abi.Types{abi.String}, "Bored Ape Yacht Club")
	symbolData, _ := abi.Pack(abi.Types{abi.String}, "BAYC")
	uriData, _ := abi.Pack(abi.Types{abi.String}, "ipfs://1234")
	ownerData, _ := abi.Pack(abi.Types{abi.Address}, testAddr1)

	data := encodeResultsData(t,
		[]bool{true, true, true, true},
		[][]byte{nameData, symbolData, uriData, ownerData},
	)

	tokenId := big.NewInt(1234)
	mt, err := DecodeMetadataWithTokenID(tokenId, data)
	require.NoError(t, err)
	require.Equal(t, "Bored Ape Yacht Club", mt.Name)
	require.Equal(t, "BAYC", mt.Symbol)
	require.Same(t, tokenId, mt.TokenId)
	require.Equal(t, "ipfs://1234", mt.TokenURI)
	require.True(t, mt.Owner.Equal(testAddr1))
}

func TestDecodeMetadataWithTokenIDFailedOwner(t *testing.T) {
	nameData, _ := abi.Pack(abi.Types{abi.String}, "Bored Ape Yacht Club")
	symbolData, _ := abi.Pack(abi.Types{abi.String}, "BAYC")
	uriData, _ := abi.Pack(abi.Types{abi.String}, "ipfs://1234")

	data := encodeResultsData(t,
		[]bool{true, true, true, false},
		[][]byte{nameData, symbolData, uriData, nil},
	)

	mt, err := DecodeMetadataWithTokenID(big.NewInt(1234), data)
	require.NoError(t, err)
	require.Nil(t, mt.Owner)
}

func TestDecodeMetadataWithTokenIDRejectsWrongResultCount(t *testing.T) {
	data := encodeResultsData(t, []bool{true}, [][]byte{{}})
	_, err := DecodeMetadataWithTokenID(big.NewInt(1234), data)
	require.Error(t, err)
}

func TestEncodeMetadataWithTokenIDMatchesManualAggregate3(t *testing.T) {
	tokenId := big.NewInt(1234)
	got, err := EncodeMetadataWithTokenID(bayc, tokenId, true)
	require.NoError(t, err)

	uriCalldata, err := EncodeTokenURI(tokenId)
	require.NoError(t, err)
	ownerCalldata, err := EncodeOwnerOf(tokenId)
	require.NoError(t, err)

	want, err := multicall3.EncodeAggregate3([]multicall3.Call3{
		multicall3.NewCall3(bayc, true, EncodeName()),
		multicall3.NewCall3(bayc, true, EncodeSymbol()),
		multicall3.NewCall3(bayc, true, uriCalldata),
		multicall3.NewCall3(bayc, true, ownerCalldata),
	})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDecodeAllowanceWithTokenID(t *testing.T) {
	ownerData, _ := abi.Pack(abi.Types{abi.Address}, testAddr1)
	approvedData, _ := abi.Pack(abi.Types{abi.Address}, testAddr2)

	data := encodeResultsData(t,
		[]bool{true, true},
		[][]byte{ownerData, approvedData},
	)

	tokenId := big.NewInt(1234)
	awt, err := DecodeAllowanceWithTokenID(tokenId, data)
	require.NoError(t, err)
	require.Same(t, tokenId, awt.TokenId)
	require.True(t, awt.Owner.Equal(testAddr1))
	require.True(t, awt.Operator.Equal(testAddr2))
}

func TestDecodeAllowanceWithTokenIDFailedOperator(t *testing.T) {
	ownerData, _ := abi.Pack(abi.Types{abi.Address}, testAddr1)

	data := encodeResultsData(t,
		[]bool{true, false},
		[][]byte{ownerData, nil},
	)

	awt, err := DecodeAllowanceWithTokenID(big.NewInt(1234), data)
	require.NoError(t, err)
	require.True(t, awt.Owner.Equal(testAddr1))
	require.Nil(t, awt.Operator)
}

func TestDecodeAllowanceWithTokenIDRejectsWrongResultCount(t *testing.T) {
	data := encodeResultsData(t, []bool{true}, [][]byte{{}})
	_, err := DecodeAllowanceWithTokenID(big.NewInt(1234), data)
	require.Error(t, err)
}

func TestEncodeAllowanceWithTokenIDMatchesManualAggregate3(t *testing.T) {
	tokenId := big.NewInt(1234)
	got, err := EncodeAllowanceWithTokenID(bayc, tokenId, true)
	require.NoError(t, err)

	ownerCalldata, err := EncodeOwnerOf(tokenId)
	require.NoError(t, err)
	approvedCalldata, err := EncodeGetApproved(tokenId)
	require.NoError(t, err)

	want, err := multicall3.EncodeAggregate3([]multicall3.Call3{
		multicall3.NewCall3(bayc, true, ownerCalldata),
		multicall3.NewCall3(bayc, true, approvedCalldata),
	})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDecodeApprovalForAllWithBalance(t *testing.T) {
	approvedData, _ := abi.Pack(abi.Types{abi.Bool}, true)
	ownerBalData, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(3))
	operatorBalData, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(5))

	data := encodeResultsData(t,
		[]bool{true, true, true},
		[][]byte{approvedData, ownerBalData, operatorBalData},
	)

	afawb, err := DecodeApprovalForAllWithBalance(testAddr1, testAddr2, data)
	require.NoError(t, err)
	require.True(t, afawb.Owner.Equal(testAddr1))
	require.True(t, afawb.Operator.Equal(testAddr2))
	require.True(t, afawb.Approved)
	require.Zero(t, big.NewInt(3).Cmp(afawb.OwnerBalance))
	require.Zero(t, big.NewInt(5).Cmp(afawb.OperatorBalance))
}

func TestDecodeApprovalForAllWithBalanceRejectsWrongResultCount(t *testing.T) {
	data := encodeResultsData(t, []bool{true, true}, [][]byte{{}, {}})
	_, err := DecodeApprovalForAllWithBalance(testAddr1, testAddr2, data)
	require.Error(t, err)
}

func TestEncodeApprovalForAllWithBalanceMatchesManualAggregate3(t *testing.T) {
	got, err := EncodeApprovalForAllWithBalance(bayc, testAddr1, testAddr2, true)
	require.NoError(t, err)

	approvedCalldata, err := EncodeIsApprovedForAll(testAddr1, testAddr2)
	require.NoError(t, err)
	ownerBalCalldata, err := EncodeBalanceOf(testAddr1)
	require.NoError(t, err)
	operatorBalCalldata, err := EncodeBalanceOf(testAddr2)
	require.NoError(t, err)

	want, err := multicall3.EncodeAggregate3([]multicall3.Call3{
		multicall3.NewCall3(bayc, true, approvedCalldata),
		multicall3.NewCall3(bayc, true, ownerBalCalldata),
		multicall3.NewCall3(bayc, true, operatorBalCalldata),
	})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDecodeTokenOwners(t *testing.T) {
	tokenIds := []*big.Int{big.NewInt(1), big.NewInt(2)}
	owner1, _ := abi.Pack(abi.Types{abi.Address}, testAddr1)
	owner2, _ := abi.Pack(abi.Types{abi.Address}, testAddr2)

	data := encodeResultsData(t, []bool{true, false}, [][]byte{owner1, owner2})

	tos, err := DecodeTokenOwners(bayc, tokenIds, data)
	require.NoError(t, err)
	require.True(t, tos.Contract.Equal(bayc))
	require.Len(t, tos.Tokens, 2)
	require.Same(t, tokenIds[0], tos.Tokens[0].TokenId)
	require.True(t, tos.Tokens[0].Owner.Equal(testAddr1))
	require.Same(t, tokenIds[1], tos.Tokens[1].TokenId)
	require.Nil(t, tos.Tokens[1].Owner)
}

func TestDecodeTokenOwnersRejectsLengthMismatch(t *testing.T) {
	data := encodeResultsData(t, []bool{true}, [][]byte{{}})
	_, err := DecodeTokenOwners(bayc, []*big.Int{big.NewInt(1), big.NewInt(2)}, data)
	require.Error(t, err)
}

func TestEncodeTokenOwnersMatchesManualAggregate3(t *testing.T) {
	tokenIds := []*big.Int{big.NewInt(1), big.NewInt(2)}
	got, err := EncodeTokenOwners(bayc, tokenIds, true)
	require.NoError(t, err)

	calldata1, err := EncodeOwnerOf(tokenIds[0])
	require.NoError(t, err)
	calldata2, err := EncodeOwnerOf(tokenIds[1])
	require.NoError(t, err)

	want, err := multicall3.EncodeAggregate3([]multicall3.Call3{
		multicall3.NewCall3(bayc, true, calldata1),
		multicall3.NewCall3(bayc, true, calldata2),
	})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDecodeTokenOperators(t *testing.T) {
	tokenIds := []*big.Int{big.NewInt(1), big.NewInt(2)}
	op1, _ := abi.Pack(abi.Types{abi.Address}, testAddr1)
	op2, _ := abi.Pack(abi.Types{abi.Address}, testAddr2)

	data := encodeResultsData(t, []bool{true, false}, [][]byte{op1, op2})

	tops, err := DecodeTokenOperators(bayc, tokenIds, data)
	require.NoError(t, err)
	require.True(t, tops.Contract.Equal(bayc))
	require.Len(t, tops.Tokens, 2)
	require.Same(t, tokenIds[0], tops.Tokens[0].TokenId)
	require.True(t, tops.Tokens[0].Operator.Equal(testAddr1))
	require.Same(t, tokenIds[1], tops.Tokens[1].TokenId)
	require.Nil(t, tops.Tokens[1].Operator)
}

func TestDecodeTokenOperatorsRejectsLengthMismatch(t *testing.T) {
	data := encodeResultsData(t, []bool{true}, [][]byte{{}})
	_, err := DecodeTokenOperators(bayc, []*big.Int{big.NewInt(1), big.NewInt(2)}, data)
	require.Error(t, err)
}

func TestEncodeTokenOperatorsMatchesManualAggregate3(t *testing.T) {
	tokenIds := []*big.Int{big.NewInt(1), big.NewInt(2)}
	got, err := EncodeTokenOperators(bayc, tokenIds, true)
	require.NoError(t, err)

	calldata1, err := EncodeGetApproved(tokenIds[0])
	require.NoError(t, err)
	calldata2, err := EncodeGetApproved(tokenIds[1])
	require.NoError(t, err)

	want, err := multicall3.EncodeAggregate3([]multicall3.Call3{
		multicall3.NewCall3(bayc, true, calldata1),
		multicall3.NewCall3(bayc, true, calldata2),
	})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDecodeAddressBalances(t *testing.T) {
	contracts := []*types.Address{bayc, cryptopunks}
	bal1, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(3))
	bal2, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(4))

	data := encodeResultsData(t, []bool{true, true}, [][]byte{bal1, bal2})

	ab, err := DecodeAddressBalances(testAddr1, contracts, data)
	require.NoError(t, err)
	require.True(t, ab.Address.Equal(testAddr1))
	require.Len(t, ab.Tokens, 2)
	require.True(t, ab.Tokens[0].Token.Equal(bayc))
	require.Zero(t, big.NewInt(3).Cmp(ab.Tokens[0].Amount))
	require.True(t, ab.Tokens[1].Token.Equal(cryptopunks))
	require.Zero(t, big.NewInt(4).Cmp(ab.Tokens[1].Amount))
}

func TestDecodeAddressBalancesRejectsLengthMismatch(t *testing.T) {
	data := encodeResultsData(t, []bool{true}, [][]byte{{}})
	_, err := DecodeAddressBalances(testAddr1, []*types.Address{bayc, cryptopunks}, data)
	require.Error(t, err)
}

func TestEncodeAddressBalancesMatchesManualAggregate3(t *testing.T) {
	contracts := []*types.Address{bayc, cryptopunks}
	got, err := EncodeAddressBalances(testAddr1, contracts, true)
	require.NoError(t, err)

	calldata, err := EncodeBalanceOf(testAddr1)
	require.NoError(t, err)

	want, err := multicall3.EncodeAggregate3([]multicall3.Call3{
		multicall3.NewCall3(bayc, true, calldata),
		multicall3.NewCall3(cryptopunks, true, calldata),
	})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDecodeTokenBalances(t *testing.T) {
	addrs := []*types.Address{testAddr1, testAddr2}
	bal1, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(1))
	bal2, _ := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(2))

	data := encodeResultsData(t, []bool{true, false}, [][]byte{bal1, bal2})

	tb, err := DecodeTokenBalances(bayc, addrs, data)
	require.NoError(t, err)
	require.True(t, tb.Contract.Equal(bayc))
	require.Len(t, tb.Addresses, 2)
	require.True(t, tb.Addresses[0].Address.Equal(testAddr1))
	require.Zero(t, big.NewInt(1).Cmp(tb.Addresses[0].Amount))
	require.True(t, tb.Addresses[1].Address.Equal(testAddr2))
	require.Nil(t, tb.Addresses[1].Amount)
}

func TestDecodeTokenBalancesRejectsLengthMismatch(t *testing.T) {
	data := encodeResultsData(t, []bool{true}, [][]byte{{}})
	_, err := DecodeTokenBalances(bayc, []*types.Address{testAddr1, testAddr2}, data)
	require.Error(t, err)
}

func TestEncodeTokenBalancesMatchesManualAggregate3(t *testing.T) {
	addrs := []*types.Address{testAddr1, testAddr2}
	got, err := EncodeTokenBalances(bayc, addrs, true)
	require.NoError(t, err)

	calldata1, err := EncodeBalanceOf(testAddr1)
	require.NoError(t, err)
	calldata2, err := EncodeBalanceOf(testAddr2)
	require.NoError(t, err)

	want, err := multicall3.EncodeAggregate3([]multicall3.Call3{
		multicall3.NewCall3(bayc, true, calldata1),
		multicall3.NewCall3(bayc, true, calldata2),
	})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestEncodeOwnerPairsRejectsLengthMismatch(t *testing.T) {
	_, err := EncodeOwnerPairs([]*types.Address{bayc}, []*big.Int{big.NewInt(1), big.NewInt(2)}, false)
	require.Error(t, err)
}

func TestDecodeOwnerPairsRejectsLengthMismatch(t *testing.T) {
	_, err := DecodeOwnerPairs([]*types.Address{bayc}, []*big.Int{big.NewInt(1), big.NewInt(2)}, []byte{})
	require.Error(t, err)
}

func TestDecodeOwnerPairs(t *testing.T) {
	contracts := []*types.Address{bayc, bayc, cryptopunks}
	tokenIds := []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(2)}
	owner1, _ := abi.Pack(abi.Types{abi.Address}, testAddr1)
	owner2, _ := abi.Pack(abi.Types{abi.Address}, testAddr2)
	owner3, _ := abi.Pack(abi.Types{abi.Address}, testAddr2)

	data := encodeResultsData(t, []bool{true, true, true}, [][]byte{owner1, owner2, owner3})

	pairs, err := DecodeOwnerPairs(contracts, tokenIds, data)
	require.NoError(t, err)
	require.Len(t, pairs, 3)
	require.True(t, pairs[0].Contract.Equal(bayc))
	require.Same(t, tokenIds[0], pairs[0].TokenId)
	require.True(t, pairs[0].Owner.Equal(testAddr1))
	require.True(t, pairs[2].Contract.Equal(cryptopunks))
	require.Same(t, tokenIds[2], pairs[2].TokenId)
	require.True(t, pairs[2].Owner.Equal(testAddr2))
}

func TestEncodeOwnerPairsMatchesManualAggregate3(t *testing.T) {
	contracts := []*types.Address{bayc, cryptopunks}
	tokenIds := []*big.Int{big.NewInt(1), big.NewInt(2)}
	got, err := EncodeOwnerPairs(contracts, tokenIds, true)
	require.NoError(t, err)

	calldata1, err := EncodeOwnerOf(tokenIds[0])
	require.NoError(t, err)
	calldata2, err := EncodeOwnerOf(tokenIds[1])
	require.NoError(t, err)

	want, err := multicall3.EncodeAggregate3([]multicall3.Call3{
		multicall3.NewCall3(bayc, true, calldata1),
		multicall3.NewCall3(cryptopunks, true, calldata2),
	})
	require.NoError(t, err)
	require.Equal(t, want, got)
}
