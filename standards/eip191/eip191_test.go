package eip191

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core"
)

func TestHashHello(t *testing.T) {
	got := Hash([]byte("hello"))
	want, err := hex.DecodeString("50b2c43fd39106bafbba0da34fc430e1f91e3c96ea2acee2bc34119f92b37750")
	require.NoError(t, err)
	require.Equal(t, want, got.Bytes())
}

func TestHashEmpty(t *testing.T) {
	got := Hash([]byte{})
	want := core.Keccak256([]byte("\x19Ethereum Signed Message:\n0"))
	require.Equal(t, want.Bytes(), got.Bytes())
}

func TestHashDiffersFromRawKeccak256(t *testing.T) {
	msg := []byte("hello")
	require.NotEqual(t, core.Keccak256(msg).Bytes(), Hash(msg).Bytes())
}
