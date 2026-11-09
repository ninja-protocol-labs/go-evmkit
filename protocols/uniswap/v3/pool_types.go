package v3

import (
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Slot0 mirrors IUniswapV3PoolState's slot0 tuple.
type Slot0 struct {
	SqrtPriceX96               *big.Int
	Tick                       *big.Int
	ObservationIndex           *big.Int
	ObservationCardinality     *big.Int
	ObservationCardinalityNext *big.Int
	FeeProtocol                *big.Int
	Unlocked                   bool
}

func NewSlot0(sqrtPriceX96, tick, observationIndex, observationCardinality, observationCardinalityNext, feeProtocol *big.Int, unlocked bool) *Slot0 {
	return &Slot0{
		SqrtPriceX96:               sqrtPriceX96,
		Tick:                       tick,
		ObservationIndex:           observationIndex,
		ObservationCardinality:     observationCardinality,
		ObservationCardinalityNext: observationCardinalityNext,
		FeeProtocol:                feeProtocol,
		Unlocked:                   unlocked,
	}
}

// ProtocolFees mirrors IUniswapV3PoolState's protocolFees tuple.
type ProtocolFees struct {
	Token0 *big.Int
	Token1 *big.Int
}

func NewProtocolFees(token0, token1 *big.Int) *ProtocolFees {
	return &ProtocolFees{
		Token0: token0,
		Token1: token1,
	}
}

// TickInfo mirrors IUniswapV3PoolState's ticks(tick) tuple.
type TickInfo struct {
	LiquidityGross                 *big.Int
	LiquidityNet                   *big.Int
	FeeGrowthOutside0X128          *big.Int
	FeeGrowthOutside1X128          *big.Int
	TickCumulativeOutside          *big.Int
	SecondsPerLiquidityOutsideX128 *big.Int
	SecondsOutside                 *big.Int
	Initialized                    bool
}

func NewTickInfo(liquidityGross, liquidityNet, feeGrowthOutside0X128, feeGrowthOutside1X128, tickCumulativeOutside, secondsPerLiquidityOutsideX128, secondsOutside *big.Int, initialized bool) *TickInfo {
	return &TickInfo{
		LiquidityGross:                 liquidityGross,
		LiquidityNet:                   liquidityNet,
		FeeGrowthOutside0X128:          feeGrowthOutside0X128,
		FeeGrowthOutside1X128:          feeGrowthOutside1X128,
		TickCumulativeOutside:          tickCumulativeOutside,
		SecondsPerLiquidityOutsideX128: secondsPerLiquidityOutsideX128,
		SecondsOutside:                 secondsOutside,
		Initialized:                    initialized,
	}
}

// PositionInfo mirrors IUniswapV3PoolState's positions(key) tuple.
type PositionInfo struct {
	Liquidity                *big.Int
	FeeGrowthInside0LastX128 *big.Int
	FeeGrowthInside1LastX128 *big.Int
	TokensOwed0              *big.Int
	TokensOwed1              *big.Int
}

func NewPositionInfo(liquidity, feeGrowthInside0LastX128, feeGrowthInside1LastX128, tokensOwed0, tokensOwed1 *big.Int) *PositionInfo {
	return &PositionInfo{
		Liquidity:                liquidity,
		FeeGrowthInside0LastX128: feeGrowthInside0LastX128,
		FeeGrowthInside1LastX128: feeGrowthInside1LastX128,
		TokensOwed0:              tokensOwed0,
		TokensOwed1:              tokensOwed1,
	}
}

// ObservationInfo mirrors IUniswapV3PoolState's observations(index) tuple.
type ObservationInfo struct {
	BlockTimestamp                    *big.Int
	TickCumulative                    *big.Int
	SecondsPerLiquidityCumulativeX128 *big.Int
	Initialized                       bool
}

func NewObservationInfo(blockTimestamp, tickCumulative, secondsPerLiquidityCumulativeX128 *big.Int, initialized bool) *ObservationInfo {
	return &ObservationInfo{
		BlockTimestamp:                    blockTimestamp,
		TickCumulative:                    tickCumulative,
		SecondsPerLiquidityCumulativeX128: secondsPerLiquidityCumulativeX128,
		Initialized:                       initialized,
	}
}

// ObserveResult mirrors IUniswapV3PoolDerivedState's observe return values, parallel slices
// indexed the same as the secondsAgos argument.
type ObserveResult struct {
	TickCumulatives                    []*big.Int
	SecondsPerLiquidityCumulativeX128s []*big.Int
}

func NewObserveResult(tickCumulatives, secondsPerLiquidityCumulativeX128s []*big.Int) *ObserveResult {
	return &ObserveResult{
		TickCumulatives:                    tickCumulatives,
		SecondsPerLiquidityCumulativeX128s: secondsPerLiquidityCumulativeX128s,
	}
}

// SnapshotCumulativesInside mirrors IUniswapV3PoolDerivedState's snapshotCumulativesInside tuple.
type SnapshotCumulativesInside struct {
	TickCumulativeInside          *big.Int
	SecondsPerLiquidityInsideX128 *big.Int
	SecondsInside                 *big.Int
}

func NewSnapshotCumulativesInside(tickCumulativeInside, secondsPerLiquidityInsideX128, secondsInside *big.Int) *SnapshotCumulativesInside {
	return &SnapshotCumulativesInside{
		TickCumulativeInside:          tickCumulativeInside,
		SecondsPerLiquidityInsideX128: secondsPerLiquidityInsideX128,
		SecondsInside:                 secondsInside,
	}
}

// InitializeParams mirrors IUniswapV3PoolEvents' Initialize event: Contract was initialized at
// SqrtPriceX96/Tick, emitted exactly once, before any Mint/Burn/Swap.
// InitializeParams is CreateInitialize's argument list.
type InitializeParams struct {
	Pool         *types.Address
	Sender       *types.Address
	SqrtPriceX96 *big.Int
}

func NewInitializeParams(pool, sender *types.Address, sqrtPriceX96 *big.Int) *InitializeParams {
	return &InitializeParams{
		Pool:         pool,
		Sender:       sender,
		SqrtPriceX96: sqrtPriceX96,
	}
}

// PoolMintParams is CreateMint's and CallMint's argument list.
type PoolMintParams struct {
	Pool         *types.Address
	Sender       *types.Address
	Recipient    *types.Address
	TickLower    *big.Int
	TickUpper    *big.Int
	Amount       *big.Int
	CallbackData []byte
}

func NewPoolMintParams(pool, sender, recipient *types.Address, tickLower, tickUpper, amount *big.Int, callbackData []byte) *PoolMintParams {
	return &PoolMintParams{
		Pool:         pool,
		Sender:       sender,
		Recipient:    recipient,
		TickLower:    tickLower,
		TickUpper:    tickUpper,
		Amount:       amount,
		CallbackData: callbackData,
	}
}

// PoolCollectParams is CreateCollect's and CallCollect's argument list.
type PoolCollectParams struct {
	Pool             *types.Address
	Sender           *types.Address
	Recipient        *types.Address
	TickLower        *big.Int
	TickUpper        *big.Int
	Amount0Requested *big.Int
	Amount1Requested *big.Int
}

func NewPoolCollectParams(pool, sender, recipient *types.Address, tickLower, tickUpper, amount0Requested, amount1Requested *big.Int) *PoolCollectParams {
	return &PoolCollectParams{
		Pool:             pool,
		Sender:           sender,
		Recipient:        recipient,
		TickLower:        tickLower,
		TickUpper:        tickUpper,
		Amount0Requested: amount0Requested,
		Amount1Requested: amount1Requested,
	}
}

// BurnParams is CreateBurn's and CallBurn's argument list.
type BurnParams struct {
	Pool      *types.Address
	Sender    *types.Address
	TickLower *big.Int
	TickUpper *big.Int
	Amount    *big.Int
}

func NewBurnParams(pool, sender *types.Address, tickLower, tickUpper, amount *big.Int) *BurnParams {
	return &BurnParams{
		Pool:      pool,
		Sender:    sender,
		TickLower: tickLower,
		TickUpper: tickUpper,
		Amount:    amount,
	}
}

// SwapParams is CreateSwap's and CallSwap's argument list.
type SwapParams struct {
	Pool              *types.Address
	Sender            *types.Address
	Recipient         *types.Address
	ZeroForOne        bool
	AmountSpecified   *big.Int
	SqrtPriceLimitX96 *big.Int
	CallbackData      []byte
}

func NewSwapParams(pool, sender, recipient *types.Address, zeroForOne bool, amountSpecified, sqrtPriceLimitX96 *big.Int, callbackData []byte) *SwapParams {
	return &SwapParams{
		Pool:              pool,
		Sender:            sender,
		Recipient:         recipient,
		ZeroForOne:        zeroForOne,
		AmountSpecified:   amountSpecified,
		SqrtPriceLimitX96: sqrtPriceLimitX96,
		CallbackData:      callbackData,
	}
}

// FlashParams is CreateFlash's argument list.
type FlashParams struct {
	Pool         *types.Address
	Sender       *types.Address
	Recipient    *types.Address
	Amount0      *big.Int
	Amount1      *big.Int
	CallbackData []byte
}

func NewFlashParams(pool, sender, recipient *types.Address, amount0, amount1 *big.Int, callbackData []byte) *FlashParams {
	return &FlashParams{
		Pool:         pool,
		Sender:       sender,
		Recipient:    recipient,
		Amount0:      amount0,
		Amount1:      amount1,
		CallbackData: callbackData,
	}
}

// IncreaseObservationCardinalityNextParams is CreateIncreaseObservationCardinalityNext's
// argument list.
type IncreaseObservationCardinalityNextParams struct {
	Pool                       *types.Address
	Sender                     *types.Address
	ObservationCardinalityNext uint16
}

func NewIncreaseObservationCardinalityNextParams(pool, sender *types.Address, observationCardinalityNext uint16) *IncreaseObservationCardinalityNextParams {
	return &IncreaseObservationCardinalityNextParams{
		Pool:                       pool,
		Sender:                     sender,
		ObservationCardinalityNext: observationCardinalityNext,
	}
}

// SetFeeProtocolParams is CreateSetFeeProtocol's argument list.
type SetFeeProtocolParams struct {
	Pool         *types.Address
	Sender       *types.Address
	FeeProtocol0 uint8
	FeeProtocol1 uint8
}

func NewSetFeeProtocolParams(pool, sender *types.Address, feeProtocol0, feeProtocol1 uint8) *SetFeeProtocolParams {
	return &SetFeeProtocolParams{
		Pool:         pool,
		Sender:       sender,
		FeeProtocol0: feeProtocol0,
		FeeProtocol1: feeProtocol1,
	}
}

// CollectProtocolParams is CreateCollectProtocol's and CallCollectProtocol's argument list.
type CollectProtocolParams struct {
	Pool             *types.Address
	Sender           *types.Address
	Recipient        *types.Address
	Amount0Requested *big.Int
	Amount1Requested *big.Int
}

func NewCollectProtocolParams(pool, sender, recipient *types.Address, amount0Requested, amount1Requested *big.Int) *CollectProtocolParams {
	return &CollectProtocolParams{
		Pool:             pool,
		Sender:           sender,
		Recipient:        recipient,
		Amount0Requested: amount0Requested,
		Amount1Requested: amount1Requested,
	}
}

type Initialize struct {
	Contract     *types.Address
	SqrtPriceX96 *big.Int
	Tick         *big.Int
}

func NewInitialize(contract *types.Address, sqrtPriceX96, tick *big.Int) *Initialize {
	return &Initialize{
		Contract:     contract,
		SqrtPriceX96: sqrtPriceX96,
		Tick:         tick,
	}
}

// Mint mirrors IUniswapV3PoolEvents' Mint event: Sender minted Amount liquidity into Owner's
// [TickLower, TickUpper] position, paying Amount0/Amount1.
type Mint struct {
	Contract  *types.Address
	Sender    *types.Address
	Owner     *types.Address
	TickLower *big.Int
	TickUpper *big.Int
	Amount    *big.Int
	Amount0   *big.Int
	Amount1   *big.Int
}

func NewMint(contract, sender, owner *types.Address, tickLower, tickUpper, amount, amount0, amount1 *big.Int) *Mint {
	return &Mint{
		Contract:  contract,
		Sender:    sender,
		Owner:     owner,
		TickLower: tickLower,
		TickUpper: tickUpper,
		Amount:    amount,
		Amount0:   amount0,
		Amount1:   amount1,
	}
}

// Collect mirrors IUniswapV3PoolEvents' Collect event: fees owed to Owner's [TickLower,
// TickUpper] position were sent to Recipient.
type Collect struct {
	Contract  *types.Address
	Owner     *types.Address
	Recipient *types.Address
	TickLower *big.Int
	TickUpper *big.Int
	Amount0   *big.Int
	Amount1   *big.Int
}

func NewCollect(contract, owner, recipient *types.Address, tickLower, tickUpper, amount0, amount1 *big.Int) *Collect {
	return &Collect{
		Contract:  contract,
		Owner:     owner,
		Recipient: recipient,
		TickLower: tickLower,
		TickUpper: tickUpper,
		Amount0:   amount0,
		Amount1:   amount1,
	}
}

// Burn mirrors IUniswapV3PoolEvents' Burn event: Amount liquidity was removed from Owner's
// [TickLower, TickUpper] position, returning Amount0/Amount1 (fees must be withdrawn separately
// via Collect).
type Burn struct {
	Contract  *types.Address
	Owner     *types.Address
	TickLower *big.Int
	TickUpper *big.Int
	Amount    *big.Int
	Amount0   *big.Int
	Amount1   *big.Int
}

func NewBurn(contract, owner *types.Address, tickLower, tickUpper, amount, amount0, amount1 *big.Int) *Burn {
	return &Burn{
		Contract:  contract,
		Owner:     owner,
		TickLower: tickLower,
		TickUpper: tickUpper,
		Amount:    amount,
		Amount0:   amount0,
		Amount1:   amount1,
	}
}

// Swap mirrors IUniswapV3PoolEvents' Swap event: Sender swapped, sending the output to
// Recipient; Amount0/Amount1 are the pool's balance deltas, SqrtPriceX96/Liquidity/Tick are the
// pool's state right after the swap.
type Swap struct {
	Contract     *types.Address
	Sender       *types.Address
	Recipient    *types.Address
	Amount0      *big.Int
	Amount1      *big.Int
	SqrtPriceX96 *big.Int
	Liquidity    *big.Int
	Tick         *big.Int
}

func NewSwap(contract, sender, recipient *types.Address, amount0, amount1, sqrtPriceX96, liquidity, tick *big.Int) *Swap {
	return &Swap{
		Contract:     contract,
		Sender:       sender,
		Recipient:    recipient,
		Amount0:      amount0,
		Amount1:      amount1,
		SqrtPriceX96: sqrtPriceX96,
		Liquidity:    liquidity,
		Tick:         tick,
	}
}

// Flash mirrors IUniswapV3PoolEvents' Flash event: Sender flash-borrowed Amount0/Amount1 to
// Recipient, repaying Paid0/Paid1 (which can exceed Amount plus the fee).
type Flash struct {
	Contract  *types.Address
	Sender    *types.Address
	Recipient *types.Address
	Amount0   *big.Int
	Amount1   *big.Int
	Paid0     *big.Int
	Paid1     *big.Int
}

func NewFlash(contract, sender, recipient *types.Address, amount0, amount1, paid0, paid1 *big.Int) *Flash {
	return &Flash{
		Contract:  contract,
		Sender:    sender,
		Recipient: recipient,
		Amount0:   amount0,
		Amount1:   amount1,
		Paid0:     paid0,
		Paid1:     paid1,
	}
}

// IncreaseObservationCardinalityNext mirrors IUniswapV3PoolEvents' same-named event: the
// pool's observationCardinalityNext moved from Old to New.
type IncreaseObservationCardinalityNext struct {
	Contract *types.Address
	Old      *big.Int
	New      *big.Int
}

func NewIncreaseObservationCardinalityNext(contract *types.Address, old, new_ *big.Int) *IncreaseObservationCardinalityNext {
	return &IncreaseObservationCardinalityNext{
		Contract: contract,
		Old:      old,
		New:      new_,
	}
}

// SetFeeProtocol mirrors IUniswapV3PoolEvents' same-named event: the pool's protocol fee for
// token0/token1 moved from the Old pair to the New pair.
type SetFeeProtocol struct {
	Contract        *types.Address
	FeeProtocol0Old *big.Int
	FeeProtocol1Old *big.Int
	FeeProtocol0New *big.Int
	FeeProtocol1New *big.Int
}

func NewSetFeeProtocol(contract *types.Address, feeProtocol0Old, feeProtocol1Old, feeProtocol0New, feeProtocol1New *big.Int) *SetFeeProtocol {
	return &SetFeeProtocol{
		Contract:        contract,
		FeeProtocol0Old: feeProtocol0Old,
		FeeProtocol1Old: feeProtocol1Old,
		FeeProtocol0New: feeProtocol0New,
		FeeProtocol1New: feeProtocol1New,
	}
}

// CollectProtocol mirrors IUniswapV3PoolEvents' same-named event: Sender withdrew the
// pool's accrued protocol fees to Recipient.
type CollectProtocol struct {
	Contract  *types.Address
	Sender    *types.Address
	Recipient *types.Address
	Amount0   *big.Int
	Amount1   *big.Int
}

func NewCollectProtocol(contract, sender, recipient *types.Address, amount0, amount1 *big.Int) *CollectProtocol {
	return &CollectProtocol{
		Contract:  contract,
		Sender:    sender,
		Recipient: recipient,
		Amount0:   amount0,
		Amount1:   amount1,
	}
}
