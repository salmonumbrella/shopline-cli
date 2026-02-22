package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStorefrontTokensList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/storefront_tokens" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := StorefrontTokensListResponse{
			Items: []StorefrontToken{
				{
					ID:    "sft_123",
					Title: "Mobile App",
				},
				{
					ID:    "sft_456",
					Title: "Web Storefront",
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

	tokens, err := client.ListStorefrontTokens(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListStorefrontTokens failed: %v", err)
	}

	if len(tokens.Items) != 2 {
		t.Errorf("Expected 2 tokens, got %d", len(tokens.Items))
	}
	if tokens.Items[0].ID != "sft_123" {
		t.Errorf("Unexpected token ID: %s", tokens.Items[0].ID)
	}
	if tokens.Items[0].Title != "Mobile App" {
		t.Errorf("Unexpected title: %s", tokens.Items[0].Title)
	}
}

func TestStorefrontTokensListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("page_size") != "10" {
			t.Errorf("Expected page_size=10, got %s", r.URL.Query().Get("page_size"))
		}

		resp := StorefrontTokensListResponse{
			Items:      []StorefrontToken{{ID: "sft_123", Title: "Test"}},
			Page:       2,
			PageSize:   10,
			TotalCount: 1,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &StorefrontTokensListOptions{
		Page:     2,
		PageSize: 10,
	}
	tokens, err := client.ListStorefrontTokens(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListStorefrontTokens failed: %v", err)
	}

	if len(tokens.Items) != 1 {
		t.Errorf("Expected 1 token, got %d", len(tokens.Items))
	}
}

func TestStorefrontTokensGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/storefront_tokens/sft_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		token := StorefrontToken{
			ID:    "sft_123",
			Title: "Mobile App",
		}
		_ = json.NewEncoder(w).Encode(token)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	token, err := client.GetStorefrontToken(context.Background(), "sft_123")
	if err != nil {
		t.Fatalf("GetStorefrontToken failed: %v", err)
	}

	if token.ID != "sft_123" {
		t.Errorf("Unexpected token ID: %s", token.ID)
	}
	if token.Title != "Mobile App" {
		t.Errorf("Unexpected title: %s", token.Title)
	}
}

func TestStorefrontTokensCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/storefront_tokens" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req StorefrontTokenCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Title != "New Storefront" {
			t.Errorf("Unexpected title: %s", req.Title)
		}

		token := StorefrontToken{
			ID:          "sft_new",
			Title:       req.Title,
			AccessToken: "shpsf_xxxxx",
		}
		_ = json.NewEncoder(w).Encode(token)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &StorefrontTokenCreateRequest{
		Title: "New Storefront",
	}
	token, err := client.CreateStorefrontToken(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateStorefrontToken failed: %v", err)
	}

	if token.ID != "sft_new" {
		t.Errorf("Unexpected token ID: %s", token.ID)
	}
	if token.AccessToken != "shpsf_xxxxx" {
		t.Errorf("Unexpected access token: %s", token.AccessToken)
	}
}

func TestStorefrontTokensDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/storefront_tokens/sft_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteStorefrontToken(context.Background(), "sft_123")
	if err != nil {
		t.Fatalf("DeleteStorefrontToken failed: %v", err)
	}
}

func TestGetStorefrontTokenEmptyID(t *testing.T) {
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
			_, err := client.GetStorefrontToken(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "storefront token id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteStorefrontTokenEmptyID(t *testing.T) {
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
			err := client.DeleteStorefrontToken(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "storefront token id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
