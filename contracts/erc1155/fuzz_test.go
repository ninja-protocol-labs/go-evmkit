package erc1155

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

var (
	fuzzContract1 = types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	fuzzAddr1     = types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	fuzzAddr2     = types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
)

// go test -run '^$' -fuzz '^FuzzTransferBatchExpand$' -fuzztime=10s ./contracts/erc1155
func FuzzTransferBatchExpand(f *testing.F) {
	f.Add(uint8(0), uint8(0))
	f.Add(uint8(1), uint8(1))
	f.Add(uint8(3), uint8(3))
	f.Add(uint8(255), uint8(255))
	f.Add(uint8(2), uint8(1)) // mismatched lengths: must error, never panic
	f.Add(uint8(0), uint8(1))
	f.Add(uint8(1), uint8(0))

	f.Fuzz(func(t *testing.T, nIds, nValues uint8) {
		ids := make([]*big.Int, nIds)
		for i := range ids {
			ids[i] = big.NewInt(int64(i))
		}
		values := make([]*big.Int, nValues)
		for i := range values {
			values[i] = big.NewInt(int64(i) * 10)
		}

		tb := NewTransferBatch(fuzzContract1, fuzzAddr1, fuzzAddr1, fuzzAddr2, ids, values)

		transfers, err := tb.Expand()
		if nIds != nValues {
			require.Error(t, err)
			require.Nil(t, transfers)
			return
		}

		require.NoError(t, err)
		require.Len(t, transfers, int(nIds))
		for i, tr := range transfers {
			require.Same(t, fuzzContract1, tr.Contract)
			require.Same(t, fuzzAddr1, tr.Operator)
			require.Same(t, fuzzAddr1, tr.From)
			require.Same(t, fuzzAddr2, tr.To)
			require.Same(t, ids[i], tr.Id)
			require.Same(t, values[i], tr.Value)
		}
	})
}

// go test -run '^$' -fuzz '^FuzzSimpleDecodersDoNotPanic$' -fuzztime=10s ./contracts/erc1155
func FuzzSimpleDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 33))
	f.Add(make([]byte, 64))
	f.Add(make([]byte, 96))

	uintData, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(123456789))
	require.NoError(f, err)
	f.Add(uintData)

	boolData, err := abi.Pack(abi.Types{abi.Bool}, true)
	require.NoError(f, err)
	f.Add(boolData)

	addrData, err := abi.Pack(abi.Types{abi.Address}, fuzzAddr1)
	require.NoError(f, err)
	f.Add(addrData)

	strData, err := abi.Pack(abi.Types{abi.String}, "ipfs://Qm.../{id}.json")
	require.NoError(f, err)
	f.Add(strData)

	sliceData, err := abi.Pack(abi.Types{abi.Slice(abi.Uint256)}, []*big.Int{big.NewInt(1), big.NewInt(2)})
	require.NoError(f, err)
	f.Add(sliceData)

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeBalanceOf(data)
		_, _ = DecodeIsApprovedForAll(data)
		_, _ = DecodeURI(data)
		_, _ = DecodeBalanceOfBatch(data)
	})
}

// go test -run '^$' -fuzz '^FuzzBatchDecodersDoNotPanic$' -fuzztime=10s ./contracts/erc1155
//
// Feeds arbitrary garbage, including lengths that don't match the fixed
// ids/accounts/contracts the decoders are told about, through every
// Decode*WithBalance/Decode*Batch* decoder: these must always return an
// error on malformed or mismatched input, never panic.
func FuzzBatchDecodersDoNotPanic(f *testing.F) {
	ids := []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3)}
	accounts := []*types.Address{fuzzAddr1, fuzzAddr2, fuzzAddr1}
	contracts := []*types.Address{fuzzContract1, fuzzContract1}

	f.Add([]byte{})
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 192))

	if seed, err := EncodeAccountIdBalances(fuzzAddr1, big.NewInt(1), contracts, false); err == nil {
		f.Add(seed)
	}
	if seed, err := EncodeApprovalForAllWithBalance(fuzzContract1, fuzzAddr1, fuzzAddr2, big.NewInt(1), false); err == nil {
		f.Add(seed)
	}
	if seed, err := EncodeBalanceOfBatch(accounts, ids); err == nil {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeIdBalances(ids, data)
		_, _ = DecodeIdBalancePairs(accounts, ids, data)
		_, _ = DecodeAccountIdBalances(fuzzAddr1, big.NewInt(1), contracts, data)
		_, _ = DecodeApprovalForAllWithBalance(fuzzAddr1, fuzzAddr2, big.NewInt(1), data)
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeBalanceOfRoundTrip$' -fuzztime=10s ./contracts/erc1155
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

// go test -run '^$' -fuzz '^FuzzDecodeURIRoundTrip$' -fuzztime=10s ./contracts/erc1155
func FuzzDecodeURIRoundTrip(f *testing.F) {
	f.Add("")
	f.Add("ipfs://Qm.../{id}.json")
	f.Add("https://example.com/metadata/0.json")

	f.Fuzz(func(t *testing.T, s string) {
		data, err := abi.Pack(abi.Types{abi.String}, s)
		require.NoError(t, err)

		got, err := DecodeURI(data)
		require.NoError(t, err)
		require.Equal(t, s, got)
	})
}

// go test -run '^$' -fuzz '^FuzzBalanceOfBatchRoundTrip$' -fuzztime=10s ./contracts/erc1155
//
// Covers the exact shape that originally broke DecodeBalanceOfBatch (a
// Slice(Uint256) return decoding to []any, not []*big.Int) across random
// lengths and values, both through EncodeBalanceOfBatch/DecodeCall and
// through DecodeBalanceOfBatch's own return-data decoding.
func FuzzBalanceOfBatchRoundTrip(f *testing.F) {
	f.Add(uint8(0), []byte{})
	f.Add(uint8(1), []byte{0x01})
	f.Add(uint8(5), []byte{0xde, 0xad, 0xbe, 0xef})
	f.Add(uint8(32), []byte{0xff})

	f.Fuzz(func(t *testing.T, n uint8, seed []byte) {
		count := int(n % 24)
		accounts := make([]*types.Address, count)
		ids := make([]*big.Int, count)
		amounts := make([]*big.Int, count)
		for i := range accounts {
			h := append(append([]byte{}, seed...), byte(i), 0xa0)
			accounts[i] = types.NewAddressFromBytes(h)
			ids[i] = clampUint256(append(append([]byte{}, seed...), byte(i), 0xd0))
			amounts[i] = clampUint256(append(append([]byte{}, seed...), byte(i), 0xb0))
		}

		calldata, err := EncodeBalanceOfBatch(accounts, ids)
		require.NoError(t, err)

		vals, err := balanceOfBatchFn.DecodeCall(calldata)
		require.NoError(t, err)
		gotAccounts, ok := vals[0].([]any)
		require.True(t, ok)
		require.Len(t, gotAccounts, count)
		gotIds, ok := vals[1].([]any)
		require.True(t, ok)
		require.Len(t, gotIds, count)
		for i := range accounts {
			require.True(t, gotAccounts[i].(*types.Address).Equal(accounts[i]))
			require.Zero(t, ids[i].Cmp(gotIds[i].(*big.Int)))
		}

		retData, err := abi.Pack(abi.Types{abi.Slice(abi.Uint256)}, amounts)
		require.NoError(t, err)

		decoded, err := DecodeBalanceOfBatch(retData)
		require.NoError(t, err)
		require.Len(t, decoded, count)
		for i := range amounts {
			require.Zero(t, amounts[i].Cmp(decoded[i]))
		}

		pairs, err := DecodeIdBalancePairs(accounts, ids, retData)
		require.NoError(t, err)
		require.Len(t, pairs, count)
		for i := range pairs {
			require.Same(t, accounts[i], pairs[i].Account)
			require.Same(t, ids[i], pairs[i].Id)
			require.Zero(t, amounts[i].Cmp(pairs[i].Amount))
		}
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeBalanceOfBatchRejectsLengthMismatch$' -fuzztime=10s ./contracts/erc1155
func FuzzEncodeBalanceOfBatchRejectsLengthMismatch(f *testing.F) {
	f.Add(uint8(0), uint8(1))
	f.Add(uint8(3), uint8(2))
	f.Add(uint8(255), uint8(0))

	f.Fuzz(func(t *testing.T, nAccounts, nIds uint8) {
		accounts := make([]*types.Address, nAccounts)
		for i := range accounts {
			accounts[i] = fuzzAddr1
		}
		ids := make([]*big.Int, nIds)
		for i := range ids {
			ids[i] = big.NewInt(int64(i))
		}

		data, err := EncodeBalanceOfBatch(accounts, ids)
		if nAccounts != nIds {
			require.Error(t, err)
			require.Nil(t, data)
			return
		}
		require.NoError(t, err)
	})
}

// go test -run '^$' -fuzz '^FuzzSafeBatchTransferFromRoundTrip$' -fuzztime=10s ./contracts/erc1155
//
// Stresses encoding two dynamic arrays plus a dynamic bytes argument
// together in one call, across random lengths/values/callData.
func FuzzSafeBatchTransferFromRoundTrip(f *testing.F) {
	f.Add(uint8(0), []byte{}, []byte{})
	f.Add(uint8(1), []byte{0x01}, []byte{0xde, 0xad})
	f.Add(uint8(10), []byte{0xff, 0x00}, []byte{})

	f.Fuzz(func(t *testing.T, n uint8, seed, callData []byte) {
		count := int(n % 24)
		ids := make([]*big.Int, count)
		values := make([]*big.Int, count)
		for i := range ids {
			ids[i] = clampUint256(append(append([]byte{}, seed...), byte(i), 0x01))
			values[i] = clampUint256(append(append([]byte{}, seed...), byte(i), 0x02))
		}

		data, err := EncodeSafeBatchTransferFrom(fuzzAddr1, fuzzAddr2, ids, values, callData)
		require.NoError(t, err)

		vals, err := safeBatchTransferFromFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(fuzzAddr1))
		require.True(t, vals[1].(*types.Address).Equal(fuzzAddr2))

		gotIds, ok := vals[2].([]any)
		require.True(t, ok)
		require.Len(t, gotIds, count)
		gotValues, ok := vals[3].([]any)
		require.True(t, ok)
		require.Len(t, gotValues, count)
		for i := range ids {
			require.Zero(t, ids[i].Cmp(gotIds[i].(*big.Int)))
			require.Zero(t, values[i].Cmp(gotValues[i].(*big.Int)))
		}
		require.Equal(t, callData, vals[4].([]byte))
	})
}

func clampUint256(raw []byte) *big.Int {
	if len(raw) > 32 {
		raw = raw[len(raw)-32:]
	}
	return new(big.Int).SetBytes(raw)
}
