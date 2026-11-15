// Package v4 encodes calls to and decodes results from Uniswap v4's PoolManager.
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
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc6909"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// IERC6909Claims mirrors IERC6909Claims.sol; same method set as erc6909.IERC6909, so reused directly.
type IERC6909Claims = erc6909.IERC6909

// IProtocolFees mirrors IProtocolFees.sol.
type IProtocolFees interface {
	ProtocolFeesAccrued(ctx context.Context, contract *types.Address, currency Currency, block string) (*big.Int, error)
	ProtocolFeeController(ctx context.Context, contract *types.Address, block string) (*types.Address, error)

	SetProtocolFee(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SetProtocolFeeController(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	CollectProtocolFees(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

// IExtsload mirrors IExtsload.sol; its 3 overloaded extsload(...) become distinctly-named methods.
type IExtsload interface {
	Extsload(ctx context.Context, contract *types.Address, slot [32]byte, block string) ([32]byte, error)
	ExtsloadRange(ctx context.Context, contract *types.Address, startSlot [32]byte, nSlots *big.Int, block string) ([][32]byte, error)
	ExtsloadSlots(ctx context.Context, contract *types.Address, slots [][32]byte, block string) ([][32]byte, error)
}

// IExttload mirrors IExttload.sol; its 2 overloaded exttload(...) become distinctly-named methods.
type IExttload interface {
	Exttload(ctx context.Context, contract *types.Address, slot [32]byte, block string) ([32]byte, error)
	ExttloadSlots(ctx context.Context, contract *types.Address, slots [][32]byte, block string) ([][32]byte, error)
}

// IPoolManager mirrors IPoolManager.sol: its own action surface (all state-mutating, no view
// functions of its own) plus IProtocolFees, IERC6909Claims, IExtsload, and IExttload.
type IPoolManager interface {
	IProtocolFees
	IERC6909Claims
	IExtsload
	IExttload

	Initialize(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	ModifyLiquidity(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Swap(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Donate(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Unlock(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Sync(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Take(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Settle(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SettleFor(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Clear(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Mint(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Burn(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	UpdateDynamicLPFee(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

var _ IPoolManager = (*PoolManager)(nil)

// PoolManager calls any PoolManager contract over a Client; it holds no contract address, so one
// instance serves every deployment. The embedded *erc6909.ERC6909 promotes BalanceOf/Allowance/
// IsOperator/Approve/SetOperator/Transfer/TransferFrom/SupportsInterface, satisfying
// IERC6909Claims.
type PoolManager struct {
	*erc6909.ERC6909

	c rpc.Client
}

// New returns a PoolManager calling deployments via cli.
func New(cli rpc.Client) *PoolManager {
	return &PoolManager{
		c:       cli,
		ERC6909: erc6909.New(cli),
	}
}

// ProtocolFeesAccrued returns the protocol fees accrued in currency on contract.
func (m *PoolManager) ProtocolFeesAccrued(ctx context.Context, contract *types.Address, currency Currency, block string) (*big.Int, error) {
	calldata, err := EncodeProtocolFeesAccrued(currency)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: protocolFeesAccrued: %w", err)
	}
	data, err := m.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: protocolFeesAccrued: %w", err)
	}
	amount, err := DecodeProtocolFeesAccrued(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: protocolFeesAccrued: %w", err)
	}
	return amount, nil
}

// ProtocolFeeController returns contract's current protocol fee controller address.
func (m *PoolManager) ProtocolFeeController(ctx context.Context, contract *types.Address, block string) (*types.Address, error) {
	calldata, err := EncodeProtocolFeeController()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: protocolFeeController: %w", err)
	}
	data, err := m.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: protocolFeeController: %w", err)
	}
	controller, err := DecodeProtocolFeeController(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: protocolFeeController: %w", err)
	}
	return controller, nil
}

// Extsload returns the raw value at slot on contract.
func (m *PoolManager) Extsload(ctx context.Context, contract *types.Address, slot [32]byte, block string) ([32]byte, error) {
	calldata, err := EncodeExtsload(slot)
	if err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv4: extsload: %w", err)
	}
	data, err := m.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv4: extsload: %w", err)
	}
	value, err := DecodeExtsload(data)
	if err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv4: extsload: %w", err)
	}
	return value, nil
}

// ExtsloadRange returns the nSlots raw values starting at startSlot on contract.
func (m *PoolManager) ExtsloadRange(ctx context.Context, contract *types.Address, startSlot [32]byte, nSlots *big.Int, block string) ([][32]byte, error) {
	calldata, err := EncodeExtsloadRange(startSlot, nSlots)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: extsloadRange: %w", err)
	}
	data, err := m.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: extsloadRange: %w", err)
	}
	values, err := DecodeExtsloadRange(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: extsloadRange: %w", err)
	}
	return values, nil
}

// ExtsloadSlots returns the raw value at each of slots on contract.
func (m *PoolManager) ExtsloadSlots(ctx context.Context, contract *types.Address, slots [][32]byte, block string) ([][32]byte, error) {
	calldata, err := EncodeExtsloadSlots(slots)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: extsloadSlots: %w", err)
	}
	data, err := m.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: extsloadSlots: %w", err)
	}
	values, err := DecodeExtsloadSlots(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: extsloadSlots: %w", err)
	}
	return values, nil
}

// Exttload returns the raw transient-storage value at slot on contract.
func (m *PoolManager) Exttload(ctx context.Context, contract *types.Address, slot [32]byte, block string) ([32]byte, error) {
	calldata, err := EncodeExttload(slot)
	if err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv4: exttload: %w", err)
	}
	data, err := m.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv4: exttload: %w", err)
	}
	value, err := DecodeExttload(data)
	if err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv4: exttload: %w", err)
	}
	return value, nil
}

// ExttloadSlots returns the raw transient-storage value at each of slots on contract.
func (m *PoolManager) ExttloadSlots(ctx context.Context, contract *types.Address, slots [][32]byte, block string) ([][32]byte, error) {
	calldata, err := EncodeExttloadSlots(slots)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: exttloadSlots: %w", err)
	}
	data, err := m.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: exttloadSlots: %w", err)
	}
	values, err := DecodeExttloadSlots(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: exttloadSlots: %w", err)
	}
	return values, nil
}

// CreateInitialize builds an unsigned initialize(key, sqrtPriceX96) transaction via b, sent by sender.
func (m *PoolManager) CreateInitialize(ctx context.Context, b tx.Builder, contract, sender *types.Address, key *PoolKey, sqrtPriceX96 *big.Int) (core.Transaction, error) {
	calldata, err := EncodeInitialize(key, sqrtPriceX96)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createInitialize: %w", err)
	}
	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, m.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createInitialize: %w", m.wrapRevertError(err))
	}
	return unsigned, nil
}

// Initialize broadcasts an already-signed initialize transaction.
func (m *PoolManager) Initialize(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: initialize: %w", err)
	}
	hash, err := tx.Broadcast(ctx, m.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: initialize: %w", err)
	}
	return hash, nil
}

// CreateModifyLiquidity builds an unsigned modifyLiquidity(key, params, hookData) transaction via
// b, sent by sender.
func (m *PoolManager) CreateModifyLiquidity(ctx context.Context, b tx.Builder, contract, sender *types.Address, key *PoolKey, params *LiquidityParams, hookData []byte) (core.Transaction, error) {
	calldata, err := EncodeModifyLiquidity(key, params, hookData)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createModifyLiquidity: %w", err)
	}
	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, m.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createModifyLiquidity: %w", m.wrapRevertError(err))
	}
	return unsigned, nil
}

// ModifyLiquidity broadcasts an already-signed modifyLiquidity transaction.
func (m *PoolManager) ModifyLiquidity(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: modifyLiquidity: %w", err)
	}
	hash, err := tx.Broadcast(ctx, m.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: modifyLiquidity: %w", err)
	}
	return hash, nil
}

// CreateSwap builds an unsigned swap(key, params, hookData) transaction via b, sent by sender.
func (m *PoolManager) CreateSwap(ctx context.Context, b tx.Builder, contract, sender *types.Address, key *PoolKey, params *SwapAmountParams, hookData []byte) (core.Transaction, error) {
	calldata, err := EncodeSwap(key, params, hookData)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createSwap: %w", err)
	}
	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, m.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createSwap: %w", m.wrapRevertError(err))
	}
	return unsigned, nil
}

// Swap broadcasts an already-signed swap transaction.
func (m *PoolManager) Swap(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: swap: %w", err)
	}
	hash, err := tx.Broadcast(ctx, m.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: swap: %w", err)
	}
	return hash, nil
}

// CreateDonate builds an unsigned donate(key, amount0, amount1, hookData) transaction via b, sent
// by sender.
func (m *PoolManager) CreateDonate(ctx context.Context, b tx.Builder, contract, sender *types.Address, key *PoolKey, amount0, amount1 *big.Int, hookData []byte) (core.Transaction, error) {
	calldata, err := EncodeDonate(key, amount0, amount1, hookData)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createDonate: %w", err)
	}
	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, m.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createDonate: %w", m.wrapRevertError(err))
	}
	return unsigned, nil
}

// Donate broadcasts an already-signed donate transaction.
func (m *PoolManager) Donate(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: donate: %w", err)
	}
	hash, err := tx.Broadcast(ctx, m.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: donate: %w", err)
	}
	return hash, nil
}

// CreateUnlock builds an unsigned unlock(data) transaction via b, sent by sender.
func (m *PoolManager) CreateUnlock(ctx context.Context, b tx.Builder, contract, sender *types.Address, unlockData []byte) (core.Transaction, error) {
	calldata, err := EncodeUnlock(unlockData)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createUnlock: %w", err)
	}
	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, m.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createUnlock: %w", m.wrapRevertError(err))
	}
	return unsigned, nil
}

// Unlock broadcasts an already-signed unlock transaction.
func (m *PoolManager) Unlock(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: unlock: %w", err)
	}
	hash, err := tx.Broadcast(ctx, m.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: unlock: %w", err)
	}
	return hash, nil
}

// CreateSync builds an unsigned sync(currency) transaction via b, sent by sender.
func (m *PoolManager) CreateSync(ctx context.Context, b tx.Builder, contract, sender *types.Address, currency Currency) (core.Transaction, error) {
	calldata, err := EncodeSync(currency)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createSync: %w", err)
	}
	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, m.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createSync: %w", m.wrapRevertError(err))
	}
	return unsigned, nil
}

// Sync broadcasts an already-signed sync transaction.
func (m *PoolManager) Sync(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: sync: %w", err)
	}
	hash, err := tx.Broadcast(ctx, m.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: sync: %w", err)
	}
	return hash, nil
}

// CreateTake builds an unsigned take(currency, to, amount) transaction via b, sent by sender.
func (m *PoolManager) CreateTake(ctx context.Context, b tx.Builder, contract, sender *types.Address, currency Currency, to *types.Address, amount *big.Int) (core.Transaction, error) {
	calldata, err := EncodeTake(currency, to, amount)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createTake: %w", err)
	}
	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, m.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createTake: %w", m.wrapRevertError(err))
	}
	return unsigned, nil
}

// Take broadcasts an already-signed take transaction.
func (m *PoolManager) Take(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: take: %w", err)
	}
	hash, err := tx.Broadcast(ctx, m.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: take: %w", err)
	}
	return hash, nil
}

// CreateSettle builds an unsigned settle() transaction via b, sent by sender with value wei attached.
func (m *PoolManager) CreateSettle(ctx context.Context, b tx.Builder, contract, sender *types.Address, value *big.Int) (core.Transaction, error) {
	calldata, err := EncodeSettle()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createSettle: %w", err)
	}
	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, m.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createSettle: %w", m.wrapRevertError(err))
	}
	return unsigned, nil
}

// Settle broadcasts an already-signed settle transaction.
func (m *PoolManager) Settle(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: settle: %w", err)
	}
	hash, err := tx.Broadcast(ctx, m.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: settle: %w", err)
	}
	return hash, nil
}

// CreateSettleFor builds an unsigned settleFor(recipient) transaction via b, sent by sender with
// value wei attached.
func (m *PoolManager) CreateSettleFor(ctx context.Context, b tx.Builder, contract, sender *types.Address, value *big.Int, recipient *types.Address) (core.Transaction, error) {
	calldata, err := EncodeSettleFor(recipient)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createSettleFor: %w", err)
	}
	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, m.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createSettleFor: %w", m.wrapRevertError(err))
	}
	return unsigned, nil
}

// SettleFor broadcasts an already-signed settleFor transaction.
func (m *PoolManager) SettleFor(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: settleFor: %w", err)
	}
	hash, err := tx.Broadcast(ctx, m.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: settleFor: %w", err)
	}
	return hash, nil
}

// CreateClear builds an unsigned clear(currency, amount) transaction via b, sent by sender.
func (m *PoolManager) CreateClear(ctx context.Context, b tx.Builder, contract, sender *types.Address, currency Currency, amount *big.Int) (core.Transaction, error) {
	calldata, err := EncodeClear(currency, amount)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createClear: %w", err)
	}
	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, m.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createClear: %w", m.wrapRevertError(err))
	}
	return unsigned, nil
}

// Clear broadcasts an already-signed clear transaction.
func (m *PoolManager) Clear(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: clear: %w", err)
	}
	hash, err := tx.Broadcast(ctx, m.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: clear: %w", err)
	}
	return hash, nil
}

// CreateMint builds an unsigned mint(to, id, amount) transaction via b, sent by sender -- the
// ERC6909 claim-token mint, not a liquidity position.
func (m *PoolManager) CreateMint(ctx context.Context, b tx.Builder, contract, sender, to *types.Address, id, amount *big.Int) (core.Transaction, error) {
	calldata, err := EncodeMint(to, id, amount)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createMint: %w", err)
	}
	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, m.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createMint: %w", m.wrapRevertError(err))
	}
	return unsigned, nil
}

// Mint broadcasts an already-signed mint transaction.
func (m *PoolManager) Mint(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: mint: %w", err)
	}
	hash, err := tx.Broadcast(ctx, m.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: mint: %w", err)
	}
	return hash, nil
}

// CreateBurn builds an unsigned burn(from, id, amount) transaction via b, sent by sender.
func (m *PoolManager) CreateBurn(ctx context.Context, b tx.Builder, contract, sender, from *types.Address, id, amount *big.Int) (core.Transaction, error) {
	calldata, err := EncodeBurn(from, id, amount)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createBurn: %w", err)
	}
	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, m.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createBurn: %w", m.wrapRevertError(err))
	}
	return unsigned, nil
}

// Burn broadcasts an already-signed burn transaction.
func (m *PoolManager) Burn(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: burn: %w", err)
	}
	hash, err := tx.Broadcast(ctx, m.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: burn: %w", err)
	}
	return hash, nil
}

// CreateUpdateDynamicLPFee builds an unsigned updateDynamicLPFee(key, newDynamicLPFee) transaction
// via b, sent by sender.
func (m *PoolManager) CreateUpdateDynamicLPFee(ctx context.Context, b tx.Builder, contract, sender *types.Address, key *PoolKey, newDynamicLPFee *big.Int) (core.Transaction, error) {
	calldata, err := EncodeUpdateDynamicLPFee(key, newDynamicLPFee)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createUpdateDynamicLPFee: %w", err)
	}
	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, m.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createUpdateDynamicLPFee: %w", m.wrapRevertError(err))
	}
	return unsigned, nil
}

// UpdateDynamicLPFee broadcasts an already-signed updateDynamicLPFee transaction.
func (m *PoolManager) UpdateDynamicLPFee(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: updateDynamicLPFee: %w", err)
	}
	hash, err := tx.Broadcast(ctx, m.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: updateDynamicLPFee: %w", err)
	}
	return hash, nil
}

// CreateSetProtocolFee builds an unsigned setProtocolFee(key, newProtocolFee) transaction via b,
// sent by sender.
func (m *PoolManager) CreateSetProtocolFee(ctx context.Context, b tx.Builder, contract, sender *types.Address, key *PoolKey, newProtocolFee *big.Int) (core.Transaction, error) {
	calldata, err := EncodeSetProtocolFee(key, newProtocolFee)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createSetProtocolFee: %w", err)
	}
	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, m.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createSetProtocolFee: %w", m.wrapRevertError(err))
	}
	return unsigned, nil
}

// SetProtocolFee broadcasts an already-signed setProtocolFee transaction.
func (m *PoolManager) SetProtocolFee(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: setProtocolFee: %w", err)
	}
	hash, err := tx.Broadcast(ctx, m.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: setProtocolFee: %w", err)
	}
	return hash, nil
}

// CreateSetProtocolFeeController builds an unsigned setProtocolFeeController(controller)
// transaction via b, sent by sender.
func (m *PoolManager) CreateSetProtocolFeeController(ctx context.Context, b tx.Builder, contract, sender, controller *types.Address) (core.Transaction, error) {
	calldata, err := EncodeSetProtocolFeeController(controller)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createSetProtocolFeeController: %w", err)
	}
	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, m.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createSetProtocolFeeController: %w", m.wrapRevertError(err))
	}
	return unsigned, nil
}

// SetProtocolFeeController broadcasts an already-signed setProtocolFeeController transaction.
func (m *PoolManager) SetProtocolFeeController(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: setProtocolFeeController: %w", err)
	}
	hash, err := tx.Broadcast(ctx, m.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: setProtocolFeeController: %w", err)
	}
	return hash, nil
}

// CreateCollectProtocolFees builds an unsigned collectProtocolFees(recipient, currency, amount)
// transaction via b, sent by sender.
func (m *PoolManager) CreateCollectProtocolFees(ctx context.Context, b tx.Builder, contract, sender, recipient *types.Address, currency Currency, amount *big.Int) (core.Transaction, error) {
	calldata, err := EncodeCollectProtocolFees(recipient, currency, amount)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createCollectProtocolFees: %w", err)
	}
	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, m.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createCollectProtocolFees: %w", m.wrapRevertError(err))
	}
	return unsigned, nil
}

// CollectProtocolFees broadcasts an already-signed collectProtocolFees transaction.
func (m *PoolManager) CollectProtocolFees(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: collectProtocolFees: %w", err)
	}
	hash, err := tx.Broadcast(ctx, m.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: collectProtocolFees: %w", err)
	}
	return hash, nil
}

// call eth_calls contract with calldata (from msg.sender from, if non-nil), decoding a reverted
// custom error to its sentinel.
func (m *PoolManager) call(ctx context.Context, from, contract *types.Address, calldata []byte, block string) ([]byte, error) {
	params := map[string]any{
		"to":   contract.String(),
		"data": encoding.Hex.EncodePrefixed(calldata),
	}
	if from != nil {
		params["from"] = from.String()
	}

	var raw string
	if err := m.c.Call(ctx, rpc.ETHCall("call", params, block, &raw)); err != nil {
		return nil, m.wrapRevertError(err)
	}
	return encoding.Hex.Decode(raw)
}

// wrapRevertError maps a known IPoolManager/IProtocolFees custom error in err to its sentinel.
func (_ *PoolManager) wrapRevertError(err error) error {
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
	case errCurrencyNotSettled.Selector():
		return fmt.Errorf("%w: %w", ErrCurrencyNotSettled, err)
	case errPoolNotInitialized.Selector():
		return fmt.Errorf("%w: %w", ErrPoolNotInitialized, err)
	case errAlreadyUnlocked.Selector():
		return fmt.Errorf("%w: %w", ErrAlreadyUnlocked, err)
	case errManagerLocked.Selector():
		return fmt.Errorf("%w: %w", ErrManagerLocked, err)
	case errTickSpacingTooLarge.Selector():
		return fmt.Errorf("%w: %w", ErrTickSpacingTooLarge, err)
	case errTickSpacingTooSmall.Selector():
		return fmt.Errorf("%w: %w", ErrTickSpacingTooSmall, err)
	case errCurrenciesOutOfOrderOrEqual.Selector():
		return fmt.Errorf("%w: %w", ErrCurrenciesOutOfOrderOrEqual, err)
	case errUnauthorizedDynamicLPFeeUpdate.Selector():
		return fmt.Errorf("%w: %w", ErrUnauthorizedDynamicLPFeeUpdate, err)
	case errSwapAmountCannotBeZero.Selector():
		return fmt.Errorf("%w: %w", ErrSwapAmountCannotBeZero, err)
	case errNonzeroNativeValue.Selector():
		return fmt.Errorf("%w: %w", ErrNonzeroNativeValue, err)
	case errMustClearExactPositiveDelta.Selector():
		return fmt.Errorf("%w: %w", ErrMustClearExactPositiveDelta, err)
	case errProtocolFeeTooLarge.Selector():
		return fmt.Errorf("%w: %w", ErrProtocolFeeTooLarge, err)
	case errInvalidCaller.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidCaller, err)
	case errProtocolFeeCurrencySynced.Selector():
		return fmt.Errorf("%w: %w", ErrProtocolFeeCurrencySynced, err)
	default:
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}
}
