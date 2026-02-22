package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// DiscountCode represents a Shopline discount code.
type DiscountCode struct {
	ID            string    `json:"id"`
	Code          string    `json:"code"`
	PriceRuleID   string    `json:"price_rule_id"`
	DiscountType  string    `json:"discount_type"`
	DiscountValue float64   `json:"discount_value"`
	UsageLimit    int       `json:"usage_limit"`
	UsageCount    int       `json:"usage_count"`
	MinPurchase   float64   `json:"min_purchase"`
	StartsAt      time.Time `json:"starts_at"`
	EndsAt        time.Time `json:"ends_at"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// DiscountCodesListOptions contains options for listing discount codes.
type DiscountCodesListOptions struct {
	Page        int
	PageSize    int
	PriceRuleID string
	Status      string
}

// DiscountCodesListResponse is the paginated response for discount codes.
type DiscountCodesListResponse = ListResponse[DiscountCode]

// DiscountCodeCreateRequest contains the request body for creating a discount code.
type DiscountCodeCreateRequest struct {
	Code          string    `json:"code"`
	PriceRuleID   string    `json:"price_rule_id,omitempty"`
	DiscountType  string    `json:"discount_type"`
	DiscountValue float64   `json:"discount_value"`
	UsageLimit    int       `json:"usage_limit,omitempty"`
	MinPurchase   float64   `json:"min_purchase,omitempty"`
	StartsAt      time.Time `json:"starts_at,omitempty"`
	EndsAt        time.Time `json:"ends_at,omitempty"`
}

// ListDiscountCodes retrieves a list of discount codes.
func (c *Client) ListDiscountCodes(ctx context.Context, opts *DiscountCodesListOptions) (*DiscountCodesListResponse, error) {
	path := "/discount_codes"
	if opts != nil {
		params := url.Values{}
		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			params.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.PriceRuleID != "" {
			params.Set("price_rule_id", opts.PriceRuleID)
		}
		if opts.Status != "" {
			params.Set("status", opts.Status)
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp DiscountCodesListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetDiscountCode retrieves a single discount code by ID.
func (c *Client) GetDiscountCode(ctx context.Context, id string) (*DiscountCode, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("discount code id is required")
	}
	var code DiscountCode
	if err := c.Get(ctx, fmt.Sprintf("/discount_codes/%s", id), &code); err != nil {
		return nil, err
	}
	return &code, nil
}

// GetDiscountCodeByCode retrieves a discount code by its code string.
func (c *Client) GetDiscountCodeByCode(ctx context.Context, code string) (*DiscountCode, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("discount code is required")
	}
	var discountCode DiscountCode
	path := fmt.Sprintf("/discount_codes/lookup?code=%s", url.QueryEscape(code))
	if err := c.Get(ctx, path, &discountCode); err != nil {
		return nil, err
	}
	return &discountCode, nil
}

// CreateDiscountCode creates a new discount code.
func (c *Client) CreateDiscountCode(ctx context.Context, req *DiscountCodeCreateRequest) (*DiscountCode, error) {
	var code DiscountCode
	if err := c.Post(ctx, "/discount_codes", req, &code); err != nil {
		return nil, err
	}
	return &code, nil
}

// DeleteDiscountCode deletes a discount code.
func (c *Client) DeleteDiscountCode(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("discount code id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/discount_codes/%s", id))
}
