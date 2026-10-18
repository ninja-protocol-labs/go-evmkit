package v2

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// go test -run '^$' -fuzz '^FuzzDecodeGetReservesRoundTrip$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzDecodeGetReservesRoundTrip(f *testing.F) {
	f.Add([]byte{0x01}, []byte{0x02}, uint32(0))
	f.Add(make([]byte, 14), make([]byte, 14), uint32(4294967295))

	f.Fuzz(func(t *testing.T, r0Raw, r1Raw []byte, blockTimestampLast uint32) {
		if len(r0Raw) > 14 {
			r0Raw = r0Raw[len(r0Raw)-14:]
		}
		if len(r1Raw) > 14 {
			r1Raw = r1Raw[len(r1Raw)-14:]
		}
		reserve0 := new(big.Int).SetBytes(r0Raw)
		reserve1 := new(big.Int).SetBytes(r1Raw)

		data, err := abi.Pack(abi.Types{abi.Uint112, abi.Uint112, abi.Uint32}, reserve0, reserve1, blockTimestampLast)
		require.NoError(t, err)

		got, err := DecodeGetReserves(data)
		require.NoError(t, err)
		require.Zero(t, reserve0.Cmp(got.Reserve0))
		require.Zero(t, reserve1.Cmp(got.Reserve1))
		require.Equal(t, blockTimestampLast, got.BlockTimestampLast)
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeBurnRoundTrip$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzDecodeBurnRoundTrip(f *testing.F) {
	f.Add([]byte{0x01}, []byte{0x02})
	f.Add(make([]byte, 32), make([]byte, 32))

	f.Fuzz(func(t *testing.T, a0Raw, a1Raw []byte) {
		if len(a0Raw) > 32 {
			a0Raw = a0Raw[len(a0Raw)-32:]
		}
		if len(a1Raw) > 32 {
			a1Raw = a1Raw[len(a1Raw)-32:]
		}
		amount0 := new(big.Int).SetBytes(a0Raw)
		amount1 := new(big.Int).SetBytes(a1Raw)

		data, err := abi.Pack(abi.Types{abi.Uint256, abi.Uint256}, amount0, amount1)
		require.NoError(t, err)

		gotAmount0, gotAmount1, err := DecodeBurn(data)
		require.NoError(t, err)
		require.Zero(t, amount0.Cmp(gotAmount0))
		require.Zero(t, amount1.Cmp(gotAmount1))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeMintRoundTrip$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzEncodeMintRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20))
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef})

	f.Fuzz(func(t *testing.T, toRaw []byte) {
		to := types.NewAddressFromBytes(toRaw)

		data, err := EncodeMint(to)
		require.NoError(t, err)
		require.Equal(t, []byte{0x6a, 0x62, 0x78, 0x42}, data[:4])

		vals, err := mintFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(to))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeBurnRoundTrip$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzEncodeBurnRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20))
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef})

	f.Fuzz(func(t *testing.T, toRaw []byte) {
		to := types.NewAddressFromBytes(toRaw)

		data, err := EncodeBurn(to)
		require.NoError(t, err)
		require.Equal(t, []byte{0x89, 0xaf, 0xcb, 0x44}, data[:4])

		vals, err := burnFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(to))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeSwapRoundTrip$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzEncodeSwapRoundTrip(f *testing.F) {
	f.Add([]byte{}, []byte{}, make([]byte, 20), []byte{})
	f.Add(make([]byte, 32), make([]byte, 32), []byte{0xde, 0xad, 0xbe, 0xef}, []byte{1, 2, 3})

	f.Fuzz(func(t *testing.T, a0OutRaw, a1OutRaw, toRaw, swapData []byte) {
		if len(a0OutRaw) > 32 {
			a0OutRaw = a0OutRaw[len(a0OutRaw)-32:]
		}
		if len(a1OutRaw) > 32 {
			a1OutRaw = a1OutRaw[len(a1OutRaw)-32:]
		}
		amount0Out := new(big.Int).SetBytes(a0OutRaw)
		amount1Out := new(big.Int).SetBytes(a1OutRaw)
		to := types.NewAddressFromBytes(toRaw)

		data, err := EncodeSwap(amount0Out, amount1Out, to, swapData)
		require.NoError(t, err)
		require.Equal(t, []byte{0x02, 0x2c, 0x0d, 0x9f}, data[:4])

		vals, err := swapFn.DecodeCall(data)
		require.NoError(t, err)
		require.Zero(t, amount0Out.Cmp(vals[0].(*big.Int)))
		require.Zero(t, amount1Out.Cmp(vals[1].(*big.Int)))
		require.True(t, vals[2].(*types.Address).Equal(to))
		require.Equal(t, swapData, vals[3].([]byte))
	})
}

// go test -run '^$' -fuzz '^FuzzPairSimpleDecodersDoNotPanic$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzPairSimpleDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 33))
	f.Add(make([]byte, 64))
	f.Add(make([]byte, 96))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeMinimumLiquidity(data)
		_, _ = DecodeFactory(data)
		_, _ = DecodeToken0(data)
		_, _ = DecodeToken1(data)
		_, _ = DecodeGetReserves(data)
		_, _ = DecodePrice0CumulativeLast(data)
		_, _ = DecodePrice1CumulativeLast(data)
		_, _ = DecodeKLast(data)
		_, _ = DecodeMint(data)
		_, _, _ = DecodeBurn(data)
		_, _ = DecodePermitTypehash(data)
	})
}
