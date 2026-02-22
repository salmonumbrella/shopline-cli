package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Collection represents a Shopline collection.
type Collection struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Handle         string    `json:"handle"`
	Description    string    `json:"description"`
	SortOrder      string    `json:"sort_order"`
	ProductsCount  int       `json:"products_count"`
	PublishedScope string    `json:"published_scope"`
	PublishedAt    time.Time `json:"published_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CollectionsListOptions contains options for listing collections.
type CollectionsListOptions struct {
	Page           int
	PageSize       int
	Title          string
	Handle         string
	PublishedScope string
	SortBy         string
	SortOrder      string
}

// CollectionsListResponse is the paginated response for collections.
type CollectionsListResponse = ListResponse[Collection]

// CollectionCreateRequest contains the request body for creating a collection.
type CollectionCreateRequest struct {
	Title          string   `json:"title"`
	Handle         string   `json:"handle,omitempty"`
	Description    string   `json:"description,omitempty"`
	SortOrder      string   `json:"sort_order,omitempty"`
	PublishedScope string   `json:"published_scope,omitempty"`
	ProductIDs     []string `json:"product_ids,omitempty"`
}

// CollectionUpdateRequest contains the request body for updating a collection.
type CollectionUpdateRequest struct {
	Title          string `json:"title,omitempty"`
	Handle         string `json:"handle,omitempty"`
	Description    string `json:"description,omitempty"`
	SortOrder      string `json:"sort_order,omitempty"`
	PublishedScope string `json:"published_scope,omitempty"`
}

// ListCollections retrieves a list of collections.
func (c *Client) ListCollections(ctx context.Context, opts *CollectionsListOptions) (*CollectionsListResponse, error) {
	path := "/collections"
	if opts != nil {
		params := url.Values{}
		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			params.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.Title != "" {
			params.Set("title", opts.Title)
		}
		if opts.Handle != "" {
			params.Set("handle", opts.Handle)
		}
		if opts.PublishedScope != "" {
			params.Set("published_scope", opts.PublishedScope)
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

	var resp CollectionsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCollection retrieves a single collection by ID.
func (c *Client) GetCollection(ctx context.Context, id string) (*Collection, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("collection id is required")
	}
	var collection Collection
	if err := c.Get(ctx, fmt.Sprintf("/collections/%s", id), &collection); err != nil {
		return nil, err
	}
	return &collection, nil
}

// CreateCollection creates a new collection.
func (c *Client) CreateCollection(ctx context.Context, req *CollectionCreateRequest) (*Collection, error) {
	var collection Collection
	if err := c.Post(ctx, "/collections", req, &collection); err != nil {
		return nil, err
	}
	return &collection, nil
}

// UpdateCollection updates an existing collection.
func (c *Client) UpdateCollection(ctx context.Context, id string, req *CollectionUpdateRequest) (*Collection, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("collection id is required")
	}
	var collection Collection
	if err := c.Put(ctx, fmt.Sprintf("/collections/%s", id), req, &collection); err != nil {
		return nil, err
	}
	return &collection, nil
}

// DeleteCollection deletes a collection.
func (c *Client) DeleteCollection(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("collection id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/collections/%s", id))
}

// AddProductsToCollection adds products to a collection.
func (c *Client) AddProductsToCollection(ctx context.Context, id string, productIDs []string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("collection id is required")
	}
	body := map[string][]string{"product_ids": productIDs}
	return c.Post(ctx, fmt.Sprintf("/collections/%s/products", id), body, nil)
}

// RemoveProductFromCollection removes a product from a collection.
func (c *Client) RemoveProductFromCollection(ctx context.Context, id, productID string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("collection id is required")
	}
	if strings.TrimSpace(productID) == "" {
		return fmt.Errorf("product id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/collections/%s/products/%s", id, productID))
}
