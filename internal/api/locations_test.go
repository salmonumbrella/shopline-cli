package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLocationsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/locations" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := LocationsListResponse{
			Items: []Location{
				{ID: "loc_123", Name: "Main Warehouse", City: "New York", Active: true},
				{ID: "loc_456", Name: "West Coast", City: "Los Angeles", Active: true},
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

	locations, err := client.ListLocations(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListLocations failed: %v", err)
	}

	if len(locations.Items) != 2 {
		t.Errorf("Expected 2 locations, got %d", len(locations.Items))
	}
	if locations.Items[0].ID != "loc_123" {
		t.Errorf("Unexpected location ID: %s", locations.Items[0].ID)
	}
}

func TestLocationsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/locations/loc_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		location := Location{ID: "loc_123", Name: "Main Warehouse", City: "New York"}
		_ = json.NewEncoder(w).Encode(location)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	location, err := client.GetLocation(context.Background(), "loc_123")
	if err != nil {
		t.Fatalf("GetLocation failed: %v", err)
	}

	if location.ID != "loc_123" {
		t.Errorf("Unexpected location ID: %s", location.ID)
	}
}

func TestGetLocationEmptyID(t *testing.T) {
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
			_, err := client.GetLocation(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "location id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestLocationsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/locations" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		location := Location{ID: "loc_new", Name: "New Location", City: "Chicago"}
		_ = json.NewEncoder(w).Encode(location)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &LocationCreateRequest{Name: "New Location", Address1: "123 Main St", City: "Chicago", Country: "US"}
	location, err := client.CreateLocation(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateLocation failed: %v", err)
	}

	if location.ID != "loc_new" {
		t.Errorf("Unexpected location ID: %s", location.ID)
	}
}

func TestLocationsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/locations/loc_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteLocation(context.Background(), "loc_123")
	if err != nil {
		t.Fatalf("DeleteLocation failed: %v", err)
	}
}

func TestDeleteLocationEmptyID(t *testing.T) {
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
			err := client.DeleteLocation(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "location id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestLocationsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/locations/loc_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		location := Location{ID: "loc_123", Name: "Updated Warehouse", City: "New York"}
		_ = json.NewEncoder(w).Encode(location)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &LocationUpdateRequest{Name: "Updated Warehouse"}
	location, err := client.UpdateLocation(context.Background(), "loc_123", req)
	if err != nil {
		t.Fatalf("UpdateLocation failed: %v", err)
	}

	if location.ID != "loc_123" {
		t.Errorf("Unexpected location ID: %s", location.ID)
	}
	if location.Name != "Updated Warehouse" {
		t.Errorf("Unexpected location name: %s", location.Name)
	}
}

func TestUpdateLocationEmptyID(t *testing.T) {
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
			req := &LocationUpdateRequest{Name: "Test"}
			_, err := client.UpdateLocation(context.Background(), tc.id, req)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "location id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestLocationsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", query.Get("page"))
		}
		if query.Get("page_size") != "50" {
			t.Errorf("Expected page_size=50, got %s", query.Get("page_size"))
		}
		if query.Get("active") != "true" {
			t.Errorf("Expected active=true, got %s", query.Get("active"))
		}

		resp := LocationsListResponse{
			Items:      []Location{{ID: "loc_123", Name: "Main Warehouse", Active: true}},
			Page:       2,
			PageSize:   50,
			TotalCount: 100,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	active := true
	opts := &LocationsListOptions{
		Page:     2,
		PageSize: 50,
		Active:   &active,
	}
	locations, err := client.ListLocations(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListLocations failed: %v", err)
	}

	if len(locations.Items) != 1 {
		t.Errorf("Expected 1 location, got %d", len(locations.Items))
	}
}
