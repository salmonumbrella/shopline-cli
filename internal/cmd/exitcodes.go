package cmd

import (
	"errors"

	"github.com/salmonumbrella/shopline-cli/internal/api"
)

// Exit codes for the CLI.
const (
	ExitSuccess    = 0
	ExitGeneral    = 1
	ExitAuth       = 2
	ExitValidation = 3
	ExitNotFound   = 4
	ExitRateLimit  = 5
)

// ExitError wraps an error with an exit code.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	return e.Err.Error()
}

func (e *ExitError) Unwrap() error {
	return e.Err
}

// GetExitCode returns the appropriate exit code for an error.
func GetExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}

	var exitErr *ExitError
	if errors.As(err, &exitErr) {
		return exitErr.Code
	}

	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.Status {
		case 401, 403:
			return ExitAuth
		case 404:
			return ExitNotFound
		case 400, 422:
			return ExitValidation
		case 429:
			return ExitRateLimit
		}
	}

	var authErr *api.AuthError
	if errors.As(err, &authErr) {
		return ExitAuth
	}

	var valErr *api.ValidationError
	if errors.As(err, &valErr) {
		return ExitValidation
	}

	var rateErr *api.RateLimitError
	if errors.As(err, &rateErr) {
		return ExitRateLimit
	}

	return ExitGeneral
}
