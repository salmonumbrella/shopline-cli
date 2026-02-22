package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeliveryOptionsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/delivery_options" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := DeliveryOptionsListResponse{
			Items: []DeliveryOption{
				{ID: "do_123", Name: "Standard Shipping", Type: "shipping", Status: "active"},
				{ID: "do_456", Name: "Express Delivery", Type: "shipping", Status: "active"},
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

	options, err := client.ListDeliveryOptions(context.Background(), &DeliveryOptionsListOptions{})
	if err != nil {
		t.Fatalf("ListDeliveryOptions failed: %v", err)
	}

	if len(options.Items) != 2 {
		t.Errorf("Expected 2 delivery options, got %d", len(options.Items))
	}
	if options.Items[0].ID != "do_123" {
		t.Errorf("Unexpected delivery option ID: %s", options.Items[0].ID)
	}
}

func TestDeliveryOptionsListWithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("status") != "active" {
			t.Errorf("Expected status=active, got %s", r.URL.Query().Get("status"))
		}
		if r.URL.Query().Get("type") != "pickup" {
			t.Errorf("Expected type=pickup, got %s", r.URL.Query().Get("type"))
		}

		resp := DeliveryOptionsListResponse{Items: []DeliveryOption{}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.ListDeliveryOptions(context.Background(), &DeliveryOptionsListOptions{
		Status: "active",
		Type:   "pickup",
	})
	if err != nil {
		t.Fatalf("ListDeliveryOptions failed: %v", err)
	}
}

func TestDeliveryOptionsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/delivery_options/do_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		opt := DeliveryOption{
			ID:          "do_123",
			Name:        "Standard Shipping",
			Type:        "shipping",
			Status:      "active",
			Description: "5-7 business days",
		}
		_ = json.NewEncoder(w).Encode(opt)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opt, err := client.GetDeliveryOption(context.Background(), "do_123")
	if err != nil {
		t.Fatalf("GetDeliveryOption failed: %v", err)
	}

	if opt.ID != "do_123" {
		t.Errorf("Unexpected delivery option ID: %s", opt.ID)
	}
	if opt.Name != "Standard Shipping" {
		t.Errorf("Unexpected delivery option name: %s", opt.Name)
	}
}

func TestGetDeliveryOptionEmptyID(t *testing.T) {
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
			_, err := client.GetDeliveryOption(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "delivery option id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeliveryOptionsUpdatePickupStore(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/delivery_options/do_123/pickup_store" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		opt := DeliveryOption{
			ID:     "do_123",
			Name:   "Store Pickup",
			Type:   "pickup",
			Status: "active",
		}
		_ = json.NewEncoder(w).Encode(opt)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &PickupStoreUpdateRequest{
		StoreID:   "store_456",
		StoreName: "Downtown Store",
		Address:   "123 Main St",
	}

	opt, err := client.UpdateDeliveryOptionPickupStore(context.Background(), "do_123", req)
	if err != nil {
		t.Fatalf("UpdateDeliveryOptionPickupStore failed: %v", err)
	}

	if opt.ID != "do_123" {
		t.Errorf("Unexpected delivery option ID: %s", opt.ID)
	}
}

func TestUpdateDeliveryOptionPickupStoreEmptyID(t *testing.T) {
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
			_, err := client.UpdateDeliveryOptionPickupStore(context.Background(), tc.id, &PickupStoreUpdateRequest{})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "delivery option id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeliveryTimeSlotsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/delivery_options/do_123/time_slots" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := DeliveryTimeSlotsListResponse{
			Items: []DeliveryTimeSlot{
				{ID: "ts_1", Date: "2024-01-15", StartTime: "09:00", EndTime: "12:00", Available: true},
				{ID: "ts_2", Date: "2024-01-15", StartTime: "14:00", EndTime: "18:00", Available: true},
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

	slots, err := client.ListDeliveryTimeSlots(context.Background(), "do_123", &DeliveryTimeSlotsListOptions{})
	if err != nil {
		t.Fatalf("ListDeliveryTimeSlots failed: %v", err)
	}

	if len(slots.Items) != 2 {
		t.Errorf("Expected 2 time slots, got %d", len(slots.Items))
	}
	if slots.Items[0].ID != "ts_1" {
		t.Errorf("Unexpected time slot ID: %s", slots.Items[0].ID)
	}
}

func TestDeliveryTimeSlotsWithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("start_date") != "2024-01-15" {
			t.Errorf("Expected start_date=2024-01-15, got %s", r.URL.Query().Get("start_date"))
		}
		if r.URL.Query().Get("end_date") != "2024-01-20" {
			t.Errorf("Expected end_date=2024-01-20, got %s", r.URL.Query().Get("end_date"))
		}

		resp := DeliveryTimeSlotsListResponse{Items: []DeliveryTimeSlot{}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.ListDeliveryTimeSlots(context.Background(), "do_123", &DeliveryTimeSlotsListOptions{
		StartDate: "2024-01-15",
		EndDate:   "2024-01-20",
	})
	if err != nil {
		t.Fatalf("ListDeliveryTimeSlots failed: %v", err)
	}
}

func TestListDeliveryTimeSlotsEmptyID(t *testing.T) {
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
			_, err := client.ListDeliveryTimeSlots(context.Background(), tc.id, &DeliveryTimeSlotsListOptions{})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "delivery option id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestGetDeliveryConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/delivery_options/delivery_config" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "shipping" {
			t.Errorf("Expected type=shipping, got %s", r.URL.Query().Get("type"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.GetDeliveryConfig(context.Background(), &DeliveryConfigOptions{Type: "shipping"})
	if err != nil {
		t.Fatalf("GetDeliveryConfig failed: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("failed to unmarshal raw response: %v", err)
	}
	if got["ok"] != true {
		t.Fatalf("expected ok=true, got %v", got["ok"])
	}
}

func TestGetDeliveryConfigEmptyType(t *testing.T) {
	client := NewClient("token")
	_, err := client.GetDeliveryConfig(context.Background(), &DeliveryConfigOptions{Type: " "})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "type is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetDeliveryConfigNilOptions(t *testing.T) {
	client := NewClient("token")
	_, err := client.GetDeliveryConfig(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "delivery config options are required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetDeliveryTimeSlotsOpenAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/delivery_options/do_123/delivery_time_slots" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.GetDeliveryTimeSlotsOpenAPI(context.Background(), "do_123")
	if err != nil {
		t.Fatalf("GetDeliveryTimeSlotsOpenAPI failed: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("failed to unmarshal raw response: %v", err)
	}
	if _, ok := got["items"]; !ok {
		t.Fatalf("expected items key in response, got %v", got)
	}
}

func TestGetDeliveryTimeSlotsOpenAPIEmptyID(t *testing.T) {
	client := NewClient("token")
	_, err := client.GetDeliveryTimeSlotsOpenAPI(context.Background(), " ")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "delivery option id is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateDeliveryOptionStoresInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/delivery_options/do_123/stores_info" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["stores"] == nil {
			t.Fatalf("expected stores in body, got %v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"updated": true})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.UpdateDeliveryOptionStoresInfo(context.Background(), "do_123", map[string]any{"stores": []any{}})
	if err != nil {
		t.Fatalf("UpdateDeliveryOptionStoresInfo failed: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("failed to unmarshal raw response: %v", err)
	}
	if got["updated"] != true {
		t.Fatalf("expected updated=true, got %v", got["updated"])
	}
}

func TestUpdateDeliveryOptionStoresInfoEmptyID(t *testing.T) {
	client := NewClient("token")
	_, err := client.UpdateDeliveryOptionStoresInfo(context.Background(), " ", map[string]any{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "delivery option id is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}
