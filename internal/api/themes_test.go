package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestThemesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/themes" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := ThemesListResponse{
			Items: []Theme{
				{ID: "thm_123", Name: "Main Theme", Role: "main"},
				{ID: "thm_456", Name: "Development Theme", Role: "unpublished"},
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

	themes, err := client.ListThemes(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListThemes failed: %v", err)
	}

	if len(themes.Items) != 2 {
		t.Errorf("Expected 2 themes, got %d", len(themes.Items))
	}
	if themes.Items[0].ID != "thm_123" {
		t.Errorf("Unexpected theme ID: %s", themes.Items[0].ID)
	}
}

func TestThemesGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/themes/thm_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		theme := Theme{ID: "thm_123", Name: "Main Theme", Role: "main"}
		_ = json.NewEncoder(w).Encode(theme)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	theme, err := client.GetTheme(context.Background(), "thm_123")
	if err != nil {
		t.Fatalf("GetTheme failed: %v", err)
	}

	if theme.ID != "thm_123" {
		t.Errorf("Unexpected theme ID: %s", theme.ID)
	}
}

func TestGetThemeEmptyID(t *testing.T) {
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
			_, err := client.GetTheme(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "theme id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestThemesCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/themes" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ThemeCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Name != "My New Theme" {
			t.Errorf("Unexpected name: %s", req.Name)
		}
		if req.Role != "unpublished" {
			t.Errorf("Unexpected role: %s", req.Role)
		}

		theme := Theme{
			ID:   "thm_new",
			Name: req.Name,
			Role: req.Role,
		}
		_ = json.NewEncoder(w).Encode(theme)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ThemeCreateRequest{
		Name: "My New Theme",
		Role: "unpublished",
	}
	theme, err := client.CreateTheme(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateTheme failed: %v", err)
	}

	if theme.ID != "thm_new" {
		t.Errorf("Unexpected theme ID: %s", theme.ID)
	}
	if theme.Name != "My New Theme" {
		t.Errorf("Unexpected name: %s", theme.Name)
	}
}

func TestThemesUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/themes/thm_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ThemeUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Name != "Updated Theme Name" {
			t.Errorf("Unexpected name: %s", req.Name)
		}

		theme := Theme{
			ID:   "thm_123",
			Name: req.Name,
			Role: "main",
		}
		_ = json.NewEncoder(w).Encode(theme)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ThemeUpdateRequest{
		Name: "Updated Theme Name",
	}
	theme, err := client.UpdateTheme(context.Background(), "thm_123", req)
	if err != nil {
		t.Fatalf("UpdateTheme failed: %v", err)
	}

	if theme.ID != "thm_123" {
		t.Errorf("Unexpected theme ID: %s", theme.ID)
	}
	if theme.Name != "Updated Theme Name" {
		t.Errorf("Unexpected name: %s", theme.Name)
	}
}

func TestThemesDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/themes/thm_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteTheme(context.Background(), "thm_123")
	if err != nil {
		t.Fatalf("DeleteTheme failed: %v", err)
	}
}

func TestUpdateThemeEmptyID(t *testing.T) {
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
			_, err := client.UpdateTheme(context.Background(), tc.id, &ThemeUpdateRequest{Name: "Test"})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "theme id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteThemeEmptyID(t *testing.T) {
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
			err := client.DeleteTheme(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "theme id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
