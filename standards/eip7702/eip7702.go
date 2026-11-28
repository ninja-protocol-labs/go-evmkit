// Package eip7702 prepares and reads EIP-7702 delegations: the delegation designator codec, account code lookup and authorization constructors.
package eip7702

import (
	"context"
	"fmt"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

// IEIP7702 is implemented by *EIP7702.
type IEIP7702 interface {
	GetDelegation(ctx context.Context, opts *bind.CallOptions, req *GetDelegationRequest) (*AccountCode, error)
}

// EIP7702 reads account code over a Client; it holds no address, so one instance serves every account.
type EIP7702 struct {
	c rpc.Client
}

var _ IEIP7702 = (*EIP7702)(nil)

// NewEIP7702 returns an EIP7702 reading via c.
func NewEIP7702(c rpc.Client) *EIP7702 {
	return &EIP7702{
		c: c,
	}
}

// GetDelegationElement validates opts and req and builds the eth_getCode element decoding into result, for sending in an rpc batch.
func GetDelegationElement(id string, opts *bind.CallOptions, req *GetDelegationRequest, result *AccountCode) (rpc.Element, error) {
	if err := req.Validate(); err != nil {
		return rpc.Element{}, fmt.Errorf("eip7702: getDelegation: %w", err)
	}
	if result == nil {
		return rpc.Element{}, fmt.Errorf("eip7702: getDelegation: %w: result is nil", bind.ErrInvalidRequest)
	}
	if opts == nil {
		opts = bind.NewCallOptions()
	}
	if opts.StateOverride != nil {
		return rpc.Element{}, fmt.Errorf("eip7702: getDelegation: %w: eth_getCode takes no state override", bind.ErrInvalidOptions)
	}

	block := opts.Block
	if block == "" {
		block = rpc.BlockTagLatest
	}

	*result = AccountCode{}
	return rpc.NewElement(id, rpc.MethodETHGetCode, []any{req.Account.String(), block}, result), nil
}

// GetDelegation classifies req.Account's code at opts.Block via eth_getCode; only Block is read from opts.
func (e *EIP7702) GetDelegation(ctx context.Context, opts *bind.CallOptions, req *GetDelegationRequest) (*AccountCode, error) {
	var code AccountCode
	elem, err := GetDelegationElement("code", opts, req, &code)
	if err != nil {
		return nil, err
	}

	if err := e.c.Call(ctx, elem); err != nil {
		return nil, fmt.Errorf("eip7702: getDelegation: %w", err)
	}
	return &code, nil
}
