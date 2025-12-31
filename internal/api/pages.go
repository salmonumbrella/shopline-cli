package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Page represents a Shopline page.
type Page struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Handle         string    `json:"handle"`
	BodyHTML       string    `json:"body_html"`
	Author         string    `json:"author"`
	TemplateSuffix string    `json:"template_suffix"`
	Published      bool      `json:"published"`
	PublishedAt    time.Time `json:"published_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// PagesListOptions contains options for listing pages.
type PagesListOptions struct {
	Page      int
	PageSize  int
	Published *bool
	Title     string
}

// PagesListResponse contains the list response.
type PagesListResponse struct {
	Items      []Page `json:"items"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	TotalCount int    `json:"total_count"`
	HasMore    bool   `json:"has_more"`
}

// PageCreateRequest contains the request body for creating a page.
type PageCreateRequest struct {
	Title          string     `json:"title"`
	Handle         string     `json:"handle,omitempty"`
	BodyHTML       string     `json:"body_html,omitempty"`
	Author         string     `json:"author,omitempty"`
	TemplateSuffix string     `json:"template_suffix,omitempty"`
	Published      bool       `json:"published,omitempty"`
	PublishedAt    *time.Time `json:"published_at,omitempty"`
}

// PageUpdateRequest contains the request body for updating a page.
type PageUpdateRequest struct {
	Title          string     `json:"title,omitempty"`
	Handle         string     `json:"handle,omitempty"`
	BodyHTML       string     `json:"body_html,omitempty"`
	Author         string     `json:"author,omitempty"`
	TemplateSuffix string     `json:"template_suffix,omitempty"`
	Published      *bool      `json:"published,omitempty"`
	PublishedAt    *time.Time `json:"published_at,omitempty"`
}

// ListPages retrieves a list of pages.
func (c *Client) ListPages(ctx context.Context, opts *PagesListOptions) (*PagesListResponse, error) {
	path := "/pages"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			BoolPtr("published", opts.Published).
			String("title", opts.Title).
			Build()
	}

	var resp PagesListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPage retrieves a single page by ID.
func (c *Client) GetPage(ctx context.Context, id string) (*Page, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("page id is required")
	}
	var page Page
	if err := c.Get(ctx, fmt.Sprintf("/pages/%s", id), &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// CreatePage creates a new page.
func (c *Client) CreatePage(ctx context.Context, req *PageCreateRequest) (*Page, error) {
	var page Page
	if err := c.Post(ctx, "/pages", req, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// UpdatePage updates an existing page.
func (c *Client) UpdatePage(ctx context.Context, id string, req *PageUpdateRequest) (*Page, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("page id is required")
	}
	var page Page
	if err := c.Put(ctx, fmt.Sprintf("/pages/%s", id), req, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// DeletePage deletes a page.
func (c *Client) DeletePage(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("page id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/pages/%s", id))
}
