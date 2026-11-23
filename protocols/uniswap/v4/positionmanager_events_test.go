package v4

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
	testPositionManager = types.MustNewAddressFromHex("0x4b2c77d209d3405f41a037ec6c77f7f5b8e2ca80")
	testSubscriber      = types.MustNewAddressFromHex("0x9a8ca723f5dccb7926d00b71dec55c2fea1f50f7")
)

func TestExtractModifyPositionSampleReceipt(t *testing.T) {
	topic0, ok := modifyPositionEvent.Topic0()
	require.True(t, ok)
	salt := make([]byte, 32)
	salt[0] = 0x0b
	data, err := abi.Pack(abi.Types{abi.Int24, abi.Int24, abi.Int256, abi.Bytes32},
		big.NewInt(-200), big.NewInt(200), big.NewInt(500), salt)
	require.NoError(t, err)

	logs := []tx.Log{
		{
			Address: testPositionManager,
			Topics: []*types.Hash{
				types.NewHashFromBytes(topic0[:]),
				types.NewHashFromBytes([]byte{0x03}),
				types.NewHashFromBytes(testSender.Bytes()),
			},
			Data: data,
		},
	}

	mods, err := ExtractModifyPosition(logs)
	require.NoError(t, err)
	require.Len(t, mods, 1)
	require.True(t, mods[0].Contract.Equal(testPositionManager))
	require.True(t, mods[0].Sender.Equal(testSender))
	require.Zero(t, big.NewInt(-200).Cmp(mods[0].TickLower))
	require.Zero(t, big.NewInt(200).Cmp(mods[0].TickUpper))
	require.Zero(t, big.NewInt(500).Cmp(mods[0].LiquidityDelta))
	require.Equal(t, [32]byte{0x0b}, mods[0].Salt)
}

func TestExtractSubscriptionSampleReceipt(t *testing.T) {
	topic0, ok := subscriptionEvent.Topic0()
	require.True(t, ok)

	logs := []tx.Log{
		{
			Address: testPositionManager,
			Topics: []*types.Hash{
				types.NewHashFromBytes(topic0[:]),
				types.NewHashFromBytes(big.NewInt(7).Bytes()),
				types.NewHashFromBytes(testSubscriber.Bytes()),
			},
		},
	}

	subs, err := ExtractSubscription(logs)
	require.NoError(t, err)
	require.Len(t, subs, 1)
	require.True(t, subs[0].Contract.Equal(testPositionManager))
	require.Zero(t, big.NewInt(7).Cmp(subs[0].TokenID))
	require.True(t, subs[0].Subscriber.Equal(testSubscriber))
}

func TestExtractUnsubscriptionSampleReceipt(t *testing.T) {
	topic0, ok := unsubscriptionEvent.Topic0()
	require.True(t, ok)

	logs := []tx.Log{
		{
			Address: testPositionManager,
			Topics: []*types.Hash{
				types.NewHashFromBytes(topic0[:]),
				types.NewHashFromBytes(big.NewInt(9).Bytes()),
				types.NewHashFromBytes(testSubscriber.Bytes()),
			},
		},
	}

	unsubs, err := ExtractUnsubscription(logs)
	require.NoError(t, err)
	require.Len(t, unsubs, 1)
	require.True(t, unsubs[0].Contract.Equal(testPositionManager))
	require.Zero(t, big.NewInt(9).Cmp(unsubs[0].TokenID))
	require.True(t, unsubs[0].Subscriber.Equal(testSubscriber))
}

func TestExtractModifyPositionEmptyLogs(t *testing.T) {
	mods, err := ExtractModifyPosition(nil)
	require.NoError(t, err)
	require.Empty(t, mods)
}

func TestExtractModifyPositionNoTopics(t *testing.T) {
	logs := []tx.Log{{Address: testPositionManager, Topics: nil, Data: nil}}
	mods, err := ExtractModifyPosition(logs)
	require.NoError(t, err)
	require.Empty(t, mods)
}

func TestExtractModifyPositionUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{
		Address: testPositionManager,
		Topics:  []*types.Hash{unrelated},
		Data:    []byte{0xde, 0xad},
	}}
	mods, err := ExtractModifyPosition(logs)
	require.NoError(t, err)
	require.Empty(t, mods)
}

func TestExtractModifyPositionRejectsNilTopic(t *testing.T) {
	logs := []tx.Log{{
		Address: testPositionManager,
		Topics:  []*types.Hash{nil},
		Data:    nil,
	}}
	mods, err := ExtractModifyPosition(logs)
	require.NoError(t, err)
	require.Empty(t, mods)

	t0, ok := modifyPositionEvent.Topic0()
	require.True(t, ok)
	logs = []tx.Log{{
		Address: testPositionManager,
		Topics:  []*types.Hash{types.NewHashFromBytes(t0[:]), nil, nil},
		Data:    nil,
	}}
	_, err = ExtractModifyPosition(logs)
	require.Error(t, err)
}

func TestExtractModifyPositionRejectsMalformedTopicCount(t *testing.T) {
	t0, ok := modifyPositionEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testPositionManager,
		Topics: []*types.Hash{
			types.NewHashFromBytes(t0[:]),
			types.NewHashFromBytes([]byte{0x03}),
		},
		Data: nil,
	}}
	_, err := ExtractModifyPosition(logs)
	require.Error(t, err)
}

func TestExtractModifyPositionRejectsMalformedData(t *testing.T) {
	t0, ok := modifyPositionEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testPositionManager,
		Topics: []*types.Hash{
			types.NewHashFromBytes(t0[:]),
			types.NewHashFromBytes([]byte{0x03}),
			types.NewHashFromBytes(testSender.Bytes()),
		},
		Data: []byte{0x01, 0x02},
	}}
	_, err := ExtractModifyPosition(logs)
	require.Error(t, err)
}

func TestExtractSubscriptionEmptyLogs(t *testing.T) {
	subs, err := ExtractSubscription(nil)
	require.NoError(t, err)
	require.Empty(t, subs)
}

func TestExtractSubscriptionNoTopics(t *testing.T) {
	logs := []tx.Log{{Address: testPositionManager, Topics: nil, Data: nil}}
	subs, err := ExtractSubscription(logs)
	require.NoError(t, err)
	require.Empty(t, subs)
}

func TestExtractSubscriptionUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x22", 32)))
	logs := []tx.Log{{
		Address: testPositionManager,
		Topics:  []*types.Hash{unrelated},
		Data:    []byte{0xbe, 0xef},
	}}
	subs, err := ExtractSubscription(logs)
	require.NoError(t, err)
	require.Empty(t, subs)
}

func TestExtractSubscriptionRejectsNilTopic(t *testing.T) {
	logs := []tx.Log{{
		Address: testPositionManager,
		Topics:  []*types.Hash{nil},
		Data:    nil,
	}}
	subs, err := ExtractSubscription(logs)
	require.NoError(t, err)
	require.Empty(t, subs)

	t0, ok := subscriptionEvent.Topic0()
	require.True(t, ok)
	logs = []tx.Log{{
		Address: testPositionManager,
		Topics:  []*types.Hash{types.NewHashFromBytes(t0[:]), nil, nil},
		Data:    nil,
	}}
	_, err = ExtractSubscription(logs)
	require.Error(t, err)
}

func TestExtractSubscriptionRejectsMalformedTopicCount(t *testing.T) {
	t0, ok := subscriptionEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testPositionManager,
		Topics: []*types.Hash{
			types.NewHashFromBytes(t0[:]),
			types.NewHashFromBytes(big.NewInt(7).Bytes()),
		},
		Data: nil,
	}}
	_, err := ExtractSubscription(logs)
	require.Error(t, err)
}

func TestExtractUnsubscriptionEmptyLogs(t *testing.T) {
	unsubs, err := ExtractUnsubscription(nil)
	require.NoError(t, err)
	require.Empty(t, unsubs)
}

func TestExtractUnsubscriptionNoTopics(t *testing.T) {
	logs := []tx.Log{{Address: testPositionManager, Topics: nil, Data: nil}}
	unsubs, err := ExtractUnsubscription(logs)
	require.NoError(t, err)
	require.Empty(t, unsubs)
}

func TestExtractUnsubscriptionUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x33", 32)))
	logs := []tx.Log{{
		Address: testPositionManager,
		Topics:  []*types.Hash{unrelated},
		Data:    []byte{0xca, 0xfe},
	}}
	unsubs, err := ExtractUnsubscription(logs)
	require.NoError(t, err)
	require.Empty(t, unsubs)
}

func TestExtractUnsubscriptionRejectsNilTopic(t *testing.T) {
	logs := []tx.Log{{
		Address: testPositionManager,
		Topics:  []*types.Hash{nil},
		Data:    nil,
	}}
	unsubs, err := ExtractUnsubscription(logs)
	require.NoError(t, err)
	require.Empty(t, unsubs)

	t0, ok := unsubscriptionEvent.Topic0()
	require.True(t, ok)
	logs = []tx.Log{{
		Address: testPositionManager,
		Topics:  []*types.Hash{types.NewHashFromBytes(t0[:]), nil, nil},
		Data:    nil,
	}}
	_, err = ExtractUnsubscription(logs)
	require.Error(t, err)
}

func TestExtractUnsubscriptionRejectsMalformedTopicCount(t *testing.T) {
	t0, ok := unsubscriptionEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testPositionManager,
		Topics: []*types.Hash{
			types.NewHashFromBytes(t0[:]),
			types.NewHashFromBytes(big.NewInt(9).Bytes()),
		},
		Data: nil,
	}}
	_, err := ExtractUnsubscription(logs)
	require.Error(t, err)
}
