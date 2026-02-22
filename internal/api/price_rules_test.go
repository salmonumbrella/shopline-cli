package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPriceRulesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/price_rules" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := PriceRulesListResponse{
			Items: []PriceRule{
				{ID: "pr_123", Title: "10% off", ValueType: "percentage", Value: "-10"},
				{ID: "pr_456", Title: "$20 off", ValueType: "fixed_amount", Value: "-20"},
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

	rules, err := client.ListPriceRules(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListPriceRules failed: %v", err)
	}

	if len(rules.Items) != 2 {
		t.Errorf("Expected 2 price rules, got %d", len(rules.Items))
	}
	if rules.Items[0].ID != "pr_123" {
		t.Errorf("Unexpected price rule ID: %s", rules.Items[0].ID)
	}
}

func TestPriceRulesGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/price_rules/pr_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		rule := PriceRule{ID: "pr_123", Title: "10% off", ValueType: "percentage"}
		_ = json.NewEncoder(w).Encode(rule)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	rule, err := client.GetPriceRule(context.Background(), "pr_123")
	if err != nil {
		t.Fatalf("GetPriceRule failed: %v", err)
	}

	if rule.ID != "pr_123" {
		t.Errorf("Unexpected price rule ID: %s", rule.ID)
	}
}

func TestGetPriceRuleEmptyID(t *testing.T) {
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
			_, err := client.GetPriceRule(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "price rule id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestPriceRulesCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		rule := PriceRule{ID: "pr_new", Title: "New Rule"}
		_ = json.NewEncoder(w).Encode(rule)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &PriceRuleCreateRequest{
		Title:            "New Rule",
		TargetType:       "line_item",
		TargetSelection:  "all",
		AllocationMethod: "across",
		ValueType:        "percentage",
		Value:            "-15",
	}

	rule, err := client.CreatePriceRule(context.Background(), req)
	if err != nil {
		t.Fatalf("CreatePriceRule failed: %v", err)
	}

	if rule.Title != "New Rule" {
		t.Errorf("Unexpected rule title: %s", rule.Title)
	}
}

func TestPriceRulesUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}

		rule := PriceRule{ID: "pr_123", Title: "Updated Rule"}
		_ = json.NewEncoder(w).Encode(rule)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &PriceRuleUpdateRequest{
		Title: "Updated Rule",
	}

	rule, err := client.UpdatePriceRule(context.Background(), "pr_123", req)
	if err != nil {
		t.Fatalf("UpdatePriceRule failed: %v", err)
	}

	if rule.Title != "Updated Rule" {
		t.Errorf("Unexpected rule title: %s", rule.Title)
	}
}

func TestPriceRulesDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/price_rules/pr_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeletePriceRule(context.Background(), "pr_123")
	if err != nil {
		t.Fatalf("DeletePriceRule failed: %v", err)
	}
}

func TestDeletePriceRuleEmptyID(t *testing.T) {
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
			err := client.DeletePriceRule(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "price rule id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestListPriceRulesWithOptions(t *testing.T) {
	testCases := []struct {
		name          string
		opts          *PriceRulesListOptions
		expectedQuery map[string]string
	}{
		{
			name:          "page only",
			opts:          &PriceRulesListOptions{Page: 2},
			expectedQuery: map[string]string{"page": "2"},
		},
		{
			name:          "page_size only",
			opts:          &PriceRulesListOptions{PageSize: 50},
			expectedQuery: map[string]string{"page_size": "50"},
		},
		{
			name:          "all options combined",
			opts:          &PriceRulesListOptions{Page: 3, PageSize: 25},
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

				resp := PriceRulesListResponse{
					Items:      []PriceRule{{ID: "pr_123", Title: "Test Rule"}},
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

			_, err := client.ListPriceRules(context.Background(), tc.opts)
			if err != nil {
				t.Fatalf("ListPriceRules failed: %v", err)
			}
		})
	}
}

func TestUpdatePriceRuleEmptyID(t *testing.T) {
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
			_, err := client.UpdatePriceRule(context.Background(), tc.id, &PriceRuleUpdateRequest{})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "price rule id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
