package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// GiftCardStatus represents the status of a gift card.
type GiftCardStatus string

const (
	GiftCardStatusEnabled  GiftCardStatus = "enabled"
	GiftCardStatusDisabled GiftCardStatus = "disabled"
)

// GiftCard represents a gift card.
type GiftCard struct {
	ID           string         `json:"id"`
	Code         string         `json:"code"`
	MaskedCode   string         `json:"masked_code"`
	InitialValue string         `json:"initial_value"`
	Balance      string         `json:"balance"`
	Currency     string         `json:"currency"`
	Status       GiftCardStatus `json:"status"`
	CustomerID   string         `json:"customer_id"`
	Note         string         `json:"note"`
	ExpiresAt    time.Time      `json:"expires_at"`
	DisabledAt   time.Time      `json:"disabled_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// GiftCardsListOptions contains options for listing gift cards.
type GiftCardsListOptions struct {
	Page       int
	PageSize   int
	Status     string
	CustomerID string
}

// GiftCardsListResponse is the paginated response for gift cards.
type GiftCardsListResponse = ListResponse[GiftCard]

// GiftCardCreateRequest contains the data for creating a gift card.
type GiftCardCreateRequest struct {
	InitialValue string     `json:"initial_value"`
	Currency     string     `json:"currency,omitempty"`
	Code         string     `json:"code,omitempty"`
	CustomerID   string     `json:"customer_id,omitempty"`
	Note         string     `json:"note,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

// ListGiftCards retrieves a list of gift cards.
func (c *Client) ListGiftCards(ctx context.Context, opts *GiftCardsListOptions) (*GiftCardsListResponse, error) {
	path := "/gift_cards"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("status", opts.Status).
			String("customer_id", opts.CustomerID).
			Build()
	}

	var resp GiftCardsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetGiftCard retrieves a single gift card by ID.
func (c *Client) GetGiftCard(ctx context.Context, id string) (*GiftCard, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("gift card id is required")
	}
	var giftCard GiftCard
	if err := c.Get(ctx, fmt.Sprintf("/gift_cards/%s", id), &giftCard); err != nil {
		return nil, err
	}
	return &giftCard, nil
}

// CreateGiftCard creates a new gift card.
func (c *Client) CreateGiftCard(ctx context.Context, req *GiftCardCreateRequest) (*GiftCard, error) {
	var giftCard GiftCard
	if err := c.Post(ctx, "/gift_cards", req, &giftCard); err != nil {
		return nil, err
	}
	return &giftCard, nil
}

// DeleteGiftCard disables a gift card.
func (c *Client) DeleteGiftCard(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("gift card id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/gift_cards/%s", id))
}
