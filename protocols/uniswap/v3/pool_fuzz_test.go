package v3

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

// go test -run '^$' -fuzz '^FuzzEncodeTicksRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeTicksRoundTrip(f *testing.F) {
	f.Add([]byte{0x00, 0x64})
	f.Add([]byte{0xff, 0x9c})

	f.Fuzz(func(t *testing.T, raw []byte) {
		tick := new(big.Int).Mod(new(big.Int).SetBytes(raw), new(big.Int).Lsh(big.NewInt(1), 24))
		if tick.Cmp(new(big.Int).Lsh(big.NewInt(1), 23)) >= 0 {
			tick.Sub(tick, new(big.Int).Lsh(big.NewInt(1), 24))
		}

		data, err := EncodeTicks(tick)
		require.NoError(t, err)

		vals, err := ticksFn.DecodeCall(data)
		require.NoError(t, err)
		require.Zero(t, tick.Cmp(vals[0].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeTickBitmapRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeTickBitmapRoundTrip(f *testing.F) {
	f.Add(int16(0))
	f.Add(int16(-32768))

	f.Fuzz(func(t *testing.T, wordPosition int16) {
		data, err := EncodeTickBitmap(wordPosition)
		require.NoError(t, err)

		vals, err := tickBitmapFn.DecodeCall(data)
		require.NoError(t, err)
		require.Equal(t, wordPosition, vals[0].(int16))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodePositionsRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodePositionsRoundTrip(f *testing.F) {
	f.Add(make([]byte, 32))
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef})

	f.Fuzz(func(t *testing.T, raw []byte) {
		var key [32]byte
		copy(key[:], raw)

		data, err := EncodePositions(key)
		require.NoError(t, err)

		vals, err := positionsFn.DecodeCall(data)
		require.NoError(t, err)
		require.Equal(t, key[:], vals[0].([]byte))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodePoolObservationsRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodePoolObservationsRoundTrip(f *testing.F) {
	f.Add([]byte{0x01})
	f.Add(make([]byte, 32))

	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > 32 {
			raw = raw[len(raw)-32:]
		}
		index := new(big.Int).SetBytes(raw)

		data, err := EncodePoolObservations(index)
		require.NoError(t, err)

		vals, err := poolObservationsFn.DecodeCall(data)
		require.NoError(t, err)
		require.Zero(t, index.Cmp(vals[0].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeObserveRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeObserveRoundTrip(f *testing.F) {
	f.Add(uint32(0), uint32(3600))
	f.Add(uint32(4294967295), uint32(1))

	f.Fuzz(func(t *testing.T, a, b uint32) {
		secondsAgos := []uint32{a, b}

		data, err := EncodeObserve(secondsAgos)
		require.NoError(t, err)

		vals, err := observeFn.DecodeCall(data)
		require.NoError(t, err)
		got := vals[0].([]any)
		require.Len(t, got, 2)
		require.Equal(t, a, got[0].(uint32))
		require.Equal(t, b, got[1].(uint32))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeSnapshotCumulativesInsideRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeSnapshotCumulativesInsideRoundTrip(f *testing.F) {
	f.Add([]byte{0x00, 0x64}, []byte{0xff, 0x9c})
	f.Add([]byte{0x00}, []byte{0x00})

	f.Fuzz(func(t *testing.T, rawLower, rawUpper []byte) {
		tickLower := new(big.Int).Mod(new(big.Int).SetBytes(rawLower), new(big.Int).Lsh(big.NewInt(1), 24))
		if tickLower.Cmp(new(big.Int).Lsh(big.NewInt(1), 23)) >= 0 {
			tickLower.Sub(tickLower, new(big.Int).Lsh(big.NewInt(1), 24))
		}
		tickUpper := new(big.Int).Mod(new(big.Int).SetBytes(rawUpper), new(big.Int).Lsh(big.NewInt(1), 24))
		if tickUpper.Cmp(new(big.Int).Lsh(big.NewInt(1), 23)) >= 0 {
			tickUpper.Sub(tickUpper, new(big.Int).Lsh(big.NewInt(1), 24))
		}

		data, err := EncodeSnapshotCumulativesInside(tickLower, tickUpper)
		require.NoError(t, err)

		vals, err := snapshotCumulativesInsideFn.DecodeCall(data)
		require.NoError(t, err)
		require.Zero(t, tickLower.Cmp(vals[0].(*big.Int)))
		require.Zero(t, tickUpper.Cmp(vals[1].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeInitializeRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeInitializeRoundTrip(f *testing.F) {
	f.Add([]byte{0x01})
	f.Add(make([]byte, 20))

	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > 20 {
			raw = raw[len(raw)-20:]
		}
		sqrtPriceX96 := new(big.Int).SetBytes(raw)

		data, err := EncodeInitialize(sqrtPriceX96)
		require.NoError(t, err)

		vals, err := initializeFn.DecodeCall(data)
		require.NoError(t, err)
		require.Zero(t, sqrtPriceX96.Cmp(vals[0].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeMintRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeMintRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), []byte{0x00, 0x64}, []byte{0xff, 0x9c}, []byte{0x01, 0xf4}, []byte{0xde, 0xad})
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, []byte{0x00}, []byte{0x00}, []byte{0x00}, []byte{})

	f.Fuzz(func(t *testing.T, recipientRaw, tickLowerRaw, tickUpperRaw, amountRaw, callbackData []byte) {
		recipient := types.NewAddressFromBytes(recipientRaw)
		tickLower := new(big.Int).Mod(new(big.Int).SetBytes(tickLowerRaw), new(big.Int).Lsh(big.NewInt(1), 24))
		if tickLower.Cmp(new(big.Int).Lsh(big.NewInt(1), 23)) >= 0 {
			tickLower.Sub(tickLower, new(big.Int).Lsh(big.NewInt(1), 24))
		}
		tickUpper := new(big.Int).Mod(new(big.Int).SetBytes(tickUpperRaw), new(big.Int).Lsh(big.NewInt(1), 24))
		if tickUpper.Cmp(new(big.Int).Lsh(big.NewInt(1), 23)) >= 0 {
			tickUpper.Sub(tickUpper, new(big.Int).Lsh(big.NewInt(1), 24))
		}
		if len(amountRaw) > 16 {
			amountRaw = amountRaw[len(amountRaw)-16:]
		}
		amount := new(big.Int).SetBytes(amountRaw)

		data, err := EncodeMint(recipient, tickLower, tickUpper, amount, callbackData)
		require.NoError(t, err)

		vals, err := mintFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(recipient))
		require.Zero(t, tickLower.Cmp(vals[1].(*big.Int)))
		require.Zero(t, tickUpper.Cmp(vals[2].(*big.Int)))
		require.Zero(t, amount.Cmp(vals[3].(*big.Int)))
		require.Equal(t, callbackData, vals[4].([]byte))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeCollectRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeCollectRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), []byte{0x01}, []byte{0x02})
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, recipientRaw, amount0Raw, amount1Raw []byte) {
		recipient := types.NewAddressFromBytes(recipientRaw)
		if len(amount0Raw) > 16 {
			amount0Raw = amount0Raw[len(amount0Raw)-16:]
		}
		amount0Requested := new(big.Int).SetBytes(amount0Raw)
		if len(amount1Raw) > 16 {
			amount1Raw = amount1Raw[len(amount1Raw)-16:]
		}
		amount1Requested := new(big.Int).SetBytes(amount1Raw)

		data, err := EncodeCollect(recipient, big.NewInt(-10), big.NewInt(10), amount0Requested, amount1Requested)
		require.NoError(t, err)

		vals, err := collectFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(recipient))
		require.Zero(t, amount0Requested.Cmp(vals[3].(*big.Int)))
		require.Zero(t, amount1Requested.Cmp(vals[4].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeBurnRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeBurnRoundTrip(f *testing.F) {
	f.Add([]byte{0x01})
	f.Add(make([]byte, 16))

	f.Fuzz(func(t *testing.T, amountRaw []byte) {
		if len(amountRaw) > 16 {
			amountRaw = amountRaw[len(amountRaw)-16:]
		}
		amount := new(big.Int).SetBytes(amountRaw)

		data, err := EncodeBurn(big.NewInt(-10), big.NewInt(10), amount)
		require.NoError(t, err)

		vals, err := burnFn.DecodeCall(data)
		require.NoError(t, err)
		require.Zero(t, amount.Cmp(vals[2].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeSwapRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeSwapRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), true, []byte{0x01}, []byte{0x02})
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, false, []byte{0xff}, []byte{0x00})

	f.Fuzz(func(t *testing.T, recipientRaw []byte, zeroForOne bool, amountSpecifiedRaw, sqrtPriceLimitRaw []byte) {
		recipient := types.NewAddressFromBytes(recipientRaw)
		if len(amountSpecifiedRaw) > 32 {
			amountSpecifiedRaw = amountSpecifiedRaw[len(amountSpecifiedRaw)-32:]
		}
		amountSpecified := new(big.Int).Mod(new(big.Int).SetBytes(amountSpecifiedRaw), new(big.Int).Lsh(big.NewInt(1), 256))
		if amountSpecified.Cmp(new(big.Int).Lsh(big.NewInt(1), 255)) >= 0 {
			amountSpecified.Sub(amountSpecified, new(big.Int).Lsh(big.NewInt(1), 256))
		}
		if amountSpecified.Sign() == 0 {
			amountSpecified = big.NewInt(1)
		}
		if len(sqrtPriceLimitRaw) > 20 {
			sqrtPriceLimitRaw = sqrtPriceLimitRaw[len(sqrtPriceLimitRaw)-20:]
		}
		sqrtPriceLimitX96 := new(big.Int).SetBytes(sqrtPriceLimitRaw)

		data, err := EncodeSwap(recipient, zeroForOne, amountSpecified, sqrtPriceLimitX96, []byte{0x01, 0x02})
		require.NoError(t, err)

		vals, err := swapFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(recipient))
		require.Equal(t, zeroForOne, vals[1].(bool))
		require.Zero(t, amountSpecified.Cmp(vals[2].(*big.Int)))
		require.Zero(t, sqrtPriceLimitX96.Cmp(vals[3].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeFlashRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeFlashRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), []byte{0x01}, []byte{0x02})
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, recipientRaw, amount0Raw, amount1Raw []byte) {
		recipient := types.NewAddressFromBytes(recipientRaw)
		if len(amount0Raw) > 32 {
			amount0Raw = amount0Raw[len(amount0Raw)-32:]
		}
		amount0 := new(big.Int).SetBytes(amount0Raw)
		if len(amount1Raw) > 32 {
			amount1Raw = amount1Raw[len(amount1Raw)-32:]
		}
		amount1 := new(big.Int).SetBytes(amount1Raw)

		data, err := EncodeFlash(recipient, amount0, amount1, []byte{0x09})
		require.NoError(t, err)

		vals, err := flashFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(recipient))
		require.Zero(t, amount0.Cmp(vals[1].(*big.Int)))
		require.Zero(t, amount1.Cmp(vals[2].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeIncreaseObservationCardinalityNextRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeIncreaseObservationCardinalityNextRoundTrip(f *testing.F) {
	f.Add(uint16(0))
	f.Add(uint16(65535))

	f.Fuzz(func(t *testing.T, observationCardinalityNext uint16) {
		data, err := EncodeIncreaseObservationCardinalityNext(observationCardinalityNext)
		require.NoError(t, err)

		vals, err := increaseObservationCardinalityNextFn.DecodeCall(data)
		require.NoError(t, err)
		require.Equal(t, observationCardinalityNext, vals[0].(uint16))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeSetFeeProtocolRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeSetFeeProtocolRoundTrip(f *testing.F) {
	f.Add(uint8(0), uint8(0))
	f.Add(uint8(255), uint8(255))

	f.Fuzz(func(t *testing.T, feeProtocol0, feeProtocol1 uint8) {
		data, err := EncodeSetFeeProtocol(feeProtocol0, feeProtocol1)
		require.NoError(t, err)

		vals, err := setFeeProtocolFn.DecodeCall(data)
		require.NoError(t, err)
		require.Equal(t, feeProtocol0, vals[0].(uint8))
		require.Equal(t, feeProtocol1, vals[1].(uint8))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeCollectProtocolRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeCollectProtocolRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), []byte{0x01}, []byte{0x02})
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, recipientRaw, amount0Raw, amount1Raw []byte) {
		recipient := types.NewAddressFromBytes(recipientRaw)
		if len(amount0Raw) > 16 {
			amount0Raw = amount0Raw[len(amount0Raw)-16:]
		}
		amount0Requested := new(big.Int).SetBytes(amount0Raw)
		if len(amount1Raw) > 16 {
			amount1Raw = amount1Raw[len(amount1Raw)-16:]
		}
		amount1Requested := new(big.Int).SetBytes(amount1Raw)

		data, err := EncodeCollectProtocol(recipient, amount0Requested, amount1Requested)
		require.NoError(t, err)

		vals, err := collectProtocolFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(recipient))
		require.Zero(t, amount0Requested.Cmp(vals[1].(*big.Int)))
		require.Zero(t, amount1Requested.Cmp(vals[2].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodePoolAddressReturnsRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzDecodePoolAddressReturnsRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20))
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef})

	f.Fuzz(func(t *testing.T, addrRaw []byte) {
		addr := types.NewAddressFromBytes(addrRaw)

		data, err := abi.Pack(abi.Types{abi.Address}, addr)
		require.NoError(t, err)

		for _, decode := range []func([]byte) (*types.Address, error){DecodeFactory, DecodeToken0, DecodeToken1} {
			got, err := decode(data)
			require.NoError(t, err)
			require.True(t, got.Equal(addr))
		}
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeSlot0RoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzDecodeSlot0RoundTrip(f *testing.F) {
	f.Add([]byte{0x01}, []byte{0x00, 0x64}, uint16(1), uint16(2), uint16(3), uint8(4), true)
	f.Add(make([]byte, 20), []byte{0xff, 0x9c}, uint16(0), uint16(0), uint16(0), uint8(0), false)

	f.Fuzz(func(t *testing.T, sqrtPriceRaw, tickRaw []byte, observationIndex, observationCardinality, observationCardinalityNext uint16, feeProtocol uint8, unlocked bool) {
		if len(sqrtPriceRaw) > 20 {
			sqrtPriceRaw = sqrtPriceRaw[len(sqrtPriceRaw)-20:]
		}
		sqrtPriceX96 := new(big.Int).SetBytes(sqrtPriceRaw)
		tick := new(big.Int).Mod(new(big.Int).SetBytes(tickRaw), new(big.Int).Lsh(big.NewInt(1), 24))
		if tick.Cmp(new(big.Int).Lsh(big.NewInt(1), 23)) >= 0 {
			tick.Sub(tick, new(big.Int).Lsh(big.NewInt(1), 24))
		}

		data, err := abi.Pack(abi.Types{abi.Uint160, abi.Int24, abi.Uint16, abi.Uint16, abi.Uint16, abi.Uint8, abi.Bool},
			sqrtPriceX96, tick, observationIndex, observationCardinality, observationCardinalityNext, feeProtocol, unlocked)
		require.NoError(t, err)

		got, err := DecodeSlot0(data)
		require.NoError(t, err)
		require.Zero(t, sqrtPriceX96.Cmp(got.SqrtPriceX96))
		require.Zero(t, tick.Cmp(got.Tick))
		require.Zero(t, big.NewInt(int64(observationIndex)).Cmp(got.ObservationIndex))
		require.Zero(t, big.NewInt(int64(observationCardinality)).Cmp(got.ObservationCardinality))
		require.Zero(t, big.NewInt(int64(observationCardinalityNext)).Cmp(got.ObservationCardinalityNext))
		require.Zero(t, big.NewInt(int64(feeProtocol)).Cmp(got.FeeProtocol))
		require.Equal(t, unlocked, got.Unlocked)
	})
}

// go test -run '^$' -fuzz '^FuzzPoolSimpleDecodersDoNotPanic$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzPoolSimpleDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 33))
	f.Add(make([]byte, 64))
	f.Add(make([]byte, 256))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeFactory(data)
		_, _ = DecodeToken0(data)
		_, _ = DecodeToken1(data)
		_, _ = DecodePoolFee(data)
		_, _ = DecodePoolTickSpacing(data)
		_, _ = DecodeMaxLiquidityPerTick(data)
		_, _ = DecodeFeeGrowthGlobal0X128(data)
		_, _ = DecodeFeeGrowthGlobal1X128(data)
		_, _ = DecodeProtocolFees(data)
		_, _ = DecodePoolLiquidity(data)
		_, _ = DecodeTicks(data)
		_, _ = DecodeTickBitmap(data)
		_, _ = DecodePositions(data)
		_, _ = DecodePoolObservations(data)
		_, _ = DecodeObserve(data)
		_, _ = DecodeSnapshotCumulativesInside(data)
		_, _ = DecodeSlot0(data)
		_, _, _ = DecodeMint(data)
		_, _, _ = DecodeCollect(data)
		_, _, _ = DecodeBurn(data)
		_, _, _ = DecodeSwap(data)
		_, _, _ = DecodeCollectProtocol(data)
	})
}

// go test -run '^$' -fuzz '^FuzzExtractMintRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzExtractMintRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), []byte{0x00, 0x64}, []byte{0xff, 0x9c}, []byte{0x01, 0xf4})
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, make([]byte, 20), []byte{0x00}, []byte{0x00}, []byte{0x00})

	f.Fuzz(func(t *testing.T, senderRaw, ownerRaw, tickLowerRaw, tickUpperRaw, amountRaw []byte) {
		sender := types.NewAddressFromBytes(senderRaw)
		owner := types.NewAddressFromBytes(ownerRaw)
		tickLower := new(big.Int).Mod(new(big.Int).SetBytes(tickLowerRaw), new(big.Int).Lsh(big.NewInt(1), 24))
		if tickLower.Cmp(new(big.Int).Lsh(big.NewInt(1), 23)) >= 0 {
			tickLower.Sub(tickLower, new(big.Int).Lsh(big.NewInt(1), 24))
		}
		tickUpper := new(big.Int).Mod(new(big.Int).SetBytes(tickUpperRaw), new(big.Int).Lsh(big.NewInt(1), 24))
		if tickUpper.Cmp(new(big.Int).Lsh(big.NewInt(1), 23)) >= 0 {
			tickUpper.Sub(tickUpper, new(big.Int).Lsh(big.NewInt(1), 24))
		}
		if len(amountRaw) > 16 {
			amountRaw = amountRaw[len(amountRaw)-16:]
		}
		amount := new(big.Int).SetBytes(amountRaw)

		topic0, ok := mintEvent.Topic0()
		require.True(t, ok)
		tickLowerWord, err := abi.Pack(abi.Types{abi.Int24}, tickLower)
		require.NoError(t, err)
		tickUpperWord, err := abi.Pack(abi.Types{abi.Int24}, tickUpper)
		require.NoError(t, err)
		data, err := abi.Pack(abi.Types{abi.Address, abi.Uint128, abi.Uint256, abi.Uint256}, sender, amount, big.NewInt(1), big.NewInt(2))
		require.NoError(t, err)

		log := tx.Log{
			Address: testPool,
			Topics: []*types.Hash{
				types.NewHashFromBytes(topic0[:]),
				types.NewHashFromBytes(owner.Bytes()),
				types.NewHashFromBytes(tickLowerWord),
				types.NewHashFromBytes(tickUpperWord),
			},
			Data: data,
		}

		minted, err := ExtractMint([]tx.Log{log})
		require.NoError(t, err)
		require.Len(t, minted, 1)
		require.True(t, minted[0].Sender.Equal(sender))
		require.True(t, minted[0].Owner.Equal(owner))
		require.Zero(t, tickLower.Cmp(minted[0].TickLower))
		require.Zero(t, tickUpper.Cmp(minted[0].TickUpper))
		require.Zero(t, amount.Cmp(minted[0].Amount))
	})
}

// go test -run '^$' -fuzz '^FuzzExtractSwapRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzExtractSwapRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), []byte{0x01}, []byte{0x02})
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, make([]byte, 20), []byte{0xff}, []byte{0x00})

	f.Fuzz(func(t *testing.T, senderRaw, recipientRaw, amount0Raw, amount1Raw []byte) {
		sender := types.NewAddressFromBytes(senderRaw)
		recipient := types.NewAddressFromBytes(recipientRaw)
		if len(amount0Raw) > 32 {
			amount0Raw = amount0Raw[len(amount0Raw)-32:]
		}
		amount0 := new(big.Int).Mod(new(big.Int).SetBytes(amount0Raw), new(big.Int).Lsh(big.NewInt(1), 256))
		if amount0.Cmp(new(big.Int).Lsh(big.NewInt(1), 255)) >= 0 {
			amount0.Sub(amount0, new(big.Int).Lsh(big.NewInt(1), 256))
		}
		if len(amount1Raw) > 32 {
			amount1Raw = amount1Raw[len(amount1Raw)-32:]
		}
		amount1 := new(big.Int).Mod(new(big.Int).SetBytes(amount1Raw), new(big.Int).Lsh(big.NewInt(1), 256))
		if amount1.Cmp(new(big.Int).Lsh(big.NewInt(1), 255)) >= 0 {
			amount1.Sub(amount1, new(big.Int).Lsh(big.NewInt(1), 256))
		}

		topic0, ok := swapEvent.Topic0()
		require.True(t, ok)
		data, err := abi.Pack(abi.Types{abi.Int256, abi.Int256, abi.Uint160, abi.Uint128, abi.Int24}, amount0, amount1, big.NewInt(1), big.NewInt(0), big.NewInt(0))
		require.NoError(t, err)

		log := tx.Log{
			Address: testPool,
			Topics: []*types.Hash{
				types.NewHashFromBytes(topic0[:]),
				types.NewHashFromBytes(sender.Bytes()),
				types.NewHashFromBytes(recipient.Bytes()),
			},
			Data: data,
		}

		swapped, err := ExtractSwap([]tx.Log{log})
		require.NoError(t, err)
		require.Len(t, swapped, 1)
		require.True(t, swapped[0].Sender.Equal(sender))
		require.True(t, swapped[0].Recipient.Equal(recipient))
		require.Zero(t, amount0.Cmp(swapped[0].Amount0))
		require.Zero(t, amount1.Cmp(swapped[0].Amount1))
	})
}

// go test -run '^$' -fuzz '^FuzzExtractEventsDoNotPanic$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzExtractEventsDoNotPanic(f *testing.F) {
	f.Add(make([]byte, 32), []byte{})
	f.Add(make([]byte, 32), make([]byte, 128))

	f.Fuzz(func(t *testing.T, topicRaw, data []byte) {
		logs := []tx.Log{{Address: testPool, Topics: []*types.Hash{types.NewHashFromBytes(topicRaw)}, Data: data}}
		_, _ = ExtractInitialize(logs)
		_, _ = ExtractMint(logs)
		_, _ = ExtractCollect(logs)
		_, _ = ExtractBurn(logs)
		_, _ = ExtractSwap(logs)
		_, _ = ExtractFlash(logs)
		_, _ = ExtractIncreaseObservationCardinalityNext(logs)
		_, _ = ExtractSetFeeProtocol(logs)
		_, _ = ExtractCollectProtocol(logs)
	})
}
