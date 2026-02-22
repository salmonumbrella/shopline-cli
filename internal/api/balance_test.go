package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetBalance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/balance" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		balance := Balance{
			Currency:  "USD",
			Available: "5000.00",
			Pending:   "1000.00",
			Total:     "6000.00",
		}
		_ = json.NewEncoder(w).Encode(balance)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	balance, err := client.GetBalance(context.Background())
	if err != nil {
		t.Fatalf("GetBalance failed: %v", err)
	}

	if balance.Currency != "USD" {
		t.Errorf("Expected USD, got %s", balance.Currency)
	}
	if balance.Available != "5000.00" {
		t.Errorf("Expected 5000.00, got %s", balance.Available)
	}
}

func TestBalanceTransactionsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/balance/transactions" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := BalanceTransactionsListResponse{
			Items: []BalanceTransaction{
				{ID: "bt_123", Type: "payment", Amount: "100.00", Currency: "USD", Status: "available"},
				{ID: "bt_456", Type: "refund", Amount: "-50.00", Currency: "USD", Status: "pending"},
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

	transactions, err := client.ListBalanceTransactions(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListBalanceTransactions failed: %v", err)
	}

	if len(transactions.Items) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(transactions.Items))
	}
	if transactions.Items[0].ID != "bt_123" {
		t.Errorf("Unexpected transaction ID: %s", transactions.Items[0].ID)
	}
}

func TestBalanceTransactionsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("type") != "payment" {
			t.Errorf("Expected type=payment, got %s", r.URL.Query().Get("type"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}

		resp := BalanceTransactionsListResponse{
			Items:      []BalanceTransaction{},
			Page:       2,
			PageSize:   20,
			TotalCount: 0,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &BalanceTransactionsListOptions{
		Page: 2,
		Type: "payment",
	}
	_, err := client.ListBalanceTransactions(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListBalanceTransactions failed: %v", err)
	}
}

func TestGetBalanceTransaction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/balance/transactions/bt_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		txn := BalanceTransaction{ID: "bt_123", Type: "payment", Amount: "100.00", Currency: "USD"}
		_ = json.NewEncoder(w).Encode(txn)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	txn, err := client.GetBalanceTransaction(context.Background(), "bt_123")
	if err != nil {
		t.Fatalf("GetBalanceTransaction failed: %v", err)
	}

	if txn.ID != "bt_123" {
		t.Errorf("Unexpected transaction ID: %s", txn.ID)
	}
}

func TestGetBalanceTransactionEmptyID(t *testing.T) {
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
			_, err := client.GetBalanceTransaction(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "balance transaction id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
