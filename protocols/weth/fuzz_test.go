package weth

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
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
