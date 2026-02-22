package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAffiliateCampaignOrders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/affiliate_campaigns/camp_123/orders" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("page"); got != "2" {
			t.Fatalf("expected page=2, got %q", got)
		}
		if got := r.URL.Query().Get("page_size"); got != "50" {
			t.Fatalf("expected page_size=50, got %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.GetAffiliateCampaignOrders(context.Background(), "camp_123", &AffiliateCampaignOrdersOptions{
		Page:     2,
		PageSize: 50,
	})
	if err != nil {
		t.Fatalf("GetAffiliateCampaignOrders failed: %v", err)
	}
	if len(raw) == 0 {
		t.Fatalf("expected response body")
	}
}

func TestGetAffiliateCampaignSummary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/affiliate_campaigns/camp_123/summary" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.GetAffiliateCampaignSummary(context.Background(), "camp_123")
	if err != nil {
		t.Fatalf("GetAffiliateCampaignSummary failed: %v", err)
	}
	if len(raw) == 0 {
		t.Fatalf("expected response body")
	}
}

func TestGetAffiliateCampaignProductsSalesRanking(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/affiliate_campaigns/camp_123/get_products_sales_ranking" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("page"); got != "1" {
			t.Fatalf("expected page=1, got %q", got)
		}
		if got := r.URL.Query().Get("page_size"); got != "20" {
			t.Fatalf("expected page_size=20, got %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.GetAffiliateCampaignProductsSalesRanking(context.Background(), "camp_123", &AffiliateCampaignProductsSalesRankingOptions{
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("GetAffiliateCampaignProductsSalesRanking failed: %v", err)
	}
	if len(raw) == 0 {
		t.Fatalf("expected response body")
	}
}

func TestExportAffiliateCampaignReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/affiliate_campaigns/camp_123/export_report" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body["ok"] != true {
			t.Fatalf("expected ok=true in body, got %v", body)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{"job_id": "job_1"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	raw, err := client.ExportAffiliateCampaignReport(context.Background(), "camp_123", map[string]any{"ok": true})
	if err != nil {
		t.Fatalf("ExportAffiliateCampaignReport failed: %v", err)
	}
	if len(raw) == 0 {
		t.Fatalf("expected response body")
	}
}

func TestAffiliateCampaignDocsEmptyID(t *testing.T) {
	client := NewClient("token")
	client.SetUseOpenAPI(false)

	if _, err := client.GetAffiliateCampaignOrders(context.Background(), " ", nil); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := client.GetAffiliateCampaignSummary(context.Background(), " "); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := client.GetAffiliateCampaignProductsSalesRanking(context.Background(), " ", nil); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := client.ExportAffiliateCampaignReport(context.Background(), " ", map[string]any{}); err == nil {
		t.Fatalf("expected error")
	}
}
