package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGiftsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/gifts" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := GiftsListResponse{
			Items: []Gift{
				{ID: "gift_123", Title: "Free Sample", GiftProductName: "Sample Product", Status: "active"},
				{ID: "gift_456", Title: "Holiday Gift", GiftProductName: "Holiday Item", Status: "scheduled"},
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

	gifts, err := client.ListGifts(context.Background(), &GiftsListOptions{})
	if err != nil {
		t.Fatalf("ListGifts failed: %v", err)
	}

	if len(gifts.Items) != 2 {
		t.Errorf("Expected 2 gifts, got %d", len(gifts.Items))
	}
	if gifts.Items[0].ID != "gift_123" {
		t.Errorf("Unexpected gift ID: %s", gifts.Items[0].ID)
	}
}

func TestGiftsListWithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("status") != "active" {
			t.Errorf("Expected status=active, got %s", r.URL.Query().Get("status"))
		}

		resp := GiftsListResponse{Items: []Gift{}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.ListGifts(context.Background(), &GiftsListOptions{
		Status: "active",
	})
	if err != nil {
		t.Fatalf("ListGifts failed: %v", err)
	}
}

func TestGiftsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/gifts/gift_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		gift := Gift{ID: "gift_123", Title: "Free Sample", GiftProductName: "Sample Product"}
		_ = json.NewEncoder(w).Encode(gift)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	gift, err := client.GetGift(context.Background(), "gift_123")
	if err != nil {
		t.Fatalf("GetGift failed: %v", err)
	}

	if gift.ID != "gift_123" {
		t.Errorf("Unexpected gift ID: %s", gift.ID)
	}
}

func TestGetGiftEmptyID(t *testing.T) {
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
			_, err := client.GetGift(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "gift id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestGiftsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/gifts" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		gift := Gift{ID: "gift_new", Title: "New Gift", GiftProductID: "prod_123"}
		_ = json.NewEncoder(w).Encode(gift)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &GiftCreateRequest{
		Title:         "New Gift",
		GiftProductID: "prod_123",
		TriggerType:   "min_purchase",
		TriggerValue:  100.00,
	}

	gift, err := client.CreateGift(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateGift failed: %v", err)
	}

	if gift.Title != "New Gift" {
		t.Errorf("Unexpected gift title: %s", gift.Title)
	}
}

func TestGiftsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/gifts/gift_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteGift(context.Background(), "gift_123")
	if err != nil {
		t.Fatalf("DeleteGift failed: %v", err)
	}
}

func TestDeleteGiftEmptyID(t *testing.T) {
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
			err := client.DeleteGift(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "gift id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestGiftsActivate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/gifts/gift_123/activate" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		gift := Gift{ID: "gift_123", Status: "active"}
		_ = json.NewEncoder(w).Encode(gift)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	gift, err := client.ActivateGift(context.Background(), "gift_123")
	if err != nil {
		t.Fatalf("ActivateGift failed: %v", err)
	}

	if gift.Status != "active" {
		t.Errorf("Expected active status, got: %s", gift.Status)
	}
}

func TestGiftsDeactivate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/gifts/gift_123/deactivate" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		gift := Gift{ID: "gift_123", Status: "inactive"}
		_ = json.NewEncoder(w).Encode(gift)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	gift, err := client.DeactivateGift(context.Background(), "gift_123")
	if err != nil {
		t.Fatalf("DeactivateGift failed: %v", err)
	}

	if gift.Status != "inactive" {
		t.Errorf("Expected inactive status, got: %s", gift.Status)
	}
}

func TestGiftsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/gifts/gift_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req["title"] != "Updated Gift" {
			t.Errorf("Expected title 'Updated Gift', got %v", req["title"])
		}

		gift := Gift{ID: "gift_123", Title: "Updated Gift", Status: "active"}
		_ = json.NewEncoder(w).Encode(gift)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &GiftUpdateRequest{
		Title: "Updated Gift",
	}
	gift, err := client.UpdateGift(context.Background(), "gift_123", req)
	if err != nil {
		t.Fatalf("UpdateGift failed: %v", err)
	}

	if gift.Title != "Updated Gift" {
		t.Errorf("Unexpected gift title: %s", gift.Title)
	}
}

func TestUpdateGiftEmptyID(t *testing.T) {
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
			_, err := client.UpdateGift(context.Background(), tc.id, &GiftUpdateRequest{})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "gift id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestGiftsSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/gifts/search" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("query") != "sample" {
			t.Errorf("Expected query=sample, got %s", r.URL.Query().Get("query"))
		}
		if r.URL.Query().Get("status") != "active" {
			t.Errorf("Expected status=active, got %s", r.URL.Query().Get("status"))
		}

		resp := GiftsListResponse{
			Items: []Gift{
				{ID: "gift_123", Title: "Free Sample", Status: "active"},
			},
			Page:       1,
			PageSize:   20,
			TotalCount: 1,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	gifts, err := client.SearchGifts(context.Background(), &GiftSearchOptions{
		Query:  "sample",
		Status: "active",
	})
	if err != nil {
		t.Fatalf("SearchGifts failed: %v", err)
	}

	if len(gifts.Items) != 1 {
		t.Errorf("Expected 1 gift, got %d", len(gifts.Items))
	}
	if gifts.Items[0].ID != "gift_123" {
		t.Errorf("Unexpected gift ID: %s", gifts.Items[0].ID)
	}
}

func TestGiftsUpdateQuantity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/gifts/gift_123/update_quantity" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if int(req["quantity"].(float64)) != 100 {
			t.Errorf("Expected quantity 100, got %v", req["quantity"])
		}

		gift := Gift{ID: "gift_123", Quantity: 100}
		_ = json.NewEncoder(w).Encode(gift)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	gift, err := client.UpdateGiftQuantity(context.Background(), "gift_123", 100)
	if err != nil {
		t.Fatalf("UpdateGiftQuantity failed: %v", err)
	}

	if gift.Quantity != 100 {
		t.Errorf("Unexpected gift quantity: %d", gift.Quantity)
	}
}

func TestUpdateGiftQuantityEmptyID(t *testing.T) {
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
			_, err := client.UpdateGiftQuantity(context.Background(), tc.id, 100)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "gift id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestGiftsUpdateQuantityBySKU(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/gifts/update_quantity" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req["sku"] != "SKU-123" {
			t.Errorf("Expected sku 'SKU-123', got %v", req["sku"])
		}
		if int(req["quantity"].(float64)) != 50 {
			t.Errorf("Expected quantity 50, got %v", req["quantity"])
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.UpdateGiftsQuantityBySKU(context.Background(), "SKU-123", 50)
	if err != nil {
		t.Fatalf("UpdateGiftsQuantityBySKU failed: %v", err)
	}
}

func TestGiftsGetStocks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/gifts/gift_123/stocks" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.GetGiftStocks(context.Background(), "gift_123")
	if err != nil {
		t.Fatalf("GetGiftStocks failed: %v", err)
	}
	if len(raw) == 0 {
		t.Fatalf("expected response body")
	}
}

func TestGiftsUpdateStocks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/gifts/gift_123/stocks" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body["ok"] != true {
			t.Fatalf("expected ok=true, got %v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"updated": true})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.UpdateGiftStocks(context.Background(), "gift_123", map[string]any{"ok": true})
	if err != nil {
		t.Fatalf("UpdateGiftStocks failed: %v", err)
	}
	if len(raw) == 0 {
		t.Fatalf("expected response body")
	}
}

func TestUpdateGiftsQuantityBySKUEmptySKU(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name string
		sku  string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.UpdateGiftsQuantityBySKU(context.Background(), tc.sku, 100)
			if err == nil {
				t.Error("Expected error for empty SKU, got nil")
			}
			if err != nil && err.Error() != "sku is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
