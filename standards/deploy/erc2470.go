package deploy

import (
	"context"
	"errors"
	"fmt"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/eip1014"
)

var (
	// DefaultERC2470SingletonFactory is the EIP-2470 Singleton Factory, deployed at the same
	// address on chains that support it.
	DefaultERC2470SingletonFactory = types.MustNewAddressFromHex("0xce0042B868300000d44A59004Da54A005ffdcf9f")

	erc2470DeployFn = abi.NewFunction("deploy", abi.NewTypes(abi.Bytes, abi.Bytes32), abi.NewTypes(abi.Address))
)

// ErrDeploymentFailed is returned when the Singleton Factory's deploy returns the zero address,
// which is how it reports a failed CREATE2 instead of reverting.
var ErrDeploymentFailed = errors.New("deployment failed")

// ERC2470Deployer deploys contracts via CREATE2 through the EIP-2470 Singleton Factory.
type ERC2470Deployer struct {
	c rpc.Client
}

// NewERC2470Deployer returns an ERC2470Deployer sending deployments via cli.
func NewERC2470Deployer(cli rpc.Client) *ERC2470Deployer {
	return &ERC2470Deployer{
		c: cli,
	}
}

// CallDeploy runs deploy(initCode, salt) via eth_call, returning the address it would deploy to or
// ErrDeploymentFailed when the factory returns the zero address.
func (d *ERC2470Deployer) CallDeploy(ctx context.Context, opts *bind.CallOptions, req *Create2Request) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("deploy: callERC2470: %w", err)
	}

	calldata, err := EncodeERC2470Deploy(req.Deployment)
	if err != nil {
		return nil, fmt.Errorf("deploy: callERC2470: %w", err)
	}

	data, err := bind.Call(ctx, d.c, opts, req.Factory, calldata)
	if err != nil {
		return nil, fmt.Errorf("deploy: callERC2470: %w", err)
	}

	addr, err := erc2470DeployFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("deploy: callERC2470: %w", err)
	}

	if addr.Equal(types.ZeroAddress()) {
		return nil, fmt.Errorf("deploy: callERC2470: %w", ErrDeploymentFailed)
	}

	return addr, nil
}

// Deploy sends deploy(initCode, salt) to req.Factory, signed by opts.Signer. The factory does not
// revert on a failed CREATE2, so check CallDeploy first or the code at DeployAddress afterwards.
func (d *ERC2470Deployer) Deploy(ctx context.Context, opts *bind.TransactOptions, req *Create2Request) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("deploy: erc2470: %w", err)
	}

	calldata, err := EncodeERC2470Deploy(req.Deployment)
	if err != nil {
		return nil, nil, fmt.Errorf("deploy: erc2470: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, d.c, opts, req.Factory, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("deploy: erc2470: %w", err)
	}

	return signed, hash, nil
}

// DeployAddress returns the address req deploys to through req.Factory (EIP-1014).
func (d *ERC2470Deployer) DeployAddress(req *Create2Request) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("deploy: deployAddress: %w", err)
	}

	initCode, err := req.Deployment.InitCode()
	if err != nil {
		return nil, fmt.Errorf("deploy: deployAddress: %w", err)
	}

	return eip1014.Address(req.Factory, req.Deployment.Salt, initCode), nil
}

// EncodeERC2470Deploy returns the calldata for deploy(deployment.InitCode(), deployment.Salt).
func EncodeERC2470Deploy(deployment *DeterministicContractDeployment) ([]byte, error) {
	initCode, err := deployment.InitCode()
	if err != nil {
		return nil, fmt.Errorf("deploy: encodeERC2470Deploy: %w", err)
	}

	data, err := erc2470DeployFn.EncodeCall(initCode, deployment.Salt[:])
	if err != nil {
		return nil, fmt.Errorf("deploy: encodeERC2470Deploy: %w", err)
	}

	return data, nil
}
