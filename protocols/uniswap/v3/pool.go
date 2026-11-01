package v3

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// IPoolImmutables mirrors IUniswapV3PoolImmutables.
type IPoolImmutables interface {
	Factory(ctx context.Context, contract *types.Address, block string) (*types.Address, error)
	Token0(ctx context.Context, contract *types.Address, block string) (*types.Address, error)
	Token1(ctx context.Context, contract *types.Address, block string) (*types.Address, error)
	Fee(ctx context.Context, contract *types.Address, block string) (*big.Int, error)
	TickSpacing(ctx context.Context, contract *types.Address, block string) (*big.Int, error)
	MaxLiquidityPerTick(ctx context.Context, contract *types.Address, block string) (*big.Int, error)
}

// IPoolState mirrors IUniswapV3PoolState.
type IPoolState interface {
	Slot0(ctx context.Context, contract *types.Address, block string) (*Slot0, error)
	FeeGrowthGlobal0X128(ctx context.Context, contract *types.Address, block string) (*big.Int, error)
	FeeGrowthGlobal1X128(ctx context.Context, contract *types.Address, block string) (*big.Int, error)
	ProtocolFees(ctx context.Context, contract *types.Address, block string) (*ProtocolFees, error)
	Liquidity(ctx context.Context, contract *types.Address, block string) (*big.Int, error)
	Ticks(ctx context.Context, contract *types.Address, tick *big.Int, block string) (*TickInfo, error)
	TickBitmap(ctx context.Context, contract *types.Address, wordPosition int16, block string) (*big.Int, error)
	Positions(ctx context.Context, contract *types.Address, key [32]byte, block string) (*PositionInfo, error)
	Observations(ctx context.Context, contract *types.Address, index *big.Int, block string) (*ObservationInfo, error)
}

// IPoolDerivedState mirrors IUniswapV3PoolDerivedState.
type IPoolDerivedState interface {
	Observe(ctx context.Context, contract *types.Address, secondsAgos []uint32, block string) (*ObserveResult, error)
	SnapshotCumulativesInside(ctx context.Context, contract *types.Address, tickLower, tickUpper *big.Int, block string) (*SnapshotCumulativesInside, error)
}

// IPoolActions mirrors IUniswapV3PoolActions. No tx.Builder-taking Create* constructors here --
// see CreateInitialize, CreateMint, CreateCollect, CreateBurn, CreateSwap, CreateFlash, and
// CreateIncreaseObservationCardinalityNext.
type IPoolActions interface {
	Initialize(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Mint(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Collect(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Burn(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Swap(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Flash(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	IncreaseObservationCardinalityNext(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

// IPoolOwnerActions mirrors IUniswapV3PoolOwnerActions. No tx.Builder-taking Create*
// constructors here -- see CreateSetFeeProtocol and CreateCollectProtocol.
type IPoolOwnerActions interface {
	SetFeeProtocol(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	CollectProtocol(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

// IPool mirrors IUniswapV3Pool, composed the same way the Solidity interface is: one small
// interface per concern (immutables/state/derived state/actions/owner actions).
type IPool interface {
	IPoolImmutables
	IPoolState
	IPoolDerivedState
	IPoolActions
	IPoolOwnerActions
}

var _ IPool = (*Pool)(nil)

// Pool calls any UniswapV3Pool contract over a Client; it holds no contract address, so one
// instance serves every pool on every network.
type Pool struct {
	c rpc.Client
}

// NewPool returns a Pool calling pools via c.
func NewPool(c rpc.Client) *Pool {
	return &Pool{
		c: c,
	}
}

// Factory returns the UniswapV3Factory that deployed contract.
func (p *Pool) Factory(ctx context.Context, contract *types.Address, block string) (*types.Address, error) {
	data, err := p.call(ctx, nil, contract, EncodeFactory(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: factory: %w", err)
	}
	addr, err := DecodeFactory(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: factory: %w", err)
	}
	return addr, nil
}

// Token0 returns contract's first token, sorted by address.
func (p *Pool) Token0(ctx context.Context, contract *types.Address, block string) (*types.Address, error) {
	data, err := p.call(ctx, nil, contract, EncodeToken0(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: token0: %w", err)
	}
	addr, err := DecodeToken0(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: token0: %w", err)
	}
	return addr, nil
}

// Token1 returns contract's second token, sorted by address.
func (p *Pool) Token1(ctx context.Context, contract *types.Address, block string) (*types.Address, error) {
	data, err := p.call(ctx, nil, contract, EncodeToken1(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: token1: %w", err)
	}
	addr, err := DecodeToken1(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: token1: %w", err)
	}
	return addr, nil
}

// Fee returns contract's swap fee, denominated in hundredths of a bip.
func (p *Pool) Fee(ctx context.Context, contract *types.Address, block string) (*big.Int, error) {
	data, err := p.call(ctx, nil, contract, EncodePoolFee(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: fee: %w", err)
	}
	n, err := DecodePoolFee(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: fee: %w", err)
	}
	return n, nil
}

// TickSpacing returns the minimum number of ticks between initialized ticks in contract.
func (p *Pool) TickSpacing(ctx context.Context, contract *types.Address, block string) (*big.Int, error) {
	data, err := p.call(ctx, nil, contract, EncodePoolTickSpacing(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: tickSpacing: %w", err)
	}
	n, err := DecodePoolTickSpacing(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: tickSpacing: %w", err)
	}
	return n, nil
}

// MaxLiquidityPerTick returns the max amount of liquidity any single tick in contract can hold.
func (p *Pool) MaxLiquidityPerTick(ctx context.Context, contract *types.Address, block string) (*big.Int, error) {
	data, err := p.call(ctx, nil, contract, EncodeMaxLiquidityPerTick(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: maxLiquidityPerTick: %w", err)
	}
	n, err := DecodeMaxLiquidityPerTick(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: maxLiquidityPerTick: %w", err)
	}
	return n, nil
}

// Slot0 returns contract's packed 0th storage slot (current price, tick, observation state,
// protocol fee, and reentrancy lock) in one call.
func (p *Pool) Slot0(ctx context.Context, contract *types.Address, block string) (*Slot0, error) {
	data, err := p.call(ctx, nil, contract, EncodeSlot0(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: slot0: %w", err)
	}
	s, err := DecodeSlot0(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: slot0: %w", err)
	}
	return s, nil
}

// FeeGrowthGlobal0X128 returns the Q128.128 fees of token0 collected per unit of liquidity over
// contract's entire lifetime.
func (p *Pool) FeeGrowthGlobal0X128(ctx context.Context, contract *types.Address, block string) (*big.Int, error) {
	data, err := p.call(ctx, nil, contract, EncodeFeeGrowthGlobal0X128(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: feeGrowthGlobal0X128: %w", err)
	}
	n, err := DecodeFeeGrowthGlobal0X128(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: feeGrowthGlobal0X128: %w", err)
	}
	return n, nil
}

// FeeGrowthGlobal1X128 returns the Q128.128 fees of token1 collected per unit of liquidity over
// contract's entire lifetime.
func (p *Pool) FeeGrowthGlobal1X128(ctx context.Context, contract *types.Address, block string) (*big.Int, error) {
	data, err := p.call(ctx, nil, contract, EncodeFeeGrowthGlobal1X128(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: feeGrowthGlobal1X128: %w", err)
	}
	n, err := DecodeFeeGrowthGlobal1X128(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: feeGrowthGlobal1X128: %w", err)
	}
	return n, nil
}

// ProtocolFees returns the amounts of token0/token1 currently owed to the protocol.
func (p *Pool) ProtocolFees(ctx context.Context, contract *types.Address, block string) (*ProtocolFees, error) {
	data, err := p.call(ctx, nil, contract, EncodeProtocolFees(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: protocolFees: %w", err)
	}
	pf, err := DecodeProtocolFees(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: protocolFees: %w", err)
	}
	return pf, nil
}

// Liquidity returns contract's currently in-range liquidity -- unrelated to total liquidity
// across every tick.
func (p *Pool) Liquidity(ctx context.Context, contract *types.Address, block string) (*big.Int, error) {
	data, err := p.call(ctx, nil, contract, EncodePoolLiquidity(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: liquidity: %w", err)
	}
	n, err := DecodePoolLiquidity(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: liquidity: %w", err)
	}
	return n, nil
}

// Ticks returns the state of tick in contract.
func (p *Pool) Ticks(ctx context.Context, contract *types.Address, tick *big.Int, block string) (*TickInfo, error) {
	calldata, err := EncodeTicks(tick)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: ticks: %w", err)
	}
	data, err := p.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: ticks: %w", err)
	}
	ti, err := DecodeTicks(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: ticks: %w", err)
	}
	return ti, nil
}

// TickBitmap returns the 256 packed tick-initialized flags for wordPosition in contract.
func (p *Pool) TickBitmap(ctx context.Context, contract *types.Address, wordPosition int16, block string) (*big.Int, error) {
	calldata, err := EncodeTickBitmap(wordPosition)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: tickBitmap: %w", err)
	}
	data, err := p.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: tickBitmap: %w", err)
	}
	n, err := DecodeTickBitmap(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: tickBitmap: %w", err)
	}
	return n, nil
}

// Positions returns the state of the position identified by key (hash of owner, tickLower,
// tickUpper) in contract.
func (p *Pool) Positions(ctx context.Context, contract *types.Address, key [32]byte, block string) (*PositionInfo, error) {
	calldata, err := EncodePositions(key)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: positions: %w", err)
	}
	data, err := p.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: positions: %w", err)
	}
	pi, err := DecodePositions(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: positions: %w", err)
	}
	return pi, nil
}

// Observations returns the oracle observation at index in contract.
func (p *Pool) Observations(ctx context.Context, contract *types.Address, index *big.Int, block string) (*ObservationInfo, error) {
	calldata, err := EncodePoolObservations(index)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: observations: %w", err)
	}
	data, err := p.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: observations: %w", err)
	}
	oi, err := DecodePoolObservations(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: observations: %w", err)
	}
	return oi, nil
}

// Observe returns contract's cumulative tick and liquidity as of each secondsAgos entry.
func (p *Pool) Observe(ctx context.Context, contract *types.Address, secondsAgos []uint32, block string) (*ObserveResult, error) {
	calldata, err := EncodeObserve(secondsAgos)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: observe: %w", err)
	}
	data, err := p.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: observe: %w", err)
	}
	or, err := DecodeObserve(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: observe: %w", err)
	}
	return or, nil
}

// SnapshotCumulativesInside returns a snapshot of the tick/seconds-per-liquidity accumulators
// for the [tickLower, tickUpper] range in contract.
func (p *Pool) SnapshotCumulativesInside(ctx context.Context, contract *types.Address, tickLower, tickUpper *big.Int, block string) (*SnapshotCumulativesInside, error) {
	calldata, err := EncodeSnapshotCumulativesInside(tickLower, tickUpper)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: snapshotCumulativesInside: %w", err)
	}
	data, err := p.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: snapshotCumulativesInside: %w", err)
	}
	sci, err := DecodeSnapshotCumulativesInside(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: snapshotCumulativesInside: %w", err)
	}
	return sci, nil
}

// CreateInitialize builds an unsigned initialize(sqrtPriceX96) transaction via b, sent by sender.
func (p *Pool) CreateInitialize(ctx context.Context, b tx.Builder, contract, sender *types.Address, sqrtPriceX96 *big.Int) (core.Transaction, error) {
	calldata, err := EncodeInitialize(sqrtPriceX96)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createInitialize: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createInitialize: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// Initialize broadcasts an already-signed initialize transaction.
func (p *Pool) Initialize(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: initialize: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: initialize: %w", err)
	}
	return hash, nil
}

// CreateMint builds an unsigned mint(recipient, tickLower, tickUpper, amount, data) transaction
// via b, sent by sender -- which must be a contract implementing IUniswapV3MintCallback, since
// the pool calls back into msg.sender mid-transaction to collect payment.
func (p *Pool) CreateMint(ctx context.Context, b tx.Builder, contract, sender, recipient *types.Address, tickLower, tickUpper, amount *big.Int, callbackData []byte) (core.Transaction, error) {
	calldata, err := EncodeMint(recipient, tickLower, tickUpper, amount, callbackData)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createMint: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createMint: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// CallMint simulates mint(recipient, tickLower, tickUpper, amount, data) via eth_call, as
// sender, returning the amounts of token0/token1 it would require -- without broadcasting
// anything. sender must already be a contract implementing IUniswapV3MintCallback on-chain, or
// the simulated callback is a no-op and the simulation reverts the same way a real EOA-sent
// mint would.
func (p *Pool) CallMint(ctx context.Context, contract, sender, recipient *types.Address, tickLower, tickUpper, amount *big.Int, callbackData []byte, block string) (amount0, amount1 *big.Int, err error) {
	calldata, err := EncodeMint(recipient, tickLower, tickUpper, amount, callbackData)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callMint: %w", err)
	}
	data, err := p.call(ctx, sender, contract, calldata, block)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callMint: %w", err)
	}
	amount0, amount1, err = DecodeMint(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callMint: %w", err)
	}
	return amount0, amount1, nil
}

// Mint broadcasts an already-signed mint transaction. The caller must be a contract implementing
// IUniswapV3MintCallback, since the pool calls back into msg.sender mid-transaction to collect
// payment -- a plain EOA-signed call reverts there.
func (p *Pool) Mint(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: mint: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: mint: %w", err)
	}
	return hash, nil
}

// CreateCollect builds an unsigned collect(recipient, tickLower, tickUpper, amount0Requested,
// amount1Requested) transaction via b, sent by sender -- which must be the position's owner.
func (p *Pool) CreateCollect(ctx context.Context, b tx.Builder, contract, sender, recipient *types.Address, tickLower, tickUpper, amount0Requested, amount1Requested *big.Int) (core.Transaction, error) {
	calldata, err := EncodeCollect(recipient, tickLower, tickUpper, amount0Requested, amount1Requested)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createCollect: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createCollect: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// CallCollect simulates collect(recipient, tickLower, tickUpper, amount0Requested,
// amount1Requested) via eth_call, as sender, returning the amounts of token0/token1 it would
// collect -- without broadcasting anything.
func (p *Pool) CallCollect(ctx context.Context, contract, sender, recipient *types.Address, tickLower, tickUpper, amount0Requested, amount1Requested *big.Int, block string) (amount0, amount1 *big.Int, err error) {
	calldata, err := EncodeCollect(recipient, tickLower, tickUpper, amount0Requested, amount1Requested)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callCollect: %w", err)
	}
	data, err := p.call(ctx, sender, contract, calldata, block)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callCollect: %w", err)
	}
	amount0, amount1, err = DecodeCollect(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callCollect: %w", err)
	}
	return amount0, amount1, nil
}

// Collect broadcasts an already-signed collect transaction.
func (p *Pool) Collect(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: collect: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: collect: %w", err)
	}
	return hash, nil
}

// CreateBurn builds an unsigned burn(tickLower, tickUpper, amount) transaction via b, sent by
// sender -- which must be the position's owner.
func (p *Pool) CreateBurn(ctx context.Context, b tx.Builder, contract, sender *types.Address, tickLower, tickUpper, amount *big.Int) (core.Transaction, error) {
	calldata, err := EncodeBurn(tickLower, tickUpper, amount)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createBurn: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createBurn: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// CallBurn simulates burn(tickLower, tickUpper, amount) via eth_call, as sender, returning the
// amounts of token0/token1 it would credit to the position's tokensOwed -- without broadcasting
// anything.
func (p *Pool) CallBurn(ctx context.Context, contract, sender *types.Address, tickLower, tickUpper, amount *big.Int, block string) (amount0, amount1 *big.Int, err error) {
	calldata, err := EncodeBurn(tickLower, tickUpper, amount)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callBurn: %w", err)
	}
	data, err := p.call(ctx, sender, contract, calldata, block)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callBurn: %w", err)
	}
	amount0, amount1, err = DecodeBurn(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callBurn: %w", err)
	}
	return amount0, amount1, nil
}

// Burn broadcasts an already-signed burn transaction.
func (p *Pool) Burn(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: burn: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: burn: %w", err)
	}
	return hash, nil
}

// CreateSwap builds an unsigned swap(recipient, zeroForOne, amountSpecified, sqrtPriceLimitX96,
// data) transaction via b, sent by sender -- which must be a contract implementing
// IUniswapV3SwapCallback, since the pool calls back into msg.sender mid-transaction to collect
// payment.
func (p *Pool) CreateSwap(ctx context.Context, b tx.Builder, contract, sender, recipient *types.Address, zeroForOne bool, amountSpecified, sqrtPriceLimitX96 *big.Int, callbackData []byte) (core.Transaction, error) {
	calldata, err := EncodeSwap(recipient, zeroForOne, amountSpecified, sqrtPriceLimitX96, callbackData)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createSwap: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createSwap: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// CallSwap simulates swap(recipient, zeroForOne, amountSpecified, sqrtPriceLimitX96, data) via
// eth_call, as sender, returning the pool's balance deltas for token0/token1 -- without
// broadcasting anything. sender must already be a contract implementing IUniswapV3SwapCallback
// on-chain, or the simulated callback is a no-op and the simulation reverts the same way a real
// EOA-sent swap would.
func (p *Pool) CallSwap(ctx context.Context, contract, sender, recipient *types.Address, zeroForOne bool, amountSpecified, sqrtPriceLimitX96 *big.Int, callbackData []byte, block string) (amount0, amount1 *big.Int, err error) {
	calldata, err := EncodeSwap(recipient, zeroForOne, amountSpecified, sqrtPriceLimitX96, callbackData)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callSwap: %w", err)
	}
	data, err := p.call(ctx, sender, contract, calldata, block)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callSwap: %w", err)
	}
	amount0, amount1, err = DecodeSwap(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callSwap: %w", err)
	}
	return amount0, amount1, nil
}

// Swap broadcasts an already-signed swap transaction. The caller must be a contract implementing
// IUniswapV3SwapCallback, since the pool calls back into msg.sender mid-transaction to collect
// payment -- a plain EOA-signed call reverts there.
func (p *Pool) Swap(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: swap: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: swap: %w", err)
	}
	return hash, nil
}

// CreateFlash builds an unsigned flash(recipient, amount0, amount1, data) transaction via b,
// sent by sender -- which must be a contract implementing IUniswapV3FlashCallback, since the
// pool calls back into msg.sender mid-transaction to collect repayment.
func (p *Pool) CreateFlash(ctx context.Context, b tx.Builder, contract, sender, recipient *types.Address, amount0, amount1 *big.Int, callbackData []byte) (core.Transaction, error) {
	calldata, err := EncodeFlash(recipient, amount0, amount1, callbackData)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createFlash: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createFlash: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// Flash broadcasts an already-signed flash transaction. The caller must be a contract
// implementing IUniswapV3FlashCallback, since the pool calls back into msg.sender mid-transaction
// to collect repayment -- a plain EOA-signed call reverts there.
func (p *Pool) Flash(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: flash: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: flash: %w", err)
	}
	return hash, nil
}

// CreateIncreaseObservationCardinalityNext builds an unsigned
// increaseObservationCardinalityNext(observationCardinalityNext) transaction via b, sent by sender.
func (p *Pool) CreateIncreaseObservationCardinalityNext(ctx context.Context, b tx.Builder, contract, sender *types.Address, observationCardinalityNext uint16) (core.Transaction, error) {
	calldata, err := EncodeIncreaseObservationCardinalityNext(observationCardinalityNext)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createIncreaseObservationCardinalityNext: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createIncreaseObservationCardinalityNext: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// IncreaseObservationCardinalityNext broadcasts an already-signed
// increaseObservationCardinalityNext transaction.
func (p *Pool) IncreaseObservationCardinalityNext(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: increaseObservationCardinalityNext: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: increaseObservationCardinalityNext: %w", err)
	}
	return hash, nil
}

// CreateSetFeeProtocol builds an unsigned setFeeProtocol(feeProtocol0, feeProtocol1) transaction
// via b, sent by sender -- which must be contract's factory's owner, or the transaction reverts.
func (p *Pool) CreateSetFeeProtocol(ctx context.Context, b tx.Builder, contract, sender *types.Address, feeProtocol0, feeProtocol1 uint8) (core.Transaction, error) {
	calldata, err := EncodeSetFeeProtocol(feeProtocol0, feeProtocol1)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createSetFeeProtocol: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createSetFeeProtocol: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// SetFeeProtocol broadcasts an already-signed setFeeProtocol transaction -- sender must be
// contract's factory's owner, or the transaction reverts.
func (p *Pool) SetFeeProtocol(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: setFeeProtocol: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: setFeeProtocol: %w", err)
	}
	return hash, nil
}

// CreateCollectProtocol builds an unsigned collectProtocol(recipient, amount0Requested,
// amount1Requested) transaction via b, sent by sender -- which must be contract's factory's
// owner, or the transaction reverts.
func (p *Pool) CreateCollectProtocol(ctx context.Context, b tx.Builder, contract, sender, recipient *types.Address, amount0Requested, amount1Requested *big.Int) (core.Transaction, error) {
	calldata, err := EncodeCollectProtocol(recipient, amount0Requested, amount1Requested)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createCollectProtocol: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createCollectProtocol: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// CallCollectProtocol simulates collectProtocol(recipient, amount0Requested, amount1Requested)
// via eth_call, as sender, returning the protocol fee amounts it would collect -- without
// broadcasting anything.
func (p *Pool) CallCollectProtocol(ctx context.Context, contract, sender, recipient *types.Address, amount0Requested, amount1Requested *big.Int, block string) (amount0, amount1 *big.Int, err error) {
	calldata, err := EncodeCollectProtocol(recipient, amount0Requested, amount1Requested)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callCollectProtocol: %w", err)
	}
	data, err := p.call(ctx, sender, contract, calldata, block)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callCollectProtocol: %w", err)
	}
	amount0, amount1, err = DecodeCollectProtocol(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callCollectProtocol: %w", err)
	}
	return amount0, amount1, nil
}

// CollectProtocol broadcasts an already-signed collectProtocol transaction -- sender must be
// contract's factory's owner, or the transaction reverts.
func (p *Pool) CollectProtocol(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: collectProtocol: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: collectProtocol: %w", err)
	}
	return hash, nil
}

// call eth_calls contract with calldata (from msg.sender from, if non-nil), decoding a reverted
// require(cond, reason) to its sentinel.
func (p *Pool) call(ctx context.Context, from, contract *types.Address, calldata []byte, block string) ([]byte, error) {
	params := map[string]any{
		"to":   contract.String(),
		"data": encoding.Hex.EncodePrefixed(calldata),
	}
	if from != nil {
		params["from"] = from.String()
	}

	var raw string
	if err := p.c.Call(ctx, rpc.ETHCall("call", params, block, &raw)); err != nil {
		return nil, p.wrapRevertError(err)
	}
	return encoding.Hex.Decode(raw)
}

// wrapRevertError maps a known UniswapV3Pool require(cond, reason) string in err to its
// sentinel. Many of Pool's requires carry no reason string at all (the onlyFactoryOwner check,
// a couple of internal invariants); those fall back to ErrReverted, same as UniswapV3Factory.
func (_ *Pool) wrapRevertError(err error) error {
	var rpcErr *rpc.ResponseError
	if !errors.As(err, &rpcErr) {
		return err
	}
	if len(rpcErr.Data) == 0 {
		return fmt.Errorf("%w: %w", ErrReverted, err)
	}

	var hexData string
	if jsonErr := json.Unmarshal(rpcErr.Data, &hexData); jsonErr != nil {
		return fmt.Errorf("%w: %w", ErrReverted, err)
	}

	data, decErr := encoding.Hex.Decode(hexData)
	if decErr != nil {
		return fmt.Errorf("%w: %w", ErrReverted, err)
	}

	reason, decErr := abi.DecodeRevert(data)
	if decErr != nil {
		return fmt.Errorf("%w: %w", ErrReverted, err)
	}

	switch reason {
	case reasonLocked:
		return fmt.Errorf("%w: %w", ErrLocked, err)
	case reasonTickLowerGreaterThanOrEqualUpper:
		return fmt.Errorf("%w: %w", ErrTickLowerGreaterThanOrEqual, err)
	case reasonTickLowerTooLow:
		return fmt.Errorf("%w: %w", ErrTickLowerTooLow, err)
	case reasonTickUpperTooHigh:
		return fmt.Errorf("%w: %w", ErrTickUpperTooHigh, err)
	case reasonAlreadyInitialized:
		return fmt.Errorf("%w: %w", ErrAlreadyInitialized, err)
	case reasonMintToken0Underpaid:
		return fmt.Errorf("%w: %w", ErrMintToken0Underpaid, err)
	case reasonMintToken1Underpaid:
		return fmt.Errorf("%w: %w", ErrMintToken1Underpaid, err)
	case reasonAmountSpecifiedZero:
		return fmt.Errorf("%w: %w", ErrAmountSpecifiedZero, err)
	case reasonInvalidSqrtPriceLimit:
		return fmt.Errorf("%w: %w", ErrInvalidSqrtPriceLimit, err)
	case reasonSwapUnderpaid:
		return fmt.Errorf("%w: %w", ErrSwapUnderpaid, err)
	case reasonNoLiquidity:
		return fmt.Errorf("%w: %w", ErrNoLiquidity, err)
	case reasonFlashToken0Underpaid:
		return fmt.Errorf("%w: %w", ErrFlashToken0Underpaid, err)
	case reasonFlashToken1Underpaid:
		return fmt.Errorf("%w: %w", ErrFlashToken1Underpaid, err)
	case reasonLiquiditySubUnderflow:
		return fmt.Errorf("%w: %w", ErrLiquiditySubUnderflow, err)
	case reasonLiquidityAddOverflow:
		return fmt.Errorf("%w: %w", ErrLiquidityAddOverflow, err)
	case reasonOracleNotInitialized:
		return fmt.Errorf("%w: %w", ErrOracleNotInitialized, err)
	case reasonOracleObservationTooOld:
		return fmt.Errorf("%w: %w", ErrOracleObservationTooOld, err)
	case reasonNoPosition:
		return fmt.Errorf("%w: %w", ErrNoPosition, err)
	case reasonLiquidityOverflow:
		return fmt.Errorf("%w: %w", ErrLiquidityOverflow, err)
	case reasonTickOutOfBounds:
		return fmt.Errorf("%w: %w", ErrTickOutOfBounds, err)
	case reasonSqrtRatioOutOfBounds:
		return fmt.Errorf("%w: %w", ErrSqrtRatioOutOfBounds, err)
	case reasonTransferFailed:
		return fmt.Errorf("%w: %w", ErrTransferFailed, err)
	default:
		return fmt.Errorf("%w: %w", ErrReverted, err)
	}
}
