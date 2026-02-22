package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Label represents a Shopline product label.
type Label struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	Icon        string    `json:"icon"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LabelsListOptions contains options for listing labels.
type LabelsListOptions struct {
	Page     int
	PageSize int
	Active   *bool
}

// LabelsListResponse is the paginated response for labels.
type LabelsListResponse = ListResponse[Label]

// LabelCreateRequest contains the data for creating a label.
type LabelCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Active      bool   `json:"active"`
}

// LabelUpdateRequest contains the data for updating a label.
type LabelUpdateRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Active      *bool  `json:"active,omitempty"`
}

// ListLabels retrieves a list of labels.
func (c *Client) ListLabels(ctx context.Context, opts *LabelsListOptions) (*LabelsListResponse, error) {
	path := "/labels"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			BoolPtr("active", opts.Active).
			Build()
	}

	var resp LabelsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetLabel retrieves a single label by ID.
func (c *Client) GetLabel(ctx context.Context, id string) (*Label, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("label id is required")
	}
	var label Label
	if err := c.Get(ctx, fmt.Sprintf("/labels/%s", id), &label); err != nil {
		return nil, err
	}
	return &label, nil
}

// CreateLabel creates a new label.
func (c *Client) CreateLabel(ctx context.Context, req *LabelCreateRequest) (*Label, error) {
	var label Label
	if err := c.Post(ctx, "/labels", req, &label); err != nil {
		return nil, err
	}
	return &label, nil
}

// UpdateLabel updates an existing label.
func (c *Client) UpdateLabel(ctx context.Context, id string, req *LabelUpdateRequest) (*Label, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("label id is required")
	}
	var label Label
	if err := c.Put(ctx, fmt.Sprintf("/labels/%s", id), req, &label); err != nil {
		return nil, err
	}
	return &label, nil
}

// DeleteLabel deletes a label.
func (c *Client) DeleteLabel(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("label id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/labels/%s", id))
}
