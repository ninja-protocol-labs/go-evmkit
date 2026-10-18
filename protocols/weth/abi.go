package weth

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
)

// Function signatures from WETH9's own surface -- the rest come from erc20.
var (
	depositFn  = abi.NewFunction("deposit", nil, nil)
	withdrawFn = abi.NewFunction("withdraw", abi.NewTypes(abi.Uint256), nil)
)

// Events from WETH9's own surface -- Transfer/Approval come from erc20 instead.
var (
	depositEvent = abi.NewEvent("Deposit", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Uint256},
	}, false)
	withdrawalEvent = abi.NewEvent("Withdrawal", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Uint256},
	}, false)
)

// EncodeDeposit returns the calldata for deposit().
func EncodeDeposit() []byte {
	return depositFn.SelectorBytes()
}

// EncodeWithdraw returns the calldata for withdraw(wad).
func EncodeWithdraw(wad *big.Int) ([]byte, error) {
	data, err := withdrawFn.EncodeCall(wad)
	if err != nil {
		return nil, fmt.Errorf("weth: encodeWithdraw: %w", err)
	}
	return data, nil
}
