package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Customer store credits

// GetCustomerStoreCredits retrieves store credit history for a customer.
//
// Docs: GET /customers/{id}/store_credits
func (c *Client) GetCustomerStoreCredits(ctx context.Context, customerID string) (json.RawMessage, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/customers/%s/store_credits", customerID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateCustomerStoreCredits updates a customer's store credits.
//
// Docs: POST /customers/{id}/store_credits
func (c *Client) CreateCustomerStoreCredits(ctx context.Context, customerID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	var resp json.RawMessage
	if err := c.Post(ctx, fmt.Sprintf("/customers/%s/store_credits", customerID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
