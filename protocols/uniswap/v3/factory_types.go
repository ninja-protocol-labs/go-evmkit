package v3

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// OwnerRequest is owner()'s input on Contract.
type OwnerRequest struct {
	Contract *types.Address
}

func NewOwnerRequest(contract *types.Address) *OwnerRequest {
	return &OwnerRequest{
		Contract: contract,
	}
}

func (r *OwnerRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// FeeAmountTickSpacingRequest is feeAmountTickSpacing(fee)'s input on Contract.
type FeeAmountTickSpacingRequest struct {
	Contract *types.Address
	Fee      *big.Int
}

func NewFeeAmountTickSpacingRequest(contract *types.Address, fee *big.Int) *FeeAmountTickSpacingRequest {
	return &FeeAmountTickSpacingRequest{
		Contract: contract,
		Fee:      fee,
	}
}

func (r *FeeAmountTickSpacingRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Fee == nil {
		return fmt.Errorf("%w: fee is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetPoolRequest is getPool(tokenA, tokenB, fee)'s input on Contract.
type GetPoolRequest struct {
	Contract *types.Address
	TokenA   *types.Address
	TokenB   *types.Address
	Fee      *big.Int
}

func NewGetPoolRequest(contract, tokenA, tokenB *types.Address, fee *big.Int) *GetPoolRequest {
	return &GetPoolRequest{
		Contract: contract,
		TokenA:   tokenA,
		TokenB:   tokenB,
		Fee:      fee,
	}
}

func (r *GetPoolRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.TokenA == nil {
		return fmt.Errorf("%w: tokenA is nil", bind.ErrInvalidRequest)
	}

	if r.TokenB == nil {
		return fmt.Errorf("%w: tokenB is nil", bind.ErrInvalidRequest)
	}

	if r.Fee == nil {
		return fmt.Errorf("%w: fee is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// CreatePoolRequest is createPool(tokenA, tokenB, fee)'s input on Contract.
type CreatePoolRequest struct {
	Contract *types.Address
	TokenA   *types.Address
	TokenB   *types.Address
	Fee      *big.Int
}

func NewCreatePoolRequest(contract, tokenA, tokenB *types.Address, fee *big.Int) *CreatePoolRequest {
	return &CreatePoolRequest{
		Contract: contract,
		TokenA:   tokenA,
		TokenB:   tokenB,
		Fee:      fee,
	}
}

func (r *CreatePoolRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.TokenA == nil {
		return fmt.Errorf("%w: tokenA is nil", bind.ErrInvalidRequest)
	}

	if r.TokenB == nil {
		return fmt.Errorf("%w: tokenB is nil", bind.ErrInvalidRequest)
	}

	if r.Fee == nil {
		return fmt.Errorf("%w: fee is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SetOwnerRequest is setOwner(_owner)'s input on Contract.
type SetOwnerRequest struct {
	Contract *types.Address
	Owner    *types.Address
}

func NewSetOwnerRequest(contract, owner *types.Address) *SetOwnerRequest {
	return &SetOwnerRequest{
		Contract: contract,
		Owner:    owner,
	}
}

func (r *SetOwnerRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Owner == nil {
		return fmt.Errorf("%w: owner is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// EnableFeeAmountRequest is enableFeeAmount(fee, tickSpacing)'s input on Contract.
type EnableFeeAmountRequest struct {
	Contract    *types.Address
	Fee         *big.Int
	TickSpacing *big.Int
}

func NewEnableFeeAmountRequest(contract *types.Address, fee, tickSpacing *big.Int) *EnableFeeAmountRequest {
	return &EnableFeeAmountRequest{
		Contract:    contract,
		Fee:         fee,
		TickSpacing: tickSpacing,
	}
}

func (r *EnableFeeAmountRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Fee == nil {
		return fmt.Errorf("%w: fee is nil", bind.ErrInvalidRequest)
	}

	if r.TickSpacing == nil {
		return fmt.Errorf("%w: tickSpacing is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// OwnerChangedEvent mirrors IUniswapV3Factory's OwnerChanged event: Contract's owner moved from OldOwner to NewOwner via setOwner.
type OwnerChangedEvent struct {
	Contract *types.Address
	OldOwner *types.Address
	NewOwner *types.Address
}

func NewOwnerChangedEvent(contract, oldOwner, newOwner *types.Address) *OwnerChangedEvent {
	return &OwnerChangedEvent{
		Contract: contract,
		OldOwner: oldOwner,
		NewOwner: newOwner,
	}
}

// PoolCreatedEvent mirrors IUniswapV3Factory's PoolCreated event: a pool for (Token0, Token1, Fee) was deployed at Pool.
type PoolCreatedEvent struct {
	Contract    *types.Address
	Token0      *types.Address
	Token1      *types.Address
	Fee         *big.Int
	TickSpacing *big.Int
	Pool        *types.Address
}

func NewPoolCreatedEvent(contract, token0, token1 *types.Address, fee, tickSpacing *big.Int, pool *types.Address) *PoolCreatedEvent {
	return &PoolCreatedEvent{
		Contract:    contract,
		Token0:      token0,
		Token1:      token1,
		Fee:         fee,
		TickSpacing: tickSpacing,
		Pool:        pool,
	}
}

// FeeAmountEnabledEvent mirrors IUniswapV3Factory's FeeAmountEnabled event: Fee became a valid tier with TickSpacing.
type FeeAmountEnabledEvent struct {
	Contract    *types.Address
	Fee         *big.Int
	TickSpacing *big.Int
}

func NewFeeAmountEnabledEvent(contract *types.Address, fee, tickSpacing *big.Int) *FeeAmountEnabledEvent {
	return &FeeAmountEnabledEvent{
		Contract:    contract,
		Fee:         fee,
		TickSpacing: tickSpacing,
	}
}
