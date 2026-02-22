package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPromotionsCouponCenter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/promotions/coupon-center" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.GetPromotionsCouponCenter(context.Background())
	if err != nil {
		t.Fatalf("GetPromotionsCouponCenter failed: %v", err)
	}
	var got map[string]any
	_ = json.Unmarshal(raw, &got)
	if _, ok := got["items"]; !ok {
		t.Fatalf("expected items key, got %v", got)
	}
}
