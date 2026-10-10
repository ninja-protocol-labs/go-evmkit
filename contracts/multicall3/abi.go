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

var (
	callT       = abi.Tuple(abi.Address, abi.Bytes)
	call3T      = abi.Tuple(abi.Address, abi.Bool, abi.Bytes)
	call3ValueT = abi.Tuple(abi.Address, abi.Bool, abi.Uint256, abi.Bytes)
	resultT     = abi.Tuple(abi.Bool, abi.Bytes)
)

var (
	aggregate            = abi.NewFunction("aggregate", abi.NewTypes(abi.Slice(callT)), abi.NewTypes(abi.Uint256, abi.Slice(abi.Bytes)))
	aggregate3           = abi.NewFunction("aggregate3", abi.NewTypes(abi.Slice(call3T)), abi.NewTypes(abi.Slice(resultT)))
	aggregate3Value      = abi.NewFunction("aggregate3Value", abi.NewTypes(abi.Slice(call3ValueT)), abi.NewTypes(abi.Slice(resultT)))
	blockAndAggregate    = abi.NewFunction("blockAndAggregate", abi.NewTypes(abi.Slice(callT)), abi.NewTypes(abi.Uint256, abi.Bytes32, abi.Slice(resultT)))
	tryAggregate         = abi.NewFunction("tryAggregate", abi.NewTypes(abi.Bool, abi.Slice(callT)), abi.NewTypes(abi.Slice(resultT)))
	tryBlockAndAggregate = abi.NewFunction("tryBlockAndAggregate", abi.NewTypes(abi.Bool, abi.Slice(callT)), abi.NewTypes(abi.Uint256, abi.Bytes32, abi.Slice(resultT)))

	getBasefee              = abi.NewFunction("getBasefee", nil, abi.NewTypes(abi.Uint256))
	getBlockHash            = abi.NewFunction("getBlockHash", abi.NewTypes(abi.Uint256), abi.NewTypes(abi.Bytes32))
	getBlockNumber          = abi.NewFunction("getBlockNumber", nil, abi.NewTypes(abi.Uint256))
	getChainId              = abi.NewFunction("getChainId", nil, abi.NewTypes(abi.Uint256))
	getCurrentBlockCoinbase = abi.NewFunction("getCurrentBlockCoinbase", nil, abi.NewTypes(abi.Address))
	getCurrentBlockDiff     = abi.NewFunction("getCurrentBlockDifficulty", nil, abi.NewTypes(abi.Uint256))
	getCurrentBlockGasLimit = abi.NewFunction("getCurrentBlockGasLimit", nil, abi.NewTypes(abi.Uint256))
	getCurrentBlockTime     = abi.NewFunction("getCurrentBlockTimestamp", nil, abi.NewTypes(abi.Uint256))
	getEthBalance           = abi.NewFunction("getEthBalance", abi.NewTypes(abi.Address), abi.NewTypes(abi.Uint256))
	getLastBlockHash        = abi.NewFunction("getLastBlockHash", nil, abi.NewTypes(abi.Bytes32))
)

// EncodeAggregate returns the calldata for aggregate(calls).
func EncodeAggregate(calls []Call) ([]byte, error) {
	data, err := aggregate.EncodeCall(callTuples(calls))
	if err != nil {
		return nil, fmt.Errorf("multicall3: encodeAggregate: %w", err)
	}
	return data, nil
}

// EncodeBlockAndAggregate returns the calldata for blockAndAggregate(calls).
func EncodeBlockAndAggregate(calls []Call) ([]byte, error) {
	data, err := blockAndAggregate.EncodeCall(callTuples(calls))
	if err != nil {
		return nil, fmt.Errorf("multicall3: encodeBlockAndAggregate: %w", err)
	}
	return data, nil
}

// EncodeTryAggregate returns the calldata for tryAggregate(requireSuccess, calls).
func EncodeTryAggregate(requireSuccess bool, calls []Call) ([]byte, error) {
	data, err := tryAggregate.EncodeCall(requireSuccess, callTuples(calls))
	if err != nil {
		return nil, fmt.Errorf("multicall3: encodeTryAggregate: %w", err)
	}
	return data, nil
}

// EncodeTryBlockAndAggregate returns the calldata for tryBlockAndAggregate(requireSuccess, calls).
func EncodeTryBlockAndAggregate(requireSuccess bool, calls []Call) ([]byte, error) {
	data, err := tryBlockAndAggregate.EncodeCall(requireSuccess, callTuples(calls))
	if err != nil {
		return nil, fmt.Errorf("multicall3: encodeTryBlockAndAggregate: %w", err)
	}
	return data, nil
}

// EncodeAggregate3 returns the calldata for aggregate3(calls).
func EncodeAggregate3(calls []Call3) ([]byte, error) {
	tuples := make([][]any, len(calls))
	for i, c := range calls {
		tuples[i] = c.tuple()
	}
	data, err := aggregate3.EncodeCall(tuples)
	if err != nil {
		return nil, fmt.Errorf("multicall3: encodeAggregate3: %w", err)
	}
	return data, nil
}

// EncodeAggregate3Value returns the calldata for aggregate3Value(calls). The
// transaction's value must equal the sum of the calls' values.
func EncodeAggregate3Value(calls []Call3Value) ([]byte, error) {
	tuples := make([][]any, len(calls))
	for i, c := range calls {
		tuples[i] = c.tuple()
	}
	data, err := aggregate3Value.EncodeCall(tuples)
	if err != nil {
		return nil, fmt.Errorf("multicall3: encodeAggregate3Value: %w", err)
	}
	return data, nil
}

// DecodeResults decodes the return data of aggregate3, aggregate3Value and tryAggregate.
func DecodeResults(data []byte) ([]Result, error) {
	vals, err := aggregate3.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: decodeResults: %w", err)
	}
	results, err := toResults(vals[0])
	if err != nil {
		return nil, fmt.Errorf("multicall3: decodeResults: %w", err)
	}
	return results, nil
}

// DecodeAggregate decodes the return data of aggregate.
func DecodeAggregate(data []byte) (*AggregateResult, error) {
	vals, err := aggregate.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: decodeAggregate: %w", err)
	}

	blockNumber, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("multicall3: decodeAggregate: blockNumber: unexpected type %T", vals[0])
	}

	elems, ok := vals[1].([]any)
	if !ok {
		return nil, fmt.Errorf("multicall3: decodeAggregate: returnData: unexpected type %T", vals[1])
	}
	returnData := make([][]byte, len(elems))
	for i, e := range elems {
		b, ok := e.([]byte)
		if !ok {
			return nil, fmt.Errorf("multicall3: decodeAggregate: returnData[%d]: unexpected type %T", i, e)
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
		return nil, fmt.Errorf("multicall3: decodeBlockAndAggregate: %w", err)
	}

	blockNumber, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("multicall3: decodeBlockAndAggregate: blockNumber: unexpected type %T", vals[0])
	}

	blockHashBytes, ok := vals[1].([]byte)
	if !ok {
		return nil, fmt.Errorf("multicall3: decodeBlockAndAggregate: blockHash: unexpected type %T", vals[1])
	}

	results, err := toResults(vals[2])
	if err != nil {
		return nil, fmt.Errorf("multicall3: decodeBlockAndAggregate: %w", err)
	}

	result := NewBlockResults(blockNumber, types.NewHashFromBytes(blockHashBytes), results)
	return &result, nil
}

// EncodeGetBasefee returns the calldata for getBasefee().
func EncodeGetBasefee() []byte {
	return getBasefee.SelectorBytes()
}

// DecodeGetBasefee decodes the return data of getBasefee.
func DecodeGetBasefee(data []byte) (*big.Int, error) {
	n, err := getBasefee.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: decodeGetBasefee: %w", err)
	}
	return n, nil
}

// EncodeGetBlockHash returns the calldata for getBlockHash(blockNumber).
func EncodeGetBlockHash(blockNumber *big.Int) ([]byte, error) {
	data, err := getBlockHash.EncodeCall(blockNumber)
	if err != nil {
		return nil, fmt.Errorf("multicall3: encodeGetBlockHash: %w", err)
	}
	return data, nil
}

// DecodeGetBlockHash decodes the return data of getBlockHash.
func DecodeGetBlockHash(data []byte) (*types.Hash, error) {
	h, err := getBlockHash.DecodeSingleReturnAs(data, types.NewHashFromBytes)
	if err != nil {
		return nil, fmt.Errorf("multicall3: decodeGetBlockHash: %w", err)
	}
	return h, nil
}

// EncodeGetBlockNumber returns the calldata for getBlockNumber().
func EncodeGetBlockNumber() []byte {
	return getBlockNumber.SelectorBytes()
}

// DecodeGetBlockNumber decodes the return data of getBlockNumber.
func DecodeGetBlockNumber(data []byte) (*big.Int, error) {
	n, err := getBlockNumber.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: decodeGetBlockNumber: %w", err)
	}
	return n, nil
}

// EncodeGetChainID returns the calldata for getChainId().
func EncodeGetChainID() []byte {
	return getChainId.SelectorBytes()
}

// DecodeGetChainID decodes the return data of getChainId.
func DecodeGetChainID(data []byte) (*big.Int, error) {
	n, err := getChainId.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: decodeGetChainID: %w", err)
	}
	return n, nil
}

// EncodeGetCurrentBlockCoinbase returns the calldata for getCurrentBlockCoinbase().
func EncodeGetCurrentBlockCoinbase() []byte {
	return getCurrentBlockCoinbase.SelectorBytes()
}

// DecodeGetCurrentBlockCoinbase decodes the return data of getCurrentBlockCoinbase.
func DecodeGetCurrentBlockCoinbase(data []byte) (*types.Address, error) {
	addr, err := getCurrentBlockCoinbase.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: decodeGetCurrentBlockCoinbase: %w", err)
	}
	return addr, nil
}

// EncodeGetCurrentBlockDifficulty returns the calldata for getCurrentBlockDifficulty().
func EncodeGetCurrentBlockDifficulty() []byte {
	return getCurrentBlockDiff.SelectorBytes()
}

// DecodeGetCurrentBlockDifficulty decodes the return data of getCurrentBlockDifficulty.
func DecodeGetCurrentBlockDifficulty(data []byte) (*big.Int, error) {
	n, err := getCurrentBlockDiff.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: decodeGetCurrentBlockDifficulty: %w", err)
	}
	return n, nil
}

// EncodeGetCurrentBlockGasLimit returns the calldata for getCurrentBlockGasLimit().
func EncodeGetCurrentBlockGasLimit() []byte {
	return getCurrentBlockGasLimit.SelectorBytes()
}

// DecodeGetCurrentBlockGasLimit decodes the return data of getCurrentBlockGasLimit.
func DecodeGetCurrentBlockGasLimit(data []byte) (*big.Int, error) {
	n, err := getCurrentBlockGasLimit.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: decodeGetCurrentBlockGasLimit: %w", err)
	}
	return n, nil
}

// EncodeGetCurrentBlockTimestamp returns the calldata for getCurrentBlockTimestamp().
func EncodeGetCurrentBlockTimestamp() []byte {
	return getCurrentBlockTime.SelectorBytes()
}

// DecodeGetCurrentBlockTimestamp decodes the return data of getCurrentBlockTimestamp.
func DecodeGetCurrentBlockTimestamp(data []byte) (*big.Int, error) {
	n, err := getCurrentBlockTime.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: decodeGetCurrentBlockTimestamp: %w", err)
	}
	return n, nil
}

// EncodeGetEthBalance returns the calldata for getEthBalance(addr).
func EncodeGetEthBalance(addr *types.Address) ([]byte, error) {
	data, err := getEthBalance.EncodeCall(addr)
	if err != nil {
		return nil, fmt.Errorf("multicall3: encodeGetEthBalance: %w", err)
	}
	return data, nil
}

// DecodeGetEthBalance decodes the return data of getEthBalance.
func DecodeGetEthBalance(data []byte) (*big.Int, error) {
	n, err := getEthBalance.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: decodeGetEthBalance: %w", err)
	}
	return n, nil
}

// EncodeGetLastBlockHash returns the calldata for getLastBlockHash().
func EncodeGetLastBlockHash() []byte {
	return getLastBlockHash.SelectorBytes()
}

// DecodeGetLastBlockHash decodes the return data of getLastBlockHash.
func DecodeGetLastBlockHash(data []byte) (*types.Hash, error) {
	h, err := getLastBlockHash.DecodeSingleReturnAs(data, types.NewHashFromBytes)
	if err != nil {
		return nil, fmt.Errorf("multicall3: decodeGetLastBlockHash: %w", err)
	}
	return h, nil
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
