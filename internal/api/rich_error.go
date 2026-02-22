package api

import (
	"errors"
	"fmt"
)

// RichError wraps an error with actionable suggestions.
type RichError struct {
	Message     string   // Human-readable message
	Code        string   // Error code (e.g., "NOT_FOUND", "RATE_LIMITED")
	Suggestions []string // Actionable next steps
	Resource    string   // Resource type (e.g., "orders")
	ResourceID  string   // Resource ID if applicable
	RetryAfter  int      // Seconds to wait before retry (for rate limits)
	Cause       error    // Underlying error
}

func (e *RichError) Error() string {
	return e.Message
}

func (e *RichError) Unwrap() error {
	return e.Cause
}

// EnrichError wraps an error with context-aware suggestions.
func EnrichError(err error, resource, resourceID string) error {
	if err == nil {
		return nil
	}

	rich := &RichError{
		Message:    err.Error(),
		Resource:   resource,
		ResourceID: resourceID,
		Cause:      err,
	}

	// Enrich based on error type
	var apiErr *APIError
	var rateLimitErr *RateLimitError
	var authErr *AuthError
	var validationErr *ValidationError

	switch {
	case errors.As(err, &rateLimitErr):
		rich.Code = "RATE_LIMITED"
		rich.RetryAfter = int(rateLimitErr.RetryAfter.Seconds())
		rich.Suggestions = []string{
			fmt.Sprintf("Wait %d seconds before retrying", rich.RetryAfter),
			"Consider using --limit to reduce request size",
			"Batch operations when possible",
		}

	case errors.As(err, &authErr):
		rich.Code = "AUTH_ERROR"
		rich.Suggestions = []string{
			"Run 'spl auth login' to re-authenticate",
			"Check your API credentials are valid",
			"Verify you have permission for this operation",
		}

	case errors.As(err, &validationErr):
		rich.Code = "VALIDATION_ERROR"
		rich.Suggestions = []string{
			fmt.Sprintf("Check the '%s' field value", validationErr.Field),
			fmt.Sprintf("Run 'spl schema %s' for field requirements", resource),
		}

	case errors.As(err, &apiErr):
		switch apiErr.Status {
		case 404:
			rich.Code = "NOT_FOUND"
			rich.Suggestions = []string{
				fmt.Sprintf("Verify the %s ID '%s' is correct", resource, resourceID),
				fmt.Sprintf("Run 'spl %s list' to see available %s", resource, resource),
			}
		case 409:
			rich.Code = "CONFLICT"
			rich.Suggestions = []string{
				"The resource may have been modified by another process",
				"Fetch the latest version and retry",
			}
		case 500, 502, 503:
			rich.Code = "SERVER_ERROR"
			rich.Suggestions = []string{
				"This is a temporary server issue",
				"Wait a moment and retry the command",
				"Check Shopline status page if the issue persists",
			}
		default:
			rich.Code = fmt.Sprintf("HTTP_%d", apiErr.Status)
		}

	default:
		rich.Code = "UNKNOWN"
	}

	return rich
}

// FormatRichError returns a formatted error message with suggestions.
func FormatRichError(err error) string {
	var rich *RichError
	if !errors.As(err, &rich) {
		return err.Error()
	}

	msg := rich.Message
	if len(rich.Suggestions) > 0 {
		msg += "\n\nSuggestions:"
		for _, s := range rich.Suggestions {
			msg += "\n  • " + s
		}
	}
	return msg
}
