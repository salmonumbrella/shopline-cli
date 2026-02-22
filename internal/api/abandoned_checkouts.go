package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// AbandonedCheckoutLineItem represents a line item in an abandoned checkout.
type AbandonedCheckoutLineItem struct {
	VariantID   string  `json:"variant_id"`
	ProductID   string  `json:"product_id"`
	Title       string  `json:"title"`
	VariantName string  `json:"variant_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

// AbandonedCheckout represents a Shopline abandoned checkout.
type AbandonedCheckout struct {
	ID                     string                      `json:"id"`
	Token                  string                      `json:"token"`
	CartToken              string                      `json:"cart_token"`
	Email                  string                      `json:"email"`
	CustomerID             string                      `json:"customer_id"`
	CustomerLocale         string                      `json:"customer_locale"`
	Phone                  string                      `json:"phone"`
	TotalPrice             string                      `json:"total_price"`
	SubtotalPrice          string                      `json:"subtotal_price"`
	TotalTax               string                      `json:"total_tax"`
	TotalDiscounts         string                      `json:"total_discounts"`
	Currency               string                      `json:"currency"`
	LineItems              []AbandonedCheckoutLineItem `json:"line_items"`
	AbandonedCheckoutURL   string                      `json:"abandoned_checkout_url"`
	RecoveryURL            string                      `json:"recovery_url"`
	CompletedAt            *time.Time                  `json:"completed_at"`
	ClosedAt               *time.Time                  `json:"closed_at"`
	RecoveryEmailSentCount int                         `json:"recovery_email_sent_count"`
	CreatedAt              time.Time                   `json:"created_at"`
	UpdatedAt              time.Time                   `json:"updated_at"`
}

// AbandonedCheckoutsListOptions contains options for listing abandoned checkouts.
type AbandonedCheckoutsListOptions struct {
	Page       int
	PageSize   int
	Status     string
	CustomerID string
	Since      *time.Time
	Until      *time.Time
}

// AbandonedCheckoutsListResponse is the paginated response for abandoned checkouts.
type AbandonedCheckoutsListResponse = ListResponse[AbandonedCheckout]

// ListAbandonedCheckouts retrieves a list of abandoned checkouts.
func (c *Client) ListAbandonedCheckouts(ctx context.Context, opts *AbandonedCheckoutsListOptions) (*AbandonedCheckoutsListResponse, error) {
	path := "/abandoned_checkouts"
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

	var resp AbandonedCheckoutsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAbandonedCheckout retrieves a single abandoned checkout by ID.
func (c *Client) GetAbandonedCheckout(ctx context.Context, id string) (*AbandonedCheckout, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("abandoned checkout id is required")
	}
	var checkout AbandonedCheckout
	if err := c.Get(ctx, fmt.Sprintf("/abandoned_checkouts/%s", id), &checkout); err != nil {
		return nil, err
	}
	return &checkout, nil
}

// SendAbandonedCheckoutRecoveryEmail sends a recovery email for an abandoned checkout.
func (c *Client) SendAbandonedCheckoutRecoveryEmail(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("abandoned checkout id is required")
	}
	return c.Post(ctx, fmt.Sprintf("/abandoned_checkouts/%s/send_recovery_email", id), nil, nil)
}
