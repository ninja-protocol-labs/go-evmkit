package eip7702

import (
	"encoding/json/v2"
	"errors"
	"fmt"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

var (
	ErrInvalidDelegate  = errors.New("invalid delegate")
	ErrInvalidChainID   = errors.New("invalid chain id")
	ErrInvalidNonce     = errors.New("invalid nonce")
	ErrInvalidSignature = errors.New("invalid signature")
	ErrInvalidAuthority = errors.New("invalid authority")
)

// CodeKind classifies an account's code as empty, an EIP-7702 delegation or contract code.
type CodeKind uint8

const (
	EmptyCode CodeKind = iota
	DelegatedCode
	ContractCode
)

func (k CodeKind) String() string {
	switch k {
	case EmptyCode:
		return "empty"
	case DelegatedCode:
		return "delegated"
	case ContractCode:
		return "contract"
	default:
		return fmt.Sprintf("CodeKind(%d)", uint8(k))
	}
}

// AccountCode is an account's code classification; Delegate is set only for DelegatedCode.
type AccountCode struct {
	Kind     CodeKind
	Delegate *types.Address
}

func NewAccountCode(kind CodeKind, delegate *types.Address) *AccountCode {
	return &AccountCode{
		Kind:     kind,
		Delegate: delegate,
	}
}

// UnmarshalJSON classifies an eth_getCode hex result; null and "" are rejected rather than read as empty code.
func (a *AccountCode) UnmarshalJSON(data []byte) error {
	var raw *string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("eip7702: account code: %w", err)
	}
	if raw == nil || *raw == "" {
		return errors.New("eip7702: account code: empty eth_getCode result")
	}

	code, err := encoding.Hex.Decode(*raw)
	if err != nil {
		return fmt.Errorf("eip7702: account code: %w", err)
	}

	switch delegate, ok := ParseDelegation(code); {
	case len(code) == 0:
		*a = AccountCode{Kind: EmptyCode}
	case ok:
		*a = AccountCode{Kind: DelegatedCode, Delegate: delegate}
	default:
		*a = AccountCode{Kind: ContractCode}
	}
	return nil
}

// GetDelegationRequest is eth_getCode's input for Account.
type GetDelegationRequest struct {
	Account *types.Address
}

func NewGetDelegationRequest(account *types.Address) *GetDelegationRequest {
	return &GetDelegationRequest{
		Account: account,
	}
}

func (r *GetDelegationRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Account == nil {
		return fmt.Errorf("%w: account is nil", bind.ErrInvalidRequest)
	}

	return nil
}
