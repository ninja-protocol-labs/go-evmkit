package v4

import (
	"context"
	"fmt"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

// IRouter mirrors a V4Router-based deployment's executeActions(bytes) entry point, which unlocks PoolManager and dispatches the encoded actions.
type IRouter interface {
	CallExecuteActions(ctx context.Context, opts *bind.CallOptions, req *ExecuteActionsRequest) error

	ExecuteActions(ctx context.Context, opts *bind.TransactOptions, req *ExecuteActionsRequest) (core.Transaction, *types.Hash, error)
}

var _ IRouter = (*Router)(nil)

// Router calls any V4Router-based deployment exposing executeActions(bytes) over a Client; it holds no contract address.
type Router struct {
	c rpc.Client
}

// NewRouter returns a Router calling deployments via cli.
func NewRouter(cli rpc.Client) *Router {
	return &Router{
		c: cli,
	}
}

// CallExecuteActions runs executeActions via eth_call as opts.From with req.NativeAmount as msg.value, returning nil or the revert; opts.Value must be nil.
func (r *Router) CallExecuteActions(ctx context.Context, opts *bind.CallOptions, req *ExecuteActionsRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("uniswapv4: callExecuteActions: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return fmt.Errorf("uniswapv4: callExecuteActions: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeExecuteActions(req.Actions, req.Params)
	if err != nil {
		return fmt.Errorf("uniswapv4: callExecuteActions: %w", err)
	}

	callOpts := bind.NewCallOptions()
	if opts != nil {
		*callOpts = *opts
	}
	callOpts.Value = req.NativeAmount

	if _, err := bind.Call(ctx, r.c, callOpts, req.Contract, calldata); err != nil {
		return fmt.Errorf("uniswapv4: callExecuteActions: %w", mapRouterRevertError(err))
	}

	return nil
}

// ExecuteActions sends executeActions(abi.encode(actions, params)) on req.Contract, signed by opts.Signer with req.NativeAmount (nil sends none) as msg.value; opts.Value must be nil.
func (r *Router) ExecuteActions(ctx context.Context, opts *bind.TransactOptions, req *ExecuteActionsRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: executeActions: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv4: executeActions: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeExecuteActions(req.Actions, req.Params)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: executeActions: %w", err)
	}

	var txOpts *bind.TransactOptions
	if opts != nil {
		o := *opts
		o.Value = req.NativeAmount
		txOpts = &o
	}

	signed, hash, err := bind.Transact(ctx, r.c, txOpts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: executeActions: %w", mapRouterRevertError(err))
	}

	return signed, hash, nil
}
