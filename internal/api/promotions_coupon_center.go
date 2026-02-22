package api

import (
	"context"
	"encoding/json"
)

// GetPromotionsCouponCenter retrieves coupon center promotions (documented endpoint).
//
// Docs: GET /promotions/coupon-center
func (c *Client) GetPromotionsCouponCenter(ctx context.Context) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/promotions/coupon-center", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
