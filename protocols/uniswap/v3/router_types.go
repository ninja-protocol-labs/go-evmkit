package v3

import (
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// ExactInputSingleParams mirrors IV3SwapRouter.ExactInputSingleParams.
type ExactInputSingleParams struct {
	TokenIn           *types.Address
	TokenOut          *types.Address
	Fee               *big.Int
	Recipient         *types.Address
	AmountIn          *big.Int
	AmountOutMinimum  *big.Int
	SqrtPriceLimitX96 *big.Int
}

func NewExactInputSingleParams(
	tokenIn, tokenOut *types.Address,
	fee *big.Int,
	recipient *types.Address,
	amountIn, amountOutMinimum, sqrtPriceLimitX96 *big.Int,
) *ExactInputSingleParams {
	return &ExactInputSingleParams{
		TokenIn:           tokenIn,
		TokenOut:          tokenOut,
		Fee:               fee,
		Recipient:         recipient,
		AmountIn:          amountIn,
		AmountOutMinimum:  amountOutMinimum,
		SqrtPriceLimitX96: sqrtPriceLimitX96,
	}
}

// ExactInputParams mirrors IV3SwapRouter.ExactInputParams.
type ExactInputParams struct {
	Path             []byte
	Recipient        *types.Address
	AmountIn         *big.Int
	AmountOutMinimum *big.Int
}

func NewExactInputParams(
	path []byte,
	recipient *types.Address,
	amountIn, amountOutMinimum *big.Int,
) *ExactInputParams {
	return &ExactInputParams{
		Path:             path,
		Recipient:        recipient,
		AmountIn:         amountIn,
		AmountOutMinimum: amountOutMinimum,
	}
}

// ExactOutputSingleParams mirrors IV3SwapRouter.ExactOutputSingleParams.
type ExactOutputSingleParams struct {
	TokenIn           *types.Address
	TokenOut          *types.Address
	Fee               *big.Int
	Recipient         *types.Address
	AmountOut         *big.Int
	AmountInMaximum   *big.Int
	SqrtPriceLimitX96 *big.Int
}

func NewExactOutputSingleParams(
	tokenIn, tokenOut *types.Address,
	fee *big.Int,
	recipient *types.Address,
	amountOut, amountInMaximum, sqrtPriceLimitX96 *big.Int,
) *ExactOutputSingleParams {
	return &ExactOutputSingleParams{
		TokenIn:           tokenIn,
		TokenOut:          tokenOut,
		Fee:               fee,
		Recipient:         recipient,
		AmountOut:         amountOut,
		AmountInMaximum:   amountInMaximum,
		SqrtPriceLimitX96: sqrtPriceLimitX96,
	}
}

// ExactOutputParams mirrors IV3SwapRouter.ExactOutputParams.
type ExactOutputParams struct {
	Path            []byte
	Recipient       *types.Address
	AmountOut       *big.Int
	AmountInMaximum *big.Int
}

func NewExactOutputParams(
	path []byte,
	recipient *types.Address,
	amountOut, amountInMaximum *big.Int,
) *ExactOutputParams {
	return &ExactOutputParams{
		Path:            path,
		Recipient:       recipient,
		AmountOut:       amountOut,
		AmountInMaximum: amountInMaximum,
	}
}
