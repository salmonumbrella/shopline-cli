package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTokensList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/tokens" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := TokensListResponse{
			Items: []Token{
				{
					ID:     "tok_123",
					Title:  "Production API",
					Scopes: []string{"read_products", "write_orders"},
				},
				{
					ID:     "tok_456",
					Title:  "Development API",
					Scopes: []string{"read_products"},
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

	tokens, err := client.ListTokens(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTokens failed: %v", err)
	}

	if len(tokens.Items) != 2 {
		t.Errorf("Expected 2 tokens, got %d", len(tokens.Items))
	}
	if tokens.Items[0].ID != "tok_123" {
		t.Errorf("Unexpected token ID: %s", tokens.Items[0].ID)
	}
	if tokens.Items[0].Title != "Production API" {
		t.Errorf("Unexpected title: %s", tokens.Items[0].Title)
	}
}

func TestTokensListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("page_size") != "10" {
			t.Errorf("Expected page_size=10, got %s", r.URL.Query().Get("page_size"))
		}

		resp := TokensListResponse{
			Items:      []Token{{ID: "tok_123", Title: "Test Token"}},
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

	opts := &TokensListOptions{
		Page:     2,
		PageSize: 10,
	}
	tokens, err := client.ListTokens(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListTokens failed: %v", err)
	}

	if len(tokens.Items) != 1 {
		t.Errorf("Expected 1 token, got %d", len(tokens.Items))
	}
}

func TestTokensGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tokens/tok_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		token := Token{
			ID:     "tok_123",
			Title:  "Production API",
			Scopes: []string{"read_products", "write_orders"},
		}
		_ = json.NewEncoder(w).Encode(token)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	token, err := client.GetToken(context.Background(), "tok_123")
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if token.ID != "tok_123" {
		t.Errorf("Unexpected token ID: %s", token.ID)
	}
	if token.Title != "Production API" {
		t.Errorf("Unexpected title: %s", token.Title)
	}
	if len(token.Scopes) != 2 {
		t.Errorf("Expected 2 scopes, got %d", len(token.Scopes))
	}
}

func TestTokensCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/tokens" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req TokenCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Title != "New Token" {
			t.Errorf("Unexpected title: %s", req.Title)
		}

		token := Token{
			ID:          "tok_new",
			Title:       req.Title,
			AccessToken: "shpat_xxxxx",
			Scopes:      req.Scopes,
		}
		_ = json.NewEncoder(w).Encode(token)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &TokenCreateRequest{
		Title:  "New Token",
		Scopes: []string{"read_products"},
	}
	token, err := client.CreateToken(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateToken failed: %v", err)
	}

	if token.ID != "tok_new" {
		t.Errorf("Unexpected token ID: %s", token.ID)
	}
	if token.AccessToken != "shpat_xxxxx" {
		t.Errorf("Unexpected access token: %s", token.AccessToken)
	}
}

func TestTokensDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/tokens/tok_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteToken(context.Background(), "tok_123")
	if err != nil {
		t.Fatalf("DeleteToken failed: %v", err)
	}
}

func TestGetTokenEmptyID(t *testing.T) {
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
			_, err := client.GetToken(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "token id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteTokenEmptyID(t *testing.T) {
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
			err := client.DeleteToken(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "token id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
