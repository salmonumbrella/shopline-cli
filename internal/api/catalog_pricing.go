package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CatalogPricing represents B2B catalog-specific pricing.
type CatalogPricing struct {
	ID            string    `json:"id"`
	CatalogID     string    `json:"catalog_id"`
	ProductID     string    `json:"product_id"`
	VariantID     string    `json:"variant_id"`
	OriginalPrice float64   `json:"original_price"`
	CatalogPrice  float64   `json:"catalog_price"`
	DiscountPct   float64   `json:"discount_pct"`
	MinQuantity   int       `json:"min_quantity"`
	MaxQuantity   int       `json:"max_quantity"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CatalogPricingListOptions contains options for listing catalog pricing.
type CatalogPricingListOptions struct {
	Page      int
	PageSize  int
	CatalogID string
	ProductID string
}

// CatalogPricingListResponse is the paginated response for catalog pricing.
type CatalogPricingListResponse = ListResponse[CatalogPricing]

// CatalogPricingCreateRequest contains the request body for creating catalog pricing.
type CatalogPricingCreateRequest struct {
	CatalogID    string  `json:"catalog_id"`
	ProductID    string  `json:"product_id"`
	VariantID    string  `json:"variant_id,omitempty"`
	CatalogPrice float64 `json:"catalog_price"`
	MinQuantity  int     `json:"min_quantity,omitempty"`
	MaxQuantity  int     `json:"max_quantity,omitempty"`
}

// CatalogPricingUpdateRequest contains the request body for updating catalog pricing.
type CatalogPricingUpdateRequest struct {
	CatalogPrice *float64 `json:"catalog_price,omitempty"`
	MinQuantity  *int     `json:"min_quantity,omitempty"`
	MaxQuantity  *int     `json:"max_quantity,omitempty"`
}

// ListCatalogPricing retrieves a list of catalog pricing entries.
func (c *Client) ListCatalogPricing(ctx context.Context, opts *CatalogPricingListOptions) (*CatalogPricingListResponse, error) {
	path := "/catalog_pricing" + NewQuery().
		Int("page", opts.Page).
		Int("page_size", opts.PageSize).
		String("catalog_id", opts.CatalogID).
		String("product_id", opts.ProductID).
		Build()

	var resp CatalogPricingListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCatalogPricing retrieves a single catalog pricing entry by ID.
func (c *Client) GetCatalogPricing(ctx context.Context, id string) (*CatalogPricing, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("catalog pricing id is required")
	}
	var pricing CatalogPricing
	if err := c.Get(ctx, fmt.Sprintf("/catalog_pricing/%s", id), &pricing); err != nil {
		return nil, err
	}
	return &pricing, nil
}

// CreateCatalogPricing creates a new catalog pricing entry.
func (c *Client) CreateCatalogPricing(ctx context.Context, req *CatalogPricingCreateRequest) (*CatalogPricing, error) {
	var pricing CatalogPricing
	if err := c.Post(ctx, "/catalog_pricing", req, &pricing); err != nil {
		return nil, err
	}
	return &pricing, nil
}

// UpdateCatalogPricing updates an existing catalog pricing entry.
func (c *Client) UpdateCatalogPricing(ctx context.Context, id string, req *CatalogPricingUpdateRequest) (*CatalogPricing, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("catalog pricing id is required")
	}
	var pricing CatalogPricing
	if err := c.Put(ctx, fmt.Sprintf("/catalog_pricing/%s", id), req, &pricing); err != nil {
		return nil, err
	}
	return &pricing, nil
}

// DeleteCatalogPricing deletes a catalog pricing entry.
func (c *Client) DeleteCatalogPricing(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("catalog pricing id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/catalog_pricing/%s", id))
}
