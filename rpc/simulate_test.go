package rpc

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestETHSimulateV1Params(t *testing.T) {
	payload := map[string]any{"blockStateCalls": []any{}}
	var result []map[string]any

	e := ETHSimulateV1("1", payload, "0x10", &result)
	require.Equal(t, "1", e.ID)
	require.Equal(t, MethodETHSimulateV1, e.Method)
	require.Equal(t, "eth_simulateV1", e.Method)
	require.Equal(t, []any{payload, "0x10"}, e.Params)
	require.Same(t, &result, e.Result)
}

func TestClientSendsSimulateV1PayloadAndDecodesBlocks(t *testing.T) {
	owner, err := types.NewAddressFromHex("0x8888888888888888888888888888888888888888")
	require.NoError(t, err)
	token, err := types.NewAddressFromHex("0x9999999999999999999999999999999999999999")
	require.NoError(t, err)
	var slot, value [32]byte
	slot[31] = 0x01
	value[31] = 0x2a
	override := NewStateOverride().WithCode(token, []byte{0x00}).WithStateDiff(token, slot, value)

	var gotMethod string
	var gotParams []json.RawMessage
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     string            `json:"id"`
			Method string            `json:"method"`
			Params []json.RawMessage `json:"params"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		gotMethod = req.Method
		gotParams = req.Params
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"` + req.ID + `","result":[{"number":"0x11","calls":[` +
			`{"status":"0x1","gasUsed":"0xb4a8","returnData":"0x0000000000000000000000000000000000000000000000000000000000000001","logs":[{"address":"0x9999999999999999999999999999999999999999","topics":["0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"],"data":"0x"}]},` +
			`{"status":"0x0","gasUsed":"0x5f6a","returnData":"0x","logs":[],"error":{"code":3,"message":"execution reverted","data":"0xfb8f41b2"}}` +
			`]}]}`))
	}))
	defer srv.Close()

	payload := map[string]any{
		"blockStateCalls": []any{
			map[string]any{
				"stateOverrides": override,
				"calls": []any{
					map[string]any{"from": owner.String(), "to": token.String(), "data": "0x095ea7b3"},
					map[string]any{"from": owner.String(), "to": token.String(), "data": "0x23b872dd"},
				},
			},
		},
		"traceTransfers": true,
		"validation":     false,
	}
	var result []map[string]any
	require.NoError(t, NewClient(srv.URL).Call(t.Context(), ETHSimulateV1("sim", payload, BlockTagLatest, &result)))

	require.Equal(t, MethodETHSimulateV1, gotMethod)
	require.Len(t, gotParams, 2)
	require.JSONEq(t, `{
		"blockStateCalls": [{
			"stateOverrides": {
				"0x9999999999999999999999999999999999999999": {
					"code": "0x00",
					"stateDiff": {"0x0000000000000000000000000000000000000000000000000000000000000001": "0x000000000000000000000000000000000000000000000000000000000000002a"}
				}
			},
			"calls": [
				{"from": "0x8888888888888888888888888888888888888888", "to": "0x9999999999999999999999999999999999999999", "data": "0x095ea7b3"},
				{"from": "0x8888888888888888888888888888888888888888", "to": "0x9999999999999999999999999999999999999999", "data": "0x23b872dd"}
			]
		}],
		"traceTransfers": true,
		"validation": false
	}`, string(gotParams[0]))
	var gotBlock string
	require.NoError(t, json.Unmarshal(gotParams[1], &gotBlock))
	require.Equal(t, BlockTagLatest, gotBlock)

	require.Len(t, result, 1)
	require.Equal(t, "0x11", result[0]["number"])
	calls, ok := result[0]["calls"].([]any)
	require.True(t, ok)
	require.Len(t, calls, 2)
	first, ok := calls[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "0x1", first["status"])
	require.Equal(t, "0xb4a8", first["gasUsed"])
	require.Len(t, first["logs"], 1)
	second, ok := calls[1].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "0x0", second["status"])
	callErr, ok := second["error"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "execution reverted", callErr["message"])
	require.Equal(t, "0xfb8f41b2", callErr["data"])
}
