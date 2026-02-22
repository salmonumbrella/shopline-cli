package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestScriptTagsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/script_tags" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := ScriptTagsListResponse{
			Items: []ScriptTag{
				{ID: "st_123", Src: "https://example.com/script.js", Event: "onload"},
				{ID: "st_456", Src: "https://example.com/analytics.js", Event: "onload"},
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

	tags, err := client.ListScriptTags(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListScriptTags failed: %v", err)
	}

	if len(tags.Items) != 2 {
		t.Errorf("Expected 2 script tags, got %d", len(tags.Items))
	}
	if tags.Items[0].ID != "st_123" {
		t.Errorf("Unexpected script tag ID: %s", tags.Items[0].ID)
	}
}

func TestScriptTagsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/script_tags/st_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		tag := ScriptTag{ID: "st_123", Src: "https://example.com/script.js"}
		_ = json.NewEncoder(w).Encode(tag)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	tag, err := client.GetScriptTag(context.Background(), "st_123")
	if err != nil {
		t.Fatalf("GetScriptTag failed: %v", err)
	}

	if tag.ID != "st_123" {
		t.Errorf("Unexpected script tag ID: %s", tag.ID)
	}
}

func TestGetScriptTagEmptyID(t *testing.T) {
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
			_, err := client.GetScriptTag(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "script tag id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestScriptTagsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		tag := ScriptTag{ID: "st_new", Src: "https://example.com/new.js"}
		_ = json.NewEncoder(w).Encode(tag)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ScriptTagCreateRequest{
		Src:   "https://example.com/new.js",
		Event: "onload",
	}

	tag, err := client.CreateScriptTag(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateScriptTag failed: %v", err)
	}

	if tag.ID != "st_new" {
		t.Errorf("Unexpected script tag ID: %s", tag.ID)
	}
}

func TestScriptTagsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/script_tags/st_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteScriptTag(context.Background(), "st_123")
	if err != nil {
		t.Fatalf("DeleteScriptTag failed: %v", err)
	}
}

func TestUpdateScriptTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/script_tags/st_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ScriptTagUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.Src != "https://example.com/updated.js" {
			t.Errorf("Unexpected src: %s", req.Src)
		}

		tag := ScriptTag{
			ID:    "st_123",
			Src:   req.Src,
			Event: "onload",
		}
		_ = json.NewEncoder(w).Encode(tag)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ScriptTagUpdateRequest{
		Src: "https://example.com/updated.js",
	}

	tag, err := client.UpdateScriptTag(context.Background(), "st_123", req)
	if err != nil {
		t.Fatalf("UpdateScriptTag failed: %v", err)
	}

	if tag.ID != "st_123" {
		t.Errorf("Unexpected script tag ID: %s", tag.ID)
	}
	if tag.Src != "https://example.com/updated.js" {
		t.Errorf("Unexpected script tag Src: %s", tag.Src)
	}
}

func TestUpdateScriptTagEmptyID(t *testing.T) {
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
			req := &ScriptTagUpdateRequest{Src: "https://example.com/test.js"}
			_, err := client.UpdateScriptTag(context.Background(), tc.id, req)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "script tag id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteScriptTagEmptyID(t *testing.T) {
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
			err := client.DeleteScriptTag(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "script tag id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestListScriptTagsWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		query := r.URL.Query()
		if query.Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", query.Get("page"))
		}
		if query.Get("page_size") != "10" {
			t.Errorf("Expected page_size=10, got %s", query.Get("page_size"))
		}
		if query.Get("src") != "https://example.com" {
			t.Errorf("Expected src=https://example.com, got %s", query.Get("src"))
		}

		resp := ScriptTagsListResponse{
			Items: []ScriptTag{
				{ID: "st_789", Src: "https://example.com/filtered.js", Event: "onload"},
			},
			Page:       2,
			PageSize:   10,
			TotalCount: 11,
			HasMore:    false,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &ScriptTagsListOptions{
		Page:     2,
		PageSize: 10,
		Src:      "https://example.com",
	}

	tags, err := client.ListScriptTags(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListScriptTags with options failed: %v", err)
	}

	if len(tags.Items) != 1 {
		t.Errorf("Expected 1 script tag, got %d", len(tags.Items))
	}
	if tags.Page != 2 {
		t.Errorf("Expected page 2, got %d", tags.Page)
	}
	if tags.PageSize != 10 {
		t.Errorf("Expected page_size 10, got %d", tags.PageSize)
	}
}

func TestListScriptTagsWithSrcOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		// Verify page and page_size are not set when they are 0
		if query.Get("page") != "" {
			t.Errorf("Expected page to be empty, got %s", query.Get("page"))
		}
		if query.Get("page_size") != "" {
			t.Errorf("Expected page_size to be empty, got %s", query.Get("page_size"))
		}
		if query.Get("src") != "https://cdn.example.com" {
			t.Errorf("Expected src=https://cdn.example.com, got %s", query.Get("src"))
		}

		resp := ScriptTagsListResponse{
			Items:      []ScriptTag{{ID: "st_filter", Src: "https://cdn.example.com/script.js"}},
			TotalCount: 1,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &ScriptTagsListOptions{
		Src: "https://cdn.example.com",
	}

	tags, err := client.ListScriptTags(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListScriptTags with src only failed: %v", err)
	}

	if len(tags.Items) != 1 {
		t.Errorf("Expected 1 script tag, got %d", len(tags.Items))
	}
}

func TestListScriptTagsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "server error"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.ListScriptTags(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for server error response, got nil")
	}
}

func TestGetScriptTagError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.GetScriptTag(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Expected error for not found response, got nil")
	}
}

func TestCreateScriptTagError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ScriptTagCreateRequest{
		Src: "invalid-url",
	}

	_, err := client.CreateScriptTag(context.Background(), req)
	if err == nil {
		t.Error("Expected error for bad request response, got nil")
	}
}

func TestCreateScriptTagVerifyRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/script_tags" {
			t.Errorf("Expected path /script_tags, got %s", r.URL.Path)
		}

		var req ScriptTagCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.Src != "https://example.com/analytics.js" {
			t.Errorf("Unexpected src: %s", req.Src)
		}
		if req.Event != "onload" {
			t.Errorf("Unexpected event: %s", req.Event)
		}
		if req.DisplayScope != "all" {
			t.Errorf("Unexpected display_scope: %s", req.DisplayScope)
		}

		tag := ScriptTag{
			ID:           "st_created",
			Src:          req.Src,
			Event:        req.Event,
			DisplayScope: req.DisplayScope,
		}
		_ = json.NewEncoder(w).Encode(tag)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ScriptTagCreateRequest{
		Src:          "https://example.com/analytics.js",
		Event:        "onload",
		DisplayScope: "all",
	}

	tag, err := client.CreateScriptTag(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateScriptTag failed: %v", err)
	}

	if tag.ID != "st_created" {
		t.Errorf("Unexpected script tag ID: %s", tag.ID)
	}
	if tag.DisplayScope != "all" {
		t.Errorf("Unexpected display_scope: %s", tag.DisplayScope)
	}
}

func TestUpdateScriptTagError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "script tag not found"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ScriptTagUpdateRequest{
		Src: "https://example.com/updated.js",
	}

	_, err := client.UpdateScriptTag(context.Background(), "nonexistent", req)
	if err == nil {
		t.Error("Expected error for not found response, got nil")
	}
}

func TestUpdateScriptTagAllFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ScriptTagUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.Src != "https://example.com/new-script.js" {
			t.Errorf("Unexpected src: %s", req.Src)
		}
		if req.Event != "DOMContentLoaded" {
			t.Errorf("Unexpected event: %s", req.Event)
		}
		if req.DisplayScope != "order_status" {
			t.Errorf("Unexpected display_scope: %s", req.DisplayScope)
		}

		tag := ScriptTag{
			ID:           "st_update_all",
			Src:          req.Src,
			Event:        req.Event,
			DisplayScope: req.DisplayScope,
		}
		_ = json.NewEncoder(w).Encode(tag)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ScriptTagUpdateRequest{
		Src:          "https://example.com/new-script.js",
		Event:        "DOMContentLoaded",
		DisplayScope: "order_status",
	}

	tag, err := client.UpdateScriptTag(context.Background(), "st_update_all", req)
	if err != nil {
		t.Fatalf("UpdateScriptTag failed: %v", err)
	}

	if tag.Event != "DOMContentLoaded" {
		t.Errorf("Unexpected event: %s", tag.Event)
	}
	if tag.DisplayScope != "order_status" {
		t.Errorf("Unexpected display_scope: %s", tag.DisplayScope)
	}
}

func TestDeleteScriptTagError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "script tag not found"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteScriptTag(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Expected error for not found response, got nil")
	}
}
