package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStorefrontProductsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/storefront/products" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := StorefrontProductsListResponse{
			Items: []StorefrontProduct{
				{ID: "prod_123", Title: "Widget", Price: "29.99", Available: true},
				{ID: "prod_456", Title: "Gadget", Price: "49.99", Available: true},
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

	products, err := client.ListStorefrontProducts(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListStorefrontProducts failed: %v", err)
	}

	if len(products.Items) != 2 {
		t.Errorf("Expected 2 products, got %d", len(products.Items))
	}
	if products.Items[0].ID != "prod_123" {
		t.Errorf("Unexpected product ID: %s", products.Items[0].ID)
	}
}

func TestStorefrontProductsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("vendor") != "Acme" {
			t.Errorf("Expected vendor=Acme, got %s", r.URL.Query().Get("vendor"))
		}
		if r.URL.Query().Get("available") != "true" {
			t.Errorf("Expected available=true, got %s", r.URL.Query().Get("available"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}

		resp := StorefrontProductsListResponse{
			Items:      []StorefrontProduct{{ID: "prod_123", Title: "Widget", Vendor: "Acme"}},
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

	available := true
	opts := &StorefrontProductsListOptions{
		Page:      2,
		Vendor:    "Acme",
		Available: &available,
	}
	products, err := client.ListStorefrontProducts(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListStorefrontProducts failed: %v", err)
	}

	if len(products.Items) != 1 {
		t.Errorf("Expected 1 product, got %d", len(products.Items))
	}
}

func TestStorefrontProductsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/storefront/products/prod_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		product := StorefrontProduct{
			ID:        "prod_123",
			Title:     "Widget",
			Handle:    "widget",
			Price:     "29.99",
			Available: true,
			ViewCount: 150,
		}
		_ = json.NewEncoder(w).Encode(product)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	product, err := client.GetStorefrontProduct(context.Background(), "prod_123")
	if err != nil {
		t.Fatalf("GetStorefrontProduct failed: %v", err)
	}

	if product.ID != "prod_123" {
		t.Errorf("Unexpected product ID: %s", product.ID)
	}
	if product.Price != "29.99" {
		t.Errorf("Unexpected price: %s", product.Price)
	}
	if product.ViewCount != 150 {
		t.Errorf("Unexpected view count: %d", product.ViewCount)
	}
}

func TestStorefrontProductsGetByHandle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/storefront/products/handle/widget" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		product := StorefrontProduct{
			ID:     "prod_123",
			Title:  "Widget",
			Handle: "widget",
			Price:  "29.99",
		}
		_ = json.NewEncoder(w).Encode(product)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	product, err := client.GetStorefrontProductByHandle(context.Background(), "widget")
	if err != nil {
		t.Fatalf("GetStorefrontProductByHandle failed: %v", err)
	}

	if product.Handle != "widget" {
		t.Errorf("Unexpected product handle: %s", product.Handle)
	}
}

func TestGetStorefrontProductEmptyID(t *testing.T) {
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
			_, err := client.GetStorefrontProduct(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "product id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestGetStorefrontProductByHandleEmpty(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name   string
		handle string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab only", "\t"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.GetStorefrontProductByHandle(context.Background(), tc.handle)
			if err == nil {
				t.Error("Expected error for empty handle, got nil")
			}
			if err != nil && err.Error() != "product handle is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
