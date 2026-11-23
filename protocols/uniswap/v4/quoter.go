package v4

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// IQuoterImmutables mirrors IImmutableState plus IMsgSender, the only state V4Quoter exposes.
type IQuoterImmutables interface {
	PoolManager(ctx context.Context, opts *bind.CallOptions, req *PoolManagerRequest) (*types.Address, error)
	MsgSender(ctx context.Context, opts *bind.CallOptions, req *MsgSenderRequest) (*types.Address, error)
}

// IQuoterQuotes mirrors IV4Quoter's quoting functions; they only simulate, so only eth_call-backed Call* methods exist.
type IQuoterQuotes interface {
	CallQuoteExactInputSingle(ctx context.Context, opts *bind.CallOptions, req *QuoteExactInputSingleRequest) (*QuoteExactInputResult, error)
	CallQuoteExactInput(ctx context.Context, opts *bind.CallOptions, req *QuoteExactInputRequest) (*QuoteExactInputResult, error)
	CallQuoteExactOutputSingle(ctx context.Context, opts *bind.CallOptions, req *QuoteExactOutputSingleRequest) (*QuoteExactOutputResult, error)
	CallQuoteExactOutput(ctx context.Context, opts *bind.CallOptions, req *QuoteExactOutputRequest) (*QuoteExactOutputResult, error)
}

// IQuoter mirrors IV4Quoter.
type IQuoter interface {
	IQuoterImmutables
	IQuoterQuotes
}

var _ IQuoter = (*Quoter)(nil)

// Quoter calls any V4Quoter deployment over a Client; it holds no contract address.
type Quoter struct {
	c rpc.Client
}

// NewQuoter returns a Quoter calling V4Quoter deployments via cli.
func NewQuoter(cli rpc.Client) *Quoter {
	return &Quoter{
		c: cli,
	}
}

// PoolManager returns the PoolManager req.Contract quotes against.
func (q *Quoter) PoolManager(ctx context.Context, opts *bind.CallOptions, req *PoolManagerRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: quoterPoolManager: %w", err)
	}

	data, err := bind.Call(ctx, q.c, opts, req.Contract, EncodeQuoterPoolManager())
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: quoterPoolManager: %w", q.wrapRevertError(err))
	}

	out, err := DecodeQuoterPoolManager(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: quoterPoolManager: %w", err)
	}

	return out, nil
}

// MsgSender returns req.Contract's current msgSender, the zero address outside a quote.
func (q *Quoter) MsgSender(ctx context.Context, opts *bind.CallOptions, req *MsgSenderRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: quoterMsgSender: %w", err)
	}

	data, err := bind.Call(ctx, q.c, opts, req.Contract, EncodeQuoterMsgSender())
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: quoterMsgSender: %w", q.wrapRevertError(err))
	}

	out, err := DecodeQuoterMsgSender(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: quoterMsgSender: %w", err)
	}

	return out, nil
}

// CallQuoteExactInputSingle runs quoteExactInputSingle via eth_call as opts.From, returning the output for req.Params.ExactAmount in or the revert.
func (q *Quoter) CallQuoteExactInputSingle(ctx context.Context, opts *bind.CallOptions, req *QuoteExactInputSingleRequest) (*QuoteExactInputResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactInputSingle: %w", err)
	}

	calldata, err := EncodeQuoteExactInputSingle(req.Params)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactInputSingle: %w", err)
	}

	data, err := bind.Call(ctx, q.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactInputSingle: %w", q.wrapRevertError(err))
	}

	out, err := DecodeQuoteExactInputSingle(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactInputSingle: %w", err)
	}

	return out, nil
}

// CallQuoteExactInput runs quoteExactInput via eth_call as opts.From, returning the final output for req.Params.ExactAmount of ExactCurrency in or the revert.
func (q *Quoter) CallQuoteExactInput(ctx context.Context, opts *bind.CallOptions, req *QuoteExactInputRequest) (*QuoteExactInputResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactInput: %w", err)
	}

	calldata, err := EncodeQuoteExactInput(req.Params)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactInput: %w", err)
	}

	data, err := bind.Call(ctx, q.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactInput: %w", q.wrapRevertError(err))
	}

	out, err := DecodeQuoteExactInput(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactInput: %w", err)
	}

	return out, nil
}

// CallQuoteExactOutputSingle runs quoteExactOutputSingle via eth_call as opts.From, returning the input required for req.Params.ExactAmount out or the revert.
func (q *Quoter) CallQuoteExactOutputSingle(ctx context.Context, opts *bind.CallOptions, req *QuoteExactOutputSingleRequest) (*QuoteExactOutputResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactOutputSingle: %w", err)
	}

	calldata, err := EncodeQuoteExactOutputSingle(req.Params)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactOutputSingle: %w", err)
	}

	data, err := bind.Call(ctx, q.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactOutputSingle: %w", q.wrapRevertError(err))
	}

	out, err := DecodeQuoteExactOutputSingle(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactOutputSingle: %w", err)
	}

	return out, nil
}

// CallQuoteExactOutput runs quoteExactOutput via eth_call as opts.From, returning the input required for req.Params.ExactAmount of ExactCurrency out or the revert.
func (q *Quoter) CallQuoteExactOutput(ctx context.Context, opts *bind.CallOptions, req *QuoteExactOutputRequest) (*QuoteExactOutputResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactOutput: %w", err)
	}

	calldata, err := EncodeQuoteExactOutput(req.Params)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactOutput: %w", err)
	}

	data, err := bind.Call(ctx, q.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactOutput: %w", q.wrapRevertError(err))
	}

	out, err := DecodeQuoteExactOutput(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callQuoteExactOutput: %w", err)
	}

	return out, nil
}

// wrapRevertError maps a V4Quoter custom error in err to its sentinel, including the selector inside UnexpectedRevertBytes; other top-level reverts fall through to PoolManager's mapper.
func (_ *Quoter) wrapRevertError(err error) error {
	rpcErr, ok := errors.AsType[*rpc.ResponseError](err)
	if !ok || len(rpcErr.Data) == 0 {
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
		return (*PoolManager)(nil).wrapRevertError(err) // HACK
	}
}
