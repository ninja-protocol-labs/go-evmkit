package v4

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

type (
	Currency = *types.Address
	PoolID   = [32]byte
)

// PoolKey mirrors the PoolKey struct in types/PoolKey.sol -- it identifies a pool the same way a
// Pair/Pool contract address did in v2/v3, since PoolManager is a singleton shared by every pool.
// Currency0/Currency1 are Currency (a *types.Address alias; the zero address means native ETH).
type PoolKey struct {
	Currency0   Currency
	Currency1   Currency
	Fee         *big.Int
	TickSpacing *big.Int
	Hooks       *types.Address
}

func NewPoolKey(currency0, currency1 Currency, fee, tickSpacing *big.Int, hooks *types.Address) *PoolKey {
	return &PoolKey{
		Currency0:   currency0,
		Currency1:   currency1,
		Fee:         fee,
		TickSpacing: tickSpacing,
		Hooks:       hooks,
	}
}

// PoolID returns k's PoolId: keccak256(abi.encode(k)). PoolKey's five fields are all static, so
// its ABI encoding is just those five words concatenated (0xa0 bytes), matching
// PoolIdLibrary.toId's own assembly.
func (k *PoolKey) PoolID() (PoolID, error) {
	encoded, err := abi.Pack(abi.Types{poolKeyTuple}, []any{k.Currency0, k.Currency1, k.Fee, k.TickSpacing, k.Hooks})
	if err != nil {
		return PoolID{}, fmt.Errorf("uniswapv4: poolKey.id: %w", err)
	}
	var id PoolID
	copy(id[:], core.Keccak256(encoded).Bytes())
	return id, nil
}

// BalanceDelta mirrors types/BalanceDelta.sol: a single int256 packing two int128s, upper bits
// Amount0 and lower bits Amount1.
type BalanceDelta struct {
	Amount0 *big.Int
	Amount1 *big.Int
}

func NewBalanceDelta(amount0, amount1 *big.Int) *BalanceDelta {
	return &BalanceDelta{
		Amount0: amount0,
		Amount1: amount1,
	}
}

// LiquidityParams mirrors IPoolManager's own ModifyLiquidityParams tuple (tickLower/tickUpper/
// liquidityDelta/salt).
type LiquidityParams struct {
	TickLower      *big.Int
	TickUpper      *big.Int
	LiquidityDelta *big.Int
	Salt           [32]byte
}

func NewLiquidityParams(tickLower, tickUpper, liquidityDelta *big.Int, salt [32]byte) *LiquidityParams {
	return &LiquidityParams{
		TickLower:      tickLower,
		TickUpper:      tickUpper,
		LiquidityDelta: liquidityDelta,
		Salt:           salt,
	}
}

// SwapAmountParams mirrors IPoolManager's own SwapParams tuple (zeroForOne/amountSpecified/
// sqrtPriceLimitX96). Named SwapAmountParams, not SwapParams, to leave SwapParams free for the
// eventual Create/Call-method call wrapper.
type SwapAmountParams struct {
	ZeroForOne        bool
	AmountSpecified   *big.Int
	SqrtPriceLimitX96 *big.Int
}

func NewSwapAmountParams(zeroForOne bool, amountSpecified, sqrtPriceLimitX96 *big.Int) *SwapAmountParams {
	return &SwapAmountParams{
		ZeroForOne:        zeroForOne,
		AmountSpecified:   amountSpecified,
		SqrtPriceLimitX96: sqrtPriceLimitX96,
	}
}

// Initialize mirrors IPoolManager's Initialize event: a new pool was initialized for Key at
// SqrtPriceX96, landing at the given Tick.
type Initialize struct {
	PoolManager  *types.Address
	Id           PoolID
	Currency0    Currency
	Currency1    Currency
	Fee          *big.Int
	TickSpacing  *big.Int
	Hooks        *types.Address
	SqrtPriceX96 *big.Int
	Tick         *big.Int
}

func NewInitialize(poolManager *types.Address, id PoolID, currency0, currency1 Currency, fee, tickSpacing *big.Int, hooks *types.Address, sqrtPriceX96, tick *big.Int) *Initialize {
	return &Initialize{
		PoolManager:  poolManager,
		Id:           id,
		Currency0:    currency0,
		Currency1:    currency1,
		Fee:          fee,
		TickSpacing:  tickSpacing,
		Hooks:        hooks,
		SqrtPriceX96: sqrtPriceX96,
		Tick:         tick,
	}
}

// ModifyLiquidity mirrors IPoolManager's ModifyLiquidity event.
type ModifyLiquidity struct {
	PoolManager    *types.Address
	Id             PoolID
	Sender         *types.Address
	TickLower      *big.Int
	TickUpper      *big.Int
	LiquidityDelta *big.Int
	Salt           [32]byte
}

func NewModifyLiquidity(poolManager *types.Address, id PoolID, sender *types.Address, tickLower, tickUpper, liquidityDelta *big.Int, salt [32]byte) *ModifyLiquidity {
	return &ModifyLiquidity{
		PoolManager:    poolManager,
		Id:             id,
		Sender:         sender,
		TickLower:      tickLower,
		TickUpper:      tickUpper,
		LiquidityDelta: liquidityDelta,
		Salt:           salt,
	}
}

// Swap mirrors IPoolManager's Swap event: Sender swapped in pool Id; Amount0/Amount1 are the
// pool's balance deltas, SqrtPriceX96/Liquidity/Tick/Fee are the pool's state right after.
type Swap struct {
	PoolManager  *types.Address
	Id           PoolID
	Sender       *types.Address
	Amount0      *big.Int
	Amount1      *big.Int
	SqrtPriceX96 *big.Int
	Liquidity    *big.Int
	Tick         *big.Int
	Fee          *big.Int
}

func NewSwap(poolManager *types.Address, id PoolID, sender *types.Address, amount0, amount1, sqrtPriceX96, liquidity, tick, fee *big.Int) *Swap {
	return &Swap{
		PoolManager:  poolManager,
		Id:           id,
		Sender:       sender,
		Amount0:      amount0,
		Amount1:      amount1,
		SqrtPriceX96: sqrtPriceX96,
		Liquidity:    liquidity,
		Tick:         tick,
		Fee:          fee,
	}
}

// Donate mirrors IPoolManager's Donate event.
type Donate struct {
	PoolManager *types.Address
	Id          PoolID
	Sender      *types.Address
	Amount0     *big.Int
	Amount1     *big.Int
}

func NewDonate(poolManager *types.Address, id PoolID, sender *types.Address, amount0, amount1 *big.Int) *Donate {
	return &Donate{
		PoolManager: poolManager,
		Id:          id,
		Sender:      sender,
		Amount0:     amount0,
		Amount1:     amount1,
	}
}

// ProtocolFeeControllerUpdated mirrors IProtocolFees' ProtocolFeeControllerUpdated event.
type ProtocolFeeControllerUpdated struct {
	PoolManager           *types.Address
	ProtocolFeeController *types.Address
}

func NewProtocolFeeControllerUpdated(poolManager, protocolFeeController *types.Address) *ProtocolFeeControllerUpdated {
	return &ProtocolFeeControllerUpdated{
		PoolManager:           poolManager,
		ProtocolFeeController: protocolFeeController,
	}
}

// ProtocolFeeUpdated mirrors IProtocolFees' ProtocolFeeUpdated event.
type ProtocolFeeUpdated struct {
	PoolManager *types.Address
	Id          PoolID
	ProtocolFee *big.Int
}

func NewProtocolFeeUpdated(poolManager *types.Address, id PoolID, protocolFee *big.Int) *ProtocolFeeUpdated {
	return &ProtocolFeeUpdated{
		PoolManager: poolManager,
		Id:          id,
		ProtocolFee: protocolFee,
	}
}
