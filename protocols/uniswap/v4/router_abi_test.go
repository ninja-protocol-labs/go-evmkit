package v4

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestEncodeDecodeActionSwapExactInSingle(t *testing.T) {
	poolKey := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	params := NewExactInputSingleParams(poolKey, true, big.NewInt(1000), big.NewInt(900), big.NewInt(123456789), []byte{0xde, 0xad})

	data, err := EncodeActionSwapExactInSingle(params)
	require.NoError(t, err)

	decoded, err := DecodeActionSwapExactInSingle(data)
	require.NoError(t, err)
	require.True(t, decoded.PoolKey.Currency0.Equal(testCurrency0))
	require.True(t, decoded.PoolKey.Currency1.Equal(testCurrency1))
	require.Zero(t, big.NewInt(3000).Cmp(decoded.PoolKey.Fee))
	require.Zero(t, big.NewInt(60).Cmp(decoded.PoolKey.TickSpacing))
	require.True(t, decoded.PoolKey.Hooks.Equal(testHooks))
	require.True(t, decoded.ZeroForOne)
	require.Zero(t, big.NewInt(1000).Cmp(decoded.AmountIn))
	require.Zero(t, big.NewInt(900).Cmp(decoded.AmountOutMinimum))
	require.Zero(t, big.NewInt(123456789).Cmp(decoded.MinHopPriceX36))
	require.Equal(t, []byte{0xde, 0xad}, decoded.HookData)
}

func TestEncodeActionSwapExactInSingleRejectsNilArgs(t *testing.T) {
	_, err := EncodeActionSwapExactInSingle(nil)
	require.Error(t, err)

	_, err = EncodeActionSwapExactInSingle(NewExactInputSingleParams(nil, true, big.NewInt(1), big.NewInt(1), big.NewInt(0), nil))
	require.Error(t, err)
}

func TestDecodeActionSwapExactInSingleRejectsGarbage(t *testing.T) {
	_, err := DecodeActionSwapExactInSingle([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeActionSwapExactIn(t *testing.T) {
	path := []*PathKey{
		NewPathKey(testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks, []byte{0x01}),
	}
	minHopPriceX36 := []*big.Int{big.NewInt(42)}
	params := NewExactInputParams(testCurrency0, path, minHopPriceX36, big.NewInt(1000), big.NewInt(900))

	data, err := EncodeActionSwapExactIn(params)
	require.NoError(t, err)

	decoded, err := DecodeActionSwapExactIn(data)
	require.NoError(t, err)
	require.True(t, decoded.CurrencyIn.Equal(testCurrency0))
	require.Len(t, decoded.Path, 1)
	require.True(t, decoded.Path[0].IntermediateCurrency.Equal(testCurrency1))
	require.Zero(t, big.NewInt(3000).Cmp(decoded.Path[0].Fee))
	require.Zero(t, big.NewInt(60).Cmp(decoded.Path[0].TickSpacing))
	require.True(t, decoded.Path[0].Hooks.Equal(testHooks))
	require.Equal(t, []byte{0x01}, decoded.Path[0].HookData)
	require.Len(t, decoded.MinHopPriceX36, 1)
	require.Zero(t, big.NewInt(42).Cmp(decoded.MinHopPriceX36[0]))
	require.Zero(t, big.NewInt(1000).Cmp(decoded.AmountIn))
	require.Zero(t, big.NewInt(900).Cmp(decoded.AmountOutMinimum))
}

func TestEncodeActionSwapExactInRejectsNilArgs(t *testing.T) {
	_, err := EncodeActionSwapExactIn(nil)
	require.Error(t, err)

	_, err = EncodeActionSwapExactIn(NewExactInputParams(testCurrency0, []*PathKey{nil}, nil, big.NewInt(1), big.NewInt(1)))
	require.Error(t, err)
}

func TestDecodeActionSwapExactInRejectsGarbage(t *testing.T) {
	_, err := DecodeActionSwapExactIn([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeActionSwapExactOutSingle(t *testing.T) {
	poolKey := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	params := NewExactOutputSingleParams(poolKey, false, big.NewInt(1000), big.NewInt(1100), big.NewInt(123456789), []byte{0xbe, 0xef})

	data, err := EncodeActionSwapExactOutSingle(params)
	require.NoError(t, err)

	decoded, err := DecodeActionSwapExactOutSingle(data)
	require.NoError(t, err)
	require.True(t, decoded.PoolKey.Currency0.Equal(testCurrency0))
	require.True(t, decoded.PoolKey.Currency1.Equal(testCurrency1))
	require.Zero(t, big.NewInt(3000).Cmp(decoded.PoolKey.Fee))
	require.Zero(t, big.NewInt(60).Cmp(decoded.PoolKey.TickSpacing))
	require.True(t, decoded.PoolKey.Hooks.Equal(testHooks))
	require.False(t, decoded.ZeroForOne)
	require.Zero(t, big.NewInt(1000).Cmp(decoded.AmountOut))
	require.Zero(t, big.NewInt(1100).Cmp(decoded.AmountInMaximum))
	require.Zero(t, big.NewInt(123456789).Cmp(decoded.MinHopPriceX36))
	require.Equal(t, []byte{0xbe, 0xef}, decoded.HookData)
}

func TestEncodeActionSwapExactOutSingleRejectsNilArgs(t *testing.T) {
	_, err := EncodeActionSwapExactOutSingle(nil)
	require.Error(t, err)

	_, err = EncodeActionSwapExactOutSingle(NewExactOutputSingleParams(nil, true, big.NewInt(1), big.NewInt(1), big.NewInt(0), nil))
	require.Error(t, err)
}

func TestDecodeActionSwapExactOutSingleRejectsGarbage(t *testing.T) {
	_, err := DecodeActionSwapExactOutSingle([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeActionSwapExactOut(t *testing.T) {
	path := []*PathKey{
		NewPathKey(testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks, []byte{0x02}),
	}
	minHopPriceX36 := []*big.Int{big.NewInt(43)}
	params := NewExactOutputParams(testCurrency0, path, minHopPriceX36, big.NewInt(1000), big.NewInt(1100))

	data, err := EncodeActionSwapExactOut(params)
	require.NoError(t, err)

	decoded, err := DecodeActionSwapExactOut(data)
	require.NoError(t, err)
	require.True(t, decoded.CurrencyOut.Equal(testCurrency0))
	require.Len(t, decoded.Path, 1)
	require.True(t, decoded.Path[0].IntermediateCurrency.Equal(testCurrency1))
	require.Zero(t, big.NewInt(3000).Cmp(decoded.Path[0].Fee))
	require.Zero(t, big.NewInt(60).Cmp(decoded.Path[0].TickSpacing))
	require.True(t, decoded.Path[0].Hooks.Equal(testHooks))
	require.Equal(t, []byte{0x02}, decoded.Path[0].HookData)
	require.Len(t, decoded.MinHopPriceX36, 1)
	require.Zero(t, big.NewInt(43).Cmp(decoded.MinHopPriceX36[0]))
	require.Zero(t, big.NewInt(1000).Cmp(decoded.AmountOut))
	require.Zero(t, big.NewInt(1100).Cmp(decoded.AmountInMaximum))
}

func TestEncodeActionSwapExactOutRejectsNilArgs(t *testing.T) {
	_, err := EncodeActionSwapExactOut(nil)
	require.Error(t, err)

	_, err = EncodeActionSwapExactOut(NewExactOutputParams(testCurrency0, []*PathKey{nil}, nil, big.NewInt(1), big.NewInt(1)))
	require.Error(t, err)
}

func TestDecodeActionSwapExactOutRejectsGarbage(t *testing.T) {
	_, err := DecodeActionSwapExactOut([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeActionSettle(t *testing.T) {
	data, err := EncodeActionSettle(testCurrency0, big.NewInt(1000), true)
	require.NoError(t, err)

	currency, amount, payerIsUser, err := DecodeActionSettle(data)
	require.NoError(t, err)
	require.True(t, currency.Equal(testCurrency0))
	require.Zero(t, big.NewInt(1000).Cmp(amount))
	require.True(t, payerIsUser)
}

func TestDecodeActionSettleRejectsGarbage(t *testing.T) {
	_, _, _, err := DecodeActionSettle([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeActionSettleAll(t *testing.T) {
	data, err := EncodeActionSettleAll(testCurrency0, big.NewInt(1000))
	require.NoError(t, err)

	currency, maxAmount, err := DecodeActionSettleAll(data)
	require.NoError(t, err)
	require.True(t, currency.Equal(testCurrency0))
	require.Zero(t, big.NewInt(1000).Cmp(maxAmount))
}

func TestDecodeActionSettleAllRejectsGarbage(t *testing.T) {
	_, _, err := DecodeActionSettleAll([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeActionTake(t *testing.T) {
	recipient := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	data, err := EncodeActionTake(testCurrency0, recipient, big.NewInt(1000))
	require.NoError(t, err)

	currency, decodedRecipient, amount, err := DecodeActionTake(data)
	require.NoError(t, err)
	require.True(t, currency.Equal(testCurrency0))
	require.True(t, decodedRecipient.Equal(recipient))
	require.Zero(t, big.NewInt(1000).Cmp(amount))
}

func TestDecodeActionTakeRejectsGarbage(t *testing.T) {
	_, _, _, err := DecodeActionTake([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeActionTakeAll(t *testing.T) {
	data, err := EncodeActionTakeAll(testCurrency0, big.NewInt(1000))
	require.NoError(t, err)

	currency, minAmount, err := DecodeActionTakeAll(data)
	require.NoError(t, err)
	require.True(t, currency.Equal(testCurrency0))
	require.Zero(t, big.NewInt(1000).Cmp(minAmount))
}

func TestDecodeActionTakeAllRejectsGarbage(t *testing.T) {
	_, _, err := DecodeActionTakeAll([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeActionTakePortion(t *testing.T) {
	recipient := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	data, err := EncodeActionTakePortion(testCurrency0, recipient, big.NewInt(5000))
	require.NoError(t, err)

	currency, decodedRecipient, bips, err := DecodeActionTakePortion(data)
	require.NoError(t, err)
	require.True(t, currency.Equal(testCurrency0))
	require.True(t, decodedRecipient.Equal(recipient))
	require.Zero(t, big.NewInt(5000).Cmp(bips))
}

func TestDecodeActionTakePortionRejectsGarbage(t *testing.T) {
	_, _, _, err := DecodeActionTakePortion([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeDecodeActions(t *testing.T) {
	actions := []byte{ActionSettleAll, ActionTakeAll}
	params := [][]byte{{0x01, 0x02}, {0x03, 0x04, 0x05}}

	data, err := EncodeActions(actions, params)
	require.NoError(t, err)

	decodedActions, decodedParams, err := DecodeActions(data)
	require.NoError(t, err)
	require.Equal(t, actions, decodedActions)
	require.Equal(t, params, decodedParams)
}

func TestDecodeActionsRejectsGarbage(t *testing.T) {
	_, _, err := DecodeActions([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeExecuteActions(t *testing.T) {
	actions := []byte{ActionSettleAll, ActionTakeAll}
	params := [][]byte{{0x01, 0x02}, {0x03}}

	data, err := EncodeExecuteActions(actions, params)
	require.NoError(t, err)

	vals, err := executeActionsFn.DecodeCall(data)
	require.NoError(t, err)
	inner, ok := vals[0].([]byte)
	require.True(t, ok)

	decodedActions, decodedParams, err := DecodeActions(inner)
	require.NoError(t, err)
	require.Equal(t, actions, decodedActions)
	require.Equal(t, params, decodedParams)
}

func TestActionConstants(t *testing.T) {
	require.True(t, MsgSender().Equal(types.MustNewAddressFromHex("0x0000000000000000000000000000000000000001")))
	require.True(t, AddressThis().Equal(types.MustNewAddressFromHex("0x0000000000000000000000000000000000000002")))
	require.Zero(t, big.NewInt(0).Cmp(OpenDelta()))
	require.Zero(t, new(big.Int).Lsh(big.NewInt(1), 255).Cmp(ContractBalance()))
}

func TestActionConstantsReturnFreshValuesEachCall(t *testing.T) {
	a := MsgSender()
	a.SetBytes([]byte{0xff})
	require.True(t, MsgSender().Equal(types.MustNewAddressFromHex("0x0000000000000000000000000000000000000001")), "mutating one MsgSender() call's result must not affect the next call")

	d := OpenDelta()
	d.Add(d, big.NewInt(1))
	require.Zero(t, big.NewInt(0).Cmp(OpenDelta()), "mutating one OpenDelta() call's result must not affect the next call")
}
