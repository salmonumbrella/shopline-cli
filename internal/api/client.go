package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/debug"
	"github.com/salmonumbrella/shopline-cli/internal/env"
)

const (
	defaultAPIVersion = "v20251201"
	httpTimeout       = 30 * time.Second
	maxRetries        = 3
	circuitThreshold  = 5
	circuitTimeout    = 30 * time.Second
	retryBaseDelay    = 200 * time.Millisecond
	retryMaxDelay     = 2 * time.Second
	retryBudget       = 5 * time.Second
	retryJitter       = 0.2

	// OpenAPIBaseURL is the base URL for Shopline Open API (token-scoped)
	OpenAPIBaseURL = "https://open.shopline.io/v1"
)

type retryConfig struct {
	baseDelay time.Duration
	maxDelay  time.Duration
	budget    time.Duration
	jitter    float64
}

// Client is the Shopline API client.
type Client struct {
	handle      string
	accessToken string
	apiVersion  string
	BaseURL     string
	httpClient  *http.Client
	useOpenAPI  bool // true for open.shopline.io, false for {handle}.myshopline.com
	retry       retryConfig
	debug       *debug.Logger

	mu               sync.RWMutex
	consecutiveFails int
	circuitOpen      bool
	circuitOpenedAt  time.Time
}

// NewClient creates a new API client using the Open API (open.shopline.io).
// This is the recommended client for most Shopline stores.
func NewClient(accessToken string) *Client {
	return NewOpenAPIClient(accessToken)
}

// NewOpenAPIClient creates a client for the Shopline Open API (open.shopline.io/v1).
// The bearer token is scoped to the store, so no handle is needed in the URL.
func NewOpenAPIClient(accessToken string) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	return &Client{
		accessToken: accessToken,
		apiVersion:  "v1",
		BaseURL:     OpenAPIBaseURL,
		useOpenAPI:  true,
		retry:       retryConfigFromEnv(),
		debug:       debugLoggerFromEnv(),
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   httpTimeout,
		},
	}
}

// NewAdminAPIClient creates a client for the Shopline Admin OpenAPI ({handle}.myshopline.com).
// Use this for stores on the myshopline.com domain.
func NewAdminAPIClient(handle, accessToken string) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	return &Client{
		handle:      handle,
		accessToken: accessToken,
		apiVersion:  defaultAPIVersion,
		BaseURL:     fmt.Sprintf("https://%s.myshopline.com/admin/openapi/%s", handle, defaultAPIVersion),
		useOpenAPI:  false,
		retry:       retryConfigFromEnv(),
		debug:       debugLoggerFromEnv(),
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   httpTimeout,
		},
	}
}

// SetUseOpenAPI controls the base URL format.
// When true, uses open.shopline.io (token-scoped, recommended).
// When false, uses {handle}.myshopline.com (legacy Admin API).
func (c *Client) SetUseOpenAPI(use bool) {
	c.useOpenAPI = use
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	return c.do(ctx, http.MethodGet, path, nil, result)
}

// Post performs a POST request.
func (c *Client) Post(ctx context.Context, path string, body, result interface{}) error {
	return c.do(ctx, http.MethodPost, path, body, result)
}

// Put performs a PUT request.
func (c *Client) Put(ctx context.Context, path string, body, result interface{}) error {
	return c.do(ctx, http.MethodPut, path, body, result)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

// DeleteWithBody performs a DELETE request with a JSON body. Some Shopline endpoints
// (bulk deletes, delete images, etc.) require a body.
func (c *Client) DeleteWithBody(ctx context.Context, path string, body, result interface{}) error {
	return c.do(ctx, http.MethodDelete, path, body, result)
}

// Patch performs a PATCH request.
func (c *Client) Patch(ctx context.Context, path string, body, result interface{}) error {
	return c.do(ctx, http.MethodPatch, path, body, result)
}

func (c *Client) do(ctx context.Context, method, path string, body, result interface{}) error {
	if c.isCircuitOpen() {
		return &CircuitBreakerError{}
	}

	// Build URL - Shopline Open API does NOT use .json extensions
	// Unlike Shopify, Shopline endpoints work without .json and some fail with it
	url := c.BaseURL + path

	start := time.Now()
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		attemptStart := time.Now()
		var bodyReader io.Reader
		if body != nil {
			data, err := json.Marshal(body)
			if err != nil {
				return fmt.Errorf("failed to marshal request body: %w", err)
			}
			bodyReader = bytes.NewReader(data)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.accessToken)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("X-Shopline-Access-Token", c.accessToken)
		req.Header.Set("User-Agent", "shopline-cli")

		c.logf("api request method=%s url=%s attempt=%d", method, url, attempt+1)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			c.logf("api error method=%s url=%s attempt=%d duration=%s err=%v", method, url, attempt+1, time.Since(attemptStart), err)
			if c.shouldRetryNetworkError(ctx, method, attempt, start, err) {
				continue
			}
			break
		}
		c.logf("api response method=%s url=%s status=%d duration=%s", method, url, resp.StatusCode, time.Since(attemptStart))

		// Handle rate limiting
		// Only retry for safe/idempotent methods to avoid duplicate resources
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
			resp.Body.Close() //nolint:errcheck
			isIdempotent := method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
			if isIdempotent && attempt < maxRetries-1 {
				jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
				c.logf("api rate limited retryAfter=%s", retryAfter)
				time.Sleep(retryAfter + jitter)
				continue
			}
			return &RateLimitError{RetryAfter: retryAfter}
		}

		// Handle server errors
		if resp.StatusCode >= 500 {
			c.recordFailure()
			var serverBody []byte
			if attempt >= maxRetries-1 || method != http.MethodGet {
				serverBody, _ = io.ReadAll(io.LimitReader(resp.Body, 512))
			}
			resp.Body.Close() //nolint:errcheck
			if method == http.MethodGet && attempt < maxRetries-1 {
				c.logf("api server error status=%d attempt=%d", resp.StatusCode, attempt+1)
				time.Sleep(time.Second)
				continue
			}
			msg := http.StatusText(resp.StatusCode)
			if len(serverBody) > 0 && !bytes.Contains(serverBody, []byte("<html")) && !bytes.Contains(serverBody, []byte("<HTML")) {
				snippet := string(serverBody)
				if runes := []rune(snippet); len(runes) > 200 {
					snippet = string(runes[:200]) + "..."
				}
				msg += ": " + snippet
			}
			return &APIError{
				Code:    fmt.Sprintf("HTTP_%d", resp.StatusCode),
				Message: msg,
				Status:  resp.StatusCode,
			}
		}

		c.recordSuccess()

		// Handle client errors
		if resp.StatusCode >= 400 {
			var apiErr APIError
			bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
			resp.Body.Close() //nolint:errcheck

			if err := json.Unmarshal(bodyBytes, &apiErr); err != nil {
				msg := http.StatusText(resp.StatusCode)
				if len(bodyBytes) > 0 && !bytes.Contains(bodyBytes, []byte("<html")) && !bytes.Contains(bodyBytes, []byte("<HTML")) {
					snippet := string(bodyBytes)
					if runes := []rune(snippet); len(runes) > 200 {
						snippet = string(runes[:200]) + "..."
					}
					msg += ": " + snippet
				}
				return &APIError{
					Code:    fmt.Sprintf("HTTP_%d", resp.StatusCode),
					Message: msg,
					Status:  resp.StatusCode,
				}
			}
			apiErr.Status = resp.StatusCode
			return &apiErr
		}

		// Decode successful response
		if result != nil && resp.StatusCode != http.StatusNoContent {
			if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
				if err != io.EOF {
					resp.Body.Close() //nolint:errcheck
					return fmt.Errorf("failed to decode response: %w", err)
				}
			}
		}
		resp.Body.Close() //nolint:errcheck

		return nil
	}

	return lastErr
}

func (c *Client) isCircuitOpen() bool {
	c.mu.RLock()
	if !c.circuitOpen {
		c.mu.RUnlock()
		return false
	}

	if time.Since(c.circuitOpenedAt) <= circuitTimeout {
		c.mu.RUnlock()
		return true
	}

	c.mu.RUnlock()

	c.mu.Lock()
	if c.circuitOpen && time.Since(c.circuitOpenedAt) > circuitTimeout {
		c.circuitOpen = false
		c.consecutiveFails = 0
	}
	c.mu.Unlock()

	return false
}

func (c *Client) recordFailure() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.consecutiveFails++
	if c.consecutiveFails >= circuitThreshold {
		c.circuitOpen = true
		c.circuitOpenedAt = time.Now()
	}
}

func (c *Client) recordSuccess() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.consecutiveFails = 0
	c.circuitOpen = false
}

func parseRetryAfter(header string) time.Duration {
	if header == "" {
		return time.Second
	}

	if seconds, err := strconv.Atoi(header); err == nil {
		return time.Duration(seconds) * time.Second
	}

	if t, err := time.Parse(time.RFC1123, header); err == nil {
		until := time.Until(t)
		if until > 0 {
			return until
		}
		return 0
	}

	return time.Second
}

func (c *Client) logf(format string, args ...interface{}) {
	if c.debug == nil {
		return
	}
	c.debug.Printf(format, args...)
}

func (c *Client) shouldRetryNetworkError(ctx context.Context, method string, attempt int, start time.Time, err error) bool {
	if !isIdempotentMethod(method) || attempt >= maxRetries-1 {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	delay := c.retryDelay(attempt)
	if !c.withinRetryBudget(start, delay) {
		c.logf("api retry budget_exceeded=true")
		return false
	}
	if delay > 0 {
		c.logf("api retry delay=%s", delay)
		time.Sleep(delay)
	}
	return true
}

func (c *Client) retryDelay(attempt int) time.Duration {
	if c.retry.baseDelay <= 0 {
		return 0
	}
	multiplier := time.Duration(1 << attempt)
	delay := c.retry.baseDelay * multiplier
	if c.retry.maxDelay > 0 && delay > c.retry.maxDelay {
		delay = c.retry.maxDelay
	}
	if c.retry.jitter > 0 {
		jitterRange := c.retry.jitter * float64(delay)
		jitter := (rand.Float64()*2 - 1) * jitterRange
		delay = time.Duration(float64(delay) + jitter)
		if delay < 0 {
			return 0
		}
	}
	return delay
}

func (c *Client) withinRetryBudget(start time.Time, delay time.Duration) bool {
	if c.retry.budget == 0 {
		return false // 0 explicitly disables retries
	}
	if c.retry.budget < 0 {
		return true // negative means unlimited
	}
	return time.Since(start)+delay <= c.retry.budget
}

func isIdempotentMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPut, http.MethodDelete:
		return true
	default:
		return false
	}
}

func retryConfigFromEnv() retryConfig {
	cfg := retryConfig{
		baseDelay: retryBaseDelay,
		maxDelay:  retryMaxDelay,
		budget:    retryBudget,
		jitter:    retryJitter,
	}
	if val := strings.TrimSpace(os.Getenv("SHOPLINE_RETRY_BASE")); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			cfg.baseDelay = d
		}
	}
	if val := strings.TrimSpace(os.Getenv("SHOPLINE_RETRY_MAX")); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			cfg.maxDelay = d
		}
	}
	if val := strings.TrimSpace(os.Getenv("SHOPLINE_RETRY_BUDGET")); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			cfg.budget = d
		}
	}
	if val := strings.TrimSpace(os.Getenv("SHOPLINE_RETRY_JITTER")); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			if f < 0 {
				f = 0
			}
			if f > 1 {
				f = 1
			}
			cfg.jitter = f
		}
	}
	return cfg
}

func debugLoggerFromEnv() *debug.Logger {
	if env.Bool("SHOPLINE_DEBUG") {
		return debug.New(os.Stderr)
	}
	return debug.Nop()
}
