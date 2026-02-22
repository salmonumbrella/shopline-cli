package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Gift represents a gift promotion.
type Gift struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	GiftProductID   string    `json:"gift_product_id"`
	GiftVariantID   string    `json:"gift_variant_id"`
	GiftProductName string    `json:"gift_product_name"`
	TriggerType     string    `json:"trigger_type"`
	TriggerValue    float64   `json:"trigger_value"`
	Quantity        int       `json:"quantity"`
	QuantityUsed    int       `json:"quantity_used"`
	LimitPerUser    int       `json:"limit_per_user"`
	Status          string    `json:"status"`
	StartsAt        time.Time `json:"starts_at"`
	EndsAt          time.Time `json:"ends_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// GiftsListOptions contains options for listing gifts.
type GiftsListOptions struct {
	Page     int
	PageSize int
	Status   string
}

// GiftsListResponse is the paginated response for gifts.
type GiftsListResponse = ListResponse[Gift]

// GiftCreateRequest contains the request body for creating a gift.
type GiftCreateRequest struct {
	Title         string     `json:"title"`
	Description   string     `json:"description,omitempty"`
	GiftProductID string     `json:"gift_product_id"`
	GiftVariantID string     `json:"gift_variant_id,omitempty"`
	TriggerType   string     `json:"trigger_type"`
	TriggerValue  float64    `json:"trigger_value"`
	Quantity      int        `json:"quantity,omitempty"`
	LimitPerUser  int        `json:"limit_per_user,omitempty"`
	StartsAt      *time.Time `json:"starts_at"`
	EndsAt        *time.Time `json:"ends_at,omitempty"`
}

// GiftUpdateRequest contains the request body for updating a gift.
type GiftUpdateRequest struct {
	Title         string     `json:"title,omitempty"`
	Description   string     `json:"description,omitempty"`
	GiftProductID string     `json:"gift_product_id,omitempty"`
	GiftVariantID string     `json:"gift_variant_id,omitempty"`
	TriggerType   string     `json:"trigger_type,omitempty"`
	TriggerValue  *float64   `json:"trigger_value,omitempty"`
	Quantity      *int       `json:"quantity,omitempty"`
	LimitPerUser  *int       `json:"limit_per_user,omitempty"`
	StartsAt      *time.Time `json:"starts_at,omitempty"`
	EndsAt        *time.Time `json:"ends_at,omitempty"`
}

// GiftSearchOptions contains options for searching gifts.
type GiftSearchOptions struct {
	Query    string
	Status   string
	Page     int
	PageSize int
}

// GiftQuantityUpdateRequest contains the request body for updating gift quantity.
type GiftQuantityUpdateRequest struct {
	Quantity int `json:"quantity"`
}

// GiftQuantityBySKURequest contains the request body for bulk updating gift quantity by SKU.
type GiftQuantityBySKURequest struct {
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
}

// ListGifts retrieves a list of gift promotions.
func (c *Client) ListGifts(ctx context.Context, opts *GiftsListOptions) (*GiftsListResponse, error) {
	path := "/gifts" + NewQuery().
		Int("page", opts.Page).
		Int("page_size", opts.PageSize).
		String("status", opts.Status).
		Build()

	var resp GiftsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetGift retrieves a single gift by ID.
func (c *Client) GetGift(ctx context.Context, id string) (*Gift, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("gift id is required")
	}
	var gift Gift
	if err := c.Get(ctx, fmt.Sprintf("/gifts/%s", id), &gift); err != nil {
		return nil, err
	}
	return &gift, nil
}

// CreateGift creates a new gift promotion.
func (c *Client) CreateGift(ctx context.Context, req *GiftCreateRequest) (*Gift, error) {
	var gift Gift
	if err := c.Post(ctx, "/gifts", req, &gift); err != nil {
		return nil, err
	}
	return &gift, nil
}

// DeleteGift deletes a gift promotion.
func (c *Client) DeleteGift(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("gift id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/gifts/%s", id))
}

// ActivateGift activates a gift promotion.
func (c *Client) ActivateGift(ctx context.Context, id string) (*Gift, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("gift id is required")
	}
	var gift Gift
	if err := c.Post(ctx, fmt.Sprintf("/gifts/%s/activate", id), nil, &gift); err != nil {
		return nil, err
	}
	return &gift, nil
}

// DeactivateGift deactivates a gift promotion.
func (c *Client) DeactivateGift(ctx context.Context, id string) (*Gift, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("gift id is required")
	}
	var gift Gift
	if err := c.Post(ctx, fmt.Sprintf("/gifts/%s/deactivate", id), nil, &gift); err != nil {
		return nil, err
	}
	return &gift, nil
}

// UpdateGift updates an existing gift promotion.
func (c *Client) UpdateGift(ctx context.Context, id string, req *GiftUpdateRequest) (*Gift, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("gift id is required")
	}
	var gift Gift
	if err := c.Put(ctx, fmt.Sprintf("/gifts/%s", id), req, &gift); err != nil {
		return nil, err
	}
	return &gift, nil
}

// SearchGifts searches for gifts with query parameters.
func (c *Client) SearchGifts(ctx context.Context, opts *GiftSearchOptions) (*GiftsListResponse, error) {
	path := "/gifts/search" + NewQuery().
		String("query", opts.Query).
		String("status", opts.Status).
		Int("page", opts.Page).
		Int("page_size", opts.PageSize).
		Build()

	var resp GiftsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateGiftQuantity updates the quantity of a gift promotion.
func (c *Client) UpdateGiftQuantity(ctx context.Context, id string, quantity int) (*Gift, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("gift id is required")
	}
	req := &GiftQuantityUpdateRequest{Quantity: quantity}
	var gift Gift
	if err := c.Put(ctx, fmt.Sprintf("/gifts/%s/update_quantity", id), req, &gift); err != nil {
		return nil, err
	}
	return &gift, nil
}

// UpdateGiftsQuantityBySKU bulk updates gift quantities by SKU.
func (c *Client) UpdateGiftsQuantityBySKU(ctx context.Context, sku string, quantity int) error {
	if strings.TrimSpace(sku) == "" {
		return fmt.Errorf("sku is required")
	}
	req := &GiftQuantityBySKURequest{SKU: sku, Quantity: quantity}
	return c.Put(ctx, "/gifts/update_quantity", req, nil)
}

// GetGiftStocks retrieves stock info for a gift (documented endpoint).
//
// Docs: GET /gifts/{id}/stocks
func (c *Client) GetGiftStocks(ctx context.Context, id string) (json.RawMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("gift id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/gifts/%s/stocks", id), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateGiftStocks updates stock info for a gift (documented endpoint; raw JSON body).
//
// Docs: PUT /gifts/{id}/stocks
func (c *Client) UpdateGiftStocks(ctx context.Context, id string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("gift id is required")
	}
	var resp json.RawMessage
	if err := c.Put(ctx, fmt.Sprintf("/gifts/%s/stocks", id), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
