package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckoutSettingsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/checkout_settings" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		settings := CheckoutSettings{
			ID:                  "cs_123",
			RequirePhone:        true,
			EnableGuestCheckout: true,
			EnableTipping:       true,
			TippingOptions:      []float64{10, 15, 20},
			AbandonedCartDelay:  24,
		}
		_ = json.NewEncoder(w).Encode(settings)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	settings, err := client.GetCheckoutSettings(context.Background())
	if err != nil {
		t.Fatalf("GetCheckoutSettings failed: %v", err)
	}

	if settings.ID != "cs_123" {
		t.Errorf("Unexpected settings ID: %s", settings.ID)
	}
	if !settings.RequirePhone {
		t.Error("Expected RequirePhone to be true")
	}
	if !settings.EnableGuestCheckout {
		t.Error("Expected EnableGuestCheckout to be true")
	}
	if len(settings.TippingOptions) != 3 {
		t.Errorf("Expected 3 tipping options, got %d", len(settings.TippingOptions))
	}
}

func TestCheckoutSettingsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/checkout_settings" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req CheckoutSettingsUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.EnableGuestCheckout == nil || *req.EnableGuestCheckout != false {
			t.Error("Expected EnableGuestCheckout to be false")
		}
		if req.AbandonedCartDelay != 48 {
			t.Errorf("Expected AbandonedCartDelay to be 48, got %d", req.AbandonedCartDelay)
		}

		settings := CheckoutSettings{
			ID:                  "cs_123",
			EnableGuestCheckout: false,
			AbandonedCartDelay:  48,
		}
		_ = json.NewEncoder(w).Encode(settings)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	enableGuest := false
	req := &CheckoutSettingsUpdateRequest{
		EnableGuestCheckout: &enableGuest,
		AbandonedCartDelay:  48,
	}

	settings, err := client.UpdateCheckoutSettings(context.Background(), req)
	if err != nil {
		t.Fatalf("UpdateCheckoutSettings failed: %v", err)
	}

	if settings.EnableGuestCheckout {
		t.Error("Expected EnableGuestCheckout to be false")
	}
	if settings.AbandonedCartDelay != 48 {
		t.Errorf("Expected AbandonedCartDelay to be 48, got %d", settings.AbandonedCartDelay)
	}
}

func TestCheckoutSettingsUpdatePolicies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CheckoutSettingsUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.TermsOfServiceURL != "https://example.com/terms" {
			t.Errorf("Unexpected terms URL: %s", req.TermsOfServiceURL)
		}
		if req.PrivacyPolicyURL != "https://example.com/privacy" {
			t.Errorf("Unexpected privacy URL: %s", req.PrivacyPolicyURL)
		}

		settings := CheckoutSettings{
			ID:                "cs_123",
			TermsOfServiceURL: req.TermsOfServiceURL,
			PrivacyPolicyURL:  req.PrivacyPolicyURL,
		}
		_ = json.NewEncoder(w).Encode(settings)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &CheckoutSettingsUpdateRequest{
		TermsOfServiceURL: "https://example.com/terms",
		PrivacyPolicyURL:  "https://example.com/privacy",
	}

	settings, err := client.UpdateCheckoutSettings(context.Background(), req)
	if err != nil {
		t.Fatalf("UpdateCheckoutSettings failed: %v", err)
	}

	if settings.TermsOfServiceURL != "https://example.com/terms" {
		t.Errorf("Unexpected terms URL: %s", settings.TermsOfServiceURL)
	}
}
