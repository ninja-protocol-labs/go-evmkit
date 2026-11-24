// Package universalrouter encodes, decodes and executes Uniswap Universal Router 2.1 command plans.
package universalrouter

import (
	"context"
	"fmt"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

// IUniversalRouter mirrors IUniversalRouter plus RouteSigner's public noncesUsed getter.
type IUniversalRouter interface {
	CallExecute(ctx context.Context, opts *bind.CallOptions, req *ExecuteRequest) error
	CallExecuteSigned(ctx context.Context, opts *bind.CallOptions, req *ExecuteSignedRequest) error
	NoncesUsed(ctx context.Context, opts *bind.CallOptions, req *NoncesUsedRequest) (bool, error)
	SignedRouteContext(ctx context.Context, opts *bind.CallOptions, req *SignedRouteContextRequest) (*RouteContext, error)

	Execute(ctx context.Context, opts *bind.TransactOptions, req *ExecuteRequest) (core.Transaction, *types.Hash, error)
	ExecuteSigned(ctx context.Context, opts *bind.TransactOptions, req *ExecuteSignedRequest) (core.Transaction, *types.Hash, error)
}

var _ IUniversalRouter = (*UniversalRouter)(nil)

// UniversalRouter calls any Universal Router 2.1 deployment over a Client; it holds no contract address.
type UniversalRouter struct {
	c rpc.Client
}

// New returns a UniversalRouter calling deployments via cli.
func New(cli rpc.Client) *UniversalRouter {
	return &UniversalRouter{
		c: cli,
	}
}

// CallExecute runs req.Plan via eth_call as opts.From with req.NativeAmount as msg.value, returning nil or the revert; opts.Value must be nil.
func (u *UniversalRouter) CallExecute(ctx context.Context, opts *bind.CallOptions, req *ExecuteRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("universalrouter: callExecute: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return fmt.Errorf("universalrouter: callExecute: %w: value must be nil", bind.ErrInvalidOptions)
	}

	commands, inputs, err := req.Plan.Encode()
	if err != nil {
		return fmt.Errorf("universalrouter: callExecute: %w", err)
	}

	var calldata []byte
	if req.Deadline == nil {
		calldata, err = EncodeExecute(commands, inputs)
	} else {
		calldata, err = EncodeExecuteWithDeadline(commands, inputs, req.Deadline)
	}
	if err != nil {
		return fmt.Errorf("universalrouter: callExecute: %w", err)
	}

	callOpts := bind.NewCallOptions()
	if opts != nil {
		*callOpts = *opts
	}
	callOpts.Value = req.NativeAmount

	if _, err := bind.Call(ctx, u.c, callOpts, req.Contract, calldata); err != nil {
		return fmt.Errorf("universalrouter: callExecute: %w", MapRevertError(err))
	}

	return nil
}

// CallExecuteSigned runs executeSigned via eth_call as opts.From with req.NativeAmount as msg.value, returning nil or the revert; opts.Value must be nil.
func (u *UniversalRouter) CallExecuteSigned(ctx context.Context, opts *bind.CallOptions, req *ExecuteSignedRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("universalrouter: callExecuteSigned: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return fmt.Errorf("universalrouter: callExecuteSigned: %w: value must be nil", bind.ErrInvalidOptions)
	}

	commands, inputs, err := req.Plan.Encode()
	if err != nil {
		return fmt.Errorf("universalrouter: callExecuteSigned: %w", err)
	}

	calldata, err := EncodeExecuteSigned(commands, inputs, req.Intent, req.Data, req.VerifySender, req.Nonce, req.Signature, req.Deadline)
	if err != nil {
		return fmt.Errorf("universalrouter: callExecuteSigned: %w", err)
	}

	callOpts := bind.NewCallOptions()
	if opts != nil {
		*callOpts = *opts
	}
	callOpts.Value = req.NativeAmount

	if _, err := bind.Call(ctx, u.c, callOpts, req.Contract, calldata); err != nil {
		return fmt.Errorf("universalrouter: callExecuteSigned: %w", MapRevertError(err))
	}

	return nil
}

// NoncesUsed reports whether req.User's executeSigned nonce req.Nonce has been consumed on req.Contract.
func (u *UniversalRouter) NoncesUsed(ctx context.Context, opts *bind.CallOptions, req *NoncesUsedRequest) (bool, error) {
	if err := req.Validate(); err != nil {
		return false, fmt.Errorf("universalrouter: noncesUsed: %w", err)
	}

	calldata, err := EncodeNoncesUsed(req.User, req.Nonce)
	if err != nil {
		return false, fmt.Errorf("universalrouter: noncesUsed: %w", err)
	}

	data, err := bind.Call(ctx, u.c, opts, req.Contract, calldata)
	if err != nil {
		return false, fmt.Errorf("universalrouter: noncesUsed: %w", MapRevertError(err))
	}

	used, err := DecodeNoncesUsed(data)
	if err != nil {
		return false, fmt.Errorf("universalrouter: noncesUsed: %w", err)
	}

	return used, nil
}

// SignedRouteContext returns req.Contract's signed route context, all zero outside an executeSigned call.
func (u *UniversalRouter) SignedRouteContext(ctx context.Context, opts *bind.CallOptions, req *SignedRouteContextRequest) (*RouteContext, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("universalrouter: signedRouteContext: %w", err)
	}

	calldata, err := EncodeSignedRouteContext()
	if err != nil {
		return nil, fmt.Errorf("universalrouter: signedRouteContext: %w", err)
	}

	data, err := bind.Call(ctx, u.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: signedRouteContext: %w", MapRevertError(err))
	}

	out, err := DecodeSignedRouteContext(data)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: signedRouteContext: %w", err)
	}

	return out, nil
}

// Execute sends execute(commands, inputs[, deadline]) for req.Plan on req.Contract, signed by opts.Signer with req.NativeAmount (nil sends none) as msg.value; opts.Value must be nil.
func (u *UniversalRouter) Execute(ctx context.Context, opts *bind.TransactOptions, req *ExecuteRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("universalrouter: execute: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("universalrouter: execute: %w: value must be nil", bind.ErrInvalidOptions)
	}

	commands, inputs, err := req.Plan.Encode()
	if err != nil {
		return nil, nil, fmt.Errorf("universalrouter: execute: %w", err)
	}

	var calldata []byte
	if req.Deadline == nil {
		calldata, err = EncodeExecute(commands, inputs)
	} else {
		calldata, err = EncodeExecuteWithDeadline(commands, inputs, req.Deadline)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("universalrouter: execute: %w", err)
	}

	var txOpts *bind.TransactOptions
	if opts != nil {
		o := *opts
		o.Value = req.NativeAmount
		txOpts = &o
	}

	signed, hash, err := bind.Transact(ctx, u.c, txOpts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("universalrouter: execute: %w", MapRevertError(err))
	}

	return signed, hash, nil
}

// ExecuteSigned sends executeSigned for req.Plan on req.Contract, signed by opts.Signer with req.NativeAmount (nil sends none) as msg.value; opts.Value must be nil.
func (u *UniversalRouter) ExecuteSigned(ctx context.Context, opts *bind.TransactOptions, req *ExecuteSignedRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("universalrouter: executeSigned: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("universalrouter: executeSigned: %w: value must be nil", bind.ErrInvalidOptions)
	}

	commands, inputs, err := req.Plan.Encode()
	if err != nil {
		return nil, nil, fmt.Errorf("universalrouter: executeSigned: %w", err)
	}

	calldata, err := EncodeExecuteSigned(commands, inputs, req.Intent, req.Data, req.VerifySender, req.Nonce, req.Signature, req.Deadline)
	if err != nil {
		return nil, nil, fmt.Errorf("universalrouter: executeSigned: %w", err)
	}

	var txOpts *bind.TransactOptions
	if opts != nil {
		o := *opts
		o.Value = req.NativeAmount
		txOpts = &o
	}

	signed, hash, err := bind.Transact(ctx, u.c, txOpts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("universalrouter: executeSigned: %w", MapRevertError(err))
	}

	return signed, hash, nil
}
