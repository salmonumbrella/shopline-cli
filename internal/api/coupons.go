package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Coupon represents a Shopline coupon.
type Coupon struct {
	ID            string    `json:"id"`
	Code          string    `json:"code"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	DiscountType  string    `json:"discount_type"`
	DiscountValue float64   `json:"discount_value"`
	MinPurchase   float64   `json:"min_purchase"`
	MaxDiscount   float64   `json:"max_discount"`
	UsageLimit    int       `json:"usage_limit"`
	UsageCount    int       `json:"usage_count"`
	PerCustomer   int       `json:"per_customer"`
	Status        string    `json:"status"`
	StartsAt      time.Time `json:"starts_at"`
	EndsAt        time.Time `json:"ends_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CouponsListOptions contains options for listing coupons.
type CouponsListOptions struct {
	Page     int
	PageSize int
	Status   string
	Code     string
}

// CouponsListResponse contains the list response.
type CouponsListResponse struct {
	Items      []Coupon `json:"items"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	TotalCount int      `json:"total_count"`
	HasMore    bool     `json:"has_more"`
}

// CouponCreateRequest contains the request body for creating a coupon.
type CouponCreateRequest struct {
	Code          string     `json:"code"`
	Title         string     `json:"title"`
	Description   string     `json:"description,omitempty"`
	DiscountType  string     `json:"discount_type"`
	DiscountValue float64    `json:"discount_value"`
	MinPurchase   float64    `json:"min_purchase,omitempty"`
	MaxDiscount   float64    `json:"max_discount,omitempty"`
	UsageLimit    int        `json:"usage_limit,omitempty"`
	PerCustomer   int        `json:"per_customer,omitempty"`
	StartsAt      *time.Time `json:"starts_at,omitempty"`
	EndsAt        *time.Time `json:"ends_at,omitempty"`
}

// CouponUpdateRequest contains the request body for updating a coupon.
type CouponUpdateRequest struct {
	Title         string     `json:"title,omitempty"`
	Description   string     `json:"description,omitempty"`
	DiscountType  string     `json:"discount_type,omitempty"`
	DiscountValue *float64   `json:"discount_value,omitempty"`
	MinPurchase   *float64   `json:"min_purchase,omitempty"`
	MaxDiscount   *float64   `json:"max_discount,omitempty"`
	UsageLimit    *int       `json:"usage_limit,omitempty"`
	PerCustomer   *int       `json:"per_customer,omitempty"`
	StartsAt      *time.Time `json:"starts_at,omitempty"`
	EndsAt        *time.Time `json:"ends_at,omitempty"`
}

// ListCoupons retrieves a list of coupons.
func (c *Client) ListCoupons(ctx context.Context, opts *CouponsListOptions) (*CouponsListResponse, error) {
	path := "/coupons"
	if opts != nil {
		params := url.Values{}
		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			params.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.Status != "" {
			params.Set("status", opts.Status)
		}
		if opts.Code != "" {
			params.Set("code", opts.Code)
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp CouponsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCoupon retrieves a single coupon by ID.
func (c *Client) GetCoupon(ctx context.Context, id string) (*Coupon, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("coupon id is required")
	}
	var coupon Coupon
	if err := c.Get(ctx, fmt.Sprintf("/coupons/%s", id), &coupon); err != nil {
		return nil, err
	}
	return &coupon, nil
}

// GetCouponByCode retrieves a coupon by its code.
func (c *Client) GetCouponByCode(ctx context.Context, code string) (*Coupon, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("coupon code is required")
	}
	var coupon Coupon
	if err := c.Get(ctx, fmt.Sprintf("/coupons/code/%s", code), &coupon); err != nil {
		return nil, err
	}
	return &coupon, nil
}

// CreateCoupon creates a new coupon.
func (c *Client) CreateCoupon(ctx context.Context, req *CouponCreateRequest) (*Coupon, error) {
	var coupon Coupon
	if err := c.Post(ctx, "/coupons", req, &coupon); err != nil {
		return nil, err
	}
	return &coupon, nil
}

// UpdateCoupon updates an existing coupon.
func (c *Client) UpdateCoupon(ctx context.Context, id string, req *CouponUpdateRequest) (*Coupon, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("coupon id is required")
	}
	var coupon Coupon
	if err := c.Put(ctx, fmt.Sprintf("/coupons/%s", id), req, &coupon); err != nil {
		return nil, err
	}
	return &coupon, nil
}

// ActivateCoupon activates a coupon.
func (c *Client) ActivateCoupon(ctx context.Context, id string) (*Coupon, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("coupon id is required")
	}
	var coupon Coupon
	if err := c.Post(ctx, fmt.Sprintf("/coupons/%s/activate", id), nil, &coupon); err != nil {
		return nil, err
	}
	return &coupon, nil
}

// DeactivateCoupon deactivates a coupon.
func (c *Client) DeactivateCoupon(ctx context.Context, id string) (*Coupon, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("coupon id is required")
	}
	var coupon Coupon
	if err := c.Post(ctx, fmt.Sprintf("/coupons/%s/deactivate", id), nil, &coupon); err != nil {
		return nil, err
	}
	return &coupon, nil
}

// DeleteCoupon deletes a coupon.
func (c *Client) DeleteCoupon(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("coupon id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/coupons/%s", id))
}
