package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	defaultAPIVersion = "v20251201"
	httpTimeout       = 30 * time.Second
	maxRetries        = 3
	circuitThreshold  = 5
	circuitTimeout    = 30 * time.Second

	// OpenAPIBaseURL is the base URL for Shopline Open API (token-scoped)
	OpenAPIBaseURL = "https://open.shopline.io/v1"
)

// Client is the Shopline API client.
type Client struct {
	handle      string
	accessToken string
	apiVersion  string
	BaseURL     string
	httpClient  *http.Client
	useOpenAPI  bool // true for open.shopline.io, false for {handle}.myshopline.com

	mu               sync.RWMutex
	consecutiveFails int
	circuitOpen      bool
	circuitOpenedAt  time.Time
}

// NewClient creates a new API client using the Open API (open.shopline.io).
// This is the recommended client for most Shopline stores.
func NewClient(handle, accessToken string) *Client {
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

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
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

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		// Handle rate limiting
		// Only retry for safe/idempotent methods to avoid duplicate resources
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
			resp.Body.Close() //nolint:errcheck
			isIdempotent := method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
			if isIdempotent && attempt < maxRetries-1 {
				jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
				time.Sleep(retryAfter + jitter)
				continue
			}
			return &RateLimitError{RetryAfter: retryAfter}
		}

		// Handle server errors
		if resp.StatusCode >= 500 {
			c.recordFailure()
			resp.Body.Close() //nolint:errcheck
			if method == http.MethodGet && attempt < maxRetries-1 {
				time.Sleep(time.Second)
				continue
			}
			return &APIError{
				Code:    "SERVER_ERROR",
				Message: "Internal server error",
				Status:  resp.StatusCode,
			}
		}

		c.recordSuccess()

		// Handle client errors
		if resp.StatusCode >= 400 {
			var apiErr APIError
			if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
				resp.Body.Close() //nolint:errcheck
				return &APIError{
					Code:    "UNKNOWN_ERROR",
					Message: "Failed to decode error response",
					Status:  resp.StatusCode,
				}
			}
			resp.Body.Close() //nolint:errcheck
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
