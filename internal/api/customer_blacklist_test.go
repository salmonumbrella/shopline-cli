package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCustomerBlacklistList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/customer_blacklist" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := CustomerBlacklistListResponse{
			Items: []CustomerBlacklist{
				{ID: "bl_123", CustomerID: "cust_123", Email: "bad@example.com", Reason: "Fraud"},
				{ID: "bl_456", CustomerID: "cust_456", Email: "spam@example.com", Reason: "Spam"},
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

	entries, err := client.ListCustomerBlacklist(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListCustomerBlacklist failed: %v", err)
	}

	if len(entries.Items) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries.Items))
	}
	if entries.Items[0].ID != "bl_123" {
		t.Errorf("Unexpected entry ID: %s", entries.Items[0].ID)
	}
}

func TestCustomerBlacklistListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("email") != "bad@example.com" {
			t.Errorf("Expected email=bad@example.com, got %s", r.URL.Query().Get("email"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}

		resp := CustomerBlacklistListResponse{
			Items: []CustomerBlacklist{
				{ID: "bl_123", Email: "bad@example.com"},
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

	opts := &CustomerBlacklistListOptions{
		Page:  2,
		Email: "bad@example.com",
	}
	entries, err := client.ListCustomerBlacklist(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListCustomerBlacklist failed: %v", err)
	}

	if len(entries.Items) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries.Items))
	}
}

func TestCustomerBlacklistGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customer_blacklist/bl_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		entry := CustomerBlacklist{
			ID:         "bl_123",
			CustomerID: "cust_123",
			Email:      "bad@example.com",
			Reason:     "Fraud",
		}
		_ = json.NewEncoder(w).Encode(entry)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	entry, err := client.GetCustomerBlacklist(context.Background(), "bl_123")
	if err != nil {
		t.Fatalf("GetCustomerBlacklist failed: %v", err)
	}

	if entry.ID != "bl_123" {
		t.Errorf("Unexpected entry ID: %s", entry.ID)
	}
	if entry.Reason != "Fraud" {
		t.Errorf("Unexpected reason: %s", entry.Reason)
	}
}

func TestCustomerBlacklistCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/customer_blacklist" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req CustomerBlacklistCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Email != "bad@example.com" {
			t.Errorf("Unexpected email: %s", req.Email)
		}
		if req.Reason != "Fraud" {
			t.Errorf("Unexpected reason: %s", req.Reason)
		}

		entry := CustomerBlacklist{
			ID:     "bl_new",
			Email:  req.Email,
			Reason: req.Reason,
		}
		_ = json.NewEncoder(w).Encode(entry)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CustomerBlacklistCreateRequest{
		Email:  "bad@example.com",
		Reason: "Fraud",
	}
	entry, err := client.CreateCustomerBlacklist(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateCustomerBlacklist failed: %v", err)
	}

	if entry.ID != "bl_new" {
		t.Errorf("Unexpected entry ID: %s", entry.ID)
	}
}

func TestCustomerBlacklistDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/customer_blacklist/bl_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteCustomerBlacklist(context.Background(), "bl_123")
	if err != nil {
		t.Fatalf("DeleteCustomerBlacklist failed: %v", err)
	}
}

func TestGetCustomerBlacklistEmptyID(t *testing.T) {
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
			_, err := client.GetCustomerBlacklist(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "blacklist entry id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteCustomerBlacklistEmptyID(t *testing.T) {
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
			err := client.DeleteCustomerBlacklist(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "blacklist entry id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
