package v4

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc721"
)

// IPositionManagerImmutables mirrors IImmutableState plus NativeWrapper's and Permit2Forwarder's immutable getters.
type IPositionManagerImmutables interface {
	PoolManager(ctx context.Context, opts *bind.CallOptions, req *PoolManagerRequest) (*types.Address, error)
	WETH9(ctx context.Context, opts *bind.CallOptions, req *WETH9Request) (*types.Address, error)
	Permit2(ctx context.Context, opts *bind.CallOptions, req *Permit2Request) (*types.Address, error)
}

// IPositionManagerActions mirrors IPositionManager's batched-action entry points and views.
type IPositionManagerActions interface {
	NextTokenID(ctx context.Context, opts *bind.CallOptions, req *NextTokenIDRequest) (*big.Int, error)
	GetPositionLiquidity(ctx context.Context, opts *bind.CallOptions, req *GetPositionLiquidityRequest) (*big.Int, error)
	GetPoolAndPositionInfo(ctx context.Context, opts *bind.CallOptions, req *GetPoolAndPositionInfoRequest) (*PoolKey, *PositionInfo, error)
	PositionInfo(ctx context.Context, opts *bind.CallOptions, req *PositionInfoRequest) (*PositionInfo, error)
	PoolKeys(ctx context.Context, opts *bind.CallOptions, req *PoolKeysRequest) (*PoolKey, error)

	ModifyLiquidities(ctx context.Context, opts *bind.TransactOptions, req *ModifyLiquiditiesRequest) (core.Transaction, *types.Hash, error)
	ModifyLiquiditiesWithoutUnlock(ctx context.Context, opts *bind.TransactOptions, req *ModifyLiquiditiesWithoutUnlockRequest) (core.Transaction, *types.Hash, error)
}

// IPositionManagerNotifier mirrors INotifier.
type IPositionManagerNotifier interface {
	Subscriber(ctx context.Context, opts *bind.CallOptions, req *SubscriberRequest) (*types.Address, error)
	UnsubscribeGasLimit(ctx context.Context, opts *bind.CallOptions, req *UnsubscribeGasLimitRequest) (*big.Int, error)

	Subscribe(ctx context.Context, opts *bind.TransactOptions, req *SubscribeRequest) (core.Transaction, *types.Hash, error)
	Unsubscribe(ctx context.Context, opts *bind.TransactOptions, req *UnsubscribeRequest) (core.Transaction, *types.Hash, error)
}

// IPositionManagerPoolInitializer mirrors IPoolInitializer_v4.
type IPositionManagerPoolInitializer interface {
	CallInitializePool(ctx context.Context, opts *bind.CallOptions, req *InitializePoolRequest) (*big.Int, error)

	InitializePool(ctx context.Context, opts *bind.TransactOptions, req *InitializePoolRequest) (core.Transaction, *types.Hash, error)
}

// IPositionManagerMulticall mirrors IMulticall_v4.
type IPositionManagerMulticall interface {
	CallMulticall(ctx context.Context, opts *bind.CallOptions, req *MulticallRequest) ([][]byte, error)

	Multicall(ctx context.Context, opts *bind.TransactOptions, req *MulticallRequest) (core.Transaction, *types.Hash, error)
}

// IPositionManagerUnorderedNonce mirrors IUnorderedNonce.
type IPositionManagerUnorderedNonce interface {
	Nonces(ctx context.Context, opts *bind.CallOptions, req *NoncesRequest) (*big.Int, error)

	RevokeNonce(ctx context.Context, opts *bind.TransactOptions, req *RevokeNonceRequest) (core.Transaction, *types.Hash, error)
}

// IPositionManagerPermit mirrors IERC721Permit_v4 plus IPermit2Forwarder.
type IPositionManagerPermit interface {
	CallPermit2Permit(ctx context.Context, opts *bind.CallOptions, req *Permit2PermitRequest) ([]byte, error)
	CallPermit2PermitBatch(ctx context.Context, opts *bind.CallOptions, req *Permit2PermitBatchRequest) ([]byte, error)

	ERC721Permit(ctx context.Context, opts *bind.TransactOptions, req *ERC721PermitRequest) (core.Transaction, *types.Hash, error)
	ERC721PermitForAll(ctx context.Context, opts *bind.TransactOptions, req *ERC721PermitForAllRequest) (core.Transaction, *types.Hash, error)
	Permit2Permit(ctx context.Context, opts *bind.TransactOptions, req *Permit2PermitRequest) (core.Transaction, *types.Hash, error)
	Permit2PermitBatch(ctx context.Context, opts *bind.TransactOptions, req *Permit2PermitBatchRequest) (core.Transaction, *types.Hash, error)
}

// IPositionManager mirrors PositionManager: one interface per Solidity base contract plus the ERC721 surface of its position NFTs.
type IPositionManager interface {
	erc721.IERC721Metadata
	IPositionManagerImmutables
	IPositionManagerActions
	IPositionManagerNotifier
	IPositionManagerPoolInitializer
	IPositionManagerMulticall
	IPositionManagerUnorderedNonce
	IPositionManagerPermit
}

var _ IPositionManager = (*PositionManager)(nil)

// PositionManager calls any PositionManager deployment over a Client; ERC721 methods are promoted from the embedded *erc721.ERC721.
type PositionManager struct {
	*erc721.ERC721

	c rpc.Client
}

// NewPositionManager returns a PositionManager calling deployments via cli.
func NewPositionManager(cli rpc.Client) *PositionManager {
	return &PositionManager{
		ERC721: erc721.New(cli),
		c:      cli,
	}
}

// PoolManager returns the PoolManager req.Contract was deployed against.
func (p *PositionManager) PoolManager(ctx context.Context, opts *bind.CallOptions, req *PoolManagerRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: poolManager: %w", err)
	}

	calldata, err := EncodePoolManager()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: poolManager: %w", err)
	}

	data, err := bind.Call(ctx, p.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: poolManager: %w", mapPositionManagerRevertError(err))
	}

	out, err := DecodePoolManager(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: poolManager: %w", err)
	}

	return out, nil
}

// WETH9 returns the WETH9 req.Contract wraps and unwraps native ETH through.
func (p *PositionManager) WETH9(ctx context.Context, opts *bind.CallOptions, req *WETH9Request) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: weth9: %w", err)
	}

	calldata, err := EncodeWETH9()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: weth9: %w", err)
	}

	data, err := bind.Call(ctx, p.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: weth9: %w", mapPositionManagerRevertError(err))
	}

	out, err := DecodeWETH9(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: weth9: %w", err)
	}

	return out, nil
}

// Permit2 returns the Permit2 req.Contract forwards permit and permitBatch to.
func (p *PositionManager) Permit2(ctx context.Context, opts *bind.CallOptions, req *Permit2Request) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: permit2: %w", err)
	}

	calldata, err := EncodePermit2Contract()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: permit2: %w", err)
	}

	data, err := bind.Call(ctx, p.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: permit2: %w", mapPositionManagerRevertError(err))
	}

	out, err := DecodePermit2Contract(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: permit2: %w", err)
	}

	return out, nil
}

// NextTokenID returns the token ID req.Contract will assign to its next minted position.
func (p *PositionManager) NextTokenID(ctx context.Context, opts *bind.CallOptions, req *NextTokenIDRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: nextTokenId: %w", err)
	}

	calldata, err := EncodeNextTokenID()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: nextTokenId: %w", err)
	}

	data, err := bind.Call(ctx, p.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: nextTokenId: %w", mapPositionManagerRevertError(err))
	}

	out, err := DecodeNextTokenID(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: nextTokenId: %w", err)
	}

	return out, nil
}

// GetPositionLiquidity returns the liquidity of position req.TokenID.
func (p *PositionManager) GetPositionLiquidity(ctx context.Context, opts *bind.CallOptions, req *GetPositionLiquidityRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: getPositionLiquidity: %w", err)
	}

	calldata, err := EncodeGetPositionLiquidity(req.TokenID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getPositionLiquidity: %w", err)
	}

	data, err := bind.Call(ctx, p.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getPositionLiquidity: %w", mapPositionManagerRevertError(err))
	}

	out, err := DecodeGetPositionLiquidity(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getPositionLiquidity: %w", err)
	}

	return out, nil
}

// GetPoolAndPositionInfo returns the pool key and packed position info of position req.TokenID.
func (p *PositionManager) GetPoolAndPositionInfo(ctx context.Context, opts *bind.CallOptions, req *GetPoolAndPositionInfoRequest) (*PoolKey, *PositionInfo, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: getPoolAndPositionInfo: %w", err)
	}

	calldata, err := EncodeGetPoolAndPositionInfo(req.TokenID)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: getPoolAndPositionInfo: %w", err)
	}

	data, err := bind.Call(ctx, p.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: getPoolAndPositionInfo: %w", mapPositionManagerRevertError(err))
	}

	key, info, err := DecodeGetPoolAndPositionInfo(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: getPoolAndPositionInfo: %w", err)
	}

	return key, info, nil
}

// PositionInfo returns the packed position info of position req.TokenID.
func (p *PositionManager) PositionInfo(ctx context.Context, opts *bind.CallOptions, req *PositionInfoRequest) (*PositionInfo, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: positionInfo: %w", err)
	}

	calldata, err := EncodePositionInfo(req.TokenID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: positionInfo: %w", err)
	}

	data, err := bind.Call(ctx, p.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: positionInfo: %w", mapPositionManagerRevertError(err))
	}

	out, err := DecodePositionInfo(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: positionInfo: %w", err)
	}

	return out, nil
}

// PoolKeys returns the pool key req.Contract stored under the truncated 25-byte req.PoolID.
func (p *PositionManager) PoolKeys(ctx context.Context, opts *bind.CallOptions, req *PoolKeysRequest) (*PoolKey, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: poolKeys: %w", err)
	}

	calldata, err := EncodePoolKeys(req.PoolID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: poolKeys: %w", err)
	}

	data, err := bind.Call(ctx, p.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: poolKeys: %w", mapPositionManagerRevertError(err))
	}

	out, err := DecodePoolKeys(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: poolKeys: %w", err)
	}

	return out, nil
}

// Subscriber returns the subscriber of position req.TokenID, the zero address if none.
func (p *PositionManager) Subscriber(ctx context.Context, opts *bind.CallOptions, req *SubscriberRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: subscriber: %w", err)
	}

	calldata, err := EncodeSubscriber(req.TokenID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: subscriber: %w", err)
	}

	data, err := bind.Call(ctx, p.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: subscriber: %w", mapPositionManagerRevertError(err))
	}

	out, err := DecodeSubscriber(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: subscriber: %w", err)
	}

	return out, nil
}

// UnsubscribeGasLimit returns the gas req.Contract forwards to a subscriber's notifyUnsubscribe.
func (p *PositionManager) UnsubscribeGasLimit(ctx context.Context, opts *bind.CallOptions, req *UnsubscribeGasLimitRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: unsubscribeGasLimit: %w", err)
	}

	calldata, err := EncodeUnsubscribeGasLimit()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: unsubscribeGasLimit: %w", err)
	}

	data, err := bind.Call(ctx, p.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: unsubscribeGasLimit: %w", mapPositionManagerRevertError(err))
	}

	out, err := DecodeUnsubscribeGasLimit(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: unsubscribeGasLimit: %w", err)
	}

	return out, nil
}

// Nonces returns req.Owner's unordered-nonce bitmap word req.Word.
func (p *PositionManager) Nonces(ctx context.Context, opts *bind.CallOptions, req *NoncesRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: nonces: %w", err)
	}

	calldata, err := EncodeNonces(req.Owner, req.Word)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: nonces: %w", err)
	}

	data, err := bind.Call(ctx, p.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: nonces: %w", mapPositionManagerRevertError(err))
	}

	out, err := DecodeNonces(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: nonces: %w", err)
	}

	return out, nil
}

// CallInitializePool runs initializePool via eth_call as opts.From, returning the new tick (type(int24).max if the pool existed) or the revert.
func (p *PositionManager) CallInitializePool(ctx context.Context, opts *bind.CallOptions, req *InitializePoolRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: callInitializePool: %w", err)
	}

	calldata, err := EncodeInitializePool(req.Key, req.SqrtPriceX96)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callInitializePool: %w", err)
	}

	data, err := bind.Call(ctx, p.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callInitializePool: %w", mapPositionManagerRevertError(err))
	}

	tick, err := DecodeInitializePool(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callInitializePool: %w", err)
	}

	return tick, nil
}

// CallMulticall runs multicall via eth_call as opts.From with req.NativeAmount as msg.value, returning each sub-call's result or the revert; opts.Value must be nil.
func (p *PositionManager) CallMulticall(ctx context.Context, opts *bind.CallOptions, req *MulticallRequest) ([][]byte, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: callMulticall: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, fmt.Errorf("uniswapv4: callMulticall: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeMulticall(req.Data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callMulticall: %w", err)
	}

	callOpts := bind.NewCallOptions()
	if opts != nil {
		*callOpts = *opts
	}
	callOpts.Value = req.NativeAmount

	data, err := bind.Call(ctx, p.c, callOpts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callMulticall: %w", mapPositionManagerRevertError(err))
	}

	results, err := DecodeMulticall(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callMulticall: %w", err)
	}

	return results, nil
}

// CallPermit2Permit runs Permit2Forwarder.permit via eth_call as opts.From, returning the inner permit2 revert data (empty on success) or the revert.
func (p *PositionManager) CallPermit2Permit(ctx context.Context, opts *bind.CallOptions, req *Permit2PermitRequest) ([]byte, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: callPermit2Permit: %w", err)
	}

	calldata, err := EncodePermit2Permit(req.Owner, req.PermitSingle, req.Signature)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callPermit2Permit: %w", err)
	}

	data, err := bind.Call(ctx, p.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callPermit2Permit: %w", mapPositionManagerRevertError(err))
	}

	out, err := DecodePermit2Permit(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callPermit2Permit: %w", err)
	}

	return out, nil
}

// CallPermit2PermitBatch runs Permit2Forwarder.permitBatch via eth_call as opts.From, returning the inner permit2 revert data (empty on success) or the revert.
func (p *PositionManager) CallPermit2PermitBatch(ctx context.Context, opts *bind.CallOptions, req *Permit2PermitBatchRequest) ([]byte, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: callPermit2PermitBatch: %w", err)
	}

	calldata, err := EncodePermit2PermitBatch(req.Owner, req.PermitBatch, req.Signature)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callPermit2PermitBatch: %w", err)
	}

	data, err := bind.Call(ctx, p.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callPermit2PermitBatch: %w", mapPositionManagerRevertError(err))
	}

	out, err := DecodePermit2PermitBatch(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callPermit2PermitBatch: %w", err)
	}

	return out, nil
}

// ModifyLiquidities sends modifyLiquidities(abi.encode(actions, params), deadline) on req.Contract, signed by opts.Signer with req.NativeAmount (nil sends none) as msg.value; opts.Value must be nil.
func (p *PositionManager) ModifyLiquidities(ctx context.Context, opts *bind.TransactOptions, req *ModifyLiquiditiesRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: modifyLiquidities: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv4: modifyLiquidities: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeModifyLiquidities(req.Actions, req.Params, req.Deadline)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: modifyLiquidities: %w", err)
	}

	var txOpts *bind.TransactOptions
	if opts != nil {
		o := *opts
		o.Value = req.NativeAmount
		txOpts = &o
	}

	signed, hash, err := bind.Transact(ctx, p.c, txOpts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: modifyLiquidities: %w", mapPositionManagerRevertError(err))
	}

	return signed, hash, nil
}

// ModifyLiquiditiesWithoutUnlock sends modifyLiquiditiesWithoutUnlock(actions, params) on req.Contract, signed by opts.Signer with req.NativeAmount (nil sends none) as msg.value; opts.Value must be nil.
func (p *PositionManager) ModifyLiquiditiesWithoutUnlock(ctx context.Context, opts *bind.TransactOptions, req *ModifyLiquiditiesWithoutUnlockRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: modifyLiquiditiesWithoutUnlock: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv4: modifyLiquiditiesWithoutUnlock: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeModifyLiquiditiesWithoutUnlock(req.Actions, req.Params)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: modifyLiquiditiesWithoutUnlock: %w", err)
	}

	var txOpts *bind.TransactOptions
	if opts != nil {
		o := *opts
		o.Value = req.NativeAmount
		txOpts = &o
	}

	signed, hash, err := bind.Transact(ctx, p.c, txOpts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: modifyLiquiditiesWithoutUnlock: %w", mapPositionManagerRevertError(err))
	}

	return signed, hash, nil
}

// InitializePool sends initializePool(key, sqrtPriceX96) on req.Contract, signed by opts.Signer.
func (p *PositionManager) InitializePool(ctx context.Context, opts *bind.TransactOptions, req *InitializePoolRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: initializePool: %w", err)
	}

	calldata, err := EncodeInitializePool(req.Key, req.SqrtPriceX96)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: initializePool: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, p.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: initializePool: %w", mapPositionManagerRevertError(err))
	}

	return signed, hash, nil
}

// Multicall sends multicall(data) on req.Contract, signed by opts.Signer with req.NativeAmount (nil sends none) as msg.value; opts.Value must be nil.
func (p *PositionManager) Multicall(ctx context.Context, opts *bind.TransactOptions, req *MulticallRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: multicall: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv4: multicall: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeMulticall(req.Data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: multicall: %w", err)
	}

	var txOpts *bind.TransactOptions
	if opts != nil {
		o := *opts
		o.Value = req.NativeAmount
		txOpts = &o
	}

	signed, hash, err := bind.Transact(ctx, p.c, txOpts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: multicall: %w", mapPositionManagerRevertError(err))
	}

	return signed, hash, nil
}

// Subscribe sends subscribe(tokenId, newSubscriber, data) on req.Contract, signed by opts.Signer.
func (p *PositionManager) Subscribe(ctx context.Context, opts *bind.TransactOptions, req *SubscribeRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: subscribe: %w", err)
	}

	calldata, err := EncodeSubscribe(req.TokenID, req.Subscriber, req.Data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: subscribe: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, p.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: subscribe: %w", mapPositionManagerRevertError(err))
	}

	return signed, hash, nil
}

// Unsubscribe sends unsubscribe(tokenId) on req.Contract, signed by opts.Signer.
func (p *PositionManager) Unsubscribe(ctx context.Context, opts *bind.TransactOptions, req *UnsubscribeRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: unsubscribe: %w", err)
	}

	calldata, err := EncodeUnsubscribe(req.TokenID)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: unsubscribe: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, p.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: unsubscribe: %w", mapPositionManagerRevertError(err))
	}

	return signed, hash, nil
}

// RevokeNonce sends revokeNonce(nonce) on req.Contract, signed by opts.Signer.
func (p *PositionManager) RevokeNonce(ctx context.Context, opts *bind.TransactOptions, req *RevokeNonceRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: revokeNonce: %w", err)
	}

	calldata, err := EncodeRevokeNonce(req.Nonce)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: revokeNonce: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, p.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: revokeNonce: %w", mapPositionManagerRevertError(err))
	}

	return signed, hash, nil
}

// ERC721Permit sends permit(spender, tokenId, deadline, nonce, signature) on req.Contract, signed by opts.Signer.
func (p *PositionManager) ERC721Permit(ctx context.Context, opts *bind.TransactOptions, req *ERC721PermitRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: erc721Permit: %w", err)
	}

	calldata, err := EncodeERC721Permit(req.Spender, req.TokenID, req.Deadline, req.Nonce, req.Signature)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: erc721Permit: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, p.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: erc721Permit: %w", mapPositionManagerRevertError(err))
	}

	return signed, hash, nil
}

// ERC721PermitForAll sends permitForAll(owner, operator, approved, deadline, nonce, signature) on req.Contract, signed by opts.Signer.
func (p *PositionManager) ERC721PermitForAll(ctx context.Context, opts *bind.TransactOptions, req *ERC721PermitForAllRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: erc721PermitForAll: %w", err)
	}

	calldata, err := EncodeERC721PermitForAll(req.Owner, req.Operator, req.Approved, req.Deadline, req.Nonce, req.Signature)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: erc721PermitForAll: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, p.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: erc721PermitForAll: %w", mapPositionManagerRevertError(err))
	}

	return signed, hash, nil
}

// Permit2Permit sends Permit2Forwarder.permit(owner, permitSingle, signature) on req.Contract, signed by opts.Signer.
func (p *PositionManager) Permit2Permit(ctx context.Context, opts *bind.TransactOptions, req *Permit2PermitRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: permit2Permit: %w", err)
	}

	calldata, err := EncodePermit2Permit(req.Owner, req.PermitSingle, req.Signature)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: permit2Permit: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, p.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: permit2Permit: %w", mapPositionManagerRevertError(err))
	}

	return signed, hash, nil
}

// Permit2PermitBatch sends Permit2Forwarder.permitBatch(owner, permitBatch, signature) on req.Contract, signed by opts.Signer.
func (p *PositionManager) Permit2PermitBatch(ctx context.Context, opts *bind.TransactOptions, req *Permit2PermitBatchRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: permit2PermitBatch: %w", err)
	}

	calldata, err := EncodePermit2PermitBatch(req.Owner, req.PermitBatch, req.Signature)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: permit2PermitBatch: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, p.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: permit2PermitBatch: %w", mapPositionManagerRevertError(err))
	}

	return signed, hash, nil
}
