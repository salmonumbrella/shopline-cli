package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAffiliateCampaignsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/affiliate_campaigns" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := AffiliateCampaignsListResponse{
			Items: []AffiliateCampaign{
				{ID: "aff_123", Name: "Summer Sale", Status: "active", CommissionType: "percentage", CommissionValue: 10.0},
				{ID: "aff_456", Name: "Holiday Promo", Status: "paused", CommissionType: "fixed", CommissionValue: 5.0},
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

	campaigns, err := client.ListAffiliateCampaigns(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListAffiliateCampaigns failed: %v", err)
	}

	if len(campaigns.Items) != 2 {
		t.Errorf("Expected 2 campaigns, got %d", len(campaigns.Items))
	}
	if campaigns.Items[0].ID != "aff_123" {
		t.Errorf("Unexpected campaign ID: %s", campaigns.Items[0].ID)
	}
}

func TestAffiliateCampaignsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("status") != "active" {
			t.Errorf("Expected status=active, got %s", r.URL.Query().Get("status"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}

		resp := AffiliateCampaignsListResponse{
			Items:      []AffiliateCampaign{},
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

	opts := &AffiliateCampaignsListOptions{
		Page:   2,
		Status: "active",
	}
	_, err := client.ListAffiliateCampaigns(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListAffiliateCampaigns failed: %v", err)
	}
}

func TestAffiliateCampaignsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/affiliate_campaigns/aff_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		campaign := AffiliateCampaign{ID: "aff_123", Name: "Summer Sale", Status: "active", CommissionType: "percentage", CommissionValue: 10.0}
		_ = json.NewEncoder(w).Encode(campaign)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	campaign, err := client.GetAffiliateCampaign(context.Background(), "aff_123")
	if err != nil {
		t.Fatalf("GetAffiliateCampaign failed: %v", err)
	}

	if campaign.ID != "aff_123" {
		t.Errorf("Unexpected campaign ID: %s", campaign.ID)
	}
}

func TestGetAffiliateCampaignEmptyID(t *testing.T) {
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
			_, err := client.GetAffiliateCampaign(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "affiliate campaign id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestAffiliateCampaignsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/affiliate_campaigns" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req AffiliateCampaignCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Name != "New Campaign" {
			t.Errorf("Unexpected name: %s", req.Name)
		}

		campaign := AffiliateCampaign{
			ID:              "aff_new",
			Name:            req.Name,
			CommissionType:  req.CommissionType,
			CommissionValue: req.CommissionValue,
			Status:          "active",
		}
		_ = json.NewEncoder(w).Encode(campaign)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &AffiliateCampaignCreateRequest{
		Name:            "New Campaign",
		CommissionType:  "percentage",
		CommissionValue: 15.0,
	}
	campaign, err := client.CreateAffiliateCampaign(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateAffiliateCampaign failed: %v", err)
	}

	if campaign.ID != "aff_new" {
		t.Errorf("Unexpected campaign ID: %s", campaign.ID)
	}
	if campaign.Name != "New Campaign" {
		t.Errorf("Unexpected name: %s", campaign.Name)
	}
}

func TestAffiliateCampaignsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/affiliate_campaigns/aff_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req AffiliateCampaignUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		campaign := AffiliateCampaign{
			ID:              "aff_123",
			Name:            req.Name,
			Status:          "paused",
			CommissionType:  "percentage",
			CommissionValue: 10.0,
		}
		_ = json.NewEncoder(w).Encode(campaign)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &AffiliateCampaignUpdateRequest{
		Name:   "Updated Campaign",
		Status: "paused",
	}
	campaign, err := client.UpdateAffiliateCampaign(context.Background(), "aff_123", req)
	if err != nil {
		t.Fatalf("UpdateAffiliateCampaign failed: %v", err)
	}

	if campaign.Name != "Updated Campaign" {
		t.Errorf("Unexpected campaign name: %s", campaign.Name)
	}
}

func TestAffiliateCampaignsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/affiliate_campaigns/aff_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteAffiliateCampaign(context.Background(), "aff_123")
	if err != nil {
		t.Fatalf("DeleteAffiliateCampaign failed: %v", err)
	}
}

func TestUpdateAffiliateCampaignEmptyID(t *testing.T) {
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
			_, err := client.UpdateAffiliateCampaign(context.Background(), tc.id, &AffiliateCampaignUpdateRequest{Name: "Test"})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "affiliate campaign id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteAffiliateCampaignEmptyID(t *testing.T) {
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
			err := client.DeleteAffiliateCampaign(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "affiliate campaign id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
