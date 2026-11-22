package erc165

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

var (
	interfaceIDERC165  = [4]byte{0x01, 0xff, 0xc9, 0xa7}
	interfaceIDERC721  = [4]byte{0x80, 0xac, 0x58, 0xcd}
	interfaceIDInvalid = [4]byte{0xff, 0xff, 0xff, 0xff}
)

type ERC165Suite struct {
	suite.Suite

	Enable               bool
	RPCURL               string
	ContractHex          string
	NonERC165ContractHex string

	ctx    context.Context
	cancel context.CancelFunc

	client            rpc.Client
	contract          *types.Address
	nonERC165Contract *types.Address
	erc165            *ERC165
}

func TestERC165Suite(t *testing.T) {
	suite.Run(t, &ERC165Suite{
		Enable:               false,
		RPCURL:               "",
		ContractHex:          "",
		NonERC165ContractHex: "",
	})
}

func (s *ERC165Suite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(s.T().Context(), 60*time.Second)

	s.client = rpc.NewClient(s.RPCURL)

	var err error
	s.contract, err = types.NewAddressFromHex(s.ContractHex)
	s.Require().NoError(err)

	s.nonERC165Contract, err = types.NewAddressFromHex(s.NonERC165ContractHex)
	s.Require().NoError(err)

	s.erc165 = New(s.client)
}

func (s *ERC165Suite) TearDownSuite() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *ERC165Suite) Test01SupportsInterfaceERC165() {
	ok, err := s.erc165.SupportsInterface(s.ctx, nil, NewSupportsInterfaceRequest(s.contract, interfaceIDERC165))
	s.Require().NoError(err)
	s.Require().True(ok)
}

func (s *ERC165Suite) Test02SupportsInterfaceERC721() {
	ok, err := s.erc165.SupportsInterface(s.ctx, nil, NewSupportsInterfaceRequest(s.contract, interfaceIDERC721))
	s.Require().NoError(err)
	s.Require().True(ok)
}

func (s *ERC165Suite) Test03SupportsInterfaceUnsupported() {
	ok, err := s.erc165.SupportsInterface(s.ctx, nil, NewSupportsInterfaceRequest(s.contract, interfaceIDInvalid))
	s.Require().NoError(err)
	s.Require().False(ok)
}

func (s *ERC165Suite) Test04SupportsInterfaceNotImplemented() {
	ok, err := s.erc165.SupportsInterface(s.ctx, nil, NewSupportsInterfaceRequest(s.nonERC165Contract, interfaceIDERC165))
	s.Require().Error(err)
	s.Require().False(ok)
	s.Require().ErrorIs(err, ErrNotImplementedERC165)
}
