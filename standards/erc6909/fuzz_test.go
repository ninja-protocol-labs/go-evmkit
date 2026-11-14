package erc6909

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

// go test -run '^$' -fuzz '^FuzzExtractOperatorSetsDoesNotPanic$' -fuzztime=10s ./standards/erc6909
func FuzzExtractOperatorSetsDoesNotPanic(f *testing.F) {
	t0, ok := operatorSetEvent.Topic0()
	require.True(f, ok)

	f.Add(t0[:], make([]byte, 32), make([]byte, 32), []byte{})
	f.Add(t0[:], make([]byte, 32), make([]byte, 32), make([]byte, 32))
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
			logs := []tx.Log{{Address: testTokens, Topics: topics, Data: data}}
			_, _ = ExtractOperatorSets(logs)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzExtractApprovalsDoesNotPanic$' -fuzztime=10s ./standards/erc6909
func FuzzExtractApprovalsDoesNotPanic(f *testing.F) {
	t0, ok := approvalEvent.Topic0()
	require.True(f, ok)

	f.Add(t0[:], make([]byte, 32), make([]byte, 32), []byte{})
	f.Add(t0[:], make([]byte, 32), make([]byte, 32), make([]byte, 32))
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
			logs := []tx.Log{{Address: testTokens, Topics: topics, Data: data}}
			_, _ = ExtractApprovals(logs)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzExtractTransfersDoesNotPanic$' -fuzztime=10s ./standards/erc6909
func FuzzExtractTransfersDoesNotPanic(f *testing.F) {
	t0, ok := transferEvent.Topic0()
	require.True(f, ok)

	f.Add(t0[:], make([]byte, 32), make([]byte, 32), []byte{})
	f.Add(t0[:], make([]byte, 32), make([]byte, 32), make([]byte, 32))
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
			logs := []tx.Log{{Address: testTokens, Topics: topics, Data: data}}
			_, _ = ExtractTransfers(logs)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzExtractNameUpdatedDoesNotPanic$' -fuzztime=10s ./standards/erc6909
func FuzzExtractNameUpdatedDoesNotPanic(f *testing.F) {
	t0, ok := nameUpdatedEvent.Topic0()
	require.True(f, ok)

	f.Add(t0[:], make([]byte, 32), make([]byte, 32), []byte{})
	f.Add(t0[:], make([]byte, 32), make([]byte, 32), make([]byte, 32))
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
			logs := []tx.Log{{Address: testTokens, Topics: topics, Data: data}}
			_, _ = ExtractNameUpdated(logs)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzExtractSymbolUpdatedDoesNotPanic$' -fuzztime=10s ./standards/erc6909
func FuzzExtractSymbolUpdatedDoesNotPanic(f *testing.F) {
	t0, ok := symbolUpdatedEvent.Topic0()
	require.True(f, ok)

	f.Add(t0[:], make([]byte, 32), make([]byte, 32), []byte{})
	f.Add(t0[:], make([]byte, 32), make([]byte, 32), make([]byte, 32))
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
			logs := []tx.Log{{Address: testTokens, Topics: topics, Data: data}}
			_, _ = ExtractSymbolUpdated(logs)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzExtractDecimalsUpdatedDoesNotPanic$' -fuzztime=10s ./standards/erc6909
func FuzzExtractDecimalsUpdatedDoesNotPanic(f *testing.F) {
	t0, ok := decimalsUpdatedEvent.Topic0()
	require.True(f, ok)

	f.Add(t0[:], make([]byte, 32), make([]byte, 32), []byte{})
	f.Add(t0[:], make([]byte, 32), make([]byte, 32), make([]byte, 32))
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
			logs := []tx.Log{{Address: testTokens, Topics: topics, Data: data}}
			_, _ = ExtractDecimalsUpdated(logs)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzExtractTransfersRoundTrip$' -fuzztime=10s ./standards/erc6909
func FuzzExtractTransfersRoundTrip(f *testing.F) {
	f.Add(
		[]byte{0x00}, []byte{0x11}, []byte{0x22},
		[]byte{0x2a}, []byte{0x07},
	)
	f.Add(
		make([]byte, 20), make([]byte, 20), make([]byte, 20),
		[]byte{}, []byte{},
	)

	f.Fuzz(func(t *testing.T, callerRaw, senderRaw, receiverRaw, idRaw, amountRaw []byte) {
		caller := types.NewAddressFromBytes(callerRaw)
		sender := types.NewAddressFromBytes(senderRaw)
		receiver := types.NewAddressFromBytes(receiverRaw)

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

		t0, ok := transferEvent.Topic0()
		require.True(t, ok)

		data, err := abi.Pack(abi.Types{abi.Address, abi.Uint256}, caller, amount)
		require.NoError(t, err)

		logs := []tx.Log{{
			Address: testTokens,
			Topics: []*types.Hash{
				types.NewHashFromBytes(t0[:]),
				types.NewHashFromBytes(sender.Bytes()),
				types.NewHashFromBytes(receiver.Bytes()),
				types.NewHashFromBytes(id.FillBytes(make([]byte, 32))),
			},
			Data: data,
		}}

		transfers, err := ExtractTransfers(logs)
		require.NoError(t, err)
		require.Len(t, transfers, 1)
		require.True(t, transfers[0].Contract.Equal(testTokens))
		require.True(t, transfers[0].Caller.Equal(caller))
		require.True(t, transfers[0].Sender.Equal(sender))
		require.True(t, transfers[0].Receiver.Equal(receiver))
		require.Zero(t, id.Cmp(transfers[0].Id))
		require.Zero(t, amount.Cmp(transfers[0].Amount))
	})
}

// go test -run '^$' -fuzz '^FuzzExtractApprovalsRoundTrip$' -fuzztime=10s ./standards/erc6909
func FuzzExtractApprovalsRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x11}, []byte{0x01}, []byte{0x64})
	f.Add(make([]byte, 20), make([]byte, 20), []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, ownerRaw, spenderRaw, idRaw, amountRaw []byte) {
		owner := types.NewAddressFromBytes(ownerRaw)
		spender := types.NewAddressFromBytes(spenderRaw)

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

		t0, ok := approvalEvent.Topic0()
		require.True(t, ok)

		data, err := abi.Pack(abi.Types{abi.Uint256}, amount)
		require.NoError(t, err)

		logs := []tx.Log{{
			Address: testTokens,
			Topics: []*types.Hash{
				types.NewHashFromBytes(t0[:]),
				types.NewHashFromBytes(owner.Bytes()),
				types.NewHashFromBytes(spender.Bytes()),
				types.NewHashFromBytes(id.FillBytes(make([]byte, 32))),
			},
			Data: data,
		}}

		approvals, err := ExtractApprovals(logs)
		require.NoError(t, err)
		require.Len(t, approvals, 1)
		require.True(t, approvals[0].Owner.Equal(owner))
		require.True(t, approvals[0].Spender.Equal(spender))
		require.Zero(t, id.Cmp(approvals[0].Id))
		require.Zero(t, amount.Cmp(approvals[0].Amount))
	})
}

// go test -run '^$' -fuzz '^FuzzExtractDecimalsUpdatedRoundTrip$' -fuzztime=10s ./standards/erc6909
func FuzzExtractDecimalsUpdatedRoundTrip(f *testing.F) {
	f.Add(uint8(0), []byte{0x01})
	f.Add(uint8(18), []byte{0x01})
	f.Add(uint8(255), []byte{0xff})

	f.Fuzz(func(t *testing.T, decimals uint8, idRaw []byte) {
		idClamped := idRaw
		if len(idClamped) > 32 {
			idClamped = idClamped[len(idClamped)-32:]
		}
		id := new(big.Int).SetBytes(idClamped)

		t0, ok := decimalsUpdatedEvent.Topic0()
		require.True(t, ok)

		data, err := abi.Pack(abi.Types{abi.Uint8}, decimals)
		require.NoError(t, err)

		logs := []tx.Log{{
			Address: testTokens,
			Topics: []*types.Hash{
				types.NewHashFromBytes(t0[:]),
				types.NewHashFromBytes(id.FillBytes(make([]byte, 32))),
			},
			Data: data,
		}}

		updates, err := ExtractDecimalsUpdated(logs)
		require.NoError(t, err)
		require.Len(t, updates, 1)
		require.Zero(t, id.Cmp(updates[0].Id))
		require.Equal(t, decimals, updates[0].NewDecimals)
	})
}

// go test -run '^$' -fuzz '^FuzzExtractNameUpdatedRoundTrip$' -fuzztime=10s ./standards/erc6909
func FuzzExtractNameUpdatedRoundTrip(f *testing.F) {
	f.Add("", []byte{0x01})
	f.Add("evmkit test token", []byte{0x01})
	f.Add("\x00\x01\xff unicode éè", []byte{0xff})

	f.Fuzz(func(t *testing.T, name string, idRaw []byte) {
		idClamped := idRaw
		if len(idClamped) > 32 {
			idClamped = idClamped[len(idClamped)-32:]
		}
		id := new(big.Int).SetBytes(idClamped)

		t0, ok := nameUpdatedEvent.Topic0()
		require.True(t, ok)

		data, err := abi.Pack(abi.Types{abi.String}, name)
		require.NoError(t, err)

		logs := []tx.Log{{
			Address: testTokens,
			Topics: []*types.Hash{
				types.NewHashFromBytes(t0[:]),
				types.NewHashFromBytes(id.FillBytes(make([]byte, 32))),
			},
			Data: data,
		}}

		updates, err := ExtractNameUpdated(logs)
		require.NoError(t, err)
		require.Len(t, updates, 1)
		require.Zero(t, id.Cmp(updates[0].Id))
		require.Equal(t, name, updates[0].NewName)
	})
}

// go test -run '^$' -fuzz '^FuzzSimpleDecodersDoNotPanic$' -fuzztime=10s ./standards/erc6909
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

	uint8Data, err := abi.Pack(abi.Types{abi.Uint8}, uint8(18))
	require.NoError(f, err)
	f.Add(uint8Data)

	strData, err := abi.Pack(abi.Types{abi.String}, "evmkit test token")
	require.NoError(f, err)
	f.Add(strData)

	f.Fuzz(func(t *testing.T, data []byte) {
		require.NotPanics(t, func() {
			_, _ = DecodeBalanceOf(data)
			_, _ = DecodeAllowance(data)
			_, _ = DecodeIsOperator(data)
			_, _ = DecodeName(data)
			_, _ = DecodeSymbol(data)
			_, _ = DecodeDecimals(data)
		})
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeBalanceOfRoundTrip$' -fuzztime=10s ./standards/erc6909
func FuzzDecodeBalanceOfRoundTrip(f *testing.F) {
	f.Add([]byte{0x01})
	f.Add([]byte{})
	f.Add(make([]byte, 32))

	f.Fuzz(func(t *testing.T, raw []byte) {
		clamped := raw
		if len(clamped) > 32 {
			clamped = clamped[len(clamped)-32:]
		}
		value := new(big.Int).SetBytes(clamped)

		data, err := abi.Pack(abi.Types{abi.Uint256}, value)
		require.NoError(t, err)

		got, err := DecodeBalanceOf(data)
		require.NoError(t, err)
		require.Zero(t, value.Cmp(got))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeAllowanceRoundTrip$' -fuzztime=10s ./standards/erc6909
func FuzzDecodeAllowanceRoundTrip(f *testing.F) {
	f.Add([]byte{0x01})
	f.Add([]byte{})
	f.Add(make([]byte, 32))

	f.Fuzz(func(t *testing.T, raw []byte) {
		clamped := raw
		if len(clamped) > 32 {
			clamped = clamped[len(clamped)-32:]
		}
		value := new(big.Int).SetBytes(clamped)

		data, err := abi.Pack(abi.Types{abi.Uint256}, value)
		require.NoError(t, err)

		got, err := DecodeAllowance(data)
		require.NoError(t, err)
		require.Zero(t, value.Cmp(got))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeDecimalsRoundTrip$' -fuzztime=10s ./standards/erc6909
func FuzzDecodeDecimalsRoundTrip(f *testing.F) {
	f.Add(uint8(0))
	f.Add(uint8(18))
	f.Add(uint8(255))

	f.Fuzz(func(t *testing.T, decimals uint8) {
		data, err := abi.Pack(abi.Types{abi.Uint8}, decimals)
		require.NoError(t, err)

		got, err := DecodeDecimals(data)
		require.NoError(t, err)
		require.Equal(t, decimals, got)
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeNameRoundTrip$' -fuzztime=10s ./standards/erc6909
func FuzzDecodeNameRoundTrip(f *testing.F) {
	f.Add("")
	f.Add("evmkit test token")
	f.Add("\x00\x01\xff unicode éè")

	f.Fuzz(func(t *testing.T, name string) {
		data, err := abi.Pack(abi.Types{abi.String}, name)
		require.NoError(t, err)

		got, err := DecodeName(data)
		require.NoError(t, err)
		require.Equal(t, name, got)
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeSymbolRoundTrip$' -fuzztime=10s ./standards/erc6909
func FuzzDecodeSymbolRoundTrip(f *testing.F) {
	f.Add("")
	f.Add("EVK6909")
	f.Add("\x00\x01\xff unicode éè")

	f.Fuzz(func(t *testing.T, symbol string) {
		data, err := abi.Pack(abi.Types{abi.String}, symbol)
		require.NoError(t, err)

		got, err := DecodeSymbol(data)
		require.NoError(t, err)
		require.Equal(t, symbol, got)
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeBalanceOfRoundTrip$' -fuzztime=10s ./standards/erc6909
func FuzzEncodeBalanceOfRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x01})
	f.Add(make([]byte, 20), []byte{})

	f.Fuzz(func(t *testing.T, ownerRaw, idRaw []byte) {
		owner := types.NewAddressFromBytes(ownerRaw)
		idClamped := idRaw
		if len(idClamped) > 32 {
			idClamped = idClamped[len(idClamped)-32:]
		}
		id := new(big.Int).SetBytes(idClamped)

		data, err := EncodeBalanceOf(owner, id)
		require.NoError(t, err)

		vals, err := balanceOfFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(owner))
		require.Zero(t, id.Cmp(vals[1].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeAllowanceRoundTrip$' -fuzztime=10s ./standards/erc6909
func FuzzEncodeAllowanceRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x11}, []byte{0x01})
	f.Add(make([]byte, 20), make([]byte, 20), []byte{})

	f.Fuzz(func(t *testing.T, ownerRaw, spenderRaw, idRaw []byte) {
		owner := types.NewAddressFromBytes(ownerRaw)
		spender := types.NewAddressFromBytes(spenderRaw)
		idClamped := idRaw
		if len(idClamped) > 32 {
			idClamped = idClamped[len(idClamped)-32:]
		}
		id := new(big.Int).SetBytes(idClamped)

		data, err := EncodeAllowance(owner, spender, id)
		require.NoError(t, err)

		vals, err := allowanceFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(owner))
		require.True(t, vals[1].(*types.Address).Equal(spender))
		require.Zero(t, id.Cmp(vals[2].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeIsOperatorRoundTrip$' -fuzztime=10s ./standards/erc6909
func FuzzEncodeIsOperatorRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x11})
	f.Add(make([]byte, 20), make([]byte, 20))

	f.Fuzz(func(t *testing.T, ownerRaw, spenderRaw []byte) {
		owner := types.NewAddressFromBytes(ownerRaw)
		spender := types.NewAddressFromBytes(spenderRaw)

		data, err := EncodeIsOperator(owner, spender)
		require.NoError(t, err)

		vals, err := isOperatorFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(owner))
		require.True(t, vals[1].(*types.Address).Equal(spender))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeApproveRoundTrip$' -fuzztime=10s ./standards/erc6909
func FuzzEncodeApproveRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x01}, []byte{0x64})
	f.Add(make([]byte, 20), []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, spenderRaw, idRaw, amountRaw []byte) {
		spender := types.NewAddressFromBytes(spenderRaw)

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

		data, err := EncodeApprove(spender, id, amount)
		require.NoError(t, err)

		vals, err := approveFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(spender))
		require.Zero(t, id.Cmp(vals[1].(*big.Int)))
		require.Zero(t, amount.Cmp(vals[2].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeSetOperatorRoundTrip$' -fuzztime=10s ./standards/erc6909
func FuzzEncodeSetOperatorRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, true)
	f.Add(make([]byte, 20), false)

	f.Fuzz(func(t *testing.T, spenderRaw []byte, approved bool) {
		spender := types.NewAddressFromBytes(spenderRaw)

		data, err := EncodeSetOperator(spender, approved)
		require.NoError(t, err)

		vals, err := setOperatorFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(spender))
		require.Equal(t, approved, vals[1].(bool))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeTransferRoundTrip$' -fuzztime=10s ./standards/erc6909
func FuzzEncodeTransferRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x01}, []byte{0x64})
	f.Add(make([]byte, 20), []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, receiverRaw, idRaw, amountRaw []byte) {
		receiver := types.NewAddressFromBytes(receiverRaw)

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

		data, err := EncodeTransfer(receiver, id, amount)
		require.NoError(t, err)

		vals, err := transferFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(receiver))
		require.Zero(t, id.Cmp(vals[1].(*big.Int)))
		require.Zero(t, amount.Cmp(vals[2].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeTransferFromRoundTrip$' -fuzztime=10s ./standards/erc6909
func FuzzEncodeTransferFromRoundTrip(f *testing.F) {
	f.Add([]byte{0x00}, []byte{0x11}, []byte{0x01}, []byte{0x64})
	f.Add(make([]byte, 20), make([]byte, 20), []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, senderRaw, receiverRaw, idRaw, amountRaw []byte) {
		sender := types.NewAddressFromBytes(senderRaw)
		receiver := types.NewAddressFromBytes(receiverRaw)

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

		data, err := EncodeTransferFrom(sender, receiver, id, amount)
		require.NoError(t, err)

		vals, err := transferFromFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(sender))
		require.True(t, vals[1].(*types.Address).Equal(receiver))
		require.Zero(t, id.Cmp(vals[2].(*big.Int)))
		require.Zero(t, amount.Cmp(vals[3].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeNameRoundTrip$' -fuzztime=10s ./standards/erc6909
func FuzzEncodeNameRoundTrip(f *testing.F) {
	f.Add([]byte{0x01})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, idRaw []byte) {
		idClamped := idRaw
		if len(idClamped) > 32 {
			idClamped = idClamped[len(idClamped)-32:]
		}
		id := new(big.Int).SetBytes(idClamped)

		data, err := EncodeName(id)
		require.NoError(t, err)

		vals, err := nameFn.DecodeCall(data)
		require.NoError(t, err)
		require.Zero(t, id.Cmp(vals[0].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeSymbolRoundTrip$' -fuzztime=10s ./standards/erc6909
func FuzzEncodeSymbolRoundTrip(f *testing.F) {
	f.Add([]byte{0x01})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, idRaw []byte) {
		idClamped := idRaw
		if len(idClamped) > 32 {
			idClamped = idClamped[len(idClamped)-32:]
		}
		id := new(big.Int).SetBytes(idClamped)

		data, err := EncodeSymbol(id)
		require.NoError(t, err)

		vals, err := symbolFn.DecodeCall(data)
		require.NoError(t, err)
		require.Zero(t, id.Cmp(vals[0].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeDecimalsRoundTrip$' -fuzztime=10s ./standards/erc6909
func FuzzEncodeDecimalsRoundTrip(f *testing.F) {
	f.Add([]byte{0x01})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, idRaw []byte) {
		idClamped := idRaw
		if len(idClamped) > 32 {
			idClamped = idClamped[len(idClamped)-32:]
		}
		id := new(big.Int).SetBytes(idClamped)

		data, err := EncodeDecimals(id)
		require.NoError(t, err)

		vals, err := decimalsFn.DecodeCall(data)
		require.NoError(t, err)
		require.Zero(t, id.Cmp(vals[0].(*big.Int)))
	})
}
