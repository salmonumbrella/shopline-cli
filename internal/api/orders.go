package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Order represents a Shopline order.
type Order struct {
	ID            string `json:"id"`
	OrderNumber   string `json:"order_number"`
	Status        string `json:"status"`
	PaymentStatus string `json:"payment_status"`
	FulfillStatus string `json:"fulfill_status"`
	TotalPrice    string `json:"total_price"`
	Currency      string `json:"currency"`
	CustomerEmail string `json:"customer_email"`
	CustomerName  string `json:"customer_name"`
	CustomerID    string `json:"customer_id,omitempty"`
	// Customer is populated when the API includes it or when expanded via the CLI.
	Customer *Customer `json:"customer,omitempty"`
	// LineItems are typically present on order detail endpoints.
	LineItems []OrderLineItem `json:"line_items"`
	// Common optional fields returned by the order detail endpoint.
	Note            string    `json:"note,omitempty"`
	Tags            []string  `json:"tags,omitempty"`
	ShippingAddress *Address  `json:"shipping_address,omitempty"`
	BillingAddress  *Address  `json:"billing_address,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// OrderLineItem represents a line item on an order (read side).
// Fields vary by endpoint; keep this permissive to avoid losing data.
type OrderLineItem struct {
	ID        string `json:"id,omitempty"`
	ProductID string `json:"product_id,omitempty"`
	VariantID string `json:"variant_id,omitempty"`
	SKU       string `json:"sku,omitempty"`
	Title     string `json:"title,omitempty"`
	Name      string `json:"name,omitempty"`
	Vendor    string `json:"vendor,omitempty"`
	Brand     string `json:"brand,omitempty"`
	Quantity  int    `json:"quantity,omitempty"`

	// These often vary in shape (number, string, or object). Preserve raw JSON.
	Price    json.RawMessage `json:"price,omitempty"`
	Currency string          `json:"currency,omitempty"`
	Total    json.RawMessage `json:"total,omitempty"`
	Subtotal json.RawMessage `json:"subtotal,omitempty"`
	Tax      json.RawMessage `json:"tax,omitempty"`
	Discount json.RawMessage `json:"discount,omitempty"`

	// Product is populated when expanded via the CLI.
	Product *Product `json:"product,omitempty"`
}

// OrdersListOptions contains options for listing orders.
type OrdersListOptions struct {
	Page      int
	PageSize  int
	Status    string
	Since     *time.Time
	Until     *time.Time
	SortBy    string
	SortOrder string
}

// OrdersListResponse is the paginated response for orders.
type OrdersListResponse = ListResponse[Order]

// ListOrders retrieves a list of orders.
func (c *Client) ListOrders(ctx context.Context, opts *OrdersListOptions) (*OrdersListResponse, error) {
	path := "/orders"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("status", opts.Status).
			Time("created_at_min", opts.Since).
			Time("created_at_max", opts.Until).
			String("sort_by", opts.SortBy).
			String("sort_order", opts.SortOrder).
			Build()
	}

	var resp OrdersListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetOrder retrieves a single order by ID.
func (c *Client) GetOrder(ctx context.Context, id string) (*Order, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	var order Order
	if err := c.Get(ctx, fmt.Sprintf("/orders/%s", id), &order); err != nil {
		return nil, err
	}
	return &order, nil
}

// CancelOrder cancels an order.
func (c *Client) CancelOrder(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("order id is required")
	}
	return c.Post(ctx, fmt.Sprintf("/orders/%s/cancel", id), nil, nil)
}

// OrderSearchOptions contains options for searching orders.
type OrderSearchOptions struct {
	Query    string
	Status   string
	Since    *time.Time
	Until    *time.Time
	Page     int
	PageSize int
}

// ArchivedOrdersListOptions contains options for listing archived orders.
type ArchivedOrdersListOptions struct {
	Page     int
	PageSize int
	Since    *time.Time
	Until    *time.Time
}

// OrderItem represents a line item in an order.
type OrderItem struct {
	ProductID   string  `json:"product_id"`
	VariationID string  `json:"variation_id,omitempty"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price,omitempty"`
}

// OrderCreateRequest contains the request body for creating an order.
type OrderCreateRequest struct {
	CustomerID      string      `json:"customer_id,omitempty"`
	CustomerEmail   string      `json:"customer_email,omitempty"`
	LineItems       []OrderItem `json:"line_items"`
	ShippingAddress *Address    `json:"shipping_address,omitempty"`
	BillingAddress  *Address    `json:"billing_address,omitempty"`
	Note            string      `json:"note,omitempty"`
}

// OrderUpdateRequest contains the request body for updating an order.
type OrderUpdateRequest struct {
	Note            *string  `json:"note,omitempty"`
	ShippingAddress *Address `json:"shipping_address,omitempty"`
	BillingAddress  *Address `json:"billing_address,omitempty"`
}

// OrderStatusUpdateRequest contains the request body for updating order status.
type OrderStatusUpdateRequest struct {
	Status string `json:"status"`
}

// OrderDeliveryStatusUpdateRequest contains the request body for updating delivery status.
type OrderDeliveryStatusUpdateRequest struct {
	DeliveryStatus string `json:"delivery_status"`
}

// OrderPaymentStatusUpdateRequest contains the request body for updating payment status.
type OrderPaymentStatusUpdateRequest struct {
	PaymentStatus string `json:"payment_status"`
}

// OrderTagsResponse represents order tags.
type OrderTagsResponse struct {
	Tags []string `json:"tags"`
}

// OrderTagsUpdateRequest contains the request body for updating order tags.
type OrderTagsUpdateRequest struct {
	Tags []string `json:"tags"`
}

// OrderSplitRequest contains the request body for splitting an order.
type OrderSplitRequest struct {
	LineItemIDs []string `json:"line_item_ids"`
}

// OrderSplitResponse represents the result of splitting an order.
type OrderSplitResponse struct {
	OriginalOrder *Order `json:"original_order"`
	NewOrder      *Order `json:"new_order"`
}

// BulkShipmentRequest contains the request body for bulk executing shipments.
type BulkShipmentRequest struct {
	OrderIDs []string `json:"order_ids"`
}

// BulkShipmentFailure represents a failed shipment in bulk execution.
type BulkShipmentFailure struct {
	OrderID string `json:"order_id"`
	Error   string `json:"error"`
}

// BulkShipmentResponse represents the result of bulk shipment execution.
type BulkShipmentResponse struct {
	Successful []string              `json:"successful"`
	Failed     []BulkShipmentFailure `json:"failed"`
}

// SearchOrders searches for orders with query parameters.
func (c *Client) SearchOrders(ctx context.Context, opts *OrderSearchOptions) (*OrdersListResponse, error) {
	path := "/orders/search" + NewQuery().
		String("query", opts.Query).
		String("status", opts.Status).
		Time("created_at_min", opts.Since).
		Time("created_at_max", opts.Until).
		Int("page", opts.Page).
		Int("page_size", opts.PageSize).
		Build()

	var resp OrdersListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListArchivedOrders retrieves archived orders.
func (c *Client) ListArchivedOrders(ctx context.Context, opts *ArchivedOrdersListOptions) (*OrdersListResponse, error) {
	path := "/orders/archived" + NewQuery().
		Int("page", opts.Page).
		Int("page_size", opts.PageSize).
		Time("created_at_min", opts.Since).
		Time("created_at_max", opts.Until).
		Build()

	var resp OrdersListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateOrder creates a new order.
func (c *Client) CreateOrder(ctx context.Context, req *OrderCreateRequest) (*Order, error) {
	var order Order
	if err := c.Post(ctx, "/orders", req, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

// UpdateOrder updates an existing order.
func (c *Client) UpdateOrder(ctx context.Context, id string, req *OrderUpdateRequest) (*Order, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	var order Order
	if err := c.Put(ctx, fmt.Sprintf("/orders/%s", id), req, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

// UpdateOrderStatus updates the status of an order.
func (c *Client) UpdateOrderStatus(ctx context.Context, id string, status string) (*Order, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	req := &OrderStatusUpdateRequest{Status: status}
	var order Order
	if err := c.Patch(ctx, fmt.Sprintf("/orders/%s/status", id), req, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

// UpdateOrderDeliveryStatus updates the delivery status of an order.
func (c *Client) UpdateOrderDeliveryStatus(ctx context.Context, id string, status string) (*Order, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	req := &OrderDeliveryStatusUpdateRequest{DeliveryStatus: status}
	var order Order
	if err := c.Patch(ctx, fmt.Sprintf("/orders/%s/delivery-status", id), req, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

// UpdateOrderPaymentStatus updates the payment status of an order.
func (c *Client) UpdateOrderPaymentStatus(ctx context.Context, id string, status string) (*Order, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	req := &OrderPaymentStatusUpdateRequest{PaymentStatus: status}
	var order Order
	if err := c.Patch(ctx, fmt.Sprintf("/orders/%s/payment-status", id), req, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

// GetOrderTags retrieves tags for an order.
func (c *Client) GetOrderTags(ctx context.Context, id string) (*OrderTagsResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	var tags OrderTagsResponse
	if err := c.Get(ctx, fmt.Sprintf("/orders/%s/tags", id), &tags); err != nil {
		return nil, err
	}
	return &tags, nil
}

// UpdateOrderTags updates tags for an order.
func (c *Client) UpdateOrderTags(ctx context.Context, id string, tags []string) (*Order, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	req := &OrderTagsUpdateRequest{Tags: tags}
	var order Order
	if err := c.Put(ctx, fmt.Sprintf("/orders/%s/tags", id), req, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

// SplitOrder splits an order into two orders.
func (c *Client) SplitOrder(ctx context.Context, id string, lineItemIDs []string) (*OrderSplitResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	if len(lineItemIDs) == 0 {
		return nil, fmt.Errorf("at least one line item id is required")
	}
	req := &OrderSplitRequest{LineItemIDs: lineItemIDs}
	var resp OrderSplitResponse
	if err := c.Post(ctx, fmt.Sprintf("/orders/%s/split", id), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// BulkExecuteShipment executes shipments for multiple orders.
func (c *Client) BulkExecuteShipment(ctx context.Context, orderIDs []string) (*BulkShipmentResponse, error) {
	if len(orderIDs) == 0 {
		return nil, fmt.Errorf("at least one order id is required")
	}
	req := &BulkShipmentRequest{OrderIDs: orderIDs}
	var resp BulkShipmentResponse
	if err := c.Post(ctx, "/orders/bulk-execute-shipment", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
