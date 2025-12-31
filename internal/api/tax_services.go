package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// TaxService represents a tax service provider.
type TaxService struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Provider    string    `json:"provider"`
	APIKey      string    `json:"api_key"`
	APISecret   string    `json:"api_secret"`
	Sandbox     bool      `json:"sandbox"`
	Active      bool      `json:"active"`
	CallbackURL string    `json:"callback_url"`
	Countries   []string  `json:"countries"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TaxServicesListOptions contains options for listing tax services.
type TaxServicesListOptions struct {
	Page     int
	PageSize int
	Provider string
	Active   *bool
}

// TaxServicesListResponse is the paginated response for tax services.
type TaxServicesListResponse = ListResponse[TaxService]

// TaxServiceCreateRequest contains the request body for creating a tax service.
type TaxServiceCreateRequest struct {
	Name        string   `json:"name"`
	Provider    string   `json:"provider"`
	APIKey      string   `json:"api_key"`
	APISecret   string   `json:"api_secret,omitempty"`
	Sandbox     bool     `json:"sandbox,omitempty"`
	Active      bool     `json:"active,omitempty"`
	CallbackURL string   `json:"callback_url,omitempty"`
	Countries   []string `json:"countries,omitempty"`
}

// TaxServiceUpdateRequest contains the request body for updating a tax service.
type TaxServiceUpdateRequest struct {
	Name        string   `json:"name,omitempty"`
	APIKey      string   `json:"api_key,omitempty"`
	APISecret   string   `json:"api_secret,omitempty"`
	Sandbox     *bool    `json:"sandbox,omitempty"`
	Active      *bool    `json:"active,omitempty"`
	CallbackURL string   `json:"callback_url,omitempty"`
	Countries   []string `json:"countries,omitempty"`
}

// ListTaxServices retrieves a list of tax services.
func (c *Client) ListTaxServices(ctx context.Context, opts *TaxServicesListOptions) (*TaxServicesListResponse, error) {
	path := "/tax_services"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("provider", opts.Provider).
			BoolPtr("active", opts.Active).
			Build()
	}

	var resp TaxServicesListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTaxService retrieves a single tax service by ID.
func (c *Client) GetTaxService(ctx context.Context, id string) (*TaxService, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("tax service id is required")
	}
	var service TaxService
	if err := c.Get(ctx, fmt.Sprintf("/tax_services/%s", id), &service); err != nil {
		return nil, err
	}
	return &service, nil
}

// CreateTaxService creates a new tax service.
func (c *Client) CreateTaxService(ctx context.Context, req *TaxServiceCreateRequest) (*TaxService, error) {
	var service TaxService
	if err := c.Post(ctx, "/tax_services", req, &service); err != nil {
		return nil, err
	}
	return &service, nil
}

// UpdateTaxService updates an existing tax service.
func (c *Client) UpdateTaxService(ctx context.Context, id string, req *TaxServiceUpdateRequest) (*TaxService, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("tax service id is required")
	}
	var service TaxService
	if err := c.Put(ctx, fmt.Sprintf("/tax_services/%s", id), req, &service); err != nil {
		return nil, err
	}
	return &service, nil
}

// DeleteTaxService deletes a tax service.
func (c *Client) DeleteTaxService(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("tax service id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/tax_services/%s", id))
}
