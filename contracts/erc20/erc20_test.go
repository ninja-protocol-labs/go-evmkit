package erc20

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

type ERC20Suite struct {
	suite.Suite

	Enable        bool
	RPCURL        string
	PrivateKeyHex string
	TokenHex      string
	ToHex         string
	MulticallHex  string
	Value         string

	ctx    context.Context
	cancel context.CancelFunc

	client    rpc.Client
	signer    core.Signer
	from      *types.Address
	token     *types.Address
	to        *types.Address
	multicall *types.Address
	value     *big.Int

	erc20 *ERC20
}

func TestERC20LiveSuite(t *testing.T) {
	suite.Run(t, &ERC20Suite{
		Enable:        false,
		RPCURL:        "",
		PrivateKeyHex: "",
		TokenHex:      "",
		ToHex:         "",
		MulticallHex:  "",
		Value:         "",
	})
}

func (s *ERC20Suite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 60*time.Second)

	s.client = rpc.NewClient(s.RPCURL)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())

	s.token, err = types.NewAddressFromHex(s.TokenHex)
	s.Require().NoError(err)

	s.to, err = types.NewAddressFromHex(s.ToHex)
	s.Require().NoError(err)

	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)

	value, ok := new(big.Int).SetString(s.Value, 10)
	s.Require().True(ok)
	s.value = value

	s.erc20 = NewERC20(s.client)
}

func (s *ERC20Suite) TearDownSuite() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *ERC20Suite) TestReadMethods() {
	totalSupply, err := s.erc20.TotalSupply(s.ctx, s.token, rpc.BlockTagLatest)
	s.Require().NoError(err)

	balance, err := s.erc20.BalanceOf(s.ctx, s.token, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)

	allowance, err := s.erc20.Allowance(s.ctx, s.token, s.from, s.to, rpc.BlockTagLatest)
	s.Require().NoError(err)

	name, err := s.erc20.Name(s.ctx, s.token, rpc.BlockTagLatest)
	s.Require().NoError(err)

	symbol, err := s.erc20.Symbol(s.ctx, s.token, rpc.BlockTagLatest)
	s.Require().NoError(err)

	decimals, err := s.erc20.Decimals(s.ctx, s.token, rpc.BlockTagLatest)
	s.Require().NoError(err)

	s.T().Logf("ERC20 %s TestReadMethods Result", s.token)
	s.T().Logf("\tName: %s", name)
	s.T().Logf("\tSymbol: %s", symbol)
	s.T().Logf("\tDecimals: %d", decimals)
	s.T().Logf("\tTotalSupply: %s", totalSupply)
	s.T().Logf("\tBalance: %s", balance)
	s.T().Logf("\tAllowance")
	s.T().Logf("\t\tOwner: %s", s.from)
	s.T().Logf("\t\tSpender: %s", s.to)
	s.T().Logf("\t\tValue: %s", allowance)
}

func (s *ERC20Suite) TestMetadata() {
	m, err := s.erc20.Metadata(s.ctx, s.multicall, s.token, false, rpc.BlockTagLatest)
	s.Require().NoError(err)

	name, err := s.erc20.Name(s.ctx, s.token, rpc.BlockTagLatest)
	s.Require().NoError(err)
	symbol, err := s.erc20.Symbol(s.ctx, s.token, rpc.BlockTagLatest)
	s.Require().NoError(err)
	decimals, err := s.erc20.Decimals(s.ctx, s.token, rpc.BlockTagLatest)
	s.Require().NoError(err)
	totalSupply, err := s.erc20.TotalSupply(s.ctx, s.token, rpc.BlockTagLatest)
	s.Require().NoError(err)

	s.Require().Equal(name, m.Name)
	s.Require().Equal(symbol, m.Symbol)
	s.Require().Equal(decimals, m.Decimals)
	s.Require().Zero(totalSupply.Cmp(m.TotalSupply))

	s.T().Logf("ERC20 %s TestMetadata Result", s.token)
	s.T().Logf("\tName: %s", m.Name)
	s.T().Logf("\tSymbol: %s", m.Symbol)
	s.T().Logf("\tDecimals: %d", m.Decimals)
	s.T().Logf("\tTotalSupply: %s", m.TotalSupply)
}

func (s *ERC20Suite) TestMetadataAllowFailure() {
	m, err := s.erc20.Metadata(s.ctx, s.multicall, s.multicall, true, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Empty(m.Name)
	s.Require().Empty(m.Symbol)
	s.Require().Zero(m.Decimals)
	s.Require().Nil(m.TotalSupply)

	s.T().Logf("ERC20 %s TestMetadataAllowFailure Result", s.multicall)
	s.T().Logf("\tName: %q", m.Name)
	s.T().Logf("\tSymbol: %q", m.Symbol)
	s.T().Logf("\tDecimals: %d", m.Decimals)
	s.T().Logf("\tTotalSupply: %v", m.TotalSupply)
}

func (s *ERC20Suite) TestMetadataWithBalance() {
	mb, err := s.erc20.MetadataWithBalance(s.ctx, s.multicall, s.token, s.from, false, rpc.BlockTagLatest)
	s.Require().NoError(err)

	name, err := s.erc20.Name(s.ctx, s.token, rpc.BlockTagLatest)
	s.Require().NoError(err)
	symbol, err := s.erc20.Symbol(s.ctx, s.token, rpc.BlockTagLatest)
	s.Require().NoError(err)
	decimals, err := s.erc20.Decimals(s.ctx, s.token, rpc.BlockTagLatest)
	s.Require().NoError(err)
	totalSupply, err := s.erc20.TotalSupply(s.ctx, s.token, rpc.BlockTagLatest)
	s.Require().NoError(err)
	balance, err := s.erc20.BalanceOf(s.ctx, s.token, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)

	s.Require().Equal(name, mb.Name)
	s.Require().Equal(symbol, mb.Symbol)
	s.Require().Equal(decimals, mb.Decimals)
	s.Require().Zero(totalSupply.Cmp(mb.TotalSupply))
	s.Require().Zero(balance.Cmp(mb.Balance))

	s.T().Logf("ERC20 %s TestMetadataWithBalance Result", s.token)
	s.T().Logf("\tName: %s", mb.Name)
	s.T().Logf("\tSymbol: %s", mb.Symbol)
	s.T().Logf("\tDecimals: %d", mb.Decimals)
	s.T().Logf("\tTotalSupply: %s", mb.TotalSupply)
	s.T().Logf("\tBalance: %s", mb.Balance)
}

func (s *ERC20Suite) TestMetadataWithBalanceAllowFailure() {
	mb, err := s.erc20.MetadataWithBalance(s.ctx, s.multicall, s.multicall, s.from, true, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Empty(mb.Name)
	s.Require().Empty(mb.Symbol)
	s.Require().Zero(mb.Decimals)
	s.Require().Nil(mb.TotalSupply)
	s.Require().Nil(mb.Balance)

	s.T().Logf("ERC20 %s TestMetadataWithBalanceAllowFailure Result", s.multicall)
	s.T().Logf("\tName: %q", mb.Name)
	s.T().Logf("\tSymbol: %q", mb.Symbol)
	s.T().Logf("\tDecimals: %d", mb.Decimals)
	s.T().Logf("\tTotalSupply: %v", mb.TotalSupply)
	s.T().Logf("\tBalance: %v", mb.Balance)
}

func (s *ERC20Suite) TestAllowanceWithBalance() {
	awb, err := s.erc20.AllowanceWithBalance(s.ctx, s.multicall, s.token, s.from, s.to, false, rpc.BlockTagLatest)
	s.Require().NoError(err)

	allowance, err := s.erc20.Allowance(s.ctx, s.token, s.from, s.to, rpc.BlockTagLatest)
	s.Require().NoError(err)
	ownerBalance, err := s.erc20.BalanceOf(s.ctx, s.token, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	spenderBalance, err := s.erc20.BalanceOf(s.ctx, s.token, s.to, rpc.BlockTagLatest)
	s.Require().NoError(err)

	s.Require().Zero(allowance.Cmp(awb.Value))
	s.Require().Zero(ownerBalance.Cmp(awb.OwnerBalance))
	s.Require().Zero(spenderBalance.Cmp(awb.SpenderBalance))

	s.T().Logf("ERC20 %s TestAllowanceWithBalance Result", s.token)
	s.T().Logf("\tAllowance")
	s.T().Logf("\t\tOwner: %s", s.from)
	s.T().Logf("\t\tSpender: %s", s.to)
	s.T().Logf("\t\tValue: %s", awb.Value)
	s.T().Logf("\tOwnerBalance: %s", awb.OwnerBalance)
	s.T().Logf("\tSpenderBalance: %s", awb.SpenderBalance)
}

func (s *ERC20Suite) TestAllowanceWithBalanceAllowFailure() {
	awb, err := s.erc20.AllowanceWithBalance(s.ctx, s.multicall, s.multicall, s.from, s.to, true, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Nil(awb.Value)
	s.Require().Nil(awb.OwnerBalance)
	s.Require().Nil(awb.SpenderBalance)

	s.T().Logf("ERC20 %s TestAllowanceWithBalanceAllowFailure Result", s.multicall)
	s.T().Logf("\tAllowance")
	s.T().Logf("\t\tOwner: %s", s.from)
	s.T().Logf("\t\tSpender: %s", s.to)
	s.T().Logf("\t\tValue: %v", awb.Value)
	s.T().Logf("\tOwnerBalance: %v", awb.OwnerBalance)
	s.T().Logf("\tSpenderBalance: %v", awb.SpenderBalance)
}

func (s *ERC20Suite) TestTokenBalances() {
	token, err := types.NewAddressFromHex("0xE4aB69C077896252FAFBD49EFD26B5D171A32410") // Base Sepolia - Link
	s.Require().NoError(err)

	addrs := []string{
		"0x833e1D0b8Bc979D49d57b65dCF18364694B16D52",
		"0xc10B4374F9654187DeB5eE2d2715c935f5C1Cb02",
		"0x432311F8926294B3C525634d204a7F1145272282",
		"0xA4e59B140846Cf0a9Be5DE0792eF9bB4a85C378D",
		"0xD6F7d9027cf4F60415A4670041695D1e1705A86d",
	}

	addresses := make([]*types.Address, len(addrs))
	for i, addr := range addrs {
		a, err := types.NewAddressFromHex(addr)
		s.Require().NoError(err)

		addresses[i] = a
	}

	tb, err := s.erc20.TokenBalances(s.ctx, s.multicall, token, addresses, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(tb.Token.Equal(token))
	s.Require().Len(tb.Addresses, len(addresses))

	s.T().Logf("ERC20 %s TestTokenBalances Result", token)
	for i, addr := range addresses {
		balance, err := s.erc20.BalanceOf(s.ctx, token, addr, rpc.BlockTagLatest)
		s.Require().NoError(err)

		s.Require().True(tb.Addresses[i].Address.Equal(addr))
		s.Require().Zero(balance.Cmp(tb.Addresses[i].Amount))
		s.T().Logf("\t%s: %s", addr, tb.Addresses[i].Amount)
	}
}

func (s *ERC20Suite) TestAddressBalances() {
	tks := []string{
		"0xE4aB69C077896252FAFBD49EFD26B5D171A32410", // Base Sepolia - LINK
		"0x808456652fdb597867f38412077A9182bf77359F", // Base Sepolia - EURC
		"0x036CbD53842c5426634e7929541eC2318f3dCF7e", // Base Sepolia - USDC
	}

	tokens := make([]*types.Address, len(tks))
	for i, tk := range tks {
		token, err := types.NewAddressFromHex(tk)
		s.Require().NoError(err)

		tokens[i] = token
	}

	ab, err := s.erc20.AddressBalances(s.ctx, s.multicall, s.from, tokens, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(ab.Address.Equal(s.from))
	s.Require().Len(ab.Tokens, len(tokens))

	s.T().Logf("ERC20 %s TestAddressBalances Result", s.from)
	for i, token := range tokens {
		balance, err := s.erc20.BalanceOf(s.ctx, token, s.from, rpc.BlockTagLatest)
		s.Require().NoError(err)

		s.Require().True(ab.Tokens[i].Token.Equal(token))
		s.Require().Zero(balance.Cmp(ab.Tokens[i].Amount))
		s.T().Logf("\t%s: %s", token, ab.Tokens[i].Amount)
	}
}

func (s *ERC20Suite) TestBalancePairs() {
	tks := []string{
		"0xE4aB69C077896252FAFBD49EFD26B5D171A32410", // Base Sepolia - LINK
		"0x808456652fdb597867f38412077A9182bf77359F", // Base Sepolia - EURC
		"0x036CbD53842c5426634e7929541eC2318f3dCF7e", // Base Sepolia - USDC
	}
	addrs := []string{
		"0x833e1D0b8Bc979D49d57b65dCF18364694B16D52",
		"0xc10B4374F9654187DeB5eE2d2715c935f5C1Cb02",
		"0x432311F8926294B3C525634d204a7F1145272282",
	}

	tokens := make([]*types.Address, len(tks))
	for i, tk := range tks {
		token, err := types.NewAddressFromHex(tk)
		s.Require().NoError(err)

		tokens[i] = token
	}

	addresses := make([]*types.Address, len(addrs))
	for i, addr := range addrs {
		a, err := types.NewAddressFromHex(addr)
		s.Require().NoError(err)

		addresses[i] = a
	}

	pairs, err := s.erc20.BalancePairs(s.ctx, s.multicall, tokens, addresses, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(pairs, len(tokens))

	s.T().Logf("ERC20 TestBalancePairs Result")
	for i, pair := range pairs {
		balance, err := s.erc20.BalanceOf(s.ctx, tokens[i], addresses[i], rpc.BlockTagLatest)
		s.Require().NoError(err)

		s.Require().True(pair.Token.Equal(tokens[i]))
		s.Require().True(pair.Address.Equal(addresses[i]))
		s.Require().Zero(balance.Cmp(pair.Amount))
		s.T().Logf("\t%s / %s: %s", pair.Token, pair.Address, pair.Amount)
	}
}

func (s *ERC20Suite) TestTransfer() {
	before, err := s.erc20.BalanceOf(s.ctx, s.token, s.to, rpc.BlockTagLatest)
	s.Require().NoError(err)

	ok, err := s.erc20.CallTransfer(s.ctx, s.token, s.from, s.to, s.value, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(ok)

	gas, err := s.erc20.TransferGas(s.ctx, tx.NewDynamicFeeBuilder(), s.token, s.from, s.to, s.value)
	s.Require().NoError(err)

	unsigned, err := s.erc20.CreateTransfer(s.ctx, tx.NewDynamicFeeBuilder(), s.token, s.from, s.to, s.value)
	s.Require().NoError(err)

	s.Require().NoError(unsigned.Sign(s.signer))

	hash, err := s.erc20.Transfer(s.ctx, unsigned)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, 0, 0)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transaction reverted")

	transfers, err := ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].Contract.Equal(s.token))
	s.Require().True(transfers[0].From.Equal(s.from))
	s.Require().True(transfers[0].To.Equal(s.to))
	s.Require().Zero(s.value.Cmp(transfers[0].Value))

	want := new(big.Int).Add(before, s.value)
	var after *big.Int
	s.Require().Eventually(func() bool {
		after, err = s.erc20.BalanceOf(s.ctx, s.token, s.to, rpc.BlockTagLatest)
		return err == nil && want.Cmp(after) == 0
	}, 10*time.Second, 500*time.Millisecond, "balance not visible after transfer")

	s.T().Logf("ERC20 %s TestTransfer Result", s.token)
	s.T().Logf("\tFrom: %s", s.from)
	s.T().Logf("\tTo: %s", s.to)
	s.T().Logf("\tValue: %s", s.value)
	s.T().Logf("\tTx: %s", hash)
	s.T().Logf("\tGas: %d", gas)
	s.T().Logf("\tBalance")
	s.T().Logf("\t\tBefore: %s", before)
	s.T().Logf("\t\tAfter: %s", after)
}

func (s *ERC20Suite) TestTransferInsufficientBalance() {
	decimals, err := s.erc20.Decimals(s.ctx, s.token, rpc.BlockTagLatest)
	s.Require().NoError(err)

	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	huge := new(big.Int).Mul(big.NewInt(100_000_000), scale)

	ok, err := s.erc20.CallTransfer(s.ctx, s.token, s.from, s.to, huge, rpc.BlockTagLatest)
	s.Require().Error(err)
	s.Require().False(ok)
	s.Require().ErrorIs(err, ErrUnknown)

	var rpcErr *rpc.ResponseError
	s.Require().ErrorAs(err, &rpcErr)
	s.Require().NotEmpty(rpcErr.Data)

	s.T().Logf("ERC20 %s TestTransferInsufficientBalance Result", s.token)
	s.T().Logf("\tFrom: %s", s.from)
	s.T().Logf("\tTo: %s", s.to)
	s.T().Logf("\tValue: %s", huge)
	s.T().Logf("\tError: %v", err)
	s.T().Logf("\tRevertData: %s", rpcErr.Data)
}

func (s *ERC20Suite) TestApprove() {
	ok, err := s.erc20.CallApprove(s.ctx, s.token, s.from, s.to, s.value, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(ok)

	gas, err := s.erc20.ApproveGas(s.ctx, tx.NewDynamicFeeBuilder(), s.token, s.from, s.to, s.value)
	s.Require().NoError(err)

	unsigned, err := s.erc20.CreateApprove(s.ctx, tx.NewDynamicFeeBuilder(), s.token, s.from, s.to, s.value)
	s.Require().NoError(err)

	s.Require().NoError(unsigned.Sign(s.signer))

	hash, err := s.erc20.Approve(s.ctx, unsigned)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, 0, 0)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transaction reverted")

	approvals, err := ExtractApprovals(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(approvals, 1)
	s.Require().True(approvals[0].Contract.Equal(s.token))
	s.Require().True(approvals[0].Owner.Equal(s.from))
	s.Require().True(approvals[0].Spender.Equal(s.to))
	s.Require().Zero(s.value.Cmp(approvals[0].Value))

	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, s.token, s.from, s.to, rpc.BlockTagLatest)
		return err == nil && s.value.Cmp(allowance) == 0
	}, 10*time.Second, 500*time.Millisecond, "allowance not visible after approve")

	s.T().Logf("ERC20 %s TestApprove Result", s.token)
	s.T().Logf("\tOwner: %s", s.from)
	s.T().Logf("\tSpender: %s", s.to)
	s.T().Logf("\tValue: %s", s.value)
	s.T().Logf("\tTx: %s", hash)
	s.T().Logf("\tGas: %d", gas)
	s.T().Logf("\tAllowance: %s", allowance)
}

func (s *ERC20Suite) TestTransferFrom() {
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, tx.NewDynamicFeeBuilder(), s.token, s.from, s.from, s.value)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, 0, 0)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")

	approvals, err := ExtractApprovals(approveReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(approvals, 1)
	s.Require().Zero(s.value.Cmp(approvals[0].Value))

	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, s.token, s.from, s.from, rpc.BlockTagLatest)
		return err == nil && s.value.Cmp(allowance) == 0
	}, 10*time.Second, 500*time.Millisecond, "allowance not visible after approve")

	ok, err := s.erc20.CallTransferFrom(s.ctx, s.token, s.from, s.from, s.to, s.value, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(ok)

	gas, err := s.erc20.TransferFromGas(s.ctx, tx.NewDynamicFeeBuilder(), s.token, s.from, s.from, s.to, s.value)
	s.Require().NoError(err)

	unsigned, err := s.erc20.CreateTransferFrom(s.ctx, tx.NewDynamicFeeBuilder(), s.token, s.from, s.from, s.to, s.value)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))

	hash, err := s.erc20.TransferFrom(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("transferFrom tx: %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, 0, 0)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transaction reverted")

	transfers, err := ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].Contract.Equal(s.token))
	s.Require().True(transfers[0].From.Equal(s.from))
	s.Require().True(transfers[0].To.Equal(s.to))
	s.Require().Zero(s.value.Cmp(transfers[0].Value))

	s.T().Logf("ERC20 %s TestTransferFrom Result", s.token)
	s.T().Logf("\tApprove")
	s.T().Logf("\t\tOwner: %s", s.from)
	s.T().Logf("\t\tSpender: %s", s.from)
	s.T().Logf("\t\tValue: %s", s.value)
	s.T().Logf("\t\tTx: %s", approveHash)
	s.T().Logf("\t\tGas: %d", gas)
	s.T().Logf("\t\tAllowance: %s", allowance)
	s.T().Logf("\tTransferFrom")
	s.T().Logf("\t\tSender: %s", s.from)
	s.T().Logf("\t\tFrom: %s", s.from)
	s.T().Logf("\t\tTo: %s", s.to)
	s.T().Logf("\t\tValue: %s", s.value)
	s.T().Logf("\t\tTx: %s", hash)
}

func (s *ERC20Suite) TestTransferFromInsufficientAllowance() {
	sender := types.DeadAddress()

	ok, err := s.erc20.CallTransferFrom(s.ctx, s.token, sender, s.from, s.to, s.value, rpc.BlockTagLatest)
	s.Require().Error(err)
	s.Require().False(ok)
	s.Require().ErrorIs(err, ErrUnknown)

	var rpcErr *rpc.ResponseError
	s.Require().ErrorAs(err, &rpcErr)
	s.Require().NotEmpty(rpcErr.Data)

	s.T().Logf("ERC20 %s TestTransferFromInsufficientAllowance Result", s.token)
	s.T().Logf("\tSender: %s", sender)
	s.T().Logf("\tFrom: %s", s.from)
	s.T().Logf("\tTo: %s", s.to)
	s.T().Logf("\tValue: %s", s.value)
	s.T().Logf("\tError: %v", err)
	s.T().Logf("\tRevertData: %s", rpcErr.Data)
}

func (s *ERC20Suite) TestApproveInvalidSpender() {
	spender := types.ZeroAddress()

	ok, err := s.erc20.CallApprove(s.ctx, s.token, s.from, spender, s.value, rpc.BlockTagLatest)
	s.Require().Error(err)
	s.Require().False(ok)
	s.Require().ErrorIs(err, ErrUnknown)

	var rpcErr *rpc.ResponseError
	s.Require().ErrorAs(err, &rpcErr)
	s.Require().NotEmpty(rpcErr.Data)

	s.T().Logf("ERC20 %s TestApproveInvalidSpender Result", s.token)
	s.T().Logf("\tOwner: %s", s.from)
	s.T().Logf("\tSpender: %s", spender)
	s.T().Logf("\tValue: %s", s.value)
	s.T().Logf("\tError: %v", err)
	s.T().Logf("\tRevertData: %s", rpcErr.Data)
}
