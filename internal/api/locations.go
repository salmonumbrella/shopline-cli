package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Location represents a Shopline location.
type Location struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Address1    string    `json:"address1"`
	Address2    string    `json:"address2"`
	City        string    `json:"city"`
	Province    string    `json:"province"`
	Country     string    `json:"country"`
	CountryCode string    `json:"country_code"`
	Zip         string    `json:"zip"`
	Phone       string    `json:"phone"`
	Active      bool      `json:"active"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LocationsListOptions contains options for listing locations.
type LocationsListOptions struct {
	Page     int
	PageSize int
	Active   *bool
}

// LocationsListResponse is the paginated response for locations.
type LocationsListResponse = ListResponse[Location]

// LocationCreateRequest contains the request body for creating a location.
type LocationCreateRequest struct {
	Name        string `json:"name"`
	Address1    string `json:"address1"`
	Address2    string `json:"address2,omitempty"`
	City        string `json:"city"`
	Province    string `json:"province,omitempty"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code,omitempty"`
	Zip         string `json:"zip,omitempty"`
	Phone       string `json:"phone,omitempty"`
}

// LocationUpdateRequest contains the request body for updating a location.
type LocationUpdateRequest struct {
	Name     string `json:"name,omitempty"`
	Address1 string `json:"address1,omitempty"`
	Address2 string `json:"address2,omitempty"`
	City     string `json:"city,omitempty"`
	Province string `json:"province,omitempty"`
	Country  string `json:"country,omitempty"`
	Zip      string `json:"zip,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Active   *bool  `json:"active,omitempty"`
}

// ListLocations retrieves a list of locations.
func (c *Client) ListLocations(ctx context.Context, opts *LocationsListOptions) (*LocationsListResponse, error) {
	path := "/locations"
	if opts != nil {
		params := url.Values{}
		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			params.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.Active != nil {
			params.Set("active", strconv.FormatBool(*opts.Active))
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp LocationsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetLocation retrieves a single location by ID.
func (c *Client) GetLocation(ctx context.Context, id string) (*Location, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("location id is required")
	}
	var location Location
	if err := c.Get(ctx, fmt.Sprintf("/locations/%s", id), &location); err != nil {
		return nil, err
	}
	return &location, nil
}

// CreateLocation creates a new location.
func (c *Client) CreateLocation(ctx context.Context, req *LocationCreateRequest) (*Location, error) {
	var location Location
	if err := c.Post(ctx, "/locations", req, &location); err != nil {
		return nil, err
	}
	return &location, nil
}

// UpdateLocation updates an existing location.
func (c *Client) UpdateLocation(ctx context.Context, id string, req *LocationUpdateRequest) (*Location, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("location id is required")
	}
	var location Location
	if err := c.Put(ctx, fmt.Sprintf("/locations/%s", id), req, &location); err != nil {
		return nil, err
	}
	return &location, nil
}

// DeleteLocation deletes a location.
func (c *Client) DeleteLocation(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("location id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/locations/%s", id))
}
