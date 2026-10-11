package multicall3

import (
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Call is IMulticall3.Call.
type Call struct {
	Target   *types.Address
	CallData []byte
}

func NewCall(target *types.Address, callData []byte) Call {
	return Call{
		Target:   target,
		CallData: callData,
	}
}

// tuple returns c as the []any ABI encoding expects for a callT value.
func (c Call) tuple() []any {
	return []any{
		c.Target,
		c.CallData,
	}
}

// Call3 is IMulticall3.Call3.
type Call3 struct {
	Call
	AllowFailure bool
}

func NewCall3(target *types.Address, allowFailure bool, callData []byte) Call3 {
	return Call3{
		Target:       target,
		AllowFailure: allowFailure,
		CallData:     callData,
	}
}

// tuple returns c as the []any ABI encoding expects for a call3T value.
func (c Call3) tuple() []any {
	return []any{
		c.Target,
		c.AllowFailure,
		c.CallData,
	}
}

// Call3Value is IMulticall3.Call3Value.
type Call3Value struct {
	Call3
	Value *big.Int
}

func NewCall3Value(target *types.Address, allowFailure bool, value *big.Int, callData []byte) Call3Value {
	return Call3Value{
		Target:       target,
		AllowFailure: allowFailure,
		Value:        value,
		CallData:     callData,
	}
}

// tuple returns c as the []any ABI encoding expects for a call3ValueT value.
func (c Call3Value) tuple() []any {
	return []any{
		c.Target,
		c.AllowFailure,
		c.Value,
		c.CallData,
	}
}

// Result is IMulticall3.Result.
type Result struct {
	Success    bool
	ReturnData []byte
}

func NewResult(Success bool, ReturnData []byte) Result {
	return Result{
		Success:    Success,
		ReturnData: ReturnData,
	}
}

// AggregateResult is the return value of aggregate.
type AggregateResult struct {
	BlockNumber *big.Int
	ReturnData  [][]byte
}

func NewAggregateResult(blockNumber *big.Int, returnData [][]byte) AggregateResult {
	return AggregateResult{
		BlockNumber: blockNumber,
		ReturnData:  returnData,
	}
}

// BlockResults is the return value of blockAndAggregate and tryBlockAndAggregate.
type BlockResults struct {
	BlockNumber *big.Int
	BlockHash   *types.Hash
	Results     []Result
}

func NewBlockResults(blockNumber *big.Int, blockHash *types.Hash, results []Result) BlockResults {
	return BlockResults{
		BlockNumber: blockNumber,
		BlockHash:   blockHash,
		Results:     results,
	}
}
