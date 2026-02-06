package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Currency represents a Shopline currency.
type Currency struct {
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	Symbol       string    `json:"symbol"`
	Primary      bool      `json:"primary"`
	Enabled      bool      `json:"enabled"`
	RoundingMode string    `json:"rounding_mode"`
	ExchangeRate float64   `json:"exchange_rate"`
	AutoUpdate   bool      `json:"auto_update"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CurrenciesListResponse contains the list response.
type CurrenciesListResponse struct {
	Items []Currency `json:"items"`
}

// CurrencyUpdateRequest contains the request body for updating a currency.
type CurrencyUpdateRequest struct {
	Enabled      *bool    `json:"enabled,omitempty"`
	ExchangeRate *float64 `json:"exchange_rate,omitempty"`
	RoundingMode string   `json:"rounding_mode,omitempty"`
	AutoUpdate   *bool    `json:"auto_update,omitempty"`
}

// ListCurrencies retrieves a list of currencies.
func (c *Client) ListCurrencies(ctx context.Context) (*CurrenciesListResponse, error) {
	var resp CurrenciesListResponse
	if err := c.Get(ctx, "/currencies", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCurrency retrieves a single currency by code.
func (c *Client) GetCurrency(ctx context.Context, code string) (*Currency, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("currency code is required")
	}
	var currency Currency
	if err := c.Get(ctx, fmt.Sprintf("/currencies/%s", code), &currency); err != nil {
		return nil, err
	}
	return &currency, nil
}

// UpdateCurrency updates a currency.
func (c *Client) UpdateCurrency(ctx context.Context, code string, req *CurrencyUpdateRequest) (*Currency, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("currency code is required")
	}
	var currency Currency
	if err := c.Put(ctx, fmt.Sprintf("/currencies/%s", code), req, &currency); err != nil {
		return nil, err
	}
	return &currency, nil
}
