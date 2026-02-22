package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLocalDeliveryList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/local_delivery" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := LocalDeliveryListResponse{
			Items: []LocalDeliveryOption{
				{ID: "ld_123", Name: "Same Day Delivery", Price: "5.00", Active: true},
				{ID: "ld_456", Name: "Next Day Delivery", Price: "3.00", Active: true},
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

	options, err := client.ListLocalDeliveryOptions(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListLocalDeliveryOptions failed: %v", err)
	}

	if len(options.Items) != 2 {
		t.Errorf("Expected 2 options, got %d", len(options.Items))
	}
	if options.Items[0].ID != "ld_123" {
		t.Errorf("Unexpected option ID: %s", options.Items[0].ID)
	}
}

func TestLocalDeliveryGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/local_delivery/ld_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		option := LocalDeliveryOption{
			ID:     "ld_123",
			Name:   "Same Day Delivery",
			Price:  "5.00",
			Active: true,
			Zones: []LocalDeliveryZone{
				{ID: "zone_1", Name: "Downtown", Type: "zip_code", ZipCodes: []string{"10001", "10002"}},
			},
		}
		_ = json.NewEncoder(w).Encode(option)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	option, err := client.GetLocalDeliveryOption(context.Background(), "ld_123")
	if err != nil {
		t.Fatalf("GetLocalDeliveryOption failed: %v", err)
	}

	if option.ID != "ld_123" {
		t.Errorf("Unexpected option ID: %s", option.ID)
	}
	if len(option.Zones) != 1 {
		t.Errorf("Expected 1 zone, got %d", len(option.Zones))
	}
}

func TestGetLocalDeliveryOptionEmptyID(t *testing.T) {
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
			_, err := client.GetLocalDeliveryOption(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "local delivery option id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCreateLocalDeliveryOption(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/local_delivery" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		option := LocalDeliveryOption{ID: "ld_new", Name: "Express Delivery", Price: "10.00", Active: true}
		_ = json.NewEncoder(w).Encode(option)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &LocalDeliveryCreateRequest{
		Name:   "Express Delivery",
		Price:  "10.00",
		Active: true,
	}

	option, err := client.CreateLocalDeliveryOption(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateLocalDeliveryOption failed: %v", err)
	}

	if option.ID != "ld_new" {
		t.Errorf("Unexpected option ID: %s", option.ID)
	}
}

func TestUpdateLocalDeliveryOption(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/local_delivery/ld_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		option := LocalDeliveryOption{ID: "ld_123", Name: "Updated Delivery", Price: "7.00", Active: true}
		_ = json.NewEncoder(w).Encode(option)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	newPrice := "7.00"
	req := &LocalDeliveryUpdateRequest{
		Price: &newPrice,
	}

	option, err := client.UpdateLocalDeliveryOption(context.Background(), "ld_123", req)
	if err != nil {
		t.Fatalf("UpdateLocalDeliveryOption failed: %v", err)
	}

	if option.Price != "7.00" {
		t.Errorf("Expected price '7.00', got '%s'", option.Price)
	}
}

func TestDeleteLocalDeliveryOption(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/local_delivery/ld_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteLocalDeliveryOption(context.Background(), "ld_123")
	if err != nil {
		t.Fatalf("DeleteLocalDeliveryOption failed: %v", err)
	}
}

func TestDeleteLocalDeliveryOptionEmptyID(t *testing.T) {
	client := NewClient("token")

	err := client.DeleteLocalDeliveryOption(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
	if err != nil && err.Error() != "local delivery option id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestUpdateLocalDeliveryOptionEmptyID(t *testing.T) {
	client := NewClient("token")

	_, err := client.UpdateLocalDeliveryOption(context.Background(), "", nil)
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
	if err != nil && err.Error() != "local delivery option id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}
