package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CompanyCatalog represents a B2B company catalog.
type CompanyCatalog struct {
	ID          string    `json:"id"`
	CompanyID   string    `json:"company_id"`
	CompanyName string    `json:"company_name"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ProductIDs  []string  `json:"product_ids"`
	IsDefault   bool      `json:"is_default"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CompanyCatalogsListOptions contains options for listing company catalogs.
type CompanyCatalogsListOptions struct {
	Page      int
	PageSize  int
	CompanyID string
	Status    string
}

// CompanyCatalogsListResponse is the paginated response for company catalogs.
type CompanyCatalogsListResponse = ListResponse[CompanyCatalog]

// CompanyCatalogCreateRequest contains the request body for creating a company catalog.
type CompanyCatalogCreateRequest struct {
	CompanyID   string   `json:"company_id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	ProductIDs  []string `json:"product_ids,omitempty"`
	IsDefault   bool     `json:"is_default,omitempty"`
}

// CompanyCatalogUpdateRequest contains the request body for updating a company catalog.
type CompanyCatalogUpdateRequest struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	ProductIDs  []string `json:"product_ids,omitempty"`
	IsDefault   *bool    `json:"is_default,omitempty"`
}

// ListCompanyCatalogs retrieves a list of company catalogs.
func (c *Client) ListCompanyCatalogs(ctx context.Context, opts *CompanyCatalogsListOptions) (*CompanyCatalogsListResponse, error) {
	path := "/company_catalogs" + NewQuery().
		Int("page", opts.Page).
		Int("page_size", opts.PageSize).
		String("company_id", opts.CompanyID).
		String("status", opts.Status).
		Build()

	var resp CompanyCatalogsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCompanyCatalog retrieves a single company catalog by ID.
func (c *Client) GetCompanyCatalog(ctx context.Context, id string) (*CompanyCatalog, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("company catalog id is required")
	}
	var catalog CompanyCatalog
	if err := c.Get(ctx, fmt.Sprintf("/company_catalogs/%s", id), &catalog); err != nil {
		return nil, err
	}
	return &catalog, nil
}

// CreateCompanyCatalog creates a new company catalog.
func (c *Client) CreateCompanyCatalog(ctx context.Context, req *CompanyCatalogCreateRequest) (*CompanyCatalog, error) {
	var catalog CompanyCatalog
	if err := c.Post(ctx, "/company_catalogs", req, &catalog); err != nil {
		return nil, err
	}
	return &catalog, nil
}

// UpdateCompanyCatalog updates an existing company catalog.
func (c *Client) UpdateCompanyCatalog(ctx context.Context, id string, req *CompanyCatalogUpdateRequest) (*CompanyCatalog, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("company catalog id is required")
	}
	var catalog CompanyCatalog
	if err := c.Put(ctx, fmt.Sprintf("/company_catalogs/%s", id), req, &catalog); err != nil {
		return nil, err
	}
	return &catalog, nil
}

// DeleteCompanyCatalog deletes a company catalog.
func (c *Client) DeleteCompanyCatalog(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("company catalog id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/company_catalogs/%s", id))
}
