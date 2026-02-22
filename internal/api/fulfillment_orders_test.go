package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFulfillmentOrdersList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/fulfillment_orders" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := FulfillmentOrdersListResponse{
			Items: []FulfillmentOrder{
				{ID: "fo_123", OrderID: "ord_123", Status: "open"},
				{ID: "fo_456", OrderID: "ord_456", Status: "in_progress"},
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

	orders, err := client.ListFulfillmentOrders(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListFulfillmentOrders failed: %v", err)
	}

	if len(orders.Items) != 2 {
		t.Errorf("Expected 2 fulfillment orders, got %d", len(orders.Items))
	}
	if orders.Items[0].ID != "fo_123" {
		t.Errorf("Unexpected fulfillment order ID: %s", orders.Items[0].ID)
	}
}

func TestFulfillmentOrdersGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/fulfillment_orders/fo_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		fo := FulfillmentOrder{ID: "fo_123", OrderID: "ord_123", Status: "open"}
		_ = json.NewEncoder(w).Encode(fo)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	fo, err := client.GetFulfillmentOrder(context.Background(), "fo_123")
	if err != nil {
		t.Fatalf("GetFulfillmentOrder failed: %v", err)
	}

	if fo.ID != "fo_123" {
		t.Errorf("Unexpected fulfillment order ID: %s", fo.ID)
	}
}

func TestGetFulfillmentOrderEmptyID(t *testing.T) {
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
			_, err := client.GetFulfillmentOrder(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "fulfillment order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestListOrderFulfillmentOrders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orders/ord_123/fulfillment_orders" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := FulfillmentOrdersListResponse{
			Items: []FulfillmentOrder{
				{ID: "fo_123", OrderID: "ord_123", Status: "open"},
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

	orders, err := client.ListOrderFulfillmentOrders(context.Background(), "ord_123")
	if err != nil {
		t.Fatalf("ListOrderFulfillmentOrders failed: %v", err)
	}

	if len(orders.Items) != 1 {
		t.Errorf("Expected 1 fulfillment order, got %d", len(orders.Items))
	}
}

func TestMoveFulfillmentOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/fulfillment_orders/fo_123/move" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		fo := FulfillmentOrder{ID: "fo_123", AssignedLocationID: "loc_456"}
		_ = json.NewEncoder(w).Encode(fo)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	fo, err := client.MoveFulfillmentOrder(context.Background(), "fo_123", "loc_456")
	if err != nil {
		t.Fatalf("MoveFulfillmentOrder failed: %v", err)
	}

	if fo.AssignedLocationID != "loc_456" {
		t.Errorf("Unexpected location ID: %s", fo.AssignedLocationID)
	}
}

func TestCancelFulfillmentOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/fulfillment_orders/fo_123/cancel" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		fo := FulfillmentOrder{ID: "fo_123", Status: "cancelled"}
		_ = json.NewEncoder(w).Encode(fo)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	fo, err := client.CancelFulfillmentOrder(context.Background(), "fo_123")
	if err != nil {
		t.Fatalf("CancelFulfillmentOrder failed: %v", err)
	}

	if fo.Status != "cancelled" {
		t.Errorf("Unexpected status: %s", fo.Status)
	}
}

func TestCancelFulfillmentOrderEmptyID(t *testing.T) {
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
			_, err := client.CancelFulfillmentOrder(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "fulfillment order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCloseFulfillmentOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/fulfillment_orders/fo_123/close" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		fo := FulfillmentOrder{ID: "fo_123", Status: "closed"}
		_ = json.NewEncoder(w).Encode(fo)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	fo, err := client.CloseFulfillmentOrder(context.Background(), "fo_123")
	if err != nil {
		t.Fatalf("CloseFulfillmentOrder failed: %v", err)
	}

	if fo.Status != "closed" {
		t.Errorf("Unexpected status: %s", fo.Status)
	}
}

func TestCloseFulfillmentOrderEmptyID(t *testing.T) {
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
			_, err := client.CloseFulfillmentOrder(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "fulfillment order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestListFulfillmentOrdersWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/fulfillment_orders" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		// Verify query parameters
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
		if query.Get("order_id") != "ord_789" {
			t.Errorf("Expected order_id=ord_789, got %s", query.Get("order_id"))
		}

		resp := FulfillmentOrdersListResponse{
			Items: []FulfillmentOrder{
				{ID: "fo_123", OrderID: "ord_789", Status: "open"},
			},
			Page:       2,
			PageSize:   50,
			TotalCount: 51,
			HasMore:    true,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &FulfillmentOrdersListOptions{
		Page:     2,
		PageSize: 50,
		Status:   "open",
		OrderID:  "ord_789",
	}
	orders, err := client.ListFulfillmentOrders(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListFulfillmentOrders with options failed: %v", err)
	}

	if len(orders.Items) != 1 {
		t.Errorf("Expected 1 fulfillment order, got %d", len(orders.Items))
	}
	if orders.Page != 2 {
		t.Errorf("Expected page 2, got %d", orders.Page)
	}
	if orders.PageSize != 50 {
		t.Errorf("Expected page_size 50, got %d", orders.PageSize)
	}
	if !orders.HasMore {
		t.Error("Expected HasMore to be true")
	}
}

func TestListOrderFulfillmentOrdersEmptyID(t *testing.T) {
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
			_, err := client.ListOrderFulfillmentOrders(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty order ID, got nil")
			}
			if err != nil && err.Error() != "order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestMoveFulfillmentOrderEmptyID(t *testing.T) {
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
			_, err := client.MoveFulfillmentOrder(context.Background(), tc.id, "loc_123")
			if err == nil {
				t.Error("Expected error for empty fulfillment order ID, got nil")
			}
			if err != nil && err.Error() != "fulfillment order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestMoveFulfillmentOrderEmptyLocationID(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name       string
		locationID string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab only", "\t"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.MoveFulfillmentOrder(context.Background(), "fo_123", tc.locationID)
			if err == nil {
				t.Error("Expected error for empty location ID, got nil")
			}
			if err != nil && err.Error() != "new location id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestMoveFulfillmentOrderRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		// Verify the request body contains the new location ID
		var reqBody FulfillmentOrderMoveRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		if reqBody.NewLocationID != "loc_new_789" {
			t.Errorf("Expected new_location_id=loc_new_789, got %s", reqBody.NewLocationID)
		}

		fo := FulfillmentOrder{ID: "fo_123", AssignedLocationID: "loc_new_789"}
		_ = json.NewEncoder(w).Encode(fo)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	fo, err := client.MoveFulfillmentOrder(context.Background(), "fo_123", "loc_new_789")
	if err != nil {
		t.Fatalf("MoveFulfillmentOrder failed: %v", err)
	}

	if fo.AssignedLocationID != "loc_new_789" {
		t.Errorf("Unexpected location ID: %s", fo.AssignedLocationID)
	}
}

func TestFulfillmentOrderWithLineItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fo := FulfillmentOrder{
			ID:      "fo_123",
			OrderID: "ord_123",
			Status:  "open",
			LineItems: []FulfillmentOrderItem{
				{
					ID:                  "foi_1",
					LineItemID:          "li_1",
					VariantID:           "var_1",
					Quantity:            3,
					FulfillableQuantity: 2,
					FulfilledQuantity:   1,
				},
				{
					ID:                  "foi_2",
					LineItemID:          "li_2",
					VariantID:           "var_2",
					Quantity:            5,
					FulfillableQuantity: 5,
					FulfilledQuantity:   0,
				},
			},
			DeliveryMethod: FulfillmentDeliveryMethod{
				MethodType:  "shipping",
				ServiceCode: "standard",
			},
		}
		_ = json.NewEncoder(w).Encode(fo)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	fo, err := client.GetFulfillmentOrder(context.Background(), "fo_123")
	if err != nil {
		t.Fatalf("GetFulfillmentOrder failed: %v", err)
	}

	if len(fo.LineItems) != 2 {
		t.Errorf("Expected 2 line items, got %d", len(fo.LineItems))
	}

	// Verify first line item
	if fo.LineItems[0].ID != "foi_1" {
		t.Errorf("Unexpected line item ID: %s", fo.LineItems[0].ID)
	}
	if fo.LineItems[0].Quantity != 3 {
		t.Errorf("Expected quantity 3, got %d", fo.LineItems[0].Quantity)
	}
	if fo.LineItems[0].FulfillableQuantity != 2 {
		t.Errorf("Expected fulfillable quantity 2, got %d", fo.LineItems[0].FulfillableQuantity)
	}
	if fo.LineItems[0].FulfilledQuantity != 1 {
		t.Errorf("Expected fulfilled quantity 1, got %d", fo.LineItems[0].FulfilledQuantity)
	}

	// Verify delivery method
	if fo.DeliveryMethod.MethodType != "shipping" {
		t.Errorf("Expected method_type shipping, got %s", fo.DeliveryMethod.MethodType)
	}
	if fo.DeliveryMethod.ServiceCode != "standard" {
		t.Errorf("Expected service_code standard, got %s", fo.DeliveryMethod.ServiceCode)
	}
}

func TestFulfillmentOrderServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.GetFulfillmentOrder(context.Background(), "fo_123")
	if err == nil {
		t.Error("Expected error for server error response, got nil")
	}
}

func TestListFulfillmentOrdersServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "bad request"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.ListFulfillmentOrders(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for server error response, got nil")
	}
}

func TestCancelFulfillmentOrderMethod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method for cancel, got %s", r.Method)
		}
		if r.URL.Path != "/fulfillment_orders/fo_cancel_test/cancel" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		fo := FulfillmentOrder{ID: "fo_cancel_test", Status: "cancelled"}
		_ = json.NewEncoder(w).Encode(fo)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	fo, err := client.CancelFulfillmentOrder(context.Background(), "fo_cancel_test")
	if err != nil {
		t.Fatalf("CancelFulfillmentOrder failed: %v", err)
	}

	if fo.ID != "fo_cancel_test" {
		t.Errorf("Unexpected ID: %s", fo.ID)
	}
}

func TestListOrderFulfillmentOrdersMethod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/orders/ord_method_test/fulfillment_orders" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := FulfillmentOrdersListResponse{
			Items:      []FulfillmentOrder{},
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

	orders, err := client.ListOrderFulfillmentOrders(context.Background(), "ord_method_test")
	if err != nil {
		t.Fatalf("ListOrderFulfillmentOrders failed: %v", err)
	}

	if len(orders.Items) != 0 {
		t.Errorf("Expected 0 fulfillment orders, got %d", len(orders.Items))
	}
}

func TestFulfillmentOrderNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "fulfillment order not found"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.GetFulfillmentOrder(context.Background(), "fo_nonexistent")
	if err == nil {
		t.Error("Expected error for not found response, got nil")
	}
}

func TestMoveFulfillmentOrderServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "cannot move fulfillment order"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.MoveFulfillmentOrder(context.Background(), "fo_123", "loc_456")
	if err == nil {
		t.Error("Expected error for unprocessable entity response, got nil")
	}
}

func TestCancelFulfillmentOrderServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "fulfillment order already cancelled"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.CancelFulfillmentOrder(context.Background(), "fo_123")
	if err == nil {
		t.Error("Expected error for conflict response, got nil")
	}
}

func TestCloseFulfillmentOrderServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "cannot close fulfillment order"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.CloseFulfillmentOrder(context.Background(), "fo_123")
	if err == nil {
		t.Error("Expected error for forbidden response, got nil")
	}
}

func TestListFulfillmentOrdersPartialOptions(t *testing.T) {
	testCases := []struct {
		name           string
		opts           *FulfillmentOrdersListOptions
		expectedParams map[string]string
	}{
		{
			name:           "page only",
			opts:           &FulfillmentOrdersListOptions{Page: 3},
			expectedParams: map[string]string{"page": "3"},
		},
		{
			name:           "page_size only",
			opts:           &FulfillmentOrdersListOptions{PageSize: 25},
			expectedParams: map[string]string{"page_size": "25"},
		},
		{
			name:           "status only",
			opts:           &FulfillmentOrdersListOptions{Status: "in_progress"},
			expectedParams: map[string]string{"status": "in_progress"},
		},
		{
			name:           "order_id only",
			opts:           &FulfillmentOrdersListOptions{OrderID: "ord_single"},
			expectedParams: map[string]string{"order_id": "ord_single"},
		},
		{
			name: "page and status",
			opts: &FulfillmentOrdersListOptions{Page: 2, Status: "open"},
			expectedParams: map[string]string{
				"page":   "2",
				"status": "open",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				query := r.URL.Query()
				for key, expectedValue := range tc.expectedParams {
					if query.Get(key) != expectedValue {
						t.Errorf("Expected %s=%s, got %s", key, expectedValue, query.Get(key))
					}
				}

				resp := FulfillmentOrdersListResponse{
					Items:      []FulfillmentOrder{},
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

			_, err := client.ListFulfillmentOrders(context.Background(), tc.opts)
			if err != nil {
				t.Fatalf("ListFulfillmentOrders failed: %v", err)
			}
		})
	}
}
