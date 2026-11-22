package weth

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// DepositRequest is deposit()'s input on Contract; Amount is sent as msg.value.
type DepositRequest struct {
	Contract *types.Address
	Amount   *big.Int
}

func NewDepositRequest(contract *types.Address, amount *big.Int) *DepositRequest {
	return &DepositRequest{
		Contract: contract,
		Amount:   amount,
	}
}

func (r *DepositRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Amount == nil {
		return fmt.Errorf("%w: amount is nil", bind.ErrInvalidRequest)
	}

	if r.Amount.Sign() < 0 {
		return fmt.Errorf("%w: amount is negative", bind.ErrInvalidRequest)
	}

	return nil
}

// WithdrawRequest is withdraw(wad)'s input on Contract.
type WithdrawRequest struct {
	Contract *types.Address
	Amount   *big.Int
}

func NewWithdrawRequest(contract *types.Address, amount *big.Int) *WithdrawRequest {
	return &WithdrawRequest{
		Contract: contract,
		Amount:   amount,
	}
}

func (r *WithdrawRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Amount == nil {
		return fmt.Errorf("%w: amount is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// DepositEvent mirrors WETH9's Deposit event.
type DepositEvent struct {
	Contract *types.Address
	Dst      *types.Address
	Wad      *big.Int
}

func NewDepositEvent(contract, dst *types.Address, wad *big.Int) *DepositEvent {
	return &DepositEvent{
		Contract: contract,
		Dst:      dst,
		Wad:      wad,
	}
}

// WithdrawalEvent mirrors WETH9's Withdrawal event.
type WithdrawalEvent struct {
	Contract *types.Address
	Src      *types.Address
	Wad      *big.Int
}

func NewWithdrawalEvent(contract, src *types.Address, wad *big.Int) *WithdrawalEvent {
	return &WithdrawalEvent{
		Contract: contract,
		Src:      src,
		Wad:      wad,
	}
}
