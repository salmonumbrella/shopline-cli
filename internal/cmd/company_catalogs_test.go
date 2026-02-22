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

func TestCompanyCatalogsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := companyCatalogsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCompanyCatalogsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := companyCatalogsGetCmd.RunE(cmd, []string{"catalog-123"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCompanyCatalogsCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("company-id", "comp-123", "")
	cmd.Flags().String("name", "Test Catalog", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().StringSlice("product-ids", nil, "")
	cmd.Flags().Bool("default", false, "")

	err := companyCatalogsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCompanyCatalogsUpdateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().StringSlice("product-ids", nil, "")
	cmd.Flags().Bool("default", false, "")

	err := companyCatalogsUpdateCmd.RunE(cmd, []string{"catalog-123"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCompanyCatalogsDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	// yes flag already added by newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")

	err := companyCatalogsDeleteCmd.RunE(cmd, []string{"catalog-123"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCompanyCatalogsListFlags(t *testing.T) {
	flags := companyCatalogsListCmd.Flags()

	if flags.Lookup("page") == nil {
		t.Error("Expected page flag")
	}
	if flags.Lookup("page-size") == nil {
		t.Error("Expected page-size flag")
	}
	if flags.Lookup("company-id") == nil {
		t.Error("Expected company-id flag")
	}
	if flags.Lookup("status") == nil {
		t.Error("Expected status flag")
	}
}

func TestCompanyCatalogsCommandStructure(t *testing.T) {
	if companyCatalogsCmd.Use != "company-catalogs" {
		t.Errorf("Expected Use 'company-catalogs', got %s", companyCatalogsCmd.Use)
	}

	subcommands := companyCatalogsCmd.Commands()
	expectedCmds := map[string]bool{
		"list":   false,
		"get":    false,
		"create": false,
		"update": false,
		"delete": false,
	}

	for _, cmd := range subcommands {
		if _, ok := expectedCmds[cmd.Use]; ok || startsWithUse(cmd.Use, expectedCmds) {
			expectedCmds[getBaseUse(cmd.Use)] = true
		}
	}

	for name, found := range expectedCmds {
		if !found {
			t.Errorf("Expected subcommand %s not found", name)
		}
	}
}

func startsWithUse(use string, expected map[string]bool) bool {
	for name := range expected {
		if len(use) >= len(name) && use[:len(name)] == name {
			return true
		}
	}
	return false
}

func getBaseUse(use string) string {
	for i, c := range use {
		if c == ' ' {
			return use[:i]
		}
	}
	return use
}

// companyCatalogsTestClient is a mock implementation for company catalogs testing.
type companyCatalogsTestClient struct {
	api.MockClient

	listCompanyCatalogsResp  *api.CompanyCatalogsListResponse
	listCompanyCatalogsErr   error
	getCompanyCatalogResp    *api.CompanyCatalog
	getCompanyCatalogErr     error
	createCompanyCatalogResp *api.CompanyCatalog
	createCompanyCatalogErr  error
	updateCompanyCatalogResp *api.CompanyCatalog
	updateCompanyCatalogErr  error
	deleteCompanyCatalogErr  error
}

func (m *companyCatalogsTestClient) ListCompanyCatalogs(ctx context.Context, opts *api.CompanyCatalogsListOptions) (*api.CompanyCatalogsListResponse, error) {
	return m.listCompanyCatalogsResp, m.listCompanyCatalogsErr
}

func (m *companyCatalogsTestClient) GetCompanyCatalog(ctx context.Context, id string) (*api.CompanyCatalog, error) {
	return m.getCompanyCatalogResp, m.getCompanyCatalogErr
}

func (m *companyCatalogsTestClient) CreateCompanyCatalog(ctx context.Context, req *api.CompanyCatalogCreateRequest) (*api.CompanyCatalog, error) {
	return m.createCompanyCatalogResp, m.createCompanyCatalogErr
}

func (m *companyCatalogsTestClient) UpdateCompanyCatalog(ctx context.Context, id string, req *api.CompanyCatalogUpdateRequest) (*api.CompanyCatalog, error) {
	return m.updateCompanyCatalogResp, m.updateCompanyCatalogErr
}

func (m *companyCatalogsTestClient) DeleteCompanyCatalog(ctx context.Context, id string) error {
	return m.deleteCompanyCatalogErr
}

// setupCompanyCatalogsMockFactories sets up mock factories for company catalogs tests
// and returns a cleanup function.
func setupCompanyCatalogsMockFactories(mockClient *companyCatalogsTestClient) func() {
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

	return func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}
}

// TestCompanyCatalogsListRunE tests the company catalogs list command execution with mock API.
func TestCompanyCatalogsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.CompanyCatalogsListResponse
		mockErr    error
		jsonOutput bool
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list with items",
			mockResp: &api.CompanyCatalogsListResponse{
				Items: []api.CompanyCatalog{
					{
						ID:          "cat_123",
						CompanyID:   "comp_456",
						CompanyName: "Acme Corp",
						Name:        "Premium Products",
						Description: "Premium catalog for Acme",
						ProductIDs:  []string{"prod_1", "prod_2"},
						IsDefault:   true,
						Status:      "active",
						CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
						UpdatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
					{
						ID:          "cat_456",
						CompanyID:   "comp_789",
						CompanyName: "Beta Inc",
						Name:        "Standard Products",
						Description: "Standard catalog",
						ProductIDs:  []string{},
						IsDefault:   false,
						Status:      "inactive",
						CreatedAt:   time.Time{}, // Zero time to test "-" display
						UpdatedAt:   time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 2,
			},
			wantOutput: "cat_123",
		},
		{
			name: "successful list JSON output",
			mockResp: &api.CompanyCatalogsListResponse{
				Items: []api.CompanyCatalog{
					{
						ID:          "cat_123",
						CompanyID:   "comp_456",
						CompanyName: "Acme Corp",
						Name:        "Premium Products",
						ProductIDs:  []string{"prod_1"},
						IsDefault:   true,
						Status:      "active",
						CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
						UpdatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			jsonOutput: true,
			wantOutput: "cat_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.CompanyCatalogsListResponse{
				Items:      []api.CompanyCatalog{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &companyCatalogsTestClient{
				listCompanyCatalogsResp: tt.mockResp,
				listCompanyCatalogsErr:  tt.mockErr,
			}
			cleanup := setupCompanyCatalogsMockFactories(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("company-id", "", "")
			cmd.Flags().String("status", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			if tt.jsonOutput {
				cmd.Flags().String("output", "json", "")
			} else {
				cmd.Flags().String("output", "", "")
			}

			err := companyCatalogsListCmd.RunE(cmd, []string{})

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

// TestCompanyCatalogsGetRunE tests the company catalogs get command execution with mock API.
func TestCompanyCatalogsGetRunE(t *testing.T) {
	tests := []struct {
		name       string
		catalogID  string
		mockResp   *api.CompanyCatalog
		mockErr    error
		jsonOutput bool
		wantErr    bool
	}{
		{
			name:      "successful get",
			catalogID: "cat_123",
			mockResp: &api.CompanyCatalog{
				ID:          "cat_123",
				CompanyID:   "comp_456",
				CompanyName: "Acme Corp",
				Name:        "Premium Products",
				Description: "Premium catalog for Acme",
				ProductIDs:  []string{"prod_1", "prod_2", "prod_3"},
				IsDefault:   true,
				Status:      "active",
				CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:      "successful get with many products (no product IDs displayed)",
			catalogID: "cat_789",
			mockResp: &api.CompanyCatalog{
				ID:          "cat_789",
				CompanyID:   "comp_456",
				CompanyName: "Acme Corp",
				Name:        "Large Catalog",
				Description: "Catalog with many products",
				ProductIDs:  []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8", "p9", "p10", "p11"},
				IsDefault:   false,
				Status:      "active",
				CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:      "successful get JSON output",
			catalogID: "cat_123",
			mockResp: &api.CompanyCatalog{
				ID:          "cat_123",
				CompanyID:   "comp_456",
				CompanyName: "Acme Corp",
				Name:        "Premium Products",
				ProductIDs:  []string{"prod_1"},
				IsDefault:   true,
				Status:      "active",
				CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			jsonOutput: true,
		},
		{
			name:      "catalog not found",
			catalogID: "cat_999",
			mockErr:   errors.New("company catalog not found"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &companyCatalogsTestClient{
				getCompanyCatalogResp: tt.mockResp,
				getCompanyCatalogErr:  tt.mockErr,
			}
			cleanup := setupCompanyCatalogsMockFactories(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			if tt.jsonOutput {
				cmd.Flags().String("output", "json", "")
			} else {
				cmd.Flags().String("output", "", "")
			}

			err := companyCatalogsGetCmd.RunE(cmd, []string{tt.catalogID})

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

// TestCompanyCatalogsCreateRunE tests the company catalogs create command execution with mock API.
func TestCompanyCatalogsCreateRunE(t *testing.T) {
	tests := []struct {
		name        string
		companyID   string
		catalogName string
		productIDs  []string
		isDefault   bool
		mockResp    *api.CompanyCatalog
		mockErr     error
		jsonOutput  bool
		wantErr     bool
	}{
		{
			name:        "successful create",
			companyID:   "comp_123",
			catalogName: "New Catalog",
			productIDs:  []string{"prod_1", "prod_2"},
			isDefault:   false,
			mockResp: &api.CompanyCatalog{
				ID:          "cat_new",
				CompanyID:   "comp_123",
				CompanyName: "Test Company",
				Name:        "New Catalog",
				Description: "A test catalog",
				ProductIDs:  []string{"prod_1", "prod_2"},
				IsDefault:   false,
				Status:      "active",
				CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:        "successful create as default",
			companyID:   "comp_123",
			catalogName: "Default Catalog",
			isDefault:   true,
			mockResp: &api.CompanyCatalog{
				ID:        "cat_default",
				CompanyID: "comp_123",
				Name:      "Default Catalog",
				IsDefault: true,
				Status:    "active",
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:        "successful create JSON output",
			companyID:   "comp_123",
			catalogName: "JSON Catalog",
			mockResp: &api.CompanyCatalog{
				ID:        "cat_json",
				CompanyID: "comp_123",
				Name:      "JSON Catalog",
				Status:    "active",
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			jsonOutput: true,
		},
		{
			name:        "create fails",
			companyID:   "comp_123",
			catalogName: "Failed Catalog",
			mockErr:     errors.New("failed to create company catalog"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &companyCatalogsTestClient{
				createCompanyCatalogResp: tt.mockResp,
				createCompanyCatalogErr:  tt.mockErr,
			}
			cleanup := setupCompanyCatalogsMockFactories(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("company-id", tt.companyID, "")
			cmd.Flags().String("name", tt.catalogName, "")
			cmd.Flags().String("description", "A test catalog", "")
			cmd.Flags().StringSlice("product-ids", tt.productIDs, "")
			cmd.Flags().Bool("default", tt.isDefault, "")
			if tt.jsonOutput {
				cmd.Flags().String("output", "json", "")
			} else {
				cmd.Flags().String("output", "", "")
			}

			err := companyCatalogsCreateCmd.RunE(cmd, []string{})

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

// TestCompanyCatalogsUpdateRunE tests the company catalogs update command execution with mock API.
func TestCompanyCatalogsUpdateRunE(t *testing.T) {
	tests := []struct {
		name        string
		catalogID   string
		updateName  bool
		updateDesc  bool
		updateProds bool
		updateDef   bool
		mockResp    *api.CompanyCatalog
		mockErr     error
		jsonOutput  bool
		wantErr     bool
	}{
		{
			name:       "successful update name",
			catalogID:  "cat_123",
			updateName: true,
			mockResp: &api.CompanyCatalog{
				ID:          "cat_123",
				CompanyID:   "comp_456",
				CompanyName: "Acme Corp",
				Name:        "Updated Catalog",
				Status:      "active",
				CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			name:       "successful update description",
			catalogID:  "cat_123",
			updateDesc: true,
			mockResp: &api.CompanyCatalog{
				ID:          "cat_123",
				CompanyID:   "comp_456",
				Description: "Updated description",
				Status:      "active",
				CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			name:        "successful update product IDs",
			catalogID:   "cat_123",
			updateProds: true,
			mockResp: &api.CompanyCatalog{
				ID:         "cat_123",
				CompanyID:  "comp_456",
				ProductIDs: []string{"prod_new_1", "prod_new_2"},
				Status:     "active",
				CreatedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:  time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			name:      "successful update default flag",
			catalogID: "cat_123",
			updateDef: true,
			mockResp: &api.CompanyCatalog{
				ID:        "cat_123",
				CompanyID: "comp_456",
				IsDefault: true,
				Status:    "active",
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			name:       "successful update JSON output",
			catalogID:  "cat_123",
			updateName: true,
			mockResp: &api.CompanyCatalog{
				ID:        "cat_123",
				CompanyID: "comp_456",
				Name:      "Updated Catalog",
				Status:    "active",
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC),
			},
			jsonOutput: true,
		},
		{
			name:       "update fails",
			catalogID:  "cat_456",
			updateName: true,
			mockErr:    errors.New("failed to update company catalog"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &companyCatalogsTestClient{
				updateCompanyCatalogResp: tt.mockResp,
				updateCompanyCatalogErr:  tt.mockErr,
			}
			cleanup := setupCompanyCatalogsMockFactories(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("name", "", "")
			cmd.Flags().String("description", "", "")
			cmd.Flags().StringSlice("product-ids", nil, "")
			cmd.Flags().Bool("default", false, "")
			if tt.jsonOutput {
				cmd.Flags().String("output", "json", "")
			} else {
				cmd.Flags().String("output", "", "")
			}

			// Set flags based on test case to trigger Changed() detection
			if tt.updateName {
				_ = cmd.Flags().Set("name", "Updated Catalog")
			}
			if tt.updateDesc {
				_ = cmd.Flags().Set("description", "Updated description")
			}
			if tt.updateProds {
				_ = cmd.Flags().Set("product-ids", "prod_new_1,prod_new_2")
			}
			if tt.updateDef {
				_ = cmd.Flags().Set("default", "true")
			}

			err := companyCatalogsUpdateCmd.RunE(cmd, []string{tt.catalogID})

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

// TestCompanyCatalogsDeleteRunE tests the company catalogs delete command execution with mock API.
func TestCompanyCatalogsDeleteRunE(t *testing.T) {
	tests := []struct {
		name      string
		catalogID string
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful delete",
			catalogID: "cat_123",
			mockErr:   nil,
		},
		{
			name:      "delete fails - not found",
			catalogID: "cat_456",
			mockErr:   errors.New("company catalog not found"),
			wantErr:   true,
		},
		{
			name:      "delete fails - API error",
			catalogID: "cat_789",
			mockErr:   errors.New("API unavailable"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &companyCatalogsTestClient{
				deleteCompanyCatalogErr: tt.mockErr,
			}
			cleanup := setupCompanyCatalogsMockFactories(mockClient)
			defer cleanup()

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().Bool("yes", true, "") // Skip confirmation

			err := companyCatalogsDeleteCmd.RunE(cmd, []string{tt.catalogID})

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

// TestCompanyCatalogsCreateFlags tests the create command flags.
func TestCompanyCatalogsCreateFlags(t *testing.T) {
	flags := companyCatalogsCreateCmd.Flags()

	requiredFlags := []string{"company-id", "name"}
	for _, flag := range requiredFlags {
		f := flags.Lookup(flag)
		if f == nil {
			t.Errorf("Expected flag %s", flag)
		}
	}

	optionalFlags := []string{"description", "product-ids", "default"}
	for _, flag := range optionalFlags {
		f := flags.Lookup(flag)
		if f == nil {
			t.Errorf("Expected optional flag %s", flag)
		}
	}
}

// TestCompanyCatalogsUpdateFlags tests the update command flags.
func TestCompanyCatalogsUpdateFlags(t *testing.T) {
	flags := companyCatalogsUpdateCmd.Flags()

	expectedFlags := []string{"name", "description", "product-ids", "default"}
	for _, flag := range expectedFlags {
		f := flags.Lookup(flag)
		if f == nil {
			t.Errorf("Expected flag %s", flag)
		}
	}
}

// TestCompanyCatalogsDeleteFlags tests the delete command flags.
func TestCompanyCatalogsDeleteFlags(t *testing.T) {
	flags := companyCatalogsDeleteCmd.Flags()

	if flags.Lookup("yes") == nil {
		t.Error("Expected yes flag")
	}
}

// TestCompanyCatalogsListWithFilters tests list command with filter options.
func TestCompanyCatalogsListWithFilters(t *testing.T) {
	mockClient := &companyCatalogsTestClient{
		listCompanyCatalogsResp: &api.CompanyCatalogsListResponse{
			Items: []api.CompanyCatalog{
				{
					ID:          "cat_123",
					CompanyID:   "comp_filter",
					CompanyName: "Filtered Company",
					Name:        "Filtered Catalog",
					Status:      "active",
					CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					UpdatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
		listCompanyCatalogsErr: nil,
	}
	cleanup := setupCompanyCatalogsMockFactories(mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("company-id", "comp_filter", "")
	cmd.Flags().String("status", "active", "")
	cmd.Flags().Int("page", 2, "")
	cmd.Flags().Int("page-size", 50, "")

	err := companyCatalogsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestCompanyCatalogsGetEmptyProductIDs tests get command with empty product IDs.
func TestCompanyCatalogsGetEmptyProductIDs(t *testing.T) {
	mockClient := &companyCatalogsTestClient{
		getCompanyCatalogResp: &api.CompanyCatalog{
			ID:          "cat_empty",
			CompanyID:   "comp_456",
			CompanyName: "Acme Corp",
			Name:        "Empty Catalog",
			Description: "Catalog with no products",
			ProductIDs:  []string{},
			IsDefault:   false,
			Status:      "active",
			CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		getCompanyCatalogErr: nil,
	}
	cleanup := setupCompanyCatalogsMockFactories(mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := companyCatalogsGetCmd.RunE(cmd, []string{"cat_empty"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Output goes to stdout (fmt.Printf), not the formatter buffer, so we just verify no error
}
