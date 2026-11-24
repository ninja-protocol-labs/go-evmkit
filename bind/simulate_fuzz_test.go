package bind

import (
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// go test -run '^$' -fuzz '^FuzzSimulateParams$' -fuzztime=10s ./bind
func FuzzSimulateParams(f *testing.F) {
	f.Add([]byte{}, false, []byte{}, []byte{}, "", false, false)
	f.Add(make([]byte, 20), false, []byte{0x01}, []byte{0xa9, 0x05, 0x9c, 0xbb}, rpc.BlockTagLatest, true, false)
	f.Add([]byte{0xff, 0xee}, true, make([]byte, 32), []byte{0x60, 0x80, 0x60, 0x40}, "0x1b4", false, true)
	f.Add(make([]byte, 40), false, []byte{0xff, 0xff}, make([]byte, 200), rpc.BlockTagPending, true, true)

	var payload map[string]any
	var gotBlock string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     string            `json:"id"`
			Params []json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Params) != 2 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		payload = nil
		_ = json.Unmarshal(req.Params[0], &payload)
		_ = json.Unmarshal(req.Params[1], &gotBlock)
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"` + req.ID + `","result":[{"calls":[` +
			`{"status":"0x1","gasUsed":"0x5208","returnData":"0x","logs":[]},` +
			`{"status":"0x0","gasUsed":"0x1","returnData":"0x","logs":[],"error":{"code":3,"message":"execution reverted"}}]}]}`))
	}))
	f.Cleanup(srv.Close)
	client := rpc.NewClient(srv.URL)

	f.Fuzz(func(t *testing.T, toBytes []byte, create bool, valueBytes, calldata []byte, block string, traceTransfers, validation bool) {
		if !utf8.ValidString(block) {
			t.Skip()
		}
		if len(valueBytes) > 32 {
			valueBytes = valueBytes[:32]
		}
		from := types.NewAddressFromBytes(core.Keccak256(toBytes).Bytes()[:20])
		value := new(big.Int).SetBytes(valueBytes)
		var to *types.Address
		if !create {
			to = types.NewAddressFromBytes(toBytes)
		}

		opts := NewSimulateOptions().WithBlock(block)
		if traceTransfers {
			opts.WithTraceTransfers()
		}
		if validation {
			opts.WithValidation()
		}

		results, err := Simulate(t.Context(), client, opts, []SimulateCall{
			NewSimulateCall(from, to, value, calldata),
			NewSimulateCall(from, to, nil, nil),
		})
		require.NoError(t, err)
		require.Len(t, results, 2)
		require.True(t, results[0].Success)
		require.Equal(t, uint64(21000), results[0].GasUsed)
		require.False(t, results[1].Success)
		require.Equal(t, 3, results[1].Error.Code)

		require.Equal(t, traceTransfers, payload["traceTransfers"])
		require.Equal(t, validation, payload["validation"])
		blockStateCalls, ok := payload["blockStateCalls"].([]any)
		require.True(t, ok)
		require.Len(t, blockStateCalls, 1)
		blockStateCall, ok := blockStateCalls[0].(map[string]any)
		require.True(t, ok)
		require.NotContains(t, blockStateCall, "stateOverrides")
		calls, ok := blockStateCall["calls"].([]any)
		require.True(t, ok)
		require.Len(t, calls, 2)

		want := map[string]any{
			"from":  from.String(),
			"data":  encoding.Hex.EncodePrefixed(calldata),
			"value": "0x" + value.Text(16),
		}
		if !create {
			want["to"] = to.String()
		}
		require.Equal(t, want, calls[0])
		bare := map[string]any{
			"from": from.String(),
			"data": "0x",
		}
		if !create {
			bare["to"] = to.String()
		}
		require.Equal(t, bare, calls[1])

		if block == "" {
			require.Equal(t, rpc.BlockTagLatest, gotBlock)
		} else {
			require.Equal(t, block, gotBlock)
		}
	})
}

// go test -run '^$' -fuzz '^FuzzSimulateArbitraryResult$' -fuzztime=10s ./bind
func FuzzSimulateArbitraryResult(f *testing.F) {
	f.Add(`[]`)
	f.Add(`null`)
	f.Add(`[{"calls":[{"status":"0x1","gasUsed":"0x5208","returnData":"0x","logs":[]}]}]`)
	f.Add(`[{"calls":[{"status":"0x0","gasUsed":"0x1","returnData":"0x","error":{"code":3,"message":"m","data":"0xfb8f41b2"}}]}]`)
	f.Add(`[{"calls":[{"status":"0x1","gasUsed":"0x1","returnData":"0x","logs":[{"address":"0x2222222222222222222222222222222222222222","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"],"data":"0x","blockNumber":"0x1","transactionHash":"0x6de9ecde85473347b57a96c8eb71f9dce601416b0d062ac367404fa3a8bd2f15","logIndex":"0x0"}]}]}]`)
	f.Add(`[{"calls":[{"status":"0x1","gasUsed":"0xffffffffffffffffff","returnData":"0x1","logs":[null],"error":{"code":"x"}}]}]`)
	f.Add(`[{"calls":"x"},{"calls":[]}]`)
	f.Add(`{"calls":[]}`)

	from, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(f, err)

	var reply string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"` + req.ID + `","result":` + reply + `}`))
	}))
	f.Cleanup(srv.Close)
	client := rpc.NewClient(srv.URL)

	f.Fuzz(func(t *testing.T, result string) {
		if !json.Valid([]byte(result)) {
			t.Skip()
		}
		reply = result

		results, err := Simulate(t.Context(), client, nil, []SimulateCall{NewSimulateCall(from, nil, nil, nil)})
		if err != nil {
			require.Nil(t, results)
			return
		}
		require.Len(t, results, 1)
	})
}
