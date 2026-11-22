package v2

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// MinimumLiquidityRequest is MINIMUM_LIQUIDITY()'s input on Contract.
type MinimumLiquidityRequest struct {
	Contract *types.Address
}

func NewMinimumLiquidityRequest(contract *types.Address) *MinimumLiquidityRequest {
	return &MinimumLiquidityRequest{
		Contract: contract,
	}
}

func (r *MinimumLiquidityRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// FactoryRequest is factory()'s input on Contract.
type FactoryRequest struct {
	Contract *types.Address
}

func NewFactoryRequest(contract *types.Address) *FactoryRequest {
	return &FactoryRequest{
		Contract: contract,
	}
}

func (r *FactoryRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// Token0Request is token0()'s input on Contract.
type Token0Request struct {
	Contract *types.Address
}

func NewToken0Request(contract *types.Address) *Token0Request {
	return &Token0Request{
		Contract: contract,
	}
}

func (r *Token0Request) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// Token1Request is token1()'s input on Contract.
type Token1Request struct {
	Contract *types.Address
}

func NewToken1Request(contract *types.Address) *Token1Request {
	return &Token1Request{
		Contract: contract,
	}
}

func (r *Token1Request) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetReservesRequest is getReserves()'s input on Contract.
type GetReservesRequest struct {
	Contract *types.Address
}

func NewGetReservesRequest(contract *types.Address) *GetReservesRequest {
	return &GetReservesRequest{
		Contract: contract,
	}
}

func (r *GetReservesRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// Price0CumulativeLastRequest is price0CumulativeLast()'s input on Contract.
type Price0CumulativeLastRequest struct {
	Contract *types.Address
}

func NewPrice0CumulativeLastRequest(contract *types.Address) *Price0CumulativeLastRequest {
	return &Price0CumulativeLastRequest{
		Contract: contract,
	}
}

func (r *Price0CumulativeLastRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// Price1CumulativeLastRequest is price1CumulativeLast()'s input on Contract.
type Price1CumulativeLastRequest struct {
	Contract *types.Address
}

func NewPrice1CumulativeLastRequest(contract *types.Address) *Price1CumulativeLastRequest {
	return &Price1CumulativeLastRequest{
		Contract: contract,
	}
}

func (r *Price1CumulativeLastRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// KLastRequest is kLast()'s input on Contract.
type KLastRequest struct {
	Contract *types.Address
}

func NewKLastRequest(contract *types.Address) *KLastRequest {
	return &KLastRequest{
		Contract: contract,
	}
}

func (r *KLastRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// PermitTypehashRequest is PERMIT_TYPEHASH()'s input on Contract.
type PermitTypehashRequest struct {
	Contract *types.Address
}

func NewPermitTypehashRequest(contract *types.Address) *PermitTypehashRequest {
	return &PermitTypehashRequest{
		Contract: contract,
	}
}

func (r *PermitTypehashRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// InitializeRequest is initialize(token0, token1)'s input on Contract.
type InitializeRequest struct {
	Contract *types.Address
	Token0   *types.Address
	Token1   *types.Address
}

func NewInitializeRequest(contract, token0, token1 *types.Address) *InitializeRequest {
	return &InitializeRequest{
		Contract: contract,
		Token0:   token0,
		Token1:   token1,
	}
}

func (r *InitializeRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Token0 == nil {
		return fmt.Errorf("%w: token0 is nil", bind.ErrInvalidRequest)
	}

	if r.Token1 == nil {
		return fmt.Errorf("%w: token1 is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// MintRequest is mint(to)'s input on Contract.
type MintRequest struct {
	Contract *types.Address
	To       *types.Address
}

func NewMintRequest(contract, to *types.Address) *MintRequest {
	return &MintRequest{
		Contract: contract,
		To:       to,
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

	return nil
}

// BurnRequest is burn(to)'s input on Contract.
type BurnRequest struct {
	Contract *types.Address
	To       *types.Address
}

func NewBurnRequest(contract, to *types.Address) *BurnRequest {
	return &BurnRequest{
		Contract: contract,
		To:       to,
	}
}

func (r *BurnRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SwapRequest is swap(amount0Out, amount1Out, to, data)'s input on Contract.
type SwapRequest struct {
	Contract   *types.Address
	Amount0Out *big.Int
	Amount1Out *big.Int
	To         *types.Address
	Data       []byte
}

func NewSwapRequest(contract *types.Address, amount0Out, amount1Out *big.Int, to *types.Address, data []byte) *SwapRequest {
	return &SwapRequest{
		Contract:   contract,
		Amount0Out: amount0Out,
		Amount1Out: amount1Out,
		To:         to,
		Data:       data,
	}
}

func (r *SwapRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Amount0Out == nil {
		return fmt.Errorf("%w: amount0Out is nil", bind.ErrInvalidRequest)
	}

	if r.Amount1Out == nil {
		return fmt.Errorf("%w: amount1Out is nil", bind.ErrInvalidRequest)
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SkimRequest is skim(to)'s input on Contract.
type SkimRequest struct {
	Contract *types.Address
	To       *types.Address
}

func NewSkimRequest(contract, to *types.Address) *SkimRequest {
	return &SkimRequest{
		Contract: contract,
		To:       to,
	}
}

func (r *SkimRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SyncRequest is sync()'s input on Contract.
type SyncRequest struct {
	Contract *types.Address
}

func NewSyncRequest(contract *types.Address) *SyncRequest {
	return &SyncRequest{
		Contract: contract,
	}
}

func (r *SyncRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// MintEvent mirrors UniswapV2Pair's Mint event, emitted when mint(to) adds liquidity.
type MintEvent struct {
	Contract *types.Address
	Sender   *types.Address
	Amount0  *big.Int
	Amount1  *big.Int
}

func NewMintEvent(contract, sender *types.Address, amount0, amount1 *big.Int) *MintEvent {
	return &MintEvent{
		Contract: contract,
		Sender:   sender,
		Amount0:  amount0,
		Amount1:  amount1,
	}
}

// BurnEvent mirrors UniswapV2Pair's Burn event, emitted when burn(to) removes liquidity.
type BurnEvent struct {
	Contract *types.Address
	Sender   *types.Address
	Amount0  *big.Int
	Amount1  *big.Int
	To       *types.Address
}

func NewBurnEvent(contract, sender *types.Address, amount0, amount1 *big.Int, to *types.Address) *BurnEvent {
	return &BurnEvent{
		Contract: contract,
		Sender:   sender,
		Amount0:  amount0,
		Amount1:  amount1,
		To:       to,
	}
}

// SwapEvent mirrors UniswapV2Pair's Swap event, emitted on every swap.
type SwapEvent struct {
	Contract   *types.Address
	Sender     *types.Address
	Amount0In  *big.Int
	Amount1In  *big.Int
	Amount0Out *big.Int
	Amount1Out *big.Int
	To         *types.Address
}

func NewSwapEvent(contract, sender *types.Address, amount0In, amount1In, amount0Out, amount1Out *big.Int, to *types.Address) *SwapEvent {
	return &SwapEvent{
		Contract:   contract,
		Sender:     sender,
		Amount0In:  amount0In,
		Amount1In:  amount1In,
		Amount0Out: amount0Out,
		Amount1Out: amount1Out,
		To:         to,
	}
}

// SyncEvent mirrors UniswapV2Pair's Sync event, emitted whenever the pair's reserves are updated.
type SyncEvent struct {
	Contract *types.Address
	Reserve0 *big.Int
	Reserve1 *big.Int
}

func NewSyncEvent(contract *types.Address, reserve0, reserve1 *big.Int) *SyncEvent {
	return &SyncEvent{
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
