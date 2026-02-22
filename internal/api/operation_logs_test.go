package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOperationLogsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/operation_logs" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := OperationLogsListResponse{
			Items: []OperationLog{
				{
					ID:           "log_123",
					Action:       OperationLogActionCreate,
					ResourceType: "product",
					ResourceID:   "prod_123",
					UserEmail:    "admin@example.com",
				},
				{
					ID:           "log_456",
					Action:       OperationLogActionUpdate,
					ResourceType: "order",
					ResourceID:   "ord_456",
					UserEmail:    "staff@example.com",
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

	logs, err := client.ListOperationLogs(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListOperationLogs failed: %v", err)
	}

	if len(logs.Items) != 2 {
		t.Errorf("Expected 2 operation logs, got %d", len(logs.Items))
	}
	if logs.Items[0].ID != "log_123" {
		t.Errorf("Unexpected log ID: %s", logs.Items[0].ID)
	}
	if logs.Items[0].Action != OperationLogActionCreate {
		t.Errorf("Unexpected action: %s", logs.Items[0].Action)
	}
}

func TestOperationLogsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("action") != "create" {
			t.Errorf("Expected action=create, got %s", r.URL.Query().Get("action"))
		}
		if r.URL.Query().Get("resource_type") != "product" {
			t.Errorf("Expected resource_type=product, got %s", r.URL.Query().Get("resource_type"))
		}

		resp := OperationLogsListResponse{
			Items:      []OperationLog{{ID: "log_123", Action: OperationLogActionCreate}},
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

	opts := &OperationLogsListOptions{
		Action:       OperationLogActionCreate,
		ResourceType: "product",
	}
	logs, err := client.ListOperationLogs(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListOperationLogs failed: %v", err)
	}

	if len(logs.Items) != 1 {
		t.Errorf("Expected 1 operation log, got %d", len(logs.Items))
	}
}

func TestOperationLogsListWithDateRange(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startDate := r.URL.Query().Get("start_date")
		if startDate == "" {
			t.Error("Expected start_date to be set")
		}

		resp := OperationLogsListResponse{
			Items:      []OperationLog{{ID: "log_123"}},
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

	startDate := time.Now().AddDate(0, 0, -7)
	opts := &OperationLogsListOptions{
		StartDate: &startDate,
	}
	logs, err := client.ListOperationLogs(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListOperationLogs failed: %v", err)
	}

	if len(logs.Items) != 1 {
		t.Errorf("Expected 1 operation log, got %d", len(logs.Items))
	}
}

func TestOperationLogsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/operation_logs/log_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		log := OperationLog{
			ID:           "log_123",
			Action:       OperationLogActionUpdate,
			ResourceType: "product",
			ResourceID:   "prod_123",
			ResourceName: "Test Product",
			UserID:       "user_123",
			UserEmail:    "admin@example.com",
			UserName:     "Admin User",
			IPAddress:    "192.168.1.1",
			UserAgent:    "Mozilla/5.0",
			Changes: map[string]Change{
				"title": {From: "Old Title", To: "New Title"},
				"price": {From: 10.00, To: 15.00},
			},
			Metadata: map[string]string{
				"source": "admin_panel",
			},
		}
		_ = json.NewEncoder(w).Encode(log)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	log, err := client.GetOperationLog(context.Background(), "log_123")
	if err != nil {
		t.Fatalf("GetOperationLog failed: %v", err)
	}

	if log.ID != "log_123" {
		t.Errorf("Unexpected log ID: %s", log.ID)
	}
	if log.ResourceName != "Test Product" {
		t.Errorf("Unexpected resource name: %s", log.ResourceName)
	}
	if len(log.Changes) != 2 {
		t.Errorf("Expected 2 changes, got %d", len(log.Changes))
	}
	if log.Changes["title"].From != "Old Title" {
		t.Errorf("Unexpected title from value: %v", log.Changes["title"].From)
	}
}

func TestGetOperationLogEmptyID(t *testing.T) {
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
			_, err := client.GetOperationLog(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "operation log id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
