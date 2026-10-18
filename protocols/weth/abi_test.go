package weth

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeDeposit(t *testing.T) {
	require.Equal(t, []byte{0xd0, 0xe3, 0x0d, 0xb0}, EncodeDeposit())
}

func TestEncodeWithdraw(t *testing.T) {
	data, err := EncodeWithdraw(big.NewInt(500))
	require.NoError(t, err)
	require.Equal(t, []byte{0x2e, 0x1a, 0x7d, 0x4d}, data[:4])

	vals, err := withdrawFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(500).Cmp(vals[0].(*big.Int)))
}
