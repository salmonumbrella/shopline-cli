package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBlogsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/blogs" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := BlogsListResponse{
			Items: []Blog{
				{ID: "blog_123", Title: "News", Handle: "news", Commentable: "moderate"},
				{ID: "blog_456", Title: "Updates", Handle: "updates", Commentable: "yes"},
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

	blogs, err := client.ListBlogs(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListBlogs failed: %v", err)
	}

	if len(blogs.Items) != 2 {
		t.Errorf("Expected 2 blogs, got %d", len(blogs.Items))
	}
	if blogs.Items[0].ID != "blog_123" {
		t.Errorf("Unexpected blog ID: %s", blogs.Items[0].ID)
	}
}

func TestBlogsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/blogs/blog_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		blog := Blog{ID: "blog_123", Title: "News", Handle: "news", Commentable: "moderate"}
		_ = json.NewEncoder(w).Encode(blog)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	blog, err := client.GetBlog(context.Background(), "blog_123")
	if err != nil {
		t.Fatalf("GetBlog failed: %v", err)
	}

	if blog.ID != "blog_123" {
		t.Errorf("Unexpected blog ID: %s", blog.ID)
	}
}

func TestGetBlogEmptyID(t *testing.T) {
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
			_, err := client.GetBlog(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "blog id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestBlogsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/blogs" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req BlogCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Title != "My Blog" {
			t.Errorf("Unexpected title: %s", req.Title)
		}

		blog := Blog{
			ID:          "blog_new",
			Title:       req.Title,
			Handle:      "my-blog",
			Commentable: "moderate",
		}
		_ = json.NewEncoder(w).Encode(blog)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &BlogCreateRequest{
		Title: "My Blog",
	}
	blog, err := client.CreateBlog(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateBlog failed: %v", err)
	}

	if blog.ID != "blog_new" {
		t.Errorf("Unexpected blog ID: %s", blog.ID)
	}
	if blog.Title != "My Blog" {
		t.Errorf("Unexpected title: %s", blog.Title)
	}
}

func TestBlogsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/blogs/blog_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req BlogUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		blog := Blog{
			ID:          "blog_123",
			Title:       req.Title,
			Handle:      "updated-blog",
			Commentable: "yes",
		}
		_ = json.NewEncoder(w).Encode(blog)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &BlogUpdateRequest{
		Title: "Updated Blog",
	}
	blog, err := client.UpdateBlog(context.Background(), "blog_123", req)
	if err != nil {
		t.Fatalf("UpdateBlog failed: %v", err)
	}

	if blog.Title != "Updated Blog" {
		t.Errorf("Unexpected blog title: %s", blog.Title)
	}
}

func TestBlogsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/blogs/blog_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteBlog(context.Background(), "blog_123")
	if err != nil {
		t.Fatalf("DeleteBlog failed: %v", err)
	}
}

func TestUpdateBlogEmptyID(t *testing.T) {
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
			_, err := client.UpdateBlog(context.Background(), tc.id, &BlogUpdateRequest{Title: "Test"})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "blog id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteBlogEmptyID(t *testing.T) {
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
			err := client.DeleteBlog(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "blog id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
