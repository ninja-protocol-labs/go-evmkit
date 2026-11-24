package v4

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlannerMatchesEncoders(t *testing.T) {
	poolKey := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	path := []*PathKey{NewPathKey(testCurrency1, big.NewInt(500), big.NewInt(10), testHooks, []byte{0x01})}
	exactInSingle := NewExactInputSingleParams(poolKey, true, big.NewInt(1000), big.NewInt(900), big.NewInt(0), []byte{0xde, 0xad})
	exactIn := NewExactInputParams(testCurrency0, path, []*big.Int{}, big.NewInt(1000), big.NewInt(900))
	exactOutSingle := NewExactOutputSingleParams(poolKey, false, big.NewInt(500), big.NewInt(600), big.NewInt(0), []byte{})
	exactOut := NewExactOutputParams(testCurrency1, path, []*big.Int{}, big.NewInt(500), big.NewInt(600))
	mint := NewMintPositionParams(poolKey, big.NewInt(-60), big.NewInt(60), big.NewInt(1_000_000), big.NewInt(10), big.NewInt(20), testSender, []byte{})
	mintFromDeltas := NewMintPositionFromDeltasParams(poolKey, big.NewInt(-120), big.NewInt(120), big.NewInt(30), big.NewInt(40), testSender, []byte{0x02})
	tokenID := big.NewInt(7)

	plan := NewPlanner()
	require.NoError(t, plan.AppendSwapExactInSingle(exactInSingle))
	require.NoError(t, plan.AppendSwapExactIn(exactIn))
	require.NoError(t, plan.AppendSwapExactOutSingle(exactOutSingle))
	require.NoError(t, plan.AppendSwapExactOut(exactOut))
	require.NoError(t, plan.AppendSettle(testCurrency0, big.NewInt(11), true))
	require.NoError(t, plan.AppendSettleAll(testCurrency0, big.NewInt(12)))
	require.NoError(t, plan.AppendSettlePair(testCurrency0, testCurrency1))
	require.NoError(t, plan.AppendTake(testCurrency1, testSender, big.NewInt(13)))
	require.NoError(t, plan.AppendTakeAll(testCurrency1, big.NewInt(14)))
	require.NoError(t, plan.AppendTakePortion(testCurrency1, testSender, big.NewInt(25)))
	require.NoError(t, plan.AppendTakePair(testCurrency0, testCurrency1, testSender))
	require.NoError(t, plan.AppendCloseCurrency(testCurrency0))
	require.NoError(t, plan.AppendClearOrTake(testCurrency1, big.NewInt(15)))
	require.NoError(t, plan.AppendSweep(testCurrency0, testSender))
	require.NoError(t, plan.AppendWrap(big.NewInt(16)))
	require.NoError(t, plan.AppendUnwrap(big.NewInt(17)))
	require.NoError(t, plan.AppendMintPosition(mint))
	require.NoError(t, plan.AppendMintPositionFromDeltas(mintFromDeltas))
	require.NoError(t, plan.AppendIncreaseLiquidity(tokenID, big.NewInt(18), big.NewInt(19), big.NewInt(20), []byte{}))
	require.NoError(t, plan.AppendIncreaseLiquidityFromDeltas(tokenID, big.NewInt(21), big.NewInt(22), []byte{}))
	require.NoError(t, plan.AppendDecreaseLiquidity(tokenID, big.NewInt(23), big.NewInt(24), big.NewInt(25), []byte{}))
	require.NoError(t, plan.AppendBurnPosition(tokenID, big.NewInt(26), big.NewInt(27), []byte{}))
	plan.Append(ActionDonate, []byte{0xaa, 0xbb})

	actions, params := plan.Actions(), plan.Params()
	require.Equal(t, 23, plan.Len())
	require.Equal(t, []byte{
		ActionSwapExactInSingle,
		ActionSwapExactIn,
		ActionSwapExactOutSingle,
		ActionSwapExactOut,
		ActionSettle,
		ActionSettleAll,
		ActionSettlePair,
		ActionTake,
		ActionTakeAll,
		ActionTakePortion,
		ActionTakePair,
		ActionCloseCurrency,
		ActionClearOrTake,
		ActionSweep,
		ActionWrap,
		ActionUnwrap,
		ActionMintPosition,
		ActionMintPositionFromDeltas,
		ActionIncreaseLiquidity,
		ActionIncreaseLiquidityFromDeltas,
		ActionDecreaseLiquidity,
		ActionBurnPosition,
		ActionDonate,
	}, actions)

	want := make([][]byte, 0, 23)
	data, err := EncodeActionSwapExactInSingle(exactInSingle)
	require.NoError(t, err)
	want = append(want, data)
	data, err = EncodeActionSwapExactIn(exactIn)
	require.NoError(t, err)
	want = append(want, data)
	data, err = EncodeActionSwapExactOutSingle(exactOutSingle)
	require.NoError(t, err)
	want = append(want, data)
	data, err = EncodeActionSwapExactOut(exactOut)
	require.NoError(t, err)
	want = append(want, data)
	data, err = EncodeActionSettle(testCurrency0, big.NewInt(11), true)
	require.NoError(t, err)
	want = append(want, data)
	data, err = EncodeActionSettleAll(testCurrency0, big.NewInt(12))
	require.NoError(t, err)
	want = append(want, data)
	data, err = EncodeActionSettlePair(testCurrency0, testCurrency1)
	require.NoError(t, err)
	want = append(want, data)
	data, err = EncodeActionTake(testCurrency1, testSender, big.NewInt(13))
	require.NoError(t, err)
	want = append(want, data)
	data, err = EncodeActionTakeAll(testCurrency1, big.NewInt(14))
	require.NoError(t, err)
	want = append(want, data)
	data, err = EncodeActionTakePortion(testCurrency1, testSender, big.NewInt(25))
	require.NoError(t, err)
	want = append(want, data)
	data, err = EncodeActionTakePair(testCurrency0, testCurrency1, testSender)
	require.NoError(t, err)
	want = append(want, data)
	data, err = EncodeActionCloseCurrency(testCurrency0)
	require.NoError(t, err)
	want = append(want, data)
	data, err = EncodeActionClearOrTake(testCurrency1, big.NewInt(15))
	require.NoError(t, err)
	want = append(want, data)
	data, err = EncodeActionSweep(testCurrency0, testSender)
	require.NoError(t, err)
	want = append(want, data)
	data, err = EncodeActionWrap(big.NewInt(16))
	require.NoError(t, err)
	want = append(want, data)
	data, err = EncodeActionUnwrap(big.NewInt(17))
	require.NoError(t, err)
	want = append(want, data)
	data, err = EncodeActionMintPosition(mint)
	require.NoError(t, err)
	want = append(want, data)
	data, err = EncodeActionMintPositionFromDeltas(mintFromDeltas)
	require.NoError(t, err)
	want = append(want, data)
	data, err = EncodeActionIncreaseLiquidity(tokenID, big.NewInt(18), big.NewInt(19), big.NewInt(20), []byte{})
	require.NoError(t, err)
	want = append(want, data)
	data, err = EncodeActionIncreaseLiquidityFromDeltas(tokenID, big.NewInt(21), big.NewInt(22), []byte{})
	require.NoError(t, err)
	want = append(want, data)
	data, err = EncodeActionDecreaseLiquidity(tokenID, big.NewInt(23), big.NewInt(24), big.NewInt(25), []byte{})
	require.NoError(t, err)
	want = append(want, data)
	data, err = EncodeActionBurnPosition(tokenID, big.NewInt(26), big.NewInt(27), []byte{})
	require.NoError(t, err)
	want = append(want, data)
	want = append(want, []byte{0xaa, 0xbb})
	require.Equal(t, want, params)

	unlockData, err := EncodeActions(actions, params)
	require.NoError(t, err)
	decodedActions, decodedParams, err := DecodeActions(unlockData)
	require.NoError(t, err)
	require.Equal(t, actions, decodedActions)
	require.Equal(t, params, decodedParams)
}

func TestPlannerAppendReturnsErrorWithoutAppending(t *testing.T) {
	plan := NewPlanner()
	require.NoError(t, plan.AppendSettleAll(testCurrency0, big.NewInt(1)))

	err := plan.AppendSwapExactInSingle(nil)
	require.ErrorContains(t, err, "uniswapv4: planner: append SWAP_EXACT_IN_SINGLE: ")
	require.ErrorContains(t, err, "params must not be nil")
	require.Equal(t, 1, plan.Len())

	require.NoError(t, plan.AppendTakeAll(testCurrency1, big.NewInt(2)))
	require.Equal(t, []byte{ActionSettleAll, ActionTakeAll}, plan.Actions())
	require.Len(t, plan.Params(), 2)
}

func TestPlannerStartsEmpty(t *testing.T) {
	plan := NewPlanner()
	require.Zero(t, plan.Len())
	require.Empty(t, plan.Actions())
	require.Empty(t, plan.Params())
}

func TestPlannerCopiesParams(t *testing.T) {
	raw := []byte{0x01, 0x02}
	plan := NewPlanner()
	plan.Append(ActionDonate, raw)
	raw[0] = 0xff
	require.Equal(t, [][]byte{{0x01, 0x02}}, plan.Params())

	actions, params := plan.Actions(), plan.Params()
	actions[0] = ActionSweep
	params[0][1] = 0xee
	require.Equal(t, []byte{ActionDonate}, plan.Actions())
	require.Equal(t, [][]byte{{0x01, 0x02}}, plan.Params())
}
