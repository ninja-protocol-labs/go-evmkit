package universalrouter

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/protocols/permit2"
	v4 "github.com/ninja-protocol-labs/go-evmkit/protocols/uniswap/v4"
)

// Step is one command of a plan; AllowRevert lets the plan continue when the command fails.
type Step struct {
	Input       Input
	AllowRevert bool
}

func NewStep(input Input, allowRevert bool) *Step {
	return &Step{
		Input:       input,
		AllowRevert: allowRevert,
	}
}

// Plan is an ordered list of commands that encodes to execute's commands and inputs.
type Plan struct {
	Steps []Step
}

func NewPlan() *Plan {
	return new(Plan)
}

// Add appends input as a step that reverts the whole plan when it fails.
func (p *Plan) Add(input Input) *Plan {
	p.Steps = append(p.Steps, *NewStep(input, false))
	return p
}

// AddAllowRevert appends input as a step whose failure is tolerated.
func (p *Plan) AddAllowRevert(input Input) *Plan {
	p.Steps = append(p.Steps, *NewStep(input, true))
	return p
}

// Validate checks every step's input, reporting the first failure with its index and command.
func (p *Plan) Validate() error {
	if p == nil {
		return fmt.Errorf("%w: plan is nil", bind.ErrInvalidRequest)
	}

	if len(p.Steps) == 0 {
		return fmt.Errorf("%w: plan has no steps", bind.ErrInvalidRequest)
	}

	for i, s := range p.Steps {
		if s.Input == nil {
			return fmt.Errorf("%w: steps[%d] input is nil", bind.ErrInvalidRequest, i)
		}

		if err := s.Input.Validate(); err != nil {
			return fmt.Errorf("steps[%d] %s: %w", i, s.Input.Command(), err)
		}
	}

	return nil
}

// Encode returns execute's commands byte string and inputs, one per step.
func (p *Plan) Encode() ([]byte, [][]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, nil, fmt.Errorf("universalrouter: encodePlan: %w", err)
	}

	commands := make([]byte, len(p.Steps))
	inputs := make([][]byte, len(p.Steps))
	for i, s := range p.Steps {
		data, err := s.Input.Encode()
		if err != nil {
			return nil, nil, fmt.Errorf("universalrouter: encodePlan: steps[%d]: %w", i, err)
		}
		commands[i] = s.Input.Command().Byte(s.AllowRevert)
		inputs[i] = data
	}

	return commands, inputs, nil
}

// DecodePlan decodes execute's commands and inputs back into typed steps, recursing into sub-plans.
func DecodePlan(commands []byte, inputs [][]byte) (*Plan, error) {
	if len(commands) != len(inputs) {
		return nil, fmt.Errorf("universalrouter: decodePlan: %d commands but %d inputs", len(commands), len(inputs))
	}

	plan := NewPlan()
	for i, b := range commands {
		cmd, allowRevert := ParseCommandByte(b)
		input, err := DecodeInput(cmd, inputs[i])
		if err != nil {
			return nil, fmt.Errorf("universalrouter: decodePlan: steps[%d]: %w", i, err)
		}
		plan.Steps = append(plan.Steps, *NewStep(input, allowRevert))
	}

	return plan, nil
}

// DecodeInput decodes one command's input bytes into its typed Input via the command's Decode function.
func DecodeInput(cmd Command, data []byte) (Input, error) {
	var (
		in  Input
		err error
	)

	switch cmd {
	case CommandV3SwapExactIn:
		in, err = DecodeV3SwapExactInInput(data)
	case CommandV3SwapExactOut:
		in, err = DecodeV3SwapExactOutInput(data)
	case CommandPermit2TransferFrom:
		in, err = DecodePermit2TransferFromInput(data)
	case CommandPermit2PermitBatch:
		in, err = DecodePermit2PermitBatchInput(data)
	case CommandSweep:
		in, err = DecodeSweepInput(data)
	case CommandTransfer:
		in, err = DecodeTransferInput(data)
	case CommandPayPortion:
		in, err = DecodePayPortionInput(data)
	case CommandPayPortionFullPrecision:
		in, err = DecodePayPortionFullPrecisionInput(data)
	case CommandV2SwapExactIn:
		in, err = DecodeV2SwapExactInInput(data)
	case CommandV2SwapExactOut:
		in, err = DecodeV2SwapExactOutInput(data)
	case CommandPermit2Permit:
		in, err = DecodePermit2PermitInput(data)
	case CommandWrapETH:
		in, err = DecodeWrapETHInput(data)
	case CommandUnwrapWETH:
		in, err = DecodeUnwrapWETHInput(data)
	case CommandPermit2TransferFromBatch:
		in, err = DecodePermit2TransferFromBatchInput(data)
	case CommandBalanceCheckERC20:
		in, err = DecodeBalanceCheckERC20Input(data)
	case CommandV4Swap:
		in, err = DecodeV4SwapInput(data)
	case CommandV3PositionManagerPermit:
		in, err = DecodeV3PositionManagerPermitInput(data)
	case CommandV3PositionManagerCall:
		in, err = DecodeV3PositionManagerCallInput(data)
	case CommandV4InitializePool:
		in, err = DecodeV4InitializePoolInput(data)
	case CommandV4PositionManagerCall:
		in, err = DecodeV4PositionManagerCallInput(data)
	case CommandExecuteSubPlan:
		in, err = DecodeExecuteSubPlanInput(data)
	case CommandAcrossV4DepositV3:
		in, err = DecodeAcrossV4DepositV3Input(data)
	default:
		return nil, fmt.Errorf("universalrouter: decode %s: %w", cmd, ErrUnknownCommand)
	}

	if err != nil {
		return nil, err
	}

	return in, nil
}

// SignedRoute holds executeSigned's signature-specific arguments.
type SignedRoute struct {
	Intent       [32]byte
	Data         [32]byte
	VerifySender bool
	Nonce        [32]byte
	Signature    []byte
}

func NewSignedRoute(intent, data [32]byte, verifySender bool, nonce [32]byte, signature []byte) *SignedRoute {
	return &SignedRoute{
		Intent:       intent,
		Data:         data,
		VerifySender: verifySender,
		Nonce:        nonce,
		Signature:    signature,
	}
}

// ExecuteCall is decoded execute or executeSigned calldata; Deadline is nil for execute(commands, inputs) and Signed is nil unless executeSigned.
type ExecuteCall struct {
	Plan     *Plan
	Deadline *big.Int
	Signed   *SignedRoute
}

func NewExecuteCall(plan *Plan, deadline *big.Int, signed *SignedRoute) *ExecuteCall {
	return &ExecuteCall{
		Plan:     plan,
		Deadline: deadline,
		Signed:   signed,
	}
}

// DecodeV3SwapExactInInput decodes V3_SWAP_EXACT_IN input bytes.
func DecodeV3SwapExactInInput(data []byte) (*V3SwapExactInInput, error) {
	vals, err := abi.Unpack(v3SwapTypes, data)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: decode V3_SWAP_EXACT_IN: %w", err)
	}

	recipient, ok := vals[0].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V3_SWAP_EXACT_IN: unexpected decoded type for recipient: %T", vals[0])
	}

	amount, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V3_SWAP_EXACT_IN: unexpected decoded type for amountIn: %T", vals[1])
	}

	limit, ok := vals[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V3_SWAP_EXACT_IN: unexpected decoded type for amountOutMin: %T", vals[2])
	}

	path, ok := vals[3].([]byte)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V3_SWAP_EXACT_IN: unexpected decoded type for path: %T", vals[3])
	}

	payerIsUser, ok := vals[4].(bool)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V3_SWAP_EXACT_IN: unexpected decoded type for payerIsUser: %T", vals[4])
	}

	rawHops, ok := vals[5].([]any)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V3_SWAP_EXACT_IN: unexpected decoded type for minHopPriceX36: %T", vals[5])
	}

	minHopPriceX36 := make([]*big.Int, len(rawHops))
	for i, e := range rawHops {
		n, ok := e.(*big.Int)
		if !ok {
			return nil, fmt.Errorf("universalrouter: decode V3_SWAP_EXACT_IN: unexpected decoded type for minHopPriceX36[%d]: %T", i, e)
		}
		minHopPriceX36[i] = n
	}

	return NewV3SwapExactInInput(recipient, amount, limit, path, payerIsUser, minHopPriceX36), nil
}

// DecodeV3SwapExactOutInput decodes V3_SWAP_EXACT_OUT input bytes.
func DecodeV3SwapExactOutInput(data []byte) (*V3SwapExactOutInput, error) {
	vals, err := abi.Unpack(v3SwapTypes, data)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: decode V3_SWAP_EXACT_OUT: %w", err)
	}

	recipient, ok := vals[0].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V3_SWAP_EXACT_OUT: unexpected decoded type for recipient: %T", vals[0])
	}

	amount, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V3_SWAP_EXACT_OUT: unexpected decoded type for amountOut: %T", vals[1])
	}

	limit, ok := vals[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V3_SWAP_EXACT_OUT: unexpected decoded type for amountInMax: %T", vals[2])
	}

	path, ok := vals[3].([]byte)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V3_SWAP_EXACT_OUT: unexpected decoded type for path: %T", vals[3])
	}

	payerIsUser, ok := vals[4].(bool)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V3_SWAP_EXACT_OUT: unexpected decoded type for payerIsUser: %T", vals[4])
	}

	rawHops, ok := vals[5].([]any)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V3_SWAP_EXACT_OUT: unexpected decoded type for minHopPriceX36: %T", vals[5])
	}

	minHopPriceX36 := make([]*big.Int, len(rawHops))
	for i, e := range rawHops {
		n, ok := e.(*big.Int)
		if !ok {
			return nil, fmt.Errorf("universalrouter: decode V3_SWAP_EXACT_OUT: unexpected decoded type for minHopPriceX36[%d]: %T", i, e)
		}
		minHopPriceX36[i] = n
	}

	return NewV3SwapExactOutInput(recipient, amount, limit, path, payerIsUser, minHopPriceX36), nil
}

// DecodeV2SwapExactInInput decodes V2_SWAP_EXACT_IN input bytes.
func DecodeV2SwapExactInInput(data []byte) (*V2SwapExactInInput, error) {
	vals, err := abi.Unpack(v2SwapTypes, data)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: decode V2_SWAP_EXACT_IN: %w", err)
	}

	recipient, ok := vals[0].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V2_SWAP_EXACT_IN: unexpected decoded type for recipient: %T", vals[0])
	}

	amount, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V2_SWAP_EXACT_IN: unexpected decoded type for amountIn: %T", vals[1])
	}

	limit, ok := vals[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V2_SWAP_EXACT_IN: unexpected decoded type for amountOutMin: %T", vals[2])
	}

	rawPath, ok := vals[3].([]any)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V2_SWAP_EXACT_IN: unexpected decoded type for path: %T", vals[3])
	}

	path := make([]*types.Address, len(rawPath))
	for i, v := range rawPath {
		a, ok := v.(*types.Address)
		if !ok {
			return nil, fmt.Errorf("universalrouter: decode V2_SWAP_EXACT_IN: unexpected decoded type for path[%d]: %T", i, v)
		}
		path[i] = a
	}

	payerIsUser, ok := vals[4].(bool)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V2_SWAP_EXACT_IN: unexpected decoded type for payerIsUser: %T", vals[4])
	}

	rawHops, ok := vals[5].([]any)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V2_SWAP_EXACT_IN: unexpected decoded type for minHopPriceX36: %T", vals[5])
	}

	minHopPriceX36 := make([]*big.Int, len(rawHops))
	for i, e := range rawHops {
		n, ok := e.(*big.Int)
		if !ok {
			return nil, fmt.Errorf("universalrouter: decode V2_SWAP_EXACT_IN: unexpected decoded type for minHopPriceX36[%d]: %T", i, e)
		}
		minHopPriceX36[i] = n
	}

	return NewV2SwapExactInInput(recipient, amount, limit, path, payerIsUser, minHopPriceX36), nil
}

// DecodeV2SwapExactOutInput decodes V2_SWAP_EXACT_OUT input bytes.
func DecodeV2SwapExactOutInput(data []byte) (*V2SwapExactOutInput, error) {
	vals, err := abi.Unpack(v2SwapTypes, data)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: decode V2_SWAP_EXACT_OUT: %w", err)
	}

	recipient, ok := vals[0].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V2_SWAP_EXACT_OUT: unexpected decoded type for recipient: %T", vals[0])
	}

	amount, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V2_SWAP_EXACT_OUT: unexpected decoded type for amountOut: %T", vals[1])
	}

	limit, ok := vals[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V2_SWAP_EXACT_OUT: unexpected decoded type for amountInMax: %T", vals[2])
	}

	rawPath, ok := vals[3].([]any)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V2_SWAP_EXACT_OUT: unexpected decoded type for path: %T", vals[3])
	}

	path := make([]*types.Address, len(rawPath))
	for i, v := range rawPath {
		a, ok := v.(*types.Address)
		if !ok {
			return nil, fmt.Errorf("universalrouter: decode V2_SWAP_EXACT_OUT: unexpected decoded type for path[%d]: %T", i, v)
		}
		path[i] = a
	}

	payerIsUser, ok := vals[4].(bool)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V2_SWAP_EXACT_OUT: unexpected decoded type for payerIsUser: %T", vals[4])
	}

	rawHops, ok := vals[5].([]any)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V2_SWAP_EXACT_OUT: unexpected decoded type for minHopPriceX36: %T", vals[5])
	}

	minHopPriceX36 := make([]*big.Int, len(rawHops))
	for i, e := range rawHops {
		n, ok := e.(*big.Int)
		if !ok {
			return nil, fmt.Errorf("universalrouter: decode V2_SWAP_EXACT_OUT: unexpected decoded type for minHopPriceX36[%d]: %T", i, e)
		}
		minHopPriceX36[i] = n
	}

	return NewV2SwapExactOutInput(recipient, amount, limit, path, payerIsUser, minHopPriceX36), nil
}

// DecodeSweepInput decodes SWEEP input bytes.
func DecodeSweepInput(data []byte) (*SweepInput, error) {
	vals, err := abi.Unpack(tokenRecipientAmountTypes, data)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: decode SWEEP: %w", err)
	}

	token, ok := vals[0].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode SWEEP: unexpected decoded type for token: %T", vals[0])
	}

	recipient, ok := vals[1].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode SWEEP: unexpected decoded type for recipient: %T", vals[1])
	}

	amountMin, ok := vals[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode SWEEP: unexpected decoded type for amountMin: %T", vals[2])
	}

	return NewSweepInput(token, recipient, amountMin), nil
}

// DecodeTransferInput decodes TRANSFER input bytes.
func DecodeTransferInput(data []byte) (*TransferInput, error) {
	vals, err := abi.Unpack(tokenRecipientAmountTypes, data)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: decode TRANSFER: %w", err)
	}

	token, ok := vals[0].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode TRANSFER: unexpected decoded type for token: %T", vals[0])
	}

	recipient, ok := vals[1].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode TRANSFER: unexpected decoded type for recipient: %T", vals[1])
	}

	value, ok := vals[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode TRANSFER: unexpected decoded type for value: %T", vals[2])
	}

	return NewTransferInput(token, recipient, value), nil
}

// DecodePayPortionInput decodes PAY_PORTION input bytes.
func DecodePayPortionInput(data []byte) (*PayPortionInput, error) {
	vals, err := abi.Unpack(tokenRecipientAmountTypes, data)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: decode PAY_PORTION: %w", err)
	}

	token, ok := vals[0].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode PAY_PORTION: unexpected decoded type for token: %T", vals[0])
	}

	recipient, ok := vals[1].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode PAY_PORTION: unexpected decoded type for recipient: %T", vals[1])
	}

	bips, ok := vals[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode PAY_PORTION: unexpected decoded type for bips: %T", vals[2])
	}

	return NewPayPortionInput(token, recipient, bips), nil
}

// DecodePayPortionFullPrecisionInput decodes PAY_PORTION_FULL_PRECISION input bytes.
func DecodePayPortionFullPrecisionInput(data []byte) (*PayPortionFullPrecisionInput, error) {
	vals, err := abi.Unpack(tokenRecipientAmountTypes, data)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: decode PAY_PORTION_FULL_PRECISION: %w", err)
	}

	token, ok := vals[0].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode PAY_PORTION_FULL_PRECISION: unexpected decoded type for token: %T", vals[0])
	}

	recipient, ok := vals[1].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode PAY_PORTION_FULL_PRECISION: unexpected decoded type for recipient: %T", vals[1])
	}

	portion, ok := vals[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode PAY_PORTION_FULL_PRECISION: unexpected decoded type for portion: %T", vals[2])
	}

	return NewPayPortionFullPrecisionInput(token, recipient, portion), nil
}

// DecodeBalanceCheckERC20Input decodes BALANCE_CHECK_ERC20 input bytes.
func DecodeBalanceCheckERC20Input(data []byte) (*BalanceCheckERC20Input, error) {
	vals, err := abi.Unpack(ownerTokenAmountTypes, data)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: decode BALANCE_CHECK_ERC20: %w", err)
	}

	owner, ok := vals[0].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode BALANCE_CHECK_ERC20: unexpected decoded type for owner: %T", vals[0])
	}

	token, ok := vals[1].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode BALANCE_CHECK_ERC20: unexpected decoded type for token: %T", vals[1])
	}

	minBalance, ok := vals[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode BALANCE_CHECK_ERC20: unexpected decoded type for minBalance: %T", vals[2])
	}

	return NewBalanceCheckERC20Input(owner, token, minBalance), nil
}

// DecodePermit2TransferFromInput decodes PERMIT2_TRANSFER_FROM input bytes.
func DecodePermit2TransferFromInput(data []byte) (*Permit2TransferFromInput, error) {
	vals, err := abi.Unpack(permit2TransferFromTypes, data)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: decode PERMIT2_TRANSFER_FROM: %w", err)
	}

	token, ok := vals[0].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode PERMIT2_TRANSFER_FROM: unexpected decoded type for token: %T", vals[0])
	}

	recipient, ok := vals[1].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode PERMIT2_TRANSFER_FROM: unexpected decoded type for recipient: %T", vals[1])
	}

	amount, ok := vals[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode PERMIT2_TRANSFER_FROM: unexpected decoded type for amount: %T", vals[2])
	}

	return NewPermit2TransferFromInput(token, recipient, amount), nil
}

// DecodeWrapETHInput decodes WRAP_ETH input bytes.
func DecodeWrapETHInput(data []byte) (*WrapETHInput, error) {
	vals, err := abi.Unpack(recipientAmountTypes, data)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: decode WRAP_ETH: %w", err)
	}

	recipient, ok := vals[0].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode WRAP_ETH: unexpected decoded type for recipient: %T", vals[0])
	}

	amount, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode WRAP_ETH: unexpected decoded type for amount: %T", vals[1])
	}

	return NewWrapETHInput(recipient, amount), nil
}

// DecodeUnwrapWETHInput decodes UNWRAP_WETH input bytes.
func DecodeUnwrapWETHInput(data []byte) (*UnwrapWETHInput, error) {
	vals, err := abi.Unpack(recipientAmountTypes, data)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: decode UNWRAP_WETH: %w", err)
	}

	recipient, ok := vals[0].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode UNWRAP_WETH: unexpected decoded type for recipient: %T", vals[0])
	}

	amountMin, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode UNWRAP_WETH: unexpected decoded type for amountMin: %T", vals[1])
	}

	return NewUnwrapWETHInput(recipient, amountMin), nil
}

// DecodePermit2PermitInput decodes PERMIT2_PERMIT input bytes.
func DecodePermit2PermitInput(data []byte) (*Permit2PermitInput, error) {
	vals, err := abi.Unpack(permit2PermitTypes, data)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: decode PERMIT2_PERMIT: %w", err)
	}

	single, ok := vals[0].([]any)
	if !ok || len(single) != 3 {
		return nil, fmt.Errorf("universalrouter: decode PERMIT2_PERMIT: unexpected decoded type for permitSingle: %T", vals[0])
	}

	fields, ok := single[0].([]any)
	if !ok || len(fields) != 4 {
		return nil, fmt.Errorf("universalrouter: decode PERMIT2_PERMIT: unexpected decoded type for permitSingle.details: %T", single[0])
	}

	token, ok := fields[0].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode PERMIT2_PERMIT: unexpected decoded type for permitSingle.details.token: %T", fields[0])
	}

	amount, ok := fields[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode PERMIT2_PERMIT: unexpected decoded type for permitSingle.details.amount: %T", fields[1])
	}

	expiration, ok := fields[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode PERMIT2_PERMIT: unexpected decoded type for permitSingle.details.expiration: %T", fields[2])
	}

	nonce, ok := fields[3].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode PERMIT2_PERMIT: unexpected decoded type for permitSingle.details.nonce: %T", fields[3])
	}

	spender, ok := single[1].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode PERMIT2_PERMIT: unexpected decoded type for permitSingle.spender: %T", single[1])
	}

	sigDeadline, ok := single[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode PERMIT2_PERMIT: unexpected decoded type for permitSingle.sigDeadline: %T", single[2])
	}

	signature, ok := vals[1].([]byte)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode PERMIT2_PERMIT: unexpected decoded type for signature: %T", vals[1])
	}

	return NewPermit2PermitInput(permit2.NewPermitSingle(permit2.NewPermitDetails(token, amount, expiration, nonce), spender, sigDeadline), signature), nil
}

// DecodePermit2PermitBatchInput decodes PERMIT2_PERMIT_BATCH input bytes.
func DecodePermit2PermitBatchInput(data []byte) (*Permit2PermitBatchInput, error) {
	vals, err := abi.Unpack(permit2PermitBatchTypes, data)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: decode PERMIT2_PERMIT_BATCH: %w", err)
	}

	batch, ok := vals[0].([]any)
	if !ok || len(batch) != 3 {
		return nil, fmt.Errorf("universalrouter: decode PERMIT2_PERMIT_BATCH: unexpected decoded type for permitBatch: %T", vals[0])
	}

	rawDetails, ok := batch[0].([]any)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode PERMIT2_PERMIT_BATCH: unexpected decoded type for permitBatch.details: %T", batch[0])
	}

	details := make([]permit2.PermitDetails, len(rawDetails))
	for i, v := range rawDetails {
		fields, ok := v.([]any)
		if !ok || len(fields) != 4 {
			return nil, fmt.Errorf("universalrouter: decode PERMIT2_PERMIT_BATCH: unexpected decoded type for permitBatch.details[%d]: %T", i, v)
		}

		token, ok := fields[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("universalrouter: decode PERMIT2_PERMIT_BATCH: unexpected decoded type for permitBatch.details[%d].token: %T", i, fields[0])
		}

		amount, ok := fields[1].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("universalrouter: decode PERMIT2_PERMIT_BATCH: unexpected decoded type for permitBatch.details[%d].amount: %T", i, fields[1])
		}

		expiration, ok := fields[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("universalrouter: decode PERMIT2_PERMIT_BATCH: unexpected decoded type for permitBatch.details[%d].expiration: %T", i, fields[2])
		}

		nonce, ok := fields[3].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("universalrouter: decode PERMIT2_PERMIT_BATCH: unexpected decoded type for permitBatch.details[%d].nonce: %T", i, fields[3])
		}

		details[i] = *permit2.NewPermitDetails(token, amount, expiration, nonce)
	}

	spender, ok := batch[1].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode PERMIT2_PERMIT_BATCH: unexpected decoded type for permitBatch.spender: %T", batch[1])
	}

	sigDeadline, ok := batch[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode PERMIT2_PERMIT_BATCH: unexpected decoded type for permitBatch.sigDeadline: %T", batch[2])
	}

	signature, ok := vals[1].([]byte)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode PERMIT2_PERMIT_BATCH: unexpected decoded type for signature: %T", vals[1])
	}

	return NewPermit2PermitBatchInput(permit2.NewPermitBatch(details, spender, sigDeadline), signature), nil
}

// DecodePermit2TransferFromBatchInput decodes PERMIT2_TRANSFER_FROM_BATCH input bytes.
func DecodePermit2TransferFromBatchInput(data []byte) (*Permit2TransferFromBatchInput, error) {
	vals, err := abi.Unpack(permit2TransferFromBatchTypes, data)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: decode PERMIT2_TRANSFER_FROM_BATCH: %w", err)
	}

	raw, ok := vals[0].([]any)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode PERMIT2_TRANSFER_FROM_BATCH: unexpected decoded type for details: %T", vals[0])
	}

	details := make([]permit2.AllowanceTransferDetails, len(raw))
	for i, v := range raw {
		fields, ok := v.([]any)
		if !ok || len(fields) != 4 {
			return nil, fmt.Errorf("universalrouter: decode PERMIT2_TRANSFER_FROM_BATCH: unexpected decoded type for details[%d]: %T", i, v)
		}

		from, ok := fields[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("universalrouter: decode PERMIT2_TRANSFER_FROM_BATCH: unexpected decoded type for details[%d].from: %T", i, fields[0])
		}

		to, ok := fields[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("universalrouter: decode PERMIT2_TRANSFER_FROM_BATCH: unexpected decoded type for details[%d].to: %T", i, fields[1])
		}

		amount, ok := fields[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("universalrouter: decode PERMIT2_TRANSFER_FROM_BATCH: unexpected decoded type for details[%d].amount: %T", i, fields[2])
		}

		token, ok := fields[3].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("universalrouter: decode PERMIT2_TRANSFER_FROM_BATCH: unexpected decoded type for details[%d].token: %T", i, fields[3])
		}

		details[i] = *permit2.NewAllowanceTransferDetails(from, to, amount, token)
	}

	return NewPermit2TransferFromBatchInput(details), nil
}

// DecodeV4SwapInput decodes V4_SWAP input bytes into V4Router actions and params.
func DecodeV4SwapInput(data []byte) (*V4SwapInput, error) {
	actions, params, err := v4.DecodeActions(data)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: decode V4_SWAP: %w", err)
	}

	return NewV4SwapInput(actions, params), nil
}

// DecodeV3PositionManagerPermitInput decodes V3_POSITION_MANAGER_PERMIT input bytes, which are the raw NFPM calldata.
func DecodeV3PositionManagerPermitInput(data []byte) (*V3PositionManagerPermitInput, error) {
	return NewV3PositionManagerPermitInput(bytes.Clone(data)), nil
}

// DecodeV3PositionManagerCallInput decodes V3_POSITION_MANAGER_CALL input bytes, which are the raw NFPM calldata.
func DecodeV3PositionManagerCallInput(data []byte) (*V3PositionManagerCallInput, error) {
	return NewV3PositionManagerCallInput(bytes.Clone(data)), nil
}

// DecodeV4PositionManagerCallInput decodes V4_POSITION_MANAGER_CALL input bytes, which are the raw PositionManager calldata.
func DecodeV4PositionManagerCallInput(data []byte) (*V4PositionManagerCallInput, error) {
	return NewV4PositionManagerCallInput(bytes.Clone(data)), nil
}

// DecodeV4InitializePoolInput decodes V4_INITIALIZE_POOL input bytes.
func DecodeV4InitializePoolInput(data []byte) (*V4InitializePoolInput, error) {
	vals, err := abi.Unpack(v4InitializePoolTypes, data)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: decode V4_INITIALIZE_POOL: %w", err)
	}

	fields, ok := vals[0].([]any)
	if !ok || len(fields) != 5 {
		return nil, fmt.Errorf("universalrouter: decode V4_INITIALIZE_POOL: unexpected decoded type for key: %T", vals[0])
	}

	currency0, ok := fields[0].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V4_INITIALIZE_POOL: unexpected decoded type for key.currency0: %T", fields[0])
	}

	currency1, ok := fields[1].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V4_INITIALIZE_POOL: unexpected decoded type for key.currency1: %T", fields[1])
	}

	fee, ok := fields[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V4_INITIALIZE_POOL: unexpected decoded type for key.fee: %T", fields[2])
	}

	tickSpacing, ok := fields[3].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V4_INITIALIZE_POOL: unexpected decoded type for key.tickSpacing: %T", fields[3])
	}

	hooks, ok := fields[4].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V4_INITIALIZE_POOL: unexpected decoded type for key.hooks: %T", fields[4])
	}

	sqrtPriceX96, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode V4_INITIALIZE_POOL: unexpected decoded type for sqrtPriceX96: %T", vals[1])
	}

	return NewV4InitializePoolInput(v4.NewPoolKey(currency0, currency1, fee, tickSpacing, hooks), sqrtPriceX96), nil
}

// DecodeExecuteSubPlanInput decodes EXECUTE_SUB_PLAN input bytes, recursing into the nested plan.
func DecodeExecuteSubPlanInput(data []byte) (*ExecuteSubPlanInput, error) {
	vals, err := abi.Unpack(subPlanTypes, data)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: decode EXECUTE_SUB_PLAN: %w", err)
	}

	commands, ok := vals[0].([]byte)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode EXECUTE_SUB_PLAN: unexpected decoded type for commands: %T", vals[0])
	}

	raw, ok := vals[1].([]any)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode EXECUTE_SUB_PLAN: unexpected decoded type for inputs: %T", vals[1])
	}

	inputs := make([][]byte, len(raw))
	for i, v := range raw {
		b, ok := v.([]byte)
		if !ok {
			return nil, fmt.Errorf("universalrouter: decode EXECUTE_SUB_PLAN: unexpected decoded type for inputs[%d]: %T", i, v)
		}
		inputs[i] = b
	}

	plan, err := DecodePlan(commands, inputs)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: decode EXECUTE_SUB_PLAN: %w", err)
	}

	return NewExecuteSubPlanInput(plan), nil
}

// DecodeAcrossV4DepositV3Input decodes ACROSS_V4_DEPOSIT_V3 input bytes.
func DecodeAcrossV4DepositV3Input(data []byte) (*AcrossV4DepositV3Input, error) {
	vals, err := abi.Unpack(acrossV4DepositV3Types, data)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: decode ACROSS_V4_DEPOSIT_V3: %w", err)
	}

	f, ok := vals[0].([]any)
	if !ok || len(f) != 13 {
		return nil, fmt.Errorf("universalrouter: decode ACROSS_V4_DEPOSIT_V3: unexpected decoded type for params: %T", vals[0])
	}

	var addrs [5]*types.Address
	for i, idx := range []int{0, 1, 2, 3, 7} {
		a, ok := f[idx].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("universalrouter: decode ACROSS_V4_DEPOSIT_V3: unexpected decoded type for params[%d]: %T", idx, f[idx])
		}
		addrs[i] = a
	}

	var amounts [3]*big.Int
	for i, idx := range []int{4, 5, 6} {
		n, ok := f[idx].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("universalrouter: decode ACROSS_V4_DEPOSIT_V3: unexpected decoded type for params[%d]: %T", idx, f[idx])
		}
		amounts[i] = n
	}

	var stamps [3]uint32
	for i, idx := range []int{8, 9, 10} {
		n, ok := f[idx].(uint32)
		if !ok {
			return nil, fmt.Errorf("universalrouter: decode ACROSS_V4_DEPOSIT_V3: unexpected decoded type for params[%d]: %T", idx, f[idx])
		}
		stamps[i] = n
	}

	message, ok := f[11].([]byte)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode ACROSS_V4_DEPOSIT_V3: unexpected decoded type for message: %T", f[11])
	}

	useNative, ok := f[12].(bool)
	if !ok {
		return nil, fmt.Errorf("universalrouter: decode ACROSS_V4_DEPOSIT_V3: unexpected decoded type for useNative: %T", f[12])
	}

	return NewAcrossV4DepositV3Input(addrs[0], addrs[1], addrs[2], addrs[3], amounts[0], amounts[1], amounts[2], addrs[4], stamps[0], stamps[1], stamps[2], message, useNative), nil
}
