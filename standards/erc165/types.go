package erc165

import (
	"fmt"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// SupportsInterfaceRequest is supportsInterface(interfaceId)'s input on Contract.
type SupportsInterfaceRequest struct {
	Contract    *types.Address
	InterfaceID [4]byte
}

func NewSupportsInterfaceRequest(contract *types.Address, interfaceID [4]byte) *SupportsInterfaceRequest {
	return &SupportsInterfaceRequest{
		Contract:    contract,
		InterfaceID: interfaceID,
	}
}

func (r *SupportsInterfaceRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}
