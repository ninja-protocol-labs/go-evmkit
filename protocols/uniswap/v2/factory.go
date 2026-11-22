package v2

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

// IFactory mirrors IUniswapV2Factory.
type IFactory interface {
	FeeTo(ctx context.Context, opts *bind.CallOptions, req *FeeToRequest) (*types.Address, error)
	FeeToSetter(ctx context.Context, opts *bind.CallOptions, req *FeeToSetterRequest) (*types.Address, error)
	GetPair(ctx context.Context, opts *bind.CallOptions, req *GetPairRequest) (*types.Address, error)
	AllPairs(ctx context.Context, opts *bind.CallOptions, req *AllPairsRequest) (*types.Address, error)
	AllPairsLength(ctx context.Context, opts *bind.CallOptions, req *AllPairsLengthRequest) (*big.Int, error)
	CallCreatePair(ctx context.Context, opts *bind.CallOptions, req *CreatePairRequest) (*types.Address, error)

	CreatePair(ctx context.Context, opts *bind.TransactOptions, req *CreatePairRequest) (core.Transaction, *types.Hash, error)
	SetFeeTo(ctx context.Context, opts *bind.TransactOptions, req *SetFeeToRequest) (core.Transaction, *types.Hash, error)
	SetFeeToSetter(ctx context.Context, opts *bind.TransactOptions, req *SetFeeToSetterRequest) (core.Transaction, *types.Hash, error)
}

var _ IFactory = (*Factory)(nil)

// Factory calls any UniswapV2Factory deployment over a Client.
type Factory struct {
	c rpc.Client
}

// NewFactory returns a Factory calling factories via cli.
func NewFactory(cli rpc.Client) *Factory {
	return &Factory{
		c: cli,
	}
}

// FeeTo returns req.Contract's protocol fee recipient (zero when the fee switch is off).
func (f *Factory) FeeTo(ctx context.Context, opts *bind.CallOptions, req *FeeToRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: feeTo: %w", err)
	}

	calldata := EncodeFeeTo()

	data, err := bind.Call(ctx, f.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: feeTo: %w", f.wrapRevertError(err))
	}

	out, err := DecodeFeeTo(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: feeTo: %w", err)
	}

	return out, nil
}

// FeeToSetter returns req.Contract's governance address, the only account allowed to call SetFeeTo/SetFeeToSetter.
func (f *Factory) FeeToSetter(ctx context.Context, opts *bind.CallOptions, req *FeeToSetterRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: feeToSetter: %w", err)
	}

	calldata := EncodeFeeToSetter()

	data, err := bind.Call(ctx, f.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: feeToSetter: %w", f.wrapRevertError(err))
	}

	out, err := DecodeFeeToSetter(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: feeToSetter: %w", err)
	}

	return out, nil
}

// GetPair returns the pair for (req.TokenA, req.TokenB), or the zero address if none exists yet.
func (f *Factory) GetPair(ctx context.Context, opts *bind.CallOptions, req *GetPairRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: getPair: %w", err)
	}

	calldata, err := EncodeGetPair(req.TokenA, req.TokenB)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getPair: %w", err)
	}

	data, err := bind.Call(ctx, f.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getPair: %w", f.wrapRevertError(err))
	}

	out, err := DecodeGetPair(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getPair: %w", err)
	}

	return out, nil
}

// AllPairs returns the pair at req.Index in req.Contract's allPairs array.
func (f *Factory) AllPairs(ctx context.Context, opts *bind.CallOptions, req *AllPairsRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: allPairs: %w", err)
	}

	calldata, err := EncodeAllPairs(req.Index)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: allPairs: %w", err)
	}

	data, err := bind.Call(ctx, f.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: allPairs: %w", f.wrapRevertError(err))
	}

	out, err := DecodeAllPairs(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: allPairs: %w", err)
	}

	return out, nil
}

// AllPairsLength returns how many pairs req.Contract has created.
func (f *Factory) AllPairsLength(ctx context.Context, opts *bind.CallOptions, req *AllPairsLengthRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: allPairsLength: %w", err)
	}

	calldata := EncodeAllPairsLength()

	data, err := bind.Call(ctx, f.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: allPairsLength: %w", f.wrapRevertError(err))
	}

	out, err := DecodeAllPairsLength(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: allPairsLength: %w", err)
	}

	return out, nil
}

// CallCreatePair runs createPair(tokenA, tokenB) via eth_call as opts.From, returning the pair it would create or the revert.
func (f *Factory) CallCreatePair(ctx context.Context, opts *bind.CallOptions, req *CreatePairRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: callCreatePair: %w", err)
	}

	calldata, err := EncodeCreatePair(req.TokenA, req.TokenB)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callCreatePair: %w", err)
	}

	data, err := bind.Call(ctx, f.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callCreatePair: %w", f.wrapRevertError(err))
	}

	out, err := DecodeCreatePair(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callCreatePair: %w", err)
	}

	return out, nil
}

// CreatePair sends createPair(tokenA, tokenB) on req.Contract, signed by opts.Signer; the new pair comes back via ExtractPairCreated.
func (f *Factory) CreatePair(ctx context.Context, opts *bind.TransactOptions, req *CreatePairRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: createPair: %w", err)
	}

	calldata, err := EncodeCreatePair(req.TokenA, req.TokenB)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: createPair: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, f.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: createPair: %w", f.wrapRevertError(err))
	}

	return signed, hash, nil
}

// SetFeeTo sends setFeeTo(feeTo) on req.Contract, signed by opts.Signer, which must be the feeToSetter or it reverts with ErrForbidden.
func (f *Factory) SetFeeTo(ctx context.Context, opts *bind.TransactOptions, req *SetFeeToRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: setFeeTo: %w", err)
	}

	calldata, err := EncodeSetFeeTo(req.FeeTo)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: setFeeTo: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, f.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: setFeeTo: %w", f.wrapRevertError(err))
	}

	return signed, hash, nil
}

// SetFeeToSetter sends setFeeToSetter(feeToSetter) on req.Contract, signed by opts.Signer, which must be the feeToSetter or it reverts with ErrForbidden.
func (f *Factory) SetFeeToSetter(ctx context.Context, opts *bind.TransactOptions, req *SetFeeToSetterRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: setFeeToSetter: %w", err)
	}

	calldata, err := EncodeSetFeeToSetter(req.FeeToSetter)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: setFeeToSetter: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, f.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: setFeeToSetter: %w", f.wrapRevertError(err))
	}

	return signed, hash, nil
}

// wrapRevertError maps a known UniswapV2Factory require(cond, reason) string in err to its sentinel.
// UniswapV2Factory predates Solidity custom errors, so there's no 4-byte selector to switch on --
// just the plain Error(string) reason, decoded via abi.DecodeRevert.
func (_ *Factory) wrapRevertError(err error) error {
	var rpcErr *rpc.ResponseError
	if !errors.As(err, &rpcErr) || len(rpcErr.Data) == 0 {
		return err
	}

	var hexData string
	if jsonErr := json.Unmarshal(rpcErr.Data, &hexData); jsonErr != nil {
		return err
	}

	data, decErr := encoding.Hex.Decode(hexData)
	if decErr != nil {
		return err
	}

	reason, decErr := abi.DecodeRevert(data)
	if decErr != nil {
		return err
	}

	switch reason {
	case reasonIdenticalAddresses:
		return fmt.Errorf("%w: %w", ErrIdenticalAddresses, err)
	case reasonZeroAddress:
		return fmt.Errorf("%w: %w", ErrZeroAddress, err)
	case reasonPairExists:
		return fmt.Errorf("%w: %w", ErrPairExists, err)
	case reasonForbidden:
		return fmt.Errorf("%w: %w", ErrForbidden, err)
	default:
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}
}
