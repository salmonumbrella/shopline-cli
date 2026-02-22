package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCustomerAddressesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/customers/cust_123/addresses" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := CustomerAddressesListResponse{
			Items: []CustomerAddress{
				{ID: "addr_123", CustomerID: "cust_123", City: "New York"},
				{ID: "addr_456", CustomerID: "cust_123", City: "Los Angeles"},
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

	addresses, err := client.ListCustomerAddresses(context.Background(), "cust_123", nil)
	if err != nil {
		t.Fatalf("ListCustomerAddresses failed: %v", err)
	}

	if len(addresses.Items) != 2 {
		t.Errorf("Expected 2 addresses, got %d", len(addresses.Items))
	}
	if addresses.Items[0].ID != "addr_123" {
		t.Errorf("Unexpected address ID: %s", addresses.Items[0].ID)
	}
}

func TestListCustomerAddressesEmptyCustomerID(t *testing.T) {
	client := NewClient("token")

	_, err := client.ListCustomerAddresses(context.Background(), "", nil)
	if err == nil {
		t.Error("Expected error for empty customer ID, got nil")
	}
	if err != nil && err.Error() != "customer id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestCustomerAddressesGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customers/cust_123/addresses/addr_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		address := CustomerAddress{ID: "addr_123", CustomerID: "cust_123", City: "New York"}
		_ = json.NewEncoder(w).Encode(address)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	address, err := client.GetCustomerAddress(context.Background(), "cust_123", "addr_123")
	if err != nil {
		t.Fatalf("GetCustomerAddress failed: %v", err)
	}

	if address.ID != "addr_123" {
		t.Errorf("Unexpected address ID: %s", address.ID)
	}
}

func TestGetCustomerAddressEmptyIDs(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name       string
		customerID string
		addressID  string
		errMsg     string
	}{
		{"empty customer id", "", "addr_123", "customer id is required"},
		{"empty address id", "cust_123", "", "address id is required"},
		{"both empty", "", "", "customer id is required"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.GetCustomerAddress(context.Background(), tc.customerID, tc.addressID)
			if err == nil {
				t.Error("Expected error, got nil")
			}
			if err != nil && err.Error() != tc.errMsg {
				t.Errorf("Expected error '%s', got '%s'", tc.errMsg, err.Error())
			}
		})
	}
}

func TestCustomerAddressesCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/customers/cust_123/addresses" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		address := CustomerAddress{ID: "addr_new", CustomerID: "cust_123", City: "Chicago"}
		_ = json.NewEncoder(w).Encode(address)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CustomerAddressCreateRequest{Address1: "123 Main St", City: "Chicago", Country: "US"}
	address, err := client.CreateCustomerAddress(context.Background(), "cust_123", req)
	if err != nil {
		t.Fatalf("CreateCustomerAddress failed: %v", err)
	}

	if address.ID != "addr_new" {
		t.Errorf("Unexpected address ID: %s", address.ID)
	}
}

func TestCustomerAddressesDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/customers/cust_123/addresses/addr_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteCustomerAddress(context.Background(), "cust_123", "addr_123")
	if err != nil {
		t.Fatalf("DeleteCustomerAddress failed: %v", err)
	}
}

func TestSetDefaultCustomerAddress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/customers/cust_123/addresses/addr_123/default" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		address := CustomerAddress{ID: "addr_123", CustomerID: "cust_123", Default: true}
		_ = json.NewEncoder(w).Encode(address)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	address, err := client.SetDefaultCustomerAddress(context.Background(), "cust_123", "addr_123")
	if err != nil {
		t.Fatalf("SetDefaultCustomerAddress failed: %v", err)
	}

	if !address.Default {
		t.Error("Expected address to be marked as default")
	}
}
