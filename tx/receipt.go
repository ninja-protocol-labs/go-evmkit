package tx

import (
	"context"
	"fmt"
	"time"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// defaultReceiptPollInterval is used by WaitForReceipt when pollInterval is <= 0.
const defaultReceiptPollInterval = 2 * time.Second

// Receipt holds the fields common to every EVM chain's transaction
// receipt. Chain-specific fields (an L2's L1 fee, blob gas fields, ...)
// are not modeled here; read them from Raw instead.
type Receipt struct {
	TxHash          *types.Hash
	BlockHash       *types.Hash
	BlockNumber     uint64
	Status          bool // true for 0x1 (success)
	GasUsed         uint64
	ContractAddress *types.Address // nil unless this was a contract creation
	Logs            []Log

	// Raw is the full decoded eth_getTransactionReceipt result.
	Raw map[string]any
}

// Log holds the fields common to every EVM chain's event log.
type Log struct {
	Address     *types.Address
	Topics      []*types.Hash
	Data        []byte
	BlockNumber uint64
	TxHash      *types.Hash
	LogIndex    uint64
	Removed     bool

	// Raw is the full decoded log object.
	Raw map[string]any
}

// WaitForReceipt polls eth_getTransactionReceipt for hash every
// pollInterval (or defaultReceiptPollInterval if pollInterval <= 0) until a
// receipt appears, ctx is done, or timeout elapses. A timeout of 0 applies
// none, leaving ctx's own deadline (if any) as the only bound.
func WaitForReceipt(ctx context.Context, client rpc.Client, hash *types.Hash, pollInterval, timeout time.Duration) (*Receipt, error) {
	if pollInterval <= 0 {
		pollInterval = defaultReceiptPollInterval
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	for {
		var raw map[string]any
		if err := client.Call(ctx, rpc.ETHGetTransactionReceipt("receipt", hash.String(), &raw)); err != nil {
			return nil, fmt.Errorf("tx: wait for receipt: %w", err)
		}
		if raw != nil {
			return ParseReceipt(raw)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(pollInterval):
		}
	}
}

// ParseReceipt decodes an eth_getTransactionReceipt result into a Receipt.
func ParseReceipt(raw map[string]any) (*Receipt, error) {
	txHashHex, ok := raw["transactionHash"].(string)
	if !ok {
		return nil, fmt.Errorf("tx: receipt: missing transactionHash")
	}
	txHash, err := types.NewHashFromHex(txHashHex)
	if err != nil {
		return nil, fmt.Errorf("tx: receipt: transactionHash: %w", err)
	}

	blockHashHex, ok := raw["blockHash"].(string)
	if !ok {
		return nil, fmt.Errorf("tx: receipt: missing blockHash")
	}
	blockHash, err := types.NewHashFromHex(blockHashHex)
	if err != nil {
		return nil, fmt.Errorf("tx: receipt: blockHash: %w", err)
	}

	blockNumberHex, ok := raw["blockNumber"].(string)
	if !ok {
		return nil, fmt.Errorf("tx: receipt: missing blockNumber")
	}
	blockNumber, err := parseQuantity(blockNumberHex)
	if err != nil {
		return nil, fmt.Errorf("tx: receipt: blockNumber: %w", err)
	}

	statusHex, ok := raw["status"].(string)
	if !ok {
		return nil, fmt.Errorf("tx: receipt: missing status")
	}
	status, err := parseQuantity(statusHex)
	if err != nil {
		return nil, fmt.Errorf("tx: receipt: status: %w", err)
	}

	gasUsedHex, ok := raw["gasUsed"].(string)
	if !ok {
		return nil, fmt.Errorf("tx: receipt: missing gasUsed")
	}
	gasUsed, err := parseQuantity(gasUsedHex)
	if err != nil {
		return nil, fmt.Errorf("tx: receipt: gasUsed: %w", err)
	}

	var contractAddress *types.Address
	if ca, ok := raw["contractAddress"].(string); ok && ca != "" {
		contractAddress, err = types.NewAddressFromHex(ca)
		if err != nil {
			return nil, fmt.Errorf("tx: receipt: contractAddress: %w", err)
		}
	}

	var logs []Log
	if raw["logs"] != nil {
		rawLogs, ok := raw["logs"].([]any)
		if !ok {
			return nil, fmt.Errorf("tx: receipt: logs: unexpected type %T", raw["logs"])
		}
		logs = make([]Log, len(rawLogs))
		for i, l := range rawLogs {
			rawLog, ok := l.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("tx: receipt: logs[%d]: unexpected type %T", i, l)
			}
			log, err := ParseLog(rawLog)
			if err != nil {
				return nil, fmt.Errorf("tx: receipt: logs[%d]: %w", i, err)
			}
			logs[i] = log
		}
	}

	return &Receipt{
		TxHash:          txHash,
		BlockHash:       blockHash,
		BlockNumber:     blockNumber.Uint64(),
		Status:          status.Sign() != 0,
		GasUsed:         gasUsed.Uint64(),
		ContractAddress: contractAddress,
		Logs:            logs,
		Raw:             raw,
	}, nil
}

// ParseLog decodes one entry of a receipt's "logs" array into a Log.
func ParseLog(raw map[string]any) (Log, error) {
	addressHex, ok := raw["address"].(string)
	if !ok {
		return Log{}, fmt.Errorf("missing address")
	}
	address, err := types.NewAddressFromHex(addressHex)
	if err != nil {
		return Log{}, fmt.Errorf("address: %w", err)
	}

	rawTopics, ok := raw["topics"].([]any)
	if !ok {
		return Log{}, fmt.Errorf("missing topics")
	}
	topics := make([]*types.Hash, len(rawTopics))
	for i, t := range rawTopics {
		topicHex, ok := t.(string)
		if !ok {
			return Log{}, fmt.Errorf("topics[%d]: unexpected type %T", i, t)
		}
		topics[i], err = types.NewHashFromHex(topicHex)
		if err != nil {
			return Log{}, fmt.Errorf("topics[%d]: %w", i, err)
		}
	}

	dataHex, ok := raw["data"].(string)
	if !ok {
		return Log{}, fmt.Errorf("missing data")
	}
	data, err := encoding.Hex.Decode(dataHex)
	if err != nil {
		return Log{}, fmt.Errorf("data: %w", err)
	}

	blockNumberHex, ok := raw["blockNumber"].(string)
	if !ok {
		return Log{}, fmt.Errorf("missing blockNumber")
	}
	blockNumber, err := parseQuantity(blockNumberHex)
	if err != nil {
		return Log{}, fmt.Errorf("blockNumber: %w", err)
	}

	txHashHex, ok := raw["transactionHash"].(string)
	if !ok {
		return Log{}, fmt.Errorf("missing transactionHash")
	}
	txHash, err := types.NewHashFromHex(txHashHex)
	if err != nil {
		return Log{}, fmt.Errorf("transactionHash: %w", err)
	}

	logIndexHex, ok := raw["logIndex"].(string)
	if !ok {
		return Log{}, fmt.Errorf("missing logIndex")
	}
	logIndex, err := parseQuantity(logIndexHex)
	if err != nil {
		return Log{}, fmt.Errorf("logIndex: %w", err)
	}

	removed, _ := raw["removed"].(bool)

	return Log{
		Address:     address,
		Topics:      topics,
		Data:        data,
		BlockNumber: blockNumber.Uint64(),
		TxHash:      txHash,
		LogIndex:    logIndex.Uint64(),
		Removed:     removed,
		Raw:         raw,
	}, nil
}
