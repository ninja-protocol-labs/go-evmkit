package eip7702

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
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

type EIP7702Suite struct {
	suite.Suite

	PrivateKeyHex       string
	RPCURL              string
	ChainID             *big.Int
	MulticallHex        string
	NicksFactoryHex     string
	NativeSymbol        string
	NativeDecimals      int
	Timeout             time.Duration
	ReceiptPollInterval time.Duration
	ReceiptTimeout      time.Duration
	LimitBuffer         uint64

	client    rpc.Client
	key       *types.PrivateKey
	from      *types.Address
	multicall *types.Address
	recipient *types.Address

	mu          *sync.Mutex
	totalFee    *big.Int
	authorities []*types.PrivateKey
}

func TestEIP7702Suite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)
	if c.Addresses.Standard.Multicall == "" {
		t.Skip("testenv: no Multicall3 configured for this network")
	}

	suite.Run(t, &EIP7702Suite{
		PrivateKeyHex:       c.ActorPrivateKeyHex,
		RPCURL:              c.RPCURL,
		ChainID:             big.NewInt(int64(c.Constant.ChainID)),
		MulticallHex:        c.Addresses.Standard.Multicall,
		NicksFactoryHex:     c.Addresses.Standard.NicksFactory,
		NativeSymbol:        c.Constant.NativeSymbol,
		NativeDecimals:      c.Constant.NativeDecimals,
		Timeout:             10 * time.Minute,
		ReceiptPollInterval: c.Constant.ReceiptPollInterval,
		ReceiptTimeout:      c.Constant.ReceiptTimeout,
		LimitBuffer:         130,
	})
}

func (s *EIP7702Suite) SetupSuite() {
	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	var err error
	s.key, err = types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.from = core.PubkeyToAddress(s.key.PublicKey())
	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)
	s.recipient = types.NewAddressFromBytes(core.Keccak256(fmt.Appendf(nil, "go-evmkit-eip7702-recipient-%d", time.Now().UnixNano())).Bytes()[12:])
	s.T().Logf("SetupSuite: signing as %s, delegate Multicall3 %s", s.from, s.multicall)
}

func (s *EIP7702Suite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	if len(s.authorities) > 0 {
		revocations := make([]core.Authorization, len(s.authorities))
		authorities := make([]*types.Address, len(s.authorities))
		for i, authorityKey := range s.authorities {
			authorities[i] = core.PubkeyToAddress(authorityKey.PublicKey())
			var nonceHex string
			s.Require().NoError(s.client.Call(ctx, rpc.ETHGetTransactionCount("nonce", authorities[i].String(), rpc.BlockTagLatest, &nonceHex)))
			nonce, ok := new(big.Int).SetString(strings.TrimPrefix(nonceHex, "0x"), 16)
			s.Require().True(ok)
			revocation, err := NewRevocation(s.ChainID, nonce.Uint64())
			s.Require().NoError(err)
			s.Require().NoError(revocation.Sign(authorityKey))
			revocations[i] = *revocation
		}

		revokeHash, err := tx.Submit(ctx, s.client, tx.NewSetCodeTxConfig(s.from, s.recipient, big.NewInt(0), nil, revocations).WithGasLimitBuffer(s.LimitBuffer), s.key)
		s.Require().NoError(err)
		revokeReceipt, err := tx.WaitForReceipt(ctx, s.client, revokeHash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(revokeReceipt.Status, "revocation transaction reverted")
		revokePrice, ok := new(big.Int).SetString(strings.TrimPrefix(revokeReceipt.Raw["effectiveGasPrice"].(string), "0x"), 16)
		s.Require().True(ok)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, new(big.Int).Mul(revokePrice, new(big.Int).SetUint64(revokeReceipt.GasUsed)))
		s.mu.Unlock()

		opts := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", revokeReceipt.BlockNumber))
		for _, authority := range authorities {
			var code *AccountCode
			s.Require().Eventually(func() bool {
				code, err = NewEIP7702(s.client).GetDelegation(ctx, opts, NewGetDelegationRequest(authority))
				return err == nil
			}, 20*time.Second, s.ReceiptPollInterval, "authority code not readable at the revocation block")
			s.Require().Equal(EmptyCode, code.Kind, "authority %s still delegated after teardown", authority)
		}
		s.T().Logf("TearDownSuite: revoked %d delegations in one type-4 tx (tx %s, block %d, gas %d)", len(authorities), revokeHash, revokeReceipt.BlockNumber, revokeReceipt.GasUsed)
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
		s.T().Logf("EIP7702Suite total fee spent: %s %s", decimal, s.NativeSymbol)
	}
}

func (s *EIP7702Suite) Test01GetDelegationContractCode() {
	code, err := NewEIP7702(s.client).GetDelegation(s.T().Context(), nil, NewGetDelegationRequest(s.multicall))
	s.Require().NoError(err)
	s.Require().Equal(ContractCode, code.Kind)
	s.Require().Nil(code.Delegate)

	s.T().Logf("EIP7702 Test01GetDelegationContractCode Result")
	s.T().Logf("\tMulticall3 %s: %s", s.multicall, code.Kind)
}

func (s *EIP7702Suite) Test02GetDelegationEmptyCode() {
	authorityKey, err := types.GeneratePrivateKey()
	s.Require().NoError(err)
	authority := core.PubkeyToAddress(authorityKey.PublicKey())

	code, err := NewEIP7702(s.client).GetDelegation(s.T().Context(), bind.NewCallOptions().WithBlock(rpc.BlockTagLatest), NewGetDelegationRequest(authority))
	s.Require().NoError(err)
	s.Require().Equal(EmptyCode, code.Kind)
	s.Require().Nil(code.Delegate)

	s.T().Logf("EIP7702 Test02GetDelegationEmptyCode Result")
	s.T().Logf("\tFresh account %s: %s", authority, code.Kind)
}

func (s *EIP7702Suite) Test03SponsoredDelegation() {
	authorityKey, err := types.GeneratePrivateKey()
	s.Require().NoError(err)
	authority := core.PubkeyToAddress(authorityKey.PublicKey())
	auth, err := NewAuthorization(s.ChainID, s.multicall, 0)
	s.Require().NoError(err)
	s.Require().NoError(auth.Sign(authorityKey))
	signer, err := auth.Authority()
	s.Require().NoError(err)
	s.Require().True(signer.Equal(authority))
	s.Require().NoError(CheckAuthorization(auth, authority, s.ChainID, s.multicall))

	data, err := abi.NewFunction("getBlockNumber", nil, abi.NewTypes(abi.Uint256)).EncodeCall()
	s.Require().NoError(err)
	hash, err := tx.Submit(s.T().Context(), s.client, tx.NewSetCodeTxConfig(s.from, authority, big.NewInt(0), data, []core.Authorization{*auth}).WithGasLimitBuffer(s.LimitBuffer), s.key)
	s.Require().NoError(err)
	s.T().Logf("Test03SponsoredDelegation: sent type-4 call into %s, tx %s", authority, hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "set code transaction reverted")
	price, ok := new(big.Int).SetString(strings.TrimPrefix(receipt.Raw["effectiveGasPrice"].(string), "0x"), 16)
	s.Require().True(ok)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(price, new(big.Int).SetUint64(receipt.GasUsed)))
	s.authorities = append(s.authorities, authorityKey)
	s.mu.Unlock()

	var code *AccountCode
	s.Require().Eventually(func() bool {
		code, err = NewEIP7702(s.client).GetDelegation(s.T().Context(), bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber)), NewGetDelegationRequest(authority))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "authority code not readable at the receipt block")
	s.Require().Equal(DelegatedCode, code.Kind)
	s.Require().True(s.multicall.Equal(code.Delegate))

	before, err := NewEIP7702(s.client).GetDelegation(s.T().Context(), bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber-1)), NewGetDelegationRequest(authority))
	s.Require().NoError(err)
	s.Require().Equal(EmptyCode, before.Kind)

	out, err := bind.Call(s.T().Context(), s.client, bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber)), authority, data)
	s.Require().NoError(err)
	s.Require().Len(out, 32)
	s.Require().Positive(new(big.Int).SetBytes(out).Sign())

	var nonceHex string
	s.Require().NoError(s.client.Call(s.T().Context(), rpc.ETHGetTransactionCount("nonce", authority.String(), fmt.Sprintf("0x%x", receipt.BlockNumber), &nonceHex)))
	s.Require().Equal("0x1", nonceHex)

	s.T().Logf("EIP7702 Test03SponsoredDelegation Result")
	s.T().Logf("\tAuthority %s: %s -> %s (delegate %s), nonce %s", authority, before.Kind, code.Kind, code.Delegate, nonceHex)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *EIP7702Suite) Test04SelfSponsoredDelegation() {
	authorityKey, err := types.GeneratePrivateKey()
	s.Require().NoError(err)
	authority := core.PubkeyToAddress(authorityKey.PublicKey())

	txNonce := uint64(0)
	authNonce, err := SelfAuthorizationNonce(txNonce)
	s.Require().NoError(err)
	auth, err := NewAuthorization(s.ChainID, s.multicall, authNonce)
	s.Require().NoError(err)
	s.Require().NoError(auth.Sign(authorityKey))

	data, err := abi.NewFunction("getBlockNumber", nil, abi.NewTypes(abi.Uint256)).EncodeCall()
	s.Require().NoError(err)
	selfTx, err := tx.NewSetCodeTxConfig(authority, authority, big.NewInt(0), data, []core.Authorization{*auth}).
		WithChainID(s.ChainID).
		WithNonce(txNonce).
		WithGasLimitBuffer(s.LimitBuffer).
		Pack(s.T().Context(), s.client)
	s.Require().NoError(err)

	funding := new(big.Int).Mul(selfTx.MaxFee(), big.NewInt(2))
	fundHash, err := tx.Submit(s.T().Context(), s.client, tx.NewDynamicFeeTxConfig(s.from, authority, funding, nil).WithGasLimitBuffer(s.LimitBuffer), s.key)
	s.Require().NoError(err)
	fundReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, fundHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(fundReceipt.Status, "funding transaction reverted")
	fundPrice, ok := new(big.Int).SetString(strings.TrimPrefix(fundReceipt.Raw["effectiveGasPrice"].(string), "0x"), 16)
	s.Require().True(ok)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(fundPrice, new(big.Int).SetUint64(fundReceipt.GasUsed)))
	s.totalFee.Add(s.totalFee, funding)
	s.mu.Unlock()
	s.T().Logf("Test04SelfSponsoredDelegation: funded %s with %s wei (2x max L2 fee), tx %s", authority, funding, fundHash)

	s.Require().NoError(selfTx.Sign(authorityKey))
	raw, err := selfTx.EncodeRLP()
	s.Require().NoError(err)
	var hash *types.Hash
	deadline := time.Now().Add(20 * time.Second)
	for {
		hash, err = tx.Broadcast(s.T().Context(), s.client, raw)
		if err == nil || time.Now().After(deadline) {
			break
		}
		time.Sleep(s.ReceiptPollInterval)
	}
	s.Require().NoError(err, "self-sponsored tx not accepted after funding")
	s.T().Logf("Test04SelfSponsoredDelegation: sent self-sponsored type-4 tx, tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "self-sponsored set code transaction reverted")
	s.mu.Lock()
	s.authorities = append(s.authorities, authorityKey)
	s.mu.Unlock()

	var code *AccountCode
	s.Require().Eventually(func() bool {
		code, err = NewEIP7702(s.client).GetDelegation(s.T().Context(), bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber)), NewGetDelegationRequest(authority))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "authority code not readable at the receipt block")
	s.Require().Equal(DelegatedCode, code.Kind, "authorization nonce txNonce+1 must apply for a self-sponsored tx")
	s.Require().True(s.multicall.Equal(code.Delegate))

	var nonceHex string
	s.Require().NoError(s.client.Call(s.T().Context(), rpc.ETHGetTransactionCount("nonce", authority.String(), fmt.Sprintf("0x%x", receipt.BlockNumber), &nonceHex)))
	s.Require().Equal("0x2", nonceHex, "the tx and the authorization each advance the nonce")

	s.T().Logf("EIP7702 Test04SelfSponsoredDelegation Result")
	s.T().Logf("\tAuthority %s: tx nonce %d, authorization nonce %d, %s (delegate %s), nonce now %s", authority, txNonce, authNonce, code.Kind, code.Delegate, nonceHex)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *EIP7702Suite) Test05StaleNonceAuthorizationIsSkipped() {
	authorityKey, err := types.GeneratePrivateKey()
	s.Require().NoError(err)
	authority := core.PubkeyToAddress(authorityKey.PublicKey())
	auth, err := NewAuthorization(s.ChainID, s.multicall, 1)
	s.Require().NoError(err)
	s.Require().NoError(auth.Sign(authorityKey))

	hash, err := tx.Submit(s.T().Context(), s.client, tx.NewSetCodeTxConfig(s.from, s.recipient, big.NewInt(0), nil, []core.Authorization{*auth}).WithGasLimitBuffer(s.LimitBuffer), s.key)
	s.Require().NoError(err)
	s.T().Logf("Test05StaleNonceAuthorizationIsSkipped: sent type-4 tx with a wrong-nonce authorization, tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "a skipped authorization must not revert the transaction")
	price, ok := new(big.Int).SetString(strings.TrimPrefix(receipt.Raw["effectiveGasPrice"].(string), "0x"), 16)
	s.Require().True(ok)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(price, new(big.Int).SetUint64(receipt.GasUsed)))
	s.mu.Unlock()

	var code *AccountCode
	s.Require().Eventually(func() bool {
		code, err = NewEIP7702(s.client).GetDelegation(s.T().Context(), bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber)), NewGetDelegationRequest(authority))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "authority code not readable at the receipt block")
	s.Require().Equal(EmptyCode, code.Kind, "an authorization with the wrong nonce must not delegate")

	var nonceHex string
	s.Require().NoError(s.client.Call(s.T().Context(), rpc.ETHGetTransactionCount("nonce", authority.String(), fmt.Sprintf("0x%x", receipt.BlockNumber), &nonceHex)))
	s.Require().Equal("0x0", nonceHex, "a skipped authorization must not advance the authority nonce")

	s.T().Logf("EIP7702 Test05StaleNonceAuthorizationIsSkipped Result")
	s.T().Logf("\tAuthority %s: authorization nonce %d against nonce 0, %s, nonce %s", authority, auth.Nonce, code.Kind, nonceHex)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *EIP7702Suite) Test06Revocation() {
	authorityKey, err := types.GeneratePrivateKey()
	s.Require().NoError(err)
	authority := core.PubkeyToAddress(authorityKey.PublicKey())

	auth, err := NewAuthorization(s.ChainID, s.multicall, 0)
	s.Require().NoError(err)
	s.Require().NoError(auth.Sign(authorityKey))
	delegateHash, err := tx.Submit(s.T().Context(), s.client, tx.NewSetCodeTxConfig(s.from, s.recipient, big.NewInt(0), nil, []core.Authorization{*auth}).WithGasLimitBuffer(s.LimitBuffer), s.key)
	s.Require().NoError(err)
	delegateReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, delegateHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(delegateReceipt.Status, "delegation transaction reverted")
	delegatePrice, ok := new(big.Int).SetString(strings.TrimPrefix(delegateReceipt.Raw["effectiveGasPrice"].(string), "0x"), 16)
	s.Require().True(ok)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(delegatePrice, new(big.Int).SetUint64(delegateReceipt.GasUsed)))
	s.mu.Unlock()

	var delegated *AccountCode
	s.Require().Eventually(func() bool {
		delegated, err = NewEIP7702(s.client).GetDelegation(s.T().Context(), bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", delegateReceipt.BlockNumber)), NewGetDelegationRequest(authority))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "authority code not readable at the delegation block")
	s.Require().Equal(DelegatedCode, delegated.Kind)

	revocation, err := NewRevocation(s.ChainID, 1)
	s.Require().NoError(err)
	s.Require().NoError(revocation.Sign(authorityKey))
	var revokeHash *types.Hash
	s.Require().Eventually(func() bool {
		revokeHash, err = tx.Submit(s.T().Context(), s.client, tx.NewSetCodeTxConfig(s.from, s.recipient, big.NewInt(0), nil, []core.Authorization{*revocation}).WithGasLimitBuffer(s.LimitBuffer), s.key)
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "revocation not accepted by the estimating node")
	s.T().Logf("Test06Revocation: delegated in %s, revoking in %s", delegateHash, revokeHash)

	revokeReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, revokeHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(revokeReceipt.Status, "revocation transaction reverted")
	revokePrice, ok := new(big.Int).SetString(strings.TrimPrefix(revokeReceipt.Raw["effectiveGasPrice"].(string), "0x"), 16)
	s.Require().True(ok)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(revokePrice, new(big.Int).SetUint64(revokeReceipt.GasUsed)))
	s.mu.Unlock()

	var revoked *AccountCode
	s.Require().Eventually(func() bool {
		revoked, err = NewEIP7702(s.client).GetDelegation(s.T().Context(), bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", revokeReceipt.BlockNumber)), NewGetDelegationRequest(authority))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "authority code not readable at the revocation block")
	s.Require().Equal(EmptyCode, revoked.Kind, "a revocation must clear the delegation")
	s.Require().Nil(revoked.Delegate)

	var nonceHex string
	s.Require().NoError(s.client.Call(s.T().Context(), rpc.ETHGetTransactionCount("nonce", authority.String(), fmt.Sprintf("0x%x", revokeReceipt.BlockNumber), &nonceHex)))
	s.Require().Equal("0x2", nonceHex, "each applied authorization advances the authority nonce")

	s.T().Logf("EIP7702 Test06Revocation Result")
	s.T().Logf("\tAuthority %s: %s -> %s, nonce %s", authority, delegated.Kind, revoked.Kind, nonceHex)
	s.T().Logf("\tTx: %s (block %d, gas %d)", revokeHash, revokeReceipt.BlockNumber, revokeReceipt.GasUsed)
}

func (s *EIP7702Suite) Test07DelegationOverrideCall() {
	authorityKey, err := types.GeneratePrivateKey()
	s.Require().NoError(err)
	authority := core.PubkeyToAddress(authorityKey.PublicKey())

	designator, err := EncodeDelegation(s.multicall)
	s.Require().NoError(err)
	data, err := abi.NewFunction("getBlockNumber", nil, abi.NewTypes(abi.Uint256)).EncodeCall()
	s.Require().NoError(err)

	plain, err := bind.Call(s.T().Context(), s.client, nil, authority, data)
	s.Require().NoError(err)
	s.Require().Empty(plain, "a call into an undelegated EOA returns no data")

	opts := bind.NewCallOptions().WithStateOverride(rpc.NewStateOverride().WithCode(authority, designator))
	out, err := bind.Call(s.T().Context(), s.client, opts, authority, data)
	s.Require().NoError(err)
	s.Require().Len(out, 32)
	s.Require().Positive(new(big.Int).SetBytes(out).Sign())

	code, err := NewEIP7702(s.client).GetDelegation(s.T().Context(), nil, NewGetDelegationRequest(authority))
	s.Require().NoError(err)
	s.Require().Equal(EmptyCode, code.Kind, "an override must not persist")

	s.T().Logf("EIP7702 Test07DelegationOverrideCall Result")
	s.T().Logf("\tAuthority %s with override %x: getBlockNumber %s, onchain %s", authority, designator, new(big.Int).SetBytes(out), code.Kind)
}

func (s *EIP7702Suite) Test08AnyChainAuthorization() {
	authorityKey, err := types.GeneratePrivateKey()
	s.Require().NoError(err)
	authority := core.PubkeyToAddress(authorityKey.PublicKey())
	auth := core.NewAuthorization(big.NewInt(0), s.multicall, 0)
	s.Require().NoError(auth.Sign(authorityKey))
	s.Require().ErrorIs(CheckAuthorization(auth, authority, s.ChainID, s.multicall), ErrInvalidChainID)

	hash, err := tx.Submit(s.T().Context(), s.client, tx.NewSetCodeTxConfig(s.from, s.recipient, big.NewInt(0), nil, []core.Authorization{*auth}).WithGasLimitBuffer(s.LimitBuffer), s.key)
	s.Require().NoError(err)
	s.T().Logf("Test08AnyChainAuthorization: sent type-4 tx with a chainId 0 authorization, tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "set code transaction reverted")
	price, ok := new(big.Int).SetString(strings.TrimPrefix(receipt.Raw["effectiveGasPrice"].(string), "0x"), 16)
	s.Require().True(ok)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(price, new(big.Int).SetUint64(receipt.GasUsed)))
	s.authorities = append(s.authorities, authorityKey)
	s.mu.Unlock()

	var code *AccountCode
	s.Require().Eventually(func() bool {
		code, err = NewEIP7702(s.client).GetDelegation(s.T().Context(), bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber)), NewGetDelegationRequest(authority))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "authority code not readable at the receipt block")
	s.Require().Equal(DelegatedCode, code.Kind, "a chainId 0 authorization applies on every chain")
	s.Require().True(s.multicall.Equal(code.Delegate))

	s.T().Logf("EIP7702 Test08AnyChainAuthorization Result")
	s.T().Logf("\tAuthority %s: chainId 0 authorization -> %s (delegate %s)", authority, code.Kind, code.Delegate)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *EIP7702Suite) Test09Redelegation() {
	if s.NicksFactoryHex == "" {
		s.T().Skip("testenv: no Nick's factory configured as a second delegate")
	}
	second, err := types.NewAddressFromHex(s.NicksFactoryHex)
	s.Require().NoError(err)

	authorityKey, err := types.GeneratePrivateKey()
	s.Require().NoError(err)
	authority := core.PubkeyToAddress(authorityKey.PublicKey())

	first, err := NewAuthorization(s.ChainID, s.multicall, 0)
	s.Require().NoError(err)
	s.Require().NoError(first.Sign(authorityKey))
	firstHash, err := tx.Submit(s.T().Context(), s.client, tx.NewSetCodeTxConfig(s.from, s.recipient, big.NewInt(0), nil, []core.Authorization{*first}).WithGasLimitBuffer(s.LimitBuffer), s.key)
	s.Require().NoError(err)
	firstReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, firstHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(firstReceipt.Status, "first delegation transaction reverted")
	firstPrice, ok := new(big.Int).SetString(strings.TrimPrefix(firstReceipt.Raw["effectiveGasPrice"].(string), "0x"), 16)
	s.Require().True(ok)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(firstPrice, new(big.Int).SetUint64(firstReceipt.GasUsed)))
	s.authorities = append(s.authorities, authorityKey)
	s.mu.Unlock()

	var before *AccountCode
	s.Require().Eventually(func() bool {
		before, err = NewEIP7702(s.client).GetDelegation(s.T().Context(), bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", firstReceipt.BlockNumber)), NewGetDelegationRequest(authority))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "authority code not readable at the first delegation block")
	s.Require().Equal(DelegatedCode, before.Kind)
	s.Require().True(s.multicall.Equal(before.Delegate))

	redelegate, err := NewAuthorization(s.ChainID, second, 1)
	s.Require().NoError(err)
	s.Require().NoError(redelegate.Sign(authorityKey))
	s.Require().NoError(CheckAuthorization(redelegate, authority, s.ChainID, second))
	var hash *types.Hash
	s.Require().Eventually(func() bool {
		hash, err = tx.Submit(s.T().Context(), s.client, tx.NewSetCodeTxConfig(s.from, s.recipient, big.NewInt(0), nil, []core.Authorization{*redelegate}).WithGasLimitBuffer(s.LimitBuffer), s.key)
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "redelegation not accepted by the estimating node")
	s.T().Logf("Test09Redelegation: delegated in %s, redelegating in %s", firstHash, hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "redelegation transaction reverted")
	price, ok := new(big.Int).SetString(strings.TrimPrefix(receipt.Raw["effectiveGasPrice"].(string), "0x"), 16)
	s.Require().True(ok)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(price, new(big.Int).SetUint64(receipt.GasUsed)))
	s.mu.Unlock()

	var after *AccountCode
	s.Require().Eventually(func() bool {
		after, err = NewEIP7702(s.client).GetDelegation(s.T().Context(), bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber)), NewGetDelegationRequest(authority))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "authority code not readable at the redelegation block")
	s.Require().Equal(DelegatedCode, after.Kind)
	s.Require().True(second.Equal(after.Delegate), "a new authorization replaces the previous delegate")

	var nonceHex string
	s.Require().NoError(s.client.Call(s.T().Context(), rpc.ETHGetTransactionCount("nonce", authority.String(), fmt.Sprintf("0x%x", receipt.BlockNumber), &nonceHex)))
	s.Require().Equal("0x2", nonceHex)

	s.T().Logf("EIP7702 Test09Redelegation Result")
	s.T().Logf("\tAuthority %s: delegate %s -> %s, nonce %s", authority, before.Delegate, after.Delegate, nonceHex)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *EIP7702Suite) Test10GetDelegationElementBatch() {
	authorityKey, err := types.GeneratePrivateKey()
	s.Require().NoError(err)
	fresh := core.PubkeyToAddress(authorityKey.PublicKey())

	var blockHex string
	s.Require().NoError(s.client.Call(s.T().Context(), rpc.ETHBlockNumber("block", &blockHex)))
	opts := bind.NewCallOptions().WithBlock(blockHex)

	var contractCode, emptyCode AccountCode
	var nonceHex string
	contractElem, err := GetDelegationElement("multicall", opts, NewGetDelegationRequest(s.multicall), &contractCode)
	s.Require().NoError(err)
	emptyElem, err := GetDelegationElement("fresh", opts, NewGetDelegationRequest(fresh), &emptyCode)
	s.Require().NoError(err)
	elems := rpc.Elements{}
	elems.With(contractElem)
	elems.With(emptyElem)
	elems.With(rpc.ETHGetTransactionCount("nonce", fresh.String(), blockHex, &nonceHex))
	s.Require().NoError(s.client.Batch(s.T().Context(), elems))

	s.Require().Equal(ContractCode, contractCode.Kind)
	s.Require().Equal(EmptyCode, emptyCode.Kind)
	s.Require().Equal("0x0", nonceHex)

	s.T().Logf("EIP7702 Test10GetDelegationElementBatch Result")
	s.T().Logf("\tBlock %s: Multicall3 %s, fresh %s %s (nonce %s) in one batch", blockHex, contractCode.Kind, fresh, emptyCode.Kind, nonceHex)
}
