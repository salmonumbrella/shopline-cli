package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// FlashPrice represents a flash sale pricing configuration.
type FlashPrice struct {
	ID            string    `json:"id"`
	ProductID     string    `json:"product_id"`
	VariantID     string    `json:"variant_id"`
	OriginalPrice float64   `json:"original_price"`
	FlashPrice    float64   `json:"flash_price"`
	DiscountPct   float64   `json:"discount_pct"`
	Quantity      int       `json:"quantity"`
	QuantitySold  int       `json:"quantity_sold"`
	LimitPerUser  int       `json:"limit_per_user"`
	Status        string    `json:"status"`
	StartsAt      time.Time `json:"starts_at"`
	EndsAt        time.Time `json:"ends_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// FlashPriceListOptions contains options for listing flash prices.
type FlashPriceListOptions struct {
	Page      int
	PageSize  int
	ProductID string
	Status    string
}

// FlashPriceListResponse is the paginated response for flash prices.
type FlashPriceListResponse = ListResponse[FlashPrice]

// FlashPriceCreateRequest contains the request body for creating a flash price.
type FlashPriceCreateRequest struct {
	ProductID    string     `json:"product_id"`
	VariantID    string     `json:"variant_id,omitempty"`
	FlashPrice   float64    `json:"flash_price"`
	Quantity     int        `json:"quantity,omitempty"`
	LimitPerUser int        `json:"limit_per_user,omitempty"`
	StartsAt     *time.Time `json:"starts_at"`
	EndsAt       *time.Time `json:"ends_at"`
}

// FlashPriceUpdateRequest contains the request body for updating a flash price.
type FlashPriceUpdateRequest struct {
	FlashPrice   *float64   `json:"flash_price,omitempty"`
	Quantity     *int       `json:"quantity,omitempty"`
	LimitPerUser *int       `json:"limit_per_user,omitempty"`
	StartsAt     *time.Time `json:"starts_at,omitempty"`
	EndsAt       *time.Time `json:"ends_at,omitempty"`
}

// ListFlashPrices retrieves a list of flash price campaigns.
func (c *Client) ListFlashPrices(ctx context.Context, opts *FlashPriceListOptions) (*FlashPriceListResponse, error) {
	path := "/flash-price-campaigns" + NewQuery().
		Int("page", opts.Page).
		Int("page_size", opts.PageSize).
		String("product_id", opts.ProductID).
		String("status", opts.Status).
		Build()

	var resp FlashPriceListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetFlashPrice retrieves a single flash price campaign by ID.
func (c *Client) GetFlashPrice(ctx context.Context, id string) (*FlashPrice, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("flash price campaign id is required")
	}
	var flashPrice FlashPrice
	if err := c.Get(ctx, fmt.Sprintf("/flash-price-campaigns/%s", id), &flashPrice); err != nil {
		return nil, err
	}
	return &flashPrice, nil
}

// CreateFlashPrice creates a new flash price campaign.
func (c *Client) CreateFlashPrice(ctx context.Context, req *FlashPriceCreateRequest) (*FlashPrice, error) {
	var flashPrice FlashPrice
	if err := c.Post(ctx, "/flash-price-campaigns", req, &flashPrice); err != nil {
		return nil, err
	}
	return &flashPrice, nil
}

// UpdateFlashPrice updates an existing flash price campaign.
func (c *Client) UpdateFlashPrice(ctx context.Context, id string, req *FlashPriceUpdateRequest) (*FlashPrice, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("flash price campaign id is required")
	}
	var flashPrice FlashPrice
	if err := c.Put(ctx, fmt.Sprintf("/flash-price-campaigns/%s", id), req, &flashPrice); err != nil {
		return nil, err
	}
	return &flashPrice, nil
}

// DeleteFlashPrice deletes a flash price campaign.
func (c *Client) DeleteFlashPrice(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("flash price campaign id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/flash-price-campaigns/%s", id))
}

// ActivateFlashPrice activates a flash price campaign.
func (c *Client) ActivateFlashPrice(ctx context.Context, id string) (*FlashPrice, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("flash price campaign id is required")
	}
	var flashPrice FlashPrice
	if err := c.Post(ctx, fmt.Sprintf("/flash-price-campaigns/%s/activate", id), nil, &flashPrice); err != nil {
		return nil, err
	}
	return &flashPrice, nil
}

// DeactivateFlashPrice deactivates a flash price campaign.
func (c *Client) DeactivateFlashPrice(ctx context.Context, id string) (*FlashPrice, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("flash price campaign id is required")
	}
	var flashPrice FlashPrice
	if err := c.Post(ctx, fmt.Sprintf("/flash-price-campaigns/%s/deactivate", id), nil, &flashPrice); err != nil {
		return nil, err
	}
	return &flashPrice, nil
}
