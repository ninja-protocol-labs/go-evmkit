package v2

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// IRouter mirrors IUniswapV2Router02 (which embeds IUniswapV2Router01).
type IRouter interface {
	Factory(ctx context.Context, opts *bind.CallOptions, req *FactoryRequest) (*types.Address, error)
	WETH(ctx context.Context, opts *bind.CallOptions, req *WETHRequest) (*types.Address, error)
	Quote(ctx context.Context, opts *bind.CallOptions, req *QuoteRequest) (*big.Int, error)
	GetAmountOut(ctx context.Context, opts *bind.CallOptions, req *GetAmountOutRequest) (*big.Int, error)
	GetAmountIn(ctx context.Context, opts *bind.CallOptions, req *GetAmountInRequest) (*big.Int, error)
	GetAmountsOut(ctx context.Context, opts *bind.CallOptions, req *GetAmountsOutRequest) ([]*big.Int, error)
	GetAmountsIn(ctx context.Context, opts *bind.CallOptions, req *GetAmountsInRequest) ([]*big.Int, error)

	CallAddLiquidity(ctx context.Context, opts *bind.CallOptions, req *AddLiquidityRequest) (*AddLiquidityResult, error)
	CallAddLiquidityETH(ctx context.Context, opts *bind.CallOptions, req *AddLiquidityETHRequest) (*AddLiquidityETHResult, error)
	CallRemoveLiquidity(ctx context.Context, opts *bind.CallOptions, req *RemoveLiquidityRequest) (*RemoveLiquidityResult, error)
	CallRemoveLiquidityETH(ctx context.Context, opts *bind.CallOptions, req *RemoveLiquidityETHRequest) (*RemoveLiquidityETHResult, error)
	CallRemoveLiquidityWithPermit(ctx context.Context, opts *bind.CallOptions, req *RemoveLiquidityWithPermitRequest) (*RemoveLiquidityResult, error)
	CallRemoveLiquidityETHWithPermit(ctx context.Context, opts *bind.CallOptions, req *RemoveLiquidityETHWithPermitRequest) (*RemoveLiquidityETHResult, error)
	CallSwapExactTokensForTokens(ctx context.Context, opts *bind.CallOptions, req *SwapExactTokensForTokensRequest) ([]*big.Int, error)
	CallSwapTokensForExactTokens(ctx context.Context, opts *bind.CallOptions, req *SwapTokensForExactTokensRequest) ([]*big.Int, error)
	CallSwapExactETHForTokens(ctx context.Context, opts *bind.CallOptions, req *SwapExactETHForTokensRequest) ([]*big.Int, error)
	CallSwapTokensForExactETH(ctx context.Context, opts *bind.CallOptions, req *SwapTokensForExactETHRequest) ([]*big.Int, error)
	CallSwapExactTokensForETH(ctx context.Context, opts *bind.CallOptions, req *SwapExactTokensForETHRequest) ([]*big.Int, error)
	CallSwapETHForExactTokens(ctx context.Context, opts *bind.CallOptions, req *SwapETHForExactTokensRequest) ([]*big.Int, error)
	CallRemoveLiquidityETHSupportingFeeOnTransferTokens(ctx context.Context, opts *bind.CallOptions, req *RemoveLiquidityETHSupportingFeeOnTransferTokensRequest) (*big.Int, error)
	CallRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(ctx context.Context, opts *bind.CallOptions, req *RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest) (*big.Int, error)

	AddLiquidity(ctx context.Context, opts *bind.TransactOptions, req *AddLiquidityRequest) (core.Transaction, *types.Hash, error)
	AddLiquidityETH(ctx context.Context, opts *bind.TransactOptions, req *AddLiquidityETHRequest) (core.Transaction, *types.Hash, error)
	RemoveLiquidity(ctx context.Context, opts *bind.TransactOptions, req *RemoveLiquidityRequest) (core.Transaction, *types.Hash, error)
	RemoveLiquidityETH(ctx context.Context, opts *bind.TransactOptions, req *RemoveLiquidityETHRequest) (core.Transaction, *types.Hash, error)
	RemoveLiquidityWithPermit(ctx context.Context, opts *bind.TransactOptions, req *RemoveLiquidityWithPermitRequest) (core.Transaction, *types.Hash, error)
	RemoveLiquidityETHWithPermit(ctx context.Context, opts *bind.TransactOptions, req *RemoveLiquidityETHWithPermitRequest) (core.Transaction, *types.Hash, error)
	SwapExactTokensForTokens(ctx context.Context, opts *bind.TransactOptions, req *SwapExactTokensForTokensRequest) (core.Transaction, *types.Hash, error)
	SwapTokensForExactTokens(ctx context.Context, opts *bind.TransactOptions, req *SwapTokensForExactTokensRequest) (core.Transaction, *types.Hash, error)
	SwapExactETHForTokens(ctx context.Context, opts *bind.TransactOptions, req *SwapExactETHForTokensRequest) (core.Transaction, *types.Hash, error)
	SwapTokensForExactETH(ctx context.Context, opts *bind.TransactOptions, req *SwapTokensForExactETHRequest) (core.Transaction, *types.Hash, error)
	SwapExactTokensForETH(ctx context.Context, opts *bind.TransactOptions, req *SwapExactTokensForETHRequest) (core.Transaction, *types.Hash, error)
	SwapETHForExactTokens(ctx context.Context, opts *bind.TransactOptions, req *SwapETHForExactTokensRequest) (core.Transaction, *types.Hash, error)
	RemoveLiquidityETHSupportingFeeOnTransferTokens(ctx context.Context, opts *bind.TransactOptions, req *RemoveLiquidityETHSupportingFeeOnTransferTokensRequest) (core.Transaction, *types.Hash, error)
	RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(ctx context.Context, opts *bind.TransactOptions, req *RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest) (core.Transaction, *types.Hash, error)
	SwapExactTokensForTokensSupportingFeeOnTransferTokens(ctx context.Context, opts *bind.TransactOptions, req *SwapExactTokensForTokensSupportingFeeOnTransferTokensRequest) (core.Transaction, *types.Hash, error)
	SwapExactETHForTokensSupportingFeeOnTransferTokens(ctx context.Context, opts *bind.TransactOptions, req *SwapExactETHForTokensSupportingFeeOnTransferTokensRequest) (core.Transaction, *types.Hash, error)
	SwapExactTokensForETHSupportingFeeOnTransferTokens(ctx context.Context, opts *bind.TransactOptions, req *SwapExactTokensForETHSupportingFeeOnTransferTokensRequest) (core.Transaction, *types.Hash, error)
}

var _ IRouter = (*Router)(nil)

// Router calls any UniswapV2Router02 contract over a Client; it holds no contract address, so one instance serves every deployment.
type Router struct {
	c rpc.Client
}

// NewRouter returns a Router calling routers via cli.
func NewRouter(cli rpc.Client) *Router {
	return &Router{
		c: cli,
	}
}

// Factory returns req.Contract's UniswapV2Factory.
func (r *Router) Factory(ctx context.Context, opts *bind.CallOptions, req *FactoryRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: factory: %w", err)
	}

	data, err := bind.Call(ctx, r.c, opts, req.Contract, EncodeRouterFactory())
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: factory: %w", r.wrapRevertError(err))
	}

	out, err := DecodeRouterFactory(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: factory: %w", err)
	}

	return out, nil
}

// WETH returns req.Contract's wrapped native token.
func (r *Router) WETH(ctx context.Context, opts *bind.CallOptions, req *WETHRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: weth: %w", err)
	}

	data, err := bind.Call(ctx, r.c, opts, req.Contract, EncodeWETH())
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: weth: %w", r.wrapRevertError(err))
	}

	out, err := DecodeWETH(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: weth: %w", err)
	}

	return out, nil
}

// Quote returns the amount of the other asset equivalent to req.AmountA at the given reserves, with no fee applied.
func (r *Router) Quote(ctx context.Context, opts *bind.CallOptions, req *QuoteRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: quote: %w", err)
	}

	calldata, err := EncodeQuote(req.AmountA, req.ReserveA, req.ReserveB)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: quote: %w", err)
	}

	data, err := bind.Call(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: quote: %w", r.wrapRevertError(err))
	}

	out, err := DecodeQuote(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: quote: %w", err)
	}

	return out, nil
}

// GetAmountOut returns the output of a single swap of req.AmountIn at the given reserves, after the 0.3% fee.
func (r *Router) GetAmountOut(ctx context.Context, opts *bind.CallOptions, req *GetAmountOutRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountOut: %w", err)
	}

	calldata, err := EncodeGetAmountOut(req.AmountIn, req.ReserveIn, req.ReserveOut)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountOut: %w", err)
	}

	data, err := bind.Call(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountOut: %w", r.wrapRevertError(err))
	}

	out, err := DecodeGetAmountOut(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountOut: %w", err)
	}

	return out, nil
}

// GetAmountIn returns the input a single swap needs to produce req.AmountOut at the given reserves.
func (r *Router) GetAmountIn(ctx context.Context, opts *bind.CallOptions, req *GetAmountInRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountIn: %w", err)
	}

	calldata, err := EncodeGetAmountIn(req.AmountOut, req.ReserveIn, req.ReserveOut)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountIn: %w", err)
	}

	data, err := bind.Call(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountIn: %w", r.wrapRevertError(err))
	}

	out, err := DecodeGetAmountIn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountIn: %w", err)
	}

	return out, nil
}

// GetAmountsOut returns per-hop amounts for swapping req.AmountIn along req.Path; the last is the final output.
func (r *Router) GetAmountsOut(ctx context.Context, opts *bind.CallOptions, req *GetAmountsOutRequest) ([]*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountsOut: %w", err)
	}

	calldata, err := EncodeGetAmountsOut(req.AmountIn, req.Path)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountsOut: %w", err)
	}

	data, err := bind.Call(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountsOut: %w", r.wrapRevertError(err))
	}

	out, err := DecodeGetAmountsOut(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountsOut: %w", err)
	}

	return out, nil
}

// GetAmountsIn returns per-hop amounts needed to produce req.AmountOut at the end of req.Path; the first is the required input.
func (r *Router) GetAmountsIn(ctx context.Context, opts *bind.CallOptions, req *GetAmountsInRequest) ([]*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountsIn: %w", err)
	}

	calldata, err := EncodeGetAmountsIn(req.AmountOut, req.Path)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountsIn: %w", err)
	}

	data, err := bind.Call(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountsIn: %w", r.wrapRevertError(err))
	}

	out, err := DecodeGetAmountsIn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountsIn: %w", err)
	}

	return out, nil
}

// CallAddLiquidity runs addLiquidity via eth_call as opts.From, returning its result or the revert.
func (r *Router) CallAddLiquidity(ctx context.Context, opts *bind.CallOptions, req *AddLiquidityRequest) (*AddLiquidityResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: callAddLiquidity: %w", err)
	}

	calldata, err := EncodeAddLiquidity(req.TokenA, req.TokenB, req.AmountADesired, req.AmountBDesired, req.AmountAMin, req.AmountBMin, req.To, req.Deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callAddLiquidity: %w", err)
	}

	data, err := bind.Call(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callAddLiquidity: %w", r.wrapRevertError(err))
	}

	out, err := DecodeAddLiquidity(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callAddLiquidity: %w", err)
	}

	return out, nil
}

// CallAddLiquidityETH runs addLiquidityETH via eth_call as opts.From with req.AmountETHDesired as msg.value, returning its result or the revert.
func (r *Router) CallAddLiquidityETH(ctx context.Context, opts *bind.CallOptions, req *AddLiquidityETHRequest) (*AddLiquidityETHResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: callAddLiquidityETH: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, fmt.Errorf("uniswapv2: callAddLiquidityETH: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeAddLiquidityETH(req.Token, req.AmountTokenDesired, req.AmountTokenMin, req.AmountETHMin, req.To, req.Deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callAddLiquidityETH: %w", err)
	}

	callOpts := bind.NewCallOptions()
	if opts != nil {
		*callOpts = *opts
	}
	callOpts.Value = req.AmountETHDesired

	data, err := bind.Call(ctx, r.c, callOpts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callAddLiquidityETH: %w", r.wrapRevertError(err))
	}

	out, err := DecodeAddLiquidityETH(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callAddLiquidityETH: %w", err)
	}

	return out, nil
}

// CallRemoveLiquidity runs removeLiquidity via eth_call as opts.From, returning its result or the revert.
func (r *Router) CallRemoveLiquidity(ctx context.Context, opts *bind.CallOptions, req *RemoveLiquidityRequest) (*RemoveLiquidityResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidity: %w", err)
	}

	calldata, err := EncodeRemoveLiquidity(req.TokenA, req.TokenB, req.Liquidity, req.AmountAMin, req.AmountBMin, req.To, req.Deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidity: %w", err)
	}

	data, err := bind.Call(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidity: %w", r.wrapRevertError(err))
	}

	out, err := DecodeRemoveLiquidity(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidity: %w", err)
	}

	return out, nil
}

// CallRemoveLiquidityETH runs removeLiquidityETH via eth_call as opts.From, returning its result or the revert.
func (r *Router) CallRemoveLiquidityETH(ctx context.Context, opts *bind.CallOptions, req *RemoveLiquidityETHRequest) (*RemoveLiquidityETHResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETH: %w", err)
	}

	calldata, err := EncodeRemoveLiquidityETH(req.Token, req.Liquidity, req.AmountTokenMin, req.AmountETHMin, req.To, req.Deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETH: %w", err)
	}

	data, err := bind.Call(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETH: %w", r.wrapRevertError(err))
	}

	out, err := DecodeRemoveLiquidityETH(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETH: %w", err)
	}

	return out, nil
}

// CallRemoveLiquidityWithPermit runs removeLiquidityWithPermit via eth_call as opts.From, returning its result or the revert.
func (r *Router) CallRemoveLiquidityWithPermit(ctx context.Context, opts *bind.CallOptions, req *RemoveLiquidityWithPermitRequest) (*RemoveLiquidityResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityWithPermit: %w", err)
	}

	calldata, err := EncodeRemoveLiquidityWithPermit(req.TokenA, req.TokenB, req.Liquidity, req.AmountAMin, req.AmountBMin, req.To, req.Deadline, req.ApproveMax, req.Sig)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityWithPermit: %w", err)
	}

	data, err := bind.Call(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityWithPermit: %w", r.wrapRevertError(err))
	}

	out, err := DecodeRemoveLiquidityWithPermit(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityWithPermit: %w", err)
	}

	return out, nil
}

// CallRemoveLiquidityETHWithPermit runs removeLiquidityETHWithPermit via eth_call as opts.From, returning its result or the revert.
func (r *Router) CallRemoveLiquidityETHWithPermit(ctx context.Context, opts *bind.CallOptions, req *RemoveLiquidityETHWithPermitRequest) (*RemoveLiquidityETHResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETHWithPermit: %w", err)
	}

	calldata, err := EncodeRemoveLiquidityETHWithPermit(req.Token, req.Liquidity, req.AmountTokenMin, req.AmountETHMin, req.To, req.Deadline, req.ApproveMax, req.Sig)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETHWithPermit: %w", err)
	}

	data, err := bind.Call(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETHWithPermit: %w", r.wrapRevertError(err))
	}

	out, err := DecodeRemoveLiquidityETHWithPermit(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETHWithPermit: %w", err)
	}

	return out, nil
}

// CallSwapExactTokensForTokens runs swapExactTokensForTokens via eth_call as opts.From, returning its result or the revert.
func (r *Router) CallSwapExactTokensForTokens(ctx context.Context, opts *bind.CallOptions, req *SwapExactTokensForTokensRequest) ([]*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapExactTokensForTokens: %w", err)
	}

	calldata, err := EncodeSwapExactTokensForTokens(req.AmountIn, req.AmountOutMin, req.Path, req.To, req.Deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapExactTokensForTokens: %w", err)
	}

	data, err := bind.Call(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapExactTokensForTokens: %w", r.wrapRevertError(err))
	}

	out, err := DecodeSwapExactTokensForTokens(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapExactTokensForTokens: %w", err)
	}

	return out, nil
}

// CallSwapTokensForExactTokens runs swapTokensForExactTokens via eth_call as opts.From, returning its result or the revert.
func (r *Router) CallSwapTokensForExactTokens(ctx context.Context, opts *bind.CallOptions, req *SwapTokensForExactTokensRequest) ([]*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapTokensForExactTokens: %w", err)
	}

	calldata, err := EncodeSwapTokensForExactTokens(req.AmountOut, req.AmountInMax, req.Path, req.To, req.Deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapTokensForExactTokens: %w", err)
	}

	data, err := bind.Call(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapTokensForExactTokens: %w", r.wrapRevertError(err))
	}

	out, err := DecodeSwapTokensForExactTokens(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapTokensForExactTokens: %w", err)
	}

	return out, nil
}

// CallSwapExactETHForTokens runs swapExactETHForTokens via eth_call as opts.From with req.AmountIn as msg.value, returning its result or the revert.
func (r *Router) CallSwapExactETHForTokens(ctx context.Context, opts *bind.CallOptions, req *SwapExactETHForTokensRequest) ([]*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapExactETHForTokens: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapExactETHForTokens: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeSwapExactETHForTokens(req.AmountOutMin, req.Path, req.To, req.Deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapExactETHForTokens: %w", err)
	}

	callOpts := bind.NewCallOptions()
	if opts != nil {
		*callOpts = *opts
	}
	callOpts.Value = req.AmountIn

	data, err := bind.Call(ctx, r.c, callOpts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapExactETHForTokens: %w", r.wrapRevertError(err))
	}

	out, err := DecodeSwapExactETHForTokens(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapExactETHForTokens: %w", err)
	}

	return out, nil
}

// CallSwapTokensForExactETH runs swapTokensForExactETH via eth_call as opts.From, returning its result or the revert.
func (r *Router) CallSwapTokensForExactETH(ctx context.Context, opts *bind.CallOptions, req *SwapTokensForExactETHRequest) ([]*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapTokensForExactETH: %w", err)
	}

	calldata, err := EncodeSwapTokensForExactETH(req.AmountOut, req.AmountInMax, req.Path, req.To, req.Deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapTokensForExactETH: %w", err)
	}

	data, err := bind.Call(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapTokensForExactETH: %w", r.wrapRevertError(err))
	}

	out, err := DecodeSwapTokensForExactETH(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapTokensForExactETH: %w", err)
	}

	return out, nil
}

// CallSwapExactTokensForETH runs swapExactTokensForETH via eth_call as opts.From, returning its result or the revert.
func (r *Router) CallSwapExactTokensForETH(ctx context.Context, opts *bind.CallOptions, req *SwapExactTokensForETHRequest) ([]*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapExactTokensForETH: %w", err)
	}

	calldata, err := EncodeSwapExactTokensForETH(req.AmountIn, req.AmountOutMin, req.Path, req.To, req.Deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapExactTokensForETH: %w", err)
	}

	data, err := bind.Call(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapExactTokensForETH: %w", r.wrapRevertError(err))
	}

	out, err := DecodeSwapExactTokensForETH(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapExactTokensForETH: %w", err)
	}

	return out, nil
}

// CallSwapETHForExactTokens runs swapETHForExactTokens via eth_call as opts.From with req.AmountInMax as msg.value, returning its result or the revert.
func (r *Router) CallSwapETHForExactTokens(ctx context.Context, opts *bind.CallOptions, req *SwapETHForExactTokensRequest) ([]*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapETHForExactTokens: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapETHForExactTokens: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeSwapETHForExactTokens(req.AmountOut, req.Path, req.To, req.Deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapETHForExactTokens: %w", err)
	}

	callOpts := bind.NewCallOptions()
	if opts != nil {
		*callOpts = *opts
	}
	callOpts.Value = req.AmountInMax

	data, err := bind.Call(ctx, r.c, callOpts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapETHForExactTokens: %w", r.wrapRevertError(err))
	}

	out, err := DecodeSwapETHForExactTokens(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapETHForExactTokens: %w", err)
	}

	return out, nil
}

// CallRemoveLiquidityETHSupportingFeeOnTransferTokens runs removeLiquidityETHSupportingFeeOnTransferTokens via eth_call as opts.From, returning its result or the revert.
func (r *Router) CallRemoveLiquidityETHSupportingFeeOnTransferTokens(ctx context.Context, opts *bind.CallOptions, req *RemoveLiquidityETHSupportingFeeOnTransferTokensRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETHSupportingFeeOnTransferTokens: %w", err)
	}

	calldata, err := EncodeRemoveLiquidityETHSupportingFeeOnTransferTokens(req.Token, req.Liquidity, req.AmountTokenMin, req.AmountETHMin, req.To, req.Deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETHSupportingFeeOnTransferTokens: %w", err)
	}

	data, err := bind.Call(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETHSupportingFeeOnTransferTokens: %w", r.wrapRevertError(err))
	}

	out, err := DecodeRemoveLiquidityETHSupportingFeeOnTransferTokens(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETHSupportingFeeOnTransferTokens: %w", err)
	}

	return out, nil
}

// CallRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens runs removeLiquidityETHWithPermitSupportingFeeOnTransferTokens via eth_call as opts.From, returning its result or the revert.
func (r *Router) CallRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(ctx context.Context, opts *bind.CallOptions, req *RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens: %w", err)
	}

	calldata, err := EncodeRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(req.Token, req.Liquidity, req.AmountTokenMin, req.AmountETHMin, req.To, req.Deadline, req.ApproveMax, req.Sig)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens: %w", err)
	}

	data, err := bind.Call(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens: %w", r.wrapRevertError(err))
	}

	out, err := DecodeRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens: %w", err)
	}

	return out, nil
}

// AddLiquidity adds liquidity to the TokenA/TokenB pair, creating it if needed, signed by opts.Signer.
func (r *Router) AddLiquidity(ctx context.Context, opts *bind.TransactOptions, req *AddLiquidityRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: addLiquidity: %w", err)
	}

	calldata, err := EncodeAddLiquidity(req.TokenA, req.TokenB, req.AmountADesired, req.AmountBDesired, req.AmountAMin, req.AmountBMin, req.To, req.Deadline)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: addLiquidity: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: addLiquidity: %w", r.wrapRevertError(err))
	}

	return signed, hash, nil
}

// AddLiquidityETH adds liquidity to the Token/WETH pair with req.AmountETHDesired attached as msg.value; unused ETH is refunded, signed by opts.Signer.
func (r *Router) AddLiquidityETH(ctx context.Context, opts *bind.TransactOptions, req *AddLiquidityETHRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: addLiquidityETH: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv2: addLiquidityETH: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeAddLiquidityETH(req.Token, req.AmountTokenDesired, req.AmountTokenMin, req.AmountETHMin, req.To, req.Deadline)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: addLiquidityETH: %w", err)
	}

	var txOpts *bind.TransactOptions
	if opts != nil {
		o := *opts
		o.Value = req.AmountETHDesired
		txOpts = &o
	}

	signed, hash, err := bind.Transact(ctx, r.c, txOpts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: addLiquidityETH: %w", r.wrapRevertError(err))
	}

	return signed, hash, nil
}

// RemoveLiquidity burns req.Liquidity LP tokens of the TokenA/TokenB pair; the router must be approved for them, signed by opts.Signer.
func (r *Router) RemoveLiquidity(ctx context.Context, opts *bind.TransactOptions, req *RemoveLiquidityRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: removeLiquidity: %w", err)
	}

	calldata, err := EncodeRemoveLiquidity(req.TokenA, req.TokenB, req.Liquidity, req.AmountAMin, req.AmountBMin, req.To, req.Deadline)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: removeLiquidity: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: removeLiquidity: %w", r.wrapRevertError(err))
	}

	return signed, hash, nil
}

// RemoveLiquidityETH burns req.Liquidity LP tokens of the Token/WETH pair, paying the WETH side out as ETH, signed by opts.Signer.
func (r *Router) RemoveLiquidityETH(ctx context.Context, opts *bind.TransactOptions, req *RemoveLiquidityETHRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: removeLiquidityETH: %w", err)
	}

	calldata, err := EncodeRemoveLiquidityETH(req.Token, req.Liquidity, req.AmountTokenMin, req.AmountETHMin, req.To, req.Deadline)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: removeLiquidityETH: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: removeLiquidityETH: %w", r.wrapRevertError(err))
	}

	return signed, hash, nil
}

// RemoveLiquidityWithPermit is RemoveLiquidity with the LP approval given by req.Sig, an EIP-2612 permit signature, signed by opts.Signer.
func (r *Router) RemoveLiquidityWithPermit(ctx context.Context, opts *bind.TransactOptions, req *RemoveLiquidityWithPermitRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: removeLiquidityWithPermit: %w", err)
	}

	calldata, err := EncodeRemoveLiquidityWithPermit(req.TokenA, req.TokenB, req.Liquidity, req.AmountAMin, req.AmountBMin, req.To, req.Deadline, req.ApproveMax, req.Sig)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: removeLiquidityWithPermit: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: removeLiquidityWithPermit: %w", r.wrapRevertError(err))
	}

	return signed, hash, nil
}

// RemoveLiquidityETHWithPermit is RemoveLiquidityETH with the LP approval given by req.Sig, an EIP-2612 permit signature, signed by opts.Signer.
func (r *Router) RemoveLiquidityETHWithPermit(ctx context.Context, opts *bind.TransactOptions, req *RemoveLiquidityETHWithPermitRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: removeLiquidityETHWithPermit: %w", err)
	}

	calldata, err := EncodeRemoveLiquidityETHWithPermit(req.Token, req.Liquidity, req.AmountTokenMin, req.AmountETHMin, req.To, req.Deadline, req.ApproveMax, req.Sig)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: removeLiquidityETHWithPermit: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: removeLiquidityETHWithPermit: %w", r.wrapRevertError(err))
	}

	return signed, hash, nil
}

// SwapExactTokensForTokens swaps exactly req.AmountIn of Path[0] for at least req.AmountOutMin of the last token, signed by opts.Signer.
func (r *Router) SwapExactTokensForTokens(ctx context.Context, opts *bind.TransactOptions, req *SwapExactTokensForTokensRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: swapExactTokensForTokens: %w", err)
	}

	calldata, err := EncodeSwapExactTokensForTokens(req.AmountIn, req.AmountOutMin, req.Path, req.To, req.Deadline)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: swapExactTokensForTokens: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: swapExactTokensForTokens: %w", r.wrapRevertError(err))
	}

	return signed, hash, nil
}

// SwapTokensForExactTokens swaps at most req.AmountInMax of Path[0] for exactly req.AmountOut of the last token, signed by opts.Signer.
func (r *Router) SwapTokensForExactTokens(ctx context.Context, opts *bind.TransactOptions, req *SwapTokensForExactTokensRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: swapTokensForExactTokens: %w", err)
	}

	calldata, err := EncodeSwapTokensForExactTokens(req.AmountOut, req.AmountInMax, req.Path, req.To, req.Deadline)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: swapTokensForExactTokens: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: swapTokensForExactTokens: %w", r.wrapRevertError(err))
	}

	return signed, hash, nil
}

// SwapExactETHForTokens swaps exactly req.AmountIn of ETH, attached as msg.value, for at least req.AmountOutMin of the last token, signed by opts.Signer.
func (r *Router) SwapExactETHForTokens(ctx context.Context, opts *bind.TransactOptions, req *SwapExactETHForTokensRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: swapExactETHForTokens: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv2: swapExactETHForTokens: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeSwapExactETHForTokens(req.AmountOutMin, req.Path, req.To, req.Deadline)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: swapExactETHForTokens: %w", err)
	}

	var txOpts *bind.TransactOptions
	if opts != nil {
		o := *opts
		o.Value = req.AmountIn
		txOpts = &o
	}

	signed, hash, err := bind.Transact(ctx, r.c, txOpts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: swapExactETHForTokens: %w", r.wrapRevertError(err))
	}

	return signed, hash, nil
}

// SwapTokensForExactETH swaps at most req.AmountInMax of Path[0] for exactly req.AmountOut of ETH, signed by opts.Signer.
func (r *Router) SwapTokensForExactETH(ctx context.Context, opts *bind.TransactOptions, req *SwapTokensForExactETHRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: swapTokensForExactETH: %w", err)
	}

	calldata, err := EncodeSwapTokensForExactETH(req.AmountOut, req.AmountInMax, req.Path, req.To, req.Deadline)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: swapTokensForExactETH: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: swapTokensForExactETH: %w", r.wrapRevertError(err))
	}

	return signed, hash, nil
}

// SwapExactTokensForETH swaps exactly req.AmountIn of Path[0] for at least req.AmountOutMin of ETH, signed by opts.Signer.
func (r *Router) SwapExactTokensForETH(ctx context.Context, opts *bind.TransactOptions, req *SwapExactTokensForETHRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: swapExactTokensForETH: %w", err)
	}

	calldata, err := EncodeSwapExactTokensForETH(req.AmountIn, req.AmountOutMin, req.Path, req.To, req.Deadline)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: swapExactTokensForETH: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: swapExactTokensForETH: %w", r.wrapRevertError(err))
	}

	return signed, hash, nil
}

// SwapETHForExactTokens swaps at most req.AmountInMax of ETH, attached as msg.value, for exactly req.AmountOut of the last token; unused ETH is refunded, signed by opts.Signer.
func (r *Router) SwapETHForExactTokens(ctx context.Context, opts *bind.TransactOptions, req *SwapETHForExactTokensRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: swapETHForExactTokens: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv2: swapETHForExactTokens: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeSwapETHForExactTokens(req.AmountOut, req.Path, req.To, req.Deadline)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: swapETHForExactTokens: %w", err)
	}

	var txOpts *bind.TransactOptions
	if opts != nil {
		o := *opts
		o.Value = req.AmountInMax
		txOpts = &o
	}

	signed, hash, err := bind.Transact(ctx, r.c, txOpts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: swapETHForExactTokens: %w", r.wrapRevertError(err))
	}

	return signed, hash, nil
}

// RemoveLiquidityETHSupportingFeeOnTransferTokens is RemoveLiquidityETH for tokens that take a fee on transfer, signed by opts.Signer.
func (r *Router) RemoveLiquidityETHSupportingFeeOnTransferTokens(ctx context.Context, opts *bind.TransactOptions, req *RemoveLiquidityETHSupportingFeeOnTransferTokensRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: removeLiquidityETHSupportingFeeOnTransferTokens: %w", err)
	}

	calldata, err := EncodeRemoveLiquidityETHSupportingFeeOnTransferTokens(req.Token, req.Liquidity, req.AmountTokenMin, req.AmountETHMin, req.To, req.Deadline)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: removeLiquidityETHSupportingFeeOnTransferTokens: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: removeLiquidityETHSupportingFeeOnTransferTokens: %w", r.wrapRevertError(err))
	}

	return signed, hash, nil
}

// RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens is RemoveLiquidityETHWithPermit for tokens that take a fee on transfer, signed by opts.Signer.
func (r *Router) RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(ctx context.Context, opts *bind.TransactOptions, req *RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: removeLiquidityETHWithPermitSupportingFeeOnTransferTokens: %w", err)
	}

	calldata, err := EncodeRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(req.Token, req.Liquidity, req.AmountTokenMin, req.AmountETHMin, req.To, req.Deadline, req.ApproveMax, req.Sig)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: removeLiquidityETHWithPermitSupportingFeeOnTransferTokens: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: removeLiquidityETHWithPermitSupportingFeeOnTransferTokens: %w", r.wrapRevertError(err))
	}

	return signed, hash, nil
}

// SwapExactTokensForTokensSupportingFeeOnTransferTokens is SwapExactTokensForTokens for tokens that take a fee on transfer, signed by opts.Signer.
func (r *Router) SwapExactTokensForTokensSupportingFeeOnTransferTokens(ctx context.Context, opts *bind.TransactOptions, req *SwapExactTokensForTokensSupportingFeeOnTransferTokensRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: swapExactTokensForTokensSupportingFeeOnTransferTokens: %w", err)
	}

	calldata, err := EncodeSwapExactTokensForTokensSupportingFeeOnTransferTokens(req.AmountIn, req.AmountOutMin, req.Path, req.To, req.Deadline)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: swapExactTokensForTokensSupportingFeeOnTransferTokens: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: swapExactTokensForTokensSupportingFeeOnTransferTokens: %w", r.wrapRevertError(err))
	}

	return signed, hash, nil
}

// SwapExactETHForTokensSupportingFeeOnTransferTokens is SwapExactETHForTokens for tokens that take a fee on transfer, signed by opts.Signer.
func (r *Router) SwapExactETHForTokensSupportingFeeOnTransferTokens(ctx context.Context, opts *bind.TransactOptions, req *SwapExactETHForTokensSupportingFeeOnTransferTokensRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: swapExactETHForTokensSupportingFeeOnTransferTokens: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv2: swapExactETHForTokensSupportingFeeOnTransferTokens: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeSwapExactETHForTokensSupportingFeeOnTransferTokens(req.AmountOutMin, req.Path, req.To, req.Deadline)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: swapExactETHForTokensSupportingFeeOnTransferTokens: %w", err)
	}

	var txOpts *bind.TransactOptions
	if opts != nil {
		o := *opts
		o.Value = req.AmountIn
		txOpts = &o
	}

	signed, hash, err := bind.Transact(ctx, r.c, txOpts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: swapExactETHForTokensSupportingFeeOnTransferTokens: %w", r.wrapRevertError(err))
	}

	return signed, hash, nil
}

// SwapExactTokensForETHSupportingFeeOnTransferTokens is SwapExactTokensForETH for tokens that take a fee on transfer, signed by opts.Signer.
func (r *Router) SwapExactTokensForETHSupportingFeeOnTransferTokens(ctx context.Context, opts *bind.TransactOptions, req *SwapExactTokensForETHSupportingFeeOnTransferTokensRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: swapExactTokensForETHSupportingFeeOnTransferTokens: %w", err)
	}

	calldata, err := EncodeSwapExactTokensForETHSupportingFeeOnTransferTokens(req.AmountIn, req.AmountOutMin, req.Path, req.To, req.Deadline)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: swapExactTokensForETHSupportingFeeOnTransferTokens: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, r.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: swapExactTokensForETHSupportingFeeOnTransferTokens: %w", r.wrapRevertError(err))
	}

	return signed, hash, nil
}

// wrapRevertError maps a known UniswapV2Router/UniswapV2Library revert reason in err to its sentinel.
func (_ *Router) wrapRevertError(err error) error {
	var rpcErr *rpc.ResponseError
	if !errors.As(err, &rpcErr) || len(rpcErr.Data) == 0 {
		return err
	}

	var hexData string
	if jsonErr := json.Unmarshal(rpcErr.Data, &hexData); jsonErr != nil {
		return err
	}

	data, decErr := encoding.Hex.Decode(hexData)
	if decErr != nil {
		return err
	}

	reason, decErr := abi.DecodeRevert(data)
	if decErr != nil {
		return err
	}

	switch reason {
	case reasonExpired:
		return fmt.Errorf("%w: %w", ErrExpired, err)
	case reasonInsufficientAAmount:
		return fmt.Errorf("%w: %w", ErrInsufficientAAmount, err)
	case reasonInsufficientBAmount:
		return fmt.Errorf("%w: %w", ErrInsufficientBAmount, err)
	case reasonRouterInsufficientOutput, reasonLibraryInsufficientOutputOld, reasonLibraryInsufficientOutputNew:
		return fmt.Errorf("%w: %w", ErrInsufficientOutputAmount, err)
	case reasonExcessiveInputAmount:
		return fmt.Errorf("%w: %w", ErrExcessiveInputAmount, err)
	case reasonRouterInvalidPath, reasonLibraryInvalidPathOld, reasonLibraryInvalidPathNew:
		return fmt.Errorf("%w: %w", ErrInvalidPath, err)
	case reasonLibraryIdenticalAddressesOld, reasonLibraryIdenticalAddressesNew:
		return fmt.Errorf("%w: %w", ErrIdenticalAddresses, err)
	case reasonLibraryZeroAddressOld, reasonLibraryZeroAddressNew:
		return fmt.Errorf("%w: %w", ErrZeroAddress, err)
	case reasonLibraryInsufficientAmountOld, reasonLibraryInsufficientAmountNew:
		return fmt.Errorf("%w: %w", ErrInsufficientAmount, err)
	case reasonLibraryInsufficientLiquidityOld, reasonLibraryInsufficientLiquidityNew:
		return fmt.Errorf("%w: %w", ErrInsufficientLiquidity, err)
	case reasonLibraryInsufficientInputOld, reasonLibraryInsufficientInputNew:
		return fmt.Errorf("%w: %w", ErrInsufficientInputAmount, err)
	default:
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}
}

// MapRevertError maps a v2 Router, Library or Pair revert reason in err to its sentinel, else ErrUnknown.
func MapRevertError(err error) error {
	wrapped := (*Router)(nil).wrapRevertError(err)
	if !errors.Is(wrapped, ErrUnknown) {
		return wrapped
	}

	return (*Pair)(nil).wrapRevertError(err)
}
