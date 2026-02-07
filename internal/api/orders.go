package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// OrderSummary represents the fields returned by list/search endpoints.
// Use Order for order detail (which may include items and expanded resources).
type OrderSummary struct {
	ID            string    `json:"id"`
	OrderNumber   string    `json:"order_number"`
	Status        string    `json:"status"`
	PaymentStatus string    `json:"payment_status"`
	FulfillStatus string    `json:"fulfill_status"`
	TotalPrice    string    `json:"total_price"`
	Currency      string    `json:"currency"`
	CustomerEmail string    `json:"customer_email"`
	CustomerName  string    `json:"customer_name"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Order represents a Shopline order detail.
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
	// SubtotalItems is the API-native representation of order items (see docs).
	SubtotalItems []OrderSubtotalItem `json:"subtotal_items,omitempty"`
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

// UnmarshalJSON preserves normal unmarshaling and derives LineItems from SubtotalItems when needed.
func (o *Order) UnmarshalJSON(data []byte) error {
	type Alias Order
	aux := (*Alias)(o)
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	if o.LineItems == nil {
		o.LineItems = []OrderLineItem{}
	}

	// Some Shopline APIs return order items under `subtotal_items` rather than `line_items`.
	// Derive line_items for CLI convenience.
	if len(o.LineItems) == 0 && len(o.SubtotalItems) > 0 {
		derived := make([]OrderLineItem, 0, len(o.SubtotalItems))
		for _, si := range o.SubtotalItems {
			li := OrderLineItem{
				ID:        strings.TrimSpace(si.ID),
				ProductID: strings.TrimSpace(si.ItemID),
				Quantity:  si.Quantity,
			}
			if si.ItemVariationKey != "" {
				li.VariantID = strings.TrimSpace(si.ItemVariationKey)
			} else {
				li.VariantID = strings.TrimSpace(si.ItemVariationID)
			}

			if si.ItemPrice != nil {
				if b, err := json.Marshal(si.ItemPrice); err == nil {
					li.Price = b
					li.Currency = si.ItemPrice.CurrencyISO
				}
			} else if si.Price != nil {
				if b, err := json.Marshal(si.Price); err == nil {
					li.Price = b
					li.Currency = si.Price.CurrencyISO
				}
			}
			if si.TotalPrice != nil {
				if b, err := json.Marshal(si.TotalPrice); err == nil {
					li.Total = b
				}
			}

			if title, sku := extractTitleSKUFromItemData(si.ItemData); title != "" || sku != "" {
				if title != "" {
					li.Title = title
				}
				if sku != "" {
					li.SKU = sku
				}
			}
			derived = append(derived, li)
		}
		o.LineItems = derived
	}

	return nil
}

// OrderSubtotalItem represents an order item in the `subtotal_items` array.
// Shape is based on Shopline docs; keep permissive to avoid losing fields.
type OrderSubtotalItem struct {
	ID               string          `json:"id,omitempty"`
	ItemType         string          `json:"item_type,omitempty"`
	ItemID           string          `json:"item_id,omitempty"`
	ItemVariationID  string          `json:"item_variation_id,omitempty"`
	ItemVariationKey string          `json:"item_variation_key,omitempty"`
	Quantity         int             `json:"quantity,omitempty"`
	ItemPrice        *Price          `json:"item_price,omitempty"`
	Price            *Price          `json:"price,omitempty"`
	PriceSale        *Price          `json:"price_sale,omitempty"`
	DiscountedPrice  *Price          `json:"discounted_price,omitempty"`
	TotalPrice       *Price          `json:"total_price,omitempty"`
	ItemData         json.RawMessage `json:"item_data,omitempty"`
	ObjectData       json.RawMessage `json:"object_data,omitempty"`
}

func extractTitleSKUFromItemData(raw json.RawMessage) (string, string) {
	if len(raw) == 0 {
		return "", ""
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return "", ""
	}

	// Common paths seen in Shopline docs.
	if title := nestedString(m, "variation_data", "title"); title != "" {
		return title, nestedString(m, "variation_data", "sku")
	}
	if title := nestedString(m, "product_data", "title"); title != "" {
		return title, nestedString(m, "variation_data", "sku")
	}

	// Translation map fallbacks (e.g. title_translations.en).
	if title := nestedString(m, "variation_data", "title_translations", "en"); title != "" {
		return title, nestedString(m, "variation_data", "sku")
	}
	if title := nestedString(m, "product_data", "title_translations", "en"); title != "" {
		return title, nestedString(m, "variation_data", "sku")
	}

	return "", nestedString(m, "variation_data", "sku")
}

func nestedString(m map[string]any, path ...string) string {
	var cur any = m
	for _, p := range path {
		obj, ok := cur.(map[string]any)
		if !ok {
			return ""
		}
		cur, ok = obj[p]
		if !ok {
			return ""
		}
	}
	switch v := cur.(type) {
	case string:
		return v
	case map[string]any:
		// Try a few common translation keys.
		for _, k := range []string{"en", "en-US", "default"} {
			if s, ok := v[k].(string); ok {
				return s
			}
		}
		// As a last resort, return the first string value.
		for _, vv := range v {
			if s, ok := vv.(string); ok {
				return s
			}
		}
	}
	return ""
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
type OrdersListResponse = ListResponse[OrderSummary]

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
	// Docs: PATCH /orders/{orderId}/cancel
	return c.Patch(ctx, fmt.Sprintf("/orders/%s/cancel", id), nil, nil)
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
	Tags            []string    `json:"tags,omitempty"`
}

// OrderUpdateRequest contains the request body for updating an order.
type OrderUpdateRequest struct {
	Note            *string  `json:"note,omitempty"`
	Tags            []string `json:"tags,omitempty"`
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

// CreateArchivedOrdersReport creates an archived orders report.
// Docs: POST /orders/archived_orders
func (c *Client) CreateArchivedOrdersReport(ctx context.Context, body any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Post(ctx, "/orders/archived_orders", body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
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
	// Docs: PATCH /orders/{id}
	if err := c.Patch(ctx, fmt.Sprintf("/orders/%s", id), req, &order); err != nil {
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
	// Docs: PATCH /orders/{id}/order_delivery_status
	if err := c.Patch(ctx, fmt.Sprintf("/orders/%s/order_delivery_status", id), req, &order); err != nil {
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
	// Docs: PATCH /orders/{id}/order_payment_status
	if err := c.Patch(ctx, fmt.Sprintf("/orders/%s/order_payment_status", id), req, &order); err != nil {
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
	// Docs: PATCH /orders/execute_shipment
	if err := c.Patch(ctx, "/orders/execute_shipment", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ExecuteShipment executes shipment for a single order.
// Docs: PATCH /orders/{id}/execute_shipment
func (c *Client) ExecuteShipment(ctx context.Context, id string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	var resp json.RawMessage
	if err := c.Patch(ctx, fmt.Sprintf("/orders/%s/execute_shipment", id), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetOrderLabels retrieves delivery labels for order IDs.
// Docs: GET /orders/label
func (c *Client) GetOrderLabels(ctx context.Context, opts any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/orders/label", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// ListOrderTags retrieves all order tags.
// Docs: GET /orders/tags
func (c *Client) ListOrderTags(ctx context.Context) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/orders/tags", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetOrderTransactions retrieves order transaction info for order IDs.
// Docs: GET /orders/transactions
func (c *Client) GetOrderTransactions(ctx context.Context, opts any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/orders/transactions", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetOrderActionLogs retrieves action logs for an order.
// Docs: GET /orders/{id}/action_logs
func (c *Client) GetOrderActionLogs(ctx context.Context, id string) (json.RawMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/orders/%s/action_logs", id), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// PostOrderMessage creates an order message.
// Docs: POST /orders/{id}/messages
func (c *Client) PostOrderMessage(ctx context.Context, id string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	var resp json.RawMessage
	if err := c.Post(ctx, fmt.Sprintf("/orders/%s/messages", id), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
