package api

import (
	"context"
	"fmt"
	"strings"
)

// Country represents a Shopline country.
type Country struct {
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	Tax       float64    `json:"tax"`
	TaxName   string     `json:"tax_name"`
	Provinces []Province `json:"provinces"`
}

// Province represents a province/state within a country.
type Province struct {
	Code    string  `json:"code"`
	Name    string  `json:"name"`
	Tax     float64 `json:"tax"`
	TaxName string  `json:"tax_name"`
	TaxType string  `json:"tax_type"`
}

// CountriesListResponse contains the list response.
type CountriesListResponse struct {
	Items []Country `json:"items"`
}

// ListCountries retrieves a list of countries.
func (c *Client) ListCountries(ctx context.Context) (*CountriesListResponse, error) {
	var resp CountriesListResponse
	if err := c.Get(ctx, "/countries", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCountry retrieves a single country by code.
func (c *Client) GetCountry(ctx context.Context, code string) (*Country, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("country code is required")
	}
	var country Country
	if err := c.Get(ctx, fmt.Sprintf("/countries/%s", code), &country); err != nil {
		return nil, err
	}
	return &country, nil
}
