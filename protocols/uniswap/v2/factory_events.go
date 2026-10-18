package v2

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

// ExtractPairCreated pulls every PairCreated-shaped log out of logs, from any factory;
// filter by Contract afterward.
func ExtractPairCreated(logs []tx.Log) ([]PairCreated, error) {
	var created []PairCreated
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !pairCreatedEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv2: extractPairCreated: logs[%d]: %w", i, err)
		}
		vals, err := pairCreatedEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv2: extractPairCreated: logs[%d]: %w", i, err)
		}
		token0, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv2: extractPairCreated: logs[%d]: unexpected decoded type for token0: %T", i, vals[0])
		}
		token1, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv2: extractPairCreated: logs[%d]: unexpected decoded type for token1: %T", i, vals[1])
		}
		pair, ok := vals[2].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv2: extractPairCreated: logs[%d]: unexpected decoded type for pair: %T", i, vals[2])
		}
		allPairsLength, ok := vals[3].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv2: extractPairCreated: logs[%d]: unexpected decoded type for allPairsLength: %T", i, vals[3])
		}
		created = append(created, *NewPairCreated(log.Address, token0, token1, pair, allPairsLength))
	}
	return created, nil
}
