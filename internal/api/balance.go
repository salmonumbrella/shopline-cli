package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Balance represents a Shopline account balance.
type Balance struct {
	Currency  string    `json:"currency"`
	Available string    `json:"available"`
	Pending   string    `json:"pending"`
	Reserved  string    `json:"reserved,omitempty"`
	Total     string    `json:"total"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BalanceTransaction represents a balance transaction.
type BalanceTransaction struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"`
	Amount      string     `json:"amount"`
	Currency    string     `json:"currency"`
	Net         string     `json:"net"`
	Fee         string     `json:"fee,omitempty"`
	Status      string     `json:"status"`
	Description string     `json:"description,omitempty"`
	SourceID    string     `json:"source_id,omitempty"`
	SourceType  string     `json:"source_type,omitempty"`
	AvailableOn *time.Time `json:"available_on,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// BalanceTransactionsListOptions contains options for listing balance transactions.
type BalanceTransactionsListOptions struct {
	Page       int
	PageSize   int
	Type       string
	SourceType string
	DateMin    *time.Time
	DateMax    *time.Time
}

// BalanceTransactionsListResponse is the paginated response for balance transactions.
type BalanceTransactionsListResponse = ListResponse[BalanceTransaction]

// GetBalance retrieves the current account balance.
func (c *Client) GetBalance(ctx context.Context) (*Balance, error) {
	var balance Balance
	if err := c.Get(ctx, "/balance", &balance); err != nil {
		return nil, err
	}
	return &balance, nil
}

// ListBalanceTransactions retrieves a list of balance transactions.
func (c *Client) ListBalanceTransactions(ctx context.Context, opts *BalanceTransactionsListOptions) (*BalanceTransactionsListResponse, error) {
	path := "/balance/transactions"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("type", opts.Type).
			String("source_type", opts.SourceType).
			Time("date_min", opts.DateMin).
			Time("date_max", opts.DateMax).
			Build()
	}

	var resp BalanceTransactionsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetBalanceTransaction retrieves a single balance transaction by ID.
func (c *Client) GetBalanceTransaction(ctx context.Context, id string) (*BalanceTransaction, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("balance transaction id is required")
	}
	var txn BalanceTransaction
	if err := c.Get(ctx, fmt.Sprintf("/balance/transactions/%s", id), &txn); err != nil {
		return nil, err
	}
	return &txn, nil
}
