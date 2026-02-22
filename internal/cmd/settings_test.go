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

// settingsAPIClient is a mock implementation of api.APIClient for settings tests.
type settingsAPIClient struct {
	api.MockClient

	getMerchantResp    *api.Merchant
	getMerchantErr     error
	getSettingsResp    *api.SettingsResponse
	getSettingsErr     error
	updateSettingsResp *api.SettingsResponse
	updateSettingsErr  error
}

func (m *settingsAPIClient) GetMerchant(ctx context.Context) (*api.Merchant, error) {
	return m.getMerchantResp, m.getMerchantErr
}

func (m *settingsAPIClient) GetSettings(ctx context.Context) (*api.SettingsResponse, error) {
	return m.getSettingsResp, m.getSettingsErr
}

func (m *settingsAPIClient) UpdateSettings(ctx context.Context, req *api.UserSettingsUpdateRequest) (*api.SettingsResponse, error) {
	return m.updateSettingsResp, m.updateSettingsErr
}

// setupSettingsTest sets up the test environment for settings tests.
func setupSettingsTest(t *testing.T) (cleanup func(), buf *bytes.Buffer) {
	t.Helper()

	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	buf = &bytes.Buffer{}
	formatterWriter = buf

	cleanup = func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}

	return cleanup, buf
}

func TestSettingsCommandSetup(t *testing.T) {
	if settingsCmd.Use != "settings" {
		t.Errorf("expected Use 'settings', got %q", settingsCmd.Use)
	}
	if settingsCmd.Short != "Manage store settings" {
		t.Errorf("expected Short 'Manage store settings', got %q", settingsCmd.Short)
	}
}

func TestSettingsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"get":    "Get store settings",
		"update": "Update user settings",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range settingsCmd.Commands() {
				if sub.Use == name || (len(sub.Use) > len(name) && sub.Use[:len(name)] == name) {
					found = true
					if sub.Short != short {
						t.Errorf("expected Short %q, got %q", short, sub.Short)
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

func TestSettingsUpdateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"min-age-limit", ""},
		{"pos-apply-credit", "false"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := settingsUpdateCmd.Flags().Lookup(f.name)
			if flag == nil {
				t.Errorf("flag %q not found", f.name)
				return
			}
			if flag.DefValue != f.defaultValue {
				t.Errorf("expected default %q, got %q", f.defaultValue, flag.DefValue)
			}
		})
	}
}

func TestSettingsGetRunE_TextOutput(t *testing.T) {
	cleanup, _ := setupSettingsTest(t)
	defer cleanup()

	tests := []struct {
		name         string
		merchantResp *api.Merchant
		merchantErr  error
		settingsResp *api.SettingsResponse
		settingsErr  error
		wantErr      bool
		wantOutput   []string
	}{
		{
			name: "successful get with all fields",
			merchantResp: &api.Merchant{
				ID:            "merchant_123",
				Name:          "My Store",
				Email:         "store@example.com",
				Domain:        "mystore.myshopline.com",
				Phone:         "+1234567890",
				Currency:      "USD",
				Timezone:      "America/New_York",
				WeightUnit:    "kg",
				TaxesIncluded: true,
				TaxShipping:   false,
				CountryCode:   "US",
				Province:      "NY",
				Address1:      "123 Main St",
				Address2:      "Suite 100",
				City:          "New York",
				Zip:           "10001",
				CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC),
			},
			settingsResp: &api.SettingsResponse{
				Users: api.UserSettings{
					MinimumAgeLimit: "18",
					PosApplyCredit:  true,
				},
			},
			wantOutput: []string{
				"Store Settings",
				"Name:",
				"My Store",
				"Email:",
				"store@example.com",
				"Currency:",
				"USD",
				"Min Age Limit:",
				"18",
			},
		},
		{
			name:        "merchant API error",
			merchantErr: errors.New("API unavailable"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &settingsAPIClient{
				getMerchantResp: tt.merchantResp,
				getMerchantErr:  tt.merchantErr,
				getSettingsResp: tt.settingsResp,
				getSettingsErr:  tt.settingsErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "text", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")

			output, err := captureStdout(func() error {
				return settingsGetCmd.RunE(cmd, []string{})
			})

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

			for _, expected := range tt.wantOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain %q, got %q", expected, output)
				}
			}
		})
	}
}

func TestSettingsGetRunE_JSONOutput(t *testing.T) {
	cleanup, buf := setupSettingsTest(t)
	defer cleanup()

	mockClient := &settingsAPIClient{
		getMerchantResp: &api.Merchant{
			ID:       "merchant_123",
			Name:     "JSON Store",
			Email:    "json@example.com",
			Domain:   "json.myshopline.com",
			Currency: "USD",
		},
		getSettingsResp: &api.SettingsResponse{
			Users: api.UserSettings{
				MinimumAgeLimit: "13",
				PosApplyCredit:  false,
			},
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := settingsGetCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "JSON Store") {
		t.Errorf("expected JSON output to contain store name, got %q", output)
	}
}

func TestSettingsGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := settingsGetCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestSettingsGetRunE_NoProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()

	_ = os.Unsetenv("SHOPLINE_STORE")
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{names: []string{}}, nil
	}

	cmd := newTestCmdWithFlags()
	err := settingsGetCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

func TestSettingsGetRunE_MultipleProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()

	_ = os.Unsetenv("SHOPLINE_STORE")
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"store1", "store2"},
			creds: map[string]*secrets.StoreCredentials{
				"store1": {Handle: "test1", AccessToken: "token1"},
				"store2": {Handle: "test2", AccessToken: "token2"},
			},
		}, nil
	}

	cmd := newTestCmdWithFlags()
	err := settingsGetCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for multiple profiles, got nil")
	}
}

func TestSettingsUpdateRunE_TextOutput(t *testing.T) {
	cleanup, _ := setupSettingsTest(t)
	defer cleanup()

	tests := []struct {
		name       string
		flags      map[string]string
		boolFlags  map[string]bool
		mockResp   *api.SettingsResponse
		mockErr    error
		wantErr    bool
		wantOutput []string
	}{
		{
			name: "successful update",
			flags: map[string]string{
				"min-age-limit": "21",
			},
			mockResp: &api.SettingsResponse{
				Users: api.UserSettings{
					MinimumAgeLimit: "21",
					PosApplyCredit:  false,
				},
			},
			wantOutput: []string{
				"User settings updated successfully",
				"Min Age Limit:",
				"21",
			},
		},
		{
			name: "update with boolean flag",
			flags: map[string]string{
				"min-age-limit": "18",
			},
			boolFlags: map[string]bool{
				"pos-apply-credit": true,
			},
			mockResp: &api.SettingsResponse{
				Users: api.UserSettings{
					MinimumAgeLimit: "18",
					PosApplyCredit:  true,
				},
			},
			wantOutput: []string{
				"User settings updated successfully",
				"POS Apply Credit:",
				"true",
			},
		},
		{
			name:    "API error",
			flags:   map[string]string{"min-age-limit": "21"},
			mockErr: errors.New("update failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &settingsAPIClient{
				updateSettingsResp: tt.mockResp,
				updateSettingsErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "text", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("min-age-limit", "", "")
			cmd.Flags().Bool("pos-apply-credit", false, "")
			cmd.Flags().Bool("dry-run", false, "")

			for k, v := range tt.flags {
				_ = cmd.Flags().Set(k, v)
			}
			for k, v := range tt.boolFlags {
				if v {
					_ = cmd.Flags().Set(k, "true")
				}
			}

			output, err := captureStdout(func() error {
				return settingsUpdateCmd.RunE(cmd, []string{})
			})

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

			for _, expected := range tt.wantOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain %q, got %q", expected, output)
				}
			}
		})
	}
}

func TestSettingsUpdateRunE_JSONOutput(t *testing.T) {
	cleanup, buf := setupSettingsTest(t)
	defer cleanup()

	mockClient := &settingsAPIClient{
		updateSettingsResp: &api.SettingsResponse{
			Users: api.UserSettings{
				MinimumAgeLimit: "21",
				PosApplyCredit:  true,
			},
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("min-age-limit", "21", "")
	cmd.Flags().Bool("pos-apply-credit", false, "")
	cmd.Flags().Bool("dry-run", false, "")

	err := settingsUpdateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "21") {
		t.Errorf("expected JSON output to contain minimum_age_limit, got %q", output)
	}
}

func TestSettingsUpdateRunE_DryRun(t *testing.T) {
	cleanup, _ := setupSettingsTest(t)
	defer cleanup()

	// Set up a mock client that should NOT be called in dry-run mode
	mockClient := &settingsAPIClient{
		updateSettingsErr: errors.New("should not be called"),
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("min-age-limit", "21", "")
	cmd.Flags().Bool("pos-apply-credit", false, "")
	cmd.Flags().Bool("dry-run", true, "")

	output, err := captureStdout(func() error {
		return settingsUpdateCmd.RunE(cmd, []string{})
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "[DRY-RUN]") {
		t.Errorf("expected dry-run message, got %q", output)
	}
}

func TestSettingsUpdateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("min-age-limit", "21", "")
	cmd.Flags().Bool("pos-apply-credit", false, "")

	err := settingsUpdateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestSettingsUpdateRunE_NoProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()

	_ = os.Unsetenv("SHOPLINE_STORE")
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{names: []string{}}, nil
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("min-age-limit", "21", "")
	cmd.Flags().Bool("pos-apply-credit", false, "")

	err := settingsUpdateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

func TestSettingsUpdateRunE_MultipleProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()

	_ = os.Unsetenv("SHOPLINE_STORE")
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"store1", "store2"},
			creds: map[string]*secrets.StoreCredentials{
				"store1": {Handle: "test1", AccessToken: "token1"},
				"store2": {Handle: "test2", AccessToken: "token2"},
			},
		}, nil
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("min-age-limit", "21", "")
	cmd.Flags().Bool("pos-apply-credit", false, "")

	err := settingsUpdateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for multiple profiles, got nil")
	}
}

func TestSettingsWithMockStore(t *testing.T) {
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
