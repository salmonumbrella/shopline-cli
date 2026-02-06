package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// StorefrontCart represents a shopping cart in the storefront.
type StorefrontCart struct {
	ID            string               `json:"id"`
	CustomerID    string               `json:"customer_id"`
	Email         string               `json:"email"`
	Currency      string               `json:"currency"`
	Subtotal      string               `json:"subtotal"`
	TotalPrice    string               `json:"total_price"`
	TotalTax      string               `json:"total_tax"`
	TotalDiscount string               `json:"total_discount"`
	ItemCount     int                  `json:"item_count"`
	Items         []StorefrontCartItem `json:"items"`
	AbandonedAt   *time.Time           `json:"abandoned_at"`
	RecoveredAt   *time.Time           `json:"recovered_at"`
	CompletedAt   *time.Time           `json:"completed_at"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

// StorefrontCartItem represents an item in a storefront cart.
type StorefrontCartItem struct {
	ID           string `json:"id"`
	ProductID    string `json:"product_id"`
	VariantID    string `json:"variant_id"`
	Title        string `json:"title"`
	VariantTitle string `json:"variant_title"`
	Quantity     int    `json:"quantity"`
	Price        string `json:"price"`
	LineTotal    string `json:"line_total"`
	ImageURL     string `json:"image_url"`
}

// StorefrontCartsListOptions contains options for listing storefront carts.
type StorefrontCartsListOptions struct {
	Page       int
	PageSize   int
	CustomerID string
	Status     string // active, abandoned, completed
	SortBy     string
	SortOrder  string
}

// StorefrontCartsListResponse is the paginated response for storefront carts.
type StorefrontCartsListResponse = ListResponse[StorefrontCart]

// StorefrontCartCreateRequest contains the data for creating a storefront cart.
type StorefrontCartCreateRequest struct {
	CustomerID string                      `json:"customer_id,omitempty"`
	Email      string                      `json:"email,omitempty"`
	Currency   string                      `json:"currency,omitempty"`
	Items      []StorefrontCartItemRequest `json:"items,omitempty"`
}

// StorefrontCartItemRequest represents an item to add to a cart.
type StorefrontCartItemRequest struct {
	ProductID string `json:"product_id"`
	VariantID string `json:"variant_id"`
	Quantity  int    `json:"quantity"`
}

// ListStorefrontCarts retrieves a list of storefront carts.
func (c *Client) ListStorefrontCarts(ctx context.Context, opts *StorefrontCartsListOptions) (*StorefrontCartsListResponse, error) {
	path := "/storefront/carts"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("customer_id", opts.CustomerID).
			String("status", opts.Status).
			String("sort_by", opts.SortBy).
			String("sort_order", opts.SortOrder).
			Build()
	}

	var resp StorefrontCartsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetStorefrontCart retrieves a single storefront cart by ID.
func (c *Client) GetStorefrontCart(ctx context.Context, id string) (*StorefrontCart, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("cart id is required")
	}
	var cart StorefrontCart
	if err := c.Get(ctx, fmt.Sprintf("/storefront/carts/%s", id), &cart); err != nil {
		return nil, err
	}
	return &cart, nil
}

// CreateStorefrontCart creates a new storefront cart.
func (c *Client) CreateStorefrontCart(ctx context.Context, req *StorefrontCartCreateRequest) (*StorefrontCart, error) {
	var cart StorefrontCart
	if err := c.Post(ctx, "/storefront/carts", req, &cart); err != nil {
		return nil, err
	}
	return &cart, nil
}

// DeleteStorefrontCart deletes a storefront cart.
func (c *Client) DeleteStorefrontCart(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("cart id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/storefront/carts/%s", id))
}
