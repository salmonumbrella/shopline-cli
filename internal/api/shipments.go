package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Shipment represents a Shopline shipment.
type Shipment struct {
	ID                string    `json:"id"`
	OrderID           string    `json:"order_id"`
	FulfillmentID     string    `json:"fulfillment_id"`
	TrackingCompany   string    `json:"tracking_company"`
	TrackingNumber    string    `json:"tracking_number"`
	TrackingURL       string    `json:"tracking_url"`
	Status            string    `json:"status"`
	EstimatedDelivery time.Time `json:"estimated_delivery"`
	DeliveredAt       time.Time `json:"delivered_at"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ShipmentsListOptions contains options for listing shipments.
type ShipmentsListOptions struct {
	Page           int
	PageSize       int
	OrderID        string
	FulfillmentID  string
	Status         string
	TrackingNumber string
}

// ShipmentsListResponse is the paginated response for shipments.
type ShipmentsListResponse = ListResponse[Shipment]

// ShipmentCreateRequest contains the request body for creating a shipment.
type ShipmentCreateRequest struct {
	OrderID         string `json:"order_id"`
	FulfillmentID   string `json:"fulfillment_id"`
	TrackingCompany string `json:"tracking_company"`
	TrackingNumber  string `json:"tracking_number"`
	TrackingURL     string `json:"tracking_url,omitempty"`
}

// ShipmentUpdateRequest contains the request body for updating a shipment.
type ShipmentUpdateRequest struct {
	TrackingCompany string `json:"tracking_company,omitempty"`
	TrackingNumber  string `json:"tracking_number,omitempty"`
	TrackingURL     string `json:"tracking_url,omitempty"`
	Status          string `json:"status,omitempty"`
}

// ListShipments retrieves a list of shipments.
func (c *Client) ListShipments(ctx context.Context, opts *ShipmentsListOptions) (*ShipmentsListResponse, error) {
	path := "/shipments"
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
		if opts.FulfillmentID != "" {
			params.Set("fulfillment_id", opts.FulfillmentID)
		}
		if opts.Status != "" {
			params.Set("status", opts.Status)
		}
		if opts.TrackingNumber != "" {
			params.Set("tracking_number", opts.TrackingNumber)
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp ShipmentsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetShipment retrieves a single shipment by ID.
func (c *Client) GetShipment(ctx context.Context, id string) (*Shipment, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("shipment id is required")
	}
	var shipment Shipment
	if err := c.Get(ctx, fmt.Sprintf("/shipments/%s", id), &shipment); err != nil {
		return nil, err
	}
	return &shipment, nil
}

// CreateShipment creates a new shipment.
func (c *Client) CreateShipment(ctx context.Context, req *ShipmentCreateRequest) (*Shipment, error) {
	var shipment Shipment
	if err := c.Post(ctx, "/shipments", req, &shipment); err != nil {
		return nil, err
	}
	return &shipment, nil
}

// UpdateShipment updates an existing shipment.
func (c *Client) UpdateShipment(ctx context.Context, id string, req *ShipmentUpdateRequest) (*Shipment, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("shipment id is required")
	}
	var shipment Shipment
	if err := c.Put(ctx, fmt.Sprintf("/shipments/%s", id), req, &shipment); err != nil {
		return nil, err
	}
	return &shipment, nil
}

// DeleteShipment deletes a shipment.
func (c *Client) DeleteShipment(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("shipment id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/shipments/%s", id))
}
