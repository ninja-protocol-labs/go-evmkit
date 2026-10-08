package abi

import (
	"encoding/hex"
	"math/big"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func mustHexBytes(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	require.NoError(t, err)
	return b
}

func topicFromAddress(t *testing.T, addrHex string) Topic {
	t.Helper()
	addr := mustHexBytes(t, addrHex)
	require.Len(t, addr, 20)
	var topic Topic
	copy(topic[12:], addr)
	return topic
}

func TestParseEventBasic(t *testing.T) {
	e, err := ParseEvent("Transfer(address indexed from, address indexed to, uint256 value)")
	require.NoError(t, err)
	require.Equal(t, "Transfer", e.Name)
	require.Equal(t, "Transfer(address,address,uint256)", e.Signature())
	require.False(t, e.Anonymous)
}

func TestParseEventWithKeyword(t *testing.T) {
	e, err := ParseEvent("event Transfer(address indexed from, address indexed to, uint256 value)")
	require.NoError(t, err)
	require.Equal(t, "Transfer", e.Name)
	require.Equal(t, "Transfer(address,address,uint256)", e.Signature())
}

func TestParseEventIndexedFlags(t *testing.T) {
	e, err := ParseEvent("Transfer(address indexed from, address indexed to, uint256 value)")
	require.NoError(t, err)
	require.Len(t, e.Inputs, 3)
	require.True(t, e.Inputs[0].Indexed)
	require.True(t, e.Inputs[1].Indexed)
	require.False(t, e.Inputs[2].Indexed)
}

func TestParseEventIndexedNoName(t *testing.T) {
	e, err := ParseEvent("Foo(uint256 indexed)")
	require.NoError(t, err)
	require.True(t, e.Inputs[0].Indexed)
	require.Equal(t, "Foo(uint256)", e.Signature())
}

func TestParseEventNameStartingWithIndexedIsNotStripped(t *testing.T) {
	e, err := ParseEvent("Foo(uint256 indexedCount)")
	require.NoError(t, err)
	require.False(t, e.Inputs[0].Indexed)
}

func TestParseEventAnonymous(t *testing.T) {
	e, err := ParseEvent("Foo(uint256 indexed x) anonymous")
	require.NoError(t, err)
	require.True(t, e.Anonymous)

	_, ok := e.Topic0()
	require.False(t, ok)
}

func TestParseEventWithKeywordAndAnonymous(t *testing.T) {
	e, err := ParseEvent("event Foo(uint256 indexed x) anonymous")
	require.NoError(t, err)
	require.True(t, e.Anonymous)
}

func TestParseEventNoArgs(t *testing.T) {
	e, err := ParseEvent("Paused()")
	require.NoError(t, err)
	require.Equal(t, "Paused()", e.Signature())
	require.Empty(t, e.Inputs)
}

func TestParseEventTupleAndArrayArgs(t *testing.T) {
	e, err := ParseEvent("BatchProcessed((address,uint256)[] indexed items)")
	require.NoError(t, err)
	require.Equal(t, "BatchProcessed((address,uint256)[])", e.Signature())
	require.True(t, e.Inputs[0].Indexed)
	require.True(t, e.Inputs[0].Type.IsDynamic())
}

func TestParseEventRejectsMissingName(t *testing.T) {
	_, err := ParseEvent("(uint256)")
	require.ErrorIs(t, err, ErrInvalidTypeString)
}

func TestParseEventRejectsMissingParen(t *testing.T) {
	_, err := ParseEvent("Foo")
	require.ErrorIs(t, err, ErrInvalidTypeString)
}

func TestParseEventRejectsTrailingGarbage(t *testing.T) {
	_, err := ParseEvent("Foo(uint256) extra")
	require.ErrorIs(t, err, ErrInvalidTypeString)
}

func TestParseEventRejectsTrailingGarbageAfterAnonymous(t *testing.T) {
	_, err := ParseEvent("Foo(uint256) anonymous extra")
	require.ErrorIs(t, err, ErrInvalidTypeString)
}

func TestParseEventRejectsMalformedParamList(t *testing.T) {
	tests := []string{"Foo(uint256", "Foo(uint256,)", "Foo(,uint256)", "Foo(uint256 a uint256 b)"}
	for _, sig := range tests {
		_, err := ParseEvent(sig)
		require.Error(t, err, sig)
	}
}

func TestNewEventAndParseEventAgree(t *testing.T) {
	byHand := NewEvent("Transfer", []EventParam{
		{Type: Address, Indexed: true},
		{Type: Address, Indexed: true},
		{Type: Uint256},
	}, false)
	parsed, err := ParseEvent("Transfer(address indexed from, address indexed to, uint256 value)")
	require.NoError(t, err)

	require.Equal(t, byHand.Signature(), parsed.Signature())
	t0a, _ := byHand.Topic0()
	t0b, _ := parsed.Topic0()
	require.Equal(t, t0a, t0b)
}

func TestEventTopic0KnownValueTransfer(t *testing.T) {
	e, err := ParseEvent("Transfer(address indexed from, address indexed to, uint256 value)")
	require.NoError(t, err)

	t0, ok := e.Topic0()
	require.True(t, ok)
	require.Equal(t, "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", hex.EncodeToString(t0[:]))
}

func TestEventTopic0Concurrent(t *testing.T) {
	e := NewEvent("Transfer", []EventParam{
		{Type: Address, Indexed: true},
		{Type: Address, Indexed: true},
		{Type: Uint256},
	}, false)

	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			a, _ := e.Topic0()
			b, _ := e.Topic0()
			require.Equal(t, a, b)
		})
	}
	wg.Wait()
}

func TestEventDecodeRealTransferLog(t *testing.T) {
	e, err := ParseEvent("Transfer(address indexed from, address indexed to, uint256 value)")
	require.NoError(t, err)

	t0, _ := e.Topic0()
	topics := []Topic{
		t0,
		topicFromAddress(t, "df6142dfa49fd954e660b1f5425698142194e074"),
		topicFromAddress(t, "05bd7a3d18324c1f7e216f7fbf2b15985ae5281a"),
	}
	data, err := Pack(Types{Uint256}, big.NewInt(13301442975181365))
	require.NoError(t, err)

	vals, err := e.Decode(topics, data)
	require.NoError(t, err)
	require.Len(t, vals, 3)

	wantFrom, err := types.NewAddressFromHex("0xdf6142dfa49fd954e660b1f5425698142194e074")
	require.NoError(t, err)
	wantTo, err := types.NewAddressFromHex("0x05bd7a3d18324c1f7e216f7fbf2b15985ae5281a")
	require.NoError(t, err)

	require.True(t, vals[0].(*types.Address).Equal(wantFrom))
	require.True(t, vals[1].(*types.Address).Equal(wantTo))
	require.Zero(t, vals[2].(*big.Int).Cmp(big.NewInt(13301442975181365)))
}

func TestEventDecodeIndexedDynamicReturnsRawTopic(t *testing.T) {
	e, err := ParseEvent("Foo(string indexed s, uint256 x)")
	require.NoError(t, err)

	t0, _ := e.Topic0()
	var sTopic Topic
	sTopic[31] = 0x01
	data, err := Pack(Types{Uint256}, big.NewInt(42))
	require.NoError(t, err)

	vals, err := e.Decode([]Topic{t0, sTopic}, data)
	require.NoError(t, err)
	require.Equal(t, sTopic, vals[0])
	require.Zero(t, vals[1].(*big.Int).Cmp(big.NewInt(42)))
}

func TestEventDecodeRejectsTopicMismatch(t *testing.T) {
	e, err := ParseEvent("Foo(uint256 indexed x)")
	require.NoError(t, err)

	t0, _ := e.Topic0()
	wrongTopic0 := t0
	wrongTopic0[0] ^= 0xff
	_, err = e.Decode([]Topic{wrongTopic0, {}}, nil)
	require.ErrorIs(t, err, ErrTopicMismatch)
}

func TestEventDecodeRejectsMissingTopic0(t *testing.T) {
	e, err := ParseEvent("Foo(uint256 indexed x)")
	require.NoError(t, err)

	_, err = e.Decode(nil, nil)
	require.ErrorIs(t, err, ErrByteLengthMismatch)
}

func TestEventDecodeRejectsTooFewIndexedTopics(t *testing.T) {
	e, err := ParseEvent("Foo(uint256 indexed x, uint256 indexed y)")
	require.NoError(t, err)

	t0, _ := e.Topic0()
	_, err = e.Decode([]Topic{t0, {}}, nil)
	require.ErrorIs(t, err, ErrByteLengthMismatch)
}

func TestEventDecodeRejectsTooManyTopics(t *testing.T) {
	e, err := ParseEvent("Foo(uint256 indexed x)")
	require.NoError(t, err)

	t0, _ := e.Topic0()
	_, err = e.Decode([]Topic{t0, {}, {}}, nil)
	require.ErrorIs(t, err, ErrByteLengthMismatch)
}

func TestEventDecodeRejectsMalformedData(t *testing.T) {
	e, err := ParseEvent("Foo(uint256 x)")
	require.NoError(t, err)

	t0, _ := e.Topic0()
	_, err = e.Decode([]Topic{t0}, make([]byte, 10))
	require.Error(t, err)
}

func TestEventDecodeAnonymousSkipsTopic0Check(t *testing.T) {
	e, err := ParseEvent("Foo(uint256 indexed x) anonymous")
	require.NoError(t, err)

	xTopic, err := Pack(Types{Uint256}, big.NewInt(7))
	require.NoError(t, err)
	var topic Topic
	copy(topic[:], xTopic)

	vals, err := e.Decode([]Topic{topic}, nil)
	require.NoError(t, err)
	require.Zero(t, vals[0].(*big.Int).Cmp(big.NewInt(7)))
}

func TestEventDecodeNoArgs(t *testing.T) {
	e, err := ParseEvent("Paused()")
	require.NoError(t, err)

	t0, _ := e.Topic0()
	vals, err := e.Decode([]Topic{t0}, nil)
	require.NoError(t, err)
	require.Empty(t, vals)
}
