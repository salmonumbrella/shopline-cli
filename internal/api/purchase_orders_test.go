package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPurchaseOrdersList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/purchase_orders" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := PurchaseOrdersListResponse{
			Items: []PurchaseOrder{
				{ID: "po_123", Number: "PO-001", Status: "pending", Total: "1000.00"},
				{ID: "po_456", Number: "PO-002", Status: "received", Total: "500.00"},
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

	orders, err := client.ListPurchaseOrders(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListPurchaseOrders failed: %v", err)
	}

	if len(orders.Items) != 2 {
		t.Errorf("Expected 2 purchase orders, got %d", len(orders.Items))
	}
	if orders.Items[0].ID != "po_123" {
		t.Errorf("Unexpected purchase order ID: %s", orders.Items[0].ID)
	}
}

func TestPurchaseOrdersGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/purchase_orders/po_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		po := PurchaseOrder{ID: "po_123", Number: "PO-001", Status: "pending"}
		_ = json.NewEncoder(w).Encode(po)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	po, err := client.GetPurchaseOrder(context.Background(), "po_123")
	if err != nil {
		t.Fatalf("GetPurchaseOrder failed: %v", err)
	}

	if po.ID != "po_123" {
		t.Errorf("Unexpected purchase order ID: %s", po.ID)
	}
}

func TestGetPurchaseOrderEmptyID(t *testing.T) {
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
			_, err := client.GetPurchaseOrder(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "purchase order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestPurchaseOrdersCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		po := PurchaseOrder{ID: "po_new", Number: "PO-003", Status: "draft"}
		_ = json.NewEncoder(w).Encode(po)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &PurchaseOrderCreateRequest{
		SupplierID:  "sup_123",
		WarehouseID: "wh_123",
		LineItems: []PurchaseOrderItemRequest{
			{VariantID: "var_123", Quantity: 10, UnitCost: "50.00"},
		},
	}

	po, err := client.CreatePurchaseOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("CreatePurchaseOrder failed: %v", err)
	}

	if po.ID != "po_new" {
		t.Errorf("Unexpected purchase order ID: %s", po.ID)
	}
}

func TestPurchaseOrdersReceive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/purchase_orders/po_123/receive" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		po := PurchaseOrder{ID: "po_123", Status: "received"}
		_ = json.NewEncoder(w).Encode(po)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	po, err := client.ReceivePurchaseOrder(context.Background(), "po_123")
	if err != nil {
		t.Fatalf("ReceivePurchaseOrder failed: %v", err)
	}

	if po.Status != "received" {
		t.Errorf("Unexpected status: %s", po.Status)
	}
}

func TestPurchaseOrdersCancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/purchase_orders/po_123/cancel" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		po := PurchaseOrder{ID: "po_123", Status: "cancelled"}
		_ = json.NewEncoder(w).Encode(po)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	po, err := client.CancelPurchaseOrder(context.Background(), "po_123")
	if err != nil {
		t.Fatalf("CancelPurchaseOrder failed: %v", err)
	}

	if po.Status != "cancelled" {
		t.Errorf("Unexpected status: %s", po.Status)
	}
}

func TestPurchaseOrdersDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/purchase_orders/po_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeletePurchaseOrder(context.Background(), "po_123")
	if err != nil {
		t.Fatalf("DeletePurchaseOrder failed: %v", err)
	}
}

func TestReceivePurchaseOrderEmptyID(t *testing.T) {
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
			_, err := client.ReceivePurchaseOrder(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "purchase order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCancelPurchaseOrderEmptyID(t *testing.T) {
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
			_, err := client.CancelPurchaseOrder(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "purchase order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeletePurchaseOrderEmptyID(t *testing.T) {
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
			err := client.DeletePurchaseOrder(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "purchase order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestPurchaseOrdersListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		query := r.URL.Query()
		if query.Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", query.Get("page"))
		}
		if query.Get("page_size") != "10" {
			t.Errorf("Expected page_size=10, got %s", query.Get("page_size"))
		}
		if query.Get("status") != "pending" {
			t.Errorf("Expected status=pending, got %s", query.Get("status"))
		}
		if query.Get("supplier_id") != "sup_123" {
			t.Errorf("Expected supplier_id=sup_123, got %s", query.Get("supplier_id"))
		}
		if query.Get("warehouse_id") != "wh_456" {
			t.Errorf("Expected warehouse_id=wh_456, got %s", query.Get("warehouse_id"))
		}

		resp := PurchaseOrdersListResponse{
			Items:      []PurchaseOrder{{ID: "po_123", Number: "PO-001"}},
			Page:       2,
			PageSize:   10,
			TotalCount: 1,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &PurchaseOrdersListOptions{
		Page:        2,
		PageSize:    10,
		Status:      "pending",
		SupplierID:  "sup_123",
		WarehouseID: "wh_456",
	}
	orders, err := client.ListPurchaseOrders(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListPurchaseOrders failed: %v", err)
	}

	if len(orders.Items) != 1 {
		t.Errorf("Expected 1 purchase order, got %d", len(orders.Items))
	}
}
