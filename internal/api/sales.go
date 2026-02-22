package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Sale represents a sale campaign.
type Sale struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	DiscountType  string    `json:"discount_type"`
	DiscountValue float64   `json:"discount_value"`
	AppliesTo     string    `json:"applies_to"`
	ProductIDs    []string  `json:"product_ids"`
	CollectionIDs []string  `json:"collection_ids"`
	Status        string    `json:"status"`
	StartsAt      time.Time `json:"starts_at"`
	EndsAt        time.Time `json:"ends_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// SalesListOptions contains options for listing sales.
type SalesListOptions struct {
	Page     int
	PageSize int
	Status   string
}

// SalesListResponse is the paginated response for sales.
type SalesListResponse = ListResponse[Sale]

// SaleCreateRequest contains the request body for creating a sale.
type SaleCreateRequest struct {
	Title         string     `json:"title"`
	Description   string     `json:"description,omitempty"`
	DiscountType  string     `json:"discount_type"`
	DiscountValue float64    `json:"discount_value"`
	AppliesTo     string     `json:"applies_to"`
	ProductIDs    []string   `json:"product_ids,omitempty"`
	CollectionIDs []string   `json:"collection_ids,omitempty"`
	StartsAt      *time.Time `json:"starts_at"`
	EndsAt        *time.Time `json:"ends_at,omitempty"`
}

// ListSales retrieves a list of sale campaigns.
func (c *Client) ListSales(ctx context.Context, opts *SalesListOptions) (*SalesListResponse, error) {
	path := "/sales" + NewQuery().
		Int("page", opts.Page).
		Int("page_size", opts.PageSize).
		String("status", opts.Status).
		Build()

	var resp SalesListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetSale retrieves a single sale by ID.
func (c *Client) GetSale(ctx context.Context, id string) (*Sale, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("sale id is required")
	}
	var sale Sale
	if err := c.Get(ctx, fmt.Sprintf("/sales/%s", id), &sale); err != nil {
		return nil, err
	}
	return &sale, nil
}

// CreateSale creates a new sale campaign.
func (c *Client) CreateSale(ctx context.Context, req *SaleCreateRequest) (*Sale, error) {
	var sale Sale
	if err := c.Post(ctx, "/sales", req, &sale); err != nil {
		return nil, err
	}
	return &sale, nil
}

// DeleteSale deletes a sale campaign.
func (c *Client) DeleteSale(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("sale id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/sales/%s", id))
}

// ActivateSale activates a sale campaign.
func (c *Client) ActivateSale(ctx context.Context, id string) (*Sale, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("sale id is required")
	}
	var sale Sale
	if err := c.Post(ctx, fmt.Sprintf("/sales/%s/activate", id), nil, &sale); err != nil {
		return nil, err
	}
	return &sale, nil
}

// DeactivateSale deactivates a sale campaign.
func (c *Client) DeactivateSale(ctx context.Context, id string) (*Sale, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("sale id is required")
	}
	var sale Sale
	if err := c.Post(ctx, fmt.Sprintf("/sales/%s/deactivate", id), nil, &sale); err != nil {
		return nil, err
	}
	return &sale, nil
}

// =============================================================================
// Live Streaming Endpoints
// Note: Live streaming endpoints have a rate limit of 1 request/second.
// =============================================================================

// SaleProduct represents a product in a live sale.
type SaleProduct struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	VariantID string    `json:"variant_id,omitempty"`
	Price     float64   `json:"price"`
	Quantity  int       `json:"quantity"`
	Keywords  []string  `json:"keywords,omitempty"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SaleProductsListResponse is the paginated response for sale products.
type SaleProductsListResponse = ListResponse[SaleProduct]

// SaleProductsListOptions contains options for listing sale products.
type SaleProductsListOptions struct {
	Page     int
	PageSize int
}

// SaleProductRequest represents a product to add or update in a sale.
type SaleProductRequest struct {
	ProductID string   `json:"product_id"`
	VariantID string   `json:"variant_id,omitempty"`
	Price     float64  `json:"price,omitempty"`
	Quantity  int      `json:"quantity,omitempty"`
	Keywords  []string `json:"keywords,omitempty"`
}

// SaleAddProductsRequest contains the request body for adding products to a sale.
// Maximum 100 products per request.
type SaleAddProductsRequest struct {
	Products []SaleProductRequest `json:"products"`
}

// SaleUpdateProductsRequest contains the request body for updating products in a sale.
type SaleUpdateProductsRequest struct {
	Products []SaleProductRequest `json:"products"`
}

// SaleDeleteProductsRequest contains the request body for deleting products from a sale.
type SaleDeleteProductsRequest struct {
	ProductIDs []string `json:"product_ids"`
}

// SaleComment represents a comment from a live stream.
type SaleComment struct {
	ID           string    `json:"id"`
	CustomerID   string    `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"created_at"`
}

// SaleCommentsListResponse is the paginated response for sale comments.
type SaleCommentsListResponse = ListResponse[SaleComment]

// SaleCommentsListOptions contains options for listing sale comments.
type SaleCommentsListOptions struct {
	Page     int
	PageSize int
}

// SaleCustomer represents a customer who commented in a live stream.
type SaleCustomer struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email,omitempty"`
	Phone        string `json:"phone,omitempty"`
	CommentCount int    `json:"comment_count"`
}

// SaleCustomersListResponse is the paginated response for sale customers.
type SaleCustomersListResponse = ListResponse[SaleCustomer]

// SaleCustomersListOptions contains options for listing sale customers.
type SaleCustomersListOptions struct {
	Page     int
	PageSize int
}

// GetSaleProducts retrieves products in a live sale.
func (c *Client) GetSaleProducts(ctx context.Context, saleID string, opts *SaleProductsListOptions) (*SaleProductsListResponse, error) {
	if strings.TrimSpace(saleID) == "" {
		return nil, fmt.Errorf("sale id is required")
	}

	path := fmt.Sprintf("/sales/%s/products", saleID) + NewQuery().
		Int("page", opts.Page).
		Int("page_size", opts.PageSize).
		Build()

	var resp SaleProductsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AddSaleProducts adds products to a live sale. Maximum 100 products per request.
func (c *Client) AddSaleProducts(ctx context.Context, saleID string, req *SaleAddProductsRequest) (*SaleProductsListResponse, error) {
	if strings.TrimSpace(saleID) == "" {
		return nil, fmt.Errorf("sale id is required")
	}
	if len(req.Products) == 0 {
		return nil, fmt.Errorf("at least one product is required")
	}
	if len(req.Products) > 100 {
		return nil, fmt.Errorf("maximum 100 products per request")
	}

	var resp SaleProductsListResponse
	if err := c.Post(ctx, fmt.Sprintf("/sales/%s/products", saleID), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateSaleProducts modifies products in a live sale.
func (c *Client) UpdateSaleProducts(ctx context.Context, saleID string, req *SaleUpdateProductsRequest) (*SaleProductsListResponse, error) {
	if strings.TrimSpace(saleID) == "" {
		return nil, fmt.Errorf("sale id is required")
	}
	if len(req.Products) == 0 {
		return nil, fmt.Errorf("at least one product is required")
	}

	var resp SaleProductsListResponse
	if err := c.Put(ctx, fmt.Sprintf("/sales/%s/products", saleID), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteSaleProducts removes products from a live sale.
func (c *Client) DeleteSaleProducts(ctx context.Context, saleID string, req *SaleDeleteProductsRequest) error {
	if strings.TrimSpace(saleID) == "" {
		return fmt.Errorf("sale id is required")
	}
	if len(req.ProductIDs) == 0 {
		return fmt.Errorf("at least one product id is required")
	}

	return c.Post(ctx, fmt.Sprintf("/sales/%s/delete_products", saleID), req, nil)
}

// GetSaleComments retrieves comments from a live stream.
func (c *Client) GetSaleComments(ctx context.Context, saleID string, opts *SaleCommentsListOptions) (*SaleCommentsListResponse, error) {
	if strings.TrimSpace(saleID) == "" {
		return nil, fmt.Errorf("sale id is required")
	}

	path := fmt.Sprintf("/sales/%s/comments", saleID) + NewQuery().
		Int("page", opts.Page).
		Int("page_size", opts.PageSize).
		Build()

	var resp SaleCommentsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetSaleCustomers retrieves customers who commented in a live stream.
func (c *Client) GetSaleCustomers(ctx context.Context, saleID string, opts *SaleCustomersListOptions) (*SaleCustomersListResponse, error) {
	if strings.TrimSpace(saleID) == "" {
		return nil, fmt.Errorf("sale id is required")
	}

	path := fmt.Sprintf("/sales/%s/customers", saleID) + NewQuery().
		Int("page", opts.Page).
		Int("page_size", opts.PageSize).
		Build()

	var resp SaleCustomersListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
