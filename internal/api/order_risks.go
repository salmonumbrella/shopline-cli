package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// OrderRisk represents a fraud risk assessment for an order.
type OrderRisk struct {
	ID             string    `json:"id"`
	OrderID        string    `json:"order_id"`
	Score          float64   `json:"score"`
	Recommendation string    `json:"recommendation"`
	Source         string    `json:"source"`
	Message        string    `json:"message"`
	Display        bool      `json:"display"`
	CauseCancel    bool      `json:"cause_cancel"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// OrderRisksListOptions contains options for listing order risks.
type OrderRisksListOptions struct {
	Page     int
	PageSize int
}

// OrderRisksListResponse is the paginated response for order risks.
type OrderRisksListResponse = ListResponse[OrderRisk]

// OrderRiskCreateRequest contains the request body for creating an order risk.
type OrderRiskCreateRequest struct {
	Score          float64 `json:"score"`
	Recommendation string  `json:"recommendation"`
	Source         string  `json:"source,omitempty"`
	Message        string  `json:"message,omitempty"`
	Display        bool    `json:"display,omitempty"`
	CauseCancel    bool    `json:"cause_cancel,omitempty"`
}

// OrderRiskUpdateRequest contains the request body for updating an order risk.
type OrderRiskUpdateRequest struct {
	Score          *float64 `json:"score,omitempty"`
	Recommendation *string  `json:"recommendation,omitempty"`
	Message        *string  `json:"message,omitempty"`
	Display        *bool    `json:"display,omitempty"`
	CauseCancel    *bool    `json:"cause_cancel,omitempty"`
}

// ListOrderRisks retrieves a list of risks for an order.
func (c *Client) ListOrderRisks(ctx context.Context, orderID string, opts *OrderRisksListOptions) (*OrderRisksListResponse, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("order id is required")
	}

	path := fmt.Sprintf("/orders/%s/risks", orderID)
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			Build()
	}

	var resp OrderRisksListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetOrderRisk retrieves a single order risk by ID.
func (c *Client) GetOrderRisk(ctx context.Context, orderID, riskID string) (*OrderRisk, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	if strings.TrimSpace(riskID) == "" {
		return nil, fmt.Errorf("risk id is required")
	}

	var risk OrderRisk
	if err := c.Get(ctx, fmt.Sprintf("/orders/%s/risks/%s", orderID, riskID), &risk); err != nil {
		return nil, err
	}
	return &risk, nil
}

// CreateOrderRisk creates a new risk assessment for an order.
func (c *Client) CreateOrderRisk(ctx context.Context, orderID string, req *OrderRiskCreateRequest) (*OrderRisk, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("order id is required")
	}

	var risk OrderRisk
	if err := c.Post(ctx, fmt.Sprintf("/orders/%s/risks", orderID), req, &risk); err != nil {
		return nil, err
	}
	return &risk, nil
}

// UpdateOrderRisk updates an existing order risk.
func (c *Client) UpdateOrderRisk(ctx context.Context, orderID, riskID string, req *OrderRiskUpdateRequest) (*OrderRisk, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	if strings.TrimSpace(riskID) == "" {
		return nil, fmt.Errorf("risk id is required")
	}

	var risk OrderRisk
	if err := c.Put(ctx, fmt.Sprintf("/orders/%s/risks/%s", orderID, riskID), req, &risk); err != nil {
		return nil, err
	}
	return &risk, nil
}

// DeleteOrderRisk deletes an order risk.
func (c *Client) DeleteOrderRisk(ctx context.Context, orderID, riskID string) error {
	if strings.TrimSpace(orderID) == "" {
		return fmt.Errorf("order id is required")
	}
	if strings.TrimSpace(riskID) == "" {
		return fmt.Errorf("risk id is required")
	}

	return c.Delete(ctx, fmt.Sprintf("/orders/%s/risks/%s", orderID, riskID))
}
