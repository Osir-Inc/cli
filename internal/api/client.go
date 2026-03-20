package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/osir/cli/internal/auth"
	"github.com/osir/cli/internal/config"
)

// APIError is a typed error for API responses with HTTP status >= 400.
type APIError struct {
	StatusCode int
	Body       string
	Method     string
	Path       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (HTTP %d): %s", e.StatusCode, e.Body)
}

// IsRetryable returns true for server errors (5xx) that are worth retrying.
func (e *APIError) IsRetryable() bool {
	return e.StatusCode >= 500
}

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Config     *config.Config
	Session    *auth.Session
	Verbose    bool
	Version    string
}

func NewClient(cfg *config.Config, session *auth.Session) *Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Force IPv4 to avoid ~20s IPv6 timeout on servers without IPv6 routing
			d := &net.Dialer{Timeout: 10 * time.Second}
			return d.DialContext(ctx, "tcp4", addr)
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	return &Client{
		BaseURL: cfg.BackendURL,
		HTTPClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		Config:  cfg,
		Session: session,
	}
}

func (c *Client) Get(ctx context.Context, path string, query url.Values, result any) error {
	return c.doRequest(ctx, http.MethodGet, path, query, nil, result)
}

func (c *Client) Post(ctx context.Context, path string, body any, result any) error {
	return c.doRequest(ctx, http.MethodPost, path, nil, body, result)
}

func (c *Client) Put(ctx context.Context, path string, body any, result any) error {
	return c.doRequest(ctx, http.MethodPut, path, nil, body, result)
}

func (c *Client) Delete(ctx context.Context, path string, result any) error {
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil, result)
}

func (c *Client) PostForm(ctx context.Context, fullURL string, form url.Values, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return c.executeRequest(req, result)
}

func (c *Client) doRequest(ctx context.Context, method, path string, query url.Values, body any, result any) error {
	fullURL := c.BaseURL + path
	if query != nil && len(query) > 0 {
		fullURL += "?" + query.Encode()
	}

	var bodyData []byte
	if body != nil {
		var err error
		bodyData, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
	}

	// Retry loop: up to 3 attempts for retryable errors
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s
			delay := time.Duration(1<<(attempt-1)) * time.Second
			if c.Verbose {
				log.Printf("[RETRY] Attempt %d after %v delay", attempt+1, delay)
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		var bodyReader io.Reader
		if bodyData != nil {
			bodyReader = bytes.NewReader(bodyData)
		}

		req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("Accept", "application/json")
		if c.Version != "" {
			req.Header.Set("User-Agent", "osir-cli/"+c.Version)
		} else {
			req.Header.Set("User-Agent", "osir-cli")
		}

		// Inject auth token if session exists
		if c.Session != nil {
			tokenStart := time.Now()
			token, err := c.Session.GetToken(ctx)
			if c.Verbose {
				log.Printf("[TIMING] GetToken: %v", time.Since(tokenStart))
			}
			if err != nil {
				return err
			}
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
		}

		lastErr = c.executeRequest(req, result)
		if lastErr == nil {
			return nil
		}

		// Only retry idempotent methods (GET, HEAD, OPTIONS) on retryable errors
		// Never retry POST/PUT/DELETE — they may have side effects (e.g. duplicate orders)
		if method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions {
			return lastErr
		}
		if apiErr, ok := lastErr.(*APIError); ok {
			if !apiErr.IsRetryable() {
				return lastErr
			}
		} else if ctx.Err() != nil {
			return lastErr
		}
		// Network errors on GET are retryable — continue loop
	}

	return lastErr
}

func (c *Client) executeRequest(req *http.Request, result any) error {
	if c.Verbose {
		log.Printf("[TIMING] HTTP %s %s ...", req.Method, req.URL)
	}
	httpStart := time.Now()
	resp, err := c.HTTPClient.Do(req)
	if c.Verbose {
		log.Printf("[TIMING] HTTP response: %v", time.Since(httpStart))
	}
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
			Method:     req.Method,
			Path:       req.URL.Path,
		}
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}
