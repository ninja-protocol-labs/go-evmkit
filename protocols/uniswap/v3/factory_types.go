package v3

import (
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// GetPoolParams is GetPool's argument list.
type GetPoolParams struct {
	Factory *types.Address
	TokenA  *types.Address
	TokenB  *types.Address
	Fee     *big.Int
}

func NewGetPoolParams(factory, tokenA, tokenB *types.Address, fee *big.Int) *GetPoolParams {
	return &GetPoolParams{
		Factory: factory,
		TokenA:  tokenA,
		TokenB:  tokenB,
		Fee:     fee,
	}
}

// CreatePoolParams is CreateCreatePool's and CallCreatePool's argument list.
type CreatePoolParams struct {
	Factory *types.Address
	Sender  *types.Address
	TokenA  *types.Address
	TokenB  *types.Address
	Fee     *big.Int
}

func NewCreatePoolParams(factory, sender, tokenA, tokenB *types.Address, fee *big.Int) *CreatePoolParams {
	return &CreatePoolParams{
		Factory: factory,
		Sender:  sender,
		TokenA:  tokenA,
		TokenB:  tokenB,
		Fee:     fee,
	}
}

// SetOwnerParams is CreateSetOwner's argument list.
type SetOwnerParams struct {
	Factory *types.Address
	Sender  *types.Address
	Owner   *types.Address
}

func NewSetOwnerParams(factory, sender, owner *types.Address) *SetOwnerParams {
	return &SetOwnerParams{
		Factory: factory,
		Sender:  sender,
		Owner:   owner,
	}
}

// EnableFeeAmountParams is CreateEnableFeeAmount's argument list.
type EnableFeeAmountParams struct {
	Factory     *types.Address
	Sender      *types.Address
	Fee         *big.Int
	TickSpacing *big.Int
}

func NewEnableFeeAmountParams(factory, sender *types.Address, fee, tickSpacing *big.Int) *EnableFeeAmountParams {
	return &EnableFeeAmountParams{
		Factory:     factory,
		Sender:      sender,
		Fee:         fee,
		TickSpacing: tickSpacing,
	}
}

// OwnerChanged mirrors IUniswapV3Factory's OwnerChanged event: Contract's owner moved
// from OldOwner to NewOwner via setOwner.
type OwnerChanged struct {
	Contract *types.Address
	OldOwner *types.Address
	NewOwner *types.Address
}

func NewOwnerChanged(contract, oldOwner, newOwner *types.Address) *OwnerChanged {
	return &OwnerChanged{
		Contract: contract,
		OldOwner: oldOwner,
		NewOwner: newOwner,
	}
}

// PoolCreated mirrors IUniswapV3Factory's PoolCreated event: a new pool was deployed for
// (Token0, Token1, Fee) via createPool, now reachable at Pool.
type PoolCreated struct {
	Contract    *types.Address
	Token0      *types.Address
	Token1      *types.Address
	Fee         *big.Int
	TickSpacing *big.Int
	Pool        *types.Address
}

func NewPoolCreated(contract, token0, token1 *types.Address, fee, tickSpacing *big.Int, pool *types.Address) *PoolCreated {
	return &PoolCreated{
		Contract:    contract,
		Token0:      token0,
		Token1:      token1,
		Fee:         fee,
		TickSpacing: tickSpacing,
		Pool:        pool,
	}
}

// FeeAmountEnabled mirrors IUniswapV3Factory's FeeAmountEnabled event: Fee became a valid
// fee tier for createPool, with TickSpacing enforced for every pool created at that fee.
type FeeAmountEnabled struct {
	Contract    *types.Address
	Fee         *big.Int
	TickSpacing *big.Int
}

func NewFeeAmountEnabled(contract *types.Address, fee, tickSpacing *big.Int) *FeeAmountEnabled {
	return &FeeAmountEnabled{
		Contract:    contract,
		Fee:         fee,
		TickSpacing: tickSpacing,
	}
}
