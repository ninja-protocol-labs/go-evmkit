package erc1155

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

// ExtractTransfers pulls every TransferSingle-shaped log out of logs, from any contract; filter by Contract afterward.
func ExtractTransfers(logs []tx.Log) ([]Transfer, error) {
	var transfers []Transfer
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !transferSingleEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("erc1155: extractTransfers: logs[%d]: %w", i, err)
		}
		vals, err := transferSingleEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("erc1155: extractTransfers: logs[%d]: %w", i, err)
		}
		operator, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("erc1155: extractTransfers: logs[%d]: unexpected decoded type for operator: %T", i, vals[0])
		}
		from, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("erc1155: extractTransfers: logs[%d]: unexpected decoded type for from: %T", i, vals[1])
		}
		to, ok := vals[2].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("erc1155: extractTransfers: logs[%d]: unexpected decoded type for to: %T", i, vals[2])
		}
		id, ok := vals[3].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("erc1155: extractTransfers: logs[%d]: unexpected decoded type for id: %T", i, vals[3])
		}
		value, ok := vals[4].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("erc1155: extractTransfers: logs[%d]: unexpected decoded type for value: %T", i, vals[4])
		}
		transfers = append(transfers, *NewTransfer(log.Address, operator, from, to, id, value))
	}
	return transfers, nil
}

// ExtractTransferBatches pulls every TransferBatch-shaped log out of logs; see ExtractTransfers.
func ExtractTransferBatches(logs []tx.Log) ([]TransferBatch, error) {
	var batches []TransferBatch
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !transferBatchEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("erc1155: extractTransferBatches: logs[%d]: %w", i, err)
		}
		vals, err := transferBatchEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("erc1155: extractTransferBatches: logs[%d]: %w", i, err)
		}
		operator, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("erc1155: extractTransferBatches: logs[%d]: unexpected decoded type for operator: %T", i, vals[0])
		}
		from, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("erc1155: extractTransferBatches: logs[%d]: unexpected decoded type for from: %T", i, vals[1])
		}
		to, ok := vals[2].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("erc1155: extractTransferBatches: logs[%d]: unexpected decoded type for to: %T", i, vals[2])
		}
		idElems, ok := vals[3].([]any)
		if !ok {
			return nil, fmt.Errorf("erc1155: extractTransferBatches: logs[%d]: unexpected decoded type for ids: %T", i, vals[3])
		}
		ids := make([]*big.Int, len(idElems))
		for j, e := range idElems {
			id, ok := e.(*big.Int)
			if !ok {
				return nil, fmt.Errorf("erc1155: extractTransferBatches: logs[%d]: ids[%d]: unexpected decoded type: %T", i, j, e)
			}
			ids[j] = id
		}
		valueElems, ok := vals[4].([]any)
		if !ok {
			return nil, fmt.Errorf("erc1155: extractTransferBatches: logs[%d]: unexpected decoded type for values: %T", i, vals[4])
		}
		values := make([]*big.Int, len(valueElems))
		for j, e := range valueElems {
			value, ok := e.(*big.Int)
			if !ok {
				return nil, fmt.Errorf("erc1155: extractTransferBatches: logs[%d]: values[%d]: unexpected decoded type: %T", i, j, e)
			}
			values[j] = value
		}
		batches = append(batches, *NewTransferBatch(log.Address, operator, from, to, ids, values))
	}
	return batches, nil
}

// ExtractApprovalForAll pulls every ApprovalForAll-shaped log out of logs; see ExtractTransfers.
func ExtractApprovalForAll(logs []tx.Log) ([]ApprovalForAll, error) {
	var approvals []ApprovalForAll
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !approvalForAllEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("erc1155: extractApprovalForAll: logs[%d]: %w", i, err)
		}
		vals, err := approvalForAllEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("erc1155: extractApprovalForAll: logs[%d]: %w", i, err)
		}
		account, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("erc1155: extractApprovalForAll: logs[%d]: unexpected decoded type for account: %T", i, vals[0])
		}
		operator, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("erc1155: extractApprovalForAll: logs[%d]: unexpected decoded type for operator: %T", i, vals[1])
		}
		approved, ok := vals[2].(bool)
		if !ok {
			return nil, fmt.Errorf("erc1155: extractApprovalForAll: logs[%d]: unexpected decoded type for approved: %T", i, vals[2])
		}
		approvals = append(approvals, *NewApprovalForAll(log.Address, account, operator, approved))
	}
	return approvals, nil
}

// ExtractURI pulls every URI-shaped log out of logs; see ExtractTransfers. Implementations aren't
// required to emit this event, so its absence doesn't mean the uri() function's result is stale.
func ExtractURI(logs []tx.Log) ([]URI, error) {
	var uris []URI
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !uriEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("erc1155: extractURI: logs[%d]: %w", i, err)
		}
		vals, err := uriEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("erc1155: extractURI: logs[%d]: %w", i, err)
		}
		value, ok := vals[0].(string)
		if !ok {
			return nil, fmt.Errorf("erc1155: extractURI: logs[%d]: unexpected decoded type for value: %T", i, vals[0])
		}
		id, ok := vals[1].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("erc1155: extractURI: logs[%d]: unexpected decoded type for id: %T", i, vals[1])
		}
		uris = append(uris, *NewURI(log.Address, value, id))
	}
	return uris, nil
}
