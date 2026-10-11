// Package deploy deploys contracts via CREATE (a plain contract-creation
// transaction) or CREATE2 (via a well-known deterministic deployment
// factory). Each deployer is a standalone type; pick the one matching how
// you want the deployed address determined.
package deploy

import (
	"context"
	"fmt"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// IDeployer is implemented by *CreateDeployer, *NicksDeployer, and *ERC2470Deployer.
type IDeployer interface {
	Deploy(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

var (
	_ IDeployer = (*CreateDeployer)(nil)
	_ IDeployer = (*ERC2470Deployer)(nil)
	_ IDeployer = (*NicksDeployer)(nil)
)

// ContractDeployment is a contract's creation bytecode plus its
// constructor's arguments, shared by every Deployer implementation.
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
