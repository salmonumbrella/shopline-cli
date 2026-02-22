package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFulfillmentsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/fulfillments" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := FulfillmentsListResponse{
			Items: []Fulfillment{
				{
					ID:              "ful_123",
					OrderID:         "ord_456",
					Status:          FulfillmentStatusSuccess,
					TrackingCompany: "FedEx",
					TrackingNumber:  "1234567890",
				},
				{
					ID:              "ful_789",
					OrderID:         "ord_012",
					Status:          FulfillmentStatusPending,
					TrackingCompany: "UPS",
					TrackingNumber:  "0987654321",
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

	fulfillments, err := client.ListFulfillments(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListFulfillments failed: %v", err)
	}

	if len(fulfillments.Items) != 2 {
		t.Errorf("Expected 2 fulfillments, got %d", len(fulfillments.Items))
	}
	if fulfillments.Items[0].ID != "ful_123" {
		t.Errorf("Unexpected fulfillment ID: %s", fulfillments.Items[0].ID)
	}
	if fulfillments.Items[0].Status != FulfillmentStatusSuccess {
		t.Errorf("Unexpected status: %s", fulfillments.Items[0].Status)
	}
}

func TestFulfillmentsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("order_id") != "ord_456" {
			t.Errorf("Expected order_id=ord_456, got %s", r.URL.Query().Get("order_id"))
		}
		if r.URL.Query().Get("status") != "success" {
			t.Errorf("Expected status=success, got %s", r.URL.Query().Get("status"))
		}

		resp := FulfillmentsListResponse{
			Items: []Fulfillment{
				{ID: "ful_123", OrderID: "ord_456", Status: FulfillmentStatusSuccess},
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

	opts := &FulfillmentsListOptions{
		OrderID: "ord_456",
		Status:  "success",
	}
	fulfillments, err := client.ListFulfillments(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListFulfillments failed: %v", err)
	}

	if len(fulfillments.Items) != 1 {
		t.Errorf("Expected 1 fulfillment, got %d", len(fulfillments.Items))
	}
}

func TestFulfillmentsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/fulfillments/ful_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		fulfillment := Fulfillment{
			ID:              "ful_123",
			OrderID:         "ord_456",
			Status:          FulfillmentStatusSuccess,
			TrackingCompany: "FedEx",
			TrackingNumber:  "1234567890",
			TrackingURL:     "https://fedex.com/track/1234567890",
			LineItems: []FulfillmentLineItem{
				{ID: "li_1", ProductID: "prod_1", Title: "Widget", Quantity: 2},
			},
		}
		_ = json.NewEncoder(w).Encode(fulfillment)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	fulfillment, err := client.GetFulfillment(context.Background(), "ful_123")
	if err != nil {
		t.Fatalf("GetFulfillment failed: %v", err)
	}

	if fulfillment.ID != "ful_123" {
		t.Errorf("Unexpected fulfillment ID: %s", fulfillment.ID)
	}
	if fulfillment.TrackingCompany != "FedEx" {
		t.Errorf("Unexpected tracking company: %s", fulfillment.TrackingCompany)
	}
	if len(fulfillment.LineItems) != 1 {
		t.Errorf("Expected 1 line item, got %d", len(fulfillment.LineItems))
	}
}

func TestGetFulfillmentEmptyID(t *testing.T) {
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
			_, err := client.GetFulfillment(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "fulfillment id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
