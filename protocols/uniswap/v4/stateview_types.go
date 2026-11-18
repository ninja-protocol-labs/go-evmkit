package v4

import (
	"fmt"
	"math/big"

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

// PositionState is getPositionInfo's return value, mirroring Position.State. Named apart from
// PositionInfo, which is PositionManager's packed per-token info.
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

// CalculatePositionKey mirrors Position.calculatePositionKey: keccak256(abi.encodePacked(owner,
// tickLower, tickUpper, salt)), 58 bytes with each int24 packed as 3 two's-complement bytes.
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
