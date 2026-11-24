package v4

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

var (
	// quoteExactSingleParamsTuple mirrors IV4Quoter's QuoteExactSingleParams.
	quoteExactSingleParamsTuple = abi.Tuple(poolKeyTuple, abi.Bool, abi.Uint128, abi.Bytes)

	// quoteExactParamsTuple mirrors IV4Quoter's QuoteExactParams.
	quoteExactParamsTuple = abi.Tuple(abi.Address, abi.Slice(pathKeyTuple), abi.Uint128)

	quoteExactInputSingleFn  = abi.NewFunction("quoteExactInputSingle", abi.NewTypes(quoteExactSingleParamsTuple), abi.NewTypes(abi.Uint256, abi.Uint256))
	quoteExactInputFn        = abi.NewFunction("quoteExactInput", abi.NewTypes(quoteExactParamsTuple), abi.NewTypes(abi.Uint256, abi.Uint256))
	quoteExactOutputSingleFn = abi.NewFunction("quoteExactOutputSingle", abi.NewTypes(quoteExactSingleParamsTuple), abi.NewTypes(abi.Uint256, abi.Uint256))
	quoteExactOutputFn       = abi.NewFunction("quoteExactOutput", abi.NewTypes(quoteExactParamsTuple), abi.NewTypes(abi.Uint256, abi.Uint256))

	// Quoter-prefixed since PositionManager has its own poolManager() binding.
	quoterPoolManagerFn = abi.NewFunction("poolManager", nil, abi.NewTypes(abi.Address))
	quoterMsgSenderFn   = abi.NewFunction("msgSender", nil, abi.NewTypes(abi.Address))
)

// Custom errors from BaseV4Quoter and QuoterRevert, for decoding reverts via abi.Error.Decode.
var (
	errNotEnoughLiquidity    = abi.NewError("NotEnoughLiquidity", abi.NewTypes(abi.Bytes32))
	errNotSelf               = abi.NewError("NotSelf", abi.NewTypes())
	errUnexpectedCallSuccess = abi.NewError("UnexpectedCallSuccess", abi.NewTypes())
	errUnexpectedRevertBytes = abi.NewError("UnexpectedRevertBytes", abi.NewTypes(abi.Bytes))
)

// v4-core library errors a simulated swap can revert with, surfaced inside UnexpectedRevertBytes.
var (
	errSafeCastOverflow                = abi.NewError("SafeCastOverflow", abi.NewTypes())
	errPriceLimitAlreadyExceeded       = abi.NewError("PriceLimitAlreadyExceeded", abi.NewTypes(abi.Uint160, abi.Uint160))
	errPriceLimitOutOfBounds           = abi.NewError("PriceLimitOutOfBounds", abi.NewTypes(abi.Uint160))
	errInvalidFeeForExactOut           = abi.NewError("InvalidFeeForExactOut", abi.NewTypes())
	errInvalidHookResponse             = abi.NewError("InvalidHookResponse", abi.NewTypes())
	errHookCallFailed                  = abi.NewError("HookCallFailed", abi.NewTypes())
	errHookDeltaExceedsSwapAmount      = abi.NewError("HookDeltaExceedsSwapAmount", abi.NewTypes())
	errLPFeeTooLarge                   = abi.NewError("LPFeeTooLarge", abi.NewTypes(abi.Uint24))
	errWrappedError                    = abi.NewError("WrappedError", abi.NewTypes(abi.Address, abi.Bytes4, abi.Bytes, abi.Bytes))
	errInvalidPriceOrLiquidity         = abi.NewError("InvalidPriceOrLiquidity", abi.NewTypes())
	errInvalidPrice                    = abi.NewError("InvalidPrice", abi.NewTypes())
	errPriceOverflow                   = abi.NewError("PriceOverflow", abi.NewTypes())
	errSqrtPriceMathNotEnoughLiquidity = abi.NewError("NotEnoughLiquidity", abi.NewTypes())
)

// ErrNotEnoughLiquidity, etc. mirror the above custom errors; ErrNotEnoughLiquidity covers both NotEnoughLiquidity variants, and ErrHookCallFailed joins ErrWrappedError when its details say so.
var (
	ErrNotEnoughLiquidity         = errors.New("not enough liquidity")
	ErrNotSelf                    = errors.New("not self")
	ErrUnexpectedCallSuccess      = errors.New("unexpected call success")
	ErrUnexpectedRevertBytes      = errors.New("unexpected revert bytes")
	ErrSafeCastOverflow           = errors.New("safe cast overflow")
	ErrPriceLimitAlreadyExceeded  = errors.New("price limit already exceeded")
	ErrPriceLimitOutOfBounds      = errors.New("price limit out of bounds")
	ErrInvalidFeeForExactOut      = errors.New("invalid fee for exact out")
	ErrInvalidHookResponse        = errors.New("invalid hook response")
	ErrHookCallFailed             = errors.New("hook call failed")
	ErrHookDeltaExceedsSwapAmount = errors.New("hook delta exceeds swap amount")
	ErrLPFeeTooLarge              = errors.New("lp fee too large")
	ErrWrappedError               = errors.New("wrapped error")
	ErrInvalidPriceOrLiquidity    = errors.New("invalid price or liquidity")
	ErrInvalidPrice               = errors.New("invalid price")
	ErrPriceOverflow              = errors.New("price overflow")
)

// EncodeQuoteExactInputSingle returns the calldata for quoteExactInputSingle(params).
func EncodeQuoteExactInputSingle(p *QuoteExactSingleParams) ([]byte, error) {
	if p == nil {
		return nil, fmt.Errorf("uniswapv4: encodeQuoteExactInputSingle: params must not be nil")
	}
	if p.PoolKey == nil {
		return nil, fmt.Errorf("uniswapv4: encodeQuoteExactInputSingle: poolKey must not be nil")
	}
	data, err := quoteExactInputSingleFn.EncodeCall([]any{
		[]any{p.PoolKey.Currency0, p.PoolKey.Currency1, p.PoolKey.Fee, p.PoolKey.TickSpacing, p.PoolKey.Hooks},
		p.ZeroForOne, p.ExactAmount, p.HookData,
	})
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeQuoteExactInputSingle: %w", err)
	}
	return data, nil
}

// DecodeQuoteExactInputSingle decodes the return data of quoteExactInputSingle.
func DecodeQuoteExactInputSingle(data []byte) (*QuoteExactInputResult, error) {
	vals, err := quoteExactInputSingleFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeQuoteExactInputSingle: %w", err)
	}
	if len(vals) != 2 {
		return nil, fmt.Errorf("uniswapv4: decodeQuoteExactInputSingle: expected 2 return values, got %d", len(vals))
	}
	amountOut, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeQuoteExactInputSingle: unexpected decoded type for amountOut: %T", vals[0])
	}
	gasEstimate, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeQuoteExactInputSingle: unexpected decoded type for gasEstimate: %T", vals[1])
	}
	return NewQuoteExactInputResult(amountOut, gasEstimate), nil
}

// EncodeQuoteExactInput returns the calldata for quoteExactInput(params).
func EncodeQuoteExactInput(p *QuoteExactParams) ([]byte, error) {
	if p == nil {
		return nil, fmt.Errorf("uniswapv4: encodeQuoteExactInput: params must not be nil")
	}
	path := make([]any, len(p.Path))
	for i, k := range p.Path {
		if k == nil {
			return nil, fmt.Errorf("uniswapv4: encodeQuoteExactInput: path[%d] must not be nil", i)
		}
		path[i] = []any{k.IntermediateCurrency, k.Fee, k.TickSpacing, k.Hooks, k.HookData}
	}
	data, err := quoteExactInputFn.EncodeCall([]any{p.ExactCurrency, path, p.ExactAmount})
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeQuoteExactInput: %w", err)
	}
	return data, nil
}

// DecodeQuoteExactInput decodes the return data of quoteExactInput.
func DecodeQuoteExactInput(data []byte) (*QuoteExactInputResult, error) {
	vals, err := quoteExactInputFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeQuoteExactInput: %w", err)
	}
	if len(vals) != 2 {
		return nil, fmt.Errorf("uniswapv4: decodeQuoteExactInput: expected 2 return values, got %d", len(vals))
	}
	amountOut, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeQuoteExactInput: unexpected decoded type for amountOut: %T", vals[0])
	}
	gasEstimate, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeQuoteExactInput: unexpected decoded type for gasEstimate: %T", vals[1])
	}
	return NewQuoteExactInputResult(amountOut, gasEstimate), nil
}

// EncodeQuoteExactOutputSingle returns the calldata for quoteExactOutputSingle(params).
func EncodeQuoteExactOutputSingle(p *QuoteExactSingleParams) ([]byte, error) {
	if p == nil {
		return nil, fmt.Errorf("uniswapv4: encodeQuoteExactOutputSingle: params must not be nil")
	}
	if p.PoolKey == nil {
		return nil, fmt.Errorf("uniswapv4: encodeQuoteExactOutputSingle: poolKey must not be nil")
	}
	data, err := quoteExactOutputSingleFn.EncodeCall([]any{
		[]any{p.PoolKey.Currency0, p.PoolKey.Currency1, p.PoolKey.Fee, p.PoolKey.TickSpacing, p.PoolKey.Hooks},
		p.ZeroForOne, p.ExactAmount, p.HookData,
	})
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeQuoteExactOutputSingle: %w", err)
	}
	return data, nil
}

// DecodeQuoteExactOutputSingle decodes the return data of quoteExactOutputSingle.
func DecodeQuoteExactOutputSingle(data []byte) (*QuoteExactOutputResult, error) {
	vals, err := quoteExactOutputSingleFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeQuoteExactOutputSingle: %w", err)
	}
	if len(vals) != 2 {
		return nil, fmt.Errorf("uniswapv4: decodeQuoteExactOutputSingle: expected 2 return values, got %d", len(vals))
	}
	amountIn, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeQuoteExactOutputSingle: unexpected decoded type for amountIn: %T", vals[0])
	}
	gasEstimate, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeQuoteExactOutputSingle: unexpected decoded type for gasEstimate: %T", vals[1])
	}
	return NewQuoteExactOutputResult(amountIn, gasEstimate), nil
}

// EncodeQuoteExactOutput returns the calldata for quoteExactOutput(params); ExactCurrency is the output and Path[0] holds the swap's input.
func EncodeQuoteExactOutput(p *QuoteExactParams) ([]byte, error) {
	if p == nil {
		return nil, fmt.Errorf("uniswapv4: encodeQuoteExactOutput: params must not be nil")
	}
	path := make([]any, len(p.Path))
	for i, k := range p.Path {
		if k == nil {
			return nil, fmt.Errorf("uniswapv4: encodeQuoteExactOutput: path[%d] must not be nil", i)
		}
		path[i] = []any{k.IntermediateCurrency, k.Fee, k.TickSpacing, k.Hooks, k.HookData}
	}
	data, err := quoteExactOutputFn.EncodeCall([]any{p.ExactCurrency, path, p.ExactAmount})
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeQuoteExactOutput: %w", err)
	}
	return data, nil
}

// DecodeQuoteExactOutput decodes the return data of quoteExactOutput.
func DecodeQuoteExactOutput(data []byte) (*QuoteExactOutputResult, error) {
	vals, err := quoteExactOutputFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeQuoteExactOutput: %w", err)
	}
	if len(vals) != 2 {
		return nil, fmt.Errorf("uniswapv4: decodeQuoteExactOutput: expected 2 return values, got %d", len(vals))
	}
	amountIn, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeQuoteExactOutput: unexpected decoded type for amountIn: %T", vals[0])
	}
	gasEstimate, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeQuoteExactOutput: unexpected decoded type for gasEstimate: %T", vals[1])
	}
	return NewQuoteExactOutputResult(amountIn, gasEstimate), nil
}

// EncodeQuoterPoolManager returns the calldata for poolManager().
func EncodeQuoterPoolManager() []byte {
	return quoterPoolManagerFn.SelectorBytes()
}

// DecodeQuoterPoolManager decodes the return data of poolManager.
func DecodeQuoterPoolManager(data []byte) (*types.Address, error) {
	addr, err := quoterPoolManagerFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeQuoterPoolManager: %w", err)
	}
	return addr, nil
}

// EncodeQuoterMsgSender returns the calldata for msgSender().
func EncodeQuoterMsgSender() []byte {
	return quoterMsgSenderFn.SelectorBytes()
}

// DecodeQuoterMsgSender decodes the return data of msgSender, always the zero address outside a quote.
func DecodeQuoterMsgSender(data []byte) (*types.Address, error) {
	addr, err := quoterMsgSenderFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeQuoterMsgSender: %w", err)
	}
	return addr, nil
}

// mapQuoterRevertError maps a V4Quoter custom error in err to its sentinel, including the selector inside UnexpectedRevertBytes; other top-level reverts fall through to PoolManager's mapper.
func mapQuoterRevertError(err error) error {
	if rpcErr, ok := errors.AsType[*rpc.ResponseError](err); ok && rpcErr == nil {
		return nil
	}

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
		return mapPoolManagerRevertError(err) // HACK
	}
}
