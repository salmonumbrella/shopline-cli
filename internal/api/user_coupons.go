package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// UserCoupon represents a coupon assigned to a user.
type UserCoupon struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	CouponID      string    `json:"coupon_id"`
	CouponCode    string    `json:"coupon_code"`
	Title         string    `json:"title"`
	DiscountType  string    `json:"discount_type"`
	DiscountValue float64   `json:"discount_value"`
	Status        string    `json:"status"`
	UsedAt        time.Time `json:"used_at,omitempty"`
	ExpiresAt     time.Time `json:"expires_at"`
	CreatedAt     time.Time `json:"created_at"`
}

// UserCouponsListOptions contains options for listing user coupons.
type UserCouponsListOptions struct {
	Page        int
	PageSize    int
	PromotionID string // Required: the promotion/coupon campaign ID
	UserID      string
	Status      string
}

// UserCouponsListResponse is the paginated response for user coupons.
type UserCouponsListResponse = ListResponse[UserCoupon]

// UserCouponAssignRequest contains the request body for assigning a coupon to a user.
type UserCouponAssignRequest struct {
	UserID   string `json:"user_id"`
	CouponID string `json:"coupon_id"`
}

// ListUserCoupons retrieves a list of user-assigned coupons.
// The PromotionID parameter is required by the Shopline API.
func (c *Client) ListUserCoupons(ctx context.Context, opts *UserCouponsListOptions) (*UserCouponsListResponse, error) {
	if opts == nil || opts.PromotionID == "" {
		return nil, fmt.Errorf("promotion_id is required")
	}
	path := "/user_coupons" + NewQuery().
		Int("page", opts.Page).
		Int("page_size", opts.PageSize).
		String("promotion_id", opts.PromotionID).
		String("user_id", opts.UserID).
		String("status", opts.Status).
		Build()

	var resp UserCouponsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetUserCoupon retrieves a single user coupon by ID.
func (c *Client) GetUserCoupon(ctx context.Context, id string) (*UserCoupon, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("user coupon id is required")
	}
	var userCoupon UserCoupon
	if err := c.Get(ctx, fmt.Sprintf("/user_coupons/%s", id), &userCoupon); err != nil {
		return nil, err
	}
	return &userCoupon, nil
}

// AssignUserCoupon assigns a coupon to a user.
func (c *Client) AssignUserCoupon(ctx context.Context, req *UserCouponAssignRequest) (*UserCoupon, error) {
	var userCoupon UserCoupon
	if err := c.Post(ctx, "/user_coupons", req, &userCoupon); err != nil {
		return nil, err
	}
	return &userCoupon, nil
}

// RevokeUserCoupon revokes a user's coupon.
func (c *Client) RevokeUserCoupon(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("user coupon id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/user_coupons/%s", id))
}
