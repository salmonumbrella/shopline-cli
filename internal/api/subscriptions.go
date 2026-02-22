package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// SubscriptionStatus represents the status of a subscription.
type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusPaused    SubscriptionStatus = "paused"
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"
	SubscriptionStatusExpired   SubscriptionStatus = "expired"
)

// Subscription represents a customer subscription.
type Subscription struct {
	ID            string             `json:"id"`
	CustomerID    string             `json:"customer_id"`
	ProductID     string             `json:"product_id"`
	VariantID     string             `json:"variant_id"`
	Status        SubscriptionStatus `json:"status"`
	Interval      string             `json:"interval"`
	IntervalCount int                `json:"interval_count"`
	Price         string             `json:"price"`
	Currency      string             `json:"currency"`
	NextBillingAt time.Time          `json:"next_billing_at"`
	CancelledAt   time.Time          `json:"cancelled_at"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

// SubscriptionsListOptions contains options for listing subscriptions.
type SubscriptionsListOptions struct {
	Page       int
	PageSize   int
	CustomerID string
	ProductID  string
	Status     string
}

// SubscriptionsListResponse is the paginated response for subscriptions.
type SubscriptionsListResponse = ListResponse[Subscription]

// SubscriptionCreateRequest contains the data for creating a subscription.
type SubscriptionCreateRequest struct {
	CustomerID    string `json:"customer_id"`
	ProductID     string `json:"product_id"`
	VariantID     string `json:"variant_id,omitempty"`
	Interval      string `json:"interval"`
	IntervalCount int    `json:"interval_count,omitempty"`
}

// ListSubscriptions retrieves a list of subscriptions.
func (c *Client) ListSubscriptions(ctx context.Context, opts *SubscriptionsListOptions) (*SubscriptionsListResponse, error) {
	path := "/subscriptions"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("customer_id", opts.CustomerID).
			String("product_id", opts.ProductID).
			String("status", opts.Status).
			Build()
	}

	var resp SubscriptionsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetSubscription retrieves a single subscription by ID.
func (c *Client) GetSubscription(ctx context.Context, id string) (*Subscription, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("subscription id is required")
	}
	var subscription Subscription
	if err := c.Get(ctx, fmt.Sprintf("/subscriptions/%s", id), &subscription); err != nil {
		return nil, err
	}
	return &subscription, nil
}

// CreateSubscription creates a new subscription.
func (c *Client) CreateSubscription(ctx context.Context, req *SubscriptionCreateRequest) (*Subscription, error) {
	var subscription Subscription
	if err := c.Post(ctx, "/subscriptions", req, &subscription); err != nil {
		return nil, err
	}
	return &subscription, nil
}

// DeleteSubscription cancels a subscription.
func (c *Client) DeleteSubscription(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("subscription id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/subscriptions/%s", id))
}
