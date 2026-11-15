package v4

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// go test -run '^$' -fuzz '^FuzzAllRouterDecodersDoNotPanic$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzAllRouterDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x01})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 33))
	f.Add(make([]byte, 64))
	f.Add(make([]byte, 96))
	f.Add(make([]byte, 160))
	f.Add(make([]byte, 224))

	f.Fuzz(func(t *testing.T, data []byte) {
		require.NotPanics(t, func() {
			_, _ = DecodeActionSwapExactInSingle(data)
			_, _ = DecodeActionSwapExactIn(data)
			_, _ = DecodeActionSwapExactOutSingle(data)
			_, _ = DecodeActionSwapExactOut(data)
			_, _, _, _ = DecodeActionSettle(data)
			_, _, _ = DecodeActionSettleAll(data)
			_, _, _, _ = DecodeActionTake(data)
			_, _, _ = DecodeActionTakeAll(data)
			_, _, _, _ = DecodeActionTakePortion(data)
			_, _, _ = DecodeActions(data)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeActionSwapExactInSingleRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeActionSwapExactInSingleRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x11}, []byte{0x22}, []byte{0x0b, 0xb8}, []byte{0x00, 0x3c}, true, []byte{0x03, 0xe8}, []byte{0x03, 0x84}, []byte{0x01}, []byte{0xde, 0xad})
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 20), []byte{}, []byte{}, false, []byte{}, []byte{}, []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, currency0Raw, currency1Raw, hooksRaw, feeRaw, tickSpacingRaw []byte, zeroForOne bool, amountInRaw, amountOutMinimumRaw, minHopPriceX36Raw, hookData []byte) {
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

		amountInClamped := amountInRaw
		if len(amountInClamped) > 16 {
			amountInClamped = amountInClamped[len(amountInClamped)-16:]
		}
		amountIn := new(big.Int).SetBytes(amountInClamped)

		amountOutMinimumClamped := amountOutMinimumRaw
		if len(amountOutMinimumClamped) > 16 {
			amountOutMinimumClamped = amountOutMinimumClamped[len(amountOutMinimumClamped)-16:]
		}
		amountOutMinimum := new(big.Int).SetBytes(amountOutMinimumClamped)

		minHopPriceX36Clamped := minHopPriceX36Raw
		if len(minHopPriceX36Clamped) > 32 {
			minHopPriceX36Clamped = minHopPriceX36Clamped[len(minHopPriceX36Clamped)-32:]
		}
		minHopPriceX36 := new(big.Int).SetBytes(minHopPriceX36Clamped)

		poolKey := NewPoolKey(currency0, currency1, fee, tickSpacing, hooks)
		params := NewExactInputSingleParams(poolKey, zeroForOne, amountIn, amountOutMinimum, minHopPriceX36, hookData)

		data, err := EncodeActionSwapExactInSingle(params)
		require.NoError(t, err)

		decoded, err := DecodeActionSwapExactInSingle(data)
		require.NoError(t, err)
		require.True(t, decoded.PoolKey.Currency0.Equal(currency0))
		require.True(t, decoded.PoolKey.Currency1.Equal(currency1))
		require.Zero(t, fee.Cmp(decoded.PoolKey.Fee))
		require.Zero(t, tickSpacing.Cmp(decoded.PoolKey.TickSpacing))
		require.True(t, decoded.PoolKey.Hooks.Equal(hooks))
		require.Equal(t, zeroForOne, decoded.ZeroForOne)
		require.Zero(t, amountIn.Cmp(decoded.AmountIn))
		require.Zero(t, amountOutMinimum.Cmp(decoded.AmountOutMinimum))
		require.Zero(t, minHopPriceX36.Cmp(decoded.MinHopPriceX36))
		require.Equal(t, hookData, decoded.HookData)
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeActionSwapExactOutSingleRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeActionSwapExactOutSingleRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x11}, []byte{0x22}, []byte{0x0b, 0xb8}, []byte{0x00, 0x3c}, true, []byte{0x03, 0xe8}, []byte{0x04, 0x4c}, []byte{0x01}, []byte{0xbe, 0xef})
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 20), []byte{}, []byte{}, false, []byte{}, []byte{}, []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, currency0Raw, currency1Raw, hooksRaw, feeRaw, tickSpacingRaw []byte, zeroForOne bool, amountOutRaw, amountInMaximumRaw, minHopPriceX36Raw, hookData []byte) {
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

		amountOutClamped := amountOutRaw
		if len(amountOutClamped) > 16 {
			amountOutClamped = amountOutClamped[len(amountOutClamped)-16:]
		}
		amountOut := new(big.Int).SetBytes(amountOutClamped)

		amountInMaximumClamped := amountInMaximumRaw
		if len(amountInMaximumClamped) > 16 {
			amountInMaximumClamped = amountInMaximumClamped[len(amountInMaximumClamped)-16:]
		}
		amountInMaximum := new(big.Int).SetBytes(amountInMaximumClamped)

		minHopPriceX36Clamped := minHopPriceX36Raw
		if len(minHopPriceX36Clamped) > 32 {
			minHopPriceX36Clamped = minHopPriceX36Clamped[len(minHopPriceX36Clamped)-32:]
		}
		minHopPriceX36 := new(big.Int).SetBytes(minHopPriceX36Clamped)

		poolKey := NewPoolKey(currency0, currency1, fee, tickSpacing, hooks)
		params := NewExactOutputSingleParams(poolKey, zeroForOne, amountOut, amountInMaximum, minHopPriceX36, hookData)

		data, err := EncodeActionSwapExactOutSingle(params)
		require.NoError(t, err)

		decoded, err := DecodeActionSwapExactOutSingle(data)
		require.NoError(t, err)
		require.True(t, decoded.PoolKey.Currency0.Equal(currency0))
		require.True(t, decoded.PoolKey.Currency1.Equal(currency1))
		require.Zero(t, fee.Cmp(decoded.PoolKey.Fee))
		require.Zero(t, tickSpacing.Cmp(decoded.PoolKey.TickSpacing))
		require.True(t, decoded.PoolKey.Hooks.Equal(hooks))
		require.Equal(t, zeroForOne, decoded.ZeroForOne)
		require.Zero(t, amountOut.Cmp(decoded.AmountOut))
		require.Zero(t, amountInMaximum.Cmp(decoded.AmountInMaximum))
		require.Zero(t, minHopPriceX36.Cmp(decoded.MinHopPriceX36))
		require.Equal(t, hookData, decoded.HookData)
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeActionSwapExactInRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeActionSwapExactInRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, uint8(2), []byte{0x01}, []byte{0x03, 0xe8}, []byte{0x03, 0x84})
	f.Add(make([]byte, 20), uint8(0), []byte{}, []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, currencyInRaw []byte, n uint8, seed []byte, amountInRaw, amountOutMinimumRaw []byte) {
		currencyIn := types.NewAddressFromBytes(currencyInRaw)

		count := int(n % 8)
		path := make([]*PathKey, count)
		minHopPriceX36 := make([]*big.Int, count)
		for i := range path {
			h := append(append([]byte{}, seed...), byte(i))
			intermediate := types.NewAddressFromBytes(h)
			fee := new(big.Int).SetBytes([]byte{byte(i), byte(i * 7)})
			tickSpacing := new(big.Int).SetBytes([]byte{byte(i * 3)})
			hooks := types.NewAddressFromBytes(append(h, byte(i)))
			path[i] = NewPathKey(intermediate, fee, tickSpacing, hooks, []byte{byte(i)})
			hopPriceSeed := append(h, byte(i*5))
			if len(hopPriceSeed) > 32 {
				hopPriceSeed = hopPriceSeed[len(hopPriceSeed)-32:]
			}
			minHopPriceX36[i] = new(big.Int).SetBytes(hopPriceSeed)
		}

		amountInClamped := amountInRaw
		if len(amountInClamped) > 16 {
			amountInClamped = amountInClamped[len(amountInClamped)-16:]
		}
		amountIn := new(big.Int).SetBytes(amountInClamped)

		amountOutMinimumClamped := amountOutMinimumRaw
		if len(amountOutMinimumClamped) > 16 {
			amountOutMinimumClamped = amountOutMinimumClamped[len(amountOutMinimumClamped)-16:]
		}
		amountOutMinimum := new(big.Int).SetBytes(amountOutMinimumClamped)

		params := NewExactInputParams(currencyIn, path, minHopPriceX36, amountIn, amountOutMinimum)

		data, err := EncodeActionSwapExactIn(params)
		require.NoError(t, err)

		decoded, err := DecodeActionSwapExactIn(data)
		require.NoError(t, err)
		require.True(t, decoded.CurrencyIn.Equal(currencyIn))
		require.Len(t, decoded.Path, count)
		for i := range path {
			require.True(t, decoded.Path[i].IntermediateCurrency.Equal(path[i].IntermediateCurrency))
			require.Zero(t, path[i].Fee.Cmp(decoded.Path[i].Fee))
			require.Zero(t, path[i].TickSpacing.Cmp(decoded.Path[i].TickSpacing))
			require.True(t, decoded.Path[i].Hooks.Equal(path[i].Hooks))
			require.Equal(t, path[i].HookData, decoded.Path[i].HookData)
		}
		require.Len(t, decoded.MinHopPriceX36, count)
		for i := range minHopPriceX36 {
			require.Zero(t, minHopPriceX36[i].Cmp(decoded.MinHopPriceX36[i]))
		}
		require.Zero(t, amountIn.Cmp(decoded.AmountIn))
		require.Zero(t, amountOutMinimum.Cmp(decoded.AmountOutMinimum))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeActionSwapExactOutRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeActionSwapExactOutRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, uint8(2), []byte{0x01}, []byte{0x03, 0xe8}, []byte{0x04, 0x4c})
	f.Add(make([]byte, 20), uint8(0), []byte{}, []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, currencyOutRaw []byte, n uint8, seed []byte, amountOutRaw, amountInMaximumRaw []byte) {
		currencyOut := types.NewAddressFromBytes(currencyOutRaw)

		count := int(n % 8)
		path := make([]*PathKey, count)
		minHopPriceX36 := make([]*big.Int, count)
		for i := range path {
			h := append(append([]byte{}, seed...), byte(i))
			intermediate := types.NewAddressFromBytes(h)
			fee := new(big.Int).SetBytes([]byte{byte(i), byte(i * 7)})
			tickSpacing := new(big.Int).SetBytes([]byte{byte(i * 3)})
			hooks := types.NewAddressFromBytes(append(h, byte(i)))
			path[i] = NewPathKey(intermediate, fee, tickSpacing, hooks, []byte{byte(i)})
			hopPriceSeed := append(h, byte(i*5))
			if len(hopPriceSeed) > 32 {
				hopPriceSeed = hopPriceSeed[len(hopPriceSeed)-32:]
			}
			minHopPriceX36[i] = new(big.Int).SetBytes(hopPriceSeed)
		}

		amountOutClamped := amountOutRaw
		if len(amountOutClamped) > 16 {
			amountOutClamped = amountOutClamped[len(amountOutClamped)-16:]
		}
		amountOut := new(big.Int).SetBytes(amountOutClamped)

		amountInMaximumClamped := amountInMaximumRaw
		if len(amountInMaximumClamped) > 16 {
			amountInMaximumClamped = amountInMaximumClamped[len(amountInMaximumClamped)-16:]
		}
		amountInMaximum := new(big.Int).SetBytes(amountInMaximumClamped)

		params := NewExactOutputParams(currencyOut, path, minHopPriceX36, amountOut, amountInMaximum)

		data, err := EncodeActionSwapExactOut(params)
		require.NoError(t, err)

		decoded, err := DecodeActionSwapExactOut(data)
		require.NoError(t, err)
		require.True(t, decoded.CurrencyOut.Equal(currencyOut))
		require.Len(t, decoded.Path, count)
		for i := range path {
			require.True(t, decoded.Path[i].IntermediateCurrency.Equal(path[i].IntermediateCurrency))
			require.Zero(t, path[i].Fee.Cmp(decoded.Path[i].Fee))
			require.Zero(t, path[i].TickSpacing.Cmp(decoded.Path[i].TickSpacing))
			require.True(t, decoded.Path[i].Hooks.Equal(path[i].Hooks))
			require.Equal(t, path[i].HookData, decoded.Path[i].HookData)
		}
		require.Len(t, decoded.MinHopPriceX36, count)
		for i := range minHopPriceX36 {
			require.Zero(t, minHopPriceX36[i].Cmp(decoded.MinHopPriceX36[i]))
		}
		require.Zero(t, amountOut.Cmp(decoded.AmountOut))
		require.Zero(t, amountInMaximum.Cmp(decoded.AmountInMaximum))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeActionSettleRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeActionSettleRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x03, 0xe8}, true)
	f.Add(make([]byte, 20), []byte{}, false)

	f.Fuzz(func(t *testing.T, currencyRaw, amountRaw []byte, payerIsUser bool) {
		currency := types.NewAddressFromBytes(currencyRaw)

		amountClamped := amountRaw
		if len(amountClamped) > 32 {
			amountClamped = amountClamped[len(amountClamped)-32:]
		}
		amount := new(big.Int).SetBytes(amountClamped)

		data, err := EncodeActionSettle(currency, amount, payerIsUser)
		require.NoError(t, err)

		decodedCurrency, decodedAmount, decodedPayerIsUser, err := DecodeActionSettle(data)
		require.NoError(t, err)
		require.True(t, decodedCurrency.Equal(currency))
		require.Zero(t, amount.Cmp(decodedAmount))
		require.Equal(t, payerIsUser, decodedPayerIsUser)
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeActionSettleAllRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeActionSettleAllRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x03, 0xe8})
	f.Add(make([]byte, 20), []byte{})

	f.Fuzz(func(t *testing.T, currencyRaw, maxAmountRaw []byte) {
		currency := types.NewAddressFromBytes(currencyRaw)

		maxAmountClamped := maxAmountRaw
		if len(maxAmountClamped) > 32 {
			maxAmountClamped = maxAmountClamped[len(maxAmountClamped)-32:]
		}
		maxAmount := new(big.Int).SetBytes(maxAmountClamped)

		data, err := EncodeActionSettleAll(currency, maxAmount)
		require.NoError(t, err)

		decodedCurrency, decodedMaxAmount, err := DecodeActionSettleAll(data)
		require.NoError(t, err)
		require.True(t, decodedCurrency.Equal(currency))
		require.Zero(t, maxAmount.Cmp(decodedMaxAmount))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeActionTakeRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeActionTakeRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x11}, []byte{0x03, 0xe8})
	f.Add(make([]byte, 20), make([]byte, 20), []byte{})

	f.Fuzz(func(t *testing.T, currencyRaw, recipientRaw, amountRaw []byte) {
		currency := types.NewAddressFromBytes(currencyRaw)
		recipient := types.NewAddressFromBytes(recipientRaw)

		amountClamped := amountRaw
		if len(amountClamped) > 32 {
			amountClamped = amountClamped[len(amountClamped)-32:]
		}
		amount := new(big.Int).SetBytes(amountClamped)

		data, err := EncodeActionTake(currency, recipient, amount)
		require.NoError(t, err)

		decodedCurrency, decodedRecipient, decodedAmount, err := DecodeActionTake(data)
		require.NoError(t, err)
		require.True(t, decodedCurrency.Equal(currency))
		require.True(t, decodedRecipient.Equal(recipient))
		require.Zero(t, amount.Cmp(decodedAmount))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeActionTakeAllRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeActionTakeAllRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x03, 0xe8})
	f.Add(make([]byte, 20), []byte{})

	f.Fuzz(func(t *testing.T, currencyRaw, minAmountRaw []byte) {
		currency := types.NewAddressFromBytes(currencyRaw)

		minAmountClamped := minAmountRaw
		if len(minAmountClamped) > 32 {
			minAmountClamped = minAmountClamped[len(minAmountClamped)-32:]
		}
		minAmount := new(big.Int).SetBytes(minAmountClamped)

		data, err := EncodeActionTakeAll(currency, minAmount)
		require.NoError(t, err)

		decodedCurrency, decodedMinAmount, err := DecodeActionTakeAll(data)
		require.NoError(t, err)
		require.True(t, decodedCurrency.Equal(currency))
		require.Zero(t, minAmount.Cmp(decodedMinAmount))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeActionTakePortionRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeActionTakePortionRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x11}, []byte{0x13, 0x88})
	f.Add(make([]byte, 20), make([]byte, 20), []byte{})

	f.Fuzz(func(t *testing.T, currencyRaw, recipientRaw, bipsRaw []byte) {
		currency := types.NewAddressFromBytes(currencyRaw)
		recipient := types.NewAddressFromBytes(recipientRaw)

		bipsClamped := bipsRaw
		if len(bipsClamped) > 32 {
			bipsClamped = bipsClamped[len(bipsClamped)-32:]
		}
		bips := new(big.Int).SetBytes(bipsClamped)

		data, err := EncodeActionTakePortion(currency, recipient, bips)
		require.NoError(t, err)

		decodedCurrency, decodedRecipient, decodedBips, err := DecodeActionTakePortion(data)
		require.NoError(t, err)
		require.True(t, decodedCurrency.Equal(currency))
		require.True(t, decodedRecipient.Equal(recipient))
		require.Zero(t, bips.Cmp(decodedBips))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeActionsRoundTrip$' -fuzztime=10s ./protocols/uniswap/v4
func FuzzEncodeActionsRoundTrip(f *testing.F) {
	f.Add(uint8(0), []byte{0x01, 0x02, 0x03})
	f.Add(uint8(3), []byte{0xff})

	f.Fuzz(func(t *testing.T, n uint8, seed []byte) {
		count := int(n % 16)
		actions := make([]byte, count)
		params := make([][]byte, count)
		for i := range actions {
			actions[i] = byte(i)
			params[i] = append(append([]byte{}, seed...), byte(i))
		}

		data, err := EncodeActions(actions, params)
		require.NoError(t, err)

		decodedActions, decodedParams, err := DecodeActions(data)
		require.NoError(t, err)
		require.Equal(t, actions, decodedActions)
		require.Equal(t, params, decodedParams)
	})
}
