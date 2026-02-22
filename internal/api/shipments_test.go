package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestShipmentsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/shipments" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := ShipmentsListResponse{
			Items: []Shipment{
				{ID: "ship_123", OrderID: "ord_123", TrackingNumber: "1Z999AA10123456784", Status: "in_transit"},
				{ID: "ship_456", OrderID: "ord_456", TrackingNumber: "9400111899223456789012", Status: "delivered"},
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

	shipments, err := client.ListShipments(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListShipments failed: %v", err)
	}

	if len(shipments.Items) != 2 {
		t.Errorf("Expected 2 shipments, got %d", len(shipments.Items))
	}
	if shipments.Items[0].ID != "ship_123" {
		t.Errorf("Unexpected shipment ID: %s", shipments.Items[0].ID)
	}
}

func TestShipmentsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/shipments/ship_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		shipment := Shipment{ID: "ship_123", OrderID: "ord_123", TrackingNumber: "1Z999AA10123456784"}
		_ = json.NewEncoder(w).Encode(shipment)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	shipment, err := client.GetShipment(context.Background(), "ship_123")
	if err != nil {
		t.Fatalf("GetShipment failed: %v", err)
	}

	if shipment.ID != "ship_123" {
		t.Errorf("Unexpected shipment ID: %s", shipment.ID)
	}
}

func TestGetShipmentEmptyID(t *testing.T) {
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
			_, err := client.GetShipment(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "shipment id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestShipmentsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/shipments" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		shipment := Shipment{ID: "ship_new", OrderID: "ord_123", TrackingNumber: "ABC123456"}
		_ = json.NewEncoder(w).Encode(shipment)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ShipmentCreateRequest{OrderID: "ord_123", FulfillmentID: "ful_123", TrackingCompany: "UPS", TrackingNumber: "ABC123456"}
	shipment, err := client.CreateShipment(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateShipment failed: %v", err)
	}

	if shipment.ID != "ship_new" {
		t.Errorf("Unexpected shipment ID: %s", shipment.ID)
	}
}

func TestShipmentsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/shipments/ship_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteShipment(context.Background(), "ship_123")
	if err != nil {
		t.Fatalf("DeleteShipment failed: %v", err)
	}
}

func TestDeleteShipmentEmptyID(t *testing.T) {
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
			err := client.DeleteShipment(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "shipment id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestShipmentsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/shipments/ship_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		shipment := Shipment{ID: "ship_123", TrackingNumber: "UPDATED123", Status: "delivered"}
		_ = json.NewEncoder(w).Encode(shipment)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ShipmentUpdateRequest{TrackingNumber: "UPDATED123", Status: "delivered"}
	shipment, err := client.UpdateShipment(context.Background(), "ship_123", req)
	if err != nil {
		t.Fatalf("UpdateShipment failed: %v", err)
	}

	if shipment.TrackingNumber != "UPDATED123" {
		t.Errorf("Unexpected tracking number: %s", shipment.TrackingNumber)
	}
}

func TestUpdateShipmentEmptyID(t *testing.T) {
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
			_, err := client.UpdateShipment(context.Background(), tc.id, &ShipmentUpdateRequest{})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "shipment id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestShipmentsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		query := r.URL.Query()
		if query.Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", query.Get("page"))
		}
		if query.Get("page_size") != "10" {
			t.Errorf("Expected page_size=10, got %s", query.Get("page_size"))
		}
		if query.Get("order_id") != "ord_123" {
			t.Errorf("Expected order_id=ord_123, got %s", query.Get("order_id"))
		}
		if query.Get("fulfillment_id") != "ful_123" {
			t.Errorf("Expected fulfillment_id=ful_123, got %s", query.Get("fulfillment_id"))
		}
		if query.Get("status") != "in_transit" {
			t.Errorf("Expected status=in_transit, got %s", query.Get("status"))
		}
		if query.Get("tracking_number") != "ABC123" {
			t.Errorf("Expected tracking_number=ABC123, got %s", query.Get("tracking_number"))
		}

		resp := ShipmentsListResponse{
			Items:      []Shipment{{ID: "ship_123"}},
			Page:       2,
			PageSize:   10,
			TotalCount: 1,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &ShipmentsListOptions{
		Page:           2,
		PageSize:       10,
		OrderID:        "ord_123",
		FulfillmentID:  "ful_123",
		Status:         "in_transit",
		TrackingNumber: "ABC123",
	}
	shipments, err := client.ListShipments(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListShipments failed: %v", err)
	}

	if len(shipments.Items) != 1 {
		t.Errorf("Expected 1 shipment, got %d", len(shipments.Items))
	}
}
