package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompanyCreditsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/company_credits" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := CompanyCreditsListResponse{
			Items: []CompanyCredit{
				{ID: "cc_123", CompanyID: "comp_1", CreditBalance: 5000.00, CreditLimit: 10000.00, Status: "active"},
				{ID: "cc_456", CompanyID: "comp_2", CreditBalance: 2500.00, CreditLimit: 5000.00, Status: "active"},
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

	credits, err := client.ListCompanyCredits(context.Background(), &CompanyCreditsListOptions{})
	if err != nil {
		t.Fatalf("ListCompanyCredits failed: %v", err)
	}

	if len(credits.Items) != 2 {
		t.Errorf("Expected 2 credits, got %d", len(credits.Items))
	}
	if credits.Items[0].ID != "cc_123" {
		t.Errorf("Unexpected credit ID: %s", credits.Items[0].ID)
	}
}

func TestCompanyCreditsListWithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("company_id") != "comp_123" {
			t.Errorf("Expected company_id=comp_123, got %s", r.URL.Query().Get("company_id"))
		}
		if r.URL.Query().Get("status") != "active" {
			t.Errorf("Expected status=active, got %s", r.URL.Query().Get("status"))
		}

		resp := CompanyCreditsListResponse{Items: []CompanyCredit{}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.ListCompanyCredits(context.Background(), &CompanyCreditsListOptions{
		CompanyID: "comp_123",
		Status:    "active",
	})
	if err != nil {
		t.Fatalf("ListCompanyCredits failed: %v", err)
	}
}

func TestCompanyCreditsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/company_credits/cc_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		credit := CompanyCredit{ID: "cc_123", CompanyID: "comp_1", CreditBalance: 5000.00, CreditLimit: 10000.00}
		_ = json.NewEncoder(w).Encode(credit)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	credit, err := client.GetCompanyCredit(context.Background(), "cc_123")
	if err != nil {
		t.Fatalf("GetCompanyCredit failed: %v", err)
	}

	if credit.ID != "cc_123" {
		t.Errorf("Unexpected credit ID: %s", credit.ID)
	}
}

func TestGetCompanyCreditEmptyID(t *testing.T) {
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
			_, err := client.GetCompanyCredit(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "company credit id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCompanyCreditsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/company_credits" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		credit := CompanyCredit{ID: "cc_new", CompanyID: "comp_1", CreditLimit: 10000.00}
		_ = json.NewEncoder(w).Encode(credit)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CompanyCreditCreateRequest{
		CompanyID:   "comp_1",
		CreditLimit: 10000.00,
	}

	credit, err := client.CreateCompanyCredit(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateCompanyCredit failed: %v", err)
	}

	if credit.CreditLimit != 10000.00 {
		t.Errorf("Unexpected credit limit: %.2f", credit.CreditLimit)
	}
}

func TestCompanyCreditsAdjust(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/company_credits/cc_123/adjust" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		credit := CompanyCredit{ID: "cc_123", CreditBalance: 6000.00}
		_ = json.NewEncoder(w).Encode(credit)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CompanyCreditAdjustRequest{
		Amount:      1000.00,
		Description: "Credit top-up",
	}

	credit, err := client.AdjustCompanyCredit(context.Background(), "cc_123", req)
	if err != nil {
		t.Fatalf("AdjustCompanyCredit failed: %v", err)
	}

	if credit.CreditBalance != 6000.00 {
		t.Errorf("Unexpected credit balance: %.2f", credit.CreditBalance)
	}
}

func TestCompanyCreditTransactionsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/company_credits/cc_123/transactions" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := CompanyCreditTransactionsListResponse{
			Items: []CompanyCreditTransaction{
				{ID: "tx_1", CreditID: "cc_123", Type: "credit", Amount: 1000.00, Balance: 6000.00},
				{ID: "tx_2", CreditID: "cc_123", Type: "debit", Amount: -500.00, Balance: 5500.00},
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

	transactions, err := client.ListCompanyCreditTransactions(context.Background(), "cc_123", 1, 20)
	if err != nil {
		t.Fatalf("ListCompanyCreditTransactions failed: %v", err)
	}

	if len(transactions.Items) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(transactions.Items))
	}
}

func TestCompanyCreditsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/company_credits/cc_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteCompanyCredit(context.Background(), "cc_123")
	if err != nil {
		t.Fatalf("DeleteCompanyCredit failed: %v", err)
	}
}

func TestDeleteCompanyCreditEmptyID(t *testing.T) {
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
			err := client.DeleteCompanyCredit(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "company credit id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
