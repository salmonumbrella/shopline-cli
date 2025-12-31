package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// RefundLineItem represents a line item in a refund.
type RefundLineItem struct {
	LineItemID  string  `json:"line_item_id"`
	Quantity    int     `json:"quantity"`
	RestockType string  `json:"restock_type"`
	Subtotal    float64 `json:"subtotal"`
}

// Refund represents a Shopline refund.
type Refund struct {
	ID          string           `json:"id"`
	OrderID     string           `json:"order_id"`
	Note        string           `json:"note"`
	Restock     bool             `json:"restock"`
	Amount      string           `json:"amount"`
	Currency    string           `json:"currency"`
	Status      string           `json:"status"`
	LineItems   []RefundLineItem `json:"line_items"`
	ProcessedAt time.Time        `json:"processed_at"`
	CreatedAt   time.Time        `json:"created_at"`
}

// RefundsListOptions contains options for listing refunds.
type RefundsListOptions struct {
	Page     int
	PageSize int
	OrderID  string
	Status   string
}

// RefundsListResponse is the paginated response for refunds.
type RefundsListResponse = ListResponse[Refund]

// RefundCreateRequest contains the request body for creating a refund.
type RefundCreateRequest struct {
	OrderID   string           `json:"order_id"`
	Note      string           `json:"note,omitempty"`
	Restock   bool             `json:"restock,omitempty"`
	Amount    float64          `json:"amount,omitempty"`
	LineItems []RefundLineItem `json:"line_items,omitempty"`
}

// ListRefunds retrieves a list of refunds.
func (c *Client) ListRefunds(ctx context.Context, opts *RefundsListOptions) (*RefundsListResponse, error) {
	path := "/refunds"
	if opts != nil {
		params := url.Values{}
		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			params.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.OrderID != "" {
			params.Set("order_id", opts.OrderID)
		}
		if opts.Status != "" {
			params.Set("status", opts.Status)
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp RefundsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetRefund retrieves a single refund by ID.
func (c *Client) GetRefund(ctx context.Context, id string) (*Refund, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("refund id is required")
	}
	var refund Refund
	if err := c.Get(ctx, fmt.Sprintf("/refunds/%s", id), &refund); err != nil {
		return nil, err
	}
	return &refund, nil
}

// CreateRefund creates a new refund.
func (c *Client) CreateRefund(ctx context.Context, req *RefundCreateRequest) (*Refund, error) {
	var refund Refund
	if err := c.Post(ctx, "/refunds", req, &refund); err != nil {
		return nil, err
	}
	return &refund, nil
}

// ListOrderRefunds retrieves refunds for a specific order.
func (c *Client) ListOrderRefunds(ctx context.Context, orderID string) (*RefundsListResponse, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	var resp RefundsListResponse
	if err := c.Get(ctx, fmt.Sprintf("/orders/%s/refunds", orderID), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
