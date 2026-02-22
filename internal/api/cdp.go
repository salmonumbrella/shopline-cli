package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CDPCustomerProfile represents a customer profile in the CDP.
type CDPCustomerProfile struct {
	ID                string                  `json:"id"`
	CustomerID        string                  `json:"customer_id"`
	Email             string                  `json:"email"`
	Phone             string                  `json:"phone"`
	FirstName         string                  `json:"first_name"`
	LastName          string                  `json:"last_name"`
	Segments          []string                `json:"segments"`
	Tags              []string                `json:"tags"`
	TotalOrders       int                     `json:"total_orders"`
	TotalSpent        string                  `json:"total_spent"`
	AverageOrderValue string                  `json:"average_order_value"`
	LastOrderAt       *time.Time              `json:"last_order_at"`
	FirstOrderAt      *time.Time              `json:"first_order_at"`
	LifetimeValue     string                  `json:"lifetime_value"`
	PredictedLTV      string                  `json:"predicted_ltv"`
	ChurnRisk         string                  `json:"churn_risk"`
	RFMScore          *CDPRFMScore            `json:"rfm_score"`
	Preferences       *CDPCustomerPreferences `json:"preferences"`
	Attributes        map[string]interface{}  `json:"attributes"`
	CreatedAt         time.Time               `json:"created_at"`
	UpdatedAt         time.Time               `json:"updated_at"`
}

// CDPRFMScore represents RFM (Recency, Frequency, Monetary) analysis.
type CDPRFMScore struct {
	Recency   int    `json:"recency"`
	Frequency int    `json:"frequency"`
	Monetary  int    `json:"monetary"`
	Total     int    `json:"total"`
	Segment   string `json:"segment"`
}

// CDPCustomerPreferences represents customer preferences.
type CDPCustomerPreferences struct {
	EmailMarketing      bool     `json:"email_marketing"`
	SMSMarketing        bool     `json:"sms_marketing"`
	PushNotifications   bool     `json:"push_notifications"`
	PreferredChannel    string   `json:"preferred_channel"`
	PreferredLanguage   string   `json:"preferred_language"`
	PreferredCategories []string `json:"preferred_categories"`
}

// CDPEvent represents a customer event tracked in the CDP.
type CDPEvent struct {
	ID         string                 `json:"id"`
	CustomerID string                 `json:"customer_id"`
	SessionID  string                 `json:"session_id"`
	EventType  string                 `json:"event_type"`
	EventName  string                 `json:"event_name"`
	Source     string                 `json:"source"`
	Channel    string                 `json:"channel"`
	Properties map[string]interface{} `json:"properties"`
	Timestamp  time.Time              `json:"timestamp"`
	CreatedAt  time.Time              `json:"created_at"`
}

// CDPSegment represents a customer segment in the CDP.
type CDPSegment struct {
	ID            string                `json:"id"`
	Name          string                `json:"name"`
	Description   string                `json:"description"`
	Type          string                `json:"type"`
	Conditions    []CDPSegmentCondition `json:"conditions"`
	CustomerCount int                   `json:"customer_count"`
	Status        string                `json:"status"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
}

// CDPSegmentCondition represents a segment condition.
type CDPSegmentCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// CDPProfilesListOptions contains options for listing CDP profiles.
type CDPProfilesListOptions struct {
	Page      int
	PageSize  int
	Segment   string
	Tag       string
	ChurnRisk string
	SortBy    string
	SortOrder string
}

// CDPEventsListOptions contains options for listing CDP events.
type CDPEventsListOptions struct {
	Page       int
	PageSize   int
	CustomerID string
	EventType  string
	EventName  string
	Source     string
	StartDate  *time.Time
	EndDate    *time.Time
	SortBy     string
	SortOrder  string
}

// CDPSegmentsListOptions contains options for listing CDP segments.
type CDPSegmentsListOptions struct {
	Page      int
	PageSize  int
	Type      string
	Status    string
	SortBy    string
	SortOrder string
}

// CDPProfilesListResponse is the paginated response for CDP profiles.
type CDPProfilesListResponse = ListResponse[CDPCustomerProfile]

// CDPEventsListResponse is the paginated response for CDP events.
type CDPEventsListResponse = ListResponse[CDPEvent]

// CDPSegmentsListResponse is the paginated response for CDP segments.
type CDPSegmentsListResponse = ListResponse[CDPSegment]

// ListCDPProfiles retrieves a list of CDP customer profiles.
func (c *Client) ListCDPProfiles(ctx context.Context, opts *CDPProfilesListOptions) (*CDPProfilesListResponse, error) {
	path := "/cdp/profiles"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("segment", opts.Segment).
			String("tag", opts.Tag).
			String("churn_risk", opts.ChurnRisk).
			String("sort_by", opts.SortBy).
			String("sort_order", opts.SortOrder).
			Build()
	}

	var resp CDPProfilesListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCDPProfile retrieves a single CDP profile by ID.
func (c *Client) GetCDPProfile(ctx context.Context, id string) (*CDPCustomerProfile, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("profile id is required")
	}
	var profile CDPCustomerProfile
	if err := c.Get(ctx, fmt.Sprintf("/cdp/profiles/%s", id), &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

// ListCDPEvents retrieves a list of CDP events.
func (c *Client) ListCDPEvents(ctx context.Context, opts *CDPEventsListOptions) (*CDPEventsListResponse, error) {
	path := "/cdp/events"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("customer_id", opts.CustomerID).
			String("event_type", opts.EventType).
			String("event_name", opts.EventName).
			String("source", opts.Source).
			Time("start_date", opts.StartDate).
			Time("end_date", opts.EndDate).
			String("sort_by", opts.SortBy).
			String("sort_order", opts.SortOrder).
			Build()
	}

	var resp CDPEventsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCDPEvent retrieves a single CDP event by ID.
func (c *Client) GetCDPEvent(ctx context.Context, id string) (*CDPEvent, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("event id is required")
	}
	var event CDPEvent
	if err := c.Get(ctx, fmt.Sprintf("/cdp/events/%s", id), &event); err != nil {
		return nil, err
	}
	return &event, nil
}

// ListCDPSegments retrieves a list of CDP segments.
func (c *Client) ListCDPSegments(ctx context.Context, opts *CDPSegmentsListOptions) (*CDPSegmentsListResponse, error) {
	path := "/cdp/segments"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("type", opts.Type).
			String("status", opts.Status).
			String("sort_by", opts.SortBy).
			String("sort_order", opts.SortOrder).
			Build()
	}

	var resp CDPSegmentsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCDPSegment retrieves a single CDP segment by ID.
func (c *Client) GetCDPSegment(ctx context.Context, id string) (*CDPSegment, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("segment id is required")
	}
	var segment CDPSegment
	if err := c.Get(ctx, fmt.Sprintf("/cdp/segments/%s", id), &segment); err != nil {
		return nil, err
	}
	return &segment, nil
}
