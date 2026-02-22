package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Merchant represents a merchant/store in Shopline.
type Merchant struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Handle          string            `json:"handle"`
	Email           string            `json:"email"`
	Phone           string            `json:"phone"`
	Domain          string            `json:"domain"`
	PrimaryDomain   string            `json:"primary_domain"`
	Currency        string            `json:"currency"`
	Timezone        string            `json:"timezone"`
	Country         string            `json:"country"`
	CountryCode     string            `json:"country_code"`
	Province        string            `json:"province"`
	City            string            `json:"city"`
	Address1        string            `json:"address1"`
	Address2        string            `json:"address2"`
	Zip             string            `json:"zip"`
	Plan            string            `json:"plan"`
	PlanDisplayName string            `json:"plan_display_name"`
	ShopOwner       string            `json:"shop_owner"`
	WeightUnit      string            `json:"weight_unit"`
	TaxesIncluded   bool              `json:"taxes_included"`
	TaxShipping     bool              `json:"tax_shipping"`
	PasswordEnabled bool              `json:"password_enabled"`
	HasStorefront   bool              `json:"has_storefront"`
	HasDiscounts    bool              `json:"has_discounts"`
	HasGiftCards    bool              `json:"has_gift_cards"`
	SetupRequired   bool              `json:"setup_required"`
	Finances        *MerchantFinances `json:"finances"`
	Features        *MerchantFeatures `json:"features"`
	BillingAddress  *MerchantAddress  `json:"billing_address"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// MerchantFinances represents merchant financial settings.
type MerchantFinances struct {
	Currency                     string   `json:"currency"`
	MoneyFormat                  string   `json:"money_format"`
	MoneyWithCurrencyFormat      string   `json:"money_with_currency_format"`
	MoneyInEmailsFormat          string   `json:"money_in_emails_format"`
	SetupRequired                bool     `json:"setup_required"`
	EnabledPresentmentCurrencies []string `json:"enabled_presentment_currencies"`
}

// MerchantFeatures represents enabled merchant features.
type MerchantFeatures struct {
	Checkout                bool `json:"checkout"`
	MultiLocation           bool `json:"multi_location"`
	MultiCurrency           bool `json:"multi_currency"`
	GiftCards               bool `json:"gift_cards"`
	Subscriptions           bool `json:"subscriptions"`
	BuyOnline               bool `json:"buy_online"`
	PickupInStore           bool `json:"pickup_in_store"`
	LocalDelivery           bool `json:"local_delivery"`
	InternationalDomains    bool `json:"international_domains"`
	InternationalPriceRules bool `json:"international_price_rules"`
	Discounts               bool `json:"discounts"`
}

// MerchantAddress represents a merchant's address.
type MerchantAddress struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Company      string `json:"company"`
	Address1     string `json:"address1"`
	Address2     string `json:"address2"`
	City         string `json:"city"`
	Province     string `json:"province"`
	ProvinceCode string `json:"province_code"`
	Country      string `json:"country"`
	CountryCode  string `json:"country_code"`
	Zip          string `json:"zip"`
	Phone        string `json:"phone"`
}

// MerchantStaff represents a staff member of a merchant.
type MerchantStaff struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	Phone        string     `json:"phone"`
	Role         string     `json:"role"`
	Permissions  []string   `json:"permissions"`
	AccountOwner bool       `json:"account_owner"`
	Active       bool       `json:"active"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// MerchantStaffListOptions contains options for listing merchant staff.
type MerchantStaffListOptions struct {
	Page     int
	PageSize int
	Role     string
	Active   *bool
}

// MerchantStaffListResponse is the paginated response for merchant staff.
type MerchantStaffListResponse = ListResponse[MerchantStaff]

// MerchantsListResponse wraps the merchants list response.
type MerchantsListResponse struct {
	Items []Merchant `json:"items"`
}

// GetMerchant retrieves the current merchant information.
// Note: The API returns {"items": [...]} with the merchant in the items array.
func (c *Client) GetMerchant(ctx context.Context) (*Merchant, error) {
	var resp MerchantsListResponse
	if err := c.Get(ctx, "/merchants", &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("no merchant found")
	}
	return &resp.Items[0], nil
}

// ListMerchants retrieves all merchants (for multi-merchant accounts).
func (c *Client) ListMerchants(ctx context.Context) ([]Merchant, error) {
	var resp MerchantsListResponse
	if err := c.Get(ctx, "/merchants", &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// ListMerchantStaff retrieves a list of merchant staff members.
func (c *Client) ListMerchantStaff(ctx context.Context, opts *MerchantStaffListOptions) (*MerchantStaffListResponse, error) {
	path := "/merchant/staff"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("role", opts.Role).
			BoolPtr("active", opts.Active).
			Build()
	}

	var resp MerchantStaffListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetMerchantStaff retrieves a single staff member by ID.
func (c *Client) GetMerchantStaff(ctx context.Context, id string) (*MerchantStaff, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("staff id is required")
	}
	var staff MerchantStaff
	if err := c.Get(ctx, fmt.Sprintf("/merchant/staff/%s", id), &staff); err != nil {
		return nil, err
	}
	return &staff, nil
}
