package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// CarrierService represents a Shopline carrier service.
type CarrierService struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	CallbackURL        string    `json:"callback_url"`
	Active             bool      `json:"active"`
	ServiceDiscovery   bool      `json:"service_discovery"`
	CarrierServiceType string    `json:"carrier_service_type"`
	Format             string    `json:"format"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// CarrierServicesListOptions contains options for listing carrier services.
type CarrierServicesListOptions struct {
	Page     int
	PageSize int
	Active   *bool
}

// CarrierServicesListResponse contains the list response.
type CarrierServicesListResponse struct {
	Items      []CarrierService `json:"items"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalCount int              `json:"total_count"`
	HasMore    bool             `json:"has_more"`
}

// CarrierServiceCreateRequest contains the request body for creating a carrier service.
type CarrierServiceCreateRequest struct {
	Name               string `json:"name"`
	CallbackURL        string `json:"callback_url"`
	Active             bool   `json:"active,omitempty"`
	ServiceDiscovery   bool   `json:"service_discovery,omitempty"`
	CarrierServiceType string `json:"carrier_service_type,omitempty"`
	Format             string `json:"format,omitempty"`
}

// CarrierServiceUpdateRequest contains the request body for updating a carrier service.
type CarrierServiceUpdateRequest struct {
	Name             string `json:"name,omitempty"`
	CallbackURL      string `json:"callback_url,omitempty"`
	Active           *bool  `json:"active,omitempty"`
	ServiceDiscovery *bool  `json:"service_discovery,omitempty"`
}

// ListCarrierServices retrieves a list of carrier services.
func (c *Client) ListCarrierServices(ctx context.Context, opts *CarrierServicesListOptions) (*CarrierServicesListResponse, error) {
	path := "/carrier_services"
	if opts != nil {
		params := url.Values{}
		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			params.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.Active != nil {
			params.Set("active", strconv.FormatBool(*opts.Active))
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp CarrierServicesListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCarrierService retrieves a single carrier service by ID.
func (c *Client) GetCarrierService(ctx context.Context, id string) (*CarrierService, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("carrier service id is required")
	}
	var cs CarrierService
	if err := c.Get(ctx, fmt.Sprintf("/carrier_services/%s", id), &cs); err != nil {
		return nil, err
	}
	return &cs, nil
}

// CreateCarrierService creates a new carrier service.
func (c *Client) CreateCarrierService(ctx context.Context, req *CarrierServiceCreateRequest) (*CarrierService, error) {
	var cs CarrierService
	if err := c.Post(ctx, "/carrier_services", req, &cs); err != nil {
		return nil, err
	}
	return &cs, nil
}

// UpdateCarrierService updates an existing carrier service.
func (c *Client) UpdateCarrierService(ctx context.Context, id string, req *CarrierServiceUpdateRequest) (*CarrierService, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("carrier service id is required")
	}
	var cs CarrierService
	if err := c.Put(ctx, fmt.Sprintf("/carrier_services/%s", id), req, &cs); err != nil {
		return nil, err
	}
	return &cs, nil
}

// DeleteCarrierService deletes a carrier service.
func (c *Client) DeleteCarrierService(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("carrier service id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/carrier_services/%s", id))
}
