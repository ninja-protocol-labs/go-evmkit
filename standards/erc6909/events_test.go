package erc6909

import (
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

var (
	testAddr1  = types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	testAddr2  = types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	testTokens = types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
)

func TestExtractOperatorSetsSampleReceipt(t *testing.T) {
	operatorSetTopic, ok := operatorSetEvent.Topic0()
	require.True(t, ok)
	operatorSetData, err := abi.Pack(abi.Types{abi.Bool}, true)
	require.NoError(t, err)

	logs := []tx.Log{
		{
			Address: testTokens,
			Topics: []*types.Hash{
				types.NewHashFromBytes(operatorSetTopic[:]),
				types.NewHashFromBytes(testAddr1.Bytes()),
				types.NewHashFromBytes(testAddr2.Bytes()),
			},
			Data: operatorSetData,
		},
	}

	operatorSets, err := ExtractOperatorSets(logs)
	require.NoError(t, err)
	require.Len(t, operatorSets, 1)
	require.True(t, operatorSets[0].Contract.Equal(testTokens))
	require.True(t, operatorSets[0].Owner.Equal(testAddr1))
	require.True(t, operatorSets[0].Operator.Equal(testAddr2))
	require.True(t, operatorSets[0].Approved)
}

func TestExtractApprovalsSampleReceipt(t *testing.T) {
	approvalTopic, ok := approvalEvent.Topic0()
	require.True(t, ok)
	approvalData, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(100))
	require.NoError(t, err)

	logs := []tx.Log{
		{
			Address: testTokens,
			Topics: []*types.Hash{
				types.NewHashFromBytes(approvalTopic[:]),
				types.NewHashFromBytes(testAddr1.Bytes()),
				types.NewHashFromBytes(testAddr2.Bytes()),
				types.NewHashFromBytes(big.NewInt(1).FillBytes(make([]byte, 32))),
			},
			Data: approvalData,
		},
	}

	approvals, err := ExtractApprovals(logs)
	require.NoError(t, err)
	require.Len(t, approvals, 1)
	require.True(t, approvals[0].Contract.Equal(testTokens))
	require.True(t, approvals[0].Owner.Equal(testAddr1))
	require.True(t, approvals[0].Spender.Equal(testAddr2))
	require.Zero(t, big.NewInt(1).Cmp(approvals[0].Id))
	require.Zero(t, big.NewInt(100).Cmp(approvals[0].Amount))
}

func TestExtractTransfersSampleReceipt(t *testing.T) {
	transferTopic, ok := transferEvent.Topic0()
	require.True(t, ok)
	transferData, err := abi.Pack(abi.Types{abi.Address, abi.Uint256}, testAddr1, big.NewInt(7))
	require.NoError(t, err)

	logs := []tx.Log{
		{
			Address: testTokens,
			Topics: []*types.Hash{
				types.NewHashFromBytes(transferTopic[:]),
				types.NewHashFromBytes(testAddr1.Bytes()),
				types.NewHashFromBytes(testAddr2.Bytes()),
				types.NewHashFromBytes(big.NewInt(42).FillBytes(make([]byte, 32))),
			},
			Data: transferData,
		},
	}

	transfers, err := ExtractTransfers(logs)
	require.NoError(t, err)
	require.Len(t, transfers, 1)
	require.True(t, transfers[0].Contract.Equal(testTokens))
	require.True(t, transfers[0].Caller.Equal(testAddr1))
	require.True(t, transfers[0].Sender.Equal(testAddr1))
	require.True(t, transfers[0].Receiver.Equal(testAddr2))
	require.Zero(t, big.NewInt(42).Cmp(transfers[0].Id))
	require.Zero(t, big.NewInt(7).Cmp(transfers[0].Amount))
}

func TestExtractNameUpdatedSampleReceipt(t *testing.T) {
	nameUpdatedTopic, ok := nameUpdatedEvent.Topic0()
	require.True(t, ok)
	nameUpdatedData, err := abi.Pack(abi.Types{abi.String}, "evmkit test token")
	require.NoError(t, err)

	logs := []tx.Log{
		{
			Address: testTokens,
			Topics: []*types.Hash{
				types.NewHashFromBytes(nameUpdatedTopic[:]),
				types.NewHashFromBytes(big.NewInt(1).FillBytes(make([]byte, 32))),
			},
			Data: nameUpdatedData,
		},
	}

	updates, err := ExtractNameUpdated(logs)
	require.NoError(t, err)
	require.Len(t, updates, 1)
	require.True(t, updates[0].Contract.Equal(testTokens))
	require.Zero(t, big.NewInt(1).Cmp(updates[0].Id))
	require.Equal(t, "evmkit test token", updates[0].NewName)
}

func TestExtractSymbolUpdatedSampleReceipt(t *testing.T) {
	symbolUpdatedTopic, ok := symbolUpdatedEvent.Topic0()
	require.True(t, ok)
	symbolUpdatedData, err := abi.Pack(abi.Types{abi.String}, "EVK6909")
	require.NoError(t, err)

	logs := []tx.Log{
		{
			Address: testTokens,
			Topics: []*types.Hash{
				types.NewHashFromBytes(symbolUpdatedTopic[:]),
				types.NewHashFromBytes(big.NewInt(1).FillBytes(make([]byte, 32))),
			},
			Data: symbolUpdatedData,
		},
	}

	updates, err := ExtractSymbolUpdated(logs)
	require.NoError(t, err)
	require.Len(t, updates, 1)
	require.True(t, updates[0].Contract.Equal(testTokens))
	require.Zero(t, big.NewInt(1).Cmp(updates[0].Id))
	require.Equal(t, "EVK6909", updates[0].NewSymbol)
}

func TestExtractDecimalsUpdatedSampleReceipt(t *testing.T) {
	decimalsUpdatedTopic, ok := decimalsUpdatedEvent.Topic0()
	require.True(t, ok)
	decimalsUpdatedData, err := abi.Pack(abi.Types{abi.Uint8}, uint8(18))
	require.NoError(t, err)

	logs := []tx.Log{
		{
			Address: testTokens,
			Topics: []*types.Hash{
				types.NewHashFromBytes(decimalsUpdatedTopic[:]),
				types.NewHashFromBytes(big.NewInt(1).FillBytes(make([]byte, 32))),
			},
			Data: decimalsUpdatedData,
		},
	}

	updates, err := ExtractDecimalsUpdated(logs)
	require.NoError(t, err)
	require.Len(t, updates, 1)
	require.True(t, updates[0].Contract.Equal(testTokens))
	require.Zero(t, big.NewInt(1).Cmp(updates[0].Id))
	require.Equal(t, uint8(18), updates[0].NewDecimals)
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

	t0, ok := transferEvent.Topic0()
	require.True(t, ok)
	logs = []tx.Log{{
		Address: testTokens,
		Topics:  []*types.Hash{types.NewHashFromBytes(t0[:]), nil, nil, nil, nil},
		Data:    nil,
	}}
	_, err = ExtractTransfers(logs)
	require.Error(t, err)
}

func TestExtractTransfersRejectsMalformedTopicCount(t *testing.T) {
	t0, ok := transferEvent.Topic0()
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

func TestExtractOperatorSetsRejectsMalformedData(t *testing.T) {
	t0, ok := operatorSetEvent.Topic0()
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
	_, err := ExtractOperatorSets(logs)
	require.Error(t, err)
}

func TestExtractOperatorSetsRejectsHandCraftedShortData(t *testing.T) {
	t0, ok := operatorSetEvent.Topic0()
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
		_, err := ExtractOperatorSets(logs)
		require.Error(t, err)
	})
}

func TestExtractApprovalsRejectsMalformedData(t *testing.T) {
	t0, ok := approvalEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testTokens,
		Topics: []*types.Hash{
			types.NewHashFromBytes(t0[:]),
			types.NewHashFromBytes(testAddr1.Bytes()),
			types.NewHashFromBytes(testAddr2.Bytes()),
			types.NewHashFromBytes(big.NewInt(1).FillBytes(make([]byte, 32))),
		},
		Data: []byte{0x01, 0x02},
	}}
	_, err := ExtractApprovals(logs)
	require.Error(t, err)
}

func TestExtractNameUpdatedRejectsMalformedData(t *testing.T) {
	t0, ok := nameUpdatedEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testTokens,
		Topics: []*types.Hash{
			types.NewHashFromBytes(t0[:]),
			types.NewHashFromBytes(big.NewInt(1).FillBytes(make([]byte, 32))),
		},
		Data: []byte{0x01, 0x02},
	}}
	_, err := ExtractNameUpdated(logs)
	require.Error(t, err)
}

func TestExtractSymbolUpdatedRejectsMalformedData(t *testing.T) {
	t0, ok := symbolUpdatedEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testTokens,
		Topics: []*types.Hash{
			types.NewHashFromBytes(t0[:]),
			types.NewHashFromBytes(big.NewInt(1).FillBytes(make([]byte, 32))),
		},
		Data: []byte{0x01, 0x02},
	}}
	_, err := ExtractSymbolUpdated(logs)
	require.Error(t, err)
}

func TestExtractDecimalsUpdatedRejectsMalformedData(t *testing.T) {
	t0, ok := decimalsUpdatedEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testTokens,
		Topics: []*types.Hash{
			types.NewHashFromBytes(t0[:]),
			types.NewHashFromBytes(big.NewInt(1).FillBytes(make([]byte, 32))),
		},
		Data: []byte{},
	}}
	_, err := ExtractDecimalsUpdated(logs)
	require.Error(t, err)
}
