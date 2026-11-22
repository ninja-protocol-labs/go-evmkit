package erc20

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
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

type ERC20Suite struct {
	suite.Suite

	RPCURL              string
	PrivateKeyHex       string
	MulticallHex        string
	NicksFactoryHex     string
	NativeSymbol        string
	NativeDecimals      int
	Timeout             time.Duration
	ReceiptPollInterval time.Duration
	ReceiptTimeout      time.Duration
	ToHex               string
	Value               string
	MintValue           string
	TokenCount          int
	ByteCode            string
	Constructor         *abi.Function
	Args                []any
	Salt                *types.Hash

	ctx    context.Context
	cancel context.CancelFunc

	client    rpc.Client
	signer    core.Signer
	deployer  *deploy.NicksDeployer
	from      *types.Address
	tokens    []*types.Address
	to        *types.Address
	multicall *types.Address
	factory   *types.Address
	value     *big.Int
	mintValue *big.Int

	erc20 *ERC20

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestERC20Suite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)

	suite.Run(t, &ERC20Suite{
		RPCURL:              c.RPCURL,
		PrivateKeyHex:       c.ActorPrivateKeyHex,
		MulticallHex:        c.Multicall,
		NicksFactoryHex:     c.NicksFactory,
		NativeSymbol:        c.NativeSymbol,
		NativeDecimals:      c.NativeDecimals,
		Timeout:             10 * time.Minute,
		ReceiptPollInterval: c.ReceiptPollInterval,
		ReceiptTimeout:      c.ReceiptTimeout,
		ToHex:               "",
		Value:               "100000",
		MintValue:           "1000000000000000000000000",
		TokenCount:          5,
		ByteCode:            testenv.ERC20ByteCodeHex,
		Constructor:         abi.NewFunction("constructor", abi.NewTypes(abi.String, abi.String, abi.Uint8), nil),
		Args:                []any{"evmkit test ERC20", "EVK20", uint8(18)},
		Salt:                core.Keccak256([]byte("go-evmkit")),
	})
}

func (s *ERC20Suite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(s.T().Context(), s.Timeout)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.erc20 = New(s.client)

	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())

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

		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
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
		}, 10*time.Second, 500*time.Millisecond, "no code visible at deployed token %s", token)

		s.T().Logf("SetupSuite: deployed token %d/%d %s (nonce %d, tx %s)", i+1, s.TokenCount, token, unsigned.AddressNonce(), hash)

		for _, recipient := range []*types.Address{s.from, s.to} {
			calldata, err := mintFn.EncodeCall(recipient, s.mintValue)
			s.Require().NoError(err)

			var unsignedMint core.Transaction
			s.Require().Eventually(func() bool {
				u, err := tx.NewDynamicFeeBuilder().Build(s.from, token, big.NewInt(0), calldata).Pack(s.ctx, s.client)
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

			s.T().Logf("SetupSuite: minted %s of %s to %s (tx %s)", s.mintValue, token, recipient, mintHash)
		}
	}
}

func (s *ERC20Suite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(s.T().Context(), s.Timeout)
	defer cancel()

	if s.tokens != nil {
		burnFn := abi.NewFunction("burn", abi.NewTypes(abi.Address, abi.Uint256), nil)

		for _, token := range s.tokens {
			for _, holder := range []*types.Address{s.from, s.to} {
				balance, err := s.erc20.BalanceOf(ctx, nil, NewBalanceOfRequest(token, holder))
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

				burnReceipt, err := tx.WaitForReceipt(ctx, s.client, burnHash, s.ReceiptPollInterval, s.ReceiptTimeout)
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

		s.T().Logf("ERC20Suite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *ERC20Suite) Test01ReadMethods() {
	for _, token := range s.tokens {
		totalSupply, err := s.erc20.TotalSupply(s.ctx, nil, NewTotalSupplyRequest(token))
		s.Require().NoError(err)

		balance, err := s.erc20.BalanceOf(s.ctx, nil, NewBalanceOfRequest(token, s.from))
		s.Require().NoError(err)

		allowance, err := s.erc20.Allowance(s.ctx, nil, NewAllowanceRequest(token, s.from, s.to))
		s.Require().NoError(err)

		name, err := s.erc20.Name(s.ctx, nil, NewNameRequest(token))
		s.Require().NoError(err)

		symbol, err := s.erc20.Symbol(s.ctx, nil, NewSymbolRequest(token))
		s.Require().NoError(err)

		decimals, err := s.erc20.Decimals(s.ctx, nil, NewDecimalsRequest(token))
		s.Require().NoError(err)

		s.T().Logf("ERC20 %s Test01ReadMethods Result", token)
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
}

func (s *ERC20Suite) Test02Metadata() {
	for _, token := range s.tokens {
		m, err := s.erc20.Metadata(s.ctx, nil, NewMetadataRequest(s.multicall, token, false))
		s.Require().NoError(err)

		name, err := s.erc20.Name(s.ctx, nil, NewNameRequest(token))
		s.Require().NoError(err)
		symbol, err := s.erc20.Symbol(s.ctx, nil, NewSymbolRequest(token))
		s.Require().NoError(err)
		decimals, err := s.erc20.Decimals(s.ctx, nil, NewDecimalsRequest(token))
		s.Require().NoError(err)
		totalSupply, err := s.erc20.TotalSupply(s.ctx, nil, NewTotalSupplyRequest(token))
		s.Require().NoError(err)

		s.Require().Equal(name, m.Name)
		s.Require().Equal(symbol, m.Symbol)
		s.Require().Equal(decimals, m.Decimals)
		s.Require().Zero(totalSupply.Cmp(m.TotalSupply))

		s.T().Logf("ERC20 %s Test02Metadata Result", token)
		s.T().Logf("\tName: %s", m.Name)
		s.T().Logf("\tSymbol: %s", m.Symbol)
		s.T().Logf("\tDecimals: %d", m.Decimals)
		s.T().Logf("\tTotalSupply: %s", m.TotalSupply)
	}
}

func (s *ERC20Suite) Test03MetadataAllowFailure() {
	m, err := s.erc20.Metadata(s.ctx, nil, NewMetadataRequest(s.multicall, s.multicall, true))
	s.Require().NoError(err)
	s.Require().Empty(m.Name)
	s.Require().Empty(m.Symbol)
	s.Require().Zero(m.Decimals)
	s.Require().Nil(m.TotalSupply)

	s.T().Logf("ERC20 %s Test03MetadataAllowFailure Result", s.multicall)
	s.T().Logf("\tName: %q", m.Name)
	s.T().Logf("\tSymbol: %q", m.Symbol)
	s.T().Logf("\tDecimals: %d", m.Decimals)
	s.T().Logf("\tTotalSupply: %v", m.TotalSupply)
}

func (s *ERC20Suite) Test04MetadataWithBalance() {
	for _, token := range s.tokens {
		mb, err := s.erc20.MetadataWithBalance(s.ctx, nil, NewMetadataWithBalanceRequest(s.multicall, token, s.from, false))
		s.Require().NoError(err)

		name, err := s.erc20.Name(s.ctx, nil, NewNameRequest(token))
		s.Require().NoError(err)
		symbol, err := s.erc20.Symbol(s.ctx, nil, NewSymbolRequest(token))
		s.Require().NoError(err)
		decimals, err := s.erc20.Decimals(s.ctx, nil, NewDecimalsRequest(token))
		s.Require().NoError(err)
		totalSupply, err := s.erc20.TotalSupply(s.ctx, nil, NewTotalSupplyRequest(token))
		s.Require().NoError(err)
		balance, err := s.erc20.BalanceOf(s.ctx, nil, NewBalanceOfRequest(token, s.from))
		s.Require().NoError(err)

		s.Require().Equal(name, mb.Name)
		s.Require().Equal(symbol, mb.Symbol)
		s.Require().Equal(decimals, mb.Decimals)
		s.Require().Zero(totalSupply.Cmp(mb.TotalSupply))
		s.Require().Zero(balance.Cmp(mb.Balance))

		s.T().Logf("ERC20 %s Test04MetadataWithBalance Result", token)
		s.T().Logf("\tName: %s", mb.Name)
		s.T().Logf("\tSymbol: %s", mb.Symbol)
		s.T().Logf("\tDecimals: %d", mb.Decimals)
		s.T().Logf("\tTotalSupply: %s", mb.TotalSupply)
		s.T().Logf("\tBalance: %s", mb.Balance)
	}
}

func (s *ERC20Suite) Test05MetadataWithBalanceAllowFailure() {
	mb, err := s.erc20.MetadataWithBalance(s.ctx, nil, NewMetadataWithBalanceRequest(s.multicall, s.multicall, s.from, true))
	s.Require().NoError(err)
	s.Require().Empty(mb.Name)
	s.Require().Empty(mb.Symbol)
	s.Require().Zero(mb.Decimals)
	s.Require().Nil(mb.TotalSupply)
	s.Require().Nil(mb.Balance)

	s.T().Logf("ERC20 %s Test05MetadataWithBalanceAllowFailure Result", s.multicall)
	s.T().Logf("\tName: %q", mb.Name)
	s.T().Logf("\tSymbol: %q", mb.Symbol)
	s.T().Logf("\tDecimals: %d", mb.Decimals)
	s.T().Logf("\tTotalSupply: %v", mb.TotalSupply)
	s.T().Logf("\tBalance: %v", mb.Balance)
}

func (s *ERC20Suite) Test06AllowanceWithBalance() {
	for _, token := range s.tokens {
		awb, err := s.erc20.AllowanceWithBalance(s.ctx, nil, NewAllowanceWithBalanceRequest(s.multicall, token, s.from, s.to, false))
		s.Require().NoError(err)

		allowance, err := s.erc20.Allowance(s.ctx, nil, NewAllowanceRequest(token, s.from, s.to))
		s.Require().NoError(err)
		ownerBalance, err := s.erc20.BalanceOf(s.ctx, nil, NewBalanceOfRequest(token, s.from))
		s.Require().NoError(err)
		spenderBalance, err := s.erc20.BalanceOf(s.ctx, nil, NewBalanceOfRequest(token, s.to))
		s.Require().NoError(err)

		s.Require().Zero(allowance.Cmp(awb.Value))
		s.Require().Zero(ownerBalance.Cmp(awb.OwnerBalance))
		s.Require().Zero(spenderBalance.Cmp(awb.SpenderBalance))

		s.T().Logf("ERC20 %s Test06AllowanceWithBalance Result", token)
		s.T().Logf("\tAllowance")
		s.T().Logf("\t\tOwner: %s", s.from)
		s.T().Logf("\t\tSpender: %s", s.to)
		s.T().Logf("\t\tValue: %s", awb.Value)
		s.T().Logf("\tOwnerBalance: %s", awb.OwnerBalance)
		s.T().Logf("\tSpenderBalance: %s", awb.SpenderBalance)
	}
}

func (s *ERC20Suite) Test07AllowanceWithBalanceAllowFailure() {
	awb, err := s.erc20.AllowanceWithBalance(s.ctx, nil, NewAllowanceWithBalanceRequest(s.multicall, s.multicall, s.from, s.to, true))
	s.Require().NoError(err)
	s.Require().Nil(awb.Value)
	s.Require().Nil(awb.OwnerBalance)
	s.Require().Nil(awb.SpenderBalance)

	s.T().Logf("ERC20 %s Test07AllowanceWithBalanceAllowFailure Result", s.multicall)
	s.T().Logf("\tAllowance")
	s.T().Logf("\t\tOwner: %s", s.from)
	s.T().Logf("\t\tSpender: %s", s.to)
	s.T().Logf("\t\tValue: %v", awb.Value)
	s.T().Logf("\tOwnerBalance: %v", awb.OwnerBalance)
	s.T().Logf("\tSpenderBalance: %v", awb.SpenderBalance)
}

func (s *ERC20Suite) Test08TokenBalances() {
	addresses := []*types.Address{s.from, s.to}

	for _, token := range s.tokens {
		tb, err := s.erc20.TokenBalances(s.ctx, nil, NewTokenBalancesRequest(s.multicall, token, addresses, false))
		s.Require().NoError(err)
		s.Require().True(tb.Token.Equal(token))
		s.Require().Len(tb.Addresses, len(addresses))

		s.T().Logf("ERC20 %s Test08TokenBalances Result", token)
		for i, addr := range addresses {
			balance, err := s.erc20.BalanceOf(s.ctx, nil, NewBalanceOfRequest(token, addr))
			s.Require().NoError(err)

			s.Require().True(tb.Addresses[i].Address.Equal(addr))
			s.Require().Zero(balance.Cmp(tb.Addresses[i].Amount))
			s.T().Logf("\t%s: %s", addr, tb.Addresses[i].Amount)
		}
	}
}

func (s *ERC20Suite) Test09AddressBalances() {
	for _, addr := range []*types.Address{s.from, s.to} {
		ab, err := s.erc20.AddressBalances(s.ctx, nil, NewAddressBalancesRequest(s.multicall, addr, s.tokens, false))
		s.Require().NoError(err)
		s.Require().True(ab.Address.Equal(addr))
		s.Require().Len(ab.Tokens, len(s.tokens))

		s.T().Logf("ERC20 %s Test09AddressBalances Result", addr)
		for i, token := range s.tokens {
			balance, err := s.erc20.BalanceOf(s.ctx, nil, NewBalanceOfRequest(token, addr))
			s.Require().NoError(err)

			s.Require().True(ab.Tokens[i].Token.Equal(token))
			s.Require().Zero(balance.Cmp(ab.Tokens[i].Amount))
			s.T().Logf("\t%s: %s", token, ab.Tokens[i].Amount)
		}
	}
}

func (s *ERC20Suite) Test10BalancePairs() {
	for _, addr := range []*types.Address{s.from, s.to} {
		addresses := make([]*types.Address, len(s.tokens))
		for i := range addresses {
			addresses[i] = addr
		}

		pairs, err := s.erc20.BalancePairs(s.ctx, nil, NewBalancePairsRequest(s.multicall, s.tokens, addresses, false))
		s.Require().NoError(err)
		s.Require().Len(pairs, len(s.tokens))

		s.T().Logf("ERC20 Test10BalancePairs Result")
		for i, pair := range pairs {
			balance, err := s.erc20.BalanceOf(s.ctx, nil, NewBalanceOfRequest(s.tokens[i], addresses[i]))
			s.Require().NoError(err)

			s.Require().True(pair.Token.Equal(s.tokens[i]))
			s.Require().True(pair.Address.Equal(addresses[i]))
			s.Require().Zero(balance.Cmp(pair.Amount))
			s.T().Logf("\t%s / %s: %s", pair.Token, pair.Address, pair.Amount)
		}
	}
}

func (s *ERC20Suite) Test11Transfer() {
	befores := make([]*big.Int, len(s.tokens))
	for i, token := range s.tokens {
		before, err := s.erc20.BalanceOf(s.ctx, nil, NewBalanceOfRequest(token, s.to))
		s.Require().NoError(err)
		befores[i] = before

		ok, err := s.erc20.CallTransfer(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewTransferRequest(token, s.to, s.value))
		s.Require().NoError(err)
		s.Require().True(ok)
	}

	txs := make([]core.Transaction, len(s.tokens))
	hashes := make([]*types.Hash, len(s.tokens))
	for i, token := range s.tokens {
		builder := tx.NewDynamicFeeBuilder()
		if i > 0 {
			builder.Nonce(txs[0].AddressNonce() + uint64(i))
		}

		signed, hash, err := s.erc20.Transfer(s.ctx, bind.NewTransactOptions(s.signer, builder), NewTransferRequest(token, s.to, s.value))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, signed.MaxFee())
		s.mu.Unlock()

		txs[i] = signed
		hashes[i] = hash
	}

	for i, token := range s.tokens {
		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hashes[i], s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "transaction reverted")

		transfers, err := ExtractTransfers(receipt.Logs)
		s.Require().NoError(err)
		s.Require().Len(transfers, 1)
		s.Require().True(transfers[0].Contract.Equal(token))
		s.Require().True(transfers[0].From.Equal(s.from))
		s.Require().True(transfers[0].To.Equal(s.to))
		s.Require().Zero(s.value.Cmp(transfers[0].Value))

		want := new(big.Int).Add(befores[i], s.value)
		var after *big.Int
		s.Require().Eventually(func() bool {
			after, err = s.erc20.BalanceOf(s.ctx, nil, NewBalanceOfRequest(token, s.to))
			return err == nil && want.Cmp(after) == 0
		}, 10*time.Second, 500*time.Millisecond, "balance not visible after transfer")

		s.T().Logf("ERC20 %s Test11Transfer Result", token)
		s.T().Logf("\tFrom: %s", s.from)
		s.T().Logf("\tTo: %s", s.to)
		s.T().Logf("\tValue: %s", s.value)
		s.T().Logf("\tNonce: %d", txs[i].AddressNonce())
		s.T().Logf("\tGas: %d", txs[i].Gas())
		s.T().Logf("\tMaxFee: %s", txs[i].MaxFee())
		s.T().Logf("\tTx: %s", hashes[i])
		s.T().Logf("\tBalance")
		s.T().Logf("\t\tBefore: %s", befores[i])
		s.T().Logf("\t\tAfter: %s", after)
	}
}

func (s *ERC20Suite) Test12TransferInsufficientBalance() {
	for _, token := range s.tokens {
		decimals, err := s.erc20.Decimals(s.ctx, nil, NewDecimalsRequest(token))
		s.Require().NoError(err)

		scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
		huge := new(big.Int).Mul(big.NewInt(100_000_000), scale)

		ok, err := s.erc20.CallTransfer(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewTransferRequest(token, s.to, huge))
		s.Require().Error(err)
		s.Require().False(ok)
		s.Require().ErrorIs(err, ErrInsufficientBalance)

		var rpcErr *rpc.ResponseError
		s.Require().ErrorAs(err, &rpcErr)
		s.Require().NotEmpty(rpcErr.Data)

		s.T().Logf("ERC20 %s Test12TransferInsufficientBalance Result", token)
		s.T().Logf("\tFrom: %s", s.from)
		s.T().Logf("\tTo: %s", s.to)
		s.T().Logf("\tValue: %s", huge)
		s.T().Logf("\tError: %v", err)
		s.T().Logf("\tRevertData: %s", rpcErr.Data)
	}
}

func (s *ERC20Suite) Test13Approve() {
	for _, token := range s.tokens {
		ok, err := s.erc20.CallApprove(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewApproveRequest(token, s.to, s.value))
		s.Require().NoError(err)
		s.Require().True(ok)
	}

	txs := make([]core.Transaction, len(s.tokens))
	hashes := make([]*types.Hash, len(s.tokens))
	for i, token := range s.tokens {
		builder := tx.NewDynamicFeeBuilder()
		if i > 0 {
			builder.Nonce(txs[0].AddressNonce() + uint64(i))
		}

		signed, hash, err := s.erc20.Approve(s.ctx, bind.NewTransactOptions(s.signer, builder), NewApproveRequest(token, s.to, s.value))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, signed.MaxFee())
		s.mu.Unlock()

		txs[i] = signed
		hashes[i] = hash
	}

	for i, token := range s.tokens {
		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hashes[i], s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "transaction reverted")

		approvals, err := ExtractApprovals(receipt.Logs)
		s.Require().NoError(err)
		s.Require().Len(approvals, 1)
		s.Require().True(approvals[0].Contract.Equal(token))
		s.Require().True(approvals[0].Owner.Equal(s.from))
		s.Require().True(approvals[0].Spender.Equal(s.to))
		s.Require().Zero(s.value.Cmp(approvals[0].Value))

		var allowance *big.Int
		s.Require().Eventually(func() bool {
			allowance, err = s.erc20.Allowance(s.ctx, nil, NewAllowanceRequest(token, s.from, s.to))
			return err == nil && s.value.Cmp(allowance) == 0
		}, 10*time.Second, 500*time.Millisecond, "allowance not visible after approve")

		s.T().Logf("ERC20 %s Test13Approve Result", token)
		s.T().Logf("\tOwner: %s", s.from)
		s.T().Logf("\tSpender: %s", s.to)
		s.T().Logf("\tValue: %s", s.value)
		s.T().Logf("\tNonce: %d", txs[i].AddressNonce())
		s.T().Logf("\tGas: %d", txs[i].Gas())
		s.T().Logf("\tMaxFee: %s", txs[i].MaxFee())
		s.T().Logf("\tTx: %s", hashes[i])
		s.T().Logf("\tAllowance: %s", allowance)
	}
}

func (s *ERC20Suite) Test14TransferFrom() {
	approveTxs := make([]core.Transaction, len(s.tokens))
	approveHashes := make([]*types.Hash, len(s.tokens))
	for i, token := range s.tokens {
		builder := tx.NewDynamicFeeBuilder()
		if i > 0 {
			builder.Nonce(approveTxs[0].AddressNonce() + uint64(i))
		}

		signed, hash, err := s.erc20.Approve(s.ctx, bind.NewTransactOptions(s.signer, builder), NewApproveRequest(token, s.from, s.value))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, signed.MaxFee())
		s.mu.Unlock()

		approveTxs[i] = signed
		approveHashes[i] = hash
	}

	allowances := make([]*big.Int, len(s.tokens))
	for i, token := range s.tokens {
		approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHashes[i], s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(approveReceipt.Status, "approve transaction reverted")

		approvals, err := ExtractApprovals(approveReceipt.Logs)
		s.Require().NoError(err)
		s.Require().Len(approvals, 1)
		s.Require().Zero(s.value.Cmp(approvals[0].Value))

		s.Require().Eventually(func() bool {
			allowances[i], err = s.erc20.Allowance(s.ctx, nil, NewAllowanceRequest(token, s.from, s.from))
			return err == nil && s.value.Cmp(allowances[i]) == 0
		}, 10*time.Second, 500*time.Millisecond, "allowance not visible after approve")
	}

	transferTxs := make([]core.Transaction, len(s.tokens))
	transferHashes := make([]*types.Hash, len(s.tokens))
	for i, token := range s.tokens {
		builder := tx.NewDynamicFeeBuilder()
		if i > 0 {
			builder.Nonce(transferTxs[0].AddressNonce() + uint64(i))
		}

		signed, hash, err := s.erc20.TransferFrom(s.ctx, bind.NewTransactOptions(s.signer, builder), NewTransferFromRequest(token, s.from, s.to, s.value))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, signed.MaxFee())
		s.mu.Unlock()

		transferTxs[i] = signed
		transferHashes[i] = hash
	}

	for i, token := range s.tokens {
		allowance := allowances[i]

		receipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHashes[i], s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "transaction reverted")

		transfers, err := ExtractTransfers(receipt.Logs)
		s.Require().NoError(err)
		s.Require().Len(transfers, 1)
		s.Require().True(transfers[0].Contract.Equal(token))
		s.Require().True(transfers[0].From.Equal(s.from))
		s.Require().True(transfers[0].To.Equal(s.to))
		s.Require().Zero(s.value.Cmp(transfers[0].Value))

		s.T().Logf("ERC20 %s Test14TransferFrom Result", token)
		s.T().Logf("\tApprove")
		s.T().Logf("\t\tOwner: %s", s.from)
		s.T().Logf("\t\tSpender: %s", s.from)
		s.T().Logf("\t\tValue: %s", s.value)
		s.T().Logf("\t\tNonce: %d", approveTxs[i].AddressNonce())
		s.T().Logf("\t\tGas: %d", approveTxs[i].Gas())
		s.T().Logf("\t\tMaxFee: %s", approveTxs[i].MaxFee())
		s.T().Logf("\t\tTx: %s", approveHashes[i])
		s.T().Logf("\t\tAllowance: %s", allowance)
		s.T().Logf("\tTransferFrom")
		s.T().Logf("\t\tSender: %s", s.from)
		s.T().Logf("\t\tFrom: %s", s.from)
		s.T().Logf("\t\tTo: %s", s.to)
		s.T().Logf("\t\tValue: %s", s.value)
		s.T().Logf("\t\tNonce: %d", transferTxs[i].AddressNonce())
		s.T().Logf("\t\tGas: %d", transferTxs[i].Gas())
		s.T().Logf("\t\tMaxFee: %s", transferTxs[i].MaxFee())
		s.T().Logf("\t\tTx: %s", transferHashes[i])
	}
}

func (s *ERC20Suite) Test15TransferFromInsufficientAllowance() {
	sender := types.DeadAddress()

	for _, token := range s.tokens {
		ok, err := s.erc20.CallTransferFrom(s.ctx, bind.NewCallOptions().WithFrom(sender), NewTransferFromRequest(token, s.from, s.to, s.value))
		s.Require().Error(err)
		s.Require().False(ok)
		s.Require().ErrorIs(err, ErrInsufficientAllowance)

		var rpcErr *rpc.ResponseError
		s.Require().ErrorAs(err, &rpcErr)
		s.Require().NotEmpty(rpcErr.Data)

		s.T().Logf("ERC20 %s Test15TransferFromInsufficientAllowance Result", token)
		s.T().Logf("\tSender: %s", sender)
		s.T().Logf("\tFrom: %s", s.from)
		s.T().Logf("\tTo: %s", s.to)
		s.T().Logf("\tValue: %s", s.value)
		s.T().Logf("\tError: %v", err)
		s.T().Logf("\tRevertData: %s", rpcErr.Data)
	}
}

func (s *ERC20Suite) Test16ApproveInvalidSpender() {
	spender := types.ZeroAddress()

	for _, token := range s.tokens {
		ok, err := s.erc20.CallApprove(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewApproveRequest(token, spender, s.value))
		s.Require().Error(err)
		s.Require().False(ok)
		s.Require().ErrorIs(err, ErrInvalidSpender)

		var rpcErr *rpc.ResponseError
		s.Require().ErrorAs(err, &rpcErr)
		s.Require().NotEmpty(rpcErr.Data)

		s.T().Logf("ERC20 %s Test16ApproveInvalidSpender Result", token)
		s.T().Logf("\tOwner: %s", s.from)
		s.T().Logf("\tSpender: %s", spender)
		s.T().Logf("\tValue: %s", s.value)
		s.T().Logf("\tError: %v", err)
		s.T().Logf("\tRevertData: %s", rpcErr.Data)
	}
}
