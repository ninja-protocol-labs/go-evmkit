package v3

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
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

// FactoryRequest is factory()'s input on Contract.
type FactoryRequest struct {
	Contract *types.Address
}

func NewFactoryRequest(contract *types.Address) *FactoryRequest {
	return &FactoryRequest{
		Contract: contract,
	}
}

func (r *FactoryRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// Token0Request is token0()'s input on Contract.
type Token0Request struct {
	Contract *types.Address
}

func NewToken0Request(contract *types.Address) *Token0Request {
	return &Token0Request{
		Contract: contract,
	}
}

func (r *Token0Request) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// Token1Request is token1()'s input on Contract.
type Token1Request struct {
	Contract *types.Address
}

func NewToken1Request(contract *types.Address) *Token1Request {
	return &Token1Request{
		Contract: contract,
	}
}

func (r *Token1Request) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// FeeRequest is fee()'s input on Contract.
type FeeRequest struct {
	Contract *types.Address
}

func NewFeeRequest(contract *types.Address) *FeeRequest {
	return &FeeRequest{
		Contract: contract,
	}
}

func (r *FeeRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// TickSpacingRequest is tickSpacing()'s input on Contract.
type TickSpacingRequest struct {
	Contract *types.Address
}

func NewTickSpacingRequest(contract *types.Address) *TickSpacingRequest {
	return &TickSpacingRequest{
		Contract: contract,
	}
}

func (r *TickSpacingRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// MaxLiquidityPerTickRequest is maxLiquidityPerTick()'s input on Contract.
type MaxLiquidityPerTickRequest struct {
	Contract *types.Address
}

func NewMaxLiquidityPerTickRequest(contract *types.Address) *MaxLiquidityPerTickRequest {
	return &MaxLiquidityPerTickRequest{
		Contract: contract,
	}
}

func (r *MaxLiquidityPerTickRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// Slot0Request is slot0()'s input on Contract.
type Slot0Request struct {
	Contract *types.Address
}

func NewSlot0Request(contract *types.Address) *Slot0Request {
	return &Slot0Request{
		Contract: contract,
	}
}

func (r *Slot0Request) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// FeeGrowthGlobal0X128Request is feeGrowthGlobal0X128()'s input on Contract.
type FeeGrowthGlobal0X128Request struct {
	Contract *types.Address
}

func NewFeeGrowthGlobal0X128Request(contract *types.Address) *FeeGrowthGlobal0X128Request {
	return &FeeGrowthGlobal0X128Request{
		Contract: contract,
	}
}

func (r *FeeGrowthGlobal0X128Request) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// FeeGrowthGlobal1X128Request is feeGrowthGlobal1X128()'s input on Contract.
type FeeGrowthGlobal1X128Request struct {
	Contract *types.Address
}

func NewFeeGrowthGlobal1X128Request(contract *types.Address) *FeeGrowthGlobal1X128Request {
	return &FeeGrowthGlobal1X128Request{
		Contract: contract,
	}
}

func (r *FeeGrowthGlobal1X128Request) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// ProtocolFeesRequest is protocolFees()'s input on Contract.
type ProtocolFeesRequest struct {
	Contract *types.Address
}

func NewProtocolFeesRequest(contract *types.Address) *ProtocolFeesRequest {
	return &ProtocolFeesRequest{
		Contract: contract,
	}
}

func (r *ProtocolFeesRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// LiquidityRequest is liquidity()'s input on Contract.
type LiquidityRequest struct {
	Contract *types.Address
}

func NewLiquidityRequest(contract *types.Address) *LiquidityRequest {
	return &LiquidityRequest{
		Contract: contract,
	}
}

func (r *LiquidityRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// TicksRequest is ticks(tick)'s input on Contract.
type TicksRequest struct {
	Contract *types.Address
	Tick     *big.Int
}

func NewTicksRequest(contract *types.Address, tick *big.Int) *TicksRequest {
	return &TicksRequest{
		Contract: contract,
		Tick:     tick,
	}
}

func (r *TicksRequest) Validate() error {
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

// TickBitmapRequest is tickBitmap(wordPosition)'s input on Contract.
type TickBitmapRequest struct {
	Contract     *types.Address
	WordPosition int16
}

func NewTickBitmapRequest(contract *types.Address, wordPosition int16) *TickBitmapRequest {
	return &TickBitmapRequest{
		Contract:     contract,
		WordPosition: wordPosition,
	}
}

func (r *TickBitmapRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// PositionsRequest is positions(key)'s input on Contract.
type PositionsRequest struct {
	Contract *types.Address
	Key      [32]byte
}

func NewPositionsRequest(contract *types.Address, key [32]byte) *PositionsRequest {
	return &PositionsRequest{
		Contract: contract,
		Key:      key,
	}
}

func (r *PositionsRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// ObservationsRequest is observations(index)'s input on Contract.
type ObservationsRequest struct {
	Contract *types.Address
	Index    *big.Int
}

func NewObservationsRequest(contract *types.Address, index *big.Int) *ObservationsRequest {
	return &ObservationsRequest{
		Contract: contract,
		Index:    index,
	}
}

func (r *ObservationsRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Index == nil {
		return fmt.Errorf("%w: index is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// ObserveRequest is observe(secondsAgos)'s input on Contract.
type ObserveRequest struct {
	Contract    *types.Address
	SecondsAgos []uint32
}

func NewObserveRequest(contract *types.Address, secondsAgos []uint32) *ObserveRequest {
	return &ObserveRequest{
		Contract:    contract,
		SecondsAgos: secondsAgos,
	}
}

func (r *ObserveRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SnapshotCumulativesInsideRequest is snapshotCumulativesInside(tickLower, tickUpper)'s input on Contract.
type SnapshotCumulativesInsideRequest struct {
	Contract  *types.Address
	TickLower *big.Int
	TickUpper *big.Int
}

func NewSnapshotCumulativesInsideRequest(contract *types.Address, tickLower, tickUpper *big.Int) *SnapshotCumulativesInsideRequest {
	return &SnapshotCumulativesInsideRequest{
		Contract:  contract,
		TickLower: tickLower,
		TickUpper: tickUpper,
	}
}

func (r *SnapshotCumulativesInsideRequest) Validate() error {
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

// InitializeRequest is initialize(sqrtPriceX96)'s input on Contract.
type InitializeRequest struct {
	Contract     *types.Address
	SqrtPriceX96 *big.Int
}

func NewInitializeRequest(contract *types.Address, sqrtPriceX96 *big.Int) *InitializeRequest {
	return &InitializeRequest{
		Contract:     contract,
		SqrtPriceX96: sqrtPriceX96,
	}
}

func (r *InitializeRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.SqrtPriceX96 == nil {
		return fmt.Errorf("%w: sqrtPriceX96 is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// PoolMintRequest is mint(recipient, tickLower, tickUpper, amount, data)'s input on Contract.
type PoolMintRequest struct {
	Contract     *types.Address
	Recipient    *types.Address
	TickLower    *big.Int
	TickUpper    *big.Int
	Amount       *big.Int
	CallbackData []byte
}

func NewPoolMintRequest(contract, recipient *types.Address, tickLower, tickUpper, amount *big.Int, callbackData []byte) *PoolMintRequest {
	return &PoolMintRequest{
		Contract:     contract,
		Recipient:    recipient,
		TickLower:    tickLower,
		TickUpper:    tickUpper,
		Amount:       amount,
		CallbackData: callbackData,
	}
}

func (r *PoolMintRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	if r.TickLower == nil {
		return fmt.Errorf("%w: tickLower is nil", bind.ErrInvalidRequest)
	}

	if r.TickUpper == nil {
		return fmt.Errorf("%w: tickUpper is nil", bind.ErrInvalidRequest)
	}

	if r.Amount == nil {
		return fmt.Errorf("%w: amount is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// PoolCollectRequest is collect(recipient, tickLower, tickUpper, amount0Requested, amount1Requested)'s input on Contract.
type PoolCollectRequest struct {
	Contract         *types.Address
	Recipient        *types.Address
	TickLower        *big.Int
	TickUpper        *big.Int
	Amount0Requested *big.Int
	Amount1Requested *big.Int
}

func NewPoolCollectRequest(contract, recipient *types.Address, tickLower, tickUpper, amount0Requested, amount1Requested *big.Int) *PoolCollectRequest {
	return &PoolCollectRequest{
		Contract:         contract,
		Recipient:        recipient,
		TickLower:        tickLower,
		TickUpper:        tickUpper,
		Amount0Requested: amount0Requested,
		Amount1Requested: amount1Requested,
	}
}

func (r *PoolCollectRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	if r.TickLower == nil {
		return fmt.Errorf("%w: tickLower is nil", bind.ErrInvalidRequest)
	}

	if r.TickUpper == nil {
		return fmt.Errorf("%w: tickUpper is nil", bind.ErrInvalidRequest)
	}

	if r.Amount0Requested == nil {
		return fmt.Errorf("%w: amount0Requested is nil", bind.ErrInvalidRequest)
	}

	if r.Amount1Requested == nil {
		return fmt.Errorf("%w: amount1Requested is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// BurnRequest is burn(tickLower, tickUpper, amount)'s input on Contract.
type BurnRequest struct {
	Contract  *types.Address
	TickLower *big.Int
	TickUpper *big.Int
	Amount    *big.Int
}

func NewBurnRequest(contract *types.Address, tickLower, tickUpper, amount *big.Int) *BurnRequest {
	return &BurnRequest{
		Contract:  contract,
		TickLower: tickLower,
		TickUpper: tickUpper,
		Amount:    amount,
	}
}

func (r *BurnRequest) Validate() error {
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

	if r.Amount == nil {
		return fmt.Errorf("%w: amount is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SwapRequest is swap(recipient, zeroForOne, amountSpecified, sqrtPriceLimitX96, data)'s input on Contract.
type SwapRequest struct {
	Contract          *types.Address
	Recipient         *types.Address
	ZeroForOne        bool
	AmountSpecified   *big.Int
	SqrtPriceLimitX96 *big.Int
	CallbackData      []byte
}

func NewSwapRequest(contract, recipient *types.Address, zeroForOne bool, amountSpecified, sqrtPriceLimitX96 *big.Int, callbackData []byte) *SwapRequest {
	return &SwapRequest{
		Contract:          contract,
		Recipient:         recipient,
		ZeroForOne:        zeroForOne,
		AmountSpecified:   amountSpecified,
		SqrtPriceLimitX96: sqrtPriceLimitX96,
		CallbackData:      callbackData,
	}
}

func (r *SwapRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	if r.AmountSpecified == nil {
		return fmt.Errorf("%w: amountSpecified is nil", bind.ErrInvalidRequest)
	}

	if r.SqrtPriceLimitX96 == nil {
		return fmt.Errorf("%w: sqrtPriceLimitX96 is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// FlashRequest is flash(recipient, amount0, amount1, data)'s input on Contract.
type FlashRequest struct {
	Contract     *types.Address
	Recipient    *types.Address
	Amount0      *big.Int
	Amount1      *big.Int
	CallbackData []byte
}

func NewFlashRequest(contract, recipient *types.Address, amount0, amount1 *big.Int, callbackData []byte) *FlashRequest {
	return &FlashRequest{
		Contract:     contract,
		Recipient:    recipient,
		Amount0:      amount0,
		Amount1:      amount1,
		CallbackData: callbackData,
	}
}

func (r *FlashRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	if r.Amount0 == nil {
		return fmt.Errorf("%w: amount0 is nil", bind.ErrInvalidRequest)
	}

	if r.Amount1 == nil {
		return fmt.Errorf("%w: amount1 is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// IncreaseObservationCardinalityNextRequest is increaseObservationCardinalityNext(observationCardinalityNext)'s input on Contract.
type IncreaseObservationCardinalityNextRequest struct {
	Contract                   *types.Address
	ObservationCardinalityNext uint16
}

func NewIncreaseObservationCardinalityNextRequest(contract *types.Address, observationCardinalityNext uint16) *IncreaseObservationCardinalityNextRequest {
	return &IncreaseObservationCardinalityNextRequest{
		Contract:                   contract,
		ObservationCardinalityNext: observationCardinalityNext,
	}
}

func (r *IncreaseObservationCardinalityNextRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SetFeeProtocolRequest is setFeeProtocol(feeProtocol0, feeProtocol1)'s input on Contract.
type SetFeeProtocolRequest struct {
	Contract     *types.Address
	FeeProtocol0 uint8
	FeeProtocol1 uint8
}

func NewSetFeeProtocolRequest(contract *types.Address, feeProtocol0, feeProtocol1 uint8) *SetFeeProtocolRequest {
	return &SetFeeProtocolRequest{
		Contract:     contract,
		FeeProtocol0: feeProtocol0,
		FeeProtocol1: feeProtocol1,
	}
}

func (r *SetFeeProtocolRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// CollectProtocolRequest is collectProtocol(recipient, amount0Requested, amount1Requested)'s input on Contract.
type CollectProtocolRequest struct {
	Contract         *types.Address
	Recipient        *types.Address
	Amount0Requested *big.Int
	Amount1Requested *big.Int
}

func NewCollectProtocolRequest(contract, recipient *types.Address, amount0Requested, amount1Requested *big.Int) *CollectProtocolRequest {
	return &CollectProtocolRequest{
		Contract:         contract,
		Recipient:        recipient,
		Amount0Requested: amount0Requested,
		Amount1Requested: amount1Requested,
	}
}

func (r *CollectProtocolRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	if r.Amount0Requested == nil {
		return fmt.Errorf("%w: amount0Requested is nil", bind.ErrInvalidRequest)
	}

	if r.Amount1Requested == nil {
		return fmt.Errorf("%w: amount1Requested is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// InitializeEvent mirrors IUniswapV3PoolEvents' Initialize event: Contract was initialized at SqrtPriceX96/Tick.
type InitializeEvent struct {
	Contract     *types.Address
	SqrtPriceX96 *big.Int
	Tick         *big.Int
}

func NewInitializeEvent(contract *types.Address, sqrtPriceX96, tick *big.Int) *InitializeEvent {
	return &InitializeEvent{
		Contract:     contract,
		SqrtPriceX96: sqrtPriceX96,
		Tick:         tick,
	}
}

// MintEvent mirrors IUniswapV3PoolEvents' Mint event: Sender minted Amount liquidity into Owner's
// [TickLower, TickUpper] position, paying Amount0/Amount1.
type MintEvent struct {
	Contract  *types.Address
	Sender    *types.Address
	Owner     *types.Address
	TickLower *big.Int
	TickUpper *big.Int
	Amount    *big.Int
	Amount0   *big.Int
	Amount1   *big.Int
}

func NewMintEvent(contract, sender, owner *types.Address, tickLower, tickUpper, amount, amount0, amount1 *big.Int) *MintEvent {
	return &MintEvent{
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

// CollectEvent mirrors IUniswapV3PoolEvents' Collect event: fees owed to Owner's [TickLower,
// TickUpper] position were sent to Recipient.
type CollectEvent struct {
	Contract  *types.Address
	Owner     *types.Address
	Recipient *types.Address
	TickLower *big.Int
	TickUpper *big.Int
	Amount0   *big.Int
	Amount1   *big.Int
}

func NewCollectEvent(contract, owner, recipient *types.Address, tickLower, tickUpper, amount0, amount1 *big.Int) *CollectEvent {
	return &CollectEvent{
		Contract:  contract,
		Owner:     owner,
		Recipient: recipient,
		TickLower: tickLower,
		TickUpper: tickUpper,
		Amount0:   amount0,
		Amount1:   amount1,
	}
}

// BurnEvent mirrors IUniswapV3PoolEvents' Burn event: Amount liquidity was removed from Owner's
// [TickLower, TickUpper] position, returning Amount0/Amount1 (fees must be withdrawn separately
// via Collect).
type BurnEvent struct {
	Contract  *types.Address
	Owner     *types.Address
	TickLower *big.Int
	TickUpper *big.Int
	Amount    *big.Int
	Amount0   *big.Int
	Amount1   *big.Int
}

func NewBurnEvent(contract, owner *types.Address, tickLower, tickUpper, amount, amount0, amount1 *big.Int) *BurnEvent {
	return &BurnEvent{
		Contract:  contract,
		Owner:     owner,
		TickLower: tickLower,
		TickUpper: tickUpper,
		Amount:    amount,
		Amount0:   amount0,
		Amount1:   amount1,
	}
}

// SwapEvent mirrors IUniswapV3PoolEvents' Swap event: Sender swapped, sending the output to
// Recipient; Amount0/Amount1 are the pool's balance deltas, SqrtPriceX96/Liquidity/Tick are the
// pool's state right after the swap.
type SwapEvent struct {
	Contract     *types.Address
	Sender       *types.Address
	Recipient    *types.Address
	Amount0      *big.Int
	Amount1      *big.Int
	SqrtPriceX96 *big.Int
	Liquidity    *big.Int
	Tick         *big.Int
}

func NewSwapEvent(contract, sender, recipient *types.Address, amount0, amount1, sqrtPriceX96, liquidity, tick *big.Int) *SwapEvent {
	return &SwapEvent{
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

// FlashEvent mirrors IUniswapV3PoolEvents' Flash event: Sender flash-borrowed Amount0/Amount1 to
// Recipient, repaying Paid0/Paid1 (which can exceed Amount plus the fee).
type FlashEvent struct {
	Contract  *types.Address
	Sender    *types.Address
	Recipient *types.Address
	Amount0   *big.Int
	Amount1   *big.Int
	Paid0     *big.Int
	Paid1     *big.Int
}

func NewFlashEvent(contract, sender, recipient *types.Address, amount0, amount1, paid0, paid1 *big.Int) *FlashEvent {
	return &FlashEvent{
		Contract:  contract,
		Sender:    sender,
		Recipient: recipient,
		Amount0:   amount0,
		Amount1:   amount1,
		Paid0:     paid0,
		Paid1:     paid1,
	}
}

// IncreaseObservationCardinalityNextEvent mirrors IUniswapV3PoolEvents' same-named event: the
// pool's observationCardinalityNext moved from Old to New.
type IncreaseObservationCardinalityNextEvent struct {
	Contract *types.Address
	Old      *big.Int
	New      *big.Int
}

func NewIncreaseObservationCardinalityNextEvent(contract *types.Address, old, new_ *big.Int) *IncreaseObservationCardinalityNextEvent {
	return &IncreaseObservationCardinalityNextEvent{
		Contract: contract,
		Old:      old,
		New:      new_,
	}
}

// SetFeeProtocolEvent mirrors IUniswapV3PoolEvents' same-named event: the pool's protocol fee for
// token0/token1 moved from the Old pair to the New pair.
type SetFeeProtocolEvent struct {
	Contract        *types.Address
	FeeProtocol0Old *big.Int
	FeeProtocol1Old *big.Int
	FeeProtocol0New *big.Int
	FeeProtocol1New *big.Int
}

func NewSetFeeProtocolEvent(contract *types.Address, feeProtocol0Old, feeProtocol1Old, feeProtocol0New, feeProtocol1New *big.Int) *SetFeeProtocolEvent {
	return &SetFeeProtocolEvent{
		Contract:        contract,
		FeeProtocol0Old: feeProtocol0Old,
		FeeProtocol1Old: feeProtocol1Old,
		FeeProtocol0New: feeProtocol0New,
		FeeProtocol1New: feeProtocol1New,
	}
}

// CollectProtocolEvent mirrors IUniswapV3PoolEvents' same-named event: Sender withdrew the
// pool's accrued protocol fees to Recipient.
type CollectProtocolEvent struct {
	Contract  *types.Address
	Sender    *types.Address
	Recipient *types.Address
	Amount0   *big.Int
	Amount1   *big.Int
}

func NewCollectProtocolEvent(contract, sender, recipient *types.Address, amount0, amount1 *big.Int) *CollectProtocolEvent {
	return &CollectProtocolEvent{
		Contract:  contract,
		Sender:    sender,
		Recipient: recipient,
		Amount0:   amount0,
		Amount1:   amount1,
	}
}
