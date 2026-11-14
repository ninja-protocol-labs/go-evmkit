package erc6909

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

// ExtractOperatorSets pulls every OperatorSet-shaped log out of logs, from any contract; filter by Contract afterward.
func ExtractOperatorSets(logs []tx.Log) ([]OperatorSet, error) {
	var operatorSets []OperatorSet
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !operatorSetEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("erc6909: extractOperatorSets: logs[%d]: %w", i, err)
		}
		vals, err := operatorSetEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("erc6909: extractOperatorSets: logs[%d]: %w", i, err)
		}
		owner, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("erc6909: extractOperatorSets: logs[%d]: unexpected decoded type for owner: %T", i, vals[0])
		}
		operator, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("erc6909: extractOperatorSets: logs[%d]: unexpected decoded type for operator: %T", i, vals[1])
		}
		approved, ok := vals[2].(bool)
		if !ok {
			return nil, fmt.Errorf("erc6909: extractOperatorSets: logs[%d]: unexpected decoded type for approved: %T", i, vals[2])
		}
		operatorSets = append(operatorSets, *NewOperatorSet(log.Address, owner, operator, approved))
	}
	return operatorSets, nil
}

// ExtractApprovals pulls every Approval-shaped log out of logs; see ExtractOperatorSets.
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
			return nil, fmt.Errorf("erc6909: extractApprovals: logs[%d]: %w", i, err)
		}
		vals, err := approvalEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("erc6909: extractApprovals: logs[%d]: %w", i, err)
		}
		owner, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("erc6909: extractApprovals: logs[%d]: unexpected decoded type for owner: %T", i, vals[0])
		}
		spender, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("erc6909: extractApprovals: logs[%d]: unexpected decoded type for spender: %T", i, vals[1])
		}
		id, ok := vals[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("erc6909: extractApprovals: logs[%d]: unexpected decoded type for id: %T", i, vals[2])
		}
		amount, ok := vals[3].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("erc6909: extractApprovals: logs[%d]: unexpected decoded type for amount: %T", i, vals[3])
		}
		approvals = append(approvals, *NewApproval(log.Address, owner, spender, id, amount))
	}
	return approvals, nil
}

// ExtractTransfers pulls every Transfer-shaped log out of logs; see ExtractOperatorSets.
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
			return nil, fmt.Errorf("erc6909: extractTransfers: logs[%d]: %w", i, err)
		}
		vals, err := transferEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("erc6909: extractTransfers: logs[%d]: %w", i, err)
		}
		caller, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("erc6909: extractTransfers: logs[%d]: unexpected decoded type for caller: %T", i, vals[0])
		}
		sender, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("erc6909: extractTransfers: logs[%d]: unexpected decoded type for sender: %T", i, vals[1])
		}
		receiver, ok := vals[2].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("erc6909: extractTransfers: logs[%d]: unexpected decoded type for receiver: %T", i, vals[2])
		}
		id, ok := vals[3].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("erc6909: extractTransfers: logs[%d]: unexpected decoded type for id: %T", i, vals[3])
		}
		amount, ok := vals[4].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("erc6909: extractTransfers: logs[%d]: unexpected decoded type for amount: %T", i, vals[4])
		}
		transfers = append(transfers, *NewTransfer(log.Address, caller, sender, receiver, id, amount))
	}
	return transfers, nil
}

// ExtractNameUpdated pulls every ERC6909NameUpdated-shaped log out of logs; see ExtractOperatorSets.
func ExtractNameUpdated(logs []tx.Log) ([]NameUpdated, error) {
	var updates []NameUpdated
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !nameUpdatedEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("erc6909: extractNameUpdated: logs[%d]: %w", i, err)
		}
		vals, err := nameUpdatedEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("erc6909: extractNameUpdated: logs[%d]: %w", i, err)
		}
		id, ok := vals[0].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("erc6909: extractNameUpdated: logs[%d]: unexpected decoded type for id: %T", i, vals[0])
		}
		newName, ok := vals[1].(string)
		if !ok {
			return nil, fmt.Errorf("erc6909: extractNameUpdated: logs[%d]: unexpected decoded type for newName: %T", i, vals[1])
		}
		updates = append(updates, *NewNameUpdated(log.Address, id, newName))
	}
	return updates, nil
}

// ExtractSymbolUpdated pulls every ERC6909SymbolUpdated-shaped log out of logs; see ExtractOperatorSets.
func ExtractSymbolUpdated(logs []tx.Log) ([]SymbolUpdated, error) {
	var updates []SymbolUpdated
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !symbolUpdatedEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("erc6909: extractSymbolUpdated: logs[%d]: %w", i, err)
		}
		vals, err := symbolUpdatedEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("erc6909: extractSymbolUpdated: logs[%d]: %w", i, err)
		}
		id, ok := vals[0].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("erc6909: extractSymbolUpdated: logs[%d]: unexpected decoded type for id: %T", i, vals[0])
		}
		newSymbol, ok := vals[1].(string)
		if !ok {
			return nil, fmt.Errorf("erc6909: extractSymbolUpdated: logs[%d]: unexpected decoded type for newSymbol: %T", i, vals[1])
		}
		updates = append(updates, *NewSymbolUpdated(log.Address, id, newSymbol))
	}
	return updates, nil
}

// ExtractDecimalsUpdated pulls every ERC6909DecimalsUpdated-shaped log out of logs; see ExtractOperatorSets.
func ExtractDecimalsUpdated(logs []tx.Log) ([]DecimalsUpdated, error) {
	var updates []DecimalsUpdated
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !decimalsUpdatedEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("erc6909: extractDecimalsUpdated: logs[%d]: %w", i, err)
		}
		vals, err := decimalsUpdatedEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("erc6909: extractDecimalsUpdated: logs[%d]: %w", i, err)
		}
		id, ok := vals[0].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("erc6909: extractDecimalsUpdated: logs[%d]: unexpected decoded type for id: %T", i, vals[0])
		}
		newDecimals, ok := vals[1].(uint8)
		if !ok {
			return nil, fmt.Errorf("erc6909: extractDecimalsUpdated: logs[%d]: unexpected decoded type for newDecimals: %T", i, vals[1])
		}
		updates = append(updates, *NewDecimalsUpdated(log.Address, id, newDecimals))
	}
	return updates, nil
}
