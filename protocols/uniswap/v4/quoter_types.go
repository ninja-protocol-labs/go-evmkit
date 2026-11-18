package v4

import (
	"math/big"
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
