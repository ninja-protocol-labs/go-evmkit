package v4

import (
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// PathKey mirrors libraries/PathKey.sol: one hop of a multi-hop route, identifying the next pool
// by its other three PoolKey fields plus the intermediate currency being swapped into.
type PathKey struct {
	IntermediateCurrency Currency
	Fee                  *big.Int
	TickSpacing          *big.Int
	Hooks                *types.Address
	HookData             []byte
}

func NewPathKey(intermediateCurrency Currency, fee, tickSpacing *big.Int, hooks *types.Address, hookData []byte) *PathKey {
	return &PathKey{
		IntermediateCurrency: intermediateCurrency,
		Fee:                  fee,
		TickSpacing:          tickSpacing,
		Hooks:                hooks,
		HookData:             hookData,
	}
}

// ExactInputSingleParams mirrors IV4Router's ExactInputSingleParams: a single-hop exact-input swap.
type ExactInputSingleParams struct {
	PoolKey          *PoolKey
	ZeroForOne       bool
	AmountIn         *big.Int
	AmountOutMinimum *big.Int
	MinHopPriceX36   *big.Int
	HookData         []byte
}

func NewExactInputSingleParams(poolKey *PoolKey, zeroForOne bool, amountIn, amountOutMinimum, minHopPriceX36 *big.Int, hookData []byte) *ExactInputSingleParams {
	return &ExactInputSingleParams{
		PoolKey:          poolKey,
		ZeroForOne:       zeroForOne,
		AmountIn:         amountIn,
		AmountOutMinimum: amountOutMinimum,
		MinHopPriceX36:   minHopPriceX36,
		HookData:         hookData,
	}
}

// ExactInputParams mirrors IV4Router's ExactInputParams: a multi-hop exact-input swap starting
// from CurrencyIn and following Path.
type ExactInputParams struct {
	CurrencyIn       Currency
	Path             []*PathKey
	MinHopPriceX36   []*big.Int
	AmountIn         *big.Int
	AmountOutMinimum *big.Int
}

func NewExactInputParams(currencyIn Currency, path []*PathKey, minHopPriceX36 []*big.Int, amountIn, amountOutMinimum *big.Int) *ExactInputParams {
	return &ExactInputParams{
		CurrencyIn:       currencyIn,
		Path:             path,
		MinHopPriceX36:   minHopPriceX36,
		AmountIn:         amountIn,
		AmountOutMinimum: amountOutMinimum,
	}
}

// ExactOutputSingleParams mirrors IV4Router's ExactOutputSingleParams: a single-hop exact-output swap.
type ExactOutputSingleParams struct {
	PoolKey         *PoolKey
	ZeroForOne      bool
	AmountOut       *big.Int
	AmountInMaximum *big.Int
	MinHopPriceX36  *big.Int
	HookData        []byte
}

func NewExactOutputSingleParams(poolKey *PoolKey, zeroForOne bool, amountOut, amountInMaximum, minHopPriceX36 *big.Int, hookData []byte) *ExactOutputSingleParams {
	return &ExactOutputSingleParams{
		PoolKey:         poolKey,
		ZeroForOne:      zeroForOne,
		AmountOut:       amountOut,
		AmountInMaximum: amountInMaximum,
		MinHopPriceX36:  minHopPriceX36,
		HookData:        hookData,
	}
}

// ExactOutputParams mirrors IV4Router's ExactOutputParams: a multi-hop exact-output swap ending
// at CurrencyOut and following Path.
type ExactOutputParams struct {
	CurrencyOut     Currency
	Path            []*PathKey
	MinHopPriceX36  []*big.Int
	AmountOut       *big.Int
	AmountInMaximum *big.Int
}

func NewExactOutputParams(currencyOut Currency, path []*PathKey, minHopPriceX36 []*big.Int, amountOut, amountInMaximum *big.Int) *ExactOutputParams {
	return &ExactOutputParams{
		CurrencyOut:     currencyOut,
		Path:            path,
		MinHopPriceX36:  minHopPriceX36,
		AmountOut:       amountOut,
		AmountInMaximum: amountInMaximum,
	}
}
