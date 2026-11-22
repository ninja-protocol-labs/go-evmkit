package v4

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/protocols/permit2"
)

var (
	// modifyLiquiditiesFn mirrors IPositionManager's modifyLiquidities(bytes unlockData, uint256
	// deadline): the standard entry point, which unlocks PoolManager and batches unlockData's actions.
	modifyLiquiditiesFn = abi.NewFunction("modifyLiquidities", abi.NewTypes(abi.Bytes, abi.Uint256), nil)

	// modifyLiquiditiesWithoutUnlockFn mirrors IPositionManager's modifyLiquiditiesWithoutUnlock
	// (bytes actions, bytes[] params): must be called by a contract that has already unlocked PoolManager.
	modifyLiquiditiesWithoutUnlockFn = abi.NewFunction("modifyLiquiditiesWithoutUnlock", abi.NewTypes(abi.Bytes, abi.Slice(abi.Bytes)), nil)

	// nextTokenIdFn mirrors IPositionManager's nextTokenId() view.
	nextTokenIdFn = abi.NewFunction("nextTokenId", nil, abi.NewTypes(abi.Uint256))

	// getPositionLiquidityFn mirrors IPositionManager's getPositionLiquidity(uint256) view.
	getPositionLiquidityFn = abi.NewFunction("getPositionLiquidity", abi.NewTypes(abi.Uint256), abi.NewTypes(abi.Uint128))

	// getPoolAndPositionInfoFn mirrors IPositionManager's getPoolAndPositionInfo(uint256) view,
	// returning (PoolKey, PositionInfo); PositionInfo is a user-defined uint256 on the Solidity side.
	getPoolAndPositionInfoFn = abi.NewFunction("getPoolAndPositionInfo", abi.NewTypes(abi.Uint256), abi.NewTypes(poolKeyTuple, abi.Uint256))

	// positionInfoFn mirrors IPositionManager's positionInfo(uint256) view.
	positionInfoFn = abi.NewFunction("positionInfo", abi.NewTypes(abi.Uint256), abi.NewTypes(abi.Uint256))

	// poolKeysFn mirrors PositionManager's auto-generated public getter for its
	// poolKeys(bytes25 => PoolKey) mapping: looks up a pool's full PoolKey from the truncated
	// 25-byte poolId stored in a PositionInfo.
	poolKeysFn = abi.NewFunction("poolKeys", abi.NewTypes(abi.Bytes25), abi.NewTypes(abi.Address, abi.Address, abi.Uint24, abi.Int24, abi.Address))

	// subscribeFn mirrors INotifier's subscribe(uint256 tokenId, address newSubscriber, bytes data).
	subscribeFn = abi.NewFunction("subscribe", abi.NewTypes(abi.Uint256, abi.Address, abi.Bytes), nil)

	// unsubscribeFn mirrors INotifier's unsubscribe(uint256 tokenId).
	unsubscribeFn = abi.NewFunction("unsubscribe", abi.NewTypes(abi.Uint256), nil)

	// subscriberFn mirrors INotifier's subscriber(uint256 tokenId) view.
	subscriberFn = abi.NewFunction("subscriber", abi.NewTypes(abi.Uint256), abi.NewTypes(abi.Address))

	// unsubscribeGasLimitFn mirrors INotifier's unsubscribeGasLimit() view.
	unsubscribeGasLimitFn = abi.NewFunction("unsubscribeGasLimit", nil, abi.NewTypes(abi.Uint256))

	// initializePoolFn mirrors IPoolInitializer_v4's initializePool(PoolKey key, uint160 sqrtPriceX96).
	initializePoolFn = abi.NewFunction("initializePool", abi.NewTypes(poolKeyTuple, abi.Uint160), abi.NewTypes(abi.Int24))

	// multicallFn mirrors IMulticall_v4's multicall(bytes[] data).
	multicallFn = abi.NewFunction("multicall", abi.NewTypes(abi.Slice(abi.Bytes)), abi.NewTypes(abi.Slice(abi.Bytes)))

	// poolManagerFn mirrors IImmutableState's poolManager() view.
	poolManagerFn = abi.NewFunction("poolManager", nil, abi.NewTypes(abi.Address))

	// weth9Fn mirrors NativeWrapper's auto-generated public getter for its immutable WETH9 field.
	weth9Fn = abi.NewFunction("WETH9", nil, abi.NewTypes(abi.Address))

	// permit2ContractFn mirrors Permit2Forwarder's auto-generated public getter for its immutable
	// permit2 field: the Permit2 contract this PositionManager forwards permit/permitBatch to.
	permit2ContractFn = abi.NewFunction("permit2", nil, abi.NewTypes(abi.Address))

	// noncesFn mirrors IUnorderedNonce's nonces(address owner, uint256 word) view.
	noncesFn = abi.NewFunction("nonces", abi.NewTypes(abi.Address, abi.Uint256), abi.NewTypes(abi.Uint256))

	// revokeNonceFn mirrors IUnorderedNonce's revokeNonce(uint256 nonce).
	revokeNonceFn = abi.NewFunction("revokeNonce", abi.NewTypes(abi.Uint256), nil)

	// erc721PermitFn mirrors IERC721Permit_v4's permit(address spender, uint256 tokenId, uint256
	// deadline, uint256 nonce, bytes signature): single-tokenId approval via signature.
	erc721PermitFn = abi.NewFunction("permit", abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256, abi.Uint256, abi.Bytes), nil)

	// erc721PermitForAllFn mirrors IERC721Permit_v4's permitForAll(address owner, address
	// operator, bool approved, uint256 deadline, uint256 nonce, bytes signature).
	erc721PermitForAllFn = abi.NewFunction("permitForAll", abi.NewTypes(abi.Address, abi.Address, abi.Bool, abi.Uint256, abi.Uint256, abi.Bytes), nil)

	// permitDetailsTuple/permitSingleTuple/permitBatchTuple mirror IAllowanceTransfer's
	// PermitDetails/PermitSingle/PermitBatch structs (same shapes as the permit2 package's own,
	// redeclared here per this repo's per-package duplication convention for ABI shapes).
	permitDetailsTuple = abi.Tuple(abi.Address, abi.Uint160, abi.Uint48, abi.Uint48)
	permitSingleTuple  = abi.Tuple(permitDetailsTuple, abi.Address, abi.Uint256)
	permitBatchTuple   = abi.Tuple(abi.Slice(permitDetailsTuple), abi.Address, abi.Uint256)

	// permit2PermitFn mirrors IPermit2Forwarder's permit(address owner, PermitSingle permitSingle,
	// bytes signature), forwarding a single-token permit2 approval.
	permit2PermitFn = abi.NewFunction("permit", abi.NewTypes(abi.Address, permitSingleTuple, abi.Bytes), abi.NewTypes(abi.Bytes))

	// permit2PermitBatchFn mirrors IPermit2Forwarder's permitBatch(address owner, PermitBatch
	// permitBatch, bytes signature), forwarding a batch permit2 approval.
	permit2PermitBatchFn = abi.NewFunction("permitBatch", abi.NewTypes(abi.Address, permitBatchTuple, abi.Bytes), abi.NewTypes(abi.Bytes))
)

// Custom errors from IPositionManager, for decoding reverts via abi.Error.Decode.
var (
	errNotApproved             = abi.NewError("NotApproved", abi.NewTypes(abi.Address))
	errDeadlinePassed          = abi.NewError("DeadlinePassed", abi.NewTypes(abi.Uint256))
	errPoolManagerMustBeLocked = abi.NewError("PoolManagerMustBeLocked", abi.NewTypes())
)

// Custom errors from IERC721Permit_v4, for decoding reverts via abi.Error.Decode.
var (
	errSignatureDeadlineExpired = abi.NewError("SignatureDeadlineExpired", abi.NewTypes())
	errNoSelfPermit             = abi.NewError("NoSelfPermit", abi.NewTypes())
	errUnauthorized             = abi.NewError("Unauthorized", abi.NewTypes())
)

// Custom errors from INotifier, for decoding reverts via abi.Error.Decode.
var (
	errNotSubscribed                       = abi.NewError("NotSubscribed", abi.NewTypes())
	errNoCodeSubscriber                    = abi.NewError("NoCodeSubscriber", abi.NewTypes())
	errGasLimitTooLow                      = abi.NewError("GasLimitTooLow", abi.NewTypes())
	errSubscriptionReverted                = abi.NewError("SubscriptionReverted", abi.NewTypes(abi.Address, abi.Bytes))
	errModifyLiquidityNotificationReverted = abi.NewError("ModifyLiquidityNotificationReverted", abi.NewTypes(abi.Address, abi.Bytes))
	errBurnNotificationReverted            = abi.NewError("BurnNotificationReverted", abi.NewTypes(abi.Address, abi.Bytes))
	errAlreadySubscribed                   = abi.NewError("AlreadySubscribed", abi.NewTypes(abi.Uint256, abi.Address))
)

// Custom errors from IUnorderedNonce, for decoding reverts via abi.Error.Decode.
var (
	errNonceAlreadyUsed = abi.NewError("NonceAlreadyUsed", abi.NewTypes())
)

// Custom errors from ReentrancyLock and NativeWrapper, for decoding reverts via abi.Error.Decode.
var (
	errContractLocked   = abi.NewError("ContractLocked", abi.NewTypes())
	errInvalidEthSender = abi.NewError("InvalidEthSender", abi.NewTypes())
)

// Custom errors from DeltaResolver, for decoding reverts via abi.Error.Decode. Shared with
// router.go's Router (same base contract), but not mapped there; mapped here for PositionManager.
var (
	errDeltaNotPositive    = abi.NewError("DeltaNotPositive", abi.NewTypes(abi.Address))
	errDeltaNotNegative    = abi.NewError("DeltaNotNegative", abi.NewTypes(abi.Address))
	errInsufficientBalance = abi.NewError("InsufficientBalance", abi.NewTypes())
)

// Custom errors from SlippageCheck, for decoding reverts via abi.Error.Decode.
var (
	errMaximumAmountExceeded     = abi.NewError("MaximumAmountExceeded", abi.NewTypes(abi.Uint128, abi.Uint128))
	errMinimumAmountInsufficient = abi.NewError("MinimumAmountInsufficient", abi.NewTypes(abi.Uint128, abi.Uint128))
)

// Custom errors from permit2's SignatureVerification, used internally by ERC721Permit_v4.permit/
// permitForAll's signature checks, for decoding reverts via abi.Error.Decode.
var (
	errInvalidSignatureLength   = abi.NewError("InvalidSignatureLength", abi.NewTypes())
	errInvalidSignature         = abi.NewError("InvalidSignature", abi.NewTypes())
	errInvalidSigner            = abi.NewError("InvalidSigner", abi.NewTypes())
	errInvalidContractSignature = abi.NewError("InvalidContractSignature", abi.NewTypes())
)

// ErrNotApproved, etc. mirror the above custom errors, for errors.Is.
var (
	ErrNotApproved                         = errors.New("uniswapv4: caller not approved to modify position")
	ErrDeadlinePassed                      = errors.New("uniswapv4: deadline passed")
	ErrPoolManagerMustBeLocked             = errors.New("uniswapv4: pool manager must be locked")
	ErrSignatureDeadlineExpired            = errors.New("uniswapv4: signature deadline expired")
	ErrNoSelfPermit                        = errors.New("uniswapv4: cannot self-permit")
	ErrUnauthorized                        = errors.New("uniswapv4: unauthorized")
	ErrNotSubscribed                       = errors.New("uniswapv4: not subscribed")
	ErrNoCodeSubscriber                    = errors.New("uniswapv4: subscriber has no code")
	ErrGasLimitTooLow                      = errors.New("uniswapv4: gas limit too low to notify unsubscribe")
	ErrSubscriptionReverted                = errors.New("uniswapv4: subscription reverted")
	ErrModifyLiquidityNotificationReverted = errors.New("uniswapv4: modify-liquidity notification reverted")
	ErrBurnNotificationReverted            = errors.New("uniswapv4: burn notification reverted")
	ErrAlreadySubscribed                   = errors.New("uniswapv4: already subscribed")
	ErrNonceAlreadyUsed                    = errors.New("uniswapv4: nonce already used")
	ErrContractLocked                      = errors.New("uniswapv4: contract locked")
	ErrInvalidEthSender                    = errors.New("uniswapv4: invalid eth sender")
	ErrDeltaNotPositive                    = errors.New("uniswapv4: delta not positive")
	ErrDeltaNotNegative                    = errors.New("uniswapv4: delta not negative")
	ErrInsufficientBalance                 = errors.New("uniswapv4: insufficient balance")
	ErrMaximumAmountExceeded               = errors.New("uniswapv4: maximum amount exceeded")
	ErrMinimumAmountInsufficient           = errors.New("uniswapv4: minimum amount insufficient")
	ErrInvalidSignatureLength              = errors.New("uniswapv4: invalid signature length")
	ErrInvalidSignature                    = errors.New("uniswapv4: invalid signature")
	ErrInvalidSigner                       = errors.New("uniswapv4: invalid signer")
	ErrInvalidContractSignature            = errors.New("uniswapv4: invalid contract signature")
)

// EncodeModifyLiquidities returns the calldata for modifyLiquidities(abi.encode(actions, params),
// deadline): the standard entry point for batching position actions.
func EncodeModifyLiquidities(actions []byte, params [][]byte, deadline *big.Int) ([]byte, error) {
	unlockData, err := EncodeActions(actions, params)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeModifyLiquidities: %w", err)
	}
	data, err := modifyLiquiditiesFn.EncodeCall(unlockData, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeModifyLiquidities: %w", err)
	}
	return data, nil
}

// EncodeModifyLiquiditiesWithoutUnlock returns the calldata for modifyLiquiditiesWithoutUnlock
// (actions, params). Only callable by a contract that has already unlocked PoolManager itself.
func EncodeModifyLiquiditiesWithoutUnlock(actions []byte, params [][]byte) ([]byte, error) {
	data, err := modifyLiquiditiesWithoutUnlockFn.EncodeCall(actions, params)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeModifyLiquiditiesWithoutUnlock: %w", err)
	}
	return data, nil
}

// EncodeNextTokenID returns the calldata for nextTokenId().
func EncodeNextTokenID() ([]byte, error) {
	data, err := nextTokenIdFn.EncodeCall()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeNextTokenID: %w", err)
	}
	return data, nil
}

// DecodeNextTokenID decodes the return data of nextTokenId: the next tokenId to be minted.
func DecodeNextTokenID(data []byte) (*big.Int, error) {
	tokenID, err := nextTokenIdFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeNextTokenID: %w", err)
	}
	return tokenID, nil
}

// EncodeGetPositionLiquidity returns the calldata for getPositionLiquidity(tokenId).
func EncodeGetPositionLiquidity(tokenID *big.Int) ([]byte, error) {
	data, err := getPositionLiquidityFn.EncodeCall(tokenID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeGetPositionLiquidity: %w", err)
	}
	return data, nil
}

// DecodeGetPositionLiquidity decodes the return data of getPositionLiquidity: the position's
// liquidity amount.
func DecodeGetPositionLiquidity(data []byte) (*big.Int, error) {
	liquidity, err := getPositionLiquidityFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeGetPositionLiquidity: %w", err)
	}
	return liquidity, nil
}

// EncodeGetPoolAndPositionInfo returns the calldata for getPoolAndPositionInfo(tokenId).
func EncodeGetPoolAndPositionInfo(tokenID *big.Int) ([]byte, error) {
	data, err := getPoolAndPositionInfoFn.EncodeCall(tokenID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeGetPoolAndPositionInfo: %w", err)
	}
	return data, nil
}

// DecodeGetPoolAndPositionInfo decodes the return data of getPoolAndPositionInfo: the position's
// pool key and its packed PositionInfo.
func DecodeGetPoolAndPositionInfo(data []byte) (*PoolKey, *PositionInfo, error) {
	vals, err := getPoolAndPositionInfoFn.DecodeReturn(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: decodeGetPoolAndPositionInfo: %w", err)
	}
	poolKeyFields, ok := vals[0].([]any)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv4: decodeGetPoolAndPositionInfo: unexpected decoded type for poolKey: %T", vals[0])
	}
	currency0, ok := poolKeyFields[0].(*types.Address)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv4: decodeGetPoolAndPositionInfo: unexpected decoded type for currency0: %T", poolKeyFields[0])
	}
	currency1, ok := poolKeyFields[1].(*types.Address)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv4: decodeGetPoolAndPositionInfo: unexpected decoded type for currency1: %T", poolKeyFields[1])
	}
	fee, ok := poolKeyFields[2].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv4: decodeGetPoolAndPositionInfo: unexpected decoded type for fee: %T", poolKeyFields[2])
	}
	tickSpacing, ok := poolKeyFields[3].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv4: decodeGetPoolAndPositionInfo: unexpected decoded type for tickSpacing: %T", poolKeyFields[3])
	}
	hooks, ok := poolKeyFields[4].(*types.Address)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv4: decodeGetPoolAndPositionInfo: unexpected decoded type for hooks: %T", poolKeyFields[4])
	}
	raw, ok := vals[1].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv4: decodeGetPoolAndPositionInfo: unexpected decoded type for positionInfo: %T", vals[1])
	}
	word := make([]byte, 32)
	b := raw.Bytes()
	if len(b) > 32 {
		b = b[len(b)-32:]
	}
	copy(word[32-len(b):], b)

	var poolID [25]byte
	copy(poolID[:], word[0:25])

	tickUpper := new(big.Int).SetBytes(word[25:28])
	if word[25]&0x80 != 0 {
		tickUpper.Sub(tickUpper, new(big.Int).Lsh(big.NewInt(1), 24))
	}

	tickLower := new(big.Int).SetBytes(word[28:31])
	if word[28]&0x80 != 0 {
		tickLower.Sub(tickLower, new(big.Int).Lsh(big.NewInt(1), 24))
	}

	hasSubscriber := word[31] != 0

	return NewPoolKey(currency0, currency1, fee, tickSpacing, hooks), NewPositionInfo(poolID, tickLower, tickUpper, hasSubscriber), nil
}

// EncodePositionInfo returns the calldata for positionInfo(tokenId).
func EncodePositionInfo(tokenID *big.Int) ([]byte, error) {
	data, err := positionInfoFn.EncodeCall(tokenID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodePositionInfo: %w", err)
	}
	return data, nil
}

// DecodePositionInfo decodes the return data of positionInfo: the position's packed PositionInfo.
// PositionInfoLibrary.sol packs it as upper 200 bits truncated poolId | next 24 bits tickUpper |
// next 24 bits tickLower | lowest 8 bits hasSubscriber; each field's bit offset is byte-aligned,
// so this is plain slicing of the 32-byte big-endian word, with ticks sign-extended from int24.
func DecodePositionInfo(data []byte) (*PositionInfo, error) {
	raw, err := positionInfoFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodePositionInfo: %w", err)
	}
	word := make([]byte, 32)
	b := raw.Bytes()
	if len(b) > 32 {
		b = b[len(b)-32:]
	}
	copy(word[32-len(b):], b)

	var poolID [25]byte
	copy(poolID[:], word[0:25])

	tickUpper := new(big.Int).SetBytes(word[25:28])
	if word[25]&0x80 != 0 {
		tickUpper.Sub(tickUpper, new(big.Int).Lsh(big.NewInt(1), 24))
	}

	tickLower := new(big.Int).SetBytes(word[28:31])
	if word[28]&0x80 != 0 {
		tickLower.Sub(tickLower, new(big.Int).Lsh(big.NewInt(1), 24))
	}

	hasSubscriber := word[31] != 0

	return NewPositionInfo(poolID, tickLower, tickUpper, hasSubscriber), nil
}

// EncodePoolKeys returns the calldata for poolKeys(poolID).
func EncodePoolKeys(poolID [25]byte) ([]byte, error) {
	data, err := poolKeysFn.EncodeCall(poolID[:])
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodePoolKeys: %w", err)
	}
	return data, nil
}

// DecodePoolKeys decodes the return data of poolKeys: the full PoolKey for a truncated poolId.
func DecodePoolKeys(data []byte) (*PoolKey, error) {
	vals, err := poolKeysFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodePoolKeys: %w", err)
	}
	currency0, ok := vals[0].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodePoolKeys: unexpected decoded type for currency0: %T", vals[0])
	}
	currency1, ok := vals[1].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodePoolKeys: unexpected decoded type for currency1: %T", vals[1])
	}
	fee, ok := vals[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodePoolKeys: unexpected decoded type for fee: %T", vals[2])
	}
	tickSpacing, ok := vals[3].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodePoolKeys: unexpected decoded type for tickSpacing: %T", vals[3])
	}
	hooks, ok := vals[4].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodePoolKeys: unexpected decoded type for hooks: %T", vals[4])
	}
	return NewPoolKey(currency0, currency1, fee, tickSpacing, hooks), nil
}

// EncodeSubscribe returns the calldata for subscribe(tokenId, newSubscriber, data).
func EncodeSubscribe(tokenID *big.Int, newSubscriber *types.Address, data []byte) ([]byte, error) {
	calldata, err := subscribeFn.EncodeCall(tokenID, newSubscriber, data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeSubscribe: %w", err)
	}
	return calldata, nil
}

// EncodeUnsubscribe returns the calldata for unsubscribe(tokenId).
func EncodeUnsubscribe(tokenID *big.Int) ([]byte, error) {
	calldata, err := unsubscribeFn.EncodeCall(tokenID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeUnsubscribe: %w", err)
	}
	return calldata, nil
}

// EncodeSubscriber returns the calldata for subscriber(tokenId).
func EncodeSubscriber(tokenID *big.Int) ([]byte, error) {
	calldata, err := subscriberFn.EncodeCall(tokenID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeSubscriber: %w", err)
	}
	return calldata, nil
}

// DecodeSubscriber decodes the return data of subscriber: the position's subscriber contract
// address (the zero address if unsubscribed).
func DecodeSubscriber(data []byte) (*types.Address, error) {
	subscriber, err := subscriberFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeSubscriber: %w", err)
	}
	return subscriber, nil
}

// EncodeUnsubscribeGasLimit returns the calldata for unsubscribeGasLimit().
func EncodeUnsubscribeGasLimit() ([]byte, error) {
	calldata, err := unsubscribeGasLimitFn.EncodeCall()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeUnsubscribeGasLimit: %w", err)
	}
	return calldata, nil
}

// DecodeUnsubscribeGasLimit decodes the return data of unsubscribeGasLimit.
func DecodeUnsubscribeGasLimit(data []byte) (*big.Int, error) {
	gasLimit, err := unsubscribeGasLimitFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeUnsubscribeGasLimit: %w", err)
	}
	return gasLimit, nil
}

// EncodeInitializePool returns the calldata for initializePool(key, sqrtPriceX96).
func EncodeInitializePool(key *PoolKey, sqrtPriceX96 *big.Int) ([]byte, error) {
	if key == nil {
		return nil, fmt.Errorf("uniswapv4: encodeInitializePool: key must not be nil")
	}
	calldata, err := initializePoolFn.EncodeCall([]any{key.Currency0, key.Currency1, key.Fee, key.TickSpacing, key.Hooks}, sqrtPriceX96)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeInitializePool: %w", err)
	}
	return calldata, nil
}

// DecodeInitializePool decodes the return data of initializePool: the pool's resulting tick.
func DecodeInitializePool(data []byte) (*big.Int, error) {
	tick, err := initializePoolFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeInitializePool: %w", err)
	}
	return tick, nil
}

// EncodeMulticall returns the calldata for multicall(data).
func EncodeMulticall(data [][]byte) ([]byte, error) {
	calldata, err := multicallFn.EncodeCall(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeMulticall: %w", err)
	}
	return calldata, nil
}

// DecodeMulticall decodes the return data of multicall -- each sub-call's own return data.
func DecodeMulticall(data []byte) ([][]byte, error) {
	vals, err := multicallFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeMulticall: %w", err)
	}
	if len(vals) != 1 {
		return nil, fmt.Errorf("uniswapv4: decodeMulticall: expected 1 return value, got %d", len(vals))
	}
	elems, ok := vals[0].([]any)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeMulticall: unexpected decoded type for results: %T", vals[0])
	}
	results := make([][]byte, len(elems))
	for i, e := range elems {
		b, ok := e.([]byte)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: decodeMulticall: unexpected decoded type for results[%d]: %T", i, e)
		}
		results[i] = b
	}
	return results, nil
}

// EncodePoolManager returns the calldata for poolManager().
func EncodePoolManager() ([]byte, error) {
	calldata, err := poolManagerFn.EncodeCall()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodePoolManager: %w", err)
	}
	return calldata, nil
}

// DecodePoolManager decodes the return data of poolManager: the PoolManager this PositionManager
// is wired to.
func DecodePoolManager(data []byte) (*types.Address, error) {
	addr, err := poolManagerFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodePoolManager: %w", err)
	}
	return addr, nil
}

// EncodeWETH9 returns the calldata for WETH9().
func EncodeWETH9() ([]byte, error) {
	data, err := weth9Fn.EncodeCall()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeWETH9: %w", err)
	}
	return data, nil
}

// DecodeWETH9 decodes the return data of WETH9: the WETH9 contract this PositionManager wraps
// and unwraps native ETH through.
func DecodeWETH9(data []byte) (*types.Address, error) {
	addr, err := weth9Fn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeWETH9: %w", err)
	}
	return addr, nil
}

// EncodePermit2Contract returns the calldata for permit2().
func EncodePermit2Contract() ([]byte, error) {
	data, err := permit2ContractFn.EncodeCall()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodePermit2Contract: %w", err)
	}
	return data, nil
}

// DecodePermit2Contract decodes the return data of permit2: the Permit2 contract this
// PositionManager forwards permit/permitBatch to.
func DecodePermit2Contract(data []byte) (*types.Address, error) {
	addr, err := permit2ContractFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodePermit2Contract: %w", err)
	}
	return addr, nil
}

// EncodeNonces returns the calldata for nonces(owner, word).
func EncodeNonces(owner *types.Address, word *big.Int) ([]byte, error) {
	calldata, err := noncesFn.EncodeCall(owner, word)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeNonces: %w", err)
	}
	return calldata, nil
}

// DecodeNonces decodes the return data of nonces: the bitmap word of consumed nonces.
func DecodeNonces(data []byte) (*big.Int, error) {
	bitmap, err := noncesFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeNonces: %w", err)
	}
	return bitmap, nil
}

// EncodeRevokeNonce returns the calldata for revokeNonce(nonce).
func EncodeRevokeNonce(nonce *big.Int) ([]byte, error) {
	calldata, err := revokeNonceFn.EncodeCall(nonce)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeRevokeNonce: %w", err)
	}
	return calldata, nil
}

// EncodeERC721Permit returns the calldata for permit(spender, tokenId, deadline, nonce,
// signature): approves spender for a single position NFT via signature.
func EncodeERC721Permit(spender *types.Address, tokenID, deadline, nonce *big.Int, signature []byte) ([]byte, error) {
	calldata, err := erc721PermitFn.EncodeCall(spender, tokenID, deadline, nonce, signature)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeERC721Permit: %w", err)
	}
	return calldata, nil
}

// EncodeERC721PermitForAll returns the calldata for permitForAll(owner, operator, approved,
// deadline, nonce, signature): sets an operator over all of owner's position NFTs via signature.
func EncodeERC721PermitForAll(owner, operator *types.Address, approved bool, deadline, nonce *big.Int, signature []byte) ([]byte, error) {
	calldata, err := erc721PermitForAllFn.EncodeCall(owner, operator, approved, deadline, nonce, signature)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeERC721PermitForAll: %w", err)
	}
	return calldata, nil
}

// EncodePermit2Permit returns the calldata for Permit2Forwarder's permit(owner, permitSingle,
// signature): forwards a single-token permit2 approval on owner's behalf.
func EncodePermit2Permit(owner *types.Address, permitSingle *permit2.PermitSingle, signature []byte) ([]byte, error) {
	if permitSingle == nil {
		return nil, fmt.Errorf("uniswapv4: encodePermit2Permit: permitSingle must not be nil")
	}
	if permitSingle.Details == nil {
		return nil, fmt.Errorf("uniswapv4: encodePermit2Permit: permitSingle.Details must not be nil")
	}
	calldata, err := permit2PermitFn.EncodeCall(owner, []any{
		[]any{permitSingle.Details.Token, permitSingle.Details.Amount, permitSingle.Details.Expiration, permitSingle.Details.Nonce},
		permitSingle.Spender, permitSingle.SigDeadline,
	}, signature)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodePermit2Permit: %w", err)
	}
	return calldata, nil
}

// DecodePermit2Permit decodes the return data of Permit2Forwarder's permit: the revert reason
// from a reverting inner permit2 call, empty if it succeeded.
func DecodePermit2Permit(data []byte) ([]byte, error) {
	errData, err := permit2PermitFn.DecodeSingleReturn[[]byte](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodePermit2Permit: %w", err)
	}
	return errData, nil
}

// EncodePermit2PermitBatch returns the calldata for Permit2Forwarder's permitBatch(owner,
// permitBatch, signature): forwards a batch permit2 approval on owner's behalf.
func EncodePermit2PermitBatch(owner *types.Address, permitBatch *permit2.PermitBatch, signature []byte) ([]byte, error) {
	if permitBatch == nil {
		return nil, fmt.Errorf("uniswapv4: encodePermit2PermitBatch: permitBatch must not be nil")
	}
	details := make([]any, len(permitBatch.Details))
	for i, d := range permitBatch.Details {
		details[i] = []any{d.Token, d.Amount, d.Expiration, d.Nonce}
	}
	calldata, err := permit2PermitBatchFn.EncodeCall(owner, []any{details, permitBatch.Spender, permitBatch.SigDeadline}, signature)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodePermit2PermitBatch: %w", err)
	}
	return calldata, nil
}

// DecodePermit2PermitBatch decodes the return data of Permit2Forwarder's permitBatch: the revert
// reason from a reverting inner permit2 call, empty if it succeeded.
func DecodePermit2PermitBatch(data []byte) ([]byte, error) {
	errData, err := permit2PermitBatchFn.DecodeSingleReturn[[]byte](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodePermit2PermitBatch: %w", err)
	}
	return errData, nil
}

// Events from IPositionManager, for decoding logs via Event.Decode.
var (
	modifyPositionEvent = abi.NewEvent("ModifyPosition", []abi.EventParam{
		{Type: abi.Bytes32, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Int24},
		{Type: abi.Int24},
		{Type: abi.Int256},
		{Type: abi.Bytes32},
	}, false)
)

// Events from INotifier, for decoding logs via Event.Decode.
var (
	subscriptionEvent = abi.NewEvent("Subscription", []abi.EventParam{
		{Type: abi.Uint256, Indexed: true},
		{Type: abi.Address, Indexed: true},
	}, false)
	unsubscriptionEvent = abi.NewEvent("Unsubscription", []abi.EventParam{
		{Type: abi.Uint256, Indexed: true},
		{Type: abi.Address, Indexed: true},
	}, false)
)

// mintPositionTypes mirrors CalldataDecoder.decodeMintParams' calldata layout: poolKey, tickLower,
// tickUpper, liquidity, amount0Max, amount1Max, and owner as flat top-level words (poolKey
// inlined, since it's static), followed by hookData's offset and tail -- NOT a single struct
// argument with its own leading offset (that's the shape Router's swap params use instead).
var mintPositionTypes = abi.NewTypes(poolKeyTuple, abi.Int24, abi.Int24, abi.Uint256, abi.Uint128, abi.Uint128, abi.Address, abi.Bytes)

// EncodeActionMintPosition returns the params blob for Actions.MINT_POSITION.
func EncodeActionMintPosition(p *MintPositionParams) ([]byte, error) {
	if p == nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionMintPosition: params must not be nil")
	}
	if p.PoolKey == nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionMintPosition: poolKey must not be nil")
	}
	data, err := abi.Pack(mintPositionTypes,
		[]any{p.PoolKey.Currency0, p.PoolKey.Currency1, p.PoolKey.Fee, p.PoolKey.TickSpacing, p.PoolKey.Hooks},
		p.TickLower, p.TickUpper, p.Liquidity, p.Amount0Max, p.Amount1Max, p.Owner, p.HookData,
	)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionMintPosition: %w", err)
	}
	return data, nil
}

// DecodeActionMintPosition decodes a Actions.MINT_POSITION params blob.
func DecodeActionMintPosition(data []byte) (*MintPositionParams, error) {
	vals, err := abi.Unpack(mintPositionTypes, data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPosition: %w", err)
	}
	poolKeyFields, ok := vals[0].([]any)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPosition: unexpected decoded type for poolKey: %T", vals[0])
	}
	currency0, ok := poolKeyFields[0].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPosition: unexpected decoded type for currency0: %T", poolKeyFields[0])
	}
	currency1, ok := poolKeyFields[1].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPosition: unexpected decoded type for currency1: %T", poolKeyFields[1])
	}
	fee, ok := poolKeyFields[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPosition: unexpected decoded type for fee: %T", poolKeyFields[2])
	}
	tickSpacing, ok := poolKeyFields[3].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPosition: unexpected decoded type for tickSpacing: %T", poolKeyFields[3])
	}
	hooks, ok := poolKeyFields[4].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPosition: unexpected decoded type for hooks: %T", poolKeyFields[4])
	}
	tickLower, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPosition: unexpected decoded type for tickLower: %T", vals[1])
	}
	tickUpper, ok := vals[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPosition: unexpected decoded type for tickUpper: %T", vals[2])
	}
	liquidity, ok := vals[3].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPosition: unexpected decoded type for liquidity: %T", vals[3])
	}
	amount0Max, ok := vals[4].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPosition: unexpected decoded type for amount0Max: %T", vals[4])
	}
	amount1Max, ok := vals[5].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPosition: unexpected decoded type for amount1Max: %T", vals[5])
	}
	owner, ok := vals[6].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPosition: unexpected decoded type for owner: %T", vals[6])
	}
	hookData, ok := vals[7].([]byte)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPosition: unexpected decoded type for hookData: %T", vals[7])
	}
	return NewMintPositionParams(
		NewPoolKey(currency0, currency1, fee, tickSpacing, hooks),
		tickLower, tickUpper, liquidity, amount0Max, amount1Max, owner, hookData,
	), nil
}

// mintPositionFromDeltasTypes mirrors CalldataDecoder.decodeMintFromDeltasParams' calldata
// layout: flat top-level words like mintPositionTypes, minus liquidity.
var mintPositionFromDeltasTypes = abi.NewTypes(poolKeyTuple, abi.Int24, abi.Int24, abi.Uint128, abi.Uint128, abi.Address, abi.Bytes)

// EncodeActionMintPositionFromDeltas returns the params blob for Actions.MINT_POSITION_FROM_DELTAS.
// Deprecated upstream (sandwich-attack vulnerable); prefer EncodeActionMintPosition.
func EncodeActionMintPositionFromDeltas(p *MintPositionFromDeltasParams) ([]byte, error) {
	if p == nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionMintPositionFromDeltas: params must not be nil")
	}
	if p.PoolKey == nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionMintPositionFromDeltas: poolKey must not be nil")
	}
	data, err := abi.Pack(mintPositionFromDeltasTypes,
		[]any{p.PoolKey.Currency0, p.PoolKey.Currency1, p.PoolKey.Fee, p.PoolKey.TickSpacing, p.PoolKey.Hooks},
		p.TickLower, p.TickUpper, p.Amount0Max, p.Amount1Max, p.Owner, p.HookData,
	)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionMintPositionFromDeltas: %w", err)
	}
	return data, nil
}

// DecodeActionMintPositionFromDeltas decodes a Actions.MINT_POSITION_FROM_DELTAS params blob.
func DecodeActionMintPositionFromDeltas(data []byte) (*MintPositionFromDeltasParams, error) {
	vals, err := abi.Unpack(mintPositionFromDeltasTypes, data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPositionFromDeltas: %w", err)
	}
	poolKeyFields, ok := vals[0].([]any)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPositionFromDeltas: unexpected decoded type for poolKey: %T", vals[0])
	}
	currency0, ok := poolKeyFields[0].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPositionFromDeltas: unexpected decoded type for currency0: %T", poolKeyFields[0])
	}
	currency1, ok := poolKeyFields[1].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPositionFromDeltas: unexpected decoded type for currency1: %T", poolKeyFields[1])
	}
	fee, ok := poolKeyFields[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPositionFromDeltas: unexpected decoded type for fee: %T", poolKeyFields[2])
	}
	tickSpacing, ok := poolKeyFields[3].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPositionFromDeltas: unexpected decoded type for tickSpacing: %T", poolKeyFields[3])
	}
	hooks, ok := poolKeyFields[4].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPositionFromDeltas: unexpected decoded type for hooks: %T", poolKeyFields[4])
	}
	tickLower, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPositionFromDeltas: unexpected decoded type for tickLower: %T", vals[1])
	}
	tickUpper, ok := vals[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPositionFromDeltas: unexpected decoded type for tickUpper: %T", vals[2])
	}
	amount0Max, ok := vals[3].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPositionFromDeltas: unexpected decoded type for amount0Max: %T", vals[3])
	}
	amount1Max, ok := vals[4].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPositionFromDeltas: unexpected decoded type for amount1Max: %T", vals[4])
	}
	owner, ok := vals[5].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPositionFromDeltas: unexpected decoded type for owner: %T", vals[5])
	}
	hookData, ok := vals[6].([]byte)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionMintPositionFromDeltas: unexpected decoded type for hookData: %T", vals[6])
	}
	return NewMintPositionFromDeltasParams(
		NewPoolKey(currency0, currency1, fee, tickSpacing, hooks),
		tickLower, tickUpper, amount0Max, amount1Max, owner, hookData,
	), nil
}

// EncodeActionIncreaseLiquidity returns the params blob for Actions.INCREASE_LIQUIDITY:
// abi.encode(tokenId, liquidity, amount0Max, amount1Max, hookData).
func EncodeActionIncreaseLiquidity(tokenID, liquidity, amount0Max, amount1Max *big.Int, hookData []byte) ([]byte, error) {
	data, err := abi.Pack(abi.NewTypes(abi.Uint256, abi.Uint256, abi.Uint128, abi.Uint128, abi.Bytes), tokenID, liquidity, amount0Max, amount1Max, hookData)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionIncreaseLiquidity: %w", err)
	}
	return data, nil
}

// DecodeActionIncreaseLiquidity decodes a Actions.INCREASE_LIQUIDITY params blob.
func DecodeActionIncreaseLiquidity(data []byte) (*big.Int, *big.Int, *big.Int, *big.Int, []byte, error) {
	vals, err := abi.Unpack(abi.NewTypes(abi.Uint256, abi.Uint256, abi.Uint128, abi.Uint128, abi.Bytes), data)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionIncreaseLiquidity: %w", err)
	}
	tokenID, ok := vals[0].(*big.Int)
	if !ok {
		return nil, nil, nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionIncreaseLiquidity: unexpected decoded type for tokenId: %T", vals[0])
	}
	liquidity, ok := vals[1].(*big.Int)
	if !ok {
		return nil, nil, nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionIncreaseLiquidity: unexpected decoded type for liquidity: %T", vals[1])
	}
	amount0Max, ok := vals[2].(*big.Int)
	if !ok {
		return nil, nil, nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionIncreaseLiquidity: unexpected decoded type for amount0Max: %T", vals[2])
	}
	amount1Max, ok := vals[3].(*big.Int)
	if !ok {
		return nil, nil, nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionIncreaseLiquidity: unexpected decoded type for amount1Max: %T", vals[3])
	}
	hookData, ok := vals[4].([]byte)
	if !ok {
		return nil, nil, nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionIncreaseLiquidity: unexpected decoded type for hookData: %T", vals[4])
	}
	return tokenID, liquidity, amount0Max, amount1Max, hookData, nil
}

// EncodeActionIncreaseLiquidityFromDeltas returns the params blob for
// Actions.INCREASE_LIQUIDITY_FROM_DELTAS: abi.encode(tokenId, amount0Max, amount1Max, hookData).
// Deprecated upstream (sandwich-attack vulnerable); prefer EncodeActionIncreaseLiquidity.
func EncodeActionIncreaseLiquidityFromDeltas(tokenID, amount0Max, amount1Max *big.Int, hookData []byte) ([]byte, error) {
	data, err := abi.Pack(abi.NewTypes(abi.Uint256, abi.Uint128, abi.Uint128, abi.Bytes), tokenID, amount0Max, amount1Max, hookData)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionIncreaseLiquidityFromDeltas: %w", err)
	}
	return data, nil
}

// DecodeActionIncreaseLiquidityFromDeltas decodes a Actions.INCREASE_LIQUIDITY_FROM_DELTAS params blob.
func DecodeActionIncreaseLiquidityFromDeltas(data []byte) (*big.Int, *big.Int, *big.Int, []byte, error) {
	vals, err := abi.Unpack(abi.NewTypes(abi.Uint256, abi.Uint128, abi.Uint128, abi.Bytes), data)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionIncreaseLiquidityFromDeltas: %w", err)
	}
	tokenID, ok := vals[0].(*big.Int)
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionIncreaseLiquidityFromDeltas: unexpected decoded type for tokenId: %T", vals[0])
	}
	amount0Max, ok := vals[1].(*big.Int)
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionIncreaseLiquidityFromDeltas: unexpected decoded type for amount0Max: %T", vals[1])
	}
	amount1Max, ok := vals[2].(*big.Int)
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionIncreaseLiquidityFromDeltas: unexpected decoded type for amount1Max: %T", vals[2])
	}
	hookData, ok := vals[3].([]byte)
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionIncreaseLiquidityFromDeltas: unexpected decoded type for hookData: %T", vals[3])
	}
	return tokenID, amount0Max, amount1Max, hookData, nil
}

// EncodeActionDecreaseLiquidity returns the params blob for Actions.DECREASE_LIQUIDITY:
// abi.encode(tokenId, liquidity, amount0Min, amount1Min, hookData).
func EncodeActionDecreaseLiquidity(tokenID, liquidity, amount0Min, amount1Min *big.Int, hookData []byte) ([]byte, error) {
	data, err := abi.Pack(abi.NewTypes(abi.Uint256, abi.Uint256, abi.Uint128, abi.Uint128, abi.Bytes), tokenID, liquidity, amount0Min, amount1Min, hookData)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionDecreaseLiquidity: %w", err)
	}
	return data, nil
}

// DecodeActionDecreaseLiquidity decodes a Actions.DECREASE_LIQUIDITY params blob.
func DecodeActionDecreaseLiquidity(data []byte) (*big.Int, *big.Int, *big.Int, *big.Int, []byte, error) {
	vals, err := abi.Unpack(abi.NewTypes(abi.Uint256, abi.Uint256, abi.Uint128, abi.Uint128, abi.Bytes), data)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionDecreaseLiquidity: %w", err)
	}
	tokenID, ok := vals[0].(*big.Int)
	if !ok {
		return nil, nil, nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionDecreaseLiquidity: unexpected decoded type for tokenId: %T", vals[0])
	}
	liquidity, ok := vals[1].(*big.Int)
	if !ok {
		return nil, nil, nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionDecreaseLiquidity: unexpected decoded type for liquidity: %T", vals[1])
	}
	amount0Min, ok := vals[2].(*big.Int)
	if !ok {
		return nil, nil, nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionDecreaseLiquidity: unexpected decoded type for amount0Min: %T", vals[2])
	}
	amount1Min, ok := vals[3].(*big.Int)
	if !ok {
		return nil, nil, nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionDecreaseLiquidity: unexpected decoded type for amount1Min: %T", vals[3])
	}
	hookData, ok := vals[4].([]byte)
	if !ok {
		return nil, nil, nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionDecreaseLiquidity: unexpected decoded type for hookData: %T", vals[4])
	}
	return tokenID, liquidity, amount0Min, amount1Min, hookData, nil
}

// EncodeActionBurnPosition returns the params blob for Actions.BURN_POSITION:
// abi.encode(tokenId, amount0Min, amount1Min, hookData).
func EncodeActionBurnPosition(tokenID, amount0Min, amount1Min *big.Int, hookData []byte) ([]byte, error) {
	data, err := abi.Pack(abi.NewTypes(abi.Uint256, abi.Uint128, abi.Uint128, abi.Bytes), tokenID, amount0Min, amount1Min, hookData)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionBurnPosition: %w", err)
	}
	return data, nil
}

// DecodeActionBurnPosition decodes a Actions.BURN_POSITION params blob.
func DecodeActionBurnPosition(data []byte) (*big.Int, *big.Int, *big.Int, []byte, error) {
	vals, err := abi.Unpack(abi.NewTypes(abi.Uint256, abi.Uint128, abi.Uint128, abi.Bytes), data)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionBurnPosition: %w", err)
	}
	tokenID, ok := vals[0].(*big.Int)
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionBurnPosition: unexpected decoded type for tokenId: %T", vals[0])
	}
	amount0Min, ok := vals[1].(*big.Int)
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionBurnPosition: unexpected decoded type for amount0Min: %T", vals[1])
	}
	amount1Min, ok := vals[2].(*big.Int)
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionBurnPosition: unexpected decoded type for amount1Min: %T", vals[2])
	}
	hookData, ok := vals[3].([]byte)
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionBurnPosition: unexpected decoded type for hookData: %T", vals[3])
	}
	return tokenID, amount0Min, amount1Min, hookData, nil
}

// EncodeActionSettlePair returns the params blob for Actions.SETTLE_PAIR: abi.encode(currency0, currency1).
func EncodeActionSettlePair(currency0, currency1 Currency) ([]byte, error) {
	data, err := abi.Pack(abi.NewTypes(abi.Address, abi.Address), currency0, currency1)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionSettlePair: %w", err)
	}
	return data, nil
}

// DecodeActionSettlePair decodes a Actions.SETTLE_PAIR params blob.
func DecodeActionSettlePair(data []byte) (Currency, Currency, error) {
	vals, err := abi.Unpack(abi.NewTypes(abi.Address, abi.Address), data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: decodeActionSettlePair: %w", err)
	}
	currency0, ok := vals[0].(*types.Address)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv4: decodeActionSettlePair: unexpected decoded type for currency0: %T", vals[0])
	}
	currency1, ok := vals[1].(*types.Address)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv4: decodeActionSettlePair: unexpected decoded type for currency1: %T", vals[1])
	}
	return currency0, currency1, nil
}

// EncodeActionTakePair returns the params blob for Actions.TAKE_PAIR: abi.encode(currency0, currency1, recipient).
func EncodeActionTakePair(currency0, currency1, recipient Currency) ([]byte, error) {
	data, err := abi.Pack(abi.NewTypes(abi.Address, abi.Address, abi.Address), currency0, currency1, recipient)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionTakePair: %w", err)
	}
	return data, nil
}

// DecodeActionTakePair decodes a Actions.TAKE_PAIR params blob.
func DecodeActionTakePair(data []byte) (Currency, Currency, Currency, error) {
	vals, err := abi.Unpack(abi.NewTypes(abi.Address, abi.Address, abi.Address), data)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionTakePair: %w", err)
	}
	currency0, ok := vals[0].(*types.Address)
	if !ok {
		return nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionTakePair: unexpected decoded type for currency0: %T", vals[0])
	}
	currency1, ok := vals[1].(*types.Address)
	if !ok {
		return nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionTakePair: unexpected decoded type for currency1: %T", vals[1])
	}
	recipient, ok := vals[2].(*types.Address)
	if !ok {
		return nil, nil, nil, fmt.Errorf("uniswapv4: decodeActionTakePair: unexpected decoded type for recipient: %T", vals[2])
	}
	return currency0, currency1, recipient, nil
}

// EncodeActionCloseCurrency returns the params blob for Actions.CLOSE_CURRENCY: abi.encode(currency).
func EncodeActionCloseCurrency(currency Currency) ([]byte, error) {
	data, err := abi.Pack(abi.NewTypes(abi.Address), currency)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionCloseCurrency: %w", err)
	}
	return data, nil
}

// DecodeActionCloseCurrency decodes a Actions.CLOSE_CURRENCY params blob.
func DecodeActionCloseCurrency(data []byte) (Currency, error) {
	vals, err := abi.Unpack(abi.NewTypes(abi.Address), data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeActionCloseCurrency: %w", err)
	}
	currency, ok := vals[0].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionCloseCurrency: unexpected decoded type for currency: %T", vals[0])
	}
	return currency, nil
}

// EncodeActionClearOrTake returns the params blob for Actions.CLEAR_OR_TAKE: abi.encode(currency, amountMax).
func EncodeActionClearOrTake(currency Currency, amountMax *big.Int) ([]byte, error) {
	data, err := abi.Pack(abi.NewTypes(abi.Address, abi.Uint256), currency, amountMax)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionClearOrTake: %w", err)
	}
	return data, nil
}

// DecodeActionClearOrTake decodes a Actions.CLEAR_OR_TAKE params blob.
func DecodeActionClearOrTake(data []byte) (Currency, *big.Int, error) {
	vals, err := abi.Unpack(abi.NewTypes(abi.Address, abi.Uint256), data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: decodeActionClearOrTake: %w", err)
	}
	currency, ok := vals[0].(*types.Address)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv4: decodeActionClearOrTake: unexpected decoded type for currency: %T", vals[0])
	}
	amountMax, ok := vals[1].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv4: decodeActionClearOrTake: unexpected decoded type for amountMax: %T", vals[1])
	}
	return currency, amountMax, nil
}

// EncodeActionSweep returns the params blob for Actions.SWEEP: abi.encode(currency, to).
func EncodeActionSweep(currency Currency, to *types.Address) ([]byte, error) {
	data, err := abi.Pack(abi.NewTypes(abi.Address, abi.Address), currency, to)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionSweep: %w", err)
	}
	return data, nil
}

// DecodeActionSweep decodes a Actions.SWEEP params blob.
func DecodeActionSweep(data []byte) (Currency, *types.Address, error) {
	vals, err := abi.Unpack(abi.NewTypes(abi.Address, abi.Address), data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: decodeActionSweep: %w", err)
	}
	currency, ok := vals[0].(*types.Address)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv4: decodeActionSweep: unexpected decoded type for currency: %T", vals[0])
	}
	to, ok := vals[1].(*types.Address)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv4: decodeActionSweep: unexpected decoded type for to: %T", vals[1])
	}
	return currency, to, nil
}

// EncodeActionWrap returns the params blob for Actions.WRAP: abi.encode(amount).
func EncodeActionWrap(amount *big.Int) ([]byte, error) {
	data, err := abi.Pack(abi.NewTypes(abi.Uint256), amount)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionWrap: %w", err)
	}
	return data, nil
}

// DecodeActionWrap decodes a Actions.WRAP params blob.
func DecodeActionWrap(data []byte) (*big.Int, error) {
	vals, err := abi.Unpack(abi.NewTypes(abi.Uint256), data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeActionWrap: %w", err)
	}
	amount, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionWrap: unexpected decoded type for amount: %T", vals[0])
	}
	return amount, nil
}

// EncodeActionUnwrap returns the params blob for Actions.UNWRAP: abi.encode(amount).
func EncodeActionUnwrap(amount *big.Int) ([]byte, error) {
	data, err := abi.Pack(abi.NewTypes(abi.Uint256), amount)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: encodeActionUnwrap: %w", err)
	}
	return data, nil
}

// DecodeActionUnwrap decodes a Actions.UNWRAP params blob.
func DecodeActionUnwrap(data []byte) (*big.Int, error) {
	vals, err := abi.Unpack(abi.NewTypes(abi.Uint256), data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: decodeActionUnwrap: %w", err)
	}
	amount, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv4: decodeActionUnwrap: unexpected decoded type for amount: %T", vals[0])
	}
	return amount, nil
}
