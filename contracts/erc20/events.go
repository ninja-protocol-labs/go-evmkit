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

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("erc20: extractTransfers: logs[%d]: %w", i, err)
		}
		vals, err := transferEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("erc20: extractTransfers: logs[%d]: %w", i, err)
		}
		from, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("erc20: extractTransfers: logs[%d]: unexpected decoded type for from: %T", i, vals[0])
		}
		to, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("erc20: extractTransfers: logs[%d]: unexpected decoded type for to: %T", i, vals[1])
		}
		value, ok := vals[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("erc20: extractTransfers: logs[%d]: unexpected decoded type for value: %T", i, vals[2])
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

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("erc20: extractApprovals: logs[%d]: %w", i, err)
		}
		vals, err := approvalEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("erc20: extractApprovals: logs[%d]: %w", i, err)
		}
		owner, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("erc20: extractApprovals: logs[%d]: unexpected decoded type for owner: %T", i, vals[0])
		}
		spender, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("erc20: extractApprovals: logs[%d]: unexpected decoded type for spender: %T", i, vals[1])
		}
		value, ok := vals[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("erc20: extractApprovals: logs[%d]: unexpected decoded type for value: %T", i, vals[2])
		}
		approvals = append(approvals, *NewApproval(log.Address, owner, spender, value))
	}
	return approvals, nil
}
