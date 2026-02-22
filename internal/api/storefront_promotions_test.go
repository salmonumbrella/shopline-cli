package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStorefrontPromotionsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/storefront/promotions" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := StorefrontPromotionsListResponse{
			Items: []StorefrontPromotion{
				{ID: "promo_123", Title: "Summer Sale", DiscountType: "percentage", DiscountValue: "20"},
				{ID: "promo_456", Title: "Free Shipping", DiscountType: "shipping", DiscountValue: "0"},
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

	promos, err := client.ListStorefrontPromotions(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListStorefrontPromotions failed: %v", err)
	}

	if len(promos.Items) != 2 {
		t.Errorf("Expected 2 promotions, got %d", len(promos.Items))
	}
	if promos.Items[0].ID != "promo_123" {
		t.Errorf("Unexpected promotion ID: %s", promos.Items[0].ID)
	}
}

func TestStorefrontPromotionsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("status") != "active" {
			t.Errorf("Expected status=active, got %s", r.URL.Query().Get("status"))
		}
		if r.URL.Query().Get("auto_apply") != "true" {
			t.Errorf("Expected auto_apply=true, got %s", r.URL.Query().Get("auto_apply"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}

		resp := StorefrontPromotionsListResponse{
			Items:      []StorefrontPromotion{{ID: "promo_123", Title: "Auto Promo", AutoApply: true}},
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

	autoApply := true
	opts := &StorefrontPromotionsListOptions{
		Page:      2,
		Status:    "active",
		AutoApply: &autoApply,
	}
	promos, err := client.ListStorefrontPromotions(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListStorefrontPromotions failed: %v", err)
	}

	if len(promos.Items) != 1 {
		t.Errorf("Expected 1 promotion, got %d", len(promos.Items))
	}
}

func TestStorefrontPromotionsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/storefront/promotions/promo_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		promo := StorefrontPromotion{
			ID:            "promo_123",
			Title:         "Summer Sale",
			DiscountType:  "percentage",
			DiscountValue: "20",
			UsageLimit:    1000,
			UsageCount:    250,
			StartsAt:      time.Now(),
		}
		_ = json.NewEncoder(w).Encode(promo)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	promo, err := client.GetStorefrontPromotion(context.Background(), "promo_123")
	if err != nil {
		t.Fatalf("GetStorefrontPromotion failed: %v", err)
	}

	if promo.ID != "promo_123" {
		t.Errorf("Unexpected promotion ID: %s", promo.ID)
	}
	if promo.DiscountValue != "20" {
		t.Errorf("Unexpected discount value: %s", promo.DiscountValue)
	}
}

func TestStorefrontPromotionsGetByCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/storefront/promotions/code/SUMMER20" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		promo := StorefrontPromotion{
			ID:            "promo_123",
			Title:         "Summer Sale",
			Code:          "SUMMER20",
			DiscountType:  "percentage",
			DiscountValue: "20",
		}
		_ = json.NewEncoder(w).Encode(promo)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	promo, err := client.GetStorefrontPromotionByCode(context.Background(), "SUMMER20")
	if err != nil {
		t.Fatalf("GetStorefrontPromotionByCode failed: %v", err)
	}

	if promo.Code != "SUMMER20" {
		t.Errorf("Unexpected promotion code: %s", promo.Code)
	}
}

func TestGetStorefrontPromotionEmptyID(t *testing.T) {
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
			_, err := client.GetStorefrontPromotion(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "promotion id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestGetStorefrontPromotionByCodeEmpty(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name string
		code string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab only", "\t"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.GetStorefrontPromotionByCode(context.Background(), tc.code)
			if err == nil {
				t.Error("Expected error for empty code, got nil")
			}
			if err != nil && err.Error() != "promotion code is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
