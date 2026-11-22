package v3

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// IPoolImmutables mirrors IUniswapV3PoolImmutables.
type IPoolImmutables interface {
	Factory(ctx context.Context, opts *bind.CallOptions, req *FactoryRequest) (*types.Address, error)
	Token0(ctx context.Context, opts *bind.CallOptions, req *Token0Request) (*types.Address, error)
	Token1(ctx context.Context, opts *bind.CallOptions, req *Token1Request) (*types.Address, error)
	Fee(ctx context.Context, opts *bind.CallOptions, req *FeeRequest) (*big.Int, error)
	TickSpacing(ctx context.Context, opts *bind.CallOptions, req *TickSpacingRequest) (*big.Int, error)
	MaxLiquidityPerTick(ctx context.Context, opts *bind.CallOptions, req *MaxLiquidityPerTickRequest) (*big.Int, error)
}

// IPoolState mirrors IUniswapV3PoolState.
type IPoolState interface {
	Slot0(ctx context.Context, opts *bind.CallOptions, req *Slot0Request) (*Slot0, error)
	FeeGrowthGlobal0X128(ctx context.Context, opts *bind.CallOptions, req *FeeGrowthGlobal0X128Request) (*big.Int, error)
	FeeGrowthGlobal1X128(ctx context.Context, opts *bind.CallOptions, req *FeeGrowthGlobal1X128Request) (*big.Int, error)
	ProtocolFees(ctx context.Context, opts *bind.CallOptions, req *ProtocolFeesRequest) (*ProtocolFees, error)
	Liquidity(ctx context.Context, opts *bind.CallOptions, req *LiquidityRequest) (*big.Int, error)
	Ticks(ctx context.Context, opts *bind.CallOptions, req *TicksRequest) (*TickInfo, error)
	TickBitmap(ctx context.Context, opts *bind.CallOptions, req *TickBitmapRequest) (*big.Int, error)
	Positions(ctx context.Context, opts *bind.CallOptions, req *PositionsRequest) (*PositionInfo, error)
	Observations(ctx context.Context, opts *bind.CallOptions, req *ObservationsRequest) (*ObservationInfo, error)
}

// IPoolDerivedState mirrors IUniswapV3PoolDerivedState.
type IPoolDerivedState interface {
	Observe(ctx context.Context, opts *bind.CallOptions, req *ObserveRequest) (*ObserveResult, error)
	SnapshotCumulativesInside(ctx context.Context, opts *bind.CallOptions, req *SnapshotCumulativesInsideRequest) (*SnapshotCumulativesInside, error)
}

// IPoolActions mirrors IUniswapV3PoolActions.
type IPoolActions interface {
	CallMint(ctx context.Context, opts *bind.CallOptions, req *PoolMintRequest) (amount0, amount1 *big.Int, err error)
	CallCollect(ctx context.Context, opts *bind.CallOptions, req *PoolCollectRequest) (amount0, amount1 *big.Int, err error)
	CallBurn(ctx context.Context, opts *bind.CallOptions, req *BurnRequest) (amount0, amount1 *big.Int, err error)
	CallSwap(ctx context.Context, opts *bind.CallOptions, req *SwapRequest) (amount0, amount1 *big.Int, err error)

	Initialize(ctx context.Context, opts *bind.TransactOptions, req *InitializeRequest) (core.Transaction, *types.Hash, error)
	Mint(ctx context.Context, opts *bind.TransactOptions, req *PoolMintRequest) (core.Transaction, *types.Hash, error)
	Collect(ctx context.Context, opts *bind.TransactOptions, req *PoolCollectRequest) (core.Transaction, *types.Hash, error)
	Burn(ctx context.Context, opts *bind.TransactOptions, req *BurnRequest) (core.Transaction, *types.Hash, error)
	Swap(ctx context.Context, opts *bind.TransactOptions, req *SwapRequest) (core.Transaction, *types.Hash, error)
	Flash(ctx context.Context, opts *bind.TransactOptions, req *FlashRequest) (core.Transaction, *types.Hash, error)
	IncreaseObservationCardinalityNext(ctx context.Context, opts *bind.TransactOptions, req *IncreaseObservationCardinalityNextRequest) (core.Transaction, *types.Hash, error)
}

// IPoolOwnerActions mirrors IUniswapV3PoolOwnerActions.
type IPoolOwnerActions interface {
	CallCollectProtocol(ctx context.Context, opts *bind.CallOptions, req *CollectProtocolRequest) (amount0, amount1 *big.Int, err error)

	SetFeeProtocol(ctx context.Context, opts *bind.TransactOptions, req *SetFeeProtocolRequest) (core.Transaction, *types.Hash, error)
	CollectProtocol(ctx context.Context, opts *bind.TransactOptions, req *CollectProtocolRequest) (core.Transaction, *types.Hash, error)
}

// IPool mirrors IUniswapV3Pool, composed the same way the Solidity interface is.
type IPool interface {
	IPoolImmutables
	IPoolState
	IPoolDerivedState
	IPoolActions
	IPoolOwnerActions
}

var _ IPool = (*Pool)(nil)

// Pool calls any UniswapV3Pool contract over a Client; it holds no contract address, so one instance serves every pool.
type Pool struct {
	c rpc.Client
}

// NewPool returns a Pool calling pools via cli.
func NewPool(cli rpc.Client) *Pool {
	return &Pool{
		c: cli,
	}
}

// Factory returns the UniswapV3Factory that deployed req.Contract.
func (pl *Pool) Factory(ctx context.Context, opts *bind.CallOptions, req *FactoryRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: factory: %w", err)
	}

	data, err := bind.Call(ctx, pl.c, opts, req.Contract, EncodeFactory())
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: factory: %w", pl.wrapRevertError(err))
	}

	out, err := DecodeFactory(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: factory: %w", err)
	}

	return out, nil
}

// Token0 returns req.Contract's token0 (the lower-sorted token).
func (pl *Pool) Token0(ctx context.Context, opts *bind.CallOptions, req *Token0Request) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: token0: %w", err)
	}

	data, err := bind.Call(ctx, pl.c, opts, req.Contract, EncodeToken0())
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: token0: %w", pl.wrapRevertError(err))
	}

	out, err := DecodeToken0(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: token0: %w", err)
	}

	return out, nil
}

// Token1 returns req.Contract's token1 (the higher-sorted token).
func (pl *Pool) Token1(ctx context.Context, opts *bind.CallOptions, req *Token1Request) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: token1: %w", err)
	}

	data, err := bind.Call(ctx, pl.c, opts, req.Contract, EncodeToken1())
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: token1: %w", pl.wrapRevertError(err))
	}

	out, err := DecodeToken1(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: token1: %w", err)
	}

	return out, nil
}

// Fee returns req.Contract's fee in hundredths of a bip.
func (pl *Pool) Fee(ctx context.Context, opts *bind.CallOptions, req *FeeRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: fee: %w", err)
	}

	data, err := bind.Call(ctx, pl.c, opts, req.Contract, EncodePoolFee())
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: fee: %w", pl.wrapRevertError(err))
	}

	out, err := DecodePoolFee(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: fee: %w", err)
	}

	return out, nil
}

// TickSpacing returns req.Contract's tick spacing.
func (pl *Pool) TickSpacing(ctx context.Context, opts *bind.CallOptions, req *TickSpacingRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: tickSpacing: %w", err)
	}

	data, err := bind.Call(ctx, pl.c, opts, req.Contract, EncodePoolTickSpacing())
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: tickSpacing: %w", pl.wrapRevertError(err))
	}

	out, err := DecodePoolTickSpacing(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: tickSpacing: %w", err)
	}

	return out, nil
}

// MaxLiquidityPerTick returns the most liquidity any tick of req.Contract can hold.
func (pl *Pool) MaxLiquidityPerTick(ctx context.Context, opts *bind.CallOptions, req *MaxLiquidityPerTickRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: maxLiquidityPerTick: %w", err)
	}

	data, err := bind.Call(ctx, pl.c, opts, req.Contract, EncodeMaxLiquidityPerTick())
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: maxLiquidityPerTick: %w", pl.wrapRevertError(err))
	}

	out, err := DecodeMaxLiquidityPerTick(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: maxLiquidityPerTick: %w", err)
	}

	return out, nil
}

// Slot0 returns req.Contract's current price, tick and oracle/fee-protocol state.
func (pl *Pool) Slot0(ctx context.Context, opts *bind.CallOptions, req *Slot0Request) (*Slot0, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: slot0: %w", err)
	}

	data, err := bind.Call(ctx, pl.c, opts, req.Contract, EncodeSlot0())
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: slot0: %w", pl.wrapRevertError(err))
	}

	out, err := DecodeSlot0(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: slot0: %w", err)
	}

	return out, nil
}

// FeeGrowthGlobal0X128 returns req.Contract's all-time token0 fees per unit of liquidity, as Q128.
func (pl *Pool) FeeGrowthGlobal0X128(ctx context.Context, opts *bind.CallOptions, req *FeeGrowthGlobal0X128Request) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: feeGrowthGlobal0X128: %w", err)
	}

	data, err := bind.Call(ctx, pl.c, opts, req.Contract, EncodeFeeGrowthGlobal0X128())
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: feeGrowthGlobal0X128: %w", pl.wrapRevertError(err))
	}

	out, err := DecodeFeeGrowthGlobal0X128(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: feeGrowthGlobal0X128: %w", err)
	}

	return out, nil
}

// FeeGrowthGlobal1X128 returns req.Contract's all-time token1 fees per unit of liquidity, as Q128.
func (pl *Pool) FeeGrowthGlobal1X128(ctx context.Context, opts *bind.CallOptions, req *FeeGrowthGlobal1X128Request) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: feeGrowthGlobal1X128: %w", err)
	}

	data, err := bind.Call(ctx, pl.c, opts, req.Contract, EncodeFeeGrowthGlobal1X128())
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: feeGrowthGlobal1X128: %w", pl.wrapRevertError(err))
	}

	out, err := DecodeFeeGrowthGlobal1X128(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: feeGrowthGlobal1X128: %w", err)
	}

	return out, nil
}

// ProtocolFees returns req.Contract's uncollected protocol fees.
func (pl *Pool) ProtocolFees(ctx context.Context, opts *bind.CallOptions, req *ProtocolFeesRequest) (*ProtocolFees, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: protocolFees: %w", err)
	}

	data, err := bind.Call(ctx, pl.c, opts, req.Contract, EncodeProtocolFees())
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: protocolFees: %w", pl.wrapRevertError(err))
	}

	out, err := DecodeProtocolFees(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: protocolFees: %w", err)
	}

	return out, nil
}

// Liquidity returns req.Contract's in-range liquidity.
func (pl *Pool) Liquidity(ctx context.Context, opts *bind.CallOptions, req *LiquidityRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: liquidity: %w", err)
	}

	data, err := bind.Call(ctx, pl.c, opts, req.Contract, EncodePoolLiquidity())
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: liquidity: %w", pl.wrapRevertError(err))
	}

	out, err := DecodePoolLiquidity(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: liquidity: %w", err)
	}

	return out, nil
}

// Ticks returns req.Contract's state for req.Tick.
func (pl *Pool) Ticks(ctx context.Context, opts *bind.CallOptions, req *TicksRequest) (*TickInfo, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: ticks: %w", err)
	}

	calldata, err := EncodeTicks(req.Tick)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: ticks: %w", err)
	}

	data, err := bind.Call(ctx, pl.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: ticks: %w", pl.wrapRevertError(err))
	}

	out, err := DecodeTicks(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: ticks: %w", err)
	}

	return out, nil
}

// TickBitmap returns req.Contract's initialized-tick bitmap word at req.WordPosition.
func (pl *Pool) TickBitmap(ctx context.Context, opts *bind.CallOptions, req *TickBitmapRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: tickBitmap: %w", err)
	}

	calldata, err := EncodeTickBitmap(req.WordPosition)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: tickBitmap: %w", err)
	}

	data, err := bind.Call(ctx, pl.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: tickBitmap: %w", pl.wrapRevertError(err))
	}

	out, err := DecodeTickBitmap(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: tickBitmap: %w", err)
	}

	return out, nil
}

// Positions returns req.Contract's position for req.Key, keccak256(owner, tickLower, tickUpper).
func (pl *Pool) Positions(ctx context.Context, opts *bind.CallOptions, req *PositionsRequest) (*PositionInfo, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: positions: %w", err)
	}

	calldata, err := EncodePositions(req.Key)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: positions: %w", err)
	}

	data, err := bind.Call(ctx, pl.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: positions: %w", pl.wrapRevertError(err))
	}

	out, err := DecodePositions(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: positions: %w", err)
	}

	return out, nil
}

// Observations returns req.Contract's oracle observation at req.Index.
func (pl *Pool) Observations(ctx context.Context, opts *bind.CallOptions, req *ObservationsRequest) (*ObservationInfo, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: observations: %w", err)
	}

	calldata, err := EncodePoolObservations(req.Index)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: observations: %w", err)
	}

	data, err := bind.Call(ctx, pl.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: observations: %w", pl.wrapRevertError(err))
	}

	out, err := DecodePoolObservations(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: observations: %w", err)
	}

	return out, nil
}

// Observe returns req.Contract's cumulative tick and seconds-per-liquidity as of each of req.SecondsAgos.
func (pl *Pool) Observe(ctx context.Context, opts *bind.CallOptions, req *ObserveRequest) (*ObserveResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: observe: %w", err)
	}

	calldata, err := EncodeObserve(req.SecondsAgos)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: observe: %w", err)
	}

	data, err := bind.Call(ctx, pl.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: observe: %w", pl.wrapRevertError(err))
	}

	out, err := DecodeObserve(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: observe: %w", err)
	}

	return out, nil
}

// SnapshotCumulativesInside returns req.Contract's cumulatives inside [req.TickLower, req.TickUpper].
func (pl *Pool) SnapshotCumulativesInside(ctx context.Context, opts *bind.CallOptions, req *SnapshotCumulativesInsideRequest) (*SnapshotCumulativesInside, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: snapshotCumulativesInside: %w", err)
	}

	calldata, err := EncodeSnapshotCumulativesInside(req.TickLower, req.TickUpper)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: snapshotCumulativesInside: %w", err)
	}

	data, err := bind.Call(ctx, pl.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: snapshotCumulativesInside: %w", pl.wrapRevertError(err))
	}

	out, err := DecodeSnapshotCumulativesInside(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: snapshotCumulativesInside: %w", err)
	}

	return out, nil
}

// CallMint runs mint via eth_call as opts.From, returning (amount0, amount1) or the revert.
func (pl *Pool) CallMint(ctx context.Context, opts *bind.CallOptions, req *PoolMintRequest) (amount0, amount1 *big.Int, err error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callMint: %w", err)
	}

	calldata, err := EncodeMint(req.Recipient, req.TickLower, req.TickUpper, req.Amount, req.CallbackData)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callMint: %w", err)
	}

	data, err := bind.Call(ctx, pl.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callMint: %w", pl.wrapRevertError(err))
	}

	amount0, amount1, err = DecodeMint(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callMint: %w", err)
	}

	return amount0, amount1, nil
}

// CallCollect runs collect via eth_call as opts.From, returning (amount0, amount1) or the revert.
func (pl *Pool) CallCollect(ctx context.Context, opts *bind.CallOptions, req *PoolCollectRequest) (amount0, amount1 *big.Int, err error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callCollect: %w", err)
	}

	calldata, err := EncodeCollect(req.Recipient, req.TickLower, req.TickUpper, req.Amount0Requested, req.Amount1Requested)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callCollect: %w", err)
	}

	data, err := bind.Call(ctx, pl.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callCollect: %w", pl.wrapRevertError(err))
	}

	amount0, amount1, err = DecodeCollect(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callCollect: %w", err)
	}

	return amount0, amount1, nil
}

// CallBurn runs burn via eth_call as opts.From, returning (amount0, amount1) or the revert.
func (pl *Pool) CallBurn(ctx context.Context, opts *bind.CallOptions, req *BurnRequest) (amount0, amount1 *big.Int, err error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callBurn: %w", err)
	}

	calldata, err := EncodeBurn(req.TickLower, req.TickUpper, req.Amount)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callBurn: %w", err)
	}

	data, err := bind.Call(ctx, pl.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callBurn: %w", pl.wrapRevertError(err))
	}

	amount0, amount1, err = DecodeBurn(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callBurn: %w", err)
	}

	return amount0, amount1, nil
}

// CallSwap runs swap via eth_call as opts.From, returning (amount0, amount1) or the revert.
func (pl *Pool) CallSwap(ctx context.Context, opts *bind.CallOptions, req *SwapRequest) (amount0, amount1 *big.Int, err error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callSwap: %w", err)
	}

	calldata, err := EncodeSwap(req.Recipient, req.ZeroForOne, req.AmountSpecified, req.SqrtPriceLimitX96, req.CallbackData)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callSwap: %w", err)
	}

	data, err := bind.Call(ctx, pl.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callSwap: %w", pl.wrapRevertError(err))
	}

	amount0, amount1, err = DecodeSwap(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callSwap: %w", err)
	}

	return amount0, amount1, nil
}

// CallCollectProtocol runs collectProtocol via eth_call as opts.From, returning (amount0, amount1) or the revert.
func (pl *Pool) CallCollectProtocol(ctx context.Context, opts *bind.CallOptions, req *CollectProtocolRequest) (amount0, amount1 *big.Int, err error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callCollectProtocol: %w", err)
	}

	calldata, err := EncodeCollectProtocol(req.Recipient, req.Amount0Requested, req.Amount1Requested)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callCollectProtocol: %w", err)
	}

	data, err := bind.Call(ctx, pl.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callCollectProtocol: %w", pl.wrapRevertError(err))
	}

	amount0, amount1, err = DecodeCollectProtocol(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callCollectProtocol: %w", err)
	}

	return amount0, amount1, nil
}

// Initialize sets req.Contract's starting price; it can be called once, by anyone, signed by opts.Signer.
func (pl *Pool) Initialize(ctx context.Context, opts *bind.TransactOptions, req *InitializeRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: initialize: %w", err)
	}

	calldata, err := EncodeInitialize(req.SqrtPriceX96)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: initialize: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, pl.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: initialize: %w", pl.wrapRevertError(err))
	}

	return signed, hash, nil
}

// Mint adds req.Amount liquidity for req.Recipient; the sender must pay through uniswapV3MintCallback, signed by opts.Signer.
func (pl *Pool) Mint(ctx context.Context, opts *bind.TransactOptions, req *PoolMintRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: mint: %w", err)
	}

	calldata, err := EncodeMint(req.Recipient, req.TickLower, req.TickUpper, req.Amount, req.CallbackData)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: mint: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, pl.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: mint: %w", pl.wrapRevertError(err))
	}

	return signed, hash, nil
}

// Collect collects tokens owed to the sender's position, up to the requested amounts, signed by opts.Signer.
func (pl *Pool) Collect(ctx context.Context, opts *bind.TransactOptions, req *PoolCollectRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: collect: %w", err)
	}

	calldata, err := EncodeCollect(req.Recipient, req.TickLower, req.TickUpper, req.Amount0Requested, req.Amount1Requested)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: collect: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, pl.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: collect: %w", pl.wrapRevertError(err))
	}

	return signed, hash, nil
}

// Burn removes req.Amount liquidity from the sender's position, crediting it as tokens owed, signed by opts.Signer.
func (pl *Pool) Burn(ctx context.Context, opts *bind.TransactOptions, req *BurnRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: burn: %w", err)
	}

	calldata, err := EncodeBurn(req.TickLower, req.TickUpper, req.Amount)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: burn: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, pl.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: burn: %w", pl.wrapRevertError(err))
	}

	return signed, hash, nil
}

// Swap swaps on req.Contract; the sender must pay through uniswapV3SwapCallback, signed by opts.Signer.
func (pl *Pool) Swap(ctx context.Context, opts *bind.TransactOptions, req *SwapRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: swap: %w", err)
	}

	calldata, err := EncodeSwap(req.Recipient, req.ZeroForOne, req.AmountSpecified, req.SqrtPriceLimitX96, req.CallbackData)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: swap: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, pl.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: swap: %w", pl.wrapRevertError(err))
	}

	return signed, hash, nil
}

// Flash lends tokens to req.Recipient; the sender must repay plus fee through uniswapV3FlashCallback, signed by opts.Signer.
func (pl *Pool) Flash(ctx context.Context, opts *bind.TransactOptions, req *FlashRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: flash: %w", err)
	}

	calldata, err := EncodeFlash(req.Recipient, req.Amount0, req.Amount1, req.CallbackData)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: flash: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, pl.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: flash: %w", pl.wrapRevertError(err))
	}

	return signed, hash, nil
}

// IncreaseObservationCardinalityNext grows req.Contract's oracle buffer to at least req.ObservationCardinalityNext, signed by opts.Signer.
func (pl *Pool) IncreaseObservationCardinalityNext(ctx context.Context, opts *bind.TransactOptions, req *IncreaseObservationCardinalityNextRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: increaseObservationCardinalityNext: %w", err)
	}

	calldata, err := EncodeIncreaseObservationCardinalityNext(req.ObservationCardinalityNext)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: increaseObservationCardinalityNext: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, pl.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: increaseObservationCardinalityNext: %w", pl.wrapRevertError(err))
	}

	return signed, hash, nil
}

// SetFeeProtocol sets req.Contract's protocol fee share; only the factory owner may call it, signed by opts.Signer.
func (pl *Pool) SetFeeProtocol(ctx context.Context, opts *bind.TransactOptions, req *SetFeeProtocolRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: setFeeProtocol: %w", err)
	}

	calldata, err := EncodeSetFeeProtocol(req.FeeProtocol0, req.FeeProtocol1)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: setFeeProtocol: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, pl.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: setFeeProtocol: %w", pl.wrapRevertError(err))
	}

	return signed, hash, nil
}

// CollectProtocol collects req.Contract's protocol fees; only the factory owner may call it, signed by opts.Signer.
func (pl *Pool) CollectProtocol(ctx context.Context, opts *bind.TransactOptions, req *CollectProtocolRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: collectProtocol: %w", err)
	}

	calldata, err := EncodeCollectProtocol(req.Recipient, req.Amount0Requested, req.Amount1Requested)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: collectProtocol: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, pl.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: collectProtocol: %w", pl.wrapRevertError(err))
	}

	return signed, hash, nil
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
