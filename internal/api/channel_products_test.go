package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListChannelProductListings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/channels/ch_123/product_listings" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := ChannelProductsListResponse{
			Items: []ChannelProductListing{
				{ID: "cpl_123", ProductID: "prod_123", ChannelID: "ch_123", Title: "Product 1", Published: true},
				{ID: "cpl_456", ProductID: "prod_456", ChannelID: "ch_123", Title: "Product 2", Published: false},
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

	listings, err := client.ListChannelProductListings(context.Background(), "ch_123", nil)
	if err != nil {
		t.Fatalf("ListChannelProductListings failed: %v", err)
	}

	if len(listings.Items) != 2 {
		t.Errorf("Expected 2 listings, got %d", len(listings.Items))
	}
	if listings.Items[0].ID != "cpl_123" {
		t.Errorf("Unexpected listing ID: %s", listings.Items[0].ID)
	}
}

func TestListChannelProductListingsWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("published") != "true" {
			t.Errorf("Expected published=true, got %s", r.URL.Query().Get("published"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}

		resp := ChannelProductsListResponse{
			Items:      []ChannelProductListing{},
			Page:       2,
			PageSize:   20,
			TotalCount: 0,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	published := true
	opts := &ChannelProductsListOptions{
		Page:      2,
		Published: &published,
	}
	_, err := client.ListChannelProductListings(context.Background(), "ch_123", opts)
	if err != nil {
		t.Fatalf("ListChannelProductListings failed: %v", err)
	}
}

func TestListChannelProductListingsEmptyChannelID(t *testing.T) {
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
			_, err := client.ListChannelProductListings(context.Background(), tc.id, nil)
			if err == nil {
				t.Error("Expected error for empty channel ID, got nil")
			}
			if err != nil && err.Error() != "channel id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestGetChannelProductListing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/ch_123/product_listings/prod_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		listing := ChannelProductListing{
			ID:        "cpl_123",
			ProductID: "prod_123",
			ChannelID: "ch_123",
			Title:     "Test Product",
			Published: true,
		}
		_ = json.NewEncoder(w).Encode(listing)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	listing, err := client.GetChannelProductListing(context.Background(), "ch_123", "prod_123")
	if err != nil {
		t.Fatalf("GetChannelProductListing failed: %v", err)
	}

	if listing.ID != "cpl_123" {
		t.Errorf("Unexpected listing ID: %s", listing.ID)
	}
}

func TestGetChannelProductListingEmptyIDs(t *testing.T) {
	client := NewClient("token")

	// Test empty channel ID
	_, err := client.GetChannelProductListing(context.Background(), "", "prod_123")
	if err == nil || err.Error() != "channel id is required" {
		t.Errorf("Expected 'channel id is required' error, got: %v", err)
	}

	// Test empty product ID
	_, err = client.GetChannelProductListing(context.Background(), "ch_123", "")
	if err == nil || err.Error() != "product id is required" {
		t.Errorf("Expected 'product id is required' error, got: %v", err)
	}
}

func TestPublishProductToChannelListing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/channels/ch_123/product_listings" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		listing := ChannelProductListing{
			ID:        "cpl_123",
			ProductID: "prod_123",
			ChannelID: "ch_123",
			Published: true,
		}
		_ = json.NewEncoder(w).Encode(listing)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ChannelProductPublishRequest{
		ProductID: "prod_123",
	}
	listing, err := client.PublishProductToChannelListing(context.Background(), "ch_123", req)
	if err != nil {
		t.Fatalf("PublishProductToChannelListing failed: %v", err)
	}

	if listing.ProductID != "prod_123" {
		t.Errorf("Unexpected product ID: %s", listing.ProductID)
	}
}

func TestUpdateChannelProductListing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/channels/ch_123/product_listings/prod_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		listing := ChannelProductListing{
			ID:        "cpl_123",
			ProductID: "prod_123",
			ChannelID: "ch_123",
			Published: false,
		}
		_ = json.NewEncoder(w).Encode(listing)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	published := false
	req := &ChannelProductUpdateRequest{
		Published: &published,
	}
	listing, err := client.UpdateChannelProductListing(context.Background(), "ch_123", "prod_123", req)
	if err != nil {
		t.Fatalf("UpdateChannelProductListing failed: %v", err)
	}

	if listing.Published != false {
		t.Errorf("Expected published=false, got %t", listing.Published)
	}
}

func TestUnpublishProductFromChannelListing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/channels/ch_123/product_listings/prod_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.UnpublishProductFromChannelListing(context.Background(), "ch_123", "prod_123")
	if err != nil {
		t.Fatalf("UnpublishProductFromChannelListing failed: %v", err)
	}
}
