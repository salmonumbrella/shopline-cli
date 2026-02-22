package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// MemberPoints represents a customer's point balance.
type MemberPoints struct {
	CustomerID      string    `json:"customer_id"`
	TotalPoints     int       `json:"total_points"`
	AvailablePoints int       `json:"available_points"`
	PendingPoints   int       `json:"pending_points"`
	ExpiredPoints   int       `json:"expired_points"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// PointsTransaction represents a points transaction.
type PointsTransaction struct {
	ID          string    `json:"id"`
	CustomerID  string    `json:"customer_id"`
	Type        string    `json:"type"`
	Points      int       `json:"points"`
	Balance     int       `json:"balance"`
	Description string    `json:"description"`
	OrderID     string    `json:"order_id"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// PointsTransactionsListOptions contains options for listing transactions.
type PointsTransactionsListOptions struct {
	Page     int
	PageSize int
	Type     string
}

// PointsTransactionsListResponse is the paginated response for points transactions.
type PointsTransactionsListResponse = ListResponse[PointsTransaction]

// PointsAdjustRequest contains the request body for adjusting points.
type PointsAdjustRequest struct {
	Points      int    `json:"points"`
	Description string `json:"description,omitempty"`
}

// GetMemberPoints retrieves a customer's point balance.
func (c *Client) GetMemberPoints(ctx context.Context, customerID string) (*MemberPoints, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	var points MemberPoints
	if err := c.Get(ctx, fmt.Sprintf("/customers/%s/points", customerID), &points); err != nil {
		return nil, err
	}
	return &points, nil
}

// ListPointsTransactions retrieves a list of points transactions for a customer.
func (c *Client) ListPointsTransactions(ctx context.Context, customerID string, opts *PointsTransactionsListOptions) (*PointsTransactionsListResponse, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	path := fmt.Sprintf("/customers/%s/points/transactions", customerID)
	if opts != nil {
		params := url.Values{}
		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			params.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.Type != "" {
			params.Set("type", opts.Type)
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp PointsTransactionsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AdjustMemberPoints adjusts a customer's points balance.
func (c *Client) AdjustMemberPoints(ctx context.Context, customerID string, points int, description string) (*MemberPoints, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	req := PointsAdjustRequest{
		Points:      points,
		Description: description,
	}
	var result MemberPoints
	if err := c.Post(ctx, fmt.Sprintf("/customers/%s/points/adjust", customerID), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
