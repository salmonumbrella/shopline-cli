package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// FlashPriceCampaignsListOptions contains options for listing flash price campaigns.
type FlashPriceCampaignsListOptions struct {
	Page     int
	PageSize int
}

// ListFlashPriceCampaigns retrieves flash price campaigns (documented endpoint).
//
// Docs: GET /flash_price_campaigns
func (c *Client) ListFlashPriceCampaigns(ctx context.Context, opts *FlashPriceCampaignsListOptions) (json.RawMessage, error) {
	q := NewQuery()
	if opts != nil {
		q = q.Int("page", opts.Page).Int("page_size", opts.PageSize)
	}
	path := "/flash_price_campaigns" + q.Build()

	var resp json.RawMessage
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetFlashPriceCampaign retrieves a flash price campaign (documented endpoint).
//
// Docs: GET /flash_price_campaigns/{id}
func (c *Client) GetFlashPriceCampaign(ctx context.Context, id string) (json.RawMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("flash price campaign id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/flash_price_campaigns/%s", id), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateFlashPriceCampaign creates a flash price campaign (documented endpoint).
//
// Docs: POST /flash_price_campaigns
func (c *Client) CreateFlashPriceCampaign(ctx context.Context, body any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Post(ctx, "/flash_price_campaigns", body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateFlashPriceCampaign updates a flash price campaign (documented endpoint).
//
// Docs: PUT /flash_price_campaigns/{id}
func (c *Client) UpdateFlashPriceCampaign(ctx context.Context, id string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("flash price campaign id is required")
	}
	var resp json.RawMessage
	if err := c.Put(ctx, fmt.Sprintf("/flash_price_campaigns/%s", id), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// DeleteFlashPriceCampaign deletes a flash price campaign (documented endpoint).
//
// Docs: DELETE /flash_price_campaigns/{id}
func (c *Client) DeleteFlashPriceCampaign(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("flash price campaign id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/flash_price_campaigns/%s", id))
}
