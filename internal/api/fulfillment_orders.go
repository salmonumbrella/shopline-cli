package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// FulfillmentOrder represents a Shopline fulfillment order.
type FulfillmentOrder struct {
	ID                 string                    `json:"id"`
	OrderID            string                    `json:"order_id"`
	Status             string                    `json:"status"`
	FulfillmentStatus  string                    `json:"fulfillment_status"`
	AssignedLocationID string                    `json:"assigned_location_id"`
	RequestStatus      string                    `json:"request_status"`
	LineItems          []FulfillmentOrderItem    `json:"line_items"`
	DeliveryMethod     FulfillmentDeliveryMethod `json:"delivery_method"`
	CreatedAt          time.Time                 `json:"created_at"`
	UpdatedAt          time.Time                 `json:"updated_at"`
}

// FulfillmentOrderItem represents a line item in a fulfillment order.
type FulfillmentOrderItem struct {
	ID                  string `json:"id"`
	LineItemID          string `json:"line_item_id"`
	VariantID           string `json:"variant_id"`
	Quantity            int    `json:"quantity"`
	FulfillableQuantity int    `json:"fulfillable_quantity"`
	FulfilledQuantity   int    `json:"fulfilled_quantity"`
}

// FulfillmentDeliveryMethod represents the delivery method for a fulfillment order.
type FulfillmentDeliveryMethod struct {
	MethodType  string `json:"method_type"`
	ServiceCode string `json:"service_code"`
}

// FulfillmentOrdersListOptions contains options for listing fulfillment orders.
type FulfillmentOrdersListOptions struct {
	Page     int
	PageSize int
	Status   string
	OrderID  string
}

// FulfillmentOrdersListResponse contains the list response.
type FulfillmentOrdersListResponse struct {
	Items      []FulfillmentOrder `json:"items"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalCount int                `json:"total_count"`
	HasMore    bool               `json:"has_more"`
}

// FulfillmentOrderMoveRequest contains the request body for moving a fulfillment order.
type FulfillmentOrderMoveRequest struct {
	NewLocationID string `json:"new_location_id"`
}

// ListFulfillmentOrders retrieves a list of fulfillment orders.
func (c *Client) ListFulfillmentOrders(ctx context.Context, opts *FulfillmentOrdersListOptions) (*FulfillmentOrdersListResponse, error) {
	path := "/fulfillment_orders"
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
		if opts.OrderID != "" {
			params.Set("order_id", opts.OrderID)
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp FulfillmentOrdersListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetFulfillmentOrder retrieves a single fulfillment order by ID.
func (c *Client) GetFulfillmentOrder(ctx context.Context, id string) (*FulfillmentOrder, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("fulfillment order id is required")
	}
	var fo FulfillmentOrder
	if err := c.Get(ctx, fmt.Sprintf("/fulfillment_orders/%s", id), &fo); err != nil {
		return nil, err
	}
	return &fo, nil
}

// ListOrderFulfillmentOrders retrieves fulfillment orders for a specific order.
func (c *Client) ListOrderFulfillmentOrders(ctx context.Context, orderID string) (*FulfillmentOrdersListResponse, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	var resp FulfillmentOrdersListResponse
	if err := c.Get(ctx, fmt.Sprintf("/orders/%s/fulfillment_orders", orderID), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// MoveFulfillmentOrder moves a fulfillment order to a new location.
func (c *Client) MoveFulfillmentOrder(ctx context.Context, id string, newLocationID string) (*FulfillmentOrder, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("fulfillment order id is required")
	}
	if strings.TrimSpace(newLocationID) == "" {
		return nil, fmt.Errorf("new location id is required")
	}
	req := &FulfillmentOrderMoveRequest{NewLocationID: newLocationID}
	var fo FulfillmentOrder
	if err := c.Post(ctx, fmt.Sprintf("/fulfillment_orders/%s/move", id), req, &fo); err != nil {
		return nil, err
	}
	return &fo, nil
}

// CancelFulfillmentOrder cancels a fulfillment order.
func (c *Client) CancelFulfillmentOrder(ctx context.Context, id string) (*FulfillmentOrder, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("fulfillment order id is required")
	}
	var fo FulfillmentOrder
	if err := c.Post(ctx, fmt.Sprintf("/fulfillment_orders/%s/cancel", id), nil, &fo); err != nil {
		return nil, err
	}
	return &fo, nil
}

// CloseFulfillmentOrder closes a fulfillment order.
func (c *Client) CloseFulfillmentOrder(ctx context.Context, id string) (*FulfillmentOrder, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("fulfillment order id is required")
	}
	var fo FulfillmentOrder
	if err := c.Post(ctx, fmt.Sprintf("/fulfillment_orders/%s/close", id), nil, &fo); err != nil {
		return nil, err
	}
	return &fo, nil
}
