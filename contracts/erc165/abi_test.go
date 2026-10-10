package erc165

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
)

func TestEncodeSupportsInterface(t *testing.T) {
	interfaceId := [4]byte{0x80, 0xac, 0x58, 0xcd} // IERC721
	data, err := EncodeSupportsInterface(interfaceId)
	require.NoError(t, err)
	require.Equal(t, []byte{0x01, 0xff, 0xc9, 0xa7}, data[:4])

	vals, err := supportsInterfaceFn.DecodeCall(data)
	require.NoError(t, err)
	require.Equal(t, interfaceId[:], vals[0].([]byte))
}

func TestDecodeSupportsInterface(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Bool}, true)
	require.NoError(t, err)
	ok, err := DecodeSupportsInterface(data)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestDecodeSupportsInterfaceRejectsGarbage(t *testing.T) {
	_, err := DecodeSupportsInterface([]byte{0x01, 0x02})
	require.Error(t, err)
}
