package api

import (
	"errors"
	"testing"
	"time"
)

func TestRichError(t *testing.T) {
	err := &RichError{
		Message: "Order not found",
		Code:    "NOT_FOUND",
		Suggestions: []string{
			"Check the order ID is correct",
			"Run 'spl orders list' to see available orders",
		},
	}

	// Test error interface
	if err.Error() != "Order not found" {
		t.Errorf("unexpected error message: %s", err.Error())
	}

	// Test suggestions
	if len(err.Suggestions) != 2 {
		t.Errorf("expected 2 suggestions, got %d", len(err.Suggestions))
	}
}

func TestEnrichError_NotFound(t *testing.T) {
	baseErr := &APIError{
		Status:  404,
		Message: "Resource not found",
	}

	rich := EnrichError(baseErr, "orders", "12345")

	var richErr *RichError
	if !errors.As(rich, &richErr) {
		t.Fatal("expected RichError")
	}

	if len(richErr.Suggestions) == 0 {
		t.Error("expected suggestions for 404 error")
	}
}

func TestEnrichError_RateLimit(t *testing.T) {
	baseErr := &RateLimitError{
		RetryAfter: 30 * time.Second,
	}

	rich := EnrichError(baseErr, "orders", "")

	var richErr *RichError
	if !errors.As(rich, &richErr) {
		t.Fatal("expected RichError")
	}

	if richErr.RetryAfter != 30 {
		t.Errorf("expected RetryAfter=30, got %d", richErr.RetryAfter)
	}
}

func TestEnrichError_AuthError(t *testing.T) {
	baseErr := &AuthError{
		Reason: "token expired",
	}

	rich := EnrichError(baseErr, "orders", "")

	var richErr *RichError
	if !errors.As(rich, &richErr) {
		t.Fatal("expected RichError")
	}

	if richErr.Code != "AUTH_ERROR" {
		t.Errorf("expected code AUTH_ERROR, got %s", richErr.Code)
	}

	if len(richErr.Suggestions) == 0 {
		t.Error("expected suggestions for auth error")
	}
}

func TestEnrichError_ValidationError(t *testing.T) {
	baseErr := &ValidationError{
		Field:   "email",
		Message: "invalid format",
	}

	rich := EnrichError(baseErr, "customers", "")

	var richErr *RichError
	if !errors.As(rich, &richErr) {
		t.Fatal("expected RichError")
	}

	if richErr.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code VALIDATION_ERROR, got %s", richErr.Code)
	}

	if len(richErr.Suggestions) == 0 {
		t.Error("expected suggestions for validation error")
	}
}

func TestEnrichError_ServerError(t *testing.T) {
	for _, status := range []int{500, 502, 503} {
		baseErr := &APIError{
			Status:  status,
			Message: "Internal server error",
		}

		rich := EnrichError(baseErr, "orders", "")

		var richErr *RichError
		if !errors.As(rich, &richErr) {
			t.Fatalf("expected RichError for status %d", status)
		}

		if richErr.Code != "SERVER_ERROR" {
			t.Errorf("expected code SERVER_ERROR for status %d, got %s", status, richErr.Code)
		}
	}
}

func TestEnrichError_Conflict(t *testing.T) {
	baseErr := &APIError{
		Status:  409,
		Message: "Conflict",
	}

	rich := EnrichError(baseErr, "orders", "12345")

	var richErr *RichError
	if !errors.As(rich, &richErr) {
		t.Fatal("expected RichError")
	}

	if richErr.Code != "CONFLICT" {
		t.Errorf("expected code CONFLICT, got %s", richErr.Code)
	}
}

func TestEnrichError_Nil(t *testing.T) {
	rich := EnrichError(nil, "orders", "")

	if rich != nil {
		t.Error("expected nil for nil error")
	}
}

func TestRichError_Unwrap(t *testing.T) {
	cause := &APIError{Status: 404, Message: "not found"}
	rich := &RichError{
		Message: "Order not found",
		Cause:   cause,
	}

	if !errors.Is(rich, cause) {
		t.Error("expected to unwrap to cause")
	}
}

func TestFormatRichError(t *testing.T) {
	rich := &RichError{
		Message: "Order not found",
		Suggestions: []string{
			"Check the order ID",
			"Run 'spl orders list'",
		},
	}

	formatted := FormatRichError(rich)

	expected := "Order not found\n\nSuggestions:\n  • Check the order ID\n  • Run 'spl orders list'"
	if formatted != expected {
		t.Errorf("unexpected format:\ngot:  %q\nwant: %q", formatted, expected)
	}
}

func TestFormatRichError_NonRichError(t *testing.T) {
	err := errors.New("plain error")

	formatted := FormatRichError(err)

	if formatted != "plain error" {
		t.Errorf("unexpected format: %s", formatted)
	}
}

func TestFormatRichError_NoSuggestions(t *testing.T) {
	rich := &RichError{
		Message:     "Something went wrong",
		Suggestions: []string{},
	}

	formatted := FormatRichError(rich)

	if formatted != "Something went wrong" {
		t.Errorf("unexpected format: %s", formatted)
	}
}
