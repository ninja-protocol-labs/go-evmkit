package weth

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/internal/testenv"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

type WETHSuite struct {
	suite.Suite

	RPCURL               string
	PrivateKeyHex        string
	MulticallHex         string
	WETHHex              string
	NativeSymbol         string
	NativeDecimals       int
	Timeout              time.Duration
	ReceiptPollInterval  time.Duration
	ReceiptTimeout       time.Duration
	SpenderPrivateKeyHex string
	Amount               *big.Int

	ctx    context.Context
	cancel context.CancelFunc

	client        rpc.Client
	signer        core.Signer
	spenderSigner core.Signer
	from          *types.Address
	spender       *types.Address
	contract      *types.Address
	multicall     *types.Address

	w *WETH

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestWETHSuite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)
	if c.Addresses.Standard.WETH == "" {
		t.Skip("testenv: no WETH configured for this network")
	}

	suite.Run(t, &WETHSuite{
		RPCURL:               c.RPCURL,
		PrivateKeyHex:        c.ActorPrivateKeyHex,
		MulticallHex:         c.Addresses.Standard.Multicall,
		WETHHex:              c.Addresses.Standard.WETH,
		NativeSymbol:         c.Constant.NativeSymbol,
		NativeDecimals:       c.Constant.NativeDecimals,
		Timeout:              10 * time.Minute,
		ReceiptPollInterval:  c.Constant.ReceiptPollInterval,
		ReceiptTimeout:       c.Constant.ReceiptTimeout,
		SpenderPrivateKeyHex: "",
		Amount:               big.NewInt(1_000_000_000_000),
	})
}

func (s *WETHSuite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(s.T().Context(), s.Timeout)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.w = New(s.client)

	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())

	spenderKey, err := types.NewPrivateKeyFromHex(s.SpenderPrivateKeyHex)
	s.Require().NoError(err)
	s.spenderSigner = spenderKey
	s.spender = core.PubkeyToAddress(spenderKey.PublicKey())

	s.contract, err = types.NewAddressFromHex(s.WETHHex)
	s.Require().NoError(err)

	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)
}

func (s *WETHSuite) TearDownSuite() {
	if s.totalFee != nil {
		base := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(s.NativeDecimals)), nil)
		whole, frac := new(big.Int), new(big.Int)
		whole.DivMod(s.totalFee, base, frac)

		fracStr := strings.TrimRight(fmt.Sprintf("%0*s", s.NativeDecimals, frac.String()), "0")
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
	name, err := s.w.Name(s.ctx, nil, erc20.NewNameRequest(s.contract))
	s.Require().NoError(err)

	symbol, err := s.w.Symbol(s.ctx, nil, erc20.NewSymbolRequest(s.contract))
	s.Require().NoError(err)

	decimals, err := s.w.Decimals(s.ctx, nil, erc20.NewDecimalsRequest(s.contract))
	s.Require().NoError(err)
	s.Require().Equal(uint8(18), decimals)

	totalSupply, err := s.w.TotalSupply(s.ctx, nil, erc20.NewTotalSupplyRequest(s.contract))
	s.Require().NoError(err)

	balances, err := s.w.TokenBalances(s.ctx, nil, erc20.NewTokenBalancesRequest(s.multicall, s.contract, []*types.Address{s.from, s.spender}, false))
	s.Require().NoError(err)

	s.T().Logf("WETH %s Test01ReadMethods Result", s.contract)
	s.T().Logf("\tName: %s", name)
	s.T().Logf("\tSymbol: %s", symbol)
	s.T().Logf("\tDecimals: %d", decimals)
	s.T().Logf("\tTotalSupply: %s", totalSupply)
	for _, ab := range balances.Addresses {
		s.T().Logf("\tBalanceOf(%s): %s", ab.Address, ab.Amount)
	}
}

func (s *WETHSuite) Test02Deposit() {
	depositTx, hash, err := s.w.Deposit(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewDepositRequest(s.contract, s.Amount))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, depositTx.MaxFee())
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "deposit transaction reverted")

	deposits, err := ExtractDeposits(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(deposits, 1)
	s.Require().True(deposits[0].Contract.Equal(s.contract))
	s.Require().True(deposits[0].Dst.Equal(s.from))
	s.Require().Zero(s.Amount.Cmp(deposits[0].Wad))

	before, err := s.w.BalanceOf(s.ctx, bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber-1)), erc20.NewBalanceOfRequest(s.contract, s.from))
	s.Require().NoError(err)
	after, err := s.w.BalanceOf(s.ctx, bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber)), erc20.NewBalanceOfRequest(s.contract, s.from))
	s.Require().NoError(err)
	s.Require().Zero(s.Amount.Cmp(new(big.Int).Sub(after, before)))

	s.T().Logf("WETH %s Test02Deposit Result", s.contract)
	s.T().Logf("\tTx: %s (block %d, value %s)", hash, receipt.BlockNumber, s.Amount)
	s.T().Logf("\tFrom(%s) balance: %s -> %s", s.from, before, after)
}

func (s *WETHSuite) Test03Approve() {
	approveTx, hash, err := s.w.Approve(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), erc20.NewApproveRequest(s.contract, s.spender, s.Amount))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, approveTx.MaxFee())
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "approve transaction reverted")

	allowance, err := s.w.Allowance(s.ctx, bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber)), erc20.NewAllowanceRequest(s.contract, s.from, s.spender))
	s.Require().NoError(err)
	s.Require().Zero(s.Amount.Cmp(allowance))

	s.T().Logf("WETH %s Test03Approve Result", s.contract)
	s.T().Logf("\tTx: %s (block %d)", hash, receipt.BlockNumber)
	s.T().Logf("\tAllowance(%s -> %s): %s", s.from, s.spender, allowance)
}

func (s *WETHSuite) Test04TransferFrom() {
	transferFromTx, hash, err := s.w.TransferFrom(s.ctx, bind.NewTransactOptions(s.spenderSigner, tx.NewDynamicFeeBuilder()), erc20.NewTransferFromRequest(s.contract, s.from, s.spender, s.Amount))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, transferFromTx.MaxFee())
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transferFrom transaction reverted")

	before := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber-1))
	after := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))

	fromBefore, err := s.w.BalanceOf(s.ctx, before, erc20.NewBalanceOfRequest(s.contract, s.from))
	s.Require().NoError(err)
	fromAfter, err := s.w.BalanceOf(s.ctx, after, erc20.NewBalanceOfRequest(s.contract, s.from))
	s.Require().NoError(err)
	spenderBefore, err := s.w.BalanceOf(s.ctx, before, erc20.NewBalanceOfRequest(s.contract, s.spender))
	s.Require().NoError(err)
	spenderAfter, err := s.w.BalanceOf(s.ctx, after, erc20.NewBalanceOfRequest(s.contract, s.spender))
	s.Require().NoError(err)
	allowance, err := s.w.Allowance(s.ctx, after, erc20.NewAllowanceRequest(s.contract, s.from, s.spender))
	s.Require().NoError(err)

	s.Require().Zero(s.Amount.Cmp(new(big.Int).Sub(fromBefore, fromAfter)))
	s.Require().Zero(s.Amount.Cmp(new(big.Int).Sub(spenderAfter, spenderBefore)))
	s.Require().Zero(allowance.Sign())

	s.T().Logf("WETH %s Test04TransferFrom Result", s.contract)
	s.T().Logf("\tTx: %s (block %d)", hash, receipt.BlockNumber)
	s.T().Logf("\tFrom(%s) balance: %s -> %s", s.from, fromBefore, fromAfter)
	s.T().Logf("\tSpender(%s) balance: %s -> %s", s.spender, spenderBefore, spenderAfter)
}

func (s *WETHSuite) Test05Transfer() {
	transferTx, hash, err := s.w.Transfer(s.ctx, bind.NewTransactOptions(s.spenderSigner, tx.NewDynamicFeeBuilder()), erc20.NewTransferRequest(s.contract, s.from, s.Amount))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, transferTx.MaxFee())
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transfer transaction reverted")

	before := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber-1))
	after := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))

	spenderBefore, err := s.w.BalanceOf(s.ctx, before, erc20.NewBalanceOfRequest(s.contract, s.spender))
	s.Require().NoError(err)
	spenderAfter, err := s.w.BalanceOf(s.ctx, after, erc20.NewBalanceOfRequest(s.contract, s.spender))
	s.Require().NoError(err)
	fromBefore, err := s.w.BalanceOf(s.ctx, before, erc20.NewBalanceOfRequest(s.contract, s.from))
	s.Require().NoError(err)
	fromAfter, err := s.w.BalanceOf(s.ctx, after, erc20.NewBalanceOfRequest(s.contract, s.from))
	s.Require().NoError(err)

	s.Require().Zero(s.Amount.Cmp(new(big.Int).Sub(spenderBefore, spenderAfter)))
	s.Require().Zero(s.Amount.Cmp(new(big.Int).Sub(fromAfter, fromBefore)))

	s.T().Logf("WETH %s Test05Transfer Result", s.contract)
	s.T().Logf("\tTx: %s (block %d)", hash, receipt.BlockNumber)
	s.T().Logf("\tSpender(%s) balance: %s -> %s", s.spender, spenderBefore, spenderAfter)
	s.T().Logf("\tFrom(%s) balance: %s -> %s", s.from, fromBefore, fromAfter)
}

func (s *WETHSuite) Test06Withdraw() {
	withdrawTx, hash, err := s.w.Withdraw(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewWithdrawRequest(s.contract, s.Amount))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, withdrawTx.MaxFee())
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "withdraw transaction reverted")

	withdrawals, err := ExtractWithdrawals(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(withdrawals, 1)
	s.Require().True(withdrawals[0].Contract.Equal(s.contract))
	s.Require().True(withdrawals[0].Src.Equal(s.from))
	s.Require().Zero(s.Amount.Cmp(withdrawals[0].Wad))

	before, err := s.w.BalanceOf(s.ctx, bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber-1)), erc20.NewBalanceOfRequest(s.contract, s.from))
	s.Require().NoError(err)
	after, err := s.w.BalanceOf(s.ctx, bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber)), erc20.NewBalanceOfRequest(s.contract, s.from))
	s.Require().NoError(err)
	s.Require().Zero(s.Amount.Cmp(new(big.Int).Sub(before, after)))

	s.T().Logf("WETH %s Test06Withdraw Result", s.contract)
	s.T().Logf("\tTx: %s (block %d)", hash, receipt.BlockNumber)
	s.T().Logf("\tFrom(%s) balance: %s -> %s", s.from, before, after)
}

func (s *WETHSuite) Test07WithdrawInsufficientBalance() {
	balance, err := s.w.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.contract, s.spender))
	s.Require().NoError(err)

	withdrawTx, hash, err := s.w.Withdraw(s.ctx, bind.NewTransactOptions(s.spenderSigner, tx.NewDynamicFeeBuilder()), NewWithdrawRequest(s.contract, new(big.Int).Add(balance, big.NewInt(1))))
	s.Require().Error(err)
	s.Require().Nil(withdrawTx)
	s.Require().Nil(hash)

	s.T().Logf("WETH %s Test07WithdrawInsufficientBalance Result", s.contract)
	s.T().Logf("\tError: %v", err)
}
