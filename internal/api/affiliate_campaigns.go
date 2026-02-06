package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// AffiliateCampaign represents a Shopline affiliate marketing campaign.
type AffiliateCampaign struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Status          string    `json:"status"`          // active, paused, ended
	CommissionType  string    `json:"commission_type"` // percentage, fixed
	CommissionValue float64   `json:"commission_value"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	TotalClicks     int       `json:"total_clicks"`
	TotalSales      int       `json:"total_sales"`
	TotalRevenue    float64   `json:"total_revenue"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// AffiliateCampaignsListOptions contains options for listing affiliate campaigns.
type AffiliateCampaignsListOptions struct {
	Page     int
	PageSize int
	Status   string
}

// AffiliateCampaignsListResponse is the paginated response for affiliate campaigns.
type AffiliateCampaignsListResponse = ListResponse[AffiliateCampaign]

// AffiliateCampaignCreateRequest contains the data for creating an affiliate campaign.
type AffiliateCampaignCreateRequest struct {
	Name            string     `json:"name"`
	Description     string     `json:"description,omitempty"`
	CommissionType  string     `json:"commission_type"`
	CommissionValue float64    `json:"commission_value"`
	StartDate       *time.Time `json:"start_date,omitempty"`
	EndDate         *time.Time `json:"end_date,omitempty"`
}

// AffiliateCampaignUpdateRequest contains the data for updating an affiliate campaign.
type AffiliateCampaignUpdateRequest struct {
	Name            string     `json:"name,omitempty"`
	Description     string     `json:"description,omitempty"`
	Status          string     `json:"status,omitempty"`
	CommissionType  string     `json:"commission_type,omitempty"`
	CommissionValue float64    `json:"commission_value,omitempty"`
	EndDate         *time.Time `json:"end_date,omitempty"`
}

// ListAffiliateCampaigns retrieves a list of affiliate campaigns.
func (c *Client) ListAffiliateCampaigns(ctx context.Context, opts *AffiliateCampaignsListOptions) (*AffiliateCampaignsListResponse, error) {
	path := "/affiliate_campaigns"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("status", opts.Status).
			Build()
	}

	var resp AffiliateCampaignsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAffiliateCampaign retrieves a single affiliate campaign by ID.
func (c *Client) GetAffiliateCampaign(ctx context.Context, id string) (*AffiliateCampaign, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("affiliate campaign id is required")
	}
	var campaign AffiliateCampaign
	if err := c.Get(ctx, fmt.Sprintf("/affiliate_campaigns/%s", id), &campaign); err != nil {
		return nil, err
	}
	return &campaign, nil
}

// CreateAffiliateCampaign creates a new affiliate campaign.
func (c *Client) CreateAffiliateCampaign(ctx context.Context, req *AffiliateCampaignCreateRequest) (*AffiliateCampaign, error) {
	var campaign AffiliateCampaign
	if err := c.Post(ctx, "/affiliate_campaigns", req, &campaign); err != nil {
		return nil, err
	}
	return &campaign, nil
}

// UpdateAffiliateCampaign updates an existing affiliate campaign.
func (c *Client) UpdateAffiliateCampaign(ctx context.Context, id string, req *AffiliateCampaignUpdateRequest) (*AffiliateCampaign, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("affiliate campaign id is required")
	}
	var campaign AffiliateCampaign
	if err := c.Put(ctx, fmt.Sprintf("/affiliate_campaigns/%s", id), req, &campaign); err != nil {
		return nil, err
	}
	return &campaign, nil
}

// DeleteAffiliateCampaign deletes an affiliate campaign.
func (c *Client) DeleteAffiliateCampaign(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("affiliate campaign id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/affiliate_campaigns/%s", id))
}
