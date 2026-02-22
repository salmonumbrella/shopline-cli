package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// LocalDeliveryZone represents a geographic zone for local delivery.
type LocalDeliveryZone struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	ZipCodes    []string `json:"zip_codes"`
	MinDistance float64  `json:"min_distance"`
	MaxDistance float64  `json:"max_distance"`
}

// LocalDeliveryOption represents a local delivery option configuration.
type LocalDeliveryOption struct {
	ID               string              `json:"id"`
	Name             string              `json:"name"`
	Description      string              `json:"description"`
	Active           bool                `json:"active"`
	Price            string              `json:"price"`
	FreeAbove        string              `json:"free_above"`
	Currency         string              `json:"currency"`
	MinOrderAmount   string              `json:"min_order_amount"`
	MaxOrderAmount   string              `json:"max_order_amount"`
	DeliveryTimeMin  int                 `json:"delivery_time_min"`
	DeliveryTimeMax  int                 `json:"delivery_time_max"`
	DeliveryTimeUnit string              `json:"delivery_time_unit"`
	Zones            []LocalDeliveryZone `json:"zones"`
	LocationID       string              `json:"location_id"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

// LocalDeliveryListOptions contains options for listing local delivery options.
type LocalDeliveryListOptions struct {
	Page       int
	PageSize   int
	Active     *bool
	LocationID string
}

// LocalDeliveryListResponse is the paginated response for local delivery options.
type LocalDeliveryListResponse = ListResponse[LocalDeliveryOption]

// LocalDeliveryCreateRequest contains the request body for creating a local delivery option.
type LocalDeliveryCreateRequest struct {
	Name             string              `json:"name"`
	Description      string              `json:"description,omitempty"`
	Active           bool                `json:"active"`
	Price            string              `json:"price"`
	FreeAbove        string              `json:"free_above,omitempty"`
	MinOrderAmount   string              `json:"min_order_amount,omitempty"`
	MaxOrderAmount   string              `json:"max_order_amount,omitempty"`
	DeliveryTimeMin  int                 `json:"delivery_time_min,omitempty"`
	DeliveryTimeMax  int                 `json:"delivery_time_max,omitempty"`
	DeliveryTimeUnit string              `json:"delivery_time_unit,omitempty"`
	Zones            []LocalDeliveryZone `json:"zones,omitempty"`
	LocationID       string              `json:"location_id,omitempty"`
}

// LocalDeliveryUpdateRequest contains the request body for updating a local delivery option.
type LocalDeliveryUpdateRequest struct {
	Name             *string              `json:"name,omitempty"`
	Description      *string              `json:"description,omitempty"`
	Active           *bool                `json:"active,omitempty"`
	Price            *string              `json:"price,omitempty"`
	FreeAbove        *string              `json:"free_above,omitempty"`
	MinOrderAmount   *string              `json:"min_order_amount,omitempty"`
	MaxOrderAmount   *string              `json:"max_order_amount,omitempty"`
	DeliveryTimeMin  *int                 `json:"delivery_time_min,omitempty"`
	DeliveryTimeMax  *int                 `json:"delivery_time_max,omitempty"`
	DeliveryTimeUnit *string              `json:"delivery_time_unit,omitempty"`
	Zones            *[]LocalDeliveryZone `json:"zones,omitempty"`
}

// ListLocalDeliveryOptions retrieves a list of local delivery options.
func (c *Client) ListLocalDeliveryOptions(ctx context.Context, opts *LocalDeliveryListOptions) (*LocalDeliveryListResponse, error) {
	path := "/local_delivery"
	if opts != nil {
		q := NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("location_id", opts.LocationID).
			BoolPtr("active", opts.Active)
		path += q.Build()
	}

	var resp LocalDeliveryListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetLocalDeliveryOption retrieves a single local delivery option by ID.
func (c *Client) GetLocalDeliveryOption(ctx context.Context, id string) (*LocalDeliveryOption, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("local delivery option id is required")
	}
	var option LocalDeliveryOption
	if err := c.Get(ctx, fmt.Sprintf("/local_delivery/%s", id), &option); err != nil {
		return nil, err
	}
	return &option, nil
}

// CreateLocalDeliveryOption creates a new local delivery option.
func (c *Client) CreateLocalDeliveryOption(ctx context.Context, req *LocalDeliveryCreateRequest) (*LocalDeliveryOption, error) {
	var option LocalDeliveryOption
	if err := c.Post(ctx, "/local_delivery", req, &option); err != nil {
		return nil, err
	}
	return &option, nil
}

// UpdateLocalDeliveryOption updates an existing local delivery option.
func (c *Client) UpdateLocalDeliveryOption(ctx context.Context, id string, req *LocalDeliveryUpdateRequest) (*LocalDeliveryOption, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("local delivery option id is required")
	}
	var option LocalDeliveryOption
	if err := c.Put(ctx, fmt.Sprintf("/local_delivery/%s", id), req, &option); err != nil {
		return nil, err
	}
	return &option, nil
}

// DeleteLocalDeliveryOption deletes a local delivery option.
func (c *Client) DeleteLocalDeliveryOption(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("local delivery option id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/local_delivery/%s", id))
}
