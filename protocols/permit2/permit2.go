package permit2

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

// DefaultPermit2Address is Permit2's canonical CREATE2 address, identical on nearly every EVM chain.
var DefaultPermit2Address = types.MustNewAddressFromHex("0x000000000022D473030F116dDEE9F6B43aC78BA3")

// IEIP712 is DOMAIN_SEPARATOR() -- both IAllowanceTransfer and ISignatureTransfer inherit it in
// Permit2.sol, each from their own "is IEIP712" declaration.
type IEIP712 interface {
	DomainSeparator(ctx context.Context, contract *types.Address, block string) (*types.Hash, error)
}

type IAllowanceTransfer interface {
	IEIP712

	Allowance(ctx context.Context, contract, owner, token, spender *types.Address, block string) (*PackedAllowance, error)

	Approve(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	PermitSingle(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	PermitBatch(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	TransferFrom(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	TransferFromBatch(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Lockdown(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	InvalidateNonces(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

type ISignatureTransfer interface {
	IEIP712

	NonceBitmap(ctx context.Context, contract, owner *types.Address, wordPos *big.Int, block string) (*big.Int, error)

	PermitTransferFrom(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	PermitWitnessTransferFrom(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	PermitTransferFromBatch(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	PermitWitnessTransferFromBatch(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	InvalidateUnorderedNonces(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

type IPermit2 interface {
	IAllowanceTransfer
	ISignatureTransfer
}

var _ IPermit2 = (*Permit2)(nil)

type Permit2 struct {
	c rpc.Client
}

func New(c rpc.Client) *Permit2 {
	return &Permit2{
		c: c,
	}
}

// DomainSeparator returns contract's EIP-712 domain separator.
func (p *Permit2) DomainSeparator(ctx context.Context, contract *types.Address, block string) (*types.Hash, error) {
	data, err := p.call(ctx, nil, contract, EncodeDomainSeparator(), block)
	if err != nil {
		return nil, fmt.Errorf("permit2: domainSeparator: %w", err)
	}
	h, err := DecodeDomainSeparator(data)
	if err != nil {
		return nil, fmt.Errorf("permit2: domainSeparator: %w", err)
	}
	return h, nil
}

// Allowance returns owner's current allowance of token for spender.
func (p *Permit2) Allowance(ctx context.Context, contract, owner, token, spender *types.Address, block string) (*PackedAllowance, error) {
	calldata, err := EncodeAllowance(owner, token, spender)
	if err != nil {
		return nil, fmt.Errorf("permit2: allowance: %w", err)
	}
	data, err := p.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("permit2: allowance: %w", err)
	}
	a, err := DecodeAllowance(data)
	if err != nil {
		return nil, fmt.Errorf("permit2: allowance: %w", err)
	}
	return a, nil
}

// CreateApprove builds an unsigned approve(token, spender, amount, expiration) transaction via b.
func (p *Permit2) CreateApprove(ctx context.Context, b tx.Builder, contract, sender, token, spender *types.Address, amount, expiration *big.Int) (core.Transaction, error) {
	calldata, err := EncodeApprove(token, spender, amount, expiration)
	if err != nil {
		return nil, fmt.Errorf("permit2: createApprove: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("permit2: createApprove: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// Approve broadcasts an already-signed approve transaction.
func (p *Permit2) Approve(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("permit2: approve: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("permit2: approve: %w", err)
	}
	return hash, nil
}

// CreatePermitSingle builds an unsigned permit(owner, permitSingle, signature) transaction via b.
func (p *Permit2) CreatePermitSingle(ctx context.Context, b tx.Builder, contract, sender, owner *types.Address, permitSingle *PermitSingle, sig *types.Signature) (core.Transaction, error) {
	calldata, err := EncodePermitSingle(owner, permitSingle, sig)
	if err != nil {
		return nil, fmt.Errorf("permit2: createPermitSingle: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("permit2: createPermitSingle: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// PermitSingle broadcasts an already-signed permit(PermitSingle) transaction.
func (p *Permit2) PermitSingle(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("permit2: permitSingle: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("permit2: permitSingle: %w", err)
	}
	return hash, nil
}

// CreatePermitBatch builds an unsigned permit(owner, permitBatch, signature) transaction via b.
func (p *Permit2) CreatePermitBatch(ctx context.Context, b tx.Builder, contract, sender, owner *types.Address, permitBatch *PermitBatch, sig *types.Signature) (core.Transaction, error) {
	calldata, err := EncodePermitBatch(owner, permitBatch, sig)
	if err != nil {
		return nil, fmt.Errorf("permit2: createPermitBatch: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("permit2: createPermitBatch: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// PermitBatch broadcasts an already-signed permit(PermitBatch) transaction.
func (p *Permit2) PermitBatch(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("permit2: permitBatch: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("permit2: permitBatch: %w", err)
	}
	return hash, nil
}

// CreateTransferFrom builds an unsigned transferFrom(from, to, amount, token) transaction via b.
func (p *Permit2) CreateTransferFrom(ctx context.Context, b tx.Builder, contract, sender, from, to *types.Address, amount *big.Int, token *types.Address) (core.Transaction, error) {
	calldata, err := EncodeTransferFrom(from, to, amount, token)
	if err != nil {
		return nil, fmt.Errorf("permit2: createTransferFrom: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("permit2: createTransferFrom: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// TransferFrom broadcasts an already-signed transferFrom transaction.
func (p *Permit2) TransferFrom(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("permit2: transferFrom: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("permit2: transferFrom: %w", err)
	}
	return hash, nil
}

// CreateTransferFromBatch builds an unsigned transferFrom(transferDetails) transaction via b.
func (p *Permit2) CreateTransferFromBatch(ctx context.Context, b tx.Builder, contract, sender *types.Address, transferDetails []AllowanceTransferDetails) (core.Transaction, error) {
	calldata, err := EncodeTransferFromBatch(transferDetails)
	if err != nil {
		return nil, fmt.Errorf("permit2: createTransferFromBatch: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("permit2: createTransferFromBatch: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// TransferFromBatch broadcasts an already-signed transferFrom(batch) transaction.
func (p *Permit2) TransferFromBatch(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("permit2: transferFromBatch: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("permit2: transferFromBatch: %w", err)
	}
	return hash, nil
}

// CreateLockdown builds an unsigned lockdown(approvals) transaction via b.
func (p *Permit2) CreateLockdown(ctx context.Context, b tx.Builder, contract, sender *types.Address, approvals []TokenSpenderPair) (core.Transaction, error) {
	calldata, err := EncodeLockdown(approvals)
	if err != nil {
		return nil, fmt.Errorf("permit2: createLockdown: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("permit2: createLockdown: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// Lockdown broadcasts an already-signed lockdown transaction.
func (p *Permit2) Lockdown(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("permit2: lockdown: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("permit2: lockdown: %w", err)
	}
	return hash, nil
}

// CreateInvalidateNonces builds an unsigned invalidateNonces(token, spender, newNonce) transaction via b.
func (p *Permit2) CreateInvalidateNonces(ctx context.Context, b tx.Builder, contract, sender, token, spender *types.Address, newNonce *big.Int) (core.Transaction, error) {
	calldata, err := EncodeInvalidateNonces(token, spender, newNonce)
	if err != nil {
		return nil, fmt.Errorf("permit2: createInvalidateNonces: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("permit2: createInvalidateNonces: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// InvalidateNonces broadcasts an already-signed invalidateNonces transaction.
func (p *Permit2) InvalidateNonces(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("permit2: invalidateNonces: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("permit2: invalidateNonces: %w", err)
	}
	return hash, nil
}

// NonceBitmap returns owner's unordered-nonce bitmap word at wordPos.
func (p *Permit2) NonceBitmap(ctx context.Context, contract, owner *types.Address, wordPos *big.Int, block string) (*big.Int, error) {
	calldata, err := EncodeNonceBitmap(owner, wordPos)
	if err != nil {
		return nil, fmt.Errorf("permit2: nonceBitmap: %w", err)
	}
	data, err := p.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("permit2: nonceBitmap: %w", err)
	}
	n, err := DecodeNonceBitmap(data)
	if err != nil {
		return nil, fmt.Errorf("permit2: nonceBitmap: %w", err)
	}
	return n, nil
}

// CreatePermitTransferFrom builds an unsigned
// permitTransferFrom(permit, transferDetails, owner, signature) transaction via b.
func (p *Permit2) CreatePermitTransferFrom(ctx context.Context, b tx.Builder, contract, sender *types.Address, permit *PermitTransferFrom, transferDetails *SignatureTransferDetails, owner *types.Address, sig *types.Signature) (core.Transaction, error) {
	calldata, err := EncodePermitTransferFrom(permit, transferDetails, owner, sig)
	if err != nil {
		return nil, fmt.Errorf("permit2: createPermitTransferFrom: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("permit2: createPermitTransferFrom: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// PermitTransferFrom broadcasts an already-signed permitTransferFrom transaction.
func (p *Permit2) PermitTransferFrom(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("permit2: permitTransferFrom: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("permit2: permitTransferFrom: %w", err)
	}
	return hash, nil
}

// CreatePermitWitnessTransferFrom builds an unsigned
// permitWitnessTransferFrom(permit, transferDetails, owner, witness, witnessTypeString, signature)
// transaction via b.
func (p *Permit2) CreatePermitWitnessTransferFrom(ctx context.Context, b tx.Builder, contract, sender *types.Address, permit *PermitTransferFrom, transferDetails *SignatureTransferDetails, owner *types.Address, witness *types.Hash, witnessTypeString string, sig *types.Signature) (core.Transaction, error) {
	calldata, err := EncodePermitWitnessTransferFrom(permit, transferDetails, owner, witness, witnessTypeString, sig)
	if err != nil {
		return nil, fmt.Errorf("permit2: createPermitWitnessTransferFrom: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("permit2: createPermitWitnessTransferFrom: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// PermitWitnessTransferFrom broadcasts an already-signed permitWitnessTransferFrom transaction.
func (p *Permit2) PermitWitnessTransferFrom(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("permit2: permitWitnessTransferFrom: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("permit2: permitWitnessTransferFrom: %w", err)
	}
	return hash, nil
}

// CreatePermitTransferFromBatch builds an unsigned
// permitTransferFrom(permit, transferDetails, owner, signature) (batch overload) transaction via b.
func (p *Permit2) CreatePermitTransferFromBatch(ctx context.Context, b tx.Builder, contract, sender *types.Address, permit *PermitBatchTransferFrom, transferDetails []SignatureTransferDetails, owner *types.Address, sig *types.Signature) (core.Transaction, error) {
	calldata, err := EncodePermitTransferFromBatch(permit, transferDetails, owner, sig)
	if err != nil {
		return nil, fmt.Errorf("permit2: createPermitTransferFromBatch: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("permit2: createPermitTransferFromBatch: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// PermitTransferFromBatch broadcasts an already-signed permitTransferFrom(batch) transaction.
func (p *Permit2) PermitTransferFromBatch(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("permit2: permitTransferFromBatch: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("permit2: permitTransferFromBatch: %w", err)
	}
	return hash, nil
}

// CreatePermitWitnessTransferFromBatch builds an unsigned
// permitWitnessTransferFrom(permit, transferDetails, owner, witness, witnessTypeString, signature)
// (batch overload) transaction via b.
func (p *Permit2) CreatePermitWitnessTransferFromBatch(ctx context.Context, b tx.Builder, contract, sender *types.Address, permit *PermitBatchTransferFrom, transferDetails []SignatureTransferDetails, owner *types.Address, witness *types.Hash, witnessTypeString string, sig *types.Signature) (core.Transaction, error) {
	calldata, err := EncodePermitWitnessTransferFromBatch(permit, transferDetails, owner, witness, witnessTypeString, sig)
	if err != nil {
		return nil, fmt.Errorf("permit2: createPermitWitnessTransferFromBatch: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("permit2: createPermitWitnessTransferFromBatch: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// PermitWitnessTransferFromBatch broadcasts an already-signed permitWitnessTransferFrom(batch) transaction.
func (p *Permit2) PermitWitnessTransferFromBatch(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("permit2: permitWitnessTransferFromBatch: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("permit2: permitWitnessTransferFromBatch: %w", err)
	}
	return hash, nil
}

// CreateInvalidateUnorderedNonces builds an unsigned invalidateUnorderedNonces(wordPos, mask) transaction via b.
func (p *Permit2) CreateInvalidateUnorderedNonces(ctx context.Context, b tx.Builder, contract, sender *types.Address, wordPos, mask *big.Int) (core.Transaction, error) {
	calldata, err := EncodeInvalidateUnorderedNonces(wordPos, mask)
	if err != nil {
		return nil, fmt.Errorf("permit2: createInvalidateUnorderedNonces: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("permit2: createInvalidateUnorderedNonces: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// InvalidateUnorderedNonces broadcasts an already-signed invalidateUnorderedNonces transaction.
func (p *Permit2) InvalidateUnorderedNonces(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("permit2: invalidateUnorderedNonces: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("permit2: invalidateUnorderedNonces: %w", err)
	}
	return hash, nil
}

func (p *Permit2) call(ctx context.Context, from, contract *types.Address, calldata []byte, block string) ([]byte, error) {
	params := map[string]any{
		"to":   contract.String(),
		"data": encoding.Hex.EncodePrefixed(calldata),
	}
	if from != nil {
		params["from"] = from.String()
	}

	var raw string
	if err := p.c.Call(ctx, rpc.ETHCall("call", params, block, &raw)); err != nil {
		return nil, p.wrapRevertError(err)
	}
	return encoding.Hex.Decode(raw)
}

// wrapRevertError maps a known Permit2 custom error in err to its sentinel.
func (_ *Permit2) wrapRevertError(err error) error {
	var rpcErr *rpc.ResponseError
	if !errors.As(err, &rpcErr) || len(rpcErr.Data) == 0 {
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
	case errAllowanceExpired.Selector():
		return fmt.Errorf("%w: %w", ErrAllowanceExpired, err)
	case errInsufficientAllowance.Selector():
		return fmt.Errorf("%w: %w", ErrInsufficientAllowance, err)
	case errExcessiveInvalidation.Selector():
		return fmt.Errorf("%w: %w", ErrExcessiveInvalidation, err)
	case errSignatureExpired.Selector():
		return fmt.Errorf("%w: %w", ErrSignatureExpired, err)
	case errInvalidNonce.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidNonce, err)
	case errInvalidSignatureLength.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidSignatureLength, err)
	case errInvalidSignature.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidSignature, err)
	case errInvalidSigner.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidSigner, err)
	case errInvalidContractSignature.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidContractSignature, err)
	case errUnsafeCast.Selector():
		return fmt.Errorf("%w: %w", ErrUnsafeCast, err)
	case errInvalidAmount.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidAmount, err)
	case errLengthMismatch.Selector():
		return fmt.Errorf("%w: %w", ErrLengthMismatch, err)
	default:
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}
}
