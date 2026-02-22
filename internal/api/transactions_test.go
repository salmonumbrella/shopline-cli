package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTransactionsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/transactions" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := TransactionsListResponse{
			Items: []Transaction{
				{ID: "txn_123", OrderID: "ord_123", Kind: "sale", Status: "success", Amount: "99.99"},
				{ID: "txn_456", OrderID: "ord_456", Kind: "refund", Status: "success", Amount: "50.00"},
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

	transactions, err := client.ListTransactions(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTransactions failed: %v", err)
	}

	if len(transactions.Items) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(transactions.Items))
	}
	if transactions.Items[0].ID != "txn_123" {
		t.Errorf("Unexpected transaction ID: %s", transactions.Items[0].ID)
	}
}

func TestTransactionsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transactions/txn_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		transaction := Transaction{ID: "txn_123", OrderID: "ord_123", Kind: "sale", Amount: "99.99"}
		_ = json.NewEncoder(w).Encode(transaction)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	transaction, err := client.GetTransaction(context.Background(), "txn_123")
	if err != nil {
		t.Fatalf("GetTransaction failed: %v", err)
	}

	if transaction.ID != "txn_123" {
		t.Errorf("Unexpected transaction ID: %s", transaction.ID)
	}
}

func TestGetTransactionEmptyID(t *testing.T) {
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
			_, err := client.GetTransaction(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "transaction id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestListOrderTransactions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/orders/ord_123/transactions" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := TransactionsListResponse{
			Items: []Transaction{
				{ID: "txn_123", OrderID: "ord_123", Kind: "sale", Status: "success"},
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

	transactions, err := client.ListOrderTransactions(context.Background(), "ord_123")
	if err != nil {
		t.Fatalf("ListOrderTransactions failed: %v", err)
	}

	if len(transactions.Items) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(transactions.Items))
	}
}

func TestListOrderTransactionsEmptyID(t *testing.T) {
	client := NewClient("token")

	_, err := client.ListOrderTransactions(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty order ID, got nil")
	}
	if err != nil && err.Error() != "order id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestListTransactionsServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.ListTransactions(context.Background(), nil)
	if err == nil {
		t.Fatal("Expected error for server error response, got nil")
	}
}

func TestGetTransactionServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "transaction not found"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.GetTransaction(context.Background(), "txn_nonexistent")
	if err == nil {
		t.Fatal("Expected error for not found response, got nil")
	}
}

func TestListOrderTransactionsServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid order"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.ListOrderTransactions(context.Background(), "ord_invalid")
	if err == nil {
		t.Fatal("Expected error for bad request response, got nil")
	}
}

func TestListOrderTransactionsWhitespaceID(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name    string
		orderID string
	}{
		{"whitespace only", "   "},
		{"tab only", "\t"},
		{"newline", "\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.ListOrderTransactions(context.Background(), tc.orderID)
			if err == nil {
				t.Error("Expected error for whitespace order ID, got nil")
			}
			if err != nil && err.Error() != "order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestTransactionFieldsParsed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		transaction := Transaction{
			ID:            "txn_full",
			OrderID:       "ord_full",
			Kind:          "sale",
			Gateway:       "stripe",
			Status:        "success",
			Amount:        "199.99",
			Currency:      "USD",
			Authorization: "auth_123",
			ParentID:      "txn_parent",
			ErrorCode:     "",
			Message:       "Transaction approved",
			ProcessedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			CreatedAt:     time.Date(2024, 1, 15, 10, 29, 0, 0, time.UTC),
		}
		_ = json.NewEncoder(w).Encode(transaction)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	transaction, err := client.GetTransaction(context.Background(), "txn_full")
	if err != nil {
		t.Fatalf("GetTransaction failed: %v", err)
	}

	if transaction.ID != "txn_full" {
		t.Errorf("Expected ID txn_full, got %s", transaction.ID)
	}
	if transaction.OrderID != "ord_full" {
		t.Errorf("Expected OrderID ord_full, got %s", transaction.OrderID)
	}
	if transaction.Kind != "sale" {
		t.Errorf("Expected Kind sale, got %s", transaction.Kind)
	}
	if transaction.Gateway != "stripe" {
		t.Errorf("Expected Gateway stripe, got %s", transaction.Gateway)
	}
	if transaction.Status != "success" {
		t.Errorf("Expected Status success, got %s", transaction.Status)
	}
	if transaction.Amount != "199.99" {
		t.Errorf("Expected Amount 199.99, got %s", transaction.Amount)
	}
	if transaction.Currency != "USD" {
		t.Errorf("Expected Currency USD, got %s", transaction.Currency)
	}
	if transaction.Authorization != "auth_123" {
		t.Errorf("Expected Authorization auth_123, got %s", transaction.Authorization)
	}
	if transaction.ParentID != "txn_parent" {
		t.Errorf("Expected ParentID txn_parent, got %s", transaction.ParentID)
	}
	if transaction.Message != "Transaction approved" {
		t.Errorf("Expected Message 'Transaction approved', got %s", transaction.Message)
	}
}

func TestListTransactionsEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := TransactionsListResponse{
			Items:      []Transaction{},
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

	transactions, err := client.ListTransactions(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTransactions failed: %v", err)
	}

	if len(transactions.Items) != 0 {
		t.Errorf("Expected 0 transactions, got %d", len(transactions.Items))
	}
	if transactions.TotalCount != 0 {
		t.Errorf("Expected TotalCount 0, got %d", transactions.TotalCount)
	}
}

func TestListTransactionsWithOptions(t *testing.T) {
	testCases := []struct {
		name           string
		opts           *TransactionsListOptions
		expectedParams map[string]string
	}{
		{
			name: "page and page_size",
			opts: &TransactionsListOptions{
				Page:     2,
				PageSize: 50,
			},
			expectedParams: map[string]string{
				"page":      "2",
				"page_size": "50",
			},
		},
		{
			name: "order_id filter",
			opts: &TransactionsListOptions{
				OrderID: "ord_999",
			},
			expectedParams: map[string]string{
				"order_id": "ord_999",
			},
		},
		{
			name: "status filter",
			opts: &TransactionsListOptions{
				Status: "success",
			},
			expectedParams: map[string]string{
				"status": "success",
			},
		},
		{
			name: "kind filter",
			opts: &TransactionsListOptions{
				Kind: "refund",
			},
			expectedParams: map[string]string{
				"kind": "refund",
			},
		},
		{
			name: "all options combined",
			opts: &TransactionsListOptions{
				Page:     3,
				PageSize: 25,
				OrderID:  "ord_abc",
				Status:   "pending",
				Kind:     "sale",
			},
			expectedParams: map[string]string{
				"page":      "3",
				"page_size": "25",
				"order_id":  "ord_abc",
				"status":    "pending",
				"kind":      "sale",
			},
		},
		{
			name: "zero values are not included",
			opts: &TransactionsListOptions{
				Page:     0,
				PageSize: 0,
				OrderID:  "",
				Status:   "success",
				Kind:     "",
			},
			expectedParams: map[string]string{
				"status": "success",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET, got %s", r.Method)
				}
				if r.URL.Path != "/transactions" {
					t.Errorf("Unexpected path: %s", r.URL.Path)
				}

				// Verify query parameters
				query := r.URL.Query()
				for key, expectedValue := range tc.expectedParams {
					actualValue := query.Get(key)
					if actualValue != expectedValue {
						t.Errorf("Expected %s=%s, got %s=%s", key, expectedValue, key, actualValue)
					}
				}

				// Verify no unexpected parameters are present
				for key := range query {
					if _, expected := tc.expectedParams[key]; !expected {
						t.Errorf("Unexpected query parameter: %s=%s", key, query.Get(key))
					}
				}

				resp := TransactionsListResponse{
					Items: []Transaction{
						{ID: "txn_test", OrderID: "ord_test", Kind: "sale", Status: "success"},
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

			transactions, err := client.ListTransactions(context.Background(), tc.opts)
			if err != nil {
				t.Fatalf("ListTransactions failed: %v", err)
			}

			if len(transactions.Items) != 1 {
				t.Errorf("Expected 1 transaction, got %d", len(transactions.Items))
			}
		})
	}
}
