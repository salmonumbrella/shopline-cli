package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStoreCreditsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/store_credits" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := StoreCreditsListResponse{
			Items: []StoreCredit{
				{ID: "sc_123", CustomerID: "cust_123", Amount: "100.00", Balance: "75.00", Currency: "USD"},
				{ID: "sc_456", CustomerID: "cust_456", Amount: "50.00", Balance: "50.00", Currency: "USD"},
			},
			Page:       1,
			PageSize:   20,
			TotalCount: 2,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("test", "token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	credits, err := client.ListStoreCredits(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListStoreCredits failed: %v", err)
	}

	if len(credits.Items) != 2 {
		t.Errorf("Expected 2 credits, got %d", len(credits.Items))
	}
	if credits.Items[0].ID != "sc_123" {
		t.Errorf("Unexpected credit ID: %s", credits.Items[0].ID)
	}
	if credits.Items[0].Balance != "75.00" {
		t.Errorf("Unexpected balance: %s", credits.Items[0].Balance)
	}
}

func TestStoreCreditsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("customer_id") != "cust_123" {
			t.Errorf("Expected customer_id=cust_123, got %s", r.URL.Query().Get("customer_id"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}

		resp := StoreCreditsListResponse{
			Items: []StoreCredit{
				{ID: "sc_123", CustomerID: "cust_123", Amount: "100.00"},
			},
			Page:       2,
			PageSize:   20,
			TotalCount: 1,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("test", "token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &StoreCreditsListOptions{
		Page:       2,
		CustomerID: "cust_123",
	}
	credits, err := client.ListStoreCredits(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListStoreCredits failed: %v", err)
	}

	if len(credits.Items) != 1 {
		t.Errorf("Expected 1 credit, got %d", len(credits.Items))
	}
}

func TestStoreCreditsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/store_credits/sc_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		credit := StoreCredit{
			ID:          "sc_123",
			CustomerID:  "cust_123",
			Amount:      "100.00",
			Balance:     "75.00",
			Currency:    "USD",
			Description: "Loyalty reward",
		}
		_ = json.NewEncoder(w).Encode(credit)
	}))
	defer server.Close()

	client := NewClient("test", "token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	credit, err := client.GetStoreCredit(context.Background(), "sc_123")
	if err != nil {
		t.Fatalf("GetStoreCredit failed: %v", err)
	}

	if credit.ID != "sc_123" {
		t.Errorf("Unexpected credit ID: %s", credit.ID)
	}
	if credit.Description != "Loyalty reward" {
		t.Errorf("Unexpected description: %s", credit.Description)
	}
}

func TestStoreCreditsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/store_credits" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req StoreCreditCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.CustomerID != "cust_123" {
			t.Errorf("Unexpected customer ID: %s", req.CustomerID)
		}
		if req.Amount != "100.00" {
			t.Errorf("Unexpected amount: %s", req.Amount)
		}

		credit := StoreCredit{
			ID:         "sc_new",
			CustomerID: req.CustomerID,
			Amount:     req.Amount,
			Balance:    req.Amount,
			Currency:   req.Currency,
		}
		_ = json.NewEncoder(w).Encode(credit)
	}))
	defer server.Close()

	client := NewClient("test", "token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &StoreCreditCreateRequest{
		CustomerID: "cust_123",
		Amount:     "100.00",
		Currency:   "USD",
	}
	credit, err := client.CreateStoreCredit(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateStoreCredit failed: %v", err)
	}

	if credit.ID != "sc_new" {
		t.Errorf("Unexpected credit ID: %s", credit.ID)
	}
}

func TestStoreCreditsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/store_credits/sc_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("test", "token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteStoreCredit(context.Background(), "sc_123")
	if err != nil {
		t.Fatalf("DeleteStoreCredit failed: %v", err)
	}
}

func TestGetStoreCreditEmptyID(t *testing.T) {
	client := NewClient("test", "token")

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
			_, err := client.GetStoreCredit(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "store credit id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteStoreCreditEmptyID(t *testing.T) {
	client := NewClient("test", "token")

	testCases := []struct {
		name string
		id   string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.DeleteStoreCredit(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "store credit id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
