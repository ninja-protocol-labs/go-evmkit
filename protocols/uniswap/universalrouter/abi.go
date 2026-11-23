package universalrouter

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
)

var (
	// permitDetailsTuple/permitSingleTuple/permitBatchTuple mirror IAllowanceTransfer's PermitDetails/PermitSingle/PermitBatch.
	permitDetailsTuple = abi.Tuple(abi.Address, abi.Uint160, abi.Uint48, abi.Uint48)
	permitSingleTuple  = abi.Tuple(permitDetailsTuple, abi.Address, abi.Uint256)
	permitBatchTuple   = abi.Tuple(abi.Slice(permitDetailsTuple), abi.Address, abi.Uint256)

	// allowanceTransferDetailsTuple mirrors IAllowanceTransfer's AllowanceTransferDetails (from, to, amount, token).
	allowanceTransferDetailsTuple = abi.Tuple(abi.Address, abi.Address, abi.Uint160, abi.Address)

	// poolKeyTuple mirrors v4-core's PoolKey (currency0, currency1, fee, tickSpacing, hooks).
	poolKeyTuple = abi.Tuple(abi.Address, abi.Address, abi.Uint24, abi.Int24, abi.Address)

	// acrossV4DepositV3Tuple mirrors IUniversalRouter's AcrossV4DepositV3Params.
	acrossV4DepositV3Tuple = abi.Tuple(abi.Address, abi.Address, abi.Address, abi.Address, abi.Uint256, abi.Uint256, abi.Uint256, abi.Address, abi.Uint32, abi.Uint32, abi.Uint32, abi.Bytes, abi.Bool)
)

// Input layouts the 2.1 Dispatcher decodes, one per command shape.
var (
	v3SwapTypes                   = abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256, abi.Bytes, abi.Bool, abi.Slice(abi.Uint256))
	v2SwapTypes                   = abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256, abi.Slice(abi.Address), abi.Bool, abi.Slice(abi.Uint256))
	tokenRecipientAmountTypes     = abi.NewTypes(abi.Address, abi.Address, abi.Uint256)
	permit2TransferFromTypes      = abi.NewTypes(abi.Address, abi.Address, abi.Uint160)
	permit2PermitTypes            = abi.NewTypes(permitSingleTuple, abi.Bytes)
	permit2PermitBatchTypes       = abi.NewTypes(permitBatchTuple, abi.Bytes)
	permit2TransferFromBatchTypes = abi.NewTypes(abi.Slice(allowanceTransferDetailsTuple))
	recipientAmountTypes          = abi.NewTypes(abi.Address, abi.Uint256)
	ownerTokenAmountTypes         = abi.NewTypes(abi.Address, abi.Address, abi.Uint256)
	v4InitializePoolTypes         = abi.NewTypes(poolKeyTuple, abi.Uint160)
	subPlanTypes                  = abi.NewTypes(abi.Bytes, abi.Slice(abi.Bytes))
	acrossV4DepositV3Types        = abi.NewTypes(acrossV4DepositV3Tuple)
)

// Function signatures from IUniversalRouter and UniversalRouter, for encoding and decoding execute calldata.
var (
	executeFn             = abi.NewFunction("execute", abi.NewTypes(abi.Bytes, abi.Slice(abi.Bytes)), nil)
	executeWithDeadlineFn = abi.NewFunction("execute", abi.NewTypes(abi.Bytes, abi.Slice(abi.Bytes), abi.Uint256), nil)
	executeSignedFn       = abi.NewFunction("executeSigned", abi.NewTypes(abi.Bytes, abi.Slice(abi.Bytes), abi.Bytes32, abi.Bytes32, abi.Bool, abi.Bytes32, abi.Bytes, abi.Uint256), nil)
)

// Selectors of the position-manager calls the V3/V4 position-manager commands may carry (V3ToV4Migrator).
var (
	v3PermitSelector            = abi.NewFunction("permit", abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256, abi.Uint8, abi.Bytes32, abi.Bytes32), nil).Selector()
	v3DecreaseLiquiditySelector = abi.NewFunction("decreaseLiquidity", abi.NewTypes(abi.Tuple(abi.Uint256, abi.Uint128, abi.Uint256, abi.Uint256, abi.Uint256)), nil).Selector()
	v3CollectSelector           = abi.NewFunction("collect", abi.NewTypes(abi.Tuple(abi.Uint256, abi.Address, abi.Uint128, abi.Uint128)), nil).Selector()
	v3BurnSelector              = abi.NewFunction("burn", abi.NewTypes(abi.Uint256), nil).Selector()
	v4ModifyLiquiditiesFn       = abi.NewFunction("modifyLiquidities", abi.NewTypes(abi.Bytes, abi.Uint256), nil)
)

// ErrUnknownCommand is returned when decoding a command the 2.1 dispatcher would reject with InvalidCommandType.
var (
	ErrUnknownCommand  = errors.New("unknown command")
	ErrUnknownSelector = errors.New("unknown execute selector")
)

// EncodeExecute returns the calldata for execute(commands, inputs), which runs without a deadline.
func EncodeExecute(commands []byte, inputs [][]byte) ([]byte, error) {
	args := make([]any, len(inputs))
	for i, b := range inputs {
		if b == nil {
			b = []byte{}
		}
		args[i] = b
	}

	data, err := executeFn.EncodeCall(commands, args)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encodeExecute: %w", err)
	}
	return data, nil
}

// EncodeExecuteWithDeadline returns the calldata for execute(commands, inputs, deadline).
func EncodeExecuteWithDeadline(commands []byte, inputs [][]byte, deadline *big.Int) ([]byte, error) {
	args := make([]any, len(inputs))
	for i, b := range inputs {
		if b == nil {
			b = []byte{}
		}
		args[i] = b
	}

	data, err := executeWithDeadlineFn.EncodeCall(commands, args, deadline)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encodeExecuteWithDeadline: %w", err)
	}
	return data, nil
}

// EncodeExecuteSigned returns the calldata for executeSigned(commands, inputs, intent, data, verifySender, nonce, signature, deadline).
func EncodeExecuteSigned(commands []byte, inputs [][]byte, intent, data [32]byte, verifySender bool, nonce [32]byte, signature []byte, deadline *big.Int) ([]byte, error) {
	args := make([]any, len(inputs))
	for i, b := range inputs {
		if b == nil {
			b = []byte{}
		}
		args[i] = b
	}

	calldata, err := executeSignedFn.EncodeCall(commands, args, intent[:], data[:], verifySender, nonce[:], signature, deadline)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encodeExecuteSigned: %w", err)
	}
	return calldata, nil
}

// DecodeExecuteCall decodes execute or executeSigned calldata into its plan and call parameters.
func DecodeExecuteCall(calldata []byte) (*ExecuteCall, error) {
	if len(calldata) < 4 {
		return nil, fmt.Errorf("universalrouter: decodeExecuteCall: calldata is %d bytes, shorter than a selector", len(calldata))
	}

	var sel abi.Selector
	copy(sel[:], calldata[:4])

	var (
		vals     []any
		deadline *big.Int
		signed   *SignedRoute
		err      error
	)

	switch sel {
	case executeFn.Selector():
		vals, err = executeFn.DecodeCall(calldata)
		if err != nil {
			return nil, fmt.Errorf("universalrouter: decodeExecuteCall: %w", err)
		}
	case executeWithDeadlineFn.Selector():
		vals, err = executeWithDeadlineFn.DecodeCall(calldata)
		if err != nil {
			return nil, fmt.Errorf("universalrouter: decodeExecuteCall: %w", err)
		}

		var ok bool
		deadline, ok = vals[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("universalrouter: decodeExecuteCall: unexpected decoded type for deadline: %T", vals[2])
		}
	case executeSignedFn.Selector():
		vals, err = executeSignedFn.DecodeCall(calldata)
		if err != nil {
			return nil, fmt.Errorf("universalrouter: decodeExecuteCall: %w", err)
		}

		var words [3][32]byte
		for i, idx := range []int{2, 3, 5} {
			b, ok := vals[idx].([]byte)
			if !ok || len(b) != 32 {
				return nil, fmt.Errorf("universalrouter: decodeExecuteCall: unexpected decoded type for executeSigned argument %d: %T", idx, vals[idx])
			}
			copy(words[i][:], b)
		}

		verifySender, ok := vals[4].(bool)
		if !ok {
			return nil, fmt.Errorf("universalrouter: decodeExecuteCall: unexpected decoded type for verifySender: %T", vals[4])
		}

		signature, ok := vals[6].([]byte)
		if !ok {
			return nil, fmt.Errorf("universalrouter: decodeExecuteCall: unexpected decoded type for signature: %T", vals[6])
		}

		deadline, ok = vals[7].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("universalrouter: decodeExecuteCall: unexpected decoded type for deadline: %T", vals[7])
		}

		signed = NewSignedRoute(words[0], words[1], verifySender, words[2], signature)
	default:
		return nil, fmt.Errorf("universalrouter: decodeExecuteCall: %w: %x", ErrUnknownSelector, sel[:])
	}

	commands, ok := vals[0].([]byte)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decodeExecuteCall: unexpected decoded type for commands: %T", vals[0])
	}

	raw, ok := vals[1].([]any)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decodeExecuteCall: unexpected decoded type for inputs: %T", vals[1])
	}

	inputs := make([][]byte, len(raw))
	for i, v := range raw {
		b, ok := v.([]byte)
		if !ok {
			return nil, fmt.Errorf("universalrouter: decodeExecuteCall: unexpected decoded type for inputs[%d]: %T", i, v)
		}
		inputs[i] = b
	}

	plan, err := DecodePlan(commands, inputs)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: decodeExecuteCall: %w", err)
	}

	return NewExecuteCall(plan, deadline, signed), nil
}
