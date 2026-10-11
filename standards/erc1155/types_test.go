package erc1155

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewTransfer(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	operator := types.MustNewAddressFromHex("0x1231deb6f5749ef6ce6943a275a1d3e7486f4eae")
	from := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	to := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	id := big.NewInt(42)
	value := big.NewInt(100)

	tr := NewTransfer(contract, operator, from, to, id, value)
	require.Same(t, contract, tr.Contract)
	require.Same(t, operator, tr.Operator)
	require.Same(t, from, tr.From)
	require.Same(t, to, tr.To)
	require.Same(t, id, tr.Id)
	require.Same(t, value, tr.Value)
}

func TestNewTransferBatch(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	operator := types.MustNewAddressFromHex("0x1231deb6f5749ef6ce6943a275a1d3e7486f4eae")
	from := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	to := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	ids := []*big.Int{big.NewInt(1), big.NewInt(2)}
	values := []*big.Int{big.NewInt(10), big.NewInt(20)}

	tb := NewTransferBatch(contract, operator, from, to, ids, values)
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

	tb := NewTransferBatch(contract, operator, from, to, ids, values)
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
	tb := NewTransferBatch(nil, nil, nil, nil, nil, nil)
	transfers, err := tb.Expand()
	require.NoError(t, err)
	require.NotNil(t, transfers)
	require.Empty(t, transfers)
}

func TestTransferBatchExpandLengthMismatch(t *testing.T) {
	tb := NewTransferBatch(nil, nil, nil, nil, []*big.Int{big.NewInt(1), big.NewInt(2)}, []*big.Int{big.NewInt(1)})
	transfers, err := tb.Expand()
	require.Error(t, err)
	require.Nil(t, transfers)
}

func TestNewApprovalForAll(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	account := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	operator := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	afa := NewApprovalForAll(contract, account, operator, true)
	require.Same(t, contract, afa.Contract)
	require.Same(t, account, afa.Account)
	require.Same(t, operator, afa.Operator)
	require.True(t, afa.Approved)
}

func TestNewURI(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	id := big.NewInt(7)

	u := NewURI(contract, "ipfs://1234", id)
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
