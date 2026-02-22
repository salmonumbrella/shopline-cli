package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTagsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/tags" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := TagsListResponse{
			Items: []Tag{
				{
					ID:           "tag_123",
					Name:         "Sale",
					Handle:       "sale",
					ProductCount: 42,
				},
				{
					ID:           "tag_456",
					Name:         "New Arrival",
					Handle:       "new-arrival",
					ProductCount: 15,
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

	tags, err := client.ListTags(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}

	if len(tags.Items) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tags.Items))
	}
	if tags.Items[0].ID != "tag_123" {
		t.Errorf("Unexpected tag ID: %s", tags.Items[0].ID)
	}
	if tags.Items[0].Name != "Sale" {
		t.Errorf("Unexpected name: %s", tags.Items[0].Name)
	}
}

func TestTagsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("query") != "sale" {
			t.Errorf("Expected query=sale, got %s", r.URL.Query().Get("query"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}

		resp := TagsListResponse{
			Items: []Tag{
				{ID: "tag_123", Name: "Sale"},
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

	opts := &TagsListOptions{
		Page:  2,
		Query: "sale",
	}
	tags, err := client.ListTags(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}

	if len(tags.Items) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(tags.Items))
	}
}

func TestTagsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tags/tag_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		tag := Tag{
			ID:           "tag_123",
			Name:         "Sale",
			Handle:       "sale",
			ProductCount: 42,
		}
		_ = json.NewEncoder(w).Encode(tag)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	tag, err := client.GetTag(context.Background(), "tag_123")
	if err != nil {
		t.Fatalf("GetTag failed: %v", err)
	}

	if tag.ID != "tag_123" {
		t.Errorf("Unexpected tag ID: %s", tag.ID)
	}
	if tag.Name != "Sale" {
		t.Errorf("Unexpected name: %s", tag.Name)
	}
	if tag.ProductCount != 42 {
		t.Errorf("Unexpected product count: %d", tag.ProductCount)
	}
}

func TestTagsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/tags" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req TagCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Name != "Featured" {
			t.Errorf("Unexpected name: %s", req.Name)
		}

		tag := Tag{
			ID:           "tag_new",
			Name:         req.Name,
			Handle:       "featured",
			ProductCount: 0,
		}
		_ = json.NewEncoder(w).Encode(tag)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &TagCreateRequest{
		Name: "Featured",
	}
	tag, err := client.CreateTag(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateTag failed: %v", err)
	}

	if tag.ID != "tag_new" {
		t.Errorf("Unexpected tag ID: %s", tag.ID)
	}
	if tag.Name != "Featured" {
		t.Errorf("Unexpected name: %s", tag.Name)
	}
}

func TestTagsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/tags/tag_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteTag(context.Background(), "tag_123")
	if err != nil {
		t.Fatalf("DeleteTag failed: %v", err)
	}
}

func TestGetTagEmptyID(t *testing.T) {
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
			_, err := client.GetTag(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "tag id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteTagEmptyID(t *testing.T) {
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
			err := client.DeleteTag(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "tag id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
