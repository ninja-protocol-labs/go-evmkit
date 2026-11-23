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

// Input is one command's typed input; Encode validates and returns the bytes the Dispatcher decodes.
type Input interface {
	Command() Command
	Validate() error
	Encode() ([]byte, error)
}

// MsgSender returns ActionConstants.MSG_SENDER, which a recipient resolves to the locker (the execute caller).
func MsgSender() *types.Address {
	return v4.MsgSender()
}

// AddressThis returns ActionConstants.ADDRESS_THIS, which a recipient resolves to the router itself.
func AddressThis() *types.Address {
	return v4.AddressThis()
}

// ContractBalance returns ActionConstants.CONTRACT_BALANCE, which an amount resolves to the router's whole balance.
func ContractBalance() *big.Int {
	return v4.ContractBalance()
}

// V3SwapExactInInput is V3_SWAP_EXACT_IN's input; Path is a packed v3 path and MinHopPriceX36 is empty or one entry per hop.
type V3SwapExactInInput struct {
	Recipient      *types.Address
	AmountIn       *big.Int
	AmountOutMin   *big.Int
	Path           []byte
	PayerIsUser    bool
	MinHopPriceX36 []*big.Int
}

func NewV3SwapExactInInput(recipient *types.Address, amountIn, amountOutMin *big.Int, path []byte, payerIsUser bool, minHopPriceX36 []*big.Int) *V3SwapExactInInput {
	return &V3SwapExactInInput{
		Recipient:      recipient,
		AmountIn:       amountIn,
		AmountOutMin:   amountOutMin,
		Path:           path,
		PayerIsUser:    payerIsUser,
		MinHopPriceX36: minHopPriceX36,
	}
}

func (*V3SwapExactInInput) Command() Command {
	return CommandV3SwapExactIn
}

func (in *V3SwapExactInInput) Validate() error {
	if in == nil {
		return fmt.Errorf("%w: input is nil", bind.ErrInvalidRequest)
	}

	if in.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	if in.AmountIn == nil {
		return fmt.Errorf("%w: amountIn is nil", bind.ErrInvalidRequest)
	}

	if in.AmountOutMin == nil {
		return fmt.Errorf("%w: amountOutMin is nil", bind.ErrInvalidRequest)
	}

	if len(in.Path) < 43 || (len(in.Path)-20)%23 != 0 {
		return fmt.Errorf("%w: path length %d is not 20 + 23*n with n >= 1", bind.ErrInvalidRequest, len(in.Path))
	}

	hops := (len(in.Path) - 20) / 23
	if len(in.MinHopPriceX36) != 0 && len(in.MinHopPriceX36) != hops {
		return fmt.Errorf("%w: minHopPriceX36 length %d does not match %d hops", bind.ErrInvalidRequest, len(in.MinHopPriceX36), hops)
	}

	for i, p := range in.MinHopPriceX36 {
		if p == nil {
			return fmt.Errorf("%w: minHopPriceX36[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

func (in *V3SwapExactInInput) Encode() ([]byte, error) {
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandV3SwapExactIn, err)
	}

	minHopPriceX36 := make([]any, len(in.MinHopPriceX36))
	for i, p := range in.MinHopPriceX36 {
		minHopPriceX36[i] = p
	}

	data, err := abi.Pack(v3SwapTypes, in.Recipient, in.AmountIn, in.AmountOutMin, in.Path, in.PayerIsUser, minHopPriceX36)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandV3SwapExactIn, err)
	}

	return data, nil
}

// V3SwapExactOutInput is V3_SWAP_EXACT_OUT's input; Path is the packed v3 path reversed (output token first).
type V3SwapExactOutInput struct {
	Recipient      *types.Address
	AmountOut      *big.Int
	AmountInMax    *big.Int
	Path           []byte
	PayerIsUser    bool
	MinHopPriceX36 []*big.Int
}

func NewV3SwapExactOutInput(recipient *types.Address, amountOut, amountInMax *big.Int, path []byte, payerIsUser bool, minHopPriceX36 []*big.Int) *V3SwapExactOutInput {
	return &V3SwapExactOutInput{
		Recipient:      recipient,
		AmountOut:      amountOut,
		AmountInMax:    amountInMax,
		Path:           path,
		PayerIsUser:    payerIsUser,
		MinHopPriceX36: minHopPriceX36,
	}
}

func (*V3SwapExactOutInput) Command() Command {
	return CommandV3SwapExactOut
}

func (in *V3SwapExactOutInput) Validate() error {
	if in == nil {
		return fmt.Errorf("%w: input is nil", bind.ErrInvalidRequest)
	}

	if in.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	if in.AmountOut == nil {
		return fmt.Errorf("%w: amountOut is nil", bind.ErrInvalidRequest)
	}

	if in.AmountInMax == nil {
		return fmt.Errorf("%w: amountInMax is nil", bind.ErrInvalidRequest)
	}

	if len(in.Path) < 43 || (len(in.Path)-20)%23 != 0 {
		return fmt.Errorf("%w: path length %d is not 20 + 23*n with n >= 1", bind.ErrInvalidRequest, len(in.Path))
	}

	hops := (len(in.Path) - 20) / 23
	if len(in.MinHopPriceX36) != 0 && len(in.MinHopPriceX36) != hops {
		return fmt.Errorf("%w: minHopPriceX36 length %d does not match %d hops", bind.ErrInvalidRequest, len(in.MinHopPriceX36), hops)
	}

	for i, p := range in.MinHopPriceX36 {
		if p == nil {
			return fmt.Errorf("%w: minHopPriceX36[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

func (in *V3SwapExactOutInput) Encode() ([]byte, error) {
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandV3SwapExactOut, err)
	}

	minHopPriceX36 := make([]any, len(in.MinHopPriceX36))
	for i, p := range in.MinHopPriceX36 {
		minHopPriceX36[i] = p
	}

	data, err := abi.Pack(v3SwapTypes, in.Recipient, in.AmountOut, in.AmountInMax, in.Path, in.PayerIsUser, minHopPriceX36)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandV3SwapExactOut, err)
	}

	return data, nil
}

// Permit2TransferFromInput is PERMIT2_TRANSFER_FROM's input: pull Amount of Token from the locker to Recipient via Permit2.
type Permit2TransferFromInput struct {
	Token     *types.Address
	Recipient *types.Address
	Amount    *big.Int
}

func NewPermit2TransferFromInput(token, recipient *types.Address, amount *big.Int) *Permit2TransferFromInput {
	return &Permit2TransferFromInput{
		Token:     token,
		Recipient: recipient,
		Amount:    amount,
	}
}

func (*Permit2TransferFromInput) Command() Command {
	return CommandPermit2TransferFrom
}

func (in *Permit2TransferFromInput) Validate() error {
	if in == nil {
		return fmt.Errorf("%w: input is nil", bind.ErrInvalidRequest)
	}

	if in.Token == nil {
		return fmt.Errorf("%w: token is nil", bind.ErrInvalidRequest)
	}

	if in.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	if in.Amount == nil {
		return fmt.Errorf("%w: amount is nil", bind.ErrInvalidRequest)
	}

	return nil
}

func (in *Permit2TransferFromInput) Encode() ([]byte, error) {
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandPermit2TransferFrom, err)
	}

	data, err := abi.Pack(permit2TransferFromTypes, in.Token, in.Recipient, in.Amount)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandPermit2TransferFrom, err)
	}

	return data, nil
}

// Permit2PermitBatchInput is PERMIT2_PERMIT_BATCH's input: forward a signed PermitBatch from the locker to Permit2.
type Permit2PermitBatchInput struct {
	PermitBatch *permit2.PermitBatch
	Signature   []byte
}

func NewPermit2PermitBatchInput(permitBatch *permit2.PermitBatch, signature []byte) *Permit2PermitBatchInput {
	return &Permit2PermitBatchInput{
		PermitBatch: permitBatch,
		Signature:   signature,
	}
}

func (*Permit2PermitBatchInput) Command() Command {
	return CommandPermit2PermitBatch
}

func (in *Permit2PermitBatchInput) Validate() error {
	if in == nil {
		return fmt.Errorf("%w: input is nil", bind.ErrInvalidRequest)
	}

	if in.PermitBatch == nil {
		return fmt.Errorf("%w: permitBatch is nil", bind.ErrInvalidRequest)
	}

	if len(in.PermitBatch.Details) == 0 {
		return fmt.Errorf("%w: permitBatch.details is empty", bind.ErrInvalidRequest)
	}

	for i, d := range in.PermitBatch.Details {
		if d.Token == nil {
			return fmt.Errorf("%w: permitBatch.details[%d].token is nil", bind.ErrInvalidRequest, i)
		}

		if d.Amount == nil {
			return fmt.Errorf("%w: permitBatch.details[%d].amount is nil", bind.ErrInvalidRequest, i)
		}

		if d.Expiration == nil {
			return fmt.Errorf("%w: permitBatch.details[%d].expiration is nil", bind.ErrInvalidRequest, i)
		}

		if d.Nonce == nil {
			return fmt.Errorf("%w: permitBatch.details[%d].nonce is nil", bind.ErrInvalidRequest, i)
		}
	}

	if in.PermitBatch.Spender == nil {
		return fmt.Errorf("%w: permitBatch.spender is nil", bind.ErrInvalidRequest)
	}

	if in.PermitBatch.SigDeadline == nil {
		return fmt.Errorf("%w: permitBatch.sigDeadline is nil", bind.ErrInvalidRequest)
	}

	if len(in.Signature) == 0 {
		return fmt.Errorf("%w: signature is empty", bind.ErrInvalidRequest)
	}

	return nil
}

func (in *Permit2PermitBatchInput) Encode() ([]byte, error) {
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandPermit2PermitBatch, err)
	}

	details := make([]any, len(in.PermitBatch.Details))
	for i, d := range in.PermitBatch.Details {
		details[i] = []any{d.Token, d.Amount, d.Expiration, d.Nonce}
	}

	data, err := abi.Pack(permit2PermitBatchTypes, []any{details, in.PermitBatch.Spender, in.PermitBatch.SigDeadline}, in.Signature)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandPermit2PermitBatch, err)
	}

	return data, nil
}

// SweepInput is SWEEP's input: send the router's whole Token balance (zero address for ETH) to Recipient, reverting below AmountMin.
type SweepInput struct {
	Token     *types.Address
	Recipient *types.Address
	AmountMin *big.Int
}

func NewSweepInput(token, recipient *types.Address, amountMin *big.Int) *SweepInput {
	return &SweepInput{
		Token:     token,
		Recipient: recipient,
		AmountMin: amountMin,
	}
}

func (*SweepInput) Command() Command {
	return CommandSweep
}

func (in *SweepInput) Validate() error {
	if in == nil {
		return fmt.Errorf("%w: input is nil", bind.ErrInvalidRequest)
	}

	if in.Token == nil {
		return fmt.Errorf("%w: token is nil", bind.ErrInvalidRequest)
	}

	if in.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	if in.AmountMin == nil {
		return fmt.Errorf("%w: amountMin is nil", bind.ErrInvalidRequest)
	}

	if in.AmountMin.BitLen() > 160 {
		return fmt.Errorf("%w: amountMin %s exceeds uint160", bind.ErrInvalidRequest, in.AmountMin)
	}

	return nil
}

func (in *SweepInput) Encode() ([]byte, error) {
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandSweep, err)
	}

	data, err := abi.Pack(tokenRecipientAmountTypes, in.Token, in.Recipient, in.AmountMin)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandSweep, err)
	}

	return data, nil
}

// TransferInput is TRANSFER's input: send Value of the router's Token (zero address for ETH) to Recipient.
type TransferInput struct {
	Token     *types.Address
	Recipient *types.Address
	Value     *big.Int
}

func NewTransferInput(token, recipient *types.Address, value *big.Int) *TransferInput {
	return &TransferInput{
		Token:     token,
		Recipient: recipient,
		Value:     value,
	}
}

func (*TransferInput) Command() Command {
	return CommandTransfer
}

func (in *TransferInput) Validate() error {
	if in == nil {
		return fmt.Errorf("%w: input is nil", bind.ErrInvalidRequest)
	}

	if in.Token == nil {
		return fmt.Errorf("%w: token is nil", bind.ErrInvalidRequest)
	}

	if in.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	if in.Value == nil {
		return fmt.Errorf("%w: value is nil", bind.ErrInvalidRequest)
	}

	return nil
}

func (in *TransferInput) Encode() ([]byte, error) {
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandTransfer, err)
	}

	data, err := abi.Pack(tokenRecipientAmountTypes, in.Token, in.Recipient, in.Value)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandTransfer, err)
	}

	return data, nil
}

// PayPortionInput is PAY_PORTION's input: send Bips/10000 of the router's Token balance to Recipient.
type PayPortionInput struct {
	Token     *types.Address
	Recipient *types.Address
	Bips      *big.Int
}

func NewPayPortionInput(token, recipient *types.Address, bips *big.Int) *PayPortionInput {
	return &PayPortionInput{
		Token:     token,
		Recipient: recipient,
		Bips:      bips,
	}
}

func (*PayPortionInput) Command() Command {
	return CommandPayPortion
}

func (in *PayPortionInput) Validate() error {
	if in == nil {
		return fmt.Errorf("%w: input is nil", bind.ErrInvalidRequest)
	}

	if in.Token == nil {
		return fmt.Errorf("%w: token is nil", bind.ErrInvalidRequest)
	}

	if in.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	if in.Bips == nil {
		return fmt.Errorf("%w: bips is nil", bind.ErrInvalidRequest)
	}

	if in.Bips.Cmp(big.NewInt(BipsDenominator)) > 0 {
		return fmt.Errorf("%w: bips %s exceeds %d", bind.ErrInvalidRequest, in.Bips, BipsDenominator)
	}

	return nil
}

func (in *PayPortionInput) Encode() ([]byte, error) {
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandPayPortion, err)
	}

	data, err := abi.Pack(tokenRecipientAmountTypes, in.Token, in.Recipient, in.Bips)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandPayPortion, err)
	}

	return data, nil
}

// PayPortionFullPrecisionInput is PAY_PORTION_FULL_PRECISION's input: send Portion/1e18 of the router's Token balance to Recipient.
type PayPortionFullPrecisionInput struct {
	Token     *types.Address
	Recipient *types.Address
	Portion   *big.Int
}

func NewPayPortionFullPrecisionInput(token, recipient *types.Address, portion *big.Int) *PayPortionFullPrecisionInput {
	return &PayPortionFullPrecisionInput{
		Token:     token,
		Recipient: recipient,
		Portion:   portion,
	}
}

func (*PayPortionFullPrecisionInput) Command() Command {
	return CommandPayPortionFullPrecision
}

func (in *PayPortionFullPrecisionInput) Validate() error {
	if in == nil {
		return fmt.Errorf("%w: input is nil", bind.ErrInvalidRequest)
	}

	if in.Token == nil {
		return fmt.Errorf("%w: token is nil", bind.ErrInvalidRequest)
	}

	if in.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	if in.Portion == nil {
		return fmt.Errorf("%w: portion is nil", bind.ErrInvalidRequest)
	}

	if in.Portion.Cmp(big.NewInt(FullPrecisionPortionDenominator)) > 0 {
		return fmt.Errorf("%w: portion %s exceeds %d", bind.ErrInvalidRequest, in.Portion, int64(FullPrecisionPortionDenominator))
	}

	return nil
}

func (in *PayPortionFullPrecisionInput) Encode() ([]byte, error) {
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandPayPortionFullPrecision, err)
	}

	data, err := abi.Pack(tokenRecipientAmountTypes, in.Token, in.Recipient, in.Portion)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandPayPortionFullPrecision, err)
	}

	return data, nil
}

// V2SwapExactInInput is V2_SWAP_EXACT_IN's input; Path lists at least two tokens and MinHopPriceX36 is empty or len(Path)-1 long.
type V2SwapExactInInput struct {
	Recipient      *types.Address
	AmountIn       *big.Int
	AmountOutMin   *big.Int
	Path           []*types.Address
	PayerIsUser    bool
	MinHopPriceX36 []*big.Int
}

func NewV2SwapExactInInput(recipient *types.Address, amountIn, amountOutMin *big.Int, path []*types.Address, payerIsUser bool, minHopPriceX36 []*big.Int) *V2SwapExactInInput {
	return &V2SwapExactInInput{
		Recipient:      recipient,
		AmountIn:       amountIn,
		AmountOutMin:   amountOutMin,
		Path:           path,
		PayerIsUser:    payerIsUser,
		MinHopPriceX36: minHopPriceX36,
	}
}

func (*V2SwapExactInInput) Command() Command {
	return CommandV2SwapExactIn
}

func (in *V2SwapExactInInput) Validate() error {
	if in == nil {
		return fmt.Errorf("%w: input is nil", bind.ErrInvalidRequest)
	}

	if in.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	if in.AmountIn == nil {
		return fmt.Errorf("%w: amountIn is nil", bind.ErrInvalidRequest)
	}

	if in.AmountOutMin == nil {
		return fmt.Errorf("%w: amountOutMin is nil", bind.ErrInvalidRequest)
	}

	if len(in.Path) < 2 {
		return fmt.Errorf("%w: path needs at least 2 tokens, got %d", bind.ErrInvalidRequest, len(in.Path))
	}

	for i, a := range in.Path {
		if a == nil {
			return fmt.Errorf("%w: path[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	if len(in.MinHopPriceX36) != 0 && len(in.MinHopPriceX36) != len(in.Path)-1 {
		return fmt.Errorf("%w: minHopPriceX36 length %d does not match %d hops", bind.ErrInvalidRequest, len(in.MinHopPriceX36), len(in.Path)-1)
	}

	for i, p := range in.MinHopPriceX36 {
		if p == nil {
			return fmt.Errorf("%w: minHopPriceX36[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

func (in *V2SwapExactInInput) Encode() ([]byte, error) {
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandV2SwapExactIn, err)
	}

	path := make([]any, len(in.Path))
	for i, a := range in.Path {
		path[i] = a
	}

	minHopPriceX36 := make([]any, len(in.MinHopPriceX36))
	for i, p := range in.MinHopPriceX36 {
		minHopPriceX36[i] = p
	}

	data, err := abi.Pack(v2SwapTypes, in.Recipient, in.AmountIn, in.AmountOutMin, path, in.PayerIsUser, minHopPriceX36)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandV2SwapExactIn, err)
	}

	return data, nil
}

// V2SwapExactOutInput is V2_SWAP_EXACT_OUT's input; Path is in swap order (input token first).
type V2SwapExactOutInput struct {
	Recipient      *types.Address
	AmountOut      *big.Int
	AmountInMax    *big.Int
	Path           []*types.Address
	PayerIsUser    bool
	MinHopPriceX36 []*big.Int
}

func NewV2SwapExactOutInput(recipient *types.Address, amountOut, amountInMax *big.Int, path []*types.Address, payerIsUser bool, minHopPriceX36 []*big.Int) *V2SwapExactOutInput {
	return &V2SwapExactOutInput{
		Recipient:      recipient,
		AmountOut:      amountOut,
		AmountInMax:    amountInMax,
		Path:           path,
		PayerIsUser:    payerIsUser,
		MinHopPriceX36: minHopPriceX36,
	}
}

func (*V2SwapExactOutInput) Command() Command {
	return CommandV2SwapExactOut
}

func (in *V2SwapExactOutInput) Validate() error {
	if in == nil {
		return fmt.Errorf("%w: input is nil", bind.ErrInvalidRequest)
	}

	if in.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	if in.AmountOut == nil {
		return fmt.Errorf("%w: amountOut is nil", bind.ErrInvalidRequest)
	}

	if in.AmountInMax == nil {
		return fmt.Errorf("%w: amountInMax is nil", bind.ErrInvalidRequest)
	}

	if len(in.Path) < 2 {
		return fmt.Errorf("%w: path needs at least 2 tokens, got %d", bind.ErrInvalidRequest, len(in.Path))
	}

	for i, a := range in.Path {
		if a == nil {
			return fmt.Errorf("%w: path[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	if len(in.MinHopPriceX36) != 0 && len(in.MinHopPriceX36) != len(in.Path)-1 {
		return fmt.Errorf("%w: minHopPriceX36 length %d does not match %d hops", bind.ErrInvalidRequest, len(in.MinHopPriceX36), len(in.Path)-1)
	}

	for i, p := range in.MinHopPriceX36 {
		if p == nil {
			return fmt.Errorf("%w: minHopPriceX36[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

func (in *V2SwapExactOutInput) Encode() ([]byte, error) {
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandV2SwapExactOut, err)
	}

	path := make([]any, len(in.Path))
	for i, a := range in.Path {
		path[i] = a
	}

	minHopPriceX36 := make([]any, len(in.MinHopPriceX36))
	for i, p := range in.MinHopPriceX36 {
		minHopPriceX36[i] = p
	}

	data, err := abi.Pack(v2SwapTypes, in.Recipient, in.AmountOut, in.AmountInMax, path, in.PayerIsUser, minHopPriceX36)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandV2SwapExactOut, err)
	}

	return data, nil
}

// Permit2PermitInput is PERMIT2_PERMIT's input: forward a signed PermitSingle from the locker to Permit2.
type Permit2PermitInput struct {
	PermitSingle *permit2.PermitSingle
	Signature    []byte
}

func NewPermit2PermitInput(permitSingle *permit2.PermitSingle, signature []byte) *Permit2PermitInput {
	return &Permit2PermitInput{
		PermitSingle: permitSingle,
		Signature:    signature,
	}
}

func (*Permit2PermitInput) Command() Command {
	return CommandPermit2Permit
}

func (in *Permit2PermitInput) Validate() error {
	if in == nil {
		return fmt.Errorf("%w: input is nil", bind.ErrInvalidRequest)
	}

	if in.PermitSingle == nil {
		return fmt.Errorf("%w: permitSingle is nil", bind.ErrInvalidRequest)
	}

	if in.PermitSingle.Details == nil {
		return fmt.Errorf("%w: permitSingle.details is nil", bind.ErrInvalidRequest)
	}

	if in.PermitSingle.Details.Token == nil {
		return fmt.Errorf("%w: permitSingle.details.token is nil", bind.ErrInvalidRequest)
	}

	if in.PermitSingle.Details.Amount == nil {
		return fmt.Errorf("%w: permitSingle.details.amount is nil", bind.ErrInvalidRequest)
	}

	if in.PermitSingle.Details.Expiration == nil {
		return fmt.Errorf("%w: permitSingle.details.expiration is nil", bind.ErrInvalidRequest)
	}

	if in.PermitSingle.Details.Nonce == nil {
		return fmt.Errorf("%w: permitSingle.details.nonce is nil", bind.ErrInvalidRequest)
	}

	if in.PermitSingle.Spender == nil {
		return fmt.Errorf("%w: permitSingle.spender is nil", bind.ErrInvalidRequest)
	}

	if in.PermitSingle.SigDeadline == nil {
		return fmt.Errorf("%w: permitSingle.sigDeadline is nil", bind.ErrInvalidRequest)
	}

	if len(in.Signature) == 0 {
		return fmt.Errorf("%w: signature is empty", bind.ErrInvalidRequest)
	}

	return nil
}

func (in *Permit2PermitInput) Encode() ([]byte, error) {
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandPermit2Permit, err)
	}

	d := in.PermitSingle.Details
	data, err := abi.Pack(permit2PermitTypes, []any{[]any{d.Token, d.Amount, d.Expiration, d.Nonce}, in.PermitSingle.Spender, in.PermitSingle.SigDeadline}, in.Signature)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandPermit2Permit, err)
	}

	return data, nil
}

// WrapETHInput is WRAP_ETH's input: wrap Amount of the router's ETH (or ContractBalance) into WETH for Recipient.
type WrapETHInput struct {
	Recipient *types.Address
	Amount    *big.Int
}

func NewWrapETHInput(recipient *types.Address, amount *big.Int) *WrapETHInput {
	return &WrapETHInput{
		Recipient: recipient,
		Amount:    amount,
	}
}

func (*WrapETHInput) Command() Command {
	return CommandWrapETH
}

func (in *WrapETHInput) Validate() error {
	if in == nil {
		return fmt.Errorf("%w: input is nil", bind.ErrInvalidRequest)
	}

	if in.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	if in.Amount == nil {
		return fmt.Errorf("%w: amount is nil", bind.ErrInvalidRequest)
	}

	return nil
}

func (in *WrapETHInput) Encode() ([]byte, error) {
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandWrapETH, err)
	}

	data, err := abi.Pack(recipientAmountTypes, in.Recipient, in.Amount)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandWrapETH, err)
	}

	return data, nil
}

// UnwrapWETHInput is UNWRAP_WETH's input: unwrap the router's whole WETH balance to Recipient, reverting below AmountMin.
type UnwrapWETHInput struct {
	Recipient *types.Address
	AmountMin *big.Int
}

func NewUnwrapWETHInput(recipient *types.Address, amountMin *big.Int) *UnwrapWETHInput {
	return &UnwrapWETHInput{
		Recipient: recipient,
		AmountMin: amountMin,
	}
}

func (*UnwrapWETHInput) Command() Command {
	return CommandUnwrapWETH
}

func (in *UnwrapWETHInput) Validate() error {
	if in == nil {
		return fmt.Errorf("%w: input is nil", bind.ErrInvalidRequest)
	}

	if in.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	if in.AmountMin == nil {
		return fmt.Errorf("%w: amountMin is nil", bind.ErrInvalidRequest)
	}

	return nil
}

func (in *UnwrapWETHInput) Encode() ([]byte, error) {
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandUnwrapWETH, err)
	}

	data, err := abi.Pack(recipientAmountTypes, in.Recipient, in.AmountMin)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandUnwrapWETH, err)
	}

	return data, nil
}

// Permit2TransferFromBatchInput is PERMIT2_TRANSFER_FROM_BATCH's input; every Details[i].From must be the locker.
type Permit2TransferFromBatchInput struct {
	Details []permit2.AllowanceTransferDetails
}

func NewPermit2TransferFromBatchInput(details []permit2.AllowanceTransferDetails) *Permit2TransferFromBatchInput {
	return &Permit2TransferFromBatchInput{
		Details: details,
	}
}

func (*Permit2TransferFromBatchInput) Command() Command {
	return CommandPermit2TransferFromBatch
}

func (in *Permit2TransferFromBatchInput) Validate() error {
	if in == nil {
		return fmt.Errorf("%w: input is nil", bind.ErrInvalidRequest)
	}

	if len(in.Details) == 0 {
		return fmt.Errorf("%w: details is empty", bind.ErrInvalidRequest)
	}

	for i, d := range in.Details {
		if d.From == nil {
			return fmt.Errorf("%w: details[%d].from is nil", bind.ErrInvalidRequest, i)
		}

		if d.To == nil {
			return fmt.Errorf("%w: details[%d].to is nil", bind.ErrInvalidRequest, i)
		}

		if d.Amount == nil {
			return fmt.Errorf("%w: details[%d].amount is nil", bind.ErrInvalidRequest, i)
		}

		if d.Token == nil {
			return fmt.Errorf("%w: details[%d].token is nil", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

func (in *Permit2TransferFromBatchInput) Encode() ([]byte, error) {
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandPermit2TransferFromBatch, err)
	}

	details := make([]any, len(in.Details))
	for i, d := range in.Details {
		details[i] = []any{d.From, d.To, d.Amount, d.Token}
	}

	data, err := abi.Pack(permit2TransferFromBatchTypes, details)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandPermit2TransferFromBatch, err)
	}

	return data, nil
}

// BalanceCheckERC20Input is BALANCE_CHECK_ERC20's input: fail (BalanceTooLow) unless Owner holds at least MinBalance of Token.
type BalanceCheckERC20Input struct {
	Owner      *types.Address
	Token      *types.Address
	MinBalance *big.Int
}

func NewBalanceCheckERC20Input(owner, token *types.Address, minBalance *big.Int) *BalanceCheckERC20Input {
	return &BalanceCheckERC20Input{
		Owner:      owner,
		Token:      token,
		MinBalance: minBalance,
	}
}

func (*BalanceCheckERC20Input) Command() Command {
	return CommandBalanceCheckERC20
}

func (in *BalanceCheckERC20Input) Validate() error {
	if in == nil {
		return fmt.Errorf("%w: input is nil", bind.ErrInvalidRequest)
	}

	if in.Owner == nil {
		return fmt.Errorf("%w: owner is nil", bind.ErrInvalidRequest)
	}

	if in.Token == nil {
		return fmt.Errorf("%w: token is nil", bind.ErrInvalidRequest)
	}

	if in.MinBalance == nil {
		return fmt.Errorf("%w: minBalance is nil", bind.ErrInvalidRequest)
	}

	return nil
}

func (in *BalanceCheckERC20Input) Encode() ([]byte, error) {
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandBalanceCheckERC20, err)
	}

	data, err := abi.Pack(ownerTokenAmountTypes, in.Owner, in.Token, in.MinBalance)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandBalanceCheckERC20, err)
	}

	return data, nil
}

// V4SwapInput is V4_SWAP's input: V4Router actions and their params, encoded with the v4 package's EncodeAction* helpers.
type V4SwapInput struct {
	Actions []byte
	Params  [][]byte
}

func NewV4SwapInput(actions []byte, params [][]byte) *V4SwapInput {
	return &V4SwapInput{
		Actions: actions,
		Params:  params,
	}
}

func (*V4SwapInput) Command() Command {
	return CommandV4Swap
}

func (in *V4SwapInput) Validate() error {
	if in == nil {
		return fmt.Errorf("%w: input is nil", bind.ErrInvalidRequest)
	}

	if len(in.Actions) == 0 {
		return fmt.Errorf("%w: actions is empty", bind.ErrInvalidRequest)
	}

	if len(in.Params) != len(in.Actions) {
		return fmt.Errorf("%w: params length %d does not match actions length %d", bind.ErrInvalidRequest, len(in.Params), len(in.Actions))
	}

	return nil
}

func (in *V4SwapInput) Encode() ([]byte, error) {
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandV4Swap, err)
	}

	data, err := v4.EncodeActions(in.Actions, in.Params)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandV4Swap, err)
	}

	return data, nil
}

// V3PositionManagerPermitInput is V3_POSITION_MANAGER_PERMIT's input: raw calldata of the v3 NFPM's ERC721 permit.
type V3PositionManagerPermitInput struct {
	Calldata []byte
}

func NewV3PositionManagerPermitInput(calldata []byte) *V3PositionManagerPermitInput {
	return &V3PositionManagerPermitInput{
		Calldata: calldata,
	}
}

func (*V3PositionManagerPermitInput) Command() Command {
	return CommandV3PositionManagerPermit
}

func (in *V3PositionManagerPermitInput) Validate() error {
	if in == nil {
		return fmt.Errorf("%w: input is nil", bind.ErrInvalidRequest)
	}

	if len(in.Calldata) < 4 {
		return fmt.Errorf("%w: calldata is shorter than a selector", bind.ErrInvalidRequest)
	}

	if !bytes.Equal(in.Calldata[:4], v3PermitSelector[:]) {
		return fmt.Errorf("%w: calldata selector %x is not permit", bind.ErrInvalidRequest, in.Calldata[:4])
	}

	return nil
}

func (in *V3PositionManagerPermitInput) Encode() ([]byte, error) {
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandV3PositionManagerPermit, err)
	}

	return bytes.Clone(in.Calldata), nil
}

// V3PositionManagerCallInput is V3_POSITION_MANAGER_CALL's input: raw calldata of the v3 NFPM's decreaseLiquidity, collect or burn.
type V3PositionManagerCallInput struct {
	Calldata []byte
}

func NewV3PositionManagerCallInput(calldata []byte) *V3PositionManagerCallInput {
	return &V3PositionManagerCallInput{
		Calldata: calldata,
	}
}

func (*V3PositionManagerCallInput) Command() Command {
	return CommandV3PositionManagerCall
}

func (in *V3PositionManagerCallInput) Validate() error {
	if in == nil {
		return fmt.Errorf("%w: input is nil", bind.ErrInvalidRequest)
	}

	if len(in.Calldata) < 4+32 {
		return fmt.Errorf("%w: calldata is shorter than a selector and tokenId", bind.ErrInvalidRequest)
	}

	sel := in.Calldata[:4]
	if !bytes.Equal(sel, v3DecreaseLiquiditySelector[:]) && !bytes.Equal(sel, v3CollectSelector[:]) && !bytes.Equal(sel, v3BurnSelector[:]) {
		return fmt.Errorf("%w: calldata selector %x is not decreaseLiquidity, collect or burn", bind.ErrInvalidRequest, sel)
	}

	return nil
}

func (in *V3PositionManagerCallInput) Encode() ([]byte, error) {
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandV3PositionManagerCall, err)
	}

	return bytes.Clone(in.Calldata), nil
}

// V4InitializePoolInput is V4_INITIALIZE_POOL's input: PoolManager.initialize(Key, SqrtPriceX96), allowed to fail if already initialized.
type V4InitializePoolInput struct {
	Key          *v4.PoolKey
	SqrtPriceX96 *big.Int
}

func NewV4InitializePoolInput(key *v4.PoolKey, sqrtPriceX96 *big.Int) *V4InitializePoolInput {
	return &V4InitializePoolInput{
		Key:          key,
		SqrtPriceX96: sqrtPriceX96,
	}
}

func (*V4InitializePoolInput) Command() Command {
	return CommandV4InitializePool
}

func (in *V4InitializePoolInput) Validate() error {
	if in == nil {
		return fmt.Errorf("%w: input is nil", bind.ErrInvalidRequest)
	}

	if in.Key == nil {
		return fmt.Errorf("%w: key is nil", bind.ErrInvalidRequest)
	}

	if in.Key.Currency0 == nil {
		return fmt.Errorf("%w: key.currency0 is nil", bind.ErrInvalidRequest)
	}

	if in.Key.Currency1 == nil {
		return fmt.Errorf("%w: key.currency1 is nil", bind.ErrInvalidRequest)
	}

	if in.Key.Fee == nil {
		return fmt.Errorf("%w: key.fee is nil", bind.ErrInvalidRequest)
	}

	if in.Key.TickSpacing == nil {
		return fmt.Errorf("%w: key.tickSpacing is nil", bind.ErrInvalidRequest)
	}

	if in.Key.Hooks == nil {
		return fmt.Errorf("%w: key.hooks is nil", bind.ErrInvalidRequest)
	}

	if in.SqrtPriceX96 == nil {
		return fmt.Errorf("%w: sqrtPriceX96 is nil", bind.ErrInvalidRequest)
	}

	return nil
}

func (in *V4InitializePoolInput) Encode() ([]byte, error) {
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandV4InitializePool, err)
	}

	k := in.Key
	data, err := abi.Pack(v4InitializePoolTypes, []any{k.Currency0, k.Currency1, k.Fee, k.TickSpacing, k.Hooks}, in.SqrtPriceX96)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandV4InitializePool, err)
	}

	return data, nil
}

// V4PositionManagerCallInput is V4_POSITION_MANAGER_CALL's input: raw modifyLiquidities calldata whose actions never increase, decrease or burn.
type V4PositionManagerCallInput struct {
	Calldata []byte
}

func NewV4PositionManagerCallInput(calldata []byte) *V4PositionManagerCallInput {
	return &V4PositionManagerCallInput{
		Calldata: calldata,
	}
}

func (*V4PositionManagerCallInput) Command() Command {
	return CommandV4PositionManagerCall
}

func (in *V4PositionManagerCallInput) Validate() error {
	if in == nil {
		return fmt.Errorf("%w: input is nil", bind.ErrInvalidRequest)
	}

	if len(in.Calldata) < 4 {
		return fmt.Errorf("%w: calldata is shorter than a selector", bind.ErrInvalidRequest)
	}

	sel := v4ModifyLiquiditiesFn.Selector()
	if !bytes.Equal(in.Calldata[:4], sel[:]) {
		return fmt.Errorf("%w: calldata selector %x is not modifyLiquidities", bind.ErrInvalidRequest, in.Calldata[:4])
	}

	vals, err := v4ModifyLiquiditiesFn.DecodeCall(in.Calldata)
	if err != nil {
		return fmt.Errorf("%w: calldata: %w", bind.ErrInvalidRequest, err)
	}

	unlockData, ok := vals[0].([]byte)
	if !ok {
		return fmt.Errorf("%w: calldata: unexpected decoded type for unlockData: %T", bind.ErrInvalidRequest, vals[0])
	}

	actions, _, err := v4.DecodeActions(unlockData)
	if err != nil {
		return fmt.Errorf("%w: calldata: %w", bind.ErrInvalidRequest, err)
	}

	for i, a := range actions {
		switch a {
		case v4.ActionIncreaseLiquidity, v4.ActionIncreaseLiquidityFromDeltas, v4.ActionDecreaseLiquidity, v4.ActionBurnPosition:
			return fmt.Errorf("%w: calldata actions[%d] 0x%02x is not allowed, only mint", bind.ErrInvalidRequest, i, a)
		}
	}

	return nil
}

func (in *V4PositionManagerCallInput) Encode() ([]byte, error) {
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandV4PositionManagerCall, err)
	}

	return bytes.Clone(in.Calldata), nil
}

// ExecuteSubPlanInput is EXECUTE_SUB_PLAN's input: a nested plan the router executes as a self-call.
type ExecuteSubPlanInput struct {
	Plan *Plan
}

func NewExecuteSubPlanInput(plan *Plan) *ExecuteSubPlanInput {
	return &ExecuteSubPlanInput{
		Plan: plan,
	}
}

func (*ExecuteSubPlanInput) Command() Command {
	return CommandExecuteSubPlan
}

func (in *ExecuteSubPlanInput) Validate() error {
	if in == nil {
		return fmt.Errorf("%w: input is nil", bind.ErrInvalidRequest)
	}

	if in.Plan == nil {
		return fmt.Errorf("%w: plan is nil", bind.ErrInvalidRequest)
	}

	if err := in.Plan.Validate(); err != nil {
		return fmt.Errorf("plan: %w", err)
	}

	return nil
}

func (in *ExecuteSubPlanInput) Encode() ([]byte, error) {
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandExecuteSubPlan, err)
	}

	commands, inputs, err := in.Plan.Encode()
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandExecuteSubPlan, err)
	}

	args := make([]any, len(inputs))
	for i, b := range inputs {
		args[i] = b
	}

	data, err := abi.Pack(subPlanTypes, commands, args)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandExecuteSubPlan, err)
	}

	return data, nil
}

// AcrossV4DepositV3Input is ACROSS_V4_DEPOSIT_V3's input (AcrossV4DepositV3Params); InputAmount may be ContractBalance.
type AcrossV4DepositV3Input struct {
	Depositor           *types.Address
	Recipient           *types.Address
	InputToken          *types.Address
	OutputToken         *types.Address
	InputAmount         *big.Int
	OutputAmount        *big.Int
	DestinationChainID  *big.Int
	ExclusiveRelayer    *types.Address
	QuoteTimestamp      uint32
	FillDeadline        uint32
	ExclusivityDeadline uint32
	Message             []byte
	UseNative           bool
}

func NewAcrossV4DepositV3Input(depositor, recipient, inputToken, outputToken *types.Address, inputAmount, outputAmount, destinationChainID *big.Int, exclusiveRelayer *types.Address, quoteTimestamp, fillDeadline, exclusivityDeadline uint32, message []byte, useNative bool) *AcrossV4DepositV3Input {
	return &AcrossV4DepositV3Input{
		Depositor:           depositor,
		Recipient:           recipient,
		InputToken:          inputToken,
		OutputToken:         outputToken,
		InputAmount:         inputAmount,
		OutputAmount:        outputAmount,
		DestinationChainID:  destinationChainID,
		ExclusiveRelayer:    exclusiveRelayer,
		QuoteTimestamp:      quoteTimestamp,
		FillDeadline:        fillDeadline,
		ExclusivityDeadline: exclusivityDeadline,
		Message:             message,
		UseNative:           useNative,
	}
}

func (*AcrossV4DepositV3Input) Command() Command {
	return CommandAcrossV4DepositV3
}

func (in *AcrossV4DepositV3Input) Validate() error {
	if in == nil {
		return fmt.Errorf("%w: input is nil", bind.ErrInvalidRequest)
	}

	if in.Depositor == nil {
		return fmt.Errorf("%w: depositor is nil", bind.ErrInvalidRequest)
	}

	if in.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	if in.InputToken == nil {
		return fmt.Errorf("%w: inputToken is nil", bind.ErrInvalidRequest)
	}

	if in.OutputToken == nil {
		return fmt.Errorf("%w: outputToken is nil", bind.ErrInvalidRequest)
	}

	if in.InputAmount == nil {
		return fmt.Errorf("%w: inputAmount is nil", bind.ErrInvalidRequest)
	}

	if in.OutputAmount == nil {
		return fmt.Errorf("%w: outputAmount is nil", bind.ErrInvalidRequest)
	}

	if in.DestinationChainID == nil {
		return fmt.Errorf("%w: destinationChainID is nil", bind.ErrInvalidRequest)
	}

	if in.ExclusiveRelayer == nil {
		return fmt.Errorf("%w: exclusiveRelayer is nil", bind.ErrInvalidRequest)
	}

	return nil
}

func (in *AcrossV4DepositV3Input) Encode() ([]byte, error) {
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandAcrossV4DepositV3, err)
	}

	message := in.Message
	if message == nil {
		message = []byte{}
	}

	data, err := abi.Pack(acrossV4DepositV3Types, []any{
		in.Depositor, in.Recipient, in.InputToken, in.OutputToken, in.InputAmount, in.OutputAmount, in.DestinationChainID,
		in.ExclusiveRelayer, in.QuoteTimestamp, in.FillDeadline, in.ExclusivityDeadline, message, in.UseNative,
	})
	if err != nil {
		return nil, fmt.Errorf("universalrouter: encode %s: %w", CommandAcrossV4DepositV3, err)
	}

	return data, nil
}
