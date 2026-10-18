package v2

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

// ExtractMint pulls every Mint-shaped log out of logs, from any pair; filter by Contract afterward.
func ExtractMint(logs []tx.Log) ([]Mint, error) {
	var mints []Mint
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !mintEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv2: extractMint: logs[%d]: %w", i, err)
		}
		vals, err := mintEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv2: extractMint: logs[%d]: %w", i, err)
		}
		sender, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv2: extractMint: logs[%d]: unexpected decoded type for sender: %T", i, vals[0])
		}
		amount0, ok := vals[1].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv2: extractMint: logs[%d]: unexpected decoded type for amount0: %T", i, vals[1])
		}
		amount1, ok := vals[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv2: extractMint: logs[%d]: unexpected decoded type for amount1: %T", i, vals[2])
		}
		mints = append(mints, *NewMint(log.Address, sender, amount0, amount1))
	}
	return mints, nil
}

// ExtractBurn pulls every Burn-shaped log out of logs, from any pair; filter by Contract afterward.
func ExtractBurn(logs []tx.Log) ([]Burn, error) {
	var burns []Burn
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !burnEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv2: extractBurn: logs[%d]: %w", i, err)
		}
		vals, err := burnEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv2: extractBurn: logs[%d]: %w", i, err)
		}
		sender, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv2: extractBurn: logs[%d]: unexpected decoded type for sender: %T", i, vals[0])
		}
		amount0, ok := vals[1].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv2: extractBurn: logs[%d]: unexpected decoded type for amount0: %T", i, vals[1])
		}
		amount1, ok := vals[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv2: extractBurn: logs[%d]: unexpected decoded type for amount1: %T", i, vals[2])
		}
		to, ok := vals[3].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv2: extractBurn: logs[%d]: unexpected decoded type for to: %T", i, vals[3])
		}
		burns = append(burns, *NewBurn(log.Address, sender, amount0, amount1, to))
	}
	return burns, nil
}

// ExtractSwap pulls every Swap-shaped log out of logs, from any pair; filter by Contract afterward.
func ExtractSwap(logs []tx.Log) ([]Swap, error) {
	var swaps []Swap
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !swapEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv2: extractSwap: logs[%d]: %w", i, err)
		}
		vals, err := swapEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv2: extractSwap: logs[%d]: %w", i, err)
		}
		sender, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv2: extractSwap: logs[%d]: unexpected decoded type for sender: %T", i, vals[0])
		}
		amount0In, ok := vals[1].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv2: extractSwap: logs[%d]: unexpected decoded type for amount0In: %T", i, vals[1])
		}
		amount1In, ok := vals[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv2: extractSwap: logs[%d]: unexpected decoded type for amount1In: %T", i, vals[2])
		}
		amount0Out, ok := vals[3].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv2: extractSwap: logs[%d]: unexpected decoded type for amount0Out: %T", i, vals[3])
		}
		amount1Out, ok := vals[4].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv2: extractSwap: logs[%d]: unexpected decoded type for amount1Out: %T", i, vals[4])
		}
		to, ok := vals[5].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv2: extractSwap: logs[%d]: unexpected decoded type for to: %T", i, vals[5])
		}
		swaps = append(swaps, *NewSwap(log.Address, sender, amount0In, amount1In, amount0Out, amount1Out, to))
	}
	return swaps, nil
}

// ExtractSync pulls every Sync-shaped log out of logs, from any pair; filter by Contract afterward.
func ExtractSync(logs []tx.Log) ([]Sync, error) {
	var syncs []Sync
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !syncEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv2: extractSync: logs[%d]: %w", i, err)
		}
		vals, err := syncEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv2: extractSync: logs[%d]: %w", i, err)
		}
		reserve0, ok := vals[0].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv2: extractSync: logs[%d]: unexpected decoded type for reserve0: %T", i, vals[0])
		}
		reserve1, ok := vals[1].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv2: extractSync: logs[%d]: unexpected decoded type for reserve1: %T", i, vals[1])
		}
		syncs = append(syncs, *NewSync(log.Address, reserve0, reserve1))
	}
	return syncs, nil
}
