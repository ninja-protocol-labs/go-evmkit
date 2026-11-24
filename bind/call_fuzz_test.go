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

// go test -run '^$' -fuzz '^FuzzCallParams$' -fuzztime=10s ./bind
func FuzzCallParams(f *testing.F) {
	f.Add([]byte{}, []byte{}, false, false, "", []byte{})
	f.Add(make([]byte, 20), []byte{0x00}, true, true, rpc.BlockTagLatest, []byte{0x70, 0xa0, 0x82, 0x31})
	f.Add([]byte{0xff, 0xee, 0xdd}, []byte{0x01, 0x00}, true, false, "0x1b4", make([]byte, 68))
	f.Add(make([]byte, 40), make([]byte, 33), false, true, rpc.BlockTagPending, []byte{0x00})

	var callParams map[string]string
	var gotBlock string
	var reply []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     string            `json:"id"`
			Params []json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Params) != 2 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		callParams = nil
		_ = json.Unmarshal(req.Params[0], &callParams)
		_ = json.Unmarshal(req.Params[1], &gotBlock)
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"` + req.ID + `","result":"` + encoding.Hex.EncodePrefixed(reply) + `"}`))
	}))
	f.Cleanup(srv.Close)
	client := rpc.NewClient(srv.URL)

	f.Fuzz(func(t *testing.T, toBytes, valueBytes []byte, withFrom, withValue bool, block string, calldata []byte) {
		if !utf8.ValidString(block) {
			t.Skip()
		}
		if len(valueBytes) > 32 {
			valueBytes = valueBytes[:32]
		}
		to := types.NewAddressFromBytes(toBytes)
		from := types.NewAddressFromBytes(core.Keccak256(toBytes).Bytes()[:20])
		value := new(big.Int).SetBytes(valueBytes)

		opts := &CallOptions{Block: block}
		if withFrom {
			opts.From = from
		}
		if withValue {
			opts.Value = value
		}

		reply = calldata
		out, err := Call(t.Context(), client, opts, to, calldata)
		require.NoError(t, err)
		require.Equal(t, len(calldata), len(out))
		if len(calldata) > 0 {
			require.Equal(t, calldata, out)
		}

		want := map[string]string{
			"to":   to.String(),
			"data": encoding.Hex.EncodePrefixed(calldata),
		}
		if withFrom {
			want["from"] = from.String()
		}
		if withValue {
			want["value"] = "0x" + value.Text(16)
		}
		require.Equal(t, want, callParams)

		if block == "" {
			require.Equal(t, rpc.BlockTagLatest, gotBlock)
		} else {
			require.Equal(t, block, gotBlock)
		}
	})
}
