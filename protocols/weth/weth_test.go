package weth

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/multicall3"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

type WETHSuite struct {
	suite.Suite

	Enable        bool
	RPCURL        string
	PrivateKeyHex string
	SpenderKeyHex string
	WETHHex       string
	MulticallHex  string
	DepositValue  string
	NativeSymbol  string
	Interval      int
	Timeout       int

	ctx    context.Context
	cancel context.CancelFunc

	client        rpc.Client
	signer        core.Signer
	spenderSigner core.Signer
	from          *types.Address
	spender       *types.Address
	contract      *types.Address
	multicall     *types.Address
	depositValue  *big.Int

	w *WETH

	totalFee *big.Int
}

func TestWETHSuite(t *testing.T) {
	suite.Run(t, &WETHSuite{
		Enable:        false,
		RPCURL:        "",
		PrivateKeyHex: "",
		SpenderKeyHex: "",
		WETHHex:       "",
		MulticallHex:  multicall3.DefaultMulticall3Address.String(),
		DepositValue:  "",
		NativeSymbol:  "",
		Interval:      200,
		Timeout:       60,
	})
}

func (s *WETHSuite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Minute)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.w = New(s.client)

	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())

	spenderKey, err := types.NewPrivateKeyFromHex(s.SpenderKeyHex)
	s.Require().NoError(err)
	s.spenderSigner = spenderKey
	s.spender = core.PubkeyToAddress(spenderKey.PublicKey())

	s.contract, err = types.NewAddressFromHex(s.WETHHex)
	s.Require().NoError(err)

	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)

	depositValue, ok := new(big.Int).SetString(s.DepositValue, 10)
	s.Require().True(ok)
	s.depositValue = depositValue
}

func (s *WETHSuite) TearDownSuite() {
	if s.totalFee != nil {
		base := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
		whole, frac := new(big.Int), new(big.Int)
		whole.DivMod(s.totalFee, base, frac)

		fracStr := strings.TrimRight(fmt.Sprintf("%018s", frac.String()), "0")
		decimal := whole.String()
		if fracStr != "" {
			decimal += "." + fracStr
		}

		s.T().Logf("WETHSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *WETHSuite) Test01ReadMethods() {
	name, err := s.w.Name(s.ctx, s.contract, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("name: %s", name)

	symbol, err := s.w.Symbol(s.ctx, s.contract, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("symbol: %s", symbol)

	decimals, err := s.w.Decimals(s.ctx, s.contract, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Equal(uint8(18), decimals)

	totalSupply, err := s.w.TotalSupply(s.ctx, s.contract, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("totalSupply: %s", totalSupply)

	balances, err := s.w.TokenBalances(s.ctx, s.multicall, s.contract, []*types.Address{s.from, s.spender}, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	for _, ab := range balances.Addresses {
		s.T().Logf("balance of %s: %s", ab.Address, ab.Amount)
	}
}

func (s *WETHSuite) Test02Deposit() {
	unsigned, err := s.w.CreateDeposit(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.from, s.depositValue)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.w.Deposit(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("deposit: broadcast tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "deposit transaction reverted")
	s.T().Logf("deposit: confirmed in block %d", receipt.BlockNumber)

	deposits, err := ExtractDeposits(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(deposits, 1)
	s.Require().True(deposits[0].Dst.Equal(s.from))
	s.Require().Zero(s.depositValue.Cmp(deposits[0].Wad))

	var balance *big.Int
	s.Require().Eventually(func() bool {
		balance, err = s.w.BalanceOf(s.ctx, s.contract, s.from, rpc.BlockTagLatest)
		return err == nil && s.depositValue.Cmp(balance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balance not visible after deposit")
}

func (s *WETHSuite) Test03Approve() {
	unsigned, err := s.w.CreateApprove(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.from, s.spender, s.depositValue)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.w.Approve(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("approve: broadcast tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "approve transaction reverted")
	s.T().Logf("approve: confirmed in block %d", receipt.BlockNumber)

	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.w.Allowance(s.ctx, s.contract, s.from, s.spender, rpc.BlockTagLatest)
		return err == nil && s.depositValue.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
}

func (s *WETHSuite) Test04TransferFrom() {
	unsigned, err := s.w.CreateTransferFrom(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.spender, s.from, s.spender, s.depositValue)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.spenderSigner))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.w.TransferFrom(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("transferFrom: broadcast tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transferFrom transaction reverted")
	s.T().Logf("transferFrom: confirmed in block %d", receipt.BlockNumber)

	var fromBalance, spenderBalance, allowance *big.Int
	s.Require().Eventually(func() bool {
		fromBalance, err = s.w.BalanceOf(s.ctx, s.contract, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		spenderBalance, err = s.w.BalanceOf(s.ctx, s.contract, s.spender, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		allowance, err = s.w.Allowance(s.ctx, s.contract, s.from, s.spender, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		return fromBalance.Sign() == 0 && s.depositValue.Cmp(spenderBalance) == 0 && allowance.Sign() == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances/allowance not visible after transferFrom")
}

func (s *WETHSuite) Test05Transfer() {
	unsigned, err := s.w.CreateTransfer(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.spender, s.from, s.depositValue)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.spenderSigner))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.w.Transfer(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("transfer: broadcast tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transfer transaction reverted")
	s.T().Logf("transfer: confirmed in block %d", receipt.BlockNumber)

	var spenderBalance, fromBalance *big.Int
	s.Require().Eventually(func() bool {
		spenderBalance, err = s.w.BalanceOf(s.ctx, s.contract, s.spender, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		fromBalance, err = s.w.BalanceOf(s.ctx, s.contract, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		return spenderBalance.Sign() == 0 && s.depositValue.Cmp(fromBalance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not visible after transfer")
}

func (s *WETHSuite) Test06Withdraw() {
	unsigned, err := s.w.CreateWithdraw(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.from, s.depositValue)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.w.Withdraw(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("withdraw: broadcast tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "withdraw transaction reverted")
	s.T().Logf("withdraw: confirmed in block %d", receipt.BlockNumber)

	withdrawals, err := ExtractWithdrawals(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(withdrawals, 1)
	s.Require().True(withdrawals[0].Src.Equal(s.from))
	s.Require().Zero(s.depositValue.Cmp(withdrawals[0].Wad))

	var balance *big.Int
	s.Require().Eventually(func() bool {
		balance, err = s.w.BalanceOf(s.ctx, s.contract, s.from, rpc.BlockTagLatest)
		return err == nil && balance.Sign() == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balance not visible after withdraw")
}

func (s *WETHSuite) Test07WithdrawInsufficientBalance() {
	balance, err := s.w.BalanceOf(s.ctx, s.contract, s.spender, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(balance.Sign(), "spender must hold no WETH for this revert case")

	_, err = s.w.CreateWithdraw(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.spender, s.depositValue)
	s.Require().Error(err)
}
