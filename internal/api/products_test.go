package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProductsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/products" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := ProductsListResponse{
			Items: []Product{
				{ID: "prod_123", Title: "Widget", Status: "active", Price: &Price{Cents: 2999, Label: "29.99"}},
				{ID: "prod_456", Title: "Gadget", Status: "draft", Price: &Price{Cents: 4999, Label: "49.99"}},
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

	products, err := client.ListProducts(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListProducts failed: %v", err)
	}

	if len(products.Items) != 2 {
		t.Errorf("Expected 2 products, got %d", len(products.Items))
	}
	if products.Items[0].ID != "prod_123" {
		t.Errorf("Unexpected product ID: %s", products.Items[0].ID)
	}
}

func TestProductsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/products/prod_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		product := Product{ID: "prod_123", Title: "Widget", Status: "active", Price: &Price{Cents: 2999, Label: "29.99"}}
		_ = json.NewEncoder(w).Encode(product)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	product, err := client.GetProduct(context.Background(), "prod_123")
	if err != nil {
		t.Fatalf("GetProduct failed: %v", err)
	}

	if product.ID != "prod_123" {
		t.Errorf("Unexpected product ID: %s", product.ID)
	}
}

func TestGetProductEmptyID(t *testing.T) {
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
			_, err := client.GetProduct(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "product id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestProductsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", query.Get("page"))
		}
		if query.Get("page_size") != "50" {
			t.Errorf("Expected page_size=50, got %s", query.Get("page_size"))
		}
		if query.Get("status") != "active" {
			t.Errorf("Expected status=active, got %s", query.Get("status"))
		}
		if query.Get("vendor") != "TestVendor" {
			t.Errorf("Expected vendor=TestVendor, got %s", query.Get("vendor"))
		}
		if query.Get("product_type") != "widgets" {
			t.Errorf("Expected product_type=widgets, got %s", query.Get("product_type"))
		}
		if query.Get("sort_by") != "title" {
			t.Errorf("Expected sort_by=title, got %s", query.Get("sort_by"))
		}
		if query.Get("sort_order") != "asc" {
			t.Errorf("Expected sort_order=asc, got %s", query.Get("sort_order"))
		}

		resp := ProductsListResponse{
			Items:      []Product{{ID: "prod_123", Title: "Widget"}},
			Page:       2,
			PageSize:   50,
			TotalCount: 100,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &ProductsListOptions{
		Page:        2,
		PageSize:    50,
		Status:      "active",
		Vendor:      "TestVendor",
		ProductType: "widgets",
		SortBy:      "title",
		SortOrder:   "asc",
	}
	products, err := client.ListProducts(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListProducts failed: %v", err)
	}

	if len(products.Items) != 1 {
		t.Errorf("Expected 1 product, got %d", len(products.Items))
	}
}

func TestProductsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	// Test List API error
	_, err := client.ListProducts(context.Background(), nil)
	if err == nil {
		t.Error("Expected error from ListProducts")
	}

	// Test Get API error
	_, err = client.GetProduct(context.Background(), "prod_123")
	if err == nil {
		t.Error("Expected error from GetProduct")
	}
}

func TestSearchProducts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/products/search" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		query := r.URL.Query()
		if query.Get("query") != "widget" {
			t.Errorf("Expected query=widget, got %s", query.Get("query"))
		}
		if query.Get("status") != "active" {
			t.Errorf("Expected status=active, got %s", query.Get("status"))
		}
		if query.Get("vendor") != "TestVendor" {
			t.Errorf("Expected vendor=TestVendor, got %s", query.Get("vendor"))
		}
		if query.Get("page") != "1" {
			t.Errorf("Expected page=1, got %s", query.Get("page"))
		}
		if query.Get("page_size") != "25" {
			t.Errorf("Expected page_size=25, got %s", query.Get("page_size"))
		}

		resp := ProductsListResponse{
			Items: []Product{
				{ID: "prod_123", Title: "Widget Pro", Status: "active"},
			},
			Page:       1,
			PageSize:   25,
			TotalCount: 1,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &ProductSearchOptions{
		Query:    "widget",
		Status:   "active",
		Vendor:   "TestVendor",
		Page:     1,
		PageSize: 25,
	}
	products, err := client.SearchProducts(context.Background(), opts)
	if err != nil {
		t.Fatalf("SearchProducts failed: %v", err)
	}

	if len(products.Items) != 1 {
		t.Errorf("Expected 1 product, got %d", len(products.Items))
	}
	if products.Items[0].ID != "prod_123" {
		t.Errorf("Unexpected product ID: %s", products.Items[0].ID)
	}
}

func TestSearchProductsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &ProductSearchOptions{Query: "test"}
	_, err := client.SearchProducts(context.Background(), opts)
	if err == nil {
		t.Error("Expected error from SearchProducts")
	}
}

func TestCreateProduct(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/products" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ProductCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.Title != "New Product" {
			t.Errorf("Expected title=New Product, got %s", req.Title)
		}
		if req.Description != "A great product" {
			t.Errorf("Expected description=A great product, got %s", req.Description)
		}
		if req.Vendor != "TestVendor" {
			t.Errorf("Expected vendor=TestVendor, got %s", req.Vendor)
		}
		if req.ProductType != "widgets" {
			t.Errorf("Expected product_type=widgets, got %s", req.ProductType)
		}
		if len(req.Tags) != 2 || req.Tags[0] != "new" || req.Tags[1] != "sale" {
			t.Errorf("Unexpected tags: %v", req.Tags)
		}
		if req.Status != "draft" {
			t.Errorf("Expected status=draft, got %s", req.Status)
		}

		product := Product{ID: "prod_new", Title: "New Product", Status: "draft"}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(product)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ProductCreateRequest{
		Title:       "New Product",
		Description: "A great product",
		Vendor:      "TestVendor",
		ProductType: "widgets",
		Tags:        []string{"new", "sale"},
		Status:      "draft",
	}
	product, err := client.CreateProduct(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateProduct failed: %v", err)
	}

	if product.ID != "prod_new" {
		t.Errorf("Unexpected product ID: %s", product.ID)
	}
}

func TestCreateProductAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ProductCreateRequest{Title: "Test"}
	_, err := client.CreateProduct(context.Background(), req)
	if err == nil {
		t.Error("Expected error from CreateProduct")
	}
}

func TestUpdateProduct(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/products/prod_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ProductUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.Title == nil || *req.Title != "Updated Title" {
			t.Errorf("Expected title=Updated Title, got %v", req.Title)
		}
		if req.Description == nil || *req.Description != "Updated description" {
			t.Errorf("Expected description=Updated description, got %v", req.Description)
		}

		product := Product{ID: "prod_123", Title: "Updated Title", Description: "Updated description"}
		_ = json.NewEncoder(w).Encode(product)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	title := "Updated Title"
	description := "Updated description"
	req := &ProductUpdateRequest{
		Title:       &title,
		Description: &description,
	}
	product, err := client.UpdateProduct(context.Background(), "prod_123", req)
	if err != nil {
		t.Fatalf("UpdateProduct failed: %v", err)
	}

	if product.Title != "Updated Title" {
		t.Errorf("Unexpected product title: %s", product.Title)
	}
}

func TestUpdateProductEmptyID(t *testing.T) {
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
			_, err := client.UpdateProduct(context.Background(), tc.id, &ProductUpdateRequest{})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "product id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestUpdateProductAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.UpdateProduct(context.Background(), "prod_123", &ProductUpdateRequest{})
	if err == nil {
		t.Error("Expected error from UpdateProduct")
	}
}

func TestDeleteProduct(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/products/prod_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteProduct(context.Background(), "prod_123")
	if err != nil {
		t.Fatalf("DeleteProduct failed: %v", err)
	}
}

func TestDeleteProductEmptyID(t *testing.T) {
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
			err := client.DeleteProduct(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "product id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteProductAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteProduct(context.Background(), "prod_123")
	if err == nil {
		t.Error("Expected error from DeleteProduct")
	}
}

func TestBulkDeleteProducts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/products/bulk" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ProductBulkDeleteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		if len(req.ProductIDs) != 2 || req.ProductIDs[0] != "prod_1" || req.ProductIDs[1] != "prod_2" {
			t.Errorf("Unexpected product IDs: %v", req.ProductIDs)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	if err := client.BulkDeleteProducts(context.Background(), []string{"prod_1", "prod_2"}); err != nil {
		t.Fatalf("BulkDeleteProducts failed: %v", err)
	}
}

func TestUpdateProductQuantity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/products/prod_123/update_quantity" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ProductQuantityUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.Quantity != 100 {
			t.Errorf("Expected quantity=100, got %d", req.Quantity)
		}

		product := Product{ID: "prod_123", Title: "Widget"}
		_ = json.NewEncoder(w).Encode(product)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	product, err := client.UpdateProductQuantity(context.Background(), "prod_123", 100)
	if err != nil {
		t.Fatalf("UpdateProductQuantity failed: %v", err)
	}

	if product.ID != "prod_123" {
		t.Errorf("Unexpected product ID: %s", product.ID)
	}
}

func TestUpdateProductQuantityEmptyID(t *testing.T) {
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
			_, err := client.UpdateProductQuantity(context.Background(), tc.id, 10)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "product id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestUpdateProductQuantityAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.UpdateProductQuantity(context.Background(), "prod_123", 100)
	if err == nil {
		t.Error("Expected error from UpdateProductQuantity")
	}
}

func TestUpdateProductVariationQuantity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/products/prod_123/variations/var_456/update_quantity" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ProductQuantityUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.Quantity != 50 {
			t.Errorf("Expected quantity=50, got %d", req.Quantity)
		}

		product := Product{ID: "prod_123", Title: "Widget"}
		_ = json.NewEncoder(w).Encode(product)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	product, err := client.UpdateProductVariationQuantity(context.Background(), "prod_123", "var_456", 50)
	if err != nil {
		t.Fatalf("UpdateProductVariationQuantity failed: %v", err)
	}

	if product.ID != "prod_123" {
		t.Errorf("Unexpected product ID: %s", product.ID)
	}
}

func TestUpdateProductVariationQuantityEmptyIDs(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name        string
		productID   string
		variationID string
		expectedErr string
	}{
		{"empty product ID", "", "var_456", "product id is required"},
		{"whitespace product ID", "   ", "var_456", "product id is required"},
		{"empty variation ID", "prod_123", "", "variation id is required"},
		{"whitespace variation ID", "prod_123", "   ", "variation id is required"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.UpdateProductVariationQuantity(context.Background(), tc.productID, tc.variationID, 10)
			if err == nil {
				t.Error("Expected error, got nil")
			}
			if err != nil && err.Error() != tc.expectedErr {
				t.Errorf("Expected error '%s', got '%s'", tc.expectedErr, err.Error())
			}
		})
	}
}

func TestUpdateProductVariationQuantityAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.UpdateProductVariationQuantity(context.Background(), "prod_123", "var_456", 50)
	if err == nil {
		t.Error("Expected error from UpdateProductVariationQuantity")
	}
}

func TestUpdateProductPrice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/products/prod_123/update_price" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ProductPriceUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.Price != 29.99 {
			t.Errorf("Expected price=29.99, got %f", req.Price)
		}

		product := Product{ID: "prod_123", Title: "Widget", Price: &Price{Cents: 2999, Label: "29.99"}}
		_ = json.NewEncoder(w).Encode(product)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	product, err := client.UpdateProductPrice(context.Background(), "prod_123", 29.99)
	if err != nil {
		t.Fatalf("UpdateProductPrice failed: %v", err)
	}

	if product.ID != "prod_123" {
		t.Errorf("Unexpected product ID: %s", product.ID)
	}
}

func TestUpdateProductPriceEmptyID(t *testing.T) {
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
			_, err := client.UpdateProductPrice(context.Background(), tc.id, 10.00)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "product id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestUpdateProductPriceAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.UpdateProductPrice(context.Background(), "prod_123", 29.99)
	if err == nil {
		t.Error("Expected error from UpdateProductPrice")
	}
}

func TestUpdateProductVariationPrice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/products/prod_123/variations/var_456/update_price" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ProductPriceUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.Price != 39.99 {
			t.Errorf("Expected price=39.99, got %f", req.Price)
		}

		product := Product{ID: "prod_123", Title: "Widget"}
		_ = json.NewEncoder(w).Encode(product)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	product, err := client.UpdateProductVariationPrice(context.Background(), "prod_123", "var_456", 39.99)
	if err != nil {
		t.Fatalf("UpdateProductVariationPrice failed: %v", err)
	}

	if product.ID != "prod_123" {
		t.Errorf("Unexpected product ID: %s", product.ID)
	}
}

func TestUpdateProductVariationPriceEmptyIDs(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name        string
		productID   string
		variationID string
		expectedErr string
	}{
		{"empty product ID", "", "var_456", "product id is required"},
		{"whitespace product ID", "   ", "var_456", "product id is required"},
		{"empty variation ID", "prod_123", "", "variation id is required"},
		{"whitespace variation ID", "prod_123", "   ", "variation id is required"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.UpdateProductVariationPrice(context.Background(), tc.productID, tc.variationID, 10.00)
			if err == nil {
				t.Error("Expected error, got nil")
			}
			if err != nil && err.Error() != tc.expectedErr {
				t.Errorf("Expected error '%s', got '%s'", tc.expectedErr, err.Error())
			}
		})
	}
}

func TestUpdateProductVariationPriceAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.UpdateProductVariationPrice(context.Background(), "prod_123", "var_456", 39.99)
	if err == nil {
		t.Error("Expected error from UpdateProductVariationPrice")
	}
}

func TestUpdateProductQuantityBySKU(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/products/update_quantity" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ProductQuantityBySKURequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.SKU != "SKU-12345" {
			t.Errorf("Expected sku=SKU-12345, got %s", req.SKU)
		}
		if req.Quantity != 75 {
			t.Errorf("Expected quantity=75, got %d", req.Quantity)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.UpdateProductQuantityBySKU(context.Background(), "SKU-12345", 75)
	if err != nil {
		t.Fatalf("UpdateProductQuantityBySKU failed: %v", err)
	}
}

func TestUpdateProductQuantityBySKUEmptySKU(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name string
		sku  string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab only", "\t"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.UpdateProductQuantityBySKU(context.Background(), tc.sku, 10)
			if err == nil {
				t.Error("Expected error for empty SKU, got nil")
			}
			if err != nil && err.Error() != "sku is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestUpdateProductQuantityBySKUAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.UpdateProductQuantityBySKU(context.Background(), "SKU-12345", 75)
	if err == nil {
		t.Error("Expected error from UpdateProductQuantityBySKU")
	}
}

func TestSearchProductsPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/products/search" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ProductSearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		if req.Query != "shoes" {
			t.Errorf("Expected query=shoes, got %q", req.Query)
		}

		resp := ProductsListResponse{Items: []Product{{ID: "prod_1", Title: "Shoe"}}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	resp, err := client.SearchProductsPost(context.Background(), &ProductSearchRequest{Query: "shoes"})
	if err != nil {
		t.Fatalf("SearchProductsPost failed: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].ID != "prod_1" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestAddProductImages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/products/prod_123/add_images" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ProductAddImagesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if len(req.Images) != 2 {
			t.Errorf("Expected 2 images, got %d", len(req.Images))
		}
		if req.Images[0].Src != "https://example.com/image1.jpg" {
			t.Errorf("Unexpected image src: %s", req.Images[0].Src)
		}
		if req.Images[0].Position != 1 {
			t.Errorf("Expected position=1, got %d", req.Images[0].Position)
		}

		images := []ProductImage{
			{ID: "img_1", ProductID: "prod_123", Src: "https://example.com/image1.jpg", Position: 1},
			{ID: "img_2", ProductID: "prod_123", Src: "https://example.com/image2.jpg", Position: 2},
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(images)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ProductAddImagesRequest{
		Images: []ProductImageInput{
			{Src: "https://example.com/image1.jpg", Position: 1},
			{Src: "https://example.com/image2.jpg", Position: 2},
		},
	}
	images, err := client.AddProductImages(context.Background(), "prod_123", req)
	if err != nil {
		t.Fatalf("AddProductImages failed: %v", err)
	}

	if len(images) != 2 {
		t.Errorf("Expected 2 images, got %d", len(images))
	}
	if images[0].ID != "img_1" {
		t.Errorf("Unexpected image ID: %s", images[0].ID)
	}
}

func TestAddProductImagesEmptyID(t *testing.T) {
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
			req := &ProductAddImagesRequest{Images: []ProductImageInput{{Src: "test.jpg"}}}
			_, err := client.AddProductImages(context.Background(), tc.id, req)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "product id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestAddProductImagesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ProductAddImagesRequest{Images: []ProductImageInput{{Src: "test.jpg"}}}
	_, err := client.AddProductImages(context.Background(), "prod_123", req)
	if err == nil {
		t.Error("Expected error from AddProductImages")
	}
}

func TestDeleteProductImages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/products/prod_123/delete_images" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ProductDeleteImagesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if len(req.ImageIDs) != 2 {
			t.Errorf("Expected 2 image IDs, got %d", len(req.ImageIDs))
		}
		if req.ImageIDs[0] != "img_1" || req.ImageIDs[1] != "img_2" {
			t.Errorf("Unexpected image IDs: %v", req.ImageIDs)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteProductImages(context.Background(), "prod_123", []string{"img_1", "img_2"})
	if err != nil {
		t.Fatalf("DeleteProductImages failed: %v", err)
	}
}

func TestDeleteProductImagesEmptyID(t *testing.T) {
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
			err := client.DeleteProductImages(context.Background(), tc.id, []string{"img_1"})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "product id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteProductImagesEmptyImageIDs(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name     string
		imageIDs []string
	}{
		{"nil slice", nil},
		{"empty slice", []string{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.DeleteProductImages(context.Background(), "prod_123", tc.imageIDs)
			if err == nil {
				t.Error("Expected error for empty image IDs, got nil")
			}
			if err != nil && err.Error() != "at least one image id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteProductImagesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteProductImages(context.Background(), "prod_123", []string{"img_1"})
	if err == nil {
		t.Error("Expected error from DeleteProductImages")
	}
}

func TestReplaceProductTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/products/prod_123/tags" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ProductTagsReplaceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		if len(req.Tags) != 2 || req.Tags[0] != "tag1" || req.Tags[1] != "tag2" {
			t.Errorf("Unexpected tags: %v", req.Tags)
		}

		product := Product{ID: "prod_123", Title: "Widget", Tags: req.Tags}
		_ = json.NewEncoder(w).Encode(product)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	product, err := client.ReplaceProductTags(context.Background(), "prod_123", []string{"tag1", "tag2"})
	if err != nil {
		t.Fatalf("ReplaceProductTags failed: %v", err)
	}
	if len(product.Tags) != 2 {
		t.Fatalf("unexpected tags: %+v", product.Tags)
	}
}

func TestAddProductVariation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/products/prod_123/variations" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ProductVariationCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.Title != "Large" {
			t.Errorf("Expected title=Large, got %s", req.Title)
		}
		if req.SKU != "SKU-LARGE" {
			t.Errorf("Expected sku=SKU-LARGE, got %s", req.SKU)
		}
		if req.Price != 34.99 {
			t.Errorf("Expected price=34.99, got %f", req.Price)
		}
		if req.Quantity != 50 {
			t.Errorf("Expected quantity=50, got %d", req.Quantity)
		}

		variation := ProductVariation{
			ID:        "var_new",
			ProductID: "prod_123",
			Title:     "Large",
			SKU:       "SKU-LARGE",
			Quantity:  50,
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(variation)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ProductVariationCreateRequest{
		Title:    "Large",
		SKU:      "SKU-LARGE",
		Price:    34.99,
		Quantity: 50,
	}
	variation, err := client.AddProductVariation(context.Background(), "prod_123", req)
	if err != nil {
		t.Fatalf("AddProductVariation failed: %v", err)
	}

	if variation.ID != "var_new" {
		t.Errorf("Unexpected variation ID: %s", variation.ID)
	}
	if variation.Title != "Large" {
		t.Errorf("Unexpected variation title: %s", variation.Title)
	}
}

func TestAddProductVariationEmptyID(t *testing.T) {
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
			req := &ProductVariationCreateRequest{Title: "Test"}
			_, err := client.AddProductVariation(context.Background(), tc.id, req)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "product id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestAddProductVariationAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ProductVariationCreateRequest{Title: "Test"}
	_, err := client.AddProductVariation(context.Background(), "prod_123", req)
	if err == nil {
		t.Error("Expected error from AddProductVariation")
	}
}

func TestUpdateProductVariation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/products/prod_123/variations/var_456" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ProductVariationUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.Title == nil || *req.Title != "Extra Large" {
			t.Errorf("Expected title=Extra Large, got %v", req.Title)
		}
		if req.SKU == nil || *req.SKU != "SKU-XL" {
			t.Errorf("Expected sku=SKU-XL, got %v", req.SKU)
		}

		variation := ProductVariation{
			ID:        "var_456",
			ProductID: "prod_123",
			Title:     "Extra Large",
			SKU:       "SKU-XL",
		}
		_ = json.NewEncoder(w).Encode(variation)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	title := "Extra Large"
	sku := "SKU-XL"
	req := &ProductVariationUpdateRequest{
		Title: &title,
		SKU:   &sku,
	}
	variation, err := client.UpdateProductVariation(context.Background(), "prod_123", "var_456", req)
	if err != nil {
		t.Fatalf("UpdateProductVariation failed: %v", err)
	}

	if variation.Title != "Extra Large" {
		t.Errorf("Unexpected variation title: %s", variation.Title)
	}
}

func TestUpdateProductVariationEmptyIDs(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name        string
		productID   string
		variationID string
		expectedErr string
	}{
		{"empty product ID", "", "var_456", "product id is required"},
		{"whitespace product ID", "   ", "var_456", "product id is required"},
		{"empty variation ID", "prod_123", "", "variation id is required"},
		{"whitespace variation ID", "prod_123", "   ", "variation id is required"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.UpdateProductVariation(context.Background(), tc.productID, tc.variationID, &ProductVariationUpdateRequest{})
			if err == nil {
				t.Error("Expected error, got nil")
			}
			if err != nil && err.Error() != tc.expectedErr {
				t.Errorf("Expected error '%s', got '%s'", tc.expectedErr, err.Error())
			}
		})
	}
}

func TestUpdateProductVariationAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.UpdateProductVariation(context.Background(), "prod_123", "var_456", &ProductVariationUpdateRequest{})
	if err == nil {
		t.Error("Expected error from UpdateProductVariation")
	}
}

func TestDeleteProductVariation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/products/prod_123/variations/var_456" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteProductVariation(context.Background(), "prod_123", "var_456")
	if err != nil {
		t.Fatalf("DeleteProductVariation failed: %v", err)
	}
}

func TestDeleteProductVariationEmptyIDs(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name        string
		productID   string
		variationID string
		expectedErr string
	}{
		{"empty product ID", "", "var_456", "product id is required"},
		{"whitespace product ID", "   ", "var_456", "product id is required"},
		{"empty variation ID", "prod_123", "", "variation id is required"},
		{"whitespace variation ID", "prod_123", "   ", "variation id is required"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.DeleteProductVariation(context.Background(), tc.productID, tc.variationID)
			if err == nil {
				t.Error("Expected error, got nil")
			}
			if err != nil && err.Error() != tc.expectedErr {
				t.Errorf("Expected error '%s', got '%s'", tc.expectedErr, err.Error())
			}
		})
	}
}

func TestDeleteProductVariationAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteProductVariation(context.Background(), "prod_123", "var_456")
	if err == nil {
		t.Error("Expected error from DeleteProductVariation")
	}
}

func TestUpdateProductTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("Expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/products/prod_123/tags" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ProductTagsUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if len(req.Add) != 2 || req.Add[0] != "new-tag" || req.Add[1] != "sale" {
			t.Errorf("Unexpected add tags: %v", req.Add)
		}
		if len(req.Remove) != 1 || req.Remove[0] != "old-tag" {
			t.Errorf("Unexpected remove tags: %v", req.Remove)
		}

		product := Product{ID: "prod_123", Title: "Widget", Tags: []string{"new-tag", "sale"}}
		_ = json.NewEncoder(w).Encode(product)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ProductTagsUpdateRequest{
		Add:    []string{"new-tag", "sale"},
		Remove: []string{"old-tag"},
	}
	product, err := client.UpdateProductTags(context.Background(), "prod_123", req)
	if err != nil {
		t.Fatalf("UpdateProductTags failed: %v", err)
	}

	if len(product.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(product.Tags))
	}
}

func TestUpdateProductTagsEmptyID(t *testing.T) {
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
			_, err := client.UpdateProductTags(context.Background(), tc.id, &ProductTagsUpdateRequest{})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "product id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestUpdateProductTagsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ProductTagsUpdateRequest{Add: []string{"tag1"}}
	_, err := client.UpdateProductTags(context.Background(), "prod_123", req)
	if err == nil {
		t.Error("Expected error from UpdateProductTags")
	}
}

func TestBulkAssignProductsToCategory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/categories/bulk_assign" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ProductCategoryAssignRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if len(req.ProductIDs) != 3 {
			t.Errorf("Expected 3 product IDs, got %d", len(req.ProductIDs))
		}
		if req.ProductIDs[0] != "prod_1" || req.ProductIDs[1] != "prod_2" || req.ProductIDs[2] != "prod_3" {
			t.Errorf("Unexpected product IDs: %v", req.ProductIDs)
		}
		if req.CategoryID != "cat_123" {
			t.Errorf("Expected category_id=cat_123, got %s", req.CategoryID)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ProductCategoryAssignRequest{
		ProductIDs: []string{"prod_1", "prod_2", "prod_3"},
		CategoryID: "cat_123",
	}
	err := client.BulkAssignProductsToCategory(context.Background(), req)
	if err != nil {
		t.Fatalf("BulkAssignProductsToCategory failed: %v", err)
	}
}

func TestBulkAssignProductsToCategoryEmptyProductIDs(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name       string
		productIDs []string
	}{
		{"nil slice", nil},
		{"empty slice", []string{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &ProductCategoryAssignRequest{ProductIDs: tc.productIDs, CategoryID: "cat_123"}
			err := client.BulkAssignProductsToCategory(context.Background(), req)
			if err == nil {
				t.Error("Expected error for empty product IDs, got nil")
			}
			if err != nil && err.Error() != "at least one product id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestBulkAssignProductsToCategoryEmptyCategoryID(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name       string
		categoryID string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab only", "\t"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &ProductCategoryAssignRequest{ProductIDs: []string{"prod_1"}, CategoryID: tc.categoryID}
			err := client.BulkAssignProductsToCategory(context.Background(), req)
			if err == nil {
				t.Error("Expected error for empty category ID, got nil")
			}
			if err != nil && err.Error() != "category id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestBulkAssignProductsToCategoryAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ProductCategoryAssignRequest{ProductIDs: []string{"prod_1"}, CategoryID: "cat_123"}
	err := client.BulkAssignProductsToCategory(context.Background(), req)
	if err == nil {
		t.Error("Expected error from BulkAssignProductsToCategory")
	}
}

func TestGetLockedInventoryCount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/products/locked_inventory_count" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		count := LockedInventoryCount{
			ProductID:   "prod_123",
			VariationID: "var_456",
			LockedCount: 10,
		}
		_ = json.NewEncoder(w).Encode(count)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	count, err := client.GetLockedInventoryCount(context.Background(), "prod_123")
	if err != nil {
		t.Fatalf("GetLockedInventoryCount failed: %v", err)
	}

	if count.ProductID != "prod_123" {
		t.Errorf("Unexpected product ID: %s", count.ProductID)
	}
	if count.VariationID != "var_456" {
		t.Errorf("Unexpected variation ID: %s", count.VariationID)
	}
	if count.LockedCount != 10 {
		t.Errorf("Expected locked count=10, got %d", count.LockedCount)
	}
}

func TestGetLockedInventoryCountEmptyID(t *testing.T) {
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
			_, err := client.GetLockedInventoryCount(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "product id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestGetLockedInventoryCountAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.GetLockedInventoryCount(context.Background(), "prod_123")
	if err == nil {
		t.Error("Expected error from GetLockedInventoryCount")
	}
}

func TestGetProductPromotions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/products/prod_123/promotions" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[{"id":"promo_1"}]}`))
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.GetProductPromotions(context.Background(), "prod_123")
	if err != nil {
		t.Fatalf("GetProductPromotions failed: %v", err)
	}
	if len(raw) == 0 {
		t.Fatalf("expected non-empty response")
	}
}

func TestGetProductStocks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/products/prod_123/stocks" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[{"location_id":"loc_1","available":3}]}`))
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.GetProductStocks(context.Background(), "prod_123")
	if err != nil {
		t.Fatalf("GetProductStocks failed: %v", err)
	}
	if len(raw) == 0 {
		t.Fatalf("expected non-empty response")
	}
}

func TestUpdateProductStocks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/products/prod_123/stocks" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.UpdateProductStocks(context.Background(), "prod_123", map[string]any{"items": []any{}})
	if err != nil {
		t.Fatalf("UpdateProductStocks failed: %v", err)
	}
	if len(raw) == 0 {
		t.Fatalf("expected non-empty response")
	}
}

func TestBulkUpdateProductStocks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/products/bulk_update_stocks" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	if err := client.BulkUpdateProductStocks(context.Background(), map[string]any{"items": []any{}}); err != nil {
		t.Fatalf("BulkUpdateProductStocks failed: %v", err)
	}
}

func TestUpdateProductsStatusBulk(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/products/status/bulk" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	if err := client.UpdateProductsStatusBulk(context.Background(), map[string]any{"product_ids": []string{"prod_1"}, "status": "active"}); err != nil {
		t.Fatalf("UpdateProductsStatusBulk failed: %v", err)
	}
}

func TestUpdateProductsRetailStatusBulk(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/products/retail_status/bulk" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	if err := client.UpdateProductsRetailStatusBulk(context.Background(), map[string]any{"product_ids": []string{"prod_1"}, "status": "active"}); err != nil {
		t.Fatalf("UpdateProductsRetailStatusBulk failed: %v", err)
	}
}

func TestUpdateProductsLabelsBulk(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("Expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/products/labels" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	if err := client.UpdateProductsLabelsBulk(context.Background(), map[string]any{"items": []any{}}); err != nil {
		t.Fatalf("UpdateProductsLabelsBulk failed: %v", err)
	}
}
