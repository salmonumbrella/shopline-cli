package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// AffiliateCampaignOrdersOptions contains options for listing affiliate campaign orders.
type AffiliateCampaignOrdersOptions struct {
	Page     int
	PageSize int
}

// GetAffiliateCampaignOrders retrieves orders for a specific affiliate campaign (documented endpoint).
//
// Docs: GET /affiliate_campaigns/{id}/orders
func (c *Client) GetAffiliateCampaignOrders(ctx context.Context, id string, opts *AffiliateCampaignOrdersOptions) (json.RawMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("affiliate campaign id is required")
	}
	path := fmt.Sprintf("/affiliate_campaigns/%s/orders", id)
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

// GetAffiliateCampaignSummary retrieves summary for a specific affiliate campaign (documented endpoint).
//
// Docs: GET /affiliate_campaigns/{id}/summary
func (c *Client) GetAffiliateCampaignSummary(ctx context.Context, id string) (json.RawMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("affiliate campaign id is required")
	}

	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/affiliate_campaigns/%s/summary", id), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// AffiliateCampaignProductsSalesRankingOptions contains options for listing products sales ranking.
type AffiliateCampaignProductsSalesRankingOptions struct {
	Page     int
	PageSize int
}

// GetAffiliateCampaignProductsSalesRanking retrieves products sales ranking for a campaign (documented endpoint).
//
// Docs: GET /affiliate_campaigns/{id}/get_products_sales_ranking
func (c *Client) GetAffiliateCampaignProductsSalesRanking(ctx context.Context, id string, opts *AffiliateCampaignProductsSalesRankingOptions) (json.RawMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("affiliate campaign id is required")
	}
	path := fmt.Sprintf("/affiliate_campaigns/%s/get_products_sales_ranking", id)
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

// ExportAffiliateCampaignReport exports a campaign report (documented endpoint; raw JSON body).
//
// Docs: POST /affiliate_campaigns/{id}/export_report
func (c *Client) ExportAffiliateCampaignReport(ctx context.Context, id string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("affiliate campaign id is required")
	}

	var resp json.RawMessage
	if err := c.Post(ctx, fmt.Sprintf("/affiliate_campaigns/%s/export_report", id), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
