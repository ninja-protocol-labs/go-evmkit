package tx

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

func TestBroadcastConfirmsByHashAfterSendError(t *testing.T) {
	raw := []byte{0x02, 0xf8, 0x6b, 0x01, 0x09}
	wantHash := core.Keccak256(raw)

	cases := []struct {
		name        string
		send        string
		lookup      string
		wantErr     string
		wantMethods []string
	}{
		{"sent", `"result":"` + wantHash.String() + `"`, "", "", []string{rpc.MethodETHSendRawTransaction}},
		{"send error but node knows the tx", `"error":{"code":-32000,"message":"nonce too low"}`, `"result":{"hash":"` + wantHash.String() + `","nonce":"0x11"}`, "", []string{rpc.MethodETHSendRawTransaction, rpc.MethodETHGetTransactionByHash}},
		{"send error and tx unknown", `"error":{"code":-32000,"message":"nonce too low"}`, `"result":null`, "tx: broadcast: rpc error -32000: nonce too low", []string{rpc.MethodETHSendRawTransaction, rpc.MethodETHGetTransactionByHash}},
		{"send error and lookup error", `"error":{"code":-32000,"message":"insufficient funds"}`, `"error":{"code":-32601,"message":"method not found"}`, "tx: broadcast: rpc error -32000: insufficient funds", []string{rpc.MethodETHSendRawTransaction, rpc.MethodETHGetTransactionByHash}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var methods []string
			var lookupParams []string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req struct {
					ID     string   `json:"id"`
					Method string   `json:"method"`
					Params []string `json:"params"`
				}
				require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
				methods = append(methods, req.Method)
				body := tc.send
				if req.Method == rpc.MethodETHGetTransactionByHash {
					lookupParams = req.Params
					body = tc.lookup
				}
				_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"` + req.ID + `",` + body + `}`))
			}))
			defer srv.Close()

			hash, err := Broadcast(t.Context(), rpc.NewClient(srv.URL), raw)
			require.Equal(t, tc.wantMethods, methods)
			if len(tc.wantMethods) == 2 {
				require.Equal(t, []string{wantHash.String()}, lookupParams)
			}
			if tc.wantErr == "" {
				require.NoError(t, err)
				require.True(t, hash.Equal(wantHash))
				return
			}
			require.Nil(t, hash)
			require.EqualError(t, err, tc.wantErr)
			rpcErr, ok := errors.AsType[*rpc.ResponseError](err)
			require.True(t, ok)
			require.Equal(t, -32000, rpcErr.Code)
		})
	}
}
