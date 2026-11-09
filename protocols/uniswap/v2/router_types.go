package v2

import (
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// QuoteParams is Quote's argument list.
type QuoteParams struct {
	Router   *types.Address
	AmountA  *big.Int
	ReserveA *big.Int
	ReserveB *big.Int
}

func NewQuoteParams(router *types.Address, amountA, reserveA, reserveB *big.Int) *QuoteParams {
	return &QuoteParams{
		Router:   router,
		AmountA:  amountA,
		ReserveA: reserveA,
		ReserveB: reserveB,
	}
}

// GetAmountOutParams is GetAmountOut's argument list.
type GetAmountOutParams struct {
	Router     *types.Address
	AmountIn   *big.Int
	ReserveIn  *big.Int
	ReserveOut *big.Int
}

func NewGetAmountOutParams(router *types.Address, amountIn, reserveIn, reserveOut *big.Int) *GetAmountOutParams {
	return &GetAmountOutParams{
		Router:     router,
		AmountIn:   amountIn,
		ReserveIn:  reserveIn,
		ReserveOut: reserveOut,
	}
}

// GetAmountInParams is GetAmountIn's argument list.
type GetAmountInParams struct {
	Router     *types.Address
	AmountOut  *big.Int
	ReserveIn  *big.Int
	ReserveOut *big.Int
}

func NewGetAmountInParams(router *types.Address, amountOut, reserveIn, reserveOut *big.Int) *GetAmountInParams {
	return &GetAmountInParams{
		Router:     router,
		AmountOut:  amountOut,
		ReserveIn:  reserveIn,
		ReserveOut: reserveOut,
	}
}

// GetAmountsOutParams is GetAmountsOut's argument list.
type GetAmountsOutParams struct {
	Router   *types.Address
	AmountIn *big.Int
	Path     []*types.Address
}

func NewGetAmountsOutParams(router *types.Address, amountIn *big.Int, path []*types.Address) *GetAmountsOutParams {
	return &GetAmountsOutParams{
		Router:   router,
		AmountIn: amountIn,
		Path:     path,
	}
}

// GetAmountsInParams is GetAmountsIn's argument list.
type GetAmountsInParams struct {
	Router    *types.Address
	AmountOut *big.Int
	Path      []*types.Address
}

func NewGetAmountsInParams(router *types.Address, amountOut *big.Int, path []*types.Address) *GetAmountsInParams {
	return &GetAmountsInParams{
		Router:    router,
		AmountOut: amountOut,
		Path:      path,
	}
}

// AddLiquidityParams is CreateAddLiquidity's and CallAddLiquidity's argument list.
type AddLiquidityParams struct {
	Router         *types.Address
	Sender         *types.Address
	TokenA         *types.Address
	TokenB         *types.Address
	AmountADesired *big.Int
	AmountBDesired *big.Int
	AmountAMin     *big.Int
	AmountBMin     *big.Int
	To             *types.Address
	Deadline       *big.Int
}

func NewAddLiquidityParams(router, sender, tokenA, tokenB *types.Address, amountADesired, amountBDesired, amountAMin, amountBMin *big.Int, to *types.Address, deadline *big.Int) *AddLiquidityParams {
	return &AddLiquidityParams{
		Router:         router,
		Sender:         sender,
		TokenA:         tokenA,
		TokenB:         tokenB,
		AmountADesired: amountADesired,
		AmountBDesired: amountBDesired,
		AmountAMin:     amountAMin,
		AmountBMin:     amountBMin,
		To:             to,
		Deadline:       deadline,
	}
}

// AddLiquidityETHParams is CreateAddLiquidityETH's and CallAddLiquidityETH's argument list.
type AddLiquidityETHParams struct {
	Router             *types.Address
	Sender             *types.Address
	Token              *types.Address
	AmountTokenDesired *big.Int
	AmountTokenMin     *big.Int
	AmountETHMin       *big.Int
	To                 *types.Address
	Deadline           *big.Int
	Value              *big.Int
}

func NewAddLiquidityETHParams(router, sender, token *types.Address, amountTokenDesired, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline, value *big.Int) *AddLiquidityETHParams {
	return &AddLiquidityETHParams{
		Router:             router,
		Sender:             sender,
		Token:              token,
		AmountTokenDesired: amountTokenDesired,
		AmountTokenMin:     amountTokenMin,
		AmountETHMin:       amountETHMin,
		To:                 to,
		Deadline:           deadline,
		Value:              value,
	}
}

// RemoveLiquidityParams is CreateRemoveLiquidity's and CallRemoveLiquidity's argument list.
type RemoveLiquidityParams struct {
	Router     *types.Address
	Sender     *types.Address
	TokenA     *types.Address
	TokenB     *types.Address
	Liquidity  *big.Int
	AmountAMin *big.Int
	AmountBMin *big.Int
	To         *types.Address
	Deadline   *big.Int
}

func NewRemoveLiquidityParams(router, sender, tokenA, tokenB *types.Address, liquidity, amountAMin, amountBMin *big.Int, to *types.Address, deadline *big.Int) *RemoveLiquidityParams {
	return &RemoveLiquidityParams{
		Router:     router,
		Sender:     sender,
		TokenA:     tokenA,
		TokenB:     tokenB,
		Liquidity:  liquidity,
		AmountAMin: amountAMin,
		AmountBMin: amountBMin,
		To:         to,
		Deadline:   deadline,
	}
}

// RemoveLiquidityETHParams is CreateRemoveLiquidityETH's and CallRemoveLiquidityETH's argument
// list.
type RemoveLiquidityETHParams struct {
	Router         *types.Address
	Sender         *types.Address
	Token          *types.Address
	Liquidity      *big.Int
	AmountTokenMin *big.Int
	AmountETHMin   *big.Int
	To             *types.Address
	Deadline       *big.Int
}

func NewRemoveLiquidityETHParams(router, sender, token *types.Address, liquidity, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline *big.Int) *RemoveLiquidityETHParams {
	return &RemoveLiquidityETHParams{
		Router:         router,
		Sender:         sender,
		Token:          token,
		Liquidity:      liquidity,
		AmountTokenMin: amountTokenMin,
		AmountETHMin:   amountETHMin,
		To:             to,
		Deadline:       deadline,
	}
}

// RemoveLiquidityWithPermitParams is CreateRemoveLiquidityWithPermit's and
// CallRemoveLiquidityWithPermit's argument list.
type RemoveLiquidityWithPermitParams struct {
	Router     *types.Address
	Sender     *types.Address
	TokenA     *types.Address
	TokenB     *types.Address
	Liquidity  *big.Int
	AmountAMin *big.Int
	AmountBMin *big.Int
	To         *types.Address
	Deadline   *big.Int
	ApproveMax bool
	Sig        *types.Signature
}

func NewRemoveLiquidityWithPermitParams(router, sender, tokenA, tokenB *types.Address, liquidity, amountAMin, amountBMin *big.Int, to *types.Address, deadline *big.Int, approveMax bool, sig *types.Signature) *RemoveLiquidityWithPermitParams {
	return &RemoveLiquidityWithPermitParams{
		Router:     router,
		Sender:     sender,
		TokenA:     tokenA,
		TokenB:     tokenB,
		Liquidity:  liquidity,
		AmountAMin: amountAMin,
		AmountBMin: amountBMin,
		To:         to,
		Deadline:   deadline,
		ApproveMax: approveMax,
		Sig:        sig,
	}
}

// RemoveLiquidityETHWithPermitParams is CreateRemoveLiquidityETHWithPermit's and
// CallRemoveLiquidityETHWithPermit's argument list.
type RemoveLiquidityETHWithPermitParams struct {
	Router         *types.Address
	Sender         *types.Address
	Token          *types.Address
	Liquidity      *big.Int
	AmountTokenMin *big.Int
	AmountETHMin   *big.Int
	To             *types.Address
	Deadline       *big.Int
	ApproveMax     bool
	Sig            *types.Signature
}

func NewRemoveLiquidityETHWithPermitParams(router, sender, token *types.Address, liquidity, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline *big.Int, approveMax bool, sig *types.Signature) *RemoveLiquidityETHWithPermitParams {
	return &RemoveLiquidityETHWithPermitParams{
		Router:         router,
		Sender:         sender,
		Token:          token,
		Liquidity:      liquidity,
		AmountTokenMin: amountTokenMin,
		AmountETHMin:   amountETHMin,
		To:             to,
		Deadline:       deadline,
		ApproveMax:     approveMax,
		Sig:            sig,
	}
}

// SwapExactTokensForTokensParams is CreateSwapExactTokensForTokens's and
// CallSwapExactTokensForTokens's argument list.
type SwapExactTokensForTokensParams struct {
	Router       *types.Address
	Sender       *types.Address
	AmountIn     *big.Int
	AmountOutMin *big.Int
	Path         []*types.Address
	To           *types.Address
	Deadline     *big.Int
}

func NewSwapExactTokensForTokensParams(router, sender *types.Address, amountIn, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) *SwapExactTokensForTokensParams {
	return &SwapExactTokensForTokensParams{
		Router:       router,
		Sender:       sender,
		AmountIn:     amountIn,
		AmountOutMin: amountOutMin,
		Path:         path,
		To:           to,
		Deadline:     deadline,
	}
}

// SwapTokensForExactTokensParams is CreateSwapTokensForExactTokens's and
// CallSwapTokensForExactTokens's argument list.
type SwapTokensForExactTokensParams struct {
	Router      *types.Address
	Sender      *types.Address
	AmountOut   *big.Int
	AmountInMax *big.Int
	Path        []*types.Address
	To          *types.Address
	Deadline    *big.Int
}

func NewSwapTokensForExactTokensParams(router, sender *types.Address, amountOut, amountInMax *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) *SwapTokensForExactTokensParams {
	return &SwapTokensForExactTokensParams{
		Router:      router,
		Sender:      sender,
		AmountOut:   amountOut,
		AmountInMax: amountInMax,
		Path:        path,
		To:          to,
		Deadline:    deadline,
	}
}

// SwapExactETHForTokensParams is CreateSwapExactETHForTokens's and CallSwapExactETHForTokens's
// argument list.
type SwapExactETHForTokensParams struct {
	Router       *types.Address
	Sender       *types.Address
	AmountOutMin *big.Int
	Path         []*types.Address
	To           *types.Address
	Deadline     *big.Int
	Value        *big.Int
}

func NewSwapExactETHForTokensParams(router, sender *types.Address, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline, value *big.Int) *SwapExactETHForTokensParams {
	return &SwapExactETHForTokensParams{
		Router:       router,
		Sender:       sender,
		AmountOutMin: amountOutMin,
		Path:         path,
		To:           to,
		Deadline:     deadline,
		Value:        value,
	}
}

// SwapTokensForExactETHParams is CreateSwapTokensForExactETH's and CallSwapTokensForExactETH's
// argument list.
type SwapTokensForExactETHParams struct {
	Router      *types.Address
	Sender      *types.Address
	AmountOut   *big.Int
	AmountInMax *big.Int
	Path        []*types.Address
	To          *types.Address
	Deadline    *big.Int
}

func NewSwapTokensForExactETHParams(router, sender *types.Address, amountOut, amountInMax *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) *SwapTokensForExactETHParams {
	return &SwapTokensForExactETHParams{
		Router:      router,
		Sender:      sender,
		AmountOut:   amountOut,
		AmountInMax: amountInMax,
		Path:        path,
		To:          to,
		Deadline:    deadline,
	}
}

// SwapExactTokensForETHParams is CreateSwapExactTokensForETH's and CallSwapExactTokensForETH's
// argument list.
type SwapExactTokensForETHParams struct {
	Router       *types.Address
	Sender       *types.Address
	AmountIn     *big.Int
	AmountOutMin *big.Int
	Path         []*types.Address
	To           *types.Address
	Deadline     *big.Int
}

func NewSwapExactTokensForETHParams(router, sender *types.Address, amountIn, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) *SwapExactTokensForETHParams {
	return &SwapExactTokensForETHParams{
		Router:       router,
		Sender:       sender,
		AmountIn:     amountIn,
		AmountOutMin: amountOutMin,
		Path:         path,
		To:           to,
		Deadline:     deadline,
	}
}

// SwapETHForExactTokensParams is CreateSwapETHForExactTokens's and CallSwapETHForExactTokens's
// argument list.
type SwapETHForExactTokensParams struct {
	Router    *types.Address
	Sender    *types.Address
	AmountOut *big.Int
	Path      []*types.Address
	To        *types.Address
	Deadline  *big.Int
	Value     *big.Int
}

func NewSwapETHForExactTokensParams(router, sender *types.Address, amountOut *big.Int, path []*types.Address, to *types.Address, deadline, value *big.Int) *SwapETHForExactTokensParams {
	return &SwapETHForExactTokensParams{
		Router:    router,
		Sender:    sender,
		AmountOut: amountOut,
		Path:      path,
		To:        to,
		Deadline:  deadline,
		Value:     value,
	}
}

// RemoveLiquidityETHSupportingFeeOnTransferTokensParams is
// CreateRemoveLiquidityETHSupportingFeeOnTransferTokens's and
// CallRemoveLiquidityETHSupportingFeeOnTransferTokens's argument list.
type RemoveLiquidityETHSupportingFeeOnTransferTokensParams struct {
	Router         *types.Address
	Sender         *types.Address
	Token          *types.Address
	Liquidity      *big.Int
	AmountTokenMin *big.Int
	AmountETHMin   *big.Int
	To             *types.Address
	Deadline       *big.Int
}

func NewRemoveLiquidityETHSupportingFeeOnTransferTokensParams(router, sender, token *types.Address, liquidity, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline *big.Int) *RemoveLiquidityETHSupportingFeeOnTransferTokensParams {
	return &RemoveLiquidityETHSupportingFeeOnTransferTokensParams{
		Router:         router,
		Sender:         sender,
		Token:          token,
		Liquidity:      liquidity,
		AmountTokenMin: amountTokenMin,
		AmountETHMin:   amountETHMin,
		To:             to,
		Deadline:       deadline,
	}
}

// RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensParams is
// CreateRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens's and
// CallRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens's argument list.
type RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensParams struct {
	Router         *types.Address
	Sender         *types.Address
	Token          *types.Address
	Liquidity      *big.Int
	AmountTokenMin *big.Int
	AmountETHMin   *big.Int
	To             *types.Address
	Deadline       *big.Int
	ApproveMax     bool
	Sig            *types.Signature
}

func NewRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensParams(router, sender, token *types.Address, liquidity, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline *big.Int, approveMax bool, sig *types.Signature) *RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensParams {
	return &RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensParams{
		Router:         router,
		Sender:         sender,
		Token:          token,
		Liquidity:      liquidity,
		AmountTokenMin: amountTokenMin,
		AmountETHMin:   amountETHMin,
		To:             to,
		Deadline:       deadline,
		ApproveMax:     approveMax,
		Sig:            sig,
	}
}

// SwapExactTokensForTokensSupportingFeeOnTransferTokensParams is
// CreateSwapExactTokensForTokensSupportingFeeOnTransferTokens's argument list.
type SwapExactTokensForTokensSupportingFeeOnTransferTokensParams struct {
	Router       *types.Address
	Sender       *types.Address
	AmountIn     *big.Int
	AmountOutMin *big.Int
	Path         []*types.Address
	To           *types.Address
	Deadline     *big.Int
}

func NewSwapExactTokensForTokensSupportingFeeOnTransferTokensParams(router, sender *types.Address, amountIn, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) *SwapExactTokensForTokensSupportingFeeOnTransferTokensParams {
	return &SwapExactTokensForTokensSupportingFeeOnTransferTokensParams{
		Router:       router,
		Sender:       sender,
		AmountIn:     amountIn,
		AmountOutMin: amountOutMin,
		Path:         path,
		To:           to,
		Deadline:     deadline,
	}
}

// SwapExactETHForTokensSupportingFeeOnTransferTokensParams is
// CreateSwapExactETHForTokensSupportingFeeOnTransferTokens's argument list.
type SwapExactETHForTokensSupportingFeeOnTransferTokensParams struct {
	Router       *types.Address
	Sender       *types.Address
	AmountOutMin *big.Int
	Path         []*types.Address
	To           *types.Address
	Deadline     *big.Int
	Value        *big.Int
}

func NewSwapExactETHForTokensSupportingFeeOnTransferTokensParams(router, sender *types.Address, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline, value *big.Int) *SwapExactETHForTokensSupportingFeeOnTransferTokensParams {
	return &SwapExactETHForTokensSupportingFeeOnTransferTokensParams{
		Router:       router,
		Sender:       sender,
		AmountOutMin: amountOutMin,
		Path:         path,
		To:           to,
		Deadline:     deadline,
		Value:        value,
	}
}

// SwapExactTokensForETHSupportingFeeOnTransferTokensParams is
// CreateSwapExactTokensForETHSupportingFeeOnTransferTokens's argument list.
type SwapExactTokensForETHSupportingFeeOnTransferTokensParams struct {
	Router       *types.Address
	Sender       *types.Address
	AmountIn     *big.Int
	AmountOutMin *big.Int
	Path         []*types.Address
	To           *types.Address
	Deadline     *big.Int
}

func NewSwapExactTokensForETHSupportingFeeOnTransferTokensParams(router, sender *types.Address, amountIn, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) *SwapExactTokensForETHSupportingFeeOnTransferTokensParams {
	return &SwapExactTokensForETHSupportingFeeOnTransferTokensParams{
		Router:       router,
		Sender:       sender,
		AmountIn:     amountIn,
		AmountOutMin: amountOutMin,
		Path:         path,
		To:           to,
		Deadline:     deadline,
	}
}

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
