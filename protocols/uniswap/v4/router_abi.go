package v4

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// pathKeyTuple mirrors libraries/PathKey.sol's PathKey struct.
var pathKeyTuple = abi.Tuple(abi.Address, abi.Uint24, abi.Int24, abi.Address, abi.Bytes)

// exactInputSingleParamsTuple mirrors IV4Router's ExactInputSingleParams.
var exactInputSingleParamsTuple = abi.Tuple(poolKeyTuple, abi.Bool, abi.Uint128, abi.Uint128, abi.Uint256, abi.Bytes)

// exactInputParamsTuple mirrors IV4Router's ExactInputParams.
var exactInputParamsTuple = abi.Tuple(abi.Address, abi.Slice(pathKeyTuple), abi.Slice(abi.Uint256), abi.Uint128, abi.Uint128)

// exactOutputSingleParamsTuple mirrors IV4Router's ExactOutputSingleParams.
var exactOutputSingleParamsTuple = abi.Tuple(poolKeyTuple, abi.Bool, abi.Uint128, abi.Uint128, abi.Uint256, abi.Bytes)

// exactOutputParamsTuple mirrors IV4Router's ExactOutputParams.
var exactOutputParamsTuple = abi.Tuple(abi.Address, abi.Slice(pathKeyTuple), abi.Slice(abi.Uint256), abi.Uint128, abi.Uint128)

// Action IDs from libraries/Actions.sol, the full set (0x00-0x19)
const (
	ActionIncreaseLiquidity           byte = iota // 0x00
	ActionDecreaseLiquidity                       // 0x01
	ActionMintPosition                            // 0x02
	ActionBurnPosition                            // 0x03
	ActionIncreaseLiquidityFromDeltas             // 0x04
	ActionMintPositionFromDeltas                  // 0x05
	ActionSwapExactInSingle                       // 0x06
	ActionSwapExactIn                             // 0x07
	ActionSwapExactOutSingle                      // 0x08
	ActionSwapExactOut                            // 0x09
	ActionDonate                                  // 0x0a
	ActionSettle                                  // 0x0b
	ActionSettleAll                               // 0x0c
	ActionSettlePair                              // 0x0d
	ActionTake                                    // 0x0e
	ActionTakeAll                                 // 0x0f
	ActionTakePortion                             // 0x10
	ActionTakePair                                // 0x11
	ActionCloseCurrency                           // 0x12
	ActionClearOrTake                             // 0x13
	ActionSweep                                   // 0x14
	ActionWrap                                    // 0x15
	ActionUnwrap                                  // 0x16
	ActionMint6909                                // 0x17
	ActionBurn6909                                // 0x18
	ActionUnwindWithFallback
)

// MsgSender returns ActionConstants.sol's MSG_SENDER sentinel. Fresh value each call since *types.Address is mutable in place.
func MsgSender() *types.Address {
	return types.MustNewAddressFromHex("0x0000000000000000000000000000000000000001")
}

// AddressThis returns ActionConstants.sol's ADDRESS_THIS sentinel.
func AddressThis() *types.Address {
	return types.MustNewAddressFromHex("0x0000000000000000000000000000000000000002")
}

// OpenDelta returns ActionConstants.sol's OPEN_DELTA sentinel. Fresh value each call since *big.Int is mutable in place.
func OpenDelta() *big.Int {
	return big.NewInt(0)
}

// ContractBalance returns ActionConstants.sol's CONTRACT_BALANCE sentinel.
func ContractBalance() *big.Int {
	return new(big.Int).Lsh(big.NewInt(1), 255)
}

// executeActionsFn mirrors a V4Router-based deployment's executeActions(bytes), forwarding to BaseActionsRouter._executeActions.
var executeActionsFn = abi.NewFunction("executeActions", abi.NewTypes(abi.Bytes), nil)

// Custom errors from IV4Router, for decoding reverts via abi.Error.Decode.
var (
	errV4TooLittleReceived             = abi.NewError("V4TooLittleReceived", abi.NewTypes(abi.Uint256, abi.Uint256))
	errV4TooMuchRequested              = abi.NewError("V4TooMuchRequested", abi.NewTypes(abi.Uint256, abi.Uint256))
	errV4TooLittleReceivedPerHop       = abi.NewError("V4TooLittleReceivedPerHop", abi.NewTypes(abi.Uint256, abi.Uint256, abi.Uint256))
	errV4TooMuchRequestedPerHop        = abi.NewError("V4TooMuchRequestedPerHop", abi.NewTypes(abi.Uint256, abi.Uint256, abi.Uint256))
	errV4TooLittleReceivedPerHopSingle = abi.NewError("V4TooLittleReceivedPerHopSingle", abi.NewTypes(abi.Uint256, abi.Uint256))
	errV4TooMuchRequestedPerHopSingle  = abi.NewError("V4TooMuchRequestedPerHopSingle", abi.NewTypes(abi.Uint256, abi.Uint256))
	errInvalidHopPriceLength           = abi.NewError("InvalidHopPriceLength", abi.NewTypes())
	errV4ExactOutputUnfilled           = abi.NewError("V4ExactOutputUnfilled", abi.NewTypes(abi.Uint256, abi.Uint256))
)

// Custom errors from BaseActionsRouter, for decoding reverts via abi.Error.Decode.
var (
	errInputLengthMismatch = abi.NewError("InputLengthMismatch", abi.NewTypes())
	errUnsupportedAction   = abi.NewError("UnsupportedAction", abi.NewTypes(abi.Uint256))
)

// ErrV4TooLittleReceived, etc. mirror IV4Router's/BaseActionsRouter's own custom errors, for errors.Is.
var (
	ErrV4TooLittleReceived             = errors.New("too little received")
	ErrV4TooMuchRequested              = errors.New("too much requested")
	ErrV4TooLittleReceivedPerHop       = errors.New("too little received per hop")
	ErrV4TooMuchRequestedPerHop        = errors.New("too much requested per hop")
	ErrV4TooLittleReceivedPerHopSingle = errors.New("too little received per hop (single)")
	ErrV4TooMuchRequestedPerHopSingle  = errors.New("too much requested per hop (single)")
	ErrInvalidHopPriceLength           = errors.New("invalid hop price length")
	ErrV4ExactOutputUnfilled           = errors.New("exact output unfilled")
	ErrInputLengthMismatch             = errors.New("input length mismatch")
	ErrUnsupportedAction               = errors.New("unsupported action")
)

// EncodeExecuteActions returns the calldata for our Kit router's executeActions(abi.encode(actions, params)).
func EncodeExecuteActions(actions []byte, params [][]byte) ([]byte, error) {
	actionsData, err := EncodeActions(actions, params)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeExecuteActions: %w", err)
	}
	data, err := executeActionsFn.EncodeCall(actionsData)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeExecuteActions: %w", err)
	}
	return data, nil
}

// EncodeActions returns BaseActionsRouter's unlock data abi.encode(actions, params), where params[i] belongs to actions[i].
func EncodeActions(actions []byte, params [][]byte) ([]byte, error) {
	paramsArg := make([]any, len(params))
	for i, p := range params {
		paramsArg[i] = p
	}
	data, err := abi.Pack(abi.NewTypes(abi.Bytes, abi.Slice(abi.Bytes)), actions, paramsArg)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeActions: %w", err)
	}
	return data, nil
}

// DecodeActions decodes BaseActionsRouter's unlockCallback data back into its actions and params.
func DecodeActions(data []byte) ([]byte, [][]byte, error) {
	vals, err := abi.Unpack(abi.NewTypes(abi.Bytes, abi.Slice(abi.Bytes)), data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: decodeActions: %w", err)
	}
	actions, ok := vals[0].([]byte)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv4: decodeActions: unexpected decoded type for actions: %T", vals[0])
	}
	paramsVals, ok := vals[1].([]any)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv4: decodeActions: unexpected decoded type for params: %T", vals[1])
	}
	params := make([][]byte, len(paramsVals))
	for i, v := range paramsVals {
		p, ok := v.([]byte)
		if !ok {
			return nil, nil, fmt.Errorf("uniswapv4: decodeActions: unexpected decoded type for params[%d]: %T", i, v)
		}
		params[i] = p
	}
	return actions, params, nil
}

// EncodeActionSwapExactInSingle returns the params blob for Actions.SWAP_EXACT_IN_SINGLE.
func EncodeActionSwapExactInSingle(p *ExactInputSingleParams) ([]byte, error) {
	if p == nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionSwapExactInSingle: params must not be nil")
	}
	if p.PoolKey == nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionSwapExactInSingle: poolKey must not be nil")
	}
	data, err := abi.Pack(abi.NewTypes(exactInputSingleParamsTuple), []any{
		[]any{p.PoolKey.Currency0, p.PoolKey.Currency1, p.PoolKey.Fee, p.PoolKey.TickSpacing, p.PoolKey.Hooks},
		p.ZeroForOne, p.AmountIn, p.AmountOutMinimum, p.MinHopPriceX36, p.HookData,
	})
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionSwapExactInSingle: %w", err)
	}
	return data, nil
}

// DecodeActionSwapExactInSingle decodes a Actions.SWAP_EXACT_IN_SINGLE params blob.
func DecodeActionSwapExactInSingle(data []byte) (*ExactInputSingleParams, error) {
	vals, err := abi.Unpack(abi.NewTypes(exactInputSingleParamsTuple), data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactInSingle: %w", err)
	}
	fields, ok := vals[0].([]any)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactInSingle: unexpected decoded type: %T", vals[0])
	}
	poolKeyFields, ok := fields[0].([]any)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactInSingle: unexpected decoded type for poolKey: %T", fields[0])
	}
	currency0, ok := poolKeyFields[0].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactInSingle: unexpected decoded type for currency0: %T", poolKeyFields[0])
	}
	currency1, ok := poolKeyFields[1].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactInSingle: unexpected decoded type for currency1: %T", poolKeyFields[1])
	}
	fee, ok := poolKeyFields[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactInSingle: unexpected decoded type for fee: %T", poolKeyFields[2])
	}
	tickSpacing, ok := poolKeyFields[3].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactInSingle: unexpected decoded type for tickSpacing: %T", poolKeyFields[3])
	}
	hooks, ok := poolKeyFields[4].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactInSingle: unexpected decoded type for hooks: %T", poolKeyFields[4])
	}
	zeroForOne, ok := fields[1].(bool)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactInSingle: unexpected decoded type for zeroForOne: %T", fields[1])
	}
	amountIn, ok := fields[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactInSingle: unexpected decoded type for amountIn: %T", fields[2])
	}
	amountOutMinimum, ok := fields[3].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactInSingle: unexpected decoded type for amountOutMinimum: %T", fields[3])
	}
	minHopPriceX36, ok := fields[4].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactInSingle: unexpected decoded type for minHopPriceX36: %T", fields[4])
	}
	hookData, ok := fields[5].([]byte)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactInSingle: unexpected decoded type for hookData: %T", fields[5])
	}
	return NewExactInputSingleParams(
		NewPoolKey(currency0, currency1, fee, tickSpacing, hooks),
		zeroForOne, amountIn, amountOutMinimum, minHopPriceX36, hookData,
	), nil
}

// EncodeActionSwapExactIn returns the params blob for Actions.SWAP_EXACT_IN.
func EncodeActionSwapExactIn(p *ExactInputParams) ([]byte, error) {
	if p == nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionSwapExactIn: params must not be nil")
	}
	path := make([]any, len(p.Path))
	for i, k := range p.Path {
		if k == nil {
			return nil, fmt.Errorf("uniswapv4: encodeActionSwapExactIn: path[%d] must not be nil", i)
		}
		path[i] = []any{k.IntermediateCurrency, k.Fee, k.TickSpacing, k.Hooks, k.HookData}
	}
	minHopPriceX36 := make([]any, len(p.MinHopPriceX36))
	for i, v := range p.MinHopPriceX36 {
		minHopPriceX36[i] = v
	}
	data, err := abi.Pack(abi.NewTypes(exactInputParamsTuple), []any{
		p.CurrencyIn, path, minHopPriceX36, p.AmountIn, p.AmountOutMinimum,
	})
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionSwapExactIn: %w", err)
	}
	return data, nil
}

// DecodeActionSwapExactIn decodes a Actions.SWAP_EXACT_IN params blob.
func DecodeActionSwapExactIn(data []byte) (*ExactInputParams, error) {
	vals, err := abi.Unpack(abi.NewTypes(exactInputParamsTuple), data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactIn: %w", err)
	}
	fields, ok := vals[0].([]any)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactIn: unexpected decoded type: %T", vals[0])
	}
	currencyIn, ok := fields[0].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactIn: unexpected decoded type for currencyIn: %T", fields[0])
	}
	pathVals, ok := fields[1].([]any)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactIn: unexpected decoded type for path: %T", fields[1])
	}
	path := make([]*PathKey, len(pathVals))
	for i, v := range pathVals {
		pathKeyFields, ok := v.([]any)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactIn: unexpected decoded type for path[%d]: %T", i, v)
		}
		intermediateCurrency, ok := pathKeyFields[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactIn: unexpected decoded type for path[%d].intermediateCurrency: %T", i, pathKeyFields[0])
		}
		fee, ok := pathKeyFields[1].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactIn: unexpected decoded type for path[%d].fee: %T", i, pathKeyFields[1])
		}
		tickSpacing, ok := pathKeyFields[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactIn: unexpected decoded type for path[%d].tickSpacing: %T", i, pathKeyFields[2])
		}
		hooks, ok := pathKeyFields[3].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactIn: unexpected decoded type for path[%d].hooks: %T", i, pathKeyFields[3])
		}
		hookData, ok := pathKeyFields[4].([]byte)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactIn: unexpected decoded type for path[%d].hookData: %T", i, pathKeyFields[4])
		}
		path[i] = NewPathKey(intermediateCurrency, fee, tickSpacing, hooks, hookData)
	}
	minHopPriceX36Vals, ok := fields[2].([]any)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactIn: unexpected decoded type for minHopPriceX36: %T", fields[2])
	}
	minHopPriceX36 := make([]*big.Int, len(minHopPriceX36Vals))
	for i, v := range minHopPriceX36Vals {
		n, ok := v.(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactIn: unexpected decoded type for minHopPriceX36[%d]: %T", i, v)
		}
		minHopPriceX36[i] = n
	}
	amountIn, ok := fields[3].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactIn: unexpected decoded type for amountIn: %T", fields[3])
	}
	amountOutMinimum, ok := fields[4].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactIn: unexpected decoded type for amountOutMinimum: %T", fields[4])
	}
	return NewExactInputParams(currencyIn, path, minHopPriceX36, amountIn, amountOutMinimum), nil
}

// EncodeActionSwapExactOutSingle returns the params blob for Actions.SWAP_EXACT_OUT_SINGLE.
func EncodeActionSwapExactOutSingle(p *ExactOutputSingleParams) ([]byte, error) {
	if p == nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionSwapExactOutSingle: params must not be nil")
	}
	if p.PoolKey == nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionSwapExactOutSingle: poolKey must not be nil")
	}
	data, err := abi.Pack(abi.NewTypes(exactOutputSingleParamsTuple), []any{
		[]any{p.PoolKey.Currency0, p.PoolKey.Currency1, p.PoolKey.Fee, p.PoolKey.TickSpacing, p.PoolKey.Hooks},
		p.ZeroForOne, p.AmountOut, p.AmountInMaximum, p.MinHopPriceX36, p.HookData,
	})
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionSwapExactOutSingle: %w", err)
	}
	return data, nil
}

// DecodeActionSwapExactOutSingle decodes a Actions.SWAP_EXACT_OUT_SINGLE params blob.
func DecodeActionSwapExactOutSingle(data []byte) (*ExactOutputSingleParams, error) {
	vals, err := abi.Unpack(abi.NewTypes(exactOutputSingleParamsTuple), data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOutSingle: %w", err)
	}
	fields, ok := vals[0].([]any)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOutSingle: unexpected decoded type: %T", vals[0])
	}
	poolKeyFields, ok := fields[0].([]any)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOutSingle: unexpected decoded type for poolKey: %T", fields[0])
	}
	currency0, ok := poolKeyFields[0].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOutSingle: unexpected decoded type for currency0: %T", poolKeyFields[0])
	}
	currency1, ok := poolKeyFields[1].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOutSingle: unexpected decoded type for currency1: %T", poolKeyFields[1])
	}
	fee, ok := poolKeyFields[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOutSingle: unexpected decoded type for fee: %T", poolKeyFields[2])
	}
	tickSpacing, ok := poolKeyFields[3].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOutSingle: unexpected decoded type for tickSpacing: %T", poolKeyFields[3])
	}
	hooks, ok := poolKeyFields[4].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOutSingle: unexpected decoded type for hooks: %T", poolKeyFields[4])
	}
	zeroForOne, ok := fields[1].(bool)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOutSingle: unexpected decoded type for zeroForOne: %T", fields[1])
	}
	amountOut, ok := fields[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOutSingle: unexpected decoded type for amountOut: %T", fields[2])
	}
	amountInMaximum, ok := fields[3].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOutSingle: unexpected decoded type for amountInMaximum: %T", fields[3])
	}
	minHopPriceX36, ok := fields[4].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOutSingle: unexpected decoded type for minHopPriceX36: %T", fields[4])
	}
	hookData, ok := fields[5].([]byte)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOutSingle: unexpected decoded type for hookData: %T", fields[5])
	}
	return NewExactOutputSingleParams(
		NewPoolKey(currency0, currency1, fee, tickSpacing, hooks),
		zeroForOne, amountOut, amountInMaximum, minHopPriceX36, hookData,
	), nil
}

// EncodeActionSwapExactOut returns the params blob for Actions.SWAP_EXACT_OUT.
func EncodeActionSwapExactOut(p *ExactOutputParams) ([]byte, error) {
	if p == nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionSwapExactOut: params must not be nil")
	}
	path := make([]any, len(p.Path))
	for i, k := range p.Path {
		if k == nil {
			return nil, fmt.Errorf("uniswapv4: encodeActionSwapExactOut: path[%d] must not be nil", i)
		}
		path[i] = []any{k.IntermediateCurrency, k.Fee, k.TickSpacing, k.Hooks, k.HookData}
	}
	minHopPriceX36 := make([]any, len(p.MinHopPriceX36))
	for i, v := range p.MinHopPriceX36 {
		minHopPriceX36[i] = v
	}
	data, err := abi.Pack(abi.NewTypes(exactOutputParamsTuple), []any{
		p.CurrencyOut, path, minHopPriceX36, p.AmountOut, p.AmountInMaximum,
	})
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionSwapExactOut: %w", err)
	}
	return data, nil
}

// DecodeActionSwapExactOut decodes a Actions.SWAP_EXACT_OUT params blob.
func DecodeActionSwapExactOut(data []byte) (*ExactOutputParams, error) {
	vals, err := abi.Unpack(abi.NewTypes(exactOutputParamsTuple), data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOut: %w", err)
	}
	fields, ok := vals[0].([]any)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOut: unexpected decoded type: %T", vals[0])
	}
	currencyOut, ok := fields[0].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOut: unexpected decoded type for currencyOut: %T", fields[0])
	}
	pathVals, ok := fields[1].([]any)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOut: unexpected decoded type for path: %T", fields[1])
	}
	path := make([]*PathKey, len(pathVals))
	for i, v := range pathVals {
		pathKeyFields, ok := v.([]any)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOut: unexpected decoded type for path[%d]: %T", i, v)
		}
		intermediateCurrency, ok := pathKeyFields[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOut: unexpected decoded type for path[%d].intermediateCurrency: %T", i, pathKeyFields[0])
		}
		fee, ok := pathKeyFields[1].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOut: unexpected decoded type for path[%d].fee: %T", i, pathKeyFields[1])
		}
		tickSpacing, ok := pathKeyFields[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOut: unexpected decoded type for path[%d].tickSpacing: %T", i, pathKeyFields[2])
		}
		hooks, ok := pathKeyFields[3].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOut: unexpected decoded type for path[%d].hooks: %T", i, pathKeyFields[3])
		}
		hookData, ok := pathKeyFields[4].([]byte)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOut: unexpected decoded type for path[%d].hookData: %T", i, pathKeyFields[4])
		}
		path[i] = NewPathKey(intermediateCurrency, fee, tickSpacing, hooks, hookData)
	}
	minHopPriceX36Vals, ok := fields[2].([]any)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOut: unexpected decoded type for minHopPriceX36: %T", fields[2])
	}
	minHopPriceX36 := make([]*big.Int, len(minHopPriceX36Vals))
	for i, v := range minHopPriceX36Vals {
		n, ok := v.(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOut: unexpected decoded type for minHopPriceX36[%d]: %T", i, v)
		}
		minHopPriceX36[i] = n
	}
	amountOut, ok := fields[3].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOut: unexpected decoded type for amountOut: %T", fields[3])
	}
	amountInMaximum, ok := fields[4].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionSwapExactOut: unexpected decoded type for amountInMaximum: %T", fields[4])
	}
	return NewExactOutputParams(currencyOut, path, minHopPriceX36, amountOut, amountInMaximum), nil
}

// EncodeActionSettle returns the params blob for Actions.SETTLE: abi.encode(currency, amount, payerIsUser).
func EncodeActionSettle(currency Currency, amount *big.Int, payerIsUser bool) ([]byte, error) {
	data, err := abi.Pack(abi.NewTypes(abi.Address, abi.Uint256, abi.Bool), currency, amount, payerIsUser)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionSettle: %w", err)
	}
	return data, nil
}

// DecodeActionSettle decodes a Actions.SETTLE params blob.
func DecodeActionSettle(data []byte) (Currency, *big.Int, bool, error) {
	vals, err := abi.Unpack(abi.NewTypes(abi.Address, abi.Uint256, abi.Bool), data)
	if err != nil {
		return nil, nil, false, fmt.Errorf("uniswapv4: decodeActionSettle: %w", err)
	}
	currency, ok := vals[0].(*types.Address)
	if !ok {
		return nil, nil, false, fmt.Errorf("uniswapv4: decodeActionSettle: unexpected decoded type for currency: %T", vals[0])
	}
	amount, ok := vals[1].(*big.Int)
	if !ok {
		return nil, nil, false, fmt.Errorf("uniswapv4: decodeActionSettle: unexpected decoded type for amount: %T", vals[1])
	}
	payerIsUser, ok := vals[2].(bool)
	if !ok {
		return nil, nil, false, fmt.Errorf("uniswapv4: decodeActionSettle: unexpected decoded type for payerIsUser: %T", vals[2])
	}
	return currency, amount, payerIsUser, nil
}

// EncodeActionSettleAll returns the params blob for Actions.SETTLE_ALL: abi.encode(currency, maxAmount).
func EncodeActionSettleAll(currency Currency, maxAmount *big.Int) ([]byte, error) {
	data, err := abi.Pack(abi.NewTypes(abi.Address, abi.Uint256), currency, maxAmount)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionSettleAll: %w", err)
	}
	return data, nil
}

// DecodeActionSettleAll decodes a Actions.SETTLE_ALL params blob.
func DecodeActionSettleAll(data []byte) (Currency, *big.Int, error) {
	vals, err := abi.Unpack(abi.NewTypes(abi.Address, abi.Uint256), data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: decodeActionSettleAll: %w", err)
	}
	currency, ok := vals[0].(*types.Address)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv4: decodeActionSettleAll: unexpected decoded type for currency: %T", vals[0])
	}
	maxAmount, ok := vals[1].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv4: decodeActionSettleAll: unexpected decoded type for maxAmount: %T", vals[1])
	}
	return currency, maxAmount, nil
}

// EncodeActionTake returns the params blob for Actions.TAKE: abi.encode(currency, recipient, amount).
func EncodeActionTake(currency Currency, recipient *types.Address, amount *big.Int) ([]byte, error) {
	data, err := abi.Pack(abi.NewTypes(abi.Address, abi.Address, abi.Uint256), currency, recipient, amount)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionTake: %w", err)
	}
	return data, nil
}

// DecodeActionTake decodes a Actions.TAKE params blob.
func DecodeActionTake(data []byte) (Currency, *types.Address, *big.Int, error) {
	vals, err := abi.Unpack(abi.NewTypes(abi.Address, abi.Address, abi.Uint256), data)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionTake: %w", err)
	}
	currency, ok := vals[0].(*types.Address)
	if !ok {
		return nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionTake: unexpected decoded type for currency: %T", vals[0])
	}
	recipient, ok := vals[1].(*types.Address)
	if !ok {
		return nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionTake: unexpected decoded type for recipient: %T", vals[1])
	}
	amount, ok := vals[2].(*big.Int)
	if !ok {
		return nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionTake: unexpected decoded type for amount: %T", vals[2])
	}
	return currency, recipient, amount, nil
}

// EncodeActionTakeAll returns the params blob for Actions.TAKE_ALL: abi.encode(currency, minAmount).
func EncodeActionTakeAll(currency Currency, minAmount *big.Int) ([]byte, error) {
	data, err := abi.Pack(abi.NewTypes(abi.Address, abi.Uint256), currency, minAmount)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionTakeAll: %w", err)
	}
	return data, nil
}

// DecodeActionTakeAll decodes a Actions.TAKE_ALL params blob.
func DecodeActionTakeAll(data []byte) (Currency, *big.Int, error) {
	vals, err := abi.Unpack(abi.NewTypes(abi.Address, abi.Uint256), data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: decodeActionTakeAll: %w", err)
	}
	currency, ok := vals[0].(*types.Address)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv4: decodeActionTakeAll: unexpected decoded type for currency: %T", vals[0])
	}
	minAmount, ok := vals[1].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv4: decodeActionTakeAll: unexpected decoded type for minAmount: %T", vals[1])
	}
	return currency, minAmount, nil
}

// EncodeActionTakePortion returns the params blob for Actions.TAKE_PORTION: abi.encode(currency, recipient, bips).
func EncodeActionTakePortion(currency Currency, recipient *types.Address, bips *big.Int) ([]byte, error) {
	data, err := abi.Pack(abi.NewTypes(abi.Address, abi.Address, abi.Uint256), currency, recipient, bips)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionTakePortion: %w", err)
	}
	return data, nil
}

// DecodeActionTakePortion decodes a Actions.TAKE_PORTION params blob.
func DecodeActionTakePortion(data []byte) (Currency, *types.Address, *big.Int, error) {
	vals, err := abi.Unpack(abi.NewTypes(abi.Address, abi.Address, abi.Uint256), data)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionTakePortion: %w", err)
	}
	currency, ok := vals[0].(*types.Address)
	if !ok {
		return nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionTakePortion: unexpected decoded type for currency: %T", vals[0])
	}
	recipient, ok := vals[1].(*types.Address)
	if !ok {
		return nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionTakePortion: unexpected decoded type for recipient: %T", vals[1])
	}
	bips, ok := vals[2].(*big.Int)
	if !ok {
		return nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionTakePortion: unexpected decoded type for bips: %T", vals[2])
	}
	return currency, recipient, bips, nil
}
