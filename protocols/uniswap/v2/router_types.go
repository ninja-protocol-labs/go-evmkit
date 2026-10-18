package v2

import "math/big"

// AddLiquidityResult is addLiquidity()'s (amountA, amountB, liquidity) return value.
type AddLiquidityResult struct {
	AmountA   *big.Int
	AmountB   *big.Int
	Liquidity *big.Int
}

func NewAddLiquidityResult(amountA, amountB, liquidity *big.Int) *AddLiquidityResult {
	return &AddLiquidityResult{
		AmountA:   amountA,
		AmountB:   amountB,
		Liquidity: liquidity,
	}
}

// AddLiquidityETHResult is addLiquidityETH()'s (amountToken, amountETH, liquidity) return value.
type AddLiquidityETHResult struct {
	AmountToken *big.Int
	AmountETH   *big.Int
	Liquidity   *big.Int
}

func NewAddLiquidityETHResult(amountToken, amountETH, liquidity *big.Int) *AddLiquidityETHResult {
	return &AddLiquidityETHResult{
		AmountToken: amountToken,
		AmountETH:   amountETH,
		Liquidity:   liquidity,
	}
}

// RemoveLiquidityResult is removeLiquidity()'s (and removeLiquidityWithPermit()'s) (amountA,
// amountB) return value.
type RemoveLiquidityResult struct {
	AmountA *big.Int
	AmountB *big.Int
}

func NewRemoveLiquidityResult(amountA, amountB *big.Int) *RemoveLiquidityResult {
	return &RemoveLiquidityResult{
		AmountA: amountA,
		AmountB: amountB,
	}
}

// RemoveLiquidityETHResult is removeLiquidityETH()'s (and removeLiquidityETHWithPermit()'s)
// (amountToken, amountETH) return value.
type RemoveLiquidityETHResult struct {
	AmountToken *big.Int
	AmountETH   *big.Int
}

func NewRemoveLiquidityETHResult(amountToken, amountETH *big.Int) *RemoveLiquidityETHResult {
	return &RemoveLiquidityETHResult{
		AmountToken: amountToken,
		AmountETH:   amountETH,
	}
}
