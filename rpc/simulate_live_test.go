package rpc

import (
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

type SimulateV1Suite struct {
	suite.Suite

	RPCURL          string
	ChainID         int
	TokenByteCode   string
	TokenName       string
	TokenSymbol     string
	TokenDecimals   uint8
	TokenSeed       string
	OwnerSeed       string
	SpenderSeed     string
	RecipientSeed   string
	TokenAmount     *big.Int
	OverrideBalance *big.Int
	TransferValue   *big.Int

	client Client

	token     *types.Address
	owner     *types.Address
	spender   *types.Address
	recipient *types.Address
	tokenCode []byte

	balanceSlot  [32]byte
	amountWord   [32]byte
	approveData  string
	transferData string
	ownerOfData  string
	recipOfData  string

	transferTopic string
	approvalTopic string
}

func TestSimulateV1Suite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)

	suite.Run(t, &SimulateV1Suite{
		RPCURL:          c.RPCURL,
		ChainID:         c.Constant.ChainID,
		TokenByteCode:   testenv.ERC20ByteCodeHex,
		TokenName:       "evmkit simulate token",
		TokenSymbol:     "EVKS",
		TokenDecimals:   18,
		TokenSeed:       "go-evmkit-rpc-simulate-token",
		OwnerSeed:       "go-evmkit-rpc-simulate-owner",
		SpenderSeed:     "go-evmkit-rpc-simulate-spender",
		RecipientSeed:   "go-evmkit-rpc-simulate-recipient",
		TokenAmount:     new(big.Int).Mul(big.NewInt(250), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),
		OverrideBalance: new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil),
		TransferValue:   new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
	})
}

func (s *SimulateV1Suite) SetupSuite() {
	s.client = NewClient(s.RPCURL, WithRetry(5, ExponentialBackoff(500*time.Millisecond, 10*time.Second)))

	s.token = types.NewAddressFromBytes(core.Keccak256([]byte(s.TokenSeed)).Bytes())
	s.owner = types.NewAddressFromBytes(core.Keccak256([]byte(s.OwnerSeed)).Bytes())
	s.spender = types.NewAddressFromBytes(core.Keccak256([]byte(s.SpenderSeed)).Bytes())
	s.recipient = types.NewAddressFromBytes(core.Keccak256([]byte(s.RecipientSeed)).Bytes())

	var probe []map[string]any
	probeErr := s.client.Call(s.T().Context(), ETHSimulateV1(uuid.New().String(), map[string]any{
		"blockStateCalls": []any{
			map[string]any{"calls": []any{map[string]any{"from": s.owner.String(), "to": s.recipient.String()}}},
		},
	}, BlockTagLatest, &probe))
	if probeErr != nil {
		s.T().Skipf("eth_simulateV1 unavailable on chain %d: %v", s.ChainID, probeErr)
	}

	bytecode, err := encoding.Hex.Decode(s.TokenByteCode)
	s.Require().NoError(err)
	args, err := abi.Pack(abi.NewTypes(abi.String, abi.String, abi.Uint8), s.TokenName, s.TokenSymbol, s.TokenDecimals)
	s.Require().NoError(err)
	var runtime string
	e := ETHCall(uuid.New().String(), map[string]any{"data": encoding.Hex.EncodePrefixed(append(bytecode, args...))}, BlockTagLatest, nil, &runtime)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.tokenCode, err = encoding.Hex.Decode(runtime)
	s.Require().NoError(err)
	s.Require().NotEmpty(s.tokenCode)

	balanceKey, err := abi.Pack(abi.NewTypes(abi.Address, abi.Uint256), s.owner, big.NewInt(0))
	s.Require().NoError(err)
	copy(s.balanceSlot[:], core.Keccak256(balanceKey).Bytes())
	s.TokenAmount.FillBytes(s.amountWord[:])

	approveData, err := abi.NewFunction("approve", abi.NewTypes(abi.Address, abi.Uint256), abi.NewTypes(abi.Bool)).
		EncodeCall(s.spender, s.TokenAmount)
	s.Require().NoError(err)
	s.approveData = encoding.Hex.EncodePrefixed(approveData)
	transferData, err := abi.NewFunction("transferFrom", abi.NewTypes(abi.Address, abi.Address, abi.Uint256), abi.NewTypes(abi.Bool)).
		EncodeCall(s.owner, s.recipient, s.TokenAmount)
	s.Require().NoError(err)
	s.transferData = encoding.Hex.EncodePrefixed(transferData)
	balanceOf := abi.NewFunction("balanceOf", abi.NewTypes(abi.Address), abi.NewTypes(abi.Uint256))
	ownerOfData, err := balanceOf.EncodeCall(s.owner)
	s.Require().NoError(err)
	s.ownerOfData = encoding.Hex.EncodePrefixed(ownerOfData)
	recipOfData, err := balanceOf.EncodeCall(s.recipient)
	s.Require().NoError(err)
	s.recipOfData = encoding.Hex.EncodePrefixed(recipOfData)

	s.transferTopic = core.Keccak256([]byte("Transfer(address,address,uint256)")).String()
	s.approvalTopic = core.Keccak256([]byte("Approval(address,address,uint256)")).String()

	s.T().Logf("SetupSuite: chain %d, probe returned %d block(s)", s.ChainID, len(probe))
	s.T().Logf("SetupSuite: token %s (%d runtime bytes)", s.token, len(s.tokenCode))
	s.T().Logf("SetupSuite: owner %s, spender %s, recipient %s", s.owner, s.spender, s.recipient)
	s.T().Logf("SetupSuite: balance slot %x, token amount %s", s.balanceSlot, s.TokenAmount)
}

func (s *SimulateV1Suite) Test01ApproveThenTransferFromRunInOrder() {
	override := NewStateOverride().
		WithCode(s.token, s.tokenCode).
		WithStateDiff(s.token, s.balanceSlot, s.amountWord)
	payload := map[string]any{
		"blockStateCalls": []any{
			map[string]any{
				"stateOverrides": override,
				"calls": []any{
					map[string]any{"from": s.owner.String(), "to": s.token.String(), "data": s.approveData},
					map[string]any{"from": s.spender.String(), "to": s.token.String(), "data": s.transferData},
					map[string]any{"to": s.token.String(), "data": s.recipOfData},
					map[string]any{"to": s.token.String(), "data": s.ownerOfData},
				},
			},
		},
	}

	var blocks []map[string]any
	s.Require().NoError(s.client.Call(s.T().Context(), ETHSimulateV1(uuid.New().String(), payload, BlockTagLatest, &blocks)))
	s.Require().Len(blocks, 1)
	calls, ok := blocks[0]["calls"].([]any)
	s.Require().True(ok)
	s.Require().Len(calls, 4)

	approve, ok := calls[0].(map[string]any)
	s.Require().True(ok)
	s.Require().Equal("0x1", approve["status"], "%v", approve["error"])
	approveLogs, ok := approve["logs"].([]any)
	s.Require().True(ok)
	s.Require().Len(approveLogs, 1)
	approveLog, ok := approveLogs[0].(map[string]any)
	s.Require().True(ok)
	s.Require().True(strings.EqualFold(s.token.String(), approveLog["address"].(string)))
	s.Require().Equal(s.approvalTopic, approveLog["topics"].([]any)[0])

	transfer, ok := calls[1].(map[string]any)
	s.Require().True(ok)
	s.Require().Equal("0x1", transfer["status"], "%v", transfer["error"])
	s.Require().Equal("0x"+strings.Repeat("00", 31)+"01", transfer["returnData"])
	transferLogs, ok := transfer["logs"].([]any)
	s.Require().True(ok)
	s.Require().Len(transferLogs, 1)
	transferLog, ok := transferLogs[0].(map[string]any)
	s.Require().True(ok)
	s.Require().Equal(s.transferTopic, transferLog["topics"].([]any)[0])
	s.Require().Equal(encoding.Hex.EncodePrefixed(s.amountWord[:]), transferLog["data"])

	recipientBalance, ok := calls[2].(map[string]any)
	s.Require().True(ok)
	s.Require().Equal(encoding.Hex.EncodePrefixed(s.amountWord[:]), recipientBalance["returnData"])
	ownerBalance, ok := calls[3].(map[string]any)
	s.Require().True(ok)
	s.Require().Equal("0x"+strings.Repeat("00", 32), ownerBalance["returnData"])

	approveGas, err := strconv.ParseUint(strings.TrimPrefix(approve["gasUsed"].(string), "0x"), 16, 64)
	s.Require().NoError(err)
	s.Require().Positive(approveGas)
	transferGas, err := strconv.ParseUint(strings.TrimPrefix(transfer["gasUsed"].(string), "0x"), 16, 64)
	s.Require().NoError(err)
	s.Require().Positive(transferGas)

	s.T().Logf("SimulateV1 chain %d Test01ApproveThenTransferFromRunInOrder Result", s.ChainID)
	s.T().Logf("\tblock %v", blocks[0]["number"])
	s.T().Logf("\toverride: code %d bytes on %s, balance[%x] = %s", len(s.tokenCode), s.token, s.balanceSlot, s.TokenAmount)
	s.T().Logf("\tapprove(%s, %s) from owner: gasUsed %d", s.spender, s.TokenAmount, approveGas)
	s.T().Logf("\ttransferFrom(owner, %s, %s) from spender: gasUsed %d, returnData %v", s.recipient, s.TokenAmount, transferGas, transfer["returnData"])
	s.T().Logf("\tbalanceOf(recipient): %v", recipientBalance["returnData"])
	s.T().Logf("\tbalanceOf(owner): %v", ownerBalance["returnData"])
}

func (s *SimulateV1Suite) Test02RevertedCallReportsErrorAndLaterCallsRun() {
	override := NewStateOverride().
		WithCode(s.token, s.tokenCode).
		WithStateDiff(s.token, s.balanceSlot, s.amountWord)
	payload := map[string]any{
		"blockStateCalls": []any{
			map[string]any{
				"stateOverrides": override,
				"calls": []any{
					map[string]any{"from": s.spender.String(), "to": s.token.String(), "data": s.transferData},
					map[string]any{"to": s.token.String(), "data": s.ownerOfData},
				},
			},
		},
	}

	var blocks []map[string]any
	s.Require().NoError(s.client.Call(s.T().Context(), ETHSimulateV1(uuid.New().String(), payload, BlockTagLatest, &blocks)))
	s.Require().Len(blocks, 1)
	calls, ok := blocks[0]["calls"].([]any)
	s.Require().True(ok)
	s.Require().Len(calls, 2)

	failed, ok := calls[0].(map[string]any)
	s.Require().True(ok)
	s.Require().Equal("0x0", failed["status"])
	callErr, ok := failed["error"].(map[string]any)
	s.Require().True(ok, "%v", failed)
	s.Require().NotZero(callErr["code"])
	s.Require().NotEmpty(callErr["message"])
	insufficientAllowance := abi.NewError("ERC20InsufficientAllowance", abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256)).Selector()
	revertData, ok := callErr["data"].(string)
	s.Require().True(ok, "%v", callErr)
	s.Require().True(strings.HasPrefix(strings.ToLower(revertData), encoding.Hex.EncodePrefixed(insufficientAllowance[:])), revertData)

	ownerBalance, ok := calls[1].(map[string]any)
	s.Require().True(ok)
	s.Require().Equal("0x1", ownerBalance["status"])
	s.Require().Equal(encoding.Hex.EncodePrefixed(s.amountWord[:]), ownerBalance["returnData"])

	s.T().Logf("SimulateV1 chain %d Test02RevertedCallReportsErrorAndLaterCallsRun Result", s.ChainID)
	s.T().Logf("\ttransferFrom without approve: status %v, gasUsed %v", failed["status"], failed["gasUsed"])
	s.T().Logf("\terror: code %v, message %v, data %v", callErr["code"], callErr["message"], callErr["data"])
	s.T().Logf("\tbalanceOf(owner) after the revert: %v", ownerBalance["returnData"])
}

func (s *SimulateV1Suite) Test03TraceTransfersAddsNativeTransferLog() {
	override := NewStateOverride().WithBalance(s.owner, s.OverrideBalance)
	call := map[string]any{"from": s.owner.String(), "to": s.recipient.String(), "value": "0x" + s.TransferValue.Text(16)}

	var traced, plain []map[string]any
	s.Require().NoError(s.client.Batch(s.T().Context(), Elements{
		ETHSimulateV1(uuid.New().String(), map[string]any{
			"blockStateCalls": []any{map[string]any{"stateOverrides": override, "calls": []any{call}}},
			"traceTransfers":  true,
		}, BlockTagLatest, &traced),
		ETHSimulateV1(uuid.New().String(), map[string]any{
			"blockStateCalls": []any{map[string]any{"stateOverrides": override, "calls": []any{call}}},
		}, BlockTagLatest, &plain),
	}))

	s.Require().Len(traced, 1)
	tracedCalls, ok := traced[0]["calls"].([]any)
	s.Require().True(ok)
	s.Require().Len(tracedCalls, 1)
	tracedCall, ok := tracedCalls[0].(map[string]any)
	s.Require().True(ok)
	s.Require().Equal("0x1", tracedCall["status"], "%v", tracedCall["error"])
	tracedLogs, ok := tracedCall["logs"].([]any)
	s.Require().True(ok)
	s.Require().Len(tracedLogs, 1)
	nativeLog, ok := tracedLogs[0].(map[string]any)
	s.Require().True(ok)
	topics, ok := nativeLog["topics"].([]any)
	s.Require().True(ok)
	s.Require().Len(topics, 3)
	s.Require().Equal(s.transferTopic, topics[0])
	s.Require().True(strings.EqualFold("0x"+strings.Repeat("00", 12)+strings.TrimPrefix(s.owner.String(), "0x"), topics[1].(string)))
	s.Require().True(strings.EqualFold("0x"+strings.Repeat("00", 12)+strings.TrimPrefix(s.recipient.String(), "0x"), topics[2].(string)))
	var valueWord [32]byte
	s.TransferValue.FillBytes(valueWord[:])
	s.Require().Equal(encoding.Hex.EncodePrefixed(valueWord[:]), nativeLog["data"])

	s.Require().Len(plain, 1)
	plainCalls, ok := plain[0]["calls"].([]any)
	s.Require().True(ok)
	s.Require().Len(plainCalls, 1)
	plainCall, ok := plainCalls[0].(map[string]any)
	s.Require().True(ok)
	s.Require().Equal("0x1", plainCall["status"], "%v", plainCall["error"])

	s.T().Logf("SimulateV1 chain %d Test03TraceTransfersAddsNativeTransferLog Result", s.ChainID)
	s.T().Logf("\toverride: balance %s on %s", s.OverrideBalance, s.owner)
	s.T().Logf("\tsend %s wei to %s", s.TransferValue, s.recipient)
	s.T().Logf("\ttraceTransfers=true log: address %v, topics %v, data %v", nativeLog["address"], topics, nativeLog["data"])
	s.T().Logf("\ttraceTransfers=false logs: %v", plainCall["logs"])
}

func (s *SimulateV1Suite) Test04StateCarriesAcrossBlockStateCalls() {
	override := NewStateOverride().
		WithCode(s.token, s.tokenCode).
		WithStateDiff(s.token, s.balanceSlot, s.amountWord)
	payload := map[string]any{
		"blockStateCalls": []any{
			map[string]any{
				"stateOverrides": override,
				"calls":          []any{map[string]any{"from": s.owner.String(), "to": s.token.String(), "data": s.approveData}},
			},
			map[string]any{
				"calls": []any{map[string]any{"from": s.spender.String(), "to": s.token.String(), "data": s.transferData}},
			},
		},
	}

	var blocks []map[string]any
	s.Require().NoError(s.client.Call(s.T().Context(), ETHSimulateV1(uuid.New().String(), payload, BlockTagLatest, &blocks)))
	s.Require().Len(blocks, 2)

	first, err := strconv.ParseUint(strings.TrimPrefix(blocks[0]["number"].(string), "0x"), 16, 64)
	s.Require().NoError(err)
	second, err := strconv.ParseUint(strings.TrimPrefix(blocks[1]["number"].(string), "0x"), 16, 64)
	s.Require().NoError(err)
	s.Require().Equal(first+1, second)

	approveCalls, ok := blocks[0]["calls"].([]any)
	s.Require().True(ok)
	s.Require().Len(approveCalls, 1)
	approve, ok := approveCalls[0].(map[string]any)
	s.Require().True(ok)
	s.Require().Equal("0x1", approve["status"], "%v", approve["error"])

	transferCalls, ok := blocks[1]["calls"].([]any)
	s.Require().True(ok)
	s.Require().Len(transferCalls, 1)
	transfer, ok := transferCalls[0].(map[string]any)
	s.Require().True(ok)
	s.Require().Equal("0x1", transfer["status"], "%v", transfer["error"])
	s.Require().Equal("0x"+strings.Repeat("00", 31)+"01", transfer["returnData"])

	s.T().Logf("SimulateV1 chain %d Test04StateCarriesAcrossBlockStateCalls Result", s.ChainID)
	s.T().Logf("\tblock %d: approve status %v, gasUsed %v", first, approve["status"], approve["gasUsed"])
	s.T().Logf("\tblock %d: transferFrom status %v, gasUsed %v, returnData %v", second, transfer["status"], transfer["gasUsed"], transfer["returnData"])
}
