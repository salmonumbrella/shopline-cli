package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

// merchantsMockAPIClient is a mock implementation of api.APIClient for merchants tests.
type merchantsMockAPIClient struct {
	api.MockClient

	getMerchantResp       *api.Merchant
	getMerchantErr        error
	listMerchantStaffResp *api.MerchantStaffListResponse
	listMerchantStaffErr  error
	getMerchantStaffResp  *api.MerchantStaff
	getMerchantStaffErr   error

	getMerchantByIDResp     json.RawMessage
	getMerchantByIDErr      error
	generateExpressLinkResp json.RawMessage
	generateExpressLinkErr  error
}

func (m *merchantsMockAPIClient) GetMerchant(ctx context.Context) (*api.Merchant, error) {
	return m.getMerchantResp, m.getMerchantErr
}

func (m *merchantsMockAPIClient) ListMerchantStaff(ctx context.Context, opts *api.MerchantStaffListOptions) (*api.MerchantStaffListResponse, error) {
	return m.listMerchantStaffResp, m.listMerchantStaffErr
}

func (m *merchantsMockAPIClient) GetMerchantStaff(ctx context.Context, id string) (*api.MerchantStaff, error) {
	return m.getMerchantStaffResp, m.getMerchantStaffErr
}

func (m *merchantsMockAPIClient) GetMerchantByID(ctx context.Context, merchantID string) (json.RawMessage, error) {
	return m.getMerchantByIDResp, m.getMerchantByIDErr
}

func (m *merchantsMockAPIClient) GenerateMerchantExpressLink(ctx context.Context, body any) (json.RawMessage, error) {
	return m.generateExpressLinkResp, m.generateExpressLinkErr
}

// setupMerchantsMockFactories configures factories for merchants command tests.
func setupMerchantsMockFactories(mockClient *merchantsMockAPIClient) (cleanup func()) {
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
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}
	var buf bytes.Buffer
	formatterWriter = &buf

	return func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}
}

// TestMerchantsCommandSetup verifies merchants command initialization
func TestMerchantsCommandSetup(t *testing.T) {
	if merchantsCmd.Use != "merchants" {
		t.Errorf("expected Use 'merchants', got %q", merchantsCmd.Use)
	}
	if merchantsCmd.Short != "View merchant information" {
		t.Errorf("expected Short 'View merchant information', got %q", merchantsCmd.Short)
	}
}

// TestMerchantsSubcommands verifies all subcommands are registered
func TestMerchantsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"get":          "Get current merchant details",
		"get-by-id":    "Get merchant details by id (documented endpoint; raw JSON)",
		"staff":        "Manage merchant staff",
		"express-link": "Generate express cart link (documented endpoint)",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range merchantsCmd.Commands() {
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

func TestMerchantsGetByIDRunE_WithMockAPI(t *testing.T) {
	mockClient := &merchantsMockAPIClient{
		getMerchantByIDResp: json.RawMessage(`{"id":"m_123"}`),
	}
	cleanup := setupMerchantsMockFactories(mockClient)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")

	if err := merchantsGetByIDCmd.RunE(cmd, []string{"m_123"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMerchantsExpressLinkGenerateRunE_WithMockAPI(t *testing.T) {
	mockClient := &merchantsMockAPIClient{
		generateExpressLinkResp: json.RawMessage(`{"url":"https://example.com/express"}`),
	}
	cleanup := setupMerchantsMockFactories(mockClient)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().String("body", `{"ok":true}`, "")
	cmd.Flags().String("body-file", "", "")

	if err := merchantsExpressLinkGenerateCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestMerchantsGetRunE tests the merchants get command execution with mock API.
func TestMerchantsGetRunE(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	tests := []struct {
		name       string
		mockResp   *api.Merchant
		mockErr    error
		outputJSON bool
		wantErr    bool
	}{
		{
			name: "successful get with all fields",
			mockResp: &api.Merchant{
				ID:              "merchant_123",
				Name:            "Test Store",
				Handle:          "test-store",
				Email:           "owner@test.com",
				Phone:           "+1234567890",
				Domain:          "test-store.myshopline.com",
				PrimaryDomain:   "www.test-store.com",
				Currency:        "USD",
				Timezone:        "America/New_York",
				Country:         "United States",
				CountryCode:     "US",
				Province:        "New York",
				City:            "New York",
				Address1:        "123 Main St",
				Address2:        "Suite 100",
				Zip:             "10001",
				Plan:            "premium",
				PlanDisplayName: "Premium Plan",
				ShopOwner:       "John Doe",
				WeightUnit:      "lb",
				TaxesIncluded:   true,
				TaxShipping:     false,
				PasswordEnabled: false,
				SetupRequired:   false,
				Features: &api.MerchantFeatures{
					Checkout:      true,
					MultiLocation: true,
					MultiCurrency: true,
					GiftCards:     true,
					Subscriptions: true,
					Discounts:     true,
				},
				Finances: &api.MerchantFinances{
					EnabledPresentmentCurrencies: []string{"USD", "EUR", "GBP"},
				},
				CreatedAt: now,
				UpdatedAt: now.Add(24 * time.Hour),
			},
		},
		{
			name: "successful get without features",
			mockResp: &api.Merchant{
				ID:              "merchant_456",
				Name:            "Simple Store",
				Handle:          "simple-store",
				Email:           "simple@test.com",
				Currency:        "CAD",
				Timezone:        "America/Toronto",
				Country:         "Canada",
				CountryCode:     "CA",
				Province:        "Ontario",
				City:            "Toronto",
				Address1:        "456 Oak Ave",
				Plan:            "basic",
				PlanDisplayName: "Basic Plan",
				ShopOwner:       "Jane Smith",
				WeightUnit:      "kg",
				CreatedAt:       now,
				UpdatedAt:       now,
			},
		},
		{
			name: "successful get with empty finances currencies",
			mockResp: &api.Merchant{
				ID:        "merchant_789",
				Name:      "Minimal Store",
				Handle:    "minimal",
				Email:     "min@test.com",
				Currency:  "EUR",
				Country:   "Germany",
				Finances:  &api.MerchantFinances{EnabledPresentmentCurrencies: []string{}},
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		{
			name:       "successful get with JSON output",
			outputJSON: true,
			mockResp: &api.Merchant{
				ID:        "merchant_json",
				Name:      "JSON Store",
				Handle:    "json-store",
				Email:     "json@test.com",
				Currency:  "USD",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name:    "authentication error",
			mockErr: errors.New("unauthorized"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &merchantsMockAPIClient{
				getMerchantResp: tt.mockResp,
				getMerchantErr:  tt.mockErr,
			}
			cleanup := setupMerchantsMockFactories(mockClient)
			defer cleanup()

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")

			if tt.outputJSON {
				_ = cmd.Flags().Set("output", "json")
			}

			err := merchantsGetCmd.RunE(cmd, []string{})

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

// TestMerchantsStaffListRunE tests the merchants staff list command execution with mock API.
func TestMerchantsStaffListRunE(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	lastLogin := now.Add(-48 * time.Hour)

	tests := []struct {
		name       string
		mockResp   *api.MerchantStaffListResponse
		mockErr    error
		outputJSON bool
		wantErr    bool
	}{
		{
			name: "successful list with multiple staff",
			mockResp: &api.MerchantStaffListResponse{
				Items: []api.MerchantStaff{
					{
						ID:           "staff_1",
						Email:        "owner@test.com",
						FirstName:    "John",
						LastName:     "Doe",
						Role:         "owner",
						AccountOwner: true,
						Active:       true,
						LastLoginAt:  &lastLogin,
						CreatedAt:    now,
						UpdatedAt:    now,
					},
					{
						ID:           "staff_2",
						Email:        "admin@test.com",
						FirstName:    "Jane",
						LastName:     "Smith",
						Role:         "admin",
						AccountOwner: false,
						Active:       true,
						LastLoginAt:  nil,
						CreatedAt:    now,
						UpdatedAt:    now,
					},
					{
						ID:           "staff_3",
						Email:        "inactive@test.com",
						FirstName:    "Bob",
						LastName:     "Inactive",
						Role:         "staff",
						AccountOwner: false,
						Active:       false,
						LastLoginAt:  nil,
						CreatedAt:    now,
						UpdatedAt:    now,
					},
				},
				TotalCount: 3,
			},
		},
		{
			name: "empty staff list",
			mockResp: &api.MerchantStaffListResponse{
				Items:      []api.MerchantStaff{},
				TotalCount: 0,
			},
		},
		{
			name:       "successful list with JSON output",
			outputJSON: true,
			mockResp: &api.MerchantStaffListResponse{
				Items: []api.MerchantStaff{
					{
						ID:        "staff_json",
						Email:     "json@test.com",
						FirstName: "JSON",
						LastName:  "User",
						Role:      "admin",
						Active:    true,
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
				TotalCount: 1,
			},
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name:    "permission denied",
			mockErr: errors.New("forbidden"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &merchantsMockAPIClient{
				listMerchantStaffResp: tt.mockResp,
				listMerchantStaffErr:  tt.mockErr,
			}
			cleanup := setupMerchantsMockFactories(mockClient)
			defer cleanup()

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("role", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			if tt.outputJSON {
				_ = cmd.Flags().Set("output", "json")
			}

			err := merchantsStaffListCmd.RunE(cmd, []string{})

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

// TestMerchantsStaffGetRunE tests the merchants staff get command execution with mock API.
func TestMerchantsStaffGetRunE(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	lastLogin := now.Add(-24 * time.Hour)

	tests := []struct {
		name       string
		staffID    string
		mockResp   *api.MerchantStaff
		mockErr    error
		outputJSON bool
		wantErr    bool
	}{
		{
			name:    "successful get with all fields",
			staffID: "staff_123",
			mockResp: &api.MerchantStaff{
				ID:           "staff_123",
				Email:        "staff@test.com",
				FirstName:    "John",
				LastName:     "Doe",
				Phone:        "+1234567890",
				Role:         "admin",
				Permissions:  []string{"products", "orders", "customers"},
				AccountOwner: false,
				Active:       true,
				LastLoginAt:  &lastLogin,
				CreatedAt:    now,
				UpdatedAt:    now.Add(time.Hour),
			},
		},
		{
			name:    "successful get owner with no permissions listed",
			staffID: "staff_owner",
			mockResp: &api.MerchantStaff{
				ID:           "staff_owner",
				Email:        "owner@test.com",
				FirstName:    "Jane",
				LastName:     "Owner",
				Role:         "owner",
				Permissions:  []string{},
				AccountOwner: true,
				Active:       true,
				LastLoginAt:  nil,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
		},
		{
			name:    "successful get inactive staff",
			staffID: "staff_inactive",
			mockResp: &api.MerchantStaff{
				ID:           "staff_inactive",
				Email:        "inactive@test.com",
				FirstName:    "Bob",
				LastName:     "Former",
				Role:         "staff",
				Permissions:  []string{},
				AccountOwner: false,
				Active:       false,
				LastLoginAt:  nil,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
		},
		{
			name:       "successful get with JSON output",
			staffID:    "staff_json",
			outputJSON: true,
			mockResp: &api.MerchantStaff{
				ID:        "staff_json",
				Email:     "json@test.com",
				FirstName: "JSON",
				LastName:  "User",
				Role:      "admin",
				Active:    true,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		{
			name:    "staff not found",
			staffID: "staff_nonexistent",
			mockErr: errors.New("staff not found"),
			wantErr: true,
		},
		{
			name:    "API error",
			staffID: "staff_error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &merchantsMockAPIClient{
				getMerchantStaffResp: tt.mockResp,
				getMerchantStaffErr:  tt.mockErr,
			}
			cleanup := setupMerchantsMockFactories(mockClient)
			defer cleanup()

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")

			if tt.outputJSON {
				_ = cmd.Flags().Set("output", "json")
			}

			err := merchantsStaffGetCmd.RunE(cmd, []string{tt.staffID})

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

// TestMerchantsGetGetClientError verifies error handling when getClient fails for get command
func TestMerchantsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := merchantsGetCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// TestMerchantsStaffListGetClientError verifies error handling when getClient fails for staff list command
func TestMerchantsStaffListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("role", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := merchantsStaffListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// TestMerchantsStaffGetGetClientError verifies error handling when getClient fails for staff get command
func TestMerchantsStaffGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := merchantsStaffGetCmd.RunE(cmd, []string{"staff-123"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// TestMerchantsStaffListFlags verifies list command flags exist with correct defaults
func TestMerchantsStaffListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"role", ""},
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := merchantsStaffListCmd.Flags().Lookup(f.name)
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

// TestMerchantsCommandStructure verifies the command hierarchy
func TestMerchantsCommandStructure(t *testing.T) {
	if merchantsCmd.Use != "merchants" {
		t.Errorf("Expected Use 'merchants', got %s", merchantsCmd.Use)
	}

	subcommands := merchantsCmd.Commands()
	expectedCmds := map[string]bool{
		"get":   false,
		"staff": false,
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

// TestMerchantsStaffCommandStructure verifies the staff command hierarchy
func TestMerchantsStaffCommandStructure(t *testing.T) {
	subcommands := merchantsStaffCmd.Commands()
	expectedCmds := map[string]bool{
		"list": false,
		"get":  false,
	}

	for _, cmd := range subcommands {
		if startsWithUse(cmd.Use, expectedCmds) {
			expectedCmds[getBaseUse(cmd.Use)] = true
		}
	}

	for name, found := range expectedCmds {
		if !found {
			t.Errorf("Expected staff subcommand %s not found", name)
		}
	}
}

// TestMerchantsStaffGetCmdUse verifies the get command has correct use string
func TestMerchantsStaffGetCmdUse(t *testing.T) {
	if merchantsStaffGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", merchantsStaffGetCmd.Use)
	}
}

// TestMerchantsGetRunE_NoProfiles verifies error when no profiles are configured
func TestMerchantsGetRunE_NoProfiles(t *testing.T) {
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
	err := merchantsGetCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

// TestMerchantsStaffListRunE_NoProfiles verifies error when no profiles are configured
func TestMerchantsStaffListRunE_NoProfiles(t *testing.T) {
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
	cmd.Flags().String("role", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := merchantsStaffListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

// TestMerchantsStaffGetRunE_MultipleProfiles verifies error when multiple profiles exist without selection
func TestMerchantsStaffGetRunE_MultipleProfiles(t *testing.T) {
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
	err := merchantsStaffGetCmd.RunE(cmd, []string{"staff_123"})
	if err == nil {
		t.Fatal("Expected error for multiple profiles, got nil")
	}
}

// TestMerchantsGetRunE_WithRoleFilter verifies staff list with role filter
func TestMerchantsStaffListRunE_WithRoleFilter(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &merchantsMockAPIClient{
		listMerchantStaffResp: &api.MerchantStaffListResponse{
			Items: []api.MerchantStaff{
				{
					ID:        "staff_admin_1",
					Email:     "admin1@test.com",
					FirstName: "Admin",
					LastName:  "One",
					Role:      "admin",
					Active:    true,
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup := setupMerchantsMockFactories(mockClient)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("role", "admin", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := merchantsStaffListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestMerchantsGetRunE_OutputContainsExpectedFields verifies output formatting
func TestMerchantsGetRunE_OutputContainsExpectedFields(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	mockClient := &merchantsMockAPIClient{
		getMerchantResp: &api.Merchant{
			ID:              "merchant_output_test",
			Name:            "Output Test Store",
			Handle:          "output-test",
			Email:           "output@test.com",
			Phone:           "+1234567890",
			Domain:          "output.myshopline.com",
			PrimaryDomain:   "www.output.com",
			Currency:        "USD",
			Timezone:        "UTC",
			Country:         "United States",
			CountryCode:     "US",
			Province:        "California",
			City:            "San Francisco",
			Address1:        "100 Test St",
			Address2:        "Floor 5",
			Zip:             "94107",
			Plan:            "enterprise",
			PlanDisplayName: "Enterprise Plan",
			ShopOwner:       "Test Owner",
			WeightUnit:      "lb",
			TaxesIncluded:   true,
			TaxShipping:     true,
			PasswordEnabled: false,
			SetupRequired:   false,
			Features: &api.MerchantFeatures{
				Checkout:      true,
				MultiLocation: true,
				MultiCurrency: false,
				GiftCards:     true,
				Subscriptions: false,
				Discounts:     true,
			},
			Finances: &api.MerchantFinances{
				EnabledPresentmentCurrencies: []string{"USD", "CAD"},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := merchantsGetCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	expectedFields := []string{
		"merchant_output_test",
		"Output Test Store",
		"output-test",
		"output@test.com",
		"United States",
		"Enterprise Plan",
		"USD, CAD",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("expected output to contain %q", field)
		}
	}
}
