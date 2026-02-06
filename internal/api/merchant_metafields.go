package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Merchant (current) metafields (non-app-scoped)
//
// Docs use /merchants/current/... endpoints.

func (c *Client) ListMerchantMetafields(ctx context.Context) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/merchants/current/metafields", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetMerchantMetafield(ctx context.Context, metafieldID string) (json.RawMessage, error) {
	if strings.TrimSpace(metafieldID) == "" {
		return nil, fmt.Errorf("metafield id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/merchants/current/metafields/%s", metafieldID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) CreateMerchantMetafield(ctx context.Context, body any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Post(ctx, "/merchants/current/metafields", body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) UpdateMerchantMetafield(ctx context.Context, metafieldID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(metafieldID) == "" {
		return nil, fmt.Errorf("metafield id is required")
	}
	var resp json.RawMessage
	if err := c.Put(ctx, fmt.Sprintf("/merchants/current/metafields/%s", metafieldID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) DeleteMerchantMetafield(ctx context.Context, metafieldID string) error {
	if strings.TrimSpace(metafieldID) == "" {
		return fmt.Errorf("metafield id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/merchants/current/metafields/%s", metafieldID))
}

func (c *Client) BulkCreateMerchantMetafields(ctx context.Context, body any) error {
	return c.Post(ctx, "/merchants/current/metafields/bulk", body, nil)
}

func (c *Client) BulkUpdateMerchantMetafields(ctx context.Context, body any) error {
	return c.Put(ctx, "/merchants/current/metafields/bulk", body, nil)
}

func (c *Client) BulkDeleteMerchantMetafields(ctx context.Context, body any) error {
	return c.DeleteWithBody(ctx, "/merchants/current/metafields/bulk", body, nil)
}

// Merchant (current) app metafields (app-scoped)

func (c *Client) ListMerchantAppMetafields(ctx context.Context) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Get(ctx, "/merchants/current/app_metafields", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetMerchantAppMetafield(ctx context.Context, metafieldID string) (json.RawMessage, error) {
	if strings.TrimSpace(metafieldID) == "" {
		return nil, fmt.Errorf("metafield id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/merchants/current/app_metafields/%s", metafieldID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) CreateMerchantAppMetafield(ctx context.Context, body any) (json.RawMessage, error) {
	var resp json.RawMessage
	if err := c.Post(ctx, "/merchants/current/app_metafields", body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) UpdateMerchantAppMetafield(ctx context.Context, metafieldID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(metafieldID) == "" {
		return nil, fmt.Errorf("metafield id is required")
	}
	var resp json.RawMessage
	if err := c.Put(ctx, fmt.Sprintf("/merchants/current/app_metafields/%s", metafieldID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) DeleteMerchantAppMetafield(ctx context.Context, metafieldID string) error {
	if strings.TrimSpace(metafieldID) == "" {
		return fmt.Errorf("metafield id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/merchants/current/app_metafields/%s", metafieldID))
}

func (c *Client) BulkCreateMerchantAppMetafields(ctx context.Context, body any) error {
	return c.Post(ctx, "/merchants/current/app_metafields/bulk", body, nil)
}

func (c *Client) BulkUpdateMerchantAppMetafields(ctx context.Context, body any) error {
	return c.Put(ctx, "/merchants/current/app_metafields/bulk", body, nil)
}

func (c *Client) BulkDeleteMerchantAppMetafields(ctx context.Context, body any) error {
	return c.DeleteWithBody(ctx, "/merchants/current/app_metafields/bulk", body, nil)
}
