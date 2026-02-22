package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStorefrontOAuthClientsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/storefront_oauth/clients" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := StorefrontOAuthClientsListResponse{
			Items: []StorefrontOAuthClient{
				{
					ID:           "oauth_123",
					Name:         "Mobile App",
					ClientID:     "client_abc",
					RedirectURIs: []string{"https://app.example.com/callback"},
					Scopes:       []string{"read_products", "read_customers"},
				},
				{
					ID:           "oauth_456",
					Name:         "Partner Integration",
					ClientID:     "client_def",
					RedirectURIs: []string{"https://partner.example.com/auth"},
					Scopes:       []string{"read_orders"},
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

	clients, err := client.ListStorefrontOAuthClients(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListStorefrontOAuthClients failed: %v", err)
	}

	if len(clients.Items) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(clients.Items))
	}
	if clients.Items[0].ID != "oauth_123" {
		t.Errorf("Unexpected client ID: %s", clients.Items[0].ID)
	}
	if clients.Items[0].Name != "Mobile App" {
		t.Errorf("Unexpected name: %s", clients.Items[0].Name)
	}
}

func TestStorefrontOAuthClientsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("page_size") != "10" {
			t.Errorf("Expected page_size=10, got %s", r.URL.Query().Get("page_size"))
		}

		resp := StorefrontOAuthClientsListResponse{
			Items:      []StorefrontOAuthClient{{ID: "oauth_123", Name: "Test"}},
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

	opts := &StorefrontOAuthClientsListOptions{
		Page:     2,
		PageSize: 10,
	}
	clients, err := client.ListStorefrontOAuthClients(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListStorefrontOAuthClients failed: %v", err)
	}

	if len(clients.Items) != 1 {
		t.Errorf("Expected 1 client, got %d", len(clients.Items))
	}
}

func TestStorefrontOAuthClientsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/storefront_oauth/clients/oauth_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		oauthClient := StorefrontOAuthClient{
			ID:           "oauth_123",
			Name:         "Mobile App",
			ClientID:     "client_abc",
			RedirectURIs: []string{"https://app.example.com/callback"},
			Scopes:       []string{"read_products"},
		}
		_ = json.NewEncoder(w).Encode(oauthClient)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	oauthClient, err := client.GetStorefrontOAuthClient(context.Background(), "oauth_123")
	if err != nil {
		t.Fatalf("GetStorefrontOAuthClient failed: %v", err)
	}

	if oauthClient.ID != "oauth_123" {
		t.Errorf("Unexpected client ID: %s", oauthClient.ID)
	}
	if oauthClient.ClientID != "client_abc" {
		t.Errorf("Unexpected client_id: %s", oauthClient.ClientID)
	}
}

func TestStorefrontOAuthClientsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/storefront_oauth/clients" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req StorefrontOAuthClientCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Name != "New OAuth Client" {
			t.Errorf("Unexpected name: %s", req.Name)
		}

		oauthClient := StorefrontOAuthClient{
			ID:           "oauth_new",
			Name:         req.Name,
			ClientID:     "client_xyz",
			ClientSecret: "secret_xyz",
			RedirectURIs: req.RedirectURIs,
			Scopes:       req.Scopes,
		}
		_ = json.NewEncoder(w).Encode(oauthClient)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &StorefrontOAuthClientCreateRequest{
		Name:         "New OAuth Client",
		RedirectURIs: []string{"https://example.com/callback"},
		Scopes:       []string{"read_products"},
	}
	oauthClient, err := client.CreateStorefrontOAuthClient(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateStorefrontOAuthClient failed: %v", err)
	}

	if oauthClient.ID != "oauth_new" {
		t.Errorf("Unexpected client ID: %s", oauthClient.ID)
	}
	if oauthClient.ClientSecret != "secret_xyz" {
		t.Errorf("Unexpected client secret: %s", oauthClient.ClientSecret)
	}
}

func TestStorefrontOAuthClientsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/storefront_oauth/clients/oauth_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req StorefrontOAuthClientUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		oauthClient := StorefrontOAuthClient{
			ID:       "oauth_123",
			Name:     req.Name,
			ClientID: "client_abc",
		}
		_ = json.NewEncoder(w).Encode(oauthClient)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &StorefrontOAuthClientUpdateRequest{
		Name: "Updated Name",
	}
	oauthClient, err := client.UpdateStorefrontOAuthClient(context.Background(), "oauth_123", req)
	if err != nil {
		t.Fatalf("UpdateStorefrontOAuthClient failed: %v", err)
	}

	if oauthClient.Name != "Updated Name" {
		t.Errorf("Unexpected name: %s", oauthClient.Name)
	}
}

func TestStorefrontOAuthClientsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/storefront_oauth/clients/oauth_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteStorefrontOAuthClient(context.Background(), "oauth_123")
	if err != nil {
		t.Fatalf("DeleteStorefrontOAuthClient failed: %v", err)
	}
}

func TestStorefrontOAuthClientsRotateSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/storefront_oauth/clients/oauth_123/rotate_secret" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		oauthClient := StorefrontOAuthClient{
			ID:           "oauth_123",
			Name:         "Test Client",
			ClientID:     "client_abc",
			ClientSecret: "new_secret_123",
		}
		_ = json.NewEncoder(w).Encode(oauthClient)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	oauthClient, err := client.RotateStorefrontOAuthClientSecret(context.Background(), "oauth_123")
	if err != nil {
		t.Fatalf("RotateStorefrontOAuthClientSecret failed: %v", err)
	}

	if oauthClient.ClientSecret != "new_secret_123" {
		t.Errorf("Unexpected client secret: %s", oauthClient.ClientSecret)
	}
}

func TestGetStorefrontOAuthClientEmptyID(t *testing.T) {
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
			_, err := client.GetStorefrontOAuthClient(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "oauth client id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestUpdateStorefrontOAuthClientEmptyID(t *testing.T) {
	client := NewClient("token")

	_, err := client.UpdateStorefrontOAuthClient(context.Background(), "", &StorefrontOAuthClientUpdateRequest{Name: "Test"})
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
	if err != nil && err.Error() != "oauth client id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestDeleteStorefrontOAuthClientEmptyID(t *testing.T) {
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
			err := client.DeleteStorefrontOAuthClient(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "oauth client id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestRotateStorefrontOAuthClientSecretEmptyID(t *testing.T) {
	client := NewClient("token")

	_, err := client.RotateStorefrontOAuthClientSecret(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
	if err != nil && err.Error() != "oauth client id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}
