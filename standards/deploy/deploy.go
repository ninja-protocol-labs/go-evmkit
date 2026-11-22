// Package deploy deploys contracts via CREATE (a contract-creation transaction) or CREATE2 (through a
// deterministic deployment factory). Pick the deployer matching how the address should be determined.
package deploy

import (
	"context"
	"fmt"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// ICreateDeployer is implemented by *CreateDeployer.
type ICreateDeployer interface {
	Deploy(ctx context.Context, opts *bind.TransactOptions, req *CreateRequest) (core.Transaction, *types.Hash, error)
	DeployAddress(signed core.Transaction) (*types.Address, error)
}

// ICreate2Deployer is implemented by *NicksDeployer and *ERC2470Deployer.
type ICreate2Deployer interface {
	Deploy(ctx context.Context, opts *bind.TransactOptions, req *Create2Request) (core.Transaction, *types.Hash, error)
	DeployAddress(req *Create2Request) (*types.Address, error)
}

// IERC2470Deployer adds the Singleton Factory's deploy simulation to ICreate2Deployer.
type IERC2470Deployer interface {
	ICreate2Deployer

	CallDeploy(ctx context.Context, opts *bind.CallOptions, req *Create2Request) (*types.Address, error)
}

var (
	_ ICreateDeployer  = (*CreateDeployer)(nil)
	_ ICreate2Deployer = (*NicksDeployer)(nil)
	_ IERC2470Deployer = (*ERC2470Deployer)(nil)
)

// ContractDeployment is a contract's creation bytecode plus its constructor's arguments.
type ContractDeployment struct {
	Bytecode    []byte
	Constructor abi.Types
	Args        []any
}

func NewContractDeployment(code []byte, constructor abi.Types, args []any) *ContractDeployment {
	return &ContractDeployment{
		Bytecode:    code,
		Constructor: constructor,
		Args:        args,
	}
}

// InitCode returns Bytecode followed by Args ABI-encoded per Constructor.
func (d *ContractDeployment) InitCode() ([]byte, error) {
	encodedArgs, err := abi.Pack(d.Constructor, d.Args...)
	if err != nil {
		return nil, fmt.Errorf("deploy: initCode: %w", err)
	}

	initCode := make([]byte, 0, len(d.Bytecode)+len(encodedArgs))
	initCode = append(initCode, d.Bytecode...)
	initCode = append(initCode, encodedArgs...)
	return initCode, nil
}

// DeterministicContractDeployment is a ContractDeployment plus the CREATE2 salt.
type DeterministicContractDeployment struct {
	*ContractDeployment
	Salt [32]byte
}

func NewDeterministicContractDeployment(deployment *ContractDeployment, salt [32]byte) *DeterministicContractDeployment {
	return &DeterministicContractDeployment{
		ContractDeployment: deployment,
		Salt:               salt,
	}
}
