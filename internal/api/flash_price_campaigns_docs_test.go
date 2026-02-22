package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListFlashPriceCampaigns(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/flash_price_campaigns" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.ListFlashPriceCampaigns(context.Background(), &FlashPriceCampaignsListOptions{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListFlashPriceCampaigns failed: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("failed to unmarshal raw: %v", err)
	}
	if _, ok := got["items"]; !ok {
		t.Fatalf("expected items key, got %v", got)
	}
}

func TestGetFlashPriceCampaign(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/flash_price_campaigns/fpc_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "fpc_123"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.GetFlashPriceCampaign(context.Background(), "fpc_123")
	if err != nil {
		t.Fatalf("GetFlashPriceCampaign failed: %v", err)
	}
	var got map[string]any
	_ = json.Unmarshal(raw, &got)
	if got["id"] != "fpc_123" {
		t.Fatalf("expected id=fpc_123, got %v", got["id"])
	}
}

func TestCreateFlashPriceCampaign(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/flash_price_campaigns" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "Test" {
			t.Fatalf("expected name=Test, got %v", body["name"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"created": true})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.CreateFlashPriceCampaign(context.Background(), map[string]any{"name": "Test"})
	if err != nil {
		t.Fatalf("CreateFlashPriceCampaign failed: %v", err)
	}
	var got map[string]any
	_ = json.Unmarshal(raw, &got)
	if got["created"] != true {
		t.Fatalf("expected created=true, got %v", got["created"])
	}
}

func TestUpdateFlashPriceCampaign(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/flash_price_campaigns/fpc_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"updated": true})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.UpdateFlashPriceCampaign(context.Background(), "fpc_123", map[string]any{"name": "New"})
	if err != nil {
		t.Fatalf("UpdateFlashPriceCampaign failed: %v", err)
	}
	var got map[string]any
	_ = json.Unmarshal(raw, &got)
	if got["updated"] != true {
		t.Fatalf("expected updated=true, got %v", got["updated"])
	}
}

func TestDeleteFlashPriceCampaign(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/flash_price_campaigns/fpc_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	if err := client.DeleteFlashPriceCampaign(context.Background(), "fpc_123"); err != nil {
		t.Fatalf("DeleteFlashPriceCampaign failed: %v", err)
	}
}
