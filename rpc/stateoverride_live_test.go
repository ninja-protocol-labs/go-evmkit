package rpc

import (
	"math/big"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/internal/testenv"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

type StateOverrideSuite struct {
	suite.Suite

	RPCURL           string
	WETHHex          string
	Permit2Hex       string
	SlotReaderCode   string
	BalanceCode      string
	CreateCode       string
	GuardCode        string
	TargetSeed       string
	SenderSeed       string
	RecipientSeed    string
	OverrideBalance  *big.Int
	TransferValue    *big.Int
	StorageValue     *big.Int
	OverrideNonce    uint64
	HistoricalOffset uint64

	client    Client
	target    *types.Address
	sender    *types.Address
	recipient *types.Address
	weth      *types.Address
	permit2   *types.Address

	slotReader []byte
	balanceOf  []byte
	creator    []byte
	guard      []byte
	slot0      [32]byte
	slot1      [32]byte
	value      [32]byte
	one        [32]byte
}

func TestStateOverrideSuite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)

	suite.Run(t, &StateOverrideSuite{
		RPCURL:           c.RPCURL,
		WETHHex:          c.Addresses.Standard.WETH,
		Permit2Hex:       c.Addresses.Standard.Permit2,
		SlotReaderCode:   "0x60005460005260206000f3",
		BalanceCode:      "0x303160005260206000f3",
		CreateCode:       "0x600060006000f060005260206000f3",
		GuardCode:        "0x600054600114600e5760006000fd5b00",
		TargetSeed:       "go-evmkit-rpc-stateoverride-target",
		SenderSeed:       "go-evmkit-rpc-stateoverride-sender",
		RecipientSeed:    "go-evmkit-rpc-stateoverride-recipient",
		OverrideBalance:  new(big.Int).Add(new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil), big.NewInt(7)),
		TransferValue:    new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
		StorageValue:     big.NewInt(0x2a),
		OverrideNonce:    5,
		HistoricalOffset: 10,
	})
}

func (s *StateOverrideSuite) SetupSuite() {
	s.client = NewClient(s.RPCURL, WithRetry(5, ExponentialBackoff(500*time.Millisecond, 10*time.Second)))

	s.target = types.NewAddressFromBytes(core.Keccak256([]byte(s.TargetSeed)).Bytes())
	s.sender = types.NewAddressFromBytes(core.Keccak256([]byte(s.SenderSeed)).Bytes())
	s.recipient = types.NewAddressFromBytes(core.Keccak256([]byte(s.RecipientSeed)).Bytes())

	var err error
	if s.WETHHex != "" {
		s.weth, err = types.NewAddressFromHex(s.WETHHex)
		s.Require().NoError(err)
	}
	if s.Permit2Hex != "" {
		s.permit2, err = types.NewAddressFromHex(s.Permit2Hex)
		s.Require().NoError(err)
	}

	s.slotReader, err = encoding.Hex.Decode(s.SlotReaderCode)
	s.Require().NoError(err)
	s.balanceOf, err = encoding.Hex.Decode(s.BalanceCode)
	s.Require().NoError(err)
	s.creator, err = encoding.Hex.Decode(s.CreateCode)
	s.Require().NoError(err)
	s.guard, err = encoding.Hex.Decode(s.GuardCode)
	s.Require().NoError(err)

	s.slot1[31] = 0x01
	s.StorageValue.FillBytes(s.value[:])
	s.one[31] = 0x01

	var code string
	e := ETHGetCode("code", s.target.String(), BlockTagLatest, &code)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Equal("0x", code, "target %s must be an empty account", s.target)
	e = ETHGetCode("code", s.sender.String(), BlockTagLatest, &code)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Equal("0x", code, "sender %s must be an empty account", s.sender)

	s.T().Logf("SetupSuite: target %s, sender %s, recipient %s", s.target, s.sender, s.recipient)
	s.T().Logf("SetupSuite: WETH %s, Permit2 %s", s.WETHHex, s.Permit2Hex)
	s.T().Logf("SetupSuite: slot reader code %s", s.SlotReaderCode)
	s.T().Logf("SetupSuite: balance code %s", s.BalanceCode)
	s.T().Logf("SetupSuite: create code %s", s.CreateCode)
	s.T().Logf("SetupSuite: guard code %s", s.GuardCode)
	s.T().Logf("SetupSuite: storage value %s", s.StorageValue)
	s.T().Logf("SetupSuite: override balance %s, nonce %d", s.OverrideBalance, s.OverrideNonce)
}

func (s *StateOverrideSuite) Test01CallWithoutOverrideSeesEmptyAccount() {
	var result string
	params := map[string]any{
		"to": s.target.String(),
	}
	e := ETHCall("call", params, BlockTagLatest, nil, &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Equal("0x", result)

	var balance string
	e = ETHGetBalance("balance", s.sender.String(), BlockTagLatest, &balance)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Equal("0x0", balance)
	s.T().Logf("StateOverride target %s Test01CallWithoutOverrideSeesEmptyAccount Result", s.target)
	s.T().Logf("\teth_call(%s): %s", s.target, result)
	s.T().Logf("\teth_getBalance(%s): %s", s.sender, balance)
}

func (s *StateOverrideSuite) Test02CallReadsStateDiffSlot() {
	override := NewStateOverride().
		WithCode(s.target, s.slotReader).
		WithStateDiff(s.target, s.slot0, s.value)

	var result string
	params := map[string]any{
		"to": s.target.String(),
	}
	e := ETHCall("call", params, BlockTagLatest, override, &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Equal(encoding.Hex.EncodePrefixed(s.value[:]), result)
	s.T().Logf("StateOverride target %s Test02CallReadsStateDiffSlot Result", s.target)
	overrideJSON, err := override.MarshalJSON()
	s.Require().NoError(err)
	s.T().Logf("\toverride: %s", overrideJSON)
	s.T().Logf("\tSLOAD(0): %s", result)
}

func (s *StateOverrideSuite) Test03CallReadsStateSlot() {
	override := NewStateOverride().
		WithCode(s.target, s.slotReader).
		WithState(s.target, s.slot0, s.value)

	var result string
	params := map[string]any{
		"to": s.target.String(),
	}
	e := ETHCall("call", params, BlockTagLatest, override, &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Equal(encoding.Hex.EncodePrefixed(s.value[:]), result)
	s.T().Logf("StateOverride target %s Test03CallReadsStateSlot Result", s.target)
	overrideJSON, err := override.MarshalJSON()
	s.Require().NoError(err)
	s.T().Logf("\toverride: %s", overrideJSON)
	s.T().Logf("\tSLOAD(0): %s", result)
}

func (s *StateOverrideSuite) Test04CallReadsOverriddenBalance() {
	override := NewStateOverride().
		WithCode(s.target, s.balanceOf).
		WithBalance(s.target, s.OverrideBalance)

	var result string
	params := map[string]any{
		"to": s.target.String(),
	}
	e := ETHCall("call", params, BlockTagLatest, override, &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	out, err := encoding.Hex.Decode(result)
	s.Require().NoError(err)
	s.Require().Zero(s.OverrideBalance.Cmp(new(big.Int).SetBytes(out)))
	s.T().Logf("StateOverride target %s Test04CallReadsOverriddenBalance Result", s.target)
	overrideJSON, err := override.MarshalJSON()
	s.Require().NoError(err)
	s.T().Logf("\toverride: %s", overrideJSON)
	s.T().Logf("\tBALANCE(ADDRESS): %s (%s)", result, new(big.Int).SetBytes(out))
}

func (s *StateOverrideSuite) Test05CreateUsesOverriddenNonce() {
	override := NewStateOverride().
		WithCode(s.target, s.creator).
		WithNonce(s.target, s.OverrideNonce)

	var result string
	params := map[string]any{
		"to": s.target.String(),
	}
	e := ETHCall("call", params, BlockTagLatest, override, &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	out, err := encoding.Hex.Decode(result)
	s.Require().NoError(err)
	s.Require().Len(out, 32)

	want, err := core.CreateAddress(s.target, s.OverrideNonce)
	s.Require().NoError(err)
	s.Require().True(want.Equal(types.NewAddressFromBytes(out)), "created %x, want %s", out, want)

	other, err := core.CreateAddress(s.target, s.OverrideNonce+1)
	s.Require().NoError(err)
	s.Require().False(other.Equal(types.NewAddressFromBytes(out)))
	s.T().Logf("StateOverride target %s Test05CreateUsesOverriddenNonce Result", s.target)
	overrideJSON, err := override.MarshalJSON()
	s.Require().NoError(err)
	s.T().Logf("\toverride: %s", overrideJSON)
	s.T().Logf("\tCREATE returned %s", types.NewAddressFromBytes(out))
	s.T().Logf("\tCreateAddress(target, %d) = %s, (target, %d) = %s", s.OverrideNonce, want, s.OverrideNonce+1, other)
}

func (s *StateOverrideSuite) Test06EmptyCodeClearsDeployedContract() {
	if s.permit2 == nil {
		s.T().Skip("testenv: no Permit2 configured for this network")
	}
	params := map[string]any{
		"to":   s.permit2.String(),
		"data": "0x3644e515",
	}

	var plain string
	e := ETHCall("call", params, BlockTagLatest, nil, &plain)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Len(plain, 66)

	var cleared string
	override := NewStateOverride().WithCode(s.permit2, []byte{})
	e = ETHCall("call", params, BlockTagLatest, override, &cleared)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Equal("0x", cleared)
	s.T().Logf("StateOverride target %s Test06EmptyCodeClearsDeployedContract Result", s.target)
	overrideJSON, err := override.MarshalJSON()
	s.Require().NoError(err)
	s.T().Logf("\tDOMAIN_SEPARATOR() on %s: %s", s.permit2, plain)
	s.T().Logf("\toverride: %s", overrideJSON)
	s.T().Logf("\tDOMAIN_SEPARATOR() with code cleared: %s", cleared)
}

func (s *StateOverrideSuite) Test07StateDiffKeepsAndStateZeroesUntouchedSlots() {
	if s.weth == nil {
		s.T().Skip("testenv: no WETH configured for this network")
	}
	var stored string
	e := ETHGetStorageAt("slot0", s.weth.String(), "0x0", BlockTagLatest, &stored)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	storedBytes, err := encoding.Hex.Decode(stored)
	s.Require().NoError(err)
	if new(big.Int).SetBytes(storedBytes).Sign() == 0 {
		s.T().Skip("WETH slot 0 is empty on this network")
	}
	params := map[string]any{
		"to": s.weth.String(),
	}

	var kept string
	diff := NewStateOverride().
		WithCode(s.weth, s.slotReader).
		WithStateDiff(s.weth, s.slot1, s.value)
	e = ETHCall("call", params, BlockTagLatest, diff, &kept)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	keptBytes, err := encoding.Hex.Decode(kept)
	s.Require().NoError(err)
	s.Require().Zero(new(big.Int).SetBytes(storedBytes).Cmp(new(big.Int).SetBytes(keptBytes)))

	var zeroed string
	full := NewStateOverride().
		WithCode(s.weth, s.slotReader).
		WithState(s.weth, s.slot1, s.value)
	e = ETHCall("call", params, BlockTagLatest, full, &zeroed)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Equal("0x"+strings.Repeat("00", 32), zeroed)
	s.T().Logf("StateOverride target %s Test07StateDiffKeepsAndStateZeroesUntouchedSlots Result", s.target)
	diffJSON, err := diff.MarshalJSON()
	s.Require().NoError(err)
	fullJSON, err := full.MarshalJSON()
	s.Require().NoError(err)
	s.T().Logf("\teth_getStorageAt(%s, 0x0): %s", s.weth, stored)
	s.T().Logf("\tstateDiff override: %s", diffJSON)
	s.T().Logf("\tSLOAD(0) with stateDiff on slot 1: %s", kept)
	s.T().Logf("\tstate override: %s", fullJSON)
	s.T().Logf("\tSLOAD(0) with state on slot 1: %s", zeroed)
}

func (s *StateOverrideSuite) Test08EstimateGasFailsWhenGuardIsUnset() {
	var result string
	params := map[string]any{
		"from": s.sender.String(),
		"to":   s.target.String(),
	}
	override := NewStateOverride().WithCode(s.target, s.guard)
	e := ETHEstimateGas("gas", params, BlockTagLatest, override, &result)
	err := s.client.Call(s.T().Context(), e)
	s.Require().Error(err)
	s.T().Logf("StateOverride target %s Test08EstimateGasFailsWhenGuardIsUnset Result", s.target)
	overrideJSON, marshalErr := override.MarshalJSON()
	s.Require().NoError(marshalErr)
	s.T().Logf("\toverride: %s", overrideJSON)
	s.T().Logf("\teth_estimateGas(%s -> %s): %v", s.sender, s.target, err)
}

func (s *StateOverrideSuite) Test09EstimateGasPassesGuardWithStateDiff() {
	override := NewStateOverride().
		WithCode(s.target, s.guard).
		WithStateDiff(s.target, s.slot0, s.one)

	var result string
	params := map[string]any{
		"from": s.sender.String(),
		"to":   s.target.String(),
	}
	e := ETHEstimateGas("gas", params, BlockTagLatest, override, &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	gas, err := strconv.ParseUint(strings.TrimPrefix(result, "0x"), 16, 64)
	s.Require().NoError(err)
	s.Require().Positive(gas)
	s.T().Logf("StateOverride target %s Test09EstimateGasPassesGuardWithStateDiff Result", s.target)
	overrideJSON, err := override.MarshalJSON()
	s.Require().NoError(err)
	s.T().Logf("\toverride: %s", overrideJSON)
	s.T().Logf("\teth_estimateGas(%s -> %s): %s (%d gas)", s.sender, s.target, result, gas)
}

func (s *StateOverrideSuite) Test10EstimateGasFundsSenderWithBalanceOverride() {
	params := map[string]any{
		"from":  s.sender.String(),
		"to":    s.recipient.String(),
		"value": "0x" + s.TransferValue.Text(16),
	}

	var unfunded string
	e := ETHEstimateGas("gas", params, BlockTagLatest, nil, &unfunded)
	unfundedErr := s.client.Call(s.T().Context(), e)
	s.Require().Error(unfundedErr)

	var funded string
	override := NewStateOverride().WithBalance(s.sender, s.OverrideBalance)
	e = ETHEstimateGas("gas", params, BlockTagLatest, override, &funded)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	gas, err := strconv.ParseUint(strings.TrimPrefix(funded, "0x"), 16, 64)
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(gas, uint64(21_000))
	s.T().Logf("StateOverride target %s Test10EstimateGasFundsSenderWithBalanceOverride Result", s.target)
	overrideJSON, err := override.MarshalJSON()
	s.Require().NoError(err)
	s.T().Logf("\ttransfer: %s -> %s, %s wei", s.sender, s.recipient, s.TransferValue)
	s.T().Logf("\twithout override: %v", unfundedErr)
	s.T().Logf("\toverride: %s", overrideJSON)
	s.T().Logf("\twith override: %s (%d gas)", funded, gas)
}

func (s *StateOverrideSuite) Test11BatchAppliesOverridePerElement() {
	override := NewStateOverride().
		WithCode(s.target, s.slotReader).
		WithStateDiff(s.target, s.slot0, s.value)
	params := map[string]any{
		"to": s.target.String(),
	}

	var overridden, plain string
	s.Require().NoError(s.client.Batch(s.T().Context(), Elements{
		ETHCall("overridden", params, BlockTagLatest, override, &overridden),
		ETHCall("plain", params, BlockTagLatest, nil, &plain),
	}))
	s.Require().Equal(encoding.Hex.EncodePrefixed(s.value[:]), overridden)
	s.Require().Equal("0x", plain)
	s.T().Logf("StateOverride target %s Test11BatchAppliesOverridePerElement Result", s.target)
	overrideJSON, err := override.MarshalJSON()
	s.Require().NoError(err)
	s.T().Logf("\toverride on element 0: %s", overrideJSON)
	s.T().Logf("\telement 0 (override): %s", overridden)
	s.T().Logf("\telement 1 (plain): %s", plain)
}

func (s *StateOverrideSuite) Test12OverrideAppliesAtHistoricalBlock() {
	var head string
	e := ETHBlockNumber("head", &head)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	number, err := strconv.ParseUint(strings.TrimPrefix(head, "0x"), 16, 64)
	s.Require().NoError(err)
	s.Require().Greater(number, s.HistoricalOffset)
	block := "0x" + strconv.FormatUint(number-s.HistoricalOffset, 16)

	override := NewStateOverride().
		WithCode(s.target, s.slotReader).
		WithStateDiff(s.target, s.slot0, s.value)
	var result string
	params := map[string]any{
		"to": s.target.String(),
	}
	e = ETHCall("call", params, block, override, &result)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	s.Require().Equal(encoding.Hex.EncodePrefixed(s.value[:]), result)
	s.T().Logf("StateOverride target %s Test12OverrideAppliesAtHistoricalBlock Result", s.target)
	overrideJSON, err := override.MarshalJSON()
	s.Require().NoError(err)
	s.T().Logf("\thead %s, queried block %s (head - %d)", head, block, s.HistoricalOffset)
	s.T().Logf("\toverride: %s", overrideJSON)
	s.T().Logf("\tSLOAD(0) at %s: %s", block, result)
}
