package api

import (
	"errors"
	"testing"
	"time"
)

func TestAPIErrorError(t *testing.T) {
	err := &APIError{
		Code:    "NOT_FOUND",
		Message: "Order not found",
		Status:  404,
	}

	got := err.Error()
	if got != "NOT_FOUND: Order not found (status 404)" {
		t.Errorf("Unexpected error message: %s", got)
	}
}

func TestIsRateLimitError(t *testing.T) {
	err := &RateLimitError{RetryAfter: 60}

	var rle *RateLimitError
	if !errors.As(err, &rle) {
		t.Error("Expected error to be RateLimitError")
	}
}

func TestRateLimitErrorError(t *testing.T) {
	err := &RateLimitError{RetryAfter: 30 * time.Second}

	got := err.Error()
	expected := "rate limit exceeded, retry after 30s"
	if got != expected {
		t.Errorf("Expected %q, got %q", expected, got)
	}
}

func TestAuthErrorError(t *testing.T) {
	err := &AuthError{Reason: "invalid token"}

	got := err.Error()
	expected := "authentication failed: invalid token"
	if got != expected {
		t.Errorf("Expected %q, got %q", expected, got)
	}

	// Test errors.As compatibility
	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Error("Expected error to be AuthError")
	}
}

func TestValidationErrorError(t *testing.T) {
	err := &ValidationError{
		Field:   "email",
		Message: "must be a valid email address",
	}

	got := err.Error()
	expected := "validation error on field email: must be a valid email address"
	if got != expected {
		t.Errorf("Expected %q, got %q", expected, got)
	}

	// Test errors.As compatibility
	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Error("Expected error to be ValidationError")
	}
}

func TestCircuitBreakerErrorError(t *testing.T) {
	err := &CircuitBreakerError{}

	got := err.Error()
	expected := "circuit breaker is open, requests are temporarily blocked"
	if got != expected {
		t.Errorf("Expected %q, got %q", expected, got)
	}

	// Test errors.As compatibility
	var cbErr *CircuitBreakerError
	if !errors.As(err, &cbErr) {
		t.Error("Expected error to be CircuitBreakerError")
	}
}
