package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestReturnOrdersList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/return_orders" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := ReturnOrdersListResponse{
			Items: []ReturnOrder{
				{ID: "ret_123", OrderID: "ord_123", Status: "pending", TotalAmount: "99.99"},
				{ID: "ret_456", OrderID: "ord_456", Status: "received", TotalAmount: "149.99"},
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

	returns, err := client.ListReturnOrders(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListReturnOrders failed: %v", err)
	}

	if len(returns.Items) != 2 {
		t.Errorf("Expected 2 returns, got %d", len(returns.Items))
	}
	if returns.Items[0].ID != "ret_123" {
		t.Errorf("Unexpected return order ID: %s", returns.Items[0].ID)
	}
}

func TestReturnOrdersGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/return_orders/ret_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		returnOrder := ReturnOrder{ID: "ret_123", OrderID: "ord_123", Status: "pending", TotalAmount: "99.99"}
		_ = json.NewEncoder(w).Encode(returnOrder)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	returnOrder, err := client.GetReturnOrder(context.Background(), "ret_123")
	if err != nil {
		t.Fatalf("GetReturnOrder failed: %v", err)
	}

	if returnOrder.ID != "ret_123" {
		t.Errorf("Unexpected return order ID: %s", returnOrder.ID)
	}
}

func TestGetReturnOrderEmptyID(t *testing.T) {
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
			_, err := client.GetReturnOrder(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "return order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCreateReturnOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/return_orders" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		returnOrder := ReturnOrder{ID: "ret_new", OrderID: "ord_123", Status: "pending"}
		_ = json.NewEncoder(w).Encode(returnOrder)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ReturnOrderCreateRequest{
		OrderID: "ord_123",
		Reason:  "Defective product",
		LineItems: []ReturnOrderLineItem{
			{LineItemID: "li_123", Quantity: 1, ReturnReason: "defective"},
		},
	}

	returnOrder, err := client.CreateReturnOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateReturnOrder failed: %v", err)
	}

	if returnOrder.ID != "ret_new" {
		t.Errorf("Unexpected return order ID: %s", returnOrder.ID)
	}
}

func TestCancelReturnOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/return_orders/ret_123/cancel" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.CancelReturnOrder(context.Background(), "ret_123")
	if err != nil {
		t.Fatalf("CancelReturnOrder failed: %v", err)
	}
}

func TestCancelReturnOrderEmptyID(t *testing.T) {
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
			err := client.CancelReturnOrder(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "return order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCompleteReturnOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/return_orders/ret_123/complete" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		returnOrder := ReturnOrder{ID: "ret_123", Status: "completed"}
		_ = json.NewEncoder(w).Encode(returnOrder)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	returnOrder, err := client.CompleteReturnOrder(context.Background(), "ret_123")
	if err != nil {
		t.Fatalf("CompleteReturnOrder failed: %v", err)
	}

	if returnOrder.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", returnOrder.Status)
	}
}

func TestReceiveReturnOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/return_orders/ret_123/receive" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		returnOrder := ReturnOrder{ID: "ret_123", Status: "received"}
		_ = json.NewEncoder(w).Encode(returnOrder)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	returnOrder, err := client.ReceiveReturnOrder(context.Background(), "ret_123")
	if err != nil {
		t.Fatalf("ReceiveReturnOrder failed: %v", err)
	}

	if returnOrder.Status != "received" {
		t.Errorf("Expected status 'received', got '%s'", returnOrder.Status)
	}
}

func TestUpdateReturnOrderEmptyID(t *testing.T) {
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
			_, err := client.UpdateReturnOrder(context.Background(), tc.id, nil)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "return order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestListReturnOrdersWithOptions(t *testing.T) {
	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	testCases := []struct {
		name           string
		opts           *ReturnOrdersListOptions
		expectedParams map[string]string
	}{
		{
			name: "page option",
			opts: &ReturnOrdersListOptions{Page: 2},
			expectedParams: map[string]string{
				"page": "2",
			},
		},
		{
			name: "page_size option",
			opts: &ReturnOrdersListOptions{PageSize: 50},
			expectedParams: map[string]string{
				"page_size": "50",
			},
		},
		{
			name: "status option",
			opts: &ReturnOrdersListOptions{Status: "pending"},
			expectedParams: map[string]string{
				"status": "pending",
			},
		},
		{
			name: "order_id option",
			opts: &ReturnOrdersListOptions{OrderID: "ord_123"},
			expectedParams: map[string]string{
				"order_id": "ord_123",
			},
		},
		{
			name: "customer_id option",
			opts: &ReturnOrdersListOptions{CustomerID: "cust_456"},
			expectedParams: map[string]string{
				"customer_id": "cust_456",
			},
		},
		{
			name: "return_type option",
			opts: &ReturnOrdersListOptions{ReturnType: "refund"},
			expectedParams: map[string]string{
				"return_type": "refund",
			},
		},
		{
			name: "since option (created_at_min)",
			opts: &ReturnOrdersListOptions{Since: &since},
			expectedParams: map[string]string{
				"created_at_min": "2024-01-01T00:00:00Z",
			},
		},
		{
			name: "until option (created_at_max)",
			opts: &ReturnOrdersListOptions{Until: &until},
			expectedParams: map[string]string{
				"created_at_max": "2024-12-31T23:59:59Z",
			},
		},
		{
			name: "all options combined",
			opts: &ReturnOrdersListOptions{
				Page:       3,
				PageSize:   25,
				Status:     "received",
				OrderID:    "ord_789",
				CustomerID: "cust_012",
				ReturnType: "exchange",
				Since:      &since,
				Until:      &until,
			},
			expectedParams: map[string]string{
				"page":           "3",
				"page_size":      "25",
				"status":         "received",
				"order_id":       "ord_789",
				"customer_id":    "cust_012",
				"return_type":    "exchange",
				"created_at_min": "2024-01-01T00:00:00Z",
				"created_at_max": "2024-12-31T23:59:59Z",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET, got %s", r.Method)
				}
				if r.URL.Path != "/return_orders" {
					t.Errorf("Unexpected path: %s", r.URL.Path)
				}

				// Verify query parameters
				query := r.URL.Query()
				for key, expectedValue := range tc.expectedParams {
					if got := query.Get(key); got != expectedValue {
						t.Errorf("Expected query param %s=%s, got %s", key, expectedValue, got)
					}
				}

				resp := ReturnOrdersListResponse{
					Items:      []ReturnOrder{{ID: "ret_123", OrderID: "ord_123", Status: "pending"}},
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

			returns, err := client.ListReturnOrders(context.Background(), tc.opts)
			if err != nil {
				t.Fatalf("ListReturnOrders failed: %v", err)
			}

			if len(returns.Items) != 1 {
				t.Errorf("Expected 1 return order, got %d", len(returns.Items))
			}
		})
	}
}

func TestUpdateReturnOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/return_orders/ret_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ReturnOrderUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		returnOrder := ReturnOrder{
			ID:              "ret_123",
			OrderID:         "ord_123",
			Status:          "received",
			TrackingNumber:  "TRACK123",
			TrackingCompany: "UPS",
			Note:            "Customer returned item",
		}
		_ = json.NewEncoder(w).Encode(returnOrder)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	status := "received"
	trackingNum := "TRACK123"
	trackingCo := "UPS"
	note := "Customer returned item"
	req := &ReturnOrderUpdateRequest{
		Status:          &status,
		TrackingNumber:  &trackingNum,
		TrackingCompany: &trackingCo,
		Note:            &note,
	}

	returnOrder, err := client.UpdateReturnOrder(context.Background(), "ret_123", req)
	if err != nil {
		t.Fatalf("UpdateReturnOrder failed: %v", err)
	}

	if returnOrder.ID != "ret_123" {
		t.Errorf("Unexpected return order ID: %s", returnOrder.ID)
	}
	if returnOrder.Status != "received" {
		t.Errorf("Expected status 'received', got '%s'", returnOrder.Status)
	}
	if returnOrder.TrackingNumber != "TRACK123" {
		t.Errorf("Expected tracking number 'TRACK123', got '%s'", returnOrder.TrackingNumber)
	}
}

func TestCompleteReturnOrderEmptyID(t *testing.T) {
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
			_, err := client.CompleteReturnOrder(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "return order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestReceiveReturnOrderEmptyID(t *testing.T) {
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
			_, err := client.ReceiveReturnOrder(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "return order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestListReturnOrdersError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.ListReturnOrders(context.Background(), nil)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestGetReturnOrderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Return order not found"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.GetReturnOrder(context.Background(), "ret_nonexistent")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCreateReturnOrderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ReturnOrderCreateRequest{
		OrderID: "ord_123",
		LineItems: []ReturnOrderLineItem{
			{LineItemID: "li_123", Quantity: 1},
		},
	}

	_, err := client.CreateReturnOrder(context.Background(), req)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestUpdateReturnOrderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Return order not found"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	note := "Updated note"
	req := &ReturnOrderUpdateRequest{Note: &note}

	_, err := client.UpdateReturnOrder(context.Background(), "ret_nonexistent", req)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCancelReturnOrderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Cannot cancel return order"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.CancelReturnOrder(context.Background(), "ret_123")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCompleteReturnOrderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Cannot complete return order"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.CompleteReturnOrder(context.Background(), "ret_123")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestReceiveReturnOrderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Cannot receive return order"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.ReceiveReturnOrder(context.Background(), "ret_123")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCreateReturnOrderWithLineItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		var req ReturnOrderCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Verify line items
		if len(req.LineItems) != 2 {
			t.Errorf("Expected 2 line items, got %d", len(req.LineItems))
		}
		if req.LineItems[0].LineItemID != "li_001" {
			t.Errorf("Unexpected line item ID: %s", req.LineItems[0].LineItemID)
		}
		if req.LineItems[0].Quantity != 2 {
			t.Errorf("Expected quantity 2, got %d", req.LineItems[0].Quantity)
		}
		if req.LineItems[1].ReturnReason != "wrong_size" {
			t.Errorf("Unexpected return reason: %s", req.LineItems[1].ReturnReason)
		}

		now := time.Now()
		returnOrder := ReturnOrder{
			ID:        "ret_full",
			OrderID:   req.OrderID,
			Status:    "pending",
			Reason:    req.Reason,
			Note:      req.Note,
			LineItems: req.LineItems,
			CreatedAt: now,
			UpdatedAt: now,
		}
		_ = json.NewEncoder(w).Encode(returnOrder)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ReturnOrderCreateRequest{
		OrderID: "ord_456",
		Reason:  "Multiple items to return",
		Note:    "Customer wants refund",
		LineItems: []ReturnOrderLineItem{
			{
				LineItemID:   "li_001",
				VariantID:    "var_001",
				ProductID:    "prod_001",
				Title:        "Blue T-Shirt",
				Quantity:     2,
				ReturnReason: "defective",
			},
			{
				LineItemID:   "li_002",
				VariantID:    "var_002",
				ProductID:    "prod_002",
				Title:        "Red Pants",
				Quantity:     1,
				ReturnReason: "wrong_size",
			},
		},
	}

	returnOrder, err := client.CreateReturnOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateReturnOrder failed: %v", err)
	}

	if returnOrder.ID != "ret_full" {
		t.Errorf("Unexpected return order ID: %s", returnOrder.ID)
	}
	if len(returnOrder.LineItems) != 2 {
		t.Errorf("Expected 2 line items, got %d", len(returnOrder.LineItems))
	}
}

func TestGetReturnOrderWithFullData(t *testing.T) {
	receivedAt := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	completedAt := time.Date(2024, 6, 16, 14, 0, 0, 0, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/return_orders/ret_full" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		returnOrder := ReturnOrder{
			ID:              "ret_full",
			OrderID:         "ord_999",
			OrderNumber:     "ORD-999",
			Status:          "completed",
			ReturnType:      "refund",
			CustomerID:      "cust_123",
			CustomerEmail:   "customer@example.com",
			TotalAmount:     "299.99",
			RefundAmount:    "250.00",
			Currency:        "USD",
			Reason:          "Product defective",
			Note:            "Full refund issued",
			TrackingNumber:  "TRACK789",
			TrackingCompany: "FedEx",
			ReceivedAt:      &receivedAt,
			CompletedAt:     &completedAt,
			CancelledAt:     nil,
			CreatedAt:       time.Date(2024, 6, 10, 9, 0, 0, 0, time.UTC),
			UpdatedAt:       time.Date(2024, 6, 16, 14, 0, 0, 0, time.UTC),
			LineItems: []ReturnOrderLineItem{
				{
					LineItemID:   "li_full",
					VariantID:    "var_full",
					ProductID:    "prod_full",
					Title:        "Premium Widget",
					Quantity:     1,
					ReturnReason: "defective",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(returnOrder)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	returnOrder, err := client.GetReturnOrder(context.Background(), "ret_full")
	if err != nil {
		t.Fatalf("GetReturnOrder failed: %v", err)
	}

	// Verify all fields
	if returnOrder.ID != "ret_full" {
		t.Errorf("Unexpected ID: %s", returnOrder.ID)
	}
	if returnOrder.OrderNumber != "ORD-999" {
		t.Errorf("Unexpected OrderNumber: %s", returnOrder.OrderNumber)
	}
	if returnOrder.Status != "completed" {
		t.Errorf("Unexpected Status: %s", returnOrder.Status)
	}
	if returnOrder.ReturnType != "refund" {
		t.Errorf("Unexpected ReturnType: %s", returnOrder.ReturnType)
	}
	if returnOrder.CustomerEmail != "customer@example.com" {
		t.Errorf("Unexpected CustomerEmail: %s", returnOrder.CustomerEmail)
	}
	if returnOrder.TotalAmount != "299.99" {
		t.Errorf("Unexpected TotalAmount: %s", returnOrder.TotalAmount)
	}
	if returnOrder.RefundAmount != "250.00" {
		t.Errorf("Unexpected RefundAmount: %s", returnOrder.RefundAmount)
	}
	if returnOrder.Currency != "USD" {
		t.Errorf("Unexpected Currency: %s", returnOrder.Currency)
	}
	if returnOrder.TrackingNumber != "TRACK789" {
		t.Errorf("Unexpected TrackingNumber: %s", returnOrder.TrackingNumber)
	}
	if returnOrder.TrackingCompany != "FedEx" {
		t.Errorf("Unexpected TrackingCompany: %s", returnOrder.TrackingCompany)
	}
	if returnOrder.ReceivedAt == nil {
		t.Error("Expected ReceivedAt to be set")
	}
	if returnOrder.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}
	if returnOrder.CancelledAt != nil {
		t.Error("Expected CancelledAt to be nil")
	}
	if len(returnOrder.LineItems) != 1 {
		t.Errorf("Expected 1 line item, got %d", len(returnOrder.LineItems))
	}
}
