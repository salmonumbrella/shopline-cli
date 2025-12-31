package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Redirect represents a Shopline URL redirect.
type Redirect struct {
	ID        string    `json:"id"`
	Path      string    `json:"path"`
	Target    string    `json:"target"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RedirectsListOptions contains options for listing redirects.
type RedirectsListOptions struct {
	Page     int
	PageSize int
	Path     string
	Target   string
}

// RedirectsListResponse is the paginated response for redirects.
type RedirectsListResponse = ListResponse[Redirect]

// RedirectCreateRequest contains the data for creating a redirect.
type RedirectCreateRequest struct {
	Path   string `json:"path"`
	Target string `json:"target"`
}

// RedirectUpdateRequest contains the data for updating a redirect.
type RedirectUpdateRequest struct {
	Path   string `json:"path,omitempty"`
	Target string `json:"target,omitempty"`
}

// ListRedirects retrieves a list of redirects.
func (c *Client) ListRedirects(ctx context.Context, opts *RedirectsListOptions) (*RedirectsListResponse, error) {
	path := "/redirects"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("path", opts.Path).
			String("target", opts.Target).
			Build()
	}

	var resp RedirectsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetRedirect retrieves a single redirect by ID.
func (c *Client) GetRedirect(ctx context.Context, id string) (*Redirect, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("redirect id is required")
	}
	var redirect Redirect
	if err := c.Get(ctx, fmt.Sprintf("/redirects/%s", id), &redirect); err != nil {
		return nil, err
	}
	return &redirect, nil
}

// CreateRedirect creates a new redirect.
func (c *Client) CreateRedirect(ctx context.Context, req *RedirectCreateRequest) (*Redirect, error) {
	var redirect Redirect
	if err := c.Post(ctx, "/redirects", req, &redirect); err != nil {
		return nil, err
	}
	return &redirect, nil
}

// UpdateRedirect updates an existing redirect.
func (c *Client) UpdateRedirect(ctx context.Context, id string, req *RedirectUpdateRequest) (*Redirect, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("redirect id is required")
	}
	var redirect Redirect
	if err := c.Put(ctx, fmt.Sprintf("/redirects/%s", id), req, &redirect); err != nil {
		return nil, err
	}
	return &redirect, nil
}

// DeleteRedirect deletes a redirect.
func (c *Client) DeleteRedirect(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("redirect id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/redirects/%s", id))
}
