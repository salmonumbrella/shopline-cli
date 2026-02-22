package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCDPProfilesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/cdp/profiles" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := CDPProfilesListResponse{
			Items: []CDPCustomerProfile{
				{ID: "prof_123", CustomerID: "cust_1", Email: "john@example.com", TotalOrders: 5},
				{ID: "prof_456", CustomerID: "cust_2", Email: "jane@example.com", TotalOrders: 12},
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

	profiles, err := client.ListCDPProfiles(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListCDPProfiles failed: %v", err)
	}

	if len(profiles.Items) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(profiles.Items))
	}
	if profiles.Items[0].ID != "prof_123" {
		t.Errorf("Unexpected profile ID: %s", profiles.Items[0].ID)
	}
}

func TestCDPProfilesListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("segment") != "vip" {
			t.Errorf("Expected segment=vip, got %s", r.URL.Query().Get("segment"))
		}
		if r.URL.Query().Get("churn_risk") != "high" {
			t.Errorf("Expected churn_risk=high, got %s", r.URL.Query().Get("churn_risk"))
		}

		resp := CDPProfilesListResponse{
			Items:      []CDPCustomerProfile{{ID: "prof_123", Segments: []string{"vip"}, ChurnRisk: "high"}},
			Page:       1,
			PageSize:   20,
			TotalCount: 1,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &CDPProfilesListOptions{
		Segment:   "vip",
		ChurnRisk: "high",
	}
	profiles, err := client.ListCDPProfiles(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListCDPProfiles failed: %v", err)
	}

	if len(profiles.Items) != 1 {
		t.Errorf("Expected 1 profile, got %d", len(profiles.Items))
	}
}

func TestCDPProfilesGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cdp/profiles/prof_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		profile := CDPCustomerProfile{
			ID:            "prof_123",
			CustomerID:    "cust_1",
			Email:         "john@example.com",
			TotalOrders:   5,
			TotalSpent:    "500.00",
			LifetimeValue: "750.00",
		}
		_ = json.NewEncoder(w).Encode(profile)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	profile, err := client.GetCDPProfile(context.Background(), "prof_123")
	if err != nil {
		t.Fatalf("GetCDPProfile failed: %v", err)
	}

	if profile.ID != "prof_123" {
		t.Errorf("Unexpected profile ID: %s", profile.ID)
	}
	if profile.TotalSpent != "500.00" {
		t.Errorf("Unexpected total spent: %s", profile.TotalSpent)
	}
}

func TestCDPEventsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/cdp/events" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := CDPEventsListResponse{
			Items: []CDPEvent{
				{ID: "evt_123", CustomerID: "cust_1", EventType: "page_view", EventName: "product_viewed"},
				{ID: "evt_456", CustomerID: "cust_1", EventType: "purchase", EventName: "order_completed"},
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

	events, err := client.ListCDPEvents(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListCDPEvents failed: %v", err)
	}

	if len(events.Items) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events.Items))
	}
}

func TestCDPEventsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cdp/events/evt_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		event := CDPEvent{
			ID:         "evt_123",
			CustomerID: "cust_1",
			EventType:  "page_view",
			EventName:  "product_viewed",
		}
		_ = json.NewEncoder(w).Encode(event)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	event, err := client.GetCDPEvent(context.Background(), "evt_123")
	if err != nil {
		t.Fatalf("GetCDPEvent failed: %v", err)
	}

	if event.ID != "evt_123" {
		t.Errorf("Unexpected event ID: %s", event.ID)
	}
}

func TestCDPSegmentsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/cdp/segments" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := CDPSegmentsListResponse{
			Items: []CDPSegment{
				{ID: "seg_123", Name: "VIP Customers", CustomerCount: 150},
				{ID: "seg_456", Name: "At Risk", CustomerCount: 45},
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

	segments, err := client.ListCDPSegments(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListCDPSegments failed: %v", err)
	}

	if len(segments.Items) != 2 {
		t.Errorf("Expected 2 segments, got %d", len(segments.Items))
	}
}

func TestCDPSegmentsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cdp/segments/seg_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		segment := CDPSegment{
			ID:            "seg_123",
			Name:          "VIP Customers",
			CustomerCount: 150,
			Status:        "active",
		}
		_ = json.NewEncoder(w).Encode(segment)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	segment, err := client.GetCDPSegment(context.Background(), "seg_123")
	if err != nil {
		t.Fatalf("GetCDPSegment failed: %v", err)
	}

	if segment.ID != "seg_123" {
		t.Errorf("Unexpected segment ID: %s", segment.ID)
	}
}

func TestGetCDPProfileEmptyID(t *testing.T) {
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
			_, err := client.GetCDPProfile(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "profile id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestGetCDPEventEmptyID(t *testing.T) {
	client := NewClient("token")

	_, err := client.GetCDPEvent(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
	if err != nil && err.Error() != "event id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestGetCDPSegmentEmptyID(t *testing.T) {
	client := NewClient("token")

	_, err := client.GetCDPSegment(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
	if err != nil && err.Error() != "segment id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}
