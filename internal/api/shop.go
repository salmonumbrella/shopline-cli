package api

import (
	"context"
	"time"
)

// Shop represents the Shopline shop details.
type Shop struct {
	ID                      string    `json:"id"`
	Name                    string    `json:"name"`
	Email                   string    `json:"email"`
	Domain                  string    `json:"domain"`
	MyshoplineDomain        string    `json:"myshopline_domain"`
	Phone                   string    `json:"phone"`
	Address1                string    `json:"address1"`
	Address2                string    `json:"address2"`
	City                    string    `json:"city"`
	Province                string    `json:"province"`
	ProvinceCode            string    `json:"province_code"`
	Country                 string    `json:"country"`
	CountryCode             string    `json:"country_code"`
	Zip                     string    `json:"zip"`
	Currency                string    `json:"currency"`
	MoneyFormat             string    `json:"money_format"`
	MoneyWithCurrencyFormat string    `json:"money_with_currency_format"`
	Timezone                string    `json:"timezone"`
	IanaTimezone            string    `json:"iana_timezone"`
	WeightUnit              string    `json:"weight_unit"`
	PlanName                string    `json:"plan_name"`
	PlanDisplayName         string    `json:"plan_display_name"`
	ShopOwner               string    `json:"shop_owner"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

// ShopSettings represents configurable shop settings.
type ShopSettings struct {
	ID                           string   `json:"id"`
	Currency                     string   `json:"currency"`
	WeightUnit                   string   `json:"weight_unit"`
	Timezone                     string   `json:"timezone"`
	OrderPrefix                  string   `json:"order_prefix"`
	OrderSuffix                  string   `json:"order_suffix"`
	TaxesIncluded                bool     `json:"taxes_included"`
	TaxShipping                  bool     `json:"tax_shipping"`
	AutomaticFulfillment         bool     `json:"automatic_fulfillment"`
	EnabledPresentmentCurrencies []string `json:"enabled_presentment_currencies"`
}

// ShopSettingsUpdateRequest contains the request body for updating settings.
type ShopSettingsUpdateRequest struct {
	Currency             string `json:"currency,omitempty"`
	WeightUnit           string `json:"weight_unit,omitempty"`
	Timezone             string `json:"timezone,omitempty"`
	OrderPrefix          string `json:"order_prefix,omitempty"`
	OrderSuffix          string `json:"order_suffix,omitempty"`
	TaxesIncluded        *bool  `json:"taxes_included,omitempty"`
	TaxShipping          *bool  `json:"tax_shipping,omitempty"`
	AutomaticFulfillment *bool  `json:"automatic_fulfillment,omitempty"`
}

// GetShop retrieves the current shop details.
func (c *Client) GetShop(ctx context.Context) (*Shop, error) {
	var shop Shop
	if err := c.Get(ctx, "/shop", &shop); err != nil {
		return nil, err
	}
	return &shop, nil
}

// GetShopSettings retrieves the current shop settings.
func (c *Client) GetShopSettings(ctx context.Context) (*ShopSettings, error) {
	var settings ShopSettings
	if err := c.Get(ctx, "/shop/settings", &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

// UpdateShopSettings updates the shop settings.
func (c *Client) UpdateShopSettings(ctx context.Context, req *ShopSettingsUpdateRequest) (*ShopSettings, error) {
	var settings ShopSettings
	if err := c.Put(ctx, "/shop/settings", req, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}
