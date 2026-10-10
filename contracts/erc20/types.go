package erc20

import (
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Transfer is IERC20's Transfer event: Transfer(address indexed from, address indexed to, uint256 value).
type Transfer struct {
	Contract *types.Address
	From     *types.Address
	To       *types.Address
	Value    *big.Int
}

func NewTransfer(contract, from, to *types.Address, value *big.Int) *Transfer {
	return &Transfer{
		Contract: contract,
		From:     from,
		To:       to,
		Value:    value,
	}
}

// Approval is IERC20's Approval event: Approval(address indexed owner, address indexed spender, uint256 value).
type Approval struct {
	Contract *types.Address
	Owner    *types.Address
	Spender  *types.Address
	Value    *big.Int
}

func NewApproval(contract, owner, spender *types.Address, value *big.Int) *Approval {
	return &Approval{
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
