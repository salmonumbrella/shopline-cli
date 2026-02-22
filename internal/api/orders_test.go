package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOrdersList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/orders" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := OrdersListResponse{
			Items: []OrderSummary{
				{ID: "ord_123", Status: "pending", TotalPrice: "99.99"},
				{ID: "ord_456", Status: "completed", TotalPrice: "149.99"},
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

	orders, err := client.ListOrders(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListOrders failed: %v", err)
	}

	if len(orders.Items) != 2 {
		t.Errorf("Expected 2 orders, got %d", len(orders.Items))
	}
	if orders.Items[0].ID != "ord_123" {
		t.Errorf("Unexpected order ID: %s", orders.Items[0].ID)
	}
}

func TestOrdersGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orders/ord_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		order := Order{ID: "ord_123", Status: "pending", TotalPrice: "99.99"}
		_ = json.NewEncoder(w).Encode(order)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	order, err := client.GetOrder(context.Background(), "ord_123")
	if err != nil {
		t.Fatalf("GetOrder failed: %v", err)
	}

	if order.ID != "ord_123" {
		t.Errorf("Unexpected order ID: %s", order.ID)
	}
}

func TestGetOrderEmptyID(t *testing.T) {
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
			_, err := client.GetOrder(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCancelOrderEmptyID(t *testing.T) {
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
			err := client.CancelOrder(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestOrdersListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", query.Get("page"))
		}
		if query.Get("page_size") != "50" {
			t.Errorf("Expected page_size=50, got %s", query.Get("page_size"))
		}
		if query.Get("status") != "pending" {
			t.Errorf("Expected status=pending, got %s", query.Get("status"))
		}
		if query.Get("sort_by") != "created_at" {
			t.Errorf("Expected sort_by=created_at, got %s", query.Get("sort_by"))
		}
		if query.Get("sort_order") != "desc" {
			t.Errorf("Expected sort_order=desc, got %s", query.Get("sort_order"))
		}

		resp := OrdersListResponse{
			Items:      []OrderSummary{{ID: "ord_123", Status: "pending"}},
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

	opts := &OrdersListOptions{
		Page:      2,
		PageSize:  50,
		Status:    "pending",
		SortBy:    "created_at",
		SortOrder: "desc",
	}
	orders, err := client.ListOrders(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListOrders failed: %v", err)
	}

	if len(orders.Items) != 1 {
		t.Errorf("Expected 1 order, got %d", len(orders.Items))
	}
}

func TestCancelOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("Expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/orders/ord_123/cancel" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.CancelOrder(context.Background(), "ord_123")
	if err != nil {
		t.Fatalf("CancelOrder failed: %v", err)
	}
}

func TestOrdersAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	// Test List API error
	_, err := client.ListOrders(context.Background(), nil)
	if err == nil {
		t.Error("Expected error from ListOrders")
	}

	// Test Get API error
	_, err = client.GetOrder(context.Background(), "ord_123")
	if err == nil {
		t.Error("Expected error from GetOrder")
	}

	// Test Cancel API error
	err = client.CancelOrder(context.Background(), "ord_123")
	if err == nil {
		t.Error("Expected error from CancelOrder")
	}
}

// Tests for SearchOrders

func TestSearchOrders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/orders/search" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		query := r.URL.Query()
		if query.Get("query") != "test@example.com" {
			t.Errorf("Expected query=test@example.com, got %s", query.Get("query"))
		}
		if query.Get("status") != "pending" {
			t.Errorf("Expected status=pending, got %s", query.Get("status"))
		}
		if query.Get("page") != "1" {
			t.Errorf("Expected page=1, got %s", query.Get("page"))
		}
		if query.Get("page_size") != "25" {
			t.Errorf("Expected page_size=25, got %s", query.Get("page_size"))
		}

		resp := OrdersListResponse{
			Items: []OrderSummary{
				{ID: "ord_123", Status: "pending", CustomerEmail: "test@example.com"},
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

	opts := &OrderSearchOptions{
		Query:    "test@example.com",
		Status:   "pending",
		Page:     1,
		PageSize: 25,
	}
	orders, err := client.SearchOrders(context.Background(), opts)
	if err != nil {
		t.Fatalf("SearchOrders failed: %v", err)
	}

	if len(orders.Items) != 1 {
		t.Errorf("Expected 1 order, got %d", len(orders.Items))
	}
	if orders.Items[0].ID != "ord_123" {
		t.Errorf("Unexpected order ID: %s", orders.Items[0].ID)
	}
}

func TestSearchOrdersAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.SearchOrders(context.Background(), &OrderSearchOptions{Query: "test"})
	if err == nil {
		t.Error("Expected error from SearchOrders")
	}
}

// Tests for ListArchivedOrders

func TestListArchivedOrders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/orders/archived" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		query := r.URL.Query()
		if query.Get("page") != "1" {
			t.Errorf("Expected page=1, got %s", query.Get("page"))
		}
		if query.Get("page_size") != "20" {
			t.Errorf("Expected page_size=20, got %s", query.Get("page_size"))
		}

		resp := OrdersListResponse{
			Items: []OrderSummary{
				{ID: "ord_archived_1", Status: "archived"},
				{ID: "ord_archived_2", Status: "archived"},
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

	opts := &ArchivedOrdersListOptions{
		Page:     1,
		PageSize: 20,
	}
	orders, err := client.ListArchivedOrders(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListArchivedOrders failed: %v", err)
	}

	if len(orders.Items) != 2 {
		t.Errorf("Expected 2 orders, got %d", len(orders.Items))
	}
	if orders.Items[0].ID != "ord_archived_1" {
		t.Errorf("Unexpected order ID: %s", orders.Items[0].ID)
	}
}

func TestCreateArchivedOrdersReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/orders/archived_orders" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	resp, err := client.CreateArchivedOrdersReport(context.Background(), json.RawMessage(`{"from":"2026-01-01"}`))
	if err != nil {
		t.Fatalf("CreateArchivedOrdersReport failed: %v", err)
	}
	if string(resp) == "" {
		t.Fatalf("expected response, got empty")
	}
}

func TestListArchivedOrdersAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.ListArchivedOrders(context.Background(), &ArchivedOrdersListOptions{})
	if err == nil {
		t.Error("Expected error from ListArchivedOrders")
	}
}

// Tests for CreateOrder

func TestCreateOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/orders" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req OrderCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.CustomerEmail != "customer@example.com" {
			t.Errorf("Expected CustomerEmail=customer@example.com, got %s", req.CustomerEmail)
		}
		if len(req.LineItems) != 2 {
			t.Errorf("Expected 2 line items, got %d", len(req.LineItems))
		}
		if req.LineItems[0].ProductID != "prod_1" {
			t.Errorf("Expected first line item product_id=prod_1, got %s", req.LineItems[0].ProductID)
		}
		if req.LineItems[0].Quantity != 2 {
			t.Errorf("Expected first line item quantity=2, got %d", req.LineItems[0].Quantity)
		}

		order := Order{
			ID:            "ord_new_123",
			OrderNumber:   "ORD-001",
			Status:        "pending",
			CustomerEmail: req.CustomerEmail,
			TotalPrice:    "199.98",
		}
		_ = json.NewEncoder(w).Encode(order)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &OrderCreateRequest{
		CustomerEmail: "customer@example.com",
		LineItems: []OrderItem{
			{ProductID: "prod_1", Quantity: 2, Price: 49.99},
			{ProductID: "prod_2", VariationID: "var_1", Quantity: 1, Price: 99.99},
		},
		Note: "Test order",
	}
	order, err := client.CreateOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	if order.ID != "ord_new_123" {
		t.Errorf("Unexpected order ID: %s", order.ID)
	}
	if order.Status != "pending" {
		t.Errorf("Expected status=pending, got %s", order.Status)
	}
}

func TestCreateOrderAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.CreateOrder(context.Background(), &OrderCreateRequest{
		LineItems: []OrderItem{{ProductID: "prod_1", Quantity: 1}},
	})
	if err == nil {
		t.Error("Expected error from CreateOrder")
	}
}

// Tests for UpdateOrder

func TestUpdateOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("Expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/orders/ord_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req OrderUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.Note == nil || *req.Note != "Updated note" {
			t.Errorf("Expected Note='Updated note', got %v", req.Note)
		}

		order := Order{
			ID:     "ord_123",
			Status: "pending",
		}
		_ = json.NewEncoder(w).Encode(order)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	note := "Updated note"
	req := &OrderUpdateRequest{
		Note: &note,
	}
	order, err := client.UpdateOrder(context.Background(), "ord_123", req)
	if err != nil {
		t.Fatalf("UpdateOrder failed: %v", err)
	}

	if order.ID != "ord_123" {
		t.Errorf("Unexpected order ID: %s", order.ID)
	}
}

func TestUpdateOrderEmptyID(t *testing.T) {
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
			_, err := client.UpdateOrder(context.Background(), tc.id, &OrderUpdateRequest{})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestUpdateOrderAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.UpdateOrder(context.Background(), "ord_123", &OrderUpdateRequest{})
	if err == nil {
		t.Error("Expected error from UpdateOrder")
	}
}

// Tests for UpdateOrderStatus

func TestUpdateOrderStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("Expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/orders/ord_123/status" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req OrderStatusUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.Status != "completed" {
			t.Errorf("Expected Status='completed', got %s", req.Status)
		}

		order := Order{
			ID:     "ord_123",
			Status: "completed",
		}
		_ = json.NewEncoder(w).Encode(order)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	order, err := client.UpdateOrderStatus(context.Background(), "ord_123", "completed")
	if err != nil {
		t.Fatalf("UpdateOrderStatus failed: %v", err)
	}

	if order.Status != "completed" {
		t.Errorf("Expected status=completed, got %s", order.Status)
	}
}

func TestUpdateOrderStatusEmptyID(t *testing.T) {
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
			_, err := client.UpdateOrderStatus(context.Background(), tc.id, "completed")
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestUpdateOrderStatusAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.UpdateOrderStatus(context.Background(), "ord_123", "completed")
	if err == nil {
		t.Error("Expected error from UpdateOrderStatus")
	}
}

// Tests for UpdateOrderDeliveryStatus

func TestUpdateOrderDeliveryStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("Expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/orders/ord_123/order_delivery_status" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req OrderDeliveryStatusUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.DeliveryStatus != "shipped" {
			t.Errorf("Expected DeliveryStatus='shipped', got %s", req.DeliveryStatus)
		}

		order := Order{
			ID:            "ord_123",
			FulfillStatus: "shipped",
		}
		_ = json.NewEncoder(w).Encode(order)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	order, err := client.UpdateOrderDeliveryStatus(context.Background(), "ord_123", "shipped")
	if err != nil {
		t.Fatalf("UpdateOrderDeliveryStatus failed: %v", err)
	}

	if order.FulfillStatus != "shipped" {
		t.Errorf("Expected fulfill_status=shipped, got %s", order.FulfillStatus)
	}
}

func TestUpdateOrderDeliveryStatusEmptyID(t *testing.T) {
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
			_, err := client.UpdateOrderDeliveryStatus(context.Background(), tc.id, "shipped")
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestUpdateOrderDeliveryStatusAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.UpdateOrderDeliveryStatus(context.Background(), "ord_123", "shipped")
	if err == nil {
		t.Error("Expected error from UpdateOrderDeliveryStatus")
	}
}

// Tests for UpdateOrderPaymentStatus

func TestUpdateOrderPaymentStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("Expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/orders/ord_123/order_payment_status" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req OrderPaymentStatusUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.PaymentStatus != "paid" {
			t.Errorf("Expected PaymentStatus='paid', got %s", req.PaymentStatus)
		}

		order := Order{
			ID:            "ord_123",
			PaymentStatus: "paid",
		}
		_ = json.NewEncoder(w).Encode(order)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	order, err := client.UpdateOrderPaymentStatus(context.Background(), "ord_123", "paid")
	if err != nil {
		t.Fatalf("UpdateOrderPaymentStatus failed: %v", err)
	}

	if order.PaymentStatus != "paid" {
		t.Errorf("Expected payment_status=paid, got %s", order.PaymentStatus)
	}
}

func TestUpdateOrderPaymentStatusEmptyID(t *testing.T) {
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
			_, err := client.UpdateOrderPaymentStatus(context.Background(), tc.id, "paid")
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestUpdateOrderPaymentStatusAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.UpdateOrderPaymentStatus(context.Background(), "ord_123", "paid")
	if err == nil {
		t.Error("Expected error from UpdateOrderPaymentStatus")
	}
}

// Tests for GetOrderTags

func TestGetOrderTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/orders/ord_123/tags" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := OrderTagsResponse{
			Tags: []string{"vip", "rush", "gift"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	tags, err := client.GetOrderTags(context.Background(), "ord_123")
	if err != nil {
		t.Fatalf("GetOrderTags failed: %v", err)
	}

	if len(tags.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(tags.Tags))
	}
	if tags.Tags[0] != "vip" {
		t.Errorf("Expected first tag='vip', got %s", tags.Tags[0])
	}
}

func TestGetOrderTagsEmptyID(t *testing.T) {
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
			_, err := client.GetOrderTags(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestGetOrderTagsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.GetOrderTags(context.Background(), "ord_123")
	if err == nil {
		t.Error("Expected error from GetOrderTags")
	}
}

// Tests for UpdateOrderTags

func TestUpdateOrderTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/orders/ord_123/tags" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req OrderTagsUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if len(req.Tags) != 2 {
			t.Errorf("Expected 2 tags, got %d", len(req.Tags))
		}
		if req.Tags[0] != "priority" {
			t.Errorf("Expected first tag='priority', got %s", req.Tags[0])
		}

		order := Order{
			ID:     "ord_123",
			Status: "pending",
		}
		_ = json.NewEncoder(w).Encode(order)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	order, err := client.UpdateOrderTags(context.Background(), "ord_123", []string{"priority", "express"})
	if err != nil {
		t.Fatalf("UpdateOrderTags failed: %v", err)
	}

	if order.ID != "ord_123" {
		t.Errorf("Unexpected order ID: %s", order.ID)
	}
}

func TestUpdateOrderTagsEmptyID(t *testing.T) {
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
			_, err := client.UpdateOrderTags(context.Background(), tc.id, []string{"tag1"})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestUpdateOrderTagsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.UpdateOrderTags(context.Background(), "ord_123", []string{"tag1"})
	if err == nil {
		t.Error("Expected error from UpdateOrderTags")
	}
}

// Tests for SplitOrder

func TestSplitOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/orders/ord_123/split" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req OrderSplitRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if len(req.LineItemIDs) != 2 {
			t.Errorf("Expected 2 line item IDs, got %d", len(req.LineItemIDs))
		}
		if req.LineItemIDs[0] != "item_1" {
			t.Errorf("Expected first line item ID='item_1', got %s", req.LineItemIDs[0])
		}

		resp := OrderSplitResponse{
			OriginalOrder: &Order{
				ID:         "ord_123",
				Status:     "pending",
				TotalPrice: "50.00",
			},
			NewOrder: &Order{
				ID:         "ord_456",
				Status:     "pending",
				TotalPrice: "49.99",
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	result, err := client.SplitOrder(context.Background(), "ord_123", []string{"item_1", "item_2"})
	if err != nil {
		t.Fatalf("SplitOrder failed: %v", err)
	}

	if result.OriginalOrder == nil {
		t.Fatal("Expected OriginalOrder, got nil")
	}
	if result.OriginalOrder.ID != "ord_123" {
		t.Errorf("Expected original order ID='ord_123', got %s", result.OriginalOrder.ID)
	}
	if result.NewOrder == nil {
		t.Fatal("Expected NewOrder, got nil")
	}
	if result.NewOrder.ID != "ord_456" {
		t.Errorf("Expected new order ID='ord_456', got %s", result.NewOrder.ID)
	}
}

func TestSplitOrderEmptyID(t *testing.T) {
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
			_, err := client.SplitOrder(context.Background(), tc.id, []string{"item_1"})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestSplitOrderEmptyLineItems(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name        string
		lineItemIDs []string
	}{
		{"nil slice", nil},
		{"empty slice", []string{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.SplitOrder(context.Background(), "ord_123", tc.lineItemIDs)
			if err == nil {
				t.Error("Expected error for empty line items, got nil")
			}
			if err != nil && err.Error() != "at least one line item id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestSplitOrderAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.SplitOrder(context.Background(), "ord_123", []string{"item_1"})
	if err == nil {
		t.Error("Expected error from SplitOrder")
	}
}

// Tests for BulkExecuteShipment

func TestBulkExecuteShipment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("Expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/orders/execute_shipment" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req BulkShipmentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if len(req.OrderIDs) != 3 {
			t.Errorf("Expected 3 order IDs, got %d", len(req.OrderIDs))
		}

		resp := BulkShipmentResponse{
			Successful: []string{"ord_1", "ord_2"},
			Failed: []BulkShipmentFailure{
				{OrderID: "ord_3", Error: "insufficient inventory"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	result, err := client.BulkExecuteShipment(context.Background(), []string{"ord_1", "ord_2", "ord_3"})
	if err != nil {
		t.Fatalf("BulkExecuteShipment failed: %v", err)
	}

	if len(result.Successful) != 2 {
		t.Errorf("Expected 2 successful, got %d", len(result.Successful))
	}
	if result.Successful[0] != "ord_1" {
		t.Errorf("Expected first successful='ord_1', got %s", result.Successful[0])
	}
	if len(result.Failed) != 1 {
		t.Errorf("Expected 1 failed, got %d", len(result.Failed))
	}
	if result.Failed[0].OrderID != "ord_3" {
		t.Errorf("Expected failed order ID='ord_3', got %s", result.Failed[0].OrderID)
	}
	if result.Failed[0].Error != "insufficient inventory" {
		t.Errorf("Expected error='insufficient inventory', got %s", result.Failed[0].Error)
	}
}

func TestBulkExecuteShipmentEmptyOrderIDs(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name     string
		orderIDs []string
	}{
		{"nil slice", nil},
		{"empty slice", []string{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.BulkExecuteShipment(context.Background(), tc.orderIDs)
			if err == nil {
				t.Error("Expected error for empty order IDs, got nil")
			}
			if err != nil && err.Error() != "at least one order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestBulkExecuteShipmentAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.BulkExecuteShipment(context.Background(), []string{"ord_1"})
	if err == nil {
		t.Error("Expected error from BulkExecuteShipment")
	}
}

func TestBulkExecuteShipmentAllSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := BulkShipmentResponse{
			Successful: []string{"ord_1", "ord_2", "ord_3"},
			Failed:     []BulkShipmentFailure{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	result, err := client.BulkExecuteShipment(context.Background(), []string{"ord_1", "ord_2", "ord_3"})
	if err != nil {
		t.Fatalf("BulkExecuteShipment failed: %v", err)
	}

	if len(result.Successful) != 3 {
		t.Errorf("Expected 3 successful, got %d", len(result.Successful))
	}
	if len(result.Failed) != 0 {
		t.Errorf("Expected 0 failed, got %d", len(result.Failed))
	}
}

func TestBulkExecuteShipmentAllFailed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := BulkShipmentResponse{
			Successful: []string{},
			Failed: []BulkShipmentFailure{
				{OrderID: "ord_1", Error: "order not found"},
				{OrderID: "ord_2", Error: "already shipped"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	result, err := client.BulkExecuteShipment(context.Background(), []string{"ord_1", "ord_2"})
	if err != nil {
		t.Fatalf("BulkExecuteShipment failed: %v", err)
	}

	if len(result.Successful) != 0 {
		t.Errorf("Expected 0 successful, got %d", len(result.Successful))
	}
	if len(result.Failed) != 2 {
		t.Errorf("Expected 2 failed, got %d", len(result.Failed))
	}
}
