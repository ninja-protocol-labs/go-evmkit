package v3

import (
	"context"
	"fmt"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

// IQuoterImmutables mirrors PeripheryImmutableState, the only state QuoterV2 itself holds.
type IQuoterImmutables interface {
	Factory(ctx context.Context, opts *bind.CallOptions, req *FactoryRequest) (*types.Address, error)
	WETH9(ctx context.Context, opts *bind.CallOptions, req *WETH9Request) (*types.Address, error)
}

// IQuoterQuotes mirrors IQuoterV2's quoting functions, which are only ever simulated via eth_call.
type IQuoterQuotes interface {
	CallQuoteExactInputSingle(ctx context.Context, opts *bind.CallOptions, req *QuoteExactInputSingleRequest) (*QuoteExactInputSingleResult, error)
	CallQuoteExactInput(ctx context.Context, opts *bind.CallOptions, req *QuoteExactInputRequest) (*QuoteExactInputResult, error)
	CallQuoteExactOutputSingle(ctx context.Context, opts *bind.CallOptions, req *QuoteExactOutputSingleRequest) (*QuoteExactOutputSingleResult, error)
	CallQuoteExactOutput(ctx context.Context, opts *bind.CallOptions, req *QuoteExactOutputRequest) (*QuoteExactOutputResult, error)
}

// IQuoter mirrors IQuoterV2.
type IQuoter interface {
	IQuoterImmutables
	IQuoterQuotes
}

var _ IQuoter = (*Quoter)(nil)

// Quoter calls any QuoterV2 contract over a Client; it holds no contract address, so one instance serves every deployment.
type Quoter struct {
	c rpc.Client
}

// NewQuoter returns a Quoter calling quoters via cli.
func NewQuoter(cli rpc.Client) *Quoter {
	return &Quoter{
		c: cli,
	}
}

// Factory returns the UniswapV3Factory req.Contract quotes against.
func (q *Quoter) Factory(ctx context.Context, opts *bind.CallOptions, req *FactoryRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: factory: %w", err)
	}

	data, err := bind.Call(ctx, q.c, opts, req.Contract, EncodeQuoterFactory())
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: factory: %w", q.wrapRevertError(err))
	}

	out, err := DecodeQuoterFactory(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: factory: %w", err)
	}

	return out, nil
}

// WETH9 returns req.Contract's wrapped native token.
func (q *Quoter) WETH9(ctx context.Context, opts *bind.CallOptions, req *WETH9Request) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: weth9: %w", err)
	}

	data, err := bind.Call(ctx, q.c, opts, req.Contract, EncodeQuoterWETH9())
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: weth9: %w", q.wrapRevertError(err))
	}

	out, err := DecodeQuoterWETH9(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: weth9: %w", err)
	}

	return out, nil
}

// CallQuoteExactInputSingle quotes a single-pool exact-input swap via eth_call as opts.From.
func (q *Quoter) CallQuoteExactInputSingle(ctx context.Context, opts *bind.CallOptions, req *QuoteExactInputSingleRequest) (*QuoteExactInputSingleResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactInputSingle: %w", err)
	}

	calldata, err := EncodeQuoteExactInputSingle(req.Params.TokenIn, req.Params.TokenOut, req.Params.AmountIn, req.Params.Fee, req.Params.SqrtPriceLimitX96)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactInputSingle: %w", err)
	}

	data, err := bind.Call(ctx, q.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactInputSingle: %w", q.wrapRevertError(err))
	}

	out, err := DecodeQuoteExactInputSingle(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactInputSingle: %w", err)
	}

	return out, nil
}

// CallQuoteExactInput quotes an exact-input swap along req.Path (see EncodePath) via eth_call as opts.From.
func (q *Quoter) CallQuoteExactInput(ctx context.Context, opts *bind.CallOptions, req *QuoteExactInputRequest) (*QuoteExactInputResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactInput: %w", err)
	}

	calldata, err := EncodeQuoteExactInput(req.Path, req.AmountIn)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactInput: %w", err)
	}

	data, err := bind.Call(ctx, q.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactInput: %w", q.wrapRevertError(err))
	}

	out, err := DecodeQuoteExactInput(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactInput: %w", err)
	}

	return out, nil
}

// CallQuoteExactOutputSingle quotes a single-pool exact-output swap via eth_call as opts.From.
func (q *Quoter) CallQuoteExactOutputSingle(ctx context.Context, opts *bind.CallOptions, req *QuoteExactOutputSingleRequest) (*QuoteExactOutputSingleResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactOutputSingle: %w", err)
	}

	calldata, err := EncodeQuoteExactOutputSingle(req.Params.TokenIn, req.Params.TokenOut, req.Params.Amount, req.Params.Fee, req.Params.SqrtPriceLimitX96)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactOutputSingle: %w", err)
	}

	data, err := bind.Call(ctx, q.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactOutputSingle: %w", q.wrapRevertError(err))
	}

	out, err := DecodeQuoteExactOutputSingle(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactOutputSingle: %w", err)
	}

	return out, nil
}

// CallQuoteExactOutput quotes an exact-output swap along req.Path, which runs tokenOut to tokenIn, via eth_call as opts.From.
func (q *Quoter) CallQuoteExactOutput(ctx context.Context, opts *bind.CallOptions, req *QuoteExactOutputRequest) (*QuoteExactOutputResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactOutput: %w", err)
	}

	calldata, err := EncodeQuoteExactOutput(req.Path, req.AmountOut)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactOutput: %w", err)
	}

	data, err := bind.Call(ctx, q.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactOutput: %w", q.wrapRevertError(err))
	}

	out, err := DecodeQuoteExactOutput(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactOutput: %w", err)
	}

	return out, nil
}

// wrapRevertError maps a revert in err the same way Pool does, since QuoterV2 bubbles up the pool's own revert reasons.
func (_ *Quoter) wrapRevertError(err error) error {
	return (*Pool)(nil).wrapRevertError(err)
}
