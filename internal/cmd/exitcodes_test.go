package cmd

import (
	"errors"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
)

func TestGetExitCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "nil error returns success",
			err:      nil,
			expected: ExitSuccess,
		},
		{
			name:     "generic error returns general",
			err:      errors.New("something went wrong"),
			expected: ExitGeneral,
		},
		{
			name: "ExitError returns its code",
			err: &ExitError{
				Code: ExitValidation,
				Err:  errors.New("bad input"),
			},
			expected: ExitValidation,
		},
		{
			name: "wrapped ExitError returns its code",
			err: errors.Join(
				errors.New("context"),
				&ExitError{Code: ExitAuth, Err: errors.New("auth failed")},
			),
			expected: ExitAuth,
		},
		{
			name: "API 401 returns auth",
			err: &api.APIError{
				Code:    "UNAUTHORIZED",
				Message: "Invalid token",
				Status:  401,
			},
			expected: ExitAuth,
		},
		{
			name: "API 403 returns auth",
			err: &api.APIError{
				Code:    "FORBIDDEN",
				Message: "Access denied",
				Status:  403,
			},
			expected: ExitAuth,
		},
		{
			name: "API 404 returns not found",
			err: &api.APIError{
				Code:    "NOT_FOUND",
				Message: "Order not found",
				Status:  404,
			},
			expected: ExitNotFound,
		},
		{
			name: "API 400 returns validation",
			err: &api.APIError{
				Code:    "BAD_REQUEST",
				Message: "Invalid parameter",
				Status:  400,
			},
			expected: ExitValidation,
		},
		{
			name: "API 422 returns validation",
			err: &api.APIError{
				Code:    "UNPROCESSABLE_ENTITY",
				Message: "Validation failed",
				Status:  422,
			},
			expected: ExitValidation,
		},
		{
			name: "API 429 returns rate limit",
			err: &api.APIError{
				Code:    "RATE_LIMITED",
				Message: "Too many requests",
				Status:  429,
			},
			expected: ExitRateLimit,
		},
		{
			name: "API 500 returns general",
			err: &api.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Server error",
				Status:  500,
			},
			expected: ExitGeneral,
		},
		{
			name: "AuthError returns auth",
			err: &api.AuthError{
				Reason: "no credentials",
			},
			expected: ExitAuth,
		},
		{
			name: "ValidationError returns validation",
			err: &api.ValidationError{
				Field:   "email",
				Message: "invalid format",
			},
			expected: ExitValidation,
		},
		{
			name: "RateLimitError returns rate limit",
			err: &api.RateLimitError{
				RetryAfter: 30 * time.Second,
			},
			expected: ExitRateLimit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetExitCode(tt.err)
			if got != tt.expected {
				t.Errorf("GetExitCode() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestExitErrorUnwrap(t *testing.T) {
	inner := errors.New("inner error")
	exitErr := &ExitError{
		Code: ExitAuth,
		Err:  inner,
	}

	if !errors.Is(exitErr, inner) {
		t.Error("ExitError should unwrap to inner error")
	}
}

func TestExitErrorError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ExitError
		expected string
	}{
		{
			name: "simple error message",
			err: &ExitError{
				Code: ExitGeneral,
				Err:  errors.New("something went wrong"),
			},
			expected: "something went wrong",
		},
		{
			name: "auth error message",
			err: &ExitError{
				Code: ExitAuth,
				Err:  errors.New("authentication failed"),
			},
			expected: "authentication failed",
		},
		{
			name: "validation error message",
			err: &ExitError{
				Code: ExitValidation,
				Err:  errors.New("invalid input"),
			},
			expected: "invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}
