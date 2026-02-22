package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCatalogPricingList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/catalog_pricing" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := CatalogPricingListResponse{
			Items: []CatalogPricing{
				{ID: "cp_123", CatalogID: "cat_1", ProductID: "prod_1", CatalogPrice: 99.99},
				{ID: "cp_456", CatalogID: "cat_1", ProductID: "prod_2", CatalogPrice: 149.99},
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

	pricing, err := client.ListCatalogPricing(context.Background(), &CatalogPricingListOptions{})
	if err != nil {
		t.Fatalf("ListCatalogPricing failed: %v", err)
	}

	if len(pricing.Items) != 2 {
		t.Errorf("Expected 2 pricing entries, got %d", len(pricing.Items))
	}
	if pricing.Items[0].ID != "cp_123" {
		t.Errorf("Unexpected pricing ID: %s", pricing.Items[0].ID)
	}
}

func TestCatalogPricingListWithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("catalog_id") != "cat_123" {
			t.Errorf("Expected catalog_id=cat_123, got %s", r.URL.Query().Get("catalog_id"))
		}
		if r.URL.Query().Get("product_id") != "prod_456" {
			t.Errorf("Expected product_id=prod_456, got %s", r.URL.Query().Get("product_id"))
		}

		resp := CatalogPricingListResponse{Items: []CatalogPricing{}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.ListCatalogPricing(context.Background(), &CatalogPricingListOptions{
		CatalogID: "cat_123",
		ProductID: "prod_456",
	})
	if err != nil {
		t.Fatalf("ListCatalogPricing failed: %v", err)
	}
}

func TestCatalogPricingGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/catalog_pricing/cp_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		pricing := CatalogPricing{ID: "cp_123", CatalogID: "cat_1", ProductID: "prod_1", CatalogPrice: 99.99}
		_ = json.NewEncoder(w).Encode(pricing)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	pricing, err := client.GetCatalogPricing(context.Background(), "cp_123")
	if err != nil {
		t.Fatalf("GetCatalogPricing failed: %v", err)
	}

	if pricing.ID != "cp_123" {
		t.Errorf("Unexpected pricing ID: %s", pricing.ID)
	}
}

func TestGetCatalogPricingEmptyID(t *testing.T) {
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
			_, err := client.GetCatalogPricing(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "catalog pricing id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCatalogPricingCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/catalog_pricing" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		pricing := CatalogPricing{ID: "cp_new", CatalogID: "cat_1", ProductID: "prod_1", CatalogPrice: 79.99}
		_ = json.NewEncoder(w).Encode(pricing)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CatalogPricingCreateRequest{
		CatalogID:    "cat_1",
		ProductID:    "prod_1",
		CatalogPrice: 79.99,
	}

	pricing, err := client.CreateCatalogPricing(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateCatalogPricing failed: %v", err)
	}

	if pricing.CatalogPrice != 79.99 {
		t.Errorf("Unexpected catalog price: %.2f", pricing.CatalogPrice)
	}
}

func TestCatalogPricingUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/catalog_pricing/cp_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		pricing := CatalogPricing{ID: "cp_123", CatalogPrice: 89.99}
		_ = json.NewEncoder(w).Encode(pricing)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	newPrice := 89.99
	req := &CatalogPricingUpdateRequest{CatalogPrice: &newPrice}
	pricing, err := client.UpdateCatalogPricing(context.Background(), "cp_123", req)
	if err != nil {
		t.Fatalf("UpdateCatalogPricing failed: %v", err)
	}

	if pricing.CatalogPrice != 89.99 {
		t.Errorf("Unexpected catalog price: %.2f", pricing.CatalogPrice)
	}
}

func TestCatalogPricingDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/catalog_pricing/cp_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteCatalogPricing(context.Background(), "cp_123")
	if err != nil {
		t.Fatalf("DeleteCatalogPricing failed: %v", err)
	}
}

func TestDeleteCatalogPricingEmptyID(t *testing.T) {
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
			err := client.DeleteCatalogPricing(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "catalog pricing id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
