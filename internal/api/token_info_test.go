package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTokenInfo(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method: got %s want %s", r.Method, http.MethodGet)
		}
		if r.URL.Path != "/token/info" {
			t.Fatalf("path: got %s want %s", r.URL.Path, "/token/info")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	t.Cleanup(srv.Close)

	client := NewClient("token")
	client.BaseURL = srv.URL

	got, err := client.GetTokenInfo(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) == "" {
		t.Fatal("expected non-empty response")
	}
}
