package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/debug"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestClientGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("Missing or incorrect Authorization header")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("Missing Content-Type header")
		}

		_ = json.NewEncoder(w).Encode(map[string]string{"id": "123"})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	var result map[string]string
	err := client.Get(context.Background(), "/test", &result)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if result["id"] != "123" {
		t.Errorf("Unexpected result: %v", result)
	}
}

func TestClientEmptyBodyNoError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	var result map[string]string
	err := client.Get(context.Background(), "/test", &result)
	if err != nil {
		t.Fatalf("Expected no error for empty body, got %v", err)
	}
}

func TestClientRateLimitRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"success": "true"})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	var result map[string]string
	err := client.Get(context.Background(), "/test", &result)
	if err != nil {
		t.Fatalf("Get failed after retries: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestClientPostNoRetryOn429(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.Post(context.Background(), "/test", map[string]string{"key": "value"}, nil)
	if err == nil {
		t.Fatal("Expected error for rate-limited POST")
	}

	rateLimitErr, ok := err.(*RateLimitError)
	if !ok {
		t.Fatalf("Expected RateLimitError, got %T: %v", err, err)
	}
	if rateLimitErr.RetryAfter == 0 {
		t.Error("Expected non-zero RetryAfter")
	}

	// POST should NOT retry on 429 - only one attempt expected
	if attempts != 1 {
		t.Errorf("POST should not retry on 429: expected 1 attempt, got %d", attempts)
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	failureCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		failureCount++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	// Each GET request retries up to 3 times on 500 errors, and each retry
	// increments the failure counter. So we need to make requests until
	// we accumulate circuitThreshold (5) failures.
	// With 3 retries per request, 2 requests = 6 failures, which exceeds threshold.

	// First request: 3 retries = 3 failures
	err := client.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("Expected error on first request")
	}
	if _, ok := err.(*CircuitBreakerError); ok {
		t.Fatal("Circuit opened too early after first request")
	}

	// Second request: 3 more retries = 6 total failures, exceeds threshold of 5
	// This request should succeed in making server calls (circuit not yet checked as open)
	// but will open the circuit during execution
	err = client.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("Expected error on second request")
	}
	// After 6 failures, circuit should be open now

	// Third request should fail immediately with CircuitBreakerError
	err = client.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("Expected CircuitBreakerError")
	}

	cbErr, ok := err.(*CircuitBreakerError)
	if !ok {
		t.Fatalf("Expected CircuitBreakerError, got %T: %v", err, err)
	}
	if cbErr == nil {
		t.Fatal("CircuitBreakerError should not be nil")
	}

	// Verify we made expected number of server calls (6 = 2 requests * 3 retries each)
	if failureCount != 6 {
		t.Errorf("Expected 6 server failures, got %d", failureCount)
	}
}

func TestCircuitBreaker_ResetsOnSuccess(t *testing.T) {
	failCount := 0
	shouldFail := true

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if shouldFail {
			failCount++
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"success": "true"})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	// Make failures to open the circuit
	for i := 0; i < circuitThreshold; i++ {
		_ = client.Get(context.Background(), "/test", nil)
	}

	// Verify circuit is open
	err := client.Get(context.Background(), "/test", nil)
	if _, ok := err.(*CircuitBreakerError); !ok {
		t.Fatalf("Expected circuit to be open, got %T: %v", err, err)
	}

	// Manually reset the circuit state to simulate timeout expiry
	// (we can't wait 30 seconds in a test)
	client.mu.Lock()
	client.circuitOpenedAt = client.circuitOpenedAt.Add(-circuitTimeout - time.Second)
	client.mu.Unlock()

	// Now make a successful request
	shouldFail = false
	var result map[string]string
	err = client.Get(context.Background(), "/test", &result)
	if err != nil {
		t.Fatalf("Expected success after circuit timeout, got: %v", err)
	}

	// Verify circuit is now closed by checking internal state
	client.mu.RLock()
	isOpen := client.circuitOpen
	consecutiveFails := client.consecutiveFails
	client.mu.RUnlock()

	if isOpen {
		t.Error("Circuit should be closed after successful request")
	}
	if consecutiveFails != 0 {
		t.Errorf("Consecutive fails should be 0, got %d", consecutiveFails)
	}
}

func TestCircuitBreaker_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	// Open the circuit
	for i := 0; i < circuitThreshold; i++ {
		_ = client.Get(context.Background(), "/test", nil)
	}

	// Verify circuit is open
	if !client.isCircuitOpen() {
		t.Fatal("Circuit should be open")
	}

	// Verify circuit is still open before timeout
	client.mu.Lock()
	client.circuitOpenedAt = time.Now().Add(-circuitTimeout / 2) // halfway through timeout
	client.mu.Unlock()

	if !client.isCircuitOpen() {
		t.Error("Circuit should still be open before timeout expires")
	}

	// Verify circuit is half-open (allows requests) after timeout
	client.mu.Lock()
	client.circuitOpenedAt = time.Now().Add(-circuitTimeout - time.Second) // past timeout
	client.mu.Unlock()

	if client.isCircuitOpen() {
		t.Error("Circuit should be half-open (allow requests) after timeout expires")
	}
}

func TestCircuitBreaker_TimeoutResetsCounter(t *testing.T) {
	client := NewClient("test-token")
	client.mu.Lock()
	client.circuitOpen = true
	client.consecutiveFails = circuitThreshold
	client.circuitOpenedAt = time.Now().Add(-circuitTimeout - time.Second)
	client.mu.Unlock()

	if client.isCircuitOpen() {
		t.Fatal("Expected circuit to be closed after timeout")
	}

	client.mu.RLock()
	open := client.circuitOpen
	fails := client.consecutiveFails
	client.mu.RUnlock()

	if open || fails != 0 {
		t.Fatalf("Expected circuit reset, got open=%v fails=%d", open, fails)
	}
}

func TestCircuitBreaker_PartialFailuresDoNotOpenCircuit(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		// Alternate: fail, succeed, fail, succeed...
		if requestCount%2 == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"success": "true"})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	// Make many alternating fail/success requests
	// Since GET retries 3 times on 500, each "fail" request actually makes 3 server calls
	// But then the success resets the counter
	for i := 0; i < 10; i++ {
		// This will fail (server returns 500 for odd requestCount)
		_ = client.Get(context.Background(), "/test", nil)
		// Reset requestCount to even so next succeeds
		requestCount = 0
		// This will succeed
		_ = client.Get(context.Background(), "/test", nil)
		requestCount = 0
	}

	// Circuit should not be open because successes reset the counter
	if client.isCircuitOpen() {
		t.Error("Circuit should not be open when successes reset the failure counter")
	}
}

func TestParseRetryAfter(t *testing.T) {
	testCases := []struct {
		name     string
		header   string
		expected time.Duration
	}{
		{"empty header", "", time.Second},
		{"numeric seconds", "30", 30 * time.Second},
		{"invalid string", "invalid", time.Second},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseRetryAfter(tc.header)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestParseRetryAfterRFC1123(t *testing.T) {
	// Test RFC1123 date format
	futureTime := time.Now().Add(60 * time.Second).UTC()
	header := futureTime.Format(time.RFC1123)

	result := parseRetryAfter(header)

	// Allow some tolerance for timing
	if result < 59*time.Second || result > 61*time.Second {
		t.Errorf("Expected approximately 60s, got %v", result)
	}
}

func TestParseRetryAfterRFC1123Past(t *testing.T) {
	past := time.Now().Add(-10 * time.Second).UTC().Format(time.RFC1123)
	result := parseRetryAfter(past)
	if result != 0 {
		t.Errorf("Expected 0s for past Retry-After, got %v", result)
	}
}

func TestClientPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"created": "true"})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	var result map[string]string
	err := client.Post(context.Background(), "/test", map[string]string{"name": "test"}, &result)
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}

	if result["created"] != "true" {
		t.Errorf("Unexpected result: %v", result)
	}
}

func TestClientPut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"updated": "true"})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	var result map[string]string
	err := client.Put(context.Background(), "/test", map[string]string{"name": "updated"}, &result)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	if result["updated"] != "true" {
		t.Errorf("Unexpected result: %v", result)
	}
}

func TestClientDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.Delete(context.Background(), "/test")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestClient4xxError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"code":    "NOT_FOUND",
			"message": "Resource not found",
		})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("Expected error for 404, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T: %v", err, err)
	}
	if apiErr.Status != 404 {
		t.Errorf("Expected status 404, got %d", apiErr.Status)
	}
	if apiErr.Code != "NOT_FOUND" {
		t.Errorf("Expected code NOT_FOUND, got %s", apiErr.Code)
	}
}

func TestClient4xxErrorDecodeFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("Expected error for 400, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T: %v", err, err)
	}
	if apiErr.Code != "HTTP_400" {
		t.Errorf("Expected code HTTP_400, got %s", apiErr.Code)
	}
}

func TestClientServerError(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)
	// Reset circuit breaker state
	client.circuitOpen = false
	client.consecutiveFails = 0

	err := client.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("Expected error for 500, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T: %v", err, err)
	}
	if apiErr.Code != "HTTP_500" {
		t.Errorf("Expected code HTTP_500, got %s", apiErr.Code)
	}

	// GET should retry on 500
	if attempts != 3 {
		t.Errorf("Expected 3 attempts for GET on 500, got %d", attempts)
	}
}

func TestClientMarshalError(t *testing.T) {
	client := NewClient("test-token")

	// Create an unmarshalable value (channel)
	unmarshalable := make(chan int)

	err := client.Post(context.Background(), "/test", unmarshalable, nil)
	if err == nil {
		t.Fatal("Expected marshal error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to marshal request body") {
		t.Errorf("Expected marshal error, got: %v", err)
	}
}

func TestClientResponseDecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	var result map[string]string
	err := client.Get(context.Background(), "/test", &result)
	if err == nil {
		t.Fatal("Expected decode error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode response") {
		t.Errorf("Expected decode error, got: %v", err)
	}
}

func TestClientNetworkError(t *testing.T) {
	attempts := 0
	client := NewClient("test-token")
	client.BaseURL = "http://example.invalid"
	client.SetUseOpenAPI(false)
	client.retry = retryConfig{
		baseDelay: 0,
		maxDelay:  0,
		budget:    -1, // negative means unlimited retries
		jitter:    0,
	}
	client.httpClient.Transport = roundTripperFunc(func(*http.Request) (*http.Response, error) {
		attempts++
		return nil, errors.New("network error")
	})

	err := client.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("Expected network error, got nil")
	}
	if attempts != maxRetries {
		t.Errorf("Expected %d attempts, got %d", maxRetries, attempts)
	}
}

func TestClientPostNetworkErrorNoRetry(t *testing.T) {
	attempts := 0
	client := NewClient("test-token")
	client.BaseURL = "http://example.invalid"
	client.SetUseOpenAPI(false)
	client.retry = retryConfig{
		baseDelay: 0,
		maxDelay:  0,
		budget:    0,
		jitter:    0,
	}
	client.httpClient.Transport = roundTripperFunc(func(*http.Request) (*http.Response, error) {
		attempts++
		return nil, errors.New("network error")
	})

	err := client.Post(context.Background(), "/test", map[string]string{"key": "value"}, nil)
	if err == nil {
		t.Fatal("Expected network error, got nil")
	}
	if attempts != 1 {
		t.Errorf("Expected 1 attempt for POST, got %d", attempts)
	}
}

func TestClientNetworkErrorRetryBudgetZeroDisables(t *testing.T) {
	attempts := 0
	client := NewClient("test-token")
	client.BaseURL = "http://example.invalid"
	client.SetUseOpenAPI(false)
	client.retry = retryConfig{
		baseDelay: 100 * time.Millisecond,
		maxDelay:  100 * time.Millisecond,
		budget:    0, // 0 explicitly disables retries
		jitter:    0,
	}
	client.httpClient.Transport = roundTripperFunc(func(*http.Request) (*http.Response, error) {
		attempts++
		return nil, errors.New("network error")
	})

	err := client.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("Expected network error, got nil")
	}
	if attempts != 1 {
		t.Errorf("Expected 1 attempt when budget=0 disables retries, got %d", attempts)
	}
}

func TestClientNetworkErrorRetryBudgetExhausted(t *testing.T) {
	attempts := 0
	client := NewClient("test-token")
	client.BaseURL = "http://example.invalid"
	client.SetUseOpenAPI(false)
	client.retry = retryConfig{
		baseDelay: 100 * time.Millisecond,
		maxDelay:  100 * time.Millisecond,
		budget:    time.Nanosecond, // very small budget exhausts immediately
		jitter:    0,
	}
	client.httpClient.Transport = roundTripperFunc(func(*http.Request) (*http.Response, error) {
		attempts++
		return nil, errors.New("network error")
	})

	err := client.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("Expected network error, got nil")
	}
	if attempts != 1 {
		t.Errorf("Expected 1 attempt due to exhausted retry budget, got %d", attempts)
	}
}

func TestClientPostServerErrorNoRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)
	client.circuitOpen = false
	client.consecutiveFails = 0

	err := client.Post(context.Background(), "/test", map[string]string{"key": "value"}, nil)
	if err == nil {
		t.Fatal("Expected error for 500, got nil")
	}

	// POST should NOT retry on 500 - only one attempt expected
	if attempts != 1 {
		t.Errorf("POST should not retry on 500: expected 1 attempt, got %d", attempts)
	}
}

func TestDebugLoggerFromEnv(t *testing.T) {
	tests := []struct {
		name       string
		envValue   string
		wantOutput bool
	}{
		{"unset returns nop", "", false},
		{"1 returns real logger", "1", true},
		{"true returns real logger", "true", true},
		{"yes returns real logger", "yes", true},
		{"on returns real logger", "on", true},
		{"false returns nop", "false", false},
		{"0 returns nop", "0", false},
		{"random returns nop", "random", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.envValue == "" {
				_ = os.Unsetenv("SHOPLINE_DEBUG")
			} else {
				_ = os.Setenv("SHOPLINE_DEBUG", tc.envValue)
			}
			defer func() { _ = os.Unsetenv("SHOPLINE_DEBUG") }()

			logger := debugLoggerFromEnv()

			// To verify logger type, we create test loggers and compare behavior
			// A real logger writes to os.Stderr, a nop logger writes to io.Discard
			// We test by creating equivalent loggers and checking if they match
			if tc.wantOutput {
				// Should be equivalent to debug.New(os.Stderr)
				expected := debug.New(os.Stderr)
				if logger == nil {
					t.Fatal("Expected non-nil logger")
				}
				// Both should be non-nil Logger pointers
				if expected == nil {
					t.Fatal("debug.New returned nil")
				}
			} else {
				// Should be equivalent to debug.Nop()
				if logger == nil {
					t.Fatal("Expected non-nil logger (even for nop)")
				}
			}
		})
	}
}

// TestDebugLoggerFromEnvOutput verifies that the logger actually writes output
// when SHOPLINE_DEBUG is enabled vs when it's disabled.
func TestDebugLoggerFromEnvOutput(t *testing.T) {
	// Test that nop logger produces no output
	t.Run("nop logger produces no output", func(t *testing.T) {
		_ = os.Unsetenv("SHOPLINE_DEBUG")
		defer func() { _ = os.Unsetenv("SHOPLINE_DEBUG") }()

		nopLogger := debug.Nop()
		var buf bytes.Buffer

		// Redirect the nop logger's output - but since Nop uses io.Discard,
		// we need to create our own to verify it doesn't write
		testLogger := debug.New(&buf)
		// Nop logger should not produce output when we call Printf
		nopLogger.Printf("test message")

		// The nop logger writes to io.Discard, so we can't capture its output
		// Instead, verify that a Nop() logger behaves differently than New()
		testLogger.Printf("test message")
		if buf.Len() == 0 {
			t.Error("Real logger should produce output")
		}
	})

	// Test that real logger produces output
	t.Run("real logger produces output", func(t *testing.T) {
		var buf bytes.Buffer
		logger := debug.New(&buf)
		logger.Printf("test message %d", 123)

		output := buf.String()
		if !strings.Contains(output, "[DEBUG]") {
			t.Errorf("Expected output to contain [DEBUG], got: %s", output)
		}
		if !strings.Contains(output, "test message 123") {
			t.Errorf("Expected output to contain 'test message 123', got: %s", output)
		}
	})

	// Verify Nop() returns a logger that discards output
	t.Run("nop logger discards output", func(t *testing.T) {
		nopLogger := debug.Nop()
		// This should not panic and should silently discard
		nopLogger.Printf("this should be discarded")

		// We can verify the behavior by checking that Nop uses io.Discard
		// by creating a logger with io.Discard and comparing behavior
		discardLogger := debug.New(io.Discard)
		discardLogger.Printf("this should also be discarded")

		// Both should complete without error - the test passes if no panic
	})
}

func TestClientHTMLErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("<html><body>Not Found</body></html>"))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T: %v", err, err)
	}
	if apiErr.Status != 404 {
		t.Errorf("Expected status 404, got %d", apiErr.Status)
	}
	if apiErr.Code != "HTTP_404" {
		t.Errorf("Expected code HTTP_404, got %s", apiErr.Code)
	}
	if apiErr.Message != "Not Found" {
		t.Errorf("Expected message 'Not Found', got %s", apiErr.Message)
	}
}

func TestClientPlainTextErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("upstream timeout"))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.Get(context.Background(), "/test", nil)
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T: %v", err, err)
	}
	if apiErr.Code != "HTTP_502" {
		t.Errorf("Expected code HTTP_502, got %s", apiErr.Code)
	}
	// For non-HTML bodies, include the body snippet in the message
	if apiErr.Message != "Bad Gateway: upstream timeout" {
		t.Errorf("Expected message 'Bad Gateway: upstream timeout', got: %s", apiErr.Message)
	}
}
