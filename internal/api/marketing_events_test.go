package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMarketingEventsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/marketing_events" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := MarketingEventsListResponse{
			Items: []MarketingEvent{
				{ID: "mkt_123", EventType: "ad", MarketingType: "cpc", UTMCampaign: "summer-sale"},
				{ID: "mkt_456", EventType: "email", MarketingType: "email", UTMCampaign: "newsletter"},
			},
			Page:       1,
			PageSize:   20,
			TotalCount: 2,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	events, err := client.ListMarketingEvents(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListMarketingEvents failed: %v", err)
	}

	if len(events.Items) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events.Items))
	}
	if events.Items[0].ID != "mkt_123" {
		t.Errorf("Unexpected event ID: %s", events.Items[0].ID)
	}
}

func TestMarketingEventsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("event_type") != "ad" {
			t.Errorf("Expected event_type=ad, got %s", r.URL.Query().Get("event_type"))
		}
		if r.URL.Query().Get("marketing_type") != "cpc" {
			t.Errorf("Expected marketing_type=cpc, got %s", r.URL.Query().Get("marketing_type"))
		}

		resp := MarketingEventsListResponse{
			Items:      []MarketingEvent{},
			Page:       1,
			PageSize:   20,
			TotalCount: 0,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &MarketingEventsListOptions{
		EventType:     "ad",
		MarketingType: "cpc",
	}
	_, err := client.ListMarketingEvents(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListMarketingEvents failed: %v", err)
	}
}

func TestMarketingEventsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/marketing_events/mkt_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		event := MarketingEvent{ID: "mkt_123", EventType: "ad", MarketingType: "cpc", UTMCampaign: "summer-sale"}
		_ = json.NewEncoder(w).Encode(event)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	event, err := client.GetMarketingEvent(context.Background(), "mkt_123")
	if err != nil {
		t.Fatalf("GetMarketingEvent failed: %v", err)
	}

	if event.ID != "mkt_123" {
		t.Errorf("Unexpected event ID: %s", event.ID)
	}
}

func TestGetMarketingEventEmptyID(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name string
		id   string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab only", "\t"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.GetMarketingEvent(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "marketing event id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestMarketingEventsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/marketing_events" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req MarketingEventCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.EventType != "ad" {
			t.Errorf("Unexpected event type: %s", req.EventType)
		}

		event := MarketingEvent{
			ID:            "mkt_new",
			EventType:     req.EventType,
			MarketingType: req.MarketingType,
			UTMCampaign:   req.UTMCampaign,
		}
		_ = json.NewEncoder(w).Encode(event)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &MarketingEventCreateRequest{
		EventType:     "ad",
		MarketingType: "cpc",
		UTMCampaign:   "new-campaign",
	}
	event, err := client.CreateMarketingEvent(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateMarketingEvent failed: %v", err)
	}

	if event.ID != "mkt_new" {
		t.Errorf("Unexpected event ID: %s", event.ID)
	}
	if event.EventType != "ad" {
		t.Errorf("Unexpected event type: %s", event.EventType)
	}
}

func TestMarketingEventsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/marketing_events/mkt_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req MarketingEventUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		event := MarketingEvent{
			ID:          "mkt_123",
			EventType:   "ad",
			Budget:      req.Budget,
			Description: req.Description,
		}
		_ = json.NewEncoder(w).Encode(event)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &MarketingEventUpdateRequest{
		Budget:      1000.0,
		Description: "Updated description",
	}
	event, err := client.UpdateMarketingEvent(context.Background(), "mkt_123", req)
	if err != nil {
		t.Fatalf("UpdateMarketingEvent failed: %v", err)
	}

	if event.Budget != 1000.0 {
		t.Errorf("Unexpected budget: %f", event.Budget)
	}
}

func TestMarketingEventsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/marketing_events/mkt_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteMarketingEvent(context.Background(), "mkt_123")
	if err != nil {
		t.Fatalf("DeleteMarketingEvent failed: %v", err)
	}
}

func TestUpdateMarketingEventEmptyID(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name string
		id   string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.UpdateMarketingEvent(context.Background(), tc.id, &MarketingEventUpdateRequest{Description: "Test"})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "marketing event id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteMarketingEventEmptyID(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name string
		id   string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.DeleteMarketingEvent(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "marketing event id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
