package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Channel represents a Shopline sales channel.
type Channel struct {
	ID                        string    `json:"id"`
	Name                      string    `json:"name"`
	Handle                    string    `json:"handle"`
	Type                      string    `json:"type"`
	Active                    bool      `json:"active"`
	SupportsRemoteFulfillment bool      `json:"supports_remote_fulfillment"`
	ProductCount              int       `json:"product_count"`
	CollectionCount           int       `json:"collection_count"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

// ChannelProduct represents a product in a channel.
type ChannelProduct struct {
	ProductID string `json:"product_id"`
	Published bool   `json:"published"`
}

// ChannelsListOptions contains options for listing channels.
type ChannelsListOptions struct {
	Page     int
	PageSize int
	Active   *bool
	Platform string // Required: "ec" for ecommerce, "pos" for point of sale, etc.
}

// ChannelsListResponse contains the list response.
type ChannelsListResponse struct {
	Items      []Channel `json:"items"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalCount int       `json:"total_count"`
	HasMore    bool      `json:"has_more"`
}

// ChannelProductsResponse contains the list response for simple channel products.
type ChannelProductsResponse struct {
	Items      []ChannelProduct `json:"items"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalCount int              `json:"total_count"`
	HasMore    bool             `json:"has_more"`
}

// ChannelCreateRequest contains the request body for creating a channel.
type ChannelCreateRequest struct {
	Name   string `json:"name"`
	Handle string `json:"handle,omitempty"`
	Type   string `json:"type"`
}

// ChannelUpdateRequest contains the request body for updating a channel.
type ChannelUpdateRequest struct {
	Name   *string `json:"name,omitempty"`
	Active *bool   `json:"active,omitempty"`
}

// ChannelPublishProductRequest contains the request body for publishing a product to a channel.
type ChannelPublishProductRequest struct {
	ProductID string `json:"product_id"`
}

// ListChannels retrieves a list of channels.
// The Platform parameter is required by the Shopline API (e.g., "ec" for ecommerce).
func (c *Client) ListChannels(ctx context.Context, opts *ChannelsListOptions) (*ChannelsListResponse, error) {
	path := "/channels"
	params := url.Values{}
	// Platform is required by the API - default to "ec" if not specified
	platform := "ec"
	if opts != nil {
		if opts.Platform != "" {
			platform = opts.Platform
		}
		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			params.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.Active != nil {
			params.Set("active", strconv.FormatBool(*opts.Active))
		}
	}
	params.Set("platform", platform)
	path += "?" + params.Encode()

	var resp ChannelsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetChannel retrieves a single channel by ID.
func (c *Client) GetChannel(ctx context.Context, id string) (*Channel, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("channel id is required")
	}
	var channel Channel
	if err := c.Get(ctx, fmt.Sprintf("/channels/%s", id), &channel); err != nil {
		return nil, err
	}
	return &channel, nil
}

// CreateChannel creates a new channel.
func (c *Client) CreateChannel(ctx context.Context, req *ChannelCreateRequest) (*Channel, error) {
	var channel Channel
	if err := c.Post(ctx, "/channels", req, &channel); err != nil {
		return nil, err
	}
	return &channel, nil
}

// UpdateChannel updates a channel.
func (c *Client) UpdateChannel(ctx context.Context, id string, req *ChannelUpdateRequest) (*Channel, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("channel id is required")
	}
	var channel Channel
	if err := c.Put(ctx, fmt.Sprintf("/channels/%s", id), req, &channel); err != nil {
		return nil, err
	}
	return &channel, nil
}

// DeleteChannel deletes a channel.
func (c *Client) DeleteChannel(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("channel id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/channels/%s", id))
}

// ListChannelProducts retrieves products in a channel.
func (c *Client) ListChannelProducts(ctx context.Context, channelID string, page, pageSize int) (*ChannelProductsResponse, error) {
	if strings.TrimSpace(channelID) == "" {
		return nil, fmt.Errorf("channel id is required")
	}
	path := fmt.Sprintf("/channels/%s/products", channelID)
	params := url.Values{}
	if page > 0 {
		params.Set("page", strconv.Itoa(page))
	}
	if pageSize > 0 {
		params.Set("page_size", strconv.Itoa(pageSize))
	}
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var resp ChannelProductsResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// PublishProductToChannel publishes a product to a channel.
func (c *Client) PublishProductToChannel(ctx context.Context, channelID string, req *ChannelPublishProductRequest) error {
	if strings.TrimSpace(channelID) == "" {
		return fmt.Errorf("channel id is required")
	}
	return c.Post(ctx, fmt.Sprintf("/channels/%s/products", channelID), req, nil)
}

// UnpublishProductFromChannel removes a product from a channel.
func (c *Client) UnpublishProductFromChannel(ctx context.Context, channelID, productID string) error {
	if strings.TrimSpace(channelID) == "" {
		return fmt.Errorf("channel id is required")
	}
	if strings.TrimSpace(productID) == "" {
		return fmt.Errorf("product id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/channels/%s/products/%s", channelID, productID))
}
