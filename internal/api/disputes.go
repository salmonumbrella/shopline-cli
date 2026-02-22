package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Dispute represents a Shopline payment dispute.
type Dispute struct {
	ID                string           `json:"id"`
	OrderID           string           `json:"order_id"`
	PaymentID         string           `json:"payment_id"`
	Amount            string           `json:"amount"`
	Currency          string           `json:"currency"`
	Status            string           `json:"status"`
	Reason            string           `json:"reason"`
	NetworkReasonCode string           `json:"network_reason_code,omitempty"`
	Evidence          *DisputeEvidence `json:"evidence,omitempty"`
	EvidenceDueBy     *time.Time       `json:"evidence_due_by,omitempty"`
	ResolvedAt        *time.Time       `json:"resolved_at,omitempty"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
}

// DisputeEvidence contains evidence for a dispute.
type DisputeEvidence struct {
	CustomerName           string `json:"customer_name,omitempty"`
	CustomerEmail          string `json:"customer_email,omitempty"`
	CustomerPurchaseIP     string `json:"customer_purchase_ip,omitempty"`
	ProductDescription     string `json:"product_description,omitempty"`
	ShippingCarrier        string `json:"shipping_carrier,omitempty"`
	ShippingTrackingNumber string `json:"shipping_tracking_number,omitempty"`
	ShippingDate           string `json:"shipping_date,omitempty"`
	RefundPolicy           string `json:"refund_policy,omitempty"`
	UncategorizedText      string `json:"uncategorized_text,omitempty"`
}

// DisputesListOptions contains options for listing disputes.
type DisputesListOptions struct {
	Page     int
	PageSize int
	Status   string
	Reason   string
}

// DisputesListResponse is the paginated response for disputes.
type DisputesListResponse = ListResponse[Dispute]

// DisputeUpdateEvidenceRequest contains the request body for updating dispute evidence.
type DisputeUpdateEvidenceRequest struct {
	CustomerName           string `json:"customer_name,omitempty"`
	CustomerEmail          string `json:"customer_email,omitempty"`
	CustomerPurchaseIP     string `json:"customer_purchase_ip,omitempty"`
	ProductDescription     string `json:"product_description,omitempty"`
	ShippingCarrier        string `json:"shipping_carrier,omitempty"`
	ShippingTrackingNumber string `json:"shipping_tracking_number,omitempty"`
	ShippingDate           string `json:"shipping_date,omitempty"`
	RefundPolicy           string `json:"refund_policy,omitempty"`
	UncategorizedText      string `json:"uncategorized_text,omitempty"`
}

// ListDisputes retrieves a list of disputes.
func (c *Client) ListDisputes(ctx context.Context, opts *DisputesListOptions) (*DisputesListResponse, error) {
	path := "/disputes"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("status", opts.Status).
			String("reason", opts.Reason).
			Build()
	}

	var resp DisputesListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetDispute retrieves a single dispute by ID.
func (c *Client) GetDispute(ctx context.Context, id string) (*Dispute, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("dispute id is required")
	}
	var dispute Dispute
	if err := c.Get(ctx, fmt.Sprintf("/disputes/%s", id), &dispute); err != nil {
		return nil, err
	}
	return &dispute, nil
}

// UpdateDisputeEvidence updates the evidence for a dispute.
func (c *Client) UpdateDisputeEvidence(ctx context.Context, id string, req *DisputeUpdateEvidenceRequest) (*Dispute, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("dispute id is required")
	}
	var dispute Dispute
	if err := c.Put(ctx, fmt.Sprintf("/disputes/%s/evidence", id), req, &dispute); err != nil {
		return nil, err
	}
	return &dispute, nil
}

// SubmitDispute submits dispute evidence for review.
func (c *Client) SubmitDispute(ctx context.Context, id string) (*Dispute, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("dispute id is required")
	}
	var dispute Dispute
	if err := c.Post(ctx, fmt.Sprintf("/disputes/%s/submit", id), nil, &dispute); err != nil {
		return nil, err
	}
	return &dispute, nil
}

// AcceptDispute accepts a dispute (concede to the customer).
func (c *Client) AcceptDispute(ctx context.Context, id string) (*Dispute, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("dispute id is required")
	}
	var dispute Dispute
	if err := c.Post(ctx, fmt.Sprintf("/disputes/%s/accept", id), nil, &dispute); err != nil {
		return nil, err
	}
	return &dispute, nil
}
