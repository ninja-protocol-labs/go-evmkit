package bind

import (
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
	"github.com/stretchr/testify/require"
)

func TestTransactOptionsFrom(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)

	var nilOpts *TransactOptions
	from, err := nilOpts.From()
	require.ErrorIs(t, err, ErrInvalidOptions)
	require.ErrorContains(t, err, "options is nil")
	require.Nil(t, from)

	from, err = (&TransactOptions{}).From()
	require.ErrorIs(t, err, ErrInvalidOptions)
	require.ErrorContains(t, err, "signer is nil")
	require.Nil(t, from)

	from, err = (&TransactOptions{Signer: key}).From()
	require.NoError(t, err)
	require.True(t, from.Equal(core.PubkeyToAddress(key.PublicKey())))
}

func TestTransactRejectsIncompleteOptions(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	to, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)

	cases := []struct {
		name string
		opts *TransactOptions
		want string
	}{
		{"nil options", nil, "options is nil"},
		{"nil signer", &TransactOptions{Builder: tx.NewDynamicFeeBuilder()}, "signer is nil"},
		{"nil builder", &TransactOptions{Signer: key}, "builder is nil"},
		{"nil signer and builder", &TransactOptions{}, "signer is nil"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hits := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits++
			}))
			defer srv.Close()

			signed, hash, err := Transact(t.Context(), rpc.NewClient(srv.URL), tc.opts, to, []byte{0x01})
			require.ErrorIs(t, err, ErrInvalidOptions)
			require.ErrorContains(t, err, tc.want)
			require.Nil(t, signed)
			require.Nil(t, hash)
			require.Zero(t, hits)
		})
	}
}

func TestTransactDynamicFee(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	to, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)
	calldata := []byte{0xa9, 0x05, 0x9c, 0xbb, 0x00, 0x01}

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(42)
	builder.Limit(100_000)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	var methods []string
	var raw []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     string   `json:"id"`
			Method string   `json:"method"`
			Params []string `json:"params"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		methods = append(methods, req.Method)
		require.Len(t, req.Params, 1)
		decoded, err := encoding.Hex.Decode(req.Params[0])
		require.NoError(t, err)
		raw = decoded
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"` + req.ID + `","result":"` + core.Keccak256(decoded).String() + `"}`))
	}))
	defer srv.Close()

	signed, hash, err := Transact(t.Context(), rpc.NewClient(srv.URL), &TransactOptions{Signer: key, Builder: builder, Value: big.NewInt(12345)}, to, calldata)
	require.NoError(t, err)
	require.NotNil(t, signed)
	require.Equal(t, []string{rpc.MethodETHSendRawTransaction}, methods)

	require.True(t, hash.Equal(core.Keccak256(raw)))
	signedHash, err := signed.Hash()
	require.NoError(t, err)
	require.True(t, hash.Equal(signedHash))

	decoded, err := core.DecodeTransaction(raw)
	require.NoError(t, err)
	dyn, ok := decoded.(*core.DynamicFeeTx)
	require.True(t, ok, "got %T", decoded)
	require.Zero(t, dyn.ChainID.Cmp(big.NewInt(1)))
	require.Equal(t, uint64(42), dyn.Nonce)
	require.Equal(t, uint64(100_000), dyn.GasLimit)
	require.Zero(t, dyn.GasFeeCap.Cmp(big.NewInt(30_000_000_000)))
	require.Zero(t, dyn.GasTipCap.Cmp(big.NewInt(1_000_000_000)))
	require.True(t, dyn.To.Equal(to))
	require.Zero(t, dyn.Value.Cmp(big.NewInt(12345)))
	require.Equal(t, calldata, dyn.Data)

	sender, err := decoded.Sender()
	require.NoError(t, err)
	require.True(t, sender.Equal(core.PubkeyToAddress(key.PublicKey())))
	require.Equal(t, uint64(42), signed.AddressNonce())
}

func TestTransactLegacy(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	to, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)
	calldata := []byte{0x09, 0x5e, 0xa7, 0xb3}

	builder := tx.NewLegacyBuilder()
	builder.ChainID(big.NewInt(56))
	builder.Nonce(7)
	builder.Limit(60_000)
	builder.Price(big.NewInt(3_000_000_000), nil)

	var raw []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     string   `json:"id"`
			Params []string `json:"params"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		decoded, err := encoding.Hex.Decode(req.Params[0])
		require.NoError(t, err)
		raw = decoded
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"` + req.ID + `","result":"` + core.Keccak256(decoded).String() + `"}`))
	}))
	defer srv.Close()

	signed, hash, err := Transact(t.Context(), rpc.NewClient(srv.URL), &TransactOptions{Signer: key, Builder: builder}, to, calldata)
	require.NoError(t, err)
	require.True(t, hash.Equal(core.Keccak256(raw)))
	require.Equal(t, uint64(7), signed.AddressNonce())

	decoded, err := core.DecodeTransaction(raw)
	require.NoError(t, err)
	legacy, ok := decoded.(*core.LegacyTx)
	require.True(t, ok, "got %T", decoded)
	require.Zero(t, legacy.ChainID.Cmp(big.NewInt(56)))
	require.Equal(t, uint64(7), legacy.Nonce)
	require.Equal(t, uint64(60_000), legacy.GasLimit)
	require.Zero(t, legacy.GasPrice.Cmp(big.NewInt(3_000_000_000)))
	require.True(t, legacy.To.Equal(to))
	require.Zero(t, legacy.Value.Sign())
	require.Equal(t, calldata, legacy.Data)

	sender, err := decoded.Sender()
	require.NoError(t, err)
	require.True(t, sender.Equal(core.PubkeyToAddress(key.PublicKey())))
}

func TestTransactContractCreation(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	initCode := []byte{0x60, 0x80, 0x60, 0x40, 0x52, 0x34, 0x80, 0x15}

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(3)
	builder.Limit(500_000)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	var raw []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     string   `json:"id"`
			Params []string `json:"params"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		decoded, err := encoding.Hex.Decode(req.Params[0])
		require.NoError(t, err)
		raw = decoded
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"` + req.ID + `","result":"` + core.Keccak256(decoded).String() + `"}`))
	}))
	defer srv.Close()

	signed, hash, err := Transact(t.Context(), rpc.NewClient(srv.URL), &TransactOptions{Signer: key, Builder: builder}, nil, initCode)
	require.NoError(t, err)
	require.True(t, hash.Equal(core.Keccak256(raw)))

	decoded, err := core.DecodeTransaction(raw)
	require.NoError(t, err)
	dyn, ok := decoded.(*core.DynamicFeeTx)
	require.True(t, ok, "got %T", decoded)
	require.Nil(t, dyn.To)
	require.Zero(t, dyn.Value.Sign())
	require.Equal(t, initCode, dyn.Data)

	from := core.PubkeyToAddress(key.PublicKey())
	want, err := core.CreateAddress(from, 3)
	require.NoError(t, err)
	got, err := core.CreateAddress(from, signed.AddressNonce())
	require.NoError(t, err)
	require.True(t, want.Equal(got))
}

func TestTransactBuilderReusedAcrossCalls(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	to, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Limit(21_000)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	nonce := uint64(10)
	var seen []uint64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := json.RawMessage{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		if len(body) > 0 && body[0] == '[' {
			var reqs []struct {
				ID     string `json:"id"`
				Method string `json:"method"`
			}
			require.NoError(t, json.Unmarshal(body, &reqs))
			require.Len(t, reqs, 1)
			require.Equal(t, rpc.MethodETHGetTransactionCount, reqs[0].Method)
			_, _ = w.Write([]byte(`[{"jsonrpc":"2.0","id":"` + reqs[0].ID + `","result":"0x` + big.NewInt(int64(nonce)).Text(16) + `"}]`))
			return
		}
		var req struct {
			ID     string   `json:"id"`
			Params []string `json:"params"`
		}
		require.NoError(t, json.Unmarshal(body, &req))
		raw, err := encoding.Hex.Decode(req.Params[0])
		require.NoError(t, err)
		decoded, err := core.DecodeTransaction(raw)
		require.NoError(t, err)
		seen = append(seen, decoded.AddressNonce())
		nonce++
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"` + req.ID + `","result":"` + core.Keccak256(raw).String() + `"}`))
	}))
	defer srv.Close()

	opts := &TransactOptions{Signer: key, Builder: builder}
	for range 3 {
		_, _, err := Transact(t.Context(), rpc.NewClient(srv.URL), opts, to, nil)
		require.NoError(t, err)
	}
	require.Equal(t, []uint64{10, 11, 12}, seen)
}

func TestTransactPackError(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	to, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)

	sends := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := json.RawMessage{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		if len(body) > 0 && body[0] != '[' {
			sends++
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer srv.Close()

	signed, hash, err := Transact(t.Context(), rpc.NewClient(srv.URL), &TransactOptions{Signer: key, Builder: tx.NewDynamicFeeBuilder()}, to, []byte{0x01})
	require.Error(t, err)
	require.ErrorContains(t, err, "bind: transact:")
	require.ErrorContains(t, err, "http status 400")
	require.Nil(t, signed)
	require.Nil(t, hash)
	require.Zero(t, sends)
}

func TestTransactBroadcastErrorReturnsSignedTx(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	to, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(9)
	builder.Limit(21_000)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	var raw []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     string   `json:"id"`
			Params []string `json:"params"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		decoded, err := encoding.Hex.Decode(req.Params[0])
		require.NoError(t, err)
		raw = decoded
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"` + req.ID + `","error":{"code":-32000,"message":"nonce too low"}}`))
	}))
	defer srv.Close()

	signed, hash, err := Transact(t.Context(), rpc.NewClient(srv.URL), &TransactOptions{Signer: key, Builder: builder}, to, nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "bind: transact:")
	require.Nil(t, hash)
	require.NotNil(t, signed)
	require.Equal(t, uint64(9), signed.AddressNonce())

	var rpcErr *rpc.ResponseError
	require.True(t, errors.As(err, &rpcErr))
	require.Equal(t, -32000, rpcErr.Code)
	require.Equal(t, "nonce too low", rpcErr.Message)

	signedHash, err := signed.Hash()
	require.NoError(t, err)
	require.True(t, signedHash.Equal(core.Keccak256(raw)))
}

func TestNewTransactOptions(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	builder := tx.NewDynamicFeeBuilder()

	opts := NewTransactOptions(key, builder)
	require.Equal(t, core.Signer(key), opts.Signer)
	require.Equal(t, tx.Builder(builder), opts.Builder)
	require.Nil(t, opts.Value)

	same := opts.WithValue(big.NewInt(42))
	require.Same(t, opts, same)
	require.Zero(t, opts.Value.Cmp(big.NewInt(42)))

	from, err := opts.From()
	require.NoError(t, err)
	require.True(t, from.Equal(core.PubkeyToAddress(key.PublicKey())))

	signed, hash, err := Transact(t.Context(), rpc.NewClient("http://127.0.0.1:0"), NewTransactOptions(nil, builder), nil, nil)
	require.ErrorIs(t, err, ErrInvalidOptions)
	require.ErrorContains(t, err, "signer is nil")
	require.Nil(t, signed)
	require.Nil(t, hash)

	signed, hash, err = Transact(t.Context(), rpc.NewClient("http://127.0.0.1:0"), NewTransactOptions(key, nil), nil, nil)
	require.ErrorIs(t, err, ErrInvalidOptions)
	require.ErrorContains(t, err, "builder is nil")
	require.Nil(t, signed)
	require.Nil(t, hash)
}
