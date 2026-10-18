package v2

import (
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Mint mirrors UniswapV2Pair's Mint event, emitted when mint(to) adds liquidity.
type Mint struct {
	Contract *types.Address
	Sender   *types.Address
	Amount0  *big.Int
	Amount1  *big.Int
}

func NewMint(contract, sender *types.Address, amount0, amount1 *big.Int) *Mint {
	return &Mint{
		Contract: contract,
		Sender:   sender,
		Amount0:  amount0,
		Amount1:  amount1,
	}
}

// Burn mirrors UniswapV2Pair's Burn event, emitted when burn(to) removes liquidity.
type Burn struct {
	Contract *types.Address
	Sender   *types.Address
	Amount0  *big.Int
	Amount1  *big.Int
	To       *types.Address
}

func NewBurn(contract, sender *types.Address, amount0, amount1 *big.Int, to *types.Address) *Burn {
	return &Burn{
		Contract: contract,
		Sender:   sender,
		Amount0:  amount0,
		Amount1:  amount1,
		To:       to,
	}
}

// Swap mirrors UniswapV2Pair's Swap event, emitted on every swap.
type Swap struct {
	Contract   *types.Address
	Sender     *types.Address
	Amount0In  *big.Int
	Amount1In  *big.Int
	Amount0Out *big.Int
	Amount1Out *big.Int
	To         *types.Address
}

func NewSwap(contract, sender *types.Address, amount0In, amount1In, amount0Out, amount1Out *big.Int, to *types.Address) *Swap {
	return &Swap{
		Contract:   contract,
		Sender:     sender,
		Amount0In:  amount0In,
		Amount1In:  amount1In,
		Amount0Out: amount0Out,
		Amount1Out: amount1Out,
		To:         to,
	}
}

// Sync mirrors UniswapV2Pair's Sync event, emitted whenever the pair's reserves are updated
// (mint, burn, swap, or sync itself).
type Sync struct {
	Contract *types.Address
	Reserve0 *big.Int
	Reserve1 *big.Int
}

func NewSync(contract *types.Address, reserve0, reserve1 *big.Int) *Sync {
	return &Sync{
		Contract: contract,
		Reserve0: reserve0,
		Reserve1: reserve1,
	}
}

// Reserves is getReserves()'s (reserve0, reserve1, blockTimestampLast) return value.
type Reserves struct {
	Reserve0           *big.Int
	Reserve1           *big.Int
	BlockTimestampLast uint32
}

func NewReserves(reserve0, reserve1 *big.Int, blockTimestampLast uint32) *Reserves {
	return &Reserves{
		Reserve0:           reserve0,
		Reserve1:           reserve1,
		BlockTimestampLast: blockTimestampLast,
	}
}
