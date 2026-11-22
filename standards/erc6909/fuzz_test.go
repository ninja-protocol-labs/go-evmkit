package erc6909

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

		data, err := abi.Pack(abi.NewTypes(abi.Address, abi.Uint256), caller, amount)
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

		data, err := abi.Pack(abi.NewTypes(abi.Uint256), amount)
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

		data, err := abi.Pack(abi.NewTypes(abi.Uint8), decimals)
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

		data, err := abi.Pack(abi.NewTypes(abi.String), name)
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

	uintData, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(123456789))
	require.NoError(f, err)
	f.Add(uintData)

	boolData, err := abi.Pack(abi.NewTypes(abi.Bool), true)
	require.NoError(f, err)
	f.Add(boolData)

	uint8Data, err := abi.Pack(abi.NewTypes(abi.Uint8), uint8(18))
	require.NoError(f, err)
	f.Add(uint8Data)

	strData, err := abi.Pack(abi.NewTypes(abi.String), "evmkit test token")
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

		data, err := abi.Pack(abi.NewTypes(abi.Uint256), value)
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

		data, err := abi.Pack(abi.NewTypes(abi.Uint256), value)
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
		data, err := abi.Pack(abi.NewTypes(abi.Uint8), decimals)
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
		data, err := abi.Pack(abi.NewTypes(abi.String), name)
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
		data, err := abi.Pack(abi.NewTypes(abi.String), symbol)
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

// go test -run '^$' -fuzz '^FuzzTransferFromRequestValidate$' -fuzztime=10s ./standards/erc6909
func FuzzTransferFromRequestValidate(f *testing.F) {
	f.Add(uint8(0))
	f.Add(uint8(0x1f))
	f.Add(uint8(0x20))
	f.Add(uint8(0x0f))

	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	f.Fuzz(func(t *testing.T, mask uint8) {
		var req *TransferFromRequest
		if mask&0x20 == 0 {
			req = &TransferFromRequest{}
			if mask&0x01 != 0 {
				req.Contract = contract
			}
			if mask&0x02 != 0 {
				req.Sender = a
			}
			if mask&0x04 != 0 {
				req.Receiver = a
			}
			if mask&0x08 != 0 {
				req.ID = big.NewInt(int64(mask))
			}
			if mask&0x10 != 0 {
				req.Amount = big.NewInt(0)
			}
		}

		want := ""
		switch {
		case req == nil:
			want = "invalid request: request is nil"
		case req.Contract == nil:
			want = "invalid request: contract is nil"
		case req.Sender == nil:
			want = "invalid request: sender is nil"
		case req.Receiver == nil:
			want = "invalid request: receiver is nil"
		case req.ID == nil:
			want = "invalid request: id is nil"
		case req.Amount == nil:
			want = "invalid request: amount is nil"
		}

		err := req.Validate()
		if want == "" {
			require.NoError(t, err)
			return
		}
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.EqualError(t, err, want)
	})
}

// go test -run '^$' -fuzz '^FuzzTransferFromThroughClient$' -fuzztime=10s ./standards/erc6909
func FuzzTransferFromThroughClient(f *testing.F) {
	f.Add([]byte{}, []byte{}, []byte{}, []byte{}, uint64(0), true)
	f.Add([]byte{0x01}, []byte{0x02}, []byte{0x07}, []byte{0x01, 0x00}, uint64(42), false)
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 32), make([]byte, 32), ^uint64(0)-1, true)

	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(f, err)
	signer := core.PubkeyToAddress(key.PublicKey())
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")

	f.Fuzz(func(t *testing.T, senderRaw, receiverRaw, idRaw, amountRaw []byte, nonce uint64, dynamic bool) {
		if len(idRaw) > 32 {
			idRaw = idRaw[len(idRaw)-32:]
		}
		if len(amountRaw) > 32 {
			amountRaw = amountRaw[len(amountRaw)-32:]
		}
		sender := types.NewAddressFromBytes(senderRaw)
		receiver := types.NewAddressFromBytes(receiverRaw)
		id := new(big.Int).SetBytes(idRaw)
		amount := new(big.Int).SetBytes(amountRaw)

		calldata, err := EncodeTransferFrom(sender, receiver, id, amount)
		require.NoError(t, err)

		var builder tx.Builder
		if dynamic {
			builder = tx.NewDynamicFeeBuilder()
		} else {
			builder = tx.NewLegacyBuilder()
		}
		builder.ChainID(big.NewInt(1))
		builder.Nonce(nonce)
		builder.Limit(120_000)
		builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

		var raw []byte
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			raw, err = encoding.Hex.Decode(e.Params.([]any)[0].(string))
			require.NoError(t, err)
			*e.Result.(*string) = core.Keccak256(raw).String()
			return nil
		}).Once()

		signed, hash, err := New(c).TransferFrom(t.Context(), bind.NewTransactOptions(key, builder), NewTransferFromRequest(contract, sender, receiver, id, amount))
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
		require.True(t, gotTo.Equal(contract))
		require.Equal(t, calldata, gotData)
	})
}

// go test -run '^$' -fuzz '^FuzzReadsSurviveArbitraryResult$' -fuzztime=10s ./standards/erc6909
func FuzzReadsSurviveArbitraryResult(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 96))

	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	f.Fuzz(func(t *testing.T, result []byte) {
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(result)
			return nil
		}).Times(6)
		e := New(c)

		n, err := e.BalanceOf(t.Context(), nil, NewBalanceOfRequest(contract, a, big.NewInt(1)))
		require.True(t, (err == nil) != (n == nil))
		n, err = e.Allowance(t.Context(), nil, NewAllowanceRequest(contract, a, a, big.NewInt(1)))
		require.True(t, (err == nil) != (n == nil))
		_, _ = e.IsOperator(t.Context(), nil, NewIsOperatorRequest(contract, a, a))
		_, _ = e.Name(t.Context(), nil, NewNameRequest(contract, big.NewInt(1)))
		_, _ = e.Decimals(t.Context(), nil, NewDecimalsRequest(contract, big.NewInt(1)))
		_, _ = e.CallTransferFrom(t.Context(), nil, NewTransferFromRequest(contract, a, a, big.NewInt(1), big.NewInt(1)))
	})
}
