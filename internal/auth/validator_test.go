package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPICredentialValidator_Validate(t *testing.T) {
	tests := []struct {
		name        string
		handle      string
		accessToken string
		serverFunc  func(w http.ResponseWriter, r *http.Request)
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty handle",
			handle:      "",
			accessToken: "valid-token",
			wantErr:     true,
			errContains: "handle is required",
		},
		{
			name:        "empty access token",
			handle:      "test-store",
			accessToken: "",
			wantErr:     true,
			errContains: "access token is required",
		},
		{
			name:        "invalid handle with dots",
			handle:      "evil.com",
			accessToken: "valid-token",
			wantErr:     true,
			errContains: "handle contains invalid characters",
		},
		{
			name:        "invalid handle with slashes",
			handle:      "evil/path",
			accessToken: "valid-token",
			wantErr:     true,
			errContains: "handle contains invalid characters",
		},
		{
			name:        "invalid handle injection attempt",
			handle:      "evil.com/fake?x=",
			accessToken: "valid-token",
			wantErr:     true,
			errContains: "handle contains invalid characters",
		},
		{
			name:        "valid credentials",
			handle:      "test-store",
			accessToken: "valid-token",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				// Verify headers are set correctly
				if r.Header.Get("Authorization") != "Bearer valid-token" {
					t.Error("Missing or incorrect Authorization header")
				}
				if r.Header.Get("X-Shopline-Access-Token") != "valid-token" {
					t.Error("Missing or incorrect X-Shopline-Access-Token header")
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"orders": []interface{}{}})
			},
			wantErr: false,
		},
		{
			name:        "unauthorized - 401",
			handle:      "test-store",
			accessToken: "invalid-token",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"code":    "UNAUTHORIZED",
					"message": "Invalid access token",
				})
			},
			wantErr:     true,
			errContains: "invalid credentials: Invalid access token",
		},
		{
			name:        "forbidden - 403",
			handle:      "test-store",
			accessToken: "forbidden-token",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"code":    "FORBIDDEN",
					"message": "Access denied",
				})
			},
			wantErr:     true,
			errContains: "invalid credentials: Access denied",
		},
		{
			name:        "unauthorized with no message",
			handle:      "test-store",
			accessToken: "bad-token",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte("not json"))
			},
			wantErr:     true,
			errContains: "invalid credentials",
		},
		{
			name:        "server error - 500",
			handle:      "test-store",
			accessToken: "valid-token",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr:     true,
			errContains: "could not validate credentials: server error",
		},
		{
			name:        "other client error - 400",
			handle:      "test-store",
			accessToken: "valid-token",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"code":    "BAD_REQUEST",
					"message": "Invalid request format",
				})
			},
			wantErr:     true,
			errContains: "Invalid request format",
		},
		{
			name:        "unauthorized with empty message in JSON",
			handle:      "test-store",
			accessToken: "bad-token",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"code":    "UNAUTHORIZED",
					"message": "", // Empty message
				})
			},
			wantErr:     true,
			errContains: "invalid credentials",
		},
		{
			name:        "client error 400 with non-JSON response",
			handle:      "test-store",
			accessToken: "valid-token",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("not json"))
			},
			wantErr:     true,
			errContains: "validation failed (status 400)",
		},
		{
			name:        "client error 400 with empty message in JSON",
			handle:      "test-store",
			accessToken: "valid-token",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"code":    "BAD_REQUEST",
					"message": "", // Empty message
				})
			},
			wantErr:     true,
			errContains: "validation failed (status 400)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			var validator *APICredentialValidator

			if tt.serverFunc != nil {
				server = httptest.NewServer(http.HandlerFunc(tt.serverFunc))
				defer server.Close()

				// Create validator with custom URL builder that points to test server
				validator = NewAPICredentialValidatorWithURLBuilder(func(handle string) string {
					return server.URL + "/orders?page=1&page_size=1"
				})
			} else {
				// For tests that don't need a server (validation errors),
				// use default validator
				validator = NewAPICredentialValidator()
			}

			err := validator.Validate(context.Background(), tt.handle, tt.accessToken)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCredentialValidationError(t *testing.T) {
	err := &CredentialValidationError{Message: "test error"}
	if err.Error() != "test error" {
		t.Errorf("expected 'test error', got %q", err.Error())
	}
}

func TestValidate_NetworkError(t *testing.T) {
	// Use a URL builder that points to a non-existent server
	validator := NewAPICredentialValidatorWithURLBuilder(func(handle string) string {
		return "http://localhost:59999/nonexistent" // Port that's unlikely to be in use
	})

	err := validator.Validate(context.Background(), "test-store", "valid-token")
	if err == nil {
		t.Error("expected error for network failure, got nil")
	}
	if !strings.Contains(err.Error(), "could not validate credentials") {
		t.Errorf("expected 'could not validate credentials' error, got %q", err.Error())
	}
}

func TestValidate_InvalidURL(t *testing.T) {
	// Use a URL builder that returns an invalid URL (containing control characters)
	validator := NewAPICredentialValidatorWithURLBuilder(func(handle string) string {
		return "http://invalid\x00url.com/" // Null character makes URL invalid
	})

	err := validator.Validate(context.Background(), "test-store", "valid-token")
	if err == nil {
		t.Error("expected error for invalid URL, got nil")
	}
	if !strings.Contains(err.Error(), "could not validate credentials") {
		t.Errorf("expected 'could not validate credentials' error, got %q", err.Error())
	}
}

func TestIsValidHandle(t *testing.T) {
	tests := []struct {
		handle string
		valid  bool
	}{
		{"test-store", true},
		{"TestStore123", true},
		{"my-store-name", true},
		{"store", true},
		{"123", true},
		{"", false},
		{"evil.com", false},
		{"path/to/evil", false},
		{"store?query=1", false},
		{"store&param=2", false},
		{"evil.com/fake?x=", false},
		{"store name", false},
		{"store@domain", false},
	}

	for _, tt := range tests {
		t.Run(tt.handle, func(t *testing.T) {
			got := isValidHandle(tt.handle)
			if got != tt.valid {
				t.Errorf("isValidHandle(%q) = %v, want %v", tt.handle, got, tt.valid)
			}
		})
	}
}
