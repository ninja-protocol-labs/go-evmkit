package universalrouter

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/protocols/permit2"
	v4 "github.com/ninja-protocol-labs/go-evmkit/protocols/uniswap/v4"
)

// go test -run '^$' -fuzz '^FuzzParseCommandByteRoundTrip$' -fuzztime=10s ./protocols/uniswap/universalrouter
func FuzzParseCommandByteRoundTrip(f *testing.F) {
	for _, b := range []byte{0x00, 0x0b, 0x0f, 0x10, 0x21, 0x40, 0x7f, 0x80, 0x8e, 0xa1, 0xc0, 0xff} {
		f.Add(b)
	}

	f.Fuzz(func(t *testing.T, b byte) {
		cmd, allowRevert := ParseCommandByte(b)
		require.Equal(t, b, cmd.Byte(allowRevert))
		require.Equal(t, b&0x80 != 0, allowRevert)
		require.Zero(t, byte(cmd)&0x80)
		if cmd.Valid() {
			require.NotContains(t, cmd.String(), "0x")
		} else {
			require.Len(t, cmd.String(), 4)
		}
	})
}

// go test -run '^$' -fuzz '^FuzzV3SwapInputRoundTrip$' -fuzztime=10s ./protocols/uniswap/universalrouter
func FuzzV3SwapInputRoundTrip(f *testing.F) {
	f.Add([]byte{0x01}, []byte{0x0f, 0x42, 0x40}, []byte{0x01}, uint8(1), []byte{0xaa}, true, true, false)
	f.Add([]byte{}, []byte{}, []byte{}, uint8(3), []byte{}, false, false, true)
	f.Add(bytes.Repeat([]byte{0xff}, 20), bytes.Repeat([]byte{0xff}, 32), bytes.Repeat([]byte{0xff}, 32), uint8(5), bytes.Repeat([]byte{0x5a}, 64), true, true, true)

	f.Fuzz(func(t *testing.T, recipientRaw, amountRaw, limitRaw []byte, hopsRaw uint8, seed []byte, payerIsUser, withHopPrices, exactOut bool) {
		if len(recipientRaw) > 20 || len(amountRaw) > 32 || len(limitRaw) > 32 {
			return
		}
		hops := int(hopsRaw%6) + 1

		recipient := types.NewAddressFromBytes(append(make([]byte, 20-len(recipientRaw)), recipientRaw...))
		amount := new(big.Int).SetBytes(amountRaw)
		limit := new(big.Int).SetBytes(limitRaw)

		path := make([]byte, 0, 20+23*hops)
		for i := 0; i < 20+23*hops; i++ {
			var b byte
			if len(seed) > 0 {
				b = seed[i%len(seed)] ^ byte(i)
			}
			path = append(path, b)
		}

		var minHopPriceX36 []*big.Int
		if withHopPrices {
			for i := range hops {
				minHopPriceX36 = append(minHopPriceX36, new(big.Int).Lsh(big.NewInt(int64(i+1)), uint(i*40)))
			}
		}

		var input Input
		if exactOut {
			input = NewV3SwapExactOutInput(recipient, amount, limit, path, payerIsUser, minHopPriceX36)
		} else {
			input = NewV3SwapExactInInput(recipient, amount, limit, path, payerIsUser, minHopPriceX36)
		}

		data, err := input.Encode()
		require.NoError(t, err)

		decoded, err := DecodeInput(input.Command(), data)
		require.NoError(t, err)
		require.Equal(t, input.Command(), decoded.Command())

		again, err := decoded.Encode()
		require.NoError(t, err)
		require.Equal(t, data, again)

		if exactOut {
			out, err := DecodeV3SwapExactOutInput(data)
			require.NoError(t, err)
			require.True(t, out.Recipient.Equal(recipient))
			require.Zero(t, amount.Cmp(out.AmountOut))
			require.Zero(t, limit.Cmp(out.AmountInMax))
			require.Equal(t, path, out.Path)
			require.Equal(t, payerIsUser, out.PayerIsUser)
			require.Len(t, out.MinHopPriceX36, len(minHopPriceX36))
		} else {
			in, err := DecodeV3SwapExactInInput(data)
			require.NoError(t, err)
			require.True(t, in.Recipient.Equal(recipient))
			require.Zero(t, amount.Cmp(in.AmountIn))
			require.Zero(t, limit.Cmp(in.AmountOutMin))
			require.Equal(t, path, in.Path)
			require.Equal(t, payerIsUser, in.PayerIsUser)
			require.Len(t, in.MinHopPriceX36, len(minHopPriceX36))
			for i, p := range minHopPriceX36 {
				require.Zero(t, p.Cmp(in.MinHopPriceX36[i]))
			}
		}

		_, err = NewV3SwapExactInInput(recipient, amount, limit, path[:len(path)-1], payerIsUser, nil).Encode()
		require.ErrorIs(t, err, bind.ErrInvalidRequest)

		_, err = NewV3SwapExactInInput(recipient, amount, limit, path, payerIsUser, make([]*big.Int, hops+1)).Encode()
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
	})
}

// go test -run '^$' -fuzz '^FuzzV2SwapInputRoundTrip$' -fuzztime=10s ./protocols/uniswap/universalrouter
func FuzzV2SwapInputRoundTrip(f *testing.F) {
	f.Add([]byte{0x02}, uint64(1000), uint64(990), uint8(2), []byte{0x11}, true, false, false)
	f.Add([]byte{}, uint64(0), uint64(0), uint8(7), []byte{}, false, true, true)
	f.Add(bytes.Repeat([]byte{0xff}, 20), ^uint64(0), ^uint64(0), uint8(255), bytes.Repeat([]byte{0x99}, 40), true, true, true)

	f.Fuzz(func(t *testing.T, recipientRaw []byte, amount, limit uint64, lenRaw uint8, seed []byte, payerIsUser, withHopPrices, exactOut bool) {
		if len(recipientRaw) > 20 {
			return
		}
		n := int(lenRaw%8) + 2

		recipient := types.NewAddressFromBytes(append(make([]byte, 20-len(recipientRaw)), recipientRaw...))
		path := make([]*types.Address, n)
		for i := range n {
			raw := make([]byte, 20)
			for j := range raw {
				if len(seed) > 0 {
					raw[j] = seed[(i*20+j)%len(seed)]
				}
				raw[j] ^= byte(i)
			}
			path[i] = types.NewAddressFromBytes(raw)
		}

		var minHopPriceX36 []*big.Int
		if withHopPrices {
			for i := range n - 1 {
				minHopPriceX36 = append(minHopPriceX36, new(big.Int).SetUint64(uint64(i)*amount))
			}
		}

		var input Input
		if exactOut {
			input = NewV2SwapExactOutInput(recipient, new(big.Int).SetUint64(amount), new(big.Int).SetUint64(limit), path, payerIsUser, minHopPriceX36)
		} else {
			input = NewV2SwapExactInInput(recipient, new(big.Int).SetUint64(amount), new(big.Int).SetUint64(limit), path, payerIsUser, minHopPriceX36)
		}

		data, err := input.Encode()
		require.NoError(t, err)

		decoded, err := DecodeInput(input.Command(), data)
		require.NoError(t, err)

		again, err := decoded.Encode()
		require.NoError(t, err)
		require.Equal(t, data, again)

		var gotPath []*types.Address
		if exactOut {
			out, err := DecodeV2SwapExactOutInput(data)
			require.NoError(t, err)
			gotPath = out.Path
		} else {
			in, err := DecodeV2SwapExactInInput(data)
			require.NoError(t, err)
			gotPath = in.Path
		}
		require.Len(t, gotPath, n)
		for i := range n {
			require.True(t, path[i].Equal(gotPath[i]))
		}

		_, err = NewV2SwapExactInInput(recipient, big.NewInt(1), big.NewInt(1), path[:1], payerIsUser, nil).Encode()
		require.ErrorIs(t, err, bind.ErrInvalidRequest)

		_, err = NewV2SwapExactInInput(recipient, big.NewInt(1), big.NewInt(1), path, payerIsUser, make([]*big.Int, n)).Encode()
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
	})
}

// go test -run '^$' -fuzz '^FuzzPaymentInputsRoundTrip$' -fuzztime=10s ./protocols/uniswap/universalrouter
func FuzzPaymentInputsRoundTrip(f *testing.F) {
	f.Add(uint8(0), []byte{0x01}, []byte{0x02}, []byte{0x07})
	f.Add(uint8(2), []byte{}, []byte{}, []byte{0x27, 0x10})
	f.Add(uint8(2), []byte{}, []byte{}, []byte{0x27, 0x11})
	f.Add(uint8(3), bytes.Repeat([]byte{0xff}, 20), bytes.Repeat([]byte{0xee}, 20), []byte{0x0d, 0xe0, 0xb6, 0xb3, 0xa7, 0x64, 0x00, 0x00})
	f.Add(uint8(3), []byte{0x01}, []byte{0x01}, []byte{0x0d, 0xe0, 0xb6, 0xb3, 0xa7, 0x64, 0x00, 0x01})
	f.Add(uint8(6), []byte{0x01}, []byte{0x01}, bytes.Repeat([]byte{0xff}, 32))

	f.Fuzz(func(t *testing.T, kind uint8, aRaw, bRaw, amountRaw []byte) {
		if len(aRaw) > 20 || len(bRaw) > 20 || len(amountRaw) > 32 {
			return
		}
		a := types.NewAddressFromBytes(append(make([]byte, 20-len(aRaw)), aRaw...))
		b := types.NewAddressFromBytes(append(make([]byte, 20-len(bRaw)), bRaw...))
		amount := new(big.Int).SetBytes(amountRaw)

		var input Input
		valid := true
		switch kind % 8 {
		case 0:
			input = NewSweepInput(a, b, amount)
		case 1:
			input = NewTransferInput(a, b, amount)
		case 2:
			input = NewPayPortionInput(a, b, amount)
			valid = amount.Cmp(big.NewInt(10_000)) <= 0
		case 3:
			input = NewPayPortionFullPrecisionInput(a, b, amount)
			valid = amount.Cmp(big.NewInt(1_000_000_000_000_000_000)) <= 0
		case 4:
			input = NewBalanceCheckERC20Input(a, b, amount)
		case 5:
			input = NewWrapETHInput(a, amount)
		case 6:
			input = NewUnwrapWETHInput(a, amount)
		default:
			input = NewPermit2TransferFromInput(a, b, amount)
			valid = amount.BitLen() <= 160
		}

		data, err := input.Encode()
		if !valid {
			require.Error(t, err)
			require.Nil(t, data)
			return
		}
		require.NoError(t, err)

		decoded, err := DecodeInput(input.Command(), data)
		require.NoError(t, err)
		require.Equal(t, input.Command(), decoded.Command())

		again, err := decoded.Encode()
		require.NoError(t, err)
		require.Equal(t, data, again)
	})
}

// go test -run '^$' -fuzz '^FuzzPermit2InputsRoundTrip$' -fuzztime=10s ./protocols/uniswap/universalrouter
func FuzzPermit2InputsRoundTrip(f *testing.F) {
	f.Add(uint8(1), uint64(1000), uint64(1_700_000_000), uint64(3), []byte{0xaa, 0xbb})
	f.Add(uint8(4), uint64(0), uint64(0), uint64(0), bytes.Repeat([]byte{0x01}, 65))
	f.Add(uint8(16), ^uint64(0), uint64(1<<48-1), uint64(1<<48-1), []byte{0x01})

	f.Fuzz(func(t *testing.T, countRaw uint8, amount, expiration, nonce uint64, signature []byte) {
		if len(signature) == 0 {
			return
		}
		count := int(countRaw%8) + 1
		expiration &= 1<<48 - 1
		nonce &= 1<<48 - 1

		details := make([]permit2.PermitDetails, count)
		transfers := make([]permit2.AllowanceTransferDetails, count)
		for i := range count {
			token := types.NewAddressFromBytes(append(make([]byte, 19), byte(i+1)))
			details[i] = *permit2.NewPermitDetails(token, new(big.Int).SetUint64(amount+uint64(i)), new(big.Int).SetUint64(expiration), new(big.Int).SetUint64(nonce))
			transfers[i] = *permit2.NewAllowanceTransferDetails(testTokenA, testTokenB, new(big.Int).SetUint64(amount), token)
		}

		for _, input := range []Input{
			NewPermit2PermitInput(permit2.NewPermitSingle(&details[0], testTokenC, new(big.Int).SetUint64(expiration)), signature),
			NewPermit2PermitBatchInput(permit2.NewPermitBatch(details, testTokenC, new(big.Int).SetUint64(nonce)), signature),
			NewPermit2TransferFromBatchInput(transfers),
		} {
			data, err := input.Encode()
			require.NoError(t, err)

			decoded, err := DecodeInput(input.Command(), data)
			require.NoError(t, err)

			again, err := decoded.Encode()
			require.NoError(t, err)
			require.Equal(t, data, again)
		}

		batchData, err := NewPermit2PermitBatchInput(permit2.NewPermitBatch(details, testTokenC, big.NewInt(1)), signature).Encode()
		require.NoError(t, err)
		batch, err := DecodePermit2PermitBatchInput(batchData)
		require.NoError(t, err)
		require.Len(t, batch.PermitBatch.Details, count)
		require.Equal(t, signature, batch.Signature)
		for i := range count {
			require.True(t, details[i].Token.Equal(batch.PermitBatch.Details[i].Token))
			require.Zero(t, details[i].Amount.Cmp(batch.PermitBatch.Details[i].Amount))
		}
	})
}

// go test -run '^$' -fuzz '^FuzzAcrossInputRoundTrip$' -fuzztime=10s ./protocols/uniswap/universalrouter
func FuzzAcrossInputRoundTrip(f *testing.F) {
	f.Add([]byte{0x01}, uint64(1000), uint64(990), uint64(8453), uint32(1_700_000_000), uint32(1_700_003_600), uint32(0), []byte{0xca, 0xfe}, true)
	f.Add([]byte{}, uint64(0), uint64(0), uint64(0), uint32(0), uint32(0), uint32(0), []byte{}, false)
	f.Add(bytes.Repeat([]byte{0xff}, 20), ^uint64(0), ^uint64(0), ^uint64(0), ^uint32(0), ^uint32(0), ^uint32(0), bytes.Repeat([]byte{0x77}, 100), true)

	f.Fuzz(func(t *testing.T, addrRaw []byte, inputAmount, outputAmount, chainID uint64, quote, fill, exclusivity uint32, message []byte, useNative bool) {
		if len(addrRaw) > 20 {
			return
		}
		addr := types.NewAddressFromBytes(append(make([]byte, 20-len(addrRaw)), addrRaw...))

		input := NewAcrossV4DepositV3Input(addr, testTokenA, testTokenB, addr, new(big.Int).SetUint64(inputAmount), new(big.Int).SetUint64(outputAmount), new(big.Int).SetUint64(chainID), testTokenC, quote, fill, exclusivity, message, useNative)
		data, err := input.Encode()
		require.NoError(t, err)

		decoded, err := DecodeInput(CommandAcrossV4DepositV3, data)
		require.NoError(t, err)
		got, err := DecodeAcrossV4DepositV3Input(data)
		require.NoError(t, err)
		require.True(t, got.Depositor.Equal(addr))
		require.True(t, got.OutputToken.Equal(addr))
		require.Equal(t, quote, got.QuoteTimestamp)
		require.Equal(t, fill, got.FillDeadline)
		require.Equal(t, exclusivity, got.ExclusivityDeadline)
		require.Equal(t, useNative, got.UseNative)
		require.Equal(t, len(message), len(got.Message))
		require.Zero(t, new(big.Int).SetUint64(chainID).Cmp(got.DestinationChainID))

		again, err := decoded.Encode()
		require.NoError(t, err)
		require.Equal(t, data, again)
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeInputDoesNotPanic$' -fuzztime=10s ./protocols/uniswap/universalrouter
func FuzzDecodeInputDoesNotPanic(f *testing.F) {
	wrap, err := NewWrapETHInput(testTokenA, big.NewInt(1)).Encode()
	require.NoError(f, err)
	sub, err := NewExecuteSubPlanInput(NewPlan().Add(NewWrapETHInput(testTokenA, big.NewInt(1)))).Encode()
	require.NoError(f, err)
	f.Add(byte(0x0b), wrap)
	f.Add(byte(0x21), sub)
	f.Add(byte(0x00), make([]byte, 192))
	f.Add(byte(0x03), make([]byte, 64))
	f.Add(byte(0x10), []byte{})
	f.Add(byte(0x40), bytes.Repeat([]byte{0xff}, 64))
	f.Add(byte(0x13), make([]byte, 192))

	f.Fuzz(func(t *testing.T, b byte, data []byte) {
		cmd, _ := ParseCommandByte(b)

		var (
			input Input
			err   error
		)
		require.NotPanics(t, func() { input, err = DecodeInput(cmd, data) })
		if !cmd.Valid() {
			require.ErrorIs(t, err, ErrUnknownCommand)
			return
		}
		if err != nil {
			require.Nil(t, input)
			return
		}
		require.Equal(t, cmd, input.Command())

		var validateErr error
		require.NotPanics(t, func() { validateErr = input.Validate() })
		if validateErr != nil {
			return
		}

		encoded, err := input.Encode()
		require.NoError(t, err)
		redecoded, err := DecodeInput(cmd, encoded)
		require.NoError(t, err)
		again, err := redecoded.Encode()
		require.NoError(t, err)
		require.Equal(t, encoded, again)
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeExecuteCallDoesNotPanic$' -fuzztime=10s ./protocols/uniswap/universalrouter
func FuzzDecodeExecuteCallDoesNotPanic(f *testing.F) {
	commands, inputs, err := NewPlan().Add(NewWrapETHInput(AddressThis(), ContractBalance())).AddAllowRevert(NewSweepInput(testTokenA, MsgSender(), big.NewInt(0))).Encode()
	require.NoError(f, err)
	plain, err := EncodeExecute(commands, inputs)
	require.NoError(f, err)
	withDeadline, err := EncodeExecuteWithDeadline(commands, inputs, big.NewInt(1))
	require.NoError(f, err)
	signed, err := EncodeExecuteSigned(commands, inputs, [32]byte{0x01}, [32]byte{0x02}, true, [32]byte{0x03}, []byte{0x04}, big.NewInt(1))
	require.NoError(f, err)
	f.Add(plain)
	f.Add(withDeadline)
	f.Add(signed)
	f.Add([]byte{})
	f.Add([]byte{0x35, 0x93, 0x56, 0x4c})

	f.Fuzz(func(t *testing.T, calldata []byte) {
		var (
			call *ExecuteCall
			err  error
		)
		require.NotPanics(t, func() { call, err = DecodeExecuteCall(calldata) })
		if err != nil {
			require.Nil(t, call)
			return
		}
		require.NotNil(t, call.Plan)
		if bytes.HasPrefix(calldata, []byte{0x24, 0x85, 0x6b, 0xc3}) {
			require.Nil(t, call.Deadline)
		} else {
			require.NotNil(t, call.Deadline)
		}
		require.Equal(t, bytes.HasPrefix(calldata, []byte{0x44, 0x68, 0x4d, 0xc3}), call.Signed != nil)
	})
}

// go test -run '^$' -fuzz '^FuzzPlanRoundTripThroughExecute$' -fuzztime=10s ./protocols/uniswap/universalrouter
func FuzzPlanRoundTripThroughExecute(f *testing.F) {
	f.Add([]byte{0x00, 0x11, 0x22, 0x33}, uint64(1_700_000_000), uint8(0))
	f.Add([]byte{0x87, 0x86, 0x05}, uint64(0), uint8(1))
	f.Add(bytes.Repeat([]byte{0xa5}, 24), ^uint64(0), uint8(2))

	f.Fuzz(func(t *testing.T, script []byte, deadline uint64, mode uint8) {
		if len(script) == 0 || len(script) > 32 {
			return
		}

		root := NewPlan()
		current := root
		for i, op := range script {
			amount := big.NewInt(int64(op) * int64(i+1))
			var input Input
			switch op & 0x07 {
			case 0:
				input = NewWrapETHInput(AddressThis(), amount)
			case 1:
				input = NewUnwrapWETHInput(MsgSender(), amount)
			case 2:
				input = NewSweepInput(testTokenA, MsgSender(), amount)
			case 3:
				input = NewTransferInput(testTokenB, testTokenC, amount)
			case 4:
				input = NewPayPortionInput(testTokenA, testTokenB, big.NewInt(int64(op)))
			case 5:
				input = NewBalanceCheckERC20Input(testTokenC, testTokenA, amount)
			case 6:
				input = NewV4SwapInput([]byte{op, byte(i)}, [][]byte{{op}, {}})
			default:
				nested := NewPlan().Add(NewSweepInput(types.ZeroAddress(), MsgSender(), amount))
				input = NewExecuteSubPlanInput(nested)
				if op&0x08 != 0 {
					current.AddAllowRevert(input)
					current = nested
					continue
				}
			}
			if op&0x80 != 0 {
				current.AddAllowRevert(input)
			} else {
				current.Add(input)
			}
		}

		commands, inputs, err := root.Encode()
		require.NoError(t, err)
		require.Len(t, commands, len(root.Steps))
		for i, s := range root.Steps {
			cmd, allowRevert := ParseCommandByte(commands[i])
			require.Equal(t, s.Input.Command(), cmd)
			require.Equal(t, s.AllowRevert, allowRevert)
		}

		var calldata []byte
		switch mode % 3 {
		case 0:
			calldata, err = EncodeExecute(commands, inputs)
		case 1:
			calldata, err = EncodeExecuteWithDeadline(commands, inputs, new(big.Int).SetUint64(deadline))
		default:
			calldata, err = EncodeExecuteSigned(commands, inputs, [32]byte{byte(deadline)}, [32]byte{}, deadline%2 == 0, [32]byte{0xff}, []byte{0x01}, new(big.Int).SetUint64(deadline))
		}
		require.NoError(t, err)

		call, err := DecodeExecuteCall(calldata)
		require.NoError(t, err)
		if mode%3 == 0 {
			require.Nil(t, call.Deadline)
		} else {
			require.Zero(t, new(big.Int).SetUint64(deadline).Cmp(call.Deadline))
		}
		require.Equal(t, mode%3 == 2, call.Signed != nil)

		againCommands, againInputs, err := call.Plan.Encode()
		require.NoError(t, err)
		require.Equal(t, commands, againCommands)
		require.Equal(t, inputs, againInputs)
	})
}

// go test -run '^$' -fuzz '^FuzzV4PositionManagerCallOnlyAllowsMint$' -fuzztime=10s ./protocols/uniswap/universalrouter
func FuzzV4PositionManagerCallOnlyAllowsMint(f *testing.F) {
	f.Add([]byte{v4.ActionMintPosition, v4.ActionSettlePair})
	f.Add([]byte{v4.ActionMintPosition, v4.ActionIncreaseLiquidity})
	f.Add([]byte{v4.ActionBurnPosition})
	f.Add([]byte{v4.ActionDecreaseLiquidity, v4.ActionTakePair})
	f.Add([]byte{v4.ActionIncreaseLiquidityFromDeltas})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, actions []byte) {
		if len(actions) > 64 {
			return
		}
		params := make([][]byte, len(actions))
		for i := range params {
			params[i] = []byte{byte(i)}
		}
		calldata, err := v4.EncodeModifyLiquidities(actions, params, big.NewInt(1))
		require.NoError(t, err)

		forbidden := false
		for _, a := range actions {
			switch a {
			case v4.ActionIncreaseLiquidity, v4.ActionIncreaseLiquidityFromDeltas, v4.ActionDecreaseLiquidity, v4.ActionBurnPosition:
				forbidden = true
			}
		}

		err = NewV4PositionManagerCallInput(calldata).Validate()
		if forbidden {
			require.ErrorIs(t, err, bind.ErrInvalidRequest)
			require.ErrorContains(t, err, "only mint")
		} else {
			require.NoError(t, err)
		}
	})
}
