package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSellingPlansList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/selling_plans" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := SellingPlansListResponse{
			Items: []SellingPlan{
				{
					ID:                "sp_123",
					Name:              "Monthly Subscription",
					Description:       "Subscribe and save 10%",
					Frequency:         "monthly",
					FrequencyInterval: 1,
					DiscountType:      "percentage",
					DiscountValue:     "10",
					Status:            "active",
				},
				{
					ID:                "sp_456",
					Name:              "Weekly Delivery",
					Description:       "Fresh products every week",
					Frequency:         "weekly",
					FrequencyInterval: 1,
					DiscountType:      "fixed",
					DiscountValue:     "5.00",
					Status:            "active",
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

	plans, err := client.ListSellingPlans(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListSellingPlans failed: %v", err)
	}

	if len(plans.Items) != 2 {
		t.Errorf("Expected 2 selling plans, got %d", len(plans.Items))
	}
	if plans.Items[0].ID != "sp_123" {
		t.Errorf("Unexpected selling plan ID: %s", plans.Items[0].ID)
	}
	if plans.Items[0].Frequency != "monthly" {
		t.Errorf("Unexpected frequency: %s", plans.Items[0].Frequency)
	}
}

func TestSellingPlansListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("status") != "active" {
			t.Errorf("Expected status=active, got %s", r.URL.Query().Get("status"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}

		resp := SellingPlansListResponse{
			Items: []SellingPlan{
				{ID: "sp_123", Status: "active"},
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

	opts := &SellingPlansListOptions{
		Page:   2,
		Status: "active",
	}
	plans, err := client.ListSellingPlans(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListSellingPlans failed: %v", err)
	}

	if len(plans.Items) != 1 {
		t.Errorf("Expected 1 selling plan, got %d", len(plans.Items))
	}
}

func TestSellingPlansGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/selling_plans/sp_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		plan := SellingPlan{
			ID:                "sp_123",
			Name:              "Monthly Subscription",
			Description:       "Subscribe and save 10%",
			BillingPolicy:     "recurring",
			DeliveryPolicy:    "recurring",
			Frequency:         "monthly",
			FrequencyInterval: 1,
			TrialDays:         7,
			DiscountType:      "percentage",
			DiscountValue:     "10",
			Status:            "active",
			Position:          1,
		}
		_ = json.NewEncoder(w).Encode(plan)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	plan, err := client.GetSellingPlan(context.Background(), "sp_123")
	if err != nil {
		t.Fatalf("GetSellingPlan failed: %v", err)
	}

	if plan.ID != "sp_123" {
		t.Errorf("Unexpected selling plan ID: %s", plan.ID)
	}
	if plan.Name != "Monthly Subscription" {
		t.Errorf("Unexpected name: %s", plan.Name)
	}
	if plan.TrialDays != 7 {
		t.Errorf("Unexpected trial days: %d", plan.TrialDays)
	}
}

func TestSellingPlansCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/selling_plans" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req SellingPlanCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Name != "Quarterly Plan" {
			t.Errorf("Unexpected name: %s", req.Name)
		}
		if req.Frequency != "quarterly" {
			t.Errorf("Unexpected frequency: %s", req.Frequency)
		}

		plan := SellingPlan{
			ID:                "sp_new",
			Name:              req.Name,
			Description:       req.Description,
			Frequency:         req.Frequency,
			FrequencyInterval: req.FrequencyInterval,
			DiscountType:      req.DiscountType,
			DiscountValue:     req.DiscountValue,
			Status:            "active",
		}
		_ = json.NewEncoder(w).Encode(plan)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &SellingPlanCreateRequest{
		Name:          "Quarterly Plan",
		Description:   "Save 15% with quarterly billing",
		Frequency:     "quarterly",
		DiscountType:  "percentage",
		DiscountValue: "15",
	}
	plan, err := client.CreateSellingPlan(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateSellingPlan failed: %v", err)
	}

	if plan.ID != "sp_new" {
		t.Errorf("Unexpected selling plan ID: %s", plan.ID)
	}
	if plan.Name != "Quarterly Plan" {
		t.Errorf("Unexpected name: %s", plan.Name)
	}
}

func TestSellingPlansDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/selling_plans/sp_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteSellingPlan(context.Background(), "sp_123")
	if err != nil {
		t.Fatalf("DeleteSellingPlan failed: %v", err)
	}
}

func TestGetSellingPlanEmptyID(t *testing.T) {
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
			_, err := client.GetSellingPlan(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "selling plan id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteSellingPlanEmptyID(t *testing.T) {
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
			err := client.DeleteSellingPlan(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "selling plan id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
