package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Payout represents a Shopline payout.
type Payout struct {
	ID            string         `json:"id"`
	Amount        string         `json:"amount"`
	Currency      string         `json:"currency"`
	Status        string         `json:"status"`
	Type          string         `json:"type"`
	BankAccount   string         `json:"bank_account"`
	TransactionID string         `json:"transaction_id,omitempty"`
	Fee           string         `json:"fee,omitempty"`
	Net           string         `json:"net,omitempty"`
	Summary       *PayoutSummary `json:"summary,omitempty"`
	ScheduledDate *time.Time     `json:"scheduled_date,omitempty"`
	ArrivalDate   *time.Time     `json:"arrival_date,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// PayoutSummary contains the breakdown of a payout.
type PayoutSummary struct {
	Sales         string `json:"sales"`
	Refunds       string `json:"refunds"`
	Adjustments   string `json:"adjustments"`
	Charges       string `json:"charges"`
	ReservedFunds string `json:"reserved_funds"`
}

// PayoutsListOptions contains options for listing payouts.
type PayoutsListOptions struct {
	Page     int
	PageSize int
	Status   string
	DateMin  *time.Time
	DateMax  *time.Time
}

// PayoutsListResponse is the paginated response for payouts.
type PayoutsListResponse = ListResponse[Payout]

// ListPayouts retrieves a list of payouts.
func (c *Client) ListPayouts(ctx context.Context, opts *PayoutsListOptions) (*PayoutsListResponse, error) {
	path := "/payouts"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("status", opts.Status).
			Time("date_min", opts.DateMin).
			Time("date_max", opts.DateMax).
			Build()
	}

	var resp PayoutsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPayout retrieves a single payout by ID.
func (c *Client) GetPayout(ctx context.Context, id string) (*Payout, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("payout id is required")
	}
	var payout Payout
	if err := c.Get(ctx, fmt.Sprintf("/payouts/%s", id), &payout); err != nil {
		return nil, err
	}
	return &payout, nil
}
