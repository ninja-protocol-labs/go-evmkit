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

	plan := NewPlanner().Append(wrap).Append(swap).AppendAllowRevert(check).Append(sweep)
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

	decoded, err := DecodePlanner(commands, inputs)
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
	inner := NewPlanner().Append(NewUnwrapWETHInput(MsgSender(), big.NewInt(0))).AppendAllowRevert(NewTransferInput(testTokenA, testTokenB, big.NewInt(3)))
	middle := NewPlanner().AppendAllowRevert(NewExecuteSubPlanInput(inner)).Append(NewPayPortionInput(testTokenA, testTokenC, big.NewInt(25)))
	outer := NewPlanner().Append(NewWrapETHInput(AddressThis(), big.NewInt(1))).AppendAllowRevert(NewExecuteSubPlanInput(middle))

	commands, inputs, err := outer.Encode()
	require.NoError(t, err)
	require.Equal(t, []byte{0x0b, 0xa1}, commands)

	decoded, err := DecodePlanner(commands, inputs)
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
		plan *Planner
		want string
	}{
		{"nil", nil, "invalid request: plan is nil"},
		{"empty", NewPlanner(), "invalid request: plan has no steps"},
		{"zero step", &Planner{Steps: []Step{{}}}, "invalid request: steps[0] input is nil"},
		{"nil input", &Planner{Steps: []Step{*NewStep(NewSweepInput(testTokenA, testTokenB, big.NewInt(1)), false), *NewStep(nil, true)}}, "invalid request: steps[1] input is nil"},
		{"invalid input", NewPlanner().Append(NewWrapETHInput(testTokenA, big.NewInt(1))).Append(NewTransferInput(testTokenA, nil, big.NewInt(1))), "steps[1] TRANSFER: invalid request: recipient is nil"},
		{"invalid nested", NewPlanner().Append(NewExecuteSubPlanInput(NewPlanner().Append(&PayPortionInput{Token: testTokenA, Recipient: testTokenB, Bips: big.NewInt(10_001)}))), "steps[0] EXECUTE_SUB_PLAN: plan: steps[0] PAY_PORTION: invalid request: bips 10001 exceeds 10000"},
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

	_, err = DecodePlanner([]byte{0x0b, 0x0b}, [][]byte{wrap})
	require.EqualError(t, err, "universalrouter: decodePlan: 2 commands but 1 inputs")

	_, err = DecodePlanner([]byte{0x0b}, [][]byte{wrap, wrap})
	require.EqualError(t, err, "universalrouter: decodePlan: 1 commands but 2 inputs")

	_, err = DecodePlanner([]byte{0x0b, 0x8f}, [][]byte{wrap, wrap})
	require.ErrorIs(t, err, ErrUnknownCommand)
	require.ErrorContains(t, err, "universalrouter: decodePlan: steps[1]: universalrouter: decode 0x0f: ")

	_, err = DecodePlanner([]byte{0x04}, [][]byte{wrap})
	require.ErrorContains(t, err, "universalrouter: decodePlan: steps[0]: universalrouter: decode SWEEP: ")

	empty, err := DecodePlanner(nil, nil)
	require.NoError(t, err)
	require.Empty(t, empty.Steps)
	require.Error(t, empty.Validate())

	badSub, err := abi.Pack(subPlanTypes, []byte{0x0b}, []any{[]byte{0x01}})
	require.NoError(t, err)
	_, err = DecodePlanner([]byte{byte(CommandExecuteSubPlan)}, [][]byte{badSub})
	require.ErrorContains(t, err, "universalrouter: decode EXECUTE_SUB_PLAN: universalrouter: decodePlan: steps[0]: universalrouter: decode WRAP_ETH: ")

	mismatchedSub, err := abi.Pack(subPlanTypes, []byte{0x0b, 0x0b}, []any{wrap})
	require.NoError(t, err)
	_, err = DecodePlanner([]byte{byte(CommandExecuteSubPlan)}, [][]byte{mismatchedSub})
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

	commands, inputs, err := NewPlanner().Append(NewV4SwapInput(actions, params)).Encode()
	require.NoError(t, err)
	require.Equal(t, []byte{0x10}, commands)

	unlockData, err := v4.EncodeActions(actions, params)
	require.NoError(t, err)
	require.Equal(t, unlockData, inputs[0])

	decoded, err := DecodePlanner(commands, inputs)
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

	plan := NewPlanner()
	require.Same(t, plan, plan.Append(sweep))
	require.Same(t, plan, plan.AppendAllowRevert(sweep))
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

func TestPlanCalldataMatchesExecuteEncoders(t *testing.T) {
	plan := NewPlanner().
		Append(NewWrapETHInput(AddressThis(), big.NewInt(1))).
		Append(NewSweepInput(testTokenA, MsgSender(), big.NewInt(0)))
	commands, inputs, err := plan.Encode()
	require.NoError(t, err)

	withoutDeadline, err := plan.Calldata(nil)
	require.NoError(t, err)
	want, err := EncodeExecute(commands, inputs)
	require.NoError(t, err)
	require.Equal(t, want, withoutDeadline)
	decoded, err := DecodeExecuteCall(withoutDeadline)
	require.NoError(t, err)
	require.Nil(t, decoded.Deadline)
	require.Nil(t, decoded.Signed)
	gotCommands, gotInputs, err := decoded.Plan.Encode()
	require.NoError(t, err)
	require.Equal(t, commands, gotCommands)
	require.Equal(t, inputs, gotInputs)

	deadline := big.NewInt(1_900_000_000)
	withDeadline, err := plan.Calldata(deadline)
	require.NoError(t, err)
	want, err = EncodeExecuteWithDeadline(commands, inputs, deadline)
	require.NoError(t, err)
	require.Equal(t, want, withDeadline)
	decoded, err = DecodeExecuteCall(withDeadline)
	require.NoError(t, err)
	require.Zero(t, deadline.Cmp(decoded.Deadline))

	empty, err := NewPlanner().Calldata(deadline)
	require.Nil(t, empty)
	require.ErrorIs(t, err, bind.ErrInvalidRequest)
	require.ErrorContains(t, err, "plan has no steps")

	var nilPlan *Planner
	nilCalldata, err := nilPlan.Calldata(nil)
	require.Nil(t, nilCalldata)
	require.ErrorContains(t, err, "plan is nil")
}

func TestPlanExecuteRequest(t *testing.T) {
	router := types.MustNewAddressFromHex("0x66a9893cC07D91D95644AEDD05D03f95e1dBA8Af")
	plan := NewPlanner().Append(NewWrapETHInput(AddressThis(), big.NewInt(1)))
	deadline := big.NewInt(1_900_000_000)
	native := big.NewInt(1)

	req := plan.ExecuteRequest(router, deadline, native)
	require.Same(t, router, req.Contract)
	require.Same(t, plan, req.Plan)
	require.Same(t, deadline, req.Deadline)
	require.Same(t, native, req.NativeAmount)
	require.NoError(t, req.Validate())

	req = plan.ExecuteRequest(router, nil, nil)
	require.Nil(t, req.Deadline)
	require.Nil(t, req.NativeAmount)
	require.NoError(t, req.Validate())
}

func TestNewV4SwapInputFromPlanner(t *testing.T) {
	key := v4.NewPoolKey(testTokenA, testTokenB, big.NewInt(3000), big.NewInt(60), types.ZeroAddress())
	swap := v4.NewExactInputSingleParams(key, true, big.NewInt(1000), big.NewInt(900), big.NewInt(0), []byte{})
	planner := v4.NewPlanner()
	require.NoError(t, planner.AppendSwapExactInSingle(swap))
	require.NoError(t, planner.AppendSettleAll(testTokenA, big.NewInt(1000)))
	require.NoError(t, planner.AppendTakeAll(testTokenB, big.NewInt(900)))

	input := NewV4SwapInputFromPlanner(planner)
	require.NoError(t, input.Validate())
	require.Equal(t, []byte{v4.ActionSwapExactInSingle, v4.ActionSettleAll, v4.ActionTakeAll}, input.Actions)

	swapParams, err := v4.EncodeActionSwapExactInSingle(swap)
	require.NoError(t, err)
	settleParams, err := v4.EncodeActionSettleAll(testTokenA, big.NewInt(1000))
	require.NoError(t, err)
	takeParams, err := v4.EncodeActionTakeAll(testTokenB, big.NewInt(900))
	require.NoError(t, err)
	manual := NewV4SwapInput(
		[]byte{v4.ActionSwapExactInSingle, v4.ActionSettleAll, v4.ActionTakeAll},
		[][]byte{swapParams, settleParams, takeParams},
	)
	got, err := input.Encode()
	require.NoError(t, err)
	want, err := manual.Encode()
	require.NoError(t, err)
	require.Equal(t, want, got)

	input.Actions[0] = v4.ActionSweep
	require.Equal(t, []byte{v4.ActionSwapExactInSingle, v4.ActionSettleAll, v4.ActionTakeAll}, planner.Actions())

	empty := NewV4SwapInputFromPlanner(v4.NewPlanner())
	err = empty.Validate()
	require.ErrorIs(t, err, bind.ErrInvalidRequest)
	require.ErrorContains(t, err, "actions is empty")
	_, _, err = NewPlanner().Append(NewWrapETHInput(AddressThis(), big.NewInt(1))).Append(empty).Encode()
	require.ErrorContains(t, err, "steps[1] V4_SWAP")
}
