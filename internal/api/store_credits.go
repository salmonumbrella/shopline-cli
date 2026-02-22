package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// StoreCreditUpdateRequest is the request body for POST /customers/{id}/store_credits.
// Field names match the Shopline Open API spec exactly.
type StoreCreditUpdateRequest struct {
	Value                 int    `json:"value"`                             // Credits to add (positive) or deduct (negative), -999999~999999
	Remarks               string `json:"remarks"`                           // Reason, max 50 chars
	ExpiresAt             string `json:"expires_at,omitempty"`              // Optional expiry date (ISO 8601)
	Type                  string `json:"type,omitempty"`                    // Defaults to "manual_credit"
	EmailTarget           *int   `json:"email_target,omitempty"`            // 1=not send, 3=send all
	SMSNotificationTarget *int   `json:"sms_notification_target,omitempty"` // 1=not send, 2=verified only, 3=send all
	Replace               *bool  `json:"replace,omitempty"`                 // Replace all credits with this value
	PerformerID           string `json:"performer_id,omitempty"`            // Performer ID
	PerformerType         string `json:"performer_type,omitempty"`          // "User" or "Agent"
}

// ListCustomerStoreCredits retrieves store credit history for a customer.
//
// Docs: GET /customers/{id}/store_credits
func (c *Client) ListCustomerStoreCredits(ctx context.Context, customerID string, page, perPage int) (json.RawMessage, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	path := fmt.Sprintf("/customers/%s/store_credits", customerID)
	q := NewQuery()
	if page > 0 {
		q = q.Int("page", page)
	}
	if perPage > 0 {
		q = q.Int("per_page", perPage)
	}
	path += q.Build()

	var resp json.RawMessage
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateCustomerStoreCredits adds or deducts store credits for a customer.
//
// Docs: POST /customers/{id}/store_credits
func (c *Client) UpdateCustomerStoreCredits(ctx context.Context, customerID string, req *StoreCreditUpdateRequest) (json.RawMessage, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	if req == nil {
		return nil, fmt.Errorf("request body is required")
	}
	var resp json.RawMessage
	if err := c.Post(ctx, fmt.Sprintf("/customers/%s/store_credits", customerID), req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
