package erc721

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// BalanceOfRequest balanceOf(owner)'s input on Contract.
type BalanceOfRequest struct {
	Contract *types.Address
	Owner    *types.Address
}

func NewBalanceOfRequest(contract, owner *types.Address) *BalanceOfRequest {
	return &BalanceOfRequest{
		Contract: contract,
		Owner:    owner,
	}
}

func (r *BalanceOfRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Owner == nil {
		return fmt.Errorf("%w: owner is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// OwnerOfRequest ownerOf(tokenId)'s input on Contract.
type OwnerOfRequest struct {
	Contract *types.Address
	TokenID  *big.Int
}

func NewOwnerOfRequest(contract *types.Address, tokenID *big.Int) *OwnerOfRequest {
	return &OwnerOfRequest{
		Contract: contract,
		TokenID:  tokenID,
	}
}

func (r *OwnerOfRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.TokenID == nil {
		return fmt.Errorf("%w: tokenID is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetApprovedRequest getApproved(tokenId)'s input on Contract.
type GetApprovedRequest struct {
	Contract *types.Address
	TokenID  *big.Int
}

func NewGetApprovedRequest(contract *types.Address, tokenID *big.Int) *GetApprovedRequest {
	return &GetApprovedRequest{
		Contract: contract,
		TokenID:  tokenID,
	}
}

func (r *GetApprovedRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.TokenID == nil {
		return fmt.Errorf("%w: tokenID is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// IsApprovedForAllRequest isApprovedForAll(owner, operator)'s input on Contract.
type IsApprovedForAllRequest struct {
	Contract *types.Address
	Owner    *types.Address
	Operator *types.Address
}

func NewIsApprovedForAllRequest(contract, owner, operator *types.Address) *IsApprovedForAllRequest {
	return &IsApprovedForAllRequest{
		Contract: contract,
		Owner:    owner,
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

	if r.Owner == nil {
		return fmt.Errorf("%w: owner is nil", bind.ErrInvalidRequest)
	}

	if r.Operator == nil {
		return fmt.Errorf("%w: operator is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// NameRequest name()'s input on Contract.
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

// SymbolRequest symbol()'s input on Contract.
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

// TokenURIRequest tokenURI(tokenId)'s input on Contract.
type TokenURIRequest struct {
	Contract *types.Address
	TokenID  *big.Int
}

func NewTokenURIRequest(contract *types.Address, tokenID *big.Int) *TokenURIRequest {
	return &TokenURIRequest{
		Contract: contract,
		TokenID:  tokenID,
	}
}

func (r *TokenURIRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.TokenID == nil {
		return fmt.Errorf("%w: tokenID is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// ApproveRequest approve(to, tokenId)'s input on Contract.
type ApproveRequest struct {
	Contract *types.Address
	To       *types.Address
	TokenID  *big.Int
}

func NewApproveRequest(contract, to *types.Address, tokenID *big.Int) *ApproveRequest {
	return &ApproveRequest{
		Contract: contract,
		To:       to,
		TokenID:  tokenID,
	}
}

func (r *ApproveRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.To == nil {
		return fmt.Errorf("%w: to is nil", bind.ErrInvalidRequest)
	}

	if r.TokenID == nil {
		return fmt.Errorf("%w: tokenID is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SetApprovalForAllRequest setApprovalForAll(operator, approved)'s input on Contract.
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

// TransferFromRequest transferFrom(from, to, tokenId)'s input on Contract.
type TransferFromRequest struct {
	Contract *types.Address
	From     *types.Address
	To       *types.Address
	TokenID  *big.Int
}

func NewTransferFromRequest(contract, from, to *types.Address, tokenID *big.Int) *TransferFromRequest {
	return &TransferFromRequest{
		Contract: contract,
		From:     from,
		To:       to,
		TokenID:  tokenID,
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

	if r.TokenID == nil {
		return fmt.Errorf("%w: tokenID is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SafeTransferFromRequest safeTransferFrom(from, to, tokenId)'s input on Contract.
type SafeTransferFromRequest struct {
	Contract *types.Address
	From     *types.Address
	To       *types.Address
	TokenID  *big.Int
}

func NewSafeTransferFromRequest(contract, from, to *types.Address, tokenID *big.Int) *SafeTransferFromRequest {
	return &SafeTransferFromRequest{
		Contract: contract,
		From:     from,
		To:       to,
		TokenID:  tokenID,
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

	if r.TokenID == nil {
		return fmt.Errorf("%w: tokenID is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SafeTransferFromWithDataRequest safeTransferFrom(from, to, tokenId, data)'s input on Contract; Data may be empty.
type SafeTransferFromWithDataRequest struct {
	Contract *types.Address
	From     *types.Address
	To       *types.Address
	TokenID  *big.Int
	Data     []byte
}

func NewSafeTransferFromWithDataRequest(contract, from, to *types.Address, tokenID *big.Int, data []byte) *SafeTransferFromWithDataRequest {
	return &SafeTransferFromWithDataRequest{
		Contract: contract,
		From:     from,
		To:       to,
		TokenID:  tokenID,
		Data:     data,
	}
}

func (r *SafeTransferFromWithDataRequest) Validate() error {
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

	if r.TokenID == nil {
		return fmt.Errorf("%w: tokenID is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// MetadataRequest batches name/symbol on Contract via the Multicall3 at Multicall.
type MetadataRequest struct {
	Multicall    *types.Address
	Contract     *types.Address
	AllowFailure bool
}

func NewMetadataRequest(multicall, contract *types.Address, allowFailure bool) *MetadataRequest {
	return &MetadataRequest{
		Multicall:    multicall,
		Contract:     contract,
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

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// MetadataWithTokenIDRequest adds TokenID's URI and owner to MetadataRequest.
type MetadataWithTokenIDRequest struct {
	Multicall    *types.Address
	Contract     *types.Address
	TokenID      *big.Int
	AllowFailure bool
}

func NewMetadataWithTokenIDRequest(multicall, contract *types.Address, tokenID *big.Int, allowFailure bool) *MetadataWithTokenIDRequest {
	return &MetadataWithTokenIDRequest{
		Multicall:    multicall,
		Contract:     contract,
		TokenID:      tokenID,
		AllowFailure: allowFailure,
	}
}

func (r *MetadataWithTokenIDRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Multicall == nil {
		return fmt.Errorf("%w: multicall is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.TokenID == nil {
		return fmt.Errorf("%w: tokenID is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// AllowanceWithTokenIDRequest batches TokenID's owner and approved operator on Contract.
type AllowanceWithTokenIDRequest struct {
	Multicall    *types.Address
	Contract     *types.Address
	TokenID      *big.Int
	AllowFailure bool
}

func NewAllowanceWithTokenIDRequest(multicall, contract *types.Address, tokenID *big.Int, allowFailure bool) *AllowanceWithTokenIDRequest {
	return &AllowanceWithTokenIDRequest{
		Multicall:    multicall,
		Contract:     contract,
		TokenID:      tokenID,
		AllowFailure: allowFailure,
	}
}

func (r *AllowanceWithTokenIDRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Multicall == nil {
		return fmt.Errorf("%w: multicall is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.TokenID == nil {
		return fmt.Errorf("%w: tokenID is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// ApprovalForAllWithBalanceRequest batches isApprovedForAll(Owner, Operator) and both balances on Contract.
type ApprovalForAllWithBalanceRequest struct {
	Multicall    *types.Address
	Contract     *types.Address
	Owner        *types.Address
	Operator     *types.Address
	AllowFailure bool
}

func NewApprovalForAllWithBalanceRequest(multicall, contract, owner, operator *types.Address, allowFailure bool) *ApprovalForAllWithBalanceRequest {
	return &ApprovalForAllWithBalanceRequest{
		Multicall:    multicall,
		Contract:     contract,
		Owner:        owner,
		Operator:     operator,
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

	if r.Owner == nil {
		return fmt.Errorf("%w: owner is nil", bind.ErrInvalidRequest)
	}

	if r.Operator == nil {
		return fmt.Errorf("%w: operator is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// TokenOwnersRequest batches ownerOf on Contract for each of TokenIDs.
type TokenOwnersRequest struct {
	Multicall    *types.Address
	Contract     *types.Address
	TokenIDs     []*big.Int
	AllowFailure bool
}

func NewTokenOwnersRequest(multicall, contract *types.Address, tokenIDs []*big.Int, allowFailure bool) *TokenOwnersRequest {
	return &TokenOwnersRequest{
		Multicall:    multicall,
		Contract:     contract,
		TokenIDs:     tokenIDs,
		AllowFailure: allowFailure,
	}
}

func (r *TokenOwnersRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Multicall == nil {
		return fmt.Errorf("%w: multicall is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if len(r.TokenIDs) == 0 {
		return fmt.Errorf("%w: tokenIDs is empty", bind.ErrInvalidRequest)
	}

	for i, v := range r.TokenIDs {
		if v == nil {
			return fmt.Errorf("%w: tokenIDs[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

// TokenOperatorsRequest batches getApproved on Contract for each of TokenIDs.
type TokenOperatorsRequest struct {
	Multicall    *types.Address
	Contract     *types.Address
	TokenIDs     []*big.Int
	AllowFailure bool
}

func NewTokenOperatorsRequest(multicall, contract *types.Address, tokenIDs []*big.Int, allowFailure bool) *TokenOperatorsRequest {
	return &TokenOperatorsRequest{
		Multicall:    multicall,
		Contract:     contract,
		TokenIDs:     tokenIDs,
		AllowFailure: allowFailure,
	}
}

func (r *TokenOperatorsRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Multicall == nil {
		return fmt.Errorf("%w: multicall is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if len(r.TokenIDs) == 0 {
		return fmt.Errorf("%w: tokenIDs is empty", bind.ErrInvalidRequest)
	}

	for i, v := range r.TokenIDs {
		if v == nil {
			return fmt.Errorf("%w: tokenIDs[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

// AddressBalancesRequest batches Account's balance on each of Contracts.
type AddressBalancesRequest struct {
	Multicall    *types.Address
	Account      *types.Address
	Contracts    []*types.Address
	AllowFailure bool
}

func NewAddressBalancesRequest(multicall, account *types.Address, contracts []*types.Address, allowFailure bool) *AddressBalancesRequest {
	return &AddressBalancesRequest{
		Multicall:    multicall,
		Account:      account,
		Contracts:    contracts,
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

// TokenBalancesRequest batches each of Accounts' balance on Contract.
type TokenBalancesRequest struct {
	Multicall    *types.Address
	Contract     *types.Address
	Accounts     []*types.Address
	AllowFailure bool
}

func NewTokenBalancesRequest(multicall, contract *types.Address, accounts []*types.Address, allowFailure bool) *TokenBalancesRequest {
	return &TokenBalancesRequest{
		Multicall:    multicall,
		Contract:     contract,
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

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if len(r.Accounts) == 0 {
		return fmt.Errorf("%w: accounts is empty", bind.ErrInvalidRequest)
	}

	for i, v := range r.Accounts {
		if v == nil {
			return fmt.Errorf("%w: accounts[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

// OwnerPairsRequest batches ownerOf(TokenIDs[i]) on Contracts[i], paired by index.
type OwnerPairsRequest struct {
	Multicall    *types.Address
	Contracts    []*types.Address
	TokenIDs     []*big.Int
	AllowFailure bool
}

func NewOwnerPairsRequest(multicall *types.Address, contracts []*types.Address, tokenIDs []*big.Int, allowFailure bool) *OwnerPairsRequest {
	return &OwnerPairsRequest{
		Multicall:    multicall,
		Contracts:    contracts,
		TokenIDs:     tokenIDs,
		AllowFailure: allowFailure,
	}
}

func (r *OwnerPairsRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Multicall == nil {
		return fmt.Errorf("%w: multicall is nil", bind.ErrInvalidRequest)
	}

	if len(r.Contracts) != len(r.TokenIDs) {
		return fmt.Errorf("%w: contracts and tokenIDs differ in length", bind.ErrInvalidRequest)
	}

	if len(r.Contracts) == 0 {
		return fmt.Errorf("%w: contracts is empty", bind.ErrInvalidRequest)
	}

	for i := range r.Contracts {
		if r.Contracts[i] == nil {
			return fmt.Errorf("%w: contracts[%d] is nil", bind.ErrInvalidRequest, i)
		}

		if r.TokenIDs[i] == nil {
			return fmt.Errorf("%w: tokenIDs[%d] is nil", bind.ErrInvalidRequest, i)
		}
	}

	return nil
}

// TransferEvent is IERC721's Transfer event: Transfer(address indexed from, address indexed to, uint256 indexed tokenId).
type TransferEvent struct {
	Contract *types.Address
	From     *types.Address
	To       *types.Address
	TokenId  *big.Int
}

func NewTransferEvent(contract, from, to *types.Address, tokenId *big.Int) *TransferEvent {
	return &TransferEvent{
		Contract: contract,
		From:     from,
		To:       to,
		TokenId:  tokenId,
	}
}

// ApprovalEvent is IERC721's Approval event: Approval(address indexed owner, address indexed approved, uint256 indexed tokenId).
type ApprovalEvent struct {
	Contract *types.Address
	Owner    *types.Address
	Approved *types.Address
	TokenId  *big.Int
}

func NewApprovalEvent(contract, owner, approved *types.Address, tokenId *big.Int) *ApprovalEvent {
	return &ApprovalEvent{
		Contract: contract,
		Owner:    owner,
		Approved: approved,
		TokenId:  tokenId,
	}
}

// ApprovalForAllEvent is IERC721's ApprovalForAll event: ApprovalForAll(address indexed owner, address indexed operator, bool approved).
type ApprovalForAllEvent struct {
	Contract *types.Address
	Owner    *types.Address
	Operator *types.Address
	Approved bool
}

func NewApprovalForAllEvent(contract, owner, operator *types.Address, approved bool) *ApprovalForAllEvent {
	return &ApprovalForAllEvent{
		Contract: contract,
		Owner:    owner,
		Operator: operator,
		Approved: approved,
	}
}

// Metadata is name/symbol, batched via multicall3 into one call.
type Metadata struct {
	Name   string
	Symbol string
}

func NewMetadata(name, symbol string) *Metadata {
	return &Metadata{
		Name:   name,
		Symbol: symbol,
	}
}

// MetadataWithTokenID is Metadata plus tokenId's URI and owner, batched via multicall3 into one call.
type MetadataWithTokenID struct {
	Metadata
	TokenId  *big.Int
	TokenURI string
	Owner    *types.Address
}

func NewMetadataWithTokenID(metadata *Metadata, tokenId *big.Int, tokenURI string, owner *types.Address) *MetadataWithTokenID {
	return &MetadataWithTokenID{
		Metadata: *metadata,
		TokenId:  tokenId,
		TokenURI: tokenURI,
		Owner:    owner,
	}
}

// AllowanceWithTokenID is a tokenId's owner and approved operator, batched via multicall3 into one call.
type AllowanceWithTokenID struct {
	TokenId  *big.Int
	Owner    *types.Address
	Operator *types.Address
}

func NewAllowanceWithTokenID(tokenId *big.Int, owner, operator *types.Address) *AllowanceWithTokenID {
	return &AllowanceWithTokenID{
		TokenId:  tokenId,
		Owner:    owner,
		Operator: operator,
	}
}

// ApprovalForAllWithBalance is isApprovedForAll(owner, operator) plus both their balances, batched via multicall3 into one call.
type ApprovalForAllWithBalance struct {
	Owner           *types.Address
	Operator        *types.Address
	Approved        bool
	OwnerBalance    *big.Int
	OperatorBalance *big.Int
}

func NewApprovalForAllWithBalance(owner, operator *types.Address, approved bool, ownerBalance, operatorBalance *big.Int) *ApprovalForAllWithBalance {
	return &ApprovalForAllWithBalance{
		Owner:           owner,
		Operator:        operator,
		Approved:        approved,
		OwnerBalance:    ownerBalance,
		OperatorBalance: operatorBalance,
	}
}

// TokenOwner is one tokenId's owner, one entry in a TokenOwners batch; Owner is nil on a failed call (e.g. a nonexistent token).
type TokenOwner struct {
	TokenId *big.Int
	Owner   *types.Address
}

func NewTokenOwner(tokenId *big.Int, owner *types.Address) *TokenOwner {
	return &TokenOwner{
		TokenId: tokenId,
		Owner:   owner,
	}
}

// TokenOwners is one contract's owners across many tokenIds, batched via multicall3.
type TokenOwners struct {
	Contract *types.Address
	Tokens   []TokenOwner
}

func NewTokenOwners(contract *types.Address, tokens []TokenOwner) *TokenOwners {
	return &TokenOwners{
		Contract: contract,
		Tokens:   tokens,
	}
}

// TokenOperator is one tokenId's approved operator, one entry in a TokenOperators batch; Operator is nil on a failed call.
type TokenOperator struct {
	TokenId  *big.Int
	Operator *types.Address
}

func NewTokenOperator(tokenId *big.Int, operator *types.Address) *TokenOperator {
	return &TokenOperator{
		TokenId:  tokenId,
		Operator: operator,
	}
}

// TokenOperators is one contract's approved operators across many tokenIds, batched via multicall3.
type TokenOperators struct {
	Contract *types.Address
	Tokens   []TokenOperator
}

func NewTokenOperators(contract *types.Address, tokens []TokenOperator) *TokenOperators {
	return &TokenOperators{
		Contract: contract,
		Tokens:   tokens,
	}
}

// TokenBalance is one contract's balanceOf result for an address, one entry in an AddressBalances batch; Amount is nil on a failed call.
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

// AddressBalances is one address's balanceOf across many standards, batched via multicall3.
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

// TokenOwnerPair is one (contract, tokenId) pair's owner, one entry in a TokenOwnerPairs batch; Owner is nil on a failed call.
type TokenOwnerPair struct {
	Contract *types.Address
	TokenId  *big.Int
	Owner    *types.Address
}

func NewTokenOwnerPair(contract, owner *types.Address, tokenId *big.Int) *TokenOwnerPair {
	return &TokenOwnerPair{
		Contract: contract,
		TokenId:  tokenId,
		Owner:    owner,
	}
}

// AddressBalance is one address's balanceOf result for the contract, one entry in a TokenBalances batch; Amount is nil on a failed call.
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

// TokenBalances is one contract's balanceOf across many addresses, batched via multicall3.
type TokenBalances struct {
	Contract  *types.Address
	Addresses []AddressBalance
}

func NewTokenBalances(contract *types.Address, addresses []AddressBalance) *TokenBalances {
	return &TokenBalances{
		Contract:  contract,
		Addresses: addresses,
	}
}
