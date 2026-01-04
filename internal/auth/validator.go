package auth

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"
)

const (
	validationTimeout = 10 * time.Second
)

// handlePattern validates store handles: alphanumeric and hyphens only
var handlePattern = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

// isValidHandle checks if a handle contains only safe characters.
// This prevents URL injection attacks where a malicious handle could
// redirect requests to arbitrary hosts.
func isValidHandle(handle string) bool {
	return handlePattern.MatchString(handle)
}

// CredentialValidator validates Shopline API credentials.
type CredentialValidator interface {
	Validate(ctx context.Context, handle, accessToken string) error
}

// URLBuilder constructs the validation URL for a given handle.
type URLBuilder func(handle string) string

// DefaultURLBuilder returns the production Shopline Open API URL.
// The handle is not used in the URL since Open API tokens are store-scoped.
func DefaultURLBuilder(handle string) string {
	return "https://open.shopline.io/v1/orders?per_page=1"
}

// APICredentialValidator validates credentials by making a test API call.
type APICredentialValidator struct {
	httpClient *http.Client
	urlBuilder URLBuilder
}

// NewAPICredentialValidator creates a validator with a short timeout for auth flow.
func NewAPICredentialValidator() *APICredentialValidator {
	return NewAPICredentialValidatorWithURLBuilder(DefaultURLBuilder)
}

// NewAPICredentialValidatorWithURLBuilder creates a validator with a custom URL builder.
// This is useful for testing with a mock server.
func NewAPICredentialValidatorWithURLBuilder(urlBuilder URLBuilder) *APICredentialValidator {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	return &APICredentialValidator{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   validationTimeout,
		},
		urlBuilder: urlBuilder,
	}
}

// Validate checks if the credentials are valid by making a test API call.
func (v *APICredentialValidator) Validate(ctx context.Context, handle, accessToken string) error {
	if handle == "" {
		return &CredentialValidationError{Message: "handle is required"}
	}
	if !isValidHandle(handle) {
		return &CredentialValidationError{Message: "handle contains invalid characters (only alphanumeric and hyphens allowed)"}
	}
	if accessToken == "" {
		return &CredentialValidationError{Message: "access token is required"}
	}

	url := v.urlBuilder(handle)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("could not validate credentials: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Shopline-Access-Token", accessToken)

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not validate credentials: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		var apiErr struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return &CredentialValidationError{Message: "invalid credentials"}
		}
		if apiErr.Message != "" {
			return &CredentialValidationError{Message: fmt.Sprintf("invalid credentials: %s", apiErr.Message)}
		}
		return &CredentialValidationError{Message: "invalid credentials"}
	}

	if resp.StatusCode >= 500 {
		return fmt.Errorf("could not validate credentials: server error (status %d)", resp.StatusCode)
	}

	if resp.StatusCode >= 400 {
		var apiErr struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err == nil && apiErr.Message != "" {
			return &CredentialValidationError{Message: apiErr.Message}
		}
		return &CredentialValidationError{Message: fmt.Sprintf("validation failed (status %d)", resp.StatusCode)}
	}

	return nil
}

// CredentialValidationError represents a credential validation failure.
// This is distinct from api.ValidationError which represents field validation errors.
type CredentialValidationError struct {
	Message string
}

func (e *CredentialValidationError) Error() string {
	return e.Message
}
