package v2

import (
	"context"
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

// go test -run '^$' -fuzz '^FuzzExtractEventsDoNotPanic$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzExtractEventsDoNotPanic(f *testing.F) {
	f.Add(uint8(0), []byte{}, []byte{})
	f.Add(uint8(4), make([]byte, 128), make([]byte, 64))
	f.Add(uint8(7), []byte{0xde, 0xad, 0xbe, 0xef}, []byte{0x01})

	f.Fuzz(func(t *testing.T, topicCount uint8, topicsRaw, data []byte) {
		n := int(topicCount % 8)
		topics := make([]*types.Hash, n)
		for i := range topics {
			start := i * 32
			if start >= len(topicsRaw) {
				topics[i] = types.NewHashFromBytes(nil)
				continue
			}
			end := min(start+32, len(topicsRaw))
			topics[i] = types.NewHashFromBytes(topicsRaw[start:end])
		}
		logs := []tx.Log{{Address: testPair, Topics: topics, Data: data}}

		_, _ = ExtractPairCreated(logs)
		_, _ = ExtractMint(logs)
		_, _ = ExtractBurn(logs)
		_, _ = ExtractSwap(logs)
		_, _ = ExtractSync(logs)
	})
}

// go test -run '^$' -fuzz '^FuzzPairSwapThroughClient$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzPairSwapThroughClient(f *testing.F) {
	f.Add([]byte{}, []byte{}, []byte{}, []byte{}, []byte{}, uint64(0), true)
	f.Add([]byte{0x01}, []byte{0x00}, []byte{0x5f}, []byte{0x02}, []byte{0xca, 0xfe}, uint64(42), false)
	f.Add(make([]byte, 20), make([]byte, 32), make([]byte, 32), make([]byte, 20), make([]byte, 100), ^uint64(0)-1, true)

	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(f, err)
	signer := core.PubkeyToAddress(key.PublicKey())

	f.Fuzz(func(t *testing.T, pairRaw, amount0Raw, amount1Raw, toRaw, swapData []byte, nonce uint64, dynamic bool) {
		if len(amount0Raw) > 32 {
			amount0Raw = amount0Raw[len(amount0Raw)-32:]
		}
		if len(amount1Raw) > 32 {
			amount1Raw = amount1Raw[len(amount1Raw)-32:]
		}
		pair := types.NewAddressFromBytes(pairRaw)
		amount0Out := new(big.Int).SetBytes(amount0Raw)
		amount1Out := new(big.Int).SetBytes(amount1Raw)
		to := types.NewAddressFromBytes(toRaw)

		var builder tx.Builder
		if dynamic {
			builder = tx.NewDynamicFeeBuilder()
		} else {
			builder = tx.NewLegacyBuilder()
		}
		builder.ChainID(big.NewInt(1))
		builder.Nonce(nonce)
		builder.Limit(3_000_000)
		builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

		var raw []byte
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			raw, err = encoding.Hex.Decode(e.Params.([]any)[0].(string))
			require.NoError(t, err)
			*e.Result.(*string) = core.Keccak256(raw).String()
			return nil
		}).Once()

		signed, hash, err := NewPair(c).Swap(t.Context(), bind.NewTransactOptions(key, builder), NewSwapRequest(pair, amount0Out, amount1Out, to, swapData))
		require.NoError(t, err)
		require.True(t, hash.Equal(core.Keccak256(raw)))
		require.Equal(t, nonce, signed.AddressNonce())

		decoded, err := core.DecodeTransaction(raw)
		require.NoError(t, err)
		from, err := decoded.Sender()
		require.NoError(t, err)
		require.True(t, from.Equal(signer))

		var gotTo *types.Address
		var gotData []byte
		switch v := decoded.(type) {
		case *core.DynamicFeeTx:
			require.True(t, dynamic)
			gotTo, gotData = v.To, v.Data
		case *core.LegacyTx:
			require.False(t, dynamic)
			gotTo, gotData = v.To, v.Data
		default:
			t.Fatalf("unexpected tx type %T", decoded)
		}
		require.True(t, gotTo.Equal(pair))

		vals, err := swapFn.DecodeCall(gotData)
		require.NoError(t, err)
		require.Zero(t, amount0Out.Cmp(vals[0].(*big.Int)))
		require.Zero(t, amount1Out.Cmp(vals[1].(*big.Int)))
		require.True(t, to.Equal(vals[2].(*types.Address)))
		require.Equal(t, len(swapData), len(vals[3].([]byte)))
		if len(swapData) > 0 {
			require.Equal(t, swapData, vals[3].([]byte))
		}
	})
}

// go test -run '^$' -fuzz '^FuzzPairReadsSurviveArbitraryResult$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzPairReadsSurviveArbitraryResult(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 64))
	f.Add(make([]byte, 96))

	pair := types.MustNewAddressFromHex("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	f.Fuzz(func(t *testing.T, result []byte) {
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(result)
			return nil
		}).Times(6)
		p := NewPair(c)

		reserves, err := p.GetReserves(t.Context(), nil, NewGetReservesRequest(pair))
		require.True(t, (err == nil) != (reserves == nil))
		token0, err := p.Token0(t.Context(), nil, NewToken0Request(pair))
		require.True(t, (err == nil) != (token0 == nil))
		kLast, err := p.KLast(t.Context(), nil, NewKLastRequest(pair))
		require.True(t, (err == nil) != (kLast == nil))
		typehash, err := p.PermitTypehash(t.Context(), nil, NewPermitTypehashRequest(pair))
		require.True(t, (err == nil) != (typehash == nil))
		liquidity, err := p.CallMint(t.Context(), nil, NewMintRequest(pair, a))
		require.True(t, (err == nil) != (liquidity == nil))
		amount0, amount1, err := p.CallBurn(t.Context(), nil, NewBurnRequest(pair, a))
		require.True(t, (err == nil) != (amount0 == nil))
		require.True(t, (err == nil) != (amount1 == nil))
	})
}
