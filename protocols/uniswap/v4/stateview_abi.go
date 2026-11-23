package v4

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

var (
	getSlot0Fn                = abi.NewFunction("getSlot0", abi.NewTypes(abi.Bytes32), abi.NewTypes(abi.Uint160, abi.Int24, abi.Uint24, abi.Uint24))
	getTickInfoFn             = abi.NewFunction("getTickInfo", abi.NewTypes(abi.Bytes32, abi.Int24), abi.NewTypes(abi.Uint128, abi.Int128, abi.Uint256, abi.Uint256))
	getTickLiquidityFn        = abi.NewFunction("getTickLiquidity", abi.NewTypes(abi.Bytes32, abi.Int24), abi.NewTypes(abi.Uint128, abi.Int128))
	getTickFeeGrowthOutsideFn = abi.NewFunction("getTickFeeGrowthOutside", abi.NewTypes(abi.Bytes32, abi.Int24), abi.NewTypes(abi.Uint256, abi.Uint256))
	getFeeGrowthGlobalsFn     = abi.NewFunction("getFeeGrowthGlobals", abi.NewTypes(abi.Bytes32), abi.NewTypes(abi.Uint256, abi.Uint256))
	getLiquidityFn            = abi.NewFunction("getLiquidity", abi.NewTypes(abi.Bytes32), abi.NewTypes(abi.Uint128))
	getTickBitmapFn           = abi.NewFunction("getTickBitmap", abi.NewTypes(abi.Bytes32, abi.Int16), abi.NewTypes(abi.Uint256))
	getFeeGrowthInsideFn      = abi.NewFunction("getFeeGrowthInside", abi.NewTypes(abi.Bytes32, abi.Int24, abi.Int24), abi.NewTypes(abi.Uint256, abi.Uint256))

	// getPositionInfoFn and getPositionInfoByIDFn are IStateView's two getPositionInfo overloads.
	getPositionInfoFn     = abi.NewFunction("getPositionInfo", abi.NewTypes(abi.Bytes32, abi.Address, abi.Int24, abi.Int24, abi.Bytes32), abi.NewTypes(abi.Uint128, abi.Uint256, abi.Uint256))
	getPositionInfoByIDFn = abi.NewFunction("getPositionInfo", abi.NewTypes(abi.Bytes32, abi.Bytes32), abi.NewTypes(abi.Uint128, abi.Uint256, abi.Uint256))

	// StateView-prefixed since PositionManager has its own getPositionLiquidity and poolManager bindings.
	stateViewGetPositionLiquidityFn = abi.NewFunction("getPositionLiquidity", abi.NewTypes(abi.Bytes32, abi.Bytes32), abi.NewTypes(abi.Uint128))
	stateViewPoolManagerFn          = abi.NewFunction("poolManager", nil, abi.NewTypes(abi.Address))
)

// EncodeGetSlot0 returns the calldata for getSlot0(poolId).
func EncodeGetSlot0(poolID PoolID) ([]byte, error) {
	data, err := getSlot0Fn.EncodeCall(poolID[:])
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeGetSlot0: %w", err)
	}
	return data, nil
}

// DecodeGetSlot0 decodes the return data of getSlot0.
func DecodeGetSlot0(data []byte) (*Slot0, error) {
	vals, err := getSlot0Fn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeGetSlot0: %w", err)
	}
	if len(vals) != 4 {
		return nil, fmt.Errorf("uniswapv4: decodeGetSlot0: expected 4 return values, got %d", len(vals))
	}
	sqrtPriceX96, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeGetSlot0: unexpected decoded type for sqrtPriceX96: %T", vals[0])
	}
	tick, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeGetSlot0: unexpected decoded type for tick: %T", vals[1])
	}
	protocolFee, ok := vals[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeGetSlot0: unexpected decoded type for protocolFee: %T", vals[2])
	}
	lpFee, ok := vals[3].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeGetSlot0: unexpected decoded type for lpFee: %T", vals[3])
	}
	return NewSlot0(sqrtPriceX96, tick, protocolFee, lpFee), nil
}

// EncodeGetTickInfo returns the calldata for getTickInfo(poolId, tick).
func EncodeGetTickInfo(poolID PoolID, tick *big.Int) ([]byte, error) {
	data, err := getTickInfoFn.EncodeCall(poolID[:], tick)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeGetTickInfo: %w", err)
	}
	return data, nil
}

// DecodeGetTickInfo decodes the return data of getTickInfo.
func DecodeGetTickInfo(data []byte) (*TickInfo, error) {
	vals, err := getTickInfoFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeGetTickInfo: %w", err)
	}
	if len(vals) != 4 {
		return nil, fmt.Errorf("uniswapv4: decodeGetTickInfo: expected 4 return values, got %d", len(vals))
	}
	liquidityGross, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeGetTickInfo: unexpected decoded type for liquidityGross: %T", vals[0])
	}
	liquidityNet, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeGetTickInfo: unexpected decoded type for liquidityNet: %T", vals[1])
	}
	feeGrowthOutside0X128, ok := vals[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeGetTickInfo: unexpected decoded type for feeGrowthOutside0X128: %T", vals[2])
	}
	feeGrowthOutside1X128, ok := vals[3].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeGetTickInfo: unexpected decoded type for feeGrowthOutside1X128: %T", vals[3])
	}
	return NewTickInfo(liquidityGross, liquidityNet, feeGrowthOutside0X128, feeGrowthOutside1X128), nil
}

// EncodeGetTickLiquidity returns the calldata for getTickLiquidity(poolId, tick).
func EncodeGetTickLiquidity(poolID PoolID, tick *big.Int) ([]byte, error) {
	data, err := getTickLiquidityFn.EncodeCall(poolID[:], tick)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeGetTickLiquidity: %w", err)
	}
	return data, nil
}

// DecodeGetTickLiquidity decodes the return data of getTickLiquidity.
func DecodeGetTickLiquidity(data []byte) (*TickLiquidity, error) {
	vals, err := getTickLiquidityFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeGetTickLiquidity: %w", err)
	}
	if len(vals) != 2 {
		return nil, fmt.Errorf("uniswapv4: decodeGetTickLiquidity: expected 2 return values, got %d", len(vals))
	}
	liquidityGross, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeGetTickLiquidity: unexpected decoded type for liquidityGross: %T", vals[0])
	}
	liquidityNet, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeGetTickLiquidity: unexpected decoded type for liquidityNet: %T", vals[1])
	}
	return NewTickLiquidity(liquidityGross, liquidityNet), nil
}

// EncodeGetTickFeeGrowthOutside returns the calldata for getTickFeeGrowthOutside(poolId, tick).
func EncodeGetTickFeeGrowthOutside(poolID PoolID, tick *big.Int) ([]byte, error) {
	data, err := getTickFeeGrowthOutsideFn.EncodeCall(poolID[:], tick)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeGetTickFeeGrowthOutside: %w", err)
	}
	return data, nil
}

// DecodeGetTickFeeGrowthOutside decodes the return data of getTickFeeGrowthOutside.
func DecodeGetTickFeeGrowthOutside(data []byte) (*TickFeeGrowthOutside, error) {
	vals, err := getTickFeeGrowthOutsideFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeGetTickFeeGrowthOutside: %w", err)
	}
	if len(vals) != 2 {
		return nil, fmt.Errorf("uniswapv4: decodeGetTickFeeGrowthOutside: expected 2 return values, got %d", len(vals))
	}
	feeGrowthOutside0X128, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeGetTickFeeGrowthOutside: unexpected decoded type for feeGrowthOutside0X128: %T", vals[0])
	}
	feeGrowthOutside1X128, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeGetTickFeeGrowthOutside: unexpected decoded type for feeGrowthOutside1X128: %T", vals[1])
	}
	return NewTickFeeGrowthOutside(feeGrowthOutside0X128, feeGrowthOutside1X128), nil
}

// EncodeGetFeeGrowthGlobals returns the calldata for getFeeGrowthGlobals(poolId).
func EncodeGetFeeGrowthGlobals(poolID PoolID) ([]byte, error) {
	data, err := getFeeGrowthGlobalsFn.EncodeCall(poolID[:])
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeGetFeeGrowthGlobals: %w", err)
	}
	return data, nil
}

// DecodeGetFeeGrowthGlobals decodes the return data of getFeeGrowthGlobals.
func DecodeGetFeeGrowthGlobals(data []byte) (*FeeGrowthGlobals, error) {
	vals, err := getFeeGrowthGlobalsFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeGetFeeGrowthGlobals: %w", err)
	}
	if len(vals) != 2 {
		return nil, fmt.Errorf("uniswapv4: decodeGetFeeGrowthGlobals: expected 2 return values, got %d", len(vals))
	}
	feeGrowthGlobal0, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeGetFeeGrowthGlobals: unexpected decoded type for feeGrowthGlobal0: %T", vals[0])
	}
	feeGrowthGlobal1, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeGetFeeGrowthGlobals: unexpected decoded type for feeGrowthGlobal1: %T", vals[1])
	}
	return NewFeeGrowthGlobals(feeGrowthGlobal0, feeGrowthGlobal1), nil
}

// EncodeGetLiquidity returns the calldata for getLiquidity(poolId).
func EncodeGetLiquidity(poolID PoolID) ([]byte, error) {
	data, err := getLiquidityFn.EncodeCall(poolID[:])
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeGetLiquidity: %w", err)
	}
	return data, nil
}

// DecodeGetLiquidity decodes the return data of getLiquidity: the pool's current in-range liquidity.
func DecodeGetLiquidity(data []byte) (*big.Int, error) {
	liquidity, err := getLiquidityFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeGetLiquidity: %w", err)
	}
	return liquidity, nil
}

// EncodeGetTickBitmap returns the calldata for getTickBitmap(poolId, word), where word is (tick / tickSpacing) >> 8.
func EncodeGetTickBitmap(poolID PoolID, word int16) ([]byte, error) {
	data, err := getTickBitmapFn.EncodeCall(poolID[:], word)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeGetTickBitmap: %w", err)
	}
	return data, nil
}

// DecodeGetTickBitmap decodes the return data of getTickBitmap.
func DecodeGetTickBitmap(data []byte) (*big.Int, error) {
	bitmap, err := getTickBitmapFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeGetTickBitmap: %w", err)
	}
	return bitmap, nil
}

// EncodeGetPositionInfo returns the calldata for getPositionInfo(poolId, owner, tickLower, tickUpper, salt).
func EncodeGetPositionInfo(poolID PoolID, owner *types.Address, tickLower, tickUpper *big.Int, salt [32]byte) ([]byte, error) {
	data, err := getPositionInfoFn.EncodeCall(poolID[:], owner, tickLower, tickUpper, salt[:])
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeGetPositionInfo: %w", err)
	}
	return data, nil
}

// DecodeGetPositionInfo decodes the return data of either getPositionInfo overload.
func DecodeGetPositionInfo(data []byte) (*PositionState, error) {
	vals, err := getPositionInfoFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeGetPositionInfo: %w", err)
	}
	if len(vals) != 3 {
		return nil, fmt.Errorf("uniswapv4: decodeGetPositionInfo: expected 3 return values, got %d", len(vals))
	}
	liquidity, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeGetPositionInfo: unexpected decoded type for liquidity: %T", vals[0])
	}
	feeGrowthInside0LastX128, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeGetPositionInfo: unexpected decoded type for feeGrowthInside0LastX128: %T", vals[1])
	}
	feeGrowthInside1LastX128, ok := vals[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeGetPositionInfo: unexpected decoded type for feeGrowthInside1LastX128: %T", vals[2])
	}
	return NewPositionState(liquidity, feeGrowthInside0LastX128, feeGrowthInside1LastX128), nil
}

// EncodeGetPositionInfoByID returns the calldata for getPositionInfo(poolId, positionId); see CalculatePositionKey for positionId.
func EncodeGetPositionInfoByID(poolID PoolID, positionID [32]byte) ([]byte, error) {
	data, err := getPositionInfoByIDFn.EncodeCall(poolID[:], positionID[:])
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeGetPositionInfoByID: %w", err)
	}
	return data, nil
}

// EncodeStateViewGetPositionLiquidity returns the calldata for getPositionLiquidity(poolId, positionId).
func EncodeStateViewGetPositionLiquidity(poolID PoolID, positionID [32]byte) ([]byte, error) {
	data, err := stateViewGetPositionLiquidityFn.EncodeCall(poolID[:], positionID[:])
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeStateViewGetPositionLiquidity: %w", err)
	}
	return data, nil
}

// DecodeStateViewGetPositionLiquidity decodes the return data of getPositionLiquidity.
func DecodeStateViewGetPositionLiquidity(data []byte) (*big.Int, error) {
	liquidity, err := stateViewGetPositionLiquidityFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeStateViewGetPositionLiquidity: %w", err)
	}
	return liquidity, nil
}

// EncodeGetFeeGrowthInside returns the calldata for getFeeGrowthInside(poolId, tickLower, tickUpper).
func EncodeGetFeeGrowthInside(poolID PoolID, tickLower, tickUpper *big.Int) ([]byte, error) {
	data, err := getFeeGrowthInsideFn.EncodeCall(poolID[:], tickLower, tickUpper)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeGetFeeGrowthInside: %w", err)
	}
	return data, nil
}

// DecodeGetFeeGrowthInside decodes the return data of getFeeGrowthInside.
func DecodeGetFeeGrowthInside(data []byte) (*FeeGrowthInside, error) {
	vals, err := getFeeGrowthInsideFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeGetFeeGrowthInside: %w", err)
	}
	if len(vals) != 2 {
		return nil, fmt.Errorf("uniswapv4: decodeGetFeeGrowthInside: expected 2 return values, got %d", len(vals))
	}
	feeGrowthInside0X128, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeGetFeeGrowthInside: unexpected decoded type for feeGrowthInside0X128: %T", vals[0])
	}
	feeGrowthInside1X128, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeGetFeeGrowthInside: unexpected decoded type for feeGrowthInside1X128: %T", vals[1])
	}
	return NewFeeGrowthInside(feeGrowthInside0X128, feeGrowthInside1X128), nil
}

// EncodeStateViewPoolManager returns the calldata for poolManager().
func EncodeStateViewPoolManager() []byte {
	return stateViewPoolManagerFn.SelectorBytes()
}

// DecodeStateViewPoolManager decodes the return data of poolManager.
func DecodeStateViewPoolManager(data []byte) (*types.Address, error) {
	addr, err := stateViewPoolManagerFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeStateViewPoolManager: %w", err)
	}
	return addr, nil
}
