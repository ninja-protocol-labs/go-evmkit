package v4

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Slot0 is getSlot0's return value: a pool's packed Slot0 unpacked into its four fields.
type Slot0 struct {
	SqrtPriceX96 *big.Int
	Tick         *big.Int
	ProtocolFee  *big.Int
	LPFee        *big.Int
}

func NewSlot0(sqrtPriceX96, tick, protocolFee, lpFee *big.Int) *Slot0 {
	return &Slot0{
		SqrtPriceX96: sqrtPriceX96,
		Tick:         tick,
		ProtocolFee:  protocolFee,
		LPFee:        lpFee,
	}
}

// TickInfo is getTickInfo's return value, mirroring Pool.TickInfo.
type TickInfo struct {
	LiquidityGross        *big.Int
	LiquidityNet          *big.Int
	FeeGrowthOutside0X128 *big.Int
	FeeGrowthOutside1X128 *big.Int
}

func NewTickInfo(liquidityGross, liquidityNet, feeGrowthOutside0X128, feeGrowthOutside1X128 *big.Int) *TickInfo {
	return &TickInfo{
		LiquidityGross:        liquidityGross,
		LiquidityNet:          liquidityNet,
		FeeGrowthOutside0X128: feeGrowthOutside0X128,
		FeeGrowthOutside1X128: feeGrowthOutside1X128,
	}
}

// TickLiquidity is getTickLiquidity's return value.
type TickLiquidity struct {
	LiquidityGross *big.Int
	LiquidityNet   *big.Int
}

func NewTickLiquidity(liquidityGross, liquidityNet *big.Int) *TickLiquidity {
	return &TickLiquidity{
		LiquidityGross: liquidityGross,
		LiquidityNet:   liquidityNet,
	}
}

// TickFeeGrowthOutside is getTickFeeGrowthOutside's return value.
type TickFeeGrowthOutside struct {
	FeeGrowthOutside0X128 *big.Int
	FeeGrowthOutside1X128 *big.Int
}

func NewTickFeeGrowthOutside(feeGrowthOutside0X128, feeGrowthOutside1X128 *big.Int) *TickFeeGrowthOutside {
	return &TickFeeGrowthOutside{
		FeeGrowthOutside0X128: feeGrowthOutside0X128,
		FeeGrowthOutside1X128: feeGrowthOutside1X128,
	}
}

// FeeGrowthGlobals is getFeeGrowthGlobals's return value; both values are X128 fixed point.
type FeeGrowthGlobals struct {
	FeeGrowthGlobal0 *big.Int
	FeeGrowthGlobal1 *big.Int
}

func NewFeeGrowthGlobals(feeGrowthGlobal0, feeGrowthGlobal1 *big.Int) *FeeGrowthGlobals {
	return &FeeGrowthGlobals{
		FeeGrowthGlobal0: feeGrowthGlobal0,
		FeeGrowthGlobal1: feeGrowthGlobal1,
	}
}

// PositionState is getPositionInfo's return value, mirroring Position.State; PositionInfo is PositionManager's packed per-token info.
type PositionState struct {
	Liquidity                *big.Int
	FeeGrowthInside0LastX128 *big.Int
	FeeGrowthInside1LastX128 *big.Int
}

func NewPositionState(liquidity, feeGrowthInside0LastX128, feeGrowthInside1LastX128 *big.Int) *PositionState {
	return &PositionState{
		Liquidity:                liquidity,
		FeeGrowthInside0LastX128: feeGrowthInside0LastX128,
		FeeGrowthInside1LastX128: feeGrowthInside1LastX128,
	}
}

// FeeGrowthInside is getFeeGrowthInside's return value: up-to-date fee growth within a tick range.
type FeeGrowthInside struct {
	FeeGrowthInside0X128 *big.Int
	FeeGrowthInside1X128 *big.Int
}

func NewFeeGrowthInside(feeGrowthInside0X128, feeGrowthInside1X128 *big.Int) *FeeGrowthInside {
	return &FeeGrowthInside{
		FeeGrowthInside0X128: feeGrowthInside0X128,
		FeeGrowthInside1X128: feeGrowthInside1X128,
	}
}

// CalculatePositionKey mirrors Position.calculatePositionKey: keccak256(abi.encodePacked(owner, tickLower, tickUpper, salt)) with int24 ticks packed as 3 bytes.
func CalculatePositionKey(owner *types.Address, tickLower, tickUpper *big.Int, salt [32]byte) ([32]byte, error) {
	if owner == nil {
		return [32]byte{}, fmt.Errorf("uniswapv4: calculatePositionKey: owner must not be nil")
	}
	minInt24 := big.NewInt(-1 << 23)
	maxInt24 := big.NewInt(1<<23 - 1)

	packed := make([]byte, 0, 58)
	packed = append(packed, owner.Bytes()...)
	for _, tick := range []*big.Int{tickLower, tickUpper} {
		if tick == nil || tick.Cmp(minInt24) < 0 || tick.Cmp(maxInt24) > 0 {
			return [32]byte{}, fmt.Errorf("uniswapv4: calculatePositionKey: tick %v out of int24 range", tick)
		}
		v := tick.Int64()
		if v < 0 {
			v += 1 << 24
		}
		packed = append(packed, byte(v>>16), byte(v>>8), byte(v))
	}
	packed = append(packed, salt[:]...)

	var key [32]byte
	copy(key[:], core.Keccak256(packed).Bytes())
	return key, nil
}

// PoolManagerRequest is poolManager()'s input on Contract, shared by every contract exposing IImmutableState.
type PoolManagerRequest struct {
	Contract *types.Address
}

func NewPoolManagerRequest(contract *types.Address) *PoolManagerRequest {
	return &PoolManagerRequest{
		Contract: contract,
	}
}

func (r *PoolManagerRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetSlot0Request is getSlot0(poolId)'s input on Contract.
type GetSlot0Request struct {
	Contract *types.Address
	PoolID   PoolID
}

func NewGetSlot0Request(contract *types.Address, poolID PoolID) *GetSlot0Request {
	return &GetSlot0Request{
		Contract: contract,
		PoolID:   poolID,
	}
}

func (r *GetSlot0Request) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetLiquidityRequest is getLiquidity(poolId)'s input on Contract.
type GetLiquidityRequest struct {
	Contract *types.Address
	PoolID   PoolID
}

func NewGetLiquidityRequest(contract *types.Address, poolID PoolID) *GetLiquidityRequest {
	return &GetLiquidityRequest{
		Contract: contract,
		PoolID:   poolID,
	}
}

func (r *GetLiquidityRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetFeeGrowthGlobalsRequest is getFeeGrowthGlobals(poolId)'s input on Contract.
type GetFeeGrowthGlobalsRequest struct {
	Contract *types.Address
	PoolID   PoolID
}

func NewGetFeeGrowthGlobalsRequest(contract *types.Address, poolID PoolID) *GetFeeGrowthGlobalsRequest {
	return &GetFeeGrowthGlobalsRequest{
		Contract: contract,
		PoolID:   poolID,
	}
}

func (r *GetFeeGrowthGlobalsRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetFeeGrowthInsideRequest is getFeeGrowthInside(poolId, tickLower, tickUpper)'s input on Contract.
type GetFeeGrowthInsideRequest struct {
	Contract  *types.Address
	PoolID    PoolID
	TickLower *big.Int
	TickUpper *big.Int
}

func NewGetFeeGrowthInsideRequest(contract *types.Address, poolID PoolID, tickLower, tickUpper *big.Int) *GetFeeGrowthInsideRequest {
	return &GetFeeGrowthInsideRequest{
		Contract:  contract,
		PoolID:    poolID,
		TickLower: tickLower,
		TickUpper: tickUpper,
	}
}

func (r *GetFeeGrowthInsideRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.TickLower == nil {
		return fmt.Errorf("%w: tickLower is nil", bind.ErrInvalidRequest)
	}

	if r.TickUpper == nil {
		return fmt.Errorf("%w: tickUpper is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetTickInfoRequest is getTickInfo(poolId, tick)'s input on Contract.
type GetTickInfoRequest struct {
	Contract *types.Address
	PoolID   PoolID
	Tick     *big.Int
}

func NewGetTickInfoRequest(contract *types.Address, poolID PoolID, tick *big.Int) *GetTickInfoRequest {
	return &GetTickInfoRequest{
		Contract: contract,
		PoolID:   poolID,
		Tick:     tick,
	}
}

func (r *GetTickInfoRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Tick == nil {
		return fmt.Errorf("%w: tick is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetTickLiquidityRequest is getTickLiquidity(poolId, tick)'s input on Contract.
type GetTickLiquidityRequest struct {
	Contract *types.Address
	PoolID   PoolID
	Tick     *big.Int
}

func NewGetTickLiquidityRequest(contract *types.Address, poolID PoolID, tick *big.Int) *GetTickLiquidityRequest {
	return &GetTickLiquidityRequest{
		Contract: contract,
		PoolID:   poolID,
		Tick:     tick,
	}
}

func (r *GetTickLiquidityRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Tick == nil {
		return fmt.Errorf("%w: tick is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetTickFeeGrowthOutsideRequest is getTickFeeGrowthOutside(poolId, tick)'s input on Contract.
type GetTickFeeGrowthOutsideRequest struct {
	Contract *types.Address
	PoolID   PoolID
	Tick     *big.Int
}

func NewGetTickFeeGrowthOutsideRequest(contract *types.Address, poolID PoolID, tick *big.Int) *GetTickFeeGrowthOutsideRequest {
	return &GetTickFeeGrowthOutsideRequest{
		Contract: contract,
		PoolID:   poolID,
		Tick:     tick,
	}
}

func (r *GetTickFeeGrowthOutsideRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Tick == nil {
		return fmt.Errorf("%w: tick is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetTickBitmapRequest is getTickBitmap(poolId, tick)'s input on Contract; WordPosition is (tick / tickSpacing) >> 8.
type GetTickBitmapRequest struct {
	Contract     *types.Address
	PoolID       PoolID
	WordPosition int16
}

func NewGetTickBitmapRequest(contract *types.Address, poolID PoolID, wordPosition int16) *GetTickBitmapRequest {
	return &GetTickBitmapRequest{
		Contract:     contract,
		PoolID:       poolID,
		WordPosition: wordPosition,
	}
}

func (r *GetTickBitmapRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetPositionInfoRequest is getPositionInfo(poolId, owner, tickLower, tickUpper, salt)'s input on Contract.
type GetPositionInfoRequest struct {
	Contract  *types.Address
	PoolID    PoolID
	Owner     *types.Address
	TickLower *big.Int
	TickUpper *big.Int
	Salt      [32]byte
}

func NewGetPositionInfoRequest(contract *types.Address, poolID PoolID, owner *types.Address, tickLower, tickUpper *big.Int, salt [32]byte) *GetPositionInfoRequest {
	return &GetPositionInfoRequest{
		Contract:  contract,
		PoolID:    poolID,
		Owner:     owner,
		TickLower: tickLower,
		TickUpper: tickUpper,
		Salt:      salt,
	}
}

func (r *GetPositionInfoRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Owner == nil {
		return fmt.Errorf("%w: owner is nil", bind.ErrInvalidRequest)
	}

	if r.TickLower == nil {
		return fmt.Errorf("%w: tickLower is nil", bind.ErrInvalidRequest)
	}

	if r.TickUpper == nil {
		return fmt.Errorf("%w: tickUpper is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetPositionInfoByIDRequest is getPositionInfo(poolId, positionId)'s input on Contract.
type GetPositionInfoByIDRequest struct {
	Contract   *types.Address
	PoolID     PoolID
	PositionID [32]byte
}

func NewGetPositionInfoByIDRequest(contract *types.Address, poolID PoolID, positionID [32]byte) *GetPositionInfoByIDRequest {
	return &GetPositionInfoByIDRequest{
		Contract:   contract,
		PoolID:     poolID,
		PositionID: positionID,
	}
}

func (r *GetPositionInfoByIDRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// StateViewGetPositionLiquidityRequest is StateView's getPositionLiquidity(poolId, positionId)'s input on Contract.
type StateViewGetPositionLiquidityRequest struct {
	Contract   *types.Address
	PoolID     PoolID
	PositionID [32]byte
}

func NewStateViewGetPositionLiquidityRequest(contract *types.Address, poolID PoolID, positionID [32]byte) *StateViewGetPositionLiquidityRequest {
	return &StateViewGetPositionLiquidityRequest{
		Contract:   contract,
		PoolID:     poolID,
		PositionID: positionID,
	}
}

func (r *StateViewGetPositionLiquidityRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}
