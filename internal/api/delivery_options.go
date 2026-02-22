package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// DeliveryOption represents a Shopline delivery option.
type DeliveryOption struct {
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	Type               string                 `json:"type"`
	Status             string                 `json:"status"`
	Description        string                 `json:"description"`
	SupportedCountries []string               `json:"supported_countries"`
	Rates              map[string]interface{} `json:"rates"`
	Config             map[string]interface{} `json:"config"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

// DeliveryTimeSlot represents a time slot for delivery.
type DeliveryTimeSlot struct {
	ID        string    `json:"id"`
	Date      string    `json:"date"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
	Available bool      `json:"available"`
	Capacity  int       `json:"capacity"`
	Booked    int       `json:"booked"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DeliveryOptionsListOptions contains options for listing delivery options.
type DeliveryOptionsListOptions struct {
	Page     int
	PageSize int
	Status   string
	Type     string
}

// DeliveryOptionsListResponse is the paginated response for delivery options.
type DeliveryOptionsListResponse = ListResponse[DeliveryOption]

// DeliveryTimeSlotsListOptions contains options for listing delivery time slots.
type DeliveryTimeSlotsListOptions struct {
	Page      int
	PageSize  int
	StartDate string
	EndDate   string
}

// DeliveryTimeSlotsListResponse is the paginated response for delivery time slots.
type DeliveryTimeSlotsListResponse = ListResponse[DeliveryTimeSlot]

// PickupStoreUpdateRequest contains the request body for updating pickup store.
type PickupStoreUpdateRequest struct {
	StoreID   string `json:"store_id"`
	StoreName string `json:"store_name,omitempty"`
	Address   string `json:"address,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

// DeliveryConfigOptions contains options for retrieving delivery config.
type DeliveryConfigOptions struct {
	// Type is required by the API (observed in production); e.g. "store_pickup".
	Type string
	// DeliveryOptionID is required for some types (e.g. store pickup).
	DeliveryOptionID string
}

// ListDeliveryOptions retrieves a list of delivery options.
func (c *Client) ListDeliveryOptions(ctx context.Context, opts *DeliveryOptionsListOptions) (*DeliveryOptionsListResponse, error) {
	path := "/delivery_options" + NewQuery().
		Int("page", opts.Page).
		Int("page_size", opts.PageSize).
		String("status", opts.Status).
		String("type", opts.Type).
		Build()

	var resp DeliveryOptionsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetDeliveryOption retrieves a single delivery option by ID.
func (c *Client) GetDeliveryOption(ctx context.Context, id string) (*DeliveryOption, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("delivery option id is required")
	}
	var opt DeliveryOption
	if err := c.Get(ctx, fmt.Sprintf("/delivery_options/%s", id), &opt); err != nil {
		return nil, err
	}
	return &opt, nil
}

// UpdateDeliveryOptionPickupStore updates the pickup store for a delivery option.
func (c *Client) UpdateDeliveryOptionPickupStore(ctx context.Context, id string, req *PickupStoreUpdateRequest) (*DeliveryOption, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("delivery option id is required")
	}
	var opt DeliveryOption
	if err := c.Put(ctx, fmt.Sprintf("/delivery_options/%s/pickup_store", id), req, &opt); err != nil {
		return nil, err
	}
	return &opt, nil
}

// ListDeliveryTimeSlots retrieves time slots for a delivery option.
func (c *Client) ListDeliveryTimeSlots(ctx context.Context, id string, opts *DeliveryTimeSlotsListOptions) (*DeliveryTimeSlotsListResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("delivery option id is required")
	}

	q := NewQuery().
		Int("page", opts.Page).
		Int("page_size", opts.PageSize).
		String("start_date", opts.StartDate).
		String("end_date", opts.EndDate)

	path := fmt.Sprintf("/delivery_options/%s/time_slots", id) + q.Build()

	var resp DeliveryTimeSlotsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetDeliveryConfig retrieves the delivery configuration (documented endpoint).
//
// Docs: GET /delivery_options/delivery_config?type=...
func (c *Client) GetDeliveryConfig(ctx context.Context, opts *DeliveryConfigOptions) (json.RawMessage, error) {
	if opts == nil {
		return nil, fmt.Errorf("delivery config options are required")
	}
	if strings.TrimSpace(opts.Type) == "" {
		return nil, fmt.Errorf("type is required")
	}
	var resp json.RawMessage
	path := "/delivery_options/delivery_config" + NewQuery().
		String("type", opts.Type).
		String("delivery_option_id", opts.DeliveryOptionID).
		Build()
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetDeliveryTimeSlotsOpenAPI retrieves delivery time slots via the documented endpoint.
//
// Docs: GET /delivery_options/{delivery_option_id}/delivery_time_slots
func (c *Client) GetDeliveryTimeSlotsOpenAPI(ctx context.Context, id string) (json.RawMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("delivery option id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/delivery_options/%s/delivery_time_slots", id), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateDeliveryOptionStoresInfo updates store pickup stores info (documented endpoint).
//
// Docs: PUT /delivery_options/{id}/stores_info
func (c *Client) UpdateDeliveryOptionStoresInfo(ctx context.Context, id string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("delivery option id is required")
	}
	var resp json.RawMessage
	if err := c.Put(ctx, fmt.Sprintf("/delivery_options/%s/stores_info", id), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
