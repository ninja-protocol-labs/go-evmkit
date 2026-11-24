package v3

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

// IRouterImmutables mirrors PeripheryImmutableState, the only state SwapRouter02 itself holds.
type IRouterImmutables interface {
	Factory(ctx context.Context, opts *bind.CallOptions, req *FactoryRequest) (*types.Address, error)
	WETH9(ctx context.Context, opts *bind.CallOptions, req *WETH9Request) (*types.Address, error)
}

// IRouterSwap mirrors IV3SwapRouter (SwapRouter02: no deadline in the params; use Multicall with a Deadline).
type IRouterSwap interface {
	CallExactInputSingle(ctx context.Context, opts *bind.CallOptions, req *ExactInputSingleRequest) (*big.Int, error)
	CallExactInput(ctx context.Context, opts *bind.CallOptions, req *ExactInputRequest) (*big.Int, error)
	CallExactOutputSingle(ctx context.Context, opts *bind.CallOptions, req *ExactOutputSingleRequest) (*big.Int, error)
	CallExactOutput(ctx context.Context, opts *bind.CallOptions, req *ExactOutputRequest) (*big.Int, error)

	ExactInputSingle(ctx context.Context, opts *bind.TransactOptions, req *ExactInputSingleRequest) (core.Transaction, *types.Hash, error)
	ExactInput(ctx context.Context, opts *bind.TransactOptions, req *ExactInputRequest) (core.Transaction, *types.Hash, error)
	ExactOutputSingle(ctx context.Context, opts *bind.TransactOptions, req *ExactOutputSingleRequest) (core.Transaction, *types.Hash, error)
	ExactOutput(ctx context.Context, opts *bind.TransactOptions, req *ExactOutputRequest) (core.Transaction, *types.Hash, error)
}

// IRouterPayments mirrors IPeripheryPaymentsWithFee.
type IRouterPayments interface {
	UnwrapWETH9(ctx context.Context, opts *bind.TransactOptions, req *RouterUnwrapWETH9Request) (core.Transaction, *types.Hash, error)
	UnwrapWETH9WithFee(ctx context.Context, opts *bind.TransactOptions, req *RouterUnwrapWETH9WithFeeRequest) (core.Transaction, *types.Hash, error)
	SweepToken(ctx context.Context, opts *bind.TransactOptions, req *RouterSweepTokenRequest) (core.Transaction, *types.Hash, error)
	SweepTokenWithFee(ctx context.Context, opts *bind.TransactOptions, req *RouterSweepTokenWithFeeRequest) (core.Transaction, *types.Hash, error)
	RefundETH(ctx context.Context, opts *bind.TransactOptions, req *RouterRefundETHRequest) (core.Transaction, *types.Hash, error)
}

// IRouterMulticall mirrors MulticallExtended.
type IRouterMulticall interface {
	CallMulticall(ctx context.Context, opts *bind.CallOptions, req *RouterMulticallRequest) ([][]byte, error)

	Multicall(ctx context.Context, opts *bind.TransactOptions, req *RouterMulticallRequest) (core.Transaction, *types.Hash, error)
}

// IRouterSelfPermit mirrors ISelfPermit.
type IRouterSelfPermit interface {
	SelfPermit(ctx context.Context, opts *bind.TransactOptions, req *SelfPermitRequest) (core.Transaction, *types.Hash, error)
	SelfPermitIfNecessary(ctx context.Context, opts *bind.TransactOptions, req *SelfPermitIfNecessaryRequest) (core.Transaction, *types.Hash, error)
	SelfPermitAllowed(ctx context.Context, opts *bind.TransactOptions, req *SelfPermitAllowedRequest) (core.Transaction, *types.Hash, error)
	SelfPermitAllowedIfNecessary(ctx context.Context, opts *bind.TransactOptions, req *SelfPermitAllowedIfNecessaryRequest) (core.Transaction, *types.Hash, error)
}

// IRouter mirrors SwapRouter02's V3 surface.
type IRouter interface {
	IRouterImmutables
	IRouterSwap
	IRouterPayments
	IRouterMulticall
	IRouterSelfPermit
}

var _ IRouter = (*Router)(nil)

// Router calls any SwapRouter02 contract over a Client; it holds no contract address, so one instance serves every deployment.
type Router struct {
	c rpc.Client
}

// NewRouter returns a Router calling routers via cli.
func NewRouter(cli rpc.Client) *Router {
	return &Router{
		c: cli,
	}
}

// Factory returns the UniswapV3Factory req.Contract routes through.
func (r *Router) Factory(ctx context.Context, opts *bind.CallOptions, req *FactoryRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: factory: %w", err)
	}

	data, err := bind.Call(ctx, r.c, opts, req.Contract, EncodeRouterFactory())
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: factory: %w", mapRouterRevertError(err))
	}

	out, err := DecodeRouterFactory(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: factory: %w", err)
	}

	return out, nil
}

// WETH9 returns req.Contract's wrapped native token.
func (r *Router) WETH9(ctx context.Context, opts *bind.CallOptions, req *WETH9Request) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: weth9: %w", err)
	}

	data, err := bind.Call(ctx, r.c, opts, req.Contract, EncodeRouterWETH9())
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: weth9: %w", mapRouterRevertError(err))
	}

	out, err := DecodeRouterWETH9(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: weth9: %w", err)
	}

	return out, nil
}

// CallExactInputSingle runs exactInputSingle via eth_call as opts.From with req.NativeAmount as msg.value, returning its result or the revert.
func (r *Router) CallExactInputSingle(ctx context.Context, opts *bind.CallOptions, req *ExactInputSingleRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactInputSingle: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, fmt.Errorf("uniswapv3: callExactInputSingle: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeExactInputSingle(req.Params.TokenIn, req.Params.TokenOut, req.Params.Fee, req.Params.Recipient, req.Params.AmountIn, req.Params.AmountOutMinimum, req.Params.SqrtPriceLimitX96)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactInputSingle: %w", err)
	}

	callOpts := bind.NewCallOptions()
	if opts != nil {
		*callOpts = *opts
	}
	callOpts.Value = req.NativeAmount

	data, err := bind.Call(ctx, r.c, callOpts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactInputSingle: %w", mapRouterRevertError(err))
	}

	amountOut, err := DecodeExactInputSingle(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactInputSingle: %w", err)
	}

	return amountOut, nil
}

// CallExactInput runs exactInput via eth_call as opts.From with req.NativeAmount as msg.value, returning its result or the revert.
func (r *Router) CallExactInput(ctx context.Context, opts *bind.CallOptions, req *ExactInputRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactInput: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, fmt.Errorf("uniswapv3: callExactInput: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeExactInput(req.Params.Path, req.Params.Recipient, req.Params.AmountIn, req.Params.AmountOutMinimum)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactInput: %w", err)
	}

	callOpts := bind.NewCallOptions()
	if opts != nil {
		*callOpts = *opts
	}
	callOpts.Value = req.NativeAmount

	data, err := bind.Call(ctx, r.c, callOpts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactInput: %w", mapRouterRevertError(err))
	}

	amountOut, err := DecodeExactInput(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactInput: %w", err)
	}

	return amountOut, nil
}

// CallExactOutputSingle runs exactOutputSingle via eth_call as opts.From with req.NativeAmount as msg.value, returning its result or the revert.
func (r *Router) CallExactOutputSingle(ctx context.Context, opts *bind.CallOptions, req *ExactOutputSingleRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactOutputSingle: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, fmt.Errorf("uniswapv3: callExactOutputSingle: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeExactOutputSingle(req.Params.TokenIn, req.Params.TokenOut, req.Params.Fee, req.Params.Recipient, req.Params.AmountOut, req.Params.AmountInMaximum, req.Params.SqrtPriceLimitX96)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactOutputSingle: %w", err)
	}

	callOpts := bind.NewCallOptions()
	if opts != nil {
		*callOpts = *opts
	}
	callOpts.Value = req.NativeAmount

	data, err := bind.Call(ctx, r.c, callOpts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactOutputSingle: %w", mapRouterRevertError(err))
	}

	amountIn, err := DecodeExactOutputSingle(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactOutputSingle: %w", err)
	}

	return amountIn, nil
}

// CallExactOutput runs exactOutput via eth_call as opts.From with req.NativeAmount as msg.value, returning its result or the revert.
func (r *Router) CallExactOutput(ctx context.Context, opts *bind.CallOptions, req *ExactOutputRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactOutput: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, fmt.Errorf("uniswapv3: callExactOutput: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeExactOutput(req.Params.Path, req.Params.Recipient, req.Params.AmountOut, req.Params.AmountInMaximum)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactOutput: %w", err)
	}

	callOpts := bind.NewCallOptions()
	if opts != nil {
		*callOpts = *opts
	}
	callOpts.Value = req.NativeAmount

	data, err := bind.Call(ctx, r.c, callOpts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactOutput: %w", mapRouterRevertError(err))
	}

	amountIn, err := DecodeExactOutput(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactOutput: %w", err)
	}

	return amountIn, nil
}

// CallMulticall runs multicall via eth_call as opts.From with req.NativeAmount as msg.value, returning its result or the revert.
func (r *Router) CallMulticall(ctx context.Context, opts *bind.CallOptions, req *RouterMulticallRequest) ([][]byte, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: callMulticall: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, fmt.Errorf("uniswapv3: callMulticall: %w: value must be nil", bind.ErrInvalidOptions)
	}

	var calldata []byte
	var err error
	if req.Deadline != nil {
		calldata, err = EncodeRouterMulticallWithDeadline(req.Deadline, req.Data)
	} else {
		calldata, err = EncodeRouterMulticall(req.Data)
	}
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callMulticall: %w", err)
	}

	callOpts := bind.NewCallOptions()
	if opts != nil {
		*callOpts = *opts
	}
	callOpts.Value = req.NativeAmount

	data, err := bind.Call(ctx, r.c, callOpts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callMulticall: %w", mapRouterRevertError(err))
	}

	results, err := DecodeRouterMulticall(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callMulticall: %w", err)
	}

	return results, nil
}

// ExactInputSingle swaps exactly Params.AmountIn of TokenIn for as much TokenOut as possible in one pool, signed by opts.Signer with req.NativeAmount (nil sends none) as msg.value.
func (r *Router) ExactInputSingle(ctx context.Context, opts *bind.TransactOptions, req *ExactInputSingleRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: exactInputSingle: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv3: exactInputSingle: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeExactInputSingle(req.Params.TokenIn, req.Params.TokenOut, req.Params.Fee, req.Params.Recipient, req.Params.AmountIn, req.Params.AmountOutMinimum, req.Params.SqrtPriceLimitX96)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: exactInputSingle: %w", err)
	}

	var txOpts *bind.TransactOptions
	if opts != nil {
		o := *opts
		o.Value = req.NativeAmount
		txOpts = &o
	}

	signed, hash, err := bind.Transact(ctx, r.c, txOpts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: exactInputSingle: %w", mapRouterRevertError(err))
	}

	return signed, hash, nil
}

// ExactInput swaps exactly Params.AmountIn along Params.Path (see EncodePath), signed by opts.Signer with req.NativeAmount (nil sends none) as msg.value.
func (r *Router) ExactInput(ctx context.Context, opts *bind.TransactOptions, req *ExactInputRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: exactInput: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv3: exactInput: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeExactInput(req.Params.Path, req.Params.Recipient, req.Params.AmountIn, req.Params.AmountOutMinimum)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: exactInput: %w", err)
	}

	var txOpts *bind.TransactOptions
	if opts != nil {
		o := *opts
		o.Value = req.NativeAmount
		txOpts = &o
	}

	signed, hash, err := bind.Transact(ctx, r.c, txOpts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: exactInput: %w", mapRouterRevertError(err))
	}

	return signed, hash, nil
}

// ExactOutputSingle swaps as little TokenIn as possible for exactly Params.AmountOut in one pool, signed by opts.Signer with req.NativeAmount (nil sends none) as msg.value.
func (r *Router) ExactOutputSingle(ctx context.Context, opts *bind.TransactOptions, req *ExactOutputSingleRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: exactOutputSingle: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv3: exactOutputSingle: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeExactOutputSingle(req.Params.TokenIn, req.Params.TokenOut, req.Params.Fee, req.Params.Recipient, req.Params.AmountOut, req.Params.AmountInMaximum, req.Params.SqrtPriceLimitX96)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: exactOutputSingle: %w", err)
	}

	var txOpts *bind.TransactOptions
	if opts != nil {
		o := *opts
		o.Value = req.NativeAmount
		txOpts = &o
	}

	signed, hash, err := bind.Transact(ctx, r.c, txOpts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: exactOutputSingle: %w", mapRouterRevertError(err))
	}

	return signed, hash, nil
}

// ExactOutput swaps for exactly Params.AmountOut along Params.Path, which runs tokenOut to tokenIn, signed by opts.Signer with req.NativeAmount (nil sends none) as msg.value.
func (r *Router) ExactOutput(ctx context.Context, opts *bind.TransactOptions, req *ExactOutputRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: exactOutput: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv3: exactOutput: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeExactOutput(req.Params.Path, req.Params.Recipient, req.Params.AmountOut, req.Params.AmountInMaximum)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: exactOutput: %w", err)
	}

	var txOpts *bind.TransactOptions
	if opts != nil {
		o := *opts
		o.Value = req.NativeAmount
		txOpts = &o
	}

	signed, hash, err := bind.Transact(ctx, r.c, txOpts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: exactOutput: %w", mapRouterRevertError(err))
	}

	return signed, hash, nil
}

// UnwrapWETH9 unwraps req.Contract's whole WETH9 balance to req.Recipient as native currency, signed by opts.Signer.
func (r *Router) UnwrapWETH9(ctx context.Context, opts *bind.TransactOptions, req *RouterUnwrapWETH9Request) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: unwrapWETH9: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv3: unwrapWETH9: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeRouterUnwrapWETH9(req.AmountMinimum, req.Recipient)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: unwrapWETH9: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: unwrapWETH9: %w", mapRouterRevertError(err))
	}

	return signed, hash, nil
}

// UnwrapWETH9WithFee is UnwrapWETH9 with req.FeeBips of the amount sent to req.FeeRecipient, signed by opts.Signer.
func (r *Router) UnwrapWETH9WithFee(ctx context.Context, opts *bind.TransactOptions, req *RouterUnwrapWETH9WithFeeRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: unwrapWETH9WithFee: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv3: unwrapWETH9WithFee: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeUnwrapWETH9WithFee(req.AmountMinimum, req.Recipient, req.FeeBips, req.FeeRecipient)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: unwrapWETH9WithFee: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: unwrapWETH9WithFee: %w", mapRouterRevertError(err))
	}

	return signed, hash, nil
}

// SweepToken sends req.Contract's whole req.Token balance to req.Recipient, signed by opts.Signer.
func (r *Router) SweepToken(ctx context.Context, opts *bind.TransactOptions, req *RouterSweepTokenRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: sweepToken: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv3: sweepToken: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeRouterSweepToken(req.Token, req.AmountMinimum, req.Recipient)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: sweepToken: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: sweepToken: %w", mapRouterRevertError(err))
	}

	return signed, hash, nil
}

// SweepTokenWithFee is SweepToken with req.FeeBips of the amount sent to req.FeeRecipient, signed by opts.Signer.
func (r *Router) SweepTokenWithFee(ctx context.Context, opts *bind.TransactOptions, req *RouterSweepTokenWithFeeRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: sweepTokenWithFee: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv3: sweepTokenWithFee: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeSweepTokenWithFee(req.Token, req.AmountMinimum, req.Recipient, req.FeeBips, req.FeeRecipient)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: sweepTokenWithFee: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: sweepTokenWithFee: %w", mapRouterRevertError(err))
	}

	return signed, hash, nil
}

// RefundETH returns req.Contract's native balance to the sender, signed by opts.Signer.
func (r *Router) RefundETH(ctx context.Context, opts *bind.TransactOptions, req *RouterRefundETHRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: refundETH: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv3: refundETH: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata := EncodeRouterRefundETH()

	signed, hash, err := bind.Transact(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: refundETH: %w", mapRouterRevertError(err))
	}

	return signed, hash, nil
}

// Multicall runs req.Data as one transaction on req.Contract; a non-nil req.Deadline uses the multicall(deadline, data) overload, signed by opts.Signer with req.NativeAmount (nil sends none) as msg.value.
func (r *Router) Multicall(ctx context.Context, opts *bind.TransactOptions, req *RouterMulticallRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: multicall: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv3: multicall: %w: value must be nil", bind.ErrInvalidOptions)
	}

	var calldata []byte
	var err error
	if req.Deadline != nil {
		calldata, err = EncodeRouterMulticallWithDeadline(req.Deadline, req.Data)
	} else {
		calldata, err = EncodeRouterMulticall(req.Data)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: multicall: %w", err)
	}

	var txOpts *bind.TransactOptions
	if opts != nil {
		o := *opts
		o.Value = req.NativeAmount
		txOpts = &o
	}

	signed, hash, err := bind.Transact(ctx, r.c, txOpts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: multicall: %w", mapRouterRevertError(err))
	}

	return signed, hash, nil
}

// SelfPermit approves req.Contract for req.Amount of req.Token with an EIP-2612 signature, signed by opts.Signer.
func (r *Router) SelfPermit(ctx context.Context, opts *bind.TransactOptions, req *SelfPermitRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: selfPermit: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv3: selfPermit: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeSelfPermit(req.Token, req.Amount, req.Deadline, req.Sig)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: selfPermit: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: selfPermit: %w", mapRouterRevertError(err))
	}

	return signed, hash, nil
}

// SelfPermitIfNecessary is SelfPermit, skipped when the allowance already covers req.Amount, signed by opts.Signer.
func (r *Router) SelfPermitIfNecessary(ctx context.Context, opts *bind.TransactOptions, req *SelfPermitIfNecessaryRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: selfPermitIfNecessary: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv3: selfPermitIfNecessary: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeSelfPermitIfNecessary(req.Token, req.Amount, req.Deadline, req.Sig)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: selfPermitIfNecessary: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: selfPermitIfNecessary: %w", mapRouterRevertError(err))
	}

	return signed, hash, nil
}

// SelfPermitAllowed approves req.Contract for all of req.Token with a DAI-style permit signature, signed by opts.Signer.
func (r *Router) SelfPermitAllowed(ctx context.Context, opts *bind.TransactOptions, req *SelfPermitAllowedRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: selfPermitAllowed: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv3: selfPermitAllowed: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeSelfPermitAllowed(req.Token, req.Nonce, req.Expiry, req.Sig)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: selfPermitAllowed: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: selfPermitAllowed: %w", mapRouterRevertError(err))
	}

	return signed, hash, nil
}

// SelfPermitAllowedIfNecessary is SelfPermitAllowed, skipped when the allowance is already unlimited, signed by opts.Signer.
func (r *Router) SelfPermitAllowedIfNecessary(ctx context.Context, opts *bind.TransactOptions, req *SelfPermitAllowedIfNecessaryRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: selfPermitAllowedIfNecessary: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv3: selfPermitAllowedIfNecessary: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeSelfPermitAllowedIfNecessary(req.Token, req.Nonce, req.Expiry, req.Sig)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: selfPermitAllowedIfNecessary: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: selfPermitAllowedIfNecessary: %w", mapRouterRevertError(err))
	}

	return signed, hash, nil
}
