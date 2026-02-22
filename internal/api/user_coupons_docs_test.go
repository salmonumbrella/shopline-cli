package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserCouponsDocsEndpoints(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/user_coupons/list", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method: %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{map[string]any{"id": "uc_1"}}})
	})
	mux.HandleFunc("/user_coupons/CODE/claim", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method: %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"claimed": true})
	})
	mux.HandleFunc("/user_coupons/CODE/redeem", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method: %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"redeemed": true})
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := NewClient("token")
	client.BaseURL = srv.URL

	if _, err := client.ListUserCouponsListEndpoint(context.Background(), &UserCouponsListEndpointOptions{Page: 1, PageSize: 2}); err != nil {
		t.Fatalf("ListUserCouponsListEndpoint: %v", err)
	}
	if _, err := client.ClaimUserCoupon(context.Background(), "CODE", nil); err != nil {
		t.Fatalf("ClaimUserCoupon: %v", err)
	}
	if _, err := client.RedeemUserCoupon(context.Background(), "CODE", nil); err != nil {
		t.Fatalf("RedeemUserCoupon: %v", err)
	}
}
