package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// FulfillmentService represents a Shopline fulfillment service.
type FulfillmentService struct {
	ID                     string    `json:"id"`
	Name                   string    `json:"name"`
	Handle                 string    `json:"handle"`
	CallbackURL            string    `json:"callback_url"`
	InventoryManagement    bool      `json:"inventory_management"`
	TrackingSupport        bool      `json:"tracking_support"`
	RequiresShippingMethod bool      `json:"requires_shipping_method"`
	Format                 string    `json:"format"`
	IncludePendingStock    bool      `json:"include_pending_stock"`
	ServiceDiscovery       bool      `json:"service_discovery"`
	FulfillmentOrdersOptIn bool      `json:"fulfillment_orders_opt_in"`
	PermitsSkuSharing      bool      `json:"permits_sku_sharing"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// FulfillmentServicesListOptions contains options for listing fulfillment services.
type FulfillmentServicesListOptions struct {
	Page     int
	PageSize int
}

// FulfillmentServicesListResponse contains the list response.
type FulfillmentServicesListResponse struct {
	Items      []FulfillmentService `json:"items"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalCount int                  `json:"total_count"`
	HasMore    bool                 `json:"has_more"`
}

// FulfillmentServiceCreateRequest contains the request body for creating a fulfillment service.
type FulfillmentServiceCreateRequest struct {
	Name                   string `json:"name"`
	CallbackURL            string `json:"callback_url"`
	InventoryManagement    bool   `json:"inventory_management,omitempty"`
	TrackingSupport        bool   `json:"tracking_support,omitempty"`
	RequiresShippingMethod bool   `json:"requires_shipping_method,omitempty"`
	Format                 string `json:"format,omitempty"`
}

// FulfillmentServiceUpdateRequest contains the request body for updating a fulfillment service.
type FulfillmentServiceUpdateRequest struct {
	Name                   string `json:"name,omitempty"`
	CallbackURL            string `json:"callback_url,omitempty"`
	InventoryManagement    *bool  `json:"inventory_management,omitempty"`
	TrackingSupport        *bool  `json:"tracking_support,omitempty"`
	RequiresShippingMethod *bool  `json:"requires_shipping_method,omitempty"`
}

// ListFulfillmentServices retrieves a list of fulfillment services.
func (c *Client) ListFulfillmentServices(ctx context.Context, opts *FulfillmentServicesListOptions) (*FulfillmentServicesListResponse, error) {
	path := "/fulfillment_services"
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

	var resp FulfillmentServicesListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetFulfillmentService retrieves a single fulfillment service by ID.
func (c *Client) GetFulfillmentService(ctx context.Context, id string) (*FulfillmentService, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("fulfillment service id is required")
	}
	var fs FulfillmentService
	if err := c.Get(ctx, fmt.Sprintf("/fulfillment_services/%s", id), &fs); err != nil {
		return nil, err
	}
	return &fs, nil
}

// CreateFulfillmentService creates a new fulfillment service.
func (c *Client) CreateFulfillmentService(ctx context.Context, req *FulfillmentServiceCreateRequest) (*FulfillmentService, error) {
	var fs FulfillmentService
	if err := c.Post(ctx, "/fulfillment_services", req, &fs); err != nil {
		return nil, err
	}
	return &fs, nil
}

// UpdateFulfillmentService updates an existing fulfillment service.
func (c *Client) UpdateFulfillmentService(ctx context.Context, id string, req *FulfillmentServiceUpdateRequest) (*FulfillmentService, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("fulfillment service id is required")
	}
	var fs FulfillmentService
	if err := c.Put(ctx, fmt.Sprintf("/fulfillment_services/%s", id), req, &fs); err != nil {
		return nil, err
	}
	return &fs, nil
}

// DeleteFulfillmentService deletes a fulfillment service.
func (c *Client) DeleteFulfillmentService(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("fulfillment service id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/fulfillment_services/%s", id))
}
