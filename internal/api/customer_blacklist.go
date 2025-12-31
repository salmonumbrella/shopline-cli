package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CustomerBlacklist represents a blacklisted customer entry.
type CustomerBlacklist struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	Reason     string    `json:"reason"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CustomerBlacklistListOptions contains options for listing blacklisted customers.
type CustomerBlacklistListOptions struct {
	Page     int
	PageSize int
	Email    string
	Phone    string
}

// CustomerBlacklistListResponse is the paginated response for customer blacklist.
type CustomerBlacklistListResponse = ListResponse[CustomerBlacklist]

// CustomerBlacklistCreateRequest contains the data for creating a blacklist entry.
type CustomerBlacklistCreateRequest struct {
	CustomerID string `json:"customer_id,omitempty"`
	Email      string `json:"email,omitempty"`
	Phone      string `json:"phone,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

// ListCustomerBlacklist retrieves a list of blacklisted customers.
func (c *Client) ListCustomerBlacklist(ctx context.Context, opts *CustomerBlacklistListOptions) (*CustomerBlacklistListResponse, error) {
	path := "/customer_blacklist"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("email", opts.Email).
			String("phone", opts.Phone).
			Build()
	}

	var resp CustomerBlacklistListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCustomerBlacklist retrieves a single blacklist entry by ID.
func (c *Client) GetCustomerBlacklist(ctx context.Context, id string) (*CustomerBlacklist, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("blacklist entry id is required")
	}
	var entry CustomerBlacklist
	if err := c.Get(ctx, fmt.Sprintf("/customer_blacklist/%s", id), &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// CreateCustomerBlacklist creates a new blacklist entry.
func (c *Client) CreateCustomerBlacklist(ctx context.Context, req *CustomerBlacklistCreateRequest) (*CustomerBlacklist, error) {
	var entry CustomerBlacklist
	if err := c.Post(ctx, "/customer_blacklist", req, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// DeleteCustomerBlacklist deletes a blacklist entry.
func (c *Client) DeleteCustomerBlacklist(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("blacklist entry id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/customer_blacklist/%s", id))
}
