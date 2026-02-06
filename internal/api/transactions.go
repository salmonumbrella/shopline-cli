package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Transaction represents a Shopline transaction.
type Transaction struct {
	ID            string    `json:"id"`
	OrderID       string    `json:"order_id"`
	Kind          string    `json:"kind"`
	Gateway       string    `json:"gateway"`
	Status        string    `json:"status"`
	Amount        string    `json:"amount"`
	Currency      string    `json:"currency"`
	Authorization string    `json:"authorization"`
	ParentID      string    `json:"parent_id"`
	ErrorCode     string    `json:"error_code"`
	Message       string    `json:"message"`
	ProcessedAt   time.Time `json:"processed_at"`
	CreatedAt     time.Time `json:"created_at"`
}

// TransactionsListOptions contains options for listing transactions.
type TransactionsListOptions struct {
	Page     int
	PageSize int
	OrderID  string
	Status   string
	Kind     string
}

// TransactionsListResponse is the paginated response for transactions.
type TransactionsListResponse = ListResponse[Transaction]

// ListTransactions retrieves a list of transactions.
func (c *Client) ListTransactions(ctx context.Context, opts *TransactionsListOptions) (*TransactionsListResponse, error) {
	path := "/transactions"
	if opts != nil {
		params := url.Values{}
		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			params.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.OrderID != "" {
			params.Set("order_id", opts.OrderID)
		}
		if opts.Status != "" {
			params.Set("status", opts.Status)
		}
		if opts.Kind != "" {
			params.Set("kind", opts.Kind)
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp TransactionsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTransaction retrieves a single transaction by ID.
func (c *Client) GetTransaction(ctx context.Context, id string) (*Transaction, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("transaction id is required")
	}
	var transaction Transaction
	if err := c.Get(ctx, fmt.Sprintf("/transactions/%s", id), &transaction); err != nil {
		return nil, err
	}
	return &transaction, nil
}

// ListOrderTransactions retrieves transactions for a specific order.
func (c *Client) ListOrderTransactions(ctx context.Context, orderID string) (*TransactionsListResponse, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	var resp TransactionsListResponse
	if err := c.Get(ctx, fmt.Sprintf("/orders/%s/transactions", orderID), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
