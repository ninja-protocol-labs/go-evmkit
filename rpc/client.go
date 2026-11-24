package rpc

import (
	"bytes"
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a JSON-RPC client; wrappers and test doubles can stand in for *DefaultClient.
type Client interface {
	Call(ctx context.Context, e Element) error
	Batch(ctx context.Context, elems Elements) error
}

var _ Client = (*DefaultClient)(nil)

// DefaultClient is a stateless JSON-RPC client over HTTP, safe for concurrent use.
type DefaultClient struct {
	url        string
	cli        *http.Client
	headers    http.Header
	maxRetries int
	backoff    func(attempt int) time.Duration
}

// Option configures a DefaultClient.
type Option func(*DefaultClient)

// WithHTTPClient overrides the default http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *DefaultClient) {
		c.cli = hc
	}
}

// WithTimeout sets the default http.Client's timeout; WithHTTPClient overrides it.
func WithTimeout(d time.Duration) Option {
	return func(c *DefaultClient) {
		c.cli.Timeout = d
	}
}

// WithHeader sets a header sent with every request, replacing any earlier value for key.
func WithHeader(key, value string) Option {
	return func(c *DefaultClient) {
		c.headers.Set(key, value)
	}
}

// WithRetry retries network errors and HTTP 429/5xx up to maxRetries times, waiting backoff(attempt); a nil or non-positive backoff retries immediately.
func WithRetry(maxRetries int, backoff func(attempt int) time.Duration) Option {
	return func(c *DefaultClient) {
		c.maxRetries = maxRetries
		c.backoff = backoff
	}
}

// ExponentialBackoff returns a WithRetry backoff doubling from base, capped at max.
func ExponentialBackoff(base, max time.Duration) func(attempt int) time.Duration {
	return func(attempt int) time.Duration {
		d := base
		for range attempt {
			d *= 2
			if d >= max {
				return max
			}
		}
		return d
	}
}

// NewClient returns a DefaultClient that sends requests to url.
func NewClient(url string, opts ...Option) *DefaultClient {
	c := &DefaultClient{
		url: url,
		cli: &http.Client{
			Timeout: 15 * time.Second,
		},
		headers: make(http.Header),
	}
	c.headers.Set("Content-Type", "application/json")

	for _, opt := range opts {
		opt(c)
	}
	return c
}

type response struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      string         `json:"id"`
	Result  jsontext.Value `json:"result"`
	Error   *ResponseError `json:"error,omitempty"`
}

// ResponseError is a JSON-RPC error object.
type ResponseError struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    jsontext.Value `json:"data,omitempty"`
}

// Error implements the error interface.
func (e *ResponseError) Error() string {
	if len(e.Data) == 0 {
		return fmt.Sprintf("rpc error %d: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("rpc error %d: %s: %s", e.Code, e.Message, string(e.Data))
}

func (c *DefaultClient) do(ctx context.Context, body []byte) ([]byte, error) {
	var lastErr error
	for attempt := 0; ; attempt++ {
		respBytes, retryable, err := c.doOnce(ctx, body)
		if err == nil {
			return respBytes, nil
		}
		lastErr = err
		if attempt >= c.maxRetries || !retryable {
			return nil, lastErr
		}

		var delay time.Duration
		if c.backoff != nil {
			delay = c.backoff(attempt)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}
}

// doOnce sends body once and reports whether a failure is retryable (network errors, HTTP 429/5xx).
func (c *DefaultClient) doOnce(ctx context.Context, body []byte) ([]byte, bool, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return nil, false, fmt.Errorf("rpc: create http request: %w", err)
	}
	httpReq.Header = c.headers.Clone()

	httpResp, err := c.cli.Do(httpReq)
	if err != nil {
		return nil, true, fmt.Errorf("rpc: send http request: %w", err)
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, true, fmt.Errorf("rpc: read http response: %w", err)
	}
	if c.isError(httpResp.StatusCode) {
		retryable := c.isRetryable(httpResp.StatusCode)
		var resp response
		if err = json.Unmarshal(respBytes, &resp); err == nil && resp.Error != nil {
			return nil, retryable, fmt.Errorf("rpc: http status %d: %w", httpResp.StatusCode, resp.Error)
		}

		return nil, retryable, fmt.Errorf("rpc: http status %d: %s", httpResp.StatusCode, string(respBytes))
	}
	return respBytes, false, nil
}

// Call sends a single JSON-RPC request and decodes the result into e.Result.
func (c *DefaultClient) Call(ctx context.Context, e Element) error {
	req, err := json.Marshal(e.toRequest())
	if err != nil {
		return fmt.Errorf("rpc: marshal request: %w", err)
	}

	b, err := c.do(ctx, req)
	if err != nil {
		return err
	}

	var resp response
	if err = json.Unmarshal(b, &resp); err != nil {
		return fmt.Errorf("rpc: decode response: %w: %s", err, string(b))
	}
	if resp.Error != nil {
		return resp.Error
	}
	if e.Result == nil {
		return nil
	}
	if err = json.Unmarshal(resp.Result, e.Result); err != nil {
		return fmt.Errorf("rpc: decode result for %s: %w: %s", e.Method, err, string(resp.Result))
	}
	return nil
}

// Batch sends elems as one JSON-RPC batch and decodes each result into its Element by ID.
func (c *DefaultClient) Batch(ctx context.Context, elems Elements) error {
	n := elems.Len()
	if n == 0 {
		return nil
	}

	reqs := make([]request, n)
	for i := range n {
		reqs[i] = elems[i].toRequest()
	}

	reqBody, err := json.Marshal(reqs)
	if err != nil {
		return fmt.Errorf("rpc: marshal batch request: %w", err)
	}

	respBytes, err := c.do(ctx, reqBody)
	if err != nil {
		return err
	}

	var resps []response
	if err = json.Unmarshal(respBytes, &resps); err != nil {
		return fmt.Errorf("rpc: decode batch response: %w: %s", err, string(respBytes))
	}

	byID := make(map[string]response, len(resps))
	for _, resp := range resps {
		byID[resp.ID] = resp
	}

	for i := range n {
		id := elems.GetID(i)
		resp, ok := byID[id]
		if !ok {
			return fmt.Errorf("rpc: missing response for %s (id %s)", elems.GetMethod(i), id)
		}
		if resp.Error != nil {
			return fmt.Errorf("rpc: %s: %w", elems.GetMethod(i), resp.Error)
		}
		if elems.GetResult(i) == nil {
			continue
		}
		if err = json.Unmarshal(resp.Result, elems.GetResult(i)); err != nil {
			return fmt.Errorf("rpc: decode result for %s: %w: %s", elems.GetMethod(i), err, string(resp.Result))
		}
	}
	return nil
}

// isError reports whether statusCode is outside the 2xx range.
func (*DefaultClient) isError(statusCode int) bool {
	return statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices
}

// isRetryable reports whether a failed status is worth retrying: 429 and 5xx.
func (*DefaultClient) isRetryable(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || statusCode >= http.StatusInternalServerError
}
