package api

import (
	"context"
	"time"
)

// Settings represents the store settings.
// Note: Store settings are available via /merchants endpoint.
// The /settings endpoint returns user-specific settings.
type Settings struct {
	ID                           string    `json:"id"`
	Name                         string    `json:"name"`
	Email                        string    `json:"email"`
	Domain                       string    `json:"domain"`
	Currency                     string    `json:"currency"`
	Timezone                     string    `json:"timezone"`
	IanaTimezone                 string    `json:"iana_timezone"`
	WeightUnit                   string    `json:"weight_unit"`
	LengthUnit                   string    `json:"length_unit"`
	MoneyFormat                  string    `json:"money_format"`
	MoneyWithCurrencyFormat      string    `json:"money_with_currency_format"`
	TaxesIncluded                bool      `json:"taxes_included"`
	TaxShipping                  bool      `json:"tax_shipping"`
	CountryCode                  string    `json:"country_code"`
	ProvinceCode                 string    `json:"province_code"`
	Address1                     string    `json:"address1"`
	Address2                     string    `json:"address2"`
	City                         string    `json:"city"`
	Zip                          string    `json:"zip"`
	Phone                        string    `json:"phone"`
	PrimaryLocale                string    `json:"primary_locale"`
	OrderNumberFormat            string    `json:"order_number_format"`
	OrderPrefix                  string    `json:"order_prefix"`
	OrderSuffix                  string    `json:"order_suffix"`
	PasswordEnabled              bool      `json:"password_enabled"`
	SetupRequired                bool      `json:"setup_required"`
	EnabledPresentmentCurrencies []string  `json:"enabled_presentment_currencies"`
	CreatedAt                    time.Time `json:"created_at"`
	UpdatedAt                    time.Time `json:"updated_at"`
}

// UserSettings represents user-specific settings from /settings endpoint.
type UserSettings struct {
	PosApplyCredit  bool   `json:"pos_apply_credit"`
	MinimumAgeLimit string `json:"minimum_age_limit"`
}

// SettingsResponse wraps the settings API response.
type SettingsResponse struct {
	Users UserSettings `json:"users"`
}

// UserSettingsUpdateRequest contains the request body for updating user settings via /settings endpoint.
// Note: Store settings (name, currency, etc.) are updated via /merchants endpoint.
type UserSettingsUpdateRequest struct {
	Users UserSettingsUpdate `json:"users"`
}

// UserSettingsUpdate contains the user-specific settings fields that can be updated.
type UserSettingsUpdate struct {
	PosApplyCredit  *bool  `json:"pos_apply_credit,omitempty"`
	MinimumAgeLimit string `json:"minimum_age_limit,omitempty"`
}

// GetSettings retrieves user-specific settings.
// Note: For store settings (name, currency, etc.), use GetMerchant() instead.
func (c *Client) GetSettings(ctx context.Context) (*SettingsResponse, error) {
	var resp SettingsResponse
	if err := c.Get(ctx, "/settings", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateSettings updates user-specific settings via /settings endpoint.
// Note: Store settings (name, currency, etc.) are updated via /merchants endpoint.
func (c *Client) UpdateSettings(ctx context.Context, req *UserSettingsUpdateRequest) (*SettingsResponse, error) {
	var resp SettingsResponse
	if err := c.Put(ctx, "/settings", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
