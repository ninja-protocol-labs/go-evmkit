// Package erc165 is a client for ERC-165 introspection, mirroring
// solidity/interfaces/IERC165.sol.
package erc165

import (
	"context"
	"errors"
	"fmt"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

// IERC165 is implemented by *ERC165.
type IERC165 interface {
	SupportsInterface(ctx context.Context, opts *bind.CallOptions, req *SupportsInterfaceRequest) (bool, error)
}

var _ IERC165 = (*ERC165)(nil)

// ERC165 calls any ERC-165-compliant contract over a Client; it holds no
// contract address, so one instance serves every contract.
type ERC165 struct {
	c rpc.Client
}

// New returns an ERC165 calling contracts via cli.
func New(cli rpc.Client) *ERC165 {
	return &ERC165{
		c: cli,
	}
}

// SupportsInterface returns whether req.Contract implements req.InterfaceID.
func (e *ERC165) SupportsInterface(ctx context.Context, opts *bind.CallOptions, req *SupportsInterfaceRequest) (bool, error) {
	if err := req.Validate(); err != nil {
		return false, fmt.Errorf("erc165: supportsInterface: %w", err)
	}

	calldata, err := EncodeSupportsInterface(req.InterfaceID)
	if err != nil {
		return false, fmt.Errorf("erc165: supportsInterface: %w", err)
	}

	data, err := bind.Call(ctx, e.c, opts, req.Contract, calldata)
	if err != nil {
		return false, fmt.Errorf("erc165: supportsInterface: %w", e.wrapRevertError(err))
	}

	ok, err := DecodeSupportsInterface(data)
	if err != nil {
		return false, fmt.Errorf("erc165: supportsInterface: %w: %w", ErrNotImplementedERC165, err)
	}

	return ok, nil
}

// wrapRevertError maps an on-chain revert to ErrNotImplementedERC165; a
// transport failure says nothing about the contract and is returned unchanged.
func (_ *ERC165) wrapRevertError(err error) error {
	if _, ok := errors.AsType[*rpc.ResponseError](err); !ok {
		return err
	}
	return fmt.Errorf("%w: %w", ErrNotImplementedERC165, err)
}
