package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPagesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/pages" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := PagesListResponse{
			Items: []Page{
				{ID: "pg_123", Title: "About Us", Handle: "about-us", Published: true},
				{ID: "pg_456", Title: "Contact", Handle: "contact", Published: true},
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

	pages, err := client.ListPages(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListPages failed: %v", err)
	}

	if len(pages.Items) != 2 {
		t.Errorf("Expected 2 pages, got %d", len(pages.Items))
	}
	if pages.Items[0].ID != "pg_123" {
		t.Errorf("Unexpected page ID: %s", pages.Items[0].ID)
	}
}

func TestPagesGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pages/pg_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		page := Page{ID: "pg_123", Title: "About Us", Handle: "about-us"}
		_ = json.NewEncoder(w).Encode(page)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	page, err := client.GetPage(context.Background(), "pg_123")
	if err != nil {
		t.Fatalf("GetPage failed: %v", err)
	}

	if page.ID != "pg_123" {
		t.Errorf("Unexpected page ID: %s", page.ID)
	}
}

func TestGetPageEmptyID(t *testing.T) {
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
			_, err := client.GetPage(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "page id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestPagesCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		page := Page{ID: "pg_new", Title: "New Page"}
		_ = json.NewEncoder(w).Encode(page)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &PageCreateRequest{
		Title:    "New Page",
		BodyHTML: "<p>Page content</p>",
	}

	page, err := client.CreatePage(context.Background(), req)
	if err != nil {
		t.Fatalf("CreatePage failed: %v", err)
	}

	if page.ID != "pg_new" {
		t.Errorf("Unexpected page ID: %s", page.ID)
	}
}

func TestPagesUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}

		page := Page{ID: "pg_123", Title: "Updated Page"}
		_ = json.NewEncoder(w).Encode(page)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &PageUpdateRequest{
		Title: "Updated Page",
	}

	page, err := client.UpdatePage(context.Background(), "pg_123", req)
	if err != nil {
		t.Fatalf("UpdatePage failed: %v", err)
	}

	if page.Title != "Updated Page" {
		t.Errorf("Unexpected page title: %s", page.Title)
	}
}

func TestPagesDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/pages/pg_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeletePage(context.Background(), "pg_123")
	if err != nil {
		t.Fatalf("DeletePage failed: %v", err)
	}
}
