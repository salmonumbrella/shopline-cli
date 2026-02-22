package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// NOTE: The Shopline Open API "carts" endpoints are distinct from the Storefront API
// carts endpoints (which live under /storefront/carts). These Open API endpoints
// typically operate on carts created in admin/open-api contexts.
//
// The reference docs are not always consistent about request/response schemas here,
// so we model responses as json.RawMessage and accept "any" request bodies.

// ExchangeCart exchanges a cart (e.g. currency/market/session exchange).
//
// Docs: POST /carts/exchange
func (c *Client) ExchangeCart(ctx context.Context, body any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Post(ctx, "/carts/exchange", body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// PrepareCart prepares a cart for checkout / further operations.
//
// Docs: POST /carts/{cart_id}/prepare
func (c *Client) PrepareCart(ctx context.Context, cartID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(cartID) == "" {
		return nil, fmt.Errorf("cart id is required")
	}
	var resp json.RawMessage
	if err := c.Post(ctx, fmt.Sprintf("/carts/%s/prepare", cartID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// AddCartItems adds items to a cart.
//
// Docs: POST /carts/{cart_id}/items
func (c *Client) AddCartItems(ctx context.Context, cartID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(cartID) == "" {
		return nil, fmt.Errorf("cart id is required")
	}
	var resp json.RawMessage
	if err := c.Post(ctx, fmt.Sprintf("/carts/%s/items", cartID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateCartItems updates items in a cart.
//
// Docs: PATCH /carts/{cart_id}/items
func (c *Client) UpdateCartItems(ctx context.Context, cartID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(cartID) == "" {
		return nil, fmt.Errorf("cart id is required")
	}
	var resp json.RawMessage
	if err := c.Patch(ctx, fmt.Sprintf("/carts/%s/items", cartID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// DeleteCartItems deletes items from a cart.
//
// Docs: DELETE /carts/{cart_id}/items
func (c *Client) DeleteCartItems(ctx context.Context, cartID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(cartID) == "" {
		return nil, fmt.Errorf("cart id is required")
	}
	var resp json.RawMessage
	if err := c.DeleteWithBody(ctx, fmt.Sprintf("/carts/%s/items", cartID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
