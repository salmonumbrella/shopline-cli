package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Promotion represents a Shopline promotion.
type Promotion struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Type          string    `json:"type"`
	Status        string    `json:"status"`
	DiscountType  string    `json:"discount_type"`
	DiscountValue float64   `json:"discount_value"`
	MinPurchase   float64   `json:"min_purchase"`
	UsageLimit    int       `json:"usage_limit"`
	UsageCount    int       `json:"usage_count"`
	StartsAt      time.Time `json:"starts_at"`
	EndsAt        time.Time `json:"ends_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PromotionsListOptions contains options for listing promotions.
type PromotionsListOptions struct {
	Page     int
	PageSize int
	Status   string
	Type     string
}

// PromotionsListResponse is the paginated response for promotions.
type PromotionsListResponse = ListResponse[Promotion]

// PromotionCreateRequest contains the request body for creating a promotion.
type PromotionCreateRequest struct {
	Title         string    `json:"title"`
	Description   string    `json:"description,omitempty"`
	Type          string    `json:"type"`
	DiscountType  string    `json:"discount_type"`
	DiscountValue float64   `json:"discount_value"`
	MinPurchase   float64   `json:"min_purchase,omitempty"`
	UsageLimit    int       `json:"usage_limit,omitempty"`
	StartsAt      time.Time `json:"starts_at"`
	EndsAt        time.Time `json:"ends_at,omitempty"`
}

// PromotionUpdateRequest contains the request body for updating a promotion.
type PromotionUpdateRequest struct {
	Title         *string    `json:"title,omitempty"`
	Description   *string    `json:"description,omitempty"`
	Type          *string    `json:"type,omitempty"`
	DiscountType  *string    `json:"discount_type,omitempty"`
	DiscountValue *float64   `json:"discount_value,omitempty"`
	MinPurchase   *float64   `json:"min_purchase,omitempty"`
	UsageLimit    *int       `json:"usage_limit,omitempty"`
	StartsAt      *time.Time `json:"starts_at,omitempty"`
	EndsAt        *time.Time `json:"ends_at,omitempty"`
}

// ListPromotions retrieves a list of promotions.
func (c *Client) ListPromotions(ctx context.Context, opts *PromotionsListOptions) (*PromotionsListResponse, error) {
	path := "/promotions"
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
		if opts.Type != "" {
			params.Set("type", opts.Type)
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp PromotionsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPromotion retrieves a single promotion by ID.
func (c *Client) GetPromotion(ctx context.Context, id string) (*Promotion, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("promotion id is required")
	}
	var promotion Promotion
	if err := c.Get(ctx, fmt.Sprintf("/promotions/%s", id), &promotion); err != nil {
		return nil, err
	}
	return &promotion, nil
}

// CreatePromotion creates a new promotion.
func (c *Client) CreatePromotion(ctx context.Context, req *PromotionCreateRequest) (*Promotion, error) {
	var promotion Promotion
	if err := c.Post(ctx, "/promotions", req, &promotion); err != nil {
		return nil, err
	}
	return &promotion, nil
}

// UpdatePromotion updates an existing promotion.
func (c *Client) UpdatePromotion(ctx context.Context, id string, req *PromotionUpdateRequest) (*Promotion, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("promotion id is required")
	}
	var promotion Promotion
	if err := c.Put(ctx, fmt.Sprintf("/promotions/%s", id), req, &promotion); err != nil {
		return nil, err
	}
	return &promotion, nil
}

// DeletePromotion deletes a promotion.
func (c *Client) DeletePromotion(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("promotion id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/promotions/%s", id))
}

// ActivatePromotion activates a promotion.
func (c *Client) ActivatePromotion(ctx context.Context, id string) (*Promotion, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("promotion id is required")
	}
	var promotion Promotion
	if err := c.Post(ctx, fmt.Sprintf("/promotions/%s/activate", id), nil, &promotion); err != nil {
		return nil, err
	}
	return &promotion, nil
}

// DeactivatePromotion deactivates a promotion.
func (c *Client) DeactivatePromotion(ctx context.Context, id string) (*Promotion, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("promotion id is required")
	}
	var promotion Promotion
	if err := c.Post(ctx, fmt.Sprintf("/promotions/%s/deactivate", id), nil, &promotion); err != nil {
		return nil, err
	}
	return &promotion, nil
}

// PromotionSearchOptions contains options for searching promotions.
type PromotionSearchOptions struct {
	Query    string
	Status   string
	Type     string
	Page     int
	PageSize int
}

// CouponSendRequest contains the request body for sending a coupon.
type CouponSendRequest struct {
	PromotionID string   `json:"promotion_id"`
	CustomerIDs []string `json:"customer_ids"`
}

// CouponRedeemRequest contains the request body for redeeming a coupon.
type CouponRedeemRequest struct {
	Code       string `json:"code"`
	CustomerID string `json:"customer_id"`
	OrderID    string `json:"order_id,omitempty"`
}

// CouponClaimRequest contains the request body for claiming a coupon.
type CouponClaimRequest struct {
	Code       string `json:"code"`
	CustomerID string `json:"customer_id"`
}

// SearchPromotions searches for promotions with query parameters.
func (c *Client) SearchPromotions(ctx context.Context, opts *PromotionSearchOptions) (*PromotionsListResponse, error) {
	path := "/promotions/search" + NewQuery().
		String("query", opts.Query).
		String("status", opts.Status).
		String("type", opts.Type).
		Int("page", opts.Page).
		Int("page_size", opts.PageSize).
		Build()

	var resp PromotionsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendCoupon sends a coupon to customers.
func (c *Client) SendCoupon(ctx context.Context, promotionID string, customerIDs []string) error {
	if strings.TrimSpace(promotionID) == "" {
		return fmt.Errorf("promotion id is required")
	}
	if len(customerIDs) == 0 {
		return fmt.Errorf("at least one customer id is required")
	}
	req := &CouponSendRequest{PromotionID: promotionID, CustomerIDs: customerIDs}
	return c.Post(ctx, "/promotions/send-coupon", req, nil)
}

// RedeemCoupon redeems a coupon for a customer.
func (c *Client) RedeemCoupon(ctx context.Context, code, customerID string, orderID string) error {
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("coupon code is required")
	}
	if strings.TrimSpace(customerID) == "" {
		return fmt.Errorf("customer id is required")
	}
	req := &CouponRedeemRequest{Code: code, CustomerID: customerID, OrderID: orderID}
	return c.Post(ctx, "/promotions/redeem-coupon", req, nil)
}

// ClaimCoupon claims a coupon for a customer.
func (c *Client) ClaimCoupon(ctx context.Context, code, customerID string) error {
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("coupon code is required")
	}
	if strings.TrimSpace(customerID) == "" {
		return fmt.Errorf("customer id is required")
	}
	req := &CouponClaimRequest{Code: code, CustomerID: customerID}
	return c.Post(ctx, "/promotions/claim-coupon", req, nil)
}
