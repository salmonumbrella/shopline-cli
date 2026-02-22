package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTaxServicesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/tax_services" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := TaxServicesListResponse{
			Items: []TaxService{
				{ID: "ts_123", Name: "Avalara", Provider: "avalara", Active: true},
				{ID: "ts_456", Name: "TaxJar", Provider: "taxjar", Active: true},
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

	services, err := client.ListTaxServices(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTaxServices failed: %v", err)
	}

	if len(services.Items) != 2 {
		t.Errorf("Expected 2 tax services, got %d", len(services.Items))
	}
	if services.Items[0].ID != "ts_123" {
		t.Errorf("Unexpected tax service ID: %s", services.Items[0].ID)
	}
}

func TestTaxServicesListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("provider") != "avalara" {
			t.Errorf("Expected provider=avalara, got %s", r.URL.Query().Get("provider"))
		}
		if r.URL.Query().Get("active") != "true" {
			t.Errorf("Expected active=true, got %s", r.URL.Query().Get("active"))
		}

		resp := TaxServicesListResponse{
			Items:      []TaxService{{ID: "ts_123", Provider: "avalara"}},
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

	active := true
	opts := &TaxServicesListOptions{
		Provider: "avalara",
		Active:   &active,
	}
	services, err := client.ListTaxServices(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListTaxServices failed: %v", err)
	}

	if len(services.Items) != 1 {
		t.Errorf("Expected 1 tax service, got %d", len(services.Items))
	}
}

func TestTaxServicesGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tax_services/ts_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		service := TaxService{
			ID:        "ts_123",
			Name:      "Avalara",
			Provider:  "avalara",
			Sandbox:   true,
			Active:    true,
			Countries: []string{"US", "CA"},
		}
		_ = json.NewEncoder(w).Encode(service)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	service, err := client.GetTaxService(context.Background(), "ts_123")
	if err != nil {
		t.Fatalf("GetTaxService failed: %v", err)
	}

	if service.ID != "ts_123" {
		t.Errorf("Unexpected tax service ID: %s", service.ID)
	}
	if !service.Sandbox {
		t.Error("Expected Sandbox to be true")
	}
	if len(service.Countries) != 2 {
		t.Errorf("Expected 2 countries, got %d", len(service.Countries))
	}
}

func TestGetTaxServiceEmptyID(t *testing.T) {
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
			_, err := client.GetTaxService(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "tax service id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestTaxServicesCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		var req TaxServiceCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Provider != "avalara" {
			t.Errorf("Unexpected provider: %s", req.Provider)
		}
		if req.APIKey != "test_key" {
			t.Errorf("Unexpected API key: %s", req.APIKey)
		}

		service := TaxService{
			ID:       "ts_new",
			Name:     req.Name,
			Provider: req.Provider,
			Active:   true,
		}
		_ = json.NewEncoder(w).Encode(service)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &TaxServiceCreateRequest{
		Name:     "My Avalara",
		Provider: "avalara",
		APIKey:   "test_key",
		Active:   true,
	}

	service, err := client.CreateTaxService(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateTaxService failed: %v", err)
	}

	if service.ID != "ts_new" {
		t.Errorf("Unexpected tax service ID: %s", service.ID)
	}
}

func TestTaxServicesUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/tax_services/ts_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req TaxServiceUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		service := TaxService{ID: "ts_123", Name: req.Name, Active: true}
		_ = json.NewEncoder(w).Encode(service)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	active := true
	req := &TaxServiceUpdateRequest{
		Name:   "Updated Avalara",
		Active: &active,
	}

	service, err := client.UpdateTaxService(context.Background(), "ts_123", req)
	if err != nil {
		t.Fatalf("UpdateTaxService failed: %v", err)
	}

	if service.Name != "Updated Avalara" {
		t.Errorf("Unexpected name: %s", service.Name)
	}
}

func TestTaxServicesDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/tax_services/ts_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteTaxService(context.Background(), "ts_123")
	if err != nil {
		t.Fatalf("DeleteTaxService failed: %v", err)
	}
}

func TestDeleteTaxServiceEmptyID(t *testing.T) {
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
			err := client.DeleteTaxService(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "tax service id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
