package api

import (
	"context"
	"time"
)

// CheckoutSettings represents checkout configuration.
type CheckoutSettings struct {
	ID                     string    `json:"id"`
	RequirePhone           bool      `json:"require_phone"`
	RequireShippingAddress bool      `json:"require_shipping_address"`
	RequireBillingAddress  bool      `json:"require_billing_address"`
	RequireCompany         bool      `json:"require_company"`
	RequireFullName        bool      `json:"require_full_name"`
	EnableGuestCheckout    bool      `json:"enable_guest_checkout"`
	EnableExpressCheckout  bool      `json:"enable_express_checkout"`
	EnableOrderNotes       bool      `json:"enable_order_notes"`
	EnableTipping          bool      `json:"enable_tipping"`
	TippingOptions         []float64 `json:"tipping_options"`
	DefaultTippingOption   float64   `json:"default_tipping_option"`
	AutoFulfillDigital     bool      `json:"auto_fulfill_digital"`
	EnableAddressAutofill  bool      `json:"enable_address_autofill"`
	EnableMultiCurrency    bool      `json:"enable_multi_currency"`
	AbandonedCartEnabled   bool      `json:"abandoned_cart_enabled"`
	AbandonedCartDelay     int       `json:"abandoned_cart_delay"`
	TermsOfServiceURL      string    `json:"terms_of_service_url"`
	PrivacyPolicyURL       string    `json:"privacy_policy_url"`
	RefundPolicyURL        string    `json:"refund_policy_url"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// CheckoutSettingsUpdateRequest contains the request body for updating checkout settings.
type CheckoutSettingsUpdateRequest struct {
	RequirePhone           *bool     `json:"require_phone,omitempty"`
	RequireShippingAddress *bool     `json:"require_shipping_address,omitempty"`
	RequireBillingAddress  *bool     `json:"require_billing_address,omitempty"`
	RequireCompany         *bool     `json:"require_company,omitempty"`
	RequireFullName        *bool     `json:"require_full_name,omitempty"`
	EnableGuestCheckout    *bool     `json:"enable_guest_checkout,omitempty"`
	EnableExpressCheckout  *bool     `json:"enable_express_checkout,omitempty"`
	EnableOrderNotes       *bool     `json:"enable_order_notes,omitempty"`
	EnableTipping          *bool     `json:"enable_tipping,omitempty"`
	TippingOptions         []float64 `json:"tipping_options,omitempty"`
	DefaultTippingOption   float64   `json:"default_tipping_option,omitempty"`
	AutoFulfillDigital     *bool     `json:"auto_fulfill_digital,omitempty"`
	EnableAddressAutofill  *bool     `json:"enable_address_autofill,omitempty"`
	EnableMultiCurrency    *bool     `json:"enable_multi_currency,omitempty"`
	AbandonedCartEnabled   *bool     `json:"abandoned_cart_enabled,omitempty"`
	AbandonedCartDelay     int       `json:"abandoned_cart_delay,omitempty"`
	TermsOfServiceURL      string    `json:"terms_of_service_url,omitempty"`
	PrivacyPolicyURL       string    `json:"privacy_policy_url,omitempty"`
	RefundPolicyURL        string    `json:"refund_policy_url,omitempty"`
}

// GetCheckoutSettings retrieves the checkout settings.
func (c *Client) GetCheckoutSettings(ctx context.Context) (*CheckoutSettings, error) {
	var settings CheckoutSettings
	if err := c.Get(ctx, "/checkout_settings", &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

// UpdateCheckoutSettings updates the checkout settings.
func (c *Client) UpdateCheckoutSettings(ctx context.Context, req *CheckoutSettingsUpdateRequest) (*CheckoutSettings, error) {
	var settings CheckoutSettings
	if err := c.Put(ctx, "/checkout_settings", req, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}
