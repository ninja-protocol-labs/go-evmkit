package v4

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

type (
	Currency = *types.Address
	PoolID   = [32]byte
)

// PoolKey mirrors PoolKey.sol and identifies a pool on the singleton PoolManager; a zero Currency is native ETH.
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

// PoolID returns k's PoolId, keccak256(abi.encode(k)), matching PoolIdLibrary.toId.
func (k *PoolKey) PoolID() (PoolID, error) {
	encoded, err := abi.Pack(abi.Types{poolKeyTuple}, []any{k.Currency0, k.Currency1, k.Fee, k.TickSpacing, k.Hooks})
	if err != nil {
		return PoolID{}, fmt.Errorf("uniswapv4: poolKey.id: %w", err)
	}
	var id PoolID
	copy(id[:], core.Keccak256(encoded).Bytes())
	return id, nil
}

// BalanceDelta mirrors BalanceDelta.sol: two int128s packed into one int256, Amount0 in the upper half.
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

// LiquidityParams mirrors IPoolManager's ModifyLiquidityParams tuple.
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

// SwapAmountParams mirrors IPoolManager's SwapParams tuple, carried by SwapRequest.Params.
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

// ProtocolFeesAccruedRequest is protocolFeesAccrued(currency)'s input on Contract.
type ProtocolFeesAccruedRequest struct {
	Contract *types.Address
	Currency Currency
}

func NewProtocolFeesAccruedRequest(contract *types.Address, currency Currency) *ProtocolFeesAccruedRequest {
	return &ProtocolFeesAccruedRequest{
		Contract: contract,
		Currency: currency,
	}
}

func (r *ProtocolFeesAccruedRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Currency == nil {
		return fmt.Errorf("%w: currency is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// ProtocolFeeControllerRequest is protocolFeeController()'s input on Contract.
type ProtocolFeeControllerRequest struct {
	Contract *types.Address
}

func NewProtocolFeeControllerRequest(contract *types.Address) *ProtocolFeeControllerRequest {
	return &ProtocolFeeControllerRequest{
		Contract: contract,
	}
}

func (r *ProtocolFeeControllerRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// ExtsloadRequest is extsload(slot)'s input on Contract.
type ExtsloadRequest struct {
	Contract *types.Address
	Slot     [32]byte
}

func NewExtsloadRequest(contract *types.Address, slot [32]byte) *ExtsloadRequest {
	return &ExtsloadRequest{
		Contract: contract,
		Slot:     slot,
	}
}

func (r *ExtsloadRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// ExtsloadRangeRequest is extsload(startSlot, nSlots)'s input on Contract.
type ExtsloadRangeRequest struct {
	Contract  *types.Address
	StartSlot [32]byte
	NSlots    *big.Int
}

func NewExtsloadRangeRequest(contract *types.Address, startSlot [32]byte, nSlots *big.Int) *ExtsloadRangeRequest {
	return &ExtsloadRangeRequest{
		Contract:  contract,
		StartSlot: startSlot,
		NSlots:    nSlots,
	}
}

func (r *ExtsloadRangeRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.NSlots == nil {
		return fmt.Errorf("%w: nSlots is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// ExtsloadSlotsRequest is extsload(slots)'s input on Contract.
type ExtsloadSlotsRequest struct {
	Contract *types.Address
	Slots    [][32]byte
}

func NewExtsloadSlotsRequest(contract *types.Address, slots [][32]byte) *ExtsloadSlotsRequest {
	return &ExtsloadSlotsRequest{
		Contract: contract,
		Slots:    slots,
	}
}

func (r *ExtsloadSlotsRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// ExttloadRequest is exttload(slot)'s input on Contract.
type ExttloadRequest struct {
	Contract *types.Address
	Slot     [32]byte
}

func NewExttloadRequest(contract *types.Address, slot [32]byte) *ExttloadRequest {
	return &ExttloadRequest{
		Contract: contract,
		Slot:     slot,
	}
}

func (r *ExttloadRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// ExttloadSlotsRequest is exttload(slots)'s input on Contract.
type ExttloadSlotsRequest struct {
	Contract *types.Address
	Slots    [][32]byte
}

func NewExttloadSlotsRequest(contract *types.Address, slots [][32]byte) *ExttloadSlotsRequest {
	return &ExttloadSlotsRequest{
		Contract: contract,
		Slots:    slots,
	}
}

func (r *ExttloadSlotsRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// InitializeRequest is initialize(key, sqrtPriceX96)'s input on Contract.
type InitializeRequest struct {
	Contract     *types.Address
	Key          *PoolKey
	SqrtPriceX96 *big.Int
}

func NewInitializeRequest(contract *types.Address, key *PoolKey, sqrtPriceX96 *big.Int) *InitializeRequest {
	return &InitializeRequest{
		Contract:     contract,
		Key:          key,
		SqrtPriceX96: sqrtPriceX96,
	}
}

func (r *InitializeRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if err := validatePoolKey(r.Key); err != nil {
		return err
	}

	if r.SqrtPriceX96 == nil {
		return fmt.Errorf("%w: sqrtPriceX96 is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// ModifyLiquidityRequest is modifyLiquidity(key, params, hookData)'s input on Contract.
type ModifyLiquidityRequest struct {
	Contract *types.Address
	Key      *PoolKey
	Params   *LiquidityParams
	HookData []byte
}

func NewModifyLiquidityRequest(contract *types.Address, key *PoolKey, params *LiquidityParams, hookData []byte) *ModifyLiquidityRequest {
	return &ModifyLiquidityRequest{
		Contract: contract,
		Key:      key,
		Params:   params,
		HookData: hookData,
	}
}

func (r *ModifyLiquidityRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if err := validatePoolKey(r.Key); err != nil {
		return err
	}

	if err := validateLiquidityParams(r.Params); err != nil {
		return err
	}

	return nil
}

// SwapRequest is swap(key, params, hookData)'s input on Contract.
type SwapRequest struct {
	Contract *types.Address
	Key      *PoolKey
	Params   *SwapAmountParams
	HookData []byte
}

func NewSwapRequest(contract *types.Address, key *PoolKey, params *SwapAmountParams, hookData []byte) *SwapRequest {
	return &SwapRequest{
		Contract: contract,
		Key:      key,
		Params:   params,
		HookData: hookData,
	}
}

func (r *SwapRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if err := validatePoolKey(r.Key); err != nil {
		return err
	}

	if err := validateSwapAmountParams(r.Params); err != nil {
		return err
	}

	return nil
}

// DonateRequest is donate(key, amount0, amount1, hookData)'s input on Contract.
type DonateRequest struct {
	Contract *types.Address
	Key      *PoolKey
	Amount0  *big.Int
	Amount1  *big.Int
	HookData []byte
}

func NewDonateRequest(contract *types.Address, key *PoolKey, amount0, amount1 *big.Int, hookData []byte) *DonateRequest {
	return &DonateRequest{
		Contract: contract,
		Key:      key,
		Amount0:  amount0,
		Amount1:  amount1,
		HookData: hookData,
	}
}

func (r *DonateRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if err := validatePoolKey(r.Key); err != nil {
		return err
	}

	if r.Amount0 == nil {
		return fmt.Errorf("%w: amount0 is nil", bind.ErrInvalidRequest)
	}

	if r.Amount1 == nil {
		return fmt.Errorf("%w: amount1 is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// UnlockRequest is unlock(data)'s input on Contract; Data is handed to the caller's unlockCallback.
type UnlockRequest struct {
	Contract *types.Address
	Data     []byte
}

func NewUnlockRequest(contract *types.Address, data []byte) *UnlockRequest {
	return &UnlockRequest{
		Contract: contract,
		Data:     data,
	}
}

func (r *UnlockRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SyncRequest is sync(currency)'s input on Contract.
type SyncRequest struct {
	Contract *types.Address
	Currency Currency
}

func NewSyncRequest(contract *types.Address, currency Currency) *SyncRequest {
	return &SyncRequest{
		Contract: contract,
		Currency: currency,
	}
}

func (r *SyncRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Currency == nil {
		return fmt.Errorf("%w: currency is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// TakeRequest is take(currency, to, amount)'s input on Contract.
type TakeRequest struct {
	Contract *types.Address
	Currency Currency
	To       *types.Address
	Amount   *big.Int
}

func NewTakeRequest(contract *types.Address, currency Currency, to *types.Address, amount *big.Int) *TakeRequest {
	return &TakeRequest{
		Contract: contract,
		Currency: currency,
		To:       to,
		Amount:   amount,
	}
}

func (r *TakeRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Currency == nil {
		return fmt.Errorf("%w: currency is nil", bind.ErrInvalidRequest)
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	if r.Amount == nil {
		return fmt.Errorf("%w: amount is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SettleRequest is settle()'s input on Contract; NativeAmount, if set, is sent as msg.value.
type SettleRequest struct {
	Contract     *types.Address
	NativeAmount *big.Int
}

func NewSettleRequest(contract *types.Address, nativeAmount *big.Int) *SettleRequest {
	return &SettleRequest{
		Contract:     contract,
		NativeAmount: nativeAmount,
	}
}

func (r *SettleRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.NativeAmount != nil && r.NativeAmount.Sign() < 0 {
		return fmt.Errorf("%w: nativeAmount is negative", bind.ErrInvalidRequest)
	}

	if r.NativeAmount != nil && r.NativeAmount.BitLen() > 256 {
		return fmt.Errorf("%w: nativeAmount exceeds uint256", bind.ErrInvalidRequest)
	}

	return nil
}

// SettleForRequest is settleFor(recipient)'s input on Contract; NativeAmount, if set, is sent as msg.value.
type SettleForRequest struct {
	Contract     *types.Address
	Recipient    *types.Address
	NativeAmount *big.Int
}

func NewSettleForRequest(contract, recipient *types.Address, nativeAmount *big.Int) *SettleForRequest {
	return &SettleForRequest{
		Contract:     contract,
		Recipient:    recipient,
		NativeAmount: nativeAmount,
	}
}

func (r *SettleForRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	if r.NativeAmount != nil && r.NativeAmount.Sign() < 0 {
		return fmt.Errorf("%w: nativeAmount is negative", bind.ErrInvalidRequest)
	}

	if r.NativeAmount != nil && r.NativeAmount.BitLen() > 256 {
		return fmt.Errorf("%w: nativeAmount exceeds uint256", bind.ErrInvalidRequest)
	}

	return nil
}

// ClearRequest is clear(currency, amount)'s input on Contract.
type ClearRequest struct {
	Contract *types.Address
	Currency Currency
	Amount   *big.Int
}

func NewClearRequest(contract *types.Address, currency Currency, amount *big.Int) *ClearRequest {
	return &ClearRequest{
		Contract: contract,
		Currency: currency,
		Amount:   amount,
	}
}

func (r *ClearRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Currency == nil {
		return fmt.Errorf("%w: currency is nil", bind.ErrInvalidRequest)
	}

	if r.Amount == nil {
		return fmt.Errorf("%w: amount is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// MintRequest is mint(to, id, amount)'s input on Contract -- the ERC6909 claim-token mint, not a liquidity position.
type MintRequest struct {
	Contract *types.Address
	To       *types.Address
	ID       *big.Int
	Amount   *big.Int
}

func NewMintRequest(contract, to *types.Address, id, amount *big.Int) *MintRequest {
	return &MintRequest{
		Contract: contract,
		To:       to,
		ID:       id,
		Amount:   amount,
	}
}

func (r *MintRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	if r.ID == nil {
		return fmt.Errorf("%w: id is nil", bind.ErrInvalidRequest)
	}

	if r.Amount == nil {
		return fmt.Errorf("%w: amount is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// BurnRequest is burn(from, id, amount)'s input on Contract -- the ERC6909 claim-token burn.
type BurnRequest struct {
	Contract *types.Address
	From     *types.Address
	ID       *big.Int
	Amount   *big.Int
}

func NewBurnRequest(contract, from *types.Address, id, amount *big.Int) *BurnRequest {
	return &BurnRequest{
		Contract: contract,
		From:     from,
		ID:       id,
		Amount:   amount,
	}
}

func (r *BurnRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.From == nil {
		return fmt.Errorf("%w: from is nil", bind.ErrInvalidRequest)
	}

	if r.ID == nil {
		return fmt.Errorf("%w: id is nil", bind.ErrInvalidRequest)
	}

	if r.Amount == nil {
		return fmt.Errorf("%w: amount is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// UpdateDynamicLPFeeRequest is updateDynamicLPFee(key, newDynamicLPFee)'s input on Contract.
type UpdateDynamicLPFeeRequest struct {
	Contract        *types.Address
	Key             *PoolKey
	NewDynamicLPFee *big.Int
}

func NewUpdateDynamicLPFeeRequest(contract *types.Address, key *PoolKey, newDynamicLPFee *big.Int) *UpdateDynamicLPFeeRequest {
	return &UpdateDynamicLPFeeRequest{
		Contract:        contract,
		Key:             key,
		NewDynamicLPFee: newDynamicLPFee,
	}
}

func (r *UpdateDynamicLPFeeRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if err := validatePoolKey(r.Key); err != nil {
		return err
	}

	if r.NewDynamicLPFee == nil {
		return fmt.Errorf("%w: newDynamicLPFee is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SetProtocolFeeRequest is setProtocolFee(key, newProtocolFee)'s input on Contract.
type SetProtocolFeeRequest struct {
	Contract       *types.Address
	Key            *PoolKey
	NewProtocolFee *big.Int
}

func NewSetProtocolFeeRequest(contract *types.Address, key *PoolKey, newProtocolFee *big.Int) *SetProtocolFeeRequest {
	return &SetProtocolFeeRequest{
		Contract:       contract,
		Key:            key,
		NewProtocolFee: newProtocolFee,
	}
}

func (r *SetProtocolFeeRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if err := validatePoolKey(r.Key); err != nil {
		return err
	}

	if r.NewProtocolFee == nil {
		return fmt.Errorf("%w: newProtocolFee is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SetProtocolFeeControllerRequest is setProtocolFeeController(controller)'s input on Contract.
type SetProtocolFeeControllerRequest struct {
	Contract   *types.Address
	Controller *types.Address
}

func NewSetProtocolFeeControllerRequest(contract, controller *types.Address) *SetProtocolFeeControllerRequest {
	return &SetProtocolFeeControllerRequest{
		Contract:   contract,
		Controller: controller,
	}
}

func (r *SetProtocolFeeControllerRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Controller == nil {
		return fmt.Errorf("%w: controller is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// CollectProtocolFeesRequest is collectProtocolFees(recipient, currency, amount)'s input on Contract.
type CollectProtocolFeesRequest struct {
	Contract  *types.Address
	Recipient *types.Address
	Currency  Currency
	Amount    *big.Int
}

func NewCollectProtocolFeesRequest(contract, recipient *types.Address, currency Currency, amount *big.Int) *CollectProtocolFeesRequest {
	return &CollectProtocolFeesRequest{
		Contract:  contract,
		Recipient: recipient,
		Currency:  currency,
		Amount:    amount,
	}
}

func (r *CollectProtocolFeesRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	if r.Currency == nil {
		return fmt.Errorf("%w: currency is nil", bind.ErrInvalidRequest)
	}

	if r.Amount == nil {
		return fmt.Errorf("%w: amount is nil", bind.ErrInvalidRequest)
	}

	return nil
}

func validatePoolKey(k *PoolKey) error {
	if k == nil {
		return fmt.Errorf("%w: key is nil", bind.ErrInvalidRequest)
	}

	if k.Currency0 == nil {
		return fmt.Errorf("%w: key.currency0 is nil", bind.ErrInvalidRequest)
	}

	if k.Currency1 == nil {
		return fmt.Errorf("%w: key.currency1 is nil", bind.ErrInvalidRequest)
	}

	if k.Fee == nil {
		return fmt.Errorf("%w: key.fee is nil", bind.ErrInvalidRequest)
	}

	if k.TickSpacing == nil {
		return fmt.Errorf("%w: key.tickSpacing is nil", bind.ErrInvalidRequest)
	}

	if k.Hooks == nil {
		return fmt.Errorf("%w: key.hooks is nil", bind.ErrInvalidRequest)
	}

	return nil
}

func validateLiquidityParams(p *LiquidityParams) error {
	if p == nil {
		return fmt.Errorf("%w: params is nil", bind.ErrInvalidRequest)
	}

	if p.TickLower == nil {
		return fmt.Errorf("%w: params.tickLower is nil", bind.ErrInvalidRequest)
	}

	if p.TickUpper == nil {
		return fmt.Errorf("%w: params.tickUpper is nil", bind.ErrInvalidRequest)
	}

	if p.LiquidityDelta == nil {
		return fmt.Errorf("%w: params.liquidityDelta is nil", bind.ErrInvalidRequest)
	}

	return nil
}

func validateSwapAmountParams(p *SwapAmountParams) error {
	if p == nil {
		return fmt.Errorf("%w: params is nil", bind.ErrInvalidRequest)
	}

	if p.AmountSpecified == nil {
		return fmt.Errorf("%w: params.amountSpecified is nil", bind.ErrInvalidRequest)
	}

	if p.SqrtPriceLimitX96 == nil {
		return fmt.Errorf("%w: params.sqrtPriceLimitX96 is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// InitializeEvent mirrors IPoolManager's Initialize event: pool Id was initialized on Contract at SqrtPriceX96, landing at Tick.
type InitializeEvent struct {
	Contract     *types.Address
	Id           PoolID
	Currency0    Currency
	Currency1    Currency
	Fee          *big.Int
	TickSpacing  *big.Int
	Hooks        *types.Address
	SqrtPriceX96 *big.Int
	Tick         *big.Int
}

func NewInitializeEvent(contract *types.Address, id PoolID, currency0, currency1 Currency, fee, tickSpacing *big.Int, hooks *types.Address, sqrtPriceX96, tick *big.Int) *InitializeEvent {
	return &InitializeEvent{
		Contract:     contract,
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

// ModifyLiquidityEvent mirrors IPoolManager's ModifyLiquidity event.
type ModifyLiquidityEvent struct {
	Contract       *types.Address
	Id             PoolID
	Sender         *types.Address
	TickLower      *big.Int
	TickUpper      *big.Int
	LiquidityDelta *big.Int
	Salt           [32]byte
}

func NewModifyLiquidityEvent(contract *types.Address, id PoolID, sender *types.Address, tickLower, tickUpper, liquidityDelta *big.Int, salt [32]byte) *ModifyLiquidityEvent {
	return &ModifyLiquidityEvent{
		Contract:       contract,
		Id:             id,
		Sender:         sender,
		TickLower:      tickLower,
		TickUpper:      tickUpper,
		LiquidityDelta: liquidityDelta,
		Salt:           salt,
	}
}

// SwapEvent mirrors IPoolManager's Swap event; amounts are the pool's deltas and the rest is post-swap pool state.
type SwapEvent struct {
	Contract     *types.Address
	Id           PoolID
	Sender       *types.Address
	Amount0      *big.Int
	Amount1      *big.Int
	SqrtPriceX96 *big.Int
	Liquidity    *big.Int
	Tick         *big.Int
	Fee          *big.Int
}

func NewSwapEvent(contract *types.Address, id PoolID, sender *types.Address, amount0, amount1, sqrtPriceX96, liquidity, tick, fee *big.Int) *SwapEvent {
	return &SwapEvent{
		Contract:     contract,
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

// DonateEvent mirrors IPoolManager's Donate event.
type DonateEvent struct {
	Contract *types.Address
	Id       PoolID
	Sender   *types.Address
	Amount0  *big.Int
	Amount1  *big.Int
}

func NewDonateEvent(contract *types.Address, id PoolID, sender *types.Address, amount0, amount1 *big.Int) *DonateEvent {
	return &DonateEvent{
		Contract: contract,
		Id:       id,
		Sender:   sender,
		Amount0:  amount0,
		Amount1:  amount1,
	}
}

// ProtocolFeeControllerUpdatedEvent mirrors IProtocolFees' ProtocolFeeControllerUpdated event.
type ProtocolFeeControllerUpdatedEvent struct {
	Contract              *types.Address
	ProtocolFeeController *types.Address
}

func NewProtocolFeeControllerUpdatedEvent(contract, protocolFeeController *types.Address) *ProtocolFeeControllerUpdatedEvent {
	return &ProtocolFeeControllerUpdatedEvent{
		Contract:              contract,
		ProtocolFeeController: protocolFeeController,
	}
}

// ProtocolFeeUpdatedEvent mirrors IProtocolFees' ProtocolFeeUpdated event.
type ProtocolFeeUpdatedEvent struct {
	Contract    *types.Address
	Id          PoolID
	ProtocolFee *big.Int
}

func NewProtocolFeeUpdatedEvent(contract *types.Address, id PoolID, protocolFee *big.Int) *ProtocolFeeUpdatedEvent {
	return &ProtocolFeeUpdatedEvent{
		Contract:    contract,
		Id:          id,
		ProtocolFee: protocolFee,
	}
}
