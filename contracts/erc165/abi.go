// Package erc165 encodes calls to and decodes results from an ERC-165
// introspection call. Go types mirror the interface in
// solidity/interfaces/IERC165.sol.
package erc165

import (
	"errors"
	"fmt"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
)

// Function signatures from IERC165, for encoding calls and decoding results.
var supportsInterfaceFn = abi.NewFunction("supportsInterface", abi.NewTypes(abi.Bytes4), abi.NewTypes(abi.Bool))

// ErrNotImplementedERC165 means contract either reverted the
// supportsInterface call or returned data that doesn't decode as a bool —
// both signs it doesn't correctly implement ERC-165, checkable with errors.Is.
var ErrNotImplementedERC165 = errors.New("erc165: contract does not implement ERC-165")

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
