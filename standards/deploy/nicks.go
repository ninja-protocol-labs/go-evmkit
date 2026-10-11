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

// DefaultNicksMethodFactory is the commonly-used CREATE2 deterministic
// deployment proxy ("Nick's method"), deployed at the same address on most
// EVM chains via a keyless signature. Not standardized by any EIP/ERC.
var DefaultNicksMethodFactory = types.MustNewAddressFromHex("0x4e59b44847b379578588920cA78FbF26c0B4956C")

// NicksDeployer deploys standards via CREATE2 through the Nick's method
// deterministic deployment proxy; the deployed address is computed from the
// signed transaction's own calldata (salt ++ initCode, no selector).
type NicksDeployer struct {
	c rpc.Client
}

// NewNicksDeployer returns a NicksDeployer broadcasting deployments via cli.
func NewNicksDeployer(cli rpc.Client) *NicksDeployer {
	return &NicksDeployer{
		c: cli,
	}
}

// CreateDeployment builds an unsigned transaction calling to with calldata
// = deployment.Salt ++ deployment.InitCode(), via b. to is the Nick's
// method proxy address; pass DefaultNicksMethodFactory unless the chain
// deploys it somewhere else. Unlike the EIP-2470 Singleton Factory, the
// proxy reverts if the CREATE2 fails, so a failed deployment surfaces as a
// Pack/broadcast error rather than silently succeeding with no code
// deployed.
func (d *NicksDeployer) CreateDeployment(
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

	calldata := make([]byte, 0, len(deployment.Salt)+len(initCode))
	calldata = append(calldata, deployment.Salt[:]...)
	calldata = append(calldata, initCode...)

	unsigned, err := b.Build(from, factory, value, calldata).Pack(ctx, d.c)
	if err != nil {
		return nil, fmt.Errorf("deploy: createDeployment: %w", err)
	}
	return unsigned, nil
}

// DeployAddress returns the address deployment deploys to via the Nick's
// method proxy at factory (EIP-1014), independent of any transaction.
func (d *NicksDeployer) DeployAddress(factory *types.Address, deployment *DeterministicContractDeployment) (*types.Address, error) {
	initCode, err := deployment.InitCode()
	if err != nil {
		return nil, fmt.Errorf("deploy: deployAddress: %w", err)
	}
	return core.EIP1014Address(factory, deployment.Salt, initCode), nil
}

// Deploy broadcasts an already-signed CREATE2-via-Nick's-method transaction.
func (d *NicksDeployer) Deploy(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
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
