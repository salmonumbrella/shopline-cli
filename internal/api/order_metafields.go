package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// NOTE: The Shopline reference pages for metafields often don't provide stable request/response schemas.
// To keep the CLI resilient and still provide 100% endpoint coverage, we model these as json.RawMessage
// and accept "any" request bodies. Callers can supply strongly-typed structs if they want.

// Order metafields (non-app-scoped)

func (c *Client) ListOrderMetafields(ctx context.Context, orderID string) (json.RawMessage, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/orders/%s/metafields", orderID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetOrderMetafield(ctx context.Context, orderID, metafieldID string) (json.RawMessage, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	if strings.TrimSpace(metafieldID) == "" {
		return nil, fmt.Errorf("metafield id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/orders/%s/metafields/%s", orderID, metafieldID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) CreateOrderMetafield(ctx context.Context, orderID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	var resp json.RawMessage
	if err := c.Post(ctx, fmt.Sprintf("/orders/%s/metafields", orderID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) UpdateOrderMetafield(ctx context.Context, orderID, metafieldID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	if strings.TrimSpace(metafieldID) == "" {
		return nil, fmt.Errorf("metafield id is required")
	}
	var resp json.RawMessage
	if err := c.Put(ctx, fmt.Sprintf("/orders/%s/metafields/%s", orderID, metafieldID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) DeleteOrderMetafield(ctx context.Context, orderID, metafieldID string) error {
	if strings.TrimSpace(orderID) == "" {
		return fmt.Errorf("order id is required")
	}
	if strings.TrimSpace(metafieldID) == "" {
		return fmt.Errorf("metafield id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/orders/%s/metafields/%s", orderID, metafieldID))
}

func (c *Client) BulkCreateOrderMetafields(ctx context.Context, orderID string, body any) error {
	if strings.TrimSpace(orderID) == "" {
		return fmt.Errorf("order id is required")
	}
	return c.Post(ctx, fmt.Sprintf("/orders/%s/metafields/bulk", orderID), body, nil)
}

func (c *Client) BulkUpdateOrderMetafields(ctx context.Context, orderID string, body any) error {
	if strings.TrimSpace(orderID) == "" {
		return fmt.Errorf("order id is required")
	}
	return c.Put(ctx, fmt.Sprintf("/orders/%s/metafields/bulk", orderID), body, nil)
}

func (c *Client) BulkDeleteOrderMetafields(ctx context.Context, orderID string, body any) error {
	if strings.TrimSpace(orderID) == "" {
		return fmt.Errorf("order id is required")
	}
	return c.DeleteWithBody(ctx, fmt.Sprintf("/orders/%s/metafields/bulk", orderID), body, nil)
}

// Order app metafields (app-scoped)

func (c *Client) ListOrderAppMetafields(ctx context.Context, orderID string) (json.RawMessage, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/orders/%s/app_metafields", orderID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetOrderAppMetafield(ctx context.Context, orderID, metafieldID string) (json.RawMessage, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	if strings.TrimSpace(metafieldID) == "" {
		return nil, fmt.Errorf("metafield id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/orders/%s/app_metafields/%s", orderID, metafieldID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) CreateOrderAppMetafield(ctx context.Context, orderID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	var resp json.RawMessage
	if err := c.Post(ctx, fmt.Sprintf("/orders/%s/app_metafields", orderID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) UpdateOrderAppMetafield(ctx context.Context, orderID, metafieldID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	if strings.TrimSpace(metafieldID) == "" {
		return nil, fmt.Errorf("metafield id is required")
	}
	var resp json.RawMessage
	if err := c.Put(ctx, fmt.Sprintf("/orders/%s/app_metafields/%s", orderID, metafieldID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) DeleteOrderAppMetafield(ctx context.Context, orderID, metafieldID string) error {
	if strings.TrimSpace(orderID) == "" {
		return fmt.Errorf("order id is required")
	}
	if strings.TrimSpace(metafieldID) == "" {
		return fmt.Errorf("metafield id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/orders/%s/app_metafields/%s", orderID, metafieldID))
}

func (c *Client) BulkCreateOrderAppMetafields(ctx context.Context, orderID string, body any) error {
	if strings.TrimSpace(orderID) == "" {
		return fmt.Errorf("order id is required")
	}
	return c.Post(ctx, fmt.Sprintf("/orders/%s/app_metafields/bulk", orderID), body, nil)
}

func (c *Client) BulkUpdateOrderAppMetafields(ctx context.Context, orderID string, body any) error {
	if strings.TrimSpace(orderID) == "" {
		return fmt.Errorf("order id is required")
	}
	return c.Put(ctx, fmt.Sprintf("/orders/%s/app_metafields/bulk", orderID), body, nil)
}

func (c *Client) BulkDeleteOrderAppMetafields(ctx context.Context, orderID string, body any) error {
	if strings.TrimSpace(orderID) == "" {
		return fmt.Errorf("order id is required")
	}
	return c.DeleteWithBody(ctx, fmt.Sprintf("/orders/%s/app_metafields/bulk", orderID), body, nil)
}

// Order item metafields (non-app-scoped)

func (c *Client) ListOrderItemMetafields(ctx context.Context, orderID string) (json.RawMessage, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/orders/%s/items/metafields", orderID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) BulkCreateOrderItemMetafields(ctx context.Context, orderID string, body any) error {
	if strings.TrimSpace(orderID) == "" {
		return fmt.Errorf("order id is required")
	}
	return c.Post(ctx, fmt.Sprintf("/orders/%s/items/metafields/bulk", orderID), body, nil)
}

func (c *Client) BulkUpdateOrderItemMetafields(ctx context.Context, orderID string, body any) error {
	if strings.TrimSpace(orderID) == "" {
		return fmt.Errorf("order id is required")
	}
	return c.Put(ctx, fmt.Sprintf("/orders/%s/items/metafields/bulk", orderID), body, nil)
}

func (c *Client) BulkDeleteOrderItemMetafields(ctx context.Context, orderID string, body any) error {
	if strings.TrimSpace(orderID) == "" {
		return fmt.Errorf("order id is required")
	}
	return c.DeleteWithBody(ctx, fmt.Sprintf("/orders/%s/items/metafields/bulk", orderID), body, nil)
}

// Order item app metafields (app-scoped)

func (c *Client) ListOrderItemAppMetafields(ctx context.Context, orderID string) (json.RawMessage, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/orders/%s/items/app_metafields", orderID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) BulkCreateOrderItemAppMetafields(ctx context.Context, orderID string, body any) error {
	if strings.TrimSpace(orderID) == "" {
		return fmt.Errorf("order id is required")
	}
	return c.Post(ctx, fmt.Sprintf("/orders/%s/items/app_metafields/bulk", orderID), body, nil)
}

func (c *Client) BulkUpdateOrderItemAppMetafields(ctx context.Context, orderID string, body any) error {
	if strings.TrimSpace(orderID) == "" {
		return fmt.Errorf("order id is required")
	}
	return c.Put(ctx, fmt.Sprintf("/orders/%s/items/app_metafields/bulk", orderID), body, nil)
}

func (c *Client) BulkDeleteOrderItemAppMetafields(ctx context.Context, orderID string, body any) error {
	if strings.TrimSpace(orderID) == "" {
		return fmt.Errorf("order id is required")
	}
	return c.DeleteWithBody(ctx, fmt.Sprintf("/orders/%s/items/app_metafields/bulk", orderID), body, nil)
}
