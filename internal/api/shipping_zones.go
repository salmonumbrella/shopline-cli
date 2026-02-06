package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ShippingZone represents a Shopline shipping zone.
type ShippingZone struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	Countries        []ZoneCountry  `json:"countries"`
	PriceBasedRates  []ShippingRate `json:"price_based_shipping_rates"`
	WeightBasedRates []ShippingRate `json:"weight_based_shipping_rates"`
	CarrierServices  []string       `json:"carrier_shipping_rate_provider_services"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// ZoneCountry represents a country in a shipping zone.
type ZoneCountry struct {
	Code      string   `json:"code"`
	Name      string   `json:"name"`
	Provinces []string `json:"provinces"`
}

// ShippingRate represents a shipping rate.
type ShippingRate struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Price     string  `json:"price"`
	MinValue  string  `json:"min_order_subtotal,omitempty"`
	MaxValue  string  `json:"max_order_subtotal,omitempty"`
	MinWeight float64 `json:"min_weight,omitempty"`
	MaxWeight float64 `json:"max_weight,omitempty"`
}

// ShippingZonesListOptions contains options for listing shipping zones.
type ShippingZonesListOptions struct {
	Page     int
	PageSize int
}

// ShippingZonesListResponse contains the list response.
type ShippingZonesListResponse struct {
	Items      []ShippingZone `json:"items"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalCount int            `json:"total_count"`
	HasMore    bool           `json:"has_more"`
}

// ShippingZoneCreateRequest contains the request body for creating a shipping zone.
type ShippingZoneCreateRequest struct {
	Name      string        `json:"name"`
	Countries []ZoneCountry `json:"countries,omitempty"`
}

// ListShippingZones retrieves a list of shipping zones.
func (c *Client) ListShippingZones(ctx context.Context, opts *ShippingZonesListOptions) (*ShippingZonesListResponse, error) {
	path := "/shipping_zones"
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

	var resp ShippingZonesListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetShippingZone retrieves a single shipping zone by ID.
func (c *Client) GetShippingZone(ctx context.Context, id string) (*ShippingZone, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("shipping zone id is required")
	}
	var zone ShippingZone
	if err := c.Get(ctx, fmt.Sprintf("/shipping_zones/%s", id), &zone); err != nil {
		return nil, err
	}
	return &zone, nil
}

// CreateShippingZone creates a new shipping zone.
func (c *Client) CreateShippingZone(ctx context.Context, req *ShippingZoneCreateRequest) (*ShippingZone, error) {
	var zone ShippingZone
	if err := c.Post(ctx, "/shipping_zones", req, &zone); err != nil {
		return nil, err
	}
	return &zone, nil
}

// DeleteShippingZone deletes a shipping zone.
func (c *Client) DeleteShippingZone(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("shipping zone id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/shipping_zones/%s", id))
}
