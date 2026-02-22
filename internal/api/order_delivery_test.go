package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOrderDeliveryGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/orders/ord_123/delivery" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		now := time.Now()
		estimatedAt := now.Add(48 * time.Hour)
		shippedAt := now.Add(-24 * time.Hour)
		delivery := OrderDelivery{
			ID:             "del_123",
			OrderID:        "ord_123",
			Status:         "in_transit",
			TrackingNumber: "1Z999AA10123456784",
			TrackingURL:    "https://tracking.example.com/1Z999AA10123456784",
			Carrier:        "UPS",
			EstimatedAt:    &estimatedAt,
			ShippedAt:      &shippedAt,
			DeliveredAt:    nil,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		_ = json.NewEncoder(w).Encode(delivery)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	delivery, err := client.GetOrderDelivery(context.Background(), "ord_123")
	if err != nil {
		t.Fatalf("GetOrderDelivery failed: %v", err)
	}

	if delivery.ID != "del_123" {
		t.Errorf("Unexpected delivery ID: %s", delivery.ID)
	}
	if delivery.OrderID != "ord_123" {
		t.Errorf("Unexpected order ID: %s", delivery.OrderID)
	}
	if delivery.Status != "in_transit" {
		t.Errorf("Unexpected status: %s", delivery.Status)
	}
	if delivery.TrackingNumber != "1Z999AA10123456784" {
		t.Errorf("Unexpected tracking number: %s", delivery.TrackingNumber)
	}
	if delivery.Carrier != "UPS" {
		t.Errorf("Unexpected carrier: %s", delivery.Carrier)
	}
	if delivery.EstimatedAt == nil {
		t.Error("Expected estimated_at to be set")
	}
	if delivery.ShippedAt == nil {
		t.Error("Expected shipped_at to be set")
	}
	if delivery.DeliveredAt != nil {
		t.Error("Expected delivered_at to be nil")
	}
}

func TestGetOrderDeliveryEmptyID(t *testing.T) {
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
			_, err := client.GetOrderDelivery(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestGetOrderDeliveryAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Order not found",
		})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.GetOrderDelivery(context.Background(), "nonexistent_order")
	if err == nil {
		t.Error("Expected error for API failure, got nil")
	}
}

func TestOrderDeliveryUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/orders/ord_123/delivery" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req OrderDeliveryUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Status == nil || *req.Status != "delivered" {
			t.Errorf("Unexpected status: %v", req.Status)
		}
		if req.TrackingNumber == nil || *req.TrackingNumber != "UPDATED123" {
			t.Errorf("Unexpected tracking number: %v", req.TrackingNumber)
		}
		if req.Carrier == nil || *req.Carrier != "FedEx" {
			t.Errorf("Unexpected carrier: %v", req.Carrier)
		}

		now := time.Now()
		deliveredAt := now
		delivery := OrderDelivery{
			ID:             "del_123",
			OrderID:        "ord_123",
			Status:         *req.Status,
			TrackingNumber: *req.TrackingNumber,
			TrackingURL:    "https://tracking.example.com/UPDATED123",
			Carrier:        *req.Carrier,
			DeliveredAt:    &deliveredAt,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		_ = json.NewEncoder(w).Encode(delivery)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	status := "delivered"
	trackingNumber := "UPDATED123"
	carrier := "FedEx"
	req := &OrderDeliveryUpdateRequest{
		Status:         &status,
		TrackingNumber: &trackingNumber,
		Carrier:        &carrier,
	}

	delivery, err := client.UpdateOrderDelivery(context.Background(), "ord_123", req)
	if err != nil {
		t.Fatalf("UpdateOrderDelivery failed: %v", err)
	}

	if delivery.Status != "delivered" {
		t.Errorf("Unexpected status: %s", delivery.Status)
	}
	if delivery.TrackingNumber != "UPDATED123" {
		t.Errorf("Unexpected tracking number: %s", delivery.TrackingNumber)
	}
	if delivery.Carrier != "FedEx" {
		t.Errorf("Unexpected carrier: %s", delivery.Carrier)
	}
	if delivery.DeliveredAt == nil {
		t.Error("Expected delivered_at to be set")
	}
}

func TestUpdateOrderDeliveryEmptyID(t *testing.T) {
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
			_, err := client.UpdateOrderDelivery(context.Background(), tc.id, &OrderDeliveryUpdateRequest{})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestUpdateOrderDeliveryAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid update request",
		})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	status := "invalid_status"
	req := &OrderDeliveryUpdateRequest{
		Status: &status,
	}

	_, err := client.UpdateOrderDelivery(context.Background(), "ord_123", req)
	if err == nil {
		t.Error("Expected error for API failure, got nil")
	}
}

func TestUpdateOrderDeliveryPartialUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}

		var req OrderDeliveryUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		// Only tracking URL should be set
		if req.TrackingURL == nil || *req.TrackingURL != "https://newtrack.example.com/123" {
			t.Errorf("Unexpected tracking URL: %v", req.TrackingURL)
		}
		if req.Status != nil {
			t.Errorf("Expected status to be nil, got %v", req.Status)
		}
		if req.TrackingNumber != nil {
			t.Errorf("Expected tracking number to be nil, got %v", req.TrackingNumber)
		}
		if req.Carrier != nil {
			t.Errorf("Expected carrier to be nil, got %v", req.Carrier)
		}

		now := time.Now()
		delivery := OrderDelivery{
			ID:             "del_123",
			OrderID:        "ord_123",
			Status:         "in_transit",
			TrackingNumber: "ABC123",
			TrackingURL:    *req.TrackingURL,
			Carrier:        "UPS",
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		_ = json.NewEncoder(w).Encode(delivery)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	trackingURL := "https://newtrack.example.com/123"
	req := &OrderDeliveryUpdateRequest{
		TrackingURL: &trackingURL,
	}

	delivery, err := client.UpdateOrderDelivery(context.Background(), "ord_123", req)
	if err != nil {
		t.Fatalf("UpdateOrderDelivery failed: %v", err)
	}

	if delivery.TrackingURL != "https://newtrack.example.com/123" {
		t.Errorf("Unexpected tracking URL: %s", delivery.TrackingURL)
	}
}

func TestUpdateOrderDeliveryWithEstimatedAt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}

		var req OrderDeliveryUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.EstimatedAt == nil {
			t.Error("Expected estimated_at to be set")
		}

		now := time.Now()
		delivery := OrderDelivery{
			ID:          "del_123",
			OrderID:     "ord_123",
			Status:      "in_transit",
			EstimatedAt: req.EstimatedAt,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		_ = json.NewEncoder(w).Encode(delivery)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	estimatedAt := time.Now().Add(72 * time.Hour)
	req := &OrderDeliveryUpdateRequest{
		EstimatedAt: &estimatedAt,
	}

	delivery, err := client.UpdateOrderDelivery(context.Background(), "ord_123", req)
	if err != nil {
		t.Fatalf("UpdateOrderDelivery failed: %v", err)
	}

	if delivery.EstimatedAt == nil {
		t.Error("Expected estimated_at to be set in response")
	}
}

func TestGetOrderDeliveryWithNoDeliveryInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		// Order exists but has minimal delivery info
		delivery := OrderDelivery{
			ID:        "del_123",
			OrderID:   "ord_123",
			Status:    "pending",
			CreatedAt: now,
			UpdatedAt: now,
		}
		_ = json.NewEncoder(w).Encode(delivery)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	delivery, err := client.GetOrderDelivery(context.Background(), "ord_123")
	if err != nil {
		t.Fatalf("GetOrderDelivery failed: %v", err)
	}

	if delivery.Status != "pending" {
		t.Errorf("Unexpected status: %s", delivery.Status)
	}
	if delivery.TrackingNumber != "" {
		t.Errorf("Expected empty tracking number, got: %s", delivery.TrackingNumber)
	}
	if delivery.Carrier != "" {
		t.Errorf("Expected empty carrier, got: %s", delivery.Carrier)
	}
}
