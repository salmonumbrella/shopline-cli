package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// GetChannelPrices retrieves product channel prices for a channel.
//
// Docs: GET /channels/{id}/prices
func (c *Client) GetChannelPrices(ctx context.Context, channelID string) (json.RawMessage, error) {
	if strings.TrimSpace(channelID) == "" {
		return nil, fmt.Errorf("channel id is required")
	}

	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/channels/%s/prices", channelID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateChannelProductPrice creates a new product channel price for the given channel/product.
//
// Docs: POST /channels/{channel_id}/products/{id}/prices
func (c *Client) CreateChannelProductPrice(ctx context.Context, channelID, productID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(channelID) == "" {
		return nil, fmt.Errorf("channel id is required")
	}
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("product id is required")
	}

	var resp json.RawMessage
	if err := c.Post(ctx, fmt.Sprintf("/channels/%s/products/%s/prices", channelID, productID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateChannelProductPrice updates a product channel price.
//
// Docs: PUT /channels/{channel_id}/products/{product_id}/prices/{id}
func (c *Client) UpdateChannelProductPrice(ctx context.Context, channelID, productID, priceID string, body any) (json.RawMessage, error) {
	if strings.TrimSpace(channelID) == "" {
		return nil, fmt.Errorf("channel id is required")
	}
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	if strings.TrimSpace(priceID) == "" {
		return nil, fmt.Errorf("price id is required")
	}

	var resp json.RawMessage
	if err := c.Put(ctx, fmt.Sprintf("/channels/%s/products/%s/prices/%s", channelID, productID, priceID), body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
