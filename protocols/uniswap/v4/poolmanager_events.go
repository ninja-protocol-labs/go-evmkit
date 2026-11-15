package v4

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

// ExtractInitialize pulls every Initialize-shaped log out of logs, from any contract; filter by PoolManager afterward.
func ExtractInitialize(logs []tx.Log) ([]Initialize, error) {
	var inits []Initialize
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
			return nil, fmt.Errorf("uniswapv4: extractInitialize: logs[%d]: %w", i, err)
		}
		vals, err := initializeEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv4: extractInitialize: logs[%d]: %w", i, err)
		}
		idBytes, ok := vals[0].([]byte)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractInitialize: logs[%d]: unexpected decoded type for id: %T", i, vals[0])
		}
		var id PoolID
		copy(id[:], idBytes)
		currency0, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractInitialize: logs[%d]: unexpected decoded type for currency0: %T", i, vals[1])
		}
		currency1, ok := vals[2].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractInitialize: logs[%d]: unexpected decoded type for currency1: %T", i, vals[2])
		}
		fee, ok := vals[3].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractInitialize: logs[%d]: unexpected decoded type for fee: %T", i, vals[3])
		}
		tickSpacing, ok := vals[4].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractInitialize: logs[%d]: unexpected decoded type for tickSpacing: %T", i, vals[4])
		}
		hooks, ok := vals[5].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractInitialize: logs[%d]: unexpected decoded type for hooks: %T", i, vals[5])
		}
		sqrtPriceX96, ok := vals[6].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractInitialize: logs[%d]: unexpected decoded type for sqrtPriceX96: %T", i, vals[6])
		}
		tick, ok := vals[7].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractInitialize: logs[%d]: unexpected decoded type for tick: %T", i, vals[7])
		}
		inits = append(inits, *NewInitialize(log.Address, id, currency0, currency1, fee, tickSpacing, hooks, sqrtPriceX96, tick))
	}
	return inits, nil
}

// ExtractModifyLiquidity pulls every ModifyLiquidity-shaped log out of logs; see ExtractInitialize.
func ExtractModifyLiquidity(logs []tx.Log) ([]ModifyLiquidity, error) {
	var mods []ModifyLiquidity
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !modifyLiquidityEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv4: extractModifyLiquidity: logs[%d]: %w", i, err)
		}
		vals, err := modifyLiquidityEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv4: extractModifyLiquidity: logs[%d]: %w", i, err)
		}
		idBytes, ok := vals[0].([]byte)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractModifyLiquidity: logs[%d]: unexpected decoded type for id: %T", i, vals[0])
		}
		var id PoolID
		copy(id[:], idBytes)
		sender, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractModifyLiquidity: logs[%d]: unexpected decoded type for sender: %T", i, vals[1])
		}
		tickLower, ok := vals[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractModifyLiquidity: logs[%d]: unexpected decoded type for tickLower: %T", i, vals[2])
		}
		tickUpper, ok := vals[3].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractModifyLiquidity: logs[%d]: unexpected decoded type for tickUpper: %T", i, vals[3])
		}
		liquidityDelta, ok := vals[4].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractModifyLiquidity: logs[%d]: unexpected decoded type for liquidityDelta: %T", i, vals[4])
		}
		saltBytes, ok := vals[5].([]byte)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractModifyLiquidity: logs[%d]: unexpected decoded type for salt: %T", i, vals[5])
		}
		var salt [32]byte
		copy(salt[:], saltBytes)
		mods = append(mods, *NewModifyLiquidity(log.Address, id, sender, tickLower, tickUpper, liquidityDelta, salt))
	}
	return mods, nil
}

// ExtractSwap pulls every Swap-shaped log out of logs; see ExtractInitialize.
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
			return nil, fmt.Errorf("uniswapv4: extractSwap: logs[%d]: %w", i, err)
		}
		vals, err := swapEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv4: extractSwap: logs[%d]: %w", i, err)
		}
		idBytes, ok := vals[0].([]byte)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractSwap: logs[%d]: unexpected decoded type for id: %T", i, vals[0])
		}
		var id PoolID
		copy(id[:], idBytes)
		sender, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractSwap: logs[%d]: unexpected decoded type for sender: %T", i, vals[1])
		}
		amount0, ok := vals[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractSwap: logs[%d]: unexpected decoded type for amount0: %T", i, vals[2])
		}
		amount1, ok := vals[3].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractSwap: logs[%d]: unexpected decoded type for amount1: %T", i, vals[3])
		}
		sqrtPriceX96, ok := vals[4].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractSwap: logs[%d]: unexpected decoded type for sqrtPriceX96: %T", i, vals[4])
		}
		liquidity, ok := vals[5].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractSwap: logs[%d]: unexpected decoded type for liquidity: %T", i, vals[5])
		}
		tick, ok := vals[6].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractSwap: logs[%d]: unexpected decoded type for tick: %T", i, vals[6])
		}
		fee, ok := vals[7].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractSwap: logs[%d]: unexpected decoded type for fee: %T", i, vals[7])
		}
		swaps = append(swaps, *NewSwap(log.Address, id, sender, amount0, amount1, sqrtPriceX96, liquidity, tick, fee))
	}
	return swaps, nil
}

// ExtractDonate pulls every Donate-shaped log out of logs; see ExtractInitialize.
func ExtractDonate(logs []tx.Log) ([]Donate, error) {
	var donates []Donate
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !donateEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv4: extractDonate: logs[%d]: %w", i, err)
		}
		vals, err := donateEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv4: extractDonate: logs[%d]: %w", i, err)
		}
		idBytes, ok := vals[0].([]byte)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractDonate: logs[%d]: unexpected decoded type for id: %T", i, vals[0])
		}
		var id PoolID
		copy(id[:], idBytes)
		sender, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractDonate: logs[%d]: unexpected decoded type for sender: %T", i, vals[1])
		}
		amount0, ok := vals[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractDonate: logs[%d]: unexpected decoded type for amount0: %T", i, vals[2])
		}
		amount1, ok := vals[3].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractDonate: logs[%d]: unexpected decoded type for amount1: %T", i, vals[3])
		}
		donates = append(donates, *NewDonate(log.Address, id, sender, amount0, amount1))
	}
	return donates, nil
}

// ExtractProtocolFeeControllerUpdated pulls every ProtocolFeeControllerUpdated-shaped log out of logs; see ExtractInitialize.
func ExtractProtocolFeeControllerUpdated(logs []tx.Log) ([]ProtocolFeeControllerUpdated, error) {
	var updates []ProtocolFeeControllerUpdated
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !protocolFeeControllerUpdatedEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv4: extractProtocolFeeControllerUpdated: logs[%d]: %w", i, err)
		}
		vals, err := protocolFeeControllerUpdatedEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv4: extractProtocolFeeControllerUpdated: logs[%d]: %w", i, err)
		}
		controller, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractProtocolFeeControllerUpdated: logs[%d]: unexpected decoded type for protocolFeeController: %T", i, vals[0])
		}
		updates = append(updates, *NewProtocolFeeControllerUpdated(log.Address, controller))
	}
	return updates, nil
}

// ExtractProtocolFeeUpdated pulls every ProtocolFeeUpdated-shaped log out of logs; see ExtractInitialize.
func ExtractProtocolFeeUpdated(logs []tx.Log) ([]ProtocolFeeUpdated, error) {
	var updates []ProtocolFeeUpdated
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !protocolFeeUpdatedEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv4: extractProtocolFeeUpdated: logs[%d]: %w", i, err)
		}
		vals, err := protocolFeeUpdatedEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv4: extractProtocolFeeUpdated: logs[%d]: %w", i, err)
		}
		idBytes, ok := vals[0].([]byte)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractProtocolFeeUpdated: logs[%d]: unexpected decoded type for id: %T", i, vals[0])
		}
		var id PoolID
		copy(id[:], idBytes)
		protocolFee, ok := vals[1].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractProtocolFeeUpdated: logs[%d]: unexpected decoded type for protocolFee: %T", i, vals[1])
		}
		updates = append(updates, *NewProtocolFeeUpdated(log.Address, id, protocolFee))
	}
	return updates, nil
}
