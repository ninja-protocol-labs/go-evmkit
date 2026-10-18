package v2

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

func TestExtractPairCreatedSampleLog(t *testing.T) {
	topic0, ok := pairCreatedEvent.Topic0()
	require.True(t, ok)

	data, err := abi.Pack(abi.Types{abi.Address, abi.Uint256}, testPair, big.NewInt(7))
	require.NoError(t, err)

	log := tx.Log{
		Address: testFactory,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testTokenA.Bytes()),
			types.NewHashFromBytes(testTokenB.Bytes()),
		},
		Data: data,
	}

	created, err := ExtractPairCreated([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, created, 1)
	require.True(t, created[0].Contract.Equal(testFactory))
	require.True(t, created[0].Token0.Equal(testTokenA))
	require.True(t, created[0].Token1.Equal(testTokenB))
	require.True(t, created[0].Pair.Equal(testPair))
	require.Zero(t, big.NewInt(7).Cmp(created[0].AllPairsLength))
}

func TestExtractPairCreatedEmptyLogs(t *testing.T) {
	created, err := ExtractPairCreated(nil)
	require.NoError(t, err)
	require.Empty(t, created)
}

func TestExtractPairCreatedNoTopics(t *testing.T) {
	created, err := ExtractPairCreated([]tx.Log{{Address: testFactory, Topics: nil, Data: nil}})
	require.NoError(t, err)
	require.Empty(t, created)
}

func TestExtractPairCreatedUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{
		Address: testFactory,
		Topics:  []*types.Hash{unrelated},
		Data:    []byte{0xde, 0xad},
	}}
	created, err := ExtractPairCreated(logs)
	require.NoError(t, err)
	require.Empty(t, created)
}

func TestExtractPairCreatedRejectsNilTopic(t *testing.T) {
	logs := []tx.Log{{
		Address: testFactory,
		Topics:  []*types.Hash{nil},
		Data:    nil,
	}}
	created, err := ExtractPairCreated(logs)
	require.NoError(t, err)
	require.Empty(t, created)

	topic0, ok := pairCreatedEvent.Topic0()
	require.True(t, ok)
	logs = []tx.Log{{
		Address: testFactory,
		Topics:  []*types.Hash{types.NewHashFromBytes(topic0[:]), nil, nil},
		Data:    nil,
	}}
	_, err = ExtractPairCreated(logs)
	require.Error(t, err)
}

func TestExtractPairCreatedRejectsMalformedTopicCount(t *testing.T) {
	topic0, ok := pairCreatedEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testFactory,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testTokenA.Bytes()),
		},
		Data: nil,
	}}
	_, err := ExtractPairCreated(logs)
	require.Error(t, err)
}

func TestExtractPairCreatedRejectsMalformedData(t *testing.T) {
	topic0, ok := pairCreatedEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testFactory,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testTokenA.Bytes()),
			types.NewHashFromBytes(testTokenB.Bytes()),
		},
		Data: []byte{0xde, 0xad},
	}}
	_, err := ExtractPairCreated(logs)
	require.Error(t, err)
}
