package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Theme represents a Shopline theme.
type Theme struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Role        string    `json:"role"` // main, mobile, unpublished
	Previewable bool      `json:"previewable"`
	Processing  bool      `json:"processing"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ThemesListOptions contains options for listing themes.
type ThemesListOptions struct {
	Page     int
	PageSize int
	Role     string
}

// ThemesListResponse is the paginated response for themes.
type ThemesListResponse = ListResponse[Theme]

// ThemeCreateRequest contains the data for creating a theme.
type ThemeCreateRequest struct {
	Name string `json:"name"`
	Role string `json:"role,omitempty"`
	Src  string `json:"src,omitempty"` // URL to theme zip
}

// ThemeUpdateRequest contains the data for updating a theme.
type ThemeUpdateRequest struct {
	Name string `json:"name,omitempty"`
	Role string `json:"role,omitempty"`
}

// ListThemes retrieves a list of themes.
func (c *Client) ListThemes(ctx context.Context, opts *ThemesListOptions) (*ThemesListResponse, error) {
	path := "/themes"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("role", opts.Role).
			Build()
	}

	var resp ThemesListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTheme retrieves a single theme by ID.
func (c *Client) GetTheme(ctx context.Context, id string) (*Theme, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("theme id is required")
	}
	var theme Theme
	if err := c.Get(ctx, fmt.Sprintf("/themes/%s", id), &theme); err != nil {
		return nil, err
	}
	return &theme, nil
}

// CreateTheme creates a new theme.
func (c *Client) CreateTheme(ctx context.Context, req *ThemeCreateRequest) (*Theme, error) {
	var theme Theme
	if err := c.Post(ctx, "/themes", req, &theme); err != nil {
		return nil, err
	}
	return &theme, nil
}

// UpdateTheme updates an existing theme.
func (c *Client) UpdateTheme(ctx context.Context, id string, req *ThemeUpdateRequest) (*Theme, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("theme id is required")
	}
	var theme Theme
	if err := c.Put(ctx, fmt.Sprintf("/themes/%s", id), req, &theme); err != nil {
		return nil, err
	}
	return &theme, nil
}

// DeleteTheme deletes a theme.
func (c *Client) DeleteTheme(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("theme id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/themes/%s", id))
}
