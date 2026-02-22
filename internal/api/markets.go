package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Market represents a Shopline market.
type Market struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Handle     string    `json:"handle"`
	Enabled    bool      `json:"enabled"`
	Primary    bool      `json:"primary"`
	Countries  []string  `json:"countries"`
	Currencies []string  `json:"currencies"`
	Languages  []string  `json:"languages"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// MarketsListOptions contains options for listing markets.
type MarketsListOptions struct {
	Page     int
	PageSize int
}

// MarketsListResponse contains the list response.
type MarketsListResponse struct {
	Items      []Market `json:"items"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	TotalCount int      `json:"total_count"`
	HasMore    bool     `json:"has_more"`
}

// MarketCreateRequest contains the request body for creating a market.
type MarketCreateRequest struct {
	Name       string   `json:"name"`
	Handle     string   `json:"handle,omitempty"`
	Enabled    bool     `json:"enabled,omitempty"`
	Countries  []string `json:"countries,omitempty"`
	Currencies []string `json:"currencies,omitempty"`
	Languages  []string `json:"languages,omitempty"`
}

// ListMarkets retrieves a list of markets.
func (c *Client) ListMarkets(ctx context.Context, opts *MarketsListOptions) (*MarketsListResponse, error) {
	path := "/markets"
	if opts != nil {
		params := url.Values{}
		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			params.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp MarketsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetMarket retrieves a single market by ID.
func (c *Client) GetMarket(ctx context.Context, id string) (*Market, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("market id is required")
	}
	var market Market
	if err := c.Get(ctx, fmt.Sprintf("/markets/%s", id), &market); err != nil {
		return nil, err
	}
	return &market, nil
}

// CreateMarket creates a new market.
func (c *Client) CreateMarket(ctx context.Context, req *MarketCreateRequest) (*Market, error) {
	var market Market
	if err := c.Post(ctx, "/markets", req, &market); err != nil {
		return nil, err
	}
	return &market, nil
}

// DeleteMarket deletes a market.
func (c *Client) DeleteMarket(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("market id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/markets/%s", id))
}
