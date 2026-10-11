package erc1155

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Transfer is IERC1155's TransferSingle event: TransferSingle(address indexed operator, address indexed from, address indexed to, uint256 id, uint256 value).
type Transfer struct {
	Contract *types.Address
	Operator *types.Address
	From     *types.Address
	To       *types.Address
	Id       *big.Int
	Value    *big.Int
}

func NewTransfer(contract, operator, from, to *types.Address, id, value *big.Int) *Transfer {
	return &Transfer{
		Contract: contract,
		Operator: operator,
		From:     from,
		To:       to,
		Id:       id,
		Value:    value,
	}
}

// TransferBatch is IERC1155's TransferBatch event: TransferBatch(address indexed operator, address indexed from, address indexed to, uint256[] ids, uint256[] values). Ids and Values are parallel arrays of equal length.
type TransferBatch struct {
	Contract *types.Address
	Operator *types.Address
	From     *types.Address
	To       *types.Address
	Ids      []*big.Int
	Values   []*big.Int
}

func NewTransferBatch(contract, operator, from, to *types.Address, ids, values []*big.Int) *TransferBatch {
	return &TransferBatch{
		Contract: contract,
		Operator: operator,
		From:     from,
		To:       to,
		Ids:      ids,
		Values:   values,
	}
}

// Expand splits b into one Transfer per (id, value) pair, as if each had been emitted as its own TransferSingle event.
func (b *TransferBatch) Expand() ([]Transfer, error) {
	if len(b.Ids) != len(b.Values) {
		return nil, fmt.Errorf("erc1155: expand: %d ids but %d values", len(b.Ids), len(b.Values))
	}

	transfers := make([]Transfer, len(b.Ids))
	for i := range b.Ids {
		transfers[i] = *NewTransfer(b.Contract, b.Operator, b.From, b.To, b.Ids[i], b.Values[i])
	}
	return transfers, nil
}

// ApprovalForAll is IERC1155's ApprovalForAll event: ApprovalForAll(address indexed account, address indexed operator, bool approved).
type ApprovalForAll struct {
	Contract *types.Address
	Account  *types.Address
	Operator *types.Address
	Approved bool
}

func NewApprovalForAll(contract, account, operator *types.Address, approved bool) *ApprovalForAll {
	return &ApprovalForAll{
		Contract: contract,
		Account:  account,
		Operator: operator,
		Approved: approved,
	}
}

// URI is IERC1155's URI event: URI(string value, uint256 indexed id). Only emitted by implementations that choose to (it's not required to keep the result of the uri function in sync).
type URI struct {
	Contract *types.Address
	Value    string
	Id       *big.Int
}

func NewURI(contract *types.Address, value string, id *big.Int) *URI {
	return &URI{
		Contract: contract,
		Value:    value,
		Id:       id,
	}
}

// IdBalance is one id's balanceOf result, one entry in a single contract's batch across many ids.
type IdBalance struct {
	Id     *big.Int
	Amount *big.Int
}

func NewIdBalance(id, amount *big.Int) *IdBalance {
	return &IdBalance{
		Id:     id,
		Amount: amount,
	}
}

// IdBalancePair is one (account, id) pair's balanceOf result, one entry in a balanceOfBatch(accounts, ids) call: accounts and ids are matched pairwise by index, not a full cross product.
type IdBalancePair struct {
	Account *types.Address
	Id      *big.Int
	Amount  *big.Int
}

func NewIdBalancePair(account *types.Address, id, amount *big.Int) *IdBalancePair {
	return &IdBalancePair{
		Account: account,
		Id:      id,
		Amount:  amount,
	}
}

// ContractBalance is one contract's balanceOf result for a fixed account and id, one entry in an AccountIdBalances batch; Amount is nil on a failed call.
type ContractBalance struct {
	Contract *types.Address
	Amount   *big.Int
}

func NewContractBalance(contract *types.Address, amount *big.Int) *ContractBalance {
	return &ContractBalance{
		Contract: contract,
		Amount:   amount,
	}
}

// AccountIdBalances is one (account, id) pair's balanceOf across many contracts, batched via multicall3 (balanceOf itself has no native cross-contract batch, unlike balanceOfBatch's cross-id one).
type AccountIdBalances struct {
	Account   *types.Address
	Id        *big.Int
	Contracts []ContractBalance
}

func NewAccountIdBalances(account *types.Address, id *big.Int, contracts []ContractBalance) *AccountIdBalances {
	return &AccountIdBalances{
		Account:   account,
		Id:        id,
		Contracts: contracts,
	}
}

// ApprovalForAllWithBalance is isApprovedForAll(account, operator) plus both their balanceOf(id) results, batched via multicall3 into one call.
type ApprovalForAllWithBalance struct {
	Account         *types.Address
	Operator        *types.Address
	Id              *big.Int
	Approved        bool
	AccountBalance  *big.Int
	OperatorBalance *big.Int
}

func NewApprovalForAllWithBalance(account, operator *types.Address, id *big.Int, approved bool, accountBalance, operatorBalance *big.Int) *ApprovalForAllWithBalance {
	return &ApprovalForAllWithBalance{
		Account:         account,
		Operator:        operator,
		Id:              id,
		Approved:        approved,
		AccountBalance:  accountBalance,
		OperatorBalance: operatorBalance,
	}
}
