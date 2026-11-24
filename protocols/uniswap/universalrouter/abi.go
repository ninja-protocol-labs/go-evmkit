package universalrouter

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/protocols/permit2"
	v2 "github.com/ninja-protocol-labs/go-evmkit/protocols/uniswap/v2"
	v3 "github.com/ninja-protocol-labs/go-evmkit/protocols/uniswap/v3"
	v4 "github.com/ninja-protocol-labs/go-evmkit/protocols/uniswap/v4"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
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

// Function signatures of the Universal Router's read-only entry points.
var (
	noncesUsedFn         = abi.NewFunction("noncesUsed", abi.NewTypes(abi.Address, abi.Bytes32), abi.NewTypes(abi.Bool))
	signedRouteContextFn = abi.NewFunction("signedRouteContext", nil, abi.NewTypes(abi.Address, abi.Bytes32, abi.Bytes32))
)

// Custom errors from IUniversalRouter, Lock, RouteSigner and Dispatcher, for decoding reverts via abi.Error.Decode.
var (
	errExecutionFailed           = abi.NewError("ExecutionFailed", abi.NewTypes(abi.Uint256, abi.Bytes))
	errETHNotAccepted            = abi.NewError("ETHNotAccepted", abi.NewTypes())
	errTransactionDeadlinePassed = abi.NewError("TransactionDeadlinePassed", abi.NewTypes())
	errLengthMismatch            = abi.NewError("LengthMismatch", abi.NewTypes())
	errInvalidEthSender          = abi.NewError("InvalidEthSender", abi.NewTypes())
	errContractLocked            = abi.NewError("ContractLocked", abi.NewTypes())
	errNonceAlreadyUsed          = abi.NewError("NonceAlreadyUsed", abi.NewTypes())
	errInvalidCommandType        = abi.NewError("InvalidCommandType", abi.NewTypes(abi.Uint256))
	errBalanceTooLow             = abi.NewError("BalanceTooLow", abi.NewTypes())
)

// Custom errors from Payments, Permit2Payments, BipsLibrary, V3ToV4Migrator and ChainedActions, for decoding reverts via abi.Error.Decode.
var (
	errInsufficientToken        = abi.NewError("InsufficientToken", abi.NewTypes())
	errInsufficientETH          = abi.NewError("InsufficientETH", abi.NewTypes())
	errInvalidPortion           = abi.NewError("InvalidPortion", abi.NewTypes())
	errInvalidBips              = abi.NewError("InvalidBips", abi.NewTypes())
	errFromAddressIsNotOwner    = abi.NewError("FromAddressIsNotOwner", abi.NewTypes())
	errInvalidAction            = abi.NewError("InvalidAction", abi.NewTypes(abi.Bytes4))
	errOnlyMintAllowed          = abi.NewError("OnlyMintAllowed", abi.NewTypes())
	errNotAuthorizedForToken    = abi.NewError("NotAuthorizedForToken", abi.NewTypes(abi.Uint256))
	errSafeERC20FailedOperation = abi.NewError("SafeERC20FailedOperation", abi.NewTypes(abi.Address))
	errAddressEmptyCode         = abi.NewError("AddressEmptyCode", abi.NewTypes(abi.Address))
	errFailedInnerCall          = abi.NewError("FailedInnerCall", abi.NewTypes())
)

// Custom errors from the V2 and V3 swap modules, UniswapV2Library and BytesLib, for decoding reverts via abi.Error.Decode.
var (
	errInvalidReserves                 = abi.NewError("InvalidReserves", abi.NewTypes())
	errInvalidPath                     = abi.NewError("InvalidPath", abi.NewTypes())
	errSliceOutOfBounds                = abi.NewError("SliceOutOfBounds", abi.NewTypes())
	errV2TooLittleReceived             = abi.NewError("V2TooLittleReceived", abi.NewTypes())
	errV2TooMuchRequested              = abi.NewError("V2TooMuchRequested", abi.NewTypes())
	errV2InvalidPath                   = abi.NewError("V2InvalidPath", abi.NewTypes())
	errV2TooLittleReceivedPerHop       = abi.NewError("V2TooLittleReceivedPerHop", abi.NewTypes(abi.Uint256, abi.Uint256, abi.Uint256))
	errV2InvalidHopPriceLength         = abi.NewError("V2InvalidHopPriceLength", abi.NewTypes())
	errV3InvalidSwap                   = abi.NewError("V3InvalidSwap", abi.NewTypes())
	errV3TooLittleReceived             = abi.NewError("V3TooLittleReceived", abi.NewTypes())
	errV3TooMuchRequested              = abi.NewError("V3TooMuchRequested", abi.NewTypes())
	errV3InvalidAmountOut              = abi.NewError("V3InvalidAmountOut", abi.NewTypes())
	errV3InvalidCaller                 = abi.NewError("V3InvalidCaller", abi.NewTypes())
	errV3TooLittleReceivedPerHop       = abi.NewError("V3TooLittleReceivedPerHop", abi.NewTypes(abi.Uint256, abi.Uint256, abi.Uint256))
	errV3TooMuchRequestedPerHop        = abi.NewError("V3TooMuchRequestedPerHop", abi.NewTypes(abi.Uint256, abi.Uint256, abi.Uint256))
	errV3HopPriceAndPathLengthMismatch = abi.NewError("V3HopPriceAndPathLengthMismatch", abi.NewTypes())
)

// Custom errors from OpenZeppelin's ECDSA, raised when executeSigned cannot recover a signer.
var (
	errECDSAInvalidSignature       = abi.NewError("ECDSAInvalidSignature", abi.NewTypes())
	errECDSAInvalidSignatureLength = abi.NewError("ECDSAInvalidSignatureLength", abi.NewTypes(abi.Uint256))
	errECDSAInvalidSignatureS      = abi.NewError("ECDSAInvalidSignatureS", abi.NewTypes(abi.Bytes32))
)

// ErrExecutionFailed, etc. mirror the above custom errors, for errors.Is; ErrUnknown covers reverts no router, Permit2 or v4 mapper recognises.
var (
	ErrExecutionFailed                 = errors.New("execution failed")
	ErrETHNotAccepted                  = errors.New("eth not accepted")
	ErrTransactionDeadlinePassed       = errors.New("transaction deadline passed")
	ErrLengthMismatch                  = errors.New("length mismatch")
	ErrInvalidEthSender                = errors.New("invalid eth sender")
	ErrContractLocked                  = errors.New("contract locked")
	ErrNonceAlreadyUsed                = errors.New("nonce already used")
	ErrInvalidCommandType              = errors.New("invalid command type")
	ErrBalanceTooLow                   = errors.New("balance too low")
	ErrInsufficientToken               = errors.New("insufficient token")
	ErrInsufficientETH                 = errors.New("insufficient eth")
	ErrInvalidPortion                  = errors.New("invalid portion")
	ErrInvalidBips                     = errors.New("invalid bips")
	ErrFromAddressIsNotOwner           = errors.New("from address is not owner")
	ErrInvalidAction                   = errors.New("invalid position manager action")
	ErrOnlyMintAllowed                 = errors.New("only mint allowed")
	ErrNotAuthorizedForToken           = errors.New("not authorized for token")
	ErrSafeERC20FailedOperation        = errors.New("safe erc20 operation failed")
	ErrAddressEmptyCode                = errors.New("address has no code")
	ErrFailedInnerCall                 = errors.New("inner call failed")
	ErrInvalidReserves                 = errors.New("invalid reserves")
	ErrInvalidPath                     = errors.New("invalid path")
	ErrSliceOutOfBounds                = errors.New("slice out of bounds")
	ErrV2TooLittleReceived             = errors.New("v2 too little received")
	ErrV2TooMuchRequested              = errors.New("v2 too much requested")
	ErrV2InvalidPath                   = errors.New("v2 invalid path")
	ErrV2TooLittleReceivedPerHop       = errors.New("v2 too little received per hop")
	ErrV2InvalidHopPriceLength         = errors.New("v2 invalid hop price length")
	ErrV3InvalidSwap                   = errors.New("v3 invalid swap")
	ErrV3TooLittleReceived             = errors.New("v3 too little received")
	ErrV3TooMuchRequested              = errors.New("v3 too much requested")
	ErrV3InvalidAmountOut              = errors.New("v3 invalid amount out")
	ErrV3InvalidCaller                 = errors.New("v3 invalid caller")
	ErrV3TooLittleReceivedPerHop       = errors.New("v3 too little received per hop")
	ErrV3TooMuchRequestedPerHop        = errors.New("v3 too much requested per hop")
	ErrV3HopPriceAndPathLengthMismatch = errors.New("v3 hop price and path length mismatch")
	ErrECDSAInvalidSignature           = errors.New("ecdsa invalid signature")
	ErrECDSAInvalidSignatureLength     = errors.New("ecdsa invalid signature length")
	ErrECDSAInvalidSignatureS          = errors.New("ecdsa invalid signature s")
	ErrUnknown                         = errors.New("unknown revert")
)

// EncodeNoncesUsed returns the calldata for noncesUsed(user, nonce).
func EncodeNoncesUsed(user *types.Address, nonce [32]byte) ([]byte, error) {
	data, err := noncesUsedFn.EncodeCall(user, nonce[:])
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encodeNoncesUsed: %w", err)
	}
	return data, nil
}

// DecodeNoncesUsed decodes the return data of noncesUsed.
func DecodeNoncesUsed(data []byte) (bool, error) {
	used, err := noncesUsedFn.DecodeSingleReturn[bool](data)
	if err != nil {
		return false, fmt.Errorf("universalrouter: decodeNoncesUsed: %w", err)
	}
	return used, nil
}

// EncodeSignedRouteContext returns the calldata for signedRouteContext().
func EncodeSignedRouteContext() ([]byte, error) {
	data, err := signedRouteContextFn.EncodeCall()
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encodeSignedRouteContext: %w", err)
	}
	return data, nil
}

// DecodeSignedRouteContext decodes the return data of signedRouteContext: signer, intent and data, all zero outside an executeSigned call.
func DecodeSignedRouteContext(data []byte) (*RouteContext, error) {
	vals, err := signedRouteContextFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: decodeSignedRouteContext: %w", err)
	}

	signer, ok := vals[0].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decodeSignedRouteContext: unexpected decoded type for signer: %T", vals[0])
	}

	intent, ok := vals[1].([]byte)
	if !ok || len(intent) != 32 {
		return nil, fmt.Errorf("universalrouter: decodeSignedRouteContext: unexpected decoded type for intent: %T", vals[1])
	}

	data32, ok := vals[2].([]byte)
	if !ok || len(data32) != 32 {
		return nil, fmt.Errorf("universalrouter: decodeSignedRouteContext: unexpected decoded type for data: %T", vals[2])
	}

	return NewRouteContext(signer, [32]byte(intent), [32]byte(data32)), nil
}

// MapRevertError maps a revert in err to its UR, Permit2, v4, v3 or v2 sentinel, else ErrUnknown.
func MapRevertError(err error) error {
	rpcErr, ok := errors.AsType[*rpc.ResponseError](err)
	if ok && rpcErr == nil {
		return nil
	}
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
	case errExecutionFailed.Selector():
		vals, decodeErr := errExecutionFailed.Decode(data)
		if decodeErr != nil || len(vals) != 2 {
			return fmt.Errorf("%w: %w", ErrExecutionFailed, err)
		}

		commandIndex, ok := vals[0].(*big.Int)
		if !ok {
			return fmt.Errorf("%w: %w", ErrExecutionFailed, err)
		}

		message, ok := vals[1].([]byte)
		if !ok {
			return fmt.Errorf("%w: %w", ErrExecutionFailed, err)
		}

		var cause error
		if len(message) >= 4 {
			inner := &rpc.ResponseError{Code: rpcErr.Code, Message: rpcErr.Message, Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(message) + `"`)}
			cause = MapRevertError(inner)
		}

		return fmt.Errorf("%w: %w", NewExecutionFailedError(commandIndex, message, cause), err)
	case errETHNotAccepted.Selector():
		return fmt.Errorf("%w: %w", ErrETHNotAccepted, err)
	case errTransactionDeadlinePassed.Selector():
		return fmt.Errorf("%w: %w", ErrTransactionDeadlinePassed, err)
	case errLengthMismatch.Selector():
		return fmt.Errorf("%w: %w", ErrLengthMismatch, err)
	case errInvalidEthSender.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidEthSender, err)
	case errContractLocked.Selector():
		return fmt.Errorf("%w: %w", ErrContractLocked, err)
	case errNonceAlreadyUsed.Selector():
		return fmt.Errorf("%w: %w", ErrNonceAlreadyUsed, err)
	case errInvalidCommandType.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidCommandType, err)
	case errBalanceTooLow.Selector():
		return fmt.Errorf("%w: %w", ErrBalanceTooLow, err)
	case errInsufficientToken.Selector():
		return fmt.Errorf("%w: %w", ErrInsufficientToken, err)
	case errInsufficientETH.Selector():
		return fmt.Errorf("%w: %w", ErrInsufficientETH, err)
	case errInvalidPortion.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidPortion, err)
	case errInvalidBips.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidBips, err)
	case errFromAddressIsNotOwner.Selector():
		return fmt.Errorf("%w: %w", ErrFromAddressIsNotOwner, err)
	case errInvalidAction.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidAction, err)
	case errOnlyMintAllowed.Selector():
		return fmt.Errorf("%w: %w", ErrOnlyMintAllowed, err)
	case errNotAuthorizedForToken.Selector():
		return fmt.Errorf("%w: %w", ErrNotAuthorizedForToken, err)
	case errSafeERC20FailedOperation.Selector():
		return fmt.Errorf("%w: %w", ErrSafeERC20FailedOperation, err)
	case errAddressEmptyCode.Selector():
		return fmt.Errorf("%w: %w", ErrAddressEmptyCode, err)
	case errFailedInnerCall.Selector():
		return fmt.Errorf("%w: %w", ErrFailedInnerCall, err)
	case errInvalidReserves.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidReserves, err)
	case errInvalidPath.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidPath, err)
	case errSliceOutOfBounds.Selector():
		return fmt.Errorf("%w: %w", ErrSliceOutOfBounds, err)
	case errV2TooLittleReceived.Selector():
		return fmt.Errorf("%w: %w", ErrV2TooLittleReceived, err)
	case errV2TooMuchRequested.Selector():
		return fmt.Errorf("%w: %w", ErrV2TooMuchRequested, err)
	case errV2InvalidPath.Selector():
		return fmt.Errorf("%w: %w", ErrV2InvalidPath, err)
	case errV2TooLittleReceivedPerHop.Selector():
		return fmt.Errorf("%w: %w", ErrV2TooLittleReceivedPerHop, err)
	case errV2InvalidHopPriceLength.Selector():
		return fmt.Errorf("%w: %w", ErrV2InvalidHopPriceLength, err)
	case errV3InvalidSwap.Selector():
		return fmt.Errorf("%w: %w", ErrV3InvalidSwap, err)
	case errV3TooLittleReceived.Selector():
		return fmt.Errorf("%w: %w", ErrV3TooLittleReceived, err)
	case errV3TooMuchRequested.Selector():
		return fmt.Errorf("%w: %w", ErrV3TooMuchRequested, err)
	case errV3InvalidAmountOut.Selector():
		return fmt.Errorf("%w: %w", ErrV3InvalidAmountOut, err)
	case errV3InvalidCaller.Selector():
		return fmt.Errorf("%w: %w", ErrV3InvalidCaller, err)
	case errV3TooLittleReceivedPerHop.Selector():
		return fmt.Errorf("%w: %w", ErrV3TooLittleReceivedPerHop, err)
	case errV3TooMuchRequestedPerHop.Selector():
		return fmt.Errorf("%w: %w", ErrV3TooMuchRequestedPerHop, err)
	case errV3HopPriceAndPathLengthMismatch.Selector():
		return fmt.Errorf("%w: %w", ErrV3HopPriceAndPathLengthMismatch, err)
	case errECDSAInvalidSignature.Selector():
		return fmt.Errorf("%w: %w", ErrECDSAInvalidSignature, err)
	case errECDSAInvalidSignatureLength.Selector():
		return fmt.Errorf("%w: %w", ErrECDSAInvalidSignatureLength, err)
	case errECDSAInvalidSignatureS.Selector():
		return fmt.Errorf("%w: %w", ErrECDSAInvalidSignatureS, err)
	}

	if wrapped := permit2.MapRevertError(err); !errors.Is(wrapped, permit2.ErrUnknown) {
		return wrapped
	}

	if wrapped := v4.MapRevertError(err); !errors.Is(wrapped, v4.ErrUnknown) {
		return wrapped
	}

	if _, reasonErr := abi.DecodeRevert(data); reasonErr == nil {
		if wrapped := v3.MapRevertError(err); !errors.Is(wrapped, v3.ErrReverted) {
			return wrapped
		}

		if wrapped := v2.MapRevertError(err); !errors.Is(wrapped, v2.ErrUnknown) {
			return wrapped
		}
	}

	return fmt.Errorf("%w: %w", ErrUnknown, err)
}
