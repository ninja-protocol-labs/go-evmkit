package v4

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// QuoteExactSingleParams mirrors IV4Quoter.QuoteExactSingleParams: a single-pool quote.
type QuoteExactSingleParams struct {
	PoolKey     *PoolKey
	ZeroForOne  bool
	ExactAmount *big.Int
	HookData    []byte
}

func NewQuoteExactSingleParams(poolKey *PoolKey, zeroForOne bool, exactAmount *big.Int, hookData []byte) *QuoteExactSingleParams {
	return &QuoteExactSingleParams{
		PoolKey:     poolKey,
		ZeroForOne:  zeroForOne,
		ExactAmount: exactAmount,
		HookData:    hookData,
	}
}

// QuoteExactParams mirrors IV4Quoter.QuoteExactParams: a multi-hop quote starting from ExactCurrency.
type QuoteExactParams struct {
	ExactCurrency Currency
	Path          []*PathKey
	ExactAmount   *big.Int
}

func NewQuoteExactParams(exactCurrency Currency, path []*PathKey, exactAmount *big.Int) *QuoteExactParams {
	return &QuoteExactParams{
		ExactCurrency: exactCurrency,
		Path:          path,
		ExactAmount:   exactAmount,
	}
}

// QuoteExactInputResult is quoteExactInputSingle's and quoteExactInput's return value.
type QuoteExactInputResult struct {
	AmountOut   *big.Int
	GasEstimate *big.Int
}

func NewQuoteExactInputResult(amountOut, gasEstimate *big.Int) *QuoteExactInputResult {
	return &QuoteExactInputResult{
		AmountOut:   amountOut,
		GasEstimate: gasEstimate,
	}
}

// QuoteExactOutputResult is quoteExactOutputSingle's and quoteExactOutput's return value.
type QuoteExactOutputResult struct {
	AmountIn    *big.Int
	GasEstimate *big.Int
}

func NewQuoteExactOutputResult(amountIn, gasEstimate *big.Int) *QuoteExactOutputResult {
	return &QuoteExactOutputResult{
		AmountIn:    amountIn,
		GasEstimate: gasEstimate,
	}
}

// MsgSenderRequest is msgSender()'s input on Contract.
type MsgSenderRequest struct {
	Contract *types.Address
}

func NewMsgSenderRequest(contract *types.Address) *MsgSenderRequest {
	return &MsgSenderRequest{
		Contract: contract,
	}
}

func (r *MsgSenderRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// QuoteExactInputSingleRequest is quoteExactInputSingle(params)'s input on Contract.
type QuoteExactInputSingleRequest struct {
	Contract *types.Address
	Params   *QuoteExactSingleParams
}

func NewQuoteExactInputSingleRequest(contract *types.Address, params *QuoteExactSingleParams) *QuoteExactInputSingleRequest {
	return &QuoteExactInputSingleRequest{
		Contract: contract,
		Params:   params,
	}
}

func (r *QuoteExactInputSingleRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return validateQuoteExactSingleParams(r.Params)
}

// QuoteExactInputRequest is quoteExactInput(params)'s input on Contract.
type QuoteExactInputRequest struct {
	Contract *types.Address
	Params   *QuoteExactParams
}

func NewQuoteExactInputRequest(contract *types.Address, params *QuoteExactParams) *QuoteExactInputRequest {
	return &QuoteExactInputRequest{
		Contract: contract,
		Params:   params,
	}
}

func (r *QuoteExactInputRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return validateQuoteExactParams(r.Params)
}

// QuoteExactOutputSingleRequest is quoteExactOutputSingle(params)'s input on Contract.
type QuoteExactOutputSingleRequest struct {
	Contract *types.Address
	Params   *QuoteExactSingleParams
}

func NewQuoteExactOutputSingleRequest(contract *types.Address, params *QuoteExactSingleParams) *QuoteExactOutputSingleRequest {
	return &QuoteExactOutputSingleRequest{
		Contract: contract,
		Params:   params,
	}
}

func (r *QuoteExactOutputSingleRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return validateQuoteExactSingleParams(r.Params)
}

// QuoteExactOutputRequest is quoteExactOutput(params)'s input on Contract.
type QuoteExactOutputRequest struct {
	Contract *types.Address
	Params   *QuoteExactParams
}

func NewQuoteExactOutputRequest(contract *types.Address, params *QuoteExactParams) *QuoteExactOutputRequest {
	return &QuoteExactOutputRequest{
		Contract: contract,
		Params:   params,
	}
}

func (r *QuoteExactOutputRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return validateQuoteExactParams(r.Params)
}

func validateQuoteExactSingleParams(p *QuoteExactSingleParams) error {
	if p == nil {
		return fmt.Errorf("%w: params is nil", bind.ErrInvalidRequest)
	}

	if p.PoolKey == nil {
		return fmt.Errorf("%w: params.poolKey is nil", bind.ErrInvalidRequest)
	}

	if p.PoolKey.Currency0 == nil {
		return fmt.Errorf("%w: params.poolKey.currency0 is nil", bind.ErrInvalidRequest)
	}

	if p.PoolKey.Currency1 == nil {
		return fmt.Errorf("%w: params.poolKey.currency1 is nil", bind.ErrInvalidRequest)
	}

	if p.PoolKey.Fee == nil {
		return fmt.Errorf("%w: params.poolKey.fee is nil", bind.ErrInvalidRequest)
	}

	if p.PoolKey.TickSpacing == nil {
		return fmt.Errorf("%w: params.poolKey.tickSpacing is nil", bind.ErrInvalidRequest)
	}

	if p.PoolKey.Hooks == nil {
		return fmt.Errorf("%w: params.poolKey.hooks is nil", bind.ErrInvalidRequest)
	}

	if p.ExactAmount == nil {
		return fmt.Errorf("%w: params.exactAmount is nil", bind.ErrInvalidRequest)
	}

	return nil
}

func validateQuoteExactParams(p *QuoteExactParams) error {
	if p == nil {
		return fmt.Errorf("%w: params is nil", bind.ErrInvalidRequest)
	}

	if p.ExactCurrency == nil {
		return fmt.Errorf("%w: params.exactCurrency is nil", bind.ErrInvalidRequest)
	}

	for i, k := range p.Path {
		if k == nil {
			return fmt.Errorf("%w: params.path[%d] is nil", bind.ErrInvalidRequest, i)
		}

		if k.IntermediateCurrency == nil {
			return fmt.Errorf("%w: params.path[%d].intermediateCurrency is nil", bind.ErrInvalidRequest, i)
		}

		if k.Fee == nil {
			return fmt.Errorf("%w: params.path[%d].fee is nil", bind.ErrInvalidRequest, i)
		}

		if k.TickSpacing == nil {
			return fmt.Errorf("%w: params.path[%d].tickSpacing is nil", bind.ErrInvalidRequest, i)
		}

		if k.Hooks == nil {
			return fmt.Errorf("%w: params.path[%d].hooks is nil", bind.ErrInvalidRequest, i)
		}
	}

	if p.ExactAmount == nil {
		return fmt.Errorf("%w: params.exactAmount is nil", bind.ErrInvalidRequest)
	}

	return nil
}
