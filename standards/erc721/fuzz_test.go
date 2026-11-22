package erc721

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

var (
	fuzzToken1 = types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	fuzzAddr1  = types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	fuzzAddr2  = types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
)

// go test -run '^$' -fuzz '^FuzzSimpleDecodersDoNotPanic$' -fuzztime=10s ./standards/erc721
func FuzzSimpleDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 33))
	f.Add(make([]byte, 64))

	nameData, err := abi.Pack(abi.NewTypes(abi.String), "Bored Ape Yacht Club")
	require.NoError(f, err)
	f.Add(nameData)

	uintData, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(123456789))
	require.NoError(f, err)
	f.Add(uintData)

	boolData, err := abi.Pack(abi.NewTypes(abi.Bool), true)
	require.NoError(f, err)
	f.Add(boolData)

	addrData, err := abi.Pack(abi.NewTypes(abi.Address), fuzzAddr1)
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

// go test -run '^$' -fuzz '^FuzzDecodeBalanceOfRoundTrip$' -fuzztime=10s ./standards/erc721
func FuzzDecodeBalanceOfRoundTrip(f *testing.F) {
	f.Add([]byte{0x01})
	f.Add([]byte{})
	f.Add(make([]byte, 32))

	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > 32 {
			raw = raw[len(raw)-32:]
		}
		value := new(big.Int).SetBytes(raw)

		data, err := abi.Pack(abi.NewTypes(abi.Uint256), value)
		require.NoError(t, err)

		got, err := DecodeBalanceOf(data)
		require.NoError(t, err)
		require.Zero(t, value.Cmp(got))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeOwnerOfRoundTrip$' -fuzztime=10s ./standards/erc721
func FuzzDecodeOwnerOfRoundTrip(f *testing.F) {
	f.Add(fuzzAddr1.Bytes())
	f.Add(make([]byte, 20))

	f.Fuzz(func(t *testing.T, raw []byte) {
		addr := types.NewAddressFromBytes(raw)

		data, err := abi.Pack(abi.NewTypes(abi.Address), addr)
		require.NoError(t, err)

		got, err := DecodeOwnerOf(data)
		require.NoError(t, err)
		require.True(t, got.Equal(addr))
	})
}

// go test -run '^$' -fuzz '^FuzzBatchDecodersDoNotPanic$' -fuzztime=10s ./standards/erc721
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

// go test -run '^$' -fuzz '^FuzzExtractEventsCrashSafety$' -fuzztime=10s ./standards/erc721
func FuzzExtractEventsCrashSafety(f *testing.F) {
	t0, ok := transferEvent.Topic0()
	require.True(f, ok)

	f.Add(t0[:], fuzzAddr1.Bytes(), fuzzAddr2.Bytes(), []byte{0x04, 0xd2}, []byte{})
	f.Add([]byte{}, []byte{}, []byte{}, []byte{}, []byte{})
	f.Add(make([]byte, 32), make([]byte, 20), make([]byte, 20), make([]byte, 32), make([]byte, 32))
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

// go test -run '^$' -fuzz '^FuzzExtractTransfersRoundTrip$' -fuzztime=10s ./standards/erc721
func FuzzExtractTransfersRoundTrip(f *testing.F) {
	f.Add(fuzzAddr1.Bytes(), fuzzAddr2.Bytes(), []byte{0x04, 0xd2})
	f.Add(make([]byte, 20), make([]byte, 20), []byte{})

	f.Fuzz(func(t *testing.T, fromRaw, toRaw, tokenIdRaw []byte) {
		from := types.NewAddressFromBytes(fromRaw)
		to := types.NewAddressFromBytes(toRaw)
		if len(tokenIdRaw) > 32 {
			tokenIdRaw = tokenIdRaw[len(tokenIdRaw)-32:]
		}
		tokenId := new(big.Int).SetBytes(tokenIdRaw)

		t0, ok := transferEvent.Topic0()
		require.True(t, ok)

		logs := []tx.Log{
			{
				Address: fuzzToken1,
				Topics: []*types.Hash{
					types.NewHashFromBytes(t0[:]),
					types.NewHashFromBytes(from.Bytes()),
					types.NewHashFromBytes(to.Bytes()),
					types.NewHashFromBytes(tokenId.FillBytes(make([]byte, 32))),
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

// go test -run '^$' -fuzz '^FuzzExtractApprovalForAllRoundTrip$' -fuzztime=10s ./standards/erc721
func FuzzExtractApprovalForAllRoundTrip(f *testing.F) {
	f.Add(fuzzAddr1.Bytes(), fuzzAddr2.Bytes(), true)
	f.Add(make([]byte, 20), make([]byte, 20), false)

	f.Fuzz(func(t *testing.T, ownerRaw, operatorRaw []byte, approved bool) {
		owner := types.NewAddressFromBytes(ownerRaw)
		operator := types.NewAddressFromBytes(operatorRaw)

		t0, ok := approvalForAllEvent.Topic0()
		require.True(t, ok)

		data, err := abi.Pack(abi.NewTypes(abi.Bool), approved)
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

// go test -run '^$' -fuzz '^FuzzTransferFromRequestValidate$' -fuzztime=10s ./standards/erc721
func FuzzTransferFromRequestValidate(f *testing.F) {
	f.Add(uint8(0), []byte{})
	f.Add(uint8(0x0f), []byte{0x01})
	f.Add(uint8(0x1f), []byte{0xff})
	f.Add(uint8(0x07), []byte{})
	f.Add(uint8(0x10), []byte{0x00})

	f.Fuzz(func(t *testing.T, mask uint8, tokenIDRaw []byte) {
		if len(tokenIDRaw) > 32 {
			tokenIDRaw = tokenIDRaw[:32]
		}

		var req *TransferFromRequest
		if mask&0x10 == 0 {
			req = &TransferFromRequest{}
			if mask&0x01 != 0 {
				req.Contract = fuzzToken1
			}
			if mask&0x02 != 0 {
				req.From = fuzzAddr1
			}
			if mask&0x04 != 0 {
				req.To = fuzzAddr2
			}
			if mask&0x08 != 0 {
				req.TokenID = new(big.Int).SetBytes(tokenIDRaw)
			}
		}

		want := ""
		switch {
		case req == nil:
			want = "invalid request: request is nil"
		case req.Contract == nil:
			want = "invalid request: contract is nil"
		case req.From == nil:
			want = "invalid request: from is nil"
		case req.To == nil:
			want = "invalid request: to is nil"
		case req.TokenID == nil:
			want = "invalid request: tokenID is nil"
		}

		err := req.Validate()
		if want == "" {
			require.NoError(t, err)
			return
		}
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.EqualError(t, err, want)
	})
}

// go test -run '^$' -fuzz '^FuzzOwnerPairsRequestValidate$' -fuzztime=10s ./standards/erc721
func FuzzOwnerPairsRequestValidate(f *testing.F) {
	f.Add(true, uint8(0), uint8(0), uint8(9), uint8(9))
	f.Add(true, uint8(2), uint8(2), uint8(9), uint8(9))
	f.Add(true, uint8(3), uint8(2), uint8(9), uint8(9))
	f.Add(true, uint8(3), uint8(3), uint8(1), uint8(9))
	f.Add(true, uint8(3), uint8(3), uint8(9), uint8(0))
	f.Add(false, uint8(1), uint8(1), uint8(9), uint8(9))

	f.Fuzz(func(t *testing.T, withMulticall bool, contractLen, tokenIDLen, nilContract, nilTokenID uint8) {
		contractLen %= 8
		tokenIDLen %= 8

		contracts := make([]*types.Address, contractLen)
		for i := range contracts {
			contracts[i] = fuzzToken1
			if uint8(i) == nilContract {
				contracts[i] = nil
			}
		}
		tokenIDs := make([]*big.Int, tokenIDLen)
		for i := range tokenIDs {
			tokenIDs[i] = big.NewInt(int64(i))
			if uint8(i) == nilTokenID {
				tokenIDs[i] = nil
			}
		}

		var multicall *types.Address
		if withMulticall {
			multicall = fuzzAddr2
		}
		req := NewOwnerPairsRequest(multicall, contracts, tokenIDs, nilContract%2 == 0)

		want := ""
		switch {
		case multicall == nil:
			want = "invalid request: multicall is nil"
		case contractLen != tokenIDLen:
			want = "invalid request: contracts and tokenIDs differ in length"
		case contractLen == 0:
			want = "invalid request: contracts is empty"
		case nilContract < contractLen && nilContract <= nilTokenID:
			want = fmt.Sprintf("invalid request: contracts[%d] is nil", nilContract)
		case nilTokenID < tokenIDLen:
			want = fmt.Sprintf("invalid request: tokenIDs[%d] is nil", nilTokenID)
		}

		err := req.Validate()
		if want == "" {
			require.NoError(t, err)
			return
		}
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.EqualError(t, err, want)
	})
}

// go test -run '^$' -fuzz '^FuzzOwnerOfThroughClient$' -fuzztime=10s ./standards/erc721
func FuzzOwnerOfThroughClient(f *testing.F) {
	f.Add([]byte{}, []byte{}, []byte{})
	f.Add(fuzzToken1.Bytes(), []byte{0x2a}, fuzzAddr1.Bytes())
	f.Add(make([]byte, 20), make([]byte, 32), make([]byte, 20))

	f.Fuzz(func(t *testing.T, contractRaw, tokenIDRaw, ownerRaw []byte) {
		if len(tokenIDRaw) > 32 {
			tokenIDRaw = tokenIDRaw[len(tokenIDRaw)-32:]
		}
		contract := types.NewAddressFromBytes(contractRaw)
		tokenID := new(big.Int).SetBytes(tokenIDRaw)
		owner := types.NewAddressFromBytes(ownerRaw)

		calldata, err := EncodeOwnerOf(tokenID)
		require.NoError(t, err)
		result, err := abi.Pack(abi.NewTypes(abi.Address), owner)
		require.NoError(t, err)

		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			call := e.Params.([]any)[0].(map[string]any)
			require.Equal(t, contract.String(), call["to"])
			require.Equal(t, encoding.Hex.EncodePrefixed(calldata), call["data"])
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(result)
			return nil
		}).Once()

		got, err := New(c).OwnerOf(t.Context(), nil, NewOwnerOfRequest(contract, tokenID))
		require.NoError(t, err)
		require.True(t, got.Equal(owner))
	})
}

// go test -run '^$' -fuzz '^FuzzSafeTransferFromWithDataThroughClient$' -fuzztime=10s ./standards/erc721
func FuzzSafeTransferFromWithDataThroughClient(f *testing.F) {
	f.Add([]byte{}, []byte{}, []byte{}, []byte{}, uint64(0))
	f.Add(fuzzAddr1.Bytes(), fuzzAddr2.Bytes(), []byte{0x2a}, []byte{0xca, 0xfe}, uint64(7))
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 32), make([]byte, 300), ^uint64(0)-1)

	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(f, err)
	sender := core.PubkeyToAddress(key.PublicKey())

	f.Fuzz(func(t *testing.T, fromRaw, toRaw, tokenIDRaw, data []byte, nonce uint64) {
		if len(tokenIDRaw) > 32 {
			tokenIDRaw = tokenIDRaw[len(tokenIDRaw)-32:]
		}
		from := types.NewAddressFromBytes(fromRaw)
		to := types.NewAddressFromBytes(toRaw)
		tokenID := new(big.Int).SetBytes(tokenIDRaw)

		calldata, err := EncodeSafeTransferFromWithData(from, to, tokenID, data)
		require.NoError(t, err)

		builder := tx.NewDynamicFeeBuilder()
		builder.ChainID(big.NewInt(1))
		builder.Nonce(nonce)
		builder.Limit(200_000)
		builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

		var raw []byte
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			raw, err = encoding.Hex.Decode(e.Params.([]any)[0].(string))
			require.NoError(t, err)
			*e.Result.(*string) = core.Keccak256(raw).String()
			return nil
		}).Once()

		signed, hash, err := New(c).SafeTransferFromWithData(t.Context(), bind.NewTransactOptions(key, builder), NewSafeTransferFromWithDataRequest(fuzzToken1, from, to, tokenID, data))
		require.NoError(t, err)
		require.True(t, hash.Equal(core.Keccak256(raw)))
		require.Equal(t, nonce, signed.AddressNonce())

		decoded, err := core.DecodeTransaction(raw)
		require.NoError(t, err)
		dyn, ok := decoded.(*core.DynamicFeeTx)
		require.True(t, ok)
		require.True(t, dyn.To.Equal(fuzzToken1))
		require.Equal(t, calldata, dyn.Data)
		got, err := decoded.Sender()
		require.NoError(t, err)
		require.True(t, got.Equal(sender))
	})
}

// go test -run '^$' -fuzz '^FuzzReadsSurviveArbitraryResult$' -fuzztime=10s ./standards/erc721
func FuzzReadsSurviveArbitraryResult(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 96))
	f.Add([]byte{0xff, 0xff, 0xff, 0xff})

	f.Fuzz(func(t *testing.T, result []byte) {
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(result)
			return nil
		}).Times(7)
		e := New(c)

		addr, err := e.OwnerOf(t.Context(), nil, NewOwnerOfRequest(fuzzToken1, big.NewInt(1)))
		require.True(t, (err == nil) != (addr == nil))
		n, err := e.BalanceOf(t.Context(), nil, NewBalanceOfRequest(fuzzToken1, fuzzAddr1))
		require.True(t, (err == nil) != (n == nil))
		_, _ = e.TokenURI(t.Context(), nil, NewTokenURIRequest(fuzzToken1, big.NewInt(1)))
		_, _ = e.IsApprovedForAll(t.Context(), nil, NewIsApprovedForAllRequest(fuzzToken1, fuzzAddr1, fuzzAddr2))
		m, err := e.MetadataWithTokenID(t.Context(), nil, NewMetadataWithTokenIDRequest(fuzzAddr2, fuzzToken1, big.NewInt(1), true))
		require.True(t, (err == nil) != (m == nil))
		owners, err := e.TokenOwners(t.Context(), nil, NewTokenOwnersRequest(fuzzAddr2, fuzzToken1, []*big.Int{big.NewInt(1), big.NewInt(2)}, true))
		require.True(t, (err == nil) != (owners == nil))
		pairs, err := e.OwnerPairs(t.Context(), nil, NewOwnerPairsRequest(fuzzAddr2, []*types.Address{fuzzToken1}, []*big.Int{big.NewInt(1)}, false))
		if err == nil {
			require.Len(t, pairs, 1)
		}
	})
}
