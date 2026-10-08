package multicall3

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

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

		encoded, err := aggregate3.EncodeReturn(toAny(results))
		require.NoError(t, err)

		again, err := DecodeResults(encoded)
		require.NoError(t, err)
		require.Equal(t, results, again)
	})
}

func FuzzEncodeAggregate3(f *testing.F) {
	f.Add([]byte{0x11}, true, []byte{0x70, 0xa0, 0x82, 0x31}, []byte{0x22}, false, []byte{})
	f.Add([]byte{}, false, []byte{}, []byte{}, false, []byte{})

	f.Fuzz(func(t *testing.T, target1 []byte, allow1 bool, data1 []byte, target2 []byte, allow2 bool, data2 []byte) {
		calls := []Call3{
			{Target: types.NewAddressFromBytes(target1), AllowFailure: allow1, CallData: data1},
			{Target: types.NewAddressFromBytes(target2), AllowFailure: allow2, CallData: data2},
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

func toAny(results []Result) [][]any {
	out := make([][]any, len(results))
	for i, r := range results {
		out[i] = []any{r.Success, r.ReturnData}
	}
	return out
}

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
		_, _ = DecodeGetCurrentBlockCoinbase(data)
		_, _ = DecodeGetLastBlockHash(data)
	})
}
