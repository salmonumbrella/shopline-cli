package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInventoryLevelsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/inventory_levels" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := InventoryListResponse{
			Items: []InventoryLevel{
				{ID: "inv_123", InventoryItemID: "item_001", LocationID: "loc_a", Available: 50},
				{ID: "inv_456", InventoryItemID: "item_002", LocationID: "loc_a", Available: 25},
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

	levels, err := client.ListInventoryLevels(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListInventoryLevels failed: %v", err)
	}

	if len(levels.Items) != 2 {
		t.Errorf("Expected 2 inventory levels, got %d", len(levels.Items))
	}
	if levels.Items[0].ID != "inv_123" {
		t.Errorf("Unexpected inventory level ID: %s", levels.Items[0].ID)
	}
	if levels.Items[0].Available != 50 {
		t.Errorf("Expected Available=50, got %d", levels.Items[0].Available)
	}
}

func TestInventoryLevelsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		// Check query parameters
		if r.URL.Query().Get("location_id") != "loc_a" {
			t.Errorf("Expected location_id=loc_a, got %s", r.URL.Query().Get("location_id"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}

		resp := InventoryListResponse{
			Items:      []InventoryLevel{},
			Page:       2,
			PageSize:   10,
			TotalCount: 0,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &InventoryListOptions{
		LocationID: "loc_a",
		Page:       2,
		PageSize:   10,
	}

	_, err := client.ListInventoryLevels(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListInventoryLevels with options failed: %v", err)
	}
}

func TestInventoryLevelGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/inventory_levels/inv_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		level := InventoryLevel{
			ID:              "inv_123",
			InventoryItemID: "item_001",
			LocationID:      "loc_a",
			Available:       75,
		}
		_ = json.NewEncoder(w).Encode(level)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	level, err := client.GetInventoryLevel(context.Background(), "inv_123")
	if err != nil {
		t.Fatalf("GetInventoryLevel failed: %v", err)
	}

	if level.ID != "inv_123" {
		t.Errorf("Unexpected inventory level ID: %s", level.ID)
	}
	if level.Available != 75 {
		t.Errorf("Expected Available=75, got %d", level.Available)
	}
}

func TestInventoryAdjust(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/inventory_levels/inv_123/adjust" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req InventoryAdjustRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		if req.Delta != -5 {
			t.Errorf("Expected delta=-5, got %d", req.Delta)
		}

		level := InventoryLevel{
			ID:              "inv_123",
			InventoryItemID: "item_001",
			LocationID:      "loc_a",
			Available:       70, // Was 75, adjusted by -5
		}
		_ = json.NewEncoder(w).Encode(level)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	level, err := client.AdjustInventory(context.Background(), "inv_123", -5)
	if err != nil {
		t.Fatalf("AdjustInventory failed: %v", err)
	}

	if level.Available != 70 {
		t.Errorf("Expected Available=70 after adjustment, got %d", level.Available)
	}
}

func TestInventoryAdjustPositive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req InventoryAdjustRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		if req.Delta != 10 {
			t.Errorf("Expected delta=10, got %d", req.Delta)
		}

		level := InventoryLevel{
			ID:              "inv_456",
			InventoryItemID: "item_002",
			LocationID:      "loc_b",
			Available:       35, // Was 25, adjusted by +10
		}
		_ = json.NewEncoder(w).Encode(level)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	level, err := client.AdjustInventory(context.Background(), "inv_456", 10)
	if err != nil {
		t.Fatalf("AdjustInventory failed: %v", err)
	}

	if level.Available != 35 {
		t.Errorf("Expected Available=35 after adjustment, got %d", level.Available)
	}
}

func TestGetInventoryLevelEmptyID(t *testing.T) {
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
			_, err := client.GetInventoryLevel(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "inventory level id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestAdjustInventoryEmptyID(t *testing.T) {
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
			_, err := client.AdjustInventory(context.Background(), tc.id, 10)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "inventory level id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestInventoryLevelsAdjust(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/inventory_levels/adjust" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req InventoryLevelAdjustRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		level := InventoryLevel{
			ID:              "invlvl_123",
			InventoryItemID: req.InventoryItemID,
			LocationID:      req.LocationID,
			Available:       100 + req.AvailableAdjustment,
			OnHand:          120 + req.AvailableAdjustment,
		}
		_ = json.NewEncoder(w).Encode(level)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &InventoryLevelAdjustRequest{
		InventoryItemID:     "inv_001",
		LocationID:          "loc_001",
		AvailableAdjustment: 10,
	}
	level, err := client.AdjustInventoryLevel(context.Background(), req)
	if err != nil {
		t.Fatalf("AdjustInventoryLevel failed: %v", err)
	}

	if level.Available != 110 {
		t.Errorf("Expected available 110, got %d", level.Available)
	}
}

func TestInventoryLevelsSet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/inventory_levels/set" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req InventoryLevelSetRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		level := InventoryLevel{
			ID:              "invlvl_123",
			InventoryItemID: req.InventoryItemID,
			LocationID:      req.LocationID,
			Available:       req.Available,
			OnHand:          req.Available + 20,
		}
		_ = json.NewEncoder(w).Encode(level)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &InventoryLevelSetRequest{
		InventoryItemID: "inv_001",
		LocationID:      "loc_001",
		Available:       200,
	}
	level, err := client.SetInventoryLevel(context.Background(), req)
	if err != nil {
		t.Fatalf("SetInventoryLevel failed: %v", err)
	}

	if level.Available != 200 {
		t.Errorf("Expected available 200, got %d", level.Available)
	}
}
