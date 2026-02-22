package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPayoutsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/payouts" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := PayoutsListResponse{
			Items: []Payout{
				{ID: "po_123", Amount: "1000.00", Currency: "USD", Status: "paid"},
				{ID: "po_456", Amount: "500.00", Currency: "USD", Status: "pending"},
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

	payouts, err := client.ListPayouts(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListPayouts failed: %v", err)
	}

	if len(payouts.Items) != 2 {
		t.Errorf("Expected 2 payouts, got %d", len(payouts.Items))
	}
	if payouts.Items[0].ID != "po_123" {
		t.Errorf("Unexpected payout ID: %s", payouts.Items[0].ID)
	}
}

func TestPayoutsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("status") != "paid" {
			t.Errorf("Expected status=paid, got %s", r.URL.Query().Get("status"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}

		resp := PayoutsListResponse{
			Items:      []Payout{},
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

	opts := &PayoutsListOptions{
		Page:   2,
		Status: "paid",
	}
	_, err := client.ListPayouts(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListPayouts failed: %v", err)
	}
}

func TestPayoutsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/payouts/po_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		payout := Payout{ID: "po_123", Amount: "1000.00", Currency: "USD", Status: "paid"}
		_ = json.NewEncoder(w).Encode(payout)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	payout, err := client.GetPayout(context.Background(), "po_123")
	if err != nil {
		t.Fatalf("GetPayout failed: %v", err)
	}

	if payout.ID != "po_123" {
		t.Errorf("Unexpected payout ID: %s", payout.ID)
	}
}

func TestGetPayoutEmptyID(t *testing.T) {
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
			_, err := client.GetPayout(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "payout id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
