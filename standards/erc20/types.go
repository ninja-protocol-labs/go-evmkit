package erc20

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// TotalSupplyRequest is totalSupply()'s input on Contract.
type TotalSupplyRequest struct {
	Contract *types.Address
}

func NewTotalSupplyRequest(contract *types.Address) *TotalSupplyRequest {
	return &TotalSupplyRequest{
		Contract: contract,
	}
}

func (r *TotalSupplyRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// BalanceOfRequest is balanceOf(account)'s input on Contract.
type BalanceOfRequest struct {
	Contract *types.Address
	Account  *types.Address
}

func NewBalanceOfRequest(contract, account *types.Address) *BalanceOfRequest {
	return &BalanceOfRequest{
		Contract: contract,
		Account:  account,
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

	return nil
}

// AllowanceRequest is allowance(owner, spender)'s input on Contract.
type AllowanceRequest struct {
	Contract *types.Address
	Owner    *types.Address
	Spender  *types.Address
}

func NewAllowanceRequest(contract, owner, spender *types.Address) *AllowanceRequest {
	return &AllowanceRequest{
		Contract: contract,
		Owner:    owner,
		Spender:  spender,
	}
}

func (r *AllowanceRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Owner == nil {
		return fmt.Errorf("%w: owner is nil", bind.ErrInvalidRequest)
	}

	if r.Spender == nil {
		return fmt.Errorf("%w: spender is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// NameRequest is name()'s input on Contract.
type NameRequest struct {
	Contract *types.Address
}

func NewNameRequest(contract *types.Address) *NameRequest {
	return &NameRequest{
		Contract: contract,
	}
}

func (r *NameRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SymbolRequest is symbol()'s input on Contract.
type SymbolRequest struct {
	Contract *types.Address
}

func NewSymbolRequest(contract *types.Address) *SymbolRequest {
	return &SymbolRequest{
		Contract: contract,
	}
}

func (r *SymbolRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// DecimalsRequest is decimals()'s input on Contract.
type DecimalsRequest struct {
	Contract *types.Address
}

func NewDecimalsRequest(contract *types.Address) *DecimalsRequest {
	return &DecimalsRequest{
		Contract: contract,
	}
}

func (r *DecimalsRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// TransferRequest is transfer(to, value)'s input on Contract, shared by CallTransfer and Transfer.
type TransferRequest struct {
	Contract *types.Address
	To       *types.Address
	Amount   *big.Int
}

func NewTransferRequest(contract, to *types.Address, amount *big.Int) *TransferRequest {
	return &TransferRequest{
		Contract: contract,
		To:       to,
		Amount:   amount,
	}
}

func (r *TransferRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	if r.Amount == nil {
		return fmt.Errorf("%w: amount is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// ApproveRequest is approve(spender, value)'s input on Contract, shared by CallApprove and Approve.
type ApproveRequest struct {
	Contract *types.Address
	Spender  *types.Address
	Amount   *big.Int
}

func NewApproveRequest(contract, spender *types.Address, amount *big.Int) *ApproveRequest {
	return &ApproveRequest{
		Contract: contract,
		Spender:  spender,
		Amount:   amount,
	}
}

func (r *ApproveRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Spender == nil {
		return fmt.Errorf("%w: spender is nil", bind.ErrInvalidRequest)
	}

	if r.Amount == nil {
		return fmt.Errorf("%w: amount is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// TransferFromRequest is transferFrom(from, to, value)'s input on Contract, shared by
// CallTransferFrom and TransferFrom.
type TransferFromRequest struct {
	Contract *types.Address
	From     *types.Address
	To       *types.Address
	Amount   *big.Int
}

func NewTransferFromRequest(contract, from, to *types.Address, amount *big.Int) *TransferFromRequest {
	return &TransferFromRequest{
		Contract: contract,
		From:     from,
		To:       to,
		Amount:   amount,
	}
}

func (r *TransferFromRequest) Validate() error {
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

	if r.Amount == nil {
		return fmt.Errorf("%w: amount is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// MetadataRequest batches name/symbol/decimals/totalSupply on Token via the Multicall3 at Multicall.
type MetadataRequest struct {
	Multicall    *types.Address
	Token        *types.Address
	AllowFailure bool
}

func NewMetadataRequest(multicall, token *types.Address, allowFailure bool) *MetadataRequest {
	return &MetadataRequest{
		Multicall:    multicall,
		Token:        token,
		AllowFailure: allowFailure,
	}
}

func (r *MetadataRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Multicall == nil {
		return fmt.Errorf("%w: multicall is nil", bind.ErrInvalidRequest)
	}

	if r.Token == nil {
		return fmt.Errorf("%w: token is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// MetadataWithBalanceRequest adds Account's balance of Token to MetadataRequest.
type MetadataWithBalanceRequest struct {
	Multicall    *types.Address
	Token        *types.Address
	Account      *types.Address
	AllowFailure bool
}

func NewMetadataWithBalanceRequest(multicall, token, account *types.Address, allowFailure bool) *MetadataWithBalanceRequest {
	return &MetadataWithBalanceRequest{
		Multicall:    multicall,
		Token:        token,
		Account:      account,
		AllowFailure: allowFailure,
	}
}

func (r *MetadataWithBalanceRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Multicall == nil {
		return fmt.Errorf("%w: multicall is nil", bind.ErrInvalidRequest)
	}

	if r.Token == nil {
		return fmt.Errorf("%w: token is nil", bind.ErrInvalidRequest)
	}

	if r.Account == nil {
		return fmt.Errorf("%w: account is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// AllowanceWithBalanceRequest batches Owner's allowance for Spender and both balances of Token.
type AllowanceWithBalanceRequest struct {
	Multicall    *types.Address
	Token        *types.Address
	Owner        *types.Address
	Spender      *types.Address
	AllowFailure bool
}

func NewAllowanceWithBalanceRequest(multicall, token, owner, spender *types.Address, allowFailure bool) *AllowanceWithBalanceRequest {
	return &AllowanceWithBalanceRequest{
		Multicall:    multicall,
		Token:        token,
		Owner:        owner,
		Spender:      spender,
		AllowFailure: allowFailure,
	}
}

func (r *AllowanceWithBalanceRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Multicall == nil {
		return fmt.Errorf("%w: multicall is nil", bind.ErrInvalidRequest)
	}

	if r.Token == nil {
		return fmt.Errorf("%w: token is nil", bind.ErrInvalidRequest)
	}

	if r.Owner == nil {
		return fmt.Errorf("%w: owner is nil", bind.ErrInvalidRequest)
	}

	if r.Spender == nil {
		return fmt.Errorf("%w: spender is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// TokenBalancesRequest batches each of Accounts' balance of Token.
type TokenBalancesRequest struct {
	Multicall    *types.Address
	Token        *types.Address
	Accounts     []*types.Address
	AllowFailure bool
}

func NewTokenBalancesRequest(multicall, token *types.Address, accounts []*types.Address, allowFailure bool) *TokenBalancesRequest {
	return &TokenBalancesRequest{
		Multicall:    multicall,
		Token:        token,
		Accounts:     accounts,
		AllowFailure: allowFailure,
	}
}

func (r *TokenBalancesRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Multicall == nil {
		return fmt.Errorf("%w: multicall is nil", bind.ErrInvalidRequest)
	}

	if r.Token == nil {
		return fmt.Errorf("%w: token is nil", bind.ErrInvalidRequest)
	}

	if len(r.Accounts) == 0 {
		return fmt.Errorf("%w: accounts is empty", bind.ErrInvalidRequest)
	}

	for i, addr := range r.Accounts {
		if addr == nil {
			return fmt.Errorf("%w: accounts[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

// AddressBalancesRequest batches Account's balance of each of Tokens.
type AddressBalancesRequest struct {
	Multicall    *types.Address
	Account      *types.Address
	Tokens       []*types.Address
	AllowFailure bool
}

func NewAddressBalancesRequest(multicall, account *types.Address, tokens []*types.Address, allowFailure bool) *AddressBalancesRequest {
	return &AddressBalancesRequest{
		Multicall:    multicall,
		Account:      account,
		Tokens:       tokens,
		AllowFailure: allowFailure,
	}
}

func (r *AddressBalancesRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Multicall == nil {
		return fmt.Errorf("%w: multicall is nil", bind.ErrInvalidRequest)
	}

	if r.Account == nil {
		return fmt.Errorf("%w: account is nil", bind.ErrInvalidRequest)
	}

	if len(r.Tokens) == 0 {
		return fmt.Errorf("%w: tokens is empty", bind.ErrInvalidRequest)
	}

	for i, addr := range r.Tokens {
		if addr == nil {
			return fmt.Errorf("%w: tokens[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

// BalancePairsRequest batches Accounts[i]'s balance of Tokens[i], paired by index.
type BalancePairsRequest struct {
	Multicall    *types.Address
	Tokens       []*types.Address
	Accounts     []*types.Address
	AllowFailure bool
}

func NewBalancePairsRequest(multicall *types.Address, tokens, accounts []*types.Address, allowFailure bool) *BalancePairsRequest {
	return &BalancePairsRequest{
		Multicall:    multicall,
		Tokens:       tokens,
		Accounts:     accounts,
		AllowFailure: allowFailure,
	}
}

func (r *BalancePairsRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Multicall == nil {
		return fmt.Errorf("%w: multicall is nil", bind.ErrInvalidRequest)
	}

	if len(r.Tokens) != len(r.Accounts) {
		return fmt.Errorf("%w: tokens and accounts differ in length", bind.ErrInvalidRequest)
	}

	if len(r.Tokens) == 0 {
		return fmt.Errorf("%w: tokens is empty", bind.ErrInvalidRequest)
	}

	for i := range r.Tokens {
		if r.Tokens[i] == nil {
			return fmt.Errorf("%w: tokens[%d] is nil", bind.ErrInvalidRequest, i)
		}

		if r.Accounts[i] == nil {
			return fmt.Errorf("%w: accounts[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

// TransferEvent is IERC20's Transfer event: Transfer(address indexed from, address indexed to, uint256 value).
type TransferEvent struct {
	Contract *types.Address
	From     *types.Address
	To       *types.Address
	Value    *big.Int
}

func NewTransferEvent(contract, from, to *types.Address, value *big.Int) *TransferEvent {
	return &TransferEvent{
		Contract: contract,
		From:     from,
		To:       to,
		Value:    value,
	}
}

// ApprovalEvent is IERC20's Approval event: Approval(address indexed owner, address indexed spender, uint256 value).
type ApprovalEvent struct {
	Contract *types.Address
	Owner    *types.Address
	Spender  *types.Address
	Value    *big.Int
}

func NewApprovalEvent(contract, owner, spender *types.Address, value *big.Int) *ApprovalEvent {
	return &ApprovalEvent{
		Contract: contract,
		Owner:    owner,
		Spender:  spender,
		Value:    value,
	}
}

// Metadata is name/symbol/decimals/totalSupply, batched via multicall3 into one call.
type Metadata struct {
	Name        string
	Symbol      string
	Decimals    uint8
	TotalSupply *big.Int
}

func NewMetadata(name string, symbol string, decimals uint8, totalSupply *big.Int) *Metadata {
	return &Metadata{
		Name:        name,
		Symbol:      symbol,
		Decimals:    decimals,
		TotalSupply: totalSupply,
	}
}

// MetadataWithBalance is Metadata plus one address's balance, batched together.
type MetadataWithBalance struct {
	Metadata
	Balance *big.Int
}

func NewMetadataWithBalance(metadata *Metadata, balance *big.Int) *MetadataWithBalance {
	return &MetadataWithBalance{
		Metadata: *metadata,
		Balance:  balance,
	}
}

// AllowanceWithBalance is Allowance plus both addresses' balances, batched together.
type AllowanceWithBalance struct {
	Value          *big.Int
	OwnerBalance   *big.Int
	SpenderBalance *big.Int
}

func NewAllowanceWithBalance(value *big.Int, ownerBalance, spenderBalance *big.Int) *AllowanceWithBalance {
	return &AllowanceWithBalance{
		Value:          value,
		OwnerBalance:   ownerBalance,
		SpenderBalance: spenderBalance,
	}
}

// AddressBalance is one address's balance, one entry in a TokenBalances batch; Amount is nil on a failed call.
type AddressBalance struct {
	Address *types.Address
	Amount  *big.Int
}

func NewAddressBalance(address *types.Address, amount *big.Int) *AddressBalance {
	return &AddressBalance{
		Address: address,
		Amount:  amount,
	}
}

// TokenBalance is one token's balance, one entry in an AddressBalances batch; Amount is nil on a failed call.
type TokenBalance struct {
	Token  *types.Address
	Amount *big.Int
}

func NewTokenBalance(token *types.Address, amount *big.Int) *TokenBalance {
	return &TokenBalance{
		Token:  token,
		Amount: amount,
	}
}

// TokenAddressBalance is one (token, address) pair's balance, one entry in a BalancePairs batch; Amount is nil on a failed call.
type TokenAddressBalance struct {
	Token   *types.Address
	Address *types.Address
	Amount  *big.Int
}

func NewTokenAddressBalance(token, address *types.Address, amount *big.Int) *TokenAddressBalance {
	return &TokenAddressBalance{
		Token:   token,
		Address: address,
		Amount:  amount,
	}
}

// AddressBalances is one address's balance across many tokens, batched via multicall3.
type AddressBalances struct {
	Address *types.Address
	Tokens  []TokenBalance
}

func NewAddressBalances(address *types.Address, tokens []TokenBalance) *AddressBalances {
	return &AddressBalances{
		Address: address,
		Tokens:  tokens,
	}
}

// TokenBalances is one token's balance across many addresses, batched via multicall3.
type TokenBalances struct {
	Token     *types.Address
	Addresses []AddressBalance
}

func NewTokenBalances(token *types.Address, addresses []AddressBalance) *TokenBalances {
	return &TokenBalances{
		Token:     token,
		Addresses: addresses,
	}
}
