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
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

func TestNewCallOptions(t *testing.T) {
	from, err := types.NewAddressFromHex("0x2222222222222222222222222222222222222222")
	require.NoError(t, err)

	empty := NewCallOptions()
	require.Equal(t, &CallOptions{}, empty)

	opts := NewCallOptions()
	same := opts.WithFrom(from).WithBlock("0x1b4").WithValue(big.NewInt(7))
	require.Same(t, opts, same)
	require.True(t, opts.From.Equal(from))
	require.Equal(t, "0x1b4", opts.Block)
	require.Zero(t, opts.Value.Cmp(big.NewInt(7)))

	opts.WithFrom(nil).WithBlock("").WithValue(nil)
	require.Equal(t, &CallOptions{}, opts)
}

func TestCallNilOptions(t *testing.T) {
	to, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)
	calldata := []byte{0x70, 0xa0, 0x82, 0x31, 0x01, 0x02}

	var method string
	var callParams map[string]string
	var block string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     string            `json:"id"`
			Method string            `json:"method"`
			Params []json.RawMessage `json:"params"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		method = req.Method
		require.Len(t, req.Params, 2)
		require.NoError(t, json.Unmarshal(req.Params[0], &callParams))
		require.NoError(t, json.Unmarshal(req.Params[1], &block))
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"` + req.ID + `","result":"0xdeadbeef"}`))
	}))
	defer srv.Close()

	out, err := Call(t.Context(), rpc.NewClient(srv.URL), nil, to, calldata)
	require.NoError(t, err)
	require.Equal(t, []byte{0xde, 0xad, 0xbe, 0xef}, out)

	require.Equal(t, rpc.MethodETHCall, method)
	require.Equal(t, rpc.BlockTagLatest, block)
	require.Equal(t, map[string]string{
		"to":   to.String(),
		"data": encoding.Hex.EncodePrefixed(calldata),
	}, callParams)
}

func TestCallWithOptions(t *testing.T) {
	to, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)
	from, err := types.NewAddressFromHex("0x2222222222222222222222222222222222222222")
	require.NoError(t, err)
	calldata := []byte{0xa9, 0x05, 0x9c, 0xbb}

	cases := []struct {
		name      string
		block     string
		value     *big.Int
		wantBlock string
		wantValue string
	}{
		{"latest explicit", rpc.BlockTagLatest, big.NewInt(1), rpc.BlockTagLatest, "0x1"},
		{"pending", rpc.BlockTagPending, big.NewInt(1_000_000_000_000_000_000), rpc.BlockTagPending, "0xde0b6b3a7640000"},
		{"hex number", "0x1b4", big.NewInt(255), "0x1b4", "0xff"},
		{"zero value", "", big.NewInt(0), rpc.BlockTagLatest, "0x0"},
		{"max uint256", rpc.BlockTagSafe, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)), rpc.BlockTagSafe, "0x" + "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var callParams map[string]string
			var block string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req struct {
					ID     string            `json:"id"`
					Params []json.RawMessage `json:"params"`
				}
				require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
				require.NoError(t, json.Unmarshal(req.Params[0], &callParams))
				require.NoError(t, json.Unmarshal(req.Params[1], &block))
				_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"` + req.ID + `","result":"0x"}`))
			}))
			defer srv.Close()

			out, err := Call(t.Context(), rpc.NewClient(srv.URL), &CallOptions{From: from, Block: tc.block, Value: tc.value}, to, calldata)
			require.NoError(t, err)
			require.Empty(t, out)

			require.Equal(t, tc.wantBlock, block)
			require.Equal(t, map[string]string{
				"to":    to.String(),
				"from":  from.String(),
				"value": tc.wantValue,
				"data":  encoding.Hex.EncodePrefixed(calldata),
			}, callParams)
		})
	}
}

func TestCallNilTo(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
	}))
	defer srv.Close()

	out, err := Call(t.Context(), rpc.NewClient(srv.URL), &CallOptions{}, nil, []byte{0x01})
	require.ErrorIs(t, err, ErrInvalidRequest)
	require.ErrorContains(t, err, "to is nil")
	require.Nil(t, out)
	require.Zero(t, hits)
}

func TestCallReturnsResponseErrorUnwrapped(t *testing.T) {
	to, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID string `json:"id"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"` + req.ID + `","error":{"code":3,"message":"execution reverted","data":"0xe450d38c0000000000000000000000002222222222222222222222222222222222222222"}}`))
	}))
	defer srv.Close()

	out, err := Call(t.Context(), rpc.NewClient(srv.URL), nil, to, []byte{0x01})
	require.Nil(t, out)

	var rpcErr *rpc.ResponseError
	ok := errors.As(err, &rpcErr)
	require.True(t, ok, "error must be the raw *rpc.ResponseError, got %T", err)
	require.Equal(t, 3, rpcErr.Code)
	require.Equal(t, "execution reverted", rpcErr.Message)
	require.JSONEq(t, `"0xe450d38c0000000000000000000000002222222222222222222222222222222222222222"`, string(rpcErr.Data))
}

func TestCallInvalidResultHex(t *testing.T) {
	to, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)

	for _, result := range []string{`"0xzz"`, `"0x123"`, `"not hex"`} {
		t.Run(result, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req struct {
					ID string `json:"id"`
				}
				require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
				_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"` + req.ID + `","result":` + result + `}`))
			}))
			defer srv.Close()

			out, err := Call(t.Context(), rpc.NewClient(srv.URL), nil, to, []byte{0x01})
			require.Error(t, err)
			require.Nil(t, out)
		})
	}
}

func TestCallOptionsWithStateOverrideSetsField(t *testing.T) {
	override := rpc.NewStateOverride()
	opts := NewCallOptions().WithStateOverride(override)
	require.Same(t, override, opts.StateOverride)
	require.Nil(t, opts.WithStateOverride(nil).StateOverride)
}

func TestCallSendsStateOverrideAsThirdParam(t *testing.T) {
	to, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)
	from, err := types.NewAddressFromHex("0x2222222222222222222222222222222222222222")
	require.NoError(t, err)
	var slot, value [32]byte
	slot[31] = 0x04
	value[31] = 0x2a
	override := rpc.NewStateOverride().
		WithCode(to, []byte{0x60, 0x00}).
		WithStateDiff(to, slot, value).
		WithBalance(from, big.NewInt(7))

	var gotParams []json.RawMessage
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     string            `json:"id"`
			Params []json.RawMessage `json:"params"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		gotParams = req.Params
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"` + req.ID + `","result":"0x2a"}`))
	}))
	defer srv.Close()
	cli := rpc.NewClient(srv.URL)

	cases := []struct {
		opts      *CallOptions
		wantBlock string
		wantThird bool
	}{
		{nil, rpc.BlockTagLatest, false},
		{NewCallOptions(), rpc.BlockTagLatest, false},
		{NewCallOptions().WithStateOverride(override), rpc.BlockTagLatest, true},
		{NewCallOptions().WithFrom(from).WithBlock("0x10").WithStateOverride(override), "0x10", true},
		{NewCallOptions().WithStateOverride(rpc.NewStateOverride()), rpc.BlockTagLatest, true},
	}
	for i, c := range cases {
		out, err := Call(t.Context(), cli, c.opts, to, []byte{0x01})
		require.NoError(t, err, i)
		require.Equal(t, []byte{0x2a}, out, i)

		var block string
		require.NoError(t, json.Unmarshal(gotParams[1], &block), i)
		require.Equal(t, c.wantBlock, block, i)
		if !c.wantThird {
			require.Len(t, gotParams, 2, i)
			continue
		}
		require.Len(t, gotParams, 3, i)
		want, err := json.Marshal(c.opts.StateOverride)
		require.NoError(t, err, i)
		require.JSONEq(t, string(want), string(gotParams[2]), i)
	}
}

func TestCallRejectsInvalidStateOverrideWithoutRPC(t *testing.T) {
	to, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)
	var slot, value [32]byte

	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
	}))
	defer srv.Close()
	cli := rpc.NewClient(srv.URL)

	cases := []struct {
		override *rpc.StateOverride
		wantErr  string
	}{
		{rpc.NewStateOverride().WithBalance(nil, big.NewInt(1)), "account is nil"},
		{rpc.NewStateOverride().WithBalance(to, big.NewInt(-1)), "balance is negative"},
		{rpc.NewStateOverride().WithBalance(to, new(big.Int).Lsh(big.NewInt(1), 256)), "balance exceeds uint256"},
		{
			rpc.NewStateOverride().WithState(to, slot, value).WithStateDiff(to, slot, value),
			"state and stateDiff are both set",
		},
	}
	for i, c := range cases {
		out, err := Call(t.Context(), cli, NewCallOptions().WithStateOverride(c.override), to, []byte{0x01})
		require.Nil(t, out, i)
		require.ErrorIs(t, err, ErrInvalidOptions, i)
		require.ErrorIs(t, err, rpc.ErrInvalidStateOverride, i)
		require.ErrorContains(t, err, "bind: call:", i)
		require.ErrorContains(t, err, c.wantErr, i)
	}
	require.Zero(t, hits.Load())
}
