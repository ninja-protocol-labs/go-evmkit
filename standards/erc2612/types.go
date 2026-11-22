package erc2612

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/standards/eip712"
)

const (
	primaryType = "Permit"

	ownerName    = "owner"
	spenderName  = "spender"
	valueName    = "value"
	nonceName    = "nonce"
	deadlineName = "deadline"
)

// DefaultVersion is OpenZeppelin's conventional domain version, not a spec-mandated one.
const DefaultVersion = "1"

// permitTypes is the EIP-712 schema for Permit.
var permitTypes = eip712.NewTypes(primaryType,
	eip712.NewField(ownerName, abi.Address.String()),
	eip712.NewField(spenderName, abi.Address.String()),
	eip712.NewField(valueName, abi.Uint256.String()),
	eip712.NewField(nonceName, abi.Uint256.String()),
	eip712.NewField(deadlineName, abi.Uint256.String()),
)

// Permit is ERC-2612's signable struct: owner authorizes spender over value, expiring at deadline.
type Permit struct {
	Owner    *types.Address
	Spender  *types.Address
	Value    *big.Int
	Nonce    *big.Int
	Deadline *big.Int
}

func NewPermit(owner, spender *types.Address, value, nonce, deadline *big.Int) *Permit {
	return &Permit{
		Owner:    owner,
		Spender:  spender,
		Value:    value,
		Nonce:    nonce,
		Deadline: deadline,
	}
}

// Hash returns the signable EIP-712 digest for p under domain.
func (p *Permit) Hash(domain *eip712.Domain) (*types.Hash, error) {
	message := map[string]any{
		ownerName:    p.Owner,
		spenderName:  p.Spender,
		valueName:    p.Value,
		nonceName:    p.Nonce,
		deadlineName: p.Deadline,
	}

	h, err := eip712.Hash(domain, primaryType, permitTypes, message)
	if err != nil {
		return nil, fmt.Errorf("erc2612: hash: %w", err)
	}
	return h, nil
}

// Sign returns signer's signature over p's EIP-712 digest under domain -- the v, r, s permit() needs.
// Use sig.LegacyV() (27/28), not V() or EIP155V(), since ecrecover/ECDSA.recover expect that encoding;
// Permit's own chain-replay protection already comes from domain's ChainID, not from v.
func (p *Permit) Sign(domain *eip712.Domain, signer core.Signer) (*types.Signature, error) {
	digest, err := p.Hash(domain)
	if err != nil {
		return nil, err
	}

	sig, err := signer.Sign(digest)
	if err != nil {
		return nil, fmt.Errorf("erc2612: sign: %w", err)
	}
	return sig, nil
}

// NoncesRequest is nonces(owner)'s input on Contract.
type NoncesRequest struct {
	Contract *types.Address
	Owner    *types.Address
}

func NewNoncesRequest(contract, owner *types.Address) *NoncesRequest {
	return &NoncesRequest{
		Contract: contract,
		Owner:    owner,
	}
}

func (r *NoncesRequest) Validate() error {
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

// DomainSeparatorRequest is DOMAIN_SEPARATOR()'s input on Contract.
type DomainSeparatorRequest struct {
	Contract *types.Address
}

func NewDomainSeparatorRequest(contract *types.Address) *DomainSeparatorRequest {
	return &DomainSeparatorRequest{
		Contract: contract,
	}
}

func (r *DomainSeparatorRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// PermitRequest is permit(owner, spender, value, deadline, v, r, s)'s input on Contract; Signature
// is Owner's signature over the matching Permit, and anyone may relay it.
type PermitRequest struct {
	Contract  *types.Address
	Owner     *types.Address
	Spender   *types.Address
	Amount    *big.Int
	Deadline  *big.Int
	Signature *types.Signature
}

func NewPermitRequest(contract, owner, spender *types.Address, amount, deadline *big.Int, signature *types.Signature) *PermitRequest {
	return &PermitRequest{
		Contract:  contract,
		Owner:     owner,
		Spender:   spender,
		Amount:    amount,
		Deadline:  deadline,
		Signature: signature,
	}
}

func (r *PermitRequest) Validate() error {
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

	if r.Amount == nil {
		return fmt.Errorf("%w: amount is nil", bind.ErrInvalidRequest)
	}

	if r.Deadline == nil {
		return fmt.Errorf("%w: deadline is nil", bind.ErrInvalidRequest)
	}

	if r.Signature == nil {
		return fmt.Errorf("%w: signature is nil", bind.ErrInvalidRequest)
	}

	return nil
}
