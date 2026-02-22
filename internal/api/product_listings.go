package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ProductListing represents a product published to a sales channel.
type ProductListing struct {
	ID          string    `json:"id"`
	ProductID   string    `json:"product_id"`
	Title       string    `json:"title"`
	Handle      string    `json:"handle"`
	BodyHTML    string    `json:"body_html"`
	Vendor      string    `json:"vendor"`
	ProductType string    `json:"product_type"`
	Available   bool      `json:"available"`
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProductListingsListOptions contains options for listing product listings.
type ProductListingsListOptions struct {
	Page     int
	PageSize int
}

// ProductListingsListResponse is the paginated response for product listings.
type ProductListingsListResponse = ListResponse[ProductListing]

// ListProductListings retrieves a list of product listings.
func (c *Client) ListProductListings(ctx context.Context, opts *ProductListingsListOptions) (*ProductListingsListResponse, error) {
	path := "/product_listings"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			Build()
	}

	var resp ProductListingsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetProductListing retrieves a single product listing by ID.
func (c *Client) GetProductListing(ctx context.Context, id string) (*ProductListing, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("product listing id is required")
	}
	var listing ProductListing
	if err := c.Get(ctx, fmt.Sprintf("/product_listings/%s", id), &listing); err != nil {
		return nil, err
	}
	return &listing, nil
}

// CreateProductListing publishes a product to a sales channel.
func (c *Client) CreateProductListing(ctx context.Context, productID string) (*ProductListing, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("product id is required")
	}

	req := struct {
		ProductID string `json:"product_id"`
	}{
		ProductID: productID,
	}

	var listing ProductListing
	if err := c.Post(ctx, "/product_listings", req, &listing); err != nil {
		return nil, err
	}
	return &listing, nil
}

// DeleteProductListing removes a product listing from a sales channel.
func (c *Client) DeleteProductListing(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("product listing id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/product_listings/%s", id))
}
