package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// MetafieldDefinition represents a Shopline metafield definition.
type MetafieldDefinition struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Namespace   string       `json:"namespace"`
	Key         string       `json:"key"`
	Description string       `json:"description"`
	Type        string       `json:"type"`
	OwnerType   string       `json:"owner_type"`
	Validations []Validation `json:"validations"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// Validation represents a metafield validation rule.
type Validation struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// MetafieldDefinitionsListOptions contains options for listing metafield definitions.
type MetafieldDefinitionsListOptions struct {
	Page      int
	PageSize  int
	OwnerType string
	Namespace string
}

// MetafieldDefinitionsListResponse contains the list response.
type MetafieldDefinitionsListResponse struct {
	Items      []MetafieldDefinition `json:"items"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalCount int                   `json:"total_count"`
	HasMore    bool                  `json:"has_more"`
}

// MetafieldDefinitionCreateRequest contains the request body for creating a metafield definition.
type MetafieldDefinitionCreateRequest struct {
	Name        string       `json:"name"`
	Namespace   string       `json:"namespace"`
	Key         string       `json:"key"`
	Description string       `json:"description,omitempty"`
	Type        string       `json:"type"`
	OwnerType   string       `json:"owner_type"`
	Validations []Validation `json:"validations,omitempty"`
}

// MetafieldDefinitionUpdateRequest contains the request body for updating a metafield definition.
type MetafieldDefinitionUpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// ListMetafieldDefinitions retrieves a list of metafield definitions.
func (c *Client) ListMetafieldDefinitions(ctx context.Context, opts *MetafieldDefinitionsListOptions) (*MetafieldDefinitionsListResponse, error) {
	path := "/metafield_definitions"
	if opts != nil {
		params := url.Values{}
		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			params.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.OwnerType != "" {
			params.Set("owner_type", opts.OwnerType)
		}
		if opts.Namespace != "" {
			params.Set("namespace", opts.Namespace)
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp MetafieldDefinitionsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetMetafieldDefinition retrieves a single metafield definition by ID.
func (c *Client) GetMetafieldDefinition(ctx context.Context, id string) (*MetafieldDefinition, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("metafield definition id is required")
	}
	var def MetafieldDefinition
	if err := c.Get(ctx, fmt.Sprintf("/metafield_definitions/%s", id), &def); err != nil {
		return nil, err
	}
	return &def, nil
}

// CreateMetafieldDefinition creates a new metafield definition.
func (c *Client) CreateMetafieldDefinition(ctx context.Context, req *MetafieldDefinitionCreateRequest) (*MetafieldDefinition, error) {
	var def MetafieldDefinition
	if err := c.Post(ctx, "/metafield_definitions", req, &def); err != nil {
		return nil, err
	}
	return &def, nil
}

// UpdateMetafieldDefinition updates a metafield definition.
func (c *Client) UpdateMetafieldDefinition(ctx context.Context, id string, req *MetafieldDefinitionUpdateRequest) (*MetafieldDefinition, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("metafield definition id is required")
	}
	var def MetafieldDefinition
	if err := c.Put(ctx, fmt.Sprintf("/metafield_definitions/%s", id), req, &def); err != nil {
		return nil, err
	}
	return &def, nil
}

// DeleteMetafieldDefinition deletes a metafield definition.
func (c *Client) DeleteMetafieldDefinition(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("metafield definition id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/metafield_definitions/%s", id))
}
