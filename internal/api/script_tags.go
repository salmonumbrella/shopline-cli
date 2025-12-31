package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ScriptTag represents a Shopline script tag.
type ScriptTag struct {
	ID           string    `json:"id"`
	Src          string    `json:"src"`
	Event        string    `json:"event"`
	DisplayScope string    `json:"display_scope"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ScriptTagsListOptions contains options for listing script tags.
type ScriptTagsListOptions struct {
	Page     int
	PageSize int
	Src      string
}

// ScriptTagsListResponse contains the list response.
type ScriptTagsListResponse struct {
	Items      []ScriptTag `json:"items"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalCount int         `json:"total_count"`
	HasMore    bool        `json:"has_more"`
}

// ScriptTagCreateRequest contains the request body for creating a script tag.
type ScriptTagCreateRequest struct {
	Src          string `json:"src"`
	Event        string `json:"event,omitempty"`
	DisplayScope string `json:"display_scope,omitempty"`
}

// ScriptTagUpdateRequest contains the request body for updating a script tag.
type ScriptTagUpdateRequest struct {
	Src          string `json:"src,omitempty"`
	Event        string `json:"event,omitempty"`
	DisplayScope string `json:"display_scope,omitempty"`
}

// ListScriptTags retrieves a list of script tags.
func (c *Client) ListScriptTags(ctx context.Context, opts *ScriptTagsListOptions) (*ScriptTagsListResponse, error) {
	path := "/script_tags"
	if opts != nil {
		params := url.Values{}
		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			params.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.Src != "" {
			params.Set("src", opts.Src)
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp ScriptTagsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetScriptTag retrieves a single script tag by ID.
func (c *Client) GetScriptTag(ctx context.Context, id string) (*ScriptTag, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("script tag id is required")
	}
	var tag ScriptTag
	if err := c.Get(ctx, fmt.Sprintf("/script_tags/%s", id), &tag); err != nil {
		return nil, err
	}
	return &tag, nil
}

// CreateScriptTag creates a new script tag.
func (c *Client) CreateScriptTag(ctx context.Context, req *ScriptTagCreateRequest) (*ScriptTag, error) {
	var tag ScriptTag
	if err := c.Post(ctx, "/script_tags", req, &tag); err != nil {
		return nil, err
	}
	return &tag, nil
}

// UpdateScriptTag updates an existing script tag.
func (c *Client) UpdateScriptTag(ctx context.Context, id string, req *ScriptTagUpdateRequest) (*ScriptTag, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("script tag id is required")
	}
	var tag ScriptTag
	if err := c.Put(ctx, fmt.Sprintf("/script_tags/%s", id), req, &tag); err != nil {
		return nil, err
	}
	return &tag, nil
}

// DeleteScriptTag deletes a script tag.
func (c *Client) DeleteScriptTag(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("script tag id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/script_tags/%s", id))
}
