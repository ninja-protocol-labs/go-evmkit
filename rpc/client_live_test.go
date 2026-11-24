package rpc

import (
	"errors"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/internal/testenv"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

type ClientSuite struct {
	suite.Suite

	RPCURL             string
	ChainID            int
	MulticallHex       string
	Permit2Hex         string
	FixtureDepth       uint64
	FixtureSearchLimit uint64
	FeeHistoryBlocks   uint64
	OverrideSeed       string
	OverrideBalance    *big.Int
	TransferValue      *big.Int
	TokenByteCode      string
	TokenName          string
	TokenSymbol        string
	TokenDecimals      uint8
	TokenSeed          string
	OwnerSeed          string
	SpenderSeed        string
	RecipientSeed      string
	TokenAmount        *big.Int

	client Client
	funded *types.Address

	token            *types.Address
	owner            *types.Address
	spender          *types.Address
	recipient        *types.Address
	tokenCode        []byte
	balanceSlot      [32]byte
	allowanceSlot    [32]byte
	amountWord       [32]byte
	approveData      string
	allowanceData    string
	transferFromData string

	blockNumber string
	blockHash   string
	txCount     int
	txHash      string
	txFrom      string
}

func TestClientSuite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)

	suite.Run(t, &ClientSuite{
		RPCURL:             c.RPCURL,
		ChainID:            c.Constant.ChainID,
		MulticallHex:       c.Addresses.Standard.Multicall,
		Permit2Hex:         c.Addresses.Standard.Permit2,
		FixtureDepth:       5,
		FixtureSearchLimit: 200,
		FeeHistoryBlocks:   4,
		OverrideSeed:       "go-evmkit-rpc-client-override",
		OverrideBalance:    new(big.Int).Add(new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil), big.NewInt(7)),
		TransferValue:      new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
		TokenByteCode:      testenv.ERC20ByteCodeHex,
		TokenName:          "evmkit override token",
		TokenSymbol:        "EVKO",
		TokenDecimals:      18,
		TokenSeed:          "go-evmkit-rpc-client-token",
		OwnerSeed:          "go-evmkit-rpc-client-owner",
		SpenderSeed:        "go-evmkit-rpc-client-spender",
		RecipientSeed:      "go-evmkit-rpc-client-recipient",
		TokenAmount:        new(big.Int).Mul(big.NewInt(250), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),
	})
}

func (s *ClientSuite) SetupSuite() {
	s.client = NewClient(s.RPCURL, WithRetry(5, ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.funded = types.NewAddressFromBytes(core.Keccak256([]byte(s.OverrideSeed)).Bytes())

	var head string
	e := ETHBlockNumber(uuid.New().String(), &head)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	headNumber, err := strconv.ParseUint(strings.TrimPrefix(head, "0x"), 16, 64)
	s.Require().NoError(err)
	s.Require().Greater(headNumber, s.FixtureDepth+s.FixtureSearchLimit)

	for number := headNumber - s.FixtureDepth; number > headNumber-s.FixtureDepth-s.FixtureSearchLimit; number-- {
		var block map[string]any
		tag := "0x" + strconv.FormatUint(number, 16)
		e = ETHGetBlockByNumber(uuid.New().String(), tag, true, &block)
		s.Require().NoError(s.client.Call(s.T().Context(), e))
		txs, ok := block["transactions"].([]any)
		s.Require().True(ok, "block %s has no transactions array", tag)
		if len(txs) == 0 {
			continue
		}
		first, ok := txs[0].(map[string]any)
		s.Require().True(ok, "block %s returned transaction hashes for fullTx=true", tag)
		s.blockNumber = tag
		s.blockHash = block["hash"].(string)
		s.txCount = len(txs)
		s.txHash = first["hash"].(string)
		s.txFrom = first["from"].(string)
		break
	}
	s.Require().NotEmpty(s.txHash, "no block with transactions in the last %d blocks", s.FixtureSearchLimit)

	s.token = types.NewAddressFromBytes(core.Keccak256([]byte(s.TokenSeed)).Bytes())
	s.owner = types.NewAddressFromBytes(core.Keccak256([]byte(s.OwnerSeed)).Bytes())
	s.spender = types.NewAddressFromBytes(core.Keccak256([]byte(s.SpenderSeed)).Bytes())
	s.recipient = types.NewAddressFromBytes(core.Keccak256([]byte(s.RecipientSeed)).Bytes())

	bytecode, err := encoding.Hex.Decode(s.TokenByteCode)
	s.Require().NoError(err)
	args, err := abi.Pack(abi.Types{abi.String, abi.String, abi.Uint8}, s.TokenName, s.TokenSymbol, s.TokenDecimals)
	s.Require().NoError(err)
	var runtime string
	params := map[string]any{
		"data": encoding.Hex.EncodePrefixed(append(bytecode, args...)),
	}
	e = ETHCall(uuid.New().String(), params, BlockTagLatest, nil, &runtime)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.tokenCode, err = encoding.Hex.Decode(runtime)
	s.Require().NoError(err)
	s.Require().NotEmpty(s.tokenCode)

	balanceKey, err := abi.Pack(abi.Types{abi.Address, abi.Uint256}, s.owner, big.NewInt(0))
	s.Require().NoError(err)
	copy(s.balanceSlot[:], core.Keccak256(balanceKey).Bytes())
	ownerKey, err := abi.Pack(abi.Types{abi.Address, abi.Uint256}, s.owner, big.NewInt(1))
	s.Require().NoError(err)
	allowanceKey, err := abi.Pack(
		abi.Types{abi.Address, abi.Uint256},
		s.spender,
		new(big.Int).SetBytes(core.Keccak256(ownerKey).Bytes()),
	)
	s.Require().NoError(err)
	copy(s.allowanceSlot[:], core.Keccak256(allowanceKey).Bytes())
	s.TokenAmount.FillBytes(s.amountWord[:])

	approveData, err := abi.NewFunction("approve", abi.Types{abi.Address, abi.Uint256}, abi.Types{abi.Bool}).
		EncodeCall(s.spender, s.TokenAmount)
	s.Require().NoError(err)
	s.approveData = encoding.Hex.EncodePrefixed(approveData)
	allowanceData, err := abi.NewFunction("allowance", abi.Types{abi.Address, abi.Address}, abi.Types{abi.Uint256}).
		EncodeCall(s.owner, s.spender)
	s.Require().NoError(err)
	s.allowanceData = encoding.Hex.EncodePrefixed(allowanceData)
	transferFromData, err := abi.NewFunction(
		"transferFrom",
		abi.Types{abi.Address, abi.Address, abi.Uint256},
		abi.Types{abi.Bool},
	).EncodeCall(s.owner, s.recipient, s.TokenAmount)
	s.Require().NoError(err)
	s.transferFromData = encoding.Hex.EncodePrefixed(transferFromData)
	s.T().Logf("SetupSuite: chain %d head %s", s.ChainID, head)
	s.T().Logf("SetupSuite: fixture block %s (%s) with %d txs", s.blockNumber, s.blockHash, s.txCount)
	s.T().Logf("SetupSuite: fixture tx %s from %s", s.txHash, s.txFrom)
	s.T().Logf("SetupSuite: Multicall %s, Permit2 %s", s.MulticallHex, s.Permit2Hex)
	s.T().Logf("SetupSuite: funded %s, override balance %s", s.funded, s.OverrideBalance)
	s.T().Logf("SetupSuite: token %s (%d runtime bytes)", s.token, len(s.tokenCode))
	s.T().Logf("SetupSuite: owner %s, spender %s, recipient %s", s.owner, s.spender, s.recipient)
	s.T().Logf("SetupSuite: balance slot %x", s.balanceSlot)
	s.T().Logf("SetupSuite: allowance slot %x", s.allowanceSlot)
	s.T().Logf("SetupSuite: token amount %s", s.TokenAmount)
}

func (s *ClientSuite) Test01ChainIDMatchesConfig() {
	var result string
	e := ETHChainID(uuid.New().String(), &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Equal("0x"+strconv.FormatInt(int64(s.ChainID), 16), result)
	s.T().Logf("Client chain %d Test01ChainIDMatchesConfig Result", s.ChainID)
	s.T().Logf("\teth_chainId: %s (config %d)", result, s.ChainID)
}

func (s *ClientSuite) Test02BlockNumberIsAtOrAboveFixture() {
	var result string
	e := ETHBlockNumber(uuid.New().String(), &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	head, err := strconv.ParseUint(strings.TrimPrefix(result, "0x"), 16, 64)
	s.Require().NoError(err)
	fixture, err := strconv.ParseUint(strings.TrimPrefix(s.blockNumber, "0x"), 16, 64)
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(head, fixture)
	s.T().Logf("Client chain %d Test02BlockNumberIsAtOrAboveFixture Result", s.ChainID)
	s.T().Logf("\teth_blockNumber: %s (%d), fixture %s (%d)", result, head, s.blockNumber, fixture)
}

func (s *ClientSuite) Test03GasPriceIsQuantity() {
	var result string
	e := ETHGasPrice(uuid.New().String(), &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	gasPrice, err := strconv.ParseUint(strings.TrimPrefix(result, "0x"), 16, 64)
	s.Require().NoError(err, result)
	s.T().Logf("Client chain %d Test03GasPriceIsQuantity Result", s.ChainID)
	s.T().Logf("\teth_gasPrice: %s (%d wei)", result, gasPrice)
}

func (s *ClientSuite) Test04MaxPriorityFeePerGasIsQuantity() {
	var result string
	e := ETHMaxPriorityFeePerGas(uuid.New().String(), &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	tip, err := strconv.ParseUint(strings.TrimPrefix(result, "0x"), 16, 64)
	s.Require().NoError(err, result)
	s.T().Logf("Client chain %d Test04MaxPriorityFeePerGasIsQuantity Result", s.ChainID)
	s.T().Logf("\teth_maxPriorityFeePerGas: %s (%d wei)", result, tip)
}

func (s *ClientSuite) Test05FeeHistoryCoversRequestedBlocks() {
	var result map[string]any
	blockCount := "0x" + strconv.FormatUint(s.FeeHistoryBlocks, 16)
	e := ETHFeeHistory(uuid.New().String(), blockCount, BlockTagLatest, []float64{25, 50, 75}, &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	baseFees, ok := result["baseFeePerGas"].([]any)
	s.Require().True(ok, "%v", result)
	s.Require().Len(baseFees, int(s.FeeHistoryBlocks)+1)
	rewards, ok := result["reward"].([]any)
	s.Require().True(ok, "%v", result)
	s.Require().Len(rewards, int(s.FeeHistoryBlocks))
	s.T().Logf("Client chain %d Test05FeeHistoryCoversRequestedBlocks Result", s.ChainID)
	s.T().Logf("\toldestBlock: %v", result["oldestBlock"])
	s.T().Logf("\tbaseFeePerGas: %v", baseFees)
	s.T().Logf("\tgasUsedRatio: %v", result["gasUsedRatio"])
	s.T().Logf("\treward [25 50 75]: %v", rewards)
}

func (s *ClientSuite) Test06BalanceOfFixtureSenderIsQuantity() {
	var result string
	e := ETHGetBalance(uuid.New().String(), s.txFrom, BlockTagLatest, &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().True(strings.HasPrefix(result, "0x"), result)
	s.T().Logf("Client chain %d Test06BalanceOfFixtureSenderIsQuantity Result", s.ChainID)
	s.T().Logf("\teth_getBalance(%s): %s", s.txFrom, result)
}

func (s *ClientSuite) Test07CodeOfMulticallIsDeployed() {
	var result string
	e := ETHGetCode(uuid.New().String(), s.MulticallHex, BlockTagLatest, &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Greater(len(result), 2)
	s.T().Logf("Client chain %d Test07CodeOfMulticallIsDeployed Result", s.ChainID)
	s.T().Logf("\teth_getCode(%s): %d bytes", s.MulticallHex, (len(result)-2)/2)
}

func (s *ClientSuite) Test08TransactionCountOfFixtureSenderIsQuantity() {
	var result string
	e := ETHGetTransactionCount(uuid.New().String(), s.txFrom, BlockTagLatest, &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	nonce, err := strconv.ParseUint(strings.TrimPrefix(result, "0x"), 16, 64)
	s.Require().NoError(err, result)
	s.T().Logf("Client chain %d Test08TransactionCountOfFixtureSenderIsQuantity Result", s.ChainID)
	s.T().Logf("\teth_getTransactionCount(%s): %s (%d)", s.txFrom, result, nonce)
}

func (s *ClientSuite) Test09StorageAtReturnsWord() {
	var result string
	e := ETHGetStorageAt(uuid.New().String(), s.Permit2Hex, "0x0", BlockTagLatest, &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Len(result, 66)
	s.T().Logf("Client chain %d Test09StorageAtReturnsWord Result", s.ChainID)
	s.T().Logf("\teth_getStorageAt(%s, 0x0): %s", s.Permit2Hex, result)
}

func (s *ClientSuite) Test10CallMulticallGetChainID() {
	var result string
	params := map[string]any{
		"to":   s.MulticallHex,
		"data": "0x3408e470",
	}
	e := ETHCall(uuid.New().String(), params, BlockTagLatest, nil, &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Len(result, 66)
	chainID, err := strconv.ParseUint(strings.TrimLeft(strings.TrimPrefix(result, "0x"), "0"), 16, 64)
	s.Require().NoError(err)
	s.Require().Equal(uint64(s.ChainID), chainID)
	s.T().Logf("Client chain %d Test10CallMulticallGetChainID Result", s.ChainID)
	s.T().Logf("\tgetChainId() on %s: %s (%d)", s.MulticallHex, result, chainID)
}

func (s *ClientSuite) Test11CallWithStateOverrideSeesOverriddenBalance() {
	data, err := abi.NewFunction("getEthBalance", abi.Types{abi.Address}, abi.Types{abi.Uint256}).EncodeCall(s.funded)
	s.Require().NoError(err)
	params := map[string]any{
		"to":   s.MulticallHex,
		"data": encoding.Hex.EncodePrefixed(data),
	}

	var plain string
	e := ETHCall(uuid.New().String(), params, BlockTagLatest, nil, &plain)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Equal("0x"+strings.Repeat("00", 32), plain)

	var overridden string
	override := NewStateOverride().WithBalance(s.funded, s.OverrideBalance)
	e = ETHCall(uuid.New().String(), params, BlockTagLatest, override, &overridden)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	out, err := encoding.Hex.Decode(overridden)
	s.Require().NoError(err)
	s.Require().Len(out, 32)
	s.Require().Zero(s.OverrideBalance.Cmp(new(big.Int).SetBytes(out)))
	s.T().Logf("Client chain %d Test11CallWithStateOverrideSeesOverriddenBalance Result", s.ChainID)
	overrideJSON, err := override.MarshalJSON()
	s.Require().NoError(err)
	s.T().Logf("\tgetEthBalance(%s) without override: %s", s.funded, plain)
	s.T().Logf("\toverride: %s", overrideJSON)
	s.T().Logf("\tgetEthBalance(%s) with override: %s (%s)", s.funded, overridden, new(big.Int).SetBytes(out))
}

func (s *ClientSuite) Test12EstimateGasForSelfTransfer() {
	var result string
	params := map[string]any{
		"from":  s.txFrom,
		"to":    s.txFrom,
		"value": "0x0",
	}
	e := ETHEstimateGas(uuid.New().String(), params, BlockTagLatest, nil, &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	gas, err := strconv.ParseUint(strings.TrimPrefix(result, "0x"), 16, 64)
	s.Require().NoError(err)
	s.Require().Positive(gas)
	s.T().Logf("Client chain %d Test12EstimateGasForSelfTransfer Result", s.ChainID)
	s.T().Logf("\teth_estimateGas(%s -> self, 0 wei): %s (%d gas)", s.txFrom, result, gas)
}

func (s *ClientSuite) Test13EstimateGasWithStateOverrideFundsSender() {
	params := map[string]any{
		"from":  s.funded.String(),
		"to":    s.txFrom,
		"value": "0x" + s.TransferValue.Text(16),
	}

	var unfunded string
	e := ETHEstimateGas(uuid.New().String(), params, BlockTagLatest, nil, &unfunded)
	unfundedErr := s.client.Call(s.T().Context(), e)
	s.Require().Error(unfundedErr)

	var funded string
	override := NewStateOverride().WithBalance(s.funded, s.OverrideBalance)
	e = ETHEstimateGas(uuid.New().String(), params, BlockTagLatest, override, &funded)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	gas, err := strconv.ParseUint(strings.TrimPrefix(funded, "0x"), 16, 64)
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(gas, uint64(21_000))
	s.T().Logf("Client chain %d Test13EstimateGasWithStateOverrideFundsSender Result", s.ChainID)
	overrideJSON, err := override.MarshalJSON()
	s.Require().NoError(err)
	s.T().Logf("\ttransfer: %s -> %s, %s wei", s.funded, s.txFrom, s.TransferValue)
	s.T().Logf("\twithout override: %v", unfundedErr)
	s.T().Logf("\toverride: %s", overrideJSON)
	s.T().Logf("\twith override: %s (%d gas)", funded, gas)
}

func (s *ClientSuite) Test14ApproveAndTransferFromCallWithStateOverride() {
	approveParams := map[string]any{
		"from": s.owner.String(),
		"to":   s.token.String(),
		"data": s.approveData,
	}
	var approved string
	override := NewStateOverride().WithCode(s.token, s.tokenCode)
	e := ETHCall(uuid.New().String(), approveParams, BlockTagLatest, override, &approved)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Equal("0x"+strings.Repeat("00", 31)+"01", approved)

	allowanceParams := map[string]any{
		"to":   s.token.String(),
		"data": s.allowanceData,
	}
	var allowance string
	override = NewStateOverride().
		WithCode(s.token, s.tokenCode).
		WithStateDiff(s.token, s.allowanceSlot, s.amountWord)
	e = ETHCall(uuid.New().String(), allowanceParams, BlockTagLatest, override, &allowance)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Equal(encoding.Hex.EncodePrefixed(s.amountWord[:]), allowance)

	transferFromParams := map[string]any{
		"from": s.spender.String(),
		"to":   s.token.String(),
		"data": s.transferFromData,
	}
	var unapproved string
	override = NewStateOverride().
		WithCode(s.token, s.tokenCode).
		WithStateDiff(s.token, s.balanceSlot, s.amountWord)
	e = ETHCall(uuid.New().String(), transferFromParams, BlockTagLatest, override, &unapproved)
	unapprovedErr := s.client.Call(s.T().Context(), e)
	s.Require().Error(unapprovedErr)

	var transferred string
	override = NewStateOverride().
		WithCode(s.token, s.tokenCode).
		WithStateDiff(s.token, s.balanceSlot, s.amountWord).
		WithStateDiff(s.token, s.allowanceSlot, s.amountWord)
	e = ETHCall(uuid.New().String(), transferFromParams, BlockTagLatest, override, &transferred)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Equal("0x"+strings.Repeat("00", 31)+"01", transferred)
	s.T().Logf("Client chain %d Test14ApproveAndTransferFromCallWithStateOverride Result", s.ChainID)
	s.T().Logf("\ttoken %s (injected runtime)", s.token)
	s.T().Logf("\towner %s, spender %s, recipient %s", s.owner, s.spender, s.recipient)
	s.T().Logf("\tapprove(%s, %s) from owner: %s", s.spender, s.TokenAmount, approved)
	s.T().Logf("\tallowance(owner, spender) with allowance slot %x: %s", s.allowanceSlot, allowance)
	s.T().Logf("\ttransferFrom without allowance: %v", unapprovedErr)
	s.T().Logf("\toverride: code %d bytes on %s", len(s.tokenCode), s.token)
	s.T().Logf("\toverride: stateDiff balance[%x] = %s", s.balanceSlot, s.TokenAmount)
	s.T().Logf("\toverride: stateDiff allowance[%x] = %s", s.allowanceSlot, s.TokenAmount)
	s.T().Logf("\ttransferFrom with balance+allowance: %s", transferred)
}

func (s *ClientSuite) Test15ApproveAndTransferFromEstimateGasWithStateOverride() {
	approveParams := map[string]any{
		"from": s.owner.String(),
		"to":   s.token.String(),
		"data": s.approveData,
	}
	var approveGas string
	override := NewStateOverride().WithCode(s.token, s.tokenCode)
	e := ETHEstimateGas(uuid.New().String(), approveParams, BlockTagLatest, override, &approveGas)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	approve, err := strconv.ParseUint(strings.TrimPrefix(approveGas, "0x"), 16, 64)
	s.Require().NoError(err)
	s.Require().Positive(approve)

	transferFromParams := map[string]any{
		"from": s.spender.String(),
		"to":   s.token.String(),
		"data": s.transferFromData,
	}
	var unapproved string
	override = NewStateOverride().
		WithCode(s.token, s.tokenCode).
		WithStateDiff(s.token, s.balanceSlot, s.amountWord)
	e = ETHEstimateGas(uuid.New().String(), transferFromParams, BlockTagLatest, override, &unapproved)
	unapprovedErr := s.client.Call(s.T().Context(), e)
	s.Require().Error(unapprovedErr)

	var transferFromGas string
	override = NewStateOverride().
		WithCode(s.token, s.tokenCode).
		WithStateDiff(s.token, s.balanceSlot, s.amountWord).
		WithStateDiff(s.token, s.allowanceSlot, s.amountWord)
	e = ETHEstimateGas(uuid.New().String(), transferFromParams, BlockTagLatest, override, &transferFromGas)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	transferFrom, err := strconv.ParseUint(strings.TrimPrefix(transferFromGas, "0x"), 16, 64)
	s.Require().NoError(err)
	s.Require().Positive(transferFrom)

	var gasPriceHex string
	e = ETHGasPrice(uuid.New().String(), &gasPriceHex)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	gasPrice, ok := new(big.Int).SetString(strings.TrimPrefix(gasPriceHex, "0x"), 16)
	s.Require().True(ok, gasPriceHex)
	approveFee := new(big.Int).Mul(new(big.Int).SetUint64(approve), gasPrice)
	transferFromFee := new(big.Int).Mul(new(big.Int).SetUint64(transferFrom), gasPrice)
	totalFee := new(big.Int).Add(approveFee, transferFromFee)

	s.T().Logf("Client chain %d Test15ApproveAndTransferFromEstimateGasWithStateOverride Result", s.ChainID)
	s.T().Logf("\ttoken %s (injected runtime)", s.token)
	s.T().Logf("\towner %s, spender %s", s.owner, s.spender)
	s.T().Logf("\tapprove(%s, %s): %s (%d gas)", s.spender, s.TokenAmount, approveGas, approve)
	s.T().Logf("\ttransferFrom without allowance: %v", unapprovedErr)
	s.T().Logf("\toverride: code %d bytes on %s", len(s.tokenCode), s.token)
	s.T().Logf("\toverride: stateDiff balance[%x] = %s", s.balanceSlot, s.TokenAmount)
	s.T().Logf("\toverride: stateDiff allowance[%x] = %s", s.allowanceSlot, s.TokenAmount)
	s.T().Logf("\ttransferFrom with balance+allowance: %s (%d gas)", transferFromGas, transferFrom)
	s.T().Logf("\tsteps at gasPrice %s wei:", gasPrice)
	s.T().Logf("\t\tstep 1 approve (owner): %d gas, %s wei", approve, approveFee)
	s.T().Logf("\t\tstep 2 transferFrom (spender): %d gas, %s wei", transferFrom, transferFromFee)
	s.T().Logf("\t\ttotal: %d gas, %s wei", approve+transferFrom, totalFee)
}

func (s *ClientSuite) Test16BlockByNumberMatchesFixture() {
	var result map[string]any
	e := ETHGetBlockByNumber(uuid.New().String(), s.blockNumber, false, &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Equal(s.blockHash, result["hash"])
	s.Require().Equal(s.blockNumber, result["number"])
	txs, ok := result["transactions"].([]any)
	s.Require().True(ok)
	s.Require().Len(txs, s.txCount)
	s.Require().Equal(s.txHash, txs[0])
	s.T().Logf("Client chain %d Test16BlockByNumberMatchesFixture Result", s.ChainID)
	s.T().Logf("\tblock %s: hash %s, %d txs, first %s", result["number"], result["hash"], len(txs), txs[0])
	s.T().Logf("\tparentHash: %v", result["parentHash"])
	s.T().Logf("\ttimestamp: %v, gasUsed: %v", result["timestamp"], result["gasUsed"])
	s.T().Logf("\tbaseFeePerGas: %v", result["baseFeePerGas"])
}

func (s *ClientSuite) Test17BlockByHashMatchesFixture() {
	var result map[string]any
	e := ETHGetBlockByHash(uuid.New().String(), s.blockHash, true, &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Equal(s.blockNumber, result["number"])
	txs, ok := result["transactions"].([]any)
	s.Require().True(ok)
	s.Require().Len(txs, s.txCount)
	s.Require().Equal(s.txHash, txs[0].(map[string]any)["hash"])
	s.T().Logf("Client chain %d Test17BlockByHashMatchesFixture Result", s.ChainID)
	s.T().Logf("\tblock %s: number %s, %d full txs", s.blockHash, result["number"], len(txs))
	s.T().Logf("\tfirst tx: %v", txs[0].(map[string]any)["hash"])
}

func (s *ClientSuite) Test18TransactionByHashMatchesFixture() {
	var result map[string]any
	e := ETHGetTransactionByHash(uuid.New().String(), s.txHash, &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Equal(s.txHash, result["hash"])
	s.Require().Equal(s.blockHash, result["blockHash"])
	s.Require().Equal(s.blockNumber, result["blockNumber"])
	s.Require().Equal(s.txFrom, result["from"])
	s.T().Logf("Client chain %d Test18TransactionByHashMatchesFixture Result", s.ChainID)
	s.T().Logf("\ttx %s: type %v, nonce %v", result["hash"], result["type"], result["nonce"])
	s.T().Logf("\tfrom %v, to %v", result["from"], result["to"])
	s.T().Logf("\tblock %v (%v), index %v", result["blockNumber"], result["blockHash"], result["transactionIndex"])
	s.T().Logf("\tvalue %v, gas %v", result["value"], result["gas"])
}

func (s *ClientSuite) Test19TransactionByBlockHashAndIndexMatchesFixture() {
	var result map[string]any
	e := ETHGetTransactionByBlockHashAndIndex(uuid.New().String(), s.blockHash, "0x0", &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Equal(s.txHash, result["hash"])
	s.T().Logf("Client chain %d Test19TransactionByBlockHashAndIndexMatchesFixture Result", s.ChainID)
	s.T().Logf("\t(%s, 0x0): %v", s.blockHash, result["hash"])
}

func (s *ClientSuite) Test20TransactionByBlockNumberAndIndexMatchesFixture() {
	var result map[string]any
	e := ETHGetTransactionByBlockNumberAndIndex(uuid.New().String(), s.blockNumber, "0x0", &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Equal(s.txHash, result["hash"])
	s.T().Logf("Client chain %d Test20TransactionByBlockNumberAndIndexMatchesFixture Result", s.ChainID)
	s.T().Logf("\t(%s, 0x0): %v", s.blockNumber, result["hash"])
}

func (s *ClientSuite) Test21ReceiptMatchesFixture() {
	var result map[string]any
	e := ETHGetTransactionReceipt(uuid.New().String(), s.txHash, &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Equal(s.txHash, result["transactionHash"])
	s.Require().Equal(s.blockHash, result["blockHash"])
	s.Require().Equal(s.blockNumber, result["blockNumber"])
	s.Require().Equal("0x0", result["transactionIndex"])
	s.T().Logf("Client chain %d Test21ReceiptMatchesFixture Result", s.ChainID)
	s.T().Logf("\treceipt %v: status %v", result["transactionHash"], result["status"])
	s.T().Logf("\tgasUsed %v, effectiveGasPrice %v", result["gasUsed"], result["effectiveGasPrice"])
	s.T().Logf("\tblock %v (%v), index %v", result["blockNumber"], result["blockHash"], result["transactionIndex"])
	s.T().Logf("\tlogs: %d", len(result["logs"].([]any)))
}

func (s *ClientSuite) Test22BlockReceiptsMatchFixture() {
	var result []map[string]any
	e := ETHGetBlockReceipts(uuid.New().String(), s.blockNumber, &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Len(result, s.txCount)
	s.Require().Equal(s.txHash, result[0]["transactionHash"])
	for i, receipt := range result {
		s.Require().Equal(s.blockHash, receipt["blockHash"], i)
	}
	s.T().Logf("Client chain %d Test22BlockReceiptsMatchFixture Result", s.ChainID)
	s.T().Logf("\teth_getBlockReceipts(%s): %d receipts", s.blockNumber, len(result))
	s.T().Logf("\tfirst receipt: %v", result[0]["transactionHash"])
}

func (s *ClientSuite) Test23LogsByBlockHashMatchReceipts() {
	var receipts []map[string]any
	e := ETHGetBlockReceipts(uuid.New().String(), s.blockNumber, &receipts)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	want := 0
	for _, receipt := range receipts {
		logs, ok := receipt["logs"].([]any)
		s.Require().True(ok)
		want += len(logs)
	}

	var result []map[string]any
	filter := map[string]any{
		"blockHash": s.blockHash,
	}
	e = ETHGetLogs(uuid.New().String(), filter, &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Len(result, want)
	for i, log := range result {
		s.Require().Equal(s.blockHash, log["blockHash"], i)
	}
	s.T().Logf("Client chain %d Test23LogsByBlockHashMatchReceipts Result", s.ChainID)
	s.T().Logf("\treceipts in %s carry %d logs", s.blockHash, want)
	s.T().Logf("\teth_getLogs(blockHash): %d logs", len(result))
}

func (s *ClientSuite) Test24TransactionCountByHashMatchesFixture() {
	var result string
	e := ETHGetBlockTransactionCountByHash(uuid.New().String(), s.blockHash, &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Equal("0x"+strconv.FormatInt(int64(s.txCount), 16), result)
	s.T().Logf("Client chain %d Test24TransactionCountByHashMatchesFixture Result", s.ChainID)
	s.T().Logf("\teth_getBlockTransactionCountByHash(%s): %s (fixture %d)", s.blockHash, result, s.txCount)
}

func (s *ClientSuite) Test25TransactionCountByNumberMatchesFixture() {
	var result string
	e := ETHGetBlockTransactionCountByNumber(uuid.New().String(), s.blockNumber, &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Equal("0x"+strconv.FormatInt(int64(s.txCount), 16), result)
	s.T().Logf("Client chain %d Test25TransactionCountByNumberMatchesFixture Result", s.ChainID)
	s.T().Logf("\teth_getBlockTransactionCountByNumber(%s): %s (fixture %d)", s.blockNumber, result, s.txCount)
}

func (s *ClientSuite) Test26Syncing() {
	var result any
	e := ETHSyncing(uuid.New().String(), &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().NotNil(result)
	s.T().Logf("Client chain %d Test26Syncing Result", s.ChainID)
	s.T().Logf("\teth_syncing: %v", result)
}

func (s *ClientSuite) Test27NetVersion() {
	var result string
	e := NetVersion(uuid.New().String(), &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().NotEmpty(result)
	s.T().Logf("Client chain %d Test27NetVersion Result", s.ChainID)
	s.T().Logf("\tnet_version: %s", result)
}

func (s *ClientSuite) Test28Web3ClientVersion() {
	var result string
	e := Web3ClientVersion(uuid.New().String(), &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().NotEmpty(result)
	s.T().Logf("Client chain %d Test28Web3ClientVersion Result", s.ChainID)
	s.T().Logf("\tweb3_clientVersion: %s", result)
}

func (s *ClientSuite) Test29BatchEIP1559FeeData() {
	var nonce, maxPriorityFeePerGas, gasPrice, chainID string
	s.Require().NoError(s.client.Batch(s.T().Context(), Elements{
		ETHGetTransactionCount(uuid.New().String(), s.txFrom, BlockTagLatest, &nonce),
		ETHMaxPriorityFeePerGas(uuid.New().String(), &maxPriorityFeePerGas),
		ETHGasPrice(uuid.New().String(), &gasPrice),
		ETHChainID(uuid.New().String(), &chainID),
	}))
	s.Require().NotEmpty(nonce)
	s.Require().NotEmpty(maxPriorityFeePerGas)
	s.Require().NotEmpty(gasPrice)
	s.Require().Equal("0x"+strconv.FormatInt(int64(s.ChainID), 16), chainID)
	s.T().Logf("Client chain %d Test29BatchEIP1559FeeData Result", s.ChainID)
	s.T().Logf("\tnonce(%s): %s", s.txFrom, nonce)
	s.T().Logf("\tmaxPriorityFeePerGas: %s", maxPriorityFeePerGas)
	s.T().Logf("\tgasPrice: %s", gasPrice)
	s.T().Logf("\tchainId: %s", chainID)
}

func (s *ClientSuite) Test30UnknownMethodReturnsResponseError() {
	var result any
	e := NewElement(uuid.New().String(), "evmkit_doesNotExist", []any{}, &result)
	err := s.client.Call(s.T().Context(), e)
	s.Require().Error(err)
	re, ok := errors.AsType[*ResponseError](err)
	s.Require().True(ok, "%v", err)
	s.Require().NotZero(re.Code)
	s.T().Logf("Client chain %d Test30UnknownMethodReturnsResponseError Result", s.ChainID)
	s.T().Logf("\terror: %v", err)
	s.T().Logf("\tResponseError: code %d, message %q", re.Code, re.Message)
}
