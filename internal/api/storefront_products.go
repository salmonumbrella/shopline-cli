package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// StorefrontProduct represents a product as displayed in the storefront.
type StorefrontProduct struct {
	ID             string                     `json:"id"`
	Handle         string                     `json:"handle"`
	Title          string                     `json:"title"`
	Description    string                     `json:"description"`
	Vendor         string                     `json:"vendor"`
	ProductType    string                     `json:"product_type"`
	Tags           []string                   `json:"tags"`
	Status         string                     `json:"status"`
	Available      bool                       `json:"available"`
	Price          string                     `json:"price"`
	CompareAtPrice string                     `json:"compare_at_price"`
	Currency       string                     `json:"currency"`
	Images         []StorefrontProductImage   `json:"images"`
	Variants       []StorefrontProductVariant `json:"variants"`
	Options        []StorefrontProductOption  `json:"options"`
	SEOTitle       string                     `json:"seo_title"`
	SEODescription string                     `json:"seo_description"`
	ViewCount      int                        `json:"view_count"`
	SalesCount     int                        `json:"sales_count"`
	ReviewCount    int                        `json:"review_count"`
	AverageRating  float64                    `json:"average_rating"`
	PublishedAt    *time.Time                 `json:"published_at"`
	CreatedAt      time.Time                  `json:"created_at"`
	UpdatedAt      time.Time                  `json:"updated_at"`
}

// StorefrontProductImage represents a product image.
type StorefrontProductImage struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	AltText  string `json:"alt_text"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Position int    `json:"position"`
}

// StorefrontProductVariant represents a product variant.
type StorefrontProductVariant struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	SKU            string `json:"sku"`
	Price          string `json:"price"`
	CompareAtPrice string `json:"compare_at_price"`
	Available      bool   `json:"available"`
	Inventory      int    `json:"inventory"`
	Weight         string `json:"weight"`
	WeightUnit     string `json:"weight_unit"`
}

// StorefrontProductOption represents a product option (e.g., size, color).
type StorefrontProductOption struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Position int      `json:"position"`
	Values   []string `json:"values"`
}

// StorefrontProductsListOptions contains options for listing storefront products.
type StorefrontProductsListOptions struct {
	Page        int
	PageSize    int
	Collection  string
	Category    string
	Vendor      string
	ProductType string
	Tag         string
	Available   *bool
	MinPrice    string
	MaxPrice    string
	SortBy      string
	SortOrder   string
	Query       string
}

// StorefrontProductsListResponse is the paginated response for storefront products.
type StorefrontProductsListResponse = ListResponse[StorefrontProduct]

// ListStorefrontProducts retrieves a list of storefront products.
func (c *Client) ListStorefrontProducts(ctx context.Context, opts *StorefrontProductsListOptions) (*StorefrontProductsListResponse, error) {
	path := "/storefront/products"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("collection", opts.Collection).
			String("category", opts.Category).
			String("vendor", opts.Vendor).
			String("product_type", opts.ProductType).
			String("tag", opts.Tag).
			BoolPtr("available", opts.Available).
			String("min_price", opts.MinPrice).
			String("max_price", opts.MaxPrice).
			String("sort_by", opts.SortBy).
			String("sort_order", opts.SortOrder).
			String("query", opts.Query).
			Build()
	}

	var resp StorefrontProductsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetStorefrontProduct retrieves a single storefront product by ID.
func (c *Client) GetStorefrontProduct(ctx context.Context, id string) (*StorefrontProduct, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	var product StorefrontProduct
	if err := c.Get(ctx, fmt.Sprintf("/storefront/products/%s", id), &product); err != nil {
		return nil, err
	}
	return &product, nil
}

// GetStorefrontProductByHandle retrieves a storefront product by handle.
func (c *Client) GetStorefrontProductByHandle(ctx context.Context, handle string) (*StorefrontProduct, error) {
	if strings.TrimSpace(handle) == "" {
		return nil, fmt.Errorf("product handle is required")
	}
	var product StorefrontProduct
	if err := c.Get(ctx, fmt.Sprintf("/storefront/products/handle/%s", handle), &product); err != nil {
		return nil, err
	}
	return &product, nil
}
