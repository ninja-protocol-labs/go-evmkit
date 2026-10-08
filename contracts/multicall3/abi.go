// Package multicall3 encodes calls to and decodes results from the Multicall3
// contract (https://github.com/mds1/multicall). Go types mirror the structs
// in solidity/interfaces/IMulticall3.sol.
package multicall3

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// EncodeAggregate returns the calldata for aggregate(calls).
func EncodeAggregate(calls []Call) ([]byte, error) {
	return aggregate.EncodeCall(callTuples(calls))
}

// EncodeBlockAndAggregate returns the calldata for blockAndAggregate(calls).
func EncodeBlockAndAggregate(calls []Call) ([]byte, error) {
	return blockAndAggregate.EncodeCall(callTuples(calls))
}

// EncodeTryAggregate returns the calldata for tryAggregate(requireSuccess, calls).
func EncodeTryAggregate(requireSuccess bool, calls []Call) ([]byte, error) {
	return tryAggregate.EncodeCall(requireSuccess, callTuples(calls))
}

// EncodeTryBlockAndAggregate returns the calldata for tryBlockAndAggregate(requireSuccess, calls).
func EncodeTryBlockAndAggregate(requireSuccess bool, calls []Call) ([]byte, error) {
	return tryBlockAndAggregate.EncodeCall(requireSuccess, callTuples(calls))
}

// EncodeAggregate3 returns the calldata for aggregate3(calls).
func EncodeAggregate3(calls []Call3) ([]byte, error) {
	tuples := make([][]any, len(calls))
	for i, c := range calls {
		tuples[i] = c.tuple()
	}
	return aggregate3.EncodeCall(tuples)
}

// EncodeAggregate3Value returns the calldata for aggregate3Value(calls). The
// transaction's value must equal the sum of the calls' values.
func EncodeAggregate3Value(calls []Call3Value) ([]byte, error) {
	tuples := make([][]any, len(calls))
	for i, c := range calls {
		tuples[i] = c.tuple()
	}
	return aggregate3Value.EncodeCall(tuples)
}

// DecodeResults decodes the return data of aggregate3, aggregate3Value and tryAggregate.
func DecodeResults(data []byte) ([]Result, error) {
	vals, err := aggregate3.DecodeReturn(data)
	if err != nil {
		return nil, err
	}
	return toResults(vals[0])
}

// DecodeAggregate decodes the return data of aggregate.
func DecodeAggregate(data []byte) (*AggregateResult, error) {
	vals, err := aggregate.DecodeReturn(data)
	if err != nil {
		return nil, err
	}

	blockNumber, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("multicall3: decode aggregate: blockNumber: unexpected type %T", vals[0])
	}

	elems, ok := vals[1].([]any)
	if !ok {
		return nil, fmt.Errorf("multicall3: decode aggregate: returnData: unexpected type %T", vals[1])
	}
	returnData := make([][]byte, len(elems))
	for i, e := range elems {
		b, ok := e.([]byte)
		if !ok {
			return nil, fmt.Errorf("multicall3: decode aggregate: returnData[%d]: unexpected type %T", i, e)
		}
		returnData[i] = b
	}

	result := NewAggregateResult(blockNumber, returnData)
	return &result, nil
}

// DecodeBlockAndAggregate decodes the return data of blockAndAggregate and tryBlockAndAggregate.
func DecodeBlockAndAggregate(data []byte) (*BlockResults, error) {
	vals, err := blockAndAggregate.DecodeReturn(data)
	if err != nil {
		return nil, err
	}

	blockNumber, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("multicall3: decode block and aggregate: blockNumber: unexpected type %T", vals[0])
	}

	blockHashBytes, ok := vals[1].([]byte)
	if !ok {
		return nil, fmt.Errorf("multicall3: decode block and aggregate: blockHash: unexpected type %T", vals[1])
	}

	results, err := toResults(vals[2])
	if err != nil {
		return nil, fmt.Errorf("multicall3: decode block and aggregate: %w", err)
	}

	result := NewBlockResults(blockNumber, types.NewHashFromBytes(blockHashBytes), results)
	return &result, nil
}

// EncodeGetBasefee returns the calldata for getBasefee().
func EncodeGetBasefee() []byte {
	return selectorOnly(getBasefee)
}

// DecodeGetBasefee decodes the return data of getBasefee.
func DecodeGetBasefee(data []byte) (*big.Int, error) {
	return decodeUint256(getBasefee, data)
}

// EncodeGetBlockHash returns the calldata for getBlockHash(blockNumber).
func EncodeGetBlockHash(blockNumber *big.Int) ([]byte, error) {
	return getBlockHash.EncodeCall(blockNumber)
}

// DecodeGetBlockHash decodes the return data of getBlockHash.
func DecodeGetBlockHash(data []byte) (*types.Hash, error) {
	return decodeHash(getBlockHash, data)
}

// EncodeGetBlockNumber returns the calldata for getBlockNumber().
func EncodeGetBlockNumber() []byte {
	return selectorOnly(getBlockNumber)
}

// DecodeGetBlockNumber decodes the return data of getBlockNumber.
func DecodeGetBlockNumber(data []byte) (*big.Int, error) {
	return decodeUint256(getBlockNumber, data)
}

// EncodeGetChainID returns the calldata for getChainId().
func EncodeGetChainID() []byte {
	return selectorOnly(getChainId)
}

// DecodeGetChainID decodes the return data of getChainId.
func DecodeGetChainID(data []byte) (*big.Int, error) {
	return decodeUint256(getChainId, data)
}

// EncodeGetCurrentBlockCoinbase returns the calldata for getCurrentBlockCoinbase().
func EncodeGetCurrentBlockCoinbase() []byte {
	return selectorOnly(getCurrentBlockCoinbase)
}

// DecodeGetCurrentBlockCoinbase decodes the return data of getCurrentBlockCoinbase.
func DecodeGetCurrentBlockCoinbase(data []byte) (*types.Address, error) {
	vals, err := getCurrentBlockCoinbase.DecodeReturn(data)
	if err != nil {
		return nil, err
	}
	addr, ok := vals[0].(*types.Address)
	if !ok {
		return nil, fmt.Errorf("multicall3: decode getCurrentBlockCoinbase: unexpected type %T", vals[0])
	}
	return addr, nil
}

// EncodeGetCurrentBlockDifficulty returns the calldata for getCurrentBlockDifficulty().
func EncodeGetCurrentBlockDifficulty() []byte {
	return selectorOnly(getCurrentBlockDiff)
}

// DecodeGetCurrentBlockDifficulty decodes the return data of getCurrentBlockDifficulty.
func DecodeGetCurrentBlockDifficulty(data []byte) (*big.Int, error) {
	return decodeUint256(getCurrentBlockDiff, data)
}

// EncodeGetCurrentBlockGasLimit returns the calldata for getCurrentBlockGasLimit().
func EncodeGetCurrentBlockGasLimit() []byte {
	return selectorOnly(getCurrentBlockGasLimit)
}

// DecodeGetCurrentBlockGasLimit decodes the return data of getCurrentBlockGasLimit.
func DecodeGetCurrentBlockGasLimit(data []byte) (*big.Int, error) {
	return decodeUint256(getCurrentBlockGasLimit, data)
}

// EncodeGetCurrentBlockTimestamp returns the calldata for getCurrentBlockTimestamp().
func EncodeGetCurrentBlockTimestamp() []byte {
	return selectorOnly(getCurrentBlockTime)
}

// DecodeGetCurrentBlockTimestamp decodes the return data of getCurrentBlockTimestamp.
func DecodeGetCurrentBlockTimestamp(data []byte) (*big.Int, error) {
	return decodeUint256(getCurrentBlockTime, data)
}

// EncodeGetEthBalance returns the calldata for getEthBalance(addr).
func EncodeGetEthBalance(addr *types.Address) ([]byte, error) {
	return getEthBalance.EncodeCall(addr)
}

// DecodeGetEthBalance decodes the return data of getEthBalance.
func DecodeGetEthBalance(data []byte) (*big.Int, error) {
	return decodeUint256(getEthBalance, data)
}

// EncodeGetLastBlockHash returns the calldata for getLastBlockHash().
func EncodeGetLastBlockHash() []byte {
	return selectorOnly(getLastBlockHash)
}

// DecodeGetLastBlockHash decodes the return data of getLastBlockHash.
func DecodeGetLastBlockHash(data []byte) (*types.Hash, error) {
	return decodeHash(getLastBlockHash, data)
}

func selectorOnly(fn *abi.Function) []byte {
	sel := fn.Selector()
	return sel[:]
}

func decodeUint256(fn *abi.Function, data []byte) (*big.Int, error) {
	vals, err := fn.DecodeReturn(data)
	if err != nil {
		return nil, err
	}
	n, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("multicall3: decode %s: unexpected type %T", fn.Name, vals[0])
	}
	return n, nil
}

func decodeHash(fn *abi.Function, data []byte) (*types.Hash, error) {
	vals, err := fn.DecodeReturn(data)
	if err != nil {
		return nil, err
	}
	b, ok := vals[0].([]byte)
	if !ok {
		return nil, fmt.Errorf("multicall3: decode %s: unexpected type %T", fn.Name, vals[0])
	}
	return types.NewHashFromBytes(b), nil
}

func callTuples(calls []Call) [][]any {
	tuples := make([][]any, len(calls))
	for i, c := range calls {
		tuples[i] = c.tuple()
	}
	return tuples
}

func toResults(v any) ([]Result, error) {
	elems, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("multicall3: decode results: unexpected type %T", v)
	}
	results := make([]Result, len(elems))
	for i, e := range elems {
		fields, ok := e.([]any)
		if !ok {
			return nil, fmt.Errorf("multicall3: decode results[%d]: unexpected type %T", i, e)
		}
		success, ok := fields[0].(bool)
		if !ok {
			return nil, fmt.Errorf("multicall3: decode results[%d]: success: unexpected type %T", i, fields[0])
		}
		returnData, ok := fields[1].([]byte)
		if !ok {
			return nil, fmt.Errorf("multicall3: decode results[%d]: returnData: unexpected type %T", i, fields[1])
		}
		results[i] = NewResult(success, returnData)
	}
	return results, nil
}
