package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Cart item metafields (non-app-scoped)

// ListCartItemMetafields lists metafields for items in a cart.
//
// Docs: GET /carts/{cart_id}/items/metafields
func (c *Client) ListCartItemMetafields(ctx context.Context, cartID string) (json.RawMessage, error) {
	if strings.TrimSpace(cartID) == "" {
		return nil, fmt.Errorf("cart id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/carts/%s/items/metafields", cartID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// BulkCreateCartItemMetafields bulk creates metafields for cart items.
//
// Docs: POST /carts/{cart_id}/items/metafields/bulk
func (c *Client) BulkCreateCartItemMetafields(ctx context.Context, cartID string, body any) error {
	if strings.TrimSpace(cartID) == "" {
		return fmt.Errorf("cart id is required")
	}
	return c.Post(ctx, fmt.Sprintf("/carts/%s/items/metafields/bulk", cartID), body, nil)
}

// BulkUpdateCartItemMetafields bulk updates metafields for cart items.
//
// Docs: PUT /carts/{cart_id}/items/metafields/bulk
func (c *Client) BulkUpdateCartItemMetafields(ctx context.Context, cartID string, body any) error {
	if strings.TrimSpace(cartID) == "" {
		return fmt.Errorf("cart id is required")
	}
	return c.Put(ctx, fmt.Sprintf("/carts/%s/items/metafields/bulk", cartID), body, nil)
}

// BulkDeleteCartItemMetafields bulk deletes metafields for cart items.
//
// Docs: DELETE /carts/{cart_id}/items/metafields/bulk
func (c *Client) BulkDeleteCartItemMetafields(ctx context.Context, cartID string, body any) error {
	if strings.TrimSpace(cartID) == "" {
		return fmt.Errorf("cart id is required")
	}
	return c.DeleteWithBody(ctx, fmt.Sprintf("/carts/%s/items/metafields/bulk", cartID), body, nil)
}

// Cart item app metafields (app-scoped)

// ListCartItemAppMetafields lists app metafields for items in a cart.
//
// Docs: GET /carts/{cart_id}/items/app_metafields
func (c *Client) ListCartItemAppMetafields(ctx context.Context, cartID string) (json.RawMessage, error) {
	if strings.TrimSpace(cartID) == "" {
		return nil, fmt.Errorf("cart id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/carts/%s/items/app_metafields", cartID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// BulkCreateCartItemAppMetafields bulk creates app metafields for cart items.
//
// Docs: POST /carts/{cart_id}/items/app_metafields/bulk
func (c *Client) BulkCreateCartItemAppMetafields(ctx context.Context, cartID string, body any) error {
	if strings.TrimSpace(cartID) == "" {
		return fmt.Errorf("cart id is required")
	}
	return c.Post(ctx, fmt.Sprintf("/carts/%s/items/app_metafields/bulk", cartID), body, nil)
}

// BulkUpdateCartItemAppMetafields bulk updates app metafields for cart items.
//
// Docs: PUT /carts/{cart_id}/items/app_metafields/bulk
func (c *Client) BulkUpdateCartItemAppMetafields(ctx context.Context, cartID string, body any) error {
	if strings.TrimSpace(cartID) == "" {
		return fmt.Errorf("cart id is required")
	}
	return c.Put(ctx, fmt.Sprintf("/carts/%s/items/app_metafields/bulk", cartID), body, nil)
}

// BulkDeleteCartItemAppMetafields bulk deletes app metafields for cart items.
//
// Docs: DELETE /carts/{cart_id}/items/app_metafields/bulk
func (c *Client) BulkDeleteCartItemAppMetafields(ctx context.Context, cartID string, body any) error {
	if strings.TrimSpace(cartID) == "" {
		return fmt.Errorf("cart id is required")
	}
	return c.DeleteWithBody(ctx, fmt.Sprintf("/carts/%s/items/app_metafields/bulk", cartID), body, nil)
}
