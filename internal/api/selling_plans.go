package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// SellingPlan represents a Shopline selling plan configuration.
type SellingPlan struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	BillingPolicy     string    `json:"billing_policy"`
	DeliveryPolicy    string    `json:"delivery_policy"`
	Frequency         string    `json:"frequency"`
	FrequencyInterval int       `json:"frequency_interval"`
	TrialDays         int       `json:"trial_days"`
	DiscountType      string    `json:"discount_type"`
	DiscountValue     string    `json:"discount_value"`
	Status            string    `json:"status"`
	Position          int       `json:"position"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// SellingPlansListOptions contains options for listing selling plans.
type SellingPlansListOptions struct {
	Page     int
	PageSize int
	Status   string
}

// SellingPlansListResponse is the paginated response for selling plans.
type SellingPlansListResponse = ListResponse[SellingPlan]

// SellingPlanCreateRequest contains the data for creating a selling plan.
type SellingPlanCreateRequest struct {
	Name              string `json:"name"`
	Description       string `json:"description,omitempty"`
	BillingPolicy     string `json:"billing_policy,omitempty"`
	DeliveryPolicy    string `json:"delivery_policy,omitempty"`
	Frequency         string `json:"frequency"`
	FrequencyInterval int    `json:"frequency_interval,omitempty"`
	TrialDays         int    `json:"trial_days,omitempty"`
	DiscountType      string `json:"discount_type,omitempty"`
	DiscountValue     string `json:"discount_value,omitempty"`
	Position          int    `json:"position,omitempty"`
}

// ListSellingPlans retrieves a list of selling plans.
func (c *Client) ListSellingPlans(ctx context.Context, opts *SellingPlansListOptions) (*SellingPlansListResponse, error) {
	path := "/selling_plans"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("status", opts.Status).
			Build()
	}

	var resp SellingPlansListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetSellingPlan retrieves a single selling plan by ID.
func (c *Client) GetSellingPlan(ctx context.Context, id string) (*SellingPlan, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("selling plan id is required")
	}
	var plan SellingPlan
	if err := c.Get(ctx, fmt.Sprintf("/selling_plans/%s", id), &plan); err != nil {
		return nil, err
	}
	return &plan, nil
}

// CreateSellingPlan creates a new selling plan.
func (c *Client) CreateSellingPlan(ctx context.Context, req *SellingPlanCreateRequest) (*SellingPlan, error) {
	var plan SellingPlan
	if err := c.Post(ctx, "/selling_plans", req, &plan); err != nil {
		return nil, err
	}
	return &plan, nil
}

// DeleteSellingPlan deletes a selling plan.
func (c *Client) DeleteSellingPlan(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("selling plan id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/selling_plans/%s", id))
}
