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

func TestExtractInitializeSampleLog(t *testing.T) {
	topic0, ok := initializeEvent.Topic0()
	require.True(t, ok)

	data, err := abi.Pack(abi.Types{abi.Uint160, abi.Int24}, big.NewInt(1000), big.NewInt(-50))
	require.NoError(t, err)

	log := tx.Log{
		Address: testPool,
		Topics:  []*types.Hash{types.NewHashFromBytes(topic0[:])},
		Data:    data,
	}

	initialized, err := ExtractInitialize([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, initialized, 1)
	require.True(t, initialized[0].Contract.Equal(testPool))
	require.Zero(t, big.NewInt(1000).Cmp(initialized[0].SqrtPriceX96))
	require.Zero(t, big.NewInt(-50).Cmp(initialized[0].Tick))
}

func TestExtractInitializeEmptyLogs(t *testing.T) {
	initialized, err := ExtractInitialize(nil)
	require.NoError(t, err)
	require.Empty(t, initialized)
}

func TestExtractInitializeUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{Address: testPool, Topics: []*types.Hash{unrelated}, Data: []byte{0xde, 0xad}}}
	initialized, err := ExtractInitialize(logs)
	require.NoError(t, err)
	require.Empty(t, initialized)
}

func TestExtractInitializeRejectsMalformedData(t *testing.T) {
	topic0, ok := initializeEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{Address: testPool, Topics: []*types.Hash{types.NewHashFromBytes(topic0[:])}, Data: []byte{0xde, 0xad}}}
	_, err := ExtractInitialize(logs)
	require.Error(t, err)
}

func TestExtractMintSampleLog(t *testing.T) {
	topic0, ok := mintEvent.Topic0()
	require.True(t, ok)

	tickLowerWord, err := abi.Pack(abi.Types{abi.Int24}, big.NewInt(-100))
	require.NoError(t, err)
	tickUpperWord, err := abi.Pack(abi.Types{abi.Int24}, big.NewInt(100))
	require.NoError(t, err)

	data, err := abi.Pack(abi.Types{abi.Address, abi.Uint128, abi.Uint256, abi.Uint256}, testTokenA, big.NewInt(500), big.NewInt(1), big.NewInt(2))
	require.NoError(t, err)

	log := tx.Log{
		Address: testPool,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testOwner.Bytes()),
			types.NewHashFromBytes(tickLowerWord),
			types.NewHashFromBytes(tickUpperWord),
		},
		Data: data,
	}

	minted, err := ExtractMint([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, minted, 1)
	require.True(t, minted[0].Contract.Equal(testPool))
	require.True(t, minted[0].Sender.Equal(testTokenA))
	require.True(t, minted[0].Owner.Equal(testOwner))
	require.Zero(t, big.NewInt(-100).Cmp(minted[0].TickLower))
	require.Zero(t, big.NewInt(100).Cmp(minted[0].TickUpper))
	require.Zero(t, big.NewInt(500).Cmp(minted[0].Amount))
	require.Zero(t, big.NewInt(1).Cmp(minted[0].Amount0))
	require.Zero(t, big.NewInt(2).Cmp(minted[0].Amount1))
}

func TestExtractMintEmptyLogs(t *testing.T) {
	minted, err := ExtractMint(nil)
	require.NoError(t, err)
	require.Empty(t, minted)
}

func TestExtractMintNoTopics(t *testing.T) {
	minted, err := ExtractMint([]tx.Log{{Address: testPool, Topics: nil, Data: nil}})
	require.NoError(t, err)
	require.Empty(t, minted)
}

func TestExtractMintUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{Address: testPool, Topics: []*types.Hash{unrelated}, Data: []byte{0xde, 0xad}}}
	minted, err := ExtractMint(logs)
	require.NoError(t, err)
	require.Empty(t, minted)
}

func TestExtractMintRejectsNilTopic(t *testing.T) {
	logs := []tx.Log{{Address: testPool, Topics: []*types.Hash{nil}, Data: nil}}
	minted, err := ExtractMint(logs)
	require.NoError(t, err)
	require.Empty(t, minted)

	topic0, ok := mintEvent.Topic0()
	require.True(t, ok)
	logs = []tx.Log{{Address: testPool, Topics: []*types.Hash{types.NewHashFromBytes(topic0[:]), nil, nil, nil}, Data: nil}}
	_, err = ExtractMint(logs)
	require.Error(t, err)
}

func TestExtractMintRejectsMalformedTopicCount(t *testing.T) {
	topic0, ok := mintEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testPool,
		Topics:  []*types.Hash{types.NewHashFromBytes(topic0[:]), types.NewHashFromBytes(testOwner.Bytes())},
		Data:    nil,
	}}
	_, err := ExtractMint(logs)
	require.Error(t, err)
}

func TestExtractMintRejectsMalformedData(t *testing.T) {
	topic0, ok := mintEvent.Topic0()
	require.True(t, ok)
	tickLowerWord, err := abi.Pack(abi.Types{abi.Int24}, big.NewInt(-100))
	require.NoError(t, err)
	tickUpperWord, err := abi.Pack(abi.Types{abi.Int24}, big.NewInt(100))
	require.NoError(t, err)
	logs := []tx.Log{{
		Address: testPool,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testOwner.Bytes()),
			types.NewHashFromBytes(tickLowerWord),
			types.NewHashFromBytes(tickUpperWord),
		},
		Data: []byte{0xde, 0xad},
	}}
	_, err = ExtractMint(logs)
	require.Error(t, err)
}

func TestExtractCollectSampleLog(t *testing.T) {
	topic0, ok := collectEvent.Topic0()
	require.True(t, ok)

	tickLowerWord, err := abi.Pack(abi.Types{abi.Int24}, big.NewInt(-200))
	require.NoError(t, err)
	tickUpperWord, err := abi.Pack(abi.Types{abi.Int24}, big.NewInt(200))
	require.NoError(t, err)

	data, err := abi.Pack(abi.Types{abi.Address, abi.Uint128, abi.Uint128}, testTokenB, big.NewInt(10), big.NewInt(20))
	require.NoError(t, err)

	log := tx.Log{
		Address: testPool,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testOwner.Bytes()),
			types.NewHashFromBytes(tickLowerWord),
			types.NewHashFromBytes(tickUpperWord),
		},
		Data: data,
	}

	collected, err := ExtractCollect([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, collected, 1)
	require.True(t, collected[0].Owner.Equal(testOwner))
	require.True(t, collected[0].Recipient.Equal(testTokenB))
	require.Zero(t, big.NewInt(-200).Cmp(collected[0].TickLower))
	require.Zero(t, big.NewInt(200).Cmp(collected[0].TickUpper))
	require.Zero(t, big.NewInt(10).Cmp(collected[0].Amount0))
	require.Zero(t, big.NewInt(20).Cmp(collected[0].Amount1))
}

func TestExtractCollectEmptyLogs(t *testing.T) {
	collected, err := ExtractCollect(nil)
	require.NoError(t, err)
	require.Empty(t, collected)
}

func TestExtractCollectUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{Address: testPool, Topics: []*types.Hash{unrelated}, Data: []byte{0xde, 0xad}}}
	collected, err := ExtractCollect(logs)
	require.NoError(t, err)
	require.Empty(t, collected)
}

func TestExtractCollectRejectsMalformedTopicCount(t *testing.T) {
	topic0, ok := collectEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testPool,
		Topics:  []*types.Hash{types.NewHashFromBytes(topic0[:]), types.NewHashFromBytes(testOwner.Bytes())},
		Data:    nil,
	}}
	_, err := ExtractCollect(logs)
	require.Error(t, err)
}

func TestExtractBurnSampleLog(t *testing.T) {
	topic0, ok := poolBurnEvent.Topic0()
	require.True(t, ok)

	tickLowerWord, err := abi.Pack(abi.Types{abi.Int24}, big.NewInt(-300))
	require.NoError(t, err)
	tickUpperWord, err := abi.Pack(abi.Types{abi.Int24}, big.NewInt(300))
	require.NoError(t, err)

	data, err := abi.Pack(abi.Types{abi.Uint128, abi.Uint256, abi.Uint256}, big.NewInt(5), big.NewInt(6), big.NewInt(7))
	require.NoError(t, err)

	log := tx.Log{
		Address: testPool,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testOwner.Bytes()),
			types.NewHashFromBytes(tickLowerWord),
			types.NewHashFromBytes(tickUpperWord),
		},
		Data: data,
	}

	burned, err := ExtractBurn([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, burned, 1)
	require.True(t, burned[0].Owner.Equal(testOwner))
	require.Zero(t, big.NewInt(-300).Cmp(burned[0].TickLower))
	require.Zero(t, big.NewInt(300).Cmp(burned[0].TickUpper))
	require.Zero(t, big.NewInt(5).Cmp(burned[0].Amount))
	require.Zero(t, big.NewInt(6).Cmp(burned[0].Amount0))
	require.Zero(t, big.NewInt(7).Cmp(burned[0].Amount1))
}

func TestExtractBurnEmptyLogs(t *testing.T) {
	burned, err := ExtractBurn(nil)
	require.NoError(t, err)
	require.Empty(t, burned)
}

func TestExtractBurnUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{Address: testPool, Topics: []*types.Hash{unrelated}, Data: []byte{0xde, 0xad}}}
	burned, err := ExtractBurn(logs)
	require.NoError(t, err)
	require.Empty(t, burned)
}

func TestExtractBurnRejectsMalformedData(t *testing.T) {
	topic0, ok := poolBurnEvent.Topic0()
	require.True(t, ok)
	tickLowerWord, err := abi.Pack(abi.Types{abi.Int24}, big.NewInt(-300))
	require.NoError(t, err)
	tickUpperWord, err := abi.Pack(abi.Types{abi.Int24}, big.NewInt(300))
	require.NoError(t, err)
	logs := []tx.Log{{
		Address: testPool,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testOwner.Bytes()),
			types.NewHashFromBytes(tickLowerWord),
			types.NewHashFromBytes(tickUpperWord),
		},
		Data: []byte{0xde, 0xad},
	}}
	_, err = ExtractBurn(logs)
	require.Error(t, err)
}

func TestExtractSwapSampleLog(t *testing.T) {
	topic0, ok := swapEvent.Topic0()
	require.True(t, ok)

	data, err := abi.Pack(abi.Types{abi.Int256, abi.Int256, abi.Uint160, abi.Uint128, abi.Int24},
		big.NewInt(1000), big.NewInt(-999), big.NewInt(12345), big.NewInt(500), big.NewInt(-10))
	require.NoError(t, err)

	log := tx.Log{
		Address: testPool,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testTokenA.Bytes()),
			types.NewHashFromBytes(testTokenB.Bytes()),
		},
		Data: data,
	}

	swapped, err := ExtractSwap([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, swapped, 1)
	require.True(t, swapped[0].Sender.Equal(testTokenA))
	require.True(t, swapped[0].Recipient.Equal(testTokenB))
	require.Zero(t, big.NewInt(1000).Cmp(swapped[0].Amount0))
	require.Zero(t, big.NewInt(-999).Cmp(swapped[0].Amount1))
	require.Zero(t, big.NewInt(12345).Cmp(swapped[0].SqrtPriceX96))
	require.Zero(t, big.NewInt(500).Cmp(swapped[0].Liquidity))
	require.Zero(t, big.NewInt(-10).Cmp(swapped[0].Tick))
}

func TestExtractSwapEmptyLogs(t *testing.T) {
	swapped, err := ExtractSwap(nil)
	require.NoError(t, err)
	require.Empty(t, swapped)
}

func TestExtractSwapUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{Address: testPool, Topics: []*types.Hash{unrelated}, Data: []byte{0xde, 0xad}}}
	swapped, err := ExtractSwap(logs)
	require.NoError(t, err)
	require.Empty(t, swapped)
}

func TestExtractSwapRejectsMalformedTopicCount(t *testing.T) {
	topic0, ok := swapEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testPool,
		Topics:  []*types.Hash{types.NewHashFromBytes(topic0[:]), types.NewHashFromBytes(testTokenA.Bytes())},
		Data:    nil,
	}}
	_, err := ExtractSwap(logs)
	require.Error(t, err)
}

func TestExtractSwapNegativeAmountsRoundTrip(t *testing.T) {
	topic0, ok := swapEvent.Topic0()
	require.True(t, ok)

	data, err := abi.Pack(abi.Types{abi.Int256, abi.Int256, abi.Uint160, abi.Uint128, abi.Int24},
		big.NewInt(-1000000), big.NewInt(500), big.NewInt(1), big.NewInt(0), big.NewInt(-887272))
	require.NoError(t, err)

	log := tx.Log{
		Address: testPool,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testTokenA.Bytes()),
			types.NewHashFromBytes(testTokenB.Bytes()),
		},
		Data: data,
	}

	swapped, err := ExtractSwap([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, swapped, 1)
	require.Zero(t, big.NewInt(-1000000).Cmp(swapped[0].Amount0))
	require.Zero(t, big.NewInt(500).Cmp(swapped[0].Amount1))
	require.Zero(t, big.NewInt(-887272).Cmp(swapped[0].Tick))
}

func TestExtractFlashSampleLog(t *testing.T) {
	topic0, ok := flashEvent.Topic0()
	require.True(t, ok)

	data, err := abi.Pack(abi.Types{abi.Uint256, abi.Uint256, abi.Uint256, abi.Uint256}, big.NewInt(1), big.NewInt(2), big.NewInt(3), big.NewInt(4))
	require.NoError(t, err)

	log := tx.Log{
		Address: testPool,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testTokenA.Bytes()),
			types.NewHashFromBytes(testTokenB.Bytes()),
		},
		Data: data,
	}

	flashed, err := ExtractFlash([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, flashed, 1)
	require.True(t, flashed[0].Sender.Equal(testTokenA))
	require.True(t, flashed[0].Recipient.Equal(testTokenB))
	require.Zero(t, big.NewInt(1).Cmp(flashed[0].Amount0))
	require.Zero(t, big.NewInt(2).Cmp(flashed[0].Amount1))
	require.Zero(t, big.NewInt(3).Cmp(flashed[0].Paid0))
	require.Zero(t, big.NewInt(4).Cmp(flashed[0].Paid1))
}

func TestExtractFlashEmptyLogs(t *testing.T) {
	flashed, err := ExtractFlash(nil)
	require.NoError(t, err)
	require.Empty(t, flashed)
}

func TestExtractFlashUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{Address: testPool, Topics: []*types.Hash{unrelated}, Data: []byte{0xde, 0xad}}}
	flashed, err := ExtractFlash(logs)
	require.NoError(t, err)
	require.Empty(t, flashed)
}

func TestExtractFlashRejectsMalformedData(t *testing.T) {
	topic0, ok := flashEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testPool,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testTokenA.Bytes()),
			types.NewHashFromBytes(testTokenB.Bytes()),
		},
		Data: []byte{0xde, 0xad},
	}}
	_, err := ExtractFlash(logs)
	require.Error(t, err)
}

func TestExtractIncreaseObservationCardinalityNextSampleLog(t *testing.T) {
	topic0, ok := increaseObservationCardinalityNextEvent.Topic0()
	require.True(t, ok)

	data, err := abi.Pack(abi.Types{abi.Uint16, abi.Uint16}, uint16(100), uint16(200))
	require.NoError(t, err)

	log := tx.Log{
		Address: testPool,
		Topics:  []*types.Hash{types.NewHashFromBytes(topic0[:])},
		Data:    data,
	}

	increased, err := ExtractIncreaseObservationCardinalityNext([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, increased, 1)
	require.Zero(t, big.NewInt(100).Cmp(increased[0].Old))
	require.Zero(t, big.NewInt(200).Cmp(increased[0].New))
}

func TestExtractIncreaseObservationCardinalityNextEmptyLogs(t *testing.T) {
	increased, err := ExtractIncreaseObservationCardinalityNext(nil)
	require.NoError(t, err)
	require.Empty(t, increased)
}

func TestExtractIncreaseObservationCardinalityNextUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{Address: testPool, Topics: []*types.Hash{unrelated}, Data: []byte{0xde, 0xad}}}
	increased, err := ExtractIncreaseObservationCardinalityNext(logs)
	require.NoError(t, err)
	require.Empty(t, increased)
}

func TestExtractSetFeeProtocolSampleLog(t *testing.T) {
	topic0, ok := setFeeProtocolEvent.Topic0()
	require.True(t, ok)

	data, err := abi.Pack(abi.Types{abi.Uint8, abi.Uint8, abi.Uint8, abi.Uint8}, uint8(0), uint8(0), uint8(4), uint8(10))
	require.NoError(t, err)

	log := tx.Log{
		Address: testPool,
		Topics:  []*types.Hash{types.NewHashFromBytes(topic0[:])},
		Data:    data,
	}

	set, err := ExtractSetFeeProtocol([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, set, 1)
	require.Zero(t, big.NewInt(0).Cmp(set[0].FeeProtocol0Old))
	require.Zero(t, big.NewInt(0).Cmp(set[0].FeeProtocol1Old))
	require.Zero(t, big.NewInt(4).Cmp(set[0].FeeProtocol0New))
	require.Zero(t, big.NewInt(10).Cmp(set[0].FeeProtocol1New))
}

func TestExtractSetFeeProtocolEmptyLogs(t *testing.T) {
	set, err := ExtractSetFeeProtocol(nil)
	require.NoError(t, err)
	require.Empty(t, set)
}

func TestExtractSetFeeProtocolUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{Address: testPool, Topics: []*types.Hash{unrelated}, Data: []byte{0xde, 0xad}}}
	set, err := ExtractSetFeeProtocol(logs)
	require.NoError(t, err)
	require.Empty(t, set)
}

func TestExtractCollectProtocolSampleLog(t *testing.T) {
	topic0, ok := collectProtocolEvent.Topic0()
	require.True(t, ok)

	data, err := abi.Pack(abi.Types{abi.Uint128, abi.Uint128}, big.NewInt(1), big.NewInt(2))
	require.NoError(t, err)

	log := tx.Log{
		Address: testPool,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testTokenA.Bytes()),
			types.NewHashFromBytes(testTokenB.Bytes()),
		},
		Data: data,
	}

	collected, err := ExtractCollectProtocol([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, collected, 1)
	require.True(t, collected[0].Sender.Equal(testTokenA))
	require.True(t, collected[0].Recipient.Equal(testTokenB))
	require.Zero(t, big.NewInt(1).Cmp(collected[0].Amount0))
	require.Zero(t, big.NewInt(2).Cmp(collected[0].Amount1))
}

func TestExtractCollectProtocolEmptyLogs(t *testing.T) {
	collected, err := ExtractCollectProtocol(nil)
	require.NoError(t, err)
	require.Empty(t, collected)
}

func TestExtractCollectProtocolUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{Address: testPool, Topics: []*types.Hash{unrelated}, Data: []byte{0xde, 0xad}}}
	collected, err := ExtractCollectProtocol(logs)
	require.NoError(t, err)
	require.Empty(t, collected)
}

func TestExtractCollectProtocolRejectsNilTopic(t *testing.T) {
	logs := []tx.Log{{Address: testPool, Topics: []*types.Hash{nil}, Data: nil}}
	collected, err := ExtractCollectProtocol(logs)
	require.NoError(t, err)
	require.Empty(t, collected)

	topic0, ok := collectProtocolEvent.Topic0()
	require.True(t, ok)
	logs = []tx.Log{{Address: testPool, Topics: []*types.Hash{types.NewHashFromBytes(topic0[:]), nil, nil}, Data: nil}}
	_, err = ExtractCollectProtocol(logs)
	require.Error(t, err)
}

func TestExtractCollectProtocolRejectsMalformedTopicCount(t *testing.T) {
	topic0, ok := collectProtocolEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testPool,
		Topics:  []*types.Hash{types.NewHashFromBytes(topic0[:]), types.NewHashFromBytes(testTokenA.Bytes())},
		Data:    nil,
	}}
	_, err := ExtractCollectProtocol(logs)
	require.Error(t, err)
}
