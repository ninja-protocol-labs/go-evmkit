// Package erc1155 is a client for ERC-1155 tokens, mirroring IERC1155.sol and
// IERC1155MetadataURI.sol.
package erc1155

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

// IERC1155 mirrors IERC1155.sol, including its IERC165 embedding.
type IERC1155 interface {
	erc165.IERC165

	BalanceOf(ctx context.Context, opts *bind.CallOptions, req *BalanceOfRequest) (*big.Int, error)
	BalanceOfBatch(ctx context.Context, opts *bind.CallOptions, req *BalanceOfBatchRequest) ([]*big.Int, error)
	IsApprovedForAll(ctx context.Context, opts *bind.CallOptions, req *IsApprovedForAllRequest) (bool, error)

	SetApprovalForAll(ctx context.Context, opts *bind.TransactOptions, req *SetApprovalForAllRequest) (core.Transaction, *types.Hash, error)
	SafeTransferFrom(ctx context.Context, opts *bind.TransactOptions, req *SafeTransferFromRequest) (core.Transaction, *types.Hash, error)
	SafeBatchTransferFrom(ctx context.Context, opts *bind.TransactOptions, req *SafeBatchTransferFromRequest) (core.Transaction, *types.Hash, error)
}

// IERC1155MetadataURI mirrors IERC1155MetadataURI.sol, including its IERC1155 embedding.
type IERC1155MetadataURI interface {
	IERC1155

	URI(ctx context.Context, opts *bind.CallOptions, req *URIRequest) (string, error)
}

// IERC1155Batcher groups the batched balance and approval reads; it mirrors no
// Solidity interface.
type IERC1155Batcher interface {
	IdBalances(ctx context.Context, opts *bind.CallOptions, req *IdBalancesRequest) ([]IdBalance, error)
	IdBalancePairs(ctx context.Context, opts *bind.CallOptions, req *IdBalancePairsRequest) ([]IdBalancePair, error)
	AccountIdBalances(ctx context.Context, opts *bind.CallOptions, req *AccountIdBalancesRequest) (*AccountIdBalances, error)
	ApprovalForAllWithBalance(ctx context.Context, opts *bind.CallOptions, req *ApprovalForAllWithBalanceRequest) (*ApprovalForAllWithBalance, error)
}

var (
	_ IERC1155            = (*ERC1155)(nil)
	_ IERC1155MetadataURI = (*ERC1155)(nil)
	_ IERC1155Batcher     = (*ERC1155)(nil)
)

// ERC1155 calls any ERC-1155 token contract over a Client; SupportsInterface is
// promoted from the embedded *erc165.ERC165, mirroring IERC1155 is IERC165.
type ERC1155 struct {
	*erc165.ERC165

	c rpc.Client
}

// New returns an ERC1155 calling contracts via cli.
func New(cli rpc.Client) *ERC1155 {
	return &ERC1155{
		c:      cli,
		ERC165: erc165.New(cli),
	}
}

// URI returns req.ID's URI on req.Contract.
func (e *ERC1155) URI(ctx context.Context, opts *bind.CallOptions, req *URIRequest) (string, error) {
	if err := req.Validate(); err != nil {
		return "", fmt.Errorf("erc1155: uri: %w", err)
	}

	calldata, err := EncodeURI(req.ID)
	if err != nil {
		return "", fmt.Errorf("erc1155: uri: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return "", fmt.Errorf("erc1155: uri: %w", MapRevertError(err))
	}

	out, err := DecodeURI(data)
	if err != nil {
		return "", fmt.Errorf("erc1155: uri: %w", err)
	}

	return out, nil
}

// BalanceOf returns req.Account's balance of req.ID on req.Contract.
func (e *ERC1155) BalanceOf(ctx context.Context, opts *bind.CallOptions, req *BalanceOfRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc1155: balanceOf: %w", err)
	}

	calldata, err := EncodeBalanceOf(req.Account, req.ID)
	if err != nil {
		return nil, fmt.Errorf("erc1155: balanceOf: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc1155: balanceOf: %w", MapRevertError(err))
	}

	out, err := DecodeBalanceOf(data)
	if err != nil {
		return nil, fmt.Errorf("erc1155: balanceOf: %w", err)
	}

	return out, nil
}

// BalanceOfBatch returns each req.Accounts[i]'s balance of req.IDs[i], in the same order.
func (e *ERC1155) BalanceOfBatch(ctx context.Context, opts *bind.CallOptions, req *BalanceOfBatchRequest) ([]*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc1155: balanceOfBatch: %w", err)
	}

	calldata, err := EncodeBalanceOfBatch(req.Accounts, req.IDs)
	if err != nil {
		return nil, fmt.Errorf("erc1155: balanceOfBatch: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc1155: balanceOfBatch: %w", MapRevertError(err))
	}

	out, err := DecodeBalanceOfBatch(data)
	if err != nil {
		return nil, fmt.Errorf("erc1155: balanceOfBatch: %w", err)
	}

	return out, nil
}

// IsApprovedForAll reports whether req.Operator may manage all of req.Account's tokens.
func (e *ERC1155) IsApprovedForAll(ctx context.Context, opts *bind.CallOptions, req *IsApprovedForAllRequest) (bool, error) {
	if err := req.Validate(); err != nil {
		return false, fmt.Errorf("erc1155: isApprovedForAll: %w", err)
	}

	calldata, err := EncodeIsApprovedForAll(req.Account, req.Operator)
	if err != nil {
		return false, fmt.Errorf("erc1155: isApprovedForAll: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return false, fmt.Errorf("erc1155: isApprovedForAll: %w", MapRevertError(err))
	}

	out, err := DecodeIsApprovedForAll(data)
	if err != nil {
		return false, fmt.Errorf("erc1155: isApprovedForAll: %w", err)
	}

	return out, nil
}

// SetApprovalForAll sends setApprovalForAll(operator, approved) on req.Contract, signed by opts.Signer.
func (e *ERC1155) SetApprovalForAll(ctx context.Context, opts *bind.TransactOptions, req *SetApprovalForAllRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("erc1155: setApprovalForAll: %w", err)
	}

	calldata, err := EncodeSetApprovalForAll(req.Operator, req.Approved)
	if err != nil {
		return nil, nil, fmt.Errorf("erc1155: setApprovalForAll: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("erc1155: setApprovalForAll: %w", MapRevertError(err))
	}

	return signed, hash, nil
}

// SafeTransferFrom sends safeTransferFrom(from, to, id, value, data) on req.Contract, signed by opts.Signer.
func (e *ERC1155) SafeTransferFrom(ctx context.Context, opts *bind.TransactOptions, req *SafeTransferFromRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("erc1155: safeTransferFrom: %w", err)
	}

	calldata, err := EncodeSafeTransferFrom(req.From, req.To, req.ID, req.Amount, req.Data)
	if err != nil {
		return nil, nil, fmt.Errorf("erc1155: safeTransferFrom: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("erc1155: safeTransferFrom: %w", MapRevertError(err))
	}

	return signed, hash, nil
}

// SafeBatchTransferFrom sends safeBatchTransferFrom(from, to, ids, values, data) on req.Contract, signed by opts.Signer.
func (e *ERC1155) SafeBatchTransferFrom(ctx context.Context, opts *bind.TransactOptions, req *SafeBatchTransferFromRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("erc1155: safeBatchTransferFrom: %w", err)
	}

	calldata, err := EncodeSafeBatchTransferFrom(req.From, req.To, req.IDs, req.Amounts, req.Data)
	if err != nil {
		return nil, nil, fmt.Errorf("erc1155: safeBatchTransferFrom: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("erc1155: safeBatchTransferFrom: %w", MapRevertError(err))
	}

	return signed, hash, nil
}

// IdBalances returns req.Account's balance of each of req.IDs via balanceOfBatch.
func (e *ERC1155) IdBalances(ctx context.Context, opts *bind.CallOptions, req *IdBalancesRequest) ([]IdBalance, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc1155: idBalances: %w", err)
	}

	accounts := make([]*types.Address, len(req.IDs))
	for i := range accounts {
		accounts[i] = req.Account
	}

	calldata, err := EncodeBalanceOfBatch(accounts, req.IDs)
	if err != nil {
		return nil, fmt.Errorf("erc1155: idBalances: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc1155: idBalances: %w", MapRevertError(err))
	}

	out, err := DecodeIdBalances(req.IDs, data)
	if err != nil {
		return nil, fmt.Errorf("erc1155: idBalances: %w", err)
	}

	return out, nil
}

// IdBalancePairs returns each req.Accounts[i]'s balance of req.IDs[i] via balanceOfBatch.
func (e *ERC1155) IdBalancePairs(ctx context.Context, opts *bind.CallOptions, req *IdBalancePairsRequest) ([]IdBalancePair, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc1155: idBalancePairs: %w", err)
	}

	calldata, err := EncodeBalanceOfBatch(req.Accounts, req.IDs)
	if err != nil {
		return nil, fmt.Errorf("erc1155: idBalancePairs: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc1155: idBalancePairs: %w", MapRevertError(err))
	}

	out, err := DecodeIdBalancePairs(req.Accounts, req.IDs, data)
	if err != nil {
		return nil, fmt.Errorf("erc1155: idBalancePairs: %w", err)
	}

	return out, nil
}

// AccountIdBalances batches req.Account's balance of req.ID on each of req.Contracts via Multicall3.
func (e *ERC1155) AccountIdBalances(ctx context.Context, opts *bind.CallOptions, req *AccountIdBalancesRequest) (*AccountIdBalances, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc1155: accountIdBalances: %w", err)
	}

	calldata, err := EncodeAccountIdBalances(req.Account, req.ID, req.Contracts, req.AllowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc1155: accountIdBalances: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Multicall, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc1155: accountIdBalances: %w", MapRevertError(err))
	}

	out, err := DecodeAccountIdBalances(req.Account, req.ID, req.Contracts, data)
	if err != nil {
		return nil, fmt.Errorf("erc1155: accountIdBalances: %w", err)
	}

	return out, nil
}

// ApprovalForAllWithBalance batches isApprovedForAll(account, operator) and both balances of req.ID via Multicall3.
func (e *ERC1155) ApprovalForAllWithBalance(ctx context.Context, opts *bind.CallOptions, req *ApprovalForAllWithBalanceRequest) (*ApprovalForAllWithBalance, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc1155: approvalForAllWithBalance: %w", err)
	}

	calldata, err := EncodeApprovalForAllWithBalance(req.Contract, req.Account, req.Operator, req.ID, req.AllowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc1155: approvalForAllWithBalance: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Multicall, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc1155: approvalForAllWithBalance: %w", MapRevertError(err))
	}

	out, err := DecodeApprovalForAllWithBalance(req.Account, req.Operator, req.ID, data)
	if err != nil {
		return nil, fmt.Errorf("erc1155: approvalForAllWithBalance: %w", err)
	}

	return out, nil
}
