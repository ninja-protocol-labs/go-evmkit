package tx

import (
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

func TestSetCodeBuilderSendsAuthorizationListToEstimateGas(t *testing.T) {
	from, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)
	to, err := types.NewAddressFromHex("0xe05fcC23807536bEe418f142D19fa0d21BB0cfF7")
	require.NoError(t, err)
	delegateA, err := types.NewAddressFromHex("0x2222222222222222222222222222222222222222")
	require.NoError(t, err)
	delegateB, err := types.NewAddressFromHex("0x3333333333333333333333333333333333333333")
	require.NoError(t, err)

	sigA, err := encoding.Hex.Decode("0xd03b6ab951642993417a43589dfa1f3bc3c68b8181841298553e8ccdcd8eaac7" + "56f6f045cd96573286df6049415d0692cd61f3e8d5a016d89373d85f436f07d0" + "01")
	require.NoError(t, err)
	signatureA, err := types.NewSignatureFromBytes(sigA)
	require.NoError(t, err)
	sigB, err := encoding.Hex.Decode("0x544f5aa497ce4094ec856b49d85f82bc6b6454519ecede7be07689a35af406d3" + "00bfc4545491b2825979b1141d0a31d62004bd72d08093e0e916fbf823b1b683" + "00")
	require.NoError(t, err)
	signatureB, err := types.NewSignatureFromBytes(sigB)
	require.NoError(t, err)

	authorizations := []core.Authorization{
		{ChainID: big.NewInt(1), Address: delegateA, Nonce: 7, Signature: signatureA},
		{ChainID: big.NewInt(0), Address: delegateB, Nonce: 0, Signature: signatureB},
	}
	for i := range authorizations {
		authority, err := authorizations[i].Authority()
		require.NoError(t, err)
		require.True(t, authority.Equal(to), "authorization %d recovers %s", i, authority)
	}

	wantList := `[
		{"chainId":"0x1","address":"0x2222222222222222222222222222222222222222","nonce":"0x7","yParity":"0x1",
		 "r":"0xd03b6ab951642993417a43589dfa1f3bc3c68b8181841298553e8ccdcd8eaac7","s":"0x56f6f045cd96573286df6049415d0692cd61f3e8d5a016d89373d85f436f07d0"},
		{"chainId":"0x0","address":"0x3333333333333333333333333333333333333333","nonce":"0x0","yParity":"0x0",
		 "r":"0x544f5aa497ce4094ec856b49d85f82bc6b6454519ecede7be07689a35af406d3","s":"0xbfc4545491b2825979b1141d0a31d62004bd72d08093e0e916fbf823b1b683"}
	]`

	var estimateParams []json.RawMessage
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqs []struct {
			ID     string            `json:"id"`
			Method string            `json:"method"`
			Params []json.RawMessage `json:"params"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&reqs))
		resps := make([]map[string]any, 0, len(reqs))
		for _, req := range reqs {
			var result any
			switch req.Method {
			case rpc.MethodETHChainID:
				result = "0x1"
			case rpc.MethodETHGetTransactionCount:
				result = "0x5"
			case rpc.MethodETHMaxPriorityFeePerGas:
				result = "0x1"
			case rpc.MethodETHFeeHistory:
				result = map[string]any{"baseFeePerGas": []any{"0x7", "0x7"}, "reward": []any{[]any{"0x1"}}}
			case rpc.MethodETHEstimateGas:
				estimateParams = req.Params
				result = "0x186a0"
			}
			resps = append(resps, map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
		}
		require.NoError(t, json.NewEncoder(w).Encode(resps))
	}))
	defer srv.Close()
	cli := rpc.NewClient(srv.URL)

	for _, c := range []struct {
		buffer  uint64
		wantGas uint64
	}{
		{0, 100_000},
		{130, 130_000},
	} {
		estimateParams = nil
		b := NewSetCodeBuilder(authorizations)
		b.LimitBuffer(c.buffer)
		packed, err := b.Build(from, to, big.NewInt(0), []byte{0x42, 0xcb, 0xb1, 0x5c}).Pack(t.Context(), cli)
		require.NoError(t, err, c.buffer)
		require.Equal(t, c.wantGas, packed.Gas(), c.buffer)

		require.Len(t, estimateParams, 2, c.buffer)
		var call map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(estimateParams[0], &call), c.buffer)
		require.JSONEq(t, wantList, string(call["authorizationList"]), c.buffer)
		require.JSONEq(t, `"0x42cbb15c"`, string(call["data"]), c.buffer)
	}
}

func TestSetCodeBuilderRejectsUnusableAuthorizationsBeforeEstimate(t *testing.T) {
	from, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)
	to, err := types.NewAddressFromHex("0x2222222222222222222222222222222222222222")
	require.NoError(t, err)
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xa11ce))
	require.NoError(t, err)
	signed := core.Authorization{ChainID: big.NewInt(1), Address: to, Nonce: 3}
	require.NoError(t, signed.Sign(key))

	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
	}))
	defer srv.Close()
	cli := rpc.NewClient(srv.URL)

	for i, c := range []struct {
		authorizations []core.Authorization
		wantIs         error
		wantContains   string
	}{
		{[]core.Authorization{{ChainID: big.NewInt(1), Address: to}}, core.ErrUnsigned, "tx: gasLimit: authorization 0:"},
		{[]core.Authorization{signed, {ChainID: big.NewInt(1), Address: to, Nonce: 4}}, core.ErrUnsigned, "tx: gasLimit: authorization 1:"},
		{[]core.Authorization{{ChainID: big.NewInt(1), Nonce: 3, Signature: signed.Signature}}, nil, "tx: gasLimit: authorization 0: address is required"},
	} {
		_, err := NewSetCodeBuilder(c.authorizations).Build(from, to, big.NewInt(0), nil).Pack(t.Context(), cli)
		require.Error(t, err, i)
		if c.wantIs != nil {
			require.ErrorIs(t, err, c.wantIs, i)
		}
		require.ErrorContains(t, err, c.wantContains, i)
	}
	require.Zero(t, hits.Load())
}
