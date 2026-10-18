// Package weth encodes calls to and decodes results from WETH9, the de facto standard
// contract for wrapping a chain's native asset into an ERC-20 token and back.
package weth

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

// IWETH mirrors WETH9, embedding erc20 for the shared ERC-20 surface.
type IWETH interface {
	erc20.IERC20Metadata

	Deposit(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Withdraw(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

var _ IWETH = (*WETH)(nil)

// WETH calls any WETH9-shaped contract over a Client; it holds no contract address.
type WETH struct {
	*erc20.ERC20

	c rpc.Client
}

// New returns a WETH calling contracts via c.
func New(c rpc.Client) *WETH {
	return &WETH{
		ERC20: erc20.New(c),
		c:     c,
	}
}

// CreateDeposit builds an unsigned deposit() transaction via b, wrapping value wei of sender's native balance.
func (w *WETH) CreateDeposit(ctx context.Context, b tx.Builder, contract, sender *types.Address, value *big.Int) (core.Transaction, error) {
	unsigned, err := b.Build(sender, contract, value, EncodeDeposit()).Pack(ctx, w.c)
	if err != nil {
		return nil, fmt.Errorf("weth: createDeposit: %w", err)
	}
	return unsigned, nil
}

// Deposit broadcasts an already-signed deposit transaction.
func (w *WETH) Deposit(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("weth: deposit: %w", err)
	}
	hash, err := tx.Broadcast(ctx, w.c, raw)
	if err != nil {
		return nil, fmt.Errorf("weth: deposit: %w", err)
	}
	return hash, nil
}

// CreateWithdraw builds an unsigned withdraw(wad) transaction via b, unwrapping wad of sender's contract balance.
func (w *WETH) CreateWithdraw(ctx context.Context, b tx.Builder, contract, sender *types.Address, wad *big.Int) (core.Transaction, error) {
	calldata, err := EncodeWithdraw(wad)
	if err != nil {
		return nil, fmt.Errorf("weth: createWithdraw: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, w.c)
	if err != nil {
		return nil, fmt.Errorf("weth: createWithdraw: %w", err)
	}
	return unsigned, nil
}

// Withdraw broadcasts an already-signed withdraw transaction.
func (w *WETH) Withdraw(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("weth: withdraw: %w", err)
	}
	hash, err := tx.Broadcast(ctx, w.c, raw)
	if err != nil {
		return nil, fmt.Errorf("weth: withdraw: %w", err)
	}
	return hash, nil
}
