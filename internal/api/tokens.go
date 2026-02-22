package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Token represents a Shopline API token.
type Token struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	AccessToken string     `json:"access_token,omitempty"`
	Scopes      []string   `json:"scopes"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TokensListOptions contains options for listing tokens.
type TokensListOptions struct {
	Page     int
	PageSize int
}

// TokensListResponse is the paginated response for tokens.
type TokensListResponse = ListResponse[Token]

// TokenCreateRequest contains the data for creating a token.
type TokenCreateRequest struct {
	Title  string   `json:"title"`
	Scopes []string `json:"scopes"`
}

// ListTokens retrieves a list of API tokens.
func (c *Client) ListTokens(ctx context.Context, opts *TokensListOptions) (*TokensListResponse, error) {
	path := "/tokens"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			Build()
	}

	var resp TokensListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetToken retrieves a single token by ID.
func (c *Client) GetToken(ctx context.Context, id string) (*Token, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("token id is required")
	}
	var token Token
	if err := c.Get(ctx, fmt.Sprintf("/tokens/%s", id), &token); err != nil {
		return nil, err
	}
	return &token, nil
}

// CreateToken creates a new API token.
func (c *Client) CreateToken(ctx context.Context, req *TokenCreateRequest) (*Token, error) {
	var token Token
	if err := c.Post(ctx, "/tokens", req, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

// DeleteToken revokes and deletes an API token.
func (c *Client) DeleteToken(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("token id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/tokens/%s", id))
}
