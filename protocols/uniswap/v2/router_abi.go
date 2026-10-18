package v2

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Function signatures from IUniswapV2Router01/IUniswapV2Router02.
var (
	routerFactoryFn = abi.NewFunction(
		"factory",
		nil,
		abi.NewTypes(abi.Address),
	)
	wethFn = abi.NewFunction(
		"WETH",
		nil,
		abi.NewTypes(abi.Address),
	)
	quoteFn = abi.NewFunction(
		"quote",
		abi.NewTypes(abi.Uint256, abi.Uint256, abi.Uint256),
		abi.NewTypes(abi.Uint256),
	)
	getAmountOutFn = abi.NewFunction(
		"getAmountOut",
		abi.NewTypes(abi.Uint256, abi.Uint256, abi.Uint256),
		abi.NewTypes(abi.Uint256),
	)
	getAmountInFn = abi.NewFunction(
		"getAmountIn",
		abi.NewTypes(abi.Uint256, abi.Uint256, abi.Uint256),
		abi.NewTypes(abi.Uint256),
	)
	getAmountsOutFn = abi.NewFunction(
		"getAmountsOut",
		abi.NewTypes(abi.Uint256, abi.Slice(abi.Address)),
		abi.NewTypes(abi.Slice(abi.Uint256)),
	)
	getAmountsInFn = abi.NewFunction(
		"getAmountsIn",
		abi.NewTypes(abi.Uint256, abi.Slice(abi.Address)),
		abi.NewTypes(abi.Slice(abi.Uint256)),
	)
	addLiquidityFn = abi.NewFunction(
		"addLiquidity",
		abi.NewTypes(abi.Address, abi.Address, abi.Uint256, abi.Uint256, abi.Uint256, abi.Uint256, abi.Address, abi.Uint256),
		abi.NewTypes(abi.Uint256, abi.Uint256, abi.Uint256),
	)
	addLiquidityETHFn = abi.NewFunction(
		"addLiquidityETH",
		abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256, abi.Uint256, abi.Address, abi.Uint256),
		abi.NewTypes(abi.Uint256, abi.Uint256, abi.Uint256),
	)
	removeLiquidityFn = abi.NewFunction(
		"removeLiquidity",
		abi.NewTypes(abi.Address, abi.Address, abi.Uint256, abi.Uint256, abi.Uint256, abi.Address, abi.Uint256),
		abi.NewTypes(abi.Uint256, abi.Uint256),
	)
	removeLiquidityETHFn = abi.NewFunction(
		"removeLiquidityETH",
		abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256, abi.Uint256, abi.Address, abi.Uint256),
		abi.NewTypes(abi.Uint256, abi.Uint256),
	)
	removeLiquidityWithPermitFn = abi.NewFunction(
		"removeLiquidityWithPermit",
		abi.NewTypes(abi.Address, abi.Address, abi.Uint256, abi.Uint256, abi.Uint256, abi.Address, abi.Uint256, abi.Bool, abi.Uint8, abi.Bytes32, abi.Bytes32),
		abi.NewTypes(abi.Uint256, abi.Uint256),
	)
	removeLiquidityETHWithPermitFn = abi.NewFunction(
		"removeLiquidityETHWithPermit",
		abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256, abi.Uint256, abi.Address, abi.Uint256, abi.Bool, abi.Uint8, abi.Bytes32, abi.Bytes32),
		abi.NewTypes(abi.Uint256, abi.Uint256),
	)
	swapExactTokensForTokensFn = abi.NewFunction(
		"swapExactTokensForTokens",
		abi.NewTypes(abi.Uint256, abi.Uint256, abi.Slice(abi.Address), abi.Address, abi.Uint256),
		abi.NewTypes(abi.Slice(abi.Uint256)),
	)
	swapTokensForExactTokensFn = abi.NewFunction(
		"swapTokensForExactTokens",
		abi.NewTypes(abi.Uint256, abi.Uint256, abi.Slice(abi.Address), abi.Address, abi.Uint256),
		abi.NewTypes(abi.Slice(abi.Uint256)),
	)
	swapExactETHForTokensFn = abi.NewFunction(
		"swapExactETHForTokens",
		abi.NewTypes(abi.Uint256, abi.Slice(abi.Address), abi.Address, abi.Uint256),
		abi.NewTypes(abi.Slice(abi.Uint256)),
	)
	swapTokensForExactETHFn = abi.NewFunction(
		"swapTokensForExactETH",
		abi.NewTypes(abi.Uint256, abi.Uint256, abi.Slice(abi.Address), abi.Address, abi.Uint256),
		abi.NewTypes(abi.Slice(abi.Uint256)),
	)
	swapExactTokensForETHFn = abi.NewFunction(
		"swapExactTokensForETH",
		abi.NewTypes(abi.Uint256, abi.Uint256, abi.Slice(abi.Address), abi.Address, abi.Uint256),
		abi.NewTypes(abi.Slice(abi.Uint256)),
	)
	swapETHForExactTokensFn = abi.NewFunction(
		"swapETHForExactTokens",
		abi.NewTypes(abi.Uint256, abi.Slice(abi.Address), abi.Address, abi.Uint256),
		abi.NewTypes(abi.Slice(abi.Uint256)),
	)
	removeLiquidityETHSupportingFeeOnTransferTokensFn = abi.NewFunction(
		"removeLiquidityETHSupportingFeeOnTransferTokens",
		abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256, abi.Uint256, abi.Address, abi.Uint256),
		abi.NewTypes(abi.Uint256),
	)
	removeLiquidityETHWithPermitSupportingFeeOnTransferTokensFn = abi.NewFunction(
		"removeLiquidityETHWithPermitSupportingFeeOnTransferTokens",
		abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256, abi.Uint256, abi.Address, abi.Uint256, abi.Bool, abi.Uint8, abi.Bytes32, abi.Bytes32),
		abi.NewTypes(abi.Uint256),
	)
	swapExactTokensForTokensSupportingFeeOnTransferTokensFn = abi.NewFunction(
		"swapExactTokensForTokensSupportingFeeOnTransferTokens",
		abi.NewTypes(abi.Uint256, abi.Uint256, abi.Slice(abi.Address), abi.Address, abi.Uint256),
		nil,
	)
	swapExactETHForTokensSupportingFeeOnTransferTokensFn = abi.NewFunction(
		"swapExactETHForTokensSupportingFeeOnTransferTokens",
		abi.NewTypes(abi.Uint256, abi.Slice(abi.Address), abi.Address, abi.Uint256),
		nil,
	)
	swapExactTokensForETHSupportingFeeOnTransferTokensFn = abi.NewFunction(
		"swapExactTokensForETHSupportingFeeOnTransferTokens",
		abi.NewTypes(abi.Uint256, abi.Uint256, abi.Slice(abi.Address), abi.Address, abi.Uint256),
		nil,
	)
)

// UniswapV2Router predates Solidity custom errors too; every revert is a plain require(cond,
// reason) string, decoded via abi.DecodeRevert. Some of these also surface from Router's calls
// into UniswapV2Library (a different exact string than Router's own, despite the same suffix).
const (
	reasonExpired                         = "UniswapV2Router: EXPIRED"
	reasonInsufficientAAmount             = "UniswapV2Router: INSUFFICIENT_A_AMOUNT"
	reasonInsufficientBAmount             = "UniswapV2Router: INSUFFICIENT_B_AMOUNT"
	reasonRouterInsufficientOutput        = "UniswapV2Router: INSUFFICIENT_OUTPUT_AMOUNT"
	reasonExcessiveInputAmount            = "UniswapV2Router: EXCESSIVE_INPUT_AMOUNT"
	reasonRouterInvalidPath               = "UniswapV2Router: INVALID_PATH"
	reasonLibraryIdenticalAddressesOld    = "UniswapV2: IDENTICAL_ADDRESSES"
	reasonLibraryZeroAddressOld           = "UniswapV2: ZERO_ADDRESS"
	reasonLibraryInsufficientAmountOld    = "UniswapV2: INSUFFICIENT_AMOUNT"
	reasonLibraryInsufficientLiquidityOld = "UniswapV2: INSUFFICIENT_LIQUIDITY"
	reasonLibraryInsufficientInputOld     = "UniswapV2: INSUFFICIENT_INPUT_AMOUNT"
	reasonLibraryInsufficientOutputOld    = "UniswapV2: INSUFFICIENT_OUTPUT_AMOUNT"
	reasonLibraryInvalidPathOld           = "UniswapV2: INVALID_PATH"
	reasonLibraryIdenticalAddressesNew    = "UniswapV2Library: IDENTICAL_ADDRESSES"
	reasonLibraryZeroAddressNew           = "UniswapV2Library: ZERO_ADDRESS"
	reasonLibraryInsufficientAmountNew    = "UniswapV2Library: INSUFFICIENT_AMOUNT"
	reasonLibraryInsufficientLiquidityNew = "UniswapV2Library: INSUFFICIENT_LIQUIDITY"
	reasonLibraryInsufficientInputNew     = "UniswapV2Library: INSUFFICIENT_INPUT_AMOUNT"
	reasonLibraryInsufficientOutputNew    = "UniswapV2Library: INSUFFICIENT_OUTPUT_AMOUNT"
	reasonLibraryInvalidPathNew           = "UniswapV2Library: INVALID_PATH"
)

// ErrExpired, etc. mirror UniswapV2Router's own revert reasons, for errors.Is. Reasons shared in
// meaning with an existing Factory/Pair sentinel (identical addresses, zero address, insufficient
// liquidity/input/output amount, invalid path) reuse that sentinel instead of a new one.
var (
	ErrExpired              = errors.New("uniswapv2: expired")
	ErrInsufficientAAmount  = errors.New("uniswapv2: insufficient a amount")
	ErrInsufficientBAmount  = errors.New("uniswapv2: insufficient b amount")
	ErrExcessiveInputAmount = errors.New("uniswapv2: excessive input amount")
	ErrInvalidPath          = errors.New("uniswapv2: invalid path")
	ErrInsufficientAmount   = errors.New("uniswapv2: insufficient amount")
)

// EncodeRouterFactory returns the calldata for factory().
func EncodeRouterFactory() []byte {
	return routerFactoryFn.SelectorBytes()
}

// DecodeRouterFactory decodes the return data of factory.
func DecodeRouterFactory(data []byte) (*types.Address, error) {
	addr, err := routerFactoryFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeRouterFactory: %w", err)
	}
	return addr, nil
}

// EncodeWETH returns the calldata for WETH().
func EncodeWETH() []byte {
	return wethFn.SelectorBytes()
}

// DecodeWETH decodes the return data of WETH.
func DecodeWETH(data []byte) (*types.Address, error) {
	addr, err := wethFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeWETH: %w", err)
	}
	return addr, nil
}

// EncodeQuote returns the calldata for quote(amountA, reserveA, reserveB).
func EncodeQuote(amountA, reserveA, reserveB *big.Int) ([]byte, error) {
	data, err := quoteFn.EncodeCall(amountA, reserveA, reserveB)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeQuote: %w", err)
	}
	return data, nil
}

// DecodeQuote decodes the return data of quote.
func DecodeQuote(data []byte) (*big.Int, error) {
	n, err := quoteFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeQuote: %w", err)
	}
	return n, nil
}

// EncodeGetAmountOut returns the calldata for getAmountOut(amountIn, reserveIn, reserveOut).
func EncodeGetAmountOut(amountIn, reserveIn, reserveOut *big.Int) ([]byte, error) {
	data, err := getAmountOutFn.EncodeCall(amountIn, reserveIn, reserveOut)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeGetAmountOut: %w", err)
	}
	return data, nil
}

// DecodeGetAmountOut decodes the return data of getAmountOut.
func DecodeGetAmountOut(data []byte) (*big.Int, error) {
	n, err := getAmountOutFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeGetAmountOut: %w", err)
	}
	return n, nil
}

// EncodeGetAmountIn returns the calldata for getAmountIn(amountOut, reserveIn, reserveOut).
func EncodeGetAmountIn(amountOut, reserveIn, reserveOut *big.Int) ([]byte, error) {
	data, err := getAmountInFn.EncodeCall(amountOut, reserveIn, reserveOut)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeGetAmountIn: %w", err)
	}
	return data, nil
}

// DecodeGetAmountIn decodes the return data of getAmountIn.
func DecodeGetAmountIn(data []byte) (*big.Int, error) {
	n, err := getAmountInFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeGetAmountIn: %w", err)
	}
	return n, nil
}

// EncodeGetAmountsOut returns the calldata for getAmountsOut(amountIn, path).
func EncodeGetAmountsOut(amountIn *big.Int, path []*types.Address) ([]byte, error) {
	data, err := getAmountsOutFn.EncodeCall(amountIn, path)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeGetAmountsOut: %w", err)
	}
	return data, nil
}

// DecodeGetAmountsOut decodes the return data of getAmountsOut.
func DecodeGetAmountsOut(data []byte) ([]*big.Int, error) {
	amounts, err := decodeAmounts(getAmountsOutFn, data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeGetAmountsOut: %w", err)
	}
	return amounts, nil
}

// EncodeGetAmountsIn returns the calldata for getAmountsIn(amountOut, path).
func EncodeGetAmountsIn(amountOut *big.Int, path []*types.Address) ([]byte, error) {
	data, err := getAmountsInFn.EncodeCall(amountOut, path)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeGetAmountsIn: %w", err)
	}
	return data, nil
}

// DecodeGetAmountsIn decodes the return data of getAmountsIn.
func DecodeGetAmountsIn(data []byte) ([]*big.Int, error) {
	amounts, err := decodeAmounts(getAmountsInFn, data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeGetAmountsIn: %w", err)
	}
	return amounts, nil
}

// EncodeAddLiquidity returns the calldata for addLiquidity(tokenA, tokenB, amountADesired,
// amountBDesired, amountAMin, amountBMin, to, deadline).
func EncodeAddLiquidity(tokenA, tokenB *types.Address, amountADesired, amountBDesired, amountAMin, amountBMin *big.Int, to *types.Address, deadline *big.Int) ([]byte, error) {
	data, err := addLiquidityFn.EncodeCall(tokenA, tokenB, amountADesired, amountBDesired, amountAMin, amountBMin, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeAddLiquidity: %w", err)
	}
	return data, nil
}

// DecodeAddLiquidity decodes the return data of addLiquidity. Only meaningful from an eth_call
// simulation; a mined transaction's return value isn't retrievable.
func DecodeAddLiquidity(data []byte) (*AddLiquidityResult, error) {
	vals, err := addLiquidityFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeAddLiquidity: %w", err)
	}
	amountA, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv2: decodeAddLiquidity: unexpected decoded type for amountA: %T", vals[0])
	}
	amountB, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv2: decodeAddLiquidity: unexpected decoded type for amountB: %T", vals[1])
	}
	liquidity, ok := vals[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv2: decodeAddLiquidity: unexpected decoded type for liquidity: %T", vals[2])
	}
	return NewAddLiquidityResult(amountA, amountB, liquidity), nil
}

// EncodeAddLiquidityETH returns the calldata for addLiquidityETH(token, amountTokenDesired,
// amountTokenMin, amountETHMin, to, deadline). It's payable; the ETH side rides the transaction's
// value, not an argument here.
func EncodeAddLiquidityETH(token *types.Address, amountTokenDesired, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline *big.Int) ([]byte, error) {
	data, err := addLiquidityETHFn.EncodeCall(token, amountTokenDesired, amountTokenMin, amountETHMin, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeAddLiquidityETH: %w", err)
	}
	return data, nil
}

// DecodeAddLiquidityETH decodes the return data of addLiquidityETH. Only meaningful from an
// eth_call simulation; a mined transaction's return value isn't retrievable.
func DecodeAddLiquidityETH(data []byte) (*AddLiquidityETHResult, error) {
	vals, err := addLiquidityETHFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeAddLiquidityETH: %w", err)
	}
	amountToken, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv2: decodeAddLiquidityETH: unexpected decoded type for amountToken: %T", vals[0])
	}
	amountETH, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv2: decodeAddLiquidityETH: unexpected decoded type for amountETH: %T", vals[1])
	}
	liquidity, ok := vals[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv2: decodeAddLiquidityETH: unexpected decoded type for liquidity: %T", vals[2])
	}
	return NewAddLiquidityETHResult(amountToken, amountETH, liquidity), nil
}

// EncodeRemoveLiquidity returns the calldata for removeLiquidity(tokenA, tokenB, liquidity,
// amountAMin, amountBMin, to, deadline).
func EncodeRemoveLiquidity(tokenA, tokenB *types.Address, liquidity, amountAMin, amountBMin *big.Int, to *types.Address, deadline *big.Int) ([]byte, error) {
	data, err := removeLiquidityFn.EncodeCall(tokenA, tokenB, liquidity, amountAMin, amountBMin, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeRemoveLiquidity: %w", err)
	}
	return data, nil
}

// DecodeRemoveLiquidity decodes the return data of removeLiquidity. Only meaningful from an
// eth_call simulation; a mined transaction's return value isn't retrievable.
func DecodeRemoveLiquidity(data []byte) (*RemoveLiquidityResult, error) {
	amountA, amountB, err := decodeAmountPair(removeLiquidityFn, data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeRemoveLiquidity: %w", err)
	}
	return NewRemoveLiquidityResult(amountA, amountB), nil
}

// EncodeRemoveLiquidityETH returns the calldata for removeLiquidityETH(token, liquidity,
// amountTokenMin, amountETHMin, to, deadline).
func EncodeRemoveLiquidityETH(token *types.Address, liquidity, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline *big.Int) ([]byte, error) {
	data, err := removeLiquidityETHFn.EncodeCall(token, liquidity, amountTokenMin, amountETHMin, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeRemoveLiquidityETH: %w", err)
	}
	return data, nil
}

// DecodeRemoveLiquidityETH decodes the return data of removeLiquidityETH. Only meaningful from
// an eth_call simulation; a mined transaction's return value isn't retrievable.
func DecodeRemoveLiquidityETH(data []byte) (*RemoveLiquidityETHResult, error) {
	amountToken, amountETH, err := decodeAmountPair(removeLiquidityETHFn, data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeRemoveLiquidityETH: %w", err)
	}
	return NewRemoveLiquidityETHResult(amountToken, amountETH), nil
}

// EncodeRemoveLiquidityWithPermit returns the calldata for removeLiquidityWithPermit(tokenA,
// tokenB, liquidity, amountAMin, amountBMin, to, deadline, approveMax, v, r, s), where v, r, s
// come from sig -- the EIP-712 permit signature, not a transaction signature.
func EncodeRemoveLiquidityWithPermit(tokenA, tokenB *types.Address, liquidity, amountAMin, amountBMin *big.Int, to *types.Address, deadline *big.Int, approveMax bool, sig *types.Signature) ([]byte, error) {
	v, r, s := permitSignatureParts(sig)
	data, err := removeLiquidityWithPermitFn.EncodeCall(tokenA, tokenB, liquidity, amountAMin, amountBMin, to, deadline, approveMax, v, r, s)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeRemoveLiquidityWithPermit: %w", err)
	}
	return data, nil
}

// DecodeRemoveLiquidityWithPermit decodes the return data of removeLiquidityWithPermit. Only
// meaningful from an eth_call simulation; a mined transaction's return value isn't retrievable.
func DecodeRemoveLiquidityWithPermit(data []byte) (*RemoveLiquidityResult, error) {
	amountA, amountB, err := decodeAmountPair(removeLiquidityWithPermitFn, data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeRemoveLiquidityWithPermit: %w", err)
	}
	return NewRemoveLiquidityResult(amountA, amountB), nil
}

// EncodeRemoveLiquidityETHWithPermit returns the calldata for removeLiquidityETHWithPermit(token,
// liquidity, amountTokenMin, amountETHMin, to, deadline, approveMax, v, r, s), where v, r, s come
// from sig -- the EIP-712 permit signature, not a transaction signature.
func EncodeRemoveLiquidityETHWithPermit(token *types.Address, liquidity, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline *big.Int, approveMax bool, sig *types.Signature) ([]byte, error) {
	v, r, s := permitSignatureParts(sig)
	data, err := removeLiquidityETHWithPermitFn.EncodeCall(token, liquidity, amountTokenMin, amountETHMin, to, deadline, approveMax, v, r, s)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeRemoveLiquidityETHWithPermit: %w", err)
	}
	return data, nil
}

// DecodeRemoveLiquidityETHWithPermit decodes the return data of removeLiquidityETHWithPermit.
// Only meaningful from an eth_call simulation; a mined transaction's return value isn't
// retrievable.
func DecodeRemoveLiquidityETHWithPermit(data []byte) (*RemoveLiquidityETHResult, error) {
	amountToken, amountETH, err := decodeAmountPair(removeLiquidityETHWithPermitFn, data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeRemoveLiquidityETHWithPermit: %w", err)
	}
	return NewRemoveLiquidityETHResult(amountToken, amountETH), nil
}

// EncodeSwapExactTokensForTokens returns the calldata for swapExactTokensForTokens(amountIn,
// amountOutMin, path, to, deadline).
func EncodeSwapExactTokensForTokens(amountIn, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) ([]byte, error) {
	data, err := swapExactTokensForTokensFn.EncodeCall(amountIn, amountOutMin, path, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeSwapExactTokensForTokens: %w", err)
	}
	return data, nil
}

// DecodeSwapExactTokensForTokens decodes the return data of swapExactTokensForTokens. Only
// meaningful from an eth_call simulation; a mined transaction's return value isn't retrievable.
func DecodeSwapExactTokensForTokens(data []byte) ([]*big.Int, error) {
	amounts, err := decodeAmounts(swapExactTokensForTokensFn, data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeSwapExactTokensForTokens: %w", err)
	}
	return amounts, nil
}

// EncodeSwapTokensForExactTokens returns the calldata for swapTokensForExactTokens(amountOut,
// amountInMax, path, to, deadline).
func EncodeSwapTokensForExactTokens(amountOut, amountInMax *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) ([]byte, error) {
	data, err := swapTokensForExactTokensFn.EncodeCall(amountOut, amountInMax, path, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeSwapTokensForExactTokens: %w", err)
	}
	return data, nil
}

// DecodeSwapTokensForExactTokens decodes the return data of swapTokensForExactTokens. Only
// meaningful from an eth_call simulation; a mined transaction's return value isn't retrievable.
func DecodeSwapTokensForExactTokens(data []byte) ([]*big.Int, error) {
	amounts, err := decodeAmounts(swapTokensForExactTokensFn, data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeSwapTokensForExactTokens: %w", err)
	}
	return amounts, nil
}

// EncodeSwapExactETHForTokens returns the calldata for swapExactETHForTokens(amountOutMin, path,
// to, deadline). It's payable; the input amount rides the transaction's value, not an argument
// here.
func EncodeSwapExactETHForTokens(amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) ([]byte, error) {
	data, err := swapExactETHForTokensFn.EncodeCall(amountOutMin, path, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeSwapExactETHForTokens: %w", err)
	}
	return data, nil
}

// DecodeSwapExactETHForTokens decodes the return data of swapExactETHForTokens. Only meaningful
// from an eth_call simulation; a mined transaction's return value isn't retrievable.
func DecodeSwapExactETHForTokens(data []byte) ([]*big.Int, error) {
	amounts, err := decodeAmounts(swapExactETHForTokensFn, data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeSwapExactETHForTokens: %w", err)
	}
	return amounts, nil
}

// EncodeSwapTokensForExactETH returns the calldata for swapTokensForExactETH(amountOut,
// amountInMax, path, to, deadline).
func EncodeSwapTokensForExactETH(amountOut, amountInMax *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) ([]byte, error) {
	data, err := swapTokensForExactETHFn.EncodeCall(amountOut, amountInMax, path, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeSwapTokensForExactETH: %w", err)
	}
	return data, nil
}

// DecodeSwapTokensForExactETH decodes the return data of swapTokensForExactETH. Only meaningful
// from an eth_call simulation; a mined transaction's return value isn't retrievable.
func DecodeSwapTokensForExactETH(data []byte) ([]*big.Int, error) {
	amounts, err := decodeAmounts(swapTokensForExactETHFn, data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeSwapTokensForExactETH: %w", err)
	}
	return amounts, nil
}

// EncodeSwapExactTokensForETH returns the calldata for swapExactTokensForETH(amountIn,
// amountOutMin, path, to, deadline).
func EncodeSwapExactTokensForETH(amountIn, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) ([]byte, error) {
	data, err := swapExactTokensForETHFn.EncodeCall(amountIn, amountOutMin, path, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeSwapExactTokensForETH: %w", err)
	}
	return data, nil
}

// DecodeSwapExactTokensForETH decodes the return data of swapExactTokensForETH. Only meaningful
// from an eth_call simulation; a mined transaction's return value isn't retrievable.
func DecodeSwapExactTokensForETH(data []byte) ([]*big.Int, error) {
	amounts, err := decodeAmounts(swapExactTokensForETHFn, data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeSwapExactTokensForETH: %w", err)
	}
	return amounts, nil
}

// EncodeSwapETHForExactTokens returns the calldata for swapETHForExactTokens(amountOut, path, to,
// deadline). It's payable; the max input amount rides the transaction's value, not an argument
// here.
func EncodeSwapETHForExactTokens(amountOut *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) ([]byte, error) {
	data, err := swapETHForExactTokensFn.EncodeCall(amountOut, path, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeSwapETHForExactTokens: %w", err)
	}
	return data, nil
}

// DecodeSwapETHForExactTokens decodes the return data of swapETHForExactTokens. Only meaningful
// from an eth_call simulation; a mined transaction's return value isn't retrievable.
func DecodeSwapETHForExactTokens(data []byte) ([]*big.Int, error) {
	amounts, err := decodeAmounts(swapETHForExactTokensFn, data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeSwapETHForExactTokens: %w", err)
	}
	return amounts, nil
}

// EncodeRemoveLiquidityETHSupportingFeeOnTransferTokens returns the calldata for
// removeLiquidityETHSupportingFeeOnTransferTokens(token, liquidity, amountTokenMin,
// amountETHMin, to, deadline).
func EncodeRemoveLiquidityETHSupportingFeeOnTransferTokens(token *types.Address, liquidity, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline *big.Int) ([]byte, error) {
	data, err := removeLiquidityETHSupportingFeeOnTransferTokensFn.EncodeCall(token, liquidity, amountTokenMin, amountETHMin, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeRemoveLiquidityETHSupportingFeeOnTransferTokens: %w", err)
	}
	return data, nil
}

// DecodeRemoveLiquidityETHSupportingFeeOnTransferTokens decodes the return data of
// removeLiquidityETHSupportingFeeOnTransferTokens -- the amountETH withdrawn. Only meaningful
// from an eth_call simulation; a mined transaction's return value isn't retrievable.
func DecodeRemoveLiquidityETHSupportingFeeOnTransferTokens(data []byte) (*big.Int, error) {
	n, err := removeLiquidityETHSupportingFeeOnTransferTokensFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeRemoveLiquidityETHSupportingFeeOnTransferTokens: %w", err)
	}
	return n, nil
}

// EncodeRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens returns the calldata for
// removeLiquidityETHWithPermitSupportingFeeOnTransferTokens(token, liquidity, amountTokenMin,
// amountETHMin, to, deadline, approveMax, v, r, s), where v, r, s come from sig -- the EIP-712
// permit signature, not a transaction signature.
func EncodeRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(token *types.Address, liquidity, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline *big.Int, approveMax bool, sig *types.Signature) ([]byte, error) {
	v, r, s := permitSignatureParts(sig)
	data, err := removeLiquidityETHWithPermitSupportingFeeOnTransferTokensFn.EncodeCall(token, liquidity, amountTokenMin, amountETHMin, to, deadline, approveMax, v, r, s)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens: %w", err)
	}
	return data, nil
}

// DecodeRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens decodes the return data of
// removeLiquidityETHWithPermitSupportingFeeOnTransferTokens -- the amountETH withdrawn. Only
// meaningful from an eth_call simulation; a mined transaction's return value isn't retrievable.
func DecodeRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(data []byte) (*big.Int, error) {
	n, err := removeLiquidityETHWithPermitSupportingFeeOnTransferTokensFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens: %w", err)
	}
	return n, nil
}

// EncodeSwapExactTokensForTokensSupportingFeeOnTransferTokens returns the calldata for
// swapExactTokensForTokensSupportingFeeOnTransferTokens(amountIn, amountOutMin, path, to,
// deadline). Unlike its non-fee-on-transfer counterpart, this has no return value.
func EncodeSwapExactTokensForTokensSupportingFeeOnTransferTokens(amountIn, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) ([]byte, error) {
	data, err := swapExactTokensForTokensSupportingFeeOnTransferTokensFn.EncodeCall(amountIn, amountOutMin, path, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeSwapExactTokensForTokensSupportingFeeOnTransferTokens: %w", err)
	}
	return data, nil
}

// EncodeSwapExactETHForTokensSupportingFeeOnTransferTokens returns the calldata for
// swapExactETHForTokensSupportingFeeOnTransferTokens(amountOutMin, path, to, deadline). It's
// payable; the input amount rides the transaction's value, not an argument here. Unlike its
// non-fee-on-transfer counterpart, this has no return value.
func EncodeSwapExactETHForTokensSupportingFeeOnTransferTokens(amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) ([]byte, error) {
	data, err := swapExactETHForTokensSupportingFeeOnTransferTokensFn.EncodeCall(amountOutMin, path, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeSwapExactETHForTokensSupportingFeeOnTransferTokens: %w", err)
	}
	return data, nil
}

// EncodeSwapExactTokensForETHSupportingFeeOnTransferTokens returns the calldata for
// swapExactTokensForETHSupportingFeeOnTransferTokens(amountIn, amountOutMin, path, to,
// deadline). Unlike its non-fee-on-transfer counterpart, this has no return value.
func EncodeSwapExactTokensForETHSupportingFeeOnTransferTokens(amountIn, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) ([]byte, error) {
	data, err := swapExactTokensForETHSupportingFeeOnTransferTokensFn.EncodeCall(amountIn, amountOutMin, path, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeSwapExactTokensForETHSupportingFeeOnTransferTokens: %w", err)
	}
	return data, nil
}

// permitSignatureParts extracts (v, r, s) from sig in the shape UniswapV2Router's
// *WithPermit functions expect -- the same shape erc2612.EncodePermit uses.
func permitSignatureParts(sig *types.Signature) (v uint8, r, s []byte) {
	v = uint8(sig.LegacyV().Uint64())
	r = make([]byte, 32)
	sig.R().FillBytes(r)
	s = make([]byte, 32)
	sig.S().FillBytes(s)
	return v, r, s
}

// decodeAmounts decodes a single uint256[] return value (the amounts path-aligned with a swap's
// or getAmounts{In,Out}'s path) via fn.
func decodeAmounts(fn *abi.Function, data []byte) ([]*big.Int, error) {
	vals, err := fn.DecodeReturn(data)
	if err != nil {
		return nil, err
	}
	if len(vals) == 0 {
		return nil, fmt.Errorf("no return values")
	}
	elems, ok := vals[0].([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected decoded type for amounts: %T", vals[0])
	}
	amounts := make([]*big.Int, len(elems))
	for i, e := range elems {
		n, ok := e.(*big.Int)
		if !ok {
			return nil, fmt.Errorf("unexpected decoded type for amounts[%d]: %T", i, e)
		}
		amounts[i] = n
	}
	return amounts, nil
}

// decodeAmountPair decodes a (uint256, uint256) return value via fn.
func decodeAmountPair(fn *abi.Function, data []byte) (first, second *big.Int, err error) {
	vals, err := fn.DecodeReturn(data)
	if err != nil {
		return nil, nil, err
	}
	first, ok := vals[0].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("unexpected decoded type for first: %T", vals[0])
	}
	second, ok = vals[1].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("unexpected decoded type for second: %T", vals[1])
	}
	return first, second, nil
}
