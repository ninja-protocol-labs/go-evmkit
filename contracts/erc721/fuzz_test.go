package erc721

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

var (
	fuzzToken1 = types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	fuzzAddr1  = types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	fuzzAddr2  = types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
)

// go test -run '^$' -fuzz '^FuzzSimpleDecodersDoNotPanic$' -fuzztime=10s ./contracts/erc721
func FuzzSimpleDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 33))
	f.Add(make([]byte, 64))

	nameData, err := abi.Pack(abi.Types{abi.String}, "Bored Ape Yacht Club")
	require.NoError(f, err)
	f.Add(nameData)

	uintData, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(123456789))
	require.NoError(f, err)
	f.Add(uintData)

	boolData, err := abi.Pack(abi.Types{abi.Bool}, true)
	require.NoError(f, err)
	f.Add(boolData)

	addrData, err := abi.Pack(abi.Types{abi.Address}, fuzzAddr1)
	require.NoError(f, err)
	f.Add(addrData)

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeName(data)
		_, _ = DecodeSymbol(data)
		_, _ = DecodeBalanceOf(data)
		_, _ = DecodeOwnerOf(data)
		_, _ = DecodeGetApproved(data)
		_, _ = DecodeIsApprovedForAll(data)
		_, _ = DecodeTokenURI(data)
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeBalanceOfRoundTrip$' -fuzztime=10s ./contracts/erc721
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

// go test -run '^$' -fuzz '^FuzzDecodeOwnerOfRoundTrip$' -fuzztime=10s ./contracts/erc721
func FuzzDecodeOwnerOfRoundTrip(f *testing.F) {
	f.Add(fuzzAddr1.Bytes())
	f.Add(make([]byte, 20))

	f.Fuzz(func(t *testing.T, raw []byte) {
		addr := types.NewAddressFromBytes(raw)

		data, err := abi.Pack(abi.Types{abi.Address}, addr)
		require.NoError(t, err)

		got, err := DecodeOwnerOf(data)
		require.NoError(t, err)
		require.True(t, got.Equal(addr))
	})
}

// go test -run '^$' -fuzz '^FuzzBatchDecodersDoNotPanic$' -fuzztime=10s ./contracts/erc721
func FuzzBatchDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 192))

	if seed, err := EncodeMetadata(fuzzToken1, false); err == nil {
		f.Add(seed)
	}
	if seed, err := EncodeMetadataWithTokenID(fuzzToken1, big.NewInt(1234), false); err == nil {
		f.Add(seed)
	}
	if seed, err := EncodeAllowanceWithTokenID(fuzzToken1, big.NewInt(1234), false); err == nil {
		f.Add(seed)
	}
	if seed, err := EncodeApprovalForAllWithBalance(fuzzToken1, fuzzAddr1, fuzzAddr2, false); err == nil {
		f.Add(seed)
	}
	tokenIds := []*big.Int{big.NewInt(1), big.NewInt(2)}
	if seed, err := EncodeTokenOwners(fuzzToken1, tokenIds, false); err == nil {
		f.Add(seed)
	}
	if seed, err := EncodeTokenOperators(fuzzToken1, tokenIds, false); err == nil {
		f.Add(seed)
	}
	contracts := []*types.Address{fuzzToken1, fuzzToken1}
	if seed, err := EncodeAddressBalances(fuzzAddr1, contracts, false); err == nil {
		f.Add(seed)
	}
	addrs := []*types.Address{fuzzAddr1, fuzzAddr2}
	if seed, err := EncodeTokenBalances(fuzzToken1, addrs, false); err == nil {
		f.Add(seed)
	}
	if seed, err := EncodeOwnerPairs(contracts, tokenIds, false); err == nil {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeMetadata(data)
		_, _ = DecodeMetadataWithTokenID(big.NewInt(1234), data)
		_, _ = DecodeAllowanceWithTokenID(big.NewInt(1234), data)
		_, _ = DecodeApprovalForAllWithBalance(fuzzAddr1, fuzzAddr2, data)
		_, _ = DecodeTokenOwners(fuzzToken1, tokenIds, data)
		_, _ = DecodeTokenOperators(fuzzToken1, tokenIds, data)
		_, _ = DecodeAddressBalances(fuzzAddr1, contracts, data)
		_, _ = DecodeTokenBalances(fuzzToken1, addrs, data)
		_, _ = DecodeOwnerPairs(contracts, tokenIds, data)
	})
}

// go test -run '^$' -fuzz '^FuzzExtractEventsCrashSafety$' -fuzztime=10s ./contracts/erc721
func FuzzExtractEventsCrashSafety(f *testing.F) {
	t0, ok := transferEvent.Topic0()
	require.True(f, ok)

	f.Add(t0[:], fuzzAddr1.Bytes(), fuzzAddr2.Bytes(), []byte{0x04, 0xd2}, []byte{})
	f.Add([]byte{}, []byte{}, []byte{}, []byte{}, []byte{})
	f.Add(make([]byte, 32), make([]byte, 20), make([]byte, 20), make([]byte, 32), make([]byte, 32))
	// a real production incident: a hand-written Yul contract emitting an
	// ApprovalForAll-shaped log with a 1-byte data payload instead of a 32-byte word.
	f.Add(t0[:], fuzzAddr1.Bytes(), fuzzAddr2.Bytes(), []byte{}, []byte{0x01})

	f.Fuzz(func(t *testing.T, topic0Raw, topic1Raw, topic2Raw, topic3Raw, data []byte) {
		h0 := types.NewHashFromBytes(topic0Raw)
		h1 := types.NewHashFromBytes(topic1Raw)
		h2 := types.NewHashFromBytes(topic2Raw)
		h3 := types.NewHashFromBytes(topic3Raw)

		variants := [][]*types.Hash{
			nil,
			{h0},
			{h0, h1},
			{h0, h1, h2},
			{h0, h1, h2, h3},
		}
		for _, topics := range variants {
			logs := []tx.Log{{Address: fuzzToken1, Topics: topics, Data: data}}
			_, _ = ExtractTransfers(logs)
			_, _ = ExtractApprovals(logs)
			_, _ = ExtractApprovalForAll(logs)
		}
	})
}

// go test -run '^$' -fuzz '^FuzzExtractTransfersRoundTrip$' -fuzztime=10s ./contracts/erc721
func FuzzExtractTransfersRoundTrip(f *testing.F) {
	f.Add(fuzzAddr1.Bytes(), fuzzAddr2.Bytes(), []byte{0x04, 0xd2})
	f.Add(make([]byte, 20), make([]byte, 20), []byte{})

	f.Fuzz(func(t *testing.T, fromRaw, toRaw, tokenIdRaw []byte) {
		from := types.NewAddressFromBytes(fromRaw)
		to := types.NewAddressFromBytes(toRaw)
		tokenId := clampUint256(tokenIdRaw)

		t0, ok := transferEvent.Topic0()
		require.True(t, ok)

		logs := []tx.Log{
			{
				Address: fuzzToken1,
				Topics: []*types.Hash{
					types.NewHashFromBytes(t0[:]),
					types.NewHashFromBytes(from.Bytes()),
					types.NewHashFromBytes(to.Bytes()),
					uint256Topic(tokenId),
				},
			},
		}

		transfers, err := ExtractTransfers(logs)
		require.NoError(t, err)
		require.Len(t, transfers, 1)
		require.True(t, transfers[0].Contract.Equal(fuzzToken1))
		require.True(t, transfers[0].From.Equal(from))
		require.True(t, transfers[0].To.Equal(to))
		require.Zero(t, tokenId.Cmp(transfers[0].TokenId))

		approvals, err := ExtractApprovals(logs)
		require.NoError(t, err)
		require.Empty(t, approvals)
	})
}

// go test -run '^$' -fuzz '^FuzzExtractApprovalForAllRoundTrip$' -fuzztime=10s ./contracts/erc721
func FuzzExtractApprovalForAllRoundTrip(f *testing.F) {
	f.Add(fuzzAddr1.Bytes(), fuzzAddr2.Bytes(), true)
	f.Add(make([]byte, 20), make([]byte, 20), false)

	f.Fuzz(func(t *testing.T, ownerRaw, operatorRaw []byte, approved bool) {
		owner := types.NewAddressFromBytes(ownerRaw)
		operator := types.NewAddressFromBytes(operatorRaw)

		t0, ok := approvalForAllEvent.Topic0()
		require.True(t, ok)

		data, err := abi.Pack(abi.Types{abi.Bool}, approved)
		require.NoError(t, err)

		logs := []tx.Log{
			{
				Address: fuzzToken1,
				Topics: []*types.Hash{
					types.NewHashFromBytes(t0[:]),
					types.NewHashFromBytes(owner.Bytes()),
					types.NewHashFromBytes(operator.Bytes()),
				},
				Data: data,
			},
		}

		approvals, err := ExtractApprovalForAll(logs)
		require.NoError(t, err)
		require.Len(t, approvals, 1)
		require.True(t, approvals[0].Contract.Equal(fuzzToken1))
		require.True(t, approvals[0].Owner.Equal(owner))
		require.True(t, approvals[0].Operator.Equal(operator))
		require.Equal(t, approved, approvals[0].Approved)
	})
}

func clampUint256(raw []byte) *big.Int {
	if len(raw) > 32 {
		raw = raw[len(raw)-32:]
	}
	return new(big.Int).SetBytes(raw)
}
