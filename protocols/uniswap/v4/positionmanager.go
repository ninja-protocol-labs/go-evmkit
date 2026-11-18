package v4

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/protocols/permit2"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc721"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// IPositionManagerImmutables mirrors IImmutableState plus NativeWrapper's and Permit2Forwarder's own
// immutable getters.
type IPositionManagerImmutables interface {
	PoolManager(ctx context.Context, contract *types.Address, block string) (*types.Address, error)
	WETH9(ctx context.Context, contract *types.Address, block string) (*types.Address, error)
	Permit2(ctx context.Context, contract *types.Address, block string) (*types.Address, error)
}

// IPositionManagerActions mirrors IPositionManager's own batched-action entry points and view
// functions. No tx.Builder-taking constructors here -- see CreateModifyLiquidities and
// CreateModifyLiquiditiesWithoutUnlock.
type IPositionManagerActions interface {
	ModifyLiquidities(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	ModifyLiquiditiesWithoutUnlock(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	NextTokenID(ctx context.Context, contract *types.Address, block string) (*big.Int, error)
	GetPositionLiquidity(ctx context.Context, contract *types.Address, tokenID *big.Int, block string) (*big.Int, error)
	GetPoolAndPositionInfo(ctx context.Context, contract *types.Address, tokenID *big.Int, block string) (*PoolKey, *PositionInfo, error)
	PositionInfo(ctx context.Context, contract *types.Address, tokenID *big.Int, block string) (*PositionInfo, error)
	PoolKeys(ctx context.Context, contract *types.Address, poolID [25]byte, block string) (*PoolKey, error)
}

// IPositionManagerNotifier mirrors INotifier. No tx.Builder-taking Create* constructors here -- see
// CreateSubscribe and CreateUnsubscribe.
type IPositionManagerNotifier interface {
	Subscribe(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Unsubscribe(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Subscriber(ctx context.Context, contract *types.Address, tokenID *big.Int, block string) (*types.Address, error)
	UnsubscribeGasLimit(ctx context.Context, contract *types.Address, block string) (*big.Int, error)
}

// IPositionManagerPoolInitializer mirrors IPoolInitializer_v4. No tx.Builder-taking Create* constructor
// here -- see CreateInitializePool.
type IPositionManagerPoolInitializer interface {
	InitializePool(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

// IPositionManagerMulticall mirrors IMulticall_v4. No tx.Builder-taking Create* constructor here -- see
// CreateMulticall.
type IPositionManagerMulticall interface {
	Multicall(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

// IPositionManagerUnorderedNonce mirrors IUnorderedNonce. No tx.Builder-taking Create* constructor here
// -- see CreateRevokeNonce.
type IPositionManagerUnorderedNonce interface {
	Nonces(ctx context.Context, contract, owner *types.Address, word *big.Int, block string) (*big.Int, error)
	RevokeNonce(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

// IPositionManagerPermit mirrors IERC721Permit_v4 plus IPermit2Forwarder. No tx.Builder-taking Create*
// constructors here -- see CreateERC721Permit, CreateERC721PermitForAll, CreatePermit2Permit,
// and CreatePermit2PermitBatch.
type IPositionManagerPermit interface {
	ERC721Permit(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	ERC721PermitForAll(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Permit2Permit(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Permit2PermitBatch(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

// IPositionManager mirrors PositionManager, composed the same way its Solidity base contracts are: one
// small interface per concern, plus the full ERC721 surface it wraps positions in.
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

// PositionManager calls any PositionManager deployment over a Client; it holds no contract
// address, so one instance serves every deployment. The embedded *erc721.ERC721 serves the plain
// ERC721 surface PositionManager wraps positions in (position NFTs).
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

// CreateModifyLiquidities builds an unsigned modifyLiquidities(abi.encode(actions, params),
// deadline) transaction via b, sent by sender with value wei attached (for native-ETH-funded
// actions). This is the standard entry point for batching position actions.
func (p *PositionManager) CreateModifyLiquidities(ctx context.Context, b tx.Builder, contract, sender *types.Address, actions []byte, params [][]byte, deadline, value *big.Int) (core.Transaction, error) {
	calldata, err := EncodeModifyLiquidities(actions, params, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createModifyLiquidities: %w", err)
	}
	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createModifyLiquidities: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// ModifyLiquidities broadcasts an already-signed modifyLiquidities transaction.
func (p *PositionManager) ModifyLiquidities(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: modifyLiquidities: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: modifyLiquidities: %w", err)
	}
	return hash, nil
}

// CreateModifyLiquiditiesWithoutUnlock builds an unsigned modifyLiquiditiesWithoutUnlock(actions,
// params) transaction via b, sent by sender with value wei attached. Only valid when sender is
// itself a contract that has already unlocked PoolManager.
func (p *PositionManager) CreateModifyLiquiditiesWithoutUnlock(ctx context.Context, b tx.Builder, contract, sender *types.Address, actions []byte, params [][]byte, value *big.Int) (core.Transaction, error) {
	calldata, err := EncodeModifyLiquiditiesWithoutUnlock(actions, params)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createModifyLiquiditiesWithoutUnlock: %w", err)
	}
	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createModifyLiquiditiesWithoutUnlock: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// ModifyLiquiditiesWithoutUnlock broadcasts an already-signed modifyLiquiditiesWithoutUnlock transaction.
func (p *PositionManager) ModifyLiquiditiesWithoutUnlock(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: modifyLiquiditiesWithoutUnlock: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: modifyLiquiditiesWithoutUnlock: %w", err)
	}
	return hash, nil
}

// NextTokenID returns the tokenId that will be used for the next minted position.
func (p *PositionManager) NextTokenID(ctx context.Context, contract *types.Address, block string) (*big.Int, error) {
	calldata, err := EncodeNextTokenID()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: nextTokenID: %w", err)
	}
	data, err := p.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: nextTokenID: %w", err)
	}
	tokenID, err := DecodeNextTokenID(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: nextTokenID: %w", err)
	}
	return tokenID, nil
}

// GetPositionLiquidity returns tokenId's liquidity amount.
func (p *PositionManager) GetPositionLiquidity(ctx context.Context, contract *types.Address, tokenID *big.Int, block string) (*big.Int, error) {
	calldata, err := EncodeGetPositionLiquidity(tokenID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getPositionLiquidity: %w", err)
	}
	data, err := p.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getPositionLiquidity: %w", err)
	}
	liquidity, err := DecodeGetPositionLiquidity(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: getPositionLiquidity: %w", err)
	}
	return liquidity, nil
}

// GetPoolAndPositionInfo returns tokenId's pool key and packed PositionInfo.
func (p *PositionManager) GetPoolAndPositionInfo(ctx context.Context, contract *types.Address, tokenID *big.Int, block string) (*PoolKey, *PositionInfo, error) {
	calldata, err := EncodeGetPoolAndPositionInfo(tokenID)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: getPoolAndPositionInfo: %w", err)
	}
	data, err := p.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: getPoolAndPositionInfo: %w", err)
	}
	poolKey, info, err := DecodeGetPoolAndPositionInfo(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: getPoolAndPositionInfo: %w", err)
	}
	return poolKey, info, nil
}

// PositionInfo returns tokenId's packed PositionInfo.
func (p *PositionManager) PositionInfo(ctx context.Context, contract *types.Address, tokenID *big.Int, block string) (*PositionInfo, error) {
	calldata, err := EncodePositionInfo(tokenID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: positionInfo: %w", err)
	}
	data, err := p.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: positionInfo: %w", err)
	}
	info, err := DecodePositionInfo(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: positionInfo: %w", err)
	}
	return info, nil
}

// PoolKeys returns the full PoolKey for a truncated 25-byte poolId (as stored in a PositionInfo).
func (p *PositionManager) PoolKeys(ctx context.Context, contract *types.Address, poolID [25]byte, block string) (*PoolKey, error) {
	calldata, err := EncodePoolKeys(poolID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: poolKeys: %w", err)
	}
	data, err := p.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: poolKeys: %w", err)
	}
	key, err := DecodePoolKeys(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: poolKeys: %w", err)
	}
	return key, nil
}

// CreateSubscribe builds an unsigned subscribe(tokenId, newSubscriber, data) transaction via b,
// sent by sender with value wei attached.
func (p *PositionManager) CreateSubscribe(ctx context.Context, b tx.Builder, contract, sender *types.Address, tokenID *big.Int, newSubscriber *types.Address, data []byte, value *big.Int) (core.Transaction, error) {
	calldata, err := EncodeSubscribe(tokenID, newSubscriber, data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createSubscribe: %w", err)
	}
	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createSubscribe: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// Subscribe broadcasts an already-signed subscribe transaction.
func (p *PositionManager) Subscribe(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: subscribe: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: subscribe: %w", err)
	}
	return hash, nil
}

// CreateUnsubscribe builds an unsigned unsubscribe(tokenId) transaction via b, sent by sender
// with value wei attached.
func (p *PositionManager) CreateUnsubscribe(ctx context.Context, b tx.Builder, contract, sender *types.Address, tokenID, value *big.Int) (core.Transaction, error) {
	calldata, err := EncodeUnsubscribe(tokenID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createUnsubscribe: %w", err)
	}
	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createUnsubscribe: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// Unsubscribe broadcasts an already-signed unsubscribe transaction.
func (p *PositionManager) Unsubscribe(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: unsubscribe: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: unsubscribe: %w", err)
	}
	return hash, nil
}

// Subscriber returns tokenId's subscriber contract address (the zero address if unsubscribed).
func (p *PositionManager) Subscriber(ctx context.Context, contract *types.Address, tokenID *big.Int, block string) (*types.Address, error) {
	calldata, err := EncodeSubscriber(tokenID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: subscriber: %w", err)
	}
	data, err := p.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: subscriber: %w", err)
	}
	subscriber, err := DecodeSubscriber(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: subscriber: %w", err)
	}
	return subscriber, nil
}

// UnsubscribeGasLimit returns the maximum gas-used allowed when notifying a subscriber's
// notifyUnsubscribe function.
func (p *PositionManager) UnsubscribeGasLimit(ctx context.Context, contract *types.Address, block string) (*big.Int, error) {
	calldata, err := EncodeUnsubscribeGasLimit()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: unsubscribeGasLimit: %w", err)
	}
	data, err := p.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: unsubscribeGasLimit: %w", err)
	}
	gasLimit, err := DecodeUnsubscribeGasLimit(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: unsubscribeGasLimit: %w", err)
	}
	return gasLimit, nil
}

// CreateInitializePool builds an unsigned initializePool(key, sqrtPriceX96) transaction via b,
// sent by sender with value wei attached.
func (p *PositionManager) CreateInitializePool(ctx context.Context, b tx.Builder, contract, sender *types.Address, key *PoolKey, sqrtPriceX96, value *big.Int) (core.Transaction, error) {
	calldata, err := EncodeInitializePool(key, sqrtPriceX96)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createInitializePool: %w", err)
	}
	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createInitializePool: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// CallInitializePool simulates initializePool(key, sqrtPriceX96) via eth_call, returning the
// resulting tick (or type(int24).max if the pool already existed) without broadcasting anything.
func (p *PositionManager) CallInitializePool(ctx context.Context, contract, sender *types.Address, key *PoolKey, sqrtPriceX96, value *big.Int, block string) (*big.Int, error) {
	calldata, err := EncodeInitializePool(key, sqrtPriceX96)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callInitializePool: %w", err)
	}
	data, err := p.call(ctx, sender, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callInitializePool: %w", err)
	}
	tick, err := DecodeInitializePool(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callInitializePool: %w", err)
	}
	return tick, nil
}

// InitializePool broadcasts an already-signed initializePool transaction.
func (p *PositionManager) InitializePool(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: initializePool: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: initializePool: %w", err)
	}
	return hash, nil
}

// CreateMulticall builds an unsigned multicall(data) transaction via b, sent by sender with value
// wei attached (shared across every sub-call).
func (p *PositionManager) CreateMulticall(ctx context.Context, b tx.Builder, contract, sender *types.Address, data [][]byte, value *big.Int) (core.Transaction, error) {
	calldata, err := EncodeMulticall(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createMulticall: %w", err)
	}
	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createMulticall: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// CallMulticall simulates multicall(data) via eth_call, returning each sub-call's own return
// data without broadcasting anything.
func (p *PositionManager) CallMulticall(ctx context.Context, contract, sender *types.Address, data [][]byte, value *big.Int, block string) ([][]byte, error) {
	calldata, err := EncodeMulticall(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callMulticall: %w", err)
	}
	raw, err := p.call(ctx, sender, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callMulticall: %w", err)
	}
	results, err := DecodeMulticall(raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callMulticall: %w", err)
	}
	return results, nil
}

// Multicall broadcasts an already-signed multicall transaction.
func (p *PositionManager) Multicall(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: multicall: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: multicall: %w", err)
	}
	return hash, nil
}

// PoolManager returns the PoolManager contract this PositionManager deployment is wired to.
func (p *PositionManager) PoolManager(ctx context.Context, contract *types.Address, block string) (*types.Address, error) {
	calldata, err := EncodePoolManager()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: poolManager: %w", err)
	}
	data, err := p.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: poolManager: %w", err)
	}
	addr, err := DecodePoolManager(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: poolManager: %w", err)
	}
	return addr, nil
}

// WETH9 returns the WETH9 contract this PositionManager wraps and unwraps native ETH through.
func (p *PositionManager) WETH9(ctx context.Context, contract *types.Address, block string) (*types.Address, error) {
	calldata, err := EncodeWETH9()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: weth9: %w", err)
	}
	data, err := p.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: weth9: %w", err)
	}
	addr, err := DecodeWETH9(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: weth9: %w", err)
	}
	return addr, nil
}

// Permit2 returns the Permit2 contract this PositionManager forwards permit/permitBatch to.
func (p *PositionManager) Permit2(ctx context.Context, contract *types.Address, block string) (*types.Address, error) {
	calldata, err := EncodePermit2Contract()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: permit2: %w", err)
	}
	data, err := p.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: permit2: %w", err)
	}
	addr, err := DecodePermit2Contract(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: permit2: %w", err)
	}
	return addr, nil
}

// Nonces returns owner's nonce bitmap word.
func (p *PositionManager) Nonces(ctx context.Context, contract, owner *types.Address, word *big.Int, block string) (*big.Int, error) {
	calldata, err := EncodeNonces(owner, word)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: nonces: %w", err)
	}
	data, err := p.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: nonces: %w", err)
	}
	bitmap, err := DecodeNonces(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: nonces: %w", err)
	}
	return bitmap, nil
}

// CreateRevokeNonce builds an unsigned revokeNonce(nonce) transaction via b, sent by sender with
// value wei attached.
func (p *PositionManager) CreateRevokeNonce(ctx context.Context, b tx.Builder, contract, sender *types.Address, nonce, value *big.Int) (core.Transaction, error) {
	calldata, err := EncodeRevokeNonce(nonce)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createRevokeNonce: %w", err)
	}
	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createRevokeNonce: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// RevokeNonce broadcasts an already-signed revokeNonce transaction.
func (p *PositionManager) RevokeNonce(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: revokeNonce: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: revokeNonce: %w", err)
	}
	return hash, nil
}

// CreateERC721Permit builds an unsigned permit(spender, tokenId, deadline, nonce, signature)
// transaction via b, sent by sender with value wei attached.
func (p *PositionManager) CreateERC721Permit(ctx context.Context, b tx.Builder, contract, sender, spender *types.Address, tokenID, deadline, nonce *big.Int, signature []byte, value *big.Int) (core.Transaction, error) {
	calldata, err := EncodeERC721Permit(spender, tokenID, deadline, nonce, signature)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createERC721Permit: %w", err)
	}
	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createERC721Permit: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// ERC721Permit broadcasts an already-signed permit transaction.
func (p *PositionManager) ERC721Permit(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: erc721Permit: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: erc721Permit: %w", err)
	}
	return hash, nil
}

// CreateERC721PermitForAll builds an unsigned permitForAll(owner, operator, approved, deadline,
// nonce, signature) transaction via b, sent by sender with value wei attached.
func (p *PositionManager) CreateERC721PermitForAll(ctx context.Context, b tx.Builder, contract, sender, owner, operator *types.Address, approved bool, deadline, nonce *big.Int, signature []byte, value *big.Int) (core.Transaction, error) {
	calldata, err := EncodeERC721PermitForAll(owner, operator, approved, deadline, nonce, signature)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createERC721PermitForAll: %w", err)
	}
	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createERC721PermitForAll: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// ERC721PermitForAll broadcasts an already-signed permitForAll transaction.
func (p *PositionManager) ERC721PermitForAll(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: erc721PermitForAll: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: erc721PermitForAll: %w", err)
	}
	return hash, nil
}

// CreatePermit2Permit builds an unsigned Permit2Forwarder.permit(owner, permitSingle, signature)
// transaction via b, sent by sender with value wei attached.
func (p *PositionManager) CreatePermit2Permit(ctx context.Context, b tx.Builder, contract, sender, owner *types.Address, permitSingle *permit2.PermitSingle, signature []byte, value *big.Int) (core.Transaction, error) {
	calldata, err := EncodePermit2Permit(owner, permitSingle, signature)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createPermit2Permit: %w", err)
	}
	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createPermit2Permit: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// CallPermit2Permit simulates Permit2Forwarder.permit(owner, permitSingle, signature) via
// eth_call, returning the inner permit2 call's revert reason (empty on success) without
// broadcasting anything.
func (p *PositionManager) CallPermit2Permit(ctx context.Context, contract, sender, owner *types.Address, permitSingle *permit2.PermitSingle, signature []byte, value *big.Int, block string) ([]byte, error) {
	calldata, err := EncodePermit2Permit(owner, permitSingle, signature)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callPermit2Permit: %w", err)
	}
	data, err := p.call(ctx, sender, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callPermit2Permit: %w", err)
	}
	errData, err := DecodePermit2Permit(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callPermit2Permit: %w", err)
	}
	return errData, nil
}

// Permit2Permit broadcasts an already-signed Permit2Forwarder.permit transaction.
func (p *PositionManager) Permit2Permit(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: permit2Permit: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: permit2Permit: %w", err)
	}
	return hash, nil
}

// CreatePermit2PermitBatch builds an unsigned Permit2Forwarder.permitBatch(owner, permitBatch,
// signature) transaction via b, sent by sender with value wei attached.
func (p *PositionManager) CreatePermit2PermitBatch(ctx context.Context, b tx.Builder, contract, sender, owner *types.Address, permitBatch *permit2.PermitBatch, signature []byte, value *big.Int) (core.Transaction, error) {
	calldata, err := EncodePermit2PermitBatch(owner, permitBatch, signature)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createPermit2PermitBatch: %w", err)
	}
	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createPermit2PermitBatch: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// CallPermit2PermitBatch simulates Permit2Forwarder.permitBatch(owner, permitBatch, signature)
// via eth_call, returning the inner permit2 call's revert reason (empty on success) without
// broadcasting anything.
func (p *PositionManager) CallPermit2PermitBatch(ctx context.Context, contract, sender, owner *types.Address, permitBatch *permit2.PermitBatch, signature []byte, value *big.Int, block string) ([]byte, error) {
	calldata, err := EncodePermit2PermitBatch(owner, permitBatch, signature)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callPermit2PermitBatch: %w", err)
	}
	data, err := p.call(ctx, sender, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callPermit2PermitBatch: %w", err)
	}
	errData, err := DecodePermit2PermitBatch(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callPermit2PermitBatch: %w", err)
	}
	return errData, nil
}

// Permit2PermitBatch broadcasts an already-signed Permit2Forwarder.permitBatch transaction.
func (p *PositionManager) Permit2PermitBatch(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: permit2PermitBatch: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: permit2PermitBatch: %w", err)
	}
	return hash, nil
}

// call eth_calls contract with calldata, optionally simulating from a specific sender.
func (p *PositionManager) call(ctx context.Context, from, contract *types.Address, calldata []byte, block string) ([]byte, error) {
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

// wrapRevertError maps a known PositionManager-specific custom error in err to its sentinel
// (IPositionManager, IERC721Permit_v4, INotifier, IUnorderedNonce, ReentrancyLock, NativeWrapper,
// DeltaResolver, SlippageCheck, and permit2's SignatureVerification, all inherited by
// PositionManager). A revert that bubbles up from the underlying PoolManager call isn't mapped
// here; it falls through to ErrUnknown, since that selector space belongs to poolmanager.go.
func (_ *PositionManager) wrapRevertError(err error) error {
	var rpcErr *rpc.ResponseError
	if !errors.As(err, &rpcErr) || len(rpcErr.Data) == 0 {
		return err
	}

	var hexData string
	if jsonErr := json.Unmarshal(rpcErr.Data, &hexData); jsonErr != nil {
		return err
	}

	data, decErr := encoding.Hex.Decode(hexData)
	if decErr != nil || len(data) < 4 {
		return err
	}

	var sel abi.Selector
	copy(sel[:], data[:4])

	switch sel {
	case errNotApproved.Selector():
		return fmt.Errorf("%w: %w", ErrNotApproved, err)
	case errDeadlinePassed.Selector():
		return fmt.Errorf("%w: %w", ErrDeadlinePassed, err)
	case errPoolManagerMustBeLocked.Selector():
		return fmt.Errorf("%w: %w", ErrPoolManagerMustBeLocked, err)
	case errSignatureDeadlineExpired.Selector():
		return fmt.Errorf("%w: %w", ErrSignatureDeadlineExpired, err)
	case errNoSelfPermit.Selector():
		return fmt.Errorf("%w: %w", ErrNoSelfPermit, err)
	case errUnauthorized.Selector():
		return fmt.Errorf("%w: %w", ErrUnauthorized, err)
	case errNotSubscribed.Selector():
		return fmt.Errorf("%w: %w", ErrNotSubscribed, err)
	case errNoCodeSubscriber.Selector():
		return fmt.Errorf("%w: %w", ErrNoCodeSubscriber, err)
	case errGasLimitTooLow.Selector():
		return fmt.Errorf("%w: %w", ErrGasLimitTooLow, err)
	case errSubscriptionReverted.Selector():
		return fmt.Errorf("%w: %w", ErrSubscriptionReverted, err)
	case errModifyLiquidityNotificationReverted.Selector():
		return fmt.Errorf("%w: %w", ErrModifyLiquidityNotificationReverted, err)
	case errBurnNotificationReverted.Selector():
		return fmt.Errorf("%w: %w", ErrBurnNotificationReverted, err)
	case errAlreadySubscribed.Selector():
		return fmt.Errorf("%w: %w", ErrAlreadySubscribed, err)
	case errNonceAlreadyUsed.Selector():
		return fmt.Errorf("%w: %w", ErrNonceAlreadyUsed, err)
	case errContractLocked.Selector():
		return fmt.Errorf("%w: %w", ErrContractLocked, err)
	case errInvalidEthSender.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidEthSender, err)
	case errDeltaNotPositive.Selector():
		return fmt.Errorf("%w: %w", ErrDeltaNotPositive, err)
	case errDeltaNotNegative.Selector():
		return fmt.Errorf("%w: %w", ErrDeltaNotNegative, err)
	case errInsufficientBalance.Selector():
		return fmt.Errorf("%w: %w", ErrInsufficientBalance, err)
	case errMaximumAmountExceeded.Selector():
		return fmt.Errorf("%w: %w", ErrMaximumAmountExceeded, err)
	case errMinimumAmountInsufficient.Selector():
		return fmt.Errorf("%w: %w", ErrMinimumAmountInsufficient, err)
	case errInvalidSignatureLength.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidSignatureLength, err)
	case errInvalidSignature.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidSignature, err)
	case errInvalidSigner.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidSigner, err)
	case errInvalidContractSignature.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidContractSignature, err)
	default:
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}
}
