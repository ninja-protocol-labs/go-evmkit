// Package erc2612 is a client for ERC-2612 (Permit) tokens, mirroring IERC20Permit.sol; pair it
// with an *erc20.ERC20 on the same contract for the rest of the ERC-20 surface.
package erc2612

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// IERC2612 mirrors IERC20Permit.sol.
type IERC2612 interface {
	Nonces(ctx context.Context, opts *bind.CallOptions, req *NoncesRequest) (*big.Int, error)
	DomainSeparator(ctx context.Context, opts *bind.CallOptions, req *DomainSeparatorRequest) (*types.Hash, error)

	Permit(ctx context.Context, opts *bind.TransactOptions, req *PermitRequest) (core.Transaction, *types.Hash, error)
}

var _ IERC2612 = (*ERC2612)(nil)

// ERC2612 calls any ERC-2612 token contract over a Client; it holds no contract address.
type ERC2612 struct {
	c rpc.Client
}

// New returns an ERC2612 calling contracts via cli.
func New(cli rpc.Client) *ERC2612 {
	return &ERC2612{
		c: cli,
	}
}

// Nonces returns req.Owner's current permit nonce on req.Contract.
func (e *ERC2612) Nonces(ctx context.Context, opts *bind.CallOptions, req *NoncesRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc2612: nonces: %w", err)
	}

	calldata, err := EncodeNonces(req.Owner)
	if err != nil {
		return nil, fmt.Errorf("erc2612: nonces: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("erc2612: nonces: %w", e.wrapRevertError(err))
	}

	out, err := DecodeNonces(data)
	if err != nil {
		return nil, fmt.Errorf("erc2612: nonces: %w", err)
	}

	return out, nil
}

// DomainSeparator returns req.Contract's EIP-712 domain separator.
func (e *ERC2612) DomainSeparator(ctx context.Context, opts *bind.CallOptions, req *DomainSeparatorRequest) (*types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("erc2612: domainSeparator: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, EncodeDomainSeparator())
	if err != nil {
		return nil, fmt.Errorf("erc2612: domainSeparator: %w", e.wrapRevertError(err))
	}

	b, err := DecodeDomainSeparator(data)
	if err != nil {
		return nil, fmt.Errorf("erc2612: domainSeparator: %w", err)
	}

	return types.NewHashFromBytes(b), nil
}

// Permit relays permit(owner, spender, value, deadline, v, r, s) on req.Contract, sent by
// opts.Signer; req.Signature must come from req.Owner.
func (e *ERC2612) Permit(ctx context.Context, opts *bind.TransactOptions, req *PermitRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("erc2612: permit: %w", err)
	}

	calldata, err := EncodePermit(req.Owner, req.Spender, req.Amount, req.Deadline, req.Signature)
	if err != nil {
		return nil, nil, fmt.Errorf("erc2612: permit: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("erc2612: permit: %w", e.wrapRevertError(err))
	}

	return signed, hash, nil
}

// wrapRevertError maps a known ERC20Permit custom error in err to its sentinel.
func (_ *ERC2612) wrapRevertError(err error) error {
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
	case errExpiredSignature.Selector():
		return fmt.Errorf("%w: %w", ErrExpiredSignature, err)
	case errInvalidSigner.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidSigner, err)
	default:
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}
}
