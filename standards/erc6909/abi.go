package erc6909

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Function signatures from IERC6909/IERC6909Metadata, for encoding calls and decoding results.
var (
	balanceOfFn    = abi.NewFunction("balanceOf", abi.NewTypes(abi.Address, abi.Uint256), abi.NewTypes(abi.Uint256))
	allowanceFn    = abi.NewFunction("allowance", abi.NewTypes(abi.Address, abi.Address, abi.Uint256), abi.NewTypes(abi.Uint256))
	isOperatorFn   = abi.NewFunction("isOperator", abi.NewTypes(abi.Address, abi.Address), abi.NewTypes(abi.Bool))
	approveFn      = abi.NewFunction("approve", abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256), abi.NewTypes(abi.Bool))
	setOperatorFn  = abi.NewFunction("setOperator", abi.NewTypes(abi.Address, abi.Bool), abi.NewTypes(abi.Bool))
	transferFn     = abi.NewFunction("transfer", abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256), abi.NewTypes(abi.Bool))
	transferFromFn = abi.NewFunction("transferFrom", abi.NewTypes(abi.Address, abi.Address, abi.Uint256, abi.Uint256), abi.NewTypes(abi.Bool))
	nameFn         = abi.NewFunction("name", abi.NewTypes(abi.Uint256), abi.NewTypes(abi.String))
	symbolFn       = abi.NewFunction("symbol", abi.NewTypes(abi.Uint256), abi.NewTypes(abi.String))
	decimalsFn     = abi.NewFunction("decimals", abi.NewTypes(abi.Uint256), abi.NewTypes(abi.Uint8))
)

// Custom errors from OpenZeppelin's ERC6909, for decoding reverts via abi.Error.Decode.
var (
	errInsufficientBalance   = abi.NewError("ERC6909InsufficientBalance", abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256, abi.Uint256))
	errInsufficientAllowance = abi.NewError("ERC6909InsufficientAllowance", abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256, abi.Uint256))
	errInvalidApprover       = abi.NewError("ERC6909InvalidApprover", abi.NewTypes(abi.Address))
	errInvalidReceiver       = abi.NewError("ERC6909InvalidReceiver", abi.NewTypes(abi.Address))
	errInvalidSender         = abi.NewError("ERC6909InvalidSender", abi.NewTypes(abi.Address))
	errInvalidSpender        = abi.NewError("ERC6909InvalidSpender", abi.NewTypes(abi.Address))
)

// ErrInsufficientBalance, etc. mirror OpenZeppelin's ERC6909 custom errors, for errors.Is.
var (
	ErrInsufficientBalance   = errors.New("insufficient balance")
	ErrInsufficientAllowance = errors.New("insufficient allowance")
	ErrInvalidApprover       = errors.New("invalid approver")
	ErrInvalidReceiver       = errors.New("invalid receiver")
	ErrInvalidSender         = errors.New("invalid sender")
	ErrInvalidSpender        = errors.New("invalid spender")
	ErrUnknown               = errors.New("unknown revert")
)

// Events from IERC6909, for decoding logs via Event.Decode.
var (
	operatorSetEvent = abi.NewEvent("OperatorSet", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Bool},
	}, false)
	approvalEvent = abi.NewEvent("Approval", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Uint256, Indexed: true},
		{Type: abi.Uint256},
	}, false)
	transferEvent = abi.NewEvent("Transfer", []abi.EventParam{
		{Type: abi.Address},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Uint256, Indexed: true},
		{Type: abi.Uint256},
	}, false)
)

// Events from ERC6909Metadata (OpenZeppelin's implementation, not part of the IERC6909Metadata
// interface itself), for decoding logs via Event.Decode.
var (
	nameUpdatedEvent = abi.NewEvent("ERC6909NameUpdated", []abi.EventParam{
		{Type: abi.Uint256, Indexed: true},
		{Type: abi.String},
	}, false)
	symbolUpdatedEvent = abi.NewEvent("ERC6909SymbolUpdated", []abi.EventParam{
		{Type: abi.Uint256, Indexed: true},
		{Type: abi.String},
	}, false)
	decimalsUpdatedEvent = abi.NewEvent("ERC6909DecimalsUpdated", []abi.EventParam{
		{Type: abi.Uint256, Indexed: true},
		{Type: abi.Uint8},
	}, false)
)

// EncodeBalanceOf returns the calldata for balanceOf(owner, id).
func EncodeBalanceOf(owner *types.Address, id *big.Int) ([]byte, error) {
	data, err := balanceOfFn.EncodeCall(owner, id)
	if err != nil {
		return nil, fmt.Errorf("erc6909: encodeBalanceOf: %w", err)
	}
	return data, nil
}

// DecodeBalanceOf decodes the return data of balanceOf.
func DecodeBalanceOf(data []byte) (*big.Int, error) {
	n, err := balanceOfFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("erc6909: decodeBalanceOf: %w", err)
	}
	return n, nil
}

// EncodeAllowance returns the calldata for allowance(owner, spender, id).
func EncodeAllowance(owner, spender *types.Address, id *big.Int) ([]byte, error) {
	data, err := allowanceFn.EncodeCall(owner, spender, id)
	if err != nil {
		return nil, fmt.Errorf("erc6909: encodeAllowance: %w", err)
	}
	return data, nil
}

// DecodeAllowance decodes the return data of allowance.
func DecodeAllowance(data []byte) (*big.Int, error) {
	n, err := allowanceFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("erc6909: decodeAllowance: %w", err)
	}
	return n, nil
}

// EncodeIsOperator returns the calldata for isOperator(owner, spender).
func EncodeIsOperator(owner, spender *types.Address) ([]byte, error) {
	data, err := isOperatorFn.EncodeCall(owner, spender)
	if err != nil {
		return nil, fmt.Errorf("erc6909: encodeIsOperator: %w", err)
	}
	return data, nil
}

// DecodeIsOperator decodes the return data of isOperator.
func DecodeIsOperator(data []byte) (bool, error) {
	ok, err := isOperatorFn.DecodeSingleReturn[bool](data)
	if err != nil {
		return false, fmt.Errorf("erc6909: decodeIsOperator: %w", err)
	}
	return ok, nil
}

// EncodeApprove returns the calldata for approve(spender, id, amount).
func EncodeApprove(spender *types.Address, id, amount *big.Int) ([]byte, error) {
	data, err := approveFn.EncodeCall(spender, id, amount)
	if err != nil {
		return nil, fmt.Errorf("erc6909: encodeApprove: %w", err)
	}
	return data, nil
}

// DecodeApprove decodes the return data of approve.
func DecodeApprove(data []byte) (bool, error) {
	ok, err := approveFn.DecodeSingleReturn[bool](data)
	if err != nil {
		return false, fmt.Errorf("erc6909: decodeApprove: %w", err)
	}
	return ok, nil
}

// EncodeSetOperator returns the calldata for setOperator(spender, approved).
func EncodeSetOperator(spender *types.Address, approved bool) ([]byte, error) {
	data, err := setOperatorFn.EncodeCall(spender, approved)
	if err != nil {
		return nil, fmt.Errorf("erc6909: encodeSetOperator: %w", err)
	}
	return data, nil
}

// DecodeSetOperator decodes the return data of setOperator.
func DecodeSetOperator(data []byte) (bool, error) {
	ok, err := setOperatorFn.DecodeSingleReturn[bool](data)
	if err != nil {
		return false, fmt.Errorf("erc6909: decodeSetOperator: %w", err)
	}
	return ok, nil
}

// EncodeTransfer returns the calldata for transfer(receiver, id, amount).
func EncodeTransfer(receiver *types.Address, id, amount *big.Int) ([]byte, error) {
	data, err := transferFn.EncodeCall(receiver, id, amount)
	if err != nil {
		return nil, fmt.Errorf("erc6909: encodeTransfer: %w", err)
	}
	return data, nil
}

// DecodeTransfer decodes the return data of transfer.
func DecodeTransfer(data []byte) (bool, error) {
	ok, err := transferFn.DecodeSingleReturn[bool](data)
	if err != nil {
		return false, fmt.Errorf("erc6909: decodeTransfer: %w", err)
	}
	return ok, nil
}

// EncodeTransferFrom returns the calldata for transferFrom(sender, receiver, id, amount).
func EncodeTransferFrom(sender, receiver *types.Address, id, amount *big.Int) ([]byte, error) {
	data, err := transferFromFn.EncodeCall(sender, receiver, id, amount)
	if err != nil {
		return nil, fmt.Errorf("erc6909: encodeTransferFrom: %w", err)
	}
	return data, nil
}

// DecodeTransferFrom decodes the return data of transferFrom.
func DecodeTransferFrom(data []byte) (bool, error) {
	ok, err := transferFromFn.DecodeSingleReturn[bool](data)
	if err != nil {
		return false, fmt.Errorf("erc6909: decodeTransferFrom: %w", err)
	}
	return ok, nil
}

// EncodeName returns the calldata for name(id).
func EncodeName(id *big.Int) ([]byte, error) {
	data, err := nameFn.EncodeCall(id)
	if err != nil {
		return nil, fmt.Errorf("erc6909: encodeName: %w", err)
	}
	return data, nil
}

// DecodeName decodes the return data of name.
func DecodeName(data []byte) (string, error) {
	s, err := nameFn.DecodeSingleReturn[string](data)
	if err != nil {
		return "", fmt.Errorf("erc6909: decodeName: %w", err)
	}
	return s, nil
}

// EncodeSymbol returns the calldata for symbol(id).
func EncodeSymbol(id *big.Int) ([]byte, error) {
	data, err := symbolFn.EncodeCall(id)
	if err != nil {
		return nil, fmt.Errorf("erc6909: encodeSymbol: %w", err)
	}
	return data, nil
}

// DecodeSymbol decodes the return data of symbol.
func DecodeSymbol(data []byte) (string, error) {
	s, err := symbolFn.DecodeSingleReturn[string](data)
	if err != nil {
		return "", fmt.Errorf("erc6909: decodeSymbol: %w", err)
	}
	return s, nil
}

// EncodeDecimals returns the calldata for decimals(id).
func EncodeDecimals(id *big.Int) ([]byte, error) {
	data, err := decimalsFn.EncodeCall(id)
	if err != nil {
		return nil, fmt.Errorf("erc6909: encodeDecimals: %w", err)
	}
	return data, nil
}

// DecodeDecimals decodes the return data of decimals.
func DecodeDecimals(data []byte) (uint8, error) {
	d, err := decimalsFn.DecodeSingleReturn[uint8](data)
	if err != nil {
		return 0, fmt.Errorf("erc6909: decodeDecimals: %w", err)
	}
	return d, nil
}
