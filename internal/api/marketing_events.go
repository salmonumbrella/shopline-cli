package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// MarketingEvent represents a Shopline marketing event.
type MarketingEvent struct {
	ID                string             `json:"id"`
	EventType         string             `json:"event_type"`     // ad, campaign, email, social
	MarketingType     string             `json:"marketing_type"` // cpc, display, social, search, email
	RemoteID          string             `json:"remote_id"`
	StartedAt         time.Time          `json:"started_at"`
	EndedAt           time.Time          `json:"ended_at"`
	ScheduledToEnd    time.Time          `json:"scheduled_to_end"`
	Budget            float64            `json:"budget"`
	Currency          string             `json:"currency"`
	ManageURL         string             `json:"manage_url"`
	PreviewURL        string             `json:"preview_url"`
	UTMCampaign       string             `json:"utm_campaign"`
	UTMSource         string             `json:"utm_source"`
	UTMMedium         string             `json:"utm_medium"`
	Description       string             `json:"description"`
	MarketedResources []MarketedResource `json:"marketed_resources"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

// MarketedResource represents a resource being marketed.
type MarketedResource struct {
	Type string `json:"type"` // product, collection, page
	ID   string `json:"id"`
}

// MarketingEventsListOptions contains options for listing marketing events.
type MarketingEventsListOptions struct {
	Page          int
	PageSize      int
	EventType     string
	MarketingType string
}

// MarketingEventsListResponse is the paginated response for marketing events.
type MarketingEventsListResponse = ListResponse[MarketingEvent]

// MarketingEventCreateRequest contains the data for creating a marketing event.
type MarketingEventCreateRequest struct {
	EventType         string             `json:"event_type"`
	MarketingType     string             `json:"marketing_type"`
	RemoteID          string             `json:"remote_id,omitempty"`
	StartedAt         *time.Time         `json:"started_at,omitempty"`
	ScheduledToEnd    *time.Time         `json:"scheduled_to_end,omitempty"`
	Budget            float64            `json:"budget,omitempty"`
	Currency          string             `json:"currency,omitempty"`
	ManageURL         string             `json:"manage_url,omitempty"`
	PreviewURL        string             `json:"preview_url,omitempty"`
	UTMCampaign       string             `json:"utm_campaign,omitempty"`
	UTMSource         string             `json:"utm_source,omitempty"`
	UTMMedium         string             `json:"utm_medium,omitempty"`
	Description       string             `json:"description,omitempty"`
	MarketedResources []MarketedResource `json:"marketed_resources,omitempty"`
}

// MarketingEventUpdateRequest contains the data for updating a marketing event.
type MarketingEventUpdateRequest struct {
	RemoteID       string     `json:"remote_id,omitempty"`
	EndedAt        *time.Time `json:"ended_at,omitempty"`
	ScheduledToEnd *time.Time `json:"scheduled_to_end,omitempty"`
	Budget         float64    `json:"budget,omitempty"`
	ManageURL      string     `json:"manage_url,omitempty"`
	PreviewURL     string     `json:"preview_url,omitempty"`
	Description    string     `json:"description,omitempty"`
}

// ListMarketingEvents retrieves a list of marketing events.
func (c *Client) ListMarketingEvents(ctx context.Context, opts *MarketingEventsListOptions) (*MarketingEventsListResponse, error) {
	path := "/marketing_events"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("event_type", opts.EventType).
			String("marketing_type", opts.MarketingType).
			Build()
	}

	var resp MarketingEventsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetMarketingEvent retrieves a single marketing event by ID.
func (c *Client) GetMarketingEvent(ctx context.Context, id string) (*MarketingEvent, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("marketing event id is required")
	}
	var event MarketingEvent
	if err := c.Get(ctx, fmt.Sprintf("/marketing_events/%s", id), &event); err != nil {
		return nil, err
	}
	return &event, nil
}

// CreateMarketingEvent creates a new marketing event.
func (c *Client) CreateMarketingEvent(ctx context.Context, req *MarketingEventCreateRequest) (*MarketingEvent, error) {
	var event MarketingEvent
	if err := c.Post(ctx, "/marketing_events", req, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

// UpdateMarketingEvent updates an existing marketing event.
func (c *Client) UpdateMarketingEvent(ctx context.Context, id string, req *MarketingEventUpdateRequest) (*MarketingEvent, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("marketing event id is required")
	}
	var event MarketingEvent
	if err := c.Put(ctx, fmt.Sprintf("/marketing_events/%s", id), req, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

// DeleteMarketingEvent deletes a marketing event.
func (c *Client) DeleteMarketingEvent(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("marketing event id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/marketing_events/%s", id))
}
