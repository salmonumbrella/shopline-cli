package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetMemberPoints(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/customers/cust_123/points" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		points := MemberPoints{
			CustomerID:      "cust_123",
			TotalPoints:     5000,
			AvailablePoints: 4500,
			PendingPoints:   500,
		}
		_ = json.NewEncoder(w).Encode(points)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	points, err := client.GetMemberPoints(context.Background(), "cust_123")
	if err != nil {
		t.Fatalf("GetMemberPoints failed: %v", err)
	}

	if points.CustomerID != "cust_123" {
		t.Errorf("Unexpected customer ID: %s", points.CustomerID)
	}
	if points.TotalPoints != 5000 {
		t.Errorf("Unexpected total points: %d", points.TotalPoints)
	}
}

func TestGetMemberPointsEmptyID(t *testing.T) {
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
			_, err := client.GetMemberPoints(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "customer id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestListPointsTransactions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/customers/cust_123/points/transactions" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := PointsTransactionsListResponse{
			Items: []PointsTransaction{
				{ID: "txn_123", CustomerID: "cust_123", Type: "earn", Points: 100},
				{ID: "txn_456", CustomerID: "cust_123", Type: "redeem", Points: -50},
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

	txns, err := client.ListPointsTransactions(context.Background(), "cust_123", nil)
	if err != nil {
		t.Fatalf("ListPointsTransactions failed: %v", err)
	}

	if len(txns.Items) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(txns.Items))
	}
	if txns.Items[0].ID != "txn_123" {
		t.Errorf("Unexpected transaction ID: %s", txns.Items[0].ID)
	}
}

func TestListPointsTransactionsEmptyID(t *testing.T) {
	client := NewClient("token")

	_, err := client.ListPointsTransactions(context.Background(), "", nil)
	if err == nil {
		t.Error("Expected error for empty customer ID, got nil")
	}
	if err != nil && err.Error() != "customer id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestAdjustMemberPoints(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/customers/cust_123/points/adjust" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		points := MemberPoints{
			CustomerID:      "cust_123",
			TotalPoints:     5100,
			AvailablePoints: 4600,
		}
		_ = json.NewEncoder(w).Encode(points)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	points, err := client.AdjustMemberPoints(context.Background(), "cust_123", 100, "Bonus points")
	if err != nil {
		t.Fatalf("AdjustMemberPoints failed: %v", err)
	}

	if points.TotalPoints != 5100 {
		t.Errorf("Unexpected total points: %d", points.TotalPoints)
	}
}

func TestAdjustMemberPointsEmptyID(t *testing.T) {
	client := NewClient("token")

	_, err := client.AdjustMemberPoints(context.Background(), "", 100, "test")
	if err == nil {
		t.Error("Expected error for empty customer ID, got nil")
	}
	if err != nil && err.Error() != "customer id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}
