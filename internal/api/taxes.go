package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Tax represents a tax configuration.
type Tax struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Rate         float64   `json:"rate"`
	CountryCode  string    `json:"country_code"`
	ProvinceCode string    `json:"province_code"`
	Priority     int       `json:"priority"`
	Compound     bool      `json:"compound"`
	Shipping     bool      `json:"shipping"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TaxesListOptions contains options for listing taxes.
type TaxesListOptions struct {
	Page        int
	PageSize    int
	CountryCode string
	Enabled     *bool
}

// TaxesListResponse is the paginated response for taxes.
type TaxesListResponse = ListResponse[Tax]

// TaxCreateRequest contains the request body for creating a tax.
type TaxCreateRequest struct {
	Name         string  `json:"name"`
	Rate         float64 `json:"rate"`
	CountryCode  string  `json:"country_code"`
	ProvinceCode string  `json:"province_code,omitempty"`
	Priority     int     `json:"priority,omitempty"`
	Compound     bool    `json:"compound,omitempty"`
	Shipping     bool    `json:"shipping,omitempty"`
	Enabled      bool    `json:"enabled,omitempty"`
}

// TaxUpdateRequest contains the request body for updating a tax.
type TaxUpdateRequest struct {
	Name     string   `json:"name,omitempty"`
	Rate     *float64 `json:"rate,omitempty"`
	Priority int      `json:"priority,omitempty"`
	Compound *bool    `json:"compound,omitempty"`
	Shipping *bool    `json:"shipping,omitempty"`
	Enabled  *bool    `json:"enabled,omitempty"`
}

// ListTaxes retrieves a list of taxes.
func (c *Client) ListTaxes(ctx context.Context, opts *TaxesListOptions) (*TaxesListResponse, error) {
	path := "/taxes"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("country_code", opts.CountryCode).
			BoolPtr("enabled", opts.Enabled).
			Build()
	}

	var resp TaxesListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTax retrieves a single tax by ID.
func (c *Client) GetTax(ctx context.Context, id string) (*Tax, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("tax id is required")
	}
	var tax Tax
	if err := c.Get(ctx, fmt.Sprintf("/taxes/%s", id), &tax); err != nil {
		return nil, err
	}
	return &tax, nil
}

// CreateTax creates a new tax.
func (c *Client) CreateTax(ctx context.Context, req *TaxCreateRequest) (*Tax, error) {
	var tax Tax
	if err := c.Post(ctx, "/taxes", req, &tax); err != nil {
		return nil, err
	}
	return &tax, nil
}

// UpdateTax updates an existing tax.
func (c *Client) UpdateTax(ctx context.Context, id string, req *TaxUpdateRequest) (*Tax, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("tax id is required")
	}
	var tax Tax
	if err := c.Put(ctx, fmt.Sprintf("/taxes/%s", id), req, &tax); err != nil {
		return nil, err
	}
	return &tax, nil
}

// DeleteTax deletes a tax.
func (c *Client) DeleteTax(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("tax id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/taxes/%s", id))
}
