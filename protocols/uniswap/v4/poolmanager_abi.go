package v4

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// poolKeyTuple mirrors PoolKey: (currency0, currency1, fee, tickSpacing, hooks).
var poolKeyTuple = abi.Tuple(abi.Address, abi.Address, abi.Uint24, abi.Int24, abi.Address)

// liquidityParamsTuple mirrors ModifyLiquidityParams: (tickLower, tickUpper, liquidityDelta, salt).
var liquidityParamsTuple = abi.Tuple(abi.Int24, abi.Int24, abi.Int256, abi.Bytes32)

// swapAmountParamsTuple mirrors SwapParams: (zeroForOne, amountSpecified, sqrtPriceLimitX96).
var swapAmountParamsTuple = abi.Tuple(abi.Bool, abi.Int256, abi.Uint160)

// Function signatures from IPoolManager; BalanceDelta returns are Bytes32 since each packs two int128s.
var (
	initializeFn         = abi.NewFunction("initialize", abi.NewTypes(poolKeyTuple, abi.Uint160), abi.NewTypes(abi.Int24))
	modifyLiquidityFn    = abi.NewFunction("modifyLiquidity", abi.NewTypes(poolKeyTuple, liquidityParamsTuple, abi.Bytes), abi.NewTypes(abi.Bytes32, abi.Bytes32))
	swapFn               = abi.NewFunction("swap", abi.NewTypes(poolKeyTuple, swapAmountParamsTuple, abi.Bytes), abi.NewTypes(abi.Bytes32))
	donateFn             = abi.NewFunction("donate", abi.NewTypes(poolKeyTuple, abi.Uint256, abi.Uint256, abi.Bytes), abi.NewTypes(abi.Bytes32))
	unlockFn             = abi.NewFunction("unlock", abi.NewTypes(abi.Bytes), abi.NewTypes(abi.Bytes))
	syncFn               = abi.NewFunction("sync", abi.NewTypes(abi.Address), nil)
	takeFn               = abi.NewFunction("take", abi.NewTypes(abi.Address, abi.Address, abi.Uint256), nil)
	settleFn             = abi.NewFunction("settle", nil, abi.NewTypes(abi.Uint256))
	settleForFn          = abi.NewFunction("settleFor", abi.NewTypes(abi.Address), abi.NewTypes(abi.Uint256))
	clearFn              = abi.NewFunction("clear", abi.NewTypes(abi.Address, abi.Uint256), nil)
	mintFn               = abi.NewFunction("mint", abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256), nil)
	burnFn               = abi.NewFunction("burn", abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256), nil)
	updateDynamicLPFeeFn = abi.NewFunction("updateDynamicLPFee", abi.NewTypes(poolKeyTuple, abi.Uint24), nil)
)

// Custom errors from IPoolManager, for decoding reverts via abi.Error.Decode.
var (
	errCurrencyNotSettled             = abi.NewError("CurrencyNotSettled", abi.NewTypes())
	errPoolNotInitialized             = abi.NewError("PoolNotInitialized", abi.NewTypes())
	errAlreadyUnlocked                = abi.NewError("AlreadyUnlocked", abi.NewTypes())
	errManagerLocked                  = abi.NewError("ManagerLocked", abi.NewTypes())
	errTickSpacingTooLarge            = abi.NewError("TickSpacingTooLarge", abi.NewTypes(abi.Int24))
	errTickSpacingTooSmall            = abi.NewError("TickSpacingTooSmall", abi.NewTypes(abi.Int24))
	errCurrenciesOutOfOrderOrEqual    = abi.NewError("CurrenciesOutOfOrderOrEqual", abi.NewTypes(abi.Address, abi.Address))
	errUnauthorizedDynamicLPFeeUpdate = abi.NewError("UnauthorizedDynamicLPFeeUpdate", abi.NewTypes())
	errSwapAmountCannotBeZero         = abi.NewError("SwapAmountCannotBeZero", abi.NewTypes())
	errNonzeroNativeValue             = abi.NewError("NonzeroNativeValue", abi.NewTypes())
	errMustClearExactPositiveDelta    = abi.NewError("MustClearExactPositiveDelta", abi.NewTypes())
)

// ErrCurrencyNotSettled, etc. mirror IPoolManager's own custom errors, for errors.Is.
var (
	ErrCurrencyNotSettled             = errors.New("currency not settled")
	ErrPoolNotInitialized             = errors.New("pool not initialized")
	ErrAlreadyUnlocked                = errors.New("already unlocked")
	ErrManagerLocked                  = errors.New("manager locked")
	ErrTickSpacingTooLarge            = errors.New("tick spacing too large")
	ErrTickSpacingTooSmall            = errors.New("tick spacing too small")
	ErrCurrenciesOutOfOrderOrEqual    = errors.New("currencies out of order or equal")
	ErrUnauthorizedDynamicLPFeeUpdate = errors.New("unauthorized dynamic lp fee update")
	ErrSwapAmountCannotBeZero         = errors.New("swap amount cannot be zero")
	ErrNonzeroNativeValue             = errors.New("nonzero native value")
	ErrMustClearExactPositiveDelta    = errors.New("must clear exact positive delta")
	ErrUnknown                        = errors.New("unknown revert")
)

// v4-core library errors PoolManager reverts with; the swap-path ones shared with V4Quoter live in quoter_abi.go.
var (
	errPoolAlreadyInitialized    = abi.NewError("PoolAlreadyInitialized", abi.NewTypes())
	errTicksMisordered           = abi.NewError("TicksMisordered", abi.NewTypes(abi.Int24, abi.Int24))
	errTickLowerOutOfBounds      = abi.NewError("TickLowerOutOfBounds", abi.NewTypes(abi.Int24))
	errTickUpperOutOfBounds      = abi.NewError("TickUpperOutOfBounds", abi.NewTypes(abi.Int24))
	errTickLiquidityOverflow     = abi.NewError("TickLiquidityOverflow", abi.NewTypes(abi.Int24))
	errNoLiquidityToReceiveFees  = abi.NewError("NoLiquidityToReceiveFees", abi.NewTypes())
	errTickMisaligned            = abi.NewError("TickMisaligned", abi.NewTypes(abi.Int24, abi.Int24))
	errInvalidTick               = abi.NewError("InvalidTick", abi.NewTypes(abi.Int24))
	errInvalidSqrtPrice          = abi.NewError("InvalidSqrtPrice", abi.NewTypes(abi.Uint160))
	errHookAddressNotValid       = abi.NewError("HookAddressNotValid", abi.NewTypes(abi.Address))
	errCannotUpdateEmptyPosition = abi.NewError("CannotUpdateEmptyPosition", abi.NewTypes())
	errDelegateCallNotAllowed    = abi.NewError("DelegateCallNotAllowed", abi.NewTypes())
	errNativeTransferFailed      = abi.NewError("NativeTransferFailed", abi.NewTypes())
	errERC20TransferFailed       = abi.NewError("ERC20TransferFailed", abi.NewTypes())
)

// ErrPoolAlreadyInitialized, etc. mirror the v4-core library errors above, for errors.Is.
var (
	ErrPoolAlreadyInitialized    = errors.New("pool already initialized")
	ErrTicksMisordered           = errors.New("ticks misordered")
	ErrTickLowerOutOfBounds      = errors.New("tick lower out of bounds")
	ErrTickUpperOutOfBounds      = errors.New("tick upper out of bounds")
	ErrTickLiquidityOverflow     = errors.New("tick liquidity overflow")
	ErrNoLiquidityToReceiveFees  = errors.New("no liquidity to receive fees")
	ErrTickMisaligned            = errors.New("tick misaligned")
	ErrInvalidTick               = errors.New("invalid tick")
	ErrInvalidSqrtPrice          = errors.New("invalid sqrt price")
	ErrHookAddressNotValid       = errors.New("hook address not valid")
	ErrCannotUpdateEmptyPosition = errors.New("cannot update empty position")
	ErrDelegateCallNotAllowed    = errors.New("delegate call not allowed")
	ErrNativeTransferFailed      = errors.New("native transfer failed")
	ErrERC20TransferFailed       = errors.New("erc20 transfer failed")
)

// Function signatures from IProtocolFees, for encoding calls and decoding results.
var (
	protocolFeesAccruedFn      = abi.NewFunction("protocolFeesAccrued", abi.NewTypes(abi.Address), abi.NewTypes(abi.Uint256))
	setProtocolFeeFn           = abi.NewFunction("setProtocolFee", abi.NewTypes(poolKeyTuple, abi.Uint24), nil)
	setProtocolFeeControllerFn = abi.NewFunction("setProtocolFeeController", abi.NewTypes(abi.Address), nil)
	collectProtocolFeesFn      = abi.NewFunction("collectProtocolFees", abi.NewTypes(abi.Address, abi.Address, abi.Uint256), abi.NewTypes(abi.Uint256))
	protocolFeeControllerFn    = abi.NewFunction("protocolFeeController", nil, abi.NewTypes(abi.Address))
)

// Custom errors from IProtocolFees, for decoding reverts via abi.Error.Decode.
var (
	errProtocolFeeTooLarge       = abi.NewError("ProtocolFeeTooLarge", abi.NewTypes(abi.Uint24))
	errInvalidCaller             = abi.NewError("InvalidCaller", abi.NewTypes())
	errProtocolFeeCurrencySynced = abi.NewError("ProtocolFeeCurrencySynced", abi.NewTypes())
)

// ErrProtocolFeeTooLarge, etc. mirror IProtocolFees' custom errors, for errors.Is.
var (
	ErrProtocolFeeTooLarge       = errors.New("protocol fee too large")
	ErrInvalidCaller             = errors.New("invalid caller")
	ErrProtocolFeeCurrencySynced = errors.New("protocol fee currency synced")
)

// Function signatures from IExtsload/IExttload, one var per Solidity overload.
var (
	extsloadFn      = abi.NewFunction("extsload", abi.NewTypes(abi.Bytes32), abi.NewTypes(abi.Bytes32))
	extsloadRangeFn = abi.NewFunction("extsload", abi.NewTypes(abi.Bytes32, abi.Uint256), abi.NewTypes(abi.Slice(abi.Bytes32)))
	extsloadSlotsFn = abi.NewFunction("extsload", abi.NewTypes(abi.Slice(abi.Bytes32)), abi.NewTypes(abi.Slice(abi.Bytes32)))
	exttloadFn      = abi.NewFunction("exttload", abi.NewTypes(abi.Bytes32), abi.NewTypes(abi.Bytes32))
	exttloadSlotsFn = abi.NewFunction("exttload", abi.NewTypes(abi.Slice(abi.Bytes32)), abi.NewTypes(abi.Slice(abi.Bytes32)))
)

// Events from IPoolManager, for decoding logs via Event.Decode.
var (
	initializeEvent = abi.NewEvent("Initialize", []abi.EventParam{
		{Type: abi.Bytes32, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Uint24},
		{Type: abi.Int24},
		{Type: abi.Address},
		{Type: abi.Uint160},
		{Type: abi.Int24},
	}, false)
	modifyLiquidityEvent = abi.NewEvent("ModifyLiquidity", []abi.EventParam{
		{Type: abi.Bytes32, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Int24},
		{Type: abi.Int24},
		{Type: abi.Int256},
		{Type: abi.Bytes32},
	}, false)
	swapEvent = abi.NewEvent("Swap", []abi.EventParam{
		{Type: abi.Bytes32, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Int128},
		{Type: abi.Int128},
		{Type: abi.Uint160},
		{Type: abi.Uint128},
		{Type: abi.Int24},
		{Type: abi.Uint24},
	}, false)
	donateEvent = abi.NewEvent("Donate", []abi.EventParam{
		{Type: abi.Bytes32, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Uint256},
		{Type: abi.Uint256},
	}, false)
)

// Events from IProtocolFees, for decoding logs via Event.Decode.
var (
	protocolFeeControllerUpdatedEvent = abi.NewEvent("ProtocolFeeControllerUpdated", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
	}, false)
	protocolFeeUpdatedEvent = abi.NewEvent("ProtocolFeeUpdated", []abi.EventParam{
		{Type: abi.Bytes32, Indexed: true},
		{Type: abi.Uint24},
	}, false)
)

// EncodeInitialize returns the calldata for initialize(key, sqrtPriceX96).
func EncodeInitialize(key *PoolKey, sqrtPriceX96 *big.Int) ([]byte, error) {
	if key == nil {
		return nil, fmt.Errorf("uniswapv4: encodeInitialize: key must not be nil")
	}
	data, err := initializeFn.EncodeCall([]any{key.Currency0, key.Currency1, key.Fee, key.TickSpacing, key.Hooks}, sqrtPriceX96)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeInitialize: %w", err)
	}
	return data, nil
}

// DecodeInitialize decodes the return data of initialize: the pool's initial tick.
func DecodeInitialize(data []byte) (*big.Int, error) {
	tick, err := initializeFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeInitialize: %w", err)
	}
	return tick, nil
}

// EncodeModifyLiquidity returns the calldata for modifyLiquidity(key, params, hookData).
func EncodeModifyLiquidity(key *PoolKey, params *LiquidityParams, hookData []byte) ([]byte, error) {
	if key == nil {
		return nil, fmt.Errorf("uniswapv4: encodeModifyLiquidity: key must not be nil")
	}
	if params == nil {
		return nil, fmt.Errorf("uniswapv4: encodeModifyLiquidity: params must not be nil")
	}
	data, err := modifyLiquidityFn.EncodeCall(
		[]any{key.Currency0, key.Currency1, key.Fee, key.TickSpacing, key.Hooks},
		[]any{params.TickLower, params.TickUpper, params.LiquidityDelta, params.Salt[:]},
		hookData,
	)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeModifyLiquidity: %w", err)
	}
	return data, nil
}

// DecodeModifyLiquidity decodes the return data of modifyLiquidity: (callerDelta, feesAccrued).
func DecodeModifyLiquidity(data []byte) (*BalanceDelta, *BalanceDelta, error) {
	vals, err := modifyLiquidityFn.DecodeReturn(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: decodeModifyLiquidity: %w", err)
	}
	callerWord, ok := vals[0].([]byte)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv4: decodeModifyLiquidity: unexpected decoded type for callerDelta: %T", vals[0])
	}
	if len(callerWord) != 32 {
		return nil, nil, fmt.Errorf("uniswapv4: decodeModifyLiquidity: callerDelta: expected 32 bytes, got %d", len(callerWord))
	}
	callerAmount0 := new(big.Int).SetBytes(callerWord[:16])
	if callerWord[0]&0x80 != 0 {
		callerAmount0.Sub(callerAmount0, new(big.Int).Lsh(big.NewInt(1), 128))
	}
	callerAmount1 := new(big.Int).SetBytes(callerWord[16:])
	if callerWord[16]&0x80 != 0 {
		callerAmount1.Sub(callerAmount1, new(big.Int).Lsh(big.NewInt(1), 128))
	}
	callerDelta := NewBalanceDelta(callerAmount0, callerAmount1)

	feesWord, ok := vals[1].([]byte)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv4: decodeModifyLiquidity: unexpected decoded type for feesAccrued: %T", vals[1])
	}
	if len(feesWord) != 32 {
		return nil, nil, fmt.Errorf("uniswapv4: decodeModifyLiquidity: feesAccrued: expected 32 bytes, got %d", len(feesWord))
	}
	feesAmount0 := new(big.Int).SetBytes(feesWord[:16])
	if feesWord[0]&0x80 != 0 {
		feesAmount0.Sub(feesAmount0, new(big.Int).Lsh(big.NewInt(1), 128))
	}
	feesAmount1 := new(big.Int).SetBytes(feesWord[16:])
	if feesWord[16]&0x80 != 0 {
		feesAmount1.Sub(feesAmount1, new(big.Int).Lsh(big.NewInt(1), 128))
	}
	feesAccrued := NewBalanceDelta(feesAmount0, feesAmount1)

	return callerDelta, feesAccrued, nil
}

// EncodeSwap returns the calldata for swap(key, params, hookData).
func EncodeSwap(key *PoolKey, params *SwapAmountParams, hookData []byte) ([]byte, error) {
	if key == nil {
		return nil, fmt.Errorf("uniswapv4: encodeSwap: key must not be nil")
	}
	if params == nil {
		return nil, fmt.Errorf("uniswapv4: encodeSwap: params must not be nil")
	}
	data, err := swapFn.EncodeCall(
		[]any{key.Currency0, key.Currency1, key.Fee, key.TickSpacing, key.Hooks},
		[]any{params.ZeroForOne, params.AmountSpecified, params.SqrtPriceLimitX96},
		hookData,
	)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeSwap: %w", err)
	}
	return data, nil
}

// DecodeSwap decodes the return data of swap: the balance delta of the swapping caller.
func DecodeSwap(data []byte) (*BalanceDelta, error) {
	word, err := swapFn.DecodeSingleReturn[[]byte](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeSwap: %w", err)
	}
	if len(word) != 32 {
		return nil, fmt.Errorf("uniswapv4: decodeSwap: expected 32 bytes, got %d", len(word))
	}
	amount0 := new(big.Int).SetBytes(word[:16])
	if word[0]&0x80 != 0 {
		amount0.Sub(amount0, new(big.Int).Lsh(big.NewInt(1), 128))
	}
	amount1 := new(big.Int).SetBytes(word[16:])
	if word[16]&0x80 != 0 {
		amount1.Sub(amount1, new(big.Int).Lsh(big.NewInt(1), 128))
	}
	return NewBalanceDelta(amount0, amount1), nil
}

// EncodeDonate returns the calldata for donate(key, amount0, amount1, hookData).
func EncodeDonate(key *PoolKey, amount0, amount1 *big.Int, hookData []byte) ([]byte, error) {
	if key == nil {
		return nil, fmt.Errorf("uniswapv4: encodeDonate: key must not be nil")
	}
	data, err := donateFn.EncodeCall([]any{key.Currency0, key.Currency1, key.Fee, key.TickSpacing, key.Hooks}, amount0, amount1, hookData)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeDonate: %w", err)
	}
	return data, nil
}

// DecodeDonate decodes the return data of donate: the caller's balance delta from the donation.
func DecodeDonate(data []byte) (*BalanceDelta, error) {
	word, err := donateFn.DecodeSingleReturn[[]byte](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeDonate: %w", err)
	}
	if len(word) != 32 {
		return nil, fmt.Errorf("uniswapv4: decodeDonate: expected 32 bytes, got %d", len(word))
	}
	amount0 := new(big.Int).SetBytes(word[:16])
	if word[0]&0x80 != 0 {
		amount0.Sub(amount0, new(big.Int).Lsh(big.NewInt(1), 128))
	}
	amount1 := new(big.Int).SetBytes(word[16:])
	if word[16]&0x80 != 0 {
		amount1.Sub(amount1, new(big.Int).Lsh(big.NewInt(1), 128))
	}
	return NewBalanceDelta(amount0, amount1), nil
}

// EncodeUnlock returns the calldata for unlock(data).
func EncodeUnlock(unlockData []byte) ([]byte, error) {
	data, err := unlockFn.EncodeCall(unlockData)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeUnlock: %w", err)
	}
	return data, nil
}

// DecodeUnlock decodes the return data of unlock: whatever bytes the unlockCallback returned.
func DecodeUnlock(data []byte) ([]byte, error) {
	out, err := unlockFn.DecodeSingleReturn[[]byte](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeUnlock: %w", err)
	}
	return out, nil
}

// EncodeSync returns the calldata for sync(currency).
func EncodeSync(currency Currency) ([]byte, error) {
	data, err := syncFn.EncodeCall(currency)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeSync: %w", err)
	}
	return data, nil
}

// EncodeTake returns the calldata for take(currency, to, amount).
func EncodeTake(currency Currency, to *types.Address, amount *big.Int) ([]byte, error) {
	data, err := takeFn.EncodeCall(currency, to, amount)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeTake: %w", err)
	}
	return data, nil
}

// EncodeSettle returns the calldata for settle().
func EncodeSettle() ([]byte, error) {
	data, err := settleFn.EncodeCall()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeSettle: %w", err)
	}
	return data, nil
}

// DecodeSettle decodes the return data of settle: the amount settled.
func DecodeSettle(data []byte) (*big.Int, error) {
	paid, err := settleFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeSettle: %w", err)
	}
	return paid, nil
}

// EncodeSettleFor returns the calldata for settleFor(recipient).
func EncodeSettleFor(recipient *types.Address) ([]byte, error) {
	data, err := settleForFn.EncodeCall(recipient)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeSettleFor: %w", err)
	}
	return data, nil
}

// DecodeSettleFor decodes the return data of settleFor: the amount settled.
func DecodeSettleFor(data []byte) (*big.Int, error) {
	paid, err := settleForFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeSettleFor: %w", err)
	}
	return paid, nil
}

// EncodeClear returns the calldata for clear(currency, amount).
func EncodeClear(currency Currency, amount *big.Int) ([]byte, error) {
	data, err := clearFn.EncodeCall(currency, amount)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeClear: %w", err)
	}
	return data, nil
}

// EncodeMint returns the calldata for mint(to, id, amount), the ERC6909 claim-token mint.
func EncodeMint(to *types.Address, id, amount *big.Int) ([]byte, error) {
	data, err := mintFn.EncodeCall(to, id, amount)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeMint: %w", err)
	}
	return data, nil
}

// EncodeBurn returns the calldata for burn(from, id, amount).
func EncodeBurn(from *types.Address, id, amount *big.Int) ([]byte, error) {
	data, err := burnFn.EncodeCall(from, id, amount)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeBurn: %w", err)
	}
	return data, nil
}

// EncodeUpdateDynamicLPFee returns the calldata for updateDynamicLPFee(key, newDynamicLPFee).
func EncodeUpdateDynamicLPFee(key *PoolKey, newDynamicLPFee *big.Int) ([]byte, error) {
	if key == nil {
		return nil, fmt.Errorf("uniswapv4: encodeUpdateDynamicLPFee: key must not be nil")
	}
	data, err := updateDynamicLPFeeFn.EncodeCall([]any{key.Currency0, key.Currency1, key.Fee, key.TickSpacing, key.Hooks}, newDynamicLPFee)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeUpdateDynamicLPFee: %w", err)
	}
	return data, nil
}

// EncodeProtocolFeesAccrued returns the calldata for protocolFeesAccrued(currency).
func EncodeProtocolFeesAccrued(currency Currency) ([]byte, error) {
	data, err := protocolFeesAccruedFn.EncodeCall(currency)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeProtocolFeesAccrued: %w", err)
	}
	return data, nil
}

// DecodeProtocolFeesAccrued decodes the return data of protocolFeesAccrued.
func DecodeProtocolFeesAccrued(data []byte) (*big.Int, error) {
	amount, err := protocolFeesAccruedFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeProtocolFeesAccrued: %w", err)
	}
	return amount, nil
}

// EncodeSetProtocolFee returns the calldata for setProtocolFee(key, newProtocolFee).
func EncodeSetProtocolFee(key *PoolKey, newProtocolFee *big.Int) ([]byte, error) {
	if key == nil {
		return nil, fmt.Errorf("uniswapv4: encodeSetProtocolFee: key must not be nil")
	}
	data, err := setProtocolFeeFn.EncodeCall([]any{key.Currency0, key.Currency1, key.Fee, key.TickSpacing, key.Hooks}, newProtocolFee)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeSetProtocolFee: %w", err)
	}
	return data, nil
}

// EncodeSetProtocolFeeController returns the calldata for setProtocolFeeController(controller).
func EncodeSetProtocolFeeController(controller *types.Address) ([]byte, error) {
	data, err := setProtocolFeeControllerFn.EncodeCall(controller)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeSetProtocolFeeController: %w", err)
	}
	return data, nil
}

// EncodeCollectProtocolFees returns the calldata for collectProtocolFees(recipient, currency, amount).
func EncodeCollectProtocolFees(recipient *types.Address, currency Currency, amount *big.Int) ([]byte, error) {
	data, err := collectProtocolFeesFn.EncodeCall(recipient, currency, amount)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeCollectProtocolFees: %w", err)
	}
	return data, nil
}

// DecodeCollectProtocolFees decodes the return data of collectProtocolFees: the amount actually collected.
func DecodeCollectProtocolFees(data []byte) (*big.Int, error) {
	amountCollected, err := collectProtocolFeesFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeCollectProtocolFees: %w", err)
	}
	return amountCollected, nil
}

// EncodeProtocolFeeController returns the calldata for protocolFeeController().
func EncodeProtocolFeeController() ([]byte, error) {
	data, err := protocolFeeControllerFn.EncodeCall()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeProtocolFeeController: %w", err)
	}
	return data, nil
}

// DecodeProtocolFeeController decodes the return data of protocolFeeController.
func DecodeProtocolFeeController(data []byte) (*types.Address, error) {
	controller, err := protocolFeeControllerFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeProtocolFeeController: %w", err)
	}
	return controller, nil
}

// EncodeExtsload returns the calldata for extsload(slot).
func EncodeExtsload(slot [32]byte) ([]byte, error) {
	data, err := extsloadFn.EncodeCall(slot[:])
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeExtsload: %w", err)
	}
	return data, nil
}

// DecodeExtsload decodes the return data of extsload(slot).
func DecodeExtsload(data []byte) ([32]byte, error) {
	raw, err := extsloadFn.DecodeSingleReturn[[]byte](data)
	if err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv4: decodeExtsload: %w", err)
	}
	var value [32]byte
	copy(value[:], raw)
	return value, nil
}

// EncodeExtsloadRange returns the calldata for extsload(startSlot, nSlots).
func EncodeExtsloadRange(startSlot [32]byte, nSlots *big.Int) ([]byte, error) {
	data, err := extsloadRangeFn.EncodeCall(startSlot[:], nSlots)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeExtsloadRange: %w", err)
	}
	return data, nil
}

// DecodeExtsloadRange decodes the return data of extsload(startSlot, nSlots).
func DecodeExtsloadRange(data []byte) ([][32]byte, error) {
	vals, err := extsloadRangeFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeExtsloadRange: %w", err)
	}
	elems, ok := vals[0].([]any)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeExtsloadRange: unexpected decoded type for values: %T", vals[0])
	}
	values := make([][32]byte, len(elems))
	for i, e := range elems {
		raw, ok := e.([]byte)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: decodeExtsloadRange: values[%d]: unexpected decoded type: %T", i, e)
		}
		copy(values[i][:], raw)
	}
	return values, nil
}

// EncodeExtsloadSlots returns the calldata for extsload(slots).
func EncodeExtsloadSlots(slots [][32]byte) ([]byte, error) {
	raw := make([][]byte, len(slots))
	for i, s := range slots {
		raw[i] = s[:]
	}
	data, err := extsloadSlotsFn.EncodeCall(raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeExtsloadSlots: %w", err)
	}
	return data, nil
}

// DecodeExtsloadSlots decodes the return data of extsload(slots).
func DecodeExtsloadSlots(data []byte) ([][32]byte, error) {
	vals, err := extsloadSlotsFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeExtsloadSlots: %w", err)
	}
	elems, ok := vals[0].([]any)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeExtsloadSlots: unexpected decoded type for values: %T", vals[0])
	}
	values := make([][32]byte, len(elems))
	for i, e := range elems {
		raw, ok := e.([]byte)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: decodeExtsloadSlots: values[%d]: unexpected decoded type: %T", i, e)
		}
		copy(values[i][:], raw)
	}
	return values, nil
}

// EncodeExttload returns the calldata for exttload(slot).
func EncodeExttload(slot [32]byte) ([]byte, error) {
	data, err := exttloadFn.EncodeCall(slot[:])
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeExttload: %w", err)
	}
	return data, nil
}

// DecodeExttload decodes the return data of exttload(slot).
func DecodeExttload(data []byte) ([32]byte, error) {
	raw, err := exttloadFn.DecodeSingleReturn[[]byte](data)
	if err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv4: decodeExttload: %w", err)
	}
	var value [32]byte
	copy(value[:], raw)
	return value, nil
}

// EncodeExttloadSlots returns the calldata for exttload(slots).
func EncodeExttloadSlots(slots [][32]byte) ([]byte, error) {
	raw := make([][]byte, len(slots))
	for i, s := range slots {
		raw[i] = s[:]
	}
	data, err := exttloadSlotsFn.EncodeCall(raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeExttloadSlots: %w", err)
	}
	return data, nil
}

// DecodeExttloadSlots decodes the return data of exttload(slots).
func DecodeExttloadSlots(data []byte) ([][32]byte, error) {
	vals, err := exttloadSlotsFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeExttloadSlots: %w", err)
	}
	elems, ok := vals[0].([]any)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeExttloadSlots: unexpected decoded type for values: %T", vals[0])
	}
	values := make([][32]byte, len(elems))
	for i, e := range elems {
		raw, ok := e.([]byte)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: decodeExttloadSlots: values[%d]: unexpected decoded type: %T", i, e)
		}
		copy(values[i][:], raw)
	}
	return values, nil
}

// mapPoolManagerRevertError maps a known PoolManager or v4-core custom error to its sentinel, an unknown selector to ErrUnknown, and passes data-less errors through.
func mapPoolManagerRevertError(err error) error {
	if rpcErr, ok := errors.AsType[*rpc.ResponseError](err); ok && rpcErr == nil {
		return nil
	}

	rpcErr, ok := errors.AsType[*rpc.ResponseError](err)
	if !ok || len(rpcErr.Data) == 0 {
		return err
	}

	var hexData string
	if jsonErr := json.Unmarshal(rpcErr.Data, &hexData); jsonErr != nil {
		return err
	}

	data, decErr := encoding.Hex.Decode(hexData)
	if decErr != nil || len(data) < 4 {
		return err
	}

	var sel abi.Selector
	copy(sel[:], data[:4])

	switch sel {
	case errCurrencyNotSettled.Selector():
		return fmt.Errorf("%w: %w", ErrCurrencyNotSettled, err)
	case errPoolNotInitialized.Selector():
		return fmt.Errorf("%w: %w", ErrPoolNotInitialized, err)
	case errAlreadyUnlocked.Selector():
		return fmt.Errorf("%w: %w", ErrAlreadyUnlocked, err)
	case errManagerLocked.Selector():
		return fmt.Errorf("%w: %w", ErrManagerLocked, err)
	case errTickSpacingTooLarge.Selector():
		return fmt.Errorf("%w: %w", ErrTickSpacingTooLarge, err)
	case errTickSpacingTooSmall.Selector():
		return fmt.Errorf("%w: %w", ErrTickSpacingTooSmall, err)
	case errCurrenciesOutOfOrderOrEqual.Selector():
		return fmt.Errorf("%w: %w", ErrCurrenciesOutOfOrderOrEqual, err)
	case errUnauthorizedDynamicLPFeeUpdate.Selector():
		return fmt.Errorf("%w: %w", ErrUnauthorizedDynamicLPFeeUpdate, err)
	case errSwapAmountCannotBeZero.Selector():
		return fmt.Errorf("%w: %w", ErrSwapAmountCannotBeZero, err)
	case errNonzeroNativeValue.Selector():
		return fmt.Errorf("%w: %w", ErrNonzeroNativeValue, err)
	case errMustClearExactPositiveDelta.Selector():
		return fmt.Errorf("%w: %w", ErrMustClearExactPositiveDelta, err)
	case errProtocolFeeTooLarge.Selector():
		return fmt.Errorf("%w: %w", ErrProtocolFeeTooLarge, err)
	case errInvalidCaller.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidCaller, err)
	case errProtocolFeeCurrencySynced.Selector():
		return fmt.Errorf("%w: %w", ErrProtocolFeeCurrencySynced, err)
	case errPoolAlreadyInitialized.Selector():
		return fmt.Errorf("%w: %w", ErrPoolAlreadyInitialized, err)
	case errTicksMisordered.Selector():
		return fmt.Errorf("%w: %w", ErrTicksMisordered, err)
	case errTickLowerOutOfBounds.Selector():
		return fmt.Errorf("%w: %w", ErrTickLowerOutOfBounds, err)
	case errTickUpperOutOfBounds.Selector():
		return fmt.Errorf("%w: %w", ErrTickUpperOutOfBounds, err)
	case errTickLiquidityOverflow.Selector():
		return fmt.Errorf("%w: %w", ErrTickLiquidityOverflow, err)
	case errNoLiquidityToReceiveFees.Selector():
		return fmt.Errorf("%w: %w", ErrNoLiquidityToReceiveFees, err)
	case errTickMisaligned.Selector():
		return fmt.Errorf("%w: %w", ErrTickMisaligned, err)
	case errInvalidTick.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidTick, err)
	case errInvalidSqrtPrice.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidSqrtPrice, err)
	case errHookAddressNotValid.Selector():
		return fmt.Errorf("%w: %w", ErrHookAddressNotValid, err)
	case errCannotUpdateEmptyPosition.Selector():
		return fmt.Errorf("%w: %w", ErrCannotUpdateEmptyPosition, err)
	case errDelegateCallNotAllowed.Selector():
		return fmt.Errorf("%w: %w", ErrDelegateCallNotAllowed, err)
	case errNativeTransferFailed.Selector():
		return fmt.Errorf("%w: %w", ErrNativeTransferFailed, err)
	case errERC20TransferFailed.Selector():
		return fmt.Errorf("%w: %w", ErrERC20TransferFailed, err)
	case errSafeCastOverflow.Selector():
		return fmt.Errorf("%w: %w", ErrSafeCastOverflow, err)
	case errPriceLimitAlreadyExceeded.Selector():
		return fmt.Errorf("%w: %w", ErrPriceLimitAlreadyExceeded, err)
	case errPriceLimitOutOfBounds.Selector():
		return fmt.Errorf("%w: %w", ErrPriceLimitOutOfBounds, err)
	case errInvalidFeeForExactOut.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidFeeForExactOut, err)
	case errInvalidHookResponse.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidHookResponse, err)
	case errHookDeltaExceedsSwapAmount.Selector():
		return fmt.Errorf("%w: %w", ErrHookDeltaExceedsSwapAmount, err)
	case errLPFeeTooLarge.Selector():
		return fmt.Errorf("%w: %w", ErrLPFeeTooLarge, err)
	case errInvalidPriceOrLiquidity.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidPriceOrLiquidity, err)
	case errInvalidPrice.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidPrice, err)
	case errPriceOverflow.Selector():
		return fmt.Errorf("%w: %w", ErrPriceOverflow, err)
	case errSqrtPriceMathNotEnoughLiquidity.Selector():
		return fmt.Errorf("%w: %w", ErrNotEnoughLiquidity, err)
	case errWrappedError.Selector():
		vals, wrappedErr := errWrappedError.Decode(data)
		if wrappedErr != nil || len(vals) != 4 {
			return fmt.Errorf("%w: %w", ErrWrappedError, err)
		}
		details, ok := vals[3].([]byte)
		if !ok || len(details) < 4 {
			return fmt.Errorf("%w: %w", ErrWrappedError, err)
		}
		var detailsSel abi.Selector
		copy(detailsSel[:], details[:4])
		if detailsSel == errHookCallFailed.Selector() {
			return fmt.Errorf("%w: %w: %w", ErrWrappedError, ErrHookCallFailed, err)
		}
		return fmt.Errorf("%w: %w", ErrWrappedError, err)
	default:
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}
}
