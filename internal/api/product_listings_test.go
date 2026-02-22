package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProductListingsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/product_listings" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := ProductListingsListResponse{
			Items: []ProductListing{
				{ID: "pl_123", ProductID: "prod_123", Title: "Widget", Available: true},
				{ID: "pl_456", ProductID: "prod_456", Title: "Gadget", Available: false},
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

	listings, err := client.ListProductListings(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListProductListings failed: %v", err)
	}

	if len(listings.Items) != 2 {
		t.Errorf("Expected 2 product listings, got %d", len(listings.Items))
	}
	if listings.Items[0].ID != "pl_123" {
		t.Errorf("Unexpected product listing ID: %s", listings.Items[0].ID)
	}
	if listings.Items[0].ProductID != "prod_123" {
		t.Errorf("Unexpected product ID: %s", listings.Items[0].ProductID)
	}
}

func TestProductListingsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("page_size") != "10" {
			t.Errorf("Expected page_size=10, got %s", r.URL.Query().Get("page_size"))
		}

		resp := ProductListingsListResponse{
			Items: []ProductListing{
				{ID: "pl_789", ProductID: "prod_789", Title: "Sprocket"},
			},
			Page:       2,
			PageSize:   10,
			TotalCount: 21,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &ProductListingsListOptions{
		Page:     2,
		PageSize: 10,
	}
	listings, err := client.ListProductListings(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListProductListings failed: %v", err)
	}

	if len(listings.Items) != 1 {
		t.Errorf("Expected 1 product listing, got %d", len(listings.Items))
	}
	if listings.Page != 2 {
		t.Errorf("Expected page 2, got %d", listings.Page)
	}
}

func TestProductListingsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/product_listings/pl_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		listing := ProductListing{
			ID:          "pl_123",
			ProductID:   "prod_123",
			Title:       "Widget",
			Handle:      "widget",
			BodyHTML:    "<p>A great widget</p>",
			Vendor:      "Acme",
			ProductType: "Widgets",
			Available:   true,
		}
		_ = json.NewEncoder(w).Encode(listing)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	listing, err := client.GetProductListing(context.Background(), "pl_123")
	if err != nil {
		t.Fatalf("GetProductListing failed: %v", err)
	}

	if listing.ID != "pl_123" {
		t.Errorf("Unexpected product listing ID: %s", listing.ID)
	}
	if listing.ProductID != "prod_123" {
		t.Errorf("Unexpected product ID: %s", listing.ProductID)
	}
	if listing.Title != "Widget" {
		t.Errorf("Unexpected title: %s", listing.Title)
	}
	if !listing.Available {
		t.Error("Expected listing to be available")
	}
}

func TestGetProductListingEmptyID(t *testing.T) {
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
			_, err := client.GetProductListing(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "product listing id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestProductListingsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/product_listings" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req struct {
			ProductID string `json:"product_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.ProductID != "prod_123" {
			t.Errorf("Unexpected product ID: %s", req.ProductID)
		}

		listing := ProductListing{
			ID:        "pl_new",
			ProductID: req.ProductID,
			Title:     "Widget",
			Available: true,
		}
		_ = json.NewEncoder(w).Encode(listing)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	listing, err := client.CreateProductListing(context.Background(), "prod_123")
	if err != nil {
		t.Fatalf("CreateProductListing failed: %v", err)
	}

	if listing.ID != "pl_new" {
		t.Errorf("Unexpected product listing ID: %s", listing.ID)
	}
	if listing.ProductID != "prod_123" {
		t.Errorf("Unexpected product ID: %s", listing.ProductID)
	}
}

func TestProductListingsCreateEmptyProductID(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name      string
		productID string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.CreateProductListing(context.Background(), tc.productID)
			if err == nil {
				t.Error("Expected error for empty product ID, got nil")
			}
			if err != nil && err.Error() != "product id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestProductListingsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/product_listings/pl_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteProductListing(context.Background(), "pl_123")
	if err != nil {
		t.Fatalf("DeleteProductListing failed: %v", err)
	}
}

func TestDeleteProductListingEmptyID(t *testing.T) {
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
			err := client.DeleteProductListing(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "product listing id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
