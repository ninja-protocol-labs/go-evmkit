package universalrouter

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	v3 "github.com/ninja-protocol-labs/go-evmkit/protocols/uniswap/v3"
	v4 "github.com/ninja-protocol-labs/go-evmkit/protocols/uniswap/v4"
)

func TestPlanEncodeBuildsCommandsAndInputs(t *testing.T) {
	path, err := v3.EncodePath([]*types.Address{testTokenA, testTokenB}, []*big.Int{big.NewInt(3000)})
	require.NoError(t, err)

	wrap := NewWrapETHInput(AddressThis(), ContractBalance())
	swap := NewV3SwapExactInInput(MsgSender(), ContractBalance(), big.NewInt(1), path, false, nil)
	check := NewBalanceCheckERC20Input(testTokenC, testTokenB, big.NewInt(1))
	sweep := NewSweepInput(types.ZeroAddress(), MsgSender(), big.NewInt(0))

	plan := NewPlan().Add(wrap).Add(swap).AddAllowRevert(check).Add(sweep)
	require.Len(t, plan.Steps, 4)

	commands, inputs, err := plan.Encode()
	require.NoError(t, err)
	require.Equal(t, []byte{0x0b, 0x00, 0x8e, 0x04}, commands)
	require.Len(t, inputs, 4)

	for i, in := range []Input{wrap, swap, check, sweep} {
		want, err := in.Encode()
		require.NoError(t, err)
		require.Equal(t, want, inputs[i])
	}

	decoded, err := DecodePlan(commands, inputs)
	require.NoError(t, err)
	require.Len(t, decoded.Steps, 4)
	for i, s := range decoded.Steps {
		require.Equal(t, plan.Steps[i].Input.Command(), s.Input.Command())
		require.Equal(t, plan.Steps[i].AllowRevert, s.AllowRevert)
	}

	againCommands, againInputs, err := decoded.Encode()
	require.NoError(t, err)
	require.Equal(t, commands, againCommands)
	require.Equal(t, inputs, againInputs)
}

func TestPlanNestedSubPlansRoundTrip(t *testing.T) {
	inner := NewPlan().Add(NewUnwrapWETHInput(MsgSender(), big.NewInt(0))).AddAllowRevert(NewTransferInput(testTokenA, testTokenB, big.NewInt(3)))
	middle := NewPlan().AddAllowRevert(NewExecuteSubPlanInput(inner)).Add(NewPayPortionInput(testTokenA, testTokenC, big.NewInt(25)))
	outer := NewPlan().Add(NewWrapETHInput(AddressThis(), big.NewInt(1))).AddAllowRevert(NewExecuteSubPlanInput(middle))

	commands, inputs, err := outer.Encode()
	require.NoError(t, err)
	require.Equal(t, []byte{0x0b, 0xa1}, commands)

	decoded, err := DecodePlan(commands, inputs)
	require.NoError(t, err)
	require.Len(t, decoded.Steps, 2)
	require.True(t, decoded.Steps[1].AllowRevert)

	sub, ok := decoded.Steps[1].Input.(*ExecuteSubPlanInput)
	require.True(t, ok)
	require.Len(t, sub.Plan.Steps, 2)
	require.True(t, sub.Plan.Steps[0].AllowRevert)
	require.Equal(t, CommandPayPortion, sub.Plan.Steps[1].Input.Command())

	innerDecoded, ok := sub.Plan.Steps[0].Input.(*ExecuteSubPlanInput)
	require.True(t, ok)
	require.Len(t, innerDecoded.Plan.Steps, 2)
	require.Equal(t, CommandUnwrapWETH, innerDecoded.Plan.Steps[0].Input.Command())
	require.False(t, innerDecoded.Plan.Steps[0].AllowRevert)
	transfer, ok := innerDecoded.Plan.Steps[1].Input.(*TransferInput)
	require.True(t, ok)
	require.True(t, innerDecoded.Plan.Steps[1].AllowRevert)
	require.True(t, transfer.Recipient.Equal(testTokenB))
	require.Zero(t, big.NewInt(3).Cmp(transfer.Value))

	againCommands, againInputs, err := decoded.Encode()
	require.NoError(t, err)
	require.Equal(t, commands, againCommands)
	require.Equal(t, inputs, againInputs)
}

func TestPlanValidateReportsStepIndexAndCommand(t *testing.T) {
	for _, tc := range []struct {
		name string
		plan *Plan
		want string
	}{
		{"nil", nil, "invalid request: plan is nil"},
		{"empty", NewPlan(), "invalid request: plan has no steps"},
		{"zero step", &Plan{Steps: []Step{{}}}, "invalid request: steps[0] input is nil"},
		{"nil input", &Plan{Steps: []Step{*NewStep(NewSweepInput(testTokenA, testTokenB, big.NewInt(1)), false), *NewStep(nil, true)}}, "invalid request: steps[1] input is nil"},
		{"invalid input", NewPlan().Add(NewWrapETHInput(testTokenA, big.NewInt(1))).Add(NewTransferInput(testTokenA, nil, big.NewInt(1))), "steps[1] TRANSFER: invalid request: recipient is nil"},
		{"invalid nested", NewPlan().Add(NewExecuteSubPlanInput(NewPlan().Add(&PayPortionInput{Token: testTokenA, Recipient: testTokenB, Bips: big.NewInt(10_001)}))), "steps[0] EXECUTE_SUB_PLAN: plan: steps[0] PAY_PORTION: invalid request: bips 10001 exceeds 10000"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.plan.Validate()
			require.ErrorIs(t, err, bind.ErrInvalidRequest)
			require.EqualError(t, err, tc.want)

			commands, inputs, err := tc.plan.Encode()
			require.Nil(t, commands)
			require.Nil(t, inputs)
			require.ErrorIs(t, err, bind.ErrInvalidRequest)
			require.EqualError(t, err, "universalrouter: encodePlan: "+tc.want)
		})
	}
}

func TestDecodePlanRejectsBadShapes(t *testing.T) {
	wrap, err := NewWrapETHInput(testTokenA, big.NewInt(1)).Encode()
	require.NoError(t, err)

	_, err = DecodePlan([]byte{0x0b, 0x0b}, [][]byte{wrap})
	require.EqualError(t, err, "universalrouter: decodePlan: 2 commands but 1 inputs")

	_, err = DecodePlan([]byte{0x0b}, [][]byte{wrap, wrap})
	require.EqualError(t, err, "universalrouter: decodePlan: 1 commands but 2 inputs")

	_, err = DecodePlan([]byte{0x0b, 0x8f}, [][]byte{wrap, wrap})
	require.ErrorIs(t, err, ErrUnknownCommand)
	require.ErrorContains(t, err, "universalrouter: decodePlan: steps[1]: universalrouter: decode 0x0f: ")

	_, err = DecodePlan([]byte{0x04}, [][]byte{wrap})
	require.ErrorContains(t, err, "universalrouter: decodePlan: steps[0]: universalrouter: decode SWEEP: ")

	empty, err := DecodePlan(nil, nil)
	require.NoError(t, err)
	require.Empty(t, empty.Steps)
	require.Error(t, empty.Validate())

	badSub, err := abi.Pack(subPlanTypes, []byte{0x0b}, []any{[]byte{0x01}})
	require.NoError(t, err)
	_, err = DecodePlan([]byte{byte(CommandExecuteSubPlan)}, [][]byte{badSub})
	require.ErrorContains(t, err, "universalrouter: decode EXECUTE_SUB_PLAN: universalrouter: decodePlan: steps[0]: universalrouter: decode WRAP_ETH: ")

	mismatchedSub, err := abi.Pack(subPlanTypes, []byte{0x0b, 0x0b}, []any{wrap})
	require.NoError(t, err)
	_, err = DecodePlan([]byte{byte(CommandExecuteSubPlan)}, [][]byte{mismatchedSub})
	require.ErrorContains(t, err, "2 commands but 1 inputs")
}

func TestPlanV4SwapUsesV4ActionEncoders(t *testing.T) {
	key := v4.NewPoolKey(types.ZeroAddress(), testTokenB, big.NewInt(3000), big.NewInt(60), types.ZeroAddress())
	swapData, err := v4.EncodeActionSwapExactInSingle(v4.NewExactInputSingleParams(key, true, big.NewInt(1_000), big.NewInt(990), big.NewInt(0), nil))
	require.NoError(t, err)
	settleData, err := v4.EncodeActionSettleAll(types.ZeroAddress(), big.NewInt(1_000))
	require.NoError(t, err)
	takeData, err := v4.EncodeActionTakeAll(testTokenB, big.NewInt(990))
	require.NoError(t, err)

	actions := []byte{v4.ActionSwapExactInSingle, v4.ActionSettleAll, v4.ActionTakeAll}
	params := [][]byte{swapData, settleData, takeData}

	commands, inputs, err := NewPlan().Add(NewV4SwapInput(actions, params)).Encode()
	require.NoError(t, err)
	require.Equal(t, []byte{0x10}, commands)

	unlockData, err := v4.EncodeActions(actions, params)
	require.NoError(t, err)
	require.Equal(t, unlockData, inputs[0])

	decoded, err := DecodePlan(commands, inputs)
	require.NoError(t, err)
	swap, ok := decoded.Steps[0].Input.(*V4SwapInput)
	require.True(t, ok)
	require.Equal(t, actions, swap.Actions)
	require.Equal(t, params, swap.Params)

	single, err := v4.DecodeActionSwapExactInSingle(swap.Params[0])
	require.NoError(t, err)
	require.True(t, single.PoolKey.Currency1.Equal(testTokenB))
	require.Zero(t, big.NewInt(990).Cmp(single.AmountOutMinimum))
}

func TestNewStepAndConstructorsSetFields(t *testing.T) {
	sweep := NewSweepInput(testTokenA, testTokenB, big.NewInt(1))
	step := NewStep(sweep, true)
	require.Same(t, sweep, step.Input)
	require.True(t, step.AllowRevert)

	plan := NewPlan()
	require.Same(t, plan, plan.Add(sweep))
	require.Same(t, plan, plan.AddAllowRevert(sweep))
	require.Len(t, plan.Steps, 2)
	require.False(t, plan.Steps[0].AllowRevert)
	require.True(t, plan.Steps[1].AllowRevert)

	signed := NewSignedRoute([32]byte{0x01}, [32]byte{0x02}, true, [32]byte{0x03}, []byte{0x04})
	require.Equal(t, &SignedRoute{Intent: [32]byte{0x01}, Data: [32]byte{0x02}, VerifySender: true, Nonce: [32]byte{0x03}, Signature: []byte{0x04}}, signed)

	call := NewExecuteCall(plan, big.NewInt(9), signed)
	require.Same(t, plan, call.Plan)
	require.Same(t, signed, call.Signed)
	require.Zero(t, big.NewInt(9).Cmp(call.Deadline))

	require.Equal(t, &V3SwapExactInInput{Recipient: testTokenA, AmountIn: big.NewInt(1), AmountOutMin: big.NewInt(2), Path: []byte{0x03}, PayerIsUser: true, MinHopPriceX36: []*big.Int{big.NewInt(4)}}, NewV3SwapExactInInput(testTokenA, big.NewInt(1), big.NewInt(2), []byte{0x03}, true, []*big.Int{big.NewInt(4)}))
	require.Equal(t, &V2SwapExactOutInput{Recipient: testTokenA, AmountOut: big.NewInt(1), AmountInMax: big.NewInt(2), Path: []*types.Address{testTokenB}, PayerIsUser: false}, NewV2SwapExactOutInput(testTokenA, big.NewInt(1), big.NewInt(2), []*types.Address{testTokenB}, false, nil))
	require.Equal(t, &PayPortionFullPrecisionInput{Token: testTokenA, Recipient: testTokenB, Portion: big.NewInt(5)}, NewPayPortionFullPrecisionInput(testTokenA, testTokenB, big.NewInt(5)))
	require.Equal(t, &V4SwapInput{Actions: []byte{0x06}, Params: [][]byte{{0x01}}}, NewV4SwapInput([]byte{0x06}, [][]byte{{0x01}}))
	require.Equal(t, &V3PositionManagerCallInput{Calldata: []byte{0x01}}, NewV3PositionManagerCallInput([]byte{0x01}))
	require.Equal(t, &AcrossV4DepositV3Input{Depositor: testTokenA, Recipient: testTokenB, InputToken: testTokenC, OutputToken: testTokenA, InputAmount: big.NewInt(1), OutputAmount: big.NewInt(2), DestinationChainID: big.NewInt(3), ExclusiveRelayer: testTokenB, QuoteTimestamp: 4, FillDeadline: 5, ExclusivityDeadline: 6, Message: []byte{0x07}, UseNative: true}, NewAcrossV4DepositV3Input(testTokenA, testTokenB, testTokenC, testTokenA, big.NewInt(1), big.NewInt(2), big.NewInt(3), testTokenB, 4, 5, 6, []byte{0x07}, true))
}

func TestRawCalldataInputsAreCopiedNotAliased(t *testing.T) {
	calldata := append([]byte{0x42, 0x96, 0x6c, 0x68}, make([]byte, 32)...)
	in := NewV3PositionManagerCallInput(calldata)

	encoded, err := in.Encode()
	require.NoError(t, err)
	encoded[4] = 0xff
	require.Equal(t, byte(0x00), calldata[4])

	decoded, err := DecodeV3PositionManagerCallInput(calldata)
	require.NoError(t, err)
	calldata[5] = 0xee
	require.Equal(t, byte(0x00), decoded.Calldata[5])
}
