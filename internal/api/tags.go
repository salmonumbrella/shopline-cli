package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Tag represents a Shopline product tag.
type Tag struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Handle       string    `json:"handle"`
	ProductCount int       `json:"product_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TagsListOptions contains options for listing tags.
type TagsListOptions struct {
	Page     int
	PageSize int
	Query    string
}

// TagsListResponse is the paginated response for tags.
type TagsListResponse = ListResponse[Tag]

// TagCreateRequest contains the data for creating a tag.
type TagCreateRequest struct {
	Name string `json:"name"`
}

// ListTags retrieves a list of tags.
func (c *Client) ListTags(ctx context.Context, opts *TagsListOptions) (*TagsListResponse, error) {
	path := "/tags"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("query", opts.Query).
			Build()
	}

	var resp TagsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTag retrieves a single tag by ID.
func (c *Client) GetTag(ctx context.Context, id string) (*Tag, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("tag id is required")
	}
	var tag Tag
	if err := c.Get(ctx, fmt.Sprintf("/tags/%s", id), &tag); err != nil {
		return nil, err
	}
	return &tag, nil
}

// CreateTag creates a new tag.
func (c *Client) CreateTag(ctx context.Context, req *TagCreateRequest) (*Tag, error) {
	var tag Tag
	if err := c.Post(ctx, "/tags", req, &tag); err != nil {
		return nil, err
	}
	return &tag, nil
}

// DeleteTag deletes a tag.
func (c *Client) DeleteTag(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("tag id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/tags/%s", id))
}
