package v4

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
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
