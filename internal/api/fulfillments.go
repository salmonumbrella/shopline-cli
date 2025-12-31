package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// FulfillmentStatus represents the status of a fulfillment.
type FulfillmentStatus string

const (
	FulfillmentStatusPending   FulfillmentStatus = "pending"
	FulfillmentStatusOpen      FulfillmentStatus = "open"
	FulfillmentStatusSuccess   FulfillmentStatus = "success"
	FulfillmentStatusCancelled FulfillmentStatus = "cancelled"
	FulfillmentStatusFailure   FulfillmentStatus = "failure"
)

// FulfillmentLineItem represents an item in a fulfillment.
type FulfillmentLineItem struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	VariantID string `json:"variant_id"`
	Title     string `json:"title"`
	Quantity  int    `json:"quantity"`
	SKU       string `json:"sku"`
}

// Fulfillment represents a Shopline fulfillment.
type Fulfillment struct {
	ID              string                `json:"id"`
	OrderID         string                `json:"order_id"`
	Status          FulfillmentStatus     `json:"status"`
	TrackingCompany string                `json:"tracking_company"`
	TrackingNumber  string                `json:"tracking_number"`
	TrackingURL     string                `json:"tracking_url"`
	LineItems       []FulfillmentLineItem `json:"line_items"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
}

// FulfillmentsListOptions contains options for listing fulfillments.
type FulfillmentsListOptions struct {
	Page     int
	PageSize int
	OrderID  string
	Status   string
}

// FulfillmentsListResponse is the paginated response for fulfillments.
type FulfillmentsListResponse = ListResponse[Fulfillment]

// ListFulfillments retrieves a list of fulfillments.
func (c *Client) ListFulfillments(ctx context.Context, opts *FulfillmentsListOptions) (*FulfillmentsListResponse, error) {
	path := "/fulfillments"
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
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp FulfillmentsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetFulfillment retrieves a single fulfillment by ID.
func (c *Client) GetFulfillment(ctx context.Context, id string) (*Fulfillment, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("fulfillment id is required")
	}
	var fulfillment Fulfillment
	if err := c.Get(ctx, fmt.Sprintf("/fulfillments/%s", id), &fulfillment); err != nil {
		return nil, err
	}
	return &fulfillment, nil
}
