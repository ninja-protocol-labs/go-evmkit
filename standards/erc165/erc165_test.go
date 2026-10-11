package erc165

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

// Well-known ERC-165 interface IDs used by the live tests below.
var (
	interfaceIDERC165  = [4]byte{0x01, 0xff, 0xc9, 0xa7}
	interfaceIDERC721  = [4]byte{0x80, 0xac, 0x58, 0xcd}
	interfaceIDInvalid = [4]byte{0xff, 0xff, 0xff, 0xff} // reserved "unsupported" sentinel, must always be false
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

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 60*time.Second)

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

func (s *ERC165Suite) TestSupportsInterfaceERC165() {
	ok, err := s.erc165.SupportsInterface(s.ctx, s.contract, interfaceIDERC165, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(ok)
}

func (s *ERC165Suite) TestSupportsInterfaceERC721() {
	ok, err := s.erc165.SupportsInterface(s.ctx, s.contract, interfaceIDERC721, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(ok)
}

func (s *ERC165Suite) TestSupportsInterfaceUnsupported() {
	ok, err := s.erc165.SupportsInterface(s.ctx, s.contract, interfaceIDInvalid, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().False(ok)
}

func (s *ERC165Suite) TestSupportsInterfaceNotImplemented() {
	ok, err := s.erc165.SupportsInterface(s.ctx, s.nonERC165Contract, interfaceIDERC165, rpc.BlockTagLatest)
	s.Require().Error(err)
	s.Require().False(ok)
	s.Require().ErrorIs(err, ErrNotImplementedERC165)
}
