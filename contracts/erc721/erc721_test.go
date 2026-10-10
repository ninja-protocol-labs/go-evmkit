package erc721

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

var interfaceIDERC721 = [4]byte{0x80, 0xac, 0x58, 0xcd}

type ERC721Suite struct {
	suite.Suite

	Enable        bool
	RPCURL        string
	ContractHex   string
	ContractBHex  string
	PrivateKeyHex string
	ToHex         string
	MulticallHex  string

	ctx    context.Context
	cancel context.CancelFunc

	client rpc.Client

	signer    core.Signer
	from      *types.Address
	contract  *types.Address
	contractB *types.Address
	to        *types.Address
	multicall *types.Address

	erc721 *ERC721
}

func TestERC721Suite(t *testing.T) {
	suite.Run(t, &ERC721Suite{
		Enable:        false,
		RPCURL:        "",
		ContractHex:   "",
		ContractBHex:  "",
		PrivateKeyHex: "",
		ToHex:         "",
		MulticallHex:  "",
	})
}

func (s *ERC721Suite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 60*time.Second)

	s.client = rpc.NewClient(s.RPCURL)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())

	s.contract, err = types.NewAddressFromHex(s.ContractHex)
	s.Require().NoError(err)

	s.contractB, err = types.NewAddressFromHex(s.ContractBHex)
	s.Require().NoError(err)

	s.to, err = types.NewAddressFromHex(s.ToHex)
	s.Require().NoError(err)

	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)

	s.erc721 = New(s.client)
}

func (s *ERC721Suite) TearDownSuite() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *ERC721Suite) TestReadMethods() {
	supports, err := s.erc721.SupportsInterface(s.ctx, s.contract, interfaceIDERC721, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(supports)

	name, err := s.erc721.Name(s.ctx, s.contract, rpc.BlockTagLatest)
	s.Require().NoError(err)

	symbol, err := s.erc721.Symbol(s.ctx, s.contract, rpc.BlockTagLatest)
	s.Require().NoError(err)

	tokenId := big.NewInt(1)

	tokenURI, err := s.erc721.TokenURI(s.ctx, s.contract, tokenId, rpc.BlockTagLatest)
	s.Require().NoError(err)

	fromBalance, err := s.erc721.BalanceOf(s.ctx, s.contract, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)

	toBalance, err := s.erc721.BalanceOf(s.ctx, s.contract, s.to, rpc.BlockTagLatest)
	s.Require().NoError(err)

	owner, err := s.erc721.OwnerOf(s.ctx, s.contract, tokenId, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(owner.Equal(s.from))

	approved, err := s.erc721.GetApproved(s.ctx, s.contract, tokenId, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(approved.Equal(s.to))

	isApprovedForAll, err := s.erc721.IsApprovedForAll(s.ctx, s.contract, s.from, s.to, rpc.BlockTagLatest)
	s.Require().NoError(err)

	s.T().Logf("ERC721 %s TestReadMethods Result", s.contract)
	s.T().Logf("\tSupports: %v", supports)
	s.T().Logf("\tName: %s", name)
	s.T().Logf("\tSymbol: %s", symbol)
	s.T().Logf("\tTokenURI(%s): %s", tokenId, tokenURI)
	s.T().Logf("\tBalanceOf")
	s.T().Logf("\t\tFrom(%s): %s", s.from, fromBalance)
	s.T().Logf("\t\tTo(%s): %s", s.to, toBalance)
	s.T().Logf("\tOwnerOf(%s): %s", tokenId, owner)
	s.T().Logf("\tGetApproved(%s): %s", tokenId, approved)
	s.T().Logf("\tIsApprovedForAll(from, to): %v", isApprovedForAll)
}

func (s *ERC721Suite) TestApprove() {
	tokenId := big.NewInt(1)

	before, err := s.erc721.GetApproved(s.ctx, s.contract, tokenId, rpc.BlockTagLatest)
	s.Require().NoError(err)

	gas, err := s.erc721.ApproveGas(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.from, s.to, tokenId)
	s.Require().NoError(err)

	unsigned, err := s.erc721.CreateApprove(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.from, s.to, tokenId)
	s.Require().NoError(err)

	s.Require().NoError(unsigned.Sign(s.signer))

	hash, err := s.erc721.Approve(s.ctx, unsigned)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, 0, 0)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transaction reverted")

	approvals, err := ExtractApprovals(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(approvals, 1)
	s.Require().True(approvals[0].Contract.Equal(s.contract))
	s.Require().True(approvals[0].Owner.Equal(s.from))
	s.Require().True(approvals[0].Approved.Equal(s.to))
	s.Require().Zero(tokenId.Cmp(approvals[0].TokenId))

	var after *types.Address
	s.Require().Eventually(func() bool {
		after, err = s.erc721.GetApproved(s.ctx, s.contract, tokenId, rpc.BlockTagLatest)
		return err == nil && after.Equal(s.to)
	}, 10*time.Second, 500*time.Millisecond, "approval not visible after approve")

	s.T().Logf("ERC721 %s TestApprove Result", s.contract)
	s.T().Logf("\tOwner: %s", s.from)
	s.T().Logf("\tApproved: %s", s.to)
	s.T().Logf("\tTokenId: %s", tokenId)
	s.T().Logf("\tTx: %s", hash)
	s.T().Logf("\tGas: %d", gas)
	s.T().Logf("\tGetApproved")
	s.T().Logf("\t\tBefore: %s", before)
	s.T().Logf("\t\tAfter: %s", after)
}

func (s *ERC721Suite) TestApproveInvalidApprover() {
	tokenId := big.NewInt(1)

	_, err := s.erc721.ApproveGas(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.to, s.from, tokenId)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInvalidApprover)

	var rpcErr *rpc.ResponseError
	s.Require().ErrorAs(err, &rpcErr)
	s.Require().NotEmpty(rpcErr.Data)

	s.T().Logf("ERC721 %s TestApproveInvalidApprover Result", s.contract)
	s.T().Logf("\tApprover (unauthorized): %s", s.to)
	s.T().Logf("\tTokenId: %s", tokenId)
	s.T().Logf("\tError: %v", err)
	s.T().Logf("\tRevertData: %s", rpcErr.Data)
}

func (s *ERC721Suite) TestSetApprovalForAll() {
	before, err := s.erc721.IsApprovedForAll(s.ctx, s.contract, s.from, s.to, rpc.BlockTagLatest)
	s.Require().NoError(err)

	gas, err := s.erc721.SetApprovalForAllGas(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.from, s.to, true)
	s.Require().NoError(err)

	unsigned, err := s.erc721.CreateSetApprovalForAll(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.from, s.to, true)
	s.Require().NoError(err)

	s.Require().NoError(unsigned.Sign(s.signer))

	hash, err := s.erc721.SetApprovalForAll(s.ctx, unsigned)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, 0, 0)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transaction reverted")

	approvals, err := ExtractApprovalForAll(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(approvals, 1)
	s.Require().True(approvals[0].Contract.Equal(s.contract))
	s.Require().True(approvals[0].Owner.Equal(s.from))
	s.Require().True(approvals[0].Operator.Equal(s.to))
	s.Require().True(approvals[0].Approved)

	var after bool
	s.Require().Eventually(func() bool {
		after, err = s.erc721.IsApprovedForAll(s.ctx, s.contract, s.from, s.to, rpc.BlockTagLatest)
		return err == nil && after
	}, 10*time.Second, 500*time.Millisecond, "approval-for-all not visible after setApprovalForAll")

	unsignedRevert, err := s.erc721.CreateSetApprovalForAll(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.from, s.to, false)
	s.Require().NoError(err)
	s.Require().NoError(unsignedRevert.Sign(s.signer))

	revertHash, err := s.erc721.SetApprovalForAll(s.ctx, unsignedRevert)
	s.Require().NoError(err)

	revertReceipt, err := tx.WaitForReceipt(s.ctx, s.client, revertHash, 0, 0)
	s.Require().NoError(err)
	s.Require().True(revertReceipt.Status, "revert transaction reverted")

	revertApprovals, err := ExtractApprovalForAll(revertReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(revertApprovals, 1)
	s.Require().False(revertApprovals[0].Approved)

	var reverted bool
	s.Require().Eventually(func() bool {
		reverted, err = s.erc721.IsApprovedForAll(s.ctx, s.contract, s.from, s.to, rpc.BlockTagLatest)
		return err == nil && !reverted
	}, 10*time.Second, 500*time.Millisecond, "approval-for-all not reverted to false")

	s.T().Logf("ERC721 %s TestSetApprovalForAll Result", s.contract)
	s.T().Logf("\tOwner: %s", s.from)
	s.T().Logf("\tOperator: %s", s.to)
	s.T().Logf("\tTx: %s", hash)
	s.T().Logf("\tGas: %d", gas)
	s.T().Logf("\tIsApprovedForAll")
	s.T().Logf("\t\tBefore: %v", before)
	s.T().Logf("\t\tAfter: %v", after)
	s.T().Logf("\tRevertTx: %s", revertHash)
	s.T().Logf("\t\tReverted: %v", reverted)
}

func (s *ERC721Suite) TestSetApprovalForAllInvalidOperator() {
	zero := types.ZeroAddress()

	_, err := s.erc721.SetApprovalForAllGas(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.from, zero, true)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInvalidOperator)

	var rpcErr *rpc.ResponseError
	s.Require().ErrorAs(err, &rpcErr)
	s.Require().NotEmpty(rpcErr.Data)

	s.T().Logf("ERC721 %s TestSetApprovalForAllInvalidOperator Result", s.contract)
	s.T().Logf("\tOperator (zero): %s", zero)
	s.T().Logf("\tError: %v", err)
	s.T().Logf("\tRevertData: %s", rpcErr.Data)
}

func (s *ERC721Suite) TestTransferFrom() {
	tokenId := big.NewInt(1)

	gas, err := s.erc721.TransferFromGas(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.from, s.from, s.from, tokenId)
	s.Require().NoError(err)

	unsigned, err := s.erc721.CreateTransferFrom(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.from, s.from, s.from, tokenId)
	s.Require().NoError(err)

	s.Require().NoError(unsigned.Sign(s.signer))

	hash, err := s.erc721.TransferFrom(s.ctx, unsigned)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, 0, 0)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transaction reverted")

	transfers, err := ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].Contract.Equal(s.contract))
	s.Require().True(transfers[0].From.Equal(s.from))
	s.Require().True(transfers[0].To.Equal(s.from))
	s.Require().Zero(tokenId.Cmp(transfers[0].TokenId))

	var owner *types.Address
	s.Require().Eventually(func() bool {
		owner, err = s.erc721.OwnerOf(s.ctx, s.contract, tokenId, rpc.BlockTagLatest)
		return err == nil && owner.Equal(s.from)
	}, 10*time.Second, 500*time.Millisecond, "owner not visible after transferFrom")

	s.T().Logf("ERC721 %s TestTransferFrom Result", s.contract)
	s.T().Logf("\tFrom: %s", s.from)
	s.T().Logf("\tTo: %s", s.from)
	s.T().Logf("\tTokenId: %s", tokenId)
	s.T().Logf("\tTx: %s", hash)
	s.T().Logf("\tGas: %d", gas)
	s.T().Logf("\tOwnerOf: %s", owner)
}

func (s *ERC721Suite) TestTransferFromInvalidReceiver() {
	tokenId := big.NewInt(1)
	zero := types.ZeroAddress()

	_, err := s.erc721.TransferFromGas(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.from, s.from, zero, tokenId)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInvalidReceiver)

	var rpcErr *rpc.ResponseError
	s.Require().ErrorAs(err, &rpcErr)
	s.Require().NotEmpty(rpcErr.Data)

	s.T().Logf("ERC721 %s TestTransferFromInvalidReceiver Result", s.contract)
	s.T().Logf("\tTokenId: %s", tokenId)
	s.T().Logf("\tError: %v", err)
	s.T().Logf("\tRevertData: %s", rpcErr.Data)
}

func (s *ERC721Suite) TestSafeTransferFrom() {
	tokenId := big.NewInt(1)

	gas, err := s.erc721.SafeTransferFromGas(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.from, s.from, s.from, tokenId)
	s.Require().NoError(err)

	unsigned, err := s.erc721.CreateSafeTransferFrom(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.from, s.from, s.from, tokenId)
	s.Require().NoError(err)

	s.Require().NoError(unsigned.Sign(s.signer))

	hash, err := s.erc721.SafeTransferFrom(s.ctx, unsigned)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, 0, 0)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transaction reverted")

	transfers, err := ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].Contract.Equal(s.contract))
	s.Require().True(transfers[0].From.Equal(s.from))
	s.Require().True(transfers[0].To.Equal(s.from))
	s.Require().Zero(tokenId.Cmp(transfers[0].TokenId))

	var owner *types.Address
	s.Require().Eventually(func() bool {
		owner, err = s.erc721.OwnerOf(s.ctx, s.contract, tokenId, rpc.BlockTagLatest)
		return err == nil && owner.Equal(s.from)
	}, 10*time.Second, 500*time.Millisecond, "owner not visible after safeTransferFrom")

	s.T().Logf("ERC721 %s TestSafeTransferFrom Result", s.contract)
	s.T().Logf("\tFrom: %s", s.from)
	s.T().Logf("\tTo: %s", s.from)
	s.T().Logf("\tTokenId: %s", tokenId)
	s.T().Logf("\tTx: %s", hash)
	s.T().Logf("\tGas: %d", gas)
	s.T().Logf("\tOwnerOf: %s", owner)
}

func (s *ERC721Suite) TestSafeTransferFromWithData() {
	tokenId := big.NewInt(1)

	var callData []byte
	gas, err := s.erc721.SafeTransferFromWithDataGas(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.from, s.from, s.from, tokenId, callData)
	s.Require().NoError(err)

	unsigned, err := s.erc721.CreateSafeTransferFromWithData(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.from, s.from, s.from, tokenId, callData)
	s.Require().NoError(err)

	s.Require().NoError(unsigned.Sign(s.signer))

	hash, err := s.erc721.SafeTransferFrom(s.ctx, unsigned)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, 0, 0)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transaction reverted")

	transfers, err := ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].Contract.Equal(s.contract))
	s.Require().Zero(tokenId.Cmp(transfers[0].TokenId))

	s.T().Logf("ERC721 %s TestSafeTransferFromWithData Result", s.contract)
	s.T().Logf("\tFrom: %s", s.from)
	s.T().Logf("\tTo: %s", s.from)
	s.T().Logf("\tTokenId: %s", tokenId)
	s.T().Logf("\tTx: %s", hash)
	s.T().Logf("\tGas: %d", gas)
}

func (s *ERC721Suite) TestSafeTransferFromInvalidReceiver() {
	tokenId := big.NewInt(1)
	zero := types.ZeroAddress()

	_, err := s.erc721.SafeTransferFromGas(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.from, s.from, zero, tokenId)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInvalidReceiver)

	var rpcErr *rpc.ResponseError
	s.Require().ErrorAs(err, &rpcErr)
	s.Require().NotEmpty(rpcErr.Data)

	s.T().Logf("ERC721 %s TestSafeTransferFromInvalidReceiver Result", s.contract)
	s.T().Logf("\tTokenId: %s", tokenId)
	s.T().Logf("\tError: %v", err)
	s.T().Logf("\tRevertData: %s", rpcErr.Data)
}

func (s *ERC721Suite) TestMetadata() {
	m, err := s.erc721.Metadata(s.ctx, s.multicall, s.contract, false, rpc.BlockTagLatest)
	s.Require().NoError(err)

	name, err := s.erc721.Name(s.ctx, s.contract, rpc.BlockTagLatest)
	s.Require().NoError(err)
	symbol, err := s.erc721.Symbol(s.ctx, s.contract, rpc.BlockTagLatest)
	s.Require().NoError(err)

	s.Require().Equal(name, m.Name)
	s.Require().Equal(symbol, m.Symbol)

	s.T().Logf("ERC721 %s TestMetadata Result", s.contract)
	s.T().Logf("\tName: %s", m.Name)
	s.T().Logf("\tSymbol: %s", m.Symbol)
}

func (s *ERC721Suite) TestMetadataAllowFailure() {
	m, err := s.erc721.Metadata(s.ctx, s.multicall, s.multicall, true, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Empty(m.Name)
	s.Require().Empty(m.Symbol)

	s.T().Logf("ERC721 %s TestMetadataAllowFailure Result", s.multicall)
	s.T().Logf("\tName: %q", m.Name)
	s.T().Logf("\tSymbol: %q", m.Symbol)
}

func (s *ERC721Suite) TestMetadataWithTokenID() {
	tokenId := big.NewInt(1)

	mt, err := s.erc721.MetadataWithTokenID(s.ctx, s.multicall, s.contract, tokenId, false, rpc.BlockTagLatest)
	s.Require().NoError(err)

	name, err := s.erc721.Name(s.ctx, s.contract, rpc.BlockTagLatest)
	s.Require().NoError(err)
	symbol, err := s.erc721.Symbol(s.ctx, s.contract, rpc.BlockTagLatest)
	s.Require().NoError(err)
	tokenURI, err := s.erc721.TokenURI(s.ctx, s.contract, tokenId, rpc.BlockTagLatest)
	s.Require().NoError(err)
	owner, err := s.erc721.OwnerOf(s.ctx, s.contract, tokenId, rpc.BlockTagLatest)
	s.Require().NoError(err)

	s.Require().Equal(name, mt.Name)
	s.Require().Equal(symbol, mt.Symbol)
	s.Require().Zero(tokenId.Cmp(mt.TokenId))
	s.Require().Equal(tokenURI, mt.TokenURI)
	s.Require().True(owner.Equal(mt.Owner))

	s.T().Logf("ERC721 %s TestMetadataWithTokenID Result", s.contract)
	s.T().Logf("\tName: %s", mt.Name)
	s.T().Logf("\tSymbol: %s", mt.Symbol)
	s.T().Logf("\tTokenId: %s", mt.TokenId)
	s.T().Logf("\tTokenURI: %s", mt.TokenURI)
	s.T().Logf("\tOwner: %s", mt.Owner)
}

func (s *ERC721Suite) TestMetadataWithTokenIDAllowFailure() {
	tokenId := big.NewInt(1)

	mt, err := s.erc721.MetadataWithTokenID(s.ctx, s.multicall, s.multicall, tokenId, true, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Empty(mt.Name)
	s.Require().Empty(mt.Symbol)
	s.Require().Empty(mt.TokenURI)
	s.Require().Nil(mt.Owner)

	s.T().Logf("ERC721 %s TestMetadataWithTokenIDAllowFailure Result", s.multicall)
	s.T().Logf("\tName: %q", mt.Name)
	s.T().Logf("\tSymbol: %q", mt.Symbol)
	s.T().Logf("\tTokenURI: %q", mt.TokenURI)
	s.T().Logf("\tOwner: %v", mt.Owner)
}

func (s *ERC721Suite) TestAllowanceWithTokenID() {
	tokenId := big.NewInt(1)

	awt, err := s.erc721.AllowanceWithTokenID(s.ctx, s.multicall, s.contract, tokenId, false, rpc.BlockTagLatest)
	s.Require().NoError(err)

	owner, err := s.erc721.OwnerOf(s.ctx, s.contract, tokenId, rpc.BlockTagLatest)
	s.Require().NoError(err)
	operator, err := s.erc721.GetApproved(s.ctx, s.contract, tokenId, rpc.BlockTagLatest)
	s.Require().NoError(err)

	s.Require().Zero(tokenId.Cmp(awt.TokenId))
	s.Require().True(owner.Equal(awt.Owner))
	s.Require().True(operator.Equal(awt.Operator))

	s.T().Logf("ERC721 %s TestAllowanceWithTokenID Result", s.contract)
	s.T().Logf("\tTokenId: %s", awt.TokenId)
	s.T().Logf("\tOwner: %s", awt.Owner)
	s.T().Logf("\tOperator: %s", awt.Operator)
}

func (s *ERC721Suite) TestAllowanceWithTokenIDAllowFailure() {
	tokenId := big.NewInt(1)

	awt, err := s.erc721.AllowanceWithTokenID(s.ctx, s.multicall, s.multicall, tokenId, true, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Nil(awt.Owner)
	s.Require().Nil(awt.Operator)

	s.T().Logf("ERC721 %s TestAllowanceWithTokenIDAllowFailure Result", s.multicall)
	s.T().Logf("\tTokenId: %s", awt.TokenId)
	s.T().Logf("\tOwner: %v", awt.Owner)
	s.T().Logf("\tOperator: %v", awt.Operator)
}

func (s *ERC721Suite) TestApprovalForAllWithBalance() {
	afawb, err := s.erc721.ApprovalForAllWithBalance(s.ctx, s.multicall, s.contract, s.from, s.to, false, rpc.BlockTagLatest)
	s.Require().NoError(err)

	isApprovedForAll, err := s.erc721.IsApprovedForAll(s.ctx, s.contract, s.from, s.to, rpc.BlockTagLatest)
	s.Require().NoError(err)
	ownerBalance, err := s.erc721.BalanceOf(s.ctx, s.contract, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	operatorBalance, err := s.erc721.BalanceOf(s.ctx, s.contract, s.to, rpc.BlockTagLatest)
	s.Require().NoError(err)

	s.Require().True(afawb.Owner.Equal(s.from))
	s.Require().True(afawb.Operator.Equal(s.to))
	s.Require().Equal(isApprovedForAll, afawb.Approved)
	s.Require().Zero(ownerBalance.Cmp(afawb.OwnerBalance))
	s.Require().Zero(operatorBalance.Cmp(afawb.OperatorBalance))

	s.T().Logf("ERC721 %s TestApprovalForAllWithBalance Result", s.contract)
	s.T().Logf("\tOwner: %s", afawb.Owner)
	s.T().Logf("\tOperator: %s", afawb.Operator)
	s.T().Logf("\tApproved: %v", afawb.Approved)
	s.T().Logf("\tOwnerBalance: %s", afawb.OwnerBalance)
	s.T().Logf("\tOperatorBalance: %s", afawb.OperatorBalance)
}

func (s *ERC721Suite) TestApprovalForAllWithBalanceAllowFailure() {
	afawb, err := s.erc721.ApprovalForAllWithBalance(s.ctx, s.multicall, s.multicall, s.from, s.to, true, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().False(afawb.Approved)
	s.Require().Nil(afawb.OwnerBalance)
	s.Require().Nil(afawb.OperatorBalance)

	s.T().Logf("ERC721 %s TestApprovalForAllWithBalanceAllowFailure Result", s.multicall)
	s.T().Logf("\tApproved: %v", afawb.Approved)
	s.T().Logf("\tOwnerBalance: %v", afawb.OwnerBalance)
	s.T().Logf("\tOperatorBalance: %v", afawb.OperatorBalance)
}

func (s *ERC721Suite) TestTokenOwners() {
	tokenIds := []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(2)}

	tos, err := s.erc721.TokenOwners(s.ctx, s.multicall, s.contract, tokenIds, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(tos.Contract.Equal(s.contract))
	s.Require().Len(tos.Tokens, len(tokenIds))

	s.T().Logf("ERC721 %s TestTokenOwners Result", s.contract)
	for i, tokenId := range tokenIds {
		owner, err := s.erc721.OwnerOf(s.ctx, s.contract, tokenId, rpc.BlockTagLatest)
		s.Require().NoError(err)

		s.Require().Zero(tokenId.Cmp(tos.Tokens[i].TokenId))
		s.Require().True(tos.Tokens[i].Owner.Equal(owner))
		s.T().Logf("\t%s: %s", tokenId, tos.Tokens[i].Owner)
	}
}

func (s *ERC721Suite) TestTokenOperators() {
	tokenIds := []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(2)}

	tops, err := s.erc721.TokenOperators(s.ctx, s.multicall, s.contract, tokenIds, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(tops.Contract.Equal(s.contract))
	s.Require().Len(tops.Tokens, len(tokenIds))

	s.T().Logf("ERC721 %s TestTokenOperators Result", s.contract)
	for i, tokenId := range tokenIds {
		operator, err := s.erc721.GetApproved(s.ctx, s.contract, tokenId, rpc.BlockTagLatest)
		s.Require().NoError(err)

		s.Require().Zero(tokenId.Cmp(tops.Tokens[i].TokenId))
		s.Require().True(tops.Tokens[i].Operator.Equal(operator))
		s.T().Logf("\t%s: %s", tokenId, tops.Tokens[i].Operator)
	}
}

func (s *ERC721Suite) TestAddressBalances() {
	contracts := []*types.Address{s.contract, s.contractB}

	ab, err := s.erc721.AddressBalances(s.ctx, s.multicall, s.from, contracts, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(ab.Address.Equal(s.from))
	s.Require().Len(ab.Tokens, len(contracts))

	s.T().Logf("ERC721 %s TestAddressBalances Result", s.from)
	for i, contract := range contracts {
		balance, err := s.erc721.BalanceOf(s.ctx, contract, s.from, rpc.BlockTagLatest)
		s.Require().NoError(err)

		s.Require().True(ab.Tokens[i].Token.Equal(contract))
		s.Require().Zero(balance.Cmp(ab.Tokens[i].Amount))
		s.T().Logf("\t%s: %s", contract, ab.Tokens[i].Amount)
	}
}

func (s *ERC721Suite) TestTokenBalances() {
	addresses := []*types.Address{s.from, s.to}

	tb, err := s.erc721.TokenBalances(s.ctx, s.multicall, s.contract, addresses, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(tb.Contract.Equal(s.contract))
	s.Require().Len(tb.Addresses, len(addresses))

	s.T().Logf("ERC721 %s TestTokenBalances Result", s.contract)
	for i, addr := range addresses {
		balance, err := s.erc721.BalanceOf(s.ctx, s.contract, addr, rpc.BlockTagLatest)
		s.Require().NoError(err)

		s.Require().True(tb.Addresses[i].Address.Equal(addr))
		s.Require().Zero(balance.Cmp(tb.Addresses[i].Amount))
		s.T().Logf("\t%s: %s", addr, tb.Addresses[i].Amount)
	}
}

func (s *ERC721Suite) TestOwnerPairs() {
	contracts := []*types.Address{s.contract, s.contractB}
	tokenIds := []*big.Int{big.NewInt(1), big.NewInt(0)}

	pairs, err := s.erc721.OwnerPairs(s.ctx, s.multicall, contracts, tokenIds, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(pairs, len(contracts))

	s.T().Logf("ERC721 TestOwnerPairs Result")
	for i, contract := range contracts {
		owner, err := s.erc721.OwnerOf(s.ctx, contract, tokenIds[i], rpc.BlockTagLatest)
		s.Require().NoError(err)

		s.Require().True(pairs[i].Contract.Equal(contract))
		s.Require().Zero(tokenIds[i].Cmp(pairs[i].TokenId))
		s.Require().True(pairs[i].Owner.Equal(owner))
		s.T().Logf("\t%s / %s: %s", pairs[i].Contract, pairs[i].TokenId, pairs[i].Owner)
	}
}
