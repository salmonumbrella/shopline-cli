package api

import (
	"context"
	"encoding/json"
)

// NOTE: Shopline has multiple "settings" surfaces:
// - /settings (user-specific settings) is implemented in settings.go (typed).
// - /shop/settings and /checkout_settings are implemented elsewhere (typed).
// - /settings/* endpoints below are documented in the Open API reference but often
//   have large/unstable schemas. We model them as json.RawMessage for resiliency.

// Docs: GET /settings/checkout
func (c *Client) GetSettingsCheckout(ctx context.Context) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/settings/checkout", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Docs: GET /settings/domains
func (c *Client) GetSettingsDomains(ctx context.Context) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/settings/domains", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Docs: PUT /settings/domains
func (c *Client) UpdateSettingsDomains(ctx context.Context, body any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Put(ctx, "/settings/domains", body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Docs: GET /settings/layouts
func (c *Client) GetSettingsLayouts(ctx context.Context) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/settings/layouts", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Docs: GET /settings/layouts/draft
func (c *Client) GetSettingsLayoutsDraft(ctx context.Context) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/settings/layouts/draft", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Docs: PUT /settings/layouts/draft
func (c *Client) UpdateSettingsLayoutsDraft(ctx context.Context, body any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Put(ctx, "/settings/layouts/draft", body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Docs: POST /settings/layouts/publish
func (c *Client) PublishSettingsLayouts(ctx context.Context, body any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Post(ctx, "/settings/layouts/publish", body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Docs: GET /settings/orders
func (c *Client) GetSettingsOrders(ctx context.Context) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/settings/orders", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Docs: GET /settings/payments
func (c *Client) GetSettingsPayments(ctx context.Context) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/settings/payments", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Docs: GET /settings/pos
func (c *Client) GetSettingsPOS(ctx context.Context) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/settings/pos", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Docs: GET /settings/product_review
func (c *Client) GetSettingsProductReview(ctx context.Context) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/settings/product_review", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Docs: GET /settings/products
func (c *Client) GetSettingsProducts(ctx context.Context) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/settings/products", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Docs: GET /settings/promotions
func (c *Client) GetSettingsPromotions(ctx context.Context) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/settings/promotions", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Docs: GET /settings/shop
func (c *Client) GetSettingsShop(ctx context.Context) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/settings/shop", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Docs: GET /settings/tax
func (c *Client) GetSettingsTax(ctx context.Context) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/settings/tax", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Docs: GET /settings/theme
func (c *Client) GetSettingsTheme(ctx context.Context) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/settings/theme", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Docs: GET /settings/theme/draft
func (c *Client) GetSettingsThemeDraft(ctx context.Context) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/settings/theme/draft", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Docs: PUT /settings/theme/draft
func (c *Client) UpdateSettingsThemeDraft(ctx context.Context, body any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Put(ctx, "/settings/theme/draft", body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Docs: POST /settings/theme/publish
func (c *Client) PublishSettingsTheme(ctx context.Context, body any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Post(ctx, "/settings/theme/publish", body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Docs: GET /settings/third_party_ads
func (c *Client) GetSettingsThirdPartyAds(ctx context.Context) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/settings/third_party_ads", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Docs: GET /settings/users
func (c *Client) GetSettingsUsers(ctx context.Context) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/settings/users", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
