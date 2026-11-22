package erc1155

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
)

func TestNewTransferSingleEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	operator := types.MustNewAddressFromHex("0x1231deb6f5749ef6ce6943a275a1d3e7486f4eae")
	from := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	to := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	id := big.NewInt(42)
	value := big.NewInt(100)

	tr := NewTransferSingleEvent(contract, operator, from, to, id, value)
	require.Same(t, contract, tr.Contract)
	require.Same(t, operator, tr.Operator)
	require.Same(t, from, tr.From)
	require.Same(t, to, tr.To)
	require.Same(t, id, tr.Id)
	require.Same(t, value, tr.Value)
}

func TestNewTransferBatchEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	operator := types.MustNewAddressFromHex("0x1231deb6f5749ef6ce6943a275a1d3e7486f4eae")
	from := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	to := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	ids := []*big.Int{big.NewInt(1), big.NewInt(2)}
	values := []*big.Int{big.NewInt(10), big.NewInt(20)}

	tb := NewTransferBatchEvent(contract, operator, from, to, ids, values)
	require.Same(t, contract, tb.Contract)
	require.Same(t, operator, tb.Operator)
	require.Same(t, from, tb.From)
	require.Same(t, to, tb.To)
	require.Equal(t, ids, tb.Ids)
	require.Equal(t, values, tb.Values)
}

func TestTransferBatchExpand(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	operator := types.MustNewAddressFromHex("0x1231deb6f5749ef6ce6943a275a1d3e7486f4eae")
	from := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	to := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	ids := []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3)}
	values := []*big.Int{big.NewInt(10), big.NewInt(20), big.NewInt(30)}

	tb := NewTransferBatchEvent(contract, operator, from, to, ids, values)
	transfers, err := tb.Expand()
	require.NoError(t, err)
	require.Len(t, transfers, 3)
	for i, tr := range transfers {
		require.Same(t, contract, tr.Contract)
		require.Same(t, operator, tr.Operator)
		require.Same(t, from, tr.From)
		require.Same(t, to, tr.To)
		require.Same(t, ids[i], tr.Id)
		require.Same(t, values[i], tr.Value)
	}
}

func TestTransferBatchExpandEmpty(t *testing.T) {
	tb := NewTransferBatchEvent(nil, nil, nil, nil, nil, nil)
	transfers, err := tb.Expand()
	require.NoError(t, err)
	require.NotNil(t, transfers)
	require.Empty(t, transfers)
}

func TestTransferBatchExpandLengthMismatch(t *testing.T) {
	tb := NewTransferBatchEvent(nil, nil, nil, nil, []*big.Int{big.NewInt(1), big.NewInt(2)}, []*big.Int{big.NewInt(1)})
	transfers, err := tb.Expand()
	require.Error(t, err)
	require.Nil(t, transfers)
}

func TestNewApprovalForAllEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	account := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	operator := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	afa := NewApprovalForAllEvent(contract, account, operator, true)
	require.Same(t, contract, afa.Contract)
	require.Same(t, account, afa.Account)
	require.Same(t, operator, afa.Operator)
	require.True(t, afa.Approved)
}

func TestNewURIEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	id := big.NewInt(7)

	u := NewURIEvent(contract, "ipfs://1234", id)
	require.Same(t, contract, u.Contract)
	require.Equal(t, "ipfs://1234", u.Value)
	require.Same(t, id, u.Id)
}

func TestNewIdBalance(t *testing.T) {
	id := big.NewInt(1)
	amount := big.NewInt(100)

	ib := NewIdBalance(id, amount)
	require.Same(t, id, ib.Id)
	require.Same(t, amount, ib.Amount)
}

func TestNewIdBalancePair(t *testing.T) {
	account := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	id := big.NewInt(2)
	amount := big.NewInt(200)

	p := NewIdBalancePair(account, id, amount)
	require.Same(t, account, p.Account)
	require.Same(t, id, p.Id)
	require.Same(t, amount, p.Amount)
}

func TestNewContractBalance(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	amount := big.NewInt(300)

	cb := NewContractBalance(contract, amount)
	require.Same(t, contract, cb.Contract)
	require.Same(t, amount, cb.Amount)
}

func TestNewAccountIdBalances(t *testing.T) {
	account := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	id := big.NewInt(3)
	contracts := []ContractBalance{
		*NewContractBalance(types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d"), big.NewInt(1)),
		*NewContractBalance(types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8"), big.NewInt(2)),
	}

	aib := NewAccountIdBalances(account, id, contracts)
	require.Same(t, account, aib.Account)
	require.Same(t, id, aib.Id)
	require.Equal(t, contracts, aib.Contracts)
}

func TestNewApprovalForAllWithBalance(t *testing.T) {
	account := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	operator := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	id := big.NewInt(4)
	accountBalance := big.NewInt(3)
	operatorBalance := big.NewInt(4)

	afawb := NewApprovalForAllWithBalance(account, operator, id, true, accountBalance, operatorBalance)
	require.Same(t, account, afawb.Account)
	require.Same(t, operator, afawb.Operator)
	require.Same(t, id, afawb.Id)
	require.True(t, afawb.Approved)
	require.Same(t, accountBalance, afawb.AccountBalance)
	require.Same(t, operatorBalance, afawb.OperatorBalance)
}

func TestScalarRequestsValidate(t *testing.T) {
	multicall := types.MustNewAddressFromHex("0xca11bde05977b3631167028862be2a173976ca11")
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"uri nil", (*URIRequest)(nil).Validate(), "invalid request: request is nil"},
		{"uri empty", (&URIRequest{}).Validate(), "invalid request: contract is nil"},
		{"uri nil contract", (&URIRequest{ID: big.NewInt(7)}).Validate(), "invalid request: contract is nil"},
		{"uri nil id", (&URIRequest{Contract: contract}).Validate(), "invalid request: id is nil"},
		{"uri valid", (&URIRequest{Contract: contract, ID: big.NewInt(7)}).Validate(), ""},
		{"uri zero id and amount", (&URIRequest{Contract: contract, ID: big.NewInt(0)}).Validate(), ""},
		{"balanceOf nil", (*BalanceOfRequest)(nil).Validate(), "invalid request: request is nil"},
		{"balanceOf empty", (&BalanceOfRequest{}).Validate(), "invalid request: contract is nil"},
		{"balanceOf nil contract", (&BalanceOfRequest{Account: a, ID: big.NewInt(7)}).Validate(), "invalid request: contract is nil"},
		{"balanceOf nil account", (&BalanceOfRequest{Contract: contract, ID: big.NewInt(7)}).Validate(), "invalid request: account is nil"},
		{"balanceOf nil id", (&BalanceOfRequest{Contract: contract, Account: a}).Validate(), "invalid request: id is nil"},
		{"balanceOf valid", (&BalanceOfRequest{Contract: contract, Account: a, ID: big.NewInt(7)}).Validate(), ""},
		{"balanceOf zero id and amount", (&BalanceOfRequest{Contract: contract, Account: a, ID: big.NewInt(0)}).Validate(), ""},
		{"isApprovedForAll nil", (*IsApprovedForAllRequest)(nil).Validate(), "invalid request: request is nil"},
		{"isApprovedForAll empty", (&IsApprovedForAllRequest{}).Validate(), "invalid request: contract is nil"},
		{"isApprovedForAll nil contract", (&IsApprovedForAllRequest{Account: a, Operator: b}).Validate(), "invalid request: contract is nil"},
		{"isApprovedForAll nil account", (&IsApprovedForAllRequest{Contract: contract, Operator: b}).Validate(), "invalid request: account is nil"},
		{"isApprovedForAll nil operator", (&IsApprovedForAllRequest{Contract: contract, Account: a}).Validate(), "invalid request: operator is nil"},
		{"isApprovedForAll valid", (&IsApprovedForAllRequest{Contract: contract, Account: a, Operator: b}).Validate(), ""},
		{"setApprovalForAll nil", (*SetApprovalForAllRequest)(nil).Validate(), "invalid request: request is nil"},
		{"setApprovalForAll empty", (&SetApprovalForAllRequest{}).Validate(), "invalid request: contract is nil"},
		{"setApprovalForAll nil contract", (&SetApprovalForAllRequest{Operator: b, Approved: true}).Validate(), "invalid request: contract is nil"},
		{"setApprovalForAll nil operator", (&SetApprovalForAllRequest{Contract: contract, Approved: true}).Validate(), "invalid request: operator is nil"},
		{"setApprovalForAll valid", (&SetApprovalForAllRequest{Contract: contract, Operator: b, Approved: true}).Validate(), ""},
		{"setApprovalForAll zero optional fields", (&SetApprovalForAllRequest{Contract: contract, Operator: b}).Validate(), ""},
		{"safeTransferFrom nil", (*SafeTransferFromRequest)(nil).Validate(), "invalid request: request is nil"},
		{"safeTransferFrom empty", (&SafeTransferFromRequest{}).Validate(), "invalid request: contract is nil"},
		{"safeTransferFrom nil contract", (&SafeTransferFromRequest{From: a, To: b, ID: big.NewInt(7), Amount: big.NewInt(100), Data: []byte{0x01}}).Validate(), "invalid request: contract is nil"},
		{"safeTransferFrom nil from", (&SafeTransferFromRequest{Contract: contract, To: b, ID: big.NewInt(7), Amount: big.NewInt(100), Data: []byte{0x01}}).Validate(), "invalid request: from is nil"},
		{"safeTransferFrom nil to", (&SafeTransferFromRequest{Contract: contract, From: a, ID: big.NewInt(7), Amount: big.NewInt(100), Data: []byte{0x01}}).Validate(), "invalid request: to is nil"},
		{"safeTransferFrom nil id", (&SafeTransferFromRequest{Contract: contract, From: a, To: b, Amount: big.NewInt(100), Data: []byte{0x01}}).Validate(), "invalid request: id is nil"},
		{"safeTransferFrom nil amount", (&SafeTransferFromRequest{Contract: contract, From: a, To: b, ID: big.NewInt(7), Data: []byte{0x01}}).Validate(), "invalid request: amount is nil"},
		{"safeTransferFrom valid", (&SafeTransferFromRequest{Contract: contract, From: a, To: b, ID: big.NewInt(7), Amount: big.NewInt(100), Data: []byte{0x01}}).Validate(), ""},
		{"safeTransferFrom zero id and amount", (&SafeTransferFromRequest{Contract: contract, From: a, To: b, ID: big.NewInt(0), Amount: big.NewInt(0), Data: []byte{0x01}}).Validate(), ""},
		{"safeTransferFrom zero optional fields", (&SafeTransferFromRequest{Contract: contract, From: a, To: b, ID: big.NewInt(7), Amount: big.NewInt(100)}).Validate(), ""},
		{"approvalForAllWithBalance nil", (*ApprovalForAllWithBalanceRequest)(nil).Validate(), "invalid request: request is nil"},
		{"approvalForAllWithBalance empty", (&ApprovalForAllWithBalanceRequest{}).Validate(), "invalid request: multicall is nil"},
		{"approvalForAllWithBalance nil multicall", (&ApprovalForAllWithBalanceRequest{Contract: contract, Account: a, Operator: b, ID: big.NewInt(7), AllowFailure: true}).Validate(), "invalid request: multicall is nil"},
		{"approvalForAllWithBalance nil contract", (&ApprovalForAllWithBalanceRequest{Multicall: multicall, Account: a, Operator: b, ID: big.NewInt(7), AllowFailure: true}).Validate(), "invalid request: contract is nil"},
		{"approvalForAllWithBalance nil account", (&ApprovalForAllWithBalanceRequest{Multicall: multicall, Contract: contract, Operator: b, ID: big.NewInt(7), AllowFailure: true}).Validate(), "invalid request: account is nil"},
		{"approvalForAllWithBalance nil operator", (&ApprovalForAllWithBalanceRequest{Multicall: multicall, Contract: contract, Account: a, ID: big.NewInt(7), AllowFailure: true}).Validate(), "invalid request: operator is nil"},
		{"approvalForAllWithBalance nil id", (&ApprovalForAllWithBalanceRequest{Multicall: multicall, Contract: contract, Account: a, Operator: b, AllowFailure: true}).Validate(), "invalid request: id is nil"},
		{"approvalForAllWithBalance valid", (&ApprovalForAllWithBalanceRequest{Multicall: multicall, Contract: contract, Account: a, Operator: b, ID: big.NewInt(7), AllowFailure: true}).Validate(), ""},
		{"approvalForAllWithBalance zero id and amount", (&ApprovalForAllWithBalanceRequest{Multicall: multicall, Contract: contract, Account: a, Operator: b, ID: big.NewInt(0), AllowFailure: true}).Validate(), ""},
		{"approvalForAllWithBalance zero optional fields", (&ApprovalForAllWithBalanceRequest{Multicall: multicall, Contract: contract, Account: a, Operator: b, ID: big.NewInt(7)}).Validate(), ""},
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

func TestListRequestsValidate(t *testing.T) {
	multicall := types.MustNewAddressFromHex("0xca11bde05977b3631167028862be2a173976ca11")
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	one := big.NewInt(1)
	two := big.NewInt(2)

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"balanceOfBatch nil", (*BalanceOfBatchRequest)(nil).Validate(), "invalid request: request is nil"},
		{"balanceOfBatch nil contract", NewBalanceOfBatchRequest(nil, []*types.Address{a}, []*big.Int{one}).Validate(), "invalid request: contract is nil"},
		{"balanceOfBatch accounts longer", NewBalanceOfBatchRequest(contract, []*types.Address{a, b}, []*big.Int{one}).Validate(), "invalid request: accounts and ids differ in length"},
		{"balanceOfBatch ids longer", NewBalanceOfBatchRequest(contract, []*types.Address{a}, []*big.Int{one, two}).Validate(), "invalid request: accounts and ids differ in length"},
		{"balanceOfBatch both empty", NewBalanceOfBatchRequest(contract, nil, nil).Validate(), "invalid request: accounts is empty"},
		{"balanceOfBatch nil account", NewBalanceOfBatchRequest(contract, []*types.Address{a, nil}, []*big.Int{one, two}).Validate(), "invalid request: accounts[1] is nil"},
		{"balanceOfBatch nil id", NewBalanceOfBatchRequest(contract, []*types.Address{a, b}, []*big.Int{nil, two}).Validate(), "invalid request: ids[0] is nil"},
		{"balanceOfBatch lower index first", NewBalanceOfBatchRequest(contract, []*types.Address{a, nil}, []*big.Int{nil, two}).Validate(), "invalid request: ids[0] is nil"},
		{"balanceOfBatch account before id at same index", NewBalanceOfBatchRequest(contract, []*types.Address{nil, b}, []*big.Int{nil, two}).Validate(), "invalid request: accounts[0] is nil"},
		{"balanceOfBatch valid", NewBalanceOfBatchRequest(contract, []*types.Address{a, a}, []*big.Int{one, one}).Validate(), ""},
		{"safeBatchTransferFrom nil", (*SafeBatchTransferFromRequest)(nil).Validate(), "invalid request: request is nil"},
		{"safeBatchTransferFrom nil contract", NewSafeBatchTransferFromRequest(nil, a, b, []*big.Int{one}, []*big.Int{one}, nil).Validate(), "invalid request: contract is nil"},
		{"safeBatchTransferFrom nil from", NewSafeBatchTransferFromRequest(contract, nil, b, []*big.Int{one}, []*big.Int{one}, nil).Validate(), "invalid request: from is nil"},
		{"safeBatchTransferFrom nil to", NewSafeBatchTransferFromRequest(contract, a, nil, []*big.Int{one}, []*big.Int{one}, nil).Validate(), "invalid request: to is nil"},
		{"safeBatchTransferFrom length mismatch", NewSafeBatchTransferFromRequest(contract, a, b, []*big.Int{one, two}, []*big.Int{one}, nil).Validate(), "invalid request: ids and amounts differ in length"},
		{"safeBatchTransferFrom empty", NewSafeBatchTransferFromRequest(contract, a, b, []*big.Int{}, []*big.Int{}, nil).Validate(), "invalid request: ids is empty"},
		{"safeBatchTransferFrom nil id", NewSafeBatchTransferFromRequest(contract, a, b, []*big.Int{one, nil}, []*big.Int{one, two}, nil).Validate(), "invalid request: ids[1] is nil"},
		{"safeBatchTransferFrom nil amount", NewSafeBatchTransferFromRequest(contract, a, b, []*big.Int{one, two}, []*big.Int{nil, two}, nil).Validate(), "invalid request: amounts[0] is nil"},
		{"safeBatchTransferFrom zero amounts and data", NewSafeBatchTransferFromRequest(contract, a, b, []*big.Int{one, two}, []*big.Int{big.NewInt(0), big.NewInt(0)}, []byte{0xca, 0xfe}).Validate(), ""},
		{"idBalances nil", (*IdBalancesRequest)(nil).Validate(), "invalid request: request is nil"},
		{"idBalances nil contract", NewIdBalancesRequest(nil, a, []*big.Int{one}).Validate(), "invalid request: contract is nil"},
		{"idBalances nil account", NewIdBalancesRequest(contract, nil, []*big.Int{one}).Validate(), "invalid request: account is nil"},
		{"idBalances empty", NewIdBalancesRequest(contract, a, nil).Validate(), "invalid request: ids is empty"},
		{"idBalances nil last", NewIdBalancesRequest(contract, a, []*big.Int{one, two, nil}).Validate(), "invalid request: ids[2] is nil"},
		{"idBalances valid", NewIdBalancesRequest(contract, a, []*big.Int{one, one, two}).Validate(), ""},
		{"idBalancePairs nil", (*IdBalancePairsRequest)(nil).Validate(), "invalid request: request is nil"},
		{"idBalancePairs nil contract", NewIdBalancePairsRequest(nil, []*types.Address{a}, []*big.Int{one}).Validate(), "invalid request: contract is nil"},
		{"idBalancePairs length mismatch", NewIdBalancePairsRequest(contract, []*types.Address{a}, nil).Validate(), "invalid request: accounts and ids differ in length"},
		{"idBalancePairs empty", NewIdBalancePairsRequest(contract, []*types.Address{}, []*big.Int{}).Validate(), "invalid request: accounts is empty"},
		{"idBalancePairs nil account", NewIdBalancePairsRequest(contract, []*types.Address{nil}, []*big.Int{one}).Validate(), "invalid request: accounts[0] is nil"},
		{"idBalancePairs valid", NewIdBalancePairsRequest(contract, []*types.Address{a, b}, []*big.Int{one, two}).Validate(), ""},
		{"accountIdBalances nil", (*AccountIdBalancesRequest)(nil).Validate(), "invalid request: request is nil"},
		{"accountIdBalances nil multicall", NewAccountIdBalancesRequest(nil, a, one, []*types.Address{contract}, false).Validate(), "invalid request: multicall is nil"},
		{"accountIdBalances nil account", NewAccountIdBalancesRequest(multicall, nil, one, []*types.Address{contract}, false).Validate(), "invalid request: account is nil"},
		{"accountIdBalances nil id", NewAccountIdBalancesRequest(multicall, a, nil, []*types.Address{contract}, false).Validate(), "invalid request: id is nil"},
		{"accountIdBalances empty contracts", NewAccountIdBalancesRequest(multicall, a, one, []*types.Address{}, false).Validate(), "invalid request: contracts is empty"},
		{"accountIdBalances nil contract entry", NewAccountIdBalancesRequest(multicall, a, one, []*types.Address{contract, nil}, false).Validate(), "invalid request: contracts[1] is nil"},
		{"accountIdBalances valid", NewAccountIdBalancesRequest(multicall, a, big.NewInt(0), []*types.Address{contract, contract}, true).Validate(), ""},
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
	multicall := types.MustNewAddressFromHex("0xca11bde05977b3631167028862be2a173976ca11")
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	id := big.NewInt(42)
	amount := big.NewInt(9)
	ids := []*big.Int{big.NewInt(1), big.NewInt(2)}
	amounts := []*big.Int{big.NewInt(3), big.NewInt(4)}
	addrs := []*types.Address{a, b}

	require.Equal(t, &URIRequest{Contract: contract, ID: id}, NewURIRequest(contract, id))
	require.Equal(t, &BalanceOfRequest{Contract: contract, Account: a, ID: id}, NewBalanceOfRequest(contract, a, id))
	require.Equal(t, &BalanceOfBatchRequest{Contract: contract, Accounts: addrs, IDs: ids}, NewBalanceOfBatchRequest(contract, addrs, ids))
	require.Equal(t, &IsApprovedForAllRequest{Contract: contract, Account: a, Operator: b}, NewIsApprovedForAllRequest(contract, a, b))
	require.Equal(t, &SetApprovalForAllRequest{Contract: contract, Operator: b, Approved: true}, NewSetApprovalForAllRequest(contract, b, true))
	require.Equal(t, &SafeTransferFromRequest{Contract: contract, From: a, To: b, ID: id, Amount: amount, Data: []byte{0x01}}, NewSafeTransferFromRequest(contract, a, b, id, amount, []byte{0x01}))
	require.Equal(t, &SafeBatchTransferFromRequest{Contract: contract, From: a, To: b, IDs: ids, Amounts: amounts}, NewSafeBatchTransferFromRequest(contract, a, b, ids, amounts, nil))
	require.Equal(t, &IdBalancesRequest{Contract: contract, Account: a, IDs: ids}, NewIdBalancesRequest(contract, a, ids))
	require.Equal(t, &IdBalancePairsRequest{Contract: contract, Accounts: addrs, IDs: ids}, NewIdBalancePairsRequest(contract, addrs, ids))
	require.Equal(t, &AccountIdBalancesRequest{Multicall: multicall, Account: a, ID: id, Contracts: addrs, AllowFailure: true}, NewAccountIdBalancesRequest(multicall, a, id, addrs, true))
	require.Equal(t, &ApprovalForAllWithBalanceRequest{Multicall: multicall, Contract: contract, Account: a, Operator: b, ID: id}, NewApprovalForAllWithBalanceRequest(multicall, contract, a, b, id, false))
}

func TestClientMethodsWrapValidateError(t *testing.T) {
	multicall := types.MustNewAddressFromHex("0xca11bde05977b3631167028862be2a173976ca11")
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	e := New(mocks.NewClient(t))
	ctx := t.Context()

	_, err := e.URI(ctx, nil, NewURIRequest(contract, nil))
	require.EqualError(t, err, "erc1155: uri: "+NewURIRequest(contract, nil).Validate().Error())
	_, err = e.BalanceOf(ctx, nil, NewBalanceOfRequest(contract, nil, big.NewInt(1)))
	require.EqualError(t, err, "erc1155: balanceOf: "+NewBalanceOfRequest(contract, nil, big.NewInt(1)).Validate().Error())
	_, err = e.BalanceOfBatch(ctx, nil, NewBalanceOfBatchRequest(contract, nil, nil))
	require.EqualError(t, err, "erc1155: balanceOfBatch: "+NewBalanceOfBatchRequest(contract, nil, nil).Validate().Error())
	_, err = e.IsApprovedForAll(ctx, nil, NewIsApprovedForAllRequest(nil, contract, contract))
	require.EqualError(t, err, "erc1155: isApprovedForAll: "+NewIsApprovedForAllRequest(nil, contract, contract).Validate().Error())
	_, _, err = e.SetApprovalForAll(ctx, nil, NewSetApprovalForAllRequest(contract, nil, true))
	require.EqualError(t, err, "erc1155: setApprovalForAll: "+NewSetApprovalForAllRequest(contract, nil, true).Validate().Error())
	_, _, err = e.SafeTransferFrom(ctx, nil, NewSafeTransferFromRequest(contract, contract, contract, big.NewInt(1), nil, nil))
	require.EqualError(t, err, "erc1155: safeTransferFrom: "+NewSafeTransferFromRequest(contract, contract, contract, big.NewInt(1), nil, nil).Validate().Error())
	_, _, err = e.SafeBatchTransferFrom(ctx, nil, NewSafeBatchTransferFromRequest(contract, contract, contract, []*big.Int{big.NewInt(1)}, nil, nil))
	require.EqualError(t, err, "erc1155: safeBatchTransferFrom: "+NewSafeBatchTransferFromRequest(contract, contract, contract, []*big.Int{big.NewInt(1)}, nil, nil).Validate().Error())
	_, err = e.IdBalances(ctx, nil, NewIdBalancesRequest(contract, contract, []*big.Int{nil}))
	require.EqualError(t, err, "erc1155: idBalances: "+NewIdBalancesRequest(contract, contract, []*big.Int{nil}).Validate().Error())
	_, err = e.IdBalancePairs(ctx, nil, NewIdBalancePairsRequest(nil, nil, nil))
	require.EqualError(t, err, "erc1155: idBalancePairs: "+NewIdBalancePairsRequest(nil, nil, nil).Validate().Error())
	_, err = e.AccountIdBalances(ctx, nil, NewAccountIdBalancesRequest(multicall, contract, nil, []*types.Address{contract}, false))
	require.EqualError(t, err, "erc1155: accountIdBalances: "+NewAccountIdBalancesRequest(multicall, contract, nil, []*types.Address{contract}, false).Validate().Error())
	_, err = e.ApprovalForAllWithBalance(ctx, nil, NewApprovalForAllWithBalanceRequest(nil, contract, contract, contract, big.NewInt(1), false))
	require.EqualError(t, err, "erc1155: approvalForAllWithBalance: "+NewApprovalForAllWithBalanceRequest(nil, contract, contract, contract, big.NewInt(1), false).Validate().Error())
	require.ErrorIs(t, err, bind.ErrInvalidRequest)

	_, _, err = e.SetApprovalForAll(ctx, nil, NewSetApprovalForAllRequest(contract, contract, true))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "erc1155: setApprovalForAll: bind: transact: invalid options: options is nil")
}
