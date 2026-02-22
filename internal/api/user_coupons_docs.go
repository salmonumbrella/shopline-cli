package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// User coupons (documented endpoints that differ from /user_coupons CRUD).

type UserCouponsListEndpointOptions struct {
	Page     int
	PageSize int
}

// ListUserCouponsListEndpoint lists user coupons via the documented /user_coupons/list endpoint.
//
// Docs: GET /user_coupons/list
func (c *Client) ListUserCouponsListEndpoint(ctx context.Context, opts *UserCouponsListEndpointOptions) (json.RawMessage, error) {
	path := "/user_coupons/list"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			Build()
	}
	var resp json.RawMessage
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// ClaimUserCoupon claims a coupon code for the current user/account context.
//
// Docs: POST /user_coupons/{coupon_code}/claim
func (c *Client) ClaimUserCoupon(ctx context.Context, couponCode string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(couponCode) == "" {
		return nil, fmt.Errorf("coupon code is required")
	}
	var resp json.RawMessage
	if err := c.Post(ctx, fmt.Sprintf("/user_coupons/%s/claim", couponCode), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// RedeemUserCoupon redeems a coupon code.
//
// Docs: POST /user_coupons/{coupon_code}/redeem
func (c *Client) RedeemUserCoupon(ctx context.Context, couponCode string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(couponCode) == "" {
		return nil, fmt.Errorf("coupon code is required")
	}
	var resp json.RawMessage
	if err := c.Post(ctx, fmt.Sprintf("/user_coupons/%s/redeem", couponCode), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
