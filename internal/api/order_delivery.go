package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// OrderDelivery represents order delivery information.
type OrderDelivery struct {
	ID             string     `json:"id"`
	OrderID        string     `json:"order_id"`
	Status         string     `json:"status"`
	TrackingNumber string     `json:"tracking_number"`
	TrackingURL    string     `json:"tracking_url"`
	Carrier        string     `json:"carrier"`
	EstimatedAt    *time.Time `json:"estimated_at"`
	ShippedAt      *time.Time `json:"shipped_at"`
	DeliveredAt    *time.Time `json:"delivered_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// OrderDeliveryUpdateRequest contains the request body for updating order delivery.
type OrderDeliveryUpdateRequest struct {
	Status         *string    `json:"status,omitempty"`
	TrackingNumber *string    `json:"tracking_number,omitempty"`
	TrackingURL    *string    `json:"tracking_url,omitempty"`
	Carrier        *string    `json:"carrier,omitempty"`
	EstimatedAt    *time.Time `json:"estimated_at,omitempty"`
}

// GetOrderDelivery retrieves delivery information for an order.
func (c *Client) GetOrderDelivery(ctx context.Context, orderID string) (*OrderDelivery, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	var delivery OrderDelivery
	if err := c.Get(ctx, fmt.Sprintf("/orders/%s/delivery", orderID), &delivery); err != nil {
		return nil, err
	}
	return &delivery, nil
}

// UpdateOrderDelivery updates delivery information for an order.
func (c *Client) UpdateOrderDelivery(ctx context.Context, orderID string, req *OrderDeliveryUpdateRequest) (*OrderDelivery, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	var delivery OrderDelivery
	if err := c.Put(ctx, fmt.Sprintf("/orders/%s/delivery", orderID), req, &delivery); err != nil {
		return nil, err
	}
	return &delivery, nil
}
