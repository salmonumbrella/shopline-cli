package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOrderAttributionList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/order_attributions" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := OrderAttributionListResponse{
			Items: []OrderAttribution{
				{ID: "attr_123", OrderID: "ord_123", Source: "google", Medium: "cpc"},
				{ID: "attr_456", OrderID: "ord_456", Source: "facebook", Medium: "social"},
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

	attributions, err := client.ListOrderAttributions(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListOrderAttributions failed: %v", err)
	}

	if len(attributions.Items) != 2 {
		t.Errorf("Expected 2 attributions, got %d", len(attributions.Items))
	}
	if attributions.Items[0].ID != "attr_123" {
		t.Errorf("Unexpected attribution ID: %s", attributions.Items[0].ID)
	}
}

func TestOrderAttributionListWithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("source") != "google" {
			t.Errorf("Expected source=google, got %s", r.URL.Query().Get("source"))
		}
		if r.URL.Query().Get("medium") != "cpc" {
			t.Errorf("Expected medium=cpc, got %s", r.URL.Query().Get("medium"))
		}

		resp := OrderAttributionListResponse{
			Items:      []OrderAttribution{{ID: "attr_123", Source: "google", Medium: "cpc"}},
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

	opts := &OrderAttributionListOptions{
		Source: "google",
		Medium: "cpc",
	}

	_, err := client.ListOrderAttributions(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListOrderAttributions with filters failed: %v", err)
	}
}

func TestGetOrderAttribution(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orders/ord_123/attribution" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		attr := OrderAttribution{
			ID:          "attr_123",
			OrderID:     "ord_123",
			Source:      "google",
			Medium:      "cpc",
			Campaign:    "summer_sale",
			UtmSource:   "google",
			UtmMedium:   "cpc",
			UtmCampaign: "summer_sale",
		}
		_ = json.NewEncoder(w).Encode(attr)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	attr, err := client.GetOrderAttribution(context.Background(), "ord_123")
	if err != nil {
		t.Fatalf("GetOrderAttribution failed: %v", err)
	}

	if attr.ID != "attr_123" {
		t.Errorf("Unexpected attribution ID: %s", attr.ID)
	}
	if attr.Source != "google" {
		t.Errorf("Unexpected source: %s", attr.Source)
	}
}

func TestGetOrderAttributionEmptyID(t *testing.T) {
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
			_, err := client.GetOrderAttribution(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
