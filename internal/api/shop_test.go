package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetShop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/shop" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		shop := Shop{
			ID:               "shop_123",
			Name:             "My Store",
			Email:            "owner@mystore.com",
			Domain:           "mystore.com",
			MyshoplineDomain: "mystore.myshopline.com",
			Currency:         "USD",
			PlanName:         "professional",
		}
		_ = json.NewEncoder(w).Encode(shop)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	shop, err := client.GetShop(context.Background())
	if err != nil {
		t.Fatalf("GetShop failed: %v", err)
	}

	if shop.ID != "shop_123" {
		t.Errorf("Unexpected shop ID: %s", shop.ID)
	}
	if shop.Name != "My Store" {
		t.Errorf("Unexpected shop name: %s", shop.Name)
	}
	if shop.Currency != "USD" {
		t.Errorf("Unexpected currency: %s", shop.Currency)
	}
}

func TestGetShopSettings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/shop/settings" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		settings := ShopSettings{
			ID:            "settings_123",
			Currency:      "USD",
			WeightUnit:    "lb",
			Timezone:      "America/New_York",
			TaxesIncluded: false,
		}
		_ = json.NewEncoder(w).Encode(settings)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	settings, err := client.GetShopSettings(context.Background())
	if err != nil {
		t.Fatalf("GetShopSettings failed: %v", err)
	}

	if settings.ID != "settings_123" {
		t.Errorf("Unexpected settings ID: %s", settings.ID)
	}
	if settings.Currency != "USD" {
		t.Errorf("Unexpected currency: %s", settings.Currency)
	}
}

func TestUpdateShopSettings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/shop/settings" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		settings := ShopSettings{
			ID:         "settings_123",
			Currency:   "EUR",
			WeightUnit: "kg",
		}
		_ = json.NewEncoder(w).Encode(settings)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ShopSettingsUpdateRequest{Currency: "EUR", WeightUnit: "kg"}
	settings, err := client.UpdateShopSettings(context.Background(), req)
	if err != nil {
		t.Fatalf("UpdateShopSettings failed: %v", err)
	}

	if settings.Currency != "EUR" {
		t.Errorf("Expected EUR, got: %s", settings.Currency)
	}
}
