package multicall3

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// ErrCallFailed, ErrValueMismatch and ErrBasefeeNotImplemented mirror Multicall3's own revert reasons, for errors.Is.
var (
	ErrCallFailed            = errors.New("multicall3: call failed")
	ErrValueMismatch         = errors.New("multicall3: value mismatch")
	ErrBasefeeNotImplemented = errors.New("multicall3: BASEFEE opcode not implemented on this chain")
)

const (
	ErrCallFailedString    = "Multicall3: call failed"
	ErrValueMismatchString = "Multicall3: value mismatch"
)

// IMulticall3 is implemented by *Multicall3.
type IMulticall3 interface {
	Aggregate(ctx context.Context, calls []Call, block string) (*AggregateResult, error)
	TryAggregate(ctx context.Context, requireSuccess bool, calls []Call, block string) ([]Result, error)
	TryBlockAndAggregate(ctx context.Context, requireSuccess bool, calls []Call, block string) (*BlockResults, error)
	BlockAndAggregate(ctx context.Context, calls []Call, block string) (*BlockResults, error)
	Aggregate3(ctx context.Context, calls []Call3, block string) ([]Result, error)
	Aggregate3Value(ctx context.Context, calls []Call3Value, block string) ([]Result, error)
	GetBlockHash(ctx context.Context, blockNumber *big.Int, block string) (*types.Hash, error)
	GetBlockNumber(ctx context.Context, block string) (*big.Int, error)
	GetCurrentBlockCoinbase(ctx context.Context, block string) (*types.Address, error)
	GetCurrentBlockDifficulty(ctx context.Context, block string) (*big.Int, error)
	GetCurrentBlockGasLimit(ctx context.Context, block string) (*big.Int, error)
	GetCurrentBlockTimestamp(ctx context.Context, block string) (*big.Int, error)
	GetEthBalance(ctx context.Context, addr *types.Address, block string) (*big.Int, error)
	GetLastBlockHash(ctx context.Context, block string) (*types.Hash, error)
	GetBasefee(ctx context.Context, block string) (*big.Int, error)
	GetChainID(ctx context.Context, block string) (*big.Int, error)
}

var _ IMulticall3 = (*Multicall3)(nil)

// Multicall3 calls a deployed Multicall3 contract over a Client.
type Multicall3 struct {
	c rpc.Client
	a *types.Address
}

// NewMulticall3 returns a Multicall3 calling the contract at multicall via cli.
func NewMulticall3(cli rpc.Client, multicall *types.Address) IMulticall3 {
	return &Multicall3{
		c: cli,
		a: multicall,
	}
}

// Aggregate calls aggregate(calls). Any single call failing reverts the whole call.
func (m *Multicall3) Aggregate(ctx context.Context, calls []Call, block string) (*AggregateResult, error) {
	calldata, err := EncodeAggregate(calls)
	if err != nil {
		return nil, fmt.Errorf("multicall3: aggregate: %w", err)
	}

	data, err := m.callMulticall3(ctx, calldata, nil, block)
	if err != nil {
		return nil, fmt.Errorf("multicall3: aggregate: %w", err)
	}

	result, err := DecodeAggregate(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: aggregate: %w", err)
	}

	return result, nil
}

// TryAggregate calls tryAggregate(requireSuccess, calls).
func (m *Multicall3) TryAggregate(ctx context.Context, requireSuccess bool, calls []Call, block string) ([]Result, error) {
	calldata, err := EncodeTryAggregate(requireSuccess, calls)
	if err != nil {
		return nil, fmt.Errorf("multicall3: tryAggregate: %w", err)
	}

	data, err := m.callMulticall3(ctx, calldata, nil, block)
	if err != nil {
		return nil, fmt.Errorf("multicall3: tryAggregate: %w", err)
	}

	results, err := DecodeResults(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: tryAggregate: %w", err)
	}

	return results, nil
}

// TryBlockAndAggregate calls tryBlockAndAggregate(requireSuccess, calls).
func (m *Multicall3) TryBlockAndAggregate(ctx context.Context, requireSuccess bool, calls []Call, block string) (*BlockResults, error) {
	calldata, err := EncodeTryBlockAndAggregate(requireSuccess, calls)
	if err != nil {
		return nil, fmt.Errorf("multicall3: tryBlockAndAggregate: %w", err)
	}

	data, err := m.callMulticall3(ctx, calldata, nil, block)
	if err != nil {
		return nil, fmt.Errorf("multicall3: tryBlockAndAggregate: %w", err)
	}

	result, err := DecodeBlockAndAggregate(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: tryBlockAndAggregate: %w", err)
	}

	return result, nil
}

// BlockAndAggregate calls blockAndAggregate(calls): tryBlockAndAggregate with requireSuccess true.
func (m *Multicall3) BlockAndAggregate(ctx context.Context, calls []Call, block string) (*BlockResults, error) {
	calldata, err := EncodeBlockAndAggregate(calls)
	if err != nil {
		return nil, fmt.Errorf("multicall3: blockAndAggregate: %w", err)
	}

	data, err := m.callMulticall3(ctx, calldata, nil, block)
	if err != nil {
		return nil, fmt.Errorf("multicall3: blockAndAggregate: %w", err)
	}

	result, err := DecodeBlockAndAggregate(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: blockAndAggregate: %w", err)
	}

	return result, nil
}

// Aggregate3 calls aggregate3(calls).
func (m *Multicall3) Aggregate3(ctx context.Context, calls []Call3, block string) ([]Result, error) {
	calldata, err := EncodeAggregate3(calls)
	if err != nil {
		return nil, fmt.Errorf("multicall3: aggregate3: %w", err)
	}

	data, err := m.callMulticall3(ctx, calldata, nil, block)
	if err != nil {
		return nil, fmt.Errorf("multicall3: aggregate3: %w", err)
	}

	results, err := DecodeResults(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: aggregate3: %w", err)
	}

	return results, nil
}

// Aggregate3Value calls aggregate3Value(calls), sending the sum of calls' Value fields as the call's value.
func (m *Multicall3) Aggregate3Value(ctx context.Context, calls []Call3Value, block string) ([]Result, error) {
	calldata, err := EncodeAggregate3Value(calls)
	if err != nil {
		return nil, fmt.Errorf("multicall3: aggregate3Value: %w", err)
	}

	total := new(big.Int)
	for _, c := range calls {
		if c.Value != nil {
			total.Add(total, c.Value)
		}
	}

	data, err := m.callMulticall3(ctx, calldata, total, block)
	if err != nil {
		return nil, fmt.Errorf("multicall3: aggregate3Value: %w", err)
	}

	results, err := DecodeResults(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: aggregate3Value: %w", err)
	}

	return results, nil
}

// GetBlockHash calls getBlockHash(blockNumber): the BLOCKHASH opcode, zero outside the last 256 blocks.
func (m *Multicall3) GetBlockHash(ctx context.Context, blockNumber *big.Int, block string) (*types.Hash, error) {
	calldata, err := EncodeGetBlockHash(blockNumber)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getBlockHash: %w", err)
	}

	data, err := m.callMulticall3(ctx, calldata, nil, block)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getBlockHash: %w", err)
	}

	hash, err := DecodeGetBlockHash(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getBlockHash: %w", err)
	}

	return hash, nil
}

// GetBlockNumber calls getBlockNumber().
func (m *Multicall3) GetBlockNumber(ctx context.Context, block string) (*big.Int, error) {
	data, err := m.callMulticall3(ctx, EncodeGetBlockNumber(), nil, block)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getBlockNumber: %w", err)
	}

	n, err := DecodeGetBlockNumber(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getBlockNumber: %w", err)
	}

	return n, nil
}

// GetCurrentBlockCoinbase calls getCurrentBlockCoinbase().
func (m *Multicall3) GetCurrentBlockCoinbase(ctx context.Context, block string) (*types.Address, error) {
	data, err := m.callMulticall3(ctx, EncodeGetCurrentBlockCoinbase(), nil, block)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getCurrentBlockCoinbase: %w", err)
	}

	addr, err := DecodeGetCurrentBlockCoinbase(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getCurrentBlockCoinbase: %w", err)
	}

	return addr, nil
}

// GetCurrentBlockDifficulty calls getCurrentBlockDifficulty().
func (m *Multicall3) GetCurrentBlockDifficulty(ctx context.Context, block string) (*big.Int, error) {
	data, err := m.callMulticall3(ctx, EncodeGetCurrentBlockDifficulty(), nil, block)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getCurrentBlockDifficulty: %w", err)
	}

	n, err := DecodeGetCurrentBlockDifficulty(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getCurrentBlockDifficulty: %w", err)
	}

	return n, nil
}

// GetCurrentBlockGasLimit calls getCurrentBlockGasLimit().
func (m *Multicall3) GetCurrentBlockGasLimit(ctx context.Context, block string) (*big.Int, error) {
	data, err := m.callMulticall3(ctx, EncodeGetCurrentBlockGasLimit(), nil, block)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getCurrentBlockGasLimit: %w", err)
	}

	n, err := DecodeGetCurrentBlockGasLimit(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getCurrentBlockGasLimit: %w", err)
	}

	return n, nil
}

// GetCurrentBlockTimestamp calls getCurrentBlockTimestamp().
func (m *Multicall3) GetCurrentBlockTimestamp(ctx context.Context, block string) (*big.Int, error) {
	data, err := m.callMulticall3(ctx, EncodeGetCurrentBlockTimestamp(), nil, block)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getCurrentBlockTimestamp: %w", err)
	}

	n, err := DecodeGetCurrentBlockTimestamp(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getCurrentBlockTimestamp: %w", err)
	}

	return n, nil
}

// GetEthBalance calls getEthBalance(addr).
func (m *Multicall3) GetEthBalance(ctx context.Context, addr *types.Address, block string) (*big.Int, error) {
	calldata, err := EncodeGetEthBalance(addr)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getEthBalance: %w", err)
	}

	data, err := m.callMulticall3(ctx, calldata, nil, block)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getEthBalance: %w", err)
	}

	n, err := DecodeGetEthBalance(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getEthBalance: %w", err)
	}

	return n, nil
}

// GetLastBlockHash calls getLastBlockHash().
func (m *Multicall3) GetLastBlockHash(ctx context.Context, block string) (*types.Hash, error) {
	data, err := m.callMulticall3(ctx, EncodeGetLastBlockHash(), nil, block)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getLastBlockHash: %w", err)
	}

	hash, err := DecodeGetLastBlockHash(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getLastBlockHash: %w", err)
	}

	return hash, nil
}

// GetBasefee calls getBasefee(). Fails with ErrBasefeeNotImplemented if the chain has no BASEFEE opcode.
func (m *Multicall3) GetBasefee(ctx context.Context, block string) (*big.Int, error) {
	data, err := m.callMulticall3(ctx, EncodeGetBasefee(), nil, block)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getBasefee: %w: %w", ErrBasefeeNotImplemented, err)
	}

	n, err := DecodeGetBasefee(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getBasefee: %w", err)
	}

	return n, nil
}

// GetChainID calls getChainId().
func (m *Multicall3) GetChainID(ctx context.Context, block string) (*big.Int, error) {
	data, err := m.callMulticall3(ctx, EncodeGetChainID(), nil, block)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getChainId: %w", err)
	}

	n, err := DecodeGetChainID(data)
	if err != nil {
		return nil, fmt.Errorf("multicall3: getChainId: %w", err)
	}

	return n, nil
}

// callMulticall3 eth_calls the contract with calldata and an optional value.
func (m *Multicall3) callMulticall3(ctx context.Context, calldata []byte, value *big.Int, block string) ([]byte, error) {
	params := map[string]any{
		"to":   m.a.String(),
		"data": encoding.Hex.EncodePrefixed(calldata),
	}
	if value != nil && value.Sign() != 0 {
		params["value"] = "0x" + value.Text(16)
	}

	var raw string
	if err := m.c.Call(ctx, rpc.ETHCall("call", params, block, &raw)); err != nil {
		return nil, m.wrapRevertError(err)
	}

	return encoding.Hex.Decode(raw)
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
	if decErr != nil {
		return err
	}

	reason, revertErr := abi.DecodeRevert(data)
	if revertErr != nil {
		return err
	}

	switch reason {
	case ErrCallFailedString:
		return fmt.Errorf("%w: %w", ErrCallFailed, err)
	case ErrValueMismatchString:
		return fmt.Errorf("%w: %w", ErrValueMismatch, err)
	default:
		return err
	}
}
