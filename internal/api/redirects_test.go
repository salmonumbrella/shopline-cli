package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRedirectsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/redirects" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := RedirectsListResponse{
			Items: []Redirect{
				{ID: "redir_123", Path: "/old-page", Target: "/new-page"},
				{ID: "redir_456", Path: "/legacy", Target: "https://example.com/new"},
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

	redirects, err := client.ListRedirects(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListRedirects failed: %v", err)
	}

	if len(redirects.Items) != 2 {
		t.Errorf("Expected 2 redirects, got %d", len(redirects.Items))
	}
	if redirects.Items[0].ID != "redir_123" {
		t.Errorf("Unexpected redirect ID: %s", redirects.Items[0].ID)
	}
}

func TestRedirectsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("path") != "/old" {
			t.Errorf("Expected path=/old, got %s", r.URL.Query().Get("path"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}

		resp := RedirectsListResponse{
			Items:      []Redirect{},
			Page:       2,
			PageSize:   20,
			TotalCount: 0,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &RedirectsListOptions{
		Page: 2,
		Path: "/old",
	}
	_, err := client.ListRedirects(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListRedirects failed: %v", err)
	}
}

func TestRedirectsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/redirects/redir_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		redirect := Redirect{ID: "redir_123", Path: "/old-page", Target: "/new-page"}
		_ = json.NewEncoder(w).Encode(redirect)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	redirect, err := client.GetRedirect(context.Background(), "redir_123")
	if err != nil {
		t.Fatalf("GetRedirect failed: %v", err)
	}

	if redirect.ID != "redir_123" {
		t.Errorf("Unexpected redirect ID: %s", redirect.ID)
	}
}

func TestGetRedirectEmptyID(t *testing.T) {
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
			_, err := client.GetRedirect(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "redirect id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestRedirectsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/redirects" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req RedirectCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Path != "/old-url" {
			t.Errorf("Unexpected path: %s", req.Path)
		}

		redirect := Redirect{
			ID:     "redir_new",
			Path:   req.Path,
			Target: req.Target,
		}
		_ = json.NewEncoder(w).Encode(redirect)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &RedirectCreateRequest{
		Path:   "/old-url",
		Target: "/new-url",
	}
	redirect, err := client.CreateRedirect(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateRedirect failed: %v", err)
	}

	if redirect.ID != "redir_new" {
		t.Errorf("Unexpected redirect ID: %s", redirect.ID)
	}
	if redirect.Path != "/old-url" {
		t.Errorf("Unexpected path: %s", redirect.Path)
	}
}

func TestRedirectsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/redirects/redir_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req RedirectUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		redirect := Redirect{
			ID:     "redir_123",
			Path:   "/old-url",
			Target: req.Target,
		}
		_ = json.NewEncoder(w).Encode(redirect)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &RedirectUpdateRequest{
		Target: "/updated-target",
	}
	redirect, err := client.UpdateRedirect(context.Background(), "redir_123", req)
	if err != nil {
		t.Fatalf("UpdateRedirect failed: %v", err)
	}

	if redirect.Target != "/updated-target" {
		t.Errorf("Unexpected target: %s", redirect.Target)
	}
}

func TestRedirectsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/redirects/redir_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteRedirect(context.Background(), "redir_123")
	if err != nil {
		t.Fatalf("DeleteRedirect failed: %v", err)
	}
}

func TestUpdateRedirectEmptyID(t *testing.T) {
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
			_, err := client.UpdateRedirect(context.Background(), tc.id, &RedirectUpdateRequest{Target: "/test"})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "redirect id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteRedirectEmptyID(t *testing.T) {
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
			err := client.DeleteRedirect(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "redirect id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
