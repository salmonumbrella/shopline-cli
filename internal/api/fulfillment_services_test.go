package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFulfillmentServicesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/fulfillment_services" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := FulfillmentServicesListResponse{
			Items: []FulfillmentService{
				{ID: "fs_123", Name: "Third Party Fulfillment", Handle: "third-party"},
				{ID: "fs_456", Name: "Warehouse Service", Handle: "warehouse"},
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

	services, err := client.ListFulfillmentServices(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListFulfillmentServices failed: %v", err)
	}

	if len(services.Items) != 2 {
		t.Errorf("Expected 2 fulfillment services, got %d", len(services.Items))
	}
	if services.Items[0].ID != "fs_123" {
		t.Errorf("Unexpected fulfillment service ID: %s", services.Items[0].ID)
	}
}

func TestFulfillmentServicesGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/fulfillment_services/fs_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		fs := FulfillmentService{ID: "fs_123", Name: "Third Party Fulfillment"}
		_ = json.NewEncoder(w).Encode(fs)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	fs, err := client.GetFulfillmentService(context.Background(), "fs_123")
	if err != nil {
		t.Fatalf("GetFulfillmentService failed: %v", err)
	}

	if fs.ID != "fs_123" {
		t.Errorf("Unexpected fulfillment service ID: %s", fs.ID)
	}
}

func TestGetFulfillmentServiceEmptyID(t *testing.T) {
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
			_, err := client.GetFulfillmentService(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "fulfillment service id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestFulfillmentServicesCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		fs := FulfillmentService{ID: "fs_new", Name: "New Service"}
		_ = json.NewEncoder(w).Encode(fs)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &FulfillmentServiceCreateRequest{
		Name:        "New Service",
		CallbackURL: "https://example.com/callback",
	}

	fs, err := client.CreateFulfillmentService(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateFulfillmentService failed: %v", err)
	}

	if fs.ID != "fs_new" {
		t.Errorf("Unexpected fulfillment service ID: %s", fs.ID)
	}
}

func TestFulfillmentServicesUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}

		fs := FulfillmentService{ID: "fs_123", Name: "Updated Service"}
		_ = json.NewEncoder(w).Encode(fs)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &FulfillmentServiceUpdateRequest{
		Name: "Updated Service",
	}

	fs, err := client.UpdateFulfillmentService(context.Background(), "fs_123", req)
	if err != nil {
		t.Fatalf("UpdateFulfillmentService failed: %v", err)
	}

	if fs.Name != "Updated Service" {
		t.Errorf("Unexpected service name: %s", fs.Name)
	}
}

func TestFulfillmentServicesDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/fulfillment_services/fs_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteFulfillmentService(context.Background(), "fs_123")
	if err != nil {
		t.Fatalf("DeleteFulfillmentService failed: %v", err)
	}
}

func TestListFulfillmentServicesWithOptions(t *testing.T) {
	testCases := []struct {
		name          string
		opts          *FulfillmentServicesListOptions
		expectedQuery map[string]string
	}{
		{
			name:          "page only",
			opts:          &FulfillmentServicesListOptions{Page: 2},
			expectedQuery: map[string]string{"page": "2"},
		},
		{
			name:          "page_size only",
			opts:          &FulfillmentServicesListOptions{PageSize: 50},
			expectedQuery: map[string]string{"page_size": "50"},
		},
		{
			name:          "all options combined",
			opts:          &FulfillmentServicesListOptions{Page: 3, PageSize: 25},
			expectedQuery: map[string]string{"page": "3", "page_size": "25"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET, got %s", r.Method)
				}

				query := r.URL.Query()
				for key, expectedValue := range tc.expectedQuery {
					if got := query.Get(key); got != expectedValue {
						t.Errorf("Expected %s=%s, got %s=%s", key, expectedValue, key, got)
					}
				}

				resp := FulfillmentServicesListResponse{
					Items:      []FulfillmentService{{ID: "fs_123", Name: "Test Service"}},
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

			_, err := client.ListFulfillmentServices(context.Background(), tc.opts)
			if err != nil {
				t.Fatalf("ListFulfillmentServices failed: %v", err)
			}
		})
	}
}

func TestUpdateFulfillmentServiceEmptyID(t *testing.T) {
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
			_, err := client.UpdateFulfillmentService(context.Background(), tc.id, &FulfillmentServiceUpdateRequest{})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "fulfillment service id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteFulfillmentServiceEmptyID(t *testing.T) {
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
			err := client.DeleteFulfillmentService(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "fulfillment service id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
