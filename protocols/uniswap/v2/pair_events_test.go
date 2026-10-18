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

func TestExtractMintSampleLog(t *testing.T) {
	topic0, ok := mintEvent.Topic0()
	require.True(t, ok)

	data, err := abi.Pack(abi.Types{abi.Uint256, abi.Uint256}, big.NewInt(100), big.NewInt(200))
	require.NoError(t, err)

	log := tx.Log{
		Address: testPair,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testTokenA.Bytes()),
		},
		Data: data,
	}

	mints, err := ExtractMint([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, mints, 1)
	require.True(t, mints[0].Contract.Equal(testPair))
	require.True(t, mints[0].Sender.Equal(testTokenA))
	require.Zero(t, big.NewInt(100).Cmp(mints[0].Amount0))
	require.Zero(t, big.NewInt(200).Cmp(mints[0].Amount1))
}

func TestExtractBurnSampleLog(t *testing.T) {
	topic0, ok := burnEvent.Topic0()
	require.True(t, ok)

	data, err := abi.Pack(abi.Types{abi.Uint256, abi.Uint256}, big.NewInt(100), big.NewInt(200))
	require.NoError(t, err)

	log := tx.Log{
		Address: testPair,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testTokenA.Bytes()),
			types.NewHashFromBytes(testTokenB.Bytes()),
		},
		Data: data,
	}

	burns, err := ExtractBurn([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, burns, 1)
	require.True(t, burns[0].Contract.Equal(testPair))
	require.True(t, burns[0].Sender.Equal(testTokenA))
	require.Zero(t, big.NewInt(100).Cmp(burns[0].Amount0))
	require.Zero(t, big.NewInt(200).Cmp(burns[0].Amount1))
	require.True(t, burns[0].To.Equal(testTokenB))
}

func TestExtractSwapSampleLog(t *testing.T) {
	topic0, ok := swapEvent.Topic0()
	require.True(t, ok)

	data, err := abi.Pack(abi.Types{abi.Uint256, abi.Uint256, abi.Uint256, abi.Uint256},
		big.NewInt(100), big.NewInt(0), big.NewInt(0), big.NewInt(95))
	require.NoError(t, err)

	log := tx.Log{
		Address: testPair,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testTokenA.Bytes()),
			types.NewHashFromBytes(testTokenB.Bytes()),
		},
		Data: data,
	}

	swaps, err := ExtractSwap([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, swaps, 1)
	require.True(t, swaps[0].Contract.Equal(testPair))
	require.True(t, swaps[0].Sender.Equal(testTokenA))
	require.Zero(t, big.NewInt(100).Cmp(swaps[0].Amount0In))
	require.Zero(t, big.NewInt(0).Cmp(swaps[0].Amount1In))
	require.Zero(t, big.NewInt(0).Cmp(swaps[0].Amount0Out))
	require.Zero(t, big.NewInt(95).Cmp(swaps[0].Amount1Out))
	require.True(t, swaps[0].To.Equal(testTokenB))
}

func TestExtractSyncSampleLog(t *testing.T) {
	topic0, ok := syncEvent.Topic0()
	require.True(t, ok)

	data, err := abi.Pack(abi.Types{abi.Uint112, abi.Uint112}, big.NewInt(1000), big.NewInt(2000))
	require.NoError(t, err)

	log := tx.Log{
		Address: testPair,
		Topics:  []*types.Hash{types.NewHashFromBytes(topic0[:])},
		Data:    data,
	}

	syncs, err := ExtractSync([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, syncs, 1)
	require.True(t, syncs[0].Contract.Equal(testPair))
	require.Zero(t, big.NewInt(1000).Cmp(syncs[0].Reserve0))
	require.Zero(t, big.NewInt(2000).Cmp(syncs[0].Reserve1))
}

func TestExtractMintEmptyLogs(t *testing.T) {
	mints, err := ExtractMint(nil)
	require.NoError(t, err)
	require.Empty(t, mints)
}

func TestExtractMintUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{
		Address: testPair,
		Topics:  []*types.Hash{unrelated},
		Data:    []byte{0xde, 0xad},
	}}
	mints, err := ExtractMint(logs)
	require.NoError(t, err)
	require.Empty(t, mints)
}

func TestExtractMintRejectsNilTopic(t *testing.T) {
	logs := []tx.Log{{
		Address: testPair,
		Topics:  []*types.Hash{nil},
		Data:    nil,
	}}
	mints, err := ExtractMint(logs)
	require.NoError(t, err)
	require.Empty(t, mints)

	topic0, ok := mintEvent.Topic0()
	require.True(t, ok)
	logs = []tx.Log{{
		Address: testPair,
		Topics:  []*types.Hash{types.NewHashFromBytes(topic0[:]), nil},
		Data:    nil,
	}}
	_, err = ExtractMint(logs)
	require.Error(t, err)
}

func TestExtractBurnRejectsMalformedData(t *testing.T) {
	topic0, ok := burnEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testPair,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testTokenA.Bytes()),
			types.NewHashFromBytes(testTokenB.Bytes()),
		},
		Data: []byte{0xde, 0xad},
	}}
	_, err := ExtractBurn(logs)
	require.Error(t, err)
}

func TestExtractSwapRejectsMalformedTopicCount(t *testing.T) {
	topic0, ok := swapEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testPair,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testTokenA.Bytes()),
		},
		Data: nil,
	}}
	_, err := ExtractSwap(logs)
	require.Error(t, err)
}

func TestExtractSyncRejectsMalformedData(t *testing.T) {
	topic0, ok := syncEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testPair,
		Topics:  []*types.Hash{types.NewHashFromBytes(topic0[:])},
		Data:    []byte{0xde, 0xad},
	}}
	_, err := ExtractSync(logs)
	require.Error(t, err)
}
