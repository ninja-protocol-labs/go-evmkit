package v2

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// FeeToRequest is feeTo()'s input on Contract.
type FeeToRequest struct {
	Contract *types.Address
}

func NewFeeToRequest(contract *types.Address) *FeeToRequest {
	return &FeeToRequest{
		Contract: contract,
	}
}

func (r *FeeToRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// FeeToSetterRequest is feeToSetter()'s input on Contract.
type FeeToSetterRequest struct {
	Contract *types.Address
}

func NewFeeToSetterRequest(contract *types.Address) *FeeToSetterRequest {
	return &FeeToSetterRequest{
		Contract: contract,
	}
}

func (r *FeeToSetterRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetPairRequest is getPair(tokenA, tokenB)'s input on Contract.
type GetPairRequest struct {
	Contract *types.Address
	TokenA   *types.Address
	TokenB   *types.Address
}

func NewGetPairRequest(contract *types.Address, tokenA *types.Address, tokenB *types.Address) *GetPairRequest {
	return &GetPairRequest{
		Contract: contract,
		TokenA:   tokenA,
		TokenB:   tokenB,
	}
}

func (r *GetPairRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.TokenA == nil {
		return fmt.Errorf("%w: tokenA is nil", bind.ErrInvalidRequest)
	}

	if r.TokenB == nil {
		return fmt.Errorf("%w: tokenB is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// AllPairsRequest is allPairs(index)'s input on Contract.
type AllPairsRequest struct {
	Contract *types.Address
	Index    *big.Int
}

func NewAllPairsRequest(contract *types.Address, index *big.Int) *AllPairsRequest {
	return &AllPairsRequest{
		Contract: contract,
		Index:    index,
	}
}

func (r *AllPairsRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Index == nil {
		return fmt.Errorf("%w: index is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// AllPairsLengthRequest is allPairsLength()'s input on Contract.
type AllPairsLengthRequest struct {
	Contract *types.Address
}

func NewAllPairsLengthRequest(contract *types.Address) *AllPairsLengthRequest {
	return &AllPairsLengthRequest{
		Contract: contract,
	}
}

func (r *AllPairsLengthRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// CreatePairRequest is createPair(tokenA, tokenB)'s input on Contract.
type CreatePairRequest struct {
	Contract *types.Address
	TokenA   *types.Address
	TokenB   *types.Address
}

func NewCreatePairRequest(contract *types.Address, tokenA *types.Address, tokenB *types.Address) *CreatePairRequest {
	return &CreatePairRequest{
		Contract: contract,
		TokenA:   tokenA,
		TokenB:   tokenB,
	}
}

func (r *CreatePairRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.TokenA == nil {
		return fmt.Errorf("%w: tokenA is nil", bind.ErrInvalidRequest)
	}

	if r.TokenB == nil {
		return fmt.Errorf("%w: tokenB is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SetFeeToRequest is setFeeTo(feeTo)'s input on Contract.
type SetFeeToRequest struct {
	Contract *types.Address
	FeeTo    *types.Address
}

func NewSetFeeToRequest(contract *types.Address, feeTo *types.Address) *SetFeeToRequest {
	return &SetFeeToRequest{
		Contract: contract,
		FeeTo:    feeTo,
	}
}

func (r *SetFeeToRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.FeeTo == nil {
		return fmt.Errorf("%w: feeTo is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SetFeeToSetterRequest is setFeeToSetter(feeToSetter)'s input on Contract.
type SetFeeToSetterRequest struct {
	Contract    *types.Address
	FeeToSetter *types.Address
}

func NewSetFeeToSetterRequest(contract *types.Address, feeToSetter *types.Address) *SetFeeToSetterRequest {
	return &SetFeeToSetterRequest{
		Contract:    contract,
		FeeToSetter: feeToSetter,
	}
}

func (r *SetFeeToSetterRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.FeeToSetter == nil {
		return fmt.Errorf("%w: feeToSetter is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// PairCreatedEvent mirrors IUniswapV2Factory's PairCreated event; AllPairsLength is allPairs.length
// right after the push, i.e. the pair's 1-indexed position.
type PairCreatedEvent struct {
	Contract       *types.Address
	Token0         *types.Address
	Token1         *types.Address
	Pair           *types.Address
	AllPairsLength *big.Int
}

func NewPairCreatedEvent(contract, token0, token1, pair *types.Address, allPairsLength *big.Int) *PairCreatedEvent {
	return &PairCreatedEvent{
		Contract:       contract,
		Token0:         token0,
		Token1:         token1,
		Pair:           pair,
		AllPairsLength: allPairsLength,
	}
}
