package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRefundsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/refunds" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := RefundsListResponse{
			Items: []Refund{
				{ID: "ref_123", OrderID: "ord_123", Amount: "50.00", Status: "success"},
				{ID: "ref_456", OrderID: "ord_456", Amount: "25.00", Status: "success"},
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

	refunds, err := client.ListRefunds(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListRefunds failed: %v", err)
	}

	if len(refunds.Items) != 2 {
		t.Errorf("Expected 2 refunds, got %d", len(refunds.Items))
	}
	if refunds.Items[0].ID != "ref_123" {
		t.Errorf("Unexpected refund ID: %s", refunds.Items[0].ID)
	}
}

func TestRefundsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/refunds/ref_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		refund := Refund{ID: "ref_123", OrderID: "ord_123", Amount: "50.00"}
		_ = json.NewEncoder(w).Encode(refund)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	refund, err := client.GetRefund(context.Background(), "ref_123")
	if err != nil {
		t.Fatalf("GetRefund failed: %v", err)
	}

	if refund.ID != "ref_123" {
		t.Errorf("Unexpected refund ID: %s", refund.ID)
	}
}

func TestGetRefundEmptyID(t *testing.T) {
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
			_, err := client.GetRefund(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "refund id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestRefundsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/refunds" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		refund := Refund{ID: "ref_new", OrderID: "ord_123", Amount: "30.00", Status: "pending"}
		_ = json.NewEncoder(w).Encode(refund)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &RefundCreateRequest{OrderID: "ord_123", Amount: 30.00, Note: "Customer request"}
	refund, err := client.CreateRefund(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateRefund failed: %v", err)
	}

	if refund.ID != "ref_new" {
		t.Errorf("Unexpected refund ID: %s", refund.ID)
	}
}

func TestListOrderRefunds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/orders/ord_123/refunds" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := RefundsListResponse{
			Items: []Refund{
				{ID: "ref_123", OrderID: "ord_123", Amount: "50.00", Status: "success"},
			},
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

	refunds, err := client.ListOrderRefunds(context.Background(), "ord_123")
	if err != nil {
		t.Fatalf("ListOrderRefunds failed: %v", err)
	}

	if len(refunds.Items) != 1 {
		t.Errorf("Expected 1 refund, got %d", len(refunds.Items))
	}
}

func TestListOrderRefundsEmptyID(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name    string
		orderID string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab only", "\t"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.ListOrderRefunds(context.Background(), tc.orderID)
			if err == nil {
				t.Error("Expected error for empty order ID, got nil")
			}
			if err != nil && err.Error() != "order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestListRefundsWithOptions(t *testing.T) {
	testCases := []struct {
		name           string
		opts           *RefundsListOptions
		expectedParams map[string]string
	}{
		{
			name: "page option",
			opts: &RefundsListOptions{Page: 2},
			expectedParams: map[string]string{
				"page": "2",
			},
		},
		{
			name: "page_size option",
			opts: &RefundsListOptions{PageSize: 50},
			expectedParams: map[string]string{
				"page_size": "50",
			},
		},
		{
			name: "order_id option",
			opts: &RefundsListOptions{OrderID: "ord_123"},
			expectedParams: map[string]string{
				"order_id": "ord_123",
			},
		},
		{
			name: "status option",
			opts: &RefundsListOptions{Status: "success"},
			expectedParams: map[string]string{
				"status": "success",
			},
		},
		{
			name: "all options combined",
			opts: &RefundsListOptions{
				Page:     3,
				PageSize: 25,
				OrderID:  "ord_456",
				Status:   "pending",
			},
			expectedParams: map[string]string{
				"page":      "3",
				"page_size": "25",
				"order_id":  "ord_456",
				"status":    "pending",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET, got %s", r.Method)
				}
				if r.URL.Path != "/refunds" {
					t.Errorf("Unexpected path: %s", r.URL.Path)
				}

				// Verify query parameters
				query := r.URL.Query()
				for key, expectedValue := range tc.expectedParams {
					if got := query.Get(key); got != expectedValue {
						t.Errorf("Expected query param %s=%s, got %s", key, expectedValue, got)
					}
				}

				resp := RefundsListResponse{
					Items:      []Refund{{ID: "ref_123", OrderID: "ord_123", Amount: "50.00", Status: "success"}},
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

			refunds, err := client.ListRefunds(context.Background(), tc.opts)
			if err != nil {
				t.Fatalf("ListRefunds failed: %v", err)
			}

			if len(refunds.Items) != 1 {
				t.Errorf("Expected 1 refund, got %d", len(refunds.Items))
			}
		})
	}
}

func TestCreateRefundWithLineItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/refunds" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		// Verify the request body contains line items
		var req RefundCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.OrderID != "ord_123" {
			t.Errorf("Expected order_id ord_123, got %s", req.OrderID)
		}
		if len(req.LineItems) != 2 {
			t.Errorf("Expected 2 line items, got %d", len(req.LineItems))
		}
		if req.LineItems[0].LineItemID != "li_001" {
			t.Errorf("Expected line item ID li_001, got %s", req.LineItems[0].LineItemID)
		}
		if req.LineItems[0].Quantity != 1 {
			t.Errorf("Expected quantity 1, got %d", req.LineItems[0].Quantity)
		}
		if req.LineItems[0].RestockType != "return" {
			t.Errorf("Expected restock_type return, got %s", req.LineItems[0].RestockType)
		}
		if req.LineItems[1].LineItemID != "li_002" {
			t.Errorf("Expected line item ID li_002, got %s", req.LineItems[1].LineItemID)
		}

		refund := Refund{
			ID:      "ref_with_items",
			OrderID: "ord_123",
			Amount:  "75.00",
			Status:  "pending",
			LineItems: []RefundLineItem{
				{LineItemID: "li_001", Quantity: 1, RestockType: "return", Subtotal: 50.00},
				{LineItemID: "li_002", Quantity: 2, RestockType: "cancel", Subtotal: 25.00},
			},
		}
		_ = json.NewEncoder(w).Encode(refund)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &RefundCreateRequest{
		OrderID: "ord_123",
		Note:    "Partial refund with line items",
		Restock: true,
		LineItems: []RefundLineItem{
			{LineItemID: "li_001", Quantity: 1, RestockType: "return", Subtotal: 50.00},
			{LineItemID: "li_002", Quantity: 2, RestockType: "cancel", Subtotal: 25.00},
		},
	}
	refund, err := client.CreateRefund(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateRefund failed: %v", err)
	}

	if refund.ID != "ref_with_items" {
		t.Errorf("Unexpected refund ID: %s", refund.ID)
	}
	if len(refund.LineItems) != 2 {
		t.Errorf("Expected 2 line items in response, got %d", len(refund.LineItems))
	}
}

func TestCreateRefundWithRestock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		var req RefundCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if !req.Restock {
			t.Error("Expected restock to be true")
		}
		if req.Note != "Restock items" {
			t.Errorf("Expected note 'Restock items', got %s", req.Note)
		}

		refund := Refund{
			ID:      "ref_restock",
			OrderID: req.OrderID,
			Amount:  "100.00",
			Status:  "success",
			Restock: true,
			Note:    req.Note,
		}
		_ = json.NewEncoder(w).Encode(refund)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &RefundCreateRequest{
		OrderID: "ord_789",
		Restock: true,
		Note:    "Restock items",
		Amount:  100.00,
	}
	refund, err := client.CreateRefund(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateRefund failed: %v", err)
	}

	if refund.ID != "ref_restock" {
		t.Errorf("Unexpected refund ID: %s", refund.ID)
	}
	if !refund.Restock {
		t.Error("Expected restock to be true in response")
	}
	if refund.Note != "Restock items" {
		t.Errorf("Unexpected note: %s", refund.Note)
	}
}

func TestRefundFullResponseFields(t *testing.T) {
	processedAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	createdAt := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		refund := Refund{
			ID:          "ref_full",
			OrderID:     "ord_full",
			Note:        "Full refund test",
			Restock:     true,
			Amount:      "250.50",
			Currency:    "USD",
			Status:      "success",
			ProcessedAt: processedAt,
			CreatedAt:   createdAt,
			LineItems: []RefundLineItem{
				{
					LineItemID:  "li_full_001",
					Quantity:    3,
					RestockType: "return",
					Subtotal:    150.25,
				},
				{
					LineItemID:  "li_full_002",
					Quantity:    1,
					RestockType: "cancel",
					Subtotal:    100.25,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(refund)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	refund, err := client.GetRefund(context.Background(), "ref_full")
	if err != nil {
		t.Fatalf("GetRefund failed: %v", err)
	}

	// Verify all fields
	if refund.ID != "ref_full" {
		t.Errorf("Unexpected ID: %s", refund.ID)
	}
	if refund.OrderID != "ord_full" {
		t.Errorf("Unexpected OrderID: %s", refund.OrderID)
	}
	if refund.Note != "Full refund test" {
		t.Errorf("Unexpected Note: %s", refund.Note)
	}
	if !refund.Restock {
		t.Error("Expected Restock to be true")
	}
	if refund.Amount != "250.50" {
		t.Errorf("Unexpected Amount: %s", refund.Amount)
	}
	if refund.Currency != "USD" {
		t.Errorf("Unexpected Currency: %s", refund.Currency)
	}
	if refund.Status != "success" {
		t.Errorf("Unexpected Status: %s", refund.Status)
	}
	if !refund.ProcessedAt.Equal(processedAt) {
		t.Errorf("Unexpected ProcessedAt: %v", refund.ProcessedAt)
	}
	if !refund.CreatedAt.Equal(createdAt) {
		t.Errorf("Unexpected CreatedAt: %v", refund.CreatedAt)
	}
	if len(refund.LineItems) != 2 {
		t.Errorf("Expected 2 line items, got %d", len(refund.LineItems))
	}
	if refund.LineItems[0].Subtotal != 150.25 {
		t.Errorf("Unexpected line item subtotal: %f", refund.LineItems[0].Subtotal)
	}
}

func TestRefundsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	// Test ListRefunds API error
	_, err := client.ListRefunds(context.Background(), nil)
	if err == nil {
		t.Error("Expected error from ListRefunds")
	}

	// Test GetRefund API error
	_, err = client.GetRefund(context.Background(), "ref_123")
	if err == nil {
		t.Error("Expected error from GetRefund")
	}

	// Test CreateRefund API error
	_, err = client.CreateRefund(context.Background(), &RefundCreateRequest{OrderID: "ord_123"})
	if err == nil {
		t.Error("Expected error from CreateRefund")
	}

	// Test ListOrderRefunds API error
	_, err = client.ListOrderRefunds(context.Background(), "ord_123")
	if err == nil {
		t.Error("Expected error from ListOrderRefunds")
	}
}

func TestRefundsNotFoundError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	// Test GetRefund not found
	_, err := client.GetRefund(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent refund")
	}

	// Test ListOrderRefunds not found
	_, err = client.ListOrderRefunds(context.Background(), "nonexistent_order")
	if err == nil {
		t.Error("Expected error for non-existent order")
	}
}

func TestListRefundsEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := RefundsListResponse{
			Items:      []Refund{},
			Page:       1,
			PageSize:   20,
			TotalCount: 0,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	refunds, err := client.ListRefunds(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListRefunds failed: %v", err)
	}

	if len(refunds.Items) != 0 {
		t.Errorf("Expected 0 refunds, got %d", len(refunds.Items))
	}
	if refunds.TotalCount != 0 {
		t.Errorf("Expected TotalCount 0, got %d", refunds.TotalCount)
	}
}

func TestListOrderRefundsMultipleRefunds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orders/ord_multi/refunds" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := RefundsListResponse{
			Items: []Refund{
				{ID: "ref_1", OrderID: "ord_multi", Amount: "10.00", Status: "success"},
				{ID: "ref_2", OrderID: "ord_multi", Amount: "20.00", Status: "success"},
				{ID: "ref_3", OrderID: "ord_multi", Amount: "30.00", Status: "pending"},
			},
			Page:       1,
			PageSize:   20,
			TotalCount: 3,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	refunds, err := client.ListOrderRefunds(context.Background(), "ord_multi")
	if err != nil {
		t.Fatalf("ListOrderRefunds failed: %v", err)
	}

	if len(refunds.Items) != 3 {
		t.Errorf("Expected 3 refunds, got %d", len(refunds.Items))
	}
	if refunds.TotalCount != 3 {
		t.Errorf("Expected TotalCount 3, got %d", refunds.TotalCount)
	}

	// Verify all refunds belong to the same order
	for _, r := range refunds.Items {
		if r.OrderID != "ord_multi" {
			t.Errorf("Expected OrderID ord_multi, got %s", r.OrderID)
		}
	}
}
