package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMembershipTiersList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/membership_tiers" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		// API returns array directly
		resp := []MembershipTier{
			{ID: "tier_123", Name: "Bronze", Level: 1, MinPoints: 0},
			{ID: "tier_456", Name: "Silver", Level: 2, MinPoints: 1000},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	tiers, err := client.ListMembershipTiers(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListMembershipTiers failed: %v", err)
	}

	if len(tiers.Items) != 2 {
		t.Errorf("Expected 2 tiers, got %d", len(tiers.Items))
	}
	if tiers.Items[0].ID != "tier_123" {
		t.Errorf("Unexpected tier ID: %s", tiers.Items[0].ID)
	}
}

func TestMembershipTiersGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/membership_tiers/tier_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		tier := MembershipTier{ID: "tier_123", Name: "Bronze", Level: 1}
		_ = json.NewEncoder(w).Encode(tier)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	tier, err := client.GetMembershipTier(context.Background(), "tier_123")
	if err != nil {
		t.Fatalf("GetMembershipTier failed: %v", err)
	}

	if tier.ID != "tier_123" {
		t.Errorf("Unexpected tier ID: %s", tier.ID)
	}
}

func TestGetMembershipTierEmptyID(t *testing.T) {
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
			_, err := client.GetMembershipTier(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "tier id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestMembershipTiersCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/membership_tiers" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		tier := MembershipTier{ID: "tier_new", Name: "Gold", Level: 3}
		_ = json.NewEncoder(w).Encode(tier)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &MembershipTierCreateRequest{Name: "Gold", Level: 3, MinPoints: 5000}
	tier, err := client.CreateMembershipTier(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateMembershipTier failed: %v", err)
	}

	if tier.ID != "tier_new" {
		t.Errorf("Unexpected tier ID: %s", tier.ID)
	}
}

func TestMembershipTiersDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/membership_tiers/tier_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteMembershipTier(context.Background(), "tier_123")
	if err != nil {
		t.Fatalf("DeleteMembershipTier failed: %v", err)
	}
}

func TestDeleteMembershipTierEmptyID(t *testing.T) {
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
			err := client.DeleteMembershipTier(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "tier id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestListMembershipTiersWithOptions(t *testing.T) {
	testCases := []struct {
		name          string
		opts          *MembershipTiersListOptions
		expectedQuery map[string]string
	}{
		{
			name:          "page only",
			opts:          &MembershipTiersListOptions{Page: 2},
			expectedQuery: map[string]string{"page": "2"},
		},
		{
			name:          "page_size only",
			opts:          &MembershipTiersListOptions{PageSize: 50},
			expectedQuery: map[string]string{"page_size": "50"},
		},
		{
			name:          "all options combined",
			opts:          &MembershipTiersListOptions{Page: 3, PageSize: 25},
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

				// API returns array directly
				resp := []MembershipTier{{ID: "tier_123", Name: "Test Tier"}}
				_ = json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := NewClient("token")
			client.BaseURL = server.URL
			client.SetUseOpenAPI(false)

			_, err := client.ListMembershipTiers(context.Background(), tc.opts)
			if err != nil {
				t.Fatalf("ListMembershipTiers failed: %v", err)
			}
		})
	}
}
