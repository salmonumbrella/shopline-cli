package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetStaffPermissions(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method: got %s want %s", r.Method, http.MethodGet)
		}
		if r.URL.Path != "/staffs/staff_123/permissions" {
			t.Fatalf("path: got %s want %s", r.URL.Path, "/staffs/staff_123/permissions")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"permissions": []string{"orders"}})
	}))
	t.Cleanup(srv.Close)

	client := NewClient("token")
	client.BaseURL = srv.URL
	got, err := client.GetStaffPermissions(context.Background(), "staff_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) == "" {
		t.Fatal("expected non-empty response")
	}
}
