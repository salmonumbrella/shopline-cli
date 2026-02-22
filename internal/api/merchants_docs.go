package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// GetMerchantByID retrieves merchant info by merchant id (documented endpoint).
//
// Docs: GET /merchants/{merchant_id}
func (c *Client) GetMerchantByID(ctx context.Context, merchantID string) (json.RawMessage, error) {
	if strings.TrimSpace(merchantID) == "" {
		return nil, fmt.Errorf("merchant id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/merchants/%s", merchantID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GenerateMerchantExpressLink generates an express cart link for the merchant (documented endpoint).
//
// Docs: POST /merchants/generate_express_link
func (c *Client) GenerateMerchantExpressLink(ctx context.Context, body any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Post(ctx, "/merchants/generate_express_link", body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
