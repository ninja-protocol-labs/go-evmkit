package bind

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/internal/testenv"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
	"github.com/stretchr/testify/suite"
)

type BindSuite struct {
	suite.Suite

	Enable        bool
	RPCURL        string
	PrivateKeyHex string
	ToHex         string
	Name          string
	Symbol        string
	Decimals      uint8
	Value         string
	MintValue     string
	NativeSymbol  string

	ctx    context.Context
	cancel context.CancelFunc

	client    rpc.Client
	signer    core.Signer
	from      *types.Address
	to        *types.Address
	token     *types.Address
	value     *big.Int
	mintValue *big.Int

	totalGasUsed uint64
	totalFee     *big.Int
}

func TestBindSuite(t *testing.T) {
	suite.Run(t, &BindSuite{
		Enable:        false,
		RPCURL:        "",
		PrivateKeyHex: "",
		ToHex:         "",
		Name:          "evmkit bind ERC20",
		Symbol:        "EVKB",
		Decimals:      18,
		Value:         "100000",
		MintValue:     "1000000000000000000000000",
		NativeSymbol:  "",
	})
}

func (s *BindSuite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(s.T().Context(), 5*time.Minute)
	s.totalFee = big.NewInt(0)
	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())

	s.to, err = types.NewAddressFromHex(s.ToHex)
	s.Require().NoError(err)

	value, ok := new(big.Int).SetString(s.Value, 10)
	s.Require().True(ok)
	s.value = value

	mintValue, ok := new(big.Int).SetString(s.MintValue, 10)
	s.Require().True(ok)
	s.mintValue = mintValue

	bytecode, err := encoding.Hex.Decode(testenv.ERC20ByteCodeHex)
	s.Require().NoError(err)
	args, err := abi.Pack(abi.Types{abi.String, abi.String, abi.Uint8}, s.Name, s.Symbol, s.Decimals)
	s.Require().NoError(err)
	initCode := append(append([]byte{}, bytecode...), args...)

	deployTx, deployHash, err := Transact(s.ctx, s.client, &TransactOptions{Signer: s.signer, Builder: tx.NewDynamicFeeBuilder()}, nil, initCode)
	s.Require().NoError(err)

	deployReceipt, err := tx.WaitForReceipt(s.ctx, s.client, deployHash, 500*time.Millisecond, 30*time.Second)
	s.Require().NoError(err)
	s.Require().True(deployReceipt.Status, "deploy transaction reverted")
	deployReceiptPriceHex, isString := deployReceipt.Raw["effectiveGasPrice"].(string)
	s.Require().True(isString, "receipt has no effectiveGasPrice")
	deployReceiptPrice, ok := new(big.Int).SetString(strings.TrimPrefix(deployReceiptPriceHex, "0x"), 16)
	s.Require().True(ok)
	s.totalGasUsed += deployReceipt.GasUsed
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(new(big.Int).SetUint64(deployReceipt.GasUsed), deployReceiptPrice))

	s.token, err = core.CreateAddress(s.from, deployTx.AddressNonce())
	s.Require().NoError(err)

	balanceOfFn := abi.NewFunction("balanceOf", abi.Types{abi.Address}, abi.Types{abi.Uint256})
	balanceOfCalldata, err := balanceOfFn.EncodeCall(s.from)
	s.Require().NoError(err)
	s.Require().Eventually(func() bool {
		out, err := Call(s.ctx, s.client, nil, s.token, balanceOfCalldata)
		return err == nil && len(out) == 32
	}, 10*time.Second, 500*time.Millisecond, "no code visible at deployed token %s", s.token)

	mintFn := abi.NewFunction("mint", abi.Types{abi.Address, abi.Uint256}, nil)
	mintCalldata, err := mintFn.EncodeCall(s.from, s.mintValue)
	s.Require().NoError(err)

	_, mintHash, err := Transact(s.ctx, s.client, &TransactOptions{Signer: s.signer, Builder: tx.NewDynamicFeeBuilder()}, s.token, mintCalldata)
	s.Require().NoError(err)

	mintReceipt, err := tx.WaitForReceipt(s.ctx, s.client, mintHash, 500*time.Millisecond, 30*time.Second)
	s.Require().NoError(err)
	s.Require().True(mintReceipt.Status, "mint transaction reverted")
	mintReceiptPriceHex, isString := mintReceipt.Raw["effectiveGasPrice"].(string)
	s.Require().True(isString, "receipt has no effectiveGasPrice")
	mintReceiptPrice, ok := new(big.Int).SetString(strings.TrimPrefix(mintReceiptPriceHex, "0x"), 16)
	s.Require().True(ok)
	s.totalGasUsed += mintReceipt.GasUsed
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(new(big.Int).SetUint64(mintReceipt.GasUsed), mintReceiptPrice))

	s.T().Logf("SetupSuite: deployed %s (nonce %d, tx %s), minted %s to %s (tx %s)", s.token, deployTx.AddressNonce(), deployHash, s.mintValue, s.from, mintHash)
}

func (s *BindSuite) TearDownSuite() {
	if s.token != nil {
		balanceOfFn := abi.NewFunction("balanceOf", abi.Types{abi.Address}, abi.Types{abi.Uint256})
		burnFn := abi.NewFunction("burn", abi.Types{abi.Address, abi.Uint256}, nil)

		for _, holder := range []*types.Address{s.from, s.to} {
			calldata, err := balanceOfFn.EncodeCall(holder)
			s.Require().NoError(err)
			out, err := Call(s.ctx, s.client, nil, s.token, calldata)
			s.Require().NoError(err)
			balance, err := balanceOfFn.DecodeSingleReturn[*big.Int](out)
			s.Require().NoError(err)
			if balance.Sign() == 0 {
				continue
			}

			burnCalldata, err := burnFn.EncodeCall(holder, balance)
			s.Require().NoError(err)
			_, burnHash, err := Transact(s.ctx, s.client, &TransactOptions{Signer: s.signer, Builder: tx.NewDynamicFeeBuilder()}, s.token, burnCalldata)
			s.Require().NoError(err)

			burnReceipt, err := tx.WaitForReceipt(s.ctx, s.client, burnHash, 500*time.Millisecond, 30*time.Second)
			s.Require().NoError(err)
			s.Require().True(burnReceipt.Status, "burn transaction reverted")
			burnReceiptPriceHex, isString := burnReceipt.Raw["effectiveGasPrice"].(string)
			s.Require().True(isString, "receipt has no effectiveGasPrice")
			burnReceiptPrice, ok := new(big.Int).SetString(strings.TrimPrefix(burnReceiptPriceHex, "0x"), 16)
			s.Require().True(ok)
			s.totalGasUsed += burnReceipt.GasUsed
			s.totalFee.Add(s.totalFee, new(big.Int).Mul(new(big.Int).SetUint64(burnReceipt.GasUsed), burnReceiptPrice))

			s.T().Logf("TearDownSuite: burned %s of %s from %s (tx %s)", balance, s.token, holder, burnHash)
		}
	}

	if s.totalFee != nil {
		base := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
		whole, frac := new(big.Int), new(big.Int)
		whole.DivMod(s.totalFee, base, frac)

		fracStr := strings.TrimRight(fmt.Sprintf("%018s", frac.String()), "0")
		decimal := whole.String()
		if fracStr != "" {
			decimal += "." + fracStr
		}

		s.T().Logf("BindSuite total gas used: %d", s.totalGasUsed)
		s.T().Logf("BindSuite total fee spent: %s %s (%s wei)", decimal, s.NativeSymbol, s.totalFee)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *BindSuite) Test01CallMetadata() {
	nameFn := abi.NewFunction("name", nil, abi.Types{abi.String})
	symbolFn := abi.NewFunction("symbol", nil, abi.Types{abi.String})
	decimalsFn := abi.NewFunction("decimals", nil, abi.Types{abi.Uint8})

	nameOut, err := Call(s.ctx, s.client, nil, s.token, nameFn.SelectorBytes())
	s.Require().NoError(err)
	name, err := nameFn.DecodeSingleReturn[string](nameOut)
	s.Require().NoError(err)
	s.Require().Equal(s.Name, name)

	symbolOut, err := Call(s.ctx, s.client, &CallOptions{Block: rpc.BlockTagLatest}, s.token, symbolFn.SelectorBytes())
	s.Require().NoError(err)
	symbol, err := symbolFn.DecodeSingleReturn[string](symbolOut)
	s.Require().NoError(err)
	s.Require().Equal(s.Symbol, symbol)

	decimalsOut, err := Call(s.ctx, s.client, &CallOptions{From: s.to}, s.token, decimalsFn.SelectorBytes())
	s.Require().NoError(err)
	decimals, err := decimalsFn.DecodeSingleReturn[uint8](decimalsOut)
	s.Require().NoError(err)
	s.Require().Equal(s.Decimals, decimals)

	s.T().Logf("Bind %s Test01CallMetadata Result", s.token)
	s.T().Logf("\tName: %s", name)
	s.T().Logf("\tSymbol: %s", symbol)
	s.T().Logf("\tDecimals: %d", decimals)
}

func (s *BindSuite) Test02CallSimulatesTransfer() {
	transferFn := abi.NewFunction("transfer", abi.Types{abi.Address, abi.Uint256}, abi.Types{abi.Bool})

	calldata, err := transferFn.EncodeCall(s.to, s.value)
	s.Require().NoError(err)
	out, err := Call(s.ctx, s.client, &CallOptions{From: s.from}, s.token, calldata)
	s.Require().NoError(err)
	ok, err := transferFn.DecodeSingleReturn[bool](out)
	s.Require().NoError(err)
	s.Require().True(ok)

	huge := new(big.Int).Add(s.mintValue, big.NewInt(1))
	hugeCalldata, err := transferFn.EncodeCall(s.to, huge)
	s.Require().NoError(err)
	out, err = Call(s.ctx, s.client, &CallOptions{From: s.from}, s.token, hugeCalldata)
	s.Require().Error(err)
	s.Require().Nil(out)

	var rpcErr *rpc.ResponseError
	isRPCErr := errors.As(err, &rpcErr)
	s.Require().True(isRPCErr, "error must be the raw *rpc.ResponseError, got %T", err)
	s.Require().NotEmpty(rpcErr.Data)
	var revertHex string
	s.Require().NoError(json.Unmarshal(rpcErr.Data, &revertHex))
	revert, err := encoding.Hex.Decode(revertHex)
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(revert), 4)
	s.Require().Equal([]byte{0xe4, 0x50, 0xd3, 0x8c}, revert[:4])

	s.T().Logf("Bind %s Test02CallSimulatesTransfer Result", s.token)
	s.T().Logf("\tFrom: %s", s.from)
	s.T().Logf("\tTo: %s", s.to)
	s.T().Logf("\tValue: %s -> %v", s.value, ok)
	s.T().Logf("\tHuge: %s -> %v", huge, err)
}

func (s *BindSuite) Test03TransactTransfer() {
	balanceOfFn := abi.NewFunction("balanceOf", abi.Types{abi.Address}, abi.Types{abi.Uint256})
	transferFn := abi.NewFunction("transfer", abi.Types{abi.Address, abi.Uint256}, abi.Types{abi.Bool})

	balanceCalldata, err := balanceOfFn.EncodeCall(s.to)
	s.Require().NoError(err)
	beforeOut, err := Call(s.ctx, s.client, nil, s.token, balanceCalldata)
	s.Require().NoError(err)
	before, err := balanceOfFn.DecodeSingleReturn[*big.Int](beforeOut)
	s.Require().NoError(err)

	calldata, err := transferFn.EncodeCall(s.to, s.value)
	s.Require().NoError(err)
	signed, hash, err := Transact(s.ctx, s.client, &TransactOptions{Signer: s.signer, Builder: tx.NewDynamicFeeBuilder()}, s.token, calldata)
	s.Require().NoError(err)

	signedHash, err := signed.Hash()
	s.Require().NoError(err)
	s.Require().True(hash.Equal(signedHash))
	sender, err := signed.Sender()
	s.Require().NoError(err)
	s.Require().True(sender.Equal(s.from))

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, 500*time.Millisecond, 30*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transfer transaction reverted")
	receiptPriceHex, isString := receipt.Raw["effectiveGasPrice"].(string)
	s.Require().True(isString, "receipt has no effectiveGasPrice")
	receiptPrice, ok := new(big.Int).SetString(strings.TrimPrefix(receiptPriceHex, "0x"), 16)
	s.Require().True(ok)
	s.totalGasUsed += receipt.GasUsed
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(new(big.Int).SetUint64(receipt.GasUsed), receiptPrice))

	s.Require().Len(receipt.Logs, 1)
	log := receipt.Logs[0]
	s.Require().True(log.Address.Equal(s.token))
	s.Require().Len(log.Topics, 3)
	s.Require().True(log.Topics[0].Equal(core.Keccak256([]byte("Transfer(address,address,uint256)"))))
	s.Require().True(types.NewAddressFromBytes(log.Topics[1].Bytes()[12:]).Equal(s.from))
	s.Require().True(types.NewAddressFromBytes(log.Topics[2].Bytes()[12:]).Equal(s.to))
	s.Require().Zero(s.value.Cmp(new(big.Int).SetBytes(log.Data)))

	want := new(big.Int).Add(before, s.value)
	var after *big.Int
	s.Require().Eventually(func() bool {
		out, err := Call(s.ctx, s.client, nil, s.token, balanceCalldata)
		if err != nil {
			return false
		}
		after, err = balanceOfFn.DecodeSingleReturn[*big.Int](out)
		return err == nil && want.Cmp(after) == 0
	}, 10*time.Second, 500*time.Millisecond, "balance not visible after transfer")

	s.T().Logf("Bind %s Test03TransactTransfer Result", s.token)
	s.T().Logf("\tFrom: %s", s.from)
	s.T().Logf("\tTo: %s", s.to)
	s.T().Logf("\tValue: %s", s.value)
	s.T().Logf("\tNonce: %d", signed.AddressNonce())
	s.T().Logf("\tGas: %d", signed.Gas())
	s.T().Logf("\tMaxFee: %s", signed.MaxFee())
	s.T().Logf("\tTx: %s", hash)
	s.T().Logf("\tBalance")
	s.T().Logf("\t\tBefore: %s", before)
	s.T().Logf("\t\tAfter: %s", after)
}

func (s *BindSuite) Test04CallUnknownSelector() {
	out, err := Call(s.ctx, s.client, &CallOptions{From: s.from}, s.token, []byte{0xde, 0xad, 0xbe, 0xef})
	s.Require().Error(err)
	s.Require().Nil(out)

	var rpcErr *rpc.ResponseError
	isRPCErr := errors.As(err, &rpcErr)
	s.Require().True(isRPCErr, "error must be the raw *rpc.ResponseError, got %T", err)

	s.T().Logf("Bind %s Test04CallUnknownSelector Result", s.token)
	s.T().Logf("\tCode: %d", rpcErr.Code)
	s.T().Logf("\tMessage: %s", rpcErr.Message)
	s.T().Logf("\tData: %s", rpcErr.Data)
}

func (s *BindSuite) Test05TransactRevertsOnChain() {
	transferFn := abi.NewFunction("transfer", abi.Types{abi.Address, abi.Uint256}, abi.Types{abi.Bool})

	huge := new(big.Int).Add(s.mintValue, big.NewInt(1))
	calldata, err := transferFn.EncodeCall(s.to, huge)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.Limit(100_000)

	signed, hash, err := Transact(s.ctx, s.client, &TransactOptions{Signer: s.signer, Builder: builder}, s.token, calldata)
	s.Require().NoError(err)
	s.Require().Equal(uint64(100_000), signed.Gas())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, 500*time.Millisecond, 30*time.Second)
	s.Require().NoError(err)
	s.Require().False(receipt.Status, "oversized transfer must revert on chain")
	receiptPriceHex, isString := receipt.Raw["effectiveGasPrice"].(string)
	s.Require().True(isString, "receipt has no effectiveGasPrice")
	receiptPrice, ok := new(big.Int).SetString(strings.TrimPrefix(receiptPriceHex, "0x"), 16)
	s.Require().True(ok)
	s.totalGasUsed += receipt.GasUsed
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(new(big.Int).SetUint64(receipt.GasUsed), receiptPrice))
	s.Require().Empty(receipt.Logs)

	s.T().Logf("Bind %s Test05TransactRevertsOnChain Result", s.token)
	s.T().Logf("\tValue: %s", huge)
	s.T().Logf("\tNonce: %d", signed.AddressNonce())
	s.T().Logf("\tTx: %s", hash)
}
