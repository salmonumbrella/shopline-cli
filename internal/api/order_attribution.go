package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// OrderAttribution represents attribution data for an order.
type OrderAttribution struct {
	ID              string     `json:"id"`
	OrderID         string     `json:"order_id"`
	Source          string     `json:"source"`
	Medium          string     `json:"medium"`
	Campaign        string     `json:"campaign"`
	Content         string     `json:"content"`
	Term            string     `json:"term"`
	ReferrerURL     string     `json:"referrer_url"`
	LandingPage     string     `json:"landing_page"`
	UtmSource       string     `json:"utm_source"`
	UtmMedium       string     `json:"utm_medium"`
	UtmCampaign     string     `json:"utm_campaign"`
	UtmContent      string     `json:"utm_content"`
	UtmTerm         string     `json:"utm_term"`
	FirstVisitAt    *time.Time `json:"first_visit_at"`
	LastVisitAt     *time.Time `json:"last_visit_at"`
	TouchpointCount int        `json:"touchpoint_count"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// OrderAttributionListOptions contains options for listing order attributions.
type OrderAttributionListOptions struct {
	Page     int
	PageSize int
	Source   string
	Medium   string
	Campaign string
	Since    *time.Time
	Until    *time.Time
}

// OrderAttributionListResponse is the paginated response for order attributions.
type OrderAttributionListResponse = ListResponse[OrderAttribution]

// ListOrderAttributions retrieves a list of order attributions.
func (c *Client) ListOrderAttributions(ctx context.Context, opts *OrderAttributionListOptions) (*OrderAttributionListResponse, error) {
	path := "/order_attributions"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("source", opts.Source).
			String("medium", opts.Medium).
			String("campaign", opts.Campaign).
			Time("created_at_min", opts.Since).
			Time("created_at_max", opts.Until).
			Build()
	}

	var resp OrderAttributionListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetOrderAttribution retrieves attribution data for a specific order.
func (c *Client) GetOrderAttribution(ctx context.Context, orderID string) (*OrderAttribution, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	var attribution OrderAttribution
	if err := c.Get(ctx, fmt.Sprintf("/orders/%s/attribution", orderID), &attribution); err != nil {
		return nil, err
	}
	return &attribution, nil
}
