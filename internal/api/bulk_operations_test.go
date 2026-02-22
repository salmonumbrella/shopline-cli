package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBulkOperationsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/bulk_operations" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := BulkOperationsListResponse{
			Items: []BulkOperation{
				{ID: "bo_123", Type: "query", Status: "completed", ObjectCount: 100},
				{ID: "bo_456", Type: "mutation", Status: "running", ObjectCount: 50},
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

	ops, err := client.ListBulkOperations(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListBulkOperations failed: %v", err)
	}

	if len(ops.Items) != 2 {
		t.Errorf("Expected 2 bulk operations, got %d", len(ops.Items))
	}
	if ops.Items[0].ID != "bo_123" {
		t.Errorf("Unexpected bulk operation ID: %s", ops.Items[0].ID)
	}
}

func TestBulkOperationsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bulk_operations/bo_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		op := BulkOperation{
			ID:          "bo_123",
			Type:        "query",
			Status:      "completed",
			ObjectCount: 100,
			FileSize:    1024,
			URL:         "https://example.com/results.jsonl",
		}
		_ = json.NewEncoder(w).Encode(op)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	op, err := client.GetBulkOperation(context.Background(), "bo_123")
	if err != nil {
		t.Fatalf("GetBulkOperation failed: %v", err)
	}

	if op.ID != "bo_123" {
		t.Errorf("Unexpected bulk operation ID: %s", op.ID)
	}
	if op.ObjectCount != 100 {
		t.Errorf("Unexpected object count: %d", op.ObjectCount)
	}
}

func TestGetBulkOperationEmptyID(t *testing.T) {
	client := NewClient("token")

	_, err := client.GetBulkOperation(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
	if err != nil && err.Error() != "bulk operation id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestCreateBulkQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/bulk_operations/queries" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		op := BulkOperation{ID: "bo_new", Type: "query", Status: "created"}
		_ = json.NewEncoder(w).Encode(op)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &BulkOperationCreateRequest{
		Query: "{ products { edges { node { id title } } } }",
	}

	op, err := client.CreateBulkQuery(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateBulkQuery failed: %v", err)
	}

	if op.ID != "bo_new" {
		t.Errorf("Unexpected bulk operation ID: %s", op.ID)
	}
}

func TestCreateBulkMutation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/bulk_operations/mutations" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		op := BulkOperation{ID: "bo_mutation", Type: "mutation", Status: "created"}
		_ = json.NewEncoder(w).Encode(op)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &BulkOperationMutationRequest{
		Mutation:         "mutation ($input: ProductInput!) { productUpdate(input: $input) { product { id } } }",
		StagedUploadPath: "tmp/bulk-mutation.jsonl",
	}

	op, err := client.CreateBulkMutation(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateBulkMutation failed: %v", err)
	}

	if op.ID != "bo_mutation" {
		t.Errorf("Unexpected bulk operation ID: %s", op.ID)
	}
}

func TestCancelBulkOperation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/bulk_operations/bo_123/cancel" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		op := BulkOperation{ID: "bo_123", Status: "cancelled"}
		_ = json.NewEncoder(w).Encode(op)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	op, err := client.CancelBulkOperation(context.Background(), "bo_123")
	if err != nil {
		t.Fatalf("CancelBulkOperation failed: %v", err)
	}

	if op.Status != "cancelled" {
		t.Errorf("Unexpected status: %s", op.Status)
	}
}

func TestGetCurrentBulkOperation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bulk_operations/current" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		op := BulkOperation{ID: "bo_running", Type: "query", Status: "running"}
		_ = json.NewEncoder(w).Encode(op)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	op, err := client.GetCurrentBulkOperation(context.Background())
	if err != nil {
		t.Fatalf("GetCurrentBulkOperation failed: %v", err)
	}

	if op.Status != "running" {
		t.Errorf("Unexpected status: %s", op.Status)
	}
}
