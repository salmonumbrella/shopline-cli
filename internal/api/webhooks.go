package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// WebhookFormat represents the format of webhook payloads.
type WebhookFormat string

const (
	WebhookFormatJSON WebhookFormat = "json"
	WebhookFormatXML  WebhookFormat = "xml"
)

// Webhook represents a Shopline webhook subscription.
type Webhook struct {
	ID         string        `json:"id"`
	Address    string        `json:"address"`
	Topic      string        `json:"topic"`
	Format     WebhookFormat `json:"format"`
	APIVersion string        `json:"api_version"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

// WebhooksListOptions contains options for listing webhooks.
type WebhooksListOptions struct {
	Page     int
	PageSize int
	Topic    string
}

// WebhooksListResponse is the paginated response for webhooks.
type WebhooksListResponse = ListResponse[Webhook]

// WebhookCreateRequest contains the data for creating a webhook.
type WebhookCreateRequest struct {
	Address    string        `json:"address"`
	Topic      string        `json:"topic"`
	Format     WebhookFormat `json:"format,omitempty"`
	APIVersion string        `json:"api_version,omitempty"`
}

// WebhookUpdateRequest contains the data for updating a webhook subscription.
// All fields are optional; only provided fields will be sent.
type WebhookUpdateRequest struct {
	Address    string        `json:"address,omitempty"`
	Topic      string        `json:"topic,omitempty"`
	Format     WebhookFormat `json:"format,omitempty"`
	APIVersion string        `json:"api_version,omitempty"`
}

// ListWebhooks retrieves a list of webhooks.
func (c *Client) ListWebhooks(ctx context.Context, opts *WebhooksListOptions) (*WebhooksListResponse, error) {
	path := "/webhooks"
	if opts != nil {
		params := url.Values{}
		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			params.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.Topic != "" {
			params.Set("topic", opts.Topic)
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp WebhooksListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetWebhook retrieves a single webhook by ID.
func (c *Client) GetWebhook(ctx context.Context, id string) (*Webhook, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("webhook id is required")
	}
	var webhook Webhook
	if err := c.Get(ctx, fmt.Sprintf("/webhooks/%s", id), &webhook); err != nil {
		return nil, err
	}
	return &webhook, nil
}

// CreateWebhook creates a new webhook subscription.
func (c *Client) CreateWebhook(ctx context.Context, req *WebhookCreateRequest) (*Webhook, error) {
	var webhook Webhook
	if err := c.Post(ctx, "/webhooks", req, &webhook); err != nil {
		return nil, err
	}
	return &webhook, nil
}

// UpdateWebhook updates an existing webhook subscription.
//
// Docs: PUT /webhooks/{id}
func (c *Client) UpdateWebhook(ctx context.Context, id string, req *WebhookUpdateRequest) (*Webhook, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("webhook id is required")
	}
	var webhook Webhook
	if err := c.Put(ctx, fmt.Sprintf("/webhooks/%s", id), req, &webhook); err != nil {
		return nil, err
	}
	return &webhook, nil
}

// DeleteWebhook deletes a webhook subscription.
func (c *Client) DeleteWebhook(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("webhook id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/webhooks/%s", id))
}
