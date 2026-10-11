package erc20

import (
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

var (
	weth     = types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	baseUSDC = types.MustNewAddressFromHex("0x833589fcd6edb6e08f4c7c32d4f71b54bda02913")
)

func realBaseReceiptLogs(t *testing.T) []tx.Log {
	t.Helper()

	mustTestData := func(t *testing.T, hexStr string) []byte {
		t.Helper()
		b, err := encoding.Hex.Decode(hexStr)
		if err != nil {
			panic(err)
		}
		return b
	}

	return []tx.Log{
		{
			Address: types.MustNewAddressFromHex("0x675d5cebae87ecb8acc7f057e95156f7be9697d1"),
			Topics: []*types.Hash{
				types.MustNewHashFromHex("0x3d0ce9bfc3ed7d6862dbb28b2dea94561fe714a1b4d019aa8af39730d1ad7c3d"),
				types.MustNewHashFromHex("0x000000000000000000000000a5c1ce365ddb5a91ff466774ec4bdf8f97cb9f55"),
			},
			Data: mustTestData(t, "0x000000000000000000000000000000000000000000000000000011a77417621c"),
		},
		{
			Address: types.MustNewAddressFromHex("0xa5c1ce365ddb5a91ff466774ec4bdf8f97cb9f55"),
			Topics: []*types.Hash{
				types.MustNewHashFromHex("0x6ded982279c8387ad8a63e73385031a3807c1862e633f06e09d11bcb6e282f60"),
			},
			Data: mustTestData(t, "0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000675d5cebae87ecb8acc7f057e95156f7be9697d1000000000000000000000000000000000000000000000000000011a77417621c"),
		},
		{
			Address: weth,
			Topics: []*types.Hash{
				types.MustNewHashFromHex("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
				types.MustNewHashFromHex("0x00000000000000000000000067d03631fe51b741c0c00c4e16eb662ac84381df"),
				types.MustNewHashFromHex("0x0000000000000000000000006667c8dc9fbfec411e7c1ee2b24de960149f930f"),
			},
			Data: mustTestData(t, "0x0000000000000000000000000000000000000000000000000007cabcaef955e9"),
		},
		{
			Address: weth,
			Topics: []*types.Hash{
				types.MustNewHashFromHex("0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"),
				types.MustNewHashFromHex("0x0000000000000000000000006667c8dc9fbfec411e7c1ee2b24de960149f930f"),
				types.MustNewHashFromHex("0x000000000000000000000000ba12222222228d8ba445958a75a0704d566bf2c8"),
			},
			Data: mustTestData(t, "0x0000000000000000000000000000000000000000000000000007cabcaef955e9"),
		},
		{
			Address: weth,
			Topics: []*types.Hash{
				types.MustNewHashFromHex("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
				types.MustNewHashFromHex("0x0000000000000000000000006667c8dc9fbfec411e7c1ee2b24de960149f930f"),
				types.MustNewHashFromHex("0x000000000000000000000000ba12222222228d8ba445958a75a0704d566bf2c8"),
			},
			Data: mustTestData(t, "0x0000000000000000000000000000000000000000000000000007cabcaef955e9"),
		},
		{
			Address: baseUSDC,
			Topics: []*types.Hash{
				types.MustNewHashFromHex("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
				types.MustNewHashFromHex("0x000000000000000000000000ba12222222228d8ba445958a75a0704d566bf2c8"),
				types.MustNewHashFromHex("0x0000000000000000000000001231deb6f5749ef6ce6943a275a1d3e7486f4eae"),
			},
			Data: mustTestData(t, "0x000000000000000000000000000000000000000000000000000000000059d67d"),
		},
		{
			Address: baseUSDC,
			Topics: []*types.Hash{
				types.MustNewHashFromHex("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
				types.MustNewHashFromHex("0x0000000000000000000000001231deb6f5749ef6ce6943a275a1d3e7486f4eae"),
				types.MustNewHashFromHex("0x00000000000000000000000009aea4b2242abc8bb4bb78d537a67a245a7bec64"),
			},
			Data: mustTestData(t, "0x000000000000000000000000000000000000000000000000000000000059d67d"),
		},
	}
}

func TestExtractTransfersRealReceipt(t *testing.T) {
	transfers, err := ExtractTransfers(realBaseReceiptLogs(t))
	require.NoError(t, err)
	require.Len(t, transfers, 4)

	want := big.NewInt(2193236575213033)
	require.True(t, transfers[0].Contract.Equal(weth))
	require.True(t, transfers[0].From.Equal(types.MustNewAddressFromHex("0x67d03631fe51b741c0c00c4e16eb662ac84381df")))
	require.True(t, transfers[0].To.Equal(types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")))
	require.Zero(t, want.Cmp(transfers[0].Value))

	require.True(t, transfers[1].Contract.Equal(weth))
	require.True(t, transfers[2].Contract.Equal(baseUSDC))
	require.True(t, transfers[3].Contract.Equal(baseUSDC))

	wantUSDC := big.NewInt(5887613)
	require.Zero(t, wantUSDC.Cmp(transfers[2].Value))
	require.Zero(t, wantUSDC.Cmp(transfers[3].Value))
}

func TestExtractApprovalsRealReceipt(t *testing.T) {
	approvals, err := ExtractApprovals(realBaseReceiptLogs(t))
	require.NoError(t, err)
	require.Len(t, approvals, 1)
	require.True(t, approvals[0].Contract.Equal(weth))
	require.True(t, approvals[0].Owner.Equal(types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")))
	require.True(t, approvals[0].Spender.Equal(types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")))
}

func TestExtractTransfersEmptyLogs(t *testing.T) {
	transfers, err := ExtractTransfers(nil)
	require.NoError(t, err)
	require.Empty(t, transfers)
}

func TestExtractTransfersNoTopics(t *testing.T) {
	logs := []tx.Log{{Address: weth, Topics: nil, Data: nil}}
	transfers, err := ExtractTransfers(logs)
	require.NoError(t, err)
	require.Empty(t, transfers)
}

func TestExtractTransfersUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{
		Address: weth,
		Topics:  []*types.Hash{unrelated},
		Data:    []byte{0xde, 0xad},
	}}
	transfers, err := ExtractTransfers(logs)
	require.NoError(t, err)
	require.Empty(t, transfers)
}

func TestExtractTransfersRejectsNilTopic(t *testing.T) {
	logs := []tx.Log{{
		Address: weth,
		Topics:  []*types.Hash{nil},
		Data:    nil,
	}}
	transfers, err := ExtractTransfers(logs)
	require.NoError(t, err)
	require.Empty(t, transfers)

	t0, ok := transferEvent.Topic0()
	require.True(t, ok)
	logs = []tx.Log{{
		Address: weth,
		Topics:  []*types.Hash{types.NewHashFromBytes(t0[:]), nil, nil},
		Data:    nil,
	}}
	_, err = ExtractTransfers(logs)
	require.Error(t, err)
}

func TestExtractTransfersRejectsMalformedData(t *testing.T) {
	t0, ok := transferEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: weth,
		Topics: []*types.Hash{
			types.NewHashFromBytes(t0[:]),
			types.MustNewHashFromHex("0x00000000000000000000000067d03631fe51b741c0c00c4e16eb662ac84381df"),
			types.MustNewHashFromHex("0x0000000000000000000000006667c8dc9fbfec411e7c1ee2b24de960149f930f"),
		},
		Data: []byte{0x01, 0x02},
	}}
	_, err := ExtractTransfers(logs)
	require.Error(t, err)
}

func TestExtractTransfersRejectsHandCraftedShortData(t *testing.T) {
	t0, ok := transferEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: weth,
		Topics: []*types.Hash{
			types.NewHashFromBytes(t0[:]),
			types.MustNewHashFromHex("0x00000000000000000000000067d03631fe51b741c0c00c4e16eb662ac84381df"),
			types.MustNewHashFromHex("0x0000000000000000000000006667c8dc9fbfec411e7c1ee2b24de960149f930f"),
		},
		Data: []byte{0x01},
	}}

	require.NotPanics(t, func() {
		_, err := ExtractTransfers(logs)
		require.Error(t, err)
	})
}
