package v3

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// IRouterImmutables mirrors PeripheryImmutableState, the only state SwapRouter itself holds.
type IRouterImmutables interface {
	Factory(ctx context.Context, contract *types.Address, block string) (*types.Address, error)
	WETH9(ctx context.Context, contract *types.Address, block string) (*types.Address, error)
}

// IRouterSwap mirrors ISwapRouter. No tx.Builder-taking Create* constructors here -- see
// CreateExactInputSingle, CreateExactInput, CreateExactOutputSingle, and CreateExactOutput.
type IRouterSwap interface {
	ExactInputSingle(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	ExactInput(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	ExactOutputSingle(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	ExactOutput(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

// IRouterPayments mirrors PeripheryPaymentsWithFee (and the PeripheryPayments it extends). No
// tx.Builder-taking Create* constructors here -- see CreateUnwrapWETH9, CreateUnwrapWETH9WithFee,
// CreateSweepToken, CreateSweepTokenWithFee, and CreateRefundETH.
type IRouterPayments interface {
	UnwrapWETH9(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	UnwrapWETH9WithFee(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SweepToken(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SweepTokenWithFee(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	RefundETH(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

// IRouterMulticall mirrors Multicall. No tx.Builder-taking Create* constructor here -- see
// CreateMulticall.
type IRouterMulticall interface {
	Multicall(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

// IRouterSelfPermit mirrors SelfPermit. No tx.Builder-taking Create* constructors here -- see
// CreateSelfPermit, CreateSelfPermitIfNecessary, CreateSelfPermitAllowed, and
// CreateSelfPermitAllowedIfNecessary.
type IRouterSelfPermit interface {
	SelfPermit(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SelfPermitIfNecessary(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SelfPermitAllowed(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SelfPermitAllowedIfNecessary(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

// IRouter mirrors SwapRouter, composed the same way its Solidity base contracts are: one small
// interface per concern (immutables/swap/payments/multicall/self-permit).
type IRouter interface {
	IRouterImmutables
	IRouterSwap
	IRouterPayments
	IRouterMulticall
	IRouterSelfPermit
}

var _ IRouter = (*Router)(nil)

// Router calls any SwapRouter contract over a Client; it holds no contract address, so one
// instance serves every router on every network.
type Router struct {
	c rpc.Client
}

// NewRouter returns a Router calling routers via c.
func NewRouter(c rpc.Client) *Router {
	return &Router{
		c: c,
	}
}

// Factory returns the UniswapV3Factory contract's address.
func (r *Router) Factory(ctx context.Context, contract *types.Address, block string) (*types.Address, error) {
	data, err := r.call(ctx, nil, contract, nil, EncodeRouterFactory(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: factory: %w", err)
	}
	addr, err := DecodeRouterFactory(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: factory: %w", err)
	}
	return addr, nil
}

// WETH9 returns the WETH9 contract's address.
func (r *Router) WETH9(ctx context.Context, contract *types.Address, block string) (*types.Address, error) {
	data, err := r.call(ctx, nil, contract, nil, EncodeRouterWETH9(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: weth9: %w", err)
	}
	addr, err := DecodeRouterWETH9(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: weth9: %w", err)
	}
	return addr, nil
}

// CreateExactInputSingle builds an unsigned exactInputSingle(params) transaction via b, sent by
// sender with value wei attached (value may be nil for a non-ETH-input swap).
func (r *Router) CreateExactInputSingle(
	ctx context.Context,
	b tx.Builder,
	contract, sender *types.Address,
	value *big.Int,
	params *ExactInputSingleParams,
) (core.Transaction, error) {
	calldata, err := EncodeExactInputSingle(
		params.TokenIn,
		params.TokenOut,
		params.Fee,
		params.Recipient,
		params.AmountIn,
		params.AmountOutMinimum,
		params.SqrtPriceLimitX96,
	)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createExactInputSingle: %w", err)
	}

	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createExactInputSingle: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// CallExactInputSingle simulates exactInputSingle(params) via eth_call, as sender with value wei
// attached, returning the amount of tokenOut it would receive -- without broadcasting anything.
func (r *Router) CallExactInputSingle(
	ctx context.Context,
	contract, sender *types.Address,
	value *big.Int,
	params *ExactInputSingleParams,
	block string,
) (amountOut *big.Int, err error) {
	calldata, err := EncodeExactInputSingle(
		params.TokenIn,
		params.TokenOut,
		params.Fee,
		params.Recipient,
		params.AmountIn,
		params.AmountOutMinimum,
		params.SqrtPriceLimitX96,
	)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactInputSingle: %w", err)
	}
	data, err := r.call(ctx, sender, contract, value, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactInputSingle: %w", err)
	}
	amountOut, err = DecodeExactInputSingle(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactInputSingle: %w", err)
	}
	return amountOut, nil
}

// ExactInputSingle broadcasts an already-signed exactInputSingle transaction.
func (r *Router) ExactInputSingle(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: exactInputSingle: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: exactInputSingle: %w", err)
	}
	return hash, nil
}

// CreateExactInput builds an unsigned exactInput(params) transaction via b, sent by sender with
// value wei attached (value may be nil for a non-ETH-input swap).
func (r *Router) CreateExactInput(
	ctx context.Context,
	b tx.Builder,
	contract, sender *types.Address,
	value *big.Int,
	params *ExactInputParams,
) (core.Transaction, error) {
	calldata, err := EncodeExactInput(
		params.Path,
		params.Recipient,
		params.AmountIn,
		params.AmountOutMinimum,
	)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createExactInput: %w", err)
	}

	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createExactInput: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// CallExactInput simulates exactInput(params) via eth_call, as sender with value wei attached,
// returning the amount of the final token in path it would receive -- without broadcasting
// anything.
func (r *Router) CallExactInput(
	ctx context.Context,
	contract, sender *types.Address,
	value *big.Int,
	params *ExactInputParams,
	block string,
) (amountOut *big.Int, err error) {
	calldata, err := EncodeExactInput(
		params.Path,
		params.Recipient,
		params.AmountIn,
		params.AmountOutMinimum,
	)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactInput: %w", err)
	}
	data, err := r.call(ctx, sender, contract, value, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactInput: %w", err)
	}
	amountOut, err = DecodeExactInput(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactInput: %w", err)
	}
	return amountOut, nil
}

// ExactInput broadcasts an already-signed exactInput transaction.
func (r *Router) ExactInput(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: exactInput: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: exactInput: %w", err)
	}
	return hash, nil
}

// CreateExactOutputSingle builds an unsigned exactOutputSingle(params) transaction via b, sent by
// sender with value wei attached (value may be nil for a non-ETH-input swap).
func (r *Router) CreateExactOutputSingle(
	ctx context.Context,
	b tx.Builder,
	contract, sender *types.Address,
	value *big.Int,
	params *ExactOutputSingleParams,
) (core.Transaction, error) {
	calldata, err := EncodeExactOutputSingle(
		params.TokenIn,
		params.TokenOut,
		params.Fee,
		params.Recipient,
		params.AmountOut,
		params.AmountInMaximum,
		params.SqrtPriceLimitX96,
	)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createExactOutputSingle: %w", err)
	}

	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createExactOutputSingle: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// CallExactOutputSingle simulates exactOutputSingle(params) via eth_call, as sender with value
// wei attached, returning the amount of tokenIn it would require -- without broadcasting
// anything.
func (r *Router) CallExactOutputSingle(
	ctx context.Context,
	contract, sender *types.Address,
	value *big.Int,
	params *ExactOutputSingleParams,
	block string,
) (amountIn *big.Int, err error) {
	calldata, err := EncodeExactOutputSingle(
		params.TokenIn,
		params.TokenOut,
		params.Fee,
		params.Recipient,
		params.AmountOut,
		params.AmountInMaximum,
		params.SqrtPriceLimitX96,
	)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactOutputSingle: %w", err)
	}
	data, err := r.call(ctx, sender, contract, value, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactOutputSingle: %w", err)
	}
	amountIn, err = DecodeExactOutputSingle(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactOutputSingle: %w", err)
	}
	return amountIn, nil
}

// ExactOutputSingle broadcasts an already-signed exactOutputSingle transaction.
func (r *Router) ExactOutputSingle(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: exactOutputSingle: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: exactOutputSingle: %w", err)
	}
	return hash, nil
}

// CreateExactOutput builds an unsigned exactOutput(params) transaction via b, sent by sender with
// value wei attached (value may be nil for a non-ETH-input swap).
func (r *Router) CreateExactOutput(
	ctx context.Context,
	b tx.Builder,
	contract, sender *types.Address,
	value *big.Int,
	params *ExactOutputParams,
) (core.Transaction, error) {
	calldata, err := EncodeExactOutput(
		params.Path,
		params.Recipient,
		params.AmountOut,
		params.AmountInMaximum,
	)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createExactOutput: %w", err)
	}

	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createExactOutput: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// CallExactOutput simulates exactOutput(params) via eth_call, as sender with value wei attached,
// returning the amount of the first token in path it would require -- without broadcasting
// anything.
func (r *Router) CallExactOutput(
	ctx context.Context,
	contract, sender *types.Address,
	value *big.Int,
	params *ExactOutputParams,
	block string,
) (amountIn *big.Int, err error) {
	calldata, err := EncodeExactOutput(
		params.Path,
		params.Recipient,
		params.AmountOut,
		params.AmountInMaximum,
	)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactOutput: %w", err)
	}
	data, err := r.call(ctx, sender, contract, value, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactOutput: %w", err)
	}
	amountIn, err = DecodeExactOutput(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callExactOutput: %w", err)
	}
	return amountIn, nil
}

// ExactOutput broadcasts an already-signed exactOutput transaction.
func (r *Router) ExactOutput(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: exactOutput: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: exactOutput: %w", err)
	}
	return hash, nil
}

// CreateUnwrapWETH9 builds an unsigned unwrapWETH9(amountMinimum, recipient) transaction via b,
// sent by sender.
func (r *Router) CreateUnwrapWETH9(
	ctx context.Context,
	b tx.Builder,
	contract, sender *types.Address,
	amountMinimum *big.Int,
	recipient *types.Address,
) (core.Transaction, error) {
	calldata, err := EncodeRouterUnwrapWETH9(amountMinimum, recipient)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createUnwrapWETH9: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createUnwrapWETH9: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// UnwrapWETH9 broadcasts an already-signed unwrapWETH9 transaction.
func (r *Router) UnwrapWETH9(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: unwrapWETH9: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: unwrapWETH9: %w", err)
	}
	return hash, nil
}

// CreateUnwrapWETH9WithFee builds an unsigned unwrapWETH9WithFee(amountMinimum, recipient,
// feeBips, feeRecipient) transaction via b, sent by sender.
func (r *Router) CreateUnwrapWETH9WithFee(
	ctx context.Context,
	b tx.Builder,
	contract, sender *types.Address,
	amountMinimum *big.Int,
	recipient *types.Address,
	feeBips *big.Int,
	feeRecipient *types.Address,
) (core.Transaction, error) {
	calldata, err := EncodeUnwrapWETH9WithFee(amountMinimum, recipient, feeBips, feeRecipient)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createUnwrapWETH9WithFee: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createUnwrapWETH9WithFee: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// UnwrapWETH9WithFee broadcasts an already-signed unwrapWETH9WithFee transaction.
func (r *Router) UnwrapWETH9WithFee(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: unwrapWETH9WithFee: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: unwrapWETH9WithFee: %w", err)
	}
	return hash, nil
}

// CreateSweepToken builds an unsigned sweepToken(token, amountMinimum, recipient) transaction
// via b, sent by sender.
func (r *Router) CreateSweepToken(
	ctx context.Context,
	b tx.Builder,
	contract, sender, token *types.Address,
	amountMinimum *big.Int,
	recipient *types.Address,
) (core.Transaction, error) {
	calldata, err := EncodeRouterSweepToken(token, amountMinimum, recipient)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createSweepToken: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createSweepToken: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// SweepToken broadcasts an already-signed sweepToken transaction.
func (r *Router) SweepToken(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: sweepToken: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: sweepToken: %w", err)
	}
	return hash, nil
}

// CreateSweepTokenWithFee builds an unsigned sweepTokenWithFee(token, amountMinimum, recipient,
// feeBips, feeRecipient) transaction via b, sent by sender.
func (r *Router) CreateSweepTokenWithFee(
	ctx context.Context,
	b tx.Builder,
	contract, sender, token *types.Address,
	amountMinimum *big.Int,
	recipient *types.Address,
	feeBips *big.Int,
	feeRecipient *types.Address,
) (core.Transaction, error) {
	calldata, err := EncodeSweepTokenWithFee(token, amountMinimum, recipient, feeBips, feeRecipient)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createSweepTokenWithFee: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createSweepTokenWithFee: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// SweepTokenWithFee broadcasts an already-signed sweepTokenWithFee transaction.
func (r *Router) SweepTokenWithFee(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: sweepTokenWithFee: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: sweepTokenWithFee: %w", err)
	}
	return hash, nil
}

// CreateRefundETH builds an unsigned refundETH() transaction via b, sent by sender.
func (r *Router) CreateRefundETH(
	ctx context.Context,
	b tx.Builder,
	contract, sender *types.Address,
) (core.Transaction, error) {
	unsigned, err := b.Build(sender, contract, nil, EncodeRouterRefundETH()).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createRefundETH: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// RefundETH broadcasts an already-signed refundETH transaction.
func (r *Router) RefundETH(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: refundETH: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: refundETH: %w", err)
	}
	return hash, nil
}

// CreateMulticall builds an unsigned multicall(data) transaction via b, sent by sender with
// value wei attached (value may be nil).
func (r *Router) CreateMulticall(
	ctx context.Context,
	b tx.Builder,
	contract, sender *types.Address,
	value *big.Int,
	data [][]byte,
) (core.Transaction, error) {
	calldata, err := EncodeRouterMulticall(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createMulticall: %w", err)
	}

	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createMulticall: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// CallMulticall simulates multicall(data) via eth_call, as sender with value wei attached,
// returning each sub-call's own return data -- without broadcasting anything.
func (r *Router) CallMulticall(
	ctx context.Context,
	contract, sender *types.Address,
	value *big.Int,
	data [][]byte,
	block string,
) ([][]byte, error) {
	calldata, err := EncodeRouterMulticall(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callMulticall: %w", err)
	}
	raw, err := r.call(ctx, sender, contract, value, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callMulticall: %w", err)
	}
	results, err := DecodeRouterMulticall(raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callMulticall: %w", err)
	}
	return results, nil
}

// Multicall broadcasts an already-signed multicall transaction.
func (r *Router) Multicall(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: multicall: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: multicall: %w", err)
	}
	return hash, nil
}

// CreateSelfPermit builds an unsigned selfPermit(token, value, deadline, v, r, s) transaction
// via b, sent by sender.
func (r *Router) CreateSelfPermit(
	ctx context.Context,
	b tx.Builder,
	contract, sender, token *types.Address,
	value, deadline *big.Int,
	sig *types.Signature,
) (core.Transaction, error) {
	calldata, err := EncodeSelfPermit(token, value, deadline, sig)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createSelfPermit: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createSelfPermit: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// SelfPermit broadcasts an already-signed selfPermit transaction.
func (r *Router) SelfPermit(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: selfPermit: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: selfPermit: %w", err)
	}
	return hash, nil
}

// CreateSelfPermitIfNecessary builds an unsigned selfPermitIfNecessary(token, value, deadline,
// v, r, s) transaction via b, sent by sender.
func (r *Router) CreateSelfPermitIfNecessary(
	ctx context.Context,
	b tx.Builder,
	contract, sender, token *types.Address,
	value, deadline *big.Int,
	sig *types.Signature,
) (core.Transaction, error) {
	calldata, err := EncodeSelfPermitIfNecessary(token, value, deadline, sig)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createSelfPermitIfNecessary: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createSelfPermitIfNecessary: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// SelfPermitIfNecessary broadcasts an already-signed selfPermitIfNecessary transaction.
func (r *Router) SelfPermitIfNecessary(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: selfPermitIfNecessary: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: selfPermitIfNecessary: %w", err)
	}
	return hash, nil
}

// CreateSelfPermitAllowed builds an unsigned selfPermitAllowed(token, nonce, expiry, v, r, s)
// transaction via b, sent by sender.
func (r *Router) CreateSelfPermitAllowed(
	ctx context.Context,
	b tx.Builder,
	contract, sender, token *types.Address,
	nonce, expiry *big.Int,
	sig *types.Signature,
) (core.Transaction, error) {
	calldata, err := EncodeSelfPermitAllowed(token, nonce, expiry, sig)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createSelfPermitAllowed: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createSelfPermitAllowed: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// SelfPermitAllowed broadcasts an already-signed selfPermitAllowed transaction.
func (r *Router) SelfPermitAllowed(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: selfPermitAllowed: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: selfPermitAllowed: %w", err)
	}
	return hash, nil
}

// CreateSelfPermitAllowedIfNecessary builds an unsigned selfPermitAllowedIfNecessary(token,
// nonce, expiry, v, r, s) transaction via b, sent by sender.
func (r *Router) CreateSelfPermitAllowedIfNecessary(
	ctx context.Context,
	b tx.Builder,
	contract, sender, token *types.Address,
	nonce, expiry *big.Int,
	sig *types.Signature,
) (core.Transaction, error) {
	calldata, err := EncodeSelfPermitAllowedIfNecessary(token, nonce, expiry, sig)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createSelfPermitAllowedIfNecessary: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createSelfPermitAllowedIfNecessary: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// SelfPermitAllowedIfNecessary broadcasts an already-signed selfPermitAllowedIfNecessary transaction.
func (r *Router) SelfPermitAllowedIfNecessary(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: selfPermitAllowedIfNecessary: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: selfPermitAllowedIfNecessary: %w", err)
	}
	return hash, nil
}

// call eth_calls contract with calldata, optionally simulating from a specific sender.
func (r *Router) call(
	ctx context.Context,
	from, contract *types.Address,
	value *big.Int,
	calldata []byte,
	block string,
) ([]byte, error) {
	params := map[string]any{
		"to":   contract.String(),
		"data": encoding.Hex.EncodePrefixed(calldata),
	}
	if from != nil {
		params["from"] = from.String()
	}
	if value != nil && value.Sign() != 0 {
		params["value"] = "0x" + value.Text(16)
	}

	var raw string
	if err := r.c.Call(ctx, rpc.ETHCall("call", params, block, &raw)); err != nil {
		return nil, r.wrapRevertError(err)
	}
	return encoding.Hex.Decode(raw)
}

// wrapRevertError maps a known SwapRouter revert reason in err to its sentinel.
func (_ *Router) wrapRevertError(err error) error {
	var rpcErr *rpc.ResponseError
	if !errors.As(err, &rpcErr) {
		return err
	}
	if len(rpcErr.Data) == 0 {
		return fmt.Errorf("%w: %w", ErrReverted, err)
	}

	var hexData string
	if jsonErr := json.Unmarshal(rpcErr.Data, &hexData); jsonErr != nil {
		return fmt.Errorf("%w: %w", ErrReverted, err)
	}

	data, decErr := encoding.Hex.Decode(hexData)
	if decErr != nil {
		return fmt.Errorf("%w: %w", ErrReverted, err)
	}

	reason, decErr := abi.DecodeRevert(data)
	if decErr != nil {
		return fmt.Errorf("%w: %w", ErrReverted, err)
	}

	switch reason {
	case reasonTransactionTooOld:
		return fmt.Errorf("%w: %w", ErrTransactionTooOld, err)
	case reasonTooLittleReceived:
		return fmt.Errorf("%w: %w", ErrTooLittleReceived, err)
	case reasonTooMuchRequested:
		return fmt.Errorf("%w: %w", ErrTooMuchRequested, err)
	case reasonInsufficientWETH9:
		return fmt.Errorf("%w: %w", ErrInsufficientWETH9, err)
	case reasonInsufficientToken:
		return fmt.Errorf("%w: %w", ErrInsufficientToken, err)
	case reasonNotWETH9:
		return fmt.Errorf("%w: %w", ErrNotWETH9, err)
	case reasonTransferFromFailed:
		return fmt.Errorf("%w: %w", ErrTransferFromFailed, err)
	case reasonTokenTransferFailed:
		return fmt.Errorf("%w: %w", ErrTokenTransferFailed, err)
	case reasonApproveFailed:
		return fmt.Errorf("%w: %w", ErrApproveFailed, err)
	case reasonTransferETHFailed:
		return fmt.Errorf("%w: %w", ErrTransferETHFailed, err)
	default:
		return fmt.Errorf("%w: %w", ErrReverted, err)
	}
}
