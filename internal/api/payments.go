package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Payment represents a Shopline payment.
type Payment struct {
	ID             string      `json:"id"`
	OrderID        string      `json:"order_id"`
	Amount         string      `json:"amount"`
	Currency       string      `json:"currency"`
	Status         string      `json:"status"`
	Gateway        string      `json:"gateway"`
	PaymentMethod  string      `json:"payment_method"`
	TransactionID  string      `json:"transaction_id"`
	ErrorMessage   string      `json:"error_message"`
	CreditCard     *CreditCard `json:"credit_card,omitempty"`
	BillingAddress *Address    `json:"billing_address,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

// CreditCard represents credit card details (masked).
type CreditCard struct {
	Brand       string `json:"brand"`
	Last4       string `json:"last4"`
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
}

// Address represents a billing/shipping address.
type Address struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Address1  string `json:"address1"`
	Address2  string `json:"address2"`
	City      string `json:"city"`
	Province  string `json:"province"`
	Country   string `json:"country"`
	Zip       string `json:"zip"`
	Phone     string `json:"phone"`
}

// PaymentsListOptions contains options for listing payments.
type PaymentsListOptions struct {
	Page     int
	PageSize int
	Status   string
	Gateway  string
}

// PaymentsListResponse is the paginated response for payments.
type PaymentsListResponse = ListResponse[Payment]

// ListPayments retrieves a list of payments.
func (c *Client) ListPayments(ctx context.Context, opts *PaymentsListOptions) (*PaymentsListResponse, error) {
	path := "/payments"
	if opts != nil {
		params := url.Values{}
		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			params.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.Status != "" {
			params.Set("status", opts.Status)
		}
		if opts.Gateway != "" {
			params.Set("gateway", opts.Gateway)
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp PaymentsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPayment retrieves a single payment by ID.
func (c *Client) GetPayment(ctx context.Context, id string) (*Payment, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("payment id is required")
	}
	var payment Payment
	if err := c.Get(ctx, fmt.Sprintf("/payments/%s", id), &payment); err != nil {
		return nil, err
	}
	return &payment, nil
}

// ListOrderPayments retrieves payments for a specific order.
func (c *Client) ListOrderPayments(ctx context.Context, orderID string) (*PaymentsListResponse, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	var resp PaymentsListResponse
	if err := c.Get(ctx, fmt.Sprintf("/orders/%s/payments", orderID), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CapturePayment captures an authorized payment.
func (c *Client) CapturePayment(ctx context.Context, id string, amount string) (*Payment, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("payment id is required")
	}
	req := map[string]string{}
	if amount != "" {
		req["amount"] = amount
	}
	var payment Payment
	if err := c.Post(ctx, fmt.Sprintf("/payments/%s/capture", id), req, &payment); err != nil {
		return nil, err
	}
	return &payment, nil
}

// VoidPayment voids an authorized payment.
func (c *Client) VoidPayment(ctx context.Context, id string) (*Payment, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("payment id is required")
	}
	var payment Payment
	if err := c.Post(ctx, fmt.Sprintf("/payments/%s/void", id), nil, &payment); err != nil {
		return nil, err
	}
	return &payment, nil
}

// RefundPayment refunds a captured payment.
func (c *Client) RefundPayment(ctx context.Context, id string, amount string, reason string) (*Payment, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("payment id is required")
	}
	req := map[string]string{}
	if amount != "" {
		req["amount"] = amount
	}
	if reason != "" {
		req["reason"] = reason
	}
	var payment Payment
	if err := c.Post(ctx, fmt.Sprintf("/payments/%s/refund", id), req, &payment); err != nil {
		return nil, err
	}
	return &payment, nil
}
