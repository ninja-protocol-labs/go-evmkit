package v3

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Function signatures from IUniswapV3PoolImmutables/State/DerivedState/Actions/OwnerActions,
// for encoding calls and decoding results.
var (
	// Immutables
	poolFactoryFn         = abi.NewFunction("factory", nil, abi.NewTypes(abi.Address))
	poolToken0Fn          = abi.NewFunction("token0", nil, abi.NewTypes(abi.Address))
	poolToken1Fn          = abi.NewFunction("token1", nil, abi.NewTypes(abi.Address))
	poolFeeFn             = abi.NewFunction("fee", nil, abi.NewTypes(abi.Uint24))
	poolTickSpacingFn     = abi.NewFunction("tickSpacing", nil, abi.NewTypes(abi.Int24))
	maxLiquidityPerTickFn = abi.NewFunction("maxLiquidityPerTick", nil, abi.NewTypes(abi.Uint128))

	// State
	slot0Fn                = abi.NewFunction("slot0", nil, abi.NewTypes(abi.Uint160, abi.Int24, abi.Uint16, abi.Uint16, abi.Uint16, abi.Uint8, abi.Bool))
	feeGrowthGlobal0X128Fn = abi.NewFunction("feeGrowthGlobal0X128", nil, abi.NewTypes(abi.Uint256))
	feeGrowthGlobal1X128Fn = abi.NewFunction("feeGrowthGlobal1X128", nil, abi.NewTypes(abi.Uint256))
	protocolFeesFn         = abi.NewFunction("protocolFees", nil, abi.NewTypes(abi.Uint128, abi.Uint128))
	poolLiquidityFn        = abi.NewFunction("liquidity", nil, abi.NewTypes(abi.Uint128))
	ticksFn                = abi.NewFunction("ticks", abi.NewTypes(abi.Int24), abi.NewTypes(abi.Uint128, abi.Int128, abi.Uint256, abi.Uint256, abi.Int56, abi.Uint160, abi.Uint32, abi.Bool))
	tickBitmapFn           = abi.NewFunction("tickBitmap", abi.NewTypes(abi.Int16), abi.NewTypes(abi.Uint256))
	positionsFn            = abi.NewFunction("positions", abi.NewTypes(abi.Bytes32), abi.NewTypes(abi.Uint128, abi.Uint256, abi.Uint256, abi.Uint128, abi.Uint128))
	poolObservationsFn     = abi.NewFunction("observations", abi.NewTypes(abi.Uint256), abi.NewTypes(abi.Uint32, abi.Int56, abi.Uint160, abi.Bool))

	// DerivedState
	observeFn                   = abi.NewFunction("observe", abi.NewTypes(abi.Slice(abi.Uint32)), abi.NewTypes(abi.Slice(abi.Int56), abi.Slice(abi.Uint160)))
	snapshotCumulativesInsideFn = abi.NewFunction("snapshotCumulativesInside", abi.NewTypes(abi.Int24, abi.Int24), abi.NewTypes(abi.Int56, abi.Uint160, abi.Uint32))

	// Actions
	initializeFn                         = abi.NewFunction("initialize", abi.NewTypes(abi.Uint160), nil)
	mintFn                               = abi.NewFunction("mint", abi.NewTypes(abi.Address, abi.Int24, abi.Int24, abi.Uint128, abi.Bytes), abi.NewTypes(abi.Uint256, abi.Uint256))
	collectFn                            = abi.NewFunction("collect", abi.NewTypes(abi.Address, abi.Int24, abi.Int24, abi.Uint128, abi.Uint128), abi.NewTypes(abi.Uint128, abi.Uint128))
	burnFn                               = abi.NewFunction("burn", abi.NewTypes(abi.Int24, abi.Int24, abi.Uint128), abi.NewTypes(abi.Uint256, abi.Uint256))
	swapFn                               = abi.NewFunction("swap", abi.NewTypes(abi.Address, abi.Bool, abi.Int256, abi.Uint160, abi.Bytes), abi.NewTypes(abi.Int256, abi.Int256))
	flashFn                              = abi.NewFunction("flash", abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256, abi.Bytes), nil)
	increaseObservationCardinalityNextFn = abi.NewFunction("increaseObservationCardinalityNext", abi.NewTypes(abi.Uint16), nil)

	// OwnerActions
	setFeeProtocolFn  = abi.NewFunction("setFeeProtocol", abi.NewTypes(abi.Uint8, abi.Uint8), nil)
	collectProtocolFn = abi.NewFunction("collectProtocol", abi.NewTypes(abi.Address, abi.Uint128, abi.Uint128), abi.NewTypes(abi.Uint128, abi.Uint128))
)

// Events from IUniswapV3PoolEvents, for decoding logs via Event.Decode.
var (
	initializeEvent = abi.NewEvent("Initialize", []abi.EventParam{
		{Type: abi.Uint160},
		{Type: abi.Int24},
	}, false)

	mintEvent = abi.NewEvent("Mint", []abi.EventParam{
		{Type: abi.Address},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Int24, Indexed: true},
		{Type: abi.Int24, Indexed: true},
		{Type: abi.Uint128},
		{Type: abi.Uint256},
		{Type: abi.Uint256},
	}, false)

	collectEvent = abi.NewEvent("Collect", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address},
		{Type: abi.Int24, Indexed: true},
		{Type: abi.Int24, Indexed: true},
		{Type: abi.Uint128},
		{Type: abi.Uint128},
	}, false)

	poolBurnEvent = abi.NewEvent("Burn", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Int24, Indexed: true},
		{Type: abi.Int24, Indexed: true},
		{Type: abi.Uint128},
		{Type: abi.Uint256},
		{Type: abi.Uint256},
	}, false)

	swapEvent = abi.NewEvent("Swap", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Int256},
		{Type: abi.Int256},
		{Type: abi.Uint160},
		{Type: abi.Uint128},
		{Type: abi.Int24},
	}, false)

	flashEvent = abi.NewEvent("Flash", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Uint256},
		{Type: abi.Uint256},
		{Type: abi.Uint256},
		{Type: abi.Uint256},
	}, false)

	increaseObservationCardinalityNextEvent = abi.NewEvent("IncreaseObservationCardinalityNext", []abi.EventParam{
		{Type: abi.Uint16},
		{Type: abi.Uint16},
	}, false)

	setFeeProtocolEvent = abi.NewEvent("SetFeeProtocol", []abi.EventParam{
		{Type: abi.Uint8},
		{Type: abi.Uint8},
		{Type: abi.Uint8},
		{Type: abi.Uint8},
	}, false)

	collectProtocolEvent = abi.NewEvent("CollectProtocol", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Uint128},
		{Type: abi.Uint128},
	}, false)
)

// UniswapV3Pool's require reasons, decoded via abi.DecodeRevert -- unlike UniswapV3Factory,
// most (not all) of Pool's requires carry a short reason string. Some of these live not in
// UniswapV3Pool.sol itself but in a library it imports (Position/Tick/TickMath/Oracle/
// LiquidityMath/TransferHelper) -- a library's require still reverts as if it were Pool's own,
// since "using X for Y" inlines the library's code into Pool's call context.
const (
	reasonLocked                           = "LOK"
	reasonTickLowerGreaterThanOrEqualUpper = "TLU"
	reasonTickLowerTooLow                  = "TLM"
	reasonTickUpperTooHigh                 = "TUM"
	reasonAlreadyInitialized               = "AI"
	reasonMintToken0Underpaid              = "M0"
	reasonMintToken1Underpaid              = "M1"
	reasonAmountSpecifiedZero              = "AS"
	reasonInvalidSqrtPriceLimit            = "SPL"
	reasonSwapUnderpaid                    = "IIA"
	reasonNoLiquidity                      = "L"
	reasonFlashToken0Underpaid             = "F0"
	reasonFlashToken1Underpaid             = "F1"
	reasonLiquiditySubUnderflow            = "LS"
	reasonLiquidityAddOverflow             = "LA"
	reasonOracleNotInitialized             = "I"
	reasonOracleObservationTooOld          = "OLD"
	reasonNoPosition                       = "NP"
	reasonLiquidityOverflow                = "LO"
	reasonTickOutOfBounds                  = "T"
	reasonSqrtRatioOutOfBounds             = "R"
	reasonTransferFailed                   = "TF"
)

// ErrLocked, etc. mirror UniswapV3Pool's revert reasons, for errors.Is. ErrReverted catches
// every reasonless require (the onlyFactoryOwner/zero-amount checks and internal invariants),
// same as UniswapV3Factory.
var (
	ErrLocked                      = errors.New("uniswapv3: pool locked")
	ErrTickLowerGreaterThanOrEqual = errors.New("uniswapv3: tickLower >= tickUpper")
	ErrTickLowerTooLow             = errors.New("uniswapv3: tickLower below MIN_TICK")
	ErrTickUpperTooHigh            = errors.New("uniswapv3: tickUpper above MAX_TICK")
	ErrAlreadyInitialized          = errors.New("uniswapv3: pool already initialized")
	ErrMintToken0Underpaid         = errors.New("uniswapv3: mint callback underpaid token0")
	ErrMintToken1Underpaid         = errors.New("uniswapv3: mint callback underpaid token1")
	ErrAmountSpecifiedZero         = errors.New("uniswapv3: amountSpecified is zero")
	ErrInvalidSqrtPriceLimit       = errors.New("uniswapv3: sqrtPriceLimitX96 invalid for swap direction")
	ErrSwapUnderpaid               = errors.New("uniswapv3: swap callback underpaid")
	ErrNoLiquidity                 = errors.New("uniswapv3: pool has no liquidity")
	ErrFlashToken0Underpaid        = errors.New("uniswapv3: flash callback underpaid token0 fee")
	ErrFlashToken1Underpaid        = errors.New("uniswapv3: flash callback underpaid token1 fee")
	ErrLiquiditySubUnderflow       = errors.New("uniswapv3: liquidity delta underflows")
	ErrLiquidityAddOverflow        = errors.New("uniswapv3: liquidity delta overflows")
	ErrOracleNotInitialized        = errors.New("uniswapv3: oracle not initialized")
	ErrOracleObservationTooOld     = errors.New("uniswapv3: oracle observation older than the oldest stored")
	ErrNoPosition                  = errors.New("uniswapv3: no liquidity in position, poke disallowed")
	ErrLiquidityOverflow           = errors.New("uniswapv3: tick liquidity exceeds maxLiquidityPerTick")
	ErrTickOutOfBounds             = errors.New("uniswapv3: tick outside MIN_TICK/MAX_TICK")
	ErrSqrtRatioOutOfBounds        = errors.New("uniswapv3: sqrtPriceX96 outside MIN_SQRT_RATIO/MAX_SQRT_RATIO")
	ErrTransferFailed              = errors.New("uniswapv3: token transfer failed")
)

// EncodeFactory returns the calldata for factory().
func EncodeFactory() []byte {
	return poolFactoryFn.SelectorBytes()
}

// DecodeFactory decodes the return data of factory.
func DecodeFactory(data []byte) (*types.Address, error) {
	addr, err := poolFactoryFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeFactory: %w", err)
	}
	return addr, nil
}

// EncodeToken0 returns the calldata for token0().
func EncodeToken0() []byte {
	return poolToken0Fn.SelectorBytes()
}

// DecodeToken0 decodes the return data of token0.
func DecodeToken0(data []byte) (*types.Address, error) {
	addr, err := poolToken0Fn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeToken0: %w", err)
	}
	return addr, nil
}

// EncodeToken1 returns the calldata for token1().
func EncodeToken1() []byte {
	return poolToken1Fn.SelectorBytes()
}

// DecodeToken1 decodes the return data of token1.
func DecodeToken1(data []byte) (*types.Address, error) {
	addr, err := poolToken1Fn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeToken1: %w", err)
	}
	return addr, nil
}

// EncodePoolFee returns the calldata for fee().
func EncodePoolFee() []byte {
	return poolFeeFn.SelectorBytes()
}

// DecodePoolFee decodes the return data of fee.
func DecodePoolFee(data []byte) (*big.Int, error) {
	n, err := poolFeeFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodePoolFee: %w", err)
	}
	return n, nil
}

// EncodePoolTickSpacing returns the calldata for tickSpacing().
func EncodePoolTickSpacing() []byte {
	return poolTickSpacingFn.SelectorBytes()
}

// DecodePoolTickSpacing decodes the return data of tickSpacing.
func DecodePoolTickSpacing(data []byte) (*big.Int, error) {
	n, err := poolTickSpacingFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodePoolTickSpacing: %w", err)
	}
	return n, nil
}

// EncodeMaxLiquidityPerTick returns the calldata for maxLiquidityPerTick().
func EncodeMaxLiquidityPerTick() []byte {
	return maxLiquidityPerTickFn.SelectorBytes()
}

// DecodeMaxLiquidityPerTick decodes the return data of maxLiquidityPerTick.
func DecodeMaxLiquidityPerTick(data []byte) (*big.Int, error) {
	n, err := maxLiquidityPerTickFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeMaxLiquidityPerTick: %w", err)
	}
	return n, nil
}

// EncodeSlot0 returns the calldata for slot0().
func EncodeSlot0() []byte {
	return slot0Fn.SelectorBytes()
}

// DecodeSlot0 decodes the return data of slot0.
func DecodeSlot0(data []byte) (*Slot0, error) {
	vals, err := slot0Fn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeSlot0: %w", err)
	}
	if len(vals) != 7 {
		return nil, fmt.Errorf("uniswapv3: decodeSlot0: expected 7 return values, got %d", len(vals))
	}

	sqrtPriceX96, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeSlot0: unexpected decoded type for sqrtPriceX96: %T", vals[0])
	}
	tick, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeSlot0: unexpected decoded type for tick: %T", vals[1])
	}
	observationIndex, ok := vals[2].(uint16)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeSlot0: unexpected decoded type for observationIndex: %T", vals[2])
	}
	observationCardinality, ok := vals[3].(uint16)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeSlot0: unexpected decoded type for observationCardinality: %T", vals[3])
	}
	observationCardinalityNext, ok := vals[4].(uint16)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeSlot0: unexpected decoded type for observationCardinalityNext: %T", vals[4])
	}
	feeProtocol, ok := vals[5].(uint8)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeSlot0: unexpected decoded type for feeProtocol: %T", vals[5])
	}
	unlocked, ok := vals[6].(bool)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeSlot0: unexpected decoded type for unlocked: %T", vals[6])
	}

	return NewSlot0(
		sqrtPriceX96,
		tick,
		big.NewInt(int64(observationIndex)),
		big.NewInt(int64(observationCardinality)),
		big.NewInt(int64(observationCardinalityNext)),
		big.NewInt(int64(feeProtocol)),
		unlocked,
	), nil
}

// EncodeFeeGrowthGlobal0X128 returns the calldata for feeGrowthGlobal0X128().
func EncodeFeeGrowthGlobal0X128() []byte {
	return feeGrowthGlobal0X128Fn.SelectorBytes()
}

// DecodeFeeGrowthGlobal0X128 decodes the return data of feeGrowthGlobal0X128.
func DecodeFeeGrowthGlobal0X128(data []byte) (*big.Int, error) {
	n, err := feeGrowthGlobal0X128Fn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeFeeGrowthGlobal0X128: %w", err)
	}
	return n, nil
}

// EncodeFeeGrowthGlobal1X128 returns the calldata for feeGrowthGlobal1X128().
func EncodeFeeGrowthGlobal1X128() []byte {
	return feeGrowthGlobal1X128Fn.SelectorBytes()
}

// DecodeFeeGrowthGlobal1X128 decodes the return data of feeGrowthGlobal1X128.
func DecodeFeeGrowthGlobal1X128(data []byte) (*big.Int, error) {
	n, err := feeGrowthGlobal1X128Fn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeFeeGrowthGlobal1X128: %w", err)
	}
	return n, nil
}

// EncodeProtocolFees returns the calldata for protocolFees().
func EncodeProtocolFees() []byte {
	return protocolFeesFn.SelectorBytes()
}

// DecodeProtocolFees decodes the return data of protocolFees.
func DecodeProtocolFees(data []byte) (*ProtocolFees, error) {
	vals, err := protocolFeesFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeProtocolFees: %w", err)
	}
	if len(vals) != 2 {
		return nil, fmt.Errorf("uniswapv3: decodeProtocolFees: expected 2 return values, got %d", len(vals))
	}
	token0, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeProtocolFees: unexpected decoded type for token0: %T", vals[0])
	}
	token1, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeProtocolFees: unexpected decoded type for token1: %T", vals[1])
	}
	return NewProtocolFees(token0, token1), nil
}

// EncodePoolLiquidity returns the calldata for liquidity().
func EncodePoolLiquidity() []byte {
	return poolLiquidityFn.SelectorBytes()
}

// DecodePoolLiquidity decodes the return data of liquidity.
func DecodePoolLiquidity(data []byte) (*big.Int, error) {
	n, err := poolLiquidityFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodePoolLiquidity: %w", err)
	}
	return n, nil
}

// EncodeTicks returns the calldata for ticks(tick).
func EncodeTicks(tick *big.Int) ([]byte, error) {
	data, err := ticksFn.EncodeCall(tick)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeTicks: %w", err)
	}
	return data, nil
}

// DecodeTicks decodes the return data of ticks.
func DecodeTicks(data []byte) (*TickInfo, error) {
	vals, err := ticksFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeTicks: %w", err)
	}
	if len(vals) != 8 {
		return nil, fmt.Errorf("uniswapv3: decodeTicks: expected 8 return values, got %d", len(vals))
	}

	liquidityGross, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeTicks: unexpected decoded type for liquidityGross: %T", vals[0])
	}
	liquidityNet, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeTicks: unexpected decoded type for liquidityNet: %T", vals[1])
	}
	feeGrowthOutside0X128, ok := vals[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeTicks: unexpected decoded type for feeGrowthOutside0X128: %T", vals[2])
	}
	feeGrowthOutside1X128, ok := vals[3].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeTicks: unexpected decoded type for feeGrowthOutside1X128: %T", vals[3])
	}
	tickCumulativeOutside, ok := vals[4].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeTicks: unexpected decoded type for tickCumulativeOutside: %T", vals[4])
	}
	secondsPerLiquidityOutsideX128, ok := vals[5].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeTicks: unexpected decoded type for secondsPerLiquidityOutsideX128: %T", vals[5])
	}
	// secondsOutside is uint32 in the ABI, not uint256 -- decodes to its native Go type, not *big.Int.
	secondsOutside, ok := vals[6].(uint32)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeTicks: unexpected decoded type for secondsOutside: %T", vals[6])
	}
	initialized, ok := vals[7].(bool)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeTicks: unexpected decoded type for initialized: %T", vals[7])
	}

	return NewTickInfo(
		liquidityGross,
		liquidityNet,
		feeGrowthOutside0X128,
		feeGrowthOutside1X128,
		tickCumulativeOutside,
		secondsPerLiquidityOutsideX128,
		big.NewInt(int64(secondsOutside)),
		initialized,
	), nil
}

// EncodeTickBitmap returns the calldata for tickBitmap(wordPosition). wordPosition is int16 in
// the ABI, not int256 -- abi.encodeInt requires the matching native Go type for widths of
// 8/16/32/64 bits, not *big.Int.
func EncodeTickBitmap(wordPosition int16) ([]byte, error) {
	data, err := tickBitmapFn.EncodeCall(wordPosition)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeTickBitmap: %w", err)
	}
	return data, nil
}

// DecodeTickBitmap decodes the return data of tickBitmap.
func DecodeTickBitmap(data []byte) (*big.Int, error) {
	n, err := tickBitmapFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeTickBitmap: %w", err)
	}
	return n, nil
}

// EncodePositions returns the calldata for positions(key).
func EncodePositions(key [32]byte) ([]byte, error) {
	data, err := positionsFn.EncodeCall(key[:])
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodePositions: %w", err)
	}
	return data, nil
}

// DecodePositions decodes the return data of positions.
func DecodePositions(data []byte) (*PositionInfo, error) {
	vals, err := positionsFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodePositions: %w", err)
	}
	if len(vals) != 5 {
		return nil, fmt.Errorf("uniswapv3: decodePositions: expected 5 return values, got %d", len(vals))
	}

	liquidity, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodePositions: unexpected decoded type for liquidity: %T", vals[0])
	}
	feeGrowthInside0LastX128, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodePositions: unexpected decoded type for feeGrowthInside0LastX128: %T", vals[1])
	}
	feeGrowthInside1LastX128, ok := vals[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodePositions: unexpected decoded type for feeGrowthInside1LastX128: %T", vals[2])
	}
	tokensOwed0, ok := vals[3].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodePositions: unexpected decoded type for tokensOwed0: %T", vals[3])
	}
	tokensOwed1, ok := vals[4].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodePositions: unexpected decoded type for tokensOwed1: %T", vals[4])
	}
	return NewPositionInfo(liquidity, feeGrowthInside0LastX128, feeGrowthInside1LastX128, tokensOwed0, tokensOwed1), nil
}

// EncodePoolObservations returns the calldata for observations(index).
func EncodePoolObservations(index *big.Int) ([]byte, error) {
	data, err := poolObservationsFn.EncodeCall(index)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodePoolObservations: %w", err)
	}
	return data, nil
}

// DecodePoolObservations decodes the return data of observations.
func DecodePoolObservations(data []byte) (*ObservationInfo, error) {
	vals, err := poolObservationsFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodePoolObservations: %w", err)
	}
	if len(vals) != 4 {
		return nil, fmt.Errorf("uniswapv3: decodePoolObservations: expected 4 return values, got %d", len(vals))
	}

	// blockTimestamp is uint32 in the ABI, not uint256 -- decodes to its native Go type, not *big.Int.
	blockTimestamp, ok := vals[0].(uint32)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodePoolObservations: unexpected decoded type for blockTimestamp: %T", vals[0])
	}
	tickCumulative, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodePoolObservations: unexpected decoded type for tickCumulative: %T", vals[1])
	}
	secondsPerLiquidityCumulativeX128, ok := vals[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodePoolObservations: unexpected decoded type for secondsPerLiquidityCumulativeX128: %T", vals[2])
	}
	initialized, ok := vals[3].(bool)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodePoolObservations: unexpected decoded type for initialized: %T", vals[3])
	}
	return NewObservationInfo(big.NewInt(int64(blockTimestamp)), tickCumulative, secondsPerLiquidityCumulativeX128, initialized), nil
}

// EncodeObserve returns the calldata for observe(secondsAgos). secondsAgos is uint32[] in the
// ABI, not uint256[] -- abi.encodeUint requires the matching native Go type for widths of
// 8/16/32/64 bits, not *big.Int.
func EncodeObserve(secondsAgos []uint32) ([]byte, error) {
	data, err := observeFn.EncodeCall(secondsAgos)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeObserve: %w", err)
	}
	return data, nil
}

// DecodeObserve decodes the return data of observe.
func DecodeObserve(data []byte) (*ObserveResult, error) {
	vals, err := observeFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeObserve: %w", err)
	}
	if len(vals) != 2 {
		return nil, fmt.Errorf("uniswapv3: decodeObserve: expected 2 return values, got %d", len(vals))
	}
	rawTickCumulatives, ok := vals[0].([]any)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeObserve: unexpected decoded type for tickCumulatives: %T", vals[0])
	}
	tickCumulatives := make([]*big.Int, len(rawTickCumulatives))
	for i, v := range rawTickCumulatives {
		n, ok := v.(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: decodeObserve: unexpected decoded type for tickCumulatives[%d]: %T", i, v)
		}
		tickCumulatives[i] = n
	}

	rawSecondsPerLiquidity, ok := vals[1].([]any)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeObserve: unexpected decoded type for secondsPerLiquidityCumulativeX128s: %T", vals[1])
	}
	secondsPerLiquidity := make([]*big.Int, len(rawSecondsPerLiquidity))
	for i, v := range rawSecondsPerLiquidity {
		n, ok := v.(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: decodeObserve: unexpected decoded type for secondsPerLiquidityCumulativeX128s[%d]: %T", i, v)
		}
		secondsPerLiquidity[i] = n
	}

	return NewObserveResult(tickCumulatives, secondsPerLiquidity), nil
}

// EncodeSnapshotCumulativesInside returns the calldata for snapshotCumulativesInside(tickLower, tickUpper).
func EncodeSnapshotCumulativesInside(tickLower, tickUpper *big.Int) ([]byte, error) {
	data, err := snapshotCumulativesInsideFn.EncodeCall(tickLower, tickUpper)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeSnapshotCumulativesInside: %w", err)
	}
	return data, nil
}

// DecodeSnapshotCumulativesInside decodes the return data of snapshotCumulativesInside.
func DecodeSnapshotCumulativesInside(data []byte) (*SnapshotCumulativesInside, error) {
	vals, err := snapshotCumulativesInsideFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeSnapshotCumulativesInside: %w", err)
	}
	if len(vals) != 3 {
		return nil, fmt.Errorf("uniswapv3: decodeSnapshotCumulativesInside: expected 3 return values, got %d", len(vals))
	}

	tickCumulativeInside, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeSnapshotCumulativesInside: unexpected decoded type for tickCumulativeInside: %T", vals[0])
	}
	secondsPerLiquidityInsideX128, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeSnapshotCumulativesInside: unexpected decoded type for secondsPerLiquidityInsideX128: %T", vals[1])
	}
	// secondsInside is uint32 in the ABI, not uint256 -- decodes to its native Go type, not *big.Int.
	secondsInside, ok := vals[2].(uint32)
	if !ok {
		return nil, fmt.Errorf("uniswapv3: decodeSnapshotCumulativesInside: unexpected decoded type for secondsInside: %T", vals[2])
	}
	return NewSnapshotCumulativesInside(tickCumulativeInside, secondsPerLiquidityInsideX128, big.NewInt(int64(secondsInside))), nil
}

// EncodeInitialize returns the calldata for initialize(sqrtPriceX96).
func EncodeInitialize(sqrtPriceX96 *big.Int) ([]byte, error) {
	data, err := initializeFn.EncodeCall(sqrtPriceX96)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeInitialize: %w", err)
	}
	return data, nil
}

// EncodeMint returns the calldata for mint(recipient, tickLower, tickUpper, amount, data).
func EncodeMint(recipient *types.Address, tickLower, tickUpper, amount *big.Int, callbackData []byte) ([]byte, error) {
	data, err := mintFn.EncodeCall(recipient, tickLower, tickUpper, amount, callbackData)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeMint: %w", err)
	}
	return data, nil
}

// DecodeMint decodes the return data of mint -- the amounts of token0/token1 paid.
func DecodeMint(data []byte) (amount0, amount1 *big.Int, err error) {
	vals, err := mintFn.DecodeReturn(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: decodeMint: %w", err)
	}
	if len(vals) != 2 {
		return nil, nil, fmt.Errorf("uniswapv3: decodeMint: expected 2 return values, got %d", len(vals))
	}
	amount0, ok := vals[0].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv3: decodeMint: unexpected decoded type for amount0: %T", vals[0])
	}
	amount1, ok = vals[1].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv3: decodeMint: unexpected decoded type for amount1: %T", vals[1])
	}
	return amount0, amount1, nil
}

// EncodeCollect returns the calldata for collect(recipient, tickLower, tickUpper, amount0Requested, amount1Requested).
func EncodeCollect(recipient *types.Address, tickLower, tickUpper, amount0Requested, amount1Requested *big.Int) ([]byte, error) {
	data, err := collectFn.EncodeCall(recipient, tickLower, tickUpper, amount0Requested, amount1Requested)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeCollect: %w", err)
	}
	return data, nil
}

// DecodeCollect decodes the return data of collect -- the amounts of token0/token1 collected.
func DecodeCollect(data []byte) (amount0, amount1 *big.Int, err error) {
	vals, err := collectFn.DecodeReturn(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: decodeCollect: %w", err)
	}
	if len(vals) != 2 {
		return nil, nil, fmt.Errorf("uniswapv3: decodeCollect: expected 2 return values, got %d", len(vals))
	}
	amount0, ok := vals[0].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv3: decodeCollect: unexpected decoded type for amount0: %T", vals[0])
	}
	amount1, ok = vals[1].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv3: decodeCollect: unexpected decoded type for amount1: %T", vals[1])
	}
	return amount0, amount1, nil
}

// EncodeBurn returns the calldata for burn(tickLower, tickUpper, amount).
func EncodeBurn(tickLower, tickUpper, amount *big.Int) ([]byte, error) {
	data, err := burnFn.EncodeCall(tickLower, tickUpper, amount)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeBurn: %w", err)
	}
	return data, nil
}

// DecodeBurn decodes the return data of burn -- the amounts of token0/token1 owed to the caller.
func DecodeBurn(data []byte) (amount0, amount1 *big.Int, err error) {
	vals, err := burnFn.DecodeReturn(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: decodeBurn: %w", err)
	}
	if len(vals) != 2 {
		return nil, nil, fmt.Errorf("uniswapv3: decodeBurn: expected 2 return values, got %d", len(vals))
	}
	amount0, ok := vals[0].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv3: decodeBurn: unexpected decoded type for amount0: %T", vals[0])
	}
	amount1, ok = vals[1].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv3: decodeBurn: unexpected decoded type for amount1: %T", vals[1])
	}
	return amount0, amount1, nil
}

// EncodeSwap returns the calldata for swap(recipient, zeroForOne, amountSpecified, sqrtPriceLimitX96, data).
func EncodeSwap(recipient *types.Address, zeroForOne bool, amountSpecified, sqrtPriceLimitX96 *big.Int, callbackData []byte) ([]byte, error) {
	data, err := swapFn.EncodeCall(recipient, zeroForOne, amountSpecified, sqrtPriceLimitX96, callbackData)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeSwap: %w", err)
	}
	return data, nil
}

// DecodeSwap decodes the return data of swap -- the pool's balance deltas for token0/token1.
func DecodeSwap(data []byte) (amount0, amount1 *big.Int, err error) {
	vals, err := swapFn.DecodeReturn(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: decodeSwap: %w", err)
	}
	if len(vals) != 2 {
		return nil, nil, fmt.Errorf("uniswapv3: decodeSwap: expected 2 return values, got %d", len(vals))
	}
	amount0, ok := vals[0].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv3: decodeSwap: unexpected decoded type for amount0: %T", vals[0])
	}
	amount1, ok = vals[1].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv3: decodeSwap: unexpected decoded type for amount1: %T", vals[1])
	}
	return amount0, amount1, nil
}

// EncodeFlash returns the calldata for flash(recipient, amount0, amount1, data).
func EncodeFlash(recipient *types.Address, amount0, amount1 *big.Int, callbackData []byte) ([]byte, error) {
	data, err := flashFn.EncodeCall(recipient, amount0, amount1, callbackData)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeFlash: %w", err)
	}
	return data, nil
}

// EncodeIncreaseObservationCardinalityNext returns the calldata for increaseObservationCardinalityNext(observationCardinalityNext).
// observationCardinalityNext is uint16 in the ABI, not uint256 -- abi.encodeUint requires the
// matching native Go type for widths of 8/16/32/64 bits, not *big.Int.
func EncodeIncreaseObservationCardinalityNext(observationCardinalityNext uint16) ([]byte, error) {
	data, err := increaseObservationCardinalityNextFn.EncodeCall(observationCardinalityNext)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeIncreaseObservationCardinalityNext: %w", err)
	}
	return data, nil
}

// EncodeSetFeeProtocol returns the calldata for setFeeProtocol(feeProtocol0, feeProtocol1).
// feeProtocol0/feeProtocol1 are uint8 in the ABI, not uint256 -- abi.encodeUint requires the
// matching native Go type for widths of 8/16/32/64 bits, not *big.Int.
func EncodeSetFeeProtocol(feeProtocol0, feeProtocol1 uint8) ([]byte, error) {
	data, err := setFeeProtocolFn.EncodeCall(feeProtocol0, feeProtocol1)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeSetFeeProtocol: %w", err)
	}
	return data, nil
}

// EncodeCollectProtocol returns the calldata for collectProtocol(recipient, amount0Requested, amount1Requested).
func EncodeCollectProtocol(recipient *types.Address, amount0Requested, amount1Requested *big.Int) ([]byte, error) {
	data, err := collectProtocolFn.EncodeCall(recipient, amount0Requested, amount1Requested)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeCollectProtocol: %w", err)
	}
	return data, nil
}

// DecodeCollectProtocol decodes the return data of collectProtocol -- the protocol fees collected.
func DecodeCollectProtocol(data []byte) (amount0, amount1 *big.Int, err error) {
	vals, err := collectProtocolFn.DecodeReturn(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: decodeCollectProtocol: %w", err)
	}
	if len(vals) != 2 {
		return nil, nil, fmt.Errorf("uniswapv3: decodeCollectProtocol: expected 2 return values, got %d", len(vals))
	}
	amount0, ok := vals[0].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv3: decodeCollectProtocol: unexpected decoded type for amount0: %T", vals[0])
	}
	amount1, ok = vals[1].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv3: decodeCollectProtocol: unexpected decoded type for amount1: %T", vals[1])
	}
	return amount0, amount1, nil
}
