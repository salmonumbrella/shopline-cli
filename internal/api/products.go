package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Price represents a Shopline price object.
type Price struct {
	Cents          int     `json:"cents"`
	CurrencySymbol string  `json:"currency_symbol"`
	CurrencyISO    string  `json:"currency_iso"`
	Label          string  `json:"label"`
	Dollars        float64 `json:"dollars"`
}

// Product represents a Shopline product.
type Product struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Handle      string    `json:"handle"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Vendor      string    `json:"vendor"`
	ProductType string    `json:"product_type"`
	Tags        []string  `json:"tags"`
	Price       *Price    `json:"price"`
	Currency    string    `json:"currency"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// OpenAPI fields (preferred). Many endpoints use translations instead of plain strings.
	TitleTranslations       map[string]string `json:"title_translations,omitempty"`
	DescriptionTranslations map[string]string `json:"description_translations,omitempty"`
	Brand                   *string           `json:"brand,omitempty"`
}

// UnmarshalJSON derives legacy fields (Title/Description) from translation maps when present.
func (p *Product) UnmarshalJSON(data []byte) error {
	type Alias Product
	aux := (*Alias)(p)
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	if p.Title == "" && len(p.TitleTranslations) > 0 {
		p.Title = pickTranslation(p.TitleTranslations)
	}
	if p.Description == "" && len(p.DescriptionTranslations) > 0 {
		p.Description = pickTranslation(p.DescriptionTranslations)
	}

	return nil
}

func pickTranslation(m map[string]string) string {
	for _, k := range []string{"en", "en-US", "zh-hant", "zh-tw", "zh-cn"} {
		if v := strings.TrimSpace(m[k]); v != "" {
			return v
		}
	}
	for _, v := range m {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

// ProductsListOptions contains options for listing products.
type ProductsListOptions struct {
	Page        int
	PageSize    int
	Status      string
	Vendor      string
	ProductType string
	SortBy      string
	SortOrder   string
}

// ProductsListResponse is the paginated response for products.
type ProductsListResponse = ListResponse[Product]

// ListProducts retrieves a list of products.
func (c *Client) ListProducts(ctx context.Context, opts *ProductsListOptions) (*ProductsListResponse, error) {
	path := "/products"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("status", opts.Status).
			String("vendor", opts.Vendor).
			String("product_type", opts.ProductType).
			String("sort_by", opts.SortBy).
			String("sort_order", opts.SortOrder).
			Build()
	}

	var resp ProductsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetProduct retrieves a single product by ID.
func (c *Client) GetProduct(ctx context.Context, id string) (*Product, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	var product Product
	if err := c.Get(ctx, fmt.Sprintf("/products/%s", id), &product); err != nil {
		return nil, err
	}
	return &product, nil
}

// ProductSearchOptions contains options for searching products.
type ProductSearchOptions struct {
	Query    string
	Status   string
	Vendor   string
	Page     int
	PageSize int
}

// ProductCreateRequest contains the request body for creating a product.
type ProductCreateRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Vendor      string   `json:"vendor,omitempty"`
	ProductType string   `json:"product_type,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Status      string   `json:"status,omitempty"`
}

// ProductUpdateRequest contains the request body for updating a product.
type ProductUpdateRequest struct {
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	Vendor      *string  `json:"vendor,omitempty"`
	ProductType *string  `json:"product_type,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Status      *string  `json:"status,omitempty"`
}

// ProductQuantityUpdateRequest contains the request body for updating product quantity.
type ProductQuantityUpdateRequest struct {
	Quantity int `json:"quantity"`
}

// ProductPriceUpdateRequest contains the request body for updating product price.
type ProductPriceUpdateRequest struct {
	Price float64 `json:"price"`
}

// ProductQuantityBySKURequest contains the request body for updating quantity by SKU.
type ProductQuantityBySKURequest struct {
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
}

type ProductLockedInventoryCountRequest struct {
	ProductID string `json:"product_id"`
}

// ProductImage represents a product image.
type ProductImage struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	Position  int       `json:"position"`
	Src       string    `json:"src"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProductAddImagesRequest contains the request body for adding product images.
type ProductAddImagesRequest struct {
	Images []ProductImageInput `json:"images"`
}

// ProductImageInput represents an image to add to a product.
type ProductImageInput struct {
	Src      string `json:"src"`
	Position int    `json:"position,omitempty"`
}

// ProductDeleteImagesRequest contains the request body for deleting product images.
type ProductDeleteImagesRequest struct {
	ImageIDs []string `json:"image_ids"`
}

type ProductTagsReplaceRequest struct {
	Tags []string `json:"tags"`
}

type ProductBulkDeleteRequest struct {
	ProductIDs []string `json:"product_ids"`
}

// ProductVariation represents a product variation.
type ProductVariation struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	Title     string    `json:"title"`
	SKU       string    `json:"sku"`
	Price     *Price    `json:"price"`
	Quantity  int       `json:"quantity"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProductVariationCreateRequest contains the request body for creating a variation.
type ProductVariationCreateRequest struct {
	Title    string  `json:"title"`
	SKU      string  `json:"sku,omitempty"`
	Price    float64 `json:"price,omitempty"`
	Quantity int     `json:"quantity,omitempty"`
}

// ProductVariationUpdateRequest contains the request body for updating a variation.
type ProductVariationUpdateRequest struct {
	Title    *string  `json:"title,omitempty"`
	SKU      *string  `json:"sku,omitempty"`
	Price    *float64 `json:"price,omitempty"`
	Quantity *int     `json:"quantity,omitempty"`
}

// ProductTagsUpdateRequest contains the request body for updating product tags.
type ProductTagsUpdateRequest struct {
	Add    []string `json:"add,omitempty"`
	Remove []string `json:"remove,omitempty"`
}

// ProductCategoryAssignRequest contains the request body for bulk assigning products to category.
type ProductCategoryAssignRequest struct {
	ProductIDs []string `json:"product_ids"`
	CategoryID string   `json:"category_id"`
}

// LockedInventoryCount represents locked inventory information.
type LockedInventoryCount struct {
	ProductID   string `json:"product_id"`
	VariationID string `json:"variation_id,omitempty"`
	LockedCount int    `json:"locked_count"`
}

// SearchProductsPost searches for products via POST /products/search.
// This exists alongside the GET version since both are documented.
type ProductSearchRequest struct {
	Query    string `json:"query,omitempty"`
	Status   string `json:"status,omitempty"`
	Vendor   string `json:"vendor,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

// SearchProducts searches for products with query parameters.
func (c *Client) SearchProducts(ctx context.Context, opts *ProductSearchOptions) (*ProductsListResponse, error) {
	path := "/products/search" + NewQuery().
		String("query", opts.Query).
		String("status", opts.Status).
		String("vendor", opts.Vendor).
		Int("page", opts.Page).
		Int("page_size", opts.PageSize).
		Build()

	var resp ProductsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) SearchProductsPost(ctx context.Context, req *ProductSearchRequest) (*ProductsListResponse, error) {
	var resp ProductsListResponse
	if err := c.Post(ctx, "/products/search", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateProduct creates a new product.
func (c *Client) CreateProduct(ctx context.Context, req *ProductCreateRequest) (*Product, error) {
	var product Product
	if err := c.Post(ctx, "/products", req, &product); err != nil {
		return nil, err
	}
	return &product, nil
}

// UpdateProduct updates an existing product.
func (c *Client) UpdateProduct(ctx context.Context, id string, req *ProductUpdateRequest) (*Product, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	var product Product
	if err := c.Put(ctx, fmt.Sprintf("/products/%s", id), req, &product); err != nil {
		return nil, err
	}
	return &product, nil
}

// DeleteProduct deletes a product.
func (c *Client) DeleteProduct(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("product id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/products/%s", id))
}

// BulkDeleteProducts deletes multiple products by id.
func (c *Client) BulkDeleteProducts(ctx context.Context, productIDs []string) error {
	if len(productIDs) == 0 {
		return fmt.Errorf("at least one product id is required")
	}
	req := &ProductBulkDeleteRequest{ProductIDs: productIDs}
	return c.DeleteWithBody(ctx, "/products/bulk", req, nil)
}

// UpdateProductQuantity updates the quantity of a product.
func (c *Client) UpdateProductQuantity(ctx context.Context, id string, quantity int) (*Product, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	req := &ProductQuantityUpdateRequest{Quantity: quantity}
	var product Product
	if err := c.Put(ctx, fmt.Sprintf("/products/%s/update_quantity", id), req, &product); err != nil {
		return nil, err
	}
	return &product, nil
}

// UpdateProductVariationQuantity updates the quantity of a product variation.
func (c *Client) UpdateProductVariationQuantity(ctx context.Context, productID, variationID string, quantity int) (*Product, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	if strings.TrimSpace(variationID) == "" {
		return nil, fmt.Errorf("variation id is required")
	}
	req := &ProductQuantityUpdateRequest{Quantity: quantity}
	var product Product
	if err := c.Put(ctx, fmt.Sprintf("/products/%s/variations/%s/update_quantity", productID, variationID), req, &product); err != nil {
		return nil, err
	}
	return &product, nil
}

// UpdateProductQuantityBySKU updates product quantity by SKU.
func (c *Client) UpdateProductQuantityBySKU(ctx context.Context, sku string, quantity int) error {
	if strings.TrimSpace(sku) == "" {
		return fmt.Errorf("sku is required")
	}
	req := &ProductQuantityBySKURequest{SKU: sku, Quantity: quantity}
	return c.Put(ctx, "/products/update_quantity", req, nil)
}

// UpdateProductPrice updates the price of a product.
func (c *Client) UpdateProductPrice(ctx context.Context, id string, price float64) (*Product, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	req := &ProductPriceUpdateRequest{Price: price}
	var product Product
	if err := c.Put(ctx, fmt.Sprintf("/products/%s/update_price", id), req, &product); err != nil {
		return nil, err
	}
	return &product, nil
}

// UpdateProductVariationPrice updates the price of a product variation.
func (c *Client) UpdateProductVariationPrice(ctx context.Context, productID, variationID string, price float64) (*Product, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	if strings.TrimSpace(variationID) == "" {
		return nil, fmt.Errorf("variation id is required")
	}
	req := &ProductPriceUpdateRequest{Price: price}
	var product Product
	if err := c.Put(ctx, fmt.Sprintf("/products/%s/variations/%s/update_price", productID, variationID), req, &product); err != nil {
		return nil, err
	}
	return &product, nil
}

// AddProductImages adds images to a product.
func (c *Client) AddProductImages(ctx context.Context, productID string, req *ProductAddImagesRequest) ([]ProductImage, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	var images []ProductImage
	if err := c.Post(ctx, fmt.Sprintf("/products/%s/add_images", productID), req, &images); err != nil {
		return nil, err
	}
	return images, nil
}

// DeleteProductImages deletes images from a product.
func (c *Client) DeleteProductImages(ctx context.Context, productID string, imageIDs []string) error {
	if strings.TrimSpace(productID) == "" {
		return fmt.Errorf("product id is required")
	}
	if len(imageIDs) == 0 {
		return fmt.Errorf("at least one image id is required")
	}
	req := &ProductDeleteImagesRequest{ImageIDs: imageIDs}
	return c.DeleteWithBody(ctx, fmt.Sprintf("/products/%s/delete_images", productID), req, nil)
}

// ReplaceProductTags replaces the full tag list for a product.
func (c *Client) ReplaceProductTags(ctx context.Context, id string, tags []string) (*Product, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	req := &ProductTagsReplaceRequest{Tags: tags}
	var product Product
	if err := c.Put(ctx, fmt.Sprintf("/products/%s/tags", id), req, &product); err != nil {
		return nil, err
	}
	return &product, nil
}

// AddProductVariation adds a variation to a product.
func (c *Client) AddProductVariation(ctx context.Context, productID string, req *ProductVariationCreateRequest) (*ProductVariation, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	var variation ProductVariation
	if err := c.Post(ctx, fmt.Sprintf("/products/%s/variations", productID), req, &variation); err != nil {
		return nil, err
	}
	return &variation, nil
}

// UpdateProductVariation updates a product variation.
func (c *Client) UpdateProductVariation(ctx context.Context, productID, variationID string, req *ProductVariationUpdateRequest) (*ProductVariation, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	if strings.TrimSpace(variationID) == "" {
		return nil, fmt.Errorf("variation id is required")
	}
	var variation ProductVariation
	if err := c.Put(ctx, fmt.Sprintf("/products/%s/variations/%s", productID, variationID), req, &variation); err != nil {
		return nil, err
	}
	return &variation, nil
}

// DeleteProductVariation deletes a product variation.
func (c *Client) DeleteProductVariation(ctx context.Context, productID, variationID string) error {
	if strings.TrimSpace(productID) == "" {
		return fmt.Errorf("product id is required")
	}
	if strings.TrimSpace(variationID) == "" {
		return fmt.Errorf("variation id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/products/%s/variations/%s", productID, variationID))
}

// UpdateProductTags adds or removes tags from a product.
func (c *Client) UpdateProductTags(ctx context.Context, id string, req *ProductTagsUpdateRequest) (*Product, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	var product Product
	if err := c.Patch(ctx, fmt.Sprintf("/products/%s/tags", id), req, &product); err != nil {
		return nil, err
	}
	return &product, nil
}

// BulkAssignProductsToCategory assigns multiple products to a category.
func (c *Client) BulkAssignProductsToCategory(ctx context.Context, req *ProductCategoryAssignRequest) error {
	if len(req.ProductIDs) == 0 {
		return fmt.Errorf("at least one product id is required")
	}
	if strings.TrimSpace(req.CategoryID) == "" {
		return fmt.Errorf("category id is required")
	}
	// Per OpenAPI docs this is a Categories endpoint (bulk assign products to a category).
	return c.Post(ctx, "/categories/bulk_assign", req, nil)
}

// GetLockedInventoryCount retrieves locked inventory counts for a product.
func (c *Client) GetLockedInventoryCount(ctx context.Context, productID string) (*LockedInventoryCount, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	var count LockedInventoryCount
	req := &ProductLockedInventoryCountRequest{ProductID: productID}
	if err := c.Post(ctx, "/products/locked_inventory_count", req, &count); err != nil {
		return nil, err
	}
	return &count, nil
}

// GetProductPromotions retrieves promotions for a product.
func (c *Client) GetProductPromotions(ctx context.Context, productID string) (json.RawMessage, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/products/%s/promotions", productID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetProductStocks retrieves product stock info.
func (c *Client) GetProductStocks(ctx context.Context, productID string) (json.RawMessage, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/products/%s/stocks", productID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateProductStocks updates product stock info.
func (c *Client) UpdateProductStocks(ctx context.Context, productID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	var resp json.RawMessage
	if err := c.Put(ctx, fmt.Sprintf("/products/%s/stocks", productID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// BulkUpdateProductStocks updates stocks for multiple products.
func (c *Client) BulkUpdateProductStocks(ctx context.Context, body any) error {
	return c.Put(ctx, "/products/bulk_update_stocks", body, nil)
}

// UpdateProductsStatusBulk updates the online-store status for multiple products.
func (c *Client) UpdateProductsStatusBulk(ctx context.Context, body any) error {
	return c.Put(ctx, "/products/status/bulk", body, nil)
}

// UpdateProductsRetailStatusBulk updates the retail-store status for multiple products.
func (c *Client) UpdateProductsRetailStatusBulk(ctx context.Context, body any) error {
	return c.Put(ctx, "/products/retail_status/bulk", body, nil)
}

// UpdateProductsLabelsBulk updates product labels.
func (c *Client) UpdateProductsLabelsBulk(ctx context.Context, body any) error {
	return c.Patch(ctx, "/products/labels", body, nil)
}
