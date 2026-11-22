package erc721

import (
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

var bayc = types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")

func TestExtractTransfersSampleReceipt(t *testing.T) {
	transferTopic, ok := transferEvent.Topic0()
	require.True(t, ok)
	approvalTopic, ok := approvalEvent.Topic0()
	require.True(t, ok)
	approvalForAllTopic, ok := approvalForAllEvent.Topic0()
	require.True(t, ok)

	approvedData, err := abi.Pack(abi.NewTypes(abi.Bool), true)
	require.NoError(t, err)

	logs := []tx.Log{
		{
			Address: bayc,
			Topics: []*types.Hash{
				types.NewHashFromBytes(transferTopic[:]),
				types.NewHashFromBytes(testAddr1.Bytes()),
				types.NewHashFromBytes(testAddr2.Bytes()),
				types.NewHashFromBytes(big.NewInt(1234).FillBytes(make([]byte, 32))),
			},
		},
		{
			Address: bayc,
			Topics: []*types.Hash{
				types.NewHashFromBytes(approvalTopic[:]),
				types.NewHashFromBytes(testAddr2.Bytes()),
				types.NewHashFromBytes(testAddr1.Bytes()),
				types.NewHashFromBytes(big.NewInt(1234).FillBytes(make([]byte, 32))),
			},
		},
		{
			Address: bayc,
			Topics: []*types.Hash{
				types.NewHashFromBytes(approvalForAllTopic[:]),
				types.NewHashFromBytes(testAddr1.Bytes()),
				types.NewHashFromBytes(testAddr2.Bytes()),
			},
			Data: approvedData,
		},
	}

	transfers, err := ExtractTransfers(logs)
	require.NoError(t, err)
	require.Len(t, transfers, 1)
	require.True(t, transfers[0].Contract.Equal(bayc))
	require.True(t, transfers[0].From.Equal(testAddr1))
	require.True(t, transfers[0].To.Equal(testAddr2))
	require.Zero(t, big.NewInt(1234).Cmp(transfers[0].TokenId))
}

func TestExtractApprovalsSampleReceipt(t *testing.T) {
	transferTopic, ok := transferEvent.Topic0()
	require.True(t, ok)
	approvalTopic, ok := approvalEvent.Topic0()
	require.True(t, ok)
	approvalForAllTopic, ok := approvalForAllEvent.Topic0()
	require.True(t, ok)

	approvedData, err := abi.Pack(abi.NewTypes(abi.Bool), true)
	require.NoError(t, err)

	logs := []tx.Log{
		{
			Address: bayc,
			Topics: []*types.Hash{
				types.NewHashFromBytes(transferTopic[:]),
				types.NewHashFromBytes(testAddr1.Bytes()),
				types.NewHashFromBytes(testAddr2.Bytes()),
				types.NewHashFromBytes(big.NewInt(1234).FillBytes(make([]byte, 32))),
			},
		},
		{
			Address: bayc,
			Topics: []*types.Hash{
				types.NewHashFromBytes(approvalTopic[:]),
				types.NewHashFromBytes(testAddr2.Bytes()),
				types.NewHashFromBytes(testAddr1.Bytes()),
				types.NewHashFromBytes(big.NewInt(1234).FillBytes(make([]byte, 32))),
			},
		},
		{
			Address: bayc,
			Topics: []*types.Hash{
				types.NewHashFromBytes(approvalForAllTopic[:]),
				types.NewHashFromBytes(testAddr1.Bytes()),
				types.NewHashFromBytes(testAddr2.Bytes()),
			},
			Data: approvedData,
		},
	}

	approvals, err := ExtractApprovals(logs)
	require.NoError(t, err)
	require.Len(t, approvals, 1)
	require.True(t, approvals[0].Contract.Equal(bayc))
	require.True(t, approvals[0].Owner.Equal(testAddr2))
	require.True(t, approvals[0].Approved.Equal(testAddr1))
	require.Zero(t, big.NewInt(1234).Cmp(approvals[0].TokenId))
}

func TestExtractApprovalForAllSampleReceipt(t *testing.T) {
	transferTopic, ok := transferEvent.Topic0()
	require.True(t, ok)
	approvalTopic, ok := approvalEvent.Topic0()
	require.True(t, ok)
	approvalForAllTopic, ok := approvalForAllEvent.Topic0()
	require.True(t, ok)

	approvedData, err := abi.Pack(abi.NewTypes(abi.Bool), true)
	require.NoError(t, err)

	logs := []tx.Log{
		{
			Address: bayc,
			Topics: []*types.Hash{
				types.NewHashFromBytes(transferTopic[:]),
				types.NewHashFromBytes(testAddr1.Bytes()),
				types.NewHashFromBytes(testAddr2.Bytes()),
				types.NewHashFromBytes(big.NewInt(1234).FillBytes(make([]byte, 32))),
			},
		},
		{
			Address: bayc,
			Topics: []*types.Hash{
				types.NewHashFromBytes(approvalTopic[:]),
				types.NewHashFromBytes(testAddr2.Bytes()),
				types.NewHashFromBytes(testAddr1.Bytes()),
				types.NewHashFromBytes(big.NewInt(1234).FillBytes(make([]byte, 32))),
			},
		},
		{
			Address: bayc,
			Topics: []*types.Hash{
				types.NewHashFromBytes(approvalForAllTopic[:]),
				types.NewHashFromBytes(testAddr1.Bytes()),
				types.NewHashFromBytes(testAddr2.Bytes()),
			},
			Data: approvedData,
		},
	}

	approvals, err := ExtractApprovalForAll(logs)
	require.NoError(t, err)
	require.Len(t, approvals, 1)
	require.True(t, approvals[0].Contract.Equal(bayc))
	require.True(t, approvals[0].Owner.Equal(testAddr1))
	require.True(t, approvals[0].Operator.Equal(testAddr2))
	require.True(t, approvals[0].Approved)
}

func TestExtractTransfersEmptyLogs(t *testing.T) {
	transfers, err := ExtractTransfers(nil)
	require.NoError(t, err)
	require.Empty(t, transfers)
}

func TestExtractTransfersNoTopics(t *testing.T) {
	logs := []tx.Log{{Address: bayc, Topics: nil, Data: nil}}
	transfers, err := ExtractTransfers(logs)
	require.NoError(t, err)
	require.Empty(t, transfers)
}

func TestExtractTransfersUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{
		Address: bayc,
		Topics:  []*types.Hash{unrelated},
		Data:    []byte{0xde, 0xad},
	}}
	transfers, err := ExtractTransfers(logs)
	require.NoError(t, err)
	require.Empty(t, transfers)
}

func TestExtractTransfersRejectsNilTopic(t *testing.T) {
	logs := []tx.Log{{
		Address: bayc,
		Topics:  []*types.Hash{nil},
		Data:    nil,
	}}
	transfers, err := ExtractTransfers(logs)
	require.NoError(t, err)
	require.Empty(t, transfers)

	t0, ok := transferEvent.Topic0()
	require.True(t, ok)
	logs = []tx.Log{{
		Address: bayc,
		Topics:  []*types.Hash{types.NewHashFromBytes(t0[:]), nil, nil, nil},
		Data:    nil,
	}}
	_, err = ExtractTransfers(logs)
	require.Error(t, err)
}

func TestExtractTransfersRejectsMalformedTopicCount(t *testing.T) {
	t0, ok := transferEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: bayc,
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

func TestExtractApprovalForAllRejectsMalformedData(t *testing.T) {
	t0, ok := approvalForAllEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: bayc,
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
		Address: bayc,
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
