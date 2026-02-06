package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// GetMultipassSecret retrieves the multipass secret (documented endpoint).
//
// Docs: GET /multipass/secret
func (c *Client) GetMultipassSecret(ctx context.Context) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/multipass/secret", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateMultipassSecret generates a new multipass secret or returns the existing secret (documented endpoint).
//
// Docs: POST /multipass/secret
func (c *Client) CreateMultipassSecret(ctx context.Context, body any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Post(ctx, "/multipass/secret", body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// ListMultipassLinkings returns active multipass linking records (documented endpoint).
//
// Docs: GET /multipass/linkings
func (c *Client) ListMultipassLinkings(ctx context.Context, customerIDs []string) (json.RawMessage, error) {
	path := "/multipass/linkings"
	if len(customerIDs) > 0 {
		params := url.Values{}
		// Docs text says "Given customer_ids". We encode as a comma-separated list.
		// If the API expects repeated params, users can still call the raw endpoint via future param flags.
		params.Set("customer_ids", strings.Join(customerIDs, ","))
		path += "?" + params.Encode()
	}

	var resp json.RawMessage
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateMultipassCustomerLinking updates customer's active linking record (documented endpoint).
//
// Docs: POST /multipass/customers/{customer_id}/linkings
func (c *Client) UpdateMultipassCustomerLinking(ctx context.Context, customerID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	var resp json.RawMessage
	if err := c.Post(ctx, fmt.Sprintf("/multipass/customers/%s/linkings", customerID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// DeleteMultipassCustomerLinking marks customer's active linking record as inactive (documented endpoint).
//
// Docs: DELETE /multipass/customers/{customer_id}/linkings
func (c *Client) DeleteMultipassCustomerLinking(ctx context.Context, customerID string) (json.RawMessage, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	var resp json.RawMessage
	if err := c.DeleteWithBody(ctx, fmt.Sprintf("/multipass/customers/%s/linkings", customerID), nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
