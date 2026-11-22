package multicall3

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// go test -run '^$' -fuzz '^FuzzDecodeResults$' -fuzztime=10s ./standards/multicall3
func FuzzDecodeResults(f *testing.F) {
	for _, s := range []string{resultsTwo, resultsEmpty} {
		b, err := hex.DecodeString(s)
		require.NoError(f, err)
		f.Add(b)
	}
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		results, err := DecodeResults(data)
		if err != nil {
			return
		}

		tuples := make([][]any, len(results))
		for i, r := range results {
			tuples[i] = []any{r.Success, r.ReturnData}
		}
		encoded, err := aggregate3.EncodeReturn(tuples)
		require.NoError(t, err)

		again, err := DecodeResults(encoded)
		require.NoError(t, err)
		require.Equal(t, results, again)
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeAggregate3$' -fuzztime=10s ./standards/multicall3
func FuzzEncodeAggregate3(f *testing.F) {
	f.Add([]byte{0x11}, true, []byte{0x70, 0xa0, 0x82, 0x31}, []byte{0x22}, false, []byte{})
	f.Add([]byte{}, false, []byte{}, []byte{}, false, []byte{})

	f.Fuzz(func(t *testing.T, target1 []byte, allow1 bool, data1 []byte, target2 []byte, allow2 bool, data2 []byte) {
		calls := []Call3{
			NewCall3(types.NewAddressFromBytes(target1), allow1, data1),
			NewCall3(types.NewAddressFromBytes(target2), allow2, data2),
		}

		encoded, err := EncodeAggregate3(calls)
		require.NoError(t, err)

		vals, err := aggregate3.DecodeCall(encoded)
		require.NoError(t, err)

		elems := vals[0].([]any)
		require.Len(t, elems, len(calls))
		for i, e := range elems {
			fields := e.([]any)
			require.True(t, calls[i].Target.Equal(fields[0].(*types.Address)))
			require.Equal(t, calls[i].AllowFailure, fields[1])
			require.Equal(t, append([]byte{}, calls[i].CallData...), fields[2])
		}
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeAggregate3ValueRoundTrip$' -fuzztime=10s ./standards/multicall3
func FuzzEncodeAggregate3ValueRoundTrip(f *testing.F) {
	f.Add([]byte{0x11}, true, []byte{0x01}, []byte{0x70, 0xa0, 0x82, 0x31})
	f.Add([]byte{}, false, []byte{}, []byte{})
	f.Add(make([]byte, 20), true, make([]byte, 32), make([]byte, 100))

	f.Fuzz(func(t *testing.T, targetRaw []byte, allowFailure bool, valueRaw, callData []byte) {
		if len(valueRaw) > 32 {
			valueRaw = valueRaw[len(valueRaw)-32:]
		}
		call := NewCall3Value(types.NewAddressFromBytes(targetRaw), allowFailure, new(big.Int).SetBytes(valueRaw), callData)

		encoded, err := EncodeAggregate3Value([]Call3Value{call})
		require.NoError(t, err)

		vals, err := aggregate3Value.DecodeCall(encoded)
		require.NoError(t, err)

		elems := vals[0].([]any)
		require.Len(t, elems, 1)
		fields := elems[0].([]any)
		require.True(t, call.Target.Equal(fields[0].(*types.Address)))
		require.Equal(t, allowFailure, fields[1])
		require.Equal(t, 0, call.Value.Cmp(fields[2].(*big.Int)))
		require.Equal(t, append([]byte{}, callData...), fields[3])
	})
}

// go test -run '^$' -fuzz '^FuzzDecodersDoNotPanic$' -fuzztime=10s ./standards/multicall3
func FuzzDecodersDoNotPanic(f *testing.F) {
	for _, s := range []string{retAggregate, retBlockAndAggregate, retUint256, retBytes32, retAddress, resultsTwo, resultsEmpty} {
		b, err := hex.DecodeString(s)
		require.NoError(f, err)
		f.Add(b)
	}
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeResults(data)
		_, _ = DecodeAggregate(data)
		_, _ = DecodeBlockAndAggregate(data)
		_, _ = DecodeGetBasefee(data)
		_, _ = DecodeGetBlockHash(data)
		_, _ = DecodeGetBlockNumber(data)
		_, _ = DecodeGetChainID(data)
		_, _ = DecodeGetCurrentBlockCoinbase(data)
		_, _ = DecodeGetCurrentBlockDifficulty(data)
		_, _ = DecodeGetCurrentBlockGasLimit(data)
		_, _ = DecodeGetCurrentBlockTimestamp(data)
		_, _ = DecodeGetEthBalance(data)
		_, _ = DecodeGetLastBlockHash(data)
	})
}

// go test -run '^$' -fuzz '^FuzzAggregate3ValueRequestValidate$' -fuzztime=10s ./standards/multicall3
func FuzzAggregate3ValueRequestValidate(f *testing.F) {
	f.Add(uint8(0), int64(1), int64(1))
	f.Add(uint8(0x7f), int64(0), int64(0))
	f.Add(uint8(0x80), int64(1), int64(1))
	f.Add(uint8(0x3f), int64(-1), int64(5))
	f.Add(uint8(0x7f), int64(5), int64(-1))

	contract := types.MustNewAddressFromHex("0xcA11bde05977b3631167028862bE2a173976CA11")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	f.Fuzz(func(t *testing.T, mask uint8, value0, value1 int64) {
		var req *Aggregate3ValueRequest
		if mask&0x80 == 0 {
			req = &Aggregate3ValueRequest{}
			if mask&0x01 != 0 {
				req.Contract = contract
			}
			if mask&0x02 != 0 {
				call := Call3Value{}
				if mask&0x04 != 0 {
					call.Target = a
				}
				if mask&0x08 != 0 {
					call.Value = big.NewInt(value0)
				}
				req.Calls = append(req.Calls, call)
			}
			if mask&0x10 != 0 {
				call := Call3Value{}
				if mask&0x20 != 0 {
					call.Target = a
				}
				if mask&0x40 != 0 {
					call.Value = big.NewInt(value1)
				}
				req.Calls = append(req.Calls, call)
			}
		}

		want := ""
		switch {
		case req == nil:
			want = "invalid request: request is nil"
		case req.Contract == nil:
			want = "invalid request: contract is nil"
		case len(req.Calls) == 0:
			want = "invalid request: calls is empty"
		default:
			for i, c := range req.Calls {
				if c.Target == nil {
					want = fmt.Sprintf("invalid request: calls[%d].target is nil", i)
					break
				}
				if c.Value == nil {
					want = fmt.Sprintf("invalid request: calls[%d].value is nil", i)
					break
				}
				if c.Value.Sign() < 0 {
					want = fmt.Sprintf("invalid request: calls[%d].value is negative", i)
					break
				}
			}
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

// go test -run '^$' -fuzz '^FuzzAggregate3ValueThroughClient$' -fuzztime=10s ./standards/multicall3
func FuzzAggregate3ValueThroughClient(f *testing.F) {
	f.Add([]byte{}, []byte{}, []byte{}, []byte{}, []byte{}, uint64(0), true)
	f.Add([]byte{0x01}, []byte{0x02}, []byte{0x07}, []byte{0x01, 0x00}, []byte{0xde, 0xad}, uint64(42), false)
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 31), make([]byte, 31), make([]byte, 64), ^uint64(0)-1, true)

	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(f, err)
	signer := core.PubkeyToAddress(key.PublicKey())
	contract := types.MustNewAddressFromHex("0xcA11bde05977b3631167028862bE2a173976CA11")

	f.Fuzz(func(t *testing.T, target0Raw, target1Raw, value0Raw, value1Raw, callData []byte, nonce uint64, dynamic bool) {
		if len(value0Raw) > 31 {
			value0Raw = value0Raw[len(value0Raw)-31:]
		}
		if len(value1Raw) > 31 {
			value1Raw = value1Raw[len(value1Raw)-31:]
		}
		value0 := new(big.Int).SetBytes(value0Raw)
		value1 := new(big.Int).SetBytes(value1Raw)
		calls := []Call3Value{
			NewCall3Value(types.NewAddressFromBytes(target0Raw), true, value0, callData),
			NewCall3Value(types.NewAddressFromBytes(target1Raw), false, value1, nil),
		}
		total := new(big.Int).Add(value0, value1)

		calldata, err := EncodeAggregate3Value(calls)
		require.NoError(t, err)

		var builder tx.Builder
		if dynamic {
			builder = tx.NewDynamicFeeBuilder()
		} else {
			builder = tx.NewLegacyBuilder()
		}
		builder.ChainID(big.NewInt(1))
		builder.Nonce(nonce)
		builder.Limit(300_000)
		builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

		var raw []byte
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			raw, err = encoding.Hex.Decode(e.Params.([]any)[0].(string))
			require.NoError(t, err)
			*e.Result.(*string) = core.Keccak256(raw).String()
			return nil
		}).Once()

		opts := bind.NewTransactOptions(key, builder)
		signed, hash, err := New(c).Aggregate3Value(t.Context(), opts, NewAggregate3ValueRequest(contract, calls))
		require.NoError(t, err)
		require.True(t, hash.Equal(core.Keccak256(raw)))
		require.Equal(t, nonce, signed.AddressNonce())
		require.Nil(t, opts.Value)

		decoded, err := core.DecodeTransaction(raw)
		require.NoError(t, err)
		from, err := decoded.Sender()
		require.NoError(t, err)
		require.True(t, from.Equal(signer))

		var gotTo *types.Address
		var gotValue *big.Int
		var gotData []byte
		switch v := decoded.(type) {
		case *core.DynamicFeeTx:
			require.True(t, dynamic)
			gotTo, gotValue, gotData = v.To, v.Value, v.Data
		case *core.LegacyTx:
			require.False(t, dynamic)
			gotTo, gotValue, gotData = v.To, v.Value, v.Data
		default:
			t.Fatalf("unexpected tx type %T", decoded)
		}
		require.True(t, gotTo.Equal(contract))
		require.Equal(t, total.String(), gotValue.String())
		require.Equal(t, calldata, gotData)
	})
}

// go test -run '^$' -fuzz '^FuzzReadsSurviveArbitraryResult$' -fuzztime=10s ./standards/multicall3
func FuzzReadsSurviveArbitraryResult(f *testing.F) {
	for _, s := range []string{retAggregate, retBlockAndAggregate, retUint256, resultsTwo} {
		b, err := hex.DecodeString(s)
		require.NoError(f, err)
		f.Add(b)
	}
	f.Add([]byte{})
	f.Add(make([]byte, 31))

	contract := types.MustNewAddressFromHex("0xcA11bde05977b3631167028862bE2a173976CA11")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	f.Fuzz(func(t *testing.T, result []byte) {
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(result)
			return nil
		}).Times(8)
		m := New(c)

		agg, err := m.CallAggregate(t.Context(), nil, NewAggregateRequest(contract, []Call{NewCall(a, nil)}))
		require.True(t, (err == nil) != (agg == nil))
		block, err := m.CallTryBlockAndAggregate(t.Context(), nil, NewTryBlockAndAggregateRequest(contract, false, []Call{NewCall(a, nil)}))
		require.True(t, (err == nil) != (block == nil))
		results, err := m.CallAggregate3(t.Context(), nil, NewAggregate3Request(contract, []Call3{NewCall3(a, true, nil)}))
		require.True(t, err == nil || results == nil)
		results, err = m.CallAggregate3Value(t.Context(), nil, NewAggregate3ValueRequest(contract, []Call3Value{NewCall3Value(a, true, big.NewInt(1), nil)}))
		require.True(t, err == nil || results == nil)
		n, err := m.GetEthBalance(t.Context(), nil, NewGetEthBalanceRequest(contract, a))
		require.True(t, (err == nil) != (n == nil))
		hash, err := m.GetBlockHash(t.Context(), nil, NewGetBlockHashRequest(contract, big.NewInt(1)))
		require.True(t, (err == nil) != (hash == nil))
		coinbase, err := m.GetCurrentBlockCoinbase(t.Context(), nil, NewGetCurrentBlockCoinbaseRequest(contract))
		require.True(t, (err == nil) != (coinbase == nil))
		_, _ = m.GetBasefee(t.Context(), nil, NewGetBasefeeRequest(contract))
	})
}
