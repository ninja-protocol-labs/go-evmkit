package erc6909

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewOperatorSetEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	owner := types.MustNewAddressFromHex("0x1231deb6f5749ef6ce6943a275a1d3e7486f4eae")
	operator := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	os := NewOperatorSetEvent(contract, owner, operator, true)
	require.Same(t, contract, os.Contract)
	require.Same(t, owner, os.Owner)
	require.Same(t, operator, os.Operator)
	require.True(t, os.Approved)
}

func TestNewApprovalEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	owner := types.MustNewAddressFromHex("0x1231deb6f5749ef6ce6943a275a1d3e7486f4eae")
	spender := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	id := big.NewInt(1)
	amount := big.NewInt(100)

	a := NewApprovalEvent(contract, owner, spender, id, amount)
	require.Same(t, contract, a.Contract)
	require.Same(t, owner, a.Owner)
	require.Same(t, spender, a.Spender)
	require.Same(t, id, a.Id)
	require.Same(t, amount, a.Amount)
}

func TestNewTransferEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	caller := types.MustNewAddressFromHex("0x1231deb6f5749ef6ce6943a275a1d3e7486f4eae")
	sender := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	receiver := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	id := big.NewInt(42)
	amount := big.NewInt(100)

	tr := NewTransferEvent(contract, caller, sender, receiver, id, amount)
	require.Same(t, contract, tr.Contract)
	require.Same(t, caller, tr.Caller)
	require.Same(t, sender, tr.Sender)
	require.Same(t, receiver, tr.Receiver)
	require.Same(t, id, tr.Id)
	require.Same(t, amount, tr.Amount)
}

func TestNewNameUpdatedEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	id := big.NewInt(1)

	n := NewNameUpdatedEvent(contract, id, "evmkit test token")
	require.Same(t, contract, n.Contract)
	require.Same(t, id, n.Id)
	require.Equal(t, "evmkit test token", n.NewName)
}

func TestNewSymbolUpdatedEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	id := big.NewInt(1)

	s := NewSymbolUpdatedEvent(contract, id, "EVK6909")
	require.Same(t, contract, s.Contract)
	require.Same(t, id, s.Id)
	require.Equal(t, "EVK6909", s.NewSymbol)
}

func TestNewDecimalsUpdatedEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	id := big.NewInt(1)

	d := NewDecimalsUpdatedEvent(contract, id, 18)
	require.Same(t, contract, d.Contract)
	require.Same(t, id, d.Id)
	require.Equal(t, uint8(18), d.NewDecimals)
}

func TestRequestsValidate(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"balanceOf nil", (*BalanceOfRequest)(nil).Validate(), "invalid request: request is nil"},
		{"balanceOf empty", (&BalanceOfRequest{}).Validate(), "invalid request: contract is nil"},
		{"balanceOf nil contract", (&BalanceOfRequest{Owner: a, ID: big.NewInt(7)}).Validate(), "invalid request: contract is nil"},
		{"balanceOf nil owner", (&BalanceOfRequest{Contract: contract, ID: big.NewInt(7)}).Validate(), "invalid request: owner is nil"},
		{"balanceOf nil id", (&BalanceOfRequest{Contract: contract, Owner: a}).Validate(), "invalid request: id is nil"},
		{"balanceOf valid", (&BalanceOfRequest{Contract: contract, Owner: a, ID: big.NewInt(7)}).Validate(), ""},
		{"balanceOf zero id and amount", (&BalanceOfRequest{Contract: contract, Owner: a, ID: big.NewInt(0)}).Validate(), ""},
		{"allowance nil", (*AllowanceRequest)(nil).Validate(), "invalid request: request is nil"},
		{"allowance empty", (&AllowanceRequest{}).Validate(), "invalid request: contract is nil"},
		{"allowance nil contract", (&AllowanceRequest{Owner: a, Spender: b, ID: big.NewInt(7)}).Validate(), "invalid request: contract is nil"},
		{"allowance nil owner", (&AllowanceRequest{Contract: contract, Spender: b, ID: big.NewInt(7)}).Validate(), "invalid request: owner is nil"},
		{"allowance nil spender", (&AllowanceRequest{Contract: contract, Owner: a, ID: big.NewInt(7)}).Validate(), "invalid request: spender is nil"},
		{"allowance nil id", (&AllowanceRequest{Contract: contract, Owner: a, Spender: b}).Validate(), "invalid request: id is nil"},
		{"allowance valid", (&AllowanceRequest{Contract: contract, Owner: a, Spender: b, ID: big.NewInt(7)}).Validate(), ""},
		{"allowance zero id and amount", (&AllowanceRequest{Contract: contract, Owner: a, Spender: b, ID: big.NewInt(0)}).Validate(), ""},
		{"isOperator nil", (*IsOperatorRequest)(nil).Validate(), "invalid request: request is nil"},
		{"isOperator empty", (&IsOperatorRequest{}).Validate(), "invalid request: contract is nil"},
		{"isOperator nil contract", (&IsOperatorRequest{Owner: a, Spender: b}).Validate(), "invalid request: contract is nil"},
		{"isOperator nil owner", (&IsOperatorRequest{Contract: contract, Spender: b}).Validate(), "invalid request: owner is nil"},
		{"isOperator nil spender", (&IsOperatorRequest{Contract: contract, Owner: a}).Validate(), "invalid request: spender is nil"},
		{"isOperator valid", (&IsOperatorRequest{Contract: contract, Owner: a, Spender: b}).Validate(), ""},
		{"name nil", (*NameRequest)(nil).Validate(), "invalid request: request is nil"},
		{"name empty", (&NameRequest{}).Validate(), "invalid request: contract is nil"},
		{"name nil contract", (&NameRequest{ID: big.NewInt(7)}).Validate(), "invalid request: contract is nil"},
		{"name nil id", (&NameRequest{Contract: contract}).Validate(), "invalid request: id is nil"},
		{"name valid", (&NameRequest{Contract: contract, ID: big.NewInt(7)}).Validate(), ""},
		{"name zero id and amount", (&NameRequest{Contract: contract, ID: big.NewInt(0)}).Validate(), ""},
		{"symbol nil", (*SymbolRequest)(nil).Validate(), "invalid request: request is nil"},
		{"symbol empty", (&SymbolRequest{}).Validate(), "invalid request: contract is nil"},
		{"symbol nil contract", (&SymbolRequest{ID: big.NewInt(7)}).Validate(), "invalid request: contract is nil"},
		{"symbol nil id", (&SymbolRequest{Contract: contract}).Validate(), "invalid request: id is nil"},
		{"symbol valid", (&SymbolRequest{Contract: contract, ID: big.NewInt(7)}).Validate(), ""},
		{"symbol zero id and amount", (&SymbolRequest{Contract: contract, ID: big.NewInt(0)}).Validate(), ""},
		{"decimals nil", (*DecimalsRequest)(nil).Validate(), "invalid request: request is nil"},
		{"decimals empty", (&DecimalsRequest{}).Validate(), "invalid request: contract is nil"},
		{"decimals nil contract", (&DecimalsRequest{ID: big.NewInt(7)}).Validate(), "invalid request: contract is nil"},
		{"decimals nil id", (&DecimalsRequest{Contract: contract}).Validate(), "invalid request: id is nil"},
		{"decimals valid", (&DecimalsRequest{Contract: contract, ID: big.NewInt(7)}).Validate(), ""},
		{"decimals zero id and amount", (&DecimalsRequest{Contract: contract, ID: big.NewInt(0)}).Validate(), ""},
		{"approve nil", (*ApproveRequest)(nil).Validate(), "invalid request: request is nil"},
		{"approve empty", (&ApproveRequest{}).Validate(), "invalid request: contract is nil"},
		{"approve nil contract", (&ApproveRequest{Spender: b, ID: big.NewInt(7), Amount: big.NewInt(100)}).Validate(), "invalid request: contract is nil"},
		{"approve nil spender", (&ApproveRequest{Contract: contract, ID: big.NewInt(7), Amount: big.NewInt(100)}).Validate(), "invalid request: spender is nil"},
		{"approve nil id", (&ApproveRequest{Contract: contract, Spender: b, Amount: big.NewInt(100)}).Validate(), "invalid request: id is nil"},
		{"approve nil amount", (&ApproveRequest{Contract: contract, Spender: b, ID: big.NewInt(7)}).Validate(), "invalid request: amount is nil"},
		{"approve valid", (&ApproveRequest{Contract: contract, Spender: b, ID: big.NewInt(7), Amount: big.NewInt(100)}).Validate(), ""},
		{"approve zero id and amount", (&ApproveRequest{Contract: contract, Spender: b, ID: big.NewInt(0), Amount: big.NewInt(0)}).Validate(), ""},
		{"setOperator nil", (*SetOperatorRequest)(nil).Validate(), "invalid request: request is nil"},
		{"setOperator empty", (&SetOperatorRequest{}).Validate(), "invalid request: contract is nil"},
		{"setOperator nil contract", (&SetOperatorRequest{Spender: b, Approved: true}).Validate(), "invalid request: contract is nil"},
		{"setOperator nil spender", (&SetOperatorRequest{Contract: contract, Approved: true}).Validate(), "invalid request: spender is nil"},
		{"setOperator valid", (&SetOperatorRequest{Contract: contract, Spender: b, Approved: true}).Validate(), ""},
		{"setOperator approved false", (&SetOperatorRequest{Contract: contract, Spender: b}).Validate(), ""},
		{"transfer nil", (*TransferRequest)(nil).Validate(), "invalid request: request is nil"},
		{"transfer empty", (&TransferRequest{}).Validate(), "invalid request: contract is nil"},
		{"transfer nil contract", (&TransferRequest{Receiver: b, ID: big.NewInt(7), Amount: big.NewInt(100)}).Validate(), "invalid request: contract is nil"},
		{"transfer nil receiver", (&TransferRequest{Contract: contract, ID: big.NewInt(7), Amount: big.NewInt(100)}).Validate(), "invalid request: receiver is nil"},
		{"transfer nil id", (&TransferRequest{Contract: contract, Receiver: b, Amount: big.NewInt(100)}).Validate(), "invalid request: id is nil"},
		{"transfer nil amount", (&TransferRequest{Contract: contract, Receiver: b, ID: big.NewInt(7)}).Validate(), "invalid request: amount is nil"},
		{"transfer valid", (&TransferRequest{Contract: contract, Receiver: b, ID: big.NewInt(7), Amount: big.NewInt(100)}).Validate(), ""},
		{"transfer zero id and amount", (&TransferRequest{Contract: contract, Receiver: b, ID: big.NewInt(0), Amount: big.NewInt(0)}).Validate(), ""},
		{"transferFrom nil", (*TransferFromRequest)(nil).Validate(), "invalid request: request is nil"},
		{"transferFrom empty", (&TransferFromRequest{}).Validate(), "invalid request: contract is nil"},
		{"transferFrom nil contract", (&TransferFromRequest{Sender: a, Receiver: b, ID: big.NewInt(7), Amount: big.NewInt(100)}).Validate(), "invalid request: contract is nil"},
		{"transferFrom nil sender", (&TransferFromRequest{Contract: contract, Receiver: b, ID: big.NewInt(7), Amount: big.NewInt(100)}).Validate(), "invalid request: sender is nil"},
		{"transferFrom nil receiver", (&TransferFromRequest{Contract: contract, Sender: a, ID: big.NewInt(7), Amount: big.NewInt(100)}).Validate(), "invalid request: receiver is nil"},
		{"transferFrom nil id", (&TransferFromRequest{Contract: contract, Sender: a, Receiver: b, Amount: big.NewInt(100)}).Validate(), "invalid request: id is nil"},
		{"transferFrom nil amount", (&TransferFromRequest{Contract: contract, Sender: a, Receiver: b, ID: big.NewInt(7)}).Validate(), "invalid request: amount is nil"},
		{"transferFrom valid", (&TransferFromRequest{Contract: contract, Sender: a, Receiver: b, ID: big.NewInt(7), Amount: big.NewInt(100)}).Validate(), ""},
		{"transferFrom zero id and amount", (&TransferFromRequest{Contract: contract, Sender: a, Receiver: b, ID: big.NewInt(0), Amount: big.NewInt(0)}).Validate(), ""},
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
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	id := big.NewInt(7)
	amount := big.NewInt(100)

	require.Equal(t, &BalanceOfRequest{Contract: contract, Owner: a, ID: id}, NewBalanceOfRequest(contract, a, id))
	require.Equal(t, &AllowanceRequest{Contract: contract, Owner: a, Spender: b, ID: id}, NewAllowanceRequest(contract, a, b, id))
	require.Equal(t, &IsOperatorRequest{Contract: contract, Owner: a, Spender: b}, NewIsOperatorRequest(contract, a, b))
	require.Equal(t, &NameRequest{Contract: contract, ID: id}, NewNameRequest(contract, id))
	require.Equal(t, &SymbolRequest{Contract: contract, ID: id}, NewSymbolRequest(contract, id))
	require.Equal(t, &DecimalsRequest{Contract: contract, ID: id}, NewDecimalsRequest(contract, id))
	require.Equal(t, &ApproveRequest{Contract: contract, Spender: b, ID: id, Amount: amount}, NewApproveRequest(contract, b, id, amount))
	require.Equal(t, &SetOperatorRequest{Contract: contract, Spender: b, Approved: true}, NewSetOperatorRequest(contract, b, true))
	require.Equal(t, &TransferRequest{Contract: contract, Receiver: b, ID: id, Amount: amount}, NewTransferRequest(contract, b, id, amount))
	require.Equal(t, &TransferFromRequest{Contract: contract, Sender: a, Receiver: b, ID: id, Amount: amount}, NewTransferFromRequest(contract, a, b, id, amount))
}
