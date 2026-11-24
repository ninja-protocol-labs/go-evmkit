// Package bind holds the options and the single call/transact path every
// contract client shares: reads go through Call, writes through Transact.
package bind

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

var (
	ErrInvalidOptions = errors.New("invalid options")
	ErrInvalidRequest = errors.New("invalid request")
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

// TransactOptions configures a signed, broadcast transaction. Signer and
// Builder are required; Signer also determines the sender address.
type TransactOptions struct {
	Signer core.Signer
	// Builder picks the transaction type and gas/nonce overrides, which
	// apply to every call it serves.
	Builder tx.Builder
	// Value is msg.value in wei; nil means zero.
	Value *big.Int
}

// NewTransactOptions returns TransactOptions with its required Signer and Builder.
func NewTransactOptions(s core.Signer, b tx.Builder) *TransactOptions {
	return &TransactOptions{
		Signer:  s,
		Builder: b,
	}
}

// WithValue sets Value in place and returns o.
func (o *TransactOptions) WithValue(value *big.Int) *TransactOptions {
	o.Value = value
	return o
}

// From returns the address Signer signs for.
func (o *TransactOptions) From() (*types.Address, error) {
	if o == nil {
		return nil, fmt.Errorf("bind: from: %w: options is nil", ErrInvalidOptions)
	}
	if o.Signer == nil {
		return nil, fmt.Errorf("bind: from: %w: signer is nil", ErrInvalidOptions)
	}
	return core.PubkeyToAddress(o.Signer.PublicKey()), nil
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
	if err := c.Call(ctx, rpc.ETHCall("call", params, block, nil, &raw)); err != nil {
		return nil, err
	}
	data, err := encoding.Hex.Decode(raw)
	if err != nil {
		return nil, fmt.Errorf("bind: call: %w", err)
	}
	return data, nil
}

// Transact builds, signs and broadcasts calldata to "to" (nil deploys a contract),
// returning the signed transaction and its hash.
func Transact(ctx context.Context, c rpc.Client, opts *TransactOptions, to *types.Address, calldata []byte) (core.Transaction, *types.Hash, error) {
	if opts == nil {
		return nil, nil, fmt.Errorf("bind: transact: %w: options is nil", ErrInvalidOptions)
	}
	if opts.Signer == nil {
		return nil, nil, fmt.Errorf("bind: transact: %w: signer is nil", ErrInvalidOptions)
	}
	if opts.Builder == nil {
		return nil, nil, fmt.Errorf("bind: transact: %w: builder is nil", ErrInvalidOptions)
	}
	from := core.PubkeyToAddress(opts.Signer.PublicKey())

	value := opts.Value
	if value == nil {
		value = big.NewInt(0)
	}

	signed, err := opts.Builder.Build(from, to, value, calldata).Pack(ctx, c)
	if err != nil {
		return nil, nil, fmt.Errorf("bind: transact: %w", err)
	}

	if err = signed.Sign(opts.Signer); err != nil {
		return nil, nil, fmt.Errorf("bind: transact: sign: %w", err)
	}

	raw, err := signed.EncodeRLP()
	if err != nil {
		return nil, nil, fmt.Errorf("bind: transact: encode: %w", err)
	}

	hash, err := tx.Broadcast(ctx, c, raw)
	if err != nil {
		return signed, nil, fmt.Errorf("bind: transact: %w", err)
	}
	return signed, hash, nil
}
