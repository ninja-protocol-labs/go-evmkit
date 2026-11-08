package v3

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Tuple types for IQuoterV2's struct parameters.
var (
	quoteExactInputSingleParamsT  = abi.Tuple(abi.Address, abi.Address, abi.Uint256, abi.Uint24, abi.Uint160)
	quoteExactOutputSingleParamsT = abi.Tuple(abi.Address, abi.Address, abi.Uint256, abi.Uint24, abi.Uint160)
)

// Function signatures from IQuoterV2 and the PeripheryImmutableState it inherits.
// Quoter-prefixed where the same ABI function also exists on Router/NFPM (different contracts,
// so different Go bindings, even though the calldata happens to be byte-identical).
var (
	quoterFactoryFn = abi.NewFunction("factory", nil, abi.NewTypes(abi.Address))
	quoterWETH9Fn   = abi.NewFunction("WETH9", nil, abi.NewTypes(abi.Address))

	quoteExactInputSingleFn = abi.NewFunction(
		"quoteExactInputSingle",
		abi.NewTypes(quoteExactInputSingleParamsT),
		abi.NewTypes(abi.Uint256, abi.Uint160, abi.Uint32, abi.Uint256),
	)
	quoteExactInputFn = abi.NewFunction(
		"quoteExactInput",
		abi.NewTypes(abi.Bytes, abi.Uint256),
		abi.NewTypes(abi.Uint256, abi.Slice(abi.Uint160), abi.Slice(abi.Uint32), abi.Uint256),
	)
	quoteExactOutputSingleFn = abi.NewFunction(
		"quoteExactOutputSingle",
		abi.NewTypes(quoteExactOutputSingleParamsT),
		abi.NewTypes(abi.Uint256, abi.Uint160, abi.Uint32, abi.Uint256),
	)
	quoteExactOutputFn = abi.NewFunction(
		"quoteExactOutput",
		abi.NewTypes(abi.Bytes, abi.Uint256),
		abi.NewTypes(abi.Uint256, abi.Slice(abi.Uint160), abi.Slice(abi.Uint32), abi.Uint256),
	)
)

// EncodeQuoterFactory returns the calldata for factory().
func EncodeQuoterFactory() []byte {
	return quoterFactoryFn.SelectorBytes()
}

// DecodeQuoterFactory decodes the return data of factory.
func DecodeQuoterFactory(data []byte) (*types.Address, error) {
	addr, err := quoterFactoryFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeQuoterFactory: %w", err)
	}
	return addr, nil
}

// EncodeQuoterWETH9 returns the calldata for WETH9().
func EncodeQuoterWETH9() []byte {
	return quoterWETH9Fn.SelectorBytes()
}

// DecodeQuoterWETH9 decodes the return data of WETH9.
func DecodeQuoterWETH9(data []byte) (*types.Address, error) {
	addr, err := quoterWETH9Fn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeQuoterWETH9: %w", err)
	}
	return addr, nil
}

// EncodeQuoteExactInputSingle returns the calldata for quoteExactInputSingle(params).
func EncodeQuoteExactInputSingle(
	tokenIn, tokenOut *types.Address,
	amountIn, fee, sqrtPriceLimitX96 *big.Int,
) ([]byte, error) {
	data, err := quoteExactInputSingleFn.EncodeCall([]any{tokenIn, tokenOut, amountIn, fee, sqrtPriceLimitX96})
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeQuoteExactInputSingle: %w", err)
	}
	return data, nil
}

// DecodeQuoteExactInputSingle decodes the return data of quoteExactInputSingle.
func DecodeQuoteExactInputSingle(data []byte) (*QuoteExactInputSingleResult, error) {
	vals, err := quoteExactInputSingleFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeQuoteExactInputSingle: %w", err)
	}
	if len(vals) != 4 {
		return nil, fmt.Errorf("uniswapv3: decodeQuoteExactInputSingle: expected 4 return values, got %d", len(vals))
	}
	amountOut, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeQuoteExactInputSingle: unexpected decoded type for amountOut: %T", vals[0])
	}
	sqrtPriceX96After, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeQuoteExactInputSingle: unexpected decoded type for sqrtPriceX96After: %T", vals[1])
	}
	initializedTicksCrossedRaw, ok := vals[2].(uint32)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeQuoteExactInputSingle: unexpected decoded type for initializedTicksCrossed: %T", vals[2])
	}
	initializedTicksCrossed := new(big.Int).SetUint64(uint64(initializedTicksCrossedRaw))
	gasEstimate, ok := vals[3].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeQuoteExactInputSingle: unexpected decoded type for gasEstimate: %T", vals[3])
	}
	return NewQuoteExactInputSingleResult(amountOut, sqrtPriceX96After, initializedTicksCrossed, gasEstimate), nil
}

// EncodeQuoteExactInput returns the calldata for quoteExactInput(path, amountIn).
func EncodeQuoteExactInput(path []byte, amountIn *big.Int) ([]byte, error) {
	data, err := quoteExactInputFn.EncodeCall(path, amountIn)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeQuoteExactInput: %w", err)
	}
	return data, nil
}

// DecodeQuoteExactInput decodes the return data of quoteExactInput.
func DecodeQuoteExactInput(data []byte) (*QuoteExactInputResult, error) {
	vals, err := quoteExactInputFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeQuoteExactInput: %w", err)
	}
	if len(vals) != 4 {
		return nil, fmt.Errorf("uniswapv3: decodeQuoteExactInput: expected 4 return values, got %d", len(vals))
	}
	amountOut, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeQuoteExactInput: unexpected decoded type for amountOut: %T", vals[0])
	}
	sqrtPriceX96AfterRaw, ok := vals[1].([]any)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeQuoteExactInput: unexpected decoded type for sqrtPriceX96AfterList: %T", vals[1])
	}
	sqrtPriceX96AfterList := make([]*big.Int, len(sqrtPriceX96AfterRaw))
	for i, v := range sqrtPriceX96AfterRaw {
		n, ok := v.(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: decodeQuoteExactInput: unexpected decoded type for sqrtPriceX96AfterList[%d]: %T", i, v)
		}
		sqrtPriceX96AfterList[i] = n
	}
	initializedTicksCrossedRaw, ok := vals[2].([]any)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeQuoteExactInput: unexpected decoded type for initializedTicksCrossedList: %T", vals[2])
	}
	initializedTicksCrossedList := make([]*big.Int, len(initializedTicksCrossedRaw))
	for i, v := range initializedTicksCrossedRaw {
		n, ok := v.(uint32)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: decodeQuoteExactInput: unexpected decoded type for initializedTicksCrossedList[%d]: %T", i, v)
		}
		initializedTicksCrossedList[i] = new(big.Int).SetUint64(uint64(n))
	}
	gasEstimate, ok := vals[3].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeQuoteExactInput: unexpected decoded type for gasEstimate: %T", vals[3])
	}
	return NewQuoteExactInputResult(amountOut, sqrtPriceX96AfterList, initializedTicksCrossedList, gasEstimate), nil
}

// EncodeQuoteExactOutputSingle returns the calldata for quoteExactOutputSingle(params).
func EncodeQuoteExactOutputSingle(
	tokenIn, tokenOut *types.Address,
	amount, fee, sqrtPriceLimitX96 *big.Int,
) ([]byte, error) {
	data, err := quoteExactOutputSingleFn.EncodeCall([]any{tokenIn, tokenOut, amount, fee, sqrtPriceLimitX96})
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeQuoteExactOutputSingle: %w", err)
	}
	return data, nil
}

// DecodeQuoteExactOutputSingle decodes the return data of quoteExactOutputSingle.
func DecodeQuoteExactOutputSingle(data []byte) (*QuoteExactOutputSingleResult, error) {
	vals, err := quoteExactOutputSingleFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeQuoteExactOutputSingle: %w", err)
	}
	if len(vals) != 4 {
		return nil, fmt.Errorf("uniswapv3: decodeQuoteExactOutputSingle: expected 4 return values, got %d", len(vals))
	}
	amountIn, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeQuoteExactOutputSingle: unexpected decoded type for amountIn: %T", vals[0])
	}
	sqrtPriceX96After, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeQuoteExactOutputSingle: unexpected decoded type for sqrtPriceX96After: %T", vals[1])
	}
	initializedTicksCrossedRaw, ok := vals[2].(uint32)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeQuoteExactOutputSingle: unexpected decoded type for initializedTicksCrossed: %T", vals[2])
	}
	initializedTicksCrossed := new(big.Int).SetUint64(uint64(initializedTicksCrossedRaw))
	gasEstimate, ok := vals[3].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeQuoteExactOutputSingle: unexpected decoded type for gasEstimate: %T", vals[3])
	}
	return NewQuoteExactOutputSingleResult(amountIn, sqrtPriceX96After, initializedTicksCrossed, gasEstimate), nil
}

// EncodeQuoteExactOutput returns the calldata for quoteExactOutput(path, amountOut).
func EncodeQuoteExactOutput(path []byte, amountOut *big.Int) ([]byte, error) {
	data, err := quoteExactOutputFn.EncodeCall(path, amountOut)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeQuoteExactOutput: %w", err)
	}
	return data, nil
}

// DecodeQuoteExactOutput decodes the return data of quoteExactOutput.
func DecodeQuoteExactOutput(data []byte) (*QuoteExactOutputResult, error) {
	vals, err := quoteExactOutputFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeQuoteExactOutput: %w", err)
	}
	if len(vals) != 4 {
		return nil, fmt.Errorf("uniswapv3: decodeQuoteExactOutput: expected 4 return values, got %d", len(vals))
	}
	amountIn, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeQuoteExactOutput: unexpected decoded type for amountIn: %T", vals[0])
	}
	sqrtPriceX96AfterRaw, ok := vals[1].([]any)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeQuoteExactOutput: unexpected decoded type for sqrtPriceX96AfterList: %T", vals[1])
	}
	sqrtPriceX96AfterList := make([]*big.Int, len(sqrtPriceX96AfterRaw))
	for i, v := range sqrtPriceX96AfterRaw {
		n, ok := v.(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: decodeQuoteExactOutput: unexpected decoded type for sqrtPriceX96AfterList[%d]: %T", i, v)
		}
		sqrtPriceX96AfterList[i] = n
	}
	initializedTicksCrossedRaw, ok := vals[2].([]any)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeQuoteExactOutput: unexpected decoded type for initializedTicksCrossedList: %T", vals[2])
	}
	initializedTicksCrossedList := make([]*big.Int, len(initializedTicksCrossedRaw))
	for i, v := range initializedTicksCrossedRaw {
		n, ok := v.(uint32)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: decodeQuoteExactOutput: unexpected decoded type for initializedTicksCrossedList[%d]: %T", i, v)
		}
		initializedTicksCrossedList[i] = new(big.Int).SetUint64(uint64(n))
	}
	gasEstimate, ok := vals[3].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeQuoteExactOutput: unexpected decoded type for gasEstimate: %T", vals[3])
	}
	return NewQuoteExactOutputResult(amountIn, sqrtPriceX96AfterList, initializedTicksCrossedList, gasEstimate), nil
}
