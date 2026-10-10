package erc721

import (
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Transfer is IERC721's Transfer event: Transfer(address indexed from, address indexed to, uint256 indexed tokenId).
type Transfer struct {
	Contract *types.Address
	From     *types.Address
	To       *types.Address
	TokenId  *big.Int
}

func NewTransfer(contract, from, to *types.Address, tokenId *big.Int) *Transfer {
	return &Transfer{
		Contract: contract,
		From:     from,
		To:       to,
		TokenId:  tokenId,
	}
}

// Approval is IERC721's Approval event: Approval(address indexed owner, address indexed approved, uint256 indexed tokenId).
type Approval struct {
	Contract *types.Address
	Owner    *types.Address
	Approved *types.Address
	TokenId  *big.Int
}

func NewApproval(contract, owner, approved *types.Address, tokenId *big.Int) *Approval {
	return &Approval{
		Contract: contract,
		Owner:    owner,
		Approved: approved,
		TokenId:  tokenId,
	}
}

// ApprovalForAll is IERC721's ApprovalForAll event: ApprovalForAll(address indexed owner, address indexed operator, bool approved).
type ApprovalForAll struct {
	Contract *types.Address
	Owner    *types.Address
	Operator *types.Address
	Approved bool
}

func NewApprovalForAll(contract, owner, operator *types.Address, approved bool) *ApprovalForAll {
	return &ApprovalForAll{
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

// AddressBalances is one address's balanceOf across many contracts, batched via multicall3.
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
