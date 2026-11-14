// Package erc6909 encodes calls to and decodes results from an ERC-6909
// multi-token contract. Go types mirror the interfaces in
// solidity/interfaces/IERC6909.sol and solidity/interfaces/IERC6909Metadata.sol.
package erc6909

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

// IERC6909 mirrors IERC6909.sol, including its IERC165 embedding.
type IERC6909 interface {
	erc165.IERC165
	BalanceOf(ctx context.Context, contract, owner *types.Address, id *big.Int, block string) (*big.Int, error)
	Allowance(ctx context.Context, contract, owner, spender *types.Address, id *big.Int, block string) (*big.Int, error)
	IsOperator(ctx context.Context, contract, owner, spender *types.Address, block string) (bool, error)

	Approve(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SetOperator(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Transfer(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	TransferFrom(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

var _ IERC6909 = (*ERC6909)(nil)

// IERC6909Metadata mirrors IERC6909Metadata.sol, including its IERC6909 embedding.
type IERC6909Metadata interface {
	IERC6909
	Name(ctx context.Context, contract *types.Address, id *big.Int, block string) (string, error)
	Symbol(ctx context.Context, contract *types.Address, id *big.Int, block string) (string, error)
	Decimals(ctx context.Context, contract *types.Address, id *big.Int, block string) (uint8, error)
}

var _ IERC6909Metadata = (*ERC6909)(nil)

// ERC6909 calls any ERC-6909 token contract over a Client; it holds no
// contract address, so one instance serves every collection. SupportsInterface
// is promoted from the embedded *erc165.ERC165, mirroring IERC6909 is IERC165.
type ERC6909 struct {
	*erc165.ERC165

	c rpc.Client
}

// New returns an ERC6909 calling collections via cli.
func New(cli rpc.Client) *ERC6909 {
	return &ERC6909{
		c:      cli,
		ERC165: erc165.New(cli),
	}
}

// BalanceOf returns owner's balance of id on contract.
func (e *ERC6909) BalanceOf(ctx context.Context, contract, owner *types.Address, id *big.Int, block string) (*big.Int, error) {
	calldata, err := EncodeBalanceOf(owner, id)
	if err != nil {
		return nil, fmt.Errorf("erc6909: balanceOf: %w", err)
	}
	data, err := e.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc6909: balanceOf: %w", err)
	}
	n, err := DecodeBalanceOf(data)
	if err != nil {
		return nil, fmt.Errorf("erc6909: balanceOf: %w", err)
	}
	return n, nil
}

// Allowance returns the amount spender may spend of id on behalf of owner on contract.
func (e *ERC6909) Allowance(ctx context.Context, contract, owner, spender *types.Address, id *big.Int, block string) (*big.Int, error) {
	calldata, err := EncodeAllowance(owner, spender, id)
	if err != nil {
		return nil, fmt.Errorf("erc6909: allowance: %w", err)
	}
	data, err := e.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc6909: allowance: %w", err)
	}
	n, err := DecodeAllowance(data)
	if err != nil {
		return nil, fmt.Errorf("erc6909: allowance: %w", err)
	}
	return n, nil
}

// IsOperator reports whether spender may manage all of owner's tokens on contract.
func (e *ERC6909) IsOperator(ctx context.Context, contract, owner, spender *types.Address, block string) (bool, error) {
	calldata, err := EncodeIsOperator(owner, spender)
	if err != nil {
		return false, fmt.Errorf("erc6909: isOperator: %w", err)
	}
	data, err := e.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return false, fmt.Errorf("erc6909: isOperator: %w", err)
	}
	ok, err := DecodeIsOperator(data)
	if err != nil {
		return false, fmt.Errorf("erc6909: isOperator: %w", err)
	}
	return ok, nil
}

// Name returns id's name on contract.
func (e *ERC6909) Name(ctx context.Context, contract *types.Address, id *big.Int, block string) (string, error) {
	calldata, err := EncodeName(id)
	if err != nil {
		return "", fmt.Errorf("erc6909: name: %w", err)
	}
	data, err := e.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return "", fmt.Errorf("erc6909: name: %w", err)
	}
	s, err := DecodeName(data)
	if err != nil {
		return "", fmt.Errorf("erc6909: name: %w", err)
	}
	return s, nil
}

// Symbol returns id's symbol on contract.
func (e *ERC6909) Symbol(ctx context.Context, contract *types.Address, id *big.Int, block string) (string, error) {
	calldata, err := EncodeSymbol(id)
	if err != nil {
		return "", fmt.Errorf("erc6909: symbol: %w", err)
	}
	data, err := e.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return "", fmt.Errorf("erc6909: symbol: %w", err)
	}
	s, err := DecodeSymbol(data)
	if err != nil {
		return "", fmt.Errorf("erc6909: symbol: %w", err)
	}
	return s, nil
}

// Decimals returns id's decimals on contract.
func (e *ERC6909) Decimals(ctx context.Context, contract *types.Address, id *big.Int, block string) (uint8, error) {
	calldata, err := EncodeDecimals(id)
	if err != nil {
		return 0, fmt.Errorf("erc6909: decimals: %w", err)
	}
	data, err := e.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return 0, fmt.Errorf("erc6909: decimals: %w", err)
	}
	d, err := DecodeDecimals(data)
	if err != nil {
		return 0, fmt.Errorf("erc6909: decimals: %w", err)
	}
	return d, nil
}

// CreateApprove builds an unsigned approve(spender, id, amount) transaction via b, sent by sender.
func (e *ERC6909) CreateApprove(ctx context.Context, b tx.Builder, contract, sender, spender *types.Address, id, amount *big.Int) (core.Transaction, error) {
	calldata, err := EncodeApprove(spender, id, amount)
	if err != nil {
		return nil, fmt.Errorf("erc6909: createApprove: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, e.c)
	if err != nil {
		return nil, fmt.Errorf("erc6909: createApprove: %w", e.wrapRevertError(err))
	}
	return unsigned, nil
}

// Approve broadcasts an already-signed approve transaction.
func (e *ERC6909) Approve(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("erc6909: approve: %w", err)
	}
	hash, err := tx.Broadcast(ctx, e.c, raw)
	if err != nil {
		return nil, fmt.Errorf("erc6909: approve: %w", err)
	}
	return hash, nil
}

// CreateSetOperator builds an unsigned setOperator(spender, approved) transaction via b, sent by sender.
func (e *ERC6909) CreateSetOperator(ctx context.Context, b tx.Builder, contract, sender, spender *types.Address, approved bool) (core.Transaction, error) {
	calldata, err := EncodeSetOperator(spender, approved)
	if err != nil {
		return nil, fmt.Errorf("erc6909: createSetOperator: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, e.c)
	if err != nil {
		return nil, fmt.Errorf("erc6909: createSetOperator: %w", e.wrapRevertError(err))
	}
	return unsigned, nil
}

// SetOperator broadcasts an already-signed setOperator transaction.
func (e *ERC6909) SetOperator(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("erc6909: setOperator: %w", err)
	}
	hash, err := tx.Broadcast(ctx, e.c, raw)
	if err != nil {
		return nil, fmt.Errorf("erc6909: setOperator: %w", err)
	}
	return hash, nil
}

// CreateTransfer builds an unsigned transfer(receiver, id, amount) transaction via b, sent by sender.
func (e *ERC6909) CreateTransfer(ctx context.Context, b tx.Builder, contract, sender, receiver *types.Address, id, amount *big.Int) (core.Transaction, error) {
	calldata, err := EncodeTransfer(receiver, id, amount)
	if err != nil {
		return nil, fmt.Errorf("erc6909: createTransfer: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, e.c)
	if err != nil {
		return nil, fmt.Errorf("erc6909: createTransfer: %w", e.wrapRevertError(err))
	}
	return unsigned, nil
}

// Transfer broadcasts an already-signed transfer transaction.
func (e *ERC6909) Transfer(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("erc6909: transfer: %w", err)
	}
	hash, err := tx.Broadcast(ctx, e.c, raw)
	if err != nil {
		return nil, fmt.Errorf("erc6909: transfer: %w", err)
	}
	return hash, nil
}

// CreateTransferFrom builds an unsigned transferFrom(from, to, id, amount) transaction via b, sent
// by sender (who may be from itself, an approved spender, or an operator of from).
func (e *ERC6909) CreateTransferFrom(ctx context.Context, b tx.Builder, contract, sender, from, to *types.Address, id, amount *big.Int) (core.Transaction, error) {
	calldata, err := EncodeTransferFrom(from, to, id, amount)
	if err != nil {
		return nil, fmt.Errorf("erc6909: createTransferFrom: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, e.c)
	if err != nil {
		return nil, fmt.Errorf("erc6909: createTransferFrom: %w", e.wrapRevertError(err))
	}
	return unsigned, nil
}

// TransferFrom broadcasts an already-signed transferFrom transaction.
func (e *ERC6909) TransferFrom(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("erc6909: transferFrom: %w", err)
	}
	hash, err := tx.Broadcast(ctx, e.c, raw)
	if err != nil {
		return nil, fmt.Errorf("erc6909: transferFrom: %w", err)
	}
	return hash, nil
}

// call eth_calls contract with calldata (from msg.sender from, if non-nil), decoding a reverted custom error to its sentinel.
func (e *ERC6909) call(ctx context.Context, from, contract *types.Address, calldata []byte, block string) ([]byte, error) {
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

// wrapRevertError maps a known OpenZeppelin ERC6909 custom error in err to its sentinel.
func (_ *ERC6909) wrapRevertError(err error) error {
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
	case errInsufficientAllowance.Selector():
		return fmt.Errorf("%w: %w", ErrInsufficientAllowance, err)
	case errInvalidApprover.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidApprover, err)
	case errInvalidReceiver.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidReceiver, err)
	case errInvalidSender.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidSender, err)
	case errInvalidSpender.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidSpender, err)
	default:
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}
}
