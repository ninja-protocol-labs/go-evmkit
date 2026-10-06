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

// Client is a JSON-RPC client. *DefaultClient satisfies it directly; other
// implementations (a retrying/load-balancing wrapper, a test double) can
// stand in for it without depending on this package's HTTP transport.
type Client interface {
	Call(ctx context.Context, e Element) error
	Batch(ctx context.Context, elems Elements) error
}

var _ Client = (*DefaultClient)(nil)

// DefaultClient is a stateless JSON-RPC client over HTTP. The caller
// supplies the request ID on each Element, so a DefaultClient holds no
// per-request state and is safe for concurrent use.
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

// WithTimeout sets the default http.Client's timeout. It has no effect if
// combined with WithHTTPClient, which replaces the client wholesale.
func WithTimeout(d time.Duration) Option {
	return func(c *DefaultClient) {
		c.cli.Timeout = d
	}
}

// WithHeader sets a header sent with every request, such as an API key or
// Authorization token. Calling it again with the same key replaces the value.
func WithHeader(key, value string) Option {
	return func(c *DefaultClient) {
		c.headers.Set(key, value)
	}
}

// WithRetry retries a failed request up to maxRetries times on transient
// failures — network errors and HTTP 429/5xx responses — waiting
// backoff(attempt) between attempts (attempt starts at 0 for the first
// retry). It never retries a well-formed JSON-RPC error response or a
// non-429 4xx: the request already reached the server and failed on its
// own terms, so retrying would just fail the same way again.
func WithRetry(maxRetries int, backoff func(attempt int) time.Duration) Option {
	return func(c *DefaultClient) {
		c.maxRetries = maxRetries
		c.backoff = backoff
	}
}

// ExponentialBackoff returns a backoff function for WithRetry
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

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(c.backoff(attempt)):
		}
	}
}

// doOnce sends body once. retryable reports whether the failure is worth
// retrying: network errors and HTTP 429/5xx are, a malformed request or
// any other 4xx is not.
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
	defer httpResp.Body.Close()

	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, true, fmt.Errorf("rpc: read http response: %w", err)
	}
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		retryable := httpResp.StatusCode == http.StatusTooManyRequests || httpResp.StatusCode >= 500
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

// Batch sends elems as a single JSON-RPC batch request and decodes each
// result into its own Element.Result, matched by ID.
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
