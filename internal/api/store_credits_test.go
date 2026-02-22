package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListCustomerStoreCredits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/customers/cust_123/store_credits" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[{"value":100}]}`))
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	resp, err := client.ListCustomerStoreCredits(context.Background(), "cust_123", 0, 0)
	if err != nil {
		t.Fatalf("ListCustomerStoreCredits failed: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected non-nil response")
	}
}

func TestListCustomerStoreCreditsWithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("per_page") != "10" {
			t.Errorf("Expected per_page=10, got %s", r.URL.Query().Get("per_page"))
		}
		if r.URL.Path != "/customers/cust_123/store_credits" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.ListCustomerStoreCredits(context.Background(), "cust_123", 2, 10)
	if err != nil {
		t.Fatalf("ListCustomerStoreCredits failed: %v", err)
	}
}

func TestListCustomerStoreCreditsEmptyID(t *testing.T) {
	client := NewClient("token")
	_, err := client.ListCustomerStoreCredits(context.Background(), "", 0, 0)
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
}

func TestUpdateCustomerStoreCredits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/customers/cust_123/store_credits" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var body StoreCreditUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}
		if body.Value != 222 {
			t.Errorf("Expected value=222, got %d", body.Value)
		}
		if body.Remarks != "livestream credit" {
			t.Errorf("Expected remarks='livestream credit', got %s", body.Remarks)
		}
		if body.Type != "manual_credit" {
			t.Errorf("Expected type='manual_credit', got %s", body.Type)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"sc_new","value":222}`))
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &StoreCreditUpdateRequest{
		Value:   222,
		Remarks: "livestream credit",
		Type:    "manual_credit",
	}
	resp, err := client.UpdateCustomerStoreCredits(context.Background(), "cust_123", req)
	if err != nil {
		t.Fatalf("UpdateCustomerStoreCredits failed: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected non-nil response")
	}
}

func TestUpdateCustomerStoreCreditsEmptyID(t *testing.T) {
	client := NewClient("token")
	req := &StoreCreditUpdateRequest{Value: 100, Remarks: "test"}
	_, err := client.UpdateCustomerStoreCredits(context.Background(), "", req)
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
}

func TestUpdateCustomerStoreCreditsNilRequest(t *testing.T) {
	client := NewClient("token")
	_, err := client.UpdateCustomerStoreCredits(context.Background(), "cust_123", nil)
	if err == nil {
		t.Error("Expected error for nil request, got nil")
	}
	if !strings.Contains(err.Error(), "request body is required") {
		t.Errorf("Expected 'request body is required' error, got: %v", err)
	}
}
