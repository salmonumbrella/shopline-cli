package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSettingsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/settings" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		// The actual API returns user settings in this format
		resp := SettingsResponse{
			Users: UserSettings{
				PosApplyCredit:  true,
				MinimumAgeLimit: "18",
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	settings, err := client.GetSettings(context.Background())
	if err != nil {
		t.Fatalf("GetSettings failed: %v", err)
	}

	if settings.Users.MinimumAgeLimit != "18" {
		t.Errorf("Unexpected minimum age limit: %s", settings.Users.MinimumAgeLimit)
	}
	if !settings.Users.PosApplyCredit {
		t.Error("Expected PosApplyCredit to be true")
	}
}

func TestSettingsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/settings" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req UserSettingsUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Users.MinimumAgeLimit != "21" {
			t.Errorf("Unexpected minimum_age_limit: %s", req.Users.MinimumAgeLimit)
		}
		if req.Users.PosApplyCredit == nil || !*req.Users.PosApplyCredit {
			t.Error("Expected pos_apply_credit to be true")
		}

		resp := SettingsResponse{
			Users: UserSettings{
				PosApplyCredit:  true,
				MinimumAgeLimit: "21",
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	posApplyCredit := true
	req := &UserSettingsUpdateRequest{
		Users: UserSettingsUpdate{
			MinimumAgeLimit: "21",
			PosApplyCredit:  &posApplyCredit,
		},
	}

	settings, err := client.UpdateSettings(context.Background(), req)
	if err != nil {
		t.Fatalf("UpdateSettings failed: %v", err)
	}

	if settings.Users.MinimumAgeLimit != "21" {
		t.Errorf("Unexpected minimum_age_limit: %s", settings.Users.MinimumAgeLimit)
	}
	if !settings.Users.PosApplyCredit {
		t.Error("Expected PosApplyCredit to be true")
	}
}

func TestSettingsUpdateWithBooleanFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req UserSettingsUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Users.PosApplyCredit == nil {
			t.Error("Expected PosApplyCredit to be set")
		} else if *req.Users.PosApplyCredit != false {
			t.Errorf("Expected PosApplyCredit to be false, got %v", *req.Users.PosApplyCredit)
		}

		resp := SettingsResponse{
			Users: UserSettings{
				PosApplyCredit:  false,
				MinimumAgeLimit: "18",
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	posApplyCredit := false
	req := &UserSettingsUpdateRequest{
		Users: UserSettingsUpdate{
			PosApplyCredit: &posApplyCredit,
		},
	}

	settings, err := client.UpdateSettings(context.Background(), req)
	if err != nil {
		t.Fatalf("UpdateSettings failed: %v", err)
	}

	if settings.Users.PosApplyCredit {
		t.Error("Expected PosApplyCredit to be false")
	}
}
