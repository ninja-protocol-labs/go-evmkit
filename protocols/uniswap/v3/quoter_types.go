package v3

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// QuoteExactInputSingleParams mirrors IQuoterV2.QuoteExactInputSingleParams.
type QuoteExactInputSingleParams struct {
	TokenIn           *types.Address
	TokenOut          *types.Address
	AmountIn          *big.Int
	Fee               *big.Int
	SqrtPriceLimitX96 *big.Int
}

func NewQuoteExactInputSingleParams(
	tokenIn, tokenOut *types.Address,
	amountIn, fee, sqrtPriceLimitX96 *big.Int,
) *QuoteExactInputSingleParams {
	return &QuoteExactInputSingleParams{
		TokenIn:           tokenIn,
		TokenOut:          tokenOut,
		AmountIn:          amountIn,
		Fee:               fee,
		SqrtPriceLimitX96: sqrtPriceLimitX96,
	}
}

// QuoteExactOutputSingleParams mirrors IQuoterV2.QuoteExactOutputSingleParams.
type QuoteExactOutputSingleParams struct {
	TokenIn           *types.Address
	TokenOut          *types.Address
	Amount            *big.Int
	Fee               *big.Int
	SqrtPriceLimitX96 *big.Int
}

func NewQuoteExactOutputSingleParams(
	tokenIn, tokenOut *types.Address,
	amount, fee, sqrtPriceLimitX96 *big.Int,
) *QuoteExactOutputSingleParams {
	return &QuoteExactOutputSingleParams{
		TokenIn:           tokenIn,
		TokenOut:          tokenOut,
		Amount:            amount,
		Fee:               fee,
		SqrtPriceLimitX96: sqrtPriceLimitX96,
	}
}

// QuoteExactInputSingleRequest is quoteExactInputSingle(params)'s input on Contract.
type QuoteExactInputSingleRequest struct {
	Contract *types.Address
	Params   *QuoteExactInputSingleParams
}

func NewQuoteExactInputSingleRequest(contract *types.Address, params *QuoteExactInputSingleParams) *QuoteExactInputSingleRequest {
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

	if err := validateQuoteExactInputSingleParams(r.Params); err != nil {
		return err
	}

	return nil
}

// QuoteExactInputRequest is quoteExactInput(path, amountIn)'s input on Contract.
type QuoteExactInputRequest struct {
	Contract *types.Address
	Path     []byte
	AmountIn *big.Int
}

func NewQuoteExactInputRequest(contract *types.Address, path []byte, amountIn *big.Int) *QuoteExactInputRequest {
	return &QuoteExactInputRequest{
		Contract: contract,
		Path:     path,
		AmountIn: amountIn,
	}
}

func (r *QuoteExactInputRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if len(r.Path) == 0 {
		return fmt.Errorf("%w: path is empty", bind.ErrInvalidRequest)
	}

	if r.AmountIn == nil {
		return fmt.Errorf("%w: amountIn is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// QuoteExactOutputSingleRequest is quoteExactOutputSingle(params)'s input on Contract.
type QuoteExactOutputSingleRequest struct {
	Contract *types.Address
	Params   *QuoteExactOutputSingleParams
}

func NewQuoteExactOutputSingleRequest(contract *types.Address, params *QuoteExactOutputSingleParams) *QuoteExactOutputSingleRequest {
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

	if err := validateQuoteExactOutputSingleParams(r.Params); err != nil {
		return err
	}

	return nil
}

// QuoteExactOutputRequest is quoteExactOutput(path, amountOut)'s input on Contract.
type QuoteExactOutputRequest struct {
	Contract  *types.Address
	Path      []byte
	AmountOut *big.Int
}

func NewQuoteExactOutputRequest(contract *types.Address, path []byte, amountOut *big.Int) *QuoteExactOutputRequest {
	return &QuoteExactOutputRequest{
		Contract:  contract,
		Path:      path,
		AmountOut: amountOut,
	}
}

func (r *QuoteExactOutputRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if len(r.Path) == 0 {
		return fmt.Errorf("%w: path is empty", bind.ErrInvalidRequest)
	}

	if r.AmountOut == nil {
		return fmt.Errorf("%w: amountOut is nil", bind.ErrInvalidRequest)
	}

	return nil
}

func validateQuoteExactInputSingleParams(p *QuoteExactInputSingleParams) error {
	if p == nil {
		return fmt.Errorf("%w: params is nil", bind.ErrInvalidRequest)
	}

	if p.TokenIn == nil {
		return fmt.Errorf("%w: params.tokenIn is nil", bind.ErrInvalidRequest)
	}

	if p.TokenOut == nil {
		return fmt.Errorf("%w: params.tokenOut is nil", bind.ErrInvalidRequest)
	}

	if p.AmountIn == nil {
		return fmt.Errorf("%w: params.amountIn is nil", bind.ErrInvalidRequest)
	}

	if p.Fee == nil {
		return fmt.Errorf("%w: params.fee is nil", bind.ErrInvalidRequest)
	}

	if p.SqrtPriceLimitX96 == nil {
		return fmt.Errorf("%w: params.sqrtPriceLimitX96 is nil", bind.ErrInvalidRequest)
	}

	return nil
}

func validateQuoteExactOutputSingleParams(p *QuoteExactOutputSingleParams) error {
	if p == nil {
		return fmt.Errorf("%w: params is nil", bind.ErrInvalidRequest)
	}

	if p.TokenIn == nil {
		return fmt.Errorf("%w: params.tokenIn is nil", bind.ErrInvalidRequest)
	}

	if p.TokenOut == nil {
		return fmt.Errorf("%w: params.tokenOut is nil", bind.ErrInvalidRequest)
	}

	if p.Amount == nil {
		return fmt.Errorf("%w: params.amount is nil", bind.ErrInvalidRequest)
	}

	if p.Fee == nil {
		return fmt.Errorf("%w: params.fee is nil", bind.ErrInvalidRequest)
	}

	if p.SqrtPriceLimitX96 == nil {
		return fmt.Errorf("%w: params.sqrtPriceLimitX96 is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// QuoteExactInputSingleResult is quoteExactInputSingle's return value.
type QuoteExactInputSingleResult struct {
	AmountOut               *big.Int
	SqrtPriceX96After       *big.Int
	InitializedTicksCrossed *big.Int
	GasEstimate             *big.Int
}

func NewQuoteExactInputSingleResult(amountOut, sqrtPriceX96After, initializedTicksCrossed, gasEstimate *big.Int) *QuoteExactInputSingleResult {
	return &QuoteExactInputSingleResult{
		AmountOut:               amountOut,
		SqrtPriceX96After:       sqrtPriceX96After,
		InitializedTicksCrossed: initializedTicksCrossed,
		GasEstimate:             gasEstimate,
	}
}

// QuoteExactInputResult is quoteExactInput's return value.
type QuoteExactInputResult struct {
	AmountOut                   *big.Int
	SqrtPriceX96AfterList       []*big.Int
	InitializedTicksCrossedList []*big.Int
	GasEstimate                 *big.Int
}

func NewQuoteExactInputResult(amountOut *big.Int, sqrtPriceX96AfterList, initializedTicksCrossedList []*big.Int, gasEstimate *big.Int) *QuoteExactInputResult {
	return &QuoteExactInputResult{
		AmountOut:                   amountOut,
		SqrtPriceX96AfterList:       sqrtPriceX96AfterList,
		InitializedTicksCrossedList: initializedTicksCrossedList,
		GasEstimate:                 gasEstimate,
	}
}

// QuoteExactOutputSingleResult is quoteExactOutputSingle's return value.
type QuoteExactOutputSingleResult struct {
	AmountIn                *big.Int
	SqrtPriceX96After       *big.Int
	InitializedTicksCrossed *big.Int
	GasEstimate             *big.Int
}

func NewQuoteExactOutputSingleResult(amountIn, sqrtPriceX96After, initializedTicksCrossed, gasEstimate *big.Int) *QuoteExactOutputSingleResult {
	return &QuoteExactOutputSingleResult{
		AmountIn:                amountIn,
		SqrtPriceX96After:       sqrtPriceX96After,
		InitializedTicksCrossed: initializedTicksCrossed,
		GasEstimate:             gasEstimate,
	}
}

// QuoteExactOutputResult is quoteExactOutput's return value.
type QuoteExactOutputResult struct {
	AmountIn                    *big.Int
	SqrtPriceX96AfterList       []*big.Int
	InitializedTicksCrossedList []*big.Int
	GasEstimate                 *big.Int
}

func NewQuoteExactOutputResult(amountIn *big.Int, sqrtPriceX96AfterList, initializedTicksCrossedList []*big.Int, gasEstimate *big.Int) *QuoteExactOutputResult {
	return &QuoteExactOutputResult{
		AmountIn:                    amountIn,
		SqrtPriceX96AfterList:       sqrtPriceX96AfterList,
		InitializedTicksCrossedList: initializedTicksCrossedList,
		GasEstimate:                 gasEstimate,
	}
}
