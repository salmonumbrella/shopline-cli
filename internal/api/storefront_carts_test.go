package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStorefrontCartsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/storefront/carts" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := StorefrontCartsListResponse{
			Items: []StorefrontCart{
				{ID: "cart_123", CustomerID: "cust_1", TotalPrice: "99.99", ItemCount: 3},
				{ID: "cart_456", CustomerID: "cust_2", TotalPrice: "149.99", ItemCount: 5},
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

	carts, err := client.ListStorefrontCarts(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListStorefrontCarts failed: %v", err)
	}

	if len(carts.Items) != 2 {
		t.Errorf("Expected 2 carts, got %d", len(carts.Items))
	}
	if carts.Items[0].ID != "cart_123" {
		t.Errorf("Unexpected cart ID: %s", carts.Items[0].ID)
	}
}

func TestStorefrontCartsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("status") != "abandoned" {
			t.Errorf("Expected status=abandoned, got %s", r.URL.Query().Get("status"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}

		resp := StorefrontCartsListResponse{
			Items:      []StorefrontCart{{ID: "cart_123", CustomerID: "cust_1"}},
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

	opts := &StorefrontCartsListOptions{
		Page:   2,
		Status: "abandoned",
	}
	carts, err := client.ListStorefrontCarts(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListStorefrontCarts failed: %v", err)
	}

	if len(carts.Items) != 1 {
		t.Errorf("Expected 1 cart, got %d", len(carts.Items))
	}
}

func TestStorefrontCartsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/storefront/carts/cart_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		cart := StorefrontCart{
			ID:         "cart_123",
			CustomerID: "cust_1",
			TotalPrice: "99.99",
			ItemCount:  3,
		}
		_ = json.NewEncoder(w).Encode(cart)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	cart, err := client.GetStorefrontCart(context.Background(), "cart_123")
	if err != nil {
		t.Fatalf("GetStorefrontCart failed: %v", err)
	}

	if cart.ID != "cart_123" {
		t.Errorf("Unexpected cart ID: %s", cart.ID)
	}
	if cart.TotalPrice != "99.99" {
		t.Errorf("Unexpected total price: %s", cart.TotalPrice)
	}
}

func TestStorefrontCartsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/storefront/carts" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req StorefrontCartCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.CustomerID != "cust_1" {
			t.Errorf("Unexpected customer ID: %s", req.CustomerID)
		}

		cart := StorefrontCart{
			ID:         "cart_new",
			CustomerID: req.CustomerID,
			TotalPrice: "0.00",
			ItemCount:  0,
		}
		_ = json.NewEncoder(w).Encode(cart)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &StorefrontCartCreateRequest{
		CustomerID: "cust_1",
	}
	cart, err := client.CreateStorefrontCart(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateStorefrontCart failed: %v", err)
	}

	if cart.ID != "cart_new" {
		t.Errorf("Unexpected cart ID: %s", cart.ID)
	}
}

func TestStorefrontCartsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/storefront/carts/cart_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteStorefrontCart(context.Background(), "cart_123")
	if err != nil {
		t.Fatalf("DeleteStorefrontCart failed: %v", err)
	}
}

func TestGetStorefrontCartEmptyID(t *testing.T) {
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
			_, err := client.GetStorefrontCart(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "cart id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteStorefrontCartEmptyID(t *testing.T) {
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
			err := client.DeleteStorefrontCart(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "cart id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
