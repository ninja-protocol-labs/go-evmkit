// Package erc721 is a client for ERC-721 tokens, mirroring IERC721.sol and
// IERC721Metadata.sol.
package erc721

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
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc165"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// IERC721 mirrors IERC721.sol, including its IERC165 embedding.
type IERC721 interface {
	erc165.IERC165

	BalanceOf(ctx context.Context, opts *bind.CallOptions, req *BalanceOfRequest) (*big.Int, error)
	OwnerOf(ctx context.Context, opts *bind.CallOptions, req *OwnerOfRequest) (*types.Address, error)
	GetApproved(ctx context.Context, opts *bind.CallOptions, req *GetApprovedRequest) (*types.Address, error)
	IsApprovedForAll(ctx context.Context, opts *bind.CallOptions, req *IsApprovedForAllRequest) (bool, error)

	Approve(ctx context.Context, opts *bind.TransactOptions, req *ApproveRequest) (core.Transaction, *types.Hash, error)
	SetApprovalForAll(ctx context.Context, opts *bind.TransactOptions, req *SetApprovalForAllRequest) (core.Transaction, *types.Hash, error)
	TransferFrom(ctx context.Context, opts *bind.TransactOptions, req *TransferFromRequest) (core.Transaction, *types.Hash, error)
	SafeTransferFrom(ctx context.Context, opts *bind.TransactOptions, req *SafeTransferFromRequest) (core.Transaction, *types.Hash, error)
	SafeTransferFromWithData(ctx context.Context, opts *bind.TransactOptions, req *SafeTransferFromWithDataRequest) (core.Transaction, *types.Hash, error)
}

// IERC721Metadata mirrors IERC721Metadata.sol, including its IERC721 embedding.
type IERC721Metadata interface {
	IERC721

	Name(ctx context.Context, opts *bind.CallOptions, req *NameRequest) (string, error)
	Symbol(ctx context.Context, opts *bind.CallOptions, req *SymbolRequest) (string, error)
	TokenURI(ctx context.Context, opts *bind.CallOptions, req *TokenURIRequest) (string, error)
}

// IERC721Batcher groups the Multicall3-backed batch reads; it mirrors no
// Solidity interface.
type IERC721Batcher interface {
	Metadata(ctx context.Context, opts *bind.CallOptions, req *MetadataRequest) (*Metadata, error)
	MetadataWithTokenID(ctx context.Context, opts *bind.CallOptions, req *MetadataWithTokenIDRequest) (*MetadataWithTokenID, error)
	AllowanceWithTokenID(ctx context.Context, opts *bind.CallOptions, req *AllowanceWithTokenIDRequest) (*AllowanceWithTokenID, error)
	ApprovalForAllWithBalance(ctx context.Context, opts *bind.CallOptions, req *ApprovalForAllWithBalanceRequest) (*ApprovalForAllWithBalance, error)
	TokenOwners(ctx context.Context, opts *bind.CallOptions, req *TokenOwnersRequest) (*TokenOwners, error)
	TokenOperators(ctx context.Context, opts *bind.CallOptions, req *TokenOperatorsRequest) (*TokenOperators, error)
	AddressBalances(ctx context.Context, opts *bind.CallOptions, req *AddressBalancesRequest) (*AddressBalances, error)
	TokenBalances(ctx context.Context, opts *bind.CallOptions, req *TokenBalancesRequest) (*TokenBalances, error)
	OwnerPairs(ctx context.Context, opts *bind.CallOptions, req *OwnerPairsRequest) ([]TokenOwnerPair, error)
}

var (
	_ IERC721         = (*ERC721)(nil)
	_ IERC721Metadata = (*ERC721)(nil)
	_ IERC721Batcher  = (*ERC721)(nil)
)

// ERC721 calls any ERC-721 token contract over a Client; SupportsInterface is
// promoted from the embedded *erc165.ERC165, mirroring IERC721 is IERC165.
type ERC721 struct {
	*erc165.ERC165

	c rpc.Client
}

// New returns an ERC721 calling contracts via cli.
func New(cli rpc.Client) *ERC721 {
	return &ERC721{
		c:      cli,
		ERC165: erc165.New(cli),
	}
}

// Name returns req.Contract's name.
func (e *ERC721) Name(ctx context.Context, opts *bind.CallOptions, req *NameRequest) (string, error) {
	if err := req.Validate(); err != nil {
		return "", fmt.Errorf("erc721: name: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, EncodeName())
	if err != nil {
		return "", fmt.Errorf("erc721: name: %w", e.wrapRevertError(err))
	}

	out, err := DecodeName(data)
	if err != nil {
		return "", fmt.Errorf("erc721: name: %w", err)
	}

	return out, nil
}

// Symbol returns req.Contract's symbol.
func (e *ERC721) Symbol(ctx context.Context, opts *bind.CallOptions, req *SymbolRequest) (string, error) {
	if err := req.Validate(); err != nil {
		return "", fmt.Errorf("erc721: symbol: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, EncodeSymbol())
	if err != nil {
		return "", fmt.Errorf("erc721: symbol: %w", e.wrapRevertError(err))
	}

	out, err := DecodeSymbol(data)
	if err != nil {
		return "", fmt.Errorf("erc721: symbol: %w", err)
	}

	return out, nil
}

// TokenURI returns req.TokenID's URI on req.Contract.
func (e *ERC721) TokenURI(ctx context.Context, opts *bind.CallOptions, req *TokenURIRequest) (string, error) {
	if err := req.Validate(); err != nil {
		return "", fmt.Errorf("erc721: tokenURI: %w", err)
	}

	calldata, err := EncodeTokenURI(req.TokenID)
	if err != nil {
		return "", fmt.Errorf("erc721: tokenURI: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return "", fmt.Errorf("erc721: tokenURI: %w", e.wrapRevertError(err))
	}

	out, err := DecodeTokenURI(data)
	if err != nil {
		return "", fmt.Errorf("erc721: tokenURI: %w", err)
	}

	return out, nil
}

// BalanceOf returns req.Owner's balance on req.Contract.
func (e *ERC721) BalanceOf(ctx context.Context, opts *bind.CallOptions, req *BalanceOfRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc721: balanceOf: %w", err)
	}

	calldata, err := EncodeBalanceOf(req.Owner)
	if err != nil {
		return nil, fmt.Errorf("erc721: balanceOf: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc721: balanceOf: %w", e.wrapRevertError(err))
	}

	out, err := DecodeBalanceOf(data)
	if err != nil {
		return nil, fmt.Errorf("erc721: balanceOf: %w", err)
	}

	return out, nil
}

// OwnerOf returns req.TokenID's owner on req.Contract.
func (e *ERC721) OwnerOf(ctx context.Context, opts *bind.CallOptions, req *OwnerOfRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc721: ownerOf: %w", err)
	}

	calldata, err := EncodeOwnerOf(req.TokenID)
	if err != nil {
		return nil, fmt.Errorf("erc721: ownerOf: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc721: ownerOf: %w", e.wrapRevertError(err))
	}

	out, err := DecodeOwnerOf(data)
	if err != nil {
		return nil, fmt.Errorf("erc721: ownerOf: %w", err)
	}

	return out, nil
}

// GetApproved returns req.TokenID's approved address on req.Contract.
func (e *ERC721) GetApproved(ctx context.Context, opts *bind.CallOptions, req *GetApprovedRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc721: getApproved: %w", err)
	}

	calldata, err := EncodeGetApproved(req.TokenID)
	if err != nil {
		return nil, fmt.Errorf("erc721: getApproved: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc721: getApproved: %w", e.wrapRevertError(err))
	}

	out, err := DecodeGetApproved(data)
	if err != nil {
		return nil, fmt.Errorf("erc721: getApproved: %w", err)
	}

	return out, nil
}

// IsApprovedForAll reports whether req.Operator may manage all of req.Owner's tokens.
func (e *ERC721) IsApprovedForAll(ctx context.Context, opts *bind.CallOptions, req *IsApprovedForAllRequest) (bool, error) {
	if err := req.Validate(); err != nil {
		return false, fmt.Errorf("erc721: isApprovedForAll: %w", err)
	}

	calldata, err := EncodeIsApprovedForAll(req.Owner, req.Operator)
	if err != nil {
		return false, fmt.Errorf("erc721: isApprovedForAll: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return false, fmt.Errorf("erc721: isApprovedForAll: %w", e.wrapRevertError(err))
	}

	out, err := DecodeIsApprovedForAll(data)
	if err != nil {
		return false, fmt.Errorf("erc721: isApprovedForAll: %w", err)
	}

	return out, nil
}

// Approve sends approve(to, tokenId) on req.Contract, signed by opts.Signer.
func (e *ERC721) Approve(ctx context.Context, opts *bind.TransactOptions, req *ApproveRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("erc721: approve: %w", err)
	}

	calldata, err := EncodeApprove(req.To, req.TokenID)
	if err != nil {
		return nil, nil, fmt.Errorf("erc721: approve: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("erc721: approve: %w", e.wrapRevertError(err))
	}

	return signed, hash, nil
}

// SetApprovalForAll sends setApprovalForAll(operator, approved) on req.Contract, signed by opts.Signer.
func (e *ERC721) SetApprovalForAll(ctx context.Context, opts *bind.TransactOptions, req *SetApprovalForAllRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("erc721: setApprovalForAll: %w", err)
	}

	calldata, err := EncodeSetApprovalForAll(req.Operator, req.Approved)
	if err != nil {
		return nil, nil, fmt.Errorf("erc721: setApprovalForAll: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("erc721: setApprovalForAll: %w", e.wrapRevertError(err))
	}

	return signed, hash, nil
}

// TransferFrom sends transferFrom(from, to, tokenId) on req.Contract, signed by opts.Signer.
func (e *ERC721) TransferFrom(ctx context.Context, opts *bind.TransactOptions, req *TransferFromRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("erc721: transferFrom: %w", err)
	}

	calldata, err := EncodeTransferFrom(req.From, req.To, req.TokenID)
	if err != nil {
		return nil, nil, fmt.Errorf("erc721: transferFrom: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("erc721: transferFrom: %w", e.wrapRevertError(err))
	}

	return signed, hash, nil
}

// SafeTransferFrom sends safeTransferFrom(from, to, tokenId) on req.Contract, signed by opts.Signer.
func (e *ERC721) SafeTransferFrom(ctx context.Context, opts *bind.TransactOptions, req *SafeTransferFromRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("erc721: safeTransferFrom: %w", err)
	}

	calldata, err := EncodeSafeTransferFrom(req.From, req.To, req.TokenID)
	if err != nil {
		return nil, nil, fmt.Errorf("erc721: safeTransferFrom: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("erc721: safeTransferFrom: %w", e.wrapRevertError(err))
	}

	return signed, hash, nil
}

// SafeTransferFromWithData sends safeTransferFrom(from, to, tokenId, data) on req.Contract, signed by opts.Signer.
func (e *ERC721) SafeTransferFromWithData(ctx context.Context, opts *bind.TransactOptions, req *SafeTransferFromWithDataRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("erc721: safeTransferFromWithData: %w", err)
	}

	calldata, err := EncodeSafeTransferFromWithData(req.From, req.To, req.TokenID, req.Data)
	if err != nil {
		return nil, nil, fmt.Errorf("erc721: safeTransferFromWithData: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("erc721: safeTransferFromWithData: %w", e.wrapRevertError(err))
	}

	return signed, hash, nil
}

// Metadata batches name/symbol on req.Contract via Multicall3 at req.Multicall.
func (e *ERC721) Metadata(ctx context.Context, opts *bind.CallOptions, req *MetadataRequest) (*Metadata, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc721: metadata: %w", err)
	}

	calldata, err := EncodeMetadata(req.Contract, req.AllowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc721: metadata: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Multicall, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc721: metadata: %w", e.wrapRevertError(err))
	}

	out, err := DecodeMetadata(data)
	if err != nil {
		return nil, fmt.Errorf("erc721: metadata: %w", err)
	}

	return out, nil
}

// MetadataWithTokenID is Metadata plus req.TokenID's URI and owner in the same call.
func (e *ERC721) MetadataWithTokenID(ctx context.Context, opts *bind.CallOptions, req *MetadataWithTokenIDRequest) (*MetadataWithTokenID, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc721: metadataWithTokenID: %w", err)
	}

	calldata, err := EncodeMetadataWithTokenID(req.Contract, req.TokenID, req.AllowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc721: metadataWithTokenID: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Multicall, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc721: metadataWithTokenID: %w", e.wrapRevertError(err))
	}

	out, err := DecodeMetadataWithTokenID(req.TokenID, data)
	if err != nil {
		return nil, fmt.Errorf("erc721: metadataWithTokenID: %w", err)
	}

	return out, nil
}

// AllowanceWithTokenID batches req.TokenID's owner and approved address.
func (e *ERC721) AllowanceWithTokenID(ctx context.Context, opts *bind.CallOptions, req *AllowanceWithTokenIDRequest) (*AllowanceWithTokenID, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc721: allowanceWithTokenID: %w", err)
	}

	calldata, err := EncodeAllowanceWithTokenID(req.Contract, req.TokenID, req.AllowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc721: allowanceWithTokenID: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Multicall, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc721: allowanceWithTokenID: %w", e.wrapRevertError(err))
	}

	out, err := DecodeAllowanceWithTokenID(req.TokenID, data)
	if err != nil {
		return nil, fmt.Errorf("erc721: allowanceWithTokenID: %w", err)
	}

	return out, nil
}

// ApprovalForAllWithBalance batches isApprovedForAll(owner, operator) and both balances.
func (e *ERC721) ApprovalForAllWithBalance(ctx context.Context, opts *bind.CallOptions, req *ApprovalForAllWithBalanceRequest) (*ApprovalForAllWithBalance, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc721: approvalForAllWithBalance: %w", err)
	}

	calldata, err := EncodeApprovalForAllWithBalance(req.Contract, req.Owner, req.Operator, req.AllowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc721: approvalForAllWithBalance: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Multicall, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc721: approvalForAllWithBalance: %w", e.wrapRevertError(err))
	}

	out, err := DecodeApprovalForAllWithBalance(req.Owner, req.Operator, data)
	if err != nil {
		return nil, fmt.Errorf("erc721: approvalForAllWithBalance: %w", err)
	}

	return out, nil
}

// TokenOwners batches ownerOf on req.Contract for each of req.TokenIDs.
func (e *ERC721) TokenOwners(ctx context.Context, opts *bind.CallOptions, req *TokenOwnersRequest) (*TokenOwners, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc721: tokenOwners: %w", err)
	}

	calldata, err := EncodeTokenOwners(req.Contract, req.TokenIDs, req.AllowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc721: tokenOwners: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Multicall, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc721: tokenOwners: %w", e.wrapRevertError(err))
	}

	out, err := DecodeTokenOwners(req.Contract, req.TokenIDs, data)
	if err != nil {
		return nil, fmt.Errorf("erc721: tokenOwners: %w", err)
	}

	return out, nil
}

// TokenOperators batches getApproved on req.Contract for each of req.TokenIDs.
func (e *ERC721) TokenOperators(ctx context.Context, opts *bind.CallOptions, req *TokenOperatorsRequest) (*TokenOperators, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc721: tokenOperators: %w", err)
	}

	calldata, err := EncodeTokenOperators(req.Contract, req.TokenIDs, req.AllowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc721: tokenOperators: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Multicall, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc721: tokenOperators: %w", e.wrapRevertError(err))
	}

	out, err := DecodeTokenOperators(req.Contract, req.TokenIDs, data)
	if err != nil {
		return nil, fmt.Errorf("erc721: tokenOperators: %w", err)
	}

	return out, nil
}

// AddressBalances batches req.Account's balance on each of req.Contracts.
func (e *ERC721) AddressBalances(ctx context.Context, opts *bind.CallOptions, req *AddressBalancesRequest) (*AddressBalances, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc721: addressBalances: %w", err)
	}

	calldata, err := EncodeAddressBalances(req.Account, req.Contracts, req.AllowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc721: addressBalances: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Multicall, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc721: addressBalances: %w", e.wrapRevertError(err))
	}

	out, err := DecodeAddressBalances(req.Account, req.Contracts, data)
	if err != nil {
		return nil, fmt.Errorf("erc721: addressBalances: %w", err)
	}

	return out, nil
}

// TokenBalances batches each of req.Accounts' balance on req.Contract.
func (e *ERC721) TokenBalances(ctx context.Context, opts *bind.CallOptions, req *TokenBalancesRequest) (*TokenBalances, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc721: tokenBalances: %w", err)
	}

	calldata, err := EncodeTokenBalances(req.Contract, req.Accounts, req.AllowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc721: tokenBalances: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Multicall, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc721: tokenBalances: %w", e.wrapRevertError(err))
	}

	out, err := DecodeTokenBalances(req.Contract, req.Accounts, data)
	if err != nil {
		return nil, fmt.Errorf("erc721: tokenBalances: %w", err)
	}

	return out, nil
}

// OwnerPairs batches ownerOf(req.TokenIDs[i]) on req.Contracts[i], paired by index.
func (e *ERC721) OwnerPairs(ctx context.Context, opts *bind.CallOptions, req *OwnerPairsRequest) ([]TokenOwnerPair, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc721: ownerPairs: %w", err)
	}

	calldata, err := EncodeOwnerPairs(req.Contracts, req.TokenIDs, req.AllowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc721: ownerPairs: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Multicall, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc721: ownerPairs: %w", e.wrapRevertError(err))
	}

	out, err := DecodeOwnerPairs(req.Contracts, req.TokenIDs, data)
	if err != nil {
		return nil, fmt.Errorf("erc721: ownerPairs: %w", err)
	}

	return out, nil
}

// wrapRevertError maps a known IERC721Errors custom error in err to its sentinel.
func (_ *ERC721) wrapRevertError(err error) error {
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
	case errInvalidOwner.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidOwner, err)
	case errNonexistentToken.Selector():
		return fmt.Errorf("%w: %w", ErrNonexistentToken, err)
	case errIncorrectOwner.Selector():
		return fmt.Errorf("%w: %w", ErrIncorrectOwner, err)
	case errInvalidSender.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidSender, err)
	case errInvalidReceiver.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidReceiver, err)
	case errInsufficientApproval.Selector():
		return fmt.Errorf("%w: %w", ErrInsufficientApproval, err)
	case errInvalidApprover.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidApprover, err)
	case errInvalidOperator.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidOperator, err)
	default:
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}
}
