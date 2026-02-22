package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// DraftOrderLineItem represents a line item in a draft order.
type DraftOrderLineItem struct {
	VariantID string  `json:"variant_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	Title     string  `json:"title"`
}

// DraftOrder represents a Shopline draft order.
type DraftOrder struct {
	ID            string               `json:"id"`
	Name          string               `json:"name"`
	Status        string               `json:"status"`
	CustomerID    string               `json:"customer_id"`
	CustomerEmail string               `json:"customer_email"`
	TotalPrice    string               `json:"total_price"`
	SubtotalPrice string               `json:"subtotal_price"`
	TotalTax      string               `json:"total_tax"`
	Currency      string               `json:"currency"`
	Note          string               `json:"note"`
	LineItems     []DraftOrderLineItem `json:"line_items"`
	InvoiceURL    string               `json:"invoice_url"`
	InvoiceSentAt *time.Time           `json:"invoice_sent_at"`
	CompletedAt   *time.Time           `json:"completed_at"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

// DraftOrdersListOptions contains options for listing draft orders.
type DraftOrdersListOptions struct {
	Page       int
	PageSize   int
	Status     string
	CustomerID string
	Since      *time.Time
	Until      *time.Time
}

// DraftOrdersListResponse is the paginated response for draft orders.
type DraftOrdersListResponse = ListResponse[DraftOrder]

// DraftOrderCreateRequest contains the request body for creating a draft order.
type DraftOrderCreateRequest struct {
	CustomerID string               `json:"customer_id,omitempty"`
	Email      string               `json:"email,omitempty"`
	Note       string               `json:"note,omitempty"`
	LineItems  []DraftOrderLineItem `json:"line_items"`
}

// ListDraftOrders retrieves a list of draft orders.
func (c *Client) ListDraftOrders(ctx context.Context, opts *DraftOrdersListOptions) (*DraftOrdersListResponse, error) {
	path := "/draft_orders"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("status", opts.Status).
			String("customer_id", opts.CustomerID).
			Time("created_at_min", opts.Since).
			Time("created_at_max", opts.Until).
			Build()
	}

	var resp DraftOrdersListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetDraftOrder retrieves a single draft order by ID.
func (c *Client) GetDraftOrder(ctx context.Context, id string) (*DraftOrder, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("draft order id is required")
	}
	var draftOrder DraftOrder
	if err := c.Get(ctx, fmt.Sprintf("/draft_orders/%s", id), &draftOrder); err != nil {
		return nil, err
	}
	return &draftOrder, nil
}

// CreateDraftOrder creates a new draft order.
func (c *Client) CreateDraftOrder(ctx context.Context, req *DraftOrderCreateRequest) (*DraftOrder, error) {
	var draftOrder DraftOrder
	if err := c.Post(ctx, "/draft_orders", req, &draftOrder); err != nil {
		return nil, err
	}
	return &draftOrder, nil
}

// DeleteDraftOrder deletes a draft order.
func (c *Client) DeleteDraftOrder(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("draft order id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/draft_orders/%s", id))
}

// CompleteDraftOrder converts a draft order to a real order.
func (c *Client) CompleteDraftOrder(ctx context.Context, id string) (*DraftOrder, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("draft order id is required")
	}
	var draftOrder DraftOrder
	if err := c.Put(ctx, fmt.Sprintf("/draft_orders/%s/complete", id), nil, &draftOrder); err != nil {
		return nil, err
	}
	return &draftOrder, nil
}

// SendDraftOrderInvoice sends an invoice for a draft order.
func (c *Client) SendDraftOrderInvoice(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("draft order id is required")
	}
	return c.Post(ctx, fmt.Sprintf("/draft_orders/%s/send_invoice", id), nil, nil)
}
