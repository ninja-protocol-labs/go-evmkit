package erc2612

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
	"github.com/ninja-protocol-labs/go-evmkit/standards/eip712"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/internal/testenv"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

type ERC2612Suite struct {
	suite.Suite

	RPCURL               string
	PrivateKeyHex        string
	MulticallHex         string
	NicksFactoryHex      string
	NativeSymbol         string
	NativeDecimals       int
	Timeout              time.Duration
	Interval             int
	SpenderPrivateKeyHex string
	ToHex                string
	Value                string
	MintValue            string
	TokenCount           int
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
	spender       *types.Address
	to            *types.Address
	multicall     *types.Address
	factory       *types.Address
	tokens        []*types.Address
	value         *big.Int
	mintValue     *big.Int
	chainID       *big.Int

	erc20   *erc20.ERC20
	erc2612 *ERC2612

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestERC2612Suite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.UnknownNetworkType, testenv.UnknownNetwork)

	suite.Run(t, &ERC2612Suite{
		RPCURL:               c.RPCURL,
		PrivateKeyHex:        c.ActorPrivateKeyHex,
		MulticallHex:         c.Multicall,
		NicksFactoryHex:      c.NicksFactory,
		NativeSymbol:         c.NativeSymbol,
		NativeDecimals:       c.NativeDecimals,
		Timeout:              10 * time.Minute,
		Interval:             200,
		SpenderPrivateKeyHex: "",
		ToHex:                "",
		Value:                "100000",
		MintValue:            "1000000000000000000000000",
		TokenCount:           5,
		ByteCode:             testenv.ERC2612ByteCodeHex,
		Constructor:          abi.NewFunction("constructor", abi.NewTypes(abi.String, abi.String), nil),
		Args:                 []any{"evmkit test ERC2612", "EVK2612"},
		Salt:                 core.Keccak256([]byte("go-evmkit")),
	})
}

func (s *ERC2612Suite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(s.T().Context(), s.Timeout)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.erc20 = erc20.New(s.client)
	s.erc2612 = New(s.client)

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

	s.to, err = types.NewAddressFromHex(s.ToHex)
	s.Require().NoError(err)

	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)

	s.factory, err = types.NewAddressFromHex(s.NicksFactoryHex)
	s.Require().NoError(err)

	value, ok := new(big.Int).SetString(s.Value, 10)
	s.Require().True(ok)
	s.value = value

	mintValue, ok := new(big.Int).SetString(s.MintValue, 10)
	s.Require().True(ok)
	s.mintValue = mintValue

	var chainIDHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHChainID("chainId", &chainIDHex)))
	chainID, ok := new(big.Int).SetString(strings.TrimPrefix(chainIDHex, "0x"), 16)
	s.Require().True(ok)
	s.chainID = chainID

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	mintFn := abi.NewFunction("mint", abi.NewTypes(abi.Address, abi.Uint256), nil)
	runtime := fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())

	s.tokens = make([]*types.Address, s.TokenCount)
	for i := 0; i < s.TokenCount; i++ {
		var salt [32]byte
		copy(salt[:], core.Keccak256(fmt.Appendf(nil, "%s-%d", runtime, i)).Bytes())

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

		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, 200, 15*time.Second)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "deploy transaction reverted")

		token, err := s.deployer.DeployAddress(s.factory, deployment)
		s.Require().NoError(err)
		s.tokens[i] = token

		s.Require().Eventually(func() bool {
			var codeHex string
			if err := s.client.Call(s.ctx, rpc.ETHGetCode("code", token.String(), rpc.BlockTagLatest, &codeHex)); err != nil {
				return false
			}
			code, err := encoding.Hex.Decode(codeHex)
			return err == nil && len(code) > 0
		}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "no code visible at deployed token %s", token)

		s.T().Logf("SetupSuite: deployed token %d/%d %s (nonce %d, tx %s)", i+1, s.TokenCount, token, unsigned.AddressNonce(), hash)

		calldata, err := mintFn.EncodeCall(s.from, s.mintValue)
		s.Require().NoError(err)

		var unsignedMint core.Transaction
		s.Require().Eventually(func() bool {
			u, err := tx.NewDynamicFeeBuilder().Build(s.from, token, big.NewInt(0), calldata).Pack(s.ctx, s.client)
			if err != nil || u.Gas() < 40000 {
				return false
			}
			unsignedMint = u
			return true
		}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "mint gas estimate looked like a no-code call")
		s.Require().NoError(unsignedMint.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, unsignedMint.MaxFee())
		s.mu.Unlock()

		raw, err := unsignedMint.EncodeRLP()
		s.Require().NoError(err)
		mintHash, err := tx.Broadcast(s.ctx, s.client, raw)
		s.Require().NoError(err)

		mintReceipt, err := tx.WaitForReceipt(s.ctx, s.client, mintHash, time.Duration(s.Interval)*time.Millisecond, 15*time.Second)
		s.Require().NoError(err)
		s.Require().True(mintReceipt.Status, "mint transaction reverted")

		s.T().Logf("SetupSuite: minted %s of %s to owner %s (tx %s)", s.mintValue, token, s.from, mintHash)
	}
}

func (s *ERC2612Suite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(s.T().Context(), s.Timeout)
	defer cancel()

	if s.tokens != nil {
		burnFn := abi.NewFunction("burn", abi.NewTypes(abi.Address, abi.Uint256), nil)

		for _, token := range s.tokens {
			for _, holder := range []*types.Address{s.from, s.spender, s.to} {
				balance, err := s.erc20.BalanceOf(ctx, nil, erc20.NewBalanceOfRequest(token, holder))
				s.Require().NoError(err)
				if balance.Sign() == 0 {
					continue
				}

				calldata, err := burnFn.EncodeCall(holder, balance)
				s.Require().NoError(err)

				unsignedBurn, err := tx.NewDynamicFeeBuilder().Build(s.from, token, big.NewInt(0), calldata).Pack(ctx, s.client)
				s.Require().NoError(err)
				s.Require().NoError(unsignedBurn.Sign(s.signer))
				s.mu.Lock()
				s.totalFee.Add(s.totalFee, unsignedBurn.MaxFee())
				s.mu.Unlock()

				raw, err := unsignedBurn.EncodeRLP()
				s.Require().NoError(err)
				burnHash, err := tx.Broadcast(ctx, s.client, raw)
				s.Require().NoError(err)

				burnReceipt, err := tx.WaitForReceipt(ctx, s.client, burnHash, time.Duration(s.Interval)*time.Millisecond, 15*time.Second)
				s.Require().NoError(err)
				s.Require().True(burnReceipt.Status, "burn transaction reverted")

				s.T().Logf("TearDownSuite: burned %s of %s from %s (tx %s)", balance, token, holder, burnHash)
			}
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

		s.T().Logf("ERC2612Suite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *ERC2612Suite) Test01DomainSeparatorMatchesOnChain() {
	for _, token := range s.tokens {
		name, err := s.erc20.Name(s.ctx, nil, erc20.NewNameRequest(token))
		s.Require().NoError(err)

		domain := eip712.NewDomain(name, DefaultVersion, s.chainID, token, nil)
		want, err := eip712.DomainSeparator(domain)
		s.Require().NoError(err)

		got, err := s.erc2612.DomainSeparator(s.ctx, nil, NewDomainSeparatorRequest(token))
		s.Require().NoError(err)

		s.Require().True(want.Equal(got), "want %s, got %s", want, got)

		s.T().Logf("ERC2612 %s Test01DomainSeparatorMatchesOnChain Result", token)
		s.T().Logf("\tName: %s", name)
		s.T().Logf("\tChainID: %s", s.chainID)
		s.T().Logf("\tDomainSeparator: %s", got)
	}
}

func (s *ERC2612Suite) Test02Permit() {
	for _, token := range s.tokens {
		name, err := s.erc20.Name(s.ctx, nil, erc20.NewNameRequest(token))
		s.Require().NoError(err)
		domain := eip712.NewDomain(name, DefaultVersion, s.chainID, token, nil)

		nonceBefore, err := s.erc2612.Nonces(s.ctx, nil, NewNoncesRequest(token, s.from))
		s.Require().NoError(err)

		deadline := big.NewInt(time.Now().Add(1 * time.Hour).Unix())
		permit := NewPermit(s.from, s.spender, s.value, nonceBefore, deadline)
		sig, err := permit.Sign(domain, s.signer)
		s.Require().NoError(err)

		permitTx, permitHash, err := s.erc2612.Permit(s.ctx, bind.NewTransactOptions(s.spenderSigner, tx.NewDynamicFeeBuilder()), NewPermitRequest(token, s.from, s.spender, s.value, deadline, sig))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, permitTx.MaxFee())
		s.mu.Unlock()

		permitReceipt, err := tx.WaitForReceipt(s.ctx, s.client, permitHash, time.Duration(s.Interval)*time.Millisecond, 15*time.Second)
		s.Require().NoError(err)
		s.Require().True(permitReceipt.Status, "permit transaction reverted")

		approvals, err := erc20.ExtractApprovals(permitReceipt.Logs)
		s.Require().NoError(err)
		s.Require().Len(approvals, 1)
		s.Require().True(approvals[0].Contract.Equal(token))
		s.Require().True(approvals[0].Owner.Equal(s.from))
		s.Require().True(approvals[0].Spender.Equal(s.spender))
		s.Require().Zero(s.value.Cmp(approvals[0].Value))

		wantNonceAfter := new(big.Int).Add(nonceBefore, big.NewInt(1))
		var nonceAfter *big.Int
		s.Require().Eventually(func() bool {
			nonceAfter, err = s.erc2612.Nonces(s.ctx, nil, NewNoncesRequest(token, s.from))
			return err == nil && wantNonceAfter.Cmp(nonceAfter) == 0
		}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "nonce not visible after permit")

		var allowance *big.Int
		s.Require().Eventually(func() bool {
			allowance, err = s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(token, s.from, s.spender))
			return err == nil && s.value.Cmp(allowance) == 0
		}, 10*time.Second, 500*time.Millisecond, "allowance not visible after permit")

		s.T().Logf("ERC2612 %s Test02Permit Permit Result", token)
		s.T().Logf("\tOwner: %s", s.from)
		s.T().Logf("\tSpender: %s", s.spender)
		s.T().Logf("\tValue: %s", s.value)
		s.T().Logf("\tNonce: %s -> %s", nonceBefore, nonceAfter)
		s.T().Logf("\tTx: %s", permitHash)

		transferFromTx, transferHash, err := s.erc20.TransferFrom(s.ctx, bind.NewTransactOptions(s.spenderSigner, tx.NewDynamicFeeBuilder()), erc20.NewTransferFromRequest(token, s.from, s.to, s.value))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, transferFromTx.MaxFee())
		s.mu.Unlock()

		transferReceipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHash, time.Duration(s.Interval)*time.Millisecond, 15*time.Second)
		s.Require().NoError(err)
		s.Require().True(transferReceipt.Status, "transferFrom transaction reverted")

		transfers, err := erc20.ExtractTransfers(transferReceipt.Logs)
		s.Require().NoError(err)
		s.Require().Len(transfers, 1)
		s.Require().True(transfers[0].Contract.Equal(token))
		s.Require().True(transfers[0].From.Equal(s.from))
		s.Require().True(transfers[0].To.Equal(s.to))
		s.Require().Zero(s.value.Cmp(transfers[0].Value))

		var allowanceAfter *big.Int
		s.Require().Eventually(func() bool {
			allowanceAfter, err = s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(token, s.from, s.spender))
			return err == nil && allowanceAfter.Sign() == 0
		}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible as consumed after transferFrom")

		var toBalance *big.Int
		s.Require().Eventually(func() bool {
			toBalance, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(token, s.to))
			return err == nil && s.value.Cmp(toBalance) == 0
		}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "balance not visible after transferFrom")

		s.T().Logf("ERC2612 %s Test02Permit TransferFrom Result", token)
		s.T().Logf("\tFrom: %s", s.from)
		s.T().Logf("\tTo: %s", s.to)
		s.T().Logf("\tValue: %s", s.value)
		s.T().Logf("\tTx: %s", transferHash)
		s.T().Logf("\tAllowance remaining: %s", allowanceAfter)
		s.T().Logf("\tTo balance: %s", toBalance)
	}
}

func (s *ERC2612Suite) Test03PermitExpiredDeadline() {
	for _, token := range s.tokens {
		name, err := s.erc20.Name(s.ctx, nil, erc20.NewNameRequest(token))
		s.Require().NoError(err)
		domain := eip712.NewDomain(name, DefaultVersion, s.chainID, token, nil)

		nonce, err := s.erc2612.Nonces(s.ctx, nil, NewNoncesRequest(token, s.from))
		s.Require().NoError(err)

		deadline := big.NewInt(time.Now().Add(-1 * time.Hour).Unix())
		permit := NewPermit(s.from, s.spender, s.value, nonce, deadline)
		sig, err := permit.Sign(domain, s.signer)
		s.Require().NoError(err)

		failedTx, failedHash, err := s.erc2612.Permit(s.ctx, bind.NewTransactOptions(s.spenderSigner, tx.NewDynamicFeeBuilder()), NewPermitRequest(token, s.from, s.spender, s.value, deadline, sig))
		s.Require().Nil(failedTx)
		s.Require().Nil(failedHash)
		s.Require().Error(err)
		s.Require().ErrorIs(err, ErrExpiredSignature)

		s.T().Logf("ERC2612 %s Test03PermitExpiredDeadline Result", token)
		s.T().Logf("\tDeadline: %s", deadline)
		s.T().Logf("\tError: %v", err)
	}
}

func (s *ERC2612Suite) Test04PermitOverwritesAllowance() {
	secondValue := new(big.Int).Quo(s.value, big.NewInt(2))

	for _, token := range s.tokens {
		name, err := s.erc20.Name(s.ctx, nil, erc20.NewNameRequest(token))
		s.Require().NoError(err)
		domain := eip712.NewDomain(name, DefaultVersion, s.chainID, token, nil)

		deadline := big.NewInt(time.Now().Add(1 * time.Hour).Unix())

		firstNonce, err := s.erc2612.Nonces(s.ctx, nil, NewNoncesRequest(token, s.from))
		s.Require().NoError(err)
		firstSig, err := NewPermit(s.from, s.spender, s.value, firstNonce, deadline).Sign(domain, s.signer)
		s.Require().NoError(err)
		firstTx, firstHash, err := s.erc2612.Permit(s.ctx, bind.NewTransactOptions(s.spenderSigner, tx.NewDynamicFeeBuilder()), NewPermitRequest(token, s.from, s.spender, s.value, deadline, firstSig))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, firstTx.MaxFee())
		s.mu.Unlock()
		firstReceipt, err := tx.WaitForReceipt(s.ctx, s.client, firstHash, time.Duration(s.Interval)*time.Millisecond, 15*time.Second)
		s.Require().NoError(err)
		s.Require().True(firstReceipt.Status, "permit transaction reverted")
		var allowance *big.Int
		s.Require().Eventually(func() bool {
			allowance, err = s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(token, s.from, s.spender))
			return err == nil && s.value.Cmp(allowance) == 0
		}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after first permit")

		secondNonce, err := s.erc2612.Nonces(s.ctx, nil, NewNoncesRequest(token, s.from))
		s.Require().NoError(err)
		secondSig, err := NewPermit(s.from, s.spender, secondValue, secondNonce, deadline).Sign(domain, s.signer)
		s.Require().NoError(err)
		secondTx, secondHash, err := s.erc2612.Permit(s.ctx, bind.NewTransactOptions(s.spenderSigner, tx.NewDynamicFeeBuilder()), NewPermitRequest(token, s.from, s.spender, secondValue, deadline, secondSig))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, secondTx.MaxFee())
		s.mu.Unlock()
		secondReceipt, err := tx.WaitForReceipt(s.ctx, s.client, secondHash, time.Duration(s.Interval)*time.Millisecond, 15*time.Second)
		s.Require().NoError(err)
		s.Require().True(secondReceipt.Status, "permit transaction reverted")
		s.Require().Eventually(func() bool {
			allowance, err = s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(token, s.from, s.spender))
			return err == nil && secondValue.Cmp(allowance) == 0
		}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after second permit")

		s.T().Logf("ERC2612 %s Test04PermitOverwritesAllowance Result", token)
		s.T().Logf("\tFirst permit (value %s): %s", s.value, firstHash)
		s.T().Logf("\tSecond permit (value %s): %s", secondValue, secondHash)
		s.T().Logf("\tFinal allowance: %s", allowance)
	}
}

func (s *ERC2612Suite) Test05PermitReplayRejected() {
	for _, token := range s.tokens {
		name, err := s.erc20.Name(s.ctx, nil, erc20.NewNameRequest(token))
		s.Require().NoError(err)
		domain := eip712.NewDomain(name, DefaultVersion, s.chainID, token, nil)

		nonce, err := s.erc2612.Nonces(s.ctx, nil, NewNoncesRequest(token, s.from))
		s.Require().NoError(err)

		deadline := big.NewInt(time.Now().Add(1 * time.Hour).Unix())
		permit := NewPermit(s.from, s.spender, s.value, nonce, deadline)
		sig, err := permit.Sign(domain, s.signer)
		s.Require().NoError(err)

		permitTx, permitHash, err := s.erc2612.Permit(s.ctx, bind.NewTransactOptions(s.spenderSigner, tx.NewDynamicFeeBuilder()), NewPermitRequest(token, s.from, s.spender, s.value, deadline, sig))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, permitTx.MaxFee())
		s.mu.Unlock()

		permitReceipt, err := tx.WaitForReceipt(s.ctx, s.client, permitHash, time.Duration(s.Interval)*time.Millisecond, 15*time.Second)
		s.Require().NoError(err)
		s.Require().True(permitReceipt.Status, "permit transaction reverted")

		var nonceAfter *big.Int
		wantNonceAfter := new(big.Int).Add(nonce, big.NewInt(1))
		s.Require().Eventually(func() bool {
			nonceAfter, err = s.erc2612.Nonces(s.ctx, nil, NewNoncesRequest(token, s.from))
			return err == nil && wantNonceAfter.Cmp(nonceAfter) == 0
		}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "nonce not visible after permit")

		failedTx, failedHash, err := s.erc2612.Permit(s.ctx, bind.NewTransactOptions(s.spenderSigner, tx.NewDynamicFeeBuilder()), NewPermitRequest(token, s.from, s.spender, s.value, deadline, sig))
		s.Require().Nil(failedTx)
		s.Require().Nil(failedHash)
		s.Require().Error(err)
		s.Require().ErrorIs(err, ErrInvalidSigner)

		s.T().Logf("ERC2612 %s Test05PermitReplayRejected Result", token)
		s.T().Logf("\tFirst use Tx: %s", permitHash)
		s.T().Logf("\tReplay error: %v", err)
	}
}

func (s *ERC2612Suite) Test06PermitWrongSigner() {
	for _, token := range s.tokens {
		name, err := s.erc20.Name(s.ctx, nil, erc20.NewNameRequest(token))
		s.Require().NoError(err)
		domain := eip712.NewDomain(name, DefaultVersion, s.chainID, token, nil)

		nonce, err := s.erc2612.Nonces(s.ctx, nil, NewNoncesRequest(token, s.from))
		s.Require().NoError(err)

		deadline := big.NewInt(time.Now().Add(1 * time.Hour).Unix())
		permit := NewPermit(s.from, s.spender, s.value, nonce, deadline)
		sig, err := permit.Sign(domain, s.spenderSigner)
		s.Require().NoError(err)

		failedTx, failedHash, err := s.erc2612.Permit(s.ctx, bind.NewTransactOptions(s.spenderSigner, tx.NewDynamicFeeBuilder()), NewPermitRequest(token, s.from, s.spender, s.value, deadline, sig))
		s.Require().Nil(failedTx)
		s.Require().Nil(failedHash)
		s.Require().Error(err)
		s.Require().ErrorIs(err, ErrInvalidSigner)

		s.T().Logf("ERC2612 %s Test06PermitWrongSigner Result", token)
		s.T().Logf("\tError: %v", err)
	}
}

func (s *ERC2612Suite) Test07ReadMethods() {
	for _, token := range s.tokens {
		nonce, err := s.erc2612.Nonces(s.ctx, nil, NewNoncesRequest(token, s.from))
		s.Require().NoError(err)

		domainSeparator, err := s.erc2612.DomainSeparator(s.ctx, nil, NewDomainSeparatorRequest(token))
		s.Require().NoError(err)

		balances, err := s.erc20.TokenBalances(s.ctx, nil, erc20.NewTokenBalancesRequest(s.multicall, token, []*types.Address{s.from, s.spender, s.to}, false))
		s.Require().NoError(err)

		s.T().Logf("ERC2612 %s Test07ReadMethods Result", token)
		s.T().Logf("\tNonces(%s): %s", s.from, nonce)
		s.T().Logf("\tDomainSeparator: %s", domainSeparator)
		s.T().Logf("\tBalances")
		for _, ab := range balances.Addresses {
			s.T().Logf("\t\t%s: %s", ab.Address, ab.Amount)
		}
	}
}
