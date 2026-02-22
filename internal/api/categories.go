package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Category represents a Shopline product category.
type Category struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Handle      string    `json:"handle"`
	Description string    `json:"description"`
	ParentID    string    `json:"parent_id"`
	Position    int       `json:"position"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CategoriesListOptions contains options for listing categories.
type CategoriesListOptions struct {
	Page      int
	PageSize  int
	ParentID  string
	SortBy    string
	SortOrder string
}

// CategoriesListResponse is the paginated response for categories.
type CategoriesListResponse = ListResponse[Category]

// CategoryCreateRequest contains the request body for creating a category.
type CategoryCreateRequest struct {
	Title       string `json:"title"`
	Handle      string `json:"handle,omitempty"`
	Description string `json:"description,omitempty"`
	ParentID    string `json:"parent_id,omitempty"`
	Position    int    `json:"position,omitempty"`
}

// CategoryUpdateRequest contains the request body for updating a category.
type CategoryUpdateRequest struct {
	Title       string `json:"title,omitempty"`
	Handle      string `json:"handle,omitempty"`
	Description string `json:"description,omitempty"`
	ParentID    string `json:"parent_id,omitempty"`
	Position    int    `json:"position,omitempty"`
}

// ListCategories retrieves a list of categories.
func (c *Client) ListCategories(ctx context.Context, opts *CategoriesListOptions) (*CategoriesListResponse, error) {
	path := "/categories"
	if opts != nil {
		params := url.Values{}
		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			params.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.ParentID != "" {
			params.Set("parent_id", opts.ParentID)
		}
		if opts.SortBy != "" {
			params.Set("sort_by", opts.SortBy)
		}
		if opts.SortOrder != "" {
			params.Set("sort_order", opts.SortOrder)
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp CategoriesListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCategory retrieves a single category by ID.
func (c *Client) GetCategory(ctx context.Context, id string) (*Category, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("category id is required")
	}
	var category Category
	if err := c.Get(ctx, fmt.Sprintf("/categories/%s", id), &category); err != nil {
		return nil, err
	}
	return &category, nil
}

// CreateCategory creates a new category.
func (c *Client) CreateCategory(ctx context.Context, req *CategoryCreateRequest) (*Category, error) {
	var category Category
	if err := c.Post(ctx, "/categories", req, &category); err != nil {
		return nil, err
	}
	return &category, nil
}

// UpdateCategory updates an existing category.
func (c *Client) UpdateCategory(ctx context.Context, id string, req *CategoryUpdateRequest) (*Category, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("category id is required")
	}
	var category Category
	if err := c.Put(ctx, fmt.Sprintf("/categories/%s", id), req, &category); err != nil {
		return nil, err
	}
	return &category, nil
}

// DeleteCategory deletes a category.
func (c *Client) DeleteCategory(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("category id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/categories/%s", id))
}

// BulkUpdateCategoryProductSorting updates the product sorting within a category.
//
// Docs: PUT /categories/{id}/products_sorting
func (c *Client) BulkUpdateCategoryProductSorting(ctx context.Context, id string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("category id is required")
	}

	var resp json.RawMessage
	if err := c.Put(ctx, fmt.Sprintf("/categories/%s/products_sorting", id), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
