package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWarehousesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/warehouses" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := WarehousesListResponse{
			Items: []Warehouse{
				{ID: "wh_123", Name: "Main Warehouse", City: "New York", Active: true},
				{ID: "wh_456", Name: "Secondary Warehouse", City: "Los Angeles", Active: true},
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

	warehouses, err := client.ListWarehouses(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListWarehouses failed: %v", err)
	}

	if len(warehouses.Items) != 2 {
		t.Errorf("Expected 2 warehouses, got %d", len(warehouses.Items))
	}
	if warehouses.Items[0].ID != "wh_123" {
		t.Errorf("Unexpected warehouse ID: %s", warehouses.Items[0].ID)
	}
}

func TestListWarehousesWithOptions(t *testing.T) {
	boolPtr := func(b bool) *bool { return &b }

	testCases := []struct {
		name           string
		opts           *WarehousesListOptions
		expectedParams map[string]string
	}{
		{
			name: "page only",
			opts: &WarehousesListOptions{Page: 2},
			expectedParams: map[string]string{
				"page": "2",
			},
		},
		{
			name: "page_size only",
			opts: &WarehousesListOptions{PageSize: 50},
			expectedParams: map[string]string{
				"page_size": "50",
			},
		},
		{
			name: "active true",
			opts: &WarehousesListOptions{Active: boolPtr(true)},
			expectedParams: map[string]string{
				"active": "true",
			},
		},
		{
			name: "active false",
			opts: &WarehousesListOptions{Active: boolPtr(false)},
			expectedParams: map[string]string{
				"active": "false",
			},
		},
		{
			name: "all options",
			opts: &WarehousesListOptions{Page: 3, PageSize: 25, Active: boolPtr(true)},
			expectedParams: map[string]string{
				"page":      "3",
				"page_size": "25",
				"active":    "true",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				query := r.URL.Query()
				for key, expected := range tc.expectedParams {
					if got := query.Get(key); got != expected {
						t.Errorf("Expected %s=%s, got %s", key, expected, got)
					}
				}

				resp := WarehousesListResponse{
					Items:      []Warehouse{{ID: "wh_123", Name: "Test"}},
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

			_, err := client.ListWarehouses(context.Background(), tc.opts)
			if err != nil {
				t.Fatalf("ListWarehouses failed: %v", err)
			}
		})
	}
}

func TestWarehousesGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/warehouses/wh_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		warehouse := Warehouse{ID: "wh_123", Name: "Main Warehouse", City: "New York"}
		_ = json.NewEncoder(w).Encode(warehouse)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	warehouse, err := client.GetWarehouse(context.Background(), "wh_123")
	if err != nil {
		t.Fatalf("GetWarehouse failed: %v", err)
	}

	if warehouse.ID != "wh_123" {
		t.Errorf("Unexpected warehouse ID: %s", warehouse.ID)
	}
}

func TestGetWarehouseEmptyID(t *testing.T) {
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
			_, err := client.GetWarehouse(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "warehouse id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestWarehousesCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		warehouse := Warehouse{ID: "wh_new", Name: "New Warehouse", City: "Chicago"}
		_ = json.NewEncoder(w).Encode(warehouse)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &WarehouseCreateRequest{
		Name:     "New Warehouse",
		Address1: "123 Main St",
		City:     "Chicago",
		Country:  "US",
	}

	warehouse, err := client.CreateWarehouse(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}

	if warehouse.ID != "wh_new" {
		t.Errorf("Unexpected warehouse ID: %s", warehouse.ID)
	}
}

func TestWarehousesUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}

		warehouse := Warehouse{ID: "wh_123", Name: "Updated Warehouse"}
		_ = json.NewEncoder(w).Encode(warehouse)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &WarehouseUpdateRequest{
		Name: "Updated Warehouse",
	}

	warehouse, err := client.UpdateWarehouse(context.Background(), "wh_123", req)
	if err != nil {
		t.Fatalf("UpdateWarehouse failed: %v", err)
	}

	if warehouse.Name != "Updated Warehouse" {
		t.Errorf("Unexpected warehouse name: %s", warehouse.Name)
	}
}

func TestUpdateWarehouseEmptyID(t *testing.T) {
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
			req := &WarehouseUpdateRequest{Name: "Test"}
			_, err := client.UpdateWarehouse(context.Background(), tc.id, req)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "warehouse id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestWarehousesDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/warehouses/wh_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteWarehouse(context.Background(), "wh_123")
	if err != nil {
		t.Fatalf("DeleteWarehouse failed: %v", err)
	}
}

func TestDeleteWarehouseEmptyID(t *testing.T) {
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
			err := client.DeleteWarehouse(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "warehouse id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestWarehousesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	// Test List API error
	_, err := client.ListWarehouses(context.Background(), nil)
	if err == nil {
		t.Error("Expected error from ListWarehouses")
	}

	// Test Get API error
	_, err = client.GetWarehouse(context.Background(), "wh_123")
	if err == nil {
		t.Error("Expected error from GetWarehouse")
	}

	// Test Create API error
	_, err = client.CreateWarehouse(context.Background(), &WarehouseCreateRequest{Name: "Test"})
	if err == nil {
		t.Error("Expected error from CreateWarehouse")
	}

	// Test Update API error
	_, err = client.UpdateWarehouse(context.Background(), "wh_123", &WarehouseUpdateRequest{Name: "Test"})
	if err == nil {
		t.Error("Expected error from UpdateWarehouse")
	}

	// Test Delete API error
	err = client.DeleteWarehouse(context.Background(), "wh_123")
	if err == nil {
		t.Error("Expected error from DeleteWarehouse")
	}
}

func TestWarehouseFullResponseFields(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 16, 14, 30, 0, 0, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		warehouse := Warehouse{
			ID:           "wh_full",
			Name:         "Full Warehouse",
			Code:         "WH001",
			Address1:     "123 Main Street",
			Address2:     "Suite 100",
			City:         "New York",
			Province:     "New York",
			ProvinceCode: "NY",
			Country:      "United States",
			CountryCode:  "US",
			Zip:          "10001",
			Phone:        "+1-555-123-4567",
			Email:        "warehouse@example.com",
			Active:       true,
			IsDefault:    true,
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
		}
		_ = json.NewEncoder(w).Encode(warehouse)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	warehouse, err := client.GetWarehouse(context.Background(), "wh_full")
	if err != nil {
		t.Fatalf("GetWarehouse failed: %v", err)
	}

	if warehouse.ID != "wh_full" {
		t.Errorf("Expected ID wh_full, got %s", warehouse.ID)
	}
	if warehouse.Name != "Full Warehouse" {
		t.Errorf("Expected Name 'Full Warehouse', got %s", warehouse.Name)
	}
	if warehouse.Code != "WH001" {
		t.Errorf("Expected Code WH001, got %s", warehouse.Code)
	}
	if warehouse.Address1 != "123 Main Street" {
		t.Errorf("Expected Address1 '123 Main Street', got %s", warehouse.Address1)
	}
	if warehouse.Address2 != "Suite 100" {
		t.Errorf("Expected Address2 'Suite 100', got %s", warehouse.Address2)
	}
	if warehouse.City != "New York" {
		t.Errorf("Expected City 'New York', got %s", warehouse.City)
	}
	if warehouse.Province != "New York" {
		t.Errorf("Expected Province 'New York', got %s", warehouse.Province)
	}
	if warehouse.ProvinceCode != "NY" {
		t.Errorf("Expected ProvinceCode NY, got %s", warehouse.ProvinceCode)
	}
	if warehouse.Country != "United States" {
		t.Errorf("Expected Country 'United States', got %s", warehouse.Country)
	}
	if warehouse.CountryCode != "US" {
		t.Errorf("Expected CountryCode US, got %s", warehouse.CountryCode)
	}
	if warehouse.Zip != "10001" {
		t.Errorf("Expected Zip 10001, got %s", warehouse.Zip)
	}
	if warehouse.Phone != "+1-555-123-4567" {
		t.Errorf("Expected Phone '+1-555-123-4567', got %s", warehouse.Phone)
	}
	if warehouse.Email != "warehouse@example.com" {
		t.Errorf("Expected Email 'warehouse@example.com', got %s", warehouse.Email)
	}
	if !warehouse.Active {
		t.Error("Expected Active to be true")
	}
	if !warehouse.IsDefault {
		t.Error("Expected IsDefault to be true")
	}
	if !warehouse.CreatedAt.Equal(createdAt) {
		t.Errorf("Expected CreatedAt %v, got %v", createdAt, warehouse.CreatedAt)
	}
	if !warehouse.UpdatedAt.Equal(updatedAt) {
		t.Errorf("Expected UpdatedAt %v, got %v", updatedAt, warehouse.UpdatedAt)
	}
}

func TestWarehouseCreateRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/warehouses" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req WarehouseCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.Name != "Test Warehouse" {
			t.Errorf("Expected Name 'Test Warehouse', got %s", req.Name)
		}
		if req.Code != "TW001" {
			t.Errorf("Expected Code TW001, got %s", req.Code)
		}
		if req.Address1 != "456 Oak Ave" {
			t.Errorf("Expected Address1 '456 Oak Ave', got %s", req.Address1)
		}
		if req.Address2 != "Floor 2" {
			t.Errorf("Expected Address2 'Floor 2', got %s", req.Address2)
		}
		if req.City != "Chicago" {
			t.Errorf("Expected City Chicago, got %s", req.City)
		}
		if req.Province != "Illinois" {
			t.Errorf("Expected Province Illinois, got %s", req.Province)
		}
		if req.Country != "US" {
			t.Errorf("Expected Country US, got %s", req.Country)
		}
		if req.Zip != "60601" {
			t.Errorf("Expected Zip 60601, got %s", req.Zip)
		}
		if req.Phone != "+1-312-555-0100" {
			t.Errorf("Expected Phone '+1-312-555-0100', got %s", req.Phone)
		}
		if req.Email != "chicago@example.com" {
			t.Errorf("Expected Email 'chicago@example.com', got %s", req.Email)
		}

		warehouse := Warehouse{ID: "wh_created", Name: req.Name}
		_ = json.NewEncoder(w).Encode(warehouse)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &WarehouseCreateRequest{
		Name:     "Test Warehouse",
		Code:     "TW001",
		Address1: "456 Oak Ave",
		Address2: "Floor 2",
		City:     "Chicago",
		Province: "Illinois",
		Country:  "US",
		Zip:      "60601",
		Phone:    "+1-312-555-0100",
		Email:    "chicago@example.com",
	}

	warehouse, err := client.CreateWarehouse(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}

	if warehouse.ID != "wh_created" {
		t.Errorf("Unexpected warehouse ID: %s", warehouse.ID)
	}
}

func TestWarehouseUpdateRequestBody(t *testing.T) {
	boolPtr := func(b bool) *bool { return &b }

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/warehouses/wh_update" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req WarehouseUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.Name != "Updated Name" {
			t.Errorf("Expected Name 'Updated Name', got %s", req.Name)
		}
		if req.Code != "UPD001" {
			t.Errorf("Expected Code UPD001, got %s", req.Code)
		}
		if req.Address1 != "789 New Street" {
			t.Errorf("Expected Address1 '789 New Street', got %s", req.Address1)
		}
		if req.Active == nil || *req.Active != false {
			t.Error("Expected Active to be false")
		}

		warehouse := Warehouse{ID: "wh_update", Name: req.Name, Active: false}
		_ = json.NewEncoder(w).Encode(warehouse)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &WarehouseUpdateRequest{
		Name:     "Updated Name",
		Code:     "UPD001",
		Address1: "789 New Street",
		Active:   boolPtr(false),
	}

	warehouse, err := client.UpdateWarehouse(context.Background(), "wh_update", req)
	if err != nil {
		t.Fatalf("UpdateWarehouse failed: %v", err)
	}

	if warehouse.ID != "wh_update" {
		t.Errorf("Unexpected warehouse ID: %s", warehouse.ID)
	}
	if warehouse.Active {
		t.Error("Expected warehouse Active to be false")
	}
}

func TestListWarehousesNoOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify no query parameters are sent when options have zero values
		if r.URL.RawQuery != "" {
			t.Errorf("Expected no query params, got: %s", r.URL.RawQuery)
		}

		resp := WarehousesListResponse{
			Items:      []Warehouse{{ID: "wh_123", Name: "Test"}},
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

	// Test with zero-value options (should not add query params)
	opts := &WarehousesListOptions{
		Page:     0,
		PageSize: 0,
		Active:   nil,
	}
	_, err := client.ListWarehouses(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListWarehouses failed: %v", err)
	}
}

func TestWarehouseDeleteWithStatusOK(t *testing.T) {
	// Some APIs return 200 OK instead of 204 No Content
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteWarehouse(context.Background(), "wh_123")
	if err != nil {
		t.Fatalf("DeleteWarehouse failed: %v", err)
	}
}

func TestWarehousesListEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := WarehousesListResponse{
			Items:      []Warehouse{},
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

	warehouses, err := client.ListWarehouses(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListWarehouses failed: %v", err)
	}

	if len(warehouses.Items) != 0 {
		t.Errorf("Expected 0 warehouses, got %d", len(warehouses.Items))
	}
	if warehouses.TotalCount != 0 {
		t.Errorf("Expected TotalCount 0, got %d", warehouses.TotalCount)
	}
}
