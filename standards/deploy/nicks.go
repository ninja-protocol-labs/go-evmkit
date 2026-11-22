package deploy

import (
	"context"
	"fmt"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/eip1014"
)

// DefaultNicksMethodFactory is the CREATE2 deterministic deployment proxy ("Nick's method"), deployed
// at the same address on most EVM chains via a keyless signature.
var DefaultNicksMethodFactory = types.MustNewAddressFromHex("0x4e59b44847b379578588920cA78FbF26c0B4956C")

// NicksDeployer deploys contracts via CREATE2 through the Nick's method proxy, whose calldata is
// salt ++ initCode with no selector.
type NicksDeployer struct {
	c rpc.Client
}

// NewNicksDeployer returns a NicksDeployer sending deployments via cli.
func NewNicksDeployer(cli rpc.Client) *NicksDeployer {
	return &NicksDeployer{
		c: cli,
	}
}

// Deploy sends salt ++ initCode to req.Factory, signed by opts.Signer. The proxy reverts when
// CREATE2 fails, so a failed deployment surfaces as an estimate error.
func (d *NicksDeployer) Deploy(ctx context.Context, opts *bind.TransactOptions, req *Create2Request) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("deploy: nicks: %w", err)
	}

	initCode, err := req.Deployment.InitCode()
	if err != nil {
		return nil, nil, fmt.Errorf("deploy: nicks: %w", err)
	}

	calldata := make([]byte, 0, len(req.Deployment.Salt)+len(initCode))
	calldata = append(calldata, req.Deployment.Salt[:]...)
	calldata = append(calldata, initCode...)

	signed, hash, err := bind.Transact(ctx, d.c, opts, req.Factory, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("deploy: nicks: %w", err)
	}

	return signed, hash, nil
}

// DeployAddress returns the address req deploys to through req.Factory (EIP-1014).
func (d *NicksDeployer) DeployAddress(req *Create2Request) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("deploy: deployAddress: %w", err)
	}

	initCode, err := req.Deployment.InitCode()
	if err != nil {
		return nil, fmt.Errorf("deploy: deployAddress: %w", err)
	}

	return eip1014.Address(req.Factory, req.Deployment.Salt, initCode), nil
}
