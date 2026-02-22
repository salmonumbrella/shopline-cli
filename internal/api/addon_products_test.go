package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddonProductsUpdateQuantity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/addon_products/ap_123/update_quantity" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req AddonProductQuantityRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}
		if req.Quantity != 7 {
			t.Errorf("Expected quantity 7, got %d", req.Quantity)
		}

		_ = json.NewEncoder(w).Encode(AddonProduct{ID: "ap_123", Quantity: req.Quantity})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	ap, err := client.UpdateAddonProductQuantity(context.Background(), "ap_123", &AddonProductQuantityRequest{Quantity: 7})
	if err != nil {
		t.Fatalf("UpdateAddonProductQuantity failed: %v", err)
	}
	if ap.ID != "ap_123" {
		t.Errorf("Unexpected addon product ID: %s", ap.ID)
	}
	if ap.Quantity != 7 {
		t.Errorf("Unexpected quantity: %d", ap.Quantity)
	}
}

func TestAddonProductsBulkUpdateQuantityBySKU(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/addon_products/update_quantity" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req AddonProductQuantityBySKURequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}
		if req.SKU != "SKU-1" || req.Quantity != 9 {
			t.Errorf("Unexpected request: sku=%q quantity=%d", req.SKU, req.Quantity)
		}

		_ = json.NewEncoder(w).Encode(AddonProduct{ID: "ap_bulk"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.UpdateAddonProductsQuantityBySKU(context.Background(), &AddonProductQuantityBySKURequest{SKU: "SKU-1", Quantity: 9})
	if err != nil {
		t.Fatalf("UpdateAddonProductsQuantityBySKU failed: %v", err)
	}
}

func TestAddonProductsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/addon_products" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := AddonProductsListResponse{
			Items: []AddonProduct{
				{
					ID:        "ap_123",
					Title:     "Extended Warranty",
					ProductID: "prod_456",
					Price:     "29.99",
					Currency:  "USD",
					Status:    "active",
				},
				{
					ID:        "ap_456",
					Title:     "Gift Wrapping",
					ProductID: "prod_789",
					Price:     "5.99",
					Currency:  "USD",
					Status:    "active",
				},
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

	addonProducts, err := client.ListAddonProducts(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListAddonProducts failed: %v", err)
	}

	if len(addonProducts.Items) != 2 {
		t.Errorf("Expected 2 addon products, got %d", len(addonProducts.Items))
	}
	if addonProducts.Items[0].ID != "ap_123" {
		t.Errorf("Unexpected addon product ID: %s", addonProducts.Items[0].ID)
	}
	if addonProducts.Items[0].Title != "Extended Warranty" {
		t.Errorf("Unexpected title: %s", addonProducts.Items[0].Title)
	}
}

func TestAddonProductsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("product_id") != "prod_123" {
			t.Errorf("Expected product_id=prod_123, got %s", r.URL.Query().Get("product_id"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("status") != "active" {
			t.Errorf("Expected status=active, got %s", r.URL.Query().Get("status"))
		}

		resp := AddonProductsListResponse{
			Items: []AddonProduct{
				{ID: "ap_123", ProductID: "prod_123", Status: "active"},
			},
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

	opts := &AddonProductsListOptions{
		Page:      2,
		ProductID: "prod_123",
		Status:    "active",
	}
	addonProducts, err := client.ListAddonProducts(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListAddonProducts failed: %v", err)
	}

	if len(addonProducts.Items) != 1 {
		t.Errorf("Expected 1 addon product, got %d", len(addonProducts.Items))
	}
}

func TestAddonProductsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/addon_products/ap_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		addonProduct := AddonProduct{
			ID:          "ap_123",
			Title:       "Extended Warranty",
			ProductID:   "prod_456",
			VariantID:   "var_789",
			Price:       "29.99",
			Currency:    "USD",
			Quantity:    1,
			Position:    1,
			Status:      "active",
			Description: "3-year extended warranty",
		}
		_ = json.NewEncoder(w).Encode(addonProduct)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	addonProduct, err := client.GetAddonProduct(context.Background(), "ap_123")
	if err != nil {
		t.Fatalf("GetAddonProduct failed: %v", err)
	}

	if addonProduct.ID != "ap_123" {
		t.Errorf("Unexpected addon product ID: %s", addonProduct.ID)
	}
	if addonProduct.Title != "Extended Warranty" {
		t.Errorf("Unexpected title: %s", addonProduct.Title)
	}
	if addonProduct.Price != "29.99" {
		t.Errorf("Unexpected price: %s", addonProduct.Price)
	}
}

func TestAddonProductsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/addon_products" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req AddonProductCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Title != "Gift Wrapping" {
			t.Errorf("Unexpected title: %s", req.Title)
		}
		if req.ProductID != "prod_123" {
			t.Errorf("Unexpected product ID: %s", req.ProductID)
		}

		addonProduct := AddonProduct{
			ID:        "ap_new",
			Title:     req.Title,
			ProductID: req.ProductID,
			Price:     req.Price,
			Status:    "active",
		}
		_ = json.NewEncoder(w).Encode(addonProduct)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &AddonProductCreateRequest{
		Title:     "Gift Wrapping",
		ProductID: "prod_123",
		Price:     "5.99",
	}
	addonProduct, err := client.CreateAddonProduct(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateAddonProduct failed: %v", err)
	}

	if addonProduct.ID != "ap_new" {
		t.Errorf("Unexpected addon product ID: %s", addonProduct.ID)
	}
	if addonProduct.Title != "Gift Wrapping" {
		t.Errorf("Unexpected title: %s", addonProduct.Title)
	}
}

func TestAddonProductsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/addon_products/ap_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteAddonProduct(context.Background(), "ap_123")
	if err != nil {
		t.Fatalf("DeleteAddonProduct failed: %v", err)
	}
}

func TestGetAddonProductEmptyID(t *testing.T) {
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
			_, err := client.GetAddonProduct(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "addon product id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteAddonProductEmptyID(t *testing.T) {
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
			err := client.DeleteAddonProduct(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "addon product id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
