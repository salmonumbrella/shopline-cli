package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompanyCatalogsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/company_catalogs" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := CompanyCatalogsListResponse{
			Items: []CompanyCatalog{
				{ID: "cat_123", CompanyID: "comp_1", Name: "Standard Catalog", Status: "active"},
				{ID: "cat_456", CompanyID: "comp_2", Name: "Premium Catalog", Status: "active"},
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

	catalogs, err := client.ListCompanyCatalogs(context.Background(), &CompanyCatalogsListOptions{})
	if err != nil {
		t.Fatalf("ListCompanyCatalogs failed: %v", err)
	}

	if len(catalogs.Items) != 2 {
		t.Errorf("Expected 2 catalogs, got %d", len(catalogs.Items))
	}
	if catalogs.Items[0].ID != "cat_123" {
		t.Errorf("Unexpected catalog ID: %s", catalogs.Items[0].ID)
	}
}

func TestCompanyCatalogsListWithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("company_id") != "comp_123" {
			t.Errorf("Expected company_id=comp_123, got %s", r.URL.Query().Get("company_id"))
		}
		if r.URL.Query().Get("status") != "active" {
			t.Errorf("Expected status=active, got %s", r.URL.Query().Get("status"))
		}

		resp := CompanyCatalogsListResponse{Items: []CompanyCatalog{}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.ListCompanyCatalogs(context.Background(), &CompanyCatalogsListOptions{
		CompanyID: "comp_123",
		Status:    "active",
	})
	if err != nil {
		t.Fatalf("ListCompanyCatalogs failed: %v", err)
	}
}

func TestCompanyCatalogsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/company_catalogs/cat_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		catalog := CompanyCatalog{ID: "cat_123", CompanyID: "comp_1", Name: "Standard Catalog"}
		_ = json.NewEncoder(w).Encode(catalog)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	catalog, err := client.GetCompanyCatalog(context.Background(), "cat_123")
	if err != nil {
		t.Fatalf("GetCompanyCatalog failed: %v", err)
	}

	if catalog.ID != "cat_123" {
		t.Errorf("Unexpected catalog ID: %s", catalog.ID)
	}
}

func TestGetCompanyCatalogEmptyID(t *testing.T) {
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
			_, err := client.GetCompanyCatalog(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "company catalog id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCompanyCatalogsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/company_catalogs" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		catalog := CompanyCatalog{ID: "cat_new", CompanyID: "comp_1", Name: "New Catalog"}
		_ = json.NewEncoder(w).Encode(catalog)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CompanyCatalogCreateRequest{
		CompanyID: "comp_1",
		Name:      "New Catalog",
	}

	catalog, err := client.CreateCompanyCatalog(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateCompanyCatalog failed: %v", err)
	}

	if catalog.Name != "New Catalog" {
		t.Errorf("Unexpected catalog name: %s", catalog.Name)
	}
}

func TestCompanyCatalogsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/company_catalogs/cat_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		catalog := CompanyCatalog{ID: "cat_123", Name: "Updated Catalog"}
		_ = json.NewEncoder(w).Encode(catalog)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CompanyCatalogUpdateRequest{Name: "Updated Catalog"}
	catalog, err := client.UpdateCompanyCatalog(context.Background(), "cat_123", req)
	if err != nil {
		t.Fatalf("UpdateCompanyCatalog failed: %v", err)
	}

	if catalog.Name != "Updated Catalog" {
		t.Errorf("Unexpected catalog name: %s", catalog.Name)
	}
}

func TestCompanyCatalogsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/company_catalogs/cat_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteCompanyCatalog(context.Background(), "cat_123")
	if err != nil {
		t.Fatalf("DeleteCompanyCatalog failed: %v", err)
	}
}

func TestDeleteCompanyCatalogEmptyID(t *testing.T) {
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
			err := client.DeleteCompanyCatalog(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "company catalog id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
