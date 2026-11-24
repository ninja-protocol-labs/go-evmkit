package tx

import (
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

func TestBuildersPassStateOverrideToEstimateGas(t *testing.T) {
	from, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)
	to, err := types.NewAddressFromHex("0x2222222222222222222222222222222222222222")
	require.NoError(t, err)
	var slot, value [32]byte
	slot[31] = 0x01
	value[31] = 0x2a
	override := rpc.NewStateOverride().
		WithBalance(from, big.NewInt(1_000_000)).
		WithStateDiff(to, slot, value)
	wantOverride, err := json.Marshal(override)
	require.NoError(t, err)

	var hits atomic.Int32
	var estimateParams []json.RawMessage
	var estimateSeen bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		var reqs []struct {
			ID     string            `json:"id"`
			Method string            `json:"method"`
			Params []json.RawMessage `json:"params"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&reqs))
		resps := make([]map[string]any, 0, len(reqs))
		for _, req := range reqs {
			var result any
			switch req.Method {
			case rpc.MethodETHChainID:
				result = "0x1"
			case rpc.MethodETHGetTransactionCount:
				result = "0x5"
			case rpc.MethodETHGasPrice:
				result = "0x3b9aca00"
			case rpc.MethodETHMaxPriorityFeePerGas:
				result = "0x1"
			case rpc.MethodETHGetBlockByNumber:
				result = map[string]any{"baseFeePerGas": "0x7"}
			case rpc.MethodETHEstimateGas:
				estimateSeen = true
				estimateParams = req.Params
				result = "0x5208"
			}
			resps = append(resps, map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
		}
		require.NoError(t, json.NewEncoder(w).Encode(resps))
	}))
	defer srv.Close()
	cli := rpc.NewClient(srv.URL)
	authorizations := []core.Authorization{{ChainID: big.NewInt(1), Address: to}}

	for kind := range 3 {
		for scenario := range 3 {
			var b Builder
			switch kind {
			case 0:
				b = NewLegacyBuilder()
			case 1:
				b = NewDynamicFeeBuilder()
			case 2:
				b = NewSetCodeBuilder(authorizations)
			}
			switch scenario {
			case 0:
				b.StateOverride(override)
			case 2:
				b.StateOverride(override)
				b.Limit(90_000)
			}

			estimateSeen = false
			estimateParams = nil
			packed, err := b.Build(from, to, big.NewInt(0), []byte{0x01}).Pack(t.Context(), cli)
			require.NoError(t, err, "kind %d scenario %d", kind, scenario)

			switch scenario {
			case 0:
				require.True(t, estimateSeen, "kind %d", kind)
				require.Len(t, estimateParams, 3, "kind %d", kind)
				require.JSONEq(t, string(wantOverride), string(estimateParams[2]), "kind %d", kind)
				require.GreaterOrEqual(t, packed.Gas(), uint64(21_000), "kind %d", kind)
			case 1:
				require.True(t, estimateSeen, "kind %d", kind)
				require.Len(t, estimateParams, 2, "kind %d", kind)
			case 2:
				require.False(t, estimateSeen, "kind %d", kind)
				require.Equal(t, uint64(90_000), packed.Gas(), "kind %d", kind)
			}
		}
	}
	require.Equal(t, int32(9), hits.Load())
}

func TestBuildersRejectInvalidStateOverrideWithoutRPC(t *testing.T) {
	from, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)
	to, err := types.NewAddressFromHex("0x2222222222222222222222222222222222222222")
	require.NoError(t, err)
	var slot, value [32]byte

	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
	}))
	defer srv.Close()
	cli := rpc.NewClient(srv.URL)
	authorizations := []core.Authorization{{ChainID: big.NewInt(1), Address: to}}

	invalid := []*rpc.StateOverride{
		rpc.NewStateOverride().WithBalance(nil, big.NewInt(1)),
		rpc.NewStateOverride().WithBalance(from, big.NewInt(-1)),
		rpc.NewStateOverride().WithBalance(from, new(big.Int).Lsh(big.NewInt(1), 256)),
		rpc.NewStateOverride().WithState(to, slot, value).WithStateDiff(to, slot, value),
	}
	for kind := range 3 {
		for i, override := range invalid {
			var b Builder
			switch kind {
			case 0:
				b = NewLegacyBuilder()
			case 1:
				b = NewDynamicFeeBuilder()
			case 2:
				b = NewSetCodeBuilder(authorizations)
			}
			b.StateOverride(override)
			_, err := b.Build(from, to, big.NewInt(0), nil).Pack(t.Context(), cli)
			require.ErrorIs(t, err, rpc.ErrInvalidStateOverride, "kind %d case %d", kind, i)
			require.ErrorContains(t, err, "tx: gasLimit:", "kind %d case %d", kind, i)
		}
	}
	require.Zero(t, hits.Load())
}
