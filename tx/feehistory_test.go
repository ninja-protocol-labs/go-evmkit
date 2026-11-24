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

func TestPackReadsFeesFromFeeHistoryAndMaxPriorityFee(t *testing.T) {
	from, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)
	to, err := types.NewAddressFromHex("0x2222222222222222222222222222222222222222")
	require.NoError(t, err)
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xa11ce))
	require.NoError(t, err)
	authorization := core.Authorization{ChainID: big.NewInt(1), Address: to}
	require.NoError(t, authorization.Sign(key))

	cases := []struct {
		name           string
		percentile     float64
		blocks         uint64
		multiplier     uint64
		rewards        []any
		minTip         string
		wantBlocksArg  string
		wantPercentile string
		wantTip        int64
		wantFeeCap     int64
	}{
		{"defaults", 0, 0, 0, []any{[]any{"0x1"}, []any{"0x5"}, []any{"0x3"}}, "0x1", "0xa", "[50]", 3, 130*2 + 3},
		{"explicit window", 75, 4, 0, []any{[]any{"0x9"}, []any{"0x1"}, []any{"0x6"}, []any{"0x4"}}, "0x1", "0x4", "[75]", 5, 130*2 + 5},
		{"price buffer scales next base fee", 50, 1, 150, []any{[]any{"0x2"}}, "0x0", "0x1", "[50]", 2, 195 + 2},
		{"zero tips raised to node minimum", 25, 2, 0, []any{[]any{"0x0"}, []any{"0x0"}}, "0x1", "0x2", "[25]", 1, 130*2 + 1},
		{"median below node minimum", 50, 3, 0, []any{[]any{"0x2"}, []any{"0x3"}, []any{"0x4"}}, "0xa", "0x3", "[50]", 10, 130*2 + 10},
		{"no rewards falls back to node minimum", 50, 1, 0, []any{}, "0x7", "0x1", "[50]", 7, 130*2 + 7},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var methods []string
			var feeHistoryParams []json.RawMessage
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var reqs []struct {
					ID     string            `json:"id"`
					Method string            `json:"method"`
					Params []json.RawMessage `json:"params"`
				}
				require.NoError(t, json.NewDecoder(r.Body).Decode(&reqs))
				methods = nil
				resps := make([]map[string]any, 0, len(reqs))
				for _, req := range reqs {
					methods = append(methods, req.Method)
					var result any
					switch req.Method {
					case rpc.MethodETHChainID:
						result = "0x1"
					case rpc.MethodETHGetTransactionCount:
						result = "0x5"
					case rpc.MethodETHFeeHistory:
						feeHistoryParams = req.Params
						result = map[string]any{"oldestBlock": "0x10", "baseFeePerGas": []any{"0x64", "0x6e", "0x82"}, "gasUsedRatio": []any{0.5}, "reward": tc.rewards}
					case rpc.MethodETHMaxPriorityFeePerGas:
						result = tc.minTip
					case rpc.MethodETHEstimateGas:
						result = "0x5208"
					}
					resps = append(resps, map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
				}
				require.NoError(t, json.NewEncoder(w).Encode(resps))
			}))
			defer srv.Close()
			cli := rpc.NewClient(srv.URL)

			for kind := range 2 {
				var packed core.Transaction
				switch kind {
				case 0:
					b := NewDynamicFeeBuilder()
					b.TipPercentile(tc.percentile, tc.blocks)
					b.PriceBuffer(tc.multiplier)
					packed, err = b.Build(from, to, big.NewInt(0), nil).Pack(t.Context(), cli)
				case 1:
					b := NewSetCodeBuilder([]core.Authorization{authorization})
					b.TipPercentile(tc.percentile, tc.blocks)
					b.PriceBuffer(tc.multiplier)
					packed, err = b.Build(from, to, big.NewInt(0), nil).Pack(t.Context(), cli)
				}
				require.NoError(t, err, kind)

				require.Contains(t, methods, rpc.MethodETHFeeHistory, kind)
				require.Contains(t, methods, rpc.MethodETHMaxPriorityFeePerGas, kind)
				require.NotContains(t, methods, rpc.MethodETHGetBlockByNumber, kind)
				require.Len(t, feeHistoryParams, 3, kind)
				require.JSONEq(t, `"`+tc.wantBlocksArg+`"`, string(feeHistoryParams[0]), kind)
				require.JSONEq(t, `"latest"`, string(feeHistoryParams[1]), kind)
				require.JSONEq(t, tc.wantPercentile, string(feeHistoryParams[2]), kind)

				var tip, feeCap *big.Int
				switch v := packed.(type) {
				case *core.DynamicFeeTx:
					tip, feeCap = v.GasTipCap, v.GasFeeCap
				case *core.SetCodeTx:
					tip, feeCap = v.GasTipCap, v.GasFeeCap
				default:
					t.Fatalf("unexpected tx type %T", packed)
				}
				require.Zero(t, big.NewInt(tc.wantTip).Cmp(tip), "kind %d tip %s", kind, tip)
				require.Zero(t, big.NewInt(tc.wantFeeCap).Cmp(feeCap), "kind %d feeCap %s", kind, feeCap)
			}
		})
	}
}

func TestPackSkipsFeeLookupsForExplicitPrice(t *testing.T) {
	from, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)
	to, err := types.NewAddressFromHex("0x2222222222222222222222222222222222222222")
	require.NoError(t, err)

	var methods []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqs []struct {
			ID     string `json:"id"`
			Method string `json:"method"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&reqs))
		methods = nil
		resps := make([]map[string]any, 0, len(reqs))
		for _, req := range reqs {
			methods = append(methods, req.Method)
			var result any
			switch req.Method {
			case rpc.MethodETHChainID:
				result = "0x1"
			case rpc.MethodETHGetTransactionCount:
				result = "0x5"
			case rpc.MethodETHFeeHistory:
				result = map[string]any{"baseFeePerGas": []any{"0x64", "0x82"}, "reward": []any{[]any{"0x9"}}}
			case rpc.MethodETHMaxPriorityFeePerGas:
				result = "0x1"
			case rpc.MethodETHEstimateGas:
				result = "0x5208"
			}
			resps = append(resps, map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
		}
		require.NoError(t, json.NewEncoder(w).Encode(resps))
	}))
	defer srv.Close()
	cli := rpc.NewClient(srv.URL)

	both := NewDynamicFeeBuilder()
	both.Price(big.NewInt(1000), big.NewInt(7))
	packed, err := both.Build(from, to, big.NewInt(0), nil).Pack(t.Context(), cli)
	require.NoError(t, err)
	require.NotContains(t, methods, rpc.MethodETHFeeHistory)
	require.NotContains(t, methods, rpc.MethodETHMaxPriorityFeePerGas)
	require.Zero(t, big.NewInt(7).Cmp(packed.(*core.DynamicFeeTx).GasTipCap))
	require.Zero(t, big.NewInt(1000).Cmp(packed.(*core.DynamicFeeTx).GasFeeCap))

	packed, err = NewDynamicFeeTxConfig(from, to, big.NewInt(0), nil).WithGasTipCap(big.NewInt(7)).Pack(t.Context(), cli)
	require.NoError(t, err)
	require.Contains(t, methods, rpc.MethodETHFeeHistory)
	require.NotContains(t, methods, rpc.MethodETHMaxPriorityFeePerGas)
	require.Zero(t, big.NewInt(7).Cmp(packed.(*core.DynamicFeeTx).GasTipCap))
	require.Zero(t, big.NewInt(130*2+7).Cmp(packed.(*core.DynamicFeeTx).GasFeeCap))

	packed, err = NewDynamicFeeTxConfig(from, to, big.NewInt(0), nil).WithGasFeeCap(big.NewInt(1000)).Pack(t.Context(), cli)
	require.NoError(t, err)
	require.Contains(t, methods, rpc.MethodETHFeeHistory)
	require.Contains(t, methods, rpc.MethodETHMaxPriorityFeePerGas)
	require.Zero(t, big.NewInt(9).Cmp(packed.(*core.DynamicFeeTx).GasTipCap))
	require.Zero(t, big.NewInt(1000).Cmp(packed.(*core.DynamicFeeTx).GasFeeCap))
}

func TestPackRejectsInvalidTipPercentileWithoutRPC(t *testing.T) {
	from, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)
	to, err := types.NewAddressFromHex("0x2222222222222222222222222222222222222222")
	require.NoError(t, err)

	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
	}))
	defer srv.Close()
	cli := rpc.NewClient(srv.URL)

	cases := []struct {
		percentile float64
		blocks     uint64
		wantErr    string
	}{
		{-1, 10, "tx: tipPercentile: percentile -1 is outside [0, 100]"},
		{100.5, 10, "tx: tipPercentile: percentile 100.5 is outside [0, 100]"},
		{50, 1025, "tx: tipPercentile: blocks 1025 exceeds 1024"},
	}
	for _, tc := range cases {
		b := NewDynamicFeeBuilder()
		b.TipPercentile(tc.percentile, tc.blocks)
		packed, err := b.Build(from, to, big.NewInt(0), nil).Pack(t.Context(), cli)
		require.Nil(t, packed)
		require.EqualError(t, err, tc.wantErr)

		s := NewSetCodeBuilder([]core.Authorization{{ChainID: big.NewInt(1), Address: to}})
		s.TipPercentile(tc.percentile, tc.blocks)
		s.Limit(100_000)
		packed, err = s.Build(from, to, big.NewInt(0), nil).Pack(t.Context(), cli)
		require.Nil(t, packed)
		require.EqualError(t, err, tc.wantErr)
	}
	require.Zero(t, hits.Load())
}

func TestResolveTipPercentile(t *testing.T) {
	cases := []struct {
		percentile     float64
		blocks         uint64
		wantPercentile float64
		wantBlocks     uint64
	}{
		{0, 0, 50, 10},
		{75, 0, 75, 10},
		{0, 20, 50, 20},
		{100, 1024, 100, 1024},
		{0.5, 1, 0.5, 1},
	}
	for _, tc := range cases {
		percentile, blocks, err := resolveTipPercentile(tc.percentile, tc.blocks)
		require.NoError(t, err)
		require.Equal(t, tc.wantPercentile, percentile)
		require.Equal(t, tc.wantBlocks, blocks)
	}
}

func TestParseFeeHistory(t *testing.T) {
	medians := []struct {
		name    string
		rewards []any
		want    int64
	}{
		{"single", []any{[]any{"0x7"}}, 7},
		{"odd", []any{[]any{"0x1"}, []any{"0x9"}, []any{"0x5"}}, 5},
		{"even", []any{[]any{"0x9"}, []any{"0x1"}, []any{"0x6"}, []any{"0x4"}}, 5},
		{"spike ignored", []any{[]any{"0x2"}, []any{"0x3"}, []any{"0x3e8"}, []any{"0x2"}, []any{"0x3"}}, 3},
	}
	for _, tc := range medians {
		t.Run(tc.name, func(t *testing.T) {
			tip, nextBaseFee, err := parseFeeHistory(map[string]any{"baseFeePerGas": []any{"0x1", "0x2a"}, "reward": tc.rewards})
			require.NoError(t, err)
			require.Zero(t, big.NewInt(tc.want).Cmp(tip), "tip %s", tip)
			require.Zero(t, big.NewInt(42).Cmp(nextBaseFee))
		})
	}

	tip, nextBaseFee, err := parseFeeHistory(map[string]any{"baseFeePerGas": []any{"0x2a"}})
	require.NoError(t, err)
	require.Nil(t, tip)
	require.Zero(t, big.NewInt(42).Cmp(nextBaseFee))

	malformed := []struct {
		name    string
		history map[string]any
		wantErr string
	}{
		{"nil", nil, "tx: feeHistory: chain does not report baseFeePerGas"},
		{"no base fees", map[string]any{"reward": []any{[]any{"0x1"}}}, "tx: feeHistory: chain does not report baseFeePerGas"},
		{"empty base fees", map[string]any{"baseFeePerGas": []any{}}, "tx: feeHistory: chain does not report baseFeePerGas"},
		{"non-string base fee", map[string]any{"baseFeePerGas": []any{1}}, "tx: feeHistory: baseFeePerGas[0]: unexpected type int"},
		{"bad next base fee", map[string]any{"baseFeePerGas": []any{"0x1", "0xzz"}}, "tx: feeHistory: baseFeePerGas[1]:"},
		{"two percentiles", map[string]any{"baseFeePerGas": []any{"0x1"}, "reward": []any{[]any{"0x1", "0x2"}}}, "tx: feeHistory: reward[0]: expected one percentile"},
		{"non-string tip", map[string]any{"baseFeePerGas": []any{"0x1"}, "reward": []any{[]any{1}}}, "tx: feeHistory: reward[0][0]: unexpected type int"},
		{"bad tip", map[string]any{"baseFeePerGas": []any{"0x1"}, "reward": []any{[]any{"0x1"}, []any{"nope"}}}, "tx: feeHistory: reward[1][0]:"},
	}
	for _, tc := range malformed {
		t.Run(tc.name, func(t *testing.T) {
			tip, nextBaseFee, err := parseFeeHistory(tc.history)
			require.Nil(t, tip)
			require.Nil(t, nextBaseFee)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}
