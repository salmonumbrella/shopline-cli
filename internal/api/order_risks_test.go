package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOrderRisksList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/orders/ord_123/risks" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := OrderRisksListResponse{
			Items: []OrderRisk{
				{ID: "risk_1", OrderID: "ord_123", Score: 0.8, Recommendation: "cancel"},
				{ID: "risk_2", OrderID: "ord_123", Score: 0.3, Recommendation: "accept"},
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

	risks, err := client.ListOrderRisks(context.Background(), "ord_123", nil)
	if err != nil {
		t.Fatalf("ListOrderRisks failed: %v", err)
	}

	if len(risks.Items) != 2 {
		t.Errorf("Expected 2 risks, got %d", len(risks.Items))
	}
	if risks.Items[0].ID != "risk_1" {
		t.Errorf("Unexpected risk ID: %s", risks.Items[0].ID)
	}
}

func TestOrderRisksGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orders/ord_123/risks/risk_1" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		risk := OrderRisk{ID: "risk_1", OrderID: "ord_123", Score: 0.8, Recommendation: "cancel"}
		_ = json.NewEncoder(w).Encode(risk)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	risk, err := client.GetOrderRisk(context.Background(), "ord_123", "risk_1")
	if err != nil {
		t.Fatalf("GetOrderRisk failed: %v", err)
	}

	if risk.ID != "risk_1" {
		t.Errorf("Unexpected risk ID: %s", risk.ID)
	}
}

func TestListOrderRisksEmptyOrderID(t *testing.T) {
	client := NewClient("token")

	_, err := client.ListOrderRisks(context.Background(), "", nil)
	if err == nil {
		t.Error("Expected error for empty order ID, got nil")
	}
	if err != nil && err.Error() != "order id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestGetOrderRiskEmptyIDs(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name      string
		orderID   string
		riskID    string
		wantError string
	}{
		{"empty order ID", "", "risk_1", "order id is required"},
		{"empty risk ID", "ord_123", "", "risk id is required"},
		{"whitespace order ID", "   ", "risk_1", "order id is required"},
		{"whitespace risk ID", "ord_123", "   ", "risk id is required"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.GetOrderRisk(context.Background(), tc.orderID, tc.riskID)
			if err == nil {
				t.Error("Expected error, got nil")
			}
			if err != nil && err.Error() != tc.wantError {
				t.Errorf("Expected error '%s', got '%s'", tc.wantError, err.Error())
			}
		})
	}
}

func TestCreateOrderRisk(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/orders/ord_123/risks" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		risk := OrderRisk{ID: "risk_new", OrderID: "ord_123", Score: 0.5, Recommendation: "investigate"}
		_ = json.NewEncoder(w).Encode(risk)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &OrderRiskCreateRequest{
		Score:          0.5,
		Recommendation: "investigate",
		Message:        "Suspicious activity detected",
	}

	risk, err := client.CreateOrderRisk(context.Background(), "ord_123", req)
	if err != nil {
		t.Fatalf("CreateOrderRisk failed: %v", err)
	}

	if risk.ID != "risk_new" {
		t.Errorf("Unexpected risk ID: %s", risk.ID)
	}
}

func TestDeleteOrderRisk(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/orders/ord_123/risks/risk_1" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteOrderRisk(context.Background(), "ord_123", "risk_1")
	if err != nil {
		t.Fatalf("DeleteOrderRisk failed: %v", err)
	}
}

func TestDeleteOrderRiskEmptyIDs(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name      string
		orderID   string
		riskID    string
		wantError string
	}{
		{"empty order ID", "", "risk_1", "order id is required"},
		{"empty risk ID", "ord_123", "", "risk id is required"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.DeleteOrderRisk(context.Background(), tc.orderID, tc.riskID)
			if err == nil {
				t.Error("Expected error, got nil")
			}
			if err != nil && err.Error() != tc.wantError {
				t.Errorf("Expected error '%s', got '%s'", tc.wantError, err.Error())
			}
		})
	}
}

func TestUpdateOrderRisk(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/orders/ord_123/risks/risk_1" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		risk := OrderRisk{ID: "risk_1", OrderID: "ord_123", Score: 0.9, Recommendation: "cancel"}
		_ = json.NewEncoder(w).Encode(risk)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	score := 0.9
	recommendation := "cancel"
	req := &OrderRiskUpdateRequest{
		Score:          &score,
		Recommendation: &recommendation,
	}

	risk, err := client.UpdateOrderRisk(context.Background(), "ord_123", "risk_1", req)
	if err != nil {
		t.Fatalf("UpdateOrderRisk failed: %v", err)
	}

	if risk.ID != "risk_1" {
		t.Errorf("Unexpected risk ID: %s", risk.ID)
	}
	if risk.Score != 0.9 {
		t.Errorf("Unexpected risk score: %f", risk.Score)
	}
}

func TestUpdateOrderRiskEmptyIDs(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name      string
		orderID   string
		riskID    string
		wantError string
	}{
		{"empty order ID", "", "risk_1", "order id is required"},
		{"empty risk ID", "ord_123", "", "risk id is required"},
		{"whitespace order ID", "   ", "risk_1", "order id is required"},
		{"whitespace risk ID", "ord_123", "   ", "risk id is required"},
	}

	score := 0.5
	req := &OrderRiskUpdateRequest{Score: &score}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.UpdateOrderRisk(context.Background(), tc.orderID, tc.riskID, req)
			if err == nil {
				t.Error("Expected error, got nil")
			}
			if err != nil && err.Error() != tc.wantError {
				t.Errorf("Expected error '%s', got '%s'", tc.wantError, err.Error())
			}
		})
	}
}

func TestCreateOrderRiskEmptyOrderID(t *testing.T) {
	client := NewClient("token")

	req := &OrderRiskCreateRequest{
		Score:          0.5,
		Recommendation: "investigate",
	}

	_, err := client.CreateOrderRisk(context.Background(), "", req)
	if err == nil {
		t.Error("Expected error for empty order ID, got nil")
	}
	if err != nil && err.Error() != "order id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestOrderRisksListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", query.Get("page"))
		}
		if query.Get("page_size") != "50" {
			t.Errorf("Expected page_size=50, got %s", query.Get("page_size"))
		}

		resp := OrderRisksListResponse{
			Items:      []OrderRisk{{ID: "risk_1", OrderID: "ord_123", Score: 0.5}},
			Page:       2,
			PageSize:   50,
			TotalCount: 100,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &OrderRisksListOptions{
		Page:     2,
		PageSize: 50,
	}
	risks, err := client.ListOrderRisks(context.Background(), "ord_123", opts)
	if err != nil {
		t.Fatalf("ListOrderRisks failed: %v", err)
	}

	if len(risks.Items) != 1 {
		t.Errorf("Expected 1 risk, got %d", len(risks.Items))
	}
}
