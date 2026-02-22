package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChannelPricesGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/channels/ch_123/prices" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.GetChannelPrices(context.Background(), "ch_123")
	if err != nil {
		t.Fatalf("GetChannelPrices failed: %v", err)
	}
}

func TestChannelPricesCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/channels/ch_123/products/prod_1/prices" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		if body["price"] != "10.00" {
			t.Errorf("Unexpected request body: %v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "price_1"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.CreateChannelProductPrice(context.Background(), "ch_123", "prod_1", map[string]any{"price": "10.00"})
	if err != nil {
		t.Fatalf("CreateChannelProductPrice failed: %v", err)
	}
}

func TestChannelPricesUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/channels/ch_123/products/prod_1/prices/price_1" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		if body["price"] != "11.00" {
			t.Errorf("Unexpected request body: %v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "price_1"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.UpdateChannelProductPrice(context.Background(), "ch_123", "prod_1", "price_1", map[string]any{"price": "11.00"})
	if err != nil {
		t.Fatalf("UpdateChannelProductPrice failed: %v", err)
	}
}
