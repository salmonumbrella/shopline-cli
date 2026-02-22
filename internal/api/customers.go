package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Customer represents a Shopline customer.
type Customer struct {
	ID               string                 `json:"id"`
	Email            string                 `json:"email"`
	FirstName        string                 `json:"first_name"`
	LastName         string                 `json:"last_name"`
	Phone            string                 `json:"phone"`
	AcceptsMarketing bool                   `json:"accepts_marketing"`
	CreditBalance    *float64               `json:"credit_balance,omitempty"`
	Subscriptions    []CustomerSubscription `json:"subscriptions,omitempty"`
	OrdersCount      int                    `json:"orders_count"`
	TotalSpent       string                 `json:"total_spent"`
	Currency         string                 `json:"currency"`
	Tags             []string               `json:"tags"`
	Note             string                 `json:"note"`
	State            string                 `json:"state"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// CustomerSubscription represents a marketing subscription state for a customer.
type CustomerSubscription struct {
	Platform string `json:"platform"`
	IsActive bool   `json:"is_active"`
}

// CustomersListOptions contains options for listing customers.
type CustomersListOptions struct {
	Page             int
	PageSize         int
	Email            string
	State            string
	Tags             string
	AcceptsMarketing *bool
	SortBy           string
	SortOrder        string
}

// CustomersListResponse is the paginated response for customers.
type CustomersListResponse = ListResponse[Customer]

// ListCustomers retrieves a list of customers.
func (c *Client) ListCustomers(ctx context.Context, opts *CustomersListOptions) (*CustomersListResponse, error) {
	path := "/customers"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("email", opts.Email).
			String("state", opts.State).
			String("tags", opts.Tags).
			BoolPtr("accepts_marketing", opts.AcceptsMarketing).
			String("sort_by", opts.SortBy).
			String("sort_order", opts.SortOrder).
			Build()
	}

	var resp CustomersListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCustomer retrieves a single customer by ID.
func (c *Client) GetCustomer(ctx context.Context, id string) (*Customer, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	var customer Customer
	if err := c.Get(ctx, fmt.Sprintf("/customers/%s", id), &customer); err != nil {
		return nil, err
	}
	return &customer, nil
}

// CustomerCreateRequest contains the request body for creating a customer.
type CustomerCreateRequest struct {
	Email            string   `json:"email"`
	FirstName        string   `json:"first_name,omitempty"`
	LastName         string   `json:"last_name,omitempty"`
	Phone            string   `json:"phone,omitempty"`
	AcceptsMarketing bool     `json:"accepts_marketing,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	Note             string   `json:"note,omitempty"`
}

// CustomerUpdateRequest contains the request body for updating a customer.
type CustomerUpdateRequest struct {
	Email            *string  `json:"email,omitempty"`
	FirstName        *string  `json:"first_name,omitempty"`
	LastName         *string  `json:"last_name,omitempty"`
	Phone            *string  `json:"phone,omitempty"`
	AcceptsMarketing *bool    `json:"accepts_marketing,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	Note             *string  `json:"note,omitempty"`
}

// CustomerSearchOptions contains options for searching customers.
type CustomerSearchOptions struct {
	Query    string
	Email    string
	Phone    string
	Page     int
	PageSize int
}

// CustomerTagsUpdateRequest contains the request body for updating customer tags.
type CustomerTagsUpdateRequest struct {
	Add    []string `json:"add,omitempty"`
	Remove []string `json:"remove,omitempty"`
}

// CustomerPromotionsResponse represents customer promotions.
type CustomerPromotionsResponse struct {
	Items []interface{} `json:"items"`
}

// CreateCustomer creates a new customer.
func (c *Client) CreateCustomer(ctx context.Context, req *CustomerCreateRequest) (*Customer, error) {
	var customer Customer
	if err := c.Post(ctx, "/customers", req, &customer); err != nil {
		return nil, err
	}
	return &customer, nil
}

// UpdateCustomer updates an existing customer.
func (c *Client) UpdateCustomer(ctx context.Context, id string, req *CustomerUpdateRequest) (*Customer, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	var customer Customer
	if err := c.Put(ctx, fmt.Sprintf("/customers/%s", id), req, &customer); err != nil {
		return nil, err
	}
	return &customer, nil
}

// DeleteCustomer deletes a customer.
func (c *Client) DeleteCustomer(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("customer id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/customers/%s", id))
}

// SearchCustomers searches for customers with query parameters.
func (c *Client) SearchCustomers(ctx context.Context, opts *CustomerSearchOptions) (*CustomersListResponse, error) {
	if opts == nil {
		return nil, fmt.Errorf("search options are required")
	}
	path := "/customers/search" + NewQuery().
		String("query", opts.Query).
		String("email", opts.Email).
		String("phone", opts.Phone).
		Int("page", opts.Page).
		Int("page_size", opts.PageSize).
		Build()

	var resp CustomersListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateCustomerTags updates tags for a customer (add/remove).
func (c *Client) UpdateCustomerTags(ctx context.Context, id string, req *CustomerTagsUpdateRequest) (*Customer, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	var customer Customer
	if err := c.Patch(ctx, fmt.Sprintf("/customers/%s/tags", id), req, &customer); err != nil {
		return nil, err
	}
	return &customer, nil
}

// SetCustomerTags replaces all tags for a customer.
func (c *Client) SetCustomerTags(ctx context.Context, id string, tags []string) (*Customer, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	req := struct {
		Tags []string `json:"tags"`
	}{Tags: tags}
	var customer Customer
	if err := c.Put(ctx, fmt.Sprintf("/customers/%s/tags", id), req, &customer); err != nil {
		return nil, err
	}
	return &customer, nil
}

// GetCustomerPromotions retrieves promotions available to a customer.
func (c *Client) GetCustomerPromotions(ctx context.Context, id string) (*CustomerPromotionsResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	var resp CustomerPromotionsResponse
	if err := c.Get(ctx, fmt.Sprintf("/customers/%s/promotions", id), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCustomerCouponPromotions retrieves coupon promotions available to a customer.
//
// Docs: GET /customers/{customer_id}/coupon_promotions
func (c *Client) GetCustomerCouponPromotions(ctx context.Context, id string) (json.RawMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/customers/%s/coupon_promotions", id), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateCustomerSubscriptions updates customer subscriptions.
//
// Docs: PUT /customers/{customer_id}/subscriptions
func (c *Client) UpdateCustomerSubscriptions(ctx context.Context, customerID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	var resp json.RawMessage
	if err := c.Put(ctx, fmt.Sprintf("/customers/%s/subscriptions", customerID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetLineCustomer retrieves a customer by LINE ID.
//
// Docs: GET /customers/line/{lineId}
func (c *Client) GetLineCustomer(ctx context.Context, lineID string) (*Customer, error) {
	if strings.TrimSpace(lineID) == "" {
		return nil, fmt.Errorf("line id is required")
	}
	var customer Customer
	if err := c.Get(ctx, fmt.Sprintf("/customers/line/%s", lineID), &customer); err != nil {
		return nil, err
	}
	return &customer, nil
}
