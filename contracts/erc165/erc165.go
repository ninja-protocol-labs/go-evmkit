package erc165

import (
	"context"
	"errors"
	"fmt"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// IERC165 is implemented by *ERC165.
type IERC165 interface {
	SupportsInterface(ctx context.Context, contract *types.Address, interfaceId [4]byte, block string) (bool, error)
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

// SupportsInterface returns whether contract implements interfaceId.
func (e *ERC165) SupportsInterface(ctx context.Context, contract *types.Address, interfaceId [4]byte, block string) (bool, error) {
	calldata, err := EncodeSupportsInterface(interfaceId)
	if err != nil {
		return false, fmt.Errorf("erc165: supportsInterface: %w", err)
	}

	data, err := e.call(ctx, contract, calldata, block)
	if err != nil {
		return false, fmt.Errorf("erc165: supportsInterface: %w", err)
	}

	ok, err := DecodeSupportsInterface(data)
	if err != nil {
		return false, fmt.Errorf("erc165: supportsInterface: %w: %w", ErrNotImplementedERC165, err)
	}
	return ok, nil
}

// call eth_calls contract with calldata, mapping a revert to ErrNotImplementedERC165.
func (e *ERC165) call(ctx context.Context, contract *types.Address, calldata []byte, block string) ([]byte, error) {
	params := map[string]any{
		"to":   contract.String(),
		"data": encoding.Hex.EncodePrefixed(calldata),
	}

	var raw string
	if err := e.c.Call(ctx, rpc.ETHCall("call", params, block, &raw)); err != nil {
		return nil, e.wrapRevertError(err)
	}
	return encoding.Hex.Decode(raw)
}

// wrapRevertError maps an on-chain revert (the node responded) to
// ErrNotImplementedERC165; a transport/connection failure, which says
// nothing about the contract, is returned unchanged.
func (_ *ERC165) wrapRevertError(err error) error {
	if _, ok := errors.AsType[*rpc.ResponseError](err); !ok {
		return err
	}
	return fmt.Errorf("%w: %w", ErrNotImplementedERC165, err)
}
