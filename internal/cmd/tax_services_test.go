package cmd

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

// taxServicesAPIClient is a mock implementation of api.APIClient for tax services tests.
type taxServicesAPIClient struct {
	api.MockClient

	listTaxServicesResp  *api.TaxServicesListResponse
	listTaxServicesErr   error
	getTaxServiceResp    *api.TaxService
	getTaxServiceErr     error
	createTaxServiceResp *api.TaxService
	createTaxServiceErr  error
	updateTaxServiceResp *api.TaxService
	updateTaxServiceErr  error
	deleteTaxServiceErr  error
}

func (m *taxServicesAPIClient) ListTaxServices(ctx context.Context, opts *api.TaxServicesListOptions) (*api.TaxServicesListResponse, error) {
	return m.listTaxServicesResp, m.listTaxServicesErr
}

func (m *taxServicesAPIClient) GetTaxService(ctx context.Context, id string) (*api.TaxService, error) {
	return m.getTaxServiceResp, m.getTaxServiceErr
}

func (m *taxServicesAPIClient) CreateTaxService(ctx context.Context, req *api.TaxServiceCreateRequest) (*api.TaxService, error) {
	return m.createTaxServiceResp, m.createTaxServiceErr
}

func (m *taxServicesAPIClient) UpdateTaxService(ctx context.Context, id string, req *api.TaxServiceUpdateRequest) (*api.TaxService, error) {
	return m.updateTaxServiceResp, m.updateTaxServiceErr
}

func (m *taxServicesAPIClient) DeleteTaxService(ctx context.Context, id string) error {
	return m.deleteTaxServiceErr
}

// setupTaxServicesTest initializes common test fixtures for tax services tests.
func setupTaxServicesTest(t *testing.T) (cleanup func()) {
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

	return func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}
}

// TestTaxServicesCommand verifies tax-services command initialization.
func TestTaxServicesCommand(t *testing.T) {
	if taxServicesCmd == nil {
		t.Fatal("taxServicesCmd is nil")
	}
	if taxServicesCmd.Use != "tax-services" {
		t.Errorf("Expected Use to be 'tax-services', got %q", taxServicesCmd.Use)
	}
	if taxServicesCmd.Short != "Manage tax service providers" {
		t.Errorf("Expected Short 'Manage tax service providers', got %q", taxServicesCmd.Short)
	}
}

// TestTaxServicesAliases verifies command aliases.
func TestTaxServicesAliases(t *testing.T) {
	expectedAliases := []string{"tax-service", "ts"}
	aliases := taxServicesCmd.Aliases
	if len(aliases) != len(expectedAliases) {
		t.Errorf("Expected %d aliases, got %d", len(expectedAliases), len(aliases))
	}
	// Check that all expected aliases are present (order-agnostic)
	for _, expected := range expectedAliases {
		found := false
		for _, actual := range aliases {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected alias %q not found in %v", expected, aliases)
		}
	}
}

// TestTaxServicesSubcommands verifies all subcommands are registered.
func TestTaxServicesSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List tax services",
		"get":    "Get tax service details",
		"create": "Create a tax service",
		"update": "Update a tax service",
		"delete": "Delete a tax service",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range taxServicesCmd.Commands() {
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

// TestTaxServicesListFlags verifies list command flags exist with correct defaults.
func TestTaxServicesListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
		{"provider", ""},
		{"active", "false"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := taxServicesListCmd.Flags().Lookup(f.name)
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

// TestTaxServicesCreateFlags verifies create command flags exist with correct defaults.
func TestTaxServicesCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"name", ""},
		{"provider", ""},
		{"api-key", ""},
		{"api-secret", ""},
		{"sandbox", "false"},
		{"active", "true"},
		{"callback-url", ""},
		{"countries", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := taxServicesCreateCmd.Flags().Lookup(f.name)
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

// TestTaxServicesUpdateFlags verifies update command flags exist with correct defaults.
func TestTaxServicesUpdateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"name", ""},
		{"api-key", ""},
		{"api-secret", ""},
		{"sandbox", "false"},
		{"active", "false"},
		{"callback-url", ""},
		{"countries", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := taxServicesUpdateCmd.Flags().Lookup(f.name)
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

// TestTaxServicesDeleteFlags verifies delete command flags exist with correct defaults.
func TestTaxServicesDeleteFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"yes", "false"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := taxServicesDeleteCmd.Flags().Lookup(f.name)
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

// TestTaxServicesGetArgsValidation verifies get command args validation.
func TestTaxServicesGetArgsValidation(t *testing.T) {
	if taxServicesGetCmd.Args == nil {
		t.Fatal("Expected Args validator on get command")
	}
	if taxServicesGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", taxServicesGetCmd.Use)
	}
	err := taxServicesGetCmd.Args(taxServicesGetCmd, []string{})
	if err == nil {
		t.Error("Expected error with no args")
	}
	err = taxServicesGetCmd.Args(taxServicesGetCmd, []string{"svc_123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}
}

// TestTaxServicesUpdateArgsValidation verifies update command args validation.
func TestTaxServicesUpdateArgsValidation(t *testing.T) {
	if taxServicesUpdateCmd.Args == nil {
		t.Fatal("Expected Args validator on update command")
	}
	if taxServicesUpdateCmd.Use != "update <id>" {
		t.Errorf("expected Use 'update <id>', got %q", taxServicesUpdateCmd.Use)
	}
	err := taxServicesUpdateCmd.Args(taxServicesUpdateCmd, []string{})
	if err == nil {
		t.Error("Expected error with no args")
	}
	err = taxServicesUpdateCmd.Args(taxServicesUpdateCmd, []string{"svc_123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}
}

// TestTaxServicesDeleteArgsValidation verifies delete command args validation.
func TestTaxServicesDeleteArgsValidation(t *testing.T) {
	if taxServicesDeleteCmd.Args == nil {
		t.Fatal("Expected Args validator on delete command")
	}
	if taxServicesDeleteCmd.Use != "delete <id>" {
		t.Errorf("expected Use 'delete <id>', got %q", taxServicesDeleteCmd.Use)
	}
	err := taxServicesDeleteCmd.Args(taxServicesDeleteCmd, []string{})
	if err == nil {
		t.Error("Expected error with no args")
	}
	err = taxServicesDeleteCmd.Args(taxServicesDeleteCmd, []string{"svc_123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}
}

// TestTaxServicesListRunE tests the tax-services list command execution with mock API.
func TestTaxServicesListRunE(t *testing.T) {
	cleanup := setupTaxServicesTest(t)
	defer cleanup()

	tests := []struct {
		name       string
		mockResp   *api.TaxServicesListResponse
		mockErr    error
		jsonOutput bool
		wantErr    bool
	}{
		{
			name: "successful list",
			mockResp: &api.TaxServicesListResponse{
				Items: []api.TaxService{
					{
						ID:        "svc_123",
						Name:      "Avalara Tax",
						Provider:  "avalara",
						Sandbox:   false,
						Active:    true,
						Countries: []string{"US", "CA"},
						CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
						UpdatedAt: time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantErr: false,
		},
		{
			name: "list with sandbox services",
			mockResp: &api.TaxServicesListResponse{
				Items: []api.TaxService{
					{
						ID:        "svc_456",
						Name:      "TaxJar Test",
						Provider:  "taxjar",
						Sandbox:   true,
						Active:    false,
						Countries: []string{"US"},
						CreatedAt: time.Date(2024, 2, 1, 8, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantErr: false,
		},
		{
			name: "list with many countries (truncated)",
			mockResp: &api.TaxServicesListResponse{
				Items: []api.TaxService{
					{
						ID:        "svc_789",
						Name:      "Global Tax",
						Provider:  "avalara",
						Active:    true,
						Countries: []string{"US", "CA", "MX", "GB", "FR", "DE", "ES", "IT"},
					},
				},
				TotalCount: 1,
			},
			wantErr: false,
		},
		{
			name: "successful list with JSON output",
			mockResp: &api.TaxServicesListResponse{
				Items: []api.TaxService{
					{
						ID:       "svc_json",
						Name:     "JSON Test",
						Provider: "taxjar",
						Active:   true,
					},
				},
				TotalCount: 1,
			},
			jsonOutput: true,
			wantErr:    false,
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.TaxServicesListResponse{
				Items:      []api.TaxService{},
				TotalCount: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &taxServicesAPIClient{
				listTaxServicesResp: tt.mockResp,
				listTaxServicesErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			if tt.jsonOutput {
				cmd.Flags().String("output", "json", "")
			} else {
				cmd.Flags().String("output", "", "")
			}
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			cmd.Flags().String("provider", "", "")
			cmd.Flags().Bool("active", false, "")

			err := taxServicesListCmd.RunE(cmd, []string{})

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

// TestTaxServicesListRunE_WithActiveFilter tests list with active filter.
func TestTaxServicesListRunE_WithActiveFilter(t *testing.T) {
	cleanup := setupTaxServicesTest(t)
	defer cleanup()

	mockClient := &taxServicesAPIClient{
		listTaxServicesResp: &api.TaxServicesListResponse{
			Items:      []api.TaxService{},
			TotalCount: 0,
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
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("provider", "avalara", "")
	cmd.Flags().Bool("active", false, "")
	_ = cmd.Flags().Set("active", "true")

	err := taxServicesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestTaxServicesGetRunE tests the tax-services get command execution with mock API.
func TestTaxServicesGetRunE(t *testing.T) {
	cleanup := setupTaxServicesTest(t)
	defer cleanup()

	tests := []struct {
		name       string
		id         string
		mockResp   *api.TaxService
		mockErr    error
		jsonOutput bool
		wantErr    bool
	}{
		{
			name: "successful get",
			id:   "svc_123",
			mockResp: &api.TaxService{
				ID:          "svc_123",
				Name:        "Avalara Tax",
				Provider:    "avalara",
				Sandbox:     false,
				Active:      true,
				CallbackURL: "https://example.com/callback",
				Countries:   []string{"US", "CA"},
				CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "successful get without callback URL",
			id:   "svc_456",
			mockResp: &api.TaxService{
				ID:        "svc_456",
				Name:      "TaxJar",
				Provider:  "taxjar",
				Sandbox:   true,
				Active:    false,
				Countries: []string{},
				CreatedAt: time.Date(2024, 2, 1, 8, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 2, 2, 8, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "successful get with JSON output",
			id:   "svc_json",
			mockResp: &api.TaxService{
				ID:       "svc_json",
				Name:     "JSON Test",
				Provider: "taxjar",
				Active:   true,
			},
			jsonOutput: true,
			wantErr:    false,
		},
		{
			name:    "not found",
			id:      "svc_999",
			mockErr: errors.New("tax service not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &taxServicesAPIClient{
				getTaxServiceResp: tt.mockResp,
				getTaxServiceErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			if tt.jsonOutput {
				cmd.Flags().String("output", "json", "")
			} else {
				cmd.Flags().String("output", "", "")
			}
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")

			err := taxServicesGetCmd.RunE(cmd, []string{tt.id})

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

// TestTaxServicesCreateRunE tests the tax-services create command execution with mock API.
func TestTaxServicesCreateRunE(t *testing.T) {
	cleanup := setupTaxServicesTest(t)
	defer cleanup()

	tests := []struct {
		name       string
		mockResp   *api.TaxService
		mockErr    error
		jsonOutput bool
		wantErr    bool
	}{
		{
			name: "successful create",
			mockResp: &api.TaxService{
				ID:       "svc_new",
				Name:     "New Tax Service",
				Provider: "avalara",
				Active:   true,
			},
			wantErr: false,
		},
		{
			name: "successful create with JSON output",
			mockResp: &api.TaxService{
				ID:       "svc_json",
				Name:     "JSON Test",
				Provider: "taxjar",
				Active:   true,
			},
			jsonOutput: true,
			wantErr:    false,
		},
		{
			name:    "create fails",
			mockErr: errors.New("validation error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &taxServicesAPIClient{
				createTaxServiceResp: tt.mockResp,
				createTaxServiceErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			if tt.jsonOutput {
				cmd.Flags().String("output", "json", "")
			} else {
				cmd.Flags().String("output", "", "")
			}
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("name", "New Tax Service", "")
			cmd.Flags().String("provider", "avalara", "")
			cmd.Flags().String("api-key", "test-key", "")
			cmd.Flags().String("api-secret", "test-secret", "")
			cmd.Flags().Bool("sandbox", false, "")
			cmd.Flags().Bool("active", true, "")
			cmd.Flags().String("callback-url", "https://example.com/callback", "")
			cmd.Flags().String("countries", "US, CA", "")
			cmd.Flags().Bool("dry-run", false, "")

			err := taxServicesCreateCmd.RunE(cmd, []string{})

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

// TestTaxServicesCreateRunE_DryRun tests create command with dry-run flag.
func TestTaxServicesCreateDryRun(t *testing.T) {
	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("dry-run", "true")
	cmd.Flags().String("name", "", "Name")
	_ = cmd.Flags().Set("name", "Avalara")
	cmd.Flags().String("provider", "", "Provider")
	_ = cmd.Flags().Set("provider", "avalara")
	cmd.Flags().String("api-key", "", "API key")
	cmd.Flags().String("api-secret", "", "API secret")
	cmd.Flags().Bool("sandbox", false, "Sandbox")
	cmd.Flags().Bool("active", true, "Active")
	cmd.Flags().String("callback-url", "", "Callback URL")
	cmd.Flags().String("countries", "", "Countries")

	err := taxServicesCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Dry run should not return error, got %v", err)
	}
}

// TestTaxServicesCreateRunE_WithCountries tests create command with countries.
func TestTaxServicesCreateRunE_WithCountries(t *testing.T) {
	cleanup := setupTaxServicesTest(t)
	defer cleanup()

	mockClient := &taxServicesAPIClient{
		createTaxServiceResp: &api.TaxService{
			ID:        "svc_new",
			Name:      "Tax Service",
			Provider:  "avalara",
			Countries: []string{"US", "CA", "MX"},
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
	cmd.Flags().String("name", "Tax Service", "")
	cmd.Flags().String("provider", "avalara", "")
	cmd.Flags().String("api-key", "key", "")
	cmd.Flags().String("api-secret", "", "")
	cmd.Flags().Bool("sandbox", false, "")
	cmd.Flags().Bool("active", true, "")
	cmd.Flags().String("callback-url", "", "")
	cmd.Flags().String("countries", "US, CA, MX", "")
	cmd.Flags().Bool("dry-run", false, "")

	err := taxServicesCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestTaxServicesUpdateRunE tests the tax-services update command execution with mock API.
func TestTaxServicesUpdateRunE(t *testing.T) {
	cleanup := setupTaxServicesTest(t)
	defer cleanup()

	tests := []struct {
		name       string
		id         string
		mockResp   *api.TaxService
		mockErr    error
		jsonOutput bool
		wantErr    bool
	}{
		{
			name: "successful update",
			id:   "svc_123",
			mockResp: &api.TaxService{
				ID:     "svc_123",
				Name:   "Updated Tax Service",
				Active: true,
			},
			wantErr: false,
		},
		{
			name: "successful update with JSON output",
			id:   "svc_json",
			mockResp: &api.TaxService{
				ID:     "svc_json",
				Name:   "JSON Test",
				Active: true,
			},
			jsonOutput: true,
			wantErr:    false,
		},
		{
			name:    "update fails",
			id:      "svc_456",
			mockErr: errors.New("validation error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &taxServicesAPIClient{
				updateTaxServiceResp: tt.mockResp,
				updateTaxServiceErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			if tt.jsonOutput {
				cmd.Flags().String("output", "json", "")
			} else {
				cmd.Flags().String("output", "", "")
			}
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("name", "", "")
			cmd.Flags().String("api-key", "", "")
			cmd.Flags().String("api-secret", "", "")
			cmd.Flags().Bool("sandbox", false, "")
			cmd.Flags().Bool("active", false, "")
			cmd.Flags().String("callback-url", "", "")
			cmd.Flags().String("countries", "", "")
			cmd.Flags().Bool("dry-run", false, "")

			err := taxServicesUpdateCmd.RunE(cmd, []string{tt.id})

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

// TestTaxServicesUpdateDryRun tests update command with dry-run flag.
func TestTaxServicesUpdateDryRun(t *testing.T) {
	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("dry-run", "true")
	cmd.Flags().String("name", "", "Name")
	cmd.Flags().String("api-key", "", "API key")
	cmd.Flags().String("api-secret", "", "API secret")
	cmd.Flags().Bool("sandbox", false, "Sandbox")
	cmd.Flags().Bool("active", false, "Active")
	cmd.Flags().String("callback-url", "", "Callback URL")
	cmd.Flags().String("countries", "", "Countries")

	err := taxServicesUpdateCmd.RunE(cmd, []string{"svc_123"})
	if err != nil {
		t.Errorf("Dry run should not return error, got %v", err)
	}
}

// TestTaxServicesUpdateRunE_WithAllFlags tests update with all flags changed.
func TestTaxServicesUpdateRunE_WithAllFlags(t *testing.T) {
	cleanup := setupTaxServicesTest(t)
	defer cleanup()

	mockClient := &taxServicesAPIClient{
		updateTaxServiceResp: &api.TaxService{
			ID:     "svc_123",
			Name:   "Updated Name",
			Active: false,
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
	cmd.Flags().String("name", "", "")
	_ = cmd.Flags().Set("name", "Updated Name")
	cmd.Flags().String("api-key", "", "")
	_ = cmd.Flags().Set("api-key", "new-key")
	cmd.Flags().String("api-secret", "", "")
	_ = cmd.Flags().Set("api-secret", "new-secret")
	cmd.Flags().Bool("sandbox", false, "")
	_ = cmd.Flags().Set("sandbox", "true")
	cmd.Flags().Bool("active", false, "")
	_ = cmd.Flags().Set("active", "false")
	cmd.Flags().String("callback-url", "", "")
	_ = cmd.Flags().Set("callback-url", "https://new.example.com/callback")
	cmd.Flags().String("countries", "", "")
	_ = cmd.Flags().Set("countries", "GB, DE, FR")
	cmd.Flags().Bool("dry-run", false, "")

	err := taxServicesUpdateCmd.RunE(cmd, []string{"svc_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestTaxServicesDeleteRunE tests the tax-services delete command execution with mock API.
func TestTaxServicesDeleteRunE(t *testing.T) {
	cleanup := setupTaxServicesTest(t)
	defer cleanup()

	tests := []struct {
		name    string
		id      string
		mockErr error
		wantErr bool
	}{
		{
			name:    "successful delete",
			id:      "svc_123",
			wantErr: false,
		},
		{
			name:    "delete fails",
			id:      "svc_456",
			mockErr: errors.New("tax service in use"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &taxServicesAPIClient{
				deleteTaxServiceErr: tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().Bool("yes", true, "")
			cmd.Flags().Bool("dry-run", false, "")

			err := taxServicesDeleteCmd.RunE(cmd, []string{tt.id})

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

// TestTaxServicesDeleteDryRun tests delete command with dry-run flag.
func TestTaxServicesDeleteDryRun(t *testing.T) {
	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("dry-run", "true")
	err := taxServicesDeleteCmd.RunE(cmd, []string{"svc_123"})
	if err != nil {
		t.Errorf("Dry run should not return error, got %v", err)
	}
}

// TestTaxServicesDeleteWithoutConfirmation tests delete without confirmation.
func TestTaxServicesDeleteWithoutConfirmation(t *testing.T) {
	cmd := newTestCmdWithFlags()
	err := taxServicesDeleteCmd.RunE(cmd, []string{"svc_123"})
	if err != nil {
		t.Errorf("Delete without confirmation should not return error, got %v", err)
	}
}

// TestTaxServicesListRunE_GetClientFails tests error handling when getClient fails.
func TestTaxServicesListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("provider", "", "")
	cmd.Flags().Bool("active", false, "")

	err := taxServicesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestTaxServicesGetRunE_GetClientFails tests error handling when getClient fails.
func TestTaxServicesGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()

	err := taxServicesGetCmd.RunE(cmd, []string{"svc_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestTaxServicesCreateRunE_GetClientFails tests error handling when getClient fails.
func TestTaxServicesCreateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "Test", "")
	cmd.Flags().String("provider", "avalara", "")
	cmd.Flags().String("api-key", "key", "")
	cmd.Flags().String("api-secret", "", "")
	cmd.Flags().Bool("sandbox", false, "")
	cmd.Flags().Bool("active", true, "")
	cmd.Flags().String("callback-url", "", "")
	cmd.Flags().String("countries", "", "")

	err := taxServicesCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestTaxServicesUpdateRunE_GetClientFails tests error handling when getClient fails.
func TestTaxServicesUpdateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("api-key", "", "")
	cmd.Flags().String("api-secret", "", "")
	cmd.Flags().Bool("sandbox", false, "")
	cmd.Flags().Bool("active", false, "")
	cmd.Flags().String("callback-url", "", "")
	cmd.Flags().String("countries", "", "")

	err := taxServicesUpdateCmd.RunE(cmd, []string{"svc_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestTaxServicesDeleteRunE_GetClientFails tests error handling when getClient fails.
func TestTaxServicesDeleteRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")

	err := taxServicesDeleteCmd.RunE(cmd, []string{"svc_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}
