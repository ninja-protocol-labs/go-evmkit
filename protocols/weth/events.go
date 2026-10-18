package weth

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

// ExtractDeposits pulls every Deposit-shaped log out of logs, from any WETH9-shaped contract;
// filter by Contract afterward.
func ExtractDeposits(logs []tx.Log) ([]Deposit, error) {
	var deposits []Deposit
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !depositEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("weth: extractDeposits: logs[%d]: %w", i, err)
		}
		vals, err := depositEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("weth: extractDeposits: logs[%d]: %w", i, err)
		}
		dst, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("weth: extractDeposits: logs[%d]: unexpected decoded type for dst: %T", i, vals[0])
		}
		wad, ok := vals[1].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("weth: extractDeposits: logs[%d]: unexpected decoded type for wad: %T", i, vals[1])
		}
		deposits = append(deposits, *NewDeposit(log.Address, dst, wad))
	}
	return deposits, nil
}

// ExtractWithdrawals pulls every Withdrawal-shaped log out of logs, from any WETH9-shaped
// contract; filter by Contract afterward.
func ExtractWithdrawals(logs []tx.Log) ([]Withdrawal, error) {
	var withdrawals []Withdrawal
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !withdrawalEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("weth: extractWithdrawals: logs[%d]: %w", i, err)
		}
		vals, err := withdrawalEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("weth: extractWithdrawals: logs[%d]: %w", i, err)
		}
		src, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("weth: extractWithdrawals: logs[%d]: unexpected decoded type for src: %T", i, vals[0])
		}
		wad, ok := vals[1].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("weth: extractWithdrawals: logs[%d]: unexpected decoded type for wad: %T", i, vals[1])
		}
		withdrawals = append(withdrawals, *NewWithdrawal(log.Address, src, wad))
	}
	return withdrawals, nil
}
