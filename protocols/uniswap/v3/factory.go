package v3

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

// IFactory mirrors IUniswapV3Factory.
type IFactory interface {
	Owner(ctx context.Context, opts *bind.CallOptions, req *OwnerRequest) (*types.Address, error)
	FeeAmountTickSpacing(ctx context.Context, opts *bind.CallOptions, req *FeeAmountTickSpacingRequest) (*big.Int, error)
	GetPool(ctx context.Context, opts *bind.CallOptions, req *GetPoolRequest) (*types.Address, error)
	CallCreatePool(ctx context.Context, opts *bind.CallOptions, req *CreatePoolRequest) (*types.Address, error)

	CreatePool(ctx context.Context, opts *bind.TransactOptions, req *CreatePoolRequest) (core.Transaction, *types.Hash, error)
	SetOwner(ctx context.Context, opts *bind.TransactOptions, req *SetOwnerRequest) (core.Transaction, *types.Hash, error)
	EnableFeeAmount(ctx context.Context, opts *bind.TransactOptions, req *EnableFeeAmountRequest) (core.Transaction, *types.Hash, error)
}

var _ IFactory = (*Factory)(nil)

// Factory calls any UniswapV3Factory contract over a Client; it holds no contract address, so one instance serves every deployment.
type Factory struct {
	c rpc.Client
}

// NewFactory returns a Factory calling factories via cli.
func NewFactory(cli rpc.Client) *Factory {
	return &Factory{
		c: cli,
	}
}

// Owner returns req.Contract's owner, the only account allowed to call SetOwner/EnableFeeAmount.
func (f *Factory) Owner(ctx context.Context, opts *bind.CallOptions, req *OwnerRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: owner: %w", err)
	}

	data, err := bind.Call(ctx, f.c, opts, req.Contract, EncodeOwner())
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: owner: %w", mapFactoryRevertError(err))
	}

	out, err := DecodeOwner(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: owner: %w", err)
	}

	return out, nil
}

// FeeAmountTickSpacing returns the tick spacing req.Contract enforces for req.Fee, or zero if the tier isn't enabled.
func (f *Factory) FeeAmountTickSpacing(ctx context.Context, opts *bind.CallOptions, req *FeeAmountTickSpacingRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: feeAmountTickSpacing: %w", err)
	}

	calldata, err := EncodeFeeAmountTickSpacing(req.Fee)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: feeAmountTickSpacing: %w", err)
	}

	data, err := bind.Call(ctx, f.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: feeAmountTickSpacing: %w", mapFactoryRevertError(err))
	}

	out, err := DecodeFeeAmountTickSpacing(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: feeAmountTickSpacing: %w", err)
	}

	return out, nil
}

// GetPool returns the pool for (req.TokenA, req.TokenB, req.Fee), or the zero address if none exists yet.
func (f *Factory) GetPool(ctx context.Context, opts *bind.CallOptions, req *GetPoolRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: getPool: %w", err)
	}

	calldata, err := EncodeGetPool(req.TokenA, req.TokenB, req.Fee)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: getPool: %w", err)
	}

	data, err := bind.Call(ctx, f.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: getPool: %w", mapFactoryRevertError(err))
	}

	out, err := DecodeGetPool(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: getPool: %w", err)
	}

	return out, nil
}

// CallCreatePool runs createPool(tokenA, tokenB, fee) via eth_call as opts.From, returning the pool it would create or the revert.
func (f *Factory) CallCreatePool(ctx context.Context, opts *bind.CallOptions, req *CreatePoolRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: callCreatePool: %w", err)
	}

	calldata, err := EncodeCreatePool(req.TokenA, req.TokenB, req.Fee)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callCreatePool: %w", err)
	}

	data, err := bind.Call(ctx, f.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callCreatePool: %w", mapFactoryRevertError(err))
	}

	out, err := DecodeCreatePool(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callCreatePool: %w", err)
	}

	return out, nil
}

// CreatePool sends createPool(tokenA, tokenB, fee) on req.Contract, signed by opts.Signer; the new pool comes back via ExtractPoolCreated.
func (f *Factory) CreatePool(ctx context.Context, opts *bind.TransactOptions, req *CreatePoolRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: createPool: %w", err)
	}

	calldata, err := EncodeCreatePool(req.TokenA, req.TokenB, req.Fee)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: createPool: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, f.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: createPool: %w", mapFactoryRevertError(err))
	}

	return signed, hash, nil
}

// SetOwner sends setOwner(_owner) on req.Contract, signed by opts.Signer, which must be the owner or it reverts with ErrReverted.
func (f *Factory) SetOwner(ctx context.Context, opts *bind.TransactOptions, req *SetOwnerRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: setOwner: %w", err)
	}

	calldata, err := EncodeSetOwner(req.Owner)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: setOwner: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, f.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: setOwner: %w", mapFactoryRevertError(err))
	}

	return signed, hash, nil
}

// EnableFeeAmount sends enableFeeAmount(fee, tickSpacing) on req.Contract, signed by opts.Signer, which must be the owner or it reverts with ErrReverted.
func (f *Factory) EnableFeeAmount(ctx context.Context, opts *bind.TransactOptions, req *EnableFeeAmountRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: enableFeeAmount: %w", err)
	}

	calldata, err := EncodeEnableFeeAmount(req.Fee, req.TickSpacing)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: enableFeeAmount: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, f.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: enableFeeAmount: %w", mapFactoryRevertError(err))
	}

	return signed, hash, nil
}
