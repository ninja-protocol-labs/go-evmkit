// Package erc6909 is a client for ERC-6909 multi-token contracts, mirroring IERC6909.sol and
// IERC6909Metadata.sol.
package erc6909

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc165"
)

// IERC6909 mirrors IERC6909.sol, including its IERC165 embedding.
type IERC6909 interface {
	erc165.IERC165

	BalanceOf(ctx context.Context, opts *bind.CallOptions, req *BalanceOfRequest) (*big.Int, error)
	Allowance(ctx context.Context, opts *bind.CallOptions, req *AllowanceRequest) (*big.Int, error)
	IsOperator(ctx context.Context, opts *bind.CallOptions, req *IsOperatorRequest) (bool, error)
	CallApprove(ctx context.Context, opts *bind.CallOptions, req *ApproveRequest) (bool, error)
	CallSetOperator(ctx context.Context, opts *bind.CallOptions, req *SetOperatorRequest) (bool, error)
	CallTransfer(ctx context.Context, opts *bind.CallOptions, req *TransferRequest) (bool, error)
	CallTransferFrom(ctx context.Context, opts *bind.CallOptions, req *TransferFromRequest) (bool, error)

	Approve(ctx context.Context, opts *bind.TransactOptions, req *ApproveRequest) (core.Transaction, *types.Hash, error)
	SetOperator(ctx context.Context, opts *bind.TransactOptions, req *SetOperatorRequest) (core.Transaction, *types.Hash, error)
	Transfer(ctx context.Context, opts *bind.TransactOptions, req *TransferRequest) (core.Transaction, *types.Hash, error)
	TransferFrom(ctx context.Context, opts *bind.TransactOptions, req *TransferFromRequest) (core.Transaction, *types.Hash, error)
}

// IERC6909Metadata mirrors IERC6909Metadata.sol, including its IERC6909 embedding.
type IERC6909Metadata interface {
	IERC6909

	Name(ctx context.Context, opts *bind.CallOptions, req *NameRequest) (string, error)
	Symbol(ctx context.Context, opts *bind.CallOptions, req *SymbolRequest) (string, error)
	Decimals(ctx context.Context, opts *bind.CallOptions, req *DecimalsRequest) (uint8, error)
}

var (
	_ IERC6909         = (*ERC6909)(nil)
	_ IERC6909Metadata = (*ERC6909)(nil)
)

// ERC6909 calls any ERC-6909 contract over a Client; SupportsInterface is promoted from the
// embedded *erc165.ERC165, mirroring IERC6909 is IERC165.
type ERC6909 struct {
	*erc165.ERC165

	c rpc.Client
}

// New returns an ERC6909 calling contracts via cli.
func New(cli rpc.Client) *ERC6909 {
	return &ERC6909{
		c:      cli,
		ERC165: erc165.New(cli),
	}
}

// BalanceOf returns req.Owner's balance of req.ID on req.Contract.
func (e *ERC6909) BalanceOf(ctx context.Context, opts *bind.CallOptions, req *BalanceOfRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc6909: balanceOf: %w", err)
	}

	calldata, err := EncodeBalanceOf(req.Owner, req.ID)
	if err != nil {
		return nil, fmt.Errorf("erc6909: balanceOf: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc6909: balanceOf: %w", MapRevertError(err))
	}

	out, err := DecodeBalanceOf(data)
	if err != nil {
		return nil, fmt.Errorf("erc6909: balanceOf: %w", err)
	}

	return out, nil
}

// Allowance returns how much of req.ID req.Spender may still spend of req.Owner's balance.
func (e *ERC6909) Allowance(ctx context.Context, opts *bind.CallOptions, req *AllowanceRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc6909: allowance: %w", err)
	}

	calldata, err := EncodeAllowance(req.Owner, req.Spender, req.ID)
	if err != nil {
		return nil, fmt.Errorf("erc6909: allowance: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc6909: allowance: %w", MapRevertError(err))
	}

	out, err := DecodeAllowance(data)
	if err != nil {
		return nil, fmt.Errorf("erc6909: allowance: %w", err)
	}

	return out, nil
}

// IsOperator reports whether req.Spender is an operator for all of req.Owner's ids.
func (e *ERC6909) IsOperator(ctx context.Context, opts *bind.CallOptions, req *IsOperatorRequest) (bool, error) {
	if err := req.Validate(); err != nil {
		return false, fmt.Errorf("erc6909: isOperator: %w", err)
	}

	calldata, err := EncodeIsOperator(req.Owner, req.Spender)
	if err != nil {
		return false, fmt.Errorf("erc6909: isOperator: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return false, fmt.Errorf("erc6909: isOperator: %w", MapRevertError(err))
	}

	out, err := DecodeIsOperator(data)
	if err != nil {
		return false, fmt.Errorf("erc6909: isOperator: %w", err)
	}

	return out, nil
}

// CallApprove runs approve(spender, id, amount) via eth_call as opts.From, returning its result or the revert.
func (e *ERC6909) CallApprove(ctx context.Context, opts *bind.CallOptions, req *ApproveRequest) (bool, error) {
	if err := req.Validate(); err != nil {
		return false, fmt.Errorf("erc6909: callApprove: %w", err)
	}

	calldata, err := EncodeApprove(req.Spender, req.ID, req.Amount)
	if err != nil {
		return false, fmt.Errorf("erc6909: callApprove: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return false, fmt.Errorf("erc6909: callApprove: %w", MapRevertError(err))
	}

	out, err := DecodeApprove(data)
	if err != nil {
		return false, fmt.Errorf("erc6909: callApprove: %w", err)
	}

	return out, nil
}

// CallSetOperator runs setOperator(spender, approved) via eth_call as opts.From, returning its result or the revert.
func (e *ERC6909) CallSetOperator(ctx context.Context, opts *bind.CallOptions, req *SetOperatorRequest) (bool, error) {
	if err := req.Validate(); err != nil {
		return false, fmt.Errorf("erc6909: callSetOperator: %w", err)
	}

	calldata, err := EncodeSetOperator(req.Spender, req.Approved)
	if err != nil {
		return false, fmt.Errorf("erc6909: callSetOperator: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return false, fmt.Errorf("erc6909: callSetOperator: %w", MapRevertError(err))
	}

	out, err := DecodeSetOperator(data)
	if err != nil {
		return false, fmt.Errorf("erc6909: callSetOperator: %w", err)
	}

	return out, nil
}

// CallTransfer runs transfer(receiver, id, amount) via eth_call as opts.From, returning its result or the revert.
func (e *ERC6909) CallTransfer(ctx context.Context, opts *bind.CallOptions, req *TransferRequest) (bool, error) {
	if err := req.Validate(); err != nil {
		return false, fmt.Errorf("erc6909: callTransfer: %w", err)
	}

	calldata, err := EncodeTransfer(req.Receiver, req.ID, req.Amount)
	if err != nil {
		return false, fmt.Errorf("erc6909: callTransfer: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return false, fmt.Errorf("erc6909: callTransfer: %w", MapRevertError(err))
	}

	out, err := DecodeTransfer(data)
	if err != nil {
		return false, fmt.Errorf("erc6909: callTransfer: %w", err)
	}

	return out, nil
}

// CallTransferFrom runs transferFrom(sender, receiver, id, amount) via eth_call as opts.From, returning its result or the revert.
func (e *ERC6909) CallTransferFrom(ctx context.Context, opts *bind.CallOptions, req *TransferFromRequest) (bool, error) {
	if err := req.Validate(); err != nil {
		return false, fmt.Errorf("erc6909: callTransferFrom: %w", err)
	}

	calldata, err := EncodeTransferFrom(req.Sender, req.Receiver, req.ID, req.Amount)
	if err != nil {
		return false, fmt.Errorf("erc6909: callTransferFrom: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return false, fmt.Errorf("erc6909: callTransferFrom: %w", MapRevertError(err))
	}

	out, err := DecodeTransferFrom(data)
	if err != nil {
		return false, fmt.Errorf("erc6909: callTransferFrom: %w", err)
	}

	return out, nil
}

// Approve sends approve(spender, id, amount) on req.Contract, signed by opts.Signer.
func (e *ERC6909) Approve(ctx context.Context, opts *bind.TransactOptions, req *ApproveRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("erc6909: approve: %w", err)
	}

	calldata, err := EncodeApprove(req.Spender, req.ID, req.Amount)
	if err != nil {
		return nil, nil, fmt.Errorf("erc6909: approve: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("erc6909: approve: %w", MapRevertError(err))
	}

	return signed, hash, nil
}

// SetOperator sends setOperator(spender, approved) on req.Contract, signed by opts.Signer.
func (e *ERC6909) SetOperator(ctx context.Context, opts *bind.TransactOptions, req *SetOperatorRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("erc6909: setOperator: %w", err)
	}

	calldata, err := EncodeSetOperator(req.Spender, req.Approved)
	if err != nil {
		return nil, nil, fmt.Errorf("erc6909: setOperator: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("erc6909: setOperator: %w", MapRevertError(err))
	}

	return signed, hash, nil
}

// Transfer sends transfer(receiver, id, amount) on req.Contract, signed by opts.Signer.
func (e *ERC6909) Transfer(ctx context.Context, opts *bind.TransactOptions, req *TransferRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("erc6909: transfer: %w", err)
	}

	calldata, err := EncodeTransfer(req.Receiver, req.ID, req.Amount)
	if err != nil {
		return nil, nil, fmt.Errorf("erc6909: transfer: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("erc6909: transfer: %w", MapRevertError(err))
	}

	return signed, hash, nil
}

// TransferFrom sends transferFrom(sender, receiver, id, amount) on req.Contract, signed by opts.Signer as the spender.
func (e *ERC6909) TransferFrom(ctx context.Context, opts *bind.TransactOptions, req *TransferFromRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("erc6909: transferFrom: %w", err)
	}

	calldata, err := EncodeTransferFrom(req.Sender, req.Receiver, req.ID, req.Amount)
	if err != nil {
		return nil, nil, fmt.Errorf("erc6909: transferFrom: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("erc6909: transferFrom: %w", MapRevertError(err))
	}

	return signed, hash, nil
}

// Name returns req.ID's name on req.Contract.
func (e *ERC6909) Name(ctx context.Context, opts *bind.CallOptions, req *NameRequest) (string, error) {
	if err := req.Validate(); err != nil {
		return "", fmt.Errorf("erc6909: name: %w", err)
	}

	calldata, err := EncodeName(req.ID)
	if err != nil {
		return "", fmt.Errorf("erc6909: name: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return "", fmt.Errorf("erc6909: name: %w", MapRevertError(err))
	}

	out, err := DecodeName(data)
	if err != nil {
		return "", fmt.Errorf("erc6909: name: %w", err)
	}

	return out, nil
}

// Symbol returns req.ID's symbol on req.Contract.
func (e *ERC6909) Symbol(ctx context.Context, opts *bind.CallOptions, req *SymbolRequest) (string, error) {
	if err := req.Validate(); err != nil {
		return "", fmt.Errorf("erc6909: symbol: %w", err)
	}

	calldata, err := EncodeSymbol(req.ID)
	if err != nil {
		return "", fmt.Errorf("erc6909: symbol: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return "", fmt.Errorf("erc6909: symbol: %w", MapRevertError(err))
	}

	out, err := DecodeSymbol(data)
	if err != nil {
		return "", fmt.Errorf("erc6909: symbol: %w", err)
	}

	return out, nil
}

// Decimals returns req.ID's decimals on req.Contract.
func (e *ERC6909) Decimals(ctx context.Context, opts *bind.CallOptions, req *DecimalsRequest) (uint8, error) {
	if err := req.Validate(); err != nil {
		return 0, fmt.Errorf("erc6909: decimals: %w", err)
	}

	calldata, err := EncodeDecimals(req.ID)
	if err != nil {
		return 0, fmt.Errorf("erc6909: decimals: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return 0, fmt.Errorf("erc6909: decimals: %w", MapRevertError(err))
	}

	out, err := DecodeDecimals(data)
	if err != nil {
		return 0, fmt.Errorf("erc6909: decimals: %w", err)
	}

	return out, nil
}
