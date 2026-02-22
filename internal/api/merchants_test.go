package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMerchantGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/merchants" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		// API returns {"items": [...]}
		resp := MerchantsListResponse{
			Items: []Merchant{{
				ID:       "merch_123",
				Name:     "Test Store",
				Handle:   "test-store",
				Email:    "admin@teststore.com",
				Currency: "USD",
				Plan:     "premium",
			}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	merchant, err := client.GetMerchant(context.Background())
	if err != nil {
		t.Fatalf("GetMerchant failed: %v", err)
	}

	if merchant.ID != "merch_123" {
		t.Errorf("Unexpected merchant ID: %s", merchant.ID)
	}
	if merchant.Name != "Test Store" {
		t.Errorf("Unexpected merchant name: %s", merchant.Name)
	}
	if merchant.Currency != "USD" {
		t.Errorf("Unexpected currency: %s", merchant.Currency)
	}
}

func TestMerchantStaffList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/merchant/staff" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := MerchantStaffListResponse{
			Items: []MerchantStaff{
				{ID: "staff_123", Email: "admin@store.com", Role: "admin", AccountOwner: true},
				{ID: "staff_456", Email: "support@store.com", Role: "support", AccountOwner: false},
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

	staff, err := client.ListMerchantStaff(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListMerchantStaff failed: %v", err)
	}

	if len(staff.Items) != 2 {
		t.Errorf("Expected 2 staff members, got %d", len(staff.Items))
	}
	if staff.Items[0].ID != "staff_123" {
		t.Errorf("Unexpected staff ID: %s", staff.Items[0].ID)
	}
}

func TestMerchantStaffListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("role") != "admin" {
			t.Errorf("Expected role=admin, got %s", r.URL.Query().Get("role"))
		}
		if r.URL.Query().Get("active") != "true" {
			t.Errorf("Expected active=true, got %s", r.URL.Query().Get("active"))
		}

		resp := MerchantStaffListResponse{
			Items:      []MerchantStaff{{ID: "staff_123", Role: "admin", Active: true}},
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

	active := true
	opts := &MerchantStaffListOptions{
		Role:   "admin",
		Active: &active,
	}
	staff, err := client.ListMerchantStaff(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListMerchantStaff failed: %v", err)
	}

	if len(staff.Items) != 1 {
		t.Errorf("Expected 1 staff member, got %d", len(staff.Items))
	}
}

func TestMerchantStaffGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/merchant/staff/staff_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		staff := MerchantStaff{
			ID:           "staff_123",
			Email:        "admin@store.com",
			FirstName:    "John",
			LastName:     "Doe",
			Role:         "admin",
			AccountOwner: true,
			Active:       true,
		}
		_ = json.NewEncoder(w).Encode(staff)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	staff, err := client.GetMerchantStaff(context.Background(), "staff_123")
	if err != nil {
		t.Fatalf("GetMerchantStaff failed: %v", err)
	}

	if staff.ID != "staff_123" {
		t.Errorf("Unexpected staff ID: %s", staff.ID)
	}
	if staff.Email != "admin@store.com" {
		t.Errorf("Unexpected email: %s", staff.Email)
	}
}

func TestGetMerchantStaffEmptyID(t *testing.T) {
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
			_, err := client.GetMerchantStaff(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "staff id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
