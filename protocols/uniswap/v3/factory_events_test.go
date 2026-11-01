package v3

import (
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

var testFactory = types.MustNewAddressFromHex("0x5c69bee701ef814a2b6a3edd4b1652cb9cc5aa6f")

func TestExtractOwnerChangedSampleLog(t *testing.T) {
	topic0, ok := ownerChangedEvent.Topic0()
	require.True(t, ok)

	log := tx.Log{
		Address: testFactory,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testOwner.Bytes()),
			types.NewHashFromBytes(testTokenA.Bytes()),
		},
		Data: nil,
	}

	changed, err := ExtractOwnerChanged([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, changed, 1)
	require.True(t, changed[0].Contract.Equal(testFactory))
	require.True(t, changed[0].OldOwner.Equal(testOwner))
	require.True(t, changed[0].NewOwner.Equal(testTokenA))
}

func TestExtractOwnerChangedEmptyLogs(t *testing.T) {
	changed, err := ExtractOwnerChanged(nil)
	require.NoError(t, err)
	require.Empty(t, changed)
}

func TestExtractOwnerChangedUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{Address: testFactory, Topics: []*types.Hash{unrelated}, Data: []byte{0xde, 0xad}}}
	changed, err := ExtractOwnerChanged(logs)
	require.NoError(t, err)
	require.Empty(t, changed)
}

func TestExtractOwnerChangedRejectsMalformedTopicCount(t *testing.T) {
	topic0, ok := ownerChangedEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testFactory,
		Topics:  []*types.Hash{types.NewHashFromBytes(topic0[:]), types.NewHashFromBytes(testOwner.Bytes())},
		Data:    nil,
	}}
	_, err := ExtractOwnerChanged(logs)
	require.Error(t, err)
}

func TestExtractPoolCreatedSampleLog(t *testing.T) {
	topic0, ok := poolCreatedEvent.Topic0()
	require.True(t, ok)

	fee := big.NewInt(3000)
	tickSpacing := big.NewInt(60)
	data, err := abi.Pack(abi.Types{abi.Int24, abi.Address}, tickSpacing, testPool)
	require.NoError(t, err)

	log := tx.Log{
		Address: testFactory,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testTokenA.Bytes()),
			types.NewHashFromBytes(testTokenB.Bytes()),
			types.NewHashFromBytes(fee.Bytes()),
		},
		Data: data,
	}

	created, err := ExtractPoolCreated([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, created, 1)
	require.True(t, created[0].Contract.Equal(testFactory))
	require.True(t, created[0].Token0.Equal(testTokenA))
	require.True(t, created[0].Token1.Equal(testTokenB))
	require.Zero(t, fee.Cmp(created[0].Fee))
	require.Zero(t, tickSpacing.Cmp(created[0].TickSpacing))
	require.True(t, created[0].Pool.Equal(testPool))
}

func TestExtractPoolCreatedEmptyLogs(t *testing.T) {
	created, err := ExtractPoolCreated(nil)
	require.NoError(t, err)
	require.Empty(t, created)
}

func TestExtractPoolCreatedNoTopics(t *testing.T) {
	created, err := ExtractPoolCreated([]tx.Log{{Address: testFactory, Topics: nil, Data: nil}})
	require.NoError(t, err)
	require.Empty(t, created)
}

func TestExtractPoolCreatedUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{Address: testFactory, Topics: []*types.Hash{unrelated}, Data: []byte{0xde, 0xad}}}
	created, err := ExtractPoolCreated(logs)
	require.NoError(t, err)
	require.Empty(t, created)
}

func TestExtractPoolCreatedRejectsNilTopic(t *testing.T) {
	logs := []tx.Log{{Address: testFactory, Topics: []*types.Hash{nil}, Data: nil}}
	created, err := ExtractPoolCreated(logs)
	require.NoError(t, err)
	require.Empty(t, created)

	topic0, ok := poolCreatedEvent.Topic0()
	require.True(t, ok)
	logs = []tx.Log{{
		Address: testFactory,
		Topics:  []*types.Hash{types.NewHashFromBytes(topic0[:]), nil, nil, nil},
		Data:    nil,
	}}
	_, err = ExtractPoolCreated(logs)
	require.Error(t, err)
}

func TestExtractPoolCreatedRejectsMalformedTopicCount(t *testing.T) {
	topic0, ok := poolCreatedEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testFactory,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testTokenA.Bytes()),
		},
		Data: nil,
	}}
	_, err := ExtractPoolCreated(logs)
	require.Error(t, err)
}

func TestExtractPoolCreatedRejectsMalformedData(t *testing.T) {
	topic0, ok := poolCreatedEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testFactory,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testTokenA.Bytes()),
			types.NewHashFromBytes(testTokenB.Bytes()),
			types.NewHashFromBytes(big.NewInt(3000).Bytes()),
		},
		Data: []byte{0xde, 0xad},
	}}
	_, err := ExtractPoolCreated(logs)
	require.Error(t, err)
}

func TestExtractFeeAmountEnabledSampleLog(t *testing.T) {
	topic0, ok := feeAmountEnabledEvent.Topic0()
	require.True(t, ok)

	fee := big.NewInt(500)
	tickSpacing := big.NewInt(10)
	log := tx.Log{
		Address: testFactory,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(fee.Bytes()),
			types.NewHashFromBytes(tickSpacing.Bytes()),
		},
		Data: nil,
	}

	enabled, err := ExtractFeeAmountEnabled([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, enabled, 1)
	require.True(t, enabled[0].Contract.Equal(testFactory))
	require.Zero(t, fee.Cmp(enabled[0].Fee))
	require.Zero(t, tickSpacing.Cmp(enabled[0].TickSpacing))
}

func TestExtractFeeAmountEnabledEmptyLogs(t *testing.T) {
	enabled, err := ExtractFeeAmountEnabled(nil)
	require.NoError(t, err)
	require.Empty(t, enabled)
}

func TestExtractFeeAmountEnabledUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{Address: testFactory, Topics: []*types.Hash{unrelated}, Data: []byte{0xde, 0xad}}}
	enabled, err := ExtractFeeAmountEnabled(logs)
	require.NoError(t, err)
	require.Empty(t, enabled)
}

func TestExtractFeeAmountEnabledRejectsMalformedTopicCount(t *testing.T) {
	topic0, ok := feeAmountEnabledEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testFactory,
		Topics:  []*types.Hash{types.NewHashFromBytes(topic0[:]), types.NewHashFromBytes(big.NewInt(500).Bytes())},
		Data:    nil,
	}}
	_, err := ExtractFeeAmountEnabled(logs)
	require.Error(t, err)
}
