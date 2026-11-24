package rpc

import (
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestStateOverrideMarshalJSONGolden(t *testing.T) {
	first, err := types.NewAddressFromHex("0x1111111111111111111111111111111111111111")
	require.NoError(t, err)
	second, err := types.NewAddressFromHex("0xAbCdEf0123456789aBcDeF0123456789AbCdEf01")
	require.NoError(t, err)

	var slot1, slot2, value42, valueMax [32]byte
	slot1[31] = 0x01
	slot2[31] = 0x02
	value42[31] = 0x2a
	for i := range valueMax {
		valueMax[i] = 0xff
	}

	override := NewStateOverride().
		WithBalance(first, big.NewInt(1_000_000_000_000_000_000)).
		WithNonce(first, 7).
		WithCode(first, []byte{0x60, 0x00}).
		WithStateDiff(first, slot2, valueMax).
		WithStateDiff(first, slot1, value42).
		WithState(second, slot1, value42)

	out, err := json.Marshal(override)
	require.NoError(t, err)

	zeros31 := strings.Repeat("00", 31)
	want := `{"0x1111111111111111111111111111111111111111":{` +
		`"balance":"0xde0b6b3a7640000",` +
		`"code":"0x6000",` +
		`"nonce":"0x7",` +
		`"stateDiff":{` +
		`"0x` + zeros31 + `01":"0x` + zeros31 + `2a",` +
		`"0x` + zeros31 + `02":"0x` + strings.Repeat("ff", 32) + `"}},` +
		`"0xabcdef0123456789abcdef0123456789abcdef01":{` +
		`"state":{"0x` + zeros31 + `01":"0x` + zeros31 + `2a"}}}`
	require.Equal(t, want, string(out))

	again, err := json.Marshal(override)
	require.NoError(t, err)
	require.Equal(t, string(out), string(again))
}

func TestStateOverrideMarshalJSONEdgeValues(t *testing.T) {
	account, err := types.NewAddressFromHex("0x2222222222222222222222222222222222222222")
	require.NoError(t, err)
	maxUint256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))

	cases := []struct {
		override *StateOverride
		want     string
	}{
		{
			NewStateOverride(),
			`{}`,
		},
		{
			NewStateOverride().WithBalance(account, big.NewInt(0)),
			`{"0x2222222222222222222222222222222222222222":{"balance":"0x0"}}`,
		},
		{
			NewStateOverride().WithBalance(account, maxUint256),
			`{"0x2222222222222222222222222222222222222222":{"balance":"0x` + strings.Repeat("f", 64) + `"}}`,
		},
		{
			NewStateOverride().WithNonce(account, 0),
			`{"0x2222222222222222222222222222222222222222":{"nonce":"0x0"}}`,
		},
		{
			NewStateOverride().WithNonce(account, ^uint64(0)),
			`{"0x2222222222222222222222222222222222222222":{"nonce":"0xffffffffffffffff"}}`,
		},
		{
			NewStateOverride().WithCode(account, []byte{}),
			`{"0x2222222222222222222222222222222222222222":{"code":"0x"}}`,
		},
		{
			NewStateOverride().WithCode(account, []byte{0x60, 0x01}).WithCode(account, nil),
			`{"0x2222222222222222222222222222222222222222":{}}`,
		},
		{
			NewStateOverride().WithBalance(account, big.NewInt(5)).WithBalance(account, nil),
			`{"0x2222222222222222222222222222222222222222":{}}`,
		},
		{
			NewStateOverride().WithBalance(account, big.NewInt(5)).WithBalance(account, big.NewInt(9)),
			`{"0x2222222222222222222222222222222222222222":{"balance":"0x9"}}`,
		},
		{
			NewStateOverride().WithNonce(account, 1).WithNonce(account, 2),
			`{"0x2222222222222222222222222222222222222222":{"nonce":"0x2"}}`,
		},
		{
			new(StateOverride).WithNonce(account, 3),
			`{"0x2222222222222222222222222222222222222222":{"nonce":"0x3"}}`,
		},
	}
	for i, c := range cases {
		out, err := json.Marshal(c.override)
		require.NoError(t, err, i)
		require.Equal(t, c.want, string(out), i)
	}
}

func TestStateOverrideMergesSameAccountAcrossCasing(t *testing.T) {
	lower, err := types.NewAddressFromHex("0xabcdef0123456789abcdef0123456789abcdef01")
	require.NoError(t, err)
	upper, err := types.NewAddressFromHex("0xABCDEF0123456789ABCDEF0123456789ABCDEF01")
	require.NoError(t, err)

	var slot, first, second [32]byte
	slot[31] = 0x05
	first[31] = 0x01
	second[31] = 0x02

	override := NewStateOverride().
		WithStateDiff(lower, slot, first).
		WithBalance(upper, big.NewInt(3)).
		WithStateDiff(upper, slot, second)

	out, err := json.Marshal(override)
	require.NoError(t, err)

	var decoded map[string]map[string]any
	require.NoError(t, json.Unmarshal(out, &decoded))
	require.Len(t, decoded, 1)
	account := decoded["0xabcdef0123456789abcdef0123456789abcdef01"]
	require.Equal(t, "0x3", account["balance"])
	require.Equal(t, map[string]any{
		"0x" + strings.Repeat("00", 31) + "05": "0x" + strings.Repeat("00", 31) + "02",
	}, account["stateDiff"])
}

func TestStateOverrideCopiesInputs(t *testing.T) {
	account, err := types.NewAddressFromHex("0x3333333333333333333333333333333333333333")
	require.NoError(t, err)
	balance := big.NewInt(10)
	code := []byte{0x60, 0x01}

	override := NewStateOverride().WithBalance(account, balance).WithCode(account, code)
	balance.SetInt64(99)
	code[1] = 0xff
	account.SetBytes(make([]byte, 20))

	out, err := json.Marshal(override)
	require.NoError(t, err)
	require.Equal(t, `{"0x3333333333333333333333333333333333333333":{"balance":"0xa","code":"0x6001"}}`, string(out))
}

func TestStateOverrideValidate(t *testing.T) {
	account, err := types.NewAddressFromHex("0x4444444444444444444444444444444444444444")
	require.NoError(t, err)
	var slot, value [32]byte
	tooBig := new(big.Int).Lsh(big.NewInt(1), 256)
	maxUint256 := new(big.Int).Sub(tooBig, big.NewInt(1))

	cases := []struct {
		override *StateOverride
		wantErr  string
	}{
		{
			nil,
			"",
		},
		{
			NewStateOverride(),
			"",
		},
		{
			new(StateOverride),
			"",
		},
		{
			NewStateOverride().
				WithBalance(account, maxUint256).
				WithNonce(account, 1).
				WithCode(account, []byte{0x00}).
				WithStateDiff(account, slot, value),
			"",
		},
		{
			NewStateOverride().WithState(account, slot, value),
			"",
		},
		{
			NewStateOverride().WithBalance(nil, big.NewInt(1)),
			"account is nil",
		},
		{
			NewStateOverride().WithNonce(nil, 1),
			"account is nil",
		},
		{
			NewStateOverride().WithCode(nil, []byte{}),
			"account is nil",
		},
		{
			NewStateOverride().WithState(nil, slot, value),
			"account is nil",
		},
		{
			NewStateOverride().WithStateDiff(nil, slot, value),
			"account is nil",
		},
		{
			NewStateOverride().WithBalance(account, big.NewInt(-1)),
			"0x4444444444444444444444444444444444444444.balance is negative",
		},
		{
			NewStateOverride().WithBalance(account, tooBig),
			"0x4444444444444444444444444444444444444444.balance exceeds uint256",
		},
		{
			NewStateOverride().WithState(account, slot, value).WithStateDiff(account, slot, value),
			"0x4444444444444444444444444444444444444444: state and stateDiff are both set",
		},
	}
	for i, c := range cases {
		err := c.override.Validate()
		if c.wantErr == "" {
			require.NoError(t, err, i)
			continue
		}
		require.ErrorIs(t, err, ErrInvalidStateOverride, i)
		require.ErrorContains(t, err, c.wantErr, i)

		_, err = json.Marshal(c.override)
		require.ErrorIs(t, err, ErrInvalidStateOverride, i)
	}
}

func TestStateOverrideValidateReportsLowestAccountFirst(t *testing.T) {
	low, err := types.NewAddressFromHex("0x0000000000000000000000000000000000000001")
	require.NoError(t, err)
	high, err := types.NewAddressFromHex("0xffffffffffffffffffffffffffffffffffffffff")
	require.NoError(t, err)

	override := NewStateOverride().
		WithBalance(high, big.NewInt(-1)).
		WithBalance(low, big.NewInt(-1))
	for range 20 {
		require.ErrorContains(t, override.Validate(), "0x0000000000000000000000000000000000000001.balance is negative")
	}
}

func TestETHCallAndEstimateGasParams(t *testing.T) {
	account, err := types.NewAddressFromHex("0x5555555555555555555555555555555555555555")
	require.NoError(t, err)
	override := NewStateOverride().WithNonce(account, 1)
	params := map[string]any{"to": account.String()}
	var result string

	cases := []struct {
		elem       Element
		method     string
		wantParams []any
	}{
		{
			ETHCall("1", params, BlockTagLatest, nil, &result),
			MethodETHCall,
			[]any{params, BlockTagLatest},
		},
		{
			ETHCall("2", params, "0x10", override, &result),
			MethodETHCall,
			[]any{params, "0x10", override},
		},
		{
			ETHEstimateGas("3", params, BlockTagPending, nil, &result),
			MethodETHEstimateGas,
			[]any{params, BlockTagPending},
		},
		{
			ETHEstimateGas("4", params, BlockTagLatest, override, &result),
			MethodETHEstimateGas,
			[]any{params, BlockTagLatest, override},
		},
	}
	for i, c := range cases {
		require.Equal(t, c.method, c.elem.Method, i)
		require.Equal(t, c.wantParams, c.elem.Params, i)
		require.Same(t, &result, c.elem.Result, i)
	}
}

func TestClientSendsStateOverrideAsThirdParam(t *testing.T) {
	account, err := types.NewAddressFromHex("0x6666666666666666666666666666666666666666")
	require.NoError(t, err)
	var slot, value [32]byte
	slot[31] = 0x01
	value[31] = 0x2a
	override := NewStateOverride().WithCode(account, []byte{0x00}).WithStateDiff(account, slot, value)
	wantOverride, err := json.Marshal(override)
	require.NoError(t, err)

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
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"` + req.ID + `","result":"0x2a"}`))
	}))
	defer srv.Close()
	cli := NewClient(srv.URL)
	params := map[string]any{"to": account.String()}

	var results [4]string
	cases := []struct {
		elem      Element
		method    string
		block     string
		withThird bool
	}{
		{
			ETHCall("a", params, BlockTagLatest, override, &results[0]),
			MethodETHCall,
			BlockTagLatest,
			true,
		},
		{
			ETHCall("b", params, "0x1b4", nil, &results[1]),
			MethodETHCall,
			"0x1b4",
			false,
		},
		{
			ETHEstimateGas("c", params, BlockTagPending, override, &results[2]),
			MethodETHEstimateGas,
			BlockTagPending,
			true,
		},
		{
			ETHEstimateGas("d", params, BlockTagLatest, nil, &results[3]),
			MethodETHEstimateGas,
			BlockTagLatest,
			false,
		},
	}
	for i, c := range cases {
		require.NoError(t, cli.Call(t.Context(), c.elem), i)
		require.Equal(t, "0x2a", results[i], i)
		require.Equal(t, c.method, gotMethod, i)

		var gotCall map[string]any
		require.NoError(t, json.Unmarshal(gotParams[0], &gotCall), i)
		require.Equal(t, params, gotCall, i)
		var gotBlock string
		require.NoError(t, json.Unmarshal(gotParams[1], &gotBlock), i)
		require.Equal(t, c.block, gotBlock, i)

		if !c.withThird {
			require.Len(t, gotParams, 2, i)
			continue
		}
		require.Len(t, gotParams, 3, i)
		require.JSONEq(t, string(wantOverride), string(gotParams[2]), i)
	}
}

func TestClientRejectsInvalidStateOverrideWithoutRPC(t *testing.T) {
	account, err := types.NewAddressFromHex("0x7777777777777777777777777777777777777777")
	require.NoError(t, err)
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"x","result":"0x"}`))
	}))
	defer srv.Close()
	cli := NewClient(srv.URL)
	params := map[string]any{"to": account.String()}
	var slot, value [32]byte

	invalid := []*StateOverride{
		NewStateOverride().WithBalance(nil, big.NewInt(1)),
		NewStateOverride().WithBalance(account, big.NewInt(-1)),
		NewStateOverride().WithBalance(account, new(big.Int).Lsh(big.NewInt(1), 256)),
		NewStateOverride().WithState(account, slot, value).WithStateDiff(account, slot, value),
	}
	for i, override := range invalid {
		var result string
		err := cli.Call(t.Context(), ETHCall("call", params, BlockTagLatest, override, &result))
		require.ErrorIs(t, err, ErrInvalidStateOverride, i)

		err = cli.Call(t.Context(), ETHEstimateGas("gas", params, BlockTagLatest, override, &result))
		require.ErrorIs(t, err, ErrInvalidStateOverride, i)

		var plain string
		err = cli.Batch(t.Context(), Elements{
			ETHCall("plain", params, BlockTagLatest, nil, &plain),
			ETHCall("bad", params, BlockTagLatest, override, &result),
		})
		require.ErrorIs(t, err, ErrInvalidStateOverride, i)
	}
	require.Zero(t, hits.Load())
}
