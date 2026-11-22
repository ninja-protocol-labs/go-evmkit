package erc1155

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// URIRequest is uri(id)'s input on Contract.
type URIRequest struct {
	Contract *types.Address
	ID       *big.Int
}

func NewURIRequest(contract *types.Address, id *big.Int) *URIRequest {
	return &URIRequest{
		Contract: contract,
		ID:       id,
	}
}

func (r *URIRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.ID == nil {
		return fmt.Errorf("%w: id is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// BalanceOfRequest is balanceOf(account, id)'s input on Contract.
type BalanceOfRequest struct {
	Contract *types.Address
	Account  *types.Address
	ID       *big.Int
}

func NewBalanceOfRequest(contract, account *types.Address, id *big.Int) *BalanceOfRequest {
	return &BalanceOfRequest{
		Contract: contract,
		Account:  account,
		ID:       id,
	}
}

func (r *BalanceOfRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Account == nil {
		return fmt.Errorf("%w: account is nil", bind.ErrInvalidRequest)
	}

	if r.ID == nil {
		return fmt.Errorf("%w: id is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// BalanceOfBatchRequest is balanceOfBatch(accounts, ids)'s input on Contract, paired by index.
type BalanceOfBatchRequest struct {
	Contract *types.Address
	Accounts []*types.Address
	IDs      []*big.Int
}

func NewBalanceOfBatchRequest(contract *types.Address, accounts []*types.Address, ids []*big.Int) *BalanceOfBatchRequest {
	return &BalanceOfBatchRequest{
		Contract: contract,
		Accounts: accounts,
		IDs:      ids,
	}
}

func (r *BalanceOfBatchRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if len(r.Accounts) != len(r.IDs) {
		return fmt.Errorf("%w: accounts and ids differ in length", bind.ErrInvalidRequest)
	}

	if len(r.Accounts) == 0 {
		return fmt.Errorf("%w: accounts is empty", bind.ErrInvalidRequest)
	}

	for i := range r.Accounts {
		if r.Accounts[i] == nil {
			return fmt.Errorf("%w: accounts[%d] is nil", bind.ErrInvalidRequest, i)
		}

		if r.IDs[i] == nil {
			return fmt.Errorf("%w: ids[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

// IsApprovedForAllRequest is isApprovedForAll(account, operator)'s input on Contract.
type IsApprovedForAllRequest struct {
	Contract *types.Address
	Account  *types.Address
	Operator *types.Address
}

func NewIsApprovedForAllRequest(contract, account, operator *types.Address) *IsApprovedForAllRequest {
	return &IsApprovedForAllRequest{
		Contract: contract,
		Account:  account,
		Operator: operator,
	}
}

func (r *IsApprovedForAllRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Account == nil {
		return fmt.Errorf("%w: account is nil", bind.ErrInvalidRequest)
	}

	if r.Operator == nil {
		return fmt.Errorf("%w: operator is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SetApprovalForAllRequest is setApprovalForAll(operator, approved)'s input on Contract.
type SetApprovalForAllRequest struct {
	Contract *types.Address
	Operator *types.Address
	Approved bool
}

func NewSetApprovalForAllRequest(contract, operator *types.Address, approved bool) *SetApprovalForAllRequest {
	return &SetApprovalForAllRequest{
		Contract: contract,
		Operator: operator,
		Approved: approved,
	}
}

func (r *SetApprovalForAllRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Operator == nil {
		return fmt.Errorf("%w: operator is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SafeTransferFromRequest is safeTransferFrom(from, to, id, value, data)'s input on Contract; Data may be empty.
type SafeTransferFromRequest struct {
	Contract *types.Address
	From     *types.Address
	To       *types.Address
	ID       *big.Int
	Amount   *big.Int
	Data     []byte
}

func NewSafeTransferFromRequest(contract, from, to *types.Address, id, amount *big.Int, data []byte) *SafeTransferFromRequest {
	return &SafeTransferFromRequest{
		Contract: contract,
		From:     from,
		To:       to,
		ID:       id,
		Amount:   amount,
		Data:     data,
	}
}

func (r *SafeTransferFromRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.From == nil {
		return fmt.Errorf("%w: from is nil", bind.ErrInvalidRequest)
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	if r.ID == nil {
		return fmt.Errorf("%w: id is nil", bind.ErrInvalidRequest)
	}

	if r.Amount == nil {
		return fmt.Errorf("%w: amount is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SafeBatchTransferFromRequest is safeBatchTransferFrom(from, to, ids, values, data)'s input on Contract; Data may be empty.
type SafeBatchTransferFromRequest struct {
	Contract *types.Address
	From     *types.Address
	To       *types.Address
	IDs      []*big.Int
	Amounts  []*big.Int
	Data     []byte
}

func NewSafeBatchTransferFromRequest(contract, from, to *types.Address, ids, amounts []*big.Int, data []byte) *SafeBatchTransferFromRequest {
	return &SafeBatchTransferFromRequest{
		Contract: contract,
		From:     from,
		To:       to,
		IDs:      ids,
		Amounts:  amounts,
		Data:     data,
	}
}

func (r *SafeBatchTransferFromRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.From == nil {
		return fmt.Errorf("%w: from is nil", bind.ErrInvalidRequest)
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	if len(r.IDs) != len(r.Amounts) {
		return fmt.Errorf("%w: ids and amounts differ in length", bind.ErrInvalidRequest)
	}

	if len(r.IDs) == 0 {
		return fmt.Errorf("%w: ids is empty", bind.ErrInvalidRequest)
	}

	for i := range r.IDs {
		if r.IDs[i] == nil {
			return fmt.Errorf("%w: ids[%d] is nil", bind.ErrInvalidRequest, i)
		}

		if r.Amounts[i] == nil {
			return fmt.Errorf("%w: amounts[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

// IdBalancesRequest batches Account's balance of each of IDs on Contract via balanceOfBatch.
type IdBalancesRequest struct {
	Contract *types.Address
	Account  *types.Address
	IDs      []*big.Int
}

func NewIdBalancesRequest(contract, account *types.Address, ids []*big.Int) *IdBalancesRequest {
	return &IdBalancesRequest{
		Contract: contract,
		Account:  account,
		IDs:      ids,
	}
}

func (r *IdBalancesRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Account == nil {
		return fmt.Errorf("%w: account is nil", bind.ErrInvalidRequest)
	}

	if len(r.IDs) == 0 {
		return fmt.Errorf("%w: ids is empty", bind.ErrInvalidRequest)
	}

	for i, v := range r.IDs {
		if v == nil {
			return fmt.Errorf("%w: ids[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

// IdBalancePairsRequest batches Accounts[i]'s balance of IDs[i] on Contract via balanceOfBatch.
type IdBalancePairsRequest struct {
	Contract *types.Address
	Accounts []*types.Address
	IDs      []*big.Int
}

func NewIdBalancePairsRequest(contract *types.Address, accounts []*types.Address, ids []*big.Int) *IdBalancePairsRequest {
	return &IdBalancePairsRequest{
		Contract: contract,
		Accounts: accounts,
		IDs:      ids,
	}
}

func (r *IdBalancePairsRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if len(r.Accounts) != len(r.IDs) {
		return fmt.Errorf("%w: accounts and ids differ in length", bind.ErrInvalidRequest)
	}

	if len(r.Accounts) == 0 {
		return fmt.Errorf("%w: accounts is empty", bind.ErrInvalidRequest)
	}

	for i := range r.Accounts {
		if r.Accounts[i] == nil {
			return fmt.Errorf("%w: accounts[%d] is nil", bind.ErrInvalidRequest, i)
		}

		if r.IDs[i] == nil {
			return fmt.Errorf("%w: ids[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

// AccountIdBalancesRequest batches Account's balance of ID on each of Contracts via the Multicall3 at Multicall.
type AccountIdBalancesRequest struct {
	Multicall    *types.Address
	Account      *types.Address
	ID           *big.Int
	Contracts    []*types.Address
	AllowFailure bool
}

func NewAccountIdBalancesRequest(multicall, account *types.Address, id *big.Int, contracts []*types.Address, allowFailure bool) *AccountIdBalancesRequest {
	return &AccountIdBalancesRequest{
		Multicall:    multicall,
		Account:      account,
		ID:           id,
		Contracts:    contracts,
		AllowFailure: allowFailure,
	}
}

func (r *AccountIdBalancesRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Multicall == nil {
		return fmt.Errorf("%w: multicall is nil", bind.ErrInvalidRequest)
	}

	if r.Account == nil {
		return fmt.Errorf("%w: account is nil", bind.ErrInvalidRequest)
	}

	if r.ID == nil {
		return fmt.Errorf("%w: id is nil", bind.ErrInvalidRequest)
	}

	if len(r.Contracts) == 0 {
		return fmt.Errorf("%w: contracts is empty", bind.ErrInvalidRequest)
	}

	for i, v := range r.Contracts {
		if v == nil {
			return fmt.Errorf("%w: contracts[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

// ApprovalForAllWithBalanceRequest batches isApprovedForAll(Account, Operator) and both balances of ID on Contract.
type ApprovalForAllWithBalanceRequest struct {
	Multicall    *types.Address
	Contract     *types.Address
	Account      *types.Address
	Operator     *types.Address
	ID           *big.Int
	AllowFailure bool
}

func NewApprovalForAllWithBalanceRequest(multicall, contract, account, operator *types.Address, id *big.Int, allowFailure bool) *ApprovalForAllWithBalanceRequest {
	return &ApprovalForAllWithBalanceRequest{
		Multicall:    multicall,
		Contract:     contract,
		Account:      account,
		Operator:     operator,
		ID:           id,
		AllowFailure: allowFailure,
	}
}

func (r *ApprovalForAllWithBalanceRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Multicall == nil {
		return fmt.Errorf("%w: multicall is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Account == nil {
		return fmt.Errorf("%w: account is nil", bind.ErrInvalidRequest)
	}

	if r.Operator == nil {
		return fmt.Errorf("%w: operator is nil", bind.ErrInvalidRequest)
	}

	if r.ID == nil {
		return fmt.Errorf("%w: id is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// TransferSingleEvent is IERC1155's TransferSingle event: TransferSingle(address indexed operator, address indexed from, address indexed to, uint256 id, uint256 value).
type TransferSingleEvent struct {
	Contract *types.Address
	Operator *types.Address
	From     *types.Address
	To       *types.Address
	Id       *big.Int
	Value    *big.Int
}

func NewTransferSingleEvent(contract, operator, from, to *types.Address, id, value *big.Int) *TransferSingleEvent {
	return &TransferSingleEvent{
		Contract: contract,
		Operator: operator,
		From:     from,
		To:       to,
		Id:       id,
		Value:    value,
	}
}

// TransferBatchEvent is IERC1155's TransferBatch event: TransferBatch(address indexed operator, address indexed from, address indexed to, uint256[] ids, uint256[] values). Ids and Values are parallel arrays of equal length.
type TransferBatchEvent struct {
	Contract *types.Address
	Operator *types.Address
	From     *types.Address
	To       *types.Address
	Ids      []*big.Int
	Values   []*big.Int
}

func NewTransferBatchEvent(contract, operator, from, to *types.Address, ids, values []*big.Int) *TransferBatchEvent {
	return &TransferBatchEvent{
		Contract: contract,
		Operator: operator,
		From:     from,
		To:       to,
		Ids:      ids,
		Values:   values,
	}
}

// Expand splits b into one TransferSingleEvent per (id, value) pair, as if each had been emitted as its own TransferSingle event.
func (b *TransferBatchEvent) Expand() ([]TransferSingleEvent, error) {
	if len(b.Ids) != len(b.Values) {
		return nil, fmt.Errorf("erc1155: expand: %d ids but %d values", len(b.Ids), len(b.Values))
	}

	transfers := make([]TransferSingleEvent, len(b.Ids))
	for i := range b.Ids {
		transfers[i] = *NewTransferSingleEvent(b.Contract, b.Operator, b.From, b.To, b.Ids[i], b.Values[i])
	}
	return transfers, nil
}

// ApprovalForAllEvent is IERC1155's ApprovalForAll event: ApprovalForAll(address indexed account, address indexed operator, bool approved).
type ApprovalForAllEvent struct {
	Contract *types.Address
	Account  *types.Address
	Operator *types.Address
	Approved bool
}

func NewApprovalForAllEvent(contract, account, operator *types.Address, approved bool) *ApprovalForAllEvent {
	return &ApprovalForAllEvent{
		Contract: contract,
		Account:  account,
		Operator: operator,
		Approved: approved,
	}
}

// URIEvent is IERC1155's URI event: URI(string value, uint256 indexed id). Emitting it is optional,
// so it may lag behind the uri function.
type URIEvent struct {
	Contract *types.Address
	Value    string
	Id       *big.Int
}

func NewURIEvent(contract *types.Address, value string, id *big.Int) *URIEvent {
	return &URIEvent{
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

// AccountIdBalances is one (account, id) pair's balanceOf across many standards, batched via multicall3 (balanceOf itself has no native cross-contract batch, unlike balanceOfBatch's cross-id one).
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
