package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebhooksList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/webhooks" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := WebhooksListResponse{
			Items: []Webhook{
				{
					ID:         "wh_123",
					Address:    "https://example.com/webhook",
					Topic:      "orders/create",
					Format:     WebhookFormatJSON,
					APIVersion: "2024-01",
				},
				{
					ID:         "wh_456",
					Address:    "https://example.com/products",
					Topic:      "products/update",
					Format:     WebhookFormatJSON,
					APIVersion: "2024-01",
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

	webhooks, err := client.ListWebhooks(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListWebhooks failed: %v", err)
	}

	if len(webhooks.Items) != 2 {
		t.Errorf("Expected 2 webhooks, got %d", len(webhooks.Items))
	}
	if webhooks.Items[0].ID != "wh_123" {
		t.Errorf("Unexpected webhook ID: %s", webhooks.Items[0].ID)
	}
	if webhooks.Items[0].Topic != "orders/create" {
		t.Errorf("Unexpected topic: %s", webhooks.Items[0].Topic)
	}
}

func TestWebhooksListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("topic") != "orders/create" {
			t.Errorf("Expected topic=orders/create, got %s", r.URL.Query().Get("topic"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}

		resp := WebhooksListResponse{
			Items: []Webhook{
				{ID: "wh_123", Topic: "orders/create"},
			},
			Page:       2,
			PageSize:   20,
			TotalCount: 1,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &WebhooksListOptions{
		Page:  2,
		Topic: "orders/create",
	}
	webhooks, err := client.ListWebhooks(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListWebhooks failed: %v", err)
	}

	if len(webhooks.Items) != 1 {
		t.Errorf("Expected 1 webhook, got %d", len(webhooks.Items))
	}
}

func TestWebhooksGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/webhooks/wh_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		webhook := Webhook{
			ID:         "wh_123",
			Address:    "https://example.com/webhook",
			Topic:      "orders/create",
			Format:     WebhookFormatJSON,
			APIVersion: "2024-01",
		}
		_ = json.NewEncoder(w).Encode(webhook)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	webhook, err := client.GetWebhook(context.Background(), "wh_123")
	if err != nil {
		t.Fatalf("GetWebhook failed: %v", err)
	}

	if webhook.ID != "wh_123" {
		t.Errorf("Unexpected webhook ID: %s", webhook.ID)
	}
	if webhook.Address != "https://example.com/webhook" {
		t.Errorf("Unexpected address: %s", webhook.Address)
	}
	if webhook.Format != WebhookFormatJSON {
		t.Errorf("Unexpected format: %s", webhook.Format)
	}
}

func TestWebhooksCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/webhooks" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req WebhookCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Address != "https://example.com/webhook" {
			t.Errorf("Unexpected address: %s", req.Address)
		}
		if req.Topic != "orders/create" {
			t.Errorf("Unexpected topic: %s", req.Topic)
		}

		webhook := Webhook{
			ID:         "wh_new",
			Address:    req.Address,
			Topic:      req.Topic,
			Format:     WebhookFormatJSON,
			APIVersion: "2024-01",
		}
		_ = json.NewEncoder(w).Encode(webhook)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &WebhookCreateRequest{
		Address: "https://example.com/webhook",
		Topic:   "orders/create",
	}
	webhook, err := client.CreateWebhook(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateWebhook failed: %v", err)
	}

	if webhook.ID != "wh_new" {
		t.Errorf("Unexpected webhook ID: %s", webhook.ID)
	}
	if webhook.Address != "https://example.com/webhook" {
		t.Errorf("Unexpected address: %s", webhook.Address)
	}
}

func TestWebhooksUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/webhooks/wh_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req WebhookUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}
		if req.Address != "https://example.com/updated" {
			t.Errorf("Unexpected address: %s", req.Address)
		}

		webhook := Webhook{
			ID:         "wh_123",
			Address:    req.Address,
			Topic:      "orders/create",
			Format:     WebhookFormatJSON,
			APIVersion: "2024-01",
		}
		_ = json.NewEncoder(w).Encode(webhook)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &WebhookUpdateRequest{
		Address: "https://example.com/updated",
	}
	webhook, err := client.UpdateWebhook(context.Background(), "wh_123", req)
	if err != nil {
		t.Fatalf("UpdateWebhook failed: %v", err)
	}

	if webhook.Address != "https://example.com/updated" {
		t.Errorf("Unexpected address: %s", webhook.Address)
	}
}

func TestWebhooksDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/webhooks/wh_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteWebhook(context.Background(), "wh_123")
	if err != nil {
		t.Fatalf("DeleteWebhook failed: %v", err)
	}
}

func TestGetWebhookEmptyID(t *testing.T) {
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
			_, err := client.GetWebhook(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "webhook id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteWebhookEmptyID(t *testing.T) {
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
			err := client.DeleteWebhook(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "webhook id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
