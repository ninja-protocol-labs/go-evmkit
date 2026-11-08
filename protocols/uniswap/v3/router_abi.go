package v3

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Tuple types for IV3SwapRouter's struct parameters (SwapRouter02 -- no deadline field;
// deadline protection moved to the multicall(deadline, bytes[]) overload instead).
var (
	exactInputSingleParamsT  = abi.Tuple(abi.Address, abi.Address, abi.Uint24, abi.Address, abi.Uint256, abi.Uint256, abi.Uint160)
	exactInputParamsT        = abi.Tuple(abi.Bytes, abi.Address, abi.Uint256, abi.Uint256)
	exactOutputSingleParamsT = abi.Tuple(abi.Address, abi.Address, abi.Uint24, abi.Address, abi.Uint256, abi.Uint256, abi.Uint160)
	exactOutputParamsT       = abi.Tuple(abi.Bytes, abi.Address, abi.Uint256, abi.Uint256)
)

// Function signatures from IV3SwapRouter and SwapRouter02's base contracts (PeripheryImmutableState,
// PeripheryPaymentsWithFee, Multicall, SelfPermit), for encoding calls and decoding results.
// Router-prefixed where the same ABI function also exists on NFPM (a different contract, so a
// different Go binding, even though the calldata happens to be byte-identical).
var (
	routerFactoryFn = abi.NewFunction("factory", nil, abi.NewTypes(abi.Address))
	routerWETH9Fn   = abi.NewFunction("WETH9", nil, abi.NewTypes(abi.Address))

	exactInputSingleFn  = abi.NewFunction("exactInputSingle", abi.NewTypes(exactInputSingleParamsT), abi.NewTypes(abi.Uint256))
	exactInputFn        = abi.NewFunction("exactInput", abi.NewTypes(exactInputParamsT), abi.NewTypes(abi.Uint256))
	exactOutputSingleFn = abi.NewFunction("exactOutputSingle", abi.NewTypes(exactOutputSingleParamsT), abi.NewTypes(abi.Uint256))
	exactOutputFn       = abi.NewFunction("exactOutput", abi.NewTypes(exactOutputParamsT), abi.NewTypes(abi.Uint256))

	routerUnwrapWETH9Fn  = abi.NewFunction("unwrapWETH9", abi.NewTypes(abi.Uint256, abi.Address), nil)
	unwrapWETH9WithFeeFn = abi.NewFunction("unwrapWETH9WithFee", abi.NewTypes(abi.Uint256, abi.Address, abi.Uint256, abi.Address), nil)
	routerSweepTokenFn   = abi.NewFunction("sweepToken", abi.NewTypes(abi.Address, abi.Uint256, abi.Address), nil)
	sweepTokenWithFeeFn  = abi.NewFunction("sweepTokenWithFee", abi.NewTypes(abi.Address, abi.Uint256, abi.Address, abi.Uint256, abi.Address), nil)
	routerRefundETHFn    = abi.NewFunction("refundETH", nil, nil)

	routerMulticallFn = abi.NewFunction("multicall", abi.NewTypes(abi.Slice(abi.Bytes)), abi.NewTypes(abi.Slice(abi.Bytes)))

	selfPermitFn                   = abi.NewFunction("selfPermit", abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256, abi.Uint8, abi.Bytes32, abi.Bytes32), nil)
	selfPermitIfNecessaryFn        = abi.NewFunction("selfPermitIfNecessary", abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256, abi.Uint8, abi.Bytes32, abi.Bytes32), nil)
	selfPermitAllowedFn            = abi.NewFunction("selfPermitAllowed", abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256, abi.Uint8, abi.Bytes32, abi.Bytes32), nil)
	selfPermitAllowedIfNecessaryFn = abi.NewFunction("selfPermitAllowedIfNecessary", abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256, abi.Uint8, abi.Bytes32, abi.Bytes32), nil)
)

// Revert reasons reachable from SwapRouter's own functions and the base contracts it inherits
// (PeripheryValidation, PeripheryPayments, PeripheryPaymentsWithFee, PeripheryTransferHelper).
// reasonTransactionTooOld/reasonInsufficientWETH9/reasonInsufficientToken/reasonNotWETH9/
// reasonTransferFromFailed/reasonTokenTransferFailed/reasonApproveFailed/reasonTransferETHFailed
// (and their Err sentinels) are declared once in nfpm_abi.go, shared with NFPM's identical base
// contracts.
const (
	reasonTooLittleReceived = "Too little received"
	reasonTooMuchRequested  = "Too much requested"
)

var (
	ErrTooLittleReceived = errors.New("uniswapv3: too little received")
	ErrTooMuchRequested  = errors.New("uniswapv3: too much requested")
)

// EncodeRouterFactory returns the calldata for factory().
func EncodeRouterFactory() []byte {
	return routerFactoryFn.SelectorBytes()
}

// DecodeRouterFactory decodes the return data of factory.
func DecodeRouterFactory(data []byte) (*types.Address, error) {
	addr, err := routerFactoryFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeRouterFactory: %w", err)
	}
	return addr, nil
}

// EncodeRouterWETH9 returns the calldata for WETH9().
func EncodeRouterWETH9() []byte {
	return routerWETH9Fn.SelectorBytes()
}

// DecodeRouterWETH9 decodes the return data of WETH9.
func DecodeRouterWETH9(data []byte) (*types.Address, error) {
	addr, err := routerWETH9Fn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeRouterWETH9: %w", err)
	}
	return addr, nil
}

// EncodeExactInputSingle returns the calldata for exactInputSingle(params).
func EncodeExactInputSingle(
	tokenIn, tokenOut *types.Address,
	fee *big.Int,
	recipient *types.Address,
	amountIn, amountOutMinimum, sqrtPriceLimitX96 *big.Int,
) ([]byte, error) {
	data, err := exactInputSingleFn.EncodeCall([]any{tokenIn, tokenOut, fee, recipient, amountIn, amountOutMinimum, sqrtPriceLimitX96})
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeExactInputSingle: %w", err)
	}
	return data, nil
}

// DecodeExactInputSingle decodes the return data of exactInputSingle -- amountOut.
func DecodeExactInputSingle(data []byte) (*big.Int, error) {
	n, err := exactInputSingleFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeExactInputSingle: %w", err)
	}
	return n, nil
}

// EncodeExactInput returns the calldata for exactInput(params).
func EncodeExactInput(
	path []byte,
	recipient *types.Address,
	amountIn, amountOutMinimum *big.Int,
) ([]byte, error) {
	data, err := exactInputFn.EncodeCall([]any{path, recipient, amountIn, amountOutMinimum})
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeExactInput: %w", err)
	}
	return data, nil
}

// DecodeExactInput decodes the return data of exactInput -- amountOut.
func DecodeExactInput(data []byte) (*big.Int, error) {
	n, err := exactInputFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeExactInput: %w", err)
	}
	return n, nil
}

// EncodeExactOutputSingle returns the calldata for exactOutputSingle(params).
func EncodeExactOutputSingle(
	tokenIn, tokenOut *types.Address,
	fee *big.Int,
	recipient *types.Address,
	amountOut, amountInMaximum, sqrtPriceLimitX96 *big.Int,
) ([]byte, error) {
	data, err := exactOutputSingleFn.EncodeCall([]any{tokenIn, tokenOut, fee, recipient, amountOut, amountInMaximum, sqrtPriceLimitX96})
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeExactOutputSingle: %w", err)
	}
	return data, nil
}

// DecodeExactOutputSingle decodes the return data of exactOutputSingle -- amountIn.
func DecodeExactOutputSingle(data []byte) (*big.Int, error) {
	n, err := exactOutputSingleFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeExactOutputSingle: %w", err)
	}
	return n, nil
}

// EncodeExactOutput returns the calldata for exactOutput(params).
func EncodeExactOutput(
	path []byte,
	recipient *types.Address,
	amountOut, amountInMaximum *big.Int,
) ([]byte, error) {
	data, err := exactOutputFn.EncodeCall([]any{path, recipient, amountOut, amountInMaximum})
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeExactOutput: %w", err)
	}
	return data, nil
}

// DecodeExactOutput decodes the return data of exactOutput -- amountIn.
func DecodeExactOutput(data []byte) (*big.Int, error) {
	n, err := exactOutputFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeExactOutput: %w", err)
	}
	return n, nil
}

// EncodeRouterUnwrapWETH9 returns the calldata for unwrapWETH9(amountMinimum, recipient).
func EncodeRouterUnwrapWETH9(amountMinimum *big.Int, recipient *types.Address) ([]byte, error) {
	data, err := routerUnwrapWETH9Fn.EncodeCall(amountMinimum, recipient)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeRouterUnwrapWETH9: %w", err)
	}
	return data, nil
}

// EncodeUnwrapWETH9WithFee returns the calldata for unwrapWETH9WithFee(amountMinimum, recipient, feeBips, feeRecipient).
func EncodeUnwrapWETH9WithFee(
	amountMinimum *big.Int,
	recipient *types.Address,
	feeBips *big.Int,
	feeRecipient *types.Address,
) ([]byte, error) {
	data, err := unwrapWETH9WithFeeFn.EncodeCall(amountMinimum, recipient, feeBips, feeRecipient)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeUnwrapWETH9WithFee: %w", err)
	}
	return data, nil
}

// EncodeRouterSweepToken returns the calldata for sweepToken(token, amountMinimum, recipient).
func EncodeRouterSweepToken(
	token *types.Address,
	amountMinimum *big.Int,
	recipient *types.Address,
) ([]byte, error) {
	data, err := routerSweepTokenFn.EncodeCall(token, amountMinimum, recipient)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeRouterSweepToken: %w", err)
	}
	return data, nil
}

// EncodeSweepTokenWithFee returns the calldata for sweepTokenWithFee(token, amountMinimum, recipient, feeBips, feeRecipient).
func EncodeSweepTokenWithFee(
	token *types.Address,
	amountMinimum *big.Int,
	recipient *types.Address,
	feeBips *big.Int,
	feeRecipient *types.Address,
) ([]byte, error) {
	data, err := sweepTokenWithFeeFn.EncodeCall(token, amountMinimum, recipient, feeBips, feeRecipient)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeSweepTokenWithFee: %w", err)
	}
	return data, nil
}

// EncodeRouterRefundETH returns the calldata for refundETH().
func EncodeRouterRefundETH() []byte {
	return routerRefundETHFn.SelectorBytes()
}

// EncodeRouterMulticall returns the calldata for multicall(data).
func EncodeRouterMulticall(data [][]byte) ([]byte, error) {
	calldata, err := routerMulticallFn.EncodeCall(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeRouterMulticall: %w", err)
	}
	return calldata, nil
}

// DecodeRouterMulticall decodes the return data of multicall -- each sub-call's own return data.
func DecodeRouterMulticall(data []byte) ([][]byte, error) {
	vals, err := routerMulticallFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeRouterMulticall: %w", err)
	}
	if len(vals) != 1 {
		return nil, fmt.Errorf("uniswapv3: decodeRouterMulticall: expected 1 return value, got %d", len(vals))
	}
	elems, ok := vals[0].([]any)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeRouterMulticall: unexpected decoded type for results: %T", vals[0])
	}
	results := make([][]byte, len(elems))
	for i, e := range elems {
		b, ok := e.([]byte)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: decodeRouterMulticall: unexpected decoded type for results[%d]: %T", i, e)
		}
		results[i] = b
	}
	return results, nil
}

// EncodeSelfPermit returns the calldata for selfPermit(token, value, deadline, v, r, s). sig's
// recovery id is encoded in its legacy (27/28) form, as EIP-2612's permit expects.
func EncodeSelfPermit(
	token *types.Address,
	value, deadline *big.Int,
	sig *types.Signature,
) ([]byte, error) {
	legacy := sig.LegacyBytes()
	data, err := selfPermitFn.EncodeCall(token, value, deadline, legacy[64], legacy[:32], legacy[32:64])
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeSelfPermit: %w", err)
	}
	return data, nil
}

// EncodeSelfPermitIfNecessary returns the calldata for selfPermitIfNecessary(token, value, deadline, v, r, s).
func EncodeSelfPermitIfNecessary(
	token *types.Address,
	value, deadline *big.Int,
	sig *types.Signature,
) ([]byte, error) {
	legacy := sig.LegacyBytes()
	data, err := selfPermitIfNecessaryFn.EncodeCall(token, value, deadline, legacy[64], legacy[:32], legacy[32:64])
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeSelfPermitIfNecessary: %w", err)
	}
	return data, nil
}

// EncodeSelfPermitAllowed returns the calldata for selfPermitAllowed(token, nonce, expiry, v, r, s).
func EncodeSelfPermitAllowed(
	token *types.Address,
	nonce, expiry *big.Int,
	sig *types.Signature,
) ([]byte, error) {
	legacy := sig.LegacyBytes()
	data, err := selfPermitAllowedFn.EncodeCall(token, nonce, expiry, legacy[64], legacy[:32], legacy[32:64])
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeSelfPermitAllowed: %w", err)
	}
	return data, nil
}

// EncodeSelfPermitAllowedIfNecessary returns the calldata for selfPermitAllowedIfNecessary(token, nonce, expiry, v, r, s).
func EncodeSelfPermitAllowedIfNecessary(
	token *types.Address,
	nonce, expiry *big.Int,
	sig *types.Signature,
) ([]byte, error) {
	legacy := sig.LegacyBytes()
	data, err := selfPermitAllowedIfNecessaryFn.EncodeCall(token, nonce, expiry, legacy[64], legacy[:32], legacy[32:64])
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeSelfPermitAllowedIfNecessary: %w", err)
	}
	return data, nil
}
