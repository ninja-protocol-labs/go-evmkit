// Package multicall3 is a client for the Multicall3 contract (https://github.com/mds1/multicall),
// mirroring solidity/interfaces/IMulticall3.sol.
package multicall3

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

// DefaultMulticall3Address is the address Multicall3 is conventionally
// deployed at on most EVM chains (https://github.com/mds1/multicall).
var DefaultMulticall3Address = types.MustNewAddressFromHex("0xcA11bde05977b3631167028862bE2a173976CA11")

// IMulticall3 mirrors IMulticall3.sol.
type IMulticall3 interface {
	CallAggregate(ctx context.Context, opts *bind.CallOptions, req *AggregateRequest) (*AggregateResult, error)
	CallTryAggregate(ctx context.Context, opts *bind.CallOptions, req *TryAggregateRequest) ([]Result, error)
	CallBlockAndAggregate(ctx context.Context, opts *bind.CallOptions, req *BlockAndAggregateRequest) (*BlockResults, error)
	CallTryBlockAndAggregate(ctx context.Context, opts *bind.CallOptions, req *TryBlockAndAggregateRequest) (*BlockResults, error)
	CallAggregate3(ctx context.Context, opts *bind.CallOptions, req *Aggregate3Request) ([]Result, error)
	CallAggregate3Value(ctx context.Context, opts *bind.CallOptions, req *Aggregate3ValueRequest) ([]Result, error)

	Aggregate(ctx context.Context, opts *bind.TransactOptions, req *AggregateRequest) (core.Transaction, *types.Hash, error)
	TryAggregate(ctx context.Context, opts *bind.TransactOptions, req *TryAggregateRequest) (core.Transaction, *types.Hash, error)
	BlockAndAggregate(ctx context.Context, opts *bind.TransactOptions, req *BlockAndAggregateRequest) (core.Transaction, *types.Hash, error)
	TryBlockAndAggregate(ctx context.Context, opts *bind.TransactOptions, req *TryBlockAndAggregateRequest) (core.Transaction, *types.Hash, error)
	Aggregate3(ctx context.Context, opts *bind.TransactOptions, req *Aggregate3Request) (core.Transaction, *types.Hash, error)
	Aggregate3Value(ctx context.Context, opts *bind.TransactOptions, req *Aggregate3ValueRequest) (core.Transaction, *types.Hash, error)

	GetBasefee(ctx context.Context, opts *bind.CallOptions, req *GetBasefeeRequest) (*big.Int, error)
	GetBlockHash(ctx context.Context, opts *bind.CallOptions, req *GetBlockHashRequest) (*types.Hash, error)
	GetBlockNumber(ctx context.Context, opts *bind.CallOptions, req *GetBlockNumberRequest) (*big.Int, error)
	GetChainID(ctx context.Context, opts *bind.CallOptions, req *GetChainIDRequest) (*big.Int, error)
	GetCurrentBlockCoinbase(ctx context.Context, opts *bind.CallOptions, req *GetCurrentBlockCoinbaseRequest) (*types.Address, error)
	GetCurrentBlockDifficulty(ctx context.Context, opts *bind.CallOptions, req *GetCurrentBlockDifficultyRequest) (*big.Int, error)
	GetCurrentBlockGasLimit(ctx context.Context, opts *bind.CallOptions, req *GetCurrentBlockGasLimitRequest) (*big.Int, error)
	GetCurrentBlockTimestamp(ctx context.Context, opts *bind.CallOptions, req *GetCurrentBlockTimestampRequest) (*big.Int, error)
	GetEthBalance(ctx context.Context, opts *bind.CallOptions, req *GetEthBalanceRequest) (*big.Int, error)
	GetLastBlockHash(ctx context.Context, opts *bind.CallOptions, req *GetLastBlockHashRequest) (*types.Hash, error)
}

var _ IMulticall3 = (*Multicall3)(nil)

// Multicall3 calls any Multicall3 deployment over a Client.
type Multicall3 struct {
	c rpc.Client
}

// New returns a Multicall3 calling contracts via cli.
func New(cli rpc.Client) *Multicall3 {
	return &Multicall3{
		c: cli,
	}
}

// CallAggregate runs aggregate(calls) via eth_call; any failing call reverts the whole call with ErrCallFailed.
func (m *Multicall3) CallAggregate(ctx context.Context, opts *bind.CallOptions, req *AggregateRequest) (*AggregateResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("multicall3: callAggregate: %w", err)
	}

	calldata, err := EncodeAggregate(req.Calls)
	if err != nil {
		return nil, fmt.Errorf("multicall3: callAggregate: %w", err)
	}

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("multicall3: callAggregate: %w", m.wrapRevertError(err))
	}

	out, err := DecodeAggregate(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: callAggregate: %w", err)
	}

	return out, nil
}

// CallTryAggregate runs tryAggregate(requireSuccess, calls) via eth_call, returning each call's result.
func (m *Multicall3) CallTryAggregate(ctx context.Context, opts *bind.CallOptions, req *TryAggregateRequest) ([]Result, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("multicall3: callTryAggregate: %w", err)
	}

	calldata, err := EncodeTryAggregate(req.RequireSuccess, req.Calls)
	if err != nil {
		return nil, fmt.Errorf("multicall3: callTryAggregate: %w", err)
	}

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("multicall3: callTryAggregate: %w", m.wrapRevertError(err))
	}

	out, err := DecodeResults(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: callTryAggregate: %w", err)
	}

	return out, nil
}

// CallBlockAndAggregate runs blockAndAggregate(calls) via eth_call: tryBlockAndAggregate with requireSuccess true.
func (m *Multicall3) CallBlockAndAggregate(ctx context.Context, opts *bind.CallOptions, req *BlockAndAggregateRequest) (*BlockResults, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("multicall3: callBlockAndAggregate: %w", err)
	}

	calldata, err := EncodeBlockAndAggregate(req.Calls)
	if err != nil {
		return nil, fmt.Errorf("multicall3: callBlockAndAggregate: %w", err)
	}

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("multicall3: callBlockAndAggregate: %w", m.wrapRevertError(err))
	}

	out, err := DecodeBlockAndAggregate(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: callBlockAndAggregate: %w", err)
	}

	return out, nil
}

// CallTryBlockAndAggregate runs tryBlockAndAggregate(requireSuccess, calls) via eth_call, returning the block and each call's result.
func (m *Multicall3) CallTryBlockAndAggregate(ctx context.Context, opts *bind.CallOptions, req *TryBlockAndAggregateRequest) (*BlockResults, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("multicall3: callTryBlockAndAggregate: %w", err)
	}

	calldata, err := EncodeTryBlockAndAggregate(req.RequireSuccess, req.Calls)
	if err != nil {
		return nil, fmt.Errorf("multicall3: callTryBlockAndAggregate: %w", err)
	}

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("multicall3: callTryBlockAndAggregate: %w", m.wrapRevertError(err))
	}

	out, err := DecodeBlockAndAggregate(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: callTryBlockAndAggregate: %w", err)
	}

	return out, nil
}

// CallAggregate3 runs aggregate3(calls) via eth_call; a failing call without AllowFailure reverts with ErrCallFailed.
func (m *Multicall3) CallAggregate3(ctx context.Context, opts *bind.CallOptions, req *Aggregate3Request) ([]Result, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("multicall3: callAggregate3: %w", err)
	}

	calldata, err := EncodeAggregate3(req.Calls)
	if err != nil {
		return nil, fmt.Errorf("multicall3: callAggregate3: %w", err)
	}

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("multicall3: callAggregate3: %w", m.wrapRevertError(err))
	}

	out, err := DecodeResults(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: callAggregate3: %w", err)
	}

	return out, nil
}

// CallAggregate3Value runs aggregate3Value(calls) via eth_call with msg.value set to req.TotalValue();
// opts.Value must be nil.
func (m *Multicall3) CallAggregate3Value(ctx context.Context, opts *bind.CallOptions, req *Aggregate3ValueRequest) ([]Result, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("multicall3: callAggregate3Value: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, fmt.Errorf("multicall3: callAggregate3Value: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeAggregate3Value(req.Calls)
	if err != nil {
		return nil, fmt.Errorf("multicall3: callAggregate3Value: %w", err)
	}

	callOpts := bind.NewCallOptions()
	if opts != nil {
		*callOpts = *opts
	}
	callOpts.Value = req.TotalValue()

	data, err := bind.Call(ctx, m.c, callOpts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("multicall3: callAggregate3Value: %w", m.wrapRevertError(err))
	}

	out, err := DecodeResults(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: callAggregate3Value: %w", err)
	}

	return out, nil
}

// Aggregate sends aggregate(calls) on req.Contract, signed by opts.Signer; any failing call reverts the whole transaction.
func (m *Multicall3) Aggregate(ctx context.Context, opts *bind.TransactOptions, req *AggregateRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("multicall3: aggregate: %w", err)
	}

	calldata, err := EncodeAggregate(req.Calls)
	if err != nil {
		return nil, nil, fmt.Errorf("multicall3: aggregate: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("multicall3: aggregate: %w", m.wrapRevertError(err))
	}

	return signed, hash, nil
}

// TryAggregate sends tryAggregate(requireSuccess, calls) on req.Contract, signed by opts.Signer.
func (m *Multicall3) TryAggregate(ctx context.Context, opts *bind.TransactOptions, req *TryAggregateRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("multicall3: tryAggregate: %w", err)
	}

	calldata, err := EncodeTryAggregate(req.RequireSuccess, req.Calls)
	if err != nil {
		return nil, nil, fmt.Errorf("multicall3: tryAggregate: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("multicall3: tryAggregate: %w", m.wrapRevertError(err))
	}

	return signed, hash, nil
}

// BlockAndAggregate sends blockAndAggregate(calls) on req.Contract, signed by opts.Signer.
func (m *Multicall3) BlockAndAggregate(ctx context.Context, opts *bind.TransactOptions, req *BlockAndAggregateRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("multicall3: blockAndAggregate: %w", err)
	}

	calldata, err := EncodeBlockAndAggregate(req.Calls)
	if err != nil {
		return nil, nil, fmt.Errorf("multicall3: blockAndAggregate: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("multicall3: blockAndAggregate: %w", m.wrapRevertError(err))
	}

	return signed, hash, nil
}

// TryBlockAndAggregate sends tryBlockAndAggregate(requireSuccess, calls) on req.Contract, signed by opts.Signer.
func (m *Multicall3) TryBlockAndAggregate(ctx context.Context, opts *bind.TransactOptions, req *TryBlockAndAggregateRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("multicall3: tryBlockAndAggregate: %w", err)
	}

	calldata, err := EncodeTryBlockAndAggregate(req.RequireSuccess, req.Calls)
	if err != nil {
		return nil, nil, fmt.Errorf("multicall3: tryBlockAndAggregate: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("multicall3: tryBlockAndAggregate: %w", m.wrapRevertError(err))
	}

	return signed, hash, nil
}

// Aggregate3 sends aggregate3(calls) on req.Contract, signed by opts.Signer.
func (m *Multicall3) Aggregate3(ctx context.Context, opts *bind.TransactOptions, req *Aggregate3Request) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("multicall3: aggregate3: %w", err)
	}

	calldata, err := EncodeAggregate3(req.Calls)
	if err != nil {
		return nil, nil, fmt.Errorf("multicall3: aggregate3: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("multicall3: aggregate3: %w", m.wrapRevertError(err))
	}

	return signed, hash, nil
}

// Aggregate3Value sends aggregate3Value(calls) on req.Contract with msg.value set to req.TotalValue(),
// signed by opts.Signer; opts.Value must be nil.
func (m *Multicall3) Aggregate3Value(ctx context.Context, opts *bind.TransactOptions, req *Aggregate3ValueRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("multicall3: aggregate3Value: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("multicall3: aggregate3Value: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeAggregate3Value(req.Calls)
	if err != nil {
		return nil, nil, fmt.Errorf("multicall3: aggregate3Value: %w", err)
	}

	var txOpts *bind.TransactOptions
	if opts != nil {
		o := *opts
		o.Value = req.TotalValue()
		txOpts = &o
	}

	signed, hash, err := bind.Transact(ctx, m.c, txOpts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("multicall3: aggregate3Value: %w", m.wrapRevertError(err))
	}

	return signed, hash, nil
}

// GetBasefee returns the block's base fee; a revert maps to ErrBasefeeNotImplemented on chains without BASEFEE.
func (m *Multicall3) GetBasefee(ctx context.Context, opts *bind.CallOptions, req *GetBasefeeRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("multicall3: getBasefee: %w", err)
	}

	calldata := EncodeGetBasefee()

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		if _, ok := errors.AsType[*rpc.ResponseError](err); ok {
			return nil, fmt.Errorf("multicall3: getBasefee: %w: %w", ErrBasefeeNotImplemented, err)
		}
		return nil, fmt.Errorf("multicall3: getBasefee: %w", err)
	}

	out, err := DecodeGetBasefee(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getBasefee: %w", err)
	}

	return out, nil
}

// GetBlockHash returns blockhash(req.BlockNumber), zero outside the last 256 blocks.
func (m *Multicall3) GetBlockHash(ctx context.Context, opts *bind.CallOptions, req *GetBlockHashRequest) (*types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("multicall3: getBlockHash: %w", err)
	}

	calldata, err := EncodeGetBlockHash(req.BlockNumber)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getBlockHash: %w", err)
	}

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getBlockHash: %w", m.wrapRevertError(err))
	}

	out, err := DecodeGetBlockHash(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getBlockHash: %w", err)
	}

	return out, nil
}

// GetBlockNumber returns block.number.
func (m *Multicall3) GetBlockNumber(ctx context.Context, opts *bind.CallOptions, req *GetBlockNumberRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("multicall3: getBlockNumber: %w", err)
	}

	calldata := EncodeGetBlockNumber()

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getBlockNumber: %w", m.wrapRevertError(err))
	}

	out, err := DecodeGetBlockNumber(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getBlockNumber: %w", err)
	}

	return out, nil
}

// GetChainID returns block.chainid.
func (m *Multicall3) GetChainID(ctx context.Context, opts *bind.CallOptions, req *GetChainIDRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("multicall3: getChainId: %w", err)
	}

	calldata := EncodeGetChainID()

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getChainId: %w", m.wrapRevertError(err))
	}

	out, err := DecodeGetChainID(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getChainId: %w", err)
	}

	return out, nil
}

// GetCurrentBlockCoinbase returns block.coinbase.
func (m *Multicall3) GetCurrentBlockCoinbase(ctx context.Context, opts *bind.CallOptions, req *GetCurrentBlockCoinbaseRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("multicall3: getCurrentBlockCoinbase: %w", err)
	}

	calldata := EncodeGetCurrentBlockCoinbase()

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getCurrentBlockCoinbase: %w", m.wrapRevertError(err))
	}

	out, err := DecodeGetCurrentBlockCoinbase(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getCurrentBlockCoinbase: %w", err)
	}

	return out, nil
}

// GetCurrentBlockDifficulty returns block.difficulty (prevrandao after the merge).
func (m *Multicall3) GetCurrentBlockDifficulty(ctx context.Context, opts *bind.CallOptions, req *GetCurrentBlockDifficultyRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("multicall3: getCurrentBlockDifficulty: %w", err)
	}

	calldata := EncodeGetCurrentBlockDifficulty()

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getCurrentBlockDifficulty: %w", m.wrapRevertError(err))
	}

	out, err := DecodeGetCurrentBlockDifficulty(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getCurrentBlockDifficulty: %w", err)
	}

	return out, nil
}

// GetCurrentBlockGasLimit returns block.gaslimit.
func (m *Multicall3) GetCurrentBlockGasLimit(ctx context.Context, opts *bind.CallOptions, req *GetCurrentBlockGasLimitRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("multicall3: getCurrentBlockGasLimit: %w", err)
	}

	calldata := EncodeGetCurrentBlockGasLimit()

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getCurrentBlockGasLimit: %w", m.wrapRevertError(err))
	}

	out, err := DecodeGetCurrentBlockGasLimit(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getCurrentBlockGasLimit: %w", err)
	}

	return out, nil
}

// GetCurrentBlockTimestamp returns block.timestamp.
func (m *Multicall3) GetCurrentBlockTimestamp(ctx context.Context, opts *bind.CallOptions, req *GetCurrentBlockTimestampRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("multicall3: getCurrentBlockTimestamp: %w", err)
	}

	calldata := EncodeGetCurrentBlockTimestamp()

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getCurrentBlockTimestamp: %w", m.wrapRevertError(err))
	}

	out, err := DecodeGetCurrentBlockTimestamp(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getCurrentBlockTimestamp: %w", err)
	}

	return out, nil
}

// GetEthBalance returns req.Addr's native balance.
func (m *Multicall3) GetEthBalance(ctx context.Context, opts *bind.CallOptions, req *GetEthBalanceRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("multicall3: getEthBalance: %w", err)
	}

	calldata, err := EncodeGetEthBalance(req.Addr)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getEthBalance: %w", err)
	}

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getEthBalance: %w", m.wrapRevertError(err))
	}

	out, err := DecodeGetEthBalance(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getEthBalance: %w", err)
	}

	return out, nil
}

// GetLastBlockHash returns blockhash(block.number - 1).
func (m *Multicall3) GetLastBlockHash(ctx context.Context, opts *bind.CallOptions, req *GetLastBlockHashRequest) (*types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("multicall3: getLastBlockHash: %w", err)
	}

	calldata := EncodeGetLastBlockHash()

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getLastBlockHash: %w", m.wrapRevertError(err))
	}

	out, err := DecodeGetLastBlockHash(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getLastBlockHash: %w", err)
	}

	return out, nil
}

// wrapRevertError maps a known Multicall3 revert reason in err to its sentinel.
func (_ *Multicall3) wrapRevertError(err error) error {
	var rpcErr *rpc.ResponseError
	if !errors.As(err, &rpcErr) || len(rpcErr.Data) == 0 {
		return err
	}

	var hexData string
	if jsonErr := json.Unmarshal(rpcErr.Data, &hexData); jsonErr != nil {
		return err
	}

	data, decErr := encoding.Hex.Decode(hexData)
	if decErr != nil || len(data) == 0 {
		return err
	}

	reason, revertErr := abi.DecodeRevert(data)
	if revertErr != nil {
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}

	switch reason {
	case ErrCallFailedString:
		return fmt.Errorf("%w: %w", ErrCallFailed, err)
	case ErrValueMismatchString:
		return fmt.Errorf("%w: %w", ErrValueMismatch, err)
	default:
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}
}
