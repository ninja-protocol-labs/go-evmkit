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
	"github.com/ninja-protocol-labs/go-evmkit/protocols/permit2"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// go test -run '^$' -fuzz '^FuzzExtractModifyPositionDoesNotPanic$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzExtractModifyPositionDoesNotPanic(f *testing.F) {
	t0, ok := modifyPositionEvent.Topic0()
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
			logs := []tx.Log{{Address: testPositionManager, Topics: topics, Data: data}}
			_, _ = ExtractModifyPosition(logs)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzExtractSubscriptionDoesNotPanic$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzExtractSubscriptionDoesNotPanic(f *testing.F) {
	t0, ok := subscriptionEvent.Topic0()
	require.True(f, ok)

	f.Add(t0[:], make([]byte, 32), make([]byte, 32), []byte{})
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
			logs := []tx.Log{{Address: testPositionManager, Topics: topics, Data: data}}
			_, _ = ExtractSubscription(logs)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzExtractUnsubscriptionDoesNotPanic$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzExtractUnsubscriptionDoesNotPanic(f *testing.F) {
	t0, ok := unsubscriptionEvent.Topic0()
	require.True(f, ok)

	f.Add(t0[:], make([]byte, 32), make([]byte, 32), []byte{})
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
			logs := []tx.Log{{Address: testPositionManager, Topics: topics, Data: data}}
			_, _ = ExtractUnsubscription(logs)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzExtractModifyPositionRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzExtractModifyPositionRoundTrip(f *testing.F) {
	f.Add([]byte{0x03}, []byte{0x00}, []byte{0xff, 0x38}, []byte{0x00, 0xc8}, []byte{0x01, 0xf4}, []byte{0x0b})
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

		t0, ok := modifyPositionEvent.Topic0()
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
			Address: testPositionManager,
			Topics: []*types.Hash{
				types.NewHashFromBytes(t0[:]),
				types.NewHashFromBytes(id),
				types.NewHashFromBytes(sender.Bytes()),
			},
			Data: data,
		}}

		mods, err := ExtractModifyPosition(logs)
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

// go test -run '^$' -fuzz '^FuzzExtractSubscriptionRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzExtractSubscriptionRoundTrip(f *testing.F) {
	f.Add([]byte{0x07}, make([]byte, 20))
	f.Add(make([]byte, 32), make([]byte, 20))

	f.Fuzz(func(t *testing.T, tokenIDRaw, subscriberRaw []byte) {
		subscriber := types.NewAddressFromBytes(subscriberRaw)

		t0, ok := subscriptionEvent.Topic0()
		require.True(t, ok)

		idClamped := tokenIDRaw
		if len(idClamped) > 32 {
			idClamped = idClamped[len(idClamped)-32:]
		}
		id := make([]byte, 32)
		copy(id[32-len(idClamped):], idClamped)
		tokenID := new(big.Int).SetBytes(idClamped)

		logs := []tx.Log{{
			Address: testPositionManager,
			Topics: []*types.Hash{
				types.NewHashFromBytes(t0[:]),
				types.NewHashFromBytes(id),
				types.NewHashFromBytes(subscriber.Bytes()),
			},
		}}

		subs, err := ExtractSubscription(logs)
		require.NoError(t, err)
		require.Len(t, subs, 1)
		require.Zero(t, tokenID.Cmp(subs[0].TokenID))
		require.True(t, subs[0].Subscriber.Equal(subscriber))
	})
}

// go test -run '^$' -fuzz '^FuzzExtractUnsubscriptionRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzExtractUnsubscriptionRoundTrip(f *testing.F) {
	f.Add([]byte{0x09}, make([]byte, 20))
	f.Add(make([]byte, 32), make([]byte, 20))

	f.Fuzz(func(t *testing.T, tokenIDRaw, subscriberRaw []byte) {
		subscriber := types.NewAddressFromBytes(subscriberRaw)

		t0, ok := unsubscriptionEvent.Topic0()
		require.True(t, ok)

		idClamped := tokenIDRaw
		if len(idClamped) > 32 {
			idClamped = idClamped[len(idClamped)-32:]
		}
		id := make([]byte, 32)
		copy(id[32-len(idClamped):], idClamped)
		tokenID := new(big.Int).SetBytes(idClamped)

		logs := []tx.Log{{
			Address: testPositionManager,
			Topics: []*types.Hash{
				types.NewHashFromBytes(t0[:]),
				types.NewHashFromBytes(id),
				types.NewHashFromBytes(subscriber.Bytes()),
			},
		}}

		unsubs, err := ExtractUnsubscription(logs)
		require.NoError(t, err)
		require.Len(t, unsubs, 1)
		require.Zero(t, tokenID.Cmp(unsubs[0].TokenID))
		require.True(t, unsubs[0].Subscriber.Equal(subscriber))
	})
}

// go test -run '^$' -fuzz '^FuzzAllPositionDecodersDoNotPanic$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzAllPositionDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x01})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 33))
	f.Add(make([]byte, 64))
	f.Add(make([]byte, 96))
	f.Add(make([]byte, 128))
	f.Add(make([]byte, 224))

	f.Fuzz(func(t *testing.T, data []byte) {
		require.NotPanics(t, func() {
			_, _ = DecodeActionMintPosition(data)
			_, _ = DecodeActionMintPositionFromDeltas(data)
			_, _, _, _, _, _ = DecodeActionIncreaseLiquidity(data)
			_, _, _, _, _ = DecodeActionIncreaseLiquidityFromDeltas(data)
			_, _, _, _, _, _ = DecodeActionDecreaseLiquidity(data)
			_, _, _, _, _ = DecodeActionBurnPosition(data)
			_, _, _ = DecodeActionSettlePair(data)
			_, _, _, _ = DecodeActionTakePair(data)
			_, _ = DecodeActionCloseCurrency(data)
			_, _, _ = DecodeActionClearOrTake(data)
			_, _, _ = DecodeActionSweep(data)
			_, _ = DecodeActionWrap(data)
			_, _ = DecodeActionUnwrap(data)
			_, _ = DecodeNextTokenID(data)
			_, _ = DecodeGetPositionLiquidity(data)
			_, _, _ = DecodeGetPoolAndPositionInfo(data)
			_, _ = DecodePositionInfo(data)
			_, _ = DecodePoolKeys(data)
			_, _ = DecodeWETH9(data)
			_, _ = DecodePermit2Contract(data)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzEncodePoolKeysRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodePoolKeysRoundTrip(f *testing.F) {
	f.Add([]byte{0x11, 0x22, 0x33})
	f.Add(make([]byte, 25))

	f.Fuzz(func(t *testing.T, poolIDRaw []byte) {
		var poolID [25]byte
		poolIDClamped := poolIDRaw
		if len(poolIDClamped) > 25 {
			poolIDClamped = poolIDClamped[:25]
		}
		copy(poolID[:], poolIDClamped)

		calldata, err := EncodePoolKeys(poolID)
		require.NoError(t, err)

		vals, err := poolKeysFn.DecodeCall(calldata)
		require.NoError(t, err)
		decodedPoolID, ok := vals[0].([]byte)
		require.True(t, ok)
		require.Equal(t, poolID[:], decodedPoolID)
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeGetPositionLiquidityRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeGetPositionLiquidityRoundTrip(f *testing.F) {
	f.Add([]byte{0x2a})
	f.Add(make([]byte, 32))

	f.Fuzz(func(t *testing.T, tokenIDRaw []byte) {
		tokenIDClamped := tokenIDRaw
		if len(tokenIDClamped) > 32 {
			tokenIDClamped = tokenIDClamped[len(tokenIDClamped)-32:]
		}
		tokenID := new(big.Int).SetBytes(tokenIDClamped)

		data, err := EncodeGetPositionLiquidity(tokenID)
		require.NoError(t, err)

		vals, err := getPositionLiquidityFn.DecodeCall(data)
		require.NoError(t, err)
		decodedTokenID, ok := vals[0].(*big.Int)
		require.True(t, ok)
		require.Zero(t, tokenID.Cmp(decodedTokenID))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodePositionInfoRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzDecodePositionInfoRoundTrip(f *testing.F) {
	f.Add([]byte{0xaa, 0xbb, 0xcc}, []byte{0x03, 0x84}, []byte{0xfc, 0x6c}, true)
	f.Add(make([]byte, 25), []byte{}, []byte{}, false)

	f.Fuzz(func(t *testing.T, poolIDRaw, tickUpperRaw, tickLowerRaw []byte, hasSubscriber bool) {
		var poolID [25]byte
		poolIDClamped := poolIDRaw
		if len(poolIDClamped) > 25 {
			poolIDClamped = poolIDClamped[:25]
		}
		copy(poolID[:], poolIDClamped)

		var tickUpperBytes [2]byte
		tickUpperClamped := tickUpperRaw
		if len(tickUpperClamped) > 2 {
			tickUpperClamped = tickUpperClamped[len(tickUpperClamped)-2:]
		}
		copy(tickUpperBytes[2-len(tickUpperClamped):], tickUpperClamped)

		var tickLowerBytes [2]byte
		tickLowerClamped := tickLowerRaw
		if len(tickLowerClamped) > 2 {
			tickLowerClamped = tickLowerClamped[len(tickLowerClamped)-2:]
		}
		copy(tickLowerBytes[2-len(tickLowerClamped):], tickLowerClamped)

		word := make([]byte, 32)
		copy(word[0:25], poolID[:])
		word[27] = tickUpperBytes[1]
		word[26] = tickUpperBytes[0]
		word[30] = tickLowerBytes[1]
		word[29] = tickLowerBytes[0]
		if hasSubscriber {
			word[31] = 0x01
		}
		raw := new(big.Int).SetBytes(word)

		wantTickUpper := new(big.Int).SetBytes(word[25:28])
		if word[25]&0x80 != 0 {
			wantTickUpper.Sub(wantTickUpper, new(big.Int).Lsh(big.NewInt(1), 24))
		}
		wantTickLower := new(big.Int).SetBytes(word[28:31])
		if word[28]&0x80 != 0 {
			wantTickLower.Sub(wantTickLower, new(big.Int).Lsh(big.NewInt(1), 24))
		}

		returnData, err := abi.Pack(abi.Types{abi.Uint256}, raw)
		require.NoError(t, err)

		info, err := DecodePositionInfo(returnData)
		require.NoError(t, err)
		require.Equal(t, poolID, info.PoolID)
		require.Zero(t, wantTickUpper.Cmp(info.TickUpper))
		require.Zero(t, wantTickLower.Cmp(info.TickLower))
		require.Equal(t, hasSubscriber, info.HasSubscriber)
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeGetPoolAndPositionInfoRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzDecodeGetPoolAndPositionInfoRoundTrip(f *testing.F) {
	f.Add([]byte{0x11}, []byte{0x22}, []byte{0x33}, []byte{0x0b, 0xb8}, []byte{0x00, 0x3c}, []byte{0xaa, 0xbb, 0xcc}, []byte{0x03, 0x84}, []byte{0xfc, 0x6c}, true)
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 20), []byte{}, []byte{}, make([]byte, 25), []byte{}, []byte{}, false)

	f.Fuzz(func(t *testing.T, currency0Raw, currency1Raw, hooksRaw, feeRaw, tickSpacingRaw, poolIDRaw, tickUpperRaw, tickLowerRaw []byte, hasSubscriber bool) {
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

		var poolID [25]byte
		poolIDClamped := poolIDRaw
		if len(poolIDClamped) > 25 {
			poolIDClamped = poolIDClamped[:25]
		}
		copy(poolID[:], poolIDClamped)

		var tickUpperBytes [2]byte
		tickUpperClamped := tickUpperRaw
		if len(tickUpperClamped) > 2 {
			tickUpperClamped = tickUpperClamped[len(tickUpperClamped)-2:]
		}
		copy(tickUpperBytes[2-len(tickUpperClamped):], tickUpperClamped)

		var tickLowerBytes [2]byte
		tickLowerClamped := tickLowerRaw
		if len(tickLowerClamped) > 2 {
			tickLowerClamped = tickLowerClamped[len(tickLowerClamped)-2:]
		}
		copy(tickLowerBytes[2-len(tickLowerClamped):], tickLowerClamped)

		word := make([]byte, 32)
		copy(word[0:25], poolID[:])
		word[27] = tickUpperBytes[1]
		word[26] = tickUpperBytes[0]
		word[30] = tickLowerBytes[1]
		word[29] = tickLowerBytes[0]
		if hasSubscriber {
			word[31] = 0x01
		}
		raw := new(big.Int).SetBytes(word)

		wantTickUpper := new(big.Int).SetBytes(word[25:28])
		if word[25]&0x80 != 0 {
			wantTickUpper.Sub(wantTickUpper, new(big.Int).Lsh(big.NewInt(1), 24))
		}
		wantTickLower := new(big.Int).SetBytes(word[28:31])
		if word[28]&0x80 != 0 {
			wantTickLower.Sub(wantTickLower, new(big.Int).Lsh(big.NewInt(1), 24))
		}

		returnData, err := abi.Pack(abi.Types{poolKeyTuple, abi.Uint256},
			[]any{currency0, currency1, fee, tickSpacing, hooks}, raw)
		require.NoError(t, err)

		poolKey, info, err := DecodeGetPoolAndPositionInfo(returnData)
		require.NoError(t, err)
		require.True(t, poolKey.Currency0.Equal(currency0))
		require.True(t, poolKey.Currency1.Equal(currency1))
		require.Zero(t, fee.Cmp(poolKey.Fee))
		require.Zero(t, tickSpacing.Cmp(poolKey.TickSpacing))
		require.True(t, poolKey.Hooks.Equal(hooks))
		require.Equal(t, poolID, info.PoolID)
		require.Zero(t, wantTickUpper.Cmp(info.TickUpper))
		require.Zero(t, wantTickLower.Cmp(info.TickLower))
		require.Equal(t, hasSubscriber, info.HasSubscriber)
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeActionMintPositionRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeActionMintPositionRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x11}, []byte{0x22}, []byte{0x0b, 0xb8}, []byte{0x00, 0x3c}, []byte{0xff, 0x38}, []byte{0x00, 0xc8}, []byte{0x0f, 0x42, 0x40}, []byte{0x03, 0xe8}, []byte{0x04, 0x4c}, []byte{0x33}, []byte{0xde, 0xad})
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 20), []byte{}, []byte{}, []byte{}, []byte{}, []byte{}, []byte{}, []byte{}, make([]byte, 20), []byte{})

	f.Fuzz(func(t *testing.T, currency0Raw, currency1Raw, hooksRaw, feeRaw, tickSpacingRaw, tickLowerRaw, tickUpperRaw, liquidityRaw, amount0MaxRaw, amount1MaxRaw, ownerRaw, hookData []byte) {
		currency0 := types.NewAddressFromBytes(currency0Raw)
		currency1 := types.NewAddressFromBytes(currency1Raw)
		hooks := types.NewAddressFromBytes(hooksRaw)
		owner := types.NewAddressFromBytes(ownerRaw)

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

		liquidityClamped := liquidityRaw
		if len(liquidityClamped) > 32 {
			liquidityClamped = liquidityClamped[len(liquidityClamped)-32:]
		}
		liquidity := new(big.Int).SetBytes(liquidityClamped)

		amount0MaxClamped := amount0MaxRaw
		if len(amount0MaxClamped) > 16 {
			amount0MaxClamped = amount0MaxClamped[len(amount0MaxClamped)-16:]
		}
		amount0Max := new(big.Int).SetBytes(amount0MaxClamped)

		amount1MaxClamped := amount1MaxRaw
		if len(amount1MaxClamped) > 16 {
			amount1MaxClamped = amount1MaxClamped[len(amount1MaxClamped)-16:]
		}
		amount1Max := new(big.Int).SetBytes(amount1MaxClamped)

		poolKey := NewPoolKey(currency0, currency1, fee, tickSpacing, hooks)
		params := NewMintPositionParams(poolKey, tickLower, tickUpper, liquidity, amount0Max, amount1Max, owner, hookData)

		data, err := EncodeActionMintPosition(params)
		require.NoError(t, err)

		decoded, err := DecodeActionMintPosition(data)
		require.NoError(t, err)
		require.True(t, decoded.PoolKey.Currency0.Equal(currency0))
		require.True(t, decoded.PoolKey.Currency1.Equal(currency1))
		require.Zero(t, fee.Cmp(decoded.PoolKey.Fee))
		require.Zero(t, tickSpacing.Cmp(decoded.PoolKey.TickSpacing))
		require.True(t, decoded.PoolKey.Hooks.Equal(hooks))
		require.Zero(t, tickLower.Cmp(decoded.TickLower))
		require.Zero(t, tickUpper.Cmp(decoded.TickUpper))
		require.Zero(t, liquidity.Cmp(decoded.Liquidity))
		require.Zero(t, amount0Max.Cmp(decoded.Amount0Max))
		require.Zero(t, amount1Max.Cmp(decoded.Amount1Max))
		require.True(t, decoded.Owner.Equal(owner))
		require.Equal(t, hookData, decoded.HookData)
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeActionMintPositionFromDeltasRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeActionMintPositionFromDeltasRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x11}, []byte{0x22}, []byte{0x0b, 0xb8}, []byte{0x00, 0x3c}, []byte{0xff, 0x38}, []byte{0x00, 0xc8}, []byte{0x03, 0xe8}, []byte{0x04, 0x4c}, []byte{0x33}, []byte{0xbe, 0xef})
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 20), []byte{}, []byte{}, []byte{}, []byte{}, []byte{}, []byte{}, make([]byte, 20), []byte{})

	f.Fuzz(func(t *testing.T, currency0Raw, currency1Raw, hooksRaw, feeRaw, tickSpacingRaw, tickLowerRaw, tickUpperRaw, amount0MaxRaw, amount1MaxRaw, ownerRaw, hookData []byte) {
		currency0 := types.NewAddressFromBytes(currency0Raw)
		currency1 := types.NewAddressFromBytes(currency1Raw)
		hooks := types.NewAddressFromBytes(hooksRaw)
		owner := types.NewAddressFromBytes(ownerRaw)

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

		amount0MaxClamped := amount0MaxRaw
		if len(amount0MaxClamped) > 16 {
			amount0MaxClamped = amount0MaxClamped[len(amount0MaxClamped)-16:]
		}
		amount0Max := new(big.Int).SetBytes(amount0MaxClamped)

		amount1MaxClamped := amount1MaxRaw
		if len(amount1MaxClamped) > 16 {
			amount1MaxClamped = amount1MaxClamped[len(amount1MaxClamped)-16:]
		}
		amount1Max := new(big.Int).SetBytes(amount1MaxClamped)

		poolKey := NewPoolKey(currency0, currency1, fee, tickSpacing, hooks)
		params := NewMintPositionFromDeltasParams(poolKey, tickLower, tickUpper, amount0Max, amount1Max, owner, hookData)

		data, err := EncodeActionMintPositionFromDeltas(params)
		require.NoError(t, err)

		decoded, err := DecodeActionMintPositionFromDeltas(data)
		require.NoError(t, err)
		require.True(t, decoded.PoolKey.Currency0.Equal(currency0))
		require.True(t, decoded.PoolKey.Currency1.Equal(currency1))
		require.Zero(t, fee.Cmp(decoded.PoolKey.Fee))
		require.Zero(t, tickSpacing.Cmp(decoded.PoolKey.TickSpacing))
		require.True(t, decoded.PoolKey.Hooks.Equal(hooks))
		require.Zero(t, tickLower.Cmp(decoded.TickLower))
		require.Zero(t, tickUpper.Cmp(decoded.TickUpper))
		require.Zero(t, amount0Max.Cmp(decoded.Amount0Max))
		require.Zero(t, amount1Max.Cmp(decoded.Amount1Max))
		require.True(t, decoded.Owner.Equal(owner))
		require.Equal(t, hookData, decoded.HookData)
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeActionIncreaseLiquidityRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeActionIncreaseLiquidityRoundTrip(f *testing.F) {
	f.Add([]byte{0x2a}, []byte{0x0f, 0x42, 0x40}, []byte{0x03, 0xe8}, []byte{0x04, 0x4c}, []byte{0x01})
	f.Add(make([]byte, 32), make([]byte, 32), []byte{}, []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, tokenIDRaw, liquidityRaw, amount0MaxRaw, amount1MaxRaw, hookData []byte) {
		tokenIDClamped := tokenIDRaw
		if len(tokenIDClamped) > 32 {
			tokenIDClamped = tokenIDClamped[len(tokenIDClamped)-32:]
		}
		tokenID := new(big.Int).SetBytes(tokenIDClamped)

		liquidityClamped := liquidityRaw
		if len(liquidityClamped) > 32 {
			liquidityClamped = liquidityClamped[len(liquidityClamped)-32:]
		}
		liquidity := new(big.Int).SetBytes(liquidityClamped)

		amount0MaxClamped := amount0MaxRaw
		if len(amount0MaxClamped) > 16 {
			amount0MaxClamped = amount0MaxClamped[len(amount0MaxClamped)-16:]
		}
		amount0Max := new(big.Int).SetBytes(amount0MaxClamped)

		amount1MaxClamped := amount1MaxRaw
		if len(amount1MaxClamped) > 16 {
			amount1MaxClamped = amount1MaxClamped[len(amount1MaxClamped)-16:]
		}
		amount1Max := new(big.Int).SetBytes(amount1MaxClamped)

		data, err := EncodeActionIncreaseLiquidity(tokenID, liquidity, amount0Max, amount1Max, hookData)
		require.NoError(t, err)

		decodedTokenID, decodedLiquidity, decodedAmount0Max, decodedAmount1Max, decodedHookData, err := DecodeActionIncreaseLiquidity(data)
		require.NoError(t, err)
		require.Zero(t, tokenID.Cmp(decodedTokenID))
		require.Zero(t, liquidity.Cmp(decodedLiquidity))
		require.Zero(t, amount0Max.Cmp(decodedAmount0Max))
		require.Zero(t, amount1Max.Cmp(decodedAmount1Max))
		require.Equal(t, hookData, decodedHookData)
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeActionIncreaseLiquidityFromDeltasRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeActionIncreaseLiquidityFromDeltasRoundTrip(f *testing.F) {
	f.Add([]byte{0x2a}, []byte{0x03, 0xe8}, []byte{0x04, 0x4c}, []byte{0x02})
	f.Add(make([]byte, 32), []byte{}, []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, tokenIDRaw, amount0MaxRaw, amount1MaxRaw, hookData []byte) {
		tokenIDClamped := tokenIDRaw
		if len(tokenIDClamped) > 32 {
			tokenIDClamped = tokenIDClamped[len(tokenIDClamped)-32:]
		}
		tokenID := new(big.Int).SetBytes(tokenIDClamped)

		amount0MaxClamped := amount0MaxRaw
		if len(amount0MaxClamped) > 16 {
			amount0MaxClamped = amount0MaxClamped[len(amount0MaxClamped)-16:]
		}
		amount0Max := new(big.Int).SetBytes(amount0MaxClamped)

		amount1MaxClamped := amount1MaxRaw
		if len(amount1MaxClamped) > 16 {
			amount1MaxClamped = amount1MaxClamped[len(amount1MaxClamped)-16:]
		}
		amount1Max := new(big.Int).SetBytes(amount1MaxClamped)

		data, err := EncodeActionIncreaseLiquidityFromDeltas(tokenID, amount0Max, amount1Max, hookData)
		require.NoError(t, err)

		decodedTokenID, decodedAmount0Max, decodedAmount1Max, decodedHookData, err := DecodeActionIncreaseLiquidityFromDeltas(data)
		require.NoError(t, err)
		require.Zero(t, tokenID.Cmp(decodedTokenID))
		require.Zero(t, amount0Max.Cmp(decodedAmount0Max))
		require.Zero(t, amount1Max.Cmp(decodedAmount1Max))
		require.Equal(t, hookData, decodedHookData)
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeActionDecreaseLiquidityRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeActionDecreaseLiquidityRoundTrip(f *testing.F) {
	f.Add([]byte{0x2a}, []byte{0x07, 0xa1, 0x20}, []byte{0x03, 0x84}, []byte{0x03, 0xb6}, []byte{0x03})
	f.Add(make([]byte, 32), make([]byte, 32), []byte{}, []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, tokenIDRaw, liquidityRaw, amount0MinRaw, amount1MinRaw, hookData []byte) {
		tokenIDClamped := tokenIDRaw
		if len(tokenIDClamped) > 32 {
			tokenIDClamped = tokenIDClamped[len(tokenIDClamped)-32:]
		}
		tokenID := new(big.Int).SetBytes(tokenIDClamped)

		liquidityClamped := liquidityRaw
		if len(liquidityClamped) > 32 {
			liquidityClamped = liquidityClamped[len(liquidityClamped)-32:]
		}
		liquidity := new(big.Int).SetBytes(liquidityClamped)

		amount0MinClamped := amount0MinRaw
		if len(amount0MinClamped) > 16 {
			amount0MinClamped = amount0MinClamped[len(amount0MinClamped)-16:]
		}
		amount0Min := new(big.Int).SetBytes(amount0MinClamped)

		amount1MinClamped := amount1MinRaw
		if len(amount1MinClamped) > 16 {
			amount1MinClamped = amount1MinClamped[len(amount1MinClamped)-16:]
		}
		amount1Min := new(big.Int).SetBytes(amount1MinClamped)

		data, err := EncodeActionDecreaseLiquidity(tokenID, liquidity, amount0Min, amount1Min, hookData)
		require.NoError(t, err)

		decodedTokenID, decodedLiquidity, decodedAmount0Min, decodedAmount1Min, decodedHookData, err := DecodeActionDecreaseLiquidity(data)
		require.NoError(t, err)
		require.Zero(t, tokenID.Cmp(decodedTokenID))
		require.Zero(t, liquidity.Cmp(decodedLiquidity))
		require.Zero(t, amount0Min.Cmp(decodedAmount0Min))
		require.Zero(t, amount1Min.Cmp(decodedAmount1Min))
		require.Equal(t, hookData, decodedHookData)
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeActionBurnPositionRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeActionBurnPositionRoundTrip(f *testing.F) {
	f.Add([]byte{0x2a}, []byte{0x03, 0x84}, []byte{0x03, 0xb6}, []byte{0x04})
	f.Add(make([]byte, 32), []byte{}, []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, tokenIDRaw, amount0MinRaw, amount1MinRaw, hookData []byte) {
		tokenIDClamped := tokenIDRaw
		if len(tokenIDClamped) > 32 {
			tokenIDClamped = tokenIDClamped[len(tokenIDClamped)-32:]
		}
		tokenID := new(big.Int).SetBytes(tokenIDClamped)

		amount0MinClamped := amount0MinRaw
		if len(amount0MinClamped) > 16 {
			amount0MinClamped = amount0MinClamped[len(amount0MinClamped)-16:]
		}
		amount0Min := new(big.Int).SetBytes(amount0MinClamped)

		amount1MinClamped := amount1MinRaw
		if len(amount1MinClamped) > 16 {
			amount1MinClamped = amount1MinClamped[len(amount1MinClamped)-16:]
		}
		amount1Min := new(big.Int).SetBytes(amount1MinClamped)

		data, err := EncodeActionBurnPosition(tokenID, amount0Min, amount1Min, hookData)
		require.NoError(t, err)

		decodedTokenID, decodedAmount0Min, decodedAmount1Min, decodedHookData, err := DecodeActionBurnPosition(data)
		require.NoError(t, err)
		require.Zero(t, tokenID.Cmp(decodedTokenID))
		require.Zero(t, amount0Min.Cmp(decodedAmount0Min))
		require.Zero(t, amount1Min.Cmp(decodedAmount1Min))
		require.Equal(t, hookData, decodedHookData)
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeActionSettlePairRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeActionSettlePairRoundTrip(f *testing.F) {
	f.Add([]byte{0x11}, []byte{0x22})
	f.Add(make([]byte, 20), make([]byte, 20))

	f.Fuzz(func(t *testing.T, currency0Raw, currency1Raw []byte) {
		currency0 := types.NewAddressFromBytes(currency0Raw)
		currency1 := types.NewAddressFromBytes(currency1Raw)

		data, err := EncodeActionSettlePair(currency0, currency1)
		require.NoError(t, err)

		decodedCurrency0, decodedCurrency1, err := DecodeActionSettlePair(data)
		require.NoError(t, err)
		require.True(t, decodedCurrency0.Equal(currency0))
		require.True(t, decodedCurrency1.Equal(currency1))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeActionTakePairRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeActionTakePairRoundTrip(f *testing.F) {
	f.Add([]byte{0x11}, []byte{0x22}, []byte{0x33})
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 20))

	f.Fuzz(func(t *testing.T, currency0Raw, currency1Raw, recipientRaw []byte) {
		currency0 := types.NewAddressFromBytes(currency0Raw)
		currency1 := types.NewAddressFromBytes(currency1Raw)
		recipient := types.NewAddressFromBytes(recipientRaw)

		data, err := EncodeActionTakePair(currency0, currency1, recipient)
		require.NoError(t, err)

		decodedCurrency0, decodedCurrency1, decodedRecipient, err := DecodeActionTakePair(data)
		require.NoError(t, err)
		require.True(t, decodedCurrency0.Equal(currency0))
		require.True(t, decodedCurrency1.Equal(currency1))
		require.True(t, decodedRecipient.Equal(recipient))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeActionCloseCurrencyRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeActionCloseCurrencyRoundTrip(f *testing.F) {
	f.Add([]byte{0x11})
	f.Add(make([]byte, 20))

	f.Fuzz(func(t *testing.T, currencyRaw []byte) {
		currency := types.NewAddressFromBytes(currencyRaw)

		data, err := EncodeActionCloseCurrency(currency)
		require.NoError(t, err)

		decodedCurrency, err := DecodeActionCloseCurrency(data)
		require.NoError(t, err)
		require.True(t, decodedCurrency.Equal(currency))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeActionClearOrTakeRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeActionClearOrTakeRoundTrip(f *testing.F) {
	f.Add([]byte{0x11}, []byte{0x03, 0xe8})
	f.Add(make([]byte, 20), []byte{})

	f.Fuzz(func(t *testing.T, currencyRaw, amountMaxRaw []byte) {
		currency := types.NewAddressFromBytes(currencyRaw)

		amountMaxClamped := amountMaxRaw
		if len(amountMaxClamped) > 32 {
			amountMaxClamped = amountMaxClamped[len(amountMaxClamped)-32:]
		}
		amountMax := new(big.Int).SetBytes(amountMaxClamped)

		data, err := EncodeActionClearOrTake(currency, amountMax)
		require.NoError(t, err)

		decodedCurrency, decodedAmountMax, err := DecodeActionClearOrTake(data)
		require.NoError(t, err)
		require.True(t, decodedCurrency.Equal(currency))
		require.Zero(t, amountMax.Cmp(decodedAmountMax))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeActionSweepRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeActionSweepRoundTrip(f *testing.F) {
	f.Add([]byte{0x11}, []byte{0x22})
	f.Add(make([]byte, 20), make([]byte, 20))

	f.Fuzz(func(t *testing.T, currencyRaw, toRaw []byte) {
		currency := types.NewAddressFromBytes(currencyRaw)
		to := types.NewAddressFromBytes(toRaw)

		data, err := EncodeActionSweep(currency, to)
		require.NoError(t, err)

		decodedCurrency, decodedTo, err := DecodeActionSweep(data)
		require.NoError(t, err)
		require.True(t, decodedCurrency.Equal(currency))
		require.True(t, decodedTo.Equal(to))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeActionWrapRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeActionWrapRoundTrip(f *testing.F) {
	f.Add([]byte{0x03, 0xe8})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, amountRaw []byte) {
		amountClamped := amountRaw
		if len(amountClamped) > 32 {
			amountClamped = amountClamped[len(amountClamped)-32:]
		}
		amount := new(big.Int).SetBytes(amountClamped)

		data, err := EncodeActionWrap(amount)
		require.NoError(t, err)

		decodedAmount, err := DecodeActionWrap(data)
		require.NoError(t, err)
		require.Zero(t, amount.Cmp(decodedAmount))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeActionUnwrapRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeActionUnwrapRoundTrip(f *testing.F) {
	f.Add([]byte{0x03, 0xe8})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, amountRaw []byte) {
		amountClamped := amountRaw
		if len(amountClamped) > 32 {
			amountClamped = amountClamped[len(amountClamped)-32:]
		}
		amount := new(big.Int).SetBytes(amountClamped)

		data, err := EncodeActionUnwrap(amount)
		require.NoError(t, err)

		decodedAmount, err := DecodeActionUnwrap(data)
		require.NoError(t, err)
		require.Zero(t, amount.Cmp(decodedAmount))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeModifyLiquiditiesRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeModifyLiquiditiesRoundTrip(f *testing.F) {
	f.Add([]byte{0x00, 0x0d}, []byte{0x01, 0x02}, []byte{0x03}, []byte{0x66, 0x1e, 0xf4, 0x00})
	f.Add([]byte{}, []byte{}, []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, actions, param0, param1, deadlineRaw []byte) {
		params := [][]byte{param0, param1}

		deadlineClamped := deadlineRaw
		if len(deadlineClamped) > 32 {
			deadlineClamped = deadlineClamped[len(deadlineClamped)-32:]
		}
		deadline := new(big.Int).SetBytes(deadlineClamped)

		data, err := EncodeModifyLiquidities(actions, params, deadline)
		require.NoError(t, err)

		vals, err := modifyLiquiditiesFn.DecodeCall(data)
		require.NoError(t, err)
		unlockData, ok := vals[0].([]byte)
		require.True(t, ok)
		decodedDeadline, ok := vals[1].(*big.Int)
		require.True(t, ok)
		require.Zero(t, deadline.Cmp(decodedDeadline))

		decodedActions, decodedParams, err := DecodeActions(unlockData)
		require.NoError(t, err)
		require.Equal(t, actions, decodedActions)
		require.Equal(t, params, decodedParams)
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeModifyLiquiditiesWithoutUnlockRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeModifyLiquiditiesWithoutUnlockRoundTrip(f *testing.F) {
	f.Add([]byte{0x01}, []byte{0x04}, []byte{0x05, 0x06})
	f.Add([]byte{}, []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, actions, param0, param1 []byte) {
		params := [][]byte{param0, param1}

		data, err := EncodeModifyLiquiditiesWithoutUnlock(actions, params)
		require.NoError(t, err)

		vals, err := modifyLiquiditiesWithoutUnlockFn.DecodeCall(data)
		require.NoError(t, err)
		decodedActions, ok := vals[0].([]byte)
		require.True(t, ok)
		require.Equal(t, actions, decodedActions)
		decodedParamsAny, ok := vals[1].([]any)
		require.True(t, ok)
		decodedParams := make([][]byte, len(decodedParamsAny))
		for i, v := range decodedParamsAny {
			decodedParams[i], ok = v.([]byte)
			require.True(t, ok)
		}
		require.Equal(t, params, decodedParams)
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeModifyLiquiditiesCallDoesNotPanic$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzDecodeModifyLiquiditiesCallDoesNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x01})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 64))
	f.Add(make([]byte, 96))

	f.Fuzz(func(t *testing.T, data []byte) {
		require.NotPanics(t, func() {
			_, _ = modifyLiquiditiesFn.DecodeCall(data)
			_, _ = modifyLiquiditiesWithoutUnlockFn.DecodeCall(data)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzAllPositionNotifierPermitDecodersDoNotPanic$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzAllPositionNotifierPermitDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x01})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 33))
	f.Add(make([]byte, 64))

	f.Fuzz(func(t *testing.T, data []byte) {
		require.NotPanics(t, func() {
			_, _ = DecodeSubscriber(data)
			_, _ = DecodeUnsubscribeGasLimit(data)
			_, _ = DecodeInitializePool(data)
			_, _ = DecodeMulticall(data)
			_, _ = DecodePoolManager(data)
			_, _ = DecodeNonces(data)
			_, _ = subscribeFn.DecodeCall(data)
			_, _ = unsubscribeFn.DecodeCall(data)
			_, _ = initializePoolFn.DecodeCall(data)
			_, _ = multicallFn.DecodeCall(data)
			_, _ = noncesFn.DecodeCall(data)
			_, _ = revokeNonceFn.DecodeCall(data)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeSubscribeRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeSubscribeRoundTrip(f *testing.F) {
	f.Add([]byte{0x2a}, []byte{0x11}, []byte{0xde, 0xad})
	f.Add(make([]byte, 32), make([]byte, 20), []byte{})

	f.Fuzz(func(t *testing.T, tokenIDRaw, newSubscriberRaw, data []byte) {
		tokenIDClamped := tokenIDRaw
		if len(tokenIDClamped) > 32 {
			tokenIDClamped = tokenIDClamped[len(tokenIDClamped)-32:]
		}
		tokenID := new(big.Int).SetBytes(tokenIDClamped)
		newSubscriber := types.NewAddressFromBytes(newSubscriberRaw)

		calldata, err := EncodeSubscribe(tokenID, newSubscriber, data)
		require.NoError(t, err)

		vals, err := subscribeFn.DecodeCall(calldata)
		require.NoError(t, err)
		decodedTokenID, ok := vals[0].(*big.Int)
		require.True(t, ok)
		require.Zero(t, tokenID.Cmp(decodedTokenID))
		decodedSubscriber, ok := vals[1].(*types.Address)
		require.True(t, ok)
		require.True(t, decodedSubscriber.Equal(newSubscriber))
		decodedData, ok := vals[2].([]byte)
		require.True(t, ok)
		require.Equal(t, data, decodedData)
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeInitializePoolRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeInitializePoolRoundTrip(f *testing.F) {
	f.Add([]byte{0x11}, []byte{0x22}, []byte{0x33}, []byte{0x0b, 0xb8}, []byte{0x00, 0x3c}, []byte{0x01, 0xf4})
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

		calldata, err := EncodeInitializePool(key, sqrtPriceX96)
		require.NoError(t, err)

		vals, err := initializePoolFn.DecodeCall(calldata)
		require.NoError(t, err)
		poolKeyFields, ok := vals[0].([]any)
		require.True(t, ok)
		decodedCurrency0, ok := poolKeyFields[0].(*types.Address)
		require.True(t, ok)
		require.True(t, decodedCurrency0.Equal(currency0))
		decodedCurrency1, ok := poolKeyFields[1].(*types.Address)
		require.True(t, ok)
		require.True(t, decodedCurrency1.Equal(currency1))
		decodedFee, ok := poolKeyFields[2].(*big.Int)
		require.True(t, ok)
		require.Zero(t, fee.Cmp(decodedFee))
		decodedTickSpacing, ok := poolKeyFields[3].(*big.Int)
		require.True(t, ok)
		require.Zero(t, tickSpacing.Cmp(decodedTickSpacing))
		decodedHooks, ok := poolKeyFields[4].(*types.Address)
		require.True(t, ok)
		require.True(t, decodedHooks.Equal(hooks))
		decodedSqrtPriceX96, ok := vals[1].(*big.Int)
		require.True(t, ok)
		require.Zero(t, sqrtPriceX96.Cmp(decodedSqrtPriceX96))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeMulticallRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeMulticallRoundTrip(f *testing.F) {
	f.Add([]byte{0x01, 0x02}, []byte{0x03})
	f.Add([]byte{}, []byte{})

	f.Fuzz(func(t *testing.T, data0, data1 []byte) {
		data := [][]byte{data0, data1}

		calldata, err := EncodeMulticall(data)
		require.NoError(t, err)

		vals, err := multicallFn.DecodeCall(calldata)
		require.NoError(t, err)
		elems, ok := vals[0].([]any)
		require.True(t, ok)
		decoded := make([][]byte, len(elems))
		for i, e := range elems {
			decoded[i], ok = e.([]byte)
			require.True(t, ok)
		}
		require.Equal(t, data, decoded)
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeNoncesRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeNoncesRoundTrip(f *testing.F) {
	f.Add([]byte{0x11}, []byte{0x01})
	f.Add(make([]byte, 20), []byte{})

	f.Fuzz(func(t *testing.T, ownerRaw, wordRaw []byte) {
		owner := types.NewAddressFromBytes(ownerRaw)

		wordClamped := wordRaw
		if len(wordClamped) > 32 {
			wordClamped = wordClamped[len(wordClamped)-32:]
		}
		word := new(big.Int).SetBytes(wordClamped)

		calldata, err := EncodeNonces(owner, word)
		require.NoError(t, err)

		vals, err := noncesFn.DecodeCall(calldata)
		require.NoError(t, err)
		decodedOwner, ok := vals[0].(*types.Address)
		require.True(t, ok)
		require.True(t, decodedOwner.Equal(owner))
		decodedWord, ok := vals[1].(*big.Int)
		require.True(t, ok)
		require.Zero(t, word.Cmp(decodedWord))
	})
}

// go test -run '^$' -fuzz '^FuzzAllPositionPermitDecodersDoNotPanic$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzAllPositionPermitDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x01})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 64))
	f.Add(make([]byte, 96))
	f.Add(make([]byte, 160))

	f.Fuzz(func(t *testing.T, data []byte) {
		require.NotPanics(t, func() {
			_, _ = erc721PermitFn.DecodeCall(data)
			_, _ = erc721PermitForAllFn.DecodeCall(data)
			_, _ = permit2PermitFn.DecodeCall(data)
			_, _ = permit2PermitBatchFn.DecodeCall(data)
			_, _ = DecodePermit2Permit(data)
			_, _ = DecodePermit2PermitBatch(data)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeERC721PermitRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeERC721PermitRoundTrip(f *testing.F) {
	f.Add([]byte{0x11}, []byte{0x2a}, []byte{0x66, 0x1e, 0xf4, 0x00}, []byte{0x07}, []byte{0xde, 0xad})
	f.Add(make([]byte, 20), make([]byte, 32), []byte{}, []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, spenderRaw, tokenIDRaw, deadlineRaw, nonceRaw, signature []byte) {
		spender := types.NewAddressFromBytes(spenderRaw)

		tokenIDClamped := tokenIDRaw
		if len(tokenIDClamped) > 32 {
			tokenIDClamped = tokenIDClamped[len(tokenIDClamped)-32:]
		}
		tokenID := new(big.Int).SetBytes(tokenIDClamped)

		deadlineClamped := deadlineRaw
		if len(deadlineClamped) > 32 {
			deadlineClamped = deadlineClamped[len(deadlineClamped)-32:]
		}
		deadline := new(big.Int).SetBytes(deadlineClamped)

		nonceClamped := nonceRaw
		if len(nonceClamped) > 32 {
			nonceClamped = nonceClamped[len(nonceClamped)-32:]
		}
		nonce := new(big.Int).SetBytes(nonceClamped)

		calldata, err := EncodeERC721Permit(spender, tokenID, deadline, nonce, signature)
		require.NoError(t, err)

		vals, err := erc721PermitFn.DecodeCall(calldata)
		require.NoError(t, err)
		decodedSpender, ok := vals[0].(*types.Address)
		require.True(t, ok)
		require.True(t, decodedSpender.Equal(spender))
		decodedTokenID, ok := vals[1].(*big.Int)
		require.True(t, ok)
		require.Zero(t, tokenID.Cmp(decodedTokenID))
		decodedDeadline, ok := vals[2].(*big.Int)
		require.True(t, ok)
		require.Zero(t, deadline.Cmp(decodedDeadline))
		decodedNonce, ok := vals[3].(*big.Int)
		require.True(t, ok)
		require.Zero(t, nonce.Cmp(decodedNonce))
		decodedSignature, ok := vals[4].([]byte)
		require.True(t, ok)
		require.Equal(t, signature, decodedSignature)
	})
}

// go test -run '^$' -fuzz '^FuzzEncodePermit2PermitRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodePermit2PermitRoundTrip(f *testing.F) {
	f.Add([]byte{0x11}, []byte{0x22}, []byte{0x03, 0xe8}, []byte{0x66, 0x1e, 0xf4, 0x00}, []byte{0x01}, []byte{0x33}, []byte{0x66, 0x1e, 0xf4, 0x00}, []byte{0xde, 0xad})
	f.Add(make([]byte, 20), make([]byte, 20), []byte{}, []byte{}, []byte{}, make([]byte, 20), []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, ownerRaw, tokenRaw, amountRaw, expirationRaw, nonceRaw, spenderRaw, sigDeadlineRaw, signature []byte) {
		owner := types.NewAddressFromBytes(ownerRaw)
		token := types.NewAddressFromBytes(tokenRaw)
		spender := types.NewAddressFromBytes(spenderRaw)

		amountClamped := amountRaw
		if len(amountClamped) > 20 {
			amountClamped = amountClamped[len(amountClamped)-20:]
		}
		amount := new(big.Int).SetBytes(amountClamped)

		expirationClamped := expirationRaw
		if len(expirationClamped) > 6 {
			expirationClamped = expirationClamped[len(expirationClamped)-6:]
		}
		expiration := new(big.Int).SetBytes(expirationClamped)

		nonceClamped := nonceRaw
		if len(nonceClamped) > 6 {
			nonceClamped = nonceClamped[len(nonceClamped)-6:]
		}
		nonce := new(big.Int).SetBytes(nonceClamped)

		sigDeadlineClamped := sigDeadlineRaw
		if len(sigDeadlineClamped) > 32 {
			sigDeadlineClamped = sigDeadlineClamped[len(sigDeadlineClamped)-32:]
		}
		sigDeadline := new(big.Int).SetBytes(sigDeadlineClamped)

		details := permit2.NewPermitDetails(token, amount, expiration, nonce)
		permitSingle := permit2.NewPermitSingle(details, spender, sigDeadline)

		calldata, err := EncodePermit2Permit(owner, permitSingle, signature)
		require.NoError(t, err)

		vals, err := permit2PermitFn.DecodeCall(calldata)
		require.NoError(t, err)
		decodedOwner, ok := vals[0].(*types.Address)
		require.True(t, ok)
		require.True(t, decodedOwner.Equal(owner))
		permitSingleFields, ok := vals[1].([]any)
		require.True(t, ok)
		detailsFields, ok := permitSingleFields[0].([]any)
		require.True(t, ok)
		decodedToken, ok := detailsFields[0].(*types.Address)
		require.True(t, ok)
		require.True(t, decodedToken.Equal(token))
		decodedAmount, ok := detailsFields[1].(*big.Int)
		require.True(t, ok)
		require.Zero(t, amount.Cmp(decodedAmount))
		decodedExpiration, ok := detailsFields[2].(*big.Int)
		require.True(t, ok)
		require.Zero(t, expiration.Cmp(decodedExpiration))
		decodedNonce, ok := detailsFields[3].(*big.Int)
		require.True(t, ok)
		require.Zero(t, nonce.Cmp(decodedNonce))
		decodedSpender, ok := permitSingleFields[1].(*types.Address)
		require.True(t, ok)
		require.True(t, decodedSpender.Equal(spender))
		decodedSigDeadline, ok := permitSingleFields[2].(*big.Int)
		require.True(t, ok)
		require.Zero(t, sigDeadline.Cmp(decodedSigDeadline))
		decodedSignature, ok := vals[2].([]byte)
		require.True(t, ok)
		require.Equal(t, signature, decodedSignature)
	})
}

// go test -run '^$' -fuzz '^FuzzPositionManagerModifyLiquiditiesThroughClient$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzPositionManagerModifyLiquiditiesThroughClient(f *testing.F) {
	f.Add([]byte{ActionMintPosition, ActionSettlePair}, []byte{0x01, 0x02}, uint64(1_700_000_000), uint64(0), false)
	f.Add([]byte{ActionBurnPosition}, []byte{}, uint64(0), uint64(1), true)
	f.Add([]byte{0xff, 0x00, 0x7f}, make([]byte, 64), uint64(1<<63), uint64(1<<40), true)

	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(f, err)

	f.Fuzz(func(t *testing.T, actions []byte, blob []byte, deadlineRaw uint64, nativeRaw uint64, withNative bool) {
		if len(actions) == 0 {
			return
		}
		params := make([][]byte, len(actions))
		for i := range params {
			params[i] = append([]byte{byte(i)}, blob...)
		}
		deadline := new(big.Int).SetUint64(deadlineRaw)
		var native *big.Int
		if withNative {
			native = new(big.Int).SetUint64(nativeRaw)
		}

		builder := tx.NewDynamicFeeBuilder()
		builder.ChainID(big.NewInt(1))
		builder.Nonce(3)
		builder.Limit(6_000_000)
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

		_, hash, err := NewPositionManager(c).ModifyLiquidities(t.Context(), bind.NewTransactOptions(key, builder), NewModifyLiquiditiesRequest(testPositionManager, actions, params, deadline, native))
		require.NoError(t, err)
		require.True(t, hash.Equal(core.Keccak256(raw)))

		decoded, err := core.DecodeTransaction(raw)
		require.NoError(t, err)
		dyn := decoded.(*core.DynamicFeeTx)
		if native == nil {
			require.Zero(t, dyn.Value.Sign())
		} else {
			require.Zero(t, native.Cmp(dyn.Value))
		}

		vals, err := modifyLiquiditiesFn.DecodeCall(dyn.Data)
		require.NoError(t, err)
		require.Zero(t, deadline.Cmp(vals[1].(*big.Int)))
		gotActions, gotParams, err := DecodeActions(vals[0].([]byte))
		require.NoError(t, err)
		require.Equal(t, actions, gotActions)
		require.Equal(t, params, gotParams)
	})
}

// go test -run '^$' -fuzz '^FuzzPositionManagerRequestsValidateNativeAmount$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzPositionManagerRequestsValidateNativeAmount(f *testing.F) {
	f.Add([]byte{}, false)
	f.Add([]byte{0x01}, true)
	f.Add(make([]byte, 32), false)
	f.Add(append([]byte{0x01}, make([]byte, 32)...), false)
	f.Add(append([]byte{0xff}, make([]byte, 31)...), true)

	f.Fuzz(func(t *testing.T, magnitude []byte, negative bool) {
		native := new(big.Int).SetBytes(magnitude)
		if negative {
			native.Neg(native)
		}
		valid := native.Sign() >= 0 && native.BitLen() <= 256

		for _, err := range []error{
			NewModifyLiquiditiesRequest(testPositionManager, []byte{ActionSettlePair}, [][]byte{{0x01}}, big.NewInt(1), native).Validate(),
			NewModifyLiquiditiesWithoutUnlockRequest(testPositionManager, []byte{ActionSettlePair}, [][]byte{{0x01}}, native).Validate(),
			NewMulticallRequest(testPositionManager, [][]byte{{0x01}}, native).Validate(),
		} {
			if valid {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, bind.ErrInvalidRequest)
			}
		}
	})
}

// go test -run '^$' -fuzz '^FuzzPositionManagerReadsSurviveArbitraryResult$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzPositionManagerReadsSurviveArbitraryResult(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 64))
	f.Add(make([]byte, 192))
	f.Add(append(make([]byte, 31), 0xff))

	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	tokenID := big.NewInt(1)

	f.Fuzz(func(t *testing.T, result []byte) {
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(result)
			return nil
		}).Times(13)
		p := NewPositionManager(c)
		ctx := t.Context()
		pm := testPositionManager

		poolManager, err := p.PoolManager(ctx, nil, NewPoolManagerRequest(pm))
		require.True(t, (err == nil) != (poolManager == nil))
		weth9, err := p.WETH9(ctx, nil, NewWETH9Request(pm))
		require.True(t, (err == nil) != (weth9 == nil))
		permit2Addr, err := p.Permit2(ctx, nil, NewPermit2Request(pm))
		require.True(t, (err == nil) != (permit2Addr == nil))
		next, err := p.NextTokenID(ctx, nil, NewNextTokenIDRequest(pm))
		require.True(t, (err == nil) != (next == nil))
		liquidity, err := p.GetPositionLiquidity(ctx, nil, NewGetPositionLiquidityRequest(pm, tokenID))
		require.True(t, (err == nil) != (liquidity == nil))
		gotKey, gotInfo, err := p.GetPoolAndPositionInfo(ctx, nil, NewGetPoolAndPositionInfoRequest(pm, tokenID))
		require.True(t, (err == nil) != (gotKey == nil))
		require.True(t, (err == nil) != (gotInfo == nil))
		info, err := p.PositionInfo(ctx, nil, NewPositionInfoRequest(pm, tokenID))
		require.True(t, (err == nil) != (info == nil))
		byID, err := p.PoolKeys(ctx, nil, NewPoolKeysRequest(pm, [25]byte{}))
		require.True(t, (err == nil) != (byID == nil))
		subscriber, err := p.Subscriber(ctx, nil, NewSubscriberRequest(pm, tokenID))
		require.True(t, (err == nil) != (subscriber == nil))
		gasLimit, err := p.UnsubscribeGasLimit(ctx, nil, NewUnsubscribeGasLimitRequest(pm))
		require.True(t, (err == nil) != (gasLimit == nil))
		nonces, err := p.Nonces(ctx, nil, NewNoncesRequest(pm, testSender, big.NewInt(0)))
		require.True(t, (err == nil) != (nonces == nil))
		tick, err := p.CallInitializePool(ctx, nil, NewInitializePoolRequest(pm, key, big.NewInt(1)))
		require.True(t, (err == nil) != (tick == nil))
		results, err := p.CallMulticall(ctx, nil, NewMulticallRequest(pm, [][]byte{{0x01}}, nil))
		if err != nil {
			require.Nil(t, results)
		}
	})
}

// go test -run '^$' -fuzz '^FuzzPositionManagerWrapRevertErrorKeepsCause$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzPositionManagerWrapRevertErrorKeepsCause(f *testing.F) {
	for _, e := range []*abi.Error{errNotApproved, errDeadlinePassed, errInvalidSigner, errCurrencyNotSettled, errWrappedError} {
		sel := e.Selector()
		f.Add(sel[:])
	}
	f.Add([]byte{})
	f.Add([]byte{0x08, 0xc3, 0x79, 0xa0})

	f.Fuzz(func(t *testing.T, data []byte) {
		rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(data) + `"`)}

		var wrapped error
		require.NotPanics(t, func() { wrapped = mapPositionManagerRevertError(rpcErr) })
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
