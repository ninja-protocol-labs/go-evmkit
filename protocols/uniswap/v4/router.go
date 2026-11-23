package v4

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
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
		return fmt.Errorf("uniswapv4: callExecuteActions: %w", r.wrapRevertError(err))
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
		return signed, nil, fmt.Errorf("uniswapv4: executeActions: %w", r.wrapRevertError(err))
	}

	return signed, hash, nil
}

// wrapRevertError maps an IV4Router, BaseActionsRouter or DeltaResolver custom error in err to its sentinel; anything else (PoolManager reverts bubbling out of unlock) falls through to PoolManager's mapper.
func (_ *Router) wrapRevertError(err error) error {
	rpcErr, ok := errors.AsType[*rpc.ResponseError](err)
	if !ok || len(rpcErr.Data) == 0 {
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
	case errV4TooLittleReceived.Selector():
		return fmt.Errorf("%w: %w", ErrV4TooLittleReceived, err)
	case errV4TooMuchRequested.Selector():
		return fmt.Errorf("%w: %w", ErrV4TooMuchRequested, err)
	case errV4TooLittleReceivedPerHop.Selector():
		return fmt.Errorf("%w: %w", ErrV4TooLittleReceivedPerHop, err)
	case errV4TooMuchRequestedPerHop.Selector():
		return fmt.Errorf("%w: %w", ErrV4TooMuchRequestedPerHop, err)
	case errV4TooLittleReceivedPerHopSingle.Selector():
		return fmt.Errorf("%w: %w", ErrV4TooLittleReceivedPerHopSingle, err)
	case errV4TooMuchRequestedPerHopSingle.Selector():
		return fmt.Errorf("%w: %w", ErrV4TooMuchRequestedPerHopSingle, err)
	case errInvalidHopPriceLength.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidHopPriceLength, err)
	case errV4ExactOutputUnfilled.Selector():
		return fmt.Errorf("%w: %w", ErrV4ExactOutputUnfilled, err)
	case errInputLengthMismatch.Selector():
		return fmt.Errorf("%w: %w", ErrInputLengthMismatch, err)
	case errUnsupportedAction.Selector():
		return fmt.Errorf("%w: %w", ErrUnsupportedAction, err)
	case errDeltaNotPositive.Selector():
		return fmt.Errorf("%w: %w", ErrDeltaNotPositive, err)
	case errDeltaNotNegative.Selector():
		return fmt.Errorf("%w: %w", ErrDeltaNotNegative, err)
	case errInsufficientBalance.Selector():
		return fmt.Errorf("%w: %w", ErrInsufficientBalance, err)
	default:
		return (*PoolManager)(nil).wrapRevertError(err)
	}
}
