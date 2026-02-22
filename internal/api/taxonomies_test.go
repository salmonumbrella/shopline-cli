package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTaxonomiesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/taxonomies" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := TaxonomiesListResponse{
			Items: []Taxonomy{
				{
					ID:           "tax_123",
					Name:         "Electronics",
					Handle:       "electronics",
					Level:        0,
					ProductCount: 150,
					Active:       true,
				},
				{
					ID:           "tax_456",
					Name:         "Clothing",
					Handle:       "clothing",
					Level:        0,
					ProductCount: 200,
					Active:       true,
				},
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

	taxonomies, err := client.ListTaxonomies(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTaxonomies failed: %v", err)
	}

	if len(taxonomies.Items) != 2 {
		t.Errorf("Expected 2 taxonomies, got %d", len(taxonomies.Items))
	}
	if taxonomies.Items[0].ID != "tax_123" {
		t.Errorf("Unexpected taxonomy ID: %s", taxonomies.Items[0].ID)
	}
	if taxonomies.Items[0].Name != "Electronics" {
		t.Errorf("Unexpected name: %s", taxonomies.Items[0].Name)
	}
}

func TestTaxonomiesListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("parent_id") != "tax_123" {
			t.Errorf("Expected parent_id=tax_123, got %s", r.URL.Query().Get("parent_id"))
		}
		if r.URL.Query().Get("active") != "true" {
			t.Errorf("Expected active=true, got %s", r.URL.Query().Get("active"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}

		resp := TaxonomiesListResponse{
			Items: []Taxonomy{
				{ID: "tax_789", Name: "Smartphones", ParentID: "tax_123", Active: true},
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

	active := true
	opts := &TaxonomiesListOptions{
		Page:     2,
		ParentID: "tax_123",
		Active:   &active,
	}
	taxonomies, err := client.ListTaxonomies(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListTaxonomies failed: %v", err)
	}

	if len(taxonomies.Items) != 1 {
		t.Errorf("Expected 1 taxonomy, got %d", len(taxonomies.Items))
	}
}

func TestTaxonomiesGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/taxonomies/tax_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		taxonomy := Taxonomy{
			ID:           "tax_123",
			Name:         "Electronics",
			Handle:       "electronics",
			Description:  "Electronic devices and accessories",
			Level:        0,
			Position:     1,
			Path:         "electronics",
			FullPath:     "Electronics",
			ProductCount: 150,
			Active:       true,
		}
		_ = json.NewEncoder(w).Encode(taxonomy)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	taxonomy, err := client.GetTaxonomy(context.Background(), "tax_123")
	if err != nil {
		t.Fatalf("GetTaxonomy failed: %v", err)
	}

	if taxonomy.ID != "tax_123" {
		t.Errorf("Unexpected taxonomy ID: %s", taxonomy.ID)
	}
	if taxonomy.Name != "Electronics" {
		t.Errorf("Unexpected name: %s", taxonomy.Name)
	}
	if taxonomy.ProductCount != 150 {
		t.Errorf("Unexpected product count: %d", taxonomy.ProductCount)
	}
}

func TestGetTaxonomyEmptyID(t *testing.T) {
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
			_, err := client.GetTaxonomy(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "taxonomy id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestTaxonomiesCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/taxonomies" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req TaxonomyCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Name != "Smartphones" {
			t.Errorf("Unexpected name: %s", req.Name)
		}
		if req.ParentID != "tax_123" {
			t.Errorf("Unexpected parent_id: %s", req.ParentID)
		}

		taxonomy := Taxonomy{
			ID:       "tax_new",
			Name:     req.Name,
			Handle:   req.Handle,
			ParentID: req.ParentID,
			Level:    1,
			Active:   req.Active,
		}
		_ = json.NewEncoder(w).Encode(taxonomy)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &TaxonomyCreateRequest{
		Name:     "Smartphones",
		Handle:   "smartphones",
		ParentID: "tax_123",
		Active:   true,
	}
	taxonomy, err := client.CreateTaxonomy(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateTaxonomy failed: %v", err)
	}

	if taxonomy.ID != "tax_new" {
		t.Errorf("Unexpected taxonomy ID: %s", taxonomy.ID)
	}
	if taxonomy.Name != "Smartphones" {
		t.Errorf("Unexpected name: %s", taxonomy.Name)
	}
}

func TestTaxonomiesUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/taxonomies/tax_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req TaxonomyUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Name != "Consumer Electronics" {
			t.Errorf("Unexpected name: %s", req.Name)
		}
		if req.Active == nil || *req.Active != false {
			t.Errorf("Unexpected active value")
		}

		taxonomy := Taxonomy{
			ID:     "tax_123",
			Name:   req.Name,
			Handle: "consumer-electronics",
			Active: *req.Active,
		}
		_ = json.NewEncoder(w).Encode(taxonomy)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	active := false
	req := &TaxonomyUpdateRequest{
		Name:   "Consumer Electronics",
		Active: &active,
	}
	taxonomy, err := client.UpdateTaxonomy(context.Background(), "tax_123", req)
	if err != nil {
		t.Fatalf("UpdateTaxonomy failed: %v", err)
	}

	if taxonomy.ID != "tax_123" {
		t.Errorf("Unexpected taxonomy ID: %s", taxonomy.ID)
	}
	if taxonomy.Name != "Consumer Electronics" {
		t.Errorf("Unexpected name: %s", taxonomy.Name)
	}
}

func TestUpdateTaxonomyEmptyID(t *testing.T) {
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
			_, err := client.UpdateTaxonomy(context.Background(), tc.id, &TaxonomyUpdateRequest{Name: "Test"})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "taxonomy id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestTaxonomiesDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/taxonomies/tax_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteTaxonomy(context.Background(), "tax_123")
	if err != nil {
		t.Fatalf("DeleteTaxonomy failed: %v", err)
	}
}

func TestDeleteTaxonomyEmptyID(t *testing.T) {
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
			err := client.DeleteTaxonomy(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "taxonomy id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
