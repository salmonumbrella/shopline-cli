package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// PickupHours represents operating hours for a pickup location.
type PickupHours struct {
	Day       string `json:"day"`
	OpenTime  string `json:"open_time"`
	CloseTime string `json:"close_time"`
	Closed    bool   `json:"closed"`
}

// PickupLocation represents a store pickup location.
type PickupLocation struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Address1     string        `json:"address1"`
	Address2     string        `json:"address2"`
	City         string        `json:"city"`
	Province     string        `json:"province"`
	Country      string        `json:"country"`
	CountryCode  string        `json:"country_code"`
	ZipCode      string        `json:"zip_code"`
	Phone        string        `json:"phone"`
	Email        string        `json:"email"`
	Active       bool          `json:"active"`
	Instructions string        `json:"instructions"`
	Hours        []PickupHours `json:"hours"`
	LocationID   string        `json:"location_id"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// PickupListOptions contains options for listing pickup locations.
type PickupListOptions struct {
	Page       int
	PageSize   int
	Active     *bool
	LocationID string
}

// PickupListResponse is the paginated response for pickup locations.
type PickupListResponse = ListResponse[PickupLocation]

// PickupCreateRequest contains the request body for creating a pickup location.
type PickupCreateRequest struct {
	Name         string        `json:"name"`
	Address1     string        `json:"address1"`
	Address2     string        `json:"address2,omitempty"`
	City         string        `json:"city"`
	Province     string        `json:"province,omitempty"`
	Country      string        `json:"country"`
	CountryCode  string        `json:"country_code,omitempty"`
	ZipCode      string        `json:"zip_code,omitempty"`
	Phone        string        `json:"phone,omitempty"`
	Email        string        `json:"email,omitempty"`
	Active       bool          `json:"active"`
	Instructions string        `json:"instructions,omitempty"`
	Hours        []PickupHours `json:"hours,omitempty"`
	LocationID   string        `json:"location_id,omitempty"`
}

// PickupUpdateRequest contains the request body for updating a pickup location.
type PickupUpdateRequest struct {
	Name         *string        `json:"name,omitempty"`
	Address1     *string        `json:"address1,omitempty"`
	Address2     *string        `json:"address2,omitempty"`
	City         *string        `json:"city,omitempty"`
	Province     *string        `json:"province,omitempty"`
	Country      *string        `json:"country,omitempty"`
	CountryCode  *string        `json:"country_code,omitempty"`
	ZipCode      *string        `json:"zip_code,omitempty"`
	Phone        *string        `json:"phone,omitempty"`
	Email        *string        `json:"email,omitempty"`
	Active       *bool          `json:"active,omitempty"`
	Instructions *string        `json:"instructions,omitempty"`
	Hours        *[]PickupHours `json:"hours,omitempty"`
}

// ListPickupLocations retrieves a list of pickup locations.
func (c *Client) ListPickupLocations(ctx context.Context, opts *PickupListOptions) (*PickupListResponse, error) {
	path := "/pickup"
	if opts != nil {
		q := NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("location_id", opts.LocationID).
			BoolPtr("active", opts.Active)
		path += q.Build()
	}

	var resp PickupListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPickupLocation retrieves a single pickup location by ID.
func (c *Client) GetPickupLocation(ctx context.Context, id string) (*PickupLocation, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("pickup location id is required")
	}
	var location PickupLocation
	if err := c.Get(ctx, fmt.Sprintf("/pickup/%s", id), &location); err != nil {
		return nil, err
	}
	return &location, nil
}

// CreatePickupLocation creates a new pickup location.
func (c *Client) CreatePickupLocation(ctx context.Context, req *PickupCreateRequest) (*PickupLocation, error) {
	var location PickupLocation
	if err := c.Post(ctx, "/pickup", req, &location); err != nil {
		return nil, err
	}
	return &location, nil
}

// UpdatePickupLocation updates an existing pickup location.
func (c *Client) UpdatePickupLocation(ctx context.Context, id string, req *PickupUpdateRequest) (*PickupLocation, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("pickup location id is required")
	}
	var location PickupLocation
	if err := c.Put(ctx, fmt.Sprintf("/pickup/%s", id), req, &location); err != nil {
		return nil, err
	}
	return &location, nil
}

// DeletePickupLocation deletes a pickup location.
func (c *Client) DeletePickupLocation(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("pickup location id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/pickup/%s", id))
}
