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

func TestTaxesCommand(t *testing.T) {
	if taxesCmd == nil {
		t.Fatal("taxesCmd is nil")
	}
	if taxesCmd.Use != "taxes" {
		t.Errorf("Expected Use to be 'taxes', got %q", taxesCmd.Use)
	}
}

func TestTaxesSubcommands(t *testing.T) {
	subcommands := taxesCmd.Commands()
	expectedCmds := map[string]bool{"list": false, "get": false, "create": false, "update": false, "delete": false}
	for _, cmd := range subcommands {
		switch cmd.Use {
		case "list":
			expectedCmds["list"] = true
		case "get <id>":
			expectedCmds["get"] = true
		case "create":
			expectedCmds["create"] = true
		case "update <id>":
			expectedCmds["update"] = true
		case "delete <id>":
			expectedCmds["delete"] = true
		}
	}
	for name, found := range expectedCmds {
		if !found {
			t.Errorf("Expected subcommand %q not found", name)
		}
	}
}

func TestTaxesListFlags(t *testing.T) {
	flags := []string{"page", "page-size", "country", "enabled"}
	for _, flag := range flags {
		if taxesListCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag --%s not found on list command", flag)
		}
	}
}

func TestTaxesCreateFlags(t *testing.T) {
	flags := []string{"name", "rate", "country", "province", "priority", "compound", "shipping", "enabled"}
	for _, flag := range flags {
		if taxesCreateCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag --%s not found on create command", flag)
		}
	}
}

func TestTaxesUpdateFlags(t *testing.T) {
	flags := []string{"name", "rate", "priority", "compound", "shipping", "enabled"}
	for _, flag := range flags {
		if taxesUpdateCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag --%s not found on update command", flag)
		}
	}
}

func TestTaxesGetArgsValidation(t *testing.T) {
	if taxesGetCmd.Args == nil {
		t.Fatal("Expected Args validator on get command")
	}
	err := taxesGetCmd.Args(taxesGetCmd, []string{})
	if err == nil {
		t.Error("Expected error with no args")
	}
	err = taxesGetCmd.Args(taxesGetCmd, []string{"tax_123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}
}

func TestTaxesUpdateArgsValidation(t *testing.T) {
	if taxesUpdateCmd.Args == nil {
		t.Fatal("Expected Args validator on update command")
	}
	err := taxesUpdateCmd.Args(taxesUpdateCmd, []string{})
	if err == nil {
		t.Error("Expected error with no args")
	}
	err = taxesUpdateCmd.Args(taxesUpdateCmd, []string{"tax_123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}
}

func TestTaxesDeleteArgsValidation(t *testing.T) {
	if taxesDeleteCmd.Args == nil {
		t.Fatal("Expected Args validator on delete command")
	}
	err := taxesDeleteCmd.Args(taxesDeleteCmd, []string{})
	if err == nil {
		t.Error("Expected error with no args")
	}
	err = taxesDeleteCmd.Args(taxesDeleteCmd, []string{"tax_123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}
}

func TestTaxesDeleteWithoutConfirmation(t *testing.T) {
	cmd := newTestCmdWithFlags()
	err := taxesDeleteCmd.RunE(cmd, []string{"tax_123"})
	if err != nil {
		t.Errorf("Delete without confirmation should not return error, got %v", err)
	}
}

// taxesMockAPIClient is a mock implementation of api.APIClient for taxes tests.
type taxesMockAPIClient struct {
	api.MockClient
	listTaxesResp *api.TaxesListResponse
	listTaxesErr  error
	getTaxResp    *api.Tax
	getTaxErr     error
	createTaxResp *api.Tax
	createTaxErr  error
	updateTaxResp *api.Tax
	updateTaxErr  error
	deleteTaxErr  error
}

func (m *taxesMockAPIClient) ListTaxes(ctx context.Context, opts *api.TaxesListOptions) (*api.TaxesListResponse, error) {
	return m.listTaxesResp, m.listTaxesErr
}

func (m *taxesMockAPIClient) GetTax(ctx context.Context, id string) (*api.Tax, error) {
	return m.getTaxResp, m.getTaxErr
}

func (m *taxesMockAPIClient) CreateTax(ctx context.Context, req *api.TaxCreateRequest) (*api.Tax, error) {
	return m.createTaxResp, m.createTaxErr
}

func (m *taxesMockAPIClient) UpdateTax(ctx context.Context, id string, req *api.TaxUpdateRequest) (*api.Tax, error) {
	return m.updateTaxResp, m.updateTaxErr
}

func (m *taxesMockAPIClient) DeleteTax(ctx context.Context, id string) error {
	return m.deleteTaxErr
}

// setupTaxesMockFactories sets up mock factories for taxes tests.
func setupTaxesMockFactories(mockClient *taxesMockAPIClient) (func(), *bytes.Buffer) {
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

// newTaxesTestCmd creates a test command with common flags for taxes tests.
func newTaxesTestCmd() *cobra.Command {
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

// TestTaxesListRunE tests the taxes list command with mock API.
func TestTaxesListRunE(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	tests := []struct {
		name       string
		mockResp   *api.TaxesListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.TaxesListResponse{
				Items: []api.Tax{
					{
						ID:           "tax_123",
						Name:         "Sales Tax",
						Rate:         8.25,
						CountryCode:  "US",
						ProvinceCode: "CA",
						Priority:     1,
						Compound:     false,
						Shipping:     true,
						Enabled:      true,
						CreatedAt:    fixedTime,
						UpdatedAt:    fixedTime,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "tax_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.TaxesListResponse{
				Items:      []api.Tax{},
				TotalCount: 0,
			},
		},
		{
			name: "multiple taxes with different states",
			mockResp: &api.TaxesListResponse{
				Items: []api.Tax{
					{
						ID:           "tax_001",
						Name:         "VAT",
						Rate:         20.0,
						CountryCode:  "GB",
						ProvinceCode: "",
						Shipping:     true,
						Enabled:      true,
						CreatedAt:    fixedTime,
						UpdatedAt:    fixedTime,
					},
					{
						ID:           "tax_002",
						Name:         "Disabled Tax",
						Rate:         5.0,
						CountryCode:  "US",
						ProvinceCode: "TX",
						Shipping:     false,
						Enabled:      false,
						CreatedAt:    fixedTime,
						UpdatedAt:    fixedTime,
					},
				},
				TotalCount: 2,
			},
			wantOutput: "tax_001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &taxesMockAPIClient{
				listTaxesResp: tt.mockResp,
				listTaxesErr:  tt.mockErr,
			}
			cleanup, buf := setupTaxesMockFactories(mockClient)
			defer cleanup()

			cmd := newTaxesTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			cmd.Flags().String("country", "", "")
			cmd.Flags().Bool("enabled", false, "")

			err := taxesListCmd.RunE(cmd, []string{})

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

			output := buf.String()
			if tt.wantOutput != "" && !strings.Contains(output, tt.wantOutput) {
				t.Errorf("output %q should contain %q", output, tt.wantOutput)
			}
		})
	}
}

// TestTaxesListJSONOutput tests the taxes list command with JSON output.
func TestTaxesListJSONOutput(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mockClient := &taxesMockAPIClient{
		listTaxesResp: &api.TaxesListResponse{
			Items: []api.Tax{
				{
					ID:          "tax_json",
					Name:        "JSON Tax",
					Rate:        15.0,
					CountryCode: "AU",
					Enabled:     true,
					CreatedAt:   fixedTime,
					UpdatedAt:   fixedTime,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupTaxesMockFactories(mockClient)
	defer cleanup()

	cmd := newTaxesTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("country", "", "")
	cmd.Flags().Bool("enabled", false, "")
	_ = cmd.Flags().Set("output", "json")

	err := taxesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "tax_json") {
		t.Errorf("JSON output should contain tax ID, got %q", output)
	}
}

// TestTaxesListWithFilters tests list with country and enabled filters.
func TestTaxesListWithFilters(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mockClient := &taxesMockAPIClient{
		listTaxesResp: &api.TaxesListResponse{
			Items: []api.Tax{
				{
					ID:          "tax_filtered",
					Name:        "US Tax",
					Rate:        10.0,
					CountryCode: "US",
					Enabled:     true,
					CreatedAt:   fixedTime,
					UpdatedAt:   fixedTime,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupTaxesMockFactories(mockClient)
	defer cleanup()

	cmd := newTaxesTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("country", "", "")
	cmd.Flags().Bool("enabled", false, "")
	_ = cmd.Flags().Set("country", "US")
	_ = cmd.Flags().Set("enabled", "true")

	err := taxesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "tax_filtered") {
		t.Errorf("output should contain filtered tax, got %q", output)
	}
}

// TestTaxesGetRunE tests the taxes get command with mock API.
func TestTaxesGetRunE(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	tests := []struct {
		name     string
		taxID    string
		mockResp *api.Tax
		mockErr  error
		wantErr  bool
	}{
		{
			name:  "successful get",
			taxID: "tax_123",
			mockResp: &api.Tax{
				ID:           "tax_123",
				Name:         "Sales Tax",
				Rate:         8.25,
				CountryCode:  "US",
				ProvinceCode: "CA",
				Priority:     1,
				Compound:     true,
				Shipping:     true,
				Enabled:      true,
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
			},
		},
		{
			name:    "tax not found",
			taxID:   "tax_999",
			mockErr: errors.New("tax not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &taxesMockAPIClient{
				getTaxResp: tt.mockResp,
				getTaxErr:  tt.mockErr,
			}
			cleanup, _ := setupTaxesMockFactories(mockClient)
			defer cleanup()

			cmd := newTaxesTestCmd()

			err := taxesGetCmd.RunE(cmd, []string{tt.taxID})

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

// TestTaxesGetJSONOutput tests the taxes get command with JSON output.
func TestTaxesGetJSONOutput(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mockClient := &taxesMockAPIClient{
		getTaxResp: &api.Tax{
			ID:          "tax_json_get",
			Name:        "JSON Get Tax",
			Rate:        12.5,
			CountryCode: "DE",
			Enabled:     true,
			CreatedAt:   fixedTime,
			UpdatedAt:   fixedTime,
		},
	}
	cleanup, buf := setupTaxesMockFactories(mockClient)
	defer cleanup()

	cmd := newTaxesTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := taxesGetCmd.RunE(cmd, []string{"tax_json_get"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "tax_json_get") {
		t.Errorf("JSON output should contain tax ID, got %q", output)
	}
}

// TestTaxesGetWithProvinceCode tests get output when province code is present.
func TestTaxesGetWithProvinceCode(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mockClient := &taxesMockAPIClient{
		getTaxResp: &api.Tax{
			ID:           "tax_province",
			Name:         "Provincial Tax",
			Rate:         7.0,
			CountryCode:  "CA",
			ProvinceCode: "BC",
			Priority:     2,
			Compound:     false,
			Shipping:     true,
			Enabled:      true,
			CreatedAt:    fixedTime,
			UpdatedAt:    fixedTime,
		},
	}
	cleanup, _ := setupTaxesMockFactories(mockClient)
	defer cleanup()

	cmd := newTaxesTestCmd()

	err := taxesGetCmd.RunE(cmd, []string{"tax_province"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestTaxesGetWithoutProvinceCode tests get output when province code is empty.
func TestTaxesGetWithoutProvinceCode(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mockClient := &taxesMockAPIClient{
		getTaxResp: &api.Tax{
			ID:           "tax_no_province",
			Name:         "National Tax",
			Rate:         5.0,
			CountryCode:  "FR",
			ProvinceCode: "",
			Priority:     1,
			Compound:     false,
			Shipping:     false,
			Enabled:      true,
			CreatedAt:    fixedTime,
			UpdatedAt:    fixedTime,
		},
	}
	cleanup, _ := setupTaxesMockFactories(mockClient)
	defer cleanup()

	cmd := newTaxesTestCmd()

	err := taxesGetCmd.RunE(cmd, []string{"tax_no_province"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestTaxesCreateDryRun verifies dry-run mode for create command.
func TestTaxesCreateDryRun(t *testing.T) {
	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("dry-run", "true")
	cmd.Flags().String("name", "", "Name")
	_ = cmd.Flags().Set("name", "Sales Tax")
	cmd.Flags().Float64("rate", 0, "Rate")
	_ = cmd.Flags().Set("rate", "8.5")
	cmd.Flags().String("country", "", "Country")
	_ = cmd.Flags().Set("country", "US")
	cmd.Flags().String("province", "", "Province")
	cmd.Flags().Int("priority", 1, "Priority")
	cmd.Flags().Bool("compound", false, "Compound")
	cmd.Flags().Bool("shipping", false, "Shipping")
	cmd.Flags().Bool("enabled", true, "Enabled")
	err := taxesCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Dry run should not return error, got %v", err)
	}
}

// TestTaxesCreateRunE tests the taxes create command with mock API.
func TestTaxesCreateRunE(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	tests := []struct {
		name     string
		mockResp *api.Tax
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			mockResp: &api.Tax{
				ID:          "tax_new",
				Name:        "New Tax",
				Rate:        10.0,
				CountryCode: "US",
				Enabled:     true,
				CreatedAt:   fixedTime,
				UpdatedAt:   fixedTime,
			},
		},
		{
			name:    "API error",
			mockErr: errors.New("failed to create tax"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &taxesMockAPIClient{
				createTaxResp: tt.mockResp,
				createTaxErr:  tt.mockErr,
			}
			cleanup, _ := setupTaxesMockFactories(mockClient)
			defer cleanup()

			cmd := newTaxesTestCmd()
			cmd.Flags().String("name", "", "")
			_ = cmd.Flags().Set("name", "New Tax")
			cmd.Flags().Float64("rate", 0, "")
			_ = cmd.Flags().Set("rate", "10.0")
			cmd.Flags().String("country", "", "")
			_ = cmd.Flags().Set("country", "US")
			cmd.Flags().String("province", "", "")
			cmd.Flags().Int("priority", 1, "")
			cmd.Flags().Bool("compound", false, "")
			cmd.Flags().Bool("shipping", false, "")
			cmd.Flags().Bool("enabled", true, "")

			err := taxesCreateCmd.RunE(cmd, []string{})

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

// TestTaxesCreateJSONOutput tests the taxes create command with JSON output.
func TestTaxesCreateJSONOutput(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mockClient := &taxesMockAPIClient{
		createTaxResp: &api.Tax{
			ID:          "tax_json_create",
			Name:        "JSON Create Tax",
			Rate:        8.0,
			CountryCode: "JP",
			Enabled:     true,
			CreatedAt:   fixedTime,
			UpdatedAt:   fixedTime,
		},
	}
	cleanup, buf := setupTaxesMockFactories(mockClient)
	defer cleanup()

	cmd := newTaxesTestCmd()
	cmd.Flags().String("name", "", "")
	_ = cmd.Flags().Set("name", "JSON Create Tax")
	cmd.Flags().Float64("rate", 0, "")
	_ = cmd.Flags().Set("rate", "8.0")
	cmd.Flags().String("country", "", "")
	_ = cmd.Flags().Set("country", "JP")
	cmd.Flags().String("province", "", "")
	cmd.Flags().Int("priority", 1, "")
	cmd.Flags().Bool("compound", false, "")
	cmd.Flags().Bool("shipping", false, "")
	cmd.Flags().Bool("enabled", true, "")
	_ = cmd.Flags().Set("output", "json")

	err := taxesCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "tax_json_create") {
		t.Errorf("JSON output should contain tax ID, got %q", output)
	}
}

// TestTaxesCreateGetClientError tests create command when getClient fails.
func TestTaxesCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTaxesTestCmd()
	cmd.Flags().String("name", "", "")
	_ = cmd.Flags().Set("name", "Tax")
	cmd.Flags().Float64("rate", 0, "")
	_ = cmd.Flags().Set("rate", "5.0")
	cmd.Flags().String("country", "", "")
	_ = cmd.Flags().Set("country", "US")
	cmd.Flags().String("province", "", "")
	cmd.Flags().Int("priority", 1, "")
	cmd.Flags().Bool("compound", false, "")
	cmd.Flags().Bool("shipping", false, "")
	cmd.Flags().Bool("enabled", true, "")

	err := taxesCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestTaxesUpdateDryRun verifies dry-run mode for update command.
func TestTaxesUpdateDryRun(t *testing.T) {
	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("dry-run", "true")
	cmd.Flags().String("name", "", "Name")
	cmd.Flags().Float64("rate", 0, "Rate")
	cmd.Flags().Int("priority", 0, "Priority")
	cmd.Flags().Bool("compound", false, "Compound")
	cmd.Flags().Bool("shipping", false, "Shipping")
	cmd.Flags().Bool("enabled", false, "Enabled")
	err := taxesUpdateCmd.RunE(cmd, []string{"tax_123"})
	if err != nil {
		t.Errorf("Dry run should not return error, got %v", err)
	}
}

// TestTaxesUpdateRunE tests the taxes update command with mock API.
func TestTaxesUpdateRunE(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	tests := []struct {
		name     string
		taxID    string
		mockResp *api.Tax
		mockErr  error
		wantErr  bool
	}{
		{
			name:  "successful update",
			taxID: "tax_123",
			mockResp: &api.Tax{
				ID:          "tax_123",
				Name:        "Updated Tax",
				Rate:        12.0,
				CountryCode: "US",
				Enabled:     true,
				CreatedAt:   fixedTime,
				UpdatedAt:   fixedTime,
			},
		},
		{
			name:    "tax not found",
			taxID:   "tax_999",
			mockErr: errors.New("tax not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &taxesMockAPIClient{
				updateTaxResp: tt.mockResp,
				updateTaxErr:  tt.mockErr,
			}
			cleanup, _ := setupTaxesMockFactories(mockClient)
			defer cleanup()

			cmd := newTaxesTestCmd()
			cmd.Flags().String("name", "", "")
			_ = cmd.Flags().Set("name", "Updated Tax")
			cmd.Flags().Float64("rate", 0, "")
			cmd.Flags().Int("priority", 0, "")
			cmd.Flags().Bool("compound", false, "")
			cmd.Flags().Bool("shipping", false, "")
			cmd.Flags().Bool("enabled", false, "")

			err := taxesUpdateCmd.RunE(cmd, []string{tt.taxID})

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

// TestTaxesUpdateJSONOutput tests the taxes update command with JSON output.
func TestTaxesUpdateJSONOutput(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mockClient := &taxesMockAPIClient{
		updateTaxResp: &api.Tax{
			ID:          "tax_json_update",
			Name:        "JSON Updated Tax",
			Rate:        15.0,
			CountryCode: "NZ",
			Enabled:     true,
			CreatedAt:   fixedTime,
			UpdatedAt:   fixedTime,
		},
	}
	cleanup, buf := setupTaxesMockFactories(mockClient)
	defer cleanup()

	cmd := newTaxesTestCmd()
	cmd.Flags().String("name", "", "")
	_ = cmd.Flags().Set("name", "JSON Updated Tax")
	cmd.Flags().Float64("rate", 0, "")
	_ = cmd.Flags().Set("rate", "15.0")
	cmd.Flags().Int("priority", 0, "")
	cmd.Flags().Bool("compound", false, "")
	cmd.Flags().Bool("shipping", false, "")
	cmd.Flags().Bool("enabled", false, "")
	_ = cmd.Flags().Set("output", "json")

	err := taxesUpdateCmd.RunE(cmd, []string{"tax_json_update"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "tax_json_update") {
		t.Errorf("JSON output should contain tax ID, got %q", output)
	}
}

// TestTaxesUpdateWithAllFlags tests update command with all optional flags set.
func TestTaxesUpdateWithAllFlags(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mockClient := &taxesMockAPIClient{
		updateTaxResp: &api.Tax{
			ID:          "tax_all_flags",
			Name:        "All Flags Tax",
			Rate:        9.5,
			CountryCode: "US",
			Priority:    2,
			Compound:    true,
			Shipping:    true,
			Enabled:     true,
			CreatedAt:   fixedTime,
			UpdatedAt:   fixedTime,
		},
	}
	cleanup, _ := setupTaxesMockFactories(mockClient)
	defer cleanup()

	cmd := newTaxesTestCmd()
	cmd.Flags().String("name", "", "")
	_ = cmd.Flags().Set("name", "All Flags Tax")
	cmd.Flags().Float64("rate", 0, "")
	_ = cmd.Flags().Set("rate", "9.5")
	cmd.Flags().Int("priority", 0, "")
	_ = cmd.Flags().Set("priority", "2")
	cmd.Flags().Bool("compound", false, "")
	_ = cmd.Flags().Set("compound", "true")
	cmd.Flags().Bool("shipping", false, "")
	_ = cmd.Flags().Set("shipping", "true")
	cmd.Flags().Bool("enabled", false, "")
	_ = cmd.Flags().Set("enabled", "true")

	err := taxesUpdateCmd.RunE(cmd, []string{"tax_all_flags"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestTaxesUpdateGetClientError tests update command when getClient fails.
func TestTaxesUpdateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTaxesTestCmd()
	cmd.Flags().String("name", "", "")
	cmd.Flags().Float64("rate", 0, "")
	cmd.Flags().Int("priority", 0, "")
	cmd.Flags().Bool("compound", false, "")
	cmd.Flags().Bool("shipping", false, "")
	cmd.Flags().Bool("enabled", false, "")

	err := taxesUpdateCmd.RunE(cmd, []string{"tax_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestTaxesDeleteDryRun verifies dry-run mode for delete command.
func TestTaxesDeleteDryRun(t *testing.T) {
	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("dry-run", "true")
	err := taxesDeleteCmd.RunE(cmd, []string{"tax_123"})
	if err != nil {
		t.Errorf("Dry run should not return error, got %v", err)
	}
}

// TestTaxesDeleteRunE tests the taxes delete command with mock API.
func TestTaxesDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		taxID   string
		mockErr error
		wantErr bool
	}{
		{
			name:  "successful delete",
			taxID: "tax_123",
		},
		{
			name:    "tax not found",
			taxID:   "tax_999",
			mockErr: errors.New("tax not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &taxesMockAPIClient{
				deleteTaxErr: tt.mockErr,
			}
			cleanup, _ := setupTaxesMockFactories(mockClient)
			defer cleanup()

			cmd := newTaxesTestCmd()
			_ = cmd.Flags().Set("yes", "true")

			err := taxesDeleteCmd.RunE(cmd, []string{tt.taxID})

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

// TestTaxesDeleteGetClientError tests delete command when getClient fails.
func TestTaxesDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTaxesTestCmd()
	_ = cmd.Flags().Set("yes", "true")

	err := taxesDeleteCmd.RunE(cmd, []string{"tax_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestTaxesListGetClientError tests list command when getClient fails.
func TestTaxesListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTaxesTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("country", "", "")
	cmd.Flags().Bool("enabled", false, "")

	err := taxesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestTaxesGetGetClientError tests get command when getClient fails.
func TestTaxesGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTaxesTestCmd()

	err := taxesGetCmd.RunE(cmd, []string{"tax_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestTaxesWithMockStore tests taxes commands with a mock credential store.
func TestTaxesWithMockStore(t *testing.T) {
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

	cmd := newTaxesTestCmd()
	client, err := getClient(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Error("expected client, got nil")
	}
}

// TestTaxesListFlagsWithDefaults tests list command flag defaults.
func TestTaxesListFlagsWithDefaults(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
		{"country", ""},
		{"enabled", "false"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := taxesListCmd.Flags().Lookup(f.name)
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

// TestTaxesCreateFlagsWithDefaults tests create command flag defaults.
func TestTaxesCreateFlagsWithDefaults(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"name", ""},
		{"rate", "0"},
		{"country", ""},
		{"province", ""},
		{"priority", "1"},
		{"compound", "false"},
		{"shipping", "false"},
		{"enabled", "true"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := taxesCreateCmd.Flags().Lookup(f.name)
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

// TestTaxesDeleteYesFlag tests delete command --yes flag.
func TestTaxesDeleteYesFlag(t *testing.T) {
	flag := taxesDeleteCmd.Flags().Lookup("yes")
	if flag == nil {
		t.Error("--yes flag not found on delete command")
		return
	}
	if flag.DefValue != "false" {
		t.Errorf("expected default 'false', got %q", flag.DefValue)
	}
}
