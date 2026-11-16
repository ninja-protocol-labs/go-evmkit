package v4

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

// ExtractModifyPosition pulls every ModifyPosition-shaped log out of logs, from any contract;
// filter by PositionManager afterward.
func ExtractModifyPosition(logs []tx.Log) ([]ModifyPosition, error) {
	var mods []ModifyPosition
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !modifyPositionEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv4: extractModifyPosition: logs[%d]: %w", i, err)
		}
		vals, err := modifyPositionEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv4: extractModifyPosition: logs[%d]: %w", i, err)
		}
		idBytes, ok := vals[0].([]byte)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractModifyPosition: logs[%d]: unexpected decoded type for id: %T", i, vals[0])
		}
		var id PoolID
		copy(id[:], idBytes)
		sender, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractModifyPosition: logs[%d]: unexpected decoded type for sender: %T", i, vals[1])
		}
		tickLower, ok := vals[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractModifyPosition: logs[%d]: unexpected decoded type for tickLower: %T", i, vals[2])
		}
		tickUpper, ok := vals[3].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractModifyPosition: logs[%d]: unexpected decoded type for tickUpper: %T", i, vals[3])
		}
		liquidityDelta, ok := vals[4].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractModifyPosition: logs[%d]: unexpected decoded type for liquidityDelta: %T", i, vals[4])
		}
		saltBytes, ok := vals[5].([]byte)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractModifyPosition: logs[%d]: unexpected decoded type for salt: %T", i, vals[5])
		}
		var salt [32]byte
		copy(salt[:], saltBytes)
		mods = append(mods, *NewModifyPosition(log.Address, id, sender, tickLower, tickUpper, liquidityDelta, salt))
	}
	return mods, nil
}

// ExtractSubscription pulls every Subscription-shaped log out of logs; see ExtractModifyPosition.
func ExtractSubscription(logs []tx.Log) ([]Subscription, error) {
	var subs []Subscription
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !subscriptionEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv4: extractSubscription: logs[%d]: %w", i, err)
		}
		vals, err := subscriptionEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv4: extractSubscription: logs[%d]: %w", i, err)
		}
		tokenID, ok := vals[0].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractSubscription: logs[%d]: unexpected decoded type for tokenId: %T", i, vals[0])
		}
		subscriber, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractSubscription: logs[%d]: unexpected decoded type for subscriber: %T", i, vals[1])
		}
		subs = append(subs, *NewSubscription(log.Address, tokenID, subscriber))
	}
	return subs, nil
}

// ExtractUnsubscription pulls every Unsubscription-shaped log out of logs; see ExtractModifyPosition.
func ExtractUnsubscription(logs []tx.Log) ([]Unsubscription, error) {
	var unsubs []Unsubscription
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !unsubscriptionEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("uniswapv4: extractUnsubscription: logs[%d]: %w", i, err)
		}
		vals, err := unsubscriptionEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("uniswapv4: extractUnsubscription: logs[%d]: %w", i, err)
		}
		tokenID, ok := vals[0].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractUnsubscription: logs[%d]: unexpected decoded type for tokenId: %T", i, vals[0])
		}
		subscriber, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("uniswapv4: extractUnsubscription: logs[%d]: unexpected decoded type for subscriber: %T", i, vals[1])
		}
		unsubs = append(unsubs, *NewUnsubscription(log.Address, tokenID, subscriber))
	}
	return unsubs, nil
}
