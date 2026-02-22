package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetMerchantByID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/merchants/m_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "m_123"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.GetMerchantByID(context.Background(), "m_123")
	if err != nil {
		t.Fatalf("GetMerchantByID failed: %v", err)
	}
	var got map[string]any
	_ = json.Unmarshal(raw, &got)
	if got["id"] != "m_123" {
		t.Fatalf("expected id=m_123, got %v", got["id"])
	}
}

func TestGetMerchantByIDEmptyID(t *testing.T) {
	client := NewClient("token")
	_, err := client.GetMerchantByID(context.Background(), " ")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "merchant id is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateMerchantExpressLink(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/merchants/generate_express_link" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["ok"] != true {
			t.Fatalf("expected ok=true in body, got %v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"url": "https://example.com/express"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.GenerateMerchantExpressLink(context.Background(), map[string]any{"ok": true})
	if err != nil {
		t.Fatalf("GenerateMerchantExpressLink failed: %v", err)
	}
	var got map[string]any
	_ = json.Unmarshal(raw, &got)
	if got["url"] == "" {
		t.Fatalf("expected url in response, got %v", got)
	}
}
