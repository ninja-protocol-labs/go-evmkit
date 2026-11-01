package v3

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

// ExtractOwnerChanged pulls every OwnerChanged-shaped log out of logs, from any factory;
// filter by Contract afterward.
func ExtractOwnerChanged(logs []tx.Log) ([]OwnerChanged, error) {
	var changed []OwnerChanged
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !ownerChangedEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractOwnerChanged: logs[%d]: %w", i, err)
		}
		vals, err := ownerChangedEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractOwnerChanged: logs[%d]: %w", i, err)
		}
		oldOwner, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractOwnerChanged: logs[%d]: unexpected decoded type for oldOwner: %T", i, vals[0])
		}
		newOwner, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractOwnerChanged: logs[%d]: unexpected decoded type for newOwner: %T", i, vals[1])
		}
		changed = append(changed, *NewOwnerChanged(log.Address, oldOwner, newOwner))
	}
	return changed, nil
}

// ExtractPoolCreated pulls every PoolCreated-shaped log out of logs, from any factory;
// filter by Contract afterward.
func ExtractPoolCreated(logs []tx.Log) ([]PoolCreated, error) {
	var created []PoolCreated
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !poolCreatedEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractPoolCreated: logs[%d]: %w", i, err)
		}
		vals, err := poolCreatedEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractPoolCreated: logs[%d]: %w", i, err)
		}
		token0, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractPoolCreated: logs[%d]: unexpected decoded type for token0: %T", i, vals[0])
		}
		token1, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractPoolCreated: logs[%d]: unexpected decoded type for token1: %T", i, vals[1])
		}
		fee, ok := vals[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractPoolCreated: logs[%d]: unexpected decoded type for fee: %T", i, vals[2])
		}
		tickSpacing, ok := vals[3].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractPoolCreated: logs[%d]: unexpected decoded type for tickSpacing: %T", i, vals[3])
		}
		pool, ok := vals[4].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractPoolCreated: logs[%d]: unexpected decoded type for pool: %T", i, vals[4])
		}
		created = append(created, *NewPoolCreated(log.Address, token0, token1, fee, tickSpacing, pool))
	}
	return created, nil
}

// ExtractFeeAmountEnabled pulls every FeeAmountEnabled-shaped log out of logs, from any
// factory; filter by Contract afterward.
func ExtractFeeAmountEnabled(logs []tx.Log) ([]FeeAmountEnabled, error) {
	var enabled []FeeAmountEnabled
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !feeAmountEnabledEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractFeeAmountEnabled: logs[%d]: %w", i, err)
		}
		vals, err := feeAmountEnabledEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractFeeAmountEnabled: logs[%d]: %w", i, err)
		}
		fee, ok := vals[0].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractFeeAmountEnabled: logs[%d]: unexpected decoded type for fee: %T", i, vals[0])
		}
		tickSpacing, ok := vals[1].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractFeeAmountEnabled: logs[%d]: unexpected decoded type for tickSpacing: %T", i, vals[1])
		}
		enabled = append(enabled, *NewFeeAmountEnabled(log.Address, fee, tickSpacing))
	}
	return enabled, nil
}
