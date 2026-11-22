package erc20

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
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
	e := New(mocks.NewClient(t))

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
	e := New(mocks.NewClient(t))

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
	e := New(mocks.NewClient(t))

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
		{NewTokenBalancesRequest(multicall, token, nil, false), "accounts is empty"},
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
		{NewAddressBalancesRequest(multicall, owner, []*types.Address{}, false), "tokens is empty"},
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
		{NewBalancePairsRequest(multicall, nil, nil, false), "tokens is empty"},
		{NewBalancePairsRequest(multicall, []*types.Address{token, nil}, []*types.Address{owner, owner}, false), "tokens[1] is nil"},
		{NewBalancePairsRequest(multicall, []*types.Address{token, token}, []*types.Address{nil, owner}, false), "accounts[0] is nil"},
	} {
		out, err := e.BalancePairs(t.Context(), nil, tc.req)
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.ErrorContains(t, err, tc.want)
		require.Nil(t, out)
	}
}

func TestSingleContractRequestsValidate(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")

	var nilTotalSupply *TotalSupplyRequest
	var nilName *NameRequest
	var nilSymbol *SymbolRequest
	var nilDecimals *DecimalsRequest

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"totalSupply nil", nilTotalSupply.Validate(), "invalid request: request is nil"},
		{"totalSupply empty", (&TotalSupplyRequest{}).Validate(), "invalid request: contract is nil"},
		{"totalSupply valid", NewTotalSupplyRequest(contract).Validate(), ""},
		{"name nil", nilName.Validate(), "invalid request: request is nil"},
		{"name empty", (&NameRequest{}).Validate(), "invalid request: contract is nil"},
		{"name valid", NewNameRequest(contract).Validate(), ""},
		{"symbol nil", nilSymbol.Validate(), "invalid request: request is nil"},
		{"symbol empty", (&SymbolRequest{}).Validate(), "invalid request: contract is nil"},
		{"symbol valid", NewSymbolRequest(contract).Validate(), ""},
		{"decimals nil", nilDecimals.Validate(), "invalid request: request is nil"},
		{"decimals empty", (&DecimalsRequest{}).Validate(), "invalid request: contract is nil"},
		{"decimals valid", NewDecimalsRequest(contract).Validate(), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.want == "" {
				require.NoError(t, tc.err)
				return
			}
			require.ErrorIs(t, tc.err, bind.ErrInvalidRequest)
			require.NotErrorIs(t, tc.err, bind.ErrInvalidOptions)
			require.EqualError(t, tc.err, tc.want)
		})
	}
}

func TestBalanceOfRequestValidate(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	account := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	for _, tc := range []struct {
		name string
		req  *BalanceOfRequest
		want string
	}{
		{"nil request", nil, "invalid request: request is nil"},
		{"empty", &BalanceOfRequest{}, "invalid request: contract is nil"},
		{"nil contract", NewBalanceOfRequest(nil, account), "invalid request: contract is nil"},
		{"nil account", NewBalanceOfRequest(contract, nil), "invalid request: account is nil"},
		{"zero addresses", NewBalanceOfRequest(types.ZeroAddress(), types.ZeroAddress()), ""},
		{"valid", NewBalanceOfRequest(contract, account), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			if tc.want == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, bind.ErrInvalidRequest)
			require.EqualError(t, err, tc.want)
		})
	}
}

func TestAllowanceRequestValidate(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	for _, tc := range []struct {
		name string
		req  *AllowanceRequest
		want string
	}{
		{"nil request", nil, "invalid request: request is nil"},
		{"empty", &AllowanceRequest{}, "invalid request: contract is nil"},
		{"nil contract", NewAllowanceRequest(nil, owner, spender), "invalid request: contract is nil"},
		{"nil owner", NewAllowanceRequest(contract, nil, spender), "invalid request: owner is nil"},
		{"nil spender", NewAllowanceRequest(contract, owner, nil), "invalid request: spender is nil"},
		{"nil owner and spender", NewAllowanceRequest(contract, nil, nil), "invalid request: owner is nil"},
		{"owner equals spender", NewAllowanceRequest(contract, owner, owner), ""},
		{"valid", NewAllowanceRequest(contract, owner, spender), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			if tc.want == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, bind.ErrInvalidRequest)
			require.EqualError(t, err, tc.want)
		})
	}
}

func TestTransferRequestValidate(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	to := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	maxUint256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))

	for _, tc := range []struct {
		name string
		req  *TransferRequest
		want string
	}{
		{"nil request", nil, "invalid request: request is nil"},
		{"empty", &TransferRequest{}, "invalid request: contract is nil"},
		{"nil contract", NewTransferRequest(nil, to, big.NewInt(1)), "invalid request: contract is nil"},
		{"nil to", NewTransferRequest(contract, nil, big.NewInt(1)), "invalid request: to is nil"},
		{"nil amount", NewTransferRequest(contract, to, nil), "invalid request: amount is nil"},
		{"nil to and amount", NewTransferRequest(contract, nil, nil), "invalid request: to is nil"},
		{"zero amount", NewTransferRequest(contract, to, big.NewInt(0)), ""},
		{"max uint256 amount", NewTransferRequest(contract, to, maxUint256), ""},
		{"zero address to", NewTransferRequest(contract, types.ZeroAddress(), big.NewInt(1)), ""},
		{"valid", NewTransferRequest(contract, to, big.NewInt(12345)), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			if tc.want == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, bind.ErrInvalidRequest)
			require.EqualError(t, err, tc.want)
		})
	}
}

func TestApproveRequestValidate(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	maxUint256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))

	for _, tc := range []struct {
		name string
		req  *ApproveRequest
		want string
	}{
		{"nil request", nil, "invalid request: request is nil"},
		{"empty", &ApproveRequest{}, "invalid request: contract is nil"},
		{"nil contract", NewApproveRequest(nil, spender, big.NewInt(1)), "invalid request: contract is nil"},
		{"nil spender", NewApproveRequest(contract, nil, big.NewInt(1)), "invalid request: spender is nil"},
		{"nil amount", NewApproveRequest(contract, spender, nil), "invalid request: amount is nil"},
		{"zero amount reset", NewApproveRequest(contract, spender, big.NewInt(0)), ""},
		{"max uint256 amount", NewApproveRequest(contract, spender, maxUint256), ""},
		{"valid", NewApproveRequest(contract, spender, big.NewInt(12345)), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			if tc.want == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, bind.ErrInvalidRequest)
			require.EqualError(t, err, tc.want)
		})
	}
}

func TestTransferFromRequestValidate(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	from := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	to := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	for _, tc := range []struct {
		name string
		req  *TransferFromRequest
		want string
	}{
		{"nil request", nil, "invalid request: request is nil"},
		{"empty", &TransferFromRequest{}, "invalid request: contract is nil"},
		{"nil contract", NewTransferFromRequest(nil, from, to, big.NewInt(1)), "invalid request: contract is nil"},
		{"nil from", NewTransferFromRequest(contract, nil, to, big.NewInt(1)), "invalid request: from is nil"},
		{"nil to", NewTransferFromRequest(contract, from, nil, big.NewInt(1)), "invalid request: to is nil"},
		{"nil amount", NewTransferFromRequest(contract, from, to, nil), "invalid request: amount is nil"},
		{"only contract", NewTransferFromRequest(contract, nil, nil, nil), "invalid request: from is nil"},
		{"from equals to", NewTransferFromRequest(contract, from, from, big.NewInt(1)), ""},
		{"zero amount", NewTransferFromRequest(contract, from, to, big.NewInt(0)), ""},
		{"valid", NewTransferFromRequest(contract, from, to, big.NewInt(12345)), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			if tc.want == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, bind.ErrInvalidRequest)
			require.EqualError(t, err, tc.want)
		})
	}
}

func TestMetadataRequestsValidate(t *testing.T) {
	multicall := types.MustNewAddressFromHex("0xca11bde05977b3631167028862be2a173976ca11")
	token := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	var nilMetadata *MetadataRequest
	var nilMetadataWithBalance *MetadataWithBalanceRequest
	var nilAllowanceWithBalance *AllowanceWithBalanceRequest

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"metadata nil", nilMetadata.Validate(), "invalid request: request is nil"},
		{"metadata empty", (&MetadataRequest{}).Validate(), "invalid request: multicall is nil"},
		{"metadata nil multicall", NewMetadataRequest(nil, token, false).Validate(), "invalid request: multicall is nil"},
		{"metadata nil token", NewMetadataRequest(multicall, nil, false).Validate(), "invalid request: token is nil"},
		{"metadata allowFailure false", NewMetadataRequest(multicall, token, false).Validate(), ""},
		{"metadata allowFailure true", NewMetadataRequest(multicall, token, true).Validate(), ""},
		{"metadataWithBalance nil", nilMetadataWithBalance.Validate(), "invalid request: request is nil"},
		{"metadataWithBalance empty", (&MetadataWithBalanceRequest{}).Validate(), "invalid request: multicall is nil"},
		{"metadataWithBalance nil multicall", NewMetadataWithBalanceRequest(nil, token, owner, false).Validate(), "invalid request: multicall is nil"},
		{"metadataWithBalance nil token", NewMetadataWithBalanceRequest(multicall, nil, owner, false).Validate(), "invalid request: token is nil"},
		{"metadataWithBalance nil account", NewMetadataWithBalanceRequest(multicall, token, nil, false).Validate(), "invalid request: account is nil"},
		{"metadataWithBalance valid", NewMetadataWithBalanceRequest(multicall, token, owner, true).Validate(), ""},
		{"allowanceWithBalance nil", nilAllowanceWithBalance.Validate(), "invalid request: request is nil"},
		{"allowanceWithBalance empty", (&AllowanceWithBalanceRequest{}).Validate(), "invalid request: multicall is nil"},
		{"allowanceWithBalance nil multicall", NewAllowanceWithBalanceRequest(nil, token, owner, spender, false).Validate(), "invalid request: multicall is nil"},
		{"allowanceWithBalance nil token", NewAllowanceWithBalanceRequest(multicall, nil, owner, spender, false).Validate(), "invalid request: token is nil"},
		{"allowanceWithBalance nil owner", NewAllowanceWithBalanceRequest(multicall, token, nil, spender, false).Validate(), "invalid request: owner is nil"},
		{"allowanceWithBalance nil spender", NewAllowanceWithBalanceRequest(multicall, token, owner, nil, false).Validate(), "invalid request: spender is nil"},
		{"allowanceWithBalance valid", NewAllowanceWithBalanceRequest(multicall, token, owner, spender, true).Validate(), ""},
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

func TestTokenBalancesRequestValidate(t *testing.T) {
	multicall := types.MustNewAddressFromHex("0xca11bde05977b3631167028862be2a173976ca11")
	token := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	for _, tc := range []struct {
		name string
		req  *TokenBalancesRequest
		want string
	}{
		{"nil request", nil, "invalid request: request is nil"},
		{"empty", &TokenBalancesRequest{}, "invalid request: multicall is nil"},
		{"nil multicall", NewTokenBalancesRequest(nil, token, []*types.Address{a}, false), "invalid request: multicall is nil"},
		{"nil token", NewTokenBalancesRequest(multicall, nil, []*types.Address{a}, false), "invalid request: token is nil"},
		{"nil accounts", NewTokenBalancesRequest(multicall, token, nil, false), "invalid request: accounts is empty"},
		{"empty accounts", NewTokenBalancesRequest(multicall, token, []*types.Address{}, false), "invalid request: accounts is empty"},
		{"nil first", NewTokenBalancesRequest(multicall, token, []*types.Address{nil, a, b}, false), "invalid request: accounts[0] is nil"},
		{"nil middle", NewTokenBalancesRequest(multicall, token, []*types.Address{a, nil, b}, false), "invalid request: accounts[1] is nil"},
		{"nil last", NewTokenBalancesRequest(multicall, token, []*types.Address{a, b, nil}, false), "invalid request: accounts[2] is nil"},
		{"all nil", NewTokenBalancesRequest(multicall, token, []*types.Address{nil, nil}, false), "invalid request: accounts[0] is nil"},
		{"single nil", NewTokenBalancesRequest(multicall, token, []*types.Address{nil}, false), "invalid request: accounts[0] is nil"},
		{"single", NewTokenBalancesRequest(multicall, token, []*types.Address{a}, false), ""},
		{"duplicates", NewTokenBalancesRequest(multicall, token, []*types.Address{a, a, a}, true), ""},
		{"valid", NewTokenBalancesRequest(multicall, token, []*types.Address{a, b}, true), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			if tc.want == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, bind.ErrInvalidRequest)
			require.EqualError(t, err, tc.want)
		})
	}
}

func TestAddressBalancesRequestValidate(t *testing.T) {
	multicall := types.MustNewAddressFromHex("0xca11bde05977b3631167028862be2a173976ca11")
	account := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	a := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	b := types.MustNewAddressFromHex("0x833589fcd6edb6e08f4c7c32d4f71b54bda02913")

	for _, tc := range []struct {
		name string
		req  *AddressBalancesRequest
		want string
	}{
		{"nil request", nil, "invalid request: request is nil"},
		{"empty", &AddressBalancesRequest{}, "invalid request: multicall is nil"},
		{"nil multicall", NewAddressBalancesRequest(nil, account, []*types.Address{a}, false), "invalid request: multicall is nil"},
		{"nil account", NewAddressBalancesRequest(multicall, nil, []*types.Address{a}, false), "invalid request: account is nil"},
		{"nil tokens", NewAddressBalancesRequest(multicall, account, nil, false), "invalid request: tokens is empty"},
		{"empty tokens", NewAddressBalancesRequest(multicall, account, []*types.Address{}, false), "invalid request: tokens is empty"},
		{"nil first", NewAddressBalancesRequest(multicall, account, []*types.Address{nil, a, b}, false), "invalid request: tokens[0] is nil"},
		{"nil middle", NewAddressBalancesRequest(multicall, account, []*types.Address{a, nil, b}, false), "invalid request: tokens[1] is nil"},
		{"nil last", NewAddressBalancesRequest(multicall, account, []*types.Address{a, b, nil}, false), "invalid request: tokens[2] is nil"},
		{"single", NewAddressBalancesRequest(multicall, account, []*types.Address{a}, false), ""},
		{"duplicates", NewAddressBalancesRequest(multicall, account, []*types.Address{a, a}, true), ""},
		{"valid", NewAddressBalancesRequest(multicall, account, []*types.Address{a, b}, true), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			if tc.want == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, bind.ErrInvalidRequest)
			require.EqualError(t, err, tc.want)
		})
	}
}

func TestBalancePairsRequestValidate(t *testing.T) {
	multicall := types.MustNewAddressFromHex("0xca11bde05977b3631167028862be2a173976ca11")
	t1 := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	t2 := types.MustNewAddressFromHex("0x833589fcd6edb6e08f4c7c32d4f71b54bda02913")
	a1 := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	a2 := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	for _, tc := range []struct {
		name string
		req  *BalancePairsRequest
		want string
	}{
		{"nil request", nil, "invalid request: request is nil"},
		{"empty", &BalancePairsRequest{}, "invalid request: multicall is nil"},
		{"nil multicall", NewBalancePairsRequest(nil, []*types.Address{t1}, []*types.Address{a1}, false), "invalid request: multicall is nil"},
		{"both nil", NewBalancePairsRequest(multicall, nil, nil, false), "invalid request: tokens is empty"},
		{"both empty", NewBalancePairsRequest(multicall, []*types.Address{}, []*types.Address{}, false), "invalid request: tokens is empty"},
		{"tokens longer", NewBalancePairsRequest(multicall, []*types.Address{t1, t2}, []*types.Address{a1}, false), "invalid request: tokens and accounts differ in length"},
		{"accounts longer", NewBalancePairsRequest(multicall, []*types.Address{t1}, []*types.Address{a1, a2}, false), "invalid request: tokens and accounts differ in length"},
		{"tokens only", NewBalancePairsRequest(multicall, []*types.Address{t1}, nil, false), "invalid request: tokens and accounts differ in length"},
		{"accounts only", NewBalancePairsRequest(multicall, nil, []*types.Address{a1}, false), "invalid request: tokens and accounts differ in length"},
		{"mismatch before nil element", NewBalancePairsRequest(multicall, []*types.Address{nil, t2}, []*types.Address{a1}, false), "invalid request: tokens and accounts differ in length"},
		{"nil token first", NewBalancePairsRequest(multicall, []*types.Address{nil, t2}, []*types.Address{a1, a2}, false), "invalid request: tokens[0] is nil"},
		{"nil token last", NewBalancePairsRequest(multicall, []*types.Address{t1, nil}, []*types.Address{a1, a2}, false), "invalid request: tokens[1] is nil"},
		{"nil account first", NewBalancePairsRequest(multicall, []*types.Address{t1, t2}, []*types.Address{nil, a2}, false), "invalid request: accounts[0] is nil"},
		{"nil account last", NewBalancePairsRequest(multicall, []*types.Address{t1, t2}, []*types.Address{a1, nil}, false), "invalid request: accounts[1] is nil"},
		{"lower index reported first", NewBalancePairsRequest(multicall, []*types.Address{t1, nil}, []*types.Address{nil, a2}, false), "invalid request: accounts[0] is nil"},
		{"token before account at same index", NewBalancePairsRequest(multicall, []*types.Address{t1, nil}, []*types.Address{a1, nil}, false), "invalid request: tokens[1] is nil"},
		{"single pair", NewBalancePairsRequest(multicall, []*types.Address{t1}, []*types.Address{a1}, false), ""},
		{"repeated pairs", NewBalancePairsRequest(multicall, []*types.Address{t1, t1}, []*types.Address{a1, a1}, true), ""},
		{"valid", NewBalancePairsRequest(multicall, []*types.Address{t1, t2}, []*types.Address{a1, a2}, true), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			if tc.want == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, bind.ErrInvalidRequest)
			require.EqualError(t, err, tc.want)
		})
	}
}

func TestValidateDoesNotMutateRequest(t *testing.T) {
	multicall := types.MustNewAddressFromHex("0xca11bde05977b3631167028862be2a173976ca11")
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	transfer := NewTransferRequest(contract, a, big.NewInt(7))
	require.NoError(t, transfer.Validate())
	require.Equal(t, NewTransferRequest(contract, a, big.NewInt(7)), transfer)

	pairs := NewBalancePairsRequest(multicall, []*types.Address{contract, nil}, []*types.Address{a, a}, true)
	require.Error(t, pairs.Validate())
	require.Equal(t, NewBalancePairsRequest(multicall, []*types.Address{contract, nil}, []*types.Address{a, a}, true), pairs)
}

func TestClientMethodsWrapValidateError(t *testing.T) {
	multicall := types.MustNewAddressFromHex("0xca11bde05977b3631167028862be2a173976ca11")
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	e := New(mocks.NewClient(t))

	for _, tc := range []struct {
		method string
		call   func() error
		inner  error
	}{
		{
			"totalSupply",
			func() error {
				_, err := e.TotalSupply(t.Context(), nil, &TotalSupplyRequest{})
				return err
			},
			(&TotalSupplyRequest{}).Validate(),
		},
		{
			"balanceOf",
			func() error {
				_, err := e.BalanceOf(t.Context(), nil, NewBalanceOfRequest(contract, nil))
				return err
			},
			NewBalanceOfRequest(contract, nil).Validate(),
		},
		{
			"allowance",
			func() error {
				_, err := e.Allowance(t.Context(), nil, NewAllowanceRequest(contract, contract, nil))
				return err
			},
			NewAllowanceRequest(contract, contract, nil).Validate(),
		},
		{
			"name",
			func() error {
				_, err := e.Name(t.Context(), nil, nil)
				return err
			},
			(*NameRequest)(nil).Validate(),
		},
		{
			"symbol",
			func() error {
				_, err := e.Symbol(t.Context(), nil, nil)
				return err
			},
			(*SymbolRequest)(nil).Validate(),
		},
		{
			"decimals",
			func() error {
				_, err := e.Decimals(t.Context(), nil, nil)
				return err
			},
			(*DecimalsRequest)(nil).Validate(),
		},
		{
			"callTransfer",
			func() error {
				_, err := e.CallTransfer(t.Context(), nil, NewTransferRequest(contract, contract, nil))
				return err
			},
			NewTransferRequest(contract, contract, nil).Validate(),
		},
		{
			"callApprove",
			func() error {
				_, err := e.CallApprove(t.Context(), nil, NewApproveRequest(contract, nil, big.NewInt(1)))
				return err
			},
			NewApproveRequest(contract, nil, big.NewInt(1)).Validate(),
		},
		{
			"callTransferFrom",
			func() error {
				_, err := e.CallTransferFrom(t.Context(), nil, NewTransferFromRequest(contract, nil, contract, big.NewInt(1)))
				return err
			},
			NewTransferFromRequest(contract, nil, contract, big.NewInt(1)).Validate(),
		},
		{
			"transfer",
			func() error {
				_, _, err := e.Transfer(t.Context(), nil, NewTransferRequest(nil, contract, big.NewInt(1)))
				return err
			},
			NewTransferRequest(nil, contract, big.NewInt(1)).Validate(),
		},
		{
			"approve",
			func() error {
				_, _, err := e.Approve(t.Context(), nil, NewApproveRequest(contract, contract, nil))
				return err
			},
			NewApproveRequest(contract, contract, nil).Validate(),
		},
		{
			"transferFrom",
			func() error {
				_, _, err := e.TransferFrom(t.Context(), nil, NewTransferFromRequest(contract, contract, nil, big.NewInt(1)))
				return err
			},
			NewTransferFromRequest(contract, contract, nil, big.NewInt(1)).Validate(),
		},
		{
			"metadata",
			func() error {
				_, err := e.Metadata(t.Context(), nil, NewMetadataRequest(multicall, nil, false))
				return err
			},
			NewMetadataRequest(multicall, nil, false).Validate(),
		},
		{
			"metadataWithBalance",
			func() error {
				_, err := e.MetadataWithBalance(t.Context(), nil, NewMetadataWithBalanceRequest(multicall, contract, nil, false))
				return err
			},
			NewMetadataWithBalanceRequest(multicall, contract, nil, false).Validate(),
		},
		{
			"allowanceWithBalance",
			func() error {
				_, err := e.AllowanceWithBalance(t.Context(), nil, NewAllowanceWithBalanceRequest(nil, contract, contract, contract, false))
				return err
			},
			NewAllowanceWithBalanceRequest(nil, contract, contract, contract, false).Validate(),
		},
		{
			"tokenBalances",
			func() error {
				_, err := e.TokenBalances(t.Context(), nil, NewTokenBalancesRequest(multicall, contract, nil, false))
				return err
			},
			NewTokenBalancesRequest(multicall, contract, nil, false).Validate(),
		},
		{
			"addressBalances",
			func() error {
				_, err := e.AddressBalances(t.Context(), nil, NewAddressBalancesRequest(multicall, contract, []*types.Address{nil}, false))
				return err
			},
			NewAddressBalancesRequest(multicall, contract, []*types.Address{nil}, false).Validate(),
		},
		{
			"balancePairs",
			func() error {
				_, err := e.BalancePairs(t.Context(), nil, NewBalancePairsRequest(multicall, []*types.Address{contract}, nil, false))
				return err
			},
			NewBalancePairsRequest(multicall, []*types.Address{contract}, nil, false).Validate(),
		},
	} {
		t.Run(tc.method, func(t *testing.T) {
			require.Error(t, tc.inner)
			err := tc.call()
			require.ErrorIs(t, err, bind.ErrInvalidRequest)
			require.EqualError(t, err, "erc20: "+tc.method+": "+tc.inner.Error())
		})
	}
}
