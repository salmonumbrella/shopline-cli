package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAbandonedCheckoutsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/abandoned_checkouts" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := AbandonedCheckoutsListResponse{
			Items: []AbandonedCheckout{
				{ID: "ac_123", Email: "test@example.com", TotalPrice: "99.99"},
				{ID: "ac_456", Email: "user@example.com", TotalPrice: "149.99"},
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

	checkouts, err := client.ListAbandonedCheckouts(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListAbandonedCheckouts failed: %v", err)
	}

	if len(checkouts.Items) != 2 {
		t.Errorf("Expected 2 checkouts, got %d", len(checkouts.Items))
	}
	if checkouts.Items[0].ID != "ac_123" {
		t.Errorf("Unexpected checkout ID: %s", checkouts.Items[0].ID)
	}
}

func TestAbandonedCheckoutsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/abandoned_checkouts/ac_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		checkout := AbandonedCheckout{ID: "ac_123", Email: "test@example.com", TotalPrice: "99.99"}
		_ = json.NewEncoder(w).Encode(checkout)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	checkout, err := client.GetAbandonedCheckout(context.Background(), "ac_123")
	if err != nil {
		t.Fatalf("GetAbandonedCheckout failed: %v", err)
	}

	if checkout.ID != "ac_123" {
		t.Errorf("Unexpected checkout ID: %s", checkout.ID)
	}
}

func TestGetAbandonedCheckoutEmptyID(t *testing.T) {
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
			_, err := client.GetAbandonedCheckout(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "abandoned checkout id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestSendAbandonedCheckoutRecoveryEmail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/abandoned_checkouts/ac_123/send_recovery_email" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.SendAbandonedCheckoutRecoveryEmail(context.Background(), "ac_123")
	if err != nil {
		t.Fatalf("SendAbandonedCheckoutRecoveryEmail failed: %v", err)
	}
}

func TestSendAbandonedCheckoutRecoveryEmailEmptyID(t *testing.T) {
	client := NewClient("token")

	err := client.SendAbandonedCheckoutRecoveryEmail(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
	if err != nil && err.Error() != "abandoned checkout id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}
