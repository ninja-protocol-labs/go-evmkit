package v4

import (
	"context"
	"encoding/json/jsontext"
	"errors"
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

// go test -run '^$' -fuzz '^FuzzExtractInitializeDoesNotPanic$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzExtractInitializeDoesNotPanic(f *testing.F) {
	t0, ok := initializeEvent.Topic0()
	require.True(f, ok)

	f.Add(t0[:], make([]byte, 32), make([]byte, 32), make([]byte, 32), []byte{})
	f.Add(t0[:], make([]byte, 32), make([]byte, 32), make([]byte, 32), make([]byte, 224))
	f.Add([]byte{}, []byte{}, []byte{}, []byte{}, []byte{})
	f.Add(make([]byte, 31), make([]byte, 31), make([]byte, 31), make([]byte, 31), make([]byte, 31))

	f.Fuzz(func(t *testing.T, topic0, topic1, topic2, topic3 []byte, data []byte) {
		require.NotPanics(t, func() {
			var topics []*types.Hash
			for _, raw := range [][]byte{topic0, topic1, topic2, topic3} {
				if raw == nil {
					continue
				}
				topics = append(topics, types.NewHashFromBytes(raw))
			}
			logs := []tx.Log{{Address: testPoolManager, Topics: topics, Data: data}}
			_, _ = ExtractInitialize(logs)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzExtractModifyLiquidityDoesNotPanic$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzExtractModifyLiquidityDoesNotPanic(f *testing.F) {
	t0, ok := modifyLiquidityEvent.Topic0()
	require.True(f, ok)

	f.Add(t0[:], make([]byte, 32), make([]byte, 32), []byte{})
	f.Add(t0[:], make([]byte, 32), make([]byte, 32), make([]byte, 128))
	f.Add([]byte{}, []byte{}, []byte{}, []byte{})
	f.Add(make([]byte, 31), make([]byte, 31), make([]byte, 31), make([]byte, 31))

	f.Fuzz(func(t *testing.T, topic0, topic1, topic2 []byte, data []byte) {
		require.NotPanics(t, func() {
			var topics []*types.Hash
			for _, raw := range [][]byte{topic0, topic1, topic2} {
				if raw == nil {
					continue
				}
				topics = append(topics, types.NewHashFromBytes(raw))
			}
			logs := []tx.Log{{Address: testPoolManager, Topics: topics, Data: data}}
			_, _ = ExtractModifyLiquidity(logs)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzExtractSwapDoesNotPanic$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzExtractSwapDoesNotPanic(f *testing.F) {
	t0, ok := swapEvent.Topic0()
	require.True(f, ok)

	f.Add(t0[:], make([]byte, 32), make([]byte, 32), []byte{})
	f.Add(t0[:], make([]byte, 32), make([]byte, 32), make([]byte, 192))
	f.Add([]byte{}, []byte{}, []byte{}, []byte{})
	f.Add(make([]byte, 31), make([]byte, 31), make([]byte, 31), make([]byte, 31))

	f.Fuzz(func(t *testing.T, topic0, topic1, topic2 []byte, data []byte) {
		require.NotPanics(t, func() {
			var topics []*types.Hash
			for _, raw := range [][]byte{topic0, topic1, topic2} {
				if raw == nil {
					continue
				}
				topics = append(topics, types.NewHashFromBytes(raw))
			}
			logs := []tx.Log{{Address: testPoolManager, Topics: topics, Data: data}}
			_, _ = ExtractSwap(logs)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzExtractDonateDoesNotPanic$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzExtractDonateDoesNotPanic(f *testing.F) {
	t0, ok := donateEvent.Topic0()
	require.True(f, ok)

	f.Add(t0[:], make([]byte, 32), make([]byte, 32), []byte{})
	f.Add(t0[:], make([]byte, 32), make([]byte, 32), make([]byte, 64))
	f.Add([]byte{}, []byte{}, []byte{}, []byte{})
	f.Add(make([]byte, 31), make([]byte, 31), make([]byte, 31), make([]byte, 31))

	f.Fuzz(func(t *testing.T, topic0, topic1, topic2 []byte, data []byte) {
		require.NotPanics(t, func() {
			var topics []*types.Hash
			for _, raw := range [][]byte{topic0, topic1, topic2} {
				if raw == nil {
					continue
				}
				topics = append(topics, types.NewHashFromBytes(raw))
			}
			logs := []tx.Log{{Address: testPoolManager, Topics: topics, Data: data}}
			_, _ = ExtractDonate(logs)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzExtractInitializeRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzExtractInitializeRoundTrip(f *testing.F) {
	f.Add([]byte{0x01}, []byte{0x00}, []byte{0x11}, []byte{0x22}, []byte{0x0b, 0xb8}, []byte{0x00, 0x3c}, []byte{0x00}, []byte{0x01})
	f.Add(make([]byte, 32), make([]byte, 20), make([]byte, 20), make([]byte, 20), []byte{0x01, 0xf4}, []byte{0x00, 0x0a}, make([]byte, 20), []byte{0xff})

	f.Fuzz(func(t *testing.T, idRaw, currency0Raw, currency1Raw, hooksRaw, feeRaw, tickSpacingRaw, sqrtPriceRaw, tickRaw []byte) {
		currency0 := types.NewAddressFromBytes(currency0Raw)
		currency1 := types.NewAddressFromBytes(currency1Raw)
		hooks := types.NewAddressFromBytes(hooksRaw)

		feeClamped := feeRaw
		if len(feeClamped) > 3 {
			feeClamped = feeClamped[len(feeClamped)-3:]
		}
		fee := new(big.Int).SetBytes(feeClamped)

		tickSpacingClamped := tickSpacingRaw
		if len(tickSpacingClamped) > 2 {
			tickSpacingClamped = tickSpacingClamped[len(tickSpacingClamped)-2:]
		}
		tickSpacing := new(big.Int).SetBytes(tickSpacingClamped)

		sqrtPriceClamped := sqrtPriceRaw
		if len(sqrtPriceClamped) > 20 {
			sqrtPriceClamped = sqrtPriceClamped[len(sqrtPriceClamped)-20:]
		}
		sqrtPriceX96 := new(big.Int).SetBytes(sqrtPriceClamped)

		tickClamped := tickRaw
		if len(tickClamped) > 2 {
			tickClamped = tickClamped[len(tickClamped)-2:]
		}
		tick := new(big.Int).SetBytes(tickClamped)

		t0, ok := initializeEvent.Topic0()
		require.True(t, ok)

		data, err := abi.Pack(abi.Types{abi.Uint24, abi.Int24, abi.Address, abi.Uint160, abi.Int24},
			fee, tickSpacing, hooks, sqrtPriceX96, tick)
		require.NoError(t, err)

		idClamped := idRaw
		if len(idClamped) > 32 {
			idClamped = idClamped[len(idClamped)-32:]
		}
		id := make([]byte, 32)
		copy(id[32-len(idClamped):], idClamped)

		logs := []tx.Log{{
			Address: testPoolManager,
			Topics: []*types.Hash{
				types.NewHashFromBytes(t0[:]),
				types.NewHashFromBytes(id),
				types.NewHashFromBytes(currency0.Bytes()),
				types.NewHashFromBytes(currency1.Bytes()),
			},
			Data: data,
		}}

		inits, err := ExtractInitialize(logs)
		require.NoError(t, err)
		require.Len(t, inits, 1)
		require.True(t, inits[0].Currency0.Equal(currency0))
		require.True(t, inits[0].Currency1.Equal(currency1))
		require.True(t, inits[0].Hooks.Equal(hooks))
		require.Zero(t, fee.Cmp(inits[0].Fee))
		require.Zero(t, tickSpacing.Cmp(inits[0].TickSpacing))
		require.Zero(t, sqrtPriceX96.Cmp(inits[0].SqrtPriceX96))
		require.Zero(t, tick.Cmp(inits[0].Tick))
	})
}

// go test -run '^$' -fuzz '^FuzzExtractModifyLiquidityRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzExtractModifyLiquidityRoundTrip(f *testing.F) {
	f.Add([]byte{0x02}, []byte{0x00}, []byte{0xff, 0x9c}, []byte{0x00, 0x64}, []byte{0x03, 0xe8}, []byte{0x0a})
	f.Add(make([]byte, 32), make([]byte, 20), []byte{}, []byte{}, []byte{}, make([]byte, 32))

	f.Fuzz(func(t *testing.T, idRaw, senderRaw, tickLowerRaw, tickUpperRaw, liquidityDeltaRaw, saltRaw []byte) {
		sender := types.NewAddressFromBytes(senderRaw)

		tickLowerClamped := tickLowerRaw
		if len(tickLowerClamped) > 2 {
			tickLowerClamped = tickLowerClamped[len(tickLowerClamped)-2:]
		}
		tickLower := new(big.Int).SetBytes(tickLowerClamped)

		tickUpperClamped := tickUpperRaw
		if len(tickUpperClamped) > 2 {
			tickUpperClamped = tickUpperClamped[len(tickUpperClamped)-2:]
		}
		tickUpper := new(big.Int).SetBytes(tickUpperClamped)

		liquidityDeltaClamped := liquidityDeltaRaw
		if len(liquidityDeltaClamped) > 31 {
			liquidityDeltaClamped = liquidityDeltaClamped[len(liquidityDeltaClamped)-31:]
		}
		liquidityDelta := new(big.Int).SetBytes(liquidityDeltaClamped)

		salt := make([]byte, 32)
		saltClamped := saltRaw
		if len(saltClamped) > 32 {
			saltClamped = saltClamped[len(saltClamped)-32:]
		}
		copy(salt[32-len(saltClamped):], saltClamped)

		t0, ok := modifyLiquidityEvent.Topic0()
		require.True(t, ok)

		data, err := abi.Pack(abi.Types{abi.Int24, abi.Int24, abi.Int256, abi.Bytes32},
			tickLower, tickUpper, liquidityDelta, salt)
		require.NoError(t, err)

		idClamped := idRaw
		if len(idClamped) > 32 {
			idClamped = idClamped[len(idClamped)-32:]
		}
		id := make([]byte, 32)
		copy(id[32-len(idClamped):], idClamped)

		logs := []tx.Log{{
			Address: testPoolManager,
			Topics: []*types.Hash{
				types.NewHashFromBytes(t0[:]),
				types.NewHashFromBytes(id),
				types.NewHashFromBytes(sender.Bytes()),
			},
			Data: data,
		}}

		mods, err := ExtractModifyLiquidity(logs)
		require.NoError(t, err)
		require.Len(t, mods, 1)
		require.True(t, mods[0].Sender.Equal(sender))
		require.Zero(t, tickLower.Cmp(mods[0].TickLower))
		require.Zero(t, tickUpper.Cmp(mods[0].TickUpper))
		require.Zero(t, liquidityDelta.Cmp(mods[0].LiquidityDelta))
		var wantSalt [32]byte
		copy(wantSalt[:], salt)
		require.Equal(t, wantSalt, mods[0].Salt)
	})
}

// go test -run '^$' -fuzz '^FuzzExtractSwapRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzExtractSwapRoundTrip(f *testing.F) {
	f.Add([]byte{0x05}, []byte{0x00}, []byte{0x00, 0x64}, []byte{0x00, 0x32}, []byte{0x01, 0x00}, []byte{0x03, 0xe8}, []byte{0x00, 0x2a}, []byte{0x0b, 0xb8})
	f.Add(make([]byte, 32), make([]byte, 20), []byte{}, []byte{}, []byte{}, []byte{}, []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, idRaw, senderRaw, amount0Raw, amount1Raw, sqrtPriceRaw, liquidityRaw, tickRaw, feeRaw []byte) {
		sender := types.NewAddressFromBytes(senderRaw)

		amount0Clamped := amount0Raw
		if len(amount0Clamped) > 15 {
			amount0Clamped = amount0Clamped[len(amount0Clamped)-15:]
		}
		amount0 := new(big.Int).SetBytes(amount0Clamped)

		amount1Clamped := amount1Raw
		if len(amount1Clamped) > 15 {
			amount1Clamped = amount1Clamped[len(amount1Clamped)-15:]
		}
		amount1 := new(big.Int).SetBytes(amount1Clamped)

		sqrtPriceClamped := sqrtPriceRaw
		if len(sqrtPriceClamped) > 20 {
			sqrtPriceClamped = sqrtPriceClamped[len(sqrtPriceClamped)-20:]
		}
		sqrtPriceX96 := new(big.Int).SetBytes(sqrtPriceClamped)

		liquidityClamped := liquidityRaw
		if len(liquidityClamped) > 16 {
			liquidityClamped = liquidityClamped[len(liquidityClamped)-16:]
		}
		liquidity := new(big.Int).SetBytes(liquidityClamped)

		tickClamped := tickRaw
		if len(tickClamped) > 2 {
			tickClamped = tickClamped[len(tickClamped)-2:]
		}
		tick := new(big.Int).SetBytes(tickClamped)

		feeClamped := feeRaw
		if len(feeClamped) > 3 {
			feeClamped = feeClamped[len(feeClamped)-3:]
		}
		fee := new(big.Int).SetBytes(feeClamped)

		t0, ok := swapEvent.Topic0()
		require.True(t, ok)

		data, err := abi.Pack(abi.Types{abi.Int128, abi.Int128, abi.Uint160, abi.Uint128, abi.Int24, abi.Uint24},
			amount0, amount1, sqrtPriceX96, liquidity, tick, fee)
		require.NoError(t, err)

		idClamped := idRaw
		if len(idClamped) > 32 {
			idClamped = idClamped[len(idClamped)-32:]
		}
		id := make([]byte, 32)
		copy(id[32-len(idClamped):], idClamped)

		logs := []tx.Log{{
			Address: testPoolManager,
			Topics: []*types.Hash{
				types.NewHashFromBytes(t0[:]),
				types.NewHashFromBytes(id),
				types.NewHashFromBytes(sender.Bytes()),
			},
			Data: data,
		}}

		swaps, err := ExtractSwap(logs)
		require.NoError(t, err)
		require.Len(t, swaps, 1)
		require.True(t, swaps[0].Sender.Equal(sender))
		require.Zero(t, amount0.Cmp(swaps[0].Amount0))
		require.Zero(t, amount1.Cmp(swaps[0].Amount1))
		require.Zero(t, sqrtPriceX96.Cmp(swaps[0].SqrtPriceX96))
		require.Zero(t, liquidity.Cmp(swaps[0].Liquidity))
		require.Zero(t, tick.Cmp(swaps[0].Tick))
		require.Zero(t, fee.Cmp(swaps[0].Fee))
	})
}

// go test -run '^$' -fuzz '^FuzzExtractDonateRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzExtractDonateRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x01}, []byte{0x64})
	f.Add(make([]byte, 20), []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, senderRaw, amount0Raw, amount1Raw []byte) {
		sender := types.NewAddressFromBytes(senderRaw)

		amount0Clamped := amount0Raw
		if len(amount0Clamped) > 32 {
			amount0Clamped = amount0Clamped[len(amount0Clamped)-32:]
		}
		amount0 := new(big.Int).SetBytes(amount0Clamped)

		amount1Clamped := amount1Raw
		if len(amount1Clamped) > 32 {
			amount1Clamped = amount1Clamped[len(amount1Clamped)-32:]
		}
		amount1 := new(big.Int).SetBytes(amount1Clamped)

		t0, ok := donateEvent.Topic0()
		require.True(t, ok)

		data, err := abi.Pack(abi.Types{abi.Uint256, abi.Uint256}, amount0, amount1)
		require.NoError(t, err)

		logs := []tx.Log{{
			Address: testPoolManager,
			Topics: []*types.Hash{
				types.NewHashFromBytes(t0[:]),
				types.NewHashFromBytes([]byte{0x01}),
				types.NewHashFromBytes(sender.Bytes()),
			},
			Data: data,
		}}

		donates, err := ExtractDonate(logs)
		require.NoError(t, err)
		require.Len(t, donates, 1)
		require.True(t, donates[0].Sender.Equal(sender))
		require.Zero(t, amount0.Cmp(donates[0].Amount0))
		require.Zero(t, amount1.Cmp(donates[0].Amount1))
	})
}

// go test -run '^$' -fuzz '^FuzzAllPoolManagerDecodersDoNotPanic$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzAllPoolManagerDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x01})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 33))
	f.Add(make([]byte, 64))
	f.Add(make([]byte, 96))

	f.Fuzz(func(t *testing.T, data []byte) {
		require.NotPanics(t, func() {
			_, _ = DecodeInitialize(data)
			_, _, _ = DecodeModifyLiquidity(data)
			_, _ = DecodeSwap(data)
			_, _ = DecodeDonate(data)
			_, _ = DecodeUnlock(data)
			_, _ = DecodeSettle(data)
			_, _ = DecodeSettleFor(data)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeInitializeRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeInitializeRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x11}, []byte{0x22}, []byte{0x0b, 0xb8}, []byte{0x00, 0x3c}, []byte{0x01, 0x00})
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 20), []byte{}, []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, currency0Raw, currency1Raw, hooksRaw, feeRaw, tickSpacingRaw, sqrtPriceRaw []byte) {
		currency0 := types.NewAddressFromBytes(currency0Raw)
		currency1 := types.NewAddressFromBytes(currency1Raw)
		hooks := types.NewAddressFromBytes(hooksRaw)

		feeClamped := feeRaw
		if len(feeClamped) > 3 {
			feeClamped = feeClamped[len(feeClamped)-3:]
		}
		fee := new(big.Int).SetBytes(feeClamped)

		tickSpacingClamped := tickSpacingRaw
		if len(tickSpacingClamped) > 2 {
			tickSpacingClamped = tickSpacingClamped[len(tickSpacingClamped)-2:]
		}
		tickSpacing := new(big.Int).SetBytes(tickSpacingClamped)

		sqrtPriceClamped := sqrtPriceRaw
		if len(sqrtPriceClamped) > 20 {
			sqrtPriceClamped = sqrtPriceClamped[len(sqrtPriceClamped)-20:]
		}
		sqrtPriceX96 := new(big.Int).SetBytes(sqrtPriceClamped)

		key := NewPoolKey(currency0, currency1, fee, tickSpacing, hooks)

		data, err := EncodeInitialize(key, sqrtPriceX96)
		require.NoError(t, err)

		vals, err := initializeFn.DecodeCall(data)
		require.NoError(t, err)
		keyVals, ok := vals[0].([]any)
		require.True(t, ok)
		require.True(t, keyVals[0].(*types.Address).Equal(currency0))
		require.True(t, keyVals[1].(*types.Address).Equal(currency1))
		require.Zero(t, fee.Cmp(keyVals[2].(*big.Int)))
		require.Zero(t, tickSpacing.Cmp(keyVals[3].(*big.Int)))
		require.True(t, keyVals[4].(*types.Address).Equal(hooks))
		require.Zero(t, sqrtPriceX96.Cmp(vals[1].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeInitializeRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzDecodeInitializeRoundTrip(f *testing.F) {
	f.Add([]byte{0x00, 0x2a})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, tickRaw []byte) {
		tickClamped := tickRaw
		if len(tickClamped) > 2 {
			tickClamped = tickClamped[len(tickClamped)-2:]
		}
		tick := new(big.Int).SetBytes(tickClamped)

		data, err := abi.Pack(abi.Types{abi.Int24}, tick)
		require.NoError(t, err)

		got, err := DecodeInitialize(data)
		require.NoError(t, err)
		require.Zero(t, tick.Cmp(got))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeModifyLiquidityRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeModifyLiquidityRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x11}, []byte{0x0b, 0xb8}, []byte{0x00, 0x3c}, []byte{0x00}, []byte{0xff, 0x9c}, []byte{0x00, 0x64}, []byte{0x03, 0xe8}, []byte{0x0a}, []byte{0xde, 0xad})
	f.Add(make([]byte, 20), make([]byte, 20), []byte{}, []byte{}, make([]byte, 20), []byte{}, []byte{}, []byte{}, make([]byte, 32), []byte{})

	f.Fuzz(func(t *testing.T, currency0Raw, currency1Raw, feeRaw, tickSpacingRaw, hooksRaw, tickLowerRaw, tickUpperRaw, liquidityDeltaRaw, saltRaw, hookData []byte) {
		currency0 := types.NewAddressFromBytes(currency0Raw)
		currency1 := types.NewAddressFromBytes(currency1Raw)
		hooks := types.NewAddressFromBytes(hooksRaw)

		feeClamped := feeRaw
		if len(feeClamped) > 3 {
			feeClamped = feeClamped[len(feeClamped)-3:]
		}
		fee := new(big.Int).SetBytes(feeClamped)

		tickSpacingClamped := tickSpacingRaw
		if len(tickSpacingClamped) > 2 {
			tickSpacingClamped = tickSpacingClamped[len(tickSpacingClamped)-2:]
		}
		tickSpacing := new(big.Int).SetBytes(tickSpacingClamped)

		tickLowerClamped := tickLowerRaw
		if len(tickLowerClamped) > 2 {
			tickLowerClamped = tickLowerClamped[len(tickLowerClamped)-2:]
		}
		tickLower := new(big.Int).SetBytes(tickLowerClamped)

		tickUpperClamped := tickUpperRaw
		if len(tickUpperClamped) > 2 {
			tickUpperClamped = tickUpperClamped[len(tickUpperClamped)-2:]
		}
		tickUpper := new(big.Int).SetBytes(tickUpperClamped)

		liquidityDeltaClamped := liquidityDeltaRaw
		if len(liquidityDeltaClamped) > 31 {
			liquidityDeltaClamped = liquidityDeltaClamped[len(liquidityDeltaClamped)-31:]
		}
		liquidityDelta := new(big.Int).SetBytes(liquidityDeltaClamped)

		var salt [32]byte
		saltClamped := saltRaw
		if len(saltClamped) > 32 {
			saltClamped = saltClamped[len(saltClamped)-32:]
		}
		copy(salt[32-len(saltClamped):], saltClamped)

		key := NewPoolKey(currency0, currency1, fee, tickSpacing, hooks)
		params := NewLiquidityParams(tickLower, tickUpper, liquidityDelta, salt)

		data, err := EncodeModifyLiquidity(key, params, hookData)
		require.NoError(t, err)

		vals, err := modifyLiquidityFn.DecodeCall(data)
		require.NoError(t, err)
		paramsVals, ok := vals[1].([]any)
		require.True(t, ok)
		require.Zero(t, tickLower.Cmp(paramsVals[0].(*big.Int)))
		require.Zero(t, tickUpper.Cmp(paramsVals[1].(*big.Int)))
		require.Zero(t, liquidityDelta.Cmp(paramsVals[2].(*big.Int)))
		require.Equal(t, salt[:], paramsVals[3].([]byte))
		require.Equal(t, hookData, vals[2].([]byte))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeModifyLiquidityRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
//
// Covers both BalanceDelta words together, each half (upper/lower 16 bytes) independently signed,
// to exercise the sign-extension logic across random positive/negative combinations.
func FuzzDecodeModifyLiquidityRoundTrip(f *testing.F) {
	f.Add([]byte{0x00, 0x64}, []byte{0x80, 0x32}, []byte{0x80, 0x07}, []byte{0x00, 0x03})
	f.Add([]byte{}, []byte{}, []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, callerAmount0Raw, callerAmount1Raw, feesAmount0Raw, feesAmount1Raw []byte) {
		callerHalf0 := callerAmount0Raw
		if len(callerHalf0) > 16 {
			callerHalf0 = callerHalf0[len(callerHalf0)-16:]
		}
		callerWord := make([]byte, 32)
		copy(callerWord[16-len(callerHalf0):16], callerHalf0)
		callerAmount0 := new(big.Int).SetBytes(callerWord[:16])
		if callerWord[0]&0x80 != 0 {
			callerAmount0.Sub(callerAmount0, new(big.Int).Lsh(big.NewInt(1), 128))
		}

		callerHalf1 := callerAmount1Raw
		if len(callerHalf1) > 16 {
			callerHalf1 = callerHalf1[len(callerHalf1)-16:]
		}
		copy(callerWord[32-len(callerHalf1):32], callerHalf1)
		callerAmount1 := new(big.Int).SetBytes(callerWord[16:])
		if callerWord[16]&0x80 != 0 {
			callerAmount1.Sub(callerAmount1, new(big.Int).Lsh(big.NewInt(1), 128))
		}

		feesHalf0 := feesAmount0Raw
		if len(feesHalf0) > 16 {
			feesHalf0 = feesHalf0[len(feesHalf0)-16:]
		}
		feesWord := make([]byte, 32)
		copy(feesWord[16-len(feesHalf0):16], feesHalf0)
		feesAmount0 := new(big.Int).SetBytes(feesWord[:16])
		if feesWord[0]&0x80 != 0 {
			feesAmount0.Sub(feesAmount0, new(big.Int).Lsh(big.NewInt(1), 128))
		}

		feesHalf1 := feesAmount1Raw
		if len(feesHalf1) > 16 {
			feesHalf1 = feesHalf1[len(feesHalf1)-16:]
		}
		copy(feesWord[32-len(feesHalf1):32], feesHalf1)
		feesAmount1 := new(big.Int).SetBytes(feesWord[16:])
		if feesWord[16]&0x80 != 0 {
			feesAmount1.Sub(feesAmount1, new(big.Int).Lsh(big.NewInt(1), 128))
		}

		data, err := abi.Pack(abi.Types{abi.Bytes32, abi.Bytes32}, callerWord, feesWord)
		require.NoError(t, err)

		callerDelta, feesAccrued, err := DecodeModifyLiquidity(data)
		require.NoError(t, err)
		require.Zero(t, callerAmount0.Cmp(callerDelta.Amount0))
		require.Zero(t, callerAmount1.Cmp(callerDelta.Amount1))
		require.Zero(t, feesAmount0.Cmp(feesAccrued.Amount0))
		require.Zero(t, feesAmount1.Cmp(feesAccrued.Amount1))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeSwapRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeSwapRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x11}, []byte{0x0b, 0xb8}, []byte{0x00, 0x3c}, []byte{0x00}, true, []byte{0x00, 0x64}, []byte{0x01, 0x00}, []byte{0xbe, 0xef})
	f.Add(make([]byte, 20), make([]byte, 20), []byte{}, []byte{}, make([]byte, 20), false, []byte{}, []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, currency0Raw, currency1Raw, feeRaw, tickSpacingRaw, hooksRaw []byte, zeroForOne bool, amountSpecifiedRaw, sqrtPriceLimitRaw, hookData []byte) {
		currency0 := types.NewAddressFromBytes(currency0Raw)
		currency1 := types.NewAddressFromBytes(currency1Raw)
		hooks := types.NewAddressFromBytes(hooksRaw)

		feeClamped := feeRaw
		if len(feeClamped) > 3 {
			feeClamped = feeClamped[len(feeClamped)-3:]
		}
		fee := new(big.Int).SetBytes(feeClamped)

		tickSpacingClamped := tickSpacingRaw
		if len(tickSpacingClamped) > 2 {
			tickSpacingClamped = tickSpacingClamped[len(tickSpacingClamped)-2:]
		}
		tickSpacing := new(big.Int).SetBytes(tickSpacingClamped)

		amountSpecifiedClamped := amountSpecifiedRaw
		if len(amountSpecifiedClamped) > 31 {
			amountSpecifiedClamped = amountSpecifiedClamped[len(amountSpecifiedClamped)-31:]
		}
		amountSpecified := new(big.Int).SetBytes(amountSpecifiedClamped)

		sqrtPriceLimitClamped := sqrtPriceLimitRaw
		if len(sqrtPriceLimitClamped) > 20 {
			sqrtPriceLimitClamped = sqrtPriceLimitClamped[len(sqrtPriceLimitClamped)-20:]
		}
		sqrtPriceLimitX96 := new(big.Int).SetBytes(sqrtPriceLimitClamped)

		key := NewPoolKey(currency0, currency1, fee, tickSpacing, hooks)
		params := NewSwapAmountParams(zeroForOne, amountSpecified, sqrtPriceLimitX96)

		data, err := EncodeSwap(key, params, hookData)
		require.NoError(t, err)

		vals, err := swapFn.DecodeCall(data)
		require.NoError(t, err)
		paramsVals, ok := vals[1].([]any)
		require.True(t, ok)
		require.Equal(t, zeroForOne, paramsVals[0].(bool))
		require.Zero(t, amountSpecified.Cmp(paramsVals[1].(*big.Int)))
		require.Zero(t, sqrtPriceLimitX96.Cmp(paramsVals[2].(*big.Int)))
		require.Equal(t, hookData, vals[2].([]byte))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeSwapRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzDecodeSwapRoundTrip(f *testing.F) {
	f.Add([]byte{0x00, 0x64}, []byte{0x80, 0x32})
	f.Add([]byte{}, []byte{})

	f.Fuzz(func(t *testing.T, amount0Raw, amount1Raw []byte) {
		half0 := amount0Raw
		if len(half0) > 16 {
			half0 = half0[len(half0)-16:]
		}
		word := make([]byte, 32)
		copy(word[16-len(half0):16], half0)
		amount0 := new(big.Int).SetBytes(word[:16])
		if word[0]&0x80 != 0 {
			amount0.Sub(amount0, new(big.Int).Lsh(big.NewInt(1), 128))
		}

		half1 := amount1Raw
		if len(half1) > 16 {
			half1 = half1[len(half1)-16:]
		}
		copy(word[32-len(half1):32], half1)
		amount1 := new(big.Int).SetBytes(word[16:])
		if word[16]&0x80 != 0 {
			amount1.Sub(amount1, new(big.Int).Lsh(big.NewInt(1), 128))
		}

		data, err := abi.Pack(abi.Types{abi.Bytes32}, word)
		require.NoError(t, err)

		delta, err := DecodeSwap(data)
		require.NoError(t, err)
		require.Zero(t, amount0.Cmp(delta.Amount0))
		require.Zero(t, amount1.Cmp(delta.Amount1))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeDonateRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeDonateRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x11}, []byte{0x0b, 0xb8}, []byte{0x00, 0x3c}, []byte{0x00}, []byte{0x00, 0x64}, []byte{0x00, 0xc8}, []byte{0x01})
	f.Add(make([]byte, 20), make([]byte, 20), []byte{}, []byte{}, make([]byte, 20), []byte{}, []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, currency0Raw, currency1Raw, feeRaw, tickSpacingRaw, hooksRaw, amount0Raw, amount1Raw, hookData []byte) {
		currency0 := types.NewAddressFromBytes(currency0Raw)
		currency1 := types.NewAddressFromBytes(currency1Raw)
		hooks := types.NewAddressFromBytes(hooksRaw)

		feeClamped := feeRaw
		if len(feeClamped) > 3 {
			feeClamped = feeClamped[len(feeClamped)-3:]
		}
		fee := new(big.Int).SetBytes(feeClamped)

		tickSpacingClamped := tickSpacingRaw
		if len(tickSpacingClamped) > 2 {
			tickSpacingClamped = tickSpacingClamped[len(tickSpacingClamped)-2:]
		}
		tickSpacing := new(big.Int).SetBytes(tickSpacingClamped)

		amount0Clamped := amount0Raw
		if len(amount0Clamped) > 32 {
			amount0Clamped = amount0Clamped[len(amount0Clamped)-32:]
		}
		amount0 := new(big.Int).SetBytes(amount0Clamped)

		amount1Clamped := amount1Raw
		if len(amount1Clamped) > 32 {
			amount1Clamped = amount1Clamped[len(amount1Clamped)-32:]
		}
		amount1 := new(big.Int).SetBytes(amount1Clamped)

		key := NewPoolKey(currency0, currency1, fee, tickSpacing, hooks)

		data, err := EncodeDonate(key, amount0, amount1, hookData)
		require.NoError(t, err)

		vals, err := donateFn.DecodeCall(data)
		require.NoError(t, err)
		require.Zero(t, amount0.Cmp(vals[1].(*big.Int)))
		require.Zero(t, amount1.Cmp(vals[2].(*big.Int)))
		require.Equal(t, hookData, vals[3].([]byte))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeDonateRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzDecodeDonateRoundTripBalanceDelta(f *testing.F) {
	f.Add([]byte{0x00, 0x32}, []byte{0x80, 0x19})
	f.Add([]byte{}, []byte{})

	f.Fuzz(func(t *testing.T, amount0Raw, amount1Raw []byte) {
		half0 := amount0Raw
		if len(half0) > 16 {
			half0 = half0[len(half0)-16:]
		}
		word := make([]byte, 32)
		copy(word[16-len(half0):16], half0)
		amount0 := new(big.Int).SetBytes(word[:16])
		if word[0]&0x80 != 0 {
			amount0.Sub(amount0, new(big.Int).Lsh(big.NewInt(1), 128))
		}

		half1 := amount1Raw
		if len(half1) > 16 {
			half1 = half1[len(half1)-16:]
		}
		copy(word[32-len(half1):32], half1)
		amount1 := new(big.Int).SetBytes(word[16:])
		if word[16]&0x80 != 0 {
			amount1.Sub(amount1, new(big.Int).Lsh(big.NewInt(1), 128))
		}

		data, err := abi.Pack(abi.Types{abi.Bytes32}, word)
		require.NoError(t, err)

		delta, err := DecodeDonate(data)
		require.NoError(t, err)
		require.Zero(t, amount0.Cmp(delta.Amount0))
		require.Zero(t, amount1.Cmp(delta.Amount1))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeUnlockRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeUnlockRoundTrip(f *testing.F) {
	f.Add([]byte{0xca, 0xfe})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, unlockData []byte) {
		data, err := EncodeUnlock(unlockData)
		require.NoError(t, err)

		vals, err := unlockFn.DecodeCall(data)
		require.NoError(t, err)
		require.Equal(t, unlockData, vals[0].([]byte))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeUnlockRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzDecodeUnlockRoundTrip(f *testing.F) {
	f.Add([]byte{0x01, 0x02, 0x03})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, out []byte) {
		data, err := abi.Pack(abi.Types{abi.Bytes}, out)
		require.NoError(t, err)

		got, err := DecodeUnlock(data)
		require.NoError(t, err)
		require.Equal(t, out, got)
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeSyncRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeSyncRoundTrip(f *testing.F) {
	f.Add([]byte{0x00})
	f.Add(make([]byte, 20))

	f.Fuzz(func(t *testing.T, currencyRaw []byte) {
		currency := types.NewAddressFromBytes(currencyRaw)

		data, err := EncodeSync(currency)
		require.NoError(t, err)

		vals, err := syncFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(currency))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeTakeRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeTakeRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x11}, []byte{0x01, 0xf4})
	f.Add(make([]byte, 20), make([]byte, 20), []byte{})

	f.Fuzz(func(t *testing.T, currencyRaw, toRaw, amountRaw []byte) {
		currency := types.NewAddressFromBytes(currencyRaw)
		to := types.NewAddressFromBytes(toRaw)

		amountClamped := amountRaw
		if len(amountClamped) > 32 {
			amountClamped = amountClamped[len(amountClamped)-32:]
		}
		amount := new(big.Int).SetBytes(amountClamped)

		data, err := EncodeTake(currency, to, amount)
		require.NoError(t, err)

		vals, err := takeFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(currency))
		require.True(t, vals[1].(*types.Address).Equal(to))
		require.Zero(t, amount.Cmp(vals[2].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeSettleRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzDecodeSettleRoundTrip(f *testing.F) {
	f.Add([]byte{0x01, 0xf4})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, raw []byte) {
		clamped := raw
		if len(clamped) > 32 {
			clamped = clamped[len(clamped)-32:]
		}
		paid := new(big.Int).SetBytes(clamped)

		data, err := abi.Pack(abi.Types{abi.Uint256}, paid)
		require.NoError(t, err)

		got, err := DecodeSettle(data)
		require.NoError(t, err)
		require.Zero(t, paid.Cmp(got))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeSettleForRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeSettleForRoundTrip(f *testing.F) {
	f.Add([]byte{0x00})
	f.Add(make([]byte, 20))

	f.Fuzz(func(t *testing.T, recipientRaw []byte) {
		recipient := types.NewAddressFromBytes(recipientRaw)

		data, err := EncodeSettleFor(recipient)
		require.NoError(t, err)

		vals, err := settleForFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(recipient))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeSettleForRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzDecodeSettleForRoundTrip(f *testing.F) {
	f.Add([]byte{0x01, 0xc8})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, raw []byte) {
		clamped := raw
		if len(clamped) > 32 {
			clamped = clamped[len(clamped)-32:]
		}
		paid := new(big.Int).SetBytes(clamped)

		data, err := abi.Pack(abi.Types{abi.Uint256}, paid)
		require.NoError(t, err)

		got, err := DecodeSettleFor(data)
		require.NoError(t, err)
		require.Zero(t, paid.Cmp(got))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeClearRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeClearRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x00, 0x2a})
	f.Add(make([]byte, 20), []byte{})

	f.Fuzz(func(t *testing.T, currencyRaw, amountRaw []byte) {
		currency := types.NewAddressFromBytes(currencyRaw)

		amountClamped := amountRaw
		if len(amountClamped) > 32 {
			amountClamped = amountClamped[len(amountClamped)-32:]
		}
		amount := new(big.Int).SetBytes(amountClamped)

		data, err := EncodeClear(currency, amount)
		require.NoError(t, err)

		vals, err := clearFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(currency))
		require.Zero(t, amount.Cmp(vals[1].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeMintRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeMintRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x01}, []byte{0x00, 0x64})
	f.Add(make([]byte, 20), []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, toRaw, idRaw, amountRaw []byte) {
		to := types.NewAddressFromBytes(toRaw)

		idClamped := idRaw
		if len(idClamped) > 32 {
			idClamped = idClamped[len(idClamped)-32:]
		}
		id := new(big.Int).SetBytes(idClamped)

		amountClamped := amountRaw
		if len(amountClamped) > 32 {
			amountClamped = amountClamped[len(amountClamped)-32:]
		}
		amount := new(big.Int).SetBytes(amountClamped)

		data, err := EncodeMint(to, id, amount)
		require.NoError(t, err)

		vals, err := mintFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(to))
		require.Zero(t, id.Cmp(vals[1].(*big.Int)))
		require.Zero(t, amount.Cmp(vals[2].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeBurnRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeBurnRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x01}, []byte{0x00, 0x64})
	f.Add(make([]byte, 20), []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, fromRaw, idRaw, amountRaw []byte) {
		from := types.NewAddressFromBytes(fromRaw)

		idClamped := idRaw
		if len(idClamped) > 32 {
			idClamped = idClamped[len(idClamped)-32:]
		}
		id := new(big.Int).SetBytes(idClamped)

		amountClamped := amountRaw
		if len(amountClamped) > 32 {
			amountClamped = amountClamped[len(amountClamped)-32:]
		}
		amount := new(big.Int).SetBytes(amountClamped)

		data, err := EncodeBurn(from, id, amount)
		require.NoError(t, err)

		vals, err := burnFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(from))
		require.Zero(t, id.Cmp(vals[1].(*big.Int)))
		require.Zero(t, amount.Cmp(vals[2].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeUpdateDynamicLPFeeRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeUpdateDynamicLPFeeRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x11}, []byte{0x0b, 0xb8}, []byte{0x00, 0x3c}, []byte{0x00}, []byte{0x01, 0xf4})
	f.Add(make([]byte, 20), make([]byte, 20), []byte{}, []byte{}, make([]byte, 20), []byte{})

	f.Fuzz(func(t *testing.T, currency0Raw, currency1Raw, feeRaw, tickSpacingRaw, hooksRaw, newFeeRaw []byte) {
		currency0 := types.NewAddressFromBytes(currency0Raw)
		currency1 := types.NewAddressFromBytes(currency1Raw)
		hooks := types.NewAddressFromBytes(hooksRaw)

		feeClamped := feeRaw
		if len(feeClamped) > 3 {
			feeClamped = feeClamped[len(feeClamped)-3:]
		}
		fee := new(big.Int).SetBytes(feeClamped)

		tickSpacingClamped := tickSpacingRaw
		if len(tickSpacingClamped) > 2 {
			tickSpacingClamped = tickSpacingClamped[len(tickSpacingClamped)-2:]
		}
		tickSpacing := new(big.Int).SetBytes(tickSpacingClamped)

		newFeeClamped := newFeeRaw
		if len(newFeeClamped) > 3 {
			newFeeClamped = newFeeClamped[len(newFeeClamped)-3:]
		}
		newFee := new(big.Int).SetBytes(newFeeClamped)

		key := NewPoolKey(currency0, currency1, fee, tickSpacing, hooks)

		data, err := EncodeUpdateDynamicLPFee(key, newFee)
		require.NoError(t, err)

		vals, err := updateDynamicLPFeeFn.DecodeCall(data)
		require.NoError(t, err)
		keyVals, ok := vals[0].([]any)
		require.True(t, ok)
		require.True(t, keyVals[0].(*types.Address).Equal(currency0))
		require.Zero(t, newFee.Cmp(vals[1].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzPoolKeyPoolIDDeterministic$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzPoolKeyPoolIDDeterministic(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x11}, []byte{0x22}, []byte{0x0b, 0xb8}, []byte{0x00, 0x3c})
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 20), []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, currency0Raw, currency1Raw, hooksRaw, feeRaw, tickSpacingRaw []byte) {
		currency0 := types.NewAddressFromBytes(currency0Raw)
		currency1 := types.NewAddressFromBytes(currency1Raw)
		hooks := types.NewAddressFromBytes(hooksRaw)

		feeClamped := feeRaw
		if len(feeClamped) > 3 {
			feeClamped = feeClamped[len(feeClamped)-3:]
		}
		fee := new(big.Int).SetBytes(feeClamped)

		tickSpacingClamped := tickSpacingRaw
		if len(tickSpacingClamped) > 2 {
			tickSpacingClamped = tickSpacingClamped[len(tickSpacingClamped)-2:]
		}
		tickSpacing := new(big.Int).SetBytes(tickSpacingClamped)

		key := NewPoolKey(currency0, currency1, fee, tickSpacing, hooks)

		id1, err := key.PoolID()
		require.NoError(t, err)
		id2, err := key.PoolID()
		require.NoError(t, err)
		require.Equal(t, id1, id2)
	})
}

// go test -run '^$' -fuzz '^FuzzExtractProtocolFeeControllerUpdatedDoesNotPanic$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzExtractProtocolFeeControllerUpdatedDoesNotPanic(f *testing.F) {
	t0, ok := protocolFeeControllerUpdatedEvent.Topic0()
	require.True(f, ok)

	f.Add(t0[:], make([]byte, 32), []byte{})
	f.Add([]byte{}, []byte{}, []byte{})
	f.Add(make([]byte, 31), make([]byte, 31), make([]byte, 31))

	f.Fuzz(func(t *testing.T, topic0, topic1 []byte, data []byte) {
		require.NotPanics(t, func() {
			var topics []*types.Hash
			for _, raw := range [][]byte{topic0, topic1} {
				if raw == nil {
					continue
				}
				topics = append(topics, types.NewHashFromBytes(raw))
			}
			logs := []tx.Log{{Address: testPoolManager, Topics: topics, Data: data}}
			_, _ = ExtractProtocolFeeControllerUpdated(logs)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzExtractProtocolFeeControllerUpdatedRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzExtractProtocolFeeControllerUpdatedRoundTrip(f *testing.F) {
	f.Add([]byte{0x00})
	f.Add(make([]byte, 20))

	f.Fuzz(func(t *testing.T, controllerRaw []byte) {
		controller := types.NewAddressFromBytes(controllerRaw)

		t0, ok := protocolFeeControllerUpdatedEvent.Topic0()
		require.True(t, ok)

		logs := []tx.Log{{
			Address: testPoolManager,
			Topics: []*types.Hash{
				types.NewHashFromBytes(t0[:]),
				types.NewHashFromBytes(controller.Bytes()),
			},
			Data: nil,
		}}

		updates, err := ExtractProtocolFeeControllerUpdated(logs)
		require.NoError(t, err)
		require.Len(t, updates, 1)
		require.True(t, updates[0].ProtocolFeeController.Equal(controller))
	})
}

// go test -run '^$' -fuzz '^FuzzExtractProtocolFeeUpdatedDoesNotPanic$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzExtractProtocolFeeUpdatedDoesNotPanic(f *testing.F) {
	t0, ok := protocolFeeUpdatedEvent.Topic0()
	require.True(f, ok)

	f.Add(t0[:], make([]byte, 32), []byte{})
	f.Add([]byte{}, []byte{}, []byte{})
	f.Add(make([]byte, 31), make([]byte, 31), make([]byte, 31))

	f.Fuzz(func(t *testing.T, topic0, topic1 []byte, data []byte) {
		require.NotPanics(t, func() {
			var topics []*types.Hash
			for _, raw := range [][]byte{topic0, topic1} {
				if raw == nil {
					continue
				}
				topics = append(topics, types.NewHashFromBytes(raw))
			}
			logs := []tx.Log{{Address: testPoolManager, Topics: topics, Data: data}}
			_, _ = ExtractProtocolFeeUpdated(logs)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzExtractProtocolFeeUpdatedRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzExtractProtocolFeeUpdatedRoundTrip(f *testing.F) {
	f.Add([]byte{0x01}, []byte{0x01, 0xf4})
	f.Add([]byte{}, []byte{})

	f.Fuzz(func(t *testing.T, idRaw, protocolFeeRaw []byte) {
		idClamped := idRaw
		if len(idClamped) > 32 {
			idClamped = idClamped[len(idClamped)-32:]
		}
		id := make([]byte, 32)
		copy(id[32-len(idClamped):], idClamped)

		feeClamped := protocolFeeRaw
		if len(feeClamped) > 3 {
			feeClamped = feeClamped[len(feeClamped)-3:]
		}
		protocolFee := new(big.Int).SetBytes(feeClamped)

		t0, ok := protocolFeeUpdatedEvent.Topic0()
		require.True(t, ok)

		data, err := abi.Pack(abi.Types{abi.Uint24}, protocolFee)
		require.NoError(t, err)

		logs := []tx.Log{{
			Address: testPoolManager,
			Topics: []*types.Hash{
				types.NewHashFromBytes(t0[:]),
				types.NewHashFromBytes(id),
			},
			Data: data,
		}}

		updates, err := ExtractProtocolFeeUpdated(logs)
		require.NoError(t, err)
		require.Len(t, updates, 1)
		require.Zero(t, protocolFee.Cmp(updates[0].ProtocolFee))
	})
}

// go test -run '^$' -fuzz '^FuzzAllProtocolFeesDecodersDoNotPanic$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzAllProtocolFeesDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x01})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 64))

	f.Fuzz(func(t *testing.T, data []byte) {
		require.NotPanics(t, func() {
			_, _ = DecodeProtocolFeesAccrued(data)
			_, _ = DecodeCollectProtocolFees(data)
			_, _ = DecodeProtocolFeeController(data)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeProtocolFeesAccruedRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeProtocolFeesAccruedRoundTrip(f *testing.F) {
	f.Add([]byte{0x00})
	f.Add(make([]byte, 20))

	f.Fuzz(func(t *testing.T, currencyRaw []byte) {
		currency := types.NewAddressFromBytes(currencyRaw)

		data, err := EncodeProtocolFeesAccrued(currency)
		require.NoError(t, err)

		vals, err := protocolFeesAccruedFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(currency))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeProtocolFeesAccruedRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzDecodeProtocolFeesAccruedRoundTrip(f *testing.F) {
	f.Add([]byte{0x01, 0xf4})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, raw []byte) {
		clamped := raw
		if len(clamped) > 32 {
			clamped = clamped[len(clamped)-32:]
		}
		amount := new(big.Int).SetBytes(clamped)

		data, err := abi.Pack(abi.Types{abi.Uint256}, amount)
		require.NoError(t, err)

		got, err := DecodeProtocolFeesAccrued(data)
		require.NoError(t, err)
		require.Zero(t, amount.Cmp(got))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeSetProtocolFeeRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeSetProtocolFeeRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x11}, []byte{0x0b, 0xb8}, []byte{0x00, 0x3c}, []byte{0x00}, []byte{0x01, 0xf4})
	f.Add(make([]byte, 20), make([]byte, 20), []byte{}, []byte{}, make([]byte, 20), []byte{})

	f.Fuzz(func(t *testing.T, currency0Raw, currency1Raw, feeRaw, tickSpacingRaw, hooksRaw, newFeeRaw []byte) {
		currency0 := types.NewAddressFromBytes(currency0Raw)
		currency1 := types.NewAddressFromBytes(currency1Raw)
		hooks := types.NewAddressFromBytes(hooksRaw)

		feeClamped := feeRaw
		if len(feeClamped) > 3 {
			feeClamped = feeClamped[len(feeClamped)-3:]
		}
		fee := new(big.Int).SetBytes(feeClamped)

		tickSpacingClamped := tickSpacingRaw
		if len(tickSpacingClamped) > 2 {
			tickSpacingClamped = tickSpacingClamped[len(tickSpacingClamped)-2:]
		}
		tickSpacing := new(big.Int).SetBytes(tickSpacingClamped)

		newFeeClamped := newFeeRaw
		if len(newFeeClamped) > 3 {
			newFeeClamped = newFeeClamped[len(newFeeClamped)-3:]
		}
		newFee := new(big.Int).SetBytes(newFeeClamped)

		key := NewPoolKey(currency0, currency1, fee, tickSpacing, hooks)

		data, err := EncodeSetProtocolFee(key, newFee)
		require.NoError(t, err)

		vals, err := setProtocolFeeFn.DecodeCall(data)
		require.NoError(t, err)
		keyVals, ok := vals[0].([]any)
		require.True(t, ok)
		require.True(t, keyVals[0].(*types.Address).Equal(currency0))
		require.Zero(t, newFee.Cmp(vals[1].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeSetProtocolFeeControllerRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeSetProtocolFeeControllerRoundTrip(f *testing.F) {
	f.Add([]byte{0x00})
	f.Add(make([]byte, 20))

	f.Fuzz(func(t *testing.T, controllerRaw []byte) {
		controller := types.NewAddressFromBytes(controllerRaw)

		data, err := EncodeSetProtocolFeeController(controller)
		require.NoError(t, err)

		vals, err := setProtocolFeeControllerFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(controller))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeCollectProtocolFeesRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeCollectProtocolFeesRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x11}, []byte{0x03, 0xe8})
	f.Add(make([]byte, 20), make([]byte, 20), []byte{})

	f.Fuzz(func(t *testing.T, recipientRaw, currencyRaw, amountRaw []byte) {
		recipient := types.NewAddressFromBytes(recipientRaw)
		currency := types.NewAddressFromBytes(currencyRaw)

		amountClamped := amountRaw
		if len(amountClamped) > 32 {
			amountClamped = amountClamped[len(amountClamped)-32:]
		}
		amount := new(big.Int).SetBytes(amountClamped)

		data, err := EncodeCollectProtocolFees(recipient, currency, amount)
		require.NoError(t, err)

		vals, err := collectProtocolFeesFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(recipient))
		require.True(t, vals[1].(*types.Address).Equal(currency))
		require.Zero(t, amount.Cmp(vals[2].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeCollectProtocolFeesRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzDecodeCollectProtocolFeesRoundTrip(f *testing.F) {
	f.Add([]byte{0x03, 0xe7})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, raw []byte) {
		clamped := raw
		if len(clamped) > 32 {
			clamped = clamped[len(clamped)-32:]
		}
		amountCollected := new(big.Int).SetBytes(clamped)

		data, err := abi.Pack(abi.Types{abi.Uint256}, amountCollected)
		require.NoError(t, err)

		got, err := DecodeCollectProtocolFees(data)
		require.NoError(t, err)
		require.Zero(t, amountCollected.Cmp(got))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeProtocolFeeControllerRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzDecodeProtocolFeeControllerRoundTrip(f *testing.F) {
	f.Add([]byte{0x00})
	f.Add(make([]byte, 20))

	f.Fuzz(func(t *testing.T, controllerRaw []byte) {
		controller := types.NewAddressFromBytes(controllerRaw)

		data, err := abi.Pack(abi.Types{abi.Address}, controller)
		require.NoError(t, err)

		got, err := DecodeProtocolFeeController(data)
		require.NoError(t, err)
		require.True(t, got.Equal(controller))
	})
}

// go test -run '^$' -fuzz '^FuzzAllExtsloadExttloadDecodersDoNotPanic$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzAllExtsloadExttloadDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x01})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 64))
	f.Add(make([]byte, 96))

	f.Fuzz(func(t *testing.T, data []byte) {
		require.NotPanics(t, func() {
			_, _ = DecodeExtsload(data)
			_, _ = DecodeExtsloadRange(data)
			_, _ = DecodeExtsloadSlots(data)
			_, _ = DecodeExttload(data)
			_, _ = DecodeExttloadSlots(data)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeExtsloadRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeExtsloadRoundTrip(f *testing.F) {
	f.Add([]byte{0x01, 0x02})
	f.Add(make([]byte, 32))
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, slotRaw []byte) {
		var slot [32]byte
		clamped := slotRaw
		if len(clamped) > 32 {
			clamped = clamped[len(clamped)-32:]
		}
		copy(slot[32-len(clamped):], clamped)

		data, err := EncodeExtsload(slot)
		require.NoError(t, err)

		vals, err := extsloadFn.DecodeCall(data)
		require.NoError(t, err)
		require.Equal(t, slot[:], vals[0].([]byte))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeExtsloadRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzDecodeExtsloadRoundTrip(f *testing.F) {
	f.Add([]byte{0x2a})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, raw []byte) {
		var value [32]byte
		clamped := raw
		if len(clamped) > 32 {
			clamped = clamped[len(clamped)-32:]
		}
		copy(value[32-len(clamped):], clamped)

		data, err := abi.Pack(abi.Types{abi.Bytes32}, value[:])
		require.NoError(t, err)

		got, err := DecodeExtsload(data)
		require.NoError(t, err)
		require.Equal(t, value, got)
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeExtsloadRangeRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeExtsloadRangeRoundTrip(f *testing.F) {
	f.Add([]byte{0x03}, []byte{0x05})
	f.Add([]byte{}, []byte{})

	f.Fuzz(func(t *testing.T, startSlotRaw, nSlotsRaw []byte) {
		var startSlot [32]byte
		slotClamped := startSlotRaw
		if len(slotClamped) > 32 {
			slotClamped = slotClamped[len(slotClamped)-32:]
		}
		copy(startSlot[32-len(slotClamped):], slotClamped)

		nSlotsClamped := nSlotsRaw
		if len(nSlotsClamped) > 32 {
			nSlotsClamped = nSlotsClamped[len(nSlotsClamped)-32:]
		}
		nSlots := new(big.Int).SetBytes(nSlotsClamped)

		data, err := EncodeExtsloadRange(startSlot, nSlots)
		require.NoError(t, err)

		vals, err := extsloadRangeFn.DecodeCall(data)
		require.NoError(t, err)
		require.Equal(t, startSlot[:], vals[0].([]byte))
		require.Zero(t, nSlots.Cmp(vals[1].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeExtsloadRangeRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzDecodeExtsloadRangeRoundTrip(f *testing.F) {
	f.Add(uint8(0), []byte{0x01})
	f.Add(uint8(3), []byte{0xff})
	f.Add(uint8(32), []byte{})

	f.Fuzz(func(t *testing.T, n uint8, seed []byte) {
		count := int(n % 24)
		want := make([][32]byte, count)
		raw := make([][]byte, count)
		for i := range want {
			h := append(append([]byte{}, seed...), byte(i))
			copy(want[i][32-min(len(h), 32):], h[max(0, len(h)-32):])
			raw[i] = want[i][:]
		}

		data, err := abi.Pack(abi.Types{abi.Slice(abi.Bytes32)}, raw)
		require.NoError(t, err)

		got, err := DecodeExtsloadRange(data)
		require.NoError(t, err)
		require.Equal(t, want, got)
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeExtsloadSlotsRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeExtsloadSlotsRoundTrip(f *testing.F) {
	f.Add(uint8(0), []byte{0x01})
	f.Add(uint8(3), []byte{0xff})

	f.Fuzz(func(t *testing.T, n uint8, seed []byte) {
		count := int(n % 24)
		slots := make([][32]byte, count)
		for i := range slots {
			h := append(append([]byte{}, seed...), byte(i))
			copy(slots[i][32-min(len(h), 32):], h[max(0, len(h)-32):])
		}

		data, err := EncodeExtsloadSlots(slots)
		require.NoError(t, err)

		vals, err := extsloadSlotsFn.DecodeCall(data)
		require.NoError(t, err)
		gotSlots, ok := vals[0].([]any)
		require.True(t, ok)
		require.Len(t, gotSlots, count)
		for i := range slots {
			require.Equal(t, slots[i][:], gotSlots[i].([]byte))
		}
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeExtsloadSlotsRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzDecodeExtsloadSlotsRoundTrip(f *testing.F) {
	f.Add(uint8(0), []byte{0x01})
	f.Add(uint8(3), []byte{0xff})

	f.Fuzz(func(t *testing.T, n uint8, seed []byte) {
		count := int(n % 24)
		want := make([][32]byte, count)
		raw := make([][]byte, count)
		for i := range want {
			h := append(append([]byte{}, seed...), byte(i))
			copy(want[i][32-min(len(h), 32):], h[max(0, len(h)-32):])
			raw[i] = want[i][:]
		}

		data, err := abi.Pack(abi.Types{abi.Slice(abi.Bytes32)}, raw)
		require.NoError(t, err)

		got, err := DecodeExtsloadSlots(data)
		require.NoError(t, err)
		require.Equal(t, want, got)
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeExttloadRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeExttloadRoundTrip(f *testing.F) {
	f.Add([]byte{0x04})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, slotRaw []byte) {
		var slot [32]byte
		clamped := slotRaw
		if len(clamped) > 32 {
			clamped = clamped[len(clamped)-32:]
		}
		copy(slot[32-len(clamped):], clamped)

		data, err := EncodeExttload(slot)
		require.NoError(t, err)

		vals, err := exttloadFn.DecodeCall(data)
		require.NoError(t, err)
		require.Equal(t, slot[:], vals[0].([]byte))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeExttloadRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzDecodeExttloadRoundTrip(f *testing.F) {
	f.Add([]byte{0x2b})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, raw []byte) {
		var value [32]byte
		clamped := raw
		if len(clamped) > 32 {
			clamped = clamped[len(clamped)-32:]
		}
		copy(value[32-len(clamped):], clamped)

		data, err := abi.Pack(abi.Types{abi.Bytes32}, value[:])
		require.NoError(t, err)

		got, err := DecodeExttload(data)
		require.NoError(t, err)
		require.Equal(t, value, got)
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeExttloadSlotsRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeExttloadSlotsRoundTrip(f *testing.F) {
	f.Add(uint8(0), []byte{0x05})
	f.Add(uint8(3), []byte{0xee})

	f.Fuzz(func(t *testing.T, n uint8, seed []byte) {
		count := int(n % 24)
		slots := make([][32]byte, count)
		for i := range slots {
			h := append(append([]byte{}, seed...), byte(i))
			copy(slots[i][32-min(len(h), 32):], h[max(0, len(h)-32):])
		}

		data, err := EncodeExttloadSlots(slots)
		require.NoError(t, err)

		vals, err := exttloadSlotsFn.DecodeCall(data)
		require.NoError(t, err)
		gotSlots, ok := vals[0].([]any)
		require.True(t, ok)
		require.Len(t, gotSlots, count)
		for i := range slots {
			require.Equal(t, slots[i][:], gotSlots[i].([]byte))
		}
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeExttloadSlotsRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzDecodeExttloadSlotsRoundTrip(f *testing.F) {
	f.Add(uint8(0), []byte{0x05})
	f.Add(uint8(3), []byte{0xee})

	f.Fuzz(func(t *testing.T, n uint8, seed []byte) {
		count := int(n % 24)
		want := make([][32]byte, count)
		raw := make([][]byte, count)
		for i := range want {
			h := append(append([]byte{}, seed...), byte(i))
			copy(want[i][32-min(len(h), 32):], h[max(0, len(h)-32):])
			raw[i] = want[i][:]
		}

		data, err := abi.Pack(abi.Types{abi.Slice(abi.Bytes32)}, raw)
		require.NoError(t, err)

		got, err := DecodeExttloadSlots(data)
		require.NoError(t, err)
		require.Equal(t, want, got)
	})
}

// go test -run '^$' -fuzz '^FuzzPoolManagerInitializeThroughClient$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzPoolManagerInitializeThroughClient(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), []byte{0x0b, 0xb8}, []byte{0x3c}, false, []byte{0x01}, uint64(0), true)
	f.Add([]byte{0xde, 0xad}, []byte{0xbe, 0xef}, []byte{0xff, 0xff, 0xff}, []byte{0x80, 0x00, 0x00}, true, make([]byte, 20), uint64(1<<40), false)

	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(f, err)
	signer := core.PubkeyToAddress(key.PublicKey())

	f.Fuzz(func(t *testing.T, c0Raw, c1Raw, feeRaw, tickSpacingRaw []byte, negativeSpacing bool, sqrtRaw []byte, nonce uint64, dynamic bool) {
		// fold raw bytes into the uint24/int24/uint160 ranges the ABI accepts
		fee := new(big.Int).Mod(new(big.Int).SetBytes(feeRaw), new(big.Int).Lsh(big.NewInt(1), 24))
		tickSpacing := new(big.Int).Mod(new(big.Int).SetBytes(tickSpacingRaw), new(big.Int).Lsh(big.NewInt(1), 23))
		if negativeSpacing {
			tickSpacing.Neg(tickSpacing)
		}
		sqrtPriceX96 := new(big.Int).Mod(new(big.Int).SetBytes(sqrtRaw), new(big.Int).Lsh(big.NewInt(1), 160))
		poolKey := NewPoolKey(types.NewAddressFromBytes(c0Raw), types.NewAddressFromBytes(c1Raw), fee, tickSpacing, testHooks)

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
			var err error
			raw, err = encoding.Hex.Decode(e.Params.([]any)[0].(string))
			require.NoError(t, err)
			*e.Result.(*string) = core.Keccak256(raw).String()
			return nil
		}).Once()

		signed, hash, err := New(c).Initialize(t.Context(), bind.NewTransactOptions(key, builder), NewInitializeRequest(testPoolManager, poolKey, sqrtPriceX96))
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
		require.True(t, gotTo.Equal(testPoolManager))

		vals, err := initializeFn.DecodeCall(gotData)
		require.NoError(t, err)
		keyVals := vals[0].([]any)
		require.True(t, poolKey.Currency0.Equal(keyVals[0].(*types.Address)))
		require.True(t, poolKey.Currency1.Equal(keyVals[1].(*types.Address)))
		require.Zero(t, fee.Cmp(keyVals[2].(*big.Int)))
		require.Zero(t, tickSpacing.Cmp(keyVals[3].(*big.Int)))
		require.True(t, testHooks.Equal(keyVals[4].(*types.Address)))
		require.Zero(t, sqrtPriceX96.Cmp(vals[1].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzPoolManagerSettleCarriesNativeAmount$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzPoolManagerSettleCarriesNativeAmount(f *testing.F) {
	f.Add([]byte{}, false)
	f.Add([]byte{0x01}, true)
	f.Add([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, true)

	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(f, err)

	f.Fuzz(func(t *testing.T, amountRaw []byte, forRecipient bool) {
		var native *big.Int
		if len(amountRaw) > 0 {
			// fold amountRaw into [0, 2^256), the range a tx value can carry
			native = new(big.Int).Mod(new(big.Int).SetBytes(amountRaw), new(big.Int).Lsh(big.NewInt(1), 256))
		}

		builder := tx.NewDynamicFeeBuilder()
		builder.ChainID(big.NewInt(1))
		builder.Nonce(3)
		builder.Limit(100_000)
		builder.Price(big.NewInt(2), big.NewInt(1))

		var raw []byte
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			var err error
			raw, err = encoding.Hex.Decode(e.Params.([]any)[0].(string))
			require.NoError(t, err)
			*e.Result.(*string) = core.Keccak256(raw).String()
			return nil
		}).Once()

		opts := bind.NewTransactOptions(key, builder)
		var want []byte
		if forRecipient {
			want, err = EncodeSettleFor(testSender)
			require.NoError(t, err)
			_, _, err = New(c).SettleFor(t.Context(), opts, NewSettleForRequest(testPoolManager, testSender, native))
		} else {
			want, err = EncodeSettle()
			require.NoError(t, err)
			_, _, err = New(c).Settle(t.Context(), opts, NewSettleRequest(testPoolManager, native))
		}
		require.NoError(t, err)
		require.Nil(t, opts.Value)

		decoded, err := core.DecodeTransaction(raw)
		require.NoError(t, err)
		dyn := decoded.(*core.DynamicFeeTx)
		require.Equal(t, want, dyn.Data)
		if native == nil {
			require.Zero(t, dyn.Value.Sign())
		} else {
			require.Zero(t, native.Cmp(dyn.Value))
		}
	})
}

// go test -run '^$' -fuzz '^FuzzPoolManagerReadsSurviveArbitraryResult$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzPoolManagerReadsSurviveArbitraryResult(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 64))
	f.Add(append(append(make([]byte, 31), 0x20), append(make([]byte, 31), 0x02)...))
	f.Add(append(append(make([]byte, 31), 0x20), append(make([]byte, 24), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff)...))

	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	liq := NewLiquidityParams(big.NewInt(-60), big.NewInt(60), big.NewInt(1), [32]byte{})

	f.Fuzz(func(t *testing.T, result []byte) {
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(result)
			return nil
		}).Times(10)
		m := New(c)
		ctx := t.Context()

		accrued, err := m.ProtocolFeesAccrued(ctx, nil, NewProtocolFeesAccruedRequest(testPoolManager, testCurrency0))
		require.True(t, (err == nil) != (accrued == nil))
		controller, err := m.ProtocolFeeController(ctx, nil, NewProtocolFeeControllerRequest(testPoolManager))
		require.True(t, (err == nil) != (controller == nil))
		word, err := m.Extsload(ctx, nil, NewExtsloadRequest(testPoolManager, [32]byte{}))
		if err != nil {
			require.Equal(t, [32]byte{}, word)
		}
		words, err := m.ExtsloadRange(ctx, nil, NewExtsloadRangeRequest(testPoolManager, [32]byte{}, big.NewInt(2)))
		if err != nil {
			require.Nil(t, words)
		}
		words, err = m.ExttloadSlots(ctx, nil, NewExttloadSlotsRequest(testPoolManager, nil))
		if err != nil {
			require.Nil(t, words)
		}
		tick, err := m.CallInitialize(ctx, nil, NewInitializeRequest(testPoolManager, key, big.NewInt(1)))
		require.True(t, (err == nil) != (tick == nil))
		caller, fees, err := m.CallModifyLiquidity(ctx, nil, NewModifyLiquidityRequest(testPoolManager, key, liq, nil))
		require.True(t, (err == nil) != (caller == nil))
		require.True(t, (err == nil) != (fees == nil))
		if err == nil {
			for _, d := range []*BalanceDelta{caller, fees} {
				require.LessOrEqual(t, d.Amount0.BitLen(), 128)
				require.LessOrEqual(t, d.Amount1.BitLen(), 128)
			}
		}
		delta, err := m.CallSwap(ctx, nil, NewSwapRequest(testPoolManager, key, NewSwapAmountParams(true, big.NewInt(-1), big.NewInt(1)), nil))
		require.True(t, (err == nil) != (delta == nil))
		out, err := m.CallUnlock(ctx, nil, NewUnlockRequest(testPoolManager, nil))
		if err != nil {
			require.Nil(t, out)
		}
		paid, err := m.CallSettle(ctx, nil, NewSettleRequest(testPoolManager, nil))
		require.True(t, (err == nil) != (paid == nil))
	})
}

// go test -run '^$' -fuzz '^FuzzPoolManagerWrapRevertErrorKeepsCause$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzPoolManagerWrapRevertErrorKeepsCause(f *testing.F) {
	for _, e := range []*abi.Error{errManagerLocked, errPoolAlreadyInitialized, errWrappedError, errTickMisaligned} {
		sel := e.Selector()
		f.Add(sel[:])
	}
	f.Add([]byte{})
	f.Add([]byte{0x08, 0xc3, 0x79, 0xa0})

	f.Fuzz(func(t *testing.T, data []byte) {
		rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(data) + `"`)}

		var wrapped error
		require.NotPanics(t, func() { wrapped = mapPoolManagerRevertError(rpcErr) })
		got, ok := errors.AsType[*rpc.ResponseError](wrapped)
		require.True(t, ok)
		require.Same(t, rpcErr, got)
		if len(data) < 4 {
			require.Same(t, rpcErr, wrapped)
		} else {
			require.NotSame(t, rpcErr, wrapped)
		}
	})
}
