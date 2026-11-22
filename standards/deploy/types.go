package deploy

import (
	"fmt"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// CreateRequest is a CREATE deployment of Deployment.
type CreateRequest struct {
	Deployment *ContractDeployment
}

func NewCreateRequest(deployment *ContractDeployment) *CreateRequest {
	return &CreateRequest{
		Deployment: deployment,
	}
}

func (r *CreateRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Deployment == nil {
		return fmt.Errorf("%w: deployment is nil", bind.ErrInvalidRequest)
	}

	if len(r.Deployment.Bytecode) == 0 {
		return fmt.Errorf("%w: deployment bytecode is empty", bind.ErrInvalidRequest)
	}

	return nil
}

// Create2Request is a CREATE2 deployment of Deployment through the factory at Factory.
type Create2Request struct {
	Factory    *types.Address
	Deployment *DeterministicContractDeployment
}

func NewCreate2Request(factory *types.Address, deployment *DeterministicContractDeployment) *Create2Request {
	return &Create2Request{
		Factory:    factory,
		Deployment: deployment,
	}
}

func (r *Create2Request) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Factory == nil {
		return fmt.Errorf("%w: factory is nil", bind.ErrInvalidRequest)
	}

	if r.Deployment == nil || r.Deployment.ContractDeployment == nil {
		return fmt.Errorf("%w: deployment is nil", bind.ErrInvalidRequest)
	}

	if len(r.Deployment.Bytecode) == 0 {
		return fmt.Errorf("%w: deployment bytecode is empty", bind.ErrInvalidRequest)
	}

	return nil
}
