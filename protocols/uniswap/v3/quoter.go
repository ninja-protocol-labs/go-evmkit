package v3

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// IQuoterImmutables mirrors PeripheryImmutableState, the only state QuoterV2 itself holds.
type IQuoterImmutables interface {
	Factory(ctx context.Context, contract *types.Address, block string) (*types.Address, error)
	WETH9(ctx context.Context, contract *types.Address, block string) (*types.Address, error)
}

// IQuoterQuotes mirrors IQuoterV2's own quoting functions. These are not gas-efficient and must
// never be broadcast as a real transaction -- they only simulate a swap via eth_call and parse
// the result out of a deliberate revert, which is all the Call* methods below do.
type IQuoterQuotes interface {
	CallQuoteExactInputSingle(ctx context.Context, contract, sender *types.Address, params *QuoteExactInputSingleParams, block string) (*QuoteExactInputSingleResult, error)
	CallQuoteExactInput(ctx context.Context, contract, sender *types.Address, path []byte, amountIn *big.Int, block string) (*QuoteExactInputResult, error)
	CallQuoteExactOutputSingle(ctx context.Context, contract, sender *types.Address, params *QuoteExactOutputSingleParams, block string) (*QuoteExactOutputSingleResult, error)
	CallQuoteExactOutput(ctx context.Context, contract, sender *types.Address, path []byte, amountOut *big.Int, block string) (*QuoteExactOutputResult, error)
}

// IQuoter mirrors IQuoterV2.
type IQuoter interface {
	IQuoterImmutables
	IQuoterQuotes
}

var _ IQuoter = (*Quoter)(nil)

// Quoter calls any QuoterV2 contract over a Client; it holds no contract address, so one
// instance serves every network's deployment.
type Quoter struct {
	c rpc.Client
}

// NewQuoter returns a Quoter calling QuoterV2 contracts via c.
func NewQuoter(c rpc.Client) *Quoter {
	return &Quoter{
		c: c,
	}
}

// Factory returns contract's immutable UniswapV3Factory address.
func (q *Quoter) Factory(ctx context.Context, contract *types.Address, block string) (*types.Address, error) {
	data, err := q.call(ctx, nil, contract, EncodeQuoterFactory(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: factory: %w", err)
	}
	addr, err := DecodeQuoterFactory(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: factory: %w", err)
	}
	return addr, nil
}

// WETH9 returns contract's immutable WETH9 address.
func (q *Quoter) WETH9(ctx context.Context, contract *types.Address, block string) (*types.Address, error) {
	data, err := q.call(ctx, nil, contract, EncodeQuoterWETH9(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: weth9: %w", err)
	}
	addr, err := DecodeQuoterWETH9(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: weth9: %w", err)
	}
	return addr, nil
}

// CallQuoteExactInputSingle simulates quoteExactInputSingle(params) via eth_call, as sender,
// returning the amount of tokenOut that would be received -- without executing a real swap.
func (q *Quoter) CallQuoteExactInputSingle(
	ctx context.Context,
	contract, sender *types.Address,
	params *QuoteExactInputSingleParams,
	block string,
) (*QuoteExactInputSingleResult, error) {
	calldata, err := EncodeQuoteExactInputSingle(params.TokenIn, params.TokenOut, params.AmountIn, params.Fee, params.SqrtPriceLimitX96)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactInputSingle: %w", err)
	}
	data, err := q.call(ctx, sender, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactInputSingle: %w", err)
	}
	result, err := DecodeQuoteExactInputSingle(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactInputSingle: %w", err)
	}
	return result, nil
}

// CallQuoteExactInput simulates quoteExactInput(path, amountIn) via eth_call, as sender,
// returning the amount of the final token in path that would be received -- without executing a
// real swap.
func (q *Quoter) CallQuoteExactInput(
	ctx context.Context,
	contract, sender *types.Address,
	path []byte,
	amountIn *big.Int,
	block string,
) (*QuoteExactInputResult, error) {
	calldata, err := EncodeQuoteExactInput(path, amountIn)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactInput: %w", err)
	}
	data, err := q.call(ctx, sender, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactInput: %w", err)
	}
	result, err := DecodeQuoteExactInput(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactInput: %w", err)
	}
	return result, nil
}

// CallQuoteExactOutputSingle simulates quoteExactOutputSingle(params) via eth_call, as sender,
// returning the amount of tokenIn that would be required -- without executing a real swap.
func (q *Quoter) CallQuoteExactOutputSingle(
	ctx context.Context,
	contract, sender *types.Address,
	params *QuoteExactOutputSingleParams,
	block string,
) (*QuoteExactOutputSingleResult, error) {
	calldata, err := EncodeQuoteExactOutputSingle(params.TokenIn, params.TokenOut, params.Amount, params.Fee, params.SqrtPriceLimitX96)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactOutputSingle: %w", err)
	}
	data, err := q.call(ctx, sender, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactOutputSingle: %w", err)
	}
	result, err := DecodeQuoteExactOutputSingle(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactOutputSingle: %w", err)
	}
	return result, nil
}

// CallQuoteExactOutput simulates quoteExactOutput(path, amountOut) via eth_call, as sender,
// returning the amount of the first token in path that would be required -- without executing a
// real swap. path must be provided in reverse order (the final output token first).
func (q *Quoter) CallQuoteExactOutput(
	ctx context.Context,
	contract, sender *types.Address,
	path []byte,
	amountOut *big.Int,
	block string,
) (*QuoteExactOutputResult, error) {
	calldata, err := EncodeQuoteExactOutput(path, amountOut)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactOutput: %w", err)
	}
	data, err := q.call(ctx, sender, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactOutput: %w", err)
	}
	result, err := DecodeQuoteExactOutput(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callQuoteExactOutput: %w", err)
	}
	return result, nil
}

// call eth_calls contract with calldata, optionally simulating from a specific sender. QuoterV2's
// own functions are never payable, so unlike Router/NFPM's call there's no value to attach.
func (q *Quoter) call(
	ctx context.Context,
	from, contract *types.Address,
	calldata []byte,
	block string,
) ([]byte, error) {
	params := map[string]any{
		"to":   contract.String(),
		"data": encoding.Hex.EncodePrefixed(calldata),
	}
	if from != nil {
		params["from"] = from.String()
	}

	var raw string
	if err := q.c.Call(ctx, rpc.ETHCall("call", params, block, &raw)); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrReverted, err)
	}
	return encoding.Hex.Decode(raw)
}
