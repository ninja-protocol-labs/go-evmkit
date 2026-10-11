package deploy

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

var (
	// DefaultERC2470SingletonFactory is the CREATE2 factory standardized by EIP-2470
	// ("Singleton Factory"), deployed at the same address on chains that support it.
	DefaultERC2470SingletonFactory = types.MustNewAddressFromHex("0xce0042B868300000d44A59004Da54A005ffdcf9f")

	// erc2470DeployFn describes the Singleton Factory's deploy(bytes,bytes32) function.
	erc2470DeployFn = abi.NewFunction("deploy", abi.Types{abi.Bytes, abi.Bytes32}, abi.Types{abi.Address})
)

// ERC2470Deployer deploys contracts via CREATE2 through the EIP-2470
// Singleton Factory.
type ERC2470Deployer struct {
	c rpc.Client
}

// NewERC2470Deployer returns an ERC2470Deployer broadcasting deployments via cli.
func NewERC2470Deployer(cli rpc.Client) *ERC2470Deployer {
	return &ERC2470Deployer{
		c: cli,
	}
}

// CreateDeployment builds an unsigned transaction calling to's
// deploy(deployment.InitCode(), deployment.Salt) via b. to is the Singleton
// Factory address; pass DefaultERC2470SingletonFactory unless the chain
// deploys it somewhere else.
func (d *ERC2470Deployer) CreateDeployment(
	ctx context.Context,
	b tx.Builder,
	from, factory *types.Address,
	value *big.Int,
	deployment *DeterministicContractDeployment,
) (core.Transaction, error) {
	initCode, err := deployment.InitCode()
	if err != nil {
		return nil, fmt.Errorf("deploy: createDeployment: %w", err)
	}

	calldata, err := erc2470DeployFn.EncodeCall(initCode, deployment.Salt[:])
	if err != nil {
		return nil, fmt.Errorf("deploy: createDeployment: %w", err)
	}

	unsigned, err := b.Build(from, factory, value, calldata).Pack(ctx, d.c)
	if err != nil {
		return nil, fmt.Errorf("deploy: createDeployment: %w", err)
	}
	return unsigned, nil
}

// DeployAddress returns the address deployment deploys to via the
// Singleton Factory at factory (EIP-1014), independent of any transaction.
func (d *ERC2470Deployer) DeployAddress(factory *types.Address, deployment *DeterministicContractDeployment) (*types.Address, error) {
	initCode, err := deployment.InitCode()
	if err != nil {
		return nil, fmt.Errorf("deploy: deployAddress: %w", err)
	}
	return core.EIP1014Address(factory, deployment.Salt, initCode), nil
}

// Deploy broadcasts an already-signed deploy(bytes,bytes32) transaction.
func (d *ERC2470Deployer) Deploy(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
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
