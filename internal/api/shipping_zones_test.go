package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestShippingZonesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/shipping_zones" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := ShippingZonesListResponse{
			Items: []ShippingZone{
				{ID: "sz_123", Name: "Domestic"},
				{ID: "sz_456", Name: "International"},
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

	zones, err := client.ListShippingZones(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListShippingZones failed: %v", err)
	}

	if len(zones.Items) != 2 {
		t.Errorf("Expected 2 shipping zones, got %d", len(zones.Items))
	}
	if zones.Items[0].ID != "sz_123" {
		t.Errorf("Unexpected shipping zone ID: %s", zones.Items[0].ID)
	}
}

func TestShippingZonesGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/shipping_zones/sz_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		zone := ShippingZone{
			ID:   "sz_123",
			Name: "Domestic",
			Countries: []ZoneCountry{
				{Code: "US", Name: "United States"},
			},
		}
		_ = json.NewEncoder(w).Encode(zone)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	zone, err := client.GetShippingZone(context.Background(), "sz_123")
	if err != nil {
		t.Fatalf("GetShippingZone failed: %v", err)
	}

	if zone.ID != "sz_123" {
		t.Errorf("Unexpected shipping zone ID: %s", zone.ID)
	}
}

func TestGetShippingZoneEmptyID(t *testing.T) {
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
			_, err := client.GetShippingZone(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "shipping zone id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestShippingZonesCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		zone := ShippingZone{ID: "sz_new", Name: "New Zone"}
		_ = json.NewEncoder(w).Encode(zone)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ShippingZoneCreateRequest{
		Name: "New Zone",
	}

	zone, err := client.CreateShippingZone(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateShippingZone failed: %v", err)
	}

	if zone.ID != "sz_new" {
		t.Errorf("Unexpected shipping zone ID: %s", zone.ID)
	}
}

func TestShippingZonesDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/shipping_zones/sz_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteShippingZone(context.Background(), "sz_123")
	if err != nil {
		t.Fatalf("DeleteShippingZone failed: %v", err)
	}
}

func TestListShippingZonesWithOptions(t *testing.T) {
	testCases := []struct {
		name          string
		opts          *ShippingZonesListOptions
		expectedQuery map[string]string
	}{
		{
			name:          "page only",
			opts:          &ShippingZonesListOptions{Page: 2},
			expectedQuery: map[string]string{"page": "2"},
		},
		{
			name:          "page_size only",
			opts:          &ShippingZonesListOptions{PageSize: 50},
			expectedQuery: map[string]string{"page_size": "50"},
		},
		{
			name:          "all options combined",
			opts:          &ShippingZonesListOptions{Page: 3, PageSize: 25},
			expectedQuery: map[string]string{"page": "3", "page_size": "25"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET, got %s", r.Method)
				}

				query := r.URL.Query()
				for key, expectedValue := range tc.expectedQuery {
					if got := query.Get(key); got != expectedValue {
						t.Errorf("Expected %s=%s, got %s=%s", key, expectedValue, key, got)
					}
				}

				resp := ShippingZonesListResponse{
					Items:      []ShippingZone{{ID: "sz_123", Name: "Test Zone"}},
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

			_, err := client.ListShippingZones(context.Background(), tc.opts)
			if err != nil {
				t.Fatalf("ListShippingZones failed: %v", err)
			}
		})
	}
}

func TestDeleteShippingZoneEmptyID(t *testing.T) {
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
			err := client.DeleteShippingZone(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "shipping zone id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
