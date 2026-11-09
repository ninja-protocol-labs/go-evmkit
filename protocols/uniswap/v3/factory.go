package v3

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

// IFactory mirrors IUniswapV3Factory. No tx.Builder-taking Create* constructors here -- see
// CreateCreatePool, CreateSetOwner, and CreateEnableFeeAmount.
type IFactory interface {
	Owner(ctx context.Context, contract *types.Address, block string) (*types.Address, error)
	FeeAmountTickSpacing(ctx context.Context, contract *types.Address, fee *big.Int, block string) (*big.Int, error)
	GetPool(ctx context.Context, p *GetPoolParams, block string) (*types.Address, error)

	CreatePool(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SetOwner(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	EnableFeeAmount(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

var _ IFactory = (*Factory)(nil)

// Factory calls any UniswapV3Factory contract over a Client; it holds no contract address, so
// one instance serves every network's deployment.
type Factory struct {
	c rpc.Client
}

// NewFactory returns a Factory calling factories via c.
func NewFactory(c rpc.Client) *Factory {
	return &Factory{
		c: c,
	}
}

// Owner returns contract's current owner, the only account allowed to call
// SetOwner/EnableFeeAmount.
func (f *Factory) Owner(ctx context.Context, contract *types.Address, block string) (*types.Address, error) {
	data, err := f.call(ctx, nil, contract, EncodeOwner(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: owner: %w", err)
	}
	addr, err := DecodeOwner(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: owner: %w", err)
	}
	return addr, nil
}

// FeeAmountTickSpacing returns the tick spacing contract enforces for fee, or zero if fee
// isn't an enabled tier.
func (f *Factory) FeeAmountTickSpacing(ctx context.Context, contract *types.Address, fee *big.Int, block string) (*big.Int, error) {
	calldata, err := EncodeFeeAmountTickSpacing(fee)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: feeAmountTickSpacing: %w", err)
	}
	data, err := f.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: feeAmountTickSpacing: %w", err)
	}
	n, err := DecodeFeeAmountTickSpacing(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: feeAmountTickSpacing: %w", err)
	}
	return n, nil
}

// GetPool returns the pool contract deployed for (tokenA, tokenB, fee), or the zero address
// if none exists yet.
func (f *Factory) GetPool(ctx context.Context, p *GetPoolParams, block string) (*types.Address, error) {
	calldata, err := EncodeGetPool(p.TokenA, p.TokenB, p.Fee)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: getPool: %w", err)
	}
	data, err := f.call(ctx, nil, p.Factory, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: getPool: %w", err)
	}
	addr, err := DecodeGetPool(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: getPool: %w", err)
	}
	return addr, nil
}

// CreateCreatePool builds an unsigned createPool(tokenA, tokenB, fee) transaction via b, sent by sender.
func (f *Factory) CreateCreatePool(ctx context.Context, b tx.Builder, p *CreatePoolParams) (core.Transaction, error) {
	calldata, err := EncodeCreatePool(p.TokenA, p.TokenB, p.Fee)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createCreatePool: %w", err)
	}

	unsigned, err := b.Build(p.Sender, p.Factory, nil, calldata).Pack(ctx, f.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createCreatePool: %w", f.wrapRevertError(err))
	}
	return unsigned, nil
}

// CallCreatePool simulates createPool(tokenA, tokenB, fee) via eth_call, as sender, returning the
// pool address it would create -- without broadcasting anything.
func (f *Factory) CallCreatePool(ctx context.Context, p *CreatePoolParams, block string) (*types.Address, error) {
	calldata, err := EncodeCreatePool(p.TokenA, p.TokenB, p.Fee)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callCreatePool: %w", err)
	}
	data, err := f.call(ctx, p.Sender, p.Factory, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callCreatePool: %w", err)
	}
	addr, err := DecodeCreatePool(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callCreatePool: %w", err)
	}
	return addr, nil
}

// CreatePool broadcasts an already-signed createPool transaction. The new pool's address comes
// back via the PoolCreated event (ExtractPoolCreated), not a return value -- a mined
// transaction's return value isn't retrievable.
func (f *Factory) CreatePool(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createPool: %w", err)
	}
	hash, err := tx.Broadcast(ctx, f.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createPool: %w", err)
	}
	return hash, nil
}

// CreateSetOwner builds an unsigned setOwner(_owner) transaction via b, sent by sender -- which
// must be contract's current owner, or the transaction reverts with ErrReverted.
func (f *Factory) CreateSetOwner(ctx context.Context, b tx.Builder, p *SetOwnerParams) (core.Transaction, error) {
	calldata, err := EncodeSetOwner(p.Owner)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createSetOwner: %w", err)
	}

	unsigned, err := b.Build(p.Sender, p.Factory, nil, calldata).Pack(ctx, f.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createSetOwner: %w", f.wrapRevertError(err))
	}
	return unsigned, nil
}

// SetOwner broadcasts an already-signed setOwner transaction.
func (f *Factory) SetOwner(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: setOwner: %w", err)
	}
	hash, err := tx.Broadcast(ctx, f.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: setOwner: %w", err)
	}
	return hash, nil
}

// CreateEnableFeeAmount builds an unsigned enableFeeAmount(fee, tickSpacing) transaction via b,
// sent by sender -- which must be contract's current owner, or the transaction reverts with
// ErrReverted.
func (f *Factory) CreateEnableFeeAmount(ctx context.Context, b tx.Builder, p *EnableFeeAmountParams) (core.Transaction, error) {
	calldata, err := EncodeEnableFeeAmount(p.Fee, p.TickSpacing)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createEnableFeeAmount: %w", err)
	}

	unsigned, err := b.Build(p.Sender, p.Factory, nil, calldata).Pack(ctx, f.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createEnableFeeAmount: %w", f.wrapRevertError(err))
	}
	return unsigned, nil
}

// EnableFeeAmount broadcasts an already-signed enableFeeAmount transaction.
func (f *Factory) EnableFeeAmount(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: enableFeeAmount: %w", err)
	}
	hash, err := tx.Broadcast(ctx, f.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: enableFeeAmount: %w", err)
	}
	return hash, nil
}

// call eth_calls contract with calldata, wrapping a revert via wrapRevertError.
func (f *Factory) call(ctx context.Context, from, contract *types.Address, calldata []byte, block string) ([]byte, error) {
	params := map[string]any{
		"to":   contract.String(),
		"data": encoding.Hex.EncodePrefixed(calldata),
	}
	if from != nil {
		params["from"] = from.String()
	}

	var raw string
	if err := f.c.Call(ctx, rpc.ETHCall("call", params, block, &raw)); err != nil {
		return nil, f.wrapRevertError(err)
	}
	return encoding.Hex.Decode(raw)
}

// wrapRevertError maps any UniswapV3Factory revert to ErrReverted -- there's no reason string
// or selector to switch on, since every require in UniswapV3Factory carries neither.
func (_ *Factory) wrapRevertError(err error) error {
	if _, ok := errors.AsType[*rpc.ResponseError](err); !ok {
		return err
	}
	return fmt.Errorf("%w: %w", ErrReverted, err)
}
