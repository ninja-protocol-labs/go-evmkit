package v3

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

// ExtractIncreaseLiquidity pulls every IncreaseLiquidity-shaped log out of logs, from any NFPM.
func ExtractIncreaseLiquidity(logs []tx.Log) ([]IncreaseLiquidity, error) {
	var increased []IncreaseLiquidity
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !increaseLiquidityEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractIncreaseLiquidity: logs[%d]: %w", i, err)
		}
		vals, err := increaseLiquidityEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractIncreaseLiquidity: logs[%d]: %w", i, err)
		}
		tokenId, ok := vals[0].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractIncreaseLiquidity: logs[%d]: unexpected decoded type for tokenId: %T", i, vals[0])
		}
		liquidity, ok := vals[1].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractIncreaseLiquidity: logs[%d]: unexpected decoded type for liquidity: %T", i, vals[1])
		}
		amount0, ok := vals[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractIncreaseLiquidity: logs[%d]: unexpected decoded type for amount0: %T", i, vals[2])
		}
		amount1, ok := vals[3].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractIncreaseLiquidity: logs[%d]: unexpected decoded type for amount1: %T", i, vals[3])
		}
		increased = append(increased, *NewIncreaseLiquidity(log.Address, tokenId, liquidity, amount0, amount1))
	}
	return increased, nil
}

// ExtractDecreaseLiquidity pulls every DecreaseLiquidity-shaped log out of logs, from any NFPM.
func ExtractDecreaseLiquidity(logs []tx.Log) ([]DecreaseLiquidity, error) {
	var decreased []DecreaseLiquidity
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !decreaseLiquidityEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractDecreaseLiquidity: logs[%d]: %w", i, err)
		}
		vals, err := decreaseLiquidityEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractDecreaseLiquidity: logs[%d]: %w", i, err)
		}
		tokenId, ok := vals[0].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractDecreaseLiquidity: logs[%d]: unexpected decoded type for tokenId: %T", i, vals[0])
		}
		liquidity, ok := vals[1].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractDecreaseLiquidity: logs[%d]: unexpected decoded type for liquidity: %T", i, vals[1])
		}
		amount0, ok := vals[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractDecreaseLiquidity: logs[%d]: unexpected decoded type for amount0: %T", i, vals[2])
		}
		amount1, ok := vals[3].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractDecreaseLiquidity: logs[%d]: unexpected decoded type for amount1: %T", i, vals[3])
		}
		decreased = append(decreased, *NewDecreaseLiquidity(log.Address, tokenId, liquidity, amount0, amount1))
	}
	return decreased, nil
}

// ExtractPositionCollect pulls every Collect-shaped log out of logs, from any NFPM.
func ExtractPositionCollect(logs []tx.Log) ([]PositionCollect, error) {
	var collected []PositionCollect
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !positionCollectEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractPositionCollect: logs[%d]: %w", i, err)
		}
		vals, err := positionCollectEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractPositionCollect: logs[%d]: %w", i, err)
		}
		tokenId, ok := vals[0].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractPositionCollect: logs[%d]: unexpected decoded type for tokenId: %T", i, vals[0])
		}
		recipient, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractPositionCollect: logs[%d]: unexpected decoded type for recipient: %T", i, vals[1])
		}
		amount0, ok := vals[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractPositionCollect: logs[%d]: unexpected decoded type for amount0: %T", i, vals[2])
		}
		amount1, ok := vals[3].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractPositionCollect: logs[%d]: unexpected decoded type for amount1: %T", i, vals[3])
		}
		collected = append(collected, *NewPositionCollect(log.Address, recipient, tokenId, amount0, amount1))
	}
	return collected, nil
}
