package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSalesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/sales" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := SalesListResponse{
			Items: []Sale{
				{ID: "sale_123", Title: "Summer Sale", DiscountType: "percentage", DiscountValue: 20, Status: "active"},
				{ID: "sale_456", Title: "Black Friday", DiscountType: "fixed_amount", DiscountValue: 50, Status: "scheduled"},
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

	sales, err := client.ListSales(context.Background(), &SalesListOptions{})
	if err != nil {
		t.Fatalf("ListSales failed: %v", err)
	}

	if len(sales.Items) != 2 {
		t.Errorf("Expected 2 sales, got %d", len(sales.Items))
	}
	if sales.Items[0].ID != "sale_123" {
		t.Errorf("Unexpected sale ID: %s", sales.Items[0].ID)
	}
}

func TestSalesListWithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("status") != "active" {
			t.Errorf("Expected status=active, got %s", r.URL.Query().Get("status"))
		}

		resp := SalesListResponse{Items: []Sale{}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.ListSales(context.Background(), &SalesListOptions{
		Status: "active",
	})
	if err != nil {
		t.Fatalf("ListSales failed: %v", err)
	}
}

func TestSalesGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sales/sale_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		sale := Sale{ID: "sale_123", Title: "Summer Sale", DiscountType: "percentage", DiscountValue: 20}
		_ = json.NewEncoder(w).Encode(sale)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	sale, err := client.GetSale(context.Background(), "sale_123")
	if err != nil {
		t.Fatalf("GetSale failed: %v", err)
	}

	if sale.ID != "sale_123" {
		t.Errorf("Unexpected sale ID: %s", sale.ID)
	}
}

func TestGetSaleEmptyID(t *testing.T) {
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
			_, err := client.GetSale(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "sale id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestSalesCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/sales" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		sale := Sale{ID: "sale_new", Title: "New Sale", DiscountType: "percentage", DiscountValue: 15}
		_ = json.NewEncoder(w).Encode(sale)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &SaleCreateRequest{
		Title:         "New Sale",
		DiscountType:  "percentage",
		DiscountValue: 15,
		AppliesTo:     "all",
	}

	sale, err := client.CreateSale(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateSale failed: %v", err)
	}

	if sale.Title != "New Sale" {
		t.Errorf("Unexpected sale title: %s", sale.Title)
	}
}

func TestSalesDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/sales/sale_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteSale(context.Background(), "sale_123")
	if err != nil {
		t.Fatalf("DeleteSale failed: %v", err)
	}
}

func TestDeleteSaleEmptyID(t *testing.T) {
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
			err := client.DeleteSale(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "sale id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestSalesActivate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/sales/sale_123/activate" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		sale := Sale{ID: "sale_123", Status: "active"}
		_ = json.NewEncoder(w).Encode(sale)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	sale, err := client.ActivateSale(context.Background(), "sale_123")
	if err != nil {
		t.Fatalf("ActivateSale failed: %v", err)
	}

	if sale.Status != "active" {
		t.Errorf("Expected active status, got: %s", sale.Status)
	}
}

func TestSalesDeactivate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/sales/sale_123/deactivate" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		sale := Sale{ID: "sale_123", Status: "inactive"}
		_ = json.NewEncoder(w).Encode(sale)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	sale, err := client.DeactivateSale(context.Background(), "sale_123")
	if err != nil {
		t.Fatalf("DeactivateSale failed: %v", err)
	}

	if sale.Status != "inactive" {
		t.Errorf("Expected inactive status, got: %s", sale.Status)
	}
}

func TestDeleteSaleProducts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/sales/sale_123/delete_products" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req SaleDeleteProductsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if len(req.ProductIDs) != 2 || req.ProductIDs[0] != "prod_1" || req.ProductIDs[1] != "prod_2" {
			t.Fatalf("unexpected product ids: %+v", req.ProductIDs)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteSaleProducts(context.Background(), "sale_123", &SaleDeleteProductsRequest{ProductIDs: []string{"prod_1", "prod_2"}})
	if err != nil {
		t.Fatalf("DeleteSaleProducts failed: %v", err)
	}
}

func TestDeleteSaleProductsEmptyID(t *testing.T) {
	client := NewClient("token")
	client.SetUseOpenAPI(false)

	err := client.DeleteSaleProducts(context.Background(), " ", &SaleDeleteProductsRequest{ProductIDs: []string{"prod_1"}})
	if err == nil {
		t.Fatalf("expected error for empty sale id")
	}
}

func TestDeleteSaleProductsEmptyProductIDs(t *testing.T) {
	client := NewClient("token")
	client.SetUseOpenAPI(false)

	err := client.DeleteSaleProducts(context.Background(), "sale_123", &SaleDeleteProductsRequest{ProductIDs: nil})
	if err == nil {
		t.Fatalf("expected error for empty product ids")
	}
}
