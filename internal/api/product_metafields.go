package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Product metafields (non-app-scoped)

func (c *Client) ListProductMetafields(ctx context.Context, productID string) (json.RawMessage, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/products/%s/metafields", productID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetProductMetafield(ctx context.Context, productID, metafieldID string) (json.RawMessage, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	if strings.TrimSpace(metafieldID) == "" {
		return nil, fmt.Errorf("metafield id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/products/%s/metafields/%s", productID, metafieldID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) CreateProductMetafield(ctx context.Context, productID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	var resp json.RawMessage
	if err := c.Post(ctx, fmt.Sprintf("/products/%s/metafields", productID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) UpdateProductMetafield(ctx context.Context, productID, metafieldID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	if strings.TrimSpace(metafieldID) == "" {
		return nil, fmt.Errorf("metafield id is required")
	}
	var resp json.RawMessage
	if err := c.Put(ctx, fmt.Sprintf("/products/%s/metafields/%s", productID, metafieldID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) DeleteProductMetafield(ctx context.Context, productID, metafieldID string) error {
	if strings.TrimSpace(productID) == "" {
		return fmt.Errorf("product id is required")
	}
	if strings.TrimSpace(metafieldID) == "" {
		return fmt.Errorf("metafield id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/products/%s/metafields/%s", productID, metafieldID))
}

func (c *Client) BulkCreateProductMetafields(ctx context.Context, productID string, body any) error {
	if strings.TrimSpace(productID) == "" {
		return fmt.Errorf("product id is required")
	}
	return c.Post(ctx, fmt.Sprintf("/products/%s/metafields/bulk", productID), body, nil)
}

func (c *Client) BulkUpdateProductMetafields(ctx context.Context, productID string, body any) error {
	if strings.TrimSpace(productID) == "" {
		return fmt.Errorf("product id is required")
	}
	return c.Put(ctx, fmt.Sprintf("/products/%s/metafields/bulk", productID), body, nil)
}

func (c *Client) BulkDeleteProductMetafields(ctx context.Context, productID string, body any) error {
	if strings.TrimSpace(productID) == "" {
		return fmt.Errorf("product id is required")
	}
	return c.DeleteWithBody(ctx, fmt.Sprintf("/products/%s/metafields/bulk", productID), body, nil)
}

// Product app metafields (app-scoped)

func (c *Client) ListProductAppMetafields(ctx context.Context, productID string) (json.RawMessage, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/products/%s/app_metafields", productID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetProductAppMetafield(ctx context.Context, productID, metafieldID string) (json.RawMessage, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	if strings.TrimSpace(metafieldID) == "" {
		return nil, fmt.Errorf("metafield id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/products/%s/app_metafields/%s", productID, metafieldID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) CreateProductAppMetafield(ctx context.Context, productID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	var resp json.RawMessage
	if err := c.Post(ctx, fmt.Sprintf("/products/%s/app_metafields", productID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) UpdateProductAppMetafield(ctx context.Context, productID, metafieldID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	if strings.TrimSpace(metafieldID) == "" {
		return nil, fmt.Errorf("metafield id is required")
	}
	var resp json.RawMessage
	if err := c.Put(ctx, fmt.Sprintf("/products/%s/app_metafields/%s", productID, metafieldID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) DeleteProductAppMetafield(ctx context.Context, productID, metafieldID string) error {
	if strings.TrimSpace(productID) == "" {
		return fmt.Errorf("product id is required")
	}
	if strings.TrimSpace(metafieldID) == "" {
		return fmt.Errorf("metafield id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/products/%s/app_metafields/%s", productID, metafieldID))
}

func (c *Client) BulkCreateProductAppMetafields(ctx context.Context, productID string, body any) error {
	if strings.TrimSpace(productID) == "" {
		return fmt.Errorf("product id is required")
	}
	return c.Post(ctx, fmt.Sprintf("/products/%s/app_metafields/bulk", productID), body, nil)
}

func (c *Client) BulkUpdateProductAppMetafields(ctx context.Context, productID string, body any) error {
	if strings.TrimSpace(productID) == "" {
		return fmt.Errorf("product id is required")
	}
	return c.Put(ctx, fmt.Sprintf("/products/%s/app_metafields/bulk", productID), body, nil)
}

func (c *Client) BulkDeleteProductAppMetafields(ctx context.Context, productID string, body any) error {
	if strings.TrimSpace(productID) == "" {
		return fmt.Errorf("product id is required")
	}
	return c.DeleteWithBody(ctx, fmt.Sprintf("/products/%s/app_metafields/bulk", productID), body, nil)
}
