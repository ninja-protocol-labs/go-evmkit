package bind

import (
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

func TestNewSimulateOptions(t *testing.T) {
	opts := NewSimulateOptions()
	require.Empty(t, opts.Block)
	require.Nil(t, opts.StateOverride)
	require.False(t, opts.TraceTransfers)
	require.False(t, opts.Validation)

	override := rpc.NewStateOverride()
	opts = NewSimulateOptions().WithBlock("0x10").WithStateOverride(override).WithTraceTransfers().WithValidation()
	require.Equal(t, "0x10", opts.Block)
	require.Same(t, override, opts.StateOverride)
	require.True(t, opts.TraceTransfers)
	require.True(t, opts.Validation)
}

func TestNewSimulateCall(t *testing.T) {
	from, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)
	to, err := types.NewAddressFromHex("0x2222222222222222222222222222222222222222")
	require.NoError(t, err)
	value := big.NewInt(7)
	data := []byte{0x01, 0x02}

	call := NewSimulateCall(from, to, value, data)
	require.Same(t, from, call.From)
	require.Same(t, to, call.To)
	require.Same(t, value, call.Value)
	require.Equal(t, data, call.Data)
}

func TestSimulateCallValidate(t *testing.T) {
	from, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)
	to, err := types.NewAddressFromHex("0x2222222222222222222222222222222222222222")
	require.NoError(t, err)

	cases := []struct {
		name    string
		call    *SimulateCall
		wantErr string
	}{
		{"nil", nil, "invalid request: call is nil"},
		{"nil from", &SimulateCall{To: to}, "invalid request: from is nil"},
		{"negative value", &SimulateCall{From: from, To: to, Value: big.NewInt(-1)}, "invalid request: value is negative"},
		{"value over uint256", &SimulateCall{From: from, To: to, Value: new(big.Int).Lsh(big.NewInt(1), 256)}, "invalid request: value exceeds uint256"},
		{"create", &SimulateCall{From: from, Data: []byte{0x60, 0x00}}, ""},
		{"nil value", &SimulateCall{From: from, To: to}, ""},
		{"max uint256", &SimulateCall{From: from, To: to, Value: new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call.Validate()
			if tc.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.EqualError(t, err, tc.wantErr)
			require.ErrorIs(t, err, ErrInvalidRequest)
		})
	}
}

func TestSimulateSendsPayloadAndBlock(t *testing.T) {
	from, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)
	to, err := types.NewAddressFromHex("0x2222222222222222222222222222222222222222")
	require.NoError(t, err)
	var slot, value [32]byte
	slot[31] = 0x01
	value[31] = 0x2a
	override := rpc.NewStateOverride().WithBalance(from, big.NewInt(9)).WithStateDiff(to, slot, value)
	wantOverride, err := json.Marshal(override)
	require.NoError(t, err)

	cases := []struct {
		name        string
		opts        *SimulateOptions
		wantPayload string
		wantBlock   string
	}{
		{
			"nil options",
			nil,
			`{"blockStateCalls":[{"calls":[
				{"from":"0x1111111111111111111111111111111111111111","to":"0x2222222222222222222222222222222222222222","data":"0xa9059cbb"},
				{"from":"0x1111111111111111111111111111111111111111","value":"0xde0b6b3a7640000","data":"0x6000"}
			]}],"traceTransfers":false,"validation":false}`,
			rpc.BlockTagLatest,
		},
		{
			"all options",
			NewSimulateOptions().WithBlock("0x10").WithStateOverride(override).WithTraceTransfers().WithValidation(),
			`{"blockStateCalls":[{"stateOverrides":` + string(wantOverride) + `,"calls":[
				{"from":"0x1111111111111111111111111111111111111111","to":"0x2222222222222222222222222222222222222222","data":"0xa9059cbb"},
				{"from":"0x1111111111111111111111111111111111111111","value":"0xde0b6b3a7640000","data":"0x6000"}
			]}],"traceTransfers":true,"validation":true}`,
			"0x10",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var method string
			var params []json.RawMessage
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req struct {
					ID     string            `json:"id"`
					Method string            `json:"method"`
					Params []json.RawMessage `json:"params"`
				}
				require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
				method = req.Method
				params = req.Params
				_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"` + req.ID + `","result":[{"calls":[` +
					`{"status":"0x1","gasUsed":"0x5208","returnData":"0x","logs":[]},` +
					`{"status":"0x1","gasUsed":"0xd431","returnData":"0x","logs":[]}]}]}`))
			}))
			defer srv.Close()

			results, err := Simulate(t.Context(), rpc.NewClient(srv.URL), tc.opts, []SimulateCall{
				NewSimulateCall(from, to, nil, []byte{0xa9, 0x05, 0x9c, 0xbb}),
				NewSimulateCall(from, nil, big.NewInt(1_000_000_000_000_000_000), []byte{0x60, 0x00}),
			})
			require.NoError(t, err)
			require.Len(t, results, 2)
			require.Equal(t, uint64(21000), results[0].GasUsed)
			require.Equal(t, uint64(54321), results[1].GasUsed)

			require.Equal(t, rpc.MethodETHSimulateV1, method)
			require.Len(t, params, 2)
			require.JSONEq(t, tc.wantPayload, string(params[0]))
			var block string
			require.NoError(t, json.Unmarshal(params[1], &block))
			require.Equal(t, tc.wantBlock, block)
		})
	}
}

func TestSimulateDecodesResults(t *testing.T) {
	from, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)
	to, err := types.NewAddressFromHex("0x2222222222222222222222222222222222222222")
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID string `json:"id"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"` + req.ID + `","result":[{"number":"0x11","calls":[` +
			`{"status":"0x1","gasUsed":"0xb4a8","returnData":"0x0000000000000000000000000000000000000000000000000000000000000001","logs":[` +
			`{"address":"0x2222222222222222222222222222222222222222","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x0000000000000000000000001111111111111111111111111111111111111111","0x0000000000000000000000003333333333333333333333333333333333333333"],` +
			`"data":"0x000000000000000000000000000000000000000000000000000000000000002a","blockNumber":"0x11","transactionHash":"0x6de9ecde85473347b57a96c8eb71f9dce601416b0d062ac367404fa3a8bd2f15","logIndex":"0x0","removed":false}]},` +
			`{"status":"0x0","gasUsed":"0x5f6a","returnData":"0xfb8f41b2","logs":[],"error":{"code":3,"message":"execution reverted","data":"0xfb8f41b2"}}` +
			`]}]}`))
	}))
	defer srv.Close()

	results, err := Simulate(t.Context(), rpc.NewClient(srv.URL), nil, []SimulateCall{
		NewSimulateCall(from, to, nil, []byte{0x23, 0xb8, 0x72, 0xdd}),
		NewSimulateCall(from, to, nil, []byte{0x23, 0xb8, 0x72, 0xdd}),
	})
	require.NoError(t, err)
	require.Len(t, results, 2)

	require.True(t, results[0].Success)
	require.Equal(t, uint64(0xb4a8), results[0].GasUsed)
	require.Equal(t, append(make([]byte, 31), 0x01), results[0].ReturnData)
	require.NoError(t, results[0].Error)
	require.Len(t, results[0].Logs, 1)
	log := results[0].Logs[0]
	require.Equal(t, to.String(), log.Address.String())
	require.Len(t, log.Topics, 3)
	require.Equal(t, "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", log.Topics[0].String())
	require.Equal(t, append(make([]byte, 31), 0x2a), log.Data)
	require.Equal(t, uint64(0x11), log.BlockNumber)
	require.Equal(t, uint64(0), log.LogIndex)

	require.False(t, results[1].Success)
	require.Equal(t, uint64(0x5f6a), results[1].GasUsed)
	require.Equal(t, []byte{0xfb, 0x8f, 0x41, 0xb2}, results[1].ReturnData)
	require.Empty(t, results[1].Logs)
	callErr, ok := errors.AsType[*rpc.ResponseError](results[1].Error)
	require.True(t, ok)
	require.Equal(t, 3, callErr.Code)
	require.Equal(t, "execution reverted", callErr.Message)
	require.JSONEq(t, `"0xfb8f41b2"`, string(callErr.Data))
	require.EqualError(t, results[1].Error, `rpc error 3: execution reverted: "0xfb8f41b2"`)
}

func TestSimulateRejectsInvalidInputWithoutRPC(t *testing.T) {
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

	cases := []struct {
		name     string
		opts     *SimulateOptions
		calls    []SimulateCall
		sentinel error
		wantErr  string
	}{
		{"nil calls", nil, nil, ErrInvalidRequest, "bind: simulate: invalid request: calls is empty"},
		{"empty calls", nil, []SimulateCall{}, ErrInvalidRequest, "bind: simulate: invalid request: calls is empty"},
		{
			"nil from in second call",
			nil,
			[]SimulateCall{NewSimulateCall(from, to, nil, nil), NewSimulateCall(nil, to, nil, nil)},
			ErrInvalidRequest,
			"bind: simulate: calls[1]: invalid request: from is nil",
		},
		{
			"negative value",
			nil,
			[]SimulateCall{NewSimulateCall(from, to, big.NewInt(-1), nil)},
			ErrInvalidRequest,
			"bind: simulate: calls[0]: invalid request: value is negative",
		},
		{
			"value over uint256",
			nil,
			[]SimulateCall{NewSimulateCall(from, to, new(big.Int).Lsh(big.NewInt(1), 256), nil)},
			ErrInvalidRequest,
			"bind: simulate: calls[0]: invalid request: value exceeds uint256",
		},
		{
			"invalid override",
			NewSimulateOptions().WithStateOverride(rpc.NewStateOverride().WithState(to, slot, value).WithStateDiff(to, slot, value)),
			[]SimulateCall{NewSimulateCall(from, to, nil, nil)},
			ErrInvalidOptions,
			"state and stateDiff are both set",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			results, err := Simulate(t.Context(), cli, tc.opts, tc.calls)
			require.Nil(t, results)
			require.ErrorIs(t, err, tc.sentinel)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
	require.Zero(t, hits.Load())
}

func TestSimulateReturnsResponseErrorUnwrapped(t *testing.T) {
	from, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID string `json:"id"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"` + req.ID + `","error":{"code":-32601,"message":"the method eth_simulateV1 does not exist/is not available"}}`))
	}))
	defer srv.Close()

	results, err := Simulate(t.Context(), rpc.NewClient(srv.URL), nil, []SimulateCall{NewSimulateCall(from, nil, nil, nil)})
	require.Nil(t, results)
	rpcErr, ok := errors.AsType[*rpc.ResponseError](err)
	require.True(t, ok, "error must be the raw *rpc.ResponseError, got %T", err)
	require.Equal(t, -32601, rpcErr.Code)
	require.NotContains(t, err.Error(), "bind: simulate")
}

func TestSimulateRejectsMalformedResults(t *testing.T) {
	from, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)

	cases := []struct {
		name    string
		result  string
		wantErr string
	}{
		{"no blocks", `[]`, "expected 1 block, got 0"},
		{"two blocks", `[{"calls":[]},{"calls":[]}]`, "expected 1 block, got 2"},
		{"no calls array", `[{"number":"0x1"}]`, "block has no calls array"},
		{"call count mismatch", `[{"calls":[]}]`, "expected 1 call results, got 0"},
		{"call not object", `[{"calls":["x"]}]`, "calls[0]: unexpected type string"},
		{"missing status", `[{"calls":[{"gasUsed":"0x1","returnData":"0x"}]}]`, "calls[0]: invalid status"},
		{"unknown status", `[{"calls":[{"status":"0x2","gasUsed":"0x1","returnData":"0x"}]}]`, "calls[0]: invalid status 0x2"},
		{"missing gasUsed", `[{"calls":[{"status":"0x1","returnData":"0x"}]}]`, "calls[0]: missing gasUsed"},
		{"bad gasUsed", `[{"calls":[{"status":"0x1","gasUsed":"0xzz","returnData":"0x"}]}]`, "calls[0]: gasUsed:"},
		{"missing returnData", `[{"calls":[{"status":"0x1","gasUsed":"0x1"}]}]`, "calls[0]: missing returnData"},
		{"bad returnData", `[{"calls":[{"status":"0x1","gasUsed":"0x1","returnData":"0xzz"}]}]`, "calls[0]: returnData:"},
		{"log not object", `[{"calls":[{"status":"0x1","gasUsed":"0x1","returnData":"0x","logs":[1]}]}]`, "calls[0]: logs[0]: unexpected type float64"},
		{"log missing address", `[{"calls":[{"status":"0x1","gasUsed":"0x1","returnData":"0x","logs":[{"topics":[],"data":"0x"}]}]}]`, "calls[0]: logs[0]: missing address"},
		{"error without code", `[{"calls":[{"status":"0x0","gasUsed":"0x1","returnData":"0x","error":{"message":"execution reverted"}}]}]`, "calls[0]: error: missing code"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req struct {
					ID string `json:"id"`
				}
				require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
				_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"` + req.ID + `","result":` + tc.result + `}`))
			}))
			defer srv.Close()

			results, err := Simulate(t.Context(), rpc.NewClient(srv.URL), nil, []SimulateCall{NewSimulateCall(from, nil, nil, nil)})
			require.Nil(t, results)
			require.ErrorContains(t, err, "bind: simulate: ")
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}
