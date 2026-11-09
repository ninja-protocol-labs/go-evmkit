package v3

import (
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// RouterUnwrapWETH9Params is CreateUnwrapWETH9's argument list.
type RouterUnwrapWETH9Params struct {
	Router        *types.Address
	Sender        *types.Address
	AmountMinimum *big.Int
	Recipient     *types.Address
}

func NewRouterUnwrapWETH9Params(router, sender *types.Address, amountMinimum *big.Int, recipient *types.Address) *RouterUnwrapWETH9Params {
	return &RouterUnwrapWETH9Params{
		Router:        router,
		Sender:        sender,
		AmountMinimum: amountMinimum,
		Recipient:     recipient,
	}
}

// RouterUnwrapWETH9WithFeeParams is CreateUnwrapWETH9WithFee's argument list.
type RouterUnwrapWETH9WithFeeParams struct {
	Router        *types.Address
	Sender        *types.Address
	AmountMinimum *big.Int
	Recipient     *types.Address
	FeeBips       *big.Int
	FeeRecipient  *types.Address
}

func NewRouterUnwrapWETH9WithFeeParams(router, sender *types.Address, amountMinimum *big.Int, recipient *types.Address, feeBips *big.Int, feeRecipient *types.Address) *RouterUnwrapWETH9WithFeeParams {
	return &RouterUnwrapWETH9WithFeeParams{
		Router:        router,
		Sender:        sender,
		AmountMinimum: amountMinimum,
		Recipient:     recipient,
		FeeBips:       feeBips,
		FeeRecipient:  feeRecipient,
	}
}

// RouterSweepTokenParams is CreateSweepToken's argument list.
type RouterSweepTokenParams struct {
	Router        *types.Address
	Sender        *types.Address
	Token         *types.Address
	AmountMinimum *big.Int
	Recipient     *types.Address
}

func NewRouterSweepTokenParams(router, sender, token *types.Address, amountMinimum *big.Int, recipient *types.Address) *RouterSweepTokenParams {
	return &RouterSweepTokenParams{
		Router:        router,
		Sender:        sender,
		Token:         token,
		AmountMinimum: amountMinimum,
		Recipient:     recipient,
	}
}

// RouterSweepTokenWithFeeParams is CreateSweepTokenWithFee's argument list.
type RouterSweepTokenWithFeeParams struct {
	Router        *types.Address
	Sender        *types.Address
	Token         *types.Address
	AmountMinimum *big.Int
	Recipient     *types.Address
	FeeBips       *big.Int
	FeeRecipient  *types.Address
}

func NewRouterSweepTokenWithFeeParams(router, sender, token *types.Address, amountMinimum *big.Int, recipient *types.Address, feeBips *big.Int, feeRecipient *types.Address) *RouterSweepTokenWithFeeParams {
	return &RouterSweepTokenWithFeeParams{
		Router:        router,
		Sender:        sender,
		Token:         token,
		AmountMinimum: amountMinimum,
		Recipient:     recipient,
		FeeBips:       feeBips,
		FeeRecipient:  feeRecipient,
	}
}

// RouterRefundETHParams is CreateRefundETH's argument list.
type RouterRefundETHParams struct {
	Router *types.Address
	Sender *types.Address
}

func NewRouterRefundETHParams(router, sender *types.Address) *RouterRefundETHParams {
	return &RouterRefundETHParams{
		Router: router,
		Sender: sender,
	}
}

// RouterMulticallParams is CreateMulticall's and CallMulticall's argument list.
type RouterMulticallParams struct {
	Router *types.Address
	Sender *types.Address
	Value  *big.Int
	Data   [][]byte
}

func NewRouterMulticallParams(router, sender *types.Address, value *big.Int, data [][]byte) *RouterMulticallParams {
	return &RouterMulticallParams{
		Router: router,
		Sender: sender,
		Value:  value,
		Data:   data,
	}
}

// SelfPermitParams is CreateSelfPermit's argument list.
type SelfPermitParams struct {
	Router   *types.Address
	Sender   *types.Address
	Token    *types.Address
	Value    *big.Int
	Deadline *big.Int
	Sig      *types.Signature
}

func NewSelfPermitParams(router, sender, token *types.Address, value, deadline *big.Int, sig *types.Signature) *SelfPermitParams {
	return &SelfPermitParams{
		Router:   router,
		Sender:   sender,
		Token:    token,
		Value:    value,
		Deadline: deadline,
		Sig:      sig,
	}
}

// SelfPermitIfNecessaryParams is CreateSelfPermitIfNecessary's argument list.
type SelfPermitIfNecessaryParams struct {
	Router   *types.Address
	Sender   *types.Address
	Token    *types.Address
	Value    *big.Int
	Deadline *big.Int
	Sig      *types.Signature
}

func NewSelfPermitIfNecessaryParams(router, sender, token *types.Address, value, deadline *big.Int, sig *types.Signature) *SelfPermitIfNecessaryParams {
	return &SelfPermitIfNecessaryParams{
		Router:   router,
		Sender:   sender,
		Token:    token,
		Value:    value,
		Deadline: deadline,
		Sig:      sig,
	}
}

// SelfPermitAllowedParams is CreateSelfPermitAllowed's argument list.
type SelfPermitAllowedParams struct {
	Router *types.Address
	Sender *types.Address
	Token  *types.Address
	Nonce  *big.Int
	Expiry *big.Int
	Sig    *types.Signature
}

func NewSelfPermitAllowedParams(router, sender, token *types.Address, nonce, expiry *big.Int, sig *types.Signature) *SelfPermitAllowedParams {
	return &SelfPermitAllowedParams{
		Router: router,
		Sender: sender,
		Token:  token,
		Nonce:  nonce,
		Expiry: expiry,
		Sig:    sig,
	}
}

// SelfPermitAllowedIfNecessaryParams is CreateSelfPermitAllowedIfNecessary's argument list.
type SelfPermitAllowedIfNecessaryParams struct {
	Router *types.Address
	Sender *types.Address
	Token  *types.Address
	Nonce  *big.Int
	Expiry *big.Int
	Sig    *types.Signature
}

func NewSelfPermitAllowedIfNecessaryParams(router, sender, token *types.Address, nonce, expiry *big.Int, sig *types.Signature) *SelfPermitAllowedIfNecessaryParams {
	return &SelfPermitAllowedIfNecessaryParams{
		Router: router,
		Sender: sender,
		Token:  token,
		Nonce:  nonce,
		Expiry: expiry,
		Sig:    sig,
	}
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
