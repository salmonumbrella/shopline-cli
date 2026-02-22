package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/secrets"
)

func TestMain(m *testing.M) {
	// Prevent real browser opens during tests.
	openBrowserFunc = func(string) error { return nil }
	os.Exit(m.Run())
}

// mockValidator is a test double for CredentialValidator
type mockValidator struct {
	validateFunc func(ctx context.Context, handle, accessToken string) error
}

func (m *mockValidator) Validate(ctx context.Context, handle, accessToken string) error {
	if m.validateFunc != nil {
		return m.validateFunc(ctx, handle, accessToken)
	}
	return nil
}

// Helper to create JSON request body
func jsonBody(t *testing.T, data map[string]string) *bytes.Buffer {
	t.Helper()
	body, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}
	return bytes.NewBuffer(body)
}

func TestHandleSubmit_ValidCredentials(t *testing.T) {
	validator := &mockValidator{
		validateFunc: func(ctx context.Context, handle, accessToken string) error {
			if handle != "my-store" {
				t.Errorf("expected handle 'my-store', got %q", handle)
			}
			if accessToken != "valid-token" {
				t.Errorf("expected accessToken 'valid-token', got %q", accessToken)
			}
			return nil
		},
	}

	server, err := NewServerWithValidator(validator)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	body := jsonBody(t, map[string]string{
		"handle":       "my-store",
		"access_token": "valid-token",
		"app_key":      "test-app-key",
		"app_secret":   "test-app-secret",
	})

	req := httptest.NewRequest(http.MethodPost, "/submit", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", server.csrfToken)

	w := httptest.NewRecorder()
	server.handleSubmit(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["success"] != true {
		t.Errorf("expected success=true, got %v", response["success"])
	}

	// Verify credentials were stored in pending result
	if server.pendingResult == nil {
		t.Fatal("expected pending result to be set")
	}
	if server.pendingResult.Credentials == nil {
		t.Fatal("expected credentials, got nil")
	}
	if server.pendingResult.Credentials.Handle != "my-store" {
		t.Errorf("expected handle 'my-store', got %q", server.pendingResult.Credentials.Handle)
	}
	if server.pendingResult.Credentials.AccessToken != "valid-token" {
		t.Errorf("expected access_token 'valid-token', got %q", server.pendingResult.Credentials.AccessToken)
	}
}

func TestHandleSubmit_InvalidCredentials(t *testing.T) {
	validator := &mockValidator{
		validateFunc: func(ctx context.Context, handle, accessToken string) error {
			return &CredentialValidationError{Message: "invalid credentials: Invalid access token"}
		},
	}

	server, err := NewServerWithValidator(validator)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	body := jsonBody(t, map[string]string{
		"handle":       "my-store",
		"access_token": "bad-token",
		"app_key":      "test-app-key",
		"app_secret":   "test-app-secret",
	})

	req := httptest.NewRequest(http.MethodPost, "/submit", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", server.csrfToken)

	w := httptest.NewRecorder()
	server.handleSubmit(w, req)

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["success"] != false {
		t.Errorf("expected success=false, got %v", response["success"])
	}

	errMsg, ok := response["error"].(string)
	if !ok {
		t.Fatalf("expected error to be string, got %T", response["error"])
	}
	if errMsg != "invalid credentials: Invalid access token" {
		t.Errorf("unexpected error message: %q", errMsg)
	}

	// Verify no pending result was set
	if server.pendingResult != nil {
		t.Errorf("expected no pending result, got %+v", server.pendingResult)
	}
}

func TestHandleSubmit_NetworkError(t *testing.T) {
	validator := &mockValidator{
		validateFunc: func(ctx context.Context, handle, accessToken string) error {
			// Simulate a network error (not a CredentialValidationError)
			return errors.New("could not validate credentials: connection refused")
		},
	}

	server, err := NewServerWithValidator(validator)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	body := jsonBody(t, map[string]string{
		"handle":       "my-store",
		"access_token": "valid-token",
		"app_key":      "test-app-key",
		"app_secret":   "test-app-secret",
	})

	req := httptest.NewRequest(http.MethodPost, "/submit", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", server.csrfToken)

	w := httptest.NewRecorder()
	server.handleSubmit(w, req)

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["success"] != false {
		t.Errorf("expected success=false, got %v", response["success"])
	}

	errMsg, ok := response["error"].(string)
	if !ok {
		t.Fatalf("expected error to be string, got %T", response["error"])
	}
	// Network errors should return a generic message (not leak internal details)
	expectedMsg := "Could not validate credentials. Please check your network connection and try again."
	if errMsg != expectedMsg {
		t.Errorf("expected generic error message %q, got %q", expectedMsg, errMsg)
	}
}

func TestHandleSubmit_EmptyHandle(t *testing.T) {
	server, err := NewServerWithValidator(&mockValidator{})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	body := jsonBody(t, map[string]string{
		"handle":       "",
		"access_token": "valid-token",
		"app_key":      "test-app-key",
		"app_secret":   "test-app-secret",
	})

	req := httptest.NewRequest(http.MethodPost, "/submit", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", server.csrfToken)

	w := httptest.NewRecorder()
	server.handleSubmit(w, req)

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["success"] != false {
		t.Errorf("expected success=false, got %v", response["success"])
	}

	errMsg, ok := response["error"].(string)
	if !ok {
		t.Fatalf("expected error to be string, got %T", response["error"])
	}
	if errMsg != "Store handle and access token are required" {
		t.Errorf("unexpected error message: %q", errMsg)
	}
}

func TestHandleSubmit_InvalidCSRFToken(t *testing.T) {
	validator := &mockValidator{
		validateFunc: func(ctx context.Context, handle, accessToken string) error {
			t.Error("validator should not be called when CSRF token is invalid")
			return nil
		},
	}

	server, err := NewServerWithValidator(validator)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	body := jsonBody(t, map[string]string{
		"handle":       "my-store",
		"access_token": "valid-token",
	})

	req := httptest.NewRequest(http.MethodPost, "/submit", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", "wrong-token")

	w := httptest.NewRecorder()
	server.handleSubmit(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", resp.StatusCode)
	}
}

func TestNewServer_CreatesValidatorByDefault(t *testing.T) {
	server, err := NewServer()
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	if server.validator == nil {
		t.Error("expected validator to be set")
	}

	// Verify it's the production validator type
	if _, ok := server.validator.(*APICredentialValidator); !ok {
		t.Errorf("expected *APICredentialValidator, got %T", server.validator)
	}
}

func TestNewServerWithValidator_UsesProvidedValidator(t *testing.T) {
	customValidator := &mockValidator{}

	server, err := NewServerWithValidator(customValidator)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	if server.validator != customValidator {
		t.Error("expected custom validator to be used")
	}
}

func TestHandleSetup_RendersExpectedFields(t *testing.T) {
	server, err := NewServerWithValidator(&mockValidator{})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	server.handleSetup(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := w.Body.String()

	// Verify CSRF token is present
	if !strings.Contains(body, server.csrfToken) {
		t.Error("form should contain CSRF token")
	}

	// Verify expected form fields are present
	expectedFields := []string{
		`id="handle"`,
		`id="accessToken"`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(body, field) {
			t.Errorf("form should contain field %q", field)
		}
	}

	// Verify page title
	if !strings.Contains(body, "Shopline CLI Setup") {
		t.Error("page should contain title 'Shopline CLI Setup'")
	}
}

func TestRateLimiting_BlocksAfterMaxAttempts(t *testing.T) {
	server, err := NewServerWithValidator(&mockValidator{})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Simulate maxAttempts requests from the same IP
	clientIP := "192.168.1.1:12345"

	// First maxAttempts should succeed
	for i := 0; i < maxAttempts; i++ {
		if !server.checkRateLimit(clientIP) {
			t.Errorf("attempt %d should not be rate limited", i+1)
		}
	}

	// Next attempt should be blocked
	if server.checkRateLimit(clientIP) {
		t.Error("attempt after max should be rate limited")
	}
}

func TestRateLimiting_DifferentIPsNotAffected(t *testing.T) {
	server, err := NewServerWithValidator(&mockValidator{})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Max out attempts for one IP
	clientIP1 := "192.168.1.1:12345"
	for i := 0; i < maxAttempts; i++ {
		server.checkRateLimit(clientIP1)
	}

	// Different IP should not be affected
	clientIP2 := "192.168.1.2:12345"
	if !server.checkRateLimit(clientIP2) {
		t.Error("different IP should not be rate limited")
	}
}

func TestHandleSubmit_RateLimitedReturnsError(t *testing.T) {
	server, err := NewServerWithValidator(&mockValidator{})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Pre-fill rate limit for this IP
	clientIP := "192.0.2.1:1234" // httptest uses this format
	for i := 0; i < maxAttempts; i++ {
		server.checkRateLimit(clientIP)
	}

	body := jsonBody(t, map[string]string{
		"handle":       "my-store",
		"access_token": "valid-token",
	})

	req := httptest.NewRequest(http.MethodPost, "/submit", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", server.csrfToken)
	req.RemoteAddr = clientIP

	w := httptest.NewRecorder()
	server.handleSubmit(w, req)

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["success"] != false {
		t.Errorf("expected success=false, got %v", response["success"])
	}

	errMsg, ok := response["error"].(string)
	if !ok {
		t.Fatalf("expected error to be string, got %T", response["error"])
	}
	if !strings.Contains(errMsg, "Too many attempts") {
		t.Errorf("expected rate limit error message, got %q", errMsg)
	}
}

func TestServerRun_StartsAndListensOnLocalhost(t *testing.T) {
	server, err := NewServerWithValidator(&mockValidator{})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Run server in background
	errCh := make(chan error, 1)
	go func() {
		_, err := server.Run(ctx)
		errCh <- err
	}()

	// Give server time to start
	time.Sleep(50 * time.Millisecond)

	// Cancel context to stop server
	cancel()

	// Wait for server to shutdown
	select {
	case err := <-errCh:
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("server did not shutdown in time")
	}
}

func TestHandleSubmit_InvalidJSON(t *testing.T) {
	server, err := NewServerWithValidator(&mockValidator{})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Send invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader("not valid json{"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", server.csrfToken)

	w := httptest.NewRecorder()
	server.handleSubmit(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["success"] != false {
		t.Errorf("expected success=false, got %v", response["success"])
	}
	if response["error"] != "Invalid request body" {
		t.Errorf("expected 'Invalid request body', got %q", response["error"])
	}
}

func TestHandleSubmit_MissingCSRFToken(t *testing.T) {
	server, err := NewServerWithValidator(&mockValidator{})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Request without CSRF token header
	body := jsonBody(t, map[string]string{
		"handle":       "my-store",
		"access_token": "valid-token",
	})

	req := httptest.NewRequest(http.MethodPost, "/submit", body)
	req.Header.Set("Content-Type", "application/json")
	// Note: No X-CSRF-Token header

	w := httptest.NewRecorder()
	server.handleSubmit(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", resp.StatusCode)
	}
}

func TestNewServer_GeneratesUniqueCSRFTokens(t *testing.T) {
	server1, err := NewServerWithValidator(&mockValidator{})
	if err != nil {
		t.Fatalf("failed to create server1: %v", err)
	}

	server2, err := NewServerWithValidator(&mockValidator{})
	if err != nil {
		t.Fatalf("failed to create server2: %v", err)
	}

	if server1.csrfToken == server2.csrfToken {
		t.Error("different servers should have different CSRF tokens")
	}

	// Token should be a valid hex string of expected length (32 bytes = 64 hex chars)
	if len(server1.csrfToken) != 64 {
		t.Errorf("CSRF token should be 64 hex chars, got %d", len(server1.csrfToken))
	}
}

func TestServerRun_ReturnsCredentialsOnSuccess(t *testing.T) {
	server, err := NewServerWithValidator(&mockValidator{})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	ctx := context.Background()

	// Run server in background
	resultCh := make(chan struct {
		creds *secrets.StoreCredentials
		err   error
	}, 1)
	go func() {
		creds, err := server.Run(ctx)
		resultCh <- struct {
			creds *secrets.StoreCredentials
			err   error
		}{creds, err}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Submit valid credentials via the result channel (simulating successful form submit)
	expectedCreds := &secrets.StoreCredentials{
		Name:        "test-profile",
		Handle:      "my-store",
		AccessToken: "valid-token",
	}
	server.result <- Result{Credentials: expectedCreds}

	// Wait for server to return
	select {
	case result := <-resultCh:
		if result.err != nil {
			t.Errorf("unexpected error: %v", result.err)
		}
		if result.creds == nil {
			t.Fatal("expected credentials, got nil")
		}
		if result.creds.Handle != "my-store" {
			t.Errorf("expected handle 'my-store', got %q", result.creds.Handle)
		}
	case <-time.After(2 * time.Second):
		t.Error("server did not return result in time")
	}
}

func TestServerRun_ReturnsErrorFromResult(t *testing.T) {
	server, err := NewServerWithValidator(&mockValidator{})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	ctx := context.Background()

	// Run server in background
	resultCh := make(chan struct {
		creds *secrets.StoreCredentials
		err   error
	}, 1)
	go func() {
		creds, err := server.Run(ctx)
		resultCh <- struct {
			creds *secrets.StoreCredentials
			err   error
		}{creds, err}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Submit an error via the result channel
	expectedErr := errors.New("validation failed")
	server.result <- Result{Error: expectedErr}

	// Wait for server to return
	select {
	case result := <-resultCh:
		if result.err == nil {
			t.Error("expected error, got nil")
		} else if result.err.Error() != "validation failed" {
			t.Errorf("expected 'validation failed' error, got %q", result.err.Error())
		}
		if result.creds != nil {
			t.Errorf("expected nil credentials, got %+v", result.creds)
		}
	case <-time.After(2 * time.Second):
		t.Error("server did not return result in time")
	}
}

func TestOpenBrowser_UsesOverridableFunc(t *testing.T) {
	origFunc := openBrowserFunc
	defer func() { openBrowserFunc = origFunc }()

	var called bool
	openBrowserFunc = func(url string) error {
		called = true
		if url != "http://localhost:12345/test" {
			t.Errorf("url = %q, want %q", url, "http://localhost:12345/test")
		}
		return nil
	}

	err := openBrowser("http://localhost:12345/test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !called {
		t.Error("openBrowserFunc was not called")
	}
}

func TestRateLimiting_ExpiredAttemptsAreCleared(t *testing.T) {
	server, err := NewServerWithValidator(&mockValidator{})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	clientIP := "192.168.1.1:12345"

	// Manually inject old attempts that should be expired
	server.mu.Lock()
	oldTime := time.Now().Add(-20 * time.Minute) // older than rateLimitWindow
	for i := 0; i < maxAttempts; i++ {
		server.attempts[clientIP] = append(server.attempts[clientIP], oldTime)
	}
	server.mu.Unlock()

	// New attempt should succeed because old attempts are expired
	if !server.checkRateLimit(clientIP) {
		t.Error("old expired attempts should not count toward rate limit")
	}
}

func TestDefaultURLBuilder(t *testing.T) {
	// Open API uses a fixed URL since tokens are store-scoped
	url := DefaultURLBuilder("my-store")
	expected := "https://open.shopline.io/v1/orders?per_page=1"
	if url != expected {
		t.Errorf("DefaultURLBuilder returned %q, expected %q", url, expected)
	}
}

func TestDefaultURLBuilder_DifferentHandles(t *testing.T) {
	// Open API uses the same URL regardless of handle since tokens are store-scoped
	expected := "https://open.shopline.io/v1/orders?per_page=1"
	tests := []struct {
		handle string
	}{
		{"test-store"},
		{"store123"},
		{"my-super-store"},
	}

	for _, tt := range tests {
		t.Run(tt.handle, func(t *testing.T) {
			url := DefaultURLBuilder(tt.handle)
			if url != expected {
				t.Errorf("DefaultURLBuilder(%q) = %q, want %q", tt.handle, url, expected)
			}
		})
	}
}

func TestHandleValidate_ValidCredentials(t *testing.T) {
	validator := &mockValidator{
		validateFunc: func(ctx context.Context, handle, accessToken string) error {
			return nil
		},
	}

	server, err := NewServerWithValidator(validator)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	body := jsonBody(t, map[string]string{
		"handle":       "my-store",
		"access_token": "valid-token",
	})

	req := httptest.NewRequest(http.MethodPost, "/validate", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", server.csrfToken)

	w := httptest.NewRecorder()
	server.handleValidate(w, req)

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["success"] != true {
		t.Errorf("expected success=true, got %v", response["success"])
	}
	if response["store_name"] != "my-store" {
		t.Errorf("expected store_name 'my-store', got %v", response["store_name"])
	}
	// order_count should be present (will be 0 since we're using a mock)
	if _, ok := response["order_count"]; !ok {
		t.Error("expected order_count field in response")
	}
}

func TestHandleValidate_InvalidCredentials(t *testing.T) {
	validator := &mockValidator{
		validateFunc: func(ctx context.Context, handle, accessToken string) error {
			return &CredentialValidationError{Message: "invalid token"}
		},
	}

	server, err := NewServerWithValidator(validator)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	body := jsonBody(t, map[string]string{
		"handle":       "my-store",
		"access_token": "bad-token",
	})

	req := httptest.NewRequest(http.MethodPost, "/validate", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", server.csrfToken)

	w := httptest.NewRecorder()
	server.handleValidate(w, req)

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["success"] != false {
		t.Errorf("expected success=false, got %v", response["success"])
	}
	if response["error"] != "invalid token" {
		t.Errorf("expected error 'invalid token', got %v", response["error"])
	}
}

func TestHandleSuccess_RendersPage(t *testing.T) {
	server, err := NewServerWithValidator(&mockValidator{})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/success?store=mystore.myshopline.com", nil)
	w := httptest.NewRecorder()

	server.handleSuccess(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, "mystore.myshopline.com") {
		t.Error("success page should contain store name")
	}
	if !strings.Contains(body, "You're all set!") {
		t.Error("success page should contain success message")
	}
}

func TestHandleComplete_SignalsShutdown(t *testing.T) {
	server, err := NewServerWithValidator(&mockValidator{})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Set pending result
	server.pendingResult = &Result{
		Credentials: &secrets.StoreCredentials{
			Handle: "test-store",
		},
	}

	// Start goroutine to receive result
	resultCh := make(chan Result, 1)
	go func() {
		result := <-server.result
		resultCh <- result
	}()

	req := httptest.NewRequest(http.MethodPost, "/complete", nil)
	w := httptest.NewRecorder()

	server.handleComplete(w, req)

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["success"] != true {
		t.Errorf("expected success=true, got %v", response["success"])
	}

	// Verify result was sent
	select {
	case result := <-resultCh:
		if result.Credentials == nil {
			t.Error("expected credentials in result")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected result to be sent")
	}
}

func TestHandleSetup_ReturnsNotFoundForOtherPaths(t *testing.T) {
	server, err := NewServerWithValidator(&mockValidator{})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/other-path", nil)
	w := httptest.NewRecorder()

	server.handleSetup(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", resp.StatusCode)
	}
}
