package erc20

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

func mustAddrPanic(s string) *types.Address {
	a, err := types.NewAddressFromHex(s)
	if err != nil {
		panic(err)
	}
	return a
}

var (
	fuzzToken1 = mustAddrPanic("0x4200000000000000000000000000000000000006")
	fuzzToken2 = mustAddrPanic("0x833589fcd6edb6e08f4c7c32d4f71b54bda02913")
	fuzzAddr1  = mustAddrPanic("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	fuzzAddr2  = mustAddrPanic("0xba12222222228d8ba445958a75a0704d566bf2c8")
)

// addrTopic left-pads a's 20 bytes into a 32-byte topic, as the EVM encodes an indexed address.
func addrTopic(a *types.Address) *types.Hash {
	b := make([]byte, 32)
	copy(b[12:], a.Bytes())
	return types.NewHashFromBytes(b)
}

// go test -run '^$' -fuzz '^FuzzSimpleDecodersDoNotPanic$' -fuzztime=10s ./contracts/erc20
func FuzzSimpleDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 33))
	f.Add(make([]byte, 64))

	nameData, err := abi.Pack(abi.Types{abi.String}, "USD Coin")
	require.NoError(f, err)
	f.Add(nameData)

	uintData, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(123456789))
	require.NoError(f, err)
	f.Add(uintData)

	boolData, err := abi.Pack(abi.Types{abi.Bool}, true)
	require.NoError(f, err)
	f.Add(boolData)

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeName(data)
		_, _ = DecodeSymbol(data)
		_, _ = DecodeDecimals(data)
		_, _ = DecodeTotalSupply(data)
		_, _ = DecodeBalanceOf(data)
		_, _ = DecodeAllowance(data)
		_, _ = DecodeTransfer(data)
		_, _ = DecodeApprove(data)
		_, _ = DecodeTransferFrom(data)
	})
}

// go test -run '^$' -fuzz '^FuzzBatchDecodersDoNotPanic$' -fuzztime=10s ./contracts/erc20
func FuzzBatchDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 192))

	if seed, err := EncodeMetadataWithBalance(fuzzToken1, fuzzAddr1, false); err == nil {
		f.Add(seed)
	}
	if seed, err := EncodeAllowanceWithBalance(fuzzToken1, fuzzAddr1, fuzzAddr2, false); err == nil {
		f.Add(seed)
	}

	addrs := []*types.Address{fuzzAddr1, fuzzAddr2}
	toks := []*types.Address{fuzzToken1, fuzzToken2}

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeMetadata(data)
		_, _ = DecodeMetadataWithBalance(data)
		_, _ = DecodeAllowanceWithBalance(data)
		_, _ = DecodeTokenBalances(fuzzToken1, addrs, data)
		_, _ = DecodeAddressBalances(fuzzAddr1, toks, data)
		_, _ = DecodeBalancePairs(toks, addrs, data)
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeBalanceOfRoundTrip$' -fuzztime=10s ./contracts/erc20
func FuzzDecodeBalanceOfRoundTrip(f *testing.F) {
	f.Add([]byte{0x01})
	f.Add([]byte{})
	f.Add(make([]byte, 32))

	f.Fuzz(func(t *testing.T, raw []byte) {
		value := clampUint256(raw)

		data, err := abi.Pack(abi.Types{abi.Uint256}, value)
		require.NoError(t, err)

		got, err := DecodeBalanceOf(data)
		require.NoError(t, err)
		require.Zero(t, value.Cmp(got))
	})
}

// go test -run '^$' -fuzz '^FuzzExtractTransfersCrashSafety$' -fuzztime=10s ./contracts/erc20
func FuzzExtractTransfersCrashSafety(f *testing.F) {
	t0, ok := transferEvent.Topic0()
	require.True(f, ok)

	f.Add(t0[:], fuzzAddr1.Bytes(), fuzzAddr2.Bytes(), []byte{0x2a})
	f.Add([]byte{}, []byte{}, []byte{}, []byte{})
	f.Add(make([]byte, 32), make([]byte, 20), make([]byte, 20), make([]byte, 32))
	// a real production incident: a hand-written Yul contract emitting a
	// Transfer-shaped log with a 1-byte data payload instead of a 32-byte word.
	f.Add(t0[:], fuzzAddr1.Bytes(), fuzzAddr2.Bytes(), []byte{0x01})

	f.Fuzz(func(t *testing.T, topic0Raw, topic1Raw, topic2Raw, data []byte) {
		h0 := types.NewHashFromBytes(topic0Raw)
		h1 := types.NewHashFromBytes(topic1Raw)
		h2 := types.NewHashFromBytes(topic2Raw)

		variants := [][]*types.Hash{
			nil,
			{h0},
			{h0, h1},
			{h0, h1, h2},
		}
		for _, topics := range variants {
			logs := []tx.Log{{Address: fuzzToken1, Topics: topics, Data: data}}
			_, _ = ExtractTransfers(logs)
			_, _ = ExtractApprovals(logs)
		}
	})
}

// go test -run '^$' -fuzz '^FuzzExtractTransfersRoundTrip$' -fuzztime=10s ./contracts/erc20
func FuzzExtractTransfersRoundTrip(f *testing.F) {
	f.Add(fuzzAddr1.Bytes(), fuzzAddr2.Bytes(), []byte{0x01, 0x02, 0x03})
	f.Add(make([]byte, 20), make([]byte, 20), []byte{})

	f.Fuzz(func(t *testing.T, fromRaw, toRaw, valueRaw []byte) {
		from := types.NewAddressFromBytes(fromRaw)
		to := types.NewAddressFromBytes(toRaw)
		value := clampUint256(valueRaw)

		data, err := abi.Pack(abi.Types{abi.Uint256}, value)
		require.NoError(t, err)

		t0, ok := transferEvent.Topic0()
		require.True(t, ok)

		logs := []tx.Log{
			{
				Address: fuzzToken1,
				Topics: []*types.Hash{
					types.NewHashFromBytes(t0[:]),
					addrTopic(from),
					addrTopic(to),
				},
				Data: data,
			},
		}

		transfers, err := ExtractTransfers(logs)
		require.NoError(t, err)
		require.Len(t, transfers, 1)
		require.True(t, transfers[0].Contract.Equal(fuzzToken1))
		require.True(t, transfers[0].From.Equal(from))
		require.True(t, transfers[0].To.Equal(to))
		require.Zero(t, value.Cmp(transfers[0].Value))

		approvals, err := ExtractApprovals(logs)
		require.NoError(t, err)
		require.Empty(t, approvals)
	})
}

func clampUint256(raw []byte) *big.Int {
	if len(raw) > 32 {
		raw = raw[len(raw)-32:]
	}
	return new(big.Int).SetBytes(raw)
}
