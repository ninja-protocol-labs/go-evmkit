package v2

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// IFactory mirrors IUniswapV2Factory. No tx.Builder-taking Create* constructors here -- see
// CreateCreatePair, CreateSetFeeTo, and CreateSetFeeToSetter.
type IFactory interface {
	FeeTo(ctx context.Context, contract *types.Address, block string) (*types.Address, error)
	FeeToSetter(ctx context.Context, contract *types.Address, block string) (*types.Address, error)
	GetPair(ctx context.Context, contract, tokenA, tokenB *types.Address, block string) (*types.Address, error)
	AllPairs(ctx context.Context, contract *types.Address, index *big.Int, block string) (*types.Address, error)
	AllPairsLength(ctx context.Context, contract *types.Address, block string) (*big.Int, error)

	CreatePair(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SetFeeTo(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SetFeeToSetter(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

var _ IFactory = (*Factory)(nil)

// Factory calls any UniswapV2Factory contract over a Client; it holds no contract
// address, so one instance serves every network's deployment.
type Factory struct {
	c rpc.Client
}

// NewFactory returns a Factory calling factories via c.
func NewFactory(c rpc.Client) *Factory {
	return &Factory{
		c: c,
	}
}

// FeeTo returns contract's current protocol fee recipient (zero address if the fee switch is off).
func (f *Factory) FeeTo(ctx context.Context, contract *types.Address, block string) (*types.Address, error) {
	data, err := f.call(ctx, nil, contract, EncodeFeeTo(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: feeTo: %w", err)
	}
	addr, err := DecodeFeeTo(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: feeTo: %w", err)
	}
	return addr, nil
}

// FeeToSetter returns contract's current governance address, the only account allowed to call
// SetFeeTo/SetFeeToSetter.
func (f *Factory) FeeToSetter(ctx context.Context, contract *types.Address, block string) (*types.Address, error) {
	data, err := f.call(ctx, nil, contract, EncodeFeeToSetter(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: feeToSetter: %w", err)
	}
	addr, err := DecodeFeeToSetter(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: feeToSetter: %w", err)
	}
	return addr, nil
}

// GetPair returns the pair deployed for (tokenA, tokenB), or the zero address if none exists yet.
func (f *Factory) GetPair(ctx context.Context, contract, tokenA, tokenB *types.Address, block string) (*types.Address, error) {
	calldata, err := EncodeGetPair(tokenA, tokenB)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getPair: %w", err)
	}
	data, err := f.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getPair: %w", err)
	}
	addr, err := DecodeGetPair(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getPair: %w", err)
	}
	return addr, nil
}

// AllPairs returns the pair at index in contract's allPairs array.
func (f *Factory) AllPairs(ctx context.Context, contract *types.Address, index *big.Int, block string) (*types.Address, error) {
	calldata, err := EncodeAllPairs(index)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: allPairs: %w", err)
	}
	data, err := f.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: allPairs: %w", err)
	}
	addr, err := DecodeAllPairs(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: allPairs: %w", err)
	}
	return addr, nil
}

// AllPairsLength returns how many pairs contract has ever created.
func (f *Factory) AllPairsLength(ctx context.Context, contract *types.Address, block string) (*big.Int, error) {
	data, err := f.call(ctx, nil, contract, EncodeAllPairsLength(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: allPairsLength: %w", err)
	}
	n, err := DecodeAllPairsLength(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: allPairsLength: %w", err)
	}
	return n, nil
}

// CreateCreatePair builds an unsigned createPair(tokenA, tokenB) transaction via b, sent by sender.
func (f *Factory) CreateCreatePair(ctx context.Context, b tx.Builder, contract, sender, tokenA, tokenB *types.Address) (core.Transaction, error) {
	calldata, err := EncodeCreatePair(tokenA, tokenB)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createCreatePair: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, f.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createCreatePair: %w", f.wrapRevertError(err))
	}
	return unsigned, nil
}

// CallCreatePair simulates createPair(tokenA, tokenB) via eth_call, as sender, returning the
// pair address it would create -- without broadcasting anything.
func (f *Factory) CallCreatePair(ctx context.Context, contract, sender, tokenA, tokenB *types.Address, block string) (*types.Address, error) {
	calldata, err := EncodeCreatePair(tokenA, tokenB)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callCreatePair: %w", err)
	}
	data, err := f.call(ctx, sender, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callCreatePair: %w", err)
	}
	addr, err := DecodeCreatePair(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callCreatePair: %w", err)
	}
	return addr, nil
}

// CreatePair broadcasts an already-signed createPair transaction. The new pair's address comes
// back via the PairCreated event (ExtractPairCreated), not a return value -- a mined
// transaction's return value isn't retrievable.
func (f *Factory) CreatePair(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createPair: %w", err)
	}
	hash, err := tx.Broadcast(ctx, f.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createPair: %w", err)
	}
	return hash, nil
}

// CreateSetFeeTo builds an unsigned setFeeTo(feeTo) transaction via b, sent by sender -- which
// must be contract's current feeToSetter, or the transaction reverts with ErrForbidden.
func (f *Factory) CreateSetFeeTo(ctx context.Context, b tx.Builder, contract, sender, feeTo *types.Address) (core.Transaction, error) {
	calldata, err := EncodeSetFeeTo(feeTo)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSetFeeTo: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, f.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSetFeeTo: %w", f.wrapRevertError(err))
	}
	return unsigned, nil
}

// SetFeeTo broadcasts an already-signed setFeeTo transaction.
func (f *Factory) SetFeeTo(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: setFeeTo: %w", err)
	}
	hash, err := tx.Broadcast(ctx, f.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: setFeeTo: %w", err)
	}
	return hash, nil
}

// CreateSetFeeToSetter builds an unsigned setFeeToSetter(feeToSetter) transaction via b, sent by
// sender -- which must be contract's current feeToSetter, or the transaction reverts with
// ErrForbidden.
func (f *Factory) CreateSetFeeToSetter(ctx context.Context, b tx.Builder, contract, sender, feeToSetter *types.Address) (core.Transaction, error) {
	calldata, err := EncodeSetFeeToSetter(feeToSetter)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSetFeeToSetter: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, f.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSetFeeToSetter: %w", f.wrapRevertError(err))
	}
	return unsigned, nil
}

// SetFeeToSetter broadcasts an already-signed setFeeToSetter transaction.
func (f *Factory) SetFeeToSetter(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: setFeeToSetter: %w", err)
	}
	hash, err := tx.Broadcast(ctx, f.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: setFeeToSetter: %w", err)
	}
	return hash, nil
}

// call eth_calls contract with calldata, decoding a reverted require(cond, reason) to its sentinel.
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
