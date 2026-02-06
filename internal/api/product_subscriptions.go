package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ProductSubscription represents a Shopline product subscription plan.
type ProductSubscription struct {
	ID                string    `json:"id"`
	ProductID         string    `json:"product_id"`
	VariantID         string    `json:"variant_id"`
	CustomerID        string    `json:"customer_id"`
	SellingPlanID     string    `json:"selling_plan_id"`
	Status            string    `json:"status"`
	Frequency         string    `json:"frequency"`
	FrequencyInterval int       `json:"frequency_interval"`
	NextBillingDate   time.Time `json:"next_billing_date"`
	LastBillingDate   time.Time `json:"last_billing_date"`
	Price             string    `json:"price"`
	Currency          string    `json:"currency"`
	Quantity          int       `json:"quantity"`
	TotalCycles       int       `json:"total_cycles"`
	CompletedCycles   int       `json:"completed_cycles"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ProductSubscriptionsListOptions contains options for listing product subscriptions.
type ProductSubscriptionsListOptions struct {
	Page       int
	PageSize   int
	ProductID  string
	CustomerID string
	Status     string
}

// ProductSubscriptionsListResponse is the paginated response for product subscriptions.
type ProductSubscriptionsListResponse = ListResponse[ProductSubscription]

// ProductSubscriptionCreateRequest contains the data for creating a product subscription.
type ProductSubscriptionCreateRequest struct {
	ProductID       string `json:"product_id"`
	VariantID       string `json:"variant_id,omitempty"`
	CustomerID      string `json:"customer_id"`
	SellingPlanID   string `json:"selling_plan_id"`
	Quantity        int    `json:"quantity,omitempty"`
	NextBillingDate string `json:"next_billing_date,omitempty"`
}

// ListProductSubscriptions retrieves a list of product subscriptions.
func (c *Client) ListProductSubscriptions(ctx context.Context, opts *ProductSubscriptionsListOptions) (*ProductSubscriptionsListResponse, error) {
	path := "/product_subscriptions"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("product_id", opts.ProductID).
			String("customer_id", opts.CustomerID).
			String("status", opts.Status).
			Build()
	}

	var resp ProductSubscriptionsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetProductSubscription retrieves a single product subscription by ID.
func (c *Client) GetProductSubscription(ctx context.Context, id string) (*ProductSubscription, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("product subscription id is required")
	}
	var subscription ProductSubscription
	if err := c.Get(ctx, fmt.Sprintf("/product_subscriptions/%s", id), &subscription); err != nil {
		return nil, err
	}
	return &subscription, nil
}

// CreateProductSubscription creates a new product subscription.
func (c *Client) CreateProductSubscription(ctx context.Context, req *ProductSubscriptionCreateRequest) (*ProductSubscription, error) {
	var subscription ProductSubscription
	if err := c.Post(ctx, "/product_subscriptions", req, &subscription); err != nil {
		return nil, err
	}
	return &subscription, nil
}

// DeleteProductSubscription deletes (cancels) a product subscription.
func (c *Client) DeleteProductSubscription(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("product subscription id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/product_subscriptions/%s", id))
}
