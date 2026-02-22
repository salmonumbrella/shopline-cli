package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCustomerSavedSearchesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/customer_saved_searches" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := CustomerSavedSearchesListResponse{
			Items: []CustomerSavedSearch{
				{ID: "ss_123", Name: "VIP Customers", Query: "total_spent:>1000"},
				{ID: "ss_456", Name: "New Customers", Query: "created_at:>2024-01-01"},
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

	searches, err := client.ListCustomerSavedSearches(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListCustomerSavedSearches failed: %v", err)
	}

	if len(searches.Items) != 2 {
		t.Errorf("Expected 2 searches, got %d", len(searches.Items))
	}
	if searches.Items[0].ID != "ss_123" {
		t.Errorf("Unexpected search ID: %s", searches.Items[0].ID)
	}
	if searches.Items[0].Name != "VIP Customers" {
		t.Errorf("Unexpected name: %s", searches.Items[0].Name)
	}
}

func TestCustomerSavedSearchesListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("name") != "VIP" {
			t.Errorf("Expected name=VIP, got %s", r.URL.Query().Get("name"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}

		resp := CustomerSavedSearchesListResponse{
			Items: []CustomerSavedSearch{
				{ID: "ss_123", Name: "VIP Customers"},
			},
			Page:       2,
			PageSize:   20,
			TotalCount: 1,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &CustomerSavedSearchesListOptions{
		Page: 2,
		Name: "VIP",
	}
	searches, err := client.ListCustomerSavedSearches(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListCustomerSavedSearches failed: %v", err)
	}

	if len(searches.Items) != 1 {
		t.Errorf("Expected 1 search, got %d", len(searches.Items))
	}
}

func TestCustomerSavedSearchesGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer_saved_searches/ss_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		search := CustomerSavedSearch{
			ID:    "ss_123",
			Name:  "VIP Customers",
			Query: "total_spent:>1000",
		}
		_ = json.NewEncoder(w).Encode(search)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	search, err := client.GetCustomerSavedSearch(context.Background(), "ss_123")
	if err != nil {
		t.Fatalf("GetCustomerSavedSearch failed: %v", err)
	}

	if search.ID != "ss_123" {
		t.Errorf("Unexpected search ID: %s", search.ID)
	}
	if search.Query != "total_spent:>1000" {
		t.Errorf("Unexpected query: %s", search.Query)
	}
}

func TestCustomerSavedSearchesCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/customer_saved_searches" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req CustomerSavedSearchCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Name != "High Value" {
			t.Errorf("Unexpected name: %s", req.Name)
		}
		if req.Query != "total_spent:>5000" {
			t.Errorf("Unexpected query: %s", req.Query)
		}

		search := CustomerSavedSearch{
			ID:    "ss_new",
			Name:  req.Name,
			Query: req.Query,
		}
		_ = json.NewEncoder(w).Encode(search)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CustomerSavedSearchCreateRequest{
		Name:  "High Value",
		Query: "total_spent:>5000",
	}
	search, err := client.CreateCustomerSavedSearch(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateCustomerSavedSearch failed: %v", err)
	}

	if search.ID != "ss_new" {
		t.Errorf("Unexpected search ID: %s", search.ID)
	}
}

func TestCustomerSavedSearchesDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/customer_saved_searches/ss_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteCustomerSavedSearch(context.Background(), "ss_123")
	if err != nil {
		t.Fatalf("DeleteCustomerSavedSearch failed: %v", err)
	}
}

func TestGetCustomerSavedSearchEmptyID(t *testing.T) {
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
			_, err := client.GetCustomerSavedSearch(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "saved search id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteCustomerSavedSearchEmptyID(t *testing.T) {
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
			err := client.DeleteCustomerSavedSearch(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "saved search id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
