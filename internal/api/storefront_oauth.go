package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// StorefrontOAuthClient represents a Shopline storefront OAuth client.
type StorefrontOAuthClient struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret,omitempty"`
	RedirectURIs []string  `json:"redirect_uris"`
	Scopes       []string  `json:"scopes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// StorefrontOAuthClientsListOptions contains options for listing OAuth clients.
type StorefrontOAuthClientsListOptions struct {
	Page     int
	PageSize int
}

// StorefrontOAuthClientsListResponse is the paginated response for OAuth clients.
type StorefrontOAuthClientsListResponse = ListResponse[StorefrontOAuthClient]

// StorefrontOAuthClientCreateRequest contains the data for creating an OAuth client.
type StorefrontOAuthClientCreateRequest struct {
	Name         string   `json:"name"`
	RedirectURIs []string `json:"redirect_uris"`
	Scopes       []string `json:"scopes,omitempty"`
}

// StorefrontOAuthClientUpdateRequest contains the data for updating an OAuth client.
type StorefrontOAuthClientUpdateRequest struct {
	Name         string   `json:"name,omitempty"`
	RedirectURIs []string `json:"redirect_uris,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
}

// ListStorefrontOAuthClients retrieves a list of storefront OAuth clients.
func (c *Client) ListStorefrontOAuthClients(ctx context.Context, opts *StorefrontOAuthClientsListOptions) (*StorefrontOAuthClientsListResponse, error) {
	path := "/storefront_oauth/clients"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			Build()
	}

	var resp StorefrontOAuthClientsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetStorefrontOAuthClient retrieves a single OAuth client by ID.
func (c *Client) GetStorefrontOAuthClient(ctx context.Context, id string) (*StorefrontOAuthClient, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("oauth client id is required")
	}
	var client StorefrontOAuthClient
	if err := c.Get(ctx, fmt.Sprintf("/storefront_oauth/clients/%s", id), &client); err != nil {
		return nil, err
	}
	return &client, nil
}

// CreateStorefrontOAuthClient creates a new OAuth client.
func (c *Client) CreateStorefrontOAuthClient(ctx context.Context, req *StorefrontOAuthClientCreateRequest) (*StorefrontOAuthClient, error) {
	var client StorefrontOAuthClient
	if err := c.Post(ctx, "/storefront_oauth/clients", req, &client); err != nil {
		return nil, err
	}
	return &client, nil
}

// UpdateStorefrontOAuthClient updates an existing OAuth client.
func (c *Client) UpdateStorefrontOAuthClient(ctx context.Context, id string, req *StorefrontOAuthClientUpdateRequest) (*StorefrontOAuthClient, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("oauth client id is required")
	}
	var client StorefrontOAuthClient
	if err := c.Put(ctx, fmt.Sprintf("/storefront_oauth/clients/%s", id), req, &client); err != nil {
		return nil, err
	}
	return &client, nil
}

// DeleteStorefrontOAuthClient deletes an OAuth client.
func (c *Client) DeleteStorefrontOAuthClient(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("oauth client id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/storefront_oauth/clients/%s", id))
}

// RotateStorefrontOAuthClientSecret generates a new client secret.
func (c *Client) RotateStorefrontOAuthClientSecret(ctx context.Context, id string) (*StorefrontOAuthClient, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("oauth client id is required")
	}
	var client StorefrontOAuthClient
	if err := c.Post(ctx, fmt.Sprintf("/storefront_oauth/clients/%s/rotate_secret", id), nil, &client); err != nil {
		return nil, err
	}
	return &client, nil
}
