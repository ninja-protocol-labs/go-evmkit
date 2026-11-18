package v4

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// IQuoterImmutables mirrors IImmutableState plus IMsgSender, the only state V4Quoter exposes.
type IQuoterImmutables interface {
	PoolManager(ctx context.Context, contract *types.Address, block string) (*types.Address, error)
	MsgSender(ctx context.Context, contract *types.Address, block string) (*types.Address, error)
}

// IQuoterQuotes mirrors IV4Quoter's quoting functions. They simulate a swap and are never meant to
// be broadcast, so only eth_call-backed Call* methods exist.
type IQuoterQuotes interface {
	CallQuoteExactInputSingle(ctx context.Context, contract, sender *types.Address, params *QuoteExactSingleParams, block string) (*QuoteExactInputResult, error)
	CallQuoteExactInput(ctx context.Context, contract, sender *types.Address, params *QuoteExactParams, block string) (*QuoteExactInputResult, error)
	CallQuoteExactOutputSingle(ctx context.Context, contract, sender *types.Address, params *QuoteExactSingleParams, block string) (*QuoteExactOutputResult, error)
	CallQuoteExactOutput(ctx context.Context, contract, sender *types.Address, params *QuoteExactParams, block string) (*QuoteExactOutputResult, error)
}

// IQuoter mirrors IV4Quoter.
type IQuoter interface {
	IQuoterImmutables
	IQuoterQuotes
}

var _ IQuoter = (*Quoter)(nil)

// Quoter calls any V4Quoter contract over a Client; it holds no contract address, so one instance
// serves every network's deployment.
type Quoter struct {
	c rpc.Client
}

// NewQuoter returns a Quoter calling V4Quoter contracts via c.
func NewQuoter(c rpc.Client) *Quoter {
	return &Quoter{
		c: c,
	}
}

// PoolManager returns contract's immutable PoolManager address.
func (q *Quoter) PoolManager(ctx context.Context, contract *types.Address, block string) (*types.Address, error) {
	data, err := q.call(ctx, nil, contract, EncodeQuoterPoolManager(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: quoterPoolManager: %w", err)
	}
	addr, err := DecodeQuoterPoolManager(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: quoterPoolManager: %w", err)
	}
	return addr, nil
}

// MsgSender returns contract's current msgSender; the zero address outside a quote.
func (q *Quoter) MsgSender(ctx context.Context, contract *types.Address, block string) (*types.Address, error) {
	data, err := q.call(ctx, nil, contract, EncodeQuoterMsgSender(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: quoterMsgSender: %w", err)
	}
	addr, err := DecodeQuoterMsgSender(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: quoterMsgSender: %w", err)
	}
	return addr, nil
}

// CallQuoteExactInputSingle simulates quoteExactInputSingle(params) via eth_call, as sender,
// returning the output amount for ExactAmount in.
func (q *Quoter) CallQuoteExactInputSingle(ctx context.Context, contract, sender *types.Address, params *QuoteExactSingleParams, block string) (*QuoteExactInputResult, error) {
	calldata, err := EncodeQuoteExactInputSingle(params)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactInputSingle: %w", err)
	}
	data, err := q.call(ctx, sender, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactInputSingle: %w", err)
	}
	result, err := DecodeQuoteExactInputSingle(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactInputSingle: %w", err)
	}
	return result, nil
}

// CallQuoteExactInput simulates quoteExactInput(params) via eth_call, as sender, returning the
// final output amount for ExactAmount of ExactCurrency in.
func (q *Quoter) CallQuoteExactInput(ctx context.Context, contract, sender *types.Address, params *QuoteExactParams, block string) (*QuoteExactInputResult, error) {
	calldata, err := EncodeQuoteExactInput(params)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactInput: %w", err)
	}
	data, err := q.call(ctx, sender, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactInput: %w", err)
	}
	result, err := DecodeQuoteExactInput(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactInput: %w", err)
	}
	return result, nil
}

// CallQuoteExactOutputSingle simulates quoteExactOutputSingle(params) via eth_call, as sender,
// returning the input amount required for ExactAmount out.
func (q *Quoter) CallQuoteExactOutputSingle(ctx context.Context, contract, sender *types.Address, params *QuoteExactSingleParams, block string) (*QuoteExactOutputResult, error) {
	calldata, err := EncodeQuoteExactOutputSingle(params)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactOutputSingle: %w", err)
	}
	data, err := q.call(ctx, sender, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactOutputSingle: %w", err)
	}
	result, err := DecodeQuoteExactOutputSingle(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactOutputSingle: %w", err)
	}
	return result, nil
}

// CallQuoteExactOutput simulates quoteExactOutput(params) via eth_call, as sender, returning the
// input amount required for ExactAmount of ExactCurrency out.
func (q *Quoter) CallQuoteExactOutput(ctx context.Context, contract, sender *types.Address, params *QuoteExactParams, block string) (*QuoteExactOutputResult, error) {
	calldata, err := EncodeQuoteExactOutput(params)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactOutput: %w", err)
	}
	data, err := q.call(ctx, sender, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactOutput: %w", err)
	}
	result, err := DecodeQuoteExactOutput(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactOutput: %w", err)
	}
	return result, nil
}

// call eth_calls contract with calldata, optionally simulating from a specific sender.
func (q *Quoter) call(ctx context.Context, from, contract *types.Address, calldata []byte, block string) ([]byte, error) {
	params := map[string]any{
		"to":   contract.String(),
		"data": encoding.Hex.EncodePrefixed(calldata),
	}
	if from != nil {
		params["from"] = from.String()
	}

	var raw string
	if err := q.c.Call(ctx, rpc.ETHCall("call", params, block, &raw)); err != nil {
		return nil, q.wrapRevertError(err)
	}
	return encoding.Hex.Decode(raw)
}

// wrapRevertError maps a known V4Quoter custom error in err to its sentinel. A failed quote reverts
// with UnexpectedRevertBytes(inner); a known BaseV4Quoter, IPoolManager, or v4-core library selector in inner is
// mapped too, so errors.Is matches both ErrUnexpectedRevertBytes and e.g. ErrPoolNotInitialized.
func (_ *Quoter) wrapRevertError(err error) error {
	var rpcErr *rpc.ResponseError
	if !errors.As(err, &rpcErr) || len(rpcErr.Data) == 0 {
		return err
	}

	var hexData string
	if jsonErr := json.Unmarshal(rpcErr.Data, &hexData); jsonErr != nil {
		return err
	}

	data, decErr := encoding.Hex.Decode(hexData)
	if decErr != nil || len(data) < 4 {
		return err
	}

	var sel abi.Selector
	copy(sel[:], data[:4])

	switch sel {
	case errNotEnoughLiquidity.Selector():
		return fmt.Errorf("%w: %w", ErrNotEnoughLiquidity, err)
	case errNotSelf.Selector():
		return fmt.Errorf("%w: %w", ErrNotSelf, err)
	case errUnexpectedCallSuccess.Selector():
		return fmt.Errorf("%w: %w", ErrUnexpectedCallSuccess, err)
	case errUnexpectedRevertBytes.Selector():
		vals, innerErr := errUnexpectedRevertBytes.Decode(data)
		if innerErr != nil || len(vals) != 1 {
			return fmt.Errorf("%w: %w", ErrUnexpectedRevertBytes, err)
		}
		inner, ok := vals[0].([]byte)
		if !ok || len(inner) < 4 {
			return fmt.Errorf("%w: %w", ErrUnexpectedRevertBytes, err)
		}

		var innerSel abi.Selector
		copy(innerSel[:], inner[:4])

		switch innerSel {
		case errNotEnoughLiquidity.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrNotEnoughLiquidity, err)
		case errNotSelf.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrNotSelf, err)
		case errUnexpectedCallSuccess.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrUnexpectedCallSuccess, err)
		case errCurrencyNotSettled.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrCurrencyNotSettled, err)
		case errPoolNotInitialized.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrPoolNotInitialized, err)
		case errAlreadyUnlocked.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrAlreadyUnlocked, err)
		case errManagerLocked.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrManagerLocked, err)
		case errTickSpacingTooLarge.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrTickSpacingTooLarge, err)
		case errTickSpacingTooSmall.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrTickSpacingTooSmall, err)
		case errCurrenciesOutOfOrderOrEqual.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrCurrenciesOutOfOrderOrEqual, err)
		case errUnauthorizedDynamicLPFeeUpdate.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrUnauthorizedDynamicLPFeeUpdate, err)
		case errSwapAmountCannotBeZero.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrSwapAmountCannotBeZero, err)
		case errNonzeroNativeValue.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrNonzeroNativeValue, err)
		case errMustClearExactPositiveDelta.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrMustClearExactPositiveDelta, err)
		case errSafeCastOverflow.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrSafeCastOverflow, err)
		case errPriceLimitAlreadyExceeded.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrPriceLimitAlreadyExceeded, err)
		case errPriceLimitOutOfBounds.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrPriceLimitOutOfBounds, err)
		case errInvalidFeeForExactOut.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrInvalidFeeForExactOut, err)
		case errInvalidHookResponse.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrInvalidHookResponse, err)
		case errHookDeltaExceedsSwapAmount.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrHookDeltaExceedsSwapAmount, err)
		case errLPFeeTooLarge.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrLPFeeTooLarge, err)
		case errInvalidPriceOrLiquidity.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrInvalidPriceOrLiquidity, err)
		case errInvalidPrice.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrInvalidPrice, err)
		case errPriceOverflow.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrPriceOverflow, err)
		case errSqrtPriceMathNotEnoughLiquidity.Selector():
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrNotEnoughLiquidity, err)
		case errWrappedError.Selector():
			wrappedVals, wrappedErr := errWrappedError.Decode(inner)
			if wrappedErr != nil || len(wrappedVals) != 4 {
				return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrWrappedError, err)
			}
			details, ok := wrappedVals[3].([]byte)
			if !ok || len(details) < 4 {
				return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrWrappedError, err)
			}
			var detailsSel abi.Selector
			copy(detailsSel[:], details[:4])
			if detailsSel == errHookCallFailed.Selector() {
				return fmt.Errorf("%w: %w: %w: %w", ErrUnexpectedRevertBytes, ErrWrappedError, ErrHookCallFailed, err)
			}
			return fmt.Errorf("%w: %w: %w", ErrUnexpectedRevertBytes, ErrWrappedError, err)
		default:
			return fmt.Errorf("%w: %w", ErrUnexpectedRevertBytes, err)
		}
	default:
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}
}
