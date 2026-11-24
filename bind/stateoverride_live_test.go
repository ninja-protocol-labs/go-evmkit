package bind

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
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
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

type StateOverrideSuite struct {
	suite.Suite

	RPCURL              string
	PrivateKeyHex       string
	NativeSymbol        string
	NativeDecimals      int
	Timeout             time.Duration
	ReceiptPollInterval time.Duration
	ReceiptTimeout      time.Duration
	ByteCode            string
	Name                string
	Symbol              string
	Decimals            uint8
	MintValue           *big.Int
	Amount              *big.Int
	RecipientSeed       string
	LimitBufferPct      uint64

	client    rpc.Client
	signer    core.Signer
	from      *types.Address
	token     *types.Address
	recipient *types.Address

	balanceSlot   [32]byte
	allowanceSlot [32]byte
	amountWord    [32]byte

	approveData      []byte
	transferFromData []byte
	allowanceData    []byte
	recipientBalance []byte

	mu       *sync.Mutex
	totalGas uint64
	totalFee *big.Int
}

func TestStateOverrideSuite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)

	suite.Run(t, &StateOverrideSuite{
		RPCURL:              c.RPCURL,
		PrivateKeyHex:       c.ActorPrivateKeyHex,
		NativeSymbol:        c.Constant.NativeSymbol,
		NativeDecimals:      c.Constant.NativeDecimals,
		Timeout:             5 * time.Minute,
		ReceiptPollInterval: c.Constant.ReceiptPollInterval,
		ReceiptTimeout:      c.Constant.ReceiptTimeout,
		ByteCode:            testenv.ERC20ByteCodeHex,
		Name:                "evmkit bind override token",
		Symbol:              "EVKBO",
		Decimals:            18,
		MintValue:           new(big.Int).Mul(big.NewInt(1_000_000), big.NewInt(1_000_000_000_000_000_000)),
		Amount:              new(big.Int).Mul(big.NewInt(250), big.NewInt(1_000_000_000_000_000_000)),
		RecipientSeed:       "go-evmkit-bind-stateoverride-recipient",
		LimitBufferPct:      130,
	})
}

func (s *StateOverrideSuite) SetupSuite() {
	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())
	s.recipient = types.NewAddressFromBytes(core.Keccak256([]byte(s.RecipientSeed)).Bytes())
	s.T().Logf("SetupSuite: signing as %s, recipient %s", s.from, s.recipient)

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)
	args, err := abi.Pack(abi.Types{abi.String, abi.String, abi.Uint8}, s.Name, s.Symbol, s.Decimals)
	s.Require().NoError(err)
	builder := tx.NewDynamicFeeBuilder()
	builder.LimitBuffer(s.LimitBufferPct)
	deployTx, deployHash, err := Transact(
		s.T().Context(),
		s.client,
		NewTransactOptions(s.signer, builder),
		nil,
		append(bytecode, args...),
	)
	s.Require().NoError(err)
	deployReceipt, err := tx.WaitForReceipt(
		s.T().Context(),
		s.client,
		deployHash,
		s.ReceiptPollInterval,
		s.ReceiptTimeout,
	)
	s.Require().NoError(err)
	s.Require().True(deployReceipt.Status, "deploy reverted")
	s.token, err = core.CreateAddress(s.from, deployTx.AddressNonce())
	s.Require().NoError(err)
	deployPriceHex, ok := deployReceipt.Raw["effectiveGasPrice"].(string)
	s.Require().True(ok)
	deployPrice, ok := new(big.Int).SetString(strings.TrimPrefix(deployPriceHex, "0x"), 16)
	s.Require().True(ok)
	s.totalGas += deployReceipt.GasUsed
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(new(big.Int).SetUint64(deployReceipt.GasUsed), deployPrice))
	s.T().Logf("SetupSuite: deployed %s %s (tx %s)", s.Symbol, s.token, deployHash)
	s.T().Logf("SetupSuite: deploy block %d, gas %d", deployReceipt.BlockNumber, deployReceipt.GasUsed)

	mintData, err := abi.NewFunction("mint", abi.Types{abi.Address, abi.Uint256}, nil).EncodeCall(s.from, s.MintValue)
	s.Require().NoError(err)
	_, mintHash, err := Transact(s.T().Context(), s.client, NewTransactOptions(s.signer, builder), s.token, mintData)
	s.Require().NoError(err)
	mintReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, mintHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(mintReceipt.Status, "mint reverted")
	mintPrice, ok := new(big.Int).SetString(strings.TrimPrefix(mintReceipt.Raw["effectiveGasPrice"].(string), "0x"), 16)
	s.Require().True(ok)
	s.totalGas += mintReceipt.GasUsed
	s.totalFee.Add(s.totalFee, new(big.Int).Mul(new(big.Int).SetUint64(mintReceipt.GasUsed), mintPrice))
	s.T().Logf("SetupSuite: minted %s %s to %s", s.MintValue, s.Symbol, s.from)
	s.T().Logf("SetupSuite: mint tx %s, block %d", mintHash, mintReceipt.BlockNumber)

	balanceKey, err := abi.Pack(abi.Types{abi.Address, abi.Uint256}, s.from, big.NewInt(0))
	s.Require().NoError(err)
	copy(s.balanceSlot[:], core.Keccak256(balanceKey).Bytes())
	ownerKey, err := abi.Pack(abi.Types{abi.Address, abi.Uint256}, s.from, big.NewInt(1))
	s.Require().NoError(err)
	allowanceKey, err := abi.Pack(
		abi.Types{abi.Address, abi.Uint256},
		s.from,
		new(big.Int).SetBytes(core.Keccak256(ownerKey).Bytes()),
	)
	s.Require().NoError(err)
	copy(s.allowanceSlot[:], core.Keccak256(allowanceKey).Bytes())
	s.Amount.FillBytes(s.amountWord[:])
	s.T().Logf("SetupSuite: balance slot %x", s.balanceSlot)
	s.T().Logf("SetupSuite: allowance slot %x", s.allowanceSlot)

	s.approveData, err = abi.NewFunction("approve", abi.Types{abi.Address, abi.Uint256}, abi.Types{abi.Bool}).
		EncodeCall(s.from, s.Amount)
	s.Require().NoError(err)
	s.transferFromData, err = abi.NewFunction(
		"transferFrom",
		abi.Types{abi.Address, abi.Address, abi.Uint256},
		abi.Types{abi.Bool},
	).EncodeCall(s.from, s.recipient, s.Amount)
	s.Require().NoError(err)
	s.allowanceData, err = abi.NewFunction("allowance", abi.Types{abi.Address, abi.Address}, abi.Types{abi.Uint256}).
		EncodeCall(s.from, s.from)
	s.Require().NoError(err)
	s.recipientBalance, err = abi.NewFunction("balanceOf", abi.Types{abi.Address}, abi.Types{abi.Uint256}).
		EncodeCall(s.recipient)
	s.Require().NoError(err)

	s.Require().Eventually(func() bool {
		out, err := Call(s.T().Context(), s.client, nil, s.token, s.allowanceData)
		return err == nil && len(out) == 32
	}, 30*time.Second, 500*time.Millisecond, "no code visible at deployed token %s", s.token)
}

func (s *StateOverrideSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	if s.token != nil {
		balanceOfFn := abi.NewFunction("balanceOf", abi.Types{abi.Address}, abi.Types{abi.Uint256})
		burnFn := abi.NewFunction("burn", abi.Types{abi.Address, abi.Uint256}, nil)

		for _, holder := range []*types.Address{s.from, s.recipient} {
			balanceData, err := balanceOfFn.EncodeCall(holder)
			s.Require().NoError(err)
			out, err := Call(ctx, s.client, nil, s.token, balanceData)
			s.Require().NoError(err)
			balance := new(big.Int).SetBytes(out)
			if balance.Sign() == 0 {
				continue
			}

			burnData, err := burnFn.EncodeCall(holder, balance)
			s.Require().NoError(err)
			builder := tx.NewDynamicFeeBuilder()
			builder.LimitBuffer(s.LimitBufferPct)
			_, burnHash, err := Transact(ctx, s.client, NewTransactOptions(s.signer, builder), s.token, burnData)
			s.Require().NoError(err)
			burnReceipt, err := tx.WaitForReceipt(ctx, s.client, burnHash, s.ReceiptPollInterval, s.ReceiptTimeout)
			s.Require().NoError(err)
			s.Require().True(burnReceipt.Status, "burn transaction reverted")
			burnPriceHex, ok := burnReceipt.Raw["effectiveGasPrice"].(string)
			s.Require().True(ok)
			burnPrice, ok := new(big.Int).SetString(strings.TrimPrefix(burnPriceHex, "0x"), 16)
			s.Require().True(ok)

			s.mu.Lock()
			s.totalGas += burnReceipt.GasUsed
			s.totalFee.Add(s.totalFee, new(big.Int).Mul(new(big.Int).SetUint64(burnReceipt.GasUsed), burnPrice))
			s.mu.Unlock()

			s.T().Logf("TearDownSuite: burned %s %s from %s", balance, s.Symbol, holder)
			s.T().Logf("TearDownSuite: burn tx %s", burnHash)
			s.T().Logf("TearDownSuite: burn block %d, gas %d", burnReceipt.BlockNumber, burnReceipt.GasUsed)
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

		s.T().Logf("StateOverrideSuite total gas used: %d", s.totalGas)
		s.T().Logf("StateOverrideSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}
}

func (s *StateOverrideSuite) Test01CallReadsAllowanceThroughOverride() {
	plain, err := Call(s.T().Context(), s.client, nil, s.token, s.allowanceData)
	s.Require().NoError(err)
	s.Require().Zero(new(big.Int).SetBytes(plain).Sign())

	override := rpc.NewStateOverride().WithStateDiff(s.token, s.allowanceSlot, s.amountWord)
	opts := NewCallOptions().WithStateOverride(override)
	overridden, err := Call(s.T().Context(), s.client, opts, s.token, s.allowanceData)
	s.Require().NoError(err)
	s.Require().Zero(s.Amount.Cmp(new(big.Int).SetBytes(overridden)))

	overrideJSON, err := override.MarshalJSON()
	s.Require().NoError(err)
	s.T().Logf("Bind token %s Test01CallReadsAllowanceThroughOverride Result", s.token)
	s.T().Logf("\tallowance(%s, %s) without override: %s", s.from, s.from, new(big.Int).SetBytes(plain))
	s.T().Logf("\toverride: %s", overrideJSON)
	s.T().Logf("\tallowance(%s, %s) with override: %s", s.from, s.from, new(big.Int).SetBytes(overridden))
}

func (s *StateOverrideSuite) Test02CallTransferFromNeedsAllowanceOverride() {
	plainOpts := NewCallOptions().WithFrom(s.from)
	_, plainErr := Call(s.T().Context(), s.client, plainOpts, s.token, s.transferFromData)
	s.Require().Error(plainErr)

	override := rpc.NewStateOverride().WithStateDiff(s.token, s.allowanceSlot, s.amountWord)
	opts := NewCallOptions().WithFrom(s.from).WithStateOverride(override)
	out, err := Call(s.T().Context(), s.client, opts, s.token, s.transferFromData)
	s.Require().NoError(err)
	s.Require().Equal(1, new(big.Int).SetBytes(out).Cmp(big.NewInt(0)))

	overrideJSON, err := override.MarshalJSON()
	s.Require().NoError(err)
	s.T().Logf("Bind token %s Test02CallTransferFromNeedsAllowanceOverride Result", s.token)
	s.T().Logf("\ttransferFrom(%s, %s, %s) from %s", s.from, s.recipient, s.Amount, s.from)
	s.T().Logf("\twithout override: %v", plainErr)
	s.T().Logf("\toverride: %s", overrideJSON)
	s.T().Logf("\twith override: %x", out)
}

func (s *StateOverrideSuite) Test03PackTransferFromFailsWithoutOverride() {
	builder := tx.NewDynamicFeeBuilder()
	_, err := builder.Build(s.from, s.token, big.NewInt(0), s.transferFromData).Pack(s.T().Context(), s.client)
	s.Require().Error(err)

	s.T().Logf("Bind token %s Test03PackTransferFromFailsWithoutOverride Result", s.token)
	s.T().Logf("\tPack(transferFrom) without override: %v", err)
}

func (s *StateOverrideSuite) Test04PackTransferFromEstimatesWithOverride() {
	override := rpc.NewStateOverride().WithStateDiff(s.token, s.allowanceSlot, s.amountWord)
	builder := tx.NewDynamicFeeBuilder()
	builder.StateOverride(override)
	packed, err := builder.Build(s.from, s.token, big.NewInt(0), s.transferFromData).Pack(s.T().Context(), s.client)
	s.Require().NoError(err)
	s.Require().Greater(packed.Gas(), uint64(21_000))

	buffered := tx.NewDynamicFeeBuilder()
	buffered.StateOverride(override)
	buffered.LimitBuffer(s.LimitBufferPct)
	bufferedTx, err := buffered.
		Build(s.from, s.token, big.NewInt(0), s.transferFromData).
		Pack(s.T().Context(), s.client)
	s.Require().NoError(err)
	s.Require().Equal(packed.Gas()*s.LimitBufferPct/100, bufferedTx.Gas())

	overrideJSON, err := override.MarshalJSON()
	s.Require().NoError(err)
	s.T().Logf("Bind token %s Test04PackTransferFromEstimatesWithOverride Result", s.token)
	s.T().Logf("\toverride: %s", overrideJSON)
	s.T().Logf("\tPack(transferFrom) gas limit: %d", packed.Gas())
	s.T().Logf("\tPack(transferFrom) with LimitBuffer(%d): %d", s.LimitBufferPct, bufferedTx.Gas())
	s.T().Logf("\tnonce %d", packed.AddressNonce())
}

func (s *StateOverrideSuite) Test05ApproveThenTransferFromBackToBack() {
	var nonceHex string
	e := rpc.ETHGetTransactionCount("nonce", s.from.String(), rpc.BlockTagPending, &nonceHex)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	nonce, err := strconv.ParseUint(strings.TrimPrefix(nonceHex, "0x"), 16, 64)
	s.Require().NoError(err)

	recipientBefore, err := Call(s.T().Context(), s.client, nil, s.token, s.recipientBalance)
	s.Require().NoError(err)

	approveBuilder := tx.NewDynamicFeeBuilder()
	approveBuilder.Nonce(nonce)
	approveTx, approveHash, err := Transact(
		s.T().Context(),
		s.client,
		NewTransactOptions(s.signer, approveBuilder),
		s.token,
		s.approveData,
	)
	s.Require().NoError(err)

	override := rpc.NewStateOverride().WithStateDiff(s.token, s.allowanceSlot, s.amountWord)
	transferBuilder := tx.NewDynamicFeeBuilder()
	transferBuilder.Nonce(nonce + 1)
	transferBuilder.StateOverride(override)
	transferTx, transferHash, err := Transact(
		s.T().Context(),
		s.client,
		NewTransactOptions(s.signer, transferBuilder),
		s.token,
		s.transferFromData,
	)
	s.Require().NoError(err)

	approveReceipt, err := tx.WaitForReceipt(
		s.T().Context(),
		s.client,
		approveHash,
		s.ReceiptPollInterval,
		s.ReceiptTimeout,
	)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve reverted")
	transferReceipt, err := tx.WaitForReceipt(
		s.T().Context(),
		s.client,
		transferHash,
		s.ReceiptPollInterval,
		s.ReceiptTimeout,
	)
	s.Require().NoError(err)
	s.Require().True(transferReceipt.Status, "transferFrom reverted with its override-based gas limit")

	s.Require().LessOrEqual(approveReceipt.GasUsed, approveTx.Gas())
	s.Require().LessOrEqual(transferReceipt.GasUsed, transferTx.Gas())
	s.Require().LessOrEqual(approveReceipt.BlockNumber, transferReceipt.BlockNumber)

	atTransfer := NewCallOptions().WithBlock("0x" + strconv.FormatUint(transferReceipt.BlockNumber, 16))
	allowance, err := Call(s.T().Context(), s.client, atTransfer, s.token, s.allowanceData)
	s.Require().NoError(err)
	s.Require().Zero(new(big.Int).SetBytes(allowance).Sign())
	recipientAfter, err := Call(s.T().Context(), s.client, atTransfer, s.token, s.recipientBalance)
	s.Require().NoError(err)
	received := new(big.Int).Sub(new(big.Int).SetBytes(recipientAfter), new(big.Int).SetBytes(recipientBefore))
	s.Require().Zero(s.Amount.Cmp(received))

	approvePriceHex, ok := approveReceipt.Raw["effectiveGasPrice"].(string)
	s.Require().True(ok)
	approvePrice, ok := new(big.Int).SetString(strings.TrimPrefix(approvePriceHex, "0x"), 16)
	s.Require().True(ok)
	transferPriceHex, ok := transferReceipt.Raw["effectiveGasPrice"].(string)
	s.Require().True(ok)
	transferPrice, ok := new(big.Int).SetString(strings.TrimPrefix(transferPriceHex, "0x"), 16)
	s.Require().True(ok)
	approveFee := new(big.Int).Mul(new(big.Int).SetUint64(approveReceipt.GasUsed), approvePrice)
	transferFee := new(big.Int).Mul(new(big.Int).SetUint64(transferReceipt.GasUsed), transferPrice)
	totalFee := new(big.Int).Add(approveFee, transferFee)

	s.mu.Lock()
	s.totalGas += approveReceipt.GasUsed + transferReceipt.GasUsed
	s.totalFee.Add(s.totalFee, totalFee)
	s.mu.Unlock()

	overrideJSON, err := override.MarshalJSON()
	s.Require().NoError(err)
	s.T().Logf("Bind token %s Test05ApproveThenTransferFromBackToBack Result", s.token)
	s.T().Logf("\toverride for transferFrom estimate: %s", overrideJSON)
	s.T().Logf("\tstep 1 approve(%s, %s)", s.from, s.Amount)
	s.T().Logf("\t\ttx %s, nonce %d, block %d", approveHash, approveTx.AddressNonce(), approveReceipt.BlockNumber)
	s.T().Logf("\t\testimated limit %d, gasUsed %d", approveTx.Gas(), approveReceipt.GasUsed)
	s.T().Logf("\t\tused %d%% of the estimate", approveReceipt.GasUsed*100/approveTx.Gas())
	s.T().Logf("\t\teffectiveGasPrice %s wei, fee %s wei", approvePrice, approveFee)
	s.T().Logf("\tstep 2 transferFrom(%s, %s, %s)", s.from, s.recipient, s.Amount)
	s.T().Logf("\t\ttx %s, nonce %d, block %d", transferHash, transferTx.AddressNonce(), transferReceipt.BlockNumber)
	s.T().Logf("\t\testimated limit %d, gasUsed %d", transferTx.Gas(), transferReceipt.GasUsed)
	s.T().Logf("\t\tused %d%% of the estimate", transferReceipt.GasUsed*100/transferTx.Gas())
	s.T().Logf("\t\teffectiveGasPrice %s wei, fee %s wei", transferPrice, transferFee)
	s.T().Logf("\ttotal")
	s.T().Logf("\t\testimated %d gas", approveTx.Gas()+transferTx.Gas())
	s.T().Logf("\t\tused %d gas", approveReceipt.GasUsed+transferReceipt.GasUsed)
	s.T().Logf("\t\tfee %s wei %s", totalFee, s.NativeSymbol)
	s.T().Logf("\tafter: allowance %s, recipient received %s", new(big.Int).SetBytes(allowance), received)
}
