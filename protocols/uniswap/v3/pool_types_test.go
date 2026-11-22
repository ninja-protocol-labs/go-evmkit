package v3

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
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

func TestNewInitializeEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	sqrtPriceX96 := big.NewInt(1)
	tick := big.NewInt(2)

	i := NewInitializeEvent(contract, sqrtPriceX96, tick)
	require.Same(t, contract, i.Contract)
	require.Same(t, sqrtPriceX96, i.SqrtPriceX96)
	require.Same(t, tick, i.Tick)
}

func TestNewMintEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	sender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	owner := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	tickLower := big.NewInt(1)
	tickUpper := big.NewInt(2)
	amount := big.NewInt(3)
	amount0 := big.NewInt(4)
	amount1 := big.NewInt(5)

	m := NewMintEvent(contract, sender, owner, tickLower, tickUpper, amount, amount0, amount1)
	require.Same(t, contract, m.Contract)
	require.Same(t, sender, m.Sender)
	require.Same(t, owner, m.Owner)
	require.Same(t, tickLower, m.TickLower)
	require.Same(t, tickUpper, m.TickUpper)
	require.Same(t, amount, m.Amount)
	require.Same(t, amount0, m.Amount0)
	require.Same(t, amount1, m.Amount1)
}

func TestNewCollectEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	owner := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	recipient := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	tickLower := big.NewInt(1)
	tickUpper := big.NewInt(2)
	amount0 := big.NewInt(3)
	amount1 := big.NewInt(4)

	c := NewCollectEvent(contract, owner, recipient, tickLower, tickUpper, amount0, amount1)
	require.Same(t, contract, c.Contract)
	require.Same(t, owner, c.Owner)
	require.Same(t, recipient, c.Recipient)
	require.Same(t, tickLower, c.TickLower)
	require.Same(t, tickUpper, c.TickUpper)
	require.Same(t, amount0, c.Amount0)
	require.Same(t, amount1, c.Amount1)
}

func TestNewBurnEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	owner := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	tickLower := big.NewInt(1)
	tickUpper := big.NewInt(2)
	amount := big.NewInt(3)
	amount0 := big.NewInt(4)
	amount1 := big.NewInt(5)

	b := NewBurnEvent(contract, owner, tickLower, tickUpper, amount, amount0, amount1)
	require.Same(t, contract, b.Contract)
	require.Same(t, owner, b.Owner)
	require.Same(t, tickLower, b.TickLower)
	require.Same(t, tickUpper, b.TickUpper)
	require.Same(t, amount, b.Amount)
	require.Same(t, amount0, b.Amount0)
	require.Same(t, amount1, b.Amount1)
}

func TestNewSwapEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	sender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	recipient := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	amount0 := big.NewInt(1)
	amount1 := big.NewInt(2)
	sqrtPriceX96 := big.NewInt(3)
	liquidity := big.NewInt(4)
	tick := big.NewInt(5)

	s := NewSwapEvent(contract, sender, recipient, amount0, amount1, sqrtPriceX96, liquidity, tick)
	require.Same(t, contract, s.Contract)
	require.Same(t, sender, s.Sender)
	require.Same(t, recipient, s.Recipient)
	require.Same(t, amount0, s.Amount0)
	require.Same(t, amount1, s.Amount1)
	require.Same(t, sqrtPriceX96, s.SqrtPriceX96)
	require.Same(t, liquidity, s.Liquidity)
	require.Same(t, tick, s.Tick)
}

func TestNewFlashEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	sender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	recipient := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	amount0 := big.NewInt(1)
	amount1 := big.NewInt(2)
	paid0 := big.NewInt(3)
	paid1 := big.NewInt(4)

	f := NewFlashEvent(contract, sender, recipient, amount0, amount1, paid0, paid1)
	require.Same(t, contract, f.Contract)
	require.Same(t, sender, f.Sender)
	require.Same(t, recipient, f.Recipient)
	require.Same(t, amount0, f.Amount0)
	require.Same(t, amount1, f.Amount1)
	require.Same(t, paid0, f.Paid0)
	require.Same(t, paid1, f.Paid1)
}

func TestNewIncreaseObservationCardinalityNextEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	old := big.NewInt(1)
	new_ := big.NewInt(2)

	i := NewIncreaseObservationCardinalityNextEvent(contract, old, new_)
	require.Same(t, contract, i.Contract)
	require.Same(t, old, i.Old)
	require.Same(t, new_, i.New)
}

func TestNewSetFeeProtocolEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	feeProtocol0Old := big.NewInt(1)
	feeProtocol1Old := big.NewInt(2)
	feeProtocol0New := big.NewInt(3)
	feeProtocol1New := big.NewInt(4)

	s := NewSetFeeProtocolEvent(contract, feeProtocol0Old, feeProtocol1Old, feeProtocol0New, feeProtocol1New)
	require.Same(t, contract, s.Contract)
	require.Same(t, feeProtocol0Old, s.FeeProtocol0Old)
	require.Same(t, feeProtocol1Old, s.FeeProtocol1Old)
	require.Same(t, feeProtocol0New, s.FeeProtocol0New)
	require.Same(t, feeProtocol1New, s.FeeProtocol1New)
}

func TestNewCollectProtocolEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	sender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	recipient := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	amount0 := big.NewInt(1)
	amount1 := big.NewInt(2)

	c := NewCollectProtocolEvent(contract, sender, recipient, amount0, amount1)
	require.Same(t, contract, c.Contract)
	require.Same(t, sender, c.Sender)
	require.Same(t, recipient, c.Recipient)
	require.Same(t, amount0, c.Amount0)
	require.Same(t, amount1, c.Amount1)
}

func TestPoolRequestsValidate(t *testing.T) {
	pool := types.MustNewAddressFromHex("0x8ad599c3a0ff1de082011efddc58f1908eb6e6d8")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	var key [32]byte
	key[0], key[31] = 0x11, 0x22

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"factory nil", (*FactoryRequest)(nil).Validate(), "invalid request: request is nil"},
		{"factory empty", (&FactoryRequest{}).Validate(), "invalid request: contract is nil"},
		{"factory nil contract", (&FactoryRequest{}).Validate(), "invalid request: contract is nil"},
		{"factory valid", (&FactoryRequest{Contract: pool}).Validate(), ""},
		{"token0 nil", (*Token0Request)(nil).Validate(), "invalid request: request is nil"},
		{"token0 empty", (&Token0Request{}).Validate(), "invalid request: contract is nil"},
		{"token0 nil contract", (&Token0Request{}).Validate(), "invalid request: contract is nil"},
		{"token0 valid", (&Token0Request{Contract: pool}).Validate(), ""},
		{"token1 nil", (*Token1Request)(nil).Validate(), "invalid request: request is nil"},
		{"token1 empty", (&Token1Request{}).Validate(), "invalid request: contract is nil"},
		{"token1 nil contract", (&Token1Request{}).Validate(), "invalid request: contract is nil"},
		{"token1 valid", (&Token1Request{Contract: pool}).Validate(), ""},
		{"fee nil", (*FeeRequest)(nil).Validate(), "invalid request: request is nil"},
		{"fee empty", (&FeeRequest{}).Validate(), "invalid request: contract is nil"},
		{"fee nil contract", (&FeeRequest{}).Validate(), "invalid request: contract is nil"},
		{"fee valid", (&FeeRequest{Contract: pool}).Validate(), ""},
		{"tickSpacing nil", (*TickSpacingRequest)(nil).Validate(), "invalid request: request is nil"},
		{"tickSpacing empty", (&TickSpacingRequest{}).Validate(), "invalid request: contract is nil"},
		{"tickSpacing nil contract", (&TickSpacingRequest{}).Validate(), "invalid request: contract is nil"},
		{"tickSpacing valid", (&TickSpacingRequest{Contract: pool}).Validate(), ""},
		{"maxLiquidityPerTick nil", (*MaxLiquidityPerTickRequest)(nil).Validate(), "invalid request: request is nil"},
		{"maxLiquidityPerTick empty", (&MaxLiquidityPerTickRequest{}).Validate(), "invalid request: contract is nil"},
		{"maxLiquidityPerTick nil contract", (&MaxLiquidityPerTickRequest{}).Validate(), "invalid request: contract is nil"},
		{"maxLiquidityPerTick valid", (&MaxLiquidityPerTickRequest{Contract: pool}).Validate(), ""},
		{"slot0 nil", (*Slot0Request)(nil).Validate(), "invalid request: request is nil"},
		{"slot0 empty", (&Slot0Request{}).Validate(), "invalid request: contract is nil"},
		{"slot0 nil contract", (&Slot0Request{}).Validate(), "invalid request: contract is nil"},
		{"slot0 valid", (&Slot0Request{Contract: pool}).Validate(), ""},
		{"feeGrowthGlobal0X128 nil", (*FeeGrowthGlobal0X128Request)(nil).Validate(), "invalid request: request is nil"},
		{"feeGrowthGlobal0X128 empty", (&FeeGrowthGlobal0X128Request{}).Validate(), "invalid request: contract is nil"},
		{"feeGrowthGlobal0X128 nil contract", (&FeeGrowthGlobal0X128Request{}).Validate(), "invalid request: contract is nil"},
		{"feeGrowthGlobal0X128 valid", (&FeeGrowthGlobal0X128Request{Contract: pool}).Validate(), ""},
		{"feeGrowthGlobal1X128 nil", (*FeeGrowthGlobal1X128Request)(nil).Validate(), "invalid request: request is nil"},
		{"feeGrowthGlobal1X128 empty", (&FeeGrowthGlobal1X128Request{}).Validate(), "invalid request: contract is nil"},
		{"feeGrowthGlobal1X128 nil contract", (&FeeGrowthGlobal1X128Request{}).Validate(), "invalid request: contract is nil"},
		{"feeGrowthGlobal1X128 valid", (&FeeGrowthGlobal1X128Request{Contract: pool}).Validate(), ""},
		{"protocolFees nil", (*ProtocolFeesRequest)(nil).Validate(), "invalid request: request is nil"},
		{"protocolFees empty", (&ProtocolFeesRequest{}).Validate(), "invalid request: contract is nil"},
		{"protocolFees nil contract", (&ProtocolFeesRequest{}).Validate(), "invalid request: contract is nil"},
		{"protocolFees valid", (&ProtocolFeesRequest{Contract: pool}).Validate(), ""},
		{"liquidity nil", (*LiquidityRequest)(nil).Validate(), "invalid request: request is nil"},
		{"liquidity empty", (&LiquidityRequest{}).Validate(), "invalid request: contract is nil"},
		{"liquidity nil contract", (&LiquidityRequest{}).Validate(), "invalid request: contract is nil"},
		{"liquidity valid", (&LiquidityRequest{Contract: pool}).Validate(), ""},
		{"ticks nil", (*TicksRequest)(nil).Validate(), "invalid request: request is nil"},
		{"ticks empty", (&TicksRequest{}).Validate(), "invalid request: contract is nil"},
		{"ticks nil contract", (&TicksRequest{Tick: big.NewInt(-60)}).Validate(), "invalid request: contract is nil"},
		{"ticks nil tick", (&TicksRequest{Contract: pool}).Validate(), "invalid request: tick is nil"},
		{"ticks valid", (&TicksRequest{Contract: pool, Tick: big.NewInt(-60)}).Validate(), ""},
		{"tickBitmap nil", (*TickBitmapRequest)(nil).Validate(), "invalid request: request is nil"},
		{"tickBitmap empty", (&TickBitmapRequest{}).Validate(), "invalid request: contract is nil"},
		{"tickBitmap nil contract", (&TickBitmapRequest{WordPosition: int16(-3)}).Validate(), "invalid request: contract is nil"},
		{"tickBitmap valid", (&TickBitmapRequest{Contract: pool, WordPosition: int16(-3)}).Validate(), ""},
		{"positions nil", (*PositionsRequest)(nil).Validate(), "invalid request: request is nil"},
		{"positions empty", (&PositionsRequest{}).Validate(), "invalid request: contract is nil"},
		{"positions nil contract", (&PositionsRequest{Key: key}).Validate(), "invalid request: contract is nil"},
		{"positions valid", (&PositionsRequest{Contract: pool, Key: key}).Validate(), ""},
		{"observations nil", (*ObservationsRequest)(nil).Validate(), "invalid request: request is nil"},
		{"observations empty", (&ObservationsRequest{}).Validate(), "invalid request: contract is nil"},
		{"observations nil contract", (&ObservationsRequest{Index: big.NewInt(107)}).Validate(), "invalid request: contract is nil"},
		{"observations nil index", (&ObservationsRequest{Contract: pool}).Validate(), "invalid request: index is nil"},
		{"observations valid", (&ObservationsRequest{Contract: pool, Index: big.NewInt(107)}).Validate(), ""},
		{"observe nil", (*ObserveRequest)(nil).Validate(), "invalid request: request is nil"},
		{"observe empty", (&ObserveRequest{}).Validate(), "invalid request: contract is nil"},
		{"observe nil contract", (&ObserveRequest{SecondsAgos: []uint32{0, 60}}).Validate(), "invalid request: contract is nil"},
		{"observe valid", (&ObserveRequest{Contract: pool, SecondsAgos: []uint32{0, 60}}).Validate(), ""},
		{"snapshotCumulativesInside nil", (*SnapshotCumulativesInsideRequest)(nil).Validate(), "invalid request: request is nil"},
		{"snapshotCumulativesInside empty", (&SnapshotCumulativesInsideRequest{}).Validate(), "invalid request: contract is nil"},
		{"snapshotCumulativesInside nil contract", (&SnapshotCumulativesInsideRequest{TickLower: big.NewInt(-60), TickUpper: big.NewInt(60)}).Validate(), "invalid request: contract is nil"},
		{"snapshotCumulativesInside nil tickLower", (&SnapshotCumulativesInsideRequest{Contract: pool, TickUpper: big.NewInt(60)}).Validate(), "invalid request: tickLower is nil"},
		{"snapshotCumulativesInside nil tickUpper", (&SnapshotCumulativesInsideRequest{Contract: pool, TickLower: big.NewInt(-60)}).Validate(), "invalid request: tickUpper is nil"},
		{"snapshotCumulativesInside valid", (&SnapshotCumulativesInsideRequest{Contract: pool, TickLower: big.NewInt(-60), TickUpper: big.NewInt(60)}).Validate(), ""},
		{"initialize nil", (*InitializeRequest)(nil).Validate(), "invalid request: request is nil"},
		{"initialize empty", (&InitializeRequest{}).Validate(), "invalid request: contract is nil"},
		{"initialize nil contract", (&InitializeRequest{SqrtPriceX96: big.NewInt(107)}).Validate(), "invalid request: contract is nil"},
		{"initialize nil sqrtPriceX96", (&InitializeRequest{Contract: pool}).Validate(), "invalid request: sqrtPriceX96 is nil"},
		{"initialize valid", (&InitializeRequest{Contract: pool, SqrtPriceX96: big.NewInt(107)}).Validate(), ""},
		{"poolMint nil", (*PoolMintRequest)(nil).Validate(), "invalid request: request is nil"},
		{"poolMint empty", (&PoolMintRequest{}).Validate(), "invalid request: contract is nil"},
		{"poolMint nil contract", (&PoolMintRequest{Recipient: a, TickLower: big.NewInt(-60), TickUpper: big.NewInt(60), Amount: big.NewInt(407), CallbackData: []byte{0xca, 0xfe}}).Validate(), "invalid request: contract is nil"},
		{"poolMint nil recipient", (&PoolMintRequest{Contract: pool, TickLower: big.NewInt(-60), TickUpper: big.NewInt(60), Amount: big.NewInt(407), CallbackData: []byte{0xca, 0xfe}}).Validate(), "invalid request: recipient is nil"},
		{"poolMint nil tickLower", (&PoolMintRequest{Contract: pool, Recipient: a, TickUpper: big.NewInt(60), Amount: big.NewInt(407), CallbackData: []byte{0xca, 0xfe}}).Validate(), "invalid request: tickLower is nil"},
		{"poolMint nil tickUpper", (&PoolMintRequest{Contract: pool, Recipient: a, TickLower: big.NewInt(-60), Amount: big.NewInt(407), CallbackData: []byte{0xca, 0xfe}}).Validate(), "invalid request: tickUpper is nil"},
		{"poolMint nil amount", (&PoolMintRequest{Contract: pool, Recipient: a, TickLower: big.NewInt(-60), TickUpper: big.NewInt(60), CallbackData: []byte{0xca, 0xfe}}).Validate(), "invalid request: amount is nil"},
		{"poolMint valid", (&PoolMintRequest{Contract: pool, Recipient: a, TickLower: big.NewInt(-60), TickUpper: big.NewInt(60), Amount: big.NewInt(407), CallbackData: []byte{0xca, 0xfe}}).Validate(), ""},
		{"poolCollect nil", (*PoolCollectRequest)(nil).Validate(), "invalid request: request is nil"},
		{"poolCollect empty", (&PoolCollectRequest{}).Validate(), "invalid request: contract is nil"},
		{"poolCollect nil contract", (&PoolCollectRequest{Recipient: a, TickLower: big.NewInt(-60), TickUpper: big.NewInt(60), Amount0Requested: big.NewInt(407), Amount1Requested: big.NewInt(507)}).Validate(), "invalid request: contract is nil"},
		{"poolCollect nil recipient", (&PoolCollectRequest{Contract: pool, TickLower: big.NewInt(-60), TickUpper: big.NewInt(60), Amount0Requested: big.NewInt(407), Amount1Requested: big.NewInt(507)}).Validate(), "invalid request: recipient is nil"},
		{"poolCollect nil tickLower", (&PoolCollectRequest{Contract: pool, Recipient: a, TickUpper: big.NewInt(60), Amount0Requested: big.NewInt(407), Amount1Requested: big.NewInt(507)}).Validate(), "invalid request: tickLower is nil"},
		{"poolCollect nil tickUpper", (&PoolCollectRequest{Contract: pool, Recipient: a, TickLower: big.NewInt(-60), Amount0Requested: big.NewInt(407), Amount1Requested: big.NewInt(507)}).Validate(), "invalid request: tickUpper is nil"},
		{"poolCollect nil amount0Requested", (&PoolCollectRequest{Contract: pool, Recipient: a, TickLower: big.NewInt(-60), TickUpper: big.NewInt(60), Amount1Requested: big.NewInt(507)}).Validate(), "invalid request: amount0Requested is nil"},
		{"poolCollect nil amount1Requested", (&PoolCollectRequest{Contract: pool, Recipient: a, TickLower: big.NewInt(-60), TickUpper: big.NewInt(60), Amount0Requested: big.NewInt(407)}).Validate(), "invalid request: amount1Requested is nil"},
		{"poolCollect valid", (&PoolCollectRequest{Contract: pool, Recipient: a, TickLower: big.NewInt(-60), TickUpper: big.NewInt(60), Amount0Requested: big.NewInt(407), Amount1Requested: big.NewInt(507)}).Validate(), ""},
		{"burn nil", (*BurnRequest)(nil).Validate(), "invalid request: request is nil"},
		{"burn empty", (&BurnRequest{}).Validate(), "invalid request: contract is nil"},
		{"burn nil contract", (&BurnRequest{TickLower: big.NewInt(-60), TickUpper: big.NewInt(60), Amount: big.NewInt(307)}).Validate(), "invalid request: contract is nil"},
		{"burn nil tickLower", (&BurnRequest{Contract: pool, TickUpper: big.NewInt(60), Amount: big.NewInt(307)}).Validate(), "invalid request: tickLower is nil"},
		{"burn nil tickUpper", (&BurnRequest{Contract: pool, TickLower: big.NewInt(-60), Amount: big.NewInt(307)}).Validate(), "invalid request: tickUpper is nil"},
		{"burn nil amount", (&BurnRequest{Contract: pool, TickLower: big.NewInt(-60), TickUpper: big.NewInt(60)}).Validate(), "invalid request: amount is nil"},
		{"burn valid", (&BurnRequest{Contract: pool, TickLower: big.NewInt(-60), TickUpper: big.NewInt(60), Amount: big.NewInt(307)}).Validate(), ""},
		{"swap nil", (*SwapRequest)(nil).Validate(), "invalid request: request is nil"},
		{"swap empty", (&SwapRequest{}).Validate(), "invalid request: contract is nil"},
		{"swap nil contract", (&SwapRequest{Recipient: a, ZeroForOne: true, AmountSpecified: big.NewInt(307), SqrtPriceLimitX96: big.NewInt(407), CallbackData: []byte{0xca, 0xfe}}).Validate(), "invalid request: contract is nil"},
		{"swap nil recipient", (&SwapRequest{Contract: pool, ZeroForOne: true, AmountSpecified: big.NewInt(307), SqrtPriceLimitX96: big.NewInt(407), CallbackData: []byte{0xca, 0xfe}}).Validate(), "invalid request: recipient is nil"},
		{"swap nil amountSpecified", (&SwapRequest{Contract: pool, Recipient: a, ZeroForOne: true, SqrtPriceLimitX96: big.NewInt(407), CallbackData: []byte{0xca, 0xfe}}).Validate(), "invalid request: amountSpecified is nil"},
		{"swap nil sqrtPriceLimitX96", (&SwapRequest{Contract: pool, Recipient: a, ZeroForOne: true, AmountSpecified: big.NewInt(307), CallbackData: []byte{0xca, 0xfe}}).Validate(), "invalid request: sqrtPriceLimitX96 is nil"},
		{"swap valid", (&SwapRequest{Contract: pool, Recipient: a, ZeroForOne: true, AmountSpecified: big.NewInt(307), SqrtPriceLimitX96: big.NewInt(407), CallbackData: []byte{0xca, 0xfe}}).Validate(), ""},
		{"flash nil", (*FlashRequest)(nil).Validate(), "invalid request: request is nil"},
		{"flash empty", (&FlashRequest{}).Validate(), "invalid request: contract is nil"},
		{"flash nil contract", (&FlashRequest{Recipient: a, Amount0: big.NewInt(207), Amount1: big.NewInt(307), CallbackData: []byte{0xca, 0xfe}}).Validate(), "invalid request: contract is nil"},
		{"flash nil recipient", (&FlashRequest{Contract: pool, Amount0: big.NewInt(207), Amount1: big.NewInt(307), CallbackData: []byte{0xca, 0xfe}}).Validate(), "invalid request: recipient is nil"},
		{"flash nil amount0", (&FlashRequest{Contract: pool, Recipient: a, Amount1: big.NewInt(307), CallbackData: []byte{0xca, 0xfe}}).Validate(), "invalid request: amount0 is nil"},
		{"flash nil amount1", (&FlashRequest{Contract: pool, Recipient: a, Amount0: big.NewInt(207), CallbackData: []byte{0xca, 0xfe}}).Validate(), "invalid request: amount1 is nil"},
		{"flash valid", (&FlashRequest{Contract: pool, Recipient: a, Amount0: big.NewInt(207), Amount1: big.NewInt(307), CallbackData: []byte{0xca, 0xfe}}).Validate(), ""},
		{"increaseObservationCardinalityNext nil", (*IncreaseObservationCardinalityNextRequest)(nil).Validate(), "invalid request: request is nil"},
		{"increaseObservationCardinalityNext empty", (&IncreaseObservationCardinalityNextRequest{}).Validate(), "invalid request: contract is nil"},
		{"increaseObservationCardinalityNext nil contract", (&IncreaseObservationCardinalityNextRequest{ObservationCardinalityNext: uint16(10)}).Validate(), "invalid request: contract is nil"},
		{"increaseObservationCardinalityNext valid", (&IncreaseObservationCardinalityNextRequest{Contract: pool, ObservationCardinalityNext: uint16(10)}).Validate(), ""},
		{"setFeeProtocol nil", (*SetFeeProtocolRequest)(nil).Validate(), "invalid request: request is nil"},
		{"setFeeProtocol empty", (&SetFeeProtocolRequest{}).Validate(), "invalid request: contract is nil"},
		{"setFeeProtocol nil contract", (&SetFeeProtocolRequest{FeeProtocol0: uint8(5), FeeProtocol1: uint8(6)}).Validate(), "invalid request: contract is nil"},
		{"setFeeProtocol valid", (&SetFeeProtocolRequest{Contract: pool, FeeProtocol0: uint8(5), FeeProtocol1: uint8(6)}).Validate(), ""},
		{"collectProtocol nil", (*CollectProtocolRequest)(nil).Validate(), "invalid request: request is nil"},
		{"collectProtocol empty", (&CollectProtocolRequest{}).Validate(), "invalid request: contract is nil"},
		{"collectProtocol nil contract", (&CollectProtocolRequest{Recipient: a, Amount0Requested: big.NewInt(207), Amount1Requested: big.NewInt(307)}).Validate(), "invalid request: contract is nil"},
		{"collectProtocol nil recipient", (&CollectProtocolRequest{Contract: pool, Amount0Requested: big.NewInt(207), Amount1Requested: big.NewInt(307)}).Validate(), "invalid request: recipient is nil"},
		{"collectProtocol nil amount0Requested", (&CollectProtocolRequest{Contract: pool, Recipient: a, Amount1Requested: big.NewInt(307)}).Validate(), "invalid request: amount0Requested is nil"},
		{"collectProtocol nil amount1Requested", (&CollectProtocolRequest{Contract: pool, Recipient: a, Amount0Requested: big.NewInt(207)}).Validate(), "invalid request: amount1Requested is nil"},
		{"collectProtocol valid", (&CollectProtocolRequest{Contract: pool, Recipient: a, Amount0Requested: big.NewInt(207), Amount1Requested: big.NewInt(307)}).Validate(), ""},
		{"poolMint nil callbackData", (&PoolMintRequest{Contract: pool, Recipient: a, TickLower: big.NewInt(-60), TickUpper: big.NewInt(60), Amount: big.NewInt(1)}).Validate(), ""},
		{"swap nil callbackData", (&SwapRequest{Contract: pool, Recipient: a, AmountSpecified: big.NewInt(-1), SqrtPriceLimitX96: big.NewInt(1)}).Validate(), ""},
		{"observe empty secondsAgos", (&ObserveRequest{Contract: pool}).Validate(), ""},
		{"positions zero key", (&PositionsRequest{Contract: pool}).Validate(), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.want == "" {
				require.NoError(t, tc.err)
				return
			}
			require.ErrorIs(t, tc.err, bind.ErrInvalidRequest)
			require.EqualError(t, tc.err, tc.want)
		})
	}
}

func TestNewPoolRequestsSetFields(t *testing.T) {
	pool := types.MustNewAddressFromHex("0x8ad599c3a0ff1de082011efddc58f1908eb6e6d8")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	var key [32]byte
	key[0], key[31] = 0x11, 0x22

	require.Equal(t, &FactoryRequest{Contract: pool}, NewFactoryRequest(pool))
	require.Equal(t, &Token0Request{Contract: pool}, NewToken0Request(pool))
	require.Equal(t, &Token1Request{Contract: pool}, NewToken1Request(pool))
	require.Equal(t, &FeeRequest{Contract: pool}, NewFeeRequest(pool))
	require.Equal(t, &TickSpacingRequest{Contract: pool}, NewTickSpacingRequest(pool))
	require.Equal(t, &MaxLiquidityPerTickRequest{Contract: pool}, NewMaxLiquidityPerTickRequest(pool))
	require.Equal(t, &Slot0Request{Contract: pool}, NewSlot0Request(pool))
	require.Equal(t, &FeeGrowthGlobal0X128Request{Contract: pool}, NewFeeGrowthGlobal0X128Request(pool))
	require.Equal(t, &FeeGrowthGlobal1X128Request{Contract: pool}, NewFeeGrowthGlobal1X128Request(pool))
	require.Equal(t, &ProtocolFeesRequest{Contract: pool}, NewProtocolFeesRequest(pool))
	require.Equal(t, &LiquidityRequest{Contract: pool}, NewLiquidityRequest(pool))
	require.Equal(t, &TicksRequest{Contract: pool, Tick: big.NewInt(-60)}, NewTicksRequest(pool, big.NewInt(-60)))
	require.Equal(t, &TickBitmapRequest{Contract: pool, WordPosition: int16(-3)}, NewTickBitmapRequest(pool, int16(-3)))
	require.Equal(t, &PositionsRequest{Contract: pool, Key: key}, NewPositionsRequest(pool, key))
	require.Equal(t, &ObservationsRequest{Contract: pool, Index: big.NewInt(107)}, NewObservationsRequest(pool, big.NewInt(107)))
	require.Equal(t, &ObserveRequest{Contract: pool, SecondsAgos: []uint32{0, 60}}, NewObserveRequest(pool, []uint32{0, 60}))
	require.Equal(t, &SnapshotCumulativesInsideRequest{Contract: pool, TickLower: big.NewInt(-60), TickUpper: big.NewInt(60)}, NewSnapshotCumulativesInsideRequest(pool, big.NewInt(-60), big.NewInt(60)))
	require.Equal(t, &InitializeRequest{Contract: pool, SqrtPriceX96: big.NewInt(107)}, NewInitializeRequest(pool, big.NewInt(107)))
	require.Equal(t, &PoolMintRequest{Contract: pool, Recipient: a, TickLower: big.NewInt(-60), TickUpper: big.NewInt(60), Amount: big.NewInt(407), CallbackData: []byte{0xca, 0xfe}}, NewPoolMintRequest(pool, a, big.NewInt(-60), big.NewInt(60), big.NewInt(407), []byte{0xca, 0xfe}))
	require.Equal(t, &PoolCollectRequest{Contract: pool, Recipient: a, TickLower: big.NewInt(-60), TickUpper: big.NewInt(60), Amount0Requested: big.NewInt(407), Amount1Requested: big.NewInt(507)}, NewPoolCollectRequest(pool, a, big.NewInt(-60), big.NewInt(60), big.NewInt(407), big.NewInt(507)))
	require.Equal(t, &BurnRequest{Contract: pool, TickLower: big.NewInt(-60), TickUpper: big.NewInt(60), Amount: big.NewInt(307)}, NewBurnRequest(pool, big.NewInt(-60), big.NewInt(60), big.NewInt(307)))
	require.Equal(t, &SwapRequest{Contract: pool, Recipient: a, ZeroForOne: true, AmountSpecified: big.NewInt(307), SqrtPriceLimitX96: big.NewInt(407), CallbackData: []byte{0xca, 0xfe}}, NewSwapRequest(pool, a, true, big.NewInt(307), big.NewInt(407), []byte{0xca, 0xfe}))
	require.Equal(t, &FlashRequest{Contract: pool, Recipient: a, Amount0: big.NewInt(207), Amount1: big.NewInt(307), CallbackData: []byte{0xca, 0xfe}}, NewFlashRequest(pool, a, big.NewInt(207), big.NewInt(307), []byte{0xca, 0xfe}))
	require.Equal(t, &IncreaseObservationCardinalityNextRequest{Contract: pool, ObservationCardinalityNext: uint16(10)}, NewIncreaseObservationCardinalityNextRequest(pool, uint16(10)))
	require.Equal(t, &SetFeeProtocolRequest{Contract: pool, FeeProtocol0: uint8(5), FeeProtocol1: uint8(6)}, NewSetFeeProtocolRequest(pool, uint8(5), uint8(6)))
	require.Equal(t, &CollectProtocolRequest{Contract: pool, Recipient: a, Amount0Requested: big.NewInt(207), Amount1Requested: big.NewInt(307)}, NewCollectProtocolRequest(pool, a, big.NewInt(207), big.NewInt(307)))
}
