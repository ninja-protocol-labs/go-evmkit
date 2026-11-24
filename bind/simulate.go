package bind

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// SimulateOptions configures Simulate; a nil *SimulateOptions means latest block and no flags.
type SimulateOptions struct {
	// Block is the parent block tag or hex number; empty means rpc.BlockTagLatest.
	Block string
	// StateOverride is applied before the first call; nil omits it.
	StateOverride *rpc.StateOverride
	// TraceTransfers adds a synthetic Transfer log for every native value transfer.
	TraceTransfers bool
	// Validation enforces nonce, balance and fee checks as for a real transaction.
	Validation bool
}

// NewSimulateOptions returns empty SimulateOptions: latest block, no override, no flags.
func NewSimulateOptions() *SimulateOptions {
	return new(SimulateOptions)
}

// WithBlock sets Block in place and returns o.
func (o *SimulateOptions) WithBlock(tag string) *SimulateOptions {
	o.Block = tag
	return o
}

// WithStateOverride sets StateOverride in place and returns o.
func (o *SimulateOptions) WithStateOverride(override *rpc.StateOverride) *SimulateOptions {
	o.StateOverride = override
	return o
}

// WithTraceTransfers enables TraceTransfers in place and returns o.
func (o *SimulateOptions) WithTraceTransfers() *SimulateOptions {
	o.TraceTransfers = true
	return o
}

// WithValidation enables Validation in place and returns o.
func (o *SimulateOptions) WithValidation() *SimulateOptions {
	o.Validation = true
	return o
}

// SimulateCall is one call of a simulation; calls run in order, each seeing the state left by the previous ones.
type SimulateCall struct {
	From *types.Address
	// To is the callee; nil creates a contract from Data.
	To *types.Address
	// Value is msg.value in wei; nil omits it.
	Value *big.Int
	Data  []byte
}

// NewSimulateCall returns a SimulateCall from from to to with value and data.
func NewSimulateCall(from, to *types.Address, value *big.Int, data []byte) SimulateCall {
	return SimulateCall{
		From:  from,
		To:    to,
		Value: value,
		Data:  data,
	}
}

// Validate reports whether c can be sent: From is required and Value must fit uint256.
func (c *SimulateCall) Validate() error {
	if c == nil {
		return fmt.Errorf("%w: call is nil", ErrInvalidRequest)
	}

	if c.From == nil {
		return fmt.Errorf("%w: from is nil", ErrInvalidRequest)
	}

	if c.Value != nil && c.Value.Sign() < 0 {
		return fmt.Errorf("%w: value is negative", ErrInvalidRequest)
	}

	if c.Value != nil && c.Value.BitLen() > 256 {
		return fmt.Errorf("%w: value exceeds uint256", ErrInvalidRequest)
	}

	return nil
}

// SimulateResult is the outcome of one SimulateCall.
type SimulateResult struct {
	Success    bool
	GasUsed    uint64
	ReturnData []byte
	// Logs are decoded like receipt logs, so event extractors accept them directly.
	Logs []tx.Log
	// Error is the call's revert as an *rpc.ResponseError, ready for the protocol MapRevertError functions; nil on success.
	Error error
}

// SimulateResults decodes an eth_simulateV1 response into one SimulateResult per call; SimulateElement sets how many calls to expect.
type SimulateResults struct {
	Calls []SimulateResult

	expected int
}

// UnmarshalJSON decodes the single simulated block, failing if its call count differs from the one SimulateElement recorded.
func (r *SimulateResults) UnmarshalJSON(data []byte) error {
	var blocks []map[string]any
	if err := json.Unmarshal(data, &blocks); err != nil {
		return fmt.Errorf("bind: simulate: %w", err)
	}

	if len(blocks) != 1 {
		return fmt.Errorf("bind: simulate: expected 1 block, got %d", len(blocks))
	}
	rawCalls, ok := blocks[0]["calls"].([]any)
	if !ok {
		return fmt.Errorf("bind: simulate: block has no calls array")
	}
	if r.expected > 0 && len(rawCalls) != r.expected {
		return fmt.Errorf("bind: simulate: expected %d call results, got %d", r.expected, len(rawCalls))
	}

	results := make([]SimulateResult, len(rawCalls))
	for i, rawCall := range rawCalls {
		raw, ok := rawCall.(map[string]any)
		if !ok {
			return fmt.Errorf("bind: simulate: calls[%d]: unexpected type %T", i, rawCall)
		}

		status, ok := raw["status"].(string)
		if !ok || (status != "0x0" && status != "0x1") {
			return fmt.Errorf("bind: simulate: calls[%d]: invalid status %v", i, raw["status"])
		}

		gasUsedHex, ok := raw["gasUsed"].(string)
		if !ok {
			return fmt.Errorf("bind: simulate: calls[%d]: missing gasUsed", i)
		}
		gasUsed, err := strconv.ParseUint(strings.TrimPrefix(gasUsedHex, "0x"), 16, 64)
		if err != nil {
			return fmt.Errorf("bind: simulate: calls[%d]: gasUsed: %w", i, err)
		}

		returnDataHex, ok := raw["returnData"].(string)
		if !ok {
			return fmt.Errorf("bind: simulate: calls[%d]: missing returnData", i)
		}
		returnData, err := encoding.Hex.Decode(returnDataHex)
		if err != nil {
			return fmt.Errorf("bind: simulate: calls[%d]: returnData: %w", i, err)
		}

		var logs []tx.Log
		if rawLogs, ok := raw["logs"].([]any); ok {
			logs = make([]tx.Log, len(rawLogs))
			for j, rawLog := range rawLogs {
				logMap, ok := rawLog.(map[string]any)
				if !ok {
					return fmt.Errorf("bind: simulate: calls[%d]: logs[%d]: unexpected type %T", i, j, rawLog)
				}
				logs[j], err = tx.ParseLog(logMap)
				if err != nil {
					return fmt.Errorf("bind: simulate: calls[%d]: logs[%d]: %w", i, j, err)
				}
			}
		}

		var callErr *rpc.ResponseError
		if rawErr, ok := raw["error"].(map[string]any); ok {
			code, ok := rawErr["code"].(float64)
			if !ok {
				return fmt.Errorf("bind: simulate: calls[%d]: error: missing code", i)
			}
			message, _ := rawErr["message"].(string)
			callErr = &rpc.ResponseError{Code: int(code), Message: message}
			if data, ok := rawErr["data"]; ok {
				encoded, err := json.Marshal(data)
				if err != nil {
					return fmt.Errorf("bind: simulate: calls[%d]: error data: %w", i, err)
				}
				callErr.Data = encoded
			}
		}

		results[i] = SimulateResult{
			Success:    status == "0x1",
			GasUsed:    gasUsed,
			ReturnData: returnData,
			Logs:       logs,
		}
		if callErr != nil {
			results[i].Error = callErr
		}
	}

	r.Calls = results
	return nil
}

// SimulateElement validates opts and calls and builds the eth_simulateV1 element decoding into result, for sending in an rpc batch.
func SimulateElement(id string, opts *SimulateOptions, calls []SimulateCall, result *SimulateResults) (rpc.Element, error) {
	if result == nil {
		return rpc.Element{}, fmt.Errorf("bind: simulate: %w: result is nil", ErrInvalidRequest)
	}

	if opts == nil {
		opts = NewSimulateOptions()
	}
	if err := opts.StateOverride.Validate(); err != nil {
		return rpc.Element{}, fmt.Errorf("bind: simulate: %w: %w", ErrInvalidOptions, err)
	}

	if len(calls) == 0 {
		return rpc.Element{}, fmt.Errorf("bind: simulate: %w: calls is empty", ErrInvalidRequest)
	}

	params := make([]map[string]any, len(calls))
	for i, call := range calls {
		if err := call.Validate(); err != nil {
			return rpc.Element{}, fmt.Errorf("bind: simulate: calls[%d]: %w", i, err)
		}

		param := map[string]any{
			"from": call.From.String(),
			"data": encoding.Hex.EncodePrefixed(call.Data),
		}
		if call.To != nil {
			param["to"] = call.To.String()
		}
		if call.Value != nil {
			param["value"] = "0x" + call.Value.Text(16)
		}
		params[i] = param
	}

	blockStateCall := map[string]any{"calls": params}
	if opts.StateOverride != nil {
		blockStateCall["stateOverrides"] = opts.StateOverride
	}
	payload := map[string]any{
		"blockStateCalls": []any{blockStateCall},
		"traceTransfers":  opts.TraceTransfers,
		"validation":      opts.Validation,
	}

	block := opts.Block
	if block == "" {
		block = rpc.BlockTagLatest
	}

	result.Calls = nil
	result.expected = len(calls)
	return rpc.NewElement(id, rpc.MethodETHSimulateV1, []any{payload, block}, result), nil
}

// Simulate runs calls in order in one eth_simulateV1 block, returning one result per call; RPC errors are returned unwrapped.
func Simulate(ctx context.Context, c rpc.Client, opts *SimulateOptions, calls []SimulateCall) ([]SimulateResult, error) {
	var results SimulateResults
	e, err := SimulateElement("simulate", opts, calls, &results)
	if err != nil {
		return nil, err
	}

	if err = c.Call(ctx, e); err != nil {
		return nil, err
	}

	if len(results.Calls) != len(calls) {
		return nil, fmt.Errorf("bind: simulate: expected %d call results, got %d", len(calls), len(results.Calls))
	}
	return results.Calls, nil
}
