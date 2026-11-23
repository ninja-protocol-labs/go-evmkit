package v4

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/protocols/permit2"
)

func TestEncodeDecodeActionMintPosition(t *testing.T) {
	poolKey := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	owner := types.MustNewAddressFromHex("0x9a8ca723f5dccb7926d00b71dec55c2fea1f50f7")
	params := NewMintPositionParams(poolKey, big.NewInt(-887220), big.NewInt(887220), big.NewInt(1000000), big.NewInt(1000), big.NewInt(1100), owner, []byte{0xde, 0xad})

	data, err := EncodeActionMintPosition(params)
	require.NoError(t, err)

	decoded, err := DecodeActionMintPosition(data)
	require.NoError(t, err)
	require.True(t, decoded.PoolKey.Currency0.Equal(testCurrency0))
	require.True(t, decoded.PoolKey.Currency1.Equal(testCurrency1))
	require.Zero(t, big.NewInt(3000).Cmp(decoded.PoolKey.Fee))
	require.Zero(t, big.NewInt(60).Cmp(decoded.PoolKey.TickSpacing))
	require.True(t, decoded.PoolKey.Hooks.Equal(testHooks))
	require.Zero(t, big.NewInt(-887220).Cmp(decoded.TickLower))
	require.Zero(t, big.NewInt(887220).Cmp(decoded.TickUpper))
	require.Zero(t, big.NewInt(1000000).Cmp(decoded.Liquidity))
	require.Zero(t, big.NewInt(1000).Cmp(decoded.Amount0Max))
	require.Zero(t, big.NewInt(1100).Cmp(decoded.Amount1Max))
	require.True(t, decoded.Owner.Equal(owner))
	require.Equal(t, []byte{0xde, 0xad}, decoded.HookData)
}

func TestEncodeActionMintPositionRejectsNilArgs(t *testing.T) {
	_, err := EncodeActionMintPosition(nil)
	require.Error(t, err)

	_, err = EncodeActionMintPosition(NewMintPositionParams(nil, big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), testSender, nil))
	require.Error(t, err)
}

func TestDecodeActionMintPositionRejectsGarbage(t *testing.T) {
	_, err := DecodeActionMintPosition([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeActionMintPositionFromDeltas(t *testing.T) {
	poolKey := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	owner := types.MustNewAddressFromHex("0x9a8ca723f5dccb7926d00b71dec55c2fea1f50f7")
	params := NewMintPositionFromDeltasParams(poolKey, big.NewInt(-887220), big.NewInt(887220), big.NewInt(1000), big.NewInt(1100), owner, []byte{0xbe, 0xef})

	data, err := EncodeActionMintPositionFromDeltas(params)
	require.NoError(t, err)

	decoded, err := DecodeActionMintPositionFromDeltas(data)
	require.NoError(t, err)
	require.True(t, decoded.PoolKey.Currency0.Equal(testCurrency0))
	require.True(t, decoded.PoolKey.Currency1.Equal(testCurrency1))
	require.Zero(t, big.NewInt(3000).Cmp(decoded.PoolKey.Fee))
	require.Zero(t, big.NewInt(60).Cmp(decoded.PoolKey.TickSpacing))
	require.True(t, decoded.PoolKey.Hooks.Equal(testHooks))
	require.Zero(t, big.NewInt(-887220).Cmp(decoded.TickLower))
	require.Zero(t, big.NewInt(887220).Cmp(decoded.TickUpper))
	require.Zero(t, big.NewInt(1000).Cmp(decoded.Amount0Max))
	require.Zero(t, big.NewInt(1100).Cmp(decoded.Amount1Max))
	require.True(t, decoded.Owner.Equal(owner))
	require.Equal(t, []byte{0xbe, 0xef}, decoded.HookData)
}

func TestEncodeActionMintPositionFromDeltasRejectsNilArgs(t *testing.T) {
	_, err := EncodeActionMintPositionFromDeltas(nil)
	require.Error(t, err)

	_, err = EncodeActionMintPositionFromDeltas(NewMintPositionFromDeltasParams(nil, big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), testSender, nil))
	require.Error(t, err)
}

func TestDecodeActionMintPositionFromDeltasRejectsGarbage(t *testing.T) {
	_, err := DecodeActionMintPositionFromDeltas([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeActionIncreaseLiquidity(t *testing.T) {
	data, err := EncodeActionIncreaseLiquidity(big.NewInt(42), big.NewInt(1000000), big.NewInt(1000), big.NewInt(1100), []byte{0x01})
	require.NoError(t, err)

	tokenID, liquidity, amount0Max, amount1Max, hookData, err := DecodeActionIncreaseLiquidity(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(42).Cmp(tokenID))
	require.Zero(t, big.NewInt(1000000).Cmp(liquidity))
	require.Zero(t, big.NewInt(1000).Cmp(amount0Max))
	require.Zero(t, big.NewInt(1100).Cmp(amount1Max))
	require.Equal(t, []byte{0x01}, hookData)
}

func TestDecodeActionIncreaseLiquidityRejectsGarbage(t *testing.T) {
	_, _, _, _, _, err := DecodeActionIncreaseLiquidity([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeActionIncreaseLiquidityFromDeltas(t *testing.T) {
	data, err := EncodeActionIncreaseLiquidityFromDeltas(big.NewInt(42), big.NewInt(1000), big.NewInt(1100), []byte{0x02})
	require.NoError(t, err)

	tokenID, amount0Max, amount1Max, hookData, err := DecodeActionIncreaseLiquidityFromDeltas(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(42).Cmp(tokenID))
	require.Zero(t, big.NewInt(1000).Cmp(amount0Max))
	require.Zero(t, big.NewInt(1100).Cmp(amount1Max))
	require.Equal(t, []byte{0x02}, hookData)
}

func TestDecodeActionIncreaseLiquidityFromDeltasRejectsGarbage(t *testing.T) {
	_, _, _, _, err := DecodeActionIncreaseLiquidityFromDeltas([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeActionDecreaseLiquidity(t *testing.T) {
	data, err := EncodeActionDecreaseLiquidity(big.NewInt(42), big.NewInt(500000), big.NewInt(900), big.NewInt(950), []byte{0x03})
	require.NoError(t, err)

	tokenID, liquidity, amount0Min, amount1Min, hookData, err := DecodeActionDecreaseLiquidity(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(42).Cmp(tokenID))
	require.Zero(t, big.NewInt(500000).Cmp(liquidity))
	require.Zero(t, big.NewInt(900).Cmp(amount0Min))
	require.Zero(t, big.NewInt(950).Cmp(amount1Min))
	require.Equal(t, []byte{0x03}, hookData)
}

func TestDecodeActionDecreaseLiquidityRejectsGarbage(t *testing.T) {
	_, _, _, _, _, err := DecodeActionDecreaseLiquidity([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeActionBurnPosition(t *testing.T) {
	data, err := EncodeActionBurnPosition(big.NewInt(42), big.NewInt(900), big.NewInt(950), []byte{0x04})
	require.NoError(t, err)

	tokenID, amount0Min, amount1Min, hookData, err := DecodeActionBurnPosition(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(42).Cmp(tokenID))
	require.Zero(t, big.NewInt(900).Cmp(amount0Min))
	require.Zero(t, big.NewInt(950).Cmp(amount1Min))
	require.Equal(t, []byte{0x04}, hookData)
}

func TestDecodeActionBurnPositionRejectsGarbage(t *testing.T) {
	_, _, _, _, err := DecodeActionBurnPosition([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeActionSettlePair(t *testing.T) {
	data, err := EncodeActionSettlePair(testCurrency0, testCurrency1)
	require.NoError(t, err)

	currency0, currency1, err := DecodeActionSettlePair(data)
	require.NoError(t, err)
	require.True(t, currency0.Equal(testCurrency0))
	require.True(t, currency1.Equal(testCurrency1))
}

func TestDecodeActionSettlePairRejectsGarbage(t *testing.T) {
	_, _, err := DecodeActionSettlePair([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeActionTakePair(t *testing.T) {
	recipient := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	data, err := EncodeActionTakePair(testCurrency0, testCurrency1, recipient)
	require.NoError(t, err)

	currency0, currency1, decodedRecipient, err := DecodeActionTakePair(data)
	require.NoError(t, err)
	require.True(t, currency0.Equal(testCurrency0))
	require.True(t, currency1.Equal(testCurrency1))
	require.True(t, decodedRecipient.Equal(recipient))
}

func TestDecodeActionTakePairRejectsGarbage(t *testing.T) {
	_, _, _, err := DecodeActionTakePair([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeActionCloseCurrency(t *testing.T) {
	data, err := EncodeActionCloseCurrency(testCurrency0)
	require.NoError(t, err)

	currency, err := DecodeActionCloseCurrency(data)
	require.NoError(t, err)
	require.True(t, currency.Equal(testCurrency0))
}

func TestDecodeActionCloseCurrencyRejectsGarbage(t *testing.T) {
	_, err := DecodeActionCloseCurrency([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeActionClearOrTake(t *testing.T) {
	data, err := EncodeActionClearOrTake(testCurrency0, big.NewInt(1000))
	require.NoError(t, err)

	currency, amountMax, err := DecodeActionClearOrTake(data)
	require.NoError(t, err)
	require.True(t, currency.Equal(testCurrency0))
	require.Zero(t, big.NewInt(1000).Cmp(amountMax))
}

func TestDecodeActionClearOrTakeRejectsGarbage(t *testing.T) {
	_, _, err := DecodeActionClearOrTake([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeActionSweep(t *testing.T) {
	to := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	data, err := EncodeActionSweep(testCurrency0, to)
	require.NoError(t, err)

	currency, decodedTo, err := DecodeActionSweep(data)
	require.NoError(t, err)
	require.True(t, currency.Equal(testCurrency0))
	require.True(t, decodedTo.Equal(to))
}

func TestDecodeActionSweepRejectsGarbage(t *testing.T) {
	_, _, err := DecodeActionSweep([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeActionWrap(t *testing.T) {
	data, err := EncodeActionWrap(big.NewInt(1000))
	require.NoError(t, err)

	amount, err := DecodeActionWrap(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1000).Cmp(amount))
}

func TestDecodeActionWrapRejectsGarbage(t *testing.T) {
	_, err := DecodeActionWrap([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeActionUnwrap(t *testing.T) {
	data, err := EncodeActionUnwrap(big.NewInt(1000))
	require.NoError(t, err)

	amount, err := DecodeActionUnwrap(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1000).Cmp(amount))
}

func TestDecodeActionUnwrapRejectsGarbage(t *testing.T) {
	_, err := DecodeActionUnwrap([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeModifyLiquidities(t *testing.T) {
	actions := []byte{ActionIncreaseLiquidity, ActionSettlePair}
	params := [][]byte{{0x01, 0x02}, {0x03}}
	deadline := big.NewInt(1234567890)

	data, err := EncodeModifyLiquidities(actions, params, deadline)
	require.NoError(t, err)

	vals, err := modifyLiquiditiesFn.DecodeCall(data)
	require.NoError(t, err)
	unlockData, ok := vals[0].([]byte)
	require.True(t, ok)
	decodedDeadline, ok := vals[1].(*big.Int)
	require.True(t, ok)
	require.Zero(t, deadline.Cmp(decodedDeadline))

	decodedActions, decodedParams, err := DecodeActions(unlockData)
	require.NoError(t, err)
	require.Equal(t, actions, decodedActions)
	require.Equal(t, params, decodedParams)
}

func TestEncodeDecodeNextTokenID(t *testing.T) {
	data, err := EncodeNextTokenID()
	require.NoError(t, err)
	vals, err := nextTokenIdFn.DecodeCall(data)
	require.NoError(t, err)
	require.Empty(t, vals)

	returnData, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(42))
	require.NoError(t, err)

	tokenID, err := DecodeNextTokenID(returnData)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(42).Cmp(tokenID))
}

func TestDecodeNextTokenIDRejectsGarbage(t *testing.T) {
	_, err := DecodeNextTokenID([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeGetPositionLiquidity(t *testing.T) {
	data, err := EncodeGetPositionLiquidity(big.NewInt(42))
	require.NoError(t, err)
	vals, err := getPositionLiquidityFn.DecodeCall(data)
	require.NoError(t, err)
	tokenID, ok := vals[0].(*big.Int)
	require.True(t, ok)
	require.Zero(t, big.NewInt(42).Cmp(tokenID))

	returnData, err := abi.Pack(abi.Types{abi.Uint128}, big.NewInt(1000000))
	require.NoError(t, err)

	liquidity, err := DecodeGetPositionLiquidity(returnData)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1000000).Cmp(liquidity))
}

func TestDecodeGetPositionLiquidityRejectsGarbage(t *testing.T) {
	_, err := DecodeGetPositionLiquidity([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeGetPoolAndPositionInfo(t *testing.T) {
	data, err := EncodeGetPoolAndPositionInfo(big.NewInt(42))
	require.NoError(t, err)
	vals, err := getPoolAndPositionInfoFn.DecodeCall(data)
	require.NoError(t, err)
	tokenID, ok := vals[0].(*big.Int)
	require.True(t, ok)
	require.Zero(t, big.NewInt(42).Cmp(tokenID))

	word := make([]byte, 32)
	copy(word[0:25], []byte{0xaa, 0xbb, 0xcc})
	word[25], word[26], word[27] = 0x00, 0x03, 0x84 // tickUpper = 900
	word[28], word[29], word[30] = 0xff, 0xfc, 0x6c // tickLower = -916
	word[31] = 0x01                                 // hasSubscriber = true
	raw := new(big.Int).SetBytes(word)

	returnData, err := abi.Pack(abi.Types{poolKeyTuple, abi.Uint256},
		[]any{testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks}, raw)
	require.NoError(t, err)

	poolKey, info, err := DecodeGetPoolAndPositionInfo(returnData)
	require.NoError(t, err)
	require.True(t, poolKey.Currency0.Equal(testCurrency0))
	require.True(t, poolKey.Currency1.Equal(testCurrency1))
	require.Zero(t, big.NewInt(3000).Cmp(poolKey.Fee))
	require.Zero(t, big.NewInt(60).Cmp(poolKey.TickSpacing))
	require.True(t, poolKey.Hooks.Equal(testHooks))
	var wantPoolID [25]byte
	copy(wantPoolID[:], []byte{0xaa, 0xbb, 0xcc})
	require.Equal(t, wantPoolID, info.PoolID)
	require.Zero(t, big.NewInt(900).Cmp(info.TickUpper))
	require.Zero(t, big.NewInt(-916).Cmp(info.TickLower))
	require.True(t, info.HasSubscriber)
}

func TestDecodeGetPoolAndPositionInfoRejectsGarbage(t *testing.T) {
	_, _, err := DecodeGetPoolAndPositionInfo([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodePositionInfo(t *testing.T) {
	data, err := EncodePositionInfo(big.NewInt(42))
	require.NoError(t, err)
	vals, err := positionInfoFn.DecodeCall(data)
	require.NoError(t, err)
	tokenID, ok := vals[0].(*big.Int)
	require.True(t, ok)
	require.Zero(t, big.NewInt(42).Cmp(tokenID))

	word := make([]byte, 32)
	copy(word[0:25], []byte{0x11, 0x22, 0x33})
	word[25], word[26], word[27] = 0x00, 0x00, 0x3c // tickUpper = 60
	word[28], word[29], word[30] = 0xff, 0xff, 0xc4 // tickLower = -60
	word[31] = 0x00                                 // hasSubscriber = false
	raw := new(big.Int).SetBytes(word)

	returnData, err := abi.Pack(abi.Types{abi.Uint256}, raw)
	require.NoError(t, err)

	info, err := DecodePositionInfo(returnData)
	require.NoError(t, err)
	var wantPoolID [25]byte
	copy(wantPoolID[:], []byte{0x11, 0x22, 0x33})
	require.Equal(t, wantPoolID, info.PoolID)
	require.Zero(t, big.NewInt(60).Cmp(info.TickUpper))
	require.Zero(t, big.NewInt(-60).Cmp(info.TickLower))
	require.False(t, info.HasSubscriber)
}

func TestDecodePositionInfoRejectsGarbage(t *testing.T) {
	_, err := DecodePositionInfo([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeModifyLiquiditiesWithoutUnlock(t *testing.T) {
	actions := []byte{ActionDecreaseLiquidity, ActionTakePair}
	params := [][]byte{{0x04}, {0x05, 0x06}}

	data, err := EncodeModifyLiquiditiesWithoutUnlock(actions, params)
	require.NoError(t, err)

	vals, err := modifyLiquiditiesWithoutUnlockFn.DecodeCall(data)
	require.NoError(t, err)
	decodedActions, ok := vals[0].([]byte)
	require.True(t, ok)
	require.Equal(t, actions, decodedActions)
	decodedParamsAny, ok := vals[1].([]any)
	require.True(t, ok)
	decodedParams := make([][]byte, len(decodedParamsAny))
	for i, v := range decodedParamsAny {
		decodedParams[i], ok = v.([]byte)
		require.True(t, ok)
	}
	require.Equal(t, params, decodedParams)
}

func TestEncodeSubscribe(t *testing.T) {
	newSubscriber := types.MustNewAddressFromHex("0x9a8ca723f5dccb7926d00b71dec55c2fea1f50f7")

	data, err := EncodeSubscribe(big.NewInt(42), newSubscriber, []byte{0xde, 0xad})
	require.NoError(t, err)

	vals, err := subscribeFn.DecodeCall(data)
	require.NoError(t, err)
	tokenID, ok := vals[0].(*big.Int)
	require.True(t, ok)
	require.Zero(t, big.NewInt(42).Cmp(tokenID))
	decodedSubscriber, ok := vals[1].(*types.Address)
	require.True(t, ok)
	require.True(t, decodedSubscriber.Equal(newSubscriber))
	decodedData, ok := vals[2].([]byte)
	require.True(t, ok)
	require.Equal(t, []byte{0xde, 0xad}, decodedData)
}

func TestEncodeUnsubscribe(t *testing.T) {
	data, err := EncodeUnsubscribe(big.NewInt(42))
	require.NoError(t, err)

	vals, err := unsubscribeFn.DecodeCall(data)
	require.NoError(t, err)
	tokenID, ok := vals[0].(*big.Int)
	require.True(t, ok)
	require.Zero(t, big.NewInt(42).Cmp(tokenID))
}

func TestEncodeDecodeSubscriber(t *testing.T) {
	data, err := EncodeSubscriber(big.NewInt(42))
	require.NoError(t, err)
	vals, err := subscriberFn.DecodeCall(data)
	require.NoError(t, err)
	tokenID, ok := vals[0].(*big.Int)
	require.True(t, ok)
	require.Zero(t, big.NewInt(42).Cmp(tokenID))

	returnData, err := abi.Pack(abi.Types{abi.Address}, testCurrency0)
	require.NoError(t, err)

	subscriber, err := DecodeSubscriber(returnData)
	require.NoError(t, err)
	require.True(t, subscriber.Equal(testCurrency0))
}

func TestDecodeSubscriberRejectsGarbage(t *testing.T) {
	_, err := DecodeSubscriber([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeUnsubscribeGasLimit(t *testing.T) {
	data, err := EncodeUnsubscribeGasLimit()
	require.NoError(t, err)
	vals, err := unsubscribeGasLimitFn.DecodeCall(data)
	require.NoError(t, err)
	require.Empty(t, vals)

	returnData, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(300000))
	require.NoError(t, err)

	gasLimit, err := DecodeUnsubscribeGasLimit(returnData)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(300000).Cmp(gasLimit))
}

func TestDecodeUnsubscribeGasLimitRejectsGarbage(t *testing.T) {
	_, err := DecodeUnsubscribeGasLimit([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeInitializePool(t *testing.T) {
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)

	data, err := EncodeInitializePool(key, big.NewInt(792281625142643375))
	require.NoError(t, err)
	vals, err := initializePoolFn.DecodeCall(data)
	require.NoError(t, err)
	poolKeyFields, ok := vals[0].([]any)
	require.True(t, ok)
	currency0, ok := poolKeyFields[0].(*types.Address)
	require.True(t, ok)
	require.True(t, currency0.Equal(testCurrency0))
	sqrtPriceX96, ok := vals[1].(*big.Int)
	require.True(t, ok)
	require.Zero(t, big.NewInt(792281625142643375).Cmp(sqrtPriceX96))

	returnData, err := abi.Pack(abi.Types{abi.Int24}, big.NewInt(0))
	require.NoError(t, err)

	tick, err := DecodeInitializePool(returnData)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(0).Cmp(tick))
}

func TestEncodeInitializePoolRejectsNilKey(t *testing.T) {
	_, err := EncodeInitializePool(nil, big.NewInt(0))
	require.Error(t, err)
}

func TestDecodeInitializePoolRejectsGarbage(t *testing.T) {
	_, err := DecodeInitializePool([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeMulticall(t *testing.T) {
	data := [][]byte{{0x01, 0x02}, {0x03}}

	calldata, err := EncodeMulticall(data)
	require.NoError(t, err)
	vals, err := multicallFn.DecodeCall(calldata)
	require.NoError(t, err)
	elems, ok := vals[0].([]any)
	require.True(t, ok)
	decoded := make([][]byte, len(elems))
	for i, e := range elems {
		decoded[i], ok = e.([]byte)
		require.True(t, ok)
	}
	require.Equal(t, data, decoded)

	results := [][]byte{{0xaa}, {0xbb, 0xcc}}
	returnData, err := abi.Pack(abi.Types{abi.Slice(abi.Bytes)}, results)
	require.NoError(t, err)

	decodedResults, err := DecodeMulticall(returnData)
	require.NoError(t, err)
	require.Equal(t, results, decodedResults)
}

func TestDecodeMulticallRejectsGarbage(t *testing.T) {
	_, err := DecodeMulticall([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodePoolManager(t *testing.T) {
	data, err := EncodePoolManager()
	require.NoError(t, err)
	vals, err := poolManagerFn.DecodeCall(data)
	require.NoError(t, err)
	require.Empty(t, vals)

	returnData, err := abi.Pack(abi.Types{abi.Address}, testCurrency0)
	require.NoError(t, err)

	addr, err := DecodePoolManager(returnData)
	require.NoError(t, err)
	require.True(t, addr.Equal(testCurrency0))
}

func TestDecodePoolManagerRejectsGarbage(t *testing.T) {
	_, err := DecodePoolManager([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeWETH9(t *testing.T) {
	data, err := EncodeWETH9()
	require.NoError(t, err)
	vals, err := weth9Fn.DecodeCall(data)
	require.NoError(t, err)
	require.Empty(t, vals)

	returnData, err := abi.Pack(abi.Types{abi.Address}, testCurrency1)
	require.NoError(t, err)

	addr, err := DecodeWETH9(returnData)
	require.NoError(t, err)
	require.True(t, addr.Equal(testCurrency1))
}

func TestDecodeWETH9RejectsGarbage(t *testing.T) {
	_, err := DecodeWETH9([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodePermit2Contract(t *testing.T) {
	data, err := EncodePermit2Contract()
	require.NoError(t, err)
	vals, err := permit2ContractFn.DecodeCall(data)
	require.NoError(t, err)
	require.Empty(t, vals)

	returnData, err := abi.Pack(abi.Types{abi.Address}, testHooks)
	require.NoError(t, err)

	addr, err := DecodePermit2Contract(returnData)
	require.NoError(t, err)
	require.True(t, addr.Equal(testHooks))
}

func TestDecodePermit2ContractRejectsGarbage(t *testing.T) {
	_, err := DecodePermit2Contract([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeNonces(t *testing.T) {
	data, err := EncodeNonces(testCurrency0, big.NewInt(1))
	require.NoError(t, err)
	vals, err := noncesFn.DecodeCall(data)
	require.NoError(t, err)
	owner, ok := vals[0].(*types.Address)
	require.True(t, ok)
	require.True(t, owner.Equal(testCurrency0))
	word, ok := vals[1].(*big.Int)
	require.True(t, ok)
	require.Zero(t, big.NewInt(1).Cmp(word))

	returnData, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(7))
	require.NoError(t, err)

	bitmap, err := DecodeNonces(returnData)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(7).Cmp(bitmap))
}

func TestDecodeNoncesRejectsGarbage(t *testing.T) {
	_, err := DecodeNonces([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeRevokeNonce(t *testing.T) {
	data, err := EncodeRevokeNonce(big.NewInt(5))
	require.NoError(t, err)
	vals, err := revokeNonceFn.DecodeCall(data)
	require.NoError(t, err)
	nonce, ok := vals[0].(*big.Int)
	require.True(t, ok)
	require.Zero(t, big.NewInt(5).Cmp(nonce))
}

func TestEncodeERC721Permit(t *testing.T) {
	spender := types.MustNewAddressFromHex("0x9a8ca723f5dccb7926d00b71dec55c2fea1f50f7")

	data, err := EncodeERC721Permit(spender, big.NewInt(42), big.NewInt(1234567890), big.NewInt(7), []byte{0xde, 0xad})
	require.NoError(t, err)

	vals, err := erc721PermitFn.DecodeCall(data)
	require.NoError(t, err)
	decodedSpender, ok := vals[0].(*types.Address)
	require.True(t, ok)
	require.True(t, decodedSpender.Equal(spender))
	tokenID, ok := vals[1].(*big.Int)
	require.True(t, ok)
	require.Zero(t, big.NewInt(42).Cmp(tokenID))
	deadline, ok := vals[2].(*big.Int)
	require.True(t, ok)
	require.Zero(t, big.NewInt(1234567890).Cmp(deadline))
	nonce, ok := vals[3].(*big.Int)
	require.True(t, ok)
	require.Zero(t, big.NewInt(7).Cmp(nonce))
	signature, ok := vals[4].([]byte)
	require.True(t, ok)
	require.Equal(t, []byte{0xde, 0xad}, signature)
}

func TestEncodeERC721PermitForAll(t *testing.T) {
	operator := types.MustNewAddressFromHex("0x9a8ca723f5dccb7926d00b71dec55c2fea1f50f7")

	data, err := EncodeERC721PermitForAll(testCurrency0, operator, true, big.NewInt(1234567890), big.NewInt(7), []byte{0xbe, 0xef})
	require.NoError(t, err)

	vals, err := erc721PermitForAllFn.DecodeCall(data)
	require.NoError(t, err)
	owner, ok := vals[0].(*types.Address)
	require.True(t, ok)
	require.True(t, owner.Equal(testCurrency0))
	decodedOperator, ok := vals[1].(*types.Address)
	require.True(t, ok)
	require.True(t, decodedOperator.Equal(operator))
	approved, ok := vals[2].(bool)
	require.True(t, ok)
	require.True(t, approved)
	deadline, ok := vals[3].(*big.Int)
	require.True(t, ok)
	require.Zero(t, big.NewInt(1234567890).Cmp(deadline))
	nonce, ok := vals[4].(*big.Int)
	require.True(t, ok)
	require.Zero(t, big.NewInt(7).Cmp(nonce))
	signature, ok := vals[5].([]byte)
	require.True(t, ok)
	require.Equal(t, []byte{0xbe, 0xef}, signature)
}

func TestEncodeDecodePermit2Permit(t *testing.T) {
	details := permit2.NewPermitDetails(testCurrency1, big.NewInt(1000), big.NewInt(9999999999), big.NewInt(1))
	permitSingle := permit2.NewPermitSingle(details, testHooks, big.NewInt(1234567890))

	data, err := EncodePermit2Permit(testCurrency0, permitSingle, []byte{0x01, 0x02})
	require.NoError(t, err)

	vals, err := permit2PermitFn.DecodeCall(data)
	require.NoError(t, err)
	owner, ok := vals[0].(*types.Address)
	require.True(t, ok)
	require.True(t, owner.Equal(testCurrency0))
	permitSingleFields, ok := vals[1].([]any)
	require.True(t, ok)
	detailsFields, ok := permitSingleFields[0].([]any)
	require.True(t, ok)
	token, ok := detailsFields[0].(*types.Address)
	require.True(t, ok)
	require.True(t, token.Equal(testCurrency1))
	amount, ok := detailsFields[1].(*big.Int)
	require.True(t, ok)
	require.Zero(t, big.NewInt(1000).Cmp(amount))
	spender, ok := permitSingleFields[1].(*types.Address)
	require.True(t, ok)
	require.True(t, spender.Equal(testHooks))
	sigDeadline, ok := permitSingleFields[2].(*big.Int)
	require.True(t, ok)
	require.Zero(t, big.NewInt(1234567890).Cmp(sigDeadline))
	signature, ok := vals[2].([]byte)
	require.True(t, ok)
	require.Equal(t, []byte{0x01, 0x02}, signature)

	returnData, err := abi.Pack(abi.Types{abi.Bytes}, []byte{0xca, 0xfe})
	require.NoError(t, err)
	errData, err := DecodePermit2Permit(returnData)
	require.NoError(t, err)
	require.Equal(t, []byte{0xca, 0xfe}, errData)
}

func TestEncodePermit2PermitRejectsNilArgs(t *testing.T) {
	_, err := EncodePermit2Permit(testCurrency0, nil, nil)
	require.Error(t, err)

	_, err = EncodePermit2Permit(testCurrency0, permit2.NewPermitSingle(nil, testHooks, big.NewInt(0)), nil)
	require.Error(t, err)
}

func TestDecodePermit2PermitRejectsGarbage(t *testing.T) {
	_, err := DecodePermit2Permit([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodePermit2PermitBatch(t *testing.T) {
	details := []permit2.PermitDetails{
		*permit2.NewPermitDetails(testCurrency1, big.NewInt(1000), big.NewInt(9999999999), big.NewInt(1)),
		*permit2.NewPermitDetails(testHooks, big.NewInt(2000), big.NewInt(8888888888), big.NewInt(2)),
	}
	permitBatch := permit2.NewPermitBatch(details, testHooks, big.NewInt(1234567890))

	data, err := EncodePermit2PermitBatch(testCurrency0, permitBatch, []byte{0x03, 0x04})
	require.NoError(t, err)

	vals, err := permit2PermitBatchFn.DecodeCall(data)
	require.NoError(t, err)
	owner, ok := vals[0].(*types.Address)
	require.True(t, ok)
	require.True(t, owner.Equal(testCurrency0))
	permitBatchFields, ok := vals[1].([]any)
	require.True(t, ok)
	detailsElems, ok := permitBatchFields[0].([]any)
	require.True(t, ok)
	require.Len(t, detailsElems, 2)
	firstDetails, ok := detailsElems[0].([]any)
	require.True(t, ok)
	token, ok := firstDetails[0].(*types.Address)
	require.True(t, ok)
	require.True(t, token.Equal(testCurrency1))

	returnData, err := abi.Pack(abi.Types{abi.Bytes}, []byte{})
	require.NoError(t, err)
	errData, err := DecodePermit2PermitBatch(returnData)
	require.NoError(t, err)
	require.Empty(t, errData)
}

func TestEncodePermit2PermitBatchRejectsNilArgs(t *testing.T) {
	_, err := EncodePermit2PermitBatch(testCurrency0, nil, nil)
	require.Error(t, err)
}

func TestDecodePermit2PermitBatchRejectsGarbage(t *testing.T) {
	_, err := DecodePermit2PermitBatch([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodePoolKeys(t *testing.T) {
	var poolID [25]byte
	copy(poolID[:], []byte{0x11, 0x22, 0x33})

	data, err := EncodePoolKeys(poolID)
	require.NoError(t, err)
	vals, err := poolKeysFn.DecodeCall(data)
	require.NoError(t, err)
	decodedPoolID, ok := vals[0].([]byte)
	require.True(t, ok)
	require.Equal(t, poolID[:], decodedPoolID)

	returnData, err := abi.Pack(abi.Types{abi.Address, abi.Address, abi.Uint24, abi.Int24, abi.Address},
		testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	require.NoError(t, err)

	key, err := DecodePoolKeys(returnData)
	require.NoError(t, err)
	require.True(t, key.Currency0.Equal(testCurrency0))
	require.True(t, key.Currency1.Equal(testCurrency1))
	require.Zero(t, big.NewInt(3000).Cmp(key.Fee))
	require.Zero(t, big.NewInt(60).Cmp(key.TickSpacing))
	require.True(t, key.Hooks.Equal(testHooks))
}

func TestDecodePoolKeysRejectsGarbage(t *testing.T) {
	_, err := DecodePoolKeys([]byte{0x01, 0x02})
	require.Error(t, err)
}

// TestEncodeActionMintPositionMatchesCalldataDecoderLayout verifies the exact byte layout
// CalldataDecoder.decodeMintParams reads via raw assembly offsets (poolKey/tickLower/tickUpper/
// liquidity/amount0Max/amount1Max/owner inlined as flat head words, then an offset to hookData at
// word 11) -- not just that our own Encode/Decode pair round-trips with itself, which would also
// pass if both sides shared the same (wrong) layout.
func TestEncodeActionMintPositionMatchesCalldataDecoderLayout(t *testing.T) {
	owner := types.MustNewAddressFromHex("0x9a8ca723f5dccb7926d00b71dec55c2fea1f50f7")
	params := NewMintPositionParams(
		NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks),
		big.NewInt(-600), big.NewInt(600), big.NewInt(1000000), big.NewInt(1000), big.NewInt(1100), owner, []byte{0xde, 0xad},
	)

	data, err := EncodeActionMintPosition(params)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(data), 0x180)

	require.Equal(t, testCurrency0.Bytes(), data[0x0c:0x20])
	require.Equal(t, testCurrency1.Bytes(), data[0x2c:0x40])
	require.Zero(t, big.NewInt(3000).Cmp(new(big.Int).SetBytes(data[0x40:0x60])))
	require.Zero(t, big.NewInt(60).Cmp(new(big.Int).SetBytes(data[0x60:0x80])))
	require.Equal(t, testHooks.Bytes(), data[0x8c:0xa0])
	tickLowerWord := new(big.Int).SetBytes(data[0xa0:0xc0])
	tickLowerWord.Sub(tickLowerWord, new(big.Int).Lsh(big.NewInt(1), 256))
	require.Zero(t, big.NewInt(-600).Cmp(tickLowerWord))
	require.Zero(t, big.NewInt(600).Cmp(new(big.Int).SetBytes(data[0xc0:0xe0])))
	require.Zero(t, big.NewInt(1000000).Cmp(new(big.Int).SetBytes(data[0xe0:0x100])))
	require.Zero(t, big.NewInt(1000).Cmp(new(big.Int).SetBytes(data[0x100:0x120])))
	require.Zero(t, big.NewInt(1100).Cmp(new(big.Int).SetBytes(data[0x120:0x140])))
	require.Equal(t, owner.Bytes(), data[0x14c:0x160])

	hookDataOffset := new(big.Int).SetBytes(data[0x160:0x180])
	require.Zero(t, big.NewInt(0x180).Cmp(hookDataOffset), "word 11 must be the offset to hookData, per decodeMintParams' params.toBytes(11)")

	hookDataLen := new(big.Int).SetBytes(data[0x180:0x1a0])
	require.Zero(t, big.NewInt(2).Cmp(hookDataLen))
	require.Equal(t, []byte{0xde, 0xad}, data[0x1a0:0x1a2])
}

// TestEncodeActionMintPositionFromDeltasMatchesCalldataDecoderLayout verifies
// decodeMintFromDeltasParams' layout, see TestEncodeActionMintPositionMatchesCalldataDecoderLayout.
func TestEncodeActionMintPositionFromDeltasMatchesCalldataDecoderLayout(t *testing.T) {
	owner := types.MustNewAddressFromHex("0x9a8ca723f5dccb7926d00b71dec55c2fea1f50f7")
	params := NewMintPositionFromDeltasParams(
		NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks),
		big.NewInt(-600), big.NewInt(600), big.NewInt(1000), big.NewInt(1100), owner, []byte{0xbe, 0xef},
	)

	data, err := EncodeActionMintPositionFromDeltas(params)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(data), 0x160)

	require.Equal(t, testCurrency0.Bytes(), data[0x0c:0x20])
	require.Equal(t, testCurrency1.Bytes(), data[0x2c:0x40])
	tickLowerWord := new(big.Int).SetBytes(data[0xa0:0xc0])
	tickLowerWord.Sub(tickLowerWord, new(big.Int).Lsh(big.NewInt(1), 256))
	require.Zero(t, big.NewInt(-600).Cmp(tickLowerWord))
	require.Zero(t, big.NewInt(600).Cmp(new(big.Int).SetBytes(data[0xc0:0xe0])))
	require.Zero(t, big.NewInt(1000).Cmp(new(big.Int).SetBytes(data[0xe0:0x100])))
	require.Zero(t, big.NewInt(1100).Cmp(new(big.Int).SetBytes(data[0x100:0x120])))
	require.Equal(t, owner.Bytes(), data[0x12c:0x140])

	hookDataOffset := new(big.Int).SetBytes(data[0x140:0x160])
	require.Zero(t, big.NewInt(0x160).Cmp(hookDataOffset), "word 10 must be the offset to hookData, per decodeMintFromDeltasParams' params.toBytes(10)")

	hookDataLen := new(big.Int).SetBytes(data[0x160:0x180])
	require.Zero(t, big.NewInt(2).Cmp(hookDataLen))
	require.Equal(t, []byte{0xbe, 0xef}, data[0x180:0x182])
}
