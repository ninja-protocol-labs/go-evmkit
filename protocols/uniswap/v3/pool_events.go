package v3

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

// ExtractInitialize pulls every Initialize-shaped log out of logs, from any pool.
func ExtractInitialize(logs []tx.Log) ([]Initialize, error) {
	var initialized []Initialize
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !initializeEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractInitialize: logs[%d]: %w", i, err)
		}
		vals, err := initializeEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractInitialize: logs[%d]: %w", i, err)
		}
		sqrtPriceX96, ok := vals[0].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractInitialize: logs[%d]: unexpected decoded type for sqrtPriceX96: %T", i, vals[0])
		}
		tick, ok := vals[1].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractInitialize: logs[%d]: unexpected decoded type for tick: %T", i, vals[1])
		}
		initialized = append(initialized, *NewInitialize(log.Address, sqrtPriceX96, tick))
	}
	return initialized, nil
}

// ExtractMint pulls every Mint-shaped log out of logs, from any pool.
func ExtractMint(logs []tx.Log) ([]Mint, error) {
	var minted []Mint
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
			return nil, fmt.Errorf("uniswapv3: extractMint: logs[%d]: %w", i, err)
		}
		vals, err := mintEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractMint: logs[%d]: %w", i, err)
		}
		sender, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractMint: logs[%d]: unexpected decoded type for sender: %T", i, vals[0])
		}
		owner, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractMint: logs[%d]: unexpected decoded type for owner: %T", i, vals[1])
		}
		tickLower, ok := vals[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractMint: logs[%d]: unexpected decoded type for tickLower: %T", i, vals[2])
		}
		tickUpper, ok := vals[3].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractMint: logs[%d]: unexpected decoded type for tickUpper: %T", i, vals[3])
		}
		amount, ok := vals[4].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractMint: logs[%d]: unexpected decoded type for amount: %T", i, vals[4])
		}
		amount0, ok := vals[5].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractMint: logs[%d]: unexpected decoded type for amount0: %T", i, vals[5])
		}
		amount1, ok := vals[6].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractMint: logs[%d]: unexpected decoded type for amount1: %T", i, vals[6])
		}
		minted = append(minted, *NewMint(log.Address, sender, owner, tickLower, tickUpper, amount, amount0, amount1))
	}
	return minted, nil
}

// ExtractCollect pulls every Collect-shaped log out of logs, from any pool.
func ExtractCollect(logs []tx.Log) ([]Collect, error) {
	var collected []Collect
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !collectEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractCollect: logs[%d]: %w", i, err)
		}
		vals, err := collectEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractCollect: logs[%d]: %w", i, err)
		}
		owner, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractCollect: logs[%d]: unexpected decoded type for owner: %T", i, vals[0])
		}
		recipient, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractCollect: logs[%d]: unexpected decoded type for recipient: %T", i, vals[1])
		}
		tickLower, ok := vals[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractCollect: logs[%d]: unexpected decoded type for tickLower: %T", i, vals[2])
		}
		tickUpper, ok := vals[3].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractCollect: logs[%d]: unexpected decoded type for tickUpper: %T", i, vals[3])
		}
		amount0, ok := vals[4].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractCollect: logs[%d]: unexpected decoded type for amount0: %T", i, vals[4])
		}
		amount1, ok := vals[5].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractCollect: logs[%d]: unexpected decoded type for amount1: %T", i, vals[5])
		}
		collected = append(collected, *NewCollect(log.Address, owner, recipient, tickLower, tickUpper, amount0, amount1))
	}
	return collected, nil
}

// ExtractBurn pulls every Burn-shaped log out of logs, from any pool.
func ExtractBurn(logs []tx.Log) ([]Burn, error) {
	var burned []Burn
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !poolBurnEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractBurn: logs[%d]: %w", i, err)
		}
		vals, err := poolBurnEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractBurn: logs[%d]: %w", i, err)
		}
		owner, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractBurn: logs[%d]: unexpected decoded type for owner: %T", i, vals[0])
		}
		tickLower, ok := vals[1].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractBurn: logs[%d]: unexpected decoded type for tickLower: %T", i, vals[1])
		}
		tickUpper, ok := vals[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractBurn: logs[%d]: unexpected decoded type for tickUpper: %T", i, vals[2])
		}
		amount, ok := vals[3].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractBurn: logs[%d]: unexpected decoded type for amount: %T", i, vals[3])
		}
		amount0, ok := vals[4].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractBurn: logs[%d]: unexpected decoded type for amount0: %T", i, vals[4])
		}
		amount1, ok := vals[5].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractBurn: logs[%d]: unexpected decoded type for amount1: %T", i, vals[5])
		}
		burned = append(burned, *NewBurn(log.Address, owner, tickLower, tickUpper, amount, amount0, amount1))
	}
	return burned, nil
}

// ExtractSwap pulls every Swap-shaped log out of logs, from any pool.
func ExtractSwap(logs []tx.Log) ([]Swap, error) {
	var swapped []Swap
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
			return nil, fmt.Errorf("uniswapv3: extractSwap: logs[%d]: %w", i, err)
		}
		vals, err := swapEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractSwap: logs[%d]: %w", i, err)
		}
		sender, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractSwap: logs[%d]: unexpected decoded type for sender: %T", i, vals[0])
		}
		recipient, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractSwap: logs[%d]: unexpected decoded type for recipient: %T", i, vals[1])
		}
		amount0, ok := vals[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractSwap: logs[%d]: unexpected decoded type for amount0: %T", i, vals[2])
		}
		amount1, ok := vals[3].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractSwap: logs[%d]: unexpected decoded type for amount1: %T", i, vals[3])
		}
		sqrtPriceX96, ok := vals[4].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractSwap: logs[%d]: unexpected decoded type for sqrtPriceX96: %T", i, vals[4])
		}
		liquidity, ok := vals[5].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractSwap: logs[%d]: unexpected decoded type for liquidity: %T", i, vals[5])
		}
		tick, ok := vals[6].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractSwap: logs[%d]: unexpected decoded type for tick: %T", i, vals[6])
		}
		swapped = append(swapped, *NewSwap(log.Address, sender, recipient, amount0, amount1, sqrtPriceX96, liquidity, tick))
	}
	return swapped, nil
}

// ExtractFlash pulls every Flash-shaped log out of logs, from any pool.
func ExtractFlash(logs []tx.Log) ([]Flash, error) {
	var flashed []Flash
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !flashEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractFlash: logs[%d]: %w", i, err)
		}
		vals, err := flashEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractFlash: logs[%d]: %w", i, err)
		}
		sender, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractFlash: logs[%d]: unexpected decoded type for sender: %T", i, vals[0])
		}
		recipient, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractFlash: logs[%d]: unexpected decoded type for recipient: %T", i, vals[1])
		}
		amount0, ok := vals[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractFlash: logs[%d]: unexpected decoded type for amount0: %T", i, vals[2])
		}
		amount1, ok := vals[3].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractFlash: logs[%d]: unexpected decoded type for amount1: %T", i, vals[3])
		}
		paid0, ok := vals[4].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractFlash: logs[%d]: unexpected decoded type for paid0: %T", i, vals[4])
		}
		paid1, ok := vals[5].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractFlash: logs[%d]: unexpected decoded type for paid1: %T", i, vals[5])
		}
		flashed = append(flashed, *NewFlash(log.Address, sender, recipient, amount0, amount1, paid0, paid1))
	}
	return flashed, nil
}

// ExtractIncreaseObservationCardinalityNext pulls every IncreaseObservationCardinalityNext-shaped
// log out of logs, from any pool.
func ExtractIncreaseObservationCardinalityNext(logs []tx.Log) ([]IncreaseObservationCardinalityNext, error) {
	var increased []IncreaseObservationCardinalityNext
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !increaseObservationCardinalityNextEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractIncreaseObservationCardinalityNext: logs[%d]: %w", i, err)
		}
		vals, err := increaseObservationCardinalityNextEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractIncreaseObservationCardinalityNext: logs[%d]: %w", i, err)
		}
		old, ok := vals[0].(uint16)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractIncreaseObservationCardinalityNext: logs[%d]: unexpected decoded type for old: %T", i, vals[0])
		}
		new_, ok := vals[1].(uint16)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractIncreaseObservationCardinalityNext: logs[%d]: unexpected decoded type for new: %T", i, vals[1])
		}
		increased = append(increased, *NewIncreaseObservationCardinalityNext(log.Address, big.NewInt(int64(old)), big.NewInt(int64(new_))))
	}
	return increased, nil
}

// ExtractSetFeeProtocol pulls every SetFeeProtocol-shaped log out of logs, from any pool.
func ExtractSetFeeProtocol(logs []tx.Log) ([]SetFeeProtocol, error) {
	var set []SetFeeProtocol
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !setFeeProtocolEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractSetFeeProtocol: logs[%d]: %w", i, err)
		}
		vals, err := setFeeProtocolEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractSetFeeProtocol: logs[%d]: %w", i, err)
		}
		feeProtocol0Old, ok := vals[0].(uint8)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractSetFeeProtocol: logs[%d]: unexpected decoded type for feeProtocol0Old: %T", i, vals[0])
		}
		feeProtocol1Old, ok := vals[1].(uint8)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractSetFeeProtocol: logs[%d]: unexpected decoded type for feeProtocol1Old: %T", i, vals[1])
		}
		feeProtocol0New, ok := vals[2].(uint8)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractSetFeeProtocol: logs[%d]: unexpected decoded type for feeProtocol0New: %T", i, vals[2])
		}
		feeProtocol1New, ok := vals[3].(uint8)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractSetFeeProtocol: logs[%d]: unexpected decoded type for feeProtocol1New: %T", i, vals[3])
		}
		set = append(set, *NewSetFeeProtocol(log.Address, big.NewInt(int64(feeProtocol0Old)), big.NewInt(int64(feeProtocol1Old)), big.NewInt(int64(feeProtocol0New)), big.NewInt(int64(feeProtocol1New))))
	}
	return set, nil
}

// ExtractCollectProtocol pulls every CollectProtocol-shaped log out of logs, from any pool.
func ExtractCollectProtocol(logs []tx.Log) ([]CollectProtocol, error) {
	var collected []CollectProtocol
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !collectProtocolEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractCollectProtocol: logs[%d]: %w", i, err)
		}
		vals, err := collectProtocolEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv3: extractCollectProtocol: logs[%d]: %w", i, err)
		}
		sender, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractCollectProtocol: logs[%d]: unexpected decoded type for sender: %T", i, vals[0])
		}
		recipient, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractCollectProtocol: logs[%d]: unexpected decoded type for recipient: %T", i, vals[1])
		}
		amount0, ok := vals[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractCollectProtocol: logs[%d]: unexpected decoded type for amount0: %T", i, vals[2])
		}
		amount1, ok := vals[3].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv3: extractCollectProtocol: logs[%d]: unexpected decoded type for amount1: %T", i, vals[3])
		}
		collected = append(collected, *NewCollectProtocol(log.Address, sender, recipient, amount0, amount1))
	}
	return collected, nil
}
