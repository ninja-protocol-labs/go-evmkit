package v2

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// WETHRequest is WETH()'s input on Contract.
type WETHRequest struct {
	Contract *types.Address
}

func NewWETHRequest(contract *types.Address) *WETHRequest {
	return &WETHRequest{
		Contract: contract,
	}
}

func (r *WETHRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// QuoteRequest is quote(amountA, reserveA, reserveB)'s input on Contract.
type QuoteRequest struct {
	Contract *types.Address
	AmountA  *big.Int
	ReserveA *big.Int
	ReserveB *big.Int
}

func NewQuoteRequest(contract *types.Address, amountA, reserveA, reserveB *big.Int) *QuoteRequest {
	return &QuoteRequest{
		Contract: contract,
		AmountA:  amountA,
		ReserveA: reserveA,
		ReserveB: reserveB,
	}
}

func (r *QuoteRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.AmountA == nil {
		return fmt.Errorf("%w: amountA is nil", bind.ErrInvalidRequest)
	}

	if r.ReserveA == nil {
		return fmt.Errorf("%w: reserveA is nil", bind.ErrInvalidRequest)
	}

	if r.ReserveB == nil {
		return fmt.Errorf("%w: reserveB is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetAmountOutRequest is getAmountOut(amountIn, reserveIn, reserveOut)'s input on Contract.
type GetAmountOutRequest struct {
	Contract   *types.Address
	AmountIn   *big.Int
	ReserveIn  *big.Int
	ReserveOut *big.Int
}

func NewGetAmountOutRequest(contract *types.Address, amountIn, reserveIn, reserveOut *big.Int) *GetAmountOutRequest {
	return &GetAmountOutRequest{
		Contract:   contract,
		AmountIn:   amountIn,
		ReserveIn:  reserveIn,
		ReserveOut: reserveOut,
	}
}

func (r *GetAmountOutRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.AmountIn == nil {
		return fmt.Errorf("%w: amountIn is nil", bind.ErrInvalidRequest)
	}

	if r.ReserveIn == nil {
		return fmt.Errorf("%w: reserveIn is nil", bind.ErrInvalidRequest)
	}

	if r.ReserveOut == nil {
		return fmt.Errorf("%w: reserveOut is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetAmountInRequest is getAmountIn(amountOut, reserveIn, reserveOut)'s input on Contract.
type GetAmountInRequest struct {
	Contract   *types.Address
	AmountOut  *big.Int
	ReserveIn  *big.Int
	ReserveOut *big.Int
}

func NewGetAmountInRequest(contract *types.Address, amountOut, reserveIn, reserveOut *big.Int) *GetAmountInRequest {
	return &GetAmountInRequest{
		Contract:   contract,
		AmountOut:  amountOut,
		ReserveIn:  reserveIn,
		ReserveOut: reserveOut,
	}
}

func (r *GetAmountInRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.AmountOut == nil {
		return fmt.Errorf("%w: amountOut is nil", bind.ErrInvalidRequest)
	}

	if r.ReserveIn == nil {
		return fmt.Errorf("%w: reserveIn is nil", bind.ErrInvalidRequest)
	}

	if r.ReserveOut == nil {
		return fmt.Errorf("%w: reserveOut is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetAmountsOutRequest is getAmountsOut(amountIn, path)'s input on Contract.
type GetAmountsOutRequest struct {
	Contract *types.Address
	AmountIn *big.Int
	Path     []*types.Address
}

func NewGetAmountsOutRequest(contract *types.Address, amountIn *big.Int, path []*types.Address) *GetAmountsOutRequest {
	return &GetAmountsOutRequest{
		Contract: contract,
		AmountIn: amountIn,
		Path:     path,
	}
}

func (r *GetAmountsOutRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.AmountIn == nil {
		return fmt.Errorf("%w: amountIn is nil", bind.ErrInvalidRequest)
	}

	if len(r.Path) == 0 {
		return fmt.Errorf("%w: path is empty", bind.ErrInvalidRequest)
	}

	for i, a := range r.Path {
		if a == nil {
			return fmt.Errorf("%w: path[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

// GetAmountsInRequest is getAmountsIn(amountOut, path)'s input on Contract.
type GetAmountsInRequest struct {
	Contract  *types.Address
	AmountOut *big.Int
	Path      []*types.Address
}

func NewGetAmountsInRequest(contract *types.Address, amountOut *big.Int, path []*types.Address) *GetAmountsInRequest {
	return &GetAmountsInRequest{
		Contract:  contract,
		AmountOut: amountOut,
		Path:      path,
	}
}

func (r *GetAmountsInRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.AmountOut == nil {
		return fmt.Errorf("%w: amountOut is nil", bind.ErrInvalidRequest)
	}

	if len(r.Path) == 0 {
		return fmt.Errorf("%w: path is empty", bind.ErrInvalidRequest)
	}

	for i, a := range r.Path {
		if a == nil {
			return fmt.Errorf("%w: path[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

// AddLiquidityRequest is addLiquidity(tokenA, tokenB, amountADesired, amountBDesired, amountAMin, amountBMin, to, deadline)'s input on Contract.
type AddLiquidityRequest struct {
	Contract       *types.Address
	TokenA         *types.Address
	TokenB         *types.Address
	AmountADesired *big.Int
	AmountBDesired *big.Int
	AmountAMin     *big.Int
	AmountBMin     *big.Int
	To             *types.Address
	Deadline       *big.Int
}

func NewAddLiquidityRequest(contract, tokenA, tokenB *types.Address, amountADesired, amountBDesired, amountAMin, amountBMin *big.Int, to *types.Address, deadline *big.Int) *AddLiquidityRequest {
	return &AddLiquidityRequest{
		Contract:       contract,
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

func (r *AddLiquidityRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.TokenA == nil {
		return fmt.Errorf("%w: tokenA is nil", bind.ErrInvalidRequest)
	}

	if r.TokenB == nil {
		return fmt.Errorf("%w: tokenB is nil", bind.ErrInvalidRequest)
	}

	if r.AmountADesired == nil {
		return fmt.Errorf("%w: amountADesired is nil", bind.ErrInvalidRequest)
	}

	if r.AmountBDesired == nil {
		return fmt.Errorf("%w: amountBDesired is nil", bind.ErrInvalidRequest)
	}

	if r.AmountAMin == nil {
		return fmt.Errorf("%w: amountAMin is nil", bind.ErrInvalidRequest)
	}

	if r.AmountBMin == nil {
		return fmt.Errorf("%w: amountBMin is nil", bind.ErrInvalidRequest)
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	if r.Deadline == nil {
		return fmt.Errorf("%w: deadline is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// AddLiquidityETHRequest is addLiquidityETH(token, amountTokenDesired, amountTokenMin, amountETHMin, to, deadline)'s input on Contract; AmountETHDesired is sent as msg.value.
type AddLiquidityETHRequest struct {
	Contract           *types.Address
	Token              *types.Address
	AmountTokenDesired *big.Int
	AmountTokenMin     *big.Int
	AmountETHMin       *big.Int
	To                 *types.Address
	Deadline           *big.Int
	AmountETHDesired   *big.Int
}

func NewAddLiquidityETHRequest(contract, token *types.Address, amountTokenDesired, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline, amountETHDesired *big.Int) *AddLiquidityETHRequest {
	return &AddLiquidityETHRequest{
		Contract:           contract,
		Token:              token,
		AmountTokenDesired: amountTokenDesired,
		AmountTokenMin:     amountTokenMin,
		AmountETHMin:       amountETHMin,
		To:                 to,
		Deadline:           deadline,
		AmountETHDesired:   amountETHDesired,
	}
}

func (r *AddLiquidityETHRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Token == nil {
		return fmt.Errorf("%w: token is nil", bind.ErrInvalidRequest)
	}

	if r.AmountTokenDesired == nil {
		return fmt.Errorf("%w: amountTokenDesired is nil", bind.ErrInvalidRequest)
	}

	if r.AmountTokenMin == nil {
		return fmt.Errorf("%w: amountTokenMin is nil", bind.ErrInvalidRequest)
	}

	if r.AmountETHMin == nil {
		return fmt.Errorf("%w: amountETHMin is nil", bind.ErrInvalidRequest)
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	if r.Deadline == nil {
		return fmt.Errorf("%w: deadline is nil", bind.ErrInvalidRequest)
	}

	if r.AmountETHDesired == nil {
		return fmt.Errorf("%w: amountETHDesired is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// RemoveLiquidityRequest is removeLiquidity(tokenA, tokenB, liquidity, amountAMin, amountBMin, to, deadline)'s input on Contract.
type RemoveLiquidityRequest struct {
	Contract   *types.Address
	TokenA     *types.Address
	TokenB     *types.Address
	Liquidity  *big.Int
	AmountAMin *big.Int
	AmountBMin *big.Int
	To         *types.Address
	Deadline   *big.Int
}

func NewRemoveLiquidityRequest(contract, tokenA, tokenB *types.Address, liquidity, amountAMin, amountBMin *big.Int, to *types.Address, deadline *big.Int) *RemoveLiquidityRequest {
	return &RemoveLiquidityRequest{
		Contract:   contract,
		TokenA:     tokenA,
		TokenB:     tokenB,
		Liquidity:  liquidity,
		AmountAMin: amountAMin,
		AmountBMin: amountBMin,
		To:         to,
		Deadline:   deadline,
	}
}

func (r *RemoveLiquidityRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.TokenA == nil {
		return fmt.Errorf("%w: tokenA is nil", bind.ErrInvalidRequest)
	}

	if r.TokenB == nil {
		return fmt.Errorf("%w: tokenB is nil", bind.ErrInvalidRequest)
	}

	if r.Liquidity == nil {
		return fmt.Errorf("%w: liquidity is nil", bind.ErrInvalidRequest)
	}

	if r.AmountAMin == nil {
		return fmt.Errorf("%w: amountAMin is nil", bind.ErrInvalidRequest)
	}

	if r.AmountBMin == nil {
		return fmt.Errorf("%w: amountBMin is nil", bind.ErrInvalidRequest)
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	if r.Deadline == nil {
		return fmt.Errorf("%w: deadline is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// RemoveLiquidityETHRequest is removeLiquidityETH(token, liquidity, amountTokenMin, amountETHMin, to, deadline)'s input on Contract.
type RemoveLiquidityETHRequest struct {
	Contract       *types.Address
	Token          *types.Address
	Liquidity      *big.Int
	AmountTokenMin *big.Int
	AmountETHMin   *big.Int
	To             *types.Address
	Deadline       *big.Int
}

func NewRemoveLiquidityETHRequest(contract, token *types.Address, liquidity, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline *big.Int) *RemoveLiquidityETHRequest {
	return &RemoveLiquidityETHRequest{
		Contract:       contract,
		Token:          token,
		Liquidity:      liquidity,
		AmountTokenMin: amountTokenMin,
		AmountETHMin:   amountETHMin,
		To:             to,
		Deadline:       deadline,
	}
}

func (r *RemoveLiquidityETHRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Token == nil {
		return fmt.Errorf("%w: token is nil", bind.ErrInvalidRequest)
	}

	if r.Liquidity == nil {
		return fmt.Errorf("%w: liquidity is nil", bind.ErrInvalidRequest)
	}

	if r.AmountTokenMin == nil {
		return fmt.Errorf("%w: amountTokenMin is nil", bind.ErrInvalidRequest)
	}

	if r.AmountETHMin == nil {
		return fmt.Errorf("%w: amountETHMin is nil", bind.ErrInvalidRequest)
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	if r.Deadline == nil {
		return fmt.Errorf("%w: deadline is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// RemoveLiquidityWithPermitRequest is removeLiquidityWithPermit(tokenA, tokenB, liquidity, amountAMin, amountBMin, to, deadline, approveMax, v, r, s)'s input on Contract.
type RemoveLiquidityWithPermitRequest struct {
	Contract   *types.Address
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

func NewRemoveLiquidityWithPermitRequest(contract, tokenA, tokenB *types.Address, liquidity, amountAMin, amountBMin *big.Int, to *types.Address, deadline *big.Int, approveMax bool, sig *types.Signature) *RemoveLiquidityWithPermitRequest {
	return &RemoveLiquidityWithPermitRequest{
		Contract:   contract,
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

func (r *RemoveLiquidityWithPermitRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.TokenA == nil {
		return fmt.Errorf("%w: tokenA is nil", bind.ErrInvalidRequest)
	}

	if r.TokenB == nil {
		return fmt.Errorf("%w: tokenB is nil", bind.ErrInvalidRequest)
	}

	if r.Liquidity == nil {
		return fmt.Errorf("%w: liquidity is nil", bind.ErrInvalidRequest)
	}

	if r.AmountAMin == nil {
		return fmt.Errorf("%w: amountAMin is nil", bind.ErrInvalidRequest)
	}

	if r.AmountBMin == nil {
		return fmt.Errorf("%w: amountBMin is nil", bind.ErrInvalidRequest)
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	if r.Deadline == nil {
		return fmt.Errorf("%w: deadline is nil", bind.ErrInvalidRequest)
	}

	if r.Sig == nil {
		return fmt.Errorf("%w: sig is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// RemoveLiquidityETHWithPermitRequest is removeLiquidityETHWithPermit(token, liquidity, amountTokenMin, amountETHMin, to, deadline, approveMax, v, r, s)'s input on Contract.
type RemoveLiquidityETHWithPermitRequest struct {
	Contract       *types.Address
	Token          *types.Address
	Liquidity      *big.Int
	AmountTokenMin *big.Int
	AmountETHMin   *big.Int
	To             *types.Address
	Deadline       *big.Int
	ApproveMax     bool
	Sig            *types.Signature
}

func NewRemoveLiquidityETHWithPermitRequest(contract, token *types.Address, liquidity, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline *big.Int, approveMax bool, sig *types.Signature) *RemoveLiquidityETHWithPermitRequest {
	return &RemoveLiquidityETHWithPermitRequest{
		Contract:       contract,
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

func (r *RemoveLiquidityETHWithPermitRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Token == nil {
		return fmt.Errorf("%w: token is nil", bind.ErrInvalidRequest)
	}

	if r.Liquidity == nil {
		return fmt.Errorf("%w: liquidity is nil", bind.ErrInvalidRequest)
	}

	if r.AmountTokenMin == nil {
		return fmt.Errorf("%w: amountTokenMin is nil", bind.ErrInvalidRequest)
	}

	if r.AmountETHMin == nil {
		return fmt.Errorf("%w: amountETHMin is nil", bind.ErrInvalidRequest)
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	if r.Deadline == nil {
		return fmt.Errorf("%w: deadline is nil", bind.ErrInvalidRequest)
	}

	if r.Sig == nil {
		return fmt.Errorf("%w: sig is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SwapExactTokensForTokensRequest is swapExactTokensForTokens(amountIn, amountOutMin, path, to, deadline)'s input on Contract.
type SwapExactTokensForTokensRequest struct {
	Contract     *types.Address
	AmountIn     *big.Int
	AmountOutMin *big.Int
	Path         []*types.Address
	To           *types.Address
	Deadline     *big.Int
}

func NewSwapExactTokensForTokensRequest(contract *types.Address, amountIn, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) *SwapExactTokensForTokensRequest {
	return &SwapExactTokensForTokensRequest{
		Contract:     contract,
		AmountIn:     amountIn,
		AmountOutMin: amountOutMin,
		Path:         path,
		To:           to,
		Deadline:     deadline,
	}
}

func (r *SwapExactTokensForTokensRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.AmountIn == nil {
		return fmt.Errorf("%w: amountIn is nil", bind.ErrInvalidRequest)
	}

	if r.AmountOutMin == nil {
		return fmt.Errorf("%w: amountOutMin is nil", bind.ErrInvalidRequest)
	}

	if len(r.Path) == 0 {
		return fmt.Errorf("%w: path is empty", bind.ErrInvalidRequest)
	}

	for i, a := range r.Path {
		if a == nil {
			return fmt.Errorf("%w: path[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	if r.Deadline == nil {
		return fmt.Errorf("%w: deadline is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SwapTokensForExactTokensRequest is swapTokensForExactTokens(amountOut, amountInMax, path, to, deadline)'s input on Contract.
type SwapTokensForExactTokensRequest struct {
	Contract    *types.Address
	AmountOut   *big.Int
	AmountInMax *big.Int
	Path        []*types.Address
	To          *types.Address
	Deadline    *big.Int
}

func NewSwapTokensForExactTokensRequest(contract *types.Address, amountOut, amountInMax *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) *SwapTokensForExactTokensRequest {
	return &SwapTokensForExactTokensRequest{
		Contract:    contract,
		AmountOut:   amountOut,
		AmountInMax: amountInMax,
		Path:        path,
		To:          to,
		Deadline:    deadline,
	}
}

func (r *SwapTokensForExactTokensRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.AmountOut == nil {
		return fmt.Errorf("%w: amountOut is nil", bind.ErrInvalidRequest)
	}

	if r.AmountInMax == nil {
		return fmt.Errorf("%w: amountInMax is nil", bind.ErrInvalidRequest)
	}

	if len(r.Path) == 0 {
		return fmt.Errorf("%w: path is empty", bind.ErrInvalidRequest)
	}

	for i, a := range r.Path {
		if a == nil {
			return fmt.Errorf("%w: path[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	if r.Deadline == nil {
		return fmt.Errorf("%w: deadline is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SwapExactETHForTokensRequest is swapExactETHForTokens(amountOutMin, path, to, deadline)'s input on Contract; AmountIn is sent as msg.value.
type SwapExactETHForTokensRequest struct {
	Contract     *types.Address
	AmountOutMin *big.Int
	Path         []*types.Address
	To           *types.Address
	Deadline     *big.Int
	AmountIn     *big.Int
}

func NewSwapExactETHForTokensRequest(contract *types.Address, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline, amountIn *big.Int) *SwapExactETHForTokensRequest {
	return &SwapExactETHForTokensRequest{
		Contract:     contract,
		AmountOutMin: amountOutMin,
		Path:         path,
		To:           to,
		Deadline:     deadline,
		AmountIn:     amountIn,
	}
}

func (r *SwapExactETHForTokensRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.AmountOutMin == nil {
		return fmt.Errorf("%w: amountOutMin is nil", bind.ErrInvalidRequest)
	}

	if len(r.Path) == 0 {
		return fmt.Errorf("%w: path is empty", bind.ErrInvalidRequest)
	}

	for i, a := range r.Path {
		if a == nil {
			return fmt.Errorf("%w: path[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	if r.Deadline == nil {
		return fmt.Errorf("%w: deadline is nil", bind.ErrInvalidRequest)
	}

	if r.AmountIn == nil {
		return fmt.Errorf("%w: amountIn is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SwapTokensForExactETHRequest is swapTokensForExactETH(amountOut, amountInMax, path, to, deadline)'s input on Contract.
type SwapTokensForExactETHRequest struct {
	Contract    *types.Address
	AmountOut   *big.Int
	AmountInMax *big.Int
	Path        []*types.Address
	To          *types.Address
	Deadline    *big.Int
}

func NewSwapTokensForExactETHRequest(contract *types.Address, amountOut, amountInMax *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) *SwapTokensForExactETHRequest {
	return &SwapTokensForExactETHRequest{
		Contract:    contract,
		AmountOut:   amountOut,
		AmountInMax: amountInMax,
		Path:        path,
		To:          to,
		Deadline:    deadline,
	}
}

func (r *SwapTokensForExactETHRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.AmountOut == nil {
		return fmt.Errorf("%w: amountOut is nil", bind.ErrInvalidRequest)
	}

	if r.AmountInMax == nil {
		return fmt.Errorf("%w: amountInMax is nil", bind.ErrInvalidRequest)
	}

	if len(r.Path) == 0 {
		return fmt.Errorf("%w: path is empty", bind.ErrInvalidRequest)
	}

	for i, a := range r.Path {
		if a == nil {
			return fmt.Errorf("%w: path[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	if r.Deadline == nil {
		return fmt.Errorf("%w: deadline is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SwapExactTokensForETHRequest is swapExactTokensForETH(amountIn, amountOutMin, path, to, deadline)'s input on Contract.
type SwapExactTokensForETHRequest struct {
	Contract     *types.Address
	AmountIn     *big.Int
	AmountOutMin *big.Int
	Path         []*types.Address
	To           *types.Address
	Deadline     *big.Int
}

func NewSwapExactTokensForETHRequest(contract *types.Address, amountIn, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) *SwapExactTokensForETHRequest {
	return &SwapExactTokensForETHRequest{
		Contract:     contract,
		AmountIn:     amountIn,
		AmountOutMin: amountOutMin,
		Path:         path,
		To:           to,
		Deadline:     deadline,
	}
}

func (r *SwapExactTokensForETHRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.AmountIn == nil {
		return fmt.Errorf("%w: amountIn is nil", bind.ErrInvalidRequest)
	}

	if r.AmountOutMin == nil {
		return fmt.Errorf("%w: amountOutMin is nil", bind.ErrInvalidRequest)
	}

	if len(r.Path) == 0 {
		return fmt.Errorf("%w: path is empty", bind.ErrInvalidRequest)
	}

	for i, a := range r.Path {
		if a == nil {
			return fmt.Errorf("%w: path[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	if r.Deadline == nil {
		return fmt.Errorf("%w: deadline is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SwapETHForExactTokensRequest is swapETHForExactTokens(amountOut, path, to, deadline)'s input on Contract; AmountInMax is sent as msg.value.
type SwapETHForExactTokensRequest struct {
	Contract    *types.Address
	AmountOut   *big.Int
	Path        []*types.Address
	To          *types.Address
	Deadline    *big.Int
	AmountInMax *big.Int
}

func NewSwapETHForExactTokensRequest(contract *types.Address, amountOut *big.Int, path []*types.Address, to *types.Address, deadline, amountInMax *big.Int) *SwapETHForExactTokensRequest {
	return &SwapETHForExactTokensRequest{
		Contract:    contract,
		AmountOut:   amountOut,
		Path:        path,
		To:          to,
		Deadline:    deadline,
		AmountInMax: amountInMax,
	}
}

func (r *SwapETHForExactTokensRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.AmountOut == nil {
		return fmt.Errorf("%w: amountOut is nil", bind.ErrInvalidRequest)
	}

	if len(r.Path) == 0 {
		return fmt.Errorf("%w: path is empty", bind.ErrInvalidRequest)
	}

	for i, a := range r.Path {
		if a == nil {
			return fmt.Errorf("%w: path[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	if r.Deadline == nil {
		return fmt.Errorf("%w: deadline is nil", bind.ErrInvalidRequest)
	}

	if r.AmountInMax == nil {
		return fmt.Errorf("%w: amountInMax is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// RemoveLiquidityETHSupportingFeeOnTransferTokensRequest is removeLiquidityETHSupportingFeeOnTransferTokens(token, liquidity, amountTokenMin, amountETHMin, to, deadline)'s input on Contract.
type RemoveLiquidityETHSupportingFeeOnTransferTokensRequest struct {
	Contract       *types.Address
	Token          *types.Address
	Liquidity      *big.Int
	AmountTokenMin *big.Int
	AmountETHMin   *big.Int
	To             *types.Address
	Deadline       *big.Int
}

func NewRemoveLiquidityETHSupportingFeeOnTransferTokensRequest(contract, token *types.Address, liquidity, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline *big.Int) *RemoveLiquidityETHSupportingFeeOnTransferTokensRequest {
	return &RemoveLiquidityETHSupportingFeeOnTransferTokensRequest{
		Contract:       contract,
		Token:          token,
		Liquidity:      liquidity,
		AmountTokenMin: amountTokenMin,
		AmountETHMin:   amountETHMin,
		To:             to,
		Deadline:       deadline,
	}
}

func (r *RemoveLiquidityETHSupportingFeeOnTransferTokensRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Token == nil {
		return fmt.Errorf("%w: token is nil", bind.ErrInvalidRequest)
	}

	if r.Liquidity == nil {
		return fmt.Errorf("%w: liquidity is nil", bind.ErrInvalidRequest)
	}

	if r.AmountTokenMin == nil {
		return fmt.Errorf("%w: amountTokenMin is nil", bind.ErrInvalidRequest)
	}

	if r.AmountETHMin == nil {
		return fmt.Errorf("%w: amountETHMin is nil", bind.ErrInvalidRequest)
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	if r.Deadline == nil {
		return fmt.Errorf("%w: deadline is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest is removeLiquidityETHWithPermitSupportingFeeOnTransferTokens(token, liquidity, amountTokenMin, amountETHMin, to, deadline, approveMax, v, r, s)'s input on Contract.
type RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest struct {
	Contract       *types.Address
	Token          *types.Address
	Liquidity      *big.Int
	AmountTokenMin *big.Int
	AmountETHMin   *big.Int
	To             *types.Address
	Deadline       *big.Int
	ApproveMax     bool
	Sig            *types.Signature
}

func NewRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest(contract, token *types.Address, liquidity, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline *big.Int, approveMax bool, sig *types.Signature) *RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest {
	return &RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest{
		Contract:       contract,
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

func (r *RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokensRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Token == nil {
		return fmt.Errorf("%w: token is nil", bind.ErrInvalidRequest)
	}

	if r.Liquidity == nil {
		return fmt.Errorf("%w: liquidity is nil", bind.ErrInvalidRequest)
	}

	if r.AmountTokenMin == nil {
		return fmt.Errorf("%w: amountTokenMin is nil", bind.ErrInvalidRequest)
	}

	if r.AmountETHMin == nil {
		return fmt.Errorf("%w: amountETHMin is nil", bind.ErrInvalidRequest)
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	if r.Deadline == nil {
		return fmt.Errorf("%w: deadline is nil", bind.ErrInvalidRequest)
	}

	if r.Sig == nil {
		return fmt.Errorf("%w: sig is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SwapExactTokensForTokensSupportingFeeOnTransferTokensRequest is swapExactTokensForTokensSupportingFeeOnTransferTokens(amountIn, amountOutMin, path, to, deadline)'s input on Contract.
type SwapExactTokensForTokensSupportingFeeOnTransferTokensRequest struct {
	Contract     *types.Address
	AmountIn     *big.Int
	AmountOutMin *big.Int
	Path         []*types.Address
	To           *types.Address
	Deadline     *big.Int
}

func NewSwapExactTokensForTokensSupportingFeeOnTransferTokensRequest(contract *types.Address, amountIn, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) *SwapExactTokensForTokensSupportingFeeOnTransferTokensRequest {
	return &SwapExactTokensForTokensSupportingFeeOnTransferTokensRequest{
		Contract:     contract,
		AmountIn:     amountIn,
		AmountOutMin: amountOutMin,
		Path:         path,
		To:           to,
		Deadline:     deadline,
	}
}

func (r *SwapExactTokensForTokensSupportingFeeOnTransferTokensRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.AmountIn == nil {
		return fmt.Errorf("%w: amountIn is nil", bind.ErrInvalidRequest)
	}

	if r.AmountOutMin == nil {
		return fmt.Errorf("%w: amountOutMin is nil", bind.ErrInvalidRequest)
	}

	if len(r.Path) == 0 {
		return fmt.Errorf("%w: path is empty", bind.ErrInvalidRequest)
	}

	for i, a := range r.Path {
		if a == nil {
			return fmt.Errorf("%w: path[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	if r.Deadline == nil {
		return fmt.Errorf("%w: deadline is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SwapExactETHForTokensSupportingFeeOnTransferTokensRequest is swapExactETHForTokensSupportingFeeOnTransferTokens(amountOutMin, path, to, deadline)'s input on Contract; AmountIn is sent as msg.value.
type SwapExactETHForTokensSupportingFeeOnTransferTokensRequest struct {
	Contract     *types.Address
	AmountOutMin *big.Int
	Path         []*types.Address
	To           *types.Address
	Deadline     *big.Int
	AmountIn     *big.Int
}

func NewSwapExactETHForTokensSupportingFeeOnTransferTokensRequest(contract *types.Address, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline, amountIn *big.Int) *SwapExactETHForTokensSupportingFeeOnTransferTokensRequest {
	return &SwapExactETHForTokensSupportingFeeOnTransferTokensRequest{
		Contract:     contract,
		AmountOutMin: amountOutMin,
		Path:         path,
		To:           to,
		Deadline:     deadline,
		AmountIn:     amountIn,
	}
}

func (r *SwapExactETHForTokensSupportingFeeOnTransferTokensRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.AmountOutMin == nil {
		return fmt.Errorf("%w: amountOutMin is nil", bind.ErrInvalidRequest)
	}

	if len(r.Path) == 0 {
		return fmt.Errorf("%w: path is empty", bind.ErrInvalidRequest)
	}

	for i, a := range r.Path {
		if a == nil {
			return fmt.Errorf("%w: path[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	if r.Deadline == nil {
		return fmt.Errorf("%w: deadline is nil", bind.ErrInvalidRequest)
	}

	if r.AmountIn == nil {
		return fmt.Errorf("%w: amountIn is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SwapExactTokensForETHSupportingFeeOnTransferTokensRequest is swapExactTokensForETHSupportingFeeOnTransferTokens(amountIn, amountOutMin, path, to, deadline)'s input on Contract.
type SwapExactTokensForETHSupportingFeeOnTransferTokensRequest struct {
	Contract     *types.Address
	AmountIn     *big.Int
	AmountOutMin *big.Int
	Path         []*types.Address
	To           *types.Address
	Deadline     *big.Int
}

func NewSwapExactTokensForETHSupportingFeeOnTransferTokensRequest(contract *types.Address, amountIn, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) *SwapExactTokensForETHSupportingFeeOnTransferTokensRequest {
	return &SwapExactTokensForETHSupportingFeeOnTransferTokensRequest{
		Contract:     contract,
		AmountIn:     amountIn,
		AmountOutMin: amountOutMin,
		Path:         path,
		To:           to,
		Deadline:     deadline,
	}
}

func (r *SwapExactTokensForETHSupportingFeeOnTransferTokensRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.AmountIn == nil {
		return fmt.Errorf("%w: amountIn is nil", bind.ErrInvalidRequest)
	}

	if r.AmountOutMin == nil {
		return fmt.Errorf("%w: amountOutMin is nil", bind.ErrInvalidRequest)
	}

	if len(r.Path) == 0 {
		return fmt.Errorf("%w: path is empty", bind.ErrInvalidRequest)
	}

	for i, a := range r.Path {
		if a == nil {
			return fmt.Errorf("%w: path[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	if r.Deadline == nil {
		return fmt.Errorf("%w: deadline is nil", bind.ErrInvalidRequest)
	}

	return nil
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
