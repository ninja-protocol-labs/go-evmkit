package weth

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

// go test -run '^$' -fuzz '^FuzzEncodeWithdrawRoundTrip$' -fuzztime=10s ./protocols/weth
func FuzzEncodeWithdrawRoundTrip(f *testing.F) {
	f.Add([]byte{0x01})
	f.Add(make([]byte, 32))

	f.Fuzz(func(t *testing.T, wadRaw []byte) {
		if len(wadRaw) > 32 {
			wadRaw = wadRaw[len(wadRaw)-32:]
		}
		wad := new(big.Int).SetBytes(wadRaw)

		data, err := EncodeWithdraw(wad)
		require.NoError(t, err)
		require.Equal(t, []byte{0x2e, 0x1a, 0x7d, 0x4d}, data[:4])

		vals, err := withdrawFn.DecodeCall(data)
		require.NoError(t, err)
		require.Zero(t, wad.Cmp(vals[0].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzWETHSimpleDecodersDoNotPanic$' -fuzztime=10s ./protocols/weth
func FuzzWETHSimpleDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 33))
	f.Add(make([]byte, 64))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = withdrawFn.DecodeCall(data)
	})
}

// go test -run '^$' -fuzz '^FuzzExtractEventsDoNotPanic$' -fuzztime=10s ./protocols/weth
func FuzzExtractEventsDoNotPanic(f *testing.F) {
	f.Add(uint8(0), []byte{}, []byte{})
	f.Add(uint8(2), make([]byte, 64), make([]byte, 32))
	f.Add(uint8(3), []byte{0xde, 0xad, 0xbe, 0xef}, []byte{0x01})

	f.Fuzz(func(t *testing.T, topicCount uint8, topicsRaw, data []byte) {
		n := int(topicCount % 4)
		topics := make([]*types.Hash, n)
		for i := range topics {
			start := i * 32
			if start >= len(topicsRaw) {
				topics[i] = types.NewHashFromBytes(nil)
				continue
			}
			end := min(start+32, len(topicsRaw))
			topics[i] = types.NewHashFromBytes(topicsRaw[start:end])
		}
		logs := []tx.Log{{Address: testWETH, Topics: topics, Data: data}}

		_, _ = ExtractDeposits(logs)
		_, _ = ExtractWithdrawals(logs)
	})
}
