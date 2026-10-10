package erc20

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

// ExtractTransfers pulls every Transfer-shaped log out of logs, from any contract; filter by Contract afterward.
func ExtractTransfers(logs []tx.Log) ([]Transfer, error) {
	var transfers []Transfer
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !transferEvent.Matches(topic0) {
			continue
		}

		topics, err := logTopics(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("erc20: extractTransfers: logs[%d]: %w", i, err)
		}
		vals, err := transferEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("erc20: extractTransfers: logs[%d]: %w", i, err)
		}
		from, to, value, err := eventArgs(vals)
		if err != nil {
			return nil, fmt.Errorf("erc20: extractTransfers: logs[%d]: %w", i, err)
		}
		transfers = append(transfers, *NewTransfer(log.Address, from, to, value))
	}
	return transfers, nil
}

// ExtractApprovals pulls every Approval-shaped log out of logs; see ExtractTransfers.
func ExtractApprovals(logs []tx.Log) ([]Approval, error) {
	var approvals []Approval
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !approvalEvent.Matches(topic0) {
			continue
		}

		topics, err := logTopics(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("erc20: extractApprovals: logs[%d]: %w", i, err)
		}
		vals, err := approvalEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("erc20: extractApprovals: logs[%d]: %w", i, err)
		}
		owner, spender, value, err := eventArgs(vals)
		if err != nil {
			return nil, fmt.Errorf("erc20: extractApprovals: logs[%d]: %w", i, err)
		}
		approvals = append(approvals, *NewApproval(log.Address, owner, spender, value))
	}
	return approvals, nil
}

// logTopics converts a log's topics to abi.Topic, erroring on a nil entry.
func logTopics(hashes []*types.Hash) ([]abi.Topic, error) {
	topics := make([]abi.Topic, len(hashes))
	for i, h := range hashes {
		if h == nil {
			return nil, fmt.Errorf("topics[%d] is nil", i)
		}
		copy(topics[i][:], h.Bytes())
	}
	return topics, nil
}

// eventArgs type-asserts the [address, address, uint256] shape shared by Transfer and Approval.
func eventArgs(vals []any) (*types.Address, *types.Address, *big.Int, error) {
	a, ok1 := vals[0].(*types.Address)
	b, ok2 := vals[1].(*types.Address)
	value, ok3 := vals[2].(*big.Int)
	if !ok1 || !ok2 || !ok3 {
		return nil, nil, nil, fmt.Errorf("unexpected decoded types: %T, %T, %T", vals[0], vals[1], vals[2])
	}
	return a, b, value, nil
}
