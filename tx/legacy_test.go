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

type LegacyTxSuite struct {
	suite.Suite

	RPCURL              string
	PrivateKeyHex       string
	ChainID             *big.Int
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
	nicks     *types.Address
	token     *types.Address
	recipient *types.Address

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestLegacyTxSuite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)
	if c.Addresses.Standard.NicksFactory == "" {
		t.Skip("testenv: no Nick's factory configured for this network")
	}

	suite.Run(t, &LegacyTxSuite{
		RPCURL:              c.RPCURL,
		PrivateKeyHex:       c.ActorPrivateKeyHex,
		ChainID:             big.NewInt(int64(c.Constant.ChainID)),
		NicksFactoryHex:     c.Addresses.Standard.NicksFactory,
		NativeSymbol:        c.Constant.NativeSymbol,
		NativeDecimals:      c.Constant.NativeDecimals,
		Timeout:             10 * time.Minute,
		ReceiptPollInterval: c.Constant.ReceiptPollInterval,
		ReceiptTimeout:      c.Constant.ReceiptTimeout,
		LimitBuffer:         130,
		ByteCode:            testenv.ERC20ByteCodeHex,
		Constructor:         abi.NewFunction("constructor", abi.Types{abi.String, abi.String, abi.Uint8}, nil),
		Salt:                core.Keccak256([]byte("go-evmkit-tx-legacy")),
		MintValue:           new(big.Int).Exp(big.NewInt(10), big.NewInt(21), nil),
		TransferValue:       new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
		NativeValue:         big.NewInt(1_000_000_000_000),
	})
}

func (s *LegacyTxSuite) SetupSuite() {
	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	var err error
	s.key, err = types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.from = core.PubkeyToAddress(s.key.PublicKey())
	s.nicks, err = types.NewAddressFromHex(s.NicksFactoryHex)
	s.Require().NoError(err)
	s.T().Logf("SetupSuite: signing as %s, Nick's factory %s", s.from, s.nicks)

	runtime := fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())
	s.recipient = types.NewAddressFromBytes(core.Keccak256(fmt.Appendf(nil, "%s-recipient", runtime)).Bytes()[12:])

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)
	args, err := abi.Pack(s.Constructor.Inputs, "evmkit legacy test token", "EVLGT", uint8(18))
	s.Require().NoError(err)
	initCode := append(append([]byte{}, bytecode...), args...)
	var salt [32]byte
	copy(salt[:], core.Keccak256(runtime).Bytes())

	deployTx, err := NewLegacyTxConfig(s.from, s.nicks, big.NewInt(0), append(append([]byte{}, salt[:]...), initCode...)).
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
	var mintTx core.Transaction
	s.Require().Eventually(func() bool {
		mintTx, err = NewLegacyTxConfig(s.from, s.token, big.NewInt(0), mintData).
			WithGasLimitBuffer(s.LimitBuffer).
			Pack(s.T().Context(), s.client)
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "token not callable on the estimating node")
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

func (s *LegacyTxSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

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
			burnTx, err := NewLegacyTxConfig(s.from, s.token, big.NewInt(0), burnData).
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
		s.T().Logf("LegacyTxSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}
}

func (s *LegacyTxSuite) Test01SubmitETHTransfer() {
	hash, err := Submit(s.T().Context(), s.client, NewLegacyTxConfig(s.from, s.recipient, s.NativeValue, nil).WithGasLimitBuffer(s.LimitBuffer), s.key)
	s.Require().NoError(err)
	s.T().Logf("Test01SubmitETHTransfer: sent legacy transfer, tx %s", hash)

	receipt, err := WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "legacy transfer reverted")
	s.Require().Equal("0x0", receipt.Raw["type"])
	price, ok := new(big.Int).SetString(strings.TrimPrefix(receipt.Raw["effectiveGasPrice"].(string), "0x"), 16)
	s.Require().True(ok)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(price, new(big.Int).SetUint64(receipt.GasUsed)))
	s.mu.Unlock()

	var afterHex, beforeHex string
	s.Require().Eventually(func() bool {
		return s.client.Call(s.T().Context(), rpc.ETHGetBalance("after", s.recipient.String(), fmt.Sprintf("0x%x", receipt.BlockNumber), &afterHex)) == nil
	}, 20*time.Second, s.ReceiptPollInterval, "balance not readable at the receipt block")
	s.Require().NoError(s.client.Call(s.T().Context(), rpc.ETHGetBalance("before", s.recipient.String(), fmt.Sprintf("0x%x", receipt.BlockNumber-1), &beforeHex)))
	after, ok := new(big.Int).SetString(strings.TrimPrefix(afterHex, "0x"), 16)
	s.Require().True(ok)
	before, ok := new(big.Int).SetString(strings.TrimPrefix(beforeHex, "0x"), 16)
	s.Require().True(ok)
	s.Require().Zero(s.NativeValue.Cmp(new(big.Int).Sub(after, before)), "recipient did not receive exactly the native value")

	s.T().Logf("Legacy Test01SubmitETHTransfer Result")
	s.T().Logf("\tRecipient %s: %s -> %s", s.recipient, before, after)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *LegacyTxSuite) Test02PackSignBroadcastTokenTransfer() {
	data, err := abi.NewFunction("transfer", abi.NewTypes(abi.Address, abi.Uint256), abi.NewTypes(abi.Bool)).EncodeCall(s.recipient, s.TransferValue)
	s.Require().NoError(err)

	txn, err := NewLegacyTxConfig(s.from, s.token, big.NewInt(0), data).
		WithGasLimitBuffer(s.LimitBuffer).
		Pack(s.T().Context(), s.client)
	s.Require().NoError(err)
	s.Require().Equal(core.LegacyTxType, txn.Type())
	s.Require().NoError(txn.Sign(s.key))
	raw, err := txn.EncodeRLP()
	s.Require().NoError(err)
	hash, err := Broadcast(s.T().Context(), s.client, raw)
	s.Require().NoError(err)
	s.T().Logf("Test02PackSignBroadcastTokenTransfer: sent legacy token transfer (gas limit %d), tx %s", txn.Gas(), hash)

	receipt, err := WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "legacy token transfer reverted")
	s.Require().True(hash.Equal(receipt.TxHash))
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

	s.T().Logf("Legacy Test02PackSignBroadcastTokenTransfer Result")
	s.T().Logf("\tTransferred %s of %s to %s", s.TransferValue, s.token, s.recipient)
	s.T().Logf("\tTx: %s (block %d, gas %d / limit %d)", hash, receipt.BlockNumber, receipt.GasUsed, txn.Gas())
}

func (s *LegacyTxSuite) Test03ExplicitFieldsAreMinedAsSet() {
	var nonceHex, gasPriceHex string
	s.Require().NoError(s.client.Batch(s.T().Context(), rpc.Elements{
		rpc.ETHGetTransactionCount("nonce", s.from.String(), rpc.BlockTagPending, &nonceHex),
		rpc.ETHGasPrice("gasPrice", &gasPriceHex),
	}))
	nonce, ok := new(big.Int).SetString(strings.TrimPrefix(nonceHex, "0x"), 16)
	s.Require().True(ok)
	gasPrice, ok := new(big.Int).SetString(strings.TrimPrefix(gasPriceHex, "0x"), 16)
	s.Require().True(ok)
	gasPrice.Mul(gasPrice, big.NewInt(2))
	gasLimit := uint64(150_000)

	txn, err := NewLegacyTxConfig(s.from, s.recipient, big.NewInt(1), nil).
		WithChainID(s.ChainID).
		WithNonce(nonce.Uint64()).
		WithGasPrice(gasPrice).
		WithGasLimit(gasLimit).
		Pack(s.T().Context(), s.client)
	s.Require().NoError(err)
	s.Require().Equal(nonce.Uint64(), txn.AddressNonce())
	s.Require().Equal(gasLimit, txn.Gas())
	s.Require().Zero(new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(gasLimit)).Cmp(txn.MaxFee()))
	s.Require().NoError(txn.Sign(s.key))
	sender, err := txn.Sender()
	s.Require().NoError(err)
	s.Require().True(sender.Equal(s.from))
	raw, err := txn.EncodeRLP()
	s.Require().NoError(err)
	hash, err := Broadcast(s.T().Context(), s.client, raw)
	s.Require().NoError(err)
	s.T().Logf("Test03ExplicitFieldsAreMinedAsSet: sent legacy tx with explicit fields, tx %s", hash)

	receipt, err := WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "legacy transfer reverted")
	price, ok := new(big.Int).SetString(strings.TrimPrefix(receipt.Raw["effectiveGasPrice"].(string), "0x"), 16)
	s.Require().True(ok)
	s.Require().LessOrEqual(price.Cmp(gasPrice), 0, "effective gas price exceeds the signed gas price")
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(price, new(big.Int).SetUint64(receipt.GasUsed)))
	s.mu.Unlock()

	var mined map[string]any
	s.Require().Eventually(func() bool {
		return s.client.Call(s.T().Context(), rpc.ETHGetTransactionByHash("tx", hash.String(), &mined)) == nil && mined != nil
	}, 20*time.Second, s.ReceiptPollInterval, "mined transaction not readable")
	s.Require().Equal("0x0", mined["type"])
	s.Require().Equal(fmt.Sprintf("0x%x", nonce.Uint64()), mined["nonce"])
	s.Require().Equal(fmt.Sprintf("0x%x", gasLimit), mined["gas"])
	s.Require().Equal("0x"+gasPrice.Text(16), mined["gasPrice"])
	v, ok := new(big.Int).SetString(strings.TrimPrefix(mined["v"].(string), "0x"), 16)
	s.Require().True(ok)
	eip155 := new(big.Int).Add(new(big.Int).Mul(s.ChainID, big.NewInt(2)), big.NewInt(35))
	s.Require().True(v.Cmp(eip155) == 0 || v.Cmp(new(big.Int).Add(eip155, big.NewInt(1))) == 0, "v %s is not EIP-155 for chain %s", v, s.ChainID)

	s.T().Logf("Legacy Test03ExplicitFieldsAreMinedAsSet Result")
	s.T().Logf("\tNonce %d, gas %d, gasPrice %s, v %s", nonce.Uint64(), gasLimit, gasPrice, v)
	s.T().Logf("\tTx: %s (block %d, gas used %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *LegacyTxSuite) Test04GasLimitBufferScalesEstimate() {
	data, err := abi.NewFunction("transfer", abi.NewTypes(abi.Address, abi.Uint256), abi.NewTypes(abi.Bool)).EncodeCall(s.recipient, big.NewInt(1))
	s.Require().NoError(err)

	plain, err := NewLegacyTxConfig(s.from, s.token, big.NewInt(0), data).Pack(s.T().Context(), s.client)
	s.Require().NoError(err)
	buffered, err := NewLegacyTxConfig(s.from, s.token, big.NewInt(0), data).WithGasLimitBuffer(200).Pack(s.T().Context(), s.client)
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(buffered.Gas(), plain.Gas()*19/10)
	s.Require().LessOrEqual(buffered.Gas(), plain.Gas()*21/10)

	s.Require().NoError(buffered.Sign(s.key))
	raw, err := buffered.EncodeRLP()
	s.Require().NoError(err)
	hash, err := Broadcast(s.T().Context(), s.client, raw)
	s.Require().NoError(err)
	s.T().Logf("Test04GasLimitBufferScalesEstimate: sent buffered legacy tx, tx %s", hash)

	receipt, err := WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "buffered legacy transfer reverted")
	s.Require().LessOrEqual(receipt.GasUsed, buffered.Gas())
	price, ok := new(big.Int).SetString(strings.TrimPrefix(receipt.Raw["effectiveGasPrice"].(string), "0x"), 16)
	s.Require().True(ok)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(price, new(big.Int).SetUint64(receipt.GasUsed)))
	s.mu.Unlock()

	s.T().Logf("Legacy Test04GasLimitBufferScalesEstimate Result")
	s.T().Logf("\tEstimate %d, buffered limit %d, used %d", plain.Gas(), buffered.Gas(), receipt.GasUsed)
	s.T().Logf("\tTx: %s (block %d)", hash, receipt.BlockNumber)
}
