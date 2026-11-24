// Package universalrouter encodes, decodes and executes Uniswap Universal Router 2.1 command plans.
package universalrouter

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/protocols/permit2"
	v2 "github.com/ninja-protocol-labs/go-evmkit/protocols/uniswap/v2"
	v3 "github.com/ninja-protocol-labs/go-evmkit/protocols/uniswap/v3"
	v4 "github.com/ninja-protocol-labs/go-evmkit/protocols/uniswap/v4"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
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
		return fmt.Errorf("universalrouter: callExecute: %w", u.wrapRevertError(err))
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
		return fmt.Errorf("universalrouter: callExecuteSigned: %w", u.wrapRevertError(err))
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
		return false, fmt.Errorf("universalrouter: noncesUsed: %w", u.wrapRevertError(err))
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
		return nil, fmt.Errorf("universalrouter: signedRouteContext: %w", u.wrapRevertError(err))
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
		return signed, nil, fmt.Errorf("universalrouter: execute: %w", u.wrapRevertError(err))
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
		return signed, nil, fmt.Errorf("universalrouter: executeSigned: %w", u.wrapRevertError(err))
	}

	return signed, hash, nil
}

// wrapRevertError maps a revert in err to its UR, Permit2, v4, v3 or v2 sentinel, else ErrUnknown.
func (u *UniversalRouter) wrapRevertError(err error) error {
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
	case errExecutionFailed.Selector():
		vals, decodeErr := errExecutionFailed.Decode(data)
		if decodeErr != nil || len(vals) != 2 {
			return fmt.Errorf("%w: %w", ErrExecutionFailed, err)
		}

		commandIndex, ok := vals[0].(*big.Int)
		if !ok {
			return fmt.Errorf("%w: %w", ErrExecutionFailed, err)
		}

		message, ok := vals[1].([]byte)
		if !ok {
			return fmt.Errorf("%w: %w", ErrExecutionFailed, err)
		}

		var cause error
		if len(message) >= 4 {
			inner := &rpc.ResponseError{Code: rpcErr.Code, Message: rpcErr.Message, Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(message) + `"`)}
			cause = u.wrapRevertError(inner)
		}

		return fmt.Errorf("%w: %w", NewExecutionFailedError(commandIndex, message, cause), err)
	case errETHNotAccepted.Selector():
		return fmt.Errorf("%w: %w", ErrETHNotAccepted, err)
	case errTransactionDeadlinePassed.Selector():
		return fmt.Errorf("%w: %w", ErrTransactionDeadlinePassed, err)
	case errLengthMismatch.Selector():
		return fmt.Errorf("%w: %w", ErrLengthMismatch, err)
	case errInvalidEthSender.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidEthSender, err)
	case errContractLocked.Selector():
		return fmt.Errorf("%w: %w", ErrContractLocked, err)
	case errNonceAlreadyUsed.Selector():
		return fmt.Errorf("%w: %w", ErrNonceAlreadyUsed, err)
	case errInvalidCommandType.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidCommandType, err)
	case errBalanceTooLow.Selector():
		return fmt.Errorf("%w: %w", ErrBalanceTooLow, err)
	case errInsufficientToken.Selector():
		return fmt.Errorf("%w: %w", ErrInsufficientToken, err)
	case errInsufficientETH.Selector():
		return fmt.Errorf("%w: %w", ErrInsufficientETH, err)
	case errInvalidPortion.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidPortion, err)
	case errInvalidBips.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidBips, err)
	case errFromAddressIsNotOwner.Selector():
		return fmt.Errorf("%w: %w", ErrFromAddressIsNotOwner, err)
	case errInvalidAction.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidAction, err)
	case errOnlyMintAllowed.Selector():
		return fmt.Errorf("%w: %w", ErrOnlyMintAllowed, err)
	case errNotAuthorizedForToken.Selector():
		return fmt.Errorf("%w: %w", ErrNotAuthorizedForToken, err)
	case errSafeERC20FailedOperation.Selector():
		return fmt.Errorf("%w: %w", ErrSafeERC20FailedOperation, err)
	case errAddressEmptyCode.Selector():
		return fmt.Errorf("%w: %w", ErrAddressEmptyCode, err)
	case errFailedInnerCall.Selector():
		return fmt.Errorf("%w: %w", ErrFailedInnerCall, err)
	case errInvalidReserves.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidReserves, err)
	case errInvalidPath.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidPath, err)
	case errSliceOutOfBounds.Selector():
		return fmt.Errorf("%w: %w", ErrSliceOutOfBounds, err)
	case errV2TooLittleReceived.Selector():
		return fmt.Errorf("%w: %w", ErrV2TooLittleReceived, err)
	case errV2TooMuchRequested.Selector():
		return fmt.Errorf("%w: %w", ErrV2TooMuchRequested, err)
	case errV2InvalidPath.Selector():
		return fmt.Errorf("%w: %w", ErrV2InvalidPath, err)
	case errV2TooLittleReceivedPerHop.Selector():
		return fmt.Errorf("%w: %w", ErrV2TooLittleReceivedPerHop, err)
	case errV2InvalidHopPriceLength.Selector():
		return fmt.Errorf("%w: %w", ErrV2InvalidHopPriceLength, err)
	case errV3InvalidSwap.Selector():
		return fmt.Errorf("%w: %w", ErrV3InvalidSwap, err)
	case errV3TooLittleReceived.Selector():
		return fmt.Errorf("%w: %w", ErrV3TooLittleReceived, err)
	case errV3TooMuchRequested.Selector():
		return fmt.Errorf("%w: %w", ErrV3TooMuchRequested, err)
	case errV3InvalidAmountOut.Selector():
		return fmt.Errorf("%w: %w", ErrV3InvalidAmountOut, err)
	case errV3InvalidCaller.Selector():
		return fmt.Errorf("%w: %w", ErrV3InvalidCaller, err)
	case errV3TooLittleReceivedPerHop.Selector():
		return fmt.Errorf("%w: %w", ErrV3TooLittleReceivedPerHop, err)
	case errV3TooMuchRequestedPerHop.Selector():
		return fmt.Errorf("%w: %w", ErrV3TooMuchRequestedPerHop, err)
	case errV3HopPriceAndPathLengthMismatch.Selector():
		return fmt.Errorf("%w: %w", ErrV3HopPriceAndPathLengthMismatch, err)
	case errECDSAInvalidSignature.Selector():
		return fmt.Errorf("%w: %w", ErrECDSAInvalidSignature, err)
	case errECDSAInvalidSignatureLength.Selector():
		return fmt.Errorf("%w: %w", ErrECDSAInvalidSignatureLength, err)
	case errECDSAInvalidSignatureS.Selector():
		return fmt.Errorf("%w: %w", ErrECDSAInvalidSignatureS, err)
	}

	if wrapped := permit2.MapRevertError(err); !errors.Is(wrapped, permit2.ErrUnknown) {
		return wrapped
	}

	if wrapped := v4.MapRevertError(err); !errors.Is(wrapped, v4.ErrUnknown) {
		return wrapped
	}

	if _, reasonErr := abi.DecodeRevert(data); reasonErr == nil {
		if wrapped := v3.MapRevertError(err); !errors.Is(wrapped, v3.ErrReverted) {
			return wrapped
		}

		if wrapped := v2.MapRevertError(err); !errors.Is(wrapped, v2.ErrUnknown) {
			return wrapped
		}
	}

	return fmt.Errorf("%w: %w", ErrUnknown, err)
}
