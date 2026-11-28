package eip7702

import (
	"fmt"
	"math"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// NewAuthorization returns an unsigned authorization delegating to delegate on chainID only.
func NewAuthorization(chainID *big.Int, delegate *types.Address, nonce uint64) (*core.Authorization, error) {
	if delegate.IsZero() {
		return nil, fmt.Errorf("eip7702: newAuthorization: %w", ErrInvalidDelegate)
	}

	if chainID == nil || chainID.Sign() <= 0 {
		return nil, fmt.Errorf("eip7702: newAuthorization: %w", ErrInvalidChainID)
	}
	if nonce == math.MaxUint64 {
		return nil, fmt.Errorf("eip7702: newAuthorization: %w", ErrInvalidNonce)
	}

	return core.NewAuthorization(chainID, delegate, nonce), nil
}

// NewRevocation returns an unsigned authorization that clears the delegation on chainID.
func NewRevocation(chainID *big.Int, nonce uint64) (*core.Authorization, error) {
	if chainID == nil || chainID.Sign() <= 0 {
		return nil, fmt.Errorf("eip7702: newRevocation: %w", ErrInvalidChainID)
	}
	if nonce == math.MaxUint64 {
		return nil, fmt.Errorf("eip7702: newRevocation: %w", ErrInvalidNonce)
	}

	return core.NewAuthorization(chainID, types.ZeroAddress(), nonce), nil
}

// SelfAuthorizationNonce returns txNonce+1, the authorization nonce when the authority also sends the tx.
func SelfAuthorizationNonce(txNonce uint64) (uint64, error) {
	if txNonce >= math.MaxUint64-1 {
		return 0, fmt.Errorf("eip7702: selfAuthorizationNonce: %w", ErrInvalidNonce)
	}
	return txNonce + 1, nil
}

// CheckAuthorization verifies offline that authority signed auth for exactly chainID and delegate; high-s signatures fail recovery.
func CheckAuthorization(auth *core.Authorization, authority *types.Address, chainID *big.Int, delegate *types.Address) error {
	if auth == nil {
		return fmt.Errorf("eip7702: checkAuthorization: %w: authorization is nil", bind.ErrInvalidRequest)
	}
	if authority == nil {
		return fmt.Errorf("eip7702: checkAuthorization: %w: authority is nil", bind.ErrInvalidRequest)
	}
	if delegate == nil {
		return fmt.Errorf("eip7702: checkAuthorization: %w: delegate is nil", bind.ErrInvalidRequest)
	}
	if chainID == nil || chainID.Sign() <= 0 {
		return fmt.Errorf("eip7702: checkAuthorization: %w: expected chain id must be positive", ErrInvalidChainID)
	}

	if auth.ChainID == nil || auth.ChainID.Cmp(chainID) != 0 {
		return fmt.Errorf("eip7702: checkAuthorization: %w: %v, want %s", ErrInvalidChainID, auth.ChainID, chainID)
	}
	if !auth.Address.Equal(delegate) {
		return fmt.Errorf("eip7702: checkAuthorization: %w: %v, want %s", ErrInvalidDelegate, auth.Address, delegate)
	}
	if auth.Nonce == math.MaxUint64 {
		return fmt.Errorf("eip7702: checkAuthorization: %w", ErrInvalidNonce)
	}
	if auth.Signature.IsZero() {
		return fmt.Errorf("eip7702: checkAuthorization: %w: unsigned", ErrInvalidSignature)
	}

	signer, err := auth.Authority()
	if err != nil {
		return fmt.Errorf("eip7702: checkAuthorization: %w: %w", ErrInvalidSignature, err)
	}
	if !signer.Equal(authority) {
		return fmt.Errorf("eip7702: checkAuthorization: %w: signed by %s, want %s", ErrInvalidAuthority, signer, authority)
	}
	return nil
}
