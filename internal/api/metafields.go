package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Metafield represents a Shopline metafield.
type Metafield struct {
	ID          string    `json:"id"`
	Namespace   string    `json:"namespace"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	ValueType   string    `json:"value_type"`
	Description string    `json:"description"`
	OwnerID     string    `json:"owner_id"`
	OwnerType   string    `json:"owner_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MetafieldsListOptions contains options for listing metafields.
type MetafieldsListOptions struct {
	Page      int
	PageSize  int
	Namespace string
	Key       string
	OwnerID   string
	OwnerType string
}

// MetafieldsListResponse is the paginated response for metafields.
type MetafieldsListResponse = ListResponse[Metafield]

// MetafieldCreateRequest contains the request body for creating a metafield.
type MetafieldCreateRequest struct {
	Namespace   string `json:"namespace"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	ValueType   string `json:"value_type"`
	Description string `json:"description,omitempty"`
	OwnerID     string `json:"owner_id,omitempty"`
	OwnerType   string `json:"owner_type,omitempty"`
}

// MetafieldUpdateRequest contains the request body for updating a metafield.
type MetafieldUpdateRequest struct {
	Value       string `json:"value,omitempty"`
	ValueType   string `json:"value_type,omitempty"`
	Description string `json:"description,omitempty"`
}

// ListMetafields retrieves a list of metafields.
func (c *Client) ListMetafields(ctx context.Context, opts *MetafieldsListOptions) (*MetafieldsListResponse, error) {
	path := "/metafields"
	if opts != nil {
		params := url.Values{}
		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			params.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.Namespace != "" {
			params.Set("namespace", opts.Namespace)
		}
		if opts.Key != "" {
			params.Set("key", opts.Key)
		}
		if opts.OwnerID != "" {
			params.Set("owner_id", opts.OwnerID)
		}
		if opts.OwnerType != "" {
			params.Set("owner_type", opts.OwnerType)
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp MetafieldsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetMetafield retrieves a single metafield by ID.
func (c *Client) GetMetafield(ctx context.Context, id string) (*Metafield, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("metafield id is required")
	}
	var metafield Metafield
	if err := c.Get(ctx, fmt.Sprintf("/metafields/%s", id), &metafield); err != nil {
		return nil, err
	}
	return &metafield, nil
}

// CreateMetafield creates a new metafield.
func (c *Client) CreateMetafield(ctx context.Context, req *MetafieldCreateRequest) (*Metafield, error) {
	var metafield Metafield
	if err := c.Post(ctx, "/metafields", req, &metafield); err != nil {
		return nil, err
	}
	return &metafield, nil
}

// UpdateMetafield updates an existing metafield.
func (c *Client) UpdateMetafield(ctx context.Context, id string, req *MetafieldUpdateRequest) (*Metafield, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("metafield id is required")
	}
	var metafield Metafield
	if err := c.Put(ctx, fmt.Sprintf("/metafields/%s", id), req, &metafield); err != nil {
		return nil, err
	}
	return &metafield, nil
}

// DeleteMetafield deletes a metafield.
func (c *Client) DeleteMetafield(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("metafield id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/metafields/%s", id))
}
