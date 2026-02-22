package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCarrierServicesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/carrier_services" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := CarrierServicesListResponse{
			Items: []CarrierService{
				{ID: "cs_123", Name: "Express Shipping", Active: true},
				{ID: "cs_456", Name: "Standard Shipping", Active: true},
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

	services, err := client.ListCarrierServices(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListCarrierServices failed: %v", err)
	}

	if len(services.Items) != 2 {
		t.Errorf("Expected 2 carrier services, got %d", len(services.Items))
	}
	if services.Items[0].ID != "cs_123" {
		t.Errorf("Unexpected carrier service ID: %s", services.Items[0].ID)
	}
}

func TestCarrierServicesGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/carrier_services/cs_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		cs := CarrierService{ID: "cs_123", Name: "Express Shipping"}
		_ = json.NewEncoder(w).Encode(cs)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	cs, err := client.GetCarrierService(context.Background(), "cs_123")
	if err != nil {
		t.Fatalf("GetCarrierService failed: %v", err)
	}

	if cs.ID != "cs_123" {
		t.Errorf("Unexpected carrier service ID: %s", cs.ID)
	}
}

func TestGetCarrierServiceEmptyID(t *testing.T) {
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
			_, err := client.GetCarrierService(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "carrier service id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCarrierServicesCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		cs := CarrierService{ID: "cs_new", Name: "New Carrier"}
		_ = json.NewEncoder(w).Encode(cs)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CarrierServiceCreateRequest{
		Name:        "New Carrier",
		CallbackURL: "https://example.com/rates",
	}

	cs, err := client.CreateCarrierService(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateCarrierService failed: %v", err)
	}

	if cs.ID != "cs_new" {
		t.Errorf("Unexpected carrier service ID: %s", cs.ID)
	}
}

func TestCarrierServicesUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}

		cs := CarrierService{ID: "cs_123", Name: "Updated Carrier"}
		_ = json.NewEncoder(w).Encode(cs)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CarrierServiceUpdateRequest{
		Name: "Updated Carrier",
	}

	cs, err := client.UpdateCarrierService(context.Background(), "cs_123", req)
	if err != nil {
		t.Fatalf("UpdateCarrierService failed: %v", err)
	}

	if cs.Name != "Updated Carrier" {
		t.Errorf("Unexpected carrier name: %s", cs.Name)
	}
}

func TestCarrierServicesDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/carrier_services/cs_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteCarrierService(context.Background(), "cs_123")
	if err != nil {
		t.Fatalf("DeleteCarrierService failed: %v", err)
	}
}

func TestListCarrierServicesWithOptions(t *testing.T) {
	boolPtr := func(b bool) *bool { return &b }

	testCases := []struct {
		name          string
		opts          *CarrierServicesListOptions
		expectedQuery map[string]string
	}{
		{
			name:          "page only",
			opts:          &CarrierServicesListOptions{Page: 2},
			expectedQuery: map[string]string{"page": "2"},
		},
		{
			name:          "page_size only",
			opts:          &CarrierServicesListOptions{PageSize: 50},
			expectedQuery: map[string]string{"page_size": "50"},
		},
		{
			name:          "active true",
			opts:          &CarrierServicesListOptions{Active: boolPtr(true)},
			expectedQuery: map[string]string{"active": "true"},
		},
		{
			name:          "active false",
			opts:          &CarrierServicesListOptions{Active: boolPtr(false)},
			expectedQuery: map[string]string{"active": "false"},
		},
		{
			name:          "all options combined",
			opts:          &CarrierServicesListOptions{Page: 3, PageSize: 25, Active: boolPtr(true)},
			expectedQuery: map[string]string{"page": "3", "page_size": "25", "active": "true"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET, got %s", r.Method)
				}

				// Verify query parameters
				query := r.URL.Query()
				for key, expectedValue := range tc.expectedQuery {
					if got := query.Get(key); got != expectedValue {
						t.Errorf("Expected %s=%s, got %s=%s", key, expectedValue, key, got)
					}
				}

				resp := CarrierServicesListResponse{
					Items:      []CarrierService{{ID: "cs_123", Name: "Test Carrier"}},
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

			services, err := client.ListCarrierServices(context.Background(), tc.opts)
			if err != nil {
				t.Fatalf("ListCarrierServices failed: %v", err)
			}

			if len(services.Items) != 1 {
				t.Errorf("Expected 1 carrier service, got %d", len(services.Items))
			}
		})
	}
}

func TestUpdateCarrierServiceEmptyID(t *testing.T) {
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
			req := &CarrierServiceUpdateRequest{Name: "Updated Name"}
			_, err := client.UpdateCarrierService(context.Background(), tc.id, req)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "carrier service id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteCarrierServiceEmptyID(t *testing.T) {
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
			err := client.DeleteCarrierService(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "carrier service id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCarrierServicesCreateRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/carrier_services" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		// Verify request body
		var req CarrierServiceCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.Name != "Express Shipping" {
			t.Errorf("Expected name 'Express Shipping', got '%s'", req.Name)
		}
		if req.CallbackURL != "https://example.com/rates" {
			t.Errorf("Expected callback_url 'https://example.com/rates', got '%s'", req.CallbackURL)
		}
		if req.CarrierServiceType != "api" {
			t.Errorf("Expected carrier_service_type 'api', got '%s'", req.CarrierServiceType)
		}
		if req.Format != "json" {
			t.Errorf("Expected format 'json', got '%s'", req.Format)
		}
		if !req.Active {
			t.Error("Expected active to be true")
		}
		if !req.ServiceDiscovery {
			t.Error("Expected service_discovery to be true")
		}

		cs := CarrierService{
			ID:                 "cs_new",
			Name:               req.Name,
			CallbackURL:        req.CallbackURL,
			Active:             req.Active,
			ServiceDiscovery:   req.ServiceDiscovery,
			CarrierServiceType: req.CarrierServiceType,
			Format:             req.Format,
		}
		_ = json.NewEncoder(w).Encode(cs)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CarrierServiceCreateRequest{
		Name:               "Express Shipping",
		CallbackURL:        "https://example.com/rates",
		Active:             true,
		ServiceDiscovery:   true,
		CarrierServiceType: "api",
		Format:             "json",
	}

	cs, err := client.CreateCarrierService(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateCarrierService failed: %v", err)
	}

	if cs.ID != "cs_new" {
		t.Errorf("Unexpected carrier service ID: %s", cs.ID)
	}
	if cs.Name != "Express Shipping" {
		t.Errorf("Unexpected name: %s", cs.Name)
	}
}

func TestCarrierServicesUpdateRequestBody(t *testing.T) {
	boolPtr := func(b bool) *bool { return &b }

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/carrier_services/cs_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		// Verify request body
		var req CarrierServiceUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.Name != "Updated Shipping" {
			t.Errorf("Expected name 'Updated Shipping', got '%s'", req.Name)
		}
		if req.CallbackURL != "https://example.com/new-rates" {
			t.Errorf("Expected callback_url 'https://example.com/new-rates', got '%s'", req.CallbackURL)
		}
		if req.Active == nil || *req.Active != false {
			t.Error("Expected active to be false")
		}
		if req.ServiceDiscovery == nil || *req.ServiceDiscovery != true {
			t.Error("Expected service_discovery to be true")
		}

		cs := CarrierService{
			ID:               "cs_123",
			Name:             req.Name,
			CallbackURL:      req.CallbackURL,
			Active:           *req.Active,
			ServiceDiscovery: *req.ServiceDiscovery,
		}
		_ = json.NewEncoder(w).Encode(cs)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CarrierServiceUpdateRequest{
		Name:             "Updated Shipping",
		CallbackURL:      "https://example.com/new-rates",
		Active:           boolPtr(false),
		ServiceDiscovery: boolPtr(true),
	}

	cs, err := client.UpdateCarrierService(context.Background(), "cs_123", req)
	if err != nil {
		t.Fatalf("UpdateCarrierService failed: %v", err)
	}

	if cs.Name != "Updated Shipping" {
		t.Errorf("Unexpected name: %s", cs.Name)
	}
	if cs.Active != false {
		t.Error("Expected active to be false")
	}
}

func TestCarrierServicesListHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.ListCarrierServices(context.Background(), nil)
	if err == nil {
		t.Fatal("Expected error for HTTP 500, got nil")
	}
}

func TestCarrierServicesGetHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.GetCarrierService(context.Background(), "nonexistent_id")
	if err == nil {
		t.Fatal("Expected error for HTTP 404, got nil")
	}
}

func TestCarrierServicesCreateHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CarrierServiceCreateRequest{
		Name:        "Test",
		CallbackURL: "invalid-url",
	}

	_, err := client.CreateCarrierService(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for HTTP 400, got nil")
	}
}

func TestCarrierServicesUpdateHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "forbidden"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CarrierServiceUpdateRequest{Name: "Updated"}
	_, err := client.UpdateCarrierService(context.Background(), "cs_123", req)
	if err == nil {
		t.Fatal("Expected error for HTTP 403, got nil")
	}
}

func TestCarrierServicesDeleteHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteCarrierService(context.Background(), "nonexistent_id")
	if err == nil {
		t.Fatal("Expected error for HTTP 404, got nil")
	}
}

func TestCarrierServicesFullResponseFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := CarrierServicesListResponse{
			Items: []CarrierService{
				{
					ID:                 "cs_full",
					Name:               "Full Service",
					CallbackURL:        "https://example.com/callback",
					Active:             true,
					ServiceDiscovery:   true,
					CarrierServiceType: "api",
					Format:             "json",
				},
			},
			Page:       2,
			PageSize:   10,
			TotalCount: 15,
			HasMore:    true,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	services, err := client.ListCarrierServices(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListCarrierServices failed: %v", err)
	}

	// Verify response metadata
	if services.Page != 2 {
		t.Errorf("Expected page 2, got %d", services.Page)
	}
	if services.PageSize != 10 {
		t.Errorf("Expected page_size 10, got %d", services.PageSize)
	}
	if services.TotalCount != 15 {
		t.Errorf("Expected total_count 15, got %d", services.TotalCount)
	}
	if !services.HasMore {
		t.Error("Expected has_more to be true")
	}

	// Verify carrier service fields
	if len(services.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(services.Items))
	}
	cs := services.Items[0]
	if cs.ID != "cs_full" {
		t.Errorf("Expected ID 'cs_full', got '%s'", cs.ID)
	}
	if cs.Name != "Full Service" {
		t.Errorf("Expected name 'Full Service', got '%s'", cs.Name)
	}
	if cs.CallbackURL != "https://example.com/callback" {
		t.Errorf("Expected callback_url 'https://example.com/callback', got '%s'", cs.CallbackURL)
	}
	if !cs.Active {
		t.Error("Expected active to be true")
	}
	if !cs.ServiceDiscovery {
		t.Error("Expected service_discovery to be true")
	}
	if cs.CarrierServiceType != "api" {
		t.Errorf("Expected carrier_service_type 'api', got '%s'", cs.CarrierServiceType)
	}
	if cs.Format != "json" {
		t.Errorf("Expected format 'json', got '%s'", cs.Format)
	}
}
