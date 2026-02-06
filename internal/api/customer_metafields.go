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

// Customer metafields (non-app-scoped)

func (c *Client) ListCustomerMetafields(ctx context.Context, customerID string) (json.RawMessage, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/customers/%s/metafields", customerID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetCustomerMetafield(ctx context.Context, customerID, metafieldID string) (json.RawMessage, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	if strings.TrimSpace(metafieldID) == "" {
		return nil, fmt.Errorf("metafield id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/customers/%s/metafields/%s", customerID, metafieldID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) CreateCustomerMetafield(ctx context.Context, customerID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	var resp json.RawMessage
	if err := c.Post(ctx, fmt.Sprintf("/customers/%s/metafields", customerID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) UpdateCustomerMetafield(ctx context.Context, customerID, metafieldID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	if strings.TrimSpace(metafieldID) == "" {
		return nil, fmt.Errorf("metafield id is required")
	}
	var resp json.RawMessage
	if err := c.Put(ctx, fmt.Sprintf("/customers/%s/metafields/%s", customerID, metafieldID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) DeleteCustomerMetafield(ctx context.Context, customerID, metafieldID string) error {
	if strings.TrimSpace(customerID) == "" {
		return fmt.Errorf("customer id is required")
	}
	if strings.TrimSpace(metafieldID) == "" {
		return fmt.Errorf("metafield id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/customers/%s/metafields/%s", customerID, metafieldID))
}

func (c *Client) BulkCreateCustomerMetafields(ctx context.Context, customerID string, body any) error {
	if strings.TrimSpace(customerID) == "" {
		return fmt.Errorf("customer id is required")
	}
	return c.Post(ctx, fmt.Sprintf("/customers/%s/metafields/bulk", customerID), body, nil)
}

func (c *Client) BulkUpdateCustomerMetafields(ctx context.Context, customerID string, body any) error {
	if strings.TrimSpace(customerID) == "" {
		return fmt.Errorf("customer id is required")
	}
	return c.Put(ctx, fmt.Sprintf("/customers/%s/metafields/bulk", customerID), body, nil)
}

func (c *Client) BulkDeleteCustomerMetafields(ctx context.Context, customerID string, body any) error {
	if strings.TrimSpace(customerID) == "" {
		return fmt.Errorf("customer id is required")
	}
	return c.DeleteWithBody(ctx, fmt.Sprintf("/customers/%s/metafields/bulk", customerID), body, nil)
}

// Customer app metafields (app-scoped)

func (c *Client) ListCustomerAppMetafields(ctx context.Context, customerID string) (json.RawMessage, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/customers/%s/app_metafields", customerID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetCustomerAppMetafield(ctx context.Context, customerID, metafieldID string) (json.RawMessage, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	if strings.TrimSpace(metafieldID) == "" {
		return nil, fmt.Errorf("metafield id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/customers/%s/app_metafields/%s", customerID, metafieldID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) CreateCustomerAppMetafield(ctx context.Context, customerID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	var resp json.RawMessage
	if err := c.Post(ctx, fmt.Sprintf("/customers/%s/app_metafields", customerID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) UpdateCustomerAppMetafield(ctx context.Context, customerID, metafieldID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	if strings.TrimSpace(metafieldID) == "" {
		return nil, fmt.Errorf("metafield id is required")
	}
	var resp json.RawMessage
	if err := c.Put(ctx, fmt.Sprintf("/customers/%s/app_metafields/%s", customerID, metafieldID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) DeleteCustomerAppMetafield(ctx context.Context, customerID, metafieldID string) error {
	if strings.TrimSpace(customerID) == "" {
		return fmt.Errorf("customer id is required")
	}
	if strings.TrimSpace(metafieldID) == "" {
		return fmt.Errorf("metafield id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/customers/%s/app_metafields/%s", customerID, metafieldID))
}

func (c *Client) BulkCreateCustomerAppMetafields(ctx context.Context, customerID string, body any) error {
	if strings.TrimSpace(customerID) == "" {
		return fmt.Errorf("customer id is required")
	}
	return c.Post(ctx, fmt.Sprintf("/customers/%s/app_metafields/bulk", customerID), body, nil)
}

func (c *Client) BulkUpdateCustomerAppMetafields(ctx context.Context, customerID string, body any) error {
	if strings.TrimSpace(customerID) == "" {
		return fmt.Errorf("customer id is required")
	}
	return c.Put(ctx, fmt.Sprintf("/customers/%s/app_metafields/bulk", customerID), body, nil)
}

func (c *Client) BulkDeleteCustomerAppMetafields(ctx context.Context, customerID string, body any) error {
	if strings.TrimSpace(customerID) == "" {
		return fmt.Errorf("customer id is required")
	}
	return c.DeleteWithBody(ctx, fmt.Sprintf("/customers/%s/app_metafields/bulk", customerID), body, nil)
}
