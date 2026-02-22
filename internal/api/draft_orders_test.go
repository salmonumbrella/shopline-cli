package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDraftOrdersList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/draft_orders" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := DraftOrdersListResponse{
			Items: []DraftOrder{
				{ID: "do_123", Status: "open", TotalPrice: "99.99"},
				{ID: "do_456", Status: "invoice_sent", TotalPrice: "149.99"},
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

	orders, err := client.ListDraftOrders(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListDraftOrders failed: %v", err)
	}

	if len(orders.Items) != 2 {
		t.Errorf("Expected 2 draft orders, got %d", len(orders.Items))
	}
	if orders.Items[0].ID != "do_123" {
		t.Errorf("Unexpected draft order ID: %s", orders.Items[0].ID)
	}
}

func TestDraftOrdersGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/draft_orders/do_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		order := DraftOrder{ID: "do_123", Status: "open", TotalPrice: "99.99"}
		_ = json.NewEncoder(w).Encode(order)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	order, err := client.GetDraftOrder(context.Background(), "do_123")
	if err != nil {
		t.Fatalf("GetDraftOrder failed: %v", err)
	}

	if order.ID != "do_123" {
		t.Errorf("Unexpected draft order ID: %s", order.ID)
	}
}

func TestGetDraftOrderEmptyID(t *testing.T) {
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
			_, err := client.GetDraftOrder(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "draft order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteDraftOrderEmptyID(t *testing.T) {
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
			err := client.DeleteDraftOrder(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "draft order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCompleteDraftOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/draft_orders/do_123/complete" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		order := DraftOrder{ID: "do_123", Status: "completed", TotalPrice: "99.99"}
		_ = json.NewEncoder(w).Encode(order)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	order, err := client.CompleteDraftOrder(context.Background(), "do_123")
	if err != nil {
		t.Fatalf("CompleteDraftOrder failed: %v", err)
	}

	if order.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", order.Status)
	}
}

func TestSendDraftOrderInvoice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/draft_orders/do_123/send_invoice" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.SendDraftOrderInvoice(context.Background(), "do_123")
	if err != nil {
		t.Fatalf("SendDraftOrderInvoice failed: %v", err)
	}
}

func TestCreateDraftOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/draft_orders" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		order := DraftOrder{
			ID:         "do_new",
			Status:     "open",
			TotalPrice: "199.99",
			LineItems: []DraftOrderLineItem{
				{VariantID: "var_1", Quantity: 2, Price: 99.99, Title: "Test Product"},
			},
		}
		_ = json.NewEncoder(w).Encode(order)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &DraftOrderCreateRequest{
		CustomerID: "cust_123",
		Note:       "Test draft order",
		LineItems: []DraftOrderLineItem{
			{VariantID: "var_1", Quantity: 2, Price: 99.99, Title: "Test Product"},
		},
	}

	order, err := client.CreateDraftOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateDraftOrder failed: %v", err)
	}

	if order.ID != "do_new" {
		t.Errorf("Unexpected draft order ID: %s", order.ID)
	}
	if len(order.LineItems) != 1 {
		t.Errorf("Expected 1 line item, got %d", len(order.LineItems))
	}
}

func TestDeleteDraftOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/draft_orders/do_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteDraftOrder(context.Background(), "do_123")
	if err != nil {
		t.Fatalf("DeleteDraftOrder failed: %v", err)
	}
}

func TestCompleteDraftOrderEmptyID(t *testing.T) {
	client := NewClient("token")

	_, err := client.CompleteDraftOrder(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
	if err != nil && err.Error() != "draft order id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestSendDraftOrderInvoiceEmptyID(t *testing.T) {
	client := NewClient("token")

	err := client.SendDraftOrderInvoice(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
	if err != nil && err.Error() != "draft order id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestDraftOrdersListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", query.Get("page"))
		}
		if query.Get("page_size") != "50" {
			t.Errorf("Expected page_size=50, got %s", query.Get("page_size"))
		}
		if query.Get("status") != "open" {
			t.Errorf("Expected status=open, got %s", query.Get("status"))
		}
		if query.Get("customer_id") != "cust_123" {
			t.Errorf("Expected customer_id=cust_123, got %s", query.Get("customer_id"))
		}

		resp := DraftOrdersListResponse{
			Items:      []DraftOrder{{ID: "do_123", Status: "open"}},
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

	opts := &DraftOrdersListOptions{
		Page:       2,
		PageSize:   50,
		Status:     "open",
		CustomerID: "cust_123",
	}
	orders, err := client.ListDraftOrders(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListDraftOrders failed: %v", err)
	}

	if len(orders.Items) != 1 {
		t.Errorf("Expected 1 draft order, got %d", len(orders.Items))
	}
}
