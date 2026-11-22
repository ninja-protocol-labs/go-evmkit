package v3

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

// go test -run '^$' -fuzz '^FuzzEncodeGetPoolRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeGetPoolRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), []byte{0x01, 0xf4})
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, make([]byte, 20), []byte{0x27, 0x10})

	f.Fuzz(func(t *testing.T, tokenARaw, tokenBRaw, feeRaw []byte) {
		tokenA := types.NewAddressFromBytes(tokenARaw)
		tokenB := types.NewAddressFromBytes(tokenBRaw)
		// fold feeRaw into [0, 2^24), matching uint24's valid range
		fee := new(big.Int).Mod(new(big.Int).SetBytes(feeRaw), new(big.Int).Lsh(big.NewInt(1), 24))

		data, err := EncodeGetPool(tokenA, tokenB, fee)
		require.NoError(t, err)
		require.Equal(t, []byte{0x16, 0x98, 0xee, 0x82}, data[:4])

		vals, err := getPoolFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(tokenA))
		require.True(t, vals[1].(*types.Address).Equal(tokenB))
		require.Zero(t, fee.Cmp(vals[2].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeCreatePoolRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeCreatePoolRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), []byte{0x01, 0xf4})
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, make([]byte, 20), []byte{0x27, 0x10})

	f.Fuzz(func(t *testing.T, tokenARaw, tokenBRaw, feeRaw []byte) {
		tokenA := types.NewAddressFromBytes(tokenARaw)
		tokenB := types.NewAddressFromBytes(tokenBRaw)
		// fold feeRaw into [0, 2^24), matching uint24's valid range
		fee := new(big.Int).Mod(new(big.Int).SetBytes(feeRaw), new(big.Int).Lsh(big.NewInt(1), 24))

		data, err := EncodeCreatePool(tokenA, tokenB, fee)
		require.NoError(t, err)
		require.Equal(t, []byte{0xa1, 0x67, 0x12, 0x95}, data[:4])

		vals, err := createPoolFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(tokenA))
		require.True(t, vals[1].(*types.Address).Equal(tokenB))
		require.Zero(t, fee.Cmp(vals[2].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeSetOwnerRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeSetOwnerRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20))
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef})

	f.Fuzz(func(t *testing.T, ownerRaw []byte) {
		owner := types.NewAddressFromBytes(ownerRaw)

		data, err := EncodeSetOwner(owner)
		require.NoError(t, err)
		require.Equal(t, []byte{0x13, 0xaf, 0x40, 0x35}, data[:4])

		vals, err := setOwnerFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(owner))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeEnableFeeAmountRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeEnableFeeAmountRoundTrip(f *testing.F) {
	f.Add([]byte{0x01, 0xf4}, []byte{0x0a})
	f.Add([]byte{0x27, 0x10}, []byte{0xc8})

	f.Fuzz(func(t *testing.T, feeRaw, tickSpacingRaw []byte) {
		// fold feeRaw into [0, 2^24), matching uint24's valid range
		fee := new(big.Int).Mod(new(big.Int).SetBytes(feeRaw), new(big.Int).Lsh(big.NewInt(1), 24))
		// fold tickSpacingRaw into [-2^23, 2^23), matching int24's valid range
		tickSpacing := new(big.Int).Mod(new(big.Int).SetBytes(tickSpacingRaw), new(big.Int).Lsh(big.NewInt(1), 24))
		if tickSpacing.Cmp(new(big.Int).Lsh(big.NewInt(1), 23)) >= 0 {
			tickSpacing.Sub(tickSpacing, new(big.Int).Lsh(big.NewInt(1), 24))
		}

		data, err := EncodeEnableFeeAmount(fee, tickSpacing)
		require.NoError(t, err)
		require.Equal(t, []byte{0x8a, 0x7c, 0x19, 0x5f}, data[:4])

		vals, err := enableFeeAmountFn.DecodeCall(data)
		require.NoError(t, err)
		require.Zero(t, fee.Cmp(vals[0].(*big.Int)))
		require.Zero(t, tickSpacing.Cmp(vals[1].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeFeeAmountTickSpacingRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeFeeAmountTickSpacingRoundTrip(f *testing.F) {
	f.Add([]byte{0x01, 0xf4})
	f.Add([]byte{0x27, 0x10})

	f.Fuzz(func(t *testing.T, feeRaw []byte) {
		// fold feeRaw into [0, 2^24), matching uint24's valid range
		fee := new(big.Int).Mod(new(big.Int).SetBytes(feeRaw), new(big.Int).Lsh(big.NewInt(1), 24))

		data, err := EncodeFeeAmountTickSpacing(fee)
		require.NoError(t, err)
		require.Equal(t, []byte{0x22, 0xaf, 0xcc, 0xcb}, data[:4])

		vals, err := feeAmountTickSpacingFn.DecodeCall(data)
		require.NoError(t, err)
		require.Zero(t, fee.Cmp(vals[0].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeAddressReturnsRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzDecodeAddressReturnsRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20))
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef})

	f.Fuzz(func(t *testing.T, addrRaw []byte) {
		addr := types.NewAddressFromBytes(addrRaw)

		data, err := abi.Pack(abi.Types{abi.Address}, addr)
		require.NoError(t, err)

		for _, decode := range []func([]byte) (*types.Address, error){DecodeOwner, DecodeGetPool, DecodeCreatePool} {
			got, err := decode(data)
			require.NoError(t, err)
			require.True(t, got.Equal(addr))
		}
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeFeeAmountTickSpacingRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzDecodeFeeAmountTickSpacingRoundTrip(f *testing.F) {
	f.Add([]byte{0x00, 0x0a})
	f.Add([]byte{0xff, 0xf6}) // -10 two's complement

	f.Fuzz(func(t *testing.T, raw []byte) {
		// fold raw into [-2^23, 2^23), matching int24's valid range
		want := new(big.Int).Mod(new(big.Int).SetBytes(raw), new(big.Int).Lsh(big.NewInt(1), 24))
		if want.Cmp(new(big.Int).Lsh(big.NewInt(1), 23)) >= 0 {
			want.Sub(want, new(big.Int).Lsh(big.NewInt(1), 24))
		}

		data, err := abi.Pack(abi.Types{abi.Int24}, want)
		require.NoError(t, err)

		got, err := DecodeFeeAmountTickSpacing(data)
		require.NoError(t, err)
		require.Zero(t, want.Cmp(got))
	})
}

// go test -run '^$' -fuzz '^FuzzSimpleDecodersDoNotPanic$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzSimpleDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 33))
	f.Add(make([]byte, 64))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeOwner(data)
		_, _ = DecodeFeeAmountTickSpacing(data)
		_, _ = DecodeGetPool(data)
		_, _ = DecodeCreatePool(data)
	})
}

// go test -run '^$' -fuzz '^FuzzFactoryCreatePoolThroughClient$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzFactoryCreatePoolThroughClient(f *testing.F) {
	f.Add([]byte{}, []byte{}, []byte{}, uint32(0), uint64(0), true)
	f.Add([]byte{0x01}, []byte{0x02}, []byte{0x03}, uint32(3000), uint64(42), false)
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 20), uint32(1<<24-1), ^uint64(0)-1, true)

	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(f, err)
	signer := core.PubkeyToAddress(key.PublicKey())

	f.Fuzz(func(t *testing.T, factoryRaw, tokenARaw, tokenBRaw []byte, feeRaw uint32, nonce uint64, dynamic bool) {
		factory := types.NewAddressFromBytes(factoryRaw)
		tokenA := types.NewAddressFromBytes(tokenARaw)
		tokenB := types.NewAddressFromBytes(tokenBRaw)
		fee := big.NewInt(int64(feeRaw % (1 << 24)))

		var builder tx.Builder
		if dynamic {
			builder = tx.NewDynamicFeeBuilder()
		} else {
			builder = tx.NewLegacyBuilder()
		}
		builder.ChainID(big.NewInt(1))
		builder.Nonce(nonce)
		builder.Limit(6_000_000)
		builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

		var raw []byte
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			raw, err = encoding.Hex.Decode(e.Params.([]any)[0].(string))
			require.NoError(t, err)
			*e.Result.(*string) = core.Keccak256(raw).String()
			return nil
		}).Once()

		signed, hash, err := NewFactory(c).CreatePool(t.Context(), bind.NewTransactOptions(key, builder), NewCreatePoolRequest(factory, tokenA, tokenB, fee))
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
		require.True(t, gotTo.Equal(factory))

		vals, err := createPoolFn.DecodeCall(gotData)
		require.NoError(t, err)
		require.True(t, tokenA.Equal(vals[0].(*types.Address)))
		require.True(t, tokenB.Equal(vals[1].(*types.Address)))
		require.Zero(t, fee.Cmp(vals[2].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzFactoryReadsSurviveArbitraryResult$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzFactoryReadsSurviveArbitraryResult(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(append(make([]byte, 29), 0x80, 0x00, 0x00))

	factory := types.MustNewAddressFromHex("0x1F98431c8aD98523631AE4a59f267346ea31F984")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	f.Fuzz(func(t *testing.T, result []byte) {
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(result)
			return nil
		}).Times(4)
		fac := NewFactory(c)

		owner, err := fac.Owner(t.Context(), nil, NewOwnerRequest(factory))
		require.True(t, (err == nil) != (owner == nil))
		tickSpacing, err := fac.FeeAmountTickSpacing(t.Context(), nil, NewFeeAmountTickSpacingRequest(factory, big.NewInt(3000)))
		require.True(t, (err == nil) != (tickSpacing == nil))
		pool, err := fac.GetPool(t.Context(), nil, NewGetPoolRequest(factory, a, a, big.NewInt(3000)))
		require.True(t, (err == nil) != (pool == nil))
		created, err := fac.CallCreatePool(t.Context(), nil, NewCreatePoolRequest(factory, a, a, big.NewInt(3000)))
		require.True(t, (err == nil) != (created == nil))
	})
}
