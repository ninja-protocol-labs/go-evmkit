package v4

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestEncodeInitialize(t *testing.T) {
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	sqrtPriceX96 := big.NewInt(792281625142643375)

	data, err := EncodeInitialize(key, sqrtPriceX96)
	require.NoError(t, err)
	require.Equal(t, []byte{0x62, 0x76, 0xcb, 0xbe}, data[:4])

	vals, err := initializeFn.DecodeCall(data)
	require.NoError(t, err)
	keyVals, ok := vals[0].([]any)
	require.True(t, ok)
	require.True(t, keyVals[0].(*types.Address).Equal(testCurrency0))
	require.True(t, keyVals[1].(*types.Address).Equal(testCurrency1))
	require.Zero(t, big.NewInt(3000).Cmp(keyVals[2].(*big.Int)))
	require.Zero(t, big.NewInt(60).Cmp(keyVals[3].(*big.Int)))
	require.True(t, keyVals[4].(*types.Address).Equal(testHooks))
	require.Zero(t, sqrtPriceX96.Cmp(vals[1].(*big.Int)))
}

func TestEncodeInitializeRejectsNilArgs(t *testing.T) {
	_, err := EncodeInitialize(nil, big.NewInt(1))
	require.Error(t, err)
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	_, err = EncodeInitialize(key, nil)
	require.Error(t, err)
}

func TestDecodeInitialize(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Int24}, big.NewInt(42))
	require.NoError(t, err)
	tick, err := DecodeInitialize(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(42).Cmp(tick))
}

func TestDecodeInitializeRejectsGarbage(t *testing.T) {
	_, err := DecodeInitialize([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeModifyLiquidity(t *testing.T) {
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	salt := [32]byte{0x0a}
	params := NewLiquidityParams(big.NewInt(-100), big.NewInt(100), big.NewInt(1000), salt)
	hookData := []byte{0xde, 0xad}

	data, err := EncodeModifyLiquidity(key, params, hookData)
	require.NoError(t, err)
	require.Equal(t, []byte{0x5a, 0x6b, 0xcf, 0xda}, data[:4])

	vals, err := modifyLiquidityFn.DecodeCall(data)
	require.NoError(t, err)
	keyVals, ok := vals[0].([]any)
	require.True(t, ok)
	require.True(t, keyVals[0].(*types.Address).Equal(testCurrency0))
	paramsVals, ok := vals[1].([]any)
	require.True(t, ok)
	require.Zero(t, big.NewInt(-100).Cmp(paramsVals[0].(*big.Int)))
	require.Zero(t, big.NewInt(100).Cmp(paramsVals[1].(*big.Int)))
	require.Zero(t, big.NewInt(1000).Cmp(paramsVals[2].(*big.Int)))
	require.Equal(t, salt[:], paramsVals[3].([]byte))
	require.Equal(t, hookData, vals[2].([]byte))
}

func TestEncodeModifyLiquidityRejectsNilArgs(t *testing.T) {
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	params := NewLiquidityParams(big.NewInt(-100), big.NewInt(100), big.NewInt(1000), [32]byte{})
	_, err := EncodeModifyLiquidity(nil, params, nil)
	require.Error(t, err)
	_, err = EncodeModifyLiquidity(key, nil, nil)
	require.Error(t, err)
}

func TestDecodeModifyLiquidity(t *testing.T) {
	callerWord := make([]byte, 32)
	callerAmount0 := big.NewInt(100)
	callerAmount1 := big.NewInt(-50)
	copy(callerWord[16-len(callerAmount0.Bytes()):16], callerAmount0.Bytes())
	callerAmount1TwosComplement := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 128), callerAmount1)
	copy(callerWord[32-len(callerAmount1TwosComplement.Bytes()):32], callerAmount1TwosComplement.Bytes())

	feesWord := make([]byte, 32)
	feesAmount0 := big.NewInt(-7)
	feesAmount1 := big.NewInt(3)
	feesAmount0TwosComplement := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 128), feesAmount0)
	copy(feesWord[16-len(feesAmount0TwosComplement.Bytes()):16], feesAmount0TwosComplement.Bytes())
	copy(feesWord[32-len(feesAmount1.Bytes()):32], feesAmount1.Bytes())

	data, err := abi.Pack(abi.Types{abi.Bytes32, abi.Bytes32}, callerWord, feesWord)
	require.NoError(t, err)

	callerDelta, feesAccrued, err := DecodeModifyLiquidity(data)
	require.NoError(t, err)
	require.Zero(t, callerAmount0.Cmp(callerDelta.Amount0))
	require.Zero(t, callerAmount1.Cmp(callerDelta.Amount1))
	require.Zero(t, feesAmount0.Cmp(feesAccrued.Amount0))
	require.Zero(t, feesAmount1.Cmp(feesAccrued.Amount1))
}

func TestDecodeModifyLiquidityRejectsGarbage(t *testing.T) {
	_, _, err := DecodeModifyLiquidity([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeSwap(t *testing.T) {
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	params := NewSwapAmountParams(true, big.NewInt(-1000), big.NewInt(792281625142643375))
	hookData := []byte{0xbe, 0xef}

	data, err := EncodeSwap(key, params, hookData)
	require.NoError(t, err)
	require.Equal(t, []byte{0xf3, 0xcd, 0x91, 0x4c}, data[:4])

	vals, err := swapFn.DecodeCall(data)
	require.NoError(t, err)
	paramsVals, ok := vals[1].([]any)
	require.True(t, ok)
	require.True(t, paramsVals[0].(bool))
	require.Zero(t, big.NewInt(-1000).Cmp(paramsVals[1].(*big.Int)))
	require.Zero(t, big.NewInt(792281625142643375).Cmp(paramsVals[2].(*big.Int)))
	require.Equal(t, hookData, vals[2].([]byte))
}

func TestEncodeSwapRejectsNilArgs(t *testing.T) {
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	_, err := EncodeSwap(nil, NewSwapAmountParams(false, big.NewInt(1), big.NewInt(1)), nil)
	require.Error(t, err)
	_, err = EncodeSwap(key, nil, nil)
	require.Error(t, err)
}

func TestDecodeSwap(t *testing.T) {
	word := make([]byte, 32)
	amount0 := big.NewInt(-1000)
	amount1 := big.NewInt(995)
	amount0TwosComplement := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 128), amount0)
	copy(word[16-len(amount0TwosComplement.Bytes()):16], amount0TwosComplement.Bytes())
	copy(word[32-len(amount1.Bytes()):32], amount1.Bytes())

	data, err := abi.Pack(abi.Types{abi.Bytes32}, word)
	require.NoError(t, err)

	delta, err := DecodeSwap(data)
	require.NoError(t, err)
	require.Zero(t, amount0.Cmp(delta.Amount0))
	require.Zero(t, amount1.Cmp(delta.Amount1))
}

func TestDecodeSwapRejectsGarbage(t *testing.T) {
	_, err := DecodeSwap([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDonate(t *testing.T) {
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	amount0 := big.NewInt(100)
	amount1 := big.NewInt(200)
	hookData := []byte{0x01}

	data, err := EncodeDonate(key, amount0, amount1, hookData)
	require.NoError(t, err)
	require.Equal(t, []byte{0x23, 0x42, 0x66, 0xd7}, data[:4])

	vals, err := donateFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, amount0.Cmp(vals[1].(*big.Int)))
	require.Zero(t, amount1.Cmp(vals[2].(*big.Int)))
	require.Equal(t, hookData, vals[3].([]byte))
}

func TestEncodeDonateRejectsNilArgs(t *testing.T) {
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	_, err := EncodeDonate(nil, big.NewInt(1), big.NewInt(1), nil)
	require.Error(t, err)
	_, err = EncodeDonate(key, nil, big.NewInt(1), nil)
	require.Error(t, err)
	_, err = EncodeDonate(key, big.NewInt(1), nil, nil)
	require.Error(t, err)
}

func TestDecodeDonate(t *testing.T) {
	word := make([]byte, 32)
	amount0 := big.NewInt(50)
	amount1 := big.NewInt(-25)
	copy(word[16-len(amount0.Bytes()):16], amount0.Bytes())
	amount1TwosComplement := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 128), amount1)
	copy(word[32-len(amount1TwosComplement.Bytes()):32], amount1TwosComplement.Bytes())

	data, err := abi.Pack(abi.Types{abi.Bytes32}, word)
	require.NoError(t, err)

	delta, err := DecodeDonate(data)
	require.NoError(t, err)
	require.Zero(t, amount0.Cmp(delta.Amount0))
	require.Zero(t, amount1.Cmp(delta.Amount1))
}

func TestDecodeDonateRejectsGarbage(t *testing.T) {
	_, err := DecodeDonate([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeUnlock(t *testing.T) {
	unlockData := []byte{0xca, 0xfe}

	data, err := EncodeUnlock(unlockData)
	require.NoError(t, err)
	require.Equal(t, []byte{0x48, 0xc8, 0x94, 0x91}, data[:4])

	vals, err := unlockFn.DecodeCall(data)
	require.NoError(t, err)
	require.Equal(t, unlockData, vals[0].([]byte))
}

func TestDecodeUnlock(t *testing.T) {
	out := []byte{0x01, 0x02, 0x03}
	data, err := abi.Pack(abi.Types{abi.Bytes}, out)
	require.NoError(t, err)
	got, err := DecodeUnlock(data)
	require.NoError(t, err)
	require.Equal(t, out, got)
}

func TestDecodeUnlockRejectsGarbage(t *testing.T) {
	_, err := DecodeUnlock([]byte{0x01})
	require.Error(t, err)
}

func TestEncodeSync(t *testing.T) {
	data, err := EncodeSync(testCurrency0)
	require.NoError(t, err)
	require.Equal(t, []byte{0xa5, 0x84, 0x11, 0x94}, data[:4])

	vals, err := syncFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testCurrency0))
}

func TestEncodeSyncRejectsNilArgs(t *testing.T) {
	_, err := EncodeSync(nil)
	require.Error(t, err)
}

func TestEncodeTake(t *testing.T) {
	amount := big.NewInt(500)

	data, err := EncodeTake(testCurrency0, testSender, amount)
	require.NoError(t, err)
	require.Equal(t, []byte{0x0b, 0x0d, 0x9c, 0x09}, data[:4])

	vals, err := takeFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testCurrency0))
	require.True(t, vals[1].(*types.Address).Equal(testSender))
	require.Zero(t, amount.Cmp(vals[2].(*big.Int)))
}

func TestEncodeTakeRejectsNilArgs(t *testing.T) {
	_, err := EncodeTake(nil, testSender, big.NewInt(1))
	require.Error(t, err)
	_, err = EncodeTake(testCurrency0, nil, big.NewInt(1))
	require.Error(t, err)
	_, err = EncodeTake(testCurrency0, testSender, nil)
	require.Error(t, err)
}

func TestEncodeSettle(t *testing.T) {
	data, err := EncodeSettle()
	require.NoError(t, err)
	require.Equal(t, []byte{0x11, 0xda, 0x60, 0xb4}, data[:4])
	require.Len(t, data, 4)
}

func TestDecodeSettle(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(123))
	require.NoError(t, err)
	paid, err := DecodeSettle(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(123).Cmp(paid))
}

func TestDecodeSettleRejectsGarbage(t *testing.T) {
	_, err := DecodeSettle([]byte{0x01})
	require.Error(t, err)
}

func TestEncodeSettleFor(t *testing.T) {
	data, err := EncodeSettleFor(testSender)
	require.NoError(t, err)
	require.Equal(t, []byte{0x3d, 0xd4, 0x5a, 0xdb}, data[:4])

	vals, err := settleForFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testSender))
}

func TestEncodeSettleForRejectsNilArgs(t *testing.T) {
	_, err := EncodeSettleFor(nil)
	require.Error(t, err)
}

func TestDecodeSettleFor(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(456))
	require.NoError(t, err)
	paid, err := DecodeSettleFor(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(456).Cmp(paid))
}

func TestDecodeSettleForRejectsGarbage(t *testing.T) {
	_, err := DecodeSettleFor([]byte{0x01})
	require.Error(t, err)
}

func TestEncodeClear(t *testing.T) {
	amount := big.NewInt(42)

	data, err := EncodeClear(testCurrency0, amount)
	require.NoError(t, err)
	require.Equal(t, []byte{0x80, 0xf0, 0xb4, 0x4c}, data[:4])

	vals, err := clearFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testCurrency0))
	require.Zero(t, amount.Cmp(vals[1].(*big.Int)))
}

func TestEncodeClearRejectsNilArgs(t *testing.T) {
	_, err := EncodeClear(nil, big.NewInt(1))
	require.Error(t, err)
	_, err = EncodeClear(testCurrency0, nil)
	require.Error(t, err)
}

func TestEncodeMint(t *testing.T) {
	id := big.NewInt(1)
	amount := big.NewInt(100)

	data, err := EncodeMint(testSender, id, amount)
	require.NoError(t, err)
	require.Equal(t, []byte{0x15, 0x6e, 0x29, 0xf6}, data[:4])

	vals, err := mintFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testSender))
	require.Zero(t, id.Cmp(vals[1].(*big.Int)))
	require.Zero(t, amount.Cmp(vals[2].(*big.Int)))
}

func TestEncodeMintRejectsNilArgs(t *testing.T) {
	_, err := EncodeMint(nil, big.NewInt(1), big.NewInt(1))
	require.Error(t, err)
	_, err = EncodeMint(testSender, nil, big.NewInt(1))
	require.Error(t, err)
	_, err = EncodeMint(testSender, big.NewInt(1), nil)
	require.Error(t, err)
}

func TestEncodeBurn(t *testing.T) {
	id := big.NewInt(1)
	amount := big.NewInt(100)

	data, err := EncodeBurn(testSender, id, amount)
	require.NoError(t, err)
	require.Equal(t, []byte{0xf5, 0x29, 0x8a, 0xca}, data[:4])

	vals, err := burnFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testSender))
	require.Zero(t, id.Cmp(vals[1].(*big.Int)))
	require.Zero(t, amount.Cmp(vals[2].(*big.Int)))
}

func TestEncodeBurnRejectsNilArgs(t *testing.T) {
	_, err := EncodeBurn(nil, big.NewInt(1), big.NewInt(1))
	require.Error(t, err)
	_, err = EncodeBurn(testSender, nil, big.NewInt(1))
	require.Error(t, err)
	_, err = EncodeBurn(testSender, big.NewInt(1), nil)
	require.Error(t, err)
}

func TestEncodeUpdateDynamicLPFee(t *testing.T) {
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	newFee := big.NewInt(500)

	data, err := EncodeUpdateDynamicLPFee(key, newFee)
	require.NoError(t, err)
	require.Equal(t, []byte{0x52, 0x75, 0x96, 0x51}, data[:4])

	vals, err := updateDynamicLPFeeFn.DecodeCall(data)
	require.NoError(t, err)
	keyVals, ok := vals[0].([]any)
	require.True(t, ok)
	require.True(t, keyVals[0].(*types.Address).Equal(testCurrency0))
	require.Zero(t, newFee.Cmp(vals[1].(*big.Int)))
}

func TestEncodeUpdateDynamicLPFeeRejectsNilArgs(t *testing.T) {
	_, err := EncodeUpdateDynamicLPFee(nil, big.NewInt(1))
	require.Error(t, err)
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	_, err = EncodeUpdateDynamicLPFee(key, nil)
	require.Error(t, err)
}

func TestPoolKeyPoolID(t *testing.T) {
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	id, err := key.PoolID()
	require.NoError(t, err)
	require.NotEqual(t, PoolID{}, id)

	again, err := key.PoolID()
	require.NoError(t, err)
	require.Equal(t, id, again)
}

func TestPoolKeyPoolIDDiffersByField(t *testing.T) {
	key1 := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	key2 := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(500), big.NewInt(60), testHooks)

	id1, err := key1.PoolID()
	require.NoError(t, err)
	id2, err := key2.PoolID()
	require.NoError(t, err)
	require.NotEqual(t, id1, id2)
}

func TestPoolKeyPoolIDRejectsNilFields(t *testing.T) {
	key := NewPoolKey(nil, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	_, err := key.PoolID()
	require.Error(t, err)
}

func TestEncodeProtocolFeesAccrued(t *testing.T) {
	data, err := EncodeProtocolFeesAccrued(testCurrency0)
	require.NoError(t, err)
	require.Equal(t, []byte{0x97, 0xe8, 0xcd, 0x4e}, data[:4])

	vals, err := protocolFeesAccruedFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testCurrency0))
}

func TestEncodeProtocolFeesAccruedRejectsNilArgs(t *testing.T) {
	_, err := EncodeProtocolFeesAccrued(nil)
	require.Error(t, err)
}

func TestDecodeProtocolFeesAccrued(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(789))
	require.NoError(t, err)
	amount, err := DecodeProtocolFeesAccrued(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(789).Cmp(amount))
}

func TestDecodeProtocolFeesAccruedRejectsGarbage(t *testing.T) {
	_, err := DecodeProtocolFeesAccrued([]byte{0x01})
	require.Error(t, err)
}

func TestEncodeSetProtocolFee(t *testing.T) {
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	newFee := big.NewInt(500)

	data, err := EncodeSetProtocolFee(key, newFee)
	require.NoError(t, err)
	require.Equal(t, []byte{0x7e, 0x87, 0xce, 0x7d}, data[:4])

	vals, err := setProtocolFeeFn.DecodeCall(data)
	require.NoError(t, err)
	keyVals, ok := vals[0].([]any)
	require.True(t, ok)
	require.True(t, keyVals[0].(*types.Address).Equal(testCurrency0))
	require.Zero(t, newFee.Cmp(vals[1].(*big.Int)))
}

func TestEncodeSetProtocolFeeRejectsNilArgs(t *testing.T) {
	_, err := EncodeSetProtocolFee(nil, big.NewInt(1))
	require.Error(t, err)
}

func TestEncodeSetProtocolFeeController(t *testing.T) {
	data, err := EncodeSetProtocolFeeController(testSender)
	require.NoError(t, err)
	require.Equal(t, []byte{0x2d, 0x77, 0x13, 0x89}, data[:4])

	vals, err := setProtocolFeeControllerFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testSender))
}

func TestEncodeSetProtocolFeeControllerRejectsNilArgs(t *testing.T) {
	_, err := EncodeSetProtocolFeeController(nil)
	require.Error(t, err)
}

func TestEncodeCollectProtocolFees(t *testing.T) {
	amount := big.NewInt(1000)

	data, err := EncodeCollectProtocolFees(testSender, testCurrency0, amount)
	require.NoError(t, err)
	require.Equal(t, []byte{0x81, 0x61, 0xb8, 0x74}, data[:4])

	vals, err := collectProtocolFeesFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testSender))
	require.True(t, vals[1].(*types.Address).Equal(testCurrency0))
	require.Zero(t, amount.Cmp(vals[2].(*big.Int)))
}

func TestEncodeCollectProtocolFeesRejectsNilArgs(t *testing.T) {
	_, err := EncodeCollectProtocolFees(nil, testCurrency0, big.NewInt(1))
	require.Error(t, err)
	_, err = EncodeCollectProtocolFees(testSender, nil, big.NewInt(1))
	require.Error(t, err)
	_, err = EncodeCollectProtocolFees(testSender, testCurrency0, nil)
	require.Error(t, err)
}

func TestDecodeCollectProtocolFees(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(999))
	require.NoError(t, err)
	amountCollected, err := DecodeCollectProtocolFees(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(999).Cmp(amountCollected))
}

func TestDecodeCollectProtocolFeesRejectsGarbage(t *testing.T) {
	_, err := DecodeCollectProtocolFees([]byte{0x01})
	require.Error(t, err)
}

func TestEncodeProtocolFeeController(t *testing.T) {
	data, err := EncodeProtocolFeeController()
	require.NoError(t, err)
	require.Equal(t, []byte{0xf0, 0x2d, 0xe3, 0xb2}, data[:4])
	require.Len(t, data, 4)
}

func TestDecodeProtocolFeeController(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Address}, testSender)
	require.NoError(t, err)
	controller, err := DecodeProtocolFeeController(data)
	require.NoError(t, err)
	require.True(t, controller.Equal(testSender))
}

func TestDecodeProtocolFeeControllerRejectsGarbage(t *testing.T) {
	_, err := DecodeProtocolFeeController([]byte{0x01})
	require.Error(t, err)
}

func TestEncodeExtsload(t *testing.T) {
	slot := [32]byte{0x01, 0x02}

	data, err := EncodeExtsload(slot)
	require.NoError(t, err)
	require.Equal(t, []byte{0x1e, 0x2e, 0xae, 0xaf}, data[:4])

	vals, err := extsloadFn.DecodeCall(data)
	require.NoError(t, err)
	require.Equal(t, slot[:], vals[0].([]byte))
}

func TestDecodeExtsload(t *testing.T) {
	value := [32]byte{}
	value[31] = 0x2a
	data, err := abi.Pack(abi.Types{abi.Bytes32}, value[:])
	require.NoError(t, err)

	got, err := DecodeExtsload(data)
	require.NoError(t, err)
	require.Equal(t, value, got)
}

func TestDecodeExtsloadRejectsGarbage(t *testing.T) {
	_, err := DecodeExtsload([]byte{0x01})
	require.Error(t, err)
}

func TestEncodeExtsloadRange(t *testing.T) {
	startSlot := [32]byte{0x03}
	nSlots := big.NewInt(5)

	data, err := EncodeExtsloadRange(startSlot, nSlots)
	require.NoError(t, err)
	require.Equal(t, []byte{0x35, 0xfd, 0x63, 0x1a}, data[:4])

	vals, err := extsloadRangeFn.DecodeCall(data)
	require.NoError(t, err)
	require.Equal(t, startSlot[:], vals[0].([]byte))
	require.Zero(t, nSlots.Cmp(vals[1].(*big.Int)))
}

func TestDecodeExtsloadRange(t *testing.T) {
	v1 := [32]byte{}
	v1[31] = 0x01
	v2 := [32]byte{}
	v2[31] = 0x02
	data, err := abi.Pack(abi.Types{abi.Slice(abi.Bytes32)}, [][]byte{v1[:], v2[:]})
	require.NoError(t, err)

	got, err := DecodeExtsloadRange(data)
	require.NoError(t, err)
	require.Equal(t, [][32]byte{v1, v2}, got)
}

func TestDecodeExtsloadRangeEmpty(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Slice(abi.Bytes32)}, [][]byte{})
	require.NoError(t, err)

	got, err := DecodeExtsloadRange(data)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Empty(t, got)
}

func TestDecodeExtsloadRangeRejectsGarbage(t *testing.T) {
	_, err := DecodeExtsloadRange([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeExtsloadSlots(t *testing.T) {
	slots := [][32]byte{{0x01}, {0x02}}

	data, err := EncodeExtsloadSlots(slots)
	require.NoError(t, err)
	require.Equal(t, []byte{0xdb, 0xd0, 0x35, 0xff}, data[:4])

	vals, err := extsloadSlotsFn.DecodeCall(data)
	require.NoError(t, err)
	gotSlots, ok := vals[0].([]any)
	require.True(t, ok)
	require.Len(t, gotSlots, 2)
	require.Equal(t, slots[0][:], gotSlots[0].([]byte))
	require.Equal(t, slots[1][:], gotSlots[1].([]byte))
}

func TestDecodeExtsloadSlots(t *testing.T) {
	v1 := [32]byte{}
	v1[31] = 0x0a
	data, err := abi.Pack(abi.Types{abi.Slice(abi.Bytes32)}, [][]byte{v1[:]})
	require.NoError(t, err)

	got, err := DecodeExtsloadSlots(data)
	require.NoError(t, err)
	require.Equal(t, [][32]byte{v1}, got)
}

func TestDecodeExtsloadSlotsRejectsGarbage(t *testing.T) {
	_, err := DecodeExtsloadSlots([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeExttload(t *testing.T) {
	slot := [32]byte{0x04}

	data, err := EncodeExttload(slot)
	require.NoError(t, err)
	require.Equal(t, []byte{0xf1, 0x35, 0xba, 0xaa}, data[:4])

	vals, err := exttloadFn.DecodeCall(data)
	require.NoError(t, err)
	require.Equal(t, slot[:], vals[0].([]byte))
}

func TestDecodeExttload(t *testing.T) {
	value := [32]byte{}
	value[31] = 0x2b
	data, err := abi.Pack(abi.Types{abi.Bytes32}, value[:])
	require.NoError(t, err)

	got, err := DecodeExttload(data)
	require.NoError(t, err)
	require.Equal(t, value, got)
}

func TestDecodeExttloadRejectsGarbage(t *testing.T) {
	_, err := DecodeExttload([]byte{0x01})
	require.Error(t, err)
}

func TestEncodeExttloadSlots(t *testing.T) {
	slots := [][32]byte{{0x05}, {0x06}}

	data, err := EncodeExttloadSlots(slots)
	require.NoError(t, err)
	require.Equal(t, []byte{0x9b, 0xf6, 0x64, 0x5f}, data[:4])

	vals, err := exttloadSlotsFn.DecodeCall(data)
	require.NoError(t, err)
	gotSlots, ok := vals[0].([]any)
	require.True(t, ok)
	require.Len(t, gotSlots, 2)
	require.Equal(t, slots[0][:], gotSlots[0].([]byte))
	require.Equal(t, slots[1][:], gotSlots[1].([]byte))
}

func TestDecodeExttloadSlots(t *testing.T) {
	v1 := [32]byte{}
	v1[31] = 0x0b
	data, err := abi.Pack(abi.Types{abi.Slice(abi.Bytes32)}, [][]byte{v1[:]})
	require.NoError(t, err)

	got, err := DecodeExttloadSlots(data)
	require.NoError(t, err)
	require.Equal(t, [][32]byte{v1}, got)
}

func TestDecodeExttloadSlotsRejectsGarbage(t *testing.T) {
	_, err := DecodeExttloadSlots([]byte{0x01, 0x02})
	require.Error(t, err)
}
