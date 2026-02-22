package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPickupLocationsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/pickup" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := PickupListResponse{
			Items: []PickupLocation{
				{ID: "pk_123", Name: "Main Store", City: "New York", Active: true},
				{ID: "pk_456", Name: "Downtown Branch", City: "Los Angeles", Active: true},
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

	locations, err := client.ListPickupLocations(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListPickupLocations failed: %v", err)
	}

	if len(locations.Items) != 2 {
		t.Errorf("Expected 2 locations, got %d", len(locations.Items))
	}
	if locations.Items[0].ID != "pk_123" {
		t.Errorf("Unexpected location ID: %s", locations.Items[0].ID)
	}
}

func TestPickupLocationGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pickup/pk_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		location := PickupLocation{
			ID:       "pk_123",
			Name:     "Main Store",
			Address1: "123 Main St",
			City:     "New York",
			Country:  "United States",
			Active:   true,
			Hours: []PickupHours{
				{Day: "monday", OpenTime: "09:00", CloseTime: "18:00", Closed: false},
				{Day: "sunday", Closed: true},
			},
		}
		_ = json.NewEncoder(w).Encode(location)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	location, err := client.GetPickupLocation(context.Background(), "pk_123")
	if err != nil {
		t.Fatalf("GetPickupLocation failed: %v", err)
	}

	if location.ID != "pk_123" {
		t.Errorf("Unexpected location ID: %s", location.ID)
	}
	if len(location.Hours) != 2 {
		t.Errorf("Expected 2 hours entries, got %d", len(location.Hours))
	}
}

func TestGetPickupLocationEmptyID(t *testing.T) {
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
			_, err := client.GetPickupLocation(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "pickup location id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCreatePickupLocation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/pickup" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		location := PickupLocation{ID: "pk_new", Name: "New Store", City: "Chicago", Active: true}
		_ = json.NewEncoder(w).Encode(location)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &PickupCreateRequest{
		Name:     "New Store",
		Address1: "456 Oak Ave",
		City:     "Chicago",
		Country:  "United States",
		Active:   true,
	}

	location, err := client.CreatePickupLocation(context.Background(), req)
	if err != nil {
		t.Fatalf("CreatePickupLocation failed: %v", err)
	}

	if location.ID != "pk_new" {
		t.Errorf("Unexpected location ID: %s", location.ID)
	}
}

func TestUpdatePickupLocation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/pickup/pk_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		location := PickupLocation{ID: "pk_123", Name: "Updated Store", Active: false}
		_ = json.NewEncoder(w).Encode(location)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	active := false
	req := &PickupUpdateRequest{
		Active: &active,
	}

	location, err := client.UpdatePickupLocation(context.Background(), "pk_123", req)
	if err != nil {
		t.Fatalf("UpdatePickupLocation failed: %v", err)
	}

	if location.Active != false {
		t.Errorf("Expected active to be false, got true")
	}
}

func TestDeletePickupLocation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/pickup/pk_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeletePickupLocation(context.Background(), "pk_123")
	if err != nil {
		t.Fatalf("DeletePickupLocation failed: %v", err)
	}
}

func TestDeletePickupLocationEmptyID(t *testing.T) {
	client := NewClient("token")

	err := client.DeletePickupLocation(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
	if err != nil && err.Error() != "pickup location id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestUpdatePickupLocationEmptyID(t *testing.T) {
	client := NewClient("token")

	_, err := client.UpdatePickupLocation(context.Background(), "", nil)
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
	if err != nil && err.Error() != "pickup location id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}
