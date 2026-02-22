package api

import (
	"context"
	"time"
)

// Multipass represents Shopline multipass configuration.
type Multipass struct {
	Enabled   bool      `json:"enabled"`
	Secret    string    `json:"secret,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MultipassToken represents a generated multipass token response.
type MultipassToken struct {
	Token     string    `json:"token"`
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// MultipassTokenRequest contains the data for generating a multipass token.
type MultipassTokenRequest struct {
	Email        string                 `json:"email"`
	ReturnTo     string                 `json:"return_to,omitempty"`
	CustomerData map[string]interface{} `json:"customer_data,omitempty"`
}

// GetMultipass retrieves the multipass configuration.
func (c *Client) GetMultipass(ctx context.Context) (*Multipass, error) {
	var multipass Multipass
	if err := c.Get(ctx, "/multipass", &multipass); err != nil {
		return nil, err
	}
	return &multipass, nil
}

// EnableMultipass enables multipass authentication and returns the secret.
func (c *Client) EnableMultipass(ctx context.Context) (*Multipass, error) {
	var multipass Multipass
	if err := c.Post(ctx, "/multipass/enable", nil, &multipass); err != nil {
		return nil, err
	}
	return &multipass, nil
}

// DisableMultipass disables multipass authentication.
func (c *Client) DisableMultipass(ctx context.Context) error {
	return c.Post(ctx, "/multipass/disable", nil, nil)
}

// RotateMultipassSecret generates a new multipass secret.
func (c *Client) RotateMultipassSecret(ctx context.Context) (*Multipass, error) {
	var multipass Multipass
	if err := c.Post(ctx, "/multipass/rotate", nil, &multipass); err != nil {
		return nil, err
	}
	return &multipass, nil
}

// GenerateMultipassToken generates a multipass login token for a customer.
func (c *Client) GenerateMultipassToken(ctx context.Context, req *MultipassTokenRequest) (*MultipassToken, error) {
	var token MultipassToken
	if err := c.Post(ctx, "/multipass/token", req, &token); err != nil {
		return nil, err
	}
	return &token, nil
}
