package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// POSPurchaseOrdersListOptions contains options for listing POS purchase orders.
type POSPurchaseOrdersListOptions struct {
	Page     int
	PageSize int
}

// ListPOSPurchaseOrders lists POS purchase orders (documented endpoint; raw JSON).
//
// Docs: GET /pos/purchase_orders
func (c *Client) ListPOSPurchaseOrders(ctx context.Context, opts *POSPurchaseOrdersListOptions) (json.RawMessage, error) {
	path := "/pos/purchase_orders"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			Build()
	}

	var resp json.RawMessage
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// CreatePOSPurchaseOrder creates a POS purchase order (documented endpoint; raw JSON body).
//
// Docs: POST /pos/purchase_orders
func (c *Client) CreatePOSPurchaseOrder(ctx context.Context, body any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Post(ctx, "/pos/purchase_orders", body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetPOSPurchaseOrder retrieves a POS purchase order by ID (documented endpoint; raw JSON).
//
// Docs: GET /pos/purchase_orders/{purchaseOrderId}
func (c *Client) GetPOSPurchaseOrder(ctx context.Context, id string) (json.RawMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("pos purchase order id is required")
	}

	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/pos/purchase_orders/%s", id), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdatePOSPurchaseOrder updates a POS purchase order (documented endpoint; raw JSON body).
//
// Docs: PUT /pos/purchase_orders/{purchaseOrderId}
func (c *Client) UpdatePOSPurchaseOrder(ctx context.Context, id string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("pos purchase order id is required")
	}

	var resp json.RawMessage
	if err := c.Put(ctx, fmt.Sprintf("/pos/purchase_orders/%s", id), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// BulkDeletePOSPurchaseOrders deletes POS purchase orders (documented endpoint; raw JSON body).
//
// Docs: PUT /pos/purchase_orders/bulk_delete
func (c *Client) BulkDeletePOSPurchaseOrders(ctx context.Context, body any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Put(ctx, "/pos/purchase_orders/bulk_delete", body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// CreatePOSPurchaseOrderChild creates a child POS purchase order (documented endpoint; raw JSON body).
//
// Docs: POST /pos/purchase_orders/{purchaseOrderId}/child
func (c *Client) CreatePOSPurchaseOrderChild(ctx context.Context, id string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("pos purchase order id is required")
	}

	var resp json.RawMessage
	if err := c.Post(ctx, fmt.Sprintf("/pos/purchase_orders/%s/child", id), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
