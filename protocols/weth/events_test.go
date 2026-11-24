package weth

import (
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

var (
	testWETH = types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	testUser = types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
)

func TestExtractDepositsSampleLog(t *testing.T) {
	topic0, ok := depositEvent.Topic0()
	require.True(t, ok)

	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(1000))
	require.NoError(t, err)

	log := tx.Log{
		Address: testWETH,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testUser.Bytes()),
		},
		Data: data,
	}

	deposits, err := ExtractDeposits([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, deposits, 1)
	require.True(t, deposits[0].Contract.Equal(testWETH))
	require.True(t, deposits[0].Dst.Equal(testUser))
	require.Zero(t, big.NewInt(1000).Cmp(deposits[0].Wad))
}

func TestExtractWithdrawalsSampleLog(t *testing.T) {
	topic0, ok := withdrawalEvent.Topic0()
	require.True(t, ok)

	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(2000))
	require.NoError(t, err)

	log := tx.Log{
		Address: testWETH,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testUser.Bytes()),
		},
		Data: data,
	}

	withdrawals, err := ExtractWithdrawals([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, withdrawals, 1)
	require.True(t, withdrawals[0].Contract.Equal(testWETH))
	require.True(t, withdrawals[0].Src.Equal(testUser))
	require.Zero(t, big.NewInt(2000).Cmp(withdrawals[0].Wad))
}

func TestExtractDepositsEmptyLogs(t *testing.T) {
	deposits, err := ExtractDeposits(nil)
	require.NoError(t, err)
	require.Empty(t, deposits)
}

func TestExtractDepositsUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{
		Address: testWETH,
		Topics:  []*types.Hash{unrelated},
		Data:    []byte{0xde, 0xad},
	}}
	deposits, err := ExtractDeposits(logs)
	require.NoError(t, err)
	require.Empty(t, deposits)
}

func TestExtractDepositsRejectsNilTopic(t *testing.T) {
	logs := []tx.Log{{
		Address: testWETH,
		Topics:  []*types.Hash{nil},
		Data:    nil,
	}}
	deposits, err := ExtractDeposits(logs)
	require.NoError(t, err)
	require.Empty(t, deposits)

	topic0, ok := depositEvent.Topic0()
	require.True(t, ok)
	logs = []tx.Log{{
		Address: testWETH,
		Topics:  []*types.Hash{types.NewHashFromBytes(topic0[:]), nil},
		Data:    nil,
	}}
	_, err = ExtractDeposits(logs)
	require.Error(t, err)
}

func TestExtractWithdrawalsRejectsMalformedData(t *testing.T) {
	topic0, ok := withdrawalEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testWETH,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testUser.Bytes()),
		},
		Data: []byte{0xde, 0xad},
	}}
	_, err := ExtractWithdrawals(logs)
	require.Error(t, err)
}

func TestExtractTransfersMintAndBurn(t *testing.T) {
	topic0 := core.Keccak256([]byte("Transfer(address,address,uint256)"))
	zero := types.NewHashFromBytes(make([]byte, 32))
	user := types.NewHashFromBytes(append(make([]byte, 12), testUser.Bytes()...))
	minted, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(1000))
	require.NoError(t, err)
	burned, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(400))
	require.NoError(t, err)
	depositTopic, ok := depositEvent.Topic0()
	require.True(t, ok)

	logs := []tx.Log{
		{Address: testWETH, Topics: []*types.Hash{topic0, zero, user}, Data: minted},
		{Address: testWETH, Topics: []*types.Hash{types.NewHashFromBytes(depositTopic[:]), user}, Data: minted},
		{Address: testWETH, Topics: []*types.Hash{topic0, user, zero}, Data: burned},
	}

	transfers, err := ExtractTransfers(logs)
	require.NoError(t, err)
	require.Len(t, transfers, 2)
	require.True(t, transfers[0].Contract.Equal(testWETH))
	require.True(t, transfers[0].From.IsZero())
	require.True(t, transfers[0].To.Equal(testUser))
	require.Zero(t, big.NewInt(1000).Cmp(transfers[0].Value))
	require.True(t, transfers[1].From.Equal(testUser))
	require.True(t, transfers[1].To.IsZero())
	require.Zero(t, big.NewInt(400).Cmp(transfers[1].Value))

	empty, err := ExtractTransfers(nil)
	require.NoError(t, err)
	require.Empty(t, empty)

	_, err = ExtractTransfers([]tx.Log{{Address: testWETH, Topics: []*types.Hash{topic0, zero}, Data: minted}})
	require.ErrorContains(t, err, "weth: extractTransfers: ")
}
