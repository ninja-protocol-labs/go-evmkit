package v2

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// go test -run '^$' -fuzz '^FuzzEncodeGetPairRoundTrip$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzEncodeGetPairRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20))
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, make([]byte, 20))

	f.Fuzz(func(t *testing.T, tokenARaw, tokenBRaw []byte) {
		tokenA := types.NewAddressFromBytes(tokenARaw)
		tokenB := types.NewAddressFromBytes(tokenBRaw)

		data, err := EncodeGetPair(tokenA, tokenB)
		require.NoError(t, err)
		require.Equal(t, []byte{0xe6, 0xa4, 0x39, 0x05}, data[:4])

		vals, err := getPairFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(tokenA))
		require.True(t, vals[1].(*types.Address).Equal(tokenB))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeAllPairsRoundTrip$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzEncodeAllPairsRoundTrip(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 32))

	f.Fuzz(func(t *testing.T, indexRaw []byte) {
		if len(indexRaw) > 32 {
			indexRaw = indexRaw[len(indexRaw)-32:]
		}
		index := new(big.Int).SetBytes(indexRaw)

		data, err := EncodeAllPairs(index)
		require.NoError(t, err)
		require.Equal(t, []byte{0x1e, 0x3d, 0xd1, 0x8b}, data[:4])

		vals, err := allPairsFn.DecodeCall(data)
		require.NoError(t, err)
		require.Zero(t, index.Cmp(vals[0].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeCreatePairRoundTrip$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzEncodeCreatePairRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20))
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, make([]byte, 20))

	f.Fuzz(func(t *testing.T, tokenARaw, tokenBRaw []byte) {
		tokenA := types.NewAddressFromBytes(tokenARaw)
		tokenB := types.NewAddressFromBytes(tokenBRaw)

		data, err := EncodeCreatePair(tokenA, tokenB)
		require.NoError(t, err)
		require.Equal(t, []byte{0xc9, 0xc6, 0x53, 0x96}, data[:4])

		vals, err := createPairFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(tokenA))
		require.True(t, vals[1].(*types.Address).Equal(tokenB))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeSetFeeToRoundTrip$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzEncodeSetFeeToRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20))
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef})

	f.Fuzz(func(t *testing.T, feeToRaw []byte) {
		feeTo := types.NewAddressFromBytes(feeToRaw)

		data, err := EncodeSetFeeTo(feeTo)
		require.NoError(t, err)
		require.Equal(t, []byte{0xf4, 0x69, 0x01, 0xed}, data[:4])

		vals, err := setFeeToFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(feeTo))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeSetFeeToSetterRoundTrip$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzEncodeSetFeeToSetterRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20))
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef})

	f.Fuzz(func(t *testing.T, feeToSetterRaw []byte) {
		feeToSetter := types.NewAddressFromBytes(feeToSetterRaw)

		data, err := EncodeSetFeeToSetter(feeToSetter)
		require.NoError(t, err)
		require.Equal(t, []byte{0xa2, 0xe7, 0x4a, 0xf6}, data[:4])

		vals, err := setFeeToSetterFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(feeToSetter))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeAddressReturnsRoundTrip$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzDecodeAddressReturnsRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20))
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef})

	f.Fuzz(func(t *testing.T, addrRaw []byte) {
		addr := types.NewAddressFromBytes(addrRaw)

		data, err := abi.Pack(abi.Types{abi.Address}, addr)
		require.NoError(t, err)

		for _, decode := range []func([]byte) (*types.Address, error){DecodeFeeTo, DecodeFeeToSetter, DecodeGetPair, DecodeAllPairs, DecodeCreatePair} {
			got, err := decode(data)
			require.NoError(t, err)
			require.True(t, got.Equal(addr))
		}
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeAllPairsLengthRoundTrip$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzDecodeAllPairsLengthRoundTrip(f *testing.F) {
	f.Add([]byte{0x01})
	f.Add(make([]byte, 32))

	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > 32 {
			raw = raw[len(raw)-32:]
		}
		want := new(big.Int).SetBytes(raw)

		data, err := abi.Pack(abi.Types{abi.Uint256}, want)
		require.NoError(t, err)

		got, err := DecodeAllPairsLength(data)
		require.NoError(t, err)
		require.Zero(t, want.Cmp(got))
	})
}

// go test -run '^$' -fuzz '^FuzzSimpleDecodersDoNotPanic$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzSimpleDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 33))
	f.Add(make([]byte, 64))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeFeeTo(data)
		_, _ = DecodeFeeToSetter(data)
		_, _ = DecodeGetPair(data)
		_, _ = DecodeAllPairs(data)
		_, _ = DecodeAllPairsLength(data)
		_, _ = DecodeCreatePair(data)
	})
}
