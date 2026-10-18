package weth

import (
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Deposit mirrors WETH9's Deposit event.
type Deposit struct {
	Contract *types.Address
	Dst      *types.Address
	Wad      *big.Int
}

func NewDeposit(contract, dst *types.Address, wad *big.Int) *Deposit {
	return &Deposit{
		Contract: contract,
		Dst:      dst,
		Wad:      wad,
	}
}

// Withdrawal mirrors WETH9's Withdrawal event.
type Withdrawal struct {
	Contract *types.Address
	Src      *types.Address
	Wad      *big.Int
}

func NewWithdrawal(contract, src *types.Address, wad *big.Int) *Withdrawal {
	return &Withdrawal{
		Contract: contract,
		Src:      src,
		Wad:      wad,
	}
}
