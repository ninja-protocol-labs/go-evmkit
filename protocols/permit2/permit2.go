// Package permit2 is a client for Uniswap's Permit2 (AllowanceTransfer and SignatureTransfer),
// mirroring solidity/interfaces/IPermit2.sol.
package permit2

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

// DefaultPermit2Address is Permit2's canonical CREATE2 address, identical on nearly every EVM chain.
var DefaultPermit2Address = types.MustNewAddressFromHex("0x000000000022D473030F116dDEE9F6B43aC78BA3")

// IEIP712 is DOMAIN_SEPARATOR(), which IAllowanceTransfer and ISignatureTransfer both inherit.
type IEIP712 interface {
	DomainSeparator(ctx context.Context, opts *bind.CallOptions, req *DomainSeparatorRequest) (*types.Hash, error)
}

// IAllowanceTransfer mirrors IAllowanceTransfer.sol.
type IAllowanceTransfer interface {
	IEIP712

	Allowance(ctx context.Context, opts *bind.CallOptions, req *AllowanceRequest) (*PackedAllowance, error)

	Approve(ctx context.Context, opts *bind.TransactOptions, req *ApproveRequest) (core.Transaction, *types.Hash, error)
	PermitSingle(ctx context.Context, opts *bind.TransactOptions, req *PermitSingleRequest) (core.Transaction, *types.Hash, error)
	PermitBatch(ctx context.Context, opts *bind.TransactOptions, req *PermitBatchRequest) (core.Transaction, *types.Hash, error)
	TransferFrom(ctx context.Context, opts *bind.TransactOptions, req *TransferFromRequest) (core.Transaction, *types.Hash, error)
	TransferFromBatch(ctx context.Context, opts *bind.TransactOptions, req *TransferFromBatchRequest) (core.Transaction, *types.Hash, error)
	Lockdown(ctx context.Context, opts *bind.TransactOptions, req *LockdownRequest) (core.Transaction, *types.Hash, error)
	InvalidateNonces(ctx context.Context, opts *bind.TransactOptions, req *InvalidateNoncesRequest) (core.Transaction, *types.Hash, error)
}

// ISignatureTransfer mirrors ISignatureTransfer.sol.
type ISignatureTransfer interface {
	IEIP712

	NonceBitmap(ctx context.Context, opts *bind.CallOptions, req *NonceBitmapRequest) (*big.Int, error)

	PermitTransferFrom(ctx context.Context, opts *bind.TransactOptions, req *PermitTransferFromRequest) (core.Transaction, *types.Hash, error)
	PermitWitnessTransferFrom(ctx context.Context, opts *bind.TransactOptions, req *PermitWitnessTransferFromRequest) (core.Transaction, *types.Hash, error)
	PermitTransferFromBatch(ctx context.Context, opts *bind.TransactOptions, req *PermitTransferFromBatchRequest) (core.Transaction, *types.Hash, error)
	PermitWitnessTransferFromBatch(ctx context.Context, opts *bind.TransactOptions, req *PermitWitnessTransferFromBatchRequest) (core.Transaction, *types.Hash, error)
	InvalidateUnorderedNonces(ctx context.Context, opts *bind.TransactOptions, req *InvalidateUnorderedNoncesRequest) (core.Transaction, *types.Hash, error)
}

// IPermit2 mirrors IPermit2.sol.
type IPermit2 interface {
	IAllowanceTransfer
	ISignatureTransfer
}

var _ IPermit2 = (*Permit2)(nil)

// Permit2 calls any Permit2 deployment over a Client.
type Permit2 struct {
	c rpc.Client
}

// New returns a Permit2 calling contracts via cli.
func New(cli rpc.Client) *Permit2 {
	return &Permit2{
		c: cli,
	}
}

// DomainSeparator returns req.Contract's EIP-712 domain separator.
func (p2 *Permit2) DomainSeparator(ctx context.Context, opts *bind.CallOptions, req *DomainSeparatorRequest) (*types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("permit2: domainSeparator: %w", err)
	}

	data, err := bind.Call(ctx, p2.c, opts, req.Contract, EncodeDomainSeparator())
	if err != nil {
		return nil, fmt.Errorf("permit2: domainSeparator: %w", MapRevertError(err))
	}

	out, err := DecodeDomainSeparator(data)
	if err != nil {
		return nil, fmt.Errorf("permit2: domainSeparator: %w", err)
	}

	return out, nil
}

// Allowance returns req.Owner's allowance of req.Token for req.Spender.
func (p2 *Permit2) Allowance(ctx context.Context, opts *bind.CallOptions, req *AllowanceRequest) (*PackedAllowance, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("permit2: allowance: %w", err)
	}

	calldata, err := EncodeAllowance(req.Owner, req.Token, req.Spender)
	if err != nil {
		return nil, fmt.Errorf("permit2: allowance: %w", err)
	}

	data, err := bind.Call(ctx, p2.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("permit2: allowance: %w", MapRevertError(err))
	}

	out, err := DecodeAllowance(data)
	if err != nil {
		return nil, fmt.Errorf("permit2: allowance: %w", err)
	}

	return out, nil
}

// NonceBitmap returns req.Owner's unordered-nonce bitmap word at req.WordPos.
func (p2 *Permit2) NonceBitmap(ctx context.Context, opts *bind.CallOptions, req *NonceBitmapRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("permit2: nonceBitmap: %w", err)
	}

	calldata, err := EncodeNonceBitmap(req.Owner, req.WordPos)
	if err != nil {
		return nil, fmt.Errorf("permit2: nonceBitmap: %w", err)
	}

	data, err := bind.Call(ctx, p2.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("permit2: nonceBitmap: %w", MapRevertError(err))
	}

	out, err := DecodeNonceBitmap(data)
	if err != nil {
		return nil, fmt.Errorf("permit2: nonceBitmap: %w", err)
	}

	return out, nil
}

// Approve sends approve(token, spender, amount, expiration) on req.Contract, signed by opts.Signer as the owner.
func (p2 *Permit2) Approve(ctx context.Context, opts *bind.TransactOptions, req *ApproveRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("permit2: approve: %w", err)
	}

	calldata, err := EncodeApprove(req.Token, req.Spender, req.Amount, req.Expiration)
	if err != nil {
		return nil, nil, fmt.Errorf("permit2: approve: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, p2.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("permit2: approve: %w", MapRevertError(err))
	}

	return signed, hash, nil
}

// PermitSingle sends permit(owner, permitSingle, signature) on req.Contract, signed by opts.Signer; any account may submit the owner's signature.
func (p2 *Permit2) PermitSingle(ctx context.Context, opts *bind.TransactOptions, req *PermitSingleRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("permit2: permitSingle: %w", err)
	}

	calldata, err := EncodePermitSingle(req.Owner, req.PermitSingle, req.Signature)
	if err != nil {
		return nil, nil, fmt.Errorf("permit2: permitSingle: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, p2.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("permit2: permitSingle: %w", MapRevertError(err))
	}

	return signed, hash, nil
}

// PermitBatch sends permit(owner, permitBatch, signature) on req.Contract, signed by opts.Signer; any account may submit the owner's signature.
func (p2 *Permit2) PermitBatch(ctx context.Context, opts *bind.TransactOptions, req *PermitBatchRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("permit2: permitBatch: %w", err)
	}

	calldata, err := EncodePermitBatch(req.Owner, req.PermitBatch, req.Signature)
	if err != nil {
		return nil, nil, fmt.Errorf("permit2: permitBatch: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, p2.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("permit2: permitBatch: %w", MapRevertError(err))
	}

	return signed, hash, nil
}

// TransferFrom sends transferFrom(from, to, amount, token) on req.Contract, signed by opts.Signer as the approved spender.
func (p2 *Permit2) TransferFrom(ctx context.Context, opts *bind.TransactOptions, req *TransferFromRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("permit2: transferFrom: %w", err)
	}

	calldata, err := EncodeTransferFrom(req.From, req.To, req.Amount, req.Token)
	if err != nil {
		return nil, nil, fmt.Errorf("permit2: transferFrom: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, p2.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("permit2: transferFrom: %w", MapRevertError(err))
	}

	return signed, hash, nil
}

// TransferFromBatch sends transferFrom(transferDetails) on req.Contract, signed by opts.Signer as the approved spender.
func (p2 *Permit2) TransferFromBatch(ctx context.Context, opts *bind.TransactOptions, req *TransferFromBatchRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("permit2: transferFromBatch: %w", err)
	}

	calldata, err := EncodeTransferFromBatch(req.TransferDetails)
	if err != nil {
		return nil, nil, fmt.Errorf("permit2: transferFromBatch: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, p2.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("permit2: transferFromBatch: %w", MapRevertError(err))
	}

	return signed, hash, nil
}

// Lockdown sends lockdown(approvals) on req.Contract, signed by opts.Signer as the owner, zeroing each listed allowance.
func (p2 *Permit2) Lockdown(ctx context.Context, opts *bind.TransactOptions, req *LockdownRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("permit2: lockdown: %w", err)
	}

	calldata, err := EncodeLockdown(req.Approvals)
	if err != nil {
		return nil, nil, fmt.Errorf("permit2: lockdown: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, p2.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("permit2: lockdown: %w", MapRevertError(err))
	}

	return signed, hash, nil
}

// InvalidateNonces sends invalidateNonces(token, spender, newNonce) on req.Contract, signed by opts.Signer as the owner.
func (p2 *Permit2) InvalidateNonces(ctx context.Context, opts *bind.TransactOptions, req *InvalidateNoncesRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("permit2: invalidateNonces: %w", err)
	}

	calldata, err := EncodeInvalidateNonces(req.Token, req.Spender, req.NewNonce)
	if err != nil {
		return nil, nil, fmt.Errorf("permit2: invalidateNonces: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, p2.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("permit2: invalidateNonces: %w", MapRevertError(err))
	}

	return signed, hash, nil
}

// PermitTransferFrom sends permitTransferFrom(permit, transferDetails, owner, signature) on req.Contract, signed by opts.Signer as the spender the owner signed for.
func (p2 *Permit2) PermitTransferFrom(ctx context.Context, opts *bind.TransactOptions, req *PermitTransferFromRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("permit2: permitTransferFrom: %w", err)
	}

	calldata, err := EncodePermitTransferFrom(req.Permit, req.TransferDetails, req.Owner, req.Signature)
	if err != nil {
		return nil, nil, fmt.Errorf("permit2: permitTransferFrom: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, p2.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("permit2: permitTransferFrom: %w", MapRevertError(err))
	}

	return signed, hash, nil
}

// PermitWitnessTransferFrom sends permitWitnessTransferFrom(permit, transferDetails, owner, witness, witnessTypeString, signature) on req.Contract, signed by opts.Signer as the spender.
func (p2 *Permit2) PermitWitnessTransferFrom(ctx context.Context, opts *bind.TransactOptions, req *PermitWitnessTransferFromRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("permit2: permitWitnessTransferFrom: %w", err)
	}

	calldata, err := EncodePermitWitnessTransferFrom(req.Permit, req.TransferDetails, req.Owner, req.Witness, req.WitnessTypeString, req.Signature)
	if err != nil {
		return nil, nil, fmt.Errorf("permit2: permitWitnessTransferFrom: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, p2.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("permit2: permitWitnessTransferFrom: %w", MapRevertError(err))
	}

	return signed, hash, nil
}

// PermitTransferFromBatch sends the batch permitTransferFrom(permit, transferDetails, owner, signature) on req.Contract, signed by opts.Signer as the spender.
func (p2 *Permit2) PermitTransferFromBatch(ctx context.Context, opts *bind.TransactOptions, req *PermitTransferFromBatchRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("permit2: permitTransferFromBatch: %w", err)
	}

	calldata, err := EncodePermitTransferFromBatch(req.Permit, req.TransferDetails, req.Owner, req.Signature)
	if err != nil {
		return nil, nil, fmt.Errorf("permit2: permitTransferFromBatch: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, p2.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("permit2: permitTransferFromBatch: %w", MapRevertError(err))
	}

	return signed, hash, nil
}

// PermitWitnessTransferFromBatch sends the batch permitWitnessTransferFrom(...) on req.Contract, signed by opts.Signer as the spender.
func (p2 *Permit2) PermitWitnessTransferFromBatch(ctx context.Context, opts *bind.TransactOptions, req *PermitWitnessTransferFromBatchRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("permit2: permitWitnessTransferFromBatch: %w", err)
	}

	calldata, err := EncodePermitWitnessTransferFromBatch(req.Permit, req.TransferDetails, req.Owner, req.Witness, req.WitnessTypeString, req.Signature)
	if err != nil {
		return nil, nil, fmt.Errorf("permit2: permitWitnessTransferFromBatch: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, p2.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("permit2: permitWitnessTransferFromBatch: %w", MapRevertError(err))
	}

	return signed, hash, nil
}

// InvalidateUnorderedNonces sends invalidateUnorderedNonces(wordPos, mask) on req.Contract, signed by opts.Signer as the owner.
func (p2 *Permit2) InvalidateUnorderedNonces(ctx context.Context, opts *bind.TransactOptions, req *InvalidateUnorderedNoncesRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("permit2: invalidateUnorderedNonces: %w", err)
	}

	calldata, err := EncodeInvalidateUnorderedNonces(req.WordPos, req.Mask)
	if err != nil {
		return nil, nil, fmt.Errorf("permit2: invalidateUnorderedNonces: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, p2.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("permit2: invalidateUnorderedNonces: %w", MapRevertError(err))
	}

	return signed, hash, nil
}
