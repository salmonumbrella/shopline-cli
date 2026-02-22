package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCollectionsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/collections" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := CollectionsListResponse{
			Items: []Collection{
				{ID: "col_123", Title: "Summer Collection", ProductsCount: 10},
				{ID: "col_456", Title: "Winter Collection", ProductsCount: 15},
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

	collections, err := client.ListCollections(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListCollections failed: %v", err)
	}

	if len(collections.Items) != 2 {
		t.Errorf("Expected 2 collections, got %d", len(collections.Items))
	}
	if collections.Items[0].ID != "col_123" {
		t.Errorf("Unexpected collection ID: %s", collections.Items[0].ID)
	}
}

func TestCollectionsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/collections/col_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		collection := Collection{ID: "col_123", Title: "Summer Collection", ProductsCount: 10}
		_ = json.NewEncoder(w).Encode(collection)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	collection, err := client.GetCollection(context.Background(), "col_123")
	if err != nil {
		t.Fatalf("GetCollection failed: %v", err)
	}

	if collection.ID != "col_123" {
		t.Errorf("Unexpected collection ID: %s", collection.ID)
	}
}

func TestGetCollectionEmptyID(t *testing.T) {
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
			_, err := client.GetCollection(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "collection id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCollectionsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/collections" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		collection := Collection{ID: "col_new", Title: "New Collection"}
		_ = json.NewEncoder(w).Encode(collection)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CollectionCreateRequest{Title: "New Collection"}
	collection, err := client.CreateCollection(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateCollection failed: %v", err)
	}

	if collection.ID != "col_new" {
		t.Errorf("Unexpected collection ID: %s", collection.ID)
	}
}

func TestCollectionsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/collections/col_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteCollection(context.Background(), "col_123")
	if err != nil {
		t.Fatalf("DeleteCollection failed: %v", err)
	}
}

func TestDeleteCollectionEmptyID(t *testing.T) {
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
			err := client.DeleteCollection(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "collection id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestAddProductsToCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/collections/col_123/products" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.AddProductsToCollection(context.Background(), "col_123", []string{"prod_1", "prod_2"})
	if err != nil {
		t.Fatalf("AddProductsToCollection failed: %v", err)
	}
}

func TestRemoveProductFromCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/collections/col_123/products/prod_456" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.RemoveProductFromCollection(context.Background(), "col_123", "prod_456")
	if err != nil {
		t.Fatalf("RemoveProductFromCollection failed: %v", err)
	}
}

func TestCollectionsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/collections/col_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		collection := Collection{ID: "col_123", Title: "Updated Collection"}
		_ = json.NewEncoder(w).Encode(collection)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CollectionUpdateRequest{Title: "Updated Collection"}
	collection, err := client.UpdateCollection(context.Background(), "col_123", req)
	if err != nil {
		t.Fatalf("UpdateCollection failed: %v", err)
	}

	if collection.ID != "col_123" {
		t.Errorf("Unexpected collection ID: %s", collection.ID)
	}
	if collection.Title != "Updated Collection" {
		t.Errorf("Unexpected collection title: %s", collection.Title)
	}
}

func TestUpdateCollectionEmptyID(t *testing.T) {
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
			req := &CollectionUpdateRequest{Title: "Test"}
			_, err := client.UpdateCollection(context.Background(), tc.id, req)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "collection id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestAddProductsToCollectionEmptyID(t *testing.T) {
	client := NewClient("token")

	err := client.AddProductsToCollection(context.Background(), "", []string{"prod_1"})
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
	if err != nil && err.Error() != "collection id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestRemoveProductFromCollectionEmptyIDs(t *testing.T) {
	client := NewClient("token")

	// Test empty collection ID
	err := client.RemoveProductFromCollection(context.Background(), "", "prod_456")
	if err == nil {
		t.Error("Expected error for empty collection ID, got nil")
	}
	if err != nil && err.Error() != "collection id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}

	// Test empty product ID
	err = client.RemoveProductFromCollection(context.Background(), "col_123", "")
	if err == nil {
		t.Error("Expected error for empty product ID, got nil")
	}
	if err != nil && err.Error() != "product id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestCollectionsListWithOptions(t *testing.T) {
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
		if query.Get("title") != "Summer" {
			t.Errorf("Expected title=Summer, got %s", query.Get("title"))
		}
		if query.Get("handle") != "summer-sale" {
			t.Errorf("Expected handle=summer-sale, got %s", query.Get("handle"))
		}
		if query.Get("published_scope") != "global" {
			t.Errorf("Expected published_scope=global, got %s", query.Get("published_scope"))
		}
		if query.Get("sort_by") != "title" {
			t.Errorf("Expected sort_by=title, got %s", query.Get("sort_by"))
		}
		if query.Get("sort_order") != "asc" {
			t.Errorf("Expected sort_order=asc, got %s", query.Get("sort_order"))
		}

		resp := CollectionsListResponse{
			Items:      []Collection{{ID: "col_123", Title: "Summer Collection"}},
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

	opts := &CollectionsListOptions{
		Page:           2,
		PageSize:       50,
		Title:          "Summer",
		Handle:         "summer-sale",
		PublishedScope: "global",
		SortBy:         "title",
		SortOrder:      "asc",
	}
	collections, err := client.ListCollections(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListCollections failed: %v", err)
	}

	if len(collections.Items) != 1 {
		t.Errorf("Expected 1 collection, got %d", len(collections.Items))
	}
}
