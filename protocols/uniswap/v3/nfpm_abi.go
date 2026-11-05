package v3

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Tuple types for INonfungiblePositionManager's struct parameters.
var (
	mintParamsT = abi.Tuple(
		abi.Address,
		abi.Address,
		abi.Uint24,
		abi.Int24,
		abi.Int24,
		abi.Uint256,
		abi.Uint256,
		abi.Uint256,
		abi.Uint256,
		abi.Address,
		abi.Uint256,
	)
	increaseLiquidityParamsT = abi.Tuple(
		abi.Uint256,
		abi.Uint256,
		abi.Uint256,
		abi.Uint256,
		abi.Uint256,
		abi.Uint256,
	)
	decreaseLiquidityParamsT = abi.Tuple(
		abi.Uint256,
		abi.Uint128,
		abi.Uint256,
		abi.Uint256,
		abi.Uint256,
	)
	collectParamsT = abi.Tuple(
		abi.Uint256,
		abi.Address,
		abi.Uint128,
		abi.Uint128,
	)
)

// Function signatures from INonfungiblePositionManager and its base contracts
// (PeripheryImmutableState, PoolInitializer, IPeripheryPayments, Multicall, IERC721Enumerable,
// IERC721Permit), for encoding calls and decoding results.
var (
	nfpmFactoryFn = abi.NewFunction(
		"factory",
		nil,
		abi.NewTypes(abi.Address),
	)
	nfpmWeth9Fn = abi.NewFunction(
		"WETH9",
		nil,
		abi.NewTypes(abi.Address),
	)

	createAndInitializePoolIfNecessaryFn = abi.NewFunction(
		"createAndInitializePoolIfNecessary",
		abi.NewTypes(abi.Address, abi.Address, abi.Uint24, abi.Uint160),
		abi.NewTypes(abi.Address),
	)

	nfpmPositionsFn = abi.NewFunction(
		"positions",
		abi.NewTypes(abi.Uint256),
		abi.NewTypes(
			abi.Uint96,
			abi.Address,
			abi.Address,
			abi.Address,
			abi.Uint24,
			abi.Int24,
			abi.Int24,
			abi.Uint128,
			abi.Uint256,
			abi.Uint256,
			abi.Uint128,
			abi.Uint128,
		),
	)

	nfpmMintFn = abi.NewFunction(
		"mint",
		abi.NewTypes(mintParamsT),
		abi.NewTypes(abi.Uint256, abi.Uint128, abi.Uint256, abi.Uint256),
	)
	increaseLiquidityFn = abi.NewFunction(
		"increaseLiquidity",
		abi.NewTypes(increaseLiquidityParamsT),
		abi.NewTypes(abi.Uint128, abi.Uint256, abi.Uint256),
	)
	decreaseLiquidityFn = abi.NewFunction(
		"decreaseLiquidity",
		abi.NewTypes(decreaseLiquidityParamsT),
		abi.NewTypes(abi.Uint256, abi.Uint256),
	)
	nfpmCollectFn = abi.NewFunction(
		"collect",
		abi.NewTypes(collectParamsT),
		abi.NewTypes(abi.Uint256, abi.Uint256),
	)
	nfpmBurnFn = abi.NewFunction(
		"burn",
		abi.NewTypes(abi.Uint256),
		nil,
	)

	unwrapWETH9Fn = abi.NewFunction(
		"unwrapWETH9",
		abi.NewTypes(abi.Uint256, abi.Address),
		nil,
	)
	sweepTokenFn = abi.NewFunction(
		"sweepToken",
		abi.NewTypes(abi.Address, abi.Uint256, abi.Address),
		nil,
	)
	refundETHFn = abi.NewFunction("refundETH", nil, nil)

	nfpmMulticallFn = abi.NewFunction(
		"multicall",
		abi.NewTypes(abi.Slice(abi.Bytes)),
		abi.NewTypes(abi.Slice(abi.Bytes)),
	)

	totalSupplyFn = abi.NewFunction(
		"totalSupply",
		nil,
		abi.NewTypes(abi.Uint256),
	)
	tokenOfOwnerByIndexFn = abi.NewFunction(
		"tokenOfOwnerByIndex",
		abi.NewTypes(abi.Address, abi.Uint256),
		abi.NewTypes(abi.Uint256),
	)
	tokenByIndexFn = abi.NewFunction(
		"tokenByIndex",
		abi.NewTypes(abi.Uint256),
		abi.NewTypes(abi.Uint256),
	)

	permitTypehashFn = abi.NewFunction(
		"PERMIT_TYPEHASH",
		nil,
		abi.NewTypes(abi.Bytes32),
	)
	domainSeparatorFn = abi.NewFunction(
		"DOMAIN_SEPARATOR",
		nil,
		abi.NewTypes(abi.Bytes32),
	)
	nftPermitFn = abi.NewFunction(
		"permit",
		abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256, abi.Uint8, abi.Bytes32, abi.Bytes32),
		nil,
	)
)

// Events from INonfungiblePositionManager. Transfer/Approval/ApprovalForAll are decoded by
// standards/erc721 instead, since NFPM is also a plain ERC721.
var (
	increaseLiquidityEvent = abi.NewEvent("IncreaseLiquidity", []abi.EventParam{
		{Type: abi.Uint256, Indexed: true},
		{Type: abi.Uint128},
		{Type: abi.Uint256},
		{Type: abi.Uint256},
	}, false)

	decreaseLiquidityEvent = abi.NewEvent("DecreaseLiquidity", []abi.EventParam{
		{Type: abi.Uint256, Indexed: true},
		{Type: abi.Uint128},
		{Type: abi.Uint256},
		{Type: abi.Uint256},
	}, false)

	positionCollectEvent = abi.NewEvent("Collect", []abi.EventParam{
		{Type: abi.Uint256, Indexed: true},
		{Type: abi.Address},
		{Type: abi.Uint256},
		{Type: abi.Uint256},
	}, false)
)

// Revert reasons reachable from NonfungiblePositionManager's own functions and the base
// contracts it inherits (PeripheryValidation, PeripheryPayments, PeripheryTransferHelper,
// ERC721Permit). Reverts from the vendored ERC721 base (which uses Solidity custom errors, not
// string reasons, in this checkout) aren't decoded here and fall back to ErrReverted.
const (
	reasonTransactionTooOld      = "Transaction too old"
	reasonInsufficientWETH9      = "Insufficient WETH9"
	reasonInsufficientToken      = "Insufficient token"
	reasonNotWETH9               = "Not WETH9"
	reasonTransferFromFailed     = "STF"
	reasonTokenTransferFailed    = "ST"
	reasonApproveFailed          = "SA"
	reasonTransferETHFailed      = "STE"
	reasonInvalidTokenID         = "Invalid token ID"
	reasonNotApproved            = "Not approved"
	reasonPriceSlippageCheck     = "Price slippage check"
	reasonNotCleared             = "Not cleared"
	reasonPermitExpired          = "Permit expired"
	reasonApprovalToCurrentOwner = "ERC721Permit: approval to current owner"
	reasonUnauthorized           = "Unauthorized"
	reasonInvalidSignature       = "Invalid signature"
)

var (
	ErrTransactionTooOld      = errors.New("uniswapv3: transaction too old")
	ErrInsufficientWETH9      = errors.New("uniswapv3: insufficient weth9")
	ErrInsufficientToken      = errors.New("uniswapv3: insufficient token")
	ErrNotWETH9               = errors.New("uniswapv3: not weth9")
	ErrTransferFromFailed     = errors.New("uniswapv3: transfer from failed")
	ErrTokenTransferFailed    = errors.New("uniswapv3: token transfer failed")
	ErrApproveFailed          = errors.New("uniswapv3: approve failed")
	ErrTransferETHFailed      = errors.New("uniswapv3: transfer eth failed")
	ErrInvalidTokenID         = errors.New("uniswapv3: invalid token id")
	ErrNotApproved            = errors.New("uniswapv3: not approved")
	ErrPriceSlippageCheck     = errors.New("uniswapv3: price slippage check")
	ErrNotCleared             = errors.New("uniswapv3: not cleared")
	ErrPermitExpired          = errors.New("uniswapv3: permit expired")
	ErrApprovalToCurrentOwner = errors.New("uniswapv3: approval to current owner")
	ErrUnauthorized           = errors.New("uniswapv3: unauthorized")
	ErrInvalidSignature       = errors.New("uniswapv3: invalid signature")
)

// EncodeNfpmFactory returns the calldata for factory().
func EncodeNfpmFactory() []byte {
	return nfpmFactoryFn.SelectorBytes()
}

// DecodeNfpmFactory decodes the return data of factory.
func DecodeNfpmFactory(data []byte) (*types.Address, error) {
	addr, err := nfpmFactoryFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeNfpmFactory: %w", err)
	}
	return addr, nil
}

// EncodeWETH9 returns the calldata for WETH9().
func EncodeWETH9() []byte {
	return nfpmWeth9Fn.SelectorBytes()
}

// DecodeWETH9 decodes the return data of WETH9.
func DecodeWETH9(data []byte) (*types.Address, error) {
	addr, err := nfpmWeth9Fn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeWETH9: %w", err)
	}
	return addr, nil
}

// EncodeCreateAndInitializePoolIfNecessary returns the calldata for
// createAndInitializePoolIfNecessary(token0, token1, fee, sqrtPriceX96).
func EncodeCreateAndInitializePoolIfNecessary(
	token0, token1 *types.Address,
	fee *big.Int,
	sqrtPriceX96 *big.Int,
) ([]byte, error) {
	data, err := createAndInitializePoolIfNecessaryFn.EncodeCall(token0, token1, fee, sqrtPriceX96)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeCreateAndInitializePoolIfNecessary: %w", err)
	}
	return data, nil
}

// DecodeCreateAndInitializePoolIfNecessary decodes the return data of
// createAndInitializePoolIfNecessary -- the pool's address.
func DecodeCreateAndInitializePoolIfNecessary(data []byte) (*types.Address, error) {
	addr, err := createAndInitializePoolIfNecessaryFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeCreateAndInitializePoolIfNecessary: %w", err)
	}
	return addr, nil
}

// EncodeNfpmPositions returns the calldata for positions(tokenId).
func EncodeNfpmPositions(tokenId *big.Int) ([]byte, error) {
	data, err := nfpmPositionsFn.EncodeCall(tokenId)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeNfpmPositions: %w", err)
	}
	return data, nil
}

// DecodeNfpmPositions decodes the return data of positions.
func DecodeNfpmPositions(data []byte) (*Position, error) {
	vals, err := nfpmPositionsFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeNfpmPositions: %w", err)
	}
	if len(vals) != 12 {
		return nil, fmt.Errorf("uniswapv3: decodeNfpmPositions: expected 12 return values, got %d", len(vals))
	}
	nonce, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeNfpmPositions: unexpected decoded type for nonce: %T", vals[0])
	}
	operator, ok := vals[1].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeNfpmPositions: unexpected decoded type for operator: %T", vals[1])
	}
	token0, ok := vals[2].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeNfpmPositions: unexpected decoded type for token0: %T", vals[2])
	}
	token1, ok := vals[3].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeNfpmPositions: unexpected decoded type for token1: %T", vals[3])
	}
	fee, ok := vals[4].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeNfpmPositions: unexpected decoded type for fee: %T", vals[4])
	}
	tickLower, ok := vals[5].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeNfpmPositions: unexpected decoded type for tickLower: %T", vals[5])
	}
	tickUpper, ok := vals[6].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeNfpmPositions: unexpected decoded type for tickUpper: %T", vals[6])
	}
	liquidity, ok := vals[7].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeNfpmPositions: unexpected decoded type for liquidity: %T", vals[7])
	}
	feeGrowthInside0LastX128, ok := vals[8].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeNfpmPositions: unexpected decoded type for feeGrowthInside0LastX128: %T", vals[8])
	}
	feeGrowthInside1LastX128, ok := vals[9].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeNfpmPositions: unexpected decoded type for feeGrowthInside1LastX128: %T", vals[9])
	}
	tokensOwed0, ok := vals[10].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeNfpmPositions: unexpected decoded type for tokensOwed0: %T", vals[10])
	}
	tokensOwed1, ok := vals[11].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeNfpmPositions: unexpected decoded type for tokensOwed1: %T", vals[11])
	}
	return NewPosition(nonce, operator, token0, token1, fee, tickLower, tickUpper, liquidity, feeGrowthInside0LastX128, feeGrowthInside1LastX128, tokensOwed0, tokensOwed1), nil
}

// EncodeNfpmMint returns the calldata for mint(params).
func EncodeNfpmMint(
	token0, token1 *types.Address,
	fee, tickLower, tickUpper, amount0Desired, amount1Desired, amount0Min, amount1Min *big.Int,
	recipient *types.Address,
	deadline *big.Int,
) ([]byte, error) {
	data, err := nfpmMintFn.EncodeCall([]any{token0, token1, fee, tickLower, tickUpper, amount0Desired, amount1Desired, amount0Min, amount1Min, recipient, deadline})
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeNfpmMint: %w", err)
	}
	return data, nil
}

// DecodeNfpmMint decodes the return data of mint -- tokenId, liquidity, amount0, amount1.
func DecodeNfpmMint(data []byte) (
	tokenId, liquidity, amount0, amount1 *big.Int,
	err error,
) {
	vals, err := nfpmMintFn.DecodeReturn(data)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("uniswapv3: decodeNfpmMint: %w", err)
	}
	if len(vals) != 4 {
		return nil, nil, nil, nil, fmt.Errorf("uniswapv3: decodeNfpmMint: expected 4 return values, got %d", len(vals))
	}
	tokenId, ok := vals[0].(*big.Int)
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("uniswapv3: decodeNfpmMint: unexpected decoded type for tokenId: %T", vals[0])
	}
	liquidity, ok = vals[1].(*big.Int)
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("uniswapv3: decodeNfpmMint: unexpected decoded type for liquidity: %T", vals[1])
	}
	amount0, ok = vals[2].(*big.Int)
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("uniswapv3: decodeNfpmMint: unexpected decoded type for amount0: %T", vals[2])
	}
	amount1, ok = vals[3].(*big.Int)
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("uniswapv3: decodeNfpmMint: unexpected decoded type for amount1: %T", vals[3])
	}
	return tokenId, liquidity, amount0, amount1, nil
}

// EncodeIncreaseLiquidity returns the calldata for increaseLiquidity(params).
func EncodeIncreaseLiquidity(tokenId, amount0Desired, amount1Desired, amount0Min, amount1Min, deadline *big.Int) ([]byte, error) {
	data, err := increaseLiquidityFn.EncodeCall([]any{tokenId, amount0Desired, amount1Desired, amount0Min, amount1Min, deadline})
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeIncreaseLiquidity: %w", err)
	}
	return data, nil
}

// DecodeIncreaseLiquidity decodes the return data of increaseLiquidity -- liquidity, amount0, amount1.
func DecodeIncreaseLiquidity(data []byte) (liquidity, amount0, amount1 *big.Int, err error) {
	vals, err := increaseLiquidityFn.DecodeReturn(data)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("uniswapv3: decodeIncreaseLiquidity: %w", err)
	}
	if len(vals) != 3 {
		return nil, nil, nil, fmt.Errorf("uniswapv3: decodeIncreaseLiquidity: expected 3 return values, got %d", len(vals))
	}
	liquidity, ok := vals[0].(*big.Int)
	if !ok {
		return nil, nil, nil, fmt.Errorf("uniswapv3: decodeIncreaseLiquidity: unexpected decoded type for liquidity: %T", vals[0])
	}
	amount0, ok = vals[1].(*big.Int)
	if !ok {
		return nil, nil, nil, fmt.Errorf("uniswapv3: decodeIncreaseLiquidity: unexpected decoded type for amount0: %T", vals[1])
	}
	amount1, ok = vals[2].(*big.Int)
	if !ok {
		return nil, nil, nil, fmt.Errorf("uniswapv3: decodeIncreaseLiquidity: unexpected decoded type for amount1: %T", vals[2])
	}
	return liquidity, amount0, amount1, nil
}

// EncodeDecreaseLiquidity returns the calldata for decreaseLiquidity(params).
func EncodeDecreaseLiquidity(tokenId, liquidity, amount0Min, amount1Min, deadline *big.Int) ([]byte, error) {
	data, err := decreaseLiquidityFn.EncodeCall([]any{tokenId, liquidity, amount0Min, amount1Min, deadline})
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeDecreaseLiquidity: %w", err)
	}
	return data, nil
}

// DecodeDecreaseLiquidity decodes the return data of decreaseLiquidity -- amount0, amount1.
func DecodeDecreaseLiquidity(data []byte) (amount0, amount1 *big.Int, err error) {
	vals, err := decreaseLiquidityFn.DecodeReturn(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: decodeDecreaseLiquidity: %w", err)
	}
	if len(vals) != 2 {
		return nil, nil, fmt.Errorf("uniswapv3: decodeDecreaseLiquidity: expected 2 return values, got %d", len(vals))
	}
	amount0, ok := vals[0].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv3: decodeDecreaseLiquidity: unexpected decoded type for amount0: %T", vals[0])
	}
	amount1, ok = vals[1].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv3: decodeDecreaseLiquidity: unexpected decoded type for amount1: %T", vals[1])
	}
	return amount0, amount1, nil
}

// EncodeNfpmCollect returns the calldata for collect(params).
func EncodeNfpmCollect(tokenId *big.Int, recipient *types.Address, amount0Max, amount1Max *big.Int) ([]byte, error) {
	data, err := nfpmCollectFn.EncodeCall([]any{tokenId, recipient, amount0Max, amount1Max})
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeNfpmCollect: %w", err)
	}
	return data, nil
}

// DecodeNfpmCollect decodes the return data of collect -- amount0, amount1.
func DecodeNfpmCollect(data []byte) (amount0, amount1 *big.Int, err error) {
	vals, err := nfpmCollectFn.DecodeReturn(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: decodeNfpmCollect: %w", err)
	}
	if len(vals) != 2 {
		return nil, nil, fmt.Errorf("uniswapv3: decodeNfpmCollect: expected 2 return values, got %d", len(vals))
	}
	amount0, ok := vals[0].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv3: decodeNfpmCollect: unexpected decoded type for amount0: %T", vals[0])
	}
	amount1, ok = vals[1].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv3: decodeNfpmCollect: unexpected decoded type for amount1: %T", vals[1])
	}
	return amount0, amount1, nil
}

// EncodeNfpmBurn returns the calldata for burn(tokenId).
func EncodeNfpmBurn(tokenId *big.Int) ([]byte, error) {
	data, err := nfpmBurnFn.EncodeCall(tokenId)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeNfpmBurn: %w", err)
	}
	return data, nil
}

// EncodeUnwrapWETH9 returns the calldata for unwrapWETH9(amountMinimum, recipient).
func EncodeUnwrapWETH9(amountMinimum *big.Int, recipient *types.Address) ([]byte, error) {
	data, err := unwrapWETH9Fn.EncodeCall(amountMinimum, recipient)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeUnwrapWETH9: %w", err)
	}
	return data, nil
}

// EncodeSweepToken returns the calldata for sweepToken(token, amountMinimum, recipient).
func EncodeSweepToken(token *types.Address, amountMinimum *big.Int, recipient *types.Address) ([]byte, error) {
	data, err := sweepTokenFn.EncodeCall(token, amountMinimum, recipient)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeSweepToken: %w", err)
	}
	return data, nil
}

// EncodeRefundETH returns the calldata for refundETH().
func EncodeRefundETH() []byte {
	return refundETHFn.SelectorBytes()
}

// EncodeMulticall returns the calldata for multicall(data).
func EncodeMulticall(data [][]byte) ([]byte, error) {
	calldata, err := nfpmMulticallFn.EncodeCall(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeMulticall: %w", err)
	}
	return calldata, nil
}

// DecodeMulticall decodes the return data of multicall -- each sub-call's own return data.
func DecodeMulticall(data []byte) ([][]byte, error) {
	vals, err := nfpmMulticallFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeMulticall: %w", err)
	}
	if len(vals) != 1 {
		return nil, fmt.Errorf("uniswapv3: decodeMulticall: expected 1 return value, got %d", len(vals))
	}
	elems, ok := vals[0].([]any)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeMulticall: unexpected decoded type for results: %T", vals[0])
	}
	results := make([][]byte, len(elems))
	for i, e := range elems {
		b, ok := e.([]byte)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: decodeMulticall: unexpected decoded type for results[%d]: %T", i, e)
		}
		results[i] = b
	}
	return results, nil
}

// EncodeTotalSupply returns the calldata for totalSupply().
func EncodeTotalSupply() []byte {
	return totalSupplyFn.SelectorBytes()
}

// DecodeTotalSupply decodes the return data of totalSupply.
func DecodeTotalSupply(data []byte) (*big.Int, error) {
	n, err := totalSupplyFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeTotalSupply: %w", err)
	}
	return n, nil
}

// EncodeTokenOfOwnerByIndex returns the calldata for tokenOfOwnerByIndex(owner, index).
func EncodeTokenOfOwnerByIndex(owner *types.Address, index *big.Int) ([]byte, error) {
	data, err := tokenOfOwnerByIndexFn.EncodeCall(owner, index)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeTokenOfOwnerByIndex: %w", err)
	}
	return data, nil
}

// DecodeTokenOfOwnerByIndex decodes the return data of tokenOfOwnerByIndex.
func DecodeTokenOfOwnerByIndex(data []byte) (*big.Int, error) {
	n, err := tokenOfOwnerByIndexFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeTokenOfOwnerByIndex: %w", err)
	}
	return n, nil
}

// EncodeTokenByIndex returns the calldata for tokenByIndex(index).
func EncodeTokenByIndex(index *big.Int) ([]byte, error) {
	data, err := tokenByIndexFn.EncodeCall(index)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeTokenByIndex: %w", err)
	}
	return data, nil
}

// DecodeTokenByIndex decodes the return data of tokenByIndex.
func DecodeTokenByIndex(data []byte) (*big.Int, error) {
	n, err := tokenByIndexFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeTokenByIndex: %w", err)
	}
	return n, nil
}

// EncodePermitTypehash returns the calldata for PERMIT_TYPEHASH().
func EncodePermitTypehash() []byte {
	return permitTypehashFn.SelectorBytes()
}

// DecodePermitTypehash decodes the return data of PERMIT_TYPEHASH.
func DecodePermitTypehash(data []byte) ([32]byte, error) {
	b, err := permitTypehashFn.DecodeSingleReturn[[]byte](data)
	if err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv3: decodePermitTypehash: %w", err)
	}
	var out [32]byte
	copy(out[:], b)
	return out, nil
}

// EncodeDomainSeparator returns the calldata for DOMAIN_SEPARATOR().
func EncodeDomainSeparator() []byte {
	return domainSeparatorFn.SelectorBytes()
}

// DecodeDomainSeparator decodes the return data of DOMAIN_SEPARATOR.
func DecodeDomainSeparator(data []byte) ([32]byte, error) {
	b, err := domainSeparatorFn.DecodeSingleReturn[[]byte](data)
	if err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv3: decodeDomainSeparator: %w", err)
	}
	var out [32]byte
	copy(out[:], b)
	return out, nil
}

// EncodeNFTPermit returns the calldata for permit(spender, tokenId, deadline, v, r, s). sig's
// recovery id is encoded in its legacy (27/28) form, as the permit's ecrecover check expects.
func EncodeNFTPermit(spender *types.Address, tokenId, deadline *big.Int, sig *types.Signature) ([]byte, error) {
	legacy := sig.LegacyBytes()
	data, err := nftPermitFn.EncodeCall(spender, tokenId, deadline, legacy[64], legacy[:32], legacy[32:64])
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeNFTPermit: %w", err)
	}
	return data, nil
}
