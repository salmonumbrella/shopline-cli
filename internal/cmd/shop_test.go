package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

func TestShopInfoGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := shopInfoCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestShopSettingsGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := shopSettingsCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestShopCommandStructure(t *testing.T) {
	if shopCmd.Use != "shop" {
		t.Errorf("Expected Use 'shop', got %s", shopCmd.Use)
	}

	subcommands := shopCmd.Commands()
	expectedCmds := map[string]bool{
		"info":     false,
		"settings": false,
	}

	for _, cmd := range subcommands {
		if startsWithUse(cmd.Use, expectedCmds) {
			expectedCmds[getBaseUse(cmd.Use)] = true
		}
	}

	for name, found := range expectedCmds {
		if !found {
			t.Errorf("Expected subcommand %s not found", name)
		}
	}
}

// shopMockAPIClient is a mock implementation of api.APIClient for shop tests.
type shopMockAPIClient struct {
	api.MockClient
	getShopResp         *api.Shop
	getShopErr          error
	getShopSettingsResp *api.ShopSettings
	getShopSettingsErr  error
}

func (m *shopMockAPIClient) GetShop(ctx context.Context) (*api.Shop, error) {
	return m.getShopResp, m.getShopErr
}

func (m *shopMockAPIClient) GetShopSettings(ctx context.Context) (*api.ShopSettings, error) {
	return m.getShopSettingsResp, m.getShopSettingsErr
}

// setupShopMockFactories sets up mock factories for shop tests.
func setupShopMockFactories(mockClient *shopMockAPIClient) (func(), *bytes.Buffer) {
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

// newShopTestCmd creates a test command with common flags.
func newShopTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	return cmd
}

// TestShopInfoRunE tests the shop info command with mock API.
func TestShopInfoRunE(t *testing.T) {
	tests := []struct {
		name     string
		mockResp *api.Shop
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful get",
			mockResp: &api.Shop{
				ID:               "shop_123",
				Name:             "Test Store",
				Email:            "store@example.com",
				Domain:           "teststore.com",
				MyshoplineDomain: "teststore.myshopline.com",
				Phone:            "+1234567890",
				ShopOwner:        "John Doe",
				Address1:         "123 Main St",
				Address2:         "Suite 100",
				City:             "New York",
				Province:         "New York",
				ProvinceCode:     "NY",
				Country:          "United States",
				CountryCode:      "US",
				Zip:              "10001",
				Currency:         "USD",
				Timezone:         "America/New_York",
				WeightUnit:       "lb",
				PlanName:         "basic",
				PlanDisplayName:  "Basic Plan",
				CreatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "shop without address2",
			mockResp: &api.Shop{
				ID:       "shop_456",
				Name:     "Simple Store",
				Address1: "456 Oak Ave",
				Address2: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &shopMockAPIClient{
				getShopResp: tt.mockResp,
				getShopErr:  tt.mockErr,
			}
			cleanup, _ := setupShopMockFactories(mockClient)
			defer cleanup()

			cmd := newShopTestCmd()

			err := shopInfoCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestShopInfoRunEWithJSON tests JSON output format.
func TestShopInfoRunEWithJSON(t *testing.T) {
	mockClient := &shopMockAPIClient{
		getShopResp: &api.Shop{
			ID:   "shop_json",
			Name: "JSON Test Store",
		},
	}
	cleanup, buf := setupShopMockFactories(mockClient)
	defer cleanup()

	cmd := newShopTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := shopInfoCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "shop_json") {
		t.Errorf("JSON output should contain shop ID, got: %s", output)
	}
}

// TestShopSettingsRunE tests the shop settings command with mock API.
func TestShopSettingsRunE(t *testing.T) {
	tests := []struct {
		name     string
		mockResp *api.ShopSettings
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful get",
			mockResp: &api.ShopSettings{
				ID:                           "settings_123",
				Currency:                     "USD",
				WeightUnit:                   "lb",
				Timezone:                     "America/New_York",
				OrderPrefix:                  "ORD",
				OrderSuffix:                  "-2024",
				TaxesIncluded:                true,
				TaxShipping:                  false,
				AutomaticFulfillment:         true,
				EnabledPresentmentCurrencies: []string{"USD", "EUR", "GBP"},
			},
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "settings without currencies",
			mockResp: &api.ShopSettings{
				ID:                           "settings_456",
				Currency:                     "EUR",
				EnabledPresentmentCurrencies: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &shopMockAPIClient{
				getShopSettingsResp: tt.mockResp,
				getShopSettingsErr:  tt.mockErr,
			}
			cleanup, _ := setupShopMockFactories(mockClient)
			defer cleanup()

			cmd := newShopTestCmd()

			err := shopSettingsCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestShopSettingsRunEWithJSON tests JSON output format.
func TestShopSettingsRunEWithJSON(t *testing.T) {
	mockClient := &shopMockAPIClient{
		getShopSettingsResp: &api.ShopSettings{
			ID:       "settings_json",
			Currency: "CAD",
		},
	}
	cleanup, buf := setupShopMockFactories(mockClient)
	defer cleanup()

	cmd := newShopTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := shopSettingsCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "settings_json") {
		t.Errorf("JSON output should contain settings ID, got: %s", output)
	}
}
