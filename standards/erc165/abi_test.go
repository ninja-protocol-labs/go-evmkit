package erc165

import (
	"encoding/json/jsontext"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

func TestEncodeSupportsInterface(t *testing.T) {
	interfaceId := [4]byte{0x80, 0xac, 0x58, 0xcd}
	data, err := EncodeSupportsInterface(interfaceId)
	require.NoError(t, err)
	require.Equal(t, []byte{0x01, 0xff, 0xc9, 0xa7}, data[:4])

	vals, err := supportsInterfaceFn.DecodeCall(data)
	require.NoError(t, err)
	require.Equal(t, interfaceId[:], vals[0].([]byte))
}

func TestDecodeSupportsInterface(t *testing.T) {
	data, err := abi.Pack(abi.NewTypes(abi.Bool), true)
	require.NoError(t, err)
	ok, err := DecodeSupportsInterface(data)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestDecodeSupportsInterfaceRejectsGarbage(t *testing.T) {
	_, err := DecodeSupportsInterface([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestMapRevertError(t *testing.T) {
	require.NoError(t, MapRevertError(nil))
	var success *rpc.ResponseError
	require.NoError(t, MapRevertError(success))

	plain := fmt.Errorf("dial tcp: connection refused")
	require.Same(t, plain, MapRevertError(plain))

	for _, rpcErr := range []*rpc.ResponseError{
		{Code: 3, Message: "execution reverted"},
		{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"0xdeadbeef"`)},
		{Code: -32000, Message: "execution reverted: invalid opcode"},
	} {
		mapped := MapRevertError(rpcErr)
		require.ErrorIs(t, mapped, ErrNotImplementedERC165)
		got, ok := errors.AsType[*rpc.ResponseError](mapped)
		require.True(t, ok)
		require.Same(t, rpcErr, got)
	}

	wrapped := fmt.Errorf("rpc: %s: %w", rpc.MethodETHCall, &rpc.ResponseError{Code: 3, Message: "execution reverted"})
	require.ErrorIs(t, MapRevertError(wrapped), ErrNotImplementedERC165)
}
