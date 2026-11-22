// Package weth is a client for WETH9, the de facto contract for wrapping a chain's native asset
// into an ERC-20 token and back.
package weth

import (
	"context"
	"fmt"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
)

// IWETH mirrors WETH9, embedding erc20 for the shared ERC-20 surface.
type IWETH interface {
	erc20.IERC20Metadata

	Deposit(ctx context.Context, opts *bind.TransactOptions, req *DepositRequest) (core.Transaction, *types.Hash, error)
	Withdraw(ctx context.Context, opts *bind.TransactOptions, req *WithdrawRequest) (core.Transaction, *types.Hash, error)
}

var _ IWETH = (*WETH)(nil)

// WETH calls any WETH9-shaped contract over a Client; ERC-20 methods are promoted from the
// embedded *erc20.ERC20.
type WETH struct {
	*erc20.ERC20

	c rpc.Client
}

// New returns a WETH calling contracts via cli.
func New(cli rpc.Client) *WETH {
	return &WETH{
		ERC20: erc20.New(cli),
		c:     cli,
	}
}

// Deposit sends deposit() on req.Contract with msg.value set to req.Amount, signed by
// opts.Signer; opts.Value must be nil.
func (w *WETH) Deposit(ctx context.Context, opts *bind.TransactOptions, req *DepositRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("weth: deposit: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("weth: deposit: %w: value must be nil", bind.ErrInvalidOptions)
	}

	var txOpts *bind.TransactOptions
	if opts != nil {
		o := *opts
		o.Value = req.Amount
		txOpts = &o
	}

	signed, hash, err := bind.Transact(ctx, w.c, txOpts, req.Contract, EncodeDeposit())
	if err != nil {
		return signed, nil, fmt.Errorf("weth: deposit: %w", err)
	}

	return signed, hash, nil
}

// Withdraw sends withdraw(amount) on req.Contract, signed by opts.Signer, unwrapping req.Amount
// back to native.
func (w *WETH) Withdraw(ctx context.Context, opts *bind.TransactOptions, req *WithdrawRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("weth: withdraw: %w", err)
	}

	calldata, err := EncodeWithdraw(req.Amount)
	if err != nil {
		return nil, nil, fmt.Errorf("weth: withdraw: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, w.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("weth: withdraw: %w", err)
	}

	return signed, hash, nil
}
