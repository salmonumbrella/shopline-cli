package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// PurchaseOrder represents a Shopline purchase order.
type PurchaseOrder struct {
	ID            string              `json:"id"`
	Number        string              `json:"number"`
	Status        string              `json:"status"`
	SupplierID    string              `json:"supplier_id"`
	SupplierName  string              `json:"supplier_name"`
	WarehouseID   string              `json:"warehouse_id"`
	WarehouseName string              `json:"warehouse_name"`
	Currency      string              `json:"currency"`
	Subtotal      string              `json:"subtotal"`
	Tax           string              `json:"tax"`
	Total         string              `json:"total"`
	Note          string              `json:"note"`
	ExpectedAt    time.Time           `json:"expected_at"`
	ReceivedAt    time.Time           `json:"received_at"`
	LineItems     []PurchaseOrderItem `json:"line_items"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

// PurchaseOrderItem represents a line item in a purchase order.
type PurchaseOrderItem struct {
	ID          string `json:"id"`
	ProductID   string `json:"product_id"`
	VariantID   string `json:"variant_id"`
	SKU         string `json:"sku"`
	Title       string `json:"title"`
	Quantity    int    `json:"quantity"`
	ReceivedQty int    `json:"received_qty"`
	UnitCost    string `json:"unit_cost"`
	Total       string `json:"total"`
}

// PurchaseOrdersListOptions contains options for listing purchase orders.
type PurchaseOrdersListOptions struct {
	Page        int
	PageSize    int
	Status      string
	SupplierID  string
	WarehouseID string
}

// PurchaseOrdersListResponse contains the list response.
type PurchaseOrdersListResponse struct {
	Items      []PurchaseOrder `json:"items"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalCount int             `json:"total_count"`
	HasMore    bool            `json:"has_more"`
}

// PurchaseOrderCreateRequest contains the request body for creating a purchase order.
type PurchaseOrderCreateRequest struct {
	SupplierID  string                     `json:"supplier_id"`
	WarehouseID string                     `json:"warehouse_id"`
	Currency    string                     `json:"currency,omitempty"`
	Note        string                     `json:"note,omitempty"`
	ExpectedAt  *time.Time                 `json:"expected_at,omitempty"`
	LineItems   []PurchaseOrderItemRequest `json:"line_items"`
}

// PurchaseOrderItemRequest contains the request body for a line item.
type PurchaseOrderItemRequest struct {
	VariantID string `json:"variant_id"`
	Quantity  int    `json:"quantity"`
	UnitCost  string `json:"unit_cost,omitempty"`
}

// ListPurchaseOrders retrieves a list of purchase orders.
func (c *Client) ListPurchaseOrders(ctx context.Context, opts *PurchaseOrdersListOptions) (*PurchaseOrdersListResponse, error) {
	path := "/purchase_orders"
	if opts != nil {
		params := url.Values{}
		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			params.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.Status != "" {
			params.Set("status", opts.Status)
		}
		if opts.SupplierID != "" {
			params.Set("supplier_id", opts.SupplierID)
		}
		if opts.WarehouseID != "" {
			params.Set("warehouse_id", opts.WarehouseID)
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp PurchaseOrdersListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPurchaseOrder retrieves a single purchase order by ID.
func (c *Client) GetPurchaseOrder(ctx context.Context, id string) (*PurchaseOrder, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("purchase order id is required")
	}
	var po PurchaseOrder
	if err := c.Get(ctx, fmt.Sprintf("/purchase_orders/%s", id), &po); err != nil {
		return nil, err
	}
	return &po, nil
}

// CreatePurchaseOrder creates a new purchase order.
func (c *Client) CreatePurchaseOrder(ctx context.Context, req *PurchaseOrderCreateRequest) (*PurchaseOrder, error) {
	var po PurchaseOrder
	if err := c.Post(ctx, "/purchase_orders", req, &po); err != nil {
		return nil, err
	}
	return &po, nil
}

// ReceivePurchaseOrder marks a purchase order as received.
func (c *Client) ReceivePurchaseOrder(ctx context.Context, id string) (*PurchaseOrder, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("purchase order id is required")
	}
	var po PurchaseOrder
	if err := c.Post(ctx, fmt.Sprintf("/purchase_orders/%s/receive", id), nil, &po); err != nil {
		return nil, err
	}
	return &po, nil
}

// CancelPurchaseOrder cancels a purchase order.
func (c *Client) CancelPurchaseOrder(ctx context.Context, id string) (*PurchaseOrder, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("purchase order id is required")
	}
	var po PurchaseOrder
	if err := c.Post(ctx, fmt.Sprintf("/purchase_orders/%s/cancel", id), nil, &po); err != nil {
		return nil, err
	}
	return &po, nil
}

// DeletePurchaseOrder deletes a purchase order.
func (c *Client) DeletePurchaseOrder(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("purchase order id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/purchase_orders/%s", id))
}
