package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// StorefrontPromotion represents a promotion displayed in the storefront.
type StorefrontPromotion struct {
	ID               string                       `json:"id"`
	Title            string                       `json:"title"`
	Description      string                       `json:"description"`
	Type             string                       `json:"type"`
	Status           string                       `json:"status"`
	DiscountType     string                       `json:"discount_type"`
	DiscountValue    string                       `json:"discount_value"`
	MinPurchase      string                       `json:"min_purchase"`
	MaxDiscount      string                       `json:"max_discount"`
	UsageLimit       int                          `json:"usage_limit"`
	UsageCount       int                          `json:"usage_count"`
	CustomerLimit    int                          `json:"customer_limit"`
	Stackable        bool                         `json:"stackable"`
	AutoApply        bool                         `json:"auto_apply"`
	Code             string                       `json:"code"`
	TargetType       string                       `json:"target_type"`
	TargetProducts   []string                     `json:"target_products"`
	TargetCategories []string                     `json:"target_categories"`
	ExcludedProducts []string                     `json:"excluded_products"`
	Banner           *StorefrontPromotionBanner   `json:"banner"`
	Schedule         *StorefrontPromotionSchedule `json:"schedule"`
	StartsAt         time.Time                    `json:"starts_at"`
	EndsAt           *time.Time                   `json:"ends_at"`
	CreatedAt        time.Time                    `json:"created_at"`
	UpdatedAt        time.Time                    `json:"updated_at"`
}

// StorefrontPromotionBanner represents promotion banner configuration.
type StorefrontPromotionBanner struct {
	Enabled         bool   `json:"enabled"`
	Text            string `json:"text"`
	ImageURL        string `json:"image_url"`
	LinkURL         string `json:"link_url"`
	Position        string `json:"position"`
	BackgroundColor string `json:"background_color"`
	TextColor       string `json:"text_color"`
}

// StorefrontPromotionSchedule represents promotion schedule.
type StorefrontPromotionSchedule struct {
	Recurring  bool     `json:"recurring"`
	DaysOfWeek []string `json:"days_of_week"`
	StartTime  string   `json:"start_time"`
	EndTime    string   `json:"end_time"`
	Timezone   string   `json:"timezone"`
}

// StorefrontPromotionsListOptions contains options for listing storefront promotions.
type StorefrontPromotionsListOptions struct {
	Page         int
	PageSize     int
	Status       string
	Type         string
	DiscountType string
	Active       *bool
	AutoApply    *bool
	SortBy       string
	SortOrder    string
}

// StorefrontPromotionsListResponse is the paginated response for storefront promotions.
type StorefrontPromotionsListResponse = ListResponse[StorefrontPromotion]

// ListStorefrontPromotions retrieves a list of storefront promotions.
func (c *Client) ListStorefrontPromotions(ctx context.Context, opts *StorefrontPromotionsListOptions) (*StorefrontPromotionsListResponse, error) {
	path := "/storefront/promotions"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("status", opts.Status).
			String("type", opts.Type).
			String("discount_type", opts.DiscountType).
			BoolPtr("active", opts.Active).
			BoolPtr("auto_apply", opts.AutoApply).
			String("sort_by", opts.SortBy).
			String("sort_order", opts.SortOrder).
			Build()
	}

	var resp StorefrontPromotionsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetStorefrontPromotion retrieves a single storefront promotion by ID.
func (c *Client) GetStorefrontPromotion(ctx context.Context, id string) (*StorefrontPromotion, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("promotion id is required")
	}
	var promo StorefrontPromotion
	if err := c.Get(ctx, fmt.Sprintf("/storefront/promotions/%s", id), &promo); err != nil {
		return nil, err
	}
	return &promo, nil
}

// GetStorefrontPromotionByCode retrieves a storefront promotion by code.
func (c *Client) GetStorefrontPromotionByCode(ctx context.Context, code string) (*StorefrontPromotion, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("promotion code is required")
	}
	var promo StorefrontPromotion
	if err := c.Get(ctx, fmt.Sprintf("/storefront/promotions/code/%s", code), &promo); err != nil {
		return nil, err
	}
	return &promo, nil
}
