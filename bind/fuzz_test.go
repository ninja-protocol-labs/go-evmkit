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
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

func FuzzCallParams(f *testing.F) {
	f.Add([]byte{}, []byte{}, false, false, "", []byte{})
	f.Add(make([]byte, 20), []byte{0x00}, true, true, "latest", []byte{0x70, 0xa0, 0x82, 0x31})
	f.Add([]byte{0xff, 0xee, 0xdd}, []byte{0x01, 0x00}, true, false, "0x1b4", make([]byte, 68))
	f.Add(make([]byte, 40), make([]byte, 33), false, true, "pending", []byte{0x00})

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

func FuzzTransactRoundTrip(f *testing.F) {
	f.Add([]byte{}, false, []byte{}, []byte{}, uint64(0), uint64(21_000), true)
	f.Add(make([]byte, 20), true, []byte{0x01}, []byte{0xa9, 0x05, 0x9c, 0xbb}, uint64(42), uint64(100_000), true)
	f.Add([]byte{0x11, 0x22}, false, make([]byte, 32), make([]byte, 200), uint64(1<<63), uint64(30_000_000), false)
	f.Add(make([]byte, 20), true, []byte{0xff, 0xff}, []byte{0x60, 0x80, 0x60, 0x40}, ^uint64(0)-1, uint64(1), false)

	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(f, err)
	sender := core.PubkeyToAddress(key.PublicKey())

	var raw []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     string   `json:"id"`
			Method string   `json:"method"`
			Params []string `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Method != rpc.MethodETHSendRawTransaction || len(req.Params) != 1 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		decoded, err := encoding.Hex.Decode(req.Params[0])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		raw = decoded
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"` + req.ID + `","result":"` + core.Keccak256(decoded).String() + `"}`))
	}))
	f.Cleanup(srv.Close)
	client := rpc.NewClient(srv.URL)

	f.Fuzz(func(t *testing.T, toBytes []byte, create bool, valueBytes, calldata []byte, nonce, gasLimit uint64, dynamic bool) {
		if len(valueBytes) > 32 {
			valueBytes = valueBytes[:32]
		}
		value := new(big.Int).SetBytes(valueBytes)
		var to *types.Address
		if !create {
			to = types.NewAddressFromBytes(toBytes)
		}

		var builder tx.Builder
		if dynamic {
			builder = tx.NewDynamicFeeBuilder()
		} else {
			builder = tx.NewLegacyBuilder()
		}
		builder.ChainID(big.NewInt(1))
		builder.Nonce(nonce)
		builder.Limit(gasLimit)
		builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

		raw = nil
		signed, hash, err := Transact(t.Context(), client, &TransactOptions{Signer: key, Builder: builder, Value: value}, to, calldata)
		require.NoError(t, err)
		require.True(t, hash.Equal(core.Keccak256(raw)))
		require.Equal(t, nonce, signed.AddressNonce())
		require.Equal(t, gasLimit, signed.Gas())

		decoded, err := core.DecodeTransaction(raw)
		require.NoError(t, err)
		gotSender, err := decoded.Sender()
		require.NoError(t, err)
		require.True(t, gotSender.Equal(sender))

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

		if create {
			require.Nil(t, gotTo)
		} else {
			require.NotNil(t, gotTo)
			require.True(t, gotTo.Equal(to))
		}
		require.Zero(t, value.Cmp(gotValue))
		require.Equal(t, len(calldata), len(gotData))
		if len(calldata) > 0 {
			require.Equal(t, calldata, gotData)
		}
	})
}
