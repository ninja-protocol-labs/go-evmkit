package v3

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// ExactInputSingleRequest is exactInputSingle(params)'s input on Contract; NativeAmount, if set, is sent as msg.value.
type ExactInputSingleRequest struct {
	Contract     *types.Address
	Params       *ExactInputSingleParams
	NativeAmount *big.Int
}

func NewExactInputSingleRequest(contract *types.Address, params *ExactInputSingleParams, nativeAmount *big.Int) *ExactInputSingleRequest {
	return &ExactInputSingleRequest{
		Contract:     contract,
		Params:       params,
		NativeAmount: nativeAmount,
	}
}

func (r *ExactInputSingleRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if err := validateExactInputSingleParams(r.Params); err != nil {
		return err
	}

	return nil
}

// ExactInputRequest is exactInput(params)'s input on Contract; NativeAmount, if set, is sent as msg.value.
type ExactInputRequest struct {
	Contract     *types.Address
	Params       *ExactInputParams
	NativeAmount *big.Int
}

func NewExactInputRequest(contract *types.Address, params *ExactInputParams, nativeAmount *big.Int) *ExactInputRequest {
	return &ExactInputRequest{
		Contract:     contract,
		Params:       params,
		NativeAmount: nativeAmount,
	}
}

func (r *ExactInputRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if err := validateExactInputParams(r.Params); err != nil {
		return err
	}

	return nil
}

// ExactOutputSingleRequest is exactOutputSingle(params)'s input on Contract; NativeAmount, if set, is sent as msg.value.
type ExactOutputSingleRequest struct {
	Contract     *types.Address
	Params       *ExactOutputSingleParams
	NativeAmount *big.Int
}

func NewExactOutputSingleRequest(contract *types.Address, params *ExactOutputSingleParams, nativeAmount *big.Int) *ExactOutputSingleRequest {
	return &ExactOutputSingleRequest{
		Contract:     contract,
		Params:       params,
		NativeAmount: nativeAmount,
	}
}

func (r *ExactOutputSingleRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if err := validateExactOutputSingleParams(r.Params); err != nil {
		return err
	}

	return nil
}

// ExactOutputRequest is exactOutput(params)'s input on Contract; NativeAmount, if set, is sent as msg.value.
type ExactOutputRequest struct {
	Contract     *types.Address
	Params       *ExactOutputParams
	NativeAmount *big.Int
}

func NewExactOutputRequest(contract *types.Address, params *ExactOutputParams, nativeAmount *big.Int) *ExactOutputRequest {
	return &ExactOutputRequest{
		Contract:     contract,
		Params:       params,
		NativeAmount: nativeAmount,
	}
}

func (r *ExactOutputRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if err := validateExactOutputParams(r.Params); err != nil {
		return err
	}

	return nil
}

// RouterUnwrapWETH9Request is unwrapWETH9(amountMinimum, recipient)'s input on Contract.
type RouterUnwrapWETH9Request struct {
	Contract      *types.Address
	AmountMinimum *big.Int
	Recipient     *types.Address
}

func NewRouterUnwrapWETH9Request(contract *types.Address, amountMinimum *big.Int, recipient *types.Address) *RouterUnwrapWETH9Request {
	return &RouterUnwrapWETH9Request{
		Contract:      contract,
		AmountMinimum: amountMinimum,
		Recipient:     recipient,
	}
}

func (r *RouterUnwrapWETH9Request) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.AmountMinimum == nil {
		return fmt.Errorf("%w: amountMinimum is nil", bind.ErrInvalidRequest)
	}

	if r.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// RouterUnwrapWETH9WithFeeRequest is unwrapWETH9WithFee(amountMinimum, recipient, feeBips, feeRecipient)'s input on Contract.
type RouterUnwrapWETH9WithFeeRequest struct {
	Contract      *types.Address
	AmountMinimum *big.Int
	Recipient     *types.Address
	FeeBips       *big.Int
	FeeRecipient  *types.Address
}

func NewRouterUnwrapWETH9WithFeeRequest(contract *types.Address, amountMinimum *big.Int, recipient *types.Address, feeBips *big.Int, feeRecipient *types.Address) *RouterUnwrapWETH9WithFeeRequest {
	return &RouterUnwrapWETH9WithFeeRequest{
		Contract:      contract,
		AmountMinimum: amountMinimum,
		Recipient:     recipient,
		FeeBips:       feeBips,
		FeeRecipient:  feeRecipient,
	}
}

func (r *RouterUnwrapWETH9WithFeeRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.AmountMinimum == nil {
		return fmt.Errorf("%w: amountMinimum is nil", bind.ErrInvalidRequest)
	}

	if r.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	if r.FeeBips == nil {
		return fmt.Errorf("%w: feeBips is nil", bind.ErrInvalidRequest)
	}

	if r.FeeRecipient == nil {
		return fmt.Errorf("%w: feeRecipient is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// RouterSweepTokenRequest is sweepToken(token, amountMinimum, recipient)'s input on Contract.
type RouterSweepTokenRequest struct {
	Contract      *types.Address
	Token         *types.Address
	AmountMinimum *big.Int
	Recipient     *types.Address
}

func NewRouterSweepTokenRequest(contract, token *types.Address, amountMinimum *big.Int, recipient *types.Address) *RouterSweepTokenRequest {
	return &RouterSweepTokenRequest{
		Contract:      contract,
		Token:         token,
		AmountMinimum: amountMinimum,
		Recipient:     recipient,
	}
}

func (r *RouterSweepTokenRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Token == nil {
		return fmt.Errorf("%w: token is nil", bind.ErrInvalidRequest)
	}

	if r.AmountMinimum == nil {
		return fmt.Errorf("%w: amountMinimum is nil", bind.ErrInvalidRequest)
	}

	if r.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// RouterSweepTokenWithFeeRequest is sweepTokenWithFee(token, amountMinimum, recipient, feeBips, feeRecipient)'s input on Contract.
type RouterSweepTokenWithFeeRequest struct {
	Contract      *types.Address
	Token         *types.Address
	AmountMinimum *big.Int
	Recipient     *types.Address
	FeeBips       *big.Int
	FeeRecipient  *types.Address
}

func NewRouterSweepTokenWithFeeRequest(contract, token *types.Address, amountMinimum *big.Int, recipient *types.Address, feeBips *big.Int, feeRecipient *types.Address) *RouterSweepTokenWithFeeRequest {
	return &RouterSweepTokenWithFeeRequest{
		Contract:      contract,
		Token:         token,
		AmountMinimum: amountMinimum,
		Recipient:     recipient,
		FeeBips:       feeBips,
		FeeRecipient:  feeRecipient,
	}
}

func (r *RouterSweepTokenWithFeeRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Token == nil {
		return fmt.Errorf("%w: token is nil", bind.ErrInvalidRequest)
	}

	if r.AmountMinimum == nil {
		return fmt.Errorf("%w: amountMinimum is nil", bind.ErrInvalidRequest)
	}

	if r.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	if r.FeeBips == nil {
		return fmt.Errorf("%w: feeBips is nil", bind.ErrInvalidRequest)
	}

	if r.FeeRecipient == nil {
		return fmt.Errorf("%w: feeRecipient is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// RouterRefundETHRequest is refundETH()'s input on Contract.
type RouterRefundETHRequest struct {
	Contract *types.Address
}

func NewRouterRefundETHRequest(contract *types.Address) *RouterRefundETHRequest {
	return &RouterRefundETHRequest{
		Contract: contract,
	}
}

func (r *RouterRefundETHRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// RouterMulticallRequest is multicall(data) or multicall(deadline, data)'s input on Contract; a non-nil Deadline selects the multicall(deadline, data) overload, and NativeAmount, if set, is sent as msg.value.
type RouterMulticallRequest struct {
	Contract     *types.Address
	Data         [][]byte
	Deadline     *big.Int
	NativeAmount *big.Int
}

func NewRouterMulticallRequest(contract *types.Address, data [][]byte, deadline, nativeAmount *big.Int) *RouterMulticallRequest {
	return &RouterMulticallRequest{
		Contract:     contract,
		Data:         data,
		Deadline:     deadline,
		NativeAmount: nativeAmount,
	}
}

func (r *RouterMulticallRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if len(r.Data) == 0 {
		return fmt.Errorf("%w: data is empty", bind.ErrInvalidRequest)
	}

	return nil
}

// SelfPermitRequest is selfPermit(token, value, deadline, v, r, s)'s input on Contract.
type SelfPermitRequest struct {
	Contract *types.Address
	Token    *types.Address
	Amount   *big.Int
	Deadline *big.Int
	Sig      *types.Signature
}

func NewSelfPermitRequest(contract, token *types.Address, amount, deadline *big.Int, sig *types.Signature) *SelfPermitRequest {
	return &SelfPermitRequest{
		Contract: contract,
		Token:    token,
		Amount:   amount,
		Deadline: deadline,
		Sig:      sig,
	}
}

func (r *SelfPermitRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Token == nil {
		return fmt.Errorf("%w: token is nil", bind.ErrInvalidRequest)
	}

	if r.Amount == nil {
		return fmt.Errorf("%w: amount is nil", bind.ErrInvalidRequest)
	}

	if r.Deadline == nil {
		return fmt.Errorf("%w: deadline is nil", bind.ErrInvalidRequest)
	}

	if r.Sig == nil {
		return fmt.Errorf("%w: sig is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SelfPermitIfNecessaryRequest is selfPermitIfNecessary(token, value, deadline, v, r, s)'s input on Contract.
type SelfPermitIfNecessaryRequest struct {
	Contract *types.Address
	Token    *types.Address
	Amount   *big.Int
	Deadline *big.Int
	Sig      *types.Signature
}

func NewSelfPermitIfNecessaryRequest(contract, token *types.Address, amount, deadline *big.Int, sig *types.Signature) *SelfPermitIfNecessaryRequest {
	return &SelfPermitIfNecessaryRequest{
		Contract: contract,
		Token:    token,
		Amount:   amount,
		Deadline: deadline,
		Sig:      sig,
	}
}

func (r *SelfPermitIfNecessaryRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Token == nil {
		return fmt.Errorf("%w: token is nil", bind.ErrInvalidRequest)
	}

	if r.Amount == nil {
		return fmt.Errorf("%w: amount is nil", bind.ErrInvalidRequest)
	}

	if r.Deadline == nil {
		return fmt.Errorf("%w: deadline is nil", bind.ErrInvalidRequest)
	}

	if r.Sig == nil {
		return fmt.Errorf("%w: sig is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SelfPermitAllowedRequest is selfPermitAllowed(token, nonce, expiry, v, r, s)'s input on Contract.
type SelfPermitAllowedRequest struct {
	Contract *types.Address
	Token    *types.Address
	Nonce    *big.Int
	Expiry   *big.Int
	Sig      *types.Signature
}

func NewSelfPermitAllowedRequest(contract, token *types.Address, nonce, expiry *big.Int, sig *types.Signature) *SelfPermitAllowedRequest {
	return &SelfPermitAllowedRequest{
		Contract: contract,
		Token:    token,
		Nonce:    nonce,
		Expiry:   expiry,
		Sig:      sig,
	}
}

func (r *SelfPermitAllowedRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Token == nil {
		return fmt.Errorf("%w: token is nil", bind.ErrInvalidRequest)
	}

	if r.Nonce == nil {
		return fmt.Errorf("%w: nonce is nil", bind.ErrInvalidRequest)
	}

	if r.Expiry == nil {
		return fmt.Errorf("%w: expiry is nil", bind.ErrInvalidRequest)
	}

	if r.Sig == nil {
		return fmt.Errorf("%w: sig is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SelfPermitAllowedIfNecessaryRequest is selfPermitAllowedIfNecessary(token, nonce, expiry, v, r, s)'s input on Contract.
type SelfPermitAllowedIfNecessaryRequest struct {
	Contract *types.Address
	Token    *types.Address
	Nonce    *big.Int
	Expiry   *big.Int
	Sig      *types.Signature
}

func NewSelfPermitAllowedIfNecessaryRequest(contract, token *types.Address, nonce, expiry *big.Int, sig *types.Signature) *SelfPermitAllowedIfNecessaryRequest {
	return &SelfPermitAllowedIfNecessaryRequest{
		Contract: contract,
		Token:    token,
		Nonce:    nonce,
		Expiry:   expiry,
		Sig:      sig,
	}
}

func (r *SelfPermitAllowedIfNecessaryRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Token == nil {
		return fmt.Errorf("%w: token is nil", bind.ErrInvalidRequest)
	}

	if r.Nonce == nil {
		return fmt.Errorf("%w: nonce is nil", bind.ErrInvalidRequest)
	}

	if r.Expiry == nil {
		return fmt.Errorf("%w: expiry is nil", bind.ErrInvalidRequest)
	}

	if r.Sig == nil {
		return fmt.Errorf("%w: sig is nil", bind.ErrInvalidRequest)
	}

	return nil
}

func validateExactInputSingleParams(p *ExactInputSingleParams) error {
	if p == nil {
		return fmt.Errorf("%w: params is nil", bind.ErrInvalidRequest)
	}

	if p.TokenIn == nil {
		return fmt.Errorf("%w: params.tokenIn is nil", bind.ErrInvalidRequest)
	}

	if p.TokenOut == nil {
		return fmt.Errorf("%w: params.tokenOut is nil", bind.ErrInvalidRequest)
	}

	if p.Fee == nil {
		return fmt.Errorf("%w: params.fee is nil", bind.ErrInvalidRequest)
	}

	if p.Recipient == nil {
		return fmt.Errorf("%w: params.recipient is nil", bind.ErrInvalidRequest)
	}

	if p.AmountIn == nil {
		return fmt.Errorf("%w: params.amountIn is nil", bind.ErrInvalidRequest)
	}

	if p.AmountOutMinimum == nil {
		return fmt.Errorf("%w: params.amountOutMinimum is nil", bind.ErrInvalidRequest)
	}

	if p.SqrtPriceLimitX96 == nil {
		return fmt.Errorf("%w: params.sqrtPriceLimitX96 is nil", bind.ErrInvalidRequest)
	}

	return nil
}

func validateExactInputParams(p *ExactInputParams) error {
	if p == nil {
		return fmt.Errorf("%w: params is nil", bind.ErrInvalidRequest)
	}

	if len(p.Path) == 0 {
		return fmt.Errorf("%w: params.path is empty", bind.ErrInvalidRequest)
	}

	if p.Recipient == nil {
		return fmt.Errorf("%w: params.recipient is nil", bind.ErrInvalidRequest)
	}

	if p.AmountIn == nil {
		return fmt.Errorf("%w: params.amountIn is nil", bind.ErrInvalidRequest)
	}

	if p.AmountOutMinimum == nil {
		return fmt.Errorf("%w: params.amountOutMinimum is nil", bind.ErrInvalidRequest)
	}

	return nil
}

func validateExactOutputSingleParams(p *ExactOutputSingleParams) error {
	if p == nil {
		return fmt.Errorf("%w: params is nil", bind.ErrInvalidRequest)
	}

	if p.TokenIn == nil {
		return fmt.Errorf("%w: params.tokenIn is nil", bind.ErrInvalidRequest)
	}

	if p.TokenOut == nil {
		return fmt.Errorf("%w: params.tokenOut is nil", bind.ErrInvalidRequest)
	}

	if p.Fee == nil {
		return fmt.Errorf("%w: params.fee is nil", bind.ErrInvalidRequest)
	}

	if p.Recipient == nil {
		return fmt.Errorf("%w: params.recipient is nil", bind.ErrInvalidRequest)
	}

	if p.AmountOut == nil {
		return fmt.Errorf("%w: params.amountOut is nil", bind.ErrInvalidRequest)
	}

	if p.AmountInMaximum == nil {
		return fmt.Errorf("%w: params.amountInMaximum is nil", bind.ErrInvalidRequest)
	}

	if p.SqrtPriceLimitX96 == nil {
		return fmt.Errorf("%w: params.sqrtPriceLimitX96 is nil", bind.ErrInvalidRequest)
	}

	return nil
}

func validateExactOutputParams(p *ExactOutputParams) error {
	if p == nil {
		return fmt.Errorf("%w: params is nil", bind.ErrInvalidRequest)
	}

	if len(p.Path) == 0 {
		return fmt.Errorf("%w: params.path is empty", bind.ErrInvalidRequest)
	}

	if p.Recipient == nil {
		return fmt.Errorf("%w: params.recipient is nil", bind.ErrInvalidRequest)
	}

	if p.AmountOut == nil {
		return fmt.Errorf("%w: params.amountOut is nil", bind.ErrInvalidRequest)
	}

	if p.AmountInMaximum == nil {
		return fmt.Errorf("%w: params.amountInMaximum is nil", bind.ErrInvalidRequest)
	}

	return nil
}

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
