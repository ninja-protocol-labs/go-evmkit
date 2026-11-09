package v2

import (
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// InitializeParams is CreateInitialize's argument list.
type InitializeParams struct {
	Pair   *types.Address
	Sender *types.Address
	Token0 *types.Address
	Token1 *types.Address
}

func NewInitializeParams(pair, sender, token0, token1 *types.Address) *InitializeParams {
	return &InitializeParams{
		Pair:   pair,
		Sender: sender,
		Token0: token0,
		Token1: token1,
	}
}

// MintParams is CreateMint's and CallMint's argument list.
type MintParams struct {
	Pair   *types.Address
	Sender *types.Address
	To     *types.Address
}

func NewMintParams(pair, sender, to *types.Address) *MintParams {
	return &MintParams{
		Pair:   pair,
		Sender: sender,
		To:     to,
	}
}

// BurnParams is CreateBurn's and CallBurn's argument list.
type BurnParams struct {
	Pair   *types.Address
	Sender *types.Address
	To     *types.Address
}

func NewBurnParams(pair, sender, to *types.Address) *BurnParams {
	return &BurnParams{
		Pair:   pair,
		Sender: sender,
		To:     to,
	}
}

// SwapParams is CreateSwap's argument list.
type SwapParams struct {
	Pair       *types.Address
	Sender     *types.Address
	Amount0Out *big.Int
	Amount1Out *big.Int
	To         *types.Address
	Data       []byte
}

func NewSwapParams(pair, sender *types.Address, amount0Out, amount1Out *big.Int, to *types.Address, data []byte) *SwapParams {
	return &SwapParams{
		Pair:       pair,
		Sender:     sender,
		Amount0Out: amount0Out,
		Amount1Out: amount1Out,
		To:         to,
		Data:       data,
	}
}

// SkimParams is CreateSkim's argument list.
type SkimParams struct {
	Pair   *types.Address
	Sender *types.Address
	To     *types.Address
}

func NewSkimParams(pair, sender, to *types.Address) *SkimParams {
	return &SkimParams{
		Pair:   pair,
		Sender: sender,
		To:     to,
	}
}

// SyncParams is CreateSync's argument list.
type SyncParams struct {
	Pair   *types.Address
	Sender *types.Address
}

func NewSyncParams(pair, sender *types.Address) *SyncParams {
	return &SyncParams{
		Pair:   pair,
		Sender: sender,
	}
}

// Mint mirrors UniswapV2Pair's Mint event, emitted when mint(to) adds liquidity.
type Mint struct {
	Pair    *types.Address
	Sender  *types.Address
	Amount0 *big.Int
	Amount1 *big.Int
}

func NewMint(pair, sender *types.Address, amount0, amount1 *big.Int) *Mint {
	return &Mint{
		Pair:    pair,
		Sender:  sender,
		Amount0: amount0,
		Amount1: amount1,
	}
}

// Burn mirrors UniswapV2Pair's Burn event, emitted when burn(to) removes liquidity.
type Burn struct {
	Pair    *types.Address
	Sender  *types.Address
	Amount0 *big.Int
	Amount1 *big.Int
	To      *types.Address
}

func NewBurn(pair, sender *types.Address, amount0, amount1 *big.Int, to *types.Address) *Burn {
	return &Burn{
		Pair:    pair,
		Sender:  sender,
		Amount0: amount0,
		Amount1: amount1,
		To:      to,
	}
}

// Swap mirrors UniswapV2Pair's Swap event, emitted on every swap.
type Swap struct {
	Pair       *types.Address
	Sender     *types.Address
	Amount0In  *big.Int
	Amount1In  *big.Int
	Amount0Out *big.Int
	Amount1Out *big.Int
	To         *types.Address
}

func NewSwap(pair, sender *types.Address, amount0In, amount1In, amount0Out, amount1Out *big.Int, to *types.Address) *Swap {
	return &Swap{
		Pair:       pair,
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
	Pair     *types.Address
	Reserve0 *big.Int
	Reserve1 *big.Int
}

func NewSync(pair *types.Address, reserve0, reserve1 *big.Int) *Sync {
	return &Sync{
		Pair:     pair,
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
