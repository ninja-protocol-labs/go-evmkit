package v3

import (
	"math/big"

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
