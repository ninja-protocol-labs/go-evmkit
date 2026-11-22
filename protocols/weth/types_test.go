package weth

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewDepositEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	dst := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	wad := big.NewInt(1000)

	d := NewDepositEvent(contract, dst, wad)
	require.Same(t, contract, d.Contract)
	require.Same(t, dst, d.Dst)
	require.Same(t, wad, d.Wad)
}

func TestNewWithdrawalEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	src := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	wad := big.NewInt(2000)

	w := NewWithdrawalEvent(contract, src, wad)
	require.Same(t, contract, w.Contract)
	require.Same(t, src, w.Src)
	require.Same(t, wad, w.Wad)
}

func TestRequestsValidate(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"deposit nil", (*DepositRequest)(nil).Validate(), "invalid request: request is nil"},
		{"deposit empty", (&DepositRequest{}).Validate(), "invalid request: contract is nil"},
		{"deposit nil contract", (&DepositRequest{Amount: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"deposit nil amount", (&DepositRequest{Contract: contract}).Validate(), "invalid request: amount is nil"},
		{"deposit negative amount", (&DepositRequest{Contract: contract, Amount: big.NewInt(-1)}).Validate(), "invalid request: amount is negative"},
		{"deposit valid", (&DepositRequest{Contract: contract, Amount: big.NewInt(1)}).Validate(), ""},
		{"deposit zero amount", (&DepositRequest{Contract: contract, Amount: big.NewInt(0)}).Validate(), ""},
		{"withdraw nil", (*WithdrawRequest)(nil).Validate(), "invalid request: request is nil"},
		{"withdraw empty", (&WithdrawRequest{}).Validate(), "invalid request: contract is nil"},
		{"withdraw nil contract", (&WithdrawRequest{Amount: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"withdraw nil amount", (&WithdrawRequest{Contract: contract}).Validate(), "invalid request: amount is nil"},
		{"withdraw valid", (&WithdrawRequest{Contract: contract, Amount: big.NewInt(1)}).Validate(), ""},
		{"withdraw zero amount", (&WithdrawRequest{Contract: contract, Amount: big.NewInt(0)}).Validate(), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.want == "" {
				require.NoError(t, tc.err)
				return
			}
			require.ErrorIs(t, tc.err, bind.ErrInvalidRequest)
			require.EqualError(t, tc.err, tc.want)
		})
	}
}

func TestNewRequestsSetFields(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	amount := big.NewInt(100)

	require.Equal(t, &DepositRequest{Contract: contract, Amount: amount}, NewDepositRequest(contract, amount))
	require.Equal(t, &WithdrawRequest{Contract: contract, Amount: amount}, NewWithdrawRequest(contract, amount))
}
