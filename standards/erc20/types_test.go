package erc20

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

func TestNewTransferEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	from := types.MustNewAddressFromHex("0x67d03631fe51b741c0c00c4e16eb662ac84381df")
	to := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	value := big.NewInt(42)

	tr := NewTransferEvent(contract, from, to, value)
	require.Same(t, contract, tr.Contract)
	require.Same(t, from, tr.From)
	require.Same(t, to, tr.To)
	require.Same(t, value, tr.Value)
}

func TestNewApprovalEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	value := big.NewInt(7)

	ap := NewApprovalEvent(contract, owner, spender, value)
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
	addr := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	amount := big.NewInt(5)

	ab := NewAddressBalance(addr, amount)
	require.Same(t, addr, ab.Address)
	require.Same(t, amount, ab.Amount)
}

func TestNewTokenBalance(t *testing.T) {
	token := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	amount := big.NewInt(6)

	tb := NewTokenBalance(token, amount)
	require.Same(t, token, tb.Token)
	require.Same(t, amount, tb.Amount)
}

func TestNewTokenAddressBalance(t *testing.T) {
	token := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	addr := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	amount := big.NewInt(9)

	tab := NewTokenAddressBalance(token, addr, amount)
	require.Same(t, token, tab.Token)
	require.Same(t, addr, tab.Address)
	require.Same(t, amount, tab.Amount)
}

func TestNewAddressBalances(t *testing.T) {
	addr := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	tokens := []TokenBalance{
		*NewTokenBalance(types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006"), big.NewInt(1)),
		*NewTokenBalance(types.MustNewAddressFromHex("0x833589fcd6edb6e08f4c7c32d4f71b54bda02913"), big.NewInt(2)),
	}

	ab := NewAddressBalances(addr, tokens)
	require.Same(t, addr, ab.Address)
	require.Equal(t, tokens, ab.Tokens)
}

func TestNewTokenBalances(t *testing.T) {
	token := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	addrs := []AddressBalance{
		*NewAddressBalance(types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f"), big.NewInt(1)),
		*NewAddressBalance(types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8"), big.NewInt(2)),
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
		{"name", nameFn, "name()", "0x06fdde03"},
		{"symbol", symbolFn, "symbol()", "0x95d89b41"},
		{"decimals", decimalsFn, "decimals()", "0x313ce567"},
		{"totalSupply", totalSupplyFn, "totalSupply()", "0x18160ddd"},
		{"balanceOf", balanceOfFn, "balanceOf(address)", "0x70a08231"},
		{"transfer", transferFn, "transfer(address,uint256)", "0xa9059cbb"},
		{"allowance", allowanceFn, "allowance(address,address)", "0xdd62ed3e"},
		{"approve", approveFn, "approve(address,uint256)", "0x095ea7b3"},
		{"transferFrom", transferFromFn, "transferFrom(address,address,uint256)", "0x23b872dd"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.sig, tc.fn.Signature())
			require.Equal(t, tc.sel, tc.fn.Selector().String())
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
		{"Transfer", transferEvent, "Transfer(address,address,uint256)", "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"},
		{"Approval", approvalEvent, "Approval(address,address,uint256)", "0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.sig, tc.event.Signature())
			t0, ok := tc.event.Topic0()
			require.True(t, ok)
			require.Equal(t, tc.topic, t0.String())
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

func TestNewTotalSupplyRequest(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")

	req := NewTotalSupplyRequest(contract)
	require.Equal(t, &TotalSupplyRequest{Contract: contract}, req)
}

func TestNewBalanceOfRequest(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	account := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	req := NewBalanceOfRequest(contract, account)
	require.Equal(t, &BalanceOfRequest{Contract: contract, Account: account}, req)
}

func TestNewAllowanceRequest(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	req := NewAllowanceRequest(contract, owner, spender)
	require.Equal(t, &AllowanceRequest{Contract: contract, Owner: owner, Spender: spender}, req)
}

func TestNewMetadataRequests(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")

	require.Equal(t, &NameRequest{Contract: contract}, NewNameRequest(contract))
	require.Equal(t, &SymbolRequest{Contract: contract}, NewSymbolRequest(contract))
	require.Equal(t, &DecimalsRequest{Contract: contract}, NewDecimalsRequest(contract))
}

func TestNewTransferRequests(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	from := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	to := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	value := big.NewInt(12345)

	require.Equal(t, &TransferRequest{Contract: contract, To: to, Amount: value}, NewTransferRequest(contract, to, value))
	require.Equal(t, &ApproveRequest{Contract: contract, Spender: to, Amount: value}, NewApproveRequest(contract, to, value))
	require.Equal(t, &TransferFromRequest{Contract: contract, From: from, To: to, Amount: value}, NewTransferFromRequest(contract, from, to, value))
}

func TestReadsRejectInvalidRequest(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	value := big.NewInt(1)
	e := New(rpc.NewClient("http://127.0.0.1:0"))

	for _, tc := range []struct {
		req  *TotalSupplyRequest
		want string
	}{
		{nil, "request is nil"},
		{NewTotalSupplyRequest(nil), "contract is nil"},
	} {
		n, err := e.TotalSupply(t.Context(), nil, tc.req)
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.ErrorContains(t, err, tc.want)
		require.Nil(t, n)
	}

	for _, tc := range []struct {
		req  *BalanceOfRequest
		want string
	}{
		{nil, "request is nil"},
		{NewBalanceOfRequest(nil, owner), "contract is nil"},
		{NewBalanceOfRequest(contract, nil), "account is nil"},
	} {
		n, err := e.BalanceOf(t.Context(), nil, tc.req)
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.ErrorContains(t, err, tc.want)
		require.Nil(t, n)
	}

	for _, tc := range []struct {
		req  *AllowanceRequest
		want string
	}{
		{nil, "request is nil"},
		{NewAllowanceRequest(nil, owner, owner), "contract is nil"},
		{NewAllowanceRequest(contract, nil, owner), "owner is nil"},
		{NewAllowanceRequest(contract, owner, nil), "spender is nil"},
	} {
		n, err := e.Allowance(t.Context(), nil, tc.req)
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.ErrorContains(t, err, tc.want)
		require.Nil(t, n)
	}

	for _, tc := range []struct {
		req  *NameRequest
		want string
	}{
		{nil, "request is nil"},
		{NewNameRequest(nil), "contract is nil"},
	} {
		out, err := e.Name(t.Context(), nil, tc.req)
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.ErrorContains(t, err, tc.want)
		require.Empty(t, out)
	}

	for _, tc := range []struct {
		req  *SymbolRequest
		want string
	}{
		{nil, "request is nil"},
		{NewSymbolRequest(nil), "contract is nil"},
	} {
		out, err := e.Symbol(t.Context(), nil, tc.req)
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.ErrorContains(t, err, tc.want)
		require.Empty(t, out)
	}

	for _, tc := range []struct {
		req  *DecimalsRequest
		want string
	}{
		{nil, "request is nil"},
		{NewDecimalsRequest(nil), "contract is nil"},
	} {
		out, err := e.Decimals(t.Context(), nil, tc.req)
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.ErrorContains(t, err, tc.want)
		require.Zero(t, out)
	}

	for _, tc := range []struct {
		req  *TransferRequest
		want string
	}{
		{nil, "request is nil"},
		{NewTransferRequest(nil, owner, value), "contract is nil"},
		{NewTransferRequest(contract, nil, value), "to is nil"},
		{NewTransferRequest(contract, owner, nil), "amount is nil"},
	} {
		ok, err := e.CallTransfer(t.Context(), nil, tc.req)
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.ErrorContains(t, err, tc.want)
		require.False(t, ok)
	}

	for _, tc := range []struct {
		req  *ApproveRequest
		want string
	}{
		{nil, "request is nil"},
		{NewApproveRequest(nil, owner, value), "contract is nil"},
		{NewApproveRequest(contract, nil, value), "spender is nil"},
		{NewApproveRequest(contract, owner, nil), "amount is nil"},
	} {
		ok, err := e.CallApprove(t.Context(), nil, tc.req)
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.ErrorContains(t, err, tc.want)
		require.False(t, ok)
	}

	for _, tc := range []struct {
		req  *TransferFromRequest
		want string
	}{
		{nil, "request is nil"},
		{NewTransferFromRequest(nil, owner, owner, value), "contract is nil"},
		{NewTransferFromRequest(contract, nil, owner, value), "from is nil"},
		{NewTransferFromRequest(contract, owner, nil, value), "to is nil"},
		{NewTransferFromRequest(contract, owner, owner, nil), "amount is nil"},
	} {
		ok, err := e.CallTransferFrom(t.Context(), nil, tc.req)
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.ErrorContains(t, err, tc.want)
		require.False(t, ok)
	}
}

func TestWritesRejectInvalidRequest(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	value := big.NewInt(1)
	e := New(rpc.NewClient("http://127.0.0.1:0"))

	for _, tc := range []struct {
		req  *TransferRequest
		want string
	}{
		{nil, "request is nil"},
		{NewTransferRequest(nil, owner, value), "contract is nil"},
		{NewTransferRequest(contract, nil, value), "to is nil"},
		{NewTransferRequest(contract, owner, nil), "amount is nil"},
	} {
		signed, hash, err := e.Transfer(t.Context(), nil, tc.req)
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.ErrorContains(t, err, tc.want)
		require.Nil(t, signed)
		require.Nil(t, hash)
	}

	for _, tc := range []struct {
		req  *ApproveRequest
		want string
	}{
		{nil, "request is nil"},
		{NewApproveRequest(nil, owner, value), "contract is nil"},
		{NewApproveRequest(contract, nil, value), "spender is nil"},
		{NewApproveRequest(contract, owner, nil), "amount is nil"},
	} {
		signed, hash, err := e.Approve(t.Context(), nil, tc.req)
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.ErrorContains(t, err, tc.want)
		require.Nil(t, signed)
		require.Nil(t, hash)
	}

	for _, tc := range []struct {
		req  *TransferFromRequest
		want string
	}{
		{nil, "request is nil"},
		{NewTransferFromRequest(nil, owner, owner, value), "contract is nil"},
		{NewTransferFromRequest(contract, nil, owner, value), "from is nil"},
		{NewTransferFromRequest(contract, owner, nil, value), "to is nil"},
		{NewTransferFromRequest(contract, owner, owner, nil), "amount is nil"},
	} {
		signed, hash, err := e.TransferFrom(t.Context(), nil, tc.req)
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.ErrorContains(t, err, tc.want)
		require.Nil(t, signed)
		require.Nil(t, hash)
	}

	signed, hash, err := e.Transfer(t.Context(), nil, NewTransferRequest(contract, owner, value))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.ErrorContains(t, err, "options is nil")
	require.Nil(t, signed)
	require.Nil(t, hash)
}

func TestNewBatcherRequests(t *testing.T) {
	multicall := types.MustNewAddressFromHex("0xca11bde05977b3631167028862be2a173976ca11")
	token := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	list := []*types.Address{owner, spender}

	require.Equal(t, &MetadataRequest{Multicall: multicall, Token: token, AllowFailure: true}, NewMetadataRequest(multicall, token, true))
	require.Equal(t, &MetadataWithBalanceRequest{Multicall: multicall, Token: token, Account: owner, AllowFailure: true}, NewMetadataWithBalanceRequest(multicall, token, owner, true))
	require.Equal(t, &AllowanceWithBalanceRequest{Multicall: multicall, Token: token, Owner: owner, Spender: spender}, NewAllowanceWithBalanceRequest(multicall, token, owner, spender, false))
	require.Equal(t, &TokenBalancesRequest{Multicall: multicall, Token: token, Accounts: list}, NewTokenBalancesRequest(multicall, token, list, false))
	require.Equal(t, &AddressBalancesRequest{Multicall: multicall, Account: owner, Tokens: list}, NewAddressBalancesRequest(multicall, owner, list, false))
	require.Equal(t, &BalancePairsRequest{Multicall: multicall, Tokens: list, Accounts: list, AllowFailure: true}, NewBalancePairsRequest(multicall, list, list, true))
}

func TestBatcherRejectsInvalidRequest(t *testing.T) {
	multicall := types.MustNewAddressFromHex("0xca11bde05977b3631167028862be2a173976ca11")
	token := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	e := New(rpc.NewClient("http://127.0.0.1:0"))

	for _, tc := range []struct {
		req  *MetadataRequest
		want string
	}{
		{nil, "request is nil"},
		{NewMetadataRequest(nil, token, false), "multicall is nil"},
		{NewMetadataRequest(multicall, nil, false), "token is nil"},
	} {
		out, err := e.Metadata(t.Context(), nil, tc.req)
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.ErrorContains(t, err, tc.want)
		require.Nil(t, out)
	}

	for _, tc := range []struct {
		req  *MetadataWithBalanceRequest
		want string
	}{
		{nil, "request is nil"},
		{NewMetadataWithBalanceRequest(nil, token, owner, false), "multicall is nil"},
		{NewMetadataWithBalanceRequest(multicall, nil, owner, false), "token is nil"},
		{NewMetadataWithBalanceRequest(multicall, token, nil, false), "account is nil"},
	} {
		out, err := e.MetadataWithBalance(t.Context(), nil, tc.req)
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.ErrorContains(t, err, tc.want)
		require.Nil(t, out)
	}

	for _, tc := range []struct {
		req  *AllowanceWithBalanceRequest
		want string
	}{
		{nil, "request is nil"},
		{NewAllowanceWithBalanceRequest(nil, token, owner, owner, false), "multicall is nil"},
		{NewAllowanceWithBalanceRequest(multicall, nil, owner, owner, false), "token is nil"},
		{NewAllowanceWithBalanceRequest(multicall, token, nil, owner, false), "owner is nil"},
		{NewAllowanceWithBalanceRequest(multicall, token, owner, nil, false), "spender is nil"},
	} {
		out, err := e.AllowanceWithBalance(t.Context(), nil, tc.req)
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.ErrorContains(t, err, tc.want)
		require.Nil(t, out)
	}

	for _, tc := range []struct {
		req  *TokenBalancesRequest
		want string
	}{
		{nil, "request is nil"},
		{NewTokenBalancesRequest(nil, token, []*types.Address{owner}, false), "multicall is nil"},
		{NewTokenBalancesRequest(multicall, nil, []*types.Address{owner}, false), "token is nil"},
		{NewTokenBalancesRequest(multicall, token, []*types.Address{owner, nil}, false), "accounts[1] is nil"},
	} {
		out, err := e.TokenBalances(t.Context(), nil, tc.req)
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.ErrorContains(t, err, tc.want)
		require.Nil(t, out)
	}

	for _, tc := range []struct {
		req  *AddressBalancesRequest
		want string
	}{
		{nil, "request is nil"},
		{NewAddressBalancesRequest(nil, owner, []*types.Address{token}, false), "multicall is nil"},
		{NewAddressBalancesRequest(multicall, nil, []*types.Address{token}, false), "account is nil"},
		{NewAddressBalancesRequest(multicall, owner, []*types.Address{nil, token}, false), "tokens[0] is nil"},
	} {
		out, err := e.AddressBalances(t.Context(), nil, tc.req)
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.ErrorContains(t, err, tc.want)
		require.Nil(t, out)
	}

	for _, tc := range []struct {
		req  *BalancePairsRequest
		want string
	}{
		{nil, "request is nil"},
		{NewBalancePairsRequest(nil, []*types.Address{token}, []*types.Address{owner}, false), "multicall is nil"},
		{NewBalancePairsRequest(multicall, []*types.Address{token}, []*types.Address{owner, owner}, false), "tokens and accounts differ in length"},
		{NewBalancePairsRequest(multicall, []*types.Address{token, nil}, []*types.Address{owner, owner}, false), "tokens[1] is nil"},
		{NewBalancePairsRequest(multicall, []*types.Address{token, token}, []*types.Address{nil, owner}, false), "accounts[0] is nil"},
	} {
		out, err := e.BalancePairs(t.Context(), nil, tc.req)
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.ErrorContains(t, err, tc.want)
		require.Nil(t, out)
	}
}
