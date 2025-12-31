package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// CustomerAddress represents a Shopline customer address.
type CustomerAddress struct {
	ID           string    `json:"id"`
	CustomerID   string    `json:"customer_id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Company      string    `json:"company"`
	Address1     string    `json:"address1"`
	Address2     string    `json:"address2"`
	City         string    `json:"city"`
	Province     string    `json:"province"`
	ProvinceCode string    `json:"province_code"`
	Country      string    `json:"country"`
	CountryCode  string    `json:"country_code"`
	Zip          string    `json:"zip"`
	Phone        string    `json:"phone"`
	Default      bool      `json:"default"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CustomerAddressesListOptions contains options for listing customer addresses.
type CustomerAddressesListOptions struct {
	Page     int
	PageSize int
}

// CustomerAddressesListResponse is the paginated response for customer addresses.
type CustomerAddressesListResponse = ListResponse[CustomerAddress]

// CustomerAddressCreateRequest contains the request body for creating an address.
type CustomerAddressCreateRequest struct {
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	Company      string `json:"company,omitempty"`
	Address1     string `json:"address1"`
	Address2     string `json:"address2,omitempty"`
	City         string `json:"city"`
	Province     string `json:"province,omitempty"`
	ProvinceCode string `json:"province_code,omitempty"`
	Country      string `json:"country"`
	CountryCode  string `json:"country_code,omitempty"`
	Zip          string `json:"zip,omitempty"`
	Phone        string `json:"phone,omitempty"`
	Default      bool   `json:"default,omitempty"`
}

// ListCustomerAddresses retrieves a list of addresses for a customer.
func (c *Client) ListCustomerAddresses(ctx context.Context, customerID string, opts *CustomerAddressesListOptions) (*CustomerAddressesListResponse, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	path := fmt.Sprintf("/customers/%s/addresses", customerID)
	if opts != nil {
		params := url.Values{}
		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			params.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp CustomerAddressesListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCustomerAddress retrieves a single customer address by ID.
func (c *Client) GetCustomerAddress(ctx context.Context, customerID, addressID string) (*CustomerAddress, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	if strings.TrimSpace(addressID) == "" {
		return nil, fmt.Errorf("address id is required")
	}
	var address CustomerAddress
	if err := c.Get(ctx, fmt.Sprintf("/customers/%s/addresses/%s", customerID, addressID), &address); err != nil {
		return nil, err
	}
	return &address, nil
}

// CreateCustomerAddress creates a new address for a customer.
func (c *Client) CreateCustomerAddress(ctx context.Context, customerID string, req *CustomerAddressCreateRequest) (*CustomerAddress, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	var address CustomerAddress
	if err := c.Post(ctx, fmt.Sprintf("/customers/%s/addresses", customerID), req, &address); err != nil {
		return nil, err
	}
	return &address, nil
}

// DeleteCustomerAddress deletes a customer address.
func (c *Client) DeleteCustomerAddress(ctx context.Context, customerID, addressID string) error {
	if strings.TrimSpace(customerID) == "" {
		return fmt.Errorf("customer id is required")
	}
	if strings.TrimSpace(addressID) == "" {
		return fmt.Errorf("address id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/customers/%s/addresses/%s", customerID, addressID))
}

// SetDefaultCustomerAddress sets an address as the default for a customer.
func (c *Client) SetDefaultCustomerAddress(ctx context.Context, customerID, addressID string) (*CustomerAddress, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	if strings.TrimSpace(addressID) == "" {
		return nil, fmt.Errorf("address id is required")
	}
	var address CustomerAddress
	if err := c.Put(ctx, fmt.Sprintf("/customers/%s/addresses/%s/default", customerID, addressID), nil, &address); err != nil {
		return nil, err
	}
	return &address, nil
}
