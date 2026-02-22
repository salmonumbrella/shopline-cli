package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProductSubscriptionsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/product_subscriptions" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := ProductSubscriptionsListResponse{
			Items: []ProductSubscription{
				{
					ID:                "sub_123",
					ProductID:         "prod_456",
					CustomerID:        "cust_789",
					SellingPlanID:     "sp_001",
					Status:            "active",
					Frequency:         "monthly",
					FrequencyInterval: 1,
					Price:             "19.99",
					Currency:          "USD",
					Quantity:          1,
				},
				{
					ID:                "sub_456",
					ProductID:         "prod_789",
					CustomerID:        "cust_123",
					SellingPlanID:     "sp_002",
					Status:            "active",
					Frequency:         "weekly",
					FrequencyInterval: 2,
					Price:             "9.99",
					Currency:          "USD",
					Quantity:          2,
				},
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

	subscriptions, err := client.ListProductSubscriptions(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListProductSubscriptions failed: %v", err)
	}

	if len(subscriptions.Items) != 2 {
		t.Errorf("Expected 2 subscriptions, got %d", len(subscriptions.Items))
	}
	if subscriptions.Items[0].ID != "sub_123" {
		t.Errorf("Unexpected subscription ID: %s", subscriptions.Items[0].ID)
	}
	if subscriptions.Items[0].Status != "active" {
		t.Errorf("Unexpected status: %s", subscriptions.Items[0].Status)
	}
}

func TestProductSubscriptionsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("customer_id") != "cust_123" {
			t.Errorf("Expected customer_id=cust_123, got %s", r.URL.Query().Get("customer_id"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("status") != "active" {
			t.Errorf("Expected status=active, got %s", r.URL.Query().Get("status"))
		}

		resp := ProductSubscriptionsListResponse{
			Items: []ProductSubscription{
				{ID: "sub_123", CustomerID: "cust_123", Status: "active"},
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

	opts := &ProductSubscriptionsListOptions{
		Page:       2,
		CustomerID: "cust_123",
		Status:     "active",
	}
	subscriptions, err := client.ListProductSubscriptions(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListProductSubscriptions failed: %v", err)
	}

	if len(subscriptions.Items) != 1 {
		t.Errorf("Expected 1 subscription, got %d", len(subscriptions.Items))
	}
}

func TestProductSubscriptionsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/product_subscriptions/sub_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		subscription := ProductSubscription{
			ID:                "sub_123",
			ProductID:         "prod_456",
			VariantID:         "var_789",
			CustomerID:        "cust_001",
			SellingPlanID:     "sp_001",
			Status:            "active",
			Frequency:         "monthly",
			FrequencyInterval: 1,
			Price:             "29.99",
			Currency:          "USD",
			Quantity:          1,
			TotalCycles:       12,
			CompletedCycles:   3,
		}
		_ = json.NewEncoder(w).Encode(subscription)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	subscription, err := client.GetProductSubscription(context.Background(), "sub_123")
	if err != nil {
		t.Fatalf("GetProductSubscription failed: %v", err)
	}

	if subscription.ID != "sub_123" {
		t.Errorf("Unexpected subscription ID: %s", subscription.ID)
	}
	if subscription.Frequency != "monthly" {
		t.Errorf("Unexpected frequency: %s", subscription.Frequency)
	}
	if subscription.CompletedCycles != 3 {
		t.Errorf("Unexpected completed cycles: %d", subscription.CompletedCycles)
	}
}

func TestProductSubscriptionsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/product_subscriptions" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ProductSubscriptionCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.ProductID != "prod_123" {
			t.Errorf("Unexpected product ID: %s", req.ProductID)
		}
		if req.CustomerID != "cust_456" {
			t.Errorf("Unexpected customer ID: %s", req.CustomerID)
		}
		if req.SellingPlanID != "sp_001" {
			t.Errorf("Unexpected selling plan ID: %s", req.SellingPlanID)
		}

		subscription := ProductSubscription{
			ID:            "sub_new",
			ProductID:     req.ProductID,
			CustomerID:    req.CustomerID,
			SellingPlanID: req.SellingPlanID,
			Status:        "active",
			Quantity:      req.Quantity,
		}
		_ = json.NewEncoder(w).Encode(subscription)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ProductSubscriptionCreateRequest{
		ProductID:     "prod_123",
		CustomerID:    "cust_456",
		SellingPlanID: "sp_001",
		Quantity:      1,
	}
	subscription, err := client.CreateProductSubscription(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateProductSubscription failed: %v", err)
	}

	if subscription.ID != "sub_new" {
		t.Errorf("Unexpected subscription ID: %s", subscription.ID)
	}
	if subscription.Status != "active" {
		t.Errorf("Unexpected status: %s", subscription.Status)
	}
}

func TestProductSubscriptionsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/product_subscriptions/sub_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteProductSubscription(context.Background(), "sub_123")
	if err != nil {
		t.Fatalf("DeleteProductSubscription failed: %v", err)
	}
}

func TestGetProductSubscriptionEmptyID(t *testing.T) {
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
			_, err := client.GetProductSubscription(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "product subscription id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteProductSubscriptionEmptyID(t *testing.T) {
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
			err := client.DeleteProductSubscription(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "product subscription id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
