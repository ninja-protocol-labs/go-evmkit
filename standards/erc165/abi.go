package erc165

import (
	"errors"
	"fmt"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

// Function signatures from IERC165, for encoding calls and decoding results.
var supportsInterfaceFn = abi.NewFunction("supportsInterface", abi.NewTypes(abi.Bytes4), abi.NewTypes(abi.Bool))

// ErrNotImplementedERC165 means the contract reverted supportsInterface or
// returned data that doesn't decode as a bool.
var ErrNotImplementedERC165 = errors.New("contract does not implement ERC-165")

// EncodeSupportsInterface returns the calldata for supportsInterface(interfaceId).
func EncodeSupportsInterface(interfaceId [4]byte) ([]byte, error) {
	data, err := supportsInterfaceFn.EncodeCall(interfaceId[:])
	if err != nil {
		return nil, fmt.Errorf("erc165: encodeSupportsInterface: %w", err)
	}
	return data, nil
}

// DecodeSupportsInterface decodes the return data of supportsInterface.
func DecodeSupportsInterface(data []byte) (bool, error) {
	ok, err := supportsInterfaceFn.DecodeSingleReturn[bool](data)
	if err != nil {
		return false, fmt.Errorf("erc165: decodeSupportsInterface: %w", err)
	}
	return ok, nil
}

// MapRevertError maps an on-chain revert to ErrNotImplementedERC165; a transport failure says nothing about the contract and is returned unchanged.
func MapRevertError(err error) error {
	if rpcErr, ok := errors.AsType[*rpc.ResponseError](err); ok && rpcErr == nil {
		return nil
	}

	if _, ok := errors.AsType[*rpc.ResponseError](err); !ok {
		return err
	}
	return fmt.Errorf("%w: %w", ErrNotImplementedERC165, err)
}
