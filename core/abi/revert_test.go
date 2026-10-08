package abi

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func mustDecodeHexRevert(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	require.NoError(t, err)
	return b
}

func TestDecodeRevertErrorString(t *testing.T) {
	// revert data captured live from Multicall3's aggregate(): require(success, "Multicall3: call failed")
	data := mustDecodeHexRevert(t, "08c379a0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000174d756c746963616c6c333a2063616c6c206661696c6564000000000000000000")

	reason, err := DecodeRevert(data)
	require.NoError(t, err)
	require.Equal(t, "Multicall3: call failed", reason)
}

func TestDecodeRevertValueMismatch(t *testing.T) {
	// revert data captured live from Multicall3's aggregate3Value(): require(msg.value == total, "Multicall3: value mismatch")
	data := mustDecodeHexRevert(t, "08c379a00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000001a4d756c746963616c6c333a2076616c7565206d69736d61746368000000000000")

	reason, err := DecodeRevert(data)
	require.NoError(t, err)
	require.Equal(t, "Multicall3: value mismatch", reason)
}

func TestDecodeRevertPanic(t *testing.T) {
	// Panic(uint256) with code 0x11 (arithmetic overflow/underflow)
	data := mustDecodeHexRevert(t, "4e487b710000000000000000000000000000000000000000000000000000000000000011")

	reason, err := DecodeRevert(data)
	require.NoError(t, err)
	require.Equal(t, "panic code 0x11", reason)
}

func TestDecodeRevertRejectsUnknownSelector(t *testing.T) {
	_, err := DecodeRevert([]byte{0xde, 0xad, 0xbe, 0xef})
	require.ErrorIs(t, err, ErrSelectorMismatch)
}

func TestDecodeRevertRejectsShortData(t *testing.T) {
	_, err := DecodeRevert([]byte{0x01, 0x02})
	require.ErrorIs(t, err, ErrByteLengthMismatch)
}
