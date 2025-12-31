package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CompanyCredit represents B2B company credit balance and transactions.
type CompanyCredit struct {
	ID            string    `json:"id"`
	CompanyID     string    `json:"company_id"`
	CompanyName   string    `json:"company_name"`
	CreditBalance float64   `json:"credit_balance"`
	CreditLimit   float64   `json:"credit_limit"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CompanyCreditTransaction represents a credit transaction.
type CompanyCreditTransaction struct {
	ID          string    `json:"id"`
	CreditID    string    `json:"credit_id"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Balance     float64   `json:"balance"`
	Description string    `json:"description"`
	ReferenceID string    `json:"reference_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// CompanyCreditsListOptions contains options for listing company credits.
type CompanyCreditsListOptions struct {
	Page      int
	PageSize  int
	CompanyID string
	Status    string
}

// CompanyCreditsListResponse is the paginated response for company credits.
type CompanyCreditsListResponse = ListResponse[CompanyCredit]

// CompanyCreditTransactionsListResponse is the paginated response for credit transactions.
type CompanyCreditTransactionsListResponse = ListResponse[CompanyCreditTransaction]

// CompanyCreditCreateRequest contains the request body for creating company credit.
type CompanyCreditCreateRequest struct {
	CompanyID   string  `json:"company_id"`
	CreditLimit float64 `json:"credit_limit"`
	Currency    string  `json:"currency,omitempty"`
}

// CompanyCreditAdjustRequest contains the request body for adjusting credit.
type CompanyCreditAdjustRequest struct {
	Amount      float64 `json:"amount"`
	Description string  `json:"description,omitempty"`
	ReferenceID string  `json:"reference_id,omitempty"`
}

// ListCompanyCredits retrieves a list of company credits.
func (c *Client) ListCompanyCredits(ctx context.Context, opts *CompanyCreditsListOptions) (*CompanyCreditsListResponse, error) {
	path := "/company_credits" + NewQuery().
		Int("page", opts.Page).
		Int("page_size", opts.PageSize).
		String("company_id", opts.CompanyID).
		String("status", opts.Status).
		Build()

	var resp CompanyCreditsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCompanyCredit retrieves a single company credit by ID.
func (c *Client) GetCompanyCredit(ctx context.Context, id string) (*CompanyCredit, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("company credit id is required")
	}
	var credit CompanyCredit
	if err := c.Get(ctx, fmt.Sprintf("/company_credits/%s", id), &credit); err != nil {
		return nil, err
	}
	return &credit, nil
}

// CreateCompanyCredit creates a new company credit.
func (c *Client) CreateCompanyCredit(ctx context.Context, req *CompanyCreditCreateRequest) (*CompanyCredit, error) {
	var credit CompanyCredit
	if err := c.Post(ctx, "/company_credits", req, &credit); err != nil {
		return nil, err
	}
	return &credit, nil
}

// AdjustCompanyCredit adjusts a company's credit balance.
func (c *Client) AdjustCompanyCredit(ctx context.Context, id string, req *CompanyCreditAdjustRequest) (*CompanyCredit, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("company credit id is required")
	}
	var credit CompanyCredit
	if err := c.Post(ctx, fmt.Sprintf("/company_credits/%s/adjust", id), req, &credit); err != nil {
		return nil, err
	}
	return &credit, nil
}

// ListCompanyCreditTransactions retrieves transactions for a company credit.
func (c *Client) ListCompanyCreditTransactions(ctx context.Context, creditID string, page, pageSize int) (*CompanyCreditTransactionsListResponse, error) {
	if strings.TrimSpace(creditID) == "" {
		return nil, fmt.Errorf("company credit id is required")
	}
	path := fmt.Sprintf("/company_credits/%s/transactions", creditID) + NewQuery().
		Int("page", page).
		Int("page_size", pageSize).
		Build()

	var resp CompanyCreditTransactionsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteCompanyCredit deletes a company credit.
func (c *Client) DeleteCompanyCredit(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("company credit id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/company_credits/%s", id))
}
