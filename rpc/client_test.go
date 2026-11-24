package rpc

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClientSurfacesResponseErrorFromNon2xxBody(t *testing.T) {
	cases := []struct {
		status    int
		body      string
		retries   int
		wantCode  int
		wantHits  int32
		wantError string
	}{
		{
			status:    http.StatusBadRequest,
			body:      `{"jsonrpc":"2.0","id":"x","error":{"code":-32600,"message":"Unsupported method: evmkit_doesNotExist"}}`,
			retries:   3,
			wantCode:  -32600,
			wantHits:  1,
			wantError: "rpc: http status 400: rpc error -32600: Unsupported method: evmkit_doesNotExist",
		},
		{
			status:    http.StatusForbidden,
			body:      `{"jsonrpc":"2.0","id":"x","error":{"code":-32601,"message":"rpc method is not whitelisted"}}`,
			retries:   3,
			wantCode:  -32601,
			wantHits:  1,
			wantError: "rpc: http status 403: rpc error -32601: rpc method is not whitelisted",
		},
		{
			status:    http.StatusTooManyRequests,
			body:      `{"jsonrpc":"2.0","id":"x","error":{"code":429,"message":"rate limited","data":"slow down"}}`,
			retries:   2,
			wantCode:  429,
			wantHits:  3,
			wantError: `rpc: http status 429: rpc error 429: rate limited: "slow down"`,
		},
		{
			status:    http.StatusServiceUnavailable,
			body:      `{"jsonrpc":"2.0","id":"x","error":{"code":-32000,"message":"backend down"}}`,
			retries:   1,
			wantCode:  -32000,
			wantHits:  2,
			wantError: "rpc: http status 503: rpc error -32000: backend down",
		},
	}
	for i, c := range cases {
		var hits atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hits.Add(1)
			w.WriteHeader(c.status)
			_, _ = w.Write([]byte(c.body))
		}))
		cli := NewClient(srv.URL, WithRetry(c.retries, func(int) time.Duration { return 0 }))

		var result string
		callErr := cli.Call(t.Context(), ETHChainID("x", &result))
		re, ok := errors.AsType[*ResponseError](callErr)
		require.True(t, ok, "%d: %v", i, callErr)
		require.Equal(t, c.wantCode, re.Code, i)
		require.EqualError(t, callErr, c.wantError, i)
		require.Equal(t, c.wantHits, hits.Load(), i)

		hits.Store(0)
		batchErr := cli.Batch(t.Context(), Elements{ETHChainID("x", &result), ETHBlockNumber("y", &result)})
		re, ok = errors.AsType[*ResponseError](batchErr)
		require.True(t, ok, "%d: %v", i, batchErr)
		require.Equal(t, c.wantCode, re.Code, i)
		require.Equal(t, c.wantHits, hits.Load(), i)
		srv.Close()
	}
}

func TestClientKeepsRawBodyWhenNon2xxIsNotResponseError(t *testing.T) {
	bodies := []string{
		`upstream connect error`,
		`{"jsonrpc":"2.0","id":"x","result":"0x1"}`,
		`[{"jsonrpc":"2.0","id":"x","error":{"code":-32600,"message":"bad"}}]`,
		`{"error":"not an object"}`,
		``,
	}
	for i, body := range bodies {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(body))
		}))
		cli := NewClient(srv.URL)

		var result string
		err := cli.Call(t.Context(), ETHChainID("x", &result))
		require.EqualError(t, err, "rpc: http status 400: "+body, i)
		_, ok := errors.AsType[*ResponseError](err)
		require.False(t, ok, i)
		srv.Close()
	}
}

func TestClientRetriesWithoutWaitingWhenBackoffIsNilOrNonPositive(t *testing.T) {
	cases := []struct {
		backoff   func(int) time.Duration
		retries   int
		failures  int32
		wantHits  int32
		wantError string
	}{
		{nil, 2, 2, 3, ""},
		{nil, 1, 2, 2, "rpc: http status 503: upstream down"},
		{nil, 0, 1, 1, "rpc: http status 503: upstream down"},
		{func(int) time.Duration { return 0 }, 3, 3, 4, ""},
		{func(int) time.Duration { return -time.Second }, 2, 2, 3, ""},
	}
	for i, c := range cases {
		var hits atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if hits.Add(1) <= c.failures {
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte("upstream down"))
				return
			}
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"x","result":"0x1"}`))
		}))
		cli := NewClient(srv.URL, WithRetry(c.retries, c.backoff))

		var result string
		var err error
		require.NotPanics(t, func() { err = cli.Call(t.Context(), ETHChainID("x", &result)) }, i)
		require.Equal(t, c.wantHits, hits.Load(), i)
		if c.wantError == "" {
			require.NoError(t, err, i)
			require.Equal(t, "0x1", result, i)
		} else {
			require.EqualError(t, err, c.wantError, i)
		}
		srv.Close()
	}
}

func TestClientStopsImmediateRetriesOnceContextIsDone(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		cancel()
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	cli := NewClient(srv.URL, WithRetry(1_000, nil))

	var result string
	err := cli.Call(ctx, ETHChainID("x", &result))
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, int32(1), hits.Load())
}
