package v3

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewSlot0(t *testing.T) {
	sqrtPriceX96 := big.NewInt(1)
	tick := big.NewInt(2)
	observationIndex := big.NewInt(3)
	observationCardinality := big.NewInt(4)
	observationCardinalityNext := big.NewInt(5)
	feeProtocol := big.NewInt(6)

	s := NewSlot0(sqrtPriceX96, tick, observationIndex, observationCardinality, observationCardinalityNext, feeProtocol, true)
	require.Same(t, sqrtPriceX96, s.SqrtPriceX96)
	require.Same(t, tick, s.Tick)
	require.Same(t, observationIndex, s.ObservationIndex)
	require.Same(t, observationCardinality, s.ObservationCardinality)
	require.Same(t, observationCardinalityNext, s.ObservationCardinalityNext)
	require.Same(t, feeProtocol, s.FeeProtocol)
	require.True(t, s.Unlocked)
}

func TestNewProtocolFees(t *testing.T) {
	token0 := big.NewInt(1)
	token1 := big.NewInt(2)

	pf := NewProtocolFees(token0, token1)
	require.Same(t, token0, pf.Token0)
	require.Same(t, token1, pf.Token1)
}

func TestNewTickInfo(t *testing.T) {
	liquidityGross := big.NewInt(1)
	liquidityNet := big.NewInt(2)
	feeGrowthOutside0X128 := big.NewInt(3)
	feeGrowthOutside1X128 := big.NewInt(4)
	tickCumulativeOutside := big.NewInt(5)
	secondsPerLiquidityOutsideX128 := big.NewInt(6)
	secondsOutside := big.NewInt(7)

	ti := NewTickInfo(liquidityGross, liquidityNet, feeGrowthOutside0X128, feeGrowthOutside1X128, tickCumulativeOutside, secondsPerLiquidityOutsideX128, secondsOutside, true)
	require.Same(t, liquidityGross, ti.LiquidityGross)
	require.Same(t, liquidityNet, ti.LiquidityNet)
	require.Same(t, feeGrowthOutside0X128, ti.FeeGrowthOutside0X128)
	require.Same(t, feeGrowthOutside1X128, ti.FeeGrowthOutside1X128)
	require.Same(t, tickCumulativeOutside, ti.TickCumulativeOutside)
	require.Same(t, secondsPerLiquidityOutsideX128, ti.SecondsPerLiquidityOutsideX128)
	require.Same(t, secondsOutside, ti.SecondsOutside)
	require.True(t, ti.Initialized)
}

func TestNewPositionInfo(t *testing.T) {
	liquidity := big.NewInt(1)
	feeGrowthInside0LastX128 := big.NewInt(2)
	feeGrowthInside1LastX128 := big.NewInt(3)
	tokensOwed0 := big.NewInt(4)
	tokensOwed1 := big.NewInt(5)

	pi := NewPositionInfo(liquidity, feeGrowthInside0LastX128, feeGrowthInside1LastX128, tokensOwed0, tokensOwed1)
	require.Same(t, liquidity, pi.Liquidity)
	require.Same(t, feeGrowthInside0LastX128, pi.FeeGrowthInside0LastX128)
	require.Same(t, feeGrowthInside1LastX128, pi.FeeGrowthInside1LastX128)
	require.Same(t, tokensOwed0, pi.TokensOwed0)
	require.Same(t, tokensOwed1, pi.TokensOwed1)
}

func TestNewObservationInfo(t *testing.T) {
	blockTimestamp := big.NewInt(1)
	tickCumulative := big.NewInt(2)
	secondsPerLiquidityCumulativeX128 := big.NewInt(3)

	oi := NewObservationInfo(blockTimestamp, tickCumulative, secondsPerLiquidityCumulativeX128, true)
	require.Same(t, blockTimestamp, oi.BlockTimestamp)
	require.Same(t, tickCumulative, oi.TickCumulative)
	require.Same(t, secondsPerLiquidityCumulativeX128, oi.SecondsPerLiquidityCumulativeX128)
	require.True(t, oi.Initialized)
}

func TestNewObserveResult(t *testing.T) {
	tickCumulatives := []*big.Int{big.NewInt(1), big.NewInt(2)}
	secondsPerLiquidityCumulativeX128s := []*big.Int{big.NewInt(3), big.NewInt(4)}

	or := NewObserveResult(tickCumulatives, secondsPerLiquidityCumulativeX128s)
	require.Same(t, &tickCumulatives[0], &or.TickCumulatives[0])
	require.Same(t, &secondsPerLiquidityCumulativeX128s[0], &or.SecondsPerLiquidityCumulativeX128s[0])
}

func TestNewSnapshotCumulativesInside(t *testing.T) {
	tickCumulativeInside := big.NewInt(1)
	secondsPerLiquidityInsideX128 := big.NewInt(2)
	secondsInside := big.NewInt(3)

	sci := NewSnapshotCumulativesInside(tickCumulativeInside, secondsPerLiquidityInsideX128, secondsInside)
	require.Same(t, tickCumulativeInside, sci.TickCumulativeInside)
	require.Same(t, secondsPerLiquidityInsideX128, sci.SecondsPerLiquidityInsideX128)
	require.Same(t, secondsInside, sci.SecondsInside)
}

func TestNewInitialize(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	sqrtPriceX96 := big.NewInt(1)
	tick := big.NewInt(2)

	i := NewInitialize(contract, sqrtPriceX96, tick)
	require.Same(t, contract, i.Contract)
	require.Same(t, sqrtPriceX96, i.SqrtPriceX96)
	require.Same(t, tick, i.Tick)
}

func TestNewMint(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	sender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	owner := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	tickLower := big.NewInt(1)
	tickUpper := big.NewInt(2)
	amount := big.NewInt(3)
	amount0 := big.NewInt(4)
	amount1 := big.NewInt(5)

	m := NewMint(contract, sender, owner, tickLower, tickUpper, amount, amount0, amount1)
	require.Same(t, contract, m.Contract)
	require.Same(t, sender, m.Sender)
	require.Same(t, owner, m.Owner)
	require.Same(t, tickLower, m.TickLower)
	require.Same(t, tickUpper, m.TickUpper)
	require.Same(t, amount, m.Amount)
	require.Same(t, amount0, m.Amount0)
	require.Same(t, amount1, m.Amount1)
}

func TestNewCollect(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	owner := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	recipient := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	tickLower := big.NewInt(1)
	tickUpper := big.NewInt(2)
	amount0 := big.NewInt(3)
	amount1 := big.NewInt(4)

	c := NewCollect(contract, owner, recipient, tickLower, tickUpper, amount0, amount1)
	require.Same(t, contract, c.Contract)
	require.Same(t, owner, c.Owner)
	require.Same(t, recipient, c.Recipient)
	require.Same(t, tickLower, c.TickLower)
	require.Same(t, tickUpper, c.TickUpper)
	require.Same(t, amount0, c.Amount0)
	require.Same(t, amount1, c.Amount1)
}

func TestNewBurn(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	owner := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	tickLower := big.NewInt(1)
	tickUpper := big.NewInt(2)
	amount := big.NewInt(3)
	amount0 := big.NewInt(4)
	amount1 := big.NewInt(5)

	b := NewBurn(contract, owner, tickLower, tickUpper, amount, amount0, amount1)
	require.Same(t, contract, b.Contract)
	require.Same(t, owner, b.Owner)
	require.Same(t, tickLower, b.TickLower)
	require.Same(t, tickUpper, b.TickUpper)
	require.Same(t, amount, b.Amount)
	require.Same(t, amount0, b.Amount0)
	require.Same(t, amount1, b.Amount1)
}

func TestNewSwap(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	sender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	recipient := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	amount0 := big.NewInt(1)
	amount1 := big.NewInt(2)
	sqrtPriceX96 := big.NewInt(3)
	liquidity := big.NewInt(4)
	tick := big.NewInt(5)

	s := NewSwap(contract, sender, recipient, amount0, amount1, sqrtPriceX96, liquidity, tick)
	require.Same(t, contract, s.Contract)
	require.Same(t, sender, s.Sender)
	require.Same(t, recipient, s.Recipient)
	require.Same(t, amount0, s.Amount0)
	require.Same(t, amount1, s.Amount1)
	require.Same(t, sqrtPriceX96, s.SqrtPriceX96)
	require.Same(t, liquidity, s.Liquidity)
	require.Same(t, tick, s.Tick)
}

func TestNewFlash(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	sender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	recipient := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	amount0 := big.NewInt(1)
	amount1 := big.NewInt(2)
	paid0 := big.NewInt(3)
	paid1 := big.NewInt(4)

	f := NewFlash(contract, sender, recipient, amount0, amount1, paid0, paid1)
	require.Same(t, contract, f.Contract)
	require.Same(t, sender, f.Sender)
	require.Same(t, recipient, f.Recipient)
	require.Same(t, amount0, f.Amount0)
	require.Same(t, amount1, f.Amount1)
	require.Same(t, paid0, f.Paid0)
	require.Same(t, paid1, f.Paid1)
}

func TestNewIncreaseObservationCardinalityNext(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	old := big.NewInt(1)
	new_ := big.NewInt(2)

	i := NewIncreaseObservationCardinalityNext(contract, old, new_)
	require.Same(t, contract, i.Contract)
	require.Same(t, old, i.Old)
	require.Same(t, new_, i.New)
}

func TestNewSetFeeProtocol(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	feeProtocol0Old := big.NewInt(1)
	feeProtocol1Old := big.NewInt(2)
	feeProtocol0New := big.NewInt(3)
	feeProtocol1New := big.NewInt(4)

	s := NewSetFeeProtocol(contract, feeProtocol0Old, feeProtocol1Old, feeProtocol0New, feeProtocol1New)
	require.Same(t, contract, s.Contract)
	require.Same(t, feeProtocol0Old, s.FeeProtocol0Old)
	require.Same(t, feeProtocol1Old, s.FeeProtocol1Old)
	require.Same(t, feeProtocol0New, s.FeeProtocol0New)
	require.Same(t, feeProtocol1New, s.FeeProtocol1New)
}

func TestNewCollectProtocol(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	sender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	recipient := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	amount0 := big.NewInt(1)
	amount1 := big.NewInt(2)

	c := NewCollectProtocol(contract, sender, recipient, amount0, amount1)
	require.Same(t, contract, c.Contract)
	require.Same(t, sender, c.Sender)
	require.Same(t, recipient, c.Recipient)
	require.Same(t, amount0, c.Amount0)
	require.Same(t, amount1, c.Amount1)
}
