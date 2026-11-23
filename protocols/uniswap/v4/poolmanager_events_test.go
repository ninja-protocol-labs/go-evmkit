package v4

import (
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

var (
	testPoolManager = types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	testSender      = types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	testCurrency0   = types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	testCurrency1   = types.MustNewAddressFromHex("0x1231deb6f5749ef6ce6943a275a1d3e7486f4eae")
	testHooks       = types.MustNewAddressFromHex("0xb47e3cd837ddf8e4c57f05d70ab865de6e193bbb")
)

func TestExtractInitializeSampleReceipt(t *testing.T) {
	topic0, ok := initializeEvent.Topic0()
	require.True(t, ok)
	data, err := abi.Pack(abi.Types{abi.Uint24, abi.Int24, abi.Address, abi.Uint160, abi.Int24},
		big.NewInt(3000), big.NewInt(60), testHooks, big.NewInt(792281625142643375), big.NewInt(0))
	require.NoError(t, err)

	logs := []tx.Log{
		{
			Address: testPoolManager,
			Topics: []*types.Hash{
				types.NewHashFromBytes(topic0[:]),
				types.NewHashFromBytes([]byte{0x01}),
				types.NewHashFromBytes(testCurrency0.Bytes()),
				types.NewHashFromBytes(testCurrency1.Bytes()),
			},
			Data: data,
		},
	}

	inits, err := ExtractInitialize(logs)
	require.NoError(t, err)
	require.Len(t, inits, 1)
	require.True(t, inits[0].Contract.Equal(testPoolManager))
	require.True(t, inits[0].Currency0.Equal(testCurrency0))
	require.True(t, inits[0].Currency1.Equal(testCurrency1))
	require.Zero(t, big.NewInt(3000).Cmp(inits[0].Fee))
	require.Zero(t, big.NewInt(60).Cmp(inits[0].TickSpacing))
	require.True(t, inits[0].Hooks.Equal(testHooks))
	require.Zero(t, big.NewInt(792281625142643375).Cmp(inits[0].SqrtPriceX96))
	require.Zero(t, big.NewInt(0).Cmp(inits[0].Tick))
}

func TestExtractModifyLiquiditySampleReceipt(t *testing.T) {
	topic0, ok := modifyLiquidityEvent.Topic0()
	require.True(t, ok)
	salt := make([]byte, 32)
	salt[0] = 0x0a
	data, err := abi.Pack(abi.Types{abi.Int24, abi.Int24, abi.Int256, abi.Bytes32},
		big.NewInt(-100), big.NewInt(100), big.NewInt(1000), salt)
	require.NoError(t, err)

	logs := []tx.Log{
		{
			Address: testPoolManager,
			Topics: []*types.Hash{
				types.NewHashFromBytes(topic0[:]),
				types.NewHashFromBytes([]byte{0x02}),
				types.NewHashFromBytes(testSender.Bytes()),
			},
			Data: data,
		},
	}

	mods, err := ExtractModifyLiquidity(logs)
	require.NoError(t, err)
	require.Len(t, mods, 1)
	require.True(t, mods[0].Contract.Equal(testPoolManager))
	require.True(t, mods[0].Sender.Equal(testSender))
	require.Zero(t, big.NewInt(-100).Cmp(mods[0].TickLower))
	require.Zero(t, big.NewInt(100).Cmp(mods[0].TickUpper))
	require.Zero(t, big.NewInt(1000).Cmp(mods[0].LiquidityDelta))
	require.Equal(t, [32]byte{0x0a}, mods[0].Salt)
}

func TestExtractSwapSampleReceipt(t *testing.T) {
	topic0, ok := swapEvent.Topic0()
	require.True(t, ok)
	data, err := abi.Pack(abi.Types{abi.Int128, abi.Int128, abi.Uint160, abi.Uint128, abi.Int24, abi.Uint24},
		big.NewInt(100), big.NewInt(-50), big.NewInt(792281625142643375), big.NewInt(1000), big.NewInt(42), big.NewInt(3000))
	require.NoError(t, err)

	logs := []tx.Log{
		{
			Address: testPoolManager,
			Topics: []*types.Hash{
				types.NewHashFromBytes(topic0[:]),
				types.NewHashFromBytes([]byte{0x05}),
				types.NewHashFromBytes(testSender.Bytes()),
			},
			Data: data,
		},
	}

	swaps, err := ExtractSwap(logs)
	require.NoError(t, err)
	require.Len(t, swaps, 1)
	require.True(t, swaps[0].Contract.Equal(testPoolManager))
	require.True(t, swaps[0].Sender.Equal(testSender))
	require.Zero(t, big.NewInt(100).Cmp(swaps[0].Amount0))
	require.Zero(t, big.NewInt(-50).Cmp(swaps[0].Amount1))
	require.Zero(t, big.NewInt(792281625142643375).Cmp(swaps[0].SqrtPriceX96))
	require.Zero(t, big.NewInt(1000).Cmp(swaps[0].Liquidity))
	require.Zero(t, big.NewInt(42).Cmp(swaps[0].Tick))
	require.Zero(t, big.NewInt(3000).Cmp(swaps[0].Fee))
}

func TestExtractDonateSampleReceipt(t *testing.T) {
	topic0, ok := donateEvent.Topic0()
	require.True(t, ok)
	data, err := abi.Pack(abi.Types{abi.Uint256, abi.Uint256}, big.NewInt(100), big.NewInt(200))
	require.NoError(t, err)

	logs := []tx.Log{
		{
			Address: testPoolManager,
			Topics: []*types.Hash{
				types.NewHashFromBytes(topic0[:]),
				types.NewHashFromBytes([]byte{0x06}),
				types.NewHashFromBytes(testSender.Bytes()),
			},
			Data: data,
		},
	}

	donates, err := ExtractDonate(logs)
	require.NoError(t, err)
	require.Len(t, donates, 1)
	require.True(t, donates[0].Contract.Equal(testPoolManager))
	require.True(t, donates[0].Sender.Equal(testSender))
	require.Zero(t, big.NewInt(100).Cmp(donates[0].Amount0))
	require.Zero(t, big.NewInt(200).Cmp(donates[0].Amount1))
}

func TestExtractInitializeEmptyLogs(t *testing.T) {
	inits, err := ExtractInitialize(nil)
	require.NoError(t, err)
	require.Empty(t, inits)
}

func TestExtractInitializeNoTopics(t *testing.T) {
	logs := []tx.Log{{Address: testPoolManager, Topics: nil, Data: nil}}
	inits, err := ExtractInitialize(logs)
	require.NoError(t, err)
	require.Empty(t, inits)
}

func TestExtractInitializeUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{
		Address: testPoolManager,
		Topics:  []*types.Hash{unrelated},
		Data:    []byte{0xde, 0xad},
	}}
	inits, err := ExtractInitialize(logs)
	require.NoError(t, err)
	require.Empty(t, inits)
}

func TestExtractInitializeRejectsNilTopic(t *testing.T) {
	logs := []tx.Log{{
		Address: testPoolManager,
		Topics:  []*types.Hash{nil},
		Data:    nil,
	}}
	inits, err := ExtractInitialize(logs)
	require.NoError(t, err)
	require.Empty(t, inits)

	t0, ok := initializeEvent.Topic0()
	require.True(t, ok)
	logs = []tx.Log{{
		Address: testPoolManager,
		Topics:  []*types.Hash{types.NewHashFromBytes(t0[:]), nil, nil, nil},
		Data:    nil,
	}}
	_, err = ExtractInitialize(logs)
	require.Error(t, err)
}

func TestExtractInitializeRejectsMalformedTopicCount(t *testing.T) {
	t0, ok := initializeEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testPoolManager,
		Topics: []*types.Hash{
			types.NewHashFromBytes(t0[:]),
			types.NewHashFromBytes([]byte{0x01}),
			types.NewHashFromBytes(testCurrency0.Bytes()),
		},
		Data: nil,
	}}
	_, err := ExtractInitialize(logs)
	require.Error(t, err)
}

func TestExtractModifyLiquidityRejectsMalformedData(t *testing.T) {
	t0, ok := modifyLiquidityEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testPoolManager,
		Topics: []*types.Hash{
			types.NewHashFromBytes(t0[:]),
			types.NewHashFromBytes([]byte{0x02}),
			types.NewHashFromBytes(testSender.Bytes()),
		},
		Data: []byte{0x01, 0x02},
	}}
	_, err := ExtractModifyLiquidity(logs)
	require.Error(t, err)
}

func TestExtractSwapRejectsMalformedData(t *testing.T) {
	t0, ok := swapEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testPoolManager,
		Topics: []*types.Hash{
			types.NewHashFromBytes(t0[:]),
			types.NewHashFromBytes([]byte{0x05}),
			types.NewHashFromBytes(testSender.Bytes()),
		},
		Data: []byte{0x01},
	}}

	require.NotPanics(t, func() {
		_, err := ExtractSwap(logs)
		require.Error(t, err)
	})
}

func TestExtractDonateRejectsMalformedData(t *testing.T) {
	t0, ok := donateEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testPoolManager,
		Topics: []*types.Hash{
			types.NewHashFromBytes(t0[:]),
			types.NewHashFromBytes([]byte{0x06}),
			types.NewHashFromBytes(testSender.Bytes()),
		},
		Data: []byte{0x01, 0x02},
	}}
	_, err := ExtractDonate(logs)
	require.Error(t, err)
}

func TestExtractProtocolFeeControllerUpdatedSampleReceipt(t *testing.T) {
	topic0, ok := protocolFeeControllerUpdatedEvent.Topic0()
	require.True(t, ok)

	logs := []tx.Log{
		{
			Address: testPoolManager,
			Topics: []*types.Hash{
				types.NewHashFromBytes(topic0[:]),
				types.NewHashFromBytes(testSender.Bytes()),
			},
			Data: nil,
		},
	}

	updates, err := ExtractProtocolFeeControllerUpdated(logs)
	require.NoError(t, err)
	require.Len(t, updates, 1)
	require.True(t, updates[0].Contract.Equal(testPoolManager))
	require.True(t, updates[0].ProtocolFeeController.Equal(testSender))
}

func TestExtractProtocolFeeControllerUpdatedEmptyLogs(t *testing.T) {
	updates, err := ExtractProtocolFeeControllerUpdated(nil)
	require.NoError(t, err)
	require.Empty(t, updates)
}

func TestExtractProtocolFeeControllerUpdatedUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{
		Address: testPoolManager,
		Topics:  []*types.Hash{unrelated},
		Data:    []byte{0xde, 0xad},
	}}
	updates, err := ExtractProtocolFeeControllerUpdated(logs)
	require.NoError(t, err)
	require.Empty(t, updates)
}

func TestExtractProtocolFeeControllerUpdatedRejectsMalformedTopicCount(t *testing.T) {
	t0, ok := protocolFeeControllerUpdatedEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testPoolManager,
		Topics: []*types.Hash{
			types.NewHashFromBytes(t0[:]),
			types.NewHashFromBytes(testSender.Bytes()),
			types.NewHashFromBytes(testSender.Bytes()),
		},
		Data: nil,
	}}
	_, err := ExtractProtocolFeeControllerUpdated(logs)
	require.Error(t, err)
}

func TestExtractProtocolFeeUpdatedSampleReceipt(t *testing.T) {
	topic0, ok := protocolFeeUpdatedEvent.Topic0()
	require.True(t, ok)
	data, err := abi.Pack(abi.Types{abi.Uint24}, big.NewInt(500))
	require.NoError(t, err)

	logs := []tx.Log{
		{
			Address: testPoolManager,
			Topics: []*types.Hash{
				types.NewHashFromBytes(topic0[:]),
				types.NewHashFromBytes([]byte{0x01}),
			},
			Data: data,
		},
	}

	updates, err := ExtractProtocolFeeUpdated(logs)
	require.NoError(t, err)
	require.Len(t, updates, 1)
	require.True(t, updates[0].Contract.Equal(testPoolManager))
	require.Zero(t, big.NewInt(500).Cmp(updates[0].ProtocolFee))
}

func TestExtractProtocolFeeUpdatedRejectsMalformedData(t *testing.T) {
	t0, ok := protocolFeeUpdatedEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testPoolManager,
		Topics: []*types.Hash{
			types.NewHashFromBytes(t0[:]),
			types.NewHashFromBytes([]byte{0x01}),
		},
		Data: []byte{0x01, 0x02},
	}}
	_, err := ExtractProtocolFeeUpdated(logs)
	require.Error(t, err)
}

func TestExtractProtocolFeeUpdatedRejectsNilTopic(t *testing.T) {
	logs := []tx.Log{{
		Address: testPoolManager,
		Topics:  []*types.Hash{nil},
		Data:    nil,
	}}
	updates, err := ExtractProtocolFeeUpdated(logs)
	require.NoError(t, err)
	require.Empty(t, updates)

	t0, ok := protocolFeeUpdatedEvent.Topic0()
	require.True(t, ok)
	logs = []tx.Log{{
		Address: testPoolManager,
		Topics:  []*types.Hash{types.NewHashFromBytes(t0[:]), nil},
		Data:    nil,
	}}
	_, err = ExtractProtocolFeeUpdated(logs)
	require.Error(t, err)
}
