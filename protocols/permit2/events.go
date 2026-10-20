package permit2

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

// ExtractNonceInvalidations pulls every NonceInvalidation-shaped log out of logs, from any
// Permit2-shaped contract; filter by Contract afterward.
func ExtractNonceInvalidations(logs []tx.Log) ([]NonceInvalidation, error) {
	var out []NonceInvalidation
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !nonceInvalidationEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("permit2: extractNonceInvalidations: logs[%d]: %w", i, err)
		}
		vals, err := nonceInvalidationEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("permit2: extractNonceInvalidations: logs[%d]: %w", i, err)
		}
		owner, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("permit2: extractNonceInvalidations: logs[%d]: unexpected decoded type for owner: %T", i, vals[0])
		}
		token, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("permit2: extractNonceInvalidations: logs[%d]: unexpected decoded type for token: %T", i, vals[1])
		}
		spender, ok := vals[2].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("permit2: extractNonceInvalidations: logs[%d]: unexpected decoded type for spender: %T", i, vals[2])
		}
		newNonce, ok := vals[3].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("permit2: extractNonceInvalidations: logs[%d]: unexpected decoded type for newNonce: %T", i, vals[3])
		}
		oldNonce, ok := vals[4].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("permit2: extractNonceInvalidations: logs[%d]: unexpected decoded type for oldNonce: %T", i, vals[4])
		}
		out = append(out, *NewNonceInvalidation(log.Address, owner, token, spender, newNonce, oldNonce))
	}
	return out, nil
}

// ExtractApprovals pulls every Approval-shaped log out of logs, from any Permit2-shaped contract;
// filter by Contract afterward.
func ExtractApprovals(logs []tx.Log) ([]Approval, error) {
	var out []Approval
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
			return nil, fmt.Errorf("permit2: extractApprovals: logs[%d]: %w", i, err)
		}
		vals, err := approvalEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("permit2: extractApprovals: logs[%d]: %w", i, err)
		}
		owner, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("permit2: extractApprovals: logs[%d]: unexpected decoded type for owner: %T", i, vals[0])
		}
		token, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("permit2: extractApprovals: logs[%d]: unexpected decoded type for token: %T", i, vals[1])
		}
		spender, ok := vals[2].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("permit2: extractApprovals: logs[%d]: unexpected decoded type for spender: %T", i, vals[2])
		}
		amount, ok := vals[3].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("permit2: extractApprovals: logs[%d]: unexpected decoded type for amount: %T", i, vals[3])
		}
		expiration, ok := vals[4].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("permit2: extractApprovals: logs[%d]: unexpected decoded type for expiration: %T", i, vals[4])
		}
		out = append(out, *NewApproval(log.Address, owner, token, spender, amount, expiration))
	}
	return out, nil
}

// ExtractPermits pulls every Permit-shaped log out of logs, from any Permit2-shaped contract;
// filter by Contract afterward.
func ExtractPermits(logs []tx.Log) ([]Permit, error) {
	var out []Permit
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !permitEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("permit2: extractPermits: logs[%d]: %w", i, err)
		}
		vals, err := permitEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("permit2: extractPermits: logs[%d]: %w", i, err)
		}
		owner, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("permit2: extractPermits: logs[%d]: unexpected decoded type for owner: %T", i, vals[0])
		}
		token, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("permit2: extractPermits: logs[%d]: unexpected decoded type for token: %T", i, vals[1])
		}
		spender, ok := vals[2].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("permit2: extractPermits: logs[%d]: unexpected decoded type for spender: %T", i, vals[2])
		}
		amount, ok := vals[3].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("permit2: extractPermits: logs[%d]: unexpected decoded type for amount: %T", i, vals[3])
		}
		expiration, ok := vals[4].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("permit2: extractPermits: logs[%d]: unexpected decoded type for expiration: %T", i, vals[4])
		}
		nonce, ok := vals[5].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("permit2: extractPermits: logs[%d]: unexpected decoded type for nonce: %T", i, vals[5])
		}
		out = append(out, *NewPermit(log.Address, owner, token, spender, amount, expiration, nonce))
	}
	return out, nil
}

// ExtractLockdowns pulls every Lockdown-shaped log out of logs, from any Permit2-shaped contract;
// filter by Contract afterward.
func ExtractLockdowns(logs []tx.Log) ([]Lockdown, error) {
	var out []Lockdown
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !lockdownEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("permit2: extractLockdowns: logs[%d]: %w", i, err)
		}
		vals, err := lockdownEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("permit2: extractLockdowns: logs[%d]: %w", i, err)
		}
		owner, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("permit2: extractLockdowns: logs[%d]: unexpected decoded type for owner: %T", i, vals[0])
		}
		token, ok := vals[1].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("permit2: extractLockdowns: logs[%d]: unexpected decoded type for token: %T", i, vals[1])
		}
		spender, ok := vals[2].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("permit2: extractLockdowns: logs[%d]: unexpected decoded type for spender: %T", i, vals[2])
		}
		out = append(out, *NewLockdown(log.Address, owner, token, spender))
	}
	return out, nil
}

// ExtractUnorderedNonceInvalidations pulls every UnorderedNonceInvalidation-shaped log out of
// logs, from any Permit2-shaped contract; filter by Contract afterward.
func ExtractUnorderedNonceInvalidations(logs []tx.Log) ([]UnorderedNonceInvalidation, error) {
	var out []UnorderedNonceInvalidation
	for i, log := range logs {
		if len(log.Topics) == 0 || log.Topics[0] == nil {
			continue
		}
		var topic0 abi.Topic
		copy(topic0[:], log.Topics[0].Bytes())
		if !unorderedNonceInvalidationEvent.Matches(topic0) {
			continue
		}

		topics, err := abi.TopicsFromHashes(log.Topics)
		if err != nil {
			return nil, fmt.Errorf("permit2: extractUnorderedNonceInvalidations: logs[%d]: %w", i, err)
		}
		vals, err := unorderedNonceInvalidationEvent.Decode(topics, log.Data)
		if err != nil {
			return nil, fmt.Errorf("permit2: extractUnorderedNonceInvalidations: logs[%d]: %w", i, err)
		}
		owner, ok := vals[0].(*types.Address)
		if !ok {
			return nil, fmt.Errorf("permit2: extractUnorderedNonceInvalidations: logs[%d]: unexpected decoded type for owner: %T", i, vals[0])
		}
		word, ok := vals[1].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("permit2: extractUnorderedNonceInvalidations: logs[%d]: unexpected decoded type for word: %T", i, vals[1])
		}
		mask, ok := vals[2].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("permit2: extractUnorderedNonceInvalidations: logs[%d]: unexpected decoded type for mask: %T", i, vals[2])
		}
		out = append(out, *NewUnorderedNonceInvalidation(log.Address, owner, word, mask))
	}
	return out, nil
}
