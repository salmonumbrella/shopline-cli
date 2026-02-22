package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCategoriesBulkUpdateProductSorting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/categories/cat_123/products_sorting" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		if body["op"] != "move" {
			t.Errorf("Unexpected body: %v", body)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.BulkUpdateCategoryProductSorting(context.Background(), "cat_123", map[string]any{"op": "move"})
	if err != nil {
		t.Fatalf("BulkUpdateCategoryProductSorting failed: %v", err)
	}
}

func TestCategoriesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/categories" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := CategoriesListResponse{
			Items: []Category{
				{ID: "cat_123", Title: "Electronics"},
				{ID: "cat_456", Title: "Clothing"},
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

	categories, err := client.ListCategories(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListCategories failed: %v", err)
	}

	if len(categories.Items) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(categories.Items))
	}
	if categories.Items[0].ID != "cat_123" {
		t.Errorf("Unexpected category ID: %s", categories.Items[0].ID)
	}
}

func TestCategoriesGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/categories/cat_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		category := Category{ID: "cat_123", Title: "Electronics"}
		_ = json.NewEncoder(w).Encode(category)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	category, err := client.GetCategory(context.Background(), "cat_123")
	if err != nil {
		t.Fatalf("GetCategory failed: %v", err)
	}

	if category.ID != "cat_123" {
		t.Errorf("Unexpected category ID: %s", category.ID)
	}
}

func TestGetCategoryEmptyID(t *testing.T) {
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
			_, err := client.GetCategory(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "category id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCategoriesCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/categories" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		category := Category{ID: "cat_new", Title: "New Category"}
		_ = json.NewEncoder(w).Encode(category)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CategoryCreateRequest{Title: "New Category"}
	category, err := client.CreateCategory(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateCategory failed: %v", err)
	}

	if category.ID != "cat_new" {
		t.Errorf("Unexpected category ID: %s", category.ID)
	}
}

func TestCategoriesDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/categories/cat_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteCategory(context.Background(), "cat_123")
	if err != nil {
		t.Fatalf("DeleteCategory failed: %v", err)
	}
}

func TestDeleteCategoryEmptyID(t *testing.T) {
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
			err := client.DeleteCategory(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "category id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCategoriesUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/categories/cat_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		category := Category{ID: "cat_123", Title: "Updated Electronics"}
		_ = json.NewEncoder(w).Encode(category)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CategoryUpdateRequest{Title: "Updated Electronics"}
	category, err := client.UpdateCategory(context.Background(), "cat_123", req)
	if err != nil {
		t.Fatalf("UpdateCategory failed: %v", err)
	}

	if category.ID != "cat_123" {
		t.Errorf("Unexpected category ID: %s", category.ID)
	}
	if category.Title != "Updated Electronics" {
		t.Errorf("Unexpected category title: %s", category.Title)
	}
}

func TestUpdateCategoryEmptyID(t *testing.T) {
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
			req := &CategoryUpdateRequest{Title: "Test"}
			_, err := client.UpdateCategory(context.Background(), tc.id, req)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "category id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCategoriesListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", query.Get("page"))
		}
		if query.Get("page_size") != "50" {
			t.Errorf("Expected page_size=50, got %s", query.Get("page_size"))
		}
		if query.Get("parent_id") != "parent_123" {
			t.Errorf("Expected parent_id=parent_123, got %s", query.Get("parent_id"))
		}
		if query.Get("sort_by") != "title" {
			t.Errorf("Expected sort_by=title, got %s", query.Get("sort_by"))
		}
		if query.Get("sort_order") != "asc" {
			t.Errorf("Expected sort_order=asc, got %s", query.Get("sort_order"))
		}

		resp := CategoriesListResponse{
			Items:      []Category{{ID: "cat_123", Title: "Electronics"}},
			Page:       2,
			PageSize:   50,
			TotalCount: 100,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &CategoriesListOptions{
		Page:      2,
		PageSize:  50,
		ParentID:  "parent_123",
		SortBy:    "title",
		SortOrder: "asc",
	}
	categories, err := client.ListCategories(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListCategories failed: %v", err)
	}

	if len(categories.Items) != 1 {
		t.Errorf("Expected 1 category, got %d", len(categories.Items))
	}
}
