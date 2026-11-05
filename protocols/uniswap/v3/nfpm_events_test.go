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

func TestExtractIncreaseLiquiditySampleLog(t *testing.T) {
	topic0, ok := increaseLiquidityEvent.Topic0()
	require.True(t, ok)

	tokenIdWord, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(1))
	require.NoError(t, err)

	data, err := abi.Pack(abi.Types{abi.Uint128, abi.Uint256, abi.Uint256}, big.NewInt(1000), big.NewInt(1), big.NewInt(2))
	require.NoError(t, err)

	log := tx.Log{
		Address: testNfpmFactoryAddr,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(tokenIdWord),
		},
		Data: data,
	}

	increased, err := ExtractIncreaseLiquidity([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, increased, 1)
	require.True(t, increased[0].Contract.Equal(testNfpmFactoryAddr))
	require.Zero(t, big.NewInt(1).Cmp(increased[0].TokenId))
	require.Zero(t, big.NewInt(1000).Cmp(increased[0].Liquidity))
	require.Zero(t, big.NewInt(1).Cmp(increased[0].Amount0))
	require.Zero(t, big.NewInt(2).Cmp(increased[0].Amount1))
}

func TestExtractIncreaseLiquidityEmptyLogs(t *testing.T) {
	increased, err := ExtractIncreaseLiquidity(nil)
	require.NoError(t, err)
	require.Empty(t, increased)
}

func TestExtractIncreaseLiquidityUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{Address: testNfpmFactoryAddr, Topics: []*types.Hash{unrelated}, Data: []byte{0xde, 0xad}}}
	increased, err := ExtractIncreaseLiquidity(logs)
	require.NoError(t, err)
	require.Empty(t, increased)
}

func TestExtractIncreaseLiquidityNoTopics(t *testing.T) {
	logs := []tx.Log{{Address: testNfpmFactoryAddr, Topics: nil, Data: []byte{0xde, 0xad}}}
	increased, err := ExtractIncreaseLiquidity(logs)
	require.NoError(t, err)
	require.Empty(t, increased)
}

func TestExtractIncreaseLiquidityRejectsNilTopic(t *testing.T) {
	logs := []tx.Log{{Address: testNfpmFactoryAddr, Topics: []*types.Hash{nil}, Data: []byte{0xde, 0xad}}}
	increased, err := ExtractIncreaseLiquidity(logs)
	require.NoError(t, err)
	require.Empty(t, increased)
}

func TestExtractIncreaseLiquidityRejectsMalformedData(t *testing.T) {
	topic0, ok := increaseLiquidityEvent.Topic0()
	require.True(t, ok)
	tokenIdWord, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(1))
	require.NoError(t, err)
	logs := []tx.Log{{Address: testNfpmFactoryAddr, Topics: []*types.Hash{types.NewHashFromBytes(topic0[:]), types.NewHashFromBytes(tokenIdWord)}, Data: []byte{0xde, 0xad}}}
	_, err = ExtractIncreaseLiquidity(logs)
	require.Error(t, err)
}

func TestExtractIncreaseLiquidityRejectsMalformedTopicCount(t *testing.T) {
	topic0, ok := increaseLiquidityEvent.Topic0()
	require.True(t, ok)
	data, err := abi.Pack(abi.Types{abi.Uint128, abi.Uint256, abi.Uint256}, big.NewInt(1000), big.NewInt(1), big.NewInt(2))
	require.NoError(t, err)
	logs := []tx.Log{{Address: testNfpmFactoryAddr, Topics: []*types.Hash{types.NewHashFromBytes(topic0[:])}, Data: data}}
	_, err = ExtractIncreaseLiquidity(logs)
	require.Error(t, err)
}

func TestExtractDecreaseLiquiditySampleLog(t *testing.T) {
	topic0, ok := decreaseLiquidityEvent.Topic0()
	require.True(t, ok)

	tokenIdWord, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(1))
	require.NoError(t, err)

	data, err := abi.Pack(abi.Types{abi.Uint128, abi.Uint256, abi.Uint256}, big.NewInt(500), big.NewInt(3), big.NewInt(4))
	require.NoError(t, err)

	log := tx.Log{
		Address: testNfpmFactoryAddr,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(tokenIdWord),
		},
		Data: data,
	}

	decreased, err := ExtractDecreaseLiquidity([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, decreased, 1)
	require.True(t, decreased[0].Contract.Equal(testNfpmFactoryAddr))
	require.Zero(t, big.NewInt(1).Cmp(decreased[0].TokenId))
	require.Zero(t, big.NewInt(500).Cmp(decreased[0].Liquidity))
	require.Zero(t, big.NewInt(3).Cmp(decreased[0].Amount0))
	require.Zero(t, big.NewInt(4).Cmp(decreased[0].Amount1))
}

func TestExtractDecreaseLiquidityEmptyLogs(t *testing.T) {
	decreased, err := ExtractDecreaseLiquidity(nil)
	require.NoError(t, err)
	require.Empty(t, decreased)
}

func TestExtractDecreaseLiquidityUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{Address: testNfpmFactoryAddr, Topics: []*types.Hash{unrelated}, Data: []byte{0xde, 0xad}}}
	decreased, err := ExtractDecreaseLiquidity(logs)
	require.NoError(t, err)
	require.Empty(t, decreased)
}

func TestExtractDecreaseLiquidityRejectsMalformedData(t *testing.T) {
	topic0, ok := decreaseLiquidityEvent.Topic0()
	require.True(t, ok)
	tokenIdWord, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(1))
	require.NoError(t, err)
	logs := []tx.Log{{Address: testNfpmFactoryAddr, Topics: []*types.Hash{types.NewHashFromBytes(topic0[:]), types.NewHashFromBytes(tokenIdWord)}, Data: []byte{0xde, 0xad}}}
	_, err = ExtractDecreaseLiquidity(logs)
	require.Error(t, err)
}

func TestExtractPositionCollectSampleLog(t *testing.T) {
	topic0, ok := positionCollectEvent.Topic0()
	require.True(t, ok)

	tokenIdWord, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(1))
	require.NoError(t, err)

	data, err := abi.Pack(abi.Types{abi.Address, abi.Uint256, abi.Uint256}, testNfpmRecipient, big.NewInt(5), big.NewInt(6))
	require.NoError(t, err)

	log := tx.Log{
		Address: testNfpmFactoryAddr,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(tokenIdWord),
		},
		Data: data,
	}

	collected, err := ExtractPositionCollect([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, collected, 1)
	require.True(t, collected[0].Contract.Equal(testNfpmFactoryAddr))
	require.Zero(t, big.NewInt(1).Cmp(collected[0].TokenId))
	require.True(t, collected[0].Recipient.Equal(testNfpmRecipient))
	require.Zero(t, big.NewInt(5).Cmp(collected[0].Amount0))
	require.Zero(t, big.NewInt(6).Cmp(collected[0].Amount1))
}

func TestExtractPositionCollectEmptyLogs(t *testing.T) {
	collected, err := ExtractPositionCollect(nil)
	require.NoError(t, err)
	require.Empty(t, collected)
}

func TestExtractPositionCollectUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{Address: testNfpmFactoryAddr, Topics: []*types.Hash{unrelated}, Data: []byte{0xde, 0xad}}}
	collected, err := ExtractPositionCollect(logs)
	require.NoError(t, err)
	require.Empty(t, collected)
}

func TestExtractPositionCollectRejectsMalformedData(t *testing.T) {
	topic0, ok := positionCollectEvent.Topic0()
	require.True(t, ok)
	tokenIdWord, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(1))
	require.NoError(t, err)
	logs := []tx.Log{{Address: testNfpmFactoryAddr, Topics: []*types.Hash{types.NewHashFromBytes(topic0[:]), types.NewHashFromBytes(tokenIdWord)}, Data: []byte{0xde, 0xad}}}
	_, err = ExtractPositionCollect(logs)
	require.Error(t, err)
}
