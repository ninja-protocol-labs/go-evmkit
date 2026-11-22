// Package erc20 is a client for ERC-20 tokens, mirroring IERC20.sol and
// IERC20Metadata.sol.
package erc20

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// IERC20 is implemented by *ERC20.
type IERC20 interface {
	TotalSupply(ctx context.Context, opts *bind.CallOptions, req *TotalSupplyRequest) (*big.Int, error)
	BalanceOf(ctx context.Context, opts *bind.CallOptions, req *BalanceOfRequest) (*big.Int, error)
	Allowance(ctx context.Context, opts *bind.CallOptions, req *AllowanceRequest) (*big.Int, error)
	CallTransfer(ctx context.Context, opts *bind.CallOptions, req *TransferRequest) (bool, error)
	CallApprove(ctx context.Context, opts *bind.CallOptions, req *ApproveRequest) (bool, error)
	CallTransferFrom(ctx context.Context, opts *bind.CallOptions, req *TransferFromRequest) (bool, error)

	Transfer(ctx context.Context, opts *bind.TransactOptions, req *TransferRequest) (core.Transaction, *types.Hash, error)
	Approve(ctx context.Context, opts *bind.TransactOptions, req *ApproveRequest) (core.Transaction, *types.Hash, error)
	TransferFrom(ctx context.Context, opts *bind.TransactOptions, req *TransferFromRequest) (core.Transaction, *types.Hash, error)
}

// IERC20Metadata mirrors IERC20Metadata.sol, including its IERC20 embedding.
// name/symbol/decimals are optional per EIP-20.
type IERC20Metadata interface {
	IERC20

	Name(ctx context.Context, opts *bind.CallOptions, req *NameRequest) (string, error)
	Symbol(ctx context.Context, opts *bind.CallOptions, req *SymbolRequest) (string, error)
	Decimals(ctx context.Context, opts *bind.CallOptions, req *DecimalsRequest) (uint8, error)
}

// IERC20Batcher groups the Multicall3-backed batch reads; it mirrors no
// Solidity interface.
type IERC20Batcher interface {
	Metadata(ctx context.Context, opts *bind.CallOptions, req *MetadataRequest) (*Metadata, error)
	MetadataWithBalance(ctx context.Context, opts *bind.CallOptions, req *MetadataWithBalanceRequest) (*MetadataWithBalance, error)
	AllowanceWithBalance(ctx context.Context, opts *bind.CallOptions, req *AllowanceWithBalanceRequest) (*AllowanceWithBalance, error)
	TokenBalances(ctx context.Context, opts *bind.CallOptions, req *TokenBalancesRequest) (*TokenBalances, error)
	AddressBalances(ctx context.Context, opts *bind.CallOptions, req *AddressBalancesRequest) (*AddressBalances, error)
	BalancePairs(ctx context.Context, opts *bind.CallOptions, req *BalancePairsRequest) ([]TokenAddressBalance, error)
}

var (
	_ IERC20         = (*ERC20)(nil)
	_ IERC20Metadata = (*ERC20)(nil)
	_ IERC20Batcher  = (*ERC20)(nil)
)

// ERC20 calls any ERC-20 token contract over a Client; it holds no token
// address, so one instance serves every token.
type ERC20 struct {
	c rpc.Client
}

// New returns an ERC20 calling tokens via cli.
func New(cli rpc.Client) *ERC20 {
	return &ERC20{
		c: cli,
	}
}

// TotalSupply returns req.Contract's total supply.
func (e *ERC20) TotalSupply(ctx context.Context, opts *bind.CallOptions, req *TotalSupplyRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc20: totalSupply: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, EncodeTotalSupply())
	if err != nil {
		return nil, fmt.Errorf("erc20: totalSupply: %w", e.wrapRevertError(err))
	}

	n, err := DecodeTotalSupply(data)
	if err != nil {
		return nil, fmt.Errorf("erc20: totalSupply: %w", err)
	}

	return n, nil
}

// BalanceOf returns req.Account's balance of req.Contract.
func (e *ERC20) BalanceOf(ctx context.Context, opts *bind.CallOptions, req *BalanceOfRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc20: balanceOf: %w", err)
	}

	calldata, err := EncodeBalanceOf(req.Account)
	if err != nil {
		return nil, fmt.Errorf("erc20: balanceOf: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc20: balanceOf: %w", e.wrapRevertError(err))
	}

	n, err := DecodeBalanceOf(data)
	if err != nil {
		return nil, fmt.Errorf("erc20: balanceOf: %w", err)
	}

	return n, nil
}

// Allowance returns how much req.Spender may still spend of req.Owner's req.Contract.
func (e *ERC20) Allowance(ctx context.Context, opts *bind.CallOptions, req *AllowanceRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc20: allowance: %w", err)
	}

	calldata, err := EncodeAllowance(req.Owner, req.Spender)
	if err != nil {
		return nil, fmt.Errorf("erc20: allowance: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc20: allowance: %w", e.wrapRevertError(err))
	}

	n, err := DecodeAllowance(data)
	if err != nil {
		return nil, fmt.Errorf("erc20: allowance: %w", err)
	}

	return n, nil
}

// CallTransfer runs transfer(to, value) via eth_call as opts.From, returning
// its result or the revert a real transfer would hit.
func (e *ERC20) CallTransfer(ctx context.Context, opts *bind.CallOptions, req *TransferRequest) (bool, error) {
	if err := req.Validate(); err != nil {
		return false, fmt.Errorf("erc20: callTransfer: %w", err)
	}

	calldata, err := EncodeTransfer(req.To, req.Amount)
	if err != nil {
		return false, fmt.Errorf("erc20: callTransfer: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return false, fmt.Errorf("erc20: callTransfer: %w", e.wrapRevertError(err))
	}

	out, err := DecodeTransfer(data)
	if err != nil {
		return false, fmt.Errorf("erc20: callTransfer: %w", err)
	}

	return out, nil
}

// CallApprove runs approve(spender, value) via eth_call as opts.From,
// returning its result or the revert a real approve would hit.
func (e *ERC20) CallApprove(ctx context.Context, opts *bind.CallOptions, req *ApproveRequest) (bool, error) {
	if err := req.Validate(); err != nil {
		return false, fmt.Errorf("erc20: callApprove: %w", err)
	}

	calldata, err := EncodeApprove(req.Spender, req.Amount)
	if err != nil {
		return false, fmt.Errorf("erc20: callApprove: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return false, fmt.Errorf("erc20: callApprove: %w", e.wrapRevertError(err))
	}

	out, err := DecodeApprove(data)
	if err != nil {
		return false, fmt.Errorf("erc20: callApprove: %w", err)
	}

	return out, nil
}

// CallTransferFrom runs transferFrom(from, to, value) via eth_call as opts.From
// (the spender), returning its result or the revert a real call would hit.
func (e *ERC20) CallTransferFrom(ctx context.Context, opts *bind.CallOptions, req *TransferFromRequest) (bool, error) {
	if err := req.Validate(); err != nil {
		return false, fmt.Errorf("erc20: callTransferFrom: %w", err)
	}

	calldata, err := EncodeTransferFrom(req.From, req.To, req.Amount)
	if err != nil {
		return false, fmt.Errorf("erc20: callTransferFrom: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return false, fmt.Errorf("erc20: callTransferFrom: %w", e.wrapRevertError(err))
	}

	out, err := DecodeTransferFrom(data)
	if err != nil {
		return false, fmt.Errorf("erc20: callTransferFrom: %w", err)
	}

	return out, nil
}

// Transfer sends transfer(to, value) on req.Contract, signed by opts.Signer.
func (e *ERC20) Transfer(ctx context.Context, opts *bind.TransactOptions, req *TransferRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("erc20: transfer: %w", err)
	}

	calldata, err := EncodeTransfer(req.To, req.Amount)
	if err != nil {
		return nil, nil, fmt.Errorf("erc20: transfer: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("erc20: transfer: %w", e.wrapRevertError(err))
	}

	return signed, hash, nil
}

// Approve sends approve(spender, value) on req.Contract, signed by opts.Signer.
func (e *ERC20) Approve(ctx context.Context, opts *bind.TransactOptions, req *ApproveRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("erc20: approve: %w", err)
	}

	calldata, err := EncodeApprove(req.Spender, req.Amount)
	if err != nil {
		return nil, nil, fmt.Errorf("erc20: approve: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("erc20: approve: %w", e.wrapRevertError(err))
	}

	return signed, hash, nil
}

// TransferFrom sends transferFrom(from, to, value) on req.Contract, signed by
// opts.Signer as the spender.
func (e *ERC20) TransferFrom(ctx context.Context, opts *bind.TransactOptions, req *TransferFromRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("erc20: transferFrom: %w", err)
	}

	calldata, err := EncodeTransferFrom(req.From, req.To, req.Amount)
	if err != nil {
		return nil, nil, fmt.Errorf("erc20: transferFrom: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("erc20: transferFrom: %w", e.wrapRevertError(err))
	}

	return signed, hash, nil
}

// Name returns req.Contract's name.
func (e *ERC20) Name(ctx context.Context, opts *bind.CallOptions, req *NameRequest) (string, error) {
	if err := req.Validate(); err != nil {
		return "", fmt.Errorf("erc20: name: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, EncodeName())
	if err != nil {
		return "", fmt.Errorf("erc20: name: %w", e.wrapRevertError(err))
	}

	out, err := DecodeName(data)
	if err != nil {
		return "", fmt.Errorf("erc20: name: %w", err)
	}

	return out, nil
}

// Symbol returns req.Contract's symbol.
func (e *ERC20) Symbol(ctx context.Context, opts *bind.CallOptions, req *SymbolRequest) (string, error) {
	if err := req.Validate(); err != nil {
		return "", fmt.Errorf("erc20: symbol: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, EncodeSymbol())
	if err != nil {
		return "", fmt.Errorf("erc20: symbol: %w", e.wrapRevertError(err))
	}

	out, err := DecodeSymbol(data)
	if err != nil {
		return "", fmt.Errorf("erc20: symbol: %w", err)
	}

	return out, nil
}

// Decimals returns req.Contract's decimals.
func (e *ERC20) Decimals(ctx context.Context, opts *bind.CallOptions, req *DecimalsRequest) (uint8, error) {
	if err := req.Validate(); err != nil {
		return 0, fmt.Errorf("erc20: decimals: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, EncodeDecimals())
	if err != nil {
		return 0, fmt.Errorf("erc20: decimals: %w", e.wrapRevertError(err))
	}

	out, err := DecodeDecimals(data)
	if err != nil {
		return 0, fmt.Errorf("erc20: decimals: %w", err)
	}

	return out, nil
}

// Metadata batches name/symbol/decimals/totalSupply on req.Token via Multicall3 at req.Multicall.
func (e *ERC20) Metadata(ctx context.Context, opts *bind.CallOptions, req *MetadataRequest) (*Metadata, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc20: metadata: %w", err)
	}

	calldata, err := EncodeMetadata(req.Token, req.AllowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc20: metadata: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Multicall, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc20: metadata: %w", e.wrapRevertError(err))
	}

	out, err := DecodeMetadata(data)
	if err != nil {
		return nil, fmt.Errorf("erc20: metadata: %w", err)
	}

	return out, nil
}

// MetadataWithBalance is Metadata plus req.Account's balance in the same call.
func (e *ERC20) MetadataWithBalance(ctx context.Context, opts *bind.CallOptions, req *MetadataWithBalanceRequest) (*MetadataWithBalance, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc20: metadataWithBalance: %w", err)
	}

	calldata, err := EncodeMetadataWithBalance(req.Token, req.Account, req.AllowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc20: metadataWithBalance: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Multicall, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc20: metadataWithBalance: %w", e.wrapRevertError(err))
	}

	out, err := DecodeMetadataWithBalance(data)
	if err != nil {
		return nil, fmt.Errorf("erc20: metadataWithBalance: %w", err)
	}

	return out, nil
}

// AllowanceWithBalance batches allowance(owner, spender) and both balances on req.Token.
func (e *ERC20) AllowanceWithBalance(ctx context.Context, opts *bind.CallOptions, req *AllowanceWithBalanceRequest) (*AllowanceWithBalance, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc20: allowanceWithBalance: %w", err)
	}

	calldata, err := EncodeAllowanceWithBalance(req.Token, req.Owner, req.Spender, req.AllowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc20: allowanceWithBalance: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Multicall, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc20: allowanceWithBalance: %w", e.wrapRevertError(err))
	}

	out, err := DecodeAllowanceWithBalance(data)
	if err != nil {
		return nil, fmt.Errorf("erc20: allowanceWithBalance: %w", err)
	}

	return out, nil
}

// TokenBalances batches balanceOf(account) on req.Token for each of req.Accounts.
func (e *ERC20) TokenBalances(ctx context.Context, opts *bind.CallOptions, req *TokenBalancesRequest) (*TokenBalances, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc20: tokenBalances: %w", err)
	}

	calldata, err := EncodeTokenBalances(req.Token, req.Accounts, req.AllowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc20: tokenBalances: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Multicall, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc20: tokenBalances: %w", e.wrapRevertError(err))
	}

	out, err := DecodeTokenBalances(req.Token, req.Accounts, data)
	if err != nil {
		return nil, fmt.Errorf("erc20: tokenBalances: %w", err)
	}

	return out, nil
}

// AddressBalances batches balanceOf(req.Account) on each of req.Tokens.
func (e *ERC20) AddressBalances(ctx context.Context, opts *bind.CallOptions, req *AddressBalancesRequest) (*AddressBalances, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc20: addressBalances: %w", err)
	}

	calldata, err := EncodeAddressBalances(req.Account, req.Tokens, req.AllowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc20: addressBalances: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Multicall, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc20: addressBalances: %w", e.wrapRevertError(err))
	}

	out, err := DecodeAddressBalances(req.Account, req.Tokens, data)
	if err != nil {
		return nil, fmt.Errorf("erc20: addressBalances: %w", err)
	}

	return out, nil
}

// BalancePairs batches balanceOf(req.Accounts[i]) on req.Tokens[i], paired by index.
func (e *ERC20) BalancePairs(ctx context.Context, opts *bind.CallOptions, req *BalancePairsRequest) ([]TokenAddressBalance, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc20: balancePairs: %w", err)
	}

	calldata, err := EncodeBalancePairs(req.Tokens, req.Accounts, req.AllowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc20: balancePairs: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Multicall, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc20: balancePairs: %w", e.wrapRevertError(err))
	}

	out, err := DecodeBalancePairs(req.Tokens, req.Accounts, data)
	if err != nil {
		return nil, fmt.Errorf("erc20: balancePairs: %w", err)
	}

	return out, nil
}

// wrapRevertError maps a known IERC20Errors custom error in err to its sentinel.
func (_ *ERC20) wrapRevertError(err error) error {
	var rpcErr *rpc.ResponseError
	if !errors.As(err, &rpcErr) || len(rpcErr.Data) == 0 {
		return err
	}

	var hexData string
	if jsonErr := json.Unmarshal(rpcErr.Data, &hexData); jsonErr != nil {
		return err
	}

	data, decErr := encoding.Hex.Decode(hexData)
	if decErr != nil || len(data) < 4 {
		return err
	}

	var sel abi.Selector
	copy(sel[:], data[:4])

	switch sel {
	case errInsufficientBalance.Selector():
		return fmt.Errorf("%w: %w", ErrInsufficientBalance, err)
	case errInvalidSender.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidSender, err)
	case errInvalidReceiver.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidReceiver, err)
	case errInsufficientAllowance.Selector():
		return fmt.Errorf("%w: %w", ErrInsufficientAllowance, err)
	case errInvalidApprover.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidApprover, err)
	case errInvalidSpender.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidSpender, err)
	default:
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}
}
