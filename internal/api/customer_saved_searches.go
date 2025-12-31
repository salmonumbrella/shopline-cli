package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CustomerSavedSearch represents a saved customer search query.
type CustomerSavedSearch struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Query     string    `json:"query"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CustomerSavedSearchesListOptions contains options for listing saved searches.
type CustomerSavedSearchesListOptions struct {
	Page     int
	PageSize int
	Name     string
}

// CustomerSavedSearchesListResponse is the paginated response for customer saved searches.
type CustomerSavedSearchesListResponse = ListResponse[CustomerSavedSearch]

// CustomerSavedSearchCreateRequest contains the data for creating a saved search.
type CustomerSavedSearchCreateRequest struct {
	Name  string `json:"name"`
	Query string `json:"query"`
}

// ListCustomerSavedSearches retrieves a list of saved customer searches.
func (c *Client) ListCustomerSavedSearches(ctx context.Context, opts *CustomerSavedSearchesListOptions) (*CustomerSavedSearchesListResponse, error) {
	path := "/customer_saved_searches"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("name", opts.Name).
			Build()
	}

	var resp CustomerSavedSearchesListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCustomerSavedSearch retrieves a single saved search by ID.
func (c *Client) GetCustomerSavedSearch(ctx context.Context, id string) (*CustomerSavedSearch, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("saved search id is required")
	}
	var search CustomerSavedSearch
	if err := c.Get(ctx, fmt.Sprintf("/customer_saved_searches/%s", id), &search); err != nil {
		return nil, err
	}
	return &search, nil
}

// CreateCustomerSavedSearch creates a new saved customer search.
func (c *Client) CreateCustomerSavedSearch(ctx context.Context, req *CustomerSavedSearchCreateRequest) (*CustomerSavedSearch, error) {
	var search CustomerSavedSearch
	if err := c.Post(ctx, "/customer_saved_searches", req, &search); err != nil {
		return nil, err
	}
	return &search, nil
}

// DeleteCustomerSavedSearch deletes a saved customer search.
func (c *Client) DeleteCustomerSavedSearch(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("saved search id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/customer_saved_searches/%s", id))
}
