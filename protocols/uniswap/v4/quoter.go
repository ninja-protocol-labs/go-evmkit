package v4

import (
	"context"
	"fmt"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

// IQuoterImmutables mirrors IImmutableState plus IMsgSender, the only state V4Quoter exposes.
type IQuoterImmutables interface {
	PoolManager(ctx context.Context, opts *bind.CallOptions, req *PoolManagerRequest) (*types.Address, error)
	MsgSender(ctx context.Context, opts *bind.CallOptions, req *MsgSenderRequest) (*types.Address, error)
}

// IQuoterQuotes mirrors IV4Quoter's quoting functions; they only simulate, so only eth_call-backed Call* methods exist.
type IQuoterQuotes interface {
	CallQuoteExactInputSingle(ctx context.Context, opts *bind.CallOptions, req *QuoteExactInputSingleRequest) (*QuoteExactInputResult, error)
	CallQuoteExactInput(ctx context.Context, opts *bind.CallOptions, req *QuoteExactInputRequest) (*QuoteExactInputResult, error)
	CallQuoteExactOutputSingle(ctx context.Context, opts *bind.CallOptions, req *QuoteExactOutputSingleRequest) (*QuoteExactOutputResult, error)
	CallQuoteExactOutput(ctx context.Context, opts *bind.CallOptions, req *QuoteExactOutputRequest) (*QuoteExactOutputResult, error)
}

// IQuoter mirrors IV4Quoter.
type IQuoter interface {
	IQuoterImmutables
	IQuoterQuotes
}

var _ IQuoter = (*Quoter)(nil)

// Quoter calls any V4Quoter deployment over a Client; it holds no contract address.
type Quoter struct {
	c rpc.Client
}

// NewQuoter returns a Quoter calling V4Quoter deployments via cli.
func NewQuoter(cli rpc.Client) *Quoter {
	return &Quoter{
		c: cli,
	}
}

// PoolManager returns the PoolManager req.Contract quotes against.
func (q *Quoter) PoolManager(ctx context.Context, opts *bind.CallOptions, req *PoolManagerRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: quoterPoolManager: %w", err)
	}

	data, err := bind.Call(ctx, q.c, opts, req.Contract, EncodeQuoterPoolManager())
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: quoterPoolManager: %w", mapQuoterRevertError(err))
	}

	out, err := DecodeQuoterPoolManager(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: quoterPoolManager: %w", err)
	}

	return out, nil
}

// MsgSender returns req.Contract's current msgSender, the zero address outside a quote.
func (q *Quoter) MsgSender(ctx context.Context, opts *bind.CallOptions, req *MsgSenderRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: quoterMsgSender: %w", err)
	}

	data, err := bind.Call(ctx, q.c, opts, req.Contract, EncodeQuoterMsgSender())
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: quoterMsgSender: %w", mapQuoterRevertError(err))
	}

	out, err := DecodeQuoterMsgSender(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: quoterMsgSender: %w", err)
	}

	return out, nil
}

// CallQuoteExactInputSingle runs quoteExactInputSingle via eth_call as opts.From, returning the output for req.Params.ExactAmount in or the revert.
func (q *Quoter) CallQuoteExactInputSingle(ctx context.Context, opts *bind.CallOptions, req *QuoteExactInputSingleRequest) (*QuoteExactInputResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactInputSingle: %w", err)
	}

	calldata, err := EncodeQuoteExactInputSingle(req.Params)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactInputSingle: %w", err)
	}

	data, err := bind.Call(ctx, q.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactInputSingle: %w", mapQuoterRevertError(err))
	}

	out, err := DecodeQuoteExactInputSingle(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactInputSingle: %w", err)
	}

	return out, nil
}

// CallQuoteExactInput runs quoteExactInput via eth_call as opts.From, returning the final output for req.Params.ExactAmount of ExactCurrency in or the revert.
func (q *Quoter) CallQuoteExactInput(ctx context.Context, opts *bind.CallOptions, req *QuoteExactInputRequest) (*QuoteExactInputResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactInput: %w", err)
	}

	calldata, err := EncodeQuoteExactInput(req.Params)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactInput: %w", err)
	}

	data, err := bind.Call(ctx, q.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactInput: %w", mapQuoterRevertError(err))
	}

	out, err := DecodeQuoteExactInput(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactInput: %w", err)
	}

	return out, nil
}

// CallQuoteExactOutputSingle runs quoteExactOutputSingle via eth_call as opts.From, returning the input required for req.Params.ExactAmount out or the revert.
func (q *Quoter) CallQuoteExactOutputSingle(ctx context.Context, opts *bind.CallOptions, req *QuoteExactOutputSingleRequest) (*QuoteExactOutputResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactOutputSingle: %w", err)
	}

	calldata, err := EncodeQuoteExactOutputSingle(req.Params)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactOutputSingle: %w", err)
	}

	data, err := bind.Call(ctx, q.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactOutputSingle: %w", mapQuoterRevertError(err))
	}

	out, err := DecodeQuoteExactOutputSingle(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactOutputSingle: %w", err)
	}

	return out, nil
}

// CallQuoteExactOutput runs quoteExactOutput via eth_call as opts.From, returning the input required for req.Params.ExactAmount of ExactCurrency out or the revert.
func (q *Quoter) CallQuoteExactOutput(ctx context.Context, opts *bind.CallOptions, req *QuoteExactOutputRequest) (*QuoteExactOutputResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactOutput: %w", err)
	}

	calldata, err := EncodeQuoteExactOutput(req.Params)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactOutput: %w", err)
	}

	data, err := bind.Call(ctx, q.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactOutput: %w", mapQuoterRevertError(err))
	}

	out, err := DecodeQuoteExactOutput(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactOutput: %w", err)
	}

	return out, nil
}
