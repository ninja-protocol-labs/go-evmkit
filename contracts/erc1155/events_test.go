package erc1155

import (
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

func TestExtractTransfersSampleReceipt(t *testing.T) {
	transfers, err := ExtractTransfers(sampleReceiptLogs(t))
	require.NoError(t, err)
	require.Len(t, transfers, 1)
	require.True(t, transfers[0].Contract.Equal(testTokens))
	require.True(t, transfers[0].Operator.Equal(testAddr1))
	require.True(t, transfers[0].From.Equal(testAddr1))
	require.True(t, transfers[0].To.Equal(testAddr2))
	require.Zero(t, big.NewInt(42).Cmp(transfers[0].Id))
	require.Zero(t, big.NewInt(7).Cmp(transfers[0].Value))
}

func TestExtractTransferBatchesSampleReceipt(t *testing.T) {
	batches, err := ExtractTransferBatches(sampleReceiptLogs(t))
	require.NoError(t, err)
	require.Len(t, batches, 1)
	require.True(t, batches[0].Contract.Equal(testTokens))
	require.True(t, batches[0].Operator.Equal(testAddr1))
	require.True(t, batches[0].From.Equal(testAddr1))
	require.True(t, batches[0].To.Equal(testAddr2))
	require.Len(t, batches[0].Ids, 2)
	require.Zero(t, big.NewInt(1).Cmp(batches[0].Ids[0]))
	require.Zero(t, big.NewInt(2).Cmp(batches[0].Ids[1]))
	require.Len(t, batches[0].Values, 2)
	require.Zero(t, big.NewInt(10).Cmp(batches[0].Values[0]))
	require.Zero(t, big.NewInt(20).Cmp(batches[0].Values[1]))
}

func TestExtractApprovalForAllSampleReceipt(t *testing.T) {
	approvals, err := ExtractApprovalForAll(sampleReceiptLogs(t))
	require.NoError(t, err)
	require.Len(t, approvals, 1)
	require.True(t, approvals[0].Contract.Equal(testTokens))
	require.True(t, approvals[0].Account.Equal(testAddr1))
	require.True(t, approvals[0].Operator.Equal(testAddr2))
	require.True(t, approvals[0].Approved)
}

func TestExtractURISampleReceipt(t *testing.T) {
	uris, err := ExtractURI(sampleReceiptLogs(t))
	require.NoError(t, err)
	require.Len(t, uris, 1)
	require.True(t, uris[0].Contract.Equal(testTokens))
	require.Equal(t, "ipfs://1234", uris[0].Value)
	require.Zero(t, big.NewInt(42).Cmp(uris[0].Id))
}

func TestExtractTransfersEmptyLogs(t *testing.T) {
	transfers, err := ExtractTransfers(nil)
	require.NoError(t, err)
	require.Empty(t, transfers)
}

func TestExtractTransfersNoTopics(t *testing.T) {
	logs := []tx.Log{{Address: testTokens, Topics: nil, Data: nil}}
	transfers, err := ExtractTransfers(logs)
	require.NoError(t, err)
	require.Empty(t, transfers)
}

func TestExtractTransfersUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{
		Address: testTokens,
		Topics:  []*types.Hash{unrelated},
		Data:    []byte{0xde, 0xad},
	}}
	transfers, err := ExtractTransfers(logs)
	require.NoError(t, err)
	require.Empty(t, transfers)
}

func TestExtractTransfersRejectsNilTopic(t *testing.T) {
	logs := []tx.Log{{
		Address: testTokens,
		Topics:  []*types.Hash{nil},
		Data:    nil,
	}}
	transfers, err := ExtractTransfers(logs)
	require.NoError(t, err)
	require.Empty(t, transfers)

	t0, ok := transferSingleEvent.Topic0()
	require.True(t, ok)
	logs = []tx.Log{{
		Address: testTokens,
		Topics:  []*types.Hash{types.NewHashFromBytes(t0[:]), nil, nil, nil},
		Data:    nil,
	}}
	_, err = ExtractTransfers(logs)
	require.Error(t, err)
}

func TestExtractTransfersRejectsMalformedTopicCount(t *testing.T) {
	t0, ok := transferSingleEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testTokens,
		Topics: []*types.Hash{
			types.NewHashFromBytes(t0[:]),
			types.NewHashFromBytes(testAddr1.Bytes()),
			types.NewHashFromBytes(testAddr2.Bytes()),
		},
		Data: nil,
	}}
	_, err := ExtractTransfers(logs)
	require.Error(t, err)
}

func TestExtractTransferBatchesRejectsMalformedData(t *testing.T) {
	t0, ok := transferBatchEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testTokens,
		Topics: []*types.Hash{
			types.NewHashFromBytes(t0[:]),
			types.NewHashFromBytes(testAddr1.Bytes()),
			types.NewHashFromBytes(testAddr1.Bytes()),
			types.NewHashFromBytes(testAddr2.Bytes()),
		},
		Data: []byte{0x01, 0x02},
	}}
	_, err := ExtractTransferBatches(logs)
	require.Error(t, err)
}

func TestExtractApprovalForAllRejectsMalformedData(t *testing.T) {
	t0, ok := approvalForAllEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testTokens,
		Topics: []*types.Hash{
			types.NewHashFromBytes(t0[:]),
			types.NewHashFromBytes(testAddr1.Bytes()),
			types.NewHashFromBytes(testAddr2.Bytes()),
		},
		Data: []byte{0x01, 0x02},
	}}
	_, err := ExtractApprovalForAll(logs)
	require.Error(t, err)
}

func TestExtractApprovalForAllRejectsHandCraftedShortData(t *testing.T) {
	t0, ok := approvalForAllEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testTokens,
		Topics: []*types.Hash{
			types.NewHashFromBytes(t0[:]),
			types.NewHashFromBytes(testAddr1.Bytes()),
			types.NewHashFromBytes(testAddr2.Bytes()),
		},
		Data: []byte{0x01},
	}}

	require.NotPanics(t, func() {
		_, err := ExtractApprovalForAll(logs)
		require.Error(t, err)
	})
}

func TestExtractURIRejectsMalformedData(t *testing.T) {
	t0, ok := uriEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testTokens,
		Topics: []*types.Hash{
			types.NewHashFromBytes(t0[:]),
			types.NewHashFromBytes(big.NewInt(1).FillBytes(make([]byte, 32))),
		},
		Data: []byte{0x01, 0x02},
	}}
	_, err := ExtractURI(logs)
	require.Error(t, err)
}

func sampleReceiptLogs(t *testing.T) []tx.Log {
	t.Helper()

	transferSingleTopic, ok := transferSingleEvent.Topic0()
	require.True(t, ok)
	transferBatchTopic, ok := transferBatchEvent.Topic0()
	require.True(t, ok)
	approvalForAllTopic, ok := approvalForAllEvent.Topic0()
	require.True(t, ok)
	uriTopic, ok := uriEvent.Topic0()
	require.True(t, ok)

	transferSingleData, err := abi.Pack(abi.Types{abi.Uint256, abi.Uint256}, big.NewInt(42), big.NewInt(7))
	require.NoError(t, err)
	transferBatchData, err := abi.Pack(abi.Types{abi.Slice(abi.Uint256), abi.Slice(abi.Uint256)},
		[]*big.Int{big.NewInt(1), big.NewInt(2)}, []*big.Int{big.NewInt(10), big.NewInt(20)})
	require.NoError(t, err)
	approvedData, err := abi.Pack(abi.Types{abi.Bool}, true)
	require.NoError(t, err)
	uriData, err := abi.Pack(abi.Types{abi.String}, "ipfs://1234")
	require.NoError(t, err)

	return []tx.Log{
		{
			Address: testTokens,
			Topics: []*types.Hash{
				types.NewHashFromBytes(transferSingleTopic[:]),
				types.NewHashFromBytes(testAddr1.Bytes()),
				types.NewHashFromBytes(testAddr1.Bytes()),
				types.NewHashFromBytes(testAddr2.Bytes()),
			},
			Data: transferSingleData,
		},
		{
			Address: testTokens,
			Topics: []*types.Hash{
				types.NewHashFromBytes(transferBatchTopic[:]),
				types.NewHashFromBytes(testAddr1.Bytes()),
				types.NewHashFromBytes(testAddr1.Bytes()),
				types.NewHashFromBytes(testAddr2.Bytes()),
			},
			Data: transferBatchData,
		},
		{
			Address: testTokens,
			Topics: []*types.Hash{
				types.NewHashFromBytes(approvalForAllTopic[:]),
				types.NewHashFromBytes(testAddr1.Bytes()),
				types.NewHashFromBytes(testAddr2.Bytes()),
			},
			Data: approvedData,
		},
		{
			Address: testTokens,
			Topics: []*types.Hash{
				types.NewHashFromBytes(uriTopic[:]),
				types.NewHashFromBytes(big.NewInt(42).FillBytes(make([]byte, 32))),
			},
			Data: uriData,
		},
	}
}
