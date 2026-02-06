package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Taxonomy represents a Shopline product taxonomy/category.
type Taxonomy struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Handle       string    `json:"handle"`
	Description  string    `json:"description"`
	ParentID     string    `json:"parent_id"`
	Level        int       `json:"level"`
	Position     int       `json:"position"`
	Path         string    `json:"path"`
	FullPath     string    `json:"full_path"`
	ProductCount int       `json:"product_count"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TaxonomiesListOptions contains options for listing taxonomies.
type TaxonomiesListOptions struct {
	Page     int
	PageSize int
	ParentID string
	Active   *bool
}

// TaxonomiesListResponse is the paginated response for taxonomies.
type TaxonomiesListResponse = ListResponse[Taxonomy]

// TaxonomyCreateRequest contains the data for creating a taxonomy.
type TaxonomyCreateRequest struct {
	Name        string `json:"name"`
	Handle      string `json:"handle,omitempty"`
	Description string `json:"description,omitempty"`
	ParentID    string `json:"parent_id,omitempty"`
	Position    int    `json:"position,omitempty"`
	Active      bool   `json:"active"`
}

// TaxonomyUpdateRequest contains the data for updating a taxonomy.
type TaxonomyUpdateRequest struct {
	Name        string `json:"name,omitempty"`
	Handle      string `json:"handle,omitempty"`
	Description string `json:"description,omitempty"`
	ParentID    string `json:"parent_id,omitempty"`
	Position    int    `json:"position,omitempty"`
	Active      *bool  `json:"active,omitempty"`
}

// ListTaxonomies retrieves a list of taxonomies.
func (c *Client) ListTaxonomies(ctx context.Context, opts *TaxonomiesListOptions) (*TaxonomiesListResponse, error) {
	path := "/taxonomies"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("parent_id", opts.ParentID).
			BoolPtr("active", opts.Active).
			Build()
	}

	var resp TaxonomiesListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTaxonomy retrieves a single taxonomy by ID.
func (c *Client) GetTaxonomy(ctx context.Context, id string) (*Taxonomy, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("taxonomy id is required")
	}
	var taxonomy Taxonomy
	if err := c.Get(ctx, fmt.Sprintf("/taxonomies/%s", id), &taxonomy); err != nil {
		return nil, err
	}
	return &taxonomy, nil
}

// CreateTaxonomy creates a new taxonomy.
func (c *Client) CreateTaxonomy(ctx context.Context, req *TaxonomyCreateRequest) (*Taxonomy, error) {
	var taxonomy Taxonomy
	if err := c.Post(ctx, "/taxonomies", req, &taxonomy); err != nil {
		return nil, err
	}
	return &taxonomy, nil
}

// UpdateTaxonomy updates an existing taxonomy.
func (c *Client) UpdateTaxonomy(ctx context.Context, id string, req *TaxonomyUpdateRequest) (*Taxonomy, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("taxonomy id is required")
	}
	var taxonomy Taxonomy
	if err := c.Put(ctx, fmt.Sprintf("/taxonomies/%s", id), req, &taxonomy); err != nil {
		return nil, err
	}
	return &taxonomy, nil
}

// DeleteTaxonomy deletes a taxonomy.
func (c *Client) DeleteTaxonomy(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("taxonomy id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/taxonomies/%s", id))
}
