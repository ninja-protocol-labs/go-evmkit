package erc721

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/contracts/erc165"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// IERC721 mirrors IERC721.sol, including its IERC165 embedding.
type IERC721 interface {
	erc165.IERC165
	BalanceOf(ctx context.Context, contract, owner *types.Address, block string) (*big.Int, error)
	OwnerOf(ctx context.Context, contract *types.Address, tokenId *big.Int, block string) (*types.Address, error)
	GetApproved(ctx context.Context, contract *types.Address, tokenId *big.Int, block string) (*types.Address, error)
	IsApprovedForAll(ctx context.Context, contract, owner, operator *types.Address, block string) (bool, error)
	Approve(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SetApprovalForAll(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	TransferFrom(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SafeTransferFrom(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

var _ IERC721 = (*ERC721)(nil)

// IERC721Metadata mirrors IERC721Metadata.sol, including its IERC721 embedding.
type IERC721Metadata interface {
	IERC721
	Name(ctx context.Context, contract *types.Address, block string) (string, error)
	Symbol(ctx context.Context, contract *types.Address, block string) (string, error)
	TokenURI(ctx context.Context, contract *types.Address, tokenId *big.Int, block string) (string, error)
}

var _ IERC721Metadata = (*ERC721)(nil)

// ERC721 calls any ERC-721 token contract over a Client; it holds no
// contract address, so one instance serves every collection. SupportsInterface
// is promoted from the embedded *erc165.ERC165, mirroring IERC721 is IERC165.
type ERC721 struct {
	*erc165.ERC165

	c rpc.Client
}

// New returns an ERC721 calling collections via cli.
func New(cli rpc.Client) *ERC721 {
	return &ERC721{
		c:      cli,
		ERC165: erc165.New(cli),
	}
}

// Name returns contract's name.
func (e *ERC721) Name(ctx context.Context, contract *types.Address, block string) (string, error) {
	data, err := e.call(ctx, nil, contract, EncodeName(), block)
	if err != nil {
		return "", fmt.Errorf("erc721: name: %w", err)
	}
	s, err := DecodeName(data)
	if err != nil {
		return "", fmt.Errorf("erc721: name: %w", err)
	}
	return s, nil
}

// Symbol returns contract's symbol.
func (e *ERC721) Symbol(ctx context.Context, contract *types.Address, block string) (string, error) {
	data, err := e.call(ctx, nil, contract, EncodeSymbol(), block)
	if err != nil {
		return "", fmt.Errorf("erc721: symbol: %w", err)
	}
	s, err := DecodeSymbol(data)
	if err != nil {
		return "", fmt.Errorf("erc721: symbol: %w", err)
	}
	return s, nil
}

// TokenURI returns tokenId's URI on contract.
func (e *ERC721) TokenURI(ctx context.Context, contract *types.Address, tokenId *big.Int, block string) (string, error) {
	calldata, err := EncodeTokenURI(tokenId)
	if err != nil {
		return "", fmt.Errorf("erc721: tokenURI: %w", err)
	}
	data, err := e.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return "", fmt.Errorf("erc721: tokenURI: %w", err)
	}
	s, err := DecodeTokenURI(data)
	if err != nil {
		return "", fmt.Errorf("erc721: tokenURI: %w", err)
	}
	return s, nil
}

// BalanceOf returns owner's balance of contract.
func (e *ERC721) BalanceOf(ctx context.Context, contract, owner *types.Address, block string) (*big.Int, error) {
	calldata, err := EncodeBalanceOf(owner)
	if err != nil {
		return nil, fmt.Errorf("erc721: balanceOf: %w", err)
	}
	data, err := e.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc721: balanceOf: %w", err)
	}
	n, err := DecodeBalanceOf(data)
	if err != nil {
		return nil, fmt.Errorf("erc721: balanceOf: %w", err)
	}
	return n, nil
}

// OwnerOf returns tokenId's owner on contract.
func (e *ERC721) OwnerOf(ctx context.Context, contract *types.Address, tokenId *big.Int, block string) (*types.Address, error) {
	calldata, err := EncodeOwnerOf(tokenId)
	if err != nil {
		return nil, fmt.Errorf("erc721: ownerOf: %w", err)
	}
	data, err := e.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc721: ownerOf: %w", err)
	}
	addr, err := DecodeOwnerOf(data)
	if err != nil {
		return nil, fmt.Errorf("erc721: ownerOf: %w", err)
	}
	return addr, nil
}

// GetApproved returns tokenId's approved operator on contract.
func (e *ERC721) GetApproved(ctx context.Context, contract *types.Address, tokenId *big.Int, block string) (*types.Address, error) {
	calldata, err := EncodeGetApproved(tokenId)
	if err != nil {
		return nil, fmt.Errorf("erc721: getApproved: %w", err)
	}
	data, err := e.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc721: getApproved: %w", err)
	}
	addr, err := DecodeGetApproved(data)
	if err != nil {
		return nil, fmt.Errorf("erc721: getApproved: %w", err)
	}
	return addr, nil
}

// IsApprovedForAll reports whether operator may manage all of owner's tokens on contract.
func (e *ERC721) IsApprovedForAll(ctx context.Context, contract, owner, operator *types.Address, block string) (bool, error) {
	calldata, err := EncodeIsApprovedForAll(owner, operator)
	if err != nil {
		return false, fmt.Errorf("erc721: isApprovedForAll: %w", err)
	}
	data, err := e.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return false, fmt.Errorf("erc721: isApprovedForAll: %w", err)
	}
	ok, err := DecodeIsApprovedForAll(data)
	if err != nil {
		return false, fmt.Errorf("erc721: isApprovedForAll: %w", err)
	}
	return ok, nil
}

// CreateApprove builds an unsigned approve(to, tokenId) transaction via b.
func (e *ERC721) CreateApprove(ctx context.Context, b tx.Builder, contract, from, to *types.Address, tokenId *big.Int) (core.Transaction, error) {
	calldata, err := EncodeApprove(to, tokenId)
	if err != nil {
		return nil, fmt.Errorf("erc721: createApprove: %w", err)
	}

	unsigned, err := b.Build(from, contract, nil, calldata).Pack(ctx, e.c)
	if err != nil {
		return nil, fmt.Errorf("erc721: createApprove: %w", e.wrapRevertError(err))
	}
	return unsigned, nil
}

// ApproveGas estimates the gas an approve transaction would use, without
// resolving chainID/nonce/gasPrice or building the transaction itself.
func (e *ERC721) ApproveGas(ctx context.Context, b tx.Builder, contract, from, to *types.Address, tokenId *big.Int) (uint64, error) {
	calldata, err := EncodeApprove(to, tokenId)
	if err != nil {
		return 0, fmt.Errorf("erc721: approveGas: %w", err)
	}

	gas, err := b.Build(from, contract, nil, calldata).Gas(ctx, e.c)
	if err != nil {
		return 0, fmt.Errorf("erc721: approveGas: %w", e.wrapRevertError(err))
	}
	return gas, nil
}

// Approve broadcasts an already-signed approve transaction.
func (e *ERC721) Approve(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("erc721: approve: %w", err)
	}
	hash, err := tx.Broadcast(ctx, e.c, raw)
	if err != nil {
		return nil, fmt.Errorf("erc721: approve: %w", err)
	}
	return hash, nil
}

// CreateSetApprovalForAll builds an unsigned setApprovalForAll(operator, approved) transaction via b.
func (e *ERC721) CreateSetApprovalForAll(ctx context.Context, b tx.Builder, contract, from, operator *types.Address, approved bool) (core.Transaction, error) {
	calldata, err := EncodeSetApprovalForAll(operator, approved)
	if err != nil {
		return nil, fmt.Errorf("erc721: createSetApprovalForAll: %w", err)
	}

	unsigned, err := b.Build(from, contract, nil, calldata).Pack(ctx, e.c)
	if err != nil {
		return nil, fmt.Errorf("erc721: createSetApprovalForAll: %w", e.wrapRevertError(err))
	}
	return unsigned, nil
}

// SetApprovalForAllGas estimates the gas a setApprovalForAll transaction would use, without
// resolving chainID/nonce/gasPrice or building the transaction itself.
func (e *ERC721) SetApprovalForAllGas(ctx context.Context, b tx.Builder, contract, from, operator *types.Address, approved bool) (uint64, error) {
	calldata, err := EncodeSetApprovalForAll(operator, approved)
	if err != nil {
		return 0, fmt.Errorf("erc721: setApprovalForAllGas: %w", err)
	}

	gas, err := b.Build(from, contract, nil, calldata).Gas(ctx, e.c)
	if err != nil {
		return 0, fmt.Errorf("erc721: setApprovalForAllGas: %w", e.wrapRevertError(err))
	}
	return gas, nil
}

// SetApprovalForAll broadcasts an already-signed setApprovalForAll transaction.
func (e *ERC721) SetApprovalForAll(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("erc721: setApprovalForAll: %w", err)
	}
	hash, err := tx.Broadcast(ctx, e.c, raw)
	if err != nil {
		return nil, fmt.Errorf("erc721: setApprovalForAll: %w", err)
	}
	return hash, nil
}

// CreateTransferFrom builds an unsigned transferFrom(from, to, tokenId) transaction via b, sent by sender.
func (e *ERC721) CreateTransferFrom(ctx context.Context, b tx.Builder, contract, sender, from, to *types.Address, tokenId *big.Int) (core.Transaction, error) {
	calldata, err := EncodeTransferFrom(from, to, tokenId)
	if err != nil {
		return nil, fmt.Errorf("erc721: createTransferFrom: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, e.c)
	if err != nil {
		return nil, fmt.Errorf("erc721: createTransferFrom: %w", e.wrapRevertError(err))
	}
	return unsigned, nil
}

// TransferFromGas estimates the gas a transferFrom transaction would use, without
// resolving chainID/nonce/gasPrice or building the transaction itself.
func (e *ERC721) TransferFromGas(ctx context.Context, b tx.Builder, contract, sender, from, to *types.Address, tokenId *big.Int) (uint64, error) {
	calldata, err := EncodeTransferFrom(from, to, tokenId)
	if err != nil {
		return 0, fmt.Errorf("erc721: transferFromGas: %w", err)
	}

	gas, err := b.Build(sender, contract, nil, calldata).Gas(ctx, e.c)
	if err != nil {
		return 0, fmt.Errorf("erc721: transferFromGas: %w", e.wrapRevertError(err))
	}
	return gas, nil
}

// TransferFrom broadcasts an already-signed transferFrom transaction.
func (e *ERC721) TransferFrom(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("erc721: transferFrom: %w", err)
	}
	hash, err := tx.Broadcast(ctx, e.c, raw)
	if err != nil {
		return nil, fmt.Errorf("erc721: transferFrom: %w", err)
	}
	return hash, nil
}

// CreateSafeTransferFrom builds an unsigned safeTransferFrom(from, to, tokenId) transaction via b, sent by sender.
func (e *ERC721) CreateSafeTransferFrom(ctx context.Context, b tx.Builder, contract, sender, from, to *types.Address, tokenId *big.Int) (core.Transaction, error) {
	calldata, err := EncodeSafeTransferFrom(from, to, tokenId)
	if err != nil {
		return nil, fmt.Errorf("erc721: createSafeTransferFrom: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, e.c)
	if err != nil {
		return nil, fmt.Errorf("erc721: createSafeTransferFrom: %w", e.wrapRevertError(err))
	}
	return unsigned, nil
}

// SafeTransferFromGas estimates the gas a safeTransferFrom transaction would use, without
// resolving chainID/nonce/gasPrice or building the transaction itself.
func (e *ERC721) SafeTransferFromGas(ctx context.Context, b tx.Builder, contract, sender, from, to *types.Address, tokenId *big.Int) (uint64, error) {
	calldata, err := EncodeSafeTransferFrom(from, to, tokenId)
	if err != nil {
		return 0, fmt.Errorf("erc721: safeTransferFromGas: %w", err)
	}

	gas, err := b.Build(sender, contract, nil, calldata).Gas(ctx, e.c)
	if err != nil {
		return 0, fmt.Errorf("erc721: safeTransferFromGas: %w", e.wrapRevertError(err))
	}
	return gas, nil
}

// CreateSafeTransferFromWithData builds an unsigned safeTransferFrom(from, to, tokenId, data) transaction via b, sent by sender.
func (e *ERC721) CreateSafeTransferFromWithData(ctx context.Context, b tx.Builder, contract, sender, from, to *types.Address, tokenId *big.Int, callData []byte) (core.Transaction, error) {
	calldata, err := EncodeSafeTransferFromWithData(from, to, tokenId, callData)
	if err != nil {
		return nil, fmt.Errorf("erc721: createSafeTransferFromWithData: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, e.c)
	if err != nil {
		return nil, fmt.Errorf("erc721: createSafeTransferFromWithData: %w", e.wrapRevertError(err))
	}
	return unsigned, nil
}

// SafeTransferFromWithDataGas estimates the gas a safeTransferFrom-with-data transaction would use, without
// resolving chainID/nonce/gasPrice or building the transaction itself.
func (e *ERC721) SafeTransferFromWithDataGas(ctx context.Context, b tx.Builder, contract, sender, from, to *types.Address, tokenId *big.Int, callData []byte) (uint64, error) {
	calldata, err := EncodeSafeTransferFromWithData(from, to, tokenId, callData)
	if err != nil {
		return 0, fmt.Errorf("erc721: safeTransferFromWithDataGas: %w", err)
	}

	gas, err := b.Build(sender, contract, nil, calldata).Gas(ctx, e.c)
	if err != nil {
		return 0, fmt.Errorf("erc721: safeTransferFromWithDataGas: %w", e.wrapRevertError(err))
	}
	return gas, nil
}

// SafeTransferFrom broadcasts an already-signed safeTransferFrom transaction.
func (e *ERC721) SafeTransferFrom(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("erc721: safeTransferFrom: %w", err)
	}
	hash, err := tx.Broadcast(ctx, e.c, raw)
	if err != nil {
		return nil, fmt.Errorf("erc721: safeTransferFrom: %w", err)
	}
	return hash, nil
}

// Metadata batches name/symbol on contract into one call via the Multicall3
// contract deployed at multicall.
func (e *ERC721) Metadata(ctx context.Context, multicall, contract *types.Address, allowFailure bool, block string) (*Metadata, error) {
	calldata, err := EncodeMetadata(contract, allowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc721: metadata: %w", err)
	}

	data, err := e.call(ctx, nil, multicall, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc721: metadata: %w", err)
	}

	m, err := DecodeMetadata(data)
	if err != nil {
		return nil, fmt.Errorf("erc721: metadata: %w", err)
	}
	return m, nil
}

// MetadataWithTokenID is Metadata plus tokenId's URI and owner in the same
// call, via the Multicall3 contract deployed at multicall.
func (e *ERC721) MetadataWithTokenID(ctx context.Context, multicall, contract *types.Address, tokenId *big.Int, allowFailure bool, block string) (*MetadataWithTokenID, error) {
	calldata, err := EncodeMetadataWithTokenID(contract, tokenId, allowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc721: metadataWithTokenID: %w", err)
	}

	data, err := e.call(ctx, nil, multicall, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc721: metadataWithTokenID: %w", err)
	}

	mt, err := DecodeMetadataWithTokenID(tokenId, data)
	if err != nil {
		return nil, fmt.Errorf("erc721: metadataWithTokenID: %w", err)
	}
	return mt, nil
}

// AllowanceWithTokenID batches ownerOf(tokenId) and getApproved(tokenId) into
// one call, via the Multicall3 contract deployed at multicall.
func (e *ERC721) AllowanceWithTokenID(ctx context.Context, multicall, contract *types.Address, tokenId *big.Int, allowFailure bool, block string) (*AllowanceWithTokenID, error) {
	calldata, err := EncodeAllowanceWithTokenID(contract, tokenId, allowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc721: allowanceWithTokenID: %w", err)
	}

	data, err := e.call(ctx, nil, multicall, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc721: allowanceWithTokenID: %w", err)
	}

	awt, err := DecodeAllowanceWithTokenID(tokenId, data)
	if err != nil {
		return nil, fmt.Errorf("erc721: allowanceWithTokenID: %w", err)
	}
	return awt, nil
}

// ApprovalForAllWithBalance batches isApprovedForAll(owner, operator) and
// both their balances into one call, via the Multicall3 contract deployed at multicall.
func (e *ERC721) ApprovalForAllWithBalance(ctx context.Context, multicall, contract, owner, operator *types.Address, allowFailure bool, block string) (*ApprovalForAllWithBalance, error) {
	calldata, err := EncodeApprovalForAllWithBalance(contract, owner, operator, allowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc721: approvalForAllWithBalance: %w", err)
	}

	data, err := e.call(ctx, nil, multicall, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc721: approvalForAllWithBalance: %w", err)
	}

	afawb, err := DecodeApprovalForAllWithBalance(owner, operator, data)
	if err != nil {
		return nil, fmt.Errorf("erc721: approvalForAllWithBalance: %w", err)
	}
	return afawb, nil
}

// TokenOwners batches ownerOf(tokenIds[i]) on contract for each i into one
// call, via the Multicall3 contract deployed at multicall.
func (e *ERC721) TokenOwners(ctx context.Context, multicall, contract *types.Address, tokenIds []*big.Int, allowFailure bool, block string) (*TokenOwners, error) {
	calldata, err := EncodeTokenOwners(contract, tokenIds, allowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc721: tokenOwners: %w", err)
	}

	data, err := e.call(ctx, nil, multicall, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc721: tokenOwners: %w", err)
	}

	tos, err := DecodeTokenOwners(contract, tokenIds, data)
	if err != nil {
		return nil, fmt.Errorf("erc721: tokenOwners: %w", err)
	}
	return tos, nil
}

// TokenOperators batches getApproved(tokenIds[i]) on contract for each i into
// one call, via the Multicall3 contract deployed at multicall.
func (e *ERC721) TokenOperators(ctx context.Context, multicall, contract *types.Address, tokenIds []*big.Int, allowFailure bool, block string) (*TokenOperators, error) {
	calldata, err := EncodeTokenOperators(contract, tokenIds, allowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc721: tokenOperators: %w", err)
	}

	data, err := e.call(ctx, nil, multicall, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc721: tokenOperators: %w", err)
	}

	tops, err := DecodeTokenOperators(contract, tokenIds, data)
	if err != nil {
		return nil, fmt.Errorf("erc721: tokenOperators: %w", err)
	}
	return tops, nil
}

// AddressBalances batches balanceOf(address) across each contract in
// contracts into one call, via the Multicall3 contract deployed at multicall.
func (e *ERC721) AddressBalances(ctx context.Context, multicall, address *types.Address, contracts []*types.Address, allowFailure bool, block string) (*AddressBalances, error) {
	calldata, err := EncodeAddressBalances(address, contracts, allowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc721: addressBalances: %w", err)
	}

	data, err := e.call(ctx, nil, multicall, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc721: addressBalances: %w", err)
	}

	ab, err := DecodeAddressBalances(address, contracts, data)
	if err != nil {
		return nil, fmt.Errorf("erc721: addressBalances: %w", err)
	}
	return ab, nil
}

// TokenBalances batches balanceOf(addresses[i]) on contract for each i into
// one call, via the Multicall3 contract deployed at multicall.
func (e *ERC721) TokenBalances(ctx context.Context, multicall, contract *types.Address, addresses []*types.Address, allowFailure bool, block string) (*TokenBalances, error) {
	calldata, err := EncodeTokenBalances(contract, addresses, allowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc721: tokenBalances: %w", err)
	}

	data, err := e.call(ctx, nil, multicall, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc721: tokenBalances: %w", err)
	}

	tb, err := DecodeTokenBalances(contract, addresses, data)
	if err != nil {
		return nil, fmt.Errorf("erc721: tokenBalances: %w", err)
	}
	return tb, nil
}

// OwnerPairs batches ownerOf(tokenIds[i]) on contracts[i] for each i into one
// call (paired positionally, not every combination), via the Multicall3
// contract deployed at multicall.
func (e *ERC721) OwnerPairs(ctx context.Context, multicall *types.Address, contracts []*types.Address, tokenIds []*big.Int, allowFailure bool, block string) ([]TokenOwnerPair, error) {
	calldata, err := EncodeOwnerPairs(contracts, tokenIds, allowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc721: ownerPairs: %w", err)
	}

	data, err := e.call(ctx, nil, multicall, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc721: ownerPairs: %w", err)
	}

	pairs, err := DecodeOwnerPairs(contracts, tokenIds, data)
	if err != nil {
		return nil, fmt.Errorf("erc721: ownerPairs: %w", err)
	}
	return pairs, nil
}

// call eth_calls contract with calldata (from msg.sender from, if non-nil), decoding a reverted custom error to its sentinel.
func (e *ERC721) call(ctx context.Context, from, contract *types.Address, calldata []byte, block string) ([]byte, error) {
	params := map[string]any{
		"to":   contract.String(),
		"data": encoding.Hex.EncodePrefixed(calldata),
	}
	if from != nil {
		params["from"] = from.String()
	}

	var raw string
	if err := e.c.Call(ctx, rpc.ETHCall("call", params, block, &raw)); err != nil {
		return nil, e.wrapRevertError(err)
	}
	return encoding.Hex.Decode(raw)
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
