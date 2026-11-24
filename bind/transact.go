package bind

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

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
