package erc20

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func mustAddr(t *testing.T, s string) *types.Address {
	t.Helper()
	a, err := types.NewAddressFromHex(s)
	require.NoError(t, err)
	return a
}

func TestNewTransfer(t *testing.T) {
	contract := mustAddr(t, "0x4200000000000000000000000000000000000006")
	from := mustAddr(t, "0x67d03631fe51b741c0c00c4e16eb662ac84381df")
	to := mustAddr(t, "0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	value := big.NewInt(42)

	tr := NewTransfer(contract, from, to, value)
	require.Same(t, contract, tr.Contract)
	require.Same(t, from, tr.From)
	require.Same(t, to, tr.To)
	require.Same(t, value, tr.Value)
}

func TestNewApproval(t *testing.T) {
	contract := mustAddr(t, "0x4200000000000000000000000000000000000006")
	owner := mustAddr(t, "0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	spender := mustAddr(t, "0xba12222222228d8ba445958a75a0704d566bf2c8")
	value := big.NewInt(7)

	ap := NewApproval(contract, owner, spender, value)
	require.Same(t, contract, ap.Contract)
	require.Same(t, owner, ap.Owner)
	require.Same(t, spender, ap.Spender)
	require.Same(t, value, ap.Value)
}

func TestNewMetadata(t *testing.T) {
	m := NewMetadata("USD Coin", "USDC", 6, big.NewInt(1000))
	require.Equal(t, "USD Coin", m.Name)
	require.Equal(t, "USDC", m.Symbol)
	require.Equal(t, uint8(6), m.Decimals)
	require.Zero(t, big.NewInt(1000).Cmp(m.TotalSupply))
}

func TestNewMetadataWithBalance(t *testing.T) {
	m := NewMetadata("USD Coin", "USDC", 6, big.NewInt(1000))
	balance := big.NewInt(99)

	mb := NewMetadataWithBalance(m, balance)
	require.Equal(t, *m, mb.Metadata)
	require.Same(t, balance, mb.Balance)
}

func TestNewAllowanceWithBalance(t *testing.T) {
	value := big.NewInt(1)
	ownerBal := big.NewInt(2)
	spenderBal := big.NewInt(3)

	awb := NewAllowanceWithBalance(value, ownerBal, spenderBal)
	require.Same(t, value, awb.Value)
	require.Same(t, ownerBal, awb.OwnerBalance)
	require.Same(t, spenderBal, awb.SpenderBalance)
}

func TestNewAddressBalance(t *testing.T) {
	addr := mustAddr(t, "0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	amount := big.NewInt(5)

	ab := NewAddressBalance(addr, amount)
	require.Same(t, addr, ab.Address)
	require.Same(t, amount, ab.Amount)
}

func TestNewTokenBalance(t *testing.T) {
	token := mustAddr(t, "0x4200000000000000000000000000000000000006")
	amount := big.NewInt(6)

	tb := NewTokenBalance(token, amount)
	require.Same(t, token, tb.Token)
	require.Same(t, amount, tb.Amount)
}

func TestNewTokenAddressBalance(t *testing.T) {
	token := mustAddr(t, "0x4200000000000000000000000000000000000006")
	addr := mustAddr(t, "0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	amount := big.NewInt(9)

	tab := NewTokenAddressBalance(token, addr, amount)
	require.Same(t, token, tab.Token)
	require.Same(t, addr, tab.Address)
	require.Same(t, amount, tab.Amount)
}

func TestNewAddressBalances(t *testing.T) {
	addr := mustAddr(t, "0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	tokens := []TokenBalance{
		*NewTokenBalance(mustAddr(t, "0x4200000000000000000000000000000000000006"), big.NewInt(1)),
		*NewTokenBalance(mustAddr(t, "0x833589fcd6edb6e08f4c7c32d4f71b54bda02913"), big.NewInt(2)),
	}

	ab := NewAddressBalances(addr, tokens)
	require.Same(t, addr, ab.Address)
	require.Equal(t, tokens, ab.Tokens)
}

func TestNewTokenBalances(t *testing.T) {
	token := mustAddr(t, "0x4200000000000000000000000000000000000006")
	addrs := []AddressBalance{
		*NewAddressBalance(mustAddr(t, "0x6667c8dc9fbfec411e7c1ee2b24de960149f930f"), big.NewInt(1)),
		*NewAddressBalance(mustAddr(t, "0xba12222222228d8ba445958a75a0704d566bf2c8"), big.NewInt(2)),
	}

	tbs := NewTokenBalances(token, addrs)
	require.Same(t, token, tbs.Token)
	require.Equal(t, addrs, tbs.Addresses)
}

func TestFunctionSignatures(t *testing.T) {
	tests := []struct {
		name string
		fn   *abi.Function
		sig  string
		sel  string
	}{
		{"name", nameFn, "name()", "06fdde03"},
		{"symbol", symbolFn, "symbol()", "95d89b41"},
		{"decimals", decimalsFn, "decimals()", "313ce567"},
		{"totalSupply", totalSupplyFn, "totalSupply()", "18160ddd"},
		{"balanceOf", balanceOfFn, "balanceOf(address)", "70a08231"},
		{"transfer", transferFn, "transfer(address,uint256)", "a9059cbb"},
		{"allowance", allowanceFn, "allowance(address,address)", "dd62ed3e"},
		{"approve", approveFn, "approve(address,uint256)", "095ea7b3"},
		{"transferFrom", transferFromFn, "transferFrom(address,address,uint256)", "23b872dd"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.sig, tc.fn.Signature())
			sel := tc.fn.Selector()
			require.Equal(t, tc.sel, hexString(sel[:]))
		})
	}
}

func TestEventSignatures(t *testing.T) {
	tests := []struct {
		name  string
		event *abi.Event
		sig   string
		topic string
	}{
		{"Transfer", transferEvent, "Transfer(address,address,uint256)", "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"},
		{"Approval", approvalEvent, "Approval(address,address,uint256)", "8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.sig, tc.event.Signature())
			t0, ok := tc.event.Topic0()
			require.True(t, ok)
			require.Equal(t, tc.topic, hexString(t0[:]))
			require.False(t, tc.event.Anonymous)
			require.Len(t, tc.event.Inputs, 3)
			require.True(t, tc.event.Inputs[0].Indexed)
			require.True(t, tc.event.Inputs[1].Indexed)
			require.False(t, tc.event.Inputs[2].Indexed)
		})
	}
}

func TestCustomErrorSignatures(t *testing.T) {
	tests := []struct {
		name string
		err  *abi.Error
		sig  string
	}{
		{"ERC20InsufficientBalance", errInsufficientBalance, "ERC20InsufficientBalance(address,uint256,uint256)"},
		{"ERC20InvalidSender", errInvalidSender, "ERC20InvalidSender(address)"},
		{"ERC20InvalidReceiver", errInvalidReceiver, "ERC20InvalidReceiver(address)"},
		{"ERC20InsufficientAllowance", errInsufficientAllowance, "ERC20InsufficientAllowance(address,uint256,uint256)"},
		{"ERC20InvalidApprover", errInvalidApprover, "ERC20InvalidApprover(address)"},
		{"ERC20InvalidSpender", errInvalidSpender, "ERC20InvalidSpender(address)"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.sig, tc.err.Signature())

			parsed, err := abi.ParseError(tc.sig)
			require.NoError(t, err)
			require.Equal(t, tc.err.Selector(), parsed.Selector())
		})
	}
}

func hexString(b []byte) string {
	const hextable = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, c := range b {
		out[i*2] = hextable[c>>4]
		out[i*2+1] = hextable[c&0x0f]
	}
	return string(out)
}
