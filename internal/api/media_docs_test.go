package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateMediaImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/media" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["ok"] != true {
			t.Fatalf("expected ok=true in body, got %v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "img_1"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.CreateMediaImage(context.Background(), map[string]any{"ok": true})
	if err != nil {
		t.Fatalf("CreateMediaImage failed: %v", err)
	}
	var got map[string]any
	_ = json.Unmarshal(raw, &got)
	if got["id"] != "img_1" {
		t.Fatalf("expected id=img_1, got %v", got["id"])
	}
}
