package v4

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewPoolKey(t *testing.T) {
	currency0 := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	currency1 := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	fee := big.NewInt(3000)
	tickSpacing := big.NewInt(60)
	hooks := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")

	k := NewPoolKey(currency0, currency1, fee, tickSpacing, hooks)
	require.Same(t, currency0, k.Currency0)
	require.Same(t, currency1, k.Currency1)
	require.Same(t, fee, k.Fee)
	require.Same(t, tickSpacing, k.TickSpacing)
	require.Same(t, hooks, k.Hooks)
}

func TestNewBalanceDelta(t *testing.T) {
	amount0 := big.NewInt(100)
	amount1 := big.NewInt(-50)

	d := NewBalanceDelta(amount0, amount1)
	require.Same(t, amount0, d.Amount0)
	require.Same(t, amount1, d.Amount1)
}

func TestNewLiquidityParams(t *testing.T) {
	tickLower := big.NewInt(-100)
	tickUpper := big.NewInt(100)
	liquidityDelta := big.NewInt(1000)
	salt := [32]byte{0x01, 0x02}

	p := NewLiquidityParams(tickLower, tickUpper, liquidityDelta, salt)
	require.Same(t, tickLower, p.TickLower)
	require.Same(t, tickUpper, p.TickUpper)
	require.Same(t, liquidityDelta, p.LiquidityDelta)
	require.Equal(t, salt, p.Salt)
}

func TestNewSwapAmountParams(t *testing.T) {
	amountSpecified := big.NewInt(-500)
	sqrtPriceLimitX96 := big.NewInt(123456789)

	p := NewSwapAmountParams(true, amountSpecified, sqrtPriceLimitX96)
	require.True(t, p.ZeroForOne)
	require.Same(t, amountSpecified, p.AmountSpecified)
	require.Same(t, sqrtPriceLimitX96, p.SqrtPriceLimitX96)
}

func TestNewInitialize(t *testing.T) {
	poolManager := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	id := PoolID{0x01}
	currency0 := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	currency1 := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	fee := big.NewInt(3000)
	tickSpacing := big.NewInt(60)
	hooks := types.MustNewAddressFromHex("0x1231deb6f5749ef6ce6943a275a1d3e7486f4eae")
	sqrtPriceX96 := big.NewInt(792281625142643375)
	tick := big.NewInt(0)

	init := NewInitialize(poolManager, id, currency0, currency1, fee, tickSpacing, hooks, sqrtPriceX96, tick)
	require.Same(t, poolManager, init.PoolManager)
	require.Equal(t, id, init.Id)
	require.Same(t, currency0, init.Currency0)
	require.Same(t, currency1, init.Currency1)
	require.Same(t, fee, init.Fee)
	require.Same(t, tickSpacing, init.TickSpacing)
	require.Same(t, hooks, init.Hooks)
	require.Same(t, sqrtPriceX96, init.SqrtPriceX96)
	require.Same(t, tick, init.Tick)
}

func TestNewModifyLiquidity(t *testing.T) {
	poolManager := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	id := PoolID{0x02}
	sender := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	tickLower := big.NewInt(-100)
	tickUpper := big.NewInt(100)
	liquidityDelta := big.NewInt(1000)
	salt := [32]byte{0x03, 0x04}

	m := NewModifyLiquidity(poolManager, id, sender, tickLower, tickUpper, liquidityDelta, salt)
	require.Same(t, poolManager, m.PoolManager)
	require.Equal(t, id, m.Id)
	require.Same(t, sender, m.Sender)
	require.Same(t, tickLower, m.TickLower)
	require.Same(t, tickUpper, m.TickUpper)
	require.Same(t, liquidityDelta, m.LiquidityDelta)
	require.Equal(t, salt, m.Salt)
}

func TestNewSwap(t *testing.T) {
	poolManager := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	id := PoolID{0x05}
	sender := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	amount0 := big.NewInt(100)
	amount1 := big.NewInt(-50)
	sqrtPriceX96 := big.NewInt(792281625142643375)
	liquidity := big.NewInt(1000)
	tick := big.NewInt(42)
	fee := big.NewInt(3000)

	s := NewSwap(poolManager, id, sender, amount0, amount1, sqrtPriceX96, liquidity, tick, fee)
	require.Same(t, poolManager, s.PoolManager)
	require.Equal(t, id, s.Id)
	require.Same(t, sender, s.Sender)
	require.Same(t, amount0, s.Amount0)
	require.Same(t, amount1, s.Amount1)
	require.Same(t, sqrtPriceX96, s.SqrtPriceX96)
	require.Same(t, liquidity, s.Liquidity)
	require.Same(t, tick, s.Tick)
	require.Same(t, fee, s.Fee)
}

func TestNewDonate(t *testing.T) {
	poolManager := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	id := PoolID{0x06}
	sender := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	amount0 := big.NewInt(100)
	amount1 := big.NewInt(200)

	d := NewDonate(poolManager, id, sender, amount0, amount1)
	require.Same(t, poolManager, d.PoolManager)
	require.Equal(t, id, d.Id)
	require.Same(t, sender, d.Sender)
	require.Same(t, amount0, d.Amount0)
	require.Same(t, amount1, d.Amount1)
}

func TestNewProtocolFeeControllerUpdated(t *testing.T) {
	poolManager := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	controller := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	u := NewProtocolFeeControllerUpdated(poolManager, controller)
	require.Same(t, poolManager, u.PoolManager)
	require.Same(t, controller, u.ProtocolFeeController)
}

func TestNewProtocolFeeUpdated(t *testing.T) {
	poolManager := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	id := PoolID{0x07}
	protocolFee := big.NewInt(500)

	u := NewProtocolFeeUpdated(poolManager, id, protocolFee)
	require.Same(t, poolManager, u.PoolManager)
	require.Equal(t, id, u.Id)
	require.Same(t, protocolFee, u.ProtocolFee)
}
