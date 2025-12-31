package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// StorefrontToken represents a Shopline storefront access token.
type StorefrontToken struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	AccessToken string    `json:"access_token,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// StorefrontTokensListOptions contains options for listing storefront tokens.
type StorefrontTokensListOptions struct {
	Page     int
	PageSize int
}

// StorefrontTokensListResponse is the paginated response for storefront tokens.
type StorefrontTokensListResponse = ListResponse[StorefrontToken]

// StorefrontTokenCreateRequest contains the data for creating a storefront token.
type StorefrontTokenCreateRequest struct {
	Title string `json:"title"`
}

// ListStorefrontTokens retrieves a list of storefront access tokens.
func (c *Client) ListStorefrontTokens(ctx context.Context, opts *StorefrontTokensListOptions) (*StorefrontTokensListResponse, error) {
	path := "/storefront_tokens"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			Build()
	}

	var resp StorefrontTokensListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetStorefrontToken retrieves a single storefront token by ID.
func (c *Client) GetStorefrontToken(ctx context.Context, id string) (*StorefrontToken, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("storefront token id is required")
	}
	var token StorefrontToken
	if err := c.Get(ctx, fmt.Sprintf("/storefront_tokens/%s", id), &token); err != nil {
		return nil, err
	}
	return &token, nil
}

// CreateStorefrontToken creates a new storefront access token.
func (c *Client) CreateStorefrontToken(ctx context.Context, req *StorefrontTokenCreateRequest) (*StorefrontToken, error) {
	var token StorefrontToken
	if err := c.Post(ctx, "/storefront_tokens", req, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

// DeleteStorefrontToken revokes and deletes a storefront access token.
func (c *Client) DeleteStorefrontToken(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("storefront token id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/storefront_tokens/%s", id))
}
