package erc6909

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// BalanceOfRequest is balanceOf(owner, id)'s input on Contract.
type BalanceOfRequest struct {
	Contract *types.Address
	Owner    *types.Address
	ID       *big.Int
}

func NewBalanceOfRequest(contract, owner *types.Address, id *big.Int) *BalanceOfRequest {
	return &BalanceOfRequest{
		Contract: contract,
		Owner:    owner,
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

	if r.Owner == nil {
		return fmt.Errorf("%w: owner is nil", bind.ErrInvalidRequest)
	}

	if r.ID == nil {
		return fmt.Errorf("%w: id is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// AllowanceRequest is allowance(owner, spender, id)'s input on Contract.
type AllowanceRequest struct {
	Contract *types.Address
	Owner    *types.Address
	Spender  *types.Address
	ID       *big.Int
}

func NewAllowanceRequest(contract, owner, spender *types.Address, id *big.Int) *AllowanceRequest {
	return &AllowanceRequest{
		Contract: contract,
		Owner:    owner,
		Spender:  spender,
		ID:       id,
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

	if r.ID == nil {
		return fmt.Errorf("%w: id is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// IsOperatorRequest is isOperator(owner, spender)'s input on Contract.
type IsOperatorRequest struct {
	Contract *types.Address
	Owner    *types.Address
	Spender  *types.Address
}

func NewIsOperatorRequest(contract, owner, spender *types.Address) *IsOperatorRequest {
	return &IsOperatorRequest{
		Contract: contract,
		Owner:    owner,
		Spender:  spender,
	}
}

func (r *IsOperatorRequest) Validate() error {
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

// NameRequest is name(id)'s input on Contract.
type NameRequest struct {
	Contract *types.Address
	ID       *big.Int
}

func NewNameRequest(contract *types.Address, id *big.Int) *NameRequest {
	return &NameRequest{
		Contract: contract,
		ID:       id,
	}
}

func (r *NameRequest) Validate() error {
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

// SymbolRequest is symbol(id)'s input on Contract.
type SymbolRequest struct {
	Contract *types.Address
	ID       *big.Int
}

func NewSymbolRequest(contract *types.Address, id *big.Int) *SymbolRequest {
	return &SymbolRequest{
		Contract: contract,
		ID:       id,
	}
}

func (r *SymbolRequest) Validate() error {
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

// DecimalsRequest is decimals(id)'s input on Contract.
type DecimalsRequest struct {
	Contract *types.Address
	ID       *big.Int
}

func NewDecimalsRequest(contract *types.Address, id *big.Int) *DecimalsRequest {
	return &DecimalsRequest{
		Contract: contract,
		ID:       id,
	}
}

func (r *DecimalsRequest) Validate() error {
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

// ApproveRequest is approve(spender, id, amount)'s input on Contract, shared by CallApprove and Approve.
type ApproveRequest struct {
	Contract *types.Address
	Spender  *types.Address
	ID       *big.Int
	Amount   *big.Int
}

func NewApproveRequest(contract, spender *types.Address, id, amount *big.Int) *ApproveRequest {
	return &ApproveRequest{
		Contract: contract,
		Spender:  spender,
		ID:       id,
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

	if r.ID == nil {
		return fmt.Errorf("%w: id is nil", bind.ErrInvalidRequest)
	}

	if r.Amount == nil {
		return fmt.Errorf("%w: amount is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SetOperatorRequest is setOperator(spender, approved)'s input on Contract, shared by CallSetOperator and SetOperator.
type SetOperatorRequest struct {
	Contract *types.Address
	Spender  *types.Address
	Approved bool
}

func NewSetOperatorRequest(contract, spender *types.Address, approved bool) *SetOperatorRequest {
	return &SetOperatorRequest{
		Contract: contract,
		Spender:  spender,
		Approved: approved,
	}
}

func (r *SetOperatorRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Spender == nil {
		return fmt.Errorf("%w: spender is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// TransferRequest is transfer(receiver, id, amount)'s input on Contract, shared by CallTransfer and Transfer.
type TransferRequest struct {
	Contract *types.Address
	Receiver *types.Address
	ID       *big.Int
	Amount   *big.Int
}

func NewTransferRequest(contract, receiver *types.Address, id, amount *big.Int) *TransferRequest {
	return &TransferRequest{
		Contract: contract,
		Receiver: receiver,
		ID:       id,
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

	if r.Receiver == nil {
		return fmt.Errorf("%w: receiver is nil", bind.ErrInvalidRequest)
	}

	if r.ID == nil {
		return fmt.Errorf("%w: id is nil", bind.ErrInvalidRequest)
	}

	if r.Amount == nil {
		return fmt.Errorf("%w: amount is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// TransferFromRequest is transferFrom(sender, receiver, id, amount)'s input on Contract, shared by CallTransferFrom and TransferFrom.
type TransferFromRequest struct {
	Contract *types.Address
	Sender   *types.Address
	Receiver *types.Address
	ID       *big.Int
	Amount   *big.Int
}

func NewTransferFromRequest(contract, sender, receiver *types.Address, id, amount *big.Int) *TransferFromRequest {
	return &TransferFromRequest{
		Contract: contract,
		Sender:   sender,
		Receiver: receiver,
		ID:       id,
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

	if r.Sender == nil {
		return fmt.Errorf("%w: sender is nil", bind.ErrInvalidRequest)
	}

	if r.Receiver == nil {
		return fmt.Errorf("%w: receiver is nil", bind.ErrInvalidRequest)
	}

	if r.ID == nil {
		return fmt.Errorf("%w: id is nil", bind.ErrInvalidRequest)
	}

	if r.Amount == nil {
		return fmt.Errorf("%w: amount is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// OperatorSetEvent is IERC6909's OperatorSet event: OperatorSet(address indexed owner, address indexed operator, bool approved).
type OperatorSetEvent struct {
	Contract *types.Address
	Owner    *types.Address
	Operator *types.Address
	Approved bool
}

func NewOperatorSetEvent(contract, owner, operator *types.Address, approved bool) *OperatorSetEvent {
	return &OperatorSetEvent{
		Contract: contract,
		Owner:    owner,
		Operator: operator,
		Approved: approved,
	}
}

// ApprovalEvent is IERC6909's Approval event: Approval(address indexed owner, address indexed spender, uint256 indexed id, uint256 amount).
type ApprovalEvent struct {
	Contract *types.Address
	Owner    *types.Address
	Spender  *types.Address
	Id       *big.Int
	Amount   *big.Int
}

func NewApprovalEvent(contract, owner, spender *types.Address, id, amount *big.Int) *ApprovalEvent {
	return &ApprovalEvent{
		Contract: contract,
		Owner:    owner,
		Spender:  spender,
		Id:       id,
		Amount:   amount,
	}
}

// TransferEvent is IERC6909's Transfer event: Transfer(address caller, address indexed sender, address indexed receiver, uint256 indexed id, uint256 amount).
// Caller (who initiated the call -- the owner itself for transfer, or an approved
// spender/operator for transferFrom) is not indexed, unlike Sender/Receiver/Id.
type TransferEvent struct {
	Contract *types.Address
	Caller   *types.Address
	Sender   *types.Address
	Receiver *types.Address
	Id       *big.Int
	Amount   *big.Int
}

func NewTransferEvent(contract, caller, sender, receiver *types.Address, id, amount *big.Int) *TransferEvent {
	return &TransferEvent{
		Contract: contract,
		Caller:   caller,
		Sender:   sender,
		Receiver: receiver,
		Id:       id,
		Amount:   amount,
	}
}

// NameUpdatedEvent mirrors ERC6909Metadata's ERC6909NameUpdated event: ERC6909NameUpdated(uint256 indexed id, string newName).
// Not part of the IERC6909Metadata interface itself (it's an implementation detail of
// OpenZeppelin's ERC6909Metadata), but emitted by any contract built on it.
type NameUpdatedEvent struct {
	Contract *types.Address
	Id       *big.Int
	NewName  string
}

func NewNameUpdatedEvent(contract *types.Address, id *big.Int, newName string) *NameUpdatedEvent {
	return &NameUpdatedEvent{
		Contract: contract,
		Id:       id,
		NewName:  newName,
	}
}

// SymbolUpdatedEvent mirrors ERC6909Metadata's ERC6909SymbolUpdated event.
type SymbolUpdatedEvent struct {
	Contract  *types.Address
	Id        *big.Int
	NewSymbol string
}

func NewSymbolUpdatedEvent(contract *types.Address, id *big.Int, newSymbol string) *SymbolUpdatedEvent {
	return &SymbolUpdatedEvent{
		Contract:  contract,
		Id:        id,
		NewSymbol: newSymbol,
	}
}

// DecimalsUpdatedEvent mirrors ERC6909Metadata's ERC6909DecimalsUpdated event.
type DecimalsUpdatedEvent struct {
	Contract    *types.Address
	Id          *big.Int
	NewDecimals uint8
}

func NewDecimalsUpdatedEvent(contract *types.Address, id *big.Int, newDecimals uint8) *DecimalsUpdatedEvent {
	return &DecimalsUpdatedEvent{
		Contract:    contract,
		Id:          id,
		NewDecimals: newDecimals,
	}
}
