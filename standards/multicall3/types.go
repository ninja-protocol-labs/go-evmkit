package multicall3

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
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

func NewResult(success bool, returnData []byte) Result {
	return Result{
		Success:    success,
		ReturnData: returnData,
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

// AggregateRequest is aggregate(calls)'s input on Contract.
type AggregateRequest struct {
	Contract *types.Address
	Calls    []Call
}

func NewAggregateRequest(contract *types.Address, calls []Call) *AggregateRequest {
	return &AggregateRequest{
		Contract: contract,
		Calls:    calls,
	}
}

func (r *AggregateRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if len(r.Calls) == 0 {
		return fmt.Errorf("%w: calls is empty", bind.ErrInvalidRequest)
	}

	for i, c := range r.Calls {
		if c.Target == nil {
			return fmt.Errorf("%w: calls[%d].target is nil", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

// TryAggregateRequest is tryAggregate(requireSuccess, calls)'s input on Contract.
type TryAggregateRequest struct {
	Contract       *types.Address
	RequireSuccess bool
	Calls          []Call
}

func NewTryAggregateRequest(contract *types.Address, requireSuccess bool, calls []Call) *TryAggregateRequest {
	return &TryAggregateRequest{
		Contract:       contract,
		RequireSuccess: requireSuccess,
		Calls:          calls,
	}
}

func (r *TryAggregateRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if len(r.Calls) == 0 {
		return fmt.Errorf("%w: calls is empty", bind.ErrInvalidRequest)
	}

	for i, c := range r.Calls {
		if c.Target == nil {
			return fmt.Errorf("%w: calls[%d].target is nil", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

// BlockAndAggregateRequest is blockAndAggregate(calls)'s input on Contract.
type BlockAndAggregateRequest struct {
	Contract *types.Address
	Calls    []Call
}

func NewBlockAndAggregateRequest(contract *types.Address, calls []Call) *BlockAndAggregateRequest {
	return &BlockAndAggregateRequest{
		Contract: contract,
		Calls:    calls,
	}
}

func (r *BlockAndAggregateRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if len(r.Calls) == 0 {
		return fmt.Errorf("%w: calls is empty", bind.ErrInvalidRequest)
	}

	for i, c := range r.Calls {
		if c.Target == nil {
			return fmt.Errorf("%w: calls[%d].target is nil", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

// TryBlockAndAggregateRequest is tryBlockAndAggregate(requireSuccess, calls)'s input on Contract.
type TryBlockAndAggregateRequest struct {
	Contract       *types.Address
	RequireSuccess bool
	Calls          []Call
}

func NewTryBlockAndAggregateRequest(contract *types.Address, requireSuccess bool, calls []Call) *TryBlockAndAggregateRequest {
	return &TryBlockAndAggregateRequest{
		Contract:       contract,
		RequireSuccess: requireSuccess,
		Calls:          calls,
	}
}

func (r *TryBlockAndAggregateRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if len(r.Calls) == 0 {
		return fmt.Errorf("%w: calls is empty", bind.ErrInvalidRequest)
	}

	for i, c := range r.Calls {
		if c.Target == nil {
			return fmt.Errorf("%w: calls[%d].target is nil", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

// Aggregate3Request is aggregate3(calls)'s input on Contract.
type Aggregate3Request struct {
	Contract *types.Address
	Calls    []Call3
}

func NewAggregate3Request(contract *types.Address, calls []Call3) *Aggregate3Request {
	return &Aggregate3Request{
		Contract: contract,
		Calls:    calls,
	}
}

func (r *Aggregate3Request) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if len(r.Calls) == 0 {
		return fmt.Errorf("%w: calls is empty", bind.ErrInvalidRequest)
	}

	for i, c := range r.Calls {
		if c.Target == nil {
			return fmt.Errorf("%w: calls[%d].target is nil", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

// Aggregate3ValueRequest is aggregate3Value(calls)'s input on Contract; msg.value is TotalValue.
type Aggregate3ValueRequest struct {
	Contract *types.Address
	Calls    []Call3Value
}

func NewAggregate3ValueRequest(contract *types.Address, calls []Call3Value) *Aggregate3ValueRequest {
	return &Aggregate3ValueRequest{
		Contract: contract,
		Calls:    calls,
	}
}

func (r *Aggregate3ValueRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if len(r.Calls) == 0 {
		return fmt.Errorf("%w: calls is empty", bind.ErrInvalidRequest)
	}

	for i, c := range r.Calls {
		if c.Target == nil {
			return fmt.Errorf("%w: calls[%d].target is nil", bind.ErrInvalidRequest, i)
		}
		if c.Value == nil {
			return fmt.Errorf("%w: calls[%d].value is nil", bind.ErrInvalidRequest, i)
		}
		if c.Value.Sign() < 0 {
			return fmt.Errorf("%w: calls[%d].value is negative", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

// TotalValue returns the sum of the calls' values, the msg.value aggregate3Value requires.
func (r *Aggregate3ValueRequest) TotalValue() *big.Int {
	total := new(big.Int)
	for _, c := range r.Calls {
		if c.Value != nil {
			total.Add(total, c.Value)
		}
	}
	return total
}

// GetBasefeeRequest is getBasefee()'s input on Contract.
type GetBasefeeRequest struct {
	Contract *types.Address
}

func NewGetBasefeeRequest(contract *types.Address) *GetBasefeeRequest {
	return &GetBasefeeRequest{
		Contract: contract,
	}
}

func (r *GetBasefeeRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetBlockHashRequest is getBlockHash(blockNumber)'s input on Contract.
type GetBlockHashRequest struct {
	Contract    *types.Address
	BlockNumber *big.Int
}

func NewGetBlockHashRequest(contract *types.Address, blockNumber *big.Int) *GetBlockHashRequest {
	return &GetBlockHashRequest{
		Contract:    contract,
		BlockNumber: blockNumber,
	}
}

func (r *GetBlockHashRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.BlockNumber == nil {
		return fmt.Errorf("%w: block number is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetBlockNumberRequest is getBlockNumber()'s input on Contract.
type GetBlockNumberRequest struct {
	Contract *types.Address
}

func NewGetBlockNumberRequest(contract *types.Address) *GetBlockNumberRequest {
	return &GetBlockNumberRequest{
		Contract: contract,
	}
}

func (r *GetBlockNumberRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetChainIDRequest is getChainId()'s input on Contract.
type GetChainIDRequest struct {
	Contract *types.Address
}

func NewGetChainIDRequest(contract *types.Address) *GetChainIDRequest {
	return &GetChainIDRequest{
		Contract: contract,
	}
}

func (r *GetChainIDRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetCurrentBlockCoinbaseRequest is getCurrentBlockCoinbase()'s input on Contract.
type GetCurrentBlockCoinbaseRequest struct {
	Contract *types.Address
}

func NewGetCurrentBlockCoinbaseRequest(contract *types.Address) *GetCurrentBlockCoinbaseRequest {
	return &GetCurrentBlockCoinbaseRequest{
		Contract: contract,
	}
}

func (r *GetCurrentBlockCoinbaseRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetCurrentBlockDifficultyRequest is getCurrentBlockDifficulty()'s input on Contract.
type GetCurrentBlockDifficultyRequest struct {
	Contract *types.Address
}

func NewGetCurrentBlockDifficultyRequest(contract *types.Address) *GetCurrentBlockDifficultyRequest {
	return &GetCurrentBlockDifficultyRequest{
		Contract: contract,
	}
}

func (r *GetCurrentBlockDifficultyRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetCurrentBlockGasLimitRequest is getCurrentBlockGasLimit()'s input on Contract.
type GetCurrentBlockGasLimitRequest struct {
	Contract *types.Address
}

func NewGetCurrentBlockGasLimitRequest(contract *types.Address) *GetCurrentBlockGasLimitRequest {
	return &GetCurrentBlockGasLimitRequest{
		Contract: contract,
	}
}

func (r *GetCurrentBlockGasLimitRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetCurrentBlockTimestampRequest is getCurrentBlockTimestamp()'s input on Contract.
type GetCurrentBlockTimestampRequest struct {
	Contract *types.Address
}

func NewGetCurrentBlockTimestampRequest(contract *types.Address) *GetCurrentBlockTimestampRequest {
	return &GetCurrentBlockTimestampRequest{
		Contract: contract,
	}
}

func (r *GetCurrentBlockTimestampRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetEthBalanceRequest is getEthBalance(addr)'s input on Contract.
type GetEthBalanceRequest struct {
	Contract *types.Address
	Addr     *types.Address
}

func NewGetEthBalanceRequest(contract, addr *types.Address) *GetEthBalanceRequest {
	return &GetEthBalanceRequest{
		Contract: contract,
		Addr:     addr,
	}
}

func (r *GetEthBalanceRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Addr == nil {
		return fmt.Errorf("%w: addr is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetLastBlockHashRequest is getLastBlockHash()'s input on Contract.
type GetLastBlockHashRequest struct {
	Contract *types.Address
}

func NewGetLastBlockHashRequest(contract *types.Address) *GetLastBlockHashRequest {
	return &GetLastBlockHashRequest{
		Contract: contract,
	}
}

func (r *GetLastBlockHashRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}
