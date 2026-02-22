package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTaxesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/taxes" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := TaxesListResponse{
			Items: []Tax{
				{ID: "tax_123", Name: "VAT", Rate: 20.0, CountryCode: "GB", Enabled: true},
				{ID: "tax_456", Name: "GST", Rate: 10.0, CountryCode: "AU", Enabled: true},
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

	taxes, err := client.ListTaxes(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTaxes failed: %v", err)
	}

	if len(taxes.Items) != 2 {
		t.Errorf("Expected 2 taxes, got %d", len(taxes.Items))
	}
	if taxes.Items[0].ID != "tax_123" {
		t.Errorf("Unexpected tax ID: %s", taxes.Items[0].ID)
	}
	if taxes.Items[0].Rate != 20.0 {
		t.Errorf("Unexpected tax rate: %f", taxes.Items[0].Rate)
	}
}

func TestTaxesListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("country_code") != "US" {
			t.Errorf("Expected country_code=US, got %s", r.URL.Query().Get("country_code"))
		}
		if r.URL.Query().Get("enabled") != "true" {
			t.Errorf("Expected enabled=true, got %s", r.URL.Query().Get("enabled"))
		}

		resp := TaxesListResponse{
			Items:      []Tax{{ID: "tax_789", Name: "Sales Tax", CountryCode: "US"}},
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

	enabled := true
	opts := &TaxesListOptions{
		CountryCode: "US",
		Enabled:     &enabled,
	}
	taxes, err := client.ListTaxes(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListTaxes failed: %v", err)
	}

	if len(taxes.Items) != 1 {
		t.Errorf("Expected 1 tax, got %d", len(taxes.Items))
	}
}

func TestTaxesGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/taxes/tax_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		tax := Tax{
			ID:          "tax_123",
			Name:        "VAT",
			Rate:        20.0,
			CountryCode: "GB",
			Shipping:    true,
			Enabled:     true,
		}
		_ = json.NewEncoder(w).Encode(tax)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	tax, err := client.GetTax(context.Background(), "tax_123")
	if err != nil {
		t.Fatalf("GetTax failed: %v", err)
	}

	if tax.ID != "tax_123" {
		t.Errorf("Unexpected tax ID: %s", tax.ID)
	}
	if !tax.Shipping {
		t.Error("Expected Shipping to be true")
	}
}

func TestGetTaxEmptyID(t *testing.T) {
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
			_, err := client.GetTax(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "tax id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestTaxesCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		var req TaxCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Name != "State Tax" {
			t.Errorf("Unexpected name: %s", req.Name)
		}
		if req.Rate != 8.25 {
			t.Errorf("Unexpected rate: %f", req.Rate)
		}

		tax := Tax{ID: "tax_new", Name: req.Name, Rate: req.Rate, CountryCode: req.CountryCode}
		_ = json.NewEncoder(w).Encode(tax)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &TaxCreateRequest{
		Name:        "State Tax",
		Rate:        8.25,
		CountryCode: "US",
	}

	tax, err := client.CreateTax(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateTax failed: %v", err)
	}

	if tax.ID != "tax_new" {
		t.Errorf("Unexpected tax ID: %s", tax.ID)
	}
}

func TestTaxesUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/taxes/tax_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req TaxUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		tax := Tax{ID: "tax_123", Name: req.Name, Rate: 21.0}
		_ = json.NewEncoder(w).Encode(tax)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	rate := 21.0
	req := &TaxUpdateRequest{
		Name: "Updated VAT",
		Rate: &rate,
	}

	tax, err := client.UpdateTax(context.Background(), "tax_123", req)
	if err != nil {
		t.Fatalf("UpdateTax failed: %v", err)
	}

	if tax.Name != "Updated VAT" {
		t.Errorf("Unexpected name: %s", tax.Name)
	}
}

func TestTaxesDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/taxes/tax_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteTax(context.Background(), "tax_123")
	if err != nil {
		t.Fatalf("DeleteTax failed: %v", err)
	}
}

func TestDeleteTaxEmptyID(t *testing.T) {
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
			err := client.DeleteTax(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "tax id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
