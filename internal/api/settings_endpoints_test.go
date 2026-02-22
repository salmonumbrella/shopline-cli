package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSettingsEndpoints(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	expect := func(path, method string) {
		t.Helper()
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != method {
				t.Fatalf("%s method: got %s want %s", path, r.Method, method)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"path": path, "method": method})
		})
	}

	expect("/settings/checkout", http.MethodGet)
	mux.HandleFunc("/settings/domains", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodPut:
		default:
			t.Fatalf("/settings/domains method: %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})

	expect("/settings/layouts", http.MethodGet)
	mux.HandleFunc("/settings/layouts/draft", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodPut:
		default:
			t.Fatalf("/settings/layouts/draft method: %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	expect("/settings/layouts/publish", http.MethodPost)

	expect("/settings/orders", http.MethodGet)
	expect("/settings/payments", http.MethodGet)
	expect("/settings/pos", http.MethodGet)
	expect("/settings/product_review", http.MethodGet)
	expect("/settings/products", http.MethodGet)
	expect("/settings/promotions", http.MethodGet)
	expect("/settings/shop", http.MethodGet)
	expect("/settings/tax", http.MethodGet)

	expect("/settings/theme", http.MethodGet)
	mux.HandleFunc("/settings/theme/draft", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodPut:
		default:
			t.Fatalf("/settings/theme/draft method: %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	expect("/settings/theme/publish", http.MethodPost)

	expect("/settings/third_party_ads", http.MethodGet)
	expect("/settings/users", http.MethodGet)

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := NewClient("token")
	client.BaseURL = srv.URL

	if _, err := client.GetSettingsCheckout(context.Background()); err != nil {
		t.Fatalf("GetSettingsCheckout: %v", err)
	}
	if _, err := client.GetSettingsDomains(context.Background()); err != nil {
		t.Fatalf("GetSettingsDomains: %v", err)
	}
	if _, err := client.UpdateSettingsDomains(context.Background(), json.RawMessage(`{"x":1}`)); err != nil {
		t.Fatalf("UpdateSettingsDomains: %v", err)
	}
	if _, err := client.GetSettingsLayouts(context.Background()); err != nil {
		t.Fatalf("GetSettingsLayouts: %v", err)
	}
	if _, err := client.GetSettingsLayoutsDraft(context.Background()); err != nil {
		t.Fatalf("GetSettingsLayoutsDraft: %v", err)
	}
	if _, err := client.UpdateSettingsLayoutsDraft(context.Background(), json.RawMessage(`{"x":1}`)); err != nil {
		t.Fatalf("UpdateSettingsLayoutsDraft: %v", err)
	}
	if _, err := client.PublishSettingsLayouts(context.Background(), nil); err != nil {
		t.Fatalf("PublishSettingsLayouts: %v", err)
	}
	if _, err := client.GetSettingsOrders(context.Background()); err != nil {
		t.Fatalf("GetSettingsOrders: %v", err)
	}
	if _, err := client.GetSettingsPayments(context.Background()); err != nil {
		t.Fatalf("GetSettingsPayments: %v", err)
	}
	if _, err := client.GetSettingsPOS(context.Background()); err != nil {
		t.Fatalf("GetSettingsPOS: %v", err)
	}
	if _, err := client.GetSettingsProductReview(context.Background()); err != nil {
		t.Fatalf("GetSettingsProductReview: %v", err)
	}
	if _, err := client.GetSettingsProducts(context.Background()); err != nil {
		t.Fatalf("GetSettingsProducts: %v", err)
	}
	if _, err := client.GetSettingsPromotions(context.Background()); err != nil {
		t.Fatalf("GetSettingsPromotions: %v", err)
	}
	if _, err := client.GetSettingsShop(context.Background()); err != nil {
		t.Fatalf("GetSettingsShop: %v", err)
	}
	if _, err := client.GetSettingsTax(context.Background()); err != nil {
		t.Fatalf("GetSettingsTax: %v", err)
	}
	if _, err := client.GetSettingsTheme(context.Background()); err != nil {
		t.Fatalf("GetSettingsTheme: %v", err)
	}
	if _, err := client.GetSettingsThemeDraft(context.Background()); err != nil {
		t.Fatalf("GetSettingsThemeDraft: %v", err)
	}
	if _, err := client.UpdateSettingsThemeDraft(context.Background(), json.RawMessage(`{"x":1}`)); err != nil {
		t.Fatalf("UpdateSettingsThemeDraft: %v", err)
	}
	if _, err := client.PublishSettingsTheme(context.Background(), nil); err != nil {
		t.Fatalf("PublishSettingsTheme: %v", err)
	}
	if _, err := client.GetSettingsThirdPartyAds(context.Background()); err != nil {
		t.Fatalf("GetSettingsThirdPartyAds: %v", err)
	}
	if _, err := client.GetSettingsUsers(context.Background()); err != nil {
		t.Fatalf("GetSettingsUsers: %v", err)
	}
}
