package bind

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// CallOptions configures an eth_call. A nil *CallOptions is the zero value:
// no from, latest block, no value.
type CallOptions struct {
	// From is the simulated msg.sender; nil omits it.
	From *types.Address
	// Block is the block tag or hex number; empty means rpc.BlockTagLatest.
	Block string
	// Value is the simulated msg.value for payable functions; nil omits it.
	Value *big.Int
	// StateOverride is the eth_call state override; nil omits it.
	StateOverride *rpc.StateOverride
}

// NewCallOptions returns empty CallOptions: no from, latest block, no value.
func NewCallOptions() *CallOptions {
	return new(CallOptions)
}

// WithFrom sets From in place and returns o.
func (o *CallOptions) WithFrom(from *types.Address) *CallOptions {
	o.From = from
	return o
}

// WithBlock sets Block in place and returns o.
func (o *CallOptions) WithBlock(tag string) *CallOptions {
	o.Block = tag
	return o
}

// WithValue sets Value in place and returns o.
func (o *CallOptions) WithValue(value *big.Int) *CallOptions {
	o.Value = value
	return o
}

// WithStateOverride sets StateOverride in place and returns o.
func (o *CallOptions) WithStateOverride(override *rpc.StateOverride) *CallOptions {
	o.StateOverride = override
	return o
}

// Call runs eth_call against to and returns the raw return data. RPC
// errors are returned unwrapped so callers can map revert data to sentinels.
func Call(ctx context.Context, c rpc.Client, opts *CallOptions, to *types.Address, calldata []byte) ([]byte, error) {
	if to == nil {
		return nil, fmt.Errorf("bind: call: %w: to is nil", ErrInvalidRequest)
	}
	if opts == nil {
		opts = NewCallOptions()
	}
	if err := opts.StateOverride.Validate(); err != nil {
		return nil, fmt.Errorf("bind: call: %w: %w", ErrInvalidOptions, err)
	}

	params := map[string]any{
		"to":   to.String(),
		"data": encoding.Hex.EncodePrefixed(calldata),
	}
	if opts.From != nil {
		params["from"] = opts.From.String()
	}
	if opts.Value != nil {
		params["value"] = "0x" + opts.Value.Text(16)
	}

	block := opts.Block
	if block == "" {
		block = rpc.BlockTagLatest
	}

	var raw string
	if err := c.Call(ctx, rpc.ETHCall("call", params, block, opts.StateOverride, &raw)); err != nil {
		return nil, err
	}
	data, err := encoding.Hex.Decode(raw)
	if err != nil {
		return nil, fmt.Errorf("bind: call: %w", err)
	}
	return data, nil
}
