package v4

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

// IStateViewImmutables mirrors IImmutableState, StateView's only state.
type IStateViewImmutables interface {
	PoolManager(ctx context.Context, opts *bind.CallOptions, req *PoolManagerRequest) (*types.Address, error)
}

// IStateViewPool mirrors IStateView's pool-wide getters.
type IStateViewPool interface {
	GetSlot0(ctx context.Context, opts *bind.CallOptions, req *GetSlot0Request) (*Slot0, error)
	GetLiquidity(ctx context.Context, opts *bind.CallOptions, req *GetLiquidityRequest) (*big.Int, error)
	GetFeeGrowthGlobals(ctx context.Context, opts *bind.CallOptions, req *GetFeeGrowthGlobalsRequest) (*FeeGrowthGlobals, error)
	GetFeeGrowthInside(ctx context.Context, opts *bind.CallOptions, req *GetFeeGrowthInsideRequest) (*FeeGrowthInside, error)
}

// IStateViewTicks mirrors IStateView's per-tick getters.
type IStateViewTicks interface {
	GetTickInfo(ctx context.Context, opts *bind.CallOptions, req *GetTickInfoRequest) (*TickInfo, error)
	GetTickLiquidity(ctx context.Context, opts *bind.CallOptions, req *GetTickLiquidityRequest) (*TickLiquidity, error)
	GetTickFeeGrowthOutside(ctx context.Context, opts *bind.CallOptions, req *GetTickFeeGrowthOutsideRequest) (*TickFeeGrowthOutside, error)
	GetTickBitmap(ctx context.Context, opts *bind.CallOptions, req *GetTickBitmapRequest) (*big.Int, error)
}

// IStateViewPositions mirrors IStateView's per-position getters; GetPositionInfoByID is the getPositionInfo(poolId, positionId) overload.
type IStateViewPositions interface {
	GetPositionInfo(ctx context.Context, opts *bind.CallOptions, req *GetPositionInfoRequest) (*PositionState, error)
	GetPositionInfoByID(ctx context.Context, opts *bind.CallOptions, req *GetPositionInfoByIDRequest) (*PositionState, error)
	GetPositionLiquidity(ctx context.Context, opts *bind.CallOptions, req *StateViewGetPositionLiquidityRequest) (*big.Int, error)
}

// IStateView mirrors IStateView.
type IStateView interface {
	IStateViewImmutables
	IStateViewPool
	IStateViewTicks
	IStateViewPositions
}

var _ IStateView = (*StateView)(nil)

// StateView calls any StateView contract over a Client; it declares no custom errors, so RPC errors pass through unwrapped.
type StateView struct {
	c rpc.Client
}

// NewStateView returns a StateView calling StateView contracts via cli.
func NewStateView(cli rpc.Client) *StateView {
	return &StateView{
		c: cli,
	}
}

// PoolManager returns req.Contract's immutable PoolManager address.
func (v *StateView) PoolManager(ctx context.Context, opts *bind.CallOptions, req *PoolManagerRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: stateViewPoolManager: %w", err)
	}

	data, err := bind.Call(ctx, v.c, opts, req.Contract, EncodeStateViewPoolManager())
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: stateViewPoolManager: %w", err)
	}

	out, err := DecodeStateViewPoolManager(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: stateViewPoolManager: %w", err)
	}

	return out, nil
}

// GetSlot0 returns req.PoolID's sqrtPriceX96, tick, protocolFee and lpFee; an uninitialized pool reads as all zero.
func (v *StateView) GetSlot0(ctx context.Context, opts *bind.CallOptions, req *GetSlot0Request) (*Slot0, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: getSlot0: %w", err)
	}

	calldata, err := EncodeGetSlot0(req.PoolID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getSlot0: %w", err)
	}

	data, err := bind.Call(ctx, v.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getSlot0: %w", err)
	}

	out, err := DecodeGetSlot0(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getSlot0: %w", err)
	}

	return out, nil
}

// GetLiquidity returns req.PoolID's current in-range liquidity.
func (v *StateView) GetLiquidity(ctx context.Context, opts *bind.CallOptions, req *GetLiquidityRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: getLiquidity: %w", err)
	}

	calldata, err := EncodeGetLiquidity(req.PoolID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getLiquidity: %w", err)
	}

	data, err := bind.Call(ctx, v.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getLiquidity: %w", err)
	}

	out, err := DecodeGetLiquidity(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getLiquidity: %w", err)
	}

	return out, nil
}

// GetFeeGrowthGlobals returns req.PoolID's global fee growth per unit of liquidity, X128.
func (v *StateView) GetFeeGrowthGlobals(ctx context.Context, opts *bind.CallOptions, req *GetFeeGrowthGlobalsRequest) (*FeeGrowthGlobals, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: getFeeGrowthGlobals: %w", err)
	}

	calldata, err := EncodeGetFeeGrowthGlobals(req.PoolID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getFeeGrowthGlobals: %w", err)
	}

	data, err := bind.Call(ctx, v.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getFeeGrowthGlobals: %w", err)
	}

	out, err := DecodeGetFeeGrowthGlobals(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getFeeGrowthGlobals: %w", err)
	}

	return out, nil
}

// GetFeeGrowthInside returns req.PoolID's up-to-date fee growth within [req.TickLower, req.TickUpper], X128.
func (v *StateView) GetFeeGrowthInside(ctx context.Context, opts *bind.CallOptions, req *GetFeeGrowthInsideRequest) (*FeeGrowthInside, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: getFeeGrowthInside: %w", err)
	}

	calldata, err := EncodeGetFeeGrowthInside(req.PoolID, req.TickLower, req.TickUpper)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getFeeGrowthInside: %w", err)
	}

	data, err := bind.Call(ctx, v.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getFeeGrowthInside: %w", err)
	}

	out, err := DecodeGetFeeGrowthInside(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getFeeGrowthInside: %w", err)
	}

	return out, nil
}

// GetTickInfo returns req.PoolID's full info for req.Tick.
func (v *StateView) GetTickInfo(ctx context.Context, opts *bind.CallOptions, req *GetTickInfoRequest) (*TickInfo, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickInfo: %w", err)
	}

	calldata, err := EncodeGetTickInfo(req.PoolID, req.Tick)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickInfo: %w", err)
	}

	data, err := bind.Call(ctx, v.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickInfo: %w", err)
	}

	out, err := DecodeGetTickInfo(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickInfo: %w", err)
	}

	return out, nil
}

// GetTickLiquidity returns req.PoolID's liquidityGross and liquidityNet for req.Tick.
func (v *StateView) GetTickLiquidity(ctx context.Context, opts *bind.CallOptions, req *GetTickLiquidityRequest) (*TickLiquidity, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickLiquidity: %w", err)
	}

	calldata, err := EncodeGetTickLiquidity(req.PoolID, req.Tick)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickLiquidity: %w", err)
	}

	data, err := bind.Call(ctx, v.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickLiquidity: %w", err)
	}

	out, err := DecodeGetTickLiquidity(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickLiquidity: %w", err)
	}

	return out, nil
}

// GetTickFeeGrowthOutside returns req.PoolID's fee growth outside req.Tick, X128.
func (v *StateView) GetTickFeeGrowthOutside(ctx context.Context, opts *bind.CallOptions, req *GetTickFeeGrowthOutsideRequest) (*TickFeeGrowthOutside, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickFeeGrowthOutside: %w", err)
	}

	calldata, err := EncodeGetTickFeeGrowthOutside(req.PoolID, req.Tick)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickFeeGrowthOutside: %w", err)
	}

	data, err := bind.Call(ctx, v.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickFeeGrowthOutside: %w", err)
	}

	out, err := DecodeGetTickFeeGrowthOutside(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickFeeGrowthOutside: %w", err)
	}

	return out, nil
}

// GetTickBitmap returns req.PoolID's tick bitmap word at req.WordPosition, i.e. (tick / tickSpacing) >> 8.
func (v *StateView) GetTickBitmap(ctx context.Context, opts *bind.CallOptions, req *GetTickBitmapRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickBitmap: %w", err)
	}

	calldata, err := EncodeGetTickBitmap(req.PoolID, req.WordPosition)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickBitmap: %w", err)
	}

	data, err := bind.Call(ctx, v.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickBitmap: %w", err)
	}

	out, err := DecodeGetTickBitmap(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getTickBitmap: %w", err)
	}

	return out, nil
}

// GetPositionInfo returns the position (req.Owner, req.TickLower, req.TickUpper, req.Salt) in req.PoolID.
func (v *StateView) GetPositionInfo(ctx context.Context, opts *bind.CallOptions, req *GetPositionInfoRequest) (*PositionState, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: getPositionInfo: %w", err)
	}

	calldata, err := EncodeGetPositionInfo(req.PoolID, req.Owner, req.TickLower, req.TickUpper, req.Salt)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getPositionInfo: %w", err)
	}

	data, err := bind.Call(ctx, v.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getPositionInfo: %w", err)
	}

	out, err := DecodeGetPositionInfo(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getPositionInfo: %w", err)
	}

	return out, nil
}

// GetPositionInfoByID returns the position req.PositionID in req.PoolID; see CalculatePositionKey.
func (v *StateView) GetPositionInfoByID(ctx context.Context, opts *bind.CallOptions, req *GetPositionInfoByIDRequest) (*PositionState, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: getPositionInfoByID: %w", err)
	}

	calldata, err := EncodeGetPositionInfoByID(req.PoolID, req.PositionID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getPositionInfoByID: %w", err)
	}

	data, err := bind.Call(ctx, v.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getPositionInfoByID: %w", err)
	}

	out, err := DecodeGetPositionInfo(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getPositionInfoByID: %w", err)
	}

	return out, nil
}

// GetPositionLiquidity returns the liquidity of position req.PositionID in req.PoolID.
func (v *StateView) GetPositionLiquidity(ctx context.Context, opts *bind.CallOptions, req *StateViewGetPositionLiquidityRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: stateViewGetPositionLiquidity: %w", err)
	}

	calldata, err := EncodeStateViewGetPositionLiquidity(req.PoolID, req.PositionID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: stateViewGetPositionLiquidity: %w", err)
	}

	data, err := bind.Call(ctx, v.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: stateViewGetPositionLiquidity: %w", err)
	}

	out, err := DecodeStateViewGetPositionLiquidity(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: stateViewGetPositionLiquidity: %w", err)
	}

	return out, nil
}
