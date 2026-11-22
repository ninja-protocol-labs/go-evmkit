package permit2

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
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/internal/testenv"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/deploy"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

type Permit2Suite struct {
	suite.Suite

	RPCURL               string
	PrivateKeyHex        string
	SpenderPrivateKeyHex string
	Permit2Hex           string
	MulticallHex         string
	NicksFactoryHex      string
	ChainID              int
	NativeSymbol         string
	NativeDecimals       int
	Timeout              time.Duration
	ReceiptPollInterval  time.Duration
	ReceiptTimeout       time.Duration
	ByteCode             string
	Constructor          *abi.Function
	Args                 []any
	Salt                 *types.Hash
	MintValue            *big.Int
	ApproveAmount        *big.Int
	DeadlineSeconds      int
	PriceBuffer          uint64
	LimitBuffer          uint64

	ctx    context.Context
	cancel context.CancelFunc

	client        rpc.Client
	deployer      *deploy.NicksDeployer
	signer        core.Signer
	spenderSigner core.Signer
	from          *types.Address
	spender       *types.Address
	contract      *types.Address
	multicall     *types.Address
	factory       *types.Address
	token         *types.Address
	chainID       *big.Int
	mintValue     *big.Int
	approveAmount *big.Int

	p     *Permit2
	erc20 *erc20.ERC20

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestPermit2Suite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)
	if c.Permit2 == "" {
		t.Skip("testenv: no Permit2 configured for this network")
	}

	suite.Run(t, &Permit2Suite{
		RPCURL:               c.RPCURL,
		PrivateKeyHex:        c.ActorPrivateKeyHex,
		SpenderPrivateKeyHex: "",
		Permit2Hex:           c.Permit2,
		MulticallHex:         c.Multicall,
		NicksFactoryHex:      c.NicksFactory,
		ChainID:              c.ChainID,
		NativeSymbol:         c.NativeSymbol,
		NativeDecimals:       c.NativeDecimals,
		Timeout:              10 * time.Minute,
		ReceiptPollInterval:  c.ReceiptPollInterval,
		ReceiptTimeout:       c.ReceiptTimeout,
		ByteCode:             testenv.ERC20ByteCodeHex,
		Constructor:          abi.NewFunction("constructor", abi.NewTypes(abi.String, abi.String, abi.Uint8), nil),
		Args:                 []any{"evmkit permit2 test ERC20", "EVKP2", uint8(18)},
		Salt:                 core.Keccak256([]byte("go-evmkit-permit2")),
		MintValue:            new(big.Int).Mul(big.NewInt(1_000_000), big.NewInt(1_000_000_000_000_000_000)),
		ApproveAmount:        new(big.Int).Mul(big.NewInt(1_000), big.NewInt(1_000_000_000_000_000_000)),
		DeadlineSeconds:      600,
		PriceBuffer:          150,
		LimitBuffer:          150,
	})
}

func (s *Permit2Suite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(s.T().Context(), s.Timeout)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.p = New(s.client)
	s.erc20 = erc20.New(s.client)

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

	s.contract, err = types.NewAddressFromHex(s.Permit2Hex)
	s.Require().NoError(err)

	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)

	s.factory, err = types.NewAddressFromHex(s.NicksFactoryHex)
	s.Require().NoError(err)

	s.mintValue = s.MintValue
	s.approveAmount = s.ApproveAmount
	s.chainID = big.NewInt(int64(s.ChainID))

	s.T().Logf("SetupSuite: chainID=%s", s.chainID.String())
	s.T().Logf("\tpermit2 = %s", s.contract.String())
	s.T().Logf("\towner = %s", s.from.String())
	s.T().Logf("\tspender = %s", s.spender.String())
	s.T().Logf("\tmulticall = %s", s.multicall.String())

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	var salt [32]byte
	copy(salt[:], core.Keccak256(fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())).Bytes())

	deployment := deploy.NewDeterministicContractDeployment(
		deploy.NewContractDeployment(bytecode, s.Constructor.Inputs, s.Args),
		salt,
	)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	deployTx, hash, err := s.deployer.Deploy(s.ctx, bind.NewTransactOptions(s.signer, builder), deploy.NewCreate2Request(s.factory, deployment))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, deployTx.MaxFee())
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "deploy transaction reverted")

	token, err := s.deployer.DeployAddress(deploy.NewCreate2Request(s.factory, deployment))
	s.Require().NoError(err)
	s.token = token

	s.Require().Eventually(func() bool {
		var codeHex string
		if err := s.client.Call(s.ctx, rpc.ETHGetCode("code", token.String(), rpc.BlockTagLatest, &codeHex)); err != nil {
			return false
		}
		code, err := encoding.Hex.Decode(codeHex)
		return err == nil && len(code) > 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "no code visible at deployed token %s", token)

	s.T().Logf("SetupSuite: deployed test ERC20")
	s.T().Logf("\ttoken = %s", token.String())
	s.T().Logf("\ttx = %s", hash.String())

	mintFn := abi.NewFunction("mint", abi.NewTypes(abi.Address, abi.Uint256), nil)
	calldata, err := mintFn.EncodeCall(s.from, s.mintValue)
	s.Require().NoError(err)

	mintBuilder := tx.NewDynamicFeeBuilder()
	mintBuilder.PriceBuffer(s.PriceBuffer)
	mintBuilder.LimitBuffer(s.LimitBuffer)
	unsignedMint, err := mintBuilder.Build(s.from, s.token, big.NewInt(0), calldata).Pack(s.ctx, s.client)
	s.Require().NoError(err)
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

	s.T().Logf("SetupSuite: minted test ERC20")
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tto = %s", s.from.String())
	s.T().Logf("\tamount = %s", s.mintValue.String())
	s.T().Logf("\ttx = %s", mintHash.String())

	approveBuilder := tx.NewDynamicFeeBuilder()
	approveBuilder.PriceBuffer(s.PriceBuffer)
	approveBuilder.LimitBuffer(s.LimitBuffer)
	approveTx, approveHash, err := s.erc20.Approve(s.ctx, bind.NewTransactOptions(s.signer, approveBuilder), erc20.NewApproveRequest(s.token, s.contract, s.mintValue))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, approveTx.MaxFee())
	s.mu.Unlock()

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")

	s.Require().Eventually(func() bool {
		a, err := s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(s.token, s.from, s.contract))
		return err == nil && a.Cmp(s.mintValue) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "token allowance to permit2 not visible after approve")

	s.T().Logf("SetupSuite: approved permit2 to spend test ERC20")
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\towner = %s", s.from.String())
	s.T().Logf("\tspender = %s", s.contract.String())
	s.T().Logf("\tamount = %s", s.mintValue.String())
	s.T().Logf("\ttx = %s", approveHash.String())
}

func (s *Permit2Suite) TearDownSuite() {
	if s.token != nil {
		balance, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.token, s.from))
		s.Require().NoError(err)
		s.T().Logf("TearDownSuite: remaining balance")
		s.T().Logf("\ttoken = %s", s.token.String())
		s.T().Logf("\towner = %s", s.from.String())
		s.T().Logf("\tbalance = %s", balance.String())
		if balance.Sign() > 0 {
			burnFn := abi.NewFunction("burn", abi.NewTypes(abi.Address, abi.Uint256), nil)
			calldata, err := burnFn.EncodeCall(s.from, balance)
			s.Require().NoError(err)

			builder := tx.NewDynamicFeeBuilder()
			builder.PriceBuffer(s.PriceBuffer)
			builder.LimitBuffer(s.LimitBuffer)
			unsignedBurn, err := builder.Build(s.from, s.token, big.NewInt(0), calldata).Pack(s.ctx, s.client)
			s.Require().NoError(err)
			s.Require().NoError(unsignedBurn.Sign(s.signer))
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, unsignedBurn.MaxFee())
			s.mu.Unlock()

			raw, err := unsignedBurn.EncodeRLP()
			s.Require().NoError(err)
			burnHash, err := tx.Broadcast(s.ctx, s.client, raw)
			s.Require().NoError(err)

			burnReceipt, err := tx.WaitForReceipt(s.ctx, s.client, burnHash, s.ReceiptPollInterval, s.ReceiptTimeout)
			s.Require().NoError(err)
			s.Require().True(burnReceipt.Status, "burn transaction reverted")
			s.T().Logf("TearDownSuite: burned remaining balance")
			s.T().Logf("\ttoken = %s", s.token.String())
			s.T().Logf("\towner = %s", s.from.String())
			s.T().Logf("\tamount = %s", balance.String())
			s.T().Logf("\ttx = %s", burnHash.String())
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

		s.T().Logf("TearDownSuite: total fee spent")
		s.T().Logf("\tamount = %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *Permit2Suite) Test01ReadMethods() {
	s.T().Logf("Test01ReadMethods: fixture")
	s.T().Logf("\tpermit2 = %s", s.contract.String())
	s.T().Logf("\towner = %s", s.from.String())
	s.T().Logf("\tspender = %s", s.spender.String())
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tchainID = %s", s.chainID.String())

	domainSep, err := s.p.DomainSeparator(s.ctx, nil, NewDomainSeparatorRequest(s.contract))
	s.Require().NoError(err)
	s.Require().Len(domainSep.Bytes(), 32)
	s.T().Logf("Test01ReadMethods: DomainSeparator")
	s.T().Logf("\tvalue = 0x%x", domainSep.Bytes())

	allowance, err := s.p.Allowance(s.ctx, nil, NewAllowanceRequest(s.contract, s.from, s.token, s.spender))
	s.Require().NoError(err)
	s.T().Logf("Test01ReadMethods: Allowance")
	s.T().Logf("\towner = %s", s.from.String())
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tspender = %s", s.spender.String())
	s.T().Logf("\tamount = %s", allowance.Amount.String())
	s.T().Logf("\texpiration = %s", allowance.Expiration.String())
	s.T().Logf("\tnonce = %s", allowance.Nonce.String())
	s.Require().Zero(allowance.Amount.Sign())
	s.Require().Zero(allowance.Expiration.Sign())
	s.Require().Zero(allowance.Nonce.Sign())

	wordPos := big.NewInt(0)
	bitmap, err := s.p.NonceBitmap(s.ctx, nil, NewNonceBitmapRequest(s.contract, s.from, wordPos))
	s.Require().NoError(err)
	s.T().Logf("Test01ReadMethods: NonceBitmap")
	s.T().Logf("\towner = %s", s.from.String())
	s.T().Logf("\twordPos = %s", wordPos.String())
	s.T().Logf("\tbitmap = %s", bitmap.String())
	s.Require().Zero(bitmap.Sign())
}

func (s *Permit2Suite) Test02Approve() {
	expiration := big.NewInt(time.Now().Add(1 * time.Hour).Unix())

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.p.Approve(s.ctx, bind.NewTransactOptions(s.signer, builder), NewApproveRequest(s.contract, s.token, s.spender, s.approveAmount, expiration))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()

	s.T().Logf("Test02Approve: broadcast approve")
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tspender = %s", s.spender.String())
	s.T().Logf("\tamount = %s", s.approveAmount.String())
	s.T().Logf("\texpiration = %s", expiration.String())
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "approve transaction reverted")

	approvals, err := ExtractApprovals(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(approvals, 1)
	s.Require().True(approvals[0].Contract.Equal(s.contract))
	s.Require().True(approvals[0].Owner.Equal(s.from))
	s.Require().True(approvals[0].Token.Equal(s.token))
	s.Require().True(approvals[0].Spender.Equal(s.spender))
	s.Require().Zero(s.approveAmount.Cmp(approvals[0].Amount))
	s.Require().Zero(expiration.Cmp(approvals[0].Expiration))

	s.T().Logf("Test02Approve: Approval event")
	s.T().Logf("\towner = %s", approvals[0].Owner.String())
	s.T().Logf("\ttoken = %s", approvals[0].Token.String())
	s.T().Logf("\tspender = %s", approvals[0].Spender.String())
	s.T().Logf("\tamount = %s", approvals[0].Amount.String())
	s.T().Logf("\texpiration = %s", approvals[0].Expiration.String())

	var allowance *PackedAllowance
	s.Require().Eventually(func() bool {
		a, err := s.p.Allowance(s.ctx, nil, NewAllowanceRequest(s.contract, s.from, s.token, s.spender))
		if err != nil || a.Amount.Cmp(s.approveAmount) != 0 || a.Expiration.Cmp(expiration) != 0 {
			return false
		}
		allowance = a
		return true
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance not visible after approve")

	s.T().Logf("Test02Approve: Allowance after approve")
	s.T().Logf("\tamount = %s", allowance.Amount.String())
	s.T().Logf("\texpiration = %s", allowance.Expiration.String())
	s.T().Logf("\tnonce = %s", allowance.Nonce.String())
}

func (s *Permit2Suite) Test03PermitSingle() {
	before, err := s.p.Allowance(s.ctx, nil, NewAllowanceRequest(s.contract, s.from, s.token, s.spender))
	s.Require().NoError(err)

	expiration := big.NewInt(time.Now().Add(1 * time.Hour).Unix())
	sigDeadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
	details := NewPermitDetails(s.token, s.approveAmount, expiration, before.Nonce)
	permitSingle := NewPermitSingle(details, s.spender, sigDeadline)

	domain := NewDomain(s.chainID, s.contract)
	sig, err := permitSingle.Sign(domain, s.signer)
	s.Require().NoError(err)

	s.T().Logf("Test03PermitSingle: signed PermitSingle")
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tspender = %s", s.spender.String())
	s.T().Logf("\tamount = %s", s.approveAmount.String())
	s.T().Logf("\texpiration = %s", expiration.String())
	s.T().Logf("\tnonce = %s", before.Nonce.String())
	s.T().Logf("\tsigDeadline = %s", sigDeadline.String())

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.p.PermitSingle(s.ctx, bind.NewTransactOptions(s.spenderSigner, builder), NewPermitSingleRequest(s.contract, s.from, permitSingle, sig))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test03PermitSingle: broadcast permit")
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "permit transaction reverted")

	permits, err := ExtractPermits(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(permits, 1)
	s.Require().True(permits[0].Contract.Equal(s.contract))
	s.Require().True(permits[0].Owner.Equal(s.from))
	s.Require().True(permits[0].Token.Equal(s.token))
	s.Require().True(permits[0].Spender.Equal(s.spender))
	s.Require().Zero(s.approveAmount.Cmp(permits[0].Amount))
	s.Require().Zero(expiration.Cmp(permits[0].Expiration))
	s.Require().Zero(before.Nonce.Cmp(permits[0].Nonce))

	s.T().Logf("Test03PermitSingle: Permit event")
	s.T().Logf("\towner = %s", permits[0].Owner.String())
	s.T().Logf("\ttoken = %s", permits[0].Token.String())
	s.T().Logf("\tspender = %s", permits[0].Spender.String())
	s.T().Logf("\tamount = %s", permits[0].Amount.String())
	s.T().Logf("\texpiration = %s", permits[0].Expiration.String())
	s.T().Logf("\tnonce = %s", permits[0].Nonce.String())

	wantNonce := new(big.Int).Add(before.Nonce, big.NewInt(1))
	var allowance *PackedAllowance
	s.Require().Eventually(func() bool {
		a, err := s.p.Allowance(s.ctx, nil, NewAllowanceRequest(s.contract, s.from, s.token, s.spender))
		if err != nil || a.Nonce.Cmp(wantNonce) != 0 {
			return false
		}
		allowance = a
		return true
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance nonce not advanced after permit")

	s.T().Logf("Test03PermitSingle: Allowance after permit")
	s.T().Logf("\tamount = %s", allowance.Amount.String())
	s.T().Logf("\texpiration = %s", allowance.Expiration.String())
	s.T().Logf("\tnonce = %s", allowance.Nonce.String())
}

func (s *Permit2Suite) Test04TransferFrom() {
	before, err := s.p.Allowance(s.ctx, nil, NewAllowanceRequest(s.contract, s.from, s.token, s.spender))
	s.Require().NoError(err)

	amount := new(big.Int).Div(before.Amount, big.NewInt(2))

	balancesBefore, err := s.erc20.TokenBalances(s.ctx, nil, erc20.NewTokenBalancesRequest(s.multicall, s.token, []*types.Address{s.from, s.spender}, false))
	s.Require().NoError(err)
	fromBalanceBefore := balancesBefore.Addresses[0].Amount
	toBalanceBefore := balancesBefore.Addresses[1].Amount

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.p.TransferFrom(s.ctx, bind.NewTransactOptions(s.spenderSigner, builder), NewTransferFromRequest(s.contract, s.from, s.spender, amount, s.token))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()

	s.T().Logf("Test04TransferFrom: broadcast transferFrom")
	s.T().Logf("\tfrom = %s", s.from.String())
	s.T().Logf("\tto = %s", s.spender.String())
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tamount = %s", amount.String())
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transferFrom transaction reverted")

	transfers, err := erc20.ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].Contract.Equal(s.token))
	s.Require().True(transfers[0].From.Equal(s.from))
	s.Require().True(transfers[0].To.Equal(s.spender))
	s.Require().Zero(amount.Cmp(transfers[0].Value))

	s.T().Logf("Test04TransferFrom: Transfer event")
	s.T().Logf("\tfrom = %s", transfers[0].From.String())
	s.T().Logf("\tto = %s", transfers[0].To.String())
	s.T().Logf("\tvalue = %s", transfers[0].Value.String())

	wantAllowance := new(big.Int).Sub(before.Amount, amount)
	var allowance *PackedAllowance
	s.Require().Eventually(func() bool {
		a, err := s.p.Allowance(s.ctx, nil, NewAllowanceRequest(s.contract, s.from, s.token, s.spender))
		if err != nil || a.Amount.Cmp(wantAllowance) != 0 {
			return false
		}
		allowance = a
		return true
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance not decremented after transferFrom")

	s.T().Logf("Test04TransferFrom: Allowance after transferFrom")
	s.T().Logf("\tamount = %s", allowance.Amount.String())
	s.T().Logf("\texpiration = %s", allowance.Expiration.String())
	s.T().Logf("\tnonce = %s", allowance.Nonce.String())

	balancesAfter, err := s.erc20.TokenBalances(s.ctx, nil, erc20.NewTokenBalancesRequest(s.multicall, s.token, []*types.Address{s.from, s.spender}, false))
	s.Require().NoError(err)
	fromBalanceAfter := balancesAfter.Addresses[0].Amount
	toBalanceAfter := balancesAfter.Addresses[1].Amount

	s.T().Logf("Test04TransferFrom: balances after transferFrom")
	s.T().Logf("\tfrom = %s", fromBalanceAfter.String())
	s.T().Logf("\tto = %s", toBalanceAfter.String())

	wantFromBalance := new(big.Int).Sub(fromBalanceBefore, amount)
	wantToBalance := new(big.Int).Add(toBalanceBefore, amount)
	s.Require().Zero(wantFromBalance.Cmp(fromBalanceAfter))
	s.Require().Zero(wantToBalance.Cmp(toBalanceAfter))
}

func (s *Permit2Suite) Test05PermitBatch() {
	before, err := s.p.Allowance(s.ctx, nil, NewAllowanceRequest(s.contract, s.from, s.token, s.spender))
	s.Require().NoError(err)

	expiration := big.NewInt(time.Now().Add(1 * time.Hour).Unix())
	sigDeadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
	details := []PermitDetails{*NewPermitDetails(s.token, s.approveAmount, expiration, before.Nonce)}
	permitBatch := NewPermitBatch(details, s.spender, sigDeadline)

	domain := NewDomain(s.chainID, s.contract)
	sig, err := permitBatch.Sign(domain, s.signer)
	s.Require().NoError(err)

	s.T().Logf("Test05PermitBatch: signed PermitBatch")
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tspender = %s", s.spender.String())
	s.T().Logf("\tamount = %s", s.approveAmount.String())
	s.T().Logf("\texpiration = %s", expiration.String())
	s.T().Logf("\tnonce = %s", before.Nonce.String())
	s.T().Logf("\tsigDeadline = %s", sigDeadline.String())

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.p.PermitBatch(s.ctx, bind.NewTransactOptions(s.spenderSigner, builder), NewPermitBatchRequest(s.contract, s.from, permitBatch, sig))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test05PermitBatch: broadcast permit")
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "permit transaction reverted")

	permits, err := ExtractPermits(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(permits, 1)
	s.Require().True(permits[0].Contract.Equal(s.contract))
	s.Require().True(permits[0].Owner.Equal(s.from))
	s.Require().True(permits[0].Token.Equal(s.token))
	s.Require().True(permits[0].Spender.Equal(s.spender))
	s.Require().Zero(s.approveAmount.Cmp(permits[0].Amount))
	s.Require().Zero(expiration.Cmp(permits[0].Expiration))
	s.Require().Zero(before.Nonce.Cmp(permits[0].Nonce))

	s.T().Logf("Test05PermitBatch: Permit event")
	s.T().Logf("\towner = %s", permits[0].Owner.String())
	s.T().Logf("\ttoken = %s", permits[0].Token.String())
	s.T().Logf("\tspender = %s", permits[0].Spender.String())
	s.T().Logf("\tamount = %s", permits[0].Amount.String())
	s.T().Logf("\texpiration = %s", permits[0].Expiration.String())
	s.T().Logf("\tnonce = %s", permits[0].Nonce.String())

	wantNonce := new(big.Int).Add(before.Nonce, big.NewInt(1))
	var allowance *PackedAllowance
	s.Require().Eventually(func() bool {
		a, err := s.p.Allowance(s.ctx, nil, NewAllowanceRequest(s.contract, s.from, s.token, s.spender))
		if err != nil || a.Nonce.Cmp(wantNonce) != 0 {
			return false
		}
		allowance = a
		return true
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance nonce not advanced after permitBatch")

	s.T().Logf("Test05PermitBatch: Allowance after permitBatch")
	s.T().Logf("\tamount = %s", allowance.Amount.String())
	s.T().Logf("\texpiration = %s", allowance.Expiration.String())
	s.T().Logf("\tnonce = %s", allowance.Nonce.String())
}

func (s *Permit2Suite) Test06TransferFromBatch() {
	before, err := s.p.Allowance(s.ctx, nil, NewAllowanceRequest(s.contract, s.from, s.token, s.spender))
	s.Require().NoError(err)

	amount := new(big.Int).Div(before.Amount, big.NewInt(2))
	transferDetails := []AllowanceTransferDetails{*NewAllowanceTransferDetails(s.from, s.spender, amount, s.token)}

	balancesBefore, err := s.erc20.TokenBalances(s.ctx, nil, erc20.NewTokenBalancesRequest(s.multicall, s.token, []*types.Address{s.from, s.spender}, false))
	s.Require().NoError(err)
	fromBalanceBefore := balancesBefore.Addresses[0].Amount
	toBalanceBefore := balancesBefore.Addresses[1].Amount

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.p.TransferFromBatch(s.ctx, bind.NewTransactOptions(s.spenderSigner, builder), NewTransferFromBatchRequest(s.contract, transferDetails))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()

	s.T().Logf("Test06TransferFromBatch: broadcast transferFrom")
	s.T().Logf("\tfrom = %s", s.from.String())
	s.T().Logf("\tto = %s", s.spender.String())
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tamount = %s", amount.String())
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transferFrom(batch) transaction reverted")

	transfers, err := erc20.ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].Contract.Equal(s.token))
	s.Require().True(transfers[0].From.Equal(s.from))
	s.Require().True(transfers[0].To.Equal(s.spender))
	s.Require().Zero(amount.Cmp(transfers[0].Value))

	s.T().Logf("Test06TransferFromBatch: Transfer event")
	s.T().Logf("\tfrom = %s", transfers[0].From.String())
	s.T().Logf("\tto = %s", transfers[0].To.String())
	s.T().Logf("\tvalue = %s", transfers[0].Value.String())

	wantAllowance := new(big.Int).Sub(before.Amount, amount)
	var allowance *PackedAllowance
	s.Require().Eventually(func() bool {
		a, err := s.p.Allowance(s.ctx, nil, NewAllowanceRequest(s.contract, s.from, s.token, s.spender))
		if err != nil || a.Amount.Cmp(wantAllowance) != 0 {
			return false
		}
		allowance = a
		return true
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance not decremented after transferFrom(batch)")

	s.T().Logf("Test06TransferFromBatch: Allowance after transferFrom")
	s.T().Logf("\tamount = %s", allowance.Amount.String())
	s.T().Logf("\texpiration = %s", allowance.Expiration.String())
	s.T().Logf("\tnonce = %s", allowance.Nonce.String())

	balancesAfter, err := s.erc20.TokenBalances(s.ctx, nil, erc20.NewTokenBalancesRequest(s.multicall, s.token, []*types.Address{s.from, s.spender}, false))
	s.Require().NoError(err)
	fromBalanceAfter := balancesAfter.Addresses[0].Amount
	toBalanceAfter := balancesAfter.Addresses[1].Amount

	s.T().Logf("Test06TransferFromBatch: balances after transferFrom")
	s.T().Logf("\tfrom = %s", fromBalanceAfter.String())
	s.T().Logf("\tto = %s", toBalanceAfter.String())

	wantFromBalance := new(big.Int).Sub(fromBalanceBefore, amount)
	wantToBalance := new(big.Int).Add(toBalanceBefore, amount)
	s.Require().Zero(wantFromBalance.Cmp(fromBalanceAfter))
	s.Require().Zero(wantToBalance.Cmp(toBalanceAfter))
}

func (s *Permit2Suite) Test07Lockdown() {
	before, err := s.p.Allowance(s.ctx, nil, NewAllowanceRequest(s.contract, s.from, s.token, s.spender))
	s.Require().NoError(err)
	s.Require().Positive(before.Amount.Sign(), "expected a nonzero allowance to lock down")

	approvals := []TokenSpenderPair{*NewTokenSpenderPair(s.token, s.spender)}

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.p.Lockdown(s.ctx, bind.NewTransactOptions(s.signer, builder), NewLockdownRequest(s.contract, approvals))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()

	s.T().Logf("Test07Lockdown: broadcast lockdown")
	s.T().Logf("\towner = %s", s.from.String())
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tspender = %s", s.spender.String())
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "lockdown transaction reverted")

	lockdowns, err := ExtractLockdowns(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(lockdowns, 1)
	s.Require().True(lockdowns[0].Contract.Equal(s.contract))
	s.Require().True(lockdowns[0].Owner.Equal(s.from))
	s.Require().True(lockdowns[0].Token.Equal(s.token))
	s.Require().True(lockdowns[0].Spender.Equal(s.spender))

	s.T().Logf("Test07Lockdown: Lockdown event")
	s.T().Logf("\towner = %s", lockdowns[0].Owner.String())
	s.T().Logf("\ttoken = %s", lockdowns[0].Token.String())
	s.T().Logf("\tspender = %s", lockdowns[0].Spender.String())

	var allowance *PackedAllowance
	s.Require().Eventually(func() bool {
		a, err := s.p.Allowance(s.ctx, nil, NewAllowanceRequest(s.contract, s.from, s.token, s.spender))
		if err != nil || a.Amount.Sign() != 0 {
			return false
		}
		allowance = a
		return true
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance not zeroed after lockdown")

	s.T().Logf("Test07Lockdown: Allowance after lockdown")
	s.T().Logf("\tamount = %s", allowance.Amount.String())
	s.T().Logf("\texpiration = %s", allowance.Expiration.String())
	s.T().Logf("\tnonce = %s", allowance.Nonce.String())
}

func (s *Permit2Suite) Test08InvalidateNonces() {
	before, err := s.p.Allowance(s.ctx, nil, NewAllowanceRequest(s.contract, s.from, s.token, s.spender))
	s.Require().NoError(err)

	newNonce := new(big.Int).Add(before.Nonce, big.NewInt(5))

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.p.InvalidateNonces(s.ctx, bind.NewTransactOptions(s.signer, builder), NewInvalidateNoncesRequest(s.contract, s.token, s.spender, newNonce))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()

	s.T().Logf("Test08InvalidateNonces: broadcast invalidateNonces")
	s.T().Logf("\towner = %s", s.from.String())
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tspender = %s", s.spender.String())
	s.T().Logf("\toldNonce = %s", before.Nonce.String())
	s.T().Logf("\tnewNonce = %s", newNonce.String())
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "invalidateNonces transaction reverted")

	invalidations, err := ExtractNonceInvalidations(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(invalidations, 1)
	s.Require().True(invalidations[0].Contract.Equal(s.contract))
	s.Require().True(invalidations[0].Owner.Equal(s.from))
	s.Require().True(invalidations[0].Token.Equal(s.token))
	s.Require().True(invalidations[0].Spender.Equal(s.spender))
	s.Require().Zero(newNonce.Cmp(invalidations[0].NewNonce))
	s.Require().Zero(before.Nonce.Cmp(invalidations[0].OldNonce))

	s.T().Logf("Test08InvalidateNonces: NonceInvalidation event")
	s.T().Logf("\towner = %s", invalidations[0].Owner.String())
	s.T().Logf("\ttoken = %s", invalidations[0].Token.String())
	s.T().Logf("\tspender = %s", invalidations[0].Spender.String())
	s.T().Logf("\tnewNonce = %s", invalidations[0].NewNonce.String())
	s.T().Logf("\toldNonce = %s", invalidations[0].OldNonce.String())

	var allowance *PackedAllowance
	s.Require().Eventually(func() bool {
		a, err := s.p.Allowance(s.ctx, nil, NewAllowanceRequest(s.contract, s.from, s.token, s.spender))
		if err != nil || a.Nonce.Cmp(newNonce) != 0 {
			return false
		}
		allowance = a
		return true
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance nonce not advanced after invalidateNonces")

	s.T().Logf("Test08InvalidateNonces: Allowance after invalidateNonces")
	s.T().Logf("\tamount = %s", allowance.Amount.String())
	s.T().Logf("\texpiration = %s", allowance.Expiration.String())
	s.T().Logf("\tnonce = %s", allowance.Nonce.String())

	expiration := big.NewInt(time.Now().Add(1 * time.Hour).Unix())
	sigDeadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
	staleDetails := NewPermitDetails(s.token, s.approveAmount, expiration, before.Nonce)
	stalePermitSingle := NewPermitSingle(staleDetails, s.spender, sigDeadline)

	domain := NewDomain(s.chainID, s.contract)
	staleSig, err := stalePermitSingle.Sign(domain, s.signer)
	s.Require().NoError(err)

	staleBuilder := tx.NewDynamicFeeBuilder()
	staleBuilder.PriceBuffer(s.PriceBuffer)
	staleBuilder.LimitBuffer(s.LimitBuffer)
	revertTx, revertHash, err := s.p.PermitSingle(s.ctx, bind.NewTransactOptions(s.spenderSigner, staleBuilder), NewPermitSingleRequest(s.contract, s.from, stalePermitSingle, staleSig))
	s.Require().Error(err)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)
	s.Require().ErrorIs(err, ErrInvalidNonce)
	s.T().Logf("Test08InvalidateNonces: stale-nonce PermitSingle correctly rejected: %v", err)
}

func (s *Permit2Suite) Test09PermitTransferFrom() {
	requestedAmount := new(big.Int).Div(s.approveAmount, big.NewInt(10))
	nonce := big.NewInt(time.Now().UnixNano())
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	permitted := NewTokenPermissions(s.token, requestedAmount)
	permit := NewPermitTransferFrom(permitted, nonce, deadline)
	transferDetails := NewSignatureTransferDetails(s.spender, requestedAmount)

	domain := NewDomain(s.chainID, s.contract)
	sig, err := permit.Sign(domain, s.spender, s.signer)
	s.Require().NoError(err)

	s.T().Logf("Test09PermitTransferFrom: signed PermitTransferFrom")
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tamount = %s", requestedAmount.String())
	s.T().Logf("\tnonce = %s", nonce.String())
	s.T().Logf("\tdeadline = %s", deadline.String())
	s.T().Logf("\tspender = %s", s.spender.String())

	balancesBefore, err := s.erc20.TokenBalances(s.ctx, nil, erc20.NewTokenBalancesRequest(s.multicall, s.token, []*types.Address{s.from, s.spender}, false))
	s.Require().NoError(err)
	fromBalanceBefore := balancesBefore.Addresses[0].Amount
	toBalanceBefore := balancesBefore.Addresses[1].Amount

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.p.PermitTransferFrom(s.ctx, bind.NewTransactOptions(s.spenderSigner, builder), NewPermitTransferFromRequest(s.contract, permit, transferDetails, s.from, sig))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test09PermitTransferFrom: broadcast permitTransferFrom")
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "permitTransferFrom transaction reverted")

	transfers, err := erc20.ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].Contract.Equal(s.token))
	s.Require().True(transfers[0].From.Equal(s.from))
	s.Require().True(transfers[0].To.Equal(s.spender))
	s.Require().Zero(requestedAmount.Cmp(transfers[0].Value))

	s.T().Logf("Test09PermitTransferFrom: Transfer event")
	s.T().Logf("\tfrom = %s", transfers[0].From.String())
	s.T().Logf("\tto = %s", transfers[0].To.String())
	s.T().Logf("\tvalue = %s", transfers[0].Value.String())

	wantFrom := new(big.Int).Sub(fromBalanceBefore, requestedAmount)
	wantTo := new(big.Int).Add(toBalanceBefore, requestedAmount)
	var balancesAfter *erc20.TokenBalances
	s.Require().Eventually(func() bool {
		b, err := s.erc20.TokenBalances(s.ctx, nil, erc20.NewTokenBalancesRequest(s.multicall, s.token, []*types.Address{s.from, s.spender}, false))
		if err != nil || b.Addresses[0].Amount.Cmp(wantFrom) != 0 || b.Addresses[1].Amount.Cmp(wantTo) != 0 {
			return false
		}
		balancesAfter = b
		return true
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "balances not updated after permitTransferFrom")

	s.T().Logf("Test09PermitTransferFrom: balances after permitTransferFrom")
	s.T().Logf("\tfrom = %s", balancesAfter.Addresses[0].Amount.String())
	s.T().Logf("\tto = %s", balancesAfter.Addresses[1].Amount.String())
}

func (s *Permit2Suite) Test10PermitWitnessTransferFrom() {
	requestedAmount := new(big.Int).Div(s.approveAmount, big.NewInt(10))
	nonce := big.NewInt(time.Now().UnixNano())
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
	witness := types.NewHashFromBytes([]byte("test witness payload, 32 bytes!"))
	witnessTypeString := "Witness witness)Witness(uint256 value)"

	permitted := NewTokenPermissions(s.token, requestedAmount)
	permit := NewPermitTransferFrom(permitted, nonce, deadline)
	transferDetails := NewSignatureTransferDetails(s.spender, requestedAmount)

	domain := NewDomain(s.chainID, s.contract)
	sig, err := permit.SignWithWitness(domain, s.spender, witness, witnessTypeString, s.signer)
	s.Require().NoError(err)

	s.T().Logf("Test10PermitWitnessTransferFrom: signed PermitWitnessTransferFrom")
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tamount = %s", requestedAmount.String())
	s.T().Logf("\tnonce = %s", nonce.String())
	s.T().Logf("\tdeadline = %s", deadline.String())
	s.T().Logf("\tspender = %s", s.spender.String())
	s.T().Logf("\twitness = 0x%x", witness.Bytes())
	s.T().Logf("\twitnessTypeString = %s", witnessTypeString)

	balancesBefore, err := s.erc20.TokenBalances(s.ctx, nil, erc20.NewTokenBalancesRequest(s.multicall, s.token, []*types.Address{s.from, s.spender}, false))
	s.Require().NoError(err)
	fromBalanceBefore := balancesBefore.Addresses[0].Amount
	toBalanceBefore := balancesBefore.Addresses[1].Amount

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.p.PermitWitnessTransferFrom(s.ctx, bind.NewTransactOptions(s.spenderSigner, builder), NewPermitWitnessTransferFromRequest(s.contract, permit, transferDetails, s.from, witness, witnessTypeString, sig))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test10PermitWitnessTransferFrom: broadcast permitWitnessTransferFrom")
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "permitWitnessTransferFrom transaction reverted")

	transfers, err := erc20.ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].Contract.Equal(s.token))
	s.Require().True(transfers[0].From.Equal(s.from))
	s.Require().True(transfers[0].To.Equal(s.spender))
	s.Require().Zero(requestedAmount.Cmp(transfers[0].Value))

	s.T().Logf("Test10PermitWitnessTransferFrom: Transfer event")
	s.T().Logf("\tfrom = %s", transfers[0].From.String())
	s.T().Logf("\tto = %s", transfers[0].To.String())
	s.T().Logf("\tvalue = %s", transfers[0].Value.String())

	wantFrom := new(big.Int).Sub(fromBalanceBefore, requestedAmount)
	wantTo := new(big.Int).Add(toBalanceBefore, requestedAmount)
	var balancesAfter *erc20.TokenBalances
	s.Require().Eventually(func() bool {
		b, err := s.erc20.TokenBalances(s.ctx, nil, erc20.NewTokenBalancesRequest(s.multicall, s.token, []*types.Address{s.from, s.spender}, false))
		if err != nil || b.Addresses[0].Amount.Cmp(wantFrom) != 0 || b.Addresses[1].Amount.Cmp(wantTo) != 0 {
			return false
		}
		balancesAfter = b
		return true
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "balances not updated after permitWitnessTransferFrom")

	s.T().Logf("Test10PermitWitnessTransferFrom: balances after permitWitnessTransferFrom")
	s.T().Logf("\tfrom = %s", balancesAfter.Addresses[0].Amount.String())
	s.T().Logf("\tto = %s", balancesAfter.Addresses[1].Amount.String())
}

func (s *Permit2Suite) Test11PermitTransferFromBatch() {
	requestedAmount := new(big.Int).Div(s.approveAmount, big.NewInt(10))
	nonce := big.NewInt(time.Now().UnixNano())
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	permitted := []TokenPermissions{*NewTokenPermissions(s.token, requestedAmount)}
	permit := NewPermitBatchTransferFrom(permitted, nonce, deadline)
	transferDetails := []SignatureTransferDetails{*NewSignatureTransferDetails(s.spender, requestedAmount)}

	domain := NewDomain(s.chainID, s.contract)
	sig, err := permit.Sign(domain, s.spender, s.signer)
	s.Require().NoError(err)

	s.T().Logf("Test11PermitTransferFromBatch: signed PermitBatchTransferFrom")
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tamount = %s", requestedAmount.String())
	s.T().Logf("\tnonce = %s", nonce.String())
	s.T().Logf("\tdeadline = %s", deadline.String())
	s.T().Logf("\tspender = %s", s.spender.String())

	balancesBefore, err := s.erc20.TokenBalances(s.ctx, nil, erc20.NewTokenBalancesRequest(s.multicall, s.token, []*types.Address{s.from, s.spender}, false))
	s.Require().NoError(err)
	fromBalanceBefore := balancesBefore.Addresses[0].Amount
	toBalanceBefore := balancesBefore.Addresses[1].Amount

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.p.PermitTransferFromBatch(s.ctx, bind.NewTransactOptions(s.spenderSigner, builder), NewPermitTransferFromBatchRequest(s.contract, permit, transferDetails, s.from, sig))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test11PermitTransferFromBatch: broadcast permitTransferFrom(batch)")
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "permitTransferFrom(batch) transaction reverted")

	transfers, err := erc20.ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].Contract.Equal(s.token))
	s.Require().True(transfers[0].From.Equal(s.from))
	s.Require().True(transfers[0].To.Equal(s.spender))
	s.Require().Zero(requestedAmount.Cmp(transfers[0].Value))

	s.T().Logf("Test11PermitTransferFromBatch: Transfer event")
	s.T().Logf("\tfrom = %s", transfers[0].From.String())
	s.T().Logf("\tto = %s", transfers[0].To.String())
	s.T().Logf("\tvalue = %s", transfers[0].Value.String())

	wantFrom := new(big.Int).Sub(fromBalanceBefore, requestedAmount)
	wantTo := new(big.Int).Add(toBalanceBefore, requestedAmount)
	var balancesAfter *erc20.TokenBalances
	s.Require().Eventually(func() bool {
		b, err := s.erc20.TokenBalances(s.ctx, nil, erc20.NewTokenBalancesRequest(s.multicall, s.token, []*types.Address{s.from, s.spender}, false))
		if err != nil || b.Addresses[0].Amount.Cmp(wantFrom) != 0 || b.Addresses[1].Amount.Cmp(wantTo) != 0 {
			return false
		}
		balancesAfter = b
		return true
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "balances not updated after permitTransferFrom(batch)")

	s.T().Logf("Test11PermitTransferFromBatch: balances after permitTransferFrom(batch)")
	s.T().Logf("\tfrom = %s", balancesAfter.Addresses[0].Amount.String())
	s.T().Logf("\tto = %s", balancesAfter.Addresses[1].Amount.String())
}

func (s *Permit2Suite) Test12PermitWitnessTransferFromBatch() {
	requestedAmount := new(big.Int).Div(s.approveAmount, big.NewInt(10))
	nonce := big.NewInt(time.Now().UnixNano())
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
	witness := types.NewHashFromBytes([]byte("test witness payload, 32 bytes!"))
	witnessTypeString := "Witness witness)Witness(uint256 value)"

	permitted := []TokenPermissions{*NewTokenPermissions(s.token, requestedAmount)}
	permit := NewPermitBatchTransferFrom(permitted, nonce, deadline)
	transferDetails := []SignatureTransferDetails{*NewSignatureTransferDetails(s.spender, requestedAmount)}

	domain := NewDomain(s.chainID, s.contract)
	sig, err := permit.SignWithWitness(domain, s.spender, witness, witnessTypeString, s.signer)
	s.Require().NoError(err)

	s.T().Logf("Test12PermitWitnessTransferFromBatch: signed PermitBatchTransferFrom with witness")
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tamount = %s", requestedAmount.String())
	s.T().Logf("\tnonce = %s", nonce.String())
	s.T().Logf("\tdeadline = %s", deadline.String())
	s.T().Logf("\tspender = %s", s.spender.String())
	s.T().Logf("\twitness = 0x%x", witness.Bytes())
	s.T().Logf("\twitnessTypeString = %s", witnessTypeString)

	balancesBefore, err := s.erc20.TokenBalances(s.ctx, nil, erc20.NewTokenBalancesRequest(s.multicall, s.token, []*types.Address{s.from, s.spender}, false))
	s.Require().NoError(err)
	fromBalanceBefore := balancesBefore.Addresses[0].Amount
	toBalanceBefore := balancesBefore.Addresses[1].Amount

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.p.PermitWitnessTransferFromBatch(s.ctx, bind.NewTransactOptions(s.spenderSigner, builder), NewPermitWitnessTransferFromBatchRequest(s.contract, permit, transferDetails, s.from, witness, witnessTypeString, sig))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test12PermitWitnessTransferFromBatch: broadcast permitWitnessTransferFrom(batch)")
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "permitWitnessTransferFrom(batch) transaction reverted")

	transfers, err := erc20.ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].Contract.Equal(s.token))
	s.Require().True(transfers[0].From.Equal(s.from))
	s.Require().True(transfers[0].To.Equal(s.spender))
	s.Require().Zero(requestedAmount.Cmp(transfers[0].Value))

	s.T().Logf("Test12PermitWitnessTransferFromBatch: Transfer event")
	s.T().Logf("\tfrom = %s", transfers[0].From.String())
	s.T().Logf("\tto = %s", transfers[0].To.String())
	s.T().Logf("\tvalue = %s", transfers[0].Value.String())

	wantFrom := new(big.Int).Sub(fromBalanceBefore, requestedAmount)
	wantTo := new(big.Int).Add(toBalanceBefore, requestedAmount)
	var balancesAfter *erc20.TokenBalances
	s.Require().Eventually(func() bool {
		b, err := s.erc20.TokenBalances(s.ctx, nil, erc20.NewTokenBalancesRequest(s.multicall, s.token, []*types.Address{s.from, s.spender}, false))
		if err != nil || b.Addresses[0].Amount.Cmp(wantFrom) != 0 || b.Addresses[1].Amount.Cmp(wantTo) != 0 {
			return false
		}
		balancesAfter = b
		return true
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "balances not updated after permitWitnessTransferFrom(batch)")

	s.T().Logf("Test12PermitWitnessTransferFromBatch: balances after permitWitnessTransferFrom(batch)")
	s.T().Logf("\tfrom = %s", balancesAfter.Addresses[0].Amount.String())
	s.T().Logf("\tto = %s", balancesAfter.Addresses[1].Amount.String())
}

func (s *Permit2Suite) Test13InvalidateUnorderedNonces() {
	wordPos := big.NewInt(time.Now().UnixNano())
	mask := big.NewInt(0b1011)

	before, err := s.p.NonceBitmap(s.ctx, nil, NewNonceBitmapRequest(s.contract, s.from, wordPos))
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.p.InvalidateUnorderedNonces(s.ctx, bind.NewTransactOptions(s.signer, builder), NewInvalidateUnorderedNoncesRequest(s.contract, wordPos, mask))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()

	s.T().Logf("Test13InvalidateUnorderedNonces: broadcast invalidateUnorderedNonces")
	s.T().Logf("\towner = %s", s.from.String())
	s.T().Logf("\twordPos = %s", wordPos.String())
	s.T().Logf("\tmask = %s", mask.String())
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "invalidateUnorderedNonces transaction reverted")

	invalidations, err := ExtractUnorderedNonceInvalidations(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(invalidations, 1)
	s.Require().True(invalidations[0].Contract.Equal(s.contract))
	s.Require().True(invalidations[0].Owner.Equal(s.from))
	s.Require().Zero(wordPos.Cmp(invalidations[0].Word))
	s.Require().Zero(mask.Cmp(invalidations[0].Mask))

	s.T().Logf("Test13InvalidateUnorderedNonces: UnorderedNonceInvalidation event")
	s.T().Logf("\towner = %s", invalidations[0].Owner.String())
	s.T().Logf("\tword = %s", invalidations[0].Word.String())
	s.T().Logf("\tmask = %s", invalidations[0].Mask.String())

	wantBitmap := new(big.Int).Or(before, mask)
	var bitmap *big.Int
	s.Require().Eventually(func() bool {
		b, err := s.p.NonceBitmap(s.ctx, nil, NewNonceBitmapRequest(s.contract, s.from, wordPos))
		if err != nil || b.Cmp(wantBitmap) != 0 {
			return false
		}
		bitmap = b
		return true
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "bitmap not updated after invalidateUnorderedNonces")

	s.T().Logf("Test13InvalidateUnorderedNonces: NonceBitmap after invalidateUnorderedNonces")
	s.T().Logf("\tbitmap = %s", bitmap.String())

	usedNonce := new(big.Int).Lsh(wordPos, 8)

	permitted := NewTokenPermissions(s.token, big.NewInt(1))
	permit := NewPermitTransferFrom(permitted, usedNonce, big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds)*time.Second).Unix()))
	transferDetails := NewSignatureTransferDetails(s.spender, big.NewInt(1))

	domain := NewDomain(s.chainID, s.contract)
	staleSig, err := permit.Sign(domain, s.spender, s.signer)
	s.Require().NoError(err)

	staleBuilder := tx.NewDynamicFeeBuilder()
	staleBuilder.PriceBuffer(s.PriceBuffer)
	staleBuilder.LimitBuffer(s.LimitBuffer)
	revertTx, revertHash, err := s.p.PermitTransferFrom(s.ctx, bind.NewTransactOptions(s.spenderSigner, staleBuilder), NewPermitTransferFromRequest(s.contract, permit, transferDetails, s.from, staleSig))
	s.Require().Error(err)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)
	s.Require().ErrorIs(err, ErrInvalidNonce)
	s.T().Logf("Test13InvalidateUnorderedNonces: reused nonce correctly rejected: %v", err)
}

func (s *Permit2Suite) Test14AllowanceExpired() {
	amount := big.NewInt(1)
	expiration := big.NewInt(time.Now().Add(5 * time.Second).Unix())

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.p.Approve(s.ctx, bind.NewTransactOptions(s.signer, builder), NewApproveRequest(s.contract, s.token, s.spender, amount, expiration))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()

	s.T().Logf("Test14AllowanceExpired: broadcast approve with short expiration")
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tspender = %s", s.spender.String())
	s.T().Logf("\tamount = %s", amount.String())
	s.T().Logf("\texpiration = %s", expiration.String())
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "approve transaction reverted")

	var allowance *PackedAllowance
	s.Require().Eventually(func() bool {
		a, err := s.p.Allowance(s.ctx, nil, NewAllowanceRequest(s.contract, s.from, s.token, s.spender))
		if err != nil || a.Amount.Cmp(amount) != 0 {
			return false
		}
		allowance = a
		return true
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance not visible after approve")

	s.T().Logf("Test14AllowanceExpired: Allowance before expiry")
	s.T().Logf("\tamount = %s", allowance.Amount.String())
	s.T().Logf("\texpiration = %s", allowance.Expiration.String())
	s.T().Logf("\tnonce = %s", allowance.Nonce.String())

	s.T().Logf("Test14AllowanceExpired: waiting for allowance to expire")
	time.Sleep(10 * time.Second)

	expiredBuilder := tx.NewDynamicFeeBuilder()
	expiredBuilder.PriceBuffer(s.PriceBuffer)
	expiredBuilder.LimitBuffer(s.LimitBuffer)
	revertTx, revertHash, err := s.p.TransferFrom(s.ctx, bind.NewTransactOptions(s.spenderSigner, expiredBuilder), NewTransferFromRequest(s.contract, s.from, s.spender, amount, s.token))
	s.Require().Error(err)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)
	s.Require().ErrorIs(err, ErrAllowanceExpired)
	s.T().Logf("Test14AllowanceExpired: expired-allowance transferFrom correctly rejected: %v", err)
}

func (s *Permit2Suite) Test15SignatureExpired() {
	before, err := s.p.Allowance(s.ctx, nil, NewAllowanceRequest(s.contract, s.from, s.token, s.spender))
	s.Require().NoError(err)

	expiration := big.NewInt(time.Now().Add(1 * time.Hour).Unix())
	sigDeadline := big.NewInt(time.Now().Add(-1 * time.Hour).Unix())
	details := NewPermitDetails(s.token, s.approveAmount, expiration, before.Nonce)
	permitSingle := NewPermitSingle(details, s.spender, sigDeadline)

	domain := NewDomain(s.chainID, s.contract)
	sig, err := permitSingle.Sign(domain, s.signer)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	revertTx, revertHash, err := s.p.PermitSingle(s.ctx, bind.NewTransactOptions(s.spenderSigner, builder), NewPermitSingleRequest(s.contract, s.from, permitSingle, sig))
	s.Require().Error(err)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)
	s.Require().ErrorIs(err, ErrSignatureExpired)
	s.T().Logf("Test15SignatureExpired: expired-deadline PermitSingle correctly rejected: %v", err)
}

func (s *Permit2Suite) Test16ExcessiveInvalidation() {
	before, err := s.p.Allowance(s.ctx, nil, NewAllowanceRequest(s.contract, s.from, s.token, s.spender))
	s.Require().NoError(err)

	newNonce := new(big.Int).Add(before.Nonce, big.NewInt(70000))

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	revertTx, revertHash, err := s.p.InvalidateNonces(s.ctx, bind.NewTransactOptions(s.signer, builder), NewInvalidateNoncesRequest(s.contract, s.token, s.spender, newNonce))
	s.Require().Error(err)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)
	s.Require().ErrorIs(err, ErrExcessiveInvalidation)
	s.T().Logf("Test16ExcessiveInvalidation: oversized nonce jump correctly rejected: %v", err)
}

func (s *Permit2Suite) Test17InvalidSignatureLength() {
	before, err := s.p.Allowance(s.ctx, nil, NewAllowanceRequest(s.contract, s.from, s.token, s.spender))
	s.Require().NoError(err)

	expiration := big.NewInt(time.Now().Add(1 * time.Hour).Unix())
	sigDeadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
	details := NewPermitDetails(s.token, s.approveAmount, expiration, before.Nonce)
	permitSingle := NewPermitSingle(details, s.spender, sigDeadline)

	shortSig := make([]byte, 63)
	calldata, err := permitSingleFn.EncodeCall(s.from, permitSingle.tuple(), shortSig)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	revertTx, revertHash, err := bind.Transact(s.ctx, s.client, bind.NewTransactOptions(s.spenderSigner, builder), s.contract, calldata)
	s.Require().Error(err)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	wrapped := s.p.wrapRevertError(err)
	s.Require().ErrorIs(wrapped, ErrInvalidSignatureLength)
	s.T().Logf("Test17InvalidSignatureLength: short signature correctly rejected: %v", wrapped)
}

func (s *Permit2Suite) Test18InvalidSignature() {
	before, err := s.p.Allowance(s.ctx, nil, NewAllowanceRequest(s.contract, s.from, s.token, s.spender))
	s.Require().NoError(err)

	expiration := big.NewInt(time.Now().Add(1 * time.Hour).Unix())
	sigDeadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
	details := NewPermitDetails(s.token, s.approveAmount, expiration, before.Nonce)
	permitSingle := NewPermitSingle(details, s.spender, sigDeadline)

	domain := NewDomain(s.chainID, s.contract)
	sig, err := permitSingle.Sign(domain, s.signer)
	s.Require().NoError(err)

	badSig := sig.LegacyBytes()
	badSig[64] = 17

	calldata, err := permitSingleFn.EncodeCall(s.from, permitSingle.tuple(), badSig)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	revertTx, revertHash, err := bind.Transact(s.ctx, s.client, bind.NewTransactOptions(s.spenderSigner, builder), s.contract, calldata)
	s.Require().Error(err)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	wrapped := s.p.wrapRevertError(err)
	s.Require().ErrorIs(wrapped, ErrInvalidSignature)
	s.T().Logf("Test18InvalidSignature: malformed-v signature correctly rejected: %v", wrapped)
}

func (s *Permit2Suite) Test19InvalidSigner() {
	before, err := s.p.Allowance(s.ctx, nil, NewAllowanceRequest(s.contract, s.from, s.token, s.spender))
	s.Require().NoError(err)

	expiration := big.NewInt(time.Now().Add(1 * time.Hour).Unix())
	sigDeadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
	details := NewPermitDetails(s.token, s.approveAmount, expiration, before.Nonce)
	permitSingle := NewPermitSingle(details, s.spender, sigDeadline)

	domain := NewDomain(s.chainID, s.contract)
	wrongSig, err := permitSingle.Sign(domain, s.spenderSigner)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	revertTx, revertHash, err := s.p.PermitSingle(s.ctx, bind.NewTransactOptions(s.spenderSigner, builder), NewPermitSingleRequest(s.contract, s.from, permitSingle, wrongSig))
	s.Require().Error(err)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)
	s.Require().ErrorIs(err, ErrInvalidSigner)
	s.T().Logf("Test19InvalidSigner: wrong-signer PermitSingle correctly rejected: %v", err)
}

func (s *Permit2Suite) Test20InvalidAmount() {
	permittedAmount := big.NewInt(100)
	requestedAmount := big.NewInt(101)
	nonce := big.NewInt(time.Now().UnixNano())
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	permitted := NewTokenPermissions(s.token, permittedAmount)
	permit := NewPermitTransferFrom(permitted, nonce, deadline)
	transferDetails := NewSignatureTransferDetails(s.spender, requestedAmount)

	domain := NewDomain(s.chainID, s.contract)
	sig, err := permit.Sign(domain, s.spender, s.signer)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	revertTx, revertHash, err := s.p.PermitTransferFrom(s.ctx, bind.NewTransactOptions(s.spenderSigner, builder), NewPermitTransferFromRequest(s.contract, permit, transferDetails, s.from, sig))
	s.Require().Error(err)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)
	s.Require().ErrorIs(err, ErrInvalidAmount)
	s.T().Logf("Test20InvalidAmount: over-requested PermitTransferFrom correctly rejected: %v", err)
}

func (s *Permit2Suite) Test21LengthMismatch() {
	nonce := big.NewInt(time.Now().UnixNano())
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	permitted := []TokenPermissions{*NewTokenPermissions(s.token, big.NewInt(1))}
	permit := NewPermitBatchTransferFrom(permitted, nonce, deadline)
	transferDetails := []SignatureTransferDetails{
		*NewSignatureTransferDetails(s.spender, big.NewInt(1)),
		*NewSignatureTransferDetails(s.spender, big.NewInt(1)),
	}

	domain := NewDomain(s.chainID, s.contract)
	sig, err := permit.Sign(domain, s.spender, s.signer)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	revertTx, revertHash, err := s.p.PermitTransferFromBatch(s.ctx, bind.NewTransactOptions(s.spenderSigner, builder), NewPermitTransferFromBatchRequest(s.contract, permit, transferDetails, s.from, sig))
	s.Require().ErrorIs(err, bind.ErrInvalidRequest)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	calldata, err := EncodePermitTransferFromBatch(permit, transferDetails, s.from, sig)
	s.Require().NoError(err)

	revertTx, revertHash, err = bind.Transact(s.ctx, s.client, bind.NewTransactOptions(s.spenderSigner, builder), s.contract, calldata)
	s.Require().Error(err)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)
	s.Require().ErrorIs(s.p.wrapRevertError(err), ErrLengthMismatch)
	s.T().Logf("Test21LengthMismatch: mismatched transferDetails length correctly rejected: %v", s.p.wrapRevertError(err))
}

func (s *Permit2Suite) Test22InsufficientAllowance() {
	amount := big.NewInt(1)
	expiration := big.NewInt(time.Now().Add(1 * time.Hour).Unix())

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.p.Approve(s.ctx, bind.NewTransactOptions(s.signer, builder), NewApproveRequest(s.contract, s.token, s.spender, amount, expiration))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "approve transaction reverted")

	s.Require().Eventually(func() bool {
		a, err := s.p.Allowance(s.ctx, nil, NewAllowanceRequest(s.contract, s.from, s.token, s.spender))
		return err == nil && a.Amount.Cmp(amount) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance not visible after approve")

	s.T().Logf("Test22InsufficientAllowance: approved small allowance")
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tspender = %s", s.spender.String())
	s.T().Logf("\tamount = %s", amount.String())
	s.T().Logf("\texpiration = %s", expiration.String())

	overAmount := new(big.Int).Add(amount, big.NewInt(1))
	overBuilder := tx.NewDynamicFeeBuilder()
	overBuilder.PriceBuffer(s.PriceBuffer)
	overBuilder.LimitBuffer(s.LimitBuffer)
	revertTx, revertHash, err := s.p.TransferFrom(s.ctx, bind.NewTransactOptions(s.spenderSigner, overBuilder), NewTransferFromRequest(s.contract, s.from, s.spender, overAmount, s.token))
	s.Require().Error(err)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)
	s.Require().ErrorIs(err, ErrInsufficientAllowance)
	s.T().Logf("Test22InsufficientAllowance: over-amount transferFrom correctly rejected: %v", err)
}
