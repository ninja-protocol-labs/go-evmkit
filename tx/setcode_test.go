package tx

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/internal/testenv"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/eip1014"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

type SetCodeTxSuite struct {
	suite.Suite

	RPCURL              string
	PrivateKeyHex       string
	ChainID             *big.Int
	MulticallHex        string
	NicksFactoryHex     string
	NativeSymbol        string
	NativeDecimals      int
	Timeout             time.Duration
	ReceiptPollInterval time.Duration
	ReceiptTimeout      time.Duration
	LimitBuffer         uint64
	ByteCode            string
	Constructor         *abi.Function
	Salt                *types.Hash
	MintValue           *big.Int
	TransferValue       *big.Int
	NativeValue         *big.Int

	client    rpc.Client
	key       *types.PrivateKey
	from      *types.Address
	multicall *types.Address
	nicks     *types.Address
	token     *types.Address
	recipient *types.Address

	mu          *sync.Mutex
	totalFee    *big.Int
	authorities []*types.PrivateKey
}

func TestSetCodeTxSuite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)
	if c.Addresses.Standard.Multicall == "" || c.Addresses.Standard.NicksFactory == "" {
		t.Skip("testenv: no Multicall3/Nick's factory configured for this network")
	}

	suite.Run(t, &SetCodeTxSuite{
		RPCURL:              c.RPCURL,
		PrivateKeyHex:       c.ActorPrivateKeyHex,
		ChainID:             big.NewInt(int64(c.Constant.ChainID)),
		MulticallHex:        c.Addresses.Standard.Multicall,
		NicksFactoryHex:     c.Addresses.Standard.NicksFactory,
		NativeSymbol:        c.Constant.NativeSymbol,
		NativeDecimals:      c.Constant.NativeDecimals,
		Timeout:             10 * time.Minute,
		ReceiptPollInterval: c.Constant.ReceiptPollInterval,
		ReceiptTimeout:      c.Constant.ReceiptTimeout,
		LimitBuffer:         130,
		ByteCode:            testenv.ERC20ByteCodeHex,
		Constructor:         abi.NewFunction("constructor", abi.Types{abi.String, abi.String, abi.Uint8}, nil),
		Salt:                core.Keccak256([]byte("go-evmkit-tx-setcode")),
		MintValue:           new(big.Int).Exp(big.NewInt(10), big.NewInt(21), nil),
		TransferValue:       new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
		NativeValue:         big.NewInt(1_000_000_000_000),
	})
}

func (s *SetCodeTxSuite) SetupSuite() {
	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	var err error
	s.key, err = types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.from = core.PubkeyToAddress(s.key.PublicKey())
	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)
	s.nicks, err = types.NewAddressFromHex(s.NicksFactoryHex)
	s.Require().NoError(err)
	s.T().Logf("SetupSuite: signing as %s, Multicall3 %s, Nick's factory %s", s.from, s.multicall, s.nicks)

	runtime := fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())
	s.recipient = types.NewAddressFromBytes(core.Keccak256(fmt.Appendf(nil, "%s-recipient", runtime)).Bytes()[12:])

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)
	args, err := abi.Pack(s.Constructor.Inputs, "evmkit setcode test token", "EVSCT", uint8(18))
	s.Require().NoError(err)
	initCode := append(append([]byte{}, bytecode...), args...)
	var salt [32]byte
	copy(salt[:], core.Keccak256(runtime).Bytes())

	deployTx, err := NewDynamicFeeTxConfig(s.from, s.nicks, big.NewInt(0), append(append([]byte{}, salt[:]...), initCode...)).
		WithGasLimitBuffer(s.LimitBuffer).
		Pack(s.T().Context(), s.client)
	s.Require().NoError(err)
	s.Require().NoError(deployTx.Sign(s.key))
	deployRaw, err := deployTx.EncodeRLP()
	s.Require().NoError(err)
	deployHash, err := Broadcast(s.T().Context(), s.client, deployRaw)
	s.Require().NoError(err)
	deployReceipt, err := WaitForReceipt(s.T().Context(), s.client, deployHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(deployReceipt.Status, "deploy transaction reverted")
	deployPrice, ok := new(big.Int).SetString(strings.TrimPrefix(deployReceipt.Raw["effectiveGasPrice"].(string), "0x"), 16)
	s.Require().True(ok)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(deployPrice, new(big.Int).SetUint64(deployReceipt.GasUsed)))
	s.mu.Unlock()

	s.token = eip1014.Address(s.nicks, salt, initCode)
	s.Require().Eventually(func() bool {
		var codeHex string
		if err := s.client.Call(s.T().Context(), rpc.ETHGetCode("code", s.token.String(), rpc.BlockTagLatest, &codeHex)); err != nil {
			return false
		}
		code, err := encoding.Hex.Decode(codeHex)
		return err == nil && len(code) > 0
	}, 20*time.Second, s.ReceiptPollInterval, "no code visible at deployed token %s", s.token)
	s.T().Logf("SetupSuite: deployed token %s (tx %s, block %d)", s.token, deployHash, deployReceipt.BlockNumber)

	mintData, err := abi.NewFunction("mint", abi.NewTypes(abi.Address, abi.Uint256), nil).EncodeCall(s.from, s.MintValue)
	s.Require().NoError(err)
	mintTx, err := NewDynamicFeeTxConfig(s.from, s.token, big.NewInt(0), mintData).
		WithGasLimitBuffer(s.LimitBuffer).
		Pack(s.T().Context(), s.client)
	s.Require().NoError(err)
	s.Require().NoError(mintTx.Sign(s.key))
	mintRaw, err := mintTx.EncodeRLP()
	s.Require().NoError(err)
	mintHash, err := Broadcast(s.T().Context(), s.client, mintRaw)
	s.Require().NoError(err)
	mintReceipt, err := WaitForReceipt(s.T().Context(), s.client, mintHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(mintReceipt.Status, "mint transaction reverted")
	mintPrice, ok := new(big.Int).SetString(strings.TrimPrefix(mintReceipt.Raw["effectiveGasPrice"].(string), "0x"), 16)
	s.Require().True(ok)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(mintPrice, new(big.Int).SetUint64(mintReceipt.GasUsed)))
	s.mu.Unlock()
	s.T().Logf("SetupSuite: minted %s to %s (tx %s)", s.MintValue, s.from, mintHash)
}

func (s *SetCodeTxSuite) TearDownSuite() {
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
			revocations[i] = core.Authorization{ChainID: s.ChainID, Address: types.ZeroAddress(), Nonce: nonce.Uint64()}
			s.Require().NoError(revocations[i].Sign(authorityKey))
		}

		revokeHash, err := Submit(ctx, s.client, NewSetCodeTxConfig(s.from, s.recipient, big.NewInt(0), nil, revocations).WithGasLimitBuffer(s.LimitBuffer), s.key)
		s.Require().NoError(err)
		revokeReceipt, err := WaitForReceipt(ctx, s.client, revokeHash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(revokeReceipt.Status, "revocation transaction reverted")
		revokePrice, ok := new(big.Int).SetString(strings.TrimPrefix(revokeReceipt.Raw["effectiveGasPrice"].(string), "0x"), 16)
		s.Require().True(ok)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, new(big.Int).Mul(revokePrice, new(big.Int).SetUint64(revokeReceipt.GasUsed)))
		s.mu.Unlock()

		atBlock := fmt.Sprintf("0x%x", revokeReceipt.BlockNumber)
		for _, authority := range authorities {
			var codeHex string
			s.Require().Eventually(func() bool {
				return s.client.Call(ctx, rpc.ETHGetCode("code", authority.String(), atBlock, &codeHex)) == nil
			}, 20*time.Second, s.ReceiptPollInterval, "authority code not readable at the revocation block")
			s.Require().Equal("0x", codeHex, "authority %s still delegated after teardown", authority)
		}
		s.T().Logf("TearDownSuite: revoked %d delegations in one type-4 tx (tx %s, block %d, gas %d)", len(authorities), revokeHash, revokeReceipt.BlockNumber, revokeReceipt.GasUsed)
	}

	if s.token != nil {
		balanceOfFn := abi.NewFunction("balanceOf", abi.NewTypes(abi.Address), abi.NewTypes(abi.Uint256))
		burnFn := abi.NewFunction("burn", abi.NewTypes(abi.Address, abi.Uint256), nil)

		for _, holder := range []*types.Address{s.from, s.recipient} {
			callData, err := balanceOfFn.EncodeCall(holder)
			s.Require().NoError(err)
			var balanceHex string
			s.Require().NoError(s.client.Call(ctx, rpc.ETHCall("balance", map[string]any{"to": s.token.String(), "data": encoding.Hex.EncodePrefixed(callData)}, rpc.BlockTagLatest, nil, &balanceHex)))
			balanceRaw, err := encoding.Hex.Decode(balanceHex)
			s.Require().NoError(err)
			balance := new(big.Int).SetBytes(balanceRaw)
			if balance.Sign() == 0 {
				continue
			}

			burnData, err := burnFn.EncodeCall(holder, balance)
			s.Require().NoError(err)
			burnTx, err := NewDynamicFeeTxConfig(s.from, s.token, big.NewInt(0), burnData).
				WithGasLimitBuffer(s.LimitBuffer).
				Pack(ctx, s.client)
			s.Require().NoError(err)
			s.Require().NoError(burnTx.Sign(s.key))
			burnRaw, err := burnTx.EncodeRLP()
			s.Require().NoError(err)
			burnHash, err := Broadcast(ctx, s.client, burnRaw)
			s.Require().NoError(err)
			burnReceipt, err := WaitForReceipt(ctx, s.client, burnHash, s.ReceiptPollInterval, s.ReceiptTimeout)
			s.Require().NoError(err)
			s.Require().True(burnReceipt.Status, "burn transaction reverted")
			burnPrice, ok := new(big.Int).SetString(strings.TrimPrefix(burnReceipt.Raw["effectiveGasPrice"].(string), "0x"), 16)
			s.Require().True(ok)
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, new(big.Int).Mul(burnPrice, new(big.Int).SetUint64(burnReceipt.GasUsed)))
			s.mu.Unlock()
			s.T().Logf("TearDownSuite: burned %s of %s from %s (tx %s)", balance, s.token, holder, burnHash)
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
		s.T().Logf("SetCodeTxSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}
}

func (s *SetCodeTxSuite) Test01SubmitETHTransferDelegatesAuthority() {
	authorityKey, err := types.GeneratePrivateKey()
	s.Require().NoError(err)
	authority := core.PubkeyToAddress(authorityKey.PublicKey())
	auth := core.Authorization{ChainID: s.ChainID, Address: s.multicall, Nonce: 0}
	s.Require().NoError(auth.Sign(authorityKey))

	cfg := NewSetCodeTxConfig(s.from, s.recipient, s.NativeValue, nil, []core.Authorization{auth}).WithGasLimitBuffer(s.LimitBuffer)
	hash, err := Submit(s.T().Context(), s.client, cfg, s.key)
	s.Require().NoError(err)
	s.T().Logf("Test01SubmitETHTransferDelegatesAuthority: sent type-4 transfer, tx %s", hash)

	receipt, err := WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "set code transaction reverted")
	price, ok := new(big.Int).SetString(strings.TrimPrefix(receipt.Raw["effectiveGasPrice"].(string), "0x"), 16)
	s.Require().True(ok)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(price, new(big.Int).SetUint64(receipt.GasUsed)))
	s.mu.Unlock()

	atBlock := fmt.Sprintf("0x%x", receipt.BlockNumber)
	beforeBlock := fmt.Sprintf("0x%x", receipt.BlockNumber-1)
	wantCode := "0xef0100" + strings.TrimPrefix(strings.ToLower(s.multicall.String()), "0x")
	var codeHex string
	s.Require().Eventually(func() bool {
		return s.client.Call(s.T().Context(), rpc.ETHGetCode("code", authority.String(), atBlock, &codeHex)) == nil
	}, 20*time.Second, s.ReceiptPollInterval, "authority code not readable at the receipt block")
	s.Require().Equal(wantCode, strings.ToLower(codeHex), "authority is not delegated to Multicall3")
	s.mu.Lock()
	s.authorities = append(s.authorities, authorityKey)
	s.mu.Unlock()

	var nonceHex, afterHex, beforeHex string
	s.Require().NoError(s.client.Call(s.T().Context(), rpc.ETHGetTransactionCount("nonce", authority.String(), atBlock, &nonceHex)))
	s.Require().Equal("0x1", nonceHex, "authority nonce should advance once")
	s.Require().NoError(s.client.Call(s.T().Context(), rpc.ETHGetBalance("after", s.recipient.String(), atBlock, &afterHex)))
	s.Require().NoError(s.client.Call(s.T().Context(), rpc.ETHGetBalance("before", s.recipient.String(), beforeBlock, &beforeHex)))
	after, ok := new(big.Int).SetString(strings.TrimPrefix(afterHex, "0x"), 16)
	s.Require().True(ok)
	before, ok := new(big.Int).SetString(strings.TrimPrefix(beforeHex, "0x"), 16)
	s.Require().True(ok)
	s.Require().Zero(s.NativeValue.Cmp(new(big.Int).Sub(after, before)), "recipient did not receive exactly the native value")

	s.T().Logf("SetCode Test01SubmitETHTransferDelegatesAuthority Result")
	s.T().Logf("\tAuthority %s code %s nonce %s", authority, codeHex, nonceHex)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *SetCodeTxSuite) Test02PackSignBroadcastTokenTransfer() {
	authorityKey, err := types.GeneratePrivateKey()
	s.Require().NoError(err)
	auth := core.Authorization{ChainID: s.ChainID, Address: s.multicall, Nonce: 0}
	s.Require().NoError(auth.Sign(authorityKey))

	data, err := abi.NewFunction("transfer", abi.NewTypes(abi.Address, abi.Uint256), abi.NewTypes(abi.Bool)).EncodeCall(s.recipient, s.TransferValue)
	s.Require().NoError(err)

	txn, err := NewSetCodeTxConfig(s.from, s.token, big.NewInt(0), data, []core.Authorization{auth}).
		WithGasLimitBuffer(s.LimitBuffer).
		Pack(s.T().Context(), s.client)
	s.Require().NoError(err)
	s.Require().Equal(core.SetCodeTxType, txn.Type())
	s.Require().NoError(txn.Sign(s.key))
	raw, err := txn.EncodeRLP()
	s.Require().NoError(err)
	hash, err := Broadcast(s.T().Context(), s.client, raw)
	s.Require().NoError(err)
	s.T().Logf("Test02PackSignBroadcastTokenTransfer: sent type-4 token transfer (gas limit %d), tx %s", txn.Gas(), hash)

	receipt, err := WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "set code transaction reverted")
	s.Require().True(hash.Equal(receipt.TxHash))
	s.mu.Lock()
	s.authorities = append(s.authorities, authorityKey)
	s.mu.Unlock()
	s.Require().LessOrEqual(receipt.GasUsed, txn.Gas())
	price, ok := new(big.Int).SetString(strings.TrimPrefix(receipt.Raw["effectiveGasPrice"].(string), "0x"), 16)
	s.Require().True(ok)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(price, new(big.Int).SetUint64(receipt.GasUsed)))
	s.mu.Unlock()

	s.Require().Len(receipt.Logs, 1)
	log := receipt.Logs[0]
	s.Require().True(s.token.Equal(log.Address))
	s.Require().Len(log.Topics, 3)
	s.Require().True(core.Keccak256([]byte("Transfer(address,address,uint256)")).Equal(log.Topics[0]))
	s.Require().True(types.NewHashFromBytes(append(make([]byte, 12), s.from.Bytes()...)).Equal(log.Topics[1]))
	s.Require().True(types.NewHashFromBytes(append(make([]byte, 12), s.recipient.Bytes()...)).Equal(log.Topics[2]))
	wantData := make([]byte, 32)
	s.TransferValue.FillBytes(wantData)
	s.Require().Equal(wantData, log.Data)

	s.T().Logf("SetCode Test02PackSignBroadcastTokenTransfer Result")
	s.T().Logf("\tTransferred %s of %s to %s", s.TransferValue, s.token, s.recipient)
	s.T().Logf("\tTx: %s (block %d, gas %d / limit %d)", hash, receipt.BlockNumber, receipt.GasUsed, txn.Gas())
}

func (s *SetCodeTxSuite) Test03EstimateIncludesAuthorizationAndDelegatedCode() {
	authorityKey, err := types.GeneratePrivateKey()
	s.Require().NoError(err)
	authority := core.PubkeyToAddress(authorityKey.PublicKey())
	auth := core.Authorization{ChainID: s.ChainID, Address: s.multicall, Nonce: 0}
	s.Require().NoError(auth.Sign(authorityKey))

	data, err := abi.NewFunction("getBlockNumber", nil, abi.NewTypes(abi.Uint256)).EncodeCall()
	s.Require().NoError(err)

	var plainHex string
	s.Require().NoError(s.client.Call(s.T().Context(), rpc.ETHEstimateGas("plain", map[string]any{
		"from": s.from.String(),
		"to":   authority.String(),
		"data": encoding.Hex.EncodePrefixed(data),
	}, rpc.BlockTagLatest, nil, &plainHex)))
	plain, ok := new(big.Int).SetString(strings.TrimPrefix(plainHex, "0x"), 16)
	s.Require().True(ok)

	estimated, err := NewSetCodeTxConfig(s.from, authority, big.NewInt(0), data, []core.Authorization{auth}).Pack(s.T().Context(), s.client)
	s.Require().NoError(err)
	s.Require().Greater(estimated.Gas(), plain.Uint64(), "estimate with authorizationList must include the authorization cost")

	txn, err := NewSetCodeTxConfig(s.from, authority, big.NewInt(0), data, []core.Authorization{auth}).
		WithGasLimit(estimated.Gas()*s.LimitBuffer/100).
		Pack(s.T().Context(), s.client)
	s.Require().NoError(err)
	s.Require().NoError(txn.Sign(s.key))
	raw, err := txn.EncodeRLP()
	s.Require().NoError(err)
	hash, err := Broadcast(s.T().Context(), s.client, raw)
	s.Require().NoError(err)
	s.T().Logf("Test03EstimateIncludesAuthorizationAndDelegatedCode: sent call into delegated authority, tx %s", hash)

	receipt, err := WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "set code transaction reverted")
	s.Require().Greater(receipt.GasUsed, plain.Uint64(), "executed gas should include the authorization cost")
	s.Require().LessOrEqual(receipt.GasUsed, txn.Gas())
	price, ok := new(big.Int).SetString(strings.TrimPrefix(receipt.Raw["effectiveGasPrice"].(string), "0x"), 16)
	s.Require().True(ok)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(price, new(big.Int).SetUint64(receipt.GasUsed)))
	s.mu.Unlock()

	wantCode := "0xef0100" + strings.TrimPrefix(strings.ToLower(s.multicall.String()), "0x")
	var codeHex string
	s.Require().Eventually(func() bool {
		return s.client.Call(s.T().Context(), rpc.ETHGetCode("code", authority.String(), fmt.Sprintf("0x%x", receipt.BlockNumber), &codeHex)) == nil
	}, 20*time.Second, s.ReceiptPollInterval, "authority code not readable at the receipt block")
	s.Require().Equal(wantCode, strings.ToLower(codeHex))
	s.mu.Lock()
	s.authorities = append(s.authorities, authorityKey)
	s.mu.Unlock()

	s.T().Logf("SetCode Test03EstimateIncludesAuthorizationAndDelegatedCode Result")
	s.T().Logf("\tEstimate without list %d, with list %d, sent limit %d, used %d", plain.Uint64(), estimated.Gas(), txn.Gas(), receipt.GasUsed)
	s.T().Logf("\tTx: %s (block %d)", hash, receipt.BlockNumber)
}

func (s *SetCodeTxSuite) Test04RevokeDelegation() {
	authorityKey, err := types.GeneratePrivateKey()
	s.Require().NoError(err)
	authority := core.PubkeyToAddress(authorityKey.PublicKey())

	delegate := core.Authorization{ChainID: s.ChainID, Address: s.multicall, Nonce: 0}
	s.Require().NoError(delegate.Sign(authorityKey))
	delegateHash, err := Submit(s.T().Context(), s.client, NewSetCodeTxConfig(s.from, s.recipient, big.NewInt(0), nil, []core.Authorization{delegate}).WithGasLimitBuffer(s.LimitBuffer), s.key)
	s.Require().NoError(err)
	delegateReceipt, err := WaitForReceipt(s.T().Context(), s.client, delegateHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(delegateReceipt.Status, "delegation transaction reverted")
	delegatePrice, ok := new(big.Int).SetString(strings.TrimPrefix(delegateReceipt.Raw["effectiveGasPrice"].(string), "0x"), 16)
	s.Require().True(ok)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(delegatePrice, new(big.Int).SetUint64(delegateReceipt.GasUsed)))
	s.mu.Unlock()

	wantCode := "0xef0100" + strings.TrimPrefix(strings.ToLower(s.multicall.String()), "0x")
	var delegatedHex string
	s.Require().Eventually(func() bool {
		return s.client.Call(s.T().Context(), rpc.ETHGetCode("code", authority.String(), fmt.Sprintf("0x%x", delegateReceipt.BlockNumber), &delegatedHex)) == nil
	}, 20*time.Second, s.ReceiptPollInterval, "authority code not readable at the delegation block")
	s.Require().Equal(wantCode, strings.ToLower(delegatedHex))

	revoke := core.Authorization{ChainID: s.ChainID, Address: types.ZeroAddress(), Nonce: 1}
	s.Require().NoError(revoke.Sign(authorityKey))
	var revokeHash *types.Hash
	s.Require().Eventually(func() bool {
		revokeHash, err = Submit(s.T().Context(), s.client, NewSetCodeTxConfig(s.from, s.recipient, big.NewInt(0), nil, []core.Authorization{revoke}).WithGasLimitBuffer(s.LimitBuffer), s.key)
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "revocation not accepted by the estimating node")
	s.T().Logf("Test04RevokeDelegation: delegated in %s, revoking in %s", delegateHash, revokeHash)

	revokeReceipt, err := WaitForReceipt(s.T().Context(), s.client, revokeHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(revokeReceipt.Status, "revocation transaction reverted")
	revokePrice, ok := new(big.Int).SetString(strings.TrimPrefix(revokeReceipt.Raw["effectiveGasPrice"].(string), "0x"), 16)
	s.Require().True(ok)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(revokePrice, new(big.Int).SetUint64(revokeReceipt.GasUsed)))
	s.mu.Unlock()

	atBlock := fmt.Sprintf("0x%x", revokeReceipt.BlockNumber)
	var revokedHex, nonceHex string
	s.Require().Eventually(func() bool {
		return s.client.Call(s.T().Context(), rpc.ETHGetCode("code", authority.String(), atBlock, &revokedHex)) == nil
	}, 20*time.Second, s.ReceiptPollInterval, "authority code not readable at the revocation block")
	s.Require().Equal("0x", revokedHex, "delegation to the zero address should clear the code")
	s.Require().NoError(s.client.Call(s.T().Context(), rpc.ETHGetTransactionCount("nonce", authority.String(), atBlock, &nonceHex)))
	s.Require().Equal("0x2", nonceHex, "each applied authorization advances the authority nonce")

	s.T().Logf("SetCode Test04RevokeDelegation Result")
	s.T().Logf("\tAuthority %s code %s -> %s, nonce %s", authority, delegatedHex, revokedHex, nonceHex)
	s.T().Logf("\tTx: %s (block %d, gas %d)", revokeHash, revokeReceipt.BlockNumber, revokeReceipt.GasUsed)
}

func (s *SetCodeTxSuite) Test05AuthorizationForOtherChainIsSkipped() {
	authorityKey, err := types.GeneratePrivateKey()
	s.Require().NoError(err)
	authority := core.PubkeyToAddress(authorityKey.PublicKey())
	auth := core.Authorization{ChainID: new(big.Int).Add(s.ChainID, big.NewInt(1)), Address: s.multicall, Nonce: 0}
	s.Require().NoError(auth.Sign(authorityKey))

	_, err = NewSetCodeTxConfig(s.from, s.recipient, big.NewInt(0), nil, []core.Authorization{auth}).WithGasLimitBuffer(s.LimitBuffer).Pack(s.T().Context(), s.client)
	s.Require().ErrorIs(err, ErrAuthorizationChainID, "Pack must reject an authorization for another chain")

	placeholder := core.Authorization{ChainID: s.ChainID, Address: s.multicall, Nonce: 0}
	s.Require().NoError(placeholder.Sign(authorityKey))
	packed, err := NewSetCodeTxConfig(s.from, s.recipient, big.NewInt(0), nil, []core.Authorization{placeholder}).WithGasLimitBuffer(s.LimitBuffer).Pack(s.T().Context(), s.client)
	s.Require().NoError(err)
	setCodeTx, ok := packed.(*core.SetCodeTx)
	s.Require().True(ok)
	setCodeTx.AuthList = []core.Authorization{auth}
	s.Require().NoError(setCodeTx.Sign(s.key))
	raw, err := setCodeTx.EncodeRLP()
	s.Require().NoError(err)
	hash, err := Broadcast(s.T().Context(), s.client, raw)
	s.Require().NoError(err)
	s.T().Logf("Test05AuthorizationForOtherChainIsSkipped: sent type-4 tx with a foreign-chain authorization, tx %s", hash)

	receipt, err := WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "a skipped authorization must not revert the transaction")
	price, ok := new(big.Int).SetString(strings.TrimPrefix(receipt.Raw["effectiveGasPrice"].(string), "0x"), 16)
	s.Require().True(ok)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(price, new(big.Int).SetUint64(receipt.GasUsed)))
	s.mu.Unlock()

	atBlock := fmt.Sprintf("0x%x", receipt.BlockNumber)
	var codeHex, nonceHex string
	s.Require().Eventually(func() bool {
		return s.client.Call(s.T().Context(), rpc.ETHGetCode("code", authority.String(), atBlock, &codeHex)) == nil
	}, 20*time.Second, s.ReceiptPollInterval, "authority code not readable at the receipt block")
	s.Require().Equal("0x", codeHex, "an authorization for another chain must not delegate")
	s.Require().NoError(s.client.Call(s.T().Context(), rpc.ETHGetTransactionCount("nonce", authority.String(), atBlock, &nonceHex)))
	s.Require().Equal("0x0", nonceHex, "a skipped authorization must not advance the authority nonce")

	s.T().Logf("SetCode Test05AuthorizationForOtherChainIsSkipped Result")
	s.T().Logf("\tAuthorization chainId %s on chain %s: code %s, nonce %s", auth.ChainID, s.ChainID, codeHex, nonceHex)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}
