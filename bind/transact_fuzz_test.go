package bind

import (
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// go test -run '^$' -fuzz '^FuzzTransactRoundTrip$' -fuzztime=10s ./bind
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
