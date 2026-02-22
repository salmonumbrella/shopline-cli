package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

func TestCheckoutSettingsCommandStructure(t *testing.T) {
	if checkoutSettingsCmd == nil {
		t.Fatal("checkoutSettingsCmd is nil")
	}
	if checkoutSettingsCmd.Use != "checkout-settings" {
		t.Errorf("Expected Use 'checkout-settings', got %q", checkoutSettingsCmd.Use)
	}
	hasCheckoutAlias := false
	for _, alias := range checkoutSettingsCmd.Aliases {
		if alias == "checkout" {
			hasCheckoutAlias = true
			break
		}
	}
	if !hasCheckoutAlias {
		t.Error("Expected 'checkout' alias")
	}
	subcommands := map[string]bool{"get": false, "update": false}
	for _, cmd := range checkoutSettingsCmd.Commands() {
		for key := range subcommands {
			if strings.HasPrefix(cmd.Use, key) {
				subcommands[key] = true
			}
		}
	}
	for name, found := range subcommands {
		if !found {
			t.Errorf("Subcommand %q not found", name)
		}
	}
}

func TestCheckoutSettingsUpdateFlags(t *testing.T) {
	cmd := checkoutSettingsUpdateCmd
	flags := []struct{ name, defaultValue string }{
		{"require-phone", "false"},
		{"guest-checkout", "false"},
		{"express-checkout", "false"},
		{"order-notes", "false"},
		{"tipping", "false"},
		{"abandoned-cart", "false"},
		{"abandoned-cart-delay", "24"},
		{"terms-url", ""},
		{"privacy-url", ""},
		{"refund-url", ""},
	}
	for _, f := range flags {
		flag := cmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("Flag %q not found", f.name)
		} else if flag.DefValue != f.defaultValue {
			t.Errorf("Flag %q default: expected %q, got %q", f.name, f.defaultValue, flag.DefValue)
		}
	}
}

func TestCheckoutSettingsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return nil, errors.New("keyring error") }
	cmd := newTestCmdWithFlags()
	if err := checkoutSettingsGetCmd.RunE(cmd, []string{}); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCheckoutSettingsGetNoProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return &mockStore{names: []string{}}, nil }
	cmd := newTestCmdWithFlags()
	err := checkoutSettingsGetCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no store profiles configured") {
		t.Errorf("Expected 'no store profiles' error, got: %v", err)
	}
}

func TestCheckoutSettingsGetWithEnvVar(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() { secretsStoreFactory = origFactory; _ = os.Setenv("SHOPLINE_STORE", origEnv) }()
	_ = os.Setenv("SHOPLINE_STORE", "envstore")
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{names: []string{"envstore", "other"}, creds: map[string]*secrets.StoreCredentials{"envstore": {Handle: "test", AccessToken: "token123"}}}, nil
	}
	cmd := newTestCmdWithFlags()
	err := checkoutSettingsGetCmd.RunE(cmd, []string{})
	if err != nil && strings.Contains(err.Error(), "multiple profiles") {
		t.Errorf("Should have selected store from env var, got: %v", err)
	}
}

// checkoutSettingsMockAPIClient is a mock implementation of api.APIClient for checkout settings tests.
type checkoutSettingsMockAPIClient struct {
	api.MockClient
	getCheckoutSettingsResp    *api.CheckoutSettings
	getCheckoutSettingsErr     error
	updateCheckoutSettingsResp *api.CheckoutSettings
	updateCheckoutSettingsErr  error
}

func (m *checkoutSettingsMockAPIClient) GetCheckoutSettings(ctx context.Context) (*api.CheckoutSettings, error) {
	return m.getCheckoutSettingsResp, m.getCheckoutSettingsErr
}

func (m *checkoutSettingsMockAPIClient) UpdateCheckoutSettings(ctx context.Context, req *api.CheckoutSettingsUpdateRequest) (*api.CheckoutSettings, error) {
	return m.updateCheckoutSettingsResp, m.updateCheckoutSettingsErr
}

// setupCheckoutSettingsMockFactories sets up mock factories for checkout settings tests.
func setupCheckoutSettingsMockFactories(mockClient *checkoutSettingsMockAPIClient) (func(), *bytes.Buffer) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter

	buf := new(bytes.Buffer)
	formatterWriter = buf

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	cleanup := func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}

	return cleanup, buf
}

// newCheckoutSettingsTestCmd creates a test command with common flags for checkout settings tests.
func newCheckoutSettingsTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("yes", true, "")
	return cmd
}

// TestCheckoutSettingsGetRunE tests the checkout settings get command with mock API.
func TestCheckoutSettingsGetRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.CheckoutSettings
		mockErr    error
		jsonOutput bool
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful get with all fields",
			mockResp: &api.CheckoutSettings{
				ID:                     "cs_123",
				RequirePhone:           true,
				RequireShippingAddress: true,
				RequireBillingAddress:  false,
				RequireCompany:         true,
				RequireFullName:        true,
				EnableGuestCheckout:    true,
				EnableExpressCheckout:  true,
				EnableOrderNotes:       true,
				EnableTipping:          true,
				TippingOptions:         []float64{10, 15, 20},
				DefaultTippingOption:   15,
				EnableAddressAutofill:  true,
				EnableMultiCurrency:    true,
				AbandonedCartEnabled:   true,
				AbandonedCartDelay:     24,
				TermsOfServiceURL:      "https://example.com/terms",
				PrivacyPolicyURL:       "https://example.com/privacy",
				RefundPolicyURL:        "https://example.com/refund",
				UpdatedAt:              time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			// Text output goes to stdout, not to buffer
		},
		{
			name: "successful get with minimal fields",
			mockResp: &api.CheckoutSettings{
				ID:                    "cs_456",
				RequirePhone:          false,
				EnableGuestCheckout:   false,
				EnableExpressCheckout: false,
				EnableOrderNotes:      false,
				EnableTipping:         false,
				AbandonedCartEnabled:  false,
				UpdatedAt:             time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name: "successful get with JSON output",
			mockResp: &api.CheckoutSettings{
				ID:                    "cs_789",
				RequirePhone:          true,
				EnableGuestCheckout:   true,
				EnableExpressCheckout: true,
				EnableTipping:         false,
				UpdatedAt:             time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			jsonOutput: true,
			wantOutput: "cs_789",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name:    "network error",
			mockErr: errors.New("connection refused"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &checkoutSettingsMockAPIClient{
				getCheckoutSettingsResp: tt.mockResp,
				getCheckoutSettingsErr:  tt.mockErr,
			}
			cleanup, buf := setupCheckoutSettingsMockFactories(mockClient)
			defer cleanup()

			cmd := newCheckoutSettingsTestCmd()
			if tt.jsonOutput {
				_ = cmd.Flags().Set("output", "json")
			}

			err := checkoutSettingsGetCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// For JSON output, check the buffer
			if tt.jsonOutput && tt.wantOutput != "" {
				output := buf.String()
				if !strings.Contains(output, tt.wantOutput) {
					t.Errorf("output %q should contain %q", output, tt.wantOutput)
				}
			}
		})
	}
}

// TestCheckoutSettingsGetRunEWithTipping tests output when tipping is enabled with options.
func TestCheckoutSettingsGetRunEWithTipping(t *testing.T) {
	mockClient := &checkoutSettingsMockAPIClient{
		getCheckoutSettingsResp: &api.CheckoutSettings{
			ID:                   "cs_tipping",
			EnableTipping:        true,
			TippingOptions:       []float64{10, 15, 20},
			DefaultTippingOption: 15,
			UpdatedAt:            time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	cleanup, _ := setupCheckoutSettingsMockFactories(mockClient)
	defer cleanup()

	cmd := newCheckoutSettingsTestCmd()
	err := checkoutSettingsGetCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCheckoutSettingsGetRunEWithoutTipping tests output when tipping is disabled.
func TestCheckoutSettingsGetRunEWithoutTipping(t *testing.T) {
	mockClient := &checkoutSettingsMockAPIClient{
		getCheckoutSettingsResp: &api.CheckoutSettings{
			ID:             "cs_no_tipping",
			EnableTipping:  false,
			TippingOptions: nil,
			UpdatedAt:      time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	cleanup, _ := setupCheckoutSettingsMockFactories(mockClient)
	defer cleanup()

	cmd := newCheckoutSettingsTestCmd()
	err := checkoutSettingsGetCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCheckoutSettingsGetRunEWithAbandonedCart tests output when abandoned cart is enabled.
func TestCheckoutSettingsGetRunEWithAbandonedCart(t *testing.T) {
	mockClient := &checkoutSettingsMockAPIClient{
		getCheckoutSettingsResp: &api.CheckoutSettings{
			ID:                   "cs_abandoned",
			AbandonedCartEnabled: true,
			AbandonedCartDelay:   48,
			UpdatedAt:            time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	cleanup, _ := setupCheckoutSettingsMockFactories(mockClient)
	defer cleanup()

	cmd := newCheckoutSettingsTestCmd()
	err := checkoutSettingsGetCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCheckoutSettingsGetRunEWithPolicies tests output when policies are set.
func TestCheckoutSettingsGetRunEWithPolicies(t *testing.T) {
	mockClient := &checkoutSettingsMockAPIClient{
		getCheckoutSettingsResp: &api.CheckoutSettings{
			ID:                "cs_policies",
			TermsOfServiceURL: "https://example.com/terms",
			PrivacyPolicyURL:  "https://example.com/privacy",
			RefundPolicyURL:   "https://example.com/refund",
			UpdatedAt:         time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	cleanup, _ := setupCheckoutSettingsMockFactories(mockClient)
	defer cleanup()

	cmd := newCheckoutSettingsTestCmd()
	err := checkoutSettingsGetCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCheckoutSettingsGetRunEWithoutPolicies tests output when policies are empty.
func TestCheckoutSettingsGetRunEWithoutPolicies(t *testing.T) {
	mockClient := &checkoutSettingsMockAPIClient{
		getCheckoutSettingsResp: &api.CheckoutSettings{
			ID:                "cs_no_policies",
			TermsOfServiceURL: "",
			PrivacyPolicyURL:  "",
			RefundPolicyURL:   "",
			UpdatedAt:         time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	cleanup, _ := setupCheckoutSettingsMockFactories(mockClient)
	defer cleanup()

	cmd := newCheckoutSettingsTestCmd()
	err := checkoutSettingsGetCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCheckoutSettingsUpdateRunE tests the checkout settings update command with mock API.
func TestCheckoutSettingsUpdateRunE(t *testing.T) {
	tests := []struct {
		name       string
		flags      map[string]string
		mockResp   *api.CheckoutSettings
		mockErr    error
		jsonOutput bool
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful update guest checkout",
			flags: map[string]string{
				"guest-checkout": "true",
			},
			mockResp: &api.CheckoutSettings{
				ID:                  "cs_123",
				EnableGuestCheckout: true,
			},
		},
		{
			name: "successful update express checkout",
			flags: map[string]string{
				"express-checkout": "true",
			},
			mockResp: &api.CheckoutSettings{
				ID:                    "cs_456",
				EnableExpressCheckout: true,
			},
		},
		{
			name: "successful update tipping",
			flags: map[string]string{
				"tipping": "true",
			},
			mockResp: &api.CheckoutSettings{
				ID:            "cs_789",
				EnableTipping: true,
			},
		},
		{
			name: "successful update require phone",
			flags: map[string]string{
				"require-phone": "true",
			},
			mockResp: &api.CheckoutSettings{
				ID:           "cs_phone",
				RequirePhone: true,
			},
		},
		{
			name: "successful update order notes",
			flags: map[string]string{
				"order-notes": "true",
			},
			mockResp: &api.CheckoutSettings{
				ID:               "cs_notes",
				EnableOrderNotes: true,
			},
		},
		{
			name: "successful update abandoned cart",
			flags: map[string]string{
				"abandoned-cart": "true",
			},
			mockResp: &api.CheckoutSettings{
				ID:                   "cs_cart",
				AbandonedCartEnabled: true,
			},
		},
		{
			name: "successful update multiple fields",
			flags: map[string]string{
				"require-phone":    "true",
				"guest-checkout":   "true",
				"express-checkout": "false",
				"order-notes":      "true",
				"tipping":          "true",
				"abandoned-cart":   "true",
			},
			mockResp: &api.CheckoutSettings{
				ID:                    "cs_multi",
				RequirePhone:          true,
				EnableGuestCheckout:   true,
				EnableExpressCheckout: false,
				EnableOrderNotes:      true,
				EnableTipping:         true,
				AbandonedCartEnabled:  true,
			},
		},
		{
			name: "successful update with URLs",
			flags: map[string]string{
				"terms-url":   "https://example.com/terms",
				"privacy-url": "https://example.com/privacy",
				"refund-url":  "https://example.com/refund",
			},
			mockResp: &api.CheckoutSettings{
				ID:                "cs_urls",
				TermsOfServiceURL: "https://example.com/terms",
				PrivacyPolicyURL:  "https://example.com/privacy",
				RefundPolicyURL:   "https://example.com/refund",
			},
		},
		{
			name: "successful update with abandoned cart delay",
			flags: map[string]string{
				"abandoned-cart":       "true",
				"abandoned-cart-delay": "48",
			},
			mockResp: &api.CheckoutSettings{
				ID:                   "cs_delay",
				AbandonedCartEnabled: true,
				AbandonedCartDelay:   48,
			},
		},
		{
			name: "successful update with JSON output",
			flags: map[string]string{
				"guest-checkout": "true",
			},
			mockResp: &api.CheckoutSettings{
				ID:                  "cs_json",
				EnableGuestCheckout: true,
			},
			jsonOutput: true,
			wantOutput: "cs_json",
		},
		{
			name: "API error",
			flags: map[string]string{
				"guest-checkout": "true",
			},
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "network error",
			flags: map[string]string{
				"tipping": "false",
			},
			mockErr: errors.New("connection refused"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &checkoutSettingsMockAPIClient{
				updateCheckoutSettingsResp: tt.mockResp,
				updateCheckoutSettingsErr:  tt.mockErr,
			}
			cleanup, buf := setupCheckoutSettingsMockFactories(mockClient)
			defer cleanup()

			cmd := newCheckoutSettingsTestCmd()
			// Add update-specific flags
			cmd.Flags().Bool("require-phone", false, "")
			cmd.Flags().Bool("guest-checkout", false, "")
			cmd.Flags().Bool("express-checkout", false, "")
			cmd.Flags().Bool("order-notes", false, "")
			cmd.Flags().Bool("tipping", false, "")
			cmd.Flags().Bool("abandoned-cart", false, "")
			cmd.Flags().Int("abandoned-cart-delay", 24, "")
			cmd.Flags().String("terms-url", "", "")
			cmd.Flags().String("privacy-url", "", "")
			cmd.Flags().String("refund-url", "", "")

			for k, v := range tt.flags {
				_ = cmd.Flags().Set(k, v)
			}
			if tt.jsonOutput {
				_ = cmd.Flags().Set("output", "json")
			}

			err := checkoutSettingsUpdateCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// For JSON output, check the buffer
			if tt.jsonOutput && tt.wantOutput != "" {
				output := buf.String()
				if !strings.Contains(output, tt.wantOutput) {
					t.Errorf("output %q should contain %q", output, tt.wantOutput)
				}
			}
		})
	}
}

// TestCheckoutSettingsUpdateDryRun tests the dry-run mode for update command.
func TestCheckoutSettingsUpdateDryRun(t *testing.T) {
	// The mock client should NOT be called in dry-run mode
	mockClient := &checkoutSettingsMockAPIClient{
		updateCheckoutSettingsResp: nil,
		updateCheckoutSettingsErr:  errors.New("should not be called"),
	}
	cleanup, _ := setupCheckoutSettingsMockFactories(mockClient)
	defer cleanup()

	cmd := newCheckoutSettingsTestCmd()
	cmd.Flags().Bool("require-phone", false, "")
	cmd.Flags().Bool("guest-checkout", false, "")
	cmd.Flags().Bool("express-checkout", false, "")
	cmd.Flags().Bool("order-notes", false, "")
	cmd.Flags().Bool("tipping", false, "")
	cmd.Flags().Bool("abandoned-cart", false, "")
	cmd.Flags().Int("abandoned-cart-delay", 24, "")
	cmd.Flags().String("terms-url", "", "")
	cmd.Flags().String("privacy-url", "", "")
	cmd.Flags().String("refund-url", "", "")

	_ = cmd.Flags().Set("dry-run", "true")
	_ = cmd.Flags().Set("guest-checkout", "true")

	err := checkoutSettingsUpdateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error in dry-run mode: %v", err)
	}
}

// TestCheckoutSettingsUpdateGetClientError tests error handling when getClient fails.
func TestCheckoutSettingsUpdateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().Bool("require-phone", false, "")
	cmd.Flags().Bool("guest-checkout", false, "")
	cmd.Flags().Bool("express-checkout", false, "")
	cmd.Flags().Bool("order-notes", false, "")
	cmd.Flags().Bool("tipping", false, "")
	cmd.Flags().Bool("abandoned-cart", false, "")
	cmd.Flags().Int("abandoned-cart-delay", 24, "")
	cmd.Flags().String("terms-url", "", "")
	cmd.Flags().String("privacy-url", "", "")
	cmd.Flags().String("refund-url", "", "")

	_ = cmd.Flags().Set("guest-checkout", "true")

	err := checkoutSettingsUpdateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// TestCheckoutSettingsUpdateNoProfiles tests error when no profiles are configured.
func TestCheckoutSettingsUpdateNoProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{names: []string{}}, nil
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().Bool("require-phone", false, "")
	cmd.Flags().Bool("guest-checkout", false, "")
	cmd.Flags().Bool("express-checkout", false, "")
	cmd.Flags().Bool("order-notes", false, "")
	cmd.Flags().Bool("tipping", false, "")
	cmd.Flags().Bool("abandoned-cart", false, "")
	cmd.Flags().Int("abandoned-cart-delay", 24, "")
	cmd.Flags().String("terms-url", "", "")
	cmd.Flags().String("privacy-url", "", "")
	cmd.Flags().String("refund-url", "", "")

	_ = cmd.Flags().Set("guest-checkout", "true")

	err := checkoutSettingsUpdateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no store profiles configured") {
		t.Errorf("expected 'no store profiles' error, got: %v", err)
	}
}

// TestCheckoutSettingsWithMockStore tests checkout settings commands with a mock credential store.
func TestCheckoutSettingsWithMockStore(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()

	_ = os.Unsetenv("SHOPLINE_STORE")

	store := &mockStore{
		names: []string{"teststore"},
		creds: map[string]*secrets.StoreCredentials{
			"teststore": {Handle: "test-handle", AccessToken: "test-token"},
		},
	}

	secretsStoreFactory = func() (CredentialStore, error) {
		return store, nil
	}

	cmd := newTestCmdWithFlags()
	client, err := getClient(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Error("expected client, got nil")
	}
}

// TestCheckoutSettingsGetJSONOutput tests that JSON output is properly formatted.
func TestCheckoutSettingsGetJSONOutput(t *testing.T) {
	mockClient := &checkoutSettingsMockAPIClient{
		getCheckoutSettingsResp: &api.CheckoutSettings{
			ID:                    "cs_json_test",
			RequirePhone:          true,
			EnableGuestCheckout:   true,
			EnableExpressCheckout: false,
			EnableTipping:         true,
			TippingOptions:        []float64{10, 15, 20},
			AbandonedCartEnabled:  true,
			AbandonedCartDelay:    24,
			UpdatedAt:             time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupCheckoutSettingsMockFactories(mockClient)
	defer cleanup()

	cmd := newCheckoutSettingsTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := checkoutSettingsGetCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Check that key JSON fields are present
	expectedFields := []string{
		"cs_json_test",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("JSON output should contain %q, got:\n%s", field, output)
		}
	}
}

// TestCheckoutSettingsUpdateJSONOutput tests that JSON output is properly formatted for update.
func TestCheckoutSettingsUpdateJSONOutput(t *testing.T) {
	mockClient := &checkoutSettingsMockAPIClient{
		updateCheckoutSettingsResp: &api.CheckoutSettings{
			ID:                    "cs_update_json",
			EnableGuestCheckout:   true,
			EnableExpressCheckout: true,
			EnableTipping:         false,
		},
	}
	cleanup, buf := setupCheckoutSettingsMockFactories(mockClient)
	defer cleanup()

	cmd := newCheckoutSettingsTestCmd()
	cmd.Flags().Bool("require-phone", false, "")
	cmd.Flags().Bool("guest-checkout", false, "")
	cmd.Flags().Bool("express-checkout", false, "")
	cmd.Flags().Bool("order-notes", false, "")
	cmd.Flags().Bool("tipping", false, "")
	cmd.Flags().Bool("abandoned-cart", false, "")
	cmd.Flags().Int("abandoned-cart-delay", 24, "")
	cmd.Flags().String("terms-url", "", "")
	cmd.Flags().String("privacy-url", "", "")
	cmd.Flags().String("refund-url", "", "")

	_ = cmd.Flags().Set("output", "json")
	_ = cmd.Flags().Set("guest-checkout", "true")

	err := checkoutSettingsUpdateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "cs_update_json") {
		t.Errorf("JSON output should contain ID, got:\n%s", output)
	}
}

// TestCheckoutSettingsSubcommandShorts tests that subcommand short descriptions are correct.
func TestCheckoutSettingsSubcommandShorts(t *testing.T) {
	subcommands := map[string]string{
		"get":    "Get checkout settings",
		"update": "Update checkout settings",
	}

	for name, expectedShort := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range checkoutSettingsCmd.Commands() {
				if strings.HasPrefix(sub.Use, name) {
					found = true
					if sub.Short != expectedShort {
						t.Errorf("expected Short %q, got %q", expectedShort, sub.Short)
					}
					break
				}
			}
			if !found {
				t.Errorf("subcommand %q not found", name)
			}
		})
	}
}

// TestCheckoutSettingsParentShort tests the parent command short description.
func TestCheckoutSettingsParentShort(t *testing.T) {
	expectedShort := "Manage checkout settings"
	if checkoutSettingsCmd.Short != expectedShort {
		t.Errorf("expected Short %q, got %q", expectedShort, checkoutSettingsCmd.Short)
	}
}

// TestCheckoutSettingsUpdateAllBooleanFlags tests that all boolean flags work correctly.
func TestCheckoutSettingsUpdateAllBooleanFlags(t *testing.T) {
	boolFlags := []string{
		"require-phone",
		"guest-checkout",
		"express-checkout",
		"order-notes",
		"tipping",
		"abandoned-cart",
	}

	for _, flag := range boolFlags {
		t.Run(flag+" true", func(t *testing.T) {
			mockClient := &checkoutSettingsMockAPIClient{
				updateCheckoutSettingsResp: &api.CheckoutSettings{
					ID: "cs_" + flag,
				},
			}
			cleanup, _ := setupCheckoutSettingsMockFactories(mockClient)
			defer cleanup()

			cmd := newCheckoutSettingsTestCmd()
			for _, f := range boolFlags {
				cmd.Flags().Bool(f, false, "")
			}
			cmd.Flags().Int("abandoned-cart-delay", 24, "")
			cmd.Flags().String("terms-url", "", "")
			cmd.Flags().String("privacy-url", "", "")
			cmd.Flags().String("refund-url", "", "")

			_ = cmd.Flags().Set(flag, "true")

			err := checkoutSettingsUpdateCmd.RunE(cmd, []string{})
			if err != nil {
				t.Errorf("unexpected error setting %s=true: %v", flag, err)
			}
		})

		t.Run(flag+" false", func(t *testing.T) {
			mockClient := &checkoutSettingsMockAPIClient{
				updateCheckoutSettingsResp: &api.CheckoutSettings{
					ID: "cs_" + flag + "_false",
				},
			}
			cleanup, _ := setupCheckoutSettingsMockFactories(mockClient)
			defer cleanup()

			cmd := newCheckoutSettingsTestCmd()
			for _, f := range boolFlags {
				cmd.Flags().Bool(f, false, "")
			}
			cmd.Flags().Int("abandoned-cart-delay", 24, "")
			cmd.Flags().String("terms-url", "", "")
			cmd.Flags().String("privacy-url", "", "")
			cmd.Flags().String("refund-url", "", "")

			_ = cmd.Flags().Set(flag, "false")

			err := checkoutSettingsUpdateCmd.RunE(cmd, []string{})
			if err != nil {
				t.Errorf("unexpected error setting %s=false: %v", flag, err)
			}
		})
	}
}

// TestCheckoutSettingsUpdateStringFlags tests that all string flags work correctly.
func TestCheckoutSettingsUpdateStringFlags(t *testing.T) {
	stringFlags := map[string]string{
		"terms-url":   "https://example.com/terms",
		"privacy-url": "https://example.com/privacy",
		"refund-url":  "https://example.com/refund",
	}

	for flag, value := range stringFlags {
		t.Run(flag, func(t *testing.T) {
			mockClient := &checkoutSettingsMockAPIClient{
				updateCheckoutSettingsResp: &api.CheckoutSettings{
					ID: "cs_" + flag,
				},
			}
			cleanup, _ := setupCheckoutSettingsMockFactories(mockClient)
			defer cleanup()

			cmd := newCheckoutSettingsTestCmd()
			cmd.Flags().Bool("require-phone", false, "")
			cmd.Flags().Bool("guest-checkout", false, "")
			cmd.Flags().Bool("express-checkout", false, "")
			cmd.Flags().Bool("order-notes", false, "")
			cmd.Flags().Bool("tipping", false, "")
			cmd.Flags().Bool("abandoned-cart", false, "")
			cmd.Flags().Int("abandoned-cart-delay", 24, "")
			cmd.Flags().String("terms-url", "", "")
			cmd.Flags().String("privacy-url", "", "")
			cmd.Flags().String("refund-url", "", "")

			_ = cmd.Flags().Set(flag, value)

			err := checkoutSettingsUpdateCmd.RunE(cmd, []string{})
			if err != nil {
				t.Errorf("unexpected error setting %s=%s: %v", flag, value, err)
			}
		})
	}
}

// TestCheckoutSettingsUpdateIntFlags tests that int flags work correctly.
func TestCheckoutSettingsUpdateIntFlags(t *testing.T) {
	testCases := []struct {
		name  string
		value string
	}{
		{"delay 24", "24"},
		{"delay 48", "48"},
		{"delay 72", "72"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &checkoutSettingsMockAPIClient{
				updateCheckoutSettingsResp: &api.CheckoutSettings{
					ID: "cs_delay",
				},
			}
			cleanup, _ := setupCheckoutSettingsMockFactories(mockClient)
			defer cleanup()

			cmd := newCheckoutSettingsTestCmd()
			cmd.Flags().Bool("require-phone", false, "")
			cmd.Flags().Bool("guest-checkout", false, "")
			cmd.Flags().Bool("express-checkout", false, "")
			cmd.Flags().Bool("order-notes", false, "")
			cmd.Flags().Bool("tipping", false, "")
			cmd.Flags().Bool("abandoned-cart", false, "")
			cmd.Flags().Int("abandoned-cart-delay", 24, "")
			cmd.Flags().String("terms-url", "", "")
			cmd.Flags().String("privacy-url", "", "")
			cmd.Flags().String("refund-url", "", "")

			_ = cmd.Flags().Set("abandoned-cart-delay", tc.value)

			err := checkoutSettingsUpdateCmd.RunE(cmd, []string{})
			if err != nil {
				t.Errorf("unexpected error setting abandoned-cart-delay=%s: %v", tc.value, err)
			}
		})
	}
}
