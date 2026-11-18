package v4

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// IStateViewImmutables mirrors IImmutableState, StateView's only state.
type IStateViewImmutables interface {
	PoolManager(ctx context.Context, contract *types.Address, block string) (*types.Address, error)
}

// IStateViewPool mirrors IStateView's pool-wide getters.
type IStateViewPool interface {
	GetSlot0(ctx context.Context, contract *types.Address, poolID PoolID, block string) (*Slot0, error)
	GetLiquidity(ctx context.Context, contract *types.Address, poolID PoolID, block string) (*big.Int, error)
	GetFeeGrowthGlobals(ctx context.Context, contract *types.Address, poolID PoolID, block string) (*FeeGrowthGlobals, error)
	GetFeeGrowthInside(ctx context.Context, contract *types.Address, poolID PoolID, tickLower, tickUpper *big.Int, block string) (*FeeGrowthInside, error)
}

// IStateViewTicks mirrors IStateView's per-tick getters.
type IStateViewTicks interface {
	GetTickInfo(ctx context.Context, contract *types.Address, poolID PoolID, tick *big.Int, block string) (*TickInfo, error)
	GetTickLiquidity(ctx context.Context, contract *types.Address, poolID PoolID, tick *big.Int, block string) (*TickLiquidity, error)
	GetTickFeeGrowthOutside(ctx context.Context, contract *types.Address, poolID PoolID, tick *big.Int, block string) (*TickFeeGrowthOutside, error)
	GetTickBitmap(ctx context.Context, contract *types.Address, poolID PoolID, word int16, block string) (*big.Int, error)
}

// IStateViewPositions mirrors IStateView's per-position getters; GetPositionInfoByID is the
// getPositionInfo(poolId, positionId) overload.
type IStateViewPositions interface {
	GetPositionInfo(ctx context.Context, contract *types.Address, poolID PoolID, owner *types.Address, tickLower, tickUpper *big.Int, salt [32]byte, block string) (*PositionState, error)
	GetPositionInfoByID(ctx context.Context, contract *types.Address, poolID PoolID, positionID [32]byte, block string) (*PositionState, error)
	GetPositionLiquidity(ctx context.Context, contract *types.Address, poolID PoolID, positionID [32]byte, block string) (*big.Int, error)
}

// IStateView mirrors IStateView.
type IStateView interface {
	IStateViewImmutables
	IStateViewPool
	IStateViewTicks
	IStateViewPositions
}

var _ IStateView = (*StateView)(nil)

// StateView calls any StateView contract over a Client; it holds no contract address, so one
// instance serves every network's deployment.
type StateView struct {
	c rpc.Client
}

// NewStateView returns a StateView calling StateView contracts via c.
func NewStateView(c rpc.Client) *StateView {
	return &StateView{
		c: c,
	}
}

// PoolManager returns contract's immutable PoolManager address.
func (v *StateView) PoolManager(ctx context.Context, contract *types.Address, block string) (*types.Address, error) {
	data, err := v.call(ctx, contract, EncodeStateViewPoolManager(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: stateViewPoolManager: %w", err)
	}
	addr, err := DecodeStateViewPoolManager(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: stateViewPoolManager: %w", err)
	}
	return addr, nil
}

// GetSlot0 returns poolID's sqrtPriceX96, tick, protocolFee, and lpFee.
func (v *StateView) GetSlot0(ctx context.Context, contract *types.Address, poolID PoolID, block string) (*Slot0, error) {
	calldata, err := EncodeGetSlot0(poolID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getSlot0: %w", err)
	}
	data, err := v.call(ctx, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getSlot0: %w", err)
	}
	slot0, err := DecodeGetSlot0(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getSlot0: %w", err)
	}
	return slot0, nil
}

// GetLiquidity returns poolID's current in-range liquidity.
func (v *StateView) GetLiquidity(ctx context.Context, contract *types.Address, poolID PoolID, block string) (*big.Int, error) {
	calldata, err := EncodeGetLiquidity(poolID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getLiquidity: %w", err)
	}
	data, err := v.call(ctx, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getLiquidity: %w", err)
	}
	liquidity, err := DecodeGetLiquidity(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getLiquidity: %w", err)
	}
	return liquidity, nil
}

// GetFeeGrowthGlobals returns poolID's global fee growth per unit of liquidity, X128.
func (v *StateView) GetFeeGrowthGlobals(ctx context.Context, contract *types.Address, poolID PoolID, block string) (*FeeGrowthGlobals, error) {
	calldata, err := EncodeGetFeeGrowthGlobals(poolID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getFeeGrowthGlobals: %w", err)
	}
	data, err := v.call(ctx, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getFeeGrowthGlobals: %w", err)
	}
	globals, err := DecodeGetFeeGrowthGlobals(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getFeeGrowthGlobals: %w", err)
	}
	return globals, nil
}

// GetFeeGrowthInside returns poolID's up-to-date fee growth within [tickLower, tickUpper], X128.
func (v *StateView) GetFeeGrowthInside(ctx context.Context, contract *types.Address, poolID PoolID, tickLower, tickUpper *big.Int, block string) (*FeeGrowthInside, error) {
	calldata, err := EncodeGetFeeGrowthInside(poolID, tickLower, tickUpper)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getFeeGrowthInside: %w", err)
	}
	data, err := v.call(ctx, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getFeeGrowthInside: %w", err)
	}
	inside, err := DecodeGetFeeGrowthInside(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getFeeGrowthInside: %w", err)
	}
	return inside, nil
}

// GetTickInfo returns poolID's full info for tick.
func (v *StateView) GetTickInfo(ctx context.Context, contract *types.Address, poolID PoolID, tick *big.Int, block string) (*TickInfo, error) {
	calldata, err := EncodeGetTickInfo(poolID, tick)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickInfo: %w", err)
	}
	data, err := v.call(ctx, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickInfo: %w", err)
	}
	info, err := DecodeGetTickInfo(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickInfo: %w", err)
	}
	return info, nil
}

// GetTickLiquidity returns poolID's liquidityGross and liquidityNet for tick.
func (v *StateView) GetTickLiquidity(ctx context.Context, contract *types.Address, poolID PoolID, tick *big.Int, block string) (*TickLiquidity, error) {
	calldata, err := EncodeGetTickLiquidity(poolID, tick)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickLiquidity: %w", err)
	}
	data, err := v.call(ctx, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickLiquidity: %w", err)
	}
	liquidity, err := DecodeGetTickLiquidity(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickLiquidity: %w", err)
	}
	return liquidity, nil
}

// GetTickFeeGrowthOutside returns poolID's fee growth outside tick, X128.
func (v *StateView) GetTickFeeGrowthOutside(ctx context.Context, contract *types.Address, poolID PoolID, tick *big.Int, block string) (*TickFeeGrowthOutside, error) {
	calldata, err := EncodeGetTickFeeGrowthOutside(poolID, tick)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickFeeGrowthOutside: %w", err)
	}
	data, err := v.call(ctx, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickFeeGrowthOutside: %w", err)
	}
	outside, err := DecodeGetTickFeeGrowthOutside(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickFeeGrowthOutside: %w", err)
	}
	return outside, nil
}

// GetTickBitmap returns poolID's tick bitmap word at position word.
func (v *StateView) GetTickBitmap(ctx context.Context, contract *types.Address, poolID PoolID, word int16, block string) (*big.Int, error) {
	calldata, err := EncodeGetTickBitmap(poolID, word)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickBitmap: %w", err)
	}
	data, err := v.call(ctx, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickBitmap: %w", err)
	}
	bitmap, err := DecodeGetTickBitmap(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickBitmap: %w", err)
	}
	return bitmap, nil
}

// GetPositionInfo returns the position (owner, tickLower, tickUpper, salt) in poolID.
func (v *StateView) GetPositionInfo(ctx context.Context, contract *types.Address, poolID PoolID, owner *types.Address, tickLower, tickUpper *big.Int, salt [32]byte, block string) (*PositionState, error) {
	calldata, err := EncodeGetPositionInfo(poolID, owner, tickLower, tickUpper, salt)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getPositionInfo: %w", err)
	}
	data, err := v.call(ctx, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getPositionInfo: %w", err)
	}
	state, err := DecodeGetPositionInfo(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getPositionInfo: %w", err)
	}
	return state, nil
}

// GetPositionInfoByID returns the position positionID in poolID.
func (v *StateView) GetPositionInfoByID(ctx context.Context, contract *types.Address, poolID PoolID, positionID [32]byte, block string) (*PositionState, error) {
	calldata, err := EncodeGetPositionInfoByID(poolID, positionID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getPositionInfoByID: %w", err)
	}
	data, err := v.call(ctx, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getPositionInfoByID: %w", err)
	}
	state, err := DecodeGetPositionInfo(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getPositionInfoByID: %w", err)
	}
	return state, nil
}

// GetPositionLiquidity returns the liquidity of position positionID in poolID.
func (v *StateView) GetPositionLiquidity(ctx context.Context, contract *types.Address, poolID PoolID, positionID [32]byte, block string) (*big.Int, error) {
	calldata, err := EncodeStateViewGetPositionLiquidity(poolID, positionID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: stateViewGetPositionLiquidity: %w", err)
	}
	data, err := v.call(ctx, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: stateViewGetPositionLiquidity: %w", err)
	}
	liquidity, err := DecodeStateViewGetPositionLiquidity(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: stateViewGetPositionLiquidity: %w", err)
	}
	return liquidity, nil
}

// call eth_calls contract with calldata. StateView declares no custom errors, so a revert is
// returned as-is.
func (v *StateView) call(ctx context.Context, contract *types.Address, calldata []byte, block string) ([]byte, error) {
	params := map[string]any{
		"to":   contract.String(),
		"data": encoding.Hex.EncodePrefixed(calldata),
	}

	var raw string
	if err := v.c.Call(ctx, rpc.ETHCall("call", params, block, &raw)); err != nil {
		return nil, err
	}
	return encoding.Hex.Decode(raw)
}
