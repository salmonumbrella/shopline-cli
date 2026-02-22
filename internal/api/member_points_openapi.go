package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Member points endpoints as documented in Shopline Open API reference.
// These endpoints don't have stable schemas in the docs mirror; keep them raw.

// GetCustomersMembershipInfo retrieves customers membership info.
//
// Docs: GET /customers/membership_info
func (c *Client) GetCustomersMembershipInfo(ctx context.Context) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/customers/membership_info", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetCustomerMemberPointsHistory retrieves member points history for a customer.
//
// Docs: GET /customers/{id}/member_points
func (c *Client) GetCustomerMemberPointsHistory(ctx context.Context, customerID string) (json.RawMessage, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/customers/%s/member_points", customerID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateCustomerMemberPoints updates customer member points.
//
// Docs: POST /customers/{id}/member_points
func (c *Client) UpdateCustomerMemberPoints(ctx context.Context, customerID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	var resp json.RawMessage
	if err := c.Post(ctx, fmt.Sprintf("/customers/%s/member_points", customerID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetCustomerMembershipTierActionLogs retrieves membership tier action logs for a customer.
//
// Docs: GET /customers/{id}/membership_tier/action_logs
func (c *Client) GetCustomerMembershipTierActionLogs(ctx context.Context, customerID string) (json.RawMessage, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/customers/%s/membership_tier/action_logs", customerID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// ListMemberPointRules lists member point rules.
//
// Docs: GET /member_point_rules
func (c *Client) ListMemberPointRules(ctx context.Context) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/member_point_rules", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// BulkUpdateMemberPoints bulk updates member points.
//
// Docs: POST /member_points/bulk_update
func (c *Client) BulkUpdateMemberPoints(ctx context.Context, body any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Post(ctx, "/member_points/bulk_update", body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
