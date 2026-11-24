// Package v4 encodes calls to and decodes results from Uniswap v4's PoolManager.
package v4

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc6909"
)

// IERC6909Claims mirrors IERC6909Claims.sol; same method set as erc6909.IERC6909, so reused directly.
type IERC6909Claims = erc6909.IERC6909

// IProtocolFees mirrors IProtocolFees.sol.
type IProtocolFees interface {
	ProtocolFeesAccrued(ctx context.Context, opts *bind.CallOptions, req *ProtocolFeesAccruedRequest) (*big.Int, error)
	ProtocolFeeController(ctx context.Context, opts *bind.CallOptions, req *ProtocolFeeControllerRequest) (*types.Address, error)
	CallCollectProtocolFees(ctx context.Context, opts *bind.CallOptions, req *CollectProtocolFeesRequest) (*big.Int, error)

	SetProtocolFee(ctx context.Context, opts *bind.TransactOptions, req *SetProtocolFeeRequest) (core.Transaction, *types.Hash, error)
	SetProtocolFeeController(ctx context.Context, opts *bind.TransactOptions, req *SetProtocolFeeControllerRequest) (core.Transaction, *types.Hash, error)
	CollectProtocolFees(ctx context.Context, opts *bind.TransactOptions, req *CollectProtocolFeesRequest) (core.Transaction, *types.Hash, error)
}

// IExtsload mirrors IExtsload.sol; its 3 overloaded extsload(...) become distinctly-named methods.
type IExtsload interface {
	Extsload(ctx context.Context, opts *bind.CallOptions, req *ExtsloadRequest) ([32]byte, error)
	ExtsloadRange(ctx context.Context, opts *bind.CallOptions, req *ExtsloadRangeRequest) ([][32]byte, error)
	ExtsloadSlots(ctx context.Context, opts *bind.CallOptions, req *ExtsloadSlotsRequest) ([][32]byte, error)
}

// IExttload mirrors IExttload.sol; its 2 overloaded exttload(...) become distinctly-named methods.
type IExttload interface {
	Exttload(ctx context.Context, opts *bind.CallOptions, req *ExttloadRequest) ([32]byte, error)
	ExttloadSlots(ctx context.Context, opts *bind.CallOptions, req *ExttloadSlotsRequest) ([][32]byte, error)
}

// IPoolManager mirrors IPoolManager.sol plus IProtocolFees, IERC6909Claims, IExtsload and IExttload; most actions revert with ManagerLocked outside unlock.
type IPoolManager interface {
	IProtocolFees
	IERC6909Claims
	IExtsload
	IExttload

	CallInitialize(ctx context.Context, opts *bind.CallOptions, req *InitializeRequest) (*big.Int, error)
	CallModifyLiquidity(ctx context.Context, opts *bind.CallOptions, req *ModifyLiquidityRequest) (*BalanceDelta, *BalanceDelta, error)
	CallSwap(ctx context.Context, opts *bind.CallOptions, req *SwapRequest) (*BalanceDelta, error)
	CallDonate(ctx context.Context, opts *bind.CallOptions, req *DonateRequest) (*BalanceDelta, error)
	CallUnlock(ctx context.Context, opts *bind.CallOptions, req *UnlockRequest) ([]byte, error)
	CallSettle(ctx context.Context, opts *bind.CallOptions, req *SettleRequest) (*big.Int, error)
	CallSettleFor(ctx context.Context, opts *bind.CallOptions, req *SettleForRequest) (*big.Int, error)

	Initialize(ctx context.Context, opts *bind.TransactOptions, req *InitializeRequest) (core.Transaction, *types.Hash, error)
	ModifyLiquidity(ctx context.Context, opts *bind.TransactOptions, req *ModifyLiquidityRequest) (core.Transaction, *types.Hash, error)
	Swap(ctx context.Context, opts *bind.TransactOptions, req *SwapRequest) (core.Transaction, *types.Hash, error)
	Donate(ctx context.Context, opts *bind.TransactOptions, req *DonateRequest) (core.Transaction, *types.Hash, error)
	Unlock(ctx context.Context, opts *bind.TransactOptions, req *UnlockRequest) (core.Transaction, *types.Hash, error)
	Sync(ctx context.Context, opts *bind.TransactOptions, req *SyncRequest) (core.Transaction, *types.Hash, error)
	Take(ctx context.Context, opts *bind.TransactOptions, req *TakeRequest) (core.Transaction, *types.Hash, error)
	Settle(ctx context.Context, opts *bind.TransactOptions, req *SettleRequest) (core.Transaction, *types.Hash, error)
	SettleFor(ctx context.Context, opts *bind.TransactOptions, req *SettleForRequest) (core.Transaction, *types.Hash, error)
	Clear(ctx context.Context, opts *bind.TransactOptions, req *ClearRequest) (core.Transaction, *types.Hash, error)
	Mint(ctx context.Context, opts *bind.TransactOptions, req *MintRequest) (core.Transaction, *types.Hash, error)
	Burn(ctx context.Context, opts *bind.TransactOptions, req *BurnRequest) (core.Transaction, *types.Hash, error)
	UpdateDynamicLPFee(ctx context.Context, opts *bind.TransactOptions, req *UpdateDynamicLPFeeRequest) (core.Transaction, *types.Hash, error)
}

var _ IPoolManager = (*PoolManager)(nil)

// PoolManager calls any PoolManager contract over a Client; ERC-6909 claim methods are promoted from the embedded *erc6909.ERC6909.
type PoolManager struct {
	*erc6909.ERC6909

	c rpc.Client
}

// New returns a PoolManager calling deployments via cli.
func New(cli rpc.Client) *PoolManager {
	return &PoolManager{
		ERC6909: erc6909.New(cli),
		c:       cli,
	}
}

// ProtocolFeesAccrued returns the protocol fees req.Contract has accrued in req.Currency.
func (m *PoolManager) ProtocolFeesAccrued(ctx context.Context, opts *bind.CallOptions, req *ProtocolFeesAccruedRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: protocolFeesAccrued: %w", err)
	}

	calldata, err := EncodeProtocolFeesAccrued(req.Currency)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: protocolFeesAccrued: %w", err)
	}

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: protocolFeesAccrued: %w", mapPoolManagerRevertError(err))
	}

	out, err := DecodeProtocolFeesAccrued(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: protocolFeesAccrued: %w", err)
	}

	return out, nil
}

// ProtocolFeeController returns req.Contract's protocol fee controller, the only account allowed to SetProtocolFee/CollectProtocolFees.
func (m *PoolManager) ProtocolFeeController(ctx context.Context, opts *bind.CallOptions, req *ProtocolFeeControllerRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: protocolFeeController: %w", err)
	}

	calldata, err := EncodeProtocolFeeController()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: protocolFeeController: %w", err)
	}

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: protocolFeeController: %w", mapPoolManagerRevertError(err))
	}

	out, err := DecodeProtocolFeeController(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: protocolFeeController: %w", err)
	}

	return out, nil
}

// Extsload returns the raw storage word at req.Slot on req.Contract.
func (m *PoolManager) Extsload(ctx context.Context, opts *bind.CallOptions, req *ExtsloadRequest) ([32]byte, error) {
	if err := req.Validate(); err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv4: extsload: %w", err)
	}

	calldata, err := EncodeExtsload(req.Slot)
	if err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv4: extsload: %w", err)
	}

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv4: extsload: %w", mapPoolManagerRevertError(err))
	}

	out, err := DecodeExtsload(data)
	if err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv4: extsload: %w", err)
	}

	return out, nil
}

// ExtsloadRange returns the req.NSlots raw storage words starting at req.StartSlot on req.Contract.
func (m *PoolManager) ExtsloadRange(ctx context.Context, opts *bind.CallOptions, req *ExtsloadRangeRequest) ([][32]byte, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: extsloadRange: %w", err)
	}

	calldata, err := EncodeExtsloadRange(req.StartSlot, req.NSlots)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: extsloadRange: %w", err)
	}

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: extsloadRange: %w", mapPoolManagerRevertError(err))
	}

	out, err := DecodeExtsloadRange(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: extsloadRange: %w", err)
	}

	return out, nil
}

// ExtsloadSlots returns the raw storage word at each of req.Slots on req.Contract.
func (m *PoolManager) ExtsloadSlots(ctx context.Context, opts *bind.CallOptions, req *ExtsloadSlotsRequest) ([][32]byte, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: extsloadSlots: %w", err)
	}

	calldata, err := EncodeExtsloadSlots(req.Slots)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: extsloadSlots: %w", err)
	}

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: extsloadSlots: %w", mapPoolManagerRevertError(err))
	}

	out, err := DecodeExtsloadSlots(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: extsloadSlots: %w", err)
	}

	return out, nil
}

// Exttload returns the raw transient-storage word at req.Slot on req.Contract; outside a transaction it is always zero.
func (m *PoolManager) Exttload(ctx context.Context, opts *bind.CallOptions, req *ExttloadRequest) ([32]byte, error) {
	if err := req.Validate(); err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv4: exttload: %w", err)
	}

	calldata, err := EncodeExttload(req.Slot)
	if err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv4: exttload: %w", err)
	}

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv4: exttload: %w", mapPoolManagerRevertError(err))
	}

	out, err := DecodeExttload(data)
	if err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv4: exttload: %w", err)
	}

	return out, nil
}

// ExttloadSlots returns the raw transient-storage word at each of req.Slots on req.Contract.
func (m *PoolManager) ExttloadSlots(ctx context.Context, opts *bind.CallOptions, req *ExttloadSlotsRequest) ([][32]byte, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: exttloadSlots: %w", err)
	}

	calldata, err := EncodeExttloadSlots(req.Slots)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: exttloadSlots: %w", err)
	}

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: exttloadSlots: %w", mapPoolManagerRevertError(err))
	}

	out, err := DecodeExttloadSlots(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: exttloadSlots: %w", err)
	}

	return out, nil
}

// CallInitialize runs initialize(key, sqrtPriceX96) via eth_call as opts.From, returning the pool's initial tick or the revert.
func (m *PoolManager) CallInitialize(ctx context.Context, opts *bind.CallOptions, req *InitializeRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: callInitialize: %w", err)
	}

	calldata, err := EncodeInitialize(req.Key, req.SqrtPriceX96)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callInitialize: %w", err)
	}

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callInitialize: %w", mapPoolManagerRevertError(err))
	}

	out, err := DecodeInitialize(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callInitialize: %w", err)
	}

	return out, nil
}

// CallModifyLiquidity runs modifyLiquidity via eth_call as opts.From, returning (callerDelta, feesAccrued) or the revert.
func (m *PoolManager) CallModifyLiquidity(ctx context.Context, opts *bind.CallOptions, req *ModifyLiquidityRequest) (*BalanceDelta, *BalanceDelta, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: callModifyLiquidity: %w", err)
	}

	calldata, err := EncodeModifyLiquidity(req.Key, req.Params, req.HookData)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: callModifyLiquidity: %w", err)
	}

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: callModifyLiquidity: %w", mapPoolManagerRevertError(err))
	}

	callerDelta, feesAccrued, err := DecodeModifyLiquidity(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: callModifyLiquidity: %w", err)
	}

	return callerDelta, feesAccrued, nil
}

// CallSwap runs swap via eth_call as opts.From, returning the swap's BalanceDelta or the revert.
func (m *PoolManager) CallSwap(ctx context.Context, opts *bind.CallOptions, req *SwapRequest) (*BalanceDelta, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: callSwap: %w", err)
	}

	calldata, err := EncodeSwap(req.Key, req.Params, req.HookData)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callSwap: %w", err)
	}

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callSwap: %w", mapPoolManagerRevertError(err))
	}

	out, err := DecodeSwap(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callSwap: %w", err)
	}

	return out, nil
}

// CallDonate runs donate via eth_call as opts.From, returning the donation's BalanceDelta or the revert.
func (m *PoolManager) CallDonate(ctx context.Context, opts *bind.CallOptions, req *DonateRequest) (*BalanceDelta, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: callDonate: %w", err)
	}

	calldata, err := EncodeDonate(req.Key, req.Amount0, req.Amount1, req.HookData)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callDonate: %w", err)
	}

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callDonate: %w", mapPoolManagerRevertError(err))
	}

	out, err := DecodeDonate(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callDonate: %w", err)
	}

	return out, nil
}

// CallUnlock runs unlock(data) via eth_call as opts.From, returning what opts.From's unlockCallback returned or the revert.
func (m *PoolManager) CallUnlock(ctx context.Context, opts *bind.CallOptions, req *UnlockRequest) ([]byte, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: callUnlock: %w", err)
	}

	calldata, err := EncodeUnlock(req.Data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callUnlock: %w", err)
	}

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callUnlock: %w", mapPoolManagerRevertError(err))
	}

	out, err := DecodeUnlock(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callUnlock: %w", err)
	}

	return out, nil
}

// CallSettle runs settle() via eth_call as opts.From with req.NativeAmount as msg.value, returning the amount paid or the revert; opts.Value must be nil.
func (m *PoolManager) CallSettle(ctx context.Context, opts *bind.CallOptions, req *SettleRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: callSettle: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, fmt.Errorf("uniswapv4: callSettle: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeSettle()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callSettle: %w", err)
	}

	callOpts := bind.NewCallOptions()
	if opts != nil {
		*callOpts = *opts
	}
	callOpts.Value = req.NativeAmount

	data, err := bind.Call(ctx, m.c, callOpts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callSettle: %w", mapPoolManagerRevertError(err))
	}

	out, err := DecodeSettle(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callSettle: %w", err)
	}

	return out, nil
}

// CallSettleFor runs settleFor(recipient) via eth_call as opts.From with req.NativeAmount as msg.value, returning the amount paid or the revert; opts.Value must be nil.
func (m *PoolManager) CallSettleFor(ctx context.Context, opts *bind.CallOptions, req *SettleForRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: callSettleFor: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, fmt.Errorf("uniswapv4: callSettleFor: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeSettleFor(req.Recipient)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callSettleFor: %w", err)
	}

	callOpts := bind.NewCallOptions()
	if opts != nil {
		*callOpts = *opts
	}
	callOpts.Value = req.NativeAmount

	data, err := bind.Call(ctx, m.c, callOpts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callSettleFor: %w", mapPoolManagerRevertError(err))
	}

	out, err := DecodeSettleFor(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callSettleFor: %w", err)
	}

	return out, nil
}

// CallCollectProtocolFees runs collectProtocolFees via eth_call as opts.From, returning the amount it would collect or the revert.
func (m *PoolManager) CallCollectProtocolFees(ctx context.Context, opts *bind.CallOptions, req *CollectProtocolFeesRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv4: callCollectProtocolFees: %w", err)
	}

	calldata, err := EncodeCollectProtocolFees(req.Recipient, req.Currency, req.Amount)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callCollectProtocolFees: %w", err)
	}

	data, err := bind.Call(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callCollectProtocolFees: %w", mapPoolManagerRevertError(err))
	}

	out, err := DecodeCollectProtocolFees(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: callCollectProtocolFees: %w", err)
	}

	return out, nil
}

// Initialize sends initialize(key, sqrtPriceX96) on req.Contract, signed by opts.Signer; the new pool comes back via ExtractInitialize.
func (m *PoolManager) Initialize(ctx context.Context, opts *bind.TransactOptions, req *InitializeRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: initialize: %w", err)
	}

	calldata, err := EncodeInitialize(req.Key, req.SqrtPriceX96)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: initialize: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: initialize: %w", mapPoolManagerRevertError(err))
	}

	return signed, hash, nil
}

// ModifyLiquidity sends modifyLiquidity(key, params, hookData) on req.Contract, signed by opts.Signer; from an EOA it reverts with ErrManagerLocked.
func (m *PoolManager) ModifyLiquidity(ctx context.Context, opts *bind.TransactOptions, req *ModifyLiquidityRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: modifyLiquidity: %w", err)
	}

	calldata, err := EncodeModifyLiquidity(req.Key, req.Params, req.HookData)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: modifyLiquidity: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: modifyLiquidity: %w", mapPoolManagerRevertError(err))
	}

	return signed, hash, nil
}

// Swap sends swap(key, params, hookData) on req.Contract, signed by opts.Signer; from an EOA it reverts with ErrManagerLocked.
func (m *PoolManager) Swap(ctx context.Context, opts *bind.TransactOptions, req *SwapRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: swap: %w", err)
	}

	calldata, err := EncodeSwap(req.Key, req.Params, req.HookData)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: swap: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: swap: %w", mapPoolManagerRevertError(err))
	}

	return signed, hash, nil
}

// Donate sends donate(key, amount0, amount1, hookData) on req.Contract, signed by opts.Signer; from an EOA it reverts with ErrManagerLocked.
func (m *PoolManager) Donate(ctx context.Context, opts *bind.TransactOptions, req *DonateRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: donate: %w", err)
	}

	calldata, err := EncodeDonate(req.Key, req.Amount0, req.Amount1, req.HookData)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: donate: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: donate: %w", mapPoolManagerRevertError(err))
	}

	return signed, hash, nil
}

// Unlock sends unlock(data) on req.Contract, signed by opts.Signer; PoolManager calls back into the sender, so an EOA sender reverts.
func (m *PoolManager) Unlock(ctx context.Context, opts *bind.TransactOptions, req *UnlockRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: unlock: %w", err)
	}

	calldata, err := EncodeUnlock(req.Data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: unlock: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: unlock: %w", mapPoolManagerRevertError(err))
	}

	return signed, hash, nil
}

// Sync sends sync(currency) on req.Contract, signed by opts.Signer, checkpointing its reserves of req.Currency before a settle.
func (m *PoolManager) Sync(ctx context.Context, opts *bind.TransactOptions, req *SyncRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: sync: %w", err)
	}

	calldata, err := EncodeSync(req.Currency)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: sync: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: sync: %w", mapPoolManagerRevertError(err))
	}

	return signed, hash, nil
}

// Take sends take(currency, to, amount) on req.Contract, signed by opts.Signer; from an EOA it reverts with ErrManagerLocked.
func (m *PoolManager) Take(ctx context.Context, opts *bind.TransactOptions, req *TakeRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: take: %w", err)
	}

	calldata, err := EncodeTake(req.Currency, req.To, req.Amount)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: take: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: take: %w", mapPoolManagerRevertError(err))
	}

	return signed, hash, nil
}

// Settle sends settle() on req.Contract, signed by opts.Signer with req.NativeAmount (nil sends none) as msg.value; opts.Value must be nil.
func (m *PoolManager) Settle(ctx context.Context, opts *bind.TransactOptions, req *SettleRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: settle: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv4: settle: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeSettle()
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: settle: %w", err)
	}

	var txOpts *bind.TransactOptions
	if opts != nil {
		o := *opts
		o.Value = req.NativeAmount
		txOpts = &o
	}

	signed, hash, err := bind.Transact(ctx, m.c, txOpts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: settle: %w", mapPoolManagerRevertError(err))
	}

	return signed, hash, nil
}

// SettleFor sends settleFor(recipient) on req.Contract, signed by opts.Signer with req.NativeAmount (nil sends none) as msg.value; opts.Value must be nil.
func (m *PoolManager) SettleFor(ctx context.Context, opts *bind.TransactOptions, req *SettleForRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: settleFor: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv4: settleFor: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeSettleFor(req.Recipient)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: settleFor: %w", err)
	}

	var txOpts *bind.TransactOptions
	if opts != nil {
		o := *opts
		o.Value = req.NativeAmount
		txOpts = &o
	}

	signed, hash, err := bind.Transact(ctx, m.c, txOpts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: settleFor: %w", mapPoolManagerRevertError(err))
	}

	return signed, hash, nil
}

// Clear sends clear(currency, amount) on req.Contract, signed by opts.Signer; from an EOA it reverts with ErrManagerLocked.
func (m *PoolManager) Clear(ctx context.Context, opts *bind.TransactOptions, req *ClearRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: clear: %w", err)
	}

	calldata, err := EncodeClear(req.Currency, req.Amount)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: clear: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: clear: %w", mapPoolManagerRevertError(err))
	}

	return signed, hash, nil
}

// Mint sends mint(to, id, amount) on req.Contract, signed by opts.Signer -- the ERC6909 claim-token mint; from an EOA it reverts with ErrManagerLocked.
func (m *PoolManager) Mint(ctx context.Context, opts *bind.TransactOptions, req *MintRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: mint: %w", err)
	}

	calldata, err := EncodeMint(req.To, req.ID, req.Amount)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: mint: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: mint: %w", mapPoolManagerRevertError(err))
	}

	return signed, hash, nil
}

// Burn sends burn(from, id, amount) on req.Contract, signed by opts.Signer -- the ERC6909 claim-token burn; from an EOA it reverts with ErrManagerLocked.
func (m *PoolManager) Burn(ctx context.Context, opts *bind.TransactOptions, req *BurnRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: burn: %w", err)
	}

	calldata, err := EncodeBurn(req.From, req.ID, req.Amount)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: burn: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: burn: %w", mapPoolManagerRevertError(err))
	}

	return signed, hash, nil
}

// UpdateDynamicLPFee sends updateDynamicLPFee(key, newDynamicLPFee) on req.Contract, signed by opts.Signer, which must be key.Hooks or it reverts with ErrUnauthorizedDynamicLPFeeUpdate.
func (m *PoolManager) UpdateDynamicLPFee(ctx context.Context, opts *bind.TransactOptions, req *UpdateDynamicLPFeeRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: updateDynamicLPFee: %w", err)
	}

	calldata, err := EncodeUpdateDynamicLPFee(req.Key, req.NewDynamicLPFee)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: updateDynamicLPFee: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: updateDynamicLPFee: %w", mapPoolManagerRevertError(err))
	}

	return signed, hash, nil
}

// SetProtocolFee sends setProtocolFee(key, newProtocolFee) on req.Contract, signed by opts.Signer, which must be the protocol fee controller or it reverts with ErrInvalidCaller.
func (m *PoolManager) SetProtocolFee(ctx context.Context, opts *bind.TransactOptions, req *SetProtocolFeeRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: setProtocolFee: %w", err)
	}

	calldata, err := EncodeSetProtocolFee(req.Key, req.NewProtocolFee)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: setProtocolFee: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: setProtocolFee: %w", mapPoolManagerRevertError(err))
	}

	return signed, hash, nil
}

// SetProtocolFeeController sends setProtocolFeeController(controller) on req.Contract, signed by opts.Signer, which must be the owner or it reverts.
func (m *PoolManager) SetProtocolFeeController(ctx context.Context, opts *bind.TransactOptions, req *SetProtocolFeeControllerRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: setProtocolFeeController: %w", err)
	}

	calldata, err := EncodeSetProtocolFeeController(req.Controller)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: setProtocolFeeController: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: setProtocolFeeController: %w", mapPoolManagerRevertError(err))
	}

	return signed, hash, nil
}

// CollectProtocolFees sends collectProtocolFees(recipient, currency, amount) on req.Contract, signed by opts.Signer, which must be the protocol fee controller or it reverts with ErrInvalidCaller.
func (m *PoolManager) CollectProtocolFees(ctx context.Context, opts *bind.TransactOptions, req *CollectProtocolFeesRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: collectProtocolFees: %w", err)
	}

	calldata, err := EncodeCollectProtocolFees(req.Recipient, req.Currency, req.Amount)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv4: collectProtocolFees: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, m.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv4: collectProtocolFees: %w", mapPoolManagerRevertError(err))
	}

	return signed, hash, nil
}
