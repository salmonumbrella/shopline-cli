package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDisputesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/disputes" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := DisputesListResponse{
			Items: []Dispute{
				{ID: "dp_123", OrderID: "ord_123", Amount: "100.00", Status: "needs_response", Reason: "fraudulent"},
				{ID: "dp_456", OrderID: "ord_456", Amount: "50.00", Status: "under_review", Reason: "product_not_received"},
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

	disputes, err := client.ListDisputes(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListDisputes failed: %v", err)
	}

	if len(disputes.Items) != 2 {
		t.Errorf("Expected 2 disputes, got %d", len(disputes.Items))
	}
	if disputes.Items[0].ID != "dp_123" {
		t.Errorf("Unexpected dispute ID: %s", disputes.Items[0].ID)
	}
}

func TestDisputesListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("status") != "needs_response" {
			t.Errorf("Expected status=needs_response, got %s", r.URL.Query().Get("status"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}

		resp := DisputesListResponse{
			Items:      []Dispute{},
			Page:       2,
			PageSize:   20,
			TotalCount: 0,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &DisputesListOptions{
		Page:   2,
		Status: "needs_response",
	}
	_, err := client.ListDisputes(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListDisputes failed: %v", err)
	}
}

func TestDisputesGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/disputes/dp_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		dispute := Dispute{ID: "dp_123", OrderID: "ord_123", Amount: "100.00", Status: "needs_response"}
		_ = json.NewEncoder(w).Encode(dispute)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	dispute, err := client.GetDispute(context.Background(), "dp_123")
	if err != nil {
		t.Fatalf("GetDispute failed: %v", err)
	}

	if dispute.ID != "dp_123" {
		t.Errorf("Unexpected dispute ID: %s", dispute.ID)
	}
}

func TestGetDisputeEmptyID(t *testing.T) {
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
			_, err := client.GetDispute(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "dispute id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestUpdateDisputeEvidence(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/disputes/dp_123/evidence" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		dispute := Dispute{ID: "dp_123", Status: "needs_response"}
		_ = json.NewEncoder(w).Encode(dispute)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &DisputeUpdateEvidenceRequest{
		CustomerName:       "John Doe",
		ProductDescription: "Test product",
	}
	dispute, err := client.UpdateDisputeEvidence(context.Background(), "dp_123", req)
	if err != nil {
		t.Fatalf("UpdateDisputeEvidence failed: %v", err)
	}

	if dispute.ID != "dp_123" {
		t.Errorf("Unexpected dispute ID: %s", dispute.ID)
	}
}

func TestSubmitDispute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/disputes/dp_123/submit" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		dispute := Dispute{ID: "dp_123", Status: "under_review"}
		_ = json.NewEncoder(w).Encode(dispute)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	dispute, err := client.SubmitDispute(context.Background(), "dp_123")
	if err != nil {
		t.Fatalf("SubmitDispute failed: %v", err)
	}

	if dispute.Status != "under_review" {
		t.Errorf("Unexpected status: %s", dispute.Status)
	}
}

func TestAcceptDispute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/disputes/dp_123/accept" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		dispute := Dispute{ID: "dp_123", Status: "lost"}
		_ = json.NewEncoder(w).Encode(dispute)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	dispute, err := client.AcceptDispute(context.Background(), "dp_123")
	if err != nil {
		t.Fatalf("AcceptDispute failed: %v", err)
	}

	if dispute.Status != "lost" {
		t.Errorf("Unexpected status: %s", dispute.Status)
	}
}
