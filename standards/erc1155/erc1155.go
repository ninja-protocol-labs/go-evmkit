// Package erc1155 encodes calls to and decodes results from an ERC-1155
// token contract. Go types mirror the interfaces in
// solidity/interfaces/IERC1155.sol and solidity/interfaces/IERC1155MetadataURI.sol.
package erc1155

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc165"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// IERC1155 mirrors IERC1155.sol, including its IERC165 embedding.
type IERC1155 interface {
	erc165.IERC165
	BalanceOf(ctx context.Context, contract, account *types.Address, id *big.Int, block string) (*big.Int, error)
	BalanceOfBatch(ctx context.Context, contract *types.Address, accounts []*types.Address, ids []*big.Int, block string) ([]*big.Int, error)
	IsApprovedForAll(ctx context.Context, contract, account, operator *types.Address, block string) (bool, error)
	SetApprovalForAll(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SafeTransferFrom(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SafeBatchTransferFrom(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

var _ IERC1155 = (*ERC1155)(nil)

// IERC1155MetadataURI mirrors IERC1155MetadataURI.sol, including its IERC1155 embedding.
type IERC1155MetadataURI interface {
	IERC1155
	URI(ctx context.Context, contract *types.Address, id *big.Int, block string) (string, error)
}

var _ IERC1155MetadataURI = (*ERC1155)(nil)

// ERC1155 calls any ERC-1155 token contract over a Client; it holds no
// contract address, so one instance serves every collection. SupportsInterface
// is promoted from the embedded *erc165.ERC165, mirroring IERC1155 is IERC165.
type ERC1155 struct {
	*erc165.ERC165

	c rpc.Client
}

// New returns an ERC1155 calling collections via cli.
func New(cli rpc.Client) *ERC1155 {
	return &ERC1155{
		c:      cli,
		ERC165: erc165.New(cli),
	}
}

// URI returns id's URI on contract.
func (e *ERC1155) URI(ctx context.Context, contract *types.Address, id *big.Int, block string) (string, error) {
	calldata, err := EncodeURI(id)
	if err != nil {
		return "", fmt.Errorf("erc1155: uri: %w", err)
	}
	data, err := e.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return "", fmt.Errorf("erc1155: uri: %w", err)
	}
	s, err := DecodeURI(data)
	if err != nil {
		return "", fmt.Errorf("erc1155: uri: %w", err)
	}
	return s, nil
}

// BalanceOf returns account's balance of id on contract.
func (e *ERC1155) BalanceOf(ctx context.Context, contract, account *types.Address, id *big.Int, block string) (*big.Int, error) {
	calldata, err := EncodeBalanceOf(account, id)
	if err != nil {
		return nil, fmt.Errorf("erc1155: balanceOf: %w", err)
	}
	data, err := e.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc1155: balanceOf: %w", err)
	}
	n, err := DecodeBalanceOf(data)
	if err != nil {
		return nil, fmt.Errorf("erc1155: balanceOf: %w", err)
	}
	return n, nil
}

// BalanceOfBatch returns each accounts[i]'s balance of ids[i] on contract, in the same order.
func (e *ERC1155) BalanceOfBatch(ctx context.Context, contract *types.Address, accounts []*types.Address, ids []*big.Int, block string) ([]*big.Int, error) {
	calldata, err := EncodeBalanceOfBatch(accounts, ids)
	if err != nil {
		return nil, fmt.Errorf("erc1155: balanceOfBatch: %w", err)
	}
	data, err := e.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc1155: balanceOfBatch: %w", err)
	}
	amounts, err := DecodeBalanceOfBatch(data)
	if err != nil {
		return nil, fmt.Errorf("erc1155: balanceOfBatch: %w", err)
	}
	return amounts, nil
}

// IdBalances returns account's balance of each id in ids on contract, via balanceOfBatch's native
// cross-id batching (account repeated once per id).
func (e *ERC1155) IdBalances(ctx context.Context, contract, account *types.Address, ids []*big.Int, block string) ([]IdBalance, error) {
	accounts := make([]*types.Address, len(ids))
	for i := range accounts {
		accounts[i] = account
	}

	calldata, err := EncodeBalanceOfBatch(accounts, ids)
	if err != nil {
		return nil, fmt.Errorf("erc1155: idBalances: %w", err)
	}
	data, err := e.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc1155: idBalances: %w", err)
	}
	balances, err := DecodeIdBalances(ids, data)
	if err != nil {
		return nil, fmt.Errorf("erc1155: idBalances: %w", err)
	}
	return balances, nil
}

// IdBalancePairs returns each accounts[i]'s balance of ids[i] on contract, via balanceOfBatch's
// native cross-id batching; accounts and ids are matched pairwise by index.
func (e *ERC1155) IdBalancePairs(ctx context.Context, contract *types.Address, accounts []*types.Address, ids []*big.Int, block string) ([]IdBalancePair, error) {
	calldata, err := EncodeBalanceOfBatch(accounts, ids)
	if err != nil {
		return nil, fmt.Errorf("erc1155: idBalancePairs: %w", err)
	}
	data, err := e.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc1155: idBalancePairs: %w", err)
	}
	pairs, err := DecodeIdBalancePairs(accounts, ids, data)
	if err != nil {
		return nil, fmt.Errorf("erc1155: idBalancePairs: %w", err)
	}
	return pairs, nil
}

// AccountIdBalances batches balanceOf(account, id) across each contract in standards into one
// call, via the Multicall3 contract deployed at multicall.
func (e *ERC1155) AccountIdBalances(ctx context.Context, multicall, account *types.Address, id *big.Int, contracts []*types.Address, allowFailure bool, block string) (*AccountIdBalances, error) {
	calldata, err := EncodeAccountIdBalances(account, id, contracts, allowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc1155: accountIdBalances: %w", err)
	}

	data, err := e.call(ctx, nil, multicall, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc1155: accountIdBalances: %w", err)
	}

	aib, err := DecodeAccountIdBalances(account, id, contracts, data)
	if err != nil {
		return nil, fmt.Errorf("erc1155: accountIdBalances: %w", err)
	}
	return aib, nil
}

// ApprovalForAllWithBalance batches isApprovedForAll(account, operator) and both their
// balanceOf(id) results into one call, via the Multicall3 contract deployed at multicall.
func (e *ERC1155) ApprovalForAllWithBalance(ctx context.Context, multicall, contract, account, operator *types.Address, id *big.Int, allowFailure bool, block string) (*ApprovalForAllWithBalance, error) {
	calldata, err := EncodeApprovalForAllWithBalance(contract, account, operator, id, allowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc1155: approvalForAllWithBalance: %w", err)
	}

	data, err := e.call(ctx, nil, multicall, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc1155: approvalForAllWithBalance: %w", err)
	}

	afawb, err := DecodeApprovalForAllWithBalance(account, operator, id, data)
	if err != nil {
		return nil, fmt.Errorf("erc1155: approvalForAllWithBalance: %w", err)
	}
	return afawb, nil
}

// IsApprovedForAll reports whether operator may manage all of account's tokens on contract.
func (e *ERC1155) IsApprovedForAll(ctx context.Context, contract, account, operator *types.Address, block string) (bool, error) {
	calldata, err := EncodeIsApprovedForAll(account, operator)
	if err != nil {
		return false, fmt.Errorf("erc1155: isApprovedForAll: %w", err)
	}
	data, err := e.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return false, fmt.Errorf("erc1155: isApprovedForAll: %w", err)
	}
	ok, err := DecodeIsApprovedForAll(data)
	if err != nil {
		return false, fmt.Errorf("erc1155: isApprovedForAll: %w", err)
	}
	return ok, nil
}

// CreateSetApprovalForAll builds an unsigned setApprovalForAll(operator, approved) transaction via b.
func (e *ERC1155) CreateSetApprovalForAll(ctx context.Context, b tx.Builder, contract, from, operator *types.Address, approved bool) (core.Transaction, error) {
	calldata, err := EncodeSetApprovalForAll(operator, approved)
	if err != nil {
		return nil, fmt.Errorf("erc1155: createSetApprovalForAll: %w", err)
	}

	unsigned, err := b.Build(from, contract, nil, calldata).Pack(ctx, e.c)
	if err != nil {
		return nil, fmt.Errorf("erc1155: createSetApprovalForAll: %w", e.wrapRevertError(err))
	}
	return unsigned, nil
}

// SetApprovalForAll broadcasts an already-signed setApprovalForAll transaction.
func (e *ERC1155) SetApprovalForAll(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("erc1155: setApprovalForAll: %w", err)
	}
	hash, err := tx.Broadcast(ctx, e.c, raw)
	if err != nil {
		return nil, fmt.Errorf("erc1155: setApprovalForAll: %w", err)
	}
	return hash, nil
}

// CreateSafeTransferFrom builds an unsigned safeTransferFrom(from, to, id, value, data) transaction via b, sent by sender.
func (e *ERC1155) CreateSafeTransferFrom(ctx context.Context, b tx.Builder, contract, sender, from, to *types.Address, id, value *big.Int, callData []byte) (core.Transaction, error) {
	calldata, err := EncodeSafeTransferFrom(from, to, id, value, callData)
	if err != nil {
		return nil, fmt.Errorf("erc1155: createSafeTransferFrom: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, e.c)
	if err != nil {
		return nil, fmt.Errorf("erc1155: createSafeTransferFrom: %w", e.wrapRevertError(err))
	}
	return unsigned, nil
}

// SafeTransferFrom broadcasts an already-signed safeTransferFrom transaction.
func (e *ERC1155) SafeTransferFrom(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("erc1155: safeTransferFrom: %w", err)
	}
	hash, err := tx.Broadcast(ctx, e.c, raw)
	if err != nil {
		return nil, fmt.Errorf("erc1155: safeTransferFrom: %w", err)
	}
	return hash, nil
}

// CreateSafeBatchTransferFrom builds an unsigned safeBatchTransferFrom(from, to, ids, values, data) transaction via b, sent by sender.
func (e *ERC1155) CreateSafeBatchTransferFrom(ctx context.Context, b tx.Builder, contract, sender, from, to *types.Address, ids, values []*big.Int, callData []byte) (core.Transaction, error) {
	calldata, err := EncodeSafeBatchTransferFrom(from, to, ids, values, callData)
	if err != nil {
		return nil, fmt.Errorf("erc1155: createSafeBatchTransferFrom: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, e.c)
	if err != nil {
		return nil, fmt.Errorf("erc1155: createSafeBatchTransferFrom: %w", e.wrapRevertError(err))
	}
	return unsigned, nil
}

// SafeBatchTransferFrom broadcasts an already-signed safeBatchTransferFrom transaction.
func (e *ERC1155) SafeBatchTransferFrom(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("erc1155: safeBatchTransferFrom: %w", err)
	}
	hash, err := tx.Broadcast(ctx, e.c, raw)
	if err != nil {
		return nil, fmt.Errorf("erc1155: safeBatchTransferFrom: %w", err)
	}
	return hash, nil
}

// call eth_calls contract with calldata (from msg.sender from, if non-nil), decoding a reverted custom error to its sentinel.
func (e *ERC1155) call(ctx context.Context, from, contract *types.Address, calldata []byte, block string) ([]byte, error) {
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

// wrapRevertError maps a known IERC1155Errors custom error in err to its sentinel.
func (_ *ERC1155) wrapRevertError(err error) error {
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
	case errMissingApprovalForAll.Selector():
		return fmt.Errorf("%w: %w", ErrMissingApprovalForAll, err)
	case errInvalidApprover.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidApprover, err)
	case errInvalidOperator.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidOperator, err)
	case errInvalidArrayLength.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidArrayLength, err)
	default:
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}
}
