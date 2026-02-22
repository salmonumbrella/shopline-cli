package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Blog represents a Shopline blog.
type Blog struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Handle         string    `json:"handle"`
	Commentable    string    `json:"commentable"` // no, moderate, yes
	Tags           string    `json:"tags"`
	TemplateSuffix string    `json:"template_suffix"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// BlogsListOptions contains options for listing blogs.
type BlogsListOptions struct {
	Page     int
	PageSize int
}

// BlogsListResponse is the paginated response for blogs.
type BlogsListResponse = ListResponse[Blog]

// BlogCreateRequest contains the data for creating a blog.
type BlogCreateRequest struct {
	Title       string `json:"title"`
	Handle      string `json:"handle,omitempty"`
	Commentable string `json:"commentable,omitempty"`
}

// BlogUpdateRequest contains the data for updating a blog.
type BlogUpdateRequest struct {
	Title       string `json:"title,omitempty"`
	Handle      string `json:"handle,omitempty"`
	Commentable string `json:"commentable,omitempty"`
}

// ListBlogs retrieves a list of blogs.
func (c *Client) ListBlogs(ctx context.Context, opts *BlogsListOptions) (*BlogsListResponse, error) {
	path := "/blogs"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			Build()
	}

	var resp BlogsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetBlog retrieves a single blog by ID.
func (c *Client) GetBlog(ctx context.Context, id string) (*Blog, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("blog id is required")
	}
	var blog Blog
	if err := c.Get(ctx, fmt.Sprintf("/blogs/%s", id), &blog); err != nil {
		return nil, err
	}
	return &blog, nil
}

// CreateBlog creates a new blog.
func (c *Client) CreateBlog(ctx context.Context, req *BlogCreateRequest) (*Blog, error) {
	var blog Blog
	if err := c.Post(ctx, "/blogs", req, &blog); err != nil {
		return nil, err
	}
	return &blog, nil
}

// UpdateBlog updates an existing blog.
func (c *Client) UpdateBlog(ctx context.Context, id string, req *BlogUpdateRequest) (*Blog, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("blog id is required")
	}
	var blog Blog
	if err := c.Put(ctx, fmt.Sprintf("/blogs/%s", id), req, &blog); err != nil {
		return nil, err
	}
	return &blog, nil
}

// DeleteBlog deletes a blog.
func (c *Client) DeleteBlog(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("blog id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/blogs/%s", id))
}
