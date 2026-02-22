package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFlashPriceList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/flash-price-campaigns" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := FlashPriceListResponse{
			Items: []FlashPrice{
				{ID: "fp_123", ProductID: "prod_1", FlashPrice: 99.99, Status: "active"},
				{ID: "fp_456", ProductID: "prod_2", FlashPrice: 49.99, Status: "scheduled"},
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

	flashPrices, err := client.ListFlashPrices(context.Background(), &FlashPriceListOptions{})
	if err != nil {
		t.Fatalf("ListFlashPrices failed: %v", err)
	}

	if len(flashPrices.Items) != 2 {
		t.Errorf("Expected 2 flash prices, got %d", len(flashPrices.Items))
	}
	if flashPrices.Items[0].ID != "fp_123" {
		t.Errorf("Unexpected flash price ID: %s", flashPrices.Items[0].ID)
	}
}

func TestFlashPriceListWithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("product_id") != "prod_123" {
			t.Errorf("Expected product_id=prod_123, got %s", r.URL.Query().Get("product_id"))
		}
		if r.URL.Query().Get("status") != "active" {
			t.Errorf("Expected status=active, got %s", r.URL.Query().Get("status"))
		}

		resp := FlashPriceListResponse{Items: []FlashPrice{}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.ListFlashPrices(context.Background(), &FlashPriceListOptions{
		ProductID: "prod_123",
		Status:    "active",
	})
	if err != nil {
		t.Fatalf("ListFlashPrices failed: %v", err)
	}
}

func TestFlashPriceGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/flash-price-campaigns/fp_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		flashPrice := FlashPrice{ID: "fp_123", ProductID: "prod_1", FlashPrice: 99.99}
		_ = json.NewEncoder(w).Encode(flashPrice)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	flashPrice, err := client.GetFlashPrice(context.Background(), "fp_123")
	if err != nil {
		t.Fatalf("GetFlashPrice failed: %v", err)
	}

	if flashPrice.ID != "fp_123" {
		t.Errorf("Unexpected flash price ID: %s", flashPrice.ID)
	}
}

func TestGetFlashPriceEmptyID(t *testing.T) {
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
			_, err := client.GetFlashPrice(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "flash price campaign id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestFlashPriceCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/flash-price-campaigns" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		flashPrice := FlashPrice{ID: "fp_new", ProductID: "prod_1", FlashPrice: 79.99}
		_ = json.NewEncoder(w).Encode(flashPrice)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &FlashPriceCreateRequest{
		ProductID:  "prod_1",
		FlashPrice: 79.99,
		Quantity:   100,
	}

	flashPrice, err := client.CreateFlashPrice(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateFlashPrice failed: %v", err)
	}

	if flashPrice.ProductID != "prod_1" {
		t.Errorf("Unexpected product ID: %s", flashPrice.ProductID)
	}
}

func TestFlashPriceDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/flash-price-campaigns/fp_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteFlashPrice(context.Background(), "fp_123")
	if err != nil {
		t.Fatalf("DeleteFlashPrice failed: %v", err)
	}
}

func TestDeleteFlashPriceEmptyID(t *testing.T) {
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
			err := client.DeleteFlashPrice(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "flash price campaign id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestFlashPriceActivate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/flash-price-campaigns/fp_123/activate" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		flashPrice := FlashPrice{ID: "fp_123", Status: "active"}
		_ = json.NewEncoder(w).Encode(flashPrice)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	flashPrice, err := client.ActivateFlashPrice(context.Background(), "fp_123")
	if err != nil {
		t.Fatalf("ActivateFlashPrice failed: %v", err)
	}

	if flashPrice.Status != "active" {
		t.Errorf("Expected active status, got: %s", flashPrice.Status)
	}
}

func TestFlashPriceDeactivate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/flash-price-campaigns/fp_123/deactivate" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		flashPrice := FlashPrice{ID: "fp_123", Status: "inactive"}
		_ = json.NewEncoder(w).Encode(flashPrice)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	flashPrice, err := client.DeactivateFlashPrice(context.Background(), "fp_123")
	if err != nil {
		t.Fatalf("DeactivateFlashPrice failed: %v", err)
	}

	if flashPrice.Status != "inactive" {
		t.Errorf("Expected inactive status, got: %s", flashPrice.Status)
	}
}

func TestFlashPriceUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/flash-price-campaigns/fp_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		flashPrice := FlashPrice{ID: "fp_123", ProductID: "prod_1", FlashPrice: 89.99, Status: "active"}
		_ = json.NewEncoder(w).Encode(flashPrice)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	newPrice := 89.99
	req := &FlashPriceUpdateRequest{
		FlashPrice: &newPrice,
	}

	flashPrice, err := client.UpdateFlashPrice(context.Background(), "fp_123", req)
	if err != nil {
		t.Fatalf("UpdateFlashPrice failed: %v", err)
	}

	if flashPrice.FlashPrice != 89.99 {
		t.Errorf("Unexpected flash price: %v", flashPrice.FlashPrice)
	}
}

func TestUpdateFlashPriceEmptyID(t *testing.T) {
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
			_, err := client.UpdateFlashPrice(context.Background(), tc.id, &FlashPriceUpdateRequest{})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "flash price campaign id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
