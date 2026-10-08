package abi

import (
	"fmt"
)

// errorSelector and panicSelector are the first 4 bytes of
// keccak256("Error(string)") and keccak256("Panic(uint256)"): the standard
// encodings Solidity uses for a require/revert reason and a
// compiler-inserted panic (overflow, array out-of-bounds, ...),
// respectively.
//
// errorSelector was confirmed against live revert data from a deployed
// Multicall3 contract in this repo's own testing. panicSelector is
// confirmed by OpenZeppelin's Panic.sol (mstore(0x00, 0x4e487b71) in
// _revert), https://github.com/OpenZeppelin/openzeppelin-contracts/blob/v5.1.0/contracts/utils/Panic.sol.
var (
	errorSelector = Selector{0x08, 0xc3, 0x79, 0xa0}
	panicSelector = Selector{0x4e, 0x48, 0x7b, 0x71}
)

// DecodeRevert decodes data — e.g. the Data field of a JSON-RPC
// "execution reverted" error — as a Solidity require/revert reason
// (Error(string)) or a compiler panic (Panic(uint256)), returning a
// human-readable message. It returns ErrSelectorMismatch if data isn't
// either standard encoding.
//
// TODO: Solidity custom errors (e.g. "error InsufficientBalance(uint256,
// uint256)") revert with their own selector and arbitrary argument list,
// the same way a function call does. Decoding those needs something like
// ParseFunction for error declarations — a signature-string parser plus a
// matching Decode, not just the two built-in encodings here.
func DecodeRevert(data []byte) (string, error) {
	if len(data) < 4 {
		return "", fmt.Errorf("%w: revert data must be at least 4 bytes, got %d", ErrByteLengthMismatch, len(data))
	}

	var sel Selector
	copy(sel[:], data[:4])

	switch sel {
	case errorSelector:
		vals, err := Unpack(Types{String}, data[4:])
		if err != nil {
			return "", fmt.Errorf("abi: decode revert reason: %w", err)
		}
		reason, _ := vals[0].(string)
		return reason, nil

	case panicSelector:
		vals, err := Unpack(Types{Uint256}, data[4:])
		if err != nil {
			return "", fmt.Errorf("abi: decode panic code: %w", err)
		}
		return fmt.Sprintf("panic code 0x%x", vals[0]), nil

	default:
		return "", fmt.Errorf("%w: %x is neither Error(string) nor Panic(uint256)", ErrSelectorMismatch, sel)
	}
}
