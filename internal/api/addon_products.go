package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// AddonProduct represents a Shopline add-on product bundle.
type AddonProduct struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	ProductID   string    `json:"product_id"`
	VariantID   string    `json:"variant_id"`
	Price       string    `json:"price"`
	Currency    string    `json:"currency"`
	Quantity    int       `json:"quantity"`
	Position    int       `json:"position"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AddonProductsListOptions contains options for listing addon products.
type AddonProductsListOptions struct {
	Page      int
	PageSize  int
	ProductID string
	Status    string
}

// AddonProductsListResponse is the paginated response for addon products.
type AddonProductsListResponse = ListResponse[AddonProduct]

// AddonProductCreateRequest contains the data for creating an addon product.
type AddonProductCreateRequest struct {
	Title       string `json:"title"`
	ProductID   string `json:"product_id"`
	VariantID   string `json:"variant_id,omitempty"`
	Price       string `json:"price,omitempty"`
	Quantity    int    `json:"quantity,omitempty"`
	Position    int    `json:"position,omitempty"`
	Description string `json:"description,omitempty"`
}

// AddonProductUpdateRequest contains the data for updating an addon product.
// All fields are optional - only provided fields will be updated.
type AddonProductUpdateRequest struct {
	Title       *string `json:"title,omitempty"`
	ProductID   *string `json:"product_id,omitempty"`
	VariantID   *string `json:"variant_id,omitempty"`
	Price       *string `json:"price,omitempty"`
	Quantity    *int    `json:"quantity,omitempty"`
	Position    *int    `json:"position,omitempty"`
	Status      *string `json:"status,omitempty"`
	Description *string `json:"description,omitempty"`
}

// AddonProductSearchOptions contains options for searching addon products.
type AddonProductSearchOptions struct {
	Query     string
	ProductID string
	Status    string
	Page      int
	PageSize  int
}

// AddonProductQuantityRequest contains the data for updating addon product quantity.
type AddonProductQuantityRequest struct {
	Quantity int `json:"quantity"`
}

// AddonProductQuantityBySKURequest contains the data for updating addon product quantity by SKU.
type AddonProductQuantityBySKURequest struct {
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
}

// AddonProductStock represents stock information for an addon product at a warehouse.
type AddonProductStock struct {
	WarehouseID string `json:"warehouse_id"`
	Quantity    int    `json:"quantity"`
	Reserved    int    `json:"reserved,omitempty"`
	Available   int    `json:"available,omitempty"`
}

// AddonProductStocksResponse is the response for addon product stocks.
type AddonProductStocksResponse struct {
	Stocks []AddonProductStock `json:"stocks"`
}

// AddonProductStocksUpdateRequest contains the data for updating addon product stocks.
type AddonProductStocksUpdateRequest struct {
	Stocks []AddonProductStock `json:"stocks"`
}

// ListAddonProducts retrieves a list of addon products.
func (c *Client) ListAddonProducts(ctx context.Context, opts *AddonProductsListOptions) (*AddonProductsListResponse, error) {
	path := "/addon_products"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("product_id", opts.ProductID).
			String("status", opts.Status).
			Build()
	}

	var resp AddonProductsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAddonProduct retrieves a single addon product by ID.
func (c *Client) GetAddonProduct(ctx context.Context, id string) (*AddonProduct, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("addon product id is required")
	}
	var addonProduct AddonProduct
	if err := c.Get(ctx, fmt.Sprintf("/addon_products/%s", id), &addonProduct); err != nil {
		return nil, err
	}
	return &addonProduct, nil
}

// CreateAddonProduct creates a new addon product.
func (c *Client) CreateAddonProduct(ctx context.Context, req *AddonProductCreateRequest) (*AddonProduct, error) {
	var addonProduct AddonProduct
	if err := c.Post(ctx, "/addon_products", req, &addonProduct); err != nil {
		return nil, err
	}
	return &addonProduct, nil
}

// DeleteAddonProduct deletes an addon product.
func (c *Client) DeleteAddonProduct(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("addon product id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/addon_products/%s", id))
}

// UpdateAddonProduct updates an existing addon product.
func (c *Client) UpdateAddonProduct(ctx context.Context, id string, req *AddonProductUpdateRequest) (*AddonProduct, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("addon product id is required")
	}
	var addonProduct AddonProduct
	if err := c.Put(ctx, fmt.Sprintf("/addon_products/%s", id), req, &addonProduct); err != nil {
		return nil, err
	}
	return &addonProduct, nil
}

// SearchAddonProducts searches for addon products based on criteria.
func (c *Client) SearchAddonProducts(ctx context.Context, opts *AddonProductSearchOptions) (*AddonProductsListResponse, error) {
	path := "/addon_products/search"
	if opts != nil {
		path += NewQuery().
			String("query", opts.Query).
			String("product_id", opts.ProductID).
			String("status", opts.Status).
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			Build()
	}

	var resp AddonProductsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateAddonProductQuantity updates the quantity of an addon product.
func (c *Client) UpdateAddonProductQuantity(ctx context.Context, id string, req *AddonProductQuantityRequest) (*AddonProduct, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("addon product id is required")
	}
	var addonProduct AddonProduct
	if err := c.Put(ctx, fmt.Sprintf("/addon_products/%s/update_quantity", id), req, &addonProduct); err != nil {
		return nil, err
	}
	return &addonProduct, nil
}

// UpdateAddonProductsQuantityBySKU updates the quantity of addon products by SKU (bulk operation).
func (c *Client) UpdateAddonProductsQuantityBySKU(ctx context.Context, req *AddonProductQuantityBySKURequest) (*AddonProduct, error) {
	var addonProduct AddonProduct
	if err := c.Put(ctx, "/addon_products/update_quantity", req, &addonProduct); err != nil {
		return nil, err
	}
	return &addonProduct, nil
}

// GetAddonProductStocks retrieves the stock information for an addon product.
func (c *Client) GetAddonProductStocks(ctx context.Context, id string) (*AddonProductStocksResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("addon product id is required")
	}
	var resp AddonProductStocksResponse
	if err := c.Get(ctx, fmt.Sprintf("/addon_products/%s/stocks", id), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateAddonProductStocks updates the stock information for an addon product.
func (c *Client) UpdateAddonProductStocks(ctx context.Context, id string, req *AddonProductStocksUpdateRequest) (*AddonProductStocksResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("addon product id is required")
	}
	var resp AddonProductStocksResponse
	if err := c.Put(ctx, fmt.Sprintf("/addon_products/%s/stocks", id), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
