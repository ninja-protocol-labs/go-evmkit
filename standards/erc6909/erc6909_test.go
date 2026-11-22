package erc6909

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/standards/deploy"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/internal/testenv"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc165"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

var interfaceIDERC6909 = [4]byte{0x0f, 0x63, 0x2f, 0xb3}

type ERC6909Suite struct {
	suite.Suite

	RPCURL               string
	PrivateKeyHex        string
	MulticallHex         string
	NicksFactoryHex      string
	NativeSymbol         string
	NativeDecimals       int
	Timeout              time.Duration
	ReceiptPollInterval  time.Duration
	ReceiptTimeout       time.Duration
	SpenderPrivateKeyHex string
	ToHex                string
	ByteCode             string
	Constructor          *abi.Function
	Args                 []any
	Salt                 *types.Hash

	ctx    context.Context
	cancel context.CancelFunc

	client        rpc.Client
	signer        core.Signer
	spenderSigner core.Signer
	deployer      *deploy.NicksDeployer
	from          *types.Address
	contract      *types.Address
	spender       *types.Address
	to            *types.Address
	multicall     *types.Address
	factory       *types.Address

	erc6909 *ERC6909

	id     *big.Int
	amount *big.Int

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestERC6909Suite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)

	suite.Run(t, &ERC6909Suite{
		RPCURL:               c.RPCURL,
		PrivateKeyHex:        c.ActorPrivateKeyHex,
		MulticallHex:         c.Multicall,
		NicksFactoryHex:      c.NicksFactory,
		NativeSymbol:         c.NativeSymbol,
		NativeDecimals:       c.NativeDecimals,
		Timeout:              10 * time.Minute,
		ReceiptPollInterval:  c.ReceiptPollInterval,
		ReceiptTimeout:       c.ReceiptTimeout,
		SpenderPrivateKeyHex: "",
		ToHex:                "",
		ByteCode:             testenv.ERC6909ByteCodeHex,
		Constructor:          abi.NewFunction("constructor", abi.NewTypes(abi.Uint256, abi.String, abi.String, abi.Uint8), nil),
		Args:                 []any{big.NewInt(1), "evmkit test token", "EVK6909", uint8(18)},
		Salt:                 core.Keccak256([]byte("hello")),
	})
}

func (s *ERC6909Suite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(s.T().Context(), s.Timeout)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.erc6909 = New(s.client)

	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	s.id = big.NewInt(1)
	s.amount = big.NewInt(100)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())

	spenderKey, err := types.NewPrivateKeyFromHex(s.SpenderPrivateKeyHex)
	s.Require().NoError(err)
	s.spenderSigner = spenderKey
	s.spender = core.PubkeyToAddress(spenderKey.PublicKey())

	s.to, err = types.NewAddressFromHex(s.ToHex)
	s.Require().NoError(err)

	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)

	s.factory, err = types.NewAddressFromHex(s.NicksFactoryHex)
	s.Require().NoError(err)

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	var salt [32]byte
	copy(salt[:], core.Keccak256(fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())).Bytes())

	deployment := deploy.NewDeterministicContractDeployment(
		deploy.NewContractDeployment(bytecode, s.Constructor.Inputs, s.Args),
		salt,
	)

	unsigned, err := s.deployer.CreateDeployment(s.ctx, tx.NewDynamicFeeBuilder(), s.from, s.factory, big.NewInt(0), deployment)
	s.Require().NoError(err)

	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()

	hash, err := s.deployer.Deploy(s.ctx, unsigned)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "deploy transaction reverted")

	contract, err := s.deployer.DeployAddress(s.factory, deployment)
	s.Require().NoError(err)
	s.contract = contract

	s.Require().Eventually(func() bool {
		var codeHex string
		if err := s.client.Call(s.ctx, rpc.ETHGetCode("code", contract.String(), rpc.BlockTagLatest, &codeHex)); err != nil {
			return false
		}
		code, err := encoding.Hex.Decode(codeHex)
		return err == nil && len(code) > 0
	}, 10*time.Second, 500*time.Millisecond, "no code visible at deployed contract %s", contract)

	s.T().Logf("SetupSuite: deployed contract %s (nonce %d, tx %s)", contract, unsigned.AddressNonce(), hash)

	mintFn := abi.NewFunction("mint", abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256), nil)
	calldata, err := mintFn.EncodeCall(s.from, s.id, s.amount)
	s.Require().NoError(err)

	var unsignedMint core.Transaction
	s.Require().Eventually(func() bool {
		u, err := tx.NewDynamicFeeBuilder().Build(s.from, contract, big.NewInt(0), calldata).Pack(s.ctx, s.client)
		if err != nil || u.Gas() < 40000 {
			return false
		}
		unsignedMint = u
		return true
	}, 10*time.Second, 500*time.Millisecond, "mint gas estimate looked like a no-code call")
	s.Require().NoError(unsignedMint.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedMint.MaxFee())
	s.mu.Unlock()

	raw, err := unsignedMint.EncodeRLP()
	s.Require().NoError(err)
	mintHash, err := tx.Broadcast(s.ctx, s.client, raw)
	s.Require().NoError(err)

	mintReceipt, err := tx.WaitForReceipt(s.ctx, s.client, mintHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(mintReceipt.Status, "mint transaction reverted")

	s.T().Logf("SetupSuite: minted id %s (amount %s) of %s to %s (tx %s)", s.id, s.amount, contract, s.from, mintHash)
}

func (s *ERC6909Suite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(s.T().Context(), s.Timeout)
	defer cancel()

	if s.contract != nil {
		burnFn := abi.NewFunction("burn", abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256), nil)

		for _, holder := range []*types.Address{s.from, s.to} {
			balance, err := s.erc6909.BalanceOf(ctx, nil, NewBalanceOfRequest(s.contract, holder, s.id))
			s.Require().NoError(err)
			if balance.Sign() == 0 {
				continue
			}

			calldata, err := burnFn.EncodeCall(holder, s.id, balance)
			s.Require().NoError(err)

			unsignedBurn, err := tx.NewDynamicFeeBuilder().Build(s.from, s.contract, big.NewInt(0), calldata).Pack(ctx, s.client)
			s.Require().NoError(err)
			s.Require().NoError(unsignedBurn.Sign(s.signer))
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, unsignedBurn.MaxFee())
			s.mu.Unlock()

			raw, err := unsignedBurn.EncodeRLP()
			s.Require().NoError(err)
			burnHash, err := tx.Broadcast(ctx, s.client, raw)
			s.Require().NoError(err)

			burnReceipt, err := tx.WaitForReceipt(ctx, s.client, burnHash, s.ReceiptPollInterval, s.ReceiptTimeout)
			s.Require().NoError(err)
			s.Require().True(burnReceipt.Status, "burn transaction reverted")

			s.T().Logf("TearDownSuite: burned id %s (amount %s) of %s from %s (tx %s)", s.id, balance, s.contract, holder, burnHash)
		}
	}

	if s.totalFee != nil {
		base := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(s.NativeDecimals)), nil)
		whole, frac := new(big.Int), new(big.Int)
		whole.DivMod(s.totalFee, base, frac)

		fracStr := strings.TrimRight(fmt.Sprintf("%0*s", s.NativeDecimals, frac.String()), "0")
		decimal := whole.String()
		if fracStr != "" {
			decimal += "." + fracStr
		}

		s.T().Logf("ERC6909Suite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *ERC6909Suite) Test01ReadMethods() {
	supports, err := s.erc6909.SupportsInterface(s.ctx, nil, erc165.NewSupportsInterfaceRequest(s.contract, interfaceIDERC6909))
	s.Require().NoError(err)
	s.Require().True(supports)

	fromBalance, err := s.erc6909.BalanceOf(s.ctx, nil, NewBalanceOfRequest(s.contract, s.from, s.id))
	s.Require().NoError(err)
	s.Require().Zero(s.amount.Cmp(fromBalance))

	toBalance, err := s.erc6909.BalanceOf(s.ctx, nil, NewBalanceOfRequest(s.contract, s.to, s.id))
	s.Require().NoError(err)
	s.Require().Zero(big.NewInt(0).Cmp(toBalance))

	allowance, err := s.erc6909.Allowance(s.ctx, nil, NewAllowanceRequest(s.contract, s.from, s.to, s.id))
	s.Require().NoError(err)
	s.Require().Zero(big.NewInt(0).Cmp(allowance))

	isOperator, err := s.erc6909.IsOperator(s.ctx, nil, NewIsOperatorRequest(s.contract, s.from, s.to))
	s.Require().NoError(err)
	s.Require().False(isOperator)

	name, err := s.erc6909.Name(s.ctx, nil, NewNameRequest(s.contract, s.id))
	s.Require().NoError(err)
	s.Require().Equal("evmkit test token", name)

	symbol, err := s.erc6909.Symbol(s.ctx, nil, NewSymbolRequest(s.contract, s.id))
	s.Require().NoError(err)
	s.Require().Equal("EVK6909", symbol)

	decimals, err := s.erc6909.Decimals(s.ctx, nil, NewDecimalsRequest(s.contract, s.id))
	s.Require().NoError(err)
	s.Require().Equal(uint8(18), decimals)

	s.T().Logf("ERC6909 %s Test01ReadMethods Result", s.contract)
	s.T().Logf("\tSupports: %v", supports)
	s.T().Logf("\tBalanceOf(id %s)", s.id)
	s.T().Logf("\t\tFrom(%s): %s", s.from, fromBalance)
	s.T().Logf("\t\tTo(%s): %s", s.to, toBalance)
	s.T().Logf("\tAllowance(from, to, id %s): %s", s.id, allowance)
	s.T().Logf("\tIsOperator(from, to): %v", isOperator)
	s.T().Logf("\tName(id %s): %s", s.id, name)
	s.T().Logf("\tSymbol(id %s): %s", s.id, symbol)
	s.T().Logf("\tDecimals(id %s): %d", s.id, decimals)
}

func (s *ERC6909Suite) Test02SetOperator() {
	sentTx, hash, err := s.erc6909.SetOperator(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSetOperatorRequest(s.contract, s.to, true))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transaction reverted")

	operatorSets, err := ExtractOperatorSets(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(operatorSets, 1)
	s.Require().True(operatorSets[0].Contract.Equal(s.contract))
	s.Require().True(operatorSets[0].Owner.Equal(s.from))
	s.Require().True(operatorSets[0].Operator.Equal(s.to))
	s.Require().True(operatorSets[0].Approved)

	var after bool
	s.Require().Eventually(func() bool {
		after, err = s.erc6909.IsOperator(s.ctx, nil, NewIsOperatorRequest(s.contract, s.from, s.to))
		return err == nil && after
	}, 10*time.Second, 500*time.Millisecond, "operator not visible after setOperator")

	s.T().Logf("ERC6909 %s Test02SetOperator Result", s.contract)
	s.T().Logf("\tOwner: %s", s.from)
	s.T().Logf("\tOperator: %s", s.to)
	s.T().Logf("\tTx: %s", hash)
	s.T().Logf("\tGas: %d", sentTx.Gas())
	s.T().Logf("\tNonce: %d", sentTx.AddressNonce())
	s.T().Logf("\tMaxFee: %s", sentTx.MaxFee())
	s.T().Logf("\tIsOperator")
	s.T().Logf("\t\tAfter: %v", after)

	revertTx, revertHash, err := s.erc6909.SetOperator(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSetOperatorRequest(s.contract, s.to, false))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, revertTx.MaxFee())
	s.mu.Unlock()

	revertReceipt, err := tx.WaitForReceipt(s.ctx, s.client, revertHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(revertReceipt.Status, "revert transaction reverted")

	revertOperatorSets, err := ExtractOperatorSets(revertReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(revertOperatorSets, 1)
	s.Require().False(revertOperatorSets[0].Approved)

	var reverted bool
	s.Require().Eventually(func() bool {
		reverted, err = s.erc6909.IsOperator(s.ctx, nil, NewIsOperatorRequest(s.contract, s.from, s.to))
		return err == nil && !reverted
	}, 10*time.Second, 500*time.Millisecond, "operator not reverted to false")

	s.T().Logf("ERC6909 %s Test02SetOperator Revert Result", s.contract)
	s.T().Logf("\tRevertTx: %s", revertHash)
	s.T().Logf("\t\tMaxFee: %s", revertTx.MaxFee())
	s.T().Logf("\t\tReverted: %v", reverted)
}

func (s *ERC6909Suite) Test03Transfer() {
	half := new(big.Int).Div(s.amount, big.NewInt(2))

	sentTx, hash, err := s.erc6909.Transfer(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewTransferRequest(s.contract, s.to, s.id, half))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transaction reverted")

	transfers, err := ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].Contract.Equal(s.contract))
	s.Require().True(transfers[0].Caller.Equal(s.from))
	s.Require().True(transfers[0].Sender.Equal(s.from))
	s.Require().True(transfers[0].Receiver.Equal(s.to))
	s.Require().Zero(s.id.Cmp(transfers[0].Id))
	s.Require().Zero(half.Cmp(transfers[0].Amount))

	var toBalance *big.Int
	s.Require().Eventually(func() bool {
		toBalance, err = s.erc6909.BalanceOf(s.ctx, nil, NewBalanceOfRequest(s.contract, s.to, s.id))
		return err == nil && half.Cmp(toBalance) == 0
	}, 10*time.Second, 500*time.Millisecond, "balance not visible after transfer")

	fromBalance, err := s.erc6909.BalanceOf(s.ctx, nil, NewBalanceOfRequest(s.contract, s.from, s.id))
	s.Require().NoError(err)
	s.Require().Zero(half.Cmp(fromBalance))

	s.T().Logf("ERC6909 %s Test03Transfer Result", s.contract)
	s.T().Logf("\tFrom: %s", s.from)
	s.T().Logf("\tTo: %s", s.to)
	s.T().Logf("\tId: %s", s.id)
	s.T().Logf("\tAmount: %s", half)
	s.T().Logf("\tTx: %s", hash)
	s.T().Logf("\tGas: %d", sentTx.Gas())
	s.T().Logf("\tMaxFee: %s", sentTx.MaxFee())
	s.T().Logf("\tBalanceOf")
	s.T().Logf("\t\tFrom: %s", fromBalance)
	s.T().Logf("\t\tTo: %s", toBalance)
}

func (s *ERC6909Suite) Test04Approve() {
	half := new(big.Int).Div(s.amount, big.NewInt(2))

	sentTx, hash, err := s.erc6909.Approve(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewApproveRequest(s.contract, s.spender, s.id, half))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transaction reverted")

	approvals, err := ExtractApprovals(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(approvals, 1)
	s.Require().True(approvals[0].Contract.Equal(s.contract))
	s.Require().True(approvals[0].Owner.Equal(s.from))
	s.Require().True(approvals[0].Spender.Equal(s.spender))
	s.Require().Zero(s.id.Cmp(approvals[0].Id))
	s.Require().Zero(half.Cmp(approvals[0].Amount))

	var after *big.Int
	s.Require().Eventually(func() bool {
		after, err = s.erc6909.Allowance(s.ctx, nil, NewAllowanceRequest(s.contract, s.from, s.spender, s.id))
		return err == nil && half.Cmp(after) == 0
	}, 10*time.Second, 500*time.Millisecond, "allowance not visible after approve")

	s.T().Logf("ERC6909 %s Test04Approve Result", s.contract)
	s.T().Logf("\tOwner: %s", s.from)
	s.T().Logf("\tSpender: %s", s.spender)
	s.T().Logf("\tId: %s", s.id)
	s.T().Logf("\tTx: %s", hash)
	s.T().Logf("\tGas: %d", sentTx.Gas())
	s.T().Logf("\tNonce: %d", sentTx.AddressNonce())
	s.T().Logf("\tMaxFee: %s", sentTx.MaxFee())
	s.T().Logf("\tAllowance")
	s.T().Logf("\t\tAfter: %s", after)
}

func (s *ERC6909Suite) Test05TransferFrom() {
	half := new(big.Int).Div(s.amount, big.NewInt(2))

	sentTx, hash, err := s.erc6909.TransferFrom(s.ctx, bind.NewTransactOptions(s.spenderSigner, tx.NewDynamicFeeBuilder()), NewTransferFromRequest(s.contract, s.from, s.to, s.id, half))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transaction reverted")

	transfers, err := ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].Contract.Equal(s.contract))
	s.Require().True(transfers[0].Caller.Equal(s.spender))
	s.Require().True(transfers[0].Sender.Equal(s.from))
	s.Require().True(transfers[0].Receiver.Equal(s.to))
	s.Require().Zero(s.id.Cmp(transfers[0].Id))
	s.Require().Zero(half.Cmp(transfers[0].Amount))

	var allowanceAfter *big.Int
	s.Require().Eventually(func() bool {
		allowanceAfter, err = s.erc6909.Allowance(s.ctx, nil, NewAllowanceRequest(s.contract, s.from, s.spender, s.id))
		return err == nil && big.NewInt(0).Cmp(allowanceAfter) == 0
	}, 10*time.Second, 500*time.Millisecond, "allowance not spent after transferFrom")

	fromBalance, err := s.erc6909.BalanceOf(s.ctx, nil, NewBalanceOfRequest(s.contract, s.from, s.id))
	s.Require().NoError(err)
	s.Require().Zero(big.NewInt(0).Cmp(fromBalance))

	toBalance, err := s.erc6909.BalanceOf(s.ctx, nil, NewBalanceOfRequest(s.contract, s.to, s.id))
	s.Require().NoError(err)
	s.Require().Zero(s.amount.Cmp(toBalance))

	s.T().Logf("ERC6909 %s Test05TransferFrom Result", s.contract)
	s.T().Logf("\tSpender: %s", s.spender)
	s.T().Logf("\tFrom: %s", s.from)
	s.T().Logf("\tTo: %s", s.to)
	s.T().Logf("\tId: %s", s.id)
	s.T().Logf("\tAmount: %s", half)
	s.T().Logf("\tTx: %s", hash)
	s.T().Logf("\tGas: %d", sentTx.Gas())
	s.T().Logf("\tMaxFee: %s", sentTx.MaxFee())
	s.T().Logf("\tAllowance")
	s.T().Logf("\t\tAfter: %s", allowanceAfter)
	s.T().Logf("\tBalanceOf")
	s.T().Logf("\t\tFrom: %s", fromBalance)
	s.T().Logf("\t\tTo: %s", toBalance)
}

func (s *ERC6909Suite) Test06TransferInvalidReceiver() {
	signed, hash, err := s.erc6909.Transfer(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewTransferRequest(s.contract, types.ZeroAddress(), s.id, big.NewInt(1)))
	s.Require().Nil(signed)
	s.Require().Nil(hash)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInvalidReceiver)

	var rpcErr *rpc.ResponseError
	s.Require().ErrorAs(err, &rpcErr)
	s.Require().NotEmpty(rpcErr.Data)

	s.T().Logf("ERC6909 %s Test06TransferInvalidReceiver Result", s.contract)
	s.T().Logf("\tError: %v", err)
	s.T().Logf("\tRevertData: %s", rpcErr.Data)
}

func (s *ERC6909Suite) Test07ApproveInvalidSpender() {
	signed, hash, err := s.erc6909.Approve(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewApproveRequest(s.contract, types.ZeroAddress(), s.id, s.amount))
	s.Require().Nil(signed)
	s.Require().Nil(hash)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInvalidSpender)

	var rpcErr *rpc.ResponseError
	s.Require().ErrorAs(err, &rpcErr)
	s.Require().NotEmpty(rpcErr.Data)

	s.T().Logf("ERC6909 %s Test07ApproveInvalidSpender Result", s.contract)
	s.T().Logf("\tError: %v", err)
	s.T().Logf("\tRevertData: %s", rpcErr.Data)
}

func (s *ERC6909Suite) Test08TransferInsufficientBalance() {
	signed, hash, err := s.erc6909.Transfer(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewTransferRequest(s.contract, s.to, s.id, big.NewInt(1)))
	s.Require().Nil(signed)
	s.Require().Nil(hash)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInsufficientBalance)

	var rpcErr *rpc.ResponseError
	s.Require().ErrorAs(err, &rpcErr)
	s.Require().NotEmpty(rpcErr.Data)

	s.T().Logf("ERC6909 %s Test08TransferInsufficientBalance Result", s.contract)
	s.T().Logf("\tError: %v", err)
	s.T().Logf("\tRevertData: %s", rpcErr.Data)
}

func (s *ERC6909Suite) Test09TransferFromInsufficientAllowance() {
	signed, hash, err := s.erc6909.TransferFrom(s.ctx, bind.NewTransactOptions(s.spenderSigner, tx.NewDynamicFeeBuilder()), NewTransferFromRequest(s.contract, s.from, s.to, s.id, big.NewInt(1)))
	s.Require().Nil(signed)
	s.Require().Nil(hash)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInsufficientAllowance)

	var rpcErr *rpc.ResponseError
	s.Require().ErrorAs(err, &rpcErr)
	s.Require().NotEmpty(rpcErr.Data)

	s.T().Logf("ERC6909 %s Test09TransferFromInsufficientAllowance Result", s.contract)
	s.T().Logf("\tError: %v", err)
	s.T().Logf("\tRevertData: %s", rpcErr.Data)
}
