package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// StoreCredit represents a customer store credit record.
type StoreCredit struct {
	ID          string    `json:"id"`
	CustomerID  string    `json:"customer_id"`
	Amount      string    `json:"amount"`
	Balance     string    `json:"balance"`
	Currency    string    `json:"currency"`
	Description string    `json:"description"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// StoreCreditsListOptions contains options for listing store credits.
type StoreCreditsListOptions struct {
	Page       int
	PageSize   int
	CustomerID string
}

// StoreCreditsListResponse is the paginated response for store credits.
type StoreCreditsListResponse = ListResponse[StoreCredit]

// StoreCreditCreateRequest contains the data for creating a store credit.
type StoreCreditCreateRequest struct {
	CustomerID  string     `json:"customer_id"`
	Amount      string     `json:"amount"`
	Currency    string     `json:"currency,omitempty"`
	Description string     `json:"description,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// ListStoreCredits retrieves a list of store credits.
func (c *Client) ListStoreCredits(ctx context.Context, opts *StoreCreditsListOptions) (*StoreCreditsListResponse, error) {
	path := "/store_credits"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("customer_id", opts.CustomerID).
			Build()
	}

	var resp StoreCreditsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetStoreCredit retrieves a single store credit by ID.
func (c *Client) GetStoreCredit(ctx context.Context, id string) (*StoreCredit, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("store credit id is required")
	}
	var credit StoreCredit
	if err := c.Get(ctx, fmt.Sprintf("/store_credits/%s", id), &credit); err != nil {
		return nil, err
	}
	return &credit, nil
}

// CreateStoreCredit creates a new store credit.
func (c *Client) CreateStoreCredit(ctx context.Context, req *StoreCreditCreateRequest) (*StoreCredit, error) {
	var credit StoreCredit
	if err := c.Post(ctx, "/store_credits", req, &credit); err != nil {
		return nil, err
	}
	return &credit, nil
}

// DeleteStoreCredit deletes a store credit.
func (c *Client) DeleteStoreCredit(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("store credit id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/store_credits/%s", id))
}
