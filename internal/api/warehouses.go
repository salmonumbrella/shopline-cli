package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Warehouse represents a Shopline warehouse.
type Warehouse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Code         string    `json:"code"`
	Address1     string    `json:"address1"`
	Address2     string    `json:"address2"`
	City         string    `json:"city"`
	Province     string    `json:"province"`
	ProvinceCode string    `json:"province_code"`
	Country      string    `json:"country"`
	CountryCode  string    `json:"country_code"`
	Zip          string    `json:"zip"`
	Phone        string    `json:"phone"`
	Email        string    `json:"email"`
	Active       bool      `json:"active"`
	IsDefault    bool      `json:"is_default"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// WarehousesListOptions contains options for listing warehouses.
type WarehousesListOptions struct {
	Page     int
	PageSize int
	Active   *bool
}

// WarehousesListResponse is the paginated response for warehouses.
type WarehousesListResponse = ListResponse[Warehouse]

// WarehouseCreateRequest contains the request body for creating a warehouse.
type WarehouseCreateRequest struct {
	Name     string `json:"name"`
	Code     string `json:"code,omitempty"`
	Address1 string `json:"address1"`
	Address2 string `json:"address2,omitempty"`
	City     string `json:"city"`
	Province string `json:"province,omitempty"`
	Country  string `json:"country"`
	Zip      string `json:"zip,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Email    string `json:"email,omitempty"`
}

// WarehouseUpdateRequest contains the request body for updating a warehouse.
type WarehouseUpdateRequest struct {
	Name     string `json:"name,omitempty"`
	Code     string `json:"code,omitempty"`
	Address1 string `json:"address1,omitempty"`
	Address2 string `json:"address2,omitempty"`
	City     string `json:"city,omitempty"`
	Province string `json:"province,omitempty"`
	Country  string `json:"country,omitempty"`
	Zip      string `json:"zip,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Email    string `json:"email,omitempty"`
	Active   *bool  `json:"active,omitempty"`
}

// ListWarehouses retrieves a list of warehouses.
func (c *Client) ListWarehouses(ctx context.Context, opts *WarehousesListOptions) (*WarehousesListResponse, error) {
	path := "/warehouses"
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

	var resp WarehousesListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetWarehouse retrieves a single warehouse by ID.
func (c *Client) GetWarehouse(ctx context.Context, id string) (*Warehouse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("warehouse id is required")
	}
	var warehouse Warehouse
	if err := c.Get(ctx, fmt.Sprintf("/warehouses/%s", id), &warehouse); err != nil {
		return nil, err
	}
	return &warehouse, nil
}

// CreateWarehouse creates a new warehouse.
func (c *Client) CreateWarehouse(ctx context.Context, req *WarehouseCreateRequest) (*Warehouse, error) {
	var warehouse Warehouse
	if err := c.Post(ctx, "/warehouses", req, &warehouse); err != nil {
		return nil, err
	}
	return &warehouse, nil
}

// UpdateWarehouse updates an existing warehouse.
func (c *Client) UpdateWarehouse(ctx context.Context, id string, req *WarehouseUpdateRequest) (*Warehouse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("warehouse id is required")
	}
	var warehouse Warehouse
	if err := c.Put(ctx, fmt.Sprintf("/warehouses/%s", id), req, &warehouse); err != nil {
		return nil, err
	}
	return &warehouse, nil
}

// DeleteWarehouse deletes a warehouse.
func (c *Client) DeleteWarehouse(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("warehouse id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/warehouses/%s", id))
}
