package bind

import (
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

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
