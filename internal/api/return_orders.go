package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ReturnOrderLineItem represents a line item in a return order.
type ReturnOrderLineItem struct {
	LineItemID   string `json:"line_item_id"`
	VariantID    string `json:"variant_id"`
	ProductID    string `json:"product_id"`
	Title        string `json:"title"`
	Quantity     int    `json:"quantity"`
	ReturnReason string `json:"return_reason"`
}

// ReturnOrder represents a Shopline return order.
type ReturnOrder struct {
	ID              string                `json:"id"`
	OrderID         string                `json:"order_id"`
	OrderNumber     string                `json:"order_number"`
	Status          string                `json:"status"`
	ReturnType      string                `json:"return_type"`
	CustomerID      string                `json:"customer_id"`
	CustomerEmail   string                `json:"customer_email"`
	TotalAmount     string                `json:"total_amount"`
	RefundAmount    string                `json:"refund_amount"`
	Currency        string                `json:"currency"`
	Reason          string                `json:"reason"`
	Note            string                `json:"note"`
	LineItems       []ReturnOrderLineItem `json:"line_items"`
	TrackingNumber  string                `json:"tracking_number"`
	TrackingCompany string                `json:"tracking_company"`
	ReceivedAt      *time.Time            `json:"received_at"`
	CompletedAt     *time.Time            `json:"completed_at"`
	CancelledAt     *time.Time            `json:"cancelled_at"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
}

// ReturnOrdersListOptions contains options for listing return orders.
type ReturnOrdersListOptions struct {
	Page       int
	PageSize   int
	Status     string
	OrderID    string
	CustomerID string
	ReturnType string
	Since      *time.Time
	Until      *time.Time
}

// ReturnOrdersListResponse is the paginated response for return orders.
type ReturnOrdersListResponse = ListResponse[ReturnOrder]

// ReturnOrderCreateRequest contains the request body for creating a return order.
type ReturnOrderCreateRequest struct {
	OrderID   string                `json:"order_id"`
	Reason    string                `json:"reason,omitempty"`
	Note      string                `json:"note,omitempty"`
	LineItems []ReturnOrderLineItem `json:"line_items"`
}

// ReturnOrderUpdateRequest contains the request body for updating a return order.
type ReturnOrderUpdateRequest struct {
	Status          *string `json:"status,omitempty"`
	TrackingNumber  *string `json:"tracking_number,omitempty"`
	TrackingCompany *string `json:"tracking_company,omitempty"`
	Note            *string `json:"note,omitempty"`
}

// ListReturnOrders retrieves a list of return orders.
func (c *Client) ListReturnOrders(ctx context.Context, opts *ReturnOrdersListOptions) (*ReturnOrdersListResponse, error) {
	path := "/return_orders"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("status", opts.Status).
			String("order_id", opts.OrderID).
			String("customer_id", opts.CustomerID).
			String("return_type", opts.ReturnType).
			Time("created_at_min", opts.Since).
			Time("created_at_max", opts.Until).
			Build()
	}

	var resp ReturnOrdersListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetReturnOrder retrieves a single return order by ID.
func (c *Client) GetReturnOrder(ctx context.Context, id string) (*ReturnOrder, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("return order id is required")
	}
	var returnOrder ReturnOrder
	if err := c.Get(ctx, fmt.Sprintf("/return_orders/%s", id), &returnOrder); err != nil {
		return nil, err
	}
	return &returnOrder, nil
}

// CreateReturnOrder creates a new return order.
func (c *Client) CreateReturnOrder(ctx context.Context, req *ReturnOrderCreateRequest) (*ReturnOrder, error) {
	var returnOrder ReturnOrder
	if err := c.Post(ctx, "/return_orders", req, &returnOrder); err != nil {
		return nil, err
	}
	return &returnOrder, nil
}

// UpdateReturnOrder updates an existing return order.
func (c *Client) UpdateReturnOrder(ctx context.Context, id string, req *ReturnOrderUpdateRequest) (*ReturnOrder, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("return order id is required")
	}
	var returnOrder ReturnOrder
	if err := c.Put(ctx, fmt.Sprintf("/return_orders/%s", id), req, &returnOrder); err != nil {
		return nil, err
	}
	return &returnOrder, nil
}

// CancelReturnOrder cancels a return order.
func (c *Client) CancelReturnOrder(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("return order id is required")
	}
	return c.Post(ctx, fmt.Sprintf("/return_orders/%s/cancel", id), nil, nil)
}

// CompleteReturnOrder marks a return order as complete.
func (c *Client) CompleteReturnOrder(ctx context.Context, id string) (*ReturnOrder, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("return order id is required")
	}
	var returnOrder ReturnOrder
	if err := c.Post(ctx, fmt.Sprintf("/return_orders/%s/complete", id), nil, &returnOrder); err != nil {
		return nil, err
	}
	return &returnOrder, nil
}

// ReceiveReturnOrder marks items as received for a return order.
func (c *Client) ReceiveReturnOrder(ctx context.Context, id string) (*ReturnOrder, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("return order id is required")
	}
	var returnOrder ReturnOrder
	if err := c.Post(ctx, fmt.Sprintf("/return_orders/%s/receive", id), nil, &returnOrder); err != nil {
		return nil, err
	}
	return &returnOrder, nil
}
