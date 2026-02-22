package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPaymentsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/payments" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := PaymentsListResponse{
			Items: []Payment{
				{ID: "pay_123", OrderID: "ord_123", Amount: "99.99", Status: "captured"},
				{ID: "pay_456", OrderID: "ord_456", Amount: "49.99", Status: "authorized"},
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

	payments, err := client.ListPayments(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListPayments failed: %v", err)
	}

	if len(payments.Items) != 2 {
		t.Errorf("Expected 2 payments, got %d", len(payments.Items))
	}
	if payments.Items[0].ID != "pay_123" {
		t.Errorf("Unexpected payment ID: %s", payments.Items[0].ID)
	}
}

func TestPaymentsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/payments/pay_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		payment := Payment{ID: "pay_123", OrderID: "ord_123", Amount: "99.99"}
		_ = json.NewEncoder(w).Encode(payment)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	payment, err := client.GetPayment(context.Background(), "pay_123")
	if err != nil {
		t.Fatalf("GetPayment failed: %v", err)
	}

	if payment.ID != "pay_123" {
		t.Errorf("Unexpected payment ID: %s", payment.ID)
	}
}

func TestGetPaymentEmptyID(t *testing.T) {
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
			_, err := client.GetPayment(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "payment id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestListOrderPayments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orders/ord_123/payments" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := PaymentsListResponse{
			Items: []Payment{
				{ID: "pay_123", OrderID: "ord_123", Amount: "99.99"},
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

	payments, err := client.ListOrderPayments(context.Background(), "ord_123")
	if err != nil {
		t.Fatalf("ListOrderPayments failed: %v", err)
	}

	if len(payments.Items) != 1 {
		t.Errorf("Expected 1 payment, got %d", len(payments.Items))
	}
}

func TestCapturePayment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/payments/pay_123/capture" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		payment := Payment{ID: "pay_123", Status: "captured"}
		_ = json.NewEncoder(w).Encode(payment)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	payment, err := client.CapturePayment(context.Background(), "pay_123", "99.99")
	if err != nil {
		t.Fatalf("CapturePayment failed: %v", err)
	}

	if payment.Status != "captured" {
		t.Errorf("Unexpected status: %s", payment.Status)
	}
}

func TestVoidPayment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/payments/pay_123/void" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		payment := Payment{ID: "pay_123", Status: "voided"}
		_ = json.NewEncoder(w).Encode(payment)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	payment, err := client.VoidPayment(context.Background(), "pay_123")
	if err != nil {
		t.Fatalf("VoidPayment failed: %v", err)
	}

	if payment.Status != "voided" {
		t.Errorf("Unexpected status: %s", payment.Status)
	}
}

func TestRefundPayment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/payments/pay_123/refund" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		payment := Payment{ID: "pay_123", Status: "refunded"}
		_ = json.NewEncoder(w).Encode(payment)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	payment, err := client.RefundPayment(context.Background(), "pay_123", "50.00", "customer request")
	if err != nil {
		t.Fatalf("RefundPayment failed: %v", err)
	}

	if payment.Status != "refunded" {
		t.Errorf("Unexpected status: %s", payment.Status)
	}
}

func TestListOrderPaymentsEmptyID(t *testing.T) {
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
			_, err := client.ListOrderPayments(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCapturePaymentEmptyID(t *testing.T) {
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
			_, err := client.CapturePayment(context.Background(), tc.id, "99.99")
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "payment id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestVoidPaymentEmptyID(t *testing.T) {
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
			_, err := client.VoidPayment(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "payment id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestRefundPaymentEmptyID(t *testing.T) {
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
			_, err := client.RefundPayment(context.Background(), tc.id, "50.00", "reason")
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "payment id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestPaymentsListWithOptions(t *testing.T) {
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
		if query.Get("status") != "captured" {
			t.Errorf("Expected status=captured, got %s", query.Get("status"))
		}
		if query.Get("gateway") != "stripe" {
			t.Errorf("Expected gateway=stripe, got %s", query.Get("gateway"))
		}

		resp := PaymentsListResponse{
			Items:      []Payment{{ID: "pay_123", Status: "captured", Gateway: "stripe"}},
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

	opts := &PaymentsListOptions{
		Page:     2,
		PageSize: 10,
		Status:   "captured",
		Gateway:  "stripe",
	}
	payments, err := client.ListPayments(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListPayments failed: %v", err)
	}

	if len(payments.Items) != 1 {
		t.Errorf("Expected 1 payment, got %d", len(payments.Items))
	}
}

func TestCapturePaymentWithEmptyAmount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/payments/pay_123/capture" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		payment := Payment{ID: "pay_123", Status: "captured"}
		_ = json.NewEncoder(w).Encode(payment)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	payment, err := client.CapturePayment(context.Background(), "pay_123", "")
	if err != nil {
		t.Fatalf("CapturePayment failed: %v", err)
	}

	if payment.Status != "captured" {
		t.Errorf("Unexpected status: %s", payment.Status)
	}
}

func TestRefundPaymentWithEmptyAmountAndReason(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/payments/pay_123/refund" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		payment := Payment{ID: "pay_123", Status: "refunded"}
		_ = json.NewEncoder(w).Encode(payment)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	payment, err := client.RefundPayment(context.Background(), "pay_123", "", "")
	if err != nil {
		t.Fatalf("RefundPayment failed: %v", err)
	}

	if payment.Status != "refunded" {
		t.Errorf("Unexpected status: %s", payment.Status)
	}
}
