package weth

import (
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// DepositParams is CreateDeposit's argument list.
type DepositParams struct {
	WETH   *types.Address
	Sender *types.Address
	Value  *big.Int
}

func NewDepositParams(weth, sender *types.Address, value *big.Int) *DepositParams {
	return &DepositParams{
		WETH:   weth,
		Sender: sender,
		Value:  value,
	}
}

// WithdrawParams is CreateWithdraw's argument list.
type WithdrawParams struct {
	WETH   *types.Address
	Sender *types.Address
	Wad    *big.Int
}

func NewWithdrawParams(weth, sender *types.Address, wad *big.Int) *WithdrawParams {
	return &WithdrawParams{
		WETH:   weth,
		Sender: sender,
		Wad:    wad,
	}
}

// Deposit mirrors WETH9's Deposit event.
type Deposit struct {
	WETH *types.Address
	Dst  *types.Address
	Wad  *big.Int
}

func NewDeposit(weth, dst *types.Address, wad *big.Int) *Deposit {
	return &Deposit{
		WETH: weth,
		Dst:  dst,
		Wad:  wad,
	}
}

// Withdrawal mirrors WETH9's Withdrawal event.
type Withdrawal struct {
	WETH *types.Address
	Src  *types.Address
	Wad  *big.Int
}

func NewWithdrawal(weth, src *types.Address, wad *big.Int) *Withdrawal {
	return &Withdrawal{
		WETH: weth,
		Src:  src,
		Wad:  wad,
	}
}
