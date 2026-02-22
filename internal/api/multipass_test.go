package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMultipassGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/multipass" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		multipass := Multipass{
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_ = json.NewEncoder(w).Encode(multipass)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	multipass, err := client.GetMultipass(context.Background())
	if err != nil {
		t.Fatalf("GetMultipass failed: %v", err)
	}

	if !multipass.Enabled {
		t.Error("Expected multipass to be enabled")
	}
}

func TestMultipassEnable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/multipass/enable" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		multipass := Multipass{
			Enabled: true,
			Secret:  "abc123secret",
		}
		_ = json.NewEncoder(w).Encode(multipass)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	multipass, err := client.EnableMultipass(context.Background())
	if err != nil {
		t.Fatalf("EnableMultipass failed: %v", err)
	}

	if !multipass.Enabled {
		t.Error("Expected multipass to be enabled")
	}
	if multipass.Secret != "abc123secret" {
		t.Errorf("Unexpected secret: %s", multipass.Secret)
	}
}

func TestMultipassDisable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/multipass/disable" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DisableMultipass(context.Background())
	if err != nil {
		t.Fatalf("DisableMultipass failed: %v", err)
	}
}

func TestMultipassRotateSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/multipass/rotate" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		multipass := Multipass{
			Enabled: true,
			Secret:  "newsecret456",
		}
		_ = json.NewEncoder(w).Encode(multipass)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	multipass, err := client.RotateMultipassSecret(context.Background())
	if err != nil {
		t.Fatalf("RotateMultipassSecret failed: %v", err)
	}

	if multipass.Secret != "newsecret456" {
		t.Errorf("Unexpected secret: %s", multipass.Secret)
	}
}

func TestMultipassGenerateToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/multipass/token" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req MultipassTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Email != "customer@example.com" {
			t.Errorf("Unexpected email: %s", req.Email)
		}

		token := MultipassToken{
			Token:     "multipass_token_xyz",
			URL:       "https://store.example.com/account/login/multipass/multipass_token_xyz",
			ExpiresAt: time.Now().Add(time.Hour),
		}
		_ = json.NewEncoder(w).Encode(token)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &MultipassTokenRequest{
		Email:    "customer@example.com",
		ReturnTo: "/collections/sale",
	}
	token, err := client.GenerateMultipassToken(context.Background(), req)
	if err != nil {
		t.Fatalf("GenerateMultipassToken failed: %v", err)
	}

	if token.Token != "multipass_token_xyz" {
		t.Errorf("Unexpected token: %s", token.Token)
	}
	if token.URL == "" {
		t.Error("Expected URL to be set")
	}
}

func TestMultipassGenerateTokenWithCustomerData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req MultipassTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.CustomerData == nil {
			t.Error("Expected customer_data to be set")
		}
		if req.CustomerData["first_name"] != "John" {
			t.Errorf("Unexpected first_name: %v", req.CustomerData["first_name"])
		}

		token := MultipassToken{
			Token:     "multipass_token_xyz",
			URL:       "https://store.example.com/account/login/multipass/multipass_token_xyz",
			ExpiresAt: time.Now().Add(time.Hour),
		}
		_ = json.NewEncoder(w).Encode(token)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &MultipassTokenRequest{
		Email: "customer@example.com",
		CustomerData: map[string]interface{}{
			"first_name": "John",
			"last_name":  "Doe",
		},
	}
	_, err := client.GenerateMultipassToken(context.Background(), req)
	if err != nil {
		t.Fatalf("GenerateMultipassToken failed: %v", err)
	}
}
