package erc2612

import (
	"fmt"
	"math/big"

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
