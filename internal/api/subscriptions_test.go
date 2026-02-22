package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSubscriptionsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/subscriptions" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := SubscriptionsListResponse{
			Items: []Subscription{
				{ID: "sub_123", CustomerID: "cust_123", ProductID: "prod_123", Status: SubscriptionStatusActive, Interval: "month"},
				{ID: "sub_456", CustomerID: "cust_456", ProductID: "prod_456", Status: SubscriptionStatusPaused, Interval: "week"},
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

	subscriptions, err := client.ListSubscriptions(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListSubscriptions failed: %v", err)
	}

	if len(subscriptions.Items) != 2 {
		t.Errorf("Expected 2 subscriptions, got %d", len(subscriptions.Items))
	}
	if subscriptions.Items[0].ID != "sub_123" {
		t.Errorf("Unexpected subscription ID: %s", subscriptions.Items[0].ID)
	}
	if subscriptions.Items[0].Status != SubscriptionStatusActive {
		t.Errorf("Unexpected status: %s", subscriptions.Items[0].Status)
	}
}

func TestSubscriptionsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("customer_id") != "cust_123" {
			t.Errorf("Expected customer_id=cust_123, got %s", r.URL.Query().Get("customer_id"))
		}
		if r.URL.Query().Get("status") != "active" {
			t.Errorf("Expected status=active, got %s", r.URL.Query().Get("status"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}

		resp := SubscriptionsListResponse{
			Items: []Subscription{
				{ID: "sub_123", CustomerID: "cust_123", Status: SubscriptionStatusActive},
			},
			Page:       2,
			PageSize:   20,
			TotalCount: 1,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &SubscriptionsListOptions{
		Page:       2,
		CustomerID: "cust_123",
		Status:     "active",
	}
	subscriptions, err := client.ListSubscriptions(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListSubscriptions failed: %v", err)
	}

	if len(subscriptions.Items) != 1 {
		t.Errorf("Expected 1 subscription, got %d", len(subscriptions.Items))
	}
}

func TestSubscriptionsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/subscriptions/sub_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		subscription := Subscription{
			ID:            "sub_123",
			CustomerID:    "cust_123",
			ProductID:     "prod_123",
			Status:        SubscriptionStatusActive,
			Interval:      "month",
			IntervalCount: 1,
			Price:         "29.99",
			Currency:      "USD",
		}
		_ = json.NewEncoder(w).Encode(subscription)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	subscription, err := client.GetSubscription(context.Background(), "sub_123")
	if err != nil {
		t.Fatalf("GetSubscription failed: %v", err)
	}

	if subscription.ID != "sub_123" {
		t.Errorf("Unexpected subscription ID: %s", subscription.ID)
	}
	if subscription.Interval != "month" {
		t.Errorf("Unexpected interval: %s", subscription.Interval)
	}
	if subscription.Price != "29.99" {
		t.Errorf("Unexpected price: %s", subscription.Price)
	}
}

func TestSubscriptionsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/subscriptions" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req SubscriptionCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.CustomerID != "cust_123" {
			t.Errorf("Unexpected customer ID: %s", req.CustomerID)
		}
		if req.ProductID != "prod_123" {
			t.Errorf("Unexpected product ID: %s", req.ProductID)
		}
		if req.Interval != "month" {
			t.Errorf("Unexpected interval: %s", req.Interval)
		}

		subscription := Subscription{
			ID:         "sub_new",
			CustomerID: req.CustomerID,
			ProductID:  req.ProductID,
			Interval:   req.Interval,
			Status:     SubscriptionStatusActive,
		}
		_ = json.NewEncoder(w).Encode(subscription)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &SubscriptionCreateRequest{
		CustomerID: "cust_123",
		ProductID:  "prod_123",
		Interval:   "month",
	}
	subscription, err := client.CreateSubscription(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateSubscription failed: %v", err)
	}

	if subscription.ID != "sub_new" {
		t.Errorf("Unexpected subscription ID: %s", subscription.ID)
	}
}

func TestSubscriptionsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/subscriptions/sub_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteSubscription(context.Background(), "sub_123")
	if err != nil {
		t.Fatalf("DeleteSubscription failed: %v", err)
	}
}

func TestGetSubscriptionEmptyID(t *testing.T) {
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
			_, err := client.GetSubscription(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "subscription id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteSubscriptionEmptyID(t *testing.T) {
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
			err := client.DeleteSubscription(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "subscription id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
