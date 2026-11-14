package erc6909

import (
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// OperatorSet is IERC6909's OperatorSet event: OperatorSet(address indexed owner, address indexed operator, bool approved).
type OperatorSet struct {
	Contract *types.Address
	Owner    *types.Address
	Operator *types.Address
	Approved bool
}

func NewOperatorSet(contract, owner, operator *types.Address, approved bool) *OperatorSet {
	return &OperatorSet{
		Contract: contract,
		Owner:    owner,
		Operator: operator,
		Approved: approved,
	}
}

// Approval is IERC6909's Approval event: Approval(address indexed owner, address indexed spender, uint256 indexed id, uint256 amount).
type Approval struct {
	Contract *types.Address
	Owner    *types.Address
	Spender  *types.Address
	Id       *big.Int
	Amount   *big.Int
}

func NewApproval(contract, owner, spender *types.Address, id, amount *big.Int) *Approval {
	return &Approval{
		Contract: contract,
		Owner:    owner,
		Spender:  spender,
		Id:       id,
		Amount:   amount,
	}
}

// Transfer is IERC6909's Transfer event: Transfer(address caller, address indexed sender, address indexed receiver, uint256 indexed id, uint256 amount).
// Caller (who initiated the call -- the owner itself for transfer, or an approved
// spender/operator for transferFrom) is not indexed, unlike Sender/Receiver/Id.
type Transfer struct {
	Contract *types.Address
	Caller   *types.Address
	Sender   *types.Address
	Receiver *types.Address
	Id       *big.Int
	Amount   *big.Int
}

func NewTransfer(contract, caller, sender, receiver *types.Address, id, amount *big.Int) *Transfer {
	return &Transfer{
		Contract: contract,
		Caller:   caller,
		Sender:   sender,
		Receiver: receiver,
		Id:       id,
		Amount:   amount,
	}
}

// NameUpdated mirrors ERC6909Metadata's ERC6909NameUpdated event: ERC6909NameUpdated(uint256 indexed id, string newName).
// Not part of the IERC6909Metadata interface itself (it's an implementation detail of
// OpenZeppelin's ERC6909Metadata), but emitted by any contract built on it.
type NameUpdated struct {
	Contract *types.Address
	Id       *big.Int
	NewName  string
}

func NewNameUpdated(contract *types.Address, id *big.Int, newName string) *NameUpdated {
	return &NameUpdated{
		Contract: contract,
		Id:       id,
		NewName:  newName,
	}
}

// SymbolUpdated mirrors ERC6909Metadata's ERC6909SymbolUpdated event.
type SymbolUpdated struct {
	Contract  *types.Address
	Id        *big.Int
	NewSymbol string
}

func NewSymbolUpdated(contract *types.Address, id *big.Int, newSymbol string) *SymbolUpdated {
	return &SymbolUpdated{
		Contract:  contract,
		Id:        id,
		NewSymbol: newSymbol,
	}
}

// DecimalsUpdated mirrors ERC6909Metadata's ERC6909DecimalsUpdated event.
type DecimalsUpdated struct {
	Contract    *types.Address
	Id          *big.Int
	NewDecimals uint8
}

func NewDecimalsUpdated(contract *types.Address, id *big.Int, newDecimals uint8) *DecimalsUpdated {
	return &DecimalsUpdated{
		Contract:    contract,
		Id:          id,
		NewDecimals: newDecimals,
	}
}
