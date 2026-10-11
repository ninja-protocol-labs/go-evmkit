package deploy

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

// CreateDeployer deploys contracts via a plain CREATE (to is nil) transaction.
type CreateDeployer struct {
	c rpc.Client
}

// NewCreateDeployer returns a CreateDeployer broadcasting deployments via cli.
func NewCreateDeployer(cli rpc.Client) *CreateDeployer {
	return &CreateDeployer{
		c: cli,
	}
}

// CreateDeployment builds an unsigned contract-creation transaction via b.
func (d *CreateDeployer) CreateDeployment(
	ctx context.Context,
	b tx.Builder,
	from *types.Address,
	value *big.Int,
	deployment *ContractDeployment,
) (core.Transaction, error) {
	initCode, err := deployment.InitCode()
	if err != nil {
		return nil, fmt.Errorf("deploy: createDeployment: %w", err)
	}

	unsigned, err := b.Build(from, nil, value, initCode).Pack(ctx, d.c)
	if err != nil {
		return nil, fmt.Errorf("deploy: createDeployment: %w", err)
	}
	return unsigned, nil
}

// DeployAddress returns the address a CREATE from from with nonce deploys
// to. nonce must be the nonce the signed transaction actually carries (e.g.
// by reading it back via the same Packer's Nonce before Pack).
func (d *CreateDeployer) DeployAddress(from *types.Address, nonce uint64) (*types.Address, error) {
	addr, err := core.CreateAddress(from, nonce)
	if err != nil {
		return nil, fmt.Errorf("deploy: deployAddress: %w", err)
	}
	return addr, nil
}

// Deploy broadcasts an already-signed contract-creation transaction.
func (d *CreateDeployer) Deploy(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("deploy: deploy: %w", err)
	}

	hash, err := tx.Broadcast(ctx, d.c, raw)
	if err != nil {
		return nil, fmt.Errorf("deploy: deploy: %w", err)
	}

	return hash, nil
}
