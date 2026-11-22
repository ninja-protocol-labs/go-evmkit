package deploy

import (
	"context"
	"fmt"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

// CreateDeployer deploys contracts via a plain CREATE transaction (no to address).
type CreateDeployer struct {
	c rpc.Client
}

// NewCreateDeployer returns a CreateDeployer sending deployments via cli.
func NewCreateDeployer(cli rpc.Client) *CreateDeployer {
	return &CreateDeployer{
		c: cli,
	}
}

// Deploy sends a contract-creation transaction for req.Deployment, signed by opts.Signer;
// opts.Value is passed to the constructor.
func (d *CreateDeployer) Deploy(ctx context.Context, opts *bind.TransactOptions, req *CreateRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("deploy: create: %w", err)
	}

	initCode, err := req.Deployment.InitCode()
	if err != nil {
		return nil, nil, fmt.Errorf("deploy: create: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, d.c, opts, nil, initCode)
	if err != nil {
		return signed, nil, fmt.Errorf("deploy: create: %w", err)
	}

	return signed, hash, nil
}

// DeployAddress returns the address signed deploys to, from its sender and nonce.
func (d *CreateDeployer) DeployAddress(signed core.Transaction) (*types.Address, error) {
	if signed == nil {
		return nil, fmt.Errorf("deploy: deployAddress: %w: transaction is nil", bind.ErrInvalidRequest)
	}

	from, err := signed.Sender()
	if err != nil {
		return nil, fmt.Errorf("deploy: deployAddress: %w", err)
	}

	addr, err := core.CreateAddress(from, signed.AddressNonce())
	if err != nil {
		return nil, fmt.Errorf("deploy: deployAddress: %w", err)
	}

	return addr, nil
}
