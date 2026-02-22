package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateConversationShopMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/conversations/message" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body["ok"] != true {
			t.Fatalf("expected ok=true in body, got %v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "msg_1"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.CreateConversationShopMessage(context.Background(), map[string]any{"ok": true})
	if err != nil {
		t.Fatalf("CreateConversationShopMessage failed: %v", err)
	}
	if len(raw) == 0 {
		t.Fatalf("expected response body")
	}
}
